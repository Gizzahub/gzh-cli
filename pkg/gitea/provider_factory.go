// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package gitea

import (
	"context"
	"fmt"
	"time"

	"github.com/gizzahub/gzh-manager-go/internal/env"
	"github.com/gizzahub/gzh-manager-go/pkg/git/provider"
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

	// Set up token authentication
	if config.Token != "" {
		// Set token in environment
		environment := env.NewOSEnvironment()
		if err := environment.Set(env.CommonEnvironmentKeys.GiteaToken, config.Token); err != nil {
			return nil, fmt.Errorf("failed to set Gitea token: %w", err)
		}

		// Authenticate
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		creds := provider.Credentials{
			Type:  provider.CredentialTypeToken,
			Token: config.Token,
		}

		if err := giteaProvider.Authenticate(ctx, creds); err != nil {
			return nil, fmt.Errorf("failed to authenticate Gitea provider: %w", err)
		}
	}

	return giteaProvider, nil
}

// RegisterGiteaProvider registers the Gitea provider with a factory.
func RegisterGiteaProvider(factory *provider.ProviderFactory) error {
	return factory.RegisterProvider("gitea", CreateGiteaProvider)
}
