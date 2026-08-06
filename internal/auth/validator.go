// Copyright (c) 2025 Gizzahub
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
	"strconv"
	"strings"
	"time"

	"github.com/gizzahub/gzh-cli/internal/constants"
	"github.com/gizzahub/gzh-cli/internal/httpclient"
	"github.com/gizzahub/gzh-cli/internal/validation"
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
	Type        TokenType       `json:"type"`
	Valid       bool            `json:"valid"`
	Username    string          `json:"username,omitempty"`
	Scopes      []string        `json:"scopes,omitempty"`
	ExpiresAt   *time.Time      `json:"expiresAt,omitempty"`
	RateLimit   *RateLimitInfo  `json:"rateLimit,omitempty"`
	Permissions map[string]bool `json:"permissions,omitempty"`
	Metadata    map[string]any  `json:"metadata,omitempty"`
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
	baseURLs   map[TokenType]string
}

// NewValidator creates a new authentication validator.
func NewValidator() *Validator {
	return &Validator{
		httpClient: httpclient.GetGlobalClient("default"),
		validator:  validation.New(),
		patterns:   initializeTokenPatterns(),
		baseURLs:   defaultBaseURLs(),
	}
}

// defaultBaseURLs returns the API 진입점을 플랫폼별로 돌려준다.
//
// 예전에는 각 validateXxxToken이 URL을 문자열 리터럴로 박아 두고 있었다.
// 그래서 (1) 테스트가 httptest 서버를 띄워도 연결할 방법이 없어 실제
// api.github.com으로 요청이 나갔고, (2) GitHub Enterprise나 자체 호스팅
// GitLab/Gitea를 쓰는 사용자는 토큰 검증 자체를 할 수 없었다. 이 도구는
// 여러 포지를 다루는 것이 목적이므로 후자가 더 큰 문제다.
func defaultBaseURLs() map[TokenType]string {
	//nolint:gosec // G101 오탐: TokenType 키 이름을 보고 자격증명으로 판단하지만 공개 API 주소다
	return map[TokenType]string{
		TokenTypeGitHub: "https://api.github.com",
		TokenTypeGitLab: "https://gitlab.com/api/v4",
		TokenTypeGitea:  "https://gitea.com/api/v1",
	}
}

// baseURL은 플랫폼의 API 진입점을 돌려준다. 등록되지 않은 타입은 기본값으로
// 되돌려 호출자가 빈 URL로 요청을 만드는 일이 없게 한다.
func (av *Validator) baseURL(tokenType TokenType) string {
	if url, ok := av.baseURLs[tokenType]; ok && url != "" {
		return url
	}

	return defaultBaseURLs()[tokenType]
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
	//
	// 빈 토큰은 여기서 먼저 막는다. 공용 검사기(internal/validation)는 빈 토큰을
	// 통과시키는데, 그쪽은 synclone 설정처럼 토큰이 선택 항목인 자리에서도
	// 쓰이기 때문이다. 반면 이 함수는 "이 자격증명이 유효한가"에 답하므로 빈
	// 문자열은 답할 대상 자체가 없는 상태다. 예전에는 그대로 통과해 빈
	// Authorization 헤더로 실제 API를 호출했고, 돌아온 401이 "토큰이 틀렸다"로
	// 읽혀 정작 "토큰을 안 줬다"는 사실을 가렸다.
	if token == "" {
		result.Errors = append(result.Errors, "Token format validation failed: token is empty")
		result.Duration = time.Since(start)

		return result, nil
	}

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
	req, err := http.NewRequestWithContext(ctx, "GET", av.baseURL(TokenTypeGitHub)+"/user", http.NoBody)
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
		Metadata:    make(map[string]any),
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

	req, err := http.NewRequestWithContext(ctx, "GET", av.baseURL(TokenTypeGitLab)+"/user", http.NoBody)
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
		Metadata:    make(map[string]any),
	}

	// GitLab rate limiting headers
	if limit := resp.Header.Get("RateLimit-Limit"); limit != "" {
		tokenInfo.RateLimit = &RateLimitInfo{
			Limit:     parseInt(limit),
			Remaining: parseInt(resp.Header.Get("RateLimit-Remaining")),
			// GitLab의 RateLimit-Reset은 RFC3339가 아니라 유닉스 epoch다.
			// 같은 저장소의 GitLab 클라이언트(pkg/gitlab/streaming_api.go)도
			// 같은 헤더를 epoch로 읽는다. RFC3339로 읽던 예전 코드는 어떤
			// 값을 받아도 파싱에 실패했는데, 실패시 time.Now()를 돌려주던
			// 자리표시자 때문에 그 사실이 드러나지 않았다.
			ResetTime: parseUnixTimestamp(resp.Header.Get("RateLimit-Reset")),
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

	req, err := http.NewRequestWithContext(ctx, "GET", av.baseURL(TokenTypeGitea)+"/user", http.NoBody)
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
		Metadata:    make(map[string]any),
	}, nil
}

// validateGenericToken provides basic validation for generic tokens.
func (av *Validator) validateGenericToken(_ context.Context, token string, tokenType TokenType) (*TokenInfo, error) {
	return &TokenInfo{
		Type:        tokenType,
		Valid:       len(token) >= constants.MinTokenLength,
		Permissions: make(map[string]bool),
		Metadata:    make(map[string]any),
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
//
// 셋 다 값을 읽지 않고 고정값을 돌려주는 자리표시자였다. 그래서 RateLimit
// 정보는 언제나 Limit=0, Remaining=0이었고, analyzeRateLimit의 소진 검사
// (Remaining < Limit/10)는 0 < 0이라 한 번도 참이 되지 않았으며
// rate_limit_available은 항상 false로 보고됐다 -- 남은 호출이 얼마든
// 결과가 같았다는 뜻이다.
//
// 헤더가 없거나 형식이 어긋나면 "모른다"에 해당하는 값(0, 제로 time)을
// 돌려준다. time.Now()를 돌려주면 파싱 실패가 "방금 초기화됨"으로 읽혀
// 값이 없다는 사실 자체가 사라진다.
func parseInt(value string) int {
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0
	}

	return parsed
}

func parseUnixTimestamp(value string) time.Time {
	seconds, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil {
		return time.Time{}
	}

	return time.Unix(seconds, 0)
}
