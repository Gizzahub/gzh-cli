package gitea

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gizzahub/gzh-cli/pkg/git/provider"
)

func TestGiteaProvider_CreateRepository(t *testing.T) {
	var createdBody map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/orgs/acme/repos":
			assert.Equal(t, "token test-token", r.Header.Get("Authorization"))
			require.NoError(t, json.NewDecoder(r.Body).Decode(&createdBody))
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":             7,
				"name":           createdBody["name"],
				"full_name":      "acme/" + createdBody["name"].(string),
				"private":        createdBody["private"],
				"clone_url":      "https://gitea.example/acme/widget.git",
				"ssh_url":        "git@gitea.example:acme/widget.git",
				"html_url":       "https://gitea.example/acme/widget",
				"default_branch": "main",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	p := NewGiteaProvider(server.URL)
	p.SetToken("test-token")

	t.Run("owner set creates under org", func(t *testing.T) {
		repo, err := p.CreateRepository(context.Background(), provider.CreateRepoRequest{
			Owner:   "acme",
			Name:    "widget",
			Private: true,
		})
		require.NoError(t, err)
		require.NotNil(t, repo)
		assert.Equal(t, "widget", repo.Name)
		assert.Equal(t, "acme/widget", repo.FullName)
		assert.Equal(t, true, createdBody["private"])
	})

	t.Run("empty owner fails fast", func(t *testing.T) {
		_, err := p.CreateRepository(context.Background(), provider.CreateRepoRequest{Name: "widget"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "owner is required")
	})
}

func TestGiteaProvider_DeleteArchiveSearch(t *testing.T) {
	var seen []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, r.Method+" "+r.URL.Path+"?"+r.URL.RawQuery)
		switch {
		case r.Method == http.MethodDelete && r.URL.Path == "/repos/acme/widget":
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodPatch && r.URL.Path == "/repos/acme/widget":
			var body map[string]bool
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			assert.Contains(t, body, "archived")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{"full_name": "acme/widget", "archived": body["archived"]})
		case r.Method == http.MethodGet && r.URL.Path == "/repos/search":
			assert.Equal(t, "widget", r.URL.Query().Get("q"))
			_ = json.NewEncoder(w).Encode(map[string]any{
				"ok": true,
				"data": []map[string]any{
					{"name": "widget", "full_name": "acme/widget"},
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	p := NewGiteaProvider(server.URL)
	p.SetToken("test-token")
	ctx := context.Background()

	require.NoError(t, p.DeleteRepository(ctx, "acme/widget"))
	require.NoError(t, p.ArchiveRepository(ctx, "acme/widget"))
	require.NoError(t, p.UnarchiveRepository(ctx, "acme/widget"))

	result, err := p.SearchRepositories(ctx, provider.SearchQuery{Query: "widget"})
	require.NoError(t, err)
	require.Len(t, result.Repositories, 1)
	assert.Equal(t, "acme/widget", result.Repositories[0].FullName)

	joined := strings.Join(seen, "\n")
	assert.Contains(t, joined, "DELETE /repos/acme/widget")
	assert.Contains(t, joined, "PATCH /repos/acme/widget")
	assert.Contains(t, joined, "GET /repos/search?")
}

func TestGiteaProvider_CreateFallsBackToUser(t *testing.T) {
	var usedUserEndpoint bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/orgs/alice/repos":
			http.NotFound(w, r)
		case r.Method == http.MethodPost && r.URL.Path == "/user/repos":
			usedUserEndpoint = true
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"name":      "widget",
				"full_name": "alice/widget",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	p := NewGiteaProvider(server.URL)
	p.SetToken("test-token")

	repo, err := p.CreateRepository(context.Background(), provider.CreateRepoRequest{
		Owner: "alice",
		Name:  "widget",
	})
	require.NoError(t, err)
	assert.True(t, usedUserEndpoint)
	assert.Equal(t, "alice/widget", repo.FullName)
}

func TestGiteaProvider_UpdateAndFork(t *testing.T) {
	var updatePayload, forkPayload map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPatch && r.URL.Path == "/repos/acme/widget":
			require.NoError(t, json.NewDecoder(r.Body).Decode(&updatePayload))
			_ = json.NewEncoder(w).Encode(map[string]any{
				"name":        "widget",
				"full_name":   "acme/widget",
				"description": updatePayload["description"],
				"private":     updatePayload["private"],
			})
		case r.Method == http.MethodPost && r.URL.Path == "/repos/acme/widget/forks":
			require.NoError(t, json.NewDecoder(r.Body).Decode(&forkPayload))
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"name":      "widget",
				"full_name": "mine/widget",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	p := NewGiteaProvider(server.URL)
	p.SetToken("test-token")
	ctx := context.Background()

	desc := "new"
	private := true
	updated, err := p.UpdateRepository(ctx, "acme/widget", provider.UpdateRepoRequest{
		Description: &desc,
		Private:     &private,
	})
	require.NoError(t, err)
	assert.Equal(t, "acme/widget", updated.FullName)
	assert.Equal(t, "new", updatePayload["description"])
	assert.Equal(t, true, updatePayload["private"])

	forked, err := p.ForkRepository(ctx, "acme/widget", provider.ForkOptions{Organization: "mine"})
	require.NoError(t, err)
	assert.Equal(t, "mine/widget", forked.FullName)
	assert.Equal(t, "mine", forkPayload["organization"])

	_, err = p.UpdateRepository(ctx, "bad", provider.UpdateRepoRequest{})
	require.Error(t, err)
}

func TestGiteaProvider_Webhooks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/hooks"):
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{"id": 3, "type": "gitea", "active": true, "config": map[string]string{"url": "https://ex/h", "content_type": "json"}, "events": []string{"push"}},
			})
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/hooks"):
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id": 4, "type": "gitea", "active": true, "config": map[string]string{"url": "https://ex/n", "content_type": "json"}, "events": []string{"push"},
			})
		case r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/tests"):
			w.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	p := NewGiteaProvider(server.URL)
	p.SetToken("t")
	ctx := context.Background()

	hooks, err := p.ListWebhooks(ctx, "acme/widget")
	require.NoError(t, err)
	require.Len(t, hooks, 1)

	created, err := p.CreateWebhook(ctx, "acme/widget", provider.CreateWebhookRequest{
		Config: provider.WebhookConfig{URL: "https://ex/n"},
		Active: true,
	})
	require.NoError(t, err)
	assert.Equal(t, "4", created.ID)

	res, err := p.TestWebhook(ctx, "acme/widget", "4")
	require.NoError(t, err)
	assert.True(t, res.Success)

	require.NoError(t, p.DeleteWebhook(ctx, "acme/widget", "4"))
}
