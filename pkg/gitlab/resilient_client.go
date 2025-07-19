package gitlab

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// ResilientGitLabClient provides GitLab API operations with network resilience - DISABLED (recovery package removed)
// Simple HTTP client implementation to replace deleted recovery package.
type ResilientGitLabClient struct {
	httpClient HTTPClient
	baseURL    string
	token      string
}

// NewResilientGitLabClient creates a new resilient GitLab client - DISABLED (recovery package removed)
// Simple HTTP client implementation to replace deleted recovery package.
func NewResilientGitLabClient(baseURL, token string) *ResilientGitLabClient {
	return &ResilientGitLabClient{
		httpClient: NewHTTPClientAdapter(),
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		token:      token,
	}
}

// NewResilientGitLabClientWithConfig creates a resilient GitLab client with custom config - DISABLED (recovery package removed)
// Simple HTTP client implementation to replace deleted recovery package.
func NewResilientGitLabClientWithConfig(baseURL, token string, timeout time.Duration) *ResilientGitLabClient {
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &ResilientGitLabClient{
		httpClient: NewHTTPClientAdapterWithClient(&http.Client{
			Timeout: timeout,
		}),
		baseURL: strings.TrimSuffix(baseURL, "/"),
		token:   token,
	}
}

// prepareRequest adds authentication and headers to requests.
func (c *ResilientGitLabClient) prepareRequest(req *http.Request) {
	if c.token != "" {
		req.Header.Set("PRIVATE-TOKEN", c.token)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "gzh-manager-go")
}

// ListGroupProjects retrieves all projects for a GitLab group with pagination and resilience.
func (c *ResilientGitLabClient) ListGroupProjects(ctx context.Context, groupID string) ([]ProjectInfo, error) {
	var allProjects []ProjectInfo

	page := 1
	perPage := 100

	for {
		projects, hasMore, err := c.getProjectPage(ctx, groupID, page, perPage)
		if err != nil {
			return nil, err
		}

		allProjects = append(allProjects, projects...)

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

	return allProjects, nil
}

// getProjectPage fetches a single page of projects.
func (c *ResilientGitLabClient) getProjectPage(ctx context.Context, groupID string, page, perPage int) ([]ProjectInfo, bool, error) {
	url := fmt.Sprintf("%s/api/v4/groups/%s/projects?page=%d&per_page=%d&include_subgroups=true",
		c.baseURL, groupID, page, perPage)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, false, fmt.Errorf("failed to create request: %w", err)
	}

	c.prepareRequest(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("failed to get projects: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			// Log error but don't override main error
			_ = err
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, false, c.handleAPIError(resp, "failed to get projects")
	}

	var projects []ProjectInfo
	if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
		return nil, false, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check for more pages using pagination headers
	hasMore := c.hasNextPage(resp.Header)

	return projects, hasMore, nil
}

// hasNextPage checks if there are more pages based on pagination headers.
func (c *ResilientGitLabClient) hasNextPage(headers http.Header) bool {
	totalPages := headers.Get("X-Total-Pages")
	currentPage := headers.Get("X-Page")

	if totalPages == "" || currentPage == "" {
		return false
	}

	total, err1 := strconv.Atoi(totalPages)

	current, err2 := strconv.Atoi(currentPage)
	if err1 != nil || err2 != nil {
		return false
	}

	return current < total
}

// ProjectInfo represents GitLab project information.
type ProjectInfo struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	Path              string `json:"path"`
	PathWithNamespace string `json:"path_with_namespace"`
	HTTPURLToRepo     string `json:"http_url_to_repo"`
	SSHURLToRepo      string `json:"ssh_url_to_repo"`
	DefaultBranch     string `json:"default_branch"`
	Archived          bool   `json:"archived"`
	Visibility        string `json:"visibility"`
}

// GetProject retrieves detailed information about a specific project.
func (c *ResilientGitLabClient) GetProject(ctx context.Context, projectID string) (*ProjectInfo, error) {
	url := fmt.Sprintf("%s/api/v4/projects/%s", c.baseURL, projectID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.prepareRequest(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			// Log error but don't override main error
			_ = err
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleAPIError(resp, "failed to get project")
	}

	var project ProjectInfo
	if err := json.NewDecoder(resp.Body).Decode(&project); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &project, nil
}

// ListGroups retrieves all accessible groups.
func (c *ResilientGitLabClient) ListGroups(ctx context.Context) ([]APIGroupInfo, error) {
	var allGroups []APIGroupInfo

	page := 1
	perPage := 100

	for {
		groups, hasMore, err := c.getGroupPage(ctx, page, perPage)
		if err != nil {
			return nil, err
		}

		allGroups = append(allGroups, groups...)

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

	return allGroups, nil
}

// getGroupPage fetches a single page of groups.
func (c *ResilientGitLabClient) getGroupPage(ctx context.Context, page, perPage int) ([]APIGroupInfo, bool, error) {
	url := fmt.Sprintf("%s/api/v4/groups?page=%d&per_page=%d&owned=true", c.baseURL, page, perPage)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, false, fmt.Errorf("failed to create request: %w", err)
	}

	c.prepareRequest(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("failed to get groups: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			// Log error but don't override main error
			_ = err
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, false, c.handleAPIError(resp, "failed to get groups")
	}

	var groups []APIGroupInfo
	if err := json.NewDecoder(resp.Body).Decode(&groups); err != nil {
		return nil, false, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check for more pages using pagination headers
	hasMore := c.hasNextPage(resp.Header)

	return groups, hasMore, nil
}

// APIGroupInfo represents GitLab group information from API.
type APIGroupInfo struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	FullPath string `json:"full_path"`
	WebURL   string `json:"web_url"`
}

// handleAPIError creates appropriate error messages based on response status.
func (c *ResilientGitLabClient) handleAPIError(resp *http.Response, operation string) error {
	switch resp.StatusCode {
	case http.StatusNotFound:
		return fmt.Errorf("%s: not found (404)", operation)
	case http.StatusUnauthorized:
		return fmt.Errorf("%s: unauthorized - check your token (401)", operation)
	case http.StatusForbidden:
		return fmt.Errorf("%s: forbidden - insufficient permissions (403)", operation)
	case http.StatusTooManyRequests:
		return fmt.Errorf("%s: rate limited - please retry later (429)", operation)
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

// GetStats returns statistics about the underlying HTTP client - DISABLED (recovery package removed)
// Simple implementation without external recovery dependency.
func (c *ResilientGitLabClient) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"type": "standard_http_client",
		"note": "recovery package removed, using standard http.Client",
	}
}

// Close closes the underlying HTTP client connections - DISABLED (recovery package removed)
// Simple implementation without external recovery dependency.
func (c *ResilientGitLabClient) Close() {
	// Standard http.Client doesn't have Close method
	// No cleanup needed for standard client
}

// SetToken updates the authentication token.
func (c *ResilientGitLabClient) SetToken(token string) {
	c.token = token
}

// SetBaseURL updates the base URL.
func (c *ResilientGitLabClient) SetBaseURL(baseURL string) {
	c.baseURL = strings.TrimSuffix(baseURL, "/")
}
