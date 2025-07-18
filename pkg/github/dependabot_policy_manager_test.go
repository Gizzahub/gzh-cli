package github

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDependabotPolicyManager(t *testing.T) {
	logger := &simpleLogger{}
	apiClient := &simpleAPIClient{}
	configManager := createTestDependabotManager()

	policyManager := NewDependabotPolicyManager(logger, apiClient, configManager)

	assert.NotNil(t, policyManager)
	assert.Equal(t, logger, policyManager.logger)
	assert.Equal(t, apiClient, policyManager.apiClient)
	assert.Equal(t, configManager, policyManager.configManager)
	assert.NotNil(t, policyManager.policies)
	assert.NotNil(t, policyManager.cache)
}

func TestDependabotPolicyManager_CreatePolicy(t *testing.T) {
	policyManager := createTestPolicyManager()
	ctx := context.Background()

	tests := []struct {
		name        string
		policy      *DependabotPolicyConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid policy",
			policy: &DependabotPolicyConfig{
				ID:           "test-policy",
				Name:         "Test Policy",
				Organization: "testorg",
				Description:  "Test policy for unit tests",
				Enabled:      true,
				DefaultConfig: DependabotConfig{
					Version: 2,
					Updates: []DependabotUpdateRule{
						{
							PackageEcosystem: EcosystemGoModules,
							Directory:        "/",
							Schedule: DependabotSchedule{
								Interval: IntervalWeekly,
							},
						},
					},
				},
				EcosystemPolicies: map[string]EcosystemPolicy{
					EcosystemGoModules: {
						Ecosystem:         EcosystemGoModules,
						Enabled:           true,
						RequiredReviewers: 1,
					},
				},
			},
			expectError: false,
		},
		{
			name: "missing policy ID",
			policy: &DependabotPolicyConfig{
				Name:         "Test Policy",
				Organization: "testorg",
				Enabled:      true,
				DefaultConfig: DependabotConfig{
					Version: 2,
					Updates: []DependabotUpdateRule{
						{
							PackageEcosystem: EcosystemGoModules,
							Directory:        "/",
							Schedule: DependabotSchedule{
								Interval: IntervalWeekly,
							},
						},
					},
				},
			},
			expectError: true,
			errorMsg:    "policy ID is required",
		},
		{
			name: "missing organization",
			policy: &DependabotPolicyConfig{
				ID:      "test-policy",
				Name:    "Test Policy",
				Enabled: true,
				DefaultConfig: DependabotConfig{
					Version: 2,
					Updates: []DependabotUpdateRule{
						{
							PackageEcosystem: EcosystemGoModules,
							Directory:        "/",
							Schedule: DependabotSchedule{
								Interval: IntervalWeekly,
							},
						},
					},
				},
			},
			expectError: true,
			errorMsg:    "organization is required",
		},
		{
			name: "invalid default configuration",
			policy: &DependabotPolicyConfig{
				ID:           "test-policy",
				Name:         "Test Policy",
				Organization: "testorg",
				Enabled:      true,
				DefaultConfig: DependabotConfig{
					Version: 1, // Invalid version
					Updates: []DependabotUpdateRule{},
				},
			},
			expectError: true,
			errorMsg:    "invalid default configuration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := policyManager.CreatePolicy(ctx, tt.policy)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
				assert.NotZero(t, tt.policy.CreatedAt)
				assert.NotZero(t, tt.policy.UpdatedAt)
				assert.Equal(t, 1, tt.policy.Version)

				// Verify policy was stored
				stored, err := policyManager.GetPolicy(ctx, tt.policy.ID)
				assert.NoError(t, err)
				assert.Equal(t, tt.policy.ID, stored.ID)
				assert.Equal(t, tt.policy.Name, stored.Name)
			}
		})
	}
}

func TestDependabotPolicyManager_GetPolicy(t *testing.T) {
	policyManager := createTestPolicyManager()
	ctx := context.Background()

	// Create a test policy
	policy := createTestDependabotPolicy()
	err := policyManager.CreatePolicy(ctx, policy)
	require.NoError(t, err)

	tests := []struct {
		name        string
		policyID    string
		expectError bool
	}{
		{
			name:        "existing policy",
			policyID:    policy.ID,
			expectError: false,
		},
		{
			name:        "non-existent policy",
			policyID:    "non-existent",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			retrieved, err := policyManager.GetPolicy(ctx, tt.policyID)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, retrieved)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, retrieved)
				assert.Equal(t, tt.policyID, retrieved.ID)
			}
		})
	}
}

func TestDependabotPolicyManager_UpdatePolicy(t *testing.T) {
	policyManager := createTestPolicyManager()
	ctx := context.Background()

	// Create a test policy
	policy := createTestDependabotPolicy()
	err := policyManager.CreatePolicy(ctx, policy)
	require.NoError(t, err)

	// Update the policy
	originalVersion := policy.Version
	originalCreatedAt := policy.CreatedAt
	policy.Description = "Updated description"
	policy.Enabled = false

	err = policyManager.UpdatePolicy(ctx, policy)
	assert.NoError(t, err)

	// Verify the update
	updated, err := policyManager.GetPolicy(ctx, policy.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated description", updated.Description)
	assert.False(t, updated.Enabled)
	assert.Equal(t, originalVersion+1, updated.Version)
	assert.Equal(t, originalCreatedAt, updated.CreatedAt)
	assert.True(t, updated.UpdatedAt.After(originalCreatedAt))

	// Test updating non-existent policy
	nonExistentPolicy := createTestDependabotPolicy()
	nonExistentPolicy.ID = "non-existent"
	err = policyManager.UpdatePolicy(ctx, nonExistentPolicy)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "policy not found")
}

func TestDependabotPolicyManager_DeletePolicy(t *testing.T) {
	policyManager := createTestPolicyManager()
	ctx := context.Background()

	// Create a test policy
	policy := createTestDependabotPolicy()
	err := policyManager.CreatePolicy(ctx, policy)
	require.NoError(t, err)

	// Delete the policy
	err = policyManager.DeletePolicy(ctx, policy.ID)
	assert.NoError(t, err)

	// Verify deletion
	_, err = policyManager.GetPolicy(ctx, policy.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "policy not found")

	// Test deleting non-existent policy
	err = policyManager.DeletePolicy(ctx, "non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "policy not found")
}

func TestDependabotPolicyManager_EvaluateRepositoryCompliance(t *testing.T) {
	policyManager := createTestPolicyManager()
	ctx := context.Background()

	// Create a test policy
	policy := createTestDependabotPolicy()
	err := policyManager.CreatePolicy(ctx, policy)
	require.NoError(t, err)

	// Test compliance evaluation
	result, err := policyManager.EvaluateRepositoryCompliance(ctx, policy.ID, "testorg", "testrepo")
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, policy.ID, result.PolicyID)
	assert.Equal(t, "testorg", result.Organization)
	assert.Equal(t, "testrepo", result.Repository)
	assert.NotZero(t, result.EvaluatedAt)
	assert.NotZero(t, result.NextEvaluation)
	assert.GreaterOrEqual(t, result.ComplianceScore, 0.0)
	assert.LessOrEqual(t, result.ComplianceScore, 100.0)

	// Test caching - second call should return cached result
	result2, err := policyManager.EvaluateRepositoryCompliance(ctx, policy.ID, "testorg", "testrepo")
	require.NoError(t, err)
	assert.Equal(t, result.EvaluatedAt, result2.EvaluatedAt)

	// Test with non-existent policy
	_, err = policyManager.EvaluateRepositoryCompliance(ctx, "non-existent", "testorg", "testrepo")
	assert.Error(t, err)
}

func TestDependabotPolicyManager_ApplyPolicyToOrganization(t *testing.T) {
	policyManager := createTestPolicyManager()
	ctx := context.Background()

	// Create a test policy
	policy := createTestDependabotPolicy()
	err := policyManager.CreatePolicy(ctx, policy)
	require.NoError(t, err)

	// Apply policy to organization
	operation, err := policyManager.ApplyPolicyToOrganization(ctx, policy.ID, "testorg")
	require.NoError(t, err)
	require.NotNil(t, operation)

	assert.Equal(t, BulkOperationTypeApplyPolicy, operation.Type)
	assert.Equal(t, "testorg", operation.Organization)
	assert.Equal(t, policy.ID, operation.PolicyID)
	assert.Equal(t, BulkOperationStatusPending, operation.Status)
	assert.NotZero(t, operation.StartedAt)
	assert.Greater(t, operation.EstimatedDuration, time.Duration(0))
	assert.NotEmpty(t, operation.ID)

	// Wait a bit for the async operation to potentially start
	time.Sleep(100 * time.Millisecond)

	// Test with non-existent policy
	_, err = policyManager.ApplyPolicyToOrganization(ctx, "non-existent", "testorg")
	assert.Error(t, err)
}

func TestDependabotPolicyManager_GenerateOrganizationReport(t *testing.T) {
	policyManager := createTestPolicyManager()
	ctx := context.Background()

	// Create a test policy
	policy := createTestDependabotPolicy()
	err := policyManager.CreatePolicy(ctx, policy)
	require.NoError(t, err)

	// Generate report
	report, err := policyManager.GenerateOrganizationReport(ctx, policy.ID, "testorg")
	require.NoError(t, err)
	require.NotNil(t, report)

	assert.Equal(t, "testorg", report.Organization)
	assert.Equal(t, policy.ID, report.PolicyID)
	assert.NotZero(t, report.GeneratedAt)
	assert.NotNil(t, report.Summary)
	assert.NotNil(t, report.RepositoryResults)
	assert.NotNil(t, report.TopViolations)
	assert.NotNil(t, report.Recommendations)
	assert.NotNil(t, report.TrendAnalysis)
	assert.NotEmpty(t, report.ExportFormats)

	// Check summary statistics
	summary := report.Summary
	assert.GreaterOrEqual(t, summary.TotalRepositories, 0)
	assert.GreaterOrEqual(t, summary.ComplianceRate, 0.0)
	assert.LessOrEqual(t, summary.ComplianceRate, 100.0)
	assert.NotNil(t, summary.EcosystemBreakdown)
	assert.NotNil(t, summary.ViolationBreakdown)

	// Test with non-existent policy
	_, err = policyManager.GenerateOrganizationReport(ctx, "non-existent", "testorg")
	assert.Error(t, err)
}

func TestDependabotPolicyViolationTypeConstants(t *testing.T) {
	violationTypes := []DependabotPolicyViolationType{
		DependabotViolationTypeMissingConfig, DependabotViolationTypeInvalidConfig,
		DependabotViolationTypeDisabledEcosystem, DependabotViolationTypeInsufficientSchedule,
		DependabotViolationTypeExcessivePermissions, DependabotViolationTypeMissingSecurityUpdates,
		DependabotViolationTypeUnauthorizedDependency, DependabotViolationTypeOutdatedPolicy,
		DependabotViolationTypeComplianceBreach,
	}

	for _, vType := range violationTypes {
		assert.NotEmpty(t, string(vType))
	}

	// Test specific values
	assert.Equal(t, DependabotPolicyViolationType("missing_config"), DependabotViolationTypeMissingConfig)
	assert.Equal(t, DependabotPolicyViolationType("invalid_config"), DependabotViolationTypeInvalidConfig)
	assert.Equal(t, DependabotPolicyViolationType("disabled_ecosystem"), DependabotViolationTypeDisabledEcosystem)
}

func TestPolicySeverityConstants(t *testing.T) {
	severities := []PolicySeverity{
		PolicySeverityCritical, PolicySeverityHigh,
		PolicySeverityMedium, PolicySeverityLow, PolicySeverityInfo,
	}

	for _, severity := range severities {
		assert.NotEmpty(t, string(severity))
	}

	// Test specific values
	assert.Equal(t, PolicySeverity("critical"), PolicySeverityCritical)
	assert.Equal(t, PolicySeverity("high"), PolicySeverityHigh)
	assert.Equal(t, PolicySeverity("medium"), PolicySeverityMedium)
}

func TestBulkOperationTypeConstants(t *testing.T) {
	operationTypes := []BulkOperationType{
		BulkOperationTypeApplyPolicy, BulkOperationTypeValidatePolicy,
		BulkOperationTypeUpdateConfig, BulkOperationTypeEnableEcosystem,
		BulkOperationTypeGenerateReport,
	}

	for _, opType := range operationTypes {
		assert.NotEmpty(t, string(opType))
	}

	// Test specific values
	assert.Equal(t, BulkOperationType("apply_policy"), BulkOperationTypeApplyPolicy)
	assert.Equal(t, BulkOperationType("validate_policy"), BulkOperationTypeValidatePolicy)
	assert.Equal(t, BulkOperationType("generate_report"), BulkOperationTypeGenerateReport)
}

func TestTrendDirectionConstants(t *testing.T) {
	directions := []TrendDirection{
		TrendDirectionImproving, TrendDirectionStable,
		TrendDirectionDeclining, TrendDirectionUnknown,
	}

	for _, direction := range directions {
		assert.NotEmpty(t, string(direction))
	}

	// Test specific values
	assert.Equal(t, TrendDirection("improving"), TrendDirectionImproving)
	assert.Equal(t, TrendDirection("stable"), TrendDirectionStable)
	assert.Equal(t, TrendDirection("declining"), TrendDirectionDeclining)
}

func TestDependabotPolicyManager_PolicyEvaluation(t *testing.T) {
	policyManager := createTestPolicyManager()

	// Create test policy
	policy := &DependabotPolicyConfig{
		ID:           "test-policy",
		Name:         "Test Policy",
		Organization: "testorg",
		Enabled:      true,
		DefaultConfig: DependabotConfig{
			Version: 2,
			Updates: []DependabotUpdateRule{
				{
					PackageEcosystem: EcosystemGoModules,
					Directory:        "/",
					Schedule: DependabotSchedule{
						Interval: IntervalWeekly,
					},
				},
			},
		},
		EcosystemPolicies: map[string]EcosystemPolicy{
			EcosystemGoModules: {
				Ecosystem: EcosystemGoModules,
				Enabled:   true,
			},
			EcosystemNPM: {
				Ecosystem: EcosystemNPM,
				Enabled:   true,
			},
		},
	}

	// Test config that complies with policy
	compliantConfig := &DependabotConfig{
		Version: 2,
		Updates: []DependabotUpdateRule{
			{
				PackageEcosystem: EcosystemGoModules,
				Directory:        "/",
				Schedule: DependabotSchedule{
					Interval: IntervalWeekly,
				},
			},
			{
				PackageEcosystem: EcosystemNPM,
				Directory:        "/frontend",
				Schedule: DependabotSchedule{
					Interval: IntervalDaily,
				},
			},
		},
	}

	compliantStatus := &DependabotStatus{
		Enabled:      true,
		ConfigValid:  true,
		ConfigExists: true,
	}

	result := policyManager.performPolicyEvaluation(policy, compliantConfig, compliantStatus, "testorg", "testrepo")
	assert.True(t, result.Compliant)
	assert.Empty(t, result.Violations)
	assert.Equal(t, 100.0, result.ComplianceScore)

	// Test non-compliant config
	nonCompliantStatus := &DependabotStatus{
		Enabled:      false, // Dependabot disabled
		ConfigValid:  false,
		ConfigExists: true,
	}

	result = policyManager.performPolicyEvaluation(policy, compliantConfig, nonCompliantStatus, "testorg", "testrepo")
	assert.False(t, result.Compliant)
	assert.NotEmpty(t, result.Violations)
	assert.Less(t, result.ComplianceScore, 100.0)

	// Check violation details
	hasDisabledViolation := false
	hasInvalidConfigViolation := false

	for _, violation := range result.Violations {
		if violation.Type == DependabotViolationTypeMissingConfig {
			hasDisabledViolation = true
		}

		if violation.Type == DependabotViolationTypeInvalidConfig {
			hasInvalidConfigViolation = true
		}
	}

	assert.True(t, hasDisabledViolation)
	assert.True(t, hasInvalidConfigViolation)
}

func TestDependabotPolicyManager_Cache(t *testing.T) {
	policyManager := createTestPolicyManager()

	// Test cache miss
	result := policyManager.getCachedResult("non-existent-key")
	assert.Nil(t, result)

	// Test cache hit
	testResult := &PolicyEvaluationResult{
		PolicyID:       "test-policy",
		Repository:     "testrepo",
		Organization:   "testorg",
		Compliant:      true,
		EvaluatedAt:    time.Now(),
		NextEvaluation: time.Now().Add(time.Hour),
	}

	policyManager.cacheResult("test-key", testResult)
	cached := policyManager.getCachedResult("test-key")
	assert.NotNil(t, cached)
	assert.Equal(t, testResult.PolicyID, cached.PolicyID)
	assert.Equal(t, testResult.Repository, cached.Repository)

	// Test cache expiration
	expiredResult := &PolicyEvaluationResult{
		PolicyID:       "expired-policy",
		Repository:     "expiredrepo",
		Organization:   "testorg",
		Compliant:      true,
		EvaluatedAt:    time.Now().Add(-2 * time.Hour),
		NextEvaluation: time.Now().Add(-time.Hour), // Expired
	}

	policyManager.cacheResult("expired-key", expiredResult)
	cached = policyManager.getCachedResult("expired-key")
	assert.Nil(t, cached) // Should be nil due to expiration

	// Test cache invalidation
	policyManager.cacheResult("org-key", testResult)
	cached = policyManager.getCachedResult("org-key")
	assert.NotNil(t, cached)

	policyManager.invalidateCacheForOrganization("testorg")
	cached = policyManager.getCachedResult("org-key")
	assert.Nil(t, cached) // Should be invalidated
}

// Benchmark tests.
func BenchmarkEvaluateRepositoryCompliance(b *testing.B) {
	policyManager := createTestPolicyManager()
	ctx := context.Background()

	policy := createTestDependabotPolicy()
	policyManager.CreatePolicy(ctx, policy)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		policyManager.EvaluateRepositoryCompliance(ctx, policy.ID, "testorg", "testrepo")
	}
}

func BenchmarkPolicyEvaluation(b *testing.B) {
	policyManager := createTestPolicyManager()

	policy := createTestDependabotPolicy()
	config := &DependabotConfig{
		Version: 2,
		Updates: []DependabotUpdateRule{
			{
				PackageEcosystem: EcosystemGoModules,
				Directory:        "/",
				Schedule: DependabotSchedule{
					Interval: IntervalWeekly,
				},
			},
		},
	}
	status := &DependabotStatus{
		Enabled:      true,
		ConfigValid:  true,
		ConfigExists: true,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		policyManager.performPolicyEvaluation(policy, config, status, "testorg", "testrepo")
	}
}

// Helper functions.
func createTestPolicyManager() *DependabotPolicyManager {
	logger := &simpleLogger{}
	apiClient := &simpleAPIClient{}
	configManager := createTestDependabotManager()

	return NewDependabotPolicyManager(logger, apiClient, configManager)
}

func createTestDependabotPolicy() *DependabotPolicyConfig {
	return &DependabotPolicyConfig{
		ID:           "test-policy-1",
		Name:         "Test Dependabot Policy",
		Organization: "testorg",
		Description:  "Test policy for unit testing",
		Enabled:      true,
		DefaultConfig: DependabotConfig{
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
				},
			},
		},
		EcosystemPolicies: map[string]EcosystemPolicy{
			EcosystemGoModules: {
				Ecosystem:             EcosystemGoModules,
				Enabled:               true,
				RequiredReviewers:     1,
				AllowedUpdateTypes:    []string{UpdateTypeAll},
				MaxPullRequestsPerDay: 10,
				AutoMergeEnabled:      false,
				RequiredStatusChecks:  []string{"ci/test"},
				MinSecuritySeverity:   "medium",
			},
		},
		SecurityPolicies: SecurityPolicySettings{
			EnableVulnerabilityAlerts: true,
			AutoFixSecurityVulns:      true,
			SecurityReviewRequired:    true,
			CriticalVulnAutoMerge:     false,
		},
		ApprovalRequirements: ApprovalRequirements{
			SecurityUpdates: ApprovalRule{
				RequiredReviewers: 1,
				RequiredApprovals: 1,
			},
			MajorUpdates: ApprovalRule{
				RequiredReviewers:      2,
				RequiredApprovals:      2,
				RequireCodeOwnerReview: true,
			},
		},
	}
}
