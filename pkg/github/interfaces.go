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
	// Do performs an HTTP request
	Do(req *http.Request) (*http.Response, error)

	// Get performs a GET request
	Get(ctx context.Context, url string) (*http.Response, error)

	// Post performs a POST request
	Post(ctx context.Context, url string, contentType string, body io.Reader) (*http.Response, error)

	// Put performs a PUT request
	Put(ctx context.Context, url string, contentType string, body io.Reader) (*http.Response, error)

	// Patch performs a PATCH request
	Patch(ctx context.Context, url string, contentType string, body io.Reader) (*http.Response, error)

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

// APIClient defines the interface for GitHub API operations.
type APIClient interface {
	// Repository operations
	GetRepository(ctx context.Context, owner, repo string) (*RepositoryInfo, error)
	ListOrganizationRepositories(ctx context.Context, org string) ([]RepositoryInfo, error)
	GetDefaultBranch(ctx context.Context, owner, repo string) (string, error)

	// Authentication and rate limiting
	SetToken(token string)
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
	SetStrategy(strategy string) error
	GetSupportedStrategies() []string
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
	GetRequiredScopes(operation string) []string
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
	SetLogLevel(level LogLevelType)
}

// LogOperationRecord represents a logged operation.
type LogOperationRecord struct {
	ID         string                 `json:"id"`
	Timestamp  time.Time              `json:"timestamp"`
	Operation  string                 `json:"operation"`
	Repository string                 `json:"repository"`
	User       string                 `json:"user"`
	Success    bool                   `json:"success"`
	Error      string                 `json:"error,omitempty"`
	Metadata   map[string]interface{} `json:"metadata"`
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
	SetConfirmationMode(mode ConfirmationModeType)
}

// ConfirmationPromptRecord represents a confirmation request.
type ConfirmationPromptRecord struct {
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Repository  string                 `json:"repository"`
	Operation   string                 `json:"operation"`
	Risk        RiskLevelType          `json:"risk"`
	Impact      string                 `json:"impact"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// OperationRecord represents an operation that requires confirmation.
type OperationRecord struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Repository  string                 `json:"repository"`
	Description string                 `json:"description"`
	Risk        RiskLevelType          `json:"risk"`
	Metadata    map[string]interface{} `json:"metadata"`
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
