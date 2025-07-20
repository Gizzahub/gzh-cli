//nolint:testpackage // White-box testing needed for internal function access
package github

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestActionsPolicy_BasicCreation(t *testing.T) {
	// Test creating a default policy
	defaultPolicy := GetDefaultActionsPolicy()

	assert.NotNil(t, defaultPolicy)
	assert.Equal(t, "Default Actions Policy", defaultPolicy.Name)
	assert.Equal(t, ActionsPermissionLocalOnly, defaultPolicy.PermissionLevel)
	assert.True(t, defaultPolicy.Enabled)
	assert.Equal(t, 1, defaultPolicy.Version)
}

func TestActionsPermissionLevel_Constants(t *testing.T) {
	assert.Equal(t, ActionsPermissionLevel("disabled"), ActionsPermissionDisabled)
	assert.Equal(t, ActionsPermissionLevel("all"), ActionsPermissionAll)
	assert.Equal(t, ActionsPermissionLevel("local_only"), ActionsPermissionLocalOnly)
	assert.Equal(t, ActionsPermissionLevel("selected"), ActionsPermissionSelectedActions)
}

func TestDefaultPermissions_Constants(t *testing.T) {
	assert.Equal(t, DefaultPermissions("read"), DefaultPermissionsRead)
	assert.Equal(t, DefaultPermissions("write"), DefaultPermissionsWrite)
	assert.Equal(t, DefaultPermissions("restricted"), DefaultPermissionsRestricted)
}

func TestActionsTokenPermission_Constants(t *testing.T) {
	assert.Equal(t, ActionsTokenPermission("none"), TokenPermissionNone)
	assert.Equal(t, ActionsTokenPermission("read"), TokenPermissionRead)
	assert.Equal(t, ActionsTokenPermission("write"), TokenPermissionWrite)
}

func TestRunnerType_Constants(t *testing.T) {
	assert.Equal(t, RunnerType("github_hosted"), RunnerTypeGitHubHosted)
	assert.Equal(t, RunnerType("self_hosted"), RunnerTypeSelfHosted)
	assert.Equal(t, RunnerType("organization"), RunnerTypeOrganization)
	assert.Equal(t, RunnerType("repository"), RunnerTypeRepository)
}

func TestPolicyViolationTypes_Constants(t *testing.T) {
	assert.Equal(t, ActionsPolicyViolationType("unauthorized_action"), ViolationTypeUnauthorizedAction)
	assert.Equal(t, ActionsPolicyViolationType("excessive_permissions"), ViolationTypeExcessivePermissions)
	assert.Equal(t, ActionsPolicyViolationType("secret_misuse"), ViolationTypeSecretMisuse)
	assert.Equal(t, ActionsPolicyViolationType("runner_policy_breach"), ViolationTypeRunnerPolicyBreach)
	assert.Equal(t, ActionsPolicyViolationType("environment_breach"), ViolationTypeEnvironmentBreach)
	assert.Equal(t, ActionsPolicyViolationType("workflow_permission_breach"), ViolationTypeWorkflowPermissionBreach)
	assert.Equal(t, ActionsPolicyViolationType("security_settings_breach"), ViolationTypeSecuritySettingsBreach)
}

func TestSecretVisibility_Constants(t *testing.T) {
	assert.Equal(t, SecretVisibility("all"), SecretVisibilityAll)
	assert.Equal(t, SecretVisibility("private"), SecretVisibilityPrivate)
	assert.Equal(t, SecretVisibility("selected"), SecretVisibilitySelectedRepos)
}

func TestMarketplacePolicy_Constants(t *testing.T) {
	assert.Equal(t, ActionsMarketplacePolicy("disabled"), MarketplacePolicyDisabled)
	assert.Equal(t, ActionsMarketplacePolicy("verified_only"), MarketplacePolicyVerifiedOnly)
	assert.Equal(t, ActionsMarketplacePolicy("all"), MarketplacePolicyAll)
	assert.Equal(t, ActionsMarketplacePolicy("selected"), MarketplacePolicySelected)
}

func TestEnvironmentBranchPolicy_Constants(t *testing.T) {
	assert.Equal(t, EnvironmentBranchPolicy("all"), EnvironmentBranchPolicyAll)
	assert.Equal(t, EnvironmentBranchPolicy("protected"), EnvironmentBranchPolicyProtected)
	assert.Equal(t, EnvironmentBranchPolicy("selected"), EnvironmentBranchPolicySelected)
	assert.Equal(t, EnvironmentBranchPolicy("none"), EnvironmentBranchPolicyNone)
}

func TestPolicyViolationSeverity_Constants(t *testing.T) {
	assert.Equal(t, PolicyViolationSeverity("low"), ViolationSeverityLow)
	assert.Equal(t, PolicyViolationSeverity("medium"), ViolationSeverityMedium)
	assert.Equal(t, PolicyViolationSeverity("high"), ViolationSeverityHigh)
	assert.Equal(t, PolicyViolationSeverity("critical"), ViolationSeverityCritical)
}

func TestPolicyViolationStatus_Constants(t *testing.T) {
	assert.Equal(t, PolicyViolationStatus("open"), ViolationStatusOpen)
	assert.Equal(t, PolicyViolationStatus("in_progress"), ViolationStatusInProgress)
	assert.Equal(t, PolicyViolationStatus("resolved"), ViolationStatusResolved)
	assert.Equal(t, PolicyViolationStatus("ignored"), ViolationStatusIgnored)
}

// Simple mock logger for basic testing.
type simpleLogger struct{}

func (l *simpleLogger) Debug(msg string, keysAndValues ...interface{}) {}
func (l *simpleLogger) Info(msg string, keysAndValues ...interface{})  {}
func (l *simpleLogger) Warn(msg string, keysAndValues ...interface{})  {}
func (l *simpleLogger) Error(msg string, keysAndValues ...interface{}) {}

// Simple mock API client for basic testing.
type simpleAPIClient struct{}

func (m *simpleAPIClient) GetRepository(ctx context.Context, owner, repo string) (*RepositoryInfo, error) {
	return &RepositoryInfo{}, nil
}

func (m *simpleAPIClient) ListOrganizationRepositories(ctx context.Context, org string) ([]RepositoryInfo, error) {
	return []RepositoryInfo{}, nil
}

func (m *simpleAPIClient) GetDefaultBranch(ctx context.Context, owner, repo string) (string, error) {
	return "main", nil
}

func (m *simpleAPIClient) SetToken(token string) {}

func (m *simpleAPIClient) GetRateLimit(ctx context.Context) (*RateLimit, error) {
	return &RateLimit{}, nil
}

func (m *simpleAPIClient) GetRepositoryConfiguration(ctx context.Context, owner, repo string) (*RepositoryConfig, error) {
	return &RepositoryConfig{}, nil
}

func (m *simpleAPIClient) UpdateRepositoryConfiguration(ctx context.Context, owner, repo string, config *RepositoryConfig) error {
	return nil
}

func TestActionsPolicy_ManagerCreation(t *testing.T) {
	logger := &simpleLogger{}
	apiClient := &simpleAPIClient{}

	manager := NewActionsPolicyManager(logger, apiClient)

	assert.NotNil(t, manager)
	assert.NotNil(t, manager.policies)
	assert.NotNil(t, manager.violations)
}
