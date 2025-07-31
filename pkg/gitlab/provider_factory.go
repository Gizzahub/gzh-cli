// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package gitlab

import (
	"context"
	"fmt"
	"time"

	"github.com/gizzahub/gzh-manager-go/internal/auth"
	"github.com/gizzahub/gzh-manager-go/internal/env"
	"github.com/gizzahub/gzh-manager-go/pkg/git/provider"
)

// CreateGitLabProvider creates a new GitLab provider instance from configuration.
func CreateGitLabProvider(config *provider.ProviderConfig) (provider.GitProvider, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://gitlab.com/api/v4"
	}

	// Create the provider
	gitLabProvider := NewGitLabProvider(baseURL)

	// Set up token authentication using common token manager
	environment := env.NewOSEnvironment()
	tokenManager := auth.NewTokenManager(environment)
	credentials, err := tokenManager.SetupTokenAuth(config.Token, "gitlab")
	if err != nil {
		return nil, err
	}

	// Authenticate if credentials are available
	if credentials != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := gitLabProvider.Authenticate(ctx, *credentials); err != nil {
			return nil, fmt.Errorf("failed to authenticate GitLab provider: %w", err)
		}
	}

	return gitLabProvider, nil
}

// RegisterGitLabProvider registers the GitLab provider with a factory.
func RegisterGitLabProvider(factory *provider.ProviderFactory) error {
	return factory.RegisterProvider("gitlab", CreateGitLabProvider)
}
