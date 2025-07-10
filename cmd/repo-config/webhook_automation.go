package repoconfig

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gizzahub/gzh-manager-go/pkg/webhook/automation"
	"github.com/google/go-github/v66/github"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// WebhookAutomationFlags represents webhook automation command flags
type WebhookAutomationFlags struct {
	GlobalFlags
	ConfigFile  string
	ConfigDir   string
	Port        int
	WebhookPath string
	Secret      string
	Workers     int
	LogLevel    string
}

// newWebhookAutomationCmd creates the webhook automation command
func newWebhookAutomationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "automation",
		Short: "Manage webhook event automation rules",
		Long: `Configure and run event-based automation rules for GitHub webhooks.

This command allows you to define rules that automatically respond to GitHub
events, such as creating issues, adding labels, merging PRs, or running workflows
based on specific conditions.

Examples:
  gz repo-config webhook automation server --config rules.yaml --port 8080
  gz repo-config webhook automation validate --config rules.yaml
  gz repo-config webhook automation test --config rules.yaml --event push.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	// Add subcommands
	cmd.AddCommand(newWebhookAutomationServerCmd())
	cmd.AddCommand(newWebhookAutomationValidateCmd())
	cmd.AddCommand(newWebhookAutomationTestCmd())
	cmd.AddCommand(newWebhookAutomationExampleCmd())

	return cmd
}

// newWebhookAutomationServerCmd creates the server command
func newWebhookAutomationServerCmd() *cobra.Command {
	flags := &WebhookAutomationFlags{}
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Run the webhook automation server",
		Long: `Start the webhook automation server to receive and process GitHub events.

The server listens for incoming webhooks and processes them according to the
configured automation rules.

Examples:
  gz repo-config webhook automation server --config rules.yaml --port 8080
  gz repo-config webhook automation server --config-dir /etc/gzh/rules --secret $WEBHOOK_SECRET`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWebhookAutomationServer(flags)
		},
	}

	addGlobalFlags(cmd, &flags.GlobalFlags)
	cmd.Flags().StringVar(&flags.ConfigDir, "config-dir", "", "Directory containing automation rule files")
	cmd.Flags().IntVar(&flags.Port, "port", 8080, "Server port")
	cmd.Flags().StringVar(&flags.WebhookPath, "webhook-path", "/webhook", "Webhook endpoint path")
	cmd.Flags().StringVar(&flags.Secret, "secret", "", "Webhook secret for signature verification")
	cmd.Flags().IntVar(&flags.Workers, "workers", 10, "Number of event processing workers")
	cmd.Flags().StringVar(&flags.LogLevel, "log-level", "info", "Log level (debug, info, warn, error)")

	return cmd
}

// newWebhookAutomationValidateCmd creates the validate command
func newWebhookAutomationValidateCmd() *cobra.Command {
	flags := &WebhookAutomationFlags{}
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate automation rule configuration",
		Long: `Validate automation rule configuration files for syntax and logic errors.

Examples:
  gz repo-config webhook automation validate --config rules.yaml
  gz repo-config webhook automation validate --config-dir /etc/gzh/rules`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWebhookAutomationValidate(flags)
		},
	}

	cmd.Flags().StringVar(&flags.ConfigDir, "config-dir", "", "Directory containing automation rule files")

	return cmd
}

// newWebhookAutomationTestCmd creates the test command
func newWebhookAutomationTestCmd() *cobra.Command {
	flags := &WebhookAutomationFlags{}
	var eventFile string

	cmd := &cobra.Command{
		Use:   "test",
		Short: "Test automation rules with sample events",
		Long: `Test automation rules by simulating webhook events.

This command allows you to test your automation rules without setting up
actual webhooks. It loads the rules and processes a sample event file.

Examples:
  gz repo-config webhook automation test --config rules.yaml --event samples/push-event.json
  gz repo-config webhook automation test --config rules.yaml --event samples/pr-opened.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWebhookAutomationTest(flags, eventFile)
		},
	}

	addGlobalFlags(cmd, &flags.GlobalFlags)
	cmd.Flags().StringVar(&eventFile, "event", "", "Event file to test (required)")
	cmd.Flags().StringVar(&flags.LogLevel, "log-level", "debug", "Log level (debug, info, warn, error)")

	cmd.MarkFlagRequired("config")
	cmd.MarkFlagRequired("event")

	return cmd
}

// newWebhookAutomationExampleCmd creates the example command
func newWebhookAutomationExampleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "example",
		Short: "Generate example automation rule configuration",
		Long: `Generate example automation rule configuration files.

This command creates sample configuration files that demonstrate various
automation patterns and best practices.

Examples:
  gz repo-config webhook automation example > rules.yaml
  gz repo-config webhook automation example --type advanced > advanced-rules.yaml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWebhookAutomationExample()
		},
	}

	return cmd
}

// runWebhookAutomationServer runs the automation server
func runWebhookAutomationServer(flags *WebhookAutomationFlags) error {
	// Setup logger
	logger, err := setupLogger(flags.LogLevel)
	if err != nil {
		return fmt.Errorf("failed to setup logger: %w", err)
	}

	// Load configuration
	configs, err := loadAutomationConfigs(flags)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	if len(configs) == 0 {
		return fmt.Errorf("no automation rules found")
	}

	// Merge all configs
	config := automation.MergeConfigs(configs...)

	// Validate configuration
	if err := automation.ValidateConfig(config); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Create GitHub client
	client := createGitHubClient(flags.Token)

	// Create automation engine
	engine := automation.NewEngine(client, logger)

	// Register handlers
	if err := registerHandlers(engine, client, logger); err != nil {
		return fmt.Errorf("failed to register handlers: %w", err)
	}

	// Add rules to engine
	for _, rule := range config.Rules {
		if err := engine.AddRule(rule); err != nil {
			return fmt.Errorf("failed to add rule %s: %w", rule.ID, err)
		}
	}

	// Create server
	serverConfig := automation.ServerConfig{
		Port:        flags.Port,
		WebhookPath: flags.WebhookPath,
		Secret:      flags.Secret,
		Workers:     flags.Workers,
	}

	server := automation.NewServer(engine, serverConfig, logger)

	// Setup signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Received shutdown signal")
		cancel()
	}()

	// Start server
	logger.Info("Starting webhook automation server",
		zap.Int("port", flags.Port),
		zap.String("path", flags.WebhookPath),
		zap.Int("rules", len(config.Rules)))

	return server.Start(ctx)
}

// runWebhookAutomationValidate validates the configuration
func runWebhookAutomationValidate(flags *WebhookAutomationFlags) error {
	configs, err := loadAutomationConfigs(flags)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	if len(configs) == 0 {
		return fmt.Errorf("no configuration files found")
	}

	allValid := true
	for i, config := range configs {
		fmt.Printf("Validating configuration %d...\n", i+1)
		if err := automation.ValidateConfig(config); err != nil {
			fmt.Printf("  ❌ Invalid: %v\n", err)
			allValid = false
		} else {
			fmt.Printf("  ✅ Valid (%d rules)\n", len(config.Rules))
		}
	}

	if !allValid {
		return fmt.Errorf("validation failed")
	}

	fmt.Println("\n✅ All configurations are valid")
	return nil
}

// runWebhookAutomationTest tests the automation rules
func runWebhookAutomationTest(flags *WebhookAutomationFlags, eventFile string) error {
	// Setup logger
	logger, err := setupLogger(flags.LogLevel)
	if err != nil {
		return fmt.Errorf("failed to setup logger: %w", err)
	}

	// Load configuration
	config, err := automation.LoadConfig(flags.GlobalFlags.ConfigFile)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Validate configuration
	if err := automation.ValidateConfig(config); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Load test event
	_, err = os.ReadFile(eventFile)
	if err != nil {
		return fmt.Errorf("failed to read event file: %w", err)
	}

	// Create GitHub client
	client := createGitHubClient(flags.Token)

	// Create automation engine
	engine := automation.NewEngine(client, logger)

	// Register handlers (in test mode, some actions might be mocked)
	if err := registerHandlers(engine, client, logger); err != nil {
		return fmt.Errorf("failed to register handlers: %w", err)
	}

	// Add rules to engine
	for _, rule := range config.Rules {
		if err := engine.AddRule(rule); err != nil {
			return fmt.Errorf("failed to add rule %s: %w", rule.ID, err)
		}
	}

	// Parse and process the test event
	fmt.Printf("Testing with event from %s\n", eventFile)
	fmt.Printf("Loaded %d rules\n\n", len(config.Rules))

	// TODO: Implement event parsing and processing for testing

	return nil
}

// runWebhookAutomationExample generates example configuration
func runWebhookAutomationExample() error {
	example := `# Webhook Automation Rules Configuration
# This file defines automated actions triggered by GitHub webhook events

version: "1.0"

# Global settings
global:
  enabled: true
  default_timeout: "30s"
  max_concurrency: 10
  notification_urls:
    slack: "${SLACK_WEBHOOK_URL}"
    discord: "${DISCORD_WEBHOOK_URL}"

# Automation rules
rules:
  # Auto-label pull requests based on files changed
  - id: "auto-label-pr"
    name: "Auto-label Pull Requests"
    description: "Automatically add labels to PRs based on files changed"
    enabled: true
    priority: 100
    conditions:
      - type: "event_type"
        operator: "equals"
        value: "pull_request.opened"
    actions:
      - type: "add_label"
        parameters:
          labels:
            - "needs-review"

  # Welcome new contributors
  - id: "welcome-contributor"
    name: "Welcome First-Time Contributors"
    description: "Post a welcome message for first-time contributors"
    enabled: true
    priority: 90
    conditions:
      - type: "event_type"
        operator: "in"
        value: ["pull_request.opened", "issues.opened"]
      - type: "sender"
        field: "type"
        operator: "equals"
        value: "User"
    actions:
      - type: "create_comment"
        parameters:
          body: |
            Welcome @{{sender.login}}! 👋
            
            Thank you for your contribution to {{repo.name}}!
            A maintainer will review your submission soon.

  # Auto-merge dependabot PRs
  - id: "auto-merge-dependabot"
    name: "Auto-merge Dependabot PRs"
    description: "Automatically merge dependency updates that pass CI"
    enabled: true
    priority: 80
    conditions:
      - type: "event_type"
        operator: "equals"
        value: "pull_request.opened"
      - type: "sender"
        field: "login"
        operator: "equals"
        value: "dependabot[bot]"
    actions:
      - type: "add_label"
        parameters:
          labels:
            - "dependencies"
            - "automerge"

  # Notify on release
  - id: "notify-release"
    name: "Notify on Release"
    description: "Send notifications when a new release is published"
    enabled: true
    priority: 70
    conditions:
      - type: "event_type"
        operator: "equals"
        value: "release.published"
    actions:
      - type: "notification"
        parameters:
          type: "slack"
          message: "🚀 New release {{repo.name}} v{{payload.release.tag_name}} is out!"
        async: true

  # Create issue for failed workflows
  - id: "workflow-failure-issue"
    name: "Create Issue for Failed Workflows"
    description: "Create an issue when critical workflows fail"
    enabled: true
    priority: 60
    conditions:
      - type: "event_type"
        operator: "equals"
        value: "workflow_run.completed"
      - type: "payload"
        field: "workflow_run.conclusion"
        operator: "equals"
        value: "failure"
      - type: "payload"
        field: "workflow_run.name"
        operator: "in"
        value: ["CI/CD", "Deploy", "Release"]
    actions:
      - type: "create_issue"
        parameters:
          title: "Workflow Failure: {{payload.workflow_run.name}}"
          body: |
            The workflow **{{payload.workflow_run.name}}** failed.
            
            **Branch:** {{payload.workflow_run.head_branch}}
            **Commit:** {{payload.workflow_run.head_sha}}
            **Run URL:** {{payload.workflow_run.html_url}}
            
            Please investigate and fix the issue.
          labels:
            - "bug"
            - "ci/cd"
          assignees:
            - "{{payload.workflow_run.triggering_actor.login}}"

  # Security alert handling
  - id: "security-alert"
    name: "Handle Security Alerts"
    description: "Create high-priority issues for security vulnerabilities"
    enabled: true
    priority: 100
    conditions:
      - type: "event_type"
        operator: "matches"
        value: "security_advisory.*"
    actions:
      - type: "create_issue"
        parameters:
          title: "🔒 Security Alert: {{payload.security_advisory.summary}}"
          body: |
            A security vulnerability has been detected.
            
            **Severity:** {{payload.security_advisory.severity}}
            **Package:** {{payload.security_advisory.affected_package_name}}
            **Description:** {{payload.security_advisory.description}}
            
            Immediate action required!
          labels:
            - "security"
            - "high-priority"
      - type: "notification"
        parameters:
          type: "slack"
          message: "⚠️ Security alert in {{repo.name}}: {{payload.security_advisory.summary}}"
        async: true
`

	fmt.Print(example)
	return nil
}

// loadAutomationConfigs loads automation configurations
func loadAutomationConfigs(flags *WebhookAutomationFlags) ([]*automation.Config, error) {
	var configs []*automation.Config

	if flags.GlobalFlags.ConfigFile != "" {
		config, err := automation.LoadConfig(flags.GlobalFlags.ConfigFile)
		if err != nil {
			return nil, err
		}
		configs = append(configs, config)
	}

	if flags.ConfigDir != "" {
		dirConfigs, err := automation.LoadConfigFromDirectory(flags.ConfigDir)
		if err != nil {
			return nil, err
		}
		configs = append(configs, dirConfigs...)
	}

	return configs, nil
}

// registerHandlers registers all action handlers
func registerHandlers(engine *automation.Engine, client *github.Client, logger *zap.Logger) error {
	// Register issue creation handler
	if err := engine.RegisterHandler("create_issue", automation.NewCreateIssueHandler(client, logger)); err != nil {
		return err
	}

	// Register label handler
	if err := engine.RegisterHandler("add_label", automation.NewAddLabelHandler(client, logger)); err != nil {
		return err
	}

	// Register comment handler
	if err := engine.RegisterHandler("create_comment", automation.NewCreateCommentHandler(client, logger)); err != nil {
		return err
	}

	// Register PR merge handler
	if err := engine.RegisterHandler("merge_pr", automation.NewMergePRHandler(client, logger)); err != nil {
		return err
	}

	// Register notification handler
	notificationHandler := automation.NewNotificationHandler(logger)
	// Configure notification webhooks from environment or config
	if slackURL := os.Getenv("SLACK_WEBHOOK_URL"); slackURL != "" {
		notificationHandler.RegisterWebhook("slack", slackURL)
	}
	if discordURL := os.Getenv("DISCORD_WEBHOOK_URL"); discordURL != "" {
		notificationHandler.RegisterWebhook("discord", discordURL)
	}
	if err := engine.RegisterHandler("notification", notificationHandler); err != nil {
		return err
	}

	// Register workflow handler
	if err := engine.RegisterHandler("run_workflow", automation.NewRunWorkflowHandler(client, logger)); err != nil {
		return err
	}

	return nil
}

// setupLogger creates a logger with the specified level
func setupLogger(level string) (*zap.Logger, error) {
	config := zap.NewProductionConfig()

	switch level {
	case "debug":
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		config.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		config.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		return nil, fmt.Errorf("invalid log level: %s", level)
	}

	return config.Build()
}
