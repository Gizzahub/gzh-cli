// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package provider

import (
	"fmt"
)

// UnifiedProviderRegistry provides a centralized registration system for all Git providers.
type UnifiedProviderRegistry struct {
	factory *ProviderFactory
}

// NewUnifiedProviderRegistry creates a new unified provider registry.
func NewUnifiedProviderRegistry() *UnifiedProviderRegistry {
	return &UnifiedProviderRegistry{
		factory: NewProviderFactory(),
	}
}

// RegisterAllProviders registers all supported Git providers with the factory.
func (r *UnifiedProviderRegistry) RegisterAllProviders() error {
	// Register GitHub provider
	if err := r.factory.RegisterProvider("github", createGitHubProvider); err != nil {
		return fmt.Errorf("failed to register GitHub provider: %w", err)
	}

	// Register GitLab provider
	if err := r.factory.RegisterProvider("gitlab", createGitLabProvider); err != nil {
		return fmt.Errorf("failed to register GitLab provider: %w", err)
	}

	// Register Gitea provider
	if err := r.factory.RegisterProvider("gitea", createGiteaProvider); err != nil {
		return fmt.Errorf("failed to register Gitea provider: %w", err)
	}

	return nil
}

// GetFactory returns the underlying provider factory.
func (r *UnifiedProviderRegistry) GetFactory() *ProviderFactory {
	return r.factory
}

// GetRegistry creates a provider registry with the configured factory.
func (r *UnifiedProviderRegistry) GetRegistry(config RegistryConfig) *ProviderRegistry {
	return NewProviderRegistry(r.factory, config)
}

// Provider constructor functions (placeholders - these need to import actual packages)
// These would normally import the provider packages but we avoid circular imports

func createGitHubProvider(config *ProviderConfig) (GitProvider, error) {
	// This would call github.CreateGitHubProvider(config)
	// For now, return a placeholder error
	return nil, fmt.Errorf("GitHub provider constructor not connected - needs integration with pkg/github")
}

func createGitLabProvider(config *ProviderConfig) (GitProvider, error) {
	// This would call gitlab.CreateGitLabProvider(config)
	// For now, return a placeholder error
	return nil, fmt.Errorf("GitLab provider constructor not connected - needs integration with pkg/gitlab")
}

func createGiteaProvider(config *ProviderConfig) (GitProvider, error) {
	// This would call gitea.CreateGiteaProvider(config)
	// For now, return a placeholder error
	return nil, fmt.Errorf("Gitea provider constructor not connected - needs integration with pkg/gitea")
}

// DefaultUnifiedRegistry provides a default registry instance.
var DefaultUnifiedRegistry = NewUnifiedProviderRegistry()

// init registers all providers in the default registry.
func init() {
	if err := DefaultUnifiedRegistry.RegisterAllProviders(); err != nil {
		// Log error but don't panic during package initialization
		fmt.Printf("Warning: failed to register all providers: %v\n", err)
	}
}
