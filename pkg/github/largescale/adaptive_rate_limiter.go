package largescale

import (
	"context"
	"sync"
	"time"
)

// AdaptiveRateLimiter provides intelligent rate limiting for GitHub API operations.
type AdaptiveRateLimiter struct {
	mu                sync.RWMutex
	remaining         int
	resetTime         time.Time
	lastRequest       time.Time
	requestHistory    []time.Time
	backoffMultiplier float64
	maxBackoff        time.Duration

	// Configuration
	maxRequestsPerSecond int
	bufferRatio          float64 // Keep this ratio of requests as buffer
	adaptiveDelay        bool
}

// NewAdaptiveRateLimiter creates a new adaptive rate limiter.
func NewAdaptiveRateLimiter() *AdaptiveRateLimiter {
	return &AdaptiveRateLimiter{
		remaining:            5000, // GitHub default
		resetTime:            time.Now().Add(time.Hour),
		requestHistory:       make([]time.Time, 0, 100),
		backoffMultiplier:    1.5,
		maxBackoff:           time.Minute * 5,
		maxRequestsPerSecond: 10,  // Conservative default
		bufferRatio:          0.1, // Keep 10% as buffer
		adaptiveDelay:        true,
	}
}

// Wait blocks until it's safe to make a request.
func (rl *AdaptiveRateLimiter) Wait(ctx context.Context) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Clean old history (keep last 100 requests)
	rl.cleanHistory(now)

	// Calculate delay based on current state
	delay := rl.calculateDelay(now)

	if delay > 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}

	// Record this request
	rl.lastRequest = now
	rl.requestHistory = append(rl.requestHistory, now)

	return nil
}

// UpdateRemaining updates the remaining request count from API response.
func (rl *AdaptiveRateLimiter) UpdateRemaining(remaining int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.remaining = remaining

	// Adapt behavior based on remaining requests
	if remaining < 100 {
		// Very low, be more conservative
		rl.maxRequestsPerSecond = 2
		rl.bufferRatio = 0.05 // 5% buffer
	} else if remaining < 500 {
		// Low, reduce rate
		rl.maxRequestsPerSecond = 5
		rl.bufferRatio = 0.08 // 8% buffer
	} else {
		// Normal operation
		rl.maxRequestsPerSecond = 10
		rl.bufferRatio = 0.1 // 10% buffer
	}
}

// UpdateResetTime updates the rate limit reset time from API response.
func (rl *AdaptiveRateLimiter) UpdateResetTime(resetTime time.Time) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.resetTime = resetTime
}

// GetStatus returns current rate limiter status.
func (rl *AdaptiveRateLimiter) GetStatus() (remaining int, resetTime time.Time, estimatedDelay time.Duration) {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	return rl.remaining, rl.resetTime, rl.calculateDelay(time.Now())
}

// calculateDelay determines how long to wait before next request.
func (rl *AdaptiveRateLimiter) calculateDelay(now time.Time) time.Duration {
	// If we're past reset time, no delay needed
	if now.After(rl.resetTime) {
		return 0
	}

	// Calculate time until reset
	timeUntilReset := rl.resetTime.Sub(now)

	// If no remaining requests, wait until reset
	if rl.remaining <= 0 {
		return timeUntilReset
	}

	// Keep a buffer of requests
	bufferRequests := int(float64(rl.remaining) * rl.bufferRatio)
	effectiveRemaining := rl.remaining - bufferRequests

	if effectiveRemaining <= 0 {
		// Use buffer requests very slowly
		return timeUntilReset / time.Duration(maxInt(1, bufferRequests))
	}

	// Calculate delay based on remaining requests and time
	baseDelay := timeUntilReset / time.Duration(effectiveRemaining)

	// Apply adaptive adjustments
	if rl.adaptiveDelay {
		// Consider recent request frequency
		recentFrequency := rl.calculateRecentFrequency(now)
		if recentFrequency > float64(rl.maxRequestsPerSecond) {
			// We're going too fast, increase delay
			baseDelay = time.Duration(float64(baseDelay) * rl.backoffMultiplier)
		}

		// Cap the delay
		if baseDelay > rl.maxBackoff {
			baseDelay = rl.maxBackoff
		}

		// Minimum delay between requests
		minDelay := time.Second / time.Duration(rl.maxRequestsPerSecond)
		if baseDelay < minDelay {
			baseDelay = minDelay
		}
	}

	// Consider time since last request
	timeSinceLastRequest := now.Sub(rl.lastRequest)
	if timeSinceLastRequest < baseDelay {
		return baseDelay - timeSinceLastRequest
	}

	return 0
}

// calculateRecentFrequency calculates requests per second for recent history.
func (rl *AdaptiveRateLimiter) calculateRecentFrequency(now time.Time) float64 {
	if len(rl.requestHistory) < 2 {
		return 0
	}

	// Look at requests in the last 10 seconds
	lookback := time.Second * 10
	cutoff := now.Add(-lookback)

	recentRequests := 0

	for i := len(rl.requestHistory) - 1; i >= 0; i-- {
		if rl.requestHistory[i].After(cutoff) {
			recentRequests++
		} else {
			break
		}
	}

	if recentRequests < 2 {
		return 0
	}

	// Calculate frequency
	timeSpan := now.Sub(rl.requestHistory[len(rl.requestHistory)-recentRequests])
	if timeSpan <= 0 {
		return 0
	}

	return float64(recentRequests) / timeSpan.Seconds()
}

// cleanHistory removes old entries from request history.
func (rl *AdaptiveRateLimiter) cleanHistory(now time.Time) {
	// Keep only last 100 requests or requests from last hour
	cutoff := now.Add(-time.Hour)

	// Find first entry to keep
	keepFrom := 0

	for i, reqTime := range rl.requestHistory {
		if reqTime.After(cutoff) {
			keepFrom = i
			break
		}

		keepFrom = i + 1
	}

	// Keep only recent entries, but at most 100
	if keepFrom > 0 || len(rl.requestHistory) > 100 {
		if len(rl.requestHistory) > 100 {
			// Keep last 100
			keepFrom = maxInt(keepFrom, len(rl.requestHistory)-100)
		}

		copy(rl.requestHistory, rl.requestHistory[keepFrom:])
		rl.requestHistory = rl.requestHistory[:len(rl.requestHistory)-keepFrom]
	}
}

// EstimateTimeToCompletion estimates how long it will take to make N requests.
func (rl *AdaptiveRateLimiter) EstimateTimeToCompletion(requestsNeeded int) time.Duration {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	now := time.Now()

	// If we have enough remaining requests
	if rl.remaining >= requestsNeeded {
		// Estimate based on current rate limiting
		avgDelay := rl.calculateDelay(now)
		return time.Duration(requestsNeeded) * avgDelay
	}

	// Need to wait for reset and possibly multiple cycles
	timeUntilReset := rl.resetTime.Sub(now)
	if timeUntilReset <= 0 {
		timeUntilReset = 0
	}

	// Requests we can make with current remaining
	requestsFromCurrent := min(rl.remaining, requestsNeeded)
	remainingNeeded := requestsNeeded - requestsFromCurrent

	// Time for current batch
	currentBatchTime := time.Duration(requestsFromCurrent) * rl.calculateDelay(now)

	if remainingNeeded <= 0 {
		return currentBatchTime
	}

	// Additional cycles needed (5000 requests per hour)
	additionalCycles := (remainingNeeded + 4999) / 5000 // Ceiling division
	additionalTime := time.Duration(additionalCycles) * time.Hour

	return currentBatchTime + timeUntilReset + additionalTime
}

// SetConfiguration allows customizing rate limiter behavior.
func (rl *AdaptiveRateLimiter) SetConfiguration(maxPerSecond int, bufferRatio float64, enableAdaptive bool) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.maxRequestsPerSecond = maxPerSecond
	rl.bufferRatio = bufferRatio
	rl.adaptiveDelay = enableAdaptive
}

// Reset resets the rate limiter state (useful for testing or manual reset).
func (rl *AdaptiveRateLimiter) Reset() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.remaining = 5000
	rl.resetTime = time.Now().Add(time.Hour)
	rl.requestHistory = rl.requestHistory[:0]
	rl.lastRequest = time.Time{}
}

// Helper function for rate limiter.
func maxInt(a, b int) int {
	if a > b {
		return a
	}

	return b
}
