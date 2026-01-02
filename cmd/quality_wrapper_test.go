// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gizzahub/gzh-cli/internal/app"
)

// TestNewQualityCmd tests the NewQualityCmd wrapper function
func TestNewQualityCmd(t *testing.T) {
	// Setup
	appCtx := &app.AppContext{}

	// Execute
	cmd := NewQualityCmd(appCtx)

	// Verify
	require.NotNil(t, cmd, "Command should not be nil")

	// Test command metadata
	assert.Equal(t, "quality", cmd.Use, "Command use should be 'quality'")
	assert.Equal(t, "통합 코드 품질 도구 (포매팅 + 린팅)", cmd.Short, "Command short description should match")
	assert.Contains(t, cmd.Aliases, "q", "Command should have 'q' alias")
	assert.Contains(t, cmd.Aliases, "qual", "Command should have 'qual' alias")

	// Test that command has subcommands from external library
	assert.True(t, cmd.HasSubCommands(), "Command should have subcommands from external library")

	// Test that command is executable (has Run or subcommands)
	assert.True(t, cmd.HasSubCommands() || cmd.RunE != nil || cmd.Run != nil,
		"Command should be executable or have subcommands")
}

// TestQualityCmdProvider tests the qualityCmdProvider implementation
func TestQualityCmdProvider(t *testing.T) {
	// Setup
	appCtx := &app.AppContext{}
	provider := qualityCmdProvider{appCtx: appCtx}

	// Execute
	cmd := provider.Command()

	// Verify
	require.NotNil(t, cmd, "Provider should return a valid command")
	assert.Equal(t, "quality", cmd.Use, "Provider should return quality command")
}

// TestRegisterQualityCmd tests command registration
func TestRegisterQualityCmd(t *testing.T) {
	// This test verifies that RegisterQualityCmd doesn't panic
	// Actual registration testing would require mocking the registry

	// Setup
	appCtx := &app.AppContext{}

	// Execute - should not panic
	assert.NotPanics(t, func() {
		RegisterQualityCmd(appCtx)
	}, "RegisterQualityCmd should not panic")
}

// TestQualityCmdIntegration tests the integration with the external library
func TestQualityCmdIntegration(t *testing.T) {
	// Setup
	appCtx := &app.AppContext{}
	cmd := NewQualityCmd(appCtx)

	// Verify command structure
	assert.NotNil(t, cmd, "Command should not be nil")
	assert.True(t, cmd.HasSubCommands(), "Command should have subcommands")

	// Verify at least some subcommands exist
	subcommands := cmd.Commands()
	subcommandNames := make([]string, len(subcommands))
	for i, subcmd := range subcommands {
		subcommandNames[i] = subcmd.Use
	}

	// Check that we have subcommands (exact list may vary with library version)
	assert.NotEmpty(t, subcommandNames, "Should have subcommands from external library")
}
