//nolint:testpackage // White-box testing needed for internal function access
package github

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewActionsPolicyEnforcer(t *testing.T) {
	logger := &simpleLogger{}
	apiClient := &simpleAPIClient{}
	policyManager := NewActionsPolicyManager(logger, apiClient)

	enforcer := NewActionsPolicyEnforcer(logger, apiClient, policyManager)

	assert.NotNil(t, enforcer)
	assert.Equal(t, logger, enforcer.logger)
	assert.Equal(t, apiClient, enforcer.apiClient)
	assert.Equal(t, policyManager, enforcer.policyManager)
	assert.NotEmpty(t, enforcer.validationRules)
}

func TestActionsPolicyEnforcer_ValidatePolicy(t *testing.T) {
	enforcer := createTestEnforcer()
	ctx := context.Background()

	tests := []struct {
		name          string
		policy        *ActionsPolicy
		currentState  *RepositoryActionsState
		expectedRules int
		expectPassed  bool
	}{
		{
			name: "compliant policy",
			policy: &ActionsPolicy{
				PermissionLevel: ActionsPermissionLocalOnly,
				WorkflowPermissions: WorkflowPermissions{
					DefaultPermissions: DefaultPermissionsRead,
					ContentsPermission: TokenPermissionRead,
				},
				SecuritySettings: ActionsSecuritySettings{
					AllowForkPRs:                false,
					AllowGitHubOwnedActions:     true,
					AllowVerifiedPartnerActions: false,
					AllowMarketplaceActions:     MarketplacePolicyDisabled,
				},
				SecretsPolicy: SecretsPolicy{
					MaxSecretCount: 10,
				},
				Runners: RunnerPolicy{
					AllowedRunnerTypes: []RunnerType{RunnerTypeGitHubHosted},
				},
			},
			currentState: &RepositoryActionsState{
				PermissionLevel: ActionsPermissionLocalOnly,
				WorkflowPermissions: WorkflowPermissions{
					DefaultPermissions: DefaultPermissionsRead,
					ContentsPermission: TokenPermissionRead,
				},
				SecuritySettings: ActionsSecuritySettings{
					AllowForkPRs:                false,
					AllowGitHubOwnedActions:     true,
					AllowVerifiedPartnerActions: false,
					AllowMarketplaceActions:     MarketplacePolicyDisabled,
				},
				Secrets: []SecretInfo{
					{Name: "API_KEY", Visibility: "repository"},
					{Name: "DATABASE_URL", Visibility: "repository"},
				},
				Runners: []RunnerInfo{
					{ID: 1, Name: "github-runner-1", OS: "ubuntu"},
				},
			},
			expectedRules: 6,
			expectPassed:  true,
		},
		{
			name: "non-compliant policy",
			policy: &ActionsPolicy{
				PermissionLevel: ActionsPermissionLocalOnly,
				WorkflowPermissions: WorkflowPermissions{
					DefaultPermissions: DefaultPermissionsRead,
				},
				SecuritySettings: ActionsSecuritySettings{
					AllowForkPRs:            false,
					AllowMarketplaceActions: MarketplacePolicyDisabled,
				},
				SecretsPolicy: SecretsPolicy{
					MaxSecretCount: 5,
				},
			},
			currentState: &RepositoryActionsState{
				PermissionLevel: ActionsPermissionAll, // Violation
				WorkflowPermissions: WorkflowPermissions{
					DefaultPermissions: DefaultPermissionsWrite, // Violation
				},
				SecuritySettings: ActionsSecuritySettings{
					AllowForkPRs:            true,                 // Violation
					AllowMarketplaceActions: MarketplacePolicyAll, // Violation
				},
				Secrets: []SecretInfo{ // Too many secrets - violation
					{Name: "SECRET_1"},
					{Name: "SECRET_2"},
					{Name: "SECRET_3"},
					{Name: "SECRET_4"},
					{Name: "SECRET_5"},
					{Name: "SECRET_6"},
				},
			},
			expectedRules: 6,
			expectPassed:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := enforcer.ValidatePolicy(ctx, tt.policy, tt.currentState)
			require.NoError(t, err)
			assert.Len(t, results, tt.expectedRules)

			if tt.expectPassed {
				for _, result := range results {
					assert.True(t, result.Passed, "Rule %s should pass", result.RuleID)
				}
			} else {
				failedRules := 0

				for _, result := range results {
					if !result.Passed {
						failedRules++
					}
				}

				assert.Greater(t, failedRules, 0, "At least some rules should fail")
			}
		})
	}
}

func TestActionsPolicyEnforcer_EnforcePolicy(t *testing.T) {
	enforcer := createTestEnforcer()
	ctx := context.Background()

	// Create a test policy
	policy := &ActionsPolicy{
		ID:              "test-policy",
		Name:            "Test Policy",
		Organization:    "testorg",
		PermissionLevel: ActionsPermissionLocalOnly,
		WorkflowPermissions: WorkflowPermissions{
			DefaultPermissions: DefaultPermissionsRead,
		},
		SecuritySettings: ActionsSecuritySettings{
			AllowForkPRs: false,
		},
		Enabled: true,
	}

	// Add policy to manager
	err := enforcer.policyManager.CreatePolicy(ctx, policy)
	require.NoError(t, err)

	result, err := enforcer.EnforcePolicy(ctx, "test-policy", "testorg", "testrepo")
	require.NoError(t, err)

	assert.Equal(t, "test-policy", result.PolicyID)
	assert.Equal(t, "testorg", result.Organization)
	assert.Equal(t, "testrepo", result.Repository)
	assert.NotEmpty(t, result.ValidationResult)
	assert.NotZero(t, result.ExecutionTime)
	assert.NotZero(t, result.Timestamp)
}

func TestActionsPolicyEnforcer_EnforcePolicy_DisabledPolicy(t *testing.T) {
	enforcer := createTestEnforcer()
	ctx := context.Background()

	// Create a disabled test policy
	policy := &ActionsPolicy{
		ID:           "disabled-policy",
		Name:         "Disabled Policy",
		Organization: "testorg",
		Enabled:      false, // Disabled
	}

	err := enforcer.policyManager.CreatePolicy(ctx, policy)
	require.NoError(t, err)

	result, err := enforcer.EnforcePolicy(ctx, "disabled-policy", "testorg", "testrepo")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "disabled")
	assert.NotNil(t, result)
}

func TestActionsPolicyEnforcer_EnforcePolicy_NonExistentPolicy(t *testing.T) {
	enforcer := createTestEnforcer()
	ctx := context.Background()

	result, err := enforcer.EnforcePolicy(ctx, "non-existent", "testorg", "testrepo")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	assert.NotNil(t, result)
}

func TestPermissionLevelValidationRule(t *testing.T) {
	rule := &PermissionLevelValidationRule{}
	ctx := context.Background()

	tests := []struct {
		name           string
		policy         ActionsPermissionLevel
		current        ActionsPermissionLevel
		expectPassed   bool
		expectSeverity PolicyViolationSeverity
	}{
		{
			name:           "matching permission levels",
			policy:         ActionsPermissionLocalOnly,
			current:        ActionsPermissionLocalOnly,
			expectPassed:   true,
			expectSeverity: ViolationSeverityLow,
		},
		{
			name:           "permission escalation",
			policy:         ActionsPermissionLocalOnly,
			current:        ActionsPermissionAll,
			expectPassed:   false,
			expectSeverity: ViolationSeverityHigh,
		},
		{
			name:           "permission reduction",
			policy:         ActionsPermissionAll,
			current:        ActionsPermissionLocalOnly,
			expectPassed:   false,
			expectSeverity: ViolationSeverityMedium,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := &ActionsPolicy{PermissionLevel: tt.policy}
			state := &RepositoryActionsState{PermissionLevel: tt.current}

			result, err := rule.Validate(ctx, policy, state)
			require.NoError(t, err)
			require.NotNil(t, result)

			assert.Equal(t, tt.expectPassed, result.Passed)
			assert.Equal(t, tt.expectSeverity, result.Severity)
			assert.Equal(t, rule.GetRuleID(), result.RuleID)
			assert.NotEmpty(t, result.Message)
		})
	}
}

func TestSecuritySettingsValidationRule(t *testing.T) {
	rule := &SecuritySettingsValidationRule{}
	ctx := context.Background()

	tests := []struct {
		name            string
		policySettings  ActionsSecuritySettings
		currentSettings ActionsSecuritySettings
		expectPassed    bool
		expectCritical  bool
	}{
		{
			name: "compliant security settings",
			policySettings: ActionsSecuritySettings{
				AllowForkPRs:            false,
				AllowGitHubOwnedActions: true,
				AllowMarketplaceActions: MarketplacePolicyDisabled,
			},
			currentSettings: ActionsSecuritySettings{
				AllowForkPRs:            false,
				AllowGitHubOwnedActions: true,
				AllowMarketplaceActions: MarketplacePolicyDisabled,
			},
			expectPassed:   true,
			expectCritical: false,
		},
		{
			name: "critical security violation",
			policySettings: ActionsSecuritySettings{
				AllowForkPRs:            false,
				AllowMarketplaceActions: MarketplacePolicyDisabled,
			},
			currentSettings: ActionsSecuritySettings{
				AllowForkPRs:            true,                 // Critical violation
				AllowMarketplaceActions: MarketplacePolicyAll, // Critical violation
			},
			expectPassed:   false,
			expectCritical: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := &ActionsPolicy{SecuritySettings: tt.policySettings}
			state := &RepositoryActionsState{SecuritySettings: tt.currentSettings}

			result, err := rule.Validate(ctx, policy, state)
			require.NoError(t, err)
			require.NotNil(t, result)

			assert.Equal(t, tt.expectPassed, result.Passed)

			if tt.expectCritical {
				assert.Equal(t, ViolationSeverityCritical, result.Severity)
			}
		})
	}
}

func TestAllowedActionsValidationRule(t *testing.T) {
	rule := &AllowedActionsValidationRule{}
	ctx := context.Background()

	tests := []struct {
		name         string
		policy       *ActionsPolicy
		workflows    []WorkflowInfo
		expectPassed bool
	}{
		{
			name: "non-selected permission level",
			policy: &ActionsPolicy{
				PermissionLevel: ActionsPermissionAll, // Not selected actions
			},
			workflows:    []WorkflowInfo{},
			expectPassed: true,
		},
		{
			name: "authorized actions only",
			policy: &ActionsPolicy{
				PermissionLevel:        ActionsPermissionSelectedActions,
				AllowedActions:         []string{"actions/checkout@v4", "actions/setup-go@v4"},
				AllowedActionsPatterns: []string{"actions/.*"},
			},
			workflows: []WorkflowInfo{
				{
					Name:    "test-workflow",
					Actions: []string{"actions/checkout@v4", "actions/setup-go@v4"},
				},
			},
			expectPassed: true,
		},
		{
			name: "unauthorized actions present",
			policy: &ActionsPolicy{
				PermissionLevel: ActionsPermissionSelectedActions,
				AllowedActions:  []string{"actions/checkout@v4"},
			},
			workflows: []WorkflowInfo{
				{
					Name:    "test-workflow",
					Actions: []string{"actions/checkout@v4", "unauthorized/action@v1"},
				},
			},
			expectPassed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &RepositoryActionsState{RecentWorkflows: tt.workflows}

			result, err := rule.Validate(ctx, tt.policy, state)
			require.NoError(t, err)
			require.NotNil(t, result)

			assert.Equal(t, tt.expectPassed, result.Passed)
		})
	}
}

func TestSecretPolicyValidationRule(t *testing.T) {
	rule := &SecretPolicyValidationRule{}
	ctx := context.Background()

	tests := []struct {
		name         string
		policy       SecretsPolicy
		secrets      []SecretInfo
		expectPassed bool
	}{
		{
			name: "within secret limit",
			policy: SecretsPolicy{
				MaxSecretCount: 5,
			},
			secrets: []SecretInfo{
				{Name: "SECRET_1"},
				{Name: "SECRET_2"},
			},
			expectPassed: true,
		},
		{
			name: "exceeds secret limit",
			policy: SecretsPolicy{
				MaxSecretCount: 2,
			},
			secrets: []SecretInfo{
				{Name: "SECRET_1"},
				{Name: "SECRET_2"},
				{Name: "SECRET_3"},
			},
			expectPassed: false,
		},
		{
			name: "restricted secret present",
			policy: SecretsPolicy{
				RestrictedSecrets: []string{"DANGEROUS_SECRET"},
			},
			secrets: []SecretInfo{
				{Name: "SAFE_SECRET"},
				{Name: "DANGEROUS_SECRET"},
			},
			expectPassed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := &ActionsPolicy{SecretsPolicy: tt.policy}
			state := &RepositoryActionsState{Secrets: tt.secrets}

			result, err := rule.Validate(ctx, policy, state)
			require.NoError(t, err)
			require.NotNil(t, result)

			assert.Equal(t, tt.expectPassed, result.Passed)
		})
	}
}

func TestRunnerPolicyValidationRule(t *testing.T) {
	rule := &RunnerPolicyValidationRule{}
	ctx := context.Background()

	tests := []struct {
		name         string
		policy       RunnerPolicy
		runners      []RunnerInfo
		expectPassed bool
	}{
		{
			name: "github hosted runners allowed",
			policy: RunnerPolicy{
				AllowedRunnerTypes: []RunnerType{RunnerTypeGitHubHosted},
			},
			runners: []RunnerInfo{
				{Name: "github-runner-1", OS: "ubuntu"},
			},
			expectPassed: true,
		},
		{
			name: "self-hosted runner not allowed",
			policy: RunnerPolicy{
				AllowedRunnerTypes: []RunnerType{RunnerTypeGitHubHosted},
			},
			runners: []RunnerInfo{
				{Name: "self-hosted-runner-1", OS: "ubuntu"},
			},
			expectPassed: false,
		},
		{
			name: "exceeds runner limit",
			policy: RunnerPolicy{
				AllowedRunnerTypes: []RunnerType{RunnerTypeSelfHosted},
				SelfHostedRunnerPolicy: SelfHostedRunnerPolicy{
					MaxRunners: 1,
				},
			},
			runners: []RunnerInfo{
				{Name: "self-hosted-runner-1", OS: "ubuntu"},
				{Name: "self-hosted-runner-2", OS: "ubuntu"},
			},
			expectPassed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := &ActionsPolicy{Runners: tt.policy}
			state := &RepositoryActionsState{Runners: tt.runners}

			result, err := rule.Validate(ctx, policy, state)
			require.NoError(t, err)
			require.NotNil(t, result)

			assert.Equal(t, tt.expectPassed, result.Passed)
		})
	}
}

func TestHelperFunctions(t *testing.T) {
	t.Run("isPermissionEscalation", func(t *testing.T) {
		assert.True(t, isPermissionEscalation(ActionsPermissionLocalOnly, ActionsPermissionAll))
		assert.False(t, isPermissionEscalation(ActionsPermissionAll, ActionsPermissionLocalOnly))
		assert.False(t, isPermissionEscalation(ActionsPermissionLocalOnly, ActionsPermissionLocalOnly))
	})

	t.Run("isPermissionTooHigh", func(t *testing.T) {
		assert.True(t, isPermissionTooHigh(TokenPermissionRead, TokenPermissionWrite))
		assert.False(t, isPermissionTooHigh(TokenPermissionWrite, TokenPermissionRead))
		assert.False(t, isPermissionTooHigh(TokenPermissionRead, TokenPermissionRead))
	})

	t.Run("isMarketplacePolicyMorePermissive", func(t *testing.T) {
		assert.True(t, isMarketplacePolicyMorePermissive(MarketplacePolicyAll, MarketplacePolicyDisabled))
		assert.False(t, isMarketplacePolicyMorePermissive(MarketplacePolicyDisabled, MarketplacePolicyAll))
		assert.False(t, isMarketplacePolicyMorePermissive(MarketplacePolicyDisabled, MarketplacePolicyDisabled))
	})
}

// Benchmark tests.
func BenchmarkEnforcePolicy(b *testing.B) {
	enforcer := createTestEnforcer()
	ctx := context.Background()

	policy := &ActionsPolicy{
		ID:              "bench-policy",
		Name:            "Benchmark Policy",
		Organization:    "testorg",
		PermissionLevel: ActionsPermissionLocalOnly,
		Enabled:         true,
	}

	_ = enforcer.policyManager.CreatePolicy(ctx, policy)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = enforcer.EnforcePolicy(ctx, "bench-policy", "testorg", "testrepo")
	}
}

func BenchmarkValidatePolicy(b *testing.B) {
	enforcer := createTestEnforcer()
	ctx := context.Background()

	policy := &ActionsPolicy{
		PermissionLevel: ActionsPermissionLocalOnly,
		WorkflowPermissions: WorkflowPermissions{
			DefaultPermissions: DefaultPermissionsRead,
		},
		SecuritySettings: ActionsSecuritySettings{
			AllowForkPRs: false,
		},
		SecretsPolicy: SecretsPolicy{
			MaxSecretCount: 10,
		},
		Runners: RunnerPolicy{
			AllowedRunnerTypes: []RunnerType{RunnerTypeGitHubHosted},
		},
	}

	state := &RepositoryActionsState{
		PermissionLevel: ActionsPermissionLocalOnly,
		WorkflowPermissions: WorkflowPermissions{
			DefaultPermissions: DefaultPermissionsRead,
		},
		SecuritySettings: ActionsSecuritySettings{
			AllowForkPRs: false,
		},
		Secrets: []SecretInfo{{Name: "TEST_SECRET"}},
		Runners: []RunnerInfo{{Name: "test-runner"}},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := enforcer.ValidatePolicy(ctx, policy, state)
		if err != nil {
			// Ignore error in benchmark
		}
	}
}

// Helper function to create a test enforcer.
func createTestEnforcer() *ActionsPolicyEnforcer {
	logger := &simpleLogger{}
	apiClient := &simpleAPIClient{}
	policyManager := NewActionsPolicyManager(logger, apiClient)

	return NewActionsPolicyEnforcer(logger, apiClient, policyManager)
}
