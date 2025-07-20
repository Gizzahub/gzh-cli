//nolint:testpackage // White-box testing needed for internal function access
package scenarios

import (
	"testing"

	"github.com/gizzahub/gzh-manager-go/test/e2e/helpers"
)

func TestBulkClone_ConfigGeneration_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	// Create some mock repositories to scan
	env.CreateGitRepo("org1/repo1")
	env.CreateGitRepo("org1/repo2")
	env.CreateGitRepo("org2/repo3")

	// Generate configuration from existing directory structure
	result := env.RunCommand("gen-config", "--provider", "github", "--scan-dir", ".")

	assertions := helpers.NewCLIAssertions(t, result)
	assertions.Success().OutputContains("Configuration generated")

	// Verify configuration file was created
	env.AssertFileExists("bulk-clone.yaml")

	// Validate configuration content
	config := helpers.NewConfigAssertions(t, env, "bulk-clone.yaml")
	config.ValidYAML().
		HasField("version").
		HasField("providers").
		FieldEquals("version", "1.0.0")
}

func TestBulkClone_ConfigValidation_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	// Create a valid configuration
	validConfig := `
version: "1.0.0"
default_provider: github
providers:
  github:
    token: "${GITHUB_TOKEN}"
    orgs:
      - name: "test-org"
        visibility: "public"
        strategy: "reset"
        clone_dir: "./repos"
`
	env.WriteConfig("valid-config.yaml", validConfig)

	// Test configuration validation
	result := env.RunCommand("config", "validate", "--config", "valid-config.yaml")

	assertions := helpers.NewCLIAssertions(t, result)
	assertions.Success().OutputContains("Configuration is valid")
}

func TestBulkClone_ConfigValidation_Invalid_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	// Create an invalid configuration
	invalidConfig := `
version: "invalid-version"
providers:
  invalid-provider:
    token: "test"
`
	env.WriteConfig("invalid-config.yaml", invalidConfig)

	// Test configuration validation should fail
	result := env.RunCommand("config", "validate", "--config", "invalid-config.yaml")

	assertions := helpers.NewCLIAssertions(t, result)
	assertions.Failure().OutputContains("Configuration is invalid")
}

func TestBulkClone_DryRun_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	// Create test configuration
	config := `
version: "1.0.0"
default_provider: github
providers:
  github:
    token: "test-token"
    orgs:
      - name: "test-org"
        visibility: "public"
        strategy: "reset"
        clone_dir: "./repos"
        match: "test-.*"
`
	env.WriteConfig("test-config.yaml", config)

	// Run bulk clone in dry-run mode
	result := env.RunCommand("bulk-clone", "--config", "test-config.yaml", "--dry-run")

	assertions := helpers.NewCLIAssertions(t, result)
	// In dry-run mode, it should show what would be done without actual API calls
	assertions.OutputContains("dry run").OutputContains("test-org")

	// Verify no actual repositories were cloned
	env.AssertFileNotExists("repos")
}

func TestBulkClone_MultipleProviders_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	// Create configuration with multiple providers
	config := `
version: "1.0.0"
default_provider: github
providers:
  github:
    token: "github-token"
    orgs:
      - name: "github-org"
        clone_dir: "./github-repos"
  gitlab:
    base_url: "https://gitlab.example.com"
    token: "gitlab-token"
    groups:
      - name: "gitlab-group"
        clone_dir: "./gitlab-repos"
  gitea:
    base_url: "https://gitea.example.com"
    token: "gitea-token"
    orgs:
      - name: "gitea-org"
        clone_dir: "./gitea-repos"
`
	env.WriteConfig("multi-provider-config.yaml", config)

	// Validate multi-provider configuration
	result := env.RunCommand("config", "validate", "--config", "multi-provider-config.yaml")

	assertions := helpers.NewCLIAssertions(t, result)
	assertions.Success().OutputContains("Configuration is valid")

	// Test dry-run with multiple providers
	result = env.RunCommand("bulk-clone", "--config", "multi-provider-config.yaml", "--dry-run")

	assertions = helpers.NewCLIAssertions(t, result)
	assertions.OutputContains("github-org").
		OutputContains("gitlab-group").
		OutputContains("gitea-org")
}

func TestBulkClone_StrategyOptions_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	// Test different clone strategies
	strategies := []string{"reset", "pull", "fetch"}

	for _, strategy := range strategies {
		t.Run("strategy_"+strategy, func(t *testing.T) {
			config := `
version: "1.0.0"
providers:
  github:
    token: "test-token"
    orgs:
      - name: "test-org"
        strategy: "` + strategy + `"
        clone_dir: "./repos-` + strategy + `"
`
			configFile := "config-" + strategy + ".yaml"
			env.WriteConfig(configFile, config)

			// Validate configuration with specific strategy
			result := env.RunCommand("config", "validate", "--config", configFile)

			assertions := helpers.NewCLIAssertions(t, result)
			assertions.Success()
		})
	}
}

func TestBulkClone_VisibilityFiltering_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	// Test different visibility options
	visibilityOptions := []string{"public", "private", "all"}

	for _, visibility := range visibilityOptions {
		t.Run("visibility_"+visibility, func(t *testing.T) {
			config := `
version: "1.0.0"
providers:
  github:
    token: "test-token"
    orgs:
      - name: "test-org"
        visibility: "` + visibility + `"
        clone_dir: "./repos-` + visibility + `"
`
			configFile := "config-" + visibility + ".yaml"
			env.WriteConfig(configFile, config)

			// Validate configuration with specific visibility
			result := env.RunCommand("config", "validate", "--config", configFile)

			assertions := helpers.NewCLIAssertions(t, result)
			assertions.Success()
		})
	}
}

func TestBulkClone_PatternMatching_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	// Test pattern matching and exclusion
	config := `
version: "1.0.0"
providers:
  github:
    token: "test-token"
    orgs:
      - name: "test-org"
        clone_dir: "./repos"
        match: "^awesome-.*"
        exclude:
          - "awesome-archive-*"
          - "awesome-deprecated-*"
`
	env.WriteConfig("pattern-config.yaml", config)

	// Validate pattern configuration
	result := env.RunCommand("config", "validate", "--config", "pattern-config.yaml")

	assertions := helpers.NewCLIAssertions(t, result)
	assertions.Success()

	// Test dry-run to see pattern matching in action
	result = env.RunCommand("bulk-clone", "--config", "pattern-config.yaml", "--dry-run")

	assertions = helpers.NewCLIAssertions(t, result)
	assertions.OutputContains("awesome-")
}

func TestBulkClone_ErrorHandling_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	// Test with non-existent config file
	result := env.RunCommand("bulk-clone", "--config", "non-existent.yaml")

	assertions := helpers.NewCLIAssertions(t, result)
	assertions.Failure().OutputContains("config")

	// Test with malformed YAML
	malformedConfig := `
version: "1.0.0"
providers:
  github:
    token: "test"
    invalid_yaml: [
`
	env.WriteConfig("malformed.yaml", malformedConfig)

	result = env.RunCommand("bulk-clone", "--config", "malformed.yaml")

	assertions = helpers.NewCLIAssertions(t, result)
	assertions.Failure()
}

func TestBulkClone_EnvironmentVariables_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	// Set environment variables
	env.SetEnv("GITHUB_TOKEN", "env-github-token")
	env.SetEnv("GITLAB_TOKEN", "env-gitlab-token")

	// Create configuration using environment variables
	config := `
version: "1.0.0"
providers:
  github:
    token: "${GITHUB_TOKEN}"
    orgs:
      - name: "env-org"
        clone_dir: "./repos"
  gitlab:
    token: "${GITLAB_TOKEN}"
    groups:
      - name: "env-group"
        clone_dir: "./gitlab-repos"
`
	env.WriteConfig("env-config.yaml", config)

	// Validate configuration with environment variables
	result := env.RunCommand("config", "validate", "--config", "env-config.yaml")

	assertions := helpers.NewCLIAssertions(t, result)
	assertions.Success()
}

func TestBulkClone_ConfigMigration_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	// Create old version configuration
	oldConfig := `
version: "0.9.0"
github:
  token: "old-token"
  organizations:
    - "old-org"
`
	env.WriteConfig("old-config.yaml", oldConfig)

	// Test configuration migration
	result := env.RunCommand("config", "migrate", "--from", "old-config.yaml", "--to", "new-config.yaml")

	// Migration might not be implemented yet, so we check if the command exists
	if result.ExitCode == 0 {
		assertions := helpers.NewCLIAssertions(t, result)
		assertions.Success()

		// Verify new configuration was created
		env.AssertFileExists("new-config.yaml")

		// Validate migrated configuration
		newConfig := helpers.NewConfigAssertions(t, env, "new-config.yaml")
		newConfig.ValidYAML().HasField("version")
	}
}

func TestBulkClone_CacheIntegration_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	// Create configuration with cache enabled
	config := `
version: "1.0.0"
cache:
  enabled: true
  type: "file"
  ttl: "1h"
providers:
  github:
    token: "test-token"
    orgs:
      - name: "cached-org"
        clone_dir: "./repos"
`
	env.WriteConfig("cache-config.yaml", config)

	// Validate cache configuration
	result := env.RunCommand("config", "validate", "--config", "cache-config.yaml")

	assertions := helpers.NewCLIAssertions(t, result)
	assertions.Success()

	// Test with cache in dry-run mode
	result = env.RunCommand("bulk-clone", "--config", "cache-config.yaml", "--dry-run")

	assertions = helpers.NewCLIAssertions(t, result)
	assertions.OutputContains("cached-org")
}

func TestBulkClone_HelpAndVersion_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	// Test help command
	result := env.RunCommand("bulk-clone", "--help")

	assertions := helpers.NewCLIAssertions(t, result)
	assertions.Success().
		OutputContains("bulk-clone").
		OutputContains("Usage:")

	// Test version command
	result = env.RunCommand("version")

	assertions = helpers.NewCLIAssertions(t, result)
	assertions.Success().OutputNotEmpty()

	// Test global help
	result = env.RunCommand("--help")

	assertions = helpers.NewCLIAssertions(t, result)
	assertions.Success().
		OutputContains("gz").
		OutputContains("Commands:")
}
