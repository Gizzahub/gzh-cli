// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package cmd

import (
	"github.com/spf13/cobra"

	repoconfig "github.com/gizzahub/gzh-manager-go/cmd/repo-config"
)

// NewGitCmd creates the unified git platform management command.
func NewGitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "git",
		Short: "🔗 통합 Git 플랫폼 관리 도구 (config, webhook, event)",
		Long: `Unified Git platform management tools for GitHub, GitLab, Gitea, and Gogs.

This command provides comprehensive Git platform management capabilities including:
- Repository configuration management
- Webhook management and automation
- Event processing and monitoring
- Cross-platform operations

Available Resources:
  config     Repository configuration management
  webhook    Webhook management and automation
  event      Event processing and monitoring

Examples:
  gz git config audit --org myorg --framework SOC2
  gz git webhook create --org myorg --repo myrepo --url https://example.com/webhook
  gz git event server --port 8080 --secret mysecret`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	// Add subcommands for each resource
	cmd.AddCommand(newGitConfigCmd())
	cmd.AddCommand(newGitWebhookCmd())
	cmd.AddCommand(newGitEventCmd())

	return cmd
}

// newGitConfigCmd creates the git config command (maps to repo-config).
func newGitConfigCmd() *cobra.Command {
	// Use existing repo-config implementation
	repoConfigCmd := repoconfig.NewRepoConfigCmd()

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

This command provides comprehensive webhook management capabilities including:
- Individual webhook CRUD operations
- Bulk webhook management across organizations
- Event-based automation and rules engine
- Webhook health monitoring and compliance

Examples:
  gz git webhook create --org myorg --repo myrepo --url https://example.com/webhook
  gz git webhook bulk create --org myorg --config webhooks.yaml
  gz git webhook automation --action deploy --rule security`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	// Get the existing webhook command from repo-config and add its subcommands
	repoWebhookCmd := repoconfig.NewRepoConfigCmd()
	for _, subCmd := range repoWebhookCmd.Commands() {
		if subCmd.Use == "webhook" {
			// Add all webhook subcommands to our git webhook command
			for _, webhookSubCmd := range subCmd.Commands() {
				cmd.AddCommand(webhookSubCmd)
			}
			break
		}
	}

	return cmd
}

// newGitEventCmd creates the git event command.
func newGitEventCmd() *cobra.Command {
	// Use existing event implementation
	eventCmd := NewEventCmd()

	// Update command metadata for git context
	eventCmd.Use = "event"
	eventCmd.Short = "Event processing and monitoring"
	eventCmd.Long = `Manage GitHub events, run webhook servers, and monitor event processing.

This command provides comprehensive event management capabilities including:
- Running webhook servers to receive GitHub events
- Querying and filtering stored events
- Managing event handlers and processors
- Monitoring event processing metrics

Examples:
  gz git event server --port 8080 --secret mysecret
  gz git event list --org myorg --type push --limit 50
  gz git event metrics --output json`

	return eventCmd
}
