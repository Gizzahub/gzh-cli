package github_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gizzahub/gzh-cli/pkg/github"
)

const testToken = "test-token"

const (
	// userEndpoint is the GitHub API user endpoint.
	userEndpoint = "/user"
	// rateLimitEndpoint is the GitHub API rate limit endpoint.
	rateLimitEndpoint = "/rate_limit"
)

func TestTokenAwareGitHubClient_Creation(t *testing.T) {
	config := github.DefaultTokenAwareGitHubClientConfig()
	config.PrimaryToken = testToken
	config.FallbackTokens = []string{"fallback-1", "fallback-2"}

	client, err := github.NewTokenAwareGitHubClient(config)
	require.NoError(t, err)
	require.NotNil(t, client)

	defer client.Stop()

	// Test getting current token
	token, err := client.GetCurrentToken()
	require.NoError(t, err)
	assert.Equal(t, testToken, token)
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
		case userEndpoint:
			w.Header().Set("X-OAuth-Scopes", "repo,read:org")

			user := github.GitHubUser{
				ID:    12345,
				Login: "testuser",
				Name:  "Test User",
				Email: "test@example.com",
			}
			if err := json.NewEncoder(w).Encode(user); err != nil {
				http.Error(w, "Failed to encode user response", http.StatusInternalServerError)
			}

		case "/orgs/testorg":
			org := github.GitHubOrganization{
				ID:          67890,
				Login:       "testorg",
				Name:        "Test Organization",
				Description: "A test organization",
			}
			if err := json.NewEncoder(w).Encode(org); err != nil {
				http.Error(w, "Failed to encode org response", http.StatusInternalServerError)
			}

		case "/repos/testuser/testrepo":
			repo := github.GitHubRepository{
				ID:            111,
				Name:          "testrepo",
				FullName:      "testuser/testrepo",
				Description:   "A test repository",
				DefaultBranch: "main",
				Private:       false,
			}
			if err := json.NewEncoder(w).Encode(repo); err != nil {
				http.Error(w, "Failed to encode repo response", http.StatusInternalServerError)
			}

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
			if err := json.NewEncoder(w).Encode(repos); err != nil {
				http.Error(w, "Failed to encode repos response", http.StatusInternalServerError)
			}

		case rateLimitEndpoint:
			rateLimit := map[string]any{
				"resources": map[string]any{
					"core": map[string]any{
						"limit":     5000,
						"remaining": 4999,
						"reset":     time.Now().Add(1 * time.Hour).Unix(),
						"used":      1,
					},
				},
			}
			if err := json.NewEncoder(w).Encode(rateLimit); err != nil {
				http.Error(w, "Failed to encode rate limit response", http.StatusInternalServerError)
			}

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Create client with mock server
	config := github.DefaultTokenAwareGitHubClientConfig()
	config.BaseURL = server.URL
	config.PrimaryToken = testToken

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
		// 경로는 GetRepository가 실제로 부르는 /repos/{owner}/{repo}다.
		// 예전에는 /unauthorized 같은 짧은 경로로 갈라서 어느 갈래에도
		// 걸리지 않았고, 전부 default의 500으로 떨어졌다. 그런데도 "unauthorized"
		// 하위 시험은 통과했다 -- 오류 문구에 저장소 이름이 그대로 들어가서
		// 500 메시지가 "unauthorized"를 포함해 버렸기 때문이다.
		switch r.URL.Path {
		case "/repos/user/unauthorized":
			w.WriteHeader(http.StatusUnauthorized)
		case "/repos/user/forbidden":
			w.Header().Set("X-RateLimit-Reset", "1640995200") // 2022-01-01 00:00:00 UTC
			w.WriteHeader(http.StatusForbidden)
		case "/repos/user/notfound":
			w.WriteHeader(http.StatusNotFound)
		case "/repos/user/unprocessable":
			w.WriteHeader(http.StatusUnprocessableEntity)
		default:
			t.Errorf("가짜 서버가 모르는 요청을 받았다: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	config := github.DefaultTokenAwareGitHubClientConfig()
	config.BaseURL = server.URL
	config.PrimaryToken = testToken

	client, err := github.NewTokenAwareGitHubClient(config)
	require.NoError(t, err)

	defer client.Stop()

	ctx := context.Background()

	t.Run("Unauthorized", func(t *testing.T) {
		_, err := client.GetRepository(ctx, "user", "unauthorized")
		require.Error(t, err)
		// 저장소 이름에 이미 "unauthorized"가 들어 있으므로 그 낱말만으로는
		// 아무것도 확인하지 못한다. 401 갈래에만 있는 문구를 본다.
		assert.Contains(t, err.Error(), "token may be expired or invalid")
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
		if r.URL.Path == userEndpoint {
			user := github.GitHubUser{
				ID:    12345,
				Login: "testuser",
			}
			_ = json.NewEncoder(w).Encode(user) // Ignore encode error

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
	config.PrimaryToken = testToken

	client, err := github.NewTokenAwareGitHubClient(config)
	require.NoError(t, err)

	defer client.Stop()

	status, err := client.GetTokenStatus()
	require.NoError(t, err)

	// GetTokenStatus returns a map[string]interface{}
	hasToken, ok := status["has_token"].(bool)
	assert.True(t, ok, "has_token should be a bool")
	assert.True(t, hasToken)
	assert.Equal(t, "recovery package removed, using simple token management", status["note"])
}

func TestDefaultTokenAwareGitHubClientConfig(t *testing.T) {
	config := github.DefaultTokenAwareGitHubClientConfig()

	assert.Equal(t, github.DefaultGitHubAPIBaseURL, config.BaseURL)
	assert.Equal(t, 30*time.Second, config.Timeout)
	// HTTPConfig and ExpirationConfig were removed when recovery package was removed
}

func TestNewTokenAwareGitHubClient_EmptyBaseURL(t *testing.T) {
	config := github.TokenAwareGitHubClientConfig{
		PrimaryToken: "token",
		// BaseURL intentionally empty
	}

	client, err := github.NewTokenAwareGitHubClient(config)
	require.NoError(t, err)
	require.NotNil(t, client)

	// Black-box: exercise via GetStats or similar if available; use Start/Stop smoke.
	// BaseURL resolution is covered via Default config + Resolve unit tests;
	// empty BaseURL must not prevent construction.
	assert.NoError(t, client.Start(context.Background()))
	client.Stop()
}
