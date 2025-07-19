package github

import (
	"context"
	"fmt"
	"time"
)

// ActionsPermissionLevel defines the permission level for GitHub Actions.
type ActionsPermissionLevel string

const (
	ActionsPermissionDisabled        ActionsPermissionLevel = "disabled"
	ActionsPermissionAll             ActionsPermissionLevel = "all"
	ActionsPermissionLocalOnly       ActionsPermissionLevel = "local_only"
	ActionsPermissionSelectedActions ActionsPermissionLevel = "selected"
)

// ActionsPolicy represents a GitHub Actions permission policy.
type ActionsPolicy struct {
	ID                     string                  `json:"id" yaml:"id"`
	Name                   string                  `json:"name" yaml:"name"`
	Description            string                  `json:"description" yaml:"description"`
	Organization           string                  `json:"organization" yaml:"organization"`
	Repository             string                  `json:"repository,omitempty" yaml:"repository,omitempty"`
	PermissionLevel        ActionsPermissionLevel  `json:"permissionLevel" yaml:"permission_level"`
	AllowedActions         []string                `json:"allowedActions,omitempty" yaml:"allowed_actions,omitempty"`
	AllowedActionsPatterns []string                `json:"allowedActionsPatterns,omitempty" yaml:"allowed_actions_patterns,omitempty"`
	WorkflowPermissions    WorkflowPermissions     `json:"workflowPermissions" yaml:"workflow_permissions"`
	SecuritySettings       ActionsSecuritySettings `json:"securitySettings" yaml:"security_settings"`
	SecretsPolicy          SecretsPolicy           `json:"secretsPolicy" yaml:"secrets_policy"`
	Variables              map[string]string       `json:"variables,omitempty" yaml:"variables,omitempty"`
	Environments           []EnvironmentPolicy     `json:"environments,omitempty" yaml:"environments,omitempty"`
	Runners                RunnerPolicy            `json:"runners" yaml:"runners"`
	CreatedAt              time.Time               `json:"createdAt" yaml:"created_at"`
	UpdatedAt              time.Time               `json:"updatedAt" yaml:"updated_at"`
	CreatedBy              string                  `json:"createdBy" yaml:"created_by"`
	UpdatedBy              string                  `json:"updatedBy" yaml:"updated_by"`
	Version                int                     `json:"version" yaml:"version"`
	Enabled                bool                    `json:"enabled" yaml:"enabled"`
	Tags                   []string                `json:"tags,omitempty" yaml:"tags,omitempty"`
}

// WorkflowPermissions defines permissions for workflow tokens.
type WorkflowPermissions struct {
	DefaultPermissions       DefaultPermissions                `json:"defaultPermissions" yaml:"default_permissions"`
	CanApproveOwnChanges     bool                              `json:"canApproveOwnChanges" yaml:"can_approve_own_changes"`
	ActionsReadPermission    ActionsTokenPermission            `json:"actionsRead" yaml:"actions_read"`
	ContentsPermission       ActionsTokenPermission            `json:"contents" yaml:"contents"`
	MetadataPermission       ActionsTokenPermission            `json:"metadata" yaml:"metadata"`
	PackagesPermission       ActionsTokenPermission            `json:"packages" yaml:"packages"`
	PullRequestsPermission   ActionsTokenPermission            `json:"pullRequests" yaml:"pull_requests"`
	IssuesPermission         ActionsTokenPermission            `json:"issues" yaml:"issues"`
	DeploymentsPermission    ActionsTokenPermission            `json:"deployments" yaml:"deployments"`
	ChecksPermission         ActionsTokenPermission            `json:"checks" yaml:"checks"`
	StatusesPermission       ActionsTokenPermission            `json:"statuses" yaml:"statuses"`
	SecurityEventsPermission ActionsTokenPermission            `json:"securityEvents" yaml:"security_events"`
	IdTokenPermission        ActionsTokenPermission            `json:"idToken" yaml:"id_token"`
	AttestationsPermission   ActionsTokenPermission            `json:"attestations" yaml:"attestations"`
	CustomPermissions        map[string]ActionsTokenPermission `json:"customPermissions,omitempty" yaml:"custom_permissions,omitempty"`
}

// DefaultPermissions defines the default permission level for workflow tokens.
type DefaultPermissions string

const (
	DefaultPermissionsRead       DefaultPermissions = "read"
	DefaultPermissionsWrite      DefaultPermissions = "write"
	DefaultPermissionsRestricted DefaultPermissions = "restricted"
)

// ActionsTokenPermission defines the permission level for a specific scope.
type ActionsTokenPermission string

const (
	TokenPermissionNone  ActionsTokenPermission = "none"
	TokenPermissionRead  ActionsTokenPermission = "read"
	TokenPermissionWrite ActionsTokenPermission = "write"
)

// ActionsSecuritySettings defines security-related settings for Actions.
type ActionsSecuritySettings struct {
	RequireCodeScanningApproval   bool                     `json:"requireCodeScanningApproval" yaml:"require_code_scanning_approval"`
	RequireSecretScanningApproval bool                     `json:"requireSecretScanningApproval" yaml:"require_secret_scanning_approval"`
	AllowForkPRs                  bool                     `json:"allowForkPRs" yaml:"allow_fork_prs"`
	RequireApprovalForForkPRs     bool                     `json:"requireApprovalForForkPRs" yaml:"require_approval_for_fork_prs"`
	AllowPrivateRepoForkRun       bool                     `json:"allowPrivateRepoForkRun" yaml:"allow_private_repo_fork_run"`
	RequireApprovalForPrivateFork bool                     `json:"requireApprovalForPrivateFork" yaml:"require_approval_for_private_fork"`
	RestrictedActionsPatterns     []string                 `json:"restrictedActionsPatterns,omitempty" yaml:"restricted_actions_patterns,omitempty"`
	AllowGitHubOwnedActions       bool                     `json:"allowGitHubOwnedActions" yaml:"allow_github_owned_actions"`
	AllowVerifiedPartnerActions   bool                     `json:"allowVerifiedPartnerActions" yaml:"allow_verified_partner_actions"`
	AllowMarketplaceActions       ActionsMarketplacePolicy `json:"allowMarketplaceActions" yaml:"allow_marketplace_actions"`
	RequireSignedCommits          bool                     `json:"requireSignedCommits" yaml:"require_signed_commits"`
	EnforceAdminsOnBranches       bool                     `json:"enforceAdminsOnBranches" yaml:"enforce_admins_on_branches"`
	OIDCCustomClaims              map[string]string        `json:"oidcCustomClaims,omitempty" yaml:"oidc_custom_claims,omitempty"`
}

// ActionsMarketplacePolicy defines the policy for marketplace actions.
type ActionsMarketplacePolicy string

const (
	MarketplacePolicyDisabled     ActionsMarketplacePolicy = "disabled"
	MarketplacePolicyVerifiedOnly ActionsMarketplacePolicy = "verified_only"
	MarketplacePolicyAll          ActionsMarketplacePolicy = "all"
	MarketplacePolicySelected     ActionsMarketplacePolicy = "selected"
)

// SecretsPolicy defines policy for managing secrets.
type SecretsPolicy struct {
	AllowedSecrets               []string             `json:"allowedSecrets,omitempty" yaml:"allowed_secrets,omitempty"`
	RestrictedSecrets            []string             `json:"restrictedSecrets,omitempty" yaml:"restricted_secrets,omitempty"`
	RequireApprovalForNewSecrets bool                 `json:"requireApprovalForNewSecrets" yaml:"require_approval_for_new_secrets"`
	SecretVisibility             SecretVisibility     `json:"secretVisibility" yaml:"secret_visibility"`
	AllowSecretsInheritance      bool                 `json:"allowSecretsInheritance" yaml:"allow_secrets_inheritance"`
	SecretNamingPatterns         []string             `json:"secretNamingPatterns,omitempty" yaml:"secret_naming_patterns,omitempty"`
	MaxSecretCount               int                  `json:"maxSecretCount,omitempty" yaml:"max_secret_count,omitempty"`
	SecretRotationPolicy         SecretRotationPolicy `json:"secretRotationPolicy" yaml:"secret_rotation_policy"`
}

// SecretVisibility defines the visibility scope for secrets.
type SecretVisibility string

const (
	SecretVisibilityAll           SecretVisibility = "all"
	SecretVisibilityPrivate       SecretVisibility = "private"
	SecretVisibilitySelectedRepos SecretVisibility = "selected"
)

// SecretRotationPolicy defines policy for secret rotation.
type SecretRotationPolicy struct {
	Enabled                bool          `json:"enabled" yaml:"enabled"`
	RotationInterval       time.Duration `json:"rotationInterval" yaml:"rotation_interval"`
	RequireRotationWarning bool          `json:"requireRotationWarning" yaml:"require_rotation_warning"`
	WarningDays            int           `json:"warningDays" yaml:"warning_days"`
	AutoRotateSecrets      []string      `json:"autoRotateSecrets,omitempty" yaml:"auto_rotate_secrets,omitempty"`
}

// EnvironmentPolicy defines policy for deployment environments.
type EnvironmentPolicy struct {
	Name                    string                  `json:"name" yaml:"name"`
	RequiredReviewers       []string                `json:"requiredReviewers,omitempty" yaml:"required_reviewers,omitempty"`
	RequiredReviewerTeams   []string                `json:"requiredReviewerTeams,omitempty" yaml:"required_reviewer_teams,omitempty"`
	WaitTimer               time.Duration           `json:"waitTimer,omitempty" yaml:"wait_timer,omitempty"`
	BranchPolicyType        EnvironmentBranchPolicy `json:"branchPolicyType" yaml:"branch_policy_type"`
	ProtectedBranches       []string                `json:"protectedBranches,omitempty" yaml:"protected_branches,omitempty"`
	BranchPatterns          []string                `json:"branchPatterns,omitempty" yaml:"branch_patterns,omitempty"`
	RequireDeploymentBranch bool                    `json:"requireDeploymentBranch" yaml:"require_deployment_branch"`
	PreventSelfReview       bool                    `json:"preventSelfReview" yaml:"prevent_self_review"`
	Secrets                 []string                `json:"secrets,omitempty" yaml:"secrets,omitempty"`
	Variables               map[string]string       `json:"variables,omitempty" yaml:"variables,omitempty"`
}

// EnvironmentBranchPolicy defines branch protection policy for environments.
type EnvironmentBranchPolicy string

const (
	EnvironmentBranchPolicyAll       EnvironmentBranchPolicy = "all"
	EnvironmentBranchPolicyProtected EnvironmentBranchPolicy = "protected"
	EnvironmentBranchPolicySelected  EnvironmentBranchPolicy = "selected"
	EnvironmentBranchPolicyNone      EnvironmentBranchPolicy = "none"
)

// RunnerPolicy defines policy for GitHub Actions runners.
type RunnerPolicy struct {
	AllowedRunnerTypes      []RunnerType           `json:"allowedRunnerTypes" yaml:"allowed_runner_types"`
	RequireSelfHostedLabels []string               `json:"requireSelfHostedLabels,omitempty" yaml:"require_self_hosted_labels,omitempty"`
	RestrictedRunnerLabels  []string               `json:"restrictedRunnerLabels,omitempty" yaml:"restricted_runner_labels,omitempty"`
	MaxConcurrentJobs       int                    `json:"maxConcurrentJobs,omitempty" yaml:"max_concurrent_jobs,omitempty"`
	MaxJobExecutionTime     time.Duration          `json:"maxJobExecutionTime,omitempty" yaml:"max_job_execution_time,omitempty"`
	RunnerGroups            []string               `json:"runnerGroups,omitempty" yaml:"runner_groups,omitempty"`
	RequireRunnerApproval   bool                   `json:"requireRunnerApproval" yaml:"require_runner_approval"`
	SelfHostedRunnerPolicy  SelfHostedRunnerPolicy `json:"selfHostedRunnerPolicy" yaml:"self_hosted_runner_policy"`
}

// RunnerType defines the type of runner allowed.
type RunnerType string

const (
	RunnerTypeGitHubHosted RunnerType = "github_hosted"
	RunnerTypeSelfHosted   RunnerType = "self_hosted"
	RunnerTypeOrganization RunnerType = "organization"
	RunnerTypeRepository   RunnerType = "repository"
)

// SelfHostedRunnerPolicy defines policy for self-hosted runners.
type SelfHostedRunnerPolicy struct {
	RequireRunnerRegistration  bool          `json:"requireRunnerRegistration" yaml:"require_runner_registration"`
	AllowedOperatingSystems    []string      `json:"allowedOperatingSystems,omitempty" yaml:"allowed_operating_systems,omitempty"`
	RequiredSecurityPatches    bool          `json:"requiredSecurityPatches" yaml:"required_security_patches"`
	DisallowPublicRepositories bool          `json:"disallowPublicRepositories" yaml:"disallow_public_repositories"`
	RequireEncryptedStorage    bool          `json:"requireEncryptedStorage" yaml:"require_encrypted_storage"`
	RunnerTimeout              time.Duration `json:"runnerTimeout,omitempty" yaml:"runner_timeout,omitempty"`
	MaxRunners                 int           `json:"maxRunners,omitempty" yaml:"max_runners,omitempty"`
}

// ActionsPolicyViolation represents a policy violation.
type ActionsPolicyViolation struct {
	ID            string                     `json:"id"`
	PolicyID      string                     `json:"policyId"`
	ViolationType ActionsPolicyViolationType `json:"violationType"`
	Severity      PolicyViolationSeverity    `json:"severity"`
	Resource      string                     `json:"resource"`
	Description   string                     `json:"description"`
	Details       map[string]interface{}     `json:"details,omitempty"`
	DetectedAt    time.Time                  `json:"detectedAt"`
	ResolvedAt    *time.Time                 `json:"resolvedAt,omitempty"`
	Status        PolicyViolationStatus      `json:"status"`
}

// ActionsPolicyViolationType defines types of policy violations.
type ActionsPolicyViolationType string

const (
	ViolationTypeUnauthorizedAction       ActionsPolicyViolationType = "unauthorized_action"
	ViolationTypeExcessivePermissions     ActionsPolicyViolationType = "excessive_permissions"
	ViolationTypeSecretMisuse             ActionsPolicyViolationType = "secret_misuse"
	ViolationTypeRunnerPolicyBreach       ActionsPolicyViolationType = "runner_policy_breach"
	ViolationTypeEnvironmentBreach        ActionsPolicyViolationType = "environment_breach"
	ViolationTypeWorkflowPermissionBreach ActionsPolicyViolationType = "workflow_permission_breach"
	ViolationTypeSecuritySettingsBreach   ActionsPolicyViolationType = "security_settings_breach"
)

// PolicyViolationSeverity defines the severity of policy violations.
type PolicyViolationSeverity string

const (
	ViolationSeverityLow      PolicyViolationSeverity = "low"
	ViolationSeverityMedium   PolicyViolationSeverity = "medium"
	ViolationSeverityHigh     PolicyViolationSeverity = "high"
	ViolationSeverityCritical PolicyViolationSeverity = "critical"
)

// PolicyViolationStatus defines the status of a policy violation.
type PolicyViolationStatus string

const (
	ViolationStatusOpen       PolicyViolationStatus = "open"
	ViolationStatusInProgress PolicyViolationStatus = "in_progress"
	ViolationStatusResolved   PolicyViolationStatus = "resolved"
	ViolationStatusIgnored    PolicyViolationStatus = "ignored"
)

// ActionsPolicyManager manages GitHub Actions policies.
type ActionsPolicyManager struct {
	logger     Logger
	apiClient  APIClient
	policies   map[string]*ActionsPolicy
	violations map[string]*ActionsPolicyViolation
}

// NewActionsPolicyManager creates a new Actions policy manager.
func NewActionsPolicyManager(logger Logger, apiClient APIClient) *ActionsPolicyManager {
	return &ActionsPolicyManager{
		logger:     logger,
		apiClient:  apiClient,
		policies:   make(map[string]*ActionsPolicy),
		violations: make(map[string]*ActionsPolicyViolation),
	}
}

// CreatePolicy creates a new Actions policy.
func (apm *ActionsPolicyManager) CreatePolicy(_ context.Context, policy *ActionsPolicy) error {
	if policy.ID == "" {
		return fmt.Errorf("policy ID is required")
	}

	if policy.Name == "" {
		return fmt.Errorf("policy name is required")
	}

	if policy.Organization == "" {
		return fmt.Errorf("organization is required")
	}

	// Validate policy configuration
	if err := apm.validatePolicy(policy); err != nil {
		return fmt.Errorf("invalid policy configuration: %w", err)
	}

	// Set timestamps
	now := time.Now()
	policy.CreatedAt = now
	policy.UpdatedAt = now
	policy.Version = 1

	// Store policy
	apm.policies[policy.ID] = policy

	apm.logger.Info("Actions policy created",
		"policy_id", policy.ID,
		"organization", policy.Organization,
		"repository", policy.Repository)

	return nil
}

// UpdatePolicy updates an existing Actions policy.
func (apm *ActionsPolicyManager) UpdatePolicy(_ context.Context, policyID string, updates *ActionsPolicy) error {
	existingPolicy, exists := apm.policies[policyID]
	if !exists {
		return fmt.Errorf("policy not found: %s", policyID)
	}

	// Validate updates
	if err := apm.validatePolicy(updates); err != nil {
		return fmt.Errorf("invalid policy updates: %w", err)
	}

	// Update policy
	updates.ID = policyID
	updates.CreatedAt = existingPolicy.CreatedAt
	updates.CreatedBy = existingPolicy.CreatedBy
	updates.UpdatedAt = time.Now()
	updates.Version = existingPolicy.Version + 1

	apm.policies[policyID] = updates

	apm.logger.Info("Actions policy updated",
		"policy_id", policyID,
		"version", updates.Version)

	return nil
}

// GetPolicy retrieves a policy by ID.
func (apm *ActionsPolicyManager) GetPolicy(_ context.Context, policyID string) (*ActionsPolicy, error) {
	policy, exists := apm.policies[policyID]
	if !exists {
		return nil, fmt.Errorf("policy not found: %s", policyID)
	}

	// Return a copy to prevent modification
	policyCopy := *policy

	return &policyCopy, nil
}

// ListPolicies lists all policies, optionally filtered by organization.
func (apm *ActionsPolicyManager) ListPolicies(_ context.Context, organization string) ([]*ActionsPolicy, error) {
	policies := make([]*ActionsPolicy, 0)

	for _, policy := range apm.policies {
		if organization == "" || policy.Organization == organization {
			policyCopy := *policy
			policies = append(policies, &policyCopy)
		}
	}

	return policies, nil
}

// DeletePolicy deletes a policy.
func (apm *ActionsPolicyManager) DeletePolicy(_ context.Context, policyID string) error {
	if _, exists := apm.policies[policyID]; !exists {
		return fmt.Errorf("policy not found: %s", policyID)
	}

	delete(apm.policies, policyID)

	apm.logger.Info("Actions policy deleted", "policy_id", policyID)

	return nil
}

// ValidatePolicy validates a policy configuration.
func (apm *ActionsPolicyManager) validatePolicy(policy *ActionsPolicy) error {
	// Validate permission level
	validPermissionLevels := map[ActionsPermissionLevel]bool{
		ActionsPermissionDisabled:        true,
		ActionsPermissionAll:             true,
		ActionsPermissionLocalOnly:       true,
		ActionsPermissionSelectedActions: true,
	}

	if !validPermissionLevels[policy.PermissionLevel] {
		return fmt.Errorf("invalid permission level: %s", policy.PermissionLevel)
	}

	// If selected actions, validate allowed actions are specified
	if policy.PermissionLevel == ActionsPermissionSelectedActions &&
		len(policy.AllowedActions) == 0 && len(policy.AllowedActionsPatterns) == 0 {
		return fmt.Errorf("allowed actions or patterns must be specified for selected permission level")
	}

	// Validate default permissions
	validDefaultPermissions := map[DefaultPermissions]bool{
		DefaultPermissionsRead:       true,
		DefaultPermissionsWrite:      true,
		DefaultPermissionsRestricted: true,
	}

	if !validDefaultPermissions[policy.WorkflowPermissions.DefaultPermissions] {
		return fmt.Errorf("invalid default permissions: %s", policy.WorkflowPermissions.DefaultPermissions)
	}

	// Validate environment policies
	for _, env := range policy.Environments {
		if env.Name == "" {
			return fmt.Errorf("environment name is required")
		}

		validBranchPolicies := map[EnvironmentBranchPolicy]bool{
			EnvironmentBranchPolicyAll:       true,
			EnvironmentBranchPolicyProtected: true,
			EnvironmentBranchPolicySelected:  true,
			EnvironmentBranchPolicyNone:      true,
		}

		if !validBranchPolicies[env.BranchPolicyType] {
			return fmt.Errorf("invalid branch policy type for environment %s: %s", env.Name, env.BranchPolicyType)
		}
	}

	// Validate runner policy
	if len(policy.Runners.AllowedRunnerTypes) == 0 {
		return fmt.Errorf("at least one runner type must be allowed")
	}

	validRunnerTypes := map[RunnerType]bool{
		RunnerTypeGitHubHosted: true,
		RunnerTypeSelfHosted:   true,
		RunnerTypeOrganization: true,
		RunnerTypeRepository:   true,
	}

	for _, runnerType := range policy.Runners.AllowedRunnerTypes {
		if !validRunnerTypes[runnerType] {
			return fmt.Errorf("invalid runner type: %s", runnerType)
		}
	}

	return nil
}

// GetDefaultActionsPolicy returns a default Actions policy template.
func GetDefaultActionsPolicy() *ActionsPolicy {
	return &ActionsPolicy{
		Name:            "Default Actions Policy",
		Description:     "Default policy for GitHub Actions",
		PermissionLevel: ActionsPermissionLocalOnly,
		WorkflowPermissions: WorkflowPermissions{
			DefaultPermissions:       DefaultPermissionsRead,
			CanApproveOwnChanges:     false,
			ActionsReadPermission:    TokenPermissionRead,
			ContentsPermission:       TokenPermissionRead,
			MetadataPermission:       TokenPermissionRead,
			PackagesPermission:       TokenPermissionNone,
			PullRequestsPermission:   TokenPermissionRead,
			IssuesPermission:         TokenPermissionRead,
			DeploymentsPermission:    TokenPermissionNone,
			ChecksPermission:         TokenPermissionNone,
			StatusesPermission:       TokenPermissionNone,
			SecurityEventsPermission: TokenPermissionNone,
			IdTokenPermission:        TokenPermissionNone,
			AttestationsPermission:   TokenPermissionNone,
		},
		SecuritySettings: ActionsSecuritySettings{
			RequireCodeScanningApproval:   true,
			RequireSecretScanningApproval: true,
			AllowForkPRs:                  false,
			RequireApprovalForForkPRs:     true,
			AllowPrivateRepoForkRun:       false,
			RequireApprovalForPrivateFork: true,
			AllowGitHubOwnedActions:       true,
			AllowVerifiedPartnerActions:   false,
			AllowMarketplaceActions:       MarketplacePolicyDisabled,
			RequireSignedCommits:          true,
			EnforceAdminsOnBranches:       true,
		},
		SecretsPolicy: SecretsPolicy{
			RequireApprovalForNewSecrets: true,
			SecretVisibility:             SecretVisibilityPrivate,
			AllowSecretsInheritance:      false,
			MaxSecretCount:               50,
			SecretRotationPolicy: SecretRotationPolicy{
				Enabled:                false,
				RotationInterval:       90 * 24 * time.Hour, // 90 days
				RequireRotationWarning: true,
				WarningDays:            7,
			},
		},
		Runners: RunnerPolicy{
			AllowedRunnerTypes:    []RunnerType{RunnerTypeGitHubHosted},
			MaxConcurrentJobs:     5,
			MaxJobExecutionTime:   6 * time.Hour,
			RequireRunnerApproval: true,
			SelfHostedRunnerPolicy: SelfHostedRunnerPolicy{
				RequireRunnerRegistration:  true,
				RequiredSecurityPatches:    true,
				DisallowPublicRepositories: true,
				RequireEncryptedStorage:    true,
				RunnerTimeout:              24 * time.Hour,
				MaxRunners:                 10,
			},
		},
		Enabled: true,
		Version: 1,
	}
}
