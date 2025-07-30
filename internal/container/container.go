// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package container

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	internalconfig "github.com/gizzahub/gzh-manager-go/internal/config"
	"github.com/gizzahub/gzh-manager-go/internal/env"
	"github.com/gizzahub/gzh-manager-go/internal/logger"
	"github.com/gizzahub/gzh-manager-go/pkg/config"
	"github.com/gizzahub/gzh-manager-go/pkg/git/provider"
	"github.com/gizzahub/gzh-manager-go/pkg/gitea"
	"github.com/gizzahub/gzh-manager-go/pkg/github"
	"github.com/gizzahub/gzh-manager-go/pkg/gitlab"
)

// Container is a dependency injection container that manages application dependencies.
type Container struct {
	mu            sync.RWMutex
	singletons    map[string]interface{}
	factories     map[string]FactoryFunc
	configuration *ContainerConfig
}

// FactoryFunc is a function that creates an instance of a dependency.
type FactoryFunc func(c *Container) (interface{}, error)

// ContainerConfig holds configuration for the dependency injection container.
type ContainerConfig struct {
	HTTPTimeout          time.Duration
	DefaultLogLevel      string
	EnableMetrics        bool
	EnableHealthChecks   bool
	ProviderCacheTimeout time.Duration
}

// DefaultContainerConfig returns default configuration for the container.
func DefaultContainerConfig() *ContainerConfig {
	return &ContainerConfig{
		HTTPTimeout:          30 * time.Second,
		DefaultLogLevel:      "info",
		EnableMetrics:        true,
		EnableHealthChecks:   true,
		ProviderCacheTimeout: 30 * time.Minute,
	}
}

// NewContainer creates a new dependency injection container.
func NewContainer(config *ContainerConfig) *Container {
	if config == nil {
		config = DefaultContainerConfig()
	}

	container := &Container{
		singletons:    make(map[string]interface{}),
		factories:     make(map[string]FactoryFunc),
		configuration: config,
	}

	// Register core dependencies
	container.registerCoreDependencies()

	return container
}

// Get retrieves a dependency by name, creating it if necessary.
func (c *Container) Get(name string) (interface{}, error) {
	c.mu.RLock()
	if instance, exists := c.singletons[name]; exists {
		c.mu.RUnlock()
		return instance, nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check pattern
	if instance, exists := c.singletons[name]; exists {
		return instance, nil
	}

	factory, exists := c.factories[name]
	if !exists {
		return nil, fmt.Errorf("dependency '%s' not registered", name)
	}

	instance, err := factory(c)
	if err != nil {
		return nil, fmt.Errorf("failed to create dependency '%s': %w", name, err)
	}

	c.singletons[name] = instance
	return instance, nil
}

// GetTyped retrieves a dependency by name and attempts to cast it to the specified type.
func GetTyped[T any](c *Container, name string) (T, error) {
	var zero T
	instance, err := c.Get(name)
	if err != nil {
		return zero, err
	}

	typed, ok := instance.(T)
	if !ok {
		return zero, fmt.Errorf("dependency '%s' is not of expected type", name)
	}

	return typed, nil
}

// Register adds a factory function for a dependency.
func (c *Container) Register(name string, factory FactoryFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.factories[name] = factory
}

// RegisterInstance registers a pre-created instance as a singleton.
func (c *Container) RegisterInstance(name string, instance interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.singletons[name] = instance
}

// Has checks if a dependency is registered.
func (c *Container) Has(name string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, exists := c.factories[name]
	return exists || c.singletons[name] != nil
}

// ListRegistered returns all registered dependency names.
func (c *Container) ListRegistered() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	names := make([]string, 0, len(c.factories)+len(c.singletons))
	for name := range c.factories {
		names = append(names, name)
	}
	for name := range c.singletons {
		names = append(names, name)
	}

	return names
}

// Clear removes all dependencies and clears the container.
func (c *Container) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.singletons = make(map[string]interface{})
	c.factories = make(map[string]FactoryFunc)
}

// registerCoreDependencies registers the core application dependencies.
func (c *Container) registerCoreDependencies() {
	// Environment service
	c.Register("env", func(_ *Container) (interface{}, error) {
		return env.NewOSEnvironment(), nil
	})

	// Logger
	c.Register("logger", func(_ *Container) (interface{}, error) {
		return logger.NewSimpleLogger("app"), nil
	})

	// HTTP client
	c.Register("httpClient", func(_ *Container) (interface{}, error) {
		return &http.Client{
			Timeout: c.configuration.HTTPTimeout,
		}, nil
	})

	// Configuration service
	c.Register("configService", func(container *Container) (interface{}, error) {
		return internalconfig.CreateDefaultConfigService()
	})

	// Provider factory
	c.Register("providerFactory", func(container *Container) (interface{}, error) {
		factory := provider.NewProviderFactory()

		// Register provider constructors
		if err := factory.RegisterProvider("github", github.CreateGitHubProvider); err != nil {
			return nil, fmt.Errorf("failed to register GitHub provider: %w", err)
		}
		if err := factory.RegisterProvider("gitlab", gitlab.CreateGitLabProvider); err != nil {
			return nil, fmt.Errorf("failed to register GitLab provider: %w", err)
		}
		if err := factory.RegisterProvider("gitea", gitea.CreateGiteaProvider); err != nil {
			return nil, fmt.Errorf("failed to register Gitea provider: %w", err)
		}

		return factory, nil
	})

	// Provider registry
	c.Register("providerRegistry", func(container *Container) (interface{}, error) {
		factory, err := GetTyped[*provider.ProviderFactory](container, "providerFactory")
		if err != nil {
			return nil, err
		}

		registryConfig := provider.RegistryConfig{
			EnableCaching:       true,
			CacheTimeout:        c.configuration.ProviderCacheTimeout,
			EnableHealthChecks:  c.configuration.EnableHealthChecks,
			HealthCheckInterval: 5 * time.Minute,
			MaxCacheSize:        100,
			EnableMetrics:       c.configuration.EnableMetrics,
			AutoCleanup:         true,
		}

		return provider.NewProviderRegistry(factory, registryConfig), nil
	})

	// Repository configuration service - placeholder
	c.Register("repoConfigService", func(container *Container) (interface{}, error) {
		return &RepoConfigService{}, nil
	})

	// GitHub client - for backward compatibility
	c.Register("githubClient", func(container *Container) (interface{}, error) {
		environment, err := GetTyped[env.Environment](container, "env")
		if err != nil {
			return nil, err
		}

		token := environment.Get(env.CommonEnvironmentKeys.GitHubToken)
		if token == "" {
			return nil, fmt.Errorf("GitHub token not found in environment")
		}

		return github.NewResilientGitHubClient(token), nil
	})

	// Config provider factory - for backward compatibility
	c.Register("configProviderFactory", func(container *Container) (interface{}, error) {
		environment, err := GetTyped[env.Environment](container, "env")
		if err != nil {
			return nil, err
		}

		return config.NewDefaultProviderFactory(environment), nil
	})

	// Bulk operation service
	c.Register("bulkOperationService", func(container *Container) (interface{}, error) {
		providerFactory, err := GetTyped[*config.DefaultProviderFactory](container, "configProviderFactory")
		if err != nil {
			return nil, err
		}

		// This would need actual config loading - placeholder for now
		emptyConfig := &config.Config{
			Providers: make(map[string]config.Provider),
		}

		return config.NewDefaultBulkOperationService(providerFactory, emptyConfig), nil
	})
}

// CreateProvider is a convenience method to create a Git provider.
func (c *Container) CreateProvider(ctx context.Context, providerType string, config *provider.ProviderConfig) (provider.GitProvider, error) {
	registry, err := GetTyped[*provider.ProviderRegistry](c, "providerRegistry")
	if err != nil {
		return nil, err
	}

	return registry.GetProviderByType(providerType, config)
}

// GetConfigService returns the configuration service.
func (c *Container) GetConfigService() (internalconfig.ConfigService, error) {
	return GetTyped[internalconfig.ConfigService](c, "configService")
}

// GetLogger returns the logger instance.
func (c *Container) GetLogger() (interface{}, error) {
	return c.Get("logger")
}

// GetEnvironment returns the environment service.
func (c *Container) GetEnvironment() (env.Environment, error) {
	return GetTyped[env.Environment](c, "env")
}

// GetProviderRegistry returns the provider registry.
func (c *Container) GetProviderRegistry() (*provider.ProviderRegistry, error) {
	return GetTyped[*provider.ProviderRegistry](c, "providerRegistry")
}

// GetRepoConfigService returns the repository configuration service.
func (c *Container) GetRepoConfigService() (*RepoConfigService, error) {
	return GetTyped[*RepoConfigService](c, "repoConfigService")
}

// Close cleans up resources and shuts down the container.
func (c *Container) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Close provider registry if it exists
	if registry, exists := c.singletons["providerRegistry"]; exists {
		if providerRegistry, ok := registry.(*provider.ProviderRegistry); ok {
			if err := providerRegistry.Close(); err != nil {
				return fmt.Errorf("failed to close provider registry: %w", err)
			}
		}
	}

	// Clear all singletons
	c.singletons = make(map[string]interface{})

	return nil
}

// DefaultContainer is the global default container instance.
var DefaultContainer = NewContainer(DefaultContainerConfig())

// Global convenience functions using the default container

// GetDefault retrieves a dependency from the default container.
func GetDefault(name string) (interface{}, error) {
	return DefaultContainer.Get(name)
}

// GetTypedDefault retrieves a typed dependency from the default container.
func GetTypedDefault[T any](name string) (T, error) {
	return GetTyped[T](DefaultContainer, name)
}

// RegisterDefault registers a dependency in the default container.
func RegisterDefault(name string, factory FactoryFunc) {
	DefaultContainer.Register(name, factory)
}

// RegisterInstanceDefault registers an instance in the default container.
func RegisterInstanceDefault(name string, instance interface{}) {
	DefaultContainer.RegisterInstance(name, instance)
}

// Placeholder types for services that need proper implementation

// RepoConfigService provides repository configuration management.
type RepoConfigService struct{}

// NewRepoConfigService creates a new repo config service.
func NewRepoConfigService() *RepoConfigService {
	return &RepoConfigService{}
}
