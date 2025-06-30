package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// RepoConfigClient provides GitHub API operations for repository configuration management
type RepoConfigClient struct {
	token       string
	baseURL     string
	httpClient  *http.Client
	rateLimiter *RateLimiter
}

// Repository represents a GitHub repository with configuration details
type Repository struct {
	ID            int64    `json:"id"`
	Name          string   `json:"name"`
	FullName      string   `json:"full_name"`
	Description   string   `json:"description"`
	Private       bool     `json:"private"`
	HTMLURL       string   `json:"html_url"`
	CloneURL      string   `json:"clone_url"`
	SSHURL        string   `json:"ssh_url"`
	DefaultBranch string   `json:"default_branch"`
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
	Language      string   `json:"language"`
	Topics        []string `json:"topics"`

	// Repository settings
	HasIssues    bool `json:"has_issues"`
	HasProjects  bool `json:"has_projects"`
	HasWiki      bool `json:"has_wiki"`
	HasDownloads bool `json:"has_downloads"`

	// Security and collaboration settings
	AllowSquashMerge    bool `json:"allow_squash_merge"`
	AllowMergeCommit    bool `json:"allow_merge_commit"`
	AllowRebaseMerge    bool `json:"allow_rebase_merge"`
	DeleteBranchOnMerge bool `json:"delete_branch_on_merge"`
}

// BranchProtection represents branch protection rule configuration
type BranchProtection struct {
	RequiredStatusChecks       *RequiredStatusChecks       `json:"required_status_checks,omitempty"`
	EnforceAdmins              bool                        `json:"enforce_admins"`
	RequiredPullRequestReviews *RequiredPullRequestReviews `json:"required_pull_request_reviews,omitempty"`
	Restrictions               *BranchRestrictions         `json:"restrictions,omitempty"`
}

// RequiredStatusChecks represents required status checks configuration
type RequiredStatusChecks struct {
	Strict   bool     `json:"strict"`
	Contexts []string `json:"contexts"`
}

// RequiredPullRequestReviews represents PR review requirements
type RequiredPullRequestReviews struct {
	DismissStaleReviews          bool                  `json:"dismiss_stale_reviews"`
	RequireCodeOwnerReviews      bool                  `json:"require_code_owner_reviews"`
	RequiredApprovingReviewCount int                   `json:"required_approving_review_count"`
	DismissalRestrictions        *UserTeamRestrictions `json:"dismissal_restrictions,omitempty"`
}

// BranchRestrictions represents branch push restrictions
type BranchRestrictions struct {
	Users []string `json:"users"`
	Teams []string `json:"teams"`
}

// UserTeamRestrictions represents user/team restrictions
type UserTeamRestrictions struct {
	Users []string `json:"users"`
	Teams []string `json:"teams"`
}

// RepositoryUpdate represents fields that can be updated in a repository
type RepositoryUpdate struct {
	Name                *string  `json:"name,omitempty"`
	Description         *string  `json:"description,omitempty"`
	Homepage            *string  `json:"homepage,omitempty"`
	Private             *bool    `json:"private,omitempty"`
	HasIssues           *bool    `json:"has_issues,omitempty"`
	HasProjects         *bool    `json:"has_projects,omitempty"`
	HasWiki             *bool    `json:"has_wiki,omitempty"`
	HasDownloads        *bool    `json:"has_downloads,omitempty"`
	DefaultBranch       *string  `json:"default_branch,omitempty"`
	AllowSquashMerge    *bool    `json:"allow_squash_merge,omitempty"`
	AllowMergeCommit    *bool    `json:"allow_merge_commit,omitempty"`
	AllowRebaseMerge    *bool    `json:"allow_rebase_merge,omitempty"`
	DeleteBranchOnMerge *bool    `json:"delete_branch_on_merge,omitempty"`
	Topics              []string `json:"topics,omitempty"`
}

// APIError represents a GitHub API error response
type APIError struct {
	Message          string `json:"message"`
	DocumentationURL string `json:"documentation_url"`
	StatusCode       int    `json:"-"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("GitHub API error (%d): %s", e.StatusCode, e.Message)
}

// NewRepoConfigClient creates a new GitHub API client for repository configuration
func NewRepoConfigClient(token string) *RepoConfigClient {
	return &RepoConfigClient{
		token:   token,
		baseURL: "https://api.github.com",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		rateLimiter: NewRateLimiter(),
	}
}

// SetTimeout configures the HTTP client timeout
func (c *RepoConfigClient) SetTimeout(timeout time.Duration) {
	c.httpClient.Timeout = timeout
}

// makeRequest performs an HTTP request with authentication, rate limiting, and retry logic
func (c *RepoConfigClient) makeRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	retries := 0
	maxRetries := 3

	for retries <= maxRetries {
		// Wait for rate limit if necessary
		if err := c.rateLimiter.Wait(ctx); err != nil {
			return nil, fmt.Errorf("rate limit wait failed: %w", err)
		}

		url := c.baseURL + path

		var bodyReader io.Reader
		if body != nil {
			jsonBody, err := json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal request body: %w", err)
			}
			bodyReader = bytes.NewReader(jsonBody)
		}

		req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		// Set headers
		req.Header.Set("Accept", "application/vnd.github.v3+json")
		req.Header.Set("User-Agent", "gzh-manager-go/1.0")
		if c.token != "" {
			req.Header.Set("Authorization", "token "+c.token)
		}
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			// Network errors are not retryable
			return nil, fmt.Errorf("request failed: %w", err)
		}

		// Update rate limit information
		c.rateLimiter.Update(resp)

		// Check if we should retry
		if ShouldRetry(resp) && retries < maxRetries {
			// Close the response body before retry
			_ = resp.Body.Close()

			// Calculate backoff
			backoff := CalculateBackoff(retries)

			// Use Retry-After if available and longer than backoff
			if retryAfter := c.rateLimiter.retryAfter; retryAfter > backoff {
				backoff = retryAfter
			}

			// Wait before retry
			if err := sleep(ctx, backoff); err != nil {
				return nil, err
			}

			retries++
			continue
		}

		// Handle API errors
		if resp.StatusCode >= 400 {
			defer resp.Body.Close()
			var apiError APIError
			if err := json.NewDecoder(resp.Body).Decode(&apiError); err != nil {
				return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, resp.Status)
			}
			apiError.StatusCode = resp.StatusCode
			return nil, &apiError
		}

		return resp, nil
	}

	return nil, &RetryableError{
		Err:          fmt.Errorf("max retries exceeded after %d attempts", retries),
		AttemptsLeft: 0,
	}
}

// GetRateLimitStatus returns current rate limit status
func (c *RepoConfigClient) GetRateLimitStatus() (int, int, time.Time) {
	return c.rateLimiter.GetStatus()
}

// ListRepositories lists all repositories for an organization with pagination
func (c *RepoConfigClient) ListRepositories(ctx context.Context, org string, options *ListOptions) ([]*Repository, error) {
	if options == nil {
		options = &ListOptions{PerPage: 30}
	}

	var allRepos []*Repository
	page := 1

	for {
		path := fmt.Sprintf("/orgs/%s/repos?per_page=%d&page=%d", org, options.PerPage, page)
		if options.Type != "" {
			path += "&type=" + options.Type
		}
		if options.Sort != "" {
			path += "&sort=" + options.Sort
		}
		if options.Direction != "" {
			path += "&direction=" + options.Direction
		}

		resp, err := c.makeRequest(ctx, "GET", path, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to list repositories: %w", err)
		}
		defer resp.Body.Close()

		var repos []*Repository
		if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
			return nil, fmt.Errorf("failed to decode repositories: %w", err)
		}

		allRepos = append(allRepos, repos...)

		// Check if there are more pages
		if len(repos) < options.PerPage {
			break
		}
		page++
	}

	return allRepos, nil
}

// GetRepository gets a specific repository
func (c *RepoConfigClient) GetRepository(ctx context.Context, owner, repo string) (*Repository, error) {
	path := fmt.Sprintf("/repos/%s/%s", owner, repo)

	resp, err := c.makeRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}
	defer resp.Body.Close()

	var repository Repository
	if err := json.NewDecoder(resp.Body).Decode(&repository); err != nil {
		return nil, fmt.Errorf("failed to decode repository: %w", err)
	}

	return &repository, nil
}

// UpdateRepository updates repository settings
func (c *RepoConfigClient) UpdateRepository(ctx context.Context, owner, repo string, update *RepositoryUpdate) (*Repository, error) {
	path := fmt.Sprintf("/repos/%s/%s", owner, repo)

	resp, err := c.makeRequest(ctx, "PATCH", path, update)
	if err != nil {
		return nil, fmt.Errorf("failed to update repository: %w", err)
	}
	defer resp.Body.Close()

	var repository Repository
	if err := json.NewDecoder(resp.Body).Decode(&repository); err != nil {
		return nil, fmt.Errorf("failed to decode updated repository: %w", err)
	}

	return &repository, nil
}

// GetBranchProtection gets branch protection rules for a specific branch
func (c *RepoConfigClient) GetBranchProtection(ctx context.Context, owner, repo, branch string) (*BranchProtection, error) {
	path := fmt.Sprintf("/repos/%s/%s/branches/%s/protection", owner, repo, branch)

	resp, err := c.makeRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get branch protection: %w", err)
	}
	defer resp.Body.Close()

	var protection BranchProtection
	if err := json.NewDecoder(resp.Body).Decode(&protection); err != nil {
		return nil, fmt.Errorf("failed to decode branch protection: %w", err)
	}

	return &protection, nil
}

// UpdateBranchProtection updates branch protection rules
func (c *RepoConfigClient) UpdateBranchProtection(ctx context.Context, owner, repo, branch string, protection *BranchProtection) (*BranchProtection, error) {
	path := fmt.Sprintf("/repos/%s/%s/branches/%s/protection", owner, repo, branch)

	resp, err := c.makeRequest(ctx, "PUT", path, protection)
	if err != nil {
		return nil, fmt.Errorf("failed to update branch protection: %w", err)
	}
	defer resp.Body.Close()

	var updatedProtection BranchProtection
	if err := json.NewDecoder(resp.Body).Decode(&updatedProtection); err != nil {
		return nil, fmt.Errorf("failed to decode updated branch protection: %w", err)
	}

	return &updatedProtection, nil
}

// DeleteBranchProtection removes branch protection rules
func (c *RepoConfigClient) DeleteBranchProtection(ctx context.Context, owner, repo, branch string) error {
	path := fmt.Sprintf("/repos/%s/%s/branches/%s/protection", owner, repo, branch)

	resp, err := c.makeRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return fmt.Errorf("failed to delete branch protection: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

// ListOptions represents options for listing operations
type ListOptions struct {
	PerPage   int    // Number of items per page (default: 30, max: 100)
	Type      string // Repository type: all, owner, member
	Sort      string // Sort by: created, updated, pushed, full_name
	Direction string // Sort direction: asc, desc
}
