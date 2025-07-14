package recovery

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// TokenInfo represents token metadata including expiration
type TokenInfo struct {
	Token       string                 `json:"token"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
	Scopes      []string               `json:"scopes,omitempty"`
	TokenType   string                 `json:"token_type"` // "classic", "fine_grained", "oauth2"
	Service     string                 `json:"service"`    // "github", "gitlab", "gitea"
	LastValidAt time.Time              `json:"last_valid_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// IsExpired checks if the token is expired or will expire within the threshold
func (ti *TokenInfo) IsExpired(threshold time.Duration) bool {
	if ti.ExpiresAt == nil {
		return false // No expiration date available
	}
	return time.Now().Add(threshold).After(*ti.ExpiresAt)
}

// IsValid checks if the token is currently valid (not expired)
func (ti *TokenInfo) IsValid() bool {
	if ti.ExpiresAt == nil {
		return true // Assume valid if no expiration
	}
	return time.Now().Before(*ti.ExpiresAt)
}

// TokenExpirationHandler defines interface for handling token expiration events
type TokenExpirationHandler interface {
	OnTokenExpiring(ctx context.Context, tokenInfo *TokenInfo, timeUntilExpiry time.Duration) error
	OnTokenExpired(ctx context.Context, tokenInfo *TokenInfo) error
	OnTokenRefreshed(ctx context.Context, oldToken, newToken *TokenInfo) error
}

// DefaultTokenExpirationHandler provides default implementations
type DefaultTokenExpirationHandler struct {
	// Logger for token events (could be replaced with actual logger)
	logFunc func(format string, args ...interface{})
}

func NewDefaultTokenExpirationHandler() *DefaultTokenExpirationHandler {
	return &DefaultTokenExpirationHandler{
		logFunc: func(format string, args ...interface{}) {
			fmt.Printf("[TOKEN-MANAGER] "+format+"\n", args...)
		},
	}
}

func (h *DefaultTokenExpirationHandler) OnTokenExpiring(ctx context.Context, tokenInfo *TokenInfo, timeUntilExpiry time.Duration) error {
	h.logFunc("Token for %s expiring in %v", tokenInfo.Service, timeUntilExpiry)
	return nil
}

func (h *DefaultTokenExpirationHandler) OnTokenExpired(ctx context.Context, tokenInfo *TokenInfo) error {
	h.logFunc("Token for %s has expired", tokenInfo.Service)
	return fmt.Errorf("token for %s has expired", tokenInfo.Service)
}

func (h *DefaultTokenExpirationHandler) OnTokenRefreshed(ctx context.Context, oldToken, newToken *TokenInfo) error {
	h.logFunc("Token for %s has been refreshed", newToken.Service)
	return nil
}

// TokenRefresher defines interface for refreshing tokens
type TokenRefresher interface {
	RefreshToken(ctx context.Context, tokenInfo *TokenInfo) (*TokenInfo, error)
	CanRefresh(tokenInfo *TokenInfo) bool
}

// TokenValidator defines interface for validating token status
type TokenValidator interface {
	ValidateToken(ctx context.Context, token string, service string) (*TokenInfo, error)
	CheckTokenHealth(ctx context.Context, tokenInfo *TokenInfo) error
}

// TokenManager manages token lifecycle including expiration monitoring and refresh
type TokenManager struct {
	tokens     map[string]*TokenInfo // service -> token info
	refreshers map[string]TokenRefresher
	validator  TokenValidator
	handler    TokenExpirationHandler

	// Configuration
	expirationThreshold time.Duration // How early to warn about expiration
	checkInterval       time.Duration // How often to check token status

	// Internal state
	mu       sync.RWMutex
	stopChan chan struct{}
	running  bool
}

// TokenManagerConfig configures the token manager
type TokenManagerConfig struct {
	ExpirationThreshold time.Duration // Default: 24h before expiration
	CheckInterval       time.Duration // Default: 1h check interval
	Handler             TokenExpirationHandler
	Validator           TokenValidator
}

// DefaultTokenManagerConfig returns sensible defaults
func DefaultTokenManagerConfig() TokenManagerConfig {
	return TokenManagerConfig{
		ExpirationThreshold: 24 * time.Hour,
		CheckInterval:       1 * time.Hour,
		Handler:             NewDefaultTokenExpirationHandler(),
		Validator:           NewDefaultTokenValidator(),
	}
}

// NewTokenManager creates a new token manager
func NewTokenManager(config TokenManagerConfig) *TokenManager {
	return &TokenManager{
		tokens:              make(map[string]*TokenInfo),
		refreshers:          make(map[string]TokenRefresher),
		validator:           config.Validator,
		handler:             config.Handler,
		expirationThreshold: config.ExpirationThreshold,
		checkInterval:       config.CheckInterval,
		stopChan:            make(chan struct{}),
	}
}

// AddToken adds a token to be managed
func (tm *TokenManager) AddToken(service, token string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Validate token and get metadata
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tokenInfo, err := tm.validator.ValidateToken(ctx, token, service)
	if err != nil {
		return fmt.Errorf("failed to validate token for %s: %w", service, err)
	}

	tm.tokens[service] = tokenInfo
	return nil
}

// GetToken retrieves a valid token for the service
func (tm *TokenManager) GetToken(service string) (*TokenInfo, error) {
	tm.mu.RLock()
	tokenInfo, exists := tm.tokens[service]
	tm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no token found for service %s", service)
	}

	// Check if token is expired
	if !tokenInfo.IsValid() {
		// Try to refresh if possible
		if refresher, hasRefresher := tm.refreshers[service]; hasRefresher && refresher.CanRefresh(tokenInfo) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			newToken, err := refresher.RefreshToken(ctx, tokenInfo)
			if err != nil {
				return nil, fmt.Errorf("failed to refresh expired token for %s: %w", service, err)
			}

			tm.mu.Lock()
			tm.tokens[service] = newToken
			tm.mu.Unlock()

			// Notify handler
			tm.handler.OnTokenRefreshed(context.Background(), tokenInfo, newToken)
			return newToken, nil
		}

		// Can't refresh, token is expired
		tm.handler.OnTokenExpired(context.Background(), tokenInfo)
		return nil, fmt.Errorf("token for %s is expired and cannot be refreshed", service)
	}

	// Check if token is expiring soon
	if tokenInfo.IsExpired(tm.expirationThreshold) {
		timeUntilExpiry := tokenInfo.ExpiresAt.Sub(time.Now())
		tm.handler.OnTokenExpiring(context.Background(), tokenInfo, timeUntilExpiry)
	}

	return tokenInfo, nil
}

// AddRefresher adds a token refresher for a service
func (tm *TokenManager) AddRefresher(service string, refresher TokenRefresher) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.refreshers[service] = refresher
}

// Start begins monitoring token expiration
func (tm *TokenManager) Start(ctx context.Context) error {
	tm.mu.Lock()
	if tm.running {
		tm.mu.Unlock()
		return fmt.Errorf("token manager is already running")
	}
	tm.running = true
	tm.mu.Unlock()

	go tm.monitorTokens(ctx)
	return nil
}

// Stop stops the token manager
func (tm *TokenManager) Stop() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if tm.running {
		close(tm.stopChan)
		tm.running = false
	}
}

// monitorTokens periodically checks token status
func (tm *TokenManager) monitorTokens(ctx context.Context) {
	ticker := time.NewTicker(tm.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-tm.stopChan:
			return
		case <-ticker.C:
			tm.checkAllTokens(ctx)
		}
	}
}

// checkAllTokens validates all managed tokens
func (tm *TokenManager) checkAllTokens(ctx context.Context) {
	tm.mu.RLock()
	tokensCopy := make(map[string]*TokenInfo)
	for service, token := range tm.tokens {
		tokensCopy[service] = token
	}
	tm.mu.RUnlock()

	for _, tokenInfo := range tokensCopy {
		// Check token health
		if err := tm.validator.CheckTokenHealth(ctx, tokenInfo); err != nil {
			// Token validation failed
			continue
		}

		// Check expiration
		if tokenInfo.IsExpired(tm.expirationThreshold) {
			timeUntilExpiry := time.Until(*tokenInfo.ExpiresAt)
			if timeUntilExpiry > 0 {
				tm.handler.OnTokenExpiring(ctx, tokenInfo, timeUntilExpiry)
			} else {
				tm.handler.OnTokenExpired(ctx, tokenInfo)
			}
		}
	}
}

// GetTokenStatus returns status of all managed tokens
func (tm *TokenManager) GetTokenStatus() map[string]TokenStatus {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	status := make(map[string]TokenStatus)
	for service, tokenInfo := range tm.tokens {
		var timeUntilExpiry *time.Duration
		if tokenInfo.ExpiresAt != nil {
			duration := time.Until(*tokenInfo.ExpiresAt)
			timeUntilExpiry = &duration
		}

		status[service] = TokenStatus{
			Service:         service,
			IsValid:         tokenInfo.IsValid(),
			IsExpiring:      tokenInfo.IsExpired(tm.expirationThreshold),
			ExpiresAt:       tokenInfo.ExpiresAt,
			TimeUntilExpiry: timeUntilExpiry,
			TokenType:       tokenInfo.TokenType,
			Scopes:          tokenInfo.Scopes,
			LastValidAt:     tokenInfo.LastValidAt,
			CanRefresh:      tm.canRefreshToken(service, tokenInfo),
		}
	}

	return status
}

// TokenStatus represents the current status of a token
type TokenStatus struct {
	Service         string         `json:"service"`
	IsValid         bool           `json:"is_valid"`
	IsExpiring      bool           `json:"is_expiring"`
	ExpiresAt       *time.Time     `json:"expires_at,omitempty"`
	TimeUntilExpiry *time.Duration `json:"time_until_expiry,omitempty"`
	TokenType       string         `json:"token_type"`
	Scopes          []string       `json:"scopes,omitempty"`
	LastValidAt     time.Time      `json:"last_valid_at"`
	CanRefresh      bool           `json:"can_refresh"`
}

// canRefreshToken checks if a token can be refreshed
func (tm *TokenManager) canRefreshToken(service string, tokenInfo *TokenInfo) bool {
	refresher, exists := tm.refreshers[service]
	return exists && refresher.CanRefresh(tokenInfo)
}

// DefaultTokenValidator provides basic token validation
type DefaultTokenValidator struct {
	httpClient *http.Client
}

// NewDefaultTokenValidator creates a new default token validator
func NewDefaultTokenValidator() *DefaultTokenValidator {
	return &DefaultTokenValidator{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ValidateToken validates a token and returns its metadata
func (v *DefaultTokenValidator) ValidateToken(ctx context.Context, token, service string) (*TokenInfo, error) {
	tokenInfo := &TokenInfo{
		Token:       token,
		Service:     service,
		LastValidAt: time.Now(),
		TokenType:   "classic", // Default assumption
		Metadata:    make(map[string]interface{}),
	}

	switch service {
	case "github":
		return v.validateGitHubToken(ctx, tokenInfo)
	case "gitlab":
		return v.validateGitLabToken(ctx, tokenInfo)
	case "gitea":
		return v.validateGiteaToken(ctx, tokenInfo)
	default:
		return tokenInfo, nil // Basic validation for unknown services
	}
}

// CheckTokenHealth performs a health check on the token
func (v *DefaultTokenValidator) CheckTokenHealth(ctx context.Context, tokenInfo *TokenInfo) error {
	// Perform a lightweight API call to verify token is still valid
	switch tokenInfo.Service {
	case "github":
		return v.checkGitHubTokenHealth(ctx, tokenInfo)
	case "gitlab":
		return v.checkGitLabTokenHealth(ctx, tokenInfo)
	case "gitea":
		return v.checkGiteaTokenHealth(ctx, tokenInfo)
	default:
		return nil // Skip health check for unknown services
	}
}

// validateGitHubToken validates a GitHub token
func (v *DefaultTokenValidator) validateGitHubToken(ctx context.Context, tokenInfo *TokenInfo) (*TokenInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "token "+tokenInfo.Token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to validate GitHub token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub token validation failed with status %d", resp.StatusCode)
	}

	// Parse scopes from response headers
	if scopesHeader := resp.Header.Get("X-OAuth-Scopes"); scopesHeader != "" {
		tokenInfo.Scopes = strings.Split(strings.ReplaceAll(scopesHeader, " ", ""), ",")
	}

	// Check for token expiration in headers (for fine-grained tokens)
	if expiryHeader := resp.Header.Get("GitHub-Authentication-Token-Expiration"); expiryHeader != "" {
		if expiryTime, err := time.Parse(time.RFC3339, expiryHeader); err == nil {
			tokenInfo.ExpiresAt = &expiryTime
			tokenInfo.TokenType = "fine_grained"
		}
	}

	return tokenInfo, nil
}

// validateGitLabToken validates a GitLab token
func (v *DefaultTokenValidator) validateGitLabToken(ctx context.Context, tokenInfo *TokenInfo) (*TokenInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://gitlab.com/api/v4/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+tokenInfo.Token)

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to validate GitLab token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitLab token validation failed with status %d", resp.StatusCode)
	}

	// GitLab personal access tokens include expiration info in the token API
	// This would require a separate API call to /personal_access_tokens

	return tokenInfo, nil
}

// validateGiteaToken validates a Gitea token
func (v *DefaultTokenValidator) validateGiteaToken(ctx context.Context, tokenInfo *TokenInfo) (*TokenInfo, error) {
	// Gitea validation would require the base URL
	// For now, we'll do basic validation
	if len(tokenInfo.Token) < 10 {
		return nil, fmt.Errorf("Gitea token appears to be invalid (too short)")
	}

	return tokenInfo, nil
}

// checkGitHubTokenHealth performs a lightweight health check
func (v *DefaultTokenValidator) checkGitHubTokenHealth(ctx context.Context, tokenInfo *TokenInfo) error {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/rate_limit", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "token "+tokenInfo.Token)

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("GitHub token health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("GitHub token is no longer valid")
	}

	return nil
}

// checkGitLabTokenHealth performs a lightweight health check
func (v *DefaultTokenValidator) checkGitLabTokenHealth(ctx context.Context, tokenInfo *TokenInfo) error {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://gitlab.com/api/v4/version", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+tokenInfo.Token)

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("GitLab token health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("GitLab token is no longer valid")
	}

	return nil
}

// checkGiteaTokenHealth performs a lightweight health check
func (v *DefaultTokenValidator) checkGiteaTokenHealth(ctx context.Context, tokenInfo *TokenInfo) error {
	// Gitea health check would require the base URL
	// For now, we'll assume the token is healthy
	return nil
}
