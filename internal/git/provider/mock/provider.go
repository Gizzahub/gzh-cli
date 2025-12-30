// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package mock

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/gizzahub/gzh-cli/pkg/git/provider"
)

// Provider is a mock implementation of the GitProvider interface for testing.
type Provider struct {
	mock.Mock
	name         string
	repos        []provider.Repository
	webhooks     map[string][]provider.Webhook
	events       []provider.Event
	capabilities []provider.Capability
}

// NewProvider creates a new mock provider with the given name.
func NewProvider(name string) *Provider {
	p := &Provider{
		name:     name,
		repos:    []provider.Repository{},
		webhooks: make(map[string][]provider.Webhook),
		events:   []provider.Event{},
		capabilities: []provider.Capability{
			provider.CapabilityRepositories,
			provider.CapabilityWebhooks,
			provider.CapabilityEvents,
			provider.CapabilityIssues,
			provider.CapabilityWiki,
			provider.CapabilityReleases,
		},
	}
	return p
}

// Basic provider information

// GetName returns the provider name.
func (m *Provider) GetName() string {
	args := m.Called()
	if args.Get(0) != nil {
		return args.String(0)
	}
	return m.name
}

// GetCapabilities returns the provider capabilities.
func (m *Provider) GetCapabilities() []provider.Capability {
	args := m.Called()
	if args.Get(0) != nil {
		return args.Get(0).([]provider.Capability)
	}
	return m.capabilities
}

// GetBaseURL returns the provider base URL.
func (m *Provider) GetBaseURL() string {
	args := m.Called()
	if args.Get(0) != nil {
		return args.String(0)
	}
	return fmt.Sprintf("https://%s.example.com", m.name)
}

// Authentication

// Authenticate authenticates with the provider.
func (m *Provider) Authenticate(ctx context.Context, creds provider.Credentials) error {
	args := m.Called(ctx, creds)
	return args.Error(0)
}

// ValidateToken validates the authentication token.
func (m *Provider) ValidateToken(ctx context.Context) (*provider.TokenInfo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*provider.TokenInfo), args.Error(1)
}

// Repository management

// ListRepositories lists repositories with the given options.
func (m *Provider) ListRepositories(ctx context.Context, opts provider.ListOptions) (*provider.RepositoryList, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*provider.RepositoryList), args.Error(1)
}

// GetRepository gets a specific repository by ID.
func (m *Provider) GetRepository(ctx context.Context, id string) (*provider.Repository, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*provider.Repository), args.Error(1)
}

// CreateRepository creates a new repository.
func (m *Provider) CreateRepository(ctx context.Context, req provider.CreateRepoRequest) (*provider.Repository, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*provider.Repository), args.Error(1)
}

// UpdateRepository updates an existing repository.
func (m *Provider) UpdateRepository(ctx context.Context, id string, updates provider.UpdateRepoRequest) (*provider.Repository, error) {
	args := m.Called(ctx, id, updates)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*provider.Repository), args.Error(1)
}

// DeleteRepository deletes a repository.
func (m *Provider) DeleteRepository(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// ArchiveRepository archives a repository.
func (m *Provider) ArchiveRepository(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// UnarchiveRepository unarchives a repository.
func (m *Provider) UnarchiveRepository(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// CloneRepository clones a repository to the target location.
func (m *Provider) CloneRepository(ctx context.Context, repo provider.Repository, target string, opts provider.CloneOptions) error {
	args := m.Called(ctx, repo, target, opts)
	return args.Error(0)
}

// ForkRepository forks a repository.
func (m *Provider) ForkRepository(ctx context.Context, id string, opts provider.ForkOptions) (*provider.Repository, error) {
	args := m.Called(ctx, id, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*provider.Repository), args.Error(1)
}

// SearchRepositories searches for repositories.
func (m *Provider) SearchRepositories(ctx context.Context, query provider.SearchQuery) (*provider.SearchResult, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*provider.SearchResult), args.Error(1)
}

// Webhook management

// ListWebhooks lists webhooks for a repository.
func (m *Provider) ListWebhooks(ctx context.Context, repoID string) ([]provider.Webhook, error) {
	args := m.Called(ctx, repoID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]provider.Webhook), args.Error(1)
}

// GetWebhook gets a specific webhook.
func (m *Provider) GetWebhook(ctx context.Context, repoID, webhookID string) (*provider.Webhook, error) {
	args := m.Called(ctx, repoID, webhookID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*provider.Webhook), args.Error(1)
}

// CreateWebhook creates a new webhook.
func (m *Provider) CreateWebhook(ctx context.Context, repoID string, webhook provider.CreateWebhookRequest) (*provider.Webhook, error) {
	args := m.Called(ctx, repoID, webhook)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*provider.Webhook), args.Error(1)
}

// UpdateWebhook updates an existing webhook.
func (m *Provider) UpdateWebhook(ctx context.Context, repoID, webhookID string, updates provider.UpdateWebhookRequest) (*provider.Webhook, error) {
	args := m.Called(ctx, repoID, webhookID, updates)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*provider.Webhook), args.Error(1)
}

// DeleteWebhook deletes a webhook.
func (m *Provider) DeleteWebhook(ctx context.Context, repoID, webhookID string) error {
	args := m.Called(ctx, repoID, webhookID)
	return args.Error(0)
}

// TestWebhook tests a webhook.
func (m *Provider) TestWebhook(ctx context.Context, repoID, webhookID string) (*provider.WebhookTestResult, error) {
	args := m.Called(ctx, repoID, webhookID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*provider.WebhookTestResult), args.Error(1)
}

// ValidateWebhookURL validates a webhook URL.
func (m *Provider) ValidateWebhookURL(ctx context.Context, url string) error {
	args := m.Called(ctx, url)
	return args.Error(0)
}

// Event management

// ListEvents lists events.
func (m *Provider) ListEvents(ctx context.Context, opts provider.EventListOptions) ([]provider.Event, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]provider.Event), args.Error(1)
}

// GetEvent gets a specific event.
func (m *Provider) GetEvent(ctx context.Context, eventID string) (*provider.Event, error) {
	args := m.Called(ctx, eventID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*provider.Event), args.Error(1)
}

// ProcessEvent processes an event.
func (m *Provider) ProcessEvent(ctx context.Context, event provider.Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

// RegisterEventHandler registers an event handler.
func (m *Provider) RegisterEventHandler(eventType string, handler provider.EventHandler) error {
	args := m.Called(eventType, handler)
	return args.Error(0)
}

// StreamEvents streams events.
func (m *Provider) StreamEvents(ctx context.Context, opts provider.StreamOptions) (<-chan provider.Event, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(<-chan provider.Event), args.Error(1)
}

// Release management

// ListReleases lists releases for a repository.
func (m *Provider) ListReleases(ctx context.Context, repoID string, opts provider.ListReleasesOptions) (*provider.ReleaseList, error) {
	args := m.Called(ctx, repoID, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*provider.ReleaseList), args.Error(1)
}

// GetRelease gets a specific release by ID.
func (m *Provider) GetRelease(ctx context.Context, repoID, releaseID string) (*provider.Release, error) {
	args := m.Called(ctx, repoID, releaseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*provider.Release), args.Error(1)
}

// GetReleaseByTag gets a release by tag name.
func (m *Provider) GetReleaseByTag(ctx context.Context, repoID, tagName string) (*provider.Release, error) {
	args := m.Called(ctx, repoID, tagName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*provider.Release), args.Error(1)
}

// CreateRelease creates a new release.
func (m *Provider) CreateRelease(ctx context.Context, repoID string, req provider.CreateReleaseRequest) (*provider.Release, error) {
	args := m.Called(ctx, repoID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*provider.Release), args.Error(1)
}

// UpdateRelease updates an existing release.
func (m *Provider) UpdateRelease(ctx context.Context, repoID, releaseID string, updates provider.UpdateReleaseRequest) (*provider.Release, error) {
	args := m.Called(ctx, repoID, releaseID, updates)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*provider.Release), args.Error(1)
}

// DeleteRelease deletes a release.
func (m *Provider) DeleteRelease(ctx context.Context, repoID, releaseID string) error {
	args := m.Called(ctx, repoID, releaseID)
	return args.Error(0)
}

// ListReleaseAssets lists assets for a release.
func (m *Provider) ListReleaseAssets(ctx context.Context, repoID, releaseID string) ([]provider.Asset, error) {
	args := m.Called(ctx, repoID, releaseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]provider.Asset), args.Error(1)
}

// UploadReleaseAsset uploads an asset to a release.
func (m *Provider) UploadReleaseAsset(ctx context.Context, repoID string, req provider.UploadAssetRequest) (*provider.Asset, error) {
	args := m.Called(ctx, repoID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*provider.Asset), args.Error(1)
}

// DeleteReleaseAsset deletes a release asset.
func (m *Provider) DeleteReleaseAsset(ctx context.Context, repoID, assetID string) error {
	args := m.Called(ctx, repoID, assetID)
	return args.Error(0)
}

// DownloadReleaseAsset downloads a release asset.
func (m *Provider) DownloadReleaseAsset(ctx context.Context, repoID, assetID string) ([]byte, error) {
	args := m.Called(ctx, repoID, assetID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

// Health and monitoring

// HealthCheck performs a health check.
func (m *Provider) HealthCheck(ctx context.Context) (*provider.HealthStatus, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*provider.HealthStatus), args.Error(1)
}

// GetRateLimit gets the current rate limit status.
func (m *Provider) GetRateLimit(ctx context.Context) (*provider.RateLimit, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*provider.RateLimit), args.Error(1)
}

// GetMetrics gets provider metrics.
func (m *Provider) GetMetrics(ctx context.Context) (*provider.ProviderMetrics, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*provider.ProviderMetrics), args.Error(1)
}

// Test helper methods

// AddTestRepo adds a repository to the mock's internal storage.
func (m *Provider) AddTestRepo(repo provider.Repository) {
	m.repos = append(m.repos, repo)
}

// SetupListResponse sets up a mock response for ListRepositories.
func (m *Provider) SetupListResponse(org string, repos []provider.Repository) {
	result := &provider.RepositoryList{
		Repositories: repos,
		TotalCount:   len(repos),
		Page:         1,
		PerPage:      len(repos),
		HasNext:      false,
		HasPrev:      false,
	}

	m.On("ListRepositories", mock.Anything, mock.MatchedBy(func(opts provider.ListOptions) bool {
		return opts.Organization == org
	})).Return(result, nil)
}

// SetupGetResponse sets up a mock response for GetRepository.
func (m *Provider) SetupGetResponse(id string, repo *provider.Repository, err error) {
	m.On("GetRepository", mock.Anything, id).Return(repo, err)
}

// SetupCreateResponse sets up a mock response for CreateRepository.
func (m *Provider) SetupCreateResponse(matcher func(provider.CreateRepoRequest) bool, repo *provider.Repository, err error) {
	m.On("CreateRepository", mock.Anything, mock.MatchedBy(matcher)).Return(repo, err)
}

// SetupDeleteResponse sets up a mock response for DeleteRepository.
func (m *Provider) SetupDeleteResponse(id string, err error) {
	m.On("DeleteRepository", mock.Anything, id).Return(err)
}

// SetupArchiveResponse sets up a mock response for ArchiveRepository.
func (m *Provider) SetupArchiveResponse(id string, err error) {
	m.On("ArchiveRepository", mock.Anything, id).Return(err)
}

// SetupHealthResponse sets up a mock response for HealthCheck.
func (m *Provider) SetupHealthResponse(status *provider.HealthStatus, err error) {
	m.On("HealthCheck", mock.Anything).Return(status, err)
}

// Reset resets all mock expectations.
func (m *Provider) Reset() {
	m.Mock = mock.Mock{}
	m.repos = []provider.Repository{}
	m.webhooks = make(map[string][]provider.Webhook)
	m.events = []provider.Event{}
}

// FilterRepos filters repositories based on options (helper for tests).
func (m *Provider) FilterRepos(repos []provider.Repository, opts provider.ListOptions) []provider.Repository {
	var filtered []provider.Repository

	for _, repo := range repos {
		// Filter by visibility
		if opts.Visibility != "" && repo.Visibility != opts.Visibility {
			continue
		}

		// Filter by language
		if opts.Language != "" && !strings.EqualFold(repo.Language, opts.Language) {
			continue
		}

		// Filter by topic
		if opts.Topic != "" {
			hasTopicResult := false
			for _, topic := range repo.Topics {
				if strings.EqualFold(topic, opts.Topic) {
					hasTopicResult = true
					break
				}
			}
			if !hasTopicResult {
				continue
			}
		}

		// Filter by fork status
		if opts.Fork != nil && repo.Fork != *opts.Fork {
			continue
		}

		// Filter by archived status
		if opts.Archived != nil && repo.Archived != *opts.Archived {
			continue
		}

		// Filter by minimum stars
		if opts.MinStars > 0 && repo.Stars < opts.MinStars {
			continue
		}

		// Filter by maximum stars
		if opts.MaxStars > 0 && repo.Stars > opts.MaxStars {
			continue
		}

		// Filter by updated since
		if !opts.UpdatedSince.IsZero() && repo.UpdatedAt.Before(opts.UpdatedSince) {
			continue
		}

		filtered = append(filtered, repo)
	}

	return filtered
}

// GenerateTestToken generates a test token info.
func (m *Provider) GenerateTestToken() *provider.TokenInfo {
	return &provider.TokenInfo{
		Valid:     true,
		Scopes:    []string{"repo", "admin:org"},
		User:      "testuser",
		Email:     "test@example.com",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		Permissions: []string{
			"read:repository",
			"write:repository",
			"admin:repository",
		},
		RateLimit: provider.RateLimit{
			Limit:     5000,
			Remaining: 4950,
			Reset:     time.Now().Add(1 * time.Hour),
			Used:      50,
			Resource:  "core",
		},
	}
}

// GenerateTestHealthStatus generates a test health status.
func (m *Provider) GenerateTestHealthStatus() *provider.HealthStatus {
	return &provider.HealthStatus{
		Status:      provider.HealthStatusHealthy,
		LastChecked: time.Now(),
		Latency:     50 * time.Millisecond,
		Message:     "All systems operational",
		Details: map[string]interface{}{
			"api_status": "healthy",
			"db_status":  "healthy",
		},
	}
}
