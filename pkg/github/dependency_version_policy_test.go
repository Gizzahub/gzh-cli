//nolint:testpackage // White-box testing needed for internal function access
package github

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDependencyVersionPolicyManager(t *testing.T) {
	logger := &simpleLogger{}
	apiClient := &simpleAPIClient{}
	dependabotManager := createTestDependabotManager()
	securityPolicyManager := createTestSecurityPolicyManager()

	manager := NewDependencyVersionPolicyManager(logger, apiClient, dependabotManager, securityPolicyManager)

	assert.NotNil(t, manager)
	assert.Equal(t, logger, manager.logger)
	assert.Equal(t, apiClient, manager.apiClient)
	assert.Equal(t, dependabotManager, manager.dependabotManager)
	assert.Equal(t, securityPolicyManager, manager.securityPolicyManager)
	assert.NotNil(t, manager.policies)
	assert.NotNil(t, manager.versionConstraints)
}

func TestNewVersionConstraintEngine(t *testing.T) {
	logger := &simpleLogger{}
	engine := NewVersionConstraintEngine(logger)

	assert.NotNil(t, engine)
	assert.Equal(t, logger, engine.logger)
	assert.NotNil(t, engine.constraintCache)
	assert.Equal(t, time.Hour, engine.cacheTTL)
}

func TestDependencyVersionPolicyManager_CreateDependencyVersionPolicy(t *testing.T) {
	manager := createTestDependencyVersionPolicyManager()
	ctx := context.Background()

	tests := []struct {
		name        string
		policy      *DependencyVersionPolicy
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid dependency version policy",
			policy: &DependencyVersionPolicy{
				ID:           "test-version-policy",
				Name:         "Test Version Policy",
				Organization: "testorg",
				Description:  "Test dependency version policy for unit tests",
				Enabled:      true,
				VersionConstraints: map[string]VersionConstraintRule{
					"npm-lodash": {
						RuleID:            "npm-lodash-constraint",
						DependencyPattern: "lodash",
						Ecosystem:         "npm",
						MinimumVersion:    "4.17.20",
						AllowPrerelease:   false,
						UpdateStrategy:    UpdateStrategyConservative,
						AutoUpdateEnabled: true,
						UpdateFrequency:   UpdateFrequencyWeekly,
						Priority:          ConstraintPriorityMedium,
						Justification:     "Security and stability requirements",
					},
				},
				EcosystemPolicies: map[string]EcosystemVersionPolicy{
					"npm": {
						Ecosystem:              "npm",
						Enabled:                true,
						DefaultUpdateStrategy:  UpdateStrategyModerate,
						AllowMajorUpdates:      false,
						AllowMinorUpdates:      true,
						AllowPatchUpdates:      true,
						RequireSecurityUpdates: true,
						MaxVersionAge:          180 * 24 * time.Hour,
						DeprecationPolicy: DeprecationPolicy{
							AllowDeprecatedVersions:  false,
							DeprecationWarningPeriod: 30 * 24 * time.Hour,
							ForceUpgradeAfterEOL:     true,
							EOLNotificationPeriod:    60 * 24 * time.Hour,
						},
						LicenseRestrictions: []LicenseRestriction{
							{
								BlockedLicenses:  []string{"GPL-3.0", "AGPL-3.0"},
								RequiredLicenses: []string{"MIT", "Apache-2.0", "BSD-3-Clause"},
							},
						},
						PerformanceRequirements: PerformanceRequirements{
							MaxPerformanceRegression: 0.05,
							BenchmarkSuites:          []string{"npm-benchmark", "lighthouse"},
							PerformanceThresholds: map[string]float64{
								"bundle_size": 1000000, // 1MB
								"load_time":   2000,    // 2 seconds
							},
						},
					},
				},
				BreakingChangePolicy: BreakingChangePolicy{
					AllowBreakingChanges:        false,
					ImpactAnalysisRequired:      true,
					DeprecationNoticePeriod:     90 * 24 * time.Hour,
					MigrationGuidanceRequired:   true,
					BackwardCompatibilityPeriod: 365 * 24 * time.Hour,
					BreakingChangeApprovers:     []string{"architecture-team", "senior-developers"},
					BreakingChangeDetection: BreakingChangeDetection{
						Enabled:               true,
						Methods:               []DetectionMethod{DetectionMethodSemver, DetectionMethodAPI},
						SemverStrictMode:      true,
						APIChangeDetection:    true,
						SchemaChangeDetection: false,
						CustomDetectionRules: []DetectionRule{
							{
								Pattern:     "breaking.*change",
								Severity:    "high",
								Description: "Breaking change detected in changelog",
								Weight:      0.8,
							},
						},
						ThresholdConfiguration: ThresholdConfig{
							MinorChangeThreshold:    0.3,
							MajorChangeThreshold:    0.7,
							BreakingChangeThreshold: 0.9,
						},
					},
				},
				CompatibilityChecks: CompatibilityCheckConfig{
					Enabled:                   true,
					DependencyGraphAnalysis:   true,
					PerformanceImpactAnalysis: true,
					SecurityImpactAnalysis:    true,
					MatrixTesting: MatrixTestingConfig{
						Enabled:          true,
						OperatingSystems: []string{"linux", "windows", "macos"},
						RuntimeVersions:  []string{"node-16", "node-18", "node-20"},
					},
					ConflictDetection: ConflictDetectionConfig{
						Enabled:                       true,
						CheckTransitiveDependencies:   true,
						ResolveConflictsAutomatically: false,
						ConflictResolutionStrategy:    "manual",
					},
					IntegrationTesting: IntegrationTestingConfig{
						Enabled:          true,
						TestSuites:       []string{"integration", "e2e"},
						RequiredCoverage: 80.0,
						Timeout:          30 * time.Minute,
						Environment:      "staging",
					},
				},
				ApprovalRequirements: VersionUpdateApprovalRequirements{
					MajorVersionUpdates: VersionApprovalRule{
						RequiredApprovers:          2,
						RequiredApprovalTeams:      []string{"architecture-team"},
						ManualReviewRequired:       true,
						SecurityReviewRequired:     true,
						ArchitectureReviewRequired: true,
						TestingGateRequired:        true,
						WaitingPeriod:              24 * time.Hour,
						ApprovalTimeLimit:          7 * 24 * time.Hour,
					},
					MinorVersionUpdates: VersionApprovalRule{
						RequiredApprovers:      1,
						RequiredApprovalTeams:  []string{"development-team"},
						ManualReviewRequired:   false,
						SecurityReviewRequired: true,
						TestingGateRequired:    true,
						WaitingPeriod:          2 * time.Hour,
						ApprovalTimeLimit:      2 * 24 * time.Hour,
					},
					PatchVersionUpdates: VersionApprovalRule{
						RequiredApprovers:      0,
						ManualReviewRequired:   false,
						SecurityReviewRequired: false,
						TestingGateRequired:    true,
						WaitingPeriod:          0,
						ApprovalTimeLimit:      24 * time.Hour,
						AutoApprovalConditions: []AutoApprovalCondition{
							{
								Type:     "security_improvement",
								Field:    "has_security_fixes",
								Operator: "eq",
								Value:    true,
								Required: false,
							},
						},
					},
					SecurityUpdates: VersionApprovalRule{
						RequiredApprovers:      1,
						RequiredApprovalTeams:  []string{"security-team"},
						ManualReviewRequired:   false,
						SecurityReviewRequired: true,
						TestingGateRequired:    false,
						WaitingPeriod:          0,
						ApprovalTimeLimit:      4 * time.Hour,
					},
				},
				TestingRequirements: TestingRequirements{
					Enabled:                    true,
					UnitTestingRequired:        true,
					IntegrationTestingRequired: true,
					E2ETestingRequired:         false,
					PerformanceTestingRequired: true,
					SecurityTestingRequired:    true,
					MinimumTestCoverage:        85.0,
					TestSuiteConfiguration: TestSuiteConfiguration{
						DefaultSuites: []string{"unit", "integration", "performance"},
						EcosystemSpecific: map[string][]string{
							"npm": {"npm-test", "jest"},
							"go":  {"go-test", "benchstat"},
						},
					},
					AutomatedTesting: AutomatedTestingConfig{
						Enabled:               true,
						TriggerOnUpdate:       true,
						ParallelExecution:     true,
						MaxConcurrentTests:    5,
						TestEnvironments:      []string{"test", "staging"},
						NotificationOnFailure: true,
						AutoRetryOnFailure:    true,
						MaxRetries:            3,
						TestResultsRetention:  30 * 24 * time.Hour,
					},
				},
				ReleaseWindows: []ReleaseWindow{
					{
						ID:          "weekly-maintenance",
						Name:        "Weekly Maintenance Window",
						Description: "Regular weekly maintenance window for updates",
						Enabled:     true,
						Schedule: ReleaseSchedule{
							Type:       "weekly",
							DaysOfWeek: []string{"tuesday"},
							TimeOfDay:  "02:00",
							Timezone:   "UTC",
							Frequency:  "weekly",
							Duration:   4 * time.Hour,
						},
						AllowedUpdateTypes:   []string{"patch", "minor"},
						RestrictedEcosystems: []string{},
						ApprovalRequired:     false,
						BlackoutPeriods: []BlackoutPeriod{
							{
								Name:        "Holiday Blackout",
								StartDate:   time.Date(2024, 12, 20, 0, 0, 0, 0, time.UTC),
								EndDate:     time.Date(2025, 1, 5, 0, 0, 0, 0, time.UTC),
								Recurring:   true,
								Description: "Holiday season blackout period",
							},
						},
						EmergencyOverride: EmergencyOverride{
							Enabled:               true,
							AuthorizedUsers:       []string{"ops-team", "security-team"},
							JustificationRequired: true,
							AuditTrail:            true,
							PostEmergencyReview:   true,
						},
					},
				},
				NotificationSettings: VersionPolicyNotificationConfig{
					Enabled:    true,
					EventTypes: []string{"policy_violation", "approval_required", "update_approved"},
					Channels: []VersionNotificationChannel{
						{
							Type:        "email",
							Target:      "dev-team@example.com",
							Enabled:     true,
							EventFilter: []string{"policy_violation", "approval_required"},
						},
						{
							Type:        "slack",
							Target:      "#dev-alerts",
							Enabled:     true,
							EventFilter: []string{"update_approved", "policy_violation"},
						},
					},
					Recipients: []NotificationRecipient{
						{
							Type:       "team",
							Identifier: "development-team",
							EventTypes: []string{"policy_violation", "approval_required"},
							Active:     true,
						},
					},
				},
				MetricsTracking: MetricsTrackingConfig{
					Enabled:           true,
					MetricsCollectors: []string{"prometheus", "datadog"},
					TrackingFrequency: time.Hour,
					RetentionPeriod:   90 * 24 * time.Hour,
					AlertingEnabled:   true,
					DashboardEnabled:  true,
					CustomMetrics: []CustomMetric{
						{
							Name:      "dependency_update_success_rate",
							Type:      "gauge",
							Query:     "sum(dependency_updates_successful) / sum(dependency_updates_total)",
							Threshold: 0.95,
							Alerting:  true,
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "missing policy ID",
			policy: &DependencyVersionPolicy{
				Name:         "Test Policy",
				Organization: "testorg",
				Enabled:      true,
			},
			expectError: true,
			errorMsg:    "policy ID is required",
		},
		{
			name: "missing organization",
			policy: &DependencyVersionPolicy{
				ID:      "test-policy",
				Name:    "Test Policy",
				Enabled: true,
			},
			expectError: true,
			errorMsg:    "organization is required",
		},
		{
			name: "invalid version constraint",
			policy: &DependencyVersionPolicy{
				ID:           "test-policy",
				Name:         "Test Policy",
				Organization: "testorg",
				Enabled:      true,
				VersionConstraints: map[string]VersionConstraintRule{
					"invalid": {
						// Missing required fields
						DependencyPattern: "test",
					},
				},
			},
			expectError: true,
			errorMsg:    "rule ID is required",
		},
		{
			name: "invalid ecosystem policy",
			policy: &DependencyVersionPolicy{
				ID:           "test-policy",
				Name:         "Test Policy",
				Organization: "testorg",
				Enabled:      true,
				EcosystemPolicies: map[string]EcosystemVersionPolicy{
					"npm": {
						// Missing ecosystem field
						Enabled: true,
						PerformanceRequirements: PerformanceRequirements{
							MaxPerformanceRegression: 1.5, // Invalid: > 1.0
						},
					},
				},
			},
			expectError: true,
			errorMsg:    "ecosystem is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.CreateDependencyVersionPolicy(ctx, tt.policy)
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

func TestDependencyVersionPolicyManager_AnalyzeDependencyVersionUpdate(t *testing.T) {
	manager := createTestDependencyVersionPolicyManager()
	ctx := context.Background()

	// Create a test policy
	policy := createTestDependencyVersionPolicy()
	err := manager.CreateDependencyVersionPolicy(ctx, policy)
	require.NoError(t, err)

	tests := []struct {
		name            string
		dependencyName  string
		currentVersion  string
		proposedVersion string
		ecosystem       string
		expectedAllowed bool
		expectedAction  string
		expectedReason  string
	}{
		{
			name:            "allowed patch update",
			dependencyName:  "lodash",
			currentVersion:  "4.17.20",
			proposedVersion: "4.17.21",
			ecosystem:       "npm",
			expectedAllowed: true,
			expectedAction:  "approve",
			expectedReason:  "Security improvements",
		},
		{
			name:            "blocked major update",
			dependencyName:  "lodash",
			currentVersion:  "4.17.21",
			proposedVersion: "5.0.0",
			ecosystem:       "npm",
			expectedAllowed: false,
			expectedAction:  "reject",
			expectedReason:  "Major updates not allowed",
		},
		{
			name:            "prerelease blocked",
			dependencyName:  "express",
			currentVersion:  "4.18.1",
			proposedVersion: "4.18.2-beta.1",
			ecosystem:       "npm",
			expectedAllowed: false,
			expectedAction:  "reject",
			expectedReason:  "Prerelease versions not allowed",
		},
		{
			name:            "minor update requires review",
			dependencyName:  "react",
			currentVersion:  "17.0.2",
			proposedVersion: "17.1.0",
			ecosystem:       "npm",
			expectedAllowed: true,
			expectedAction:  "review",
			expectedReason:  "Manual review required",
		},
		{
			name:            "version below minimum",
			dependencyName:  "lodash",
			currentVersion:  "4.17.19",
			proposedVersion: "4.17.15",
			ecosystem:       "npm",
			expectedAllowed: false,
			expectedAction:  "reject",
			expectedReason:  "below minimum",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysis, err := manager.AnalyzeDependencyVersionUpdate(
				ctx,
				policy.ID,
				tt.dependencyName,
				tt.currentVersion,
				tt.proposedVersion,
				tt.ecosystem,
			)
			require.NoError(t, err)
			require.NotNil(t, analysis)

			assert.Equal(t, tt.dependencyName, analysis.DependencyName)
			assert.Equal(t, tt.ecosystem, analysis.Ecosystem)
			assert.Equal(t, tt.currentVersion, analysis.CurrentVersion)
			assert.Equal(t, tt.proposedVersion, analysis.ProposedVersion)
			assert.Equal(t, tt.expectedAllowed, analysis.VersionConstraintCheck.Allowed)
			assert.Equal(t, tt.expectedAction, analysis.RecommendedAction.Action)
			assert.Contains(t, analysis.RecommendedAction.Reason, tt.expectedReason)
		})
	}
}

func TestDependencyVersionPolicyManager_ApplyVersionConstraints(t *testing.T) {
	manager := createTestDependencyVersionPolicyManager()
	ctx := context.Background()

	// Create a test policy
	policy := createTestDependencyVersionPolicy()
	err := manager.CreateDependencyVersionPolicy(ctx, policy)
	require.NoError(t, err)

	updates := []DependencyUpdate{
		{
			Name:            "lodash",
			Ecosystem:       "npm",
			CurrentVersion:  "4.17.20",
			ProposedVersion: "4.17.21",
		},
		{
			Name:            "lodash",
			Ecosystem:       "npm",
			CurrentVersion:  "4.17.21",
			ProposedVersion: "5.0.0",
		},
		{
			Name:            "express",
			Ecosystem:       "npm",
			CurrentVersion:  "4.18.1",
			ProposedVersion: "4.18.2-beta.1",
		},
		{
			Name:            "react",
			Ecosystem:       "npm",
			CurrentVersion:  "17.0.2",
			ProposedVersion: "17.1.0",
		},
	}

	result, err := manager.ApplyVersionConstraints(ctx, policy.ID, updates)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, policy.ID, result.PolicyID)
	assert.Equal(t, 4, result.TotalUpdates)
	assert.GreaterOrEqual(t, result.ApprovedCount, 1)
	assert.GreaterOrEqual(t, result.RejectedCount, 1)
	assert.NotZero(t, result.ProcessedAt)

	// Verify counts match arrays
	assert.Equal(t, len(result.ApprovedUpdates), result.ApprovedCount)
	assert.Equal(t, len(result.RejectedUpdates), result.RejectedCount)
	assert.Equal(t, len(result.PendingReview), result.PendingReviewCount)
}

func TestDependencyVersionPolicyManager_UpdateTypeDetection(t *testing.T) {
	manager := createTestDependencyVersionPolicyManager()

	tests := []struct {
		name            string
		currentVersion  string
		proposedVersion string
		expectedType    string
	}{
		{
			name:            "major version update",
			currentVersion:  "1.0.0",
			proposedVersion: "2.0.0",
			expectedType:    "major",
		},
		{
			name:            "minor version update",
			currentVersion:  "1.0.0",
			proposedVersion: "1.1.0",
			expectedType:    "minor",
		},
		{
			name:            "patch version update",
			currentVersion:  "1.0.0",
			proposedVersion: "1.0.1",
			expectedType:    "patch",
		},
		{
			name:            "prerelease version",
			currentVersion:  "1.0.0",
			proposedVersion: "1.1.0-beta.1",
			expectedType:    "prerelease",
		},
		{
			name:            "complex version patch",
			currentVersion:  "1.2.3",
			proposedVersion: "1.2.4",
			expectedType:    "patch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateType := manager.determineUpdateType(tt.currentVersion, tt.proposedVersion)
			assert.Equal(t, tt.expectedType, updateType)
		})
	}
}

func TestDependencyVersionPolicyManager_VersionConstraintValidation(t *testing.T) {
	manager := createTestDependencyVersionPolicyManager()

	tests := []struct {
		name        string
		rule        VersionConstraintRule
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid constraint rule",
			rule: VersionConstraintRule{
				RuleID:            "valid-rule",
				DependencyPattern: "lodash",
				Ecosystem:         "npm",
				MinimumVersion:    "4.17.0",
				VersionPattern:    "^4\\.",
				UpdateStrategy:    UpdateStrategyConservative,
			},
			expectError: false,
		},
		{
			name: "missing rule ID",
			rule: VersionConstraintRule{
				DependencyPattern: "lodash",
				Ecosystem:         "npm",
			},
			expectError: true,
			errorMsg:    "rule ID is required",
		},
		{
			name: "missing dependency pattern",
			rule: VersionConstraintRule{
				RuleID:    "test-rule",
				Ecosystem: "npm",
			},
			expectError: true,
			errorMsg:    "dependency pattern is required",
		},
		{
			name: "missing ecosystem",
			rule: VersionConstraintRule{
				RuleID:            "test-rule",
				DependencyPattern: "lodash",
			},
			expectError: true,
			errorMsg:    "ecosystem is required",
		},
		{
			name: "invalid version pattern",
			rule: VersionConstraintRule{
				RuleID:            "test-rule",
				DependencyPattern: "lodash",
				Ecosystem:         "npm",
				VersionPattern:    "[invalid-regex",
			},
			expectError: true,
			errorMsg:    "invalid version pattern",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.validateVersionConstraintRule(&tt.rule)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDependencyVersionPolicyManager_EcosystemPolicyValidation(t *testing.T) {
	manager := createTestDependencyVersionPolicyManager()

	tests := []struct {
		name        string
		policy      EcosystemVersionPolicy
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid ecosystem policy",
			policy: EcosystemVersionPolicy{
				Ecosystem:              "npm",
				Enabled:                true,
				DefaultUpdateStrategy:  UpdateStrategyModerate,
				AllowMajorUpdates:      true,
				AllowMinorUpdates:      true,
				AllowPatchUpdates:      true,
				RequireSecurityUpdates: true,
				MaxVersionAge:          365 * 24 * time.Hour,
				PerformanceRequirements: PerformanceRequirements{
					MaxPerformanceRegression: 0.1,
				},
			},
			expectError: false,
		},
		{
			name: "missing ecosystem",
			policy: EcosystemVersionPolicy{
				Enabled: true,
			},
			expectError: true,
			errorMsg:    "ecosystem is required",
		},
		{
			name: "invalid performance regression threshold",
			policy: EcosystemVersionPolicy{
				Ecosystem: "npm",
				Enabled:   true,
				PerformanceRequirements: PerformanceRequirements{
					MaxPerformanceRegression: 1.5, // > 1.0 is invalid
				},
			},
			expectError: true,
			errorMsg:    "max performance regression must be between 0 and 1",
		},
		{
			name: "negative performance regression threshold",
			policy: EcosystemVersionPolicy{
				Ecosystem: "npm",
				Enabled:   true,
				PerformanceRequirements: PerformanceRequirements{
					MaxPerformanceRegression: -0.1, // negative is invalid
				},
			},
			expectError: true,
			errorMsg:    "max performance regression must be between 0 and 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.validateEcosystemVersionPolicy(&tt.policy)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDependencyVersionPolicyManager_DependencyPatternMatching(t *testing.T) {
	manager := createTestDependencyVersionPolicyManager()

	tests := []struct {
		name           string
		dependencyName string
		pattern        string
		expectedMatch  bool
		expectError    bool
	}{
		{
			name:           "exact match",
			dependencyName: "lodash",
			pattern:        "lodash",
			expectedMatch:  true,
			expectError:    false,
		},
		{
			name:           "wildcard match",
			dependencyName: "any-package",
			pattern:        "*",
			expectedMatch:  true,
			expectError:    false,
		},
		{
			name:           "regex pattern match",
			dependencyName: "lodash",
			pattern:        "^lo.*",
			expectedMatch:  true,
			expectError:    false,
		},
		{
			name:           "regex pattern no match",
			dependencyName: "express",
			pattern:        "^lo.*",
			expectedMatch:  false,
			expectError:    false,
		},
		{
			name:           "invalid regex pattern",
			dependencyName: "lodash",
			pattern:        "[invalid-regex",
			expectedMatch:  false,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, err := manager.matchesDependencyPattern(tt.dependencyName, tt.pattern)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedMatch, matched)
			}
		})
	}
}

func TestDependencyVersionPolicyManager_VersionComparison(t *testing.T) {
	manager := createTestDependencyVersionPolicyManager()

	tests := []struct {
		name     string
		v1       string
		v2       string
		expected int
	}{
		{
			name:     "equal versions",
			v1:       "1.0.0",
			v2:       "1.0.0",
			expected: 0,
		},
		{
			name:     "v1 less than v2",
			v1:       "1.0.0",
			v2:       "1.0.1",
			expected: -1,
		},
		{
			name:     "v1 greater than v2",
			v1:       "1.0.1",
			v2:       "1.0.0",
			expected: 1,
		},
		{
			name:     "different major versions",
			v1:       "2.0.0",
			v2:       "1.0.0",
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.compareVersions(tt.v1, tt.v2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDependencyVersionPolicyConstants(t *testing.T) {
	// Test update strategies
	strategies := []DependencyUpdateStrategy{
		UpdateStrategyConservative, UpdateStrategyModerate, UpdateStrategyAggressive,
		UpdateStrategySecurityOnly, UpdateStrategyCustom,
	}

	for _, strategy := range strategies {
		assert.NotEmpty(t, string(strategy))
	}

	// Test specific values
	assert.Equal(t, DependencyUpdateStrategy("conservative"), UpdateStrategyConservative)
	assert.Equal(t, DependencyUpdateStrategy("moderate"), UpdateStrategyModerate)
	assert.Equal(t, DependencyUpdateStrategy("aggressive"), UpdateStrategyAggressive)

	// Test update frequencies
	frequencies := []UpdateFrequency{
		UpdateFrequencyImmediate, UpdateFrequencyDaily, UpdateFrequencyWeekly,
		UpdateFrequencyBiWeekly, UpdateFrequencyMonthly, UpdateFrequencyQuarterly,
		UpdateFrequencyManual,
	}

	for _, frequency := range frequencies {
		assert.NotEmpty(t, string(frequency))
	}

	// Test constraint priorities
	priorities := []ConstraintPriority{
		ConstraintPriorityLow, ConstraintPriorityMedium,
		ConstraintPriorityHigh, ConstraintPriorityCritical,
	}

	for _, priority := range priorities {
		assert.NotEmpty(t, string(priority))
	}

	// Test detection methods
	methods := []DetectionMethod{
		DetectionMethodSemver, DetectionMethodAPI, DetectionMethodSchema,
		DetectionMethodCustom, DetectionMethodChangeLog, DetectionMethodBinary,
	}

	for _, method := range methods {
		assert.NotEmpty(t, string(method))
	}
}

func TestDependencyVersionPolicyManager_PolicyNotFound(t *testing.T) {
	manager := createTestDependencyVersionPolicyManager()
	ctx := context.Background()

	// Try to analyze with non-existent policy
	analysis, err := manager.AnalyzeDependencyVersionUpdate(
		ctx,
		"non-existent-policy",
		"lodash",
		"4.17.20",
		"4.17.21",
		"npm",
	)

	assert.Error(t, err)
	assert.Nil(t, analysis)
	assert.Contains(t, err.Error(), "dependency version policy not found")
}

func TestDependencyVersionPolicyManager_DisabledPolicy(t *testing.T) {
	manager := createTestDependencyVersionPolicyManager()
	ctx := context.Background()

	// Create a disabled policy
	policy := createTestDependencyVersionPolicy()
	policy.Enabled = false
	err := manager.CreateDependencyVersionPolicy(ctx, policy)
	require.NoError(t, err)

	analysis, err := manager.AnalyzeDependencyVersionUpdate(
		ctx,
		policy.ID,
		"lodash",
		"4.17.20",
		"4.17.21",
		"npm",
	)

	require.NoError(t, err)
	require.NotNil(t, analysis)

	assert.Equal(t, "skip", analysis.RecommendedAction.Action)
	assert.Contains(t, analysis.RecommendedAction.Reason, "policy is disabled")
}

// Benchmark tests.
func BenchmarkAnalyzeDependencyVersionUpdate(b *testing.B) {
	manager := createTestDependencyVersionPolicyManager()
	ctx := context.Background()

	policy := createTestDependencyVersionPolicy()
	_ = manager.CreateDependencyVersionPolicy(ctx, policy) //nolint:errcheck // Benchmark test setup

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = manager.AnalyzeDependencyVersionUpdate(ctx, policy.ID, "lodash", "4.17.20", "4.17.21", "npm") //nolint:errcheck // Benchmark test
	}
}

func BenchmarkVersionConstraintCheck(b *testing.B) {
	manager := createTestDependencyVersionPolicyManager()
	policy := createTestDependencyVersionPolicy()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = manager.checkVersionConstraints(policy, "lodash", "4.17.21", "npm") //nolint:errcheck // Benchmark test
	}
}

func BenchmarkDependencyPatternMatching(b *testing.B) {
	manager := createTestDependencyVersionPolicyManager()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = manager.matchesDependencyPattern("lodash", "^lo.*") //nolint:errcheck // Benchmark test
	}
}

// Helper functions.
func createTestDependencyVersionPolicyManager() *DependencyVersionPolicyManager {
	logger := &simpleLogger{}
	apiClient := &simpleAPIClient{}
	dependabotManager := createTestDependabotManager()
	securityPolicyManager := createTestSecurityPolicyManager()

	return NewDependencyVersionPolicyManager(logger, apiClient, dependabotManager, securityPolicyManager)
}

func createTestDependencyVersionPolicy() *DependencyVersionPolicy {
	return &DependencyVersionPolicy{
		ID:           "test-dependency-version-policy",
		Name:         "Test Dependency Version Policy",
		Organization: "testorg",
		Description:  "Test dependency version policy for unit testing",
		Enabled:      true,
		VersionConstraints: map[string]VersionConstraintRule{
			"npm-lodash": {
				RuleID:            "npm-lodash-constraint",
				DependencyPattern: "lodash",
				Ecosystem:         "npm",
				MinimumVersion:    "4.17.20",
				MaximumVersion:    "4.99.99",
				AllowPrerelease:   false,
				UpdateStrategy:    UpdateStrategyConservative,
				AutoUpdateEnabled: true,
				UpdateFrequency:   UpdateFrequencyWeekly,
				Priority:          ConstraintPriorityMedium,
				Justification:     "Security and stability requirements",
			},
		},
		EcosystemPolicies: map[string]EcosystemVersionPolicy{
			"npm": {
				Ecosystem:              "npm",
				Enabled:                true,
				DefaultUpdateStrategy:  UpdateStrategyModerate,
				AllowMajorUpdates:      false,
				AllowMinorUpdates:      true,
				AllowPatchUpdates:      true,
				RequireSecurityUpdates: true,
				MaxVersionAge:          180 * 24 * time.Hour,
				DeprecationPolicy: DeprecationPolicy{
					AllowDeprecatedVersions:  false,
					DeprecationWarningPeriod: 30 * 24 * time.Hour,
					ForceUpgradeAfterEOL:     true,
					EOLNotificationPeriod:    60 * 24 * time.Hour,
				},
				PerformanceRequirements: PerformanceRequirements{
					MaxPerformanceRegression: 0.05,
					BenchmarkSuites:          []string{"npm-benchmark"},
					PerformanceThresholds: map[string]float64{
						"bundle_size": 1000000,
						"load_time":   2000,
					},
				},
			},
		},
		BreakingChangePolicy: BreakingChangePolicy{
			AllowBreakingChanges:        false,
			ImpactAnalysisRequired:      true,
			DeprecationNoticePeriod:     90 * 24 * time.Hour,
			MigrationGuidanceRequired:   true,
			BackwardCompatibilityPeriod: 365 * 24 * time.Hour,
			BreakingChangeApprovers:     []string{"architecture-team"},
			BreakingChangeDetection: BreakingChangeDetection{
				Enabled:               true,
				Methods:               []DetectionMethod{DetectionMethodSemver},
				SemverStrictMode:      true,
				APIChangeDetection:    true,
				SchemaChangeDetection: false,
				ThresholdConfiguration: ThresholdConfig{
					MinorChangeThreshold:    0.3,
					MajorChangeThreshold:    0.7,
					BreakingChangeThreshold: 0.9,
				},
			},
		},
		CompatibilityChecks: CompatibilityCheckConfig{
			Enabled:                   true,
			DependencyGraphAnalysis:   true,
			PerformanceImpactAnalysis: true,
			SecurityImpactAnalysis:    true,
		},
		ApprovalRequirements: VersionUpdateApprovalRequirements{
			MajorVersionUpdates: VersionApprovalRule{
				RequiredApprovers:          2,
				RequiredApprovalTeams:      []string{"architecture-team"},
				ManualReviewRequired:       true,
				SecurityReviewRequired:     true,
				ArchitectureReviewRequired: true,
				TestingGateRequired:        true,
				WaitingPeriod:              24 * time.Hour,
				ApprovalTimeLimit:          7 * 24 * time.Hour,
			},
			MinorVersionUpdates: VersionApprovalRule{
				RequiredApprovers:      1,
				RequiredApprovalTeams:  []string{"development-team"},
				ManualReviewRequired:   false,
				SecurityReviewRequired: true,
				TestingGateRequired:    true,
				WaitingPeriod:          2 * time.Hour,
				ApprovalTimeLimit:      2 * 24 * time.Hour,
			},
			PatchVersionUpdates: VersionApprovalRule{
				RequiredApprovers:      0,
				ManualReviewRequired:   false,
				SecurityReviewRequired: false,
				TestingGateRequired:    true,
				WaitingPeriod:          0,
				ApprovalTimeLimit:      24 * time.Hour,
				AutoApprovalConditions: []AutoApprovalCondition{
					{
						Type:     "security_improvement",
						Field:    "has_security_fixes",
						Operator: "eq",
						Value:    true,
						Required: false,
					},
				},
			},
			SecurityUpdates: VersionApprovalRule{
				RequiredApprovers:      1,
				RequiredApprovalTeams:  []string{"security-team"},
				ManualReviewRequired:   false,
				SecurityReviewRequired: true,
				TestingGateRequired:    false,
				WaitingPeriod:          0,
				ApprovalTimeLimit:      4 * time.Hour,
			},
		},
		TestingRequirements: TestingRequirements{
			Enabled:                    true,
			UnitTestingRequired:        true,
			IntegrationTestingRequired: true,
			E2ETestingRequired:         false,
			PerformanceTestingRequired: true,
			SecurityTestingRequired:    true,
			MinimumTestCoverage:        85.0,
		},
		NotificationSettings: VersionPolicyNotificationConfig{
			Enabled:    true,
			EventTypes: []string{"policy_violation", "approval_required"},
			Channels: []VersionNotificationChannel{
				{
					Type:        "email",
					Target:      "dev-team@example.com",
					Enabled:     true,
					EventFilter: []string{"policy_violation"},
				},
			},
		},
		MetricsTracking: MetricsTrackingConfig{
			Enabled:           true,
			MetricsCollectors: []string{"prometheus"},
			TrackingFrequency: time.Hour,
			RetentionPeriod:   90 * 24 * time.Hour,
			AlertingEnabled:   true,
			DashboardEnabled:  true,
		},
	}
}
