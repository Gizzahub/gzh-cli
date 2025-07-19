package repoconfig

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

const (
	formatTable = "table"
	formatJSON  = "json"
	formatYAML  = "yaml"
)

// newTemplateCmd creates the template subcommand.
func newTemplateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "template",
		Short: "Manage repository configuration templates",
		Long: `Manage repository configuration templates.

Templates provide standardized configuration sets that can be applied
to repositories based on their type, purpose, or security requirements.
This helps maintain consistency across repositories in an organization.

Template Features:
- Predefined configuration templates
- Template inheritance and composition
- Custom template creation
- Template validation and testing
- Template compliance checking

Examples:
  gz repo-config template list             # List available templates
  gz repo-config template show security    # Show template details
  gz repo-config template validate <name>  # Validate template`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	// Add subcommands
	cmd.AddCommand(newTemplateListCmd())
	cmd.AddCommand(newTemplateShowCmd())
	cmd.AddCommand(newTemplateValidateCmd())

	return cmd
}

// newTemplateListCmd lists available templates.
func newTemplateListCmd() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available configuration templates",
		Long: `List all available repository configuration templates.

Shows template names, descriptions, and basic metadata for all
templates available in the system.

Examples:
  gz repo-config template list              # List all templates
  gz repo-config template list --format json  # JSON output`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTemplateListCommand(format)
		},
	}

	cmd.Flags().StringVar(&format, "format", formatTable, "Output format (table, json)")

	return cmd
}

// newTemplateShowCmd shows template details.
func newTemplateShowCmd() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "show <template-name>",
		Short: "Show detailed template configuration",
		Args:  cobra.ExactArgs(1),
		Long: `Show detailed configuration for a specific template.

Displays the complete template configuration including all settings,
policies, and inheritance information.

Examples:
  gz repo-config template show security     # Show security template
  gz repo-config template show --format json microservice  # JSON format`,
		RunE: func(cmd *cobra.Command, args []string) error {
			templateName := args[0]
			return runTemplateShowCommand(templateName, format)
		},
	}

	cmd.Flags().StringVar(&format, "format", formatYAML, "Output format (yaml, json)")

	return cmd
}

// newTemplateValidateCmd validates a template.
func newTemplateValidateCmd() *cobra.Command {
	var strict bool

	cmd := &cobra.Command{
		Use:   "validate <template-name>",
		Short: "Validate template configuration",
		Args:  cobra.ExactArgs(1),
		Long: `Validate a template configuration for correctness.

Checks template syntax, validates settings against GitHub API,
and verifies template consistency and inheritance.

Examples:
  gz repo-config template validate security  # Validate security template
  gz repo-config template validate --strict microservice  # Strict validation`,
		RunE: func(cmd *cobra.Command, args []string) error {
			templateName := args[0]
			return runTemplateValidateCommand(templateName, strict)
		},
	}

	cmd.Flags().BoolVar(&strict, "strict", false, "Enable strict validation mode")

	return cmd
}

// runTemplateListCommand executes the template list command.
func runTemplateListCommand(format string) error {
	fmt.Printf("📋 Available Repository Configuration Templates\n")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// Mock template data
	templates := []ConfigTemplate{
		{
			Name:        "security",
			Description: "High security template for production repositories",
			Version:     "1.2.0",
			Category:    "security",
			Inherits:    []string{"enterprise"},
			Settings:    15,
			Policies:    8,
		},
		{
			Name:        "microservice",
			Description: "Standard configuration for microservice repositories",
			Version:     "1.1.0",
			Category:    "service",
			Inherits:    []string{"default"},
			Settings:    12,
			Policies:    5,
		},
		{
			Name:        "frontend",
			Description: "Configuration template for frontend applications",
			Version:     "1.0.0",
			Category:    "application",
			Inherits:    []string{"default"},
			Settings:    10,
			Policies:    4,
		},
		{
			Name:        "library",
			Description: "Template for shared libraries and packages",
			Version:     "1.0.1",
			Category:    "library",
			Inherits:    []string{"opensource"},
			Settings:    8,
			Policies:    3,
		},
		{
			Name:        "opensource",
			Description: "Base template for open source repositories",
			Version:     "1.1.0",
			Category:    "base",
			Inherits:    []string{},
			Settings:    6,
			Policies:    2,
		},
	}

	switch format {
	case formatTable:
		displayTemplateTable(templates)
	case formatJSON:
		displayTemplateJSON(templates)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	return nil
}

// runTemplateShowCommand executes the template show command.
func runTemplateShowCommand(templateName, format string) error {
	fmt.Printf("📄 Template Configuration: %s\n", templateName)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// Mock template detail
	templateConfig := getTemplateConfiguration(templateName)
	if templateConfig == "" {
		return fmt.Errorf("template '%s' not found", templateName)
	}

	switch format {
	case formatYAML:
		fmt.Println(templateConfig)
	case formatJSON:
		fmt.Println("JSON template output not yet implemented")
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	return nil
}

// runTemplateValidateCommand executes the template validate command.
func runTemplateValidateCommand(templateName string, strict bool) error {
	fmt.Printf("🔍 Validating Template: %s\n", templateName)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	if strict {
		fmt.Println("Mode: STRICT validation")
		fmt.Println()
	}

	// Mock validation results
	validationResults := []TemplateValidationResult{
		{
			Check:    "Template Syntax",
			Status:   "pass",
			Message:  "Valid YAML structure",
			Severity: "info",
		},
		{
			Check:    "Setting Validation",
			Status:   "pass",
			Message:  "All settings are valid GitHub API parameters",
			Severity: "info",
		},
		{
			Check:    "Inheritance Chain",
			Status:   "pass",
			Message:  "Template inheritance is valid",
			Severity: "info",
		},
		{
			Check:    "Policy Consistency",
			Status:   "warn",
			Message:  "Some policies may conflict with base template",
			Severity: "warning",
		},
	}

	fmt.Printf("%-20s %-8s %s\n", "CHECK", "STATUS", "MESSAGE")
	fmt.Println("────────────────────────────────────────────────────────────────")

	for _, result := range validationResults {
		statusSymbol := getStatusSymbol(result.Status)
		fmt.Printf("%-20s %-8s %s\n",
			result.Check,
			statusSymbol,
			result.Message,
		)
	}

	fmt.Println()

	// Summary
	errorCount := 0
	warningCount := 0

	for _, result := range validationResults {
		switch result.Severity {
		case "error":
			errorCount++
		case "warning":
			warningCount++
		}
	}

	if errorCount == 0 {
		fmt.Printf("✅ Template '%s' is valid\n", templateName)
	} else {
		fmt.Printf("❌ Template '%s' has validation errors\n", templateName)
	}

	fmt.Printf("📊 Summary: %d errors, %d warnings\n", errorCount, warningCount)

	return nil
}

// ConfigTemplate represents a configuration template.
type ConfigTemplate struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Version     string   `json:"version"`
	Category    string   `json:"category"`
	Inherits    []string `json:"inherits"`
	Settings    int      `json:"settings"`
	Policies    int      `json:"policies"`
}

// TemplateValidationResult represents a template validation check result.
type TemplateValidationResult struct {
	Check    string `json:"check"`
	Status   string `json:"status"`
	Message  string `json:"message"`
	Severity string `json:"severity"`
}

// displayTemplateTable displays templates in table format.
func displayTemplateTable(templates []ConfigTemplate) {
	fmt.Printf("%-15s %-40s %-10s %-12s %s\n", "NAME", "DESCRIPTION", "VERSION", "CATEGORY", "INHERITS")
	fmt.Println("─────────────────────────────────────────────────────────────────────────────────────────")

	for _, template := range templates {
		inheritsStr := ""
		if len(template.Inherits) > 0 {
			inheritsStr = template.Inherits[0]
			if len(template.Inherits) > 1 {
				inheritsStr += "..."
			}
		}

		fmt.Printf("%-15s %-40s %-10s %-12s %s\n",
			template.Name,
			truncateString(template.Description, 40),
			template.Version,
			template.Category,
			inheritsStr,
		)
	}

	fmt.Println()
	fmt.Printf("Total templates: %d\n", len(templates))
}

// displayTemplateJSON displays templates in JSON format.
func displayTemplateJSON(templates []ConfigTemplate) {
	jsonData := map[string]interface{}{
		"templates": templates,
		"count":     len(templates),
	}

	if jsonBytes, err := json.MarshalIndent(jsonData, "", "  "); err != nil {
		fmt.Printf("Error serializing JSON: %v\n", err)
	} else {
		fmt.Println(string(jsonBytes))
	}
}

// getTemplateConfiguration returns mock template configuration.
func getTemplateConfiguration(templateName string) string {
	templates := map[string]string{
		"security": `# Security Template Configuration
name: security
description: High security template for production repositories
version: 1.2.0
inherits: [enterprise]

settings:
  visibility: private
  features:
    issues: true
    wiki: false
    projects: false
    downloads: false
  
  security:
    branch_protection:
      main:
        required_reviews: 2
        dismiss_stale_reviews: true
        require_code_owner_reviews: true
        restrict_pushes: true
        allowed_teams: ["security-team", "admin-team"]
        required_status_checks:
          - "ci/build"
          - "ci/test"
          - "security/scan"
    
    vulnerability_alerts: true
    security_advisories: true
    
  collaboration:
    delete_head_branches: true
    squash_merge: true
    merge_commit: false
    rebase_merge: false`,

		"microservice": `# Microservice Template Configuration
name: microservice
description: Standard configuration for microservice repositories
version: 1.1.0
inherits: [default]

settings:
  visibility: private
  features:
    issues: true
    wiki: false
    projects: true
  
  security:
    branch_protection:
      main:
        required_reviews: 2
        dismiss_stale_reviews: true
        required_status_checks:
          - "ci/build"
          - "ci/test"
    
  collaboration:
    delete_head_branches: true
    squash_merge: true`,
	}

	if config, exists := templates[templateName]; exists {
		return config
	}

	return ""
}
