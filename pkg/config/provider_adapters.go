// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package config

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Gizzahub/gzh-cli/internal/env"
)

// GitHubProviderAdapter adapts the github package to implement ProviderService.
type GitHubProviderAdapter struct {
	*BaseProviderAdapter
}

// NewGitHubProviderAdapter creates a new GitHub provider adapter.
func NewGitHubProviderAdapter(token string, environment env.Environment) *GitHubProviderAdapter {
	config := ProviderAdapterConfig{
		Name:        ProviderGitHub,
		BaseURL:     "github.com",
		APIURL:      "https://api.github.com",
		TokenEnvKey: env.CommonEnvironmentKeys.GitHubToken,
	}
	
	base := NewBaseProviderAdapter(token, environment, NewGitHubAPI(), config)
	return &GitHubProviderAdapter{BaseProviderAdapter: base}
}

// GitLabProviderAdapter adapts the gitlab package to implement ProviderService.
type GitLabProviderAdapter struct {
	*BaseProviderAdapter
}

// NewGitLabProviderAdapter creates a new GitLab provider adapter.
func NewGitLabProviderAdapter(token string, environment env.Environment) *GitLabProviderAdapter {
	config := ProviderAdapterConfig{
		Name:        ProviderGitLab,
		BaseURL:     "gitlab.com",
		APIURL:      "https://gitlab.com/api/v4",
		TokenEnvKey: env.CommonEnvironmentKeys.GitLabToken,
	}
	
	base := NewBaseProviderAdapter(token, environment, NewGitLabAPI(), config)
	return &GitLabProviderAdapter{BaseProviderAdapter: base}
}

// GiteaProviderAdapter adapts the gitea package to implement ProviderService.
type GiteaProviderAdapter struct {
	*BaseProviderAdapter
}

// NewGiteaProviderAdapter creates a new Gitea provider adapter.
func NewGiteaProviderAdapter(token string, environment env.Environment) *GiteaProviderAdapter {
	config := ProviderAdapterConfig{
		Name:        ProviderGitea,
		BaseURL:     "gitea.com",
		APIURL:      "https://gitea.com/api/v1",
		TokenEnvKey: env.CommonEnvironmentKeys.GiteaToken,
	}
	
	base := NewBaseProviderAdapter(token, environment, NewGiteaAPI(), config)
	return &GiteaProviderAdapter{BaseProviderAdapter: base}
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
func (f *DefaultProviderFactory) CreateProvider(_ context.Context, providerName string, config ProviderConfig) (ProviderService, error) {
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
func (s *DefaultBulkOperationService) CloneByFilter(ctx context.Context, _ RepositoryFilter, request *BulkCloneRequest) (*BulkCloneResult, error) {
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
func (s *DefaultBulkOperationService) RefreshByProvider(ctx context.Context, _ string, request *BulkRefreshRequest) (*BulkRefreshResult, error) {
	// Implementation would refresh only for the specified provider
	return s.RefreshAll(ctx, request)
}

// GetRepositoryStatus retrieves the status of repositories in the target path.
func (s *DefaultBulkOperationService) GetRepositoryStatus(_ context.Context, _ string) (*RepositoryStatus, error) {
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