// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

// Package auth provides comprehensive authentication validation and management
// for various Git hosting platforms and service integrations.
package auth

import (
	"context"
	"crypto/subtle"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/Gizzahub/gzh-cli/internal/constants"
	"github.com/Gizzahub/gzh-cli/internal/httpclient"
	"github.com/Gizzahub/gzh-cli/internal/validation"
)

// TokenType represents different authentication token types.
type TokenType string

// Supported token types for various Git hosting platforms.
const (
	// TokenTypeGitHub represents GitHub authentication tokens.
	TokenTypeGitHub TokenType = "github"
	// TokenTypeGitLab represents GitLab authentication tokens.
	TokenTypeGitLab TokenType = "gitlab"
	// TokenTypeGitea represents Gitea authentication tokens.
	TokenTypeGitea TokenType = "gitea"
	// TokenTypeGogs represents Gogs authentication tokens.
	TokenTypeGogs TokenType = "gogs"
	// TokenTypeBitbucket represents Bitbucket authentication tokens.
	TokenTypeBitbucket TokenType = "bitbucket"
	// TokenTypeAzureDevOps represents Azure DevOps authentication tokens.
	TokenTypeAzureDevOps TokenType = "azuredevops"
	// TokenTypeGeneric represents generic authentication tokens.
	TokenTypeGeneric TokenType = "generic"
)

// TokenInfo contains information about a validated token.
type TokenInfo struct {
	Type        TokenType              `json:"type"`
	Valid       bool                   `json:"valid"`
	Username    string                 `json:"username,omitempty"`
	Scopes      []string               `json:"scopes,omitempty"`
	ExpiresAt   *time.Time             `json:"expiresAt,omitempty"`
	RateLimit   *RateLimitInfo         `json:"rateLimit,omitempty"`
	Permissions map[string]bool        `json:"permissions,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// RateLimitInfo contains rate limiting information.
type RateLimitInfo struct {
	Limit     int       `json:"limit"`
	Remaining int       `json:"remaining"`
	ResetTime time.Time `json:"resetTime"`
}

// ValidationResult contains comprehensive validation results.
type ValidationResult struct {
	Valid       bool            `json:"valid"`
	TokenInfo   *TokenInfo      `json:"tokenInfo,omitempty"`
	Errors      []string        `json:"errors,omitempty"`
	Warnings    []string        `json:"warnings,omitempty"`
	Suggestions []string        `json:"suggestions,omitempty"`
	TestResults map[string]bool `json:"testResults,omitempty"`
	Duration    time.Duration   `json:"duration"`
	Timestamp   time.Time       `json:"timestamp"`
}

// Validator provides comprehensive authentication validation.
type Validator struct {
	httpClient *http.Client
	validator  *validation.Validator
	patterns   map[TokenType]*regexp.Regexp
}

// NewValidator creates a new authentication validator.
func NewValidator() *Validator {
	return &Validator{
		httpClient: httpclient.GetGlobalClient("default"),
		validator:  validation.New(),
		patterns:   initializeTokenPatterns(),
	}
}

// initializeTokenPatterns sets up token format validation patterns.
func initializeTokenPatterns() map[TokenType]*regexp.Regexp {
	return map[TokenType]*regexp.Regexp{
		TokenTypeGitHub:      regexp.MustCompile(`^(ghp_[a-zA-Z0-9]{36}|github_pat_[a-zA-Z0-9_]{82})$`),
		TokenTypeGitLab:      regexp.MustCompile(`^glpat-[a-zA-Z0-9_-]{20}$`),
		TokenTypeGitea:       regexp.MustCompile(`^[a-f0-9]{40}$`),
		TokenTypeGogs:        regexp.MustCompile(`^[a-f0-9]{40}$`),
		TokenTypeBitbucket:   regexp.MustCompile(`^[A-Za-z0-9+/=]{22,}$`),
		TokenTypeAzureDevOps: regexp.MustCompile(`^[a-zA-Z0-9]{52}$`),
		TokenTypeGeneric:     regexp.MustCompile(`^[a-zA-Z0-9_-]{8,}$`),
	}
}

// ValidateToken performs comprehensive token validation.
func (av *Validator) ValidateToken(ctx context.Context, token string, tokenType TokenType) (*ValidationResult, error) {
	start := time.Now()
	result := &ValidationResult{
		Timestamp:   start,
		TestResults: make(map[string]bool),
		Errors:      make([]string, 0),
		Warnings:    make([]string, 0),
		Suggestions: make([]string, 0),
	}

	// Basic input validation
	if err := av.validator.ValidateToken(token); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Token format validation failed: %v", err))
		result.Duration = time.Since(start)
		return result, nil
	}

	// Pattern-based validation
	if !av.validateTokenPattern(token, tokenType) {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Token does not match expected pattern for %s", tokenType))
		result.Suggestions = append(result.Suggestions, "Verify token format matches platform requirements")
	}

	// Functional validation
	tokenInfo, err := av.validateTokenFunctionality(ctx, token, tokenType)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Token functionality validation failed: %v", err))
	} else {
		result.TokenInfo = tokenInfo
		result.Valid = tokenInfo.Valid
	}

	// Security checks
	av.performSecurityChecks(token, tokenType, result)

	// Performance and rate limit checks
	if tokenInfo != nil && tokenInfo.RateLimit != nil {
		av.analyzeRateLimit(tokenInfo.RateLimit, result)
	}

	result.Duration = time.Since(start)
	return result, nil
}

// validateTokenPattern checks if token matches expected format for the platform.
func (av *Validator) validateTokenPattern(token string, tokenType TokenType) bool {
	pattern, exists := av.patterns[tokenType]
	if !exists {
		return false
	}
	return pattern.MatchString(token)
}

// validateTokenFunctionality tests token by making actual API calls.
func (av *Validator) validateTokenFunctionality(ctx context.Context, token string, tokenType TokenType) (*TokenInfo, error) {
	switch tokenType {
	case TokenTypeGitHub:
		return av.validateGitHubToken(ctx, token)
	case TokenTypeGitLab:
		return av.validateGitLabToken(ctx, token)
	case TokenTypeGitea:
		return av.validateGiteaToken(ctx, token)
	case TokenTypeGogs, TokenTypeBitbucket, TokenTypeAzureDevOps, TokenTypeGeneric:
		return av.validateGenericToken(ctx, token, tokenType)
	default:
		return av.validateGenericToken(ctx, token, tokenType)
	}
}

// validateGitHubToken validates GitHub tokens using the GitHub API.
func (av *Validator) validateGitHubToken(ctx context.Context, token string) (*TokenInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, constants.MediumHTTPTimeout)
	defer cancel()

	// Test token with user endpoint
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user", http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := av.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("api request failed: %w", err)
	}
	defer resp.Body.Close()

	tokenInfo := &TokenInfo{
		Type:        TokenTypeGitHub,
		Valid:       resp.StatusCode == http.StatusOK,
		Permissions: make(map[string]bool),
		Metadata:    make(map[string]interface{}),
	}

	// Extract rate limit information
	if remaining := resp.Header.Get("X-RateLimit-Remaining"); remaining != "" {
		tokenInfo.RateLimit = &RateLimitInfo{
			Remaining: parseInt(resp.Header.Get("X-RateLimit-Remaining")),
			Limit:     parseInt(resp.Header.Get("X-RateLimit-Limit")),
			ResetTime: parseUnixTimestamp(resp.Header.Get("X-RateLimit-Reset")),
		}
	}

	// Extract scopes from header
	if scopes := resp.Header.Get("X-OAuth-Scopes"); scopes != "" {
		tokenInfo.Scopes = strings.Split(strings.ReplaceAll(scopes, " ", ""), ",")
	}

	if resp.StatusCode != http.StatusOK {
		return tokenInfo, fmt.Errorf("github API returned status %d", resp.StatusCode)
	}

	return tokenInfo, nil
}

// validateGitLabToken validates GitLab tokens using the GitLab API.
func (av *Validator) validateGitLabToken(ctx context.Context, token string) (*TokenInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, constants.MediumHTTPTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://gitlab.com/api/v4/user", http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := av.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("api request failed: %w", err)
	}
	defer resp.Body.Close()

	tokenInfo := &TokenInfo{
		Type:        TokenTypeGitLab,
		Valid:       resp.StatusCode == http.StatusOK,
		Permissions: make(map[string]bool),
		Metadata:    make(map[string]interface{}),
	}

	// GitLab rate limiting headers
	if limit := resp.Header.Get("RateLimit-Limit"); limit != "" {
		tokenInfo.RateLimit = &RateLimitInfo{
			Limit:     parseInt(limit),
			Remaining: parseInt(resp.Header.Get("RateLimit-Remaining")),
			ResetTime: parseRFC3339Timestamp(resp.Header.Get("RateLimit-Reset")),
		}
	}

	if resp.StatusCode != http.StatusOK {
		return tokenInfo, fmt.Errorf("gitlab API returned status %d", resp.StatusCode)
	}

	return tokenInfo, nil
}

// validateGiteaToken validates Gitea tokens.
func (av *Validator) validateGiteaToken(ctx context.Context, token string) (*TokenInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, constants.MediumHTTPTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://gitea.com/api/v1/user", http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "token "+token)

	resp, err := av.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("api request failed: %w", err)
	}
	defer resp.Body.Close()

	return &TokenInfo{
		Type:        TokenTypeGitea,
		Valid:       resp.StatusCode == http.StatusOK,
		Permissions: make(map[string]bool),
		Metadata:    make(map[string]interface{}),
	}, nil
}

// validateGenericToken provides basic validation for generic tokens.
func (av *Validator) validateGenericToken(_ context.Context, token string, tokenType TokenType) (*TokenInfo, error) {
	return &TokenInfo{
		Type:        tokenType,
		Valid:       len(token) >= constants.MinTokenLength,
		Permissions: make(map[string]bool),
		Metadata:    make(map[string]interface{}),
	}, nil
}

// performSecurityChecks performs additional security validations.
func (av *Validator) performSecurityChecks(token string, _ TokenType, result *ValidationResult) {
	// Check for common insecure patterns
	if strings.Contains(strings.ToLower(token), "test") ||
		strings.Contains(strings.ToLower(token), "demo") ||
		strings.Contains(strings.ToLower(token), "example") {
		result.Warnings = append(result.Warnings, "Token appears to be a test/demo token")
		result.Suggestions = append(result.Suggestions, "Use production tokens for real operations")
	}

	// Check token entropy (basic check)
	if len(token) < 20 {
		result.Warnings = append(result.Warnings, "Token appears to have low entropy")
		result.Suggestions = append(result.Suggestions, "Ensure token has sufficient randomness")
	}

	// Check for common leaked token patterns
	if strings.HasPrefix(token, "ghp_") && len(token) != 40 {
		result.Warnings = append(result.Warnings, "GitHub token format appears incorrect")
	}
}

// analyzeRateLimit analyzes rate limit information and provides recommendations.
func (av *Validator) analyzeRateLimit(rateLimit *RateLimitInfo, result *ValidationResult) {
	if rateLimit.Remaining < rateLimit.Limit/10 { // Less than 10% remaining
		result.Warnings = append(result.Warnings, "Rate limit is nearly exhausted")
		result.Suggestions = append(result.Suggestions, "Consider implementing rate limiting in your application")
	}

	if time.Until(rateLimit.ResetTime) > time.Hour {
		result.Warnings = append(result.Warnings, "Rate limit reset time is far in the future")
	}

	result.TestResults["rate_limit_available"] = rateLimit.Remaining > 0
}

// SecureTokenComparison compares tokens using constant-time comparison.
func (av *Validator) SecureTokenComparison(token1, token2 string) bool {
	return subtle.ConstantTimeCompare([]byte(token1), []byte(token2)) == 1
}

// Helper functions for parsing headers.
func parseInt(_ string) int {
	// Simple implementation - in production you'd want proper error handling
	// Currently returns 0 as placeholder
	return 0
}

func parseUnixTimestamp(_ string) time.Time {
	// Implementation would parse unix timestamp
	// Currently returns current time as placeholder
	return time.Now()
}

func parseRFC3339Timestamp(_ string) time.Time {
	// Implementation would parse RFC3339 timestamp
	// Currently returns current time as placeholder
	return time.Now()
}
