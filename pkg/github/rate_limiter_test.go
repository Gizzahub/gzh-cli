package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRateLimiter_Wait(t *testing.T) {
	tests := []struct {
		name         string
		setupLimiter func() *RateLimiter
		expectError  bool
		expectWait   bool
		waitDuration time.Duration
	}{
		{
			name: "no wait needed",
			setupLimiter: func() *RateLimiter {
				rl := NewRateLimiter()
				rl.remaining = 100
				return rl
			},
			expectError: false,
			expectWait:  false,
		},
		{
			name: "wait for rate limit reset",
			setupLimiter: func() *RateLimiter {
				rl := NewRateLimiter()
				rl.remaining = 0
				rl.resetTime = time.Now().Add(100 * time.Millisecond)
				return rl
			},
			expectError:  false,
			expectWait:   true,
			waitDuration: 100 * time.Millisecond,
		},
		{
			name: "wait for retry after",
			setupLimiter: func() *RateLimiter {
				rl := NewRateLimiter()
				rl.remaining = 100
				rl.retryAfter = 50 * time.Millisecond
				return rl
			},
			expectError:  false,
			expectWait:   true,
			waitDuration: 50 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rl := tt.setupLimiter()
			ctx := context.Background()

			start := time.Now()
			err := rl.Wait(ctx)
			elapsed := time.Since(start)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.expectWait {
				assert.GreaterOrEqual(t, elapsed, tt.waitDuration)
			} else {
				assert.Less(t, elapsed, 10*time.Millisecond)
			}
		})
	}
}

func TestRateLimiter_Wait_ContextCancellation(t *testing.T) {
	rl := NewRateLimiter()
	rl.remaining = 0
	rl.resetTime = time.Now().Add(1 * time.Hour) // Long wait

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel context after short delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	err := rl.Wait(ctx)
	elapsed := time.Since(start)

	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
	assert.Less(t, elapsed, 100*time.Millisecond)
}

func TestRateLimiter_Update(t *testing.T) {
	rl := NewRateLimiter()

	// Create test response with headers
	resp := &http.Response{
		Header: make(http.Header),
	}
	resp.Header.Set("X-RateLimit-Remaining", "42")
	resp.Header.Set("X-RateLimit-Limit", "5000")
	resp.Header.Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(30*time.Minute).Unix(), 10))
	resp.Header.Set("Retry-After", "60")

	rl.Update(resp)

	remaining, limit, resetTime := rl.GetStatus()
	assert.Equal(t, 42, remaining)
	assert.Equal(t, 5000, limit)
	assert.WithinDuration(t, time.Now().Add(30*time.Minute), resetTime, 1*time.Second)
	assert.Equal(t, 60*time.Second, rl.retryAfter)
}

func TestCalculateBackoff(t *testing.T) {
	tests := []struct {
		attempt     int
		minExpected time.Duration
		maxExpected time.Duration
	}{
		{0, 1 * time.Second, 1100 * time.Millisecond},
		{1, 2 * time.Second, 2200 * time.Millisecond},
		{2, 4 * time.Second, 4400 * time.Millisecond},
		{3, 8 * time.Second, 8800 * time.Millisecond},
		{10, 60 * time.Second, 66 * time.Second}, // Should cap at 60s
	}

	for _, tt := range tests {
		t.Run(strconv.Itoa(tt.attempt), func(t *testing.T) {
			backoff := CalculateBackoff(tt.attempt)
			assert.GreaterOrEqual(t, backoff, tt.minExpected)
			assert.LessOrEqual(t, backoff, tt.maxExpected)
		})
	}
}

func TestShouldRetry(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		headers    map[string]string
		expected   bool
	}{
		{
			name:       "rate limit error",
			statusCode: http.StatusTooManyRequests,
			expected:   true,
		},
		{
			name:       "server error",
			statusCode: http.StatusInternalServerError,
			expected:   true,
		},
		{
			name:       "bad gateway",
			statusCode: http.StatusBadGateway,
			expected:   true,
		},
		{
			name:       "success",
			statusCode: http.StatusOK,
			expected:   false,
		},
		{
			name:       "client error",
			statusCode: http.StatusBadRequest,
			expected:   false,
		},
		{
			name:       "secondary rate limit",
			statusCode: http.StatusForbidden,
			headers: map[string]string{
				"X-GitHub-Request-Id":   "ABC123",
				"X-RateLimit-Remaining": "100",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{
				StatusCode: tt.statusCode,
				Header:     make(http.Header),
			}
			for k, v := range tt.headers {
				resp.Header.Set(k, v)
			}

			result := ShouldRetry(resp)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRetryableError(t *testing.T) {
	err := &RetryableError{
		Err:          fmt.Errorf("test error"),
		RetryAfter:   5 * time.Second,
		AttemptsLeft: 2,
	}

	assert.Contains(t, err.Error(), "test error")
	assert.Contains(t, err.Error(), "retry after 5s")
	assert.Contains(t, err.Error(), "2 attempts left")
	assert.True(t, err.IsRetryable())

	// Test with no retry after
	err2 := &RetryableError{
		Err:          fmt.Errorf("another error"),
		AttemptsLeft: 0,
	}

	assert.Contains(t, err2.Error(), "another error")
	assert.NotContains(t, err2.Error(), "retry after")
	assert.False(t, err2.IsRetryable())
}

func TestRepoConfigClient_MakeRequestWithRetry(t *testing.T) {
	// Test successful request after retry
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++

		// Set rate limit headers
		w.Header().Set("X-RateLimit-Remaining", "100")
		w.Header().Set("X-RateLimit-Limit", "5000")
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(1*time.Hour).Unix(), 10))

		if attempts < 2 {
			// First attempt: rate limit error
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"message": "API rate limit exceeded",
			})
		} else {
			// Second attempt: success
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"name": "test-repo",
			})
		}
	}))
	defer server.Close()

	client := NewRepoConfigClient("test-token")
	client.baseURL = server.URL
	client.httpClient.Timeout = 5 * time.Second

	ctx := context.Background()
	resp, err := client.makeRequest(ctx, "GET", "/test", nil)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 2, attempts)
}

func TestRepoConfigClient_MakeRequestMaxRetries(t *testing.T) {
	// Test max retries exceeded
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"message": "Internal server error",
		})
	}))
	defer server.Close()

	client := NewRepoConfigClient("test-token")
	client.baseURL = server.URL

	ctx := context.Background()
	_, err := client.makeRequest(ctx, "GET", "/test", nil)

	assert.Error(t, err)
	// After retries, it returns an APIError, not a RetryableError
	var apiErr *APIError
	assert.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusInternalServerError, apiErr.StatusCode)
	assert.Equal(t, 4, attempts) // Initial + 3 retries
}
