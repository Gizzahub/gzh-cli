package config

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStartupValidator_ValidateUnifiedConfig(t *testing.T) {
	validator := NewStartupValidator()

	tests := []struct {
		name          string
		config        *UnifiedConfig
		expectedValid bool
		expectedErrors int
		expectedWarnings int
		setupEnv      map[string]string
	}{
		{
			name: "valid complete configuration",
			config: &UnifiedConfig{
				Version:         "1.0.0",
				DefaultProvider: "github",
				Providers: map[string]*ProviderConfig{
					"github": {
						Token: "${GITHUB_TOKEN}",
						Organizations: []*OrganizationConfig{
							{
								Name:       "test-org",
								CloneDir:   "~/repos/test-org",
								Visibility: "all",
								Strategy:   "reset",
							},
						},
					},
				},
			},
			expectedValid:    true,
			expectedErrors:   0,
			expectedWarnings: 1, // Warning about missing environment variable
		},
		{
			name: "invalid version format",
			config: &UnifiedConfig{
				Version:         "invalid-version",
				DefaultProvider: "github",
				Providers: map[string]*ProviderConfig{
					"github": {
						Token: "test-token",
						Organizations: []*OrganizationConfig{
							{
								Name:     "test-org",
								CloneDir: "~/repos/test-org",
							},
						},
					},
				},
			},
			expectedValid:  false,
			expectedErrors: 1,
		},
		{
			name: "missing required fields",
			config: &UnifiedConfig{
				// Missing version and providers
			},
			expectedValid:  false,
			expectedErrors: 2, // Missing version and providers
		},
		{
			name: "invalid provider in default",
			config: &UnifiedConfig{
				Version:         "1.0.0",
				DefaultProvider: "nonexistent",
				Providers: map[string]*ProviderConfig{
					"github": {
						Token: "test-token",
						Organizations: []*OrganizationConfig{
							{
								Name:     "test-org",
								CloneDir: "~/repos/test-org",
							},
						},
					},
				},
			},
			expectedValid:  false,
			expectedErrors: 1, // Default provider doesn't exist
		},
		{
			name: "invalid regex patterns",
			config: &UnifiedConfig{
				Version:         "1.0.0",
				DefaultProvider: "github",
				Providers: map[string]*ProviderConfig{
					"github": {
						Token: "test-token",
						Organizations: []*OrganizationConfig{
							{
								Name:     "test-org",
								CloneDir: "~/repos/test-org",
								Include:  "[invalid-regex",
								Exclude:  []string{"[another-invalid"},
							},
						},
					},
				},
			},
			expectedValid:  false,
			expectedErrors: 2, // Invalid include and exclude patterns
		},
		{
			name: "configuration with warnings",
			config: &UnifiedConfig{
				Version:         "1.0.0",
				DefaultProvider: "github",
				Global: &GlobalSettings{
					Concurrency: &ConcurrencySettings{
						CloneWorkers: 100, // Will trigger warning
					},
					Timeouts: &TimeoutSettings{
						HTTPTimeout: 500 * time.Millisecond, // Will trigger warning
					},
				},
				Providers: map[string]*ProviderConfig{
					"github": {
						Token: "${GITHUB_TOKEN}",
						Organizations: []*OrganizationConfig{
							{
								Name:     "test-org",
								CloneDir: "~/repos/test-org",
							},
						},
					},
				},
			},
			expectedValid:    true,
			expectedErrors:   0,
			expectedWarnings: 3, // High concurrency, short timeout, missing env var
		},
		{
			name: "valid configuration with environment variable",
			config: &UnifiedConfig{
				Version:         "1.0.0",
				DefaultProvider: "github",
				Providers: map[string]*ProviderConfig{
					"github": {
						Token: "${GITHUB_TOKEN}",
						Organizations: []*OrganizationConfig{
							{
								Name:     "test-org",
								CloneDir: "~/repos/test-org",
							},
						},
					},
				},
			},
			setupEnv: map[string]string{
				"GITHUB_TOKEN": "test-token-value",
			},
			expectedValid:    true,
			expectedErrors:   0,
			expectedWarnings: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment variables
			if tt.setupEnv != nil {
				for key, value := range tt.setupEnv {
					os.Setenv(key, value)
					defer os.Unsetenv(key)
				}
			}

			result := validator.ValidateUnifiedConfig(tt.config)

			assert.Equal(t, tt.expectedValid, result.IsValid, "IsValid mismatch")
			assert.Len(t, result.Errors, tt.expectedErrors, "Error count mismatch")
			assert.Len(t, result.Warnings, tt.expectedWarnings, "Warning count mismatch")
			assert.NotEmpty(t, result.Summary, "Summary should not be empty")

			// Print details for debugging
			if len(result.Errors) > 0 {
				t.Logf("Validation errors:")
				for _, err := range result.Errors {
					t.Logf("  - %s: %s", err.Field, err.Message)
				}
			}
			if len(result.Warnings) > 0 {
				t.Logf("Validation warnings:")
				for _, warn := range result.Warnings {
					t.Logf("  - %s: %s", warn.Field, warn.Message)
				}
			}
		})
	}
}

func TestStartupValidator_ValidateConfig(t *testing.T) {
	validator := NewStartupValidator()

	tests := []struct {
		name          string
		config        *Config
		expectedValid bool
		expectedErrors int
	}{
		{
			name: "valid config",
			config: &Config{
				Version:         "1.0.0",
				DefaultProvider: "github",
				Providers: map[string]Provider{
					"github": {
						Token: "test-token",
						Orgs: []GitTarget{
							{
								Name:       "test-org",
								Visibility: "all",
								Strategy:   "reset",
								CloneDir:   "~/repos/test-org",
							},
						},
					},
				},
			},
			expectedValid:  true,
			expectedErrors: 0,
		},
		{
			name: "missing version",
			config: &Config{
				DefaultProvider: "github",
				Providers: map[string]Provider{
					"github": {
						Token: "test-token",
						Orgs:  []GitTarget{},
					},
				},
			},
			expectedValid:  false,
			expectedErrors: 1,
		},
		{
			name: "invalid target configuration",
			config: &Config{
				Version:         "1.0.0",
				DefaultProvider: "github",
				Providers: map[string]Provider{
					"github": {
						Token: "test-token",
						Orgs: []GitTarget{
							{
								// Missing name
								Visibility: "invalid-visibility",
								Strategy:   "invalid-strategy",
							},
						},
					},
				},
			},
			expectedValid:  false,
			expectedErrors: 3, // Missing name, invalid visibility, invalid strategy
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateConfig(tt.config)

			assert.Equal(t, tt.expectedValid, result.IsValid, "IsValid mismatch")
			assert.Len(t, result.Errors, tt.expectedErrors, "Error count mismatch")
			assert.NotEmpty(t, result.Summary, "Summary should not be empty")
		})
	}
}

func TestStartupValidator_CustomValidators(t *testing.T) {
	validator := NewStartupValidator()

	t.Run("strategy validation", func(t *testing.T) {
		validStrategies := []string{"reset", "pull", "fetch"}
		invalidStrategies := []string{"invalid", "push", ""}

		for _, strategy := range validStrategies {
			assert.True(t, isValidStrategy(strategy), "Strategy %s should be valid", strategy)
		}

		for _, strategy := range invalidStrategies {
			assert.False(t, isValidStrategy(strategy), "Strategy %s should be invalid", strategy)
		}
	})

	t.Run("provider validation", func(t *testing.T) {
		validProviders := []string{"github", "gitlab", "gitea", "gogs"}
		invalidProviders := []string{"invalid", "bitbucket", ""}

		for _, provider := range validProviders {
			assert.True(t, isValidProvider(provider), "Provider %s should be valid", provider)
		}

		for _, provider := range invalidProviders {
			assert.False(t, isValidProvider(provider), "Provider %s should be invalid", provider)
		}
	})

	t.Run("visibility validation", func(t *testing.T) {
		validVisibilities := []string{"public", "private", "all"}
		invalidVisibilities := []string{"invalid", "internal", ""}

		for _, visibility := range validVisibilities {
			assert.True(t, isValidVisibility(visibility), "Visibility %s should be valid", visibility)
		}

		for _, visibility := range invalidVisibilities {
			assert.False(t, isValidVisibility(visibility), "Visibility %s should be invalid", visibility)
		}
	})

	t.Run("version format validation", func(t *testing.T) {
		validVersions := []string{"1.0.0", "2.1.3", "1.0.0-alpha", "1.0.0+build.1"}
		invalidVersions := []string{"1.0", "invalid", "1.0.0.0", ""}

		for _, version := range validVersions {
			assert.True(t, isValidVersionFormat(version), "Version %s should be valid", version)
		}

		for _, version := range invalidVersions {
			assert.False(t, isValidVersionFormat(version), "Version %s should be invalid", version)
		}
	})
}

func TestStartupValidator_ValidationErrors(t *testing.T) {
	validator := NewStartupValidator()

	config := &UnifiedConfig{
		Version: "invalid-version",
		Providers: map[string]*ProviderConfig{
			"github": {
				// Missing token
				Organizations: []*OrganizationConfig{
					{
						// Missing name and clone_dir
						Visibility: "invalid-visibility",
						Strategy:   "invalid-strategy",
						Include:    "[invalid-regex",
					},
				},
			},
		},
	}

	result := validator.ValidateUnifiedConfig(config)

	assert.False(t, result.IsValid)
	assert.True(t, len(result.Errors) > 0)

	// Check that we have various types of errors
	errorFields := make(map[string]bool)
	for _, err := range result.Errors {
		errorFields[err.Tag] = true
	}

	// Should have errors for missing required fields, invalid values, etc.
	assert.True(t, len(errorFields) > 1, "Should have multiple types of validation errors")
}

func TestStartupValidator_EnvironmentVariableWarnings(t *testing.T) {
	validator := NewStartupValidator()

	// Test with missing environment variable
	config := &UnifiedConfig{
		Version:         "1.0.0",
		DefaultProvider: "github",
		Providers: map[string]*ProviderConfig{
			"github": {
				Token: "${MISSING_TOKEN}",
				Organizations: []*OrganizationConfig{
					{
						Name:     "test-org",
						CloneDir: "~/repos/test-org",
					},
				},
			},
		},
	}

	result := validator.ValidateUnifiedConfig(config)

	assert.True(t, result.IsValid, "Config should be valid even with missing env var")
	assert.True(t, len(result.Warnings) > 0, "Should have warnings about missing env var")

	// Check for environment variable warning
	foundEnvWarning := false
	for _, warning := range result.Warnings {
		if strings.Contains(warning.Message, "MISSING_TOKEN") {
			foundEnvWarning = true
			break
		}
	}
	assert.True(t, foundEnvWarning, "Should have warning about missing environment variable")
}

func TestStartupValidator_ConcurrencyAndTimeoutWarnings(t *testing.T) {
	validator := NewStartupValidator()

	config := &UnifiedConfig{
		Version:         "1.0.0",
		DefaultProvider: "github",
		Global: &GlobalSettings{
			Timeouts: &TimeoutSettings{
				HTTPTimeout:      500 * time.Millisecond, // Too short
				GitTimeout:       10 * time.Second,       // Too short
				RateLimitTimeout: 1 * time.Minute,        // Too short
			},
			Concurrency: &ConcurrencySettings{
				CloneWorkers:  100, // Too high
				UpdateWorkers: 80,  // Too high
				APIWorkers:    25,  // Too high
			},
		},
		Providers: map[string]*ProviderConfig{
			"github": {
				Token: "test-token",
				Organizations: []*OrganizationConfig{
					{
						Name:     "test-org",
						CloneDir: "~/repos/test-org",
					},
				},
			},
		},
	}

	result := validator.ValidateUnifiedConfig(config)

	assert.True(t, result.IsValid, "Config should be valid")
	assert.Equal(t, 6, len(result.Warnings), "Should have 6 warnings (3 timeout + 3 concurrency)")

	// Check for specific warnings
	warningMessages := make([]string, len(result.Warnings))
	for i, warning := range result.Warnings {
		warningMessages[i] = warning.Message
	}

	assert.Contains(t, strings.Join(warningMessages, " "), "HTTP timeout")
	assert.Contains(t, strings.Join(warningMessages, " "), "Git timeout")
	assert.Contains(t, strings.Join(warningMessages, " "), "Rate limit timeout")
	assert.Contains(t, strings.Join(warningMessages, " "), "clone worker")
	assert.Contains(t, strings.Join(warningMessages, " "), "update worker")
	assert.Contains(t, strings.Join(warningMessages, " "), "API worker")
}

func TestStartupValidator_NoOrganizationsWarning(t *testing.T) {
	validator := NewStartupValidator()

	config := &UnifiedConfig{
		Version:         "1.0.0",
		DefaultProvider: "github",
		Providers: map[string]*ProviderConfig{
			"github": {
				Token:         "test-token",
				Organizations: []*OrganizationConfig{}, // Empty organizations
			},
		},
	}

	result := validator.ValidateUnifiedConfig(config)

	assert.True(t, result.IsValid, "Config should be valid")
	assert.True(t, len(result.Warnings) > 0, "Should have warnings")

	// Check for no organizations warning
	foundNoOrgsWarning := false
	for _, warning := range result.Warnings {
		if strings.Contains(warning.Message, "No organizations configured") {
			foundNoOrgsWarning = true
			break
		}
	}
	assert.True(t, foundNoOrgsWarning, "Should have warning about no organizations")
}