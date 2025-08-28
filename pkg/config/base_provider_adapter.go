// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package config

import (
	"context"
	"fmt"
	"strings"

	"github.com/Gizzahub/gzh-cli/internal/env"
)

const (
	defaultMainBranch = "main"
)

// ProviderAPI defines the interface that each provider implementation must fulfill.
type ProviderAPI interface {
	List(ctx context.Context, owner string) ([]string, error)
	GetDefaultBranch(ctx context.Context, owner, repository string) (string, error)
	Clone(ctx context.Context, targetPath, owner, repository string, extraParams ...string) error
	RefreshAll(ctx context.Context, targetPath, owner string, extraParams ...string) error
}

// ProviderAdapterConfig holds provider-specific configuration for adapters.
type ProviderAdapterConfig struct {
	Name        string
	BaseURL     string
	APIURL      string
	TokenEnvKey string
}

// BaseProviderAdapter provides common functionality for all provider adapters.
type BaseProviderAdapter struct {
	token       string
	environment env.Environment
	api         ProviderAPI
	config      ProviderAdapterConfig
}

// NewBaseProviderAdapter creates a new base provider adapter.
func NewBaseProviderAdapter(token string, environment env.Environment, api ProviderAPI, config ProviderAdapterConfig) *BaseProviderAdapter {
	return &BaseProviderAdapter{
		token:       token,
		environment: environment,
		api:         api,
		config:      config,
	}
}

// ListRepositories lists all repositories for a given owner.
func (b *BaseProviderAdapter) ListRepositories(ctx context.Context, owner string) ([]Repository, error) {
	repoNames, err := b.api.List(ctx, owner)
	if err != nil {
		return nil, err
	}

	repositories := make([]Repository, 0, len(repoNames))
	for _, name := range repoNames {
		// Get additional repository information
		defaultBranch, err := b.api.GetDefaultBranch(ctx, owner, name)
		if err != nil {
			defaultBranch = defaultMainBranch // fallback to main if error
		}

		repo := Repository{
			Name:          name,
			FullName:      fmt.Sprintf("%s/%s", owner, name),
			DefaultBranch: defaultBranch,
			CloneURL:      fmt.Sprintf("https://%s/%s/%s.git", b.config.BaseURL, owner, name),
			SSHURL:        fmt.Sprintf("git@%s:%s/%s.git", b.config.BaseURL, owner, name),
			HTMLURL:       fmt.Sprintf("https://%s/%s/%s", b.config.BaseURL, owner, name),
		}
		repositories = append(repositories, repo)
	}

	return repositories, nil
}

// CloneRepository clones a single repository to the target path.
func (b *BaseProviderAdapter) CloneRepository(ctx context.Context, owner, repository, targetPath string) error {
	return b.api.Clone(ctx, targetPath, owner, repository, "")
}

// GetDefaultBranch retrieves the default branch name for a repository.
func (b *BaseProviderAdapter) GetDefaultBranch(ctx context.Context, owner, repository string) (string, error) {
	return b.api.GetDefaultBranch(ctx, owner, repository)
}

// RefreshAll updates all repositories in the target path using the specified strategy.
func (b *BaseProviderAdapter) RefreshAll(ctx context.Context, targetPath, owner, strategy string) error {
	return b.api.RefreshAll(ctx, targetPath, owner, strategy)
}

// CloneOrganization clones all repositories from an organization.
func (b *BaseProviderAdapter) CloneOrganization(ctx context.Context, owner, targetPath, strategy string) error {
	return b.api.RefreshAll(ctx, targetPath, owner, strategy)
}

// SetToken configures the authentication token.
func (b *BaseProviderAdapter) SetToken(ctx context.Context, token string) error {
	b.token = token
	if b.token != "" && !strings.HasPrefix(b.token, "$") {
		if err := b.environment.Set(b.config.TokenEnvKey, b.token); err != nil {
			return fmt.Errorf("failed to set %s token environment variable: %w", b.config.Name, err)
		}
	}
	return nil
}

// ValidateToken verifies that the configured token is valid.
func (b *BaseProviderAdapter) ValidateToken(ctx context.Context) error {
	// Simple validation - try to make an API call
	testOwner := b.getTestOwner()
	_, err := b.api.List(ctx, testOwner)
	return err
}

// GetProviderName returns the name identifier for this provider.
func (b *BaseProviderAdapter) GetProviderName() string {
	return b.config.Name
}

// GetAPIEndpoint returns the base URL for the provider API.
func (b *BaseProviderAdapter) GetAPIEndpoint() string {
	return b.config.APIURL
}

// IsHealthy checks if the provider is healthy and accessible.
func (b *BaseProviderAdapter) IsHealthy(ctx context.Context) error {
	return b.ValidateToken(ctx)
}

// getTestOwner returns a test owner for validation.
func (b *BaseProviderAdapter) getTestOwner() string {
	switch strings.ToLower(b.config.Name) {
	case ProviderGitHub:
		return ProviderGitHub
	case ProviderGitLab:
		return "gitlab-org"
	case ProviderGitea:
		return ProviderGitea
	default:
		return "test"
	}
}
