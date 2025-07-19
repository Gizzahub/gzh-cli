package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/gizzahub/gzh-manager-go/internal/event"
	"github.com/gizzahub/gzh-manager-go/pkg/github"
	"github.com/spf13/cobra"
)

// WebhookDependencies holds all dependencies for webhook commands.
type WebhookDependencies struct {
	WebhookService github.WebhookService
	Logger         github.Logger
	ClientFactory  func(token string) (github.APIClient, error)
}

// WebhookCommandFactory creates webhook commands with injected dependencies.
type WebhookCommandFactory struct {
	deps *WebhookDependencies
}

// NewWebhookCommandFactory creates a new webhook command factory.
func NewWebhookCommandFactory(deps *WebhookDependencies) *WebhookCommandFactory {
	// Provide defaults if not specified
	if deps == nil {
		deps = &WebhookDependencies{}
	}

	if deps.Logger == nil {
		// Use a default logger if none provided
		deps.Logger = event.NewLoggerAdapter()
	}

	if deps.ClientFactory == nil {
		// Default client factory
		deps.ClientFactory = func(token string) (github.APIClient, error) {
			config := &github.APIClientConfig{
				Token: token,
			}
			// Create a simple HTTP client that implements HTTPClientInterface
			httpClient := &simpleHTTPClient{client: &http.Client{}}
			logger := event.NewLoggerAdapter()

			return github.NewAPIClient(config, httpClient, logger), nil
		}
	}

	return &WebhookCommandFactory{
		deps: deps,
	}
}

// NewWebhookCmd creates the webhook command with dependency injection.
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
	if err := createCmd.MarkFlagRequired("url"); err != nil {
		// Error marking flag as required - continue without marking
	}
	createCmd.Flags().StringSlice("events", []string{"push"}, "Events to trigger webhook")
	createCmd.Flags().Bool("active", true, "Whether webhook is active")
	createCmd.Flags().String("content-type", "json", "Content type (json/form)")
	createCmd.Flags().String("secret", "", "Webhook secret")

	repoCmd.AddCommand(createCmd)
	cmd.AddCommand(repoCmd)

	return cmd
}

// runCreateRepositoryWebhook handles webhook creation with injected dependencies.
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

// simpleHTTPClient implements HTTPClientInterface.
type simpleHTTPClient struct {
	client *http.Client
}

func (c *simpleHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}

func (c *simpleHTTPClient) Get(url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		return nil, err
	}
	return c.client.Do(req)
}

func (c *simpleHTTPClient) Post(url, contentType string, body interface{}) (*http.Response, error) {
	var bodyReader *bytes.Reader

	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}

		bodyReader = bytes.NewReader(bodyBytes)

		req, err := http.NewRequestWithContext(context.Background(), "POST", url, bodyReader)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", contentType)
		return c.client.Do(req)
	}

	req, err := http.NewRequestWithContext(context.Background(), "POST", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return c.client.Do(req)
}
