// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package profile

import (
	"github.com/spf13/cobra"
)

// NewCmd creates the main profile command that aggregates profile and quick commands.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Network profile management",
		Long: `Network profile management including profile creation, editing, and quick actions.

This command provides comprehensive network profile management with:
- Network profile creation and editing
- Profile import/export functionality
- Quick network actions and shortcuts
- Profile switching and management

Examples:
  # Profile management
  gz net-env profile list
  gz net-env profile create office
  gz net-env profile edit home

  # Quick actions
  gz net-env profile quick vpn-toggle
  gz net-env profile quick dns-reset
  gz net-env profile quick wifi-scan`,
		SilenceUsage: true,
	}

	// Add profile management command
	cmd.AddCommand(NewProfileCmd())

	// Add quick actions command
	cmd.AddCommand(NewQuickCmd())

	return cmd
}
