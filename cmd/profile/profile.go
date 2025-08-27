// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package profile

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/Gizzahub/gzh-cli/internal/app"
	"github.com/Gizzahub/gzh-cli/internal/simpleprof"
)

// NewProfileCmd creates a simplified profile command using standard Go pprof.
func NewProfileCmd(appCtx *app.AppContext) *cobra.Command {
	_ = appCtx
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Performance profiling using standard Go pprof",
		Long: `Simple performance profiling using standard Go pprof tools.

Available commands:
  server    Start pprof HTTP server
  cpu       Collect CPU profile
  memory    Collect memory profile
  stats     Show runtime statistics

Examples:
  gz profile server --port 6060
  gz profile cpu --duration 30s
  gz profile memory
  gz profile stats`,
	}

	// Add subcommands
	cmd.AddCommand(newSimpleServerCmd())
	cmd.AddCommand(newSimpleCPUCmd())
	cmd.AddCommand(newSimpleMemoryCmd())
	cmd.AddCommand(newSimpleStatsCmd())

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

			fmt.Printf("✅ Pprof server started on http://localhost:%d/debug/pprof/\n", port)
			fmt.Println("Press Ctrl+C to stop the server")

			// Wait for interrupt
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// This would normally wait for a signal, but for CLI we'll just show the message
			select {
			case <-ctx.Done():
				return profiler.StopHTTPServer(context.Background())
			}
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
