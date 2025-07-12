package reposync

import (
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

// PythonQualityAnalyzer implements quality analysis for Python projects
type PythonQualityAnalyzer struct {
	logger *zap.Logger
}

// NewPythonQualityAnalyzer creates a new Python quality analyzer
func NewPythonQualityAnalyzer(logger *zap.Logger) *PythonQualityAnalyzer {
	return &PythonQualityAnalyzer{logger: logger}
}

func (p *PythonQualityAnalyzer) Name() string     { return "pylint" }
func (p *PythonQualityAnalyzer) Language() string { return "python" }

func (p *PythonQualityAnalyzer) IsAvailable(ctx context.Context) bool {
	// Check if pylint is available
	cmd := exec.CommandContext(ctx, "pylint", "--version")
	err := cmd.Run()
	return err == nil
}

func (p *PythonQualityAnalyzer) Analyze(ctx context.Context, path string) (*QualityResult, error) {
	result := &QualityResult{
		Repository: path,
		Issues:     make([]QualityIssue, 0),
		Metrics:    QualityMetrics{},
	}

	// Run pylint
	pylintIssues, pylintScore := p.runPylint(ctx, path)
	if pylintIssues != nil {
		result.Issues = append(result.Issues, pylintIssues...)
	}

	// Run flake8 for additional style checks
	flake8Issues, err := p.runFlake8(ctx, path)
	if err != nil {
		p.logger.Warn("Failed to run flake8", zap.Error(err))
	} else {
		result.Issues = append(result.Issues, flake8Issues...)
	}

	// Run bandit for security analysis
	securityIssues, err := p.runBandit(ctx, path)
	if err != nil {
		p.logger.Warn("Failed to run bandit", zap.Error(err))
	} else {
		result.Issues = append(result.Issues, securityIssues...)
	}

	// Run mypy for type checking
	typeIssues, err := p.runMypy(ctx, path)
	if err != nil {
		p.logger.Warn("Failed to run mypy", zap.Error(err))
	} else {
		result.Issues = append(result.Issues, typeIssues...)
	}

	// Calculate complexity using radon
	avgComplexity, err := p.calculateComplexity(ctx, path)
	if err != nil {
		p.logger.Warn("Failed to calculate complexity", zap.Error(err))
		avgComplexity = 5.0 // Default
	}
	result.Metrics.AvgComplexity = avgComplexity

	// Count files and lines
	fileMetrics, err := p.countPythonFiles(path)
	if err != nil {
		p.logger.Warn("Failed to count Python files", zap.Error(err))
	} else {
		result.Metrics.TotalFiles = fileMetrics.totalFiles
		result.Metrics.TotalLinesOfCode = fileMetrics.totalLines
	}

	// Get test coverage
	coverage, err := p.getTestCoverage(ctx, path)
	if err != nil {
		p.logger.Warn("Failed to get test coverage", zap.Error(err))
	} else {
		result.Metrics.TestCoverage = coverage
	}

	// Calculate duplication
	duplicationRate, err := p.calculateDuplication(ctx, path)
	if err != nil {
		p.logger.Warn("Failed to calculate duplication", zap.Error(err))
	} else {
		result.Metrics.DuplicationRate = duplicationRate
	}

	// Use pylint score if available, otherwise calculate
	if pylintScore > 0 {
		result.OverallScore = pylintScore
	} else {
		result.OverallScore = p.calculateScore(result)
	}

	return result, nil
}

func (p *PythonQualityAnalyzer) runPylint(ctx context.Context, path string) ([]QualityIssue, float64) {
	// Create a basic pylintrc if none exists
	pylintrc := p.findPylintConfig(path)
	if pylintrc == "" {
		if err := p.createBasicPylintrc(path); err != nil {
			p.logger.Warn("Failed to create pylintrc", zap.Error(err))
		}
		defer os.Remove(filepath.Join(path, ".pylintrc"))
	}

	// Run pylint with JSON output
	cmd := exec.CommandContext(ctx, "pylint", "--output-format=json", "--recursive=y", path)
	output, err := cmd.Output()

	var score float64
	issues := make([]QualityIssue, 0)

	if err != nil {
		// Try to get the score from stderr
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderrStr := string(exitErr.Stderr)
			if scoreMatch := regexp.MustCompile(`Your code has been rated at ([\d.]+)/10`).FindStringSubmatch(stderrStr); len(scoreMatch) > 1 {
				if s, err := strconv.ParseFloat(scoreMatch[1], 64); err == nil {
					score = s * 10 // Convert to percentage
				}
			}
		}
	}

	// Parse JSON output
	var pylintResults []struct {
		Type      string `json:"type"`
		Module    string `json:"module"`
		Object    string `json:"obj"`
		Line      int    `json:"line"`
		Column    int    `json:"column"`
		Path      string `json:"path"`
		Symbol    string `json:"symbol"`
		Message   string `json:"message"`
		MessageID string `json:"message-id"`
	}

	if err := json.Unmarshal(output, &pylintResults); err == nil {
		for _, msg := range pylintResults {
			issues = append(issues, QualityIssue{
				Type:     p.mapPylintType(msg.Type),
				Severity: p.mapPylintSeverity(msg.Type),
				File:     msg.Path,
				Line:     msg.Line,
				Column:   msg.Column,
				Message:  msg.Message,
				Rule:     msg.MessageID,
				Tool:     "pylint",
			})
		}
	} else {
		// Fallback to parsing text output
		issues = p.parsePylintTextOutput(string(output))
	}

	return issues, score
}

func (p *PythonQualityAnalyzer) runFlake8(ctx context.Context, path string) ([]QualityIssue, error) {
	cmd := exec.CommandContext(ctx, "flake8", "--format=%(path)s:%(row)d:%(col)d: %(code)s %(text)s", path)
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

		// Parse flake8 output: file.py:line:col: CODE message
		re := regexp.MustCompile(`^(.+):(\d+):(\d+): ([A-Z]\d+) (.+)$`)
		matches := re.FindStringSubmatch(line)
		if len(matches) == 6 {
			lineNum, _ := strconv.Atoi(matches[2])
			colNum, _ := strconv.Atoi(matches[3])

			severity := "minor"
			if strings.HasPrefix(matches[4], "E") {
				severity = "major"
			} else if strings.HasPrefix(matches[4], "F") {
				severity = "critical"
			}

			issues = append(issues, QualityIssue{
				Type:     "style",
				Severity: severity,
				File:     matches[1],
				Line:     lineNum,
				Column:   colNum,
				Message:  matches[5],
				Rule:     matches[4],
				Tool:     "flake8",
			})
		}
	}

	return issues, nil
}

func (p *PythonQualityAnalyzer) runBandit(ctx context.Context, path string) ([]QualityIssue, error) {
	cmd := exec.CommandContext(ctx, "bandit", "-r", "-f", "json", path)
	output, err := cmd.Output()
	if err != nil && len(output) == 0 {
		return nil, err
	}

	// Parse Bandit JSON output
	var banditResult struct {
		Results []struct {
			Filename   string `json:"filename"`
			LineNumber int    `json:"line_number"`
			LineRange  []int  `json:"line_range"`
			TestID     string `json:"test_id"`
			TestName   string `json:"test_name"`
			Severity   string `json:"issue_severity"`
			Confidence string `json:"issue_confidence"`
			IssueText  string `json:"issue_text"`
		} `json:"results"`
	}

	if err := json.Unmarshal(output, &banditResult); err != nil {
		return nil, err
	}

	issues := make([]QualityIssue, 0)
	for _, result := range banditResult.Results {
		issues = append(issues, QualityIssue{
			Type:     "security",
			Severity: p.mapBanditSeverity(result.Severity),
			File:     result.Filename,
			Line:     result.LineNumber,
			Column:   0,
			Message:  result.IssueText,
			Rule:     result.TestID,
			Tool:     "bandit",
		})
	}

	return issues, nil
}

func (p *PythonQualityAnalyzer) runMypy(ctx context.Context, path string) ([]QualityIssue, error) {
	cmd := exec.CommandContext(ctx, "mypy", "--no-error-summary", path)
	output, err := cmd.CombinedOutput()
	if err != nil && len(output) == 0 {
		return nil, err
	}

	issues := make([]QualityIssue, 0)
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		if line == "" || strings.HasPrefix(line, "Success:") {
			continue
		}

		// Parse mypy output: file.py:line: error: message
		re := regexp.MustCompile(`^(.+):(\d+): (error|warning|note): (.+)$`)
		matches := re.FindStringSubmatch(line)
		if len(matches) == 5 {
			lineNum, _ := strconv.Atoi(matches[2])

			severity := "major"
			if matches[3] == "warning" {
				severity = "minor"
			} else if matches[3] == "note" {
				severity = "info"
			}

			issues = append(issues, QualityIssue{
				Type:     "type-error",
				Severity: severity,
				File:     matches[1],
				Line:     lineNum,
				Column:   0,
				Message:  matches[4],
				Rule:     "mypy",
				Tool:     "mypy",
			})
		}
	}

	return issues, nil
}

func (p *PythonQualityAnalyzer) calculateComplexity(ctx context.Context, path string) (float64, error) {
	// Use radon for complexity calculation
	cmd := exec.CommandContext(ctx, "radon", "cc", "-a", "-j", path)
	output, err := cmd.Output()
	if err != nil {
		return p.estimateComplexity(path), nil
	}

	// Parse radon JSON output
	var radonResult map[string][]struct {
		Complexity int `json:"complexity"`
	}

	if err := json.Unmarshal(output, &radonResult); err != nil {
		// Try parsing average from text output
		cmd = exec.CommandContext(ctx, "radon", "cc", "-a", path)
		output, err = cmd.Output()
		if err != nil {
			return 5.0, err
		}

		// Look for average complexity in output
		if match := regexp.MustCompile(`Average complexity: ([A-Z]) \(([\d.]+)\)`).FindStringSubmatch(string(output)); len(match) > 2 {
			if avg, err := strconv.ParseFloat(match[2], 64); err == nil {
				return avg, nil
			}
		}
		return 5.0, nil
	}

	// Calculate average from JSON
	total := 0
	count := 0
	for _, functions := range radonResult {
		for _, fn := range functions {
			total += fn.Complexity
			count++
		}
	}

	if count > 0 {
		return float64(total) / float64(count), nil
	}
	return 5.0, nil
}

func (p *PythonQualityAnalyzer) countPythonFiles(path string) (*fileMetrics, error) {
	metrics := &fileMetrics{}

	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip virtual environments and cache directories
		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") || name == "__pycache__" || name == "venv" ||
				name == "env" || name == ".venv" || name == "site-packages" {
				return filepath.SkipDir
			}
		}

		if strings.HasSuffix(filePath, ".py") && !strings.Contains(filePath, "test_") &&
			!strings.Contains(filePath, "_test.py") {
			metrics.totalFiles++

			// Count lines
			content, err := os.ReadFile(filePath)
			if err == nil {
				lines := strings.Split(string(content), "\n")
				for _, line := range lines {
					trimmed := strings.TrimSpace(line)
					if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
						metrics.totalLines++
					}
				}
			}
		}

		return nil
	})

	return metrics, err
}

func (p *PythonQualityAnalyzer) getTestCoverage(ctx context.Context, path string) (float64, error) {
	// Try pytest with coverage
	cmd := exec.CommandContext(ctx, "pytest", "--cov=.", "--cov-report=term", "--quiet")
	cmd.Dir = path
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Try coverage.py directly
		cmd = exec.CommandContext(ctx, "coverage", "run", "-m", "pytest", "--quiet")
		cmd.Dir = path
		if err := cmd.Run(); err != nil {
			return 0, err
		}

		cmd = exec.CommandContext(ctx, "coverage", "report")
		cmd.Dir = path
		output, err = cmd.Output()
		if err != nil {
			return 0, err
		}
	}

	// Parse coverage output
	outputStr := string(output)
	if match := regexp.MustCompile(`TOTAL\s+\d+\s+\d+\s+(\d+)%`).FindStringSubmatch(outputStr); len(match) > 1 {
		if coverage, err := strconv.ParseFloat(match[1], 64); err == nil {
			return coverage, nil
		}
	}

	return 0, fmt.Errorf("could not parse coverage output")
}

func (p *PythonQualityAnalyzer) calculateDuplication(ctx context.Context, path string) (float64, error) {
	// Try to use pylint's duplicate-code checker
	cmd := exec.CommandContext(ctx, "pylint", "--disable=all", "--enable=duplicate-code", "--output-format=json", path)
	output, err := cmd.Output()
	if err == nil {
		// Count duplicate-code messages
		var pylintResults []struct {
			Symbol string `json:"symbol"`
		}

		if err := json.Unmarshal(output, &pylintResults); err == nil {
			duplicateCount := 0
			for _, msg := range pylintResults {
				if msg.Symbol == "duplicate-code" {
					duplicateCount++
				}
			}

			// Rough estimation: each duplicate warning represents ~5% duplication
			return float64(duplicateCount) * 5.0, nil
		}
	}

	// Fallback to simple detection
	return p.simpleDuplicationDetection(path), nil
}

func (p *PythonQualityAnalyzer) simpleDuplicationDetection(path string) float64 {
	lineMap := make(map[string]int)
	totalLines := 0

	filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil || !strings.HasSuffix(filePath, ".py") {
			return nil
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil
		}

		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if len(trimmed) > 30 && !strings.HasPrefix(trimmed, "#") &&
				!strings.HasPrefix(trimmed, "import") && !strings.HasPrefix(trimmed, "from") {
				lineMap[trimmed]++
				totalLines++
			}
		}

		return nil
	})

	duplicateLines := 0
	for _, count := range lineMap {
		if count > 1 {
			duplicateLines += count - 1
		}
	}

	if totalLines == 0 {
		return 0
	}

	return float64(duplicateLines) / float64(totalLines) * 100
}

func (p *PythonQualityAnalyzer) calculateScore(result *QualityResult) float64 {
	score := 100.0

	// Deduct points for issues
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

func (p *PythonQualityAnalyzer) mapPylintType(msgType string) string {
	switch msgType {
	case "convention", "refactor":
		return "style"
	case "warning":
		return "complexity"
	case "error", "fatal":
		return "bug"
	default:
		return "info"
	}
}

func (p *PythonQualityAnalyzer) mapPylintSeverity(msgType string) string {
	switch msgType {
	case "fatal":
		return "critical"
	case "error":
		return "major"
	case "warning":
		return "minor"
	default:
		return "info"
	}
}

func (p *PythonQualityAnalyzer) mapBanditSeverity(severity string) string {
	switch strings.ToUpper(severity) {
	case "HIGH":
		return "critical"
	case "MEDIUM":
		return "major"
	case "LOW":
		return "minor"
	default:
		return "info"
	}
}

func (p *PythonQualityAnalyzer) parsePylintTextOutput(output string) []QualityIssue {
	issues := make([]QualityIssue, 0)
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if line == "" || strings.HasPrefix(line, "***") || strings.HasPrefix(line, "---") {
			continue
		}

		// Parse format: file.py:line:column: CODE: message (name)
		re := regexp.MustCompile(`^(.+):(\d+):(\d+): ([A-Z]\d+): (.+) \((.+)\)$`)
		matches := re.FindStringSubmatch(line)
		if len(matches) == 7 {
			lineNum, _ := strconv.Atoi(matches[2])
			colNum, _ := strconv.Atoi(matches[3])

			issues = append(issues, QualityIssue{
				Type:     "style",
				Severity: "minor",
				File:     matches[1],
				Line:     lineNum,
				Column:   colNum,
				Message:  matches[5],
				Rule:     matches[4],
				Tool:     "pylint",
			})
		}
	}

	return issues
}

func (p *PythonQualityAnalyzer) estimateComplexity(path string) float64 {
	totalComplexity := 0
	functionCount := 0

	filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil || !strings.HasSuffix(filePath, ".py") {
			return nil
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil
		}

		code := string(content)
		// Count control flow statements
		totalComplexity += strings.Count(code, " if ")
		totalComplexity += strings.Count(code, " elif ")
		totalComplexity += strings.Count(code, " for ")
		totalComplexity += strings.Count(code, " while ")
		totalComplexity += strings.Count(code, " except ")

		// Count functions and methods
		functionCount += strings.Count(code, "def ")

		return nil
	})

	if functionCount > 0 {
		return float64(totalComplexity) / float64(functionCount)
	}
	return 5.0
}

func (p *PythonQualityAnalyzer) findPylintConfig(path string) string {
	configFiles := []string{
		".pylintrc",
		"pylintrc",
		"pyproject.toml",
		"setup.cfg",
		"tox.ini",
	}

	for _, configFile := range configFiles {
		configPath := filepath.Join(path, configFile)
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
	}

	return ""
}

func (p *PythonQualityAnalyzer) createBasicPylintrc(path string) error {
	config := `[MESSAGES CONTROL]
disable=missing-docstring,too-few-public-methods,too-many-arguments

[FORMAT]
max-line-length=120

[DESIGN]
max-complexity=10
`
	return os.WriteFile(filepath.Join(path, ".pylintrc"), []byte(config), 0644)
}
