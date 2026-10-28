// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

// Package webhook registers and manages forge webhook configurations.
package webhook

import (
	"fmt"

	"github.com/google/go-github/v66/github"
	"github.com/spf13/cobra"
)

// GlobalFlags represents global flags for all repo-config commands.
type GlobalFlags struct {
	Organization string
	ConfigFile   string
	Token        string
	DryRun       bool
	Verbose      bool
	Parallel     int
	Timeout      string
}

// WebhookFlags represents webhook command flags.
type WebhookFlags struct {
	GlobalFlags
	Repository   string
	URL          string
	Events       []string
	Active       bool
	Secret       string
	ContentType  string
	ID           int64
	OutputFormat string
}

type webhookDisplayConfig struct {
	ContentType *string `json:"content_type,omitempty"`
	InsecureSSL *string `json:"insecure_ssl,omitempty"`
	URL         *string `json:"url,omitempty"`
}

type webhookDisplay struct {
	CreatedAt    *github.Timestamp      `json:"created_at,omitempty"`
	UpdatedAt    *github.Timestamp      `json:"updated_at,omitempty"`
	URL          *string                `json:"url,omitempty"`
	ID           *int64                 `json:"id,omitempty"`
	Type         *string                `json:"type,omitempty"`
	Name         *string                `json:"name,omitempty"`
	TestURL      *string                `json:"test_url,omitempty"`
	PingURL      *string                `json:"ping_url,omitempty"`
	LastResponse map[string]interface{} `json:"last_response,omitempty"`
	Config       *webhookDisplayConfig  `json:"config,omitempty"`
	Events       []string               `json:"events,omitempty"`
	Active       *bool                  `json:"active,omitempty"`
}

type webhookConfigDisplay struct {
	URL         string   `json:"url"`
	Events      []string `json:"events"`
	Active      *bool    `json:"active,omitempty"`
	ContentType string   `json:"contentType,omitempty"`
}

// addGlobalFlags adds common flags to the command.
func addGlobalFlags(cmd *cobra.Command, flags *GlobalFlags) {
	cmd.Flags().StringVar(&flags.Organization, "org", "", "GitHub organization name")
	cmd.Flags().StringVar(&flags.ConfigFile, "config-file", "", "Configuration file path")
	cmd.Flags().StringVar(&flags.Token, "token", "", "GitHub token")
	cmd.Flags().BoolVar(&flags.DryRun, "dry-run", false, "Show what would be done without making changes")
	cmd.Flags().BoolVar(&flags.Verbose, "verbose", false, "Enable verbose output")
	cmd.Flags().IntVar(&flags.Parallel, "parallel", 5, "Number of parallel operations")
	cmd.Flags().StringVar(&flags.Timeout, "timeout", "30s", "Request timeout")
}

// truncateString truncates a string to a maximum length.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// NewCmd creates the webhook management command.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "webhook",
		Short: "Manage repository webhooks",
		Long: `Create, read, update, and delete repository webhooks.

This command provides comprehensive webhook management capabilities for GitHub repositories,
allowing you to manage webhooks across individual repositories or entire organizations.

Examples:
  gz repo-config webhook list --org myorg --repo myrepo
  gz repo-config webhook create --repo myrepo --url https://example.com/webhook --events push,pull_request
  gz repo-config webhook update --repo myrepo --id 12345 --events push,issues
  gz repo-config webhook delete --repo myrepo --id 12345
  gz repo-config webhook bulk create --org myorg --all --url https://example.com/webhook`,
	}

	// Add subcommands
	cmd.AddCommand(newWebhookListCmd())
	cmd.AddCommand(newWebhookCreateCmd())
	cmd.AddCommand(newWebhookUpdateCmd())
	cmd.AddCommand(newWebhookDeleteCmd())
	cmd.AddCommand(newWebhookGetCmd())
	cmd.AddCommand(newWebhookBulkCmd())

	return cmd
}

// addWebhookFlags adds common webhook flags to a command.
func addWebhookFlags(cmd *cobra.Command, flags *WebhookFlags) {
	addGlobalFlags(cmd, &flags.GlobalFlags)
	cmd.Flags().StringVar(&flags.Repository, "repo", "", "Repository name")
	cmd.Flags().StringVar(&flags.OutputFormat, "output", "table", "Output format (table, json, yaml)")
}

// newWebhookListCmd creates the webhook list command.
func newWebhookListCmd() *cobra.Command {
	flags := &WebhookFlags{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List webhooks for a repository",
		Long: `List all webhooks configured for a specific repository.

Examples:
  gz repo-config webhook list --org myorg --repo myrepo
  gz repo-config webhook list --org myorg --repo myrepo --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWebhookList(cmd.Context(), flags)
		},
	}

	addGlobalFlags(cmd, &flags.GlobalFlags)
	cmd.Flags().StringVar(&flags.Repository, "repo", "", "Repository name (required)")
	cmd.Flags().StringVar(&flags.OutputFormat, "output", "table", "Output format (table, json, yaml)")

	if err := cmd.MarkFlagRequired("repo"); err != nil {
		fmt.Printf("Warning: failed to mark repo flag as required: %v\n", err)
	}
	if err := cmd.MarkFlagRequired("org"); err != nil {
		fmt.Printf("Warning: failed to mark org flag as required: %v\n", err)
	}

	return cmd
}

// newWebhookCreateCmd creates the webhook create command.
func newWebhookCreateCmd() *cobra.Command {
	flags := &WebhookFlags{}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new webhook",
		Long: `Create a new webhook for a repository.

Examples:
  gz repo-config webhook create --repo myrepo --url https://example.com/webhook --events push
  gz repo-config webhook create --repo myrepo --url https://example.com/webhook --events push,pull_request --secret mysecret`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWebhookCreate(cmd.Context(), flags)
		},
	}

	addGlobalFlags(cmd, &flags.GlobalFlags)
	cmd.Flags().StringVar(&flags.Repository, "repo", "", "Repository name (required)")
	cmd.Flags().StringVar(&flags.URL, "url", "", "Webhook URL (required)")
	cmd.Flags().StringSliceVar(&flags.Events, "events", []string{"push"}, "Webhook events")
	cmd.Flags().BoolVar(&flags.Active, "active", true, "Whether webhook is active")
	cmd.Flags().StringVar(&flags.Secret, "secret", "", "Webhook secret")
	cmd.Flags().StringVar(&flags.ContentType, "content-type", "json", "Content type (json or form)")

	if err := cmd.MarkFlagRequired("repo"); err != nil {
		fmt.Printf("Warning: failed to mark repo flag as required: %v\n", err)
	}
	if err := cmd.MarkFlagRequired("org"); err != nil {
		fmt.Printf("Warning: failed to mark org flag as required: %v\n", err)
	}
	if err := cmd.MarkFlagRequired("url"); err != nil {
		fmt.Printf("Warning: failed to mark url flag as required: %v\n", err)
	}

	return cmd
}

// newWebhookUpdateCmd creates the webhook update command.
func newWebhookUpdateCmd() *cobra.Command {
	flags := &WebhookFlags{}
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update an existing webhook",
		Long: `Update an existing webhook for a repository.

Examples:
  gz repo-config webhook update --repo myrepo --id 12345 --url https://newurl.com/webhook
  gz repo-config webhook update --repo myrepo --id 12345 --events push,issues --active=false`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWebhookUpdate(cmd.Context(), flags)
		},
	}

	addGlobalFlags(cmd, &flags.GlobalFlags)
	cmd.Flags().StringVar(&flags.Repository, "repo", "", "Repository name (required)")
	cmd.Flags().Int64Var(&flags.ID, "id", 0, "Webhook ID (required)")
	cmd.Flags().StringVar(&flags.URL, "url", "", "Webhook URL")
	cmd.Flags().StringSliceVar(&flags.Events, "events", nil, "Webhook events")
	cmd.Flags().BoolVar(&flags.Active, "active", true, "Whether webhook is active")
	cmd.Flags().StringVar(&flags.Secret, "secret", "", "Webhook secret")
	cmd.Flags().StringVar(&flags.ContentType, "content-type", "", "Content type (json or form)")

	if err := cmd.MarkFlagRequired("repo"); err != nil {
		fmt.Printf("Warning: failed to mark repo flag as required: %v\n", err)
	}
	if err := cmd.MarkFlagRequired("org"); err != nil {
		fmt.Printf("Warning: failed to mark org flag as required: %v\n", err)
	}
	if err := cmd.MarkFlagRequired("id"); err != nil {
		fmt.Printf("Warning: failed to mark id flag as required: %v\n", err)
	}

	return cmd
}

// newWebhookDeleteCmd creates the webhook delete command.
func newWebhookDeleteCmd() *cobra.Command {
	flags := &WebhookFlags{}
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a webhook",
		Long: `Delete an existing webhook from a repository.

Examples:
  gz repo-config webhook delete --repo myrepo --id 12345`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWebhookDelete(cmd.Context(), flags)
		},
	}

	addGlobalFlags(cmd, &flags.GlobalFlags)
	cmd.Flags().StringVar(&flags.Repository, "repo", "", "Repository name (required)")
	cmd.Flags().Int64Var(&flags.ID, "id", 0, "Webhook ID (required)")

	if err := cmd.MarkFlagRequired("repo"); err != nil {
		fmt.Printf("Warning: failed to mark repo flag as required: %v\n", err)
	}
	if err := cmd.MarkFlagRequired("org"); err != nil {
		fmt.Printf("Warning: failed to mark org flag as required: %v\n", err)
	}
	if err := cmd.MarkFlagRequired("id"); err != nil {
		fmt.Printf("Warning: failed to mark id flag as required: %v\n", err)
	}

	return cmd
}

// newWebhookGetCmd creates the webhook get command.
func newWebhookGetCmd() *cobra.Command {
	flags := &WebhookFlags{}
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get details of a specific webhook",
		Long: `Get detailed information about a specific webhook.

Examples:
  gz repo-config webhook get --repo myrepo --id 12345
  gz repo-config webhook get --repo myrepo --id 12345 --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWebhookGet(cmd.Context(), flags)
		},
	}

	addGlobalFlags(cmd, &flags.GlobalFlags)
	cmd.Flags().StringVar(&flags.Repository, "repo", "", "Repository name (required)")
	cmd.Flags().Int64Var(&flags.ID, "id", 0, "Webhook ID (required)")
	cmd.Flags().StringVar(&flags.OutputFormat, "output", "table", "Output format (table, json, yaml)")

	if err := cmd.MarkFlagRequired("repo"); err != nil {
		fmt.Printf("Warning: failed to mark repo flag as required: %v\n", err)
	}
	if err := cmd.MarkFlagRequired("org"); err != nil {
		fmt.Printf("Warning: failed to mark org flag as required: %v\n", err)
	}
	if err := cmd.MarkFlagRequired("id"); err != nil {
		fmt.Printf("Warning: failed to mark id flag as required: %v\n", err)
	}

	return cmd
}
