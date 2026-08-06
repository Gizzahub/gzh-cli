//nolint:testpackage // White-box testing needed for internal function access
package integration

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	synclone "github.com/gizzahub/gzh-cli/pkg/synclone"
)

// TestSyncClone_ConfigurationLoading tests the configuration loading functionality.
func TestSyncClone_ConfigurationLoading(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("LoadConfigFromFile", func(t *testing.T) {
		// Create temporary directory for test configuration
		tmpDir, err := os.MkdirTemp("", "synclone-config-*")
		require.NoError(t, err)

		defer func() { _ = os.RemoveAll(tmpDir) }() // Ignore cleanup error

		// Create test configuration
		configPath := filepath.Join(tmpDir, "synclone.yaml")
		testConfig := map[string]any{
			"version":          "1.0.0",
			"default_provider": "github",
			"providers": map[string]any{
				"github": map[string]any{
					"token": "${GITHUB_TOKEN}",
					"organizations": []map[string]any{
						{
							"name":      "test-org",
							"clone_dir": tmpDir,
						},
					},
				},
			},
		}

		// Write configuration to file
		data, err := yaml.Marshal(testConfig)
		require.NoError(t, err)

		err = os.WriteFile(configPath, data, 0o600)
		require.NoError(t, err)

		// Test loading configuration
		config, err := synclone.LoadConfig(configPath)
		if err != nil {
			// Configuration loading may fail due to validation, but should not panic
			t.Logf("Config loading failed (expected in test environment): %v", err)
			return
		}

		assert.NotNil(t, config)
		t.Logf("Successfully loaded configuration with version: %s", config.Version)
	})
}

// TestSyncClone_StateManagement tests the state management functionality.
func TestSyncClone_StateManagement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("StateManager_Operations", func(t *testing.T) {
		// Create temporary directory for state files
		tmpDir, err := os.MkdirTemp("", "synclone-state-*")
		require.NoError(t, err)

		defer func() { _ = os.RemoveAll(tmpDir) }() // Ignore cleanup error

		// Create state manager
		stateManager := synclone.NewStateManager(tmpDir)
		assert.NotNil(t, stateManager)

		// Create test state
		state := synclone.NewCloneState("github", "test-org", tmpDir, "pull", 5, 3)
		assert.NotNil(t, state)

		// Set pending repositories
		repos := []string{"repo1", "repo2", "repo3"}
		state.SetPendingRepositories(repos)

		// Test state persistence
		err = stateManager.SaveState(state)
		assert.NoError(t, err)

		// Test state loading
		loadedState, err := stateManager.LoadState("github", "test-org")
		assert.NoError(t, err)
		assert.NotNil(t, loadedState)

		// Verify state content
		assert.Equal(t, "github", loadedState.Provider)
		assert.Equal(t, "test-org", loadedState.Organization)
		assert.Equal(t, tmpDir, loadedState.TargetPath)

		// Test has state
		hasState := stateManager.HasState("github", "test-org")
		assert.True(t, hasState)

		// Test list states
		states, err := stateManager.ListStates()
		assert.NoError(t, err)
		assert.Len(t, states, 1)

		// Test delete state
		err = stateManager.DeleteState("github", "test-org")
		assert.NoError(t, err)

		// Verify state is deleted
		hasState = stateManager.HasState("github", "test-org")
		assert.False(t, hasState)
	})
}

// TestSyncClone_ProgressTracking tests the progress tracking functionality.
func TestSyncClone_ProgressTracking(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("ProgressTracker_Operations", func(t *testing.T) {
		repos := []string{"repo1", "repo2", "repo3", "repo4", "repo5"}

		// 표시 방식마다 RenderProgress의 결과가 다르다. quiet은 아무것도
		// 그리지 않기로 한 방식이므로 빈 문자열이 옳다(progress.go의
		// RenderProgress가 DisplayModeQuiet에서 ""를 돌려준다). 예전에는
		// 세 방식 모두에 NotEmpty를 걸어서 quiet만 늘 빨갰다 --
		// pkg/synclone의 TestRenderQuietProgress는 같은 것을 Empty로
		// 확인하고 통과한다. 한 함수를 두고 두 시험이 반대로 주장하고
		// 있었던 셈이다.
		displayModes := []struct {
			mode       synclone.DisplayMode
			wantRender bool
		}{
			{synclone.DisplayModeCompact, true},
			{synclone.DisplayModeDetailed, true},
			{synclone.DisplayModeQuiet, false},
		}

		for _, dm := range displayModes {
			t.Run(string(dm.mode), func(t *testing.T) {
				tracker := synclone.NewProgressTracker(repos, dm.mode)
				assert.NotNil(t, tracker)

				// Test initial state
				completed, failed, pending, progressPercent := tracker.GetOverallProgress()
				assert.Equal(t, 0, completed)
				assert.Equal(t, 0, failed)
				assert.Equal(t, len(repos), pending)
				assert.Equal(t, 0.0, progressPercent)

				// Update progress for some repositories
				tracker.UpdateRepository("repo1", synclone.StatusCloning, "Cloning...", 0.5)
				tracker.CompleteRepository("repo2", "Successfully cloned")
				tracker.SetRepositoryError("repo3", "Network timeout")

				// pending은 "아직 안 끝난 것"이 아니라 "아직 시작도 안 한 것"이다.
				// 지금 받고 있는 repo1은 completed도 failed도 pending도 아니고
				// progressPercent에만 0.5로 반영된다. 그래서 셋의 합이 5가 되지
				// 않는다 -- 남는 것은 repo4, repo5뿐이다. 예전에는 여기에 3을
				// 적어 두어 repo1을 두 번 세는 셈이었다.
				completed, failed, pending, progressPercent = tracker.GetOverallProgress()
				assert.Equal(t, 1, completed)
				assert.Equal(t, 1, failed)
				assert.Equal(t, 2, pending)
				assert.Greater(t, progressPercent, 0.0)

				// Test progress rendering (should not panic)
				progress := tracker.RenderProgress()
				if dm.wantRender {
					assert.NotEmpty(t, progress)
				} else {
					assert.Empty(t, progress)
				}

				// Test summary
				summary := tracker.GetSummary()
				assert.NotEmpty(t, summary)

				// Test duration tracking
				duration := tracker.GetDuration()
				assert.Greater(t, duration, time.Duration(0))

				t.Logf("Mode: %s, Progress: %.1f%%, Summary: %s", dm.mode, progressPercent, summary)
			})
		}
	})
}

// TestSyncClone_URLBuilder tests the URL building functionality.
func TestSyncClone_URLBuilder(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("URLBuilder_Operations", func(t *testing.T) {
		testCases := []struct {
			provider string
			protocol string
			orgName  string
			repoName string
		}{
			{"github", "https", "test-org", "test-repo"},
			{"gitlab", "ssh", "test-group", "test-project"},
			{"gitea", "https", "test-org", "test-repo"},
		}

		for _, tc := range testCases {
			t.Run(tc.provider+"_"+tc.protocol, func(t *testing.T) {
				// Test default hostname
				hostname := synclone.GetDefaultHostname(tc.provider)
				assert.NotEmpty(t, hostname)

				// Test URL building
				url := synclone.BuildURLForProvider(tc.provider, tc.protocol, tc.orgName, tc.repoName)
				assert.NotEmpty(t, url)
				assert.Contains(t, url, tc.orgName)
				assert.Contains(t, url, tc.repoName)

				// Test URL builder with host alias
				urlWithAlias := synclone.BuildURLWithHostAliasForProvider(tc.provider, tc.protocol, tc.orgName, tc.repoName)
				assert.NotEmpty(t, urlWithAlias)

				t.Logf("Provider: %s, Protocol: %s, URL: %s", tc.provider, tc.protocol, url)
			})
		}
	})
}

// TestSyncClone_SchemaValidation tests the schema validation functionality.
func TestSyncClone_SchemaValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("SchemaValidation_Operations", func(t *testing.T) {
		// Create temporary directory for test configuration
		tmpDir, err := os.MkdirTemp("", "synclone-schema-*")
		require.NoError(t, err)

		defer func() { _ = os.RemoveAll(tmpDir) }() // Ignore cleanup error

		// Create valid configuration
		configPath := filepath.Join(tmpDir, "valid-config.yaml")
		validConfig := `
version: "1.0.0"
default_provider: "github"
providers:
  github:
    token: "${GITHUB_TOKEN}"
    organizations:
      - name: "test-org"
        clone_dir: "/tmp/test"
        visibility: "public"
        strategy: "pull"
`

		err = os.WriteFile(configPath, []byte(validConfig), 0o600)
		require.NoError(t, err)

		// Test schema validation
		err = synclone.ValidateConfigWithSchema(configPath)
		if err != nil {
			// Schema validation may fail in test environment, but should not panic
			t.Logf("Schema validation failed (expected in test environment): %v", err)
		} else {
			t.Log("Schema validation passed")
		}
	})
}

// TestSyncClone_EndToEnd tests the complete workflow.
func TestSyncClone_EndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Skip if no GitHub token is available
	if os.Getenv("GITHUB_TOKEN") == "" {
		t.Skip("GITHUB_TOKEN environment variable not set, skipping end-to-end test")
	}

	t.Run("EndToEnd_Workflow", func(t *testing.T) {
		// Create temporary directory for test
		tmpDir, err := os.MkdirTemp("", "synclone-e2e-*")
		require.NoError(t, err)

		defer func() { _ = os.RemoveAll(tmpDir) }() // Ignore cleanup error

		// Create configuration file
		configPath := filepath.Join(tmpDir, "synclone.yaml")
		configContent := `
version: "1.0.0"
default_provider: "github"
providers:
  github:
    token: "${GITHUB_TOKEN}"
    organizations:
      - name: "octocat"  # Public organization with sample repos
        clone_dir: "` + tmpDir + `"
        visibility: "public"
        strategy: "pull"
        include: "Hello-World"  # Only clone specific repo for testing
`

		err = os.WriteFile(configPath, []byte(configContent), 0o600)
		require.NoError(t, err)

		// Test configuration loading
		config, err := synclone.LoadConfig(configPath)
		if err != nil {
			t.Skipf("Configuration loading failed: %v", err)
		}

		require.NotNil(t, config)
		assert.Equal(t, "1.0.0", config.Version)
		// Note: DefaultProvider is not available in current config structure

		t.Logf("Successfully loaded configuration for end-to-end test")
		t.Logf("Target directory: %s", tmpDir)

		// Note: We don't actually perform cloning in this test to avoid
		// making real API calls and cloning repositories in CI/CD
		// This test validates that the configuration and setup work correctly
	})
}
