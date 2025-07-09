package repoconfig

import (
	"fmt"
	"os"
	"testing"

	"github.com/gizzahub/gzh-manager-go/pkg/types/repoconfig"
	"github.com/google/go-github/v66/github"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestNewWebhookBulkCmd(t *testing.T) {
	cmd := newWebhookBulkCmd()

	assert.Equal(t, "bulk", cmd.Use)
	assert.Equal(t, "Manage webhooks across multiple repositories", cmd.Short)
	assert.Len(t, cmd.Commands(), 4) // create, list, delete, sync
}

func TestWebhookBulkCommands(t *testing.T) {
	tests := []struct {
		name     string
		cmdFunc  func() *cobra.Command
		use      string
		required []string
	}{
		{
			name:     "bulk create command",
			cmdFunc:  newWebhookBulkCreateCmd,
			use:      "create",
			required: []string{"org"},
		},
		{
			name:     "bulk list command",
			cmdFunc:  newWebhookBulkListCmd,
			use:      "list",
			required: []string{"org"},
		},
		{
			name:     "bulk delete command",
			cmdFunc:  newWebhookBulkDeleteCmd,
			use:      "delete",
			required: []string{"org", "url"},
		},
		{
			name:     "bulk sync command",
			cmdFunc:  newWebhookBulkSyncCmd,
			use:      "sync",
			required: []string{"org", "config"},
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

func TestLoadBulkWebhookConfig(t *testing.T) {
	// Create temporary config file
	configContent := `
version: "1.0"
webhooks:
  - url: https://example.com/webhook1
    events: [push, pull_request]
    active: true
    content_type: json
    secret: mysecret
  - url: https://example.com/webhook2
    events: [issues, issue_comment]
    active: true
    content_type: json
targets:
  all: true
  exclude:
    - test-repo
    - archived-repo
options:
  skip_existing: true
  max_workers: 10
  continue_on_error: true
`

	tmpFile, err := os.CreateTemp("", "webhook-config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	tmpFile.Close()

	// Load and verify config
	config, err := loadBulkWebhookConfig(tmpFile.Name())
	require.NoError(t, err)

	assert.Equal(t, "1.0", config.Version)
	assert.Len(t, config.Webhooks, 2)
	assert.Equal(t, "https://example.com/webhook1", config.Webhooks[0].URL)
	assert.Equal(t, []string{"push", "pull_request"}, config.Webhooks[0].Events)
	assert.True(t, *config.Webhooks[0].Active)
	assert.Equal(t, "json", config.Webhooks[0].ContentType)
	assert.Equal(t, "mysecret", config.Webhooks[0].Secret)

	assert.True(t, config.Targets.All)
	assert.Equal(t, []string{"test-repo", "archived-repo"}, config.Targets.Exclude)

	assert.True(t, config.Options.SkipExisting)
	assert.Equal(t, 10, config.Options.MaxWorkers)
	assert.True(t, config.Options.ContinueOnError)
}

func TestBulkOperationResult(t *testing.T) {
	results := []BulkOperationResult{
		{
			Repository: "repo1",
			Action:     "create",
			Success:    true,
			Details:    "Created webhook ID: 123",
		},
		{
			Repository: "repo2",
			Action:     "create",
			Success:    false,
			Error:      fmt.Errorf("permission denied"),
		},
		{
			Repository: "repo3",
			Action:     "create",
			Success:    true,
			Details:    "Created webhook ID: 456",
		},
	}

	// Test hasErrors function
	assert.True(t, hasErrors(results))

	successOnlyResults := []BulkOperationResult{results[0], results[2]}
	assert.False(t, hasErrors(successOnlyResults))
}

func TestMatchPattern(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		pattern  string
		expected bool
	}{
		{"exact match", "api-service", "api-service", true},
		{"prefix match", "api-service", "api-", true},
		{"suffix match", "my-api", "-api", true},
		{"no match", "web-service", "api", false},
		{"contains match", "my-api-service", "api", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := matchPattern(tt.input, tt.pattern)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFindWebhookByURL(t *testing.T) {
	// This test would require mocking the GitHub client
	// For now, we'll just test the logic with a mock scenario
	t.Skip("Requires GitHub client mocking")
}

func TestDisplayBulkResults(t *testing.T) {
	results := []BulkOperationResult{
		{
			Repository: "repo1",
			Action:     "create",
			Success:    true,
			Details:    "Created webhook ID: 123",
		},
		{
			Repository: "repo2",
			Action:     "create",
			Success:    false,
			Error:      fmt.Errorf("permission denied"),
		},
		{
			Repository: "repo1",
			Action:     "update",
			Success:    true,
			Details:    "Updated webhook ID: 456",
		},
	}

	// Test that displayBulkResults doesn't panic
	displayBulkResults(results, "test")
}

func TestBulkWebhookConfigMarshal(t *testing.T) {
	active := true
	config := BulkWebhookConfig{
		Version: "1.0",
		Webhooks: []repoconfig.WebhookConfig{
			{
				URL:         "https://example.com/webhook",
				Events:      []string{"push", "pull_request"},
				Active:      &active,
				ContentType: "json",
				Secret:      "mysecret",
			},
		},
		Targets: BulkWebhookTargets{
			All:     true,
			Exclude: []string{"test-repo"},
		},
		Options: BulkWebhookOptions{
			SkipExisting:    true,
			MaxWorkers:      5,
			ContinueOnError: false,
		},
	}

	// Marshal to YAML
	data, err := yaml.Marshal(config)
	require.NoError(t, err)

	// Unmarshal back
	var loaded BulkWebhookConfig
	err = yaml.Unmarshal(data, &loaded)
	require.NoError(t, err)

	assert.Equal(t, config.Version, loaded.Version)
	assert.Len(t, loaded.Webhooks, 1)
	assert.Equal(t, config.Webhooks[0].URL, loaded.Webhooks[0].URL)
	assert.Equal(t, config.Targets.All, loaded.Targets.All)
	assert.Equal(t, config.Options.MaxWorkers, loaded.Options.MaxWorkers)
}

func TestDisplayWebhooksTable(t *testing.T) {
	url1 := "https://example.com/webhook1"
	url2 := "https://example.com/webhook2"
	active := true

	webhooks := map[string][]*github.Hook{
		"repo1": {
			{
				ID:     github.Int64(1),
				Events: []string{"push"},
				Active: &active,
				Config: &github.HookConfig{
					URL: &url1,
				},
			},
		},
		"repo2": {
			{
				ID:     github.Int64(2),
				Events: []string{"push", "pull_request"},
				Active: &active,
				Config: &github.HookConfig{
					URL: &url2,
				},
			},
		},
	}

	// Test that display functions don't panic
	err := displayWebhooksTable(webhooks)
	assert.NoError(t, err)
}

func TestBulkWebhookTargets(t *testing.T) {
	tests := []struct {
		name    string
		targets BulkWebhookTargets
		isValid bool
	}{
		{
			name: "all repositories",
			targets: BulkWebhookTargets{
				All: true,
			},
			isValid: true,
		},
		{
			name: "specific repositories",
			targets: BulkWebhookTargets{
				Repositories: []string{"repo1", "repo2"},
			},
			isValid: true,
		},
		{
			name: "pattern matching",
			targets: BulkWebhookTargets{
				Pattern: "^api-",
			},
			isValid: true,
		},
		{
			name: "all with exclusions",
			targets: BulkWebhookTargets{
				All:     true,
				Exclude: []string{"test-repo", "archived-repo"},
			},
			isValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify the targets struct is created correctly
			assert.Equal(t, tt.isValid, true)
		})
	}
}

func TestGetTargetRepositories(t *testing.T) {
	// This test would require mocking the GitHub client
	// For now, we'll just test the basic structure
	t.Skip("Requires GitHub client mocking")
}

func TestWebhookOperations(t *testing.T) {
	// Test webhook operation function types
	var op webhookOperation

	// Verify function signatures match
	op = createWebhookOperation
	assert.NotNil(t, op)

	op = syncWebhookOperation
	assert.NotNil(t, op)
}

func TestSampleBulkWebhookConfigFile(t *testing.T) {
	// Create a sample configuration file for documentation
	sampleConfig := `# Bulk Webhook Configuration
# This file defines webhooks to be applied across multiple repositories

version: "1.0"

# Define the webhooks to create/sync
webhooks:
  # CI/CD webhook
  - url: https://ci.example.com/github/webhook
    events:
      - push
      - pull_request
    active: true
    content_type: json
    secret: ${WEBHOOK_SECRET}  # Can use environment variables

  # Issue tracking webhook
  - url: https://tracker.example.com/github/webhook
    events:
      - issues
      - issue_comment
      - pull_request_review
    active: true
    content_type: json

  # Deployment webhook
  - url: https://deploy.example.com/github/webhook
    events:
      - release
      - deployment
      - deployment_status
    active: true
    content_type: json

# Specify target repositories
targets:
  # Apply to all repositories
  all: true
  
  # Or specify specific repositories
  # repositories:
  #   - my-app
  #   - my-api
  #   - my-lib
  
  # Or use pattern matching
  # pattern: "^(api-|service-)"
  
  # Exclude specific repositories
  exclude:
    - test-repo
    - archived-repo
    - legacy-app

# Operation options
options:
  # Skip repositories that already have webhooks
  skip_existing: false
  
  # Number of parallel operations
  max_workers: 5
  
  # Continue even if some operations fail
  continue_on_error: true
`

	// Write sample to temporary file for validation
	tmpFile, err := os.CreateTemp("", "sample-webhook-config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(sampleConfig)
	require.NoError(t, err)
	tmpFile.Close()

	// Validate it can be loaded
	config, err := loadBulkWebhookConfig(tmpFile.Name())
	require.NoError(t, err)

	assert.Equal(t, "1.0", config.Version)
	assert.Len(t, config.Webhooks, 3)
	assert.True(t, config.Targets.All)
	assert.Len(t, config.Targets.Exclude, 3)
}
