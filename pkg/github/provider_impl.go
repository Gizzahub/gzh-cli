// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package github

import (
	"context"
	"fmt"
	"time"

	"github.com/gizzahub/gzh-manager-go/pkg/git/provider"
)

// GitHubProvider implements the unified GitProvider interface for GitHub.
type GitHubProvider struct {
	client  APIClient
	cloner  CloneService
	baseURL string
	name    string
}

// Ensure GitHubProvider implements GitProvider interface
var _ provider.GitProvider = (*GitHubProvider)(nil)

// NewGitHubProvider creates a new GitHub provider instance.
func NewGitHubProvider(client APIClient, cloner CloneService) *GitHubProvider {
	return &GitHubProvider{
		client:  client,
		cloner:  cloner,
		baseURL: "https://api.github.com",
		name:    "github",
	}
}

// GetName returns the provider name.
func (g *GitHubProvider) GetName() string {
	return g.name
}

// GetCapabilities returns the list of supported capabilities.
func (g *GitHubProvider) GetCapabilities() []provider.Capability {
	return []provider.Capability{
		provider.CapabilityRepositories,
		provider.CapabilityWebhooks,
		provider.CapabilityEvents,
		provider.CapabilityIssues,
		provider.CapabilityPullRequests,
		provider.CapabilityWiki,
		provider.CapabilityProjects,
		provider.CapabilityActions,
		provider.CapabilityCICD,
		provider.CapabilityPackages,
		provider.CapabilityReleases,
		provider.CapabilityOrganizations,
		provider.CapabilityUsers,
		provider.CapabilityTeams,
		provider.CapabilityPermissions,
		provider.CapabilityBranchProtection,
		provider.CapabilitySecurityAlerts,
		provider.CapabilityDependabot,
	}
}

// GetBaseURL returns the base URL for the GitHub API.
func (g *GitHubProvider) GetBaseURL() string {
	return g.baseURL
}

// Authenticate sets up authentication credentials.
func (g *GitHubProvider) Authenticate(ctx context.Context, creds provider.Credentials) error {
	switch creds.Type {
	case provider.CredentialTypeToken:
		return g.client.SetToken(ctx, creds.Token)
	default:
		return fmt.Errorf("unsupported credential type: %s", creds.Type)
	}
}

// ValidateToken validates the authentication token.
func (g *GitHubProvider) ValidateToken(ctx context.Context) (*provider.TokenInfo, error) {
	rateLimit, err := g.client.GetRateLimit(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to validate token: %w", err)
	}

	return &provider.TokenInfo{
		Valid:  true,
		Scopes: []string{}, // GitHub doesn't expose scopes via rate limit endpoint
		User:   "",         // Would need additional API call
		Email:  "",         // Would need additional API call
		RateLimit: provider.RateLimit{
			Limit:     rateLimit.Limit,
			Remaining: rateLimit.Remaining,
			Reset:     rateLimit.Reset,
			Used:      rateLimit.Used,
		},
	}, nil
}

// ListRepositories lists repositories for an organization.
func (g *GitHubProvider) ListRepositories(ctx context.Context, opts provider.ListOptions) (*provider.RepositoryList, error) {
	owner := opts.Organization
	if owner == "" {
		owner = opts.User
	}
	if owner == "" {
		return nil, fmt.Errorf("either Organization or User must be specified in ListOptions")
	}

	repos, err := g.client.ListOrganizationRepositories(ctx, owner)
	if err != nil {
		return nil, err
	}

	repositories := make([]provider.Repository, 0, len(repos))
	for _, repo := range repos {
		repositories = append(repositories, provider.Repository{
			ID:            repo.FullName,
			Name:          repo.Name,
			FullName:      repo.FullName,
			Description:   repo.Description,
			DefaultBranch: repo.DefaultBranch,
			CloneURL:      repo.CloneURL,
			SSHURL:        repo.SSHURL,
			HTMLURL:       repo.HTMLURL,
			Private:       repo.Private,
			Archived:      repo.Archived,
			CreatedAt:     repo.CreatedAt,
			UpdatedAt:     repo.UpdatedAt,
			Language:      repo.Language,
			Size:          int64(repo.Size),
			Topics:        repo.Topics,
		})
	}

	return &provider.RepositoryList{
		Repositories: repositories,
		TotalCount:   len(repositories),
	}, nil
}

// GetRepository retrieves information about a specific repository.
func (g *GitHubProvider) GetRepository(ctx context.Context, id string) (*provider.Repository, error) {
	// Parse owner/repo from id
	owner, repo, err := parseFullName(id)
	if err != nil {
		return nil, err
	}

	repoInfo, err := g.client.GetRepository(ctx, owner, repo)
	if err != nil {
		return nil, err
	}

	return &provider.Repository{
		ID:            repoInfo.FullName,
		Name:          repoInfo.Name,
		FullName:      repoInfo.FullName,
		Description:   repoInfo.Description,
		DefaultBranch: repoInfo.DefaultBranch,
		CloneURL:      repoInfo.CloneURL,
		SSHURL:        repoInfo.SSHURL,
		HTMLURL:       repoInfo.HTMLURL,
		Private:       repoInfo.Private,
		Archived:      repoInfo.Archived,
		CreatedAt:     repoInfo.CreatedAt,
		UpdatedAt:     repoInfo.UpdatedAt,
		Language:      repoInfo.Language,
		Size:          int64(repoInfo.Size),
		Topics:        repoInfo.Topics,
	}, nil
}

// CreateRepository creates a new repository.
func (g *GitHubProvider) CreateRepository(ctx context.Context, req provider.CreateRepoRequest) (*provider.Repository, error) {
	// GitHub doesn't have a direct create repo API in the current interface
	// This would need to be implemented with the GitHub API client
	return nil, fmt.Errorf("create repository not implemented")
}

// UpdateRepository updates repository settings.
func (g *GitHubProvider) UpdateRepository(ctx context.Context, id string, updates provider.UpdateRepoRequest) (*provider.Repository, error) {
	// This would need to be implemented with the GitHub API client
	return nil, fmt.Errorf("update repository not implemented")
}

// DeleteRepository deletes a repository.
func (g *GitHubProvider) DeleteRepository(ctx context.Context, id string) error {
	// This would need to be implemented with the GitHub API client
	return fmt.Errorf("delete repository not implemented")
}

// ArchiveRepository archives a repository.
func (g *GitHubProvider) ArchiveRepository(ctx context.Context, id string) error {
	return fmt.Errorf("archive repository not implemented")
}

// UnarchiveRepository unarchives a repository.
func (g *GitHubProvider) UnarchiveRepository(ctx context.Context, id string) error {
	return fmt.Errorf("unarchive repository not implemented")
}

// CloneRepository clones a repository to the target path.
func (g *GitHubProvider) CloneRepository(ctx context.Context, repo provider.Repository, target string, opts provider.CloneOptions) error {
	// Convert to GitHub RepositoryInfo
	repoInfo := RepositoryInfo{
		Name:          repo.Name,
		FullName:      repo.FullName,
		Description:   repo.Description,
		DefaultBranch: repo.DefaultBranch,
		CloneURL:      repo.CloneURL,
		SSHURL:        repo.SSHURL,
		HTMLURL:       repo.HTMLURL,
		Private:       repo.Private,
		Archived:      repo.Archived,
		CreatedAt:     repo.CreatedAt,
		UpdatedAt:     repo.UpdatedAt,
		Language:      repo.Language,
		Size:          int(repo.Size),
		Topics:        repo.Topics,
	}

	return g.cloner.CloneRepository(ctx, repoInfo, target, opts.Strategy)
}

// ForkRepository creates a fork of a repository.
func (g *GitHubProvider) ForkRepository(ctx context.Context, id string, opts provider.ForkOptions) (*provider.Repository, error) {
	return nil, fmt.Errorf("fork repository not implemented")
}

// SearchRepositories searches for repositories.
func (g *GitHubProvider) SearchRepositories(ctx context.Context, query provider.SearchQuery) (*provider.SearchResult, error) {
	return nil, fmt.Errorf("search repositories not implemented")
}

// Webhook management methods (placeholder implementations)
func (g *GitHubProvider) ListWebhooks(ctx context.Context, repoID string) ([]provider.Webhook, error) {
	return nil, fmt.Errorf("webhook management not implemented")
}

func (g *GitHubProvider) GetWebhook(ctx context.Context, repoID, webhookID string) (*provider.Webhook, error) {
	return nil, fmt.Errorf("webhook management not implemented")
}

func (g *GitHubProvider) CreateWebhook(ctx context.Context, repoID string, webhook provider.CreateWebhookRequest) (*provider.Webhook, error) {
	return nil, fmt.Errorf("webhook management not implemented")
}

func (g *GitHubProvider) UpdateWebhook(ctx context.Context, repoID, webhookID string, updates provider.UpdateWebhookRequest) (*provider.Webhook, error) {
	return nil, fmt.Errorf("webhook management not implemented")
}

func (g *GitHubProvider) DeleteWebhook(ctx context.Context, repoID, webhookID string) error {
	return fmt.Errorf("webhook management not implemented")
}

func (g *GitHubProvider) TestWebhook(ctx context.Context, repoID, webhookID string) (*provider.WebhookTestResult, error) {
	return nil, fmt.Errorf("webhook management not implemented")
}

func (g *GitHubProvider) ValidateWebhookURL(ctx context.Context, url string) error {
	return fmt.Errorf("webhook management not implemented")
}

// Event management methods (placeholder implementations)
func (g *GitHubProvider) ListEvents(ctx context.Context, opts provider.EventListOptions) ([]provider.Event, error) {
	return nil, fmt.Errorf("event management not implemented")
}

func (g *GitHubProvider) GetEvent(ctx context.Context, eventID string) (*provider.Event, error) {
	return nil, fmt.Errorf("event management not implemented")
}

func (g *GitHubProvider) ProcessEvent(ctx context.Context, event provider.Event) error {
	return fmt.Errorf("event management not implemented")
}

func (g *GitHubProvider) RegisterEventHandler(eventType string, handler provider.EventHandler) error {
	return fmt.Errorf("event management not implemented")
}

func (g *GitHubProvider) StreamEvents(ctx context.Context, opts provider.StreamOptions) (<-chan provider.Event, error) {
	return nil, fmt.Errorf("event streaming not implemented")
}

// Health and monitoring methods
func (g *GitHubProvider) HealthCheck(ctx context.Context) (*provider.HealthStatus, error) {
	startTime := time.Now()

	// Use token validation as health check
	_, err := g.ValidateToken(ctx)
	latency := time.Since(startTime)

	status := &provider.HealthStatus{
		LastChecked: time.Now(),
		Latency:     latency,
		Details:     make(map[string]interface{}),
	}

	if err != nil {
		status.Status = provider.HealthStatusUnhealthy
		status.Message = err.Error()
	} else {
		status.Status = provider.HealthStatusHealthy
		status.Message = "GitHub API accessible"
	}

	return status, nil
}

func (g *GitHubProvider) GetRateLimit(ctx context.Context) (*provider.RateLimit, error) {
	rateLimit, err := g.client.GetRateLimit(ctx)
	if err != nil {
		return nil, err
	}

	return &provider.RateLimit{
		Limit:     rateLimit.Limit,
		Remaining: rateLimit.Remaining,
		Reset:     rateLimit.Reset,
		Used:      rateLimit.Used,
		Resource:  "core",
	}, nil
}

func (g *GitHubProvider) GetMetrics(ctx context.Context) (*provider.ProviderMetrics, error) {
	// This would need to be implemented with proper metrics collection
	return &provider.ProviderMetrics{
		RequestCount:   0,
		ErrorCount:     0,
		AverageLatency: 0,
		SuccessRate:    0.0,
		CollectedAt:    time.Now(),
	}, nil
}

// parseFullName parses owner/repo from full name
func parseFullName(fullName string) (owner, repo string, err error) {
	parts := splitFullName(fullName)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repository full name: %s", fullName)
	}
	return parts[0], parts[1], nil
}

// splitFullName splits "owner/repo" into ["owner", "repo"]
func splitFullName(fullName string) []string {
	result := make([]string, 0, 2)
	current := ""

	for _, char := range fullName {
		if char == '/' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}

	if current != "" {
		result = append(result, current)
	}

	return result
}
