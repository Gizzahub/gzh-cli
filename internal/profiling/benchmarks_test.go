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

func TestDefaultBenchmarkOptions(t *testing.T) {
	opts := DefaultBenchmarkOptions()

	assert.Equal(t, 1000, opts.Iterations)
	assert.Equal(t, time.Duration(0), opts.Duration)
	assert.Equal(t, 100, opts.WarmupRuns)
	assert.Equal(t, 1, opts.Concurrency)
	assert.True(t, opts.MemoryProfiling)
	assert.False(t, opts.CPUProfiling)
}

func TestNewBenchmarkSuite(t *testing.T) {
	profiler := NewProfiler(nil)
	suite := NewBenchmarkSuite(profiler)

	assert.NotNil(t, suite)
	assert.Equal(t, profiler, suite.profiler)
	assert.NotNil(t, suite.logger)
	assert.NotNil(t, suite.results)
	assert.Len(t, suite.results, 0)
}

func TestBenchmarkSuite_RunBenchmark_NilOptions(t *testing.T) {
	suite := NewBenchmarkSuite(nil)

	executed := false
	result, err := suite.RunBenchmark(context.Background(), "test-benchmark", func(ctx context.Context) error {
		executed = true
		return nil
	}, nil)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, executed)
	assert.Equal(t, "test-benchmark", result.Name)
	assert.Equal(t, 1000, result.Operations) // Default iterations
	assert.Greater(t, result.Duration, time.Duration(0))
}

func TestBenchmarkSuite_RunBenchmark_CustomOptions(t *testing.T) {
	suite := NewBenchmarkSuite(nil)

	opts := &BenchmarkOptions{
		Iterations:      10,
		WarmupRuns:      2,
		Concurrency:     1,
		MemoryProfiling: false,
		CPUProfiling:    false,
	}

	executed := 0
	result, err := suite.RunBenchmark(context.Background(), "custom-benchmark", func(ctx context.Context) error {
		executed++
		time.Sleep(1 * time.Millisecond)
		return nil
	}, opts)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 12, executed) // 2 warmup + 10 iterations
	assert.Equal(t, "custom-benchmark", result.Name)
	assert.Equal(t, 10, result.Operations)
	assert.Greater(t, result.Duration, 10*time.Millisecond)
	assert.Greater(t, result.NsPerOp, int64(0))
	assert.Greater(t, result.OpsPerSec, float64(0))
}

func TestBenchmarkSuite_RunBenchmark_WithDuration(t *testing.T) {
	suite := NewBenchmarkSuite(nil)

	opts := &BenchmarkOptions{
		Iterations:      10000, // High number, but duration should limit it
		Duration:        50 * time.Millisecond,
		WarmupRuns:      0,
		Concurrency:     1,
		MemoryProfiling: false,
		CPUProfiling:    false,
	}

	executed := 0
	result, err := suite.RunBenchmark(context.Background(), "duration-benchmark", func(ctx context.Context) error {
		executed++
		time.Sleep(2 * time.Millisecond)
		return nil
	}, opts)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Less(t, executed, 1000) // Should be limited by duration, not iterations
	assert.Equal(t, "duration-benchmark", result.Name)
	assert.Greater(t, result.Duration, 50*time.Millisecond)
	assert.LessOrEqual(t, result.Duration, 100*time.Millisecond) // Should not be much longer
}

func TestBenchmarkSuite_RunBenchmark_Concurrent(t *testing.T) {
	suite := NewBenchmarkSuite(nil)

	opts := &BenchmarkOptions{
		Iterations:      100,
		WarmupRuns:      0,
		Concurrency:     4,
		MemoryProfiling: false,
		CPUProfiling:    false,
	}

	executed := 0
	executedMutex := make(chan struct{}, 1)

	result, err := suite.RunBenchmark(context.Background(), "concurrent-benchmark", func(ctx context.Context) error {
		// Thread-safe increment
		executedMutex <- struct{}{}
		executed++
		<-executedMutex

		time.Sleep(1 * time.Millisecond)
		return nil
	}, opts)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 100, executed)
	assert.Equal(t, "concurrent-benchmark", result.Name)
	assert.Equal(t, 100, result.Operations)

	// Concurrent execution should be faster than sequential
	expectedSequentialTime := 100 * time.Millisecond
	assert.Less(t, result.Duration, expectedSequentialTime)
}

func TestBenchmarkSuite_RunBenchmark_WithErrors(t *testing.T) {
	suite := NewBenchmarkSuite(nil)

	opts := &BenchmarkOptions{
		Iterations:      10,
		WarmupRuns:      0,
		Concurrency:     1,
		MemoryProfiling: false,
		CPUProfiling:    false,
	}

	executed := 0
	result, err := suite.RunBenchmark(context.Background(), "error-benchmark", func(ctx context.Context) error {
		executed++
		if executed%2 == 0 {
			return errors.New("test error")
		}
		return nil
	}, opts)

	assert.NoError(t, err) // Benchmark itself should not fail
	assert.NotNil(t, result)
	assert.Equal(t, 10, executed)
	assert.Equal(t, "error-benchmark", result.Name)
	assert.Equal(t, 5, result.Operations) // Only successful operations counted
}

func TestBenchmarkSuite_RunBenchmark_ContextCancellation(t *testing.T) {
	suite := NewBenchmarkSuite(nil)

	opts := &BenchmarkOptions{
		Iterations:      1000,
		WarmupRuns:      0,
		Concurrency:     1,
		MemoryProfiling: false,
		CPUProfiling:    false,
	}

	ctx, cancel := context.WithCancel(context.Background())

	executed := 0
	result, err := suite.RunBenchmark(ctx, "cancelled-benchmark", func(ctx context.Context) error {
		executed++
		if executed == 5 {
			cancel() // Cancel after 5 iterations
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		time.Sleep(1 * time.Millisecond)
		return nil
	}, opts)

	assert.NoError(t, err) // Benchmark itself should not fail
	assert.NotNil(t, result)
	assert.Less(t, result.Operations, 1000)     // Should be cancelled early
	assert.LessOrEqual(t, result.Operations, 5) // At most 5 successful operations
}

func TestBenchmarkSuite_RunBenchmark_WithProfiling(t *testing.T) {
	config := &ProfileConfig{
		Enabled:   true,
		OutputDir: "tmp/test_profiles",
	}
	profiler := NewProfiler(config)
	suite := NewBenchmarkSuite(profiler)

	opts := &BenchmarkOptions{
		Iterations:      10,
		WarmupRuns:      0,
		Concurrency:     1,
		MemoryProfiling: true,
		CPUProfiling:    false, // CPU profiling is more complex to test
	}

	result, err := suite.RunBenchmark(context.Background(), "profiled-benchmark", func(ctx context.Context) error {
		data := make([]byte, 1024) // Allocate some memory
		for i := range data {
			data[i] = byte(i % 256)
		}
		return nil
	}, opts)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "profiled-benchmark", result.Name)
	assert.Equal(t, 10, result.Operations)
	assert.Greater(t, result.AllocBytesPerOp, int64(0)) // Should show memory allocation
}

func TestBenchmarkSuite_GetResults(t *testing.T) {
	suite := NewBenchmarkSuite(nil)

	// Initially no results
	results := suite.GetResults()
	assert.Len(t, results, 0)

	// Run some benchmarks
	opts := &BenchmarkOptions{
		Iterations:      5,
		WarmupRuns:      0,
		Concurrency:     1,
		MemoryProfiling: false,
		CPUProfiling:    false,
	}

	_, err := suite.RunBenchmark(context.Background(), "benchmark1", func(ctx context.Context) error {
		return nil
	}, opts)
	require.NoError(t, err)

	_, err = suite.RunBenchmark(context.Background(), "benchmark2", func(ctx context.Context) error {
		return nil
	}, opts)
	require.NoError(t, err)

	// Should have 2 results
	results = suite.GetResults()
	assert.Len(t, results, 2)
	assert.Equal(t, "benchmark1", results[0].Name)
	assert.Equal(t, "benchmark2", results[1].Name)
}

func TestBenchmarkSuite_CompareBenchmarks_NilInputs(t *testing.T) {
	suite := NewBenchmarkSuite(nil)

	// Test with nil baseline
	comparison := suite.CompareBenchmarks(nil, &BenchmarkResult{})
	assert.Nil(t, comparison)

	// Test with nil current
	comparison = suite.CompareBenchmarks(&BenchmarkResult{}, nil)
	assert.Nil(t, comparison)

	// Test with both nil
	comparison = suite.CompareBenchmarks(nil, nil)
	assert.Nil(t, comparison)
}

func TestBenchmarkSuite_CompareBenchmarks_ValidInputs(t *testing.T) {
	suite := NewBenchmarkSuite(nil)

	baseline := &BenchmarkResult{
		Name:            "baseline",
		NsPerOp:         1000,
		AllocBytesPerOp: 512,
		AllocsPerOp:     10,
	}

	current := &BenchmarkResult{
		Name:            "current",
		NsPerOp:         800, // 20% faster
		AllocBytesPerOp: 256, // 50% less memory
		AllocsPerOp:     8,   // 20% fewer allocations
	}

	comparison := suite.CompareBenchmarks(baseline, current)

	assert.NotNil(t, comparison)
	assert.Equal(t, "baseline", comparison["baseline_name"])
	assert.Equal(t, "current", comparison["current_name"])
	assert.Contains(t, comparison, "timestamp")

	// Performance comparison (current is faster)
	speedupRatio, ok := comparison["speedup_ratio"].(float64)
	assert.True(t, ok)
	assert.InDelta(t, 1.25, speedupRatio, 0.01) // 1000/800 = 1.25

	perfChange, ok := comparison["performance_change_percent"].(float64)
	assert.True(t, ok)
	assert.InDelta(t, 25.0, perfChange, 0.1) // 25% improvement

	// Memory comparison (current uses less memory)
	memoryRatio, ok := comparison["memory_ratio"].(float64)
	assert.True(t, ok)
	assert.InDelta(t, 0.5, memoryRatio, 0.01) // 256/512 = 0.5

	memoryChange, ok := comparison["memory_change_percent"].(float64)
	assert.True(t, ok)
	assert.InDelta(t, -50.0, memoryChange, 0.1) // 50% reduction

	// Allocation comparison
	allocRatio, ok := comparison["alloc_ratio"].(float64)
	assert.True(t, ok)
	assert.InDelta(t, 0.8, allocRatio, 0.01) // 8/10 = 0.8

	allocChange, ok := comparison["alloc_change_percent"].(float64)
	assert.True(t, ok)
	assert.InDelta(t, -20.0, allocChange, 0.1) // 20% reduction
}

func TestCalculatePercentiles(t *testing.T) {
	// Test with empty slice
	percentiles := calculatePercentiles([]time.Duration{})
	assert.Len(t, percentiles, 0)

	// Test with single value
	durations := []time.Duration{100 * time.Millisecond}
	percentiles = calculatePercentiles(durations)
	assert.Len(t, percentiles, 6)
	assert.Equal(t, 100*time.Millisecond, percentiles["p50"])
	assert.Equal(t, 100*time.Millisecond, percentiles["min"])
	assert.Equal(t, 100*time.Millisecond, percentiles["max"])

	// Test with multiple values
	durations = []time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
		30 * time.Millisecond,
		40 * time.Millisecond,
		50 * time.Millisecond,
		60 * time.Millisecond,
		70 * time.Millisecond,
		80 * time.Millisecond,
		90 * time.Millisecond,
		100 * time.Millisecond,
	}
	percentiles = calculatePercentiles(durations)

	assert.Equal(t, 10*time.Millisecond, percentiles["min"])
	assert.Equal(t, 100*time.Millisecond, percentiles["max"])
	assert.Equal(t, 60*time.Millisecond, percentiles["p50"])  // 50th percentile of 10 values is at index 5 (60ms)
	assert.Equal(t, 100*time.Millisecond, percentiles["p90"]) // 90th percentile of 10 values is at index 9 (100ms)
	assert.Equal(t, 100*time.Millisecond, percentiles["p95"]) // 95th percentile of 10 values is at index 9 (100ms)
	assert.Equal(t, 100*time.Millisecond, percentiles["p99"]) // 99th percentile of 10 values is at index 9 (100ms)
}

func TestBenchmarkResult_Structure(t *testing.T) {
	now := time.Now()
	percentiles := map[string]time.Duration{
		"p50": 50 * time.Millisecond,
		"p90": 90 * time.Millisecond,
		"p95": 95 * time.Millisecond,
		"p99": 99 * time.Millisecond,
		"min": 10 * time.Millisecond,
		"max": 100 * time.Millisecond,
	}

	result := &BenchmarkResult{
		Name:             "test-benchmark",
		Operations:       1000,
		Duration:         time.Second,
		NsPerOp:          1000000,
		OpsPerSec:        1000.0,
		AllocBytesPerOp:  512,
		AllocsPerOp:      10,
		MemoryBefore:     1024,
		MemoryAfter:      2048,
		GoroutinesBefore: 5,
		GoroutinesAfter:  7,
		Percentiles:      percentiles,
		Timestamp:        now,
	}

	assert.Equal(t, "test-benchmark", result.Name)
	assert.Equal(t, 1000, result.Operations)
	assert.Equal(t, time.Second, result.Duration)
	assert.Equal(t, int64(1000000), result.NsPerOp)
	assert.Equal(t, 1000.0, result.OpsPerSec)
	assert.Equal(t, int64(512), result.AllocBytesPerOp)
	assert.Equal(t, int64(10), result.AllocsPerOp)
	assert.Equal(t, uint64(1024), result.MemoryBefore)
	assert.Equal(t, uint64(2048), result.MemoryAfter)
	assert.Equal(t, 5, result.GoroutinesBefore)
	assert.Equal(t, 7, result.GoroutinesAfter)
	assert.Equal(t, percentiles, result.Percentiles)
	assert.Equal(t, now, result.Timestamp)
}

func TestBenchmarkOptions_Structure(t *testing.T) {
	opts := &BenchmarkOptions{
		Iterations:      500,
		Duration:        30 * time.Second,
		WarmupRuns:      50,
		Concurrency:     2,
		MemoryProfiling: true,
		CPUProfiling:    false,
	}

	assert.Equal(t, 500, opts.Iterations)
	assert.Equal(t, 30*time.Second, opts.Duration)
	assert.Equal(t, 50, opts.WarmupRuns)
	assert.Equal(t, 2, opts.Concurrency)
	assert.True(t, opts.MemoryProfiling)
	assert.False(t, opts.CPUProfiling)
}

func TestBenchmarkSuite_PrintResults_Empty(t *testing.T) {
	suite := NewBenchmarkSuite(nil)

	// Should not panic with empty results
	suite.PrintResults()

	results := suite.GetResults()
	assert.Len(t, results, 0)
}

func TestBenchmarkSuite_PrintResults_WithData(t *testing.T) {
	suite := NewBenchmarkSuite(nil)

	// Add some mock results
	suite.mu.Lock()
	suite.results = append(suite.results, BenchmarkResult{
		Name:            "test-benchmark",
		Operations:      100,
		Duration:        time.Second,
		NsPerOp:         10000000,
		OpsPerSec:       100.0,
		AllocBytesPerOp: 256,
		AllocsPerOp:     5,
		Percentiles: map[string]time.Duration{
			"p50": 10 * time.Millisecond,
			"p90": 18 * time.Millisecond,
			"p95": 19 * time.Millisecond,
			"p99": 20 * time.Millisecond,
			"min": 5 * time.Millisecond,
			"max": 20 * time.Millisecond,
		},
	})
	suite.mu.Unlock()

	// Should not panic with results
	suite.PrintResults()

	results := suite.GetResults()
	assert.Len(t, results, 1)
}
