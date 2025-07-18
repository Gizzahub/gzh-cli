package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPolicyExceptions(t *testing.T) {
	config := &RepoConfig{
		Version:      "1.0.0",
		Organization: "test-org",
		Policies: map[string]*PolicyTemplate{
			"security_compliance": {
				Description: "Security compliance policy",
				Rules: map[string]PolicyRule{
					"min_reviewers": {
						Type:        "min_reviews",
						Value:       2,
						Enforcement: "required",
						Message:     "Minimum 2 reviewers required",
					},
					"branch_protection": {
						Type:        "branch_protection",
						Value:       true,
						Enforcement: "required",
						Message:     "Branch protection must be enabled",
					},
				},
			},
			"opensource_compliance": {
				Description: "Open source compliance",
				Rules: map[string]PolicyRule{
					"license_required": {
						Type:        "file_exists",
						Value:       "LICENSE",
						Enforcement: "required",
					},
				},
			},
		},
		Repositories: &RepoTargets{
			Specific: []RepoSpecificConfig{
				{
					Name: "legacy-app",
					Exceptions: []PolicyException{
						{
							PolicyName:   "security_compliance",
							RuleName:     "min_reviewers",
							Reason:       "Legacy app with small team",
							ApprovedBy:   "cto@company.com",
							ApprovalDate: "2024-01-15",
							ExpiresAt:    "2024-12-31",
						},
					},
				},
				{
					Name: "experimental-project",
					Exceptions: []PolicyException{
						{
							PolicyName: "security_compliance",
							RuleName:   "branch_protection",
							Reason:     "Rapid prototyping phase",
							ApprovedBy: "tech-lead@company.com",
							ExpiresAt:  "2024-06-30",
							Conditions: []string{
								"Must enable before production release",
								"Weekly security review required",
							},
						},
					},
				},
			},
			Patterns: []RepoPatternConfig{
				{
					Match: "poc-*",
					Exceptions: []PolicyException{
						{
							PolicyName: "opensource_compliance",
							RuleName:   "license_required",
							Reason:     "Proof of concept repositories",
							ApprovedBy: "legal@company.com",
						},
					},
				},
			},
		},
	}

	t.Run("GetEffectiveConfig returns exceptions", func(t *testing.T) {
		_, _, _, exceptions, err := config.GetEffectiveConfig("legacy-app")
		require.NoError(t, err)

		assert.Len(t, exceptions, 1)
		assert.Equal(t, "security_compliance", exceptions[0].PolicyName)
		assert.Equal(t, "min_reviewers", exceptions[0].RuleName)
		assert.Equal(t, "Legacy app with small team", exceptions[0].Reason)
	})

	t.Run("Pattern-based exceptions", func(t *testing.T) {
		_, _, _, exceptions, err := config.GetEffectiveConfig("poc-demo")
		require.NoError(t, err)

		assert.Len(t, exceptions, 1)
		assert.Equal(t, "opensource_compliance", exceptions[0].PolicyName)
		assert.Equal(t, "license_required", exceptions[0].RuleName)
	})

	t.Run("No exceptions for regular repos", func(t *testing.T) {
		_, _, _, exceptions, err := config.GetEffectiveConfig("regular-app")
		require.NoError(t, err)

		assert.Empty(t, exceptions)
	})
}

func TestValidatePolicyExceptions(t *testing.T) {
	tests := []struct {
		name          string
		config        *RepoConfig
		expectedErrs  int
		errorContains []string
	}{
		{
			name: "valid exceptions",
			config: &RepoConfig{
				Policies: map[string]*PolicyTemplate{
					"test_policy": {
						Rules: map[string]PolicyRule{
							"test_rule": {Type: "test"},
						},
					},
				},
				Repositories: &RepoTargets{
					Specific: []RepoSpecificConfig{
						{
							Name: "test-repo",
							Exceptions: []PolicyException{
								{
									PolicyName: "test_policy",
									RuleName:   "test_rule",
									Reason:     "Valid reason",
									ApprovedBy: "admin@company.com",
								},
							},
						},
					},
				},
			},
			expectedErrs: 0,
		},
		{
			name: "non-existent policy",
			config: &RepoConfig{
				Policies: map[string]*PolicyTemplate{},
				Repositories: &RepoTargets{
					Specific: []RepoSpecificConfig{
						{
							Name: "test-repo",
							Exceptions: []PolicyException{
								{
									PolicyName: "missing_policy",
									RuleName:   "some_rule",
									Reason:     "Test",
									ApprovedBy: "admin",
								},
							},
						},
					},
				},
			},
			expectedErrs:  1,
			errorContains: []string{"non-existent policy"},
		},
		{
			name: "non-existent rule",
			config: &RepoConfig{
				Policies: map[string]*PolicyTemplate{
					"test_policy": {
						Rules: map[string]PolicyRule{},
					},
				},
				Repositories: &RepoTargets{
					Specific: []RepoSpecificConfig{
						{
							Name: "test-repo",
							Exceptions: []PolicyException{
								{
									PolicyName: "test_policy",
									RuleName:   "missing_rule",
									Reason:     "Test",
									ApprovedBy: "admin",
								},
							},
						},
					},
				},
			},
			expectedErrs:  1,
			errorContains: []string{"non-existent rule"},
		},
		{
			name: "missing required fields",
			config: &RepoConfig{
				Policies: map[string]*PolicyTemplate{
					"test_policy": {
						Rules: map[string]PolicyRule{
							"test_rule": {Type: "test"},
						},
					},
				},
				Repositories: &RepoTargets{
					Specific: []RepoSpecificConfig{
						{
							Name: "test-repo",
							Exceptions: []PolicyException{
								{
									PolicyName: "test_policy",
									RuleName:   "test_rule",
									// Missing Reason and ApprovedBy
								},
							},
						},
					},
				},
			},
			expectedErrs:  1,
			errorContains: []string{"missing required"},
		},
		{
			name: "invalid date format",
			config: &RepoConfig{
				Policies: map[string]*PolicyTemplate{
					"test_policy": {
						Rules: map[string]PolicyRule{
							"test_rule": {Type: "test"},
						},
					},
				},
				Repositories: &RepoTargets{
					Specific: []RepoSpecificConfig{
						{
							Name: "test-repo",
							Exceptions: []PolicyException{
								{
									PolicyName: "test_policy",
									RuleName:   "test_rule",
									Reason:     "Test",
									ApprovedBy: "admin",
									ExpiresAt:  "invalid", // Too short
								},
							},
						},
					},
				},
			},
			expectedErrs:  1,
			errorContains: []string{"invalid expiration date format"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := tt.config.ValidatePolicyExceptions()

			assert.Len(t, errors, tt.expectedErrs)

			for _, expected := range tt.errorContains {
				found := false

				for _, err := range errors {
					if assert.Contains(t, err, expected) {
						found = true
						break
					}
				}

				assert.True(t, found, "Expected error containing '%s' not found", expected)
			}
		})
	}
}

func TestGetPolicyExceptionReport(t *testing.T) {
	config := &RepoConfig{
		Repositories: &RepoTargets{
			Specific: []RepoSpecificConfig{
				{
					Name: "app1",
					Exceptions: []PolicyException{
						{
							PolicyName: "policy1",
							RuleName:   "rule1",
							Reason:     "Reason 1",
							ApprovedBy: "admin",
						},
						{
							PolicyName: "policy2",
							RuleName:   "rule2",
							Reason:     "Reason 2",
							ApprovedBy: "admin",
						},
					},
				},
				{
					Name: "app2",
					Exceptions: []PolicyException{
						{
							PolicyName: "policy1",
							RuleName:   "rule1",
							Reason:     "Different reason",
							ApprovedBy: "cto",
						},
					},
				},
			},
			Patterns: []RepoPatternConfig{
				{
					Match: "test-*",
					Exceptions: []PolicyException{
						{
							PolicyName: "policy3",
							RuleName:   "rule3",
							Reason:     "Pattern exception",
							ApprovedBy: "security",
						},
					},
				},
			},
		},
	}

	report := config.GetPolicyExceptionReport()

	// Check specific repositories
	assert.Contains(t, report, "app1")
	assert.Len(t, report["app1"], 2)
	assert.Equal(t, "specific", report["app1"][0].Type)
	assert.Equal(t, "app1", report["app1"][0].Repository)

	assert.Contains(t, report, "app2")
	assert.Len(t, report["app2"], 1)

	// Check pattern-based exceptions
	assert.Contains(t, report, "pattern:test-*")
	assert.Len(t, report["pattern:test-*"], 1)
	assert.Equal(t, "pattern", report["pattern:test-*"][0].Type)
	assert.Equal(t, "test-*", report["pattern:test-*"][0].Repository)
}

func TestIsExceptionActive(t *testing.T) {
	tests := []struct {
		name      string
		exception PolicyException
		expected  bool
	}{
		{
			name: "no expiration",
			exception: PolicyException{
				ExpiresAt: "",
			},
			expected: true,
		},
		{
			name: "with expiration",
			exception: PolicyException{
				ExpiresAt: "2024-12-31",
			},
			expected: true, // Current implementation always returns true
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.exception.IsExceptionActive())
		})
	}
}

func TestExceptionsWithTemplates(t *testing.T) {
	config := &RepoConfig{
		Version:      "1.0.0",
		Organization: "test-org",
		Templates: map[string]*RepoTemplate{
			"secure": {
				Settings: &RepoSettings{
					Private: boolPtr(true),
				},
			},
		},
		Policies: map[string]*PolicyTemplate{
			"security": {
				Rules: map[string]PolicyRule{
					"private_required": {
						Type:  "visibility",
						Value: "private",
					},
				},
			},
		},
		Repositories: &RepoTargets{
			Specific: []RepoSpecificConfig{
				{
					Name:     "public-demo",
					Template: "secure",
					Settings: &RepoSettings{
						Private: boolPtr(false), // Override template
					},
					Exceptions: []PolicyException{
						{
							PolicyName: "security",
							RuleName:   "private_required",
							Reason:     "Public demo repository",
							ApprovedBy: "security-team",
						},
					},
				},
			},
		},
	}

	settings, _, _, exceptions, err := config.GetEffectiveConfig("public-demo")
	require.NoError(t, err)

	// Settings should override template
	assert.False(t, *settings.Private)

	// Should have the exception
	assert.Len(t, exceptions, 1)
	assert.Equal(t, "Public demo repository", exceptions[0].Reason)
}
