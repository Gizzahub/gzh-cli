package recovery_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gizzahub/gzh-manager-go/pkg/recovery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitHubTokenRefresher(t *testing.T) {
	// Create mock OAuth2 server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/login/oauth/access_token" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Parse form data
		err := r.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		clientID := r.FormValue("client_id")
		clientSecret := r.FormValue("client_secret")
		grantType := r.FormValue("grant_type")
		refreshToken := r.FormValue("refresh_token")

		// Validate request
		if clientID != "test-client-id" || clientSecret != "test-client-secret" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error":             "invalid_client",
				"error_description": "Invalid client credentials",
			})
			return
		}

		if grantType != "refresh_token" || refreshToken != "valid-refresh-token" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error":             "invalid_grant",
				"error_description": "Invalid refresh token",
			})
			return
		}

		// Return successful refresh response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token":  "new-access-token",
			"refresh_token": "new-refresh-token",
			"token_type":    "bearer",
			"scope":         "repo,read:org",
			"expires_in":    3600,
		})
	}))
	defer server.Close()

	// Mock the GitHub OAuth2 endpoint
	originalURL := "https://github.com/login/oauth/access_token"
	refresher := recovery.NewGitHubTokenRefresher("test-client-id", "test-client-secret")

	// We need to modify the refresher to use our test server
	// Since the URL is hardcoded, we'll test the interface behavior

	t.Run("CanRefresh", func(t *testing.T) {
		tests := []struct {
			name      string
			tokenInfo *recovery.TokenInfo
			expected  bool
		}{
			{
				name: "OAuth2 token with refresh token",
				tokenInfo: &recovery.TokenInfo{
					TokenType: "oauth2",
					Metadata: map[string]interface{}{
						"refresh_token": "valid-refresh-token",
					},
				},
				expected: true,
			},
			{
				name: "Classic token",
				tokenInfo: &recovery.TokenInfo{
					TokenType: "classic",
				},
				expected: false,
			},
			{
				name: "OAuth2 token without refresh token",
				tokenInfo: &recovery.TokenInfo{
					TokenType: "oauth2",
					Metadata:  map[string]interface{}{},
				},
				expected: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := refresher.CanRefresh(tt.tokenInfo)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("RefreshToken without refresh token", func(t *testing.T) {
		tokenInfo := &recovery.TokenInfo{
			TokenType: "oauth2",
			Metadata:  map[string]interface{}{},
		}

		_, err := refresher.RefreshToken(context.Background(), tokenInfo)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no refresh token")
	})

	t.Run("RefreshToken with invalid refresh token type", func(t *testing.T) {
		tokenInfo := &recovery.TokenInfo{
			TokenType: "oauth2",
			Metadata: map[string]interface{}{
				"refresh_token": 123, // Invalid type
			},
		}

		_, err := refresher.RefreshToken(context.Background(), tokenInfo)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid refresh token format")
	})
}

func TestGitLabTokenRefresher(t *testing.T) {
	// Create mock GitLab OAuth2 server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oauth/token" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Parse form data
		err := r.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		clientID := r.FormValue("client_id")
		clientSecret := r.FormValue("client_secret")
		grantType := r.FormValue("grant_type")
		refreshToken := r.FormValue("refresh_token")

		// Validate request
		if clientID != "gitlab-client-id" || clientSecret != "gitlab-client-secret" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error":             "invalid_client",
				"error_description": "Invalid client credentials",
			})
			return
		}

		if grantType != "refresh_token" || refreshToken != "gitlab-refresh-token" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error":             "invalid_grant",
				"error_description": "Invalid refresh token",
			})
			return
		}

		// Return successful refresh response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token":  "new-gitlab-token",
			"refresh_token": "new-gitlab-refresh-token",
			"token_type":    "Bearer",
			"scope":         "read_user read_repository",
			"expires_in":    7200,
		})
	}))
	defer server.Close()

	refresher := recovery.NewGitLabTokenRefresher("gitlab-client-id", "gitlab-client-secret", server.URL)

	t.Run("CanRefresh", func(t *testing.T) {
		tokenInfo := &recovery.TokenInfo{
			TokenType: "oauth2",
			Metadata: map[string]interface{}{
				"refresh_token": "gitlab-refresh-token",
			},
		}

		result := refresher.CanRefresh(tokenInfo)
		assert.True(t, result)
	})

	t.Run("RefreshToken success", func(t *testing.T) {
		tokenInfo := &recovery.TokenInfo{
			Service:   "gitlab",
			TokenType: "oauth2",
			Metadata: map[string]interface{}{
				"refresh_token": "gitlab-refresh-token",
			},
		}

		newTokenInfo, err := refresher.RefreshToken(context.Background(), tokenInfo)
		require.NoError(t, err)

		assert.Equal(t, "new-gitlab-token", newTokenInfo.Token)
		assert.Equal(t, "gitlab", newTokenInfo.Service)
		assert.Equal(t, "oauth2", newTokenInfo.TokenType)
		assert.Equal(t, "new-gitlab-refresh-token", newTokenInfo.Metadata["refresh_token"])
		assert.NotNil(t, newTokenInfo.ExpiresAt)
		assert.Contains(t, newTokenInfo.Scopes, "read_user")
		assert.Contains(t, newTokenInfo.Scopes, "read_repository")
	})
}

func TestPersonalAccessTokenRefresher(t *testing.T) {
	mockValidator := &MockTokenValidator{}
	fallbackTokens := []string{"fallback-1", "fallback-2", "fallback-3"}

	refresher := recovery.NewPersonalAccessTokenRefresher(fallbackTokens, mockValidator)

	t.Run("CanRefresh with fallback tokens", func(t *testing.T) {
		tokenInfo := &recovery.TokenInfo{Token: "current-token"}
		result := refresher.CanRefresh(tokenInfo)
		assert.True(t, result)
	})

	t.Run("CanRefresh without fallback tokens", func(t *testing.T) {
		emptyRefresher := recovery.NewPersonalAccessTokenRefresher([]string{}, mockValidator)
		tokenInfo := &recovery.TokenInfo{Token: "current-token"}
		result := emptyRefresher.CanRefresh(tokenInfo)
		assert.False(t, result)
	})

	t.Run("RefreshToken success", func(t *testing.T) {
		tokenInfo := &recovery.TokenInfo{
			Token:   "expired-token",
			Service: "github",
		}

		newTokenInfo, err := refresher.RefreshToken(context.Background(), tokenInfo)
		require.NoError(t, err)

		assert.Equal(t, "fallback-1", newTokenInfo.Token)
		assert.Equal(t, "github", newTokenInfo.Service)
	})

	t.Run("RefreshToken skips current token", func(t *testing.T) {
		tokenInfo := &recovery.TokenInfo{
			Token:   "fallback-1", // This is in our fallback list
			Service: "github",
		}

		newTokenInfo, err := refresher.RefreshToken(context.Background(), tokenInfo)
		require.NoError(t, err)

		// Should get fallback-2 since fallback-1 is the current token
		assert.Equal(t, "fallback-2", newTokenInfo.Token)
	})
}

func TestTokenRefreshRegistry(t *testing.T) {
	registry := recovery.NewTokenRefreshRegistry()
	mockValidator := &MockTokenValidator{}

	// Add refreshers
	fallbackRefresher := recovery.NewPersonalAccessTokenRefresher([]string{"fallback"}, mockValidator)
	registry.AddRefresher("github", fallbackRefresher)

	t.Run("GetRefresher for existing service", func(t *testing.T) {
		tokenInfo := &recovery.TokenInfo{
			Service: "github",
			Token:   "current-token",
		}

		refresher := registry.GetRefresher("github", tokenInfo)
		assert.NotNil(t, refresher)
		assert.True(t, refresher.CanRefresh(tokenInfo))
	})

	t.Run("GetRefresher for non-existent service", func(t *testing.T) {
		tokenInfo := &recovery.TokenInfo{
			Service: "gitlab",
			Token:   "current-token",
		}

		refresher := registry.GetRefresher("gitlab", tokenInfo)
		assert.Nil(t, refresher)
	})

	t.Run("RefreshToken success", func(t *testing.T) {
		tokenInfo := &recovery.TokenInfo{
			Service: "github",
			Token:   "current-token",
		}

		newTokenInfo, err := registry.RefreshToken(context.Background(), tokenInfo)
		require.NoError(t, err)
		assert.Equal(t, "fallback", newTokenInfo.Token)
	})

	t.Run("RefreshToken no suitable refresher", func(t *testing.T) {
		tokenInfo := &recovery.TokenInfo{
			Service: "unknown",
			Token:   "current-token",
		}

		_, err := registry.RefreshToken(context.Background(), tokenInfo)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no suitable refresher")
	})
}

func TestTokenRefreshScheduler(t *testing.T) {
	// Create mock token manager
	mockValidator := &MockTokenValidator{returnExpiring: true}
	config := recovery.DefaultTokenManagerConfig()
	config.Validator = mockValidator
	config.ExpirationThreshold = 2 * time.Hour
	config.CheckInterval = 50 * time.Millisecond

	manager := recovery.NewTokenManager(config)

	// Create registry with fallback refresher
	registry := recovery.NewTokenRefreshRegistry()
	fallbackRefresher := recovery.NewPersonalAccessTokenRefresher([]string{"new-token"}, mockValidator)
	registry.AddRefresher("github", fallbackRefresher)

	// Create scheduler
	policy := recovery.DefaultTokenRefreshPolicy()
	policy.RefreshThreshold = 2 * time.Hour // Same as manager threshold
	scheduler := recovery.NewTokenRefreshScheduler(manager, registry, policy)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	// Start services
	err := manager.Start(ctx)
	require.NoError(t, err)
	defer manager.Stop()

	err = scheduler.Start(ctx)
	require.NoError(t, err)
	defer scheduler.Stop()

	// Add a token that will be considered expiring
	err = manager.AddToken("github", "expiring-token")
	require.NoError(t, err)

	// Wait for potential refresh attempts
	time.Sleep(150 * time.Millisecond)

	// Verify the scheduler ran without errors
	// (More detailed testing would require exposing internal state)
	status := manager.GetTokenStatus()
	assert.Contains(t, status, "github")
}

func TestDefaultTokenRefreshPolicy(t *testing.T) {
	policy := recovery.DefaultTokenRefreshPolicy()

	assert.Equal(t, recovery.RefreshStrategyProactive, policy.Strategy)
	assert.Equal(t, 24*time.Hour, policy.RefreshThreshold)
	assert.Equal(t, 3, policy.MaxRefreshAttempts)
	assert.Equal(t, 5*time.Minute, policy.RefreshRetryDelay)
	assert.True(t, policy.NotifyOnRefresh)
	assert.True(t, policy.AutoRetryOnFailure)
}
