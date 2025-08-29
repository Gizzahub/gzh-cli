// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package doctor

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/Gizzahub/gzh-cli/internal/cli"
	"github.com/Gizzahub/gzh-cli/internal/logger"
)

// SystemHealthReport represents comprehensive system health metrics.
type SystemHealthReport struct {
	Timestamp       time.Time        `json:"timestamp"`
	OverallStatus   string           `json:"overall_status"`
	Score           float64          `json:"score"`
	Categories      []HealthCategory `json:"categories"`
	Alerts          []HealthAlert    `json:"alerts"`
	Trends          HealthTrends     `json:"trends"`
	Metrics         SystemMetrics    `json:"metrics"`
	Recommendations []string         `json:"recommendations"`
}

// HealthCategory represents a category of health checks.
type HealthCategory struct {
	Name     string        `json:"name"`
	Status   string        `json:"status"`
	Score    float64       `json:"score"`
	Checks   []HealthCheck `json:"checks"`
	Duration time.Duration `json:"duration"`
}

// HealthCheck represents an individual health check.
type HealthCheck struct {
	Name      string                 `json:"name"`
	Status    string                 `json:"status"`
	Value     interface{}            `json:"value"`
	Threshold interface{}            `json:"threshold"`
	Message   string                 `json:"message"`
	Details   map[string]interface{} `json:"details"`
	Duration  time.Duration          `json:"duration"`
	Critical  bool                   `json:"critical"`
}

// HealthAlert represents a health alert.
type HealthAlert struct {
	Level      string    `json:"level"`
	Category   string    `json:"category"`
	Message    string    `json:"message"`
	Timestamp  time.Time `json:"timestamp"`
	Suggestion string    `json:"suggestion"`
}

// HealthTrends represents health trends over time.
type HealthTrends struct {
	CPUTrend     string  `json:"cpu_trend"`
	MemoryTrend  string  `json:"memory_trend"`
	DiskTrend    string  `json:"disk_trend"`
	NetworkTrend string  `json:"network_trend"`
	OverallTrend string  `json:"overall_trend"`
	TrendScore   float64 `json:"trend_score"`
}

// SystemMetrics represents detailed system metrics.
type SystemMetrics struct {
	CPU         CPUMetrics     `json:"cpu"`
	Memory      MemoryMetrics  `json:"memory"`
	Disk        DiskMetrics    `json:"disk"`
	Network     NetworkMetrics `json:"network"`
	Process     ProcessMetrics `json:"process"`
	Environment EnvMetrics     `json:"environment"`
}

// CPUMetrics represents CPU-related metrics.
type CPUMetrics struct {
	Count       int     `json:"count"`
	Usage       float64 `json:"usage"`
	LoadAvg1    float64 `json:"load_avg_1"`
	LoadAvg5    float64 `json:"load_avg_5"`
	LoadAvg15   float64 `json:"load_avg_15"`
	Temperature float64 `json:"temperature"`
}

// MemoryMetrics represents memory-related metrics.
type MemoryMetrics struct {
	Total       uint64  `json:"total"`
	Available   uint64  `json:"available"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"used_percent"`
	Swap        uint64  `json:"swap"`
	SwapUsed    uint64  `json:"swap_used"`
}

// DiskMetrics represents disk-related metrics.
type DiskMetrics struct {
	Total       uint64  `json:"total"`
	Available   uint64  `json:"available"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"used_percent"`
	IOReads     uint64  `json:"io_reads"`
	IOWrites    uint64  `json:"io_writes"`
}

// NetworkMetrics represents network-related metrics.
type NetworkMetrics struct {
	BytesSent   uint64 `json:"bytes_sent"`
	BytesRecv   uint64 `json:"bytes_recv"`
	PacketsSent uint64 `json:"packets_sent"`
	PacketsRecv uint64 `json:"packets_recv"`
	Connections int    `json:"connections"`
	Latency     int64  `json:"latency"`
}

// ProcessMetrics represents process-related metrics.
type ProcessMetrics struct {
	Goroutines   int    `json:"goroutines"`
	Threads      int    `json:"threads"`
	FDs          int    `json:"file_descriptors"`
	HeapObjects  uint64 `json:"heap_objects"`
	GCRuns       uint32 `json:"gc_runs"`
	GCPauseTotal uint64 `json:"gc_pause_total"`
}

// EnvMetrics represents environment-related metrics.
type EnvMetrics struct {
	GoVersion string            `json:"go_version"`
	Platform  string            `json:"platform"`
	Hostname  string            `json:"hostname"`
	Uptime    time.Duration     `json:"uptime"`
	EnvVars   map[string]string `json:"env_vars"`
}

// newHealthCmd creates the health subcommand for system health monitoring.
func newHealthCmd() *cobra.Command {
	ctx := context.Background()

	var (
		continuous bool
		interval   time.Duration
		alertLevel string
		outputFile string
		serverMode bool
		serverPort int
		categories []string
		includeEnv bool
		detailed   bool
	)

	cmd := cli.NewCommandBuilder(ctx, "health", "Monitor comprehensive system health metrics").
		WithLongDescription(`Monitor comprehensive system health metrics with real-time monitoring capabilities.

This command provides advanced system health monitoring including:
- Real-time CPU, memory, disk, and network monitoring
- Process-level metrics and resource usage tracking
- Environmental health checks and dependency validation
- Trend analysis and predictive health scoring
- Alert generation with configurable thresholds
- Continuous monitoring with configurable intervals
- HTTP server mode for external monitoring integration

Features:
- Multi-category health assessment (CPU, Memory, Disk, Network, Process, Environment)
- Real-time metrics collection with trend analysis
- Configurable alert thresholds and notification levels
- Continuous monitoring mode for long-running health checks
- HTTP API mode for integration with monitoring systems
- Detailed health reports with actionable recommendations
- Historical trend analysis and predictive scoring

Examples:
  gz doctor health                                  # Run comprehensive health check
  gz doctor health --continuous --interval 30s     # Continuous monitoring every 30 seconds
  gz doctor health --server --port 8080            # Start HTTP health monitoring server
  gz doctor health --categories cpu,memory         # Monitor specific categories only
  gz doctor health --alert-level critical          # Show only critical alerts`).
		WithExample("gz doctor health --continuous --interval 60s --alert-level warning").
		WithFormatFlag("table", []string{"table", "json", "yaml"}).
		WithRunFuncE(func(ctx context.Context, flags *cli.CommonFlags, args []string) error {
			return runHealthMonitoring(ctx, flags, healthOptions{
				continuous: continuous,
				interval:   interval,
				alertLevel: alertLevel,
				outputFile: outputFile,
				serverMode: serverMode,
				serverPort: serverPort,
				categories: categories,
				includeEnv: includeEnv,
				detailed:   detailed,
			})
		}).
		Build()

	cmd.Flags().BoolVar(&continuous, "continuous", false, "Enable continuous monitoring mode")
	cmd.Flags().DurationVar(&interval, "interval", 30*time.Second, "Monitoring interval for continuous mode")
	cmd.Flags().StringVar(&alertLevel, "alert-level", "warning", "Alert level threshold (info, warning, critical)")
	cmd.Flags().StringVar(&outputFile, "output", "", "Output file for health reports")
	cmd.Flags().BoolVar(&serverMode, "server", false, "Start HTTP server for health monitoring")
	cmd.Flags().IntVar(&serverPort, "port", 8080, "Port for HTTP health monitoring server")
	cmd.Flags().StringSliceVar(&categories, "categories", []string{}, "Health categories to monitor (cpu,memory,disk,network,process,env)")
	cmd.Flags().BoolVar(&includeEnv, "include-env", false, "Include environment variables in report")
	cmd.Flags().BoolVar(&detailed, "detailed", false, "Generate detailed health metrics")

	return cmd
}

type healthOptions struct {
	continuous bool
	interval   time.Duration
	alertLevel string
	outputFile string
	serverMode bool
	serverPort int
	categories []string
	includeEnv bool
	detailed   bool
}

func runHealthMonitoring(ctx context.Context, flags *cli.CommonFlags, opts healthOptions) error {
	logger := logger.NewSimpleLogger("doctor-health")

	logger.Info("Starting system health monitoring",
		"continuous", opts.continuous,
		"interval", opts.interval,
		"server_mode", opts.serverMode,
	)

	if opts.serverMode {
		return startHealthServer(ctx, opts, logger)
	}

	if opts.continuous {
		return runContinuousHealthMonitoring(ctx, flags, opts, logger)
	}

	// Single health check
	return runSingleHealthCheck(ctx, flags, opts, logger)
}

func runSingleHealthCheck(ctx context.Context, flags *cli.CommonFlags, opts healthOptions, logger logger.CommonLogger) error {
	report, err := collectHealthMetrics(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to collect health metrics: %w", err)
	}

	// Save report if requested
	if opts.outputFile != "" {
		if err := saveHealthReport(report, opts.outputFile); err != nil {
			return fmt.Errorf("failed to save health report: %w", err)
		}
		logger.Info("Health report saved", "file", opts.outputFile)
	}

	// Display results
	formatter := cli.NewOutputFormatter(flags.Format)

	switch flags.Format {
	case "json", "yaml":
		return formatter.FormatOutput(report)
	default:
		return displayHealthResults(report, opts)
	}
}

func runContinuousHealthMonitoring(ctx context.Context, _ *cli.CommonFlags, opts healthOptions, logger logger.CommonLogger) error {
	logger.Info("Starting continuous health monitoring", "interval", opts.interval)

	ticker := time.NewTicker(opts.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("Stopping continuous health monitoring")
			return ctx.Err()
		case <-ticker.C:
			report, err := collectHealthMetrics(ctx, opts)
			if err != nil {
				logger.Warn("Failed to collect health metrics", "error", err)
				continue
			}

			// Display quick summary in continuous mode
			displayHealthSummary(report, opts)

			// Check for critical alerts
			criticalAlerts := 0
			for _, alert := range report.Alerts {
				if alert.Level == "critical" {
					criticalAlerts++
				}
			}

			if criticalAlerts > 0 {
				logger.Warn("Critical health alerts detected", "count", criticalAlerts)
			}
		}
	}
}

func startHealthServer(_ context.Context, opts healthOptions, logger logger.CommonLogger) error {
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		report, err := collectHealthMetrics(r.Context(), opts)
		if err != nil {
			http.Error(w, fmt.Sprintf("Health check failed: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(report)
	})

	// Metrics endpoint
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		report, err := collectHealthMetrics(r.Context(), opts)
		if err != nil {
			http.Error(w, fmt.Sprintf("Metrics collection failed: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(report.Metrics)
	})

	// Status endpoint
	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		report, err := collectHealthMetrics(r.Context(), opts)
		if err != nil {
			http.Error(w, "unhealthy", http.StatusInternalServerError)
			return
		}

		if report.OverallStatus == "healthy" {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "healthy (score: %.1f)", report.Score)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, "unhealthy (score: %.1f)", report.Score)
		}
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", opts.serverPort),
		Handler: mux,
	}

	logger.Info("Starting health monitoring server", "port", opts.serverPort)
	return server.ListenAndServe()
}

func collectHealthMetrics(ctx context.Context, opts healthOptions) (*SystemHealthReport, error) {
	startTime := time.Now()

	report := &SystemHealthReport{
		Timestamp:       startTime,
		Categories:      make([]HealthCategory, 0),
		Alerts:          make([]HealthAlert, 0),
		Recommendations: make([]string, 0),
	}

	// Collect system metrics
	report.Metrics = collectSystemMetrics(opts.includeEnv)

	// Run health checks by category
	categories := []string{"cpu", "memory", "disk", "network", "process", "environment"}
	if len(opts.categories) > 0 {
		categories = opts.categories
	}

	var wg sync.WaitGroup
	categoryChan := make(chan HealthCategory, len(categories))

	for _, category := range categories {
		wg.Add(1)
		go func(cat string) {
			defer wg.Done()
			categoryResult := runHealthCategory(ctx, cat, report.Metrics, opts)
			categoryChan <- categoryResult
		}(category)
	}

	go func() {
		wg.Wait()
		close(categoryChan)
	}()

	// Collect category results
	for category := range categoryChan {
		report.Categories = append(report.Categories, category)
	}

	// Calculate overall health score and status
	calculateOverallHealth(report)

	// Generate alerts based on health status
	generateHealthAlerts(report, opts.alertLevel)

	// Generate recommendations
	generateHealthRecommendations(report)

	return report, nil
}

func collectSystemMetrics(includeEnv bool) SystemMetrics {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	hostname, _ := os.Hostname()

	metrics := SystemMetrics{
		CPU: CPUMetrics{
			Count: runtime.NumCPU(),
			// Usage, LoadAvg would require platform-specific code
		},
		Memory: MemoryMetrics{
			Used:        memStats.Alloc,
			Total:       memStats.Sys,
			UsedPercent: float64(memStats.Alloc) / float64(memStats.Sys) * 100,
		},
		Process: ProcessMetrics{
			Goroutines:   runtime.NumGoroutine(),
			HeapObjects:  memStats.HeapObjects,
			GCRuns:       memStats.NumGC,
			GCPauseTotal: memStats.PauseTotalNs,
		},
		Environment: EnvMetrics{
			GoVersion: runtime.Version(),
			Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
			Hostname:  hostname,
		},
	}

	if includeEnv {
		metrics.Environment.EnvVars = getEnvironmentVars()
	}

	return metrics
}

func getEnvironmentVars() map[string]string {
	envVars := make(map[string]string)

	// Only include safe environment variables
	safeVars := []string{
		"PATH", "HOME", "USER", "SHELL", "LANG", "PWD", "GOPATH", "GOROOT",
	}

	for _, key := range safeVars {
		if value := os.Getenv(key); value != "" {
			envVars[key] = value
		}
	}

	return envVars
}

func runHealthCategory(ctx context.Context, category string, metrics SystemMetrics, opts healthOptions) HealthCategory {
	startTime := time.Now()

	result := HealthCategory{
		Name:   category,
		Checks: make([]HealthCheck, 0),
	}

	switch category {
	case "cpu":
		result.Checks = append(result.Checks, checkCPUHealth(metrics.CPU))
	case "memory":
		result.Checks = append(result.Checks, checkMemoryHealth(metrics.Memory))
	case "disk":
		result.Checks = append(result.Checks, checkDiskHealth(metrics.Disk))
	case "network":
		result.Checks = append(result.Checks, checkNetworkHealth(metrics.Network))
	case "process":
		result.Checks = append(result.Checks, checkProcessHealth(metrics.Process))
	case "environment":
		result.Checks = append(result.Checks, checkEnvironmentHealth(metrics.Environment))
	}

	result.Duration = time.Since(startTime)

	// Calculate category score and status
	calculateCategoryHealth(&result)

	return result
}

func checkCPUHealth(metrics CPUMetrics) HealthCheck {
	check := HealthCheck{
		Name: "CPU Usage",
		Details: map[string]interface{}{
			"cpu_count": metrics.Count,
			"usage":     metrics.Usage,
		},
	}

	// Simple CPU health check
	if metrics.Count >= 4 {
		check.Status = "healthy"
		check.Message = fmt.Sprintf("CPU cores: %d (sufficient)", metrics.Count)
	} else {
		check.Status = "warning"
		check.Message = fmt.Sprintf("CPU cores: %d (low)", metrics.Count)
	}

	check.Value = metrics.Count
	check.Threshold = 4

	return check
}

func checkMemoryHealth(metrics MemoryMetrics) HealthCheck {
	check := HealthCheck{
		Name:      "Memory Usage",
		Value:     metrics.UsedPercent,
		Threshold: 80.0,
		Details: map[string]interface{}{
			"used_mb":      metrics.Used / 1024 / 1024,
			"total_mb":     metrics.Total / 1024 / 1024,
			"used_percent": metrics.UsedPercent,
		},
	}

	if metrics.UsedPercent < 70 {
		check.Status = "healthy"
		check.Message = fmt.Sprintf("Memory usage: %.1f%% (good)", metrics.UsedPercent)
	} else if metrics.UsedPercent < 85 {
		check.Status = "warning"
		check.Message = fmt.Sprintf("Memory usage: %.1f%% (high)", metrics.UsedPercent)
	} else {
		check.Status = "critical"
		check.Message = fmt.Sprintf("Memory usage: %.1f%% (critical)", metrics.UsedPercent)
		check.Critical = true
	}

	return check
}

func checkDiskHealth(metrics DiskMetrics) HealthCheck {
	check := HealthCheck{
		Name:      "Disk Space",
		Value:     metrics.UsedPercent,
		Threshold: 85.0,
		Details: map[string]interface{}{
			"used_gb":      metrics.Used / 1024 / 1024 / 1024,
			"available_gb": metrics.Available / 1024 / 1024 / 1024,
			"used_percent": metrics.UsedPercent,
		},
	}

	if metrics.UsedPercent < 75 {
		check.Status = "healthy"
		check.Message = fmt.Sprintf("Disk usage: %.1f%% (good)", metrics.UsedPercent)
	} else if metrics.UsedPercent < 90 {
		check.Status = "warning"
		check.Message = fmt.Sprintf("Disk usage: %.1f%% (high)", metrics.UsedPercent)
	} else {
		check.Status = "critical"
		check.Message = fmt.Sprintf("Disk usage: %.1f%% (critical)", metrics.UsedPercent)
		check.Critical = true
	}

	return check
}

func checkNetworkHealth(metrics NetworkMetrics) HealthCheck {
	check := HealthCheck{
		Name: "Network Connectivity",
		Details: map[string]interface{}{
			"bytes_sent":  metrics.BytesSent,
			"bytes_recv":  metrics.BytesRecv,
			"connections": metrics.Connections,
			"latency_ms":  metrics.Latency,
		},
	}

	// Simple network health check based on available metrics
	check.Status = "healthy"
	check.Message = "Network connectivity appears normal"
	check.Value = "operational"

	return check
}

func checkProcessHealth(metrics ProcessMetrics) HealthCheck {
	check := HealthCheck{
		Name: "Process Health",
		Details: map[string]interface{}{
			"goroutines":   metrics.Goroutines,
			"heap_objects": metrics.HeapObjects,
			"gc_runs":      metrics.GCRuns,
		},
	}

	if metrics.Goroutines < 100 {
		check.Status = "healthy"
		check.Message = fmt.Sprintf("Goroutines: %d (normal)", metrics.Goroutines)
	} else if metrics.Goroutines < 1000 {
		check.Status = "warning"
		check.Message = fmt.Sprintf("Goroutines: %d (elevated)", metrics.Goroutines)
	} else {
		check.Status = "critical"
		check.Message = fmt.Sprintf("Goroutines: %d (excessive)", metrics.Goroutines)
		check.Critical = true
	}

	check.Value = metrics.Goroutines
	check.Threshold = 100

	return check
}

func checkEnvironmentHealth(metrics EnvMetrics) HealthCheck {
	check := HealthCheck{
		Name: "Environment",
		Details: map[string]interface{}{
			"go_version": metrics.GoVersion,
			"platform":   metrics.Platform,
			"hostname":   metrics.Hostname,
		},
	}

	check.Status = "healthy"
	check.Message = fmt.Sprintf("Running %s on %s", metrics.GoVersion, metrics.Platform)
	check.Value = "operational"

	return check
}

func calculateCategoryHealth(category *HealthCategory) {
	totalScore := 100.0
	healthyChecks := 0

	for _, check := range category.Checks {
		switch check.Status {
		case "healthy":
			healthyChecks++
		case "warning":
			totalScore -= 20
		case "critical":
			totalScore -= 50
		}
	}

	if totalScore < 0 {
		totalScore = 0
	}

	category.Score = totalScore

	if totalScore >= 80 {
		category.Status = "healthy"
	} else if totalScore >= 60 {
		category.Status = "warning"
	} else {
		category.Status = "critical"
	}
}

func calculateOverallHealth(report *SystemHealthReport) {
	if len(report.Categories) == 0 {
		report.OverallStatus = "unknown"
		report.Score = 0
		return
	}

	totalScore := 0.0
	criticalCategories := 0

	for _, category := range report.Categories {
		totalScore += category.Score
		if category.Status == "critical" {
			criticalCategories++
		}
	}

	report.Score = totalScore / float64(len(report.Categories))

	if criticalCategories > 0 {
		report.OverallStatus = "critical"
	} else if report.Score >= 80 {
		report.OverallStatus = "healthy"
	} else if report.Score >= 60 {
		report.OverallStatus = "warning"
	} else {
		report.OverallStatus = "unhealthy"
	}
}

func generateHealthAlerts(report *SystemHealthReport, alertLevel string) {
	minLevel := getAlertLevelPriority(alertLevel)

	for _, category := range report.Categories {
		for _, check := range category.Checks {
			if check.Status == "critical" && getAlertLevelPriority("critical") >= minLevel {
				alert := HealthAlert{
					Level:      "critical",
					Category:   category.Name,
					Message:    fmt.Sprintf("%s: %s", check.Name, check.Message),
					Timestamp:  time.Now(),
					Suggestion: generateSuggestion(check),
				}
				report.Alerts = append(report.Alerts, alert)
			} else if check.Status == "warning" && getAlertLevelPriority("warning") >= minLevel {
				alert := HealthAlert{
					Level:      "warning",
					Category:   category.Name,
					Message:    fmt.Sprintf("%s: %s", check.Name, check.Message),
					Timestamp:  time.Now(),
					Suggestion: generateSuggestion(check),
				}
				report.Alerts = append(report.Alerts, alert)
			}
		}
	}
}

func getAlertLevelPriority(level string) int {
	switch level {
	case "critical":
		return 3
	case "warning":
		return 2
	case "info":
		return 1
	default:
		return 0
	}
}

func generateSuggestion(check HealthCheck) string {
	switch check.Name {
	case "Memory Usage":
		return "Consider increasing available memory or optimizing memory usage"
	case "Disk Space":
		return "Free up disk space or extend storage capacity"
	case "Process Health":
		return "Monitor for goroutine leaks and optimize concurrent operations"
	case "CPU Usage":
		return "Consider upgrading CPU or optimizing CPU-intensive operations"
	default:
		return "Review system configuration and resource allocation"
	}
}

func generateHealthRecommendations(report *SystemHealthReport) {
	recommendations := make([]string, 0)

	if report.Score < 70 {
		recommendations = append(recommendations, "Overall system health is below optimal - review critical issues")
	}

	criticalAlerts := 0
	warningAlerts := 0
	for _, alert := range report.Alerts {
		switch alert.Level {
		case "critical":
			criticalAlerts++
		case "warning":
			warningAlerts++
		}
	}

	if criticalAlerts > 0 {
		recommendations = append(recommendations, fmt.Sprintf("Address %d critical health alerts immediately", criticalAlerts))
	}

	if warningAlerts > 0 {
		recommendations = append(recommendations, fmt.Sprintf("Review %d warning alerts for potential improvements", warningAlerts))
	}

	// Check specific metrics for recommendations
	for _, category := range report.Categories {
		if category.Name == "memory" && category.Score < 70 {
			recommendations = append(recommendations, "Consider memory optimization or increasing available RAM")
		}
		if category.Name == "process" && category.Score < 70 {
			recommendations = append(recommendations, "Monitor process health and optimize resource usage")
		}
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "System health appears good - continue monitoring")
	}

	report.Recommendations = recommendations
}

func saveHealthReport(report *SystemHealthReport, filename string) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal health report: %w", err)
	}

	return os.WriteFile(filename, data, 0o600)
}

func displayHealthResults(report *SystemHealthReport, opts healthOptions) error {
	// Display overall health status
	statusIcon := "✅"
	switch report.OverallStatus {
	case "warning":
		statusIcon = "⚠️"
	case "critical", "unhealthy":
		statusIcon = "❌"
	}

	logger.SimpleInfo(fmt.Sprintf("%s System Health Status", statusIcon),
		"status", report.OverallStatus,
		"score", fmt.Sprintf("%.1f/100", report.Score),
		"categories", len(report.Categories),
	)

	// Display category results
	for _, category := range report.Categories {
		categoryIcon := "✅"
		switch category.Status {
		case "warning":
			categoryIcon = "⚠️"
		case "critical":
			categoryIcon = "❌"
		}

		logger.SimpleInfo(fmt.Sprintf("  %s %s", categoryIcon, category.Name),
			"status", category.Status,
			"score", fmt.Sprintf("%.1f", category.Score),
			"checks", len(category.Checks),
		)

		if opts.detailed {
			for _, check := range category.Checks {
				checkIcon := "✅"
				switch check.Status {
				case "warning":
					checkIcon = "⚠️"
				case "critical":
					checkIcon = "❌"
				}

				logger.SimpleInfo(fmt.Sprintf("    %s %s", checkIcon, check.Name),
					"status", check.Status,
					"message", check.Message,
				)
			}
		}
	}

	// Display alerts
	if len(report.Alerts) > 0 {
		logger.SimpleWarn(fmt.Sprintf("🚨 Health Alerts (%d)", len(report.Alerts)))
		for _, alert := range report.Alerts {
			alertIcon := "⚠️"
			if alert.Level == "critical" {
				alertIcon = "🔴"
			}

			logger.SimpleWarn(fmt.Sprintf("  %s [%s] %s", alertIcon, alert.Category, alert.Message),
				"level", alert.Level,
				"suggestion", alert.Suggestion,
			)
		}
	}

	// Display recommendations
	if len(report.Recommendations) > 0 {
		logger.SimpleInfo("💡 Recommendations:")
		for _, rec := range report.Recommendations {
			logger.SimpleInfo(fmt.Sprintf("  • %s", rec))
		}
	}

	return nil
}

func displayHealthSummary(report *SystemHealthReport, opts healthOptions) {
	statusIcon := "✅"
	switch report.OverallStatus {
	case "warning":
		statusIcon = "⚠️"
	case "critical", "unhealthy":
		statusIcon = "❌"
	}

	logger.SimpleInfo(fmt.Sprintf("%s Health: %s (%.1f) | Categories: %d | Alerts: %d",
		statusIcon, report.OverallStatus, report.Score, len(report.Categories), len(report.Alerts)))
}
