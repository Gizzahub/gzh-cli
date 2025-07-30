// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package gitlab

import (
	"context"
	"fmt"
	"time"

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

	// Set up token authentication
	if config.Token != "" {
		// Set token in environment
		environment := env.NewOSEnvironment()
		if err := environment.Set(env.CommonEnvironmentKeys.GitLabToken, config.Token); err != nil {
			return nil, fmt.Errorf("failed to set GitLab token: %w", err)
		}

		// Authenticate
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		creds := provider.Credentials{
			Type:  provider.CredentialTypeToken,
			Token: config.Token,
		}

		if err := gitLabProvider.Authenticate(ctx, creds); err != nil {
			return nil, fmt.Errorf("failed to authenticate GitLab provider: %w", err)
		}
	}

	return gitLabProvider, nil
}

// RegisterGitLabProvider registers the GitLab provider with a factory.
func RegisterGitLabProvider(factory *provider.ProviderFactory) error {
	return factory.RegisterProvider("gitlab", CreateGitLabProvider)
}
