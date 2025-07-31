// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package doctor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/gizzahub/gzh-manager-go/internal/logger"
	"github.com/gizzahub/gzh-manager-go/internal/profiling"
)

// PerformanceSnapshot represents a point-in-time performance measurement
type PerformanceSnapshot struct {
	ID          string                      `json:"id"`
	Timestamp   time.Time                   `json:"timestamp"`
	Version     string                      `json:"version,omitempty"`
	GitCommit   string                      `json:"git_commit,omitempty"`
	GitBranch   string                      `json:"git_branch,omitempty"`
	Environment BenchmarkEnvironment        `json:"environment"`
	Benchmarks  []profiling.BenchmarkResult `json:"benchmarks"`
	Metadata    map[string]interface{}      `json:"metadata,omitempty"`
}

// SnapshotAnalysis represents the analysis results comparing snapshots
type SnapshotAnalysis struct {
	Current      *PerformanceSnapshot     `json:"current"`
	Baseline     *PerformanceSnapshot     `json:"baseline"`
	Comparison   SnapshotComparison       `json:"comparison"`
	Trends       []PerformanceTrend       `json:"trends"`
	Regressions  []PerformanceRegression  `json:"regressions"`
	Improvements []PerformanceImprovement `json:"improvements"`
	Summary      SnapshotSummary          `json:"summary"`
}

// SnapshotComparison provides comparison metrics between two snapshots
type SnapshotComparison struct {
	TimeDifference     time.Duration `json:"time_difference"`
	EnvironmentChanged bool          `json:"environment_changed"`
	BenchmarkCount     int           `json:"benchmark_count"`
	BaselineBenchmarks int           `json:"baseline_benchmarks"`
	NewBenchmarks      []string      `json:"new_benchmarks"`
	RemovedBenchmarks  []string      `json:"removed_benchmarks"`
	OverallChange      float64       `json:"overall_change_percent"`
	PerformanceScore   float64       `json:"performance_score"`
}

// PerformanceTrend represents performance trends over time
type PerformanceTrend struct {
	BenchmarkName    string                 `json:"benchmark_name"`
	DataPoints       []TrendDataPoint       `json:"data_points"`
	TrendDirection   string                 `json:"trend_direction"` // improving, degrading, stable
	LinearRegression LinearRegressionResult `json:"linear_regression"`
	ConfidenceLevel  float64                `json:"confidence_level"`
	Prediction       TrendPrediction        `json:"prediction"`
}

// TrendDataPoint represents a single data point in a performance trend
type TrendDataPoint struct {
	Timestamp   time.Time `json:"timestamp"`
	OpsPerSec   float64   `json:"ops_per_sec"`
	MemoryUsage uint64    `json:"memory_usage"`
	GitCommit   string    `json:"git_commit,omitempty"`
}

// LinearRegressionResult contains linear regression analysis results
type LinearRegressionResult struct {
	Slope         float64 `json:"slope"`
	Intercept     float64 `json:"intercept"`
	RSquared      float64 `json:"r_squared"`
	PValue        float64 `json:"p_value"`
	IsSignificant bool    `json:"is_significant"`
}

// TrendPrediction provides performance trend predictions
type TrendPrediction struct {
	NextWeekChange    float64 `json:"next_week_change_percent"`
	NextMonthChange   float64 `json:"next_month_change_percent"`
	RecommendedAction string  `json:"recommended_action"`
}

// SnapshotSummary provides high-level summary of snapshot analysis
type SnapshotSummary struct {
	OverallHealthScore float64  `json:"overall_health_score"`
	TrendingUp         int      `json:"trending_up"`
	TrendingDown       int      `json:"trending_down"`
	StableTrends       int      `json:"stable_trends"`
	CriticalIssues     int      `json:"critical_issues"`
	RecommendedActions []string `json:"recommended_actions"`
}

// SnapshotManager manages performance snapshots and analysis
type SnapshotManager struct {
	snapshotDir string
	logger      logger.CommonLogger
}

// NewSnapshotManager creates a new snapshot manager
func NewSnapshotManager(snapshotDir string) *SnapshotManager {
	return &SnapshotManager{
		snapshotDir: snapshotDir,
		logger:      logger.NewSimpleLogger("snapshot-manager"),
	}
}

// CreateSnapshot creates a new performance snapshot
func (sm *SnapshotManager) CreateSnapshot(ctx context.Context, benchmarks []profiling.BenchmarkResult, metadata map[string]interface{}) (*PerformanceSnapshot, error) {
	// Ensure snapshot directory exists
	if err := os.MkdirAll(sm.snapshotDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create snapshot directory: %w", err)
	}

	snapshot := &PerformanceSnapshot{
		ID:          generateSnapshotID(),
		Timestamp:   time.Now(),
		Environment: getBenchmarkEnvironment(),
		Benchmarks:  benchmarks,
		Metadata:    metadata,
	}

	// Add git information if available
	if commit, err := getGitCommit(); err == nil {
		snapshot.GitCommit = commit
	}
	if branch, err := getGitBranch(); err == nil {
		snapshot.GitBranch = branch
	}

	// Save snapshot to disk
	if err := sm.saveSnapshot(snapshot); err != nil {
		return nil, fmt.Errorf("failed to save snapshot: %w", err)
	}

	sm.logger.Info("Created performance snapshot", "id", snapshot.ID, "benchmarks", len(benchmarks))
	return snapshot, nil
}

// LoadSnapshot loads a snapshot by ID
func (sm *SnapshotManager) LoadSnapshot(snapshotID string) (*PerformanceSnapshot, error) {
	filename := filepath.Join(sm.snapshotDir, fmt.Sprintf("%s.json", snapshotID))
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read snapshot file: %w", err)
	}

	var snapshot PerformanceSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil, fmt.Errorf("failed to parse snapshot: %w", err)
	}

	return &snapshot, nil
}

// ListSnapshots returns all available snapshots sorted by timestamp
func (sm *SnapshotManager) ListSnapshots() ([]*PerformanceSnapshot, error) {
	files, err := filepath.Glob(filepath.Join(sm.snapshotDir, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to list snapshot files: %w", err)
	}

	snapshots := make([]*PerformanceSnapshot, 0, len(files))
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			sm.logger.Warn("Failed to read snapshot file", "file", file, "error", err)
			continue
		}

		var snapshot PerformanceSnapshot
		if err := json.Unmarshal(data, &snapshot); err != nil {
			sm.logger.Warn("Failed to parse snapshot", "file", file, "error", err)
			continue
		}

		snapshots = append(snapshots, &snapshot)
	}

	// Sort by timestamp (newest first)
	sort.Slice(snapshots, func(i, j int) bool {
		return snapshots[i].Timestamp.After(snapshots[j].Timestamp)
	})

	return snapshots, nil
}

// AnalyzeSnapshots performs comprehensive analysis between current and baseline snapshots
func (sm *SnapshotManager) AnalyzeSnapshots(current, baseline *PerformanceSnapshot, options AnalysisOptions) (*SnapshotAnalysis, error) {
	analysis := &SnapshotAnalysis{
		Current:  current,
		Baseline: baseline,
	}

	// Perform comparison analysis
	if err := sm.performComparison(analysis, options); err != nil {
		return nil, fmt.Errorf("failed to perform comparison: %w", err)
	}

	// Analyze trends if historical data is available
	if options.IncludeTrends {
		if err := sm.analyzeTrends(analysis, options); err != nil {
			sm.logger.Warn("Failed to analyze trends", "error", err)
		}
	}

	// Generate summary and recommendations
	sm.generateAnalysisSummary(analysis)

	return analysis, nil
}

// AnalysisOptions configures snapshot analysis behavior
type AnalysisOptions struct {
	RegressionThreshold float64
	IncludeTrends       bool
	TrendWindowDays     int
	ConfidenceLevel     float64
	GeneratePredictions bool
}

// DefaultAnalysisOptions returns default analysis options
func DefaultAnalysisOptions() AnalysisOptions {
	return AnalysisOptions{
		RegressionThreshold: 10.0, // 10% threshold
		IncludeTrends:       true,
		TrendWindowDays:     30, // 30 days
		ConfidenceLevel:     0.95,
		GeneratePredictions: true,
	}
}

func (sm *SnapshotManager) performComparison(analysis *SnapshotAnalysis, options AnalysisOptions) error {
	current := analysis.Current
	baseline := analysis.Baseline

	comparison := SnapshotComparison{
		TimeDifference:     current.Timestamp.Sub(baseline.Timestamp),
		BenchmarkCount:     len(current.Benchmarks),
		BaselineBenchmarks: len(baseline.Benchmarks),
		EnvironmentChanged: !environmentEqual(current.Environment, baseline.Environment),
	}

	// Create benchmark lookup maps
	currentMap := make(map[string]profiling.BenchmarkResult)
	baselineMap := make(map[string]profiling.BenchmarkResult)

	for _, bench := range current.Benchmarks {
		currentMap[bench.Name] = bench
	}
	for _, bench := range baseline.Benchmarks {
		baselineMap[bench.Name] = bench
	}

	// Find new and removed benchmarks
	for name := range currentMap {
		if _, exists := baselineMap[name]; !exists {
			comparison.NewBenchmarks = append(comparison.NewBenchmarks, name)
		}
	}
	for name := range baselineMap {
		if _, exists := currentMap[name]; !exists {
			comparison.RemovedBenchmarks = append(comparison.RemovedBenchmarks, name)
		}
	}

	// Analyze performance changes
	var totalChange float64
	changeCount := 0

	for name, currentBench := range currentMap {
		if baselineBench, exists := baselineMap[name]; exists {
			changePercent := calculatePerformanceChange(currentBench, baselineBench)
			totalChange += changePercent
			changeCount++

			// Check for regressions and improvements
			if changePercent < -options.RegressionThreshold {
				regression := PerformanceRegression{
					BenchmarkName:     name,
					CurrentOpsPerSec:  currentBench.OpsPerSec,
					BaselineOpsPerSec: baselineBench.OpsPerSec,
					RegressionPercent: -changePercent,
					Severity:          calculateSeverity(-changePercent, options.RegressionThreshold),
					Impact:            generateImpactDescription(changePercent),
				}
				analysis.Regressions = append(analysis.Regressions, regression)
			} else if changePercent > options.RegressionThreshold {
				improvement := PerformanceImprovement{
					BenchmarkName:      name,
					CurrentOpsPerSec:   currentBench.OpsPerSec,
					BaselineOpsPerSec:  baselineBench.OpsPerSec,
					ImprovementPercent: changePercent,
					Impact:             generateImpactDescription(changePercent),
				}
				analysis.Improvements = append(analysis.Improvements, improvement)
			}
		}
	}

	if changeCount > 0 {
		comparison.OverallChange = totalChange / float64(changeCount)
	}

	// Calculate performance score
	comparison.PerformanceScore = calculateSnapshotPerformanceScore(analysis.Regressions, analysis.Improvements, len(currentMap))

	analysis.Comparison = comparison
	return nil
}

func (sm *SnapshotManager) analyzeTrends(analysis *SnapshotAnalysis, options AnalysisOptions) error {
	// Load historical snapshots for trend analysis
	snapshots, err := sm.ListSnapshots()
	if err != nil {
		return fmt.Errorf("failed to load historical snapshots: %w", err)
	}

	// Filter snapshots within the trend window
	cutoffTime := time.Now().AddDate(0, 0, -options.TrendWindowDays)
	historicalSnapshots := make([]*PerformanceSnapshot, 0)

	for _, snapshot := range snapshots {
		if snapshot.Timestamp.After(cutoffTime) {
			historicalSnapshots = append(historicalSnapshots, snapshot)
		}
	}

	if len(historicalSnapshots) < 3 {
		sm.logger.Warn("Insufficient historical data for trend analysis", "count", len(historicalSnapshots))
		return nil
	}

	// Analyze trends for each benchmark
	benchmarkNames := make(map[string]bool)
	for _, snapshot := range historicalSnapshots {
		for _, bench := range snapshot.Benchmarks {
			benchmarkNames[bench.Name] = true
		}
	}

	for benchmarkName := range benchmarkNames {
		trend, err := sm.analyzeBenchmarkTrend(benchmarkName, historicalSnapshots, options)
		if err != nil {
			sm.logger.Warn("Failed to analyze trend", "benchmark", benchmarkName, "error", err)
			continue
		}

		if trend != nil {
			analysis.Trends = append(analysis.Trends, *trend)
		}
	}

	return nil
}

func (sm *SnapshotManager) analyzeBenchmarkTrend(benchmarkName string, snapshots []*PerformanceSnapshot, options AnalysisOptions) (*PerformanceTrend, error) {
	dataPoints := make([]TrendDataPoint, 0)

	// Extract data points for this benchmark
	for _, snapshot := range snapshots {
		for _, bench := range snapshot.Benchmarks {
			if bench.Name == benchmarkName {
				dataPoint := TrendDataPoint{
					Timestamp:   snapshot.Timestamp,
					OpsPerSec:   bench.OpsPerSec,
					MemoryUsage: bench.MemoryAfter - bench.MemoryBefore,
					GitCommit:   snapshot.GitCommit,
				}
				dataPoints = append(dataPoints, dataPoint)
				break
			}
		}
	}

	if len(dataPoints) < 3 {
		return nil, errors.New("insufficient data points for trend analysis")
	}

	// Sort by timestamp
	sort.Slice(dataPoints, func(i, j int) bool {
		return dataPoints[i].Timestamp.Before(dataPoints[j].Timestamp)
	})

	trend := &PerformanceTrend{
		BenchmarkName: benchmarkName,
		DataPoints:    dataPoints,
	}

	// Perform linear regression analysis
	regression := performLinearRegression(dataPoints)
	trend.LinearRegression = regression

	// Determine trend direction
	if regression.IsSignificant {
		if regression.Slope > 0 {
			trend.TrendDirection = "improving"
		} else if regression.Slope < 0 {
			trend.TrendDirection = "degrading"
		} else {
			trend.TrendDirection = "stable"
		}
	} else {
		trend.TrendDirection = "stable"
	}

	// Calculate confidence level
	trend.ConfidenceLevel = options.ConfidenceLevel

	// Generate predictions if requested
	if options.GeneratePredictions {
		trend.Prediction = generateTrendPrediction(regression, dataPoints)
	}

	return trend, nil
}

func (sm *SnapshotManager) generateAnalysisSummary(analysis *SnapshotAnalysis) {
	summary := SnapshotSummary{
		RecommendedActions: make([]string, 0),
	}

	// Count trend directions
	for _, trend := range analysis.Trends {
		switch trend.TrendDirection {
		case "improving":
			summary.TrendingUp++
		case "degrading":
			summary.TrendingDown++
		case "stable":
			summary.StableTrends++
		}
	}

	// Count critical issues
	for _, reg := range analysis.Regressions {
		if reg.Severity == "critical" {
			summary.CriticalIssues++
		}
	}

	// Calculate overall health score
	baseScore := 100.0
	if len(analysis.Regressions) > 0 {
		regressionPenalty := float64(len(analysis.Regressions)) * 5.0
		baseScore -= regressionPenalty
	}
	if summary.CriticalIssues > 0 {
		baseScore -= float64(summary.CriticalIssues) * 20.0
	}
	if summary.TrendingDown > 0 {
		baseScore -= float64(summary.TrendingDown) * 10.0
	}
	if baseScore < 0 {
		baseScore = 0
	}

	summary.OverallHealthScore = baseScore

	// Generate recommendations
	if summary.CriticalIssues > 0 {
		summary.RecommendedActions = append(summary.RecommendedActions,
			fmt.Sprintf("Address %d critical performance regressions immediately", summary.CriticalIssues))
	}
	if summary.TrendingDown > 0 {
		summary.RecommendedActions = append(summary.RecommendedActions,
			fmt.Sprintf("Investigate %d degrading performance trends", summary.TrendingDown))
	}
	if len(analysis.Improvements) > 0 {
		summary.RecommendedActions = append(summary.RecommendedActions,
			fmt.Sprintf("Document and preserve %d performance improvements", len(analysis.Improvements)))
	}
	if len(summary.RecommendedActions) == 0 {
		summary.RecommendedActions = append(summary.RecommendedActions,
			"Performance appears stable - continue monitoring")
	}

	analysis.Summary = summary
}

// Helper functions

func (sm *SnapshotManager) saveSnapshot(snapshot *PerformanceSnapshot) error {
	filename := filepath.Join(sm.snapshotDir, fmt.Sprintf("%s.json", snapshot.ID))
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot: %w", err)
	}

	return os.WriteFile(filename, data, 0o600)
}

func generateSnapshotID() string {
	return fmt.Sprintf("snapshot-%d", time.Now().Unix())
}

func environmentEqual(env1, env2 BenchmarkEnvironment) bool {
	return env1.Platform == env2.Platform &&
		env1.GoVersion == env2.GoVersion &&
		env1.NumCPU == env2.NumCPU
}

func calculatePerformanceChange(current, baseline profiling.BenchmarkResult) float64 {
	if baseline.OpsPerSec == 0 {
		return 0
	}
	return ((current.OpsPerSec - baseline.OpsPerSec) / baseline.OpsPerSec) * 100
}

func calculateSeverity(regressionPercent, threshold float64) string {
	switch {
	case regressionPercent >= threshold*3:
		return "critical"
	case regressionPercent >= threshold*2:
		return "high"
	case regressionPercent >= threshold:
		return "medium"
	default:
		return "low"
	}
}

func calculateSnapshotPerformanceScore(regressions []PerformanceRegression, improvements []PerformanceImprovement, totalBenchmarks int) float64 {
	if totalBenchmarks == 0 {
		return 0
	}

	score := 100.0

	// Penalize regressions
	for _, reg := range regressions {
		switch reg.Severity {
		case "critical":
			score -= 30
		case "high":
			score -= 20
		case "medium":
			score -= 10
		case "low":
			score -= 5
		}
	}

	// Reward improvements (but cap the benefit)
	improvementBonus := float64(len(improvements)) * 5.0
	if improvementBonus > 20.0 {
		improvementBonus = 20.0
	}
	score += improvementBonus

	if score < 0 {
		score = 0
	} else if score > 100 {
		score = 100
	}

	return score
}

func performLinearRegression(dataPoints []TrendDataPoint) LinearRegressionResult {
	n := len(dataPoints)
	if n < 2 {
		return LinearRegressionResult{}
	}

	// Convert timestamps to numeric values (days since first point)
	baseTime := dataPoints[0].Timestamp
	var sumX, sumY, sumXY, sumX2 float64

	for _, point := range dataPoints {
		x := point.Timestamp.Sub(baseTime).Hours() / 24.0 // Days
		y := point.OpsPerSec

		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	// Calculate slope and intercept
	nFloat := float64(n)
	slope := (nFloat*sumXY - sumX*sumY) / (nFloat*sumX2 - sumX*sumX)
	intercept := (sumY - slope*sumX) / nFloat

	// Calculate R-squared
	meanY := sumY / nFloat
	var ssTotal, ssRes float64

	for _, point := range dataPoints {
		x := point.Timestamp.Sub(baseTime).Hours() / 24.0
		y := point.OpsPerSec
		predicted := slope*x + intercept

		ssTotal += (y - meanY) * (y - meanY)
		ssRes += (y - predicted) * (y - predicted)
	}

	var rSquared float64
	if ssTotal > 0 {
		rSquared = 1 - (ssRes / ssTotal)
	}

	// Simple significance test (R² > 0.5 and slope magnitude > threshold)
	isSignificant := rSquared > 0.5 && (slope > 1.0 || slope < -1.0)

	return LinearRegressionResult{
		Slope:         slope,
		Intercept:     intercept,
		RSquared:      rSquared,
		PValue:        0.05, // Simplified - would need proper statistical test
		IsSignificant: isSignificant,
	}
}

func generateTrendPrediction(regression LinearRegressionResult, dataPoints []TrendDataPoint) TrendPrediction {
	if !regression.IsSignificant {
		return TrendPrediction{
			RecommendedAction: "Continue monitoring - trend not statistically significant",
		}
	}

	// Predict changes based on regression slope
	weeklyChange := regression.Slope * 7   // 7 days
	monthlyChange := regression.Slope * 30 // 30 days

	// Convert to percentages (approximate)
	if len(dataPoints) > 0 {
		currentValue := dataPoints[len(dataPoints)-1].OpsPerSec
		if currentValue > 0 {
			weeklyChange = (weeklyChange / currentValue) * 100
			monthlyChange = (monthlyChange / currentValue) * 100
		}
	}

	var action string
	if regression.Slope < -10 { // Degrading significantly
		action = "Immediate investigation required - performance degrading rapidly"
	} else if regression.Slope < -1 {
		action = "Monitor closely - performance showing downward trend"
	} else if regression.Slope > 10 {
		action = "Document optimization - performance improving significantly"
	} else {
		action = "Continue current practices - performance trend is stable"
	}

	return TrendPrediction{
		NextWeekChange:    weeklyChange,
		NextMonthChange:   monthlyChange,
		RecommendedAction: action,
	}
}
