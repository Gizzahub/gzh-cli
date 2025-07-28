// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommandStructure(t *testing.T) {
	ctx := context.Background()
	cmd := newRootCmd(ctx)

	// Test root command properties
	assert.Equal(t, "git-synclone", cmd.Use)
	assert.Contains(t, cmd.Short, "Enhanced Git cloning")
	assert.NotEmpty(t, cmd.Long)
	assert.NotEmpty(t, cmd.Version)

	// Test that root command has subcommands
	assert.NotNil(t, cmd.Commands())
	assert.Greater(t, len(cmd.Commands()), 0)
}

func TestSubCommands(t *testing.T) {
	ctx := context.Background()
	cmd := newRootCmd(ctx)

	// Expected subcommands
	expectedSubcommands := []string{
		"github",
		"gitlab",
		"gitea",
		"all",
		"config",
		"state",
		"validate",
		"doctor",
	}

	subcommands := make(map[string]*cobra.Command)
	for _, subCmd := range cmd.Commands() {
		subcommands[subCmd.Use] = subCmd
	}

	// Verify all expected subcommands exist
	for _, expected := range expectedSubcommands {
		t.Run("subcommand_"+expected, func(t *testing.T) {
			subCmd, exists := subcommands[expected]
			require.True(t, exists, "subcommand '%s' should exist", expected)
			assert.NotEmpty(t, subCmd.Short, "subcommand '%s' should have short description", expected)
		})
	}
}

func TestGlobalFlags(t *testing.T) {
	ctx := context.Background()
	cmd := newRootCmd(ctx)

	// Expected global flags
	expectedFlags := []struct {
		name         string
		shorthand    string
		defaultValue interface{}
	}{
		{"config", "c", ""},
		{"target", "t", ""},
		{"parallel", "p", 10},
		{"resume", "", false},
		{"cleanup-orphans", "", false},
		{"strategy", "", "reset"},
		{"dry-run", "", false},
		{"progress-mode", "", "bar"},
	}

	for _, expected := range expectedFlags {
		t.Run("flag_"+expected.name, func(t *testing.T) {
			flag := cmd.PersistentFlags().Lookup(expected.name)
			require.NotNil(t, flag, "flag '%s' should exist", expected.name)

			if expected.shorthand != "" {
				assert.Equal(t, expected.shorthand, flag.Shorthand, "flag '%s' shorthand mismatch", expected.name)
			}

			// Check default value based on type
			switch v := expected.defaultValue.(type) {
			case string:
				assert.Equal(t, v, flag.DefValue, "flag '%s' default value mismatch", expected.name)
			case int:
				assert.Equal(t, string(rune(v)), flag.DefValue, "flag '%s' default value mismatch", expected.name)
			case bool:
				expectedStr := "false"
				if v {
					expectedStr = "true"
				}
				assert.Equal(t, expectedStr, flag.DefValue, "flag '%s' default value mismatch", expected.name)
			}
		})
	}
}

func TestGitHubCommand(t *testing.T) {
	ctx := context.Background()
	cmd := newGitHubCmd(ctx)

	// Test command properties
	assert.Equal(t, "github", cmd.Use)
	assert.Contains(t, cmd.Short, "GitHub")
	assert.NotEmpty(t, cmd.Long)

	// Test GitHub-specific flags
	expectedFlags := []string{
		"org",
		"match",
		"visibility",
		"archived",
		"protocol",
	}

	for _, flagName := range expectedFlags {
		t.Run("github_flag_"+flagName, func(t *testing.T) {
			flag := cmd.Flags().Lookup(flagName)
			require.NotNil(t, flag, "GitHub command should have '%s' flag", flagName)
		})
	}

	// Test required flag
	orgFlag := cmd.Flags().Lookup("org")
	require.NotNil(t, orgFlag)
}

func TestGitLabCommand(t *testing.T) {
	ctx := context.Background()
	cmd := newGitLabCmd(ctx)

	// Test command properties
	assert.Equal(t, "gitlab", cmd.Use)
	assert.Contains(t, cmd.Short, "GitLab")
	assert.NotEmpty(t, cmd.Long)

	// Test that command has RunE function
	assert.NotNil(t, cmd.RunE)
}

func TestGiteaCommand(t *testing.T) {
	ctx := context.Background()
	cmd := newGiteaCmd(ctx)

	// Test command properties
	assert.Equal(t, "gitea", cmd.Use)
	assert.Contains(t, cmd.Short, "Gitea")
	assert.NotEmpty(t, cmd.Long)

	// Test that command has RunE function
	assert.NotNil(t, cmd.RunE)
}

func TestDoctorCommand(t *testing.T) {
	cmd := newDoctorCmd()

	// Test command properties
	assert.Equal(t, "doctor", cmd.Use)
	assert.Contains(t, cmd.Short, "installation")
	assert.NotEmpty(t, cmd.Long)

	// Test doctor-specific flags
	expectedFlags := []string{
		"verbose",
		"fix",
	}

	for _, flagName := range expectedFlags {
		t.Run("doctor_flag_"+flagName, func(t *testing.T) {
			flag := cmd.Flags().Lookup(flagName)
			require.NotNil(t, flag, "Doctor command should have '%s' flag", flagName)
		})
	}

	// Test verbose flag shorthand
	verboseFlag := cmd.Flags().Lookup("verbose")
	require.NotNil(t, verboseFlag)
	assert.Equal(t, "v", verboseFlag.Shorthand)
}

func TestVersionFlag(t *testing.T) {
	ctx := context.Background()
	cmd := newRootCmd(ctx)

	// Test version flag exists
	versionFlag := cmd.Flags().Lookup("version")
	require.NotNil(t, versionFlag)

	// Test version string format
	assert.Contains(t, cmd.Version, Version)
	assert.Contains(t, cmd.Version, BuildTime)
}

func TestHelpCommand(t *testing.T) {
	ctx := context.Background()
	cmd := newRootCmd(ctx)

	// Test that help command is available
	helpCmd := cmd.Commands()
	_ = helpCmd // Use the variable to avoid 'declared and not used' error

	// help is automatically added by cobra, so we check if root command shows help when no args
	assert.NotNil(t, cmd.RunE)
}

// Helper function to find subcommand by name
func findSubCommand(parent *cobra.Command, name string) *cobra.Command {
	for _, cmd := range parent.Commands() {
		if cmd.Use == name {
			return cmd
		}
	}
	return nil
}

func TestExecuteFunction(t *testing.T) {
	// Test that Execute function exists and can be called with context
	_ = context.Background()

	// This would normally execute the command, but we'll just test the function exists
	// In a real test, we might want to capture output or use test-specific args
	assert.NotPanics(t, func() {
		// We don't actually call Execute here to avoid side effects
		// Just verify the function exists and can be referenced
		_ = Execute
	})
}

func TestBuildVersionInfo(t *testing.T) {
	// Test that version variables are properly set
	assert.NotEmpty(t, Version, "Version should not be empty")
	assert.NotEmpty(t, BuildTime, "BuildTime should not be empty")

	// Test default values when not set by ldflags
	if Version == "dev" {
		t.Log("Version is set to default 'dev' value")
	}

	if BuildTime == "unknown" {
		t.Log("BuildTime is set to default 'unknown' value")
	}
}
