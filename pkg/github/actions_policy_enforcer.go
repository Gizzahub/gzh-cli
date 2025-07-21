package github

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// ActionsPolicyEnforcer handles the enforcement and validation of Actions policies.
type ActionsPolicyEnforcer struct {
	logger          Logger
	apiClient       APIClient
	policyManager   *ActionsPolicyManager
	validationRules []PolicyValidationRule
}

// PolicyValidationRule defines a rule for validating policy compliance.
// Implementations check specific aspects of GitHub Actions configuration
// against organizational policies and return validation results.
type PolicyValidationRule interface {
	Validate(ctx context.Context, policy *ActionsPolicy, currentState *RepositoryActionsState) (*PolicyValidationResult, error)
	GetRuleID() string
	GetDescription() string
}

// PolicyValidationResult represents the result of a policy validation.
type PolicyValidationResult struct {
	RuleID        string                  `json:"ruleId"`
	Passed        bool                    `json:"passed"`
	Severity      PolicyViolationSeverity `json:"severity"`
	Message       string                  `json:"message"`
	Details       map[string]interface{}  `json:"details,omitempty"`
	Suggestions   []string                `json:"suggestions,omitempty"`
	ActualValue   interface{}             `json:"actualValue,omitempty"`
	ExpectedValue interface{}             `json:"expectedValue,omitempty"`
}

// RepositoryActionsState represents the current Actions configuration state of a repository.
type RepositoryActionsState struct {
	Organization        string                  `json:"organization"`
	Repository          string                  `json:"repository"`
	ActionsEnabled      bool                    `json:"actionsEnabled"`
	PermissionLevel     ActionsPermissionLevel  `json:"permissionLevel"`
	AllowedActions      []string                `json:"allowedActions,omitempty"`
	WorkflowPermissions WorkflowPermissions     `json:"workflowPermissions"`
	SecuritySettings    ActionsSecuritySettings `json:"securitySettings"`
	Secrets             []SecretInfo            `json:"secrets,omitempty"`
	Variables           map[string]string       `json:"variables,omitempty"`
	Environments        []EnvironmentInfo       `json:"environments,omitempty"`
	Runners             []RunnerInfo            `json:"runners,omitempty"`
	RecentWorkflows     []WorkflowInfo          `json:"recentWorkflows,omitempty"`
	LastUpdated         time.Time               `json:"lastUpdated"`
}

// SecretInfo represents information about a repository secret.
type SecretInfo struct {
	Name        string    `json:"name"`
	Visibility  string    `json:"visibility"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Environment string    `json:"environment,omitempty"`
}

// EnvironmentInfo represents information about a repository environment.
type EnvironmentInfo struct {
	Name               string            `json:"name"`
	ProtectionRules    []ProtectionRule  `json:"protectionRules,omitempty"`
	DeploymentBranches []string          `json:"deploymentBranches,omitempty"`
	Secrets            []SecretInfo      `json:"secrets,omitempty"`
	Variables          map[string]string `json:"variables,omitempty"`
}

// ProtectionRule represents an environment protection rule.
type ProtectionRule struct {
	Type      string   `json:"type"`
	Reviewers []string `json:"reviewers,omitempty"`
	WaitTimer int      `json:"waitTimer,omitempty"`
}

// RunnerInfo represents information about a repository runner.
type RunnerInfo struct {
	ID     int64    `json:"id"`
	Name   string   `json:"name"`
	Status string   `json:"status"`
	OS     string   `json:"os"`
	Labels []string `json:"labels"`
	Busy   bool     `json:"busy"`
}

// WorkflowInfo represents information about a workflow.
type WorkflowInfo struct {
	ID          int64             `json:"id"`
	Name        string            `json:"name"`
	Path        string            `json:"path"`
	State       string            `json:"state"`
	Permissions map[string]string `json:"permissions,omitempty"`
	Actions     []string          `json:"actions,omitempty"`
	LastRun     time.Time         `json:"lastRun"`
}

// PolicyEnforcementResult represents the result of applying a policy.
type PolicyEnforcementResult struct {
	PolicyID         string                   `json:"policyId"`
	Organization     string                   `json:"organization"`
	Repository       string                   `json:"repository"`
	Success          bool                     `json:"success"`
	AppliedChanges   []PolicyChange           `json:"appliedChanges"`
	FailedChanges    []PolicyChange           `json:"failedChanges"`
	ValidationResult []PolicyValidationResult `json:"validationResult"`
	Violations       []ActionsPolicyViolation `json:"violations,omitempty"`
	ExecutionTime    time.Duration            `json:"executionTime"`
	Timestamp        time.Time                `json:"timestamp"`
}

// PolicyChange represents a change made during policy enforcement.
type PolicyChange struct {
	Type     string      `json:"type"`
	Target   string      `json:"target"`
	Action   string      `json:"action"`
	OldValue interface{} `json:"oldValue,omitempty"`
	NewValue interface{} `json:"newValue"`
	Success  bool        `json:"success"`
	Error    string      `json:"error,omitempty"`
}

// NewActionsPolicyEnforcer creates a new Actions policy enforcer that validates
// and enforces GitHub Actions policies across repositories. It registers default
// validation rules and provides methods to scan workflows for compliance.
func NewActionsPolicyEnforcer(logger Logger, apiClient APIClient, policyManager *ActionsPolicyManager) *ActionsPolicyEnforcer {
	enforcer := &ActionsPolicyEnforcer{
		logger:          logger,
		apiClient:       apiClient,
		policyManager:   policyManager,
		validationRules: make([]PolicyValidationRule, 0),
	}

	// Register default validation rules
	enforcer.registerDefaultValidationRules()

	return enforcer
}

// EnforcePolicy applies an Actions policy to a repository.
func (ape *ActionsPolicyEnforcer) EnforcePolicy(ctx context.Context, policyID, organization, repository string) (*PolicyEnforcementResult, error) {
	startTime := time.Now()

	result := &PolicyEnforcementResult{
		PolicyID:       policyID,
		Organization:   organization,
		Repository:     repository,
		AppliedChanges: make([]PolicyChange, 0),
		FailedChanges:  make([]PolicyChange, 0),
		Timestamp:      startTime,
	}

	// Get policy
	policy, err := ape.policyManager.GetPolicy(ctx, policyID)
	if err != nil {
		return result, fmt.Errorf("failed to get policy %s: %w", policyID, err)
	}

	if !policy.Enabled {
		return result, fmt.Errorf("policy %s is disabled", policyID)
	}

	// Get current repository state
	currentState := ape.getRepositoryActionsState(ctx, organization, repository)

	// Validate policy before enforcement
	validationResults, err := ape.ValidatePolicy(ctx, policy, currentState)
	if err != nil {
		return result, fmt.Errorf("failed to validate policy: %w", err)
	}

	result.ValidationResult = validationResults

	// Apply policy changes
	if err := ape.applyPolicyChanges(ctx, policy, currentState, result); err != nil {
		result.Success = false
		return result, fmt.Errorf("failed to apply policy changes: %w", err)
	}

	// Check for violations after enforcement
	violations := ape.detectViolations(policy, currentState, validationResults)
	result.Violations = violations

	result.Success = len(result.FailedChanges) == 0
	result.ExecutionTime = time.Since(startTime)

	ape.logger.Info("Policy enforcement completed",
		"policy_id", policyID,
		"organization", organization,
		"repository", repository,
		"success", result.Success,
		"applied_changes", len(result.AppliedChanges),
		"failed_changes", len(result.FailedChanges),
		"violations", len(result.Violations))

	return result, nil
}

// ValidatePolicy validates a policy against current repository state.
func (ape *ActionsPolicyEnforcer) ValidatePolicy(ctx context.Context, policy *ActionsPolicy, currentState *RepositoryActionsState) ([]PolicyValidationResult, error) {
	results := make([]PolicyValidationResult, 0)

	for _, rule := range ape.validationRules {
		result, err := rule.Validate(ctx, policy, currentState)
		if err != nil {
			ape.logger.Error("Failed to validate rule",
				"rule_id", rule.GetRuleID(),
				"error", err)

			continue
		}

		if result != nil {
			results = append(results, *result)
		}
	}

	return results, nil
}

// GetRepositoryActionsState retrieves the current Actions configuration state.
func (ape *ActionsPolicyEnforcer) getRepositoryActionsState(_ context.Context, organization, repository string) *RepositoryActionsState {
	// This would typically make GitHub API calls to get the current state
	// For now, return a mock state structure
	state := &RepositoryActionsState{
		Organization:    organization,
		Repository:      repository,
		ActionsEnabled:  true,
		PermissionLevel: ActionsPermissionAll,
		WorkflowPermissions: WorkflowPermissions{
			DefaultPermissions: DefaultPermissionsWrite,
		},
		SecuritySettings: ActionsSecuritySettings{
			AllowForkPRs:                false,
			AllowGitHubOwnedActions:     true,
			AllowVerifiedPartnerActions: false,
			AllowMarketplaceActions:     MarketplacePolicyDisabled,
		},
		LastUpdated: time.Now(),
	}

	return state
}

// applyPolicyChanges applies the necessary changes to enforce the policy.
func (ape *ActionsPolicyEnforcer) applyPolicyChanges(ctx context.Context, policy *ActionsPolicy, currentState *RepositoryActionsState, result *PolicyEnforcementResult) error {
	// Apply permission level changes
	if err := ape.applyPermissionLevelChanges(ctx, policy, currentState, result); err != nil {
		return err
	}

	// Apply workflow permission changes
	if err := ape.applyWorkflowPermissionChanges(ctx, policy, currentState, result); err != nil {
		return err
	}

	// Apply security setting changes
	if err := ape.applySecuritySettingChanges(ctx, policy, currentState, result); err != nil {
		return err
	}

	// Apply secret policy changes
	if err := ape.applySecretPolicyChanges(ctx, policy, currentState, result); err != nil {
		return err
	}

	// Apply environment policy changes
	if err := ape.applyEnvironmentPolicyChanges(ctx, policy, currentState, result); err != nil {
		return err
	}

	return nil
}

// applyPermissionLevelChanges applies Actions permission level changes.
func (ape *ActionsPolicyEnforcer) applyPermissionLevelChanges(ctx context.Context, policy *ActionsPolicy, currentState *RepositoryActionsState, result *PolicyEnforcementResult) error {
	if policy.PermissionLevel != currentState.PermissionLevel {
		change := PolicyChange{
			Type:     "actions_permission",
			Target:   "permission_level",
			Action:   "update",
			OldValue: currentState.PermissionLevel,
			NewValue: policy.PermissionLevel,
		}

		// Simulate API call to update permission level
		if err := ape.updateActionsPermissionLevel(ctx, currentState.Organization, currentState.Repository, policy.PermissionLevel); err != nil {
			change.Success = false
			change.Error = err.Error()
			result.FailedChanges = append(result.FailedChanges, change)

			return err
		}

		change.Success = true
		result.AppliedChanges = append(result.AppliedChanges, change)
		currentState.PermissionLevel = policy.PermissionLevel
	}

	return nil
}

// applyWorkflowPermissionChanges applies workflow permission changes.
func (ape *ActionsPolicyEnforcer) applyWorkflowPermissionChanges(ctx context.Context, policy *ActionsPolicy, currentState *RepositoryActionsState, result *PolicyEnforcementResult) error {
	if policy.WorkflowPermissions.DefaultPermissions != currentState.WorkflowPermissions.DefaultPermissions {
		change := PolicyChange{
			Type:     "workflow_permissions",
			Target:   "default_permissions",
			Action:   "update",
			OldValue: currentState.WorkflowPermissions.DefaultPermissions,
			NewValue: policy.WorkflowPermissions.DefaultPermissions,
		}

		// Simulate API call to update workflow permissions
		if err := ape.updateWorkflowPermissions(ctx, currentState.Organization, currentState.Repository, &policy.WorkflowPermissions); err != nil {
			change.Success = false
			change.Error = err.Error()
			result.FailedChanges = append(result.FailedChanges, change)

			return err
		}

		change.Success = true
		result.AppliedChanges = append(result.AppliedChanges, change)
		currentState.WorkflowPermissions = policy.WorkflowPermissions
	}

	return nil
}

// applySecuritySettingChanges applies security setting changes.
func (ape *ActionsPolicyEnforcer) applySecuritySettingChanges(ctx context.Context, policy *ActionsPolicy, currentState *RepositoryActionsState, result *PolicyEnforcementResult) error {
	// Apply fork PR settings
	if policy.SecuritySettings.AllowForkPRs != currentState.SecuritySettings.AllowForkPRs {
		change := PolicyChange{
			Type:     "security_settings",
			Target:   "allow_fork_prs",
			Action:   "update",
			OldValue: currentState.SecuritySettings.AllowForkPRs,
			NewValue: policy.SecuritySettings.AllowForkPRs,
		}

		if err := ape.updateSecuritySettings(ctx, currentState.Organization, currentState.Repository, &policy.SecuritySettings); err != nil {
			change.Success = false
			change.Error = err.Error()
			result.FailedChanges = append(result.FailedChanges, change)

			return err
		}

		change.Success = true
		result.AppliedChanges = append(result.AppliedChanges, change)
		currentState.SecuritySettings = policy.SecuritySettings
	}

	return nil
}

// applySecretPolicyChanges applies secret policy changes.
func (ape *ActionsPolicyEnforcer) applySecretPolicyChanges(_ context.Context, policy *ActionsPolicy, currentState *RepositoryActionsState, result *PolicyEnforcementResult) error {
	// This would implement secret policy enforcement
	// For now, just log that we would apply secret policies
	ape.logger.Info("Applying secret policy changes",
		"organization", currentState.Organization,
		"repository", currentState.Repository,
		"max_secret_count", policy.SecretsPolicy.MaxSecretCount)

	return nil
}

// applyEnvironmentPolicyChanges applies environment policy changes.
func (ape *ActionsPolicyEnforcer) applyEnvironmentPolicyChanges(_ context.Context, policy *ActionsPolicy, currentState *RepositoryActionsState, result *PolicyEnforcementResult) error {
	// This would implement environment policy enforcement
	// For now, just log that we would apply environment policies
	ape.logger.Info("Applying environment policy changes",
		"organization", currentState.Organization,
		"repository", currentState.Repository,
		"environment_count", len(policy.Environments))

	return nil
}

// detectViolations detects policy violations based on validation results.
func (ape *ActionsPolicyEnforcer) detectViolations(policy *ActionsPolicy, currentState *RepositoryActionsState, validationResults []PolicyValidationResult) []ActionsPolicyViolation {
	violations := make([]ActionsPolicyViolation, 0)

	for _, result := range validationResults {
		if !result.Passed {
			violation := ActionsPolicyViolation{
				ID:            fmt.Sprintf("violation-%d", time.Now().UnixNano()),
				PolicyID:      policy.ID,
				ViolationType: ape.mapViolationType(result.RuleID),
				Severity:      result.Severity,
				Resource:      fmt.Sprintf("%s/%s", currentState.Organization, currentState.Repository),
				Description:   result.Message,
				Details: map[string]interface{}{
					"rule_id":        result.RuleID,
					"actual_value":   result.ActualValue,
					"expected_value": result.ExpectedValue,
					"suggestions":    result.Suggestions,
				},
				DetectedAt: time.Now(),
				Status:     ViolationStatusOpen,
			}
			violations = append(violations, violation)
		}
	}

	return violations
}

// mapViolationType maps a validation rule ID to a violation type.
func (ape *ActionsPolicyEnforcer) mapViolationType(ruleID string) ActionsPolicyViolationType {
	switch {
	case strings.Contains(ruleID, "permission"):
		return ViolationTypeExcessivePermissions
	case strings.Contains(ruleID, "action"):
		return ViolationTypeUnauthorizedAction
	case strings.Contains(ruleID, "secret"):
		return ViolationTypeSecretMisuse
	case strings.Contains(ruleID, "runner"):
		return ViolationTypeRunnerPolicyBreach
	case strings.Contains(ruleID, "environment"):
		return ViolationTypeEnvironmentBreach
	case strings.Contains(ruleID, "workflow"):
		return ViolationTypeWorkflowPermissionBreach
	case strings.Contains(ruleID, "security"):
		return ViolationTypeSecuritySettingsBreach
	default:
		return ViolationTypeUnauthorizedAction
	}
}

// Mock API methods (these would be real GitHub API calls in production)

func (ape *ActionsPolicyEnforcer) updateActionsPermissionLevel(ctx context.Context, org, repo string, level ActionsPermissionLevel) error {
	ape.logger.Info("Updating Actions permission level",
		"organization", org,
		"repository", repo,
		"level", level)

	return nil
}

func (ape *ActionsPolicyEnforcer) updateWorkflowPermissions(ctx context.Context, org, repo string, permissions *WorkflowPermissions) error {
	ape.logger.Info("Updating workflow permissions",
		"organization", org,
		"repository", repo,
		"default_permissions", permissions.DefaultPermissions)

	return nil
}

func (ape *ActionsPolicyEnforcer) updateSecuritySettings(ctx context.Context, org, repo string, settings *ActionsSecuritySettings) error {
	ape.logger.Info("Updating security settings",
		"organization", org,
		"repository", repo,
		"allow_fork_prs", settings.AllowForkPRs)

	return nil
}

// registerDefaultValidationRules registers the default set of validation rules.
func (ape *ActionsPolicyEnforcer) registerDefaultValidationRules() {
	ape.validationRules = append(ape.validationRules,
		&PermissionLevelValidationRule{},
		&WorkflowPermissionsValidationRule{},
		&SecuritySettingsValidationRule{},
		&AllowedActionsValidationRule{},
		&SecretPolicyValidationRule{},
		&RunnerPolicyValidationRule{},
	)
}

// AddValidationRule adds a custom validation rule.
func (ape *ActionsPolicyEnforcer) AddValidationRule(rule PolicyValidationRule) {
	ape.validationRules = append(ape.validationRules, rule)
}
