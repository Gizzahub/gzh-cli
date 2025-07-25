// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package reposync

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// newWebhookCmd creates the webhook subcommand for repo-sync
func newWebhookCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "webhook",
		Short: "Manage GitHub webhooks",
		Long: `GitHub webhook CRUD API management tool

Manage repository and organization webhooks including creation, retrieval, updates, and deletion.
Provides bulk operations and webhook status monitoring capabilities.

Features:
• Individual webhook CRUD operations
• Organization-wide webhook batch configuration
• Webhook status monitoring and testing
• Webhook delivery history`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Show deprecation warning if called through old command
			if os.Getenv("GZ_DEPRECATED_COMMAND") == "webhook" {
				fmt.Fprintf(os.Stderr, "\nWarning: 'webhook' is deprecated and will be removed in v3.0.\n")
				fmt.Fprintf(os.Stderr, "Please use 'gz repo-sync webhook' instead.\n")
				fmt.Fprintf(os.Stderr, "Run 'gz help migrate' for more information.\n\n")
			}
			return cmd.Help()
		},
	}

	// Add subcommands (from webhook command)
	cmd.AddCommand(newWebhookRepositoryCmd())
	cmd.AddCommand(newWebhookOrganizationCmd())
	cmd.AddCommand(newWebhookBulkCmd())
	cmd.AddCommand(newWebhookConfigCmd())
	cmd.AddCommand(newWebhookMonitorCmd())

	return cmd
}