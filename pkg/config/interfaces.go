package config

import (
	"context"
	"io"
)

// ConfigLoader defines the interface for configuration loading operations
type ConfigLoader interface {
	// Load configuration from default search paths
	LoadConfig(ctx context.Context) (*Config, error)

	// Load configuration from a specific file
	LoadConfigFromFile(ctx context.Context, filename string) (*Config, error)

	// Load configuration from a reader (for testing)
	LoadConfigFromReader(ctx context.Context, reader io.Reader) (*Config, error)

	// Get search paths for configuration files
	GetSearchPaths() []string

	// Set custom search paths
	SetSearchPaths(paths []string)
}

// ConfigValidator defines the interface for configuration validation
type ConfigValidator interface {
	// Validate a configuration object
	ValidateConfig(ctx context.Context, config *Config) error

	// Validate a configuration file
	ValidateConfigFile(ctx context.Context, filename string) error

	// Get validation errors with detailed messages
	GetValidationErrors(ctx context.Context, config *Config) []ValidationError

	// Check if configuration is valid
	IsValid(ctx context.Context, config *Config) bool
}

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field        string `json:"field"`
	Value        string `json:"value"`
	Message      string `json:"message"`
	Severity     string `json:"severity"` // error, warning, info
	Suggestion   string `json:"suggestion,omitempty"`
	LineNumber   int    `json:"line_number,omitempty"`
	ColumnNumber int    `json:"column_number,omitempty"`
}

// ConfigParser defines the interface for parsing configuration files
type ConfigParser interface {
	// Parse configuration from bytes
	ParseConfig(ctx context.Context, data []byte) (*Config, error)

	// Parse configuration with specific format
	ParseConfigWithFormat(ctx context.Context, data []byte, format string) (*Config, error)

	// Get supported formats
	GetSupportedFormats() []string

	// Validate format
	IsFormatSupported(format string) bool
}

// SchemaValidator defines the interface for schema validation
type SchemaValidator interface {
	// Validate against JSON schema
	ValidateSchema(ctx context.Context, data []byte, schemaPath string) error

	// Get schema for configuration
	GetConfigSchema() ([]byte, error)

	// Validate configuration structure
	ValidateStructure(ctx context.Context, config *Config) error
}

// ProviderManager defines the interface for managing provider configurations
type ProviderManager interface {
	// Get all configured providers
	GetProviders(ctx context.Context) (map[string]Provider, error)

	// Get specific provider configuration
	GetProvider(ctx context.Context, name string) (*Provider, error)

	// Create provider cloner
	CreateProviderCloner(ctx context.Context, providerName, token string) (ProviderCloner, error)

	// Validate provider configuration
	ValidateProvider(ctx context.Context, provider *Provider) error

	// Get supported providers
	GetSupportedProviders() []string
}

// DirectoryResolver defines the interface for resolving target directories
type DirectoryResolver interface {
	// Resolve directory paths for repositories
	ResolveDirectories(ctx context.Context, config *Config) ([]RepositoryPath, error)

	// Resolve single directory
	ResolveDirectory(ctx context.Context, pattern string, metadata map[string]string) (string, error)

	// Expand environment variables in paths
	ExpandPath(path string) string

	// Check if path exists and is accessible
	ValidatePath(ctx context.Context, path string) error
}

// RepositoryPath represents a resolved repository path
type RepositoryPath struct {
	Original  string            `json:"original"`
	Resolved  string            `json:"resolved"`
	Variables map[string]string `json:"variables"`
	Valid     bool              `json:"valid"`
	Error     string            `json:"error,omitempty"`
}

// FilterService defines the interface for repository filtering
type FilterService interface {
	// Apply filters to repository list
	ApplyFilters(ctx context.Context, repositories []Repository, filters *RepositoryFilter) ([]Repository, error)

	// Check if repository matches filter
	MatchesFilter(ctx context.Context, repository Repository, filter *RepositoryFilter) (bool, error)

	// Validate filter configuration
	ValidateFilter(ctx context.Context, filter *RepositoryFilter) error

	// Get filter statistics
	GetFilterStats(ctx context.Context, repositories []Repository, filter *RepositoryFilter) (*FilterStats, error)
}

// Repository represents a repository for filtering purposes
type Repository struct {
	Name       string            `json:"name"`
	FullName   string            `json:"full_name"`
	Provider   string            `json:"provider"`
	Visibility string            `json:"visibility"`
	Language   string            `json:"language"`
	Topics     []string          `json:"topics"`
	Archived   bool              `json:"archived"`
	Disabled   bool              `json:"disabled"`
	Size       int               `json:"size"`
	Metadata   map[string]string `json:"metadata"`
}

// FilterStats represents statistics about filter application
type FilterStats struct {
	TotalRepositories    int `json:"total_repositories"`
	MatchingRepositories int `json:"matching_repositories"`
	FilteredRepositories int `json:"filtered_repositories"`
	IncludedByPattern    int `json:"included_by_pattern"`
	ExcludedByPattern    int `json:"excluded_by_pattern"`
	ExcludedByVisibility int `json:"excluded_by_visibility"`
	ExcludedByArchived   int `json:"excluded_by_archived"`
}

// IntegrationService defines the interface for bulk clone integration
type IntegrationService interface {
	// Get all clone targets
	GetAllTargets(ctx context.Context) ([]BulkCloneTarget, error)

	// Get targets by provider
	GetTargetsByProvider(ctx context.Context, provider string) ([]BulkCloneTarget, error)

	// Check if target should be processed
	ShouldProcessTarget(ctx context.Context, target BulkCloneTarget, filters map[string]interface{}) bool

	// Execute bulk clone operation
	ExecuteBulkClone(ctx context.Context, targets []BulkCloneTarget) (*BulkCloneResult, error)
}

// BulkCloneTarget represents a target for bulk cloning
type BulkCloneTarget struct {
	Provider string            `json:"provider"`
	Name     string            `json:"name"`
	CloneDir string            `json:"clone_dir"`
	Strategy string            `json:"strategy"`
	Filters  *RepositoryFilter `json:"filters,omitempty"`
	Metadata map[string]string `json:"metadata"`
}

// ConfigService provides a unified interface for all configuration operations
type ConfigService interface {
	ConfigLoader
	ConfigValidator
	ConfigParser
	SchemaValidator
	ProviderManager
	DirectoryResolver
	FilterService
	IntegrationService
}

// ConfigWatcher defines the interface for watching configuration file changes
type ConfigWatcher interface {
	// Start watching configuration files
	StartWatching(ctx context.Context, paths []string) error

	// Stop watching
	StopWatching() error

	// Get notification channel for configuration changes
	Changes() <-chan ConfigChangeEvent

	// Set callback for configuration changes
	OnChange(callback func(event ConfigChangeEvent))
}

// ConfigChangeEvent represents a configuration file change event
type ConfigChangeEvent struct {
	Path      string  `json:"path"`
	Operation string  `json:"operation"` // create, write, remove, rename
	Config    *Config `json:"config,omitempty"`
	Error     string  `json:"error,omitempty"`
}
