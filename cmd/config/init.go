// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// newInitCmd creates the config init subcommand.
func newInitCmd() *cobra.Command {
	var (
		outputFile string
		force      bool
		minimal    bool
	)

	cmd := &cobra.Command{
		Use:   "init [output-file]",
		Short: "Initialize a new gzh.yaml configuration file",
		Long: `Initialize a new gzh.yaml configuration file with interactive prompts.

This command guides you through creating a configuration file by asking
questions about your Git providers, organizations, and preferences.

Examples:
  gz config init                        # Create gzh.yaml in current directory
  gz config init my-config.yaml        # Create custom config file
  gz config init --minimal             # Create minimal configuration
  gz config init --force              # Overwrite existing file`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			// Determine output file path
			if len(args) > 0 {
				outputFile = args[0]
			} else if outputFile == "" {
				outputFile = "gzh.yaml"
			}

			return initializeConfig(outputFile, force, minimal)
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path (default: gzh.yaml)")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing file")
	cmd.Flags().BoolVar(&minimal, "minimal", false, "Create minimal configuration")

	return cmd
}

// Template represents a configuration template.
type Template struct {
	Version         string
	DefaultProvider string
	Providers       map[string]ProviderTemplate
}

// ProviderTemplate represents a provider configuration template.
type ProviderTemplate struct {
	Token  string
	Orgs   []TargetTemplate
	Groups []TargetTemplate
}

// TargetTemplate represents an organization or group template.
type TargetTemplate struct {
	Name       string
	Visibility string
	CloneDir   string
	Match      string
	Exclude    []string
	Strategy   string
	Flatten    bool
	Recursive  bool
}

// initializeConfig runs the interactive configuration initialization.
func initializeConfig(outputFile string, force bool, minimal bool) error {
	// Check if file exists and force flag
	if _, err := os.Stat(outputFile); err == nil && !force {
		return fmt.Errorf("file %s already exists. Use --force to overwrite", outputFile)
	}

	fmt.Println("🚀 Welcome to gzh.yaml configuration initialization!")
	fmt.Println()

	if minimal {
		return createMinimalConfig(outputFile)
	}

	return createInteractiveConfig(outputFile)
}

// createMinimalConfig creates a minimal configuration file.
func createMinimalConfig(outputFile string) error {
	template := Template{
		Version:         "1.0.0",
		DefaultProvider: "github",
		Providers: map[string]ProviderTemplate{
			"github": {
				Token: "${GITHUB_TOKEN}",
				Orgs: []TargetTemplate{
					{
						Name:       "your-org-name",
						Visibility: "all",
						CloneDir:   "${HOME}/repos/github",
						Strategy:   "reset",
					},
				},
			},
		},
	}

	content := generateConfigContent(template)

	if err := os.WriteFile(outputFile, []byte(content), 0o644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("✓ Minimal configuration created: %s\n", outputFile)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("1. Set your GitHub token: export GITHUB_TOKEN=\"your_token\"")
	fmt.Println("2. Update the organization name in the config file")
	fmt.Println("3. Validate the configuration: gz config validate")

	return nil
}

// createInteractiveConfig creates configuration through interactive prompts.
func createInteractiveConfig(outputFile string) error {
	reader := bufio.NewReader(os.Stdin)
	template := Template{
		Version:   "1.0.0",
		Providers: make(map[string]ProviderTemplate),
	}

	// Ask for default provider
	fmt.Println("📋 Basic Configuration")
	fmt.Println("────────────────────")

	defaultProvider := promptWithDefault(reader, "Default Git provider (github/gitlab/gitea)", "github")
	template.DefaultProvider = defaultProvider

	// Configure providers
	providers := []string{}
	if promptYesNo(reader, fmt.Sprintf("Configure %s?", defaultProvider), true) {
		providers = append(providers, defaultProvider)
	}

	// Ask about additional providers
	allProviders := []string{"github", "gitlab", "gitea"}
	for _, provider := range allProviders {
		if provider != defaultProvider {
			if promptYesNo(reader, fmt.Sprintf("Configure %s?", provider), false) {
				providers = append(providers, provider)
			}
		}
	}

	// Configure each provider
	for _, provider := range providers {
		fmt.Printf("\n🔧 Configuring %s\n", strings.ToUpper(provider[:1])+provider[1:])
		fmt.Println("────────────────────")

		providerTemplate := configureProvider(reader, provider)
		template.Providers[provider] = providerTemplate
	}

	// Generate and write configuration
	content := generateConfigContent(template)

	if err := os.WriteFile(outputFile, []byte(content), 0o644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("\n✓ Configuration created: %s\n", outputFile)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("1. Set your authentication tokens:")

	for provider := range template.Providers {
		fmt.Printf("   export %s_TOKEN=\"your_%s_token\"\n", strings.ToUpper(provider), provider)
	}

	fmt.Println("2. Validate the configuration: gz config validate")
	fmt.Println("3. Test with dry run: gz bulk-clone --dry-run --use-gzh-config")

	return nil
}

// configureProvider configures a specific provider.
func configureProvider(reader *bufio.Reader, provider string) ProviderTemplate {
	template := ProviderTemplate{
		Token: fmt.Sprintf("${%s_TOKEN}", strings.ToUpper(provider)),
	}

	// Determine if it's organization-based or group-based
	isGroupBased := provider == "gitlab"

	entityType := "organization"
	if isGroupBased {
		entityType = "group"
	}

	// Ask for entities (orgs/groups)
	for {
		entityName := promptString(reader, fmt.Sprintf("Enter %s name (empty to finish)", entityType))
		if entityName == "" {
			break
		}

		target := TargetTemplate{
			Name:       entityName,
			Visibility: promptWithDefault(reader, "Repository visibility (public/private/all)", "all"),
			CloneDir:   fmt.Sprintf("${HOME}/repos/%s/%s", provider, entityName),
			Strategy:   promptWithDefault(reader, "Clone strategy (reset/pull/fetch)", "reset"),
		}

		// Additional options
		if promptYesNo(reader, "Configure advanced filtering?", false) {
			if match := promptString(reader, "Match pattern (regex, empty to skip)"); match != "" {
				target.Match = match
			}

			if exclude := promptString(reader, "Exclude patterns (comma-separated, empty to skip)"); exclude != "" {
				target.Exclude = strings.Split(exclude, ",")
				for i := range target.Exclude {
					target.Exclude[i] = strings.TrimSpace(target.Exclude[i])
				}
			}
		}

		target.Flatten = promptYesNo(reader, "Flatten directory structure?", false)

		if isGroupBased {
			target.Recursive = promptYesNo(reader, "Include subgroups recursively?", false)
			template.Groups = append(template.Groups, target)
		} else {
			template.Orgs = append(template.Orgs, target)
		}
	}

	return template
}

// promptString prompts for a string input.
func promptString(reader *bufio.Reader, prompt string) string {
	fmt.Printf("%s: ", prompt)

	input, _ := reader.ReadString('\n')

	return strings.TrimSpace(input)
}

// promptWithDefault prompts for input with a default value.
func promptWithDefault(reader *bufio.Reader, prompt, defaultValue string) string {
	fmt.Printf("%s [%s]: ", prompt, defaultValue)

	input, _ := reader.ReadString('\n')

	input = strings.TrimSpace(input)
	if input == "" {
		return defaultValue
	}

	return input
}

// promptYesNo prompts for a yes/no answer.
func promptYesNo(reader *bufio.Reader, prompt string, defaultValue bool) bool {
	defaultStr := "y/N"
	if defaultValue {
		defaultStr = "Y/n"
	}

	fmt.Printf("%s [%s]: ", prompt, defaultStr)

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input == "" {
		return defaultValue
	}

	return input == "y" || input == "yes"
}

// generateConfigContent generates the YAML configuration content.
func generateConfigContent(template Template) string {
	var content strings.Builder

	content.WriteString("# gzh.yaml - Generated configuration file\n")
	content.WriteString("# Generated by: gz config init\n")
	content.WriteString("# Edit this file to customize your configuration\n\n")

	content.WriteString(fmt.Sprintf("version: \"%s\"\n", template.Version))
	content.WriteString(fmt.Sprintf("default_provider: %s\n\n", template.DefaultProvider))

	content.WriteString("providers:\n")

	for providerName, provider := range template.Providers {
		content.WriteString(fmt.Sprintf("  %s:\n", providerName))
		content.WriteString(fmt.Sprintf("    token: \"%s\"\n", provider.Token))

		if len(provider.Orgs) > 0 {
			content.WriteString("    orgs:\n")

			for _, org := range provider.Orgs {
				writeTarget(&content, org, "      ")
			}
		}

		if len(provider.Groups) > 0 {
			content.WriteString("    groups:\n")

			for _, group := range provider.Groups {
				writeTarget(&content, group, "      ")
			}
		}

		content.WriteString("\n")
	}

	return content.String()
}

// writeTarget writes a target (org/group) configuration.
func writeTarget(content *strings.Builder, target TargetTemplate, indent string) {
	fmt.Fprintf(content, "%s- name: \"%s\"\n", indent, target.Name)
	fmt.Fprintf(content, "%s  visibility: \"%s\"\n", indent, target.Visibility)
	fmt.Fprintf(content, "%s  clone_dir: \"%s\"\n", indent, target.CloneDir)

	if target.Match != "" {
		fmt.Fprintf(content, "%s  match: \"%s\"\n", indent, target.Match)
	}

	if len(target.Exclude) > 0 {
		fmt.Fprintf(content, "%s  exclude:\n", indent)

		for _, pattern := range target.Exclude {
			fmt.Fprintf(content, "%s    - \"%s\"\n", indent, pattern)
		}
	}

	fmt.Fprintf(content, "%s  strategy: \"%s\"\n", indent, target.Strategy)

	if target.Flatten {
		fmt.Fprintf(content, "%s  flatten: true\n", indent)
	}

	if target.Recursive {
		fmt.Fprintf(content, "%s  recursive: true\n", indent)
	}
}
