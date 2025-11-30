// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

// Package auth provides common authentication utilities for Git platforms.
package auth

import (
	"fmt"

	"github.com/Gizzahub/gzh-cli/internal/env"
	"github.com/Gizzahub/gzh-cli/pkg/git/provider"
)

// TokenManager handles common token authentication logic across Git platforms.
type TokenManager struct {
	environment env.Environment
}

// NewTokenManager creates a new token manager.
func NewTokenManager(environment env.Environment) *TokenManager {
	return &TokenManager{
		environment: environment,
	}
}

// SetupTokenAuth sets up token authentication for a Git platform.
func (tm *TokenManager) SetupTokenAuth(token, platform string) (*provider.Credentials, error) {
	if token == "" {
		return nil, nil //nolint:nilnil // nil credentials with no error means no auth required
	}

	// Get the appropriate environment key for the platform
	var envKey string
	switch platform {
	case "github":
		envKey = env.CommonEnvironmentKeys.GitHubToken
	case "gitlab":
		envKey = env.CommonEnvironmentKeys.GitLabToken
	case "gitea":
		envKey = env.CommonEnvironmentKeys.GiteaToken
	default:
		return nil, fmt.Errorf("unsupported platform: %s", platform)
	}

	// Set token in environment
	if err := tm.environment.Set(envKey, token); err != nil {
		return nil, fmt.Errorf("failed to set %s token: %w", platform, err)
	}

	// Return credentials
	return &provider.Credentials{
		Type:  provider.CredentialTypeToken,
		Token: token,
	}, nil
}
