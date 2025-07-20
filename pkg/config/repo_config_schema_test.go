//nolint:testpackage // White-box testing needed for internal function access
package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadRepoConfig(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "repo-config.yaml")

	configContent := `
version: "1.0.0"
organization: "test-org"

defaults:
  template: "standard"

templates:
  standard:
    description: "Standard template"
    settings:
      private: true
      has_issues: true
    security:
      vulnerability_alerts: true

repositories:
  specific:
    - name: "test-repo"
      template: "standard"
      settings:
        description: "Test repository"
`

	err := os.WriteFile(configFile, []byte(configContent), 0o644)
	require.NoError(t, err)

	// Test loading
	config, err := LoadRepoConfig(configFile)
	require.NoError(t, err)

	assert.Equal(t, "1.0.0", config.Version)
	assert.Equal(t, "test-org", config.Organization)
	assert.NotNil(t, config.Defaults)
	assert.Equal(t, "standard", config.Defaults.Template)
	assert.Len(t, config.Templates, 1)
	assert.Len(t, config.Repositories.Specific, 1)
}

func TestValidateRepoConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  RepoConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "missing version",
			config: RepoConfig{
				Organization: "test-org",
			},
			wantErr: true,
			errMsg:  "version is required",
		},
		{
			name: "missing organization",
			config: RepoConfig{
				Version: "1.0.0",
			},
			wantErr: true,
			errMsg:  "organization is required",
		},
		{
			name: "circular template dependency",
			config: RepoConfig{
				Version:      "1.0.0",
				Organization: "test-org",
				Templates: map[string]*RepoTemplate{
					"a": {Base: "b"},
					"b": {Base: "a"},
				},
			},
			wantErr: true,
			errMsg:  "circular dependency",
		},
		{
			name: "valid config",
			config: RepoConfig{
				Version:      "1.0.0",
				Organization: "test-org",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRepoConfig(&tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetEffectiveConfig(t *testing.T) {
	config := &RepoConfig{
		Version:      "1.0.0",
		Organization: "test-org",
		Defaults: &RepoDefaults{
			Settings: &RepoSettings{
				Private:   boolPtr(true),
				HasIssues: boolPtr(true),
			},
		},
		Templates: map[string]*RepoTemplate{
			"microservice": {
				Settings: &RepoSettings{
					HasWiki:     boolPtr(false),
					HasProjects: boolPtr(false),
				},
				Security: &SecuritySettings{
					VulnerabilityAlerts: boolPtr(true),
				},
			},
		},
		Repositories: &RepoTargets{
			Specific: []RepoSpecificConfig{
				{
					Name:     "api-gateway",
					Template: "microservice",
					Settings: &RepoSettings{
						Description: stringPtr("API Gateway Service"),
					},
				},
			},
			Patterns: []RepoPatternConfig{
				{
					Match: "service-*",
					Settings: &RepoSettings{
						Topics: []string{"microservice"},
					},
				},
			},
		},
	}

	// Test specific repository
	settings, security, _, _, err := config.GetEffectiveConfig("api-gateway")
	require.NoError(t, err)

	assert.NotNil(t, settings)
	assert.Equal(t, "API Gateway Service", *settings.Description)
	assert.True(t, *settings.Private)      // From defaults
	assert.True(t, *settings.HasIssues)    // From defaults
	assert.False(t, *settings.HasWiki)     // From template
	assert.False(t, *settings.HasProjects) // From template

	assert.NotNil(t, security)
	assert.True(t, *security.VulnerabilityAlerts) // From template

	// Test pattern matching
	settings2, _, _, _, err := config.GetEffectiveConfig("service-auth")
	require.NoError(t, err)

	assert.NotNil(t, settings2)
	assert.Contains(t, settings2.Topics, "microservice")
}

func TestMergeRepoSettings(t *testing.T) {
	base := &RepoSettings{
		Private:   boolPtr(true),
		HasIssues: boolPtr(true),
		Topics:    []string{"base"},
	}

	override := &RepoSettings{
		Private:     boolPtr(false),
		HasProjects: boolPtr(true),
		Topics:      []string{"override"},
	}

	result := mergeRepoSettings(base, override)

	assert.False(t, *result.Private)                     // Overridden
	assert.True(t, *result.HasIssues)                    // From base
	assert.True(t, *result.HasProjects)                  // From override
	assert.Equal(t, []string{"override"}, result.Topics) // Overridden
}

func TestMergeSecuritySettings(t *testing.T) {
	base := &SecuritySettings{
		VulnerabilityAlerts: boolPtr(true),
		BranchProtection: map[string]*BranchProtectionRule{
			"main": {
				RequiredReviews: intPtr(1),
			},
		},
	}

	override := &SecuritySettings{
		SecurityAdvisories: boolPtr(true),
		BranchProtection: map[string]*BranchProtectionRule{
			"main": {
				RequiredReviews: intPtr(2),
			},
			"develop": {
				RequiredReviews: intPtr(1),
			},
		},
	}

	result := mergeSecuritySettings(base, override)

	assert.True(t, *result.VulnerabilityAlerts)                             // From base
	assert.True(t, *result.SecurityAdvisories)                              // From override
	assert.Equal(t, 2, *result.BranchProtection["main"].RequiredReviews)    // Overridden
	assert.Equal(t, 1, *result.BranchProtection["develop"].RequiredReviews) // From override
}

func TestMatchPattern(t *testing.T) {
	tests := []struct {
		str     string
		pattern string
		want    bool
	}{
		{"service-auth", "service-*", true},
		{"api-service", "service-*", false},
		{"service-auth", "service-auth", true},
		{"lib-common", "lib-*", true},
		{"common-lib", "*-lib", true},
	}

	for _, tt := range tests {
		t.Run(tt.str+" vs "+tt.pattern, func(t *testing.T) {
			got, err := matchPattern(tt.str, tt.pattern)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestExpandEnvVars(t *testing.T) {
	// Set test environment variables
	_ = os.Setenv("TEST_WEBHOOK_URL", "https://example.com/webhook") //nolint:errcheck // Test environment setup
	_ = os.Setenv("TEST_SECRET", "secret123")                        //nolint:errcheck // Test environment setup

	defer func() {
		_ = os.Unsetenv("TEST_WEBHOOK_URL") //nolint:errcheck // Test cleanup
		_ = os.Unsetenv("TEST_SECRET")      //nolint:errcheck // Test cleanup
	}()

	config := &RepoConfig{
		Templates: map[string]*RepoTemplate{
			"test": {
				Security: &SecuritySettings{
					Webhooks: []WebhookConfig{
						{
							URL:    "${TEST_WEBHOOK_URL}",
							Secret: "${TEST_SECRET}",
						},
					},
				},
			},
		},
	}

	err := expandRepoConfigEnvVars(config)
	require.NoError(t, err)

	webhook := config.Templates["test"].Security.Webhooks[0]
	assert.Equal(t, "https://example.com/webhook", webhook.URL)
	assert.Equal(t, "secret123", webhook.Secret)
}

// Helper functions.
func boolPtr(b bool) *bool {
	return &b
}

func intPtr(i int) *int {
	return &i
}

func stringPtr(s string) *string {
	return &s
}
