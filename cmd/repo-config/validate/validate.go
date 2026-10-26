// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package validate

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/gizzahub/gzh-cli/pkg/config"
)

// GlobalFlags represents global flags for all repo-config commands.
type GlobalFlags struct {
	Organization string
	ConfigFile   string
	Token        string
	DryRun       bool
	Verbose      bool
	Parallel     int
	Timeout      string
}

// addGlobalFlags adds common flags to the command.
func addGlobalFlags(cmd *cobra.Command, flags *GlobalFlags) {
	cmd.Flags().StringVar(&flags.Organization, "org", "", "GitHub organization name")
	cmd.Flags().StringVar(&flags.ConfigFile, "config-file", "", "Configuration file path")
	cmd.Flags().StringVar(&flags.Token, "token", "", "GitHub token")
	cmd.Flags().BoolVar(&flags.DryRun, "dry-run", false, "Show what would be done without making changes")
	cmd.Flags().BoolVar(&flags.Verbose, "verbose", false, "Enable verbose output")
	cmd.Flags().IntVar(&flags.Parallel, "parallel", 5, "Number of parallel operations")
	cmd.Flags().StringVar(&flags.Timeout, "timeout", "30s", "Request timeout")
}

// NewCmd creates the validate subcommand.
func NewCmd() *cobra.Command {
	var (
		flags  GlobalFlags
		strict bool
		format string
	)

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate repository configuration files",
		Long: `Validate repository configuration using the existing local configuration loader.

Checks YAML decoding, required version and organization, template base references
and cycles, and required repository pattern matches. Unknown-field rejection,
GitHub API compatibility, token permissions, and security policy auditing are not
performed. Strict validation is unsupported.

Examples:
  gz repo-config validate
  gz repo-config validate --config-file custom.yaml
  gz repo-config validate --format json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runValidateCommandTo(cmd.OutOrStdout(), flags, strict, format)
		},
	}

	// Add global flags
	addGlobalFlags(cmd, &flags)

	// Add validate-specific flags
	cmd.Flags().BoolVar(&strict, "strict", false, "Request unsupported strict validation mode")
	cmd.Flags().StringVar(&format, "format", "table", "Output format (table, json, yaml)")

	return cmd
}

// runValidateCommand executes the validate command.
func runValidateCommand(flags GlobalFlags, strict bool, format string) error {
	return runValidateCommandTo(os.Stdout, flags, strict, format)
}

func runValidateCommandTo(out io.Writer, flags GlobalFlags, strict bool, format string) error {
	if strict {
		return fmt.Errorf("strict repository configuration validation is unsupported")
	}
	switch format {
	case "table", "json", "yaml":
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
	targetFile := flags.ConfigFile
	if targetFile == "" {
		targetFile = discoverConfigFile()
	}
	if targetFile == "" {
		return fmt.Errorf("no configuration file found. Use --config-file to specify a file")
	}
	results, validationErr := performValidation(targetFile, false)
	var renderErr error
	switch format {
	case "table":
		renderErr = displayValidationTable(out, results)
	case "json":
		renderErr = displayValidationJSON(out, results)
	case "yaml":
		renderErr = displayValidationYAML(out, results)
	}
	if renderErr != nil {
		return fmt.Errorf("failed to render validation results: %w", renderErr)
	}
	if hasValidationErrors(results) {
		return fmt.Errorf("configuration validation failed: %w", validationErr)
	}
	return nil
}

// ValidationResult represents a single validation check result.
type ValidationResult struct {
	Check      string `json:"check" yaml:"check"`
	Status     string `json:"status" yaml:"status"` // pass, warn, fail
	Message    string `json:"message" yaml:"message"`
	Line       int    `json:"line,omitempty" yaml:"line,omitempty"`
	Severity   string `json:"severity" yaml:"severity"` // error, warning, info
	Suggestion string `json:"suggestion,omitempty" yaml:"suggestion,omitempty"`
}

// ValidationSummary contains overall validation results.
type ValidationSummary struct {
	File     string             `json:"file" yaml:"file"`
	Valid    bool               `json:"valid" yaml:"valid"`
	Errors   int                `json:"errors" yaml:"errors"`
	Warnings int                `json:"warnings" yaml:"warnings"`
	Checks   []ValidationResult `json:"checks" yaml:"checks"`
}

// performValidation runs all validation checks on the configuration file.
func performValidation(configFile string, strict bool) (ValidationSummary, error) {
	results := ValidationSummary{File: configFile, Checks: []ValidationResult{}}
	if strict {
		return results, fmt.Errorf("strict repository configuration validation is unsupported")
	}
	_, err := config.LoadRepoConfig(configFile)
	check := ValidationResult{Check: "Local configuration loader", Status: "pass", Message: "Local YAML decoding and configuration loader checks passed", Severity: "info"}
	if err != nil {
		check.Status = "fail"
		check.Message = err.Error()
		check.Severity = "error"
		results.Errors = 1
	} else {
		results.Valid = true
	}
	results.Checks = append(results.Checks, check)
	return results, err
}

// displayValidationTable displays validation results in table format.
func displayValidationTable(out io.Writer, results ValidationSummary) error {
	var text strings.Builder
	fmt.Fprintf(&text, "%-28s %-8s %s\n", "CHECK", "STATUS", "MESSAGE")
	for _, check := range results.Checks {
		fmt.Fprintf(&text, "%-28s %-8s %s\n", check.Check, getStatusSymbol(check.Status), check.Message)
	}
	if results.Valid {
		fmt.Fprintln(&text, "Configuration passed local loader validation")
	} else {
		fmt.Fprintln(&text, "Configuration has errors")
	}
	fmt.Fprintf(&text, "Summary: %d errors, %d warnings\n", results.Errors, results.Warnings)
	_, err := io.WriteString(out, text.String())
	return err
}

func displayValidationJSON(out io.Writer, results ValidationSummary) error {
	return json.NewEncoder(out).Encode(results)
}

func displayValidationYAML(out io.Writer, results ValidationSummary) error {
	data, err := yaml.Marshal(results)
	if err != nil {
		return err
	}
	_, err = out.Write(data)
	return err
}

// getStatusSymbol returns the symbol for a validation status.
func getStatusSymbol(status string) string {
	switch status {
	case "pass":
		return "✅"
	case "warn":
		return "⚠️"
	case "fail":
		return "❌"
	default:
		return "❓"
	}
}

// hasValidationErrors checks if there are any validation errors.
func hasValidationErrors(results ValidationSummary) bool {
	return results.Errors > 0
}

// discoverConfigFile attempts to find a configuration file.
func discoverConfigFile() string {
	candidates := []string{
		".gzh/repo-config.yaml",
		"repo-config.yaml",
		"gzh-repo-config.yaml",
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	return ""
}
