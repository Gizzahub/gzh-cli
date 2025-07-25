// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package reposync

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// newEventCmd creates the event subcommand for repo-sync
func newEventCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "event",
		Short: "Manage GitHub events",
		Long: `Manage and process GitHub webhook events.

This command provides tools for:
- Starting a webhook event server
- Processing incoming webhook events
- Listing and filtering historical events
- Testing webhook event handling
- Event forwarding and transformation`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Show deprecation warning if called through old command
			if os.Getenv("GZ_DEPRECATED_COMMAND") == "event" {
				fmt.Fprintf(os.Stderr, "\nWarning: 'event' is deprecated and will be removed in v3.0.\n")
				fmt.Fprintf(os.Stderr, "Please use 'gz repo-sync event' instead.\n")
				fmt.Fprintf(os.Stderr, "Run 'gz help migrate' for more information.\n\n")
			}
			return cmd.Help()
		},
	}

	// Add subcommands
	cmd.AddCommand(newEventServerCmd())
	cmd.AddCommand(newEventListCmd())
	cmd.AddCommand(newEventProcessCmd())
	cmd.AddCommand(newEventTestCmd())
	cmd.AddCommand(newEventForwardCmd())

	return cmd
}
