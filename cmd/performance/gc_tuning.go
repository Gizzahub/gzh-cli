package performance

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gizzahub/gzh-manager-go/pkg/memory"
	"github.com/spf13/cobra"
)

type gcTuningOptions struct {
	gcPercent               int
	memoryLimit             string
	profilingEnabled        bool
	profilingInterval       time.Duration
	forceGCInterval         time.Duration
	objectPoolingEnabled    bool
	memoryPressureThreshold float64
	outputDir               string
	workloadType            string
	duration                time.Duration
	captureProfiles         bool
}

func defaultGCTuningOptions() *gcTuningOptions {
	return &gcTuningOptions{
		gcPercent:               100,
		memoryLimit:             "",
		profilingEnabled:        true,
		profilingInterval:       30 * time.Second,
		forceGCInterval:         0,
		objectPoolingEnabled:    true,
		memoryPressureThreshold: 0.8,
		outputDir:               "./profiles",
		workloadType:            "balanced",
		duration:                5 * time.Minute,
		captureProfiles:         false,
	}
}

func newGCTuningCmd() *cobra.Command {
	o := defaultGCTuningOptions()

	cmd := &cobra.Command{
		Use:   "gc-tuning",
		Short: "Garbage collection tuning and memory optimization",
		Long: `Optimize garbage collection settings and monitor memory usage patterns.
Supports different workload types and provides detailed profiling capabilities.`,
		RunE: o.run,
	}

	cmd.Flags().IntVar(&o.gcPercent, "gc-percent", o.gcPercent, "GC percentage (default 100)")
	cmd.Flags().StringVar(&o.memoryLimit, "memory-limit", o.memoryLimit, "Memory limit (e.g., 1GB, 512MB)")
	cmd.Flags().BoolVar(&o.profilingEnabled, "profiling", o.profilingEnabled, "Enable memory profiling")
	cmd.Flags().DurationVar(&o.profilingInterval, "profiling-interval", o.profilingInterval, "Profiling interval")
	cmd.Flags().DurationVar(&o.forceGCInterval, "force-gc-interval", o.forceGCInterval, "Force GC interval (0 = disabled)")
	cmd.Flags().BoolVar(&o.objectPoolingEnabled, "object-pooling", o.objectPoolingEnabled, "Enable object pooling")
	cmd.Flags().Float64Var(&o.memoryPressureThreshold, "memory-pressure-threshold", o.memoryPressureThreshold, "Memory pressure threshold (0.0-1.0)")
	cmd.Flags().StringVar(&o.outputDir, "output-dir", o.outputDir, "Output directory for profiles")
	cmd.Flags().StringVar(&o.workloadType, "workload", o.workloadType, "Workload type (low-latency, high-throughput, memory-constrained, balanced)")
	cmd.Flags().DurationVar(&o.duration, "duration", o.duration, "Monitoring duration")
	cmd.Flags().BoolVar(&o.captureProfiles, "capture-profiles", o.captureProfiles, "Capture all profiling data")

	return cmd
}

func (o *gcTuningOptions) run(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	// Parse memory limit
	var memoryLimitBytes int64
	if o.memoryLimit != "" {
		limit, err := parseMemoryLimit(o.memoryLimit)
		if err != nil {
			return fmt.Errorf("invalid memory limit: %w", err)
		}
		memoryLimitBytes = limit
	}

	// Create GC tuner configuration
	gcConfig := memory.GCConfig{
		GCPercent:               o.gcPercent,
		MemoryLimit:             memoryLimitBytes,
		ProfilingEnabled:        o.profilingEnabled,
		ProfilingInterval:       o.profilingInterval,
		ForceGCInterval:         o.forceGCInterval,
		ObjectPoolingEnabled:    o.objectPoolingEnabled,
		MemoryPressureThreshold: o.memoryPressureThreshold,
	}

	// Create and start GC tuner
	tuner := memory.NewGCTuner(gcConfig)
	fmt.Printf("🔧 Starting GC tuning with configuration:\n")
	fmt.Printf("   GC Percent: %d%%\n", o.gcPercent)
	fmt.Printf("   Memory Limit: %s\n", formatMemoryLimit(memoryLimitBytes))
	fmt.Printf("   Workload Type: %s\n", o.workloadType)
	fmt.Printf("   Object Pooling: %v\n", o.objectPoolingEnabled)
	fmt.Printf("   Memory Pressure Threshold: %.1f%%\n", o.memoryPressureThreshold*100)

	// Optimize for workload type
	tuner.OptimizeForWorkload(o.workloadType)

	if err := tuner.Start(ctx); err != nil {
		return fmt.Errorf("failed to start GC tuner: %w", err)
	}
	defer tuner.Stop()

	// Create profiler if requested
	var profiler *memory.Profiler
	if o.captureProfiles {
		profilerConfig := memory.ProfilerConfig{
			OutputDir:               o.outputDir,
			CPUProfileDuration:      30 * time.Second,
			MemProfileInterval:      1 * time.Minute,
			TraceEnabled:            true,
			TraceDuration:           10 * time.Second,
			BlockProfileEnabled:     true,
			BlockProfileRate:        1,
			MutexProfileEnabled:     true,
			MutexProfileFraction:    1,
			GoroutineProfileEnabled: true,
			FilePrefix:              "gc-tuning",
		}

		profiler = memory.NewProfiler(profilerConfig)
		if err := profiler.Start(ctx); err != nil {
			return fmt.Errorf("failed to start profiler: %w", err)
		}
		defer profiler.Stop()

		fmt.Printf("📊 Profiling enabled, output directory: %s\n", o.outputDir)
	}

	// Create some memory pools for demonstration
	if o.objectPoolingEnabled {
		fmt.Printf("🔄 Creating memory pools for optimization...\n")

		// Create buffer pool
		bufferPool := tuner.CreatePool("demo-buffers",
			func() interface{} {
				return make([]byte, 0, 1024)
			},
			func(obj interface{}) {
				if slice, ok := obj.([]byte); ok {
					slice = slice[:0] // Reset length but keep capacity
				}
			})

		// Create string pool
		stringPool := tuner.CreatePool("demo-strings",
			func() interface{} {
				return make([]string, 0, 64)
			},
			func(obj interface{}) {
				if slice, ok := obj.([]string); ok {
					slice = slice[:0] // Reset length but keep capacity
				}
			})

		// Demo pool usage
		go o.demonstratePoolUsage(ctx, bufferPool, stringPool)
	}

	// Monitor for the specified duration
	fmt.Printf("⏱️  Monitoring for %v...\n", o.duration)

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	startTime := time.Now()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			elapsed := time.Since(startTime)
			if elapsed >= o.duration {
				fmt.Printf("\n✅ Monitoring completed after %v\n", elapsed)
				goto summary
			}

			// Print current stats
			stats := tuner.GetStats()
			fmt.Printf("📈 Memory: %s heap, %d GCs, %.1f%% pressure, %v avg pause\n",
				formatBytes(stats.HeapAlloc),
				stats.NumGC,
				stats.MemoryPressure*100,
				time.Duration(stats.AveragePauseNs))

			if profiler != nil {
				analysis := profiler.AnalyzeMemoryUsage()
				fmt.Printf("🔍 Analysis: %d goroutines, %.1f%% heap utilization, %.2f KB/s growth\n",
					analysis.Goroutines,
					analysis.HeapUsage.Utilization*100,
					analysis.MemoryGrowthRate/1024)
			}
		}
	}

summary:
	// Print final summary
	fmt.Printf("\n📊 Final Statistics:\n")
	tuner.PrintStats()

	if profiler != nil {
		fmt.Printf("\n🎯 Final Memory Analysis:\n")
		analysis := profiler.AnalyzeMemoryUsage()
		printMemoryAnalysis(analysis)
	}

	// Capture final profiles if requested
	if o.captureProfiles {
		fmt.Printf("\n📸 Capturing final profiles...\n")
		if err := profiler.CaptureAllProfiles(); err != nil {
			fmt.Printf("Warning: Failed to capture some profiles: %v\n", err)
		}
	}

	return nil
}

func (o *gcTuningOptions) demonstratePoolUsage(ctx context.Context, bufferPool, stringPool *memory.MemoryPool) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Use buffer pool
			for i := 0; i < 10; i++ {
				buf := bufferPool.Get().([]byte)
				// Simulate some work
				buf = append(buf, []byte("some data")...)
				bufferPool.Put(buf)
			}

			// Use string pool
			for i := 0; i < 5; i++ {
				strings := stringPool.Get().([]string)
				// Simulate some work
				strings = append(strings, "test", "data", "example")
				stringPool.Put(strings)
			}
		}
	}
}

func parseMemoryLimit(limit string) (int64, error) {
	// Simple parser for memory limits like "1GB", "512MB", etc.
	var multiplier int64 = 1
	var numberStr string

	limit = strings.ToUpper(limit)
	if strings.HasSuffix(limit, "GB") {
		multiplier = 1024 * 1024 * 1024
		numberStr = strings.TrimSuffix(limit, "GB")
	} else if strings.HasSuffix(limit, "MB") {
		multiplier = 1024 * 1024
		numberStr = strings.TrimSuffix(limit, "MB")
	} else if strings.HasSuffix(limit, "KB") {
		multiplier = 1024
		numberStr = strings.TrimSuffix(limit, "KB")
	} else {
		numberStr = limit
	}

	number, err := strconv.ParseInt(numberStr, 10, 64)
	if err != nil {
		return 0, err
	}

	return number * multiplier, nil
}

func formatMemoryLimit(bytes int64) string {
	if bytes == 0 {
		return "unlimited"
	}

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

func printMemoryAnalysis(analysis memory.MemoryAnalysis) {
	fmt.Printf("Timestamp: %s\n", analysis.Timestamp.Format(time.RFC3339))
	fmt.Printf("Heap Usage:\n")
	fmt.Printf("  Allocated: %s\n", formatBytes(analysis.HeapUsage.Allocated))
	fmt.Printf("  System: %s\n", formatBytes(analysis.HeapUsage.System))
	fmt.Printf("  In Use: %s\n", formatBytes(analysis.HeapUsage.InUse))
	fmt.Printf("  Released: %s\n", formatBytes(analysis.HeapUsage.Released))
	fmt.Printf("  Objects: %d\n", analysis.HeapUsage.Objects)
	fmt.Printf("  Utilization: %.1f%%\n", analysis.HeapUsage.Utilization*100)
	fmt.Printf("GC Stats:\n")
	fmt.Printf("  Number of GCs: %d\n", analysis.GCStats.NumGC)
	fmt.Printf("  Total Pause: %v\n", analysis.GCStats.TotalPause)
	fmt.Printf("  Average Pause: %v\n", analysis.GCStats.AvgPause)
	fmt.Printf("  Last GC: %s\n", analysis.GCStats.LastGC.Format(time.RFC3339))
	fmt.Printf("Goroutines: %d\n", analysis.Goroutines)
	if analysis.MemoryGrowthRate != 0 {
		fmt.Printf("Memory Growth Rate: %.2f KB/s\n", analysis.MemoryGrowthRate/1024)
	}
}
