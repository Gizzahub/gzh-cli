package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gizzahub/gzh-manager-go/pkg/recovery"
)

// TokenAwareGitHubClient provides GitHub API operations with automatic token expiration handling
type TokenAwareGitHubClient struct {
	httpClient       *recovery.TokenAwareHTTPClient
	tokenIntegration *recovery.TokenExpirationIntegration
	baseURL          string
	fallbackTokens   []string
}

// TokenAwareGitHubClientConfig configures the token-aware GitHub client
type TokenAwareGitHubClientConfig struct {
	BaseURL        string
	PrimaryToken   string
	FallbackTokens []string
	OAuth2Config   *recovery.OAuth2Config

	// HTTP client configuration
	HTTPConfig recovery.ResilientHTTPClientConfig

	// Token expiration configuration
	ExpirationConfig recovery.TokenExpirationConfig
}

// DefaultTokenAwareGitHubClientConfig returns sensible defaults
func DefaultTokenAwareGitHubClientConfig() TokenAwareGitHubClientConfig {
	return TokenAwareGitHubClientConfig{
		BaseURL:          "https://api.github.com",
		HTTPConfig:       recovery.DefaultResilientHTTPClientConfig(),
		ExpirationConfig: recovery.DefaultTokenExpirationConfig(),
	}
}

// NewTokenAwareGitHubClient creates a new token-aware GitHub client
func NewTokenAwareGitHubClient(config TokenAwareGitHubClientConfig) (*TokenAwareGitHubClient, error) {
	// Configure token expiration integration
	expConfig := config.ExpirationConfig
	expConfig.FallbackTokens = map[string][]string{
		"github": config.FallbackTokens,
	}

	if config.OAuth2Config != nil {
		expConfig.GitHubOAuth2 = config.OAuth2Config
	}

	// Create token integration
	integration, err := recovery.NewTokenExpirationIntegration(expConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create token integration: %w", err)
	}

	// Add primary token
	if config.PrimaryToken != "" {
		if err := integration.AddToken("github", config.PrimaryToken); err != nil {
			return nil, fmt.Errorf("failed to add primary token: %w", err)
		}
	}

	// Create HTTP factory and token-aware client
	httpFactory := recovery.NewHTTPClientFactoryWithConfig(config.HTTPConfig)
	tokenAwareFactory := recovery.NewTokenExpirationAwareFactory(integration, httpFactory)
	httpClient := tokenAwareFactory.CreateGitHubClient()

	return &TokenAwareGitHubClient{
		httpClient:       httpClient,
		tokenIntegration: integration,
		baseURL:          config.BaseURL,
		fallbackTokens:   config.FallbackTokens,
	}, nil
}

// Start initializes the token expiration monitoring
func (c *TokenAwareGitHubClient) Start(ctx context.Context) error {
	return c.tokenIntegration.Start(ctx)
}

// Stop shuts down the token expiration monitoring
func (c *TokenAwareGitHubClient) Stop() {
	c.tokenIntegration.Stop()
}

// GetCurrentToken returns the current valid token
func (c *TokenAwareGitHubClient) GetCurrentToken() (string, error) {
	return c.tokenIntegration.GetValidToken("github")
}

// GetTokenStatus returns detailed token status information
func (c *TokenAwareGitHubClient) GetTokenStatus() (recovery.TokenStatus, error) {
	statuses := c.tokenIntegration.GetTokenStatus()
	if status, exists := statuses["github"]; exists {
		return status, nil
	}
	return recovery.TokenStatus{}, fmt.Errorf("no GitHub token found")
}

// RefreshToken manually refreshes the GitHub token
func (c *TokenAwareGitHubClient) RefreshToken(ctx context.Context) error {
	return c.tokenIntegration.RefreshToken(ctx, "github")
}

// GetUser retrieves the authenticated user information
func (c *TokenAwareGitHubClient) GetUser(ctx context.Context) (*GitHubUser, error) {
	url := fmt.Sprintf("%s/user", c.baseURL)

	resp, err := c.httpClient.GetWithContext(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp, "get user")
	}

	var user GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode user response: %w", err)
	}

	return &user, nil
}

// GetOrganization retrieves organization information
func (c *TokenAwareGitHubClient) GetOrganization(ctx context.Context, org string) (*GitHubOrganization, error) {
	url := fmt.Sprintf("%s/orgs/%s", c.baseURL, org)

	resp, err := c.httpClient.GetWithContext(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization %s: %w", org, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp, fmt.Sprintf("get organization %s", org))
	}

	var organization GitHubOrganization
	if err := json.NewDecoder(resp.Body).Decode(&organization); err != nil {
		return nil, fmt.Errorf("failed to decode organization response: %w", err)
	}

	return &organization, nil
}

// ListRepositories retrieves repositories for a user or organization
func (c *TokenAwareGitHubClient) ListRepositories(ctx context.Context, owner string, page, perPage int) ([]*GitHubRepository, error) {
	var url string

	// Determine if it's a user or organization
	user, err := c.GetUser(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	if user.Login == owner {
		// User's own repositories
		url = fmt.Sprintf("%s/user/repos", c.baseURL)
	} else {
		// Organization or other user repositories
		url = fmt.Sprintf("%s/users/%s/repos", c.baseURL)
	}

	// Add pagination parameters
	url += fmt.Sprintf("?page=%d&per_page=%d&sort=updated&direction=desc", page, perPage)

	resp, err := c.httpClient.GetWithContext(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to list repositories for %s: %w", owner, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp, fmt.Sprintf("list repositories for %s", owner))
	}

	var repositories []*GitHubRepository
	if err := json.NewDecoder(resp.Body).Decode(&repositories); err != nil {
		return nil, fmt.Errorf("failed to decode repositories response: %w", err)
	}

	return repositories, nil
}

// GetRepository retrieves specific repository information
func (c *TokenAwareGitHubClient) GetRepository(ctx context.Context, owner, repo string) (*GitHubRepository, error) {
	url := fmt.Sprintf("%s/repos/%s/%s", c.baseURL, owner, repo)

	resp, err := c.httpClient.GetWithContext(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository %s/%s: %w", owner, repo, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp, fmt.Sprintf("get repository %s/%s", owner, repo))
	}

	var repository GitHubRepository
	if err := json.NewDecoder(resp.Body).Decode(&repository); err != nil {
		return nil, fmt.Errorf("failed to decode repository response: %w", err)
	}

	return &repository, nil
}

// GetDefaultBranch retrieves the default branch for a repository
func (c *TokenAwareGitHubClient) GetDefaultBranch(ctx context.Context, owner, repo string) (string, error) {
	repository, err := c.GetRepository(ctx, owner, repo)
	if err != nil {
		return "", err
	}

	if repository.DefaultBranch == "" {
		return "main", nil // Default fallback
	}

	return repository.DefaultBranch, nil
}

// GetRateLimit retrieves current rate limit information
func (c *TokenAwareGitHubClient) GetRateLimit(ctx context.Context) (*RateLimitInfo, error) {
	url := fmt.Sprintf("%s/rate_limit", c.baseURL)

	resp, err := c.httpClient.GetWithContext(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to get rate limit: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp, "get rate limit")
	}

	var rateLimitResponse struct {
		Resources struct {
			Core struct {
				Limit     int   `json:"limit"`
				Remaining int   `json:"remaining"`
				Reset     int64 `json:"reset"`
				Used      int   `json:"used"`
			} `json:"core"`
		} `json:"resources"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&rateLimitResponse); err != nil {
		return nil, fmt.Errorf("failed to decode rate limit response: %w", err)
	}

	resetTime := time.Unix(rateLimitResponse.Resources.Core.Reset, 0)

	return &RateLimitInfo{
		Limit:     rateLimitResponse.Resources.Core.Limit,
		Remaining: rateLimitResponse.Resources.Core.Remaining,
		Reset:     resetTime,
		Used:      rateLimitResponse.Resources.Core.Used,
	}, nil
}

// ValidateTokenPermissions validates token permissions for specific operations
func (c *TokenAwareGitHubClient) ValidateTokenPermissions(ctx context.Context, requiredScopes []string) error {
	url := fmt.Sprintf("%s/user", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create validation request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to validate token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.handleErrorResponse(resp, "validate token")
	}

	// Check scopes in response header
	scopesHeader := resp.Header.Get("X-OAuth-Scopes")
	if scopesHeader == "" {
		return fmt.Errorf("no scopes found in token response")
	}

	availableScopes := strings.Split(strings.ReplaceAll(scopesHeader, " ", ""), ",")

	// Check if all required scopes are available
	for _, requiredScope := range requiredScopes {
		found := false
		for _, availableScope := range availableScopes {
			if availableScope == requiredScope {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("missing required scope: %s", requiredScope)
		}
	}

	return nil
}

// handleErrorResponse creates appropriate errors from HTTP responses
func (c *TokenAwareGitHubClient) handleErrorResponse(resp *http.Response, operation string) error {
	var errorMsg string

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		errorMsg = "unauthorized - token may be expired or invalid"
	case http.StatusForbidden:
		errorMsg = "forbidden - insufficient permissions or rate limited"

		// Check for rate limit headers
		if reset := resp.Header.Get("X-RateLimit-Reset"); reset != "" {
			if resetTime, err := strconv.ParseInt(reset, 10, 64); err == nil {
				resetTimeFormatted := time.Unix(resetTime, 0).Format(time.RFC3339)
				errorMsg += fmt.Sprintf(" (rate limit resets at %s)", resetTimeFormatted)
			}
		}
	case http.StatusNotFound:
		errorMsg = "not found - repository or resource does not exist"
	case http.StatusUnprocessableEntity:
		errorMsg = "unprocessable entity - invalid request data"
	default:
		errorMsg = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	return fmt.Errorf("GitHub API error during %s: %s", operation, errorMsg)
}

// GitHubUser represents a GitHub user
type GitHubUser struct {
	ID          int       `json:"id"`
	Login       string    `json:"login"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Company     string    `json:"company"`
	Location    string    `json:"location"`
	Bio         string    `json:"bio"`
	PublicRepos int       `json:"public_repos"`
	Followers   int       `json:"followers"`
	Following   int       `json:"following"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// GitHubOrganization represents a GitHub organization
type GitHubOrganization struct {
	ID          int       `json:"id"`
	Login       string    `json:"login"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Company     string    `json:"company"`
	Location    string    `json:"location"`
	Email       string    `json:"email"`
	PublicRepos int       `json:"public_repos"`
	Followers   int       `json:"followers"`
	Following   int       `json:"following"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// GitHubRepository represents a GitHub repository
type GitHubRepository struct {
	ID              int       `json:"id"`
	Name            string    `json:"name"`
	FullName        string    `json:"full_name"`
	Description     string    `json:"description"`
	Private         bool      `json:"private"`
	Fork            bool      `json:"fork"`
	Archived        bool      `json:"archived"`
	Disabled        bool      `json:"disabled"`
	DefaultBranch   string    `json:"default_branch"`
	Language        string    `json:"language"`
	Size            int       `json:"size"`
	StargazersCount int       `json:"stargazers_count"`
	WatchersCount   int       `json:"watchers_count"`
	ForksCount      int       `json:"forks_count"`
	OpenIssuesCount int       `json:"open_issues_count"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	PushedAt        time.Time `json:"pushed_at"`
	CloneURL        string    `json:"clone_url"`
	SSHURL          string    `json:"ssh_url"`
	HTMLURL         string    `json:"html_url"`
	GitURL          string    `json:"git_url"`
}

// RateLimitInfo represents GitHub rate limit information
type RateLimitInfo struct {
	Limit     int       `json:"limit"`
	Remaining int       `json:"remaining"`
	Reset     time.Time `json:"reset"`
	Used      int       `json:"used"`
}
