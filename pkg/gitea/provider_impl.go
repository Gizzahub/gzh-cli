// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package gitea

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gizzahub/gzh-cli/internal/httpclient"
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
		// Keep package-level token in sync for List/Clone helpers that use addAuthHeader.
		SetToken(creds.Token)
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

// CreateRepository creates a repository under req.Owner (org preferred, user fallback).
// Owner is required — fail-fast when empty.
func (g *GiteaProvider) CreateRepository(ctx context.Context, req provider.CreateRepoRequest) (*provider.Repository, error) {
	if req.Owner == "" {
		return nil, g.FormatError("create repository", fmt.Errorf("owner is required"))
	}
	if req.Name == "" {
		return nil, g.FormatError("create repository", fmt.Errorf("name is required"))
	}

	payload := map[string]any{
		"name":          req.Name,
		"description":   req.Description,
		"private":       req.Private,
		"auto_init":     req.AutoInit,
		"default_branch": req.DefaultBranch,
		"has_issues":    req.HasIssues,
		"has_wiki":      req.HasWiki,
		"has_projects":  req.HasProjects,
	}

	var repo giteaRepo
	// Prefer org endpoint; fall back to user repos when owner is not an org.
	err := g.doJSON(ctx, "POST", fmt.Sprintf("orgs/%s/repos", url.PathEscape(req.Owner)), payload, &repo, http.StatusCreated, http.StatusOK)
	if err != nil {
		if !strings.Contains(err.Error(), "404") {
			return nil, g.FormatError("create repository", err)
		}
		if err2 := g.doJSON(ctx, "POST", "user/repos", payload, &repo, http.StatusCreated, http.StatusOK); err2 != nil {
			return nil, g.FormatError("create repository", err)
		}
	}
	return giteaRepoToProvider(&repo), nil
}

// UpdateRepository updates repository settings via PATCH /repos/{owner}/{repo}.
// id must be owner/repo.
func (g *GiteaProvider) UpdateRepository(ctx context.Context, id string, updates provider.UpdateRepoRequest) (*provider.Repository, error) {
	owner, repo, err := parseOwnerRepo(id)
	if err != nil {
		return nil, g.FormatError("update repository", err)
	}

	payload := map[string]any{}
	if updates.Name != nil {
		payload["name"] = *updates.Name
	}
	if updates.Description != nil {
		payload["description"] = *updates.Description
	}
	if updates.Private != nil {
		payload["private"] = *updates.Private
	}
	if updates.Archived != nil {
		payload["archived"] = *updates.Archived
	}
	if updates.DefaultBranch != nil {
		payload["default_branch"] = *updates.DefaultBranch
	}
	if updates.HasIssues != nil {
		payload["has_issues"] = *updates.HasIssues
	}
	if updates.HasWiki != nil {
		payload["has_wiki"] = *updates.HasWiki
	}
	if updates.HasProjects != nil {
		payload["has_projects"] = *updates.HasProjects
	}
	if updates.Homepage != nil {
		payload["website"] = *updates.Homepage
	}

	path := fmt.Sprintf("repos/%s/%s", url.PathEscape(owner), url.PathEscape(repo))
	var out giteaRepo
	if err := g.doJSON(ctx, "PATCH", path, payload, &out, http.StatusOK); err != nil {
		return nil, g.FormatError("update repository", err)
	}
	return giteaRepoToProvider(&out), nil
}

// DeleteRepository deletes a repository. id is owner/repo.
func (g *GiteaProvider) DeleteRepository(ctx context.Context, id string) error {
	owner, repo, err := parseOwnerRepo(id)
	if err != nil {
		return g.FormatError("delete repository", err)
	}
	path := fmt.Sprintf("repos/%s/%s", url.PathEscape(owner), url.PathEscape(repo))
	if err := g.doJSON(ctx, "DELETE", path, nil, nil, http.StatusNoContent, http.StatusOK); err != nil {
		return g.FormatError("delete repository", err)
	}
	return nil
}

// ArchiveRepository archives a repository. id is owner/repo.
func (g *GiteaProvider) ArchiveRepository(ctx context.Context, id string) error {
	return g.setArchived(ctx, id, true)
}

// UnarchiveRepository unarchives a repository. id is owner/repo.
func (g *GiteaProvider) UnarchiveRepository(ctx context.Context, id string) error {
	return g.setArchived(ctx, id, false)
}

func (g *GiteaProvider) setArchived(ctx context.Context, id string, archived bool) error {
	owner, repo, err := parseOwnerRepo(id)
	if err != nil {
		return g.FormatError("archive repository", err)
	}
	path := fmt.Sprintf("repos/%s/%s", url.PathEscape(owner), url.PathEscape(repo))
	payload := map[string]bool{"archived": archived}
	op := "unarchive repository"
	if archived {
		op = "archive repository"
	}
	if err := g.doJSON(ctx, "PATCH", path, payload, nil, http.StatusOK); err != nil {
		return g.FormatError(op, err)
	}
	return nil
}

// ForkRepository forks a repository via POST /repos/{owner}/{repo}/forks.
// id must be owner/repo.
func (g *GiteaProvider) ForkRepository(ctx context.Context, id string, opts provider.ForkOptions) (*provider.Repository, error) {
	owner, repo, err := parseOwnerRepo(id)
	if err != nil {
		return nil, g.FormatError("fork repository", err)
	}

	payload := map[string]any{}
	if opts.Organization != "" {
		payload["organization"] = opts.Organization
	}
	if opts.Name != "" {
		payload["name"] = opts.Name
	}

	path := fmt.Sprintf("repos/%s/%s/forks", url.PathEscape(owner), url.PathEscape(repo))
	var out giteaRepo
	if err := g.doJSON(ctx, "POST", path, payload, &out, http.StatusCreated, http.StatusOK, http.StatusAccepted); err != nil {
		return nil, g.FormatError("fork repository", err)
	}
	return giteaRepoToProvider(&out), nil
}

// SearchRepositories searches repositories via the Gitea search API.
func (g *GiteaProvider) SearchRepositories(ctx context.Context, query provider.SearchQuery) (*provider.SearchResult, error) {
	q := query.Query
	if q == "" {
		return nil, g.FormatError("search repositories", fmt.Errorf("search query is required"))
	}

	page := query.Page
	if page <= 0 {
		page = 1
	}
	perPage := query.PerPage
	if perPage <= 0 {
		perPage = 30
	}

	params := url.Values{}
	params.Set("q", q)
	params.Set("page", fmt.Sprintf("%d", page))
	params.Set("limit", fmt.Sprintf("%d", perPage))
	if query.User != "" {
		params.Set("user", query.User)
	}
	if query.Topic != "" {
		params.Set("topic", "true")
	}

	var result struct {
		OK   bool        `json:"ok"`
		Data []giteaRepo `json:"data"`
	}
	if err := g.doJSON(ctx, "GET", "repos/search?"+params.Encode(), nil, &result, http.StatusOK); err != nil {
		return nil, g.FormatError("search repositories", err)
	}

	repos := make([]provider.Repository, 0, len(result.Data))
	for i := range result.Data {
		repos = append(repos, *giteaRepoToProvider(&result.Data[i]))
	}

	return &provider.SearchResult{
		TotalCount:   len(repos),
		Repositories: repos,
		Page:         page,
		PerPage:      perPage,
		HasNext:      len(repos) == perPage,
		HasPrev:      page > 1,
	}, nil
}

// giteaHook is the Gitea repository hook API shape.
type giteaHook struct {
	ID     int64  `json:"id"`
	Type   string `json:"type"`
	Active bool   `json:"active"`
	Config struct {
		URL         string `json:"url"`
		ContentType string `json:"content_type"`
	} `json:"config"`
	Events []string `json:"events"`
}

// Webhook management — Gitea repository hooks API.
func (g *GiteaProvider) ListWebhooks(ctx context.Context, repoID string) ([]provider.Webhook, error) {
	owner, repo, err := parseOwnerRepo(repoID)
	if err != nil {
		return nil, err
	}
	var hooks []giteaHook
	if err := g.doJSON(ctx, "GET", fmt.Sprintf("repos/%s/%s/hooks", owner, repo), nil, &hooks, http.StatusOK); err != nil {
		return nil, g.FormatError("list webhooks", err)
	}
	out := make([]provider.Webhook, 0, len(hooks))
	for i := range hooks {
		out = append(out, giteaHookToProvider(&hooks[i]))
	}
	return out, nil
}

func (g *GiteaProvider) GetWebhook(ctx context.Context, repoID, webhookID string) (*provider.Webhook, error) {
	owner, repo, err := parseOwnerRepo(repoID)
	if err != nil {
		return nil, err
	}
	var hook giteaHook
	if err := g.doJSON(ctx, "GET", fmt.Sprintf("repos/%s/%s/hooks/%s", owner, repo, webhookID), nil, &hook, http.StatusOK); err != nil {
		return nil, g.FormatError("get webhook", err)
	}
	w := giteaHookToProvider(&hook)
	return &w, nil
}

func (g *GiteaProvider) CreateWebhook(ctx context.Context, repoID string, webhook provider.CreateWebhookRequest) (*provider.Webhook, error) {
	if err := g.ValidateWebhookURL(ctx, webhook.Config.URL); err != nil {
		return nil, err
	}
	owner, repo, err := parseOwnerRepo(repoID)
	if err != nil {
		return nil, err
	}
	events := webhook.Events
	if len(events) == 0 {
		events = []string{"push"}
	}
	ct := webhook.Config.ContentType
	if ct == "" {
		ct = "json"
	}
	payload := map[string]any{
		"type":   "gitea",
		"active": webhook.Active,
		"events": events,
		"config": map[string]any{
			"url":          webhook.Config.URL,
			"content_type": ct,
		},
	}
	var hook giteaHook
	if err := g.doJSON(ctx, "POST", fmt.Sprintf("repos/%s/%s/hooks", owner, repo), payload, &hook, http.StatusCreated, http.StatusOK); err != nil {
		return nil, g.FormatError("create webhook", err)
	}
	w := giteaHookToProvider(&hook)
	return &w, nil
}

func (g *GiteaProvider) UpdateWebhook(ctx context.Context, repoID, webhookID string, updates provider.UpdateWebhookRequest) (*provider.Webhook, error) {
	owner, repo, err := parseOwnerRepo(repoID)
	if err != nil {
		return nil, err
	}
	payload := map[string]any{}
	if updates.Active != nil {
		payload["active"] = *updates.Active
	}
	if len(updates.Events) > 0 {
		payload["events"] = updates.Events
	}
	if updates.Config != nil {
		if updates.Config.URL != "" {
			if err := g.ValidateWebhookURL(ctx, updates.Config.URL); err != nil {
				return nil, err
			}
		}
		ct := updates.Config.ContentType
		if ct == "" {
			ct = "json"
		}
		payload["config"] = map[string]any{
			"url":          updates.Config.URL,
			"content_type": ct,
		}
	}
	var hook giteaHook
	if err := g.doJSON(ctx, "PATCH", fmt.Sprintf("repos/%s/%s/hooks/%s", owner, repo, webhookID), payload, &hook, http.StatusOK); err != nil {
		return nil, g.FormatError("update webhook", err)
	}
	w := giteaHookToProvider(&hook)
	return &w, nil
}

func (g *GiteaProvider) DeleteWebhook(ctx context.Context, repoID, webhookID string) error {
	owner, repo, err := parseOwnerRepo(repoID)
	if err != nil {
		return err
	}
	if err := g.doJSON(ctx, "DELETE", fmt.Sprintf("repos/%s/%s/hooks/%s", owner, repo, webhookID), nil, nil, http.StatusNoContent, http.StatusOK); err != nil {
		return g.FormatError("delete webhook", err)
	}
	return nil
}

func (g *GiteaProvider) TestWebhook(ctx context.Context, repoID, webhookID string) (*provider.WebhookTestResult, error) {
	owner, repo, err := parseOwnerRepo(repoID)
	if err != nil {
		return nil, err
	}
	if err := g.doJSON(ctx, "POST", fmt.Sprintf("repos/%s/%s/hooks/%s/tests", owner, repo, webhookID), nil, nil, http.StatusNoContent, http.StatusOK); err != nil {
		return nil, g.FormatError("test webhook", err)
	}
	return &provider.WebhookTestResult{Success: true, StatusCode: http.StatusNoContent}, nil
}

func (g *GiteaProvider) ValidateWebhookURL(_ context.Context, webhookURL string) error {
	if strings.TrimSpace(webhookURL) == "" {
		return fmt.Errorf("webhook URL cannot be empty")
	}
	if !strings.HasPrefix(webhookURL, "http://") && !strings.HasPrefix(webhookURL, "https://") {
		return fmt.Errorf("webhook URL must be a valid HTTP/HTTPS URL")
	}
	return nil
}

func giteaHookToProvider(h *giteaHook) provider.Webhook {
	name := h.Type
	if name == "" {
		name = "gitea"
	}
	ct := h.Config.ContentType
	if ct == "" {
		ct = "json"
	}
	return provider.Webhook{
		ID:     fmt.Sprintf("%d", h.ID),
		Name:   name,
		URL:    h.Config.URL,
		Active: h.Active,
		Events: h.Events,
		Config: provider.WebhookConfig{
			URL:         h.Config.URL,
			ContentType: ct,
		},
	}
}

// Event management methods — not in CLI surface; deferred (issue 26 phase 2+)
func (g *GiteaProvider) ListEvents(ctx context.Context, opts provider.EventListOptions) ([]provider.Event, error) {
	return nil, g.FormatError("list events", fmt.Errorf("not implemented: not in CLI surface; deferred (issue 26 phase 2+)"))
}

func (g *GiteaProvider) GetEvent(ctx context.Context, eventID string) (*provider.Event, error) {
	return nil, g.FormatError("get event", fmt.Errorf("not implemented: not in CLI surface; deferred (issue 26 phase 2+)"))
}

func (g *GiteaProvider) ProcessEvent(ctx context.Context, event provider.Event) error {
	return g.FormatError("process event", fmt.Errorf("not implemented: not in CLI surface; deferred (issue 26 phase 2+)"))
}

func (g *GiteaProvider) RegisterEventHandler(eventType string, handler provider.EventHandler) error {
	return g.FormatError("register event handler", fmt.Errorf("not implemented: not in CLI surface; deferred (issue 26 phase 2+)"))
}

func (g *GiteaProvider) StreamEvents(ctx context.Context, opts provider.StreamOptions) (<-chan provider.Event, error) {
	return nil, g.FormatError("stream events", fmt.Errorf("not implemented: not in CLI surface; deferred (issue 26 phase 2+)"))
}

// Health and monitoring methods
func (g *GiteaProvider) HealthCheck(ctx context.Context) (*provider.HealthStatus, error) {
	// Use base provider health check first
	if err := g.BaseProvider.HealthCheck(ctx); err != nil {
		return &provider.HealthStatus{
			Status:      provider.HealthStatusUnhealthy,
			Message:     err.Error(),
			LastChecked: time.Now(),
			Details:     make(map[string]any),
		}, nil
	}

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
// not in CLI surface; deferred (issue 26 phase 2+)
func (g *GiteaProvider) UploadReleaseAsset(ctx context.Context, repoID string, req provider.UploadAssetRequest) (*provider.Asset, error) {
	return nil, g.FormatError("upload release asset", fmt.Errorf("not implemented: not in CLI surface; deferred (issue 26 phase 2+)"))
}

// giteaRepo is the subset of Gitea repository JSON used by mutation APIs.
type giteaRepo struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	FullName      string `json:"full_name"`
	Description   string `json:"description"`
	DefaultBranch string `json:"default_branch"`
	CloneURL      string `json:"clone_url"`
	SSHURL        string `json:"ssh_url"`
	HTMLURL       string `json:"html_url"`
	Private       bool   `json:"private"`
	Archived      bool   `json:"archived"`
}

func giteaRepoToProvider(r *giteaRepo) *provider.Repository {
	if r == nil {
		return nil
	}
	return &provider.Repository{
		ID:            r.FullName,
		Name:          r.Name,
		FullName:      r.FullName,
		Description:   r.Description,
		DefaultBranch: r.DefaultBranch,
		CloneURL:      r.CloneURL,
		SSHURL:        r.SSHURL,
		HTMLURL:       r.HTMLURL,
		Private:       r.Private,
		Archived:      r.Archived,
		ProviderType:  "gitea",
	}
}

func parseOwnerRepo(id string) (owner, repo string, err error) {
	parts := strings.Split(strings.Trim(id, "/"), "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid repository id %q (expected owner/repo)", id)
	}
	return parts[0], parts[1], nil
}

func (g *GiteaProvider) apiBase() string {
	base := strings.TrimRight(g.GetBaseURL(), "/")
	if base == "" {
		base = "https://gitea.com/api/v1"
	}
	return base
}

// doJSON performs an authenticated JSON request against this provider's base URL.
func (g *GiteaProvider) doJSON(ctx context.Context, method, endpoint string, payload any, out any, wantStatuses ...int) error {
	var bodyReader io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to encode request: %w", err)
		}
		bodyReader = bytes.NewReader(encoded)
	}

	fullURL := g.apiBase() + "/" + strings.TrimPrefix(endpoint, "/")
	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	token := g.GetToken()
	if token == "" {
		token = configuredToken
	}
	if token != "" {
		req.Header.Set("Authorization", "token "+token)
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := httpclient.GetGlobalClient("gitea")
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	ok := false
	for _, want := range wantStatuses {
		if resp.StatusCode == want {
			ok = true
			break
		}
	}
	if !ok {
		return fmt.Errorf("API %s %s: HTTP %d - %s", method, endpoint, resp.StatusCode, strings.TrimSpace(string(respBody)))
	}
	if out != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}
	return nil
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
