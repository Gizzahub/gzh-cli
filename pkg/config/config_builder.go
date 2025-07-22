// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package config

// ConfigBuilder provides a fluent interface for building test configurations.
type ConfigBuilder struct { //nolint:revive // Type name maintained for clarity in builder pattern
	config *UnifiedConfig
}

// NewConfigBuilder creates a new ConfigBuilder with default values.
func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{
		config: &UnifiedConfig{
			Version:         "1.0.0",
			DefaultProvider: "github",
			Providers:       make(map[string]*ProviderConfig),
		},
	}
}

// WithVersion sets the configuration version.
func (b *ConfigBuilder) WithVersion(version string) *ConfigBuilder {
	b.config.Version = version
	return b
}

// WithDefaultProvider sets the default provider.
func (b *ConfigBuilder) WithDefaultProvider(provider string) *ConfigBuilder {
	b.config.DefaultProvider = provider
	return b
}

// WithGitHubProvider adds a GitHub provider configuration.
func (b *ConfigBuilder) WithGitHubProvider(token string) *ConfigBuilder {
	b.config.Providers["github"] = &ProviderConfig{
		Token:         token,
		Organizations: []*OrganizationConfig{},
	}

	return b
}

// WithGitLabProvider adds a GitLab provider configuration.
func (b *ConfigBuilder) WithGitLabProvider(token string) *ConfigBuilder {
	b.config.Providers["gitlab"] = &ProviderConfig{
		Token:         token,
		Organizations: []*OrganizationConfig{},
	}

	return b
}

// WithGiteaProvider adds a Gitea provider configuration.
func (b *ConfigBuilder) WithGiteaProvider(token string) *ConfigBuilder {
	b.config.Providers["gitea"] = &ProviderConfig{
		Token:         token,
		Organizations: []*OrganizationConfig{},
	}

	return b
}

// WithOrganization adds an organization to the specified provider.
func (b *ConfigBuilder) WithOrganization(provider, name, cloneDir string) *ConfigBuilder {
	if b.config.Providers[provider] == nil {
		b.config.Providers[provider] = &ProviderConfig{
			Organizations: []*OrganizationConfig{},
		}
	}

	org := &OrganizationConfig{
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

// WithOrganizationDetails adds an organization with full configuration.
func (b *ConfigBuilder) WithOrganizationDetails(provider, name, cloneDir, visibility, strategy string) *ConfigBuilder {
	if b.config.Providers[provider] == nil {
		b.config.Providers[provider] = &ProviderConfig{
			Organizations: []*OrganizationConfig{},
		}
	}

	org := &OrganizationConfig{
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

// Build returns the constructed configuration.
func (b *ConfigBuilder) Build() *UnifiedConfig {
	return b.config
}

// BuildYAML returns the configuration as YAML content.
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
