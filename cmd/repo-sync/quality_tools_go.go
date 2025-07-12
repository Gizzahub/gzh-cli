package reposync

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

// GoQualityAnalyzer implements quality analysis for Go projects
type GoQualityAnalyzer struct {
	logger *zap.Logger
}

// NewGoQualityAnalyzer creates a new Go quality analyzer
func NewGoQualityAnalyzer(logger *zap.Logger) *GoQualityAnalyzer {
	return &GoQualityAnalyzer{logger: logger}
}

func (g *GoQualityAnalyzer) Name() string     { return "golangci-lint" }
func (g *GoQualityAnalyzer) Language() string { return "go" }

func (g *GoQualityAnalyzer) IsAvailable(ctx context.Context) bool {
	cmd := exec.CommandContext(ctx, "golangci-lint", "--version")
	err := cmd.Run()
	return err == nil
}

func (g *GoQualityAnalyzer) Analyze(ctx context.Context, path string) (*QualityResult, error) {
	result := &QualityResult{
		Repository: path,
		Issues:     make([]QualityIssue, 0),
		Metrics:    QualityMetrics{},
	}

	// Run golangci-lint
	lintIssues, err := g.runGolangciLint(ctx, path)
	if err != nil {
		g.logger.Warn("Failed to run golangci-lint", zap.Error(err))
	} else {
		result.Issues = append(result.Issues, lintIssues...)
	}

	// Run go vet
	vetIssues, err := g.runGoVet(ctx, path)
	if err != nil {
		g.logger.Warn("Failed to run go vet", zap.Error(err))
	} else {
		result.Issues = append(result.Issues, vetIssues...)
	}

	// Calculate cyclomatic complexity
	complexityMetrics, err := g.calculateComplexity(ctx, path)
	if err != nil {
		g.logger.Warn("Failed to calculate complexity", zap.Error(err))
	} else {
		result.Metrics.AvgComplexity = complexityMetrics.avgComplexity
	}

	// Count Go files and lines
	fileMetrics, err := g.countGoFiles(path)
	if err != nil {
		g.logger.Warn("Failed to count Go files", zap.Error(err))
	} else {
		result.Metrics.TotalFiles = fileMetrics.totalFiles
		result.Metrics.TotalLinesOfCode = fileMetrics.totalLines
	}

	// Get test coverage
	coverage, err := g.getTestCoverage(ctx, path)
	if err != nil {
		g.logger.Warn("Failed to get test coverage", zap.Error(err))
	} else {
		result.Metrics.TestCoverage = coverage
	}

	// Calculate duplication rate
	duplicationRate, err := g.calculateDuplication(ctx, path)
	if err != nil {
		g.logger.Warn("Failed to calculate duplication", zap.Error(err))
	} else {
		result.Metrics.DuplicationRate = duplicationRate
	}

	// Calculate overall score
	result.OverallScore = g.calculateScore(result)

	return result, nil
}

func (g *GoQualityAnalyzer) runGolangciLint(ctx context.Context, path string) ([]QualityIssue, error) {
	// Run golangci-lint with JSON output
	cmd := exec.CommandContext(ctx, "golangci-lint", "run", "--out-format", "json", path)
	output, err := cmd.Output()
	if err != nil {
		// golangci-lint returns non-zero exit code when issues are found
		if exitErr, ok := err.(*exec.ExitError); ok {
			output = exitErr.Stderr
		} else {
			return nil, err
		}
	}

	// Parse JSON output
	var lintResult struct {
		Issues []struct {
			FromLinter string `json:"FromLinter"`
			Text       string `json:"Text"`
			Severity   string `json:"Severity"`
			Pos        struct {
				Filename string `json:"Filename"`
				Line     int    `json:"Line"`
				Column   int    `json:"Column"`
			} `json:"Pos"`
		} `json:"Issues"`
	}

	if err := json.Unmarshal(output, &lintResult); err != nil {
		// Fallback to parsing text output
		return g.parseLintTextOutput(string(output)), nil
	}

	issues := make([]QualityIssue, 0, len(lintResult.Issues))
	for _, issue := range lintResult.Issues {
		severity := g.mapSeverity(issue.Severity)
		issues = append(issues, QualityIssue{
			Type:     "style",
			Severity: severity,
			File:     issue.Pos.Filename,
			Line:     issue.Pos.Line,
			Column:   issue.Pos.Column,
			Message:  issue.Text,
			Rule:     issue.FromLinter,
			Tool:     "golangci-lint",
		})
	}

	return issues, nil
}

func (g *GoQualityAnalyzer) runGoVet(ctx context.Context, path string) ([]QualityIssue, error) {
	cmd := exec.CommandContext(ctx, "go", "vet", "./...")
	cmd.Dir = path
	output, err := cmd.CombinedOutput()
	if err != nil && len(output) == 0 {
		return nil, err
	}

	issues := make([]QualityIssue, 0)
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		// Parse go vet output: file.go:line:column: message
		parts := strings.SplitN(line, ":", 4)
		if len(parts) >= 3 {
			lineNum, _ := strconv.Atoi(parts[1])
			colNum := 0
			message := parts[2]

			if len(parts) == 4 {
				colNum, _ = strconv.Atoi(parts[2])
				message = parts[3]
			}

			issues = append(issues, QualityIssue{
				Type:     "bug",
				Severity: "major",
				File:     parts[0],
				Line:     lineNum,
				Column:   colNum,
				Message:  strings.TrimSpace(message),
				Rule:     "go-vet",
				Tool:     "go vet",
			})
		}
	}

	return issues, nil
}

func (g *GoQualityAnalyzer) calculateComplexity(ctx context.Context, path string) (*complexityMetrics, error) {
	// Use gocyclo if available, otherwise estimate from code structure
	cmd := exec.CommandContext(ctx, "gocyclo", "-avg", path)
	output, err := cmd.Output()
	if err != nil {
		// Fallback to simple estimation
		return g.estimateComplexity(path), nil
	}

	// Parse gocyclo output
	avgStr := strings.TrimSpace(string(output))
	avg, err := strconv.ParseFloat(avgStr, 64)
	if err != nil {
		return &complexityMetrics{avgComplexity: 5.0}, nil
	}

	return &complexityMetrics{avgComplexity: avg}, nil
}

func (g *GoQualityAnalyzer) countGoFiles(path string) (*fileMetrics, error) {
	metrics := &fileMetrics{}

	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip vendor and hidden directories
		if info.IsDir() && (strings.HasPrefix(info.Name(), ".") || info.Name() == "vendor") {
			return filepath.SkipDir
		}

		if strings.HasSuffix(filePath, ".go") && !strings.HasSuffix(filePath, "_test.go") {
			metrics.totalFiles++

			// Count lines
			lines, err := g.countLines(filePath)
			if err == nil {
				metrics.totalLines += lines
			}
		}

		return nil
	})

	return metrics, err
}

func (g *GoQualityAnalyzer) countLines(filePath string) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lines := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip empty lines and comments
		if line != "" && !strings.HasPrefix(line, "//") {
			lines++
		}
	}

	return lines, scanner.Err()
}

func (g *GoQualityAnalyzer) getTestCoverage(ctx context.Context, path string) (float64, error) {
	// Run go test with coverage
	cmd := exec.CommandContext(ctx, "go", "test", "-cover", "./...")
	cmd.Dir = path
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	// Parse coverage output
	re := regexp.MustCompile(`coverage: (\d+\.\d+)% of statements`)
	matches := re.FindStringSubmatch(string(output))
	if len(matches) < 2 {
		return 0, fmt.Errorf("could not parse coverage output")
	}

	coverage, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0, err
	}

	return coverage, nil
}

func (g *GoQualityAnalyzer) calculateDuplication(ctx context.Context, path string) (float64, error) {
	// Simple duplication detection based on repeated code blocks
	// In a real implementation, you might use a tool like jscpd
	duplicateLines := 0
	totalLines := 0

	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil || !strings.HasSuffix(filePath, ".go") {
			return nil
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil
		}

		lines := strings.Split(string(content), "\n")
		totalLines += len(lines)

		// Simple duplicate detection: look for exact duplicate lines
		lineMap := make(map[string]int)
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" && !strings.HasPrefix(trimmed, "//") && len(trimmed) > 20 {
				lineMap[trimmed]++
			}
		}

		for _, count := range lineMap {
			if count > 1 {
				duplicateLines += count - 1
			}
		}

		return nil
	})

	if err != nil || totalLines == 0 {
		return 0, err
	}

	return float64(duplicateLines) / float64(totalLines) * 100, nil
}

func (g *GoQualityAnalyzer) calculateScore(result *QualityResult) float64 {
	score := 100.0

	// Deduct points for issues based on severity
	for _, issue := range result.Issues {
		switch issue.Severity {
		case "critical":
			score -= 5.0
		case "major":
			score -= 3.0
		case "minor":
			score -= 1.0
		case "info":
			score -= 0.5
		}
	}

	// Factor in complexity
	if result.Metrics.AvgComplexity > 10 {
		score -= (result.Metrics.AvgComplexity - 10) * 2
	}

	// Factor in test coverage
	if result.Metrics.TestCoverage < 80 {
		score -= (80 - result.Metrics.TestCoverage) * 0.5
	}

	// Factor in duplication
	if result.Metrics.DuplicationRate > 5 {
		score -= result.Metrics.DuplicationRate * 0.5
	}

	// Ensure score is between 0 and 100
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}

func (g *GoQualityAnalyzer) mapSeverity(severity string) string {
	switch strings.ToLower(severity) {
	case "error":
		return "critical"
	case "warning":
		return "major"
	case "info":
		return "info"
	default:
		return "minor"
	}
}

func (g *GoQualityAnalyzer) parseLintTextOutput(output string) []QualityIssue {
	issues := make([]QualityIssue, 0)
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		// Parse format: file.go:line:column: message (linter)
		re := regexp.MustCompile(`^(.+):(\d+):(\d+): (.+) \((.+)\)$`)
		matches := re.FindStringSubmatch(line)
		if len(matches) == 6 {
			lineNum, _ := strconv.Atoi(matches[2])
			colNum, _ := strconv.Atoi(matches[3])

			issues = append(issues, QualityIssue{
				Type:     "style",
				Severity: "minor",
				File:     matches[1],
				Line:     lineNum,
				Column:   colNum,
				Message:  matches[4],
				Rule:     matches[5],
				Tool:     "golangci-lint",
			})
		}
	}

	return issues
}

func (g *GoQualityAnalyzer) estimateComplexity(path string) *complexityMetrics {
	totalComplexity := 0
	functionCount := 0

	filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil || !strings.HasSuffix(filePath, ".go") {
			return nil
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil
		}

		// Count control flow statements as a simple complexity metric
		code := string(content)
		totalComplexity += strings.Count(code, " if ")
		totalComplexity += strings.Count(code, " for ")
		totalComplexity += strings.Count(code, " switch ")
		totalComplexity += strings.Count(code, " case ")
		totalComplexity += strings.Count(code, " select ")

		// Count functions
		functionCount += strings.Count(code, "func ")

		return nil
	})

	avgComplexity := 1.0
	if functionCount > 0 {
		avgComplexity = float64(totalComplexity) / float64(functionCount)
	}

	return &complexityMetrics{avgComplexity: avgComplexity}
}

type complexityMetrics struct {
	avgComplexity float64
}

type fileMetrics struct {
	totalFiles int
	totalLines int
}
