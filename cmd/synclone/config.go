// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package synclone

import (
	"github.com/spf13/cobra"
)

// newSyncCloneConfigCmd creates the config subcommand for synclone.
func newSyncCloneConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage synclone configuration files",
		Long: `Manage synclone configuration files including generation, validation, and conversion.

This command provides tools for working with synclone configuration files:
- Generate configurations from existing repositories
- Validate configuration syntax and structure
- Convert between configuration formats`,
	}

	// Add subcommands
	cmd.AddCommand(newConfigGenerateCmd())
	cmd.AddCommand(newConfigValidateCmd())
	cmd.AddCommand(newConfigConvertCmd())

	return cmd
}
