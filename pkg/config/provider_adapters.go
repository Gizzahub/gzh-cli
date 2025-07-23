// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package config

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gizzahub/gzh-manager-go/internal/env"
	"github.com/gizzahub/gzh-manager-go/pkg/gitea"
	"github.com/gizzahub/gzh-manager-go/pkg/github"
	"github.com/gizzahub/gzh-manager-go/pkg/gitlab"
)

const (
	defaultMainBranch = "main"
)

// GitHubProviderAdapter adapts the github package to implement ProviderService.
type GitHubProviderAdapter struct {
	token       string
	environment env.Environment
}

// NewGitHubProviderAdapter creates a new GitHub provider adapter.
func NewGitHubProviderAdapter(token string, environment env.Environment) *GitHubProviderAdapter {
	return &GitHubProviderAdapter{
		token:       token,
		environment: environment,
	}
}

// ListRepositories lists all repositories for a GitHub owner.
func (g *GitHubProviderAdapter) ListRepositories(ctx context.Context, owner string) ([]Repository, error) {
	repoNames, err := github.List(ctx, owner)
	if err != nil {
		return nil, err
	}

	repositories := make([]Repository, 0, len(repoNames))
	for _, name := range repoNames {
		// Get additional repository information
		defaultBranch, err := github.GetDefaultBranch(ctx, owner, name)
		if err != nil {
			defaultBranch = defaultMainBranch // fallback to main if error
		}

		repo := Repository{
			Name:          name,
			FullName:      fmt.Sprintf("%s/%s", owner, name),
			DefaultBranch: defaultBranch,
			CloneURL:      fmt.Sprintf("https://github.com/%s/%s.git", owner, name),
			SSHURL:        fmt.Sprintf("git@github.com:%s/%s.git", owner, name),
			HTMLURL:       fmt.Sprintf("https://github.com/%s/%s", owner, name),
		}
		repositories = append(repositories, repo)
	}

	return repositories, nil
}

// CloneRepository clones a single repository from GitHub to the target path.
func (g *GitHubProviderAdapter) CloneRepository(ctx context.Context, owner, repository, targetPath string) error {
	return github.Clone(ctx, targetPath, owner, repository)
}

// GetDefaultBranch retrieves the default branch name for a GitHub repository.
func (g *GitHubProviderAdapter) GetDefaultBranch(ctx context.Context, owner, repository string) (string, error) {
	return github.GetDefaultBranch(ctx, owner, repository)
}

// RefreshAll updates all repositories in the target path using the specified strategy.
func (g *GitHubProviderAdapter) RefreshAll(ctx context.Context, targetPath, owner, strategy string) error {
	return github.RefreshAll(ctx, targetPath, owner, strategy)
}

// CloneOrganization clones all repositories from a GitHub organization.
func (g *GitHubProviderAdapter) CloneOrganization(ctx context.Context, owner, targetPath, strategy string) error {
	return github.RefreshAll(ctx, targetPath, owner, strategy)
}

// SetToken configures the GitHub authentication token.
func (g *GitHubProviderAdapter) SetToken(token string) {
	g.token = token
	if g.token != "" && !strings.HasPrefix(g.token, "$") {
		if err := g.environment.Set(env.CommonEnvironmentKeys.GitHubToken, g.token); err != nil {
			// Environment variable setting failed - log but don't fail the operation
			fmt.Printf("Warning: failed to set GitHub token environment variable: %v\n", err)
		}
	}
}

// ValidateToken verifies that the configured GitHub token is valid.
func (g *GitHubProviderAdapter) ValidateToken(ctx context.Context) error {
	// Simple validation - try to make an API call
	_, err := github.List(ctx, "github") // Try to list github's own repositories
	return err
}

// GetProviderName returns the name identifier for this provider.
func (g *GitHubProviderAdapter) GetProviderName() string {
	return ProviderGitHub
}

// GetAPIEndpoint returns the base URL for the GitHub API.
func (g *GitHubProviderAdapter) GetAPIEndpoint() string {
	return "https://api.github.com"
}

// IsHealthy checks if the GitHub provider is healthy and accessible.
func (g *GitHubProviderAdapter) IsHealthy(ctx context.Context) error {
	// Check if we can reach the GitHub API
	return g.ValidateToken(ctx)
}

// GitLabProviderAdapter adapts the gitlab package to implement ProviderService.
type GitLabProviderAdapter struct {
	token       string
	environment env.Environment
}

// NewGitLabProviderAdapter creates a new GitLab provider adapter.
func NewGitLabProviderAdapter(token string, environment env.Environment) *GitLabProviderAdapter {
	return &GitLabProviderAdapter{
		token:       token,
		environment: environment,
	}
}

// ListRepositories retrieves all repositories for the given GitLab owner.
func (g *GitLabProviderAdapter) ListRepositories(ctx context.Context, owner string) ([]Repository, error) {
	repoNames, err := gitlab.List(ctx, owner)
	if err != nil {
		return nil, err
	}

	repositories := make([]Repository, 0, len(repoNames))
	for _, name := range repoNames {
		// Get additional repository information
		defaultBranch, err := gitlab.GetDefaultBranch(ctx, owner, name)
		if err != nil {
			defaultBranch = defaultMainBranch // fallback to main if error
		}

		repo := Repository{
			Name:          name,
			FullName:      fmt.Sprintf("%s/%s", owner, name),
			DefaultBranch: defaultBranch,
			CloneURL:      fmt.Sprintf("https://gitlab.com/%s/%s.git", owner, name),
			SSHURL:        fmt.Sprintf("git@gitlab.com:%s/%s.git", owner, name),
			HTMLURL:       fmt.Sprintf("https://gitlab.com/%s/%s", owner, name),
		}
		repositories = append(repositories, repo)
	}

	return repositories, nil
}

// CloneRepository clones a specific GitLab repository to the target path.
func (g *GitLabProviderAdapter) CloneRepository(ctx context.Context, owner, repository, targetPath string) error {
	return gitlab.Clone(ctx, targetPath, owner, repository, "")
}

// GetDefaultBranch retrieves the default branch name for a GitLab repository.
func (g *GitLabProviderAdapter) GetDefaultBranch(ctx context.Context, owner, repository string) (string, error) {
	return gitlab.GetDefaultBranch(ctx, owner, repository)
}

// RefreshAll refreshes all repositories in the target path for the given GitLab owner.
func (g *GitLabProviderAdapter) RefreshAll(ctx context.Context, targetPath, owner, strategy string) error {
	return gitlab.RefreshAll(ctx, targetPath, owner, strategy)
}

// CloneOrganization clones all repositories from a GitLab organization.
func (g *GitLabProviderAdapter) CloneOrganization(ctx context.Context, owner, targetPath, strategy string) error {
	return gitlab.RefreshAll(ctx, targetPath, owner, strategy)
}

// SetToken sets the authentication token for the GitLab provider.
func (g *GitLabProviderAdapter) SetToken(token string) {
	g.token = token
	if g.token != "" && !strings.HasPrefix(g.token, "$") {
		if err := g.environment.Set(env.CommonEnvironmentKeys.GitLabToken, g.token); err != nil {
			// Environment variable setting failed - log but don't fail the operation
			fmt.Printf("Warning: failed to set GitLab token environment variable: %v\n", err)
		}
	}
}

// ValidateToken validates the GitLab authentication token.
func (g *GitLabProviderAdapter) ValidateToken(ctx context.Context) error {
	// Simple validation - try to make an API call
	_, err := gitlab.List(ctx, "gitlab-org") // Try to list gitlab-org repositories
	return err
}

// GetProviderName returns the provider name for GitLab.
func (g *GitLabProviderAdapter) GetProviderName() string {
	return ProviderGitLab
}

// GetAPIEndpoint returns the API endpoint for GitLab.
func (g *GitLabProviderAdapter) GetAPIEndpoint() string {
	return "https://gitlab.com/api/v4"
}

// IsHealthy checks if the GitLab provider is healthy and accessible.
func (g *GitLabProviderAdapter) IsHealthy(ctx context.Context) error {
	return g.ValidateToken(ctx)
}

// GiteaProviderAdapter adapts the gitea package to implement ProviderService.
type GiteaProviderAdapter struct {
	token       string
	environment env.Environment
}

// NewGiteaProviderAdapter creates a new Gitea provider adapter.
func NewGiteaProviderAdapter(token string, environment env.Environment) *GiteaProviderAdapter {
	return &GiteaProviderAdapter{
		token:       token,
		environment: environment,
	}
}

// ListRepositories retrieves all repositories for the given Gitea owner.
func (g *GiteaProviderAdapter) ListRepositories(ctx context.Context, owner string) ([]Repository, error) {
	repoNames, err := gitea.List(ctx, owner)
	if err != nil {
		return nil, err
	}

	repositories := make([]Repository, 0, len(repoNames))
	for _, name := range repoNames {
		// Get additional repository information
		defaultBranch, err := gitea.GetDefaultBranch(ctx, owner, name)
		if err != nil {
			defaultBranch = defaultMainBranch // fallback to main if error
		}

		repo := Repository{
			Name:          name,
			FullName:      fmt.Sprintf("%s/%s", owner, name),
			DefaultBranch: defaultBranch,
			CloneURL:      fmt.Sprintf("https://gitea.com/%s/%s.git", owner, name),
			SSHURL:        fmt.Sprintf("git@gitea.com:%s/%s.git", owner, name),
			HTMLURL:       fmt.Sprintf("https://gitea.com/%s/%s", owner, name),
		}
		repositories = append(repositories, repo)
	}

	return repositories, nil
}

// CloneRepository clones a specific Gitea repository to the target path.
func (g *GiteaProviderAdapter) CloneRepository(ctx context.Context, owner, repository, targetPath string) error {
	return gitea.Clone(ctx, targetPath, owner, repository, "")
}

// GetDefaultBranch retrieves the default branch name for a Gitea repository.
func (g *GiteaProviderAdapter) GetDefaultBranch(ctx context.Context, owner, repository string) (string, error) {
	return gitea.GetDefaultBranch(ctx, owner, repository)
}

// RefreshAll refreshes all repositories in the target path for the given Gitea owner.
func (g *GiteaProviderAdapter) RefreshAll(ctx context.Context, targetPath, owner, strategy string) error {
	// Note: gitea.RefreshAll doesn't support strategy parameter
	return gitea.RefreshAll(ctx, targetPath, owner)
}

// CloneOrganization clones all repositories from a Gitea organization.
func (g *GiteaProviderAdapter) CloneOrganization(ctx context.Context, owner, targetPath, strategy string) error {
	return gitea.RefreshAll(ctx, targetPath, owner)
}

// SetToken sets the authentication token for the Gitea provider.
func (g *GiteaProviderAdapter) SetToken(token string) {
	g.token = token
	if g.token != "" && !strings.HasPrefix(g.token, "$") {
		if err := g.environment.Set(env.CommonEnvironmentKeys.GiteaToken, g.token); err != nil {
			// Environment variable setting failed - log but don't fail the operation
			fmt.Printf("Warning: failed to set Gitea token environment variable: %v\n", err)
		}
	}
}

// ValidateToken validates the Gitea authentication token.
func (g *GiteaProviderAdapter) ValidateToken(ctx context.Context) error {
	// Simple validation - try to make an API call
	_, err := gitea.List(ctx, "gitea") // Try to list gitea's own repositories
	return err
}

// GetProviderName returns the provider name for Gitea.
func (g *GiteaProviderAdapter) GetProviderName() string {
	return ProviderGitea
}

// GetAPIEndpoint returns the API endpoint for Gitea.
func (g *GiteaProviderAdapter) GetAPIEndpoint() string {
	return "https://gitea.com/api/v1"
}

// IsHealthy checks if the Gitea provider is healthy and accessible.
func (g *GiteaProviderAdapter) IsHealthy(ctx context.Context) error {
	return g.ValidateToken(ctx)
}

// DefaultProviderFactory implements ProviderFactory using adapter pattern.
type DefaultProviderFactory struct {
	environment env.Environment
}

// NewDefaultProviderFactory creates a new default provider factory.
func NewDefaultProviderFactory(environment env.Environment) *DefaultProviderFactory {
	return &DefaultProviderFactory{
		environment: environment,
	}
}

// CreateProvider creates a provider service for the given provider name and configuration.
func (f *DefaultProviderFactory) CreateProvider(ctx context.Context, providerName string, config ProviderConfig) (ProviderService, error) {
	switch strings.ToLower(providerName) {
	case ProviderGitHub:
		adapter := NewGitHubProviderAdapter(config.Token, f.environment)
		return adapter, nil
	case ProviderGitLab:
		adapter := NewGitLabProviderAdapter(config.Token, f.environment)
		return adapter, nil
	case ProviderGitea:
		adapter := NewGiteaProviderAdapter(config.Token, f.environment)
		return adapter, nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", providerName)
	}
}

// GetSupportedProviders returns a list of supported provider names.
func (f *DefaultProviderFactory) GetSupportedProviders() []string {
	return []string{ProviderGitHub, ProviderGitLab, ProviderGitea}
}

// ValidateProviderConfig validates the configuration for a given provider.
func (f *DefaultProviderFactory) ValidateProviderConfig(providerName string, config ProviderConfig) error {
	// Basic validation
	if config.Token == "" {
		return fmt.Errorf("token is required for provider %s", providerName)
	}

	// Provider-specific validation can be added here
	switch strings.ToLower(providerName) {
	case ProviderGitHub, ProviderGitLab, ProviderGitea:
		// All these providers require a token
		if strings.TrimSpace(config.Token) == "" {
			return fmt.Errorf("token cannot be empty for provider %s", providerName)
		}
	default:
		return fmt.Errorf("unsupported provider: %s", providerName)
	}

	return nil
}

// DefaultBulkOperationService implements BulkOperationService.
type DefaultBulkOperationService struct {
	factory *DefaultProviderFactory
	config  *Config
}

// NewDefaultBulkOperationService creates a new bulk operation service.
func NewDefaultBulkOperationService(factory *DefaultProviderFactory, config *Config) *DefaultBulkOperationService {
	return &DefaultBulkOperationService{
		factory: factory,
		config:  config,
	}
}

// CloneAll clones repositories from all configured providers.
func (s *DefaultBulkOperationService) CloneAll(ctx context.Context, request *BulkCloneRequest) (*BulkCloneResult, error) {
	result := &BulkCloneResult{
		Results: make([]TargetResult, 0),
	}

	// Process each configured provider
	for providerName, providerConfig := range s.config.Providers {
		if len(request.Providers) > 0 && !contains(request.Providers, providerName) {
			continue
		}

		config := ProviderConfig{Token: providerConfig.Token}

		provider, err := s.factory.CreateProvider(ctx, providerName, config)
		if err != nil {
			result.FailedTargets++
			continue
		}

		// Process organizations for this provider
		for _, org := range request.Organizations {
			if request.DryRun {
				// In dry run mode, just validate the operation
				repos, err := provider.ListRepositories(ctx, org)
				if err != nil {
					result.FailedTargets++
					continue
				}

				result.TotalTargets += len(repos)

				continue
			}

			err := provider.CloneOrganization(ctx, org, request.TargetPath, request.Strategy)

			operation := TargetResult{
				Provider: providerName,
				Name:     org,
				CloneDir: request.TargetPath,
				Strategy: request.Strategy,
				Success:  err == nil,
			}
			if err != nil {
				operation.Error = err.Error()
				result.FailedTargets++
			} else {
				result.SuccessfulTargets++
			}

			result.Results = append(result.Results, operation)
		}
	}

	// Execution completed
	return result, nil
}

// CloneByProvider clones repositories from a specific provider.
func (s *DefaultBulkOperationService) CloneByProvider(ctx context.Context, providerName string, request *BulkCloneRequest) (*BulkCloneResult, error) {
	// Create a modified request that only includes the specified provider
	modifiedRequest := *request
	modifiedRequest.Providers = []string{providerName}

	return s.CloneAll(ctx, &modifiedRequest)
}

// CloneByFilter clones repositories based on specified filter criteria.
func (s *DefaultBulkOperationService) CloneByFilter(ctx context.Context, filter RepositoryFilter, request *BulkCloneRequest) (*BulkCloneResult, error) {
	// Implementation would filter repositories based on the provided filter
	// For now, delegate to CloneAll
	return s.CloneAll(ctx, request)
}

// RefreshAll refreshes all repositories from configured providers.
func (s *DefaultBulkOperationService) RefreshAll(ctx context.Context, request *BulkRefreshRequest) (*BulkRefreshResult, error) {
	startTime := time.Now()
	result := &BulkRefreshResult{
		OperationResults: make([]RepositoryOperation, 0),
		ErrorSummary:     make(map[string]int),
	}

	// Process each configured provider
	for providerName, providerConfig := range s.config.Providers {
		config := ProviderConfig{Token: providerConfig.Token}

		provider, err := s.factory.CreateProvider(ctx, providerName, config)
		if err != nil {
			result.RefreshFailed++
			continue
		}

		// Process organizations for this provider
		for _, org := range request.Organizations {
			if request.DryRun {
				// In dry run mode, just validate the operation
				err := provider.IsHealthy(ctx)
				if err != nil {
					result.ErrorSummary[fmt.Sprintf("health_check_%s", providerName)]++
				}

				continue
			}

			err := provider.RefreshAll(ctx, request.TargetPath, org, request.Strategy)

			operation := RepositoryOperation{
				Organization: org,
				Provider:     providerName,
				Operation:    "refresh_all",
				Success:      err == nil,
				DurationMs:   time.Since(startTime).Milliseconds(),
				Path:         request.TargetPath,
			}
			if err != nil {
				operation.Error = err.Error()
				result.RefreshFailed++
				result.ErrorSummary[fmt.Sprintf("refresh_%s", providerName)]++
			} else {
				result.RefreshSuccessful++
			}

			result.OperationResults = append(result.OperationResults, operation)
		}
	}

	// Execution completed
	return result, nil
}

// RefreshByProvider refreshes repositories from a specific provider.
func (s *DefaultBulkOperationService) RefreshByProvider(ctx context.Context, providerName string, request *BulkRefreshRequest) (*BulkRefreshResult, error) {
	// Implementation would refresh only for the specified provider
	return s.RefreshAll(ctx, request)
}

// GetRepositoryStatus retrieves the status of repositories in the target path.
func (s *DefaultBulkOperationService) GetRepositoryStatus(ctx context.Context, targetPath string) (*RepositoryStatus, error) {
	startTime := time.Now()
	// Implementation would scan the target path and return status
	status := &RepositoryStatus{
		RepositoryDetails: make([]RepositoryStatusInfo, 0),
		ScanTimeMs:        time.Since(startTime).Milliseconds(),
	}

	return status, nil
}

// DiscoverRepositories discovers repositories from the specified providers.
func (s *DefaultBulkOperationService) DiscoverRepositories(ctx context.Context, providers []string) (*DiscoveryResult, error) {
	result := &DiscoveryResult{
		RepositoriesByProvider: make(map[string]int),
		Repositories:           make([]Repository, 0),
	}

	for _, providerName := range providers {
		providerConfig, exists := s.config.Providers[providerName]
		if !exists {
			continue
		}

		config := ProviderConfig{Token: providerConfig.Token}

		_, err := s.factory.CreateProvider(ctx, providerName, config)
		if err != nil {
			continue
		}

		// This would need to be enhanced to discover organizations/groups
		// For now, we'll skip the implementation
		result.RepositoriesByProvider[providerName] = 0
	}

	// Execution completed
	return result, nil
}

// Helper function to check if a slice contains a string.
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}

	return false
}
