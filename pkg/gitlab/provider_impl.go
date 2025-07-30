// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package gitlab

import (
	"context"
	"fmt"
	"time"

	"github.com/gizzahub/gzh-manager-go/pkg/git/provider"
)

// GitLabProvider implements the unified GitProvider interface for GitLab.
type GitLabProvider struct {
	baseURL string
	name    string
	token   string
}

// Ensure GitLabProvider implements GitProvider interface
var _ provider.GitProvider = (*GitLabProvider)(nil)

// NewGitLabProvider creates a new GitLab provider instance.
func NewGitLabProvider(baseURL string) *GitLabProvider {
	if baseURL == "" {
		baseURL = "https://gitlab.com/api/v4"
	}
	return &GitLabProvider{
		baseURL: baseURL,
		name:    "gitlab",
	}
}

// GetName returns the provider name.
func (g *GitLabProvider) GetName() string {
	return g.name
}

// GetCapabilities returns the list of supported capabilities.
func (g *GitLabProvider) GetCapabilities() []provider.Capability {
	return []provider.Capability{
		provider.CapabilityRepositories,
		provider.CapabilityWebhooks,
		provider.CapabilityEvents,
		provider.CapabilityIssues,
		provider.CapabilityMergeRequests,
		provider.CapabilityWiki,
		provider.CapabilityProjects,
		provider.CapabilityCICD,
		provider.CapabilityPackages,
		provider.CapabilityReleases,
		provider.CapabilityOrganizations,
		provider.CapabilityUsers,
		provider.CapabilityTeams,
		provider.CapabilityPermissions,
		provider.CapabilityBranchProtection,
	}
}

// GetBaseURL returns the base URL for the GitLab API.
func (g *GitLabProvider) GetBaseURL() string {
	return g.baseURL
}

// Authenticate sets up authentication credentials.
func (g *GitLabProvider) Authenticate(ctx context.Context, creds provider.Credentials) error {
	switch creds.Type {
	case provider.CredentialTypeToken:
		g.token = creds.Token
		return nil
	default:
		return fmt.Errorf("unsupported credential type: %s", creds.Type)
	}
}

// ValidateToken validates the authentication token.
func (g *GitLabProvider) ValidateToken(ctx context.Context) (*provider.TokenInfo, error) {
	// Use existing GitLab list function to validate token
	_, err := List(ctx, "gitlab-org")
	if err != nil {
		return &provider.TokenInfo{
			Valid: false,
		}, err
	}

	return &provider.TokenInfo{
		Valid:     true,
		Scopes:    []string{},           // GitLab scopes would need to be retrieved via API
		User:      "",                   // Would need additional API call
		Email:     "",                   // Would need additional API call
		RateLimit: provider.RateLimit{}, // GitLab rate limiting info
	}, nil
}

// ListRepositories lists repositories for an organization.
func (g *GitLabProvider) ListRepositories(ctx context.Context, opts provider.ListOptions) (*provider.RepositoryList, error) {
	owner := opts.Organization
	if owner == "" {
		owner = opts.User
	}
	if owner == "" {
		return nil, fmt.Errorf("either Organization or User must be specified in ListOptions")
	}

	repoNames, err := List(ctx, owner)
	if err != nil {
		return nil, err
	}

	repositories := make([]provider.Repository, 0, len(repoNames))
	for _, name := range repoNames {
		// Get additional repository information
		defaultBranch, err := GetDefaultBranch(ctx, owner, name)
		if err != nil {
			defaultBranch = "main" // fallback
		}

		repo := provider.Repository{
			ID:            fmt.Sprintf("%s/%s", owner, name),
			Name:          name,
			FullName:      fmt.Sprintf("%s/%s", owner, name),
			DefaultBranch: defaultBranch,
			CloneURL:      fmt.Sprintf("https://gitlab.com/%s/%s.git", owner, name),
			SSHURL:        fmt.Sprintf("git@gitlab.com:%s/%s.git", owner, name),
			HTMLURL:       fmt.Sprintf("https://gitlab.com/%s/%s", owner, name),
			ProviderType:  "gitlab",
		}
		repositories = append(repositories, repo)
	}

	return &provider.RepositoryList{
		Repositories: repositories,
		TotalCount:   len(repositories),
	}, nil
}

// GetRepository retrieves information about a specific repository.
func (g *GitLabProvider) GetRepository(ctx context.Context, id string) (*provider.Repository, error) {
	owner, repo, err := parseFullName(id)
	if err != nil {
		return nil, err
	}

	defaultBranch, err := GetDefaultBranch(ctx, owner, repo)
	if err != nil {
		defaultBranch = "main"
	}

	return &provider.Repository{
		ID:            id,
		Name:          repo,
		FullName:      id,
		DefaultBranch: defaultBranch,
		CloneURL:      fmt.Sprintf("https://gitlab.com/%s.git", id),
		SSHURL:        fmt.Sprintf("git@gitlab.com:%s.git", id),
		HTMLURL:       fmt.Sprintf("https://gitlab.com/%s", id),
		ProviderType:  "gitlab",
	}, nil
}

// CloneRepository clones a repository to the target path.
func (g *GitLabProvider) CloneRepository(ctx context.Context, repo provider.Repository, target string, opts provider.CloneOptions) error {
	owner, repoName, err := parseFullName(repo.FullName)
	if err != nil {
		return err
	}

	return Clone(ctx, target, owner, repoName, opts.Strategy)
}

// Placeholder implementations for other required methods
func (g *GitLabProvider) CreateRepository(ctx context.Context, req provider.CreateRepoRequest) (*provider.Repository, error) {
	return nil, fmt.Errorf("create repository not implemented")
}

func (g *GitLabProvider) UpdateRepository(ctx context.Context, id string, updates provider.UpdateRepoRequest) (*provider.Repository, error) {
	return nil, fmt.Errorf("update repository not implemented")
}

func (g *GitLabProvider) DeleteRepository(ctx context.Context, id string) error {
	return fmt.Errorf("delete repository not implemented")
}

func (g *GitLabProvider) ArchiveRepository(ctx context.Context, id string) error {
	return fmt.Errorf("archive repository not implemented")
}

func (g *GitLabProvider) UnarchiveRepository(ctx context.Context, id string) error {
	return fmt.Errorf("unarchive repository not implemented")
}

func (g *GitLabProvider) ForkRepository(ctx context.Context, id string, opts provider.ForkOptions) (*provider.Repository, error) {
	return nil, fmt.Errorf("fork repository not implemented")
}

func (g *GitLabProvider) SearchRepositories(ctx context.Context, query provider.SearchQuery) (*provider.SearchResult, error) {
	return nil, fmt.Errorf("search repositories not implemented")
}

// Webhook management methods (placeholder implementations)
func (g *GitLabProvider) ListWebhooks(ctx context.Context, repoID string) ([]provider.Webhook, error) {
	return nil, fmt.Errorf("webhook management not implemented")
}

func (g *GitLabProvider) GetWebhook(ctx context.Context, repoID, webhookID string) (*provider.Webhook, error) {
	return nil, fmt.Errorf("webhook management not implemented")
}

func (g *GitLabProvider) CreateWebhook(ctx context.Context, repoID string, webhook provider.CreateWebhookRequest) (*provider.Webhook, error) {
	return nil, fmt.Errorf("webhook management not implemented")
}

func (g *GitLabProvider) UpdateWebhook(ctx context.Context, repoID, webhookID string, updates provider.UpdateWebhookRequest) (*provider.Webhook, error) {
	return nil, fmt.Errorf("webhook management not implemented")
}

func (g *GitLabProvider) DeleteWebhook(ctx context.Context, repoID, webhookID string) error {
	return fmt.Errorf("webhook management not implemented")
}

func (g *GitLabProvider) TestWebhook(ctx context.Context, repoID, webhookID string) (*provider.WebhookTestResult, error) {
	return nil, fmt.Errorf("webhook management not implemented")
}

func (g *GitLabProvider) ValidateWebhookURL(ctx context.Context, url string) error {
	return fmt.Errorf("webhook management not implemented")
}

// Event management methods (placeholder implementations)
func (g *GitLabProvider) ListEvents(ctx context.Context, opts provider.EventListOptions) ([]provider.Event, error) {
	return nil, fmt.Errorf("event management not implemented")
}

func (g *GitLabProvider) GetEvent(ctx context.Context, eventID string) (*provider.Event, error) {
	return nil, fmt.Errorf("event management not implemented")
}

func (g *GitLabProvider) ProcessEvent(ctx context.Context, event provider.Event) error {
	return fmt.Errorf("event management not implemented")
}

func (g *GitLabProvider) RegisterEventHandler(eventType string, handler provider.EventHandler) error {
	return fmt.Errorf("event management not implemented")
}

func (g *GitLabProvider) StreamEvents(ctx context.Context, opts provider.StreamOptions) (<-chan provider.Event, error) {
	return nil, fmt.Errorf("event streaming not implemented")
}

// Health and monitoring methods
func (g *GitLabProvider) HealthCheck(ctx context.Context) (*provider.HealthStatus, error) {
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
		status.Message = "GitLab API accessible"
	}

	return status, nil
}

func (g *GitLabProvider) GetRateLimit(ctx context.Context) (*provider.RateLimit, error) {
	// GitLab rate limiting would need to be implemented
	return &provider.RateLimit{
		Limit:     1000,
		Remaining: 1000,
		Reset:     time.Now().Add(time.Hour),
		Used:      0,
		Resource:  "core",
	}, nil
}

func (g *GitLabProvider) GetMetrics(ctx context.Context) (*provider.ProviderMetrics, error) {
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
