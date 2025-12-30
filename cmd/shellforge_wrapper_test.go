// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package cmd

import (
	"context"
	"testing"

	"github.com/gizzahub/gzh-cli/internal/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewShellforgeCmd tests the NewShellforgeCmd wrapper function
func TestNewShellforgeCmd(t *testing.T) {
	// Setup
	ctx := context.Background()
	appCtx := &app.AppContext{}

	// Execute
	cmd := NewShellforgeCmd(ctx, appCtx)

	// Verify
	require.NotNil(t, cmd, "Command should not be nil")

	// Test command metadata
	assert.Equal(t, "shellforge", cmd.Use, "Command use should be 'shellforge'")
	assert.Equal(t, "Build tool for modular shell configurations", cmd.Short, "Command short description should match")
	assert.NotEmpty(t, cmd.Long, "Command should have long description")
	assert.Contains(t, cmd.Long, "modular shell", "Long description should mention modular shell")
	assert.Contains(t, cmd.Long, "dependency resolution", "Long description should mention dependency resolution")

	// Test that command has subcommands from external library
	assert.True(t, cmd.HasSubCommands(), "Command should have subcommands from external library")

	// Test that command is executable (has Run or subcommands)
	assert.True(t, cmd.HasSubCommands() || cmd.RunE != nil || cmd.Run != nil,
		"Command should be executable or have subcommands")
}

// TestShellforgeCmdProvider tests the shellforgeCmdProvider implementation
func TestShellforgeCmdProvider(t *testing.T) {
	// Setup
	appCtx := &app.AppContext{}
	provider := shellforgeCmdProvider{appCtx: appCtx}

	// Execute
	cmd := provider.Command()

	// Verify
	require.NotNil(t, cmd, "Provider should return a valid command")
	assert.Equal(t, "shellforge", cmd.Use, "Provider should return shellforge command")
}

// TestRegisterShellforgeCmd tests command registration
func TestRegisterShellforgeCmd(t *testing.T) {
	// This test verifies that RegisterShellforgeCmd doesn't panic
	// Actual registration testing would require mocking the registry

	// Setup
	appCtx := &app.AppContext{}

	// Execute - should not panic
	assert.NotPanics(t, func() {
		RegisterShellforgeCmd(appCtx)
	}, "RegisterShellforgeCmd should not panic")
}

// TestShellforgeCmdIntegration tests the integration with the external library
func TestShellforgeCmdIntegration(t *testing.T) {
	// Setup
	ctx := context.Background()
	appCtx := &app.AppContext{}
	cmd := NewShellforgeCmd(ctx, appCtx)

	// Verify command structure
	assert.NotNil(t, cmd, "Command should not be nil")
	assert.True(t, cmd.HasSubCommands(), "Command should have subcommands")

	// Verify subcommands exist
	subcommands := cmd.Commands()
	subcommandNames := make([]string, 0, len(subcommands))
	for _, subcmd := range subcommands {
		subcommandNames = append(subcommandNames, subcmd.Use)
	}

	// Check that we have the expected subcommands
	assert.NotEmpty(t, subcommandNames, "Should have subcommands from external library")

	// Verify key subcommands are present
	for _, expected := range []string{"build", "validate", "list"} {
		assert.Contains(t, subcommandNames, expected, "Should have '%s' subcommand", expected)
	}
}

// TestShellforgeCmdContextPropagation tests that context is properly handled
func TestShellforgeCmdContextPropagation(t *testing.T) {
	// Setup
	ctx := context.Background()
	appCtx := &app.AppContext{}

	// Execute with context
	cmd := NewShellforgeCmd(ctx, appCtx)

	// Verify command was created successfully (context handling is internal)
	require.NotNil(t, cmd, "Command should be created even with context")
	assert.Equal(t, "shellforge", cmd.Use, "Command should be properly configured")
}
