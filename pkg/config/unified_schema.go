// Package config provides unified configuration management for gzh-manager
package config

import (
	"time"
)

// UnifiedConfig represents the new unified configuration format
// This merges the functionality of both bulk-clone.yaml and gzh.yaml formats
type UnifiedConfig struct {
	// Schema version for the configuration format
	Version string `yaml:"version" json:"version" validate:"required,oneof=1.0.0"`

	// Default provider to use when not specified
	DefaultProvider string `yaml:"default_provider,omitempty" json:"default_provider,omitempty" validate:"omitempty,oneof=github gitlab gitea gogs"`

	// Global settings that apply to all providers
	Global *GlobalSettings `yaml:"global,omitempty" json:"global,omitempty"`

	// Provider-specific configurations
	Providers map[string]*ProviderConfig `yaml:"providers" json:"providers" validate:"required,min=1"`

	// Migration information from legacy formats
	Migration *MigrationInfo `yaml:"migration,omitempty" json:"migration,omitempty"`
}

// GlobalSettings contains settings that apply across all providers
type GlobalSettings struct {
	// Default clone directory base path
	CloneBaseDir string `yaml:"clone_base_dir,omitempty" json:"clone_base_dir,omitempty"`

	// Default strategy for repository operations
	DefaultStrategy string `yaml:"default_strategy,omitempty" json:"default_strategy,omitempty" validate:"omitempty,oneof=reset pull fetch"`

	// Global ignore patterns (regex)
	GlobalIgnores []string `yaml:"global_ignores,omitempty" json:"global_ignores,omitempty"`

	// Default visibility filter
	DefaultVisibility string `yaml:"default_visibility,omitempty" json:"default_visibility,omitempty" validate:"omitempty,oneof=public private all"`

	// Timeout settings
	Timeouts *TimeoutSettings `yaml:"timeouts,omitempty" json:"timeouts,omitempty"`

	// Concurrency settings
	Concurrency *ConcurrencySettings `yaml:"concurrency,omitempty" json:"concurrency,omitempty"`
}

// TimeoutSettings contains timeout configurations
type TimeoutSettings struct {
	// HTTP request timeout
	HTTPTimeout time.Duration `yaml:"http_timeout,omitempty" json:"http_timeout,omitempty"`

	// Git operation timeout
	GitTimeout time.Duration `yaml:"git_timeout,omitempty" json:"git_timeout,omitempty"`

	// API rate limit timeout
	RateLimitTimeout time.Duration `yaml:"rate_limit_timeout,omitempty" json:"rate_limit_timeout,omitempty"`
}

// ConcurrencySettings contains concurrency configurations
type ConcurrencySettings struct {
	// Maximum concurrent clone operations
	CloneWorkers int `yaml:"clone_workers,omitempty" json:"clone_workers,omitempty" validate:"omitempty,min=1,max=50"`

	// Maximum concurrent update operations
	UpdateWorkers int `yaml:"update_workers,omitempty" json:"update_workers,omitempty" validate:"omitempty,min=1,max=50"`

	// Maximum concurrent API operations
	APIWorkers int `yaml:"api_workers,omitempty" json:"api_workers,omitempty" validate:"omitempty,min=1,max=20"`
}

// ProviderConfig represents configuration for a specific Git provider
type ProviderConfig struct {
	// Authentication token (supports environment variables)
	Token string `yaml:"token,omitempty" json:"token,omitempty" validate:"required,envtoken"`

	// API endpoint URL (for self-hosted instances)
	APIURL string `yaml:"api_url,omitempty" json:"api_url,omitempty" validate:"omitempty,url"`

	// Organizations/groups to manage
	Organizations []*OrganizationConfig `yaml:"organizations,omitempty" json:"organizations,omitempty" validate:"min=1"`

	// Provider-specific settings
	Settings *ProviderSettings `yaml:"settings,omitempty" json:"settings,omitempty"`

	// Legacy support for bulk-clone.yaml format
	Legacy *LegacyProviderConfig `yaml:"legacy,omitempty" json:"legacy,omitempty"`
}

// OrganizationConfig represents configuration for an organization/group
type OrganizationConfig struct {
	// Organization/group name
	Name string `yaml:"name" json:"name" validate:"required"`

	// Clone directory for this organization
	CloneDir string `yaml:"clone_dir" json:"clone_dir" validate:"required,dirpath"`

	// Repository visibility filter
	Visibility string `yaml:"visibility,omitempty" json:"visibility,omitempty" validate:"omitempty,oneof=public private all"`

	// Update strategy for existing repositories
	Strategy string `yaml:"strategy,omitempty" json:"strategy,omitempty" validate:"omitempty,oneof=reset pull fetch"`

	// Include pattern (regex)
	Include string `yaml:"include,omitempty" json:"include,omitempty" validate:"omitempty,regexpattern"`

	// Exclude patterns (regex)
	Exclude []string `yaml:"exclude,omitempty" json:"exclude,omitempty" validate:"dive,regexpattern"`

	// Whether to flatten directory structure
	Flatten bool `yaml:"flatten,omitempty" json:"flatten,omitempty"`

	// Recursive processing (for GitLab groups)
	Recursive bool `yaml:"recursive,omitempty" json:"recursive,omitempty"`

	// Repository management settings
	RepoManagement *RepoManagementConfig `yaml:"repo_management,omitempty" json:"repo_management,omitempty"`

	// Custom labels for organization
	Labels map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
}

// ProviderSettings contains provider-specific settings
type ProviderSettings struct {
	// Rate limiting settings
	RateLimit *RateLimitConfig `yaml:"rate_limit,omitempty" json:"rate_limit,omitempty"`

	// Retry settings
	Retry *RetryConfig `yaml:"retry,omitempty" json:"retry,omitempty"`

	// Authentication settings
	Auth *AuthConfig `yaml:"auth,omitempty" json:"auth,omitempty"`
}

// RateLimitConfig contains rate limiting configuration
type RateLimitConfig struct {
	// Requests per hour
	RequestsPerHour int `yaml:"requests_per_hour,omitempty" json:"requests_per_hour,omitempty"`

	// Burst limit
	BurstLimit int `yaml:"burst_limit,omitempty" json:"burst_limit,omitempty"`

	// Enable automatic rate limit detection
	AutoDetect bool `yaml:"auto_detect,omitempty" json:"auto_detect,omitempty"`
}

// RetryConfig contains retry configuration
type RetryConfig struct {
	// Maximum retry attempts
	MaxAttempts int `yaml:"max_attempts,omitempty" json:"max_attempts,omitempty" validate:"omitempty,min=0,max=10"`

	// Base delay between retries
	BaseDelay time.Duration `yaml:"base_delay,omitempty" json:"base_delay,omitempty"`

	// Maximum delay between retries
	MaxDelay time.Duration `yaml:"max_delay,omitempty" json:"max_delay,omitempty"`

	// Enable exponential backoff
	ExponentialBackoff bool `yaml:"exponential_backoff,omitempty" json:"exponential_backoff,omitempty"`
}

// AuthConfig contains authentication configuration
type AuthConfig struct {
	// Token environment variable name
	TokenEnvVar string `yaml:"token_env_var,omitempty" json:"token_env_var,omitempty"`

	// SSH key path for Git operations
	SSHKeyPath string `yaml:"ssh_key_path,omitempty" json:"ssh_key_path,omitempty"`

	// Use SSH for Git operations
	UseSSH bool `yaml:"use_ssh,omitempty" json:"use_ssh,omitempty"`
}

// RepoManagementConfig contains repository management settings
type RepoManagementConfig struct {
	// Enable repository configuration management
	Enabled bool `yaml:"enabled,omitempty" json:"enabled,omitempty"`

	// Configuration templates to apply
	Templates []string `yaml:"templates,omitempty" json:"templates,omitempty"`

	// Branch protection settings
	BranchProtection *BranchProtectionConfig `yaml:"branch_protection,omitempty" json:"branch_protection,omitempty"`

	// Security settings
	Security *SecurityConfig `yaml:"security,omitempty" json:"security,omitempty"`
}

// BranchProtectionConfig contains branch protection settings
type BranchProtectionConfig struct {
	// Enable branch protection
	Enabled bool `yaml:"enabled,omitempty" json:"enabled,omitempty"`

	// Protected branches (patterns)
	Branches []string `yaml:"branches,omitempty" json:"branches,omitempty"`

	// Require status checks
	RequireStatusChecks bool `yaml:"require_status_checks,omitempty" json:"require_status_checks,omitempty"`

	// Require pull request reviews
	RequirePRReviews bool `yaml:"require_pr_reviews,omitempty" json:"require_pr_reviews,omitempty"`
}

// SecurityConfig contains security settings
type SecurityConfig struct {
	// Enable vulnerability alerts
	VulnerabilityAlerts bool `yaml:"vulnerability_alerts,omitempty" json:"vulnerability_alerts,omitempty"`

	// Enable automated security fixes
	AutomatedSecurityFixes bool `yaml:"automated_security_fixes,omitempty" json:"automated_security_fixes,omitempty"`

	// Required security policies
	RequiredPolicies []string `yaml:"required_policies,omitempty" json:"required_policies,omitempty"`
}

// LegacyProviderConfig supports migration from bulk-clone.yaml format
type LegacyProviderConfig struct {
	// Legacy root path
	RootPath string `yaml:"root_path,omitempty" json:"root_path,omitempty"`

	// Legacy protocol
	Protocol string `yaml:"protocol,omitempty" json:"protocol,omitempty"`

	// Legacy organization name
	OrgName string `yaml:"org_name,omitempty" json:"org_name,omitempty"`

	// Legacy group name (GitLab)
	GroupName string `yaml:"group_name,omitempty" json:"group_name,omitempty"`

	// Legacy URL (for GitLab)
	URL string `yaml:"url,omitempty" json:"url,omitempty"`
}

// MigrationInfo contains information about configuration migration
type MigrationInfo struct {
	// Source format that was migrated from
	SourceFormat string `yaml:"source_format,omitempty" json:"source_format,omitempty"`

	// Migration date
	MigrationDate time.Time `yaml:"migration_date,omitempty" json:"migration_date,omitempty"`

	// Original configuration file path
	SourcePath string `yaml:"source_path,omitempty" json:"source_path,omitempty"`

	// Migration tool version
	ToolVersion string `yaml:"tool_version,omitempty" json:"tool_version,omitempty"`
}

// DefaultUnifiedConfig returns a default unified configuration
func DefaultUnifiedConfig() *UnifiedConfig {
	return &UnifiedConfig{
		Version:         "1.0.0",
		DefaultProvider: "github",
		Global: &GlobalSettings{
			CloneBaseDir:      "$HOME/repos",
			DefaultStrategy:   "reset",
			DefaultVisibility: "all",
			Timeouts: &TimeoutSettings{
				HTTPTimeout:      30 * time.Second,
				GitTimeout:       5 * time.Minute,
				RateLimitTimeout: 1 * time.Hour,
			},
			Concurrency: &ConcurrencySettings{
				CloneWorkers:  10,
				UpdateWorkers: 15,
				APIWorkers:    5,
			},
		},
		Providers: make(map[string]*ProviderConfig),
	}
}

// SupportedProviders returns a list of supported Git providers
func SupportedProviders() []string {
	return []string{"github", "gitlab", "gitea", "gogs"}
}

// ValidateProvider checks if a provider name is supported
func ValidateProvider(provider string) bool {
	for _, supported := range SupportedProviders() {
		if provider == supported {
			return true
		}
	}
	return false
}