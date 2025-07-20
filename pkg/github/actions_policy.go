package github

import (
	"context"
	"fmt"
	"time"
)

// ActionsPermissionLevel defines the permission level for GitHub Actions.
type ActionsPermissionLevel string

const (
	// ActionsPermissionDisabled disables GitHub Actions for the repository/organization.
	ActionsPermissionDisabled ActionsPermissionLevel = "disabled"
	// ActionsPermissionAll allows all GitHub Actions to run.
	ActionsPermissionAll ActionsPermissionLevel = "all"
	// ActionsPermissionLocalOnly allows only local actions and workflows to run.
	ActionsPermissionLocalOnly ActionsPermissionLevel = "local_only"
	// ActionsPermissionSelectedActions allows only selected actions to run.
	ActionsPermissionSelectedActions ActionsPermissionLevel = "selected"
)

// ActionsPolicy represents a GitHub Actions permission policy.
type ActionsPolicy struct {
	ID                     string                  `json:"id" yaml:"id"`
	Name                   string                  `json:"name" yaml:"name"`
	Description            string                  `json:"description" yaml:"description"`
	Organization           string                  `json:"organization" yaml:"organization"`
	Repository             string                  `json:"repository,omitempty" yaml:"repository,omitempty"`
	PermissionLevel        ActionsPermissionLevel  `json:"permissionLevel" yaml:"permissionLevel"`
	AllowedActions         []string                `json:"allowedActions,omitempty" yaml:"allowedActions,omitempty"`
	AllowedActionsPatterns []string                `json:"allowedActionsPatterns,omitempty" yaml:"allowedActionsPatterns,omitempty"`
	WorkflowPermissions    WorkflowPermissions     `json:"workflowPermissions" yaml:"workflowPermissions"`
	SecuritySettings       ActionsSecuritySettings `json:"securitySettings" yaml:"securitySettings"`
	SecretsPolicy          SecretsPolicy           `json:"secretsPolicy" yaml:"secretsPolicy"`
	Variables              map[string]string       `json:"variables,omitempty" yaml:"variables,omitempty"`
	Environments           []EnvironmentPolicy     `json:"environments,omitempty" yaml:"environments,omitempty"`
	Runners                RunnerPolicy            `json:"runners" yaml:"runners"`
	CreatedAt              time.Time               `json:"createdAt" yaml:"createdAt"`
	UpdatedAt              time.Time               `json:"updatedAt" yaml:"updatedAt"`
	CreatedBy              string                  `json:"createdBy" yaml:"createdBy"`
	UpdatedBy              string                  `json:"updatedBy" yaml:"updatedBy"`
	Version                int                     `json:"version" yaml:"version"`
	Enabled                bool                    `json:"enabled" yaml:"enabled"`
	Tags                   []string                `json:"tags,omitempty" yaml:"tags,omitempty"`
}

// WorkflowPermissions defines permissions for workflow tokens.
type WorkflowPermissions struct {
	DefaultPermissions       DefaultPermissions                `json:"defaultPermissions" yaml:"defaultPermissions"`
	CanApproveOwnChanges     bool                              `json:"canApproveOwnChanges" yaml:"canApproveOwnChanges"`
	ActionsReadPermission    ActionsTokenPermission            `json:"actionsRead" yaml:"actionsRead"`
	ContentsPermission       ActionsTokenPermission            `json:"contents" yaml:"contents"`
	MetadataPermission       ActionsTokenPermission            `json:"metadata" yaml:"metadata"`
	PackagesPermission       ActionsTokenPermission            `json:"packages" yaml:"packages"`
	PullRequestsPermission   ActionsTokenPermission            `json:"pullRequests" yaml:"pullRequests"`
	IssuesPermission         ActionsTokenPermission            `json:"issues" yaml:"issues"`
	DeploymentsPermission    ActionsTokenPermission            `json:"deployments" yaml:"deployments"`
	ChecksPermission         ActionsTokenPermission            `json:"checks" yaml:"checks"`
	StatusesPermission       ActionsTokenPermission            `json:"statuses" yaml:"statuses"`
	SecurityEventsPermission ActionsTokenPermission            `json:"securityEvents" yaml:"securityEvents"`
	IdTokenPermission        ActionsTokenPermission            `json:"idToken" yaml:"idToken"`
	AttestationsPermission   ActionsTokenPermission            `json:"attestations" yaml:"attestations"`
	CustomPermissions        map[string]ActionsTokenPermission `json:"customPermissions,omitempty" yaml:"customPermissions,omitempty"`
}

// DefaultPermissions defines the default permission level for workflow tokens.
type DefaultPermissions string

const (
	// DefaultPermissionsRead grants read-only permissions to workflow tokens.
	DefaultPermissionsRead DefaultPermissions = "read"
	// DefaultPermissionsWrite grants write permissions to workflow tokens.
	DefaultPermissionsWrite DefaultPermissions = "write"
	// DefaultPermissionsRestricted restricts permissions for workflow tokens.
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
	RequireCodeScanningApproval   bool                     `json:"requireCodeScanningApproval" yaml:"requireCodeScanningApproval"`
	RequireSecretScanningApproval bool                     `json:"requireSecretScanningApproval" yaml:"requireSecretScanningApproval"`
	AllowForkPRs                  bool                     `json:"allowForkPRs" yaml:"allowForkPrs"`
	RequireApprovalForForkPRs     bool                     `json:"requireApprovalForForkPRs" yaml:"requireApprovalForForkPrs"`
	AllowPrivateRepoForkRun       bool                     `json:"allowPrivateRepoForkRun" yaml:"allowPrivateRepoForkRun"`
	RequireApprovalForPrivateFork bool                     `json:"requireApprovalForPrivateFork" yaml:"requireApprovalForPrivateFork"`
	RestrictedActionsPatterns     []string                 `json:"restrictedActionsPatterns,omitempty" yaml:"restrictedActionsPatterns,omitempty"`
	AllowGitHubOwnedActions       bool                     `json:"allowGitHubOwnedActions" yaml:"allowGithubOwnedActions"`
	AllowVerifiedPartnerActions   bool                     `json:"allowVerifiedPartnerActions" yaml:"allowVerifiedPartnerActions"`
	AllowMarketplaceActions       ActionsMarketplacePolicy `json:"allowMarketplaceActions" yaml:"allowMarketplaceActions"`
	RequireSignedCommits          bool                     `json:"requireSignedCommits" yaml:"requireSignedCommits"`
	EnforceAdminsOnBranches       bool                     `json:"enforceAdminsOnBranches" yaml:"enforceAdminsOnBranches"`
	OIDCCustomClaims              map[string]string        `json:"oidcCustomClaims,omitempty" yaml:"oidcCustomClaims,omitempty"`
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
	AllowedSecrets               []string             `json:"allowedSecrets,omitempty" yaml:"allowedSecrets,omitempty"`
	RestrictedSecrets            []string             `json:"restrictedSecrets,omitempty" yaml:"restrictedSecrets,omitempty"`
	RequireApprovalForNewSecrets bool                 `json:"requireApprovalForNewSecrets" yaml:"requireApprovalForNewSecrets"`
	SecretVisibility             SecretVisibility     `json:"secretVisibility" yaml:"secretVisibility"`
	AllowSecretsInheritance      bool                 `json:"allowSecretsInheritance" yaml:"allowSecretsInheritance"`
	SecretNamingPatterns         []string             `json:"secretNamingPatterns,omitempty" yaml:"secretNamingPatterns,omitempty"`
	MaxSecretCount               int                  `json:"maxSecretCount,omitempty" yaml:"maxSecretCount,omitempty"`
	SecretRotationPolicy         SecretRotationPolicy `json:"secretRotationPolicy" yaml:"secretRotationPolicy"`
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
	RotationInterval       time.Duration `json:"rotationInterval" yaml:"rotationInterval"`
	RequireRotationWarning bool          `json:"requireRotationWarning" yaml:"requireRotationWarning"`
	WarningDays            int           `json:"warningDays" yaml:"warningDays"`
	AutoRotateSecrets      []string      `json:"autoRotateSecrets,omitempty" yaml:"autoRotateSecrets,omitempty"`
}

// EnvironmentPolicy defines policy for deployment environments.
type EnvironmentPolicy struct {
	Name                    string                  `json:"name" yaml:"name"`
	RequiredReviewers       []string                `json:"requiredReviewers,omitempty" yaml:"requiredReviewers,omitempty"`
	RequiredReviewerTeams   []string                `json:"requiredReviewerTeams,omitempty" yaml:"requiredReviewerTeams,omitempty"`
	WaitTimer               time.Duration           `json:"waitTimer,omitempty" yaml:"waitTimer,omitempty"`
	BranchPolicyType        EnvironmentBranchPolicy `json:"branchPolicyType" yaml:"branchPolicyType"`
	ProtectedBranches       []string                `json:"protectedBranches,omitempty" yaml:"protectedBranches,omitempty"`
	BranchPatterns          []string                `json:"branchPatterns,omitempty" yaml:"branchPatterns,omitempty"`
	RequireDeploymentBranch bool                    `json:"requireDeploymentBranch" yaml:"requireDeploymentBranch"`
	PreventSelfReview       bool                    `json:"preventSelfReview" yaml:"preventSelfReview"`
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
	AllowedRunnerTypes      []RunnerType           `json:"allowedRunnerTypes" yaml:"allowedRunnerTypes"`
	RequireSelfHostedLabels []string               `json:"requireSelfHostedLabels,omitempty" yaml:"requireSelfHostedLabels,omitempty"`
	RestrictedRunnerLabels  []string               `json:"restrictedRunnerLabels,omitempty" yaml:"restrictedRunnerLabels,omitempty"`
	MaxConcurrentJobs       int                    `json:"maxConcurrentJobs,omitempty" yaml:"maxConcurrentJobs,omitempty"`
	MaxJobExecutionTime     time.Duration          `json:"maxJobExecutionTime,omitempty" yaml:"maxJobExecutionTime,omitempty"`
	RunnerGroups            []string               `json:"runnerGroups,omitempty" yaml:"runnerGroups,omitempty"`
	RequireRunnerApproval   bool                   `json:"requireRunnerApproval" yaml:"requireRunnerApproval"`
	SelfHostedRunnerPolicy  SelfHostedRunnerPolicy `json:"selfHostedRunnerPolicy" yaml:"selfHostedRunnerPolicy"`
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
	RequireRunnerRegistration  bool          `json:"requireRunnerRegistration" yaml:"requireRunnerRegistration"`
	AllowedOperatingSystems    []string      `json:"allowedOperatingSystems,omitempty" yaml:"allowedOperatingSystems,omitempty"`
	RequiredSecurityPatches    bool          `json:"requiredSecurityPatches" yaml:"requiredSecurityPatches"`
	DisallowPublicRepositories bool          `json:"disallowPublicRepositories" yaml:"disallowPublicRepositories"`
	RequireEncryptedStorage    bool          `json:"requireEncryptedStorage" yaml:"requireEncryptedStorage"`
	RunnerTimeout              time.Duration `json:"runnerTimeout,omitempty" yaml:"runnerTimeout,omitempty"`
	MaxRunners                 int           `json:"maxRunners,omitempty" yaml:"maxRunners,omitempty"`
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
