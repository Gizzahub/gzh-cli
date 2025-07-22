// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

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
	FullName      string `json:"fullName"`
	Description   string `json:"description"`
	DefaultBranch string `json:"defaultBranch"`
	CloneURL      string `json:"cloneUrl"`
	SSHURL        string `json:"sshUrl"`
	HTMLURL       string `json:"htmlUrl"`
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
	TargetPath    string            `json:"targetPath"`
	Strategy      string            `json:"strategy"`
	Filters       *RepositoryFilter `json:"filters,omitempty"`
	Concurrency   int               `json:"concurrency"`
	DryRun        bool              `json:"dryRun"`
	Credentials   map[string]string `json:"credentials,omitempty"`
}

// BulkRefreshRequest represents a request for bulk refresh operations.
type BulkRefreshRequest struct {
	TargetPath    string            `json:"targetPath"`
	Strategy      string            `json:"strategy"`
	Organizations []string          `json:"organizations,omitempty"`
	Filters       *RepositoryFilter `json:"filters,omitempty"`
	Concurrency   int               `json:"concurrency"`
	DryRun        bool              `json:"dryRun"`
}

// BulkCloneResult is defined in providers.go

// BulkRefreshResult represents the result of bulk refresh operations.
type BulkRefreshResult struct {
	TotalRepositories int                   `json:"totalRepositories"`
	RefreshSuccessful int                   `json:"refreshSuccessful"`
	RefreshFailed     int                   `json:"refreshFailed"`
	RefreshSkipped    int                   `json:"refreshSkipped"`
	OperationResults  []RepositoryOperation `json:"operationResults"`
	ExecutionTimeMs   int64                 `json:"executionTimeMs"`
	ErrorSummary      map[string]int        `json:"errorSummary"`
}

// RepositoryOperation represents the result of an operation on a single repository.
type RepositoryOperation struct {
	Repository   string `json:"repository"`
	Organization string `json:"organization"`
	Provider     string `json:"provider"`
	Operation    string `json:"operation"`
	Success      bool   `json:"success"`
	Error        string `json:"error,omitempty"`
	DurationMs   int64  `json:"durationMs"`
	Path         string `json:"path,omitempty"`
}

// DiscoveryResult represents the result of repository discovery operations.
type DiscoveryResult struct {
	TotalRepositories      int            `json:"totalRepositories"`
	RepositoriesByProvider map[string]int `json:"repositoriesByProvider"`
	Repositories           []Repository   `json:"repositories"`
	ExecutionTimeMs        int64          `json:"executionTimeMs"`
}

// RepositoryStatus represents the status of repositories in a directory.
type RepositoryStatus struct {
	TotalRepositories    int                    `json:"totalRepositories"`
	HealthyRepositories  int                    `json:"healthyRepositories"`
	BrokenRepositories   int                    `json:"brokenRepositories"`
	OutdatedRepositories int                    `json:"outdatedRepositories"`
	RepositoryDetails    []RepositoryStatusInfo `json:"repositoryDetails"`
	ScanTimeMs           int64                  `json:"scanTimeMs"`
}

// RepositoryStatusInfo contains detailed status information for a single repository.
type RepositoryStatusInfo struct {
	Path           string   `json:"path"`
	Name           string   `json:"name"`
	Organization   string   `json:"organization"`
	Provider       string   `json:"provider"`
	IsHealthy      bool     `json:"isHealthy"`
	Issues         []string `json:"issues,omitempty"`
	CurrentBranch  string   `json:"currentBranch"`
	DefaultBranch  string   `json:"defaultBranch"`
	RemoteURL      string   `json:"remoteUrl"`
	LastCommitHash string   `json:"lastCommitHash,omitempty"`
}
