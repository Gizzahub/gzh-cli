// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package testlib

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// DevEnvOptions represents common options structure for dev-env commands.
type DevEnvOptions struct {
	ConfigPath string
	StorePath  string
	Force      bool
	ListAll    bool
}

// AssertDefaultOptions verifies that default options have expected values.
func AssertDefaultOptions(t *testing.T, opts DevEnvOptions) {
	t.Helper()
	assert.NotEmpty(t, opts.ConfigPath, "ConfigPath should not be empty")
	assert.NotEmpty(t, opts.StorePath, "StorePath should not be empty")
	assert.False(t, opts.Force, "Force should be false by default")
	assert.False(t, opts.ListAll, "ListAll should be false by default")
}

// AssertDevEnvCommand verifies that a dev-env command has the expected structure.
func AssertDevEnvCommand(t *testing.T, cmd *cobra.Command, expectedUse, expectedShort string) {
	t.Helper()
	assert.Equal(t, expectedUse, cmd.Use, "Command use should match")
	assert.Equal(t, expectedShort, cmd.Short, "Command short description should match")
	assert.NotEmpty(t, cmd.Long, "Command long description should not be empty")

	// Check subcommands
	subcommands := cmd.Commands()
	assert.Len(t, subcommands, 3, "Should have exactly 3 subcommands")

	var saveCmd, loadCmd, listCmd bool
	for _, subcmd := range subcommands {
		switch subcmd.Use {
		case "save":
			saveCmd = true
		case "load":
			loadCmd = true
		case "list":
			listCmd = true
		}
	}

	assert.True(t, saveCmd, "Should have 'save' subcommand")
	assert.True(t, loadCmd, "Should have 'load' subcommand")
	assert.True(t, listCmd, "Should have 'list' subcommand")
}

// AssertCommandWithFlags verifies that a command has expected flags.
func AssertCommandWithFlags(t *testing.T, cmd *cobra.Command, expectedFlags ...string) {
	t.Helper()
	for _, flagName := range expectedFlags {
		flag := cmd.Flags().Lookup(flagName)
		assert.NotNil(t, flag, "Flag '%s' should exist", flagName)
	}
}

// AssertFileOperations provides common file operation test patterns.
type FileOperationTest struct {
	Name        string
	Setup       func(t *testing.T) string // Returns temp dir path
	Execute     func(t *testing.T, tempDir string) error
	Verify      func(t *testing.T, tempDir string, err error)
	Cleanup     func(t *testing.T, tempDir string)
}

// RunFileOperationTest executes a file operation test with common setup/cleanup.
func RunFileOperationTest(t *testing.T, test FileOperationTest) {
	t.Helper()
	t.Run(test.Name, func(t *testing.T) {
		// Setup
		tempDir := ""
		if test.Setup != nil {
			tempDir = test.Setup(t)
		}

		// Cleanup
		if test.Cleanup != nil {
			defer test.Cleanup(t, tempDir)
		}

		// Execute
		var err error
		if test.Execute != nil {
			err = test.Execute(t, tempDir)
		}

		// Verify
		if test.Verify != nil {
			test.Verify(t, tempDir, err)
		}
	})
}

// AssertConfigContent verifies common configuration content patterns.
func AssertConfigContent(t *testing.T, content []byte, expectedFields ...string) {
	t.Helper()
	contentStr := string(content)
	assert.NotEmpty(t, contentStr, "Config content should not be empty")
	
	for _, field := range expectedFields {
		assert.Contains(t, contentStr, field, "Config should contain field: %s", field)
	}
}

// CommonTestPaths provides standard test paths for different services.
type CommonTestPaths struct {
	TempDir    string
	ConfigPath string
	StorePath  string
}

// NewCommonTestPaths creates standard test paths in a temporary directory.
func NewCommonTestPaths(t *testing.T, serviceName string) CommonTestPaths {
	t.Helper()
	tempDir := t.TempDir()
	return CommonTestPaths{
		TempDir:    tempDir,
		ConfigPath: tempDir + "/" + serviceName + "_config",
		StorePath:  tempDir + "/" + serviceName + "_store",
	}
}

// AssertPathExists verifies that a path exists.
func AssertPathExists(t *testing.T, path string) {
	t.Helper()
	assert.FileExists(t, path, "Path should exist: %s", path)
}

// AssertPathNotExists verifies that a path does not exist.
func AssertPathNotExists(t *testing.T, path string) {
	t.Helper()
	assert.NoFileExists(t, path, "Path should not exist: %s", path)
}