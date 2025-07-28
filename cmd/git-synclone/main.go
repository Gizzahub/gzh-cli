// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

var (
	// Version will be set by ldflags during build
	Version = "dev"
	// BuildTime will be set by ldflags during build
	BuildTime = "unknown"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	if err := Execute(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// Execute runs the main command
func Execute(ctx context.Context) error {
	rootCmd := newRootCmd(ctx)
	return rootCmd.ExecuteContext(ctx)
}

func newRootCmd(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "git-synclone",
		Short: "Enhanced Git cloning with provider awareness",
		Long: `git synclone provides intelligent repository cloning with support for
GitHub, GitLab, Gitea, and Gogs platforms. It offers bulk cloning, parallel
execution, and resume capabilities.

This is a Git extension that can be invoked as 'git synclone' when installed
in your PATH. It provides a more intuitive interface for cloning multiple
repositories from various Git hosting services.`,
		Version: fmt.Sprintf("%s (built %s)", Version, BuildTime),
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no subcommand is specified, show help
			return cmd.Help()
		},
	}

	// Add subcommands
	cmd.AddCommand(newGitHubCmd(ctx))
	cmd.AddCommand(newGitLabCmd(ctx))
	cmd.AddCommand(newGiteaCmd(ctx))
	cmd.AddCommand(newAllCmd(ctx))
	cmd.AddCommand(newConfigCmd(ctx))
	cmd.AddCommand(newStateCmd(ctx))
	cmd.AddCommand(newValidateCmd(ctx))
	cmd.AddCommand(newDoctorCmd())

	// Global flags that apply to all subcommands
	cmd.PersistentFlags().StringP("config", "c", "", "Path to configuration file")
	cmd.PersistentFlags().StringP("target", "t", "", "Target directory for cloning")
	cmd.PersistentFlags().IntP("parallel", "p", 10, "Number of parallel workers")
	cmd.PersistentFlags().Bool("resume", false, "Resume interrupted clone operation")
	cmd.PersistentFlags().Bool("cleanup-orphans", false, "Remove directories not in organization")
	cmd.PersistentFlags().String("strategy", "reset", "Clone strategy: reset, pull, or fetch")
	cmd.PersistentFlags().Bool("dry-run", false, "Show what would be done without executing")
	cmd.PersistentFlags().String("progress-mode", "bar", "Progress display: bar, dots, spinner, quiet")

	return cmd
}
