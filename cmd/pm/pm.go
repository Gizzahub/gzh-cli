// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

// Package pm provides package manager commands for managing various package managers
// including Homebrew, asdf, SDKMAN, and others. It offers functionality for updating,
// syncing, installing, and managing packages across different systems.
package pm

import (
	"context"

	"github.com/spf13/cobra"
)

// NewPMCmd creates the package manager command for unified package management.
func NewPMCmd(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pm",
		Short: "Package manager operations",
		Long: `Manage multiple package managers with unified commands.

This command provides centralized management for multiple package managers including:
- System package managers: brew, apt, port, yum, dnf, pacman
- Version managers: asdf, rbenv, pyenv, nvm, sdkman
- Language package managers: pip, gem, npm, cargo, go, composer

Examples:
  # Show status of all package managers
  gz pm status

  # Install packages from configuration
  gz pm install

  # Update all packages
  gz pm update --all

  # Export current installations
  gz pm export --all

  # Bootstrap missing package managers
  gz pm bootstrap

  # Manage package manager caches
  gz pm cache status
  gz pm cache clean --go --npm

For detailed configuration, see: ~/.gzh/pm/`,
	}

	// Register subcommands
	cmd.AddCommand(newStatusCmd(ctx))
	cmd.AddCommand(newInstallCmd(ctx))
	cmd.AddCommand(newUpdateCmd(ctx))
	cmd.AddCommand(newExportCmd(ctx))
	cmd.AddCommand(newBootstrapCmd(ctx))
	cmd.AddCommand(newUpgradeManagersCmd(ctx))
	cmd.AddCommand(newSyncVersionsCmd(ctx))
	cmd.AddCommand(newDoctorCmd(ctx))
	cmd.AddCommand(newCacheCmd(ctx))

	return cmd
}
