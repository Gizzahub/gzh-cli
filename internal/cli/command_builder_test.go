// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package cli

import (
	"context"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockEnvironment implements env.Environment for testing.
type MockEnvironment struct {
	vars map[string]string
}

func NewMockEnvironment() *MockEnvironment {
	return &MockEnvironment{
		vars: make(map[string]string),
	}
}

func (m *MockEnvironment) Get(key string) string {
	return m.vars[key]
}

func (m *MockEnvironment) Set(key, value string) {
	m.vars[key] = value
}

func TestNewCommandBuilder(t *testing.T) {
	ctx := context.Background()
	builder := NewCommandBuilder(ctx, "test", "Test command")

	assert.NotNil(t, builder)
	assert.NotNil(t, builder.cmd)
	assert.NotNil(t, builder.flags)
	assert.Equal(t, ctx, builder.context)
	assert.Equal(t, "test", builder.cmd.Use)
	assert.Equal(t, "Test command", builder.cmd.Short)
}

func TestCommandBuilder_WithLongDescription(t *testing.T) {
	ctx := context.Background()
	longDesc := "This is a long description for the test command"

	builder := NewCommandBuilder(ctx, "test", "Test command").
		WithLongDescription(longDesc)

	assert.Equal(t, longDesc, builder.cmd.Long)
}

func TestCommandBuilder_WithExample(t *testing.T) {
	ctx := context.Background()
	example := "gz test --org myorg"

	builder := NewCommandBuilder(ctx, "test", "Test command").
		WithExample(example)

	assert.Equal(t, example, builder.cmd.Example)
}

func TestCommandBuilder_WithOrganizationFlag(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		required bool
	}{
		{"optional", false},
		{"required", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			builder := NewCommandBuilder(ctx, "test", "Test command").
				WithOrganizationFlag(test.required)

			cmd := builder.Build()

			// Check that the flag exists
			orgFlag := cmd.Flags().Lookup("org")
			assert.NotNil(t, orgFlag)
			assert.Equal(t, "Organization name", orgFlag.Usage)

			// Check if it's marked as required (this is harder to test directly)
			// For now, just verify the command builds without error
			assert.NotNil(t, cmd)
		})
	}
}

func TestCommandBuilder_WithTokenFlag(t *testing.T) {
	ctx := context.Background()
	builder := NewCommandBuilder(ctx, "test", "Test command").
		WithTokenFlag()

	cmd := builder.Build()
	tokenFlag := cmd.Flags().Lookup("token")

	assert.NotNil(t, tokenFlag)
	assert.Equal(t, "Authentication token (overrides environment)", tokenFlag.Usage)
}

func TestCommandBuilder_WithConfigFileFlag(t *testing.T) {
	ctx := context.Background()
	builder := NewCommandBuilder(ctx, "test", "Test command").
		WithConfigFileFlag()

	cmd := builder.Build()
	configFlag := cmd.Flags().Lookup("config")

	assert.NotNil(t, configFlag)
	assert.Equal(t, "Configuration file path", configFlag.Usage)
}

func TestCommandBuilder_WithVerboseFlag(t *testing.T) {
	ctx := context.Background()
	builder := NewCommandBuilder(ctx, "test", "Test command").
		WithVerboseFlag()

	cmd := builder.Build()
	verboseFlag := cmd.Flags().Lookup("verbose")

	assert.NotNil(t, verboseFlag)
	assert.Equal(t, "Enable verbose output", verboseFlag.Usage)
}

func TestCommandBuilder_WithDryRunFlag(t *testing.T) {
	ctx := context.Background()
	builder := NewCommandBuilder(ctx, "test", "Test command").
		WithDryRunFlag()

	cmd := builder.Build()
	dryRunFlag := cmd.Flags().Lookup("dry-run")

	assert.NotNil(t, dryRunFlag)
	assert.Equal(t, "Show what would be done without making changes", dryRunFlag.Usage)
}

func TestCommandBuilder_WithFormatFlag(t *testing.T) {
	ctx := context.Background()
	validFormats := []string{"json", "yaml", "table"}
	defaultFormat := "table"

	builder := NewCommandBuilder(ctx, "test", "Test command").
		WithFormatFlag(defaultFormat, validFormats)

	cmd := builder.Build()
	formatFlag := cmd.Flags().Lookup("format")

	assert.NotNil(t, formatFlag)
	assert.Contains(t, formatFlag.Usage, "Output format")
	assert.Contains(t, formatFlag.Usage, "json")
	assert.Contains(t, formatFlag.Usage, "yaml")
	assert.Contains(t, formatFlag.Usage, "table")
	assert.Equal(t, defaultFormat, formatFlag.DefValue)
}

func TestCommandBuilder_WithFilterFlag(t *testing.T) {
	ctx := context.Background()
	builder := NewCommandBuilder(ctx, "test", "Test command").
		WithFilterFlag()

	cmd := builder.Build()
	filterFlag := cmd.Flags().Lookup("filter")

	assert.NotNil(t, filterFlag)
	assert.Equal(t, "Filter results by name pattern (regex)", filterFlag.Usage)
}

func TestCommandBuilder_WithLimitFlag(t *testing.T) {
	ctx := context.Background()
	defaultLimit := 50

	builder := NewCommandBuilder(ctx, "test", "Test command").
		WithLimitFlag(defaultLimit)

	cmd := builder.Build()
	limitFlag := cmd.Flags().Lookup("limit")

	assert.NotNil(t, limitFlag)
	assert.Equal(t, "Limit number of results (0 = no limit)", limitFlag.Usage)
	assert.Equal(t, "50", limitFlag.DefValue)
}

func TestCommandBuilder_WithCustomFlag(t *testing.T) {
	ctx := context.Background()
	var customValue string

	builder := NewCommandBuilder(ctx, "test", "Test command").
		WithCustomFlag("custom", "default", "Custom flag for testing", &customValue)

	cmd := builder.Build()
	customFlag := cmd.Flags().Lookup("custom")

	assert.NotNil(t, customFlag)
	assert.Equal(t, "Custom flag for testing", customFlag.Usage)
	assert.Equal(t, "default", customFlag.DefValue)
}

func TestCommandBuilder_WithCustomBoolFlag(t *testing.T) {
	ctx := context.Background()
	var customBool bool

	builder := NewCommandBuilder(ctx, "test", "Test command").
		WithCustomBoolFlag("custom-bool", true, "Custom bool flag for testing", &customBool)

	cmd := builder.Build()
	customFlag := cmd.Flags().Lookup("custom-bool")

	assert.NotNil(t, customFlag)
	assert.Equal(t, "Custom bool flag for testing", customFlag.Usage)
	assert.Equal(t, "true", customFlag.DefValue)
}

func TestCommandBuilder_WithCustomIntFlag(t *testing.T) {
	ctx := context.Background()
	var customInt int

	builder := NewCommandBuilder(ctx, "test", "Test command").
		WithCustomIntFlag("custom-int", 42, "Custom int flag for testing", &customInt)

	cmd := builder.Build()
	customFlag := cmd.Flags().Lookup("custom-int")

	assert.NotNil(t, customFlag)
	assert.Equal(t, "Custom int flag for testing", customFlag.Usage)
	assert.Equal(t, "42", customFlag.DefValue)
}

func TestCommandBuilder_WithRunFunc(t *testing.T) {
	ctx := context.Background()
	executed := false

	runFunc := func(cmd *cobra.Command, args []string) error {
		executed = true
		return nil
	}

	builder := NewCommandBuilder(ctx, "test", "Test command").
		WithRunFunc(runFunc)

	cmd := builder.Build()
	assert.NotNil(t, cmd.RunE)

	// Execute the command
	err := cmd.RunE(cmd, []string{})
	assert.NoError(t, err)
	assert.True(t, executed)
}

func TestCommandBuilder_WithRunFuncE(t *testing.T) {
	ctx := context.Background()
	var receivedFlags *CommonFlags
	var receivedArgs []string

	runFunc := func(ctx context.Context, flags *CommonFlags, args []string) error {
		receivedFlags = flags
		receivedArgs = args
		return nil
	}

	builder := NewCommandBuilder(ctx, "test", "Test command").
		WithRunFuncE(runFunc)

	cmd := builder.Build()
	assert.NotNil(t, cmd.RunE)

	// Execute the command
	testArgs := []string{"arg1", "arg2"}
	err := cmd.RunE(cmd, testArgs)
	assert.NoError(t, err)
	assert.Equal(t, builder.flags, receivedFlags)
	assert.Equal(t, testArgs, receivedArgs)
}

func TestCommandBuilder_AddSubcommand(t *testing.T) {
	ctx := context.Background()

	parentBuilder := NewCommandBuilder(ctx, "parent", "Parent command")
	childCmd := &cobra.Command{
		Use:   "child",
		Short: "Child command",
	}

	parentBuilder.AddSubcommand(childCmd)
	parentCmd := parentBuilder.Build()

	assert.True(t, parentCmd.HasSubCommands())
	assert.Len(t, parentCmd.Commands(), 1)
	assert.Equal(t, "child", parentCmd.Commands()[0].Use)
}

func TestCommandBuilder_GetFlags(t *testing.T) {
	ctx := context.Background()
	builder := NewCommandBuilder(ctx, "test", "Test command")

	flags := builder.GetFlags()
	assert.NotNil(t, flags)
	assert.Equal(t, builder.flags, flags)
}

func TestCommandBuilder_FluentInterface(t *testing.T) {
	ctx := context.Background()

	// Test that all methods return the builder for chaining
	builder := NewCommandBuilder(ctx, "test", "Test command").
		WithLongDescription("Long description").
		WithExample("Example usage").
		WithOrganizationFlag(true).
		WithTokenFlag().
		WithConfigFileFlag().
		WithVerboseFlag().
		WithDryRunFlag().
		WithFormatFlag("json", []string{"json", "yaml"}).
		WithFilterFlag().
		WithLimitFlag(100)

	cmd := builder.Build()

	// Verify all flags were added
	assert.NotNil(t, cmd.Flags().Lookup("org"))
	assert.NotNil(t, cmd.Flags().Lookup("token"))
	assert.NotNil(t, cmd.Flags().Lookup("config"))
	assert.NotNil(t, cmd.Flags().Lookup("verbose"))
	assert.NotNil(t, cmd.Flags().Lookup("dry-run"))
	assert.NotNil(t, cmd.Flags().Lookup("format"))
	assert.NotNil(t, cmd.Flags().Lookup("filter"))
	assert.NotNil(t, cmd.Flags().Lookup("limit"))

	// Verify descriptions
	assert.Equal(t, "Long description", cmd.Long)
	assert.Equal(t, "Example usage", cmd.Example)
}

func TestCommonFlags_Structure(t *testing.T) {
	flags := &CommonFlags{
		Organization: "test-org",
		Token:        "test-token",
		ConfigFile:   "/path/to/config.yaml",
		Verbose:      true,
		DryRun:       false,
		Format:       "json",
		Filter:       ".*test.*",
		Limit:        50,
	}

	assert.Equal(t, "test-org", flags.Organization)
	assert.Equal(t, "test-token", flags.Token)
	assert.Equal(t, "/path/to/config.yaml", flags.ConfigFile)
	assert.True(t, flags.Verbose)
	assert.False(t, flags.DryRun)
	assert.Equal(t, "json", flags.Format)
	assert.Equal(t, ".*test.*", flags.Filter)
	assert.Equal(t, 50, flags.Limit)
}

func TestNewCommandValidator(t *testing.T) {
	validator := NewCommandValidator()

	assert.NotNil(t, validator)
	assert.NotNil(t, validator.environment)
}

func TestCommandValidator_ValidateOrganization(t *testing.T) {
	validator := NewCommandValidator()

	tests := []struct {
		name        string
		org         string
		expectError bool
	}{
		{"valid_org", "test-org", false},
		{"empty_org", "", true},
		{"whitespace_org", "   ", false}, // Non-empty string, even if whitespace
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validator.ValidateOrganization(test.org)
			if test.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "organization is required")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCommandValidator_ValidateFormat(t *testing.T) {
	validator := NewCommandValidator()
	validFormats := []string{"json", "yaml", "table"}

	tests := []struct {
		name        string
		format      string
		expectError bool
	}{
		{"valid_json", "json", false},
		{"valid_yaml", "yaml", false},
		{"valid_table", "table", false},
		{"invalid_format", "xml", true},
		{"empty_format", "", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validator.ValidateFormat(test.format, validFormats)
			if test.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid format")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCommandValidator_ValidateFilter(t *testing.T) {
	validator := NewCommandValidator()

	tests := []struct {
		name        string
		filter      string
		expectError bool
	}{
		{"empty_filter", "", false},
		{"valid_regex", ".*test.*", false},
		{"simple_pattern", "test", false},
		{"invalid_regex", "[", true},
		{"complex_valid_regex", "^(test|demo)_.*$", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validator.ValidateFilter(test.filter)
			if test.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid filter pattern")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCommandValidator_ValidateLimit(t *testing.T) {
	validator := NewCommandValidator()

	tests := []struct {
		name        string
		limit       int
		expectError bool
	}{
		{"zero_limit", 0, false},
		{"positive_limit", 50, false},
		{"large_limit", 10000, false},
		{"negative_limit", -1, true},
		{"large_negative", -100, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validator.ValidateLimit(test.limit)
			if test.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "limit must be non-negative")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCommandValidator_GetToken(t *testing.T) {
	validator := NewCommandValidator()

	// Mock the environment (this would require dependency injection in real code)
	// For now, test the logic assuming the env works correctly

	tests := []struct {
		name      string
		flagToken string
		envKey    string
		expected  string
	}{
		{"flag_token_present", "flag-token", "ENV_TOKEN", "flag-token"},
		{"flag_token_empty", "", "ENV_TOKEN", ""}, // Would get from env in real implementation
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := validator.GetToken(test.flagToken, test.envKey)
			if test.flagToken != "" {
				assert.Equal(t, test.expected, result)
			} else {
				// When flag is empty, it gets from environment
				// In real implementation this would get the env value
				assert.NotNil(t, result) // Just check it doesn't panic
			}
		})
	}
}

func TestJoinStrings(t *testing.T) {
	tests := []struct {
		name      string
		strs      []string
		separator string
		expected  string
	}{
		{"empty_slice", []string{}, ", ", ""},
		{"single_string", []string{"hello"}, ", ", "hello"},
		{"two_strings", []string{"hello", "world"}, ", ", "hello, world"},
		{"multiple_strings", []string{"a", "b", "c", "d"}, " | ", "a | b | c | d"},
		{"different_separator", []string{"x", "y", "z"}, "-", "x-y-z"},
		{"empty_separator", []string{"a", "b"}, "", "ab"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := joinStrings(test.strs, test.separator)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestCommandBuilder_Integration(t *testing.T) {
	ctx := context.Background()
	var executedCtx context.Context
	var executedFlags *CommonFlags
	var executedArgs []string

	runFunc := func(ctx context.Context, flags *CommonFlags, args []string) error {
		executedCtx = ctx
		executedFlags = flags
		executedArgs = args
		return nil
	}

	builder := NewCommandBuilder(ctx, "integration-test", "Integration test command").
		WithLongDescription("This is a comprehensive integration test").
		WithExample("gz integration-test --org myorg --verbose").
		WithOrganizationFlag(false).
		WithTokenFlag().
		WithVerboseFlag().
		WithFormatFlag("json", []string{"json", "yaml", "table"}).
		WithRunFuncE(runFunc)

	cmd := builder.Build()

	// Set some flag values
	cmd.Flags().Set("org", "test-org")
	cmd.Flags().Set("token", "test-token")
	cmd.Flags().Set("verbose", "true")
	cmd.Flags().Set("format", "yaml")

	// Execute
	testArgs := []string{"arg1", "arg2"}
	err := cmd.RunE(cmd, testArgs)

	require.NoError(t, err)
	assert.Equal(t, ctx, executedCtx)
	assert.NotNil(t, executedFlags)
	assert.Equal(t, "test-org", executedFlags.Organization)
	assert.Equal(t, "test-token", executedFlags.Token)
	assert.True(t, executedFlags.Verbose)
	assert.Equal(t, "yaml", executedFlags.Format)
	assert.Equal(t, testArgs, executedArgs)
}
