package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// newPerformanceCmd creates the performance analysis command
func newPerformanceCmd(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "performance",
		Short: "Performance analysis and optimization tools",
		Long: `Performance analysis and optimization tools providing:
- Runtime profiling integration (CPU, memory, goroutines)
- Bottleneck detection and analysis
- Performance optimization suggestions
- Trend analysis and reporting
- Real-time performance monitoring

Examples:
  # Start performance profiling server
  gz monitoring performance profile --enable --port 6060
  
  # Generate performance report
  gz monitoring performance report --period 1h --output json
  
  # Analyze current system performance
  gz monitoring performance analyze --component runtime
  
  # Start continuous performance monitoring
  gz monitoring performance monitor --interval 30s`,
		SilenceUsage: true,
	}

	cmd.AddCommand(newPerformanceProfileCmd(ctx))
	cmd.AddCommand(newPerformanceReportCmd(ctx))
	cmd.AddCommand(newPerformanceAnalyzeCmd(ctx))
	cmd.AddCommand(newPerformanceMonitorCmd(ctx))

	return cmd
}

// newPerformanceProfileCmd creates the profiling command
func newPerformanceProfileCmd(ctx context.Context) *cobra.Command {
	var (
		enable           bool
		port             int
		cpuProfile       bool
		memProfile       bool
		blockProfile     bool
		goroutineProfile bool
		sampleRate       string
		profileDuration  string
	)

	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Start performance profiling server",
		Long: `Start performance profiling server with pprof endpoints.

This command starts an HTTP server that exposes performance profiling endpoints:
- /debug/pprof/ - Profiling index
- /debug/pprof/profile - CPU profiling
- /debug/pprof/heap - Memory profiling
- /debug/pprof/goroutine - Goroutine profiling
- /debug/pprof/block - Block profiling
- /debug/analysis/* - Custom analysis endpoints

Examples:
  # Start profiling server on default port
  gz monitoring performance profile --enable
  
  # Start with custom port and specific profiles
  gz monitoring performance profile --enable --port 6060 --cpu --memory
  
  # Start with custom sample rate
  gz monitoring performance profile --enable --sample-rate 100ms`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !enable {
				fmt.Println("Performance profiling is disabled. Use --enable to start profiling server.")
				return nil
			}

			// Parse durations
			sampleRateDuration, err := time.ParseDuration(sampleRate)
			if err != nil {
				return fmt.Errorf("invalid sample rate: %w", err)
			}

			profileDurationParsed, err := time.ParseDuration(profileDuration)
			if err != nil {
				return fmt.Errorf("invalid profile duration: %w", err)
			}

			// Create profiling configuration
			config := &ProfilingConfig{
				Enabled:            enable,
				ListenAddress:      fmt.Sprintf(":%d", port),
				CPUProfiling:       cpuProfile,
				MemoryProfiling:    memProfile,
				BlockProfiling:     blockProfile,
				GoroutineProfiling: goroutineProfile,
				SampleRate:         sampleRateDuration,
				ProfileDuration:    profileDurationParsed,
			}

			// Create logger
			logger, _ := zap.NewDevelopment()
			defer logger.Sync()

			// Create prometheus registry
			registry := prometheus.NewRegistry()

			// Create performance analyzer
			pa := NewPerformanceAnalyzer(logger, registry, nil, config)

			fmt.Printf("🔬 Performance profiling server started on port %d\n", port)
			fmt.Printf("📊 Profiling endpoints available at:\n")
			fmt.Printf("   http://localhost:%d/debug/pprof/\n", port)
			fmt.Printf("   http://localhost:%d/debug/analysis/performance\n", port)
			fmt.Printf("   http://localhost:%d/debug/analysis/runtime\n", port)
			fmt.Printf("   http://localhost:%d/debug/analysis/memory\n", port)
			fmt.Printf("   http://localhost:%d/debug/analysis/goroutines\n", port)

			// Wait for context cancellation
			<-ctx.Done()

			fmt.Println("\n🛑 Stopping performance profiling server...")
			return pa.Stop(context.Background())
		},
	}

	cmd.Flags().BoolVar(&enable, "enable", false, "Enable performance profiling server")
	cmd.Flags().IntVar(&port, "port", 6060, "Port for profiling server")
	cmd.Flags().BoolVar(&cpuProfile, "cpu", true, "Enable CPU profiling")
	cmd.Flags().BoolVar(&memProfile, "memory", true, "Enable memory profiling")
	cmd.Flags().BoolVar(&blockProfile, "block", false, "Enable block profiling")
	cmd.Flags().BoolVar(&goroutineProfile, "goroutine", true, "Enable goroutine profiling")
	cmd.Flags().StringVar(&sampleRate, "sample-rate", "100ms", "Profiling sample rate")
	cmd.Flags().StringVar(&profileDuration, "profile-duration", "30s", "Default profile duration")

	return cmd
}

// newPerformanceReportCmd creates the performance report command
func newPerformanceReportCmd(ctx context.Context) *cobra.Command {
	var (
		period     string
		output     string
		file       string
		components []string
	)

	cmd := &cobra.Command{
		Use:   "report",
		Short: "Generate comprehensive performance analysis report",
		Long: `Generate comprehensive performance analysis report including:
- System performance metrics (CPU, memory, goroutines)
- Detected performance bottlenecks
- Optimization suggestions
- Performance trends and predictions
- Overall performance score

Examples:
  # Generate report for last hour
  gz monitoring performance report --period 1h
  
  # Generate detailed JSON report
  gz monitoring performance report --period 24h --output json --file report.json
  
  # Generate report for specific components
  gz monitoring performance report --components runtime,memory`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse analysis period
			analysisPeriod, err := time.ParseDuration(period)
			if err != nil {
				return fmt.Errorf("invalid period: %w", err)
			}

			// Create logger
			logger, _ := zap.NewDevelopment()
			defer logger.Sync()

			// Create prometheus registry
			registry := prometheus.NewRegistry()

			// Create performance analyzer with basic config
			config := &ProfilingConfig{
				Enabled: false, // Don't start profiling server for reports
			}
			pa := NewPerformanceAnalyzer(logger, registry, nil, config)

			// Generate performance report
			report, err := pa.GeneratePerformanceReport(ctx, analysisPeriod)
			if err != nil {
				return fmt.Errorf("failed to generate performance report: %w", err)
			}

			// Format and output report
			var outputData []byte
			switch output {
			case "json":
				outputData, err = json.MarshalIndent(report, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal report to JSON: %w", err)
				}
			case "table", "text":
				outputData = []byte(formatReportAsText(report))
			default:
				return fmt.Errorf("unsupported output format: %s", output)
			}

			// Write to file or stdout
			if file != "" {
				err = os.WriteFile(file, outputData, 0o644)
				if err != nil {
					return fmt.Errorf("failed to write report to file: %w", err)
				}
				fmt.Printf("📄 Performance report written to %s\n", file)
			} else {
				fmt.Println(string(outputData))
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&period, "period", "1h", "Analysis period (e.g., 1h, 24h, 7d)")
	cmd.Flags().StringVar(&output, "output", "table", "Output format (table, json)")
	cmd.Flags().StringVar(&file, "file", "", "Output file path (default: stdout)")
	cmd.Flags().StringSliceVar(&components, "components", []string{}, "Specific components to analyze")

	return cmd
}

// newPerformanceAnalyzeCmd creates the performance analysis command
func newPerformanceAnalyzeCmd(ctx context.Context) *cobra.Command {
	var (
		component string
		output    string
		detailed  bool
	)

	cmd := &cobra.Command{
		Use:   "analyze",
		Short: "Analyze current system performance",
		Long: `Analyze current system performance for specific components:
- runtime: Go runtime analysis (goroutines, GC, memory)
- memory: Detailed memory usage analysis
- goroutines: Goroutine lifecycle and leak detection
- cpu: CPU utilization and profiling analysis

Examples:
  # Analyze runtime performance
  gz monitoring performance analyze --component runtime
  
  # Detailed memory analysis
  gz monitoring performance analyze --component memory --detailed
  
  # Get JSON output for integration
  gz monitoring performance analyze --component runtime --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create logger
			logger, _ := zap.NewDevelopment()
			defer logger.Sync()

			// Create prometheus registry
			registry := prometheus.NewRegistry()

			// Create performance analyzer
			config := &ProfilingConfig{Enabled: false}
			pa := NewPerformanceAnalyzer(logger, registry, nil, config)

			var analysis map[string]interface{}
			var err error

			// Perform specific analysis based on component
			switch component {
			case "runtime":
				analysis = pa.analyzeRuntime(ctx)
			case "memory":
				analysis = pa.analyzeMemory(ctx)
			case "goroutines":
				analysis = pa.analyzeGoroutines(ctx)
			default:
				return fmt.Errorf("unsupported component: %s (supported: runtime, memory, goroutines)", component)
			}

			// Format output
			var outputData []byte
			switch output {
			case "json":
				outputData, err = json.MarshalIndent(analysis, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal analysis to JSON: %w", err)
				}
			case "table", "text":
				outputData = []byte(formatAnalysisAsText(component, analysis))
			default:
				return fmt.Errorf("unsupported output format: %s", output)
			}

			fmt.Println(string(outputData))
			return nil
		},
	}

	cmd.Flags().StringVar(&component, "component", "runtime", "Component to analyze (runtime, memory, goroutines)")
	cmd.Flags().StringVar(&output, "output", "table", "Output format (table, json)")
	cmd.Flags().BoolVar(&detailed, "detailed", false, "Include detailed analysis")

	return cmd
}

// newPerformanceMonitorCmd creates the performance monitoring command
func newPerformanceMonitorCmd(ctx context.Context) *cobra.Command {
	var (
		interval string
		duration string
		metrics  []string
	)

	cmd := &cobra.Command{
		Use:   "monitor",
		Short: "Start continuous performance monitoring",
		Long: `Start continuous performance monitoring that:
- Collects performance metrics at regular intervals
- Detects performance bottlenecks in real-time
- Generates optimization suggestions
- Provides trend analysis

Examples:
  # Monitor with default settings
  gz monitoring performance monitor
  
  # Monitor with custom interval
  gz monitoring performance monitor --interval 10s
  
  # Monitor specific metrics for limited time
  gz monitoring performance monitor --metrics cpu,memory --duration 5m`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse interval
			intervalDuration, err := time.ParseDuration(interval)
			if err != nil {
				return fmt.Errorf("invalid interval: %w", err)
			}

			// Parse duration if specified
			var monitorDuration time.Duration
			if duration != "" {
				monitorDuration, err = time.ParseDuration(duration)
				if err != nil {
					return fmt.Errorf("invalid duration: %w", err)
				}
			}

			// Create logger
			logger, _ := zap.NewDevelopment()
			defer logger.Sync()

			// Create prometheus registry
			registry := prometheus.NewRegistry()

			// Create performance analyzer
			config := &ProfilingConfig{Enabled: false}
			pa := NewPerformanceAnalyzer(logger, registry, nil, config)

			fmt.Printf("📊 Starting performance monitoring with %s interval\n", interval)
			if len(metrics) > 0 {
				fmt.Printf("🎯 Monitoring metrics: %v\n", metrics)
			}
			if duration != "" {
				fmt.Printf("⏱️  Monitoring duration: %s\n", duration)
			}

			// Create monitoring context
			monitorCtx := ctx
			if monitorDuration > 0 {
				var cancel context.CancelFunc
				monitorCtx, cancel = context.WithTimeout(ctx, monitorDuration)
				defer cancel()
			}

			// Start monitoring
			pa.StartPerformanceMonitoring(monitorCtx, intervalDuration)

			fmt.Println("🛑 Performance monitoring stopped")
			return nil
		},
	}

	cmd.Flags().StringVar(&interval, "interval", "30s", "Monitoring interval")
	cmd.Flags().StringVar(&duration, "duration", "", "Monitoring duration (default: unlimited)")
	cmd.Flags().StringSliceVar(&metrics, "metrics", []string{}, "Specific metrics to monitor")

	return cmd
}

// formatReportAsText formats a performance report as human-readable text
func formatReportAsText(report *PerformanceReport) string {
	var result strings.Builder

	result.WriteString("🔬 Performance Analysis Report\n")
	result.WriteString("==============================\n\n")

	result.WriteString(fmt.Sprintf("📅 Generated: %s\n", report.GeneratedAt.Format("2006-01-02 15:04:05")))
	result.WriteString(fmt.Sprintf("⏱️  Analysis Period: %s\n", report.AnalysisPeriod))
	result.WriteString(fmt.Sprintf("🎯 Overall Score: %.1f/100\n\n", report.OverallScore))

	// System Metrics
	result.WriteString("📊 System Metrics\n")
	result.WriteString("-----------------\n")
	result.WriteString(fmt.Sprintf("CPU Utilization: %.1f%%\n", report.SystemMetrics.CPUUtilization))
	result.WriteString(fmt.Sprintf("Memory Utilization: %.1f%%\n", report.SystemMetrics.MemoryUtilization))
	result.WriteString(fmt.Sprintf("Goroutine Count: %d\n", report.SystemMetrics.GoroutineCount))
	result.WriteString(fmt.Sprintf("GC Pause Time: %s\n", report.SystemMetrics.GCPauseTime))
	result.WriteString(fmt.Sprintf("Response Time: %s\n", report.SystemMetrics.ResponseTime))
	result.WriteString(fmt.Sprintf("Throughput: %.1f req/s\n", report.SystemMetrics.Throughput))
	result.WriteString(fmt.Sprintf("Error Rate: %.1f%%\n\n", report.SystemMetrics.ErrorRate))

	// Bottlenecks
	if len(report.DetectedBottlenecks) > 0 {
		result.WriteString("⚠️  Detected Bottlenecks\n")
		result.WriteString("------------------------\n")
		for i, bottleneck := range report.DetectedBottlenecks {
			result.WriteString(fmt.Sprintf("%d. %s (%s)\n", i+1, bottleneck.Name, bottleneck.Severity))
			result.WriteString(fmt.Sprintf("   Component: %s\n", bottleneck.Component))
			result.WriteString(fmt.Sprintf("   Current: %.2f | Threshold: %.2f\n", bottleneck.CurrentValue, bottleneck.Threshold))
			result.WriteString(fmt.Sprintf("   Impact: %s | Duration: %s\n", bottleneck.Impact, bottleneck.Duration))
		}
		result.WriteString("\n")
	}

	// Optimizations
	if len(report.Optimizations) > 0 {
		result.WriteString("🚀 Optimization Suggestions\n")
		result.WriteString("---------------------------\n")
		for i, opt := range report.Optimizations {
			if i >= 5 { // Show top 5 optimizations
				break
			}
			result.WriteString(fmt.Sprintf("%d. %s (%s impact, %s effort)\n", i+1, opt.Title, opt.Impact, opt.Effort))
			result.WriteString(fmt.Sprintf("   %s\n", opt.Description))
			result.WriteString(fmt.Sprintf("   Category: %s | Estimated Gain: %.1f%%\n", opt.Category, opt.EstimatedGain))
		}
		result.WriteString("\n")
	}

	// Recommendations
	if len(report.Recommendations) > 0 {
		result.WriteString("💡 Recommendations\n")
		result.WriteString("------------------\n")
		for i, rec := range report.Recommendations {
			result.WriteString(fmt.Sprintf("• %s\n", rec))
			if i >= 4 { // Show top 5 recommendations
				break
			}
		}
	}

	return result.String()
}

// formatAnalysisAsText formats analysis results as human-readable text
func formatAnalysisAsText(component string, analysis map[string]interface{}) string {
	var result strings.Builder

	result.WriteString(fmt.Sprintf("🔍 %s Analysis\n", strings.Title(component)))
	result.WriteString(strings.Repeat("=", len(component)+11) + "\n\n")

	switch component {
	case "runtime":
		if goroutines, ok := analysis["goroutines"].(int); ok {
			result.WriteString(fmt.Sprintf("Goroutines: %d\n", goroutines))
		}
		if cpuCores, ok := analysis["cpu_cores"].(int); ok {
			result.WriteString(fmt.Sprintf("CPU Cores: %d\n", cpuCores))
		}
		if score, ok := analysis["performance_score"].(float64); ok {
			result.WriteString(fmt.Sprintf("Performance Score: %.1f/100\n", score))
		}

	case "memory":
		if heap, ok := analysis["heap"].(map[string]interface{}); ok {
			if allocated, ok := heap["allocated"].(uint64); ok {
				result.WriteString(fmt.Sprintf("Heap Allocated: %s\n", formatBytesForAnalysis(allocated)))
			}
			if utilization, ok := heap["utilization"].(float64); ok {
				result.WriteString(fmt.Sprintf("Heap Utilization: %.1f%%\n", utilization))
			}
		}

	case "goroutines":
		if count, ok := analysis["current_count"].(int); ok {
			result.WriteString(fmt.Sprintf("Current Count: %d\n", count))
		}
		if status, ok := analysis["status"].(string); ok {
			result.WriteString(fmt.Sprintf("Health Status: %s\n", status))
		}
		if trend, ok := analysis["trend"].(string); ok {
			result.WriteString(fmt.Sprintf("Trend: %s\n", trend))
		}
	}

	// Add recommendations if available
	if recommendations, ok := analysis["recommendations"].([]string); ok && len(recommendations) > 0 {
		result.WriteString("\n💡 Recommendations:\n")
		for _, rec := range recommendations {
			result.WriteString(fmt.Sprintf("• %s\n", rec))
		}
	}

	return result.String()
}

// formatBytesForAnalysis formats byte count as human-readable string
func formatBytesForAnalysis(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
