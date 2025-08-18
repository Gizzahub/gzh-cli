// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package doctor

import (
	"context"
	"os/exec"
	"strings"
)

// ToolConfig defines configuration for a development tool check.
type ToolConfig struct {
	Name              string
	Command           string
	VersionArgs       []string
	InstallSuggestion string
}

// ToolChecker provides a common pattern for checking development tools.
type ToolChecker struct {
	config ToolConfig
}

// NewToolChecker creates a new tool checker with the given configuration.
func NewToolChecker(config ToolConfig) *ToolChecker {
	return &ToolChecker{config: config}
}

// Check performs the tool check following the common pattern.
func (tc *ToolChecker) Check(ctx context.Context) DevEnvResult {
	result := DevEnvResult{Tool: tc.config.Name, Status: "missing"}

	// Check if tool exists in PATH
	toolPath, err := exec.LookPath(tc.config.Command)
	if err != nil {
		result.Suggestion = tc.config.InstallSuggestion
		return result
	}

	result.Path = toolPath
	result.Status = "found"

	// Check version if version args are provided
	if len(tc.config.VersionArgs) > 0 {
		version, err := tc.getVersion(ctx)
		if err != nil {
			result.Status = "error"
			result.Details = map[string]interface{}{"error": err.Error()}
			return result
		}
		result.Version = version
	}

	result.Status = "ok"
	return result
}

// getVersion retrieves the version of the tool.
func (tc *ToolChecker) getVersion(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, tc.config.Command, tc.config.VersionArgs...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// CommonToolConfigs provides configurations for common development tools.
var CommonToolConfigs = map[string]ToolConfig{
	"golangci-lint": {
		Name:              "golangci-lint",
		Command:           "golangci-lint",
		VersionArgs:       []string{"version"},
		InstallSuggestion: "Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest",
	},
	"gofumpt": {
		Name:              "gofumpt",
		Command:           "gofumpt",
		VersionArgs:       []string{"-version"},
		InstallSuggestion: "Install with: go install mvdan.cc/gofumpt@latest",
	},
	"gci": {
		Name:              "gci",
		Command:           "gci",
		VersionArgs:       []string{"--version"},
		InstallSuggestion: "Install with: go install github.com/daixiang0/gci@latest",
	},
	"deadcode": {
		Name:              "deadcode",
		Command:           "deadcode",
		VersionArgs:       []string{"-version"},
		InstallSuggestion: "Install with: go install golang.org/x/tools/cmd/deadcode@latest",
	},
	"dupl": {
		Name:              "dupl",
		Command:           "dupl",
		VersionArgs:       []string{"-version"},
		InstallSuggestion: "Install with: go install github.com/mibk/dupl@latest",
	},
}

// CreateToolChecker creates a tool checker for a known tool.
func CreateToolChecker(toolName string) *ToolChecker {
	config, exists := CommonToolConfigs[toolName]
	if !exists {
		// Return a default checker for unknown tools
		config = ToolConfig{
			Name:              toolName,
			Command:           toolName,
			VersionArgs:       []string{"--version"},
			InstallSuggestion: "Please install " + toolName,
		}
	}
	return NewToolChecker(config)
}

// CheckMultipleTools checks multiple tools and returns their results.
func CheckMultipleTools(ctx context.Context, toolNames []string) []DevEnvResult {
	results := make([]DevEnvResult, 0, len(toolNames))
	for _, toolName := range toolNames {
		checker := CreateToolChecker(toolName)
		result := checker.Check(ctx)
		results = append(results, result)
	}
	return results
}

// CheckAllCommonTools checks all commonly used development tools.
func CheckAllCommonTools(ctx context.Context) []DevEnvResult {
	toolNames := make([]string, 0, len(CommonToolConfigs))
	for toolName := range CommonToolConfigs {
		toolNames = append(toolNames, toolName)
	}
	return CheckMultipleTools(ctx, toolNames)
}
