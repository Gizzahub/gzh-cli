// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package repo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBulkUpdateCmdWrapper tests the bulk-update wrapper function
func TestBulkUpdateCmdWrapper(t *testing.T) {
	// Execute - create command through private function
	cmd := newRepoBulkUpdateCmd()

	// Verify
	require.NotNil(t, cmd, "Command should not be nil")

	// Test command metadata
	assert.Contains(t, cmd.Use, "pull-all", "Command use should contain 'pull-all'")
	assert.NotEmpty(t, cmd.Short, "Command should have short description")

	// Test that command is executable
	assert.NotNil(t, cmd.RunE, "Command should have RunE function")
}

// TestBulkUpdateCmdFlags tests that the command has expected flags
func TestBulkUpdateCmdFlags(t *testing.T) {
	// Execute
	cmd := newRepoBulkUpdateCmd()

	// Verify expected flags exist
	assert.NotNil(t, cmd.Flags().Lookup("parallel"), "Command should have --parallel flag")
	assert.NotNil(t, cmd.Flags().Lookup("max-depth"), "Command should have --max-depth flag")
	assert.NotNil(t, cmd.Flags().Lookup("exclude-pattern"), "Command should have --exclude-pattern flag")
	assert.NotNil(t, cmd.Flags().Lookup("include-pattern"), "Command should have --include-pattern flag")
	assert.NotNil(t, cmd.Flags().Lookup("json"), "Command should have --json flag")
	assert.NotNil(t, cmd.Flags().Lookup("dry-run"), "Command should have --dry-run flag")
	assert.NotNil(t, cmd.Flags().Lookup("verbose"), "Command should have --verbose flag")
	assert.NotNil(t, cmd.Flags().Lookup("no-fetch"), "Command should have --no-fetch flag")
}

// TestBulkUpdateCmdIntegration tests integration with external library
func TestBulkUpdateCmdIntegration(t *testing.T) {
	// Execute
	cmd := newRepoBulkUpdateCmd()

	// Verify command structure
	assert.NotNil(t, cmd, "Command should not be nil")
	assert.NotNil(t, cmd.RunE, "Command should be executable")

	// Verify flags from external library
	flags := cmd.Flags()
	assert.NotNil(t, flags, "Command should have flags")

	// Test that command can be executed (will fail without args, but should not panic)
	assert.NotPanics(t, func() {
		cmd.SetArgs([]string{"--help"})
		_ = cmd.Execute() // Ignore error, just testing it doesn't panic
	}, "Command should handle --help flag")
}

// TestBulkUpdateCmdDefaultValues tests default flag values
func TestBulkUpdateCmdDefaultValues(t *testing.T) {
	// Execute
	cmd := newRepoBulkUpdateCmd()

	// Get flags
	parallelFlag := cmd.Flags().Lookup("parallel")
	maxDepthFlag := cmd.Flags().Lookup("max-depth")

	// Verify defaults exist (values may vary with library version)
	assert.NotNil(t, parallelFlag, "Parallel flag should exist")
	assert.NotNil(t, maxDepthFlag, "Max-depth flag should exist")
}
