// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// BlackTool implements Python formatting using black.
type BlackTool struct {
	*BaseTool
}

// NewBlackTool creates a new black tool.
func NewBlackTool() *BlackTool {
	tool := &BlackTool{
		BaseTool: NewBaseTool("black", "Python", "black", FORMAT),
	}

	tool.SetInstallCommand([]string{"pip", "install", "black"})
	tool.SetConfigPatterns([]string{"pyproject.toml", ".black", "black.toml"})

	return tool
}

// BuildCommand builds the black command.
func (t *BlackTool) BuildCommand(files []string, options ExecuteOptions) *exec.Cmd {
	args := []string{}

	// Add config file if specified
	if options.ConfigFile != "" {
		args = append(args, "--config", options.ConfigFile)
	}

	// Add line length if not in config
	args = append(args, "--line-length", "88") // black default

	// Add extra flags if provided
	args = append(args, options.ExtraArgs...)

	// Filter only Python files
	pyFiles := FilterFilesByExtensions(files, []string{".py", ".pyi"})
	if len(pyFiles) == 0 {
		// If no specific files, format all Python files in project
		args = append(args, ".")
	} else {
		args = append(args, pyFiles...)
	}

	cmd := exec.Command(t.executable, args...)

	if options.ProjectRoot != "" {
		cmd.Dir = options.ProjectRoot
	}

	return cmd
}

// RuffTool implements Python linting and formatting using ruff.
type RuffTool struct {
	*BaseTool
}

// NewRuffTool creates a new ruff tool.
func NewRuffTool() *RuffTool {
	tool := &RuffTool{
		BaseTool: NewBaseTool("ruff", "Python", "ruff", BOTH),
	}

	tool.SetInstallCommand([]string{"pip", "install", "ruff"})
	tool.SetConfigPatterns([]string{"ruff.toml", ".ruff.toml", "pyproject.toml"})

	return tool
}

// BuildCommand builds the ruff command.
func (t *RuffTool) BuildCommand(files []string, options ExecuteOptions) *exec.Cmd {
	var args []string

	// Determine operation mode
	if options.FormatOnly {
		args = append(args, "format")
	} else {
		args = append(args, "check")
	}

	// Add config file if specified
	if options.ConfigFile != "" {
		args = append(args, "--config", options.ConfigFile)
	}

	// Add fix flag if requested and not format-only
	if options.Fix && !options.FormatOnly {
		args = append(args, "--fix")
	}

	// Output format for parsing
	if !options.FormatOnly {
		args = append(args, "--output-format", "json")
	}

	// Add extra flags if provided
	args = append(args, options.ExtraArgs...)

	// Filter only Python files
	pyFiles := FilterFilesByExtensions(files, []string{".py", ".pyi"})
	if len(pyFiles) == 0 {
		// If no specific files, check all Python files in project
		args = append(args, ".")
	} else {
		args = append(args, pyFiles...)
	}

	cmd := exec.Command(t.executable, args...)

	if options.ProjectRoot != "" {
		cmd.Dir = options.ProjectRoot
	}

	return cmd
}

// Execute overrides the base Execute to handle both format and lint modes.
func (t *RuffTool) Execute(ctx context.Context, files []string, options ExecuteOptions) (*Result, error) {
	if !t.IsAvailable() {
		return &Result{
			Tool:     t.name,
			Language: t.language,
			Success:  false,
			Error:    fmt.Errorf("tool %s is not available", t.name),
		}, nil
	}

	// If BOTH is requested, run format first, then lint
	if !options.FormatOnly && !options.LintOnly {
		// Run format first
		formatOptions := options
		formatOptions.FormatOnly = true
		formatResult, err := t.executeMode(ctx, files, formatOptions)
		if err != nil {
			return formatResult, err
		}

		// Run lint second
		lintOptions := options
		lintOptions.LintOnly = true
		lintResult, err := t.executeMode(ctx, files, lintOptions)
		if err != nil {
			return lintResult, err
		}

		// Combine results
		combinedResult := &Result{
			Tool:           t.name,
			Language:       t.language,
			Success:        formatResult.Success && lintResult.Success,
			FilesProcessed: formatResult.FilesProcessed,
			Duration:       formatResult.Duration + " + " + lintResult.Duration,
			Issues:         lintResult.Issues,
			Output:         formatResult.Output + "\n" + lintResult.Output,
		}

		if !combinedResult.Success {
			if formatResult.Error != nil {
				combinedResult.Error = formatResult.Error
			} else if lintResult.Error != nil {
				combinedResult.Error = lintResult.Error
			}
		}

		return combinedResult, nil
	}

	return t.executeMode(ctx, files, options)
}

// executeMode executes ruff in a specific mode.
func (t *RuffTool) executeMode(ctx context.Context, files []string, options ExecuteOptions) (*Result, error) {
	cmd := t.BuildCommand(files, options)
	result, err := t.ExecuteCommand(ctx, cmd, files)
	if err != nil {
		return result, err
	}

	// Parse output for issues if linting and the tool found issues
	if !options.FormatOnly && result.Output != "" {
		result.Issues = t.ParseOutput(result.Output)
	}

	return result, nil
}

// ParseOutput parses ruff JSON output.
func (t *RuffTool) ParseOutput(output string) []Issue {
	if strings.TrimSpace(output) == "" {
		return []Issue{}
	}

	var ruffIssues []struct {
		Code     string `json:"code"`
		Message  string `json:"message"`
		Filename string `json:"filename"`
		Location struct {
			Row    int `json:"row"`
			Column int `json:"column"`
		} `json:"location"`
		EndLocation struct {
			Row    int `json:"row"`
			Column int `json:"column"`
		} `json:"end_location"`
		Fix *struct {
			Content string `json:"content"`
		} `json:"fix,omitempty"`
	}

	if err := json.Unmarshal([]byte(output), &ruffIssues); err != nil {
		// Fallback to plain text parsing
		return t.parseTextOutput(output)
	}

	issues := make([]Issue, 0, len(ruffIssues))
	for _, item := range ruffIssues {
		issue := Issue{
			File:     item.Filename,
			Line:     item.Location.Row,
			Column:   item.Location.Column,
			Severity: "error", // Ruff doesn't distinguish severity in JSON
			Rule:     item.Code,
			Message:  item.Message,
		}

		if item.Fix != nil {
			issue.Suggestion = item.Fix.Content
		}

		issues = append(issues, issue)
	}

	return issues
}

// parseTextOutput parses plain text output as fallback.
func (t *RuffTool) parseTextOutput(output string) []Issue {
	var issues []Issue

	// Pattern: file:line:col: code message
	re := regexp.MustCompile(`^(.+):(\d+):(\d+):\s*([A-Z]\d+)\s*(.+)$`)

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		matches := re.FindStringSubmatch(line)
		if len(matches) == 6 {
			lineNum, _ := strconv.Atoi(matches[2])
			colNum, _ := strconv.Atoi(matches[3])

			issues = append(issues, Issue{
				File:     matches[1],
				Line:     lineNum,
				Column:   colNum,
				Severity: "error",
				Rule:     matches[4],
				Message:  matches[5],
			})
		}
	}

	return issues
}

// PylintTool implements Python linting using pylint.
type PylintTool struct {
	*BaseTool
}

// NewPylintTool creates a new pylint tool.
func NewPylintTool() *PylintTool {
	tool := &PylintTool{
		BaseTool: NewBaseTool("pylint", "Python", "pylint", LINT),
	}

	tool.SetInstallCommand([]string{"pip", "install", "pylint"})
	tool.SetConfigPatterns([]string{".pylintrc", "pylint.cfg", "pyproject.toml"})

	return tool
}

// BuildCommand builds the pylint command.
func (t *PylintTool) BuildCommand(files []string, options ExecuteOptions) *exec.Cmd {
	args := []string{}

	// Add config file if specified
	if options.ConfigFile != "" {
		args = append(args, "--rcfile", options.ConfigFile)
	}

	// Output format for parsing
	args = append(args, "--output-format", "json")

	// Add extra flags if provided
	args = append(args, options.ExtraArgs...)

	// Filter only Python files
	pyFiles := FilterFilesByExtensions(files, []string{".py"})
	if len(pyFiles) == 0 {
		// If no specific files, check all Python files in project
		args = append(args, ".")
	} else {
		args = append(args, pyFiles...)
	}

	cmd := exec.Command(t.executable, args...)

	if options.ProjectRoot != "" {
		cmd.Dir = options.ProjectRoot
	}

	return cmd
}

// ParseOutput parses pylint JSON output.
func (t *PylintTool) ParseOutput(output string) []Issue {
	if strings.TrimSpace(output) == "" {
		return []Issue{}
	}

	var pylintIssues []struct {
		Type      string `json:"type"`
		Module    string `json:"module"`
		Obj       string `json:"obj"`
		Line      int    `json:"line"`
		Column    int    `json:"column"`
		Path      string `json:"path"`
		Symbol    string `json:"symbol"`
		Message   string `json:"message"`
		MessageId string `json:"message-id"`
	}

	if err := json.Unmarshal([]byte(output), &pylintIssues); err != nil {
		return []Issue{} // pylint text output is complex, skip fallback
	}

	issues := make([]Issue, 0, len(pylintIssues))
	for _, item := range pylintIssues {
		severity := "info"
		switch item.Type {
		case "error", "fatal":
			severity = "error"
		case "warning":
			severity = "warning"
		}

		issues = append(issues, Issue{
			File:     item.Path,
			Line:     item.Line,
			Column:   item.Column,
			Severity: severity,
			Rule:     item.MessageId,
			Message:  item.Message,
		})
	}

	return issues
}

// Ensure Python tools implement QualityTool interface.
var (
	_ QualityTool = (*BlackTool)(nil)
	_ QualityTool = (*RuffTool)(nil)
	_ QualityTool = (*PylintTool)(nil)
)
