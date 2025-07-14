package recovery

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// TokenExpirationIntegration provides high-level token expiration handling
type TokenExpirationIntegration struct {
	manager   *TokenManager
	registry  *TokenRefreshRegistry
	scheduler *TokenRefreshScheduler
	policy    TokenRefreshPolicy
}

// TokenExpirationConfig configures the token expiration integration
type TokenExpirationConfig struct {
	// Token manager configuration
	ExpirationThreshold time.Duration
	CheckInterval       time.Duration

	// Refresh policy
	RefreshPolicy TokenRefreshPolicy

	// OAuth2 configurations (optional)
	GitHubOAuth2 *OAuth2Config
	GitLabOAuth2 *OAuth2Config

	// Fallback tokens
	FallbackTokens map[string][]string // service -> fallback tokens

	// Custom handlers
	ExpirationHandler TokenExpirationHandler
	Validator         TokenValidator
}

// OAuth2Config contains OAuth2 client configuration
type OAuth2Config struct {
	ClientID     string
	ClientSecret string
	BaseURL      string // For GitLab self-hosted instances
}

// DefaultTokenExpirationConfig returns sensible defaults
func DefaultTokenExpirationConfig() TokenExpirationConfig {
	return TokenExpirationConfig{
		ExpirationThreshold: 24 * time.Hour,
		CheckInterval:       1 * time.Hour,
		RefreshPolicy:       DefaultTokenRefreshPolicy(),
		ExpirationHandler:   NewDefaultTokenExpirationHandler(),
		Validator:           NewDefaultTokenValidator(),
		FallbackTokens:      make(map[string][]string),
	}
}

// NewTokenExpirationIntegration creates a new token expiration integration
func NewTokenExpirationIntegration(config TokenExpirationConfig) (*TokenExpirationIntegration, error) {
	// Create token manager
	tmConfig := TokenManagerConfig{
		ExpirationThreshold: config.ExpirationThreshold,
		CheckInterval:       config.CheckInterval,
		Handler:             config.ExpirationHandler,
		Validator:           config.Validator,
	}
	manager := NewTokenManager(tmConfig)

	// Create refresh registry
	registry := NewTokenRefreshRegistry()

	// Add OAuth2 refreshers if configured
	if config.GitHubOAuth2 != nil {
		githubRefresher := NewGitHubTokenRefresher(
			config.GitHubOAuth2.ClientID,
			config.GitHubOAuth2.ClientSecret,
		)
		registry.AddRefresher("github", githubRefresher)
	}

	if config.GitLabOAuth2 != nil {
		gitlabRefresher := NewGitLabTokenRefresher(
			config.GitLabOAuth2.ClientID,
			config.GitLabOAuth2.ClientSecret,
			config.GitLabOAuth2.BaseURL,
		)
		registry.AddRefresher("gitlab", gitlabRefresher)
	}

	// Add fallback refreshers
	for service, fallbackTokens := range config.FallbackTokens {
		if len(fallbackTokens) > 0 {
			fallbackRefresher := NewPersonalAccessTokenRefresher(fallbackTokens, config.Validator)
			registry.AddRefresher(service, fallbackRefresher)
		}
	}

	// Create scheduler
	scheduler := NewTokenRefreshScheduler(manager, registry, config.RefreshPolicy)

	return &TokenExpirationIntegration{
		manager:   manager,
		registry:  registry,
		scheduler: scheduler,
		policy:    config.RefreshPolicy,
	}, nil
}

// Start initializes and starts the token expiration handling
func (tei *TokenExpirationIntegration) Start(ctx context.Context) error {
	if err := tei.manager.Start(ctx); err != nil {
		return fmt.Errorf("failed to start token manager: %w", err)
	}

	if err := tei.scheduler.Start(ctx); err != nil {
		return fmt.Errorf("failed to start refresh scheduler: %w", err)
	}

	return nil
}

// Stop shuts down the token expiration handling
func (tei *TokenExpirationIntegration) Stop() {
	tei.scheduler.Stop()
	tei.manager.Stop()
}

// AddToken adds a token to be managed
func (tei *TokenExpirationIntegration) AddToken(service, token string) error {
	return tei.manager.AddToken(service, token)
}

// GetValidToken retrieves a valid token, refreshing if necessary
func (tei *TokenExpirationIntegration) GetValidToken(service string) (string, error) {
	tokenInfo, err := tei.manager.GetToken(service)
	if err != nil {
		return "", err
	}

	return tokenInfo.Token, nil
}

// GetTokenInfo retrieves detailed token information
func (tei *TokenExpirationIntegration) GetTokenInfo(service string) (*TokenInfo, error) {
	return tei.manager.GetToken(service)
}

// RefreshToken manually refreshes a token
func (tei *TokenExpirationIntegration) RefreshToken(ctx context.Context, service string) error {
	tokenInfo, err := tei.manager.GetToken(service)
	if err != nil {
		return fmt.Errorf("failed to get token for %s: %w", service, err)
	}

	newTokenInfo, err := tei.registry.RefreshToken(ctx, tokenInfo)
	if err != nil {
		return fmt.Errorf("failed to refresh token for %s: %w", service, err)
	}

	return tei.manager.AddToken(service, newTokenInfo.Token)
}

// GetTokenStatus returns the status of all managed tokens
func (tei *TokenExpirationIntegration) GetTokenStatus() map[string]TokenStatus {
	return tei.manager.GetTokenStatus()
}

// AddFallbackToken adds a fallback token for a service
func (tei *TokenExpirationIntegration) AddFallbackToken(service, token string) {
	fallbackRefresher := NewPersonalAccessTokenRefresher([]string{token}, tei.manager.validator)
	tei.registry.AddRefresher(service, fallbackRefresher)
}

// TokenAwareHTTPClient wraps an HTTP client with automatic token refresh
type TokenAwareHTTPClient struct {
	client      *ResilientHTTPClient
	integration *TokenExpirationIntegration
	service     string
}

// NewTokenAwareHTTPClient creates an HTTP client with token expiration handling
func NewTokenAwareHTTPClient(client *ResilientHTTPClient, integration *TokenExpirationIntegration, service string) *TokenAwareHTTPClient {
	return &TokenAwareHTTPClient{
		client:      client,
		integration: integration,
		service:     service,
	}
}

// Do executes an HTTP request with automatic token refresh on expiration
func (tc *TokenAwareHTTPClient) Do(req *http.Request) (*http.Response, error) {
	// Get current valid token
	token, err := tc.integration.GetValidToken(tc.service)
	if err != nil {
		return nil, fmt.Errorf("failed to get valid token: %w", err)
	}

	// Set authorization header
	tc.setAuthHeader(req, token)

	// Execute request
	resp, err := tc.client.Do(req)

	// Check if token expired during request
	if err == nil && (resp.StatusCode == 401 || resp.StatusCode == 403) {
		// Try to refresh token and retry
		refreshCtx, cancel := context.WithTimeout(req.Context(), 30*time.Second)
		defer cancel()

		if refreshErr := tc.integration.RefreshToken(refreshCtx, tc.service); refreshErr == nil {
			// Get new token and retry
			newToken, tokenErr := tc.integration.GetValidToken(tc.service)
			if tokenErr == nil {
				// Close old response
				if resp.Body != nil {
					resp.Body.Close()
				}

				// Create new request with fresh token
				retryReq := req.Clone(req.Context())
				tc.setAuthHeader(retryReq, newToken)
				return tc.client.Do(retryReq)
			}
		}
	}

	return resp, err
}

// setAuthHeader sets the appropriate authorization header for the service
func (tc *TokenAwareHTTPClient) setAuthHeader(req *http.Request, token string) {
	switch tc.service {
	case "github":
		req.Header.Set("Authorization", "token "+token)
	case "gitlab":
		req.Header.Set("Authorization", "Bearer "+token)
	case "gitea":
		req.Header.Set("Authorization", "token "+token)
	default:
		req.Header.Set("Authorization", "Bearer "+token)
	}
}

// GetWithContext performs a GET request with token awareness
func (tc *TokenAwareHTTPClient) GetWithContext(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	return tc.Do(req)
}

// TokenExpirationAwareFactory creates token-aware HTTP clients
type TokenExpirationAwareFactory struct {
	integration *TokenExpirationIntegration
	httpFactory *HTTPClientFactory
}

// NewTokenExpirationAwareFactory creates a new token-aware factory
func NewTokenExpirationAwareFactory(integration *TokenExpirationIntegration, httpFactory *HTTPClientFactory) *TokenExpirationAwareFactory {
	return &TokenExpirationAwareFactory{
		integration: integration,
		httpFactory: httpFactory,
	}
}

// CreateGitHubClient creates a token-aware GitHub client
func (tef *TokenExpirationAwareFactory) CreateGitHubClient() *TokenAwareHTTPClient {
	resilientClient := tef.httpFactory.CreateGitHubClient()
	return NewTokenAwareHTTPClient(resilientClient, tef.integration, "github")
}

// CreateGitLabClient creates a token-aware GitLab client
func (tef *TokenExpirationAwareFactory) CreateGitLabClient() *TokenAwareHTTPClient {
	resilientClient := tef.httpFactory.CreateGitLabClient()
	return NewTokenAwareHTTPClient(resilientClient, tef.integration, "gitlab")
}

// CreateGiteaClient creates a token-aware Gitea client
func (tef *TokenExpirationAwareFactory) CreateGiteaClient() *TokenAwareHTTPClient {
	resilientClient := tef.httpFactory.CreateGiteaClient()
	return NewTokenAwareHTTPClient(resilientClient, tef.integration, "gitea")
}

// ExpirationNotifier handles token expiration notifications
type ExpirationNotifier struct {
	notificationChannels []chan TokenExpirationEvent
}

// TokenExpirationEvent represents a token expiration event
type TokenExpirationEvent struct {
	Service         string        `json:"service"`
	EventType       string        `json:"event_type"` // "expiring", "expired", "refreshed"
	TokenInfo       *TokenInfo    `json:"token_info"`
	TimeUntilExpiry time.Duration `json:"time_until_expiry,omitempty"`
	Timestamp       time.Time     `json:"timestamp"`
	Message         string        `json:"message"`
}

// NewExpirationNotifier creates a new expiration notifier
func NewExpirationNotifier() *ExpirationNotifier {
	return &ExpirationNotifier{
		notificationChannels: make([]chan TokenExpirationEvent, 0),
	}
}

// Subscribe adds a notification channel
func (en *ExpirationNotifier) Subscribe() <-chan TokenExpirationEvent {
	ch := make(chan TokenExpirationEvent, 10)
	en.notificationChannels = append(en.notificationChannels, ch)
	return ch
}

// Notify sends a notification to all subscribers
func (en *ExpirationNotifier) Notify(event TokenExpirationEvent) {
	event.Timestamp = time.Now()

	for _, ch := range en.notificationChannels {
		select {
		case ch <- event:
		default:
			// Channel is full, skip this notification
		}
	}
}

// OnTokenExpiring implements TokenExpirationHandler
func (en *ExpirationNotifier) OnTokenExpiring(ctx context.Context, tokenInfo *TokenInfo, timeUntilExpiry time.Duration) error {
	event := TokenExpirationEvent{
		Service:         tokenInfo.Service,
		EventType:       "expiring",
		TokenInfo:       tokenInfo,
		TimeUntilExpiry: timeUntilExpiry,
		Message:         fmt.Sprintf("Token for %s expires in %v", tokenInfo.Service, timeUntilExpiry),
	}
	en.Notify(event)
	return nil
}

// OnTokenExpired implements TokenExpirationHandler
func (en *ExpirationNotifier) OnTokenExpired(ctx context.Context, tokenInfo *TokenInfo) error {
	event := TokenExpirationEvent{
		Service:   tokenInfo.Service,
		EventType: "expired",
		TokenInfo: tokenInfo,
		Message:   fmt.Sprintf("Token for %s has expired", tokenInfo.Service),
	}
	en.Notify(event)
	return fmt.Errorf("token for %s has expired", tokenInfo.Service)
}

// OnTokenRefreshed implements TokenExpirationHandler
func (en *ExpirationNotifier) OnTokenRefreshed(ctx context.Context, oldToken, newToken *TokenInfo) error {
	event := TokenExpirationEvent{
		Service:   newToken.Service,
		EventType: "refreshed",
		TokenInfo: newToken,
		Message:   fmt.Sprintf("Token for %s has been refreshed", newToken.Service),
	}
	en.Notify(event)
	return nil
}
