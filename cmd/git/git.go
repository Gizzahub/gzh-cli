// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package git

import (
	"github.com/spf13/cobra"

	repopkg "github.com/gizzahub/gzh-cli/cmd/git/repo"
	webhookpkg "github.com/gizzahub/gzh-cli/cmd/git/webhook"
	repoconfig "github.com/gizzahub/gzh-cli/cmd/repo-config"
	"github.com/gizzahub/gzh-cli/internal/app"
)

// NewGitCmd creates the unified git platform management command.
func NewGitCmd(appCtx *app.AppContext) *cobra.Command {
	_ = appCtx
	cmd := &cobra.Command{
		Use:   "git",
		Short: "🔗 통합 Git 플랫폼 관리 도구 (config, webhook)",
		Long: `Unified Git platform management tools for GitHub, GitLab, Gitea, and Gogs.

This command provides Git platform management including:
- Repository lifecycle and configuration
- Webhook registration on the forge (CRUD)
- Cross-platform operations

This CLI does not receive or process forge webhooks. Register a publicly
reachable URL (GitHub Actions, Slack, your own service). GitHub cannot
POST to localhost or a laptop without a public address.

Available Resources:
  repo       Repository lifecycle management (clone, create, sync, etc.)
  config     Repository configuration management
  webhook    Webhook registration on GitHub/GitLab/Gitea

Examples:
  gz git repo clone --provider github --org myorg --target ./repos
  gz git config audit --org myorg --framework SOC2
  gz git webhook create --org myorg --repo myrepo --url https://example.com/webhook`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	// Add subcommands for each resource
	cmd.AddCommand(repopkg.NewGitRepoCmd())
	cmd.AddCommand(newGitConfigCmd(appCtx))
	cmd.AddCommand(newGitWebhookCmd())

	return cmd
}

// newGitConfigCmd creates the git config command (maps to repo-config).
func newGitConfigCmd(appCtx *app.AppContext) *cobra.Command {
	// Use existing repo-config implementation
	repoConfigCmd := repoconfig.NewRepoConfigCmd(appCtx)

	// Update command metadata for git context
	repoConfigCmd.Use = "config"
	repoConfigCmd.Short = "Repository configuration management"
	repoConfigCmd.Long = `Manage GitHub repository configurations at scale.

This command allows you to manage repository settings, security policies,
branch protection rules, and compliance auditing across entire organizations.

Examples:
  gz git config audit --org myorg --framework SOC2
  gz git config apply --config repo-config.yaml --dry-run
  gz git config diff --org myorg --repo myrepo`

	return repoConfigCmd
}

// newGitWebhookCmd creates the git webhook command.
func newGitWebhookCmd() *cobra.Command {
	// Create a wrapper command that delegates to repo-config webhook
	cmd := &cobra.Command{
		Use:   "webhook",
		Short: "Webhook management and automation",
		Long: `Manage repository and organization webhooks across Git platforms.

This command registers and manages webhook configurations on the forge.
It does not run a local receiver. The --url must be reachable from the
hosting site (not localhost).

Examples:
  gz git webhook create --org myorg --repo myrepo --url https://example.com/webhook
  gz git webhook bulk create --org myorg --config webhooks.yaml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	// Use webhook package implementation
	webhookCmd := webhookpkg.NewWebhookCmd()
	// Add all webhook subcommands to our git webhook command
	for _, webhookSubCmd := range webhookCmd.Commands() {
		cmd.AddCommand(webhookSubCmd)
	}

	return cmd
}
