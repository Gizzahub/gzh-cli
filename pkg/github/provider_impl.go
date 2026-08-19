// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gizzahub/gzh-cli/pkg/git/provider"
)

// GitHubProvider implements the unified GitProvider interface for GitHub.
type GitHubProvider struct {
	*provider.BaseProvider
	client  APIClient
	cloner  CloneService
	helpers *provider.CommonHelpers
}

// Ensure GitHubProvider implements GitProvider interface
var _ provider.GitProvider = (*GitHubProvider)(nil)

// NewGitHubProvider creates a new GitHub provider instance using the public GitHub API host.
func NewGitHubProvider(client APIClient, cloner CloneService) *GitHubProvider {
	return NewGitHubProviderWithBaseURL(client, cloner, "")
}

// NewGitHubProviderWithBaseURL creates a GitHub provider with a custom API base URL.
// Empty baseURL falls back to DefaultGitHubAPIBaseURL (for GHES or tests).
func NewGitHubProviderWithBaseURL(client APIClient, cloner CloneService, baseURL string) *GitHubProvider {
	return &GitHubProvider{
		BaseProvider: provider.NewBaseProvider("github", ResolveGitHubAPIBaseURL(baseURL), ""),
		client:       client,
		cloner:       cloner,
		helpers:      provider.NewCommonHelpers(),
	}
}

// GetCapabilities returns the list of supported capabilities.
func (g *GitHubProvider) GetCapabilities() []provider.Capability {
	capabilities := g.helpers.StandardizeCapabilities("github")
	// Add GitHub-specific capabilities
	return append(capabilities, []provider.Capability{
		provider.CapabilityCICD,
		provider.CapabilityBranchProtection,
		provider.CapabilitySecurityAlerts,
		provider.CapabilityDependabot,
	}...)
}

// Authenticate sets up authentication credentials.
func (g *GitHubProvider) Authenticate(ctx context.Context, creds provider.Credentials) error {
	switch creds.Type {
	case provider.CredentialTypeToken:
		g.SetToken(creds.Token)
		return g.client.SetToken(ctx, creds.Token)
	default:
		return g.FormatError("authenticate", fmt.Errorf("unsupported credential type: %s", creds.Type))
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

// CreateRepository creates a new repository under req.Owner (org or user).
// Owner is required — fail-fast when empty (CLI --org / sync destination).
func (g *GitHubProvider) CreateRepository(ctx context.Context, req provider.CreateRepoRequest) (*provider.Repository, error) {
	if req.Owner == "" {
		return nil, fmt.Errorf("owner is required for create repository")
	}
	if req.Name == "" {
		return nil, fmt.Errorf("name is required for create repository")
	}

	opts := &CreateRepositoryOptions{
		Name:              req.Name,
		Description:       req.Description,
		Homepage:          req.Homepage,
		Private:           req.Private,
		HasIssues:         req.HasIssues,
		HasProjects:       req.HasProjects,
		HasWiki:           req.HasWiki,
		HasDownloads:      req.HasDownloads,
		AutoInit:          req.AutoInit,
		GitignoreTemplate: req.GitignoreTemplate,
		LicenseTemplate:   req.LicenseTemplate,
		DefaultBranch:     req.DefaultBranch,
		AllowSquashMerge:  req.AllowSquashMerge,
		AllowMergeCommit:  req.AllowMergeCommit,
		AllowRebaseMerge:  req.AllowRebaseMerge,
		AllowAutoMerge:    req.AllowAutoMerge,
	}

	info, err := g.client.CreateRepository(ctx, req.Owner, opts)
	if err != nil {
		return nil, g.FormatError("create repository", err)
	}
	return repositoryInfoToProvider(info), nil
}

// UpdateRepository updates repository settings via PATCH /repos/{owner}/{repo}.
// id must be owner/repo.
func (g *GitHubProvider) UpdateRepository(ctx context.Context, id string, updates provider.UpdateRepoRequest) (*provider.Repository, error) {
	owner, repo, err := parseFullName(id)
	if err != nil {
		return nil, err
	}

	opts := &UpdateRepositoryOptions{
		Name:             updates.Name,
		Description:      updates.Description,
		Homepage:         updates.Homepage,
		Private:          updates.Private,
		HasIssues:        updates.HasIssues,
		HasProjects:      updates.HasProjects,
		HasWiki:          updates.HasWiki,
		HasDownloads:     updates.HasDownloads,
		DefaultBranch:    updates.DefaultBranch,
		AllowSquashMerge: updates.AllowSquashMerge,
		AllowMergeCommit: updates.AllowMergeCommit,
		AllowRebaseMerge: updates.AllowRebaseMerge,
		AllowAutoMerge:   updates.AllowAutoMerge,
		Archived:         updates.Archived,
	}

	info, err := g.client.UpdateRepository(ctx, owner, repo, opts)
	if err != nil {
		return nil, g.FormatError("update repository", err)
	}
	return repositoryInfoToProvider(info), nil
}

// DeleteRepository deletes a repository. id is owner/repo or full name.
func (g *GitHubProvider) DeleteRepository(ctx context.Context, id string) error {
	owner, repo, err := parseFullName(id)
	if err != nil {
		return err
	}
	if err := g.client.DeleteRepository(ctx, owner, repo); err != nil {
		return g.FormatError("delete repository", err)
	}
	return nil
}

// ArchiveRepository archives a repository. id is owner/repo or full name.
func (g *GitHubProvider) ArchiveRepository(ctx context.Context, id string) error {
	owner, repo, err := parseFullName(id)
	if err != nil {
		return err
	}
	if err := g.client.ArchiveRepository(ctx, owner, repo); err != nil {
		return g.FormatError("archive repository", err)
	}
	return nil
}

// UnarchiveRepository unarchives a repository. id is owner/repo or full name.
func (g *GitHubProvider) UnarchiveRepository(ctx context.Context, id string) error {
	owner, repo, err := parseFullName(id)
	if err != nil {
		return err
	}
	if err := g.client.UnarchiveRepository(ctx, owner, repo); err != nil {
		return g.FormatError("unarchive repository", err)
	}
	return nil
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

// ForkRepository creates a fork of a repository via POST /repos/{owner}/{repo}/forks.
// id must be owner/repo.
func (g *GitHubProvider) ForkRepository(ctx context.Context, id string, opts provider.ForkOptions) (*provider.Repository, error) {
	owner, repo, err := parseFullName(id)
	if err != nil {
		return nil, err
	}

	forkOpts := &ForkRepositoryOptions{
		Organization:      opts.Organization,
		Name:              opts.Name,
		DefaultBranchOnly: opts.DefaultBranchOnly,
	}

	info, err := g.client.ForkRepository(ctx, owner, repo, forkOpts)
	if err != nil {
		return nil, g.FormatError("fork repository", err)
	}
	return repositoryInfoToProvider(info), nil
}

// SearchRepositories searches for repositories via the GitHub search API.
func (g *GitHubProvider) SearchRepositories(ctx context.Context, query provider.SearchQuery) (*provider.SearchResult, error) {
	q := buildGitHubSearchQuery(query)
	if q == "" {
		return nil, fmt.Errorf("search query is required")
	}

	opts := &SearchRepositoriesOptions{
		Sort:    query.Sort,
		Order:   query.Order,
		Page:    query.Page,
		PerPage: query.PerPage,
	}
	result, err := g.client.SearchRepositories(ctx, q, opts)
	if err != nil {
		return nil, g.FormatError("search repositories", err)
	}

	repos := make([]provider.Repository, 0, len(result.Repositories))
	for i := range result.Repositories {
		repos = append(repos, *repositoryInfoToProvider(&result.Repositories[i]))
	}

	page := query.Page
	if page <= 0 {
		page = 1
	}
	perPage := query.PerPage
	if perPage <= 0 {
		perPage = 30
	}

	return &provider.SearchResult{
		TotalCount:        result.TotalCount,
		IncompleteResults: result.IncompleteResults,
		Repositories:      repos,
		Page:              page,
		PerPage:           perPage,
		HasNext:           page*perPage < result.TotalCount,
		HasPrev:           page > 1,
	}, nil
}

// Webhook management — GitHub hooks API via provider base URL/token (CLI: gz git webhook).

// ListWebhooks lists repository webhooks. repoID must be owner/repo.
func (g *GitHubProvider) ListWebhooks(ctx context.Context, repoID string) ([]provider.Webhook, error) {
	owner, repo, err := parseFullName(repoID)
	if err != nil {
		return nil, err
	}

	var hooks []WebhookInfo
	if err := g.doWebhookJSON(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/hooks", owner, repo), nil, &hooks, http.StatusOK); err != nil {
		return nil, g.FormatError("list webhooks", err)
	}

	out := make([]provider.Webhook, 0, len(hooks))
	for i := range hooks {
		out = append(out, webhookInfoToProvider(&hooks[i]))
	}
	return out, nil
}

// GetWebhook retrieves a single webhook by numeric ID.
func (g *GitHubProvider) GetWebhook(ctx context.Context, repoID, webhookID string) (*provider.Webhook, error) {
	owner, repo, err := parseFullName(repoID)
	if err != nil {
		return nil, err
	}
	id, err := strconv.ParseInt(webhookID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid webhook id %q: must be numeric", webhookID)
	}

	var hook WebhookInfo
	if err := g.doWebhookJSON(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/hooks/%d", owner, repo, id), nil, &hook, http.StatusOK); err != nil {
		return nil, g.FormatError("get webhook", err)
	}
	result := webhookInfoToProvider(&hook)
	return &result, nil
}

// CreateWebhook creates a repository webhook (GitHub "web" type).
func (g *GitHubProvider) CreateWebhook(ctx context.Context, repoID string, webhook provider.CreateWebhookRequest) (*provider.Webhook, error) {
	owner, repo, err := parseFullName(repoID)
	if err != nil {
		return nil, err
	}
	hookURL := webhook.Config.URL
	if hookURL == "" {
		return nil, fmt.Errorf("webhook URL is required")
	}
	if err := g.ValidateWebhookURL(ctx, hookURL); err != nil {
		return nil, err
	}
	events := webhook.Events
	if len(events) == 0 {
		events = []string{"push"}
	}
	contentType := webhook.Config.ContentType
	if contentType == "" {
		contentType = "json"
	}
	insecureSSL := "0"
	if webhook.Config.InsecureSSL {
		insecureSSL = "1"
	}

	apiRequest := map[string]any{
		"name":   "web",
		"active": webhook.Active,
		"events": events,
		"config": map[string]any{
			"url":          hookURL,
			"content_type": contentType,
			"insecure_ssl": insecureSSL,
		},
	}
	if webhook.Config.Secret != "" {
		apiRequest["config"].(map[string]any)["secret"] = webhook.Config.Secret
	}

	var hook WebhookInfo
	if err := g.doWebhookJSON(ctx, http.MethodPost, fmt.Sprintf("/repos/%s/%s/hooks", owner, repo), apiRequest, &hook, http.StatusCreated); err != nil {
		return nil, g.FormatError("create webhook", err)
	}
	result := webhookInfoToProvider(&hook)
	return &result, nil
}

// UpdateWebhook patches an existing repository webhook.
func (g *GitHubProvider) UpdateWebhook(ctx context.Context, repoID, webhookID string, updates provider.UpdateWebhookRequest) (*provider.Webhook, error) {
	owner, repo, err := parseFullName(repoID)
	if err != nil {
		return nil, err
	}
	id, err := strconv.ParseInt(webhookID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid webhook id %q: must be numeric", webhookID)
	}

	apiRequest := map[string]any{}
	if updates.Events != nil {
		apiRequest["events"] = updates.Events
	}
	if updates.Active != nil {
		apiRequest["active"] = *updates.Active
	}
	if updates.Config != nil {
		config := map[string]any{}
		if updates.Config.URL != "" {
			config["url"] = updates.Config.URL
		}
		if updates.Config.ContentType != "" {
			config["content_type"] = updates.Config.ContentType
		}
		if updates.Config.Secret != "" {
			config["secret"] = updates.Config.Secret
		}
		if updates.Config.InsecureSSL {
			config["insecure_ssl"] = "1"
		} else {
			config["insecure_ssl"] = "0"
		}
		apiRequest["config"] = config
	}

	var hook WebhookInfo
	if err := g.doWebhookJSON(ctx, http.MethodPatch, fmt.Sprintf("/repos/%s/%s/hooks/%d", owner, repo, id), apiRequest, &hook, http.StatusOK); err != nil {
		return nil, g.FormatError("update webhook", err)
	}
	result := webhookInfoToProvider(&hook)
	return &result, nil
}

// DeleteWebhook deletes a repository webhook by numeric ID.
func (g *GitHubProvider) DeleteWebhook(ctx context.Context, repoID, webhookID string) error {
	owner, repo, err := parseFullName(repoID)
	if err != nil {
		return err
	}
	id, err := strconv.ParseInt(webhookID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid webhook id %q: must be numeric", webhookID)
	}
	if err := g.doWebhookJSON(ctx, http.MethodDelete, fmt.Sprintf("/repos/%s/%s/hooks/%d", owner, repo, id), nil, nil, http.StatusNoContent); err != nil {
		return g.FormatError("delete webhook", err)
	}
	return nil
}

// TestWebhook triggers a ping delivery for a repository webhook.
func (g *GitHubProvider) TestWebhook(ctx context.Context, repoID, webhookID string) (*provider.WebhookTestResult, error) {
	owner, repo, err := parseFullName(repoID)
	if err != nil {
		return nil, err
	}
	id, err := strconv.ParseInt(webhookID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid webhook id %q: must be numeric", webhookID)
	}

	start := time.Now()
	// GitHub returns 204 No Content on success.
	if err := g.doWebhookJSON(ctx, http.MethodPost, fmt.Sprintf("/repos/%s/%s/hooks/%d/tests", owner, repo, id), nil, nil, http.StatusNoContent); err != nil {
		return &provider.WebhookTestResult{
			Success:      false,
			ResponseTime: time.Since(start),
			Error:        err.Error(),
		}, g.FormatError("test webhook", err)
	}
	return &provider.WebhookTestResult{
		Success:      true,
		StatusCode:   http.StatusNoContent,
		ResponseTime: time.Since(start),
	}, nil
}

// ValidateWebhookURL checks that a webhook URL uses http or https.
func (g *GitHubProvider) ValidateWebhookURL(_ context.Context, webhookURL string) error {
	if strings.TrimSpace(webhookURL) == "" {
		return fmt.Errorf("webhook URL cannot be empty")
	}
	if !strings.HasPrefix(webhookURL, "http://") && !strings.HasPrefix(webhookURL, "https://") {
		return fmt.Errorf("webhook URL must be a valid HTTP/HTTPS URL")
	}
	return nil
}

// doWebhookJSON performs an authenticated JSON request against the provider API base URL.
// Used for webhook CRUD so callers do not need webhook methods on APIClient.
func (g *GitHubProvider) doWebhookJSON(ctx context.Context, method, path string, payload any, out any, wantStatus int) error {
	var bodyReader io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to encode request: %w", err)
		}
		bodyReader = bytes.NewReader(encoded)
	}

	base := strings.TrimRight(g.GetBaseURL(), "/")
	if base == "" {
		base = DefaultGitHubAPIBaseURL
	}
	endpoint := base + "/" + strings.TrimPrefix(path, "/")
	req, err := http.NewRequestWithContext(ctx, method, endpoint, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	if token := g.GetToken(); token != "" {
		req.Header.Set("Authorization", "token "+token)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "gzh-cli")
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != wantStatus && !(wantStatus == http.StatusNoContent && resp.StatusCode == http.StatusOK) &&
		!(wantStatus == http.StatusCreated && resp.StatusCode == http.StatusOK) {
		return fmt.Errorf("API %s %s: HTTP %d - %s", method, path, resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	if out != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}
	return nil
}

func webhookInfoToProvider(w *WebhookInfo) provider.Webhook {
	if w == nil {
		return provider.Webhook{}
	}
	url := w.URL
	if url == "" {
		url = w.Config.URL
	}
	return provider.Webhook{
		ID:     strconv.FormatInt(w.ID, 10),
		Name:   w.Name,
		URL:    url,
		Events: w.Events,
		Active: w.Active,
		Config: provider.WebhookConfig{
			URL:         w.Config.URL,
			ContentType: w.Config.ContentType,
			Secret:      w.Config.Secret,
			InsecureSSL: w.Config.InsecureSSL,
		},
		CreatedAt: w.CreatedAt,
		UpdatedAt: w.UpdatedAt,
	}
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
		Details:     make(map[string]any),
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

// Release management

// ListReleases lists releases for a repository.
func (g *GitHubProvider) ListReleases(ctx context.Context, repoID string, opts provider.ListReleasesOptions) (*provider.ReleaseList, error) {
	// TODO: Implement GitHub releases API
	return nil, fmt.Errorf("not implemented")
}

// GetRelease gets a specific release by ID.
func (g *GitHubProvider) GetRelease(ctx context.Context, repoID, releaseID string) (*provider.Release, error) {
	return nil, fmt.Errorf("not implemented")
}

// GetReleaseByTag gets a release by tag name.
func (g *GitHubProvider) GetReleaseByTag(ctx context.Context, repoID, tagName string) (*provider.Release, error) {
	return nil, fmt.Errorf("not implemented")
}

// CreateRelease creates a new release.
func (g *GitHubProvider) CreateRelease(ctx context.Context, repoID string, req provider.CreateReleaseRequest) (*provider.Release, error) {
	return nil, fmt.Errorf("not implemented")
}

// UpdateRelease updates an existing release.
func (g *GitHubProvider) UpdateRelease(ctx context.Context, repoID, releaseID string, updates provider.UpdateReleaseRequest) (*provider.Release, error) {
	return nil, fmt.Errorf("not implemented")
}

// DeleteRelease deletes a release.
func (g *GitHubProvider) DeleteRelease(ctx context.Context, repoID, releaseID string) error {
	return fmt.Errorf("not implemented")
}

// ListReleaseAssets lists assets for a release.
// Deferred (issue 26): multipart asset APIs are non-trivial and not on CLI surface yet.
func (g *GitHubProvider) ListReleaseAssets(ctx context.Context, repoID, releaseID string) ([]provider.Asset, error) {
	return nil, fmt.Errorf("list release assets not implemented: deferred (issue 26; multipart/asset APIs not on CLI surface)")
}

// UploadReleaseAsset uploads an asset to a release.
// Deferred (issue 26): requires uploads.github.com host + Content-Type/size headers; not trivial.
func (g *GitHubProvider) UploadReleaseAsset(ctx context.Context, repoID string, req provider.UploadAssetRequest) (*provider.Asset, error) {
	return nil, fmt.Errorf("upload release asset not implemented: deferred (issue 26; multipart/asset APIs not on CLI surface)")
}

// DeleteReleaseAsset deletes a release asset.
// Deferred (issue 26): no CLI command yet; implement with release asset list/upload.
func (g *GitHubProvider) DeleteReleaseAsset(ctx context.Context, repoID, assetID string) error {
	return fmt.Errorf("delete release asset not implemented: deferred (issue 26; multipart/asset APIs not on CLI surface)")
}

// DownloadReleaseAsset downloads a release asset.
// Deferred (issue 26): redirect/binary download path differs from JSON API.
func (g *GitHubProvider) DownloadReleaseAsset(ctx context.Context, repoID, assetID string) ([]byte, error) {
	return nil, fmt.Errorf("download release asset not implemented: deferred (issue 26; multipart/asset APIs not on CLI surface)")
}

// repositoryInfoToProvider maps GitHub RepositoryInfo to provider.Repository.
func repositoryInfoToProvider(info *RepositoryInfo) *provider.Repository {
	if info == nil {
		return nil
	}
	return &provider.Repository{
		ID:            info.FullName,
		Name:          info.Name,
		FullName:      info.FullName,
		Description:   info.Description,
		DefaultBranch: info.DefaultBranch,
		CloneURL:      info.CloneURL,
		SSHURL:        info.SSHURL,
		HTMLURL:       info.HTMLURL,
		Private:       info.Private,
		Archived:      info.Archived,
		CreatedAt:     info.CreatedAt,
		UpdatedAt:     info.UpdatedAt,
		Language:      info.Language,
		Size:          int64(info.Size),
		Topics:        info.Topics,
		ProviderType:  "github",
	}
}

// buildGitHubSearchQuery composes a GitHub search q= string from SearchQuery fields.
func buildGitHubSearchQuery(query provider.SearchQuery) string {
	parts := make([]string, 0, 8)
	if query.Query != "" {
		parts = append(parts, query.Query)
	}
	if query.User != "" {
		parts = append(parts, "user:"+query.User)
	}
	if query.Organization != "" {
		parts = append(parts, "org:"+query.Organization)
	}
	if query.Language != "" {
		parts = append(parts, "language:"+query.Language)
	}
	if query.Topic != "" {
		parts = append(parts, "topic:"+query.Topic)
	}
	if query.Fork != nil {
		if *query.Fork {
			parts = append(parts, "fork:true")
		} else {
			parts = append(parts, "fork:false")
		}
	}
	if query.Archived != nil {
		if *query.Archived {
			parts = append(parts, "archived:true")
		} else {
			parts = append(parts, "archived:false")
		}
	}
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += " " + parts[i]
	}
	return result
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
