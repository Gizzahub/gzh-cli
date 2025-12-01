// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package repo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCloneOrUpdateCmdWrapper tests the clone-or-update wrapper function
func TestCloneOrUpdateCmdWrapper(t *testing.T) {
	// Execute - create command through private function
	cmd := newRepoCloneOrUpdateCmd()

	// Verify
	require.NotNil(t, cmd, "Command should not be nil")

	// Test command metadata
	assert.Contains(t, cmd.Use, "clone-or-update", "Command use should contain 'clone-or-update'")
	assert.NotEmpty(t, cmd.Short, "Command should have short description")

	// Test that command is executable
	assert.NotNil(t, cmd.RunE, "Command should have RunE function")
}

// TestCloneOrUpdateCmdFlags tests that the command has expected flags
func TestCloneOrUpdateCmdFlags(t *testing.T) {
	// Execute
	cmd := newRepoCloneOrUpdateCmd()

	// Verify expected flags exist
	assert.NotNil(t, cmd.Flags().Lookup("strategy"), "Command should have --strategy flag")
	assert.NotNil(t, cmd.Flags().Lookup("branch"), "Command should have --branch flag")
	assert.NotNil(t, cmd.Flags().Lookup("depth"), "Command should have --depth flag")
}

// TestCloneOrUpdateCmdIntegration tests integration with external library
func TestCloneOrUpdateCmdIntegration(t *testing.T) {
	// Execute
	cmd := newRepoCloneOrUpdateCmd()

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
