package api

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

// RequestDeduplicator provides request deduplication using singleflight pattern
// This prevents multiple concurrent requests for the same resource.
type RequestDeduplicator struct {
	group   singleflight.Group
	stats   DeduplicationStats
	mu      sync.RWMutex
	ttl     time.Duration
	cache   map[string]cachedResult
	cleanup chan struct{}
}

// DeduplicationStats tracks deduplication performance metrics.
type DeduplicationStats struct {
	TotalRequests     int64
	DeduplicatedCalls int64
	CacheHits         int64
	CacheMisses       int64
	TotalSavings      time.Duration
	LastUpdated       time.Time
}

// cachedResult stores cached response with expiration.
type cachedResult struct {
	value   interface{}
	err     error
	expires time.Time
}

// RequestFunc represents a function that performs an API request.
type RequestFunc func(ctx context.Context) (interface{}, error)

// NewRequestDeduplicator creates a new request deduplicator with the given TTL.
func NewRequestDeduplicator(ttl time.Duration) *RequestDeduplicator {
	d := &RequestDeduplicator{
		ttl:     ttl,
		cache:   make(map[string]cachedResult),
		cleanup: make(chan struct{}),
	}

	// Start cleanup goroutine
	go d.cleanupExpired()

	return d
}

// Do executes a request with deduplication, returning cached results for duplicate requests.
func (d *RequestDeduplicator) Do(ctx context.Context, key string, fn RequestFunc) (interface{}, error) {
	d.mu.Lock()
	d.stats.TotalRequests++
	d.stats.LastUpdated = time.Now()
	d.mu.Unlock()

	// Check cache first
	if result, found := d.getCached(key); found {
		d.mu.Lock()
		d.stats.CacheHits++
		d.mu.Unlock()

		return result.value, result.err
	}

	d.mu.Lock()
	d.stats.CacheMisses++
	d.mu.Unlock()

	// Use singleflight to deduplicate concurrent requests
	start := time.Now()
	result, err, shared := d.group.Do(key, func() (interface{}, error) {
		return fn(ctx)
	})

	if shared {
		d.mu.Lock()
		d.stats.DeduplicatedCalls++
		d.stats.TotalSavings += time.Since(start)
		d.mu.Unlock()
	}

	// Cache the result if TTL is set
	if d.ttl > 0 {
		d.setCached(key, result, err)
	}

	return result, err
}

// DoWithTimeout executes a request with deduplication and timeout.
func (d *RequestDeduplicator) DoWithTimeout(ctx context.Context, key string, timeout time.Duration, fn RequestFunc) (interface{}, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return d.Do(timeoutCtx, key, fn)
}

// getCached retrieves a cached result if it exists and hasn't expired.
func (d *RequestDeduplicator) getCached(key string) (cachedResult, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	result, exists := d.cache[key]
	if !exists {
		return cachedResult{}, false
	}

	if time.Now().After(result.expires) {
		// Result has expired, remove it
		delete(d.cache, key)
		return cachedResult{}, false
	}

	return result, true
}

// setCached stores a result in the cache with expiration.
func (d *RequestDeduplicator) setCached(key string, value interface{}, err error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.cache[key] = cachedResult{
		value:   value,
		err:     err,
		expires: time.Now().Add(d.ttl),
	}
}

// GetStats returns current deduplication statistics.
func (d *RequestDeduplicator) GetStats() DeduplicationStats {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.stats
}

// Clear clears all cached results and resets statistics.
func (d *RequestDeduplicator) Clear() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.cache = make(map[string]cachedResult)
	d.stats = DeduplicationStats{LastUpdated: time.Now()}
	d.group = singleflight.Group{} // Reset singleflight group
}

// Close stops the deduplicator and cleans up resources.
func (d *RequestDeduplicator) Close() {
	close(d.cleanup)
	d.Clear()
}

// cleanupExpired periodically removes expired cache entries.
func (d *RequestDeduplicator) cleanupExpired() {
	ticker := time.NewTicker(d.ttl / 2) // Cleanup at half the TTL interval
	defer ticker.Stop()

	for {
		select {
		case <-d.cleanup:
			return
		case <-ticker.C:
			d.performCleanup()
		}
	}
}

// performCleanup removes expired entries from cache.
func (d *RequestDeduplicator) performCleanup() {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now()
	for key, result := range d.cache {
		if now.After(result.expires) {
			delete(d.cache, key)
		}
	}
}

// ForgetKey removes a specific key from cache and singleflight group.
func (d *RequestDeduplicator) ForgetKey(key string) {
	d.mu.Lock()
	delete(d.cache, key)
	d.mu.Unlock()

	// Forget the key in singleflight group
	d.group.Forget(key)
}

// GenerateKey creates a standardized cache key from components.
func GenerateKey(service, operation string, params ...string) string {
	key := fmt.Sprintf("%s:%s", service, operation)
	for _, param := range params {
		key += ":" + param
	}

	return key
}

// GetEfficiencyRate calculates the efficiency of deduplication (0.0 to 1.0).
func (d *RequestDeduplicator) GetEfficiencyRate() float64 {
	stats := d.GetStats()
	if stats.TotalRequests == 0 {
		return 0.0
	}

	savedRequests := stats.DeduplicatedCalls + stats.CacheHits

	return float64(savedRequests) / float64(stats.TotalRequests)
}

// PrintStats prints detailed deduplication statistics.
func (d *RequestDeduplicator) PrintStats() {
	stats := d.GetStats()
	efficiency := d.GetEfficiencyRate()

	fmt.Printf("=== Request Deduplication Statistics ===\n")
	fmt.Printf("Total Requests: %d\n", stats.TotalRequests)
	fmt.Printf("Deduplicated Calls: %d\n", stats.DeduplicatedCalls)
	fmt.Printf("Cache Hits: %d\n", stats.CacheHits)
	fmt.Printf("Cache Misses: %d\n", stats.CacheMisses)
	fmt.Printf("Total Time Saved: %v\n", stats.TotalSavings)
	fmt.Printf("Efficiency Rate: %.2f%%\n", efficiency*100)
	fmt.Printf("Last Updated: %v\n", stats.LastUpdated)
}
