// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package reposync

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// newConfigCmd creates the config subcommand for repo-sync (from repo-config)
func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage repository configurations",
		Long: `Manage GitHub repository configurations across organizations.

This command provides tools for managing repository settings, security policies,
and compliance across entire GitHub organizations using infrastructure-as-code
principles.

Key Features:
- Apply consistent configuration across repositories
- Manage security policies and branch protection rules
- Template-based configuration management
- Compliance auditing and reporting
- Dry-run mode for safe changes`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Show deprecation warning if called through old command
			if os.Getenv("GZ_DEPRECATED_COMMAND") == "repo-config" {
				fmt.Fprintf(os.Stderr, "\nWarning: 'repo-config' is deprecated and will be removed in v3.0.\n")
				fmt.Fprintf(os.Stderr, "Please use 'gz repo-sync config' instead.\n")
				fmt.Fprintf(os.Stderr, "Run 'gz help migrate' for more information.\n\n")
			}
			return cmd.Help()
		},
	}

	// Add subcommands (from repo-config)
	cmd.AddCommand(newConfigListCmd())
	cmd.AddCommand(newConfigApplyCmd())
	cmd.AddCommand(newConfigValidateCmd())
	cmd.AddCommand(newConfigDiffCmd())
	cmd.AddCommand(newConfigAuditCmd())
	cmd.AddCommand(newConfigTemplateCmd())
	cmd.AddCommand(newConfigDashboardCmd())
	cmd.AddCommand(newConfigRiskAssessmentCmd())

	return cmd
}