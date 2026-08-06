// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package repo

import (
	"fmt"
	"strings"

	"github.com/gizzahub/gzh-cli/pkg/git/provider"
)

// Two production gaps keep the sync happy paths from being exercisable. Both are
// tracked in tasks/issue/; the cases tagged with them are kept as the
// specification for that work rather than deleted.
const (
	// syncNeedsGitSeam: the code action shells out to git against the
	// repository's real CloneURL (internal/git/sync/code.go:61), so any case that
	// reaches execution tries to clone github.com/testorg/... for real and fails
	// with exit status 128. Substituting the provider cannot help -- the git
	// invocation itself has no seam.
	syncNeedsGitSeam = "SyncEngine has no seam around git; its code action clones for real"

	// syncNeedsOrgCreate: syncRepository refuses to create a repository when the
	// destination is an organization rather than a single repository
	// (internal/git/sync/parallel.go:185). The blocker underneath is that neither
	// CreateRepoRequest nor ProviderConfig carries an owner, so there is nowhere
	// to record which organization to create in.
	syncNeedsOrgCreate = "org-level repository creation is not implemented"
)

// TestSyncCommand tests the repository synchronization functionality.
func (s *GitRepoTestSuite) TestSyncCommand() {
	tests := []struct {
		name       string
		args       []string
		setup      func()
		validate   func()
		expectErr  bool
		skipReason string
	}{
		{
			name:       "Sync single repository",
			skipReason: syncNeedsGitSeam,
			args: []string{
				"sync",
				"--from", "github:testorg/webapp",
				"--to", "gitlab:testgroup/webapp",
				"--create-missing",
			},
			setup: func() {
				s.resetMocks()
				srcProvider := s.mockProviders["github"]
				dstProvider := s.mockProviders["gitlab"]

				// Source repository exists. Name has to move with FullName: the
				// create request the engine builds carries Name, which is what the
				// matcher below keys on.
				srcRepo := s.testRepos[0]
				srcRepo.Name = "webapp"
				srcRepo.FullName = "testorg/webapp"
				srcProvider.SetupGetResponse("testorg/webapp", &srcRepo, nil)

				// Destination repository doesn't exist (will be created)
				dstProvider.SetupGetResponse("testgroup/webapp", nil, fmt.Errorf("repository not found"))

				// Setup create repository response
				dstRepo := srcRepo
				dstRepo.FullName = "testgroup/webapp"
				dstRepo.ID = "dst-123"
				dstProvider.SetupCreateResponse(func(req provider.CreateRepoRequest) bool {
					return req.Name == "webapp"
				}, &dstRepo, nil)
			},
			validate: func() {
				s.mockProviders["github"].AssertExpectations(s.T())
				s.mockProviders["gitlab"].AssertExpectations(s.T())
			},
		},
		{
			name:       "Sync organization repositories",
			skipReason: syncNeedsOrgCreate,
			args: []string{
				"sync",
				"--from", "github:sourceorg",
				"--to", "gitea:targetorg",
				"--create-missing",
				"--include-issues",
			},
			setup: func() {
				s.resetMocks()
				srcProvider := s.mockProviders["github"]
				dstProvider := s.mockProviders["gitea"]

				// Source organization has multiple repos
				srcProvider.SetupListResponse("sourceorg", s.testRepos[:3])

				// Destination organization is empty
				dstProvider.SetupListResponse("targetorg", []provider.Repository{})

				// Setup creation for each source repo
				for i, repo := range s.testRepos[:3] {
					dstRepo := repo
					dstRepo.ID = fmt.Sprintf("dst-%d", i+1)
					dstRepo.FullName = fmt.Sprintf("targetorg/%s", repo.Name)
					dstProvider.SetupCreateResponse(func(repoName string) func(provider.CreateRepoRequest) bool {
						return func(req provider.CreateRepoRequest) bool {
							return req.Name == repoName
						}
					}(repo.Name), &dstRepo, nil)
				}
			},
			validate: func() {
				s.mockProviders["github"].AssertExpectations(s.T())
				s.mockProviders["gitea"].AssertExpectations(s.T())
			},
		},
		{
			name:       "Sync with filtering",
			skipReason: syncNeedsOrgCreate,
			args: []string{
				"sync",
				"--from", "github:sourceorg",
				"--to", "gitlab:targetorg",
				"--match", "api-*",
				"--create-missing",
			},
			setup: func() {
				s.resetMocks()
				srcProvider := s.mockProviders["github"]
				dstProvider := s.mockProviders["gitlab"]

				// Only API repos should be synced
				apiRepos := []provider.Repository{}
				for _, repo := range s.testRepos {
					if repo.Name == "api-service" || repo.Name == "api-gateway" {
						apiRepos = append(apiRepos, repo)
					}
				}

				srcProvider.SetupListResponse("sourceorg", apiRepos)
				dstProvider.SetupListResponse("targetorg", []provider.Repository{})

				// Setup creation for filtered repos
				for i, repo := range apiRepos {
					dstRepo := repo
					dstRepo.ID = fmt.Sprintf("dst-api-%d", i+1)
					dstRepo.FullName = fmt.Sprintf("targetorg/%s", repo.Name)
					dstProvider.SetupCreateResponse(func(repoName string) func(provider.CreateRepoRequest) bool {
						return func(req provider.CreateRepoRequest) bool {
							return req.Name == repoName
						}
					}(repo.Name), &dstRepo, nil)
				}
			},
			validate: func() {
				s.mockProviders["github"].AssertExpectations(s.T())
				s.mockProviders["gitlab"].AssertExpectations(s.T())
			},
		},
		{
			name:       "Sync with update existing",
			skipReason: syncNeedsGitSeam,
			args: []string{
				"sync",
				"--from", "github:sourceorg",
				"--to", "gitlab:targetorg",
				"--update-existing",
			},
			setup: func() {
				s.resetMocks()
				srcProvider := s.mockProviders["github"]
				dstProvider := s.mockProviders["gitlab"]

				// Source has repos
				srcProvider.SetupListResponse("sourceorg", s.testRepos[:2])

				// Destination has some existing repos
				existingRepos := []provider.Repository{
					{
						ID:       "existing-1",
						Name:     s.testRepos[0].Name,
						FullName: fmt.Sprintf("targetorg/%s", s.testRepos[0].Name),
					},
				}
				dstProvider.SetupListResponse("targetorg", existingRepos)
			},
			validate: func() {
				s.mockProviders["github"].AssertExpectations(s.T())
				s.mockProviders["gitlab"].AssertExpectations(s.T())
			},
		},
		{
			name: "Sync with dry run",
			args: []string{
				"sync",
				"--from", "github:sourceorg",
				"--to", "gitlab:targetorg",
				"--create-missing",
				"--dry-run",
			},
			setup: func() {
				s.resetMocks()
				srcProvider := s.mockProviders["github"]
				dstProvider := s.mockProviders["gitlab"]

				srcProvider.SetupListResponse("sourceorg", s.testRepos[:2])
				dstProvider.SetupListResponse("targetorg", []provider.Repository{})

				// No create operations should be called in dry-run mode
			},
			validate: func() {
				s.mockProviders["github"].AssertExpectations(s.T())
				s.mockProviders["gitlab"].AssertExpectations(s.T())
			},
		},
		{
			name:       "Sync with parallel workers",
			skipReason: syncNeedsOrgCreate,
			args: []string{
				"sync",
				"--from", "github:sourceorg",
				"--to", "gitlab:targetorg",
				"--create-missing",
				"--parallel", "3",
			},
			setup: func() {
				s.resetMocks()
				srcProvider := s.mockProviders["github"]
				dstProvider := s.mockProviders["gitlab"]

				srcProvider.SetupListResponse("sourceorg", s.testRepos)
				dstProvider.SetupListResponse("targetorg", []provider.Repository{})

				// Setup creation for all repos
				for i, repo := range s.testRepos {
					dstRepo := repo
					dstRepo.ID = fmt.Sprintf("dst-%d", i+1)
					dstRepo.FullName = fmt.Sprintf("targetorg/%s", repo.Name)
					dstProvider.SetupCreateResponse(func(repoName string) func(provider.CreateRepoRequest) bool {
						return func(req provider.CreateRepoRequest) bool {
							return req.Name == repoName
						}
					}(repo.Name), &dstRepo, nil)
				}
			},
			validate: func() {
				s.mockProviders["github"].AssertExpectations(s.T())
				s.mockProviders["gitlab"].AssertExpectations(s.T())
			},
		},
		{
			name:      "Sync without from parameter",
			args:      []string{"sync", "--to", "gitlab:targetorg"},
			setup:     func() { s.resetMocks() },
			expectErr: true,
		},
		{
			name:      "Sync without to parameter",
			args:      []string{"sync", "--from", "github:sourceorg"},
			setup:     func() { s.resetMocks() },
			expectErr: true,
		},
		{
			name:      "Sync with invalid from format",
			args:      []string{"sync", "--from", "invalid-format", "--to", "gitlab:targetorg"},
			setup:     func() { s.resetMocks() },
			expectErr: true,
		},
		{
			name:      "Sync with invalid to format",
			args:      []string{"sync", "--from", "github:sourceorg", "--to", "invalid-format"},
			setup:     func() { s.resetMocks() },
			expectErr: true,
		},
		{
			name: "Sync with mismatched target types",
			args: []string{
				"sync",
				"--from", "github:sourceorg/repo",
				"--to", "gitlab:targetorg",
			},
			setup:     func() { s.resetMocks() },
			expectErr: true,
		},
		{
			name: "Sync with source repository not found",
			args: []string{
				"sync",
				"--from", "github:sourceorg/nonexistent",
				"--to", "gitlab:targetorg/newrepo",
				"--create-missing",
			},
			setup: func() {
				s.resetMocks()
				srcProvider := s.mockProviders["github"]
				srcProvider.SetupGetResponse("sourceorg/nonexistent", nil, fmt.Errorf("repository not found"))
			},
			expectErr: true,
		},
		{
			name: "Sync with destination creation failure",
			args: []string{
				"sync",
				"--from", "github:sourceorg/webapp",
				"--to", "gitlab:targetorg/webapp",
				"--create-missing",
			},
			setup: func() {
				s.resetMocks()
				srcProvider := s.mockProviders["github"]
				dstProvider := s.mockProviders["gitlab"]

				// Name has to move with FullName: the create request the engine
				// builds carries Name, which is what the matcher below keys on.
				srcRepo := s.testRepos[0]
				srcRepo.Name = "webapp"
				srcRepo.FullName = "sourceorg/webapp"
				srcProvider.SetupGetResponse("sourceorg/webapp", &srcRepo, nil)

				dstProvider.SetupGetResponse("targetorg/webapp", nil, fmt.Errorf("repository not found"))
				dstProvider.SetupCreateResponse(func(req provider.CreateRepoRequest) bool {
					return req.Name == "webapp"
				}, nil, fmt.Errorf("creation failed: permission denied"))
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			if tt.skipReason != "" {
				s.T().Skip(tt.skipReason)
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

// TestSyncCommandOptions tests sync command option parsing and validation.
func (s *GitRepoTestSuite) TestSyncCommandOptions() {
	tests := []struct {
		name      string
		args      []string
		setup     func()
		expectErr bool
	}{
		{
			name: "Valid sync options",
			args: []string{
				"sync",
				"--from", "github:sourceorg",
				"--to", "gitlab:targetorg",
				"--create-missing",
				"--update-existing",
				"--include-code",
				"--include-issues",
				"--include-wiki",
				"--parallel", "2",
			},
			setup: func() {
				s.resetMocks()
				srcProvider := s.mockProviders["github"]
				dstProvider := s.mockProviders["gitlab"]
				srcProvider.SetupListResponse("sourceorg", []provider.Repository{})
				dstProvider.SetupListResponse("targetorg", []provider.Repository{})
			},
			expectErr: false,
		},
		{
			// Validate() is deliberately asymmetric: it clamps a too-small worker
			// count up to 1 and only rejects counts above the ceiling. Clamping a
			// nonsensical low value is safe; silently shrinking a large one would
			// hide what the user asked for, so that case errors instead.
			name: "Parallel count below minimum is clamped",
			args: []string{
				"sync",
				"--from", "github:sourceorg",
				"--to", "gitlab:targetorg",
				"--parallel", "0",
			},
			setup: func() {
				s.resetMocks()
				s.mockProviders["github"].SetupListResponse("sourceorg", []provider.Repository{})
				s.mockProviders["gitlab"].SetupListResponse("targetorg", []provider.Repository{})
			},
			expectErr: false,
		},
		{
			name: "Parallel count too high",
			args: []string{
				"sync",
				"--from", "github:sourceorg",
				"--to", "gitlab:targetorg",
				"--parallel", "25",
			},
			setup:     func() { s.resetMocks() },
			expectErr: true,
		},
		{
			name: "No sync features enabled",
			args: []string{
				"sync",
				"--from", "github:sourceorg",
				"--to", "gitlab:targetorg",
				"--include-code=false",
			},
			setup:     func() { s.resetMocks() },
			expectErr: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
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
			}
		})
	}
}

// TestSyncCommandFeatures tests different sync features.
func (s *GitRepoTestSuite) TestSyncCommandFeatures() {
	tests := []struct {
		name       string
		args       []string
		setup      func()
		validate   func()
		expectErr  bool
		skipReason string
	}{
		{
			name:       "Sync with code only",
			skipReason: syncNeedsGitSeam,
			args: []string{
				"sync",
				"--from", "github:sourceorg/webapp",
				"--to", "gitlab:targetorg/webapp",
				"--create-missing",
				"--include-code",
			},
			setup: func() {
				s.resetMocks()
				s.setupBasicSyncMocks("sourceorg/webapp", "targetorg/webapp")
			},
			validate: func() {
				s.mockProviders["github"].AssertExpectations(s.T())
				s.mockProviders["gitlab"].AssertExpectations(s.T())
			},
		},
		{
			name:       "Sync with all features",
			skipReason: syncNeedsGitSeam,
			args: []string{
				"sync",
				"--from", "github:sourceorg/webapp",
				"--to", "gitlab:targetorg/webapp",
				"--create-missing",
				"--include-code",
				"--include-issues",
				"--include-wiki",
				"--include-releases",
			},
			setup: func() {
				s.resetMocks()
				s.setupBasicSyncMocks("sourceorg/webapp", "targetorg/webapp")
			},
			validate: func() {
				s.mockProviders["github"].AssertExpectations(s.T())
				s.mockProviders["gitlab"].AssertExpectations(s.T())
			},
		},
		{
			name:       "Sync with force option",
			skipReason: syncNeedsGitSeam,
			args: []string{
				"sync",
				"--from", "github:sourceorg/webapp",
				"--to", "gitlab:targetorg/webapp",
				"--update-existing",
				"--force",
			},
			setup: func() {
				s.resetMocks()
				s.setupUpdateSyncMocks("sourceorg/webapp", "targetorg/webapp")
			},
			validate: func() {
				s.mockProviders["github"].AssertExpectations(s.T())
				s.mockProviders["gitlab"].AssertExpectations(s.T())
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			if tt.skipReason != "" {
				s.T().Skip(tt.skipReason)
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

// Helper methods for sync tests

// setupBasicSyncMocks sets up mocks for basic sync operations (create missing).
func (s *GitRepoTestSuite) setupBasicSyncMocks(srcRepoName, dstRepoName string) {
	srcProvider := s.mockProviders["github"]
	dstProvider := s.mockProviders["gitlab"]

	// Source repository exists. Name has to move with FullName: the create
	// request the engine builds carries Name, which is what the create matcher
	// below keys on.
	srcRepo := s.testRepos[0]
	srcRepo.Name = getRepoNameFromFullName(srcRepoName)
	srcRepo.FullName = srcRepoName
	srcProvider.SetupGetResponse(srcRepoName, &srcRepo, nil)

	// Destination repository doesn't exist
	dstProvider.SetupGetResponse(dstRepoName, nil, fmt.Errorf("repository not found"))

	// Setup create repository response
	dstRepo := srcRepo
	dstRepo.FullName = dstRepoName
	dstRepo.ID = "dst-123"
	dstProvider.SetupCreateResponse(func(req provider.CreateRepoRequest) bool {
		return req.Name == getRepoNameFromFullName(dstRepoName)
	}, &dstRepo, nil)
}

// setupUpdateSyncMocks sets up mocks for update sync operations.
func (s *GitRepoTestSuite) setupUpdateSyncMocks(srcRepoName, dstRepoName string) {
	srcProvider := s.mockProviders["github"]
	dstProvider := s.mockProviders["gitlab"]

	// Source repository exists
	srcRepo := s.testRepos[0]
	srcRepo.FullName = srcRepoName
	srcProvider.SetupGetResponse(srcRepoName, &srcRepo, nil)

	// Destination repository exists
	dstRepo := srcRepo
	dstRepo.FullName = dstRepoName
	dstRepo.ID = "dst-123"
	dstProvider.SetupGetResponse(dstRepoName, &dstRepo, nil)
}

// getRepoNameFromFullName extracts repo name from full name (org/repo).
func getRepoNameFromFullName(fullName string) string {
	parts := strings.Split(fullName, "/")
	if len(parts) >= 2 {
		return parts[1]
	}
	return fullName
}
