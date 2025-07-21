// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package api

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// deduplication, batching, and intelligent rate limiting.
type OptimizationManager struct {
	deduplicator *RequestDeduplicator
	batcher      *BatchProcessor
	rateLimitMgr *SharedRateLimiterManager
	stats        OptimizationStats
	mu           sync.RWMutex
	config       OptimizationConfig
	enabled      bool
}

// OptimizationStats aggregates performance metrics across all optimization features.
type OptimizationStats struct {
	StartTime             time.Time
	TotalRequests         int64
	DeduplicatedRequests  int64
	BatchedRequests       int64
	RateLimitedRequests   int64
	TotalTimeSaved        time.Duration
	AverageResponseTime   time.Duration
	OverallEfficiencyGain float64
	LastUpdated           time.Time
}

// OptimizationConfig configures the optimization behavior.
type OptimizationConfig struct {
	EnableDeduplication bool
	EnableBatching      bool
	EnableRateLimit     bool
	DeduplicationTTL    time.Duration
	BatchConfig         BatchConfig
	MetricsInterval     time.Duration
}

// DefaultOptimizationConfig returns sensible defaults for API optimization.
func DefaultOptimizationConfig() OptimizationConfig {
	return OptimizationConfig{
		EnableDeduplication: true,
		EnableBatching:      true,
		EnableRateLimit:     true,
		DeduplicationTTL:    5 * time.Minute,
		BatchConfig:         DefaultBatchConfig(),
		MetricsInterval:     30 * time.Second,
	}
}

// NewOptimizationManager creates a new API optimization manager.
func NewOptimizationManager(config OptimizationConfig) *OptimizationManager {
	om := &OptimizationManager{
		config:  config,
		enabled: true,
		stats: OptimizationStats{
			StartTime:   time.Now(),
			LastUpdated: time.Now(),
		},
	}

	// Initialize components based on configuration
	if config.EnableDeduplication {
		om.deduplicator = NewRequestDeduplicator(config.DeduplicationTTL)
	}

	if config.EnableBatching {
		om.batcher = NewBatchProcessor(config.BatchConfig)
	}

	if config.EnableRateLimit {
		om.rateLimitMgr = NewSharedRateLimiterManager()
	}

	return om
}

// OptimizedRequest represents a request with optimization metadata.
type OptimizedRequest struct {
	Service   string
	Operation string
	Key       string
	Data      interface{}
	Context   context.Context
}

// OptimizedResponse contains the response and optimization metadata.
type OptimizedResponse struct {
	Data              interface{}
	Error             error
	WasDeduplicateded bool
	WasBatched        bool
	WasRateLimited    bool
	TimeSaved         time.Duration
	ResponseTime      time.Duration
}

// ExecuteRequest executes a request with all available optimizations.
func (om *OptimizationManager) ExecuteRequest(req OptimizedRequest, executor RequestFunc) (*OptimizedResponse, error) {
	if !om.enabled {
		// Execute directly without optimizations
		start := time.Now()
		result, err := executor(req.Context)

		return &OptimizedResponse{
			Data:         result,
			Error:        err,
			ResponseTime: time.Since(start),
		}, err
	}

	start := time.Now()
	response := &OptimizedResponse{}

	// Update stats
	om.mu.Lock()
	om.stats.TotalRequests++
	om.stats.LastUpdated = time.Now()
	om.mu.Unlock()

	// Apply rate limiting if enabled
	if om.config.EnableRateLimit && om.rateLimitMgr != nil {
		limiter := om.rateLimitMgr.GetLimiter(req.Service)
		if err := limiter.Wait(req.Context); err != nil {
			response.Error = fmt.Errorf("rate limit wait failed: %w", err)
			return response, response.Error
		}

		response.WasRateLimited = true

		om.mu.Lock()
		om.stats.RateLimitedRequests++
		om.mu.Unlock()
	}

	// Apply deduplication if enabled
	if om.config.EnableDeduplication && om.deduplicator != nil {
		deduplicationKey := GenerateKey(req.Service, req.Operation, req.Key)

		dedupeStart := time.Now()
		result, err := om.deduplicator.Do(req.Context, deduplicationKey, executor)
		dedupeTime := time.Since(dedupeStart)

		response.Data = result
		response.Error = err
		response.ResponseTime = dedupeTime

		// Check if this was a deduplicated call
		if dedupeTime < 10*time.Millisecond { // Likely deduplicated
			response.WasDeduplicateded = true
			response.TimeSaved = dedupeTime

			om.mu.Lock()
			om.stats.DeduplicatedRequests++
			om.stats.TotalTimeSaved += response.TimeSaved
			om.mu.Unlock()
		}
	} else {
		// Execute request directly
		result, err := executor(req.Context)
		response.Data = result
		response.Error = err
	}

	// Update overall metrics
	response.ResponseTime = time.Since(start)
	om.updateAverageResponseTime(response.ResponseTime)

	return response, response.Error
}

// ExecuteBatchRequest executes a batch of similar requests.
func (om *OptimizationManager) ExecuteBatchRequest(ctx context.Context, batchKey string, requests []*BatchRequest, processor BatchFunc) error {
	if !om.enabled || !om.config.EnableBatching || om.batcher == nil {
		// Execute requests individually
		responses := processor(ctx, requests)
		for i, req := range requests {
			if i < len(responses) {
				req.Response <- responses[i]
			} else {
				req.Response <- BatchResponse{
					ID:    req.ID,
					Error: fmt.Errorf("no response generated for request %s", req.ID),
				}
			}

			close(req.Response)
		}

		return nil
	}

	// Add requests to batch processor
	for _, req := range requests {
		err := om.batcher.Add(ctx, batchKey, req, processor)
		if err != nil {
			req.Response <- BatchResponse{
				ID:    req.ID,
				Error: fmt.Errorf("failed to add to batch: %w", err),
			}
			close(req.Response)
		}
	}

	om.mu.Lock()
	om.stats.BatchedRequests += int64(len(requests))
	om.mu.Unlock()

	return nil
}

// updateAverageResponseTime updates the running average response time.
func (om *OptimizationManager) updateAverageResponseTime(responseTime time.Duration) {
	om.mu.Lock()
	defer om.mu.Unlock()

	if om.stats.AverageResponseTime == 0 {
		om.stats.AverageResponseTime = responseTime
	} else {
		// Exponential moving average
		alpha := 0.1
		om.stats.AverageResponseTime = time.Duration(
			alpha*float64(responseTime) + (1-alpha)*float64(om.stats.AverageResponseTime),
		)
	}
}

// GetStats returns current optimization statistics.
func (om *OptimizationManager) GetStats() OptimizationStats {
	om.mu.RLock()
	defer om.mu.RUnlock()

	stats := om.stats

	// Calculate overall efficiency gain
	if stats.TotalRequests > 0 {
		optimizedRequests := stats.DeduplicatedRequests + stats.BatchedRequests
		stats.OverallEfficiencyGain = float64(optimizedRequests) / float64(stats.TotalRequests)
	}

	return stats
}

// Enable enables optimization features.
func (om *OptimizationManager) Enable() {
	om.mu.Lock()
	defer om.mu.Unlock()
	om.enabled = true
}

// Disable disables optimization features (useful for debugging).
func (om *OptimizationManager) Disable() {
	om.mu.Lock()
	defer om.mu.Unlock()
	om.enabled = false
}

// IsEnabled returns whether optimizations are currently enabled.
func (om *OptimizationManager) IsEnabled() bool {
	om.mu.RLock()
	defer om.mu.RUnlock()

	return om.enabled
}

// ClearCaches clears all optimization caches.
func (om *OptimizationManager) ClearCaches() {
	if om.deduplicator != nil {
		om.deduplicator.Clear()
	}

	om.mu.Lock()
	om.stats = OptimizationStats{
		StartTime:   time.Now(),
		LastUpdated: time.Now(),
	}
	om.mu.Unlock()
}

// Stop gracefully stops all optimization components.
func (om *OptimizationManager) Stop() {
	if om.deduplicator != nil {
		om.deduplicator.Close()
	}

	if om.batcher != nil {
		om.batcher.Stop()
	}

	if om.rateLimitMgr != nil {
		om.rateLimitMgr.Stop()
	}
}

// PrintDetailedStats prints comprehensive optimization statistics.
func (om *OptimizationManager) PrintDetailedStats() {
	stats := om.GetStats()
	uptime := time.Since(stats.StartTime)

	fmt.Printf("=== API Optimization Manager Statistics ===\n")
	fmt.Printf("Uptime: %v\n", uptime)
	fmt.Printf("Status: %s\n", map[bool]string{true: "Enabled", false: "Disabled"}[om.enabled])
	fmt.Printf("Total Requests: %d\n", stats.TotalRequests)
	fmt.Printf("Deduplicated Requests: %d\n", stats.DeduplicatedRequests)
	fmt.Printf("Batched Requests: %d\n", stats.BatchedRequests)
	fmt.Printf("Rate Limited Requests: %d\n", stats.RateLimitedRequests)
	fmt.Printf("Total Time Saved: %v\n", stats.TotalTimeSaved)
	fmt.Printf("Average Response Time: %v\n", stats.AverageResponseTime)
	fmt.Printf("Overall Efficiency Gain: %.2f%%\n", stats.OverallEfficiencyGain*100)
	fmt.Printf("Last Updated: %v\n", stats.LastUpdated)

	fmt.Printf("\n=== Component Statistics ===\n")

	if om.deduplicator != nil {
		fmt.Printf("\n--- Request Deduplication ---\n")
		om.deduplicator.PrintStats()
	}

	if om.batcher != nil {
		fmt.Printf("\n--- Batch Processing ---\n")
		om.batcher.PrintStats()
	}

	if om.rateLimitMgr != nil {
		fmt.Printf("\n--- Rate Limiting ---\n")
		om.rateLimitMgr.PrintAllStats()
	}
}

// GetRateLimiter returns the rate limiter for a specific service.
func (om *OptimizationManager) GetRateLimiter(service string) *EnhancedRateLimiter {
	if om.rateLimitMgr == nil {
		return nil
	}

	return om.rateLimitMgr.GetLimiter(service)
}

// GetBatchProcessor returns the batch processor.
func (om *OptimizationManager) GetBatchProcessor() *BatchProcessor {
	return om.batcher
}

// GetDeduplicator returns the request deduplicator.
func (om *OptimizationManager) GetDeduplicator() *RequestDeduplicator {
	return om.deduplicator
}

// UpdateConfiguration updates the optimization configuration.
func (om *OptimizationManager) UpdateConfiguration(config OptimizationConfig) {
	om.mu.Lock()
	defer om.mu.Unlock()

	om.config = config

	// Recreate components if configuration changed significantly
	if config.EnableDeduplication && om.deduplicator == nil {
		om.deduplicator = NewRequestDeduplicator(config.DeduplicationTTL)
	} else if !config.EnableDeduplication && om.deduplicator != nil {
		om.deduplicator.Close()
		om.deduplicator = nil
	}
	// Similar logic for other components...
}

// GlobalOptimizationManager provides a global instance for application-wide optimization.
var GlobalOptimizationManager = NewOptimizationManager(DefaultOptimizationConfig())

// GetGlobalOptimizer returns the global optimization manager instance.
func GetGlobalOptimizer() *OptimizationManager {
	return GlobalOptimizationManager
}

// InitializeOptimization initializes global API optimization with custom configuration.
func InitializeOptimization(config OptimizationConfig) {
	if GlobalOptimizationManager != nil {
		GlobalOptimizationManager.Stop()
	}

	GlobalOptimizationManager = NewOptimizationManager(config)
}
