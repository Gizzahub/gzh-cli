// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package tools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// BaseTool provides common functionality for quality tools.
type BaseTool struct {
	name           string
	language       string
	toolType       ToolType
	executable     string
	installCmd     []string
	configPatterns []string
}

// NewBaseTool creates a new base tool.
func NewBaseTool(name, language, executable string, toolType ToolType) *BaseTool {
	return &BaseTool{
		name:       name,
		language:   language,
		toolType:   toolType,
		executable: executable,
	}
}

// Name returns the tool name.
func (t *BaseTool) Name() string {
	return t.name
}

// Language returns the programming language.
func (t *BaseTool) Language() string {
	return t.language
}

// Type returns the tool type.
func (t *BaseTool) Type() ToolType {
	return t.toolType
}

// IsAvailable checks if the tool is installed and available.
func (t *BaseTool) IsAvailable() bool {
	_, err := exec.LookPath(t.executable)
	return err == nil
}

// SetInstallCommand sets the command to install this tool.
func (t *BaseTool) SetInstallCommand(cmd []string) {
	t.installCmd = cmd
}

// Install attempts to install the tool automatically.
func (t *BaseTool) Install() error {
	if len(t.installCmd) == 0 {
		return fmt.Errorf("no install command configured for %s", t.name)
	}

	cmd := exec.Command(t.installCmd[0], t.installCmd[1:]...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install %s: %w\nOutput: %s", t.name, err, string(output))
	}

	return nil
}

// GetVersion returns the version of the installed tool.
func (t *BaseTool) GetVersion() (string, error) {
	if !t.IsAvailable() {
		return "", fmt.Errorf("tool %s is not installed", t.name)
	}

	// Try common version flags
	versionFlags := []string{"--version", "-v", "-V", "version"}

	for _, flag := range versionFlags {
		cmd := exec.Command(t.executable, flag)
		output, err := cmd.Output()
		if err == nil {
			version := strings.TrimSpace(string(output))
			if version != "" {
				return version, nil
			}
		}
	}

	return "unknown", nil
}

// Upgrade attempts to upgrade the tool to the latest version.
func (t *BaseTool) Upgrade() error {
	if !t.IsAvailable() {
		return fmt.Errorf("tool %s is not installed, use Install() instead", t.name)
	}

	// For most tools, upgrade is the same as install
	return t.Install()
}

// SetConfigPatterns sets the configuration file patterns to search for.
func (t *BaseTool) SetConfigPatterns(patterns []string) {
	t.configPatterns = patterns
}

// FindConfigFiles returns configuration files the tool would use.
func (t *BaseTool) FindConfigFiles(projectRoot string) []string {
	var configs []string

	for _, pattern := range t.configPatterns {
		configPath := filepath.Join(projectRoot, pattern)
		if _, err := os.Stat(configPath); err == nil {
			configs = append(configs, configPath)
		}
	}

	return configs
}

// ExecuteCommand runs a command and returns the result.
func (t *BaseTool) ExecuteCommand(ctx context.Context, cmd *exec.Cmd, files []string) (*Result, error) {
	startTime := time.Now()

	result := &Result{
		Tool:     t.name,
		Language: t.language,
		Success:  false,
		Issues:   []Issue{},
	}

	output, err := cmd.CombinedOutput()
	duration := time.Since(startTime)
	result.Duration = duration.String()
	result.Output = string(output)
	result.FilesProcessed = len(files)

	if err != nil {
		result.Error = err
		return result, nil //nolint:nilerr // 오류를 결과에 캡처하여 반환하므로 에러는 무시
	}

	result.Success = true
	return result, nil
}

// ParseOutput parses tool output into issues (to be implemented by specific tools).
func (t *BaseTool) ParseOutput(output string) []Issue {
	// Default implementation returns empty slice
	// Specific tools should override this method
	return []Issue{}
}

// BuildCommand builds the command to execute (to be implemented by specific tools).
func (t *BaseTool) BuildCommand(files []string, options ExecuteOptions) *exec.Cmd {
	// Default implementation - specific tools should override
	args := append([]string{}, files...)
	if len(options.ExtraArgs) > 0 {
		args = append(args, options.ExtraArgs...)
	}

	cmd := exec.Command(t.executable, args...)

	// Set working directory
	if options.ProjectRoot != "" {
		cmd.Dir = options.ProjectRoot
	}

	// Set environment variables
	if len(options.Env) > 0 {
		env := os.Environ()
		for k, v := range options.Env {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}
		cmd.Env = env
	}

	return cmd
}

// Execute runs the tool on the specified files.
func (t *BaseTool) Execute(ctx context.Context, files []string, options ExecuteOptions) (*Result, error) {
	if !t.IsAvailable() {
		return &Result{
			Tool:     t.name,
			Language: t.language,
			Success:  false,
			Error:    fmt.Errorf("tool %s is not available", t.name),
		}, nil
	}

	cmd := t.BuildCommand(files, options)
	result, err := t.ExecuteCommand(ctx, cmd, files)
	if err != nil {
		return result, err
	}

	// Parse output for issues if the tool failed
	if !result.Success {
		result.Issues = t.ParseOutput(result.Output)
	}

	return result, nil
}

// FilterFilesByExtensions filters files by supported extensions.
func FilterFilesByExtensions(files, extensions []string) []string {
	var filtered []string
	extMap := make(map[string]bool)

	for _, ext := range extensions {
		extMap[strings.ToLower(ext)] = true
	}

	for _, file := range files {
		ext := strings.ToLower(filepath.Ext(file))
		if extMap[ext] {
			filtered = append(filtered, file)
		}
	}

	return filtered
}
