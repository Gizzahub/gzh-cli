// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package pm

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

func newCleanCmd(ctx context.Context) *cobra.Command {
	var (
		dryRun bool
		force  bool
	)

	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Clean unused packages based on strategy",
		Long:  `Remove unused packages, caches, and old versions based on configured cleanup strategy.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Cleaning unused packages...")
			if dryRun {
				fmt.Println("(dry run - no changes will be made)")
			}
			return fmt.Errorf("clean command not yet implemented")
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be cleaned")
	cmd.Flags().BoolVar(&force, "force", false, "Force cleanup without confirmation")

	return cmd
}
