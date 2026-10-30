// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package gitea

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

// 실제 서비스 대신 소유 loopback 서버에서 요청 bytes를 관측한다. 전역 client는 바꾸지 않는다.
func providerFixture(t *testing.T, status int, body string) (*GiteaProvider, <-chan observedProviderRequest) {
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
	p := NewGiteaProvider(server.URL + "/api/v1/")
	p.SetToken("fixture-token")
	return p, requests
}

func nextProviderRequest(t *testing.T, requests <-chan observedProviderRequest) observedProviderRequest {
	t.Helper()
	select {
	case request := <-requests:
		require.NoError(t, request.err)
		require.Equal(t, "token fixture-token", request.header.Get("Authorization"))
		return request
	case <-time.After(time.Second):
		t.Fatal("fixture did not observe request")
		return observedProviderRequest{}
	}
}

func TestProviderBehaviorUpdatePreservesExplicitEmptyAndFalse(t *testing.T) {
	p, requests := providerFixture(t, http.StatusOK, `{"full_name":"acme/widget","name":"renamed","private":false,"archived":true}`)
	name, empty, branch, website := "renamed", "", "stable", "https://example.invalid"
	yes, no := true, false
	result, err := p.UpdateRepository(t.Context(), "acme/widget", provider.UpdateRepoRequest{
		Name: &name, Description: &empty, Private: &no, Archived: &yes, DefaultBranch: &branch,
		HasIssues: &no, HasWiki: &yes, HasProjects: &no, Homepage: &website,
	})
	require.NoError(t, err)
	require.Equal(t, "acme/widget", result.FullName)
	require.False(t, result.Private)
	require.True(t, result.Archived)
	r := nextProviderRequest(t, requests)
	require.Equal(t, "PATCH", r.method)
	require.Equal(t, "/api/v1/repos/acme/widget", r.path)
	require.Equal(t, "application/json", r.header.Get("Content-Type"))
	var payload map[string]any
	require.NoError(t, json.Unmarshal(r.body, &payload))
	require.Equal(t, map[string]any{
		"name": name, "description": "", "private": false, "archived": true,
		"default_branch": branch, "has_issues": false, "has_wiki": true, "has_projects": false, "website": website,
	}, payload)
}

func TestProviderBehaviorSearchPaginationAndEncoding(t *testing.T) {
	for _, page := range []int{0, 2} {
		t.Run(string(rune('0'+page)), func(t *testing.T) {
			p, requests := providerFixture(t, http.StatusOK, `{"ok":true,"data":[{"full_name":"acme/a"},{"full_name":"acme/b"}]}`)
			result, err := p.SearchRepositories(t.Context(), provider.SearchQuery{Query: "go + cli", Page: page, PerPage: 2, User: "user name", Topic: "go"})
			require.NoError(t, err)
			require.Len(t, result.Repositories, 2)
			require.Equal(t, "gitea", result.Repositories[0].ProviderType)
			require.True(t, result.HasNext)
			require.Equal(t, page > 1, result.HasPrev)
			r := nextProviderRequest(t, requests)
			require.Equal(t, "GET", r.method)
			require.Equal(t, "/api/v1/repos/search", r.path)
			require.Equal(t, "go + cli", r.query.Get("q"))
			require.Equal(t, "user name", r.query.Get("user"))
			require.Equal(t, "true", r.query.Get("topic"))
			require.Equal(t, "2", r.query.Get("limit"))
			require.Equal(t, max(page, 1), result.Page)
		})
	}
}

func TestProviderBehaviorWebhookDefaultsAndUpdate(t *testing.T) {
	p, requests := providerFixture(t, http.StatusOK, `{"id":17,"active":false,"config":{"url":"https://example.invalid/hook"}}`)
	hook, err := p.CreateWebhook(t.Context(), "acme/widget", provider.CreateWebhookRequest{Config: provider.WebhookConfig{URL: "https://example.invalid/hook"}})
	require.NoError(t, err)
	require.Equal(t, "17", hook.ID)
	require.Equal(t, "gitea", hook.Name)
	require.Equal(t, "json", hook.Config.ContentType)
	r := nextProviderRequest(t, requests)
	var payload map[string]any
	require.NoError(t, json.Unmarshal(r.body, &payload))
	require.Equal(t, []any{"push"}, payload["events"])
	require.Equal(t, false, payload["active"])
	require.Equal(t, "json", payload["config"].(map[string]any)["content_type"])
	no := false
	_, err = p.UpdateWebhook(t.Context(), "acme/widget", "17", provider.UpdateWebhookRequest{Active: &no, Events: []string{"issues"}, Config: &provider.WebhookConfig{URL: "https://example.invalid/new", ContentType: "form"}})
	require.NoError(t, err)
	r = nextProviderRequest(t, requests)
	require.Equal(t, "PATCH", r.method)
	require.Equal(t, "/api/v1/repos/acme/widget/hooks/17", r.path)
	payload = nil
	require.NoError(t, json.Unmarshal(r.body, &payload))
	require.Equal(t, map[string]any{
		"active": false, "events": []any{"issues"},
		"config": map[string]any{"url": "https://example.invalid/new", "content_type": "form"},
	}, payload)
	_, err = p.GetWebhook(t.Context(), "acme/widget", "17")
	require.NoError(t, err)
	require.Equal(t, "GET", nextProviderRequest(t, requests).method)
	test, err := p.TestWebhook(t.Context(), "acme/widget", "17")
	require.NoError(t, err)
	require.True(t, test.Success)
	require.Equal(t, "/api/v1/repos/acme/widget/hooks/17/tests", nextProviderRequest(t, requests).path)
}

func TestProviderBehaviorReportsRejectedOperations(t *testing.T) {
	operations := map[string]func(*GiteaProvider) error{
		"create repository": func(p *GiteaProvider) error {
			_, err := p.CreateRepository(t.Context(), provider.CreateRepoRequest{Owner: "acme", Name: "widget"})
			return err
		},
		"update repository": func(p *GiteaProvider) error {
			_, err := p.UpdateRepository(t.Context(), "acme/widget", provider.UpdateRepoRequest{})
			return err
		},
		"delete repository":    func(p *GiteaProvider) error { return p.DeleteRepository(t.Context(), "acme/widget") },
		"archive repository":   func(p *GiteaProvider) error { return p.ArchiveRepository(t.Context(), "acme/widget") },
		"unarchive repository": func(p *GiteaProvider) error { return p.UnarchiveRepository(t.Context(), "acme/widget") },
		"fork repository": func(p *GiteaProvider) error {
			_, err := p.ForkRepository(t.Context(), "acme/widget", provider.ForkOptions{})
			return err
		},
		"search repositories": func(p *GiteaProvider) error {
			_, err := p.SearchRepositories(t.Context(), provider.SearchQuery{Query: "widget"})
			return err
		},
		"list webhooks": func(p *GiteaProvider) error { _, err := p.ListWebhooks(t.Context(), "acme/widget"); return err },
		"get webhook":   func(p *GiteaProvider) error { _, err := p.GetWebhook(t.Context(), "acme/widget", "17"); return err },
		"create webhook": func(p *GiteaProvider) error {
			_, err := p.CreateWebhook(t.Context(), "acme/widget", provider.CreateWebhookRequest{Config: provider.WebhookConfig{URL: "https://example.invalid"}})
			return err
		},
		"update webhook": func(p *GiteaProvider) error {
			_, err := p.UpdateWebhook(t.Context(), "acme/widget", "17", provider.UpdateWebhookRequest{})
			return err
		},
		"delete webhook": func(p *GiteaProvider) error { return p.DeleteWebhook(t.Context(), "acme/widget", "17") },
		"test webhook":   func(p *GiteaProvider) error { _, err := p.TestWebhook(t.Context(), "acme/widget", "17"); return err },
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
	for _, body := range []string{"{", "not-json"} {
		p, _ := providerFixture(t, http.StatusOK, body)
		_, err := p.GetWebhook(t.Context(), "acme/widget", "17")
		require.ErrorContains(t, err, "failed to decode response")
	}
	p, requests := providerFixture(t, http.StatusOK, `{}`)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err := p.GetWebhook(ctx, "acme/widget", "17")
	require.ErrorIs(t, err, context.Canceled)
	select {
	case <-requests:
		t.Fatal("canceled request reached fixture")
	default:
	}
	require.ErrorContains(t, p.doJSON(t.Context(), "POST", "repos", make(chan int), nil, http.StatusOK), "failed to encode request")
	broken := NewGiteaProvider("://invalid")
	require.ErrorContains(t, broken.doJSON(t.Context(), "GET", "repos", nil, nil, http.StatusOK), "failed to create request")
}

func TestProviderBehaviorReleaseMetadataAndAssets(t *testing.T) {
	var input giteaRelease
	require.NoError(t, json.Unmarshal([]byte(`{"id":19,"tag_name":"v1","name":"release","body":"notes","target_commitish":"stable","draft":true,"prerelease":true,"html_url":"https://example.invalid/release","tarball_url":"tar","zipball_url":"zip","created_at":"2026-01-02T03:04:05Z","published_at":"2026-01-03T03:04:05Z","author":{"id":7,"login":"bot","full_name":"Fixture","email":"bot@example.invalid","avatar_url":"avatar"},"assets":[{"id":11,"name":"binary","size":123,"download_count":4,"browser_download_url":"download","created_at":"2026-01-04T03:04:05Z"}]}`), &input))
	r := convertGiteaRelease(input)
	require.Equal(t, "19", r.ID)
	require.Equal(t, "v1", r.TagName)
	require.Equal(t, "notes", r.Body)
	require.Equal(t, "stable", r.TargetBranch)
	require.True(t, r.Draft)
	require.True(t, r.Prerelease)
	require.Equal(t, "gitea", r.ProviderType)
	require.Equal(t, "7", r.Author.ID)
	require.Equal(t, "bot@example.invalid", r.Author.Email)
	require.Len(t, r.Assets, 1)
	require.Equal(t, provider.Asset{ID: "11", Name: "binary", Size: 123, DownloadCount: 4, DownloadURL: "download", CreatedAt: input.Assets[0].CreatedAt}, r.Assets[0])
	require.Equal(t, input.PublishedAt, r.PublishedAt)
	require.Nil(t, giteaRepoToProvider(nil))
}
