// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package profiling

import (
	"context"
	"time"

	"github.com/gizzahub/gzh-manager-go/internal/logger"
)

// PerformanceMiddleware provides performance monitoring capabilities.
type PerformanceMiddleware struct {
	profiler *Profiler
	logger   *logger.SimpleLogger
	enabled  bool
}

// NewPerformanceMiddleware creates a new performance middleware.
func NewPerformanceMiddleware(profiler *Profiler, enabled bool) *PerformanceMiddleware {
	return &PerformanceMiddleware{
		profiler: profiler,
		logger:   logger.NewSimpleLogger("perf-middleware"),
		enabled:  enabled,
	}
}

// OperationMetrics holds performance metrics for an operation.
type OperationMetrics struct {
	Name             string
	StartTime        time.Time
	EndTime          time.Time
	Duration         time.Duration
	GoroutinesBefore int
	GoroutinesAfter  int
	MemoryBefore     uint64
	MemoryAfter      uint64
	Success          bool
	Error            error
}

// TrackOperation wraps an operation with performance tracking.
func (pm *PerformanceMiddleware) TrackOperation(_ context.Context, operationName string, operation func() error) error {
	if !pm.enabled {
		return operation()
	}

	metrics := &OperationMetrics{
		Name:      operationName,
		StartTime: time.Now(),
	}

	// Capture initial runtime stats
	if pm.profiler != nil {
		stats := pm.profiler.GetRuntimeStats()
		if goroutines, ok := stats["goroutines"].(int); ok {
			metrics.GoroutinesBefore = goroutines
		}
		if memory, ok := stats["memory"].(map[string]interface{}); ok {
			if alloc, ok := memory["alloc_bytes"].(uint64); ok {
				metrics.MemoryBefore = alloc
			}
		}
	}

	// Execute operation
	err := operation()

	metrics.EndTime = time.Now()
	metrics.Duration = metrics.EndTime.Sub(metrics.StartTime)
	metrics.Success = err == nil
	metrics.Error = err

	// Capture final runtime stats
	if pm.profiler != nil {
		stats := pm.profiler.GetRuntimeStats()
		if goroutines, ok := stats["goroutines"].(int); ok {
			metrics.GoroutinesAfter = goroutines
		}
		if memory, ok := stats["memory"].(map[string]interface{}); ok {
			if alloc, ok := memory["alloc_bytes"].(uint64); ok {
				metrics.MemoryAfter = alloc
			}
		}
	}

	// Log performance metrics
	pm.logMetrics(metrics)

	return err
}

// TrackOperationWithProfiling wraps an operation with performance tracking and profiling.
func (pm *PerformanceMiddleware) TrackOperationWithProfiling(ctx context.Context, operationName string, profileTypes []ProfileType, operation func() error) error {
	if !pm.enabled || pm.profiler == nil {
		return pm.TrackOperation(ctx, operationName, operation)
	}

	return pm.profiler.ProfileOperation(ctx, operationName, profileTypes, func() error {
		return pm.TrackOperation(ctx, operationName, operation)
	})
}

// logMetrics logs the performance metrics.
func (pm *PerformanceMiddleware) logMetrics(metrics *OperationMetrics) {
	goroutineDelta := metrics.GoroutinesAfter - metrics.GoroutinesBefore
	var memoryDelta int64
	if metrics.MemoryAfter >= metrics.MemoryBefore {
		memoryDelta = int64(metrics.MemoryAfter - metrics.MemoryBefore)
	} else {
		memoryDelta = -int64(metrics.MemoryBefore - metrics.MemoryAfter)
	}

	logLevel := "Info"
	if !metrics.Success {
		logLevel = "Error"
	} else if metrics.Duration > 5*time.Second {
		logLevel = "Warn" // Long operations get warning level
	}

	performanceData := map[string]interface{}{
		"duration_ms":        metrics.Duration.Milliseconds(),
		"goroutine_delta":    goroutineDelta,
		"memory_delta_bytes": memoryDelta,
		"success":            metrics.Success,
	}

	if metrics.Error != nil {
		performanceData["error"] = metrics.Error.Error()
	}

	switch logLevel {
	case "Error":
		pm.logger.LogPerformance(metrics.Name+"_failed", metrics.Duration, performanceData)
	case "Warn":
		pm.logger.LogPerformance(metrics.Name+"_slow", metrics.Duration, performanceData)
	default:
		pm.logger.LogPerformance(metrics.Name, metrics.Duration, performanceData)
	}
}

// WrapFunction creates a performance-tracked version of a function.
func (pm *PerformanceMiddleware) WrapFunction(operationName string, fn func() error) func() error {
	return func() error {
		return pm.TrackOperation(context.Background(), operationName, fn)
	}
}

// WrapFunctionWithContext creates a performance-tracked version of a function with context.
func (pm *PerformanceMiddleware) WrapFunctionWithContext(operationName string, fn func(context.Context) error) func(context.Context) error {
	return func(ctx context.Context) error {
		return pm.TrackOperation(ctx, operationName, func() error {
			return fn(ctx)
		})
	}
}

// BatchOperationTracker tracks performance metrics for batch operations.
type BatchOperationTracker struct {
	middleware    *PerformanceMiddleware
	operationName string
	batchSize     int
	startTime     time.Time
	completed     int
	failed        int
}

// NewBatchOperationTracker creates a new batch operation tracker.
func (pm *PerformanceMiddleware) NewBatchOperationTracker(operationName string, batchSize int) *BatchOperationTracker {
	return &BatchOperationTracker{
		middleware:    pm,
		operationName: operationName,
		batchSize:     batchSize,
		startTime:     time.Now(),
	}
}

// TrackItem tracks completion of a single item in the batch.
func (bot *BatchOperationTracker) TrackItem(success bool) {
	if success {
		bot.completed++
	} else {
		bot.failed++
	}
}

// Finish completes the batch operation tracking.
func (bot *BatchOperationTracker) Finish() {
	if !bot.middleware.enabled {
		return
	}

	duration := time.Since(bot.startTime)
	total := bot.completed + bot.failed
	successRate := float64(bot.completed) / float64(total) * 100

	bot.middleware.logger.LogPerformance(bot.operationName+"_batch", duration, map[string]interface{}{
		"batch_size":    bot.batchSize,
		"completed":     bot.completed,
		"failed":        bot.failed,
		"total":         total,
		"success_rate":  successRate,
		"items_per_sec": float64(total) / duration.Seconds(),
	})
}
