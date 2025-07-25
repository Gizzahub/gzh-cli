// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package synclone

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	
	// Import gen-config functionality - will be moved to pkg/synclone/config later
	genconfig "github.com/gizzahub/gzh-manager-go/cmd/gen-config"
)

// newConfigGenerateCmd creates the generate subcommand for config
func newConfigGenerateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate configuration from existing repositories",
		Long: `Generate synclone configuration files from existing repositories.

This command provides various ways to create configuration files:
- Interactive wizard for step-by-step configuration creation
- Predefined templates for common use cases
- Auto-discovery from existing repository directories
- GitHub organization cloning (legacy functionality)

Examples:
  # Interactive configuration creation
  gz synclone config generate init

  # Generate from template
  gz synclone config generate template simple

  # Auto-discover from existing repositories
  gz synclone config generate discover ~/projects --recursive`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// For now, delegate to gen-config functionality
			// TODO: Move this logic to pkg/synclone/config/generate.go
			if len(args) == 0 {
				return fmt.Errorf("subcommand required: init, template, discover, or github")
			}
			
			// Show deprecation notice if called through old command
			if os.Getenv("GZ_DEPRECATED_COMMAND") == "gen-config" {
				fmt.Fprintf(os.Stderr, "\nWarning: 'gen-config' is deprecated and will be removed in v3.0.\n")
				fmt.Fprintf(os.Stderr, "Please use 'gz synclone config generate' instead.\n")
				fmt.Fprintf(os.Stderr, "Run 'gz help migrate' for more information.\n\n")
			}
			
			return nil
		},
	}

	// Add subcommands that mirror gen-config functionality
	cmd.AddCommand(newConfigGenerateInitCmd())
	cmd.AddCommand(newConfigGenerateTemplateCmd())
	cmd.AddCommand(newConfigGenerateDiscoverCmd())
	cmd.AddCommand(newConfigGenerateGithubCmd())

	return cmd
}

func newConfigGenerateInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize configuration with interactive wizard",
		Long:  `Create a new synclone configuration file using an interactive wizard.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement interactive wizard
			// For now, use existing gen-config logic
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			
			// Temporary: Create gen-config command and run it
			genConfigCmd := genconfig.NewGenConfigCmd(ctx)
			initCmd, _, err := genConfigCmd.Find([]string{"init"})
			if err != nil {
				return err
			}
			
			return initCmd.RunE(initCmd, args)
		},
	}
}

func newConfigGenerateTemplateCmd() *cobra.Command {
	var outputFile string
	
	cmd := &cobra.Command{
		Use:   "template [template-name]",
		Short: "Generate configuration from template",
		Long: `Generate a synclone configuration file from predefined templates.

Available templates:
  simple     - Basic configuration for single organization
  multi      - Multi-organization setup
  enterprise - Enterprise configuration with advanced options`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement template generation
			// For now, use existing gen-config logic
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			
			genConfigCmd := genconfig.NewGenConfigCmd(ctx)
			templateCmd, _, err := genConfigCmd.Find([]string{"template"})
			if err != nil {
				return err
			}
			
			return templateCmd.RunE(templateCmd, args)
		},
	}
	
	cmd.Flags().StringVarP(&outputFile, "output", "o", "synclone.yaml", "Output file path")
	
	return cmd
}

func newConfigGenerateDiscoverCmd() *cobra.Command {
	var (
		outputFile string
		recursive  bool
		maxDepth   int
	)
	
	cmd := &cobra.Command{
		Use:   "discover [directory]",
		Short: "Discover repositories from directory",
		Long: `Auto-discover Git repositories from a directory and generate configuration.

This command scans the specified directory for Git repositories and creates
a synclone configuration file based on the discovered structure.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement discovery logic
			// For now, use existing gen-config logic
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			
			genConfigCmd := genconfig.NewGenConfigCmd(ctx)
			discoverCmd, _, err := genConfigCmd.Find([]string{"discover"})
			if err != nil {
				return err
			}
			
			return discoverCmd.RunE(discoverCmd, args)
		},
	}
	
	cmd.Flags().StringVarP(&outputFile, "output", "o", "synclone.yaml", "Output file path")
	cmd.Flags().BoolVarP(&recursive, "recursive", "r", true, "Scan directories recursively")
	cmd.Flags().IntVar(&maxDepth, "max-depth", 3, "Maximum directory depth for recursive scan")
	
	return cmd
}

func newConfigGenerateGithubCmd() *cobra.Command {
	var (
		outputFile string
		token      string
		targetDir  string
	)
	
	cmd := &cobra.Command{
		Use:        "github [organization]",
		Short:      "Generate configuration from GitHub organization (legacy)",
		Long:       `Generate configuration by fetching repository list from a GitHub organization.`,
		Deprecated: "Use 'gz synclone github' directly for GitHub operations",
		Args:       cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement GitHub org scanning
			// For now, use existing gen-config logic
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			
			genConfigCmd := genconfig.NewGenConfigCmd(ctx)
			githubCmd, _, err := genConfigCmd.Find([]string{"github"})
			if err != nil {
				return err
			}
			
			return githubCmd.RunE(githubCmd, args)
		},
	}
	
	cmd.Flags().StringVarP(&outputFile, "output", "o", "synclone.yaml", "Output file path")
	cmd.Flags().StringVar(&token, "token", "", "GitHub personal access token")
	cmd.Flags().StringVar(&targetDir, "target-dir", ".", "Target directory for organization")
	
	return cmd
}