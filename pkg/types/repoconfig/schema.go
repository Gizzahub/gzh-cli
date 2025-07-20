package repoconfig

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// RepoConfig represents the complete repository configuration schema.
type RepoConfig struct {
	Version      string                     `yaml:"version"`
	Organization string                     `yaml:"organization"`
	Defaults     *RepoDefaults              `yaml:"defaults,omitempty"`
	Templates    map[string]*RepoTemplate   `yaml:"templates,omitempty"`
	Repositories *RepoTargets               `yaml:"repositories,omitempty"`
	Policies     map[string]*PolicyTemplate `yaml:"policies,omitempty"`
}

// RepoDefaults represents default settings for all repositories.
type RepoDefaults struct {
	Template    string              `yaml:"template,omitempty"`
	Settings    *RepoSettings       `yaml:"settings,omitempty"`
	Security    *SecuritySettings   `yaml:"security,omitempty"`
	Permissions *PermissionSettings `yaml:"permissions,omitempty"`
}

// RepoTemplate represents a reusable configuration template.
type RepoTemplate struct {
	Base        string              `yaml:"base,omitempty"`
	Description string              `yaml:"description,omitempty"`
	Settings    *RepoSettings       `yaml:"settings,omitempty"`
	Security    *SecuritySettings   `yaml:"security,omitempty"`
	Permissions *PermissionSettings `yaml:"permissions,omitempty"`
}

// RepoTargets represents repository targeting configuration.
type RepoTargets struct {
	Specific []RepoSpecificConfig `yaml:"specific,omitempty"`
	Patterns []RepoPatternConfig  `yaml:"patterns,omitempty"`
	Default  *RepoDefaultConfig   `yaml:"default,omitempty"`
}

// RepoSpecificConfig represents configuration for specific repositories.
type RepoSpecificConfig struct {
	Name        string              `yaml:"name"`
	Template    string              `yaml:"template,omitempty"`
	Settings    *RepoSettings       `yaml:"settings,omitempty"`
	Security    *SecuritySettings   `yaml:"security,omitempty"`
	Permissions *PermissionSettings `yaml:"permissions,omitempty"`
	Exceptions  []PolicyException   `yaml:"exceptions,omitempty"`
}

// RepoPatternConfig represents configuration for repositories matching patterns.
type RepoPatternConfig struct {
	Match       string              `yaml:"match"`
	Template    string              `yaml:"template,omitempty"`
	Settings    *RepoSettings       `yaml:"settings,omitempty"`
	Security    *SecuritySettings   `yaml:"security,omitempty"`
	Permissions *PermissionSettings `yaml:"permissions,omitempty"`
	Exceptions  []PolicyException   `yaml:"exceptions,omitempty"`
}

// RepoDefaultConfig represents default configuration for all repositories.
type RepoDefaultConfig struct {
	Template    string              `yaml:"template,omitempty"`
	Settings    *RepoSettings       `yaml:"settings,omitempty"`
	Security    *SecuritySettings   `yaml:"security,omitempty"`
	Permissions *PermissionSettings `yaml:"permissions,omitempty"`
}

// RepoSettings represents basic repository settings.
type RepoSettings struct {
	Description *string  `yaml:"description,omitempty"`
	Homepage    *string  `yaml:"homepage,omitempty"`
	Topics      []string `yaml:"topics,omitempty"`
	Private     *bool    `yaml:"private,omitempty"`
	Archived    *bool    `yaml:"archived,omitempty"`

	// Features
	HasIssues    *bool `yaml:"hasIssues,omitempty"`
	HasProjects  *bool `yaml:"hasProjects,omitempty"`
	HasWiki      *bool `yaml:"hasWiki,omitempty"`
	HasDownloads *bool `yaml:"hasDownloads,omitempty"`

	// Merge settings
	AllowSquashMerge    *bool `yaml:"allowSquashMerge,omitempty"`
	AllowMergeCommit    *bool `yaml:"allowMergeCommit,omitempty"`
	AllowRebaseMerge    *bool `yaml:"allowRebaseMerge,omitempty"`
	DeleteBranchOnMerge *bool `yaml:"deleteBranchOnMerge,omitempty"`

	// Default branch
	DefaultBranch *string `yaml:"defaultBranch,omitempty"`
}

// SecuritySettings represents security-related settings.
type SecuritySettings struct {
	VulnerabilityAlerts           *bool                            `yaml:"vulnerabilityAlerts,omitempty"`
	SecurityAdvisories            *bool                            `yaml:"securityAdvisories,omitempty"`
	PrivateVulnerabilityReporting *bool                            `yaml:"privateVulnerabilityReporting,omitempty"`
	BranchProtection              map[string]*BranchProtectionRule `yaml:"branchProtection,omitempty"`
	Webhooks                      []WebhookConfig                  `yaml:"webhooks,omitempty"`
}

// BranchProtectionRule represents branch protection settings.
type BranchProtectionRule struct {
	RequiredReviews               *int     `yaml:"requiredReviews,omitempty"`
	DismissStaleReviews           *bool    `yaml:"dismissStaleReviews,omitempty"`
	RequireCodeOwnerReviews       *bool    `yaml:"requireCodeOwnerReviews,omitempty"`
	RequiredStatusChecks          []string `yaml:"requiredStatusChecks,omitempty"`
	StrictStatusChecks            *bool    `yaml:"strictStatusChecks,omitempty"`
	RestrictPushes                *bool    `yaml:"restrictPushes,omitempty"`
	AllowedUsers                  []string `yaml:"allowedUsers,omitempty"`
	AllowedTeams                  []string `yaml:"allowedTeams,omitempty"`
	RequireUpToDateBranch         *bool    `yaml:"requireUpToDateBranch,omitempty"`
	EnforceAdmins                 *bool    `yaml:"enforceAdmins,omitempty"`
	RequireConversationResolution *bool    `yaml:"requireConversationResolution,omitempty"`
	AllowForcePushes              *bool    `yaml:"allowForcePushes,omitempty"`
	AllowDeletions                *bool    `yaml:"allowDeletions,omitempty"`
}

// WebhookConfig represents webhook configuration.
type WebhookConfig struct {
	URL         string   `yaml:"url" json:"url"`
	Events      []string `yaml:"events" json:"events"`
	Active      *bool    `yaml:"active,omitempty" json:"active,omitempty"`
	ContentType string   `yaml:"contentType,omitempty" json:"contentType,omitempty"`
	Secret      string   `yaml:"secret,omitempty" json:"secret,omitempty"`
}

// PermissionSettings represents permission-related settings.
type PermissionSettings struct {
	TeamPermissions map[string]string `yaml:"teamPermissions,omitempty"`
	UserPermissions map[string]string `yaml:"userPermissions,omitempty"`
}

// PolicyTemplate represents a reusable policy configuration.
type PolicyTemplate struct {
	Description string                `yaml:"description"`
	Rules       map[string]PolicyRule `yaml:"rules"`
}

// PolicyRule represents a single policy rule.
type PolicyRule struct {
	Type        string      `yaml:"type"`
	Value       interface{} `yaml:"value"`
	Enforcement string      `yaml:"enforcement"`
	Message     string      `yaml:"message,omitempty"`
}

// LoadRepoConfig loads a repository configuration from a YAML file.
func LoadRepoConfig(path string) (*RepoConfig, error) {
	// Expand environment variables in path
	path = os.ExpandEnv(path)

	// Read the file
	data, err := os.ReadFile(path) //nolint:gosec // Loading config files is a legitimate use case
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
