package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"
)

// HTTPClient interface for dependency injection.
type HTTPClientInterface interface {
	Do(req *http.Request) (*http.Response, error)
	Get(url string) (*http.Response, error)
	Post(url, contentType string, body any) (*http.Response, error)
}

// FileSystem interface for dependency injection.
type FileSystemInterface interface {
	WriteFile(filename string, data []byte, perm int) error
	ReadFile(filename string) ([]byte, error)
	MkdirAll(path string, perm int) error
	Exists(path string) bool
}

// GitCommand interface for dependency injection.
type GitCommandInterface interface {
	Clone(ctx context.Context, url, path string) error
	Pull(ctx context.Context, path string) error
	Fetch(ctx context.Context, path string) error
	Reset(ctx context.Context, path string, hard bool) error
}

// APIClientConfig holds configuration for GitHub API client.
type APIClientConfig struct {
	BaseURL    string
	Token      string
	Timeout    time.Duration
	UserAgent  string
	RetryCount int
}

// DefaultAPIClientConfig returns default configuration.
func DefaultAPIClientConfig() *APIClientConfig {
	return &APIClientConfig{
		BaseURL:    DefaultGitHubAPIBaseURL,
		Timeout:    30 * time.Second,
		UserAgent:  "gzh-cli/1.0",
		RetryCount: 3,
	}
}

// GitHubAPIClient implements the APIClient interface.
type GitHubAPIClient struct {
	config     *APIClientConfig
	httpClient HTTPClientInterface
	logger     Logger
}

// Logger interface for dependency injection.
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

// NewAPIClient creates a new GitHub API client with dependencies.
// Empty config.BaseURL falls back to DefaultGitHubAPIBaseURL.
func NewAPIClient(config *APIClientConfig, httpClient HTTPClientInterface, logger Logger) APIClient {
	if config == nil {
		config = DefaultAPIClientConfig()
	} else {
		// Resolve without mutating the caller's config pointer unexpectedly for other fields;
		// BaseURL empty→default is intentional so GHES/tests can omit it.
		resolved := *config
		resolved.BaseURL = ResolveGitHubAPIBaseURL(config.BaseURL)
		config = &resolved
	}

	return &GitHubAPIClient{
		config:     config,
		httpClient: httpClient,
		logger:     logger,
	}
}

// SetToken implements APIClient interface.
func (c *GitHubAPIClient) SetToken(ctx context.Context, token string) error {
	c.config.Token = token
	return nil
}

// GetRepository implements APIClient interface.
func (c *GitHubAPIClient) GetRepository(ctx context.Context, owner, repo string) (*RepositoryInfo, error) {
	c.logger.Debug("Getting repository info", "owner", owner, "repo", repo)

	// Implementation would use c.httpClient instead of direct http calls
	// This is just a placeholder showing the pattern
	return nil, fmt.Errorf("GetRepository not implemented")
}

// ListOrganizationRepositories implements APIClient interface.
func (c *GitHubAPIClient) ListOrganizationRepositories(ctx context.Context, org string) ([]RepositoryInfo, error) {
	c.logger.Debug("Listing organization repositories", "org", org)

	// Implementation would use c.httpClient instead of direct http calls
	return nil, fmt.Errorf("ListOrganizationRepositories not implemented")
}

// GetDefaultBranch implements APIClient interface.
func (c *GitHubAPIClient) GetDefaultBranch(ctx context.Context, owner, repo string) (string, error) {
	c.logger.Debug("Getting default branch", "owner", owner, "repo", repo)

	// Implementation would use c.httpClient instead of direct http calls
	return "main", nil
}

// GetRateLimit implements APIClient interface.
func (c *GitHubAPIClient) GetRateLimit(ctx context.Context) (*RateLimit, error) {
	c.logger.Debug("Getting rate limit info")

	// Implementation would use c.httpClient instead of direct http calls
	return &RateLimit{
		Limit:     5000,
		Remaining: 4999,
		Reset:     time.Now().Add(time.Hour),
		Used:      1,
	}, nil
}

// GetRepositoryConfiguration implements APIClient interface.
func (c *GitHubAPIClient) GetRepositoryConfiguration(ctx context.Context, owner, repo string) (*RepositoryConfig, error) {
	c.logger.Debug("Getting repository configuration", "owner", owner, "repo", repo)

	// Implementation would use c.httpClient
	return nil, fmt.Errorf("GetRepositoryConfiguration not implemented")
}

// UpdateRepositoryConfiguration implements APIClient interface.
func (c *GitHubAPIClient) UpdateRepositoryConfiguration(ctx context.Context, owner, repo string, config *RepositoryConfig) error {
	c.logger.Debug("Updating repository configuration", "owner", owner, "repo", repo)

	// Implementation would use c.httpClient
	return nil
}

// CreateRepository implements APIClient interface.
func (c *GitHubAPIClient) CreateRepository(ctx context.Context, owner string, opts *CreateRepositoryOptions) (*RepositoryInfo, error) {
	if owner == "" {
		return nil, fmt.Errorf("owner is required")
	}
	if opts == nil || opts.Name == "" {
		return nil, fmt.Errorf("repository name is required")
	}
	c.logger.Debug("Creating repository", "owner", owner, "name", opts.Name)

	// Prefer org endpoint; fall back to authenticated user endpoint.
	info, err := c.doJSONRequest(ctx, http.MethodPost, fmt.Sprintf("/orgs/%s/repos", owner), opts, http.StatusCreated)
	if err == nil {
		var repo RepositoryInfo
		if decErr := json.Unmarshal(info, &repo); decErr != nil {
			return nil, fmt.Errorf("failed to decode create repository response: %w", decErr)
		}
		return &repo, nil
	}
	if !strings.Contains(err.Error(), "404") {
		return nil, err
	}

	body, err := c.doJSONRequest(ctx, http.MethodPost, "/user/repos", opts, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	var repo RepositoryInfo
	if err := json.Unmarshal(body, &repo); err != nil {
		return nil, fmt.Errorf("failed to decode create repository response: %w", err)
	}
	return &repo, nil
}

// DeleteRepository implements APIClient interface.
func (c *GitHubAPIClient) DeleteRepository(ctx context.Context, owner, repo string) error {
	if owner == "" || repo == "" {
		return fmt.Errorf("owner and repo are required")
	}
	c.logger.Debug("Deleting repository", "owner", owner, "repo", repo)
	_, err := c.doJSONRequest(ctx, http.MethodDelete, fmt.Sprintf("/repos/%s/%s", owner, repo), nil, http.StatusNoContent)
	return err
}

// ArchiveRepository implements APIClient interface.
func (c *GitHubAPIClient) ArchiveRepository(ctx context.Context, owner, repo string) error {
	return c.setRepositoryArchived(ctx, owner, repo, true)
}

// UnarchiveRepository implements APIClient interface.
func (c *GitHubAPIClient) UnarchiveRepository(ctx context.Context, owner, repo string) error {
	return c.setRepositoryArchived(ctx, owner, repo, false)
}

func (c *GitHubAPIClient) setRepositoryArchived(ctx context.Context, owner, repo string, archived bool) error {
	if owner == "" || repo == "" {
		return fmt.Errorf("owner and repo are required")
	}
	c.logger.Debug("Setting repository archived state", "owner", owner, "repo", repo, "archived", archived)
	_, err := c.doJSONRequest(ctx, http.MethodPatch, fmt.Sprintf("/repos/%s/%s", owner, repo), map[string]bool{"archived": archived}, http.StatusOK)
	return err
}

// SearchRepositories implements APIClient interface.
func (c *GitHubAPIClient) SearchRepositories(ctx context.Context, query string, opts *SearchRepositoriesOptions) (*RepositorySearchResult, error) {
	if query == "" {
		return nil, fmt.Errorf("search query is required")
	}
	if opts == nil {
		opts = &SearchRepositoriesOptions{}
	}
	c.logger.Debug("Searching repositories", "query", query)

	page := opts.Page
	if page <= 0 {
		page = 1
	}
	perPage := opts.PerPage
	if perPage <= 0 {
		perPage = 30
	}

	params := url.Values{}
	params.Set("q", query)
	params.Set("page", strconv.Itoa(page))
	params.Set("per_page", strconv.Itoa(perPage))
	if opts.Sort != "" {
		params.Set("sort", opts.Sort)
	}
	if opts.Order != "" {
		params.Set("order", opts.Order)
	}

	body, err := c.doJSONRequest(ctx, http.MethodGet, "/search/repositories?"+params.Encode(), nil, http.StatusOK)
	if err != nil {
		return nil, err
	}
	var result RepositorySearchResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}
	return &result, nil
}

// doJSONRequest performs an authenticated JSON request against the configured base URL.
// acceptedStatus is the primary success code; StatusOK is also accepted when primary is NoContent/Created.
func (c *GitHubAPIClient) doJSONRequest(ctx context.Context, method, path string, payload any, acceptedStatus int) ([]byte, error) {
	var bodyReader io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to encode request body: %w", err)
		}
		bodyReader = bytes.NewReader(encoded)
	}

	endpoint := strings.TrimRight(c.config.BaseURL, "/") + "/" + strings.TrimPrefix(path, "/")
	req, err := http.NewRequestWithContext(ctx, method, endpoint, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	if c.config.Token != "" {
		req.Header.Set("Authorization", "token "+c.config.Token)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if c.config.UserAgent != "" {
		req.Header.Set("User-Agent", c.config.UserAgent)
	} else {
		req.Header.Set("User-Agent", "gzh-cli")
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	var resp *http.Response
	if c.httpClient != nil {
		resp, err = c.httpClient.Do(req)
	} else {
		resp, err = http.DefaultClient.Do(req)
	}
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != acceptedStatus && resp.StatusCode != http.StatusOK {
		// Surface status so callers can branch (e.g. org 404 → user create).
		return nil, fmt.Errorf("API %s %s: HTTP %d - %s", method, path, resp.StatusCode, resp.Status)
	}
	return respBody, nil
}

// GitHubCloneService implements the CloneService interface.
type GitHubCloneService struct {
	apiClient  APIClient
	gitClient  GitCommandInterface
	fileSystem FileSystemInterface
	logger     Logger
}

// CloneServiceConfig holds configuration for clone service.
type CloneServiceConfig struct {
	DefaultStrategy string
	Concurrency     int
	Timeout         time.Duration
}

// DefaultCloneServiceConfig returns default clone service configuration.
func DefaultCloneServiceConfig() *CloneServiceConfig {
	return &CloneServiceConfig{
		DefaultStrategy: "reset",
		Concurrency:     5,
		Timeout:         10 * time.Minute,
	}
}

// NewCloneService creates a new clone service with dependencies.
func NewCloneService(
	apiClient APIClient,
	gitClient GitCommandInterface,
	fileSystem FileSystemInterface,
	logger Logger,
) CloneService {
	return &GitHubCloneService{
		apiClient:  apiClient,
		gitClient:  gitClient,
		fileSystem: fileSystem,
		logger:     logger,
	}
}

// CloneRepository implements CloneService interface.
func (s *GitHubCloneService) CloneRepository(ctx context.Context, repo RepositoryInfo, targetPath, strategy string) error {
	s.logger.Info("Cloning repository", "repo", repo.Name, "path", targetPath, "strategy", strategy)

	// Implementation would use s.gitClient instead of direct exec.Command
	return s.gitClient.Clone(ctx, repo.CloneURL, targetPath)
}

// RefreshAll implements CloneService interface.
func (s *GitHubCloneService) RefreshAll(ctx context.Context, targetPath, orgName, strategy string) error {
	s.logger.Info("Refreshing all repositories", "org", orgName, "path", targetPath, "strategy", strategy)

	// Implementation would use s.apiClient to get repos and s.gitClient for operations
	repos, err := s.apiClient.ListOrganizationRepositories(ctx, orgName)
	if err != nil {
		return err
	}

	for _, repo := range repos {
		if err := s.CloneRepository(ctx, repo, targetPath, strategy); err != nil {
			s.logger.Error("Failed to clone repository", "repo", repo.Name, "error", err)
		}
	}

	return nil
}

// CloneOrganization implements CloneService interface.
func (s *GitHubCloneService) CloneOrganization(ctx context.Context, orgName, targetPath, strategy string) error {
	return s.RefreshAll(ctx, targetPath, orgName, strategy)
}

// SetStrategy implements CloneService interface.
func (s *GitHubCloneService) SetStrategy(ctx context.Context, strategy string) error {
	// Validate strategy
	validStrategies, err := s.GetSupportedStrategies(ctx)
	if err != nil {
		return fmt.Errorf("failed to get supported strategies: %w", err)
	}
	if slices.Contains(validStrategies, strategy) {
		return nil
	}

	return fmt.Errorf("unsupported strategy: %s", strategy)
}

// GetSupportedStrategies implements CloneService interface.
func (s *GitHubCloneService) GetSupportedStrategies(ctx context.Context) ([]string, error) {
	return []string{"reset", "pull", "fetch"}, nil
}

// GitHubTokenValidator implements the TokenValidator interface.
type GitHubTokenValidator struct {
	apiClient APIClient
	logger    Logger
}

// NewGitHubTokenValidator creates a new token validator with dependencies.
func NewGitHubTokenValidator(apiClient APIClient, logger Logger) TokenValidatorInterface {
	return &GitHubTokenValidator{
		apiClient: apiClient,
		logger:    logger,
	}
}

// ValidateToken implements TokenValidator interface.
func (v *GitHubTokenValidator) ValidateToken(ctx context.Context, token string) (*TokenInfoRecord, error) {
	v.logger.Debug("Validating GitHub token")

	// Implementation would use v.apiClient
	return &TokenInfoRecord{
		Valid:  true,
		Scopes: []string{"repo", "read:org"},
		User:   "example-user",
	}, nil
}

// ValidateForOperation implements TokenValidator interface.
func (v *GitHubTokenValidator) ValidateForOperation(ctx context.Context, token, operation string) error {
	v.logger.Debug("Validating token for operation", "operation", operation)

	// Implementation logic
	return nil
}

// ValidateForRepository implements TokenValidator interface.
func (v *GitHubTokenValidator) ValidateForRepository(ctx context.Context, token, owner, repo string) error {
	v.logger.Debug("Validating token for repository", "owner", owner, "repo", repo)

	// Implementation logic
	return nil
}

// GetRequiredScopes implements TokenValidator interface.
func (v *GitHubTokenValidator) GetRequiredScopes(ctx context.Context, operation string) ([]string, error) {
	switch operation {
	case "read":
		return []string{"repo"}, nil
	case "write":
		return []string{"repo"}, nil
	case "admin":
		return []string{"repo", "admin:org"}, nil
	default:
		return []string{"repo"}, nil
	}
}

// GitHubServiceContainer holds all GitHub service implementations.
type GitHubServiceContainer struct {
	APIClient      APIClient
	CloneService   CloneService
	TokenValidator TokenValidatorInterface
}

// GitHubServiceConfig holds configuration for the GitHub service.
type GitHubServiceConfig struct {
	API   *APIClientConfig
	Clone *CloneServiceConfig
}

// NewGitHubServiceContainer creates a new GitHub service container with all dependencies.
func NewGitHubServiceContainer(
	config *GitHubServiceConfig,
	httpClient HTTPClientInterface,
	gitClient GitCommandInterface,
	fileSystem FileSystemInterface,
	logger Logger,
) *GitHubServiceContainer {
	if config == nil {
		config = &GitHubServiceConfig{
			API:   DefaultAPIClientConfig(),
			Clone: DefaultCloneServiceConfig(),
		}
	}

	apiClient := NewAPIClient(config.API, httpClient, logger)
	cloneService := NewCloneService(apiClient, gitClient, fileSystem, logger)
	tokenValidator := NewGitHubTokenValidator(apiClient, logger)

	return &GitHubServiceContainer{
		APIClient:      apiClient,
		CloneService:   cloneService,
		TokenValidator: tokenValidator,
	}
}
