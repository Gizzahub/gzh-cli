// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package pm

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

func newInstallCmd(ctx context.Context) *cobra.Command {
	var (
		manager  string
		strategy string
		force    bool
	)

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install packages from configuration files",
		Long: `Install packages based on configuration files in ~/.gzh/pm/.

This command reads the configuration files and installs all specified packages
with their configured versions.

Examples:
  # Install packages from all configured managers
  gz pm install

  # Install from specific package manager
  gz pm install --manager brew

  # Install with specific strategy
  gz pm install --strategy strict

  # Force reinstall
  gz pm install --force`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInstall(ctx, manager, strategy, force)
		},
	}

	cmd.Flags().StringVar(&manager, "manager", "", "Install packages for specific manager only")
	cmd.Flags().StringVar(&strategy, "strategy", "preserve", "Installation strategy: preserve, strict, latest")
	cmd.Flags().BoolVar(&force, "force", false, "Force reinstall even if already installed")

	return cmd
}

func runInstall(ctx context.Context, manager, strategy string, force bool) error {
	fmt.Printf("Installing packages with strategy: %s\n", strategy)
	if force {
		fmt.Println("Force reinstall enabled")
	}

	// TODO: Read configuration files from ~/.gzh/pm/
	// TODO: Install packages based on configuration

	if manager != "" {
		fmt.Printf("Installing packages for %s...\n", manager)
	} else {
		fmt.Println("Installing packages from all configured managers...")
	}

	return fmt.Errorf("install command not yet implemented")
}
