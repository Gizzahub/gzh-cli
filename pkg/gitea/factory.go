package gitea

import (
	"context"
	"fmt"

	"github.com/gizzahub/gzh-cli/internal/env"
)

// GiteaProviderFactory defines the interface for creating Gitea-specific instances.
type GiteaProviderFactory interface {
	// CreateCloner creates a Gitea cloner with the specified token
	CreateCloner(ctx context.Context, token string) (GiteaCloner, error)

	// CreateClonerWithEnv creates a Gitea cloner with a specific environment
	CreateClonerWithEnv(ctx context.Context, token string, environment env.Environment) (GiteaCloner, error)

	// GetProviderName returns the provider name
	GetProviderName() string
}

// giteaProviderFactoryImpl implements the GiteaProviderFactory interface.
type giteaProviderFactoryImpl struct {
	environment env.Environment
}

// NewGiteaProviderFactory creates a new Gitea provider factory.
func NewGiteaProviderFactory(environment env.Environment) GiteaProviderFactory {
	if environment == nil {
		environment = env.NewOSEnvironment()
	}

	return &giteaProviderFactoryImpl{
		environment: environment,
	}
}

// CreateCloner creates a Gitea cloner with the specified token.
func (f *giteaProviderFactoryImpl) CreateCloner(ctx context.Context, token string) (GiteaCloner, error) {
	return f.CreateClonerWithEnv(ctx, token, f.environment)
}

// CreateClonerWithEnv creates a Gitea cloner with a specific environment.
func (f *giteaProviderFactoryImpl) CreateClonerWithEnv(ctx context.Context, token string, environment env.Environment) (GiteaCloner, error) {
	if token == "" {
		// Try to get token from environment
		token = environment.Get(env.CommonEnvironmentKeys.GiteaToken)
	}

	if token == "" {
		return nil, fmt.Errorf("gitea token is required")
	}

	// Create a specific Gitea cloner implementation
	return &giteaClonerImpl{
		Token:       token,
		Environment: environment,
	}, nil
}

// GetProviderName returns the provider name.
func (f *giteaProviderFactoryImpl) GetProviderName() string {
	return "gitea"
}

// GiteaCloner interface defines the contract for Gitea cloning operations.
type GiteaCloner interface {
	// CloneOrganization clones all repositories from a Gitea organization
	CloneOrganization(ctx context.Context, orgName, targetPath, strategy string) error

	// CloneRepository clones a specific repository
	CloneRepository(ctx context.Context, owner, repo, targetPath, strategy string) error

	// SetToken sets the Gitea token for authentication
	SetToken(token string)

	// GetToken returns the current Gitea token
	GetToken() string

	// GetProviderName returns the provider name
	GetProviderName() string
}

// giteaClonerImpl implements the GiteaCloner interface.
type giteaClonerImpl struct {
	Token       string
	Environment env.Environment
}

// CloneOrganization clones all repositories from a Gitea organization.
func (g *giteaClonerImpl) CloneOrganization(ctx context.Context, orgName, targetPath, strategy string) error {
	// Set token as environment variable if provided
	if g.Token != "" {
		if err := g.Environment.Set(env.CommonEnvironmentKeys.GiteaToken, g.Token); err != nil {
			return fmt.Errorf("failed to set Gitea token: %w", err)
		}
	}

	// Call the existing RefreshAll function
	// Note: strategy parameter is ignored for now since gitea.RefreshAll doesn't support it
	return RefreshAll(ctx, targetPath, orgName)
}

// CloneRepository clones a specific repository.
func (g *giteaClonerImpl) CloneRepository(ctx context.Context, owner, repo, targetPath, strategy string) error {
	// Set token as environment variable if provided
	if g.Token != "" {
		if err := g.Environment.Set(env.CommonEnvironmentKeys.GiteaToken, g.Token); err != nil {
			return fmt.Errorf("failed to set Gitea token: %w", err)
		}
	}

	// Implementation would call appropriate Gitea API functions
	// For now, this is a placeholder
	return fmt.Errorf("CloneRepository not yet implemented")
}

// SetToken sets the Gitea token for authentication.
func (g *giteaClonerImpl) SetToken(token string) {
	g.Token = token
}

// GetToken returns the current Gitea token.
func (g *giteaClonerImpl) GetToken() string {
	return g.Token
}

// GetProviderName returns the provider name.
func (g *giteaClonerImpl) GetProviderName() string {
	return "gitea"
}

// GiteaFactoryConfig holds configuration for the Gitea factory.
type GiteaFactoryConfig struct {
	// DefaultToken is the default token to use when none is specified
	DefaultToken string
	// Environment is the environment to use for token resolution
	Environment env.Environment
}

// DefaultGiteaFactoryConfig returns default Gitea factory configuration.
func DefaultGiteaFactoryConfig() *GiteaFactoryConfig {
	return &GiteaFactoryConfig{
		Environment: env.NewOSEnvironment(),
	}
}

// NewGiteaProviderFactoryWithConfig creates a new Gitea provider factory with configuration.
func NewGiteaProviderFactoryWithConfig(config *GiteaFactoryConfig) GiteaProviderFactory {
	if config == nil {
		config = DefaultGiteaFactoryConfig()
	}

	return NewGiteaProviderFactory(config.Environment)
}
