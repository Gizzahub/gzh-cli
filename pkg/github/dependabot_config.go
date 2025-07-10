package github

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// DependabotConfigManager manages Dependabot configurations for repositories
type DependabotConfigManager struct {
	logger    Logger
	apiClient APIClient
}

// DependabotConfig represents the complete Dependabot configuration
type DependabotConfig struct {
	Version int                    `yaml:"version" json:"version"`
	Updates []DependabotUpdateRule `yaml:"updates" json:"updates"`
	// Registries for private package managers
	Registries map[string]DependabotRegistry `yaml:"registries,omitempty" json:"registries,omitempty"`
}

// DependabotUpdateRule defines update rules for a package ecosystem
type DependabotUpdateRule struct {
	PackageEcosystem     string                     `yaml:"package-ecosystem" json:"package_ecosystem"`
	Directory            string                     `yaml:"directory" json:"directory"`
	Schedule             DependabotSchedule         `yaml:"schedule" json:"schedule"`
	VersioningStrategy   string                     `yaml:"versioning-strategy,omitempty" json:"versioning_strategy,omitempty"`
	AllowedUpdates       []DependabotAllowedUpdate  `yaml:"allow,omitempty" json:"allowed_updates,omitempty"`
	IgnoredDependencies  []DependabotIgnoredUpdate  `yaml:"ignore,omitempty" json:"ignored_dependencies,omitempty"`
	Reviewers            []string                   `yaml:"reviewers,omitempty" json:"reviewers,omitempty"`
	Assignees            []string                   `yaml:"assignees,omitempty" json:"assignees,omitempty"`
	Labels               []string                   `yaml:"labels,omitempty" json:"labels,omitempty"`
	PullRequestLimit     int                        `yaml:"open-pull-requests-limit,omitempty" json:"pull_request_limit,omitempty"`
	RebaseStrategy       string                     `yaml:"rebase-strategy,omitempty" json:"rebase_strategy,omitempty"`
	CommitMessage        *DependabotCommitMessage   `yaml:"commit-message,omitempty" json:"commit_message,omitempty"`
	Groups               map[string]DependabotGroup `yaml:"groups,omitempty" json:"groups,omitempty"`
	RegistriesConfig     []string                   `yaml:"registries,omitempty" json:"registries_config,omitempty"`
	VendorUpdates        bool                       `yaml:"vendor,omitempty" json:"vendor_updates,omitempty"`
	InsecureExternalCode bool                       `yaml:"insecure-external-code-execution,omitempty" json:"insecure_external_code,omitempty"`
}

// DependabotSchedule defines when Dependabot checks for updates
type DependabotSchedule struct {
	Interval string `yaml:"interval" json:"interval"`
	Day      string `yaml:"day,omitempty" json:"day,omitempty"`
	Time     string `yaml:"time,omitempty" json:"time,omitempty"`
	Timezone string `yaml:"timezone,omitempty" json:"timezone,omitempty"`
}

// DependabotAllowedUpdate defines which updates are allowed
type DependabotAllowedUpdate struct {
	DependencyType string `yaml:"dependency-type,omitempty" json:"dependency_type,omitempty"`
	DependencyName string `yaml:"dependency-name,omitempty" json:"dependency_name,omitempty"`
	UpdateType     string `yaml:"update-type,omitempty" json:"update_type,omitempty"`
}

// DependabotIgnoredUpdate defines dependencies to ignore
type DependabotIgnoredUpdate struct {
	DependencyName string   `yaml:"dependency-name" json:"dependency_name"`
	Versions       []string `yaml:"versions,omitempty" json:"versions,omitempty"`
	UpdateTypes    []string `yaml:"update-types,omitempty" json:"update_types,omitempty"`
}

// DependabotCommitMessage defines commit message preferences
type DependabotCommitMessage struct {
	Prefix            string `yaml:"prefix,omitempty" json:"prefix,omitempty"`
	PrefixDevelopment string `yaml:"prefix-development,omitempty" json:"prefix_development,omitempty"`
	Include           string `yaml:"include,omitempty" json:"include,omitempty"`
}

// DependabotGroup defines dependency groups for batch updates
type DependabotGroup struct {
	DependencyType string                   `yaml:"dependency-type,omitempty" json:"dependency_type,omitempty"`
	UpdateTypes    []string                 `yaml:"update-types,omitempty" json:"update_types,omitempty"`
	Patterns       []string                 `yaml:"patterns,omitempty" json:"patterns,omitempty"`
	ExcludePattern []string                 `yaml:"exclude-patterns,omitempty" json:"exclude_patterns,omitempty"`
	AppliesTo      DependabotGroupAppliesTo `yaml:"applies-to,omitempty" json:"applies_to,omitempty"`
}

// DependabotGroupAppliesTo defines version update constraints for groups
type DependabotGroupAppliesTo struct {
	VersionUpdates  []string `yaml:"version-updates,omitempty" json:"version_updates,omitempty"`
	SecurityUpdates bool     `yaml:"security-updates,omitempty" json:"security_updates,omitempty"`
}

// DependabotRegistry defines private package registry configuration
type DependabotRegistry struct {
	Type        string `yaml:"type" json:"type"`
	URL         string `yaml:"url" json:"url"`
	Username    string `yaml:"username,omitempty" json:"username,omitempty"`
	Password    string `yaml:"password,omitempty" json:"password,omitempty"`
	Key         string `yaml:"key,omitempty" json:"key,omitempty"`
	Token       string `yaml:"token,omitempty" json:"token,omitempty"`
	ReplaceBase bool   `yaml:"replace-base,omitempty" json:"replace_base,omitempty"`
}

// DependabotPolicyConfig represents organization-wide Dependabot policies
type DependabotPolicyConfig struct {
	ID                   string                     `json:"id"`
	Name                 string                     `json:"name"`
	Organization         string                     `json:"organization"`
	Description          string                     `json:"description"`
	Enabled              bool                       `json:"enabled"`
	DefaultConfig        DependabotConfig           `json:"default_config"`
	EcosystemPolicies    map[string]EcosystemPolicy `json:"ecosystem_policies"`
	SecurityPolicies     SecurityPolicySettings     `json:"security_policies"`
	ApprovalRequirements ApprovalRequirements       `json:"approval_requirements"`
	CreatedAt            time.Time                  `json:"created_at"`
	UpdatedAt            time.Time                  `json:"updated_at"`
	Version              int                        `json:"version"`
}

// EcosystemPolicy defines policies for specific package ecosystems
type EcosystemPolicy struct {
	Ecosystem             string   `json:"ecosystem"`
	Enabled               bool     `json:"enabled"`
	RequiredReviewers     int      `json:"required_reviewers"`
	AllowedUpdateTypes    []string `json:"allowed_update_types"`
	BlockedDependencies   []string `json:"blocked_dependencies"`
	MaxPullRequestsPerDay int      `json:"max_pull_requests_per_day"`
	AutoMergeEnabled      bool     `json:"auto_merge_enabled"`
	AutoMergeUpdateTypes  []string `json:"auto_merge_update_types"`
	RequiredStatusChecks  []string `json:"required_status_checks"`
	MinSecuritySeverity   string   `json:"min_security_severity"`
}

// SecurityPolicySettings defines security-related policies for Dependabot
type SecurityPolicySettings struct {
	EnableVulnerabilityAlerts  bool     `json:"enable_vulnerability_alerts"`
	AutoFixSecurityVulns       bool     `json:"auto_fix_security_vulns"`
	AllowedSecurityUpdateTypes []string `json:"allowed_security_update_types"`
	SecurityReviewRequired     bool     `json:"security_review_required"`
	CriticalVulnAutoMerge      bool     `json:"critical_vuln_auto_merge"`
	VulnReportingWebhook       string   `json:"vuln_reporting_webhook,omitempty"`
	ExcludedVulnerabilityIDs   []string `json:"excluded_vulnerability_ids,omitempty"`
}

// ApprovalRequirements defines approval requirements for different update types
type ApprovalRequirements struct {
	MajorUpdates    ApprovalRule `json:"major_updates"`
	MinorUpdates    ApprovalRule `json:"minor_updates"`
	PatchUpdates    ApprovalRule `json:"patch_updates"`
	SecurityUpdates ApprovalRule `json:"security_updates"`
}

// ApprovalRule defines approval requirements for a specific update type
type ApprovalRule struct {
	RequiredReviewers      int      `json:"required_reviewers"`
	RequiredApprovals      int      `json:"required_approvals"`
	DismissStaleReviews    bool     `json:"dismiss_stale_reviews"`
	RequireCodeOwnerReview bool     `json:"require_code_owner_review"`
	AllowedMergeUsers      []string `json:"allowed_merge_users,omitempty"`
	RestrictedPaths        []string `json:"restricted_paths,omitempty"`
}

// DependabotStatus represents the current status of Dependabot for a repository
type DependabotStatus struct {
	Repository          string                  `json:"repository"`
	Organization        string                  `json:"organization"`
	Enabled             bool                    `json:"enabled"`
	ConfigExists        bool                    `json:"config_exists"`
	ConfigValid         bool                    `json:"config_valid"`
	LastUpdated         time.Time               `json:"last_updated"`
	ActivePullRequests  int                     `json:"active_pull_requests"`
	RecentUpdates       []DependabotUpdate      `json:"recent_updates"`
	Errors              []DependabotError       `json:"errors,omitempty"`
	SupportedEcosystems []string                `json:"supported_ecosystems"`
	ConfigSummary       DependabotConfigSummary `json:"config_summary"`
}

// DependabotUpdate represents a Dependabot update activity
type DependabotUpdate struct {
	ID               string                 `json:"id"`
	Dependency       string                 `json:"dependency"`
	FromVersion      string                 `json:"from_version"`
	ToVersion        string                 `json:"to_version"`
	UpdateType       string                 `json:"update_type"`
	Ecosystem        string                 `json:"ecosystem"`
	PullRequestURL   string                 `json:"pull_request_url,omitempty"`
	Status           DependabotUpdateStatus `json:"status"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
	SecurityAdvisory *SecurityAdvisoryInfo  `json:"security_advisory,omitempty"`
}

// DependabotError represents an error encountered by Dependabot
type DependabotError struct {
	ID        string              `json:"id"`
	Type      DependabotErrorType `json:"type"`
	Message   string              `json:"message"`
	Ecosystem string              `json:"ecosystem,omitempty"`
	Directory string              `json:"directory,omitempty"`
	Timestamp time.Time           `json:"timestamp"`
	Resolved  bool                `json:"resolved"`
}

// DependabotConfigSummary provides a summary of the current configuration
type DependabotConfigSummary struct {
	TotalEcosystems        int               `json:"total_ecosystems"`
	EnabledEcosystems      []string          `json:"enabled_ecosystems"`
	UpdateSchedules        map[string]string `json:"update_schedules"`
	TotalIgnoredDeps       int               `json:"total_ignored_deps"`
	GroupedUpdatesCount    int               `json:"grouped_updates_count"`
	SecurityUpdatesEnabled bool              `json:"security_updates_enabled"`
	RegistriesConfigured   int               `json:"registries_configured"`
}

// SecurityAdvisoryInfo represents security vulnerability information
type SecurityAdvisoryInfo struct {
	ID          string    `json:"id"`
	Summary     string    `json:"summary"`
	Severity    string    `json:"severity"`
	CVSS        float64   `json:"cvss_score"`
	References  []string  `json:"references"`
	PublishedAt time.Time `json:"published_at"`
}

// Enum types
type DependabotUpdateStatus string

const (
	DependabotUpdateStatusPending    DependabotUpdateStatus = "pending"
	DependabotUpdateStatusActive     DependabotUpdateStatus = "active"
	DependabotUpdateStatusMerged     DependabotUpdateStatus = "merged"
	DependabotUpdateStatusClosed     DependabotUpdateStatus = "closed"
	DependabotUpdateStatusSuperseded DependabotUpdateStatus = "superseded"
	DependabotUpdateStatusFailed     DependabotUpdateStatus = "failed"
)

type DependabotErrorType string

const (
	DependabotErrorTypeConfigInvalid     DependabotErrorType = "config_invalid"
	DependabotErrorTypeEcosystemNotFound DependabotErrorType = "ecosystem_not_found"
	DependabotErrorTypeRegistryAuth      DependabotErrorType = "registry_auth_failed"
	DependabotErrorTypePermissions       DependabotErrorType = "insufficient_permissions"
	DependabotErrorTypeRateLimit         DependabotErrorType = "rate_limit_exceeded"
	DependabotErrorTypeUnknown           DependabotErrorType = "unknown_error"
)

// Supported package ecosystems
const (
	EcosystemNPM           = "npm"
	EcosystemPip           = "pip"
	EcosystemBundler       = "bundler"
	EcosystemGradle        = "gradle"
	EcosystemMaven         = "maven"
	EcosystemComposer      = "composer"
	EcosystemNuGet         = "nuget"
	EcosystemCargoRust     = "cargo"
	EcosystemGoModules     = "gomod"
	EcosystemDockerfile    = "docker"
	EcosystemGitSubmodule  = "gitsubmodule"
	EcosystemGitHubActions = "github-actions"
	EcosystemTerraform     = "terraform"
	EcosystemElm           = "elm"
	EcosystemMix           = "mix"
	EcosystemPub           = "pub"
	EcosystemSwift         = "swift"
)

// Update intervals
const (
	IntervalDaily   = "daily"
	IntervalWeekly  = "weekly"
	IntervalMonthly = "monthly"
)

// Update types
const (
	UpdateTypeAll           = "all"
	UpdateTypeSecurity      = "security"
	UpdateTypeVersionUpdate = "version-update:semver-major"
	UpdateTypeVersionMinor  = "version-update:semver-minor"
	UpdateTypeVersionPatch  = "version-update:semver-patch"
)

// Versioning strategies
const (
	VersioningStrategyAuto                = "auto"
	VersioningStrategyLockfileOnly        = "lockfile-only"
	VersioningStrategyWiden               = "widen"
	VersioningStrategyIncrease            = "increase"
	VersioningStrategyIncreaseIfNecessary = "increase-if-necessary"
)

// NewDependabotConfigManager creates a new Dependabot configuration manager
func NewDependabotConfigManager(logger Logger, apiClient APIClient) *DependabotConfigManager {
	return &DependabotConfigManager{
		logger:    logger,
		apiClient: apiClient,
	}
}

// GetDependabotConfig retrieves the current Dependabot configuration for a repository
func (dm *DependabotConfigManager) GetDependabotConfig(ctx context.Context, organization, repository string) (*DependabotConfig, error) {
	dm.logger.Info("Retrieving Dependabot configuration", "organization", organization, "repository", repository)

	// In a real implementation, this would fetch from GitHub API
	// For now, return a mock configuration
	config := &DependabotConfig{
		Version: 2,
		Updates: []DependabotUpdateRule{
			{
				PackageEcosystem: EcosystemGoModules,
				Directory:        "/",
				Schedule: DependabotSchedule{
					Interval: IntervalWeekly,
					Day:      "monday",
					Time:     "06:00",
					Timezone: "UTC",
				},
				PullRequestLimit: 5,
				Labels:           []string{"dependencies", "go"},
				CommitMessage: &DependabotCommitMessage{
					Prefix:  "deps",
					Include: "scope",
				},
			},
		},
	}

	return config, nil
}

// UpdateDependabotConfig updates the Dependabot configuration for a repository
func (dm *DependabotConfigManager) UpdateDependabotConfig(ctx context.Context, organization, repository string, config *DependabotConfig) error {
	dm.logger.Info("Updating Dependabot configuration", "organization", organization, "repository", repository)

	// Validate configuration
	if err := dm.ValidateConfig(config); err != nil {
		return fmt.Errorf("invalid Dependabot configuration: %w", err)
	}

	// Convert to YAML
	configYAML, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration to YAML: %w", err)
	}

	dm.logger.Debug("Generated Dependabot configuration", "yaml", string(configYAML))

	// In a real implementation, this would update the .github/dependabot.yml file via GitHub API
	// For now, log the operation
	dm.logger.Info("Dependabot configuration updated successfully",
		"organization", organization,
		"repository", repository,
		"ecosystems", len(config.Updates))

	return nil
}

// CreateDefaultConfig creates a default Dependabot configuration for a repository
func (dm *DependabotConfigManager) CreateDefaultConfig(ctx context.Context, organization, repository string, ecosystems []string) (*DependabotConfig, error) {
	dm.logger.Info("Creating default Dependabot configuration",
		"organization", organization,
		"repository", repository,
		"ecosystems", ecosystems)

	config := &DependabotConfig{
		Version: 2,
		Updates: make([]DependabotUpdateRule, 0),
	}

	// Create update rules for each ecosystem
	for _, ecosystem := range ecosystems {
		rule := dm.createDefaultUpdateRule(ecosystem)
		config.Updates = append(config.Updates, rule)
	}

	return config, nil
}

// ValidateConfig validates a Dependabot configuration
func (dm *DependabotConfigManager) ValidateConfig(config *DependabotConfig) error {
	if config.Version != 2 {
		return fmt.Errorf("unsupported version: %d (only version 2 is supported)", config.Version)
	}

	if len(config.Updates) == 0 {
		return fmt.Errorf("at least one update rule is required")
	}

	// Validate each update rule
	for i, update := range config.Updates {
		if err := dm.validateUpdateRule(&update); err != nil {
			return fmt.Errorf("invalid update rule %d: %w", i, err)
		}
	}

	return nil
}

// GetDependabotStatus retrieves the current status of Dependabot for a repository
func (dm *DependabotConfigManager) GetDependabotStatus(ctx context.Context, organization, repository string) (*DependabotStatus, error) {
	dm.logger.Info("Retrieving Dependabot status", "organization", organization, "repository", repository)

	// In a real implementation, this would query GitHub API for Dependabot status
	status := &DependabotStatus{
		Repository:          repository,
		Organization:        organization,
		Enabled:             true,
		ConfigExists:        true,
		ConfigValid:         true,
		LastUpdated:         time.Now().Add(-24 * time.Hour),
		ActivePullRequests:  2,
		SupportedEcosystems: []string{EcosystemGoModules, EcosystemDockerfile, EcosystemGitHubActions},
		RecentUpdates: []DependabotUpdate{
			{
				ID:          "update-1",
				Dependency:  "github.com/stretchr/testify",
				FromVersion: "v1.8.0",
				ToVersion:   "v1.8.4",
				UpdateType:  UpdateTypeVersionPatch,
				Ecosystem:   EcosystemGoModules,
				Status:      DependabotUpdateStatusMerged,
				CreatedAt:   time.Now().Add(-48 * time.Hour),
				UpdatedAt:   time.Now().Add(-24 * time.Hour),
			},
		},
		ConfigSummary: DependabotConfigSummary{
			TotalEcosystems:        3,
			EnabledEcosystems:      []string{EcosystemGoModules, EcosystemDockerfile, EcosystemGitHubActions},
			UpdateSchedules:        map[string]string{EcosystemGoModules: IntervalWeekly},
			SecurityUpdatesEnabled: true,
			TotalIgnoredDeps:       0,
			GroupedUpdatesCount:    0,
			RegistriesConfigured:   0,
		},
	}

	return status, nil
}

// Helper methods

func (dm *DependabotConfigManager) createDefaultUpdateRule(ecosystem string) DependabotUpdateRule {
	rule := DependabotUpdateRule{
		PackageEcosystem: ecosystem,
		Directory:        "/",
		Schedule: DependabotSchedule{
			Interval: IntervalWeekly,
			Day:      "monday",
			Time:     "06:00",
			Timezone: "UTC",
		},
		PullRequestLimit: 5,
		Labels:           []string{"dependencies", ecosystem},
		RebaseStrategy:   "auto",
	}

	// Ecosystem-specific defaults
	switch ecosystem {
	case EcosystemGoModules:
		rule.VendorUpdates = true
		rule.CommitMessage = &DependabotCommitMessage{
			Prefix:  "deps",
			Include: "scope",
		}
	case EcosystemNPM:
		rule.VersioningStrategy = VersioningStrategyIncrease
		rule.AllowedUpdates = []DependabotAllowedUpdate{
			{DependencyType: "direct"},
			{DependencyType: "indirect", UpdateType: UpdateTypeSecurity},
		}
	case EcosystemDockerfile:
		rule.Schedule.Interval = IntervalMonthly
		rule.PullRequestLimit = 3
	case EcosystemGitHubActions:
		rule.Schedule.Interval = IntervalWeekly
		rule.Groups = map[string]DependabotGroup{
			"github-actions": {
				Patterns: []string{"actions/*"},
			},
		}
	}

	return rule
}

func (dm *DependabotConfigManager) validateUpdateRule(rule *DependabotUpdateRule) error {
	// Validate package ecosystem
	if !dm.isSupportedEcosystem(rule.PackageEcosystem) {
		return fmt.Errorf("unsupported package ecosystem: %s", rule.PackageEcosystem)
	}

	// Validate directory
	if rule.Directory == "" {
		return fmt.Errorf("directory is required")
	}

	// Validate schedule
	if !dm.isValidInterval(rule.Schedule.Interval) {
		return fmt.Errorf("invalid schedule interval: %s", rule.Schedule.Interval)
	}

	// Validate versioning strategy
	if rule.VersioningStrategy != "" && !dm.isValidVersioningStrategy(rule.VersioningStrategy) {
		return fmt.Errorf("invalid versioning strategy: %s", rule.VersioningStrategy)
	}

	return nil
}

func (dm *DependabotConfigManager) isSupportedEcosystem(ecosystem string) bool {
	supportedEcosystems := []string{
		EcosystemNPM, EcosystemPip, EcosystemBundler, EcosystemGradle,
		EcosystemMaven, EcosystemComposer, EcosystemNuGet, EcosystemCargoRust,
		EcosystemGoModules, EcosystemDockerfile, EcosystemGitSubmodule,
		EcosystemGitHubActions, EcosystemTerraform, EcosystemElm,
		EcosystemMix, EcosystemPub, EcosystemSwift,
	}

	for _, supported := range supportedEcosystems {
		if ecosystem == supported {
			return true
		}
	}
	return false
}

func (dm *DependabotConfigManager) isValidInterval(interval string) bool {
	validIntervals := []string{IntervalDaily, IntervalWeekly, IntervalMonthly}
	for _, valid := range validIntervals {
		if interval == valid {
			return true
		}
	}
	return false
}

func (dm *DependabotConfigManager) isValidVersioningStrategy(strategy string) bool {
	validStrategies := []string{
		VersioningStrategyAuto, VersioningStrategyLockfileOnly,
		VersioningStrategyWiden, VersioningStrategyIncrease,
		VersioningStrategyIncreaseIfNecessary,
	}
	for _, valid := range validStrategies {
		if strategy == valid {
			return true
		}
	}
	return false
}

// DetectEcosystems detects package ecosystems in a repository
func (dm *DependabotConfigManager) DetectEcosystems(ctx context.Context, organization, repository string) ([]string, error) {
	dm.logger.Info("Detecting package ecosystems", "organization", organization, "repository", repository)

	// In a real implementation, this would analyze repository files
	// For now, return mock detected ecosystems
	ecosystems := []string{EcosystemGoModules}

	// Mock detection logic based on repository name patterns
	repoName := strings.ToLower(repository)
	if strings.Contains(repoName, "node") || strings.Contains(repoName, "js") {
		ecosystems = append(ecosystems, EcosystemNPM)
	}
	if strings.Contains(repoName, "python") || strings.Contains(repoName, "py") {
		ecosystems = append(ecosystems, EcosystemPip)
	}
	if strings.Contains(repoName, "docker") {
		ecosystems = append(ecosystems, EcosystemDockerfile)
	}

	// Always include GitHub Actions for repositories with workflows
	ecosystems = append(ecosystems, EcosystemGitHubActions)

	dm.logger.Info("Detected ecosystems", "ecosystems", ecosystems)
	return ecosystems, nil
}
