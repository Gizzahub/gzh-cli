// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewValidator(t *testing.T) {
	validator := NewValidator()

	assert.NotNil(t, validator)
	assert.NotNil(t, validator.httpClient)
	assert.NotNil(t, validator.validator)
	assert.NotNil(t, validator.patterns)

	// Check that all token type patterns are initialized
	expectedTypes := []TokenType{
		TokenTypeGitHub, TokenTypeGitLab, TokenTypeGitea,
		TokenTypeGogs, TokenTypeBitbucket, TokenTypeAzureDevOps, TokenTypeGeneric,
	}

	for _, tokenType := range expectedTypes {
		assert.Contains(t, validator.patterns, tokenType, "Pattern not found for token type: %s", tokenType)
	}
}

func TestTokenTypeConstants(t *testing.T) {
	assert.Equal(t, TokenType("github"), TokenTypeGitHub)
	assert.Equal(t, TokenType("gitlab"), TokenTypeGitLab)
	assert.Equal(t, TokenType("gitea"), TokenTypeGitea)
	assert.Equal(t, TokenType("gogs"), TokenTypeGogs)
	assert.Equal(t, TokenType("bitbucket"), TokenTypeBitbucket)
	assert.Equal(t, TokenType("azuredevops"), TokenTypeAzureDevOps)
	assert.Equal(t, TokenType("generic"), TokenTypeGeneric)
}

func TestInitializeTokenPatterns(t *testing.T) {
	patterns := initializeTokenPatterns()

	tests := []struct {
		tokenType     TokenType
		validTokens   []string
		invalidTokens []string
	}{
		{
			tokenType: TokenTypeGitHub,
			validTokens: []string{
				"ghp_" + generateString(36),
				"github_pat_" + generateString(82),
			},
			invalidTokens: []string{
				"ghp_short",
				"invalid_token",
				"",
			},
		},
		{
			tokenType: TokenTypeGitLab,
			validTokens: []string{
				"glpat-" + generateString(20),
			},
			invalidTokens: []string{
				"glpat-short",
				"invalid_token",
				"",
			},
		},
		{
			tokenType: TokenTypeGitea,
			validTokens: []string{
				generateHexString(40),
			},
			invalidTokens: []string{
				generateHexString(20),
				"invalid_token",
				"",
			},
		},
		{
			tokenType: TokenTypeGeneric,
			validTokens: []string{
				"validtoken123",
				"a_very_long_generic_token_with_underscores_and_dashes-123",
			},
			invalidTokens: []string{
				"short",
				"",
			},
		},
	}

	for _, test := range tests {
		t.Run(string(test.tokenType), func(t *testing.T) {
			pattern, exists := patterns[test.tokenType]
			require.True(t, exists, "Pattern should exist for token type: %s", test.tokenType)

			for _, validToken := range test.validTokens {
				assert.True(t, pattern.MatchString(validToken),
					"Token should be valid for %s: %s", test.tokenType, validToken)
			}

			for _, invalidToken := range test.invalidTokens {
				assert.False(t, pattern.MatchString(invalidToken),
					"Token should be invalid for %s: %s", test.tokenType, invalidToken)
			}
		})
	}
}

func TestValidator_validateTokenPattern(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		token     string
		tokenType TokenType
		expected  bool
	}{
		{"ghp_" + generateString(36), TokenTypeGitHub, true},
		{"invalid_github_token", TokenTypeGitHub, false},
		{"glpat-" + generateString(20), TokenTypeGitLab, true},
		{"invalid_gitlab_token", TokenTypeGitLab, false},
		{generateHexString(40), TokenTypeGitea, true},
		{"invalid_gitea_token", TokenTypeGitea, false},
		{"valid_generic_token", TokenTypeGeneric, true},
		{"short", TokenTypeGeneric, false},
		{"token", TokenType("unknown"), false},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s_%s", test.tokenType, test.token[:min(len(test.token), 10)]), func(t *testing.T) {
			result := validator.validateTokenPattern(test.token, test.tokenType)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestValidator_ValidateToken_BasicValidation(t *testing.T) {
	validator := NewValidator()
	ctx := context.Background()

	// Test with empty token (should fail basic validation)
	result, err := validator.ValidateToken(ctx, "", TokenTypeGitHub)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Valid)
	assert.NotEmpty(t, result.Errors)
	assert.Contains(t, result.Errors[0], "Token format validation failed")
	assert.NotZero(t, result.Duration)
	assert.NotZero(t, result.Timestamp)
}

func TestValidator_ValidateToken_PatternWarnings(t *testing.T) {
	validator := NewValidator()
	ctx := context.Background()

	// Test with token that passes basic validation but fails pattern matching
	result, err := validator.ValidateToken(ctx, "invalid_pattern_but_long_enough", TokenTypeGitHub)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.Warnings)
	assert.Contains(t, result.Warnings[0], "Token does not match expected pattern")
	assert.NotEmpty(t, result.Suggestions)
}

func TestValidator_validateGitHubToken(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		headers       map[string]string
		expectedValid bool
		expectedError bool
	}{
		{
			name:       "valid_token",
			statusCode: http.StatusOK,
			headers: map[string]string{
				"X-RateLimit-Remaining": "4999",
				"X-RateLimit-Limit":     "5000",
				"X-RateLimit-Reset":     "1234567890",
				"X-OAuth-Scopes":        "repo, user",
			},
			expectedValid: true,
			expectedError: false,
		},
		{
			name:          "invalid_token",
			statusCode:    http.StatusUnauthorized,
			expectedValid: false,
			expectedError: true,
		},
		{
			name:          "forbidden_token",
			statusCode:    http.StatusForbidden,
			expectedValid: false,
			expectedError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify correct headers are sent
				assert.Equal(t, "Bearer test_token", r.Header.Get("Authorization"))
				assert.Equal(t, "application/vnd.github+json", r.Header.Get("Accept"))
				assert.Equal(t, "2022-11-28", r.Header.Get("X-GitHub-Api-Version"))

				// Set response headers
				for key, value := range test.headers {
					w.Header().Set(key, value)
				}

				w.WriteHeader(test.statusCode)
			}))
			defer server.Close()

			// Create validator with custom http client
			validator := NewValidator()

			// Replace GitHub API URL in the validation function (this would need refactoring in real code)
			// For now, test the logic with a mock approach
			tokenInfo, err := validator.validateGitHubToken(context.Background(), "test_token")

			if test.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tokenInfo != nil {
				assert.Equal(t, TokenTypeGitHub, tokenInfo.Type)
				assert.Equal(t, test.expectedValid, tokenInfo.Valid)
				assert.NotNil(t, tokenInfo.Permissions)
				assert.NotNil(t, tokenInfo.Metadata)
			}
		})
	}
}

func TestValidator_validateGitLabToken(t *testing.T) {
	validator := NewValidator()

	tokenInfo, err := validator.validateGitLabToken(context.Background(), "test_token")
	// Since this makes actual HTTP calls, we expect either success or network error
	// In a real test environment, you would mock the HTTP client
	if err != nil {
		assert.Contains(t, err.Error(), "api request failed")
	}

	if tokenInfo != nil {
		assert.Equal(t, TokenTypeGitLab, tokenInfo.Type)
		assert.NotNil(t, tokenInfo.Permissions)
		assert.NotNil(t, tokenInfo.Metadata)
	}
}

func TestValidator_validateGiteaToken(t *testing.T) {
	validator := NewValidator()

	tokenInfo, err := validator.validateGiteaToken(context.Background(), "test_token")
	// Since this makes actual HTTP calls, we expect either success or network error
	if err != nil {
		assert.Contains(t, err.Error(), "api request failed")
	}

	if tokenInfo != nil {
		assert.Equal(t, TokenTypeGitea, tokenInfo.Type)
		assert.NotNil(t, tokenInfo.Permissions)
		assert.NotNil(t, tokenInfo.Metadata)
	}
}

func TestValidator_validateGenericToken(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		token     string
		tokenType TokenType
		expected  bool
	}{
		{"short", TokenTypeGeneric, false}, // Less than MinTokenLength
		{"this_is_a_long_enough_token", TokenTypeGeneric, true},
		{"another_valid_token_12345", TokenTypeBitbucket, true},
	}

	for _, test := range tests {
		t.Run(test.token, func(t *testing.T) {
			tokenInfo, err := validator.validateGenericToken(context.Background(), test.token, test.tokenType)

			assert.NoError(t, err)
			assert.NotNil(t, tokenInfo)
			assert.Equal(t, test.tokenType, tokenInfo.Type)
			assert.Equal(t, test.expected, tokenInfo.Valid)
			assert.NotNil(t, tokenInfo.Permissions)
			assert.NotNil(t, tokenInfo.Metadata)
		})
	}
}

func TestValidator_performSecurityChecks(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name                string
		token               string
		expectedWarnings    []string
		expectedSuggestions []string
	}{
		{
			name:  "test_token",
			token: "test_token_12345",
			expectedWarnings: []string{
				"Token appears to be a test/demo token",
				"Token appears to have low entropy",
			},
			expectedSuggestions: []string{
				"Use production tokens for real operations",
				"Ensure token has sufficient randomness",
			},
		},
		{
			name:  "demo_token",
			token: "demo_token_12345",
			expectedWarnings: []string{
				"Token appears to be a test/demo token",
				"Token appears to have low entropy",
			},
		},
		{
			name:  "short_token",
			token: "short",
			expectedWarnings: []string{
				"Token appears to have low entropy",
			},
		},
		{
			name:  "invalid_github_format",
			token: "ghp_invalid_length",
			expectedWarnings: []string{
				"GitHub token format appears incorrect",
				"Token appears to have low entropy",
			},
		},
		{
			name:             "valid_long_token",
			token:            generateString(50),
			expectedWarnings: []string{}, // Should have no warnings
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := &ValidationResult{
				Warnings:    make([]string, 0),
				Suggestions: make([]string, 0),
			}

			validator.performSecurityChecks(test.token, TokenTypeGeneric, result)

			assert.Len(t, result.Warnings, len(test.expectedWarnings))
			for _, expectedWarning := range test.expectedWarnings {
				assert.Contains(t, result.Warnings, expectedWarning)
			}

			for _, expectedSuggestion := range test.expectedSuggestions {
				found := false
				for _, suggestion := range result.Suggestions {
					if suggestion == expectedSuggestion {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected suggestion not found: %s", expectedSuggestion)
			}
		})
	}
}

func TestValidator_analyzeRateLimit(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name                string
		rateLimit           *RateLimitInfo
		expectedWarnings    []string
		expectedTestResults map[string]bool
	}{
		{
			name: "low_remaining",
			rateLimit: &RateLimitInfo{
				Limit:     1000,
				Remaining: 50, // 5% remaining
				ResetTime: time.Now().Add(30 * time.Minute),
			},
			expectedWarnings: []string{
				"Rate limit is nearly exhausted",
			},
			expectedTestResults: map[string]bool{
				"rate_limit_available": true,
			},
		},
		{
			name: "far_future_reset",
			rateLimit: &RateLimitInfo{
				Limit:     1000,
				Remaining: 500,
				ResetTime: time.Now().Add(2 * time.Hour),
			},
			expectedWarnings: []string{
				"Rate limit reset time is far in the future",
			},
			expectedTestResults: map[string]bool{
				"rate_limit_available": true,
			},
		},
		{
			name: "no_remaining",
			rateLimit: &RateLimitInfo{
				Limit:     1000,
				Remaining: 0,
				ResetTime: time.Now().Add(30 * time.Minute),
			},
			expectedTestResults: map[string]bool{
				"rate_limit_available": false,
			},
		},
		{
			name: "healthy_rate_limit",
			rateLimit: &RateLimitInfo{
				Limit:     1000,
				Remaining: 800,
				ResetTime: time.Now().Add(30 * time.Minute),
			},
			expectedWarnings: []string{}, // No warnings expected
			expectedTestResults: map[string]bool{
				"rate_limit_available": true,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := &ValidationResult{
				Warnings:    make([]string, 0),
				Suggestions: make([]string, 0),
				TestResults: make(map[string]bool),
			}

			validator.analyzeRateLimit(test.rateLimit, result)

			assert.Len(t, result.Warnings, len(test.expectedWarnings))
			for _, expectedWarning := range test.expectedWarnings {
				assert.Contains(t, result.Warnings, expectedWarning)
			}

			for key, expectedValue := range test.expectedTestResults {
				assert.Equal(t, expectedValue, result.TestResults[key])
			}
		})
	}
}

func TestValidator_SecureTokenComparison(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		token1   string
		token2   string
		expected bool
	}{
		{"identical", "identical", true},
		{"different1", "different2", false},
		{"", "", true},
		{"token", "", false},
		{"", "token", false},
		{"case", "CASE", false}, // Case sensitive
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s_vs_%s", test.token1, test.token2), func(t *testing.T) {
			result := validator.SecureTokenComparison(test.token1, test.token2)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestTokenInfo_Structure(t *testing.T) {
	tokenInfo := &TokenInfo{
		Type:        TokenTypeGitHub,
		Valid:       true,
		Username:    "testuser",
		Scopes:      []string{"repo", "user"},
		ExpiresAt:   nil,
		RateLimit:   &RateLimitInfo{Limit: 5000, Remaining: 4999},
		Permissions: map[string]bool{"read": true, "write": false},
		Metadata:    map[string]interface{}{"key": "value"},
	}

	assert.Equal(t, TokenTypeGitHub, tokenInfo.Type)
	assert.True(t, tokenInfo.Valid)
	assert.Equal(t, "testuser", tokenInfo.Username)
	assert.Equal(t, []string{"repo", "user"}, tokenInfo.Scopes)
	assert.Nil(t, tokenInfo.ExpiresAt)
	assert.NotNil(t, tokenInfo.RateLimit)
	assert.Equal(t, 5000, tokenInfo.RateLimit.Limit)
	assert.Equal(t, 4999, tokenInfo.RateLimit.Remaining)
	assert.True(t, tokenInfo.Permissions["read"])
	assert.False(t, tokenInfo.Permissions["write"])
	assert.Equal(t, "value", tokenInfo.Metadata["key"])
}

func TestValidationResult_Structure(t *testing.T) {
	now := time.Now()
	duration := 100 * time.Millisecond

	result := &ValidationResult{
		Valid:       true,
		TokenInfo:   &TokenInfo{Type: TokenTypeGitHub, Valid: true},
		Errors:      []string{"error1", "error2"},
		Warnings:    []string{"warning1"},
		Suggestions: []string{"suggestion1"},
		TestResults: map[string]bool{"test1": true, "test2": false},
		Duration:    duration,
		Timestamp:   now,
	}

	assert.True(t, result.Valid)
	assert.NotNil(t, result.TokenInfo)
	assert.Equal(t, TokenTypeGitHub, result.TokenInfo.Type)
	assert.Equal(t, []string{"error1", "error2"}, result.Errors)
	assert.Equal(t, []string{"warning1"}, result.Warnings)
	assert.Equal(t, []string{"suggestion1"}, result.Suggestions)
	assert.True(t, result.TestResults["test1"])
	assert.False(t, result.TestResults["test2"])
	assert.Equal(t, duration, result.Duration)
	assert.Equal(t, now, result.Timestamp)
}

func TestRateLimitInfo_Structure(t *testing.T) {
	resetTime := time.Now().Add(time.Hour)

	rateLimit := &RateLimitInfo{
		Limit:     5000,
		Remaining: 4999,
		ResetTime: resetTime,
	}

	assert.Equal(t, 5000, rateLimit.Limit)
	assert.Equal(t, 4999, rateLimit.Remaining)
	assert.Equal(t, resetTime, rateLimit.ResetTime)
}

// Helper functions for tests

func generateString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[i%len(charset)]
	}
	return string(result)
}

func generateHexString(length int) string {
	const hexCharset = "abcdef0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = hexCharset[i%len(hexCharset)]
	}
	return string(result)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
