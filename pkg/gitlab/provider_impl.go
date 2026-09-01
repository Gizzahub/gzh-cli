// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package gitlab

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gizzahub/gzh-cli/internal/httpclient"
	"github.com/gizzahub/gzh-cli/pkg/git/provider"
)

// GitLabProvider implements the unified GitProvider interface for GitLab.
type GitLabProvider struct {
	*provider.BaseProvider
	helpers *provider.CommonHelpers
}

// Ensure GitLabProvider implements GitProvider interface
var _ provider.GitProvider = (*GitLabProvider)(nil)

// NewGitLabProvider creates a new GitLab provider instance.
func NewGitLabProvider(baseURL string) *GitLabProvider {
	if baseURL == "" {
		baseURL = "https://gitlab.com/api/v4"
	}
	return &GitLabProvider{
		BaseProvider: provider.NewBaseProvider("gitlab", baseURL, ""),
		helpers:      provider.NewCommonHelpers(),
	}
}

// GetCapabilities returns the list of supported capabilities.
func (g *GitLabProvider) GetCapabilities() []provider.Capability {
	capabilities := g.helpers.StandardizeCapabilities("gitlab")
	// Add GitLab-specific capabilities
	return append(capabilities, provider.CapabilityBranchProtection)
}

// Authenticate sets up authentication credentials.
func (g *GitLabProvider) Authenticate(ctx context.Context, creds provider.Credentials) error {
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

// CreateRepository creates a project under req.Owner (group/namespace).
// Owner is required — fail-fast when empty.
func (g *GitLabProvider) CreateRepository(ctx context.Context, req provider.CreateRepoRequest) (*provider.Repository, error) {
	if req.Owner == "" {
		return nil, fmt.Errorf("owner is required for create repository")
	}
	if req.Name == "" {
		return nil, fmt.Errorf("name is required for create repository")
	}

	namespaceID, err := g.resolveNamespaceID(ctx, req.Owner)
	if err != nil {
		return nil, g.FormatError("create repository", err)
	}

	visibility := "public"
	if req.Private {
		visibility = "private"
	}
	switch req.Visibility {
	case provider.VisibilityPrivate:
		visibility = "private"
	case provider.VisibilityPublic:
		visibility = "public"
	case provider.VisibilityInternal:
		visibility = "internal"
	}

	payload := map[string]any{
		"name":                   req.Name,
		"description":            req.Description,
		"visibility":             visibility,
		"initialize_with_readme": req.AutoInit,
		"namespace_id":           namespaceID,
	}
	if req.DefaultBranch != "" {
		payload["default_branch"] = req.DefaultBranch
	}

	var project gitlabProject
	if err := g.doJSON(ctx, "POST", "projects", payload, &project, http.StatusCreated, http.StatusOK); err != nil {
		return nil, g.FormatError("create repository", err)
	}
	return gitlabProjectToProvider(&project), nil
}

// UpdateRepository updates project settings via PUT /projects/:id.
// id must be owner/repo (path with namespace).
func (g *GitLabProvider) UpdateRepository(ctx context.Context, id string, updates provider.UpdateRepoRequest) (*provider.Repository, error) {
	if id == "" {
		return nil, fmt.Errorf("repository id is required")
	}

	payload := map[string]any{}
	if updates.Name != nil {
		payload["name"] = *updates.Name
	}
	if updates.Description != nil {
		payload["description"] = *updates.Description
	}
	if updates.DefaultBranch != nil {
		payload["default_branch"] = *updates.DefaultBranch
	}
	if updates.Archived != nil {
		payload["archived"] = *updates.Archived
	}
	if updates.Private != nil {
		if *updates.Private {
			payload["visibility"] = "private"
		} else {
			payload["visibility"] = "public"
		}
	}
	switch updates.Visibility {
	case provider.VisibilityPrivate:
		payload["visibility"] = "private"
	case provider.VisibilityPublic:
		payload["visibility"] = "public"
	case provider.VisibilityInternal:
		payload["visibility"] = "internal"
	}

	encoded := url.PathEscape(id)
	var project gitlabProject
	if err := g.doJSON(ctx, "PUT", "projects/"+encoded, payload, &project, http.StatusOK); err != nil {
		return nil, g.FormatError("update repository", err)
	}
	return gitlabProjectToProvider(&project), nil
}

// DeleteRepository deletes a project. id is owner/repo (path with namespace).
func (g *GitLabProvider) DeleteRepository(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("repository id is required")
	}
	encoded := url.PathEscape(id)
	if err := g.doJSON(ctx, "DELETE", "projects/"+encoded, nil, nil, http.StatusAccepted, http.StatusNoContent, http.StatusOK); err != nil {
		return g.FormatError("delete repository", err)
	}
	return nil
}

// ArchiveRepository archives a project. id is owner/repo.
func (g *GitLabProvider) ArchiveRepository(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("repository id is required")
	}
	encoded := url.PathEscape(id)
	if err := g.doJSON(ctx, "POST", "projects/"+encoded+"/archive", nil, nil, http.StatusCreated, http.StatusOK); err != nil {
		return g.FormatError("archive repository", err)
	}
	return nil
}

// UnarchiveRepository unarchives a project. id is owner/repo.
func (g *GitLabProvider) UnarchiveRepository(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("repository id is required")
	}
	encoded := url.PathEscape(id)
	if err := g.doJSON(ctx, "POST", "projects/"+encoded+"/unarchive", nil, nil, http.StatusCreated, http.StatusOK); err != nil {
		return g.FormatError("unarchive repository", err)
	}
	return nil
}

// ForkRepository forks a project via POST /projects/:id/fork.
// id must be owner/repo (path with namespace).
func (g *GitLabProvider) ForkRepository(ctx context.Context, id string, opts provider.ForkOptions) (*provider.Repository, error) {
	if id == "" {
		return nil, fmt.Errorf("repository id is required")
	}

	payload := map[string]any{}
	if opts.Organization != "" {
		payload["namespace_path"] = opts.Organization
	}
	if opts.Name != "" {
		payload["name"] = opts.Name
		payload["path"] = opts.Name
	}

	encoded := url.PathEscape(id)
	var project gitlabProject
	if err := g.doJSON(ctx, "POST", "projects/"+encoded+"/fork", payload, &project, http.StatusCreated, http.StatusOK, http.StatusAccepted); err != nil {
		return nil, g.FormatError("fork repository", err)
	}
	return gitlabProjectToProvider(&project), nil
}

// SearchRepositories searches projects via the GitLab projects API.
func (g *GitLabProvider) SearchRepositories(ctx context.Context, query provider.SearchQuery) (*provider.SearchResult, error) {
	q := query.Query
	if q == "" {
		return nil, fmt.Errorf("search query is required")
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
	params.Set("search", q)
	params.Set("page", fmt.Sprintf("%d", page))
	params.Set("per_page", fmt.Sprintf("%d", perPage))
	if query.Order != "" {
		params.Set("sort", query.Order)
	}
	if query.Organization != "" {
		params.Set("namespace_path", query.Organization)
	}

	var projects []gitlabProject
	if err := g.doJSON(ctx, "GET", "projects?"+params.Encode(), nil, &projects, http.StatusOK); err != nil {
		return nil, g.FormatError("search repositories", err)
	}

	repos := make([]provider.Repository, 0, len(projects))
	for i := range projects {
		repos = append(repos, *gitlabProjectToProvider(&projects[i]))
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

// gitlabProjectHook is the GitLab project hook API shape.
type gitlabProjectHook struct {
	ID         int    `json:"id"`
	URL        string `json:"url"`
	PushEvents bool   `json:"push_events"`
	EnableSSL  bool   `json:"enable_ssl_verification"`
	CreatedAt  string `json:"created_at"`
}

// Webhook management — GitLab project hooks API.
func (g *GitLabProvider) ListWebhooks(ctx context.Context, repoID string) ([]provider.Webhook, error) {
	if repoID == "" {
		return nil, fmt.Errorf("repository id is required")
	}
	encoded := url.PathEscape(repoID)
	var hooks []gitlabProjectHook
	if err := g.doJSON(ctx, "GET", "projects/"+encoded+"/hooks", nil, &hooks, http.StatusOK); err != nil {
		return nil, g.FormatError("list webhooks", err)
	}
	out := make([]provider.Webhook, 0, len(hooks))
	for i := range hooks {
		out = append(out, gitlabHookToProvider(&hooks[i]))
	}
	return out, nil
}

func (g *GitLabProvider) GetWebhook(ctx context.Context, repoID, webhookID string) (*provider.Webhook, error) {
	if repoID == "" || webhookID == "" {
		return nil, fmt.Errorf("repository id and webhook id are required")
	}
	encoded := url.PathEscape(repoID)
	var hook gitlabProjectHook
	if err := g.doJSON(ctx, "GET", "projects/"+encoded+"/hooks/"+webhookID, nil, &hook, http.StatusOK); err != nil {
		return nil, g.FormatError("get webhook", err)
	}
	w := gitlabHookToProvider(&hook)
	return &w, nil
}

func (g *GitLabProvider) CreateWebhook(ctx context.Context, repoID string, webhook provider.CreateWebhookRequest) (*provider.Webhook, error) {
	if err := g.ValidateWebhookURL(ctx, webhook.Config.URL); err != nil {
		return nil, err
	}
	if repoID == "" {
		return nil, fmt.Errorf("repository id is required")
	}
	payload := map[string]any{
		"url":                     webhook.Config.URL,
		"push_events":             true,
		"enable_ssl_verification": !webhook.Config.InsecureSSL,
	}
	if webhook.Config.Secret != "" {
		payload["token"] = webhook.Config.Secret
	}
	encoded := url.PathEscape(repoID)
	var hook gitlabProjectHook
	if err := g.doJSON(ctx, "POST", "projects/"+encoded+"/hooks", payload, &hook, http.StatusCreated, http.StatusOK); err != nil {
		return nil, g.FormatError("create webhook", err)
	}
	w := gitlabHookToProvider(&hook)
	return &w, nil
}

func (g *GitLabProvider) UpdateWebhook(ctx context.Context, repoID, webhookID string, updates provider.UpdateWebhookRequest) (*provider.Webhook, error) {
	if repoID == "" || webhookID == "" {
		return nil, fmt.Errorf("repository id and webhook id are required")
	}
	payload := map[string]any{}
	if updates.Config != nil {
		if updates.Config.URL != "" {
			if err := g.ValidateWebhookURL(ctx, updates.Config.URL); err != nil {
				return nil, err
			}
			payload["url"] = updates.Config.URL
		}
		payload["enable_ssl_verification"] = !updates.Config.InsecureSSL
		if updates.Config.Secret != "" {
			payload["token"] = updates.Config.Secret
		}
	}
	if updates.Active != nil {
		// GitLab has no active flag; map to push_events for best-effort.
		payload["push_events"] = *updates.Active
	}
	encoded := url.PathEscape(repoID)
	var hook gitlabProjectHook
	if err := g.doJSON(ctx, "PUT", "projects/"+encoded+"/hooks/"+webhookID, payload, &hook, http.StatusOK); err != nil {
		return nil, g.FormatError("update webhook", err)
	}
	w := gitlabHookToProvider(&hook)
	return &w, nil
}

func (g *GitLabProvider) DeleteWebhook(ctx context.Context, repoID, webhookID string) error {
	if repoID == "" || webhookID == "" {
		return fmt.Errorf("repository id and webhook id are required")
	}
	encoded := url.PathEscape(repoID)
	if err := g.doJSON(ctx, "DELETE", "projects/"+encoded+"/hooks/"+webhookID, nil, nil, http.StatusNoContent, http.StatusOK); err != nil {
		return g.FormatError("delete webhook", err)
	}
	return nil
}

func (g *GitLabProvider) TestWebhook(ctx context.Context, repoID, webhookID string) (*provider.WebhookTestResult, error) {
	// GitLab has no direct "test hook" equivalent in all versions; report not supported clearly.
	return nil, fmt.Errorf("test webhook not supported by GitLab project hooks API")
}

func (g *GitLabProvider) ValidateWebhookURL(_ context.Context, webhookURL string) error {
	if strings.TrimSpace(webhookURL) == "" {
		return fmt.Errorf("webhook URL cannot be empty")
	}
	if !strings.HasPrefix(webhookURL, "http://") && !strings.HasPrefix(webhookURL, "https://") {
		return fmt.Errorf("webhook URL must be a valid HTTP/HTTPS URL")
	}
	return nil
}

func gitlabHookToProvider(h *gitlabProjectHook) provider.Webhook {
	return provider.Webhook{
		ID:     fmt.Sprintf("%d", h.ID),
		Name:   "project_hook",
		URL:    h.URL,
		Active: h.PushEvents,
		Events: []string{"push"},
		Config: provider.WebhookConfig{
			URL:         h.URL,
			ContentType: "json",
			InsecureSSL: !h.EnableSSL,
		},
	}
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
		Details:     make(map[string]any),
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

// Release management

// ListReleases lists releases for a repository.
func (g *GitLabProvider) ListReleases(ctx context.Context, repoID string, opts provider.ListReleasesOptions) (*provider.ReleaseList, error) {
	return ListReleases(ctx, repoID, opts)
}

// GetRelease gets a specific release by ID.
// Note: GitLab uses tag name as release ID.
func (g *GitLabProvider) GetRelease(ctx context.Context, repoID, releaseID string) (*provider.Release, error) {
	return GetRelease(ctx, repoID, releaseID)
}

// GetReleaseByTag gets a release by tag name.
func (g *GitLabProvider) GetReleaseByTag(ctx context.Context, repoID, tagName string) (*provider.Release, error) {
	return GetRelease(ctx, repoID, tagName)
}

// CreateRelease creates a new release.
func (g *GitLabProvider) CreateRelease(ctx context.Context, repoID string, req provider.CreateReleaseRequest) (*provider.Release, error) {
	return CreateRelease(ctx, repoID, req)
}

// UpdateRelease updates an existing release.
// Note: GitLab uses tag name as release ID.
func (g *GitLabProvider) UpdateRelease(ctx context.Context, repoID, releaseID string, updates provider.UpdateReleaseRequest) (*provider.Release, error) {
	return UpdateRelease(ctx, repoID, releaseID, updates)
}

// DeleteRelease deletes a release.
// Note: GitLab uses tag name as release ID.
func (g *GitLabProvider) DeleteRelease(ctx context.Context, repoID, releaseID string) error {
	return DeleteRelease(ctx, repoID, releaseID)
}

// ListReleaseAssets lists assets for a release.
func (g *GitLabProvider) ListReleaseAssets(ctx context.Context, repoID, releaseID string) ([]provider.Asset, error) {
	// GitLab release에서 직접 assets를 가져옴
	release, err := GetRelease(ctx, repoID, releaseID)
	if err != nil {
		return nil, err
	}
	return release.Assets, nil
}

// UploadReleaseAsset uploads an asset to a release.
// not in CLI surface; deferred (issue 26 phase 2+)
func (g *GitLabProvider) UploadReleaseAsset(ctx context.Context, repoID string, req provider.UploadAssetRequest) (*provider.Asset, error) {
	return nil, fmt.Errorf("GitLab release asset upload not implemented: not in CLI surface; deferred (issue 26 phase 2+)")
}

// DeleteReleaseAsset deletes a release asset.
// not in CLI surface; deferred (issue 26 phase 2+)
func (g *GitLabProvider) DeleteReleaseAsset(ctx context.Context, repoID, assetID string) error {
	return fmt.Errorf("GitLab release asset deletion not implemented: not in CLI surface; deferred (issue 26 phase 2+)")
}

// DownloadReleaseAsset downloads a release asset.
// not in CLI surface; deferred (issue 26 phase 2+)
func (g *GitLabProvider) DownloadReleaseAsset(ctx context.Context, repoID, assetID string) ([]byte, error) {
	return nil, fmt.Errorf("GitLab release asset download not implemented: not in CLI surface; deferred (issue 26 phase 2+)")
}

// gitlabProject is the subset of GitLab project JSON used by mutation APIs.
type gitlabProject struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	PathWithNamespace string `json:"path_with_namespace"`
	Description       string `json:"description"`
	DefaultBranch     string `json:"default_branch"`
	HTTPURLToRepo     string `json:"http_url_to_repo"`
	SSHURLToRepo      string `json:"ssh_url_to_repo"`
	WebURL            string `json:"web_url"`
	Visibility        string `json:"visibility"`
	Archived          bool   `json:"archived"`
}

func gitlabProjectToProvider(p *gitlabProject) *provider.Repository {
	if p == nil {
		return nil
	}
	fullName := p.PathWithNamespace
	if fullName == "" {
		fullName = p.Name
	}
	return &provider.Repository{
		ID:            fullName,
		Name:          p.Name,
		FullName:      fullName,
		Description:   p.Description,
		DefaultBranch: p.DefaultBranch,
		CloneURL:      p.HTTPURLToRepo,
		SSHURL:        p.SSHURLToRepo,
		HTMLURL:       p.WebURL,
		Private:       p.Visibility == "private",
		Archived:      p.Archived,
		ProviderType:  "gitlab",
	}
}

// resolveNamespaceID looks up a GitLab namespace (group or user) by path.
func (g *GitLabProvider) resolveNamespaceID(ctx context.Context, owner string) (int, error) {
	var ns struct {
		ID int `json:"id"`
	}
	path := "namespaces/" + url.PathEscape(owner)
	if err := g.doJSON(ctx, "GET", path, nil, &ns, http.StatusOK); err != nil {
		return 0, fmt.Errorf("failed to resolve namespace %q: %w", owner, err)
	}
	if ns.ID == 0 {
		return 0, fmt.Errorf("namespace %q not found", owner)
	}
	return ns.ID, nil
}

// apiBase returns the provider API base URL (instance-configured).
func (g *GitLabProvider) apiBase() string {
	base := strings.TrimRight(g.GetBaseURL(), "/")
	if base == "" {
		base = getBaseAPIURL()
	}
	return base
}

// doJSON performs an authenticated JSON request against this provider's base URL.
// Any of wantStatuses is treated as success.
// endpoint may include PathEscape'd segments (projects/acme%2Frepo) and query strings.
// url.Parse preserves %2F via RawPath so GitLab receives a single project path segment.
func (g *GitLabProvider) doJSON(ctx context.Context, method, endpoint string, payload any, out any, wantStatuses ...int) error {
	var bodyReader io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to encode request: %w", err)
		}
		bodyReader = bytes.NewReader(encoded)
	}

	fullURL := strings.TrimRight(g.apiBase(), "/") + "/" + strings.TrimPrefix(endpoint, "/")
	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	token := g.GetToken()
	if token == "" {
		token = configuredToken
	}
	if token != "" {
		req.Header.Set("PRIVATE-TOKEN", token)
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := httpclient.GetGlobalClient("gitlab")
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
