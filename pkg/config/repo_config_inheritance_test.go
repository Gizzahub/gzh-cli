package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplateInheritance(t *testing.T) {
	tests := []struct {
		name     string
		config   *RepoConfig
		template string
		validate func(t *testing.T, resolved *RepoTemplate)
	}{
		{
			name: "single level inheritance",
			config: &RepoConfig{
				Templates: map[string]*RepoTemplate{
					"base": {
						Settings: &RepoSettings{
							Private:   boolPtr(true),
							HasIssues: boolPtr(true),
						},
						Security: &SecuritySettings{
							VulnerabilityAlerts: boolPtr(true),
						},
					},
					"derived": {
						Base: "base",
						Settings: &RepoSettings{
							HasWiki: boolPtr(false),
						},
					},
				},
			},
			template: "derived",
			validate: func(t *testing.T, resolved *RepoTemplate) {
				require.NotNil(t, resolved.Settings)
				assert.True(t, *resolved.Settings.Private)             // From base
				assert.True(t, *resolved.Settings.HasIssues)           // From base
				assert.False(t, *resolved.Settings.HasWiki)            // From derived
				assert.True(t, *resolved.Security.VulnerabilityAlerts) // From base
			},
		},
		{
			name: "multi-level inheritance",
			config: &RepoConfig{
				Templates: map[string]*RepoTemplate{
					"base": {
						Settings: &RepoSettings{
							Private: boolPtr(true),
						},
					},
					"middle": {
						Base: "base",
						Settings: &RepoSettings{
							HasIssues: boolPtr(true),
						},
						Security: &SecuritySettings{
							VulnerabilityAlerts: boolPtr(true),
						},
					},
					"derived": {
						Base: "middle",
						Settings: &RepoSettings{
							HasWiki: boolPtr(false),
						},
					},
				},
			},
			template: "derived",
			validate: func(t *testing.T, resolved *RepoTemplate) {
				require.NotNil(t, resolved.Settings)
				assert.True(t, *resolved.Settings.Private)             // From base
				assert.True(t, *resolved.Settings.HasIssues)           // From middle
				assert.False(t, *resolved.Settings.HasWiki)            // From derived
				assert.True(t, *resolved.Security.VulnerabilityAlerts) // From middle
			},
		},
		{
			name: "override precedence",
			config: &RepoConfig{
				Templates: map[string]*RepoTemplate{
					"base": {
						Settings: &RepoSettings{
							Private:   boolPtr(true),
							HasIssues: boolPtr(false),
						},
					},
					"derived": {
						Base: "base",
						Settings: &RepoSettings{
							Private: boolPtr(false), // Override
						},
					},
				},
			},
			template: "derived",
			validate: func(t *testing.T, resolved *RepoTemplate) {
				require.NotNil(t, resolved.Settings)
				assert.False(t, *resolved.Settings.Private)   // Overridden
				assert.False(t, *resolved.Settings.HasIssues) // From base
			},
		},
		{
			name: "branch protection merging",
			config: &RepoConfig{
				Templates: map[string]*RepoTemplate{
					"base": {
						Security: &SecuritySettings{
							BranchProtection: map[string]*BranchProtectionRule{
								"main": {
									RequiredReviews: intPtr(1),
									EnforceAdmins:   boolPtr(false),
								},
							},
						},
					},
					"derived": {
						Base: "base",
						Security: &SecuritySettings{
							BranchProtection: map[string]*BranchProtectionRule{
								"main": {
									RequiredReviews: intPtr(2), // Override
								},
								"develop": {
									RequiredReviews: intPtr(1),
								},
							},
						},
					},
				},
			},
			template: "derived",
			validate: func(t *testing.T, resolved *RepoTemplate) {
				require.NotNil(t, resolved.Security)
				require.NotNil(t, resolved.Security.BranchProtection)

				mainRule := resolved.Security.BranchProtection["main"]
				require.NotNil(t, mainRule)
				assert.Equal(t, 2, *mainRule.RequiredReviews) // Overridden
				assert.False(t, *mainRule.EnforceAdmins)      // From base

				developRule := resolved.Security.BranchProtection["develop"]
				require.NotNil(t, developRule)
				assert.Equal(t, 1, *developRule.RequiredReviews) // From derived
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolved, err := tt.config.resolveTemplate(tt.template)
			require.NoError(t, err)
			tt.validate(t, resolved)
		})
	}
}

func TestTemplateInheritanceErrors(t *testing.T) {
	tests := []struct {
		name        string
		config      *RepoConfig
		template    string
		expectedErr string
	}{
		{
			name: "circular dependency - self reference",
			config: &RepoConfig{
				Templates: map[string]*RepoTemplate{
					"self": {
						Base: "self",
					},
				},
			},
			template:    "self",
			expectedErr: "circular template dependency detected",
		},
		{
			name: "circular dependency - two templates",
			config: &RepoConfig{
				Templates: map[string]*RepoTemplate{
					"a": {Base: "b"},
					"b": {Base: "a"},
				},
			},
			template:    "a",
			expectedErr: "circular template dependency detected",
		},
		{
			name: "circular dependency - three templates",
			config: &RepoConfig{
				Templates: map[string]*RepoTemplate{
					"a": {Base: "b"},
					"b": {Base: "c"},
					"c": {Base: "a"},
				},
			},
			template:    "a",
			expectedErr: "circular template dependency detected",
		},
		{
			name: "template not found",
			config: &RepoConfig{
				Templates: map[string]*RepoTemplate{
					"exists": {Base: "missing"},
				},
			},
			template:    "exists",
			expectedErr: "template 'missing' not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.config.resolveTemplate(tt.template)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestValidateTemplateOverrides(t *testing.T) {
	config := &RepoConfig{
		Templates: map[string]*RepoTemplate{
			"secure-base": {
				Settings: &RepoSettings{
					Private: boolPtr(true),
				},
				Security: &SecuritySettings{
					BranchProtection: map[string]*BranchProtectionRule{
						"main": {
							RequiredReviews: intPtr(3),
							EnforceAdmins:   boolPtr(true),
						},
					},
				},
				Permissions: &PermissionSettings{
					TeamPermissions: map[string]string{
						"dev-team": "write",
					},
				},
			},
			"less-secure": {
				Base: "secure-base",
				Settings: &RepoSettings{
					Private: boolPtr(false), // Makes it public
				},
				Security: &SecuritySettings{
					BranchProtection: map[string]*BranchProtectionRule{
						"main": {
							RequiredReviews: intPtr(1),      // Reduces reviews
							EnforceAdmins:   boolPtr(false), // Disables admin enforcement
						},
					},
				},
				Permissions: &PermissionSettings{
					TeamPermissions: map[string]string{
						"dev-team": "admin", // Escalates permissions
					},
				},
			},
		},
	}

	warnings := config.ValidateTemplateOverrides()

	// Should have warnings for security downgrades
	assert.NotEmpty(t, warnings)

	// Check specific warnings exist
	warningsStr := strings.Join(warnings, "\n")
	assert.Contains(t, warningsStr, "Changes repository from private to public")
	assert.Contains(t, warningsStr, "Reduces required reviews")
	assert.Contains(t, warningsStr, "Disables admin enforcement")
	assert.Contains(t, warningsStr, "Escalates permissions")
}

func TestGetTemplateInheritanceChain(t *testing.T) {
	config := &RepoConfig{
		Templates: map[string]*RepoTemplate{
			"base":     {},
			"middle":   {Base: "base"},
			"derived":  {Base: "middle"},
			"isolated": {},
		},
	}

	tests := []struct {
		name     string
		template string
		expected []string
	}{
		{
			name:     "no inheritance",
			template: "base",
			expected: []string{"base"},
		},
		{
			name:     "single inheritance",
			template: "middle",
			expected: []string{"middle", "base"},
		},
		{
			name:     "multi-level inheritance",
			template: "derived",
			expected: []string{"derived", "middle", "base"},
		},
		{
			name:     "isolated template",
			template: "isolated",
			expected: []string{"isolated"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chain, err := config.GetTemplateInheritanceChain(tt.template)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, chain)
		})
	}
}

func TestGetAllTemplateChains(t *testing.T) {
	config := &RepoConfig{
		Templates: map[string]*RepoTemplate{
			"base":    {},
			"middle":  {Base: "base"},
			"derived": {Base: "middle"},
		},
	}

	chains := config.GetAllTemplateChains()

	assert.Len(t, chains, 3)
	assert.Equal(t, []string{"base"}, chains["base"])
	assert.Equal(t, []string{"middle", "base"}, chains["middle"])
	assert.Equal(t, []string{"derived", "middle", "base"}, chains["derived"])
}

func TestEffectiveConfigWithInheritance(t *testing.T) {
	config := &RepoConfig{
		Version:      "1.0.0",
		Organization: "test-org",
		Templates: map[string]*RepoTemplate{
			"base": {
				Settings: &RepoSettings{
					Private:          boolPtr(true),
					HasIssues:        boolPtr(true),
					AllowSquashMerge: boolPtr(true),
				},
			},
			"secure": {
				Base: "base",
				Settings: &RepoSettings{
					HasWiki: boolPtr(false),
				},
				Security: &SecuritySettings{
					VulnerabilityAlerts: boolPtr(true),
					BranchProtection: map[string]*BranchProtectionRule{
						"main": {
							RequiredReviews: intPtr(2),
						},
					},
				},
			},
		},
		Repositories: &RepoTargets{
			Specific: []RepoSpecificConfig{
				{
					Name:     "test-repo",
					Template: "secure",
				},
			},
		},
	}

	settings, security, _, err := config.GetEffectiveConfig("test-repo")
	require.NoError(t, err)

	// Should have merged settings from base and secure templates
	require.NotNil(t, settings)
	assert.True(t, *settings.Private)          // From base
	assert.True(t, *settings.HasIssues)        // From base
	assert.True(t, *settings.AllowSquashMerge) // From base
	assert.False(t, *settings.HasWiki)         // From secure

	// Should have security settings from secure template
	require.NotNil(t, security)
	assert.True(t, *security.VulnerabilityAlerts)
	assert.Equal(t, 2, *security.BranchProtection["main"].RequiredReviews)
}
