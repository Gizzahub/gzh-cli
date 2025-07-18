package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenValidator_ValidateToken(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/user":
			user := &User{
				Login:     "testuser",
				ID:        12345,
				Type:      "User",
				SiteAdmin: false,
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(user)

		case "/rate_limit":
			rateLimit := map[string]interface{}{
				"resources": map[string]interface{}{
					"core": &RateLimitInfo{
						Limit:     5000,
						Remaining: 4999,
						ResetTime: time.Now().Add(1 * time.Hour),
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(rateLimit)

		default:
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	client := NewRepoConfigClient("test-token")
	client.baseURL = server.URL

	validator := NewTokenValidator(client)

	ctx := context.Background()
	result, err := validator.ValidateToken(ctx)

	require.NoError(t, err)
	assert.True(t, result.Valid)
	assert.NotNil(t, result.TokenInfo)
	assert.Equal(t, "testuser", result.TokenInfo.User.Login)
	assert.Equal(t, int64(12345), result.TokenInfo.User.ID)
	assert.NotNil(t, result.TokenInfo.RateLimit)
	assert.Equal(t, 5000, result.TokenInfo.RateLimit.Limit)
}

func TestTokenValidator_ValidateForOperation(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/user":
			user := &User{
				Login: "testuser",
				ID:    12345,
				Type:  "User",
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(user)

		case "/rate_limit":
			rateLimit := map[string]interface{}{
				"resources": map[string]interface{}{
					"core": &RateLimitInfo{
						Limit:     5000,
						Remaining: 4999,
						ResetTime: time.Now().Add(1 * time.Hour),
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(rateLimit)

		default:
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	client := NewRepoConfigClient("test-token")
	client.baseURL = server.URL

	validator := NewTokenValidator(client)

	tests := []struct {
		name        string
		operation   string
		expectValid bool
	}{
		{
			name:        "repository read operation",
			operation:   "repository_read",
			expectValid: true,
		},
		{
			name:        "bulk operations",
			operation:   "bulk_operations",
			expectValid: false, // Should fail since we don't have admin:org scope
		},
		{
			name:        "unknown operation",
			operation:   "unknown_operation",
			expectValid: true, // Valid token but unknown operation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := validator.ValidateForOperation(ctx, tt.operation)

			require.NoError(t, err)
			assert.Equal(t, tt.expectValid, result.Valid)
			assert.NotNil(t, result.TokenInfo)
		})
	}
}

func TestTokenValidator_ValidateForRepository(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/user":
			user := &User{
				Login: "testuser",
				ID:    12345,
				Type:  "User",
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(user)

		case "/rate_limit":
			rateLimit := map[string]interface{}{
				"resources": map[string]interface{}{
					"core": &RateLimitInfo{
						Limit:     5000,
						Remaining: 4999,
						ResetTime: time.Now().Add(1 * time.Hour),
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(rateLimit)

		case "/repos/testorg/testrepo":
			repo := &Repository{
				ID:       1,
				Name:     "testrepo",
				FullName: "testorg/testrepo",
				Private:  false,
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(repo)

		case "/repos/testorg/nonexistent":
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"message": "Not Found",
			})

		default:
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	client := NewRepoConfigClient("test-token")
	client.baseURL = server.URL

	validator := NewTokenValidator(client)

	tests := []struct {
		name        string
		owner       string
		repo        string
		operation   string
		expectValid bool
	}{
		{
			name:        "accessible repository",
			owner:       "testorg",
			repo:        "testrepo",
			operation:   "repository_read",
			expectValid: true,
		},
		{
			name:        "non-existent repository",
			owner:       "testorg",
			repo:        "nonexistent",
			operation:   "repository_read",
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := validator.ValidateForRepository(ctx, tt.owner, tt.repo, tt.operation)

			require.NoError(t, err)
			assert.Equal(t, tt.expectValid, result.Valid)

			if !tt.expectValid {
				assert.NotEmpty(t, result.Warnings)
			}
		})
	}
}

func TestPermissionLevelSufficient(t *testing.T) {
	validator := &TokenValidator{}

	tests := []struct {
		name     string
		current  PermissionLevel
		required PermissionLevel
		expected bool
	}{
		{
			name:     "admin meets admin requirement",
			current:  PermissionAdmin,
			required: PermissionAdmin,
			expected: true,
		},
		{
			name:     "admin meets write requirement",
			current:  PermissionAdmin,
			required: PermissionWrite,
			expected: true,
		},
		{
			name:     "admin meets read requirement",
			current:  PermissionAdmin,
			required: PermissionRead,
			expected: true,
		},
		{
			name:     "write meets write requirement",
			current:  PermissionWrite,
			required: PermissionWrite,
			expected: true,
		},
		{
			name:     "write meets read requirement",
			current:  PermissionWrite,
			required: PermissionRead,
			expected: true,
		},
		{
			name:     "write does not meet admin requirement",
			current:  PermissionWrite,
			required: PermissionAdmin,
			expected: false,
		},
		{
			name:     "read meets read requirement",
			current:  PermissionRead,
			required: PermissionRead,
			expected: true,
		},
		{
			name:     "read does not meet write requirement",
			current:  PermissionRead,
			required: PermissionWrite,
			expected: false,
		},
		{
			name:     "none does not meet any requirement",
			current:  PermissionNone,
			required: PermissionRead,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.permissionLevelSufficient(tt.current, tt.required)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestScopeToPermissionLevel(t *testing.T) {
	validator := &TokenValidator{}

	tests := []struct {
		scope    string
		expected PermissionLevel
	}{
		{"repo", PermissionAdmin},
		{"admin:org", PermissionAdmin},
		{"public_repo", PermissionRead},
		{"read:org", PermissionRead},
		{"unknown_scope", PermissionRead},
	}

	for _, tt := range tests {
		t.Run(tt.scope, func(t *testing.T) {
			result := validator.scopeToPermissionLevel(tt.scope)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHasPermission(t *testing.T) {
	validator := &TokenValidator{}

	tokenInfo := &TokenInfo{
		Permissions: map[string]PermissionLevel{
			"repo":     PermissionAdmin,
			"read:org": PermissionRead,
		},
	}

	tests := []struct {
		name     string
		req      RequiredPermission
		expected bool
	}{
		{
			name: "has required repo permission",
			req: RequiredPermission{
				Scope:    "repo",
				Level:    PermissionWrite,
				Optional: false,
			},
			expected: true,
		},
		{
			name: "has required read permission",
			req: RequiredPermission{
				Scope:    "read:org",
				Level:    PermissionRead,
				Optional: false,
			},
			expected: true,
		},
		{
			name: "missing required permission",
			req: RequiredPermission{
				Scope:    "admin:org",
				Level:    PermissionAdmin,
				Optional: false,
			},
			expected: false,
		},
		{
			name: "missing optional permission",
			req: RequiredPermission{
				Scope:    "admin:org",
				Level:    PermissionAdmin,
				Optional: true,
			},
			expected: true,
		},
		{
			name: "insufficient permission level",
			req: RequiredPermission{
				Scope:    "read:org",
				Level:    PermissionWrite,
				Optional: false,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.hasPermission(tokenInfo, tt.req)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetPermissionHelp(t *testing.T) {
	validator := &TokenValidator{}

	help := validator.GetPermissionHelp()

	assert.NotEmpty(t, help)
	assert.Contains(t, help, "repo")
	assert.Contains(t, help, "public_repo")
	assert.Contains(t, help, "read:org")
	assert.Contains(t, help, "admin:org")

	// Check that help text is meaningful
	assert.Contains(t, help["repo"], "Full access")
	assert.Contains(t, help["read:org"], "Read access")
}

func TestOperationRequirements(t *testing.T) {
	// Test that operation requirements are properly defined
	assert.NotEmpty(t, OperationRequirements)

	// Check specific operations
	repoRead, exists := OperationRequirements["repository_read"]
	assert.True(t, exists)
	assert.NotEmpty(t, repoRead)
	assert.Equal(t, "repo", repoRead[0].Scope)

	bulkOps, exists := OperationRequirements["bulk_operations"]
	assert.True(t, exists)
	assert.Len(t, bulkOps, 2) // Should require both repo and admin:org
}
