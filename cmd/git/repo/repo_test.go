// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package repo

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/Gizzahub/gzh-cli/internal/git/provider/mock"
	"github.com/Gizzahub/gzh-cli/pkg/git/provider"
)

// GitRepoTestSuite provides a comprehensive test suite for git repo commands.
type GitRepoTestSuite struct {
	suite.Suite
	mockProviders map[string]*mock.Provider
	testRepos     []provider.Repository
	tempDir       string
	ctx           context.Context
}

// SetupSuite initializes the test suite with mock providers and test data.
func (s *GitRepoTestSuite) SetupSuite() {
	s.ctx = context.Background()

	// Initialize mock providers
	s.mockProviders = map[string]*mock.Provider{
		"github": mock.NewProvider("github"),
		"gitlab": mock.NewProvider("gitlab"),
		"gitea":  mock.NewProvider("gitea"),
		"gogs":   mock.NewProvider("gogs"),
	}

	// Generate test repository data
	s.testRepos = s.generateTestRepos()
}

// SetupTest creates a temporary directory for each test.
func (s *GitRepoTestSuite) SetupTest() {
	var err error
	s.tempDir, err = os.MkdirTemp("", "gzh-git-repo-test-*")
	s.Require().NoError(err)

	// Change to temp directory for tests
	s.Require().NoError(os.Chdir(s.tempDir))
}

// TearDownTest cleans up the temporary directory after each test.
func (s *GitRepoTestSuite) TearDownTest() {
	if s.tempDir != "" {
		os.RemoveAll(s.tempDir)
	}
}

// skipIfNoProviderToken skips the test if the required provider token is not available.
// These tests require actual provider tokens because the current implementation
// does not support mock provider injection. Tests pass with mocked setup expectations
// but fail at runtime when getGitProvider() is called.
func (s *GitRepoTestSuite) skipIfNoProviderToken(providerTypes ...string) {
	tokenEnvVars := map[string][]string{
		"github": {"GITHUB_TOKEN", "GH_TOKEN"},
		"gitlab": {"GITLAB_TOKEN", "GL_TOKEN"},
		"gitea":  {"GITEA_TOKEN"},
		"gogs":   {"GOGS_TOKEN"},
	}

	for _, pt := range providerTypes {
		envVars, ok := tokenEnvVars[pt]
		if !ok {
			continue
		}

		hasToken := false
		for _, envVar := range envVars {
			if os.Getenv(envVar) != "" {
				hasToken = true
				break
			}
		}

		if !hasToken {
			s.T().Skipf("Skipping test: %s token not available (set %v)", pt, envVars)
		}
	}
}

// generateTestRepos creates a set of test repositories for various scenarios.
func (s *GitRepoTestSuite) generateTestRepos() []provider.Repository {
	now := time.Now()
	return []provider.Repository{
		{
			ID:            "1",
			Name:          "web-app",
			FullName:      "testorg/web-app",
			Description:   "Main web application",
			Private:       false,
			Language:      "TypeScript",
			CloneURL:      "https://github.com/testorg/web-app.git",
			SSHURL:        "git@github.com:testorg/web-app.git",
			HTMLURL:       "https://github.com/testorg/web-app",
			DefaultBranch: "main",
			Topics:        []string{"webapp", "react", "typescript"},
			Stars:         125,
			Forks:         23,
			CreatedAt:     now.AddDate(-1, -6, 0),
			UpdatedAt:     now.AddDate(0, 0, -2),
			Visibility:    provider.VisibilityPublic,
		},
		{
			ID:            "2",
			Name:          "api-service",
			FullName:      "testorg/api-service",
			Description:   "REST API service",
			Private:       true,
			Language:      "Go",
			CloneURL:      "https://github.com/testorg/api-service.git",
			SSHURL:        "git@github.com:testorg/api-service.git",
			HTMLURL:       "https://github.com/testorg/api-service",
			DefaultBranch: "main",
			Topics:        []string{"api", "golang", "microservice"},
			Stars:         45,
			Forks:         8,
			CreatedAt:     now.AddDate(-2, 0, 0),
			UpdatedAt:     now.AddDate(0, 0, -1),
			Visibility:    provider.VisibilityPrivate,
		},
		{
			ID:            "3",
			Name:          "api-gateway",
			FullName:      "testorg/api-gateway",
			Description:   "API Gateway service",
			Private:       false,
			Language:      "Go",
			CloneURL:      "https://github.com/testorg/api-gateway.git",
			SSHURL:        "git@github.com:testorg/api-gateway.git",
			HTMLURL:       "https://github.com/testorg/api-gateway",
			DefaultBranch: "main",
			Topics:        []string{"api", "gateway", "golang"},
			Stars:         89,
			Forks:         15,
			CreatedAt:     now.AddDate(-1, -3, 0),
			UpdatedAt:     now.AddDate(0, 0, -5),
			Visibility:    provider.VisibilityPublic,
		},
		{
			ID:            "4",
			Name:          "mobile-app",
			FullName:      "testorg/mobile-app",
			Description:   "Mobile application",
			Private:       true,
			Language:      "Swift",
			CloneURL:      "https://github.com/testorg/mobile-app.git",
			SSHURL:        "git@github.com:testorg/mobile-app.git",
			HTMLURL:       "https://github.com/testorg/mobile-app",
			DefaultBranch: "develop",
			Topics:        []string{"mobile", "ios", "swift"},
			Stars:         67,
			Forks:         12,
			CreatedAt:     now.AddDate(-1, 0, 0),
			UpdatedAt:     now.AddDate(0, 0, -3),
			Visibility:    provider.VisibilityPrivate,
		},
		{
			ID:            "5",
			Name:          "docs",
			FullName:      "testorg/docs",
			Description:   "Documentation site",
			Private:       false,
			Language:      "Markdown",
			CloneURL:      "https://github.com/testorg/docs.git",
			SSHURL:        "git@github.com:testorg/docs.git",
			HTMLURL:       "https://github.com/testorg/docs",
			DefaultBranch: "main",
			Topics:        []string{"documentation", "mdbook"},
			Stars:         34,
			Forks:         45,
			CreatedAt:     now.AddDate(-3, 0, 0),
			UpdatedAt:     now.AddDate(0, 0, -7),
			Visibility:    provider.VisibilityPublic,
		},
	}
}

// TestGitRepoSuite runs the complete test suite.
func TestGitRepoSuite(t *testing.T) {
	suite.Run(t, new(GitRepoTestSuite))
}

// Helper methods for tests

// assertDirectoryExists checks if a directory exists.
func (s *GitRepoTestSuite) assertDirectoryExists(path string) {
	fullPath := filepath.Join(s.tempDir, path)
	info, err := os.Stat(fullPath)
	s.Require().NoError(err, "Directory should exist: %s", path)
	s.Require().True(info.IsDir(), "Path should be a directory: %s", path)
}

// assertDirectoryNotExists checks if a directory does not exist.
func (s *GitRepoTestSuite) assertDirectoryNotExists(path string) {
	fullPath := filepath.Join(s.tempDir, path)
	_, err := os.Stat(fullPath)
	s.Require().True(os.IsNotExist(err), "Directory should not exist: %s", path)
}

// createTestFile creates a test file with content.
func (s *GitRepoTestSuite) createTestFile(path, content string) {
	fullPath := filepath.Join(s.tempDir, path)
	err := os.MkdirAll(filepath.Dir(fullPath), 0o755)
	s.Require().NoError(err)
	err = os.WriteFile(fullPath, []byte(content), 0o644)
	s.Require().NoError(err)
}

// resetMocks resets all mock providers to clean state.
func (s *GitRepoTestSuite) resetMocks() {
	for _, mockProvider := range s.mockProviders {
		mockProvider.Reset()
	}
}

// TestBasicRepoCommand tests basic repo command functionality.
func (s *GitRepoTestSuite) TestBasicRepoCommand() {
	cmd := NewGitRepoCmd()
	s.NotNil(cmd)
	s.Equal("repo", cmd.Use)
	s.NotEmpty(cmd.Short)
	// Long description은 optional

	// Check that all subcommands are registered
	// pull-all 명령은 bulk-update wrapper 통해 등록됨
	subcommands := []string{"clone", "list", "create", "delete", "archive", "sync", "migrate", "search"}
	for _, subcmd := range subcommands {
		found := false
		for _, child := range cmd.Commands() {
			// Use는 "command [args]" 형식일 수 있으므로 첫 단어만 비교
			cmdName := child.Use
			if idx := len(subcmd); idx <= len(cmdName) && cmdName[:idx] == subcmd {
				found = true
				break
			}
		}
		s.True(found, "Subcommand should be registered: %s", subcmd)
	}
}

// TestCommandValidation tests command validation.
func (s *GitRepoTestSuite) TestCommandValidation() {
	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			name:      "No subcommand",
			args:      []string{},
			expectErr: false, // Should show help
		},
		{
			name:      "Valid subcommand",
			args:      []string{"list", "--help"},
			expectErr: false,
		},
		{
			name:      "Invalid subcommand",
			args:      []string{"invalid-command"},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := NewGitRepoCmd()
			cmd.SetArgs(tc.args)

			err := cmd.Execute()
			if tc.expectErr {
				s.Error(err)
			} else {
				s.NoError(err)
			}
		})
	}
}

// Benchmark tests

// BenchmarkTestRepoGeneration benchmarks test repository generation.
func BenchmarkTestRepoGeneration(b *testing.B) {
	suite := &GitRepoTestSuite{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		repos := suite.generateTestRepos()
		if len(repos) == 0 {
			b.Fatal("No test repos generated")
		}
	}
}

// BenchmarkMockProviderSetup benchmarks mock provider setup.
func BenchmarkMockProviderSetup(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mockProviders := map[string]*mock.Provider{
			"github": mock.NewProvider("github"),
			"gitlab": mock.NewProvider("gitlab"),
			"gitea":  mock.NewProvider("gitea"),
		}
		if len(mockProviders) == 0 {
			b.Fatal("No mock providers created")
		}
	}
}
