// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package gitea

import (
	"context"
	"fmt"
	"time"

	"github.com/Gizzahub/gzh-cli/internal/auth"
	"github.com/Gizzahub/gzh-cli/internal/env"
	"github.com/Gizzahub/gzh-cli/pkg/git/provider"
)

// CreateGiteaProvider creates a new Gitea provider instance from configuration.
func CreateGiteaProvider(config *provider.ProviderConfig) (provider.GitProvider, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://gitea.com/api/v1"
	}

	// Create the provider
	giteaProvider := NewGiteaProvider(baseURL)

	// Set up token authentication using common token manager
	environment := env.NewOSEnvironment()
	tokenManager := auth.NewTokenManager(environment)
	credentials, err := tokenManager.SetupTokenAuth(config.Token, "gitea")
	if err != nil {
		return nil, err
	}

	// Authenticate if credentials are available
	if credentials != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := giteaProvider.Authenticate(ctx, *credentials); err != nil {
			return nil, fmt.Errorf("failed to authenticate Gitea provider: %w", err)
		}
	}

	return giteaProvider, nil
}

// RegisterGiteaProvider registers the Gitea provider with a factory.
func RegisterGiteaProvider(factory *provider.ProviderFactory) error {
	return factory.RegisterProvider("gitea", CreateGiteaProvider)
}
