// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package genconfig

import (
	"context"

	"github.com/spf13/cobra"
)

func NewGenConfigCmd(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "gen-config",
		Short:        "Generate bulk-clone configuration files",
		SilenceUsage: true,
		Long: `Generate and manage bulk-clone.yaml configuration files.

This command provides various ways to create configuration files for managing
multiple Git repositories across different hosting services:

- Interactive wizard for step-by-step configuration creation
- Predefined templates for common use cases
- Auto-discovery from existing repository directories
- GitHub organization cloning (legacy functionality)

Examples:
  # Interactive configuration creation
  gz gen-config init

  # Generate from template
  gz gen-config template simple

  # Auto-discover from existing repositories
  gz gen-config discover ~/projects --recursive`,
	}

	cmd.AddCommand(newGenConfigInitCmd())
	cmd.AddCommand(newGenConfigTemplateCmd())
	cmd.AddCommand(newGenConfigDiscoverCmd(ctx))
	cmd.AddCommand(newGenConfigGithubCmd()) // Keep for backward compatibility

	return cmd
}
