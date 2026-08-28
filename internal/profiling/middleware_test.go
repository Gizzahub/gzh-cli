// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package profiling

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignedMemoryDelta(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		before        uint64
		after         uint64
		wantDelta     int64
		wantSaturated bool
	}{
		{name: "equal", before: 42, after: 42, wantDelta: 0},
		{name: "positive", before: 1000, after: 1200, wantDelta: 200},
		{name: "negative", before: 1200, after: 1000, wantDelta: -200},
		{name: "positive max int64", before: 0, after: math.MaxInt64, wantDelta: math.MaxInt64},
		{name: "positive overflow", before: 0, after: uint64(math.MaxInt64) + 1, wantDelta: math.MaxInt64, wantSaturated: true},
		{name: "positive max uint64", before: 0, after: math.MaxUint64, wantDelta: math.MaxInt64, wantSaturated: true},
		{name: "negative max int64", before: math.MaxInt64, after: 0, wantDelta: -math.MaxInt64},
		{name: "exact min int64", before: uint64(math.MaxInt64) + 1, after: 0, wantDelta: math.MinInt64},
		{name: "negative overflow", before: uint64(math.MaxInt64) + 2, after: 0, wantDelta: math.MinInt64, wantSaturated: true},
		{name: "negative max uint64", before: math.MaxUint64, after: 0, wantDelta: math.MinInt64, wantSaturated: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotDelta, gotSaturated := signedMemoryDelta(tt.before, tt.after)

			assert.Equal(t, tt.wantDelta, gotDelta)
			assert.Equal(t, tt.wantSaturated, gotSaturated)
		})
	}
}

func TestPerformanceLogData_MemoryDeltaSaturation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		before        uint64
		after         uint64
		wantDelta     int64
		wantSaturated bool
	}{
		{name: "normal delta", before: 100, after: 150, wantDelta: 50},
		{name: "positive saturation", before: 0, after: math.MaxUint64, wantDelta: math.MaxInt64, wantSaturated: true},
		{name: "exact minimum", before: uint64(math.MaxInt64) + 1, after: 0, wantDelta: math.MinInt64},
		{name: "negative saturation", before: math.MaxUint64, after: 0, wantDelta: math.MinInt64, wantSaturated: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data := performanceLogData(&OperationMetrics{
				MemoryBefore: tt.before,
				MemoryAfter:  tt.after,
			})

			assert.Equal(t, tt.wantDelta, data["memory_delta_bytes"])
			if tt.wantSaturated {
				assert.Equal(t, true, data["memory_delta_saturated"])
			} else {
				assert.NotContains(t, data, "memory_delta_saturated")
			}
		})
	}

	t.Run("preserves existing metrics", func(t *testing.T) {
		t.Parallel()

		operationErr := errors.New("operation failed")
		data := performanceLogData(&OperationMetrics{
			Duration:         1500 * time.Millisecond,
			GoroutinesBefore: 2,
			GoroutinesAfter:  5,
			MemoryBefore:     100,
			MemoryAfter:      125,
			Success:          false,
			Error:            operationErr,
		})

		assert.Equal(t, int64(1500), data["duration_ms"])
		assert.Equal(t, 3, data["goroutine_delta"])
		assert.Equal(t, int64(25), data["memory_delta_bytes"])
		assert.Equal(t, false, data["success"])
		assert.Equal(t, operationErr.Error(), data["error"])
		assert.NotContains(t, data, "memory_delta_saturated")
	})
}

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
		OutputDir: t.TempDir(),
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
	for range 8 {
		tracker.TrackItem(true)
	}
	for range 2 {
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
		OutputDir: t.TempDir(),
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
