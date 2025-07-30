// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package container

import (
	"context"
	"time"

	"github.com/gizzahub/gzh-manager-go/pkg/git/provider"
)

// ContainerBuilder provides a fluent interface for building and configuring containers.
type ContainerBuilder struct {
	config          *ContainerConfig
	customFactories map[string]FactoryFunc
	customInstances map[string]interface{}
}

// NewContainerBuilder creates a new container builder.
func NewContainerBuilder() *ContainerBuilder {
	return &ContainerBuilder{
		config:          DefaultContainerConfig(),
		customFactories: make(map[string]FactoryFunc),
		customInstances: make(map[string]interface{}),
	}
}

// WithHTTPTimeout sets the HTTP client timeout.
func (b *ContainerBuilder) WithHTTPTimeout(timeout time.Duration) *ContainerBuilder {
	b.config.HTTPTimeout = timeout
	return b
}

// WithLogLevel sets the default log level.
func (b *ContainerBuilder) WithLogLevel(level string) *ContainerBuilder {
	b.config.DefaultLogLevel = level
	return b
}

// WithMetrics enables or disables metrics collection.
func (b *ContainerBuilder) WithMetrics(enabled bool) *ContainerBuilder {
	b.config.EnableMetrics = enabled
	return b
}

// WithHealthChecks enables or disables health checks.
func (b *ContainerBuilder) WithHealthChecks(enabled bool) *ContainerBuilder {
	b.config.EnableHealthChecks = enabled
	return b
}

// WithProviderCacheTimeout sets the provider cache timeout.
func (b *ContainerBuilder) WithProviderCacheTimeout(timeout time.Duration) *ContainerBuilder {
	b.config.ProviderCacheTimeout = timeout
	return b
}

// Register adds a custom factory function.
func (b *ContainerBuilder) Register(name string, factory FactoryFunc) *ContainerBuilder {
	b.customFactories[name] = factory
	return b
}

// RegisterInstance adds a custom instance.
func (b *ContainerBuilder) RegisterInstance(name string, instance interface{}) *ContainerBuilder {
	b.customInstances[name] = instance
	return b
}

// Build creates and configures the container.
func (b *ContainerBuilder) Build() *Container {
	container := NewContainer(b.config)

	// Register custom factories
	for name, factory := range b.customFactories {
		container.Register(name, factory)
	}

	// Register custom instances
	for name, instance := range b.customInstances {
		container.RegisterInstance(name, instance)
	}

	return container
}

// ContextualContainer wraps a container with a context for operation-scoped dependencies.
type ContextualContainer struct {
	*Container
	ctx context.Context
}

// NewContextualContainer creates a container that can provide context to dependencies.
func NewContextualContainer(ctx context.Context, container *Container) *ContextualContainer {
	return &ContextualContainer{
		Container: container,
		ctx:       ctx,
	}
}

// GetContext returns the associated context.
func (c *ContextualContainer) GetContext() context.Context {
	return c.ctx
}

// CreateProviderWithContext creates a provider using the container's context.
func (c *ContextualContainer) CreateProviderWithContext(providerType string, config *provider.ProviderConfig) (provider.GitProvider, error) {
	return c.CreateProvider(c.ctx, providerType, config)
}

// ContainerModule represents a module that can configure a container.
type ContainerModule interface {
	Configure(container *Container) error
}

// ModuleBuilder builds containers using modules.
type ModuleBuilder struct {
	config  *ContainerConfig
	modules []ContainerModule
}

// NewModuleBuilder creates a new module-based container builder.
func NewModuleBuilder() *ModuleBuilder {
	return &ModuleBuilder{
		config:  DefaultContainerConfig(),
		modules: make([]ContainerModule, 0),
	}
}

// WithConfig sets the container configuration.
func (b *ModuleBuilder) WithConfig(config *ContainerConfig) *ModuleBuilder {
	b.config = config
	return b
}

// AddModule adds a configuration module.
func (b *ModuleBuilder) AddModule(module ContainerModule) *ModuleBuilder {
	b.modules = append(b.modules, module)
	return b
}

// Build creates the container and applies all modules.
func (b *ModuleBuilder) Build() (*Container, error) {
	container := NewContainer(b.config)

	for _, module := range b.modules {
		if err := module.Configure(container); err != nil {
			return nil, err
		}
	}

	return container, nil
}

// Pre-defined modules for common configurations

// GitHubModule configures GitHub-specific dependencies.
type GitHubModule struct {
	Token string
}

// Configure implements ContainerModule.
func (m *GitHubModule) Configure(container *Container) error {
	if m.Token != "" {
		container.Register("githubToken", func(_ *Container) (interface{}, error) {
			return m.Token, nil
		})
	}
	return nil
}

// GitLabModule configures GitLab-specific dependencies.
type GitLabModule struct {
	Token   string
	BaseURL string
}

// Configure implements ContainerModule.
func (m *GitLabModule) Configure(container *Container) error {
	if m.Token != "" {
		container.Register("gitlabToken", func(_ *Container) (interface{}, error) {
			return m.Token, nil
		})
	}
	if m.BaseURL != "" {
		container.Register("gitlabBaseURL", func(_ *Container) (interface{}, error) {
			return m.BaseURL, nil
		})
	}
	return nil
}

// GiteaModule configures Gitea-specific dependencies.
type GiteaModule struct {
	Token   string
	BaseURL string
}

// Configure implements ContainerModule.
func (m *GiteaModule) Configure(container *Container) error {
	if m.Token != "" {
		container.Register("giteaToken", func(_ *Container) (interface{}, error) {
			return m.Token, nil
		})
	}
	if m.BaseURL != "" {
		container.Register("giteaBaseURL", func(_ *Container) (interface{}, error) {
			return m.BaseURL, nil
		})
	}
	return nil
}

// TestingModule configures dependencies for testing environments.
type TestingModule struct {
	MockProviders map[string]provider.GitProvider
	MockLogger    interface{}
}

// Configure implements ContainerModule.
func (m *TestingModule) Configure(container *Container) error {
	if m.MockLogger != nil {
		container.RegisterInstance("logger", m.MockLogger)
	}

	for providerType, mockProvider := range m.MockProviders {
		container.RegisterInstance("provider:"+providerType, mockProvider)
	}

	return nil
}
