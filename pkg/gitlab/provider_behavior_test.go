// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package gitlab

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gizzahub/gzh-cli/pkg/git/provider"
)

type observedProviderRequest struct {
	method, path string
	query        url.Values
	header       http.Header
	body         []byte
	err          error
}

// 인스턴스 base URL만 loopback fixture로 지정하고 패키지 전역 설정은 보존한다.
func providerFixture(t *testing.T, status int, body string) (*GitLabProvider, <-chan observedProviderRequest) {
	t.Helper()
	requests := make(chan observedProviderRequest, 8)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		select {
		case requests <- observedProviderRequest{r.Method, r.URL.EscapedPath(), r.URL.Query(), r.Header.Clone(), data, err}:
		default:
			t.Error("fixture received too many requests")
		}
		w.WriteHeader(status)
		_, _ = io.WriteString(w, body)
	}))
	t.Cleanup(server.Close)
	p := NewGitLabProvider(server.URL + "/api/v4/")
	p.SetToken("fixture-token")
	return p, requests
}

func nextProviderRequest(t *testing.T, requests <-chan observedProviderRequest) observedProviderRequest {
	t.Helper()
	select {
	case request := <-requests:
		require.NoError(t, request.err)
		require.Equal(t, "fixture-token", request.header.Get("PRIVATE-TOKEN"))
		return request
	case <-time.After(time.Second):
		t.Fatal("fixture did not observe request")
		return observedProviderRequest{}
	}
}

func TestProviderBehaviorUpdateVisibilityAndEscapedNamespace(t *testing.T) {
	for _, visibility := range []provider.VisibilityType{provider.VisibilityPrivate, provider.VisibilityPublic, provider.VisibilityInternal} {
		t.Run(string(visibility), func(t *testing.T) {
			p, requests := providerFixture(t, http.StatusOK, `{"name":"widget","path_with_namespace":"org/sub/widget","visibility":"internal","archived":true}`)
			name, empty, branch := "widget", "", "stable"
			yes, no := true, false
			result, err := p.UpdateRepository(t.Context(), "org/sub/widget", provider.UpdateRepoRequest{Name: &name, Description: &empty, DefaultBranch: &branch, Archived: &yes, Private: &no, Visibility: visibility})
			require.NoError(t, err)
			require.Equal(t, "org/sub/widget", result.FullName)
			require.False(t, result.Private)
			require.True(t, result.Archived)
			r := nextProviderRequest(t, requests)
			require.Equal(t, "PUT", r.method)
			require.Equal(t, "/api/v4/projects/org%2Fsub%2Fwidget", r.path)
			require.Equal(t, "application/json", r.header.Get("Content-Type"))
			var payload map[string]any
			require.NoError(t, json.Unmarshal(r.body, &payload))
			require.Equal(t, map[string]any{"name": name, "description": "", "default_branch": branch, "archived": true, "visibility": string(visibility)}, payload)
		})
	}
}

func TestProviderBehaviorNamespaceFailurePreventsCreation(t *testing.T) {
	for _, status := range []int{http.StatusOK, http.StatusNotFound} {
		p, requests := providerFixture(t, status, `{}`)
		_, err := p.CreateRepository(t.Context(), provider.CreateRepoRequest{Owner: "org/sub", Name: "widget"})
		require.ErrorContains(t, err, "namespace")
		r := nextProviderRequest(t, requests)
		require.Equal(t, "GET", r.method)
		require.Equal(t, "/api/v4/namespaces/org%2Fsub", r.path)
		select {
		case <-requests:
			t.Fatal("namespace failure must not create project")
		default:
		}
	}
}

func TestProviderBehaviorCreateUsesResolvedNamespaceAndVisibility(t *testing.T) {
	requests := make(chan observedProviderRequest, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		select {
		case requests <- observedProviderRequest{r.Method, r.URL.EscapedPath(), r.URL.Query(), r.Header.Clone(), data, err}:
		default:
			t.Error("unexpected extra request")
		}
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusCreated)
		}
		_, _ = io.WriteString(w, `{"id":42,"name":"widget","path_with_namespace":"org/widget","visibility":"internal"}`)
	}))
	t.Cleanup(server.Close)
	p := NewGitLabProvider(server.URL)
	p.SetToken("fixture-token")
	result, err := p.CreateRepository(t.Context(), provider.CreateRepoRequest{Owner: "org", Name: "widget", Private: true, Visibility: provider.VisibilityInternal, DefaultBranch: "stable", AutoInit: true})
	require.NoError(t, err)
	require.Equal(t, "org/widget", result.FullName)
	require.Equal(t, "GET", nextProviderRequest(t, requests).method)
	r := nextProviderRequest(t, requests)
	require.Equal(t, "POST", r.method)
	require.Equal(t, "/projects", r.path)
	var payload map[string]any
	require.NoError(t, json.Unmarshal(r.body, &payload))
	require.Equal(t, map[string]any{"name": "widget", "namespace_id": float64(42), "description": "", "visibility": "internal", "default_branch": "stable", "initialize_with_readme": true}, payload)
}

func TestProviderBehaviorSearchPaginationAndEncoding(t *testing.T) {
	p, requests := providerFixture(t, http.StatusOK, `[{"name":"a","path_with_namespace":"org/a"},{"name":"b","path_with_namespace":"org/b"}]`)
	result, err := p.SearchRepositories(t.Context(), provider.SearchQuery{Query: "go + cli", Organization: "org/sub", Page: 2, PerPage: 2, Order: "desc"})
	require.NoError(t, err)
	require.Len(t, result.Repositories, 2)
	require.True(t, result.HasNext)
	require.True(t, result.HasPrev)
	require.Equal(t, 2, result.Page)
	require.Equal(t, "gitlab", result.Repositories[0].ProviderType)
	r := nextProviderRequest(t, requests)
	require.Equal(t, "GET", r.method)
	require.Equal(t, "/api/v4/projects", r.path)
	require.Equal(t, url.Values{"search": {"go + cli"}, "namespace_path": {"org/sub"}, "page": {"2"}, "per_page": {"2"}, "sort": {"desc"}}, r.query)
}

func TestProviderBehaviorWebhookUpdateAndRead(t *testing.T) {
	p, requests := providerFixture(t, http.StatusOK, `{"id":17,"url":"https://example.invalid/new","push_events":false,"enable_ssl_verification":false}`)
	no := false
	hook, err := p.UpdateWebhook(t.Context(), "org/sub/widget", "17", provider.UpdateWebhookRequest{Active: &no, Config: &provider.WebhookConfig{URL: "https://example.invalid/new", Secret: "fixture-secret", InsecureSSL: true}})
	require.NoError(t, err)
	require.Equal(t, "17", hook.ID)
	require.False(t, hook.Active)
	require.True(t, hook.Config.InsecureSSL)
	r := nextProviderRequest(t, requests)
	require.Equal(t, "PUT", r.method)
	require.Equal(t, "/api/v4/projects/org%2Fsub%2Fwidget/hooks/17", r.path)
	var payload map[string]any
	require.NoError(t, json.Unmarshal(r.body, &payload))
	require.Equal(t, map[string]any{"url": "https://example.invalid/new", "token": "fixture-secret", "enable_ssl_verification": false, "push_events": false}, payload)
	hook, err = p.GetWebhook(t.Context(), "org/sub/widget", "17")
	require.NoError(t, err)
	require.Equal(t, "https://example.invalid/new", hook.URL)
	require.Equal(t, "GET", nextProviderRequest(t, requests).method)
}

func TestProviderBehaviorReportsRejectedOperations(t *testing.T) {
	operations := map[string]func(*GitLabProvider) error{
		"update repository": func(p *GitLabProvider) error {
			_, err := p.UpdateRepository(t.Context(), "org/widget", provider.UpdateRepoRequest{})
			return err
		},
		"delete repository":    func(p *GitLabProvider) error { return p.DeleteRepository(t.Context(), "org/widget") },
		"archive repository":   func(p *GitLabProvider) error { return p.ArchiveRepository(t.Context(), "org/widget") },
		"unarchive repository": func(p *GitLabProvider) error { return p.UnarchiveRepository(t.Context(), "org/widget") },
		"fork repository": func(p *GitLabProvider) error {
			_, err := p.ForkRepository(t.Context(), "org/widget", provider.ForkOptions{})
			return err
		},
		"search repositories": func(p *GitLabProvider) error {
			_, err := p.SearchRepositories(t.Context(), provider.SearchQuery{Query: "widget"})
			return err
		},
		"list webhooks": func(p *GitLabProvider) error { _, err := p.ListWebhooks(t.Context(), "org/widget"); return err },
		"get webhook":   func(p *GitLabProvider) error { _, err := p.GetWebhook(t.Context(), "org/widget", "17"); return err },
		"create webhook": func(p *GitLabProvider) error {
			_, err := p.CreateWebhook(t.Context(), "org/widget", provider.CreateWebhookRequest{Config: provider.WebhookConfig{URL: "https://example.invalid"}})
			return err
		},
		"update webhook": func(p *GitLabProvider) error {
			_, err := p.UpdateWebhook(t.Context(), "org/widget", "17", provider.UpdateWebhookRequest{})
			return err
		},
		"delete webhook": func(p *GitLabProvider) error { return p.DeleteWebhook(t.Context(), "org/widget", "17") },
	}
	for name, operation := range operations {
		t.Run(name, func(t *testing.T) {
			p, requests := providerFixture(t, http.StatusForbidden, `denied fixture`)
			err := operation(p)
			require.ErrorContains(t, err, name)
			require.ErrorContains(t, err, "403")
			require.ErrorContains(t, err, "denied fixture")
			nextProviderRequest(t, requests)
		})
	}
}

func TestProviderBehaviorJSONFailuresAndCancellation(t *testing.T) {
	p, requests := providerFixture(t, http.StatusOK, "{")
	_, err := p.GetWebhook(t.Context(), "org/widget", "17")
	require.ErrorContains(t, err, "failed to decode response")
	nextProviderRequest(t, requests)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = p.GetWebhook(ctx, "org/widget", "17")
	require.ErrorIs(t, err, context.Canceled)
	select {
	case <-requests:
		t.Fatal("canceled request reached fixture")
	default:
	}
	require.ErrorContains(t, p.doJSON(t.Context(), "POST", "projects", make(chan int), nil, http.StatusOK), "failed to encode request")
	broken := NewGitLabProvider("://invalid")
	require.ErrorContains(t, broken.doJSON(t.Context(), "GET", "projects", nil, nil, http.StatusOK), "failed to create request")
}

func TestProviderBehaviorReleaseMetadataAndAssets(t *testing.T) {
	var input gitlabRelease
	require.NoError(t, json.Unmarshal([]byte(`{"name":"release","tag_name":"v1","description":"notes","created_at":"2026-01-02T03:04:05Z","released_at":"2026-01-03T03:04:05Z","author":{"id":7,"username":"bot","name":"Fixture","avatar_url":"avatar","web_url":"user"},"_links":{"self":"release-url"},"assets":{"links":[{"id":11,"name":"binary","url":"download"}],"sources":[{"format":"tar.gz","url":"source"}]}}`), &input))
	r := convertGitLabRelease(input)
	require.Equal(t, "v1", r.ID)
	require.Equal(t, "notes", r.Body)
	require.Equal(t, "release-url", r.HTMLURL)
	require.Equal(t, "gitlab", r.ProviderType)
	require.Equal(t, provider.Actor{ID: "7", Login: "bot", Name: "Fixture", AvatarURL: "avatar", HTMLURL: "user"}, r.Author)
	require.Equal(t, []provider.Asset{{ID: "11", Name: "binary", DownloadURL: "download"}, {Name: "Source code (tar.gz)", DownloadURL: "source", ContentType: "application/tar.gz"}}, r.Assets)
	require.Equal(t, input.CreatedAt, r.CreatedAt)
	require.Equal(t, input.ReleasedAt, r.PublishedAt)
	require.False(t, r.Draft)
	require.False(t, r.Prerelease)
	require.Nil(t, gitlabProjectToProvider(nil))
}
