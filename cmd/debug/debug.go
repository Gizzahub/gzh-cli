package debug

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/gizzahub/gzh-manager-go/pkg/debug"
	"github.com/spf13/cobra"
)

// DebugCmd represents the debug command
var DebugCmd = &cobra.Command{
	Use:   "debug",
	Short: "Advanced debugging and profiling tools",
	Long: `Advanced debugging and profiling tools for GZH Manager.

Provides comprehensive debugging capabilities including:
- Performance profiling (CPU, memory, goroutines)
- Execution tracing and analysis
- Enhanced logging with multiple levels
- Memory usage monitoring
- Real-time metrics collection

Examples:
  gz debug profile --cpu --memory --duration 30s
  gz debug trace --output trace.json --categories api,bulk-clone
  gz debug log --level debug --format json --file debug.log
  gz debug memory --interval 5s --duration 1m`,
}

var (
	// Profile command flags
	profileCPU       bool
	profileMemory    bool
	profileGoroutine bool
	profileBlock     bool
	profileMutex     bool
	profileOutput    string
	profileDuration  time.Duration
	profileInterval  time.Duration
	profileHTTP      string

	// Trace command flags
	traceOutput     string
	traceMaxEvents  int
	traceStack      bool
	traceStackDepth int
	traceCategories []string
	traceDuration   time.Duration

	// Log command flags
	logLevel  string
	logFile   string
	logFormat string
	logColor  bool
	logTrace  bool

	// Memory command flags
	memoryInterval time.Duration
	memoryDuration time.Duration
	memoryOutput   string
)

func init() {
	// Profile command
	profileCmd := &cobra.Command{
		Use:   "profile",
		Short: "Start performance profiling",
		Long: `Start performance profiling with various profile types.

Captures CPU, memory, goroutine, block, and mutex profiles for analysis.
Profiles can be analyzed with 'go tool pprof' for detailed performance insights.`,
		Run: runProfile,
	}

	profileCmd.Flags().BoolVar(&profileCPU, "cpu", true, "Enable CPU profiling")
	profileCmd.Flags().BoolVar(&profileMemory, "memory", true, "Enable memory profiling")
	profileCmd.Flags().BoolVar(&profileGoroutine, "goroutine", true, "Enable goroutine profiling")
	profileCmd.Flags().BoolVar(&profileBlock, "block", false, "Enable block profiling")
	profileCmd.Flags().BoolVar(&profileMutex, "mutex", false, "Enable mutex profiling")
	profileCmd.Flags().StringVar(&profileOutput, "output", "./debug-profiles", "Output directory for profiles")
	profileCmd.Flags().DurationVar(&profileDuration, "duration", 30*time.Second, "Profiling duration")
	profileCmd.Flags().DurationVar(&profileInterval, "interval", 5*time.Second, "Profiling interval")
	profileCmd.Flags().StringVar(&profileHTTP, "http", ":6060", "HTTP endpoint for pprof (empty to disable)")

	// Trace command
	traceCmd := &cobra.Command{
		Use:   "trace",
		Short: "Start execution tracing",
		Long: `Start execution tracing to capture detailed program execution.

Generates Chrome trace format files that can be viewed in chrome://tracing
for detailed execution timeline analysis.`,
		Run: runTrace,
	}

	traceCmd.Flags().StringVar(&traceOutput, "output", "./debug-trace.json", "Output file for trace data")
	traceCmd.Flags().IntVar(&traceMaxEvents, "max-events", 100000, "Maximum number of trace events")
	traceCmd.Flags().BoolVar(&traceStack, "stack", false, "Include stack traces in events")
	traceCmd.Flags().IntVar(&traceStackDepth, "stack-depth", 10, "Maximum stack trace depth")
	traceCmd.Flags().StringSliceVar(&traceCategories, "categories", []string{"default"}, "Trace event categories")
	traceCmd.Flags().DurationVar(&traceDuration, "duration", 30*time.Second, "Tracing duration")

	// Log command
	logCmd := &cobra.Command{
		Use:   "log",
		Short: "Configure enhanced logging",
		Long: `Configure enhanced logging with multiple levels and formats.

Supports structured logging with JSON format, file output, and detailed
trace information including file locations and function names.`,
		Run: runLog,
	}

	logCmd.Flags().StringVar(&logLevel, "level", "debug", "Log level (trace, debug, info, warn, error, silent)")
	logCmd.Flags().StringVar(&logFile, "file", "", "Log file path (empty for stderr)")
	logCmd.Flags().StringVar(&logFormat, "format", "text", "Log format (text or json)")
	logCmd.Flags().BoolVar(&logColor, "color", true, "Enable colored output")
	logCmd.Flags().BoolVar(&logTrace, "trace", true, "Include trace information")

	// Memory command
	memoryCmd := &cobra.Command{
		Use:   "memory",
		Short: "Monitor memory usage",
		Long: `Monitor memory usage with periodic sampling.

Captures memory statistics at regular intervals and generates
detailed reports showing memory usage patterns over time.`,
		Run: runMemory,
	}

	memoryCmd.Flags().DurationVar(&memoryInterval, "interval", 5*time.Second, "Memory sampling interval")
	memoryCmd.Flags().DurationVar(&memoryDuration, "duration", 1*time.Minute, "Monitoring duration")
	memoryCmd.Flags().StringVar(&memoryOutput, "output", "./memory-report.json", "Output file for memory report")

	// Add subcommands
	DebugCmd.AddCommand(profileCmd)
	DebugCmd.AddCommand(traceCmd)
	DebugCmd.AddCommand(logCmd)
	DebugCmd.AddCommand(memoryCmd)
}

func runProfile(cmd *cobra.Command, args []string) {
	fmt.Println("🔍 Starting performance profiling...")

	// Create profiler configuration
	config := &debug.ProfilerConfig{
		Enabled:        true,
		CPUProfile:     profileCPU,
		MemoryProfile:  profileMemory,
		GoroutineTrace: profileGoroutine,
		BlockProfile:   profileBlock,
		MutexProfile:   profileMutex,
		OutputDir:      profileOutput,
		Duration:       profileDuration,
		Interval:       profileInterval,
		HTTPEndpoint:   profileHTTP,
	}

	// Create and start profiler
	profiler := debug.NewProfiler(config)
	ctx, cancel := context.WithTimeout(context.Background(), profileDuration)
	defer cancel()

	if err := profiler.Start(ctx); err != nil {
		fmt.Printf("❌ Failed to start profiler: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("📊 Profiling for %v...\n", profileDuration)
	if profileHTTP != "" {
		fmt.Printf("🌐 Profiler endpoint: http://localhost%s/debug/pprof/\n", profileHTTP)
	}

	// Wait for profiling to complete
	<-ctx.Done()

	if err := profiler.Stop(); err != nil {
		fmt.Printf("⚠️ Error stopping profiler: %v\n", err)
	}

	// Generate report
	reportPath, err := profiler.GenerateReport()
	if err != nil {
		fmt.Printf("⚠️ Failed to generate report: %v\n", err)
	} else {
		fmt.Printf("📊 Report generated: %s\n", reportPath)
	}

	fmt.Println("✅ Profiling completed")
}

func runTrace(cmd *cobra.Command, args []string) {
	fmt.Println("🔎 Starting execution tracing...")

	// Create tracer configuration
	config := &debug.TracerConfig{
		Enabled:       true,
		OutputFile:    traceOutput,
		MaxEvents:     traceMaxEvents,
		IncludeStack:  traceStack,
		StackDepth:    traceStackDepth,
		BufferSize:    1000,
		FlushInterval: 2 * time.Second,
		Categories:    traceCategories,
	}

	// Create and start tracer
	tracer := debug.NewTracer(config)
	ctx, cancel := context.WithTimeout(context.Background(), traceDuration)
	defer cancel()

	if err := tracer.Start(ctx); err != nil {
		fmt.Printf("❌ Failed to start tracer: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("📊 Tracing for %v...\n", traceDuration)
	fmt.Printf("💾 Output file: %s\n", traceOutput)

	// Add some sample trace events
	tracer.Instant("trace_started", "debug", map[string]interface{}{
		"duration":   traceDuration.String(),
		"categories": traceCategories,
	})

	// Wait for tracing to complete
	<-ctx.Done()

	tracer.Instant("trace_completed", "debug")

	if err := tracer.Stop(); err != nil {
		fmt.Printf("⚠️ Error stopping tracer: %v\n", err)
	}

	fmt.Printf("🌐 View trace at: chrome://tracing (load %s)\n", traceOutput)
	fmt.Println("✅ Tracing completed")
}

func runLog(cmd *cobra.Command, args []string) {
	fmt.Println("📝 Configuring enhanced logging...")

	// Parse log level
	level, err := debug.ParseLogLevel(logLevel)
	if err != nil {
		fmt.Printf("❌ Invalid log level: %v\n", err)
		os.Exit(1)
	}

	// Create logger configuration
	config := &debug.LoggerConfig{
		Level:       level,
		File:        logFile,
		EnableColor: logColor,
		EnableTrace: logTrace,
		MaxFileSize: 100 * 1024 * 1024, // 100MB
		MaxBackups:  5,
		Compress:    true,
		Format:      logFormat,
	}

	// Initialize global logger
	if err := debug.InitGlobalLogger(config); err != nil {
		fmt.Printf("❌ Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Enhanced logging configured\n")
	fmt.Printf("  Level: %s\n", logLevel)
	fmt.Printf("  Format: %s\n", logFormat)
	if logFile != "" {
		fmt.Printf("  File: %s\n", logFile)
	} else {
		fmt.Printf("  Output: stderr\n")
	}
	fmt.Printf("  Color: %v\n", logColor)
	fmt.Printf("  Trace: %v\n", logTrace)

	// Demonstrate logging levels
	fmt.Println("\nDemonstrating log levels:")
	debug.Error("This is an error message", map[string]interface{}{"component": "demo"})
	debug.Warn("This is a warning message", map[string]interface{}{"component": "demo"})
	debug.Info("This is an info message", map[string]interface{}{"component": "demo"})
	debug.Debug("This is a debug message", map[string]interface{}{"component": "demo"})
	debug.Trace("This is a trace message", map[string]interface{}{"component": "demo"})
}

func runMemory(cmd *cobra.Command, args []string) {
	fmt.Println("📊 Starting memory monitoring...")

	fmt.Printf("  Interval: %v\n", memoryInterval)
	fmt.Printf("  Duration: %v\n", memoryDuration)
	fmt.Printf("  Output: %s\n", memoryOutput)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), memoryDuration)
	defer cancel()

	// Memory sampling data
	type memorySample struct {
		Timestamp time.Time              `json:"timestamp"`
		Stats     map[string]interface{} `json:"stats"`
	}

	var samples []memorySample
	ticker := time.NewTicker(memoryInterval)
	defer ticker.Stop()

	// Initial sample
	stats := debug.ProfileMemoryUsage()
	samples = append(samples, memorySample{
		Timestamp: time.Now(),
		Stats:     stats,
	})

	fmt.Printf("🔍 Initial memory: %.2f MB allocated, %d goroutines\n",
		stats["allocated_mb"], stats["goroutines"])

	sampleCount := 1
	for {
		select {
		case <-ctx.Done():
			goto done
		case <-ticker.C:
			stats := debug.ProfileMemoryUsage()
			samples = append(samples, memorySample{
				Timestamp: time.Now(),
				Stats:     stats,
			})
			sampleCount++

			fmt.Printf("📊 Sample %d: %.2f MB allocated, %d goroutines\n",
				sampleCount, stats["allocated_mb"], stats["goroutines"])
		}
	}

done:
	// Generate report
	report := map[string]interface{}{
		"start_time":      samples[0].Timestamp,
		"end_time":        time.Now(),
		"duration":        memoryDuration.String(),
		"sample_count":    len(samples),
		"sample_interval": memoryInterval.String(),
		"samples":         samples,
	}

	// Calculate statistics
	if len(samples) > 1 {
		firstMem := samples[0].Stats["allocated_mb"].(float64)
		lastMem := samples[len(samples)-1].Stats["allocated_mb"].(float64)
		memDelta := lastMem - firstMem

		firstGoroutines := samples[0].Stats["goroutines"].(int)
		lastGoroutines := samples[len(samples)-1].Stats["goroutines"].(int)
		goroutineDelta := lastGoroutines - firstGoroutines

		report["summary"] = map[string]interface{}{
			"memory_delta_mb":    memDelta,
			"goroutine_delta":    goroutineDelta,
			"initial_memory_mb":  firstMem,
			"final_memory_mb":    lastMem,
			"initial_goroutines": firstGoroutines,
			"final_goroutines":   lastGoroutines,
		}

		fmt.Printf("\n📊 Memory Summary:\n")
		fmt.Printf("  Initial: %.2f MB, %d goroutines\n", firstMem, firstGoroutines)
		fmt.Printf("  Final: %.2f MB, %d goroutines\n", lastMem, lastGoroutines)
		fmt.Printf("  Delta: %+.2f MB, %+d goroutines\n", memDelta, goroutineDelta)
	}

	// Write report to file
	if err := writeJSONReport(memoryOutput, report); err != nil {
		fmt.Printf("❌ Failed to write memory report: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("💾 Memory report saved: %s\n", memoryOutput)
	fmt.Println("✅ Memory monitoring completed")
}

func writeJSONReport(filename string, data interface{}) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}
