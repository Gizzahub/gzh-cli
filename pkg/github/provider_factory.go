// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package github

import (
	"context"
	"fmt"
	"time"

	"github.com/gizzahub/gzh-manager-go/internal/auth"
	"github.com/gizzahub/gzh-manager-go/internal/env"
	"github.com/gizzahub/gzh-manager-go/pkg/git/provider"
)

// CreateGitHubProvider creates a new GitHub provider instance from configuration.
func CreateGitHubProvider(config *provider.ProviderConfig) (provider.GitProvider, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Create environment for token management
	environment := env.NewOSEnvironment()

	// Set up token authentication using common token manager
	tokenManager := auth.NewTokenManager(environment)
	credentials, err := tokenManager.SetupTokenAuth(config.Token, "github")
	if err != nil {
		return nil, err
	}

	// Create resilient client
	resilientClient := NewResilientGitHubClientWithConfig(config.Token, time.Duration(config.Timeout)*time.Second)

	// Create adapter to bridge different interfaces
	apiClientAdapter := &GitHubAPIClientAdapter{client: resilientClient}

	// Create a simple clone service
	cloneService := &SimpleCloneService{}

	// Create the provider
	gitHubProvider := NewGitHubProvider(apiClientAdapter, cloneService)

	// Authenticate if credentials are available
	if credentials != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := gitHubProvider.Authenticate(ctx, *credentials); err != nil {
			return nil, fmt.Errorf("failed to authenticate GitHub provider: %w", err)
		}
	}

	return gitHubProvider, nil
}

// SimpleCloneService provides a minimal implementation of CloneService
type SimpleCloneService struct{}

// Ensure SimpleCloneService implements CloneService interface
var _ CloneService = (*SimpleCloneService)(nil)

func (s *SimpleCloneService) CloneRepository(ctx context.Context, repo RepositoryInfo, targetPath, strategy string) error {
	return Clone(ctx, targetPath, extractOwner(repo.FullName), repo.Name)
}

func (s *SimpleCloneService) RefreshAll(ctx context.Context, targetPath, orgName, strategy string) error {
	return RefreshAll(ctx, targetPath, orgName, strategy)
}

func (s *SimpleCloneService) CloneOrganization(ctx context.Context, orgName, targetPath, strategy string) error {
	return RefreshAll(ctx, targetPath, orgName, strategy)
}

func (s *SimpleCloneService) SetStrategy(ctx context.Context, strategy string) error {
	// Strategy is handled per-operation, no global state needed
	return nil
}

func (s *SimpleCloneService) GetSupportedStrategies(ctx context.Context) ([]string, error) {
	return []string{"reset", "pull", "fetch"}, nil
}

// extractOwner extracts owner from "owner/repo" format
func extractOwner(fullName string) string {
	for i, char := range fullName {
		if char == '/' {
			return fullName[:i]
		}
	}
	return fullName
}

// GitHubAPIClientAdapter adapts ResilientGitHubClient to APIClient interface
type GitHubAPIClientAdapter struct {
	client *ResilientGitHubClient
}

// Ensure GitHubAPIClientAdapter implements APIClient interface
var _ APIClient = (*GitHubAPIClientAdapter)(nil)

func (a *GitHubAPIClientAdapter) GetRepository(ctx context.Context, owner, repo string) (*RepositoryInfo, error) {
	// ResilientGitHubClient doesn't have GetRepository, use global function
	repos, err := List(ctx, owner)
	if err != nil {
		return nil, err
	}

	// Find the specific repository
	for _, repoName := range repos {
		if repoName == repo {
			defaultBranch, _ := a.client.GetDefaultBranch(ctx, owner, repo)
			return &RepositoryInfo{
				Name:          repo,
				FullName:      fmt.Sprintf("%s/%s", owner, repo),
				DefaultBranch: defaultBranch,
				CloneURL:      fmt.Sprintf("https://github.com/%s/%s.git", owner, repo),
				SSHURL:        fmt.Sprintf("git@github.com:%s/%s.git", owner, repo),
				HTMLURL:       fmt.Sprintf("https://github.com/%s/%s", owner, repo),
			}, nil
		}
	}

	return nil, fmt.Errorf("repository %s/%s not found", owner, repo)
}

func (a *GitHubAPIClientAdapter) ListOrganizationRepositories(ctx context.Context, org string) ([]RepositoryInfo, error) {
	repoNames, err := a.client.ListRepositories(ctx, org)
	if err != nil {
		return nil, err
	}

	repos := make([]RepositoryInfo, 0, len(repoNames))
	for _, name := range repoNames {
		defaultBranch, _ := a.client.GetDefaultBranch(ctx, org, name)
		repo := RepositoryInfo{
			Name:          name,
			FullName:      fmt.Sprintf("%s/%s", org, name),
			DefaultBranch: defaultBranch,
			CloneURL:      fmt.Sprintf("https://github.com/%s/%s.git", org, name),
			SSHURL:        fmt.Sprintf("git@github.com:%s/%s.git", org, name),
			HTMLURL:       fmt.Sprintf("https://github.com/%s/%s", org, name),
		}
		repos = append(repos, repo)
	}

	return repos, nil
}

func (a *GitHubAPIClientAdapter) GetDefaultBranch(ctx context.Context, owner, repo string) (string, error) {
	return a.client.GetDefaultBranch(ctx, owner, repo)
}

func (a *GitHubAPIClientAdapter) SetToken(ctx context.Context, token string) error {
	// ResilientGitHubClient doesn't have SetToken method, token is set during construction
	return nil
}

func (a *GitHubAPIClientAdapter) GetRateLimit(ctx context.Context) (*RateLimit, error) {
	rateLimitInfo, err := a.client.GetRateLimit(ctx)
	if err != nil {
		return nil, err
	}

	// Convert RateLimitInfo to RateLimit
	return &RateLimit{
		Limit:     rateLimitInfo.Limit,
		Remaining: rateLimitInfo.Remaining,
		Reset:     rateLimitInfo.ResetTime,
		Used:      rateLimitInfo.Limit - rateLimitInfo.Remaining,
	}, nil
}

func (a *GitHubAPIClientAdapter) GetRepositoryConfiguration(ctx context.Context, owner, repo string) (*RepositoryConfig, error) {
	// This would need to be implemented based on the actual requirements
	return nil, fmt.Errorf("GetRepositoryConfiguration not implemented")
}

func (a *GitHubAPIClientAdapter) UpdateRepositoryConfiguration(ctx context.Context, owner, repo string, config *RepositoryConfig) error {
	// This would need to be implemented based on the actual requirements
	return fmt.Errorf("UpdateRepositoryConfiguration not implemented")
}

// RegisterGitHubProvider registers the GitHub provider with a factory.
func RegisterGitHubProvider(factory *provider.ProviderFactory) error {
	return factory.RegisterProvider("github", CreateGitHubProvider)
}
