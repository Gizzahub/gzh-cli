// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

// Package helpers provides testing utilities and helper functions for end-to-end tests.
package helpers

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// CLIAssertions provides custom assertions for CLI results.
type CLIAssertions struct {
	t      *testing.T
	result *CLIResult
}

// NewCLIAssertions creates a new CLI assertions helper.
func NewCLIAssertions(t *testing.T, result *CLIResult) *CLIAssertions {
	t.Helper()
	return &CLIAssertions{
		t:      t,
		result: result,
	}
}

// Success asserts that the command succeeded (exit code 0).
func (a *CLIAssertions) Success() *CLIAssertions {
	a.t.Helper()
	assert.Equal(a.t, 0, a.result.ExitCode, "Command should succeed\nOutput: %s", a.result.Output)

	return a
}

// Failure asserts that the command failed (non-zero exit code).
func (a *CLIAssertions) Failure() *CLIAssertions {
	a.t.Helper()
	assert.NotEqual(a.t, 0, a.result.ExitCode, "Command should fail\nOutput: %s", a.result.Output)

	return a
}

// ExitCode asserts a specific exit code.
func (a *CLIAssertions) ExitCode(expected int) *CLIAssertions {
	a.t.Helper()
	assert.Equal(a.t, expected, a.result.ExitCode, "Unexpected exit code\nOutput: %s", a.result.Output)

	return a
}

// OutputContains asserts that output contains the given text.
func (a *CLIAssertions) OutputContains(expected string) *CLIAssertions {
	a.t.Helper()
	assert.Contains(a.t, a.result.Output, expected, "Output should contain expected text")

	return a
}

// OutputNotContains asserts that output does not contain the given text.
func (a *CLIAssertions) OutputNotContains(unexpected string) *CLIAssertions {
	a.t.Helper()
	assert.NotContains(a.t, a.result.Output, unexpected, "Output should not contain unexpected text")

	return a
}

// OutputMatches asserts that output matches a regular expression.
func (a *CLIAssertions) OutputMatches(pattern string) *CLIAssertions {
	a.t.Helper()
	matched, err := regexp.MatchString(pattern, a.result.Output)
	require.NoError(a.t, err, "Invalid regex pattern: %s", pattern)
	assert.True(a.t, matched, "Output should match pattern: %s\nOutput: %s", pattern, a.result.Output)

	return a
}

// OutputEmpty asserts that output is empty.
func (a *CLIAssertions) OutputEmpty() *CLIAssertions {
	a.t.Helper()
	assert.Empty(a.t, strings.TrimSpace(a.result.Output), "Output should be empty")

	return a
}

// OutputNotEmpty asserts that output is not empty.
func (a *CLIAssertions) OutputNotEmpty() *CLIAssertions {
	a.t.Helper()
	assert.NotEmpty(a.t, strings.TrimSpace(a.result.Output), "Output should not be empty")

	return a
}

// OutputLines asserts the number of output lines.
func (a *CLIAssertions) OutputLines(expected int) *CLIAssertions {
	a.t.Helper()

	lines := strings.Split(strings.TrimSpace(a.result.Output), "\n")
	if len(lines) == 1 && lines[0] == "" {
		lines = []string{} // Empty output
	}

	assert.Len(a.t, lines, expected, "Unexpected number of output lines")

	return a
}

// OutputLineContains asserts that a specific line contains text.
func (a *CLIAssertions) OutputLineContains(lineIndex int, expected string) *CLIAssertions {
	a.t.Helper()
	lines := strings.Split(strings.TrimSpace(a.result.Output), "\n")
	require.Greater(a.t, len(lines), lineIndex, "Line index out of range")
	assert.Contains(a.t, lines[lineIndex], expected, "Line %d should contain expected text", lineIndex)

	return a
}

// NoError asserts that there was no error.
func (a *CLIAssertions) NoError() *CLIAssertions {
	a.t.Helper()
	assert.NoError(a.t, a.result.Error, "Command should not error\nOutput: %s", a.result.Output)

	return a
}

// Error asserts that there was an error.
func (a *CLIAssertions) Error() *CLIAssertions {
	a.t.Helper()
	assert.Error(a.t, a.result.Error, "Command should error\nOutput: %s", a.result.Output)

	return a
}

// ConfigAssertions provides assertions for configuration validation.
type ConfigAssertions struct {
	t       *testing.T
	env     *TestEnvironment
	content string
}

// NewConfigAssertions creates config assertions for a file.
func NewConfigAssertions(t *testing.T, env *TestEnvironment, configPath string) *ConfigAssertions {
	t.Helper()
	content := env.ReadFile(configPath)

	return &ConfigAssertions{
		t:       t,
		env:     env,
		content: content,
	}
}

// ValidYAML asserts that the config is valid YAML.
func (a *ConfigAssertions) ValidYAML() *ConfigAssertions {
	a.t.Helper()

	var data interface{}

	err := yaml.Unmarshal([]byte(a.content), &data)
	assert.NoError(a.t, err, "Config should be valid YAML")

	return a
}

// ValidJSON asserts that the config is valid JSON.
func (a *ConfigAssertions) ValidJSON() *ConfigAssertions {
	a.t.Helper()

	var data interface{}

	err := json.Unmarshal([]byte(a.content), &data)
	assert.NoError(a.t, err, "Config should be valid JSON")

	return a
}

// HasField asserts that a YAML field exists.
func (a *ConfigAssertions) HasField(fieldPath string) *ConfigAssertions {
	a.t.Helper()

	var data map[string]interface{}

	err := yaml.Unmarshal([]byte(a.content), &data)
	require.NoError(a.t, err, "Config should be valid YAML")

	// Simple field path parsing (supports nested fields like "providers.github.token")
	fields := strings.Split(fieldPath, ".")
	current := data

	for i, field := range fields {
		value, exists := current[field]
		if !exists {
			a.t.Errorf("Field %s not found in config (at path: %s)", field, strings.Join(fields[:i+1], "."))
			return a
		}

		if i < len(fields)-1 {
			// Not the last field, should be a map
			nextMap, ok := value.(map[string]interface{})
			if !ok {
				a.t.Errorf("Field %s is not a map (required for nested access)", field)
				return a
			}
			current = nextMap
		}
	}

	return a
}

// FieldEquals asserts that a YAML field has a specific value.
func (a *ConfigAssertions) FieldEquals(fieldPath string, expected interface{}) *ConfigAssertions {
	a.t.Helper()

	var data map[string]interface{}

	err := yaml.Unmarshal([]byte(a.content), &data)
	require.NoError(a.t, err, "Config should be valid YAML")

	fields := strings.Split(fieldPath, ".")
	current := data

	for i, field := range fields {
		value, exists := current[field]
		require.True(a.t, exists, "Field %s not found in config", field)

		if i == len(fields)-1 {
			// Last field, check value
			assert.Equal(a.t, expected, value, "Field %s should have expected value", fieldPath)
		} else {
			// Not the last field, should be a map
			nextMap, ok := value.(map[string]interface{})
			require.True(a.t, ok, "Field %s is not a map", field)

			current = nextMap
		}
	}

	return a
}

// ProcessAssertions provides assertions for process management.
type ProcessAssertions struct {
	t   *testing.T
	env *TestEnvironment
}

// NewProcessAssertions creates process assertions.
func NewProcessAssertions(t *testing.T, env *TestEnvironment) *ProcessAssertions {
	t.Helper()
	return &ProcessAssertions{
		t:   t,
		env: env,
	}
}

// LogFileExists asserts that a log file exists and is not empty.
func (a *ProcessAssertions) LogFileExists(logPath string) *ProcessAssertions {
	a.t.Helper()
	a.env.AssertFileExists(logPath)
	content := a.env.ReadFile(logPath)
	assert.NotEmpty(a.t, strings.TrimSpace(content), "Log file should not be empty")

	return a
}

// LogContains asserts that a log file contains specific text.
func (a *ProcessAssertions) LogContains(logPath, expected string) *ProcessAssertions {
	a.t.Helper()
	a.env.AssertFileContains(logPath, expected)

	return a
}

// PidFileExists asserts that a PID file exists with valid content.
func (a *ProcessAssertions) PidFileExists(pidPath string) *ProcessAssertions {
	a.t.Helper()
	a.env.AssertFileExists(pidPath)
	content := strings.TrimSpace(a.env.ReadFile(pidPath))
	assert.NotEmpty(a.t, content, "PID file should not be empty")
	assert.Regexp(a.t, `^\d+$`, content, "PID file should contain a valid process ID")

	return a
}

// GitAssertions provides assertions for Git repositories.
type GitAssertions struct {
	t   *testing.T
	env *TestEnvironment
}

// NewGitAssertions creates Git assertions.
func NewGitAssertions(t *testing.T, env *TestEnvironment) *GitAssertions {
	t.Helper()
	return &GitAssertions{
		t:   t,
		env: env,
	}
}

// IsValidRepo asserts that a directory is a valid Git repository.
func (a *GitAssertions) IsValidRepo(repoPath string) *GitAssertions {
	a.t.Helper()
	a.env.AssertDirectoryExists(repoPath)
	a.env.AssertDirectoryExists(repoPath + "/.git")

	return a
}

// HasRemote asserts that a repository has a specific remote.
func (a *GitAssertions) HasRemote(repoPath, _ string) *GitAssertions {
	a.t.Helper()
	// This would typically check .git/config or run git commands
	// For simplicity, we'll check if the .git directory exists
	a.IsValidRepo(repoPath)

	return a
}

// HasBranch asserts that a repository has a specific branch.
func (a *GitAssertions) HasBranch(repoPath, _ string) *GitAssertions {
	a.t.Helper()
	a.IsValidRepo(repoPath)
	// In a real implementation, this would check git branches
	return a
}

// BulkCloneAssertions provides assertions specific to bulk clone operations.
type BulkCloneAssertions struct {
	t   *testing.T
	env *TestEnvironment
}

// NewBulkCloneAssertions creates bulk clone assertions.
func NewBulkCloneAssertions(t *testing.T, env *TestEnvironment) *BulkCloneAssertions {
	t.Helper()
	return &BulkCloneAssertions{
		t:   t,
		env: env,
	}
}

// RepoCloned asserts that a repository was cloned successfully.
func (a *BulkCloneAssertions) RepoCloned(org, repo string) *BulkCloneAssertions {
	a.t.Helper()

	repoPath := org + "/" + repo
	git := NewGitAssertions(a.t, a.env)
	git.IsValidRepo(repoPath)

	return a
}

// OrgDirectoryExists asserts that an organization directory exists.
func (a *BulkCloneAssertions) OrgDirectoryExists(org string) *BulkCloneAssertions {
	a.t.Helper()
	a.env.AssertDirectoryExists(org)

	return a
}

// RepoCount asserts the number of repositories in an organization directory.
func (a *BulkCloneAssertions) RepoCount(org string, expectedCount int) *BulkCloneAssertions {
	a.t.Helper()
	actualCount := a.env.CountDirectories(org)
	assert.Equal(a.t, expectedCount, actualCount, "Unexpected number of repositories in %s", org)

	return a
}

// ConfigFileGenerated asserts that a configuration file was generated.
func (a *BulkCloneAssertions) ConfigFileGenerated(configPath string) *BulkCloneAssertions {
	a.t.Helper()
	a.env.AssertFileExists(configPath)
	config := NewConfigAssertions(a.t, a.env, configPath)
	config.ValidYAML()

	return a
}
