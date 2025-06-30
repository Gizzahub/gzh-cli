package github

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// RateLimiter handles GitHub API rate limiting with retry logic
type RateLimiter struct {
	mu         sync.Mutex
	limit      int
	remaining  int
	resetTime  time.Time
	retryAfter time.Duration
	maxRetries int
}

// NewRateLimiter creates a new rate limiter with default settings
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		limit:      5000, // GitHub's default API rate limit
		remaining:  5000,
		resetTime:  time.Now().Add(1 * time.Hour),
		maxRetries: 3,
	}
}

// Wait blocks until rate limit allows making a request
func (rl *RateLimiter) Wait(ctx context.Context) error {
	rl.mu.Lock()

	// Check if we need to wait for retry-after
	if rl.retryAfter > 0 {
		waitDuration := rl.retryAfter
		rl.retryAfter = 0 // Reset after use
		rl.mu.Unlock()

		if err := sleep(ctx, waitDuration); err != nil {
			return err
		}

		rl.mu.Lock()
	}

	// Check rate limit
	if rl.remaining <= 0 && time.Now().Before(rl.resetTime) {
		waitDuration := time.Until(rl.resetTime)
		rl.mu.Unlock()

		if err := sleep(ctx, waitDuration); err != nil {
			return err
		}

		rl.mu.Lock()
	}

	rl.remaining--
	rl.mu.Unlock()
	return nil
}

// Update updates rate limit information from response headers
func (rl *RateLimiter) Update(resp *http.Response) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Update remaining requests
	if remaining := resp.Header.Get("X-RateLimit-Remaining"); remaining != "" {
		if r, err := strconv.Atoi(remaining); err == nil {
			rl.remaining = r
		}
	}

	// Update limit
	if limit := resp.Header.Get("X-RateLimit-Limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			rl.limit = l
		}
	}

	// Update reset time
	if reset := resp.Header.Get("X-RateLimit-Reset"); reset != "" {
		if r, err := strconv.ParseInt(reset, 10, 64); err == nil {
			rl.resetTime = time.Unix(r, 0)
		}
	}

	// Check for Retry-After header (used for secondary rate limits)
	if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
		if seconds, err := strconv.Atoi(retryAfter); err == nil {
			rl.retryAfter = time.Duration(seconds) * time.Second
		}
	}
}

// SetRetryAfter sets the retry-after duration
func (rl *RateLimiter) SetRetryAfter(duration time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.retryAfter = duration
}

// GetStatus returns current rate limit status
func (rl *RateLimiter) GetStatus() (int, int, time.Time) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	return rl.remaining, rl.limit, rl.resetTime
}

// CalculateBackoff calculates exponential backoff with jitter
func CalculateBackoff(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}

	// Base backoff: 2^attempt seconds
	backoff := time.Duration(1<<uint(attempt)) * time.Second

	// Cap at 60 seconds
	if backoff > 60*time.Second {
		backoff = 60 * time.Second
	}

	// Add jitter (10% of backoff)
	jitter := time.Duration(rand.Float64() * float64(backoff) * 0.1)
	return backoff + jitter
}

// ShouldRetry determines if a response indicates we should retry
func ShouldRetry(resp *http.Response) bool {
	if resp == nil {
		return false
	}

	// Retry on rate limit errors
	if resp.StatusCode == http.StatusTooManyRequests {
		return true
	}

	// Retry on server errors
	if resp.StatusCode >= 500 && resp.StatusCode < 600 {
		return true
	}

	// Check for specific GitHub headers that indicate temporary issues
	if resp.Header.Get("X-GitHub-Request-Id") != "" {
		// Secondary rate limit hit
		if resp.StatusCode == http.StatusForbidden &&
			resp.Header.Get("X-RateLimit-Remaining") != "0" {
			return true
		}
	}

	return false
}

// RetryableError represents an error that can be retried
type RetryableError struct {
	Err           error
	RetryAfter    time.Duration
	AttemptsLeft  int
	NextRetryTime time.Time
}

func (e *RetryableError) Error() string {
	if e.RetryAfter > 0 {
		return fmt.Sprintf("%v (retry after %v, %d attempts left)",
			e.Err, e.RetryAfter, e.AttemptsLeft)
	}
	return fmt.Sprintf("%v (%d attempts left)", e.Err, e.AttemptsLeft)
}

// IsRetryable returns true if the error is retryable
func (e *RetryableError) IsRetryable() bool {
	return e.AttemptsLeft > 0
}

// sleep is a context-aware sleep function
func sleep(ctx context.Context, duration time.Duration) error {
	select {
	case <-time.After(duration):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
