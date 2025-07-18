package genconfig

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

type genConfigInitOptions struct {
	outputFile  string
	interactive bool
	template    string
}

func defaultGenConfigInitOptions() *genConfigInitOptions {
	return &genConfigInitOptions{
		outputFile:  "bulk-clone.yaml",
		interactive: true,
		template:    "simple",
	}
}

func newGenConfigInitCmd() *cobra.Command {
	o := defaultGenConfigInitOptions()

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new bulk-clone configuration file",
		Long: `Initialize a new bulk-clone.yaml configuration file through an interactive wizard.

This command will guide you through creating a configuration file for managing
multiple Git repositories across different hosting services (GitHub, GitLab, Gitea).

The wizard will ask about:
- Repository hosting services you use
- Organization/group names
- Target directories for each organization
- Protocol preferences (HTTPS, SSH)
- Repository ignore patterns

Examples:
  # Interactive configuration creation
  gz gen-config init
  
  # Non-interactive with simple template
  gz gen-config init --template simple --no-interactive
  
  # Generate to custom file
  gz gen-config init --output my-config.yaml`,
		RunE: o.run,
	}

	cmd.Flags().StringVarP(&o.outputFile, "output", "o", o.outputFile, "Output configuration file")
	cmd.Flags().BoolVar(&o.interactive, "interactive", true, "Run interactive configuration wizard")
	cmd.Flags().StringVar(&o.template, "template", o.template, "Template to use (simple, comprehensive)")

	// Add --no-interactive flag as the opposite of --interactive
	cmd.Flags().BoolVar(&o.interactive, "no-interactive", false, "Skip interactive wizard")
	cmd.Flag("no-interactive").NoOptDefVal = "false"

	return cmd
}

// ConfigData and RepoRootConfig are defined in gen_config_discover.go

func (o *genConfigInitOptions) run(_ *cobra.Command, _ []string) error {
	// Check if output file already exists
	if _, err := os.Stat(o.outputFile); err == nil {
		if o.interactive {
			if !o.confirmOverwrite() {
				fmt.Println("Configuration generation cancelled.")
				return nil
			}
		} else {
			return fmt.Errorf("configuration file already exists: %s (use --interactive to confirm overwrite)", o.outputFile)
		}
	}

	var config ConfigData
	if o.interactive {
		config = o.runInteractiveWizard()
	} else {
		config = o.generateFromTemplate()
	}

	// Generate the YAML content
	yamlContent := o.generateYAMLContent(config)

	// Write to file
	err := os.WriteFile(o.outputFile, []byte(yamlContent), 0o600)
	if err != nil {
		return fmt.Errorf("failed to write configuration file: %w", err)
	}

	fmt.Printf("✅ Configuration file generated successfully: %s\n", o.outputFile)

	if o.interactive {
		fmt.Println("\nNext steps:")
		fmt.Println("1. Review and edit the configuration file as needed")
		fmt.Println("2. Set up authentication tokens as environment variables")
		fmt.Println("3. Generate SSH configuration if using SSH protocol:")
		fmt.Printf("   gz ssh-config generate --config %s --dry-run\n", o.outputFile)
		fmt.Println("4. Start cloning repositories:")
		fmt.Printf("   gz bulk-clone --config %s\n", o.outputFile)
	}

	return nil
}

func (o *genConfigInitOptions) confirmOverwrite() bool {
	fmt.Printf("Configuration file '%s' already exists. Overwrite? (y/N): ", o.outputFile)

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	response := strings.ToLower(strings.TrimSpace(scanner.Text()))

	return response == "y" || response == "yes"
}

func (o *genConfigInitOptions) runInteractiveWizard() ConfigData {
	scanner := bufio.NewScanner(os.Stdin)
	config := ConfigData{
		Version:   "0.1",
		Protocol:  "https",
		RepoRoots: []RepoRootConfig{},
		Ignores:   []string{},
	}

	fmt.Println("🚀 Bulk Clone Configuration Wizard")
	fmt.Println("===================================")
	fmt.Println()

	// Ask for default protocol
	fmt.Print("Default protocol (https/ssh) [https]: ")
	scanner.Scan()

	protocol := strings.TrimSpace(scanner.Text())
	if protocol == "" {
		protocol = "https"
	}

	if protocol != "https" && protocol != "ssh" && protocol != "http" {
		fmt.Println("⚠️  Invalid protocol, using https")

		protocol = "https"
	}

	config.Protocol = protocol

	fmt.Println()

	// Add repository roots
	fmt.Println("📁 Repository Configuration")
	fmt.Println("Add repository roots for different organizations/groups.")
	fmt.Println()

	for {
		fmt.Printf("Add a repository root? (Y/n): ")
		scanner.Scan()

		if strings.ToLower(strings.TrimSpace(scanner.Text())) == "n" {
			break
		}

		repoRoot := o.promptForRepoRoot(scanner, config.Protocol)
		if repoRoot.Provider != "" {
			config.RepoRoots = append(config.RepoRoots, repoRoot)
			fmt.Printf("✅ Added %s/%s -> %s\n\n", repoRoot.Provider, repoRoot.OrgName, repoRoot.RootPath)
		}
	}

	// Add ignore patterns
	fmt.Println("🚫 Ignore Patterns")
	fmt.Println("Add patterns for repositories to ignore (regex patterns).")
	fmt.Println("Common examples: test-.*, .*-archive, .*-deprecated")
	fmt.Println()

	for {
		fmt.Print("Add ignore pattern (empty to finish): ")
		scanner.Scan()

		pattern := strings.TrimSpace(scanner.Text())
		if pattern == "" {
			break
		}

		config.Ignores = append(config.Ignores, pattern)
		fmt.Printf("✅ Added ignore pattern: %s\n", pattern)
	}

	return config
}

func (o *genConfigInitOptions) promptForRepoRoot(scanner *bufio.Scanner, defaultProtocol string) RepoRootConfig {
	repoRoot := RepoRootConfig{}

	// Provider
	fmt.Print("Provider (github/gitlab/gitea): ")
	scanner.Scan()

	repoRoot.Provider = strings.ToLower(strings.TrimSpace(scanner.Text()))
	if repoRoot.Provider == "" {
		fmt.Println("⚠️  Provider is required")
		return RepoRootConfig{}
	}

	// Organization/Group name
	var prompt string

	switch repoRoot.Provider {
	case "gitlab":
		prompt = "GitLab group name: "
	default:
		prompt = "Organization name: "
	}

	fmt.Print(prompt)
	scanner.Scan()

	repoRoot.OrgName = strings.TrimSpace(scanner.Text())
	if repoRoot.OrgName == "" {
		fmt.Println("⚠️  Organization/group name is required")
		return RepoRootConfig{}
	}

	// Root path
	fmt.Printf("Target directory [$HOME/%s]: ", repoRoot.OrgName)
	scanner.Scan()

	repoRoot.RootPath = strings.TrimSpace(scanner.Text())
	if repoRoot.RootPath == "" {
		repoRoot.RootPath = fmt.Sprintf("$HOME/%s", repoRoot.OrgName)
	}

	// Protocol
	fmt.Printf("Protocol for this organization (https/ssh) [%s]: ", defaultProtocol)
	scanner.Scan()

	protocol := strings.TrimSpace(scanner.Text())
	if protocol == "" {
		protocol = defaultProtocol
	}

	repoRoot.Protocol = protocol

	return repoRoot
}

func (o *genConfigInitOptions) generateFromTemplate() ConfigData {
	switch o.template {
	case "comprehensive":
		return ConfigData{
			Version:  "0.1",
			Protocol: "https",
			RepoRoots: []RepoRootConfig{
				{Provider: "github", OrgName: "mycompany", RootPath: "$HOME/work/mycompany", Protocol: "ssh"},
				{Provider: "github", OrgName: "kubernetes", RootPath: "$HOME/opensource/kubernetes", Protocol: "https"},
				{Provider: "gitlab", OrgName: "mygroup", RootPath: "$HOME/gitlab/mygroup", Protocol: "ssh"},
			},
			Ignores: []string{"^test-.*", ".*-archive$", "^temp.*", ".*-deprecated$"},
		}
	default: // simple
		return ConfigData{
			Version:  "0.1",
			Protocol: "https",
			RepoRoots: []RepoRootConfig{
				{Provider: "github", OrgName: "myorg", RootPath: "$HOME/repos/myorg", Protocol: "https"},
			},
			Ignores: []string{"test-.*", ".*-archive"},
		}
	}
}

func (o *genConfigInitOptions) generateYAMLContent(config ConfigData) string {
	var content strings.Builder

	content.WriteString("# Generated by gzh-manager gen-config init\n")
	content.WriteString("# Configuration for bulk Git repository management\n\n")
	content.WriteString(fmt.Sprintf("version: \"%s\"\n\n", config.Version))

	// Default section
	content.WriteString("# Global default settings\n")
	content.WriteString("default:\n")
	content.WriteString(fmt.Sprintf("  protocol: %s\n", config.Protocol))
	content.WriteString("  github:\n")
	content.WriteString("    root_path: \"$HOME/github-repos\"\n")
	content.WriteString("  gitlab:\n")
	content.WriteString("    root_path: \"$HOME/gitlab-repos\"\n\n")

	// Repository roots
	content.WriteString("# Repository configurations\n")
	content.WriteString("repo_roots:\n")

	for _, root := range config.RepoRoots {
		content.WriteString(fmt.Sprintf("  - root_path: \"%s\"\n", root.RootPath))
		content.WriteString(fmt.Sprintf("    provider: \"%s\"\n", root.Provider))
		content.WriteString(fmt.Sprintf("    protocol: \"%s\"\n", root.Protocol))
		content.WriteString(fmt.Sprintf("    org_name: \"%s\"\n", root.OrgName))
		content.WriteString("\n")
	}

	// Ignore patterns
	if len(config.Ignores) > 0 {
		content.WriteString("# Ignore patterns for repositories\n")
		content.WriteString("ignore_names:\n")

		for _, pattern := range config.Ignores {
			content.WriteString(fmt.Sprintf("  - \"%s\"\n", pattern))
		}
	}

	return content.String()
}
