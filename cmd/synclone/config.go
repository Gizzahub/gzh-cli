// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package synclone

import (
	"github.com/spf13/cobra"

	"github.com/gizzahub/gzh-cli/internal/app"
)

// newSyncCloneConfigCmd creates the config subcommand for synclone.
func newSyncCloneConfigCmd(appCtx *app.AppContext) *cobra.Command {
	_ = appCtx
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
