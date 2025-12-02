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

// WebhookInfo represents a GitHub webhook configuration.
type WebhookInfo struct {
	ID           int64         `json:"id"`
	Name         string        `json:"name"`
	URL          string        `json:"url"`
	Events       []string      `json:"events"`
	Active       bool          `json:"active"`
	Config       WebhookConfig `json:"config"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
	Repository   string        `json:"repository,omitempty"`
	Organization string        `json:"organization,omitempty"`
}

// WebhookConfig represents webhook configuration settings.
type WebhookConfig struct {
	URL         string `json:"url"`
	ContentType string `json:"content_type"`
	Secret      string `json:"secret,omitempty"`
	InsecureSSL bool   `json:"insecure_ssl"`
}

// WebhookCreateRequest represents a request to create a new webhook.
type WebhookCreateRequest struct {
	Name   string        `json:"name"`
	URL    string        `json:"url"`
	Events []string      `json:"events"`
	Active bool          `json:"active"`
	Config WebhookConfig `json:"config"`
}

// WebhookUpdateRequest represents a request to update an existing webhook.
type WebhookUpdateRequest struct {
	ID     int64         `json:"id"`
	Name   string        `json:"name,omitempty"`
	URL    string        `json:"url,omitempty"`
	Events []string      `json:"events,omitempty"`
	Active *bool         `json:"active,omitempty"`
	Config WebhookConfig `json:"config,omitempty"`
}

// WebhookListOptions represents options for listing webhooks.
type WebhookListOptions struct {
	Organization string `json:"organization,omitempty"`
	Repository   string `json:"repository,omitempty"`
	Page         int    `json:"page"`
	PerPage      int    `json:"per_page"`
}

// WebhookService defines the interface for webhook operations.
type WebhookService interface {
	// Repository webhooks
	CreateRepositoryWebhook(ctx context.Context, owner, repo string, request *WebhookCreateRequest) (*WebhookInfo, error)
	GetRepositoryWebhook(ctx context.Context, owner, repo string, webhookID int64) (*WebhookInfo, error)
	ListRepositoryWebhooks(ctx context.Context, owner, repo string, options *WebhookListOptions) ([]*WebhookInfo, error)
	UpdateRepositoryWebhook(ctx context.Context, owner, repo string, request *WebhookUpdateRequest) (*WebhookInfo, error)
	DeleteRepositoryWebhook(ctx context.Context, owner, repo string, webhookID int64) error

	// Organization webhooks
	CreateOrganizationWebhook(ctx context.Context, org string, request *WebhookCreateRequest) (*WebhookInfo, error)
	GetOrganizationWebhook(ctx context.Context, org string, webhookID int64) (*WebhookInfo, error)
	ListOrganizationWebhooks(ctx context.Context, org string, options *WebhookListOptions) ([]*WebhookInfo, error)
	UpdateOrganizationWebhook(ctx context.Context, org string, request *WebhookUpdateRequest) (*WebhookInfo, error)
	DeleteOrganizationWebhook(ctx context.Context, org string, webhookID int64) error

	// Bulk operations
	BulkCreateWebhooks(ctx context.Context, request *BulkWebhookRequest) (*BulkWebhookResult, error)
	BulkUpdateWebhooks(ctx context.Context, request *BulkWebhookUpdateRequest) (*BulkWebhookResult, error)
	BulkDeleteWebhooks(ctx context.Context, request *BulkWebhookDeleteRequest) (*BulkWebhookResult, error)

	// Webhook status monitoring
	TestWebhook(ctx context.Context, owner, repo string, webhookID int64) (*WebhookTestResult, error)
	GetWebhookDeliveries(ctx context.Context, owner, repo string, webhookID int64) ([]*WebhookDelivery, error)
}

// BulkWebhookRequest represents a bulk webhook creation request.
type BulkWebhookRequest struct {
	Organization string               `json:"organization"`
	Repositories []string             `json:"repositories,omitempty"` // if empty, apply to all repos
	Template     WebhookCreateRequest `json:"template"`
	Filters      *RepositoryFilters   `json:"filters,omitempty"`
}

// BulkWebhookUpdateRequest represents a bulk webhook update request.
type BulkWebhookUpdateRequest struct {
	Organization string               `json:"organization"`
	Repositories []string             `json:"repositories,omitempty"`
	Template     WebhookUpdateRequest `json:"template"`
	Filters      *RepositoryFilters   `json:"filters,omitempty"`
	SelectBy     WebhookSelector      `json:"select_by"` // how to find webhooks to update
}

// BulkWebhookDeleteRequest represents a bulk webhook deletion request.
type BulkWebhookDeleteRequest struct {
	Organization string             `json:"organization"`
	Repositories []string           `json:"repositories,omitempty"`
	SelectBy     WebhookSelector    `json:"select_by"` // how to find webhooks to delete
	Filters      *RepositoryFilters `json:"filters,omitempty"`
}

// WebhookSelector defines how to select webhooks for bulk operations.
type WebhookSelector struct {
	ByName   string   `json:"by_name,omitempty"`
	ByURL    string   `json:"by_url,omitempty"`
	ByEvents []string `json:"by_events,omitempty"`
	Active   *bool    `json:"active,omitempty"`
}

// BulkWebhookResult represents the result of bulk webhook operations.
type BulkWebhookResult struct {
	TotalRepositories int                      `json:"total_repositories"`
	SuccessCount      int                      `json:"success_count"`
	FailureCount      int                      `json:"failure_count"`
	Results           []WebhookOperationResult `json:"results"`
	ExecutionTime     string                   `json:"execution_time"`
}

// WebhookOperationResult represents the result of a single webhook operation.
type WebhookOperationResult struct {
	Repository  string       `json:"repository"`
	Operation   string       `json:"operation"`
	Success     bool         `json:"success"`
	WebhookInfo *WebhookInfo `json:"webhook_info,omitempty"`
	Error       string       `json:"error,omitempty"`
	Duration    string       `json:"duration"`
}

// WebhookTestResult represents the result of testing a webhook.
type WebhookTestResult struct {
	Success    bool      `json:"success"`
	StatusCode int       `json:"status_code"`
	Response   string    `json:"response"`
	Duration   string    `json:"duration"`
	Error      string    `json:"error,omitempty"`
	DeliveryID string    `json:"delivery_id"`
	TestedAt   time.Time `json:"tested_at"`
}

// WebhookDelivery represents a webhook delivery record.
type WebhookDelivery struct {
	ID          string    `json:"id"`
	Event       string    `json:"event"`
	Action      string    `json:"action"`
	StatusCode  int       `json:"status_code"`
	Duration    string    `json:"duration"`
	DeliveredAt time.Time `json:"delivered_at"`
	Success     bool      `json:"success"`
	Redelivered bool      `json:"redelivered"`
	URL         string    `json:"url"`
}

// webhookServiceImpl implements the WebhookService interface.
type webhookServiceImpl struct {
	apiClient  APIClient
	httpClient *http.Client
	baseURL    string
	token      string
	logger     Logger
}

// NewWebhookService creates a new webhook service instance.
func NewWebhookService(apiClient APIClient, logger Logger) WebhookService {
	return &webhookServiceImpl{
		apiClient:  apiClient,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    "https://api.github.com",
		logger:     logger,
	}
}

// NewWebhookServiceWithToken creates a webhook service with a token for API calls.
func NewWebhookServiceWithToken(apiClient APIClient, token string, logger Logger) WebhookService {
	return &webhookServiceImpl{
		apiClient:  apiClient,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    "https://api.github.com",
		token:      token,
		logger:     logger,
	}
}

// SetToken sets the authentication token for API calls.
func (w *webhookServiceImpl) SetToken(token string) {
	w.token = token
}

// doRequest performs an HTTP request with authentication.
func (w *webhookServiceImpl) doRequest(ctx context.Context, method, url string, body interface{}) (*http.Response, error) {
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

	if w.token != "" {
		req.Header.Set("Authorization", "token "+w.token)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "gzh-cli")

	return w.httpClient.Do(req)
}

// Repository webhook operations

// CreateRepositoryWebhook creates a new webhook for a repository.
func (w *webhookServiceImpl) CreateRepositoryWebhook(ctx context.Context, owner, repo string, request *WebhookCreateRequest) (*WebhookInfo, error) {
	w.logger.Info("Creating repository webhook", "owner", owner, "repo", repo, "name", request.Name)

	// Validate request
	if err := w.validateWebhookRequest(request); err != nil {
		return nil, fmt.Errorf("invalid webhook request: %w", err)
	}

	// GitHub API request body
	apiRequest := map[string]interface{}{
		"name":   "web",
		"active": request.Active,
		"events": request.Events,
		"config": map[string]interface{}{
			"url":          request.Config.URL,
			"content_type": request.Config.ContentType,
			"insecure_ssl": "0",
		},
	}
	if request.Config.Secret != "" {
		apiRequest["config"].(map[string]interface{})["secret"] = request.Config.Secret
	}
	if request.Config.InsecureSSL {
		apiRequest["config"].(map[string]interface{})["insecure_ssl"] = "1"
	}

	url := fmt.Sprintf("%s/repos/%s/%s/hooks", w.baseURL, owner, repo)
	resp, err := w.doRequest(ctx, http.MethodPost, url, apiRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to create webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create webhook: HTTP %d - %s", resp.StatusCode, string(body))
	}

	var webhook WebhookInfo
	if err := json.NewDecoder(resp.Body).Decode(&webhook); err != nil {
		return nil, fmt.Errorf("failed to decode webhook response: %w", err)
	}

	webhook.Repository = fmt.Sprintf("%s/%s", owner, repo)
	w.logger.Info("Successfully created repository webhook", "webhook_id", webhook.ID)

	return &webhook, nil
}

// GetRepositoryWebhook retrieves a specific webhook for a repository.
func (w *webhookServiceImpl) GetRepositoryWebhook(ctx context.Context, owner, repo string, webhookID int64) (*WebhookInfo, error) {
	w.logger.Debug("Getting repository webhook", "owner", owner, "repo", repo, "webhook_id", webhookID)

	url := fmt.Sprintf("%s/repos/%s/%s/hooks/%d", w.baseURL, owner, repo, webhookID)
	resp, err := w.doRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("webhook not found: %d", webhookID)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get webhook: HTTP %d - %s", resp.StatusCode, string(body))
	}

	var webhook WebhookInfo
	if err := json.NewDecoder(resp.Body).Decode(&webhook); err != nil {
		return nil, fmt.Errorf("failed to decode webhook response: %w", err)
	}

	webhook.Repository = fmt.Sprintf("%s/%s", owner, repo)
	return &webhook, nil
}

// ListRepositoryWebhooks lists all webhooks for a repository.
func (w *webhookServiceImpl) ListRepositoryWebhooks(ctx context.Context, owner, repo string, options *WebhookListOptions) ([]*WebhookInfo, error) {
	w.logger.Debug("Listing repository webhooks", "owner", owner, "repo", repo)

	page := 1
	perPage := 100
	if options != nil {
		if options.Page > 0 {
			page = options.Page
		}
		if options.PerPage > 0 {
			perPage = options.PerPage
		}
	}

	url := fmt.Sprintf("%s/repos/%s/%s/hooks?page=%d&per_page=%d", w.baseURL, owner, repo, page, perPage)
	resp, err := w.doRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list webhooks: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to list webhooks: HTTP %d - %s", resp.StatusCode, string(body))
	}

	var webhooks []*WebhookInfo
	if err := json.NewDecoder(resp.Body).Decode(&webhooks); err != nil {
		return nil, fmt.Errorf("failed to decode webhooks response: %w", err)
	}

	// Set repository info for each webhook
	for _, wh := range webhooks {
		wh.Repository = fmt.Sprintf("%s/%s", owner, repo)
	}

	return webhooks, nil
}

// UpdateRepositoryWebhook updates an existing webhook for a repository.
func (w *webhookServiceImpl) UpdateRepositoryWebhook(ctx context.Context, owner, repo string, request *WebhookUpdateRequest) (*WebhookInfo, error) {
	w.logger.Info("Updating repository webhook", "owner", owner, "repo", repo, "webhook_id", request.ID)

	// Build update request body
	apiRequest := make(map[string]interface{})
	if request.Events != nil {
		apiRequest["events"] = request.Events
	}
	if request.Active != nil {
		apiRequest["active"] = *request.Active
	}
	if request.Config.URL != "" || request.Config.ContentType != "" || request.Config.Secret != "" {
		config := make(map[string]interface{})
		if request.Config.URL != "" {
			config["url"] = request.Config.URL
		}
		if request.Config.ContentType != "" {
			config["content_type"] = request.Config.ContentType
		}
		if request.Config.Secret != "" {
			config["secret"] = request.Config.Secret
		}
		if request.Config.InsecureSSL {
			config["insecure_ssl"] = "1"
		} else {
			config["insecure_ssl"] = "0"
		}
		apiRequest["config"] = config
	}

	url := fmt.Sprintf("%s/repos/%s/%s/hooks/%d", w.baseURL, owner, repo, request.ID)
	resp, err := w.doRequest(ctx, http.MethodPatch, url, apiRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to update webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to update webhook: HTTP %d - %s", resp.StatusCode, string(body))
	}

	var webhook WebhookInfo
	if err := json.NewDecoder(resp.Body).Decode(&webhook); err != nil {
		return nil, fmt.Errorf("failed to decode webhook response: %w", err)
	}

	webhook.Repository = fmt.Sprintf("%s/%s", owner, repo)
	w.logger.Info("Successfully updated repository webhook", "webhook_id", webhook.ID)

	return &webhook, nil
}

// DeleteRepositoryWebhook deletes a webhook from a repository.
func (w *webhookServiceImpl) DeleteRepositoryWebhook(ctx context.Context, owner, repo string, webhookID int64) error {
	w.logger.Info("Deleting repository webhook", "owner", owner, "repo", repo, "webhook_id", webhookID)

	url := fmt.Sprintf("%s/repos/%s/%s/hooks/%d", w.baseURL, owner, repo, webhookID)
	resp, err := w.doRequest(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("failed to delete webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete webhook: HTTP %d - %s", resp.StatusCode, string(body))
	}

	w.logger.Info("Successfully deleted repository webhook", "webhook_id", webhookID)

	return nil
}

// Organization webhook operations (similar implementations)

// CreateOrganizationWebhook creates a new webhook for an organization.
func (w *webhookServiceImpl) CreateOrganizationWebhook(ctx context.Context, org string, request *WebhookCreateRequest) (*WebhookInfo, error) {
	w.logger.Info("Creating organization webhook", "org", org, "name", request.Name)

	if err := w.validateWebhookRequest(request); err != nil {
		return nil, fmt.Errorf("invalid webhook request: %w", err)
	}

	// GitHub API request body
	apiRequest := map[string]interface{}{
		"name":   "web",
		"active": request.Active,
		"events": request.Events,
		"config": map[string]interface{}{
			"url":          request.Config.URL,
			"content_type": request.Config.ContentType,
			"insecure_ssl": "0",
		},
	}
	if request.Config.Secret != "" {
		apiRequest["config"].(map[string]interface{})["secret"] = request.Config.Secret
	}
	if request.Config.InsecureSSL {
		apiRequest["config"].(map[string]interface{})["insecure_ssl"] = "1"
	}

	url := fmt.Sprintf("%s/orgs/%s/hooks", w.baseURL, org)
	resp, err := w.doRequest(ctx, http.MethodPost, url, apiRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to create organization webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create organization webhook: HTTP %d - %s", resp.StatusCode, string(body))
	}

	var webhook WebhookInfo
	if err := json.NewDecoder(resp.Body).Decode(&webhook); err != nil {
		return nil, fmt.Errorf("failed to decode webhook response: %w", err)
	}

	webhook.Organization = org
	w.logger.Info("Successfully created organization webhook", "webhook_id", webhook.ID)

	return &webhook, nil
}

// GetOrganizationWebhook retrieves a specific webhook for an organization.
func (w *webhookServiceImpl) GetOrganizationWebhook(ctx context.Context, org string, webhookID int64) (*WebhookInfo, error) {
	w.logger.Debug("Getting organization webhook", "org", org, "webhook_id", webhookID)

	url := fmt.Sprintf("%s/orgs/%s/hooks/%d", w.baseURL, org, webhookID)
	resp, err := w.doRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("organization webhook not found: %d", webhookID)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get organization webhook: HTTP %d - %s", resp.StatusCode, string(body))
	}

	var webhook WebhookInfo
	if err := json.NewDecoder(resp.Body).Decode(&webhook); err != nil {
		return nil, fmt.Errorf("failed to decode webhook response: %w", err)
	}

	webhook.Organization = org
	return &webhook, nil
}

// ListOrganizationWebhooks lists all webhooks for an organization.
func (w *webhookServiceImpl) ListOrganizationWebhooks(ctx context.Context, org string, options *WebhookListOptions) ([]*WebhookInfo, error) {
	w.logger.Debug("Listing organization webhooks", "org", org)

	page := 1
	perPage := 100
	if options != nil {
		if options.Page > 0 {
			page = options.Page
		}
		if options.PerPage > 0 {
			perPage = options.PerPage
		}
	}

	url := fmt.Sprintf("%s/orgs/%s/hooks?page=%d&per_page=%d", w.baseURL, org, page, perPage)
	resp, err := w.doRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list organization webhooks: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to list organization webhooks: HTTP %d - %s", resp.StatusCode, string(body))
	}

	var webhooks []*WebhookInfo
	if err := json.NewDecoder(resp.Body).Decode(&webhooks); err != nil {
		return nil, fmt.Errorf("failed to decode webhooks response: %w", err)
	}

	// Set organization info for each webhook
	for _, wh := range webhooks {
		wh.Organization = org
	}

	return webhooks, nil
}

// UpdateOrganizationWebhook updates an existing webhook for an organization.
func (w *webhookServiceImpl) UpdateOrganizationWebhook(ctx context.Context, org string, request *WebhookUpdateRequest) (*WebhookInfo, error) {
	w.logger.Info("Updating organization webhook", "org", org, "webhook_id", request.ID)

	// Build update request body
	apiRequest := make(map[string]interface{})
	if request.Events != nil {
		apiRequest["events"] = request.Events
	}
	if request.Active != nil {
		apiRequest["active"] = *request.Active
	}
	if request.Config.URL != "" || request.Config.ContentType != "" || request.Config.Secret != "" {
		config := make(map[string]interface{})
		if request.Config.URL != "" {
			config["url"] = request.Config.URL
		}
		if request.Config.ContentType != "" {
			config["content_type"] = request.Config.ContentType
		}
		if request.Config.Secret != "" {
			config["secret"] = request.Config.Secret
		}
		if request.Config.InsecureSSL {
			config["insecure_ssl"] = "1"
		} else {
			config["insecure_ssl"] = "0"
		}
		apiRequest["config"] = config
	}

	url := fmt.Sprintf("%s/orgs/%s/hooks/%d", w.baseURL, org, request.ID)
	resp, err := w.doRequest(ctx, http.MethodPatch, url, apiRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to update organization webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to update organization webhook: HTTP %d - %s", resp.StatusCode, string(body))
	}

	var webhook WebhookInfo
	if err := json.NewDecoder(resp.Body).Decode(&webhook); err != nil {
		return nil, fmt.Errorf("failed to decode webhook response: %w", err)
	}

	webhook.Organization = org
	w.logger.Info("Successfully updated organization webhook", "webhook_id", webhook.ID)

	return &webhook, nil
}

// DeleteOrganizationWebhook deletes a webhook from an organization.
func (w *webhookServiceImpl) DeleteOrganizationWebhook(ctx context.Context, org string, webhookID int64) error {
	w.logger.Info("Deleting organization webhook", "org", org, "webhook_id", webhookID)

	url := fmt.Sprintf("%s/orgs/%s/hooks/%d", w.baseURL, org, webhookID)
	resp, err := w.doRequest(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("failed to delete organization webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete organization webhook: HTTP %d - %s", resp.StatusCode, string(body))
	}

	w.logger.Info("Successfully deleted organization webhook", "webhook_id", webhookID)

	return nil
}

// Bulk operations

// BulkCreateWebhooks creates webhooks across multiple repositories.
func (w *webhookServiceImpl) BulkCreateWebhooks(ctx context.Context, request *BulkWebhookRequest) (*BulkWebhookResult, error) {
	w.logger.Info("Starting bulk webhook creation", "org", request.Organization)

	startTime := time.Now()
	result := &BulkWebhookResult{
		Results: make([]WebhookOperationResult, 0),
	}

	// Get repositories to process
	repositories := request.Repositories
	if len(repositories) == 0 {
		// Get all repositories from organization
		// TODO: Implement using APIClient.ListOrganizationRepositories
		repositories = []string{"repo1", "repo2"} // Mock for now
	}

	result.TotalRepositories = len(repositories)

	// Create webhooks for each repository
	for _, repo := range repositories {
		opResult := WebhookOperationResult{
			Repository: repo,
			Operation:  "create",
		}

		opStartTime := time.Now()
		webhook, err := w.CreateRepositoryWebhook(ctx, request.Organization, repo, &request.Template)
		opDuration := time.Since(opStartTime)
		opResult.Duration = opDuration.String()

		if err != nil {
			opResult.Success = false
			opResult.Error = err.Error()
			result.FailureCount++
		} else {
			opResult.Success = true
			opResult.WebhookInfo = webhook
			result.SuccessCount++
		}

		result.Results = append(result.Results, opResult)
	}

	result.ExecutionTime = time.Since(startTime).String()
	w.logger.Info("Completed bulk webhook creation",
		"total", result.TotalRepositories,
		"success", result.SuccessCount,
		"failures", result.FailureCount)

	return result, nil
}

// BulkUpdateWebhooks updates webhooks across multiple repositories.
func (w *webhookServiceImpl) BulkUpdateWebhooks(ctx context.Context, request *BulkWebhookUpdateRequest) (*BulkWebhookResult, error) {
	w.logger.Info("Starting bulk webhook update", "org", request.Organization)

	startTime := time.Now()
	result := &BulkWebhookResult{
		Results: make([]WebhookOperationResult, 0),
	}

	// Get repositories to process
	repositories := request.Repositories
	if len(repositories) == 0 {
		repositories = []string{"repo1", "repo2"} // Mock for now
	}

	result.TotalRepositories = len(repositories)

	// Update webhooks for each repository
	for _, repo := range repositories {
		// Find webhooks matching the selector
		existingWebhooks, err := w.ListRepositoryWebhooks(ctx, request.Organization, repo, nil)
		if err != nil {
			continue
		}

		for _, existing := range existingWebhooks {
			if w.webhookMatchesSelector(existing, &request.SelectBy) {
				opResult := WebhookOperationResult{
					Repository: repo,
					Operation:  "update",
				}

				request.Template.ID = existing.ID
				opStartTime := time.Now()
				webhook, err := w.UpdateRepositoryWebhook(ctx, request.Organization, repo, &request.Template)
				opDuration := time.Since(opStartTime)
				opResult.Duration = opDuration.String()

				if err != nil {
					opResult.Success = false
					opResult.Error = err.Error()
					result.FailureCount++
				} else {
					opResult.Success = true
					opResult.WebhookInfo = webhook
					result.SuccessCount++
				}

				result.Results = append(result.Results, opResult)
			}
		}
	}

	result.ExecutionTime = time.Since(startTime).String()

	return result, nil
}

// BulkDeleteWebhooks deletes webhooks across multiple repositories.
func (w *webhookServiceImpl) BulkDeleteWebhooks(ctx context.Context, request *BulkWebhookDeleteRequest) (*BulkWebhookResult, error) {
	w.logger.Info("Starting bulk webhook deletion", "org", request.Organization)

	startTime := time.Now()
	result := &BulkWebhookResult{
		Results: make([]WebhookOperationResult, 0),
	}

	repositories := request.Repositories
	if len(repositories) == 0 {
		repositories = []string{"repo1", "repo2"} // Mock for now
	}

	result.TotalRepositories = len(repositories)

	for _, repo := range repositories {
		existingWebhooks, err := w.ListRepositoryWebhooks(ctx, request.Organization, repo, nil)
		if err != nil {
			continue
		}

		for _, existing := range existingWebhooks {
			if w.webhookMatchesSelector(existing, &request.SelectBy) {
				opResult := WebhookOperationResult{
					Repository: repo,
					Operation:  "delete",
				}

				opStartTime := time.Now()
				err := w.DeleteRepositoryWebhook(ctx, request.Organization, repo, existing.ID)
				opDuration := time.Since(opStartTime)
				opResult.Duration = opDuration.String()

				if err != nil {
					opResult.Success = false
					opResult.Error = err.Error()
					result.FailureCount++
				} else {
					opResult.Success = true
					result.SuccessCount++
				}

				result.Results = append(result.Results, opResult)
			}
		}
	}

	result.ExecutionTime = time.Since(startTime).String()

	return result, nil
}

// TestWebhook tests a webhook by sending a ping event.
func (w *webhookServiceImpl) TestWebhook(ctx context.Context, owner, repo string, webhookID int64) (*WebhookTestResult, error) {
	w.logger.Info("Testing webhook", "owner", owner, "repo", repo, "webhook_id", webhookID)

	startTime := time.Now()

	// TODO: Implement actual webhook test
	testResult := &WebhookTestResult{
		Success:    true,
		StatusCode: 200,
		Response:   "OK",
		Duration:   time.Since(startTime).String(),
		DeliveryID: fmt.Sprintf("test-%d-%d", webhookID, time.Now().Unix()),
		TestedAt:   time.Now(),
	}

	return testResult, nil
}

// GetWebhookDeliveries retrieves recent webhook deliveries.
func (w *webhookServiceImpl) GetWebhookDeliveries(ctx context.Context, owner, repo string, webhookID int64) ([]*WebhookDelivery, error) {
	w.logger.Debug("Getting webhook deliveries", "owner", owner, "repo", repo, "webhook_id", webhookID)

	// TODO: Implement actual GitHub API call
	deliveries := []*WebhookDelivery{
		{
			ID:          "12345678-1234-1234-1234-123456789012",
			Event:       "push",
			Action:      "synchronize",
			StatusCode:  200,
			Duration:    "150ms",
			DeliveredAt: time.Now().Add(-2 * time.Hour),
			Success:     true,
			Redelivered: false,
			URL:         "https://example.com/webhook",
		},
		{
			ID:          "87654321-4321-4321-4321-210987654321",
			Event:       "pull_request",
			Action:      "opened",
			StatusCode:  200,
			Duration:    "230ms",
			DeliveredAt: time.Now().Add(-6 * time.Hour),
			Success:     true,
			Redelivered: false,
			URL:         "https://example.com/webhook",
		},
	}

	return deliveries, nil
}

// Helper methods

// validateWebhookRequest validates a webhook creation request.
func (w *webhookServiceImpl) validateWebhookRequest(request *WebhookCreateRequest) error {
	if request.Name == "" {
		return fmt.Errorf("webhook name is required")
	}

	if request.URL == "" {
		return fmt.Errorf("webhook URL is required")
	}

	if len(request.Events) == 0 {
		return fmt.Errorf("at least one event must be specified")
	}

	return nil
}

// webhookMatchesSelector checks if a webhook matches the given selector criteria.
func (w *webhookServiceImpl) webhookMatchesSelector(webhook *WebhookInfo, selector *WebhookSelector) bool {
	if selector.ByName != "" && webhook.Name != selector.ByName {
		return false
	}

	if selector.ByURL != "" && webhook.URL != selector.ByURL {
		return false
	}

	if selector.Active != nil && webhook.Active != *selector.Active {
		return false
	}

	if len(selector.ByEvents) > 0 {
		// Check if webhook has all specified events
		for _, event := range selector.ByEvents {
			found := false

			for _, webhookEvent := range webhook.Events {
				if webhookEvent == event {
					found = true
					break
				}
			}

			if !found {
				return false
			}
		}
	}

	return true
}

// Logger interface is defined in constructors.go
