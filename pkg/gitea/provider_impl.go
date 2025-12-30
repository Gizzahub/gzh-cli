// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package gitea

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gizzahub/gzh-cli/pkg/git/provider"
)

// GiteaProvider implements the unified GitProvider interface for Gitea.
type GiteaProvider struct {
	*provider.BaseProvider
	helpers *provider.CommonHelpers
}

// Ensure GiteaProvider implements GitProvider interface
var _ provider.GitProvider = (*GiteaProvider)(nil)

// NewGiteaProvider creates a new Gitea provider instance.
func NewGiteaProvider(baseURL string) *GiteaProvider {
	if baseURL == "" {
		baseURL = "https://gitea.com/api/v1"
	}
	return &GiteaProvider{
		BaseProvider: provider.NewBaseProvider("gitea", baseURL, ""),
		helpers:      provider.NewCommonHelpers(),
	}
}

// GetCapabilities returns the list of supported capabilities.
func (g *GiteaProvider) GetCapabilities() []provider.Capability {
	return g.helpers.StandardizeCapabilities("gitea")
}

// Authenticate sets up authentication credentials.
func (g *GiteaProvider) Authenticate(ctx context.Context, creds provider.Credentials) error {
	switch creds.Type {
	case provider.CredentialTypeToken:
		g.SetToken(creds.Token)
		return nil
	default:
		return g.FormatError("authenticate", fmt.Errorf("unsupported credential type: %s", creds.Type))
	}
}

// ValidateToken validates the authentication token.
func (g *GiteaProvider) ValidateToken(ctx context.Context) (*provider.TokenInfo, error) {
	// Use existing Gitea list function to validate token
	_, err := List(ctx, "gitea")
	if err != nil {
		return &provider.TokenInfo{
			Valid: false,
		}, err
	}

	return &provider.TokenInfo{
		Valid:     true,
		Scopes:    []string{},           // Gitea scopes would need to be retrieved via API
		User:      "",                   // Would need additional API call
		Email:     "",                   // Would need additional API call
		RateLimit: provider.RateLimit{}, // Gitea rate limiting info
	}, nil
}

// ListRepositories lists repositories for an organization.
func (g *GiteaProvider) ListRepositories(ctx context.Context, opts provider.ListOptions) (*provider.RepositoryList, error) {
	owner := opts.Organization
	if owner == "" {
		owner = opts.User
	}
	if owner == "" {
		return nil, g.FormatError("list repositories", fmt.Errorf("either Organization or User must be specified in ListOptions"))
	}

	repoNames, err := List(ctx, owner)
	if err != nil {
		return nil, g.FormatError("list repositories", err)
	}

	repositories := make([]provider.Repository, 0, len(repoNames))
	for _, name := range repoNames {
		// Get additional repository information
		defaultBranch, err := GetDefaultBranch(ctx, owner, name)
		if err != nil {
			defaultBranch = "main" // fallback
		}

		fullName := fmt.Sprintf("%s/%s", owner, name)
		repo := provider.Repository{
			ID:            fullName,
			Name:          name,
			FullName:      fullName,
			DefaultBranch: defaultBranch,
			CloneURL:      fmt.Sprintf("https://gitea.com/%s.git", fullName),
			SSHURL:        fmt.Sprintf("git@gitea.com:%s.git", fullName),
			HTMLURL:       fmt.Sprintf("https://gitea.com/%s", fullName),
			ProviderType:  g.GetName(),
		}
		repositories = append(repositories, repo)
	}

	return &provider.RepositoryList{
		Repositories: repositories,
		TotalCount:   len(repositories),
	}, nil
}

// GetRepository retrieves information about a specific repository.
func (g *GiteaProvider) GetRepository(ctx context.Context, id string) (*provider.Repository, error) {
	owner, repo, err := g.helpers.ParseRepositoryURL(id)
	if err != nil {
		return nil, g.FormatError("get repository", err)
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
		CloneURL:      fmt.Sprintf("https://gitea.com/%s.git", id),
		SSHURL:        fmt.Sprintf("git@gitea.com:%s.git", id),
		HTMLURL:       fmt.Sprintf("https://gitea.com/%s", id),
		ProviderType:  "gitea",
	}, nil
}

// CloneRepository clones a repository to the target path.
func (g *GiteaProvider) CloneRepository(ctx context.Context, repo provider.Repository, target string, opts provider.CloneOptions) error {
	owner, repoName, err := g.helpers.ParseRepositoryURL(repo.FullName)
	if err != nil {
		return g.FormatError("clone repository", err)
	}

	err = Clone(ctx, target, owner, repoName, opts.Strategy)
	if err != nil {
		return g.FormatError("clone repository", err)
	}
	return nil
}

// Placeholder implementations for other required methods
func (g *GiteaProvider) CreateRepository(ctx context.Context, req provider.CreateRepoRequest) (*provider.Repository, error) {
	return nil, g.FormatError("create repository", fmt.Errorf("not implemented"))
}

func (g *GiteaProvider) UpdateRepository(ctx context.Context, id string, updates provider.UpdateRepoRequest) (*provider.Repository, error) {
	return nil, g.FormatError("update repository", fmt.Errorf("not implemented"))
}

func (g *GiteaProvider) DeleteRepository(ctx context.Context, id string) error {
	return g.FormatError("delete repository", fmt.Errorf("not implemented"))
}

func (g *GiteaProvider) ArchiveRepository(ctx context.Context, id string) error {
	return g.FormatError("archive repository", fmt.Errorf("not implemented"))
}

func (g *GiteaProvider) UnarchiveRepository(ctx context.Context, id string) error {
	return g.FormatError("unarchive repository", fmt.Errorf("not implemented"))
}

func (g *GiteaProvider) ForkRepository(ctx context.Context, id string, opts provider.ForkOptions) (*provider.Repository, error) {
	return nil, g.FormatError("fork repository", fmt.Errorf("not implemented"))
}

func (g *GiteaProvider) SearchRepositories(ctx context.Context, query provider.SearchQuery) (*provider.SearchResult, error) {
	return nil, g.FormatError("search repositories", fmt.Errorf("not implemented"))
}

// Webhook management methods (placeholder implementations)
func (g *GiteaProvider) ListWebhooks(ctx context.Context, repoID string) ([]provider.Webhook, error) {
	return nil, g.FormatError("list webhooks", fmt.Errorf("not implemented"))
}

func (g *GiteaProvider) GetWebhook(ctx context.Context, repoID, webhookID string) (*provider.Webhook, error) {
	return nil, g.FormatError("get webhook", fmt.Errorf("not implemented"))
}

func (g *GiteaProvider) CreateWebhook(ctx context.Context, repoID string, webhook provider.CreateWebhookRequest) (*provider.Webhook, error) {
	if err := g.helpers.ValidateWebhookRequest(repoID, "", webhook.Config.URL); err != nil {
		return nil, g.FormatError("create webhook", err)
	}
	return nil, g.FormatError("create webhook", fmt.Errorf("not implemented"))
}

func (g *GiteaProvider) UpdateWebhook(ctx context.Context, repoID, webhookID string, updates provider.UpdateWebhookRequest) (*provider.Webhook, error) {
	return nil, g.FormatError("update webhook", fmt.Errorf("not implemented"))
}

func (g *GiteaProvider) DeleteWebhook(ctx context.Context, repoID, webhookID string) error {
	return g.FormatError("delete webhook", fmt.Errorf("not implemented"))
}

func (g *GiteaProvider) TestWebhook(ctx context.Context, repoID, webhookID string) (*provider.WebhookTestResult, error) {
	return nil, g.FormatError("test webhook", fmt.Errorf("not implemented"))
}

func (g *GiteaProvider) ValidateWebhookURL(ctx context.Context, url string) error {
	if err := g.helpers.ValidateWebhookRequest("", "", url); err != nil {
		return g.FormatError("validate webhook URL", err)
	}
	return g.FormatError("validate webhook URL", fmt.Errorf("not implemented"))
}

// Event management methods (placeholder implementations)
func (g *GiteaProvider) ListEvents(ctx context.Context, opts provider.EventListOptions) ([]provider.Event, error) {
	return nil, g.FormatError("list events", fmt.Errorf("not implemented"))
}

func (g *GiteaProvider) GetEvent(ctx context.Context, eventID string) (*provider.Event, error) {
	return nil, g.FormatError("get event", fmt.Errorf("not implemented"))
}

func (g *GiteaProvider) ProcessEvent(ctx context.Context, event provider.Event) error {
	return g.FormatError("process event", fmt.Errorf("not implemented"))
}

func (g *GiteaProvider) RegisterEventHandler(eventType string, handler provider.EventHandler) error {
	return g.FormatError("register event handler", fmt.Errorf("not implemented"))
}

func (g *GiteaProvider) StreamEvents(ctx context.Context, opts provider.StreamOptions) (<-chan provider.Event, error) {
	return nil, g.FormatError("stream events", fmt.Errorf("not implemented"))
}

// Health and monitoring methods
func (g *GiteaProvider) HealthCheck(ctx context.Context) (*provider.HealthStatus, error) {
	// Use base provider health check first
	if err := g.BaseProvider.HealthCheck(ctx); err != nil {
		return &provider.HealthStatus{
			Status:      provider.HealthStatusUnhealthy,
			Message:     err.Error(),
			LastChecked: time.Now(),
			Details:     make(map[string]interface{}),
		}, nil
	}

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
		status.Message = "Gitea API accessible"
	}

	return status, nil
}

func (g *GiteaProvider) GetRateLimit(ctx context.Context) (*provider.RateLimit, error) {
	// Gitea rate limiting would need to be implemented
	return &provider.RateLimit{
		Limit:     1000,
		Remaining: 1000,
		Reset:     time.Now().Add(time.Hour),
		Used:      0,
		Resource:  "core",
	}, nil
}

func (g *GiteaProvider) GetMetrics(ctx context.Context) (*provider.ProviderMetrics, error) {
	return &provider.ProviderMetrics{
		RequestCount:   0,
		ErrorCount:     0,
		AverageLatency: 0,
		SuccessRate:    0.0,
		CollectedAt:    time.Now(),
	}, nil
}

// Release management

// ListReleases lists releases for a repository.
func (g *GiteaProvider) ListReleases(ctx context.Context, repoID string, opts provider.ListReleasesOptions) (*provider.ReleaseList, error) {
	owner, repo, err := g.helpers.ParseRepositoryURL(repoID)
	if err != nil {
		return nil, g.FormatError("list releases", err)
	}
	return ListReleases(ctx, owner, repo, opts)
}

// GetRelease gets a specific release by ID.
func (g *GiteaProvider) GetRelease(ctx context.Context, repoID, releaseID string) (*provider.Release, error) {
	owner, repo, err := g.helpers.ParseRepositoryURL(repoID)
	if err != nil {
		return nil, g.FormatError("get release", err)
	}
	id, err := strconv.ParseInt(releaseID, 10, 64)
	if err != nil {
		return nil, g.FormatError("get release", fmt.Errorf("invalid release ID: %s", releaseID))
	}
	return GetRelease(ctx, owner, repo, id)
}

// GetReleaseByTag gets a release by tag name.
func (g *GiteaProvider) GetReleaseByTag(ctx context.Context, repoID, tagName string) (*provider.Release, error) {
	owner, repo, err := g.helpers.ParseRepositoryURL(repoID)
	if err != nil {
		return nil, g.FormatError("get release by tag", err)
	}
	return GetReleaseByTag(ctx, owner, repo, tagName)
}

// CreateRelease creates a new release.
func (g *GiteaProvider) CreateRelease(ctx context.Context, repoID string, req provider.CreateReleaseRequest) (*provider.Release, error) {
	owner, repo, err := g.helpers.ParseRepositoryURL(repoID)
	if err != nil {
		return nil, g.FormatError("create release", err)
	}
	return CreateRelease(ctx, owner, repo, req)
}

// UpdateRelease updates an existing release.
func (g *GiteaProvider) UpdateRelease(ctx context.Context, repoID, releaseID string, updates provider.UpdateReleaseRequest) (*provider.Release, error) {
	owner, repo, err := g.helpers.ParseRepositoryURL(repoID)
	if err != nil {
		return nil, g.FormatError("update release", err)
	}
	id, err := strconv.ParseInt(releaseID, 10, 64)
	if err != nil {
		return nil, g.FormatError("update release", fmt.Errorf("invalid release ID: %s", releaseID))
	}
	return UpdateRelease(ctx, owner, repo, id, updates)
}

// DeleteRelease deletes a release.
func (g *GiteaProvider) DeleteRelease(ctx context.Context, repoID, releaseID string) error {
	owner, repo, err := g.helpers.ParseRepositoryURL(repoID)
	if err != nil {
		return g.FormatError("delete release", err)
	}
	id, err := strconv.ParseInt(releaseID, 10, 64)
	if err != nil {
		return g.FormatError("delete release", fmt.Errorf("invalid release ID: %s", releaseID))
	}
	return DeleteRelease(ctx, owner, repo, id)
}

// ListReleaseAssets lists assets for a release.
func (g *GiteaProvider) ListReleaseAssets(ctx context.Context, repoID, releaseID string) ([]provider.Asset, error) {
	owner, repo, err := g.helpers.ParseRepositoryURL(repoID)
	if err != nil {
		return nil, g.FormatError("list release assets", err)
	}
	id, err := strconv.ParseInt(releaseID, 10, 64)
	if err != nil {
		return nil, g.FormatError("list release assets", fmt.Errorf("invalid release ID: %s", releaseID))
	}
	return ListReleaseAssets(ctx, owner, repo, id)
}

// UploadReleaseAsset uploads an asset to a release.
func (g *GiteaProvider) UploadReleaseAsset(ctx context.Context, repoID string, req provider.UploadAssetRequest) (*provider.Asset, error) {
	// Gitea asset upload requires multipart form - not yet implemented
	// TODO: Implement Gitea release asset upload
	return nil, g.FormatError("upload release asset", fmt.Errorf("not implemented"))
}

// DeleteReleaseAsset deletes a release asset.
func (g *GiteaProvider) DeleteReleaseAsset(ctx context.Context, repoID, assetID string) error {
	owner, repo, err := g.helpers.ParseRepositoryURL(repoID)
	if err != nil {
		return g.FormatError("delete release asset", err)
	}
	id, err := strconv.ParseInt(assetID, 10, 64)
	if err != nil {
		return g.FormatError("delete release asset", fmt.Errorf("invalid asset ID: %s", assetID))
	}
	return DeleteReleaseAsset(ctx, owner, repo, id)
}

// DownloadReleaseAsset downloads a release asset.
func (g *GiteaProvider) DownloadReleaseAsset(ctx context.Context, repoID, assetID string) ([]byte, error) {
	owner, repo, err := g.helpers.ParseRepositoryURL(repoID)
	if err != nil {
		return nil, g.FormatError("download release asset", err)
	}
	id, err := strconv.ParseInt(assetID, 10, 64)
	if err != nil {
		return nil, g.FormatError("download release asset", fmt.Errorf("invalid asset ID: %s", assetID))
	}
	return DownloadReleaseAsset(ctx, owner, repo, id)
}
