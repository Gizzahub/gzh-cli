// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package tools

import (
	"encoding/json"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// GofumptTool implements Go formatting using gofumpt.
type GofumptTool struct {
	*BaseTool
}

// NewGofumptTool creates a new gofumpt tool.
func NewGofumptTool() *GofumptTool {
	tool := &GofumptTool{
		BaseTool: NewBaseTool("gofumpt", "Go", "gofumpt", FORMAT),
	}

	tool.SetInstallCommand([]string{"go", "install", "mvdan.cc/gofumpt@latest"})
	tool.SetConfigPatterns([]string{".gofumpt"})

	return tool
}

// BuildCommand builds the gofumpt command.
func (t *GofumptTool) BuildCommand(files []string, options ExecuteOptions) *exec.Cmd {
	args := []string{"-w"} // Always write changes

	// Add extra flags if provided
	args = append(args, options.ExtraArgs...)

	// Filter only Go files
	goFiles := FilterFilesByExtensions(files, []string{".go"})
	args = append(args, goFiles...)

	cmd := exec.Command(t.executable, args...)

	if options.ProjectRoot != "" {
		cmd.Dir = options.ProjectRoot
	}

	return cmd
}

// GoimportsTool implements Go import formatting using goimports.
type GoimportsTool struct {
	*BaseTool
}

// NewGoimportsTool creates a new goimports tool.
func NewGoimportsTool() *GoimportsTool {
	tool := &GoimportsTool{
		BaseTool: NewBaseTool("goimports", "Go", "goimports", FORMAT),
	}

	tool.SetInstallCommand([]string{"go", "install", "golang.org/x/tools/cmd/goimports@latest"})

	return tool
}

// BuildCommand builds the goimports command.
func (t *GoimportsTool) BuildCommand(files []string, options ExecuteOptions) *exec.Cmd {
	args := []string{"-w"} // Always write changes

	// Add local import setting if project root is available
	if options.ProjectRoot != "" {
		// Try to determine module name from go.mod
		if modName := getGoModuleName(options.ProjectRoot); modName != "" {
			args = append(args, "-local", modName)
		}
	}

	// Add extra flags if provided
	args = append(args, options.ExtraArgs...)

	// Filter only Go files
	goFiles := FilterFilesByExtensions(files, []string{".go"})
	args = append(args, goFiles...)

	cmd := exec.Command(t.executable, args...)

	if options.ProjectRoot != "" {
		cmd.Dir = options.ProjectRoot
	}

	return cmd
}

// GolangciLintTool implements Go linting using golangci-lint.
type GolangciLintTool struct {
	*BaseTool
}

// NewGolangciLintTool creates a new golangci-lint tool.
func NewGolangciLintTool() *GolangciLintTool {
	tool := &GolangciLintTool{
		BaseTool: NewBaseTool("golangci-lint", "Go", "golangci-lint", LINT),
	}

	tool.SetInstallCommand([]string{"go", "install", "github.com/golangci/golangci-lint/cmd/golangci-lint@latest"})
	tool.SetConfigPatterns([]string{".golangci.yml", ".golangci.yaml", "golangci.yml", "golangci.yaml"})

	return tool
}

// BuildCommand builds the golangci-lint command.
func (t *GolangciLintTool) BuildCommand(files []string, options ExecuteOptions) *exec.Cmd {
	args := []string{"run"}

	// Add config file if specified
	if options.ConfigFile != "" {
		args = append(args, "-c", options.ConfigFile)
	}

	// Add fix flag if requested
	if options.Fix {
		args = append(args, "--fix")
	}

	// Output format for parsing
	args = append(args, "--out-format", "json")

	// Add extra flags if provided
	args = append(args, options.ExtraArgs...)

	// Add file patterns or directories
	if len(files) > 0 {
		// Filter only Go files and their directories
		goFiles := FilterFilesByExtensions(files, []string{".go"})
		if len(goFiles) > 0 {
			// Get unique directories
			dirs := make(map[string]bool)
			for _, file := range goFiles {
				dirs[filepath.Dir(file)] = true
			}

			for dir := range dirs {
				args = append(args, dir)
			}
		}
	} else {
		args = append(args, "./...")
	}

	cmd := exec.Command(t.executable, args...)

	if options.ProjectRoot != "" {
		cmd.Dir = options.ProjectRoot
	}

	return cmd
}

// ParseOutput parses golangci-lint JSON output.
func (t *GolangciLintTool) ParseOutput(output string) []Issue {
	if strings.TrimSpace(output) == "" {
		return []Issue{}
	}

	var lintResults struct {
		Issues []struct {
			FromLinter  string   `json:"FromLinter"`
			Text        string   `json:"Text"`
			Severity    string   `json:"Severity"`
			SourceLines []string `json:"SourceLines"`
			Replacement *struct {
				NewLines []string `json:"NewLines"`
			} `json:"Replacement,omitempty"`
			Pos struct {
				Filename string `json:"Filename"`
				Offset   int    `json:"Offset"`
				Line     int    `json:"Line"`
				Column   int    `json:"Column"`
			} `json:"Pos"`
		} `json:"Issues"`
	}

	if err := json.Unmarshal([]byte(output), &lintResults); err != nil {
		// Fallback to plain text parsing
		return t.parseTextOutput(output)
	}

	var issues []Issue
	for _, item := range lintResults.Issues {
		issue := Issue{
			File:     item.Pos.Filename,
			Line:     item.Pos.Line,
			Column:   item.Pos.Column,
			Severity: item.Severity,
			Rule:     item.FromLinter,
			Message:  item.Text,
		}

		if item.Replacement != nil && len(item.Replacement.NewLines) > 0 {
			issue.Suggestion = strings.Join(item.Replacement.NewLines, "\n")
		}

		issues = append(issues, issue)
	}

	return issues
}

// parseTextOutput parses plain text output as fallback.
func (t *GolangciLintTool) parseTextOutput(output string) []Issue {
	var issues []Issue

	// Pattern: file:line:col: message (rule)
	re := regexp.MustCompile(`^(.+):(\d+):(\d+):\s*(.+)\s*\((.+)\)$`)

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
				Severity: "error", // Default severity
				Rule:     matches[5],
				Message:  matches[4],
			})
		}
	}

	return issues
}

// getGoModuleName extracts module name from go.mod file.
func getGoModuleName(projectRoot string) string {
	cmd := exec.Command("go", "list", "-m")
	cmd.Dir = projectRoot

	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(output))
}

// Ensure Go tools implement QualityTool interface.
var (
	_ QualityTool = (*GofumptTool)(nil)
	_ QualityTool = (*GoimportsTool)(nil)
	_ QualityTool = (*GolangciLintTool)(nil)
)
