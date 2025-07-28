// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

func newStateCmd(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "state",
		Short: "Manage synclone operation state",
		Long:  `Commands for managing the state of synclone operations, including resume functionality.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(newStateListCmd(ctx))
	cmd.AddCommand(newStateShowCmd(ctx))
	cmd.AddCommand(newStateCleanCmd(ctx))
	cmd.AddCommand(newStateResumeCmd(ctx))

	return cmd
}

func newStateListCmd(ctx context.Context) *cobra.Command {
	var (
		format string
		all    bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List saved operation states",
		Long: `List all saved synclone operation states that can be resumed.

Examples:
  # List active states
  git synclone state list

  # List all states including completed
  git synclone state list --all

  # List in JSON format
  git synclone state list --format json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Delegate to gz synclone state list
			gzArgs := []string{"synclone", "state", "list"}

			if all {
				gzArgs = append(gzArgs, "--all")
			}
			if format != "" {
				gzArgs = append(gzArgs, "--format", format)
			}

			gzCmd := exec.CommandContext(ctx, "gz", gzArgs...)
			gzCmd.Stdout = os.Stdout
			gzCmd.Stderr = os.Stderr
			return gzCmd.Run()
		},
	}

	cmd.Flags().StringVar(&format, "format", "table", "Output format: table or json")
	cmd.Flags().BoolVar(&all, "all", false, "Include completed states")

	return cmd
}

func newStateShowCmd(ctx context.Context) *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "show [state-id]",
		Short: "Show details of a specific state",
		Long: `Show detailed information about a specific synclone operation state.

Examples:
  # Show state details
  git synclone state show abc123

  # Show in JSON format
  git synclone state show abc123 --format json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Delegate to gz synclone state show
			gzArgs := []string{"synclone", "state", "show", args[0]}

			if format != "" {
				gzArgs = append(gzArgs, "--format", format)
			}

			gzCmd := exec.CommandContext(ctx, "gz", gzArgs...)
			gzCmd.Stdout = os.Stdout
			gzCmd.Stderr = os.Stderr
			return gzCmd.Run()
		},
	}

	cmd.Flags().StringVar(&format, "format", "text", "Output format: text, json, or yaml")

	return cmd
}

func newStateCleanCmd(ctx context.Context) *cobra.Command {
	var (
		all       bool
		completed bool
		older     string
		force     bool
	)

	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Clean up saved states",
		Long: `Clean up saved synclone operation states.

Examples:
  # Clean completed states
  git synclone state clean --completed

  # Clean states older than 7 days
  git synclone state clean --older 7d

  # Clean all states (requires confirmation)
  git synclone state clean --all`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Delegate to gz synclone state clean
			gzArgs := []string{"synclone", "state", "clean"}

			if all {
				gzArgs = append(gzArgs, "--all")
			}
			if completed {
				gzArgs = append(gzArgs, "--completed")
			}
			if older != "" {
				gzArgs = append(gzArgs, "--older", older)
			}
			if force {
				gzArgs = append(gzArgs, "--force")
			}

			gzCmd := exec.CommandContext(ctx, "gz", gzArgs...)
			gzCmd.Stdout = os.Stdout
			gzCmd.Stderr = os.Stderr
			gzCmd.Stdin = os.Stdin
			return gzCmd.Run()
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "Clean all states")
	cmd.Flags().BoolVar(&completed, "completed", false, "Clean only completed states")
	cmd.Flags().StringVar(&older, "older", "", "Clean states older than duration (e.g., 7d, 24h)")
	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation prompt")

	return cmd
}

func newStateResumeCmd(ctx context.Context) *cobra.Command {
	var (
		latest   bool
		parallel int
		strategy string
		dryRun   bool
	)

	cmd := &cobra.Command{
		Use:   "resume [state-id]",
		Short: "Resume an interrupted operation",
		Long: `Resume an interrupted synclone operation from a saved state.

Examples:
  # Resume specific state
  git synclone state resume abc123

  # Resume the latest incomplete state
  git synclone state resume --latest

  # Resume with different settings
  git synclone state resume abc123 --parallel 20`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Delegate to gz synclone state resume
			gzArgs := []string{"synclone", "state", "resume"}

			if len(args) > 0 {
				gzArgs = append(gzArgs, args[0])
			}
			if latest {
				gzArgs = append(gzArgs, "--latest")
			}
			if parallel > 0 {
				gzArgs = append(gzArgs, "--parallel", fmt.Sprintf("%d", parallel))
			}
			if strategy != "" {
				gzArgs = append(gzArgs, "--strategy", strategy)
			}
			if dryRun {
				gzArgs = append(gzArgs, "--dry-run")
			}

			gzCmd := exec.CommandContext(ctx, "gz", gzArgs...)
			gzCmd.Stdout = os.Stdout
			gzCmd.Stderr = os.Stderr
			return gzCmd.Run()
		},
	}

	cmd.Flags().BoolVar(&latest, "latest", false, "Resume the latest incomplete state")
	cmd.Flags().IntVarP(&parallel, "parallel", "p", 0, "Override parallel workers (0 = use saved value)")
	cmd.Flags().StringVar(&strategy, "strategy", "", "Override clone strategy")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be resumed without executing")

	return cmd
}
