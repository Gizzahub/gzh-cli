package gitlab

import (
	"context"
	"fmt"

	"github.com/gizzahub/gzh-manager-go/internal/env"
)

// GitLabProviderFactory defines the interface for creating GitLab-specific instances
type GitLabProviderFactory interface {
	// CreateCloner creates a GitLab cloner with the specified token
	CreateCloner(ctx context.Context, token string) (GitLabCloner, error)
	
	// CreateClonerWithEnv creates a GitLab cloner with a specific environment
	CreateClonerWithEnv(ctx context.Context, token string, environment env.Environment) (GitLabCloner, error)
	
	// GetProviderName returns the provider name
	GetProviderName() string
}

// GitLabProviderFactoryImpl implements the GitLabProviderFactory interface
type GitLabProviderFactoryImpl struct {
	environment env.Environment
}

// NewGitLabProviderFactory creates a new GitLab provider factory
func NewGitLabProviderFactory(environment env.Environment) GitLabProviderFactory {
	if environment == nil {
		environment = env.NewOSEnvironment()
	}
	
	return &GitLabProviderFactoryImpl{
		environment: environment,
	}
}

// CreateCloner creates a GitLab cloner with the specified token
func (f *GitLabProviderFactoryImpl) CreateCloner(ctx context.Context, token string) (GitLabCloner, error) {
	return f.CreateClonerWithEnv(ctx, token, f.environment)
}

// CreateClonerWithEnv creates a GitLab cloner with a specific environment
func (f *GitLabProviderFactoryImpl) CreateClonerWithEnv(ctx context.Context, token string, environment env.Environment) (GitLabCloner, error) {
	if token == "" {
		// Try to get token from environment
		token = environment.Get(env.CommonEnvironmentKeys.GitLabToken)
	}
	
	if token == "" {
		return nil, fmt.Errorf("GitLab token is required")
	}
	
	// Create a specific GitLab cloner implementation
	return &GitLabClonerImpl{
		Token:       token,
		Environment: environment,
	}, nil
}

// GetProviderName returns the provider name
func (f *GitLabProviderFactoryImpl) GetProviderName() string {
	return "gitlab"
}

// GitLabCloner interface defines the contract for GitLab cloning operations
type GitLabCloner interface {
	// CloneGroup clones all repositories from a GitLab group
	CloneGroup(ctx context.Context, groupName, targetPath, strategy string) error
	
	// CloneProject clones a specific project
	CloneProject(ctx context.Context, groupName, projectName, targetPath, strategy string) error
	
	// SetToken sets the GitLab token for authentication
	SetToken(token string)
	
	// GetToken returns the current GitLab token
	GetToken() string
	
	// GetProviderName returns the provider name
	GetProviderName() string
}

// GitLabClonerImpl implements the GitLabCloner interface
type GitLabClonerImpl struct {
	Token       string
	Environment env.Environment
}

// CloneGroup clones all repositories from a GitLab group
func (g *GitLabClonerImpl) CloneGroup(ctx context.Context, groupName, targetPath, strategy string) error {
	// Set token as environment variable if provided
	if g.Token != "" {
		g.Environment.Set(env.CommonEnvironmentKeys.GitLabToken, g.Token)
	}
	
	// Call the existing RefreshAll function
	return RefreshAll(targetPath, groupName, strategy)
}

// CloneProject clones a specific project
func (g *GitLabClonerImpl) CloneProject(ctx context.Context, groupName, projectName, targetPath, strategy string) error {
	// Set token as environment variable if provided
	if g.Token != "" {
		g.Environment.Set(env.CommonEnvironmentKeys.GitLabToken, g.Token)
	}
	
	// Implementation would call appropriate GitLab API functions
	// For now, this is a placeholder
	return fmt.Errorf("CloneProject not yet implemented")
}

// SetToken sets the GitLab token for authentication
func (g *GitLabClonerImpl) SetToken(token string) {
	g.Token = token
}

// GetToken returns the current GitLab token
func (g *GitLabClonerImpl) GetToken() string {
	return g.Token
}

// GetProviderName returns the provider name
func (g *GitLabClonerImpl) GetProviderName() string {
	return "gitlab"
}

// GitLabFactoryConfig holds configuration for the GitLab factory
type GitLabFactoryConfig struct {
	// DefaultToken is the default token to use when none is specified
	DefaultToken string
	// Environment is the environment to use for token resolution
	Environment env.Environment
}

// DefaultGitLabFactoryConfig returns default GitLab factory configuration
func DefaultGitLabFactoryConfig() *GitLabFactoryConfig {
	return &GitLabFactoryConfig{
		Environment: env.NewOSEnvironment(),
	}
}

// NewGitLabProviderFactoryWithConfig creates a new GitLab provider factory with configuration
func NewGitLabProviderFactoryWithConfig(config *GitLabFactoryConfig) GitLabProviderFactory {
	if config == nil {
		config = DefaultGitLabFactoryConfig()
	}
	
	return NewGitLabProviderFactory(config.Environment)
}