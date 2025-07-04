package fixtures

import (
	"github.com/gizzahub/gzh-manager-go/internal/testutil/builders"
	"github.com/gizzahub/gzh-manager-go/pkg/config"
)

// ConfigFixtures provides common configuration fixtures for tests
type ConfigFixtures struct{}

// NewConfigFixtures creates a new ConfigFixtures instance
func NewConfigFixtures() *ConfigFixtures {
	return &ConfigFixtures{}
}

// SimpleGitHubConfig returns a simple GitHub configuration
func (f *ConfigFixtures) SimpleGitHubConfig() *config.UnifiedConfig {
	return builders.NewConfigBuilder().
		WithVersion("1.0.0").
		WithDefaultProvider("github").
		WithGitHubProvider("${GITHUB_TOKEN}").
		WithOrganization("github", "test-org", "~/repos/test-org").
		Build()
}

// MultiProviderConfig returns a configuration with multiple providers
func (f *ConfigFixtures) MultiProviderConfig() *config.UnifiedConfig {
	return builders.NewConfigBuilder().
		WithVersion("1.0.0").
		WithDefaultProvider("github").
		WithGitHubProvider("${GITHUB_TOKEN}").
		WithGitLabProvider("${GITLAB_TOKEN}").
		WithGiteaProvider("${GITEA_TOKEN}").
		WithOrganization("github", "github-org", "~/repos/github-org").
		WithOrganization("gitlab", "gitlab-group", "~/repos/gitlab-group").
		WithOrganization("gitea", "gitea-org", "~/repos/gitea-org").
		Build()
}

// ComplexGitHubConfig returns a complex GitHub configuration with multiple organizations
func (f *ConfigFixtures) ComplexGitHubConfig() *config.UnifiedConfig {
	return builders.NewConfigBuilder().
		WithVersion("1.0.0").
		WithDefaultProvider("github").
		WithGitHubProvider("${GITHUB_TOKEN}").
		WithOrganizationDetails("github", "public-org", "~/repos/public", "public", "reset").
		WithOrganizationDetails("github", "private-org", "~/repos/private", "private", "pull").
		WithOrganizationDetails("github", "all-org", "~/repos/all", "all", "reset").
		Build()
}

// InvalidConfig returns an invalid configuration for error testing
func (f *ConfigFixtures) InvalidConfig() *config.UnifiedConfig {
	return builders.NewConfigBuilder().
		WithVersion(""). // Invalid empty version
		WithDefaultProvider("invalid-provider").
		Build()
}

// ConfigWithEnvironmentVariables returns a configuration that uses environment variables
func (f *ConfigFixtures) ConfigWithEnvironmentVariables() *config.UnifiedConfig {
	return builders.NewConfigBuilder().
		WithVersion("1.0.0").
		WithDefaultProvider("github").
		WithGitHubProvider("${GITHUB_TOKEN}").
		WithOrganization("github", "test-org", "${HOME}/repos/test").
		Build()
}

// MinimalConfig returns a minimal valid configuration
func (f *ConfigFixtures) MinimalConfig() *config.UnifiedConfig {
	return builders.NewConfigBuilder().
		WithVersion("1.0.0").
		WithDefaultProvider("github").
		Build()
}

// ConfigYAMLFixtures provides YAML configuration fixtures
type ConfigYAMLFixtures struct{}

// NewConfigYAMLFixtures creates a new ConfigYAMLFixtures instance
func NewConfigYAMLFixtures() *ConfigYAMLFixtures {
	return &ConfigYAMLFixtures{}
}

// SimpleGitHubYAML returns a simple GitHub configuration as YAML
func (f *ConfigYAMLFixtures) SimpleGitHubYAML() string {
	return builders.NewConfigBuilder().
		WithVersion("1.0.0").
		WithDefaultProvider("github").
		WithGitHubProvider("${GITHUB_TOKEN}").
		WithOrganization("github", "test-org", "~/repos/test-org").
		BuildYAML()
}

// MultiProviderYAML returns a multi-provider configuration as YAML
func (f *ConfigYAMLFixtures) MultiProviderYAML() string {
	return builders.NewConfigBuilder().
		WithVersion("1.0.0").
		WithDefaultProvider("github").
		WithGitHubProvider("${GITHUB_TOKEN}").
		WithGitLabProvider("${GITLAB_TOKEN}").
		WithOrganization("github", "github-org", "~/repos/github").
		WithOrganization("gitlab", "gitlab-group", "~/repos/gitlab").
		BuildYAML()
}

// InvalidYAML returns invalid YAML for error testing
func (f *ConfigYAMLFixtures) InvalidYAML() string {
	return `version: "1.0.0"
providers:
  github:
    token: "unclosed string
    orgs:
      - name: test
`
}

// MalformedYAML returns malformed YAML for error testing
func (f *ConfigYAMLFixtures) MalformedYAML() string {
	return `version: "1.0.0"
providers:
  github:
    token: "test-token"
    orgs:
      - name: "test-org"
        match: "[invalid"`
}

// EnvironmentVariableYAML returns YAML with environment variables
func (f *ConfigYAMLFixtures) EnvironmentVariableYAML() string {
	return `version: "1.0.0"
default_provider: github
providers:
  github:
    token: "${TEST_GITHUB_TOKEN}"
    organizations:
      - name: "test-org"
        clone_dir: "${HOME}/repos"
        visibility: "all"
        strategy: "reset"`
}

// MinimalYAML returns minimal valid YAML
func (f *ConfigYAMLFixtures) MinimalYAML() string {
	return `version: "1.0.0"
default_provider: github
providers:
  github:
    token: "test-token"`
}
