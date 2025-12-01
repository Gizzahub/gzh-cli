// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

//go:build pm_external
// +build pm_external

package cmd

import (
	"context"
	"testing"

	"github.com/Gizzahub/gzh-cli/internal/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewPMCmd tests the NewPMCmd wrapper function
func TestNewPMCmd(t *testing.T) {
	// Setup
	ctx := context.Background()
	appCtx := &app.AppContext{}

	// Execute
	cmd := NewPMCmd(ctx, appCtx)

	// Verify
	require.NotNil(t, cmd, "Command should not be nil")

	// Test command metadata
	assert.Equal(t, "pm", cmd.Use, "Command use should be 'pm'")
	assert.Equal(t, "Package manager operations", cmd.Short, "Command short description should match")
	assert.NotEmpty(t, cmd.Long, "Command should have long description")
	assert.Contains(t, cmd.Long, "package managers", "Long description should mention package managers")
	assert.Contains(t, cmd.Long, "brew", "Long description should mention specific package managers")

	// Test that command has subcommands from external library
	assert.True(t, cmd.HasSubCommands(), "Command should have subcommands from external library")

	// Test that command is executable (has Run or subcommands)
	assert.True(t, cmd.HasSubCommands() || cmd.RunE != nil || cmd.Run != nil,
		"Command should be executable or have subcommands")
}

// TestPMCmdProvider tests the pmCmdProvider implementation
func TestPMCmdProvider(t *testing.T) {
	// Setup
	appCtx := &app.AppContext{}
	provider := pmCmdProvider{appCtx: appCtx}

	// Execute
	cmd := provider.Command()

	// Verify
	require.NotNil(t, cmd, "Provider should return a valid command")
	assert.Equal(t, "pm", cmd.Use, "Provider should return pm command")
}

// TestRegisterPMCmd tests command registration
func TestRegisterPMCmd(t *testing.T) {
	// This test verifies that RegisterPMCmd doesn't panic
	// Actual registration testing would require mocking the registry

	// Setup
	appCtx := &app.AppContext{}

	// Execute - should not panic
	assert.NotPanics(t, func() {
		RegisterPMCmd(appCtx)
	}, "RegisterPMCmd should not panic")
}

// TestPMCmdIntegration tests the integration with the external library
func TestPMCmdIntegration(t *testing.T) {
	// Setup
	ctx := context.Background()
	appCtx := &app.AppContext{}
	cmd := NewPMCmd(ctx, appCtx)

	// Verify command structure
	assert.NotNil(t, cmd, "Command should not be nil")
	assert.True(t, cmd.HasSubCommands(), "Command should have subcommands")

	// Expected subcommands from gzh-cli-package-manager
	// Note: exact subcommands may vary with library version

	// Verify at least some subcommands exist
	subcommands := cmd.Commands()
	subcommandNames := make([]string, len(subcommands))
	for i, subcmd := range subcommands {
		subcommandNames[i] = subcmd.Use
	}

	// Check that we have subcommands (exact list may vary with library version)
	assert.NotEmpty(t, subcommandNames, "Should have subcommands from external library")
}

// TestPMCmdContextPropagation tests that context is properly handled
func TestPMCmdContextPropagation(t *testing.T) {
	// Setup
	ctx := context.Background()
	appCtx := &app.AppContext{}

	// Execute with context
	cmd := NewPMCmd(ctx, appCtx)

	// Verify command was created successfully (context handling is internal)
	require.NotNil(t, cmd, "Command should be created even with context")
	assert.Equal(t, "pm", cmd.Use, "Command should be properly configured")
}
