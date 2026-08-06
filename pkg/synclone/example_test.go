package bulkclone_test

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	synclone "github.com/gizzahub/gzh-cli/pkg/synclone"
)

// Provider constants.
const (
	providerGitHub = "github"
	providerGitLab = "gitlab"
	providerGitea  = "gitea"
)

// ExampleLoadConfig demonstrates how to load a bulk clone configuration
// from a YAML file.
func ExampleLoadConfig() {
	// Create a temporary configuration file
	tempDir := "/tmp/bulk-clone-config-example"

	_ = os.MkdirAll(tempDir, 0o750)
	defer func() { _ = os.RemoveAll(tempDir) }()

	configPath := filepath.Join(tempDir, "bulk-clone.yaml")
	configContent := `
version: "1.0"
default:
  protocol: https
repo_roots:
  - root_path: "./github-repos"
    provider: "github"
    protocol: "https"
    org_name: "octocat"
    strategy: "reset"
  - root_path: "./gitlab-repos"
    provider: "gitlab"
    protocol: "https"
    org_name: "gitlab-org"
    strategy: "pull"
  - root_path: "./gitea-repos"
    provider: "gitea"
    protocol: "https"
    org_name: "gitea"
    strategy: "fetch"
`

	err := os.WriteFile(configPath, []byte(configContent), 0o600)
	if err != nil {
		log.Printf("Error creating config file: %v", err)
		return
	}

	// Load the configuration
	config, err := synclone.LoadConfig(configPath)
	if err != nil {
		log.Printf("Error loading config: %v", err)
		return
	}

	fmt.Printf("Loaded configuration with:\n")
	fmt.Printf("- Version: %s\n", config.Version)
	fmt.Printf("- Repository roots: %d\n", len(config.RepoRoots))

	// Count providers by type
	githubCount, gitlabCount, giteaCount := 0, 0, 0

	for _, root := range config.RepoRoots {
		switch root.Provider {
		case providerGitHub:
			githubCount++
		case providerGitLab:
			gitlabCount++
		case providerGitea:
			giteaCount++
		}
	}

	fmt.Printf("- GitHub organizations: %d\n", githubCount)
	fmt.Printf("- GitLab groups: %d\n", gitlabCount)
	fmt.Printf("- Gitea organizations: %d\n", giteaCount)

	if len(config.RepoRoots) > 0 && config.RepoRoots[0].Provider == providerGitHub {
		fmt.Printf("- First GitHub org: %s\n", config.RepoRoots[0].OrgName)
	}

	// Output:
	// Loaded configuration with:
	// - Version: 1.0
	// - Repository roots: 3
	// - GitHub organizations: 1
	// - GitLab groups: 1
	// - Gitea organizations: 1
	// - First GitHub org: octocat
}

// ExampleLoadConfig_validation demonstrates configuration validation and error handling.
func ExampleLoadConfig_validation() {
	tempDir := "/tmp/bulk-clone-validation-example"

	_ = os.MkdirAll(tempDir, 0o750)
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create an invalid configuration file
	configPath := filepath.Join(tempDir, "invalid-config.yaml")
	invalidConfig := `
version: "1.0"
repo_roots:
  - root_path: ""  # Invalid: empty path
    provider: "github"
    org_name: ""  # Invalid: empty name
# Missing required fields
`

	err := os.WriteFile(configPath, []byte(invalidConfig), 0o600)
	if err != nil {
		log.Printf("Error creating config file: %v", err)
		return
	}

	// Attempt to load and validate
	_, err = synclone.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("Validation caught error: %v\n", err)
	}

	// Create a valid configuration
	validConfigPath := filepath.Join(tempDir, "valid-config.yaml")
	validConfig := `version: "1.0"
default:
  protocol: https
repo_roots:
  - root_path: "./repos"
    provider: "github"
    protocol: "https"
    org_name: "valid-org"
`

	err = os.WriteFile(validConfigPath, []byte(validConfig), 0o600)
	if err != nil {
		log.Printf("Error creating valid config: %v", err)
		return
	}

	config, err := synclone.LoadConfig(validConfigPath)
	if err != nil {
		log.Printf("Error loading valid config: %v", err)
		return
	}

	fmt.Printf("Valid configuration loaded successfully\n")

	if len(config.RepoRoots) > 0 {
		fmt.Printf("Organization: %s\n", config.RepoRoots[0].OrgName)
		fmt.Printf("Provider: %s\n", config.RepoRoots[0].Provider)
	}

	// 첫 줄 "Error: ..."는 이 예시가 찍는 것이 아니라 synclone_config.go의
	// printValidationErrors가 표준출력으로 내보내는 것이다. Example은 표준출력을
	// 통째로 받으므로 기대값에도 들어간다.
	//
	// 아래 주석 앞의 빈 줄은 반드시 있어야 한다. Go는 함수 본문의 마지막 주석
	// 덩어리가 "Output:"으로 시작할 때만 그것을 기대값으로 인정한다. 빈 줄이
	// 없으면 위 설명과 한 덩어리가 되어 예시가 아예 실행되지 않는다.

	// Output:
	// Error: This field is required.
	// Validation caught error: failed to validate config file: Key: 'bulkCloneConfig.Default.Protocol' Error:Field validation for 'Protocol' failed on the 'required' tag
	// Valid configuration loaded successfully
	// Organization: valid-org
	// Provider: github
}

// ExampleMultiProviderConfig demonstrates a comprehensive configuration
// with multiple Git hosting providers.
func ExampleLoadConfig_multiProvider() {
	tempDir := "/tmp/bulk-clone-multi-provider-example"

	_ = os.MkdirAll(tempDir, 0o750)
	defer func() { _ = os.RemoveAll(tempDir) }()

	configPath := filepath.Join(tempDir, "multi-provider.yaml")
	multiProviderConfig := `
version: "1.0"
default:
  protocol: https
repo_roots:
  - root_path: "./github/org1"
    provider: "github"
    protocol: "https"
    org_name: "github-org-1"
    strategy: "reset"
    include_forks: false
  - root_path: "./github/org2"
    provider: "github"
    protocol: "https"
    org_name: "github-org-2"
    strategy: "pull"
    include_forks: true
  - root_path: "./gitlab/group1"
    provider: "gitlab"
    protocol: "https"
    org_name: "gitlab-group-1"
    strategy: "fetch"
    include_subgroups: true
  - root_path: "./gitlab/group2"
    provider: "gitlab"
    protocol: "https"
    org_name: "gitlab-group-2"
    strategy: "reset"
    include_subgroups: false
  - root_path: "./gitea/org"
    provider: "gitea"
    protocol: "https"
    org_name: "gitea-org"
    strategy: "pull"
  - root_path: "./gogs/org"
    provider: "gogs"
    protocol: "https"
    org_name: "gogs-org"
    strategy: "reset"
`

	err := os.WriteFile(configPath, []byte(multiProviderConfig), 0o644)
	if err != nil {
		log.Printf("Error creating config file: %v", err)
		return
	}

	config, err := synclone.LoadConfig(configPath)
	if err != nil {
		log.Printf("Error loading config: %v", err)
		return
	}

	fmt.Printf("Multi-provider configuration loaded:\n")
	// Count providers by type
	githubCount, gitlabCount, giteaCount, gogsCount := 0, 0, 0, 0

	for _, root := range config.RepoRoots {
		switch root.Provider {
		case providerGitHub:
			githubCount++
		case providerGitLab:
			gitlabCount++
		case providerGitea:
			giteaCount++
		case "gogs":
			gogsCount++
		}
	}

	fmt.Printf("- GitHub: %d organizations\n", githubCount)
	fmt.Printf("- GitLab: %d groups\n", gitlabCount)
	fmt.Printf("- Gitea: %d organizations\n", giteaCount)
	fmt.Printf("- Gogs: %d organizations\n", gogsCount)

	// Demonstrate accessing specific configuration details
	for _, root := range config.RepoRoots {
		if root.Provider != "github" {
			continue
		}
		fmt.Printf("\nGitHub org details:\n")
		fmt.Printf("  Name: %s\n", root.OrgName)
		fmt.Printf("  Target: %s\n", root.RootPath)
		fmt.Printf("  Protocol: %s\n", root.Protocol)

		break
	}

	// Output:
	// Multi-provider configuration loaded:
	// - GitHub: 2 organizations
	// - GitLab: 2 groups
	// - Gitea: 1 organizations
	// - Gogs: 1 organizations
	//
	// GitHub org details:
	//   Name: github-org-1
	//   Target: ./github/org1
	//   Protocol: https
}

// ExampleConfigStrategies demonstrates different cloning strategies
// and their use cases.
func ExampleLoadConfig_strategies() {
	tempDir := "/tmp/bulk-clone-strategies-example"

	_ = os.MkdirAll(tempDir, 0o750)
	defer func() { _ = os.RemoveAll(tempDir) }()

	configPath := filepath.Join(tempDir, "strategies.yaml")
	strategiesConfig := `
version: "1.0"
default:
  protocol: https
repo_roots:
  - root_path: "./production"
    provider: "github"
    protocol: "https"
    org_name: "production-org"
    strategy: "reset"  # Clean slate, discard local changes
  - root_path: "./development"
    provider: "github"
    protocol: "https"
    org_name: "development-org"
    strategy: "pull"   # Merge changes, preserve local work
  - root_path: "./backup"
    provider: "github"
    protocol: "https"
    org_name: "backup-org"
    strategy: "fetch"  # Update refs only, no merge
`

	err := os.WriteFile(configPath, []byte(strategiesConfig), 0o644)
	if err != nil {
		log.Printf("Error creating config file: %v", err)
		return
	}

	config, err := synclone.LoadConfig(configPath)
	if err != nil {
		log.Printf("Error loading config: %v", err)
		return
	}

	fmt.Println("Clone strategies configured:")

	for _, root := range config.RepoRoots {
		if root.Provider == "github" {
			// Note: Strategy is configured at the operation level, not in the config
			fmt.Printf("- %s: %s\n", root.OrgName, root.RootPath)
		}
	}

	// Output:
	// Clone strategies configured:
	// - production-org: ./production
	// - development-org: ./development
	// - backup-org: ./backup
}

// ExampleEnvironmentVariables demonstrates how environment variables
// are used for authentication and configuration.
func ExampleLoadConfig_environmentVariables() {
	// Save original environment
	originalGitHub := os.Getenv("GITHUB_TOKEN")
	originalGitLab := os.Getenv("GITLAB_TOKEN")

	defer func() {
		_ = os.Setenv("GITHUB_TOKEN", originalGitHub)
		_ = os.Setenv("GITLAB_TOKEN", originalGitLab)
	}()

	// Set example tokens (these would be real tokens in practice)
	if err := os.Setenv("GITHUB_TOKEN", "ghp_example_token_here"); err != nil {
		log.Printf("Warning: failed to set GITHUB_TOKEN: %v", err)
	}
	if err := os.Setenv("GITLAB_TOKEN", "glpat_example_token_here"); err != nil {
		log.Printf("Warning: failed to set GITLAB_TOKEN: %v", err)
	}

	tempDir := "/tmp/bulk-clone-env-example"

	_ = os.MkdirAll(tempDir, 0o750)
	defer func() { _ = os.RemoveAll(tempDir) }()

	configPath := filepath.Join(tempDir, "env-config.yaml")
	envConfig := `
version: "1.0"
default:
  protocol: https
repo_roots:
  - root_path: "./private-repos"
    provider: "github"
    protocol: "https"
    org_name: "private-org"
    strategy: "reset"
  - root_path: "./private-projects"
    provider: "gitlab"
    protocol: "https"
    org_name: "private-group"
    strategy: "pull"
`

	err := os.WriteFile(configPath, []byte(envConfig), 0o644)
	if err != nil {
		log.Printf("Error creating config file: %v", err)
		return
	}

	_, err = synclone.LoadConfig(configPath)
	if err != nil {
		log.Printf("Error loading config: %v", err)
		return
	}

	fmt.Println("Environment variable configuration:")
	// In real usage, tokens would be retrieved from environment
	githubToken := os.Getenv("GITHUB_TOKEN")
	gitlabToken := os.Getenv("GITLAB_TOKEN")

	fmt.Printf("GitHub token available: %t\n", githubToken != "")
	fmt.Printf("GitLab token available: %t\n", gitlabToken != "")

	// Output:
	// Environment variable configuration:
	// GitHub token available: true
	// GitLab token available: true
}
