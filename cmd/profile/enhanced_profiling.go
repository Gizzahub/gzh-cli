// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package profile

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/gizzahub/gzh-cli/internal/simpleprof"
)

// ProfileAnalyzer collects profiles on a schedule; analysis is left to go tool pprof.
type ProfileAnalyzer struct {
	profiler  *simpleprof.SimpleProfiler
	outputDir string
}

// NewProfileAnalyzer creates a new profile analyzer.
func NewProfileAnalyzer(outputDir string) *ProfileAnalyzer {
	return &ProfileAnalyzer{
		profiler:  simpleprof.NewSimpleProfiler(outputDir),
		outputDir: outputDir,
	}
}

// newContinuousCmd creates a command for continuous profiling.
func newContinuousCmd() *cobra.Command {
	var interval time.Duration
	var duration time.Duration
	var profileType string

	cmd := &cobra.Command{
		Use:   "continuous",
		Short: "Run continuous profiling over time",
		Long: `Start continuous profiling that collects profiles at regular intervals.

This is useful for monitoring performance over time and detecting gradual performance
degradation or improvements. Profiles are saved with timestamps; inspect them with
"go tool pprof -top <file>" and compare two with
"go tool pprof -diff_base=<base> <current>".

Examples:
  gz profile continuous --interval 5m --duration 1h
  gz profile continuous --type cpu --interval 1m --duration 30m
  gz profile continuous --type memory --interval 10m --duration 2h`,
		RunE: func(cmd *cobra.Command, args []string) error {
			analyzer := NewProfileAnalyzer("tmp/profiles")

			// cmd.Context()여야 한다. 예전에는 context.Background()를
			// 줘서 RunContinuousProfiling 안의 ctx.Done()이 죽어 있었다.
			// --duration 기본값이 1시간인데 Ctrl+C로 멈출 수가 없었다 --
			// apprunner가 signal.Notify로 기본 동작(신호 받으면 죽기)을
			// 가로챈 뒤 자기 맥락만 취소하기 때문이다.
			return analyzer.RunContinuousProfiling(cmd.Context(), profileType, interval, duration)
		},
	}

	cmd.Flags().DurationVar(&interval, "interval", 5*time.Minute, "Interval between profile collections")
	cmd.Flags().DurationVar(&duration, "duration", 1*time.Hour, "Total duration to run continuous profiling")
	cmd.Flags().StringVar(&profileType, "type", "cpu", "Profile type: cpu, memory, goroutine")

	return cmd
}

// RunContinuousProfiling runs continuous profiling for the specified duration.
func (pa *ProfileAnalyzer) RunContinuousProfiling(ctx context.Context, profileType string, interval, duration time.Duration) error {
	fmt.Printf("🔄 Starting continuous %s profiling...\n", profileType)
	fmt.Printf("📊 Interval: %v, Duration: %v\n", interval, duration)

	startTime := time.Now()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	profileCount := 0

	for {
		select {
		case <-ctx.Done():
			// Return ctx.Err() so apprunner maps context.Canceled → exit 130 (POSIX SIGINT).
			fmt.Println("🛑 Continuous profiling stopped by context")
			return ctx.Err()

		case <-ticker.C:
			if time.Since(startTime) > duration {
				fmt.Printf("✅ Continuous profiling completed after %v\n", time.Since(startTime).Round(time.Second))
				fmt.Printf("📈 Collected %d profiles\n", profileCount)
				return nil
			}

			profileCount++
			timestamp := time.Now().Format("20060102_150405")
			fmt.Printf("📸 [%d] Collecting %s profile at %s...\n", profileCount, profileType, timestamp)

			var profileTypeEnum simpleprof.ProfileType
			switch profileType {
			case "cpu":
				profileTypeEnum = simpleprof.ProfileTypeCPU
			case "memory":
				profileTypeEnum = simpleprof.ProfileTypeMemory
			case "goroutine":
				profileTypeEnum = simpleprof.ProfileTypeGoroutine
			default:
				profileTypeEnum = simpleprof.ProfileTypeCPU
			}

			filename, err := pa.profiler.StartProfile(profileTypeEnum, 10*time.Second)
			if err != nil {
				fmt.Printf("⚠️  Failed to collect profile: %v\n", err)
				continue
			}

			fmt.Printf("💾 Profile saved: %s\n", filepath.Base(filename))
		}
	}
}

// Enhanced stats command with more detailed information.
//
//nolint:unused // Reserved for future enhancement
func newEnhancedStatsCmd() *cobra.Command {
	var interval time.Duration
	var count int
	var format string

	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show enhanced runtime statistics",
		Long: `Display detailed runtime statistics with optional continuous monitoring.

This enhanced version provides more detailed metrics and can continuously monitor
runtime statistics over time.

Examples:
  gz profile stats
  gz profile stats --interval 5s --count 10
  gz profile stats --format json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			profiler := simpleprof.NewSimpleProfiler("tmp/profiles")

			if interval > 0 {
				return printContinuousStats(profiler, interval, count, format)
			}

			return printEnhancedStats(profiler, format)
		},
	}

	cmd.Flags().DurationVar(&interval, "interval", 0, "Interval for continuous monitoring (e.g., 5s)")
	cmd.Flags().IntVar(&count, "count", 0, "Number of samples for continuous monitoring (0 = unlimited)")
	cmd.Flags().StringVar(&format, "format", "text", "Output format: text, json, csv")

	return cmd
}

// GetEnhancedStats returns detailed runtime statistics.
func GetEnhancedStats() map[string]any {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	stats := map[string]any{
		"timestamp": time.Now(),
		"memory": map[string]any{
			"heap_alloc":    m.HeapAlloc,
			"heap_sys":      m.HeapSys,
			"heap_inuse":    m.HeapInuse,
			"heap_released": m.HeapReleased,
			"heap_idle":     m.HeapIdle,
			"stack_inuse":   m.StackInuse,
			"stack_sys":     m.StackSys,
			"total_alloc":   m.TotalAlloc,
		},
		"gc": map[string]any{
			"num_gc":          m.NumGC,
			"last_gc":         simpleprof.LastGCTime(m.LastGC),
			"pause_total_ns":  m.PauseTotalNs,
			"gc_cpu_fraction": m.GCCPUFraction,
		},
		"runtime": map[string]any{
			"goroutines": runtime.NumGoroutine(),
			"num_cpu":    runtime.NumCPU(),
			"gomaxprocs": runtime.GOMAXPROCS(0),
			"version":    runtime.Version(),
		},
	}

	return stats
}

//nolint:unused // Reserved for future enhancement
func printEnhancedStats(profiler *simpleprof.SimpleProfiler, format string) error {
	stats := GetEnhancedStats()

	if format == "json" {
		// In real implementation, use json.Marshal
		memory, ok := stats["memory"].(map[string]any)
		if !ok {
			return fmt.Errorf("invalid memory stats format")
		}
		heapAlloc, ok := memory["heap_alloc"].(uint64)
		if !ok {
			return fmt.Errorf("invalid heap_alloc format")
		}
		heapSys, ok := memory["heap_sys"].(uint64)
		if !ok {
			return fmt.Errorf("invalid heap_sys format")
		}
		fmt.Printf(`{
  "timestamp": "%s",
  "goroutines": %d,
  "memory": {
    "heapAlloc": "%s",
    "heapSys": "%s"
  }
}`, stats["timestamp"],
			runtime.NumGoroutine(),
			formatBytes(heapAlloc),
			formatBytes(heapSys))
		return nil
	}

	// Enhanced text format
	fmt.Printf("📊 Enhanced Runtime Statistics\n")
	fmt.Printf("==============================\n")
	fmt.Printf("Timestamp: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))

	// Memory section
	fmt.Printf("💾 Memory:\n")
	memory, ok := stats["memory"].(map[string]any)
	if !ok {
		return fmt.Errorf("invalid memory stats format")
	}
	heapAlloc, ok := memory["heap_alloc"].(uint64)
	if !ok {
		return fmt.Errorf("invalid heap_alloc format")
	}
	heapSys, ok := memory["heap_sys"].(uint64)
	if !ok {
		return fmt.Errorf("invalid heap_sys format")
	}
	fmt.Printf("  Heap Allocated:   %s\n", formatBytes(heapAlloc))
	fmt.Printf("  Heap System:      %s\n", formatBytes(heapSys))

	heapInuse, ok := memory["heap_inuse"].(uint64)
	if !ok {
		return fmt.Errorf("invalid heap_inuse format")
	}
	totalAlloc, ok := memory["total_alloc"].(uint64)
	if !ok {
		return fmt.Errorf("invalid total_alloc format")
	}
	stackInuse, ok := memory["stack_inuse"].(uint64)
	if !ok {
		return fmt.Errorf("invalid stack_inuse format")
	}

	fmt.Printf("  Heap In Use:      %s\n", formatBytes(heapInuse))
	fmt.Printf("  Total Allocated:  %s\n", formatBytes(totalAlloc))
	fmt.Printf("  Stack In Use:     %s\n\n", formatBytes(stackInuse))

	// GC section
	fmt.Printf("🗑️  Garbage Collection:\n")
	gc, ok := stats["gc"].(map[string]any)
	if !ok {
		return fmt.Errorf("invalid gc stats format")
	}
	numGC, ok := gc["num_gc"].(uint32)
	if !ok {
		return fmt.Errorf("invalid num_gc format")
	}
	fmt.Printf("  GC Runs:          %d\n", numGC)

	lastGC, ok := gc["last_gc"].(time.Time)
	if !ok {
		return fmt.Errorf("invalid last_gc format")
	}
	gcCPUFraction, ok := gc["gc_cpu_fraction"].(float64)
	if !ok {
		return fmt.Errorf("invalid gc_cpu_fraction format")
	}
	fmt.Printf("  Last GC:          %s\n", lastGC.Format("15:04:05"))
	fmt.Printf("  GC CPU Fraction:  %.3f\n\n", gcCPUFraction)

	// Runtime section
	fmt.Printf("⚙️  Runtime:\n")
	runtimeStats, ok := stats["runtime"].(map[string]any)
	if !ok {
		return fmt.Errorf("invalid runtime stats format")
	}
	goroutines, ok := runtimeStats["goroutines"].(int)
	if !ok {
		return fmt.Errorf("invalid goroutines format")
	}
	numCPU, ok := runtimeStats["num_cpu"].(int)
	if !ok {
		return fmt.Errorf("invalid num_cpu format")
	}
	gomaxprocs, ok := runtimeStats["gomaxprocs"].(int)
	if !ok {
		return fmt.Errorf("invalid gomaxprocs format")
	}
	version, ok := runtimeStats["version"].(string)
	if !ok {
		return fmt.Errorf("invalid version format")
	}
	fmt.Printf("  Goroutines:       %d\n", goroutines)
	fmt.Printf("  CPU Cores:        %d\n", numCPU)
	fmt.Printf("  GOMAXPROCS:       %d\n", gomaxprocs)
	fmt.Printf("  Go Version:       %s\n", version)

	return nil
}

//nolint:unused // Reserved for future enhancement
func printContinuousStats(profiler *simpleprof.SimpleProfiler, interval time.Duration, count int, format string) error {
	fmt.Printf("📊 Continuous Statistics Monitoring\n")
	fmt.Printf("Interval: %v, Count: %s\n", interval, func() string {
		if count == 0 {
			return "unlimited"
		}
		return strconv.Itoa(count)
	}())
	fmt.Println("Press Ctrl+C to stop")

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	sampleCount := 0

	for range ticker.C {
		sampleCount++

		if format == "csv" && sampleCount == 1 {
			fmt.Println("timestamp,goroutines,heap_alloc_mb,heap_sys_mb,gc_runs")
		}

		stats := GetEnhancedStats()
		memory, ok := stats["memory"].(map[string]any)
		if !ok {
			fmt.Fprintf(os.Stderr, "Error: invalid memory stats format\n")
			break
		}

		timestamp, ok := stats["timestamp"].(time.Time)
		if !ok {
			fmt.Fprintf(os.Stderr, "Error: invalid timestamp format\n")
			break
		}

		heapAlloc, ok := memory["heap_alloc"].(uint64)
		if !ok {
			fmt.Fprintf(os.Stderr, "Error: invalid heap_alloc format\n")
			break
		}

		heapSys, ok := memory["heap_sys"].(uint64)
		if !ok {
			fmt.Fprintf(os.Stderr, "Error: invalid heap_sys format\n")
			break
		}

		gc, ok := stats["gc"].(map[string]any)
		if !ok {
			fmt.Fprintf(os.Stderr, "Error: invalid gc stats format\n")
			break
		}

		numGC, ok := gc["num_gc"].(uint32)
		if !ok {
			fmt.Fprintf(os.Stderr, "Error: invalid num_gc format\n")
			break
		}

		if format == "csv" {
			fmt.Printf("%s,%d,%.1f,%.1f,%d\n",
				timestamp.Format("15:04:05"),
				runtime.NumGoroutine(),
				float64(heapAlloc)/1024/1024,
				float64(heapSys)/1024/1024,
				numGC)
		} else {
			fmt.Printf("[%s] Goroutines: %3d | Heap: %8s | GC: %3d runs\n",
				timestamp.Format("15:04:05"),
				runtime.NumGoroutine(),
				formatBytes(heapAlloc),
				numGC)
		}

		if count > 0 && sampleCount >= count {
			break
		}
	}

	return nil
}
