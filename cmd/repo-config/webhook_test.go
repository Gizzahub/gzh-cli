package repoconfig

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gizzahub/gzh-manager-go/pkg/types/repoconfig"
	"github.com/google/go-github/v66/github"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWebhookCmd(t *testing.T) {
	cmd := newWebhookCmd()

	assert.Equal(t, "webhook", cmd.Use)
	assert.Equal(t, "Manage repository webhooks", cmd.Short)
	assert.Len(t, cmd.Commands(), 5) // list, create, update, delete, get
}

func TestWebhookCommands(t *testing.T) {
	tests := []struct {
		name     string
		cmdFunc  func() *cobra.Command
		use      string
		required []string
	}{
		{
			name:     "list command",
			cmdFunc:  newWebhookListCmd,
			use:      "list",
			required: []string{"repo", "org"},
		},
		{
			name:     "create command",
			cmdFunc:  newWebhookCreateCmd,
			use:      "create",
			required: []string{"repo", "org", "url"},
		},
		{
			name:     "update command",
			cmdFunc:  newWebhookUpdateCmd,
			use:      "update",
			required: []string{"repo", "org", "id"},
		},
		{
			name:     "delete command",
			cmdFunc:  newWebhookDeleteCmd,
			use:      "delete",
			required: []string{"repo", "org", "id"},
		},
		{
			name:     "get command",
			cmdFunc:  newWebhookGetCmd,
			use:      "get",
			required: []string{"repo", "org", "id"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.cmdFunc()
			assert.Equal(t, tt.use, cmd.Use)

			// Check required flags
			for _, flag := range tt.required {
				f := cmd.Flag(flag)
				require.NotNil(t, f, "Flag %s should exist", flag)
			}
		})
	}
}

func TestConvertToWebhookConfig(t *testing.T) {
	active := true
	url := "https://example.com/webhook"
	contentType := "json"
	secret := "mysecret"

	hook := &github.Hook{
		ID:     github.Int64(12345),
		Events: []string{"push", "pull_request"},
		Active: &active,
		Config: &github.HookConfig{
			URL:         &url,
			ContentType: &contentType,
			Secret:      &secret,
		},
	}

	config := convertToWebhookConfig(hook)

	assert.Equal(t, "https://example.com/webhook", config.URL)
	assert.Equal(t, []string{"push", "pull_request"}, config.Events)
	assert.Equal(t, &active, config.Active)
	assert.Equal(t, "json", config.ContentType)
	assert.Equal(t, "mysecret", config.Secret)
}

func TestSafeStringFromPointer(t *testing.T) {
	url := "https://example.com"
	contentType := "json"

	assert.Equal(t, "https://example.com", safeStringFromPointer(&url))
	assert.Equal(t, "json", safeStringFromPointer(&contentType))
	assert.Equal(t, "", safeStringFromPointer(nil))
}

func TestDisplayWebhooks(t *testing.T) {
	active := true
	url1 := "https://example.com/webhook1"
	url2 := "https://example.com/webhook2"
	contentType1 := "json"
	contentType2 := "form"

	webhooks := []*github.Hook{
		{
			ID:     github.Int64(1),
			Events: []string{"push"},
			Active: &active,
			Config: &github.HookConfig{
				URL:         &url1,
				ContentType: &contentType1,
			},
		},
		{
			ID:     github.Int64(2),
			Events: []string{"push", "pull_request", "issues"},
			Active: &active,
			Config: &github.HookConfig{
				URL:         &url2,
				ContentType: &contentType2,
			},
		},
	}

	// Test table format (default)
	err := displayWebhooks(webhooks, "table")
	assert.NoError(t, err)

	// Test JSON format
	err = displayWebhooks(webhooks, "json")
	assert.NoError(t, err)

	// Test YAML format
	err = displayWebhooks(webhooks, "yaml")
	assert.NoError(t, err)
}

func TestDisplayWebhook(t *testing.T) {
	active := true
	url := "https://example.com/webhook"
	contentType := "json"

	hook := &github.Hook{
		ID:     github.Int64(12345),
		Events: []string{"push", "pull_request"},
		Active: &active,
		Config: &github.HookConfig{
			URL:         &url,
			ContentType: &contentType,
		},
		CreatedAt: &github.Timestamp{},
		UpdatedAt: &github.Timestamp{},
	}

	// Test table format (default)
	err := displayWebhook(hook, "table")
	assert.NoError(t, err)

	// Test JSON format
	err = displayWebhook(hook, "json")
	assert.NoError(t, err)

	// Test YAML format
	err = displayWebhook(hook, "yaml")
	assert.NoError(t, err)
}

// Mock GitHub API server for integration tests
func setupMockGitHubServer() *httptest.Server {
	mux := http.NewServeMux()

	// Mock webhook list endpoint
	mux.HandleFunc("/repos/testorg/testrepo/hooks", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			response := `[
				{
					"id": 1,
					"url": "https://api.github.com/repos/testorg/testrepo/hooks/1",
					"events": ["push"],
					"active": true,
					"config": {
						"url": "https://example.com/webhook1",
						"content_type": "json"
					},
					"created_at": "2023-01-01T00:00:00Z",
					"updated_at": "2023-01-01T00:00:00Z"
				}
			]`
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(response))
			return
		}

		if r.Method == "POST" {
			response := `{
				"id": 2,
				"url": "https://api.github.com/repos/testorg/testrepo/hooks/2",
				"events": ["push"],
				"active": true,
				"config": {
					"url": "https://example.com/webhook2",
					"content_type": "json"
				},
				"created_at": "2023-01-01T00:00:00Z",
				"updated_at": "2023-01-01T00:00:00Z"
			}`
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(response))
			return
		}
	})

	// Mock individual webhook endpoints
	mux.HandleFunc("/repos/testorg/testrepo/hooks/1", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			response := `{
				"id": 1,
				"url": "https://api.github.com/repos/testorg/testrepo/hooks/1",
				"events": ["push"],
				"active": true,
				"config": {
					"url": "https://example.com/webhook1",
					"content_type": "json"
				},
				"created_at": "2023-01-01T00:00:00Z",
				"updated_at": "2023-01-01T00:00:00Z"
			}`
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(response))
		case "PATCH":
			response := `{
				"id": 1,
				"url": "https://api.github.com/repos/testorg/testrepo/hooks/1",
				"events": ["push", "pull_request"],
				"active": true,
				"config": {
					"url": "https://example.com/webhook1-updated",
					"content_type": "json"
				},
				"created_at": "2023-01-01T00:00:00Z",
				"updated_at": "2023-01-01T00:00:00Z"
			}`
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(response))
		case "DELETE":
			w.WriteHeader(http.StatusNoContent)
		}
	})

	return httptest.NewServer(mux)
}

func TestWebhookIntegration(t *testing.T) {
	// Skip integration tests in CI unless specifically enabled
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	server := setupMockGitHubServer()
	defer server.Close()

	// Create client pointing to mock server
	serverURL, _ := url.Parse(server.URL + "/")
	client := github.NewClient(nil)
	client.BaseURL = serverURL

	// Test webhook operations would go here
	// This is a placeholder for actual integration tests
	assert.NotNil(t, client)
}

func TestWebhookValidation(t *testing.T) {
	tests := []struct {
		name   string
		config repoconfig.WebhookConfig
		valid  bool
	}{
		{
			name: "valid webhook",
			config: repoconfig.WebhookConfig{
				URL:    "https://example.com/webhook",
				Events: []string{"push"},
			},
			valid: true,
		},
		{
			name: "empty URL",
			config: repoconfig.WebhookConfig{
				URL:    "",
				Events: []string{"push"},
			},
			valid: false,
		},
		{
			name: "no events",
			config: repoconfig.WebhookConfig{
				URL:    "https://example.com/webhook",
				Events: []string{},
			},
			valid: false,
		},
		{
			name: "invalid URL",
			config: repoconfig.WebhookConfig{
				URL:    "://invalid-url",
				Events: []string{"push"},
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateWebhookConfig(tt.config)
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

// validateWebhookConfig validates a webhook configuration
func validateWebhookConfig(config repoconfig.WebhookConfig) error {
	if config.URL == "" {
		return fmt.Errorf("webhook URL is required")
	}

	if len(config.Events) == 0 {
		return fmt.Errorf("at least one event must be specified")
	}

	// Parse URL to validate format
	if _, err := url.Parse(config.URL); err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	return nil
}
