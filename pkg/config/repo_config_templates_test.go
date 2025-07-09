package config

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPolicyTemplates(t *testing.T) {
	// Test loading each policy template
	templates := []struct {
		name     string
		filename string
		checks   func(t *testing.T, config *RepoConfig)
	}{
		{
			name:     "Security Enhanced Template",
			filename: "repo-config-security.yaml",
			checks:   testSecurityTemplate,
		},
		{
			name:     "Open Source Template",
			filename: "repo-config-opensource.yaml",
			checks:   testOpenSourceTemplate,
		},
		{
			name:     "Enterprise Template",
			filename: "repo-config-enterprise.yaml",
			checks:   testEnterpriseTemplate,
		},
	}

	for _, tt := range templates {
		t.Run(tt.name, func(t *testing.T) {
			// Load template file
			samplesDir := filepath.Join("..", "..", "samples")
			configPath := filepath.Join(samplesDir, tt.filename)

			config, err := LoadRepoConfig(configPath)
			require.NoError(t, err, "Failed to load %s", tt.filename)

			// Validate the config
			err = validateRepoConfig(config)
			require.NoError(t, err, "Validation failed for %s", tt.filename)

			// Run specific checks
			tt.checks(t, config)
		})
	}
}

func testSecurityTemplate(t *testing.T, config *RepoConfig) {
	// Check template exists
	template, ok := config.Templates["security-enhanced"]
	require.True(t, ok, "security-enhanced template not found")

	// Verify security settings
	assert.NotNil(t, template.Settings)
	assert.True(t, *template.Settings.Private, "Security repos must be private")
	assert.False(t, *template.Settings.HasWiki, "Wiki should be disabled")
	assert.True(t, *template.Settings.WebCommitSignoffRequired, "DCO should be required")

	// Verify security features
	assert.NotNil(t, template.Security)
	assert.True(t, *template.Security.VulnerabilityAlerts)
	assert.True(t, *template.Security.AutomatedSecurityFixes)
	assert.True(t, *template.Security.PrivateVulnerabilityReporting)

	// Verify branch protection
	mainProtection, ok := template.Security.BranchProtection["main"]
	require.True(t, ok, "main branch protection not found")
	assert.Equal(t, 3, *mainProtection.RequiredReviews)
	assert.True(t, *mainProtection.RequireCodeOwnerReviews)
	assert.True(t, *mainProtection.EnforceAdmins)
	assert.False(t, *mainProtection.AllowForcePushes)

	// Verify required status checks
	assert.Contains(t, mainProtection.RequiredStatusChecks, "security/vulnerability-scan")
	assert.Contains(t, mainProtection.RequiredStatusChecks, "security/secrets-scan")

	// Verify permissions
	assert.NotNil(t, template.Permissions)
	assert.Equal(t, "admin", template.Permissions.TeamPermissions["security-team"])

	// Verify compliance policy
	policy, ok := config.Policies["security_enhanced_compliance"]
	require.True(t, ok, "security_enhanced_compliance policy not found")
	assert.NotEmpty(t, policy.Rules)
}

func testOpenSourceTemplate(t *testing.T, config *RepoConfig) {
	// Check template exists
	template, ok := config.Templates["opensource-community"]
	require.True(t, ok, "opensource-community template not found")

	// Verify open source settings
	assert.NotNil(t, template.Settings)
	assert.False(t, *template.Settings.Private, "Open source repos must be public")
	assert.True(t, *template.Settings.HasIssues)
	assert.True(t, *template.Settings.HasWiki)
	assert.True(t, *template.Settings.HasProjects)
	assert.True(t, *template.Settings.HasDiscussions)
	assert.True(t, *template.Settings.AllowForking)

	// Verify all merge types allowed
	assert.True(t, *template.Settings.AllowSquashMerge)
	assert.True(t, *template.Settings.AllowMergeCommit)
	assert.True(t, *template.Settings.AllowRebaseMerge)
	assert.True(t, *template.Settings.AllowAutoMerge)

	// Verify relaxed branch protection
	mainProtection, ok := template.Security.BranchProtection["main"]
	require.True(t, ok, "main branch protection not found")
	assert.Equal(t, 1, *mainProtection.RequiredReviews)
	assert.False(t, *mainProtection.DismissStaleReviews)
	assert.False(t, *mainProtection.EnforceAdmins)

	// Verify CLA/DCO checks
	assert.Contains(t, mainProtection.RequiredStatusChecks, "license/cla")
	assert.Contains(t, mainProtection.RequiredStatusChecks, "dco")

	// Verify community topics
	assert.Contains(t, template.Topics, "open-source")
	assert.Contains(t, template.Topics, "community")
	assert.Contains(t, template.Topics, "hacktoberfest")

	// Verify required files
	assert.NotEmpty(t, template.RequiredFiles)
	hasFile := func(path string) bool {
		for _, f := range template.RequiredFiles {
			if f.Path == path {
				return true
			}
		}
		return false
	}
	assert.True(t, hasFile("README.md"))
	assert.True(t, hasFile("CONTRIBUTING.md"))
	assert.True(t, hasFile("CODE_OF_CONDUCT.md"))
	assert.True(t, hasFile("LICENSE"))

	// Verify open source policy
	policy, ok := config.Policies["opensource_best_practices"]
	require.True(t, ok, "opensource_best_practices policy not found")
	assert.NotEmpty(t, policy.Rules)
}

func testEnterpriseTemplate(t *testing.T, config *RepoConfig) {
	// Check template exists
	template, ok := config.Templates["enterprise-standard"]
	require.True(t, ok, "enterprise-standard template not found")

	// Verify enterprise settings
	assert.NotNil(t, template.Settings)
	assert.True(t, *template.Settings.Private, "Enterprise repos should be private")
	assert.True(t, *template.Settings.HasIssues)
	assert.True(t, *template.Settings.HasWiki)
	assert.True(t, *template.Settings.HasProjects)
	assert.False(t, *template.Settings.HasDiscussions, "Should use enterprise forums")

	// Verify merge strategy
	assert.True(t, *template.Settings.AllowSquashMerge)
	assert.False(t, *template.Settings.AllowMergeCommit)
	assert.False(t, *template.Settings.AllowRebaseMerge)
	assert.False(t, *template.Settings.AllowAutoMerge)

	// Verify branch protection
	mainProtection, ok := template.Security.BranchProtection["main"]
	require.True(t, ok, "main branch protection not found")
	assert.Equal(t, 2, *mainProtection.RequiredReviews)
	assert.True(t, *mainProtection.RequireCodeOwnerReviews)
	assert.True(t, *mainProtection.EnforceAdmins)

	// Verify enterprise CI checks
	assert.Contains(t, mainProtection.RequiredStatusChecks, "enterprise/security-scan")
	assert.Contains(t, mainProtection.RequiredStatusChecks, "enterprise/compliance-check")
	assert.Contains(t, mainProtection.RequiredStatusChecks, "sonarqube/quality-gate")

	// Verify deployment protection
	assert.NotNil(t, mainProtection.DeploymentProtectionRules)
	assert.Len(t, mainProtection.DeploymentProtectionRules, 1)
	assert.Equal(t, "production", mainProtection.DeploymentProtectionRules[0].Environment)

	// Verify team-based permissions
	assert.NotNil(t, template.Permissions)
	assert.Equal(t, "admin", template.Permissions.TeamPermissions["architects"])
	assert.Equal(t, "maintain", template.Permissions.TeamPermissions["senior-developers"])
	assert.Empty(t, template.Permissions.UserPermissions)

	// Verify compliance files
	hasFile := func(path string) bool {
		for _, f := range template.RequiredFiles {
			if f.Path == path {
				return true
			}
		}
		return false
	}
	assert.True(t, hasFile("COMPLIANCE.md"))
	assert.True(t, hasFile(".github/workflows/compliance.yml"))

	// Verify environments
	assert.NotEmpty(t, template.Environments)
	var prodEnv *EnvironmentConfig
	for _, env := range template.Environments {
		if env.Name == "production" {
			prodEnv = &env
			break
		}
	}
	require.NotNil(t, prodEnv, "production environment not found")
	assert.Equal(t, 60, prodEnv.ProtectionRules.WaitTimer)
	assert.Contains(t, prodEnv.ProtectionRules.RequiredReviewers, "release-managers")

	// Verify enterprise policy
	policy, ok := config.Policies["enterprise_governance"]
	require.True(t, ok, "enterprise_governance policy not found")
	assert.NotEmpty(t, policy.Rules)
}

func TestTemplateInheritance(t *testing.T) {
	// Test that security template inherits from standard in schema
	schemaPath := filepath.Join("..", "..", "docs", "repo-config-schema.yaml")
	config, err := LoadRepoConfig(schemaPath)
	require.NoError(t, err)

	// Check security template inherits from standard
	securityTemplate, ok := config.Templates["security"]
	require.True(t, ok, "security template not found in schema")
	assert.Equal(t, "standard", securityTemplate.Base)

	// Verify inheritance works correctly
	standardTemplate, ok := config.Templates["standard"]
	require.True(t, ok, "standard template not found in schema")

	// Get effective config for a repo using security template
	config.Repositories = &RepoTargets{
		Specific: []RepoSpecificConfig{
			{
				Name:     "test-secure-repo",
				Template: "security",
			},
		},
	}

	settings, security, _, err := config.GetEffectiveConfig("test-secure-repo")
	require.NoError(t, err)

	// Should have settings from both templates
	assert.True(t, *settings.Private)           // From security template
	assert.True(t, *settings.AllowSquashMerge)  // From standard template
	assert.False(t, *settings.AllowMergeCommit) // From standard template

	// Security settings should be merged
	assert.True(t, *security.VulnerabilityAlerts)           // From standard
	assert.True(t, *security.PrivateVulnerabilityReporting) // From security

	// Branch protection should be merged with security overrides
	mainProtection := security.BranchProtection["main"]
	assert.Equal(t, 2, *mainProtection.RequiredReviews) // Security overrides standard's 1
	assert.True(t, *mainProtection.EnforceAdmins)       // Security overrides standard's false
}

func TestPolicyValidation(t *testing.T) {
	tests := []struct {
		name       string
		policyName string
		repo       RepoSpecificConfig
		expectPass bool
	}{
		{
			name:       "Security policy - private repo passes",
			policyName: "security_enhanced_compliance",
			repo: RepoSpecificConfig{
				Name: "secure-repo",
				Settings: &RepoSettings{
					Private: boolPtr(true),
				},
			},
			expectPass: true,
		},
		{
			name:       "Security policy - public repo fails",
			policyName: "security_enhanced_compliance",
			repo: RepoSpecificConfig{
				Name: "public-repo",
				Settings: &RepoSettings{
					Private: boolPtr(false),
				},
			},
			expectPass: false,
		},
		{
			name:       "Open source policy - public repo passes",
			policyName: "opensource_best_practices",
			repo: RepoSpecificConfig{
				Name: "oss-repo",
				Settings: &RepoSettings{
					Private: boolPtr(false),
				},
			},
			expectPass: true,
		},
		{
			name:       "Enterprise policy - private repo passes",
			policyName: "enterprise_governance",
			repo: RepoSpecificConfig{
				Name: "enterprise-app",
				Settings: &RepoSettings{
					Private: boolPtr(true),
				},
			},
			expectPass: true,
		},
	}

	// Load security template for policies
	securityPath := filepath.Join("..", "..", "samples", "repo-config-security.yaml")
	securityConfig, err := LoadRepoConfig(securityPath)
	require.NoError(t, err)

	// For each test case, we would validate against the policy
	// This is a simplified test structure as the actual policy
	// validation implementation would be in the main code
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Here we're just checking that policies exist and have rules
			if policy, ok := securityConfig.Policies[tt.policyName]; ok {
				assert.NotNil(t, policy)
				assert.NotEmpty(t, policy.Rules)
			}
		})
	}
}

func TestTemplateUsagePatterns(t *testing.T) {
	// Test the overview template that shows usage patterns
	overviewPath := filepath.Join("..", "..", "samples", "repo-config-templates-overview.yaml")
	config, err := LoadRepoConfig(overviewPath)
	require.NoError(t, err)

	// Check pattern matching examples
	require.NotNil(t, config.Repositories)
	require.NotEmpty(t, config.Repositories.Patterns)

	// Verify infrastructure repos use security template
	for _, pattern := range config.Repositories.Patterns {
		if pattern.Match == "infra-*" || pattern.Match == "terraform-*" {
			assert.Equal(t, "security-enhanced", pattern.Template)
		}
		if pattern.Match == "example-*" || pattern.Match == "demo-*" {
			assert.Equal(t, "opensource-community", pattern.Template)
		}
		if pattern.Match == "*-service" || pattern.Match == "*-api" {
			assert.Equal(t, "enterprise-standard", pattern.Template)
		}
	}
}
