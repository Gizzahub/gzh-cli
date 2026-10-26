// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package template

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const (
	formatTable = "table"
	formatJSON  = "json"
	formatYAML  = "yaml"
)

// truncateString truncates a string to a maximum length.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// NewCmd creates the template subcommand.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "template",
		Short: "Manage repository configuration templates",
		Long: `Built-in repository template catalog and YAML examples.

List returns static catalog metadata. Show returns the security or microservice
YAML example; other catalog entries have no bundled YAML sample.
Template validation and JSON show are unsupported and return explicit errors.
Use --help on a subcommand to inspect retained flags.`,
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
		Long: `Show a built-in YAML example (security or microservice).

JSON show is unsupported. Catalog names without a bundled sample return
not found; no remote or custom template store is queried.`,
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
		Short: "Template validation (currently unsupported)",
		Args:  cobra.ExactArgs(1),
		Long: `Template validation is currently unsupported.

No resolved template schema, inheritance, or GitHub API validation backend
is implemented. Execution returns an error; --strict does not enable validation.`,
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
	// 내장 catalog이며 원격 저장소 상태를 관측한 결과가 아니다.
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
		fmt.Printf("📋 Available Repository Configuration Templates\n")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println()
		displayTemplateTable(templates)
	case formatJSON:
		return displayTemplateJSON(templates)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	return nil
}

// runTemplateShowCommand executes the template show command.
func runTemplateShowCommand(templateName, format string) error {
	templateConfig := getTemplateConfiguration(templateName)
	if templateConfig == "" {
		return fmt.Errorf("template '%s' not found", templateName)
	}
	switch format {
	case formatYAML:
		_, err := fmt.Fprintln(os.Stdout, templateConfig)
		return err
	case formatJSON:
		return fmt.Errorf("JSON template output is not supported")
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// runTemplateValidateCommand executes the template validate command.
func runTemplateValidateCommand(templateName string, strict bool) error {
	return fmt.Errorf("template validation is not supported: no resolved template validation backend is implemented")
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

		fmt.Printf(
			"%-15s %-40s %-10s %-12s %s\n",
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
func displayTemplateJSON(templates []ConfigTemplate) error {
	jsonData := map[string]any{"templates": templates, "count": len(templates)}
	jsonBytes, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(os.Stdout, string(jsonBytes))
	return err
}

// getTemplateConfiguration returns the two built-in YAML examples.
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
