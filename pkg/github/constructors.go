package github

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// HTTPClient interface for dependency injection.
type HTTPClientInterface interface {
	Do(req *http.Request) (*http.Response, error)
	Get(url string) (*http.Response, error)
	Post(url, contentType string, body interface{}) (*http.Response, error)
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
		BaseURL:    "https://api.github.com",
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
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

// NewAPIClient creates a new GitHub API client with dependencies.
func NewAPIClient(config *APIClientConfig, httpClient HTTPClientInterface, logger Logger) APIClient {
	if config == nil {
		config = DefaultAPIClientConfig()
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
	for _, valid := range validStrategies {
		if strategy == valid {
			return nil
		}
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
