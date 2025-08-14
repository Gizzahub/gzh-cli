// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

//go:build integration
// +build integration

package git

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/Gizzahub/gzh-cli/pkg/git/provider"
	"github.com/Gizzahub/gzh-cli/pkg/gitea"
	"github.com/Gizzahub/gzh-cli/pkg/github"
	"github.com/Gizzahub/gzh-cli/pkg/gitlab"
)

// GitRepoIntegrationTestSuite provides integration tests for Git repo functionality
// using real provider APIs when authentication tokens are available.
type GitRepoIntegrationTestSuite struct {
	suite.Suite
	providers    map[string]provider.GitProvider
	testOrgs     map[string]string
	hasGitHub    bool
	hasGitLab    bool
	hasGitea     bool
	tempDir      string
	createdRepos []string // Track repos created during tests for cleanup
}

// SetupSuite initializes the integration test suite.
func (s *GitRepoIntegrationTestSuite) SetupSuite() {
	s.providers = make(map[string]provider.GitProvider)
	s.testOrgs = make(map[string]string)
	s.createdRepos = []string{}

	// Create temporary directory for test operations
	tempDir, err := os.MkdirTemp("", "gzh-integration-test-*")
	s.Require().NoError(err)
	s.tempDir = tempDir

	// Initialize GitHub provider if token is available
	if githubToken := os.Getenv("GITHUB_TOKEN"); githubToken != "" {
		githubProvider, err := github.NewProvider(github.Config{
			Token: githubToken,
		})
		if err == nil {
			s.providers["github"] = githubProvider
			s.testOrgs["github"] = getEnvOrDefault("GITHUB_TEST_ORG", "gizzahub")
			s.hasGitHub = true
			s.T().Logf("GitHub integration tests enabled with org: %s", s.testOrgs["github"])
		} else {
			s.T().Logf("GitHub provider initialization failed: %v", err)
		}
	} else {
		s.T().Log("GITHUB_TOKEN not set, skipping GitHub integration tests")
	}

	// Initialize GitLab provider if token is available
	if gitlabToken := os.Getenv("GITLAB_TOKEN"); gitlabToken != "" {
		gitlabProvider, err := gitlab.NewProvider(gitlab.Config{
			Token:   gitlabToken,
			BaseURL: getEnvOrDefault("GITLAB_BASE_URL", "https://gitlab.com"),
		})
		if err == nil {
			s.providers["gitlab"] = gitlabProvider
			s.testOrgs["gitlab"] = getEnvOrDefault("GITLAB_TEST_GROUP", "gizzahub")
			s.hasGitLab = true
			s.T().Logf("GitLab integration tests enabled with group: %s", s.testOrgs["gitlab"])
		} else {
			s.T().Logf("GitLab provider initialization failed: %v", err)
		}
	} else {
		s.T().Log("GITLAB_TOKEN not set, skipping GitLab integration tests")
	}

	// Initialize Gitea provider if token is available
	if giteaToken := os.Getenv("GITEA_TOKEN"); giteaToken != "" {
		giteaProvider, err := gitea.NewProvider(gitea.Config{
			Token:   giteaToken,
			BaseURL: getEnvOrDefault("GITEA_BASE_URL", "https://gitea.com"),
		})
		if err == nil {
			s.providers["gitea"] = giteaProvider
			s.testOrgs["gitea"] = getEnvOrDefault("GITEA_TEST_ORG", "gizzahub")
			s.hasGitea = true
			s.T().Logf("Gitea integration tests enabled with org: %s", s.testOrgs["gitea"])
		} else {
			s.T().Logf("Gitea provider initialization failed: %v", err)
		}
	} else {
		s.T().Log("GITEA_TOKEN not set, skipping Gitea integration tests")
	}

	if !s.hasGitHub && !s.hasGitLab && !s.hasGitea {
		s.T().Skip("No authentication tokens available, skipping all integration tests")
	}
}

// TearDownSuite cleans up after integration tests.
func (s *GitRepoIntegrationTestSuite) TearDownSuite() {
	// Clean up any repositories created during tests
	s.cleanupCreatedRepos()

	// Remove temporary directory
	if s.tempDir != "" {
		os.RemoveAll(s.tempDir)
	}
}

// TestProviderHealthCheck tests provider health and authentication.
func (s *GitRepoIntegrationTestSuite) TestProviderHealthCheck() {
	for providerName, provider := range s.providers {
		s.Run(fmt.Sprintf("HealthCheck_%s", providerName), func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			health, err := provider.HealthCheck(ctx)
			s.NoError(err, "Health check should succeed for %s", providerName)
			s.True(health.Healthy, "Provider %s should be healthy", providerName)
			s.NotEmpty(health.Version, "Provider %s should return version info", providerName)
		})
	}
}

// TestListRepositories tests repository listing functionality.
func (s *GitRepoIntegrationTestSuite) TestListRepositories() {
	for providerName, provider := range s.providers {
		org := s.testOrgs[providerName]
		s.Run(fmt.Sprintf("ListRepos_%s", providerName), func() {
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			opts := provider.ListOptions{
				Organization: org,
				PerPage:      10,
				Page:         1,
			}

			result, err := provider.ListRepositories(ctx, opts)
			s.NoError(err, "Listing repositories should succeed for %s", providerName)
			s.NotNil(result, "Result should not be nil for %s", providerName)
			s.GreaterOrEqual(result.TotalCount, 0, "Total count should be non-negative for %s", providerName)

			// Validate repository structure
			for _, repo := range result.Repositories {
				s.NotEmpty(repo.ID, "Repository ID should not be empty")
				s.NotEmpty(repo.Name, "Repository name should not be empty")
				s.NotEmpty(repo.FullName, "Repository full name should not be empty")
				s.Contains(repo.FullName, org, "Repository full name should contain organization")
			}
		})
	}
}

// TestRepositoryLifecycle tests create, get, update, and delete operations.
func (s *GitRepoIntegrationTestSuite) TestRepositoryLifecycle() {
	for providerName, provider := range s.providers {
		s.Run(fmt.Sprintf("Lifecycle_%s", providerName), func() {
			// Skip if running in CI without permission to create repos
			if os.Getenv("CI") == "true" && os.Getenv("ALLOW_REPO_CREATION") != "true" {
				s.T().Skip("Repository creation not allowed in CI environment")
			}

			ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
			defer cancel()

			repoName := fmt.Sprintf("gzh-test-repo-%d", time.Now().Unix())
			fullName := fmt.Sprintf("%s/%s", s.testOrgs[providerName], repoName)

			// Test repository creation
			createReq := provider.CreateRepoRequest{
				Name:        repoName,
				Description: "Integration test repository",
				Private:     true,
				HasIssues:   true,
				HasWiki:     false,
			}

			createdRepo, err := provider.CreateRepository(ctx, createReq)
			if err != nil {
				s.T().Skipf("Repository creation failed for %s (may lack permissions): %v", providerName, err)
				return
			}
			s.NotNil(createdRepo, "Created repository should not be nil")
			s.Equal(repoName, createdRepo.Name, "Created repository name should match")
			s.True(createdRepo.Private, "Created repository should be private")

			// Track for cleanup
			s.createdRepos = append(s.createdRepos, fmt.Sprintf("%s:%s", providerName, createdRepo.ID))

			// Test repository retrieval
			retrievedRepo, err := provider.GetRepository(ctx, fullName)
			s.NoError(err, "Getting repository should succeed")
			s.NotNil(retrievedRepo, "Retrieved repository should not be nil")
			s.Equal(createdRepo.ID, retrievedRepo.ID, "Retrieved repository ID should match")

			// Test repository update
			updateReq := provider.UpdateRepoRequest{
				Name:        repoName,
				Description: "Updated integration test repository",
				Private:     false, // Make it public
			}

			updatedRepo, err := provider.UpdateRepository(ctx, createdRepo.ID, updateReq)
			if err == nil {
				s.NotNil(updatedRepo, "Updated repository should not be nil")
				s.Equal("Updated integration test repository", updatedRepo.Description, "Description should be updated")
				s.False(updatedRepo.Private, "Repository should now be public")
			} else {
				s.T().Logf("Repository update failed for %s (may not be supported): %v", providerName, err)
			}

			// Test repository deletion
			err = provider.DeleteRepository(ctx, createdRepo.ID)
			s.NoError(err, "Deleting repository should succeed")

			// Remove from cleanup list since we successfully deleted it
			for i, repoRef := range s.createdRepos {
				if repoRef == fmt.Sprintf("%s:%s", providerName, createdRepo.ID) {
					s.createdRepos = append(s.createdRepos[:i], s.createdRepos[i+1:]...)
					break
				}
			}

			// Verify deletion
			_, err = provider.GetRepository(ctx, fullName)
			s.Error(err, "Getting deleted repository should fail")
		})
	}
}

// TestRepositoryFiltering tests repository filtering capabilities.
func (s *GitRepoIntegrationTestSuite) TestRepositoryFiltering() {
	for providerName, provider := range s.providers {
		org := s.testOrgs[providerName]
		s.Run(fmt.Sprintf("Filtering_%s", providerName), func() {
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			// Test visibility filtering
			publicOpts := provider.ListOptions{
				Organization: org,
				Visibility:   provider.VisibilityPublic,
				PerPage:      10,
			}

			publicResult, err := provider.ListRepositories(ctx, publicOpts)
			s.NoError(err, "Listing public repositories should succeed")
			s.NotNil(publicResult, "Public result should not be nil")

			// Verify all returned repos are public
			for _, repo := range publicResult.Repositories {
				s.False(repo.Private, "All repos should be public when filtering by public visibility")
			}

			// Test language filtering if Go repos exist
			goOpts := provider.ListOptions{
				Organization: org,
				Language:     "Go",
				PerPage:      10,
			}

			goResult, err := provider.ListRepositories(ctx, goOpts)
			s.NoError(err, "Listing Go repositories should succeed")
			s.NotNil(goResult, "Go result should not be nil")

			// Verify all returned repos are Go (if any)
			for _, repo := range goResult.Repositories {
				if repo.Language != "" {
					s.Equal("Go", repo.Language, "All repos should be Go when filtering by Go language")
				}
			}
		})
	}
}

// TestSearchRepositories tests repository search functionality.
func (s *GitRepoIntegrationTestSuite) TestSearchRepositories() {
	for providerName, provider := range s.providers {
		s.Run(fmt.Sprintf("Search_%s", providerName), func() {
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			searchQuery := provider.SearchQuery{
				Query:        "api",
				Sort:         "updated",
				Order:        "desc",
				PerPage:      10,
				Page:         1,
				Organization: s.testOrgs[providerName],
			}

			result, err := provider.SearchRepositories(ctx, searchQuery)
			if err != nil {
				// Some providers might not support search or require different permissions
				s.T().Logf("Search not supported or failed for %s: %v", providerName, err)
				return
			}

			s.NotNil(result, "Search result should not be nil")
			s.GreaterOrEqual(result.TotalCount, 0, "Total count should be non-negative")

			// Validate search results
			for _, repo := range result.Repositories {
				s.NotEmpty(repo.ID, "Repository ID should not be empty")
				s.NotEmpty(repo.Name, "Repository name should not be empty")
				// The search term "api" should appear somewhere in the repo data
			}
		})
	}
}

// TestConcurrentOperations tests concurrent API operations.
func (s *GitRepoIntegrationTestSuite) TestConcurrentOperations() {
	for providerName, provider := range s.providers {
		org := s.testOrgs[providerName]
		s.Run(fmt.Sprintf("Concurrent_%s", providerName), func() {
			ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
			defer cancel()

			// Perform multiple concurrent list operations
			const numConcurrent = 5
			results := make(chan error, numConcurrent)

			for i := 0; i < numConcurrent; i++ {
				go func(index int) {
					opts := provider.ListOptions{
						Organization: org,
						PerPage:      5,
						Page:         1,
					}

					_, err := provider.ListRepositories(ctx, opts)
					results <- err
				}(i)
			}

			// Collect results
			for i := 0; i < numConcurrent; i++ {
				err := <-results
				s.NoError(err, "Concurrent operation %d should succeed", i+1)
			}
		})
	}
}

// TestRateLimitHandling tests rate limit handling and retries.
func (s *GitRepoIntegrationTestSuite) TestRateLimitHandling() {
	for providerName, provider := range s.providers {
		org := s.testOrgs[providerName]
		s.Run(fmt.Sprintf("RateLimit_%s", providerName), func() {
			// Skip this test in CI to avoid hitting rate limits
			if os.Getenv("CI") == "true" {
				s.T().Skip("Skipping rate limit test in CI environment")
			}

			ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
			defer cancel()

			// Perform rapid-fire requests to potentially trigger rate limiting
			const numRequests = 20
			for i := 0; i < numRequests; i++ {
				opts := provider.ListOptions{
					Organization: org,
					PerPage:      1,
					Page:         1,
				}

				_, err := provider.ListRepositories(ctx, opts)
				if err != nil {
					s.T().Logf("Request %d failed (possibly due to rate limiting): %v", i+1, err)
					// Don't fail the test - rate limiting is expected behavior
					break
				}

				// Small delay between requests
				time.Sleep(100 * time.Millisecond)
			}
		})
	}
}

// TestErrorHandling tests various error scenarios.
func (s *GitRepoIntegrationTestSuite) TestErrorHandling() {
	for providerName, provider := range s.providers {
		s.Run(fmt.Sprintf("ErrorHandling_%s", providerName), func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			// Test non-existent organization
			opts := provider.ListOptions{
				Organization: "this-org-should-not-exist-12345",
				PerPage:      10,
			}

			_, err := provider.ListRepositories(ctx, opts)
			s.Error(err, "Listing repos for non-existent org should fail")

			// Test non-existent repository
			_, err = provider.GetRepository(ctx, "this-org-should-not-exist-12345/this-repo-should-not-exist")
			s.Error(err, "Getting non-existent repo should fail")

			// Test invalid repository creation (if supported)
			createReq := provider.CreateRepoRequest{
				Name:        "", // Invalid empty name
				Description: "Test repo",
			}

			_, err = provider.CreateRepository(ctx, createReq)
			s.Error(err, "Creating repo with invalid name should fail")
		})
	}
}

// TestCommandLineIntegration tests CLI commands with real providers.
func (s *GitRepoIntegrationTestSuite) TestCommandLineIntegration() {
	for providerName := range s.providers {
		org := s.testOrgs[providerName]
		s.Run(fmt.Sprintf("CLI_%s", providerName), func() {
			// Test list command
			listArgs := []string{
				"list",
				"--provider", providerName,
				"--org", org,
				"--limit", "5",
			}

			cmd := NewGitRepoCmd()
			cmd.SetArgs(listArgs)
			err := cmd.Execute()
			s.NoError(err, "List command should succeed for %s", providerName)

			// Test dry-run clone
			cloneArgs := []string{
				"clone",
				"--provider", providerName,
				"--org", org,
				"--dry-run",
				"--limit", "3",
			}

			cmd = NewGitRepoCmd()
			cmd.SetArgs(cloneArgs)
			err = cmd.Execute()
			s.NoError(err, "Dry-run clone command should succeed for %s", providerName)
		})
	}
}

// Helper functions

// getEnvOrDefault returns environment variable value or default if not set.
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// cleanupCreatedRepos cleans up repositories created during tests.
func (s *GitRepoIntegrationTestSuite) cleanupCreatedRepos() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	for _, repoRef := range s.createdRepos {
		parts := splitFirst(repoRef, ":")
		if len(parts) != 2 {
			continue
		}
		providerName, repoID := parts[0], parts[1]

		if provider, exists := s.providers[providerName]; exists {
			err := provider.DeleteRepository(ctx, repoID)
			if err != nil {
				s.T().Logf("Failed to cleanup repo %s from %s: %v", repoID, providerName, err)
			} else {
				s.T().Logf("Successfully cleaned up repo %s from %s", repoID, providerName)
			}
		}
	}
}

// splitFirst splits string at first occurrence of separator.
func splitFirst(s, sep string) []string {
	idx := len(s)
	for i := 0; i < len(s)-len(sep)+1; i++ {
		if s[i:i+len(sep)] == sep {
			idx = i
			break
		}
	}
	if idx == len(s) {
		return []string{s}
	}
	return []string{s[:idx], s[idx+len(sep):]}
}

// TestIntegrationSuite runs the integration test suite.
func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(GitRepoIntegrationTestSuite))
}
