package config

import (
	"fmt"
	"os"
	"strings"

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
}

// RepoPatternConfig represents configuration for repositories matching patterns
type RepoPatternConfig struct {
	Match       string              `yaml:"match"`
	Template    string              `yaml:"template,omitempty"`
	Settings    *RepoSettings       `yaml:"settings,omitempty"`
	Security    *SecuritySettings   `yaml:"security,omitempty"`
	Permissions *PermissionSettings `yaml:"permissions,omitempty"`
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
}

// LoadRepoConfig loads repository configuration from a YAML file
func LoadRepoConfig(path string) (*RepoConfig, error) {
	// Expand environment variables in path
	path = os.ExpandEnv(path)

	// Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read repo config file: %w", err)
	}

	// Parse YAML
	var config RepoConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse repo config YAML: %w", err)
	}

	// Expand environment variables in the config
	if err := expandRepoConfigEnvVars(&config); err != nil {
		return nil, fmt.Errorf("failed to expand environment variables: %w", err)
	}

	// Validate the configuration
	if err := validateRepoConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid repo config: %w", err)
	}

	return &config, nil
}

// expandRepoConfigEnvVars expands environment variables in the configuration
func expandRepoConfigEnvVars(config *RepoConfig) error {
	// Expand environment variables in webhook URLs and secrets
	if config.Templates != nil {
		for _, template := range config.Templates {
			if template.Security != nil && template.Security.Webhooks != nil {
				for i := range template.Security.Webhooks {
					template.Security.Webhooks[i].URL = os.ExpandEnv(template.Security.Webhooks[i].URL)
					template.Security.Webhooks[i].Secret = os.ExpandEnv(template.Security.Webhooks[i].Secret)
				}
			}
		}
	}

	// Expand in repository-specific configs
	if config.Repositories != nil && config.Repositories.Specific != nil {
		for i := range config.Repositories.Specific {
			repo := &config.Repositories.Specific[i]
			if repo.Security != nil && repo.Security.Webhooks != nil {
				for j := range repo.Security.Webhooks {
					repo.Security.Webhooks[j].URL = os.ExpandEnv(repo.Security.Webhooks[j].URL)
					repo.Security.Webhooks[j].Secret = os.ExpandEnv(repo.Security.Webhooks[j].Secret)
				}
			}
		}
	}

	return nil
}

// validateRepoConfig validates the repository configuration
func validateRepoConfig(config *RepoConfig) error {
	// Check version
	if config.Version == "" {
		return fmt.Errorf("version is required")
	}

	// Check organization
	if config.Organization == "" {
		return fmt.Errorf("organization is required")
	}

	// Validate templates
	if config.Templates != nil {
		for name, template := range config.Templates {
			// Check for circular dependencies
			if err := validateTemplateInheritance(name, template, config.Templates); err != nil {
				return fmt.Errorf("template '%s': %w", name, err)
			}
		}
	}

	// Validate repository patterns
	if config.Repositories != nil && config.Repositories.Patterns != nil {
		for i, pattern := range config.Repositories.Patterns {
			if pattern.Match == "" {
				return fmt.Errorf("repository pattern %d: match is required", i)
			}
		}
	}

	return nil
}

// validateTemplateInheritance checks for circular dependencies in template inheritance
func validateTemplateInheritance(name string, template *RepoTemplate, templates map[string]*RepoTemplate) error {
	visited := make(map[string]bool)
	return checkTemplateInheritance(name, template, templates, visited)
}

func checkTemplateInheritance(name string, template *RepoTemplate, templates map[string]*RepoTemplate, visited map[string]bool) error {
	if visited[name] {
		return fmt.Errorf("circular dependency detected")
	}

	if template.Base == "" {
		return nil
	}

	visited[name] = true

	baseTemplate, ok := templates[template.Base]
	if !ok {
		return fmt.Errorf("base template '%s' not found", template.Base)
	}

	return checkTemplateInheritance(template.Base, baseTemplate, templates, visited)
}

// MergeRepoConfigs merges multiple repository configurations with priority
func MergeRepoConfigs(configs ...*RepoConfig) (*RepoConfig, error) {
	if len(configs) == 0 {
		return nil, fmt.Errorf("no configurations to merge")
	}

	// Start with the first config as base
	result := &RepoConfig{
		Version:      configs[0].Version,
		Organization: configs[0].Organization,
		Templates:    make(map[string]*RepoTemplate),
		Policies:     make(map[string]*PolicyTemplate),
	}

	// Merge all configs
	for _, config := range configs {
		if config == nil {
			continue
		}

		// Merge templates
		for name, template := range config.Templates {
			result.Templates[name] = template
		}

		// Merge policies
		for name, policy := range config.Policies {
			result.Policies[name] = policy
		}

		// Take the last non-nil defaults
		if config.Defaults != nil {
			result.Defaults = config.Defaults
		}

		// Take the last non-nil repositories
		if config.Repositories != nil {
			result.Repositories = config.Repositories
		}
	}

	return result, nil
}

// ValidateTemplateOverrides checks for conflicts in template overrides
func (rc *RepoConfig) ValidateTemplateOverrides() []string {
	var warnings []string

	// Check each template that has a base
	for name, template := range rc.Templates {
		if template.Base == "" {
			continue
		}

		baseTemplate, err := rc.resolveTemplate(template.Base)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("Template '%s': %v", name, err))
			continue
		}

		// Check for potential conflicts
		conflicts := checkTemplateConflicts(name, template, baseTemplate)
		warnings = append(warnings, conflicts...)
	}

	return warnings
}

// checkTemplateConflicts identifies potential conflicts between derived and base templates
func checkTemplateConflicts(templateName string, derived, base *RepoTemplate) []string {
	var conflicts []string

	// Check security settings conflicts
	if derived.Security != nil && base.Security != nil {
		// Check branch protection conflicts
		for branch, derivedRule := range derived.Security.BranchProtection {
			if baseRule, exists := base.Security.BranchProtection[branch]; exists {
				// Warn if derived has weaker protection
				if derivedRule.RequiredReviews != nil && baseRule.RequiredReviews != nil {
					if *derivedRule.RequiredReviews < *baseRule.RequiredReviews {
						conflicts = append(conflicts, fmt.Sprintf(
							"Template '%s': Reduces required reviews for branch '%s' from %d to %d",
							templateName, branch, *baseRule.RequiredReviews, *derivedRule.RequiredReviews))
					}
				}

				// Warn if derived disables protections that base enables
				if baseRule.EnforceAdmins != nil && *baseRule.EnforceAdmins &&
					derivedRule.EnforceAdmins != nil && !*derivedRule.EnforceAdmins {
					conflicts = append(conflicts, fmt.Sprintf(
						"Template '%s': Disables admin enforcement for branch '%s'",
						templateName, branch))
				}
			}
		}
	}

	// Check permission conflicts
	if derived.Permissions != nil && base.Permissions != nil {
		// Warn if derived grants higher permissions than base
		for team, derivedPerm := range derived.Permissions.TeamPermissions {
			if basePerm, exists := base.Permissions.TeamPermissions[team]; exists {
				if isHigherPermission(derivedPerm, basePerm) {
					conflicts = append(conflicts, fmt.Sprintf(
						"Template '%s': Escalates permissions for team '%s' from '%s' to '%s'",
						templateName, team, basePerm, derivedPerm))
				}
			}
		}
	}

	// Check repository visibility conflicts
	if derived.Settings != nil && base.Settings != nil {
		if base.Settings.Private != nil && *base.Settings.Private &&
			derived.Settings.Private != nil && !*derived.Settings.Private {
			conflicts = append(conflicts, fmt.Sprintf(
				"Template '%s': Changes repository from private to public",
				templateName))
		}
	}

	return conflicts
}

// isHigherPermission checks if perm1 grants more access than perm2
func isHigherPermission(perm1, perm2 string) bool {
	permissions := map[string]int{
		"read":     1,
		"triage":   2,
		"write":    3,
		"maintain": 4,
		"admin":    5,
	}

	level1, ok1 := permissions[strings.ToLower(perm1)]
	level2, ok2 := permissions[strings.ToLower(perm2)]

	if !ok1 || !ok2 {
		return false
	}

	return level1 > level2
}

// GetTemplateInheritanceChain returns the inheritance chain for a template
func (rc *RepoConfig) GetTemplateInheritanceChain(templateName string) ([]string, error) {
	chain := []string{}
	visited := make(map[string]bool)

	current := templateName
	for current != "" {
		if visited[current] {
			return nil, fmt.Errorf("circular dependency detected in template '%s'", templateName)
		}

		template, ok := rc.Templates[current]
		if !ok {
			return nil, fmt.Errorf("template '%s' not found", current)
		}

		chain = append(chain, current)
		visited[current] = true
		current = template.Base
	}

	return chain, nil
}

// GetAllTemplateChains returns inheritance chains for all templates
func (rc *RepoConfig) GetAllTemplateChains() map[string][]string {
	chains := make(map[string][]string)

	for name := range rc.Templates {
		if chain, err := rc.GetTemplateInheritanceChain(name); err == nil {
			chains[name] = chain
		}
	}

	return chains
}

// resolveTemplate recursively resolves a template and all its base templates
func (rc *RepoConfig) resolveTemplate(templateName string) (*RepoTemplate, error) {
	return rc.resolveTemplateWithChain(templateName, []string{})
}

// resolveTemplateWithChain recursively resolves a template with circular dependency checking
func (rc *RepoConfig) resolveTemplateWithChain(templateName string, chain []string) (*RepoTemplate, error) {
	// Check for circular dependency
	for _, name := range chain {
		if name == templateName {
			return nil, fmt.Errorf("circular template dependency detected: %s", strings.Join(append(chain, templateName), " -> "))
		}
	}

	template, ok := rc.Templates[templateName]
	if !ok {
		return nil, fmt.Errorf("template '%s' not found", templateName)
	}

	// If no base template, return as is
	if template.Base == "" {
		return template, nil
	}

	// Resolve base template
	baseTemplate, err := rc.resolveTemplateWithChain(template.Base, append(chain, templateName))
	if err != nil {
		return nil, err
	}

	// Create a new merged template
	result := &RepoTemplate{
		Description: template.Description,
		Settings:    mergeRepoSettings(baseTemplate.Settings, template.Settings),
		Security:    mergeSecuritySettings(baseTemplate.Security, template.Security),
		Permissions: mergePermissionSettings(baseTemplate.Permissions, template.Permissions),
	}

	// Merge other fields
	if len(template.Topics) > 0 {
		result.Topics = template.Topics
	} else if len(baseTemplate.Topics) > 0 {
		result.Topics = baseTemplate.Topics
	}

	if len(template.RequiredFiles) > 0 {
		result.RequiredFiles = template.RequiredFiles
	} else if len(baseTemplate.RequiredFiles) > 0 {
		result.RequiredFiles = baseTemplate.RequiredFiles
	}

	if len(template.Webhooks) > 0 {
		result.Webhooks = template.Webhooks
	} else if len(baseTemplate.Webhooks) > 0 {
		result.Webhooks = baseTemplate.Webhooks
	}

	if len(template.Environments) > 0 {
		result.Environments = template.Environments
	} else if len(baseTemplate.Environments) > 0 {
		result.Environments = baseTemplate.Environments
	}

	return result, nil
}

// GetEffectiveConfig returns the effective configuration for a specific repository
func (rc *RepoConfig) GetEffectiveConfig(repoName string) (*RepoSettings, *SecuritySettings, *PermissionSettings, error) {
	var settings *RepoSettings
	var security *SecuritySettings
	var permissions *PermissionSettings

	// Start with defaults
	if rc.Defaults != nil {
		if rc.Defaults.Template != "" {
			template, err := rc.resolveTemplate(rc.Defaults.Template)
			if err == nil {
				settings = mergeRepoSettings(settings, template.Settings)
				security = mergeSecuritySettings(security, template.Security)
				permissions = mergePermissionSettings(permissions, template.Permissions)
			}
		}
		settings = mergeRepoSettings(settings, rc.Defaults.Settings)
		security = mergeSecuritySettings(security, rc.Defaults.Security)
		permissions = mergePermissionSettings(permissions, rc.Defaults.Permissions)
	}

	// Apply repository-specific configuration
	if rc.Repositories != nil {
		// Check specific repositories
		for _, specific := range rc.Repositories.Specific {
			if specific.Name == repoName {
				if specific.Template != "" {
					template, err := rc.resolveTemplate(specific.Template)
					if err == nil {
						settings = mergeRepoSettings(settings, template.Settings)
						security = mergeSecuritySettings(security, template.Security)
						permissions = mergePermissionSettings(permissions, template.Permissions)
					}
				}
				settings = mergeRepoSettings(settings, specific.Settings)
				security = mergeSecuritySettings(security, specific.Security)
				permissions = mergePermissionSettings(permissions, specific.Permissions)
				return settings, security, permissions, nil
			}
		}

		// Check patterns
		for _, pattern := range rc.Repositories.Patterns {
			if matched, _ := matchPattern(repoName, pattern.Match); matched {
				if pattern.Template != "" {
					template, err := rc.resolveTemplate(pattern.Template)
					if err == nil {
						settings = mergeRepoSettings(settings, template.Settings)
						security = mergeSecuritySettings(security, template.Security)
						permissions = mergePermissionSettings(permissions, template.Permissions)
					}
				}
				settings = mergeRepoSettings(settings, pattern.Settings)
				security = mergeSecuritySettings(security, pattern.Security)
				permissions = mergePermissionSettings(permissions, pattern.Permissions)
			}
		}

		// Apply default if exists
		if rc.Repositories.Default != nil {
			if rc.Repositories.Default.Template != "" {
				template, err := rc.resolveTemplate(rc.Repositories.Default.Template)
				if err == nil {
					settings = mergeRepoSettings(settings, template.Settings)
					security = mergeSecuritySettings(security, template.Security)
					permissions = mergePermissionSettings(permissions, template.Permissions)
				}
			}
			settings = mergeRepoSettings(settings, rc.Repositories.Default.Settings)
			security = mergeSecuritySettings(security, rc.Repositories.Default.Security)
			permissions = mergePermissionSettings(permissions, rc.Repositories.Default.Permissions)
		}
	}

	return settings, security, permissions, nil
}

// mergeRepoSettings merges two RepoSettings, with the second taking precedence
func mergeRepoSettings(base, override *RepoSettings) *RepoSettings {
	if base == nil && override == nil {
		return nil
	}

	result := &RepoSettings{}

	// Copy from base
	if base != nil {
		*result = *base
		if base.Topics != nil {
			result.Topics = make([]string, len(base.Topics))
			copy(result.Topics, base.Topics)
		}
	}

	// Override with new values
	if override != nil {
		if override.Description != nil {
			result.Description = override.Description
		}
		if override.Homepage != nil {
			result.Homepage = override.Homepage
		}
		if override.Topics != nil {
			result.Topics = make([]string, len(override.Topics))
			copy(result.Topics, override.Topics)
		}
		if override.Private != nil {
			result.Private = override.Private
		}
		if override.Archived != nil {
			result.Archived = override.Archived
		}
		if override.HasIssues != nil {
			result.HasIssues = override.HasIssues
		}
		if override.HasProjects != nil {
			result.HasProjects = override.HasProjects
		}
		if override.HasWiki != nil {
			result.HasWiki = override.HasWiki
		}
		if override.HasDownloads != nil {
			result.HasDownloads = override.HasDownloads
		}
		if override.AllowSquashMerge != nil {
			result.AllowSquashMerge = override.AllowSquashMerge
		}
		if override.AllowMergeCommit != nil {
			result.AllowMergeCommit = override.AllowMergeCommit
		}
		if override.AllowRebaseMerge != nil {
			result.AllowRebaseMerge = override.AllowRebaseMerge
		}
		if override.DeleteBranchOnMerge != nil {
			result.DeleteBranchOnMerge = override.DeleteBranchOnMerge
		}
		if override.DefaultBranch != nil {
			result.DefaultBranch = override.DefaultBranch
		}
	}

	return result
}

// mergeSecuritySettings merges two SecuritySettings, with the second taking precedence
func mergeSecuritySettings(base, override *SecuritySettings) *SecuritySettings {
	if base == nil && override == nil {
		return nil
	}

	result := &SecuritySettings{
		BranchProtection: make(map[string]*BranchProtectionRule),
	}

	// Copy from base
	if base != nil {
		result.VulnerabilityAlerts = base.VulnerabilityAlerts
		result.SecurityAdvisories = base.SecurityAdvisories
		result.PrivateVulnerabilityReporting = base.PrivateVulnerabilityReporting

		// Deep copy branch protection
		for branch, rule := range base.BranchProtection {
			result.BranchProtection[branch] = copyBranchProtectionRule(rule)
		}

		// Copy webhooks
		if base.Webhooks != nil {
			result.Webhooks = make([]WebhookConfig, len(base.Webhooks))
			copy(result.Webhooks, base.Webhooks)
		}
	}

	// Override with new values
	if override != nil {
		if override.VulnerabilityAlerts != nil {
			result.VulnerabilityAlerts = override.VulnerabilityAlerts
		}
		if override.SecurityAdvisories != nil {
			result.SecurityAdvisories = override.SecurityAdvisories
		}
		if override.PrivateVulnerabilityReporting != nil {
			result.PrivateVulnerabilityReporting = override.PrivateVulnerabilityReporting
		}

		// Merge branch protection
		for branch, rule := range override.BranchProtection {
			result.BranchProtection[branch] = copyBranchProtectionRule(rule)
		}

		// Override webhooks (complete replacement)
		if override.Webhooks != nil {
			result.Webhooks = make([]WebhookConfig, len(override.Webhooks))
			copy(result.Webhooks, override.Webhooks)
		}
	}

	return result
}

// mergePermissionSettings merges two PermissionSettings, with the second taking precedence
func mergePermissionSettings(base, override *PermissionSettings) *PermissionSettings {
	if base == nil && override == nil {
		return nil
	}

	result := &PermissionSettings{
		TeamPermissions: make(map[string]string),
		UserPermissions: make(map[string]string),
	}

	// Copy from base
	if base != nil {
		for team, perm := range base.TeamPermissions {
			result.TeamPermissions[team] = perm
		}
		for user, perm := range base.UserPermissions {
			result.UserPermissions[user] = perm
		}
	}

	// Override with new values
	if override != nil {
		for team, perm := range override.TeamPermissions {
			result.TeamPermissions[team] = perm
		}
		for user, perm := range override.UserPermissions {
			result.UserPermissions[user] = perm
		}
	}

	return result
}

// copyBranchProtectionRule creates a deep copy of a BranchProtectionRule
func copyBranchProtectionRule(rule *BranchProtectionRule) *BranchProtectionRule {
	if rule == nil {
		return nil
	}

	result := &BranchProtectionRule{
		RequiredReviews:               rule.RequiredReviews,
		DismissStaleReviews:           rule.DismissStaleReviews,
		RequireCodeOwnerReviews:       rule.RequireCodeOwnerReviews,
		StrictStatusChecks:            rule.StrictStatusChecks,
		RestrictPushes:                rule.RestrictPushes,
		RequireUpToDateBranch:         rule.RequireUpToDateBranch,
		EnforceAdmins:                 rule.EnforceAdmins,
		RequireConversationResolution: rule.RequireConversationResolution,
		AllowForcePushes:              rule.AllowForcePushes,
		AllowDeletions:                rule.AllowDeletions,
	}

	if rule.RequiredStatusChecks != nil {
		result.RequiredStatusChecks = make([]string, len(rule.RequiredStatusChecks))
		copy(result.RequiredStatusChecks, rule.RequiredStatusChecks)
	}

	if rule.AllowedUsers != nil {
		result.AllowedUsers = make([]string, len(rule.AllowedUsers))
		copy(result.AllowedUsers, rule.AllowedUsers)
	}

	if rule.AllowedTeams != nil {
		result.AllowedTeams = make([]string, len(rule.AllowedTeams))
		copy(result.AllowedTeams, rule.AllowedTeams)
	}

	return result
}

// matchPattern checks if a string matches a pattern (simple glob support)
func matchPattern(str, pattern string) (bool, error) {
	// Simple implementation - can be enhanced with more sophisticated pattern matching
	if strings.Contains(pattern, "*") {
		// Convert simple glob to regex
		pattern = strings.ReplaceAll(pattern, ".", "\\.")
		pattern = strings.ReplaceAll(pattern, "*", ".*")
		pattern = "^" + pattern + "$"
		return strings.Contains(str, strings.Trim(pattern, "^.*$")), nil
	}
	return str == pattern, nil
}
