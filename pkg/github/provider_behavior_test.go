// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitHubAPIClientSearchRepositoriesRequestContract(t *testing.T) {
	var observed *http.Request
	client := newProviderBehaviorAPIClient(t, func(r *http.Request) (*http.Response, error) {
		observed = r
		return providerBehaviorResponse(r, http.StatusOK, `{"total_count":1,"incomplete_results":false,"items":[{"name":"cli","full_name":"octo/cli"}]}`), nil
	})

	result, err := client.SearchRepositories(context.Background(), "org:octo language:go", &SearchRepositoriesOptions{
		Sort:    "updated",
		Order:   "desc",
		Page:    3,
		PerPage: 50,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 1, result.TotalCount)
	require.Len(t, result.Repositories, 1)
	assert.Equal(t, "octo/cli", result.Repositories[0].FullName)
	require.NotNil(t, observed)
	assert.Equal(t, http.MethodGet, observed.Method)
	assert.Equal(t, "/search/repositories", observed.URL.Path)
	assert.Equal(t, url.Values{
		"q":        {"org:octo language:go"},
		"page":     {"3"},
		"per_page": {"50"},
		"sort":     {"updated"},
		"order":    {"desc"},
	}, observed.URL.Query())
	assertProviderBehaviorHeaders(t, observed)
}

func TestGitHubAPIClientCreateRepositoryFallsBackFromOrganization404(t *testing.T) {
	var paths []string
	var payloads []map[string]any
	client := newProviderBehaviorAPIClient(t, func(r *http.Request) (*http.Response, error) {
		paths = append(paths, r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)
		assertProviderBehaviorHeaders(t, r)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		var payload map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		payloads = append(payloads, payload)
		if r.URL.Path == "/orgs/octo/repos" {
			return providerBehaviorResponse(r, http.StatusNotFound, `{"message":"organization not found"}`), nil
		}
		return providerBehaviorResponse(r, http.StatusCreated, `{"name":"cli","full_name":"octo/cli","private":true}`), nil
	})

	created, err := client.CreateRepository(context.Background(), "octo", &CreateRepositoryOptions{
		Name:        "cli",
		Description: "command line tools",
		Private:     true,
		AutoInit:    true,
	})
	require.NoError(t, err)
	assert.Equal(t, "octo/cli", created.FullName)
	assert.Equal(t, []string{"/orgs/octo/repos", "/user/repos"}, paths)
	require.Len(t, payloads, 2)
	assert.Equal(t, payloads[0], payloads[1])
	assert.Equal(t, "cli", payloads[0]["name"])
	assert.Equal(t, "command line tools", payloads[0]["description"])
	assert.Equal(t, true, payloads[0]["private"])
	assert.Equal(t, true, payloads[0]["auto_init"])
}

func TestGitHubAPIClientCreateRepositoryDoesNotFallbackForForbidden(t *testing.T) {
	requests := 0
	client := newProviderBehaviorAPIClient(t, func(r *http.Request) (*http.Response, error) {
		requests++
		assert.Equal(t, "/orgs/octo/repos", r.URL.Path)
		return providerBehaviorResponse(r, http.StatusForbidden, `{"message":"permission denied"}`), nil
	})

	_, err := client.CreateRepository(context.Background(), "octo", &CreateRepositoryOptions{Name: "cli"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP 403")
	assert.Equal(t, 1, requests)
}

func TestGitHubAPIClientUpdateRepositorySendsOnlyRequestedFields(t *testing.T) {
	description := "updated description"
	private := false
	var observed *http.Request
	var payload map[string]any
	client := newProviderBehaviorAPIClient(t, func(r *http.Request) (*http.Response, error) {
		observed = r
		assertProviderBehaviorHeaders(t, r)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		return providerBehaviorResponse(r, http.StatusOK, `{"name":"cli","full_name":"octo/cli","description":"updated description","private":false}`), nil
	})

	updated, err := client.UpdateRepository(context.Background(), "octo", "cli", &UpdateRepositoryOptions{
		Description: &description,
		Private:     &private,
	})
	require.NoError(t, err)
	assert.Equal(t, "updated description", updated.Description)
	require.NotNil(t, observed)
	assert.Equal(t, http.MethodPatch, observed.Method)
	assert.Equal(t, "/repos/octo/cli", observed.URL.Path)
	assert.Equal(t, map[string]any{
		"description": "updated description",
		"private":     false,
	}, payload)
}

func TestGitHubAPIClientForkUsesAcceptedStatusAndOptions(t *testing.T) {
	var observed *http.Request
	var payload ForkRepositoryOptions
	client := newProviderBehaviorAPIClient(t, func(r *http.Request) (*http.Response, error) {
		observed = r
		assertProviderBehaviorHeaders(t, r)
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		return providerBehaviorResponse(r, http.StatusAccepted, `{"name":"cli","full_name":"forker/cli"}`), nil
	})

	forked, err := client.ForkRepository(context.Background(), "octo", "cli", &ForkRepositoryOptions{
		Organization:      "forker",
		Name:              "cli",
		DefaultBranchOnly: true,
	})
	require.NoError(t, err)
	assert.Equal(t, "forker/cli", forked.FullName)
	require.NotNil(t, observed)
	assert.Equal(t, http.MethodPost, observed.Method)
	assert.Equal(t, "/repos/octo/cli/forks", observed.URL.Path)
	assert.Equal(t, ForkRepositoryOptions{Organization: "forker", Name: "cli", DefaultBranchOnly: true}, payload)
}

func TestResilientGitHubClientCancellationUnblocksRequest(t *testing.T) {
	requestStarted := make(chan struct{})
	requestCanceled := make(chan struct{})
	cleanupRelease := make(chan struct{})
	client := NewResilientGitHubClientWithBaseURL("fake-token", "https://github.invalid")
	client.httpClient.Transport = providerBehaviorRoundTripper(func(r *http.Request) (*http.Response, error) {
		close(requestStarted)
		select {
		case <-r.Context().Done():
			close(requestCanceled)
			return nil, r.Context().Err()
		case <-cleanupRelease:
			return providerBehaviorResponse(r, http.StatusOK, `{"default_branch":"main"}`), nil
		}
	})
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	t.Cleanup(func() { close(cleanupRelease) })
	result := make(chan error, 1)
	go func() {
		_, err := client.GetDefaultBranch(ctx, "octo", "cli")
		result <- err
	}()
	select {
	case <-requestStarted:
	case <-time.After(time.Second):
		t.Fatal("transport did not receive repository request")
	}
	cancel()
	select {
	case err := <-result:
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get repository info")
	case <-time.After(time.Second):
		t.Fatal("request did not return after context cancellation")
	}
	select {
	case <-requestCanceled:
	case <-time.After(time.Second):
		t.Fatal("transport did not observe canceled request")
	}
}

func newProviderBehaviorAPIClient(t *testing.T, roundTrip func(*http.Request) (*http.Response, error)) *GitHubAPIClient {
	t.Helper()
	config := &APIClientConfig{
		BaseURL:   "https://api.github.invalid",
		Token:     "fake-token",
		UserAgent: "provider-contract-test",
	}
	return &GitHubAPIClient{
		config:     config,
		httpClient: providerBehaviorHTTPClient{client: &http.Client{Transport: providerBehaviorRoundTripper(roundTrip)}},
		logger:     &mockLogger{},
	}
}

func assertProviderBehaviorHeaders(t *testing.T, r *http.Request) {
	t.Helper()
	assert.Equal(t, "token fake-token", r.Header.Get("Authorization"))
	assert.Equal(t, "application/vnd.github.v3+json", r.Header.Get("Accept"))
	assert.Equal(t, "provider-contract-test", r.Header.Get("User-Agent"))
}

func providerBehaviorResponse(r *http.Request, status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Status:     fmt.Sprintf("%d %s", status, http.StatusText(status)),
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    r,
	}
}

type providerBehaviorRoundTripper func(*http.Request) (*http.Response, error)

func (f providerBehaviorRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

type providerBehaviorHTTPClient struct {
	client *http.Client
}

func (c providerBehaviorHTTPClient) Do(r *http.Request) (*http.Response, error) {
	return c.client.Do(r)
}

func (c providerBehaviorHTTPClient) Get(rawURL string) (*http.Response, error) {
	return c.client.Get(rawURL)
}

func (c providerBehaviorHTTPClient) Post(rawURL, contentType string, body any) (*http.Response, error) {
	var reader io.Reader
	switch value := body.(type) {
	case io.Reader:
		reader = value
	case []byte:
		reader = strings.NewReader(string(value))
	case string:
		reader = strings.NewReader(value)
	}
	return c.client.Post(rawURL, contentType, reader)
}
