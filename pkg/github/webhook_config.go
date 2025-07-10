package github

import (
	"context"
	"time"
)

// WebhookPolicy represents an organization-wide webhook policy
type WebhookPolicy struct {
	ID           string              `json:"id" yaml:"id"`
	Name         string              `json:"name" yaml:"name"`
	Description  string              `json:"description" yaml:"description"`
	Organization string              `json:"organization" yaml:"organization"`
	Enabled      bool                `json:"enabled" yaml:"enabled"`
	Priority     int                 `json:"priority" yaml:"priority"` // Higher number = higher priority
	Rules        []WebhookPolicyRule `json:"rules" yaml:"rules"`
	CreatedAt    time.Time           `json:"created_at" yaml:"created_at"`
	UpdatedAt    time.Time           `json:"updated_at" yaml:"updated_at"`
	CreatedBy    string              `json:"created_by" yaml:"created_by"`
	Tags         map[string]string   `json:"tags,omitempty" yaml:"tags,omitempty"`
}

// WebhookPolicyRule defines a rule for applying webhooks
type WebhookPolicyRule struct {
	ID         string             `json:"id" yaml:"id"`
	Name       string             `json:"name" yaml:"name"`
	Conditions WebhookConditions  `json:"conditions" yaml:"conditions"`
	Action     WebhookAction      `json:"action" yaml:"action"`
	Template   WebhookTemplate    `json:"template" yaml:"template"`
	Enabled    bool               `json:"enabled" yaml:"enabled"`
	OnConflict ConflictResolution `json:"on_conflict" yaml:"on_conflict"`
}

// WebhookConditions defines when a rule should be applied
type WebhookConditions struct {
	RepositoryName    []string          `json:"repository_name,omitempty" yaml:"repository_name,omitempty"`
	RepositoryPattern []string          `json:"repository_pattern,omitempty" yaml:"repository_pattern,omitempty"`
	Language          []string          `json:"language,omitempty" yaml:"language,omitempty"`
	Topics            []string          `json:"topics,omitempty" yaml:"topics,omitempty"`
	Visibility        []string          `json:"visibility,omitempty" yaml:"visibility,omitempty"` // public, private, internal
	IsArchived        *bool             `json:"is_archived,omitempty" yaml:"is_archived,omitempty"`
	IsTemplate        *bool             `json:"is_template,omitempty" yaml:"is_template,omitempty"`
	HasIssues         *bool             `json:"has_issues,omitempty" yaml:"has_issues,omitempty"`
	CustomFields      map[string]string `json:"custom_fields,omitempty" yaml:"custom_fields,omitempty"`
}

// WebhookAction defines what action to take
type WebhookAction string

const (
	WebhookActionCreate WebhookAction = "create"
	WebhookActionUpdate WebhookAction = "update"
	WebhookActionDelete WebhookAction = "delete"
	WebhookActionEnsure WebhookAction = "ensure" // create if not exists, update if exists
)

// ConflictResolution defines how to handle conflicts
type ConflictResolution string

const (
	ConflictResolutionSkip      ConflictResolution = "skip"      // Skip if webhook exists
	ConflictResolutionOverwrite ConflictResolution = "overwrite" // Overwrite existing webhook
	ConflictResolutionMerge     ConflictResolution = "merge"     // Merge configurations
	ConflictResolutionError     ConflictResolution = "error"     // Fail on conflict
)

// WebhookTemplate defines the webhook configuration template
type WebhookTemplate struct {
	Name      string                `json:"name" yaml:"name"`
	URL       string                `json:"url" yaml:"url"`
	Events    []string              `json:"events" yaml:"events"`
	Active    bool                  `json:"active" yaml:"active"`
	Config    WebhookConfigTemplate `json:"config" yaml:"config"`
	Variables map[string]string     `json:"variables,omitempty" yaml:"variables,omitempty"` // Template variables
}

// WebhookConfigTemplate extends WebhookConfig with template support
type WebhookConfigTemplate struct {
	URL         string `json:"url" yaml:"url"`
	ContentType string `json:"content_type" yaml:"content_type"`
	Secret      string `json:"secret,omitempty" yaml:"secret,omitempty"`
	InsecureSSL bool   `json:"insecure_ssl" yaml:"insecure_ssl"`
}

// OrganizationWebhookConfig represents the overall webhook configuration for an organization
type OrganizationWebhookConfig struct {
	Organization string                      `json:"organization" yaml:"organization"`
	Version      string                      `json:"version" yaml:"version"`
	Metadata     ConfigMetadata              `json:"metadata" yaml:"metadata"`
	Defaults     WebhookDefaults             `json:"defaults" yaml:"defaults"`
	Policies     []WebhookPolicy             `json:"policies" yaml:"policies"`
	Settings     OrganizationWebhookSettings `json:"settings" yaml:"settings"`
	Validation   ValidationConfig            `json:"validation" yaml:"validation"`
}

// ConfigMetadata contains metadata about the configuration
type ConfigMetadata struct {
	Name        string            `json:"name" yaml:"name"`
	Description string            `json:"description" yaml:"description"`
	Author      string            `json:"author" yaml:"author"`
	CreatedAt   time.Time         `json:"created_at" yaml:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at" yaml:"updated_at"`
	Version     string            `json:"version" yaml:"version"`
	Tags        map[string]string `json:"tags,omitempty" yaml:"tags,omitempty"`
}

// WebhookDefaults defines default webhook settings
type WebhookDefaults struct {
	Events    []string              `json:"events" yaml:"events"`
	Active    bool                  `json:"active" yaml:"active"`
	Config    WebhookConfigTemplate `json:"config" yaml:"config"`
	Variables map[string]string     `json:"variables,omitempty" yaml:"variables,omitempty"`
}

// OrganizationWebhookSettings contains organization-specific settings
type OrganizationWebhookSettings struct {
	AllowRepositoryOverride bool                 `json:"allow_repository_override" yaml:"allow_repository_override"`
	RequireApproval         bool                 `json:"require_approval" yaml:"require_approval"`
	MaxWebhooksPerRepo      int                  `json:"max_webhooks_per_repo" yaml:"max_webhooks_per_repo"`
	RetryOnFailure          bool                 `json:"retry_on_failure" yaml:"retry_on_failure"`
	NotificationSettings    NotificationSettings `json:"notification_settings" yaml:"notification_settings"`
}

// NotificationSettings defines notification preferences
type NotificationSettings struct {
	OnSuccess    bool     `json:"on_success" yaml:"on_success"`
	OnFailure    bool     `json:"on_failure" yaml:"on_failure"`
	OnConflict   bool     `json:"on_conflict" yaml:"on_conflict"`
	Recipients   []string `json:"recipients" yaml:"recipients"`
	SlackChannel string   `json:"slack_channel,omitempty" yaml:"slack_channel,omitempty"`
}

// ValidationConfig defines validation rules
type ValidationConfig struct {
	RequiredEvents   []string `json:"required_events,omitempty" yaml:"required_events,omitempty"`
	ForbiddenEvents  []string `json:"forbidden_events,omitempty" yaml:"forbidden_events,omitempty"`
	AllowedDomains   []string `json:"allowed_domains,omitempty" yaml:"allowed_domains,omitempty"`
	ForbiddenDomains []string `json:"forbidden_domains,omitempty" yaml:"forbidden_domains,omitempty"`
	RequireSSL       bool     `json:"require_ssl" yaml:"require_ssl"`
	RequireSecret    bool     `json:"require_secret" yaml:"require_secret"`
}

// WebhookConfigurationService provides organization-wide webhook configuration management
type WebhookConfigurationService interface {
	// Policy Management
	CreatePolicy(ctx context.Context, policy *WebhookPolicy) error
	GetPolicy(ctx context.Context, org, policyID string) (*WebhookPolicy, error)
	ListPolicies(ctx context.Context, org string) ([]*WebhookPolicy, error)
	UpdatePolicy(ctx context.Context, policy *WebhookPolicy) error
	DeletePolicy(ctx context.Context, org, policyID string) error

	// Configuration Management
	GetOrganizationConfig(ctx context.Context, org string) (*OrganizationWebhookConfig, error)
	UpdateOrganizationConfig(ctx context.Context, config *OrganizationWebhookConfig) error
	ValidateConfiguration(ctx context.Context, config *OrganizationWebhookConfig) (*WebhookValidationResult, error)

	// Policy Application
	ApplyPolicies(ctx context.Context, request *ApplyPoliciesRequest) (*ApplyPoliciesResult, error)
	PreviewPolicyApplication(ctx context.Context, request *ApplyPoliciesRequest) (*PolicyApplicationPreview, error)

	// Migration and Sync
	MigrateExistingWebhooks(ctx context.Context, request *MigrationRequest) (*MigrationResult, error)
	SyncOrganizationWebhooks(ctx context.Context, org string) (*SyncResult, error)

	// Reporting and Audit
	GenerateComplianceReport(ctx context.Context, org string) (*ComplianceReport, error)
	GetWebhookInventory(ctx context.Context, org string) (*WebhookInventory, error)
}

// ApplyPoliciesRequest represents a request to apply webhook policies
type ApplyPoliciesRequest struct {
	Organization    string   `json:"organization"`
	PolicyIDs       []string `json:"policy_ids,omitempty"`       // if empty, apply all enabled policies
	RepositoryNames []string `json:"repository_names,omitempty"` // if empty, apply to all repos
	DryRun          bool     `json:"dry_run"`
	Force           bool     `json:"force"` // Override conflict resolution
}

// ApplyPoliciesResult represents the result of applying policies
type ApplyPoliciesResult struct {
	Organization          string                    `json:"organization"`
	TotalRepositories     int                       `json:"total_repositories"`
	ProcessedRepositories int                       `json:"processed_repositories"`
	SuccessCount          int                       `json:"success_count"`
	FailureCount          int                       `json:"failure_count"`
	SkippedCount          int                       `json:"skipped_count"`
	Results               []PolicyApplicationResult `json:"results"`
	ExecutionTime         string                    `json:"execution_time"`
	Summary               PolicyApplicationSummary  `json:"summary"`
}

// PolicyApplicationResult represents the result for a single repository
type PolicyApplicationResult struct {
	Repository string        `json:"repository"`
	PolicyID   string        `json:"policy_id"`
	RuleID     string        `json:"rule_id"`
	Action     WebhookAction `json:"action"`
	Success    bool          `json:"success"`
	WebhookID  *int64        `json:"webhook_id,omitempty"`
	Error      string        `json:"error,omitempty"`
	Skipped    bool          `json:"skipped"`
	SkipReason string        `json:"skip_reason,omitempty"`
	Changes    []string      `json:"changes,omitempty"`
	Duration   string        `json:"duration"`
}

// PolicyApplicationSummary provides a summary of policy application
type PolicyApplicationSummary struct {
	WebhooksCreated int            `json:"webhooks_created"`
	WebhooksUpdated int            `json:"webhooks_updated"`
	WebhooksDeleted int            `json:"webhooks_deleted"`
	ConflictsFound  int            `json:"conflicts_found"`
	ErrorsByType    map[string]int `json:"errors_by_type"`
}

// PolicyApplicationPreview shows what would happen without making changes
type PolicyApplicationPreview struct {
	Organization      string                   `json:"organization"`
	TotalRepositories int                      `json:"total_repositories"`
	PlannedActions    []PlannedAction          `json:"planned_actions"`
	Conflicts         []PolicyConflict         `json:"conflicts"`
	Warnings          []string                 `json:"warnings"`
	Summary           PolicyApplicationSummary `json:"summary"`
}

// PlannedAction represents an action that would be taken
type PlannedAction struct {
	Repository  string        `json:"repository"`
	PolicyID    string        `json:"policy_id"`
	RuleID      string        `json:"rule_id"`
	Action      WebhookAction `json:"action"`
	WebhookName string        `json:"webhook_name"`
	Changes     []string      `json:"changes"`
	Conflicts   []string      `json:"conflicts,omitempty"`
}

// PolicyConflict represents a conflict between policies or existing webhooks
type PolicyConflict struct {
	Repository      string       `json:"repository"`
	ConflictType    string       `json:"conflict_type"`
	Description     string       `json:"description"`
	PolicyID1       string       `json:"policy_id_1"`
	PolicyID2       string       `json:"policy_id_2,omitempty"`
	ExistingWebhook *WebhookInfo `json:"existing_webhook,omitempty"`
	Resolution      string       `json:"resolution"`
}

// MigrationRequest represents a request to migrate existing webhooks
type MigrationRequest struct {
	Organization   string            `json:"organization"`
	SourceConfig   string            `json:"source_config,omitempty"` // Path to source configuration
	TargetPolicyID string            `json:"target_policy_id"`
	DryRun         bool              `json:"dry_run"`
	BackupExisting bool              `json:"backup_existing"`
	Mapping        map[string]string `json:"mapping,omitempty"` // URL mappings for migration
}

// MigrationResult represents the result of a migration operation
type MigrationResult struct {
	Organization     string                   `json:"organization"`
	TotalWebhooks    int                      `json:"total_webhooks"`
	MigratedWebhooks int                      `json:"migrated_webhooks"`
	SkippedWebhooks  int                      `json:"skipped_webhooks"`
	FailedWebhooks   int                      `json:"failed_webhooks"`
	BackupPath       string                   `json:"backup_path,omitempty"`
	Results          []WebhookMigrationResult `json:"results"`
	ExecutionTime    string                   `json:"execution_time"`
}

// WebhookMigrationResult represents the result for a single webhook migration
type WebhookMigrationResult struct {
	Repository   string   `json:"repository"`
	OldWebhookID int64    `json:"old_webhook_id"`
	NewWebhookID int64    `json:"new_webhook_id,omitempty"`
	Success      bool     `json:"success"`
	Error        string   `json:"error,omitempty"`
	Changes      []string `json:"changes"`
}

// SyncResult represents the result of synchronizing webhooks
type SyncResult struct {
	Organization       string               `json:"organization"`
	TotalRepositories  int                  `json:"total_repositories"`
	SyncedRepositories int                  `json:"synced_repositories"`
	Discrepancies      []WebhookDiscrepancy `json:"discrepancies"`
	ExecutionTime      string               `json:"execution_time"`
}

// WebhookDiscrepancy represents a difference between expected and actual webhook configuration
type WebhookDiscrepancy struct {
	Repository      string `json:"repository"`
	WebhookID       int64  `json:"webhook_id"`
	DiscrepancyType string `json:"discrepancy_type"`
	Expected        string `json:"expected"`
	Actual          string `json:"actual"`
	Severity        string `json:"severity"`
}

// ComplianceReport represents a compliance report for webhooks
type ComplianceReport struct {
	Organization      string                `json:"organization"`
	GeneratedAt       time.Time             `json:"generated_at"`
	TotalRepositories int                   `json:"total_repositories"`
	CompliantRepos    int                   `json:"compliant_repos"`
	NonCompliantRepos int                   `json:"non_compliant_repos"`
	Violations        []ComplianceViolation `json:"violations"`
	ComplianceScore   float64               `json:"compliance_score"`
	Recommendations   []string              `json:"recommendations"`
}

// ComplianceViolation represents a compliance violation
type ComplianceViolation struct {
	Repository    string `json:"repository"`
	PolicyID      string `json:"policy_id"`
	RuleID        string `json:"rule_id"`
	ViolationType string `json:"violation_type"`
	Description   string `json:"description"`
	Severity      string `json:"severity"`
	Remediation   string `json:"remediation"`
}

// WebhookInventory represents an inventory of all webhooks in an organization
type WebhookInventory struct {
	Organization    string                  `json:"organization"`
	GeneratedAt     time.Time               `json:"generated_at"`
	TotalWebhooks   int                     `json:"total_webhooks"`
	WebhooksByType  map[string]int          `json:"webhooks_by_type"`
	WebhooksByEvent map[string]int          `json:"webhooks_by_event"`
	Repositories    []RepositoryWebhookInfo `json:"repositories"`
	Summary         WebhookInventorySummary `json:"summary"`
}

// RepositoryWebhookInfo represents webhook information for a repository
type RepositoryWebhookInfo struct {
	Repository string         `json:"repository"`
	Webhooks   []*WebhookInfo `json:"webhooks"`
	Compliance string         `json:"compliance"` // compliant, non-compliant, unknown
	Issues     []string       `json:"issues,omitempty"`
}

// WebhookInventorySummary provides summary statistics
type WebhookInventorySummary struct {
	ActiveWebhooks    int     `json:"active_webhooks"`
	InactiveWebhooks  int     `json:"inactive_webhooks"`
	DuplicateWebhooks int     `json:"duplicate_webhooks"`
	OrphanedWebhooks  int     `json:"orphaned_webhooks"`
	HealthScore       float64 `json:"health_score"`
}

// WebhookValidationResult represents the result of webhook configuration validation
type WebhookValidationResult struct {
	Valid    bool                       `json:"valid"`
	Errors   []WebhookValidationError   `json:"errors,omitempty"`
	Warnings []WebhookValidationWarning `json:"warnings,omitempty"`
	Score    int                        `json:"score"` // 0-100
}

// WebhookValidationError represents a webhook validation error
type WebhookValidationError struct {
	Field      string `json:"field"`
	Message    string `json:"message"`
	Severity   string `json:"severity"`
	Suggestion string `json:"suggestion,omitempty"`
}

// WebhookValidationWarning represents a webhook validation warning
type WebhookValidationWarning struct {
	Field      string `json:"field"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion,omitempty"`
}
