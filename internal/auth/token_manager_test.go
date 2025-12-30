// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gizzahub/gzh-cli/internal/env"
	"github.com/gizzahub/gzh-cli/pkg/git/provider"
)

func TestNewTokenManager(t *testing.T) {
	environment := env.NewOSEnvironment()
	tm := NewTokenManager(environment)

	assert.NotNil(t, tm)
	assert.Equal(t, environment, tm.environment)
}

func TestTokenManager_SetupTokenAuth_EmptyToken(t *testing.T) {
	environment := env.NewOSEnvironment()
	tm := NewTokenManager(environment)

	creds, err := tm.SetupTokenAuth("", "github")

	require.NoError(t, err)
	assert.Nil(t, creds)
}

func TestTokenManager_SetupTokenAuth_UnsupportedPlatform(t *testing.T) {
	environment := env.NewOSEnvironment()
	tm := NewTokenManager(environment)

	token := "test_token"
	creds, err := tm.SetupTokenAuth(token, "unsupported")

	require.Error(t, err)
	assert.Nil(t, creds)
	assert.Contains(t, err.Error(), "unsupported platform")
}

func TestTokenManager_SetupTokenAuth_ValidToken(t *testing.T) {
	environment := env.NewOSEnvironment()
	tm := NewTokenManager(environment)

	token := "test_token_123"

	// Test all supported platforms
	platforms := []string{"github", "gitlab", "gitea"}

	for _, platform := range platforms {
		creds, err := tm.SetupTokenAuth(token, platform)

		require.NoError(t, err, "platform: %s", platform)
		require.NotNil(t, creds, "platform: %s", platform)

		assert.Equal(t, provider.CredentialTypeToken, creds.Type, "platform: %s", platform)
		assert.Equal(t, token, creds.Token, "platform: %s", platform)
	}
}
