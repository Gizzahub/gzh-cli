//nolint:testpackage // White-box testing needed for internal function access
package github

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSecurityUpdatePolicyManager(t *testing.T) {
	logger := &simpleLogger{}
	apiClient := &simpleAPIClient{}
	dependabotManager := createTestDependabotManager()

	manager := NewSecurityUpdatePolicyManager(logger, apiClient, dependabotManager)

	assert.NotNil(t, manager)
	assert.Equal(t, logger, manager.logger)
	assert.Equal(t, apiClient, manager.apiClient)
	assert.Equal(t, dependabotManager, manager.dependabotManager)
	assert.NotNil(t, manager.policies)
	assert.NotNil(t, manager.vulnerabilityDB)
}

func TestNewVulnerabilityDatabase(t *testing.T) {
	db := NewVulnerabilityDatabase()

	assert.NotNil(t, db)
	assert.NotNil(t, db.vulnerabilities)
	assert.NotNil(t, db.cveCache)
	assert.False(t, db.lastUpdated.IsZero())
}

func TestSecurityUpdatePolicyManager_CreateSecurityPolicy(t *testing.T) {
	manager := createTestSecurityPolicyManager()
	ctx := context.Background()

	tests := []struct {
		name        string
		policy      *SecurityUpdatePolicy
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid security policy",
			policy: &SecurityUpdatePolicy{
				ID:           "test-security-policy",
				Name:         "Test Security Policy",
				Organization: "testorg",
				Description:  "Test security policy for unit tests",
				Enabled:      true,
				AutoApprovalRules: []AutoApprovalRule{
					{
						ID:          "rule-1",
						Name:        "Auto-approve low severity",
						Enabled:     true,
						MaxSeverity: VulnSeverityLow,
						Conditions: []ApprovalCondition{
							{
								Type:     ConditionTypeSeverity,
								Field:    "severity",
								Operator: "lte",
								Value:    "low",
							},
						},
						Actions: []AutoApprovalAction{
							{
								Type: ActionTypeSecurityApprove,
							},
						},
					},
				},
				SeverityThresholds: SeverityThresholdConfig{
					Critical: SeverityThreshold{
						AutoApprove:         false,
						RequireManualReview: true,
						MaxResponseTime:     2 * time.Hour,
						RequiredApprovers:   2,
						NotifyImmediately:   true,
					},
					Low: SeverityThreshold{
						AutoApprove:       true,
						MaxResponseTime:   24 * time.Hour,
						RequiredApprovers: 0,
					},
				},
			},
			expectError: false,
		},
		{
			name: "missing policy ID",
			policy: &SecurityUpdatePolicy{
				Name:         "Test Policy",
				Organization: "testorg",
				Enabled:      true,
			},
			expectError: true,
			errorMsg:    "policy ID is required",
		},
		{
			name: "missing organization",
			policy: &SecurityUpdatePolicy{
				ID:      "test-policy",
				Name:    "Test Policy",
				Enabled: true,
			},
			expectError: true,
			errorMsg:    "organization is required",
		},
		{
			name: "invalid auto-approval rule",
			policy: &SecurityUpdatePolicy{
				ID:           "test-policy",
				Name:         "Test Policy",
				Organization: "testorg",
				Enabled:      true,
				AutoApprovalRules: []AutoApprovalRule{
					{
						Name:    "Invalid Rule",
						Enabled: true,
						// Missing ID and Conditions
					},
				},
			},
			expectError: true,
			errorMsg:    "ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.CreateSecurityPolicy(ctx, tt.policy)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
				assert.NotZero(t, tt.policy.CreatedAt)
				assert.NotZero(t, tt.policy.UpdatedAt)
				assert.Equal(t, 1, tt.policy.Version)

				// Verify policy was stored
				assert.Contains(t, manager.policies, tt.policy.ID)
			}
		})
	}
}

func TestSecurityUpdatePolicyManager_EvaluateSecurityUpdate(t *testing.T) {
	manager := createTestSecurityPolicyManager()
	ctx := context.Background()

	// Create a test policy
	policy := createTestSecurityPolicy()
	err := manager.CreateSecurityPolicy(ctx, policy)
	require.NoError(t, err)

	tests := []struct {
		name           string
		update         *SecurityUpdateStatus
		expectApproved bool
		expectReason   string
	}{
		{
			name: "low severity auto-approved",
			update: &SecurityUpdateStatus{
				UpdateID:        "update-1",
				VulnerabilityID: "vuln-low",
				Repository:      "test-repo",
				Organization:    "testorg",
				Package: PackageInfo{
					Name:      "test-package",
					Ecosystem: "npm",
				},
				Status: UpdateStatusPending,
			},
			expectApproved: true,
			expectReason:   "Auto-approved by rule",
		},
		{
			name: "excluded package",
			update: &SecurityUpdateStatus{
				UpdateID:        "update-2",
				VulnerabilityID: "vuln-excluded",
				Repository:      "test-repo",
				Organization:    "testorg",
				Package: PackageInfo{
					Name:      "excluded-package",
					Ecosystem: "npm",
				},
				Status: UpdateStatusPending,
			},
			expectApproved: false,
			expectReason:   "Update matches exclusion rule",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decision, err := manager.EvaluateSecurityUpdate(ctx, policy.ID, tt.update)
			require.NoError(t, err)
			require.NotNil(t, decision)

			assert.Equal(t, tt.expectApproved, decision.Approved)
			assert.Contains(t, decision.Reason, tt.expectReason)
		})
	}
}

func TestSecurityUpdatePolicyManager_ProcessSecurityUpdates(t *testing.T) {
	manager := createTestSecurityPolicyManager()
	ctx := context.Background()

	// Create a test policy
	policy := createTestSecurityPolicy()
	err := manager.CreateSecurityPolicy(ctx, policy)
	require.NoError(t, err)

	// Process security updates
	result, err := manager.ProcessSecurityUpdates(ctx, "testorg")
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "testorg", result.Organization)
	assert.GreaterOrEqual(t, result.TotalUpdates, 0)
	assert.NotZero(t, result.StartedAt)
	assert.NotZero(t, result.CompletedAt)
	assert.Greater(t, result.ProcessingTime, time.Duration(0))
}

func TestVulnerabilitySeverityConstants(t *testing.T) {
	severities := []VulnerabilitySeverity{
		VulnSeverityCritical, VulnSeverityHigh, VulnSeverityMedium, VulnSeverityLow, VulnSeverityInfo,
	}

	for _, severity := range severities {
		assert.NotEmpty(t, string(severity))
	}

	// Test specific values
	assert.Equal(t, VulnerabilitySeverity("critical"), SeverityCritical)
	assert.Equal(t, VulnerabilitySeverity("high"), SeverityHigh)
	assert.Equal(t, VulnerabilitySeverity("medium"), SeverityMedium)
	assert.Equal(t, VulnerabilitySeverity("low"), SeverityLow)
	assert.Equal(t, VulnerabilitySeverity("info"), SeverityInfo)
}

func TestConditionTypeConstants(t *testing.T) {
	conditionTypes := []ConditionType{
		ConditionTypeSeverity, ConditionTypePackage, ConditionTypeVersion,
		ConditionTypeCVSS, ConditionTypeAge, ConditionTypeRepository,
		ConditionTypeEcosystem,
	}

	for _, condType := range conditionTypes {
		assert.NotEmpty(t, string(condType))
	}

	// Test specific values
	assert.Equal(t, ConditionType("severity"), ConditionTypeSeverity)
	assert.Equal(t, ConditionType("package"), ConditionTypePackage)
	assert.Equal(t, ConditionType("cvss"), ConditionTypeCVSS)
}

func TestActionTypeConstants(t *testing.T) {
	actionTypes := []ActionType{
		ActionTypeSecurityApprove, ActionTypeSecurityMerge, ActionTypeSecurityNotify,
		ActionTypeSecurityTest, ActionTypeSecurityCreateTicket, ActionTypeSecuritySchedule,
	}

	for _, actionType := range actionTypes {
		assert.NotEmpty(t, string(actionType))
	}

	// Test specific values
	assert.Equal(t, ActionType("security_approve"), ActionTypeSecurityApprove)
	assert.Equal(t, ActionType("security_merge"), ActionTypeSecurityMerge)
	assert.Equal(t, ActionType("security_notify"), ActionTypeSecurityNotify)
}

func TestUpdateStatusConstants(t *testing.T) {
	statuses := []UpdateStatus{
		UpdateStatusPending, UpdateStatusReviewing, UpdateStatusApproved,
		UpdateStatusRejected, UpdateStatusTesting, UpdateStatusDeploying,
		UpdateStatusCompleted, UpdateStatusFailed, UpdateStatusCancelled,
	}

	for _, status := range statuses {
		assert.NotEmpty(t, string(status))
	}

	// Test specific values
	assert.Equal(t, UpdateStatus("pending"), UpdateStatusPending)
	assert.Equal(t, UpdateStatus("approved"), UpdateStatusApproved)
	assert.Equal(t, UpdateStatus("completed"), UpdateStatusCompleted)
}

func TestSecurityUpdatePolicyManager_SeverityComparison(t *testing.T) {
	manager := createTestSecurityPolicyManager()

	tests := []struct {
		name     string
		actual   VulnerabilitySeverity
		maximum  VulnerabilitySeverity
		expected bool
	}{
		{
			name:     "critical exceeds high",
			actual:   VulnSeverityCritical,
			maximum:  VulnSeverityHigh,
			expected: true,
		},
		{
			name:     "medium does not exceed high",
			actual:   VulnSeverityMedium,
			maximum:  VulnSeverityHigh,
			expected: false,
		},
		{
			name:     "high equals high",
			actual:   VulnSeverityHigh,
			maximum:  VulnSeverityHigh,
			expected: false,
		},
		{
			name:     "low does not exceed critical",
			actual:   VulnSeverityLow,
			maximum:  VulnSeverityCritical,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.severityExceeds(tt.actual, tt.maximum)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSecurityUpdatePolicyManager_ConditionEvaluation(t *testing.T) {
	manager := createTestSecurityPolicyManager()

	vuln := &VulnerabilityRecord{
		ID:       "test-vuln",
		Severity: VulnSeverityMedium,
		CVSS: CVSSScore{
			Score: 6.5,
		},
	}

	update := &SecurityUpdateStatus{
		Package: PackageInfo{
			Name:      "test-package",
			Ecosystem: "npm",
		},
	}

	tests := []struct {
		name      string
		condition ApprovalCondition
		expected  bool
	}{
		{
			name: "severity equals medium",
			condition: ApprovalCondition{
				Type:     ConditionTypeSeverity,
				Field:    "severity",
				Operator: "eq",
				Value:    "medium",
			},
			expected: true,
		},
		{
			name: "severity less than or equal to high",
			condition: ApprovalCondition{
				Type:     ConditionTypeSeverity,
				Field:    "severity",
				Operator: "lte",
				Value:    "high",
			},
			expected: true,
		},
		{
			name: "package name equals test-package",
			condition: ApprovalCondition{
				Type:     ConditionTypePackage,
				Field:    "name",
				Operator: "eq",
				Value:    "test-package",
			},
			expected: true,
		},
		{
			name: "cvss score greater than 6.0",
			condition: ApprovalCondition{
				Type:     ConditionTypeCVSS,
				Field:    "score",
				Operator: "gt",
				Value:    6.0,
			},
			expected: true,
		},
		{
			name: "negated condition",
			condition: ApprovalCondition{
				Type:     ConditionTypePackage,
				Field:    "name",
				Operator: "eq",
				Value:    "different-package",
				Negated:  true,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.evaluateCondition(tt.condition, vuln, update)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSecurityUpdatePolicyManager_ExclusionRules(t *testing.T) {
	manager := createTestSecurityPolicyManager()

	policy := &SecurityUpdatePolicy{
		ExclusionRules: []VulnerabilityExclusion{
			{
				ID:      "exclude-1",
				Type:    ExclusionTypePackage,
				Pattern: "excluded-package",
				Reason:  "Testing exclusion",
			},
			{
				ID:        "exclude-2",
				Type:      ExclusionTypeCVE,
				Pattern:   "CVE-2024-1234",
				Reason:    "False positive",
				ExpiresAt: &[]time.Time{time.Now().Add(-1 * time.Hour)}[0], // Expired
			},
		},
	}

	vuln := &VulnerabilityRecord{
		CVE: "CVE-2024-1234",
	}

	tests := []struct {
		name     string
		update   *SecurityUpdateStatus
		expected bool
	}{
		{
			name: "excluded package",
			update: &SecurityUpdateStatus{
				Package: PackageInfo{
					Name: "excluded-package",
				},
			},
			expected: true,
		},
		{
			name: "expired CVE exclusion",
			update: &SecurityUpdateStatus{
				Package: PackageInfo{
					Name: "normal-package",
				},
			},
			expected: false, // Exclusion is expired
		},
		{
			name: "non-excluded package",
			update: &SecurityUpdateStatus{
				Package: PackageInfo{
					Name: "normal-package",
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.isExcluded(policy, vuln, tt.update)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSecurityUpdatePolicyManager_StringComparison(t *testing.T) {
	manager := createTestSecurityPolicyManager()

	tests := []struct {
		name     string
		actual   string
		operator string
		expected interface{}
		result   bool
	}{
		{
			name:     "equals match",
			actual:   "test-package",
			operator: "eq",
			expected: "test-package",
			result:   true,
		},
		{
			name:     "contains match",
			actual:   "test-package-name",
			operator: "contains",
			expected: "package",
			result:   true,
		},
		{
			name:     "starts with match",
			actual:   "test-package",
			operator: "starts_with",
			expected: "test",
			result:   true,
		},
		{
			name:     "ends with match",
			actual:   "test-package",
			operator: "ends_with",
			expected: "package",
			result:   true,
		},
		{
			name:     "no match",
			actual:   "test-package",
			operator: "eq",
			expected: "different",
			result:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.compareString(tt.actual, tt.operator, tt.expected)
			assert.Equal(t, tt.result, result)
		})
	}
}

func TestSecurityUpdatePolicyManager_FloatComparison(t *testing.T) {
	manager := createTestSecurityPolicyManager()

	tests := []struct {
		name     string
		actual   float64
		operator string
		expected interface{}
		result   bool
	}{
		{
			name:     "equals float",
			actual:   6.5,
			operator: "eq",
			expected: 6.5,
			result:   true,
		},
		{
			name:     "greater than",
			actual:   7.0,
			operator: "gt",
			expected: 6.5,
			result:   true,
		},
		{
			name:     "less than or equal",
			actual:   6.5,
			operator: "lte",
			expected: 6.5,
			result:   true,
		},
		{
			name:     "equals int conversion",
			actual:   6.0,
			operator: "eq",
			expected: 6,
			result:   true,
		},
		{
			name:     "less than",
			actual:   5.0,
			operator: "lt",
			expected: 6.0,
			result:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.compareFloat(tt.actual, tt.operator, tt.expected)
			assert.Equal(t, tt.result, result)
		})
	}
}

// Benchmark tests.
func BenchmarkEvaluateSecurityUpdate(b *testing.B) {
	manager := createTestSecurityPolicyManager()
	ctx := context.Background()

	policy := createTestSecurityPolicy()
	manager.CreateSecurityPolicy(ctx, policy)

	update := &SecurityUpdateStatus{
		UpdateID:        "bench-update",
		VulnerabilityID: "bench-vuln",
		Repository:      "bench-repo",
		Organization:    "testorg",
		Package: PackageInfo{
			Name:      "bench-package",
			Ecosystem: "npm",
		},
		Status: UpdateStatusPending,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		manager.EvaluateSecurityUpdate(ctx, policy.ID, update)
	}
}

func BenchmarkConditionEvaluation(b *testing.B) {
	manager := createTestSecurityPolicyManager()

	vuln := &VulnerabilityRecord{
		Severity: VulnSeverityMedium,
		CVSS:     CVSSScore{Score: 6.5},
	}

	update := &SecurityUpdateStatus{
		Package: PackageInfo{Name: "test-package"},
	}

	condition := ApprovalCondition{
		Type:     ConditionTypeSeverity,
		Field:    "severity",
		Operator: "lte",
		Value:    "high",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		manager.evaluateCondition(condition, vuln, update)
	}
}

// Helper functions.
func createTestSecurityPolicyManager() *SecurityUpdatePolicyManager {
	logger := &simpleLogger{}
	apiClient := &simpleAPIClient{}
	dependabotManager := createTestDependabotManager()

	return NewSecurityUpdatePolicyManager(logger, apiClient, dependabotManager)
}

func createTestSecurityPolicy() *SecurityUpdatePolicy {
	return &SecurityUpdatePolicy{
		ID:           "test-security-policy-1",
		Name:         "Test Security Policy",
		Organization: "testorg",
		Description:  "Test security policy for unit testing",
		Enabled:      true,
		AutoApprovalRules: []AutoApprovalRule{
			{
				ID:          "rule-1",
				Name:        "Auto-approve low severity",
				Enabled:     true,
				MaxSeverity: VulnSeverityMedium,
				Conditions: []ApprovalCondition{
					{
						Type:     ConditionTypeSeverity,
						Field:    "severity",
						Operator: "lte",
						Value:    "medium",
					},
				},
				Actions: []AutoApprovalAction{
					{
						Type: ActionTypeSecurityApprove,
					},
				},
				TestingRequired: false,
			},
		},
		SeverityThresholds: SeverityThresholdConfig{
			Critical: SeverityThreshold{
				AutoApprove:         false,
				RequireManualReview: true,
				MaxResponseTime:     2 * time.Hour,
				RequiredApprovers:   2,
				NotifyImmediately:   true,
			},
			Medium: SeverityThreshold{
				AutoApprove:       true,
				MaxResponseTime:   24 * time.Hour,
				RequiredApprovers: 0,
			},
			Low: SeverityThreshold{
				AutoApprove:       true,
				MaxResponseTime:   72 * time.Hour,
				RequiredApprovers: 0,
			},
		},
		ResponseTimeRequirements: ResponseTimeConfig{
			CriticalVulnerabilities: 2 * time.Hour,
			HighVulnerabilities:     24 * time.Hour,
			MediumVulnerabilities:   72 * time.Hour,
			LowVulnerabilities:      7 * 24 * time.Hour,
		},
		NotificationSettings: NotificationConfig{
			Enabled: true,
			Channels: []NotificationChannel{
				{
					Type:       ChannelTypeEmail,
					Target:     "security@example.com",
					Enabled:    true,
					Severities: []VulnerabilitySeverity{VulnSeverityCritical, VulnSeverityHigh},
				},
			},
		},
		ExclusionRules: []VulnerabilityExclusion{
			{
				ID:      "exclude-test",
				Type:    ExclusionTypePackage,
				Pattern: "excluded-package",
				Reason:  "Test exclusion",
			},
		},
		ComplianceSettings: ComplianceConfig{
			AuditTrailRequired:    true,
			DocumentationRequired: true,
			RetentionPeriod:       365 * 24 * time.Hour,
		},
	}
}
