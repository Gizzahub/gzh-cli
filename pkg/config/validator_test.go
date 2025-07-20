//nolint:testpackage // White-box testing needed for internal function access
package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidator_ValidateConfig(t *testing.T) {
	tests := []struct {
		name         string
		config       *Config
		expectError  bool
		expectWarn   bool
		errorCount   int
		warningCount int
	}{
		{
			name: "valid configuration",
			config: &Config{
				Version:         "1.0.0",
				DefaultProvider: "github",
				Providers: map[string]Provider{
					"github": {
						Token: "${GITHUB_TOKEN}",
						Orgs: []GitTarget{
							{
								Name:       "test-org",
								Visibility: "public",
								Strategy:   "reset",
							},
						},
					},
				},
			},
			expectError: false,
			expectWarn:  false,
		},
		{
			name: "missing version",
			config: &Config{
				Providers: map[string]Provider{
					"github": {
						Token: "test",
						Orgs:  []GitTarget{{Name: "test"}},
					},
				},
			},
			expectError: true,
			errorCount:  1,
		},
		{
			name: "invalid version format",
			config: &Config{
				Version: "1.0",
				Providers: map[string]Provider{
					"github": {
						Token: "test",
						Orgs:  []GitTarget{{Name: "test"}},
					},
				},
			},
			expectError: true,
			errorCount:  1,
		},
		{
			name: "invalid provider name",
			config: &Config{
				Version: "1.0.0",
				Providers: map[string]Provider{
					"invalid": {
						Token: "test",
						Orgs:  []GitTarget{{Name: "test"}},
					},
				},
			},
			expectError: true,
			errorCount:  1,
		},
		{
			name: "missing token",
			config: &Config{
				Version: "1.0.0",
				Providers: map[string]Provider{
					"github": {
						Orgs: []GitTarget{{Name: "test"}},
					},
				},
			},
			expectError: true,
			errorCount:  1,
		},
		{
			name: "invalid visibility",
			config: &Config{
				Version: "1.0.0",
				Providers: map[string]Provider{
					"github": {
						Token: "test",
						Orgs: []GitTarget{
							{
								Name:       "test",
								Visibility: "invalid",
							},
						},
					},
				},
			},
			expectError: true,
			errorCount:  1,
		},
		{
			name: "invalid regex pattern",
			config: &Config{
				Version: "1.0.0",
				Providers: map[string]Provider{
					"github": {
						Token: "test",
						Orgs: []GitTarget{
							{
								Name:  "test",
								Match: "[invalid",
							},
						},
					},
				},
			},
			expectError: true,
			errorCount:  1,
		},
		{
			name: "configuration with warnings",
			config: &Config{
				Version: "1.0.0",
				Providers: map[string]Provider{
					"github": {
						Token: "short", // Should trigger warning
						Orgs: []GitTarget{
							{
								Name:     "test",
								CloneDir: "../unsafe", // Should trigger warning
							},
						},
					},
				},
			},
			expectError: false,
			expectWarn:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewValidator()
			err := validator.ValidateConfig(tt.config)

			if tt.expectError {
				assert.Error(t, err)

				if tt.errorCount > 0 {
					assert.Len(t, validator.errors, tt.errorCount)
				}
			} else {
				assert.NoError(t, err)
			}

			if tt.expectWarn {
				assert.Greater(t, len(validator.GetWarnings()), 0)
			}

			if tt.warningCount > 0 {
				assert.Len(t, validator.GetWarnings(), tt.warningCount)
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	tests := []struct {
		name         string
		providerName string
		token        string
		expectWarn   bool
	}{
		{
			name:         "valid GitHub token",
			providerName: "github",
			token:        "ghp_1234567890abcdef",
			expectWarn:   false,
		},
		{
			name:         "valid environment variable",
			providerName: "github",
			token:        "${GITHUB_TOKEN}",
			expectWarn:   false,
		},
		{
			name:         "short token",
			providerName: "github",
			token:        "short",
			expectWarn:   true,
		},
		{
			name:         "GitHub token wrong format",
			providerName: "github",
			token:        "random_token_1234567890",
			expectWarn:   true,
		},
		{
			name:         "GitLab token correct format",
			providerName: "gitlab",
			token:        "glpat-1234567890abcdef",
			expectWarn:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewValidator()
			validator.validateToken(tt.providerName, tt.token)

			if tt.expectWarn {
				assert.Greater(t, len(validator.GetWarnings()), 0)
			} else {
				assert.Len(t, validator.GetWarnings(), 0)
			}
		})
	}
}

func TestValidateGitTarget(t *testing.T) {
	tests := []struct {
		name        string
		target      GitTarget
		expectError bool
		expectWarn  bool
	}{
		{
			name: "valid target",
			target: GitTarget{
				Name:       "test-org",
				Visibility: "public",
				Strategy:   "reset",
				Match:      "^test-.*",
				CloneDir:   "./safe",
			},
			expectError: false,
			expectWarn:  false,
		},
		{
			name:        "missing name",
			target:      GitTarget{},
			expectError: true,
		},
		{
			name: "invalid visibility",
			target: GitTarget{
				Name:       "test",
				Visibility: "invalid",
			},
			expectError: true,
		},
		{
			name: "invalid strategy",
			target: GitTarget{
				Name:     "test",
				Strategy: "invalid",
			},
			expectError: true,
		},
		{
			name: "invalid regex",
			target: GitTarget{
				Name:  "test",
				Match: "[invalid",
			},
			expectError: true,
		},
		{
			name: "unsafe clone dir",
			target: GitTarget{
				Name:     "test",
				CloneDir: "../unsafe",
			},
			expectError: false,
			expectWarn:  true,
		},
		{
			name: "complex glob pattern",
			target: GitTarget{
				Name:    "test",
				Exclude: []string{"*/*/*/*/*"},
			},
			expectError: false,
			expectWarn:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewValidator()
			validator.validateGitTarget("test", tt.target)

			if tt.expectError {
				assert.Greater(t, len(validator.errors), 0)
			} else {
				assert.Len(t, validator.errors, 0)
			}

			if tt.expectWarn {
				assert.Greater(t, len(validator.GetWarnings()), 0)
			}
		})
	}
}

func TestValidationResult(t *testing.T) {
	tests := []struct {
		name            string
		result          ValidationResult
		expectHasIssues bool
	}{
		{
			name: "no issues",
			result: ValidationResult{
				Valid:    true,
				Errors:   []string{},
				Warnings: []string{},
			},
			expectHasIssues: false,
		},
		{
			name: "has errors",
			result: ValidationResult{
				Valid:  false,
				Errors: []string{"error1"},
			},
			expectHasIssues: true,
		},
		{
			name: "has warnings",
			result: ValidationResult{
				Valid:    true,
				Warnings: []string{"warning1"},
			},
			expectHasIssues: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectHasIssues, tt.result.HasIssues())
		})
	}
}
