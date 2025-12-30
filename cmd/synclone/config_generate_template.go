// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package synclone

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/gizzahub/gzh-cli/internal/synclone/template"
)

// newConfigGenerateTemplateCmd creates the config generate template command.
func newConfigGenerateTemplateCmd() *cobra.Command {
	var (
		templateName   string
		outputFile     string
		listTemplates  bool
		templateDir    string
		variables      []string
		interactiveVar bool
	)

	cmd := &cobra.Command{
		Use:   "template",
		Short: "Generate synclone configuration from templates",
		Long: `Generate synclone configuration files from predefined templates.

Templates provide pre-configured setups for common scenarios like enterprise
environments, multi-organization setups, or personal projects.

Available built-in templates:
  - enterprise: Multi-organization setup with security features
  - minimal: Simple setup for personal or small team use
  - multi-org: Setup for managing multiple organizations
  - personal: Setup for personal repositories and projects

Examples:
  # List available templates
  gz synclone config generate template --list-templates

  # Generate from enterprise template
  gz synclone config generate template --template enterprise --output config.yaml

  # Generate with variables
  gz synclone config generate template --template enterprise --var CompanyOrg=mycompany --var CompanyGroup=mygroup

  # Interactive variable input
  gz synclone config generate template --template enterprise --interactive`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigGenerateTemplate(templateName, outputFile, listTemplates, templateDir, variables, interactiveVar)
		},
	}

	cmd.Flags().StringVarP(&templateName, "template", "t", "", "Template name to use")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "synclone-template.yaml", "Output configuration file")
	cmd.Flags().BoolVar(&listTemplates, "list-templates", false, "List available templates")
	cmd.Flags().StringVar(&templateDir, "template-dir", "", "Custom template directory")
	cmd.Flags().StringSliceVar(&variables, "var", []string{}, "Template variables in key=value format")
	cmd.Flags().BoolVarP(&interactiveVar, "interactive", "i", false, "Interactive variable input")

	return cmd
}

// runConfigGenerateTemplate executes the template generation command.
func runConfigGenerateTemplate(templateName, outputFile string, listTemplates bool, templateDir string, variables []string, interactiveVar bool) error {
	// Set default template directory
	if templateDir == "" {
		homeDir, _ := os.UserHomeDir()
		templateDir = filepath.Join(homeDir, ".config", "gzh-manager", "templates")
	}

	// Create template engine
	engine := template.NewTemplateEngine(templateDir)

	// Ensure built-in templates exist
	if err := engine.CreateBuiltinTemplates(); err != nil {
		return fmt.Errorf("failed to create builtin templates: %w", err)
	}

	// Handle list templates
	if listTemplates {
		return listAvailableTemplates(engine)
	}

	// Validate template name
	if templateName == "" {
		return fmt.Errorf("template name is required (use --template flag or --list-templates to see available templates)")
	}

	// Load template information
	templateInfo, err := engine.GetTemplateInfo(templateName)
	if err != nil {
		return fmt.Errorf("failed to load template %s: %w", templateName, err)
	}

	fmt.Printf("🎨 Generating configuration from template: %s\n", templateInfo.Name)
	fmt.Printf("   Description: %s\n", templateInfo.Description)

	// Parse variables
	varMap := make(map[string]interface{})
	if err := parseVariables(variables, varMap); err != nil {
		return fmt.Errorf("failed to parse variables: %w", err)
	}

	// Interactive variable input
	if interactiveVar {
		if err := promptForVariables(templateInfo, varMap); err != nil {
			return fmt.Errorf("failed to collect variables: %w", err)
		}
	}

	// Generate configuration
	config, err := engine.GenerateConfig(templateName, varMap)
	if err != nil {
		return fmt.Errorf("failed to generate configuration: %w", err)
	}

	// Save configuration
	if err := saveTemplateConfig(config, outputFile); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("✅ Configuration generated successfully\n")
	fmt.Printf("📁 Saved to: %s\n", outputFile)

	// Display template summary
	displayTemplateSummary(templateInfo, varMap)

	return nil
}

// listAvailableTemplates lists all available templates.
func listAvailableTemplates(engine *template.TemplateEngine) error {
	templates, err := engine.ListTemplates()
	if err != nil {
		return fmt.Errorf("failed to list templates: %w", err)
	}

	if len(templates) == 0 {
		fmt.Println("No templates available")
		return nil
	}

	fmt.Println("📋 Available Templates:")
	fmt.Println()

	for _, templateName := range templates {
		templateInfo, err := engine.GetTemplateInfo(templateName)
		if err != nil {
			fmt.Printf("  ❌ %s (failed to load info)\n", templateName)
			continue
		}

		fmt.Printf("  📄 %s\n", templateInfo.Name)
		fmt.Printf("     ID: %s\n", templateName)
		fmt.Printf("     Description: %s\n", templateInfo.Description)

		if len(templateInfo.Variables) > 0 {
			fmt.Printf("     Variables:\n")
			for _, variable := range templateInfo.Variables {
				required := ""
				if variable.Required {
					required = " (required)"
				}
				defaultValue := ""
				if variable.DefaultValue != nil {
					defaultValue = fmt.Sprintf(" [default: %v]", variable.DefaultValue)
				}
				fmt.Printf("       - %s: %s%s%s\n", variable.Name, variable.Description, required, defaultValue)
			}
		}
		fmt.Println()
	}

	return nil
}

// parseVariables parses variable strings in key=value format.
func parseVariables(variables []string, varMap map[string]interface{}) error {
	for _, variable := range variables {
		parts := strings.SplitN(variable, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid variable format: %s (expected key=value)", variable)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if key == "" {
			return fmt.Errorf("variable key cannot be empty: %s", variable)
		}

		varMap[key] = value
	}

	return nil
}

// promptForVariables prompts the user for template variables.
func promptForVariables(templateInfo *template.TemplateConfig, varMap map[string]interface{}) error {
	fmt.Println("\n📝 Template Variables:")

	for _, variable := range templateInfo.Variables {
		// Skip if already provided
		if _, exists := varMap[variable.Name]; exists {
			continue
		}

		prompt := fmt.Sprintf("  %s", variable.Description)
		if variable.DefaultValue != nil {
			prompt += fmt.Sprintf(" [%v]", variable.DefaultValue)
		}
		if variable.Required {
			prompt += " (required)"
		}

		if len(variable.Options) > 0 {
			prompt += fmt.Sprintf(" (options: %s)", strings.Join(variable.Options, ", "))
		}

		prompt += ": "

		fmt.Print(prompt)

		var input string
		fmt.Scanln(&input)
		input = strings.TrimSpace(input)

		// Use default value if no input provided
		if input == "" && variable.DefaultValue != nil {
			varMap[variable.Name] = variable.DefaultValue
		} else if input != "" {
			// Validate options if provided
			if len(variable.Options) > 0 {
				validOption := false
				for _, option := range variable.Options {
					if input == option {
						validOption = true
						break
					}
				}
				if !validOption {
					fmt.Printf("    Invalid option. Valid options: %s\n", strings.Join(variable.Options, ", "))
					return fmt.Errorf("invalid option for variable %s", variable.Name)
				}
			}

			varMap[variable.Name] = input
		} else if variable.Required {
			return fmt.Errorf("required variable %s not provided", variable.Name)
		}
	}

	return nil
}

// saveTemplateConfig saves the generated configuration to a file.
func saveTemplateConfig(config map[string]interface{}, filename string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(filename, data, 0o644); err != nil {
		return fmt.Errorf("failed to write configuration file: %w", err)
	}

	return nil
}

// displayTemplateSummary displays a summary of the template generation.
func displayTemplateSummary(templateInfo *template.TemplateConfig, variables map[string]interface{}) {
	fmt.Printf("\n📊 Template Summary:\n")
	fmt.Printf("   Template: %s\n", templateInfo.Name)

	if len(variables) > 0 {
		fmt.Printf("   Variables used:\n")
		for key, value := range variables {
			fmt.Printf("     %s: %v\n", key, value)
		}
	}

	// Analyze generated configuration
	if providers, ok := templateInfo.Template["providers"].(map[string]interface{}); ok {
		fmt.Printf("   Providers configured: %d\n", len(providers))
		for provider := range providers {
			fmt.Printf("     - %s\n", provider)
		}
	}

	fmt.Printf("\n💡 Next steps:\n")
	fmt.Printf("   1. Review the generated configuration file\n")
	fmt.Printf("   2. Customize settings as needed\n")
	fmt.Printf("   3. Set required environment variables (tokens, etc.)\n")
	fmt.Printf("   4. Run: gz synclone run --config %s\n", "synclone-template.yaml")
}
