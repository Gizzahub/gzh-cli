package repoconfig

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// RepoConfig represents the complete repository configuration schema
type RepoConfig struct {
	Version      string                     `yaml:"version"`
	Organization string                     `yaml:"organization"`
	Defaults     *RepoDefaults              `yaml:"defaults,omitempty"`
	Templates    map[string]*RepoTemplate   `yaml:"templates,omitempty"`
	Repositories *RepoTargets               `yaml:"repositories,omitempty"`
	Policies     map[string]*PolicyTemplate `yaml:"policies,omitempty"`
}

// RepoDefaults represents default settings for all repositories
type RepoDefaults struct {
	Template    string              `yaml:"template,omitempty"`
	Settings    *RepoSettings       `yaml:"settings,omitempty"`
	Security    *SecuritySettings   `yaml:"security,omitempty"`
	Permissions *PermissionSettings `yaml:"permissions,omitempty"`
}

// RepoTemplate represents a reusable configuration template
type RepoTemplate struct {
	Base        string              `yaml:"base,omitempty"`
	Description string              `yaml:"description,omitempty"`
	Settings    *RepoSettings       `yaml:"settings,omitempty"`
	Security    *SecuritySettings   `yaml:"security,omitempty"`
	Permissions *PermissionSettings `yaml:"permissions,omitempty"`
}

// RepoTargets represents repository targeting configuration
type RepoTargets struct {
	Specific []RepoSpecificConfig `yaml:"specific,omitempty"`
	Patterns []RepoPatternConfig  `yaml:"patterns,omitempty"`
	Default  *RepoDefaultConfig   `yaml:"default,omitempty"`
}

// RepoSpecificConfig represents configuration for specific repositories
type RepoSpecificConfig struct {
	Name        string              `yaml:"name"`
	Template    string              `yaml:"template,omitempty"`
	Settings    *RepoSettings       `yaml:"settings,omitempty"`
	Security    *SecuritySettings   `yaml:"security,omitempty"`
	Permissions *PermissionSettings `yaml:"permissions,omitempty"`
	Exceptions  []PolicyException   `yaml:"exceptions,omitempty"`
}

// RepoPatternConfig represents configuration for repositories matching patterns
type RepoPatternConfig struct {
	Match       string              `yaml:"match"`
	Template    string              `yaml:"template,omitempty"`
	Settings    *RepoSettings       `yaml:"settings,omitempty"`
	Security    *SecuritySettings   `yaml:"security,omitempty"`
	Permissions *PermissionSettings `yaml:"permissions,omitempty"`
	Exceptions  []PolicyException   `yaml:"exceptions,omitempty"`
}

// RepoDefaultConfig represents default configuration for all repositories
type RepoDefaultConfig struct {
	Template    string              `yaml:"template,omitempty"`
	Settings    *RepoSettings       `yaml:"settings,omitempty"`
	Security    *SecuritySettings   `yaml:"security,omitempty"`
	Permissions *PermissionSettings `yaml:"permissions,omitempty"`
}

// RepoSettings represents basic repository settings
type RepoSettings struct {
	Description *string  `yaml:"description,omitempty"`
	Homepage    *string  `yaml:"homepage,omitempty"`
	Topics      []string `yaml:"topics,omitempty"`
	Private     *bool    `yaml:"private,omitempty"`
	Archived    *bool    `yaml:"archived,omitempty"`

	// Features
	HasIssues    *bool `yaml:"has_issues,omitempty"`
	HasProjects  *bool `yaml:"has_projects,omitempty"`
	HasWiki      *bool `yaml:"has_wiki,omitempty"`
	HasDownloads *bool `yaml:"has_downloads,omitempty"`

	// Merge settings
	AllowSquashMerge    *bool `yaml:"allow_squash_merge,omitempty"`
	AllowMergeCommit    *bool `yaml:"allow_merge_commit,omitempty"`
	AllowRebaseMerge    *bool `yaml:"allow_rebase_merge,omitempty"`
	DeleteBranchOnMerge *bool `yaml:"delete_branch_on_merge,omitempty"`

	// Default branch
	DefaultBranch *string `yaml:"default_branch,omitempty"`
}

// SecuritySettings represents security-related settings
type SecuritySettings struct {
	VulnerabilityAlerts           *bool                            `yaml:"vulnerability_alerts,omitempty"`
	SecurityAdvisories            *bool                            `yaml:"security_advisories,omitempty"`
	PrivateVulnerabilityReporting *bool                            `yaml:"private_vulnerability_reporting,omitempty"`
	BranchProtection              map[string]*BranchProtectionRule `yaml:"branch_protection,omitempty"`
	Webhooks                      []WebhookConfig                  `yaml:"webhooks,omitempty"`
}

// BranchProtectionRule represents branch protection settings
type BranchProtectionRule struct {
	RequiredReviews               *int     `yaml:"required_reviews,omitempty"`
	DismissStaleReviews           *bool    `yaml:"dismiss_stale_reviews,omitempty"`
	RequireCodeOwnerReviews       *bool    `yaml:"require_code_owner_reviews,omitempty"`
	RequiredStatusChecks          []string `yaml:"required_status_checks,omitempty"`
	StrictStatusChecks            *bool    `yaml:"strict_status_checks,omitempty"`
	RestrictPushes                *bool    `yaml:"restrict_pushes,omitempty"`
	AllowedUsers                  []string `yaml:"allowed_users,omitempty"`
	AllowedTeams                  []string `yaml:"allowed_teams,omitempty"`
	RequireUpToDateBranch         *bool    `yaml:"require_up_to_date_branch,omitempty"`
	EnforceAdmins                 *bool    `yaml:"enforce_admins,omitempty"`
	RequireConversationResolution *bool    `yaml:"require_conversation_resolution,omitempty"`
	AllowForcePushes              *bool    `yaml:"allow_force_pushes,omitempty"`
	AllowDeletions                *bool    `yaml:"allow_deletions,omitempty"`
}

// WebhookConfig represents webhook configuration
type WebhookConfig struct {
	URL         string   `yaml:"url"`
	Events      []string `yaml:"events"`
	Active      *bool    `yaml:"active,omitempty"`
	ContentType string   `yaml:"content_type,omitempty"`
	Secret      string   `yaml:"secret,omitempty"`
}

// PermissionSettings represents permission-related settings
type PermissionSettings struct {
	TeamPermissions map[string]string `yaml:"team_permissions,omitempty"`
	UserPermissions map[string]string `yaml:"user_permissions,omitempty"`
}

// PolicyTemplate represents a reusable policy configuration
type PolicyTemplate struct {
	Description string                `yaml:"description"`
	Rules       map[string]PolicyRule `yaml:"rules"`
}

// PolicyRule represents a single policy rule
type PolicyRule struct {
	Type        string      `yaml:"type"`
	Value       interface{} `yaml:"value"`
	Enforcement string      `yaml:"enforcement"`
	Message     string      `yaml:"message,omitempty"`
}

// LoadRepoConfig loads a repository configuration from a YAML file
func LoadRepoConfig(path string) (*RepoConfig, error) {
	// Expand environment variables in path
	path = os.ExpandEnv(path)

	// Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var config RepoConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate version
	if config.Version == "" {
		return nil, fmt.Errorf("config version is required")
	}

	return &config, nil
}