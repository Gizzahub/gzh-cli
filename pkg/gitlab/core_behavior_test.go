// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package gitlab

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResilientGitLabClientListGroupProjectsPagination(t *testing.T) {
	type requestRecord struct {
		path   string
		query  url.Values
		token  string
		accept string
		agent  string
	}
	var requests []requestRecord
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, requestRecord{r.URL.Path, r.URL.Query(), r.Header.Get("PRIVATE-TOKEN"), r.Header.Get("Accept"), r.Header.Get("User-Agent")})
		page, err := strconv.Atoi(r.URL.Query().Get("page"))
		if err != nil {
			http.Error(w, "bad page", http.StatusBadRequest)
			return
		}
		w.Header().Set("X-Total-Pages", "2")
		w.Header().Set("X-Page", strconv.Itoa(page))
		_ = json.NewEncoder(w).Encode([]ProjectInfo{{ID: page, Name: fmt.Sprintf("repo-%d", page)}})
	}))
	t.Cleanup(server.Close)

	client := NewResilientGitLabClient(server.URL, "private-token")
	client.httpClient = NewHTTPClientAdapterWithClient(server.Client())
	projects, err := client.ListGroupProjects(context.Background(), "platform/team")
	require.NoError(t, err)
	require.Len(t, projects, 2)
	assert.Equal(t, []string{"repo-1", "repo-2"}, []string{projects[0].Name, projects[1].Name})
	require.Len(t, requests, 2)
	for i, request := range requests {
		assert.Equal(t, "/api/v4/groups/platform/team/projects", request.path)
		assert.Equal(t, strconv.Itoa(i+1), request.query.Get("page"))
		assert.Equal(t, "100", request.query.Get("per_page"))
		assert.Equal(t, "true", request.query.Get("include_subgroups"))
		assert.Equal(t, "private-token", request.token)
		assert.Equal(t, "application/json", request.accept)
		assert.Equal(t, "gzh-cli", request.agent)
	}
}

func TestResilientGitLabClientFailuresAndCancellation(t *testing.T) {
	t.Run("HTTP errors retain status meaning", func(t *testing.T) {
		cases := []struct {
			status int
			want   string
		}{
			{http.StatusNotFound, "failed to get project: not found (404)"},
			{http.StatusUnauthorized, "failed to get project: unauthorized - check your token (401)"},
			{http.StatusForbidden, "failed to get project: forbidden - insufficient permissions (403)"},
			{http.StatusTooManyRequests, "failed to get project: rate limited - please retry later (429)"},
			{http.StatusInternalServerError, "failed to get project: server error (500)"},
			{http.StatusBadGateway, "failed to get project: bad gateway (502)"},
			{http.StatusServiceUnavailable, "failed to get project: service unavailable (503)"},
			{http.StatusGatewayTimeout, "failed to get project: gateway timeout (504)"},
			{http.StatusTeapot, "failed to get project: HTTP 418 - 418 I'm a teapot"},
		}
		for _, tc := range cases {
			t.Run(strconv.Itoa(tc.status), func(t *testing.T) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					http.Error(w, "denied", tc.status)
				}))
				t.Cleanup(server.Close)
				client := NewResilientGitLabClient(server.URL, "token")
				client.httpClient = NewHTTPClientAdapterWithClient(server.Client())

				_, err := client.GetProject(context.Background(), "42")
				require.Error(t, err)
				assert.EqualError(t, err, tc.want)
			})
		}
	})

	t.Run("malformed JSON is returned as a decode error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("{"))
		}))
		t.Cleanup(server.Close)
		client := NewResilientGitLabClient(server.URL, "")
		client.httpClient = NewHTTPClientAdapterWithClient(server.Client())

		_, err := client.ListGroups(context.Background())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode response")
	})

	t.Run("context cancellation reaches the request", func(t *testing.T) {
		requestStarted := make(chan struct{})
		requestCanceled := make(chan struct{})
		cleanupRelease := make(chan struct{})
		server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			close(requestStarted)
			select {
			case <-r.Context().Done():
				close(requestCanceled)
			case <-cleanupRelease:
			}
		}))
		t.Cleanup(server.Close)
		t.Cleanup(func() { close(cleanupRelease) })
		client := NewResilientGitLabClient(server.URL, "")
		client.httpClient = NewHTTPClientAdapterWithClient(server.Client())
		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)
		result := make(chan error, 1)
		go func() {
			_, err := client.GetProject(ctx, "42")
			result <- err
		}()
		select {
		case <-requestStarted:
		case <-time.After(time.Second):
			t.Fatal("server did not receive request")
		}
		cancel()
		var err error
		select {
		case err = <-result:
		case <-time.After(time.Second):
			t.Fatal("client did not return after request cancellation")
		}
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get project")
		select {
		case <-requestCanceled:
		case <-time.After(time.Second):
			t.Fatal("server did not observe canceled request")
		}
	})
}

func TestStreamingClientStreamsPaginatedProjects(t *testing.T) {
	var pages []string
	client := NewStreamingClient("stream-token", "https://gitlab.invalid", StreamingConfig{})
	client.httpClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		pages = append(pages, r.URL.Query().Get("page"))
		assert.Equal(t, "/api/v4/groups/engineering%2Fplatform/projects", r.URL.EscapedPath())
		assert.Equal(t, "2", r.URL.Query().Get("per_page"))
		assert.Equal(t, "stream-token", r.Header.Get("PRIVATE-TOKEN"))
		if r.URL.Query().Get("page") == "1" {
			header := make(http.Header)
			header.Set("X-Next-Page", "2")
			header.Set("X-Total-Pages", "2")
			return testHTTPResponse(r, http.StatusOK, `[{"id":1,"name":"first"}]`, header), nil
		}
		return testHTTPResponse(r, http.StatusOK, `[{"id":2,"name":"second"}]`, make(http.Header)), nil
	})}

	config := StreamingConfig{PageSize: 2, BufferSize: 4, MemoryLimit: 1024, RateLimitBuffer: 0}
	stream, err := client.StreamGroupProjects(context.Background(), "engineering/platform", config)
	require.NoError(t, err)
	var got []string
	for item := range stream {
		require.NoError(t, item.Error)
		require.NotNil(t, item.Project)
		got = append(got, item.Project.Name)
	}
	assert.Equal(t, []string{"first", "second"}, got)
	assert.Equal(t, []string{"1", "2"}, pages)
}

func TestStreamingClientReportsRequestFailure(t *testing.T) {
	client := NewStreamingClient("", "https://gitlab.invalid", StreamingConfig{})
	client.httpClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return testHTTPResponse(r, http.StatusServiceUnavailable, "unavailable", make(http.Header)), nil
	})}

	stream, err := client.StreamGroupProjects(context.Background(), "team", StreamingConfig{
		PageSize: 1, BufferSize: 1, MemoryLimit: 1024, RateLimitBuffer: 0,
	})
	require.NoError(t, err)
	item, ok := <-stream
	require.True(t, ok)
	require.Error(t, item.Error)
	assert.Contains(t, item.Error.Error(), "failed to fetch page 1")
	assert.Contains(t, item.Error.Error(), "503 Service Unavailable")
	_, ok = <-stream
	assert.False(t, ok)
}

func TestStreamingClientCancellationInterruptsRequest(t *testing.T) {
	requestStarted := make(chan struct{})
	requestCanceled := make(chan struct{})
	cleanupRelease := make(chan struct{})
	t.Cleanup(func() { close(cleanupRelease) })
	client := NewStreamingClient("", "https://gitlab.invalid", StreamingConfig{})
	client.httpClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		close(requestStarted)
		select {
		case <-r.Context().Done():
			close(requestCanceled)
			return nil, r.Context().Err()
		case <-cleanupRelease:
			return testHTTPResponse(r, http.StatusOK, `[]`, make(http.Header)), nil
		}
	})}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	stream, err := client.StreamGroupProjects(ctx, "team", StreamingConfig{
		PageSize: 1, BufferSize: 1, MemoryLimit: 1024, RateLimitBuffer: 0,
	})
	require.NoError(t, err)
	select {
	case <-requestStarted:
	case <-time.After(time.Second):
		t.Fatal("transport did not receive streaming request")
	}
	cancel()
	var item ProjectStream
	var ok bool
	select {
	case item, ok = <-stream:
	case <-time.After(time.Second):
		t.Fatal("stream did not return after request cancellation")
	}
	require.True(t, ok)
	require.Error(t, item.Error)
	select {
	case <-requestCanceled:
	case <-time.After(time.Second):
		t.Fatal("transport did not observe canceled streaming request")
	}
	select {
	case _, ok = <-stream:
		assert.False(t, ok)
	case <-time.After(time.Second):
		t.Fatal("stream did not close after request cancellation")
	}
}

func TestResilientGitLabClientGetStatsAndClose(t *testing.T) {
	client := NewResilientGitLabClient("https://gitlab.example", "")
	assert.Equal(t, "standard_http_client", client.GetStats()["type"])
	client.Close()
	client.SetToken("rotated")
	client.SetBaseURL("https://gitlab.example/")
	assert.Equal(t, "rotated", client.token)
	assert.Equal(t, "https://gitlab.example", client.baseURL)
}

func TestResilientGitLabClientTransportError(t *testing.T) {
	client := NewResilientGitLabClient("http://127.0.0.1:1", "")
	client.httpClient = failingHTTPClient{err: errors.New("transport unavailable")}
	_, err := client.GetProject(context.Background(), "1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get project")
}

func TestCachedGitLabClientCacheHitsMissesAndInvalidation(t *testing.T) {
	client := &CachedGitLabClient{cache: sync.Map{}}
	cachedProjects := []*Project{{ID: 17, Name: "cached"}}
	client.cache.Store("projects:group-a", cachedProjects)

	gotProjects, err := client.ListGroupProjectsWithCache(context.Background(), "group-a")
	require.NoError(t, err)
	assert.Equal(t, cachedProjects, gotProjects)
	assert.Equal(t, 1, client.InvalidateGroupCache(context.Background(), "group-a"))
	assert.Equal(t, 0, client.InvalidateGroupCache(context.Background(), "group-a"))

	project := &Project{ID: 17, Name: "single"}
	client.cache.Store("project:17", project)
	gotProject, err := client.GetProjectWithCache(context.Background(), "17")
	require.NoError(t, err)
	assert.Same(t, project, gotProject)
	assert.Equal(t, 1, client.InvalidateProjectCache(context.Background(), "17"))
	assert.Equal(t, 0, client.InvalidateProjectCache(context.Background(), "17"))

	miss, err := client.GetProjectWithCache(context.Background(), "17")
	assert.Nil(t, miss)
	assert.EqualError(t, err, "single project fetch not implemented")
	assert.Equal(t, map[string]any{
		"type": "simple_sync_map",
		"note": "cache package removed, using simple sync.Map",
	}, client.GetCacheStats())
}

func TestCachedGitLabClientMissStoresFetchedProjects(t *testing.T) {
	requests := 0
	streamingClient := NewStreamingClient("fake-token", "https://gitlab.invalid", DefaultStreamingConfig())
	streamingClient.httpClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		requests++
		assert.Equal(t, "/api/v4/groups/ops%2Fplatform/projects", r.URL.EscapedPath())
		assert.Equal(t, "fake-token", r.Header.Get("PRIVATE-TOKEN"))
		return testHTTPResponse(r, http.StatusOK, `[{"id":17,"name":"agent","path_with_namespace":"ops/agent"}]`, make(http.Header)), nil
	})}
	client := &CachedGitLabClient{streamingClient: streamingClient, cache: sync.Map{}}

	first, err := client.ListGroupProjectsWithCache(context.Background(), "ops/platform")
	require.NoError(t, err)
	require.Len(t, first, 1)
	assert.Equal(t, "agent", first[0].Name)

	second, err := client.ListGroupProjectsWithCache(context.Background(), "ops/platform")
	require.NoError(t, err)
	assert.Equal(t, first, second)
	assert.Equal(t, 1, requests)

	assert.Equal(t, 1, client.InvalidateGroupCache(context.Background(), "ops/platform"))
	third, err := client.ListGroupProjectsWithCache(context.Background(), "ops/platform")
	require.NoError(t, err)
	assert.Equal(t, first, third)
	assert.Equal(t, 2, requests)
}

func TestCachedStreamingClientCacheHit(t *testing.T) {
	client := &CachedStreamingClient{cache: sync.Map{}}
	cached := []ProjectStream{
		{Project: &Project{ID: 1, Name: "first"}, Metadata: StreamMetadata{Page: 3, CacheHit: true}},
		{Project: &Project{ID: 2, Name: "second"}, Metadata: StreamMetadata{Page: 4, CacheHit: true}},
	}
	client.cache.Store("stream_projects:group-a", cached)

	stream, err := client.StreamGroupProjectsWithCache(context.Background(), "group-a", StreamingConfig{BufferSize: 2})
	require.NoError(t, err)
	var got []ProjectStream
	for item := range stream {
		got = append(got, item)
	}
	assert.Equal(t, cached, got)
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func testHTTPResponse(request *http.Request, status int, body string, header http.Header) *http.Response {
	if header == nil {
		header = make(http.Header)
	}
	header.Set("Content-Type", "application/json")
	return &http.Response{
		StatusCode: status,
		Status:     strconv.Itoa(status) + " " + http.StatusText(status),
		Header:     header,
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    request,
	}
}

type failingHTTPClient struct{ err error }

func (f failingHTTPClient) Do(context.Context, *http.Request) (*http.Response, error) {
	return nil, f.err
}

func (f failingHTTPClient) Get(context.Context, string) (*http.Response, error) {
	return nil, f.err
}

func (f failingHTTPClient) Post(context.Context, string, string, io.Reader) (*http.Response, error) {
	return nil, f.err
}

func (f failingHTTPClient) Put(context.Context, string, string, io.Reader) (*http.Response, error) {
	return nil, f.err
}

func (f failingHTTPClient) Patch(context.Context, string, string, io.Reader) (*http.Response, error) {
	return nil, f.err
}

func (f failingHTTPClient) Delete(context.Context, string) (*http.Response, error) {
	return nil, f.err
}
