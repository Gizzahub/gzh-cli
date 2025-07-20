package github

import (
	"context"
	"fmt"

	"github.com/gizzahub/gzh-manager-go/internal/env"
)

// GitHubProviderFactory defines the interface for creating GitHub-specific instances.
type GitHubProviderFactory interface {
	// CreateCloner creates a GitHub cloner with the specified token
	CreateCloner(ctx context.Context, token string) (GitHubCloner, error)

	// CreateClonerWithEnv creates a GitHub cloner with a specific environment
	CreateClonerWithEnv(ctx context.Context, token string, environment env.Environment) (GitHubCloner, error)

	// CreateChangeLogger creates a GitHub change logger
	CreateChangeLogger(ctx context.Context, changelog *ChangeLog, options *LoggerOptions) (*ChangeLogger, error)

	// GetProviderName returns the provider name
	GetProviderName() string
}

// gitHubProviderFactoryImpl implements the GitHubProviderFactory interface.
type gitHubProviderFactoryImpl struct {
	environment env.Environment
}

// NewGitHubProviderFactory creates a new GitHub provider factory.
func NewGitHubProviderFactory(environment env.Environment) GitHubProviderFactory {
	if environment == nil {
		environment = env.NewOSEnvironment()
	}

	return &gitHubProviderFactoryImpl{
		environment: environment,
	}
}

// CreateCloner creates a GitHub cloner with the specified token.
func (f *gitHubProviderFactoryImpl) CreateCloner(ctx context.Context, token string) (GitHubCloner, error) {
	return f.CreateClonerWithEnv(ctx, token, f.environment)
}

// CreateClonerWithEnv creates a GitHub cloner with a specific environment.
func (f *gitHubProviderFactoryImpl) CreateClonerWithEnv(ctx context.Context, token string, environment env.Environment) (GitHubCloner, error) {
	if token == "" {
		// Try to get token from environment
		token = environment.Get(env.CommonEnvironmentKeys.GitHubToken)
	}

	if token == "" {
		return nil, fmt.Errorf("GitHub token is required")
	}

	// Create a specific GitHub cloner implementation
	return &gitHubClonerImpl{
		Token:       token,
		Environment: environment,
	}, nil
}

// CreateChangeLogger creates a GitHub change logger.
func (f *gitHubProviderFactoryImpl) CreateChangeLogger(ctx context.Context, changelog *ChangeLog, options *LoggerOptions) (*ChangeLogger, error) {
	if changelog == nil {
		// Create a default changelog with nil dependencies for basic logging
		// In production, these would be injected properly
		changelog = NewChangeLog(nil, nil)
	}

	return NewChangeLogger(changelog, options), nil
}

// GetProviderName returns the provider name.
func (f *gitHubProviderFactoryImpl) GetProviderName() string {
	return "github"
}

// GitHubCloner interface defines the contract for GitHub cloning operations.
type GitHubCloner interface {
	// CloneOrganization clones all repositories from a GitHub organization
	CloneOrganization(ctx context.Context, orgName, targetPath, strategy string) error

	// CloneRepository clones a specific repository
	CloneRepository(ctx context.Context, owner, repo, targetPath, strategy string) error

	// SetToken sets the GitHub token for authentication
	SetToken(token string)

	// GetToken returns the current GitHub token
	GetToken() string

	// GetProviderName returns the provider name
	GetProviderName() string
}

// gitHubClonerImpl implements the GitHubCloner interface.
type gitHubClonerImpl struct {
	Token       string
	Environment env.Environment
}

// CloneOrganization clones all repositories from a GitHub organization.
func (g *gitHubClonerImpl) CloneOrganization(ctx context.Context, orgName, targetPath, strategy string) error {
	// Set token as environment variable if provided
	if g.Token != "" {
		_ = g.Environment.Set(env.CommonEnvironmentKeys.GitHubToken, g.Token) //nolint:errcheck // Environment setup for authentication
	}

	// Call the existing RefreshAll function
	return RefreshAll(ctx, targetPath, orgName, strategy)
}

// CloneRepository clones a specific repository.
func (g *gitHubClonerImpl) CloneRepository(ctx context.Context, owner, repo, targetPath, strategy string) error {
	// Set token as environment variable if provided
	if g.Token != "" {
		_ = g.Environment.Set(env.CommonEnvironmentKeys.GitHubToken, g.Token) //nolint:errcheck // Environment setup for authentication
	}

	// Implementation would call appropriate GitHub API functions
	// For now, this is a placeholder
	return fmt.Errorf("CloneRepository not yet implemented")
}

// SetToken sets the GitHub token for authentication.
func (g *gitHubClonerImpl) SetToken(token string) {
	g.Token = token
}

// GetToken returns the current GitHub token.
func (g *gitHubClonerImpl) GetToken() string {
	return g.Token
}

// GetProviderName returns the provider name.
func (g *gitHubClonerImpl) GetProviderName() string {
	return "github"
}

// GitHubFactoryConfig holds configuration for the GitHub factory.
type GitHubFactoryConfig struct {
	// DefaultToken is the default token to use when none is specified
	DefaultToken string
	// Environment is the environment to use for token resolution
	Environment env.Environment
}

// DefaultGitHubFactoryConfig returns default GitHub factory configuration.
func DefaultGitHubFactoryConfig() *GitHubFactoryConfig {
	return &GitHubFactoryConfig{
		Environment: env.NewOSEnvironment(),
	}
}

// NewGitHubProviderFactoryWithConfig creates a new GitHub provider factory with configuration.
func NewGitHubProviderFactoryWithConfig(config *GitHubFactoryConfig) GitHubProviderFactory {
	if config == nil {
		config = DefaultGitHubFactoryConfig()
	}

	return NewGitHubProviderFactory(config.Environment)
}
