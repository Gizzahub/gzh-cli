// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

type allOptions struct {
	// Configuration
	configFile string
	useConfig  bool

	// Provider filter
	providerFilter string

	// Common options (inherited from parent)
	target         string
	parallel       int
	strategy       string
	resume         bool
	cleanupOrphans bool
	dryRun         bool
	progressMode   string
}

func newAllCmd(ctx context.Context) *cobra.Command {
	opts := &allOptions{}

	cmd := &cobra.Command{
		Use:   "all",
		Short: "Clone repositories from all configured providers",
		Long: `Clone repositories from multiple providers using a configuration file.

The configuration file supports both the unified format (git.yaml) and the
legacy synclone format (synclone.yaml).

Examples:
  # Clone using configuration file
  git synclone all -c synclone.yaml

  # Use configuration from standard locations
  git synclone all --use-config

  # Clone only from specific provider
  git synclone all -c config.yaml --provider github`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get config file from parent flag if not set locally
			if opts.configFile == "" {
				opts.configFile, _ = cmd.Flags().GetString("config")
			}

			// Inherit common flags from parent
			opts.target, _ = cmd.Flags().GetString("target")
			opts.parallel, _ = cmd.Flags().GetInt("parallel")
			opts.strategy, _ = cmd.Flags().GetString("strategy")
			opts.resume, _ = cmd.Flags().GetBool("resume")
			opts.cleanupOrphans, _ = cmd.Flags().GetBool("cleanup-orphans")
			opts.dryRun, _ = cmd.Flags().GetBool("dry-run")
			opts.progressMode, _ = cmd.Flags().GetString("progress-mode")

			return runAll(ctx, opts)
		},
	}

	// Configuration flags
	cmd.Flags().BoolVar(&opts.useConfig, "use-config", false, "Use config file from standard locations")
	cmd.Flags().StringVar(&opts.providerFilter, "provider", "", "Filter by provider: github, gitlab, gitea")

	return cmd
}

func runAll(ctx context.Context, opts *allOptions) error {
	// For now, delegate to the existing gz synclone command
	// This ensures compatibility while we transition to the new structure

	fmt.Printf("🚀 Running synclone with configuration file\n")

	// Build command arguments for gz synclone
	args := []string{"synclone"}

	if opts.configFile != "" {
		args = append(args, "--config", opts.configFile)
	} else if opts.useConfig {
		args = append(args, "--use-config")
	} else {
		return fmt.Errorf("either --config or --use-config must be specified")
	}

	// Add other flags
	if opts.strategy != "" {
		args = append(args, "--strategy", opts.strategy)
	}
	if opts.parallel > 0 {
		args = append(args, "--parallel", fmt.Sprintf("%d", opts.parallel))
	}
	if opts.resume {
		args = append(args, "--resume")
	}
	if opts.cleanupOrphans {
		args = append(args, "--cleanup-orphans")
	}
	if opts.progressMode != "" {
		args = append(args, "--progress-mode", opts.progressMode)
	}
	if opts.providerFilter != "" {
		args = append(args, "--provider", opts.providerFilter)
	}

	fmt.Printf("   Command: gz %s\n", strings.Join(args, " "))

	if opts.dryRun {
		fmt.Printf("\n⚠️ DRY RUN MODE - Command would be executed but not actually running\n")
		return nil
	}

	// Execute gz synclone command
	cmd := exec.CommandContext(ctx, "gz", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}
