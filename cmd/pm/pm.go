// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

// Package pm provides package manager commands for managing various package managers
// including Homebrew, asdf, SDKMAN, and others. It offers functionality for updating,
// syncing, installing, and managing packages across different systems.
package pm

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/Gizzahub/gzh-cli/cmd/pm/advanced"
	"github.com/Gizzahub/gzh-cli/cmd/pm/cache"
	"github.com/Gizzahub/gzh-cli/cmd/pm/doctor"
	"github.com/Gizzahub/gzh-cli/cmd/pm/export"
	"github.com/Gizzahub/gzh-cli/cmd/pm/install"
	"github.com/Gizzahub/gzh-cli/cmd/pm/status"
	"github.com/Gizzahub/gzh-cli/cmd/pm/update"
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
	cmd.AddCommand(status.NewStatusCmd(ctx))
	cmd.AddCommand(install.NewInstallCmd(ctx))
	cmd.AddCommand(update.NewUpdateCmd(ctx))
	cmd.AddCommand(export.NewExportCmd(ctx))
	cmd.AddCommand(advanced.NewBootstrapCmd(ctx))
	cmd.AddCommand(advanced.NewUpgradeManagersCmd(ctx))
	cmd.AddCommand(advanced.NewSyncVersionsCmd(ctx))
	cmd.AddCommand(doctor.NewDoctorCmd(ctx))
	cmd.AddCommand(cache.NewCacheCmd(ctx))

	return cmd
}
