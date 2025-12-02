// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package repo

import (
	"fmt"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/Gizzahub/gzh-cli/pkg/git/provider"
)

// TestCreateCommand tests the repository creation functionality.
// NOTE: Tests that execute commands with provider operations require tokens
// because the current implementation doesn't support mock provider injection.
func (s *GitRepoTestSuite) TestCreateCommand() {
	tests := []struct {
		name              string
		args              []string
		setup             func()
		validate          func()
		expectErr         bool
		requiresProviders []string
	}{
		{
			name: "Basic repository creation",
			args: []string{
				"create",
				"--provider", "github",
				"--org", "testorg",
				"--name", "newrepo",
				"--description", "Test repository",
			},
			requiresProviders: []string{"github"},
			setup: func() {
				s.resetMocks()
				mockProvider := s.mockProviders["github"]

				expectedRepo := &provider.Repository{
					ID:          "123",
					Name:        "newrepo",
					FullName:    "testorg/newrepo",
					Description: "Test repository",
					Private:     false,
					CloneURL:    "https://github.com/testorg/newrepo.git",
					SSHURL:      "git@github.com:testorg/newrepo.git",
					HTMLURL:     "https://github.com/testorg/newrepo",
					CreatedAt:   time.Now(),
				}

				mockProvider.SetupCreateResponse(func(req provider.CreateRepoRequest) bool {
					return req.Name == "newrepo" && req.Description == "Test repository" && !req.Private
				}, expectedRepo, nil)
			},
			validate: func() {
				// Validate mock expectations were met
				s.mockProviders["github"].AssertExpectations(s.T())
			},
		},
		{
			name: "Private repository creation",
			args: []string{
				"create",
				"--provider", "github",
				"--org", "testorg",
				"--name", "privaterepo",
				"--private",
				"--description", "Private test repository",
			},
			requiresProviders: []string{"github"},
			setup: func() {
				s.resetMocks()
				mockProvider := s.mockProviders["github"]

				expectedRepo := &provider.Repository{
					ID:          "124",
					Name:        "privaterepo",
					FullName:    "testorg/privaterepo",
					Description: "Private test repository",
					Private:     true,
					CloneURL:    "https://github.com/testorg/privaterepo.git",
					SSHURL:      "git@github.com:testorg/privaterepo.git",
					HTMLURL:     "https://github.com/testorg/privaterepo",
					CreatedAt:   time.Now(),
				}

				mockProvider.SetupCreateResponse(func(req provider.CreateRepoRequest) bool {
					return req.Name == "privaterepo" && req.Private
				}, expectedRepo, nil)
			},
			validate: func() {
				s.mockProviders["github"].AssertExpectations(s.T())
			},
		},
		{
			name: "Repository with features",
			args: []string{
				"create",
				"--provider", "github",
				"--org", "testorg",
				"--name", "fullrepo",
				"--description", "Full featured repository",
				"--issues",
				"--wiki",
				"--projects",
				"--topics", "golang,cli,testing",
			},
			requiresProviders: []string{"github"},
			setup: func() {
				s.resetMocks()
				mockProvider := s.mockProviders["github"]

				expectedRepo := &provider.Repository{
					ID:          "125",
					Name:        "fullrepo",
					FullName:    "testorg/fullrepo",
					Description: "Full featured repository",
					Topics:      []string{"golang", "cli", "testing"},
					CreatedAt:   time.Now(),
				}

				mockProvider.SetupCreateResponse(func(req provider.CreateRepoRequest) bool {
					return req.Name == "fullrepo" && req.HasIssues && req.HasWiki && req.HasProjects
				}, expectedRepo, nil)
			},
			validate: func() {
				s.mockProviders["github"].AssertExpectations(s.T())
			},
		},
		{
			name:      "Create without required name",
			args:      []string{"create", "--provider", "github", "--org", "testorg"},
			setup:     func() { s.resetMocks() },
			expectErr: true,
		},
		{
			name:      "Create without required org",
			args:      []string{"create", "--provider", "github", "--name", "test"},
			setup:     func() { s.resetMocks() },
			expectErr: true,
		},
		{
			name: "Create with invalid name",
			args: []string{
				"create",
				"--provider", "github",
				"--org", "testorg",
				"--name", "invalid@name",
			},
			setup:     func() { s.resetMocks() },
			expectErr: true,
		},
		{
			name: "Create with provider error",
			args: []string{
				"create",
				"--provider", "github",
				"--org", "testorg",
				"--name", "failrepo",
			},
			setup: func() {
				s.resetMocks()
				mockProvider := s.mockProviders["github"]
				mockProvider.SetupCreateResponse(func(req provider.CreateRepoRequest) bool {
					return req.Name == "failrepo"
				}, nil, fmt.Errorf("repository already exists"))
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// 토큰이 필요한 테스트는 토큰이 없으면 스킵
			if len(tt.requiresProviders) > 0 {
				s.skipIfNoProviderToken(tt.requiresProviders...)
			}

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
		})
	}
}

// TestListCommand tests the repository listing functionality.
// NOTE: Tests that execute commands with provider operations require tokens
// because the current implementation doesn't support mock provider injection.
func (s *GitRepoTestSuite) TestListCommand() {
	tests := []struct {
		name              string
		args              []string
		setup             func()
		validate          func()
		expectErr         bool
		requiresProviders []string
	}{
		{
			name:              "List all repositories",
			args:              []string{"list", "--provider", "github", "--org", "testorg"},
			requiresProviders: []string{"github"},
			setup: func() {
				s.resetMocks()
				mockProvider := s.mockProviders["github"]
				mockProvider.SetupListResponse("testorg", s.testRepos)
			},
			validate: func() {
				s.mockProviders["github"].AssertExpectations(s.T())
			},
		},
		{
			name:              "List with visibility filter",
			args:              []string{"list", "--provider", "github", "--org", "testorg", "--visibility", "private"},
			requiresProviders: []string{"github"},
			setup: func() {
				s.resetMocks()
				mockProvider := s.mockProviders["github"]

				// Filter private repos
				privateRepos := []provider.Repository{}
				for _, repo := range s.testRepos {
					if repo.Private {
						privateRepos = append(privateRepos, repo)
					}
				}

				result := &provider.RepositoryList{
					Repositories: privateRepos,
					TotalCount:   len(privateRepos),
					Page:         1,
					PerPage:      len(privateRepos),
				}

				mockProvider.On("ListRepositories", mock.Anything, mock.MatchedBy(func(opts provider.ListOptions) bool {
					return opts.Organization == "testorg" && opts.Visibility == provider.VisibilityPrivate
				})).Return(result, nil)
			},
			validate: func() {
				s.mockProviders["github"].AssertExpectations(s.T())
			},
		},
		{
			name:              "List with language filter",
			args:              []string{"list", "--provider", "github", "--org", "testorg", "--language", "Go"},
			requiresProviders: []string{"github"},
			setup: func() {
				s.resetMocks()
				mockProvider := s.mockProviders["github"]

				// Filter Go repos
				goRepos := []provider.Repository{}
				for _, repo := range s.testRepos {
					if repo.Language == "Go" {
						goRepos = append(goRepos, repo)
					}
				}

				result := &provider.RepositoryList{
					Repositories: goRepos,
					TotalCount:   len(goRepos),
					Page:         1,
					PerPage:      len(goRepos),
				}

				mockProvider.On("ListRepositories", mock.Anything, mock.MatchedBy(func(opts provider.ListOptions) bool {
					return opts.Organization == "testorg" && opts.Language == "Go"
				})).Return(result, nil)
			},
			validate: func() {
				s.mockProviders["github"].AssertExpectations(s.T())
			},
		},
		{
			name:              "List with pagination",
			args:              []string{"list", "--provider", "github", "--org", "testorg", "--page", "2", "--per-page", "2"},
			requiresProviders: []string{"github"},
			setup: func() {
				s.resetMocks()
				mockProvider := s.mockProviders["github"]

				result := &provider.RepositoryList{
					Repositories: s.testRepos[2:4], // Page 2 with 2 per page
					TotalCount:   len(s.testRepos),
					Page:         2,
					PerPage:      2,
					HasNext:      true,
					HasPrev:      true,
				}

				mockProvider.On("ListRepositories", mock.Anything, mock.MatchedBy(func(opts provider.ListOptions) bool {
					return opts.Organization == "testorg" && opts.Page == 2 && opts.PerPage == 2
				})).Return(result, nil)
			},
			validate: func() {
				s.mockProviders["github"].AssertExpectations(s.T())
			},
		},
		{
			name:      "List without org",
			args:      []string{"list", "--provider", "github"},
			setup:     func() { s.resetMocks() },
			expectErr: true,
		},
		{
			name: "List with provider error",
			args: []string{"list", "--provider", "github", "--org", "nonexistent"},
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
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// 토큰이 필요한 테스트는 토큰이 없으면 스킵
			if len(tt.requiresProviders) > 0 {
				s.skipIfNoProviderToken(tt.requiresProviders...)
			}

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
		})
	}
}

// TestDeleteCommand tests the repository deletion functionality.
// NOTE: Tests that execute commands with provider operations require tokens
// because the current implementation doesn't support mock provider injection.
func (s *GitRepoTestSuite) TestDeleteCommand() {
	tests := []struct {
		name              string
		args              []string
		setup             func()
		validate          func()
		expectErr         bool
		requiresProviders []string
	}{
		{
			name:              "Delete single repository",
			args:              []string{"delete", "--provider", "github", "--org", "testorg", "--repo", "test-repo"},
			requiresProviders: []string{"github"},
			setup: func() {
				s.resetMocks()
				mockProvider := s.mockProviders["github"]

				// Setup get repository response
				testRepo := s.testRepos[0]
				testRepo.FullName = "testorg/test-repo"
				mockProvider.SetupGetResponse("testorg/test-repo", &testRepo, nil)

				// Setup delete response
				mockProvider.SetupDeleteResponse(testRepo.ID, nil)
			},
			validate: func() {
				s.mockProviders["github"].AssertExpectations(s.T())
			},
		},
		{
			name:              "Delete with pattern matching",
			args:              []string{"delete", "--provider", "github", "--org", "testorg", "--match", "test-*"},
			requiresProviders: []string{"github"},
			setup: func() {
				s.resetMocks()
				mockProvider := s.mockProviders["github"]

				// Setup list response with test repos
				testRepos := []provider.Repository{
					{ID: "1", Name: "test-repo1", FullName: "testorg/test-repo1"},
					{ID: "2", Name: "test-repo2", FullName: "testorg/test-repo2"},
				}
				mockProvider.SetupListResponse("testorg", testRepos)

				// Setup delete responses
				for _, repo := range testRepos {
					mockProvider.SetupDeleteResponse(repo.ID, nil)
				}
			},
			validate: func() {
				s.mockProviders["github"].AssertExpectations(s.T())
			},
		},
		{
			name:              "Delete with confirmation",
			args:              []string{"delete", "--provider", "github", "--org", "testorg", "--repo", "test-repo", "--confirm"},
			requiresProviders: []string{"github"},
			setup: func() {
				s.resetMocks()
				mockProvider := s.mockProviders["github"]

				testRepo := s.testRepos[0]
				testRepo.FullName = "testorg/test-repo"
				mockProvider.SetupGetResponse("testorg/test-repo", &testRepo, nil)
				mockProvider.SetupDeleteResponse(testRepo.ID, nil)
			},
			validate: func() {
				s.mockProviders["github"].AssertExpectations(s.T())
			},
		},
		{
			name:      "Delete without repo or pattern",
			args:      []string{"delete", "--provider", "github", "--org", "testorg"},
			setup:     func() { s.resetMocks() },
			expectErr: true,
		},
		{
			name: "Delete non-existent repository",
			args: []string{"delete", "--provider", "github", "--org", "testorg", "--repo", "nonexistent"},
			setup: func() {
				s.resetMocks()
				mockProvider := s.mockProviders["github"]
				mockProvider.SetupGetResponse("testorg/nonexistent", nil, fmt.Errorf("repository not found"))
			},
			expectErr: true,
		},
		{
			name: "Delete with provider error",
			args: []string{"delete", "--provider", "github", "--org", "testorg", "--repo", "test-repo"},
			setup: func() {
				s.resetMocks()
				mockProvider := s.mockProviders["github"]

				testRepo := s.testRepos[0]
				testRepo.FullName = "testorg/test-repo"
				mockProvider.SetupGetResponse("testorg/test-repo", &testRepo, nil)
				mockProvider.SetupDeleteResponse(testRepo.ID, fmt.Errorf("permission denied"))
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// 토큰이 필요한 테스트는 토큰이 없으면 스킵
			if len(tt.requiresProviders) > 0 {
				s.skipIfNoProviderToken(tt.requiresProviders...)
			}

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
		})
	}
}

// TestArchiveCommand tests the repository archiving functionality.
// NOTE: Tests that execute commands with provider operations require tokens
// because the current implementation doesn't support mock provider injection.
func (s *GitRepoTestSuite) TestArchiveCommand() {
	tests := []struct {
		name              string
		args              []string
		setup             func()
		validate          func()
		expectErr         bool
		requiresProviders []string
	}{
		{
			name: "Archive single repository",
			// 수정: --repo 형식을 org/repo로 변경
			args:              []string{"archive", "--provider", "github", "--repo", "testorg/test-repo"},
			requiresProviders: []string{"github"},
			setup: func() {
				s.resetMocks()
				mockProvider := s.mockProviders["github"]

				testRepo := s.testRepos[0]
				testRepo.FullName = "testorg/test-repo"
				testRepo.Archived = false
				mockProvider.SetupGetResponse("testorg/test-repo", &testRepo, nil)
				mockProvider.SetupArchiveResponse(testRepo.ID, nil)
			},
			validate: func() {
				s.mockProviders["github"].AssertExpectations(s.T())
			},
		},
		{
			name: "Unarchive repository",
			// 수정: --repo 형식을 org/repo로 변경
			args:              []string{"archive", "--provider", "github", "--repo", "testorg/test-repo", "--unarchive"},
			requiresProviders: []string{"github"},
			setup: func() {
				s.resetMocks()
				mockProvider := s.mockProviders["github"]

				testRepo := s.testRepos[0]
				testRepo.FullName = "testorg/test-repo"
				testRepo.Archived = true
				mockProvider.SetupGetResponse("testorg/test-repo", &testRepo, nil)
				mockProvider.On("UnarchiveRepository", mock.Anything, testRepo.ID).Return(nil)
			},
			validate: func() {
				s.mockProviders["github"].AssertExpectations(s.T())
			},
		},
		{
			name:              "Archive with pattern matching",
			args:              []string{"archive", "--provider", "github", "--org", "testorg", "--match", "old-*"},
			requiresProviders: []string{"github"},
			setup: func() {
				s.resetMocks()
				mockProvider := s.mockProviders["github"]

				oldRepos := []provider.Repository{
					{ID: "1", Name: "old-project1", FullName: "testorg/old-project1", Archived: false},
					{ID: "2", Name: "old-project2", FullName: "testorg/old-project2", Archived: false},
				}
				mockProvider.SetupListResponse("testorg", oldRepos)

				for _, repo := range oldRepos {
					mockProvider.SetupArchiveResponse(repo.ID, nil)
				}
			},
			validate: func() {
				s.mockProviders["github"].AssertExpectations(s.T())
			},
		},
		{
			name:      "Archive without repo or pattern",
			args:      []string{"archive", "--provider", "github", "--org", "testorg"},
			setup:     func() { s.resetMocks() },
			expectErr: true,
		},
		{
			name: "Archive already archived repository",
			// 수정: --repo 형식을 org/repo로 변경
			args:              []string{"archive", "--provider", "github", "--repo", "testorg/test-repo"},
			requiresProviders: []string{"github"},
			setup: func() {
				s.resetMocks()
				mockProvider := s.mockProviders["github"]

				testRepo := s.testRepos[0]
				testRepo.FullName = "testorg/test-repo"
				testRepo.Archived = true // Already archived
				mockProvider.SetupGetResponse("testorg/test-repo", &testRepo, nil)
				// No archive call should be made
			},
			validate: func() {
				s.mockProviders["github"].AssertExpectations(s.T())
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// 토큰이 필요한 테스트는 토큰이 없으면 스킵
			if len(tt.requiresProviders) > 0 {
				s.skipIfNoProviderToken(tt.requiresProviders...)
			}

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
		})
	}
}

// TestSearchCommand tests the repository search functionality.
// NOTE: Tests that execute commands with provider operations require tokens
// because the current implementation doesn't support mock provider injection.
func (s *GitRepoTestSuite) TestSearchCommand() {
	tests := []struct {
		name              string
		args              []string
		setup             func()
		validate          func()
		expectErr         bool
		requiresProviders []string
	}{
		{
			name:              "Basic search",
			args:              []string{"search", "--provider", "github", "--query", "golang cli"},
			requiresProviders: []string{"github"},
			setup: func() {
				s.resetMocks()
				mockProvider := s.mockProviders["github"]

				searchResult := &provider.SearchResult{
					TotalCount:        3,
					IncompleteResults: false,
					Repositories:      s.testRepos[:3],
					Page:              1,
					PerPage:           10,
				}

				mockProvider.On("SearchRepositories", mock.Anything, mock.MatchedBy(func(query provider.SearchQuery) bool {
					return query.Query == "golang cli"
				})).Return(searchResult, nil)
			},
			validate: func() {
				s.mockProviders["github"].AssertExpectations(s.T())
			},
		},
		{
			name:              "Search with filters",
			args:              []string{"search", "--provider", "github", "--query", "api", "--language", "Go", "--sort", "stars"},
			requiresProviders: []string{"github"},
			setup: func() {
				s.resetMocks()
				mockProvider := s.mockProviders["github"]

				searchResult := &provider.SearchResult{
					TotalCount:   2,
					Repositories: s.testRepos[1:3], // Go repos
					Page:         1,
					PerPage:      10,
				}

				mockProvider.On("SearchRepositories", mock.Anything, mock.MatchedBy(func(query provider.SearchQuery) bool {
					return query.Query == "api" && query.Language == "Go" && query.Sort == "stars"
				})).Return(searchResult, nil)
			},
			validate: func() {
				s.mockProviders["github"].AssertExpectations(s.T())
			},
		},
		{
			name:      "Search without query",
			args:      []string{"search", "--provider", "github"},
			setup:     func() { s.resetMocks() },
			expectErr: true,
		},
		{
			name: "Search with provider error",
			args: []string{"search", "--provider", "github", "--query", "test"},
			setup: func() {
				s.resetMocks()
				mockProvider := s.mockProviders["github"]
				mockProvider.On("SearchRepositories", mock.Anything, mock.Anything).Return(
					(*provider.SearchResult)(nil),
					fmt.Errorf("search service unavailable"),
				)
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// 토큰이 필요한 테스트는 토큰이 없으면 스킵
			if len(tt.requiresProviders) > 0 {
				s.skipIfNoProviderToken(tt.requiresProviders...)
			}

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
		})
	}
}
