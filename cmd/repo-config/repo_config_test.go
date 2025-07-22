//nolint:testpackage // White-box testing needed for internal function access
package repoconfig

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRepoConfigCmd(t *testing.T) {
	cmd := NewRepoConfigCmd()

	assert.Equal(t, "repo-config", cmd.Use)
	assert.Equal(t, "GitHub repository configuration management", cmd.Short)
	assert.Contains(t, cmd.Long, "infrastructure-as-code")

	// Check that all expected subcommands are present
	expectedSubcommands := []string{"list", "apply", "validate", "diff", "audit", "template"}
	actualSubcommands := make([]string, 0, len(cmd.Commands()))

	for _, subcmd := range cmd.Commands() {
		actualSubcommands = append(actualSubcommands, subcmd.Use)
	}

	for _, expected := range expectedSubcommands {
		assert.Contains(t, actualSubcommands, expected, "Expected subcommand %s not found", expected)
	}
}

func TestGlobalFlags(t *testing.T) {
	cmd := &cobra.Command{
		Use: "test",
	}

	var flags GlobalFlags
	addGlobalFlags(cmd, &flags)

	// Test that all expected flags are present
	expectedFlags := []string{"org", "config", "token", "dry-run", "verbose", "parallel", "timeout"}

	for _, flagName := range expectedFlags {
		flag := cmd.Flags().Lookup(flagName)
		assert.NotNil(t, flag, "Expected flag %s not found", flagName)
	}

	// Test flag types and defaults
	assert.Equal(t, 5, flags.Parallel)
	assert.Equal(t, "30s", flags.Timeout)
	assert.False(t, flags.DryRun)
	assert.False(t, flags.Verbose)
}

func TestListCommand(t *testing.T) {
	cmd := newListCmd()

	assert.Equal(t, "list", cmd.Use)
	assert.Equal(t, "List repositories with current configuration", cmd.Short)
	assert.Contains(t, cmd.Long, "repository details")

	// Test that list-specific flags are present
	expectedFlags := []string{"filter", "format", "show-config", "limit"}

	for _, flagName := range expectedFlags {
		flag := cmd.Flags().Lookup(flagName)
		assert.NotNil(t, flag, "Expected flag %s not found", flagName)
	}
}

func TestApplyCommand(t *testing.T) {
	cmd := newApplyCmd()

	assert.Equal(t, "apply", cmd.Use)
	assert.Equal(t, "Apply repository configuration", cmd.Short)
	assert.Contains(t, cmd.Long, "configuration templates")

	// Test that apply-specific flags are present
	expectedFlags := []string{"filter", "template", "interactive", "force"}

	for _, flagName := range expectedFlags {
		flag := cmd.Flags().Lookup(flagName)
		assert.NotNil(t, flag, "Expected flag %s not found", flagName)
	}
}

func TestValidateCommand(t *testing.T) {
	cmd := newValidateCmd()

	assert.Equal(t, "validate", cmd.Use)
	assert.Equal(t, "Validate repository configuration files", cmd.Short)
	assert.Contains(t, cmd.Long, "schema compliance")

	// Test that validate-specific flags are present
	expectedFlags := []string{"config-file", "strict", "format"}

	for _, flagName := range expectedFlags {
		flag := cmd.Flags().Lookup(flagName)
		assert.NotNil(t, flag, "Expected flag %s not found", flagName)
	}
}

func TestDiffCommand(t *testing.T) {
	cmd := newDiffCmd()

	assert.Equal(t, "diff", cmd.Use)
	assert.Equal(t, "Show differences between current and target configuration", cmd.Short)
	assert.Contains(t, cmd.Long, "compares the current state")

	// Test that diff-specific flags are present
	expectedFlags := []string{"filter", "format", "show-values"}

	for _, flagName := range expectedFlags {
		flag := cmd.Flags().Lookup(flagName)
		assert.NotNil(t, flag, "Expected flag %s not found", flagName)
	}
}

func TestAuditCommand(t *testing.T) {
	cmd := newAuditCmd()

	assert.Equal(t, "audit", cmd.Use)
	assert.Equal(t, "Generate compliance audit report", cmd.Short)
	assert.Contains(t, cmd.Long, "compliance audit report")

	// Test that audit-specific flags are present
	expectedFlags := []string{"format", "output", "detailed", "policy"}

	for _, flagName := range expectedFlags {
		flag := cmd.Flags().Lookup(flagName)
		assert.NotNil(t, flag, "Expected flag %s not found", flagName)
	}
}

func TestTemplateCommand(t *testing.T) {
	cmd := newTemplateCmd()

	assert.Equal(t, "template", cmd.Use)
	assert.Equal(t, "Manage repository configuration templates", cmd.Short)
	assert.Contains(t, cmd.Long, "configuration templates")

	// Check that template subcommands are present
	expectedSubcommands := []string{"list", "show", "validate"}
	actualSubcommands := make([]string, 0, len(cmd.Commands()))

	for _, subcmd := range cmd.Commands() {
		// Extract just the command name (before any spaces/arguments)
		cmdName := strings.Split(subcmd.Use, " ")[0]
		actualSubcommands = append(actualSubcommands, cmdName)
	}

	for _, expected := range expectedSubcommands {
		assert.Contains(t, actualSubcommands, expected, "Expected template subcommand %s not found", expected)
	}
}

func TestRunListCommandMissingOrg(t *testing.T) {
	var flags GlobalFlags

	err := runListCommand(flags, "", "table", false, 0)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "organization is required")
}

func TestRunApplyCommandMissingOrg(t *testing.T) {
	var flags GlobalFlags

	err := runApplyCommand(flags, "", "", false, false)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "organization is required")
}

func TestConfigurationChange(t *testing.T) {
	change := ConfigurationChange{
		Repository:   "test-repo",
		Setting:      "branch_protection.main.required_reviews",
		CurrentValue: "1",
		NewValue:     "2",
		Action:       "update",
	}

	assert.Equal(t, "test-repo", change.Repository)
	assert.Equal(t, "branch_protection.main.required_reviews", change.Setting)
	assert.Equal(t, "1", change.CurrentValue)
	assert.Equal(t, "2", change.NewValue)
	assert.Equal(t, "update", change.Action)
}

func TestGetActionSymbol(t *testing.T) {
	tests := []struct {
		action   string
		expected string
	}{
		{"create", "➕"},
		{"update", "🔄"},
		{"delete", "➖"},
		{"unknown", "📝"},
	}

	for _, test := range tests {
		t.Run(test.action, func(t *testing.T) {
			result := getActionSymbol(test.action)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestGetAffectedRepoCount(t *testing.T) {
	changes := []ConfigurationChange{
		{Repository: "repo1", Action: "update"},
		{Repository: "repo2", Action: "update"},
		{Repository: "repo1", Action: "create"}, // Same repo, should count as 1
	}

	count := getAffectedRepoCount(changes)
	assert.Equal(t, 2, count)
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"this is a very long string", 10, "this is..."},
		{"exactly10", 10, "exactly10"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := truncateString(test.input, test.maxLen)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestValidationResult(t *testing.T) {
	result := ValidationResult{
		Check:    "Test Check",
		Status:   "pass",
		Message:  "Test message",
		Severity: "info",
	}

	assert.Equal(t, "Test Check", result.Check)
	assert.Equal(t, "pass", result.Status)
	assert.Equal(t, "Test message", result.Message)
	assert.Equal(t, "info", result.Severity)
}

func TestGetStatusSymbol(t *testing.T) {
	tests := []struct {
		status   string
		expected string
	}{
		{"pass", "✅"},
		{"warn", "⚠️"},
		{"fail", "❌"},
		{"unknown", "❓"},
	}

	for _, test := range tests {
		t.Run(test.status, func(t *testing.T) {
			result := getStatusSymbol(test.status)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestHasValidationErrors(t *testing.T) {
	tests := []struct {
		name     string
		results  ValidationSummary
		expected bool
	}{
		{
			name: "no errors",
			results: ValidationSummary{
				Valid:  true,
				Errors: 0,
			},
			expected: false,
		},
		{
			name: "has errors",
			results: ValidationSummary{
				Valid:  false,
				Errors: 1,
			},
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := hasValidationErrors(test.results)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestDiscoverConfigFile(t *testing.T) {
	// This test would normally create temporary files to test discovery
	// For now, we'll just test that it returns empty when no files exist
	result := discoverConfigFile()
	assert.Equal(t, "", result)
}

func TestRepoConfigCmdHelp(t *testing.T) {
	cmd := NewRepoConfigCmd()

	// Capture output
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	// Run the command without arguments (should show help)
	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "repository configurations")
	assert.Contains(t, output, "list")
	assert.Contains(t, output, "apply")
	assert.Contains(t, output, "validate")
}

func TestCommandStructureConsistency(t *testing.T) {
	// Test that all commands have consistent structure
	commands := []*cobra.Command{
		newListCmd(),
		newApplyCmd(),
		newValidateCmd(),
		newDiffCmd(),
		newAuditCmd(),
	}

	for _, cmd := range commands {
		t.Run(cmd.Use, func(t *testing.T) {
			// All commands should have global flags
			orgFlag := cmd.Flags().Lookup("org")
			assert.NotNil(t, orgFlag, "Command %s missing --org flag", cmd.Use)

			configFlag := cmd.Flags().Lookup("config")
			assert.NotNil(t, configFlag, "Command %s missing --config flag", cmd.Use)

			verboseFlag := cmd.Flags().Lookup("verbose")
			assert.NotNil(t, verboseFlag, "Command %s missing --verbose flag", cmd.Use)

			// All commands should have non-empty descriptions
			assert.NotEmpty(t, cmd.Short, "Command %s missing short description", cmd.Use)
			assert.NotEmpty(t, cmd.Long, "Command %s missing long description", cmd.Use)

			// All commands should have examples in their long description
			assert.Contains(t, strings.ToLower(cmd.Long), "example",
				"Command %s missing examples in description", cmd.Use)
		})
	}
}
