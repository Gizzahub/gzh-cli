// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package gitlab

import (
	"context"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestStreamingClientConsumesRateLimitHeaders(t *testing.T) {
	for _, status := range []int{http.StatusOK, http.StatusTooManyRequests} {
		t.Run(strconv.Itoa(status), func(t *testing.T) {
			// 오류 응답의 헤더도 보존하고 실제 reset 시각까지 기다리지 않는다.
			reset := time.Now().Add(time.Hour).Truncate(time.Second)
			requests := 0
			client := NewStreamingClient("", "https://gitlab.invalid", StreamingConfig{})
			client.httpClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				requests++
				header := make(http.Header)
				header.Set("RateLimit-Remaining", "0")
				header.Set("RateLimit-Limit", "50")
				header.Set("RateLimit-Reset", strconv.FormatInt(reset.Unix(), 10))
				return testHTTPResponse(r, status, `[]`, header), nil
			})}
			_, _, err := client.fetchProjectPage(context.Background(), "team", CursorPagination{Page: 1, PerPage: 10})
			if status == http.StatusOK {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, "429 Too Many Requests")
			}
			require.Equal(t, 1, requests)
			client.rateLimiter.mu.RLock()
			remaining, limit, observedReset := client.rateLimiter.remaining, client.rateLimiter.limit, client.rateLimiter.reset
			client.rateLimiter.mu.RUnlock()
			require.Zero(t, remaining)
			require.Equal(t, 50, limit)
			require.True(t, reset.Equal(observedReset))
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			require.ErrorIs(t, client.waitForRateLimit(ctx, 0), context.Canceled)
		})
	}
}

func TestStreamingClientInvalidRateLimitHeadersPreserveState(t *testing.T) {
	for _, value := range []string{"", "not-a-number"} {
		t.Run(value, func(t *testing.T) {
			client := NewStreamingClient("", "https://gitlab.invalid", StreamingConfig{})
			reset := time.Unix(1234, 0)
			client.rateLimiter.remaining = 7
			client.rateLimiter.limit = 20
			client.rateLimiter.reset = reset
			client.httpClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				header := make(http.Header)
				for _, key := range []string{"RateLimit-Remaining", "RateLimit-Limit", "RateLimit-Reset"} {
					header.Set(key, value)
				}
				return testHTTPResponse(r, http.StatusOK, `[]`, header), nil
			})}
			_, _, err := client.fetchProjectPage(context.Background(), "team", CursorPagination{Page: 1, PerPage: 10})
			require.NoError(t, err)
			require.Equal(t, 7, client.rateLimiter.remaining)
			require.Equal(t, 20, client.rateLimiter.limit)
			require.True(t, reset.Equal(client.rateLimiter.reset))
		})
	}
}

func TestStreamingClientRateLimitedResponseStopsWithoutRetry(t *testing.T) {
	// 429을 자동 재시도하는 새로운 정책을 도입하지 않는다.
	requests := 0
	client := NewStreamingClient("", "https://gitlab.invalid", StreamingConfig{})
	client.httpClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		requests++
		header := make(http.Header)
		header.Set("RateLimit-Remaining", "0")
		header.Set("RateLimit-Reset", strconv.FormatInt(time.Now().Add(time.Hour).Unix(), 10))
		return testHTTPResponse(r, http.StatusTooManyRequests, "limited", header), nil
	})}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	stream, err := client.StreamGroupProjects(ctx, "team", StreamingConfig{
		PageSize: 1, BufferSize: 1, MemoryLimit: 1024, RetryAttempts: 3,
	})
	require.NoError(t, err)
	select {
	case item, ok := <-stream:
		require.True(t, ok)
		require.ErrorContains(t, item.Error, "429 Too Many Requests")
	case <-ctx.Done():
		t.Fatal("rate-limited response was not returned")
	}
	select {
	case _, ok := <-stream:
		require.False(t, ok)
	case <-ctx.Done():
		t.Fatal("stream did not close after HTTP error")
	}
	require.Equal(t, 1, requests)
}
