package recovery

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// GitHubTokenRefresher handles GitHub OAuth2 token refresh
type GitHubTokenRefresher struct {
	clientID     string
	clientSecret string
	httpClient   *http.Client
}

// NewGitHubTokenRefresher creates a new GitHub token refresher
func NewGitHubTokenRefresher(clientID, clientSecret string) *GitHubTokenRefresher {
	return &GitHubTokenRefresher{
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CanRefresh checks if the token can be refreshed
func (gtr *GitHubTokenRefresher) CanRefresh(tokenInfo *TokenInfo) bool {
	// Can only refresh OAuth2 tokens with refresh tokens
	if tokenInfo.TokenType != "oauth2" {
		return false
	}

	refreshToken, exists := tokenInfo.Metadata["refresh_token"]
	return exists && refreshToken != nil
}

// RefreshToken refreshes a GitHub OAuth2 token
func (gtr *GitHubTokenRefresher) RefreshToken(ctx context.Context, tokenInfo *TokenInfo) (*TokenInfo, error) {
	refreshToken, exists := tokenInfo.Metadata["refresh_token"]
	if !exists {
		return nil, fmt.Errorf("no refresh token available")
	}

	refreshTokenStr, ok := refreshToken.(string)
	if !ok {
		return nil, fmt.Errorf("invalid refresh token format")
	}

	// Prepare refresh request
	data := url.Values{
		"client_id":     {gtr.clientID},
		"client_secret": {gtr.clientSecret},
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshTokenStr},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://github.com/login/oauth/access_token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := gtr.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh GitHub token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub token refresh failed with status %d", resp.StatusCode)
	}

	var refreshResp GitHubRefreshResponse
	if err := json.NewDecoder(resp.Body).Decode(&refreshResp); err != nil {
		return nil, fmt.Errorf("failed to decode refresh response: %w", err)
	}

	if refreshResp.Error != "" {
		return nil, fmt.Errorf("GitHub token refresh error: %s - %s", refreshResp.Error, refreshResp.ErrorDescription)
	}

	// Create new token info
	newTokenInfo := &TokenInfo{
		Token:       refreshResp.AccessToken,
		Service:     "github",
		TokenType:   "oauth2",
		LastValidAt: time.Now(),
		Metadata: map[string]interface{}{
			"refresh_token": refreshResp.RefreshToken,
			"token_type":    refreshResp.TokenType,
		},
	}

	// Set expiration if provided
	if refreshResp.ExpiresIn > 0 {
		expiresAt := time.Now().Add(time.Duration(refreshResp.ExpiresIn) * time.Second)
		newTokenInfo.ExpiresAt = &expiresAt
	}

	// Parse scopes
	if refreshResp.Scope != "" {
		newTokenInfo.Scopes = strings.Split(refreshResp.Scope, ",")
	}

	return newTokenInfo, nil
}

// GitHubRefreshResponse represents the response from GitHub token refresh
type GitHubRefreshResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	Scope            string `json:"scope"`
	ExpiresIn        int    `json:"expires_in"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// GitLabTokenRefresher handles GitLab OAuth2 token refresh
type GitLabTokenRefresher struct {
	clientID     string
	clientSecret string
	baseURL      string // e.g., "https://gitlab.com" or self-hosted instance
	httpClient   *http.Client
}

// NewGitLabTokenRefresher creates a new GitLab token refresher
func NewGitLabTokenRefresher(clientID, clientSecret, baseURL string) *GitLabTokenRefresher {
	if baseURL == "" {
		baseURL = "https://gitlab.com"
	}

	return &GitLabTokenRefresher{
		clientID:     clientID,
		clientSecret: clientSecret,
		baseURL:      baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CanRefresh checks if the token can be refreshed
func (glr *GitLabTokenRefresher) CanRefresh(tokenInfo *TokenInfo) bool {
	// Can only refresh OAuth2 tokens with refresh tokens
	if tokenInfo.TokenType != "oauth2" {
		return false
	}

	refreshToken, exists := tokenInfo.Metadata["refresh_token"]
	return exists && refreshToken != nil
}

// RefreshToken refreshes a GitLab OAuth2 token
func (glr *GitLabTokenRefresher) RefreshToken(ctx context.Context, tokenInfo *TokenInfo) (*TokenInfo, error) {
	refreshToken, exists := tokenInfo.Metadata["refresh_token"]
	if !exists {
		return nil, fmt.Errorf("no refresh token available")
	}

	refreshTokenStr, ok := refreshToken.(string)
	if !ok {
		return nil, fmt.Errorf("invalid refresh token format")
	}

	// Prepare refresh request
	data := url.Values{
		"client_id":     {glr.clientID},
		"client_secret": {glr.clientSecret},
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshTokenStr},
	}

	refreshURL := fmt.Sprintf("%s/oauth/token", glr.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", refreshURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := glr.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh GitLab token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitLab token refresh failed with status %d", resp.StatusCode)
	}

	var refreshResp GitLabRefreshResponse
	if err := json.NewDecoder(resp.Body).Decode(&refreshResp); err != nil {
		return nil, fmt.Errorf("failed to decode refresh response: %w", err)
	}

	if refreshResp.Error != "" {
		return nil, fmt.Errorf("GitLab token refresh error: %s - %s", refreshResp.Error, refreshResp.ErrorDescription)
	}

	// Create new token info
	newTokenInfo := &TokenInfo{
		Token:       refreshResp.AccessToken,
		Service:     "gitlab",
		TokenType:   "oauth2",
		LastValidAt: time.Now(),
		Metadata: map[string]interface{}{
			"refresh_token": refreshResp.RefreshToken,
			"token_type":    refreshResp.TokenType,
		},
	}

	// Set expiration if provided
	if refreshResp.ExpiresIn > 0 {
		expiresAt := time.Now().Add(time.Duration(refreshResp.ExpiresIn) * time.Second)
		newTokenInfo.ExpiresAt = &expiresAt
	}

	// Parse scopes
	if refreshResp.Scope != "" {
		newTokenInfo.Scopes = strings.Split(refreshResp.Scope, " ")
	}

	return newTokenInfo, nil
}

// GitLabRefreshResponse represents the response from GitLab token refresh
type GitLabRefreshResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	Scope            string `json:"scope"`
	ExpiresIn        int    `json:"expires_in"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// PersonalAccessTokenRefresher handles personal access token "refresh" through fallback mechanisms
type PersonalAccessTokenRefresher struct {
	fallbackTokens []string // List of backup tokens to try
	validator      TokenValidator
}

// NewPersonalAccessTokenRefresher creates a new PAT refresher with fallback tokens
func NewPersonalAccessTokenRefresher(fallbackTokens []string, validator TokenValidator) *PersonalAccessTokenRefresher {
	return &PersonalAccessTokenRefresher{
		fallbackTokens: fallbackTokens,
		validator:      validator,
	}
}

// CanRefresh checks if fallback tokens are available
func (par *PersonalAccessTokenRefresher) CanRefresh(tokenInfo *TokenInfo) bool {
	// Can "refresh" by falling back to alternative tokens
	return len(par.fallbackTokens) > 0
}

// RefreshToken attempts to use fallback tokens
func (par *PersonalAccessTokenRefresher) RefreshToken(ctx context.Context, tokenInfo *TokenInfo) (*TokenInfo, error) {
	// Try each fallback token until we find a valid one
	for i, fallbackToken := range par.fallbackTokens {
		if fallbackToken == tokenInfo.Token {
			continue // Skip the current token
		}

		newTokenInfo, err := par.validator.ValidateToken(ctx, fallbackToken, tokenInfo.Service)
		if err != nil {
			continue // Try next token
		}

		if newTokenInfo.IsValid() {
			// Remove used token from fallback list to avoid reuse
			par.fallbackTokens = append(par.fallbackTokens[:i], par.fallbackTokens[i+1:]...)
			return newTokenInfo, nil
		}
	}

	return nil, fmt.Errorf("no valid fallback tokens available for %s", tokenInfo.Service)
}

// TokenRefreshRegistry manages multiple token refreshers
type TokenRefreshRegistry struct {
	refreshers map[string][]TokenRefresher // service -> refreshers (in priority order)
}

// NewTokenRefreshRegistry creates a new token refresh registry
func NewTokenRefreshRegistry() *TokenRefreshRegistry {
	return &TokenRefreshRegistry{
		refreshers: make(map[string][]TokenRefresher),
	}
}

// AddRefresher adds a token refresher for a service
func (trr *TokenRefreshRegistry) AddRefresher(service string, refresher TokenRefresher) {
	trr.refreshers[service] = append(trr.refreshers[service], refresher)
}

// GetRefresher returns the first available refresher for a service and token
func (trr *TokenRefreshRegistry) GetRefresher(service string, tokenInfo *TokenInfo) TokenRefresher {
	refreshers, exists := trr.refreshers[service]
	if !exists {
		return nil
	}

	for _, refresher := range refreshers {
		if refresher.CanRefresh(tokenInfo) {
			return refresher
		}
	}

	return nil
}

// RefreshToken attempts to refresh a token using available refreshers
func (trr *TokenRefreshRegistry) RefreshToken(ctx context.Context, tokenInfo *TokenInfo) (*TokenInfo, error) {
	refresher := trr.GetRefresher(tokenInfo.Service, tokenInfo)
	if refresher == nil {
		return nil, fmt.Errorf("no suitable refresher found for %s token", tokenInfo.Service)
	}

	return refresher.RefreshToken(ctx, tokenInfo)
}

// TokenRefreshStrategy defines different strategies for token refresh
type TokenRefreshStrategy int

const (
	RefreshStrategyProactive TokenRefreshStrategy = iota // Refresh before expiration
	RefreshStrategyOnDemand                              // Refresh when needed
	RefreshStrategyFallback                              // Use fallback tokens only
)

// TokenRefreshPolicy defines when and how tokens should be refreshed
type TokenRefreshPolicy struct {
	Strategy           TokenRefreshStrategy
	RefreshThreshold   time.Duration // How early to refresh (for proactive strategy)
	MaxRefreshAttempts int           // Maximum number of refresh attempts
	RefreshRetryDelay  time.Duration // Delay between refresh attempts
	FallbackTokens     []string      // Fallback tokens to use
	NotifyOnRefresh    bool          // Whether to notify handlers on refresh
	AutoRetryOnFailure bool          // Whether to automatically retry failed operations with new token
}

// DefaultTokenRefreshPolicy returns a sensible default policy
func DefaultTokenRefreshPolicy() TokenRefreshPolicy {
	return TokenRefreshPolicy{
		Strategy:           RefreshStrategyProactive,
		RefreshThreshold:   24 * time.Hour, // Refresh 24h before expiration
		MaxRefreshAttempts: 3,
		RefreshRetryDelay:  5 * time.Minute,
		NotifyOnRefresh:    true,
		AutoRetryOnFailure: true,
	}
}

// TokenRefreshScheduler handles automatic token refresh scheduling
type TokenRefreshScheduler struct {
	tokenManager *TokenManager
	registry     *TokenRefreshRegistry
	policy       TokenRefreshPolicy

	// Internal state
	stopChan chan struct{}
	running  bool
}

// NewTokenRefreshScheduler creates a new token refresh scheduler
func NewTokenRefreshScheduler(tokenManager *TokenManager, registry *TokenRefreshRegistry, policy TokenRefreshPolicy) *TokenRefreshScheduler {
	return &TokenRefreshScheduler{
		tokenManager: tokenManager,
		registry:     registry,
		policy:       policy,
		stopChan:     make(chan struct{}),
	}
}

// Start begins the refresh scheduler
func (trs *TokenRefreshScheduler) Start(ctx context.Context) error {
	if trs.running {
		return fmt.Errorf("refresh scheduler is already running")
	}

	trs.running = true
	go trs.scheduleRefreshes(ctx)
	return nil
}

// Stop stops the refresh scheduler
func (trs *TokenRefreshScheduler) Stop() {
	if trs.running {
		close(trs.stopChan)
		trs.running = false
	}
}

// scheduleRefreshes handles automatic token refresh scheduling
func (trs *TokenRefreshScheduler) scheduleRefreshes(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour) // Check every hour
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-trs.stopChan:
			return
		case <-ticker.C:
			trs.checkAndRefreshTokens(ctx)
		}
	}
}

// checkAndRefreshTokens checks all tokens and refreshes those that need it
func (trs *TokenRefreshScheduler) checkAndRefreshTokens(ctx context.Context) {
	if trs.policy.Strategy != RefreshStrategyProactive {
		return // Only run for proactive strategy
	}

	tokenStatuses := trs.tokenManager.GetTokenStatus()

	for service, status := range tokenStatuses {
		if !status.IsValid {
			continue // Skip invalid tokens
		}

		// Check if token needs refresh
		shouldRefresh := false
		if status.ExpiresAt != nil {
			timeUntilExpiry := time.Until(*status.ExpiresAt)
			shouldRefresh = timeUntilExpiry <= trs.policy.RefreshThreshold
		}

		if shouldRefresh && status.CanRefresh {
			trs.refreshTokenWithRetry(ctx, service)
		}
	}
}

// refreshTokenWithRetry attempts to refresh a token with retry logic
func (trs *TokenRefreshScheduler) refreshTokenWithRetry(ctx context.Context, service string) {
	for attempt := 1; attempt <= trs.policy.MaxRefreshAttempts; attempt++ {
		tokenInfo, err := trs.tokenManager.GetToken(service)
		if err != nil {
			break // Token not found or other error
		}

		newTokenInfo, err := trs.registry.RefreshToken(ctx, tokenInfo)
		if err == nil {
			// Refresh successful, update token manager
			trs.tokenManager.AddToken(service, newTokenInfo.Token)
			return
		}

		if attempt < trs.policy.MaxRefreshAttempts {
			select {
			case <-ctx.Done():
				return
			case <-time.After(trs.policy.RefreshRetryDelay):
				// Continue to next attempt
			}
		}
	}
}
