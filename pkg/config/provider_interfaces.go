package config

import (
	"context"
)

// ProviderService defines the unified interface for all Git service providers
// This interface abstracts away the specific implementation details of GitHub, GitLab, Gitea, etc.
type ProviderService interface {
	// Repository Operations
	ListRepositories(ctx context.Context, owner string) ([]Repository, error)
	CloneRepository(ctx context.Context, owner, repository, targetPath string) error
	GetDefaultBranch(ctx context.Context, owner, repository string) (string, error)

	// Bulk Operations
	RefreshAll(ctx context.Context, targetPath, owner, strategy string) error
	CloneOrganization(ctx context.Context, owner, targetPath, strategy string) error

	// Authentication and Configuration
	SetToken(token string)
	ValidateToken(ctx context.Context) error

	// Provider Information
	GetProviderName() string
	GetAPIEndpoint() string

	// Health and Status
	IsHealthy(ctx context.Context) error
}

// Repository represents a repository across different providers.
type Repository struct {
	Name          string `json:"name"`
	FullName      string `json:"full_name"`
	Description   string `json:"description"`
	DefaultBranch string `json:"default_branch"`
	CloneURL      string `json:"clone_url"`
	SSHURL        string `json:"ssh_url"`
	HTMLURL       string `json:"html_url"`
	Private       bool   `json:"private"`
	Archived      bool   `json:"archived"`
	Language      string `json:"language"`
	Size          int64  `json:"size"`
}

// ProviderFactory is defined in factory.go

// ConfigurationService handles configuration loading and validation.
type ConfigurationService interface {
	// Configuration Loading
	LoadConfiguration(ctx context.Context) (*Config, error)
	LoadConfigurationFromFile(ctx context.Context, filename string) (*Config, error)

	// Configuration Validation
	ValidateConfiguration(ctx context.Context, config *Config) error
	ValidateProviderConfiguration(ctx context.Context, providerName string, config ProviderConfig) error

	// Configuration Discovery
	FindConfigurationFiles() ([]string, error)
	GetDefaultConfigPath() string

	// Configuration Management
	SaveConfiguration(ctx context.Context, config *Config, filename string) error
	MergeConfigurations(base *Config, overlay *Config) (*Config, error)
}

// BulkOperationService handles bulk operations across multiple repositories/providers.
type BulkOperationService interface {
	// Bulk Clone Operations
	CloneAll(ctx context.Context, request *BulkCloneRequest) (*BulkCloneResult, error)
	CloneByProvider(ctx context.Context, providerName string, request *BulkCloneRequest) (*BulkCloneResult, error)
	CloneByFilter(ctx context.Context, filter RepositoryFilter, request *BulkCloneRequest) (*BulkCloneResult, error)

	// Bulk Update Operations
	RefreshAll(ctx context.Context, request *BulkRefreshRequest) (*BulkRefreshResult, error)
	RefreshByProvider(ctx context.Context, providerName string, request *BulkRefreshRequest) (*BulkRefreshResult, error)

	// Status and Discovery
	GetRepositoryStatus(ctx context.Context, targetPath string) (*RepositoryStatus, error)
	DiscoverRepositories(ctx context.Context, providers []string) (*DiscoveryResult, error)
}

// BulkCloneRequest represents a request for bulk cloning operations.
type BulkCloneRequest struct {
	Providers     []string          `json:"providers,omitempty"`
	Organizations []string          `json:"organizations,omitempty"`
	Repositories  []string          `json:"repositories,omitempty"`
	TargetPath    string            `json:"target_path"`
	Strategy      string            `json:"strategy"`
	Filters       *RepositoryFilter `json:"filters,omitempty"`
	Concurrency   int               `json:"concurrency"`
	DryRun        bool              `json:"dry_run"`
	Credentials   map[string]string `json:"credentials,omitempty"`
}

// BulkRefreshRequest represents a request for bulk refresh operations.
type BulkRefreshRequest struct {
	TargetPath    string            `json:"target_path"`
	Strategy      string            `json:"strategy"`
	Organizations []string          `json:"organizations,omitempty"`
	Filters       *RepositoryFilter `json:"filters,omitempty"`
	Concurrency   int               `json:"concurrency"`
	DryRun        bool              `json:"dry_run"`
}

// BulkCloneResult is defined in providers.go

// BulkRefreshResult represents the result of bulk refresh operations.
type BulkRefreshResult struct {
	TotalRepositories int                   `json:"total_repositories"`
	RefreshSuccessful int                   `json:"refresh_successful"`
	RefreshFailed     int                   `json:"refresh_failed"`
	RefreshSkipped    int                   `json:"refresh_skipped"`
	OperationResults  []RepositoryOperation `json:"operation_results"`
	ExecutionTimeMs   int64                 `json:"execution_time_ms"`
	ErrorSummary      map[string]int        `json:"error_summary"`
}

// RepositoryOperation represents the result of an operation on a single repository.
type RepositoryOperation struct {
	Repository   string `json:"repository"`
	Organization string `json:"organization"`
	Provider     string `json:"provider"`
	Operation    string `json:"operation"`
	Success      bool   `json:"success"`
	Error        string `json:"error,omitempty"`
	DurationMs   int64  `json:"duration_ms"`
	Path         string `json:"path,omitempty"`
}

// DiscoveryResult represents the result of repository discovery operations.
type DiscoveryResult struct {
	TotalRepositories      int            `json:"total_repositories"`
	RepositoriesByProvider map[string]int `json:"repositories_by_provider"`
	Repositories           []Repository   `json:"repositories"`
	ExecutionTimeMs        int64          `json:"execution_time_ms"`
}

// RepositoryStatus represents the status of repositories in a directory.
type RepositoryStatus struct {
	TotalRepositories    int                    `json:"total_repositories"`
	HealthyRepositories  int                    `json:"healthy_repositories"`
	BrokenRepositories   int                    `json:"broken_repositories"`
	OutdatedRepositories int                    `json:"outdated_repositories"`
	RepositoryDetails    []RepositoryStatusInfo `json:"repository_details"`
	ScanTimeMs           int64                  `json:"scan_time_ms"`
}

// RepositoryStatusInfo contains detailed status information for a single repository.
type RepositoryStatusInfo struct {
	Path           string   `json:"path"`
	Name           string   `json:"name"`
	Organization   string   `json:"organization"`
	Provider       string   `json:"provider"`
	IsHealthy      bool     `json:"is_healthy"`
	Issues         []string `json:"issues,omitempty"`
	CurrentBranch  string   `json:"current_branch"`
	DefaultBranch  string   `json:"default_branch"`
	RemoteURL      string   `json:"remote_url"`
	LastCommitHash string   `json:"last_commit_hash,omitempty"`
}
