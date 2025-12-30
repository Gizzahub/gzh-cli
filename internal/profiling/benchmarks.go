// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package profiling

import (
	"context"
	"fmt"
	"runtime"
	"sort"
	"sync"
	"time"

	"github.com/gizzahub/gzh-cli/internal/logger"
)

// BenchmarkResult holds the results of a benchmark run.
type BenchmarkResult struct {
	Name             string                   `json:"name"`
	Operations       int                      `json:"operations"`
	Duration         time.Duration            `json:"duration"`
	NsPerOp          int64                    `json:"ns_per_op"`
	OpsPerSec        float64                  `json:"ops_per_sec"`
	AllocBytesPerOp  int64                    `json:"alloc_bytes_per_op"`
	AllocsPerOp      int64                    `json:"allocs_per_op"`
	MemoryBefore     uint64                   `json:"memory_before"`
	MemoryAfter      uint64                   `json:"memory_after"`
	GoroutinesBefore int                      `json:"goroutines_before"`
	GoroutinesAfter  int                      `json:"goroutines_after"`
	Percentiles      map[string]time.Duration `json:"percentiles"`
	Timestamp        time.Time                `json:"timestamp"`
}

// BenchmarkSuite manages and runs performance benchmarks.
type BenchmarkSuite struct {
	profiler *Profiler
	logger   *logger.SimpleLogger
	results  []BenchmarkResult
	mu       sync.RWMutex
}

// NewBenchmarkSuite creates a new benchmark suite.
func NewBenchmarkSuite(profiler *Profiler) *BenchmarkSuite {
	return &BenchmarkSuite{
		profiler: profiler,
		logger:   logger.NewSimpleLogger("benchmark"),
		results:  make([]BenchmarkResult, 0),
	}
}

// BenchmarkOptions configures benchmark execution.
type BenchmarkOptions struct {
	Iterations      int           `json:"iterations"`
	Duration        time.Duration `json:"duration"`
	WarmupRuns      int           `json:"warmup_runs"`
	Concurrency     int           `json:"concurrency"`
	MemoryProfiling bool          `json:"memory_profiling"`
	CPUProfiling    bool          `json:"cpu_profiling"`
}

// DefaultBenchmarkOptions returns default benchmark options.
func DefaultBenchmarkOptions() *BenchmarkOptions {
	return &BenchmarkOptions{
		Iterations:      1000,
		Duration:        0,
		WarmupRuns:      100,
		Concurrency:     1,
		MemoryProfiling: true,
		CPUProfiling:    false,
	}
}

// BenchmarkFunc represents a function to be benchmarked.
type BenchmarkFunc func(ctx context.Context) error

// SimpleBenchmarkFunc represents a simple function to be benchmarked (for compatibility).
type SimpleBenchmarkFunc func(ctx context.Context)

// RunBenchmark executes a benchmark with the given options.
func (bs *BenchmarkSuite) RunBenchmark(ctx context.Context, name string, fn BenchmarkFunc, opts *BenchmarkOptions) (*BenchmarkResult, error) {
	if opts == nil {
		opts = DefaultBenchmarkOptions()
	}

	bs.logger.Info("Starting benchmark", "name", name, "iterations", opts.Iterations, "concurrency", opts.Concurrency)

	// Warmup runs
	if opts.WarmupRuns > 0 {
		bs.logger.Debug("Running warmup", "name", name, "warmup_runs", opts.WarmupRuns)
		for i := 0; i < opts.WarmupRuns; i++ {
			fn(ctx) // Warmup run, errors are ignored intentionally
		}
		runtime.GC() // Force garbage collection after warmup
	}

	// Start profiling if enabled
	var profileSessions []string
	if bs.profiler != nil {
		if opts.CPUProfiling {
			if sessionID, err := bs.profiler.StartProfile(ProfileTypeCPU); err == nil {
				profileSessions = append(profileSessions, sessionID)
			}
		}
		if opts.MemoryProfiling {
			if sessionID, err := bs.profiler.StartProfile(ProfileTypeMemory); err == nil {
				profileSessions = append(profileSessions, sessionID)
			}
		}
	}

	// Capture initial runtime stats
	var initialStats runtime.MemStats
	runtime.ReadMemStats(&initialStats)
	initialGoroutines := runtime.NumGoroutine()

	// Run benchmark
	result, err := bs.executeBenchmark(ctx, name, fn, opts)
	if err != nil {
		return nil, err
	}

	// Capture final runtime stats
	var finalStats runtime.MemStats
	runtime.ReadMemStats(&finalStats)
	finalGoroutines := runtime.NumGoroutine()

	// Stop profiling sessions
	for _, sessionID := range profileSessions {
		if err := bs.profiler.StopProfile(sessionID); err != nil {
			bs.logger.Warn("Failed to stop profile session", "session_id", sessionID, "error", err)
		}
	}

	// Calculate memory and allocation metrics
	result.MemoryBefore = initialStats.Alloc
	result.MemoryAfter = finalStats.Alloc
	result.GoroutinesBefore = initialGoroutines
	result.GoroutinesAfter = finalGoroutines

	if result.Operations > 0 {
		allocDelta := finalStats.TotalAlloc - initialStats.TotalAlloc
		mallocsDelta := finalStats.Mallocs - initialStats.Mallocs

		result.AllocBytesPerOp = int64(allocDelta) / int64(result.Operations) //nolint:gosec // G115: 벤치마크 통계는 안전한 범위
		result.AllocsPerOp = int64(mallocsDelta) / int64(result.Operations)   //nolint:gosec // G115: 벤치마크 통계는 안전한 범위
	}

	result.Timestamp = time.Now()

	// Store result
	bs.mu.Lock()
	bs.results = append(bs.results, *result)
	bs.mu.Unlock()

	// Log benchmark completion
	bs.logger.LogPerformance(name+"_benchmark", result.Duration, map[string]interface{}{
		"operations":         result.Operations,
		"ns_per_op":          result.NsPerOp,
		"ops_per_sec":        result.OpsPerSec,
		"alloc_bytes_per_op": result.AllocBytesPerOp,
		"allocs_per_op":      result.AllocsPerOp,
		"memory_growth":      result.MemoryAfter - result.MemoryBefore,
		"goroutine_delta":    result.GoroutinesAfter - result.GoroutinesBefore,
	})

	return result, nil
}

// executeBenchmark performs the actual benchmark execution.
func (bs *BenchmarkSuite) executeBenchmark(ctx context.Context, name string, fn BenchmarkFunc, opts *BenchmarkOptions) (*BenchmarkResult, error) {
	result := &BenchmarkResult{
		Name:        name,
		Percentiles: make(map[string]time.Duration),
	}

	if opts.Concurrency <= 1 {
		return bs.runSequentialBenchmark(ctx, result, fn, opts)
	}

	return bs.runConcurrentBenchmark(ctx, result, fn, opts)
}

// runSequentialBenchmark runs benchmark operations sequentially.
func (bs *BenchmarkSuite) runSequentialBenchmark(ctx context.Context, result *BenchmarkResult, fn BenchmarkFunc, opts *BenchmarkOptions) (*BenchmarkResult, error) {
	durations := make([]time.Duration, 0, opts.Iterations)
	startTime := time.Now()

	operations := 0
	for i := 0; i < opts.Iterations || (opts.Duration > 0 && time.Since(startTime) < opts.Duration); i++ {
		if ctx.Err() != nil {
			break
		}

		opStart := time.Now()
		if err := fn(ctx); err != nil {
			bs.logger.Warn("Benchmark operation failed", "name", result.Name, "iteration", i, "error", err)
			continue
		}
		opDuration := time.Since(opStart)

		durations = append(durations, opDuration)
		operations++

		// Break if we've reached the duration limit
		if opts.Duration > 0 && time.Since(startTime) >= opts.Duration {
			break
		}
	}

	totalDuration := time.Since(startTime)

	result.Operations = operations
	result.Duration = totalDuration

	if operations > 0 {
		result.NsPerOp = totalDuration.Nanoseconds() / int64(operations)
		result.OpsPerSec = float64(operations) / totalDuration.Seconds()
		result.Percentiles = calculatePercentiles(durations)
	}

	return result, nil
}

// runConcurrentBenchmark runs benchmark operations concurrently.
func (bs *BenchmarkSuite) runConcurrentBenchmark(ctx context.Context, result *BenchmarkResult, fn BenchmarkFunc, opts *BenchmarkOptions) (*BenchmarkResult, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	durations := make([]time.Duration, 0, opts.Iterations)

	iterationsPerWorker := opts.Iterations / opts.Concurrency
	remainingIterations := opts.Iterations % opts.Concurrency

	startTime := time.Now()
	operations := int64(0)

	for worker := 0; worker < opts.Concurrency; worker++ {
		wg.Add(1)
		workerIterations := iterationsPerWorker
		if worker < remainingIterations {
			workerIterations++
		}

		go func(iterations int) {
			defer wg.Done()
			workerDurations := make([]time.Duration, 0, iterations)

			for i := 0; i < iterations; i++ {
				if ctx.Err() != nil {
					break
				}

				opStart := time.Now()
				if err := fn(ctx); err != nil {
					bs.logger.Debug("Concurrent benchmark operation failed", "error", err)
					continue
				}
				opDuration := time.Since(opStart)

				workerDurations = append(workerDurations, opDuration)
			}

			mu.Lock()
			durations = append(durations, workerDurations...)
			operations += int64(len(workerDurations))
			mu.Unlock()
		}(workerIterations)
	}

	wg.Wait()
	totalDuration := time.Since(startTime)

	result.Operations = int(operations)
	result.Duration = totalDuration

	if operations > 0 {
		result.NsPerOp = totalDuration.Nanoseconds() / operations
		result.OpsPerSec = float64(operations) / totalDuration.Seconds()
		result.Percentiles = calculatePercentiles(durations)
	}

	return result, nil
}

// calculatePercentiles calculates percentile durations from a slice of durations.
func calculatePercentiles(durations []time.Duration) map[string]time.Duration {
	if len(durations) == 0 {
		return map[string]time.Duration{}
	}

	sort.Slice(durations, func(i, j int) bool {
		return durations[i] < durations[j]
	})

	percentiles := map[string]time.Duration{
		"p50": durations[int(float64(len(durations))*0.50)],
		"p90": durations[int(float64(len(durations))*0.90)],
		"p95": durations[int(float64(len(durations))*0.95)],
		"p99": durations[int(float64(len(durations))*0.99)],
		"min": durations[0],
		"max": durations[len(durations)-1],
	}

	return percentiles
}

// GetResults returns all benchmark results.
func (bs *BenchmarkSuite) GetResults() []BenchmarkResult {
	bs.mu.RLock()
	defer bs.mu.RUnlock()

	results := make([]BenchmarkResult, len(bs.results))
	copy(results, bs.results)
	return results
}

// CompareBenchmarks compares two benchmark results.
func (bs *BenchmarkSuite) CompareBenchmarks(baseline, current *BenchmarkResult) map[string]interface{} {
	if baseline == nil || current == nil {
		return nil
	}

	comparison := map[string]interface{}{
		"baseline_name": baseline.Name,
		"current_name":  current.Name,
		"timestamp":     time.Now(),
	}

	// Performance comparison
	if baseline.NsPerOp > 0 {
		speedupRatio := float64(baseline.NsPerOp) / float64(current.NsPerOp)
		comparison["speedup_ratio"] = speedupRatio
		comparison["performance_change_percent"] = (speedupRatio - 1.0) * 100
	}

	// Memory comparison
	if baseline.AllocBytesPerOp > 0 {
		memoryRatio := float64(current.AllocBytesPerOp) / float64(baseline.AllocBytesPerOp)
		comparison["memory_ratio"] = memoryRatio
		comparison["memory_change_percent"] = (memoryRatio - 1.0) * 100
	}

	// Allocation comparison
	if baseline.AllocsPerOp > 0 {
		allocRatio := float64(current.AllocsPerOp) / float64(baseline.AllocsPerOp)
		comparison["alloc_ratio"] = allocRatio
		comparison["alloc_change_percent"] = (allocRatio - 1.0) * 100
	}

	return comparison
}

// PrintResults prints benchmark results in a formatted way.
func (bs *BenchmarkSuite) PrintResults() {
	results := bs.GetResults()
	if len(results) == 0 {
		bs.logger.Info("No benchmark results available")
		return
	}

	bs.logger.Info("Benchmark Results Summary", "total_benchmarks", len(results))

	for _, result := range results {
		bs.logger.Info("Benchmark Result",
			"name", result.Name,
			"operations", result.Operations,
			"duration", result.Duration,
			"ns_per_op", result.NsPerOp,
			"ops_per_sec", fmt.Sprintf("%.2f", result.OpsPerSec),
			"alloc_bytes_per_op", result.AllocBytesPerOp,
			"allocs_per_op", result.AllocsPerOp,
		)

		if len(result.Percentiles) > 0 {
			bs.logger.Debug("Benchmark Percentiles",
				"name", result.Name,
				"p50", result.Percentiles["p50"],
				"p90", result.Percentiles["p90"],
				"p95", result.Percentiles["p95"],
				"p99", result.Percentiles["p99"],
				"min", result.Percentiles["min"],
				"max", result.Percentiles["max"],
			)
		}
	}
}

// RunSimpleBenchmark executes a simple benchmark with specified parameters.
func (bs *BenchmarkSuite) RunSimpleBenchmark(ctx context.Context, name string, fn SimpleBenchmarkFunc, iterations int, duration time.Duration) (*BenchmarkResult, error) {
	// Convert simple function to BenchmarkFunc
	benchmarkFn := func(ctx context.Context) error {
		fn(ctx)
		return nil
	}

	// Create options from parameters
	opts := &BenchmarkOptions{
		Iterations:      iterations,
		Duration:        duration,
		WarmupRuns:      10, // Small warmup for simple benchmarks
		Concurrency:     1,
		MemoryProfiling: true,
		CPUProfiling:    false,
	}

	return bs.RunBenchmark(ctx, name, benchmarkFn, opts)
}
