package reposync

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

// JavaScriptQualityAnalyzer implements quality analysis for JavaScript projects
type JavaScriptQualityAnalyzer struct {
	logger *zap.Logger
}

// NewJavaScriptQualityAnalyzer creates a new JavaScript quality analyzer
func NewJavaScriptQualityAnalyzer(logger *zap.Logger) *JavaScriptQualityAnalyzer {
	return &JavaScriptQualityAnalyzer{logger: logger}
}

func (j *JavaScriptQualityAnalyzer) Name() string     { return "eslint" }
func (j *JavaScriptQualityAnalyzer) Language() string { return "javascript" }

func (j *JavaScriptQualityAnalyzer) IsAvailable(ctx context.Context) bool {
	// Check if eslint is available
	cmd := exec.CommandContext(ctx, "npx", "eslint", "--version")
	err := cmd.Run()
	return err == nil
}

func (j *JavaScriptQualityAnalyzer) Analyze(ctx context.Context, path string) (*QualityResult, error) {
	result := &QualityResult{
		Repository: path,
		Issues:     make([]QualityIssue, 0),
		Metrics:    QualityMetrics{},
	}

	// Run ESLint
	lintIssues, err := j.runESLint(ctx, path)
	if err != nil {
		j.logger.Warn("Failed to run ESLint", zap.Error(err))
	} else {
		result.Issues = append(result.Issues, lintIssues...)
	}

	// Calculate complexity using eslint complexity rule
	complexityIssues, avgComplexity := j.analyzeComplexity(result.Issues)
	result.Metrics.AvgComplexity = avgComplexity

	// Count files and lines
	fileMetrics, err := j.countJSFiles(path)
	if err != nil {
		j.logger.Warn("Failed to count JS files", zap.Error(err))
	} else {
		result.Metrics.TotalFiles = fileMetrics.totalFiles
		result.Metrics.TotalLinesOfCode = fileMetrics.totalLines
	}

	// Get test coverage if jest or mocha is available
	coverage, err := j.getTestCoverage(ctx, path)
	if err != nil {
		j.logger.Warn("Failed to get test coverage", zap.Error(err))
	} else {
		result.Metrics.TestCoverage = coverage
	}

	// Calculate duplication using jscpd if available
	duplicationRate, err := j.calculateDuplication(ctx, path)
	if err != nil {
		j.logger.Warn("Failed to calculate duplication", zap.Error(err))
	} else {
		result.Metrics.DuplicationRate = duplicationRate
	}

	// Calculate overall score
	result.OverallScore = j.calculateScore(result)

	return result, nil
}

func (j *JavaScriptQualityAnalyzer) runESLint(ctx context.Context, path string) ([]QualityIssue, error) {
	// Check if .eslintrc exists
	eslintConfig := j.findESLintConfig(path)
	if eslintConfig == "" {
		// Create a basic ESLint config if none exists
		if err := j.createBasicESLintConfig(path); err != nil {
			return nil, fmt.Errorf("failed to create ESLint config: %w", err)
		}
		defer os.Remove(filepath.Join(path, ".eslintrc.json"))
	}

	// Run ESLint with JSON output
	cmd := exec.CommandContext(ctx, "npx", "eslint", ".", "--format", "json", "--ext", ".js,.jsx,.ts,.tsx")
	cmd.Dir = path
	output, err := cmd.Output()
	if err != nil {
		// ESLint returns non-zero exit code when issues are found
		if exitErr, ok := err.(*exec.ExitError); ok && len(exitErr.Stderr) == 0 {
			// This is expected when ESLint finds issues
		} else {
			return nil, err
		}
	}

	// Parse JSON output
	var eslintResults []struct {
		FilePath string `json:"filePath"`
		Messages []struct {
			RuleID   string `json:"ruleId"`
			Severity int    `json:"severity"`
			Message  string `json:"message"`
			Line     int    `json:"line"`
			Column   int    `json:"column"`
			NodeType string `json:"nodeType"`
		} `json:"messages"`
	}

	if err := json.Unmarshal(output, &eslintResults); err != nil {
		return nil, fmt.Errorf("failed to parse ESLint output: %w", err)
	}

	issues := make([]QualityIssue, 0)
	for _, file := range eslintResults {
		for _, msg := range file.Messages {
			issueType := "style"
			if strings.Contains(msg.RuleID, "error") || msg.Severity == 2 {
				issueType = "bug"
			}

			issues = append(issues, QualityIssue{
				Type:       issueType,
				Severity:   j.mapESLintSeverity(msg.Severity),
				File:       file.FilePath,
				Line:       msg.Line,
				Column:     msg.Column,
				Message:    msg.Message,
				Rule:       msg.RuleID,
				Tool:       "eslint",
				Suggestion: j.getESLintSuggestion(msg.RuleID),
			})
		}
	}

	return issues, nil
}

func (j *JavaScriptQualityAnalyzer) analyzeComplexity(issues []QualityIssue) ([]QualityIssue, float64) {
	complexityIssues := make([]QualityIssue, 0)
	totalComplexity := 0
	complexityCount := 0

	for _, issue := range issues {
		if strings.Contains(issue.Rule, "complexity") {
			complexityIssues = append(complexityIssues, issue)
			// Extract complexity value from message if possible
			if strings.Contains(issue.Message, "Cyclomatic complexity") {
				// Parse complexity value
				complexityCount++
				totalComplexity += 10 // Default high complexity
			}
		}
	}

	avgComplexity := 5.0 // Default
	if complexityCount > 0 {
		avgComplexity = float64(totalComplexity) / float64(complexityCount)
	}

	return complexityIssues, avgComplexity
}

func (j *JavaScriptQualityAnalyzer) countJSFiles(path string) (*fileMetrics, error) {
	metrics := &fileMetrics{}
	extensions := map[string]bool{
		".js":  true,
		".jsx": true,
		".ts":  true,
		".tsx": true,
		".mjs": true,
		".cjs": true,
	}

	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip node_modules and hidden directories
		if info.IsDir() && (strings.HasPrefix(info.Name(), ".") || info.Name() == "node_modules" || info.Name() == "dist" || info.Name() == "build") {
			return filepath.SkipDir
		}

		ext := filepath.Ext(filePath)
		if extensions[ext] && !strings.Contains(filePath, ".test.") && !strings.Contains(filePath, ".spec.") {
			metrics.totalFiles++

			// Count lines
			content, err := os.ReadFile(filePath)
			if err == nil {
				lines := strings.Split(string(content), "\n")
				for _, line := range lines {
					trimmed := strings.TrimSpace(line)
					if trimmed != "" && !strings.HasPrefix(trimmed, "//") && !strings.HasPrefix(trimmed, "/*") {
						metrics.totalLines++
					}
				}
			}
		}

		return nil
	})

	return metrics, err
}

func (j *JavaScriptQualityAnalyzer) getTestCoverage(ctx context.Context, path string) (float64, error) {
	// Check package.json for test script
	packageJSON := filepath.Join(path, "package.json")
	if _, err := os.Stat(packageJSON); err != nil {
		return 0, fmt.Errorf("no package.json found")
	}

	// Try to run coverage command
	cmd := exec.CommandContext(ctx, "npm", "run", "test:coverage", "--", "--silent")
	cmd.Dir = path
	output, err := cmd.Output()
	if err != nil {
		// Try with jest directly
		cmd = exec.CommandContext(ctx, "npx", "jest", "--coverage", "--silent")
		cmd.Dir = path
		output, err = cmd.Output()
		if err != nil {
			return 0, err
		}
	}

	// Parse coverage output (looking for summary)
	outputStr := string(output)
	if strings.Contains(outputStr, "All files") {
		// Simple parsing of coverage percentage
		lines := strings.Split(outputStr, "\n")
		for _, line := range lines {
			if strings.Contains(line, "All files") {
				// Extract percentage from line
				parts := strings.Fields(line)
				for i, part := range parts {
					if strings.HasSuffix(part, "%") && i > 0 {
						coverage := strings.TrimSuffix(part, "%")
						if val, err := parseFloat(coverage); err == nil {
							return val, nil
						}
					}
				}
			}
		}
	}

	return 0, fmt.Errorf("could not parse coverage output")
}

func (j *JavaScriptQualityAnalyzer) calculateDuplication(ctx context.Context, path string) (float64, error) {
	// Try to use jscpd if available
	cmd := exec.CommandContext(ctx, "npx", "jscpd", path, "--reporters", "json", "--silent")
	output, err := cmd.Output()
	if err != nil {
		// Fallback to simple duplication detection
		return j.simpleDuplicationDetection(path), nil
	}

	// Parse jscpd JSON output
	var result struct {
		Statistics struct {
			Percentage float64 `json:"percentage"`
		} `json:"statistics"`
	}

	if err := json.Unmarshal(output, &result); err != nil {
		return 0, err
	}

	return result.Statistics.Percentage, nil
}

func (j *JavaScriptQualityAnalyzer) simpleDuplicationDetection(path string) float64 {
	// Simple line-based duplication detection
	lineMap := make(map[string]int)
	totalLines := 0

	filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		ext := filepath.Ext(filePath)
		if ext != ".js" && ext != ".jsx" && ext != ".ts" && ext != ".tsx" {
			return nil
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil
		}

		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if len(trimmed) > 30 && !strings.HasPrefix(trimmed, "//") && !strings.HasPrefix(trimmed, "import") && !strings.HasPrefix(trimmed, "export") {
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

func (j *JavaScriptQualityAnalyzer) calculateScore(result *QualityResult) float64 {
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
		score -= (80 - result.Metrics.TestCoverage) * 0.3
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

func (j *JavaScriptQualityAnalyzer) mapESLintSeverity(severity int) string {
	switch severity {
	case 2:
		return "major"
	case 1:
		return "minor"
	default:
		return "info"
	}
}

func (j *JavaScriptQualityAnalyzer) getESLintSuggestion(ruleID string) string {
	suggestions := map[string]string{
		"no-unused-vars":    "Remove unused variable or use it",
		"no-console":        "Remove console statement or use a proper logger",
		"semi":              "Add or remove semicolon as per style guide",
		"quotes":            "Use consistent quote style",
		"indent":            "Fix indentation",
		"no-undef":          "Define the variable or import it",
		"no-empty":          "Add code to empty block or remove it",
		"no-duplicate-case": "Remove duplicate case",
		"complexity":        "Reduce function complexity by extracting logic",
	}

	if suggestion, exists := suggestions[ruleID]; exists {
		return suggestion
	}
	return ""
}

func (j *JavaScriptQualityAnalyzer) findESLintConfig(path string) string {
	configFiles := []string{
		".eslintrc.js",
		".eslintrc.cjs",
		".eslintrc.yaml",
		".eslintrc.yml",
		".eslintrc.json",
		".eslintrc",
		"package.json", // Check for eslintConfig field
	}

	for _, configFile := range configFiles {
		configPath := filepath.Join(path, configFile)
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
	}

	return ""
}

func (j *JavaScriptQualityAnalyzer) createBasicESLintConfig(path string) error {
	config := map[string]interface{}{
		"env": map[string]bool{
			"browser": true,
			"es2021":  true,
			"node":    true,
		},
		"extends": []string{"eslint:recommended"},
		"parserOptions": map[string]interface{}{
			"ecmaVersion": "latest",
			"sourceType":  "module",
		},
		"rules": map[string]interface{}{
			"complexity": []interface{}{"warn", 10},
		},
	}

	configJSON, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(path, ".eslintrc.json"), configJSON, 0o644)
}

// TypeScriptQualityAnalyzer extends JavaScript analyzer for TypeScript
type TypeScriptQualityAnalyzer struct {
	*JavaScriptQualityAnalyzer
}

// NewTypeScriptQualityAnalyzer creates a new TypeScript quality analyzer
func NewTypeScriptQualityAnalyzer(logger *zap.Logger) *TypeScriptQualityAnalyzer {
	return &TypeScriptQualityAnalyzer{
		JavaScriptQualityAnalyzer: NewJavaScriptQualityAnalyzer(logger),
	}
}

func (t *TypeScriptQualityAnalyzer) Name() string     { return "typescript-eslint" }
func (t *TypeScriptQualityAnalyzer) Language() string { return "typescript" }

func (t *TypeScriptQualityAnalyzer) IsAvailable(ctx context.Context) bool {
	// Check if TypeScript and @typescript-eslint/parser are available
	cmd := exec.CommandContext(ctx, "npx", "tsc", "--version")
	if err := cmd.Run(); err != nil {
		return false
	}

	return t.JavaScriptQualityAnalyzer.IsAvailable(ctx)
}

func (t *TypeScriptQualityAnalyzer) Analyze(ctx context.Context, path string) (*QualityResult, error) {
	// Run TypeScript compiler for type checking
	typeIssues, err := t.runTypeScriptCompiler(ctx, path)
	if err != nil {
		t.logger.Warn("Failed to run TypeScript compiler", zap.Error(err))
	}

	// Run parent JavaScript analysis (ESLint with TypeScript support)
	result, err := t.JavaScriptQualityAnalyzer.Analyze(ctx, path)
	if err != nil {
		return nil, err
	}

	// Add TypeScript-specific issues
	if typeIssues != nil {
		result.Issues = append(result.Issues, typeIssues...)
	}

	return result, nil
}

func (t *TypeScriptQualityAnalyzer) runTypeScriptCompiler(ctx context.Context, path string) ([]QualityIssue, error) {
	// Check if tsconfig.json exists
	tsConfig := filepath.Join(path, "tsconfig.json")
	if _, err := os.Stat(tsConfig); err != nil {
		// Create a basic tsconfig if none exists
		if err := t.createBasicTSConfig(path); err != nil {
			return nil, err
		}
		defer os.Remove(tsConfig)
	}

	// Run tsc with no emit to just check types
	cmd := exec.CommandContext(ctx, "npx", "tsc", "--noEmit", "--pretty", "false")
	cmd.Dir = path
	output, err := cmd.CombinedOutput()
	if err != nil && len(output) == 0 {
		return nil, err
	}

	// Parse TypeScript compiler output
	return t.parseTSCOutput(string(output)), nil
}

func (t *TypeScriptQualityAnalyzer) parseTSCOutput(output string) []QualityIssue {
	issues := make([]QualityIssue, 0)
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		// Parse TypeScript error format: file.ts(line,col): error TS1234: message
		if strings.Contains(line, "): error TS") {
			parts := strings.SplitN(line, ":", 3)
			if len(parts) >= 3 {
				// Extract file and position
				fileAndPos := parts[0]
				if idx := strings.LastIndex(fileAndPos, "("); idx != -1 {
					file := fileAndPos[:idx]
					posStr := fileAndPos[idx+1 : len(fileAndPos)-1]
					positions := strings.Split(posStr, ",")

					lineNum := 0
					colNum := 0
					if len(positions) >= 1 {
						fmt.Sscanf(positions[0], "%d", &lineNum)
					}
					if len(positions) >= 2 {
						fmt.Sscanf(positions[1], "%d", &colNum)
					}

					// Extract error code and message
					message := strings.TrimSpace(parts[2])
					errorCode := ""
					if idx := strings.Index(parts[1], " TS"); idx != -1 {
						errorCode = parts[1][idx+1:]
					}

					issues = append(issues, QualityIssue{
						Type:     "bug",
						Severity: "major",
						File:     file,
						Line:     lineNum,
						Column:   colNum,
						Message:  message,
						Rule:     errorCode,
						Tool:     "tsc",
					})
				}
			}
		}
	}

	return issues
}

func (t *TypeScriptQualityAnalyzer) createBasicTSConfig(path string) error {
	config := map[string]interface{}{
		"compilerOptions": map[string]interface{}{
			"target":                           "es2020",
			"module":                           "commonjs",
			"strict":                           true,
			"esModuleInterop":                  true,
			"skipLibCheck":                     true,
			"forceConsistentCasingInFileNames": true,
		},
	}

	configJSON, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(path, "tsconfig.json"), configJSON, 0o644)
}

// Helper function to parse float
func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}
