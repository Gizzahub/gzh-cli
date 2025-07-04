package builders

import (
	"github.com/gizzahub/gzh-manager-go/internal/env"
	"github.com/gizzahub/gzh-manager-go/pkg/config"
)

// ConfigBuilder provides a fluent interface for building test configurations
type ConfigBuilder struct {
	config *config.UnifiedConfig
}

// NewConfigBuilder creates a new ConfigBuilder with default values
func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{
		config: &config.UnifiedConfig{
			Version:         "1.0.0",
			DefaultProvider: "github",
			Providers:       make(map[string]*config.ProviderConfig),
		},
	}
}

// WithVersion sets the configuration version
func (b *ConfigBuilder) WithVersion(version string) *ConfigBuilder {
	b.config.Version = version
	return b
}

// WithDefaultProvider sets the default provider
func (b *ConfigBuilder) WithDefaultProvider(provider string) *ConfigBuilder {
	b.config.DefaultProvider = provider
	return b
}

// WithGitHubProvider adds a GitHub provider configuration
func (b *ConfigBuilder) WithGitHubProvider(token string) *ConfigBuilder {
	b.config.Providers["github"] = &config.ProviderConfig{
		Token:         token,
		Organizations: []*config.OrganizationConfig{},
	}
	return b
}

// WithGitLabProvider adds a GitLab provider configuration
func (b *ConfigBuilder) WithGitLabProvider(token string) *ConfigBuilder {
	b.config.Providers["gitlab"] = &config.ProviderConfig{
		Token:         token,
		Organizations: []*config.OrganizationConfig{},
	}
	return b
}

// WithGiteaProvider adds a Gitea provider configuration
func (b *ConfigBuilder) WithGiteaProvider(token string) *ConfigBuilder {
	b.config.Providers["gitea"] = &config.ProviderConfig{
		Token:         token,
		Organizations: []*config.OrganizationConfig{},
	}
	return b
}

// WithOrganization adds an organization to the specified provider
func (b *ConfigBuilder) WithOrganization(provider, name, cloneDir string) *ConfigBuilder {
	if b.config.Providers[provider] == nil {
		b.config.Providers[provider] = &config.ProviderConfig{
			Organizations: []*config.OrganizationConfig{},
		}
	}

	org := &config.OrganizationConfig{
		Name:       name,
		CloneDir:   cloneDir,
		Visibility: "all",
		Strategy:   "reset",
	}

	b.config.Providers[provider].Organizations = append(
		b.config.Providers[provider].Organizations,
		org,
	)
	return b
}

// WithOrganizationDetails adds an organization with full configuration
func (b *ConfigBuilder) WithOrganizationDetails(provider, name, cloneDir, visibility, strategy string) *ConfigBuilder {
	if b.config.Providers[provider] == nil {
		b.config.Providers[provider] = &config.ProviderConfig{
			Organizations: []*config.OrganizationConfig{},
		}
	}

	org := &config.OrganizationConfig{
		Name:       name,
		CloneDir:   cloneDir,
		Visibility: visibility,
		Strategy:   strategy,
	}

	b.config.Providers[provider].Organizations = append(
		b.config.Providers[provider].Organizations,
		org,
	)
	return b
}

// Build returns the constructed configuration
func (b *ConfigBuilder) Build() *config.UnifiedConfig {
	return b.config
}

// BuildYAML returns the configuration as YAML content
func (b *ConfigBuilder) BuildYAML() string {
	orgs := ""
	for provider, cfg := range b.config.Providers {
		orgs += provider + ":\n"
		if cfg.Token != "" {
			orgs += "    token: \"" + cfg.Token + "\"\n"
		}
		if len(cfg.Organizations) > 0 {
			orgs += "    organizations:\n"
			for _, org := range cfg.Organizations {
				orgs += "      - name: \"" + org.Name + "\"\n"
				orgs += "        clone_dir: \"" + org.CloneDir + "\"\n"
				orgs += "        visibility: \"" + org.Visibility + "\"\n"
				orgs += "        strategy: \"" + org.Strategy + "\"\n"
			}
		}
	}

	return `version: "` + b.config.Version + `"
default_provider: ` + b.config.DefaultProvider + `
providers:
  ` + orgs
}

// EnvironmentBuilder provides a fluent interface for building test environments
type EnvironmentBuilder struct {
	vars map[string]string
}

// NewEnvironmentBuilder creates a new EnvironmentBuilder
func NewEnvironmentBuilder() *EnvironmentBuilder {
	return &EnvironmentBuilder{
		vars: make(map[string]string),
	}
}

// WithVar adds an environment variable
func (b *EnvironmentBuilder) WithVar(key, value string) *EnvironmentBuilder {
	b.vars[key] = value
	return b
}

// WithGitHubToken adds a GitHub token
func (b *EnvironmentBuilder) WithGitHubToken(token string) *EnvironmentBuilder {
	b.vars["GITHUB_TOKEN"] = token
	return b
}

// WithGitLabToken adds a GitLab token
func (b *EnvironmentBuilder) WithGitLabToken(token string) *EnvironmentBuilder {
	b.vars["GITLAB_TOKEN"] = token
	return b
}

// WithGiteaToken adds a Gitea token
func (b *EnvironmentBuilder) WithGiteaToken(token string) *EnvironmentBuilder {
	b.vars["GITEA_TOKEN"] = token
	return b
}

// WithHome sets the HOME environment variable
func (b *EnvironmentBuilder) WithHome(home string) *EnvironmentBuilder {
	b.vars["HOME"] = home
	return b
}

// WithConfigPath sets the GZH_CONFIG_PATH environment variable
func (b *EnvironmentBuilder) WithConfigPath(path string) *EnvironmentBuilder {
	b.vars["GZH_CONFIG_PATH"] = path
	return b
}

// Build returns the constructed mock environment
func (b *EnvironmentBuilder) Build() env.Environment {
	return env.NewMockEnvironment(b.vars)
}

// BuildMap returns the environment variables as a map
func (b *EnvironmentBuilder) BuildMap() map[string]string {
	result := make(map[string]string)
	for k, v := range b.vars {
		result[k] = v
	}
	return result
}
