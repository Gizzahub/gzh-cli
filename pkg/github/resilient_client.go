package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// ResilientGitHubClient provides GitHub API operations with network resilience - DISABLED (recovery package removed)
// Simple HTTP client implementation to replace deleted recovery package.
type ResilientGitHubClient struct {
	httpClient *http.Client
	baseURL    string
	token      string
}

// NewResilientGitHubClient creates a new resilient GitHub client - DISABLED (recovery package removed)
// Simple HTTP client implementation to replace deleted recovery package.
func NewResilientGitHubClient(token string) *ResilientGitHubClient {
	return NewResilientGitHubClientWithBaseURL(token, "")
}

// NewResilientGitHubClientWithBaseURL creates a resilient GitHub client with a custom API base URL.
// Empty baseURL falls back to DefaultGitHubAPIBaseURL (for GHES or tests).
func NewResilientGitHubClientWithBaseURL(token, baseURL string) *ResilientGitHubClient {
	return &ResilientGitHubClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: ResolveGitHubAPIBaseURL(baseURL),
		token:   token,
	}
}

// NewResilientGitHubClientWithConfig creates a resilient GitHub client with custom config - DISABLED (recovery package removed)
// Simple HTTP client implementation to replace deleted recovery package.
func NewResilientGitHubClientWithConfig(token string, timeout time.Duration) *ResilientGitHubClient {
	return NewResilientGitHubClientWithConfigAndBaseURL(token, timeout, "")
}

// NewResilientGitHubClientWithConfigAndBaseURL creates a resilient GitHub client with timeout and custom API base URL.
// Empty baseURL falls back to DefaultGitHubAPIBaseURL (for GHES or tests).
func NewResilientGitHubClientWithConfigAndBaseURL(token string, timeout time.Duration, baseURL string) *ResilientGitHubClient {
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &ResilientGitHubClient{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		baseURL: ResolveGitHubAPIBaseURL(baseURL),
		token:   token,
	}
}

// prepareRequest adds authentication and headers to requests.
func (c *ResilientGitHubClient) prepareRequest(req *http.Request) {
	if c.token != "" {
		req.Header.Set("Authorization", "token "+c.token)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "gzh-cli")
}

// GetDefaultBranch retrieves the default branch for a repository with network resilience.
func (c *ResilientGitHubClient) GetDefaultBranch(ctx context.Context, org, repo string) (string, error) {
	url := fmt.Sprintf("%s/repos/%s/%s", c.baseURL, org, repo)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	c.prepareRequest(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get repository info: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			// Log error but don't override the main error
			fmt.Printf("Warning: failed to close response body: %v\n", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", c.handleAPIError(resp, "failed to get repository info")
	}

	var repoInfo RepoInfo
	if err := json.NewDecoder(resp.Body).Decode(&repoInfo); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return repoInfo.DefaultBranch, nil
}

// ListRepositories retrieves all repositories for an organization with pagination and resilience.
func (c *ResilientGitHubClient) ListRepositories(ctx context.Context, org string) ([]string, error) {
	var allRepos []string

	page := 1
	perPage := 100

	for {
		repos, hasMore, err := c.getRepositoryPage(ctx, org, page, perPage)
		if err != nil {
			return nil, err
		}

		allRepos = append(allRepos, repos...)

		if !hasMore {
			break
		}

		page++

		// Check for context cancellation between pages
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
	}

	return allRepos, nil
}

// getRepositoryPage fetches a single page of repositories.
func (c *ResilientGitHubClient) getRepositoryPage(ctx context.Context, org string, page, perPage int) ([]string, bool, error) {
	url := fmt.Sprintf("%s/orgs/%s/repos?page=%d&per_page=%d", c.baseURL, org, page, perPage)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, false, fmt.Errorf("failed to create request: %w", err)
	}

	c.prepareRequest(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("failed to get repositories: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			// Log error but don't override the main error
			fmt.Printf("Warning: failed to close response body: %v\n", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, false, c.handleAPIError(resp, "failed to get repositories")
	}

	var repos []struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, false, fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract repository names
	names := make([]string, len(repos))
	for i, repo := range repos {
		names[i] = repo.Name
	}

	// Check for more pages using Link header
	hasMore := c.hasNextPage(resp.Header.Get("Link"))

	return names, hasMore, nil
}

// hasNextPage checks if there are more pages based on Link header.
func (c *ResilientGitHubClient) hasNextPage(linkHeader string) bool {
	return strings.Contains(linkHeader, `rel="next"`)
}

// handleAPIError creates appropriate error messages based on response status.
func (c *ResilientGitHubClient) handleAPIError(resp *http.Response, operation string) error {
	switch resp.StatusCode {
	case http.StatusNotFound:
		return fmt.Errorf("%s: not found (404)", operation)
	case http.StatusUnauthorized:
		return fmt.Errorf("%s: unauthorized - check your token (401)", operation)
	case http.StatusForbidden:
		return fmt.Errorf("%s: forbidden - insufficient permissions (403)", operation)
	case http.StatusTooManyRequests:
		// Extract rate limit reset time
		resetHeader := resp.Header.Get("X-RateLimit-Reset")
		if resetTime, err := strconv.ParseInt(resetHeader, 10, 64); err == nil {
			resetAt := time.Unix(resetTime, 0)
			waitTime := time.Until(resetAt)

			return fmt.Errorf("%s: rate limited - retry after %v (429)", operation, waitTime.Round(time.Second))
		}

		return fmt.Errorf("%s: rate limited (429)", operation)
	case http.StatusInternalServerError:
		return fmt.Errorf("%s: server error (500)", operation)
	case http.StatusBadGateway:
		return fmt.Errorf("%s: bad gateway (502)", operation)
	case http.StatusServiceUnavailable:
		return fmt.Errorf("%s: service unavailable (503)", operation)
	case http.StatusGatewayTimeout:
		return fmt.Errorf("%s: gateway timeout (504)", operation)
	default:
		return fmt.Errorf("%s: HTTP %d - %s", operation, resp.StatusCode, resp.Status)
	}
}

// GetRateLimit retrieves current rate limit status.
func (c *ResilientGitHubClient) GetRateLimit(ctx context.Context) (*RateLimitInfo, error) {
	url := fmt.Sprintf("%s/rate_limit", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.prepareRequest(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get rate limit: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			// Log error but don't override the main error
			fmt.Printf("Warning: failed to close response body: %v\n", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleAPIError(resp, "failed to get rate limit")
	}

	var rateLimitResponse struct {
		Rate struct {
			Limit     int   `json:"limit"`
			Remaining int   `json:"remaining"`
			Reset     int64 `json:"reset"`
		} `json:"rate"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&rateLimitResponse); err != nil {
		return nil, fmt.Errorf("failed to decode rate limit response: %w", err)
	}

	return &RateLimitInfo{
		Limit:     rateLimitResponse.Rate.Limit,
		Remaining: rateLimitResponse.Rate.Remaining,
		ResetTime: time.Unix(rateLimitResponse.Rate.Reset, 0),
	}, nil
}

// RateLimitInfo contains GitHub API rate limit information
// RateLimitInfo type is defined in token_aware_client.go to avoid duplication

// GetStats returns statistics about the underlying HTTP client - DISABLED (recovery package removed)
// Simple implementation without external recovery dependency.
func (c *ResilientGitHubClient) GetStats() map[string]any {
	return map[string]any{
		"type": "standard_http_client",
		"note": "recovery package removed, using standard http.Client",
		"config": map[string]any{
			"timeout": c.httpClient.Timeout,
			"baseURL": c.baseURL,
		},
	}
}

// Close closes the underlying HTTP client connections - DISABLED (recovery package removed)
// Simple implementation without external recovery dependency.
func (c *ResilientGitHubClient) Close() {
	// Standard http.Client doesn't have Close method
	// No cleanup needed for standard client
}

// SetToken updates the authentication token.
func (c *ResilientGitHubClient) SetToken(token string) {
	c.token = token
}

// SetBaseURL updates the base URL (useful for GitHub Enterprise).
// Empty baseURL falls back to DefaultGitHubAPIBaseURL.
func (c *ResilientGitHubClient) SetBaseURL(baseURL string) {
	c.baseURL = ResolveGitHubAPIBaseURL(baseURL)
}

// CreateRepository creates a repository under the given owner (org preferred, user fallback).
func (c *ResilientGitHubClient) CreateRepository(ctx context.Context, owner string, opts *CreateRepositoryOptions) (*RepositoryInfo, error) {
	if owner == "" {
		return nil, fmt.Errorf("owner is required")
	}
	if opts == nil || opts.Name == "" {
		return nil, fmt.Errorf("repository name is required")
	}

	// Prefer org endpoint; fall back to user endpoint when owner is a user namespace.
	info, err := c.doCreateRepository(ctx, fmt.Sprintf("%s/orgs/%s/repos", c.baseURL, owner), opts)
	if err == nil {
		return info, nil
	}
	// 404 from org endpoint typically means owner is not an organization.
	if !strings.Contains(err.Error(), "not found (404)") {
		return nil, err
	}
	return c.doCreateRepository(ctx, fmt.Sprintf("%s/user/repos", c.baseURL), opts)
}

func (c *ResilientGitHubClient) doCreateRepository(ctx context.Context, endpoint string, opts *CreateRepositoryOptions) (*RepositoryInfo, error) {
	body, err := json.Marshal(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to encode create repository request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	c.prepareRequest(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, c.handleAPIError(resp, "failed to create repository")
	}

	var info RepositoryInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("failed to decode create repository response: %w", err)
	}
	return &info, nil
}

// UpdateRepository patches repository settings (description, private, archived, etc.).
func (c *ResilientGitHubClient) UpdateRepository(ctx context.Context, owner, repo string, opts *UpdateRepositoryOptions) (*RepositoryInfo, error) {
	if owner == "" || repo == "" {
		return nil, fmt.Errorf("owner and repo are required")
	}
	if opts == nil {
		opts = &UpdateRepositoryOptions{}
	}

	body, err := json.Marshal(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to encode update repository request: %w", err)
	}

	endpoint := fmt.Sprintf("%s/repos/%s/%s", c.baseURL, owner, repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	c.prepareRequest(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to update repository: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleAPIError(resp, "failed to update repository")
	}

	var info RepositoryInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("failed to decode update repository response: %w", err)
	}
	return &info, nil
}

// ForkRepository forks a repository. Returns 202 Accepted on GitHub.
func (c *ResilientGitHubClient) ForkRepository(ctx context.Context, owner, repo string, opts *ForkRepositoryOptions) (*RepositoryInfo, error) {
	if owner == "" || repo == "" {
		return nil, fmt.Errorf("owner and repo are required")
	}
	if opts == nil {
		opts = &ForkRepositoryOptions{}
	}

	body, err := json.Marshal(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to encode fork repository request: %w", err)
	}

	endpoint := fmt.Sprintf("%s/repos/%s/%s/forks", c.baseURL, owner, repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	c.prepareRequest(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fork repository: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// GitHub returns 202 Accepted for forks; some proxies may return 201/200.
	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, c.handleAPIError(resp, "failed to fork repository")
	}

	var info RepositoryInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("failed to decode fork repository response: %w", err)
	}
	return &info, nil
}

// DeleteRepository deletes a repository identified by owner/repo.
func (c *ResilientGitHubClient) DeleteRepository(ctx context.Context, owner, repo string) error {
	if owner == "" || repo == "" {
		return fmt.Errorf("owner and repo are required")
	}

	url := fmt.Sprintf("%s/repos/%s/%s", c.baseURL, owner, repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	c.prepareRequest(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete repository: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return c.handleAPIError(resp, "failed to delete repository")
	}
	return nil
}

// ArchiveRepository archives a repository.
func (c *ResilientGitHubClient) ArchiveRepository(ctx context.Context, owner, repo string) error {
	return c.setRepositoryArchived(ctx, owner, repo, true)
}

// UnarchiveRepository unarchives a repository.
func (c *ResilientGitHubClient) UnarchiveRepository(ctx context.Context, owner, repo string) error {
	return c.setRepositoryArchived(ctx, owner, repo, false)
}

func (c *ResilientGitHubClient) setRepositoryArchived(ctx context.Context, owner, repo string, archived bool) error {
	if owner == "" || repo == "" {
		return fmt.Errorf("owner and repo are required")
	}

	body, err := json.Marshal(map[string]bool{"archived": archived})
	if err != nil {
		return fmt.Errorf("failed to encode archive request: %w", err)
	}

	endpoint := fmt.Sprintf("%s/repos/%s/%s", c.baseURL, owner, repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	c.prepareRequest(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to update repository archive state: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		op := "failed to unarchive repository"
		if archived {
			op = "failed to archive repository"
		}
		return c.handleAPIError(resp, op)
	}
	return nil
}

// SearchRepositories searches repositories via the GitHub search API.
func (c *ResilientGitHubClient) SearchRepositories(ctx context.Context, query string, opts *SearchRepositoriesOptions) (*RepositorySearchResult, error) {
	if query == "" {
		return nil, fmt.Errorf("search query is required")
	}
	if opts == nil {
		opts = &SearchRepositoriesOptions{}
	}

	page := opts.Page
	if page <= 0 {
		page = 1
	}
	perPage := opts.PerPage
	if perPage <= 0 {
		perPage = 30
	}

	params := url.Values{}
	params.Set("q", query)
	params.Set("page", strconv.Itoa(page))
	params.Set("per_page", strconv.Itoa(perPage))
	if opts.Sort != "" {
		params.Set("sort", opts.Sort)
	}
	if opts.Order != "" {
		params.Set("order", opts.Order)
	}

	endpoint := fmt.Sprintf("%s/search/repositories?%s", c.baseURL, params.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	c.prepareRequest(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to search repositories: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleAPIError(resp, "failed to search repositories")
	}

	var result RepositorySearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}
	return &result, nil
}
