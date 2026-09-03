// Package github provides interfaces and types for GitHub API integration.
// It defines contracts for HTTP operations, repository management, token validation,
// change logging, and confirmation services used throughout the application.
package github

import (
	"context"
	"io"
	"net/http"
	"time"
)

// HTTPClient defines the interface for HTTP operations.
type HTTPClient interface {
	// Do performs an HTTP request with context
	Do(ctx context.Context, req *http.Request) (*http.Response, error)

	// Get performs a GET request
	Get(ctx context.Context, url string) (*http.Response, error)

	// Post performs a POST request
	Post(ctx context.Context, url, contentType string, body io.Reader) (*http.Response, error)

	// Put performs a PUT request
	Put(ctx context.Context, url, contentType string, body io.Reader) (*http.Response, error)

	// Patch performs a PATCH request
	Patch(ctx context.Context, url, contentType string, body io.Reader) (*http.Response, error)

	// Delete performs a DELETE request
	Delete(ctx context.Context, url string) (*http.Response, error)
}

// RepositoryInfo represents a GitHub repository with essential information for interfaces.
type RepositoryInfo struct {
	Name          string    `json:"name"`
	FullName      string    `json:"full_name"`
	Description   string    `json:"description"`
	DefaultBranch string    `json:"default_branch"`
	CloneURL      string    `json:"clone_url"`
	SSHURL        string    `json:"ssh_url"`
	HTMLURL       string    `json:"html_url"`
	Private       bool      `json:"private"`
	Archived      bool      `json:"archived"`
	Disabled      bool      `json:"disabled"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Language      string    `json:"language"`
	Size          int       `json:"size"`
	Topics        []string  `json:"topics"`
	Visibility    string    `json:"visibility"`
	IsTemplate    bool      `json:"is_template"`
}

// CreateRepositoryOptions holds parameters for creating a GitHub repository.
type CreateRepositoryOptions struct {
	Name              string `json:"name"`
	Description       string `json:"description,omitempty"`
	Homepage          string `json:"homepage,omitempty"`
	Private           bool   `json:"private"`
	HasIssues         bool   `json:"has_issues"`
	HasProjects       bool   `json:"has_projects"`
	HasWiki           bool   `json:"has_wiki"`
	HasDownloads      bool   `json:"has_downloads"`
	AutoInit          bool   `json:"auto_init"`
	GitignoreTemplate string `json:"gitignore_template,omitempty"`
	LicenseTemplate   string `json:"license_template,omitempty"`
	DefaultBranch     string `json:"default_branch,omitempty"`
	AllowSquashMerge  bool   `json:"allow_squash_merge"`
	AllowMergeCommit  bool   `json:"allow_merge_commit"`
	AllowRebaseMerge  bool   `json:"allow_rebase_merge"`
	AllowAutoMerge    bool   `json:"allow_auto_merge"`
}

// UpdateRepositoryOptions holds parameters for PATCH /repos/{owner}/{repo}.
// Pointer fields omit when nil so partial updates are sent correctly.
type UpdateRepositoryOptions struct {
	Name             *string `json:"name,omitempty"`
	Description      *string `json:"description,omitempty"`
	Homepage         *string `json:"homepage,omitempty"`
	Private          *bool   `json:"private,omitempty"`
	HasIssues        *bool   `json:"has_issues,omitempty"`
	HasProjects      *bool   `json:"has_projects,omitempty"`
	HasWiki          *bool   `json:"has_wiki,omitempty"`
	HasDownloads     *bool   `json:"has_downloads,omitempty"`
	DefaultBranch    *string `json:"default_branch,omitempty"`
	AllowSquashMerge *bool   `json:"allow_squash_merge,omitempty"`
	AllowMergeCommit *bool   `json:"allow_merge_commit,omitempty"`
	AllowRebaseMerge *bool   `json:"allow_rebase_merge,omitempty"`
	AllowAutoMerge   *bool   `json:"allow_auto_merge,omitempty"`
	Archived         *bool   `json:"archived,omitempty"`
}

// ForkRepositoryOptions holds parameters for POST /repos/{owner}/{repo}/forks.
type ForkRepositoryOptions struct {
	Organization      string `json:"organization,omitempty"`
	Name              string `json:"name,omitempty"`
	DefaultBranchOnly bool   `json:"default_branch_only,omitempty"`
}

// SearchRepositoriesOptions holds parameters for searching repositories.
type SearchRepositoriesOptions struct {
	Sort    string
	Order   string
	Page    int
	PerPage int
}

// RepositorySearchResult holds GitHub repository search results.
type RepositorySearchResult struct {
	TotalCount        int              `json:"total_count"`
	IncompleteResults bool             `json:"incomplete_results"`
	Repositories      []RepositoryInfo `json:"items"`
}

// APIClient defines the interface for GitHub API operations.
type APIClient interface {
	// Repository operations
	GetRepository(ctx context.Context, owner, repo string) (*RepositoryInfo, error)
	ListOrganizationRepositories(ctx context.Context, org string) ([]RepositoryInfo, error)
	GetDefaultBranch(ctx context.Context, owner, repo string) (string, error)

	// Repository mutations (CLI surface: create/delete/archive/search + update/fork)
	CreateRepository(ctx context.Context, owner string, opts *CreateRepositoryOptions) (*RepositoryInfo, error)
	UpdateRepository(ctx context.Context, owner, repo string, opts *UpdateRepositoryOptions) (*RepositoryInfo, error)
	DeleteRepository(ctx context.Context, owner, repo string) error
	ArchiveRepository(ctx context.Context, owner, repo string) error
	UnarchiveRepository(ctx context.Context, owner, repo string) error
	ForkRepository(ctx context.Context, owner, repo string, opts *ForkRepositoryOptions) (*RepositoryInfo, error)
	SearchRepositories(ctx context.Context, query string, opts *SearchRepositoriesOptions) (*RepositorySearchResult, error)

	// Authentication and rate limiting
	SetToken(ctx context.Context, token string) error
	GetRateLimit(ctx context.Context) (*RateLimit, error)

	// Repository configuration
	GetRepositoryConfiguration(ctx context.Context, owner, repo string) (*RepositoryConfig, error)
	UpdateRepositoryConfiguration(ctx context.Context, owner, repo string, config *RepositoryConfig) error
}

// CloneService defines the interface for repository cloning operations.
type CloneService interface {
	// Clone a single repository
	CloneRepository(ctx context.Context, repo RepositoryInfo, targetPath, strategy string) error

	// Bulk operations
	RefreshAll(ctx context.Context, targetPath, orgName, strategy string) error
	CloneOrganization(ctx context.Context, orgName, targetPath, strategy string) error

	// Strategy management
	SetStrategy(ctx context.Context, strategy string) error
	GetSupportedStrategies(ctx context.Context) ([]string, error)
}

// RateLimit represents GitHub API rate limit information.
type RateLimit struct {
	Limit     int       `json:"limit"`
	Remaining int       `json:"remaining"`
	Reset     time.Time `json:"reset"`
	Used      int       `json:"used"`
}

// TokenValidatorInterface defines the interface for GitHub token validation.
type TokenValidatorInterface interface {
	ValidateToken(ctx context.Context, token string) (*TokenInfoRecord, error)
	ValidateForOperation(ctx context.Context, token, operation string) error
	ValidateForRepository(ctx context.Context, token, owner, repo string) error
	GetRequiredScopes(ctx context.Context, operation string) ([]string, error)
}

// TokenInfoRecord represents information about a GitHub token.
type TokenInfoRecord struct {
	Valid       bool      `json:"valid"`
	Scopes      []string  `json:"scopes"`
	RateLimit   RateLimit `json:"rate_limit"`
	User        string    `json:"user"`
	ExpiresAt   time.Time `json:"expires_at,omitempty"`
	Permissions []string  `json:"permissions"`
}

// ChangeLoggerInterface defines the interface for logging repository changes.
type ChangeLoggerInterface interface {
	LogOperation(ctx context.Context, operation LogOperationRecord) error
	GetOperationHistory(ctx context.Context, filters LogFilters) ([]LogOperationRecord, error)
	SetLogLevel(ctx context.Context, level LogLevelType) error
}

// LogOperationRecord represents a logged operation.
type LogOperationRecord struct {
	ID         string         `json:"id"`
	Timestamp  time.Time      `json:"timestamp"`
	Operation  string         `json:"operation"`
	Repository string         `json:"repository"`
	User       string         `json:"user"`
	Success    bool           `json:"success"`
	Error      string         `json:"error,omitempty"`
	Metadata   map[string]any `json:"metadata"`
}

// LogLevelType represents the logging level.
type LogLevelType int

const (
	LogLevelTypeDebug LogLevelType = iota
	LogLevelTypeInfo
	LogLevelTypeWarn
	LogLevelTypeError
)

// LogFilters defines filters for operation history queries.
type LogFilters struct {
	Repository string    `json:"repository,omitempty"`
	Operation  string    `json:"operation,omitempty"`
	User       string    `json:"user,omitempty"`
	StartTime  time.Time `json:"start_time,omitempty"`
	EndTime    time.Time `json:"end_time,omitempty"`
	Success    *bool     `json:"success,omitempty"`
}

// ConfirmationServiceInterface defines the interface for user confirmation operations.
type ConfirmationServiceInterface interface {
	ConfirmOperation(ctx context.Context, prompt *ConfirmationPromptRecord) (bool, error)
	ConfirmBulkOperation(ctx context.Context, operations []OperationRecord) ([]bool, error)
	SetConfirmationMode(ctx context.Context, mode ConfirmationModeType) error
}

// ConfirmationPromptRecord represents a confirmation request.
type ConfirmationPromptRecord struct {
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Repository  string         `json:"repository"`
	Operation   string         `json:"operation"`
	Risk        RiskLevelType  `json:"risk"`
	Impact      string         `json:"impact"`
	Metadata    map[string]any `json:"metadata"`
}

// OperationRecord represents an operation that requires confirmation.
type OperationRecord struct {
	ID          string         `json:"id"`
	Type        string         `json:"type"`
	Repository  string         `json:"repository"`
	Description string         `json:"description"`
	Risk        RiskLevelType  `json:"risk"`
	Metadata    map[string]any `json:"metadata"`
}

// RiskLevelType represents the risk level of an operation.
type RiskLevelType int

const (
	RiskLevelLow RiskLevelType = iota
	RiskLevelMedium
	RiskLevelHigh
	RiskLevelCritical
)

// ConfirmationModeType represents the confirmation mode.
type ConfirmationModeType int

const (
	ConfirmationModeInteractive ConfirmationModeType = iota
	ConfirmationModeAutoApprove
	ConfirmationModeAutoDeny
	ConfirmationModeDryRun
)

// GitHubService provides a unified interface for all GitHub operations.
type GitHubService interface {
	APIClient
	CloneService
	TokenValidatorInterface
	ChangeLoggerInterface
	ConfirmationServiceInterface
}
