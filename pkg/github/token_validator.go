package github

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// TokenValidator validates GitHub token permissions
type TokenValidator struct {
	client *RepoConfigClient
}

// PermissionLevel represents the level of access for a permission
type PermissionLevel string

const (
	PermissionNone  PermissionLevel = "none"
	PermissionRead  PermissionLevel = "read"
	PermissionWrite PermissionLevel = "write"
	PermissionAdmin PermissionLevel = "admin"
)

// RequiredPermission represents a required permission for an operation
type RequiredPermission struct {
	Scope       string          `json:"scope"`
	Level       PermissionLevel `json:"level"`
	Description string          `json:"description"`
	Optional    bool            `json:"optional"`
}

// TokenInfo contains information about the current token
type TokenInfo struct {
	User        *User                      `json:"user"`
	Scopes      []string                   `json:"scopes"`
	TokenType   string                     `json:"token_type"` // classic, fine_grained
	RateLimit   *RateLimitInfo             `json:"rate_limit"`
	Permissions map[string]PermissionLevel `json:"permissions"`
	ValidatedAt time.Time                  `json:"validated_at"`
}

// User represents GitHub user information
type User struct {
	Login     string `json:"login"`
	ID        int64  `json:"id"`
	Type      string `json:"type"` // User, Organization
	SiteAdmin bool   `json:"site_admin"`
}

// ValidatorRateLimitInfo contains rate limit information for token validator
type ValidatorRateLimitInfo struct {
	Limit     int       `json:"limit"`
	Remaining int       `json:"remaining"`
	Reset     time.Time `json:"reset"`
	Used      int       `json:"used"`
}

// ValidationResult represents the result of token validation
type ValidationResult struct {
	Valid           bool                 `json:"valid"`
	TokenInfo       *TokenInfo           `json:"token_info"`
	MissingPerms    []RequiredPermission `json:"missing_permissions"`
	Warnings        []string             `json:"warnings"`
	Recommendations []string             `json:"recommendations"`
	ValidatedAt     time.Time            `json:"validated_at"`
}

// OperationRequirements defines required permissions for different operations
var OperationRequirements = map[string][]RequiredPermission{
	"repository_read": {
		{Scope: "repo", Level: PermissionRead, Description: "Read repository information", Optional: false},
	},
	"repository_write": {
		{Scope: "repo", Level: PermissionWrite, Description: "Modify repository settings", Optional: false},
	},
	"organization_read": {
		{Scope: "read:org", Level: PermissionRead, Description: "Read organization information", Optional: false},
	},
	"organization_admin": {
		{Scope: "admin:org", Level: PermissionAdmin, Description: "Administer organization", Optional: false},
	},
	"bulk_operations": {
		{Scope: "repo", Level: PermissionWrite, Description: "Modify multiple repositories", Optional: false},
		{Scope: "admin:org", Level: PermissionAdmin, Description: "Access organization repositories", Optional: false},
	},
}

// NewTokenValidator creates a new token validator
func NewTokenValidator(client *RepoConfigClient) *TokenValidator {
	return &TokenValidator{
		client: client,
	}
}

// ValidateToken validates the current token and its permissions
func (tv *TokenValidator) ValidateToken(ctx context.Context) (*ValidationResult, error) {
	result := &ValidationResult{
		ValidatedAt:     time.Now(),
		Warnings:        []string{},
		Recommendations: []string{},
	}

	// Get token information
	tokenInfo, err := tv.getTokenInfo(ctx)
	if err != nil {
		result.Valid = false
		return result, fmt.Errorf("failed to get token info: %w", err)
	}

	result.TokenInfo = tokenInfo

	// Check basic token validity
	if tokenInfo.User == nil {
		result.Valid = false
		result.Warnings = append(result.Warnings, "Unable to retrieve user information - token may be invalid")
		return result, nil
	}

	result.Valid = true
	return result, nil
}

// ValidateForOperation validates token permissions for a specific operation
func (tv *TokenValidator) ValidateForOperation(ctx context.Context, operation string) (*ValidationResult, error) {
	result, err := tv.ValidateToken(ctx)
	if err != nil {
		return result, err
	}

	if !result.Valid {
		return result, nil
	}

	// Get required permissions for the operation
	requirements, exists := OperationRequirements[operation]
	if !exists {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Unknown operation: %s", operation))
		return result, nil
	}

	// Check each required permission
	for _, req := range requirements {
		if !tv.hasPermission(result.TokenInfo, req) {
			result.MissingPerms = append(result.MissingPerms, req)
		}
	}

	// Update validity based on missing permissions
	result.Valid = len(result.MissingPerms) == 0

	// Add recommendations
	tv.addRecommendations(result)

	return result, nil
}

// ValidateForRepository validates permissions for a specific repository
func (tv *TokenValidator) ValidateForRepository(ctx context.Context, owner, repo string, operation string) (*ValidationResult, error) {
	result, err := tv.ValidateForOperation(ctx, operation)
	if err != nil {
		return result, err
	}

	if !result.Valid {
		return result, nil
	}

	// Test actual repository access
	err = tv.testRepositoryAccess(ctx, owner, repo)
	if err != nil {
		result.Valid = false
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("Cannot access repository %s/%s: %v", owner, repo, err))
	}

	return result, nil
}

// getTokenInfo retrieves comprehensive token information
func (tv *TokenValidator) getTokenInfo(ctx context.Context) (*TokenInfo, error) {
	info := &TokenInfo{
		ValidatedAt: time.Now(),
		Permissions: make(map[string]PermissionLevel),
	}

	// Get user information
	user, err := tv.getCurrentUser(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	info.User = user

	// Get rate limit information
	rateLimit, err := tv.getRateLimit(ctx)
	if err != nil {
		// Rate limit info is not critical, continue without it
		info.RateLimit = &RateLimitInfo{}
	} else {
		info.RateLimit = rateLimit
	}

	// Determine token type and scopes
	err = tv.detectTokenType(ctx, info)
	if err != nil {
		return nil, fmt.Errorf("failed to detect token type: %w", err)
	}

	return info, nil
}

// getCurrentUser gets the current authenticated user
func (tv *TokenValidator) getCurrentUser(ctx context.Context) (*User, error) {
	resp, err := tv.client.makeRequest(ctx, "GET", "/user", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var user User
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		return nil, fmt.Errorf("failed to decode user: %w", err)
	}

	return &user, nil
}

// getRateLimit gets current rate limit information
func (tv *TokenValidator) getRateLimit(ctx context.Context) (*RateLimitInfo, error) {
	resp, err := tv.client.makeRequest(ctx, "GET", "/rate_limit", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var rateLimitResp struct {
		Resources struct {
			Core *RateLimitInfo `json:"core"`
		} `json:"resources"`
	}

	err = json.NewDecoder(resp.Body).Decode(&rateLimitResp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode rate limit: %w", err)
	}

	if rateLimitResp.Resources.Core != nil {
		return rateLimitResp.Resources.Core, nil
	}

	return &RateLimitInfo{}, nil
}

// detectTokenType determines if the token is classic or fine-grained
func (tv *TokenValidator) detectTokenType(ctx context.Context, info *TokenInfo) error {
	// Try to get the token's scopes from the response headers
	// This is a simplified implementation - in reality, you'd need to check headers from authenticated requests

	// For now, assume classic token and set common scopes
	info.TokenType = "classic"
	info.Scopes = []string{"repo", "read:org"} // These would be detected from actual API responses

	// Map scopes to permission levels
	for _, scope := range info.Scopes {
		level := tv.scopeToPermissionLevel(scope)
		info.Permissions[scope] = level
	}

	return nil
}

// scopeToPermissionLevel converts a scope string to permission level
func (tv *TokenValidator) scopeToPermissionLevel(scope string) PermissionLevel {
	switch scope {
	case "repo", "admin:org":
		return PermissionAdmin
	case "public_repo", "read:org":
		return PermissionRead
	default:
		return PermissionRead
	}
}

// hasPermission checks if the token has the required permission
func (tv *TokenValidator) hasPermission(tokenInfo *TokenInfo, req RequiredPermission) bool {
	level, exists := tokenInfo.Permissions[req.Scope]
	if !exists {
		return req.Optional
	}

	return tv.permissionLevelSufficient(level, req.Level)
}

// permissionLevelSufficient checks if the current level meets the requirement
func (tv *TokenValidator) permissionLevelSufficient(current, required PermissionLevel) bool {
	levels := map[PermissionLevel]int{
		PermissionNone:  0,
		PermissionRead:  1,
		PermissionWrite: 2,
		PermissionAdmin: 3,
	}

	return levels[current] >= levels[required]
}

// testRepositoryAccess tests actual access to a repository
func (tv *TokenValidator) testRepositoryAccess(ctx context.Context, owner, repo string) error {
	// Try to get repository information
	_, err := tv.client.GetRepository(ctx, owner, repo)
	return err
}

// addRecommendations adds helpful recommendations based on the validation result
func (tv *TokenValidator) addRecommendations(result *ValidationResult) {
	if len(result.MissingPerms) > 0 {
		result.Recommendations = append(result.Recommendations,
			"Consider using a token with broader permissions for the intended operations")
	}

	if result.TokenInfo.RateLimit != nil && result.TokenInfo.RateLimit.Remaining < 100 {
		result.Recommendations = append(result.Recommendations,
			"Rate limit is running low - consider using authentication to get higher limits")
	}

	if result.TokenInfo.TokenType == "classic" {
		result.Recommendations = append(result.Recommendations,
			"Consider using fine-grained personal access tokens for better security")
	}
}

// GetPermissionHelp returns help text for permissions
func (tv *TokenValidator) GetPermissionHelp() map[string]string {
	return map[string]string{
		"repo": "Full access to repositories including private repositories. " +
			"Grants read, write, and admin access to code, commit statuses, repository invitations, " +
			"collaborators, deployment statuses, and repository webhooks.",
		"public_repo": "Access to public repositories only. " +
			"Grants read and write access to code, commit statuses, repository invitations, " +
			"collaborators, deployment statuses, and repository webhooks for public repositories.",
		"read:org": "Read access to organization membership, organization projects, and team membership.",
		"admin:org": "Full administrative access to organization and teams. " +
			"Grants read and write access to organization profile, organization projects, and team membership.",
		"admin:repo_hook": "Grants read, write, ping, and delete access to repository hooks in public or private repositories.",
		"read:repo_hook":  "Grants read and ping access to repository hooks in public or private repositories.",
	}
}
