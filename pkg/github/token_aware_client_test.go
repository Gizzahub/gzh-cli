package github_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gizzahub/gzh-manager-go/pkg/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenAwareGitHubClient_Creation(t *testing.T) {
	config := github.DefaultTokenAwareGitHubClientConfig()
	config.PrimaryToken = "test-token"
	config.FallbackTokens = []string{"fallback-1", "fallback-2"}

	client, err := github.NewTokenAwareGitHubClient(config)
	require.NoError(t, err)
	require.NotNil(t, client)

	defer client.Stop()

	// Test getting current token
	token, err := client.GetCurrentToken()
	require.NoError(t, err)
	assert.Equal(t, "test-token", token)
}

func TestTokenAwareGitHubClient_WithOAuth2(t *testing.T) {
	config := github.DefaultTokenAwareGitHubClientConfig()
	// OAuth2Config disabled - recovery package removed
	// config.OAuth2Config = &recovery.OAuth2Config{
	//	ClientID:     "test-client-id",
	//	ClientSecret: "test-client-secret",
	// }

	client, err := github.NewTokenAwareGitHubClient(config)
	require.NoError(t, err)
	require.NotNil(t, client)

	defer client.Stop()
}

func TestTokenAwareGitHubClient_APIOperations(t *testing.T) {
	// Create mock GitHub API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "token test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		switch r.URL.Path {
		case "/user":
			w.Header().Set("X-OAuth-Scopes", "repo,read:org")
			user := github.GitHubUser{
				ID:    12345,
				Login: "testuser",
				Name:  "Test User",
				Email: "test@example.com",
			}
			json.NewEncoder(w).Encode(user)

		case "/orgs/testorg":
			org := github.GitHubOrganization{
				ID:          67890,
				Login:       "testorg",
				Name:        "Test Organization",
				Description: "A test organization",
			}
			json.NewEncoder(w).Encode(org)

		case "/repos/testuser/testrepo":
			repo := github.GitHubRepository{
				ID:            111,
				Name:          "testrepo",
				FullName:      "testuser/testrepo",
				Description:   "A test repository",
				DefaultBranch: "main",
				Private:       false,
			}
			json.NewEncoder(w).Encode(repo)

		case "/user/repos":
			repos := []*github.GitHubRepository{
				{
					ID:            111,
					Name:          "repo1",
					FullName:      "testuser/repo1",
					DefaultBranch: "main",
				},
				{
					ID:            222,
					Name:          "repo2",
					FullName:      "testuser/repo2",
					DefaultBranch: "master",
				},
			}
			json.NewEncoder(w).Encode(repos)

		case "/rate_limit":
			rateLimit := map[string]interface{}{
				"resources": map[string]interface{}{
					"core": map[string]interface{}{
						"limit":     5000,
						"remaining": 4999,
						"reset":     time.Now().Add(1 * time.Hour).Unix(),
						"used":      1,
					},
				},
			}
			json.NewEncoder(w).Encode(rateLimit)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Create client with mock server
	config := github.DefaultTokenAwareGitHubClientConfig()
	config.BaseURL = server.URL
	config.PrimaryToken = "test-token"

	client, err := github.NewTokenAwareGitHubClient(config)
	require.NoError(t, err)
	defer client.Stop()

	ctx := context.Background()

	t.Run("GetUser", func(t *testing.T) {
		user, err := client.GetUser(ctx)
		require.NoError(t, err)
		assert.Equal(t, "testuser", user.Login)
		assert.Equal(t, "Test User", user.Name)
	})

	t.Run("GetOrganization", func(t *testing.T) {
		org, err := client.GetOrganization(ctx, "testorg")
		require.NoError(t, err)
		assert.Equal(t, "testorg", org.Login)
		assert.Equal(t, "Test Organization", org.Name)
	})

	t.Run("GetRepository", func(t *testing.T) {
		repo, err := client.GetRepository(ctx, "testuser", "testrepo")
		require.NoError(t, err)
		assert.Equal(t, "testrepo", repo.Name)
		assert.Equal(t, "main", repo.DefaultBranch)
	})

	t.Run("ListRepositories", func(t *testing.T) {
		repos, err := client.ListRepositories(ctx, "testuser", 1, 10)
		require.NoError(t, err)
		assert.Len(t, repos, 2)
		assert.Equal(t, "repo1", repos[0].Name)
		assert.Equal(t, "repo2", repos[1].Name)
	})

	t.Run("GetDefaultBranch", func(t *testing.T) {
		branch, err := client.GetDefaultBranch(ctx, "testuser", "testrepo")
		require.NoError(t, err)
		assert.Equal(t, "main", branch)
	})

	t.Run("GetRateLimit", func(t *testing.T) {
		rateLimit, err := client.GetRateLimit(ctx)
		require.NoError(t, err)
		assert.Equal(t, 5000, rateLimit.Limit)
		assert.Equal(t, 4999, rateLimit.Remaining)
	})

	t.Run("ValidateTokenPermissions", func(t *testing.T) {
		err := client.ValidateTokenPermissions(ctx, []string{"repo"})
		assert.NoError(t, err)

		err = client.ValidateTokenPermissions(ctx, []string{"admin:org"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing required scope")
	})
}

func TestTokenAwareGitHubClient_ErrorHandling(t *testing.T) {
	// Create mock server that returns various error responses
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/unauthorized":
			w.WriteHeader(http.StatusUnauthorized)
		case "/forbidden":
			w.Header().Set("X-RateLimit-Reset", "1640995200") // 2022-01-01 00:00:00 UTC
			w.WriteHeader(http.StatusForbidden)
		case "/notfound":
			w.WriteHeader(http.StatusNotFound)
		case "/unprocessable":
			w.WriteHeader(http.StatusUnprocessableEntity)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	config := github.DefaultTokenAwareGitHubClientConfig()
	config.BaseURL = server.URL
	config.PrimaryToken = "test-token"

	client, err := github.NewTokenAwareGitHubClient(config)
	require.NoError(t, err)
	defer client.Stop()

	ctx := context.Background()

	t.Run("Unauthorized", func(t *testing.T) {
		_, err := client.GetRepository(ctx, "user", "unauthorized")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unauthorized")
	})

	t.Run("Forbidden with rate limit", func(t *testing.T) {
		_, err := client.GetRepository(ctx, "user", "forbidden")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "forbidden")
		assert.Contains(t, err.Error(), "rate limit resets")
	})

	t.Run("Not found", func(t *testing.T) {
		_, err := client.GetRepository(ctx, "user", "notfound")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("Unprocessable entity", func(t *testing.T) {
		_, err := client.GetRepository(ctx, "user", "unprocessable")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unprocessable entity")
	})
}

func TestTokenAwareGitHubClient_TokenExpiration(t *testing.T) {
	// Mock server that simulates token expiration
	tokenExpired := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if tokenExpired {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// First request succeeds
		if r.URL.Path == "/user" {
			user := github.GitHubUser{
				ID:    12345,
				Login: "testuser",
			}
			json.NewEncoder(w).Encode(user)

			// Mark token as expired for next request
			tokenExpired = true
		}
	}))
	defer server.Close()

	config := github.DefaultTokenAwareGitHubClientConfig()
	config.BaseURL = server.URL
	config.PrimaryToken = "primary-token"
	config.FallbackTokens = []string{"fallback-token"}

	client, err := github.NewTokenAwareGitHubClient(config)
	require.NoError(t, err)
	defer client.Stop()

	ctx := context.Background()

	// First request should succeed
	user, err := client.GetUser(ctx)
	require.NoError(t, err)
	assert.Equal(t, "testuser", user.Login)

	// Second request should fail due to token expiration
	// In a real scenario, the token-aware client would attempt to refresh
	_, err = client.GetUser(ctx)
	assert.Error(t, err)
}

func TestTokenAwareGitHubClient_TokenStatus(t *testing.T) {
	config := github.DefaultTokenAwareGitHubClientConfig()
	config.PrimaryToken = "test-token"

	client, err := github.NewTokenAwareGitHubClient(config)
	require.NoError(t, err)
	defer client.Stop()

	status, err := client.GetTokenStatus()
	require.NoError(t, err)

	assert.Equal(t, "github", status.Service)
	assert.True(t, status.IsValid)
	assert.False(t, status.IsExpiring) // Should not be expiring for a fresh token
}

func TestDefaultTokenAwareGitHubClientConfig(t *testing.T) {
	config := github.DefaultTokenAwareGitHubClientConfig()

	assert.Equal(t, "https://api.github.com", config.BaseURL)
	assert.NotNil(t, config.HTTPConfig)
	assert.NotNil(t, config.ExpirationConfig)
	assert.Equal(t, 24*time.Hour, config.ExpirationConfig.ExpirationThreshold)
}
