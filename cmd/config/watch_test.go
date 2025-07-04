package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	configservice "github.com/gizzahub/gzh-manager-go/internal/config"
	"github.com/gizzahub/gzh-manager-go/internal/env"
	configpkg "github.com/gizzahub/gzh-manager-go/pkg/config"
)

func TestWatchConfigHotReloading(t *testing.T) {
	// Skip this test on CI systems that may not support file watching
	if os.Getenv("CI") != "" {
		t.Skip("Skipping file watch test in CI environment")
	}

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "config-hot-reload-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "gzh.yaml")

	// Initial configuration
	initialContent := `version: "1.0.0"
default_provider: github
providers:
  github:
    token: "${GITHUB_TOKEN}"
    organizations:
      - name: "initial-org"
        clone_dir: "~/repos/initial"
        visibility: "all"
        strategy: "reset"
`

	err = os.WriteFile(configPath, []byte(initialContent), 0644)
	require.NoError(t, err)

	// Create service with hot-reloading enabled
	testEnv := env.NewMockEnvironment(map[string]string{
		"GITHUB_TOKEN": "test-token",
	})

	options := &configservice.ConfigServiceOptions{
		Environment:       testEnv,
		AutoMigrate:       false,
		WatchEnabled:      true,
		ValidationEnabled: true,
		SearchPaths:       []string{tempDir},
		ConfigName:        "gzh",
		ConfigTypes:       []string{"yaml"},
	}

	service, err := configservice.NewConfigService(options)
	require.NoError(t, err)

	// Load initial configuration
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	initialConfig, err := service.LoadConfiguration(ctx, configPath)
	require.NoError(t, err)
	assert.Equal(t, "github", initialConfig.DefaultProvider)
	assert.Len(t, initialConfig.Providers["github"].Organizations, 1)
	assert.Equal(t, "initial-org", initialConfig.Providers["github"].Organizations[0].Name)

	// Set up change tracking
	changeNotifications := make(chan *configpkg.UnifiedConfig, 3)
	callback := func(cfg *configpkg.UnifiedConfig) {
		select {
		case changeNotifications <- cfg:
		default:
			// Don't block if channel is full
		}
	}

	// Start watching
	err = service.WatchConfiguration(ctx, callback)
	require.NoError(t, err)
	defer service.StopWatching()

	t.Run("single configuration change", func(t *testing.T) {
		// Update configuration - change provider
		updatedContent := `version: "1.0.0"
default_provider: gitlab
providers:
  gitlab:
    token: "${GITLAB_TOKEN}"
    organizations:
      - name: "updated-org"
        clone_dir: "~/repos/updated"
        visibility: "public"
        strategy: "pull"
`

		err = os.WriteFile(configPath, []byte(updatedContent), 0644)
		require.NoError(t, err)

		// Wait for change notification
		select {
		case updatedConfig := <-changeNotifications:
			assert.Equal(t, "gitlab", updatedConfig.DefaultProvider)
			assert.Contains(t, updatedConfig.Providers, "gitlab")
			assert.Equal(t, "updated-org", updatedConfig.Providers["gitlab"].Organizations[0].Name)
		case <-time.After(3 * time.Second):
			t.Fatal("Configuration change notification was not received within timeout")
		}

		// Verify the service also has the updated configuration
		currentConfig := service.GetConfiguration()
		assert.Equal(t, "gitlab", currentConfig.DefaultProvider)
	})

	t.Run("multiple rapid changes", func(t *testing.T) {
		// Make several rapid changes
		changes := []struct {
			provider string
			orgName  string
		}{
			{"github", "rapid-change-1"},
			{"gitlab", "rapid-change-2"},
			{"gitea", "rapid-change-3"},
		}

		for i, change := range changes {
			content := fmt.Sprintf(`version: "1.0.0"
default_provider: %s
providers:
  %s:
    token: "${%s_TOKEN}"
    organizations:
      - name: "%s"
        clone_dir: "~/repos/%s"
        visibility: "all"
        strategy: "reset"
`, change.provider, change.provider, 
			strings.ToUpper(change.provider), change.orgName, change.orgName)

			err = os.WriteFile(configPath, []byte(content), 0644)
			require.NoError(t, err)

			// Small delay between changes
			time.Sleep(100 * time.Millisecond)

			// Check we get a notification (but don't check exact content due to rapid changes)
			select {
			case <-changeNotifications:
				// Got a notification, that's good
			case <-time.After(2 * time.Second):
				t.Fatalf("Change notification %d was not received within timeout", i+1)
			}
		}

		// Verify final state
		finalConfig := service.GetConfiguration()
		assert.Equal(t, "gitea", finalConfig.DefaultProvider)
		assert.Contains(t, finalConfig.Providers, "gitea")
	})

	t.Run("configuration validation during hot-reload", func(t *testing.T) {
		// Update with invalid configuration
		invalidContent := `version: "invalid-version"
default_provider: nonexistent
providers:
  github:
    # Missing required token
    organizations:
      - name: ""  # Invalid empty name
        clone_dir: ""  # Invalid empty clone_dir
`

		err = os.WriteFile(configPath, []byte(invalidContent), 0644)
		require.NoError(t, err)

		// Wait for change notification
		select {
		case <-changeNotifications:
			// Configuration should be loaded but validation should fail
			validationResult := service.GetValidationResult()
			assert.NotNil(t, validationResult)
			assert.False(t, validationResult.IsValid)
			assert.True(t, len(validationResult.Errors) > 0)
		case <-time.After(3 * time.Second):
			t.Fatal("Invalid configuration change notification was not received within timeout")
		}

		// Fix the configuration
		fixedContent := `version: "1.0.0"
default_provider: github
providers:
  github:
    token: "${GITHUB_TOKEN}"
    organizations:
      - name: "fixed-org"
        clone_dir: "~/repos/fixed"
        visibility: "all"
        strategy: "reset"
`

		err = os.WriteFile(configPath, []byte(fixedContent), 0644)
		require.NoError(t, err)

		// Wait for fix notification
		select {
		case fixedConfig := <-changeNotifications:
			assert.Equal(t, "1.0.0", fixedConfig.Version)
			assert.Equal(t, "github", fixedConfig.DefaultProvider)
			
			// Validation should now pass
			validationResult := service.GetValidationResult()
			assert.NotNil(t, validationResult)
			assert.True(t, validationResult.IsValid)
		case <-time.After(3 * time.Second):
			t.Fatal("Fixed configuration change notification was not received within timeout")
		}
	})
}

func TestWatchConfigCommand(t *testing.T) {
	// Create temporary directory and config file
	tempDir, err := os.MkdirTemp("", "config-watch-cmd-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "test-config.yaml")
	content := `version: "1.0.0"
default_provider: github
providers:
  github:
    token: "${GITHUB_TOKEN}"
    organizations:
      - name: "test-org"
        clone_dir: "~/repos/test"
`

	err = os.WriteFile(configPath, []byte(content), 0644)
	require.NoError(t, err)

	// Test config file discovery
	t.Run("find config file", func(t *testing.T) {
		// Change to temp directory to test auto-discovery
		originalDir, err := os.Getwd()
		require.NoError(t, err)
		defer os.Chdir(originalDir)

		err = os.Chdir(tempDir)
		require.NoError(t, err)

		// Rename to standard name
		err = os.Rename(configPath, filepath.Join(tempDir, "gzh.yaml"))
		require.NoError(t, err)

		foundPath, err := findConfigFile()
		assert.NoError(t, err)
		assert.Contains(t, foundPath, "gzh.yaml")
	})

	t.Run("print config summary", func(t *testing.T) {
		// Create test configuration
		testConfig := &configpkg.UnifiedConfig{
			Version:         "1.0.0",
			DefaultProvider: "github",
			Providers: map[string]*configpkg.ProviderConfig{
				"github": {
					Token: "${GITHUB_TOKEN}",
					Organizations: []*configpkg.OrganizationConfig{
						{
							Name:       "test-org",
							CloneDir:   "~/repos/test",
							Visibility: "all",
							Strategy:   "reset",
						},
					},
				},
			},
			Global: &configpkg.GlobalSettings{
				CloneBaseDir:    "$HOME/repos",
				DefaultStrategy: "reset",
			},
		}

		// This should not panic and should print the summary
		// In a real test environment, we might capture stdout to verify output
		printConfigSummary(testConfig)
	})
}