// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

func newValidateCmd(ctx context.Context) *cobra.Command {
	var (
		verbose bool
		strict  bool
	)

	cmd := &cobra.Command{
		Use:   "validate [config-file]",
		Short: "Validate configuration and connectivity",
		Long: `Validate synclone configuration files and test connectivity to Git providers.

This command performs the following checks:
- Configuration file syntax and schema validation
- Provider authentication verification
- API connectivity tests
- Repository access permissions

Examples:
  # Validate configuration file
  git synclone validate synclone.yaml

  # Validate with verbose output
  git synclone validate synclone.yaml --verbose

  # Validate from standard locations
  git synclone validate

  # Strict validation (fail on warnings)
  git synclone validate --strict`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Delegate to gz synclone validate
			gzArgs := []string{"synclone", "validate"}

			if len(args) > 0 {
				gzArgs = append(gzArgs, args[0])
			}
			if verbose {
				gzArgs = append(gzArgs, "--verbose")
			}
			if strict {
				gzArgs = append(gzArgs, "--strict")
			}

			gzCmd := exec.CommandContext(ctx, "gz", gzArgs...)
			gzCmd.Stdout = os.Stdout
			gzCmd.Stderr = os.Stderr
			return gzCmd.Run()
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed validation output")
	cmd.Flags().BoolVar(&strict, "strict", false, "Treat warnings as errors")

	return cmd
}
