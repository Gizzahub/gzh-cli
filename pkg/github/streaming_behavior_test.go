// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package github

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStreamingClientConvertsRepositoryAndCancelsNextRequest(t *testing.T) {
	secondRequestStarted := make(chan struct{})
	requestCanceled := make(chan struct{})
	cleanupRelease := make(chan struct{})
	t.Cleanup(func() { close(cleanupRelease) })
	var requests int
	client := NewStreamingClient("fake-token", StreamingConfig{})
	client.httpClient = &http.Client{Transport: streamingBehaviorRoundTripper(func(r *http.Request) (*http.Response, error) {
		requests++
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, DefaultGitHubAPIBaseURL+"/orgs/octo/repos", r.URL.Scheme+"://"+r.URL.Host+r.URL.Path)
		assert.Equal(t, "25", r.URL.Query().Get("per_page"))
		assert.Equal(t, "all", r.URL.Query().Get("type"))
		assert.Equal(t, "updated", r.URL.Query().Get("sort"))
		assert.Equal(t, "desc", r.URL.Query().Get("direction"))
		assert.Equal(t, "Bearer fake-token", r.Header.Get("Authorization"))
		assert.Equal(t, "application/vnd.github+json", r.Header.Get("Accept"))
		assert.Equal(t, "2022-11-28", r.Header.Get("X-GitHub-Api-Version"))
		if requests == 1 {
			return streamingBehaviorResponse(r, http.StatusOK, `[{"id":17,"name":"cli","full_name":"octo/cli","private":true,"default_branch":"main","language":"Go","topics":["tools"]}]`, http.Header{"Link": []string{`<https://api.github.com/orgs/octo/repos?page=2>; rel="next"`}}), nil
		}
		close(secondRequestStarted)
		select {
		case <-r.Context().Done():
			close(requestCanceled)
			return nil, r.Context().Err()
		case <-cleanupRelease:
			return streamingBehaviorResponse(r, http.StatusOK, `[]`, nil), nil
		}
	})}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	stream, err := client.StreamOrganizationRepositories(ctx, "octo", StreamingConfig{
		PageSize: 25, BufferSize: 2, MemoryLimit: 1024, RateLimitBuffer: 0,
	})
	require.NoError(t, err)
	var first RepositoryStream
	select {
	case first = <-stream:
	case <-time.After(time.Second):
		t.Fatal("stream did not return the first repository")
	}
	require.NoError(t, first.Error)
	require.NotNil(t, first.Repository)
	assert.Equal(t, int64(17), first.Repository.ID)
	assert.Equal(t, "octo/cli", first.Repository.FullName)
	assert.Equal(t, "main", first.Repository.DefaultBranch)
	assert.Equal(t, []string{"tools"}, first.Repository.Topics)
	assert.Equal(t, 1, first.Metadata.Page)
	select {
	case <-secondRequestStarted:
	case <-time.After(time.Second):
		t.Fatal("stream did not request the following page")
	}
	cancel()
	var next RepositoryStream
	select {
	case next = <-stream:
	case <-time.After(time.Second):
		t.Fatal("stream did not report the canceled request")
	}
	require.Error(t, next.Error)
	assert.Contains(t, next.Error.Error(), "failed to fetch page 2")
	select {
	case <-requestCanceled:
	case <-time.After(time.Second):
		t.Fatal("transport did not observe cancellation")
	}
	select {
	case _, ok := <-stream:
		assert.False(t, ok)
	case <-time.After(time.Second):
		t.Fatal("stream did not close after cancellation")
	}
}

func TestStreamingClientReturnsHTTPFailureAsStreamError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	requests := 0
	client := NewStreamingClient("fake-token", StreamingConfig{})
	client.httpClient = &http.Client{Transport: streamingBehaviorRoundTripper(func(r *http.Request) (*http.Response, error) {
		requests++
		if requests > 1 {
			return nil, fmt.Errorf("unexpected request after HTTP failure")
		}
		return streamingBehaviorResponse(r, http.StatusForbidden, `{"message":"denied"}`, nil), nil
	})}
	stream, err := client.StreamOrganizationRepositories(ctx, "octo", StreamingConfig{
		PageSize: 10, BufferSize: 1, MemoryLimit: 1024, RateLimitBuffer: 0,
	})
	require.NoError(t, err)
	var item RepositoryStream
	select {
	case item = <-stream:
	case <-time.After(time.Second):
		t.Fatal("stream did not report the HTTP failure")
	}
	require.Error(t, item.Error)
	assert.Nil(t, item.Repository)
	assert.Contains(t, item.Error.Error(), "API request failed: 403 Forbidden")
	select {
	case _, ok := <-stream:
		assert.False(t, ok)
	case <-time.After(time.Second):
		t.Fatal("stream did not close after the HTTP failure")
	}
}

func TestStreamingClientFollowsNextLinkAndStopsAfterFinalPage(t *testing.T) {
	var requestedPages []string
	var requests int
	client := NewStreamingClient("fake-token", StreamingConfig{})
	client.httpClient = &http.Client{Transport: streamingBehaviorRoundTripper(func(r *http.Request) (*http.Response, error) {
		requests++
		if requests > 2 {
			return nil, fmt.Errorf("unexpected request beyond page 2")
		}
		page := r.URL.Query().Get("page")
		requestedPages = append(requestedPages, page)
		assert.Equal(t, DefaultGitHubAPIBaseURL+"/orgs/octo/repos", r.URL.Scheme+"://"+r.URL.Host+r.URL.Path)
		switch page {
		case "":
			header := http.Header{"Link": []string{
				`<https://api.github.com/orgs/octo/repos?page=1>; rel="prev", <https://api.github.com/orgs/octo/repos?page=8>; rel="last", <https://api.github.com/orgs/octo/repos?page=2&per_page=25>; rel="next"`,
			}}
			return streamingBehaviorResponse(r, http.StatusOK, `[{"id":1,"name":"one","full_name":"octo/one"}]`, header), nil
		case "2":
			return streamingBehaviorResponse(r, http.StatusOK, `[{"id":2,"name":"two","full_name":"octo/two"}]`, nil), nil
		default:
			return nil, fmt.Errorf("unexpected page request %q", page)
		}
	})}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	stream, err := client.StreamOrganizationRepositories(ctx, "octo", StreamingConfig{
		PageSize: 25, BufferSize: 4, MemoryLimit: 1024, RateLimitBuffer: 0,
	})
	require.NoError(t, err)
	var items []RepositoryStream
	closed := false
	for range 3 {
		select {
		case item, ok := <-stream:
			if !ok {
				closed = true
				break
			}
			items = append(items, item)
		case <-time.After(time.Second):
			t.Fatal("stream did not return bounded page results")
		}
		if closed {
			break
		}
	}
	require.True(t, closed, "stream did not close after the final page")
	require.Len(t, items, 2)
	require.NoError(t, items[0].Error)
	require.NoError(t, items[1].Error)
	require.NotNil(t, items[0].Repository)
	require.NotNil(t, items[1].Repository)
	assert.Equal(t, []string{"", "2"}, requestedPages)
	assert.Equal(t, []int64{1, 2}, []int64{items[0].Repository.ID, items[1].Repository.ID})
	assert.Equal(t, []string{"octo/one", "octo/two"}, []string{items[0].Repository.FullName, items[1].Repository.FullName})
	assert.Equal(t, []int{1, 2}, []int{items[0].Metadata.Page, items[1].Metadata.Page})
}

func TestStreamingClientStopsWhenLinkHasNoNextRelation(t *testing.T) {
	requests := 0
	client := NewStreamingClient("fake-token", StreamingConfig{})
	client.httpClient = &http.Client{Transport: streamingBehaviorRoundTripper(func(r *http.Request) (*http.Response, error) {
		requests++
		if requests > 1 {
			return nil, fmt.Errorf("unexpected request after final page")
		}
		return streamingBehaviorResponse(r, http.StatusOK, `[{"id":3,"name":"three"}]`, nil), nil
	})}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	stream, err := client.StreamOrganizationRepositories(ctx, "octo", StreamingConfig{
		PageSize: 10, BufferSize: 2, MemoryLimit: 1024, RateLimitBuffer: 0,
	})
	require.NoError(t, err)
	var item RepositoryStream
	select {
	case item = <-stream:
	case <-time.After(time.Second):
		t.Fatal("stream did not return its repository")
	}
	require.NoError(t, item.Error)
	assert.Equal(t, int64(3), item.Repository.ID)
	select {
	case _, ok := <-stream:
		assert.False(t, ok)
	case <-time.After(time.Second):
		t.Fatal("stream did not close after a response without a next link")
	}
	assert.Equal(t, 1, requests)
}

func TestStreamingClientRejectsBackwardNextPage(t *testing.T) {
	var requestedPages []string
	var requests int
	client := NewStreamingClient("fake-token", StreamingConfig{})
	client.httpClient = &http.Client{Transport: streamingBehaviorRoundTripper(func(r *http.Request) (*http.Response, error) {
		requests++
		if requests > 2 {
			return nil, fmt.Errorf("unexpected request beyond page 2")
		}
		page := r.URL.Query().Get("page")
		requestedPages = append(requestedPages, page)
		if page == "" {
			return streamingBehaviorResponse(r, http.StatusOK, `[{"id":1,"name":"one"}]`, http.Header{"Link": []string{
				`<https://api.github.com/orgs/octo/repos?page=2>; rel="next"`,
			}}), nil
		}
		return streamingBehaviorResponse(r, http.StatusOK, `[{"id":2,"name":"two"}]`, http.Header{"Link": []string{
			`<https://api.github.com/orgs/octo/repos?page=1>; rel="next"`,
		}}), nil
	})}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	stream, err := client.StreamOrganizationRepositories(ctx, "octo", StreamingConfig{
		PageSize: 10, BufferSize: 3, MemoryLimit: 1024, RateLimitBuffer: 0,
	})
	require.NoError(t, err)
	var items []RepositoryStream
	closed := false
	for range 3 {
		select {
		case item, ok := <-stream:
			if !ok {
				closed = true
				break
			}
			items = append(items, item)
		case <-time.After(time.Second):
			t.Fatal("stream did not report bounded backward-link result")
		}
		if closed {
			break
		}
	}
	require.True(t, closed)
	require.Len(t, items, 2)
	require.NoError(t, items[0].Error)
	require.Error(t, items[1].Error)
	assert.Contains(t, items[1].Error.Error(), "invalid next page link")
	assert.Contains(t, items[1].Error.Error(), "greater than 2")
	assert.Equal(t, []string{"", "2"}, requestedPages)
}

func TestStreamingClientRejectsInvalidNextPageLink(t *testing.T) {
	cases := []struct {
		name string
		link string
	}{
		{name: "malformed page", link: `<https://api.github.com/orgs/octo/repos?page=oops>; rel="next"`},
		{name: "malformed link syntax", link: `https://api.github.com/orgs/octo/repos?page=2; rel="next"`},
		{name: "zero page", link: `<https://api.github.com/orgs/octo/repos?page=0>; rel="next"`},
		{name: "same page", link: `<https://api.github.com/orgs/octo/repos?page=1>; rel="next"`},
		{name: "cross host", link: `<https://attacker.invalid/orgs/octo/repos?page=2>; rel="next"`},
		{name: "different path", link: `<https://api.github.com/orgs/other/repos?page=2>; rel="next"`},
		{name: "changes page size", link: `<https://api.github.com/orgs/octo/repos?page=2&per_page=100>; rel="next"`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			requests := 0
			client := NewStreamingClient("fake-token", StreamingConfig{})
			client.httpClient = &http.Client{Transport: streamingBehaviorRoundTripper(func(r *http.Request) (*http.Response, error) {
				requests++
				if requests > 1 {
					return nil, fmt.Errorf("pagination continued after invalid next link")
				}
				return streamingBehaviorResponse(r, http.StatusOK, `[]`, http.Header{"Link": []string{tc.link}}), nil
			})}
			ctx, cancel := context.WithCancel(context.Background())
			t.Cleanup(cancel)
			stream, err := client.StreamOrganizationRepositories(ctx, "octo", StreamingConfig{
				PageSize: 10, BufferSize: 1, MemoryLimit: 1024, RateLimitBuffer: 0,
			})
			require.NoError(t, err)
			var item RepositoryStream
			select {
			case item = <-stream:
			case <-time.After(time.Second):
				t.Fatal("stream did not report invalid pagination")
			}
			require.Error(t, item.Error)
			assert.Contains(t, item.Error.Error(), "invalid next page link")
			select {
			case _, ok := <-stream:
				assert.False(t, ok)
			case <-time.After(time.Second):
				t.Fatal("stream did not close after invalid pagination")
			}
			assert.Equal(t, 1, requests)
		})
	}
}

func TestStreamingClientSendErrorWaitsForFullChannelToDrain(t *testing.T) {
	client := NewStreamingClient("fake-token", StreamingConfig{})
	results := make(chan RepositoryStream, 1)
	original := RepositoryStream{Repository: &Repository{ID: 9, Name: "queued"}}
	results <- original
	ctx, cancel := context.WithCancel(context.Background())
	observedCtx := &streamingBehaviorObservedDoneContext{Context: ctx, called: make(chan struct{})}
	returned := make(chan struct{})
	t.Cleanup(func() {
		cancel()
		select {
		case <-returned:
		case <-time.After(time.Second):
			t.Error("sendError did not stop during cleanup")
		}
	})
	go func() {
		client.sendError(observedCtx, results, errors.New("terminal stream failure"))
		close(returned)
	}()
	select {
	case <-observedCtx.called:
	case <-returned:
		t.Fatal("sendError returned before waiting on the full result channel")
	case <-time.After(time.Second):
		t.Fatal("sendError did not reach the full-channel wait")
	}
	assert.Equal(t, original, <-results)
	var terminal RepositoryStream
	select {
	case terminal = <-results:
	case <-time.After(time.Second):
		t.Fatal("terminal error was not delivered after the channel drained")
	}
	require.Error(t, terminal.Error)
	assert.EqualError(t, terminal.Error, "terminal stream failure")
	select {
	case <-returned:
	case <-time.After(time.Second):
		t.Fatal("sendError did not return after delivery")
	}
	select {
	case extra := <-results:
		t.Fatalf("sendError delivered an unexpected extra result: %+v", extra)
	default:
	}
}

func TestStreamingClientSendErrorReturnsWhenCanceledAndChannelIsFull(t *testing.T) {
	client := NewStreamingClient("fake-token", StreamingConfig{})
	results := make(chan RepositoryStream, 1)
	original := RepositoryStream{Repository: &Repository{ID: 9, Name: "queued"}}
	results <- original
	ctx, cancel := context.WithCancel(context.Background())
	observedCtx := &streamingBehaviorObservedDoneContext{Context: ctx, called: make(chan struct{})}
	returned := make(chan struct{})
	t.Cleanup(func() {
		cancel()
		select {
		case <-returned:
		case <-time.After(time.Second):
			t.Error("sendError did not stop during cleanup")
		}
	})
	go func() {
		client.sendError(observedCtx, results, errors.New("terminal stream failure"))
		close(returned)
	}()
	select {
	case <-observedCtx.called:
	case <-returned:
		t.Fatal("sendError returned before waiting on the full result channel")
	case <-time.After(time.Second):
		t.Fatal("sendError did not reach the full-channel wait")
	}
	cancel()
	select {
	case <-returned:
	case <-time.After(time.Second):
		t.Fatal("sendError did not return after cancellation")
	}
	assert.Equal(t, original, <-results)
	select {
	case extra := <-results:
		t.Fatalf("canceled sendError delivered an unexpected result: %+v", extra)
	default:
	}
}

func TestStreamingClientSendErrorDeliversWhenCanceledAndChannelIsFree(t *testing.T) {
	client := NewStreamingClient("fake-token", StreamingConfig{})
	results := make(chan RepositoryStream, 1)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	client.sendError(ctx, results, errors.New("terminal stream failure"))
	select {
	case result := <-results:
		require.Error(t, result.Error)
		assert.EqualError(t, result.Error, "terminal stream failure")
	default:
		t.Fatal("sendError did not use available buffer space after cancellation")
	}
}

type streamingBehaviorObservedDoneContext struct {
	context.Context
	called chan struct{}
	once   sync.Once
}

// Done은 실제 컨텍스트의 취소 신호를 그대로 반환하면서 접근 시점을 기록한다.
func (c *streamingBehaviorObservedDoneContext) Done() <-chan struct{} {
	c.once.Do(func() { close(c.called) })
	return c.Context.Done()
}

func streamingBehaviorResponse(r *http.Request, status int, body string, header http.Header) *http.Response {
	if header == nil {
		header = make(http.Header)
	}
	return &http.Response{
		StatusCode: status,
		Status:     fmt.Sprintf("%d %s", status, http.StatusText(status)),
		Header:     header,
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    r,
	}
}

type streamingBehaviorRoundTripper func(*http.Request) (*http.Response, error)

func (f streamingBehaviorRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
