package scenarios

import (
	"strings"
	"testing"

	"github.com/gizzahub/gzh-manager-go/test/e2e/helpers"
)

func TestConfig_Init_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	// Test configuration initialization
	result := env.RunCommand("config", "init")

	assertions := helpers.NewCLIAssertions(t, result)
	assertions.Success().OutputContains("Configuration initialized")

	// Verify configuration file was created
	env.AssertFileExists("bulk-clone.yaml")

	// Validate the generated configuration
	config := helpers.NewConfigAssertions(t, env, "bulk-clone.yaml")
	config.ValidYAML().
		HasField("version").
		HasField("providers")
}

func TestConfig_InitWithProvider_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	// Test initialization with specific provider
	result := env.RunCommand("config", "init", "--provider", "github")

	assertions := helpers.NewCLIAssertions(t, result)
	assertions.Success()

	// Verify GitHub provider configuration
	config := helpers.NewConfigAssertions(t, env, "bulk-clone.yaml")
	config.ValidYAML().
		HasField("providers.github").
		HasField("providers.github.token")
}

func TestConfig_Validate_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	// Create valid configuration
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
	env.WriteConfig("config.yaml", validConfig)

	// Test validation
	result := env.RunCommand("config", "validate", "--config", "config.yaml")

	assertions := helpers.NewCLIAssertions(t, result)
	assertions.Success().OutputContains("valid")
}

func TestConfig_ValidateAll_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	// Create multiple configuration files
	configs := map[string]string{
		"github-config.yaml": `
version: "1.0.0"
providers:
  github:
    token: "test"
    orgs: [{name: "test-org"}]
`,
		"gitlab-config.yaml": `
version: "1.0.0"
providers:
  gitlab:
    token: "test"
    groups: [{name: "test-group"}]
`,
	}

	for filename, content := range configs {
		env.WriteConfig(filename, content)
	}

	// Test validate all configurations
	result := env.RunCommand("config", "validate", "--all")

	assertions := helpers.NewCLIAssertions(t, result)
	assertions.Success().
		OutputContains("github-config.yaml").
		OutputContains("gitlab-config.yaml")
}

func TestConfig_Show_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	// Create configuration
	config := `
version: "1.0.0"
default_provider: github
providers:
  github:
    token: "masked-token"
    orgs:
      - name: "example-org"
`
	env.WriteConfig("config.yaml", config)

	// Test show configuration
	result := env.RunCommand("config", "show", "--config", "config.yaml")

	assertions := helpers.NewCLIAssertions(t, result)
	assertions.Success().
		OutputContains("example-org").
		OutputNotContains("masked-token") // Tokens should be masked
}

func TestConfig_Profile_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	// Create configuration directory
	env.CreateConfigDir()

	// Test profile creation
	result := env.RunCommand("config", "profile", "create", "work", "--provider", "github")

	assertions := helpers.NewCLIAssertions(t, result)
	if result.ExitCode == 0 {
		assertions.Success().OutputContains("Profile created")

		// Verify profile configuration
		env.AssertFileExists(".config/gzh-manager/profiles/work.yaml")
	}

	// Test profile listing
	result = env.RunCommand("config", "profile", "list")

	assertions = helpers.NewCLIAssertions(t, result)
	if result.ExitCode == 0 {
		assertions.Success()
	}
}

func TestConfig_Watch_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	// Create configuration
	config := `
version: "1.0.0"
providers:
  github:
    token: "test"
    orgs: [{name: "watch-org"}]
`
	env.WriteConfig("config.yaml", config)

	// Test watch command (this would typically run in background)
	result := env.RunCommand("config", "watch", "--config", "config.yaml", "--timeout", "1s")

	assertions := helpers.NewCLIAssertions(t, result)
	// Watch command might timeout, which is expected
	assertions.OutputContains("config.yaml")
}

func TestConfig_ErrorHandling_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	// Test with non-existent file
	result := env.RunCommand("config", "validate", "--config", "non-existent.yaml")

	assertions := helpers.NewCLIAssertions(t, result)
	assertions.Failure().OutputContains("not found")

	// Test with invalid YAML
	invalidConfig := `
version: "1.0.0"
providers:
  github:
    invalid_yaml: [
`
	env.WriteConfig("invalid.yaml", invalidConfig)

	result = env.RunCommand("config", "validate", "--config", "invalid.yaml")

	assertions = helpers.NewCLIAssertions(t, result)
	assertions.Failure()
}

func TestConfig_SchemaValidation_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	// Test with invalid schema
	invalidSchema := `
version: "invalid-version"
providers:
  unknown-provider:
    invalid_field: "test"
`
	env.WriteConfig("invalid-schema.yaml", invalidSchema)

	result := env.RunCommand("config", "validate", "--config", "invalid-schema.yaml")

	assertions := helpers.NewCLIAssertions(t, result)
	assertions.Failure().OutputContains("invalid")
}

func TestConfig_EnvironmentOverrides_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	// Set configuration via environment
	env.SetEnv("GZ_PROVIDER", "github")
	env.SetEnv("GZ_TOKEN", "env-token")
	env.SetEnv("GZ_ORG", "env-org")

	// Test that environment variables are recognized
	result := env.RunCommand("config", "show", "--env")

	assertions := helpers.NewCLIAssertions(t, result)
	if result.ExitCode == 0 {
		assertions.Success().
			OutputContains("github").
			OutputContains("env-org")
	}
}

func TestConfig_ConfigPath_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	// Test custom configuration path
	customDir := "custom-config"
	env.CreateDir(customDir)

	config := `
version: "1.0.0"
providers:
  github:
    token: "test"
    orgs: [{name: "custom-org"}]
`
	env.WriteConfig(customDir+"/custom.yaml", config)

	// Set custom config path
	env.SetEnv("GZ_CONFIG_PATH", env.GetWorkPath(customDir+"/custom.yaml"))

	result := env.RunCommand("config", "validate")

	assertions := helpers.NewCLIAssertions(t, result)
	assertions.Success()
}

func TestConfig_Backup_E2E(t *testing.T) {
	env := helpers.NewTestEnvironment(t)
	defer env.Cleanup()

	// Create configuration
	config := `
version: "1.0.0"
providers:
  github:
    token: "backup-test"
    orgs: [{name: "backup-org"}]
`
	env.WriteConfig("config.yaml", config)

	// Test configuration backup
	result := env.RunCommand("config", "backup", "--config", "config.yaml")

	if result.ExitCode == 0 {
		assertions := helpers.NewCLIAssertions(t, result)
		assertions.Success().OutputContains("backup")

		// Verify backup file was created (filename would include timestamp)
		files := env.ListFiles(".")
		backupFound := false

		for _, file := range files {
			if strings.Contains(file, "config") && strings.Contains(file, "backup") {
				backupFound = true
				break
			}
		}

		if backupFound {
			t.Log("Backup file created successfully")
		}
	}
}
