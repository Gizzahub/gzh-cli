//nolint:testpackage // White-box testing needed for internal function access
package cloud

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Gizzahub/gzh-cli/internal/env"
)

// MockProvider implements Provider interface for testing.
type MockProvider struct {
	name     string
	profiles map[string]*Profile
	synced   map[string]*Profile
}

func NewMockProvider(name string) *MockProvider {
	return &MockProvider{
		name:     name,
		profiles: make(map[string]*Profile),
		synced:   make(map[string]*Profile),
	}
}

func (m *MockProvider) Name() string {
	return m.name
}

func (m *MockProvider) Initialize(ctx context.Context, config ProviderConfig) error {
	return nil
}

func (m *MockProvider) GetProfile(ctx context.Context, profileName string) (*Profile, error) {
	if profile, exists := m.profiles[profileName]; exists {
		return profile, nil
	}

	return nil, fmt.Errorf("profile not found: %s", profileName)
}

func (m *MockProvider) ListProfiles(ctx context.Context) ([]*Profile, error) {
	profiles := make([]*Profile, 0, len(m.profiles))
	for _, profile := range m.profiles {
		profiles = append(profiles, profile)
	}

	return profiles, nil
}

func (m *MockProvider) SyncProfile(ctx context.Context, profile *Profile) error {
	// Store synced profile
	m.synced[profile.Name] = profile
	return nil
}

func (m *MockProvider) GetNetworkPolicy(ctx context.Context, profileName string) (*NetworkPolicy, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockProvider) ApplyNetworkPolicy(ctx context.Context, policy *NetworkPolicy) error {
	return fmt.Errorf("not implemented")
}

func (m *MockProvider) ValidateConfig(config ProviderConfig) error {
	return nil
}

func (m *MockProvider) HealthCheck(ctx context.Context) error {
	return nil
}

func (m *MockProvider) AddProfile(profile *Profile) {
	m.profiles[profile.Name] = profile
}

func (m *MockProvider) GetSyncedProfile(name string) *Profile {
	return m.synced[name]
}

func TestNewSyncManager(t *testing.T) {
	config := &Config{
		Sync: SyncConfig{
			Enabled:      true,
			ConflictMode: ConflictStrategySourceWins,
		},
	}

	manager := NewSyncManager(config)
	assert.NotNil(t, manager)

	// Test type assertion
	defaultManager, ok := manager.(*DefaultSyncManager)
	assert.True(t, ok)
	assert.Equal(t, config, defaultManager.config)
}

func TestSyncProfiles_Success(t *testing.T) {
	config := &Config{
		Sync: SyncConfig{
			Enabled:      true,
			ConflictMode: ConflictStrategySourceWins,
		},
	}

	manager := NewSyncManager(config)
	ctx := context.Background()

	// Create mock providers
	source := NewMockProvider("aws")
	target := NewMockProvider("gcp")

	// Add source profile
	sourceProfile := &Profile{
		Name:        "test-profile",
		Provider:    "aws",
		Environment: env.DevEnvironment,
		Region:      "us-east-1",
		Network: NetworkConfig{
			VPCId:      "vpc-123",
			DNSServers: []string{"8.8.8.8", "8.8.4.4"},
		},
		Tags: map[string]string{
			"environment": "development",
			"team":        "backend",
		},
	}
	source.AddProfile(sourceProfile)

	// Sync profiles
	err := manager.SyncProfiles(ctx, source, target, []string{"test-profile"})
	assert.NoError(t, err)

	// Verify sync
	syncedProfile := target.GetSyncedProfile("test-profile")
	require.NotNil(t, syncedProfile)
	assert.Equal(t, "test-profile", syncedProfile.Name)
	assert.Equal(t, "gcp", syncedProfile.Provider)
	assert.Equal(t, env.DevEnvironment, syncedProfile.Environment)
	assert.Equal(t, "us-east-1", syncedProfile.Region)
	assert.Equal(t, "vpc-123", syncedProfile.Network.VPCId)
	assert.Equal(t, []string{"8.8.8.8", "8.8.4.4"}, syncedProfile.Network.DNSServers)
	assert.Equal(t, "aws", syncedProfile.Tags["sync_source"])
}

func TestSyncProfiles_WithConflicts(t *testing.T) {
	config := &Config{
		Sync: SyncConfig{
			Enabled:      true,
			ConflictMode: ConflictStrategySourceWins,
		},
	}

	manager := NewSyncManager(config)
	ctx := context.Background()

	// Create mock providers
	source := NewMockProvider("aws")
	target := NewMockProvider("gcp")

	// Add source profile
	sourceProfile := &Profile{
		Name:        "test-profile",
		Provider:    "aws",
		Environment: env.DevEnvironment,
		Region:      "us-east-1",
		Network: NetworkConfig{
			VPCId:      "vpc-123",
			DNSServers: []string{"8.8.8.8"},
		},
		Tags: map[string]string{
			"environment": "development",
		},
	}
	source.AddProfile(sourceProfile)

	// Add conflicting target profile
	targetProfile := &Profile{
		Name:        "test-profile",
		Provider:    "gcp",
		Environment: "prod",
		Region:      "us-west-1",
		Network: NetworkConfig{
			VPCId:      "vpc-456",
			DNSServers: []string{"1.1.1.1"},
		},
		Tags: map[string]string{
			"environment": "production",
		},
	}
	target.AddProfile(targetProfile)

	// Sync profiles
	err := manager.SyncProfiles(ctx, source, target, []string{"test-profile"})
	assert.NoError(t, err)

	// Verify sync (source should win)
	syncedProfile := target.GetSyncedProfile("test-profile")
	require.NotNil(t, syncedProfile)
	assert.Equal(t, env.DevEnvironment, syncedProfile.Environment) // Source wins
	assert.Equal(t, "us-east-1", syncedProfile.Region)             // Source wins
}

func TestSyncAll(t *testing.T) {
	config := &Config{
		Sync: SyncConfig{
			Enabled:      true,
			ConflictMode: ConflictStrategySourceWins,
		},
	}

	manager := NewSyncManager(config)
	ctx := context.Background()

	// Create mock providers
	source := NewMockProvider("aws")
	target := NewMockProvider("gcp")

	// Add multiple source profiles
	profiles := []*Profile{
		{
			Name:        "profile1",
			Provider:    "aws",
			Environment: env.DevEnvironment,
			Region:      "us-east-1",
		},
		{
			Name:        "profile2",
			Provider:    "aws",
			Environment: "prod",
			Region:      "us-west-2",
		},
	}

	for _, profile := range profiles {
		source.AddProfile(profile)
	}

	// Sync all profiles
	err := manager.SyncAll(ctx, source, target)
	assert.NoError(t, err)

	// Verify all profiles were synced
	for _, profile := range profiles {
		syncedProfile := target.GetSyncedProfile(profile.Name)
		assert.NotNil(t, syncedProfile)
		assert.Equal(t, "gcp", syncedProfile.Provider)
	}
}

func TestDetectConflicts(t *testing.T) {
	config := &Config{
		Sync: SyncConfig{
			Enabled: true,
		},
	}

	mgr := NewSyncManager(config)
	manager, ok := mgr.(*DefaultSyncManager)
	if !ok {
		t.Fatalf("Expected DefaultSyncManager, got %T", mgr)
	}

	source := &Profile{
		Name:        "test-profile",
		Environment: env.DevEnvironment,
		Region:      "us-east-1",
		Network: NetworkConfig{
			VPCId:      "vpc-123",
			DNSServers: []string{"8.8.8.8"},
		},
		Tags: map[string]string{
			"team": "backend",
		},
	}

	target := &Profile{
		Name:        "test-profile",
		Environment: "prod",
		Region:      "us-west-1",
		Network: NetworkConfig{
			VPCId:      "vpc-456",
			DNSServers: []string{"1.1.1.1"},
		},
		Tags: map[string]string{
			"team": "frontend",
		},
	}

	conflicts := manager.detectConflicts("test-profile", source, target)

	// Should detect conflicts in environment, region, vpc_id, dns_servers, and tags
	assert.Len(t, conflicts, 5)

	// Check specific conflicts
	conflictFields := make(map[string]bool)
	for _, conflict := range conflicts {
		conflictFields[conflict.Field] = true
	}

	assert.True(t, conflictFields["environment"])
	assert.True(t, conflictFields["region"])
	assert.True(t, conflictFields["network.vpc_id"])
	assert.True(t, conflictFields["network.dns_servers"])
	assert.True(t, conflictFields["tags"])
}

func TestMergeProfiles(t *testing.T) {
	config := &Config{
		Sync: SyncConfig{
			Enabled: true,
		},
	}

	mgr := NewSyncManager(config)
	manager, ok := mgr.(*DefaultSyncManager)
	if !ok {
		t.Fatalf("Expected DefaultSyncManager, got %T", mgr)
	}

	source := &Profile{
		Name:        "test-profile",
		Provider:    "aws",
		Environment: env.DevEnvironment,
		Region:      "us-east-1",
		Network: NetworkConfig{
			VPCId:      "vpc-123",
			DNSServers: []string{"8.8.8.8"},
		},
		Services: map[string]ServiceConfig{
			"api": {
				Endpoint: "api.aws.com",
				Port:     443,
			},
		},
		Tags: map[string]string{
			"team":        "backend",
			"environment": "development",
		},
	}

	target := &Profile{
		Name:        "test-profile",
		Provider:    "gcp",
		Environment: env.DevEnvironment,
		Region:      "us-east-1",
		Services: map[string]ServiceConfig{
			"db": {
				Endpoint: "db.gcp.com",
				Port:     5432,
			},
		},
		Tags: map[string]string{
			"cost-center": "engineering",
		},
	}

	merged := manager.mergeProfiles(source, target, "gcp")

	assert.Equal(t, "test-profile", merged.Name)
	assert.Equal(t, "gcp", merged.Provider)
	assert.Equal(t, "dev", merged.Environment)
	assert.Equal(t, "vpc-123", merged.Network.VPCId)

	// Should have services from both profiles
	assert.Len(t, merged.Services, 2)
	assert.Equal(t, "api.aws.com", merged.Services["api"].Endpoint)
	assert.Equal(t, "db.gcp.com", merged.Services["db"].Endpoint)

	// Should have tags from both profiles, plus sync metadata
	assert.Equal(t, "backend", merged.Tags["team"])
	assert.Equal(t, "engineering", merged.Tags["cost-center"])
	assert.Equal(t, "aws", merged.Tags["sync_source"])
	assert.NotEmpty(t, merged.Tags["sync_timestamp"])
}

func TestMergeValues(t *testing.T) {
	config := &Config{
		Sync: SyncConfig{
			Enabled: true,
		},
	}

	mgr := NewSyncManager(config)
	manager, ok := mgr.(*DefaultSyncManager)
	if !ok {
		t.Fatalf("Expected DefaultSyncManager, got %T", mgr)
	}

	// Test string slice merge
	source := []string{"a", "b", "c"}
	target := []string{"b", "c", "d"}
	merged, err := manager.mergeValues(source, target)
	assert.NoError(t, err)

	mergedSlice, ok := merged.([]string)
	require.True(t, ok, "type assertion failed: expected []string")
	assert.Len(t, mergedSlice, 4) // Should remove duplicates
	assert.Contains(t, mergedSlice, "a")
	assert.Contains(t, mergedSlice, "b")
	assert.Contains(t, mergedSlice, "c")
	assert.Contains(t, mergedSlice, "d")

	// Test map merge
	sourceMap := map[string]string{"key1": "value1", "key2": "value2"}
	targetMap := map[string]string{"key2": "old_value", "key3": "value3"}
	merged, err = manager.mergeValues(sourceMap, targetMap)
	assert.NoError(t, err)

	mergedMap, ok := merged.(map[string]string)
	require.True(t, ok, "type assertion failed: expected map[string]string")
	assert.Len(t, mergedMap, 3)
	assert.Equal(t, "value1", mergedMap["key1"])
	assert.Equal(t, "value2", mergedMap["key2"]) // Source wins
	assert.Equal(t, "value3", mergedMap["key3"])
}

func TestConflictResolution(t *testing.T) {
	config := &Config{
		Sync: SyncConfig{
			Enabled: true,
		},
	}

	mgr := NewSyncManager(config)
	manager, ok := mgr.(*DefaultSyncManager)
	if !ok {
		t.Fatalf("Expected DefaultSyncManager, got %T", mgr)
	}

	conflicts := []SyncConflict{
		{
			ProfileName: "test",
			Field:       "environment",
			SourceValue: "dev",
			TargetValue: "prod",
		},
	}

	// Test source wins
	err := manager.ResolveSyncConflicts(conflicts, ConflictStrategySourceWins)
	assert.NoError(t, err)
	assert.Equal(t, "dev", conflicts[0].SourceValue)

	// Test target wins
	conflicts[0].SourceValue = "dev" // Reset
	err = manager.ResolveSyncConflicts(conflicts, ConflictStrategyTargetWins)
	assert.NoError(t, err)
	assert.Equal(t, "prod", conflicts[0].SourceValue)
}

func TestSyncHistory(t *testing.T) {
	config := &Config{
		Sync: SyncConfig{
			Enabled: true,
		},
	}

	manager := &DefaultSyncManager{
		config: config,
	}

	// Add sync results
	results := []SyncStatus{
		{
			ProfileName: "test1",
			Source:      "aws",
			Target:      "gcp",
			Status:      "synced",
			LastSync:    time.Now(),
		},
		{
			ProfileName: "test2",
			Source:      "aws",
			Target:      "azure",
			Status:      "error",
			Error:       "connection failed",
			LastSync:    time.Now(),
		},
	}

	manager.updateSyncHistory(results)

	assert.Len(t, manager.syncHistory, 2)
	assert.Equal(t, "test1", manager.syncHistory[0].ProfileName)
	assert.Equal(t, "synced", manager.syncHistory[0].Status)
	assert.Equal(t, "test2", manager.syncHistory[1].ProfileName)
	assert.Equal(t, "error", manager.syncHistory[1].Status)
}

func TestValidateSyncConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			config: &Config{
				Providers: map[string]ProviderConfig{
					"aws": {Type: "aws", Region: "us-east-1"},
					"gcp": {Type: "gcp", Region: "us-central1"},
				},
				Sync: SyncConfig{
					Enabled:      true,
					ConflictMode: ConflictStrategySourceWins,
					Targets: []SyncTarget{
						{Source: "aws", Target: "gcp"},
					},
				},
			},
			expectError: false,
		},
		{
			name: "disabled sync",
			config: &Config{
				Sync: SyncConfig{
					Enabled: false,
				},
			},
			expectError: false,
		},
		{
			name: "invalid conflict strategy",
			config: &Config{
				Sync: SyncConfig{
					Enabled:      true,
					ConflictMode: "invalid_strategy",
				},
			},
			expectError: true,
			errorMsg:    "invalid conflict strategy",
		},
		{
			name: "missing source provider",
			config: &Config{
				Providers: map[string]ProviderConfig{
					"gcp": {Type: "gcp", Region: "us-central1"},
				},
				Sync: SyncConfig{
					Enabled: true,
					Targets: []SyncTarget{
						{Source: "aws", Target: "gcp"},
					},
				},
			},
			expectError: true,
			errorMsg:    "source provider aws not found",
		},
		{
			name: "same source and target",
			config: &Config{
				Providers: map[string]ProviderConfig{
					"aws": {Type: "aws", Region: "us-east-1"},
				},
				Sync: SyncConfig{
					Enabled: true,
					Targets: []SyncTarget{
						{Source: "aws", Target: "aws"},
					},
				},
			},
			expectError: true,
			errorMsg:    "source and target cannot be the same",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSyncConfig(tt.config)

			if tt.expectError {
				assert.Error(t, err)

				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetSyncRecommendations(t *testing.T) {
	config := &Config{
		Providers: map[string]ProviderConfig{
			"aws-prod": {Type: "aws", Region: "us-east-1"},
			"gcp-prod": {Type: "gcp", Region: "us-central1"},
			"aws-dev":  {Type: "aws", Region: "us-west-1"},
		},
		Profiles: map[string]Profile{
			"app-prod-aws": {Provider: "aws-prod", Environment: "prod"},
			"app-prod-gcp": {Provider: "gcp-prod", Environment: "prod"},
			"app-dev-aws":  {Provider: "aws-dev", Environment: "dev"},
		},
	}

	recommendations, err := GetSyncRecommendations(config)
	assert.NoError(t, err)

	// Should recommend sync for prod environment (has both aws-prod and gcp-prod)
	assert.Len(t, recommendations, 1)
	assert.Equal(t, "aws-prod", recommendations[0].Source)
	assert.Equal(t, "gcp-prod", recommendations[0].Target)
	assert.Contains(t, recommendations[0].Profiles, "app-prod-aws")
	assert.Contains(t, recommendations[0].Profiles, "app-prod-gcp")
}
