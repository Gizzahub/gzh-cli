package github

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDependabotConfigManager(t *testing.T) {
	logger := &simpleLogger{}
	apiClient := &simpleAPIClient{}

	manager := NewDependabotConfigManager(logger, apiClient)

	assert.NotNil(t, manager)
	assert.Equal(t, logger, manager.logger)
	assert.Equal(t, apiClient, manager.apiClient)
}

func TestDependabotConfigManager_GetDependabotConfig(t *testing.T) {
	manager := createTestDependabotManager()
	ctx := context.Background()

	config, err := manager.GetDependabotConfig(ctx, "testorg", "testrepo")
	require.NoError(t, err)
	require.NotNil(t, config)

	assert.Equal(t, 2, config.Version)
	assert.Len(t, config.Updates, 1)
	assert.Equal(t, EcosystemGoModules, config.Updates[0].PackageEcosystem)
	assert.Equal(t, "/", config.Updates[0].Directory)
	assert.Equal(t, IntervalWeekly, config.Updates[0].Schedule.Interval)
}

func TestDependabotConfigManager_ValidateConfig(t *testing.T) {
	manager := createTestDependabotManager()

	tests := []struct {
		name        string
		config      *DependabotConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid configuration",
			config: &DependabotConfig{
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
			expectError: false,
		},
		{
			name: "invalid version",
			config: &DependabotConfig{
				Version: 1,
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
			expectError: true,
			errorMsg:    "unsupported version",
		},
		{
			name: "no update rules",
			config: &DependabotConfig{
				Version: 2,
				Updates: []DependabotUpdateRule{},
			},
			expectError: true,
			errorMsg:    "at least one update rule is required",
		},
		{
			name: "invalid ecosystem",
			config: &DependabotConfig{
				Version: 2,
				Updates: []DependabotUpdateRule{
					{
						PackageEcosystem: "invalid-ecosystem",
						Directory:        "/",
						Schedule: DependabotSchedule{
							Interval: IntervalWeekly,
						},
					},
				},
			},
			expectError: true,
			errorMsg:    "unsupported package ecosystem",
		},
		{
			name: "empty directory",
			config: &DependabotConfig{
				Version: 2,
				Updates: []DependabotUpdateRule{
					{
						PackageEcosystem: EcosystemGoModules,
						Directory:        "",
						Schedule: DependabotSchedule{
							Interval: IntervalWeekly,
						},
					},
				},
			},
			expectError: true,
			errorMsg:    "directory is required",
		},
		{
			name: "invalid schedule interval",
			config: &DependabotConfig{
				Version: 2,
				Updates: []DependabotUpdateRule{
					{
						PackageEcosystem: EcosystemGoModules,
						Directory:        "/",
						Schedule: DependabotSchedule{
							Interval: "invalid-interval",
						},
					},
				},
			},
			expectError: true,
			errorMsg:    "invalid schedule interval",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.ValidateConfig(tt.config)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDependabotConfigManager_CreateDefaultConfig(t *testing.T) {
	manager := createTestDependabotManager()
	ctx := context.Background()

	tests := []struct {
		name       string
		ecosystems []string
		expected   int
	}{
		{
			name:       "single ecosystem",
			ecosystems: []string{EcosystemGoModules},
			expected:   1,
		},
		{
			name:       "multiple ecosystems",
			ecosystems: []string{EcosystemGoModules, EcosystemNPM, EcosystemDockerfile},
			expected:   3,
		},
		{
			name:       "no ecosystems",
			ecosystems: []string{},
			expected:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := manager.CreateDefaultConfig(ctx, "testorg", "testrepo", tt.ecosystems)
			require.NoError(t, err)
			require.NotNil(t, config)

			assert.Equal(t, 2, config.Version)
			assert.Len(t, config.Updates, tt.expected)

			// Validate each generated rule
			for i, ecosystem := range tt.ecosystems {
				rule := config.Updates[i]
				assert.Equal(t, ecosystem, rule.PackageEcosystem)
				assert.Equal(t, "/", rule.Directory)
				assert.Equal(t, IntervalWeekly, rule.Schedule.Interval)
				assert.Equal(t, "monday", rule.Schedule.Day)
				assert.Equal(t, "06:00", rule.Schedule.Time)
				assert.Equal(t, "UTC", rule.Schedule.Timezone)
				assert.Equal(t, 5, rule.PullRequestLimit)
				assert.Contains(t, rule.Labels, "dependencies")
				assert.Contains(t, rule.Labels, ecosystem)
			}
		})
	}
}

func TestDependabotConfigManager_GetDependabotStatus(t *testing.T) {
	manager := createTestDependabotManager()
	ctx := context.Background()

	status, err := manager.GetDependabotStatus(ctx, "testorg", "testrepo")
	require.NoError(t, err)
	require.NotNil(t, status)

	assert.Equal(t, "testrepo", status.Repository)
	assert.Equal(t, "testorg", status.Organization)
	assert.True(t, status.Enabled)
	assert.True(t, status.ConfigExists)
	assert.True(t, status.ConfigValid)
	assert.Equal(t, 2, status.ActivePullRequests)
	assert.Len(t, status.RecentUpdates, 1)
	assert.Len(t, status.SupportedEcosystems, 3)

	// Check recent update details
	update := status.RecentUpdates[0]
	assert.Equal(t, "github.com/stretchr/testify", update.Dependency)
	assert.Equal(t, "v1.8.0", update.FromVersion)
	assert.Equal(t, "v1.8.4", update.ToVersion)
	assert.Equal(t, EcosystemGoModules, update.Ecosystem)
	assert.Equal(t, DependabotUpdateStatusMerged, update.Status)

	// Check config summary
	summary := status.ConfigSummary
	assert.Equal(t, 3, summary.TotalEcosystems)
	assert.Len(t, summary.EnabledEcosystems, 3)
	assert.True(t, summary.SecurityUpdatesEnabled)
}

func TestDependabotConfigManager_DetectEcosystems(t *testing.T) {
	manager := createTestDependabotManager()
	ctx := context.Background()

	tests := []struct {
		name           string
		repository     string
		expectedMin    int
		expectedCommon []string
	}{
		{
			name:           "go repository",
			repository:     "my-go-project",
			expectedMin:    2,
			expectedCommon: []string{EcosystemGoModules, EcosystemGitHubActions},
		},
		{
			name:           "node repository",
			repository:     "my-node-app",
			expectedMin:    3,
			expectedCommon: []string{EcosystemGoModules, EcosystemNPM, EcosystemGitHubActions},
		},
		{
			name:           "python repository",
			repository:     "python-service",
			expectedMin:    3,
			expectedCommon: []string{EcosystemGoModules, EcosystemPip, EcosystemGitHubActions},
		},
		{
			name:           "docker repository",
			repository:     "docker-app",
			expectedMin:    3,
			expectedCommon: []string{EcosystemGoModules, EcosystemDockerfile, EcosystemGitHubActions},
		},
		{
			name:           "generic repository",
			repository:     "generic-repo",
			expectedMin:    2,
			expectedCommon: []string{EcosystemGoModules, EcosystemGitHubActions},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ecosystems, err := manager.DetectEcosystems(ctx, "testorg", tt.repository)
			require.NoError(t, err)
			assert.GreaterOrEqual(t, len(ecosystems), tt.expectedMin)

			// Check that expected common ecosystems are present
			for _, expected := range tt.expectedCommon {
				assert.Contains(t, ecosystems, expected)
			}
		})
	}
}

func TestDependabotConfigManager_UpdateDependabotConfig(t *testing.T) {
	manager := createTestDependabotManager()
	ctx := context.Background()

	tests := []struct {
		name        string
		config      *DependabotConfig
		expectError bool
	}{
		{
			name: "valid configuration update",
			config: &DependabotConfig{
				Version: 2,
				Updates: []DependabotUpdateRule{
					{
						PackageEcosystem: EcosystemGoModules,
						Directory:        "/",
						Schedule: DependabotSchedule{
							Interval: IntervalDaily,
						},
						PullRequestLimit: 10,
					},
				},
			},
			expectError: false,
		},
		{
			name: "invalid configuration update",
			config: &DependabotConfig{
				Version: 2,
				Updates: []DependabotUpdateRule{
					{
						PackageEcosystem: "invalid-ecosystem",
						Directory:        "/",
						Schedule: DependabotSchedule{
							Interval: IntervalDaily,
						},
					},
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.UpdateDependabotConfig(ctx, "testorg", "testrepo", tt.config)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDependabotEcosystemConstants(t *testing.T) {
	// Test that ecosystem constants are defined
	ecosystems := []string{
		EcosystemNPM, EcosystemPip, EcosystemBundler, EcosystemGradle,
		EcosystemMaven, EcosystemComposer, EcosystemNuGet, EcosystemCargoRust,
		EcosystemGoModules, EcosystemDockerfile, EcosystemGitSubmodule,
		EcosystemGitHubActions, EcosystemTerraform, EcosystemElm,
		EcosystemMix, EcosystemPub, EcosystemSwift,
	}

	for _, ecosystem := range ecosystems {
		assert.NotEmpty(t, ecosystem)
	}

	// Test specific values
	assert.Equal(t, "npm", EcosystemNPM)
	assert.Equal(t, "gomod", EcosystemGoModules)
	assert.Equal(t, "docker", EcosystemDockerfile)
	assert.Equal(t, "github-actions", EcosystemGitHubActions)
}

func TestDependabotUpdateStatusConstants(t *testing.T) {
	// Test update status constants
	statuses := []DependabotUpdateStatus{
		DependabotUpdateStatusPending, DependabotUpdateStatusActive,
		DependabotUpdateStatusMerged, DependabotUpdateStatusClosed,
		DependabotUpdateStatusSuperseded, DependabotUpdateStatusFailed,
	}

	for _, status := range statuses {
		assert.NotEmpty(t, string(status))
	}

	// Test specific values
	assert.Equal(t, DependabotUpdateStatus("pending"), DependabotUpdateStatusPending)
	assert.Equal(t, DependabotUpdateStatus("merged"), DependabotUpdateStatusMerged)
	assert.Equal(t, DependabotUpdateStatus("failed"), DependabotUpdateStatusFailed)
}

func TestDependabotErrorTypeConstants(t *testing.T) {
	// Test error type constants
	errorTypes := []DependabotErrorType{
		DependabotErrorTypeConfigInvalid, DependabotErrorTypeEcosystemNotFound,
		DependabotErrorTypeRegistryAuth, DependabotErrorTypePermissions,
		DependabotErrorTypeRateLimit, DependabotErrorTypeUnknown,
	}

	for _, errorType := range errorTypes {
		assert.NotEmpty(t, string(errorType))
	}

	// Test specific values
	assert.Equal(t, DependabotErrorType("config_invalid"), DependabotErrorTypeConfigInvalid)
	assert.Equal(t, DependabotErrorType("registry_auth_failed"), DependabotErrorTypeRegistryAuth)
}

func TestDependabotConfigManager_EcosystemSpecificDefaults(t *testing.T) {
	manager := createTestDependabotManager()
	ctx := context.Background()

	tests := []struct {
		name      string
		ecosystem string
		checkFunc func(t *testing.T, rule DependabotUpdateRule)
	}{
		{
			name:      "go modules defaults",
			ecosystem: EcosystemGoModules,
			checkFunc: func(t *testing.T, rule DependabotUpdateRule) {
				assert.True(t, rule.VendorUpdates)
				assert.NotNil(t, rule.CommitMessage)
				assert.Equal(t, "deps", rule.CommitMessage.Prefix)
				assert.Equal(t, "scope", rule.CommitMessage.Include)
			},
		},
		{
			name:      "npm defaults",
			ecosystem: EcosystemNPM,
			checkFunc: func(t *testing.T, rule DependabotUpdateRule) {
				assert.Equal(t, VersioningStrategyIncrease, rule.VersioningStrategy)
				assert.Len(t, rule.AllowedUpdates, 2)
				assert.Equal(t, "direct", rule.AllowedUpdates[0].DependencyType)
			},
		},
		{
			name:      "dockerfile defaults",
			ecosystem: EcosystemDockerfile,
			checkFunc: func(t *testing.T, rule DependabotUpdateRule) {
				assert.Equal(t, IntervalMonthly, rule.Schedule.Interval)
				assert.Equal(t, 3, rule.PullRequestLimit)
			},
		},
		{
			name:      "github-actions defaults",
			ecosystem: EcosystemGitHubActions,
			checkFunc: func(t *testing.T, rule DependabotUpdateRule) {
				assert.Equal(t, IntervalWeekly, rule.Schedule.Interval)
				assert.NotNil(t, rule.Groups)
				assert.Contains(t, rule.Groups, "github-actions")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := manager.CreateDefaultConfig(ctx, "testorg", "testrepo", []string{tt.ecosystem})
			require.NoError(t, err)
			require.Len(t, config.Updates, 1)

			rule := config.Updates[0]
			assert.Equal(t, tt.ecosystem, rule.PackageEcosystem)
			tt.checkFunc(t, rule)
		})
	}
}

func TestDependabotConfigManager_ValidationHelpers(t *testing.T) {
	manager := createTestDependabotManager()

	t.Run("isSupportedEcosystem", func(t *testing.T) {
		assert.True(t, manager.isSupportedEcosystem(EcosystemGoModules))
		assert.True(t, manager.isSupportedEcosystem(EcosystemNPM))
		assert.True(t, manager.isSupportedEcosystem(EcosystemDockerfile))
		assert.False(t, manager.isSupportedEcosystem("invalid-ecosystem"))
		assert.False(t, manager.isSupportedEcosystem(""))
	})

	t.Run("isValidInterval", func(t *testing.T) {
		assert.True(t, manager.isValidInterval(IntervalDaily))
		assert.True(t, manager.isValidInterval(IntervalWeekly))
		assert.True(t, manager.isValidInterval(IntervalMonthly))
		assert.False(t, manager.isValidInterval("invalid-interval"))
		assert.False(t, manager.isValidInterval(""))
	})

	t.Run("isValidVersioningStrategy", func(t *testing.T) {
		assert.True(t, manager.isValidVersioningStrategy(VersioningStrategyAuto))
		assert.True(t, manager.isValidVersioningStrategy(VersioningStrategyIncrease))
		assert.True(t, manager.isValidVersioningStrategy(VersioningStrategyWiden))
		assert.False(t, manager.isValidVersioningStrategy("invalid-strategy"))
		assert.False(t, manager.isValidVersioningStrategy(""))
	})
}

// Benchmark tests.
func BenchmarkCreateDefaultConfig(b *testing.B) {
	manager := createTestDependabotManager()
	ctx := context.Background()
	ecosystems := []string{EcosystemGoModules, EcosystemNPM, EcosystemDockerfile, EcosystemGitHubActions}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		manager.CreateDefaultConfig(ctx, "testorg", "testrepo", ecosystems)
	}
}

func BenchmarkValidateConfig(b *testing.B) {
	manager := createTestDependabotManager()
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
			{
				PackageEcosystem: EcosystemNPM,
				Directory:        "/frontend",
				Schedule: DependabotSchedule{
					Interval: IntervalDaily,
				},
			},
		},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		manager.ValidateConfig(config)
	}
}

func BenchmarkDetectEcosystems(b *testing.B) {
	manager := createTestDependabotManager()
	ctx := context.Background()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		manager.DetectEcosystems(ctx, "testorg", "test-node-python-docker-repo")
	}
}

// Integration test.
func TestDependabotConfigManager_FullWorkflow(t *testing.T) {
	manager := createTestDependabotManager()
	ctx := context.Background()

	// 1. Detect ecosystems
	ecosystems, err := manager.DetectEcosystems(ctx, "testorg", "testrepo")
	require.NoError(t, err)
	assert.NotEmpty(t, ecosystems)

	// 2. Create default configuration
	config, err := manager.CreateDefaultConfig(ctx, "testorg", "testrepo", ecosystems)
	require.NoError(t, err)
	assert.NotNil(t, config)

	// 3. Validate configuration
	err = manager.ValidateConfig(config)
	assert.NoError(t, err)

	// 4. Update configuration
	err = manager.UpdateDependabotConfig(ctx, "testorg", "testrepo", config)
	assert.NoError(t, err)

	// 5. Get status
	status, err := manager.GetDependabotStatus(ctx, "testorg", "testrepo")
	require.NoError(t, err)
	assert.NotNil(t, status)
	assert.True(t, status.Enabled)
}

// Helper function to create a test Dependabot manager.
func createTestDependabotManager() *DependabotConfigManager {
	logger := &simpleLogger{}
	apiClient := &simpleAPIClient{}

	return NewDependabotConfigManager(logger, apiClient)
}
