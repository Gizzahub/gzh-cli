// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

//go:build integration
// +build integration

package repo

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/Gizzahub/gzh-cli/internal/env"
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
		factory := github.NewGitHubProviderFactory(env.NewOSEnvironment())
		cloner, err := factory.CreateCloner(context.Background(), githubToken)
		if err == nil {
			// Create a mock provider for testing - this would need proper implementation
			// For now, skip GitHub tests as the provider interface doesn't match
			s.T().Log("GitHub provider available but interface mismatch - skipping")
			_ = cloner // Avoid unused variable
		} else {
			s.T().Logf("GitHub provider initialization failed: %v", err)
		}
	} else {
		s.T().Log("GITHUB_TOKEN not set, skipping GitHub integration tests")
	}

	// Initialize GitLab provider if token is available
	if gitlabToken := os.Getenv("GITLAB_TOKEN"); gitlabToken != "" {
		factory := gitlab.NewGitLabProviderFactory(env.NewOSEnvironment())
		cloner, err := factory.CreateCloner(context.Background(), gitlabToken)
		if err == nil {
			// Create a mock provider for testing - this would need proper implementation
			// For now, skip GitLab tests as the provider interface doesn't match
			s.T().Log("GitLab provider available but interface mismatch - skipping")
			_ = cloner // Avoid unused variable
		} else {
			s.T().Logf("GitLab provider initialization failed: %v", err)
		}
	} else {
		s.T().Log("GITLAB_TOKEN not set, skipping GitLab integration tests")
	}

	// Initialize Gitea provider if token is available
	if giteaToken := os.Getenv("GITEA_TOKEN"); giteaToken != "" {
		factory := gitea.NewGiteaProviderFactory(env.NewOSEnvironment())
		cloner, err := factory.CreateCloner(context.Background(), giteaToken)
		if err == nil {
			// Create a mock provider for testing - this would need proper implementation
			// For now, skip Gitea tests as the provider interface doesn't match
			s.T().Log("Gitea provider available but interface mismatch - skipping")
			_ = cloner // Avoid unused variable
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
	// TODO: 통합 테스트는 provider 인터페이스 리팩토링 후 재구현 필요
	s.T().Skip("Provider interface refactoring in progress")
}

// TestListRepositories tests repository listing functionality.
func (s *GitRepoIntegrationTestSuite) TestListRepositories() {
	// TODO: 통합 테스트는 provider 인터페이스 리팩토링 후 재구현 필요
	s.T().Skip("Provider interface refactoring in progress")
}

// TestRepositoryLifecycle tests create, get, update, and delete operations.
func (s *GitRepoIntegrationTestSuite) TestRepositoryLifecycle() {
	// TODO: 통합 테스트는 provider 인터페이스 리팩토링 후 재구현 필요
	s.T().Skip("Provider interface refactoring in progress")
}

// TestRepositoryFiltering tests repository filtering capabilities.
func (s *GitRepoIntegrationTestSuite) TestRepositoryFiltering() {
	// TODO: 통합 테스트는 provider 인터페이스 리팩토링 후 재구현 필요
	s.T().Skip("Provider interface refactoring in progress")
}

// TestSearchRepositories tests repository search functionality.
func (s *GitRepoIntegrationTestSuite) TestSearchRepositories() {
	// TODO: 통합 테스트는 provider 인터페이스 리팩토링 후 재구현 필요
	s.T().Skip("Provider interface refactoring in progress")
}

// TestConcurrentOperations tests concurrent API operations.
func (s *GitRepoIntegrationTestSuite) TestConcurrentOperations() {
	// TODO: 통합 테스트는 provider 인터페이스 리팩토링 후 재구현 필요
	s.T().Skip("Provider interface refactoring in progress")
}

// TestRateLimitHandling tests rate limit handling and retries.
func (s *GitRepoIntegrationTestSuite) TestRateLimitHandling() {
	// TODO: 통합 테스트는 provider 인터페이스 리팩토링 후 재구현 필요
	s.T().Skip("Provider interface refactoring in progress")
}

// TestErrorHandling tests various error scenarios.
func (s *GitRepoIntegrationTestSuite) TestErrorHandling() {
	// TODO: 통합 테스트는 provider 인터페이스 리팩토링 후 재구현 필요
	s.T().Skip("Provider interface refactoring in progress")
}

// TestCommandLineIntegration tests CLI commands with real providers.
func (s *GitRepoIntegrationTestSuite) TestCommandLineIntegration() {
	// TODO: 통합 테스트는 provider 인터페이스 리팩토링 후 재구현 필요
	s.T().Skip("Provider interface refactoring in progress")
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
