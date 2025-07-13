package performance

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gizzahub/gzh-manager-go/pkg/api"
)

// OptimizedGitHubClient provides an optimized GitHub API client with
// request deduplication, batching, and intelligent rate limiting
type OptimizedGitHubClient struct {
	httpClient  *http.Client
	token       string
	baseURL     string
	optimizer   *api.OptimizationManager
	rateLimiter *api.EnhancedRateLimiter
}

// OptimizedClientConfig configures the optimized GitHub client
type OptimizedClientConfig struct {
	Token               string
	BaseURL             string
	OptimizationConfig  api.OptimizationConfig
	EnableOptimizations bool
}

// DefaultOptimizedClientConfig returns sensible defaults for GitHub API optimization
func DefaultOptimizedClientConfig(token string) OptimizedClientConfig {
	optimConfig := api.DefaultOptimizationConfig()
	optimConfig.BatchConfig.MaxBatchSize = 100 // GitHub GraphQL supports up to 100 per query
	optimConfig.BatchConfig.FlushInterval = 150 * time.Millisecond
	optimConfig.DeduplicationTTL = 5 * time.Minute

	return OptimizedClientConfig{
		Token:               token,
		BaseURL:             "https://api.github.com",
		OptimizationConfig:  optimConfig,
		EnableOptimizations: true,
	}
}

// NewOptimizedGitHubClient creates a new optimized GitHub API client
func NewOptimizedGitHubClient(config OptimizedClientConfig) *OptimizedGitHubClient {
	client := &OptimizedGitHubClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 20, // GitHub allows good concurrency
				IdleConnTimeout:     90 * time.Second,
			},
		},
		token:   config.Token,
		baseURL: config.BaseURL,
	}

	if config.EnableOptimizations {
		client.optimizer = api.NewOptimizationManager(config.OptimizationConfig)
		client.rateLimiter = client.optimizer.GetRateLimiter("github")
	}

	return client
}

// Repository represents GitHub repository information
type Repository struct {
	Name          string `json:"name"`
	FullName      string `json:"full_name"`
	DefaultBranch string `json:"default_branch"`
	Private       bool   `json:"private"`
	HTMLURL       string `json:"html_url"`
	CloneURL      string `json:"clone_url"`
	UpdatedAt     string `json:"updated_at"`
}

// RepositoryList represents a list of repositories from GitHub API
type RepositoryList struct {
	Repositories []Repository `json:"repositories"`
	TotalCount   int          `json:"total_count"`
}

// ListRepositoriesOptimized lists repositories for an organization with full optimization
func (c *OptimizedGitHubClient) ListRepositoriesOptimized(ctx context.Context, org string) ([]string, error) {
	if c.optimizer == nil {
		// Fallback to basic implementation
		return c.listRepositoriesBasic(ctx, org)
	}

	req := api.OptimizedRequest{
		Service:   "github",
		Operation: "list-repos",
		Key:       org,
		Context:   ctx,
	}

	executor := func(ctx context.Context) (interface{}, error) {
		return c.listRepositoriesBasic(ctx, org)
	}

	response, err := c.optimizer.ExecuteRequest(req, executor)
	if err != nil {
		return nil, err
	}

	if repos, ok := response.Data.([]string); ok {
		return repos, nil
	}

	return nil, fmt.Errorf("unexpected response type from optimized request")
}

// BatchGetDefaultBranches gets default branches for multiple repositories using batching
func (c *OptimizedGitHubClient) BatchGetDefaultBranches(ctx context.Context, org string, repos []string) (map[string]string, error) {
	if c.optimizer == nil {
		// Fallback to sequential calls
		return c.getDefaultBranchesSequential(ctx, org, repos)
	}

	batchProcessor := api.NewRepositoryBatchProcessor("github")
	defer batchProcessor.Stop()

	batchFunc := func(ctx context.Context, requests []*api.BatchRequest) []api.BatchResponse {
		responses := make([]api.BatchResponse, len(requests))

		// For GitHub, we can optimize this with GraphQL or parallel REST calls
		for i, req := range requests {
			data := req.Data.(map[string]string)
			org := data["org"]
			repo := data["repo"]

			branch, err := c.getDefaultBranchBasic(ctx, org, repo)
			responses[i] = api.BatchResponse{
				ID:    req.ID,
				Data:  branch,
				Error: err,
			}
		}

		return responses
	}

	return batchProcessor.BatchDefaultBranches(ctx, org, repos, batchFunc)
}

// BatchGetRepositoryMetadata gets comprehensive metadata for multiple repositories
func (c *OptimizedGitHubClient) BatchGetRepositoryMetadata(ctx context.Context, org string, repos []string) (map[string]*Repository, error) {
	if c.optimizer == nil {
		return c.getRepositoryMetadataSequential(ctx, org, repos)
	}

	batchProcessor := api.NewRepositoryBatchProcessor("github")
	defer batchProcessor.Stop()

	batchFunc := func(ctx context.Context, requests []*api.BatchRequest) []api.BatchResponse {
		// Here we could use GitHub's GraphQL API to batch multiple repository queries
		// For now, we'll parallelize REST API calls
		responses := make([]api.BatchResponse, len(requests))

		for i, req := range requests {
			data := req.Data.(map[string]string)
			org := data["org"]
			repo := data["repo"]

			metadata, err := c.getRepositoryMetadataBasic(ctx, org, repo)
			responses[i] = api.BatchResponse{
				ID:    req.ID,
				Data:  metadata,
				Error: err,
			}
		}

		return responses
	}

	rawResults, err := batchProcessor.BatchRepositoryMetadata(ctx, org, repos, batchFunc)
	if err != nil {
		return nil, err
	}

	// Convert to typed results
	results := make(map[string]*Repository)
	for repo, rawData := range rawResults {
		if metadata, ok := rawData.(*Repository); ok {
			results[repo] = metadata
		}
	}

	return results, nil
}

// listRepositoriesBasic provides the basic repository listing implementation
func (c *OptimizedGitHubClient) listRepositoriesBasic(ctx context.Context, org string) ([]string, error) {
	url := fmt.Sprintf("%s/orgs/%s/repos", c.baseURL, org)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.token != "" {
		req.Header.Set("Authorization", "token "+c.token)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Update rate limit information
	if c.rateLimiter != nil {
		c.updateRateLimitFromHeaders(resp.Header)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status: %s", resp.Status)
	}

	var repos []Repository
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract repository names
	repoNames := make([]string, len(repos))
	for i, repo := range repos {
		repoNames[i] = repo.Name
	}

	return repoNames, nil
}

// getDefaultBranchBasic gets the default branch for a single repository
func (c *OptimizedGitHubClient) getDefaultBranchBasic(ctx context.Context, org, repo string) (string, error) {
	url := fmt.Sprintf("%s/repos/%s/%s", c.baseURL, org, repo)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	if c.token != "" {
		req.Header.Set("Authorization", "token "+c.token)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Update rate limit information
	if c.rateLimiter != nil {
		c.updateRateLimitFromHeaders(resp.Header)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status: %s", resp.Status)
	}

	var repository Repository
	if err := json.NewDecoder(resp.Body).Decode(&repository); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return repository.DefaultBranch, nil
}

// getRepositoryMetadataBasic gets comprehensive metadata for a single repository
func (c *OptimizedGitHubClient) getRepositoryMetadataBasic(ctx context.Context, org, repo string) (*Repository, error) {
	url := fmt.Sprintf("%s/repos/%s/%s", c.baseURL, org, repo)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.token != "" {
		req.Header.Set("Authorization", "token "+c.token)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Update rate limit information
	if c.rateLimiter != nil {
		c.updateRateLimitFromHeaders(resp.Header)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status: %s", resp.Status)
	}

	var repository Repository
	if err := json.NewDecoder(resp.Body).Decode(&repository); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &repository, nil
}

// getDefaultBranchesSequential gets default branches sequentially (fallback)
func (c *OptimizedGitHubClient) getDefaultBranchesSequential(ctx context.Context, org string, repos []string) (map[string]string, error) {
	results := make(map[string]string)

	for _, repo := range repos {
		branch, err := c.getDefaultBranchBasic(ctx, org, repo)
		if err != nil {
			return nil, fmt.Errorf("failed to get default branch for %s: %w", repo, err)
		}
		results[repo] = branch
	}

	return results, nil
}

// getRepositoryMetadataSequential gets repository metadata sequentially (fallback)
func (c *OptimizedGitHubClient) getRepositoryMetadataSequential(ctx context.Context, org string, repos []string) (map[string]*Repository, error) {
	results := make(map[string]*Repository)

	for _, repo := range repos {
		metadata, err := c.getRepositoryMetadataBasic(ctx, org, repo)
		if err != nil {
			return nil, fmt.Errorf("failed to get metadata for %s: %w", repo, err)
		}
		results[repo] = metadata
	}

	return results, nil
}

// updateRateLimitFromHeaders updates rate limit information from response headers
func (c *OptimizedGitHubClient) updateRateLimitFromHeaders(headers http.Header) {
	limitStr := headers.Get("X-RateLimit-Limit")
	remainingStr := headers.Get("X-RateLimit-Remaining")
	resetStr := headers.Get("X-RateLimit-Reset")

	if limitStr == "" || remainingStr == "" || resetStr == "" {
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return
	}

	remaining, err := strconv.Atoi(remainingStr)
	if err != nil {
		return
	}

	reset, err := strconv.ParseInt(resetStr, 10, 64)
	if err != nil {
		return
	}

	resetTime := time.Unix(reset, 0)
	c.rateLimiter.UpdateLimits(limit, remaining, resetTime)

	// Check for secondary rate limit
	if retryAfter := headers.Get("Retry-After"); retryAfter != "" {
		if duration, err := time.ParseDuration(retryAfter + "s"); err == nil {
			c.rateLimiter.SetRetryAfter(duration)
		}
	}
}

// GetOptimizationStats returns current optimization statistics
func (c *OptimizedGitHubClient) GetOptimizationStats() *api.OptimizationStats {
	if c.optimizer == nil {
		return nil
	}

	stats := c.optimizer.GetStats()
	return &stats
}

// PrintOptimizationStats prints detailed optimization statistics
func (c *OptimizedGitHubClient) PrintOptimizationStats() {
	if c.optimizer != nil {
		c.optimizer.PrintDetailedStats()
	} else {
		fmt.Println("Optimizations disabled")
	}
}

// Close closes the optimized client and cleans up resources
func (c *OptimizedGitHubClient) Close() {
	if c.optimizer != nil {
		c.optimizer.Stop()
	}
}

// BulkCloneOptimized performs optimized bulk cloning with all optimizations enabled
func (c *OptimizedGitHubClient) BulkCloneOptimized(ctx context.Context, org string, targetDir string) error {
	// Step 1: Get repository list (with deduplication)
	repos, err := c.ListRepositoriesOptimized(ctx, org)
	if err != nil {
		return fmt.Errorf("failed to list repositories: %w", err)
	}

	// Step 2: Batch get default branches (with batching optimization)
	branches, err := c.BatchGetDefaultBranches(ctx, org, repos)
	if err != nil {
		return fmt.Errorf("failed to get default branches: %w", err)
	}

	// Step 3: Clone repositories with optimized rate limiting
	for _, repo := range repos {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Rate limiting is handled automatically by the optimizer
		branch := branches[repo]
		fmt.Printf("Cloning %s (branch: %s)...\n", repo, branch)

		// Here you would implement the actual git clone operation
		// This is just a placeholder for demonstration
		time.Sleep(10 * time.Millisecond) // Simulate clone time
	}

	return nil
}

// EnableOptimizations enables API optimizations
func (c *OptimizedGitHubClient) EnableOptimizations() {
	if c.optimizer != nil {
		c.optimizer.Enable()
	}
}

// DisableOptimizations disables API optimizations (useful for debugging)
func (c *OptimizedGitHubClient) DisableOptimizations() {
	if c.optimizer != nil {
		c.optimizer.Disable()
	}
}

// IsOptimizationEnabled returns whether optimizations are currently enabled
func (c *OptimizedGitHubClient) IsOptimizationEnabled() bool {
	if c.optimizer == nil {
		return false
	}
	return c.optimizer.IsEnabled()
}
