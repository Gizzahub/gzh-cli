package repoconfig

import (
	"testing"

	"github.com/gizzahub/gzh-manager-go/pkg/config"
	"github.com/gizzahub/gzh-manager-go/pkg/github"
	"github.com/stretchr/testify/assert"
)

func TestCompareRepositoryConfigurations(t *testing.T) {
	tests := []struct {
		name              string
		repoName          string
		current           *github.RepositoryConfig
		targetSettings    *config.RepoSettings
		targetSecurity    *config.SecuritySettings
		targetPermissions *config.PermissionSettings
		templateName      string
		expectedDiffs     int
	}{
		{
			name:     "No differences",
			repoName: "test-repo",
			current: &github.RepositoryConfig{
				Description: "Test repository",
				Private:     false,
				Settings: github.RepoConfigSettings{
					HasIssues: true,
					HasWiki:   false,
				},
				BranchProtection: make(map[string]github.BranchProtectionConfig),
				Permissions: github.PermissionsConfig{
					Teams: make(map[string]string),
				},
			},
			targetSettings: &config.RepoSettings{
				Description: strPtr("Test repository"),
				Private:     boolPtr(false),
				HasIssues:   boolPtr(true),
				HasWiki:     boolPtr(false),
			},
			templateName:  "default",
			expectedDiffs: 0,
		},
		{
			name:     "Basic settings differences",
			repoName: "test-repo",
			current: &github.RepositoryConfig{
				Description: "Old description",
				Private:     false,
				Settings: github.RepoConfigSettings{
					HasIssues:           true,
					HasWiki:             true,
					DeleteBranchOnMerge: false,
				},
				BranchProtection: make(map[string]github.BranchProtectionConfig),
				Permissions: github.PermissionsConfig{
					Teams: make(map[string]string),
				},
			},
			targetSettings: &config.RepoSettings{
				Description:         strPtr("New description"),
				Private:             boolPtr(true),
				HasIssues:           boolPtr(false),
				HasWiki:             boolPtr(false),
				DeleteBranchOnMerge: boolPtr(true),
			},
			templateName:  "secure",
			expectedDiffs: 5, // description, visibility, issues, wiki, delete_branch_on_merge
		},
		{
			name:     "Branch protection differences",
			repoName: "test-repo",
			current: &github.RepositoryConfig{
				Settings: github.RepoConfigSettings{},
				BranchProtection: map[string]github.BranchProtectionConfig{
					"main": {
						RequiredReviews: 1,
						EnforceAdmins:   false,
					},
				},
				Permissions: github.PermissionsConfig{
					Teams: make(map[string]string),
				},
			},
			targetSecurity: &config.SecuritySettings{
				BranchProtection: map[string]*config.BranchProtectionRule{
					"main": {
						RequiredReviews: intPtr(2),
						EnforceAdmins:   boolPtr(true),
					},
				},
			},
			templateName:  "strict",
			expectedDiffs: 2, // required_reviews, enforce_admins
		},
		{
			name:     "New branch protection",
			repoName: "test-repo",
			current: &github.RepositoryConfig{
				Settings:         github.RepoConfigSettings{},
				BranchProtection: make(map[string]github.BranchProtectionConfig),
				Permissions: github.PermissionsConfig{
					Teams: make(map[string]string),
				},
			},
			targetSecurity: &config.SecuritySettings{
				BranchProtection: map[string]*config.BranchProtectionRule{
					"main": {
						RequiredReviews: intPtr(2),
					},
				},
			},
			templateName:  "protected",
			expectedDiffs: 1, // new branch protection
		},
		{
			name:     "Permission differences",
			repoName: "test-repo",
			current: &github.RepositoryConfig{
				Settings:         github.RepoConfigSettings{},
				BranchProtection: make(map[string]github.BranchProtectionConfig),
				Permissions: github.PermissionsConfig{
					Teams: map[string]string{
						"developers": "write",
					},
				},
			},
			targetPermissions: &config.PermissionSettings{
				TeamPermissions: map[string]string{
					"developers": "admin",
					"reviewers":  "read",
				},
			},
			templateName:  "team-based",
			expectedDiffs: 2, // update developers, add reviewers
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diffs := compareRepositoryConfigurations(
				tt.repoName,
				tt.current,
				tt.targetSettings,
				tt.targetSecurity,
				tt.targetPermissions,
				tt.templateName,
				nil, // no exceptions for this test
			)

			assert.Equal(t, tt.expectedDiffs, len(diffs), "Number of differences should match")

			// Verify all differences have required fields
			for _, diff := range diffs {
				assert.NotEmpty(t, diff.Repository, "Repository should not be empty")
				assert.NotEmpty(t, diff.Setting, "Setting should not be empty")
				assert.NotEmpty(t, diff.ChangeType, "ChangeType should not be empty")
				assert.NotEmpty(t, diff.Impact, "Impact should not be empty")
				assert.Equal(t, tt.templateName, diff.Template, "Template should match")
			}
		})
	}
}

func TestGetChangeType(t *testing.T) {
	tests := []struct {
		current  string
		target   string
		expected string
	}{
		{"", "value", "create"},
		{"value", "", "delete"},
		{"old", "new", "update"},
		{"", "", "update"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := getChangeType(tt.current, tt.target)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFindAppliedTemplate(t *testing.T) {
	repoConfig := &config.RepoConfig{
		Defaults: &config.RepoDefaults{
			Template: "base",
		},
		Repositories: &config.RepoTargets{
			Specific: []config.RepoSpecificConfig{
				{Name: "api-service", Template: "microservice"},
				{Name: "web-app", Template: "frontend"},
			},
			Patterns: []config.RepoPatternConfig{
				{Match: "*-service", Template: "backend"},
				{Match: "test-*", Template: "testing"},
			},
			Default: &config.RepoDefaultConfig{
				Template: "standard",
			},
		},
	}

	tests := []struct {
		repoName string
		expected string
	}{
		{"api-service", "microservice"}, // specific match
		{"web-app", "frontend"},         // specific match
		{"auth-service", "backend"},     // pattern match
		{"test-integration", "testing"}, // pattern match
		{"random-repo", "standard"},     // default match
		{"another-service", "backend"},  // pattern match
	}

	for _, tt := range tests {
		t.Run(tt.repoName, func(t *testing.T) {
			result := findAppliedTemplate(repoConfig, tt.repoName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMatchRepoPattern(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		expected bool
	}{
		{"api-service", "*-service", true},
		{"service-api", "*-service", false},
		{"test-repo", "test-*", true},
		{"repo-test", "test-*", false},
		{"exact-match", "exact-match", true},
		{"no-match", "exact-match", false},
		{"api-complex-service", "api-*-service", true},
	}

	for _, tt := range tests {
		t.Run(tt.name+"-"+tt.pattern, func(t *testing.T) {
			matched, err := matchRepoPattern(tt.name, tt.pattern)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, matched)
		})
	}
}

// Helper functions for creating pointers
func strPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

func intPtr(i int) *int {
	return &i
}
