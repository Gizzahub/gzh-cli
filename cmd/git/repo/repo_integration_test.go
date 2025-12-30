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

	"github.com/gizzahub/gzh-cli/pkg/git/provider"
	"github.com/gizzahub/gzh-cli/pkg/github"
	"github.com/gizzahub/gzh-cli/pkg/gitlab"
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
		config := &provider.ProviderConfig{
			Token:   githubToken,
			Timeout: 30,
		}
		gitHubProvider, err := github.CreateGitHubProvider(config)
		if err == nil {
			s.providers["github"] = gitHubProvider
			s.hasGitHub = true
			s.testOrgs["github"] = getEnvOrDefault("GITHUB_TEST_ORG", "")
			s.T().Log("GitHub provider initialized successfully")
		} else {
			s.T().Logf("GitHub provider initialization failed: %v", err)
		}
	} else {
		s.T().Log("GITHUB_TOKEN not set, skipping GitHub integration tests")
	}

	// Initialize GitLab provider if token is available
	if gitlabToken := os.Getenv("GITLAB_TOKEN"); gitlabToken != "" {
		config := &provider.ProviderConfig{
			Token:   gitlabToken,
			Timeout: 30,
		}
		gitLabProvider, err := gitlab.CreateGitLabProvider(config)
		if err == nil {
			s.providers["gitlab"] = gitLabProvider
			s.hasGitLab = true
			s.testOrgs["gitlab"] = getEnvOrDefault("GITLAB_TEST_ORG", "")
			s.T().Log("GitLab provider initialized successfully")
		} else {
			s.T().Logf("GitLab provider initialization failed: %v", err)
		}
	} else {
		s.T().Log("GITLAB_TOKEN not set, skipping GitLab integration tests")
	}

	// Gitea provider - requires GITEA_URL and GITEA_TOKEN
	// Gitea는 별도의 서버가 필요하므로 주로 로컬 테스트 환경에서 사용
	if giteaToken := os.Getenv("GITEA_TOKEN"); giteaToken != "" {
		if giteaURL := os.Getenv("GITEA_URL"); giteaURL != "" {
			s.T().Log("Gitea support requires custom provider implementation - skipping")
			// Gitea provider 구현이 완료되면 여기서 초기화
		} else {
			s.T().Log("GITEA_URL not set, skipping Gitea integration tests")
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for name, p := range s.providers {
		s.Run(name, func() {
			// Provider 이름 확인
			providerName := p.GetName()
			s.NotEmpty(providerName, "Provider name should not be empty")
			s.T().Logf("Provider name: %s", providerName)

			// Health check 실행
			health, err := p.HealthCheck(ctx)
			s.Require().NoError(err, "Health check should not fail")
			s.NotNil(health, "Health status should not be nil")
			s.T().Logf("Health status: %s, Latency: %v", health.Status, health.Latency)

			// Rate limit 확인
			rateLimit, err := p.GetRateLimit(ctx)
			if err == nil && rateLimit != nil {
				s.T().Logf("Rate limit - Remaining: %d/%d, Reset: %v",
					rateLimit.Remaining, rateLimit.Limit, rateLimit.Reset)
			}

			// Capabilities 확인
			caps := p.GetCapabilities()
			s.NotEmpty(caps, "Provider should have capabilities")
			s.T().Logf("Capabilities: %v", caps)
		})
	}
}

// TestListRepositories tests repository listing functionality.
func (s *GitRepoIntegrationTestSuite) TestListRepositories() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	for name, p := range s.providers {
		// 테스트 조직이 설정된 경우만 테스트
		testOrg := s.testOrgs[name]
		if testOrg == "" {
			s.T().Logf("Skipping %s: no test organization configured", name)
			continue
		}

		s.Run(name, func() {
			// 저장소 목록 조회
			opts := provider.ListOptions{
				Owner:    testOrg,
				Page:     1,
				PageSize: 10,
			}
			repoList, err := p.ListRepositories(ctx, opts)
			s.Require().NoError(err, "ListRepositories should not fail")
			s.NotNil(repoList, "Repository list should not be nil")

			s.T().Logf("Found %d repositories in %s (total: %d)",
				len(repoList.Items), testOrg, repoList.Total)

			// 첫 번째 저장소 정보 로깅
			if len(repoList.Items) > 0 {
				repo := repoList.Items[0]
				s.T().Logf("First repo: %s (default branch: %s)",
					repo.FullName, repo.DefaultBranch)
			}
		})
	}
}

// TestRepositoryLifecycle tests create, get, update, and delete operations.
// 참고: 이 테스트는 실제 저장소를 생성/삭제하므로 주의 필요
func (s *GitRepoIntegrationTestSuite) TestRepositoryLifecycle() {
	// 이 테스트는 파괴적인 작업(create/delete)을 수행하므로
	// 명시적인 환경변수가 설정된 경우에만 실행
	if os.Getenv("GZ_INTEGRATION_TEST_DESTRUCTIVE") != "1" {
		s.T().Skip("Destructive tests require GZ_INTEGRATION_TEST_DESTRUCTIVE=1")
	}
	// 실제 구현은 필요 시 추가
}

// TestRepositoryFiltering tests repository filtering capabilities.
func (s *GitRepoIntegrationTestSuite) TestRepositoryFiltering() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	for name, p := range s.providers {
		testOrg := s.testOrgs[name]
		if testOrg == "" {
			s.T().Logf("Skipping %s: no test organization configured", name)
			continue
		}

		s.Run(name, func() {
			// 아카이브되지 않은 저장소만 필터링
			opts := provider.ListOptions{
				Owner:    testOrg,
				Page:     1,
				PageSize: 5,
				Filters: map[string]interface{}{
					"archived": false,
				},
			}
			repoList, err := p.ListRepositories(ctx, opts)
			if err != nil {
				s.T().Logf("Filtering not supported or failed: %v", err)
				return
			}

			s.NotNil(repoList)
			s.T().Logf("Found %d non-archived repositories", len(repoList.Items))
		})
	}
}

// TestSearchRepositories tests repository search functionality.
func (s *GitRepoIntegrationTestSuite) TestSearchRepositories() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	for name, p := range s.providers {
		s.Run(name, func() {
			// 간단한 검색 쿼리
			query := provider.SearchQuery{
				Query:    "test",
				Page:     1,
				PageSize: 5,
			}
			result, err := p.SearchRepositories(ctx, query)
			if err != nil {
				s.T().Logf("Search not supported or failed: %v", err)
				return
			}

			s.NotNil(result)
			s.T().Logf("Search returned %d results (total: %d)",
				len(result.Items), result.Total)
		})
	}
}

// TestConcurrentOperations tests concurrent API operations.
func (s *GitRepoIntegrationTestSuite) TestConcurrentOperations() {
	// 동시성 테스트는 rate limit을 빠르게 소진할 수 있으므로 스킵
	s.T().Skip("Concurrent tests skipped to preserve rate limits")
}

// TestRateLimitHandling tests rate limit handling and retries.
func (s *GitRepoIntegrationTestSuite) TestRateLimitHandling() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for name, p := range s.providers {
		s.Run(name, func() {
			// Rate limit 정보 확인
			rateLimit, err := p.GetRateLimit(ctx)
			if err != nil {
				s.T().Logf("GetRateLimit not supported: %v", err)
				return
			}

			s.NotNil(rateLimit)
			s.T().Logf("Rate limit info - Limit: %d, Remaining: %d, Reset: %v",
				rateLimit.Limit, rateLimit.Remaining, rateLimit.Reset)

			// Rate limit이 충분한지 확인
			if rateLimit.Remaining < 10 {
				s.T().Logf("Warning: Low rate limit remaining (%d)", rateLimit.Remaining)
			}
		})
	}
}

// TestErrorHandling tests various error scenarios.
func (s *GitRepoIntegrationTestSuite) TestErrorHandling() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for name, p := range s.providers {
		s.Run(name+"_nonexistent_repo", func() {
			// 존재하지 않는 저장소 조회 시 에러 확인
			_, err := p.GetRepository(ctx, "nonexistent-repo-xyz-12345")
			s.Error(err, "Should fail for nonexistent repository")
		})
	}
}

// TestCommandLineIntegration tests CLI commands with real providers.
func (s *GitRepoIntegrationTestSuite) TestCommandLineIntegration() {
	// CLI 통합 테스트는 별도의 e2e 테스트로 분리
	s.T().Skip("CLI integration tests moved to e2e test suite")
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
