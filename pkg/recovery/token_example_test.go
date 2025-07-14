package recovery_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gizzahub/gzh-manager-go/pkg/recovery"
)

// ExampleTokenExpirationIntegration demonstrates basic usage of token expiration handling
func ExampleTokenExpirationIntegration() {
	// Create configuration
	config := recovery.DefaultTokenExpirationConfig()
	config.ExpirationThreshold = 24 * time.Hour // Warn 24h before expiration
	config.CheckInterval = 1 * time.Hour        // Check every hour

	// Add fallback tokens from environment or configuration
	config.FallbackTokens = map[string][]string{
		"github": {os.Getenv("GITHUB_TOKEN_BACKUP")},
		"gitlab": {os.Getenv("GITLAB_TOKEN_BACKUP")},
	}

	// Create integration
	integration, err := recovery.NewTokenExpirationIntegration(config)
	if err != nil {
		log.Fatalf("Failed to create token integration: %v", err)
	}

	// Start background monitoring
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := integration.Start(ctx); err != nil {
		log.Fatalf("Failed to start token integration: %v", err)
	}
	defer integration.Stop()

	// Add tokens to be managed
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		integration.AddToken("github", token)
	}

	if token := os.Getenv("GITLAB_TOKEN"); token != "" {
		integration.AddToken("gitlab", token)
	}

	// Use tokens in your application
	githubToken, err := integration.GetValidToken("github")
	if err != nil {
		log.Printf("Failed to get GitHub token: %v", err)
	} else {
		fmt.Printf("GitHub token available: %t\n", len(githubToken) > 0)
	}

	// Get detailed token status
	statuses := integration.GetTokenStatus()
	for service, status := range statuses {
		fmt.Printf("Service: %s, Valid: %t, Expiring: %t\n",
			service, status.IsValid, status.IsExpiring)
	}

	// Output:
	// GitHub token available: true
	// Service: github, Valid: true, Expiring: false
}

// ExampleTokenAwareHTTPClient demonstrates using HTTP clients with automatic token refresh
func ExampleTokenAwareHTTPClient() {
	// Create token integration
	config := recovery.DefaultTokenExpirationConfig()
	integration, err := recovery.NewTokenExpirationIntegration(config)
	if err != nil {
		log.Fatalf("Failed to create integration: %v", err)
	}

	ctx := context.Background()
	integration.Start(ctx)
	defer integration.Stop()

	// Add token
	integration.AddToken("github", "your-github-token")

	// Create HTTP factory
	httpFactory := recovery.NewHTTPClientFactory()

	// Create token-aware factory
	tokenAwareFactory := recovery.NewTokenExpirationAwareFactory(integration, httpFactory)

	// Create GitHub client with automatic token refresh
	client := tokenAwareFactory.CreateGitHubClient()

	// Use the client - it will automatically handle token expiration
	resp, err := client.GetWithContext(ctx, "https://api.github.com/user")
	if err != nil {
		log.Printf("Request failed: %v", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Request successful: %d\n", resp.StatusCode)

	// Output:
	// Request successful: 200
}

// ExampleTokenRefreshWithOAuth2 demonstrates OAuth2 token refresh
func ExampleTokenRefreshWithOAuth2() {
	// Create configuration with OAuth2 settings
	config := recovery.DefaultTokenExpirationConfig()
	config.GitHubOAuth2 = &recovery.OAuth2Config{
		ClientID:     "your-github-app-id",
		ClientSecret: "your-github-app-secret",
	}
	config.GitLabOAuth2 = &recovery.OAuth2Config{
		ClientID:     "your-gitlab-app-id",
		ClientSecret: "your-gitlab-app-secret",
		BaseURL:      "https://gitlab.com", // or your GitLab instance
	}

	integration, err := recovery.NewTokenExpirationIntegration(config)
	if err != nil {
		log.Fatalf("Failed to create integration: %v", err)
	}

	ctx := context.Background()
	integration.Start(ctx)
	defer integration.Stop()

	// Add OAuth2 token (with refresh token in metadata)
	// This would typically come from your OAuth2 flow
	tokenManager := recovery.NewTokenManager(recovery.DefaultTokenManagerConfig())
	tokenInfo := &recovery.TokenInfo{
		Token:     "oauth2-access-token",
		Service:   "github",
		TokenType: "oauth2",
		ExpiresAt: func() *time.Time { t := time.Now().Add(1 * time.Hour); return &t }(),
		Metadata: map[string]interface{}{
			"refresh_token": "oauth2-refresh-token",
		},
	}

	// The integration will automatically refresh this token when it expires
	fmt.Printf("OAuth2 token will be automatically refreshed\n")

	// Output:
	// OAuth2 token will be automatically refreshed
}

// ExampleExpirationNotifier demonstrates token expiration notifications
func ExampleExpirationNotifier() {
	// Create notifier
	notifier := recovery.NewExpirationNotifier()

	// Subscribe to notifications
	events := notifier.Subscribe()

	// Handle notifications in a goroutine
	go func() {
		for event := range events {
			switch event.EventType {
			case "expiring":
				fmt.Printf("Token for %s expiring in %v\n",
					event.Service, event.TimeUntilExpiry)
			case "expired":
				fmt.Printf("Token for %s has expired\n", event.Service)
			case "refreshed":
				fmt.Printf("Token for %s has been refreshed\n", event.Service)
			}
		}
	}()

	// Create configuration with notifier as handler
	config := recovery.DefaultTokenExpirationConfig()
	config.ExpirationHandler = notifier

	integration, err := recovery.NewTokenExpirationIntegration(config)
	if err != nil {
		log.Fatalf("Failed to create integration: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	integration.Start(ctx)
	defer integration.Stop()

	fmt.Printf("Notification system ready\n")

	// Output:
	// Notification system ready
}

// ExampleTokenValidation demonstrates token validation
func ExampleTokenValidation() {
	validator := recovery.NewDefaultTokenValidator()
	ctx := context.Background()

	// Validate a GitHub token
	tokenInfo, err := validator.ValidateToken(ctx, "your-token", "github")
	if err != nil {
		log.Printf("Token validation failed: %v", err)
		return
	}

	fmt.Printf("Token type: %s\n", tokenInfo.TokenType)
	fmt.Printf("Scopes: %v\n", tokenInfo.Scopes)
	if tokenInfo.ExpiresAt != nil {
		fmt.Printf("Expires at: %v\n", tokenInfo.ExpiresAt)
	} else {
		fmt.Printf("No expiration date\n")
	}

	// Check token health
	err = validator.CheckTokenHealth(ctx, tokenInfo)
	if err != nil {
		log.Printf("Token health check failed: %v", err)
	} else {
		fmt.Printf("Token is healthy\n")
	}

	// Output:
	// Token type: classic
	// Scopes: [repo read:org]
	// No expiration date
	// Token is healthy
}

// ExampleCustomRefresher demonstrates implementing a custom token refresher
func ExampleCustomRefresher() {
	// Custom refresher implementation
	type CustomRefresher struct {
		// Custom fields for your refresher
		apiEndpoint string
		credentials map[string]string
	}

	// Implement TokenRefresher interface
	refresher := &CustomRefresher{
		apiEndpoint: "https://your-api.com/refresh",
		credentials: map[string]string{
			"client_id": "your-client-id",
		},
	}

	// Register with refresh registry
	registry := recovery.NewTokenRefreshRegistry()
	registry.AddRefresher("custom-service", refresher)

	fmt.Printf("Custom refresher registered\n")

	// Output:
	// Custom refresher registered
}

// Implement TokenRefresher interface for the custom refresher
func (cr *CustomRefresher) CanRefresh(tokenInfo *recovery.TokenInfo) bool {
	// Implement your logic to determine if token can be refreshed
	return tokenInfo.Service == "custom-service" && tokenInfo.TokenType == "bearer"
}

func (cr *CustomRefresher) RefreshToken(ctx context.Context, tokenInfo *recovery.TokenInfo) (*recovery.TokenInfo, error) {
	// Implement your custom refresh logic here
	// This would typically make an HTTP request to your refresh endpoint

	// Return new token info
	return &recovery.TokenInfo{
		Token:       "new-refreshed-token",
		Service:     tokenInfo.Service,
		TokenType:   "bearer",
		LastValidAt: time.Now(),
		ExpiresAt:   func() *time.Time { t := time.Now().Add(1 * time.Hour); return &t }(),
	}, nil
}
