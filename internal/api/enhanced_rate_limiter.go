package api

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"
)

// EnhancedRateLimiter provides intelligent rate limiting with adaptive behavior.
type EnhancedRateLimiter struct {
	mu                sync.RWMutex
	service           string
	limit             int
	remaining         int
	resetTime         time.Time
	retryAfter        time.Duration
	maxRetries        int
	adaptiveMode      bool
	requestQueue      chan *rateLimitedRequest
	stats             RateLimiterStats
	windowHistory     []RateLimitWindow
	backoffMultiplier float64
	minBackoff        time.Duration
	maxBackoff        time.Duration
	stopCh            chan struct{}
	wg                sync.WaitGroup
}

// RateLimiterStats tracks rate limiter performance and behavior.
type RateLimiterStats struct {
	TotalRequests     int64
	ThrottledRequests int64
	AdaptiveAdjusts   int64
	AverageWaitTime   time.Duration
	LastReset         time.Time
	EfficiencyRate    float64
}

// RateLimitWindow tracks rate limit information over time.
type RateLimitWindow struct {
	Timestamp time.Time
	Limit     int
	Remaining int
	Reset     time.Time
}

// rateLimitedRequest represents a queued request waiting for rate limit clearance.
type rateLimitedRequest struct {
	ctx        context.Context
	responseCh chan error
	timestamp  time.Time
}

// RateLimiterConfig configures the enhanced rate limiter.
type RateLimiterConfig struct {
	Service           string
	InitialLimit      int
	MaxRetries        int
	AdaptiveMode      bool
	BackoffMultiplier float64
	MinBackoff        time.Duration
	MaxBackoff        time.Duration
	QueueSize         int
}

// ServiceRateLimits defines rate limits for different services.
var ServiceRateLimits = map[string]RateLimiterConfig{
	"github": {
		Service:           "github",
		InitialLimit:      5000, // 5000 requests per hour
		MaxRetries:        3,
		AdaptiveMode:      true,
		BackoffMultiplier: 2.0,
		MinBackoff:        time.Second,
		MaxBackoff:        time.Minute * 15,
		QueueSize:         100,
	},
	"gitlab": {
		Service:           "gitlab",
		InitialLimit:      2000, // 2000 requests per minute
		MaxRetries:        3,
		AdaptiveMode:      true,
		BackoffMultiplier: 1.5,
		MinBackoff:        500 * time.Millisecond,
		MaxBackoff:        time.Minute * 5,
		QueueSize:         50,
	},
	"gitea": {
		Service:           "gitea",
		InitialLimit:      1000, // Conservative default, varies by instance
		MaxRetries:        2,
		AdaptiveMode:      true,
		BackoffMultiplier: 1.8,
		MinBackoff:        time.Second,
		MaxBackoff:        time.Minute * 10,
		QueueSize:         30,
	},
}

// NewEnhancedRateLimiter creates a new enhanced rate limiter for the specified service.
func NewEnhancedRateLimiter(service string) *EnhancedRateLimiter {
	config, exists := ServiceRateLimits[service]
	if !exists {
		config = ServiceRateLimits["gitea"] // Use conservative defaults
		config.Service = service
	}

	rl := &EnhancedRateLimiter{
		service:           config.Service,
		limit:             config.InitialLimit,
		remaining:         config.InitialLimit,
		resetTime:         time.Now().Add(time.Hour), // Default hour window
		maxRetries:        config.MaxRetries,
		adaptiveMode:      config.AdaptiveMode,
		requestQueue:      make(chan *rateLimitedRequest, config.QueueSize),
		windowHistory:     make([]RateLimitWindow, 0, 100),
		backoffMultiplier: config.BackoffMultiplier,
		minBackoff:        config.MinBackoff,
		maxBackoff:        config.MaxBackoff,
		stopCh:            make(chan struct{}),
	}

	// Start request processor
	rl.wg.Add(1)

	go rl.processRequests()

	return rl
}

// Wait blocks until a request can be made according to rate limit rules.
func (rl *EnhancedRateLimiter) Wait(ctx context.Context) error {
	rl.mu.Lock()
	rl.stats.TotalRequests++
	rl.mu.Unlock()

	// Check if we can proceed immediately
	if rl.canProceed() {
		return nil
	}

	// Queue the request
	req := &rateLimitedRequest{
		ctx:        ctx,
		responseCh: make(chan error, 1),
		timestamp:  time.Now(),
	}

	select {
	case rl.requestQueue <- req:
		// Request queued, wait for response
		select {
		case err := <-req.responseCh:
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	case <-ctx.Done():
		return ctx.Err()
	}
}

// UpdateLimits updates rate limit information from API response headers.
func (rl *EnhancedRateLimiter) UpdateLimits(limit, remaining int, resetTime time.Time) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Record historical data
	window := RateLimitWindow{
		Timestamp: time.Now(),
		Limit:     limit,
		Remaining: remaining,
		Reset:     resetTime,
	}
	rl.windowHistory = append(rl.windowHistory, window)

	// Keep only last 100 windows for analysis
	if len(rl.windowHistory) > 100 {
		rl.windowHistory = rl.windowHistory[1:]
	}

	oldLimit := rl.limit
	rl.limit = limit
	rl.remaining = remaining
	rl.resetTime = resetTime
	rl.stats.LastReset = time.Now()

	// Adaptive adjustment
	if rl.adaptiveMode && limit != oldLimit {
		rl.stats.AdaptiveAdjusts++
		rl.adaptBackoffStrategy()
	}

	// Update efficiency metrics
	rl.updateEfficiencyMetrics()
}

// SetRetryAfter sets the retry-after duration from API response.
func (rl *EnhancedRateLimiter) SetRetryAfter(duration time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.retryAfter = duration
}

// canProceed checks if a request can proceed immediately.
func (rl *EnhancedRateLimiter) canProceed() bool {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	now := time.Now()

	// Check if rate limit has reset
	if now.After(rl.resetTime) {
		rl.mu.RUnlock()
		rl.mu.Lock()
		rl.remaining = rl.limit
		rl.resetTime = now.Add(time.Hour) // Default reset window
		rl.mu.Unlock()
		rl.mu.RLock()
	}

	// Check retry-after constraint
	if rl.retryAfter > 0 {
		return false
	}

	// Check remaining quota
	return rl.remaining > 0
}

// processRequests handles queued requests based on rate limit availability.
func (rl *EnhancedRateLimiter) processRequests() {
	defer rl.wg.Done()

	ticker := time.NewTicker(100 * time.Millisecond) // Check every 100ms
	defer ticker.Stop()

	for {
		select {
		case <-rl.stopCh:
			return
		case <-ticker.C:
			rl.processQueuedRequests()
		case req := <-rl.requestQueue:
			if rl.canProceed() {
				rl.approveRequest(req)
			} else {
				// Put request back in queue and wait
				go func() {
					time.Sleep(rl.calculateBackoff())

					select {
					case rl.requestQueue <- req:
					case <-rl.stopCh:
						req.responseCh <- fmt.Errorf("rate limiter stopped")
					}
				}()
			}
		}
	}
}

// processQueuedRequests processes requests that are waiting in the queue.
func (rl *EnhancedRateLimiter) processQueuedRequests() {
	for {
		select {
		case req := <-rl.requestQueue:
			if rl.canProceed() {
				rl.approveRequest(req)
			} else {
				// Put request back and break
				go func() {
					rl.requestQueue <- req
				}()

				return
			}
		default:
			return // No more queued requests
		}
	}
}

// approveRequest approves a queued request and decrements remaining quota.
func (rl *EnhancedRateLimiter) approveRequest(req *rateLimitedRequest) {
	rl.mu.Lock()
	rl.remaining--
	waitTime := time.Since(req.timestamp)
	rl.stats.AverageWaitTime = (rl.stats.AverageWaitTime + waitTime) / 2
	rl.mu.Unlock()

	req.responseCh <- nil
}

// calculateBackoff calculates exponential backoff with jitter.
func (rl *EnhancedRateLimiter) calculateBackoff() time.Duration {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	if rl.retryAfter > 0 {
		return rl.retryAfter
	}

	// Calculate time until reset
	untilReset := time.Until(rl.resetTime)
	if untilReset <= 0 {
		return rl.minBackoff
	}

	// Adaptive backoff based on remaining quota
	quotaRatio := float64(rl.remaining) / float64(rl.limit)
	if quotaRatio > 0.5 {
		return rl.minBackoff
	}

	// Exponential backoff with jitter
	backoff := float64(rl.minBackoff) * math.Pow(rl.backoffMultiplier, 1.0-quotaRatio)
	jitter := backoff * 0.1 * (2.0*math.Max(0, math.Min(1, quotaRatio)) - 1.0)

	result := time.Duration(backoff + jitter)
	if result > rl.maxBackoff {
		result = rl.maxBackoff
	}

	return result
}

// adaptBackoffStrategy adjusts backoff parameters based on historical performance.
func (rl *EnhancedRateLimiter) adaptBackoffStrategy() {
	if len(rl.windowHistory) < 10 {
		return // Need more data for adaptation
	}

	// Analyze recent windows for patterns
	recentWindows := rl.windowHistory[len(rl.windowHistory)-10:]
	throttleCount := 0

	for _, window := range recentWindows {
		if float64(window.Remaining)/float64(window.Limit) < 0.1 { // Less than 10% remaining
			throttleCount++
		}
	}

	// Adjust strategy based on throttle frequency
	if throttleCount > 5 { // More than 50% of recent windows were throttled
		// Increase conservatism
		rl.backoffMultiplier = math.Min(rl.backoffMultiplier*1.1, 3.0)
		rl.minBackoff = time.Duration(float64(rl.minBackoff) * 1.2)
	} else if throttleCount < 2 { // Less than 20% throttled
		// Decrease conservatism
		rl.backoffMultiplier = math.Max(rl.backoffMultiplier*0.9, 1.2)
		rl.minBackoff = time.Duration(float64(rl.minBackoff) * 0.9)
	}
}

// updateEfficiencyMetrics calculates efficiency metrics.
func (rl *EnhancedRateLimiter) updateEfficiencyMetrics() {
	if rl.stats.TotalRequests == 0 {
		rl.stats.EfficiencyRate = 1.0
		return
	}

	successfulRequests := rl.stats.TotalRequests - rl.stats.ThrottledRequests
	rl.stats.EfficiencyRate = float64(successfulRequests) / float64(rl.stats.TotalRequests)
}

// GetStats returns current rate limiter statistics.
func (rl *EnhancedRateLimiter) GetStats() RateLimiterStats {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	return rl.stats
}

// GetCurrentStatus returns current rate limit status.
func (rl *EnhancedRateLimiter) GetCurrentStatus() (limit, remaining int, resetTime time.Time) {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	return rl.limit, rl.remaining, rl.resetTime
}

// Stop stops the rate limiter and cleans up resources.
func (rl *EnhancedRateLimiter) Stop() {
	close(rl.stopCh)
	rl.wg.Wait()
	close(rl.requestQueue)
}

// PrintStats prints detailed rate limiter statistics.
func (rl *EnhancedRateLimiter) PrintStats() {
	stats := rl.GetStats()
	limit, remaining, resetTime := rl.GetCurrentStatus()

	fmt.Printf("=== Rate Limiter Statistics (%s) ===\n", rl.service)
	fmt.Printf("Current Limit: %d\n", limit)
	fmt.Printf("Remaining: %d\n", remaining)
	fmt.Printf("Reset Time: %v\n", resetTime)
	fmt.Printf("Total Requests: %d\n", stats.TotalRequests)
	fmt.Printf("Throttled Requests: %d\n", stats.ThrottledRequests)
	fmt.Printf("Adaptive Adjustments: %d\n", stats.AdaptiveAdjusts)
	fmt.Printf("Average Wait Time: %v\n", stats.AverageWaitTime)
	fmt.Printf("Efficiency Rate: %.2f%%\n", stats.EfficiencyRate*100)
	fmt.Printf("Last Reset: %v\n", stats.LastReset)
}

// SharedRateLimiterManager manages multiple rate limiters for different services.
type SharedRateLimiterManager struct {
	limiters map[string]*EnhancedRateLimiter
	mu       sync.RWMutex
}

// NewSharedRateLimiterManager creates a new shared rate limiter manager.
func NewSharedRateLimiterManager() *SharedRateLimiterManager {
	return &SharedRateLimiterManager{
		limiters: make(map[string]*EnhancedRateLimiter),
	}
}

// GetLimiter returns the rate limiter for the specified service.
func (m *SharedRateLimiterManager) GetLimiter(service string) *EnhancedRateLimiter {
	m.mu.RLock()
	limiter, exists := m.limiters[service]
	m.mu.RUnlock()

	if !exists {
		m.mu.Lock()
		// Double-check pattern
		if limiter, exists = m.limiters[service]; !exists {
			limiter = NewEnhancedRateLimiter(service)
			m.limiters[service] = limiter
		}

		m.mu.Unlock()
	}

	return limiter
}

// Stop stops all rate limiters.
func (m *SharedRateLimiterManager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, limiter := range m.limiters {
		limiter.Stop()
	}
}

// PrintAllStats prints statistics for all managed rate limiters.
func (m *SharedRateLimiterManager) PrintAllStats() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for service, limiter := range m.limiters {
		fmt.Printf("\n--- %s Rate Limiter ---\n", service)
		limiter.PrintStats()
	}
}
