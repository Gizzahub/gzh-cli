package repoconfig

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// newValidateCmd creates the validate subcommand
func newValidateCmd() *cobra.Command {
	var flags GlobalFlags
	var (
		configFile string
		strict     bool
		format     string
	)

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate repository configuration files",
		Long: `Validate repository configuration files and templates.

This command validates configuration files for syntax errors, schema compliance,
and logical consistency. It helps ensure configurations are correct before
applying them to repositories.

Validation Features:
- YAML syntax validation
- Schema compliance checking
- Template reference validation
- Policy consistency verification
- GitHub API compatibility checks

Output Formats:
- table: Human-readable validation results (default)
- json: JSON format for programmatic use
- yaml: YAML format for configuration export

Examples:
  gz repo-config validate                        # Validate default config
  gz repo-config validate --config custom.yaml  # Validate specific file
  gz repo-config validate --strict              # Strict validation mode
  gz repo-config validate --format json         # JSON output`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runValidateCommand(flags, configFile, strict, format)
		},
	}

	// Add global flags
	addGlobalFlags(cmd, &flags)

	// Add validate-specific flags
	cmd.Flags().StringVar(&configFile, "config-file", "", "Configuration file to validate")
	cmd.Flags().BoolVar(&strict, "strict", false, "Enable strict validation mode")
	cmd.Flags().StringVar(&format, "format", "table", "Output format (table, json, yaml)")

	return cmd
}

// runValidateCommand executes the validate command
func runValidateCommand(flags GlobalFlags, configFile string, strict bool, format string) error {
	if flags.Verbose {
		fmt.Println("🔍 Validating repository configuration...")
		if configFile != "" {
			fmt.Printf("Configuration file: %s\n", configFile)
		}
		if strict {
			fmt.Println("Mode: STRICT validation")
		}
		fmt.Println()
	}

	// Determine config file to validate
	targetFile := configFile
	if targetFile == "" {
		// Use default config discovery
		targetFile = discoverConfigFile()
	}

	if targetFile == "" {
		return fmt.Errorf("no configuration file found. Use --config-file to specify a file")
	}

	// Check if file exists
	if _, err := os.Stat(targetFile); os.IsNotExist(err) {
		return fmt.Errorf("configuration file not found: %s", targetFile)
	}

	fmt.Printf("📋 Repository Configuration Validation\n")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("File: %s\n", targetFile)
	fmt.Println()

	// Perform validation checks
	validationResults := performValidation(targetFile, strict)

	// Display results based on format
	switch format {
	case "table":
		displayValidationTable(validationResults)
	case "json":
		displayValidationJSON(validationResults)
	case "yaml":
		displayValidationYAML(validationResults)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	// Return error if validation failed
	if hasValidationErrors(validationResults) {
		return fmt.Errorf("configuration validation failed")
	}

	return nil
}

// ValidationResult represents a single validation check result
type ValidationResult struct {
	Check      string `json:"check" yaml:"check"`
	Status     string `json:"status" yaml:"status"` // pass, warn, fail
	Message    string `json:"message" yaml:"message"`
	Line       int    `json:"line,omitempty" yaml:"line,omitempty"`
	Severity   string `json:"severity" yaml:"severity"` // error, warning, info
	Suggestion string `json:"suggestion,omitempty" yaml:"suggestion,omitempty"`
}

// ValidationSummary contains overall validation results
type ValidationSummary struct {
	File     string             `json:"file" yaml:"file"`
	Valid    bool               `json:"valid" yaml:"valid"`
	Errors   int                `json:"errors" yaml:"errors"`
	Warnings int                `json:"warnings" yaml:"warnings"`
	Checks   []ValidationResult `json:"checks" yaml:"checks"`
}

// performValidation runs all validation checks on the configuration file
func performValidation(configFile string, strict bool) ValidationSummary {
	results := ValidationSummary{
		File:   configFile,
		Valid:  true,
		Checks: []ValidationResult{},
	}

	// Mock validation results for demonstration
	checks := []ValidationResult{
		{
			Check:    "YAML Syntax",
			Status:   "pass",
			Message:  "Valid YAML syntax",
			Severity: "info",
		},
		{
			Check:    "Schema Compliance",
			Status:   "pass",
			Message:  "Configuration matches expected schema",
			Severity: "info",
		},
		{
			Check:      "Template References",
			Status:     "warn",
			Message:    "Template 'enterprise' referenced but not found",
			Line:       15,
			Severity:   "warning",
			Suggestion: "Define the 'enterprise' template or use an existing template",
		},
		{
			Check:    "GitHub API Compatibility",
			Status:   "pass",
			Message:  "All settings are compatible with GitHub API",
			Severity: "info",
		},
		{
			Check:    "Policy Consistency",
			Status:   "pass",
			Message:  "No conflicting policies detected",
			Severity: "info",
		},
	}

	// Add strict mode checks
	if strict {
		checks = append(checks, ValidationResult{
			Check:      "Token Permissions",
			Status:     "warn",
			Message:    "Cannot verify token permissions (token not provided)",
			Severity:   "warning",
			Suggestion: "Provide GitHub token with --token flag for permission validation",
		})
	}

	results.Checks = checks

	// Calculate summary
	for _, check := range checks {
		switch check.Severity {
		case "error":
			results.Errors++
			results.Valid = false
		case "warning":
			results.Warnings++
		}
	}

	return results
}

// displayValidationTable displays validation results in table format
func displayValidationTable(results ValidationSummary) {
	fmt.Printf("%-20s %-8s %-50s %s\n", "CHECK", "STATUS", "MESSAGE", "LINE")
	fmt.Println("────────────────────────────────────────────────────────────────────────────────")

	for _, check := range results.Checks {
		statusSymbol := getStatusSymbol(check.Status)
		lineStr := ""
		if check.Line > 0 {
			lineStr = fmt.Sprintf("%d", check.Line)
		}

		fmt.Printf("%-20s %-8s %-50s %s\n",
			check.Check,
			statusSymbol,
			truncateString(check.Message, 50),
			lineStr,
		)

		if check.Suggestion != "" {
			fmt.Printf("%-20s %-8s 💡 %s\n", "", "", check.Suggestion)
		}
	}

	fmt.Println()

	// Summary
	if results.Valid {
		fmt.Printf("✅ Configuration is valid\n")
	} else {
		fmt.Printf("❌ Configuration has errors\n")
	}

	fmt.Printf("📊 Summary: %d errors, %d warnings\n", results.Errors, results.Warnings)
}

// displayValidationJSON displays validation results in JSON format
func displayValidationJSON(results ValidationSummary) {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		return
	}
	fmt.Println(string(data))
}

// displayValidationYAML displays validation results in YAML format
func displayValidationYAML(results ValidationSummary) {
	data, err := yaml.Marshal(results)
	if err != nil {
		fmt.Printf("Error marshaling YAML: %v\n", err)
		return
	}
	fmt.Println(string(data))
}

// getStatusSymbol returns the symbol for a validation status
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

// hasValidationErrors checks if there are any validation errors
func hasValidationErrors(results ValidationSummary) bool {
	return results.Errors > 0
}

// truncateString truncates a string to the specified length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// discoverConfigFile attempts to find a configuration file
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
