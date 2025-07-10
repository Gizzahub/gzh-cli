package github

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

// PermissionLevelValidationRule validates Actions permission level compliance
type PermissionLevelValidationRule struct{}

func (r *PermissionLevelValidationRule) GetRuleID() string {
	return "actions_permission_level"
}

func (r *PermissionLevelValidationRule) GetDescription() string {
	return "Validates Actions permission level compliance"
}

func (r *PermissionLevelValidationRule) Validate(ctx context.Context, policy *ActionsPolicy, currentState *RepositoryActionsState) (*PolicyValidationResult, error) {
	result := &PolicyValidationResult{
		RuleID:        r.GetRuleID(),
		ActualValue:   currentState.PermissionLevel,
		ExpectedValue: policy.PermissionLevel,
	}

	if policy.PermissionLevel == currentState.PermissionLevel {
		result.Passed = true
		result.Message = "Actions permission level is compliant"
		result.Severity = ViolationSeverityLow
	} else {
		result.Passed = false
		result.Message = fmt.Sprintf("Actions permission level mismatch: expected %s, got %s",
			policy.PermissionLevel, currentState.PermissionLevel)

		// Determine severity based on permission escalation
		if isPermissionEscalation(policy.PermissionLevel, currentState.PermissionLevel) {
			result.Severity = ViolationSeverityHigh
			result.Suggestions = []string{
				"Reduce Actions permission level to match policy",
				"Review repository Actions usage before reducing permissions",
			}
		} else {
			result.Severity = ViolationSeverityMedium
			result.Suggestions = []string{
				"Update Actions permission level to match policy",
			}
		}
	}

	return result, nil
}

// WorkflowPermissionsValidationRule validates workflow token permissions
type WorkflowPermissionsValidationRule struct{}

func (r *WorkflowPermissionsValidationRule) GetRuleID() string {
	return "workflow_permissions"
}

func (r *WorkflowPermissionsValidationRule) GetDescription() string {
	return "Validates workflow token permissions compliance"
}

func (r *WorkflowPermissionsValidationRule) Validate(ctx context.Context, policy *ActionsPolicy, currentState *RepositoryActionsState) (*PolicyValidationResult, error) {
	result := &PolicyValidationResult{
		RuleID:        r.GetRuleID(),
		ActualValue:   currentState.WorkflowPermissions.DefaultPermissions,
		ExpectedValue: policy.WorkflowPermissions.DefaultPermissions,
	}

	violations := make([]string, 0)

	// Check default permissions
	if policy.WorkflowPermissions.DefaultPermissions != currentState.WorkflowPermissions.DefaultPermissions {
		violations = append(violations, fmt.Sprintf("Default permissions mismatch: expected %s, got %s",
			policy.WorkflowPermissions.DefaultPermissions, currentState.WorkflowPermissions.DefaultPermissions))
	}

	// Check individual permissions (simplified check)
	if policy.WorkflowPermissions.ContentsPermission != TokenPermissionNone &&
		isPermissionTooHigh(policy.WorkflowPermissions.ContentsPermission, currentState.WorkflowPermissions.ContentsPermission) {
		violations = append(violations, fmt.Sprintf("Contents permission too high: expected %s or lower, got %s",
			policy.WorkflowPermissions.ContentsPermission, currentState.WorkflowPermissions.ContentsPermission))
	}

	if len(violations) == 0 {
		result.Passed = true
		result.Message = "Workflow permissions are compliant"
		result.Severity = ViolationSeverityLow
	} else {
		result.Passed = false
		result.Message = strings.Join(violations, "; ")
		result.Severity = ViolationSeverityMedium
		result.Suggestions = []string{
			"Update workflow permissions to match policy requirements",
			"Review existing workflows that may depend on current permissions",
		}
		result.Details = map[string]interface{}{
			"violations": violations,
		}
	}

	return result, nil
}

// SecuritySettingsValidationRule validates security settings compliance
type SecuritySettingsValidationRule struct{}

func (r *SecuritySettingsValidationRule) GetRuleID() string {
	return "security_settings"
}

func (r *SecuritySettingsValidationRule) GetDescription() string {
	return "Validates security settings compliance"
}

func (r *SecuritySettingsValidationRule) Validate(ctx context.Context, policy *ActionsPolicy, currentState *RepositoryActionsState) (*PolicyValidationResult, error) {
	result := &PolicyValidationResult{
		RuleID: r.GetRuleID(),
	}

	violations := make([]string, 0)
	criticalViolations := 0

	// Check fork PR settings
	if policy.SecuritySettings.AllowForkPRs != currentState.SecuritySettings.AllowForkPRs {
		if currentState.SecuritySettings.AllowForkPRs && !policy.SecuritySettings.AllowForkPRs {
			violations = append(violations, "Fork PRs are enabled but policy requires them to be disabled")
			criticalViolations++
		} else {
			violations = append(violations, "Fork PR setting does not match policy")
		}
	}

	// Check GitHub owned actions
	if policy.SecuritySettings.AllowGitHubOwnedActions != currentState.SecuritySettings.AllowGitHubOwnedActions {
		if currentState.SecuritySettings.AllowGitHubOwnedActions && !policy.SecuritySettings.AllowGitHubOwnedActions {
			violations = append(violations, "GitHub owned actions are enabled but policy requires them to be disabled")
			criticalViolations++
		} else {
			violations = append(violations, "GitHub owned actions setting does not match policy")
		}
	}

	// Check marketplace actions
	if policy.SecuritySettings.AllowMarketplaceActions != currentState.SecuritySettings.AllowMarketplaceActions {
		if isMarketplacePolicyMorePermissive(currentState.SecuritySettings.AllowMarketplaceActions, policy.SecuritySettings.AllowMarketplaceActions) {
			violations = append(violations, fmt.Sprintf("Marketplace actions policy is too permissive: current %s, policy requires %s",
				currentState.SecuritySettings.AllowMarketplaceActions, policy.SecuritySettings.AllowMarketplaceActions))
			criticalViolations++
		} else {
			violations = append(violations, "Marketplace actions policy does not match")
		}
	}

	if len(violations) == 0 {
		result.Passed = true
		result.Message = "Security settings are compliant"
		result.Severity = ViolationSeverityLow
	} else {
		result.Passed = false
		result.Message = strings.Join(violations, "; ")

		if criticalViolations > 0 {
			result.Severity = ViolationSeverityCritical
			result.Suggestions = []string{
				"URGENT: Review and tighten security settings immediately",
				"Disable overly permissive settings that violate security policy",
				"Audit recent workflow runs for potential security issues",
			}
		} else {
			result.Severity = ViolationSeverityMedium
			result.Suggestions = []string{
				"Update security settings to match policy",
			}
		}

		result.Details = map[string]interface{}{
			"violations":          violations,
			"critical_violations": criticalViolations,
		}
	}

	return result, nil
}

// AllowedActionsValidationRule validates allowed actions compliance
type AllowedActionsValidationRule struct{}

func (r *AllowedActionsValidationRule) GetRuleID() string {
	return "allowed_actions"
}

func (r *AllowedActionsValidationRule) GetDescription() string {
	return "Validates allowed actions compliance"
}

func (r *AllowedActionsValidationRule) Validate(ctx context.Context, policy *ActionsPolicy, currentState *RepositoryActionsState) (*PolicyValidationResult, error) {
	result := &PolicyValidationResult{
		RuleID: r.GetRuleID(),
	}

	// Only validate if policy uses selected actions
	if policy.PermissionLevel != ActionsPermissionSelectedActions {
		result.Passed = true
		result.Message = "Actions permission level does not require specific action validation"
		result.Severity = ViolationSeverityLow
		return result, nil
	}

	violations := make([]string, 0)
	unauthorizedActions := make([]string, 0)

	// Check recent workflows for unauthorized actions
	for _, workflow := range currentState.RecentWorkflows {
		for _, action := range workflow.Actions {
			if !r.isActionAllowed(action, policy.AllowedActions, policy.AllowedActionsPatterns) {
				unauthorizedActions = append(unauthorizedActions, fmt.Sprintf("%s in workflow %s", action, workflow.Name))
			}
		}
	}

	if len(unauthorizedActions) > 0 {
		result.Passed = false
		result.Message = fmt.Sprintf("Found %d unauthorized actions in recent workflows", len(unauthorizedActions))
		result.Severity = ViolationSeverityHigh
		result.Suggestions = []string{
			"Review and remove unauthorized actions from workflows",
			"Update allowed actions list if these actions are legitimate",
			"Consider using action patterns for more flexible policies",
		}
		result.Details = map[string]interface{}{
			"unauthorized_actions": unauthorizedActions,
		}
	} else {
		result.Passed = true
		result.Message = "All actions in recent workflows are authorized"
		result.Severity = ViolationSeverityLow
	}

	return result, nil
}

func (r *AllowedActionsValidationRule) isActionAllowed(action string, allowedActions, allowedPatterns []string) bool {
	// Check exact matches
	for _, allowed := range allowedActions {
		if action == allowed {
			return true
		}
	}

	// Check pattern matches
	for _, pattern := range allowedPatterns {
		if matched, _ := regexp.MatchString(pattern, action); matched {
			return true
		}
	}

	return false
}

// SecretPolicyValidationRule validates secret policy compliance
type SecretPolicyValidationRule struct{}

func (r *SecretPolicyValidationRule) GetRuleID() string {
	return "secret_policy"
}

func (r *SecretPolicyValidationRule) GetDescription() string {
	return "Validates secret policy compliance"
}

func (r *SecretPolicyValidationRule) Validate(ctx context.Context, policy *ActionsPolicy, currentState *RepositoryActionsState) (*PolicyValidationResult, error) {
	result := &PolicyValidationResult{
		RuleID: r.GetRuleID(),
	}

	violations := make([]string, 0)

	// Check secret count limit
	if policy.SecretsPolicy.MaxSecretCount > 0 && len(currentState.Secrets) > policy.SecretsPolicy.MaxSecretCount {
		violations = append(violations, fmt.Sprintf("Too many secrets: %d exceeds limit of %d",
			len(currentState.Secrets), policy.SecretsPolicy.MaxSecretCount))
	}

	// Check secret naming patterns
	if len(policy.SecretsPolicy.SecretNamingPatterns) > 0 {
		for _, secret := range currentState.Secrets {
			if !r.matchesNamingPattern(secret.Name, policy.SecretsPolicy.SecretNamingPatterns) {
				violations = append(violations, fmt.Sprintf("Secret '%s' does not match naming patterns", secret.Name))
			}
		}
	}

	// Check restricted secrets
	for _, secret := range currentState.Secrets {
		for _, restricted := range policy.SecretsPolicy.RestrictedSecrets {
			if secret.Name == restricted {
				violations = append(violations, fmt.Sprintf("Restricted secret '%s' is present", secret.Name))
			}
		}
	}

	if len(violations) == 0 {
		result.Passed = true
		result.Message = "Secret policy is compliant"
		result.Severity = ViolationSeverityLow
	} else {
		result.Passed = false
		result.Message = strings.Join(violations, "; ")
		result.Severity = ViolationSeverityMedium
		result.Suggestions = []string{
			"Review and clean up secrets that violate policy",
			"Update secret names to match naming patterns",
			"Remove restricted secrets",
		}
		result.Details = map[string]interface{}{
			"violations":   violations,
			"secret_count": len(currentState.Secrets),
			"secret_limit": policy.SecretsPolicy.MaxSecretCount,
		}
	}

	return result, nil
}

func (r *SecretPolicyValidationRule) matchesNamingPattern(secretName string, patterns []string) bool {
	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, secretName); matched {
			return true
		}
	}
	return false
}

// RunnerPolicyValidationRule validates runner policy compliance
type RunnerPolicyValidationRule struct{}

func (r *RunnerPolicyValidationRule) GetRuleID() string {
	return "runner_policy"
}

func (r *RunnerPolicyValidationRule) GetDescription() string {
	return "Validates runner policy compliance"
}

func (r *RunnerPolicyValidationRule) Validate(ctx context.Context, policy *ActionsPolicy, currentState *RepositoryActionsState) (*PolicyValidationResult, error) {
	result := &PolicyValidationResult{
		RuleID: r.GetRuleID(),
	}

	violations := make([]string, 0)

	// Check runner count limits
	if policy.Runners.SelfHostedRunnerPolicy.MaxRunners > 0 {
		selfHostedCount := 0
		for _, runner := range currentState.Runners {
			if r.isSelfHostedRunner(runner) {
				selfHostedCount++
			}
		}

		if selfHostedCount > policy.Runners.SelfHostedRunnerPolicy.MaxRunners {
			violations = append(violations, fmt.Sprintf("Too many self-hosted runners: %d exceeds limit of %d",
				selfHostedCount, policy.Runners.SelfHostedRunnerPolicy.MaxRunners))
		}
	}

	// Check runner types
	for _, runner := range currentState.Runners {
		runnerType := r.getRunnerType(runner)
		if !r.isRunnerTypeAllowed(runnerType, policy.Runners.AllowedRunnerTypes) {
			violations = append(violations, fmt.Sprintf("Runner '%s' type '%s' is not allowed", runner.Name, runnerType))
		}
	}

	// Check required labels for self-hosted runners
	if len(policy.Runners.RequireSelfHostedLabels) > 0 {
		for _, runner := range currentState.Runners {
			if r.isSelfHostedRunner(runner) {
				if !r.hasRequiredLabels(runner, policy.Runners.RequireSelfHostedLabels) {
					violations = append(violations, fmt.Sprintf("Self-hosted runner '%s' missing required labels", runner.Name))
				}
			}
		}
	}

	if len(violations) == 0 {
		result.Passed = true
		result.Message = "Runner policy is compliant"
		result.Severity = ViolationSeverityLow
	} else {
		result.Passed = false
		result.Message = strings.Join(violations, "; ")
		result.Severity = ViolationSeverityMedium
		result.Suggestions = []string{
			"Review runner configuration and remove non-compliant runners",
			"Add required labels to self-hosted runners",
			"Ensure runner types match policy requirements",
		}
		result.Details = map[string]interface{}{
			"violations":   violations,
			"runner_count": len(currentState.Runners),
		}
	}

	return result, nil
}

func (r *RunnerPolicyValidationRule) isSelfHostedRunner(runner RunnerInfo) bool {
	// Simplified check - in reality would check runner type from GitHub API
	return strings.Contains(strings.ToLower(runner.Name), "self-hosted")
}

func (r *RunnerPolicyValidationRule) getRunnerType(runner RunnerInfo) RunnerType {
	if r.isSelfHostedRunner(runner) {
		return RunnerTypeSelfHosted
	}
	return RunnerTypeGitHubHosted
}

func (r *RunnerPolicyValidationRule) isRunnerTypeAllowed(runnerType RunnerType, allowedTypes []RunnerType) bool {
	for _, allowed := range allowedTypes {
		if runnerType == allowed {
			return true
		}
	}
	return false
}

func (r *RunnerPolicyValidationRule) hasRequiredLabels(runner RunnerInfo, requiredLabels []string) bool {
	for _, required := range requiredLabels {
		found := false
		for _, label := range runner.Labels {
			if label == required {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// Helper functions

func isPermissionEscalation(expected, actual ActionsPermissionLevel) bool {
	permissionLevels := map[ActionsPermissionLevel]int{
		ActionsPermissionDisabled:        0,
		ActionsPermissionSelectedActions: 1,
		ActionsPermissionLocalOnly:       2,
		ActionsPermissionAll:             3,
	}

	expectedLevel, expectedExists := permissionLevels[expected]
	actualLevel, actualExists := permissionLevels[actual]

	if !expectedExists || !actualExists {
		return false
	}

	return actualLevel > expectedLevel
}

func isPermissionTooHigh(expected, actual ActionsTokenPermission) bool {
	permissionLevels := map[ActionsTokenPermission]int{
		TokenPermissionNone:  0,
		TokenPermissionRead:  1,
		TokenPermissionWrite: 2,
	}

	expectedLevel, expectedExists := permissionLevels[expected]
	actualLevel, actualExists := permissionLevels[actual]

	if !expectedExists || !actualExists {
		return false
	}

	return actualLevel > expectedLevel
}

func isMarketplacePolicyMorePermissive(actual, expected ActionsMarketplacePolicy) bool {
	permissionLevels := map[ActionsMarketplacePolicy]int{
		MarketplacePolicyDisabled:     0,
		MarketplacePolicySelected:     1,
		MarketplacePolicyVerifiedOnly: 2,
		MarketplacePolicyAll:          3,
	}

	expectedLevel, expectedExists := permissionLevels[expected]
	actualLevel, actualExists := permissionLevels[actual]

	if !expectedExists || !actualExists {
		return false
	}

	return actualLevel > expectedLevel
}
