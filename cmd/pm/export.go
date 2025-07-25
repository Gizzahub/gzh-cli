// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package pm

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

func newExportCmd(ctx context.Context) *cobra.Command {
	var (
		allManagers bool
		manager     string
		outputDir   string
		format      string
	)

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export current installations to configuration files",
		Long: `Export currently installed packages to configuration files.

This command detects installed packages and generates configuration files
that can be used to recreate the same environment on another machine.

Examples:
  # Export all package managers
  gz pm export --all

  # Export specific package manager
  gz pm export --manager brew

  # Export to custom directory
  gz pm export --all --output ~/myconfigs/pm

  # Export in different format
  gz pm export --all --format json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runExport(ctx, allManagers, manager, outputDir, format)
		},
	}

	cmd.Flags().BoolVar(&allManagers, "all", false, "Export all package managers")
	cmd.Flags().StringVar(&manager, "manager", "", "Export specific package manager")
	cmd.Flags().StringVar(&outputDir, "output", "", "Output directory (default: ~/.gzh/pm)")
	cmd.Flags().StringVar(&format, "format", "yaml", "Export format: yaml, json")

	return cmd
}

func runExport(ctx context.Context, all bool, manager, outputDir, format string) error {
	if outputDir == "" {
		outputDir = "~/.gzh/pm"
	}

	fmt.Printf("Exporting package configurations to %s\n", outputDir)
	fmt.Printf("Format: %s\n", format)

	switch {
	case manager != "":
		fmt.Printf("Exporting %s packages...\n", manager)
	case all:
		fmt.Println("Exporting all package managers...")
	default:
		return fmt.Errorf("specify --manager or --all")
	}

	// TODO: Detect installed packages
	// TODO: Generate configuration files
	// TODO: Save to output directory

	return fmt.Errorf("export command not yet implemented")
}
