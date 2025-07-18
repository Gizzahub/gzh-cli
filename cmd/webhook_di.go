package cmd

import (
	"github.com/gizzahub/gzh-manager-go/pkg/github"
	"github.com/spf13/cobra"
)

// WebhookDependencies holds all dependencies for webhook commands
type WebhookDependencies struct {
	WebhookService github.WebhookService
	Logger         github.EventLogger
	ClientFactory  func(token string) (*github.Client, error)
}

// WebhookCommandFactory creates webhook commands with injected dependencies
type WebhookCommandFactory struct {
	deps *WebhookDependencies
}

// NewWebhookCommandFactory creates a new webhook command factory
func NewWebhookCommandFactory(deps *WebhookDependencies) *WebhookCommandFactory {
	// Provide defaults if not specified
	if deps == nil {
		deps = &WebhookDependencies{}
	}
	
	if deps.Logger == nil {
		// Use a default logger if none provided
		deps.Logger = &defaultLogger{}
	}
	
	if deps.ClientFactory == nil {
		// Default client factory
		deps.ClientFactory = func(token string) (*github.Client, error) {
			return github.NewClient(nil), nil
		}
	}
	
	return &WebhookCommandFactory{
		deps: deps,
	}
}

// NewWebhookCmd creates the webhook command with dependency injection
func (f *WebhookCommandFactory) NewWebhookCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "webhook",
		Short: "Manage GitHub webhooks",
		Long: `Manage GitHub webhooks for repositories and organizations.

This command provides comprehensive webhook management including:
- Creating and configuring webhooks
- Managing webhook policies
- Monitoring webhook deliveries
- Troubleshooting webhook issues`,
	}

	// Repository webhook commands
	repoCmd := &cobra.Command{
		Use:   "repo",
		Short: "Manage repository webhooks",
		Long:  `Create, update, delete and list webhooks for repositories.`,
	}

	createCmd := &cobra.Command{
		Use:   "create [owner] [repo]",
		Short: "Create a webhook for a repository",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return f.runCreateRepositoryWebhook(cmd, args)
		},
	}

	// Add flags
	createCmd.Flags().String("name", "web", "Webhook name")
	createCmd.Flags().String("url", "", "Webhook URL (required)")
	createCmd.MarkFlagRequired("url")
	createCmd.Flags().StringSlice("events", []string{"push"}, "Events to trigger webhook")
	createCmd.Flags().Bool("active", true, "Whether webhook is active")
	createCmd.Flags().String("content-type", "json", "Content type (json/form)")
	createCmd.Flags().String("secret", "", "Webhook secret")

	repoCmd.AddCommand(createCmd)
	cmd.AddCommand(repoCmd)

	return cmd
}

// runCreateRepositoryWebhook handles webhook creation with injected dependencies
func (f *WebhookCommandFactory) runCreateRepositoryWebhook(cmd *cobra.Command, args []string) error {
	owner, repo := args[0], args[1]

	name, _ := cmd.Flags().GetString("name")
	url, _ := cmd.Flags().GetString("url")
	events, _ := cmd.Flags().GetStringSlice("events")
	active, _ := cmd.Flags().GetBool("active")
	contentType, _ := cmd.Flags().GetString("content-type")
	secret, _ := cmd.Flags().GetString("secret")

	// Use injected webhook service
	webhookService := f.deps.WebhookService
	if webhookService == nil {
		// Create default service if not injected
		webhookService = github.NewWebhookService(nil, f.deps.Logger)
	}

	request := &github.WebhookCreateRequest{
		Name:   name,
		URL:    url,
		Events: events,
		Active: active,
		Config: github.WebhookConfig{
			URL:         url,
			ContentType: contentType,
			Secret:      secret,
		},
	}

	webhook, err := webhookService.CreateRepositoryWebhook(cmd.Context(), owner, repo, request)
	if err != nil {
		return err
	}

	f.deps.Logger.Info("Webhook created successfully", 
		"webhook_id", webhook.ID,
		"url", webhook.URL,
		"events", webhook.Events,
	)

	return nil
}

// defaultLogger is a simple logger implementation
type defaultLogger struct{}

func (l *defaultLogger) Info(msg string, keysAndValues ...interface{}) {}
func (l *defaultLogger) Error(msg string, err error, keysAndValues ...interface{}) {}
func (l *defaultLogger) Debug(msg string, keysAndValues ...interface{}) {}
func (l *defaultLogger) Warn(msg string, keysAndValues ...interface{}) {}