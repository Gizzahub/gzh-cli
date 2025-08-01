// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package tools

import (
	"encoding/json"
	"os/exec"
	"strings"
)

// RustfmtTool implements Rust formatting using rustfmt.
type RustfmtTool struct {
	*BaseTool
}

// NewRustfmtTool creates a new rustfmt tool.
func NewRustfmtTool() *RustfmtTool {
	tool := &RustfmtTool{
		BaseTool: NewBaseTool("rustfmt", "Rust", "rustfmt", FORMAT),
	}

	tool.SetInstallCommand([]string{"rustup", "component", "add", "rustfmt"})
	tool.SetConfigPatterns([]string{"rustfmt.toml", ".rustfmt.toml"})

	return tool
}

// BuildCommand builds the rustfmt command.
func (t *RustfmtTool) BuildCommand(files []string, options ExecuteOptions) *exec.Cmd {
	args := []string{}

	// Add config file if specified
	if options.ConfigFile != "" {
		args = append(args, "--config-path", options.ConfigFile)
	}

	// Add extra flags if provided
	args = append(args, options.ExtraArgs...)

	// Filter only Rust files
	rustFiles := FilterFilesByExtensions(files, []string{".rs"})
	args = append(args, rustFiles...)

	cmd := exec.Command(t.executable, args...)

	if options.ProjectRoot != "" {
		cmd.Dir = options.ProjectRoot
	}

	return cmd
}

// ClippyTool implements Rust linting using clippy.
type ClippyTool struct {
	*BaseTool
}

// NewClippyTool creates a new clippy tool.
func NewClippyTool() *ClippyTool {
	tool := &ClippyTool{
		BaseTool: NewBaseTool("clippy", "Rust", "cargo", LINT),
	}

	tool.SetInstallCommand([]string{"rustup", "component", "add", "clippy"})
	tool.SetConfigPatterns([]string{"clippy.toml", ".clippy.toml", "Cargo.toml"})

	return tool
}

// BuildCommand builds the clippy command.
func (t *ClippyTool) BuildCommand(files []string, options ExecuteOptions) *exec.Cmd {
	args := []string{"clippy"}

	// Add fix flag if requested
	if options.Fix {
		args = append(args, "--fix")
	}

	// Output format for parsing
	args = append(args, "--message-format", "json")

	// Add extra flags
	args = append(args, options.ExtraArgs...)

	// Clippy works on the entire project, not individual files
	args = append(args, "--", "-D", "warnings")

	cmd := exec.Command(t.executable, args...)

	if options.ProjectRoot != "" {
		cmd.Dir = options.ProjectRoot
	}

	return cmd
}

// ParseOutput parses clippy JSON output.
func (t *ClippyTool) ParseOutput(output string) []Issue {
	if strings.TrimSpace(output) == "" {
		return []Issue{}
	}

	var issues []Issue
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var clippyMessage struct {
			Message struct {
				Message string `json:"message"`
				Code    *struct {
					Code string `json:"code"`
				} `json:"code"`
				Level string `json:"level"`
				Spans []struct {
					FileName    string `json:"file_name"`
					LineStart   int    `json:"line_start"`
					ColumnStart int    `json:"column_start"`
				} `json:"spans"`
			} `json:"message"`
			Target struct {
				Name string `json:"name"`
			} `json:"target"`
		}

		if err := json.Unmarshal([]byte(line), &clippyMessage); err != nil {
			continue
		}

		msg := clippyMessage.Message
		if len(msg.Spans) == 0 {
			continue
		}

		severity := "info"
		switch msg.Level {
		case "error":
			severity = "error"
		case "warning":
			severity = "warning"
		}

		rule := ""
		if msg.Code != nil {
			rule = msg.Code.Code
		}

		span := msg.Spans[0]
		issues = append(issues, Issue{
			File:     span.FileName,
			Line:     span.LineStart,
			Column:   span.ColumnStart,
			Severity: severity,
			Rule:     rule,
			Message:  msg.Message,
		})
	}

	return issues
}

// CargoFmtTool implements Rust formatting using cargo fmt.
type CargoFmtTool struct {
	*BaseTool
}

// NewCargoFmtTool creates a new cargo fmt tool.
func NewCargoFmtTool() *CargoFmtTool {
	tool := &CargoFmtTool{
		BaseTool: NewBaseTool("cargo-fmt", "Rust", "cargo", FORMAT),
	}

	tool.SetInstallCommand([]string{"rustup", "component", "add", "rustfmt"})
	tool.SetConfigPatterns([]string{"rustfmt.toml", ".rustfmt.toml"})

	return tool
}

// BuildCommand builds the cargo fmt command.
func (t *CargoFmtTool) BuildCommand(files []string, options ExecuteOptions) *exec.Cmd {
	args := []string{"fmt"}

	// Add extra flags if provided
	args = append(args, options.ExtraArgs...)

	// cargo fmt works on the entire project
	cmd := exec.Command(t.executable, args...)

	if options.ProjectRoot != "" {
		cmd.Dir = options.ProjectRoot
	}

	return cmd
}

// Ensure Rust tools implement QualityTool interface.
var (
	_ QualityTool = (*RustfmtTool)(nil)
	_ QualityTool = (*ClippyTool)(nil)
	_ QualityTool = (*CargoFmtTool)(nil)
)
