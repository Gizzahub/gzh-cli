// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package extensions

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterWorkflowAlias(t *testing.T) {
	rootCmd := &cobra.Command{Use: "test"}
	loader := &Loader{}

	alias := AliasConfig{
		Description: "Test workflow",
		Steps: []string{
			"version",
			"help",
		},
	}

	err := loader.registerWorkflowAlias(rootCmd, "test-workflow", alias)
	require.NoError(t, err)

	// Verify command was added
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "test-workflow" {
			found = true
			assert.Contains(t, cmd.Long, "[WORKFLOW]")
			assert.Contains(t, cmd.Long, "version")
			assert.Contains(t, cmd.Long, "help")
		}
	}
	assert.True(t, found, "Workflow command should be registered")
}

func TestRegisterWorkflowAlias_EmptySteps(t *testing.T) {
	rootCmd := &cobra.Command{Use: "test"}
	loader := &Loader{}

	alias := AliasConfig{
		Description: "Empty workflow",
		Steps:       []string{},
	}

	err := loader.registerWorkflowAlias(rootCmd, "empty-workflow", alias)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no steps")
}

func TestRegisterParameterizedAlias(t *testing.T) {
	rootCmd := &cobra.Command{Use: "test"}
	loader := &Loader{}

	alias := AliasConfig{
		Command:     "git repo clone-or-update ${url}",
		Description: "Clone with parameter",
		Params: []Param{
			{Name: "url", Description: "Repository URL", Required: true},
		},
	}

	err := loader.registerParameterizedAlias(rootCmd, "clone", alias)
	require.NoError(t, err)

	// Verify command was added
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "clone" {
			found = true
			assert.Contains(t, cmd.Use, "<url>") // Required param shown with <>
			assert.Contains(t, cmd.Long, "[PARAMETERIZED]")
			assert.Contains(t, cmd.Long, "url")
		}
	}
	assert.True(t, found, "Parameterized command should be registered")
}

func TestRegisterParameterizedAlias_OptionalParams(t *testing.T) {
	rootCmd := &cobra.Command{Use: "test"}
	loader := &Loader{}

	alias := AliasConfig{
		Command:     "quality check $path",
		Description: "Check code quality",
		Params: []Param{
			{Name: "path", Description: "Path to check", Required: false},
		},
	}

	err := loader.registerParameterizedAlias(rootCmd, "check", alias)
	require.NoError(t, err)

	// Verify command was added with optional param
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "check" {
			found = true
			assert.Contains(t, cmd.Use, "[path]") // Optional param shown with []
		}
	}
	assert.True(t, found, "Parameterized command with optional params should be registered")
}

func TestFormatSteps(t *testing.T) {
	steps := []string{
		"step one",
		"step two",
		"step three",
	}

	result := formatSteps(steps)
	assert.Contains(t, result, "1. step one")
	assert.Contains(t, result, "2. step two")
	assert.Contains(t, result, "3. step three")
}

func TestFormatParams(t *testing.T) {
	params := []Param{
		{Name: "required-param", Description: "A required parameter", Required: true},
		{Name: "optional-param", Description: "An optional parameter", Required: false},
	}

	result := formatParams(params)
	assert.Contains(t, result, "required-param (required)")
	assert.Contains(t, result, "optional-param:")
	assert.Contains(t, result, "A required parameter")
	assert.Contains(t, result, "An optional parameter")
}

func TestLoadConfig_WithWorkflowAndParams(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "extensions.yaml")

	configContent := `
aliases:
  workflow-test:
    description: "Multi-step workflow test"
    steps:
      - "version"
      - "help"

  param-test:
    command: "git repo clone-or-update ${url}"
    description: "Parameterized alias test"
    params:
      - name: url
        description: "Repository URL"
        required: true

external:
  - name: test-ext
    command: /bin/echo
    description: "Test external command"
    passthrough: true
`

	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	loader := &Loader{configPath: configPath}
	cfg, err := loader.LoadConfig()
	require.NoError(t, err)

	// Verify workflow alias
	workflowAlias, ok := cfg.Aliases["workflow-test"]
	assert.True(t, ok)
	assert.Equal(t, "Multi-step workflow test", workflowAlias.Description)
	assert.Len(t, workflowAlias.Steps, 2)
	assert.Equal(t, "version", workflowAlias.Steps[0])

	// Verify parameterized alias
	paramAlias, ok := cfg.Aliases["param-test"]
	assert.True(t, ok)
	assert.Equal(t, "Parameterized alias test", paramAlias.Description)
	assert.Contains(t, paramAlias.Command, "${url}")
	assert.Len(t, paramAlias.Params, 1)
	assert.Equal(t, "url", paramAlias.Params[0].Name)
	assert.True(t, paramAlias.Params[0].Required)
}

func TestRegisterAll_MixedAliases(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "extensions.yaml")

	configContent := `
aliases:
  simple:
    command: "version"
    description: "Simple alias"

  workflow:
    description: "Workflow alias"
    steps:
      - "version"
      - "help"

  parameterized:
    command: "git repo clone-or-update ${url}"
    description: "Parameterized alias"
    params:
      - name: url
        description: "Repository URL"
        required: true
`

	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	loader := &Loader{configPath: configPath}
	rootCmd := &cobra.Command{Use: "test"}

	err = loader.RegisterAll(rootCmd)
	require.NoError(t, err)

	// Verify all aliases were registered
	commandNames := make([]string, 0)
	for _, cmd := range rootCmd.Commands() {
		commandNames = append(commandNames, cmd.Name())
	}

	assert.Contains(t, commandNames, "simple")
	assert.Contains(t, commandNames, "workflow")
	assert.Contains(t, commandNames, "parameterized")
}
