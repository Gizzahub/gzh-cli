// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package profiling

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPerformanceMiddleware(t *testing.T) {
	profiler := NewProfiler(nil)
	middleware := NewPerformanceMiddleware(profiler, true)

	assert.NotNil(t, middleware)
	assert.Equal(t, profiler, middleware.profiler)
	assert.True(t, middleware.enabled)
	assert.NotNil(t, middleware.logger)
}

func TestPerformanceMiddleware_TrackOperation_Disabled(t *testing.T) {
	middleware := NewPerformanceMiddleware(nil, false)

	executed := false
	err := middleware.TrackOperation(context.Background(), "test-op", func() error {
		executed = true
		return nil
	})

	assert.NoError(t, err)
	assert.True(t, executed)
}

func TestPerformanceMiddleware_TrackOperation_Success(t *testing.T) {
	profiler := NewProfiler(&ProfileConfig{Enabled: true})
	middleware := NewPerformanceMiddleware(profiler, true)

	executed := false
	startTime := time.Now()

	err := middleware.TrackOperation(context.Background(), "test-op", func() error {
		executed = true
		time.Sleep(10 * time.Millisecond)
		return nil
	})

	assert.NoError(t, err)
	assert.True(t, executed)
	assert.True(t, time.Since(startTime) >= 10*time.Millisecond)
}

func TestPerformanceMiddleware_TrackOperation_WithError(t *testing.T) {
	profiler := NewProfiler(&ProfileConfig{Enabled: true})
	middleware := NewPerformanceMiddleware(profiler, true)

	testError := errors.New("operation failed")
	executed := false

	err := middleware.TrackOperation(context.Background(), "test-op", func() error {
		executed = true
		return testError
	})

	assert.Equal(t, testError, err)
	assert.True(t, executed)
}

func TestPerformanceMiddleware_TrackOperationWithProfiling_Disabled(t *testing.T) {
	middleware := NewPerformanceMiddleware(nil, false)

	executed := false
	err := middleware.TrackOperationWithProfiling(context.Background(), "test-op", []ProfileType{ProfileTypeCPU}, func() error {
		executed = true
		return nil
	})

	assert.NoError(t, err)
	assert.True(t, executed)
}

func TestPerformanceMiddleware_TrackOperationWithProfiling_NilProfiler(t *testing.T) {
	middleware := NewPerformanceMiddleware(nil, true)

	executed := false
	err := middleware.TrackOperationWithProfiling(context.Background(), "test-op", []ProfileType{ProfileTypeCPU}, func() error {
		executed = true
		return nil
	})

	assert.NoError(t, err)
	assert.True(t, executed)
}

func TestPerformanceMiddleware_TrackOperationWithProfiling_Enabled(t *testing.T) {
	config := &ProfileConfig{
		Enabled:   true,
		OutputDir: "tmp/test_profiles",
	}
	profiler := NewProfiler(config)
	middleware := NewPerformanceMiddleware(profiler, true)

	executed := false
	err := middleware.TrackOperationWithProfiling(
		context.Background(),
		"test-op",
		[]ProfileType{ProfileTypeMemory},
		func() error {
			executed = true
			time.Sleep(10 * time.Millisecond)
			return nil
		},
	)

	assert.NoError(t, err)
	assert.True(t, executed)
}

func TestPerformanceMiddleware_WrapFunction(t *testing.T) {
	middleware := NewPerformanceMiddleware(nil, true)

	originalExecuted := false
	originalFunc := func() error {
		originalExecuted = true
		return nil
	}

	wrappedFunc := middleware.WrapFunction("wrapped-op", originalFunc)

	// Execute wrapped function
	err := wrappedFunc()

	assert.NoError(t, err)
	assert.True(t, originalExecuted)
}

func TestPerformanceMiddleware_WrapFunctionWithContext(t *testing.T) {
	middleware := NewPerformanceMiddleware(nil, true)

	originalExecuted := false
	var receivedCtx context.Context

	originalFunc := func(ctx context.Context) error {
		originalExecuted = true
		receivedCtx = ctx
		return nil
	}

	wrappedFunc := middleware.WrapFunctionWithContext("wrapped-op", originalFunc)

	// Execute wrapped function
	ctx := context.Background()
	err := wrappedFunc(ctx)

	assert.NoError(t, err)
	assert.True(t, originalExecuted)
	assert.Equal(t, ctx, receivedCtx)
}

func TestOperationMetrics_Structure(t *testing.T) {
	now := time.Now()
	testError := errors.New("test error")

	metrics := &OperationMetrics{
		Name:             "test-operation",
		StartTime:        now,
		EndTime:          now.Add(100 * time.Millisecond),
		Duration:         100 * time.Millisecond,
		GoroutinesBefore: 10,
		GoroutinesAfter:  12,
		MemoryBefore:     1000,
		MemoryAfter:      1200,
		Success:          false,
		Error:            testError,
	}

	assert.Equal(t, "test-operation", metrics.Name)
	assert.Equal(t, now, metrics.StartTime)
	assert.Equal(t, now.Add(100*time.Millisecond), metrics.EndTime)
	assert.Equal(t, 100*time.Millisecond, metrics.Duration)
	assert.Equal(t, 10, metrics.GoroutinesBefore)
	assert.Equal(t, 12, metrics.GoroutinesAfter)
	assert.Equal(t, uint64(1000), metrics.MemoryBefore)
	assert.Equal(t, uint64(1200), metrics.MemoryAfter)
	assert.False(t, metrics.Success)
	assert.Equal(t, testError, metrics.Error)
}

func TestPerformanceMiddleware_NewBatchOperationTracker(t *testing.T) {
	middleware := NewPerformanceMiddleware(nil, true)

	tracker := middleware.NewBatchOperationTracker("batch-test", 100)

	assert.NotNil(t, tracker)
	assert.Equal(t, middleware, tracker.middleware)
	assert.Equal(t, "batch-test", tracker.operationName)
	assert.Equal(t, 100, tracker.batchSize)
	assert.Equal(t, 0, tracker.completed)
	assert.Equal(t, 0, tracker.failed)
	assert.WithinDuration(t, time.Now(), tracker.startTime, time.Second)
}

func TestBatchOperationTracker_TrackItem(t *testing.T) {
	middleware := NewPerformanceMiddleware(nil, true)
	tracker := middleware.NewBatchOperationTracker("batch-test", 10)

	// Track successful items
	tracker.TrackItem(true)
	tracker.TrackItem(true)
	tracker.TrackItem(true)

	assert.Equal(t, 3, tracker.completed)
	assert.Equal(t, 0, tracker.failed)

	// Track failed items
	tracker.TrackItem(false)
	tracker.TrackItem(false)

	assert.Equal(t, 3, tracker.completed)
	assert.Equal(t, 2, tracker.failed)
}

func TestBatchOperationTracker_Finish_Disabled(t *testing.T) {
	middleware := NewPerformanceMiddleware(nil, false)
	tracker := middleware.NewBatchOperationTracker("batch-test", 10)

	tracker.TrackItem(true)
	tracker.TrackItem(false)

	// Should not panic or cause issues when disabled
	tracker.Finish()

	assert.Equal(t, 1, tracker.completed)
	assert.Equal(t, 1, tracker.failed)
}

func TestBatchOperationTracker_Finish_Enabled(t *testing.T) {
	middleware := NewPerformanceMiddleware(nil, true)
	tracker := middleware.NewBatchOperationTracker("batch-test", 10)

	// Simulate batch processing
	for i := 0; i < 8; i++ {
		tracker.TrackItem(true)
	}
	for i := 0; i < 2; i++ {
		tracker.TrackItem(false)
	}

	// Sleep to ensure some duration
	time.Sleep(10 * time.Millisecond)

	tracker.Finish()

	assert.Equal(t, 8, tracker.completed)
	assert.Equal(t, 2, tracker.failed)
}

func TestPerformanceMiddleware_TrackOperation_ContextCancellation(t *testing.T) {
	middleware := NewPerformanceMiddleware(nil, true)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	executed := false
	err := middleware.TrackOperation(ctx, "canceled-op", func() error {
		executed = true
		return ctx.Err()
	})

	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
	assert.True(t, executed)
}

func TestPerformanceMiddleware_Integration(t *testing.T) {
	config := &ProfileConfig{
		Enabled:   true,
		OutputDir: "tmp/test_profiles",
	}
	profiler := NewProfiler(config)
	middleware := NewPerformanceMiddleware(profiler, true)

	// Test complete workflow
	executed := false
	err := middleware.TrackOperationWithProfiling(
		context.Background(),
		"integration-test",
		[]ProfileType{ProfileTypeMemory, ProfileTypeGoroutine},
		func() error {
			executed = true

			// Simulate some work that would be interesting to profile
			data := make([]byte, 1024*1024) // Allocate 1MB
			for i := range data {
				data[i] = byte(i % 256)
			}

			time.Sleep(20 * time.Millisecond)
			return nil
		},
	)

	assert.NoError(t, err)
	assert.True(t, executed)
}

func TestPerformanceMiddleware_WrapFunction_WithError(t *testing.T) {
	middleware := NewPerformanceMiddleware(nil, true)

	testError := errors.New("wrapped function error")
	originalFunc := func() error {
		return testError
	}

	wrappedFunc := middleware.WrapFunction("error-wrapped-op", originalFunc)

	err := wrappedFunc()
	assert.Equal(t, testError, err)
}

func TestPerformanceMiddleware_WrapFunctionWithContext_WithError(t *testing.T) {
	middleware := NewPerformanceMiddleware(nil, true)

	testError := errors.New("wrapped function error")
	originalFunc := func(ctx context.Context) error {
		return testError
	}

	wrappedFunc := middleware.WrapFunctionWithContext("error-wrapped-op", originalFunc)

	err := wrappedFunc(context.Background())
	assert.Equal(t, testError, err)
}

func TestBatchOperationTracker_ZeroDivision(t *testing.T) {
	middleware := NewPerformanceMiddleware(nil, true)
	tracker := middleware.NewBatchOperationTracker("empty-batch", 0)

	// Don't track any items
	tracker.Finish()

	// Should not panic on zero division
	assert.Equal(t, 0, tracker.completed)
	assert.Equal(t, 0, tracker.failed)
}

func TestPerformanceMiddleware_StatisticsCollection(t *testing.T) {
	profiler := NewProfiler(&ProfileConfig{Enabled: true})
	middleware := NewPerformanceMiddleware(profiler, true)

	// Execute multiple operations to collect statistics
	operations := []string{"op1", "op2", "op3"}

	for _, opName := range operations {
		err := middleware.TrackOperation(context.Background(), opName, func() error {
			time.Sleep(5 * time.Millisecond)
			return nil
		})
		require.NoError(t, err)
	}

	// Verify that runtime stats can be collected without errors
	stats := profiler.GetRuntimeStats()
	assert.NotNil(t, stats)
	assert.Contains(t, stats, "goroutines")
	assert.Contains(t, stats, "memory")
}
