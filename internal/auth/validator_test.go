// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
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
	// Windows can report zero elapsed time for this immediately rejected input.
	// The result still records a valid, non-negative duration on every platform.
	assert.GreaterOrEqual(t, result.Duration, time.Duration(0))
	assert.NotZero(t, result.Timestamp)
}

func TestValidator_ValidateToken_PatternWarnings(t *testing.T) {
	// 형식 검사를 통과한 토큰은 기능 검증 단계에서 API를 호출한다. mock
	// 서버로 향하게 하지 않으면 이 테스트가 api.github.com에 실제로 붙는다.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	validator := NewValidator()
	validator.baseURLs[TokenTypeGitHub] = server.URL
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

			// 검사기를 위 mock 서버로 향하게 한다. 예전에는 서버를 띄워
			// 놓고 연결하지 않아 요청이 실제 api.github.com으로 나갔고,
			// 그래서 위 헤더 어설션은 한 번도 실행되지 않았으며
			// valid_token 케이스는 401을 받아 항상 실패했다.
			validator := NewValidator()
			validator.baseURLs[TokenTypeGitHub] = server.URL

			tokenInfo, err := validator.validateGitHubToken(context.Background(), "test_token")

			if test.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			require.NotNil(t, tokenInfo)
			assert.Equal(t, TokenTypeGitHub, tokenInfo.Type)
			assert.Equal(t, test.expectedValid, tokenInfo.Valid)
			assert.NotNil(t, tokenInfo.Permissions)
			assert.NotNil(t, tokenInfo.Metadata)

			// 헤더 값이 실제로 파싱되는지 확인한다. 예전 parseInt는 인자를
			// 버리고 0을 돌려주는 자리표시자였고, 어떤 테스트도 파싱 결과를
			// 보지 않아 RateLimit이 늘 0/0인 채로 남아 있었다.
			if test.headers["X-RateLimit-Limit"] != "" {
				require.NotNil(t, tokenInfo.RateLimit)
				assert.Equal(t, 5000, tokenInfo.RateLimit.Limit)
				assert.Equal(t, 4999, tokenInfo.RateLimit.Remaining)
				assert.Equal(t, time.Unix(1234567890, 0), tokenInfo.RateLimit.ResetTime)
				assert.Equal(t, []string{"repo", "user"}, tokenInfo.Scopes)
			}
		})
	}
}

// 예전에는 이 두 테스트가 gitlab.com/gitea.com으로 실제 요청을 보내고
// "성공하거나 네트워크 오류거나" 식으로 느슨하게 확인했다. 그 형태는 코드가
// 동작하든 아니든 통과했고 -- 실제로 401 응답의 문구가 기대와 달라 어긋났을
// 때에야 문제가 드러났다 -- CI를 외부 서비스에 묶어 두기도 했다.
func TestValidator_validateGitLabToken(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		headers       map[string]string
		expectedValid bool
		expectedError string
	}{
		{
			name:       "valid_token",
			statusCode: http.StatusOK,
			headers: map[string]string{
				"RateLimit-Limit":     "2000",
				"RateLimit-Remaining": "1999",
				"RateLimit-Reset":     "1234567890",
			},
			expectedValid: true,
		},
		{
			name:          "invalid_token",
			statusCode:    http.StatusUnauthorized,
			expectedValid: false,
			expectedError: "gitlab API returned status 401",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "Bearer test_token", r.Header.Get("Authorization"))
				assert.Equal(t, "/user", r.URL.Path)

				for key, value := range test.headers {
					w.Header().Set(key, value)
				}

				w.WriteHeader(test.statusCode)
			}))
			defer server.Close()

			validator := NewValidator()
			validator.baseURLs[TokenTypeGitLab] = server.URL

			tokenInfo, err := validator.validateGitLabToken(context.Background(), "test_token")

			if test.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.expectedError)
			} else {
				require.NoError(t, err)
			}

			require.NotNil(t, tokenInfo)
			assert.Equal(t, TokenTypeGitLab, tokenInfo.Type)
			assert.Equal(t, test.expectedValid, tokenInfo.Valid)
			assert.NotNil(t, tokenInfo.Permissions)
			assert.NotNil(t, tokenInfo.Metadata)

			if test.headers["RateLimit-Limit"] != "" {
				require.NotNil(t, tokenInfo.RateLimit)
				assert.Equal(t, 2000, tokenInfo.RateLimit.Limit)
				assert.Equal(t, 1999, tokenInfo.RateLimit.Remaining)
				assert.Equal(t, time.Unix(1234567890, 0), tokenInfo.RateLimit.ResetTime)
			}
		})
	}
}

func TestValidator_validateGiteaToken(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		expectedValid bool
	}{
		{name: "valid_token", statusCode: http.StatusOK, expectedValid: true},
		{name: "invalid_token", statusCode: http.StatusUnauthorized, expectedValid: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Gitea는 Bearer가 아니라 "token " 스킴을 쓴다.
				assert.Equal(t, "token test_token", r.Header.Get("Authorization"))
				assert.Equal(t, "/user", r.URL.Path)
				w.WriteHeader(test.statusCode)
			}))
			defer server.Close()

			validator := NewValidator()
			validator.baseURLs[TokenTypeGitea] = server.URL

			tokenInfo, err := validator.validateGiteaToken(context.Background(), "test_token")

			require.NoError(t, err)
			require.NotNil(t, tokenInfo)
			assert.Equal(t, TokenTypeGitea, tokenInfo.Type)
			assert.Equal(t, test.expectedValid, tokenInfo.Valid)
			assert.NotNil(t, tokenInfo.Permissions)
			assert.NotNil(t, tokenInfo.Metadata)
		})
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
				found := slices.Contains(result.Suggestions, expectedSuggestion)
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
			// 잔여 0은 10% 미만이므로 소진 경고가 나오는 것이 맞다. 예전에는
			// expectedWarnings를 아예 적지 않아 "경고 0개"를 기대했는데,
			// 한도를 다 쓴 상태에서 아무 말도 하지 않는 쪽이 오히려 문제다.
			// 완전 소진과 임박의 구분은 rate_limit_available이 담당한다.
			expectedWarnings: []string{
				"Rate limit is nearly exhausted",
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
		Metadata:    map[string]any{"key": "value"},
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
