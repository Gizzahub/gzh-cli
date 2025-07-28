// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

func newConfigCmd(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Configuration management commands",
		Long:  `Commands for managing synclone configuration files.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(newConfigGenerateCmd(ctx))
	cmd.AddCommand(newConfigValidateCmd(ctx))
	cmd.AddCommand(newConfigConvertCmd(ctx))

	return cmd
}

func newConfigGenerateCmd(ctx context.Context) *cobra.Command {
	var (
		output   string
		format   string
		provider string
	)

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate a sample configuration file",
		Long: `Generate a sample configuration file with example settings.

Examples:
  # Generate a basic configuration
  git synclone config generate -o synclone.yaml

  # Generate configuration for specific provider
  git synclone config generate -o github.yaml --provider github

  # Generate in legacy format
  git synclone config generate -o legacy.yaml --format legacy`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// For now, delegate to gz synclone config generate
			gzArgs := []string{"synclone", "config", "generate"}

			if output != "" {
				gzArgs = append(gzArgs, "-o", output)
			}
			if format != "" {
				gzArgs = append(gzArgs, "--format", format)
			}
			if provider != "" {
				gzArgs = append(gzArgs, "--provider", provider)
			}

			gzCmd := exec.CommandContext(ctx, "gz", gzArgs...)
			gzCmd.Stdout = os.Stdout
			gzCmd.Stderr = os.Stderr
			return gzCmd.Run()
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "synclone.yaml", "Output file path (use '-' for stdout)")
	cmd.Flags().StringVar(&format, "format", "unified", "Configuration format: unified or legacy")
	cmd.Flags().StringVar(&provider, "provider", "", "Generate config for specific provider")

	return cmd
}

func newConfigValidateCmd(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate [config-file]",
		Short: "Validate a configuration file",
		Long: `Validate a synclone configuration file for syntax and semantic errors.

Examples:
  # Validate a configuration file
  git synclone config validate synclone.yaml

  # Validate configuration from standard locations
  git synclone config validate`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Delegate to gz synclone validate
			gzArgs := []string{"synclone", "validate"}
			if len(args) > 0 {
				gzArgs = append(gzArgs, args[0])
			}

			gzCmd := exec.CommandContext(ctx, "gz", gzArgs...)
			gzCmd.Stdout = os.Stdout
			gzCmd.Stderr = os.Stderr
			return gzCmd.Run()
		},
	}

	return cmd
}

func newConfigConvertCmd(ctx context.Context) *cobra.Command {
	var (
		output string
		format string
		force  bool
	)

	cmd := &cobra.Command{
		Use:   "convert [config-file]",
		Short: "Convert configuration between formats",
		Long: `Convert synclone configuration files between different formats.

Examples:
  # Convert legacy format to unified format
  git synclone config convert legacy.yaml -o unified.yaml

  # Convert to stdout
  git synclone config convert config.yaml -o -`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Delegate to gz synclone config convert
			gzArgs := []string{"synclone", "config", "convert", args[0]}

			if output != "" {
				gzArgs = append(gzArgs, "-o", output)
			}
			if format != "" {
				gzArgs = append(gzArgs, "--format", format)
			}
			if force {
				gzArgs = append(gzArgs, "--force")
			}

			gzCmd := exec.CommandContext(ctx, "gz", gzArgs...)
			gzCmd.Stdout = os.Stdout
			gzCmd.Stderr = os.Stderr
			return gzCmd.Run()
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "Output file path (use '-' for stdout)")
	cmd.Flags().StringVar(&format, "format", "unified", "Target format: unified or legacy")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing file")

	cmd.MarkFlagRequired("output")

	return cmd
}
