//nolint:testpackage // White-box testing needed for internal function access
package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gizzahub/gzh-manager-go/internal/env"
	"github.com/gizzahub/gzh-manager-go/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigService_LoadConfiguration(t *testing.T) {
	tests := []struct {
		name           string
		setupConfig    func(dir string) string
		configPath     string
		expectError    bool
		validateConfig func(t *testing.T, cfg *config.UnifiedConfig)
	}{
		{
			name: "load valid unified configuration",
			setupConfig: func(dir string) string {
				content := `version: "1.0.0"
default_provider: github
providers:
  github:
    token: "${GITHUB_TOKEN}"
    organizations:
      - name: "test-org"
        clone_dir: "~/repos/test-org"
        visibility: "all"
        strategy: "reset"
`
				path := filepath.Join(dir, "gzh.yaml")
				err := os.WriteFile(path, []byte(content), 0o600)
				require.NoError(t, err)
				return path
			},
			expectError: false,
			validateConfig: func(t *testing.T, cfg *config.UnifiedConfig) {
				t.Helper()
				assert.Equal(t, "1.0.0", cfg.Version)
				assert.Equal(t, "github", cfg.DefaultProvider)
				assert.Len(t, cfg.Providers, 1)
				assert.Contains(t, cfg.Providers, "github")
			},
		},
		{
			name: "load invalid configuration",
			setupConfig: func(dir string) string {
				content := `invalid: yaml: content`
				path := filepath.Join(dir, "invalid.yaml")
				err := os.WriteFile(path, []byte(content), 0o600)
				require.NoError(t, err)
				return path
			},
			expectError: true,
		},
		{
			name: "configuration file not found",
			setupConfig: func(dir string) string {
				return filepath.Join(dir, "nonexistent.yaml")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory
			tempDir, err := os.MkdirTemp("", "config-test-")
			require.NoError(t, err)

			defer func() {
				if err := os.RemoveAll(tempDir); err != nil {
					t.Logf("Warning: failed to remove temp dir: %v", err)
				}
			}()

			// Setup configuration file
			configPath := tt.setupConfig(tempDir)

			// Create service with test environment
			testEnv := env.NewMockEnvironment(map[string]string{
				"GITHUB_TOKEN": "test-token",
			})

			options := &ConfigServiceOptions{
				Environment:  testEnv,
				AutoMigrate:  false,
				WatchEnabled: false,
				SearchPaths:  []string{tempDir},
				ConfigName:   "gzh",
				ConfigTypes:  []string{"yaml"},
			}

			service, err := NewConfigService(options)
			require.NoError(t, err)

			// Load configuration
			ctx := context.Background()
			cfg, err := service.LoadConfiguration(ctx, configPath)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, cfg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cfg)

				if tt.validateConfig != nil {
					tt.validateConfig(t, cfg)
				}
			}
		})
	}
}

func TestConfigService_ReloadConfiguration(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "config-reload-test-")
	require.NoError(t, err)

	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Warning: failed to remove temp dir: %v", err)
		}
	}()

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
`

	err = os.WriteFile(configPath, []byte(initialContent), 0o600)
	require.NoError(t, err)

	// Create service
	testEnv := env.NewMockEnvironment(map[string]string{
		"GITHUB_TOKEN": "test-token",
	})

	options := &ConfigServiceOptions{
		Environment:  testEnv,
		AutoMigrate:  false,
		WatchEnabled: false,
		SearchPaths:  []string{tempDir},
		ConfigName:   "gzh",
		ConfigTypes:  []string{"yaml"},
	}

	service, err := NewConfigService(options)
	require.NoError(t, err)

	// Load initial configuration
	ctx := context.Background()
	cfg, err := service.LoadConfiguration(ctx, configPath)
	require.NoError(t, err)
	assert.Len(t, cfg.Providers["github"].Organizations, 1)
	assert.Equal(t, "initial-org", cfg.Providers["github"].Organizations[0].Name)

	// Update configuration file
	updatedContent := `version: "1.0.0"
default_provider: github
providers:
  github:
    token: "${GITHUB_TOKEN}"
    organizations:
      - name: "updated-org"
        clone_dir: "~/repos/updated"
      - name: "second-org"
        clone_dir: "~/repos/second"
`

	err = os.WriteFile(configPath, []byte(updatedContent), 0o600)
	require.NoError(t, err)

	// Reload configuration
	err = service.ReloadConfiguration(ctx)
	require.NoError(t, err)

	// Verify updated configuration
	cfg = service.GetConfiguration()
	assert.Len(t, cfg.Providers["github"].Organizations, 2)
	assert.Equal(t, "updated-org", cfg.Providers["github"].Organizations[0].Name)
	assert.Equal(t, "second-org", cfg.Providers["github"].Organizations[1].Name)
}

func TestConfigService_SaveConfiguration(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "config-save-test-")
	require.NoError(t, err)

	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Warning: failed to remove temp dir: %v", err)
		}
	}()

	// Create service
	testEnv := env.NewMockEnvironment(map[string]string{
		"GITHUB_TOKEN": "test-token",
	})

	options := &ConfigServiceOptions{
		Environment:  testEnv,
		AutoMigrate:  false,
		WatchEnabled: false,
		SearchPaths:  []string{tempDir},
		ConfigName:   "gzh",
		ConfigTypes:  []string{"yaml"},
	}

	service, err := NewConfigService(options)
	require.NoError(t, err)

	// Create test configuration
	testConfig := config.DefaultUnifiedConfig()
	testConfig.Providers["github"] = &config.ProviderConfig{
		Token: "${GITHUB_TOKEN}",
		Organizations: []*config.OrganizationConfig{
			{
				Name:       "test-org",
				CloneDir:   "~/repos/test",
				Visibility: "all",
				Strategy:   "reset",
			},
		},
	}

	// Save configuration
	ctx := context.Background()
	savePath := filepath.Join(tempDir, "saved-config.yaml")
	err = service.SaveConfiguration(ctx, testConfig, savePath)
	require.NoError(t, err)

	// Verify file was created
	assert.FileExists(t, savePath)

	// Load and verify saved configuration
	cfg, err := service.LoadConfiguration(ctx, savePath)
	require.NoError(t, err)
	assert.Equal(t, testConfig.Version, cfg.Version)
	assert.Equal(t, testConfig.DefaultProvider, cfg.DefaultProvider)
}

func TestConfigService_WatchConfiguration(t *testing.T) {
	// Skip this test on CI systems that may not support file watching
	if os.Getenv("CI") != "" {
		t.Skip("Skipping file watch test in CI environment")
	}

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "config-watch-test-")
	require.NoError(t, err)

	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Warning: failed to remove temp dir: %v", err)
		}
	}()

	configPath := filepath.Join(tempDir, "gzh.yaml")

	// Initial configuration
	initialContent := `version: "1.0.0"
default_provider: github
providers:
  github:
    token: "${GITHUB_TOKEN}"
    organizations:
      - name: "watch-test-org"
        clone_dir: "~/repos/watch-test"
`

	err = os.WriteFile(configPath, []byte(initialContent), 0o600)
	require.NoError(t, err)

	// Create service with watch enabled
	testEnv := env.NewMockEnvironment(map[string]string{
		"GITHUB_TOKEN": "test-token",
	})

	options := &ConfigServiceOptions{
		Environment:  testEnv,
		AutoMigrate:  false,
		WatchEnabled: true,
		SearchPaths:  []string{tempDir},
		ConfigName:   "gzh",
		ConfigTypes:  []string{"yaml"},
	}

	service, err := NewConfigService(options)
	require.NoError(t, err)

	// Load initial configuration
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = service.LoadConfiguration(ctx, configPath)
	require.NoError(t, err)

	// Set up watch callback
	callbackCalled := make(chan *config.UnifiedConfig, 1)
	callback := func(cfg *config.UnifiedConfig) {
		select {
		case callbackCalled <- cfg:
		default:
		}
	}

	// Start watching
	err = service.WatchConfiguration(ctx, callback)
	require.NoError(t, err)

	defer service.StopWatching()

	// Update configuration file
	updatedContent := `version: "1.0.0"
default_provider: gitlab
providers:
  gitlab:
    token: "${GITLAB_TOKEN}"
    organizations:
      - name: "watched-org"
        clone_dir: "~/repos/watched"
`

	err = os.WriteFile(configPath, []byte(updatedContent), 0o600)
	require.NoError(t, err)

	// Wait for callback to be called
	select {
	case updatedCfg := <-callbackCalled:
		assert.Equal(t, "gitlab", updatedCfg.DefaultProvider)
	case <-time.After(2 * time.Second):
		t.Fatal("Configuration watch callback was not called within timeout")
	}
}

func TestConfigService_Factory(t *testing.T) {
	factory := NewServiceFactory()

	t.Run("create with default options", func(t *testing.T) {
		service, err := factory.CreateDefaultConfigService()
		assert.NoError(t, err)
		assert.NotNil(t, service)
		assert.False(t, service.IsLoaded())
	})

	t.Run("create with custom environment", func(t *testing.T) {
		testEnv := env.NewMockEnvironment(map[string]string{
			"TEST_VAR": "test-value",
		})

		service, err := factory.CreateConfigServiceWithEnvironment(testEnv)
		assert.NoError(t, err)
		assert.NotNil(t, service)
	})

	t.Run("create with custom options", func(t *testing.T) {
		options := &ConfigServiceOptions{
			Environment:  env.NewOSEnvironment(),
			AutoMigrate:  false,
			WatchEnabled: false,
			SearchPaths:  []string{"/custom/path"},
			ConfigName:   "custom",
			ConfigTypes:  []string{"yaml"},
		}

		service, err := factory.CreateConfigService(options)
		assert.NoError(t, err)
		assert.NotNil(t, service)
	})

	t.Run("create with nil options", func(t *testing.T) {
		service, err := factory.CreateConfigService(nil)
		assert.Error(t, err)
		assert.Nil(t, service)
		assert.Contains(t, err.Error(), "cannot be nil")
	})
}

func TestConfigService_BulkCloneIntegration(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "config-bulk-test-")
	require.NoError(t, err)

	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Warning: failed to remove temp dir: %v", err)
		}
	}()

	configPath := filepath.Join(tempDir, "gzh.yaml")

	// Configuration with multiple providers
	content := `version: "1.0.0"
default_provider: github
providers:
  github:
    token: "${GITHUB_TOKEN}"
    organizations:
      - name: "github-org"
        clone_dir: "~/repos/github-org"
        visibility: "all"
        strategy: "reset"
  gitlab:
    token: "${GITLAB_TOKEN}"
    organizations:
      - name: "gitlab-group"
        clone_dir: "~/repos/gitlab-group"
        visibility: "public"
        strategy: "pull"
`

	err = os.WriteFile(configPath, []byte(content), 0o600)
	require.NoError(t, err)

	// Create service
	testEnv := env.NewMockEnvironment(map[string]string{
		"GITHUB_TOKEN": "github-token",
		"GITLAB_TOKEN": "gitlab-token",
	})

	options := &ConfigServiceOptions{
		Environment:  testEnv,
		AutoMigrate:  false,
		WatchEnabled: false,
		SearchPaths:  []string{tempDir},
		ConfigName:   "gzh",
		ConfigTypes:  []string{"yaml"},
	}

	service, err := NewConfigService(options)
	require.NoError(t, err)

	// Load configuration
	ctx := context.Background()
	_, err = service.LoadConfiguration(ctx, configPath)
	require.NoError(t, err)

	// Test getting all targets
	targets, err := service.GetBulkCloneTargets(ctx, "")
	require.NoError(t, err)
	assert.Len(t, targets, 2)

	// Test filtering by provider
	githubTargets, err := service.GetBulkCloneTargets(ctx, "github")
	require.NoError(t, err)
	assert.Len(t, githubTargets, 1)
	assert.Equal(t, "github", githubTargets[0].Provider)
	assert.Equal(t, "github-org", githubTargets[0].Name)

	gitlabTargets, err := service.GetBulkCloneTargets(ctx, "gitlab")
	require.NoError(t, err)
	assert.Len(t, gitlabTargets, 1)
	assert.Equal(t, "gitlab", gitlabTargets[0].Provider)
	assert.Equal(t, "gitlab-group", gitlabTargets[0].Name)

	// Test completed - GetConfiguredProviders method not available in current implementation
}
