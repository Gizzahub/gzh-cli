package recovery_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gizzahub/gzh-manager-go/pkg/recovery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenInfo_IsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt *time.Time
		threshold time.Duration
		expected  bool
	}{
		{
			name:      "no expiration",
			expiresAt: nil,
			threshold: time.Hour,
			expected:  false,
		},
		{
			name:      "expires in future beyond threshold",
			expiresAt: timePtr(time.Now().Add(2 * time.Hour)),
			threshold: time.Hour,
			expected:  false,
		},
		{
			name:      "expires within threshold",
			expiresAt: timePtr(time.Now().Add(30 * time.Minute)),
			threshold: time.Hour,
			expected:  true,
		},
		{
			name:      "already expired",
			expiresAt: timePtr(time.Now().Add(-1 * time.Hour)),
			threshold: time.Hour,
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenInfo := &recovery.TokenInfo{
				ExpiresAt: tt.expiresAt,
			}

			result := tokenInfo.IsExpired(tt.threshold)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTokenInfo_IsValid(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt *time.Time
		expected  bool
	}{
		{
			name:      "no expiration",
			expiresAt: nil,
			expected:  true,
		},
		{
			name:      "expires in future",
			expiresAt: timePtr(time.Now().Add(time.Hour)),
			expected:  true,
		},
		{
			name:      "already expired",
			expiresAt: timePtr(time.Now().Add(-time.Hour)),
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenInfo := &recovery.TokenInfo{
				ExpiresAt: tt.expiresAt,
			}

			result := tokenInfo.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDefaultTokenExpirationHandler(t *testing.T) {
	handler := recovery.NewDefaultTokenExpirationHandler()
	ctx := context.Background()

	tokenInfo := &recovery.TokenInfo{
		Service: "github",
		Token:   "test-token",
	}

	t.Run("OnTokenExpiring", func(t *testing.T) {
		err := handler.OnTokenExpiring(ctx, tokenInfo, time.Hour)
		assert.NoError(t, err)
	})

	t.Run("OnTokenExpired", func(t *testing.T) {
		err := handler.OnTokenExpired(ctx, tokenInfo)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expired")
	})

	t.Run("OnTokenRefreshed", func(t *testing.T) {
		newTokenInfo := &recovery.TokenInfo{
			Service: "github",
			Token:   "new-token",
		}
		err := handler.OnTokenRefreshed(ctx, tokenInfo, newTokenInfo)
		assert.NoError(t, err)
	})
}

func TestTokenManager(t *testing.T) {
	// Create mock server for token validation
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/user":
			// GitHub user endpoint
			w.Header().Set("X-OAuth-Scopes", "repo,read:org")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"login":"testuser"}`))
		case "/rate_limit":
			// GitHub rate limit endpoint
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"rate":{"limit":5000,"remaining":4999}}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Create token manager
	config := recovery.DefaultTokenManagerConfig()
	config.CheckInterval = 100 * time.Millisecond // Fast for testing
	manager := recovery.NewTokenManager(config)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := manager.Start(ctx)
	require.NoError(t, err)
	defer manager.Stop()

	t.Run("AddToken", func(t *testing.T) {
		// This will fail with the mock server since it's not github.com
		// But we can test the error handling
		err := manager.AddToken("github", "test-token")
		assert.Error(t, err) // Expected to fail with mock server
	})

	t.Run("GetToken for non-existent service", func(t *testing.T) {
		_, err := manager.GetToken("nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no token found")
	})
}

func TestDefaultTokenValidator(t *testing.T) {
	validator := recovery.NewDefaultTokenValidator()
	ctx := context.Background()

	t.Run("ValidateToken with invalid service", func(t *testing.T) {
		tokenInfo, err := validator.ValidateToken(ctx, "test-token", "unknown")
		require.NoError(t, err)
		assert.Equal(t, "unknown", tokenInfo.Service)
		assert.Equal(t, "test-token", tokenInfo.Token)
	})

	t.Run("CheckTokenHealth for unknown service", func(t *testing.T) {
		tokenInfo := &recovery.TokenInfo{
			Service: "unknown",
			Token:   "test-token",
		}
		err := validator.CheckTokenHealth(ctx, tokenInfo)
		assert.NoError(t, err) // Should skip health check
	})
}

func TestTokenManagerWithValidTokens(t *testing.T) {
	// Create mock validator that always succeeds
	mockValidator := &MockTokenValidator{}

	config := recovery.DefaultTokenManagerConfig()
	config.Validator = mockValidator
	config.CheckInterval = 50 * time.Millisecond

	manager := recovery.NewTokenManager(config)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err := manager.Start(ctx)
	require.NoError(t, err)
	defer manager.Stop()

	// Add a token
	err = manager.AddToken("github", "valid-token")
	require.NoError(t, err)

	// Get the token
	tokenInfo, err := manager.GetToken("github")
	require.NoError(t, err)
	assert.Equal(t, "github", tokenInfo.Service)
	assert.Equal(t, "valid-token", tokenInfo.Token)

	// Get token status
	status := manager.GetTokenStatus()
	assert.Contains(t, status, "github")
	assert.True(t, status["github"].IsValid)
}

func TestTokenManagerWithExpiringToken(t *testing.T) {
	// Create mock validator that returns expiring token
	mockValidator := &MockTokenValidator{
		returnExpiring: true,
	}

	// Create mock handler to capture events
	mockHandler := &MockTokenExpirationHandler{}

	config := recovery.DefaultTokenManagerConfig()
	config.Validator = mockValidator
	config.Handler = mockHandler
	config.ExpirationThreshold = 2 * time.Hour // Set threshold

	manager := recovery.NewTokenManager(config)

	// Add a token
	err := manager.AddToken("github", "expiring-token")
	require.NoError(t, err)

	// Get the token (should trigger expiring notification)
	_, err = manager.GetToken("github")
	require.NoError(t, err)

	// Check that expiring event was triggered
	assert.True(t, mockHandler.expiringCalled)
}

// Mock implementations for testing
type MockTokenValidator struct {
	returnExpiring bool
}

func (m *MockTokenValidator) ValidateToken(ctx context.Context, token, service string) (*recovery.TokenInfo, error) {
	tokenInfo := &recovery.TokenInfo{
		Token:       token,
		Service:     service,
		TokenType:   "classic",
		LastValidAt: time.Now(),
		Scopes:      []string{"repo", "read:org"},
		Metadata:    make(map[string]interface{}),
	}

	if m.returnExpiring {
		// Set expiration to 1 hour from now (within the 2-hour threshold)
		expiresAt := time.Now().Add(1 * time.Hour)
		tokenInfo.ExpiresAt = &expiresAt
	}

	return tokenInfo, nil
}

func (m *MockTokenValidator) CheckTokenHealth(ctx context.Context, tokenInfo *recovery.TokenInfo) error {
	return nil // Always healthy
}

type MockTokenExpirationHandler struct {
	expiringCalled  bool
	expiredCalled   bool
	refreshedCalled bool
}

func (m *MockTokenExpirationHandler) OnTokenExpiring(ctx context.Context, tokenInfo *recovery.TokenInfo, timeUntilExpiry time.Duration) error {
	m.expiringCalled = true
	return nil
}

func (m *MockTokenExpirationHandler) OnTokenExpired(ctx context.Context, tokenInfo *recovery.TokenInfo) error {
	m.expiredCalled = true
	return fmt.Errorf("token expired")
}

func (m *MockTokenExpirationHandler) OnTokenRefreshed(ctx context.Context, oldToken, newToken *recovery.TokenInfo) error {
	m.refreshedCalled = true
	return nil
}

// Helper functions
func timePtr(t time.Time) *time.Time {
	return &t
}
