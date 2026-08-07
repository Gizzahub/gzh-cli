package gitlab

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

func TestGitLabProvider_CreateRepository(t *testing.T) {
	var createdPayload map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/namespaces/"):
			assert.Equal(t, "PRIVATE-TOKEN", "PRIVATE-TOKEN") // header key present check below
			assert.Equal(t, "glpat-test", r.Header.Get("PRIVATE-TOKEN"))
			_ = json.NewEncoder(w).Encode(map[string]any{"id": 42, "path": "acme"})
		case r.Method == http.MethodPost && r.URL.Path == "/projects":
			require.NoError(t, json.NewDecoder(r.Body).Decode(&createdPayload))
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":                  99,
				"name":                createdPayload["name"],
				"path_with_namespace": "acme/" + createdPayload["name"].(string),
				"description":         createdPayload["description"],
				"visibility":          createdPayload["visibility"],
				"http_url_to_repo":    "https://gitlab.example/acme/widget.git",
				"ssh_url_to_repo":     "git@gitlab.example:acme/widget.git",
				"web_url":             "https://gitlab.example/acme/widget",
				"default_branch":      "main",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	p := NewGitLabProvider(server.URL)
	p.SetToken("glpat-test")

	t.Run("owner set creates project", func(t *testing.T) {
		repo, err := p.CreateRepository(context.Background(), provider.CreateRepoRequest{
			Owner:       "acme",
			Name:        "widget",
			Description: "demo",
			Private:     true,
			AutoInit:    true,
		})
		require.NoError(t, err)
		require.NotNil(t, repo)
		assert.Equal(t, "widget", repo.Name)
		assert.Equal(t, "acme/widget", repo.FullName)
		assert.Equal(t, float64(42), createdPayload["namespace_id"])
		assert.Equal(t, "private", createdPayload["visibility"])
		assert.Equal(t, true, createdPayload["initialize_with_readme"])
	})

	t.Run("empty owner fails fast", func(t *testing.T) {
		_, err := p.CreateRepository(context.Background(), provider.CreateRepoRequest{Name: "widget"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "owner is required")
	})
}

func TestGitLabProvider_DeleteArchiveSearch(t *testing.T) {
	var methods []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methods = append(methods, r.Method+" "+r.URL.Path+"?"+r.URL.RawQuery)
		switch {
		case r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/projects/"):
			w.WriteHeader(http.StatusAccepted)
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/archive"):
			w.WriteHeader(http.StatusCreated)
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/unarchive"):
			w.WriteHeader(http.StatusCreated)
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/projects"):
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{
					"id":                  1,
					"name":                "widget",
					"path_with_namespace": "acme/widget",
					"web_url":             "https://gitlab.example/acme/widget",
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	p := NewGitLabProvider(server.URL)
	p.SetToken("glpat-test")
	ctx := context.Background()

	require.NoError(t, p.DeleteRepository(ctx, "acme/widget"))
	require.NoError(t, p.ArchiveRepository(ctx, "acme/widget"))
	require.NoError(t, p.UnarchiveRepository(ctx, "acme/widget"))

	result, err := p.SearchRepositories(ctx, provider.SearchQuery{Query: "widget"})
	require.NoError(t, err)
	require.Len(t, result.Repositories, 1)
	assert.Equal(t, "acme/widget", result.Repositories[0].FullName)

	joined := strings.Join(methods, "\n")
	// Server sees decoded Path; verify project path segment was sent (escaped on the wire).
	assert.Contains(t, joined, "DELETE /projects/acme/widget")
	assert.Contains(t, joined, "POST /projects/acme/widget/archive")
	assert.Contains(t, joined, "POST /projects/acme/widget/unarchive")
	assert.Contains(t, joined, "GET /projects?")
}

func TestGitLabProvider_SearchRequiresQuery(t *testing.T) {
	p := NewGitLabProvider("http://example.invalid")
	_, err := p.SearchRepositories(context.Background(), provider.SearchQuery{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "search query is required")
}

func TestGitLabProvider_UpdateAndFork(t *testing.T) {
	var updatePayload, forkPayload map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPut && strings.HasPrefix(r.URL.Path, "/projects/"):
			require.NoError(t, json.NewDecoder(r.Body).Decode(&updatePayload))
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":                  1,
				"name":                "widget",
				"path_with_namespace": "acme/widget",
				"description":         updatePayload["description"],
				"visibility":          updatePayload["visibility"],
			})
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/fork"):
			require.NoError(t, json.NewDecoder(r.Body).Decode(&forkPayload))
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":                  2,
				"name":                "widget",
				"path_with_namespace": "mine/widget",
				"web_url":             "https://gitlab.example/mine/widget",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	p := NewGitLabProvider(server.URL)
	p.SetToken("glpat-test")
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
	assert.Equal(t, "private", updatePayload["visibility"])

	forked, err := p.ForkRepository(ctx, "acme/widget", provider.ForkOptions{Organization: "mine"})
	require.NoError(t, err)
	assert.Equal(t, "mine/widget", forked.FullName)
	assert.Equal(t, "mine", forkPayload["namespace_path"])

	_, err = p.UpdateRepository(ctx, "", provider.UpdateRepoRequest{})
	require.Error(t, err)
}
