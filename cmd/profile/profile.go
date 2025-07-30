// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package profile

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/gizzahub/gzh-manager-go/internal/cli"
	"github.com/gizzahub/gzh-manager-go/internal/logger"
	"github.com/gizzahub/gzh-manager-go/internal/profiling"
)

// NewProfileCmd creates the profile command
func NewProfileCmd() *cobra.Command {
	ctx := context.Background()

	cmd := cli.NewCommandBuilder(ctx, "profile", "Performance profiling and benchmarking tools").
		WithLongDescription(`The profile command provides comprehensive performance profiling and benchmarking capabilities.

It includes support for:
- CPU profiling
- Memory profiling  
- Goroutine profiling
- Block profiling
- Mutex profiling
- HTTP pprof server
- Benchmark execution
- Performance comparison

Examples:
  gz profile start --type cpu --duration 30s
  gz profile stop --session session_id
  gz profile server --port 6060
  gz profile benchmark --name "git-clone" --iterations 100
  gz profile stats
`).
		WithExample("gz profile start --type cpu --duration 30s").
		Build()

	// Add subcommands
	cmd.AddCommand(newStartCmd())
	cmd.AddCommand(newStopCmd())
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newServerCmd())
	cmd.AddCommand(newBenchmarkCmd())
	cmd.AddCommand(newStatsCmd())
	cmd.AddCommand(newCompareCmd())

	return cmd
}

// Global profiler instance
var globalProfiler *profiling.Profiler

// initProfiler initializes the global profiler if not already done
func initProfiler() *profiling.Profiler {
	if globalProfiler == nil {
		config := &profiling.ProfileConfig{
			Enabled:     true,
			HTTPPort:    6060,
			OutputDir:   "tmp/profiles",
			AutoProfile: false,
			CPUDuration: 30 * time.Second,
		}
		globalProfiler = profiling.NewProfiler(config)
	}
	return globalProfiler
}

// newStartCmd creates the start subcommand
func newStartCmd() *cobra.Command {
	var profileType string
	var duration time.Duration
	var outputDir string

	ctx := context.Background()

	cmd := cli.NewCommandBuilder(ctx, "start", "Start a profiling session").
		WithLongDescription(`Start a new profiling session of the specified type.

Supported profile types:
- cpu: CPU profiling (requires duration)
- memory: Memory heap profiling
- goroutine: Goroutine profiling
- block: Block profiling
- mutex: Mutex profiling
- threadcreate: Thread creation profiling

Examples:
  gz profile start --type cpu --duration 30s
  gz profile start --type memory
  gz profile start --type goroutine --output-dir ./profiles
`).
		WithRunFuncE(func(ctx context.Context, flags *cli.CommonFlags, args []string) error {
			profiler := initProfiler()

			if outputDir != "" {
				// Update profiler config
				profiler = profiling.NewProfiler(&profiling.ProfileConfig{
					Enabled:   true,
					OutputDir: outputDir,
				})
			}

			sessionID, err := profiler.StartProfile(profiling.ProfileType(profileType))
			if err != nil {
				return fmt.Errorf("failed to start profile: %w", err)
			}

			logger.SimpleInfo("Started profiling session", "type", profileType, "session_id", sessionID)

			// For CPU profiling, automatically stop after duration
			if profileType == "cpu" && duration > 0 {
				go func() {
					time.Sleep(duration)
					if err := profiler.StopProfile(sessionID); err != nil {
						logger.SimpleError("Failed to auto-stop CPU profile", "error", err)
					} else {
						logger.SimpleInfo("Auto-stopped CPU profile", "session_id", sessionID, "duration", duration)
					}
				}()
			}

			return nil
		}).
		Build()

	cmd.Flags().StringVar(&profileType, "type", "cpu", "Profile type (cpu, memory, goroutine, block, mutex, threadcreate)")
	cmd.Flags().DurationVar(&duration, "duration", 30*time.Second, "Duration for CPU profiling (ignored for other types)")
	cmd.Flags().StringVar(&outputDir, "output-dir", "", "Custom output directory for profile files")

	cmd.MarkFlagRequired("type")

	return cmd
}

// newStopCmd creates the stop subcommand
func newStopCmd() *cobra.Command {
	var sessionID string

	ctx := context.Background()

	cmd := cli.NewCommandBuilder(ctx, "stop", "Stop a profiling session").
		WithLongDescription(`Stop an active profiling session and save the results.

Examples:
  gz profile stop --session cpu_1640995200
  gz profile stop --session memory_1640995300
`).
		WithRunFuncE(func(ctx context.Context, flags *cli.CommonFlags, args []string) error {
			profiler := initProfiler()

			if err := profiler.StopProfile(sessionID); err != nil {
				return fmt.Errorf("failed to stop profile: %w", err)
			}

			logger.SimpleInfo("Stopped profiling session", "session_id", sessionID)
			return nil
		}).
		Build()

	cmd.Flags().StringVar(&sessionID, "session", "", "Session ID to stop")
	cmd.MarkFlagRequired("session")

	return cmd
}

// newListCmd creates the list subcommand
func newListCmd() *cobra.Command {
	ctx := context.Background()

	cmd := cli.NewCommandBuilder(ctx, "list", "List active profiling sessions").
		WithLongDescription(`List all currently active profiling sessions.

Shows session ID, type, start time, and duration for each active session.

Examples:
  gz profile list
`).
		WithRunFuncE(func(ctx context.Context, flags *cli.CommonFlags, args []string) error {
			profiler := initProfiler()

			sessions := profiler.ListActiveSessions()
			if len(sessions) == 0 {
				logger.SimpleInfo("No active profiling sessions")
				return nil
			}

			logger.SimpleInfo("Active profiling sessions", "count", len(sessions))
			for sessionID, session := range sessions {
				duration := time.Since(session.StartTime)
				logger.SimpleInfo("Session",
					"id", sessionID,
					"type", session.Type,
					"duration", duration.Round(time.Second),
					"start_time", session.StartTime.Format("15:04:05"),
				)
			}

			return nil
		}).
		Build()

	return cmd
}

// newServerCmd creates the server subcommand
func newServerCmd() *cobra.Command {
	var port int
	var autoProfile bool

	ctx := context.Background()

	cmd := cli.NewCommandBuilder(ctx, "server", "Start HTTP profiling server").
		WithLongDescription(`Start an HTTP server that exposes pprof endpoints for profiling.

The server provides the following endpoints:
- /debug/pprof/ - Profile index
- /debug/pprof/profile - CPU profile
- /debug/pprof/heap - Heap profile
- /debug/pprof/goroutine - Goroutine profile
- /debug/pprof/block - Block profile
- /debug/pprof/mutex - Mutex profile
- /debug/stats - Runtime statistics

Examples:
  gz profile server --port 6060
  gz profile server --port 8080 --auto-profile
`).
		WithRunFuncE(func(ctx context.Context, flags *cli.CommonFlags, args []string) error {
			config := &profiling.ProfileConfig{
				Enabled:     true,
				HTTPPort:    port,
				OutputDir:   "tmp/profiles",
				AutoProfile: autoProfile,
			}

			profiler := profiling.NewProfiler(config)

			logger.SimpleInfo("Starting profiling server", "port", port, "auto_profile", autoProfile)

			if err := profiler.Start(ctx); err != nil {
				return fmt.Errorf("failed to start profiling server: %w", err)
			}

			// Keep server running until context is cancelled
			<-ctx.Done()

			return profiler.Stop()
		}).
		Build()

	cmd.Flags().IntVar(&port, "port", 6060, "HTTP server port")
	cmd.Flags().BoolVar(&autoProfile, "auto-profile", false, "Enable automatic periodic profiling")

	return cmd
}

// newBenchmarkCmd creates the benchmark subcommand
func newBenchmarkCmd() *cobra.Command {
	var benchmarkName string
	var iterations int
	var concurrency int
	var duration time.Duration
	var warmupRuns int
	var memoryProfiling bool
	var cpuProfiling bool

	ctx := context.Background()

	cmd := cli.NewCommandBuilder(ctx, "benchmark", "Run performance benchmarks").
		WithLongDescription(`Run performance benchmarks for various operations.

This command can benchmark built-in operations or be extended to benchmark
custom operations in your codebase.

Built-in benchmarks:
- memory-allocation: Memory allocation performance
- goroutine-creation: Goroutine creation performance
- channel-operations: Channel send/receive performance
- json-marshal: JSON marshaling performance
- json-unmarshal: JSON unmarshaling performance

Examples:
  gz profile benchmark --name memory-allocation --iterations 1000
  gz profile benchmark --name goroutine-creation --concurrency 4 --iterations 500
  gz profile benchmark --name json-marshal --duration 10s --memory-profiling
`).
		WithRunFuncE(func(ctx context.Context, flags *cli.CommonFlags, args []string) error {
			profiler := initProfiler()
			suite := profiling.NewBenchmarkSuite(profiler)

			opts := &profiling.BenchmarkOptions{
				Iterations:      iterations,
				Duration:        duration,
				WarmupRuns:      warmupRuns,
				Concurrency:     concurrency,
				MemoryProfiling: memoryProfiling,
				CPUProfiling:    cpuProfiling,
			}

			benchmarkFunc, err := getBenchmarkFunction(benchmarkName)
			if err != nil {
				return err
			}

			logger.SimpleInfo("Starting benchmark",
				"name", benchmarkName,
				"iterations", iterations,
				"concurrency", concurrency,
			)

			result, err := suite.RunBenchmark(ctx, benchmarkName, benchmarkFunc, opts)
			if err != nil {
				return fmt.Errorf("benchmark failed: %w", err)
			}

			// Print results
			logger.SimpleInfo("Benchmark completed",
				"name", result.Name,
				"operations", result.Operations,
				"duration", result.Duration,
				"ns_per_op", result.NsPerOp,
				"ops_per_sec", fmt.Sprintf("%.2f", result.OpsPerSec),
				"alloc_bytes_per_op", result.AllocBytesPerOp,
				"allocs_per_op", result.AllocsPerOp,
			)

			if len(result.Percentiles) > 0 {
				logger.SimpleInfo("Benchmark percentiles",
					"p50", result.Percentiles["p50"],
					"p90", result.Percentiles["p90"],
					"p95", result.Percentiles["p95"],
					"p99", result.Percentiles["p99"],
				)
			}

			return nil
		}).
		Build()

	cmd.Flags().StringVar(&benchmarkName, "name", "memory-allocation", "Benchmark name")
	cmd.Flags().IntVar(&iterations, "iterations", 1000, "Number of benchmark iterations")
	cmd.Flags().IntVar(&concurrency, "concurrency", 1, "Number of concurrent goroutines")
	cmd.Flags().DurationVar(&duration, "duration", 0, "Benchmark duration (overrides iterations)")
	cmd.Flags().IntVar(&warmupRuns, "warmup", 100, "Number of warmup runs")
	cmd.Flags().BoolVar(&memoryProfiling, "memory-profiling", true, "Enable memory profiling")
	cmd.Flags().BoolVar(&cpuProfiling, "cpu-profiling", false, "Enable CPU profiling")

	return cmd
}

// newStatsCmd creates the stats subcommand
func newStatsCmd() *cobra.Command {
	ctx := context.Background()

	cmd := cli.NewCommandBuilder(ctx, "stats", "Show runtime performance statistics").
		WithLongDescription(`Display current runtime performance statistics including:
- Number of goroutines
- Memory usage and allocation statistics  
- Garbage collection statistics
- CGO call count

Examples:
  gz profile stats
  gz profile stats --format json
  gz profile stats --format yaml
`).
		WithFormatFlag("table", []string{"table", "json", "yaml"}).
		WithRunFuncE(func(ctx context.Context, flags *cli.CommonFlags, args []string) error {
			profiler := initProfiler()
			stats := profiler.GetRuntimeStats()

			formatter := cli.NewOutputFormatter(flags.Format)
			return formatter.FormatOutput(stats)
		}).
		Build()

	return cmd
}

// newCompareCmd creates the compare subcommand
func newCompareCmd() *cobra.Command {
	var baselineName string
	var currentName string

	ctx := context.Background()

	cmd := cli.NewCommandBuilder(ctx, "compare", "Compare benchmark results").
		WithLongDescription(`Compare two benchmark results to analyze performance differences.

This command helps identify performance regressions or improvements between
different versions or configurations.

Examples:
  gz profile compare --baseline "v1.0" --current "v1.1"
`).
		WithRunFuncE(func(ctx context.Context, flags *cli.CommonFlags, args []string) error {
			// This is a simplified implementation
			// In a real scenario, you'd load benchmark results from storage
			logger.SimpleInfo("Benchmark comparison feature",
				"baseline", baselineName,
				"current", currentName,
				"status", "not_implemented",
			)

			logger.SimpleWarn("Benchmark comparison requires stored benchmark results")
			logger.SimpleInfo("Run benchmarks first, then implement result storage for comparison")

			return nil
		}).
		Build()

	cmd.Flags().StringVar(&baselineName, "baseline", "", "Baseline benchmark name")
	cmd.Flags().StringVar(&currentName, "current", "", "Current benchmark name")
	cmd.MarkFlagRequired("baseline")
	cmd.MarkFlagRequired("current")

	return cmd
}

// getBenchmarkFunction returns a benchmark function for the given name
func getBenchmarkFunction(name string) (profiling.BenchmarkFunc, error) {
	switch strings.ToLower(name) {
	case "memory-allocation":
		return func(ctx context.Context) error {
			data := make([]byte, 1024)
			for i := range data {
				data[i] = byte(i % 256)
			}
			return nil
		}, nil

	case "goroutine-creation":
		return func(ctx context.Context) error {
			done := make(chan struct{})
			go func() {
				time.Sleep(time.Microsecond)
				close(done)
			}()
			<-done
			return nil
		}, nil

	case "channel-operations":
		return func(ctx context.Context) error {
			ch := make(chan int, 1)
			ch <- 42
			<-ch
			return nil
		}, nil

	case "json-marshal":
		return func(ctx context.Context) error {
			data := map[string]interface{}{
				"name":    "test",
				"version": "1.0.0",
				"items":   []int{1, 2, 3, 4, 5},
			}
			_ = fmt.Sprintf("%+v", data) // Simulate marshaling work
			return nil
		}, nil

	case "string-operations":
		return func(ctx context.Context) error {
			var result strings.Builder
			for i := 0; i < 100; i++ {
				result.WriteString("test" + strconv.Itoa(i))
			}
			_ = result.String()
			return nil
		}, nil

	default:
		return nil, fmt.Errorf("unknown benchmark: %s", name)
	}
}
