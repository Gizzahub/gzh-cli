// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package repo

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/stretchr/testify/mock"

	"github.com/gizzahub/gzh-cli/pkg/git/provider"
)

// TestCloneCommand tests the clone command functionality.
//
// Git invocations go through the suite's RecordingGitRunner (cloneGitRunner
// seam). Directories are created by the fake runner on successful clone, not by
// pre-seeding the filesystem.
func (s *GitRepoTestSuite) TestCloneCommand() {
	tests := []struct {
		name      string
		args      []string
		setup     func()
		validate  func()
		expectErr bool
	}{
		{
			name: "Basic clone from GitHub",
			args: []string{"clone", "--provider", "github", "--org", "testorg", "--target", "repos"},
			setup: func() {
				s.resetMocks()
				mockProvider := s.mockProviders["github"]
				mockProvider.SetupListResponse("testorg", s.testRepos[:3])
			},
			validate: func() {
				s.assertDirectoryExists("repos/testorg/web-app")
				s.assertDirectoryExists("repos/testorg/api-service")
				s.assertDirectoryExists("repos/testorg/api-gateway")

				clones := s.cloneRunner.CloneArgs()
				s.Require().Len(clones, 3)
				joined := cloneArgsJoined(clones)
				s.Contains(joined, "web-app")
				s.Contains(joined, "api-service")
				s.Contains(joined, "api-gateway")
			},
		},
		{
			name: "Clone with pattern matching",
			// Match uses regexp (not shell glob).
			args: []string{"clone", "--provider", "github", "--org", "testorg", "--match", "^api-"},
			setup: func() {
				s.resetMocks()
				mockProvider := s.mockProviders["github"]
				mockProvider.SetupListResponse("testorg", s.testRepos)
			},
			validate: func() {
				s.assertDirectoryExists("testorg/api-service")
				s.assertDirectoryExists("testorg/api-gateway")
				s.assertDirectoryNotExists("testorg/web-app")
				s.assertDirectoryNotExists("testorg/mobile-app")
				s.assertDirectoryNotExists("testorg/docs")

				clones := s.cloneRunner.CloneArgs()
				s.Require().Len(clones, 2)
			},
		},
		{
			name: "Clone with exclude pattern",
			args: []string{"clone", "--provider", "github", "--org", "testorg", "--exclude", "^mobile-"},
			setup: func() {
				s.resetMocks()
				mockProvider := s.mockProviders["github"]
				mockProvider.SetupListResponse("testorg", s.testRepos)
			},
			validate: func() {
				s.assertDirectoryExists("testorg/web-app")
				s.assertDirectoryExists("testorg/api-service")
				s.assertDirectoryExists("testorg/api-gateway")
				s.assertDirectoryExists("testorg/docs")
				s.assertDirectoryNotExists("testorg/mobile-app")
			},
		},
		{
			name: "Clone with visibility filter",
			args: []string{"clone", "--provider", "github", "--org", "testorg", "--visibility", "private"},
			setup: func() {
				s.resetMocks()
				mockProvider := s.mockProviders["github"]

				privateRepos := []provider.Repository{}
				for _, repo := range s.testRepos {
					if repo.Private {
						privateRepos = append(privateRepos, repo)
					}
				}

				mockProvider.SetupListResponse("testorg", privateRepos)
			},
			validate: func() {
				s.assertDirectoryExists("testorg/api-service")    // private
				s.assertDirectoryExists("testorg/mobile-app")     // private
				s.assertDirectoryNotExists("testorg/web-app")     // public
				s.assertDirectoryNotExists("testorg/api-gateway") // public
				s.assertDirectoryNotExists("testorg/docs")        // public
			},
		},
		{
			name: "Clone with language filter",
			args: []string{"clone", "--provider", "github", "--org", "testorg", "--language", "Go"},
			setup: func() {
				s.resetMocks()
				mockProvider := s.mockProviders["github"]

				goRepos := []provider.Repository{}
				for _, repo := range s.testRepos {
					if repo.Language == "Go" {
						goRepos = append(goRepos, repo)
					}
				}

				mockProvider.SetupListResponse("testorg", goRepos)
			},
			validate: func() {
				s.assertDirectoryExists("testorg/api-service")
				s.assertDirectoryExists("testorg/api-gateway")
				s.assertDirectoryNotExists("testorg/web-app")    // TypeScript
				s.assertDirectoryNotExists("testorg/mobile-app") // Swift
				s.assertDirectoryNotExists("testorg/docs")       // Markdown
			},
		},
		{
			name: "Clone single repository",
			// No --repo flag on the clone command; filter with exact-name match.
			args: []string{"clone", "--provider", "github", "--org", "testorg", "--match", "^web-app$"},
			setup: func() {
				s.resetMocks()
				mockProvider := s.mockProviders["github"]
				mockProvider.SetupListResponse("testorg", s.testRepos)
			},
			validate: func() {
				s.assertDirectoryExists("testorg/web-app")
				s.assertDirectoryNotExists("testorg/api-service")

				clones := s.cloneRunner.CloneArgs()
				s.Require().Len(clones, 1)
				s.Contains(strings.Join(clones[0], " "), "web-app")
			},
		},
		{
			name: "Clone with parallel workers",
			args: []string{"clone", "--provider", "github", "--org", "testorg", "--parallel", "3"},
			setup: func() {
				s.resetMocks()
				mockProvider := s.mockProviders["github"]
				mockProvider.SetupListResponse("testorg", s.testRepos)
			},
			validate: func() {
				for _, repo := range s.testRepos {
					s.assertDirectoryExists(filepath.Join("testorg", repo.Name))
				}
				s.Require().Len(s.cloneRunner.CloneArgs(), len(s.testRepos))
			},
		},
		{
			name: "Clone with dry run",
			args: []string{"clone", "--provider", "github", "--org", "testorg", "--dry-run"},
			setup: func() {
				s.resetMocks()
				mockProvider := s.mockProviders["github"]
				mockProvider.SetupListResponse("testorg", s.testRepos)
			},
			validate: func() {
				// No directories and no git clone invocations in dry-run mode.
				for _, repo := range s.testRepos {
					s.assertDirectoryNotExists(filepath.Join("testorg", repo.Name))
				}
				s.Empty(s.cloneRunner.CloneArgs())
				s.Empty(s.cloneRunner.Calls)
			},
		},
		{
			name:      "Clone with invalid provider",
			args:      []string{"clone", "--provider", "invalid", "--org", "testorg"},
			setup:     func() { s.resetMocks() },
			expectErr: true,
		},
		{
			name:      "Clone without required org",
			args:      []string{"clone", "--provider", "github"},
			setup:     func() { s.resetMocks() },
			expectErr: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.cloneRunner.Reset()
			s.cleanTempDir()

			if tt.setup != nil {
				tt.setup()
			}

			cmd := NewGitRepoCmd()
			cmd.SetArgs(tt.args)
			err := cmd.Execute()

			if tt.expectErr {
				s.Error(err)
			} else {
				s.NoError(err)
				if tt.validate != nil {
					tt.validate()
				}
			}

			for _, mockProvider := range s.mockProviders {
				mockProvider.AssertExpectations(s.T())
			}
		})
	}
}

// TestCloneCommandErrorHandling tests error handling in clone operations.
func (s *GitRepoTestSuite) TestCloneCommandErrorHandling() {
	testCases := []struct {
		name      string
		args      []string
		setup     func()
		expectErr bool
	}{
		{
			name: "Provider authentication failure",
			args: []string{"clone", "--provider", "github", "--org", "testorg"},
			setup: func() {
				s.resetMocks()
				mockProvider := s.mockProviders["github"]
				// Only the failing expectation: a SetupListResponse("testorg", nil)
				// here would be registered first and win the match, handing the
				// executor an empty-but-successful list instead of the error.
				mockProvider.On("ListRepositories", mock.Anything, mock.Anything).Return(
					(*provider.RepositoryList)(nil),
					fmt.Errorf("authentication failed"),
				)
			},
			expectErr: true,
		},
		{
			name: "Organization not found",
			args: []string{"clone", "--provider", "github", "--org", "nonexistent"},
			setup: func() {
				s.resetMocks()
				mockProvider := s.mockProviders["github"]
				mockProvider.On("ListRepositories", mock.Anything, mock.Anything).Return(
					(*provider.RepositoryList)(nil),
					fmt.Errorf("organization not found"),
				)
			},
			expectErr: true,
		},
		{
			name: "Clone operation failure",
			args: []string{"clone", "--provider", "github", "--org", "testorg"},
			setup: func() {
				s.resetMocks()
				mockProvider := s.mockProviders["github"]
				mockProvider.SetupListResponse("testorg", s.testRepos[:1])
				// Fail via the git runner seam (CloneRepository is not used by the executor).
				s.cloneRunner.FailClone = true
			},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.cloneRunner.Reset()
			s.cleanTempDir()

			if tc.setup != nil {
				tc.setup()
			}

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

// cloneArgsJoined flattens recorded clone arg slices for substring asserts.
func cloneArgsJoined(clones [][]string) string {
	var parts []string
	for _, c := range clones {
		parts = append(parts, strings.Join(c, " "))
	}
	return strings.Join(parts, " | ")
}
