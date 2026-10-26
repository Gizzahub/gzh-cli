// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package doctor

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/gizzahub/gzh-cli/internal/cli"
	"github.com/gizzahub/gzh-cli/internal/logger"
	"github.com/gizzahub/gzh-cli/internal/profiling"
)

// BenchmarkReport represents a comprehensive benchmark analysis report.
type BenchmarkReport struct {
	Timestamp       time.Time                   `json:"timestamp"`
	Environment     BenchmarkEnvironment        `json:"environment"`
	Summary         BenchmarkSummary            `json:"summary"`
	Benchmarks      []profiling.BenchmarkResult `json:"benchmarks"`
	Regressions     []PerformanceRegression     `json:"regressions"`
	Improvements    []PerformanceImprovement    `json:"improvements"`
	Recommendations []string                    `json:"recommendations"`
	CIMetrics       CIBenchmarkMetrics          `json:"ci_metrics"`
}

// BenchmarkEnvironment captures the testing environment details.
type BenchmarkEnvironment struct {
	Platform      string `json:"platform"`
	GoVersion     string `json:"go_version"`
	NumCPU        int    `json:"num_cpu"`
	NumGoroutines int    `json:"num_goroutines"`
	MemoryLimit   uint64 `json:"memory_limit"`
	GitCommit     string `json:"git_commit,omitempty"`
	GitBranch     string `json:"git_branch,omitempty"`
	BuildInfo     string `json:"build_info,omitempty"`
}

// BenchmarkSummary provides high-level benchmark metrics.
type BenchmarkSummary struct {
	TotalBenchmarks  int           `json:"total_benchmarks"`
	PassedBenchmarks int           `json:"passed_benchmarks"`
	FailedBenchmarks int           `json:"failed_benchmarks"`
	TotalDuration    time.Duration `json:"total_duration"`
	AverageOpsPerSec float64       `json:"average_ops_per_sec"`
	TotalMemoryUsage uint64        `json:"total_memory_usage"`
	PerformanceScore float64       `json:"performance_score"`
}

// PerformanceRegression represents a performance regression detected.
type PerformanceRegression struct {
	BenchmarkName     string  `json:"benchmark_name"`
	CurrentOpsPerSec  float64 `json:"current_ops_per_sec"`
	BaselineOpsPerSec float64 `json:"baseline_ops_per_sec"`
	RegressionPercent float64 `json:"regression_percent"`
	Severity          string  `json:"severity"`
	Impact            string  `json:"impact"`
}

// PerformanceImprovement represents a performance improvement detected.
type PerformanceImprovement struct {
	BenchmarkName      string  `json:"benchmark_name"`
	CurrentOpsPerSec   float64 `json:"current_ops_per_sec"`
	BaselineOpsPerSec  float64 `json:"baseline_ops_per_sec"`
	ImprovementPercent float64 `json:"improvement_percent"`
	Impact             string  `json:"impact"`
}

// CIBenchmarkMetrics contains CI-specific benchmark metrics.
type CIBenchmarkMetrics struct {
	ExitCode              int      `json:"exit_code"`
	RegressionThreshold   float64  `json:"regression_threshold"`
	HasCriticalRegression bool     `json:"has_critical_regression"`
	RecommendedAction     string   `json:"recommended_action"`
	ArtifactPaths         []string `json:"artifact_paths"`
}

// newBenchmarkCmd creates the benchmark subcommand for performance testing.
func newBenchmarkCmd() *cobra.Command {
	var (
		packagePattern      string
		benchmarkFilter     string
		iterations          int
		duration            time.Duration
		cpuProfile          bool
		memProfile          bool
		outputFile          string
		baselineFile        string
		ciMode              bool
		regressionThreshold float64
		generateArtifacts   bool
		compareMode         bool
		createSnapshot      bool
		snapshotDir         string
		analyzeSnapshots    bool
		snapshotID          string
		trendAnalysis       bool
		trendWindowDays     int
	)

	cmd := cli.NewCommandBuilder("benchmark", "Package performance benchmarking (currently unsupported)").
		WithLongDescription(`Benchmark execution is currently unsupported.

Package benchmark discovery, profiling artifact generation, and current-measurement
comparison are not implemented by this CLI. Execution returns an explicit error
before starting a profiler, querying Git, or writing results and snapshots.

Existing flags remain available for compatibility; they do not enable execution.
Use --help to inspect the retained options.`).
		WithExample("gz doctor benchmark --help").
		WithFormatFlag("table", []string{"table", "json", "yaml"}).
		WithRunFuncE(func(ctx context.Context, flags *cli.CommonFlags, args []string) error {
			return runBenchmarkAnalysis(ctx, flags, benchmarkOptions{
				packagePattern:      packagePattern,
				benchmarkFilter:     benchmarkFilter,
				iterations:          iterations,
				duration:            duration,
				cpuProfile:          cpuProfile,
				memProfile:          memProfile,
				outputFile:          outputFile,
				baselineFile:        baselineFile,
				ciMode:              ciMode,
				regressionThreshold: regressionThreshold,
				generateArtifacts:   generateArtifacts,
				compareMode:         compareMode,
				createSnapshot:      createSnapshot,
				snapshotDir:         snapshotDir,
				analyzeSnapshots:    analyzeSnapshots,
				snapshotID:          snapshotID,
				trendAnalysis:       trendAnalysis,
				trendWindowDays:     trendWindowDays,
			})
		}).
		Build()

	cmd.Flags().StringVar(&packagePattern, "package", "./...", "Package pattern to benchmark")
	cmd.Flags().StringVar(&benchmarkFilter, "filter", "", "Benchmark name filter (regex)")
	cmd.Flags().IntVar(&iterations, "iterations", 10, "Number of benchmark iterations")
	cmd.Flags().DurationVar(&duration, "duration", 30*time.Second, "Maximum duration per benchmark")
	cmd.Flags().BoolVar(&cpuProfile, "cpu-profile", false, "Generate CPU profiling data")
	cmd.Flags().BoolVar(&memProfile, "mem-profile", false, "Generate memory profiling data")
	cmd.Flags().StringVar(&outputFile, "output", "", "Output file for benchmark results")
	cmd.Flags().StringVar(&baselineFile, "baseline", "", "Baseline file for comparison")
	cmd.Flags().BoolVar(&ciMode, "ci", false, "CI mode with exit codes and artifacts")
	cmd.Flags().Float64Var(&regressionThreshold, "regression-threshold", 10.0, "Performance regression threshold (%)")
	cmd.Flags().BoolVar(&generateArtifacts, "artifacts", false, "Generate performance artifacts")
	cmd.Flags().BoolVar(&compareMode, "compare", false, "Compare against baseline only")

	// Performance snapshot flags
	cmd.Flags().BoolVar(&createSnapshot, "create-snapshot", false, "Create performance snapshot after benchmarks")
	cmd.Flags().StringVar(&snapshotDir, "snapshot-dir", "performance-snapshots", "Directory to store performance snapshots")
	cmd.Flags().BoolVar(&analyzeSnapshots, "analyze-snapshots", false, "Analyze performance snapshots")
	cmd.Flags().StringVar(&snapshotID, "snapshot-id", "", "Baseline snapshot ID for comparison")
	cmd.Flags().BoolVar(&trendAnalysis, "trend-analysis", false, "Enable historical trend analysis")
	cmd.Flags().IntVar(&trendWindowDays, "trend-window-days", 30, "Historical trend analysis window in days")

	return cmd
}

type benchmarkOptions struct {
	packagePattern      string
	benchmarkFilter     string
	iterations          int
	duration            time.Duration
	cpuProfile          bool
	memProfile          bool
	outputFile          string
	baselineFile        string
	ciMode              bool
	regressionThreshold float64
	generateArtifacts   bool
	compareMode         bool
	createSnapshot      bool
	snapshotDir         string
	analyzeSnapshots    bool
	snapshotID          string
	trendAnalysis       bool
	trendWindowDays     int
}

// CLI에는 실제 package 발견·실행과 current 측정 입력 계약이 아직 없다.
func runBenchmarkAnalysis(_ context.Context, _ *cli.CommonFlags, _ benchmarkOptions) error {
	return unsupportedBenchmarkExecution()
}

func unsupportedBenchmarkExecution() error {
	return fmt.Errorf("doctor benchmark execution is not supported: package discovery and current measurement inputs are not implemented")
}

func getBenchmarkEnvironment(ctx context.Context) BenchmarkEnvironment {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	env := BenchmarkEnvironment{
		Platform:      fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		GoVersion:     runtime.Version(),
		NumCPU:        runtime.NumCPU(),
		NumGoroutines: runtime.NumGoroutine(),
		MemoryLimit:   memStats.Sys,
	}

	// Try to get Git information
	if commit, err := getGitCommit(ctx); err == nil {
		env.GitCommit = commit
	}
	if branch, err := getGitBranch(ctx); err == nil {
		env.GitBranch = branch
	}

	return env
}

// 직접 호출도 가상의 sleep을 요청한 package 측정으로 반환하지 않는다.
func runPackageBenchmarks(_ context.Context, _ *profiling.BenchmarkSuite, _ benchmarkOptions, _ *BenchmarkReport) error {
	return unsupportedBenchmarkExecution()
}

func generateBenchmarkArtifacts(_ *profiling.Profiler, report *BenchmarkReport, opts benchmarkOptions) error {
	if opts.cpuProfile || opts.memProfile {
		return fmt.Errorf("benchmark CPU/memory profile artifact generation is not supported")
	}
	artifactDir := "benchmark-artifacts"
	if err := os.MkdirAll(artifactDir, 0o755); err != nil {
		return fmt.Errorf("failed to create artifact directory: %w", err)
	}

	timestamp := report.Timestamp.Format("20060102-150405")

	// Generate benchmark report
	reportFile := filepath.Join(artifactDir, fmt.Sprintf("benchmark-report-%s.json", timestamp))
	if err := saveBenchmarkResults(report, reportFile); err != nil {
		return fmt.Errorf("failed to save benchmark report: %w", err)
	}
	report.CIMetrics.ArtifactPaths = append(report.CIMetrics.ArtifactPaths, reportFile)

	return nil
}

func compareAgainstBaseline(report *BenchmarkReport, baselineFile string, threshold float64) error {
	// Load baseline data
	baselineData, err := os.ReadFile(baselineFile)
	if err != nil {
		return fmt.Errorf("failed to read baseline file: %w", err)
	}

	var baseline BenchmarkReport
	if err := json.Unmarshal(baselineData, &baseline); err != nil {
		return fmt.Errorf("failed to parse baseline: %w", err)
	}

	// Create baseline lookup map
	baselineMap := make(map[string]profiling.BenchmarkResult)
	for _, result := range baseline.Benchmarks {
		baselineMap[result.Name] = result
	}

	// Compare benchmarks
	for _, current := range report.Benchmarks {
		if baselineResult, exists := baselineMap[current.Name]; exists {
			compareResults(current, baselineResult, threshold, report)
		}
	}

	return nil
}

func compareResults(current, baseline profiling.BenchmarkResult, threshold float64, report *BenchmarkReport) {
	if !hasMeasurableThroughput(current) || !hasMeasurableThroughput(baseline) {
		return
	}

	changePercent := ((current.OpsPerSec - baseline.OpsPerSec) / baseline.OpsPerSec) * 100

	if changePercent < -threshold {
		// Performance regression
		severity := "medium"
		if changePercent < -threshold*2 {
			severity = "high"
		}
		if changePercent < -threshold*3 {
			severity = "critical"
		}

		regression := PerformanceRegression{
			BenchmarkName:     current.Name,
			CurrentOpsPerSec:  current.OpsPerSec,
			BaselineOpsPerSec: baseline.OpsPerSec,
			RegressionPercent: -changePercent,
			Severity:          severity,
			Impact:            generateImpactDescription(changePercent),
		}

		report.Regressions = append(report.Regressions, regression)

		if severity == "critical" {
			report.CIMetrics.HasCriticalRegression = true
		}
	} else if changePercent > threshold {
		// Performance improvement
		improvement := PerformanceImprovement{
			BenchmarkName:      current.Name,
			CurrentOpsPerSec:   current.OpsPerSec,
			BaselineOpsPerSec:  baseline.OpsPerSec,
			ImprovementPercent: changePercent,
			Impact:             generateImpactDescription(changePercent),
		}

		report.Improvements = append(report.Improvements, improvement)
	}
}

func generateImpactDescription(changePercent float64) string {
	absChange := changePercent
	if absChange < 0 {
		absChange = -absChange
	}

	switch {
	case absChange < 10:
		return "minimal"
	case absChange < 25:
		return "moderate"
	case absChange < 50:
		return "significant"
	default:
		return "major"
	}
}

func generateBenchmarkSummary(report *BenchmarkReport) {
	report.Summary.TotalBenchmarks = len(report.Benchmarks)

	if report.Summary.TotalBenchmarks == 0 {
		return
	}

	var totalOpsPerSec float64
	measurableBenchmarks := 0
	var totalMemory uint64
	var totalDuration time.Duration

	for _, result := range report.Benchmarks {
		if hasMeasurableThroughput(result) {
			totalOpsPerSec += result.OpsPerSec
			measurableBenchmarks++
		}
		totalMemory += result.MemoryAfter - result.MemoryBefore
		totalDuration += result.Duration
	}

	if measurableBenchmarks > 0 {
		report.Summary.AverageOpsPerSec = totalOpsPerSec / float64(measurableBenchmarks)
	}
	report.Summary.TotalMemoryUsage = totalMemory
	report.Summary.TotalDuration = totalDuration

	// Calculate performance score (0-100)
	baseScore := 100.0
	if len(report.Regressions) > 0 {
		regressionPenalty := float64(len(report.Regressions)) * 10.0
		baseScore -= regressionPenalty
	}
	if report.CIMetrics.HasCriticalRegression {
		baseScore -= 30.0
	}
	if baseScore < 0 {
		baseScore = 0
	}

	report.Summary.PerformanceScore = baseScore
}

func hasMeasurableThroughput(result profiling.BenchmarkResult) bool {
	return result.OpsPerSec > 0
}

func generateBenchmarkRecommendations(report *BenchmarkReport) {
	recommendations := make([]string, 0)

	// Performance regression recommendations
	if len(report.Regressions) > 0 {
		criticalCount := 0
		for _, reg := range report.Regressions {
			if reg.Severity == "critical" {
				criticalCount++
			}
		}

		if criticalCount > 0 {
			recommendations = append(recommendations,
				fmt.Sprintf("Address %d critical performance regressions immediately", criticalCount))
		}

		recommendations = append(recommendations,
			fmt.Sprintf("Investigate %d performance regressions", len(report.Regressions)))
	}

	// Memory usage recommendations
	if report.Summary.TotalMemoryUsage > 100*1024*1024 { // 100MB
		recommendations = append(recommendations,
			"Consider memory optimization - high memory usage detected")
	}

	// Benchmark coverage recommendations
	if report.Summary.TotalBenchmarks < 5 {
		recommendations = append(recommendations,
			"Add more benchmark coverage for critical code paths")
	}

	// Performance score recommendations
	if report.Summary.PerformanceScore < 70 {
		recommendations = append(recommendations,
			"Performance score is below threshold - review recent changes")
	}

	report.Recommendations = recommendations
}

func handleCIMode(report *BenchmarkReport) {
	// Set exit code based on results
	switch {
	case report.CIMetrics.HasCriticalRegression:
		report.CIMetrics.ExitCode = 1
		report.CIMetrics.RecommendedAction = "Build should fail - critical performance regression detected"
	case len(report.Regressions) > 0:
		report.CIMetrics.ExitCode = 0 // Warning but don't fail
		report.CIMetrics.RecommendedAction = "Build passes with warnings - monitor performance regressions"
	default:
		report.CIMetrics.ExitCode = 0
		report.CIMetrics.RecommendedAction = "Build passes - no performance issues detected"
	}
}

func saveBenchmarkResults(report *BenchmarkReport, filename string) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal results: %w", err)
	}

	return os.WriteFile(filename, data, 0o600)
}

func displayBenchmarkResults(report *BenchmarkReport, opts benchmarkOptions, snapshotAnalysis *SnapshotAnalysis) error {
	// Display environment information
	logger.SimpleInfo(
		"🔧 Benchmark Environment",
		"platform", report.Environment.Platform,
		"go_version", report.Environment.GoVersion,
		"cpu_cores", report.Environment.NumCPU,
		"git_commit", report.Environment.GitCommit,
	)

	// Display summary
	logger.SimpleInfo(
		"📊 Benchmark Summary",
		"total_benchmarks", report.Summary.TotalBenchmarks,
		"passed", report.Summary.PassedBenchmarks,
		"failed", report.Summary.FailedBenchmarks,
		"avg_ops_per_sec", fmt.Sprintf("%.2f", report.Summary.AverageOpsPerSec),
		"performance_score", fmt.Sprintf("%.1f", report.Summary.PerformanceScore),
	)

	// Display benchmark results
	if len(report.Benchmarks) > 0 {
		logger.SimpleInfo("🏃 Benchmark Results:")

		// Sort by operations per second for better readability
		sortedBenchmarks := make([]profiling.BenchmarkResult, len(report.Benchmarks))
		copy(sortedBenchmarks, report.Benchmarks)
		sort.Slice(sortedBenchmarks, func(i, j int) bool {
			return sortedBenchmarks[i].OpsPerSec > sortedBenchmarks[j].OpsPerSec
		})

		for _, result := range sortedBenchmarks {
			logger.SimpleInfo(
				fmt.Sprintf("  %s", result.Name),
				"ops_per_sec", fmt.Sprintf("%.2f", result.OpsPerSec),
				"ns_per_op", result.NsPerOp,
				"allocs_per_op", result.AllocsPerOp,
			)
		}
	}

	// Display regressions
	if len(report.Regressions) > 0 {
		logger.SimpleWarn("⚠️ Performance Regressions:")
		for _, reg := range report.Regressions {
			severityIcon := "🟡"
			switch reg.Severity {
			case "high":
				severityIcon = "🔴"
			case "critical":
				severityIcon = "💥"
			}

			logger.SimpleWarn(
				fmt.Sprintf("  %s %s", severityIcon, reg.BenchmarkName),
				"regression", fmt.Sprintf("%.1f%%", reg.RegressionPercent),
				"severity", reg.Severity,
				"impact", reg.Impact,
			)
		}
	}

	// Display improvements
	if len(report.Improvements) > 0 {
		logger.SimpleInfo("✅ Performance Improvements:")
		for _, imp := range report.Improvements {
			logger.SimpleInfo(
				fmt.Sprintf("  🚀 %s", imp.BenchmarkName),
				"improvement", fmt.Sprintf("%.1f%%", imp.ImprovementPercent),
				"impact", imp.Impact,
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

	// Display snapshot analysis if available
	if snapshotAnalysis != nil {
		displaySnapshotAnalysis(snapshotAnalysis)
	}

	// Display CI metrics if in CI mode
	if opts.ciMode {
		logger.SimpleInfo(
			"🔄 CI Metrics",
			"exit_code", report.CIMetrics.ExitCode,
			"action", report.CIMetrics.RecommendedAction,
			"artifacts", len(report.CIMetrics.ArtifactPaths),
		)

		if len(report.CIMetrics.ArtifactPaths) > 0 {
			logger.SimpleInfo("📁 Generated Artifacts:")
			for _, path := range report.CIMetrics.ArtifactPaths {
				logger.SimpleInfo(fmt.Sprintf("  📄 %s", path))
			}
		}
	}

	// Set exit code for CI mode
	if opts.ciMode && report.CIMetrics.ExitCode != 0 {
		os.Exit(report.CIMetrics.ExitCode)
	}

	return nil
}

// Helper functions for Git integration.
func getGitCommit(ctx context.Context) (string, error) {
	return runGit(ctx, "rev-parse", "HEAD")
}

func getGitBranch(ctx context.Context) (string, error) {
	return runGit(ctx, "branch", "--show-current")
}

func runGit(ctx context.Context, args ...string) (string, error) {
	output, err := exec.CommandContext(ctx, "git", args...).Output()
	if err != nil {
		return "", fmt.Errorf("git command failed: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// BenchmarkFunction represents a discovered benchmark function.
type BenchmarkFunction struct {
	Name     string
	Package  string
	Function func(ctx context.Context)
}

// performSnapshotAnalysis performs comprehensive snapshot analysis.
func performSnapshotAnalysis(snapshotManager *SnapshotManager, report *BenchmarkReport, opts benchmarkOptions, logger logger.CommonLogger) (*SnapshotAnalysis, error) {
	// Create current snapshot for analysis
	currentSnapshot := &PerformanceSnapshot{
		ID:          "current-analysis",
		Timestamp:   report.Timestamp,
		Environment: report.Environment,
		Benchmarks:  report.Benchmarks,
		Metadata: map[string]any{
			"analysis_mode": true,
			"report_id":     fmt.Sprintf("analysis-%d", time.Now().Unix()),
		},
	}

	// Load baseline snapshot if specified
	var baselineSnapshot *PerformanceSnapshot
	var err error

	if opts.snapshotID != "" {
		logger.Info("Loading specified baseline snapshot", "id", opts.snapshotID)
		baselineSnapshot, err = snapshotManager.LoadSnapshot(opts.snapshotID)
		if err != nil {
			return nil, fmt.Errorf("failed to load baseline snapshot %s: %w", opts.snapshotID, err)
		}
	} else {
		// Find the most recent snapshot as baseline
		logger.Info("Finding most recent snapshot for baseline comparison")
		snapshots, err := snapshotManager.ListSnapshots()
		if err != nil {
			return nil, fmt.Errorf("failed to list snapshots: %w", err)
		}

		if len(snapshots) == 0 {
			return nil, fmt.Errorf("no baseline snapshots available for comparison")
		}

		// Use the most recent snapshot as baseline
		baselineSnapshot = snapshots[0]
		logger.Info("Using most recent snapshot as baseline", "id", baselineSnapshot.ID, "timestamp", baselineSnapshot.Timestamp)
	}

	// Configure analysis options
	analysisOptions := AnalysisOptions{
		RegressionThreshold: opts.regressionThreshold,
		IncludeTrends:       opts.trendAnalysis,
		TrendWindowDays:     opts.trendWindowDays,
		ConfidenceLevel:     0.95,
		GeneratePredictions: true,
	}

	// Perform comprehensive analysis
	logger.Info("Performing snapshot analysis", "baseline", baselineSnapshot.ID, "trend_analysis", opts.trendAnalysis)
	analysis, err := snapshotManager.AnalyzeSnapshots(currentSnapshot, baselineSnapshot, analysisOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze snapshots: %w", err)
	}

	logger.Info("Snapshot analysis completed",
		"regressions", len(analysis.Regressions),
		"improvements", len(analysis.Improvements),
		"trends", len(analysis.Trends),
		"health_score", analysis.Summary.OverallHealthScore)

	return analysis, nil
}

// displaySnapshotAnalysis displays comprehensive snapshot analysis results.
func displaySnapshotAnalysis(analysis *SnapshotAnalysis) {
	logger.SimpleInfo("📊 Performance Snapshot Analysis")

	// Display comparison summary
	logger.SimpleInfo(
		"🔄 Comparison Summary",
		"baseline_time", analysis.Baseline.Timestamp.Format("2006-01-02 15:04:05"),
		"time_difference", analysis.Comparison.TimeDifference.String(),
		"overall_change", fmt.Sprintf("%.2f%%", analysis.Comparison.OverallChange),
		"performance_score", fmt.Sprintf("%.1f", analysis.Comparison.PerformanceScore),
	)

	// Display benchmark counts
	logger.SimpleInfo(
		"📈 Benchmark Counts",
		"current", analysis.Comparison.BenchmarkCount,
		"baseline", analysis.Comparison.BaselineBenchmarks,
		"new_benchmarks", len(analysis.Comparison.NewBenchmarks),
		"removed_benchmarks", len(analysis.Comparison.RemovedBenchmarks),
	)

	// Display new benchmarks if any
	if len(analysis.Comparison.NewBenchmarks) > 0 {
		logger.SimpleInfo("🆕 New Benchmarks:")
		for _, name := range analysis.Comparison.NewBenchmarks {
			logger.SimpleInfo(fmt.Sprintf("  + %s", name))
		}
	}

	// Display removed benchmarks if any
	if len(analysis.Comparison.RemovedBenchmarks) > 0 {
		logger.SimpleWarn("🗑️ Removed Benchmarks:")
		for _, name := range analysis.Comparison.RemovedBenchmarks {
			logger.SimpleWarn(fmt.Sprintf("  - %s", name))
		}
	}

	// Display performance regressions from snapshot analysis
	if len(analysis.Regressions) > 0 {
		logger.SimpleWarn("⚠️ Performance Regressions (Snapshot Analysis):")
		for _, reg := range analysis.Regressions {
			severityIcon := "🟡"
			switch reg.Severity {
			case "high":
				severityIcon = "🔴"
			case "critical":
				severityIcon = "💥"
			}

			logger.SimpleWarn(
				fmt.Sprintf("  %s %s", severityIcon, reg.BenchmarkName),
				"regression", fmt.Sprintf("%.1f%%", reg.RegressionPercent),
				"severity", reg.Severity,
				"impact", reg.Impact,
				"current_ops", fmt.Sprintf("%.2f", reg.CurrentOpsPerSec),
				"baseline_ops", fmt.Sprintf("%.2f", reg.BaselineOpsPerSec),
			)
		}
	}

	// Display performance improvements from snapshot analysis
	if len(analysis.Improvements) > 0 {
		logger.SimpleInfo("✅ Performance Improvements (Snapshot Analysis):")
		for _, imp := range analysis.Improvements {
			logger.SimpleInfo(
				fmt.Sprintf("  🚀 %s", imp.BenchmarkName),
				"improvement", fmt.Sprintf("%.1f%%", imp.ImprovementPercent),
				"impact", imp.Impact,
				"current_ops", fmt.Sprintf("%.2f", imp.CurrentOpsPerSec),
				"baseline_ops", fmt.Sprintf("%.2f", imp.BaselineOpsPerSec),
			)
		}
	}

	// Display performance trends if available
	if len(analysis.Trends) > 0 {
		logger.SimpleInfo("📈 Performance Trends:")
		for _, trend := range analysis.Trends {
			var trendIcon string
			switch trend.TrendDirection {
			case "improving":
				trendIcon = "📈"
			case "degrading":
				trendIcon = "📉"
			default:
				trendIcon = "📊"
			}

			logger.SimpleInfo(
				fmt.Sprintf("  %s %s", trendIcon, trend.BenchmarkName),
				"direction", trend.TrendDirection,
				"data_points", len(trend.DataPoints),
				"confidence", fmt.Sprintf("%.2f", trend.ConfidenceLevel),
				"r_squared", fmt.Sprintf("%.3f", trend.LinearRegression.RSquared),
			)

			// Display predictions if available
			if trend.Prediction.RecommendedAction != "" {
				logger.SimpleInfo(fmt.Sprintf("    Prediction: %s", trend.Prediction.RecommendedAction))
				if trend.Prediction.NextWeekChange != 0 {
					logger.SimpleInfo(fmt.Sprintf("    Next week: %.2f%% change", trend.Prediction.NextWeekChange))
				}
				if trend.Prediction.NextMonthChange != 0 {
					logger.SimpleInfo(fmt.Sprintf("    Next month: %.2f%% change", trend.Prediction.NextMonthChange))
				}
			}
		}
	}

	// Display overall health summary
	logger.SimpleInfo(
		"🏥 Overall Health Summary",
		"health_score", fmt.Sprintf("%.1f/100", analysis.Summary.OverallHealthScore),
		"trending_up", analysis.Summary.TrendingUp,
		"trending_down", analysis.Summary.TrendingDown,
		"stable_trends", analysis.Summary.StableTrends,
		"critical_issues", analysis.Summary.CriticalIssues,
	)

	// Display recommended actions
	if len(analysis.Summary.RecommendedActions) > 0 {
		logger.SimpleInfo("💡 Snapshot Analysis Recommendations:")
		for _, action := range analysis.Summary.RecommendedActions {
			logger.SimpleInfo(fmt.Sprintf("  • %s", action))
		}
	}
}
