// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package synclone

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newConfigGenerateCmd creates the generate subcommand for config.
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
			if len(args) == 0 {
				return fmt.Errorf("subcommand required: init, template, discover, or github")
			}

			return nil
		},
	}

	// Add subcommands
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
			return fmt.Errorf("config generate init: not yet implemented")
		},
	}
}

// newConfigGenerateTemplateCmd is implemented in config_generate_template.go

// newConfigGenerateDiscoverCmd is implemented in config_generate_discover.go

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
			org := args[0]
			return fmt.Errorf("config generate github %s: not yet implemented", org)
		},
	}

	cmd.Flags().StringVarP(&outputFile, "output", "o", "synclone.yaml", "Output file path")
	cmd.Flags().StringVar(&token, "token", "", "GitHub personal access token")
	cmd.Flags().StringVar(&targetDir, "target-dir", ".", "Target directory for organization")

	return cmd
}
