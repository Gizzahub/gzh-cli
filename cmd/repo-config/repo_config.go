// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package repoconfig

import (
	"github.com/spf13/cobra"

	"github.com/gizzahub/gzh-cli/cmd/repo-config/apply"
	"github.com/gizzahub/gzh-cli/cmd/repo-config/audit"
	"github.com/gizzahub/gzh-cli/cmd/repo-config/dashboard"
	"github.com/gizzahub/gzh-cli/cmd/repo-config/diff"
	"github.com/gizzahub/gzh-cli/cmd/repo-config/list"
	"github.com/gizzahub/gzh-cli/cmd/repo-config/risk"
	"github.com/gizzahub/gzh-cli/cmd/repo-config/template"
	"github.com/gizzahub/gzh-cli/cmd/repo-config/validate"
	"github.com/gizzahub/gzh-cli/cmd/repo-config/webhook"
	"github.com/gizzahub/gzh-cli/internal/app"
)

// NewRepoConfigCmd creates the repo-config command with subcommands.
func NewRepoConfigCmd(appCtx *app.AppContext) *cobra.Command {
	_ = appCtx
	cmd := &cobra.Command{
		Use:   "repo-config",
		Short: "GitHub repository configuration management",
		Long: `Manage GitHub repository configurations across organizations.

This command provides tools for managing repository settings, security policies,
and compliance across entire GitHub organizations using infrastructure-as-code
principles.

Key Features:
- Apply consistent configuration across repositories
- Manage security policies and branch protection rules
- Template-based configuration management
- Compliance auditing and reporting
- Dry-run mode for safe changes

Examples:
  gz repo-config list                    # List repositories with current settings
  gz repo-config apply                   # Apply configuration to repositories
  gz repo-config validate               # Validate configuration files
  gz repo-config diff                   # Show differences between current and target
  gz repo-config audit                  # Generate compliance audit report
  gz repo-config webhook                # Manage repository webhooks
  gz repo-config dashboard              # Start real-time compliance dashboard
  gz repo-config risk-assessment        # Perform CVSS-based risk assessment`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	// Add subcommands
	cmd.AddCommand(list.NewCmd())
	cmd.AddCommand(apply.NewCmd())
	cmd.AddCommand(validate.NewCmd())
	cmd.AddCommand(diff.NewCmd())
	cmd.AddCommand(audit.NewCmd())
	cmd.AddCommand(template.NewCmd())
	cmd.AddCommand(webhook.NewCmd())
	cmd.AddCommand(dashboard.NewCmd())
	cmd.AddCommand(risk.NewCmd())

	return cmd
}

// Global types and functions are now defined in shared_types.go
