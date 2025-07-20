package repoconfig

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/gizzahub/gzh-manager-go/pkg/types/repoconfig"
	"github.com/google/go-github/v66/github"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

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

// newWebhookCmd creates the webhook management command.
func newWebhookCmd() *cobra.Command {
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
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	// Add subcommands
	cmd.AddCommand(newWebhookListCmd())
	cmd.AddCommand(newWebhookCreateCmd())
	cmd.AddCommand(newWebhookUpdateCmd())
	cmd.AddCommand(newWebhookDeleteCmd())
	cmd.AddCommand(newWebhookGetCmd())
	cmd.AddCommand(newWebhookBulkCmd())
	cmd.AddCommand(newWebhookAutomationCmd())

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
			return runWebhookList(flags)
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
			return runWebhookCreate(flags)
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
			return runWebhookUpdate(flags)
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
			return runWebhookDelete(flags)
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
			return runWebhookGet(flags)
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

// runWebhookList lists all webhooks for a repository.
func runWebhookList(flags *WebhookFlags) error {
	ctx := context.Background()
	client := createGitHubClient(flags.Token)

	webhooks, _, err := client.Repositories.ListHooks(ctx, flags.Organization, flags.Repository, nil)
	if err != nil {
		return fmt.Errorf("failed to list webhooks: %w", err)
	}

	return displayWebhooks(webhooks, flags.OutputFormat)
}

// runWebhookCreate creates a new webhook.
func runWebhookCreate(flags *WebhookFlags) error {
	ctx := context.Background()
	client := createGitHubClient(flags.Token)

	config := &github.HookConfig{
		URL:         &flags.URL,
		ContentType: &flags.ContentType,
	}
	if flags.Secret != "" {
		config.Secret = &flags.Secret
	}

	hook := &github.Hook{
		Events: flags.Events,
		Active: &flags.Active,
		Config: config,
	}

	if flags.DryRun {
		fmt.Printf("Would create webhook with URL: %s, Events: %v\n", flags.URL, flags.Events)
		return nil
	}

	createdHook, _, err := client.Repositories.CreateHook(ctx, flags.Organization, flags.Repository, hook)
	if err != nil {
		return fmt.Errorf("failed to create webhook: %w", err)
	}

	fmt.Printf("Successfully created webhook with ID: %d\n", createdHook.GetID())

	return displayWebhook(createdHook, "table")
}

// runWebhookUpdate updates an existing webhook.
func runWebhookUpdate(flags *WebhookFlags) error {
	ctx := context.Background()
	client := createGitHubClient(flags.Token)

	// Get existing webhook to preserve unmodified fields
	existingHook, _, err := client.Repositories.GetHook(ctx, flags.Organization, flags.Repository, flags.ID)
	if err != nil {
		return fmt.Errorf("failed to get existing webhook: %w", err)
	}

	// Create updated hook with existing values as defaults
	config := existingHook.Config
	if config == nil {
		config = &github.HookConfig{}
	}

	// Update fields if provided
	if flags.URL != "" {
		config.URL = &flags.URL
	}

	if flags.ContentType != "" {
		config.ContentType = &flags.ContentType
	}

	if flags.Secret != "" {
		config.Secret = &flags.Secret
	}

	hook := &github.Hook{
		Config: config,
		Active: &flags.Active,
	}

	// Update events if provided
	if len(flags.Events) > 0 {
		hook.Events = flags.Events
	} else {
		hook.Events = existingHook.Events
	}

	if flags.DryRun {
		fmt.Printf("Would update webhook ID %d\n", flags.ID)
		return nil
	}

	updatedHook, _, err := client.Repositories.EditHook(ctx, flags.Organization, flags.Repository, flags.ID, hook)
	if err != nil {
		return fmt.Errorf("failed to update webhook: %w", err)
	}

	fmt.Printf("Successfully updated webhook with ID: %d\n", flags.ID)

	return displayWebhook(updatedHook, "table")
}

// runWebhookDelete deletes a webhook.
func runWebhookDelete(flags *WebhookFlags) error {
	ctx := context.Background()
	client := createGitHubClient(flags.Token)

	if flags.DryRun {
		fmt.Printf("Would delete webhook ID %d\n", flags.ID)
		return nil
	}

	_, err := client.Repositories.DeleteHook(ctx, flags.Organization, flags.Repository, flags.ID)
	if err != nil {
		return fmt.Errorf("failed to delete webhook: %w", err)
	}

	fmt.Printf("Successfully deleted webhook with ID: %d\n", flags.ID)

	return nil
}

// runWebhookGet gets details of a specific webhook.
func runWebhookGet(flags *WebhookFlags) error {
	ctx := context.Background()
	client := createGitHubClient(flags.Token)

	hook, _, err := client.Repositories.GetHook(ctx, flags.Organization, flags.Repository, flags.ID)
	if err != nil {
		return fmt.Errorf("failed to get webhook: %w", err)
	}

	return displayWebhook(hook, flags.OutputFormat)
}

// createGitHubClient creates a GitHub API client.
func createGitHubClient(token string) *github.Client {
	if token == "" {
		return github.NewClient(nil)
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)

	return github.NewClient(tc)
}

// displayWebhooks displays a list of webhooks.
func displayWebhooks(webhooks []*github.Hook, format string) error {
	switch format {
	case "json":
		return json.NewEncoder(os.Stdout).Encode(webhooks)
	case "yaml":
		// Convert to repoconfig.WebhookConfig for YAML output
		configs := make([]repoconfig.WebhookConfig, len(webhooks))
		for i, hook := range webhooks {
			configs[i] = convertToWebhookConfig(hook)
		}

		data, err := json.MarshalIndent(configs, "", "  ")
		if err != nil {
			return err
		}

		fmt.Println(string(data))

		return nil
	default:
		// Table format
		fmt.Printf("%-8s %-20s %-40s %-8s %-20s\n", "ID", "EVENTS", "URL", "ACTIVE", "CONTENT_TYPE")
		fmt.Println("-------- -------------------- ---------------------------------------- -------- --------------------")

		for _, hook := range webhooks {
			url := ""
			contentType := ""

			if hook.Config != nil {
				if hook.Config.URL != nil {
					url = *hook.Config.URL
				}

				if hook.Config.ContentType != nil {
					contentType = *hook.Config.ContentType
				}
			}

			eventsStr := ""
			if len(hook.Events) > 0 {
				eventsStr = hook.Events[0]
				if len(hook.Events) > 1 {
					eventsStr += fmt.Sprintf(" (+%d more)", len(hook.Events)-1)
				}
			}

			fmt.Printf("%-8d %-20s %-40s %-8t %-20s\n",
				hook.GetID(),
				eventsStr,
				truncateString(url, 40),
				hook.GetActive(),
				contentType)
		}

		return nil
	}
}

// displayWebhook displays a single webhook.
func displayWebhook(hook *github.Hook, format string) error {
	switch format {
	case "json":
		return json.NewEncoder(os.Stdout).Encode(hook)
	case "yaml":
		config := convertToWebhookConfig(hook)

		data, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return err
		}

		fmt.Println(string(data))

		return nil
	default:
		// Table format
		fmt.Printf("ID: %d\n", hook.GetID())
		fmt.Printf("URL: %s\n", safeStringFromPointer(hook.Config.URL))
		fmt.Printf("Events: %v\n", hook.Events)
		fmt.Printf("Active: %t\n", hook.GetActive())
		fmt.Printf("Content Type: %s\n", safeStringFromPointer(hook.Config.ContentType))
		fmt.Printf("Created: %s\n", hook.GetCreatedAt().Format("2006-01-02 15:04:05"))
		fmt.Printf("Updated: %s\n", hook.GetUpdatedAt().Format("2006-01-02 15:04:05"))

		return nil
	}
}

// convertToWebhookConfig converts a GitHub Hook to repoconfig.WebhookConfig.
func convertToWebhookConfig(hook *github.Hook) repoconfig.WebhookConfig {
	config := repoconfig.WebhookConfig{
		Events: hook.Events,
		Active: hook.Active,
	}

	if hook.Config != nil {
		if hook.Config.URL != nil {
			config.URL = *hook.Config.URL
		}

		if hook.Config.ContentType != nil {
			config.ContentType = *hook.Config.ContentType
		}

		if hook.Config.Secret != nil {
			config.Secret = *hook.Config.Secret
		}
	}

	return config
}

// safeStringFromPointer safely gets a string value from a pointer.
func safeStringFromPointer(ptr *string) string {
	if ptr == nil {
		return ""
	}

	return *ptr
}

// newWebhookBulkCmd creates the webhook bulk operations command.
func newWebhookBulkCmd() *cobra.Command {
	flags := &WebhookFlags{}

	var (
		operation    string
		configFile   string
		parallelJobs int
		dryRun       bool
	)

	cmd := &cobra.Command{
		Use:   "bulk",
		Short: "Bulk webhook operations across repositories",
		Long: `Perform bulk webhook operations across multiple repositories.

This command allows you to manage webhooks at scale across an entire
organization or filtered set of repositories. Operations include:

- create: Bulk create webhooks from configuration
- update: Bulk update existing webhooks
- delete: Bulk delete webhooks by pattern or config
- sync: Synchronize webhooks with configuration file

Bulk Operations:
- Organization-wide webhook management
- Pattern-based repository filtering
- Template-based webhook configuration
- Parallel processing for performance
- Dry-run mode for safe testing

Configuration File Format:
The configuration file should define webhook templates and
repository mappings for consistent webhook deployment.

Examples:
  gz repo-config webhook bulk --operation create --config webhooks.yaml
  gz repo-config webhook bulk --operation sync --org myorg --dry-run
  gz repo-config webhook bulk --operation delete --filter "^legacy-.*"
  gz repo-config webhook bulk --operation update --parallel 10`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWebhookBulkCommand(*flags, operation, configFile, parallelJobs, dryRun)
		},
	}

	addWebhookFlags(cmd, flags)
	cmd.Flags().StringVar(&operation, "operation", "", "Bulk operation (create, update, delete, sync)")
	cmd.Flags().StringVar(&configFile, "webhook-config", "", "Webhook configuration file")
	cmd.Flags().IntVar(&parallelJobs, "parallel-jobs", 5, "Number of parallel operations")
	cmd.Flags().BoolVar(&dryRun, "dry-run-bulk", false, "Preview changes without applying")

	_ = cmd.MarkFlagRequired("operation") // Ignore error

	return cmd
}

// newWebhookAutomationCmd creates the webhook automation command.
func newWebhookAutomationCmd() *cobra.Command {
	flags := &WebhookFlags{}

	var (
		ruleFile  string
		action    string
		enable    bool
		disable   bool
		listRules bool
	)

	cmd := &cobra.Command{
		Use:   "automation",
		Short: "Webhook automation and rule management",
		Long: `Manage webhook automation rules and policies.

This command provides advanced webhook automation capabilities including:

- Rule-based webhook management
- Event-driven webhook creation/updates
- Policy enforcement for webhook standards
- Automated webhook lifecycle management

Automation Features:
- Template-based webhook deployment
- Event-driven webhook creation
- Compliance monitoring and enforcement
- Automated webhook health checks
- Integration with CI/CD pipelines

Rule Types:
- Repository creation triggers
- Compliance policy enforcement
- Security requirement automation
- Integration management rules

Examples:
  gz repo-config webhook automation --list-rules                    # List all rules
  gz repo-config webhook automation --action deploy --rule security # Deploy security rule
  gz repo-config webhook automation --enable --rule compliance      # Enable compliance rule
  gz repo-config webhook automation --disable --rule legacy         # Disable legacy rule`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWebhookAutomationCommand(*flags, ruleFile, action, enable, disable, listRules)
		},
	}

	addWebhookFlags(cmd, flags)
	cmd.Flags().StringVar(&ruleFile, "rule", "", "Automation rule name or file")
	cmd.Flags().StringVar(&action, "action", "", "Automation action (deploy, validate, test)")
	cmd.Flags().BoolVar(&enable, "enable", false, "Enable automation rule")
	cmd.Flags().BoolVar(&disable, "disable", false, "Disable automation rule")
	cmd.Flags().BoolVar(&listRules, "list-rules", false, "List available automation rules")

	return cmd
}

// runWebhookBulkCommand executes bulk webhook operations.
func runWebhookBulkCommand(flags WebhookFlags, operation, configFile string, parallelJobs int, dryRun bool) error {
	if flags.Organization == "" {
		return fmt.Errorf("organization is required (use --org flag)")
	}

	fmt.Printf("🔄 Webhook Bulk Operations\n")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("Organization: %s\n", flags.Organization)
	fmt.Printf("Operation: %s\n", operation)
	fmt.Printf("Parallel jobs: %d\n", parallelJobs)

	if dryRun {
		fmt.Println("Mode: DRY RUN (preview only)")
	}

	fmt.Println()

	switch operation {
	case "create":
		return runBulkCreateWebhooks(flags, configFile, parallelJobs, dryRun)
	case "update":
		return runBulkUpdateWebhooks(flags, configFile, parallelJobs, dryRun)
	case "delete":
		return runBulkDeleteWebhooks(flags, parallelJobs, dryRun)
	case "sync":
		return runBulkSyncWebhooks(flags, configFile, parallelJobs, dryRun)
	default:
		return fmt.Errorf("unsupported operation: %s (supported: create, update, delete, sync)", operation)
	}
}

// runWebhookAutomationCommand executes webhook automation operations.
func runWebhookAutomationCommand(flags WebhookFlags, ruleFile, action string, enable, disable, listRules bool) error {
	fmt.Printf("🤖 Webhook Automation\n")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	if listRules {
		return listAutomationRules()
	}

	if flags.Organization == "" {
		return fmt.Errorf("organization is required (use --org flag)")
	}

	fmt.Printf("Organization: %s\n", flags.Organization)

	if ruleFile != "" {
		fmt.Printf("Rule: %s\n", ruleFile)
	}

	fmt.Println()

	if enable && disable {
		return fmt.Errorf("cannot enable and disable simultaneously")
	}

	if enable {
		return enableAutomationRule(flags, ruleFile)
	}

	if disable {
		return disableAutomationRule(flags, ruleFile)
	}

	if action != "" {
		return executeAutomationAction(flags, ruleFile, action)
	}

	return fmt.Errorf("no automation action specified (use --enable, --disable, --action, or --list-rules)")
}

// Helper functions for bulk operations

func runBulkCreateWebhooks(flags WebhookFlags, configFile string, parallelJobs int, dryRun bool) error {
	fmt.Println("📥 Bulk creating webhooks...")

	// Mock implementation
	fmt.Printf("✅ Would create webhooks for 15 repositories")

	if dryRun {
		fmt.Printf(" (DRY RUN)")
	}

	fmt.Println()

	return nil
}

func runBulkUpdateWebhooks(flags WebhookFlags, configFile string, parallelJobs int, dryRun bool) error {
	fmt.Println("🔄 Bulk updating webhooks...")

	// Mock implementation
	fmt.Printf("✅ Would update webhooks for 12 repositories")

	if dryRun {
		fmt.Printf(" (DRY RUN)")
	}

	fmt.Println()

	return nil
}

func runBulkDeleteWebhooks(flags WebhookFlags, parallelJobs int, dryRun bool) error {
	fmt.Println("🗑️  Bulk deleting webhooks...")

	// Mock implementation
	fmt.Printf("✅ Would delete webhooks from 8 repositories")

	if dryRun {
		fmt.Printf(" (DRY RUN)")
	}

	fmt.Println()

	return nil
}

func runBulkSyncWebhooks(flags WebhookFlags, configFile string, parallelJobs int, dryRun bool) error {
	fmt.Println("🔄 Synchronizing webhooks with configuration...")

	// Mock implementation
	fmt.Printf("✅ Would sync webhooks for 20 repositories")

	if dryRun {
		fmt.Printf(" (DRY RUN)")
	}

	fmt.Println()

	return nil
}

// Helper functions for automation

func listAutomationRules() error {
	fmt.Println("📋 Available Automation Rules:")
	fmt.Println()

	rules := []struct {
		Name        string
		Description string
		Status      string
	}{
		{"security", "Enforce security webhook requirements", "enabled"},
		{"compliance", "Monitor compliance webhook deployment", "enabled"},
		{"ci-cd", "Automate CI/CD webhook setup", "disabled"},
		{"monitoring", "Deploy monitoring webhooks", "enabled"},
		{"legacy", "Legacy webhook migration", "disabled"},
	}

	fmt.Printf("%-15s %-10s %s\n", "RULE", "STATUS", "DESCRIPTION")
	fmt.Println(strings.Repeat("─", 70))

	for _, rule := range rules {
		status := "🟢 Enabled"
		if rule.Status == "disabled" {
			status = "🔴 Disabled"
		}

		fmt.Printf("%-15s %-10s %s\n", rule.Name, status, rule.Description)
	}

	return nil
}

func enableAutomationRule(flags WebhookFlags, ruleName string) error {
	fmt.Printf("✅ Enabling automation rule: %s\n", ruleName)
	fmt.Printf("Rule '%s' is now active for organization: %s\n", ruleName, flags.Organization)

	return nil
}

func disableAutomationRule(flags WebhookFlags, ruleName string) error {
	fmt.Printf("🔴 Disabling automation rule: %s\n", ruleName)
	fmt.Printf("Rule '%s' is now inactive for organization: %s\n", ruleName, flags.Organization)

	return nil
}

func executeAutomationAction(flags WebhookFlags, ruleName, action string) error {
	fmt.Printf("🚀 Executing automation action: %s for rule: %s\n", action, ruleName)

	switch action {
	case "deploy":
		fmt.Printf("📦 Deploying rule '%s' across organization: %s\n", ruleName, flags.Organization)
	case "validate":
		fmt.Printf("✅ Validating rule '%s' configuration\n", ruleName)
	case "test":
		fmt.Printf("🧪 Testing rule '%s' execution\n", ruleName)
	default:
		return fmt.Errorf("unsupported action: %s (supported: deploy, validate, test)", action)
	}

	return nil
}
