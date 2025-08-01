// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package tools

import (
	"encoding/json"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// PrettierTool implements JavaScript/TypeScript formatting using prettier.
type PrettierTool struct {
	*BaseTool
}

// NewPrettierTool creates a new prettier tool.
func NewPrettierTool() *PrettierTool {
	tool := &PrettierTool{
		BaseTool: NewBaseTool("prettier", "JavaScript", "prettier", FORMAT),
	}

	tool.SetInstallCommand([]string{"npm", "install", "-g", "prettier"})
	tool.SetConfigPatterns([]string{
		".prettierrc", ".prettierrc.json", ".prettierrc.js", ".prettierrc.yml", ".prettierrc.yaml",
		"prettier.config.js", "prettier.config.cjs", "package.json",
	})

	return tool
}

// BuildCommand builds the prettier command.
func (t *PrettierTool) BuildCommand(files []string, options ExecuteOptions) *exec.Cmd {
	args := []string{"--write"} // Always write changes

	// Add config file if specified
	if options.ConfigFile != "" {
		args = append(args, "--config", options.ConfigFile)
	}

	// Add extra flags if provided
	args = append(args, options.ExtraArgs...)

	// Filter supported files
	supportedFiles := FilterFilesByExtensions(files, []string{
		".js", ".jsx", ".ts", ".tsx", ".json", ".css", ".scss", ".less",
		".html", ".vue", ".md", ".yaml", ".yml",
	})

	if len(supportedFiles) == 0 {
		// If no specific files, format common patterns
		args = append(args, "**/*.{js,jsx,ts,tsx,json,css,scss,less,html,vue,md,yaml,yml}")
	} else {
		args = append(args, supportedFiles...)
	}

	cmd := exec.Command(t.executable, args...)

	if options.ProjectRoot != "" {
		cmd.Dir = options.ProjectRoot
	}

	return cmd
}

// ESLintTool implements JavaScript/TypeScript linting using eslint.
type ESLintTool struct {
	*BaseTool
}

// NewESLintTool creates a new eslint tool.
func NewESLintTool() *ESLintTool {
	tool := &ESLintTool{
		BaseTool: NewBaseTool("eslint", "JavaScript", "eslint", LINT),
	}

	tool.SetInstallCommand([]string{"npm", "install", "-g", "eslint"})
	tool.SetConfigPatterns([]string{
		".eslintrc", ".eslintrc.json", ".eslintrc.js", ".eslintrc.yml", ".eslintrc.yaml",
		"eslint.config.js", "eslint.config.mjs", "eslint.config.cjs", "package.json",
	})

	return tool
}

// BuildCommand builds the eslint command.
func (t *ESLintTool) BuildCommand(files []string, options ExecuteOptions) *exec.Cmd {
	args := []string{}

	// Add config file if specified
	if options.ConfigFile != "" {
		args = append(args, "--config", options.ConfigFile)
	}

	// Add fix flag if requested
	if options.Fix {
		args = append(args, "--fix")
	}

	// Output format for parsing
	args = append(args, "--format", "json")

	// Add extra flags if provided
	args = append(args, options.ExtraArgs...)

	// Filter JavaScript/TypeScript files
	jsFiles := FilterFilesByExtensions(files, []string{".js", ".jsx", ".ts", ".tsx", ".vue"})
	if len(jsFiles) == 0 {
		// If no specific files, lint common patterns
		args = append(args, "**/*.{js,jsx,ts,tsx,vue}")
	} else {
		args = append(args, jsFiles...)
	}

	cmd := exec.Command(t.executable, args...)

	if options.ProjectRoot != "" {
		cmd.Dir = options.ProjectRoot
	}

	return cmd
}

// ParseOutput parses eslint JSON output.
func (t *ESLintTool) ParseOutput(output string) []Issue {
	if strings.TrimSpace(output) == "" {
		return []Issue{}
	}

	var eslintResults []struct {
		FilePath string `json:"filePath"`
		Messages []struct {
			RuleID   *string `json:"ruleId"`
			Severity int     `json:"severity"`
			Message  string  `json:"message"`
			Line     int     `json:"line"`
			Column   int     `json:"column"`
			NodeType string  `json:"nodeType"`
			Source   string  `json:"source"`
			Fix      *struct {
				Range []int  `json:"range"`
				Text  string `json:"text"`
			} `json:"fix,omitempty"`
		} `json:"messages"`
		ErrorCount   int `json:"errorCount"`
		WarningCount int `json:"warningCount"`
	}

	if err := json.Unmarshal([]byte(output), &eslintResults); err != nil {
		// Fallback to plain text parsing
		return t.parseTextOutput(output)
	}

	var issues []Issue
	for _, file := range eslintResults {
		for _, msg := range file.Messages {
			severity := "info"
			switch msg.Severity {
			case 1:
				severity = "warning"
			case 2:
				severity = "error"
			}

			rule := ""
			if msg.RuleID != nil {
				rule = *msg.RuleID
			}

			issue := Issue{
				File:     file.FilePath,
				Line:     msg.Line,
				Column:   msg.Column,
				Severity: severity,
				Rule:     rule,
				Message:  msg.Message,
			}

			if msg.Fix != nil {
				issue.Suggestion = msg.Fix.Text
			}

			issues = append(issues, issue)
		}
	}

	return issues
}

// parseTextOutput parses plain text output as fallback.
func (t *ESLintTool) parseTextOutput(output string) []Issue {
	var issues []Issue

	// Pattern: file:line:col: message (rule)
	re := regexp.MustCompile(`^(.+):(\d+):(\d+):\s*(error|warning|info)\s*(.+?)\s*(?:\((.+)\))?$`)

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		matches := re.FindStringSubmatch(line)
		if len(matches) >= 6 {
			lineNum, _ := strconv.Atoi(matches[2])
			colNum, _ := strconv.Atoi(matches[3])

			rule := ""
			if len(matches) > 6 && matches[6] != "" {
				rule = matches[6]
			}

			issues = append(issues, Issue{
				File:     matches[1],
				Line:     lineNum,
				Column:   colNum,
				Severity: matches[4],
				Rule:     rule,
				Message:  matches[5],
			})
		}
	}

	return issues
}

// TSCTool implements TypeScript type checking using tsc.
type TSCTool struct {
	*BaseTool
}

// NewTSCTool creates a new tsc tool.
func NewTSCTool() *TSCTool {
	tool := &TSCTool{
		BaseTool: NewBaseTool("tsc", "TypeScript", "tsc", LINT),
	}

	tool.SetInstallCommand([]string{"npm", "install", "-g", "typescript"})
	tool.SetConfigPatterns([]string{"tsconfig.json", "jsconfig.json"})

	return tool
}

// BuildCommand builds the tsc command.
func (t *TSCTool) BuildCommand(files []string, options ExecuteOptions) *exec.Cmd {
	args := []string{"--noEmit"} // Type checking only, no output

	// Add config file if specified
	if options.ConfigFile != "" {
		args = append(args, "--project", options.ConfigFile)
	}

	// Pretty output for better parsing
	args = append(args, "--pretty", "false")

	// Add extra flags if provided
	args = append(args, options.ExtraArgs...)

	// TypeScript files - if specific files provided, use them
	tsFiles := FilterFilesByExtensions(files, []string{".ts", ".tsx"})
	if len(tsFiles) > 0 {
		args = append(args, tsFiles...)
	}
	// If no files specified, tsc will use tsconfig.json

	cmd := exec.Command(t.executable, args...)

	if options.ProjectRoot != "" {
		cmd.Dir = options.ProjectRoot
	}

	return cmd
}

// ParseOutput parses tsc output.
func (t *TSCTool) ParseOutput(output string) []Issue {
	if strings.TrimSpace(output) == "" {
		return []Issue{}
	}

	var issues []Issue

	// Pattern: file(line,col): error TS####: message
	re := regexp.MustCompile(`^(.+?)\((\d+),(\d+)\):\s*(error|warning)\s*TS(\d+):\s*(.+)$`)

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		matches := re.FindStringSubmatch(line)
		if len(matches) == 7 {
			lineNum, _ := strconv.Atoi(matches[2])
			colNum, _ := strconv.Atoi(matches[3])

			issues = append(issues, Issue{
				File:     matches[1],
				Line:     lineNum,
				Column:   colNum,
				Severity: matches[4],
				Rule:     "TS" + matches[5],
				Message:  matches[6],
			})
		}
	}

	return issues
}

// Ensure JavaScript tools implement QualityTool interface.
var (
	_ QualityTool = (*PrettierTool)(nil)
	_ QualityTool = (*ESLintTool)(nil)
	_ QualityTool = (*TSCTool)(nil)
)
