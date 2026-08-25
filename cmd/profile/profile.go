// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package profile

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/gizzahub/gzh-cli/internal/app"
	"github.com/gizzahub/gzh-cli/internal/simpleprof"
)

// NewProfileCmd creates a simplified profile command using standard Go pprof.
func NewProfileCmd(appCtx *app.AppContext) *cobra.Command {
	_ = appCtx
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Performance profiling using standard Go pprof",
		Long: `Comprehensive performance profiling using standard Go pprof tools with enhanced analysis capabilities.

Available commands:
  server      Start pprof HTTP server
  cpu         Collect CPU profile
  memory      Collect memory profile
  stats       Show runtime statistics

Enhanced commands:
  compare     Compare two profiles for performance differences
  continuous  Run continuous profiling over time
  analyze     Analyze profile for performance issues

Examples:
  gz profile server --port 6060
  gz profile cpu --duration 30s
  gz profile memory
  gz profile stats
  gz profile compare baseline.prof current.prof
  gz profile continuous --interval 5m --duration 1h
  gz profile analyze cpu.prof`,
	}

	// Add basic subcommands
	cmd.AddCommand(newSimpleServerCmd())
	cmd.AddCommand(newSimpleCPUCmd())
	cmd.AddCommand(newSimpleMemoryCmd())
	cmd.AddCommand(newSimpleStatsCmd())

	// Add enhanced subcommands
	cmd.AddCommand(newCompareCmd())
	cmd.AddCommand(newContinuousCmd())
	cmd.AddCommand(newAnalyzeCmd())

	return cmd
}

func newSimpleServerCmd() *cobra.Command {
	var port int

	cmd := &cobra.Command{
		Use:   "server",
		Short: "Start pprof HTTP server",
		Long: `Start HTTP server with pprof endpoints.

The server provides the following endpoints:
  /debug/pprof/           - Index page
  /debug/pprof/profile    - CPU profile
  /debug/pprof/heap       - Memory profile
  /debug/pprof/goroutine  - Goroutine profile
  /debug/pprof/block      - Block profile
  /debug/pprof/mutex      - Mutex profile`,
		RunE: func(cmd *cobra.Command, args []string) error {
			profiler := simpleprof.NewSimpleProfiler("tmp/profiles")

			if err := profiler.StartHTTPServer(port); err != nil {
				return fmt.Errorf("failed to start pprof server: %w", err)
			}

			fmt.Printf("✅ Pprof server started on http://127.0.0.1:%d/debug/pprof/\n", port)
			fmt.Println("Press Ctrl+C to stop the server")

			// 멈추라는 신호를 기다린다.
			//
			// 예전에는 여기서 새 맥락을 만들고 cancel을 defer한 뒤 그
			// 맥락의 Done을 기다렸다. 그 cancel은 이 함수가 끝나야 불리는데
			// 함수는 Done을 기다리느라 끝나지 않는다. 무조건 걸리는
			// 자리였다. Ctrl+C도 소용없었다 -- apprunner가 signal.Notify로
			// 기본 동작(신호 받으면 죽기)을 가로챈 뒤 자기 맥락만 취소하기
			// 때문이다. 신호 처리기를 달아 둔 것이 오히려 못 죽게 만들었다.
			//
			// cmd.Context()가 root의 ExecuteContext를 거쳐 온 그 맥락이다.
			<-cmd.Context().Done()

			// 멈추는 데 쓰는 맥락은 따로 둔다. 방금 취소된 것을 그대로
			// 넘기면 정리할 틈도 없이 끊긴다.
			return profiler.StopHTTPServer(context.Background())
		},
	}

	cmd.Flags().IntVar(&port, "port", 6060, "Port for pprof HTTP server")
	return cmd
}

func newSimpleCPUCmd() *cobra.Command {
	var duration time.Duration

	cmd := &cobra.Command{
		Use:   "cpu",
		Short: "Collect CPU profile",
		Long:  `Collect CPU profile for the specified duration and save to file.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			profiler := simpleprof.NewSimpleProfiler("tmp/profiles")

			fmt.Printf("🔄 Starting CPU profiling for %v...\n", duration)
			filename, err := profiler.StartProfile(simpleprof.ProfileTypeCPU, duration)
			if err != nil {
				return fmt.Errorf("failed to start CPU profile: %w", err)
			}

			// Wait for profiling to complete
			time.Sleep(duration)
			fmt.Printf("✅ CPU profile saved to: %s\n", filename)
			fmt.Printf("📊 Analyze with: go tool pprof %s\n", filename)

			return nil
		},
	}

	cmd.Flags().DurationVar(&duration, "duration", 30*time.Second, "Profiling duration")
	return cmd
}

func newSimpleMemoryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "memory",
		Short: "Collect memory profile",
		Long:  `Collect current memory profile and save to file.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			profiler := simpleprof.NewSimpleProfiler("tmp/profiles")

			fmt.Println("🔄 Collecting memory profile...")
			filename, err := profiler.StartProfile(simpleprof.ProfileTypeMemory, 0)
			if err != nil {
				return fmt.Errorf("failed to collect memory profile: %w", err)
			}

			fmt.Printf("✅ Memory profile saved to: %s\n", filename)
			fmt.Printf("📊 Analyze with: go tool pprof %s\n", filename)

			return nil
		},
	}

	return cmd
}

func newSimpleStatsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show runtime statistics",
		Long:  `Display current runtime statistics including memory usage and goroutines.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			profiler := simpleprof.NewSimpleProfiler("tmp/profiles")
			stats := profiler.GetStats()

			fmt.Println("📊 Runtime Statistics:")
			fmt.Println("====================")

			if goroutines, ok := stats["goroutines"].(int); ok {
				fmt.Printf("Goroutines:        %d\n", goroutines)
			}

			if heapAlloc, ok := stats["heap_alloc"].(uint64); ok {
				fmt.Printf("Heap Allocated:    %s\n", formatBytes(heapAlloc))
			}

			if heapSys, ok := stats["heap_sys"].(uint64); ok {
				fmt.Printf("Heap System:       %s\n", formatBytes(heapSys))
			}

			if heapInuse, ok := stats["heap_inuse"].(uint64); ok {
				fmt.Printf("Heap In Use:       %s\n", formatBytes(heapInuse))
			}

			if stackInuse, ok := stats["stack_inuse"].(uint64); ok {
				fmt.Printf("Stack In Use:      %s\n", formatBytes(stackInuse))
			}

			if gcRuns, ok := stats["gc_runs"].(uint32); ok {
				fmt.Printf("GC Runs:           %d\n", gcRuns)
			}

			if lastGC, ok := stats["last_gc"].(time.Time); ok && !lastGC.IsZero() {
				fmt.Printf("Last GC:           %v\n", lastGC.Format("2006-01-02 15:04:05"))
			}

			return nil
		},
	}

	return cmd
}

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
