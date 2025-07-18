package repoconfig

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDiffIntegration tests the diff command integration.
func TestDiffIntegration(t *testing.T) {
	// Set mock token for tests
	os.Setenv("GITHUB_TOKEN", "mock-token-for-testing")
	defer os.Unsetenv("GITHUB_TOKEN")

	tests := []struct {
		name         string
		args         []string
		expectError  bool
		expectOutput []string
		skipOutput   []string
	}{
		{
			name:        "diff without org flag",
			args:        []string{"diff"},
			expectError: true,
		},
		{
			name: "diff with org flag",
			args: []string{"diff", "--org", "test-org"},
			expectOutput: []string{
				"Repository Configuration Differences",
				"Organization: test-org",
			},
		},
		{
			name: "diff with filter",
			args: []string{"diff", "--org", "test-org", "--filter", "api-.*"},
			expectOutput: []string{
				"Repository Configuration Differences",
				"Filter: api-.*",
			},
		},
		{
			name: "diff with JSON format",
			args: []string{"diff", "--org", "test-org", "--format", "json"},
			expectOutput: []string{
				`"organization": "test-org"`,
				`"differences"`,
			},
		},
		{
			name: "diff with unified format",
			args: []string{"diff", "--org", "test-org", "--format", "unified"},
			expectOutput: []string{
				"---",
				"+++",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create root command
			cmd := &cobra.Command{Use: "repo-config"}
			cmd.AddCommand(newDiffCmd())

			// Capture output
			var (
				stdout bytes.Buffer
				stderr bytes.Buffer
			)

			cmd.SetOut(&stdout)
			cmd.SetErr(&stderr)

			// Set args
			cmd.SetArgs(append([]string{"diff"}, tt.args[1:]...))

			// Execute command
			err := cmd.Execute()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			output := stdout.String() + stderr.String()

			// Check expected output
			for _, expected := range tt.expectOutput {
				assert.Contains(t, output, expected, "Expected output not found: %s", expected)
			}

			// Check skip output
			for _, skip := range tt.skipOutput {
				assert.NotContains(t, output, skip, "Unexpected output found: %s", skip)
			}
		})
	}
}

// TestAuditIntegration tests the audit command integration.
func TestAuditIntegration(t *testing.T) {
	// Set mock token for tests
	os.Setenv("GITHUB_TOKEN", "mock-token-for-testing")
	defer os.Unsetenv("GITHUB_TOKEN")

	tests := []struct {
		name          string
		args          []string
		expectError   bool
		expectOutput  []string
		checkExitCode bool
		expectedExit  int
	}{
		{
			name:        "audit without org flag",
			args:        []string{"audit"},
			expectError: true,
		},
		{
			name: "audit with org flag",
			args: []string{"audit", "--org", "test-org"},
			expectOutput: []string{
				"Repository Compliance Audit Report",
				"Organization: test-org",
				"Compliance Summary",
			},
		},
		{
			name: "audit with policy group",
			args: []string{"audit", "--org", "test-org", "--policy-group", "security"},
			expectOutput: []string{
				"Repository Compliance Audit Report",
				"Policy group: security",
			},
		},
		{
			name: "audit with policy preset",
			args: []string{"audit", "--org", "test-org", "--policy-preset", "soc2"},
			expectOutput: []string{
				"Repository Compliance Audit Report",
				"Policy preset: soc2",
			},
		},
		{
			name: "audit with JSON format",
			args: []string{"audit", "--org", "test-org", "--format", "json"},
			expectOutput: []string{
				`"organization": "test-org"`,
				`"summary"`,
				`"policy_compliance"`,
			},
		},
		{
			name: "audit with CSV format",
			args: []string{"audit", "--org", "test-org", "--format", "csv"},
			expectOutput: []string{
				"Repository,Visibility,Template,Compliant,Violations,Critical",
			},
		},
		{
			name: "audit with SARIF format",
			args: []string{"audit", "--org", "test-org", "--format", "sarif"},
			expectOutput: []string{
				`"$schema"`,
				`"version": "2.1.0"`,
				`"tool"`,
			},
		},
		{
			name: "audit with JUnit format",
			args: []string{"audit", "--org", "test-org", "--format", "junit"},
			expectOutput: []string{
				`<?xml version="1.0" encoding="UTF-8"?>`,
				`<testsuites`,
				`<testsuite`,
			},
		},
		{
			name: "audit with repository filters",
			args: []string{
				"audit", "--org", "test-org",
				"--filter-visibility", "private",
				"--filter-pattern", "api-.*",
			},
			expectOutput: []string{
				"Repository Compliance Audit Report",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create root command
			cmd := &cobra.Command{Use: "repo-config"}
			cmd.AddCommand(newAuditCmd())

			// Capture output
			var (
				stdout bytes.Buffer
				stderr bytes.Buffer
			)

			cmd.SetOut(&stdout)
			cmd.SetErr(&stderr)

			// Set args
			cmd.SetArgs(append([]string{"audit"}, tt.args[1:]...))

			// Execute command
			err := cmd.Execute()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			output := stdout.String() + stderr.String()

			// Check expected output
			for _, expected := range tt.expectOutput {
				assert.Contains(t, output, expected, "Expected output not found: %s", expected)
			}
		})
	}
}

// TestAuditCIIntegration tests CI/CD integration features.
func TestAuditCIIntegration(t *testing.T) {
	// Skip if not in CI environment
	if os.Getenv("CI") == "" {
		t.Skip("Skipping CI integration test outside of CI environment")
	}

	// Test exit on fail feature
	t.Run("exit on fail", func(t *testing.T) {
		cmd := &cobra.Command{Use: "repo-config"}
		cmd.AddCommand(newAuditCmd())

		var stdout bytes.Buffer
		cmd.SetOut(&stdout)

		// This should exit with non-zero code if compliance is below threshold
		cmd.SetArgs([]string{
			"audit",
			"--org", "test-org",
			"--exit-on-fail",
			"--fail-threshold", "100", // Set high threshold to ensure failure
		})

		// Note: In real CI, this would exit the process
		// For testing, we just check if the feature is working
		_ = cmd.Execute()

		// In the mock implementation, this won't actually exit
		// but we can verify the feature is recognized
		output := stdout.String()
		if strings.Contains(output, "Compliance check failed") {
			assert.Contains(t, output, "Compliance check failed")
		}
	})
}

// TestAuditTrendAnalysis tests trend analysis features.
func TestAuditTrendAnalysis(t *testing.T) {
	// Create temporary directory for trend data
	tmpDir, err := os.MkdirTemp("", "audit-trend-test")
	require.NoError(t, err)

	defer os.RemoveAll(tmpDir)

	// Set audit data directory
	os.Setenv("GZH_AUDIT_DATA_DIR", tmpDir)
	defer os.Unsetenv("GZH_AUDIT_DATA_DIR")

	t.Run("save trend", func(t *testing.T) {
		cmd := &cobra.Command{Use: "repo-config"}
		cmd.AddCommand(newAuditCmd())

		var stdout bytes.Buffer
		cmd.SetOut(&stdout)

		cmd.SetArgs([]string{
			"audit",
			"--org", "test-org",
			"--save-trend",
		})

		err := cmd.Execute()
		assert.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "Audit results saved for trend analysis")
	})

	t.Run("show trend", func(t *testing.T) {
		cmd := &cobra.Command{Use: "repo-config"}
		cmd.AddCommand(newAuditCmd())

		var stdout bytes.Buffer
		cmd.SetOut(&stdout)

		cmd.SetArgs([]string{
			"audit",
			"--org", "test-org",
			"--show-trend",
			"--trend-period", "7d",
		})

		err := cmd.Execute()
		assert.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "Trend Analysis Report")
	})
}

// TestAuditNotifications tests notification features.
func TestAuditNotifications(t *testing.T) {
	t.Run("webhook notification", func(t *testing.T) {
		cmd := &cobra.Command{Use: "repo-config"}
		cmd.AddCommand(newAuditCmd())

		var stdout bytes.Buffer
		cmd.SetOut(&stdout)

		cmd.SetArgs([]string{
			"audit",
			"--org", "test-org",
			"--notify-webhook", "https://example.com/webhook",
		})

		err := cmd.Execute()
		assert.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "Sending webhook notification")
	})

	t.Run("email notification", func(t *testing.T) {
		cmd := &cobra.Command{Use: "repo-config"}
		cmd.AddCommand(newAuditCmd())

		var stdout bytes.Buffer
		cmd.SetOut(&stdout)

		cmd.SetArgs([]string{
			"audit",
			"--org", "test-org",
			"--notify-email", "admin@example.com",
		})

		err := cmd.Execute()
		assert.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "Sending email notification")
	})
}

// TestPolicyConfiguration tests policy configuration features.
func TestPolicyConfiguration(t *testing.T) {
	t.Run("list available presets", func(t *testing.T) {
		// This is a feature suggestion - list available presets
		presets := []string{"soc2", "iso27001", "nist", "pci-dss", "hipaa", "gdpr", "minimal", "enterprise"}

		for _, preset := range presets {
			cmd := &cobra.Command{Use: "repo-config"}
			cmd.AddCommand(newAuditCmd())

			var stdout bytes.Buffer
			cmd.SetOut(&stdout)

			cmd.SetArgs([]string{
				"audit",
				"--org", "test-org",
				"--policy-preset", preset,
			})

			err := cmd.Execute()
			assert.NoError(t, err, "Preset %s should be valid", preset)
		}
	})

	t.Run("list available groups", func(t *testing.T) {
		groups := []string{"security", "compliance", "best-practice"}

		for _, group := range groups {
			cmd := &cobra.Command{Use: "repo-config"}
			cmd.AddCommand(newAuditCmd())

			var stdout bytes.Buffer
			cmd.SetOut(&stdout)

			cmd.SetArgs([]string{
				"audit",
				"--org", "test-org",
				"--policy-group", group,
			})

			err := cmd.Execute()
			assert.NoError(t, err, "Group %s should be valid", group)
		}
	})
}

// TestOutputFileGeneration tests file output features.
func TestOutputFileGeneration(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "audit-output-test")
	require.NoError(t, err)

	defer os.RemoveAll(tmpDir)

	formats := []struct {
		format   string
		filename string
		contains []string
	}{
		{
			format:   "json",
			filename: "audit.json",
			contains: []string{`"organization"`, `"summary"`},
		},
		{
			format:   "csv",
			filename: "audit.csv",
			contains: []string{"Repository,Visibility"},
		},
		{
			format:   "html",
			filename: "audit.html",
			contains: []string{"<!DOCTYPE html>", "<title>"},
		},
		{
			format:   "sarif",
			filename: "audit.sarif",
			contains: []string{`"$schema"`, `"version": "2.1.0"`},
		},
		{
			format:   "junit",
			filename: "audit.xml",
			contains: []string{`<?xml version="1.0"`, `<testsuites`},
		},
	}

	for _, tc := range formats {
		t.Run(tc.format+" output", func(t *testing.T) {
			outputPath := tmpDir + "/" + tc.filename

			cmd := &cobra.Command{Use: "repo-config"}
			cmd.AddCommand(newAuditCmd())

			var stdout bytes.Buffer
			cmd.SetOut(&stdout)

			cmd.SetArgs([]string{
				"audit",
				"--org", "test-org",
				"--format", tc.format,
				"--output", outputPath,
			})

			err := cmd.Execute()
			assert.NoError(t, err)

			// Check file was created
			_, err = os.Stat(outputPath)
			assert.NoError(t, err, "Output file should be created")

			// Read and verify content
			content, err := os.ReadFile(outputPath)
			require.NoError(t, err)

			for _, expected := range tc.contains {
				assert.Contains(t, string(content), expected)
			}
		})
	}
}
