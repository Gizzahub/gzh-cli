package bulk_clone_test

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gizzahub/gzh-manager-go/pkg/bulk-clone"
)

// ExampleLoadConfig demonstrates how to load a bulk clone configuration
// from a YAML file.
func ExampleLoadConfig() {
	// Create a temporary configuration file
	tempDir := "/tmp/bulk-clone-config-example"
	os.MkdirAll(tempDir, 0o755)
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "bulk-clone.yaml")
	configContent := `
github:
  organizations:
    - name: "octocat"
      target: "./github-repos"
      strategy: "reset"
  token_env: "GITHUB_TOKEN"

gitlab:
  groups:
    - name: "gitlab-org"
      target: "./gitlab-repos"
      strategy: "pull"
  token_env: "GITLAB_TOKEN"

gitea:
  organizations:
    - name: "gitea"
      target: "./gitea-repos"
      strategy: "fetch"
  base_url: "https://gitea.com"
`

	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	if err != nil {
		log.Printf("Error creating config file: %v", err)
		return
	}

	// Load the configuration
	config, err := bulk_clone.LoadConfig(configPath)
	if err != nil {
		log.Printf("Error loading config: %v", err)
		return
	}

	fmt.Printf("Loaded configuration with:\n")
	fmt.Printf("- GitHub organizations: %d\n", len(config.GitHub.Organizations))
	fmt.Printf("- GitLab groups: %d\n", len(config.GitLab.Groups))
	fmt.Printf("- Gitea organizations: %d\n", len(config.Gitea.Organizations))

	if len(config.GitHub.Organizations) > 0 {
		fmt.Printf("- First GitHub org: %s\n", config.GitHub.Organizations[0].Name)
	}

	// Output: Configuration loaded successfully with multiple providers
}

// ExampleConfigValidation demonstrates configuration validation and error handling.
func ExampleConfigValidation() {
	tempDir := "/tmp/bulk-clone-validation-example"
	os.MkdirAll(tempDir, 0o755)
	defer os.RemoveAll(tempDir)

	// Create an invalid configuration file
	configPath := filepath.Join(tempDir, "invalid-config.yaml")
	invalidConfig := `
github:
  organizations:
    - name: ""  # Invalid: empty name
      target: ""  # Invalid: empty target
gitlab:
  groups: []  # Valid but empty
# Missing required fields
`

	err := os.WriteFile(configPath, []byte(invalidConfig), 0o644)
	if err != nil {
		log.Printf("Error creating config file: %v", err)
		return
	}

	// Attempt to load and validate
	_, err = bulk_clone.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("Validation caught error: %v\n", err)
	}

	// Create a valid configuration
	validConfigPath := filepath.Join(tempDir, "valid-config.yaml")
	validConfig := `
github:
  organizations:
    - name: "valid-org"
      target: "./repos"
      strategy: "reset"
  token_env: "GITHUB_TOKEN"
`

	err = os.WriteFile(validConfigPath, []byte(validConfig), 0o644)
	if err != nil {
		log.Printf("Error creating valid config: %v", err)
		return
	}

	config, err := bulk_clone.LoadConfig(validConfigPath)
	if err != nil {
		log.Printf("Error loading valid config: %v", err)
		return
	}

	fmt.Printf("Valid configuration loaded successfully\n")
	fmt.Printf("Organization: %s\n", config.GitHub.Organizations[0].Name)
	fmt.Printf("Strategy: %s\n", config.GitHub.Organizations[0].Strategy)

	// Output: Configuration validation demonstrates error detection and handling
}

// ExampleMultiProviderConfig demonstrates a comprehensive configuration
// with multiple Git hosting providers.
func ExampleMultiProviderConfig() {
	tempDir := "/tmp/bulk-clone-multi-provider-example"
	os.MkdirAll(tempDir, 0o755)
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "multi-provider.yaml")
	multiProviderConfig := `
github:
  organizations:
    - name: "github-org-1"
      target: "./github/org1"
      strategy: "reset"
      include_forks: false
    - name: "github-org-2"
      target: "./github/org2"
      strategy: "pull"
      include_forks: true
  token_env: "GITHUB_TOKEN"

gitlab:
  groups:
    - name: "gitlab-group-1"
      target: "./gitlab/group1"
      strategy: "fetch"
      include_subgroups: true
    - name: "gitlab-group-2"
      target: "./gitlab/group2"
      strategy: "reset"
      include_subgroups: false
  base_url: "https://gitlab.com"
  token_env: "GITLAB_TOKEN"

gitea:
  organizations:
    - name: "gitea-org"
      target: "./gitea/org"
      strategy: "pull"
  base_url: "https://gitea.com"
  token_env: "GITEA_TOKEN"

gogs:
  organizations:
    - name: "gogs-org"
      target: "./gogs/org"
      strategy: "reset"
  base_url: "https://try.gogs.io"
  token_env: "GOGS_TOKEN"
`

	err := os.WriteFile(configPath, []byte(multiProviderConfig), 0o644)
	if err != nil {
		log.Printf("Error creating config file: %v", err)
		return
	}

	config, err := bulk_clone.LoadConfig(configPath)
	if err != nil {
		log.Printf("Error loading config: %v", err)
		return
	}

	fmt.Printf("Multi-provider configuration loaded:\n")
	fmt.Printf("- GitHub: %d organizations\n", len(config.GitHub.Organizations))
	fmt.Printf("- GitLab: %d groups\n", len(config.GitLab.Groups))
	fmt.Printf("- Gitea: %d organizations\n", len(config.Gitea.Organizations))
	fmt.Printf("- Gogs: %d organizations\n", len(config.Gogs.Organizations))

	// Demonstrate accessing specific configuration details
	if len(config.GitHub.Organizations) > 0 {
		org := config.GitHub.Organizations[0]
		fmt.Printf("\nGitHub org details:\n")
		fmt.Printf("  Name: %s\n", org.Name)
		fmt.Printf("  Target: %s\n", org.Target)
		fmt.Printf("  Strategy: %s\n", org.Strategy)
		fmt.Printf("  Include forks: %t\n", org.IncludeForks)
	}

	// Output: Multi-provider configuration demonstrates comprehensive setup
}

// ExampleConfigStrategies demonstrates different cloning strategies
// and their use cases.
func ExampleConfigStrategies() {
	tempDir := "/tmp/bulk-clone-strategies-example"
	os.MkdirAll(tempDir, 0o755)
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "strategies.yaml")
	strategiesConfig := `
github:
  organizations:
    - name: "production-org"
      target: "./production"
      strategy: "reset"  # Clean slate, discard local changes
    - name: "development-org"
      target: "./development"
      strategy: "pull"   # Merge changes, preserve local work
    - name: "backup-org"
      target: "./backup"
      strategy: "fetch"  # Update refs only, no merge
  token_env: "GITHUB_TOKEN"
`

	err := os.WriteFile(configPath, []byte(strategiesConfig), 0o644)
	if err != nil {
		log.Printf("Error creating config file: %v", err)
		return
	}

	config, err := bulk_clone.LoadConfig(configPath)
	if err != nil {
		log.Printf("Error loading config: %v", err)
		return
	}

	fmt.Println("Clone strategies configured:")
	for _, org := range config.GitHub.Organizations {
		var description string
		switch org.Strategy {
		case "reset":
			description = "Hard reset + pull (discards local changes)"
		case "pull":
			description = "Merge remote changes with local changes"
		case "fetch":
			description = "Update remote tracking without changing working directory"
		default:
			description = "Unknown strategy"
		}

		fmt.Printf("- %s: %s - %s\n", org.Name, org.Strategy, description)
	}

	// Output: Clone strategies provide different update behaviors
}

// ExampleEnvironmentVariables demonstrates how environment variables
// are used for authentication and configuration.
func ExampleEnvironmentVariables() {
	// Save original environment
	originalGitHub := os.Getenv("GITHUB_TOKEN")
	originalGitLab := os.Getenv("GITLAB_TOKEN")
	defer func() {
		os.Setenv("GITHUB_TOKEN", originalGitHub)
		os.Setenv("GITLAB_TOKEN", originalGitLab)
	}()

	// Set example tokens (these would be real tokens in practice)
	os.Setenv("GITHUB_TOKEN", "ghp_example_token_here")
	os.Setenv("GITLAB_TOKEN", "glpat_example_token_here")

	tempDir := "/tmp/bulk-clone-env-example"
	os.MkdirAll(tempDir, 0o755)
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "env-config.yaml")
	envConfig := `
github:
  organizations:
    - name: "private-org"
      target: "./private-repos"
      strategy: "reset"
  token_env: "GITHUB_TOKEN"  # References environment variable

gitlab:
  groups:
    - name: "private-group"
      target: "./private-projects"
      strategy: "pull"
  token_env: "GITLAB_TOKEN"  # References environment variable
`

	err := os.WriteFile(configPath, []byte(envConfig), 0o644)
	if err != nil {
		log.Printf("Error creating config file: %v", err)
		return
	}

	config, err := bulk_clone.LoadConfig(configPath)
	if err != nil {
		log.Printf("Error loading config: %v", err)
		return
	}

	fmt.Println("Environment variable configuration:")
	fmt.Printf("GitHub token env: %s\n", config.GitHub.TokenEnv)
	fmt.Printf("GitLab token env: %s\n", config.GitLab.TokenEnv)

	// In real usage, tokens would be retrieved like this:
	githubToken := os.Getenv(config.GitHub.TokenEnv)
	gitlabToken := os.Getenv(config.GitLab.TokenEnv)

	fmt.Printf("GitHub token available: %t\n", githubToken != "")
	fmt.Printf("GitLab token available: %t\n", gitlabToken != "")

	// Output: Environment variables provide secure token management
}
