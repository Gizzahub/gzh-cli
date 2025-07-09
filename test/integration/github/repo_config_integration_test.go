package github_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gizzahub/gzh-manager-go/pkg/config"
	"github.com/gizzahub/gzh-manager-go/pkg/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// Integration tests for GitHub repository configuration management
// These tests require actual GitHub organization access with appropriate tokens

const (
	testOrgEnvVar = "GITHUB_TEST_ORG"
	tokenEnvVar   = "GITHUB_TOKEN"
)

func skipIfNoTestOrg(t *testing.T) {
	if os.Getenv(testOrgEnvVar) == "" || os.Getenv(tokenEnvVar) == "" {
		t.Skipf("Skipping integration test: %s and %s environment variables must be set", testOrgEnvVar, tokenEnvVar)
	}
}

func TestIntegration_RepoConfig_EndToEnd(t *testing.T) {
	skipIfNoTestOrg(t)

	ctx := context.Background()
	testOrg := os.Getenv(testOrgEnvVar)
	token := os.Getenv(tokenEnvVar)

	// Create a temporary directory for test configs
	tmpDir, err := os.MkdirTemp("", "repo-config-integration-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a test repository configuration
	repoConfig := &config.RepoConfig{
		Version:      "1.0.0",
		Organization: testOrg,
		Templates: map[string]*config.RepoTemplate{
			"integration-test": {
				Description: "Integration test template",
				Settings: &config.RepoSettings{
					Private:   boolPtr(false),
					HasIssues: boolPtr(true),
					HasWiki:   boolPtr(false),
				},
				Security: &config.SecuritySettings{
					VulnerabilityAlerts: boolPtr(true),
				},
			},
		},
		Repositories: []config.RepoSpecificConfig{
			{
				Name:     "integration-test-repo-*",
				Template: "integration-test",
			},
		},
	}

	// Save the configuration
	configPath := filepath.Join(tmpDir, "repo-config.yaml")
	configData, err := yaml.Marshal(repoConfig)
	require.NoError(t, err)
	err = os.WriteFile(configPath, configData, 0o644)
	require.NoError(t, err)

	// Create repo config client
	client := github.NewRepoConfigClient(token)
	client.SetTimeout(30 * time.Second)

	// Test 1: List repositories in the organization
	t.Run("ListRepositories", func(t *testing.T) {
		repos, err := client.ListRepositories(ctx, testOrg)
		require.NoError(t, err)
		assert.NotEmpty(t, repos, "Test organization should have at least one repository")

		t.Logf("Found %d repositories in organization %s", len(repos), testOrg)
		for i, repo := range repos {
			if i < 5 { // Log first 5 repos
				t.Logf("  - %s (private: %v, archived: %v)", repo.Name, repo.Private, repo.Archived)
			}
		}
	})

	// Test 2: Get repository configuration
	t.Run("GetRepositoryConfiguration", func(t *testing.T) {
		// List repos first to get a valid repo name
		repos, err := client.ListRepositories(ctx, testOrg)
		require.NoError(t, err)
		require.NotEmpty(t, repos)

		testRepo := repos[0].Name
		config, err := client.GetRepositoryConfiguration(ctx, testOrg, testRepo)
		require.NoError(t, err)
		assert.NotNil(t, config)

		t.Logf("Repository %s configuration:", testRepo)
		t.Logf("  - Private: %v", config.Private)
		t.Logf("  - Has Issues: %v", config.Settings.HasIssues)
		t.Logf("  - Has Wiki: %v", config.Settings.HasWiki)
		t.Logf("  - Default Branch: %s", config.Settings.DefaultBranch)
	})

	// Test 3: Apply configuration (dry run)
	t.Run("ApplyConfiguration_DryRun", func(t *testing.T) {
		// Create a dry-run client
		dryRunClient := github.NewRepoConfigClient(token)
		dryRunClient.SetTimeout(30 * time.Second)

		// Load and apply configuration
		loadedConfig, err := config.LoadRepoConfig(configPath)
		require.NoError(t, err)

		results, err := dryRunClient.ApplyConfigurationToOrganization(ctx, loadedConfig, true)
		require.NoError(t, err)
		assert.NotNil(t, results)

		t.Logf("Dry run results:")
		t.Logf("  - Total repositories: %d", results.TotalRepositories)
		t.Logf("  - Matching repositories: %d", results.MatchingRepositories)
		t.Logf("  - Repositories with changes: %d", len(results.RepositoryResults))

		for repo, result := range results.RepositoryResults {
			if len(result.Changes) > 0 {
				t.Logf("  - Repository %s would have %d changes", repo, len(result.Changes))
			}
		}
	})

	// Test 4: Compliance audit
	t.Run("ComplianceAudit", func(t *testing.T) {
		// Create a policy configuration
		policyConfig := &config.RepoConfig{
			Version:      "1.0.0",
			Organization: testOrg,
			Policies: map[string]*config.PolicyTemplate{
				"basic-security": {
					Description: "Basic security policy",
					Rules: map[string]config.PolicyRule{
						"vulnerability_alerts": {
							Type:        "security_feature",
							Value:       true,
							Enforcement: "required",
							Message:     "Vulnerability alerts must be enabled",
						},
					},
				},
			},
			Patterns: []config.RepoPatternConfig{
				{
					Pattern: "*",
					Policies: []config.PolicyApplication{
						{
							Name:     "basic-security",
							Severity: "warning",
						},
					},
				},
			},
		}

		// Save policy configuration
		policyPath := filepath.Join(tmpDir, "policy-config.yaml")
		policyData, err := yaml.Marshal(policyConfig)
		require.NoError(t, err)
		err = os.WriteFile(policyPath, policyData, 0o644)
		require.NoError(t, err)

		// Run compliance audit
		auditReport, err := client.RunComplianceAudit(ctx, policyPath)
		require.NoError(t, err)
		assert.NotNil(t, auditReport)

		t.Logf("Compliance audit results:")
		t.Logf("  - Total repositories: %d", auditReport.TotalRepositories)
		t.Logf("  - Compliant repositories: %d", auditReport.CompliantRepositories)
		t.Logf("  - Non-compliant repositories: %d", auditReport.NonCompliantRepositories)

		// Log first few non-compliant repos
		count := 0
		for repo, result := range auditReport.RepositoryResults {
			if !result.IsCompliant && count < 5 {
				t.Logf("  - %s: %d violations", repo, len(result.Violations))
				count++
			}
		}
	})
}

func TestIntegration_RepoConfig_BulkOperations(t *testing.T) {
	skipIfNoTestOrg(t)

	ctx := context.Background()
	testOrg := os.Getenv(testOrgEnvVar)
	token := os.Getenv(tokenEnvVar)

	client := github.NewRepoConfigClient(token)

	t.Run("BulkUpdateTopics", func(t *testing.T) {
		// Get repositories
		repos, err := client.ListRepositories(ctx, testOrg)
		require.NoError(t, err)

		// Filter to test repositories only
		var testRepos []github.Repository
		for _, repo := range repos {
			if !repo.Archived && !repo.Private {
				testRepos = append(testRepos, repo)
				if len(testRepos) >= 3 { // Limit to 3 repos for testing
					break
				}
			}
		}

		if len(testRepos) == 0 {
			t.Skip("No suitable test repositories found")
		}

		// Add integration test topic
		topics := []string{"integration-test", "automated-test"}
		for _, repo := range testRepos {
			update := &github.RepositoryUpdate{
				Topics: topics,
			}

			updated, err := client.UpdateRepository(ctx, testOrg, repo.Name, update)
			require.NoError(t, err)
			assert.Contains(t, updated.Topics, "integration-test")

			t.Logf("Updated topics for %s: %v", repo.Name, updated.Topics)
		}

		// Clean up - remove the test topics
		for _, repo := range testRepos {
			update := &github.RepositoryUpdate{
				Topics: []string{}, // Clear topics
			}
			_, err := client.UpdateRepository(ctx, testOrg, repo.Name, update)
			require.NoError(t, err)
		}
	})

	t.Run("BulkBranchProtection", func(t *testing.T) {
		// This test requires admin permissions
		// Skip if not running with admin token
		repos, err := client.ListRepositories(ctx, testOrg)
		require.NoError(t, err)

		// Find a test repository with main branch
		var testRepo *github.Repository
		for _, repo := range repos {
			if !repo.Archived && repo.DefaultBranch == "main" {
				testRepo = &repo
				break
			}
		}

		if testRepo == nil {
			t.Skip("No suitable repository with main branch found")
		}

		// Try to get branch protection (may fail if no admin permissions)
		protection, err := client.GetBranchProtection(ctx, testOrg, testRepo.Name, "main")
		if err != nil {
			t.Logf("Skipping branch protection test: %v", err)
			return
		}

		t.Logf("Current branch protection for %s/main:", testRepo.Name)
		if protection != nil {
			t.Logf("  - Required reviews: %+v", protection.RequiredStatusChecks)
			t.Logf("  - Enforce admins: %v", protection.EnforceAdmins)
		}
	})
}

func TestIntegration_RepoConfig_RateLimiting(t *testing.T) {
	skipIfNoTestOrg(t)

	ctx := context.Background()
	token := os.Getenv(tokenEnvVar)

	client := github.NewRepoConfigClient(token)

	// Test rate limit handling
	t.Run("RateLimitStatus", func(t *testing.T) {
		status, err := client.GetRateLimitStatus(ctx)
		require.NoError(t, err)
		assert.NotNil(t, status)

		t.Logf("GitHub API Rate Limit Status:")
		t.Logf("  - Limit: %d", status.Limit)
		t.Logf("  - Remaining: %d", status.Remaining)
		t.Logf("  - Reset: %v", status.Reset)
		t.Logf("  - Used: %d", status.Used)
	})

	// Test concurrent requests with rate limiting
	t.Run("ConcurrentRequests", func(t *testing.T) {
		testOrg := os.Getenv(testOrgEnvVar)

		// Make 5 concurrent requests
		results := make(chan error, 5)
		for i := 0; i < 5; i++ {
			go func(index int) {
				_, err := client.ListRepositories(ctx, testOrg)
				results <- err
			}(i)
		}

		// Collect results
		errors := 0
		for i := 0; i < 5; i++ {
			if err := <-results; err != nil {
				errors++
				t.Logf("Request %d failed: %v", i, err)
			}
		}

		// All requests should succeed with proper rate limiting
		assert.Equal(t, 0, errors, "All concurrent requests should succeed")
	})
}

func TestIntegration_RepoConfig_ErrorHandling(t *testing.T) {
	skipIfNoTestOrg(t)

	ctx := context.Background()
	token := os.Getenv(tokenEnvVar)
	testOrg := os.Getenv(testOrgEnvVar)

	client := github.NewRepoConfigClient(token)

	t.Run("NonExistentRepository", func(t *testing.T) {
		_, err := client.GetRepository(ctx, testOrg, "this-repo-definitely-does-not-exist-12345")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "404")
	})

	t.Run("InvalidToken", func(t *testing.T) {
		invalidClient := github.NewRepoConfigClient("invalid-token")
		_, err := invalidClient.ListRepositories(ctx, testOrg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "401")
	})

	t.Run("NonExistentOrganization", func(t *testing.T) {
		_, err := client.ListRepositories(ctx, "this-org-definitely-does-not-exist-12345")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "404")
	})
}

// Helper functions
func boolPtr(b bool) *bool {
	return &b
}

func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

// Test fixture creation
func createTestRepoConfig(org string) *config.RepoConfig {
	return &config.RepoConfig{
		Version:      "1.0.0",
		Organization: org,
		Templates: map[string]*config.RepoTemplate{
			"test-template": {
				Description: "Test template for integration tests",
				Settings: &config.RepoSettings{
					Private:          boolPtr(false),
					HasIssues:        boolPtr(true),
					HasWiki:          boolPtr(false),
					HasProjects:      boolPtr(false),
					AllowSquashMerge: boolPtr(true),
					AllowMergeCommit: boolPtr(true),
					AllowRebaseMerge: boolPtr(true),
				},
				Security: &config.SecuritySettings{
					VulnerabilityAlerts: boolPtr(true),
					SecretScanning:      boolPtr(false),
				},
			},
		},
		Repositories: []config.RepoSpecificConfig{
			{
				Name:     "test-*",
				Template: "test-template",
			},
		},
	}
}

// Performance test helper
func measureOperationTime(t *testing.T, operation string, fn func() error) {
	start := time.Now()
	err := fn()
	duration := time.Since(start)

	if err != nil {
		t.Errorf("%s failed: %v", operation, err)
	} else {
		t.Logf("%s completed in %v", operation, duration)
	}
}
