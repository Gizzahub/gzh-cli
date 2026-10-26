// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package doctor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/gizzahub/gzh-cli/internal/cli"
	"github.com/gizzahub/gzh-cli/internal/logger"
)

// CodeQualityReport represents a comprehensive code quality analysis report.
type CodeQualityReport struct {
	Timestamp       time.Time         `json:"timestamp"`
	ProjectPath     string            `json:"project_path"`
	Summary         QualitySummary    `json:"summary"`
	Metrics         QualityMetrics    `json:"metrics"`
	Issues          []QualityIssue    `json:"issues"`
	Trends          QualityTrends     `json:"trends"`
	Scores          QualityScores     `json:"scores"`
	Recommendations []string          `json:"recommendations"`
	FileAnalysis    []FileQualityInfo `json:"file_analysis"`
}

// QualitySummary provides high-level quality overview.
type QualitySummary struct {
	TotalFiles     int     `json:"total_files"`
	TotalLines     int     `json:"total_lines"`
	CodeLines      int     `json:"code_lines"`
	CommentLines   int     `json:"comment_lines"`
	TestCoverage   float64 `json:"test_coverage"`
	IssueCount     int     `json:"issue_count"`
	CriticalIssues int     `json:"critical_issues"`
	TechnicalDebt  string  `json:"technical_debt"`
	OverallScore   float64 `json:"overall_score"`
}

// QualityMetrics contains detailed code quality metrics.
type QualityMetrics struct {
	CyclomaticComplexity  int     `json:"cyclomatic_complexity"`
	AverageComplexity     float64 `json:"average_complexity"`
	DuplicationRatio      float64 `json:"duplication_ratio"`
	TestableCodeRatio     float64 `json:"testable_code_ratio"`
	DocumentationCoverage float64 `json:"documentation_coverage"`
	DependencyCount       int     `json:"dependency_count"`
	TechnicalDebtRatio    float64 `json:"technical_debt_ratio"`
	MaintainabilityIndex  float64 `json:"maintainability_index"`
}

// QualityIssue represents a code quality issue.
type QualityIssue struct {
	Type       string `json:"type"`
	Severity   string `json:"severity"`
	Category   string `json:"category"`
	File       string `json:"file"`
	Line       int    `json:"line"`
	Message    string `json:"message"`
	Rule       string `json:"rule"`
	Suggestion string `json:"suggestion"`
}

// QualityTrends tracks quality changes over time.
type QualityTrends struct {
	ScoreTrend      string  `json:"score_trend"`
	IssueCountTrend string  `json:"issue_count_trend"`
	CoverageTrend   string  `json:"coverage_trend"`
	ComplexityTrend string  `json:"complexity_trend"`
	WeeklyChange    float64 `json:"weekly_change"`
	MonthlyChange   float64 `json:"monthly_change"`
}

// QualityScores contains various quality scoring metrics.
type QualityScores struct {
	Maintainability float64 `json:"maintainability"`
	Reliability     float64 `json:"reliability"`
	Security        float64 `json:"security"`
	Performance     float64 `json:"performance"`
	Testability     float64 `json:"testability"`
	Documentation   float64 `json:"documentation"`
}

// FileQualityInfo contains per-file quality analysis.
type FileQualityInfo struct {
	Path         string    `json:"path"`
	Lines        int       `json:"lines"`
	Functions    int       `json:"functions"`
	Complexity   int       `json:"complexity"`
	IssueCount   int       `json:"issue_count"`
	TestCoverage float64   `json:"test_coverage"`
	QualityScore float64   `json:"quality_score"`
	LastModified time.Time `json:"last_modified"`
}

// newMetricsCmd creates the metrics subcommand for code quality analysis.
func newMetricsCmd() *cobra.Command {
	var (
		projectPath    string
		includeTests   bool
		outputFile     string
		threshold      float64
		historicalDays int
		detailedReport bool
		skipComplexity bool
		onlyIssues     bool
	)

	cmd := cli.NewCommandBuilder("metrics", "Analyze code quality metrics and generate dashboard").
		WithLongDescription(`Analyze comprehensive code quality metrics and generate an interactive dashboard.

This command provides detailed code quality analysis including:
- Code complexity analysis and maintainability metrics
- Test coverage integration and quality scoring
- Technical debt assessment and recommendations
- Issue tracking and trend analysis over time
- File-level quality breakdown and hotspot identification
- Integration with golangci-lint and go test coverage

Features:
- Comprehensive quality scoring across multiple dimensions
- Historical trend analysis and quality regression detection
- Actionable recommendations for code improvement
- Integration with existing CI/CD quality gates
- Detailed per-file analysis for targeted improvements
- Technical debt calculation and prioritization

Examples:
  gz doctor metrics                                    # Analyze current directory
  gz doctor metrics --path ./internal                 # Analyze specific path
  gz doctor metrics --threshold 80.0                  # Set quality threshold
  gz doctor metrics --detailed --output quality.json  # Generate detailed report
  gz doctor metrics --historical 30                   # Include 30-day trend analysis`).
		WithExample("gz doctor metrics --detailed --threshold 85.0").
		WithFormatFlag("table", []string{"table", "json", "yaml"}).
		WithRunFuncE(func(ctx context.Context, flags *cli.CommonFlags, args []string) error {
			return runQualityMetricsAnalysis(ctx, flags, metricsOptions{
				projectPath:    projectPath,
				includeTests:   includeTests,
				outputFile:     outputFile,
				threshold:      threshold,
				historicalDays: historicalDays,
				detailedReport: detailedReport,
				skipComplexity: skipComplexity,
				onlyIssues:     onlyIssues,
			})
		}).
		Build()

	cmd.Flags().StringVar(&projectPath, "path", ".", "Project path to analyze")
	cmd.Flags().BoolVar(&includeTests, "include-tests", true, "Include test files in analysis")
	cmd.Flags().StringVar(&outputFile, "output", "", "Output file for detailed report")
	cmd.Flags().Float64Var(&threshold, "threshold", 75.0, "Quality score threshold")
	cmd.Flags().IntVar(&historicalDays, "historical", 7, "Days of historical data to analyze")
	cmd.Flags().BoolVar(&detailedReport, "detailed", false, "Generate detailed per-file analysis")
	cmd.Flags().BoolVar(&skipComplexity, "skip-complexity", false, "Skip complexity analysis")
	cmd.Flags().BoolVar(&onlyIssues, "only-issues", false, "Show only issues and recommendations")

	return cmd
}

type metricsOptions struct {
	projectPath    string
	includeTests   bool
	outputFile     string
	threshold      float64
	historicalDays int
	detailedReport bool
	skipComplexity bool
	onlyIssues     bool
}

type projectRootReader struct {
	projectPath string
	root        *os.Root
}

func openProjectRootReader(projectPath string) (*projectRootReader, error) {
	root, err := os.OpenRoot(projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open project root: %w", err)
	}

	return &projectRootReader{projectPath: projectPath, root: root}, nil
}

func (r *projectRootReader) Close() error {
	return r.root.Close()
}

func (r *projectRootReader) ReadFile(path string) ([]byte, error) {
	relativePath, err := filepath.Rel(r.projectPath, path)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve project-relative path %q: %w", path, err)
	}

	content, err := r.root.ReadFile(relativePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read project file %q: %w", relativePath, err)
	}

	return content, nil
}

type metricsCommandRunner func(context.Context, string, string, ...string) ([]byte, error)

type metricsDependencies struct {
	run    metricsCommandRunner
	remove func(string) error
}

func defaultMetricsDependencies() metricsDependencies {
	return metricsDependencies{run: runMetricsCommand, remove: os.Remove}
}

func runMetricsCommand(ctx context.Context, dir, executable string, args ...string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Dir = dir
	output, err := cmd.Output()
	if ctx.Err() != nil {
		return output, fmt.Errorf("%s canceled: %w", executable, ctx.Err())
	}
	if err != nil {
		return output, fmt.Errorf("%s failed: %w", executable, err)
	}
	return output, nil
}

func runQualityMetricsAnalysis(ctx context.Context, flags *cli.CommonFlags, opts metricsOptions) error {
	return runQualityMetricsAnalysisWithDependencies(ctx, flags, opts, defaultMetricsDependencies())
}

func runQualityMetricsAnalysisWithDependencies(ctx context.Context, flags *cli.CommonFlags, opts metricsOptions, deps metricsDependencies) error {
	logger := logger.NewSimpleLogger("doctor-metrics")

	logger.Info(
		"Starting code quality metrics analysis",
		"project_path", opts.projectPath,
		"threshold", opts.threshold,
		"detailed", opts.detailedReport,
	)

	// Initialize quality report
	report := &CodeQualityReport{
		Timestamp:    time.Now(),
		ProjectPath:  opts.projectPath,
		Issues:       make([]QualityIssue, 0),
		FileAnalysis: make([]FileQualityInfo, 0),
	}

	// 분석 root를 한 번 고정해 walker callback의 경로가 바뀌어도 root 밖으로 나가지 못하게 한다.
	projectFiles, err := openProjectRootReader(opts.projectPath)
	if err != nil {
		return err
	}
	defer projectFiles.Close()

	// Collect basic project metrics
	if err := collectProjectMetrics(report, opts, projectFiles); err != nil {
		return fmt.Errorf("failed to collect basic metrics: %w", err)
	}

	// Run linting analysis
	if err := collectLintingIssuesWithRunner(ctx, report, opts, deps.run); err != nil {
		return fmt.Errorf("failed to collect linting issues: %w", err)
	}

	// Analyze test coverage
	if err := collectCoverageMetricsWithDependencies(ctx, report, opts, deps); err != nil {
		return fmt.Errorf("failed to collect coverage metrics: %w", err)
	}

	// Analyze code complexity
	if !opts.skipComplexity {
		if err := collectComplexityMetrics(report, opts, projectFiles); err != nil {
			return fmt.Errorf("failed to collect complexity metrics: %w", err)
		}
	}

	// Perform detailed file analysis
	if opts.detailedReport {
		if err := collectFileAnalysis(report, opts, projectFiles); err != nil {
			return fmt.Errorf("failed to collect file analysis: %w", err)
		}
	}

	// Calculate quality scores and trends
	calculateQualityScores(report)
	generateQualityRecommendations(report, opts.threshold)

	// Save detailed report if requested
	if opts.outputFile != "" {
		if err := saveQualityReport(report, opts.outputFile); err != nil {
			return fmt.Errorf("failed to save report: %w", err)
		}
		logger.Info("Quality report saved", "file", opts.outputFile)
	}

	// Display results
	formatter := cli.NewOutputFormatter(flags.Format)

	switch flags.Format {
	case "json", "yaml":
		return formatter.FormatOutput(report)
	default:
		return displayQualityResults(report, opts)
	}
}

func collectProjectMetrics(report *CodeQualityReport, opts metricsOptions, projectFiles *projectRootReader) error {
	// Count Go files and lines
	var totalFiles, totalLines, codeLines, commentLines int

	err := filepath.Walk(opts.projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip non-Go files and vendor directories
		if !strings.HasSuffix(path, ".go") || strings.Contains(path, "vendor/") {
			return nil
		}

		// Skip test files if not included
		if !opts.includeTests && strings.HasSuffix(path, "_test.go") {
			return nil
		}

		totalFiles++

		// Count lines in file
		content, err := projectFiles.ReadFile(path)
		if err != nil {
			return err
		}

		lines := strings.Split(string(content), "\n")
		totalLines += len(lines)

		// Simple heuristic for code vs comment lines
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				continue
			}
			if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") {
				commentLines++
			} else {
				codeLines++
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to walk project directory: %w", err)
	}

	report.Summary.TotalFiles = totalFiles
	report.Summary.TotalLines = totalLines
	report.Summary.CodeLines = codeLines
	report.Summary.CommentLines = commentLines

	return nil
}

func collectLintingIssues(ctx context.Context, report *CodeQualityReport, opts metricsOptions) error {
	return collectLintingIssuesWithRunner(ctx, report, opts, runMetricsCommand)
}

func collectLintingIssuesWithRunner(ctx context.Context, report *CodeQualityReport, opts metricsOptions, run metricsCommandRunner) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	// v2의 명시적 이슈 종료 코드로 정상 진단 결과와 실행 실패를 구분한다.
	output, err := run(ctx, opts.projectPath, "golangci-lint", "run", "--output.json.path", "stdout", "--show-stats=false", "--issues-exit-code=1", "./...")
	if ctx.Err() != nil {
		return fmt.Errorf("lint analysis canceled: %w", ctx.Err())
	}
	if err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
			return fmt.Errorf("lint analysis failed: %w", err)
		}
	}
	var lintResult struct {
		Issues *[]struct {
			FromLinter string `json:"FromLinter"`
			Severity   string `json:"Severity"`
			Text       string `json:"Text"`
			Pos        struct {
				Filename string `json:"Filename"`
				Line     int    `json:"Line"`
			} `json:"Pos"`
		} `json:"Issues"`
	}
	if decodeErr := json.Unmarshal(output, &lintResult); decodeErr != nil {
		return fmt.Errorf("invalid lint JSON (expected one JSON report; check configured output formats): %w", decodeErr)
	}
	if lintResult.Issues == nil {
		return fmt.Errorf("lint report is missing a non-null Issues array")
	}
	if err != nil && len(*lintResult.Issues) == 0 {
		return fmt.Errorf("lint issue exit without findings: %w", err)
	}

	// Convert to quality issues
	for _, issue := range *lintResult.Issues {
		var severity string
		switch issue.Severity {
		case "error":
			severity = "high"
		case "warning":
			severity = "medium"
		default:
			severity = "low"
		}

		qualityIssue := QualityIssue{
			Type:     "lint",
			Severity: severity,
			Category: issue.FromLinter,
			File:     issue.Pos.Filename,
			Line:     issue.Pos.Line,
			Message:  issue.Text,
			Rule:     issue.FromLinter,
		}

		report.Issues = append(report.Issues, qualityIssue)

		if severity == "high" {
			report.Summary.CriticalIssues++
		}
	}

	report.Summary.IssueCount = len(report.Issues)
	return nil
}

func collectCoverageMetrics(ctx context.Context, report *CodeQualityReport, opts metricsOptions) error {
	return collectCoverageMetricsWithDependencies(ctx, report, opts, defaultMetricsDependencies())
}

func collectCoverageMetricsWithDependencies(ctx context.Context, report *CodeQualityReport, opts metricsOptions, deps metricsDependencies) (resultErr error) {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return err
	}
	profileFile, err := os.CreateTemp("", "gz-metrics-coverage-*.out")
	if err != nil {
		return fmt.Errorf("create coverage profile: %w", err)
	}
	profile := profileFile.Name()
	defer func() {
		if cleanupErr := deps.remove(profile); cleanupErr != nil {
			resultErr = errors.Join(resultErr, fmt.Errorf("clean coverage profile: %w", cleanupErr))
		}
	}()
	if err := profileFile.Close(); err != nil {
		return fmt.Errorf("close coverage profile: %w", err)
	}

	if _, err = deps.run(ctx, opts.projectPath, "go", "test", "-coverprofile="+profile, "./..."); err != nil {
		return fmt.Errorf("coverage tests failed: %w", err)
	}
	output, err := deps.run(ctx, opts.projectPath, "go", "tool", "cover", "-func="+profile)
	if err != nil {
		return fmt.Errorf("coverage report failed: %w", err)
	}
	coverage, err := parseMetricsCoverage(output)
	if err != nil {
		return err
	}
	report.Summary.TestCoverage = coverage
	return nil
}

func parseMetricsCoverage(output []byte) (float64, error) {
	found := false
	var coverage float64
	for _, line := range strings.Split(string(output), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 || fields[0] != "total:" {
			continue
		}
		if found || len(fields) != 3 || fields[1] != "(statements)" || !strings.HasSuffix(fields[2], "%") {
			return 0, fmt.Errorf("malformed or duplicate coverage total")
		}
		value, err := strconv.ParseFloat(strings.TrimSuffix(fields[2], "%"), 64)
		if err != nil || math.IsNaN(value) || math.IsInf(value, 0) || value < 0 || value > 100 {
			return 0, fmt.Errorf("invalid coverage total %q", fields[2])
		}
		coverage, found = value, true
	}
	if !found {
		return 0, fmt.Errorf("coverage report has no total")
	}
	return coverage, nil
}

func collectComplexityMetrics(report *CodeQualityReport, opts metricsOptions, projectFiles *projectRootReader) error {
	// Use gocyclo or similar tool to analyze complexity
	// For now, implement a simple complexity estimation
	var totalComplexity, functionCount int

	err := filepath.Walk(opts.projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || !strings.HasSuffix(path, ".go") {
			return err
		}

		if strings.Contains(path, "vendor/") {
			return nil
		}

		if !opts.includeTests && strings.HasSuffix(path, "_test.go") {
			return nil
		}

		content, err := projectFiles.ReadFile(path)
		if err != nil {
			return err
		}

		// Simple complexity calculation based on control structures
		text := string(content)
		complexity := 1 // Base complexity

		// Count control structures
		complexity += strings.Count(text, "if ")
		complexity += strings.Count(text, "for ")
		complexity += strings.Count(text, "switch ")
		complexity += strings.Count(text, "case ")
		complexity += strings.Count(text, "&&")
		complexity += strings.Count(text, "||")

		// Count functions
		functions := strings.Count(text, "func ")
		functionCount += functions

		if functions > 0 {
			totalComplexity += complexity
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to analyze complexity: %w", err)
	}

	report.Metrics.CyclomaticComplexity = totalComplexity
	if functionCount > 0 {
		report.Metrics.AverageComplexity = float64(totalComplexity) / float64(functionCount)
	}

	return nil
}

func collectFileAnalysis(report *CodeQualityReport, opts metricsOptions, projectFiles *projectRootReader) error {
	err := filepath.Walk(opts.projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || !strings.HasSuffix(path, ".go") {
			return err
		}

		if strings.Contains(path, "vendor/") {
			return nil
		}

		if !opts.includeTests && strings.HasSuffix(path, "_test.go") {
			return nil
		}

		content, err := projectFiles.ReadFile(path)
		if err != nil {
			return err
		}

		lines := strings.Split(string(content), "\n")
		functions := strings.Count(string(content), "func ")

		// Calculate simple complexity for this file
		complexity := 1
		text := string(content)
		complexity += strings.Count(text, "if ")
		complexity += strings.Count(text, "for ")
		complexity += strings.Count(text, "switch ")

		// Count issues in this file
		issueCount := 0
		for _, issue := range report.Issues {
			if issue.File == path {
				issueCount++
			}
		}

		// Calculate quality score for file
		qualityScore := 100.0
		if functions > 0 {
			complexityPenalty := float64(complexity) / float64(functions) * 5
			qualityScore -= complexityPenalty
		}
		qualityScore -= float64(issueCount) * 10

		if qualityScore < 0 {
			qualityScore = 0
		}

		fileInfo := FileQualityInfo{
			Path:         path,
			Lines:        len(lines),
			Functions:    functions,
			Complexity:   complexity,
			IssueCount:   issueCount,
			QualityScore: qualityScore,
			LastModified: info.ModTime(),
		}

		report.FileAnalysis = append(report.FileAnalysis, fileInfo)

		return nil
	})

	return err
}

func calculateQualityScores(report *CodeQualityReport) {
	// Calculate overall quality score
	baseScore := 100.0

	// Penalize based on issues
	if report.Summary.IssueCount > 0 {
		baseScore -= float64(report.Summary.IssueCount) * 2
	}

	// Penalize critical issues more
	baseScore -= float64(report.Summary.CriticalIssues) * 10

	// Bonus for good test coverage
	if report.Summary.TestCoverage > 80 {
		baseScore += 5
	} else if report.Summary.TestCoverage < 50 {
		baseScore -= 15
	}

	// Penalize high complexity
	if report.Metrics.AverageComplexity > 10 {
		baseScore -= 10
	}

	if baseScore < 0 {
		baseScore = 0
	}

	report.Summary.OverallScore = baseScore

	// Calculate individual dimension scores
	report.Scores = QualityScores{
		Maintainability: calculateMaintainabilityScore(report),
		Reliability:     calculateReliabilityScore(report),
		Security:        calculateSecurityScore(report),
		Performance:     calculatePerformanceScore(report),
		Testability:     calculateTestabilityScore(report),
		Documentation:   calculateDocumentationScore(report),
	}

	// Calculate technical debt
	issueHours := float64(report.Summary.IssueCount) * 0.5        // 30 minutes per issue
	criticalHours := float64(report.Summary.CriticalIssues) * 2.0 // 2 hours per critical issue
	totalHours := issueHours + criticalHours

	if totalHours < 1 {
		report.Summary.TechnicalDebt = "< 1 hour"
	} else if totalHours < 8 {
		report.Summary.TechnicalDebt = fmt.Sprintf("%.1f hours", totalHours)
	} else {
		days := totalHours / 8
		report.Summary.TechnicalDebt = fmt.Sprintf("%.1f days", days)
	}

	report.Metrics.TechnicalDebtRatio = totalHours / float64(report.Summary.CodeLines) * 1000 // per 1000 lines
	report.Metrics.MaintainabilityIndex = report.Scores.Maintainability
}

func calculateMaintainabilityScore(report *CodeQualityReport) float64 {
	score := 100.0

	// Factor in complexity
	if report.Metrics.AverageComplexity > 15 {
		score -= 30
	} else if report.Metrics.AverageComplexity > 10 {
		score -= 15
	}

	// Factor in file size
	if report.Summary.TotalFiles > 0 {
		avgLinesPerFile := float64(report.Summary.CodeLines) / float64(report.Summary.TotalFiles)
		if avgLinesPerFile > 300 {
			score -= 15
		}
	}

	if score < 0 {
		score = 0
	}

	return score
}

func calculateReliabilityScore(report *CodeQualityReport) float64 {
	// Factor in test coverage
	score := report.Summary.TestCoverage

	// Penalize for critical issues
	score -= float64(report.Summary.CriticalIssues) * 15

	if score < 0 {
		score = 0
	}

	return score
}

func calculateSecurityScore(report *CodeQualityReport) float64 {
	score := 100.0

	// Count security-related issues
	securityIssues := 0
	for _, issue := range report.Issues {
		if strings.Contains(strings.ToLower(issue.Category), "sec") ||
			strings.Contains(strings.ToLower(issue.Rule), "gosec") {
			securityIssues++
		}
	}

	score -= float64(securityIssues) * 20

	if score < 0 {
		score = 0
	}

	return score
}

func calculatePerformanceScore(report *CodeQualityReport) float64 {
	score := 100.0

	// Factor in complexity as performance indicator
	if report.Metrics.AverageComplexity > 20 {
		score -= 30
	}

	// Count performance-related issues
	perfIssues := 0
	for _, issue := range report.Issues {
		if strings.Contains(strings.ToLower(issue.Message), "performance") ||
			strings.Contains(strings.ToLower(issue.Message), "inefficient") {
			perfIssues++
		}
	}

	score -= float64(perfIssues) * 10

	if score < 0 {
		score = 0
	}

	return score
}

func calculateTestabilityScore(report *CodeQualityReport) float64 {
	return report.Summary.TestCoverage // Direct correlation with test coverage
}

func calculateDocumentationScore(report *CodeQualityReport) float64 {
	if report.Summary.CodeLines == 0 {
		return 0
	}

	// Calculate documentation ratio
	docRatio := float64(report.Summary.CommentLines) / float64(report.Summary.CodeLines)

	score := docRatio * 100 * 5 // Scale up documentation ratio

	if score > 100 {
		score = 100
	}

	return score
}

func generateQualityRecommendations(report *CodeQualityReport, threshold float64) {
	recommendations := make([]string, 0)

	// Overall score recommendations
	if report.Summary.OverallScore < threshold {
		recommendations = append(recommendations,
			fmt.Sprintf("Overall quality score (%.1f) is below threshold (%.1f) - review needed",
				report.Summary.OverallScore, threshold))
	}

	// Critical issues
	if report.Summary.CriticalIssues > 0 {
		recommendations = append(recommendations,
			fmt.Sprintf("Address %d critical issues immediately", report.Summary.CriticalIssues))
	}

	// Test coverage
	if report.Summary.TestCoverage < 70 {
		recommendations = append(recommendations,
			fmt.Sprintf("Increase test coverage from %.1f%% to at least 70%%", report.Summary.TestCoverage))
	}

	// Complexity
	if report.Metrics.AverageComplexity > 10 {
		recommendations = append(recommendations,
			fmt.Sprintf("Reduce average complexity from %.1f to below 10", report.Metrics.AverageComplexity))
	}

	// Issue count
	if report.Summary.IssueCount > 50 {
		recommendations = append(recommendations,
			fmt.Sprintf("Address %d linting issues to improve code quality", report.Summary.IssueCount))
	}

	// Documentation
	if report.Scores.Documentation < 30 {
		recommendations = append(recommendations,
			"Improve code documentation and comments")
	}

	// Security
	if report.Scores.Security < 80 {
		recommendations = append(recommendations,
			"Review and address security-related issues")
	}

	report.Recommendations = recommendations
}

func saveQualityReport(report *CodeQualityReport, filename string) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	return os.WriteFile(filename, data, 0o600)
}

func displayQualityResults(report *CodeQualityReport, opts metricsOptions) error {
	if opts.onlyIssues {
		return displayIssuesOnly(report)
	}

	// Display project overview
	logger.SimpleInfo(
		"🔍 Code Quality Analysis",
		"project", report.ProjectPath,
		"files", report.Summary.TotalFiles,
		"lines", report.Summary.TotalLines,
		"coverage", fmt.Sprintf("%.1f%%", report.Summary.TestCoverage),
	)

	// Display quality scores
	logger.SimpleInfo(
		"📊 Quality Scores",
		"overall", fmt.Sprintf("%.1f/100", report.Summary.OverallScore),
		"maintainability", fmt.Sprintf("%.1f", report.Scores.Maintainability),
		"reliability", fmt.Sprintf("%.1f", report.Scores.Reliability),
		"security", fmt.Sprintf("%.1f", report.Scores.Security),
	)

	// Display metrics
	logger.SimpleInfo(
		"📈 Quality Metrics",
		"issues", report.Summary.IssueCount,
		"critical", report.Summary.CriticalIssues,
		"complexity", fmt.Sprintf("%.1f", report.Metrics.AverageComplexity),
		"tech_debt", report.Summary.TechnicalDebt,
	)

	// Display top issues by severity
	if len(report.Issues) > 0 {
		logger.SimpleWarn("⚠️ Top Quality Issues:")

		// Sort issues by severity
		sort.Slice(report.Issues, func(i, j int) bool {
			severityOrder := map[string]int{"high": 3, "medium": 2, "low": 1}
			return severityOrder[report.Issues[i].Severity] > severityOrder[report.Issues[j].Severity]
		})

		// Show top 10 issues
		maxIssues := min(len(report.Issues), 10)

		for i := range maxIssues {
			issue := report.Issues[i]
			severityIcon := "🟡"
			switch issue.Severity {
			case "high":
				severityIcon = "🔴"
			case "low":
				severityIcon = "🟢"
			}

			logger.SimpleWarn(
				fmt.Sprintf("  %s %s:%d", severityIcon, issue.File, issue.Line),
				"rule", issue.Rule,
				"message", issue.Message,
			)
		}

		if len(report.Issues) > maxIssues {
			logger.SimpleInfo(fmt.Sprintf("  ... and %d more issues", len(report.Issues)-maxIssues))
		}
	}

	// Display recommendations
	if len(report.Recommendations) > 0 {
		logger.SimpleInfo("💡 Recommendations:")
		for _, rec := range report.Recommendations {
			logger.SimpleInfo(fmt.Sprintf("  • %s", rec))
		}
	}

	// Display file analysis if detailed
	if opts.detailedReport && len(report.FileAnalysis) > 0 {
		logger.SimpleInfo("📄 File Quality Analysis:")

		// Sort by quality score (worst first)
		sort.Slice(report.FileAnalysis, func(i, j int) bool {
			return report.FileAnalysis[i].QualityScore < report.FileAnalysis[j].QualityScore
		})

		// Show top 10 files with issues
		maxFiles := min(len(report.FileAnalysis), 10)

		for i := range maxFiles {
			file := report.FileAnalysis[i]
			if file.QualityScore > 80 {
				break // Stop at high quality files
			}

			logger.SimpleInfo(
				fmt.Sprintf("  📄 %s", file.Path),
				"score", fmt.Sprintf("%.1f", file.QualityScore),
				"issues", file.IssueCount,
				"complexity", file.Complexity,
				"lines", file.Lines,
			)
		}
	}

	// Exit with error code if quality is below threshold
	if report.Summary.OverallScore < opts.threshold {
		logger.SimpleWarn(fmt.Sprintf("Quality score %.1f is below threshold %.1f",
			report.Summary.OverallScore, opts.threshold))
		return fmt.Errorf("quality score %.1f below threshold %.1f",
			report.Summary.OverallScore, opts.threshold)
	}

	return nil
}

func displayIssuesOnly(report *CodeQualityReport) error {
	if len(report.Issues) == 0 {
		logger.SimpleInfo("✅ No quality issues found!")
		return nil
	}

	// Group issues by category
	issuesByCategory := make(map[string][]QualityIssue)
	for _, issue := range report.Issues {
		issuesByCategory[issue.Category] = append(issuesByCategory[issue.Category], issue)
	}

	logger.SimpleInfo(fmt.Sprintf("🔍 Found %d quality issues in %d categories",
		len(report.Issues), len(issuesByCategory)))

	// Display issues by category
	for category, issues := range issuesByCategory {
		logger.SimpleWarn(fmt.Sprintf("📋 %s (%d issues):", category, len(issues)))

		for _, issue := range issues {
			severityIcon := "🟡"
			switch issue.Severity {
			case "high":
				severityIcon = "🔴"
			case "low":
				severityIcon = "🟢"
			}

			logger.SimpleWarn(fmt.Sprintf("  %s %s:%d - %s",
				severityIcon, issue.File, issue.Line, issue.Message))
		}
	}

	return nil
}
