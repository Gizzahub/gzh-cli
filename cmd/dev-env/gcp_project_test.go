package devenv

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGCPProjectManager(t *testing.T) {
	// Create temporary directory for testing
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tmpDir)

	// Create mock gcloud config directory
	gcloudDir := filepath.Join(tmpDir, ".config", "gcloud")
	require.NoError(t, os.MkdirAll(gcloudDir, 0o755))

	ctx := context.Background()
	manager, err := NewGCPProjectManager(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, manager)
	assert.Equal(t, gcloudDir, manager.gcloudConfigPath)
	assert.NotNil(t, manager.projects)
	assert.NotNil(t, manager.configurations)
}

func TestGCPProjectManager_LoadConfigurations(t *testing.T) {
	// Create temporary directory for testing
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tmpDir)

	// Create mock gcloud config structure
	gcloudDir := filepath.Join(tmpDir, ".config", "gcloud")
	configurationsDir := filepath.Join(gcloudDir, "configurations")
	require.NoError(t, os.MkdirAll(configurationsDir, 0o755))

	// Create active_config file
	activeConfigPath := filepath.Join(gcloudDir, "active_config")
	require.NoError(t, os.WriteFile(activeConfigPath, []byte("default"), 0o644))

	// Create default configuration
	defaultConfigDir := filepath.Join(configurationsDir, "default")
	require.NoError(t, os.MkdirAll(defaultConfigDir, 0o755))

	// Create properties file in JSON format
	properties := map[string]interface{}{
		"core": map[string]interface{}{
			"project": "test-project-123",
			"account": "test@example.com",
		},
		"compute": map[string]interface{}{
			"region": "us-central1",
			"zone":   "us-central1-a",
		},
	}
	propertiesData, err := json.Marshal(properties)
	require.NoError(t, err)

	propertiesPath := filepath.Join(defaultConfigDir, "properties")
	require.NoError(t, os.WriteFile(propertiesPath, propertiesData, 0o644))

	// Create staging configuration
	stagingConfigDir := filepath.Join(configurationsDir, "staging")
	require.NoError(t, os.MkdirAll(stagingConfigDir, 0o755))

	stagingProperties := map[string]interface{}{
		"core": map[string]interface{}{
			"project": "staging-project-456",
			"account": "staging@example.com",
		},
		"compute": map[string]interface{}{
			"region": "us-west1",
			"zone":   "us-west1-b",
		},
	}
	stagingPropertiesData, err := json.Marshal(stagingProperties)
	require.NoError(t, err)

	stagingPropertiesPath := filepath.Join(stagingConfigDir, "properties")
	require.NoError(t, os.WriteFile(stagingPropertiesPath, stagingPropertiesData, 0o644))

	ctx := context.Background()
	manager, err := NewGCPProjectManager(ctx)
	require.NoError(t, err)

	// Verify configurations were loaded
	assert.Len(t, manager.configurations, 2)

	// Check default configuration
	defaultConfig := manager.configurations["default"]
	require.NotNil(t, defaultConfig)
	assert.Equal(t, "default", defaultConfig.Name)
	assert.Equal(t, "test-project-123", defaultConfig.Project)
	assert.Equal(t, "test@example.com", defaultConfig.Account)
	assert.Equal(t, "us-central1", defaultConfig.Region)
	assert.Equal(t, "us-central1-a", defaultConfig.Zone)
	assert.True(t, defaultConfig.IsActive)

	// Check staging configuration
	stagingConfig := manager.configurations["staging"]
	require.NotNil(t, stagingConfig)
	assert.Equal(t, "staging", stagingConfig.Name)
	assert.Equal(t, "staging-project-456", stagingConfig.Project)
	assert.Equal(t, "staging@example.com", stagingConfig.Account)
	assert.Equal(t, "us-west1", stagingConfig.Region)
	assert.Equal(t, "us-west1-b", stagingConfig.Zone)
	assert.False(t, stagingConfig.IsActive)
}

func TestGCPProjectManager_LoadConfigurationsINIFormat(t *testing.T) {
	// Create temporary directory for testing
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tmpDir)

	// Create mock gcloud config structure
	gcloudDir := filepath.Join(tmpDir, ".config", "gcloud")
	configurationsDir := filepath.Join(gcloudDir, "configurations")
	require.NoError(t, os.MkdirAll(configurationsDir, 0o755))

	// Create test configuration directory
	testConfigDir := filepath.Join(configurationsDir, "test")
	require.NoError(t, os.MkdirAll(testConfigDir, 0o755))

	// Create properties file in INI format
	iniContent := `[core]
project = test-ini-project
account = ini@example.com

[compute]
region = europe-west1
zone = europe-west1-c
`
	propertiesPath := filepath.Join(testConfigDir, "properties")
	require.NoError(t, os.WriteFile(propertiesPath, []byte(iniContent), 0o644))

	ctx := context.Background()
	manager, err := NewGCPProjectManager(ctx)
	require.NoError(t, err)

	// Verify INI configuration was loaded
	assert.Len(t, manager.configurations, 1)

	testConfig := manager.configurations["test"]
	require.NotNil(t, testConfig)
	assert.Equal(t, "test", testConfig.Name)
	assert.Equal(t, "test-ini-project", testConfig.Project)
	assert.Equal(t, "ini@example.com", testConfig.Account)
	assert.Equal(t, "europe-west1", testConfig.Region)
	assert.Equal(t, "europe-west1-c", testConfig.Zone)
}

func TestGCPProject_JSONSerialization(t *testing.T) {
	now := time.Now()
	project := &GCPProject{
		ID:             "test-project-123",
		Name:           "Test Project",
		Number:         "123456789",
		LifecycleState: "ACTIVE",
		Account:        "test@example.com",
		Region:         "us-central1",
		Zone:           "us-central1-a",
		Configuration:  "default",
		ServiceAccount: "service@test-project.iam.gserviceaccount.com",
		BillingAccount: "012345-567890-ABCDEF",
		IsActive:       true,
		LastUsed:       &now,
		Tags: map[string]string{
			"environment": "test",
			"team":        "dev",
		},
		EnabledAPIs: []string{
			"compute.googleapis.com",
			"storage.googleapis.com",
		},
		IAMPermissions: []string{
			"roles/editor",
			"roles/storage.admin",
		},
	}

	// Test JSON marshaling
	data, err := json.Marshal(project)
	require.NoError(t, err)

	// Test JSON unmarshaling
	var unmarshaled GCPProject
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	// Verify all fields
	assert.Equal(t, project.ID, unmarshaled.ID)
	assert.Equal(t, project.Name, unmarshaled.Name)
	assert.Equal(t, project.Number, unmarshaled.Number)
	assert.Equal(t, project.LifecycleState, unmarshaled.LifecycleState)
	assert.Equal(t, project.Account, unmarshaled.Account)
	assert.Equal(t, project.Region, unmarshaled.Region)
	assert.Equal(t, project.Zone, unmarshaled.Zone)
	assert.Equal(t, project.Configuration, unmarshaled.Configuration)
	assert.Equal(t, project.ServiceAccount, unmarshaled.ServiceAccount)
	assert.Equal(t, project.BillingAccount, unmarshaled.BillingAccount)
	assert.Equal(t, project.IsActive, unmarshaled.IsActive)
	assert.True(t, project.LastUsed.Equal(*unmarshaled.LastUsed))
	assert.Equal(t, project.Tags, unmarshaled.Tags)
	assert.Equal(t, project.EnabledAPIs, unmarshaled.EnabledAPIs)
	assert.Equal(t, project.IAMPermissions, unmarshaled.IAMPermissions)
}

func TestGCPConfiguration_StructValidation(t *testing.T) {
	config := &GCPConfiguration{
		Name:           "production",
		Project:        "prod-project-789",
		Account:        "prod@company.com",
		Region:         "us-west2",
		Zone:           "us-west2-b",
		IsActive:       true,
		PropertiesPath: "/path/to/properties",
	}

	// Test JSON serialization
	data, err := json.Marshal(config)
	require.NoError(t, err)

	var unmarshaled GCPConfiguration
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, config.Name, unmarshaled.Name)
	assert.Equal(t, config.Project, unmarshaled.Project)
	assert.Equal(t, config.Account, unmarshaled.Account)
	assert.Equal(t, config.Region, unmarshaled.Region)
	assert.Equal(t, config.Zone, unmarshaled.Zone)
	assert.Equal(t, config.IsActive, unmarshaled.IsActive)
	assert.Equal(t, config.PropertiesPath, unmarshaled.PropertiesPath)
}

func TestGCPProjectManager_GetActiveConfiguration(t *testing.T) {
	// Test with existing active_config file
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tmpDir)

	gcloudDir := filepath.Join(tmpDir, ".config", "gcloud")
	require.NoError(t, os.MkdirAll(gcloudDir, 0o755))

	manager := &GCPProjectManager{
		gcloudConfigPath: gcloudDir,
		configurations:   make(map[string]*GCPConfiguration),
		projects:         make(map[string]*GCPProject),
		ctx:              context.Background(),
	}

	// Test with no active_config file (should default to "default")
	activeConfig, err := manager.getActiveConfiguration()
	assert.NoError(t, err)
	assert.Equal(t, "default", activeConfig)

	// Test with existing active_config file
	activeConfigPath := filepath.Join(gcloudDir, "active_config")
	require.NoError(t, os.WriteFile(activeConfigPath, []byte("production"), 0o644))

	activeConfig, err = manager.getActiveConfiguration()
	assert.NoError(t, err)
	assert.Equal(t, "production", activeConfig)

	// Test with active_config file containing whitespace
	require.NoError(t, os.WriteFile(activeConfigPath, []byte("  staging  \n"), 0o644))

	activeConfig, err = manager.getActiveConfiguration()
	assert.NoError(t, err)
	assert.Equal(t, "staging", activeConfig)
}

func TestGCPProjectManager_ParseConfigurationProperties(t *testing.T) {
	manager := &GCPProjectManager{}
	config := &GCPConfiguration{Name: "test"}

	t.Run("JSON format", func(t *testing.T) {
		tmpFile := filepath.Join(t.TempDir(), "properties")
		properties := map[string]interface{}{
			"core": map[string]interface{}{
				"project": "json-project",
				"account": "json@example.com",
			},
			"compute": map[string]interface{}{
				"region": "asia-east1",
				"zone":   "asia-east1-a",
			},
		}
		data, err := json.Marshal(properties)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(tmpFile, data, 0o644))

		err = manager.parseConfigurationProperties(tmpFile, config)
		require.NoError(t, err)

		assert.Equal(t, "json-project", config.Project)
		assert.Equal(t, "json@example.com", config.Account)
		assert.Equal(t, "asia-east1", config.Region)
		assert.Equal(t, "asia-east1-a", config.Zone)
	})

	t.Run("INI format", func(t *testing.T) {
		config := &GCPConfiguration{Name: "test-ini"}
		tmpFile := filepath.Join(t.TempDir(), "properties")
		iniContent := `[core]
project = ini-project
account = ini@example.com

[compute]
region = australia-southeast1
zone = australia-southeast1-b
`
		require.NoError(t, os.WriteFile(tmpFile, []byte(iniContent), 0o644))

		err := manager.parseConfigurationProperties(tmpFile, config)
		require.NoError(t, err)

		assert.Equal(t, "ini-project", config.Project)
		assert.Equal(t, "ini@example.com", config.Account)
		assert.Equal(t, "australia-southeast1", config.Region)
		assert.Equal(t, "australia-southeast1-b", config.Zone)
	})

	t.Run("Invalid file", func(t *testing.T) {
		config := &GCPConfiguration{Name: "test-invalid"}
		err := manager.parseConfigurationProperties("/nonexistent/file", config)
		assert.Error(t, err)
	})
}

func TestGCPProjectManager_EnrichProjectDetails(t *testing.T) {
	manager := &GCPProjectManager{
		configurations: map[string]*GCPConfiguration{
			"prod": {
				Name:    "prod",
				Project: "test-project-123",
				Account: "prod@example.com",
				Region:  "us-central1",
				Zone:    "us-central1-a",
			},
			"staging": {
				Name:    "staging",
				Project: "another-project-456",
				Account: "staging@example.com",
				Region:  "us-west1",
				Zone:    "us-west1-b",
			},
		},
		ctx: context.Background(),
	}

	project := &GCPProject{
		ID:   "test-project-123",
		Name: "Test Project",
	}

	manager.enrichProjectDetails(project)

	// Should find matching configuration
	assert.Equal(t, "prod@example.com", project.Account)
	assert.Equal(t, "us-central1", project.Region)
	assert.Equal(t, "us-central1-a", project.Zone)
	assert.Equal(t, "prod", project.Configuration)
}

func TestGCPProject_DefaultValues(t *testing.T) {
	project := &GCPProject{
		ID:   "test-project",
		Name: "Test Project",
	}

	// Verify default values for optional fields
	assert.Empty(t, project.ServiceAccount)
	assert.Empty(t, project.BillingAccount)
	assert.False(t, project.IsActive)
	assert.Nil(t, project.LastUsed)
	assert.Empty(t, project.Tags)
	assert.Empty(t, project.EnabledAPIs)
	assert.Empty(t, project.IAMPermissions)
}

func TestGCPProjectManager_Integration(t *testing.T) {
	// This test can be run with real gcloud CLI for integration testing
	// Skip if gcloud is not available or user is not authenticated
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create temporary directory for testing
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tmpDir)

	// Create minimal gcloud config structure
	gcloudDir := filepath.Join(tmpDir, ".config", "gcloud")
	require.NoError(t, os.MkdirAll(gcloudDir, 0o755))

	ctx := context.Background()
	manager, err := NewGCPProjectManager(ctx)

	// Should not fail even without real gcloud config
	assert.NoError(t, err)
	assert.NotNil(t, manager)

	// Test that empty manager can handle basic operations
	err = manager.listProjects("table")
	assert.NoError(t, err) // Should handle empty project list gracefully
}

// Benchmark tests for performance.
func BenchmarkGCPProjectManager_LoadConfigurations(b *testing.B) {
	// Create temporary directory with test data
	tmpDir := b.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tmpDir)

	// Create multiple configurations for benchmarking
	gcloudDir := filepath.Join(tmpDir, ".config", "gcloud")
	configurationsDir := filepath.Join(gcloudDir, "configurations")
	require.NoError(b, os.MkdirAll(configurationsDir, 0o755))

	// Create 10 test configurations
	for i := 0; i < 10; i++ {
		configName := fmt.Sprintf("config-%d", i)
		configDir := filepath.Join(configurationsDir, configName)
		require.NoError(b, os.MkdirAll(configDir, 0o755))

		properties := map[string]interface{}{
			"core": map[string]interface{}{
				"project": fmt.Sprintf("project-%d", i),
				"account": fmt.Sprintf("user%d@example.com", i),
			},
			"compute": map[string]interface{}{
				"region": "us-central1",
				"zone":   "us-central1-a",
			},
		}
		data, _ := json.Marshal(properties)
		propertiesPath := filepath.Join(configDir, "properties")
		require.NoError(b, os.WriteFile(propertiesPath, data, 0o644))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		manager := &GCPProjectManager{
			gcloudConfigPath: gcloudDir,
			configurations:   make(map[string]*GCPConfiguration),
			projects:         make(map[string]*GCPProject),
			ctx:              context.Background(),
		}
		_ = manager.loadConfigurations()
	}
}

func BenchmarkGCPProject_JSONSerialization(b *testing.B) {
	now := time.Now()
	project := &GCPProject{
		ID:             "benchmark-project",
		Name:           "Benchmark Project",
		Number:         "123456789",
		LifecycleState: "ACTIVE",
		Account:        "bench@example.com",
		Region:         "us-central1",
		Zone:           "us-central1-a",
		Configuration:  "default",
		IsActive:       true,
		LastUsed:       &now,
		Tags: map[string]string{
			"environment": "benchmark",
			"team":        "performance",
		},
		EnabledAPIs: []string{
			"compute.googleapis.com",
			"storage.googleapis.com",
			"bigquery.googleapis.com",
		},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		data, _ := json.Marshal(project)
		var unmarshaled GCPProject
		_ = json.Unmarshal(data, &unmarshaled)
	}
}
