// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package pm

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

func newSyncCmd(ctx context.Context) *cobra.Command {
	var (
		cleanup       bool
		preserveExtra bool
		dryRun        bool
	)

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Synchronize installed packages with configuration",
		Long: `Synchronize installed packages to match configuration files.

This command ensures that your installed packages match what's defined in
the configuration files. It can add missing packages and optionally remove
packages not in the configuration.

Examples:
  # Sync packages (add missing only)
  gz pm sync

  # Sync and remove packages not in config
  gz pm sync --cleanup

  # Preserve extra packages not in config
  gz pm sync --preserve-extra

  # Dry run to see what would change
  gz pm sync --cleanup --dry-run`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSync(ctx, cleanup, preserveExtra, dryRun)
		},
	}

	cmd.Flags().BoolVar(&cleanup, "cleanup", false, "Remove packages not in configuration")
	cmd.Flags().BoolVar(&preserveExtra, "preserve-extra", false, "Keep packages not in configuration")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would change without making changes")

	return cmd
}

func runSync(ctx context.Context, cleanup, preserveExtra, dryRun bool) error {
	fmt.Println("Synchronizing packages with configuration...")
	if dryRun {
		fmt.Println("(dry run - no changes will be made)")
	}

	// TODO: Compare installed vs configured packages
	// TODO: Install missing packages
	// TODO: Optionally remove extra packages

	return fmt.Errorf("sync command not yet implemented")
}
