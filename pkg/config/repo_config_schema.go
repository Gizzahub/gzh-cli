package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// RepoConfig represents the complete repository configuration schema.
type RepoConfig struct {
	Version       string                     `yaml:"version"`
	Organization  string                     `yaml:"organization"`
	Defaults      *RepoDefaults              `yaml:"defaults,omitempty"`
	Templates     map[string]*RepoTemplate   `yaml:"templates,omitempty"`
	Repositories  *RepoTargets               `yaml:"repositories,omitempty"`
	Policies      map[string]*PolicyTemplate `yaml:"policies,omitempty"`
	PolicyGroups  map[string]*PolicyGroup    `yaml:"policyGroups,omitempty"`  // Policy group configurations
	PolicyPresets map[string]*PolicyPreset   `yaml:"policyPresets,omitempty"` // Predefined policy sets (SOC2, ISO27001, etc.)
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
	Base          string              `yaml:"base,omitempty"`
	Description   string              `yaml:"description,omitempty"`
	Settings      *RepoSettings       `yaml:"settings,omitempty"`
	Security      *SecuritySettings   `yaml:"security,omitempty"`
	Permissions   *PermissionSettings `yaml:"permissions,omitempty"`
	Topics        []string            `yaml:"topics,omitempty"`
	RequiredFiles []string            `yaml:"requiredFiles,omitempty"`
	Webhooks      []string            `yaml:"webhooks,omitempty"`
	Environments  []string            `yaml:"environments,omitempty"`
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
	HasIssues      *bool `yaml:"hasIssues,omitempty"`
	HasProjects    *bool `yaml:"hasProjects,omitempty"`
	HasWiki        *bool `yaml:"hasWiki,omitempty"`
	HasDownloads   *bool `yaml:"hasDownloads,omitempty"`
	HasDiscussions *bool `yaml:"hasDiscussions,omitempty"`
	HasPages       *bool `yaml:"hasPages,omitempty"`

	// Merge settings
	AllowSquashMerge    *bool `yaml:"allowSquashMerge,omitempty"`
	AllowMergeCommit    *bool `yaml:"allowMergeCommit,omitempty"`
	AllowRebaseMerge    *bool `yaml:"allowRebaseMerge,omitempty"`
	AllowAutoMerge      *bool `yaml:"allowAutoMerge,omitempty"`
	DeleteBranchOnMerge *bool `yaml:"deleteBranchOnMerge,omitempty"`

	// Advanced settings
	AllowForking             *bool `yaml:"allowForking,omitempty"`
	WebCommitSignoffRequired *bool `yaml:"webCommitSignoffRequired,omitempty"`

	// Default branch
	DefaultBranch *string `yaml:"defaultBranch,omitempty"`
}

// SecuritySettings represents security-related settings.
type SecuritySettings struct {
	VulnerabilityAlerts           *bool                            `yaml:"vulnerabilityAlerts,omitempty"`
	AutomatedSecurityFixes        *bool                            `yaml:"automatedSecurityFixes,omitempty"`
	SecurityAdvisories            *bool                            `yaml:"securityAdvisories,omitempty"`
	PrivateVulnerabilityReporting *bool                            `yaml:"privateVulnerabilityReporting,omitempty"`
	SecretScanning                *bool                            `yaml:"secretScanning,omitempty"`
	SecretScanningPushProtection  *bool                            `yaml:"secretScanningPushProtection,omitempty"`
	BranchProtection              map[string]*BranchProtectionRule `yaml:"branchProtection,omitempty"`
	Webhooks                      []WebhookConfig                  `yaml:"webhooks,omitempty"`
}

// BranchProtectionRule represents branch protection settings.
type BranchProtectionRule struct {
	RequiredReviews               *int                       `yaml:"requiredReviews,omitempty"`
	DismissStaleReviews           *bool                      `yaml:"dismissStaleReviews,omitempty"`
	RequireCodeOwnerReviews       *bool                      `yaml:"requireCodeOwnerReviews,omitempty"`
	RequiredStatusChecks          []string                   `yaml:"requiredStatusChecks,omitempty"`
	StrictStatusChecks            *bool                      `yaml:"strictStatusChecks,omitempty"`
	RestrictPushes                *bool                      `yaml:"restrictPushes,omitempty"`
	AllowedUsers                  []string                   `yaml:"allowedUsers,omitempty"`
	AllowedTeams                  []string                   `yaml:"allowedTeams,omitempty"`
	RequireUpToDateBranch         *bool                      `yaml:"requireUpToDateBranch,omitempty"`
	EnforceAdmins                 *bool                      `yaml:"enforceAdmins,omitempty"`
	RequireConversationResolution *bool                      `yaml:"requireConversationResolution,omitempty"`
	AllowForcePushes              *bool                      `yaml:"allowForcePushes,omitempty"`
	AllowDeletions                *bool                      `yaml:"allowDeletions,omitempty"`
	DeploymentProtectionRules     []DeploymentProtectionRule `yaml:"deploymentProtectionRules,omitempty"`
}

// DeploymentProtectionRule represents deployment protection settings.
type DeploymentProtectionRule struct {
	Environment string `yaml:"environment"`
}

// WebhookConfig represents webhook configuration.
type WebhookConfig struct {
	URL         string   `yaml:"url"`
	Events      []string `yaml:"events"`
	Active      *bool    `yaml:"active,omitempty"`
	ContentType string   `yaml:"contentType,omitempty"`
	Secret      string   `yaml:"secret,omitempty"`
}

// PermissionSettings represents permission-related settings.
type PermissionSettings struct {
	TeamPermissions map[string]string `yaml:"teamPermissions,omitempty"`
	UserPermissions map[string]string `yaml:"userPermissions,omitempty"`
}

// PolicyTemplate represents a reusable policy configuration.
type PolicyTemplate struct {
	Description string                `yaml:"description"`
	Group       string                `yaml:"group,omitempty"`    // Policy group: security, compliance, best-practice, custom
	Severity    string                `yaml:"severity,omitempty"` // Overall severity: critical, high, medium, low
	Rules       map[string]PolicyRule `yaml:"rules"`
	Tags        []string              `yaml:"tags,omitempty"` // Additional categorization tags
}

// PolicyRule represents a single policy rule.
type PolicyRule struct {
	Type        string      `yaml:"type"`
	Value       interface{} `yaml:"value"`
	Enforcement string      `yaml:"enforcement"`
	Message     string      `yaml:"message,omitempty"`
}

// PolicyException represents an exception to a policy rule.
type PolicyException struct {
	PolicyName   string   `yaml:"policy"`
	RuleName     string   `yaml:"rule"`
	Reason       string   `yaml:"reason"`
	ApprovedBy   string   `yaml:"approvedBy"`
	ApprovalDate string   `yaml:"approvalDate,omitempty"`
	ExpiresAt    string   `yaml:"expiresAt,omitempty"`
	Conditions   []string `yaml:"conditions,omitempty"`
}

// PolicyGroup represents a group of related policies.
type PolicyGroup struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Weight      float64  `yaml:"weight"`             // Group weight in overall scoring (0.0-1.0)
	Policies    []string `yaml:"policies"`           // List of policy names in this group
	Required    bool     `yaml:"required,omitempty"` // Whether all policies in group must pass
	Tags        []string `yaml:"tags,omitempty"`     // Additional categorization
}

// PolicyPreset represents a predefined set of policies for compliance frameworks.
type PolicyPreset struct {
	Name        string                    `yaml:"name"`
	Description string                    `yaml:"description"`
	Framework   string                    `yaml:"framework"`           // SOC2, ISO27001, NIST, PCI-DSS, etc.
	Version     string                    `yaml:"version,omitempty"`   // Framework version
	Groups      []string                  `yaml:"groups"`              // Policy groups to include
	Policies    []string                  `yaml:"policies"`            // Individual policies to include
	Overrides   map[string]PolicyOverride `yaml:"overrides,omitempty"` // Policy-specific overrides
}

// PolicyOverride allows customizing policy rules for specific presets.
type PolicyOverride struct {
	Enforcement string                  `yaml:"enforcement,omitempty"`
	Rules       map[string]RuleOverride `yaml:"rules,omitempty"`
}

// RuleOverride allows customizing individual rules.
type RuleOverride struct {
	Value       interface{} `yaml:"value,omitempty"`
	Enforcement string      `yaml:"enforcement,omitempty"`
	Disabled    bool        `yaml:"disabled,omitempty"`
}

// LoadRepoConfig loads repository configuration from a YAML file.
func LoadRepoConfig(path string) (*RepoConfig, error) {
	// Expand environment variables in path
	path = os.ExpandEnv(path)

	// Read the file
	data, err := os.ReadFile(path) //nolint:gosec // path is user-provided configuration file path
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

// expandRepoConfigEnvVars expands environment variables in the configuration.
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

// validateRepoConfig validates the repository configuration.
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

// validateTemplateInheritance checks for circular dependencies in template inheritance.
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

// MergeRepoConfigs merges multiple repository configurations with priority.
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

// ValidateTemplateOverrides checks for conflicts in template overrides.
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

// checkTemplateConflicts identifies potential conflicts between derived and base templates.
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

// isHigherPermission checks if perm1 grants more access than perm2.
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

// GetTemplateInheritanceChain returns the inheritance chain for a template.
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

// GetAllTemplateChains returns inheritance chains for all templates.
func (rc *RepoConfig) GetAllTemplateChains() map[string][]string {
	chains := make(map[string][]string)

	for name := range rc.Templates {
		if chain, err := rc.GetTemplateInheritanceChain(name); err == nil {
			chains[name] = chain
		}
	}

	return chains
}

// resolveTemplate recursively resolves a template and all its base templates.
func (rc *RepoConfig) resolveTemplate(templateName string) (*RepoTemplate, error) {
	return rc.resolveTemplateWithChain(templateName, []string{})
}

// resolveTemplateWithChain recursively resolves a template with circular dependency checking.
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

// GetEffectiveConfig returns the effective configuration for a specific repository.
func (rc *RepoConfig) GetEffectiveConfig(repoName string) (*RepoSettings, *SecuritySettings, *PermissionSettings, []PolicyException, error) {
	config := &effectiveConfig{}

	// Start with defaults
	if rc.Defaults != nil {
		rc.applyConfigLayer(config, rc.Defaults.Template, rc.Defaults.Settings, rc.Defaults.Security, rc.Defaults.Permissions, nil)
	}

	// Apply repository-specific configuration
	if rc.Repositories != nil {
		// Check specific repositories
		for _, specific := range rc.Repositories.Specific {
			if specific.Name == repoName {
				rc.applyConfigLayer(config, specific.Template, specific.Settings, specific.Security, specific.Permissions, specific.Exceptions)
				return config.settings, config.security, config.permissions, config.exceptions, nil
			}
		}

		// Check patterns
		for _, pattern := range rc.Repositories.Patterns {
			if matched, _ := matchPattern(repoName, pattern.Match); matched {
				rc.applyConfigLayer(config, pattern.Template, pattern.Settings, pattern.Security, pattern.Permissions, pattern.Exceptions)
			}
		}

		// Apply default if exists
		if rc.Repositories.Default != nil {
			rc.applyConfigLayer(config, rc.Repositories.Default.Template, rc.Repositories.Default.Settings, rc.Repositories.Default.Security, rc.Repositories.Default.Permissions, nil)
		}
	}

	return config.settings, config.security, config.permissions, config.exceptions, nil
}

// effectiveConfig holds the accumulated configuration during processing.
type effectiveConfig struct {
	settings    *RepoSettings
	security    *SecuritySettings
	permissions *PermissionSettings
	exceptions  []PolicyException
}

// applyConfigLayer applies a configuration layer (template + direct settings).
func (rc *RepoConfig) applyConfigLayer(config *effectiveConfig, templateName string, settings *RepoSettings, security *SecuritySettings, permissions *PermissionSettings, exceptions []PolicyException) {
	// Apply template if specified
	if templateName != "" {
		template, err := rc.resolveTemplate(templateName)
		if err == nil {
			config.settings = mergeRepoSettings(config.settings, template.Settings)
			config.security = mergeSecuritySettings(config.security, template.Security)
			config.permissions = mergePermissionSettings(config.permissions, template.Permissions)
		}
	}

	// Apply direct settings
	config.settings = mergeRepoSettings(config.settings, settings)
	config.security = mergeSecuritySettings(config.security, security)
	config.permissions = mergePermissionSettings(config.permissions, permissions)
	config.exceptions = append(config.exceptions, exceptions...)
}

// ValidatePolicyExceptions validates all policy exceptions in the configuration.
func (rc *RepoConfig) ValidatePolicyExceptions() []string {
	var errors []string

	// Validate specific repository exceptions
	if rc.Repositories != nil {
		for _, specific := range rc.Repositories.Specific {
			for _, exception := range specific.Exceptions {
				if err := validatePolicyException(exception, rc.Policies); err != nil {
					errors = append(errors, fmt.Sprintf("Repository '%s': %v", specific.Name, err))
				}
			}
		}

		// Validate pattern-based exceptions
		for _, pattern := range rc.Repositories.Patterns {
			for _, exception := range pattern.Exceptions {
				if err := validatePolicyException(exception, rc.Policies); err != nil {
					errors = append(errors, fmt.Sprintf("Pattern '%s': %v", pattern.Match, err))
				}
			}
		}
	}

	return errors
}

// validatePolicyException validates a single policy exception.
func validatePolicyException(exception PolicyException, policies map[string]*PolicyTemplate) error {
	// Check if policy exists
	policy, exists := policies[exception.PolicyName]
	if !exists {
		return fmt.Errorf("exception references non-existent policy '%s'", exception.PolicyName)
	}

	// Check if rule exists
	if _, exists := policy.Rules[exception.RuleName]; !exists {
		return fmt.Errorf("exception references non-existent rule '%s' in policy '%s'",
			exception.RuleName, exception.PolicyName)
	}

	// Validate required fields
	if exception.Reason == "" {
		return fmt.Errorf("exception for policy '%s' rule '%s' missing required 'reason'",
			exception.PolicyName, exception.RuleName)
	}

	if exception.ApprovedBy == "" {
		return fmt.Errorf("exception for policy '%s' rule '%s' missing required 'approved_by'",
			exception.PolicyName, exception.RuleName)
	}

	// Validate expiration date format if provided
	if exception.ExpiresAt != "" {
		// Simple date format validation (could be enhanced)
		if len(exception.ExpiresAt) < 10 {
			return fmt.Errorf("exception for policy '%s' rule '%s' has invalid expiration date format",
				exception.PolicyName, exception.RuleName)
		}
	}

	return nil
}

// GetPolicyExceptionReport generates a report of all policy exceptions.
func (rc *RepoConfig) GetPolicyExceptionReport() map[string][]PolicyExceptionReport {
	report := make(map[string][]PolicyExceptionReport)

	if rc.Repositories != nil {
		// Process specific repositories
		for _, specific := range rc.Repositories.Specific {
			if len(specific.Exceptions) > 0 {
				report[specific.Name] = make([]PolicyExceptionReport, 0, len(specific.Exceptions))
				for _, exception := range specific.Exceptions {
					report[specific.Name] = append(report[specific.Name], PolicyExceptionReport{
						Repository: specific.Name,
						Exception:  exception,
						Type:       "specific",
					})
				}
			}
		}

		// Process pattern-based exceptions
		for _, pattern := range rc.Repositories.Patterns {
			if len(pattern.Exceptions) > 0 {
				patternKey := fmt.Sprintf("pattern:%s", pattern.Match)

				report[patternKey] = make([]PolicyExceptionReport, 0, len(pattern.Exceptions))
				for _, exception := range pattern.Exceptions {
					report[patternKey] = append(report[patternKey], PolicyExceptionReport{
						Repository: pattern.Match,
						Exception:  exception,
						Type:       "pattern",
					})
				}
			}
		}
	}

	return report
}

// PolicyExceptionReport represents a policy exception in the report.
type PolicyExceptionReport struct {
	Repository string
	Exception  PolicyException
	Type       string // "specific" or "pattern"
}

// IsExceptionActive checks if an exception is currently active.
func (e PolicyException) IsExceptionActive() bool {
	if e.ExpiresAt == "" {
		return true // No expiration means always active
	}

	// Simple date comparison (could be enhanced with proper date parsing)
	// For now, we'll assume the date format is valid
	return true
}

// mergeRepoSettings merges two RepoSettings, with the second taking precedence.
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

	// Override with new values using helper functions
	if override != nil {
		mergeStringField(&result.Description, override.Description)
		mergeStringField(&result.Homepage, override.Homepage)
		mergeStringField(&result.DefaultBranch, override.DefaultBranch)

		mergeBoolField(&result.Private, override.Private)
		mergeBoolField(&result.Archived, override.Archived)
		mergeBoolField(&result.HasIssues, override.HasIssues)
		mergeBoolField(&result.HasProjects, override.HasProjects)
		mergeBoolField(&result.HasWiki, override.HasWiki)
		mergeBoolField(&result.HasDownloads, override.HasDownloads)
		mergeBoolField(&result.AllowSquashMerge, override.AllowSquashMerge)
		mergeBoolField(&result.AllowMergeCommit, override.AllowMergeCommit)
		mergeBoolField(&result.AllowRebaseMerge, override.AllowRebaseMerge)
		mergeBoolField(&result.DeleteBranchOnMerge, override.DeleteBranchOnMerge)

		if override.Topics != nil {
			result.Topics = make([]string, len(override.Topics))
			copy(result.Topics, override.Topics)
		}
	}

	return result
}

// mergeStringField merges a string pointer field if the override is not nil.
func mergeStringField(target **string, override *string) {
	if override != nil {
		*target = override
	}
}

// mergeBoolField merges a bool pointer field if the override is not nil.
func mergeBoolField(target **bool, override *bool) {
	if override != nil {
		*target = override
	}
}

// mergeSecuritySettings merges two SecuritySettings, with the second taking precedence.
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

// mergePermissionSettings merges two PermissionSettings, with the second taking precedence.
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

// copyBranchProtectionRule creates a deep copy of a BranchProtectionRule.
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

// matchPattern checks if a string matches a pattern (simple glob support).
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
