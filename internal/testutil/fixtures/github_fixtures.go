// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package fixtures

import (
	"github.com/Gizzahub/gzh-manager-go/internal/testutil/builders"
	"github.com/Gizzahub/gzh-manager-go/pkg/github"
)

// GitHubFixtures provides common GitHub-related fixtures for tests.
type GitHubFixtures struct{}

// NewGitHubFixtures creates a new GitHubFixtures instance.
func NewGitHubFixtures() *GitHubFixtures {
	return &GitHubFixtures{}
}

// SimpleBulkCloneRequest returns a simple bulk clone request.
func (f *GitHubFixtures) SimpleBulkCloneRequest() *github.BulkCloneRequest {
	return builders.NewBulkCloneRequestBuilder().
		WithOrganization("test-org").
		WithTargetPath("/tmp/test").
		WithStrategy("reset").
		WithConcurrency(2).
		WithRepositories([]string{"repo1", "repo2"}).
		Build()
}

// LargeBulkCloneRequest returns a bulk clone request with many repositories.
func (f *GitHubFixtures) LargeBulkCloneRequest() *github.BulkCloneRequest {
	repos := make([]string, 100)
	for i := 0; i < 100; i++ {
		repos[i] = "repo" + string(rune(i))
	}

	return builders.NewBulkCloneRequestBuilder().
		WithOrganization("large-org").
		WithTargetPath("/tmp/large").
		WithStrategy("reset").
		WithConcurrency(10).
		WithRepositories(repos).
		Build()
}

// SuccessfulBulkCloneResult returns a successful bulk clone result.
func (f *GitHubFixtures) SuccessfulBulkCloneResult() *github.BulkCloneResult {
	return builders.NewBulkCloneResultBuilder().
		WithTotalRepositories(3).
		WithSuccessfulOperations(3).
		WithFailedOperations(0).
		WithSkippedRepositories(0).
		WithSuccessfulClone("repo1").
		WithSuccessfulClone("repo2").
		WithSuccessfulClone("repo3").
		Build()
}

// MixedBulkCloneResult returns a bulk clone result with mixed outcomes.
func (f *GitHubFixtures) MixedBulkCloneResult() *github.BulkCloneResult {
	return builders.NewBulkCloneResultBuilder().
		WithTotalRepositories(5).
		WithSuccessfulOperations(3).
		WithFailedOperations(1).
		WithSkippedRepositories(1).
		WithSuccessfulClone("repo1").
		WithSuccessfulClone("repo2").
		WithSuccessfulClone("repo3").
		WithFailedClone("repo4", "network error").
		WithSkippedRepository("repo5", "already exists").
		Build()
}

// FailedBulkCloneResult returns a bulk clone result with all failures.
func (f *GitHubFixtures) FailedBulkCloneResult() *github.BulkCloneResult {
	return builders.NewBulkCloneResultBuilder().
		WithTotalRepositories(2).
		WithSuccessfulOperations(0).
		WithFailedOperations(2).
		WithSkippedRepositories(0).
		WithFailedClone("repo1", "authentication failed").
		WithFailedClone("repo2", "repository not found").
		Build()
}

// SimpleRepositoryInfo returns a simple repository information object.
func (f *GitHubFixtures) SimpleRepositoryInfo() *github.RepositoryInfo {
	return builders.NewRepositoryInfoBuilder().
		WithName("test-repo").
		WithOrganization("test-org").
		WithDefaultBranch("main").
		WithPrivate(false).
		WithDescription("Test repository").
		Build()
}

// PrivateRepositoryInfo returns a private repository information object.
func (f *GitHubFixtures) PrivateRepositoryInfo() *github.RepositoryInfo {
	return builders.NewRepositoryInfoBuilder().
		WithName("private-repo").
		WithOrganization("private-org").
		WithDefaultBranch("main").
		WithPrivate(true).
		WithDescription("Private test repository").
		Build()
}

// LegacyRepositoryInfo returns a repository with legacy default branch.
func (f *GitHubFixtures) LegacyRepositoryInfo() *github.RepositoryInfo {
	return builders.NewRepositoryInfoBuilder().
		WithName("legacy-repo").
		WithOrganization("legacy-org").
		WithDefaultBranch("master").
		WithPrivate(false).
		WithDescription("Legacy repository with master branch").
		Build()
}

// RepositoryInfoList returns a list of repository information objects.
func (f *GitHubFixtures) RepositoryInfoList() []*github.RepositoryInfo {
	return []*github.RepositoryInfo{
		f.SimpleRepositoryInfo(),
		f.PrivateRepositoryInfo(),
		f.LegacyRepositoryInfo(),
	}
}

// SimpleRepositoryFilters returns simple repository filters.
func (f *GitHubFixtures) SimpleRepositoryFilters() *github.RepositoryFilters {
	return builders.NewRepositoryFiltersBuilder().
		WithIncludeNames([]string{"repo1", "test-repo"}).
		WithExcludeNames([]string{"repo2"}).
		Build()
}

// IncludeOnlyFilters returns filters that only include specific repositories.
func (f *GitHubFixtures) IncludeOnlyFilters() *github.RepositoryFilters {
	return builders.NewRepositoryFiltersBuilder().
		WithIncludeNames([]string{"important-repo", "critical-repo"}).
		Build()
}

// ExcludeOnlyFilters returns filters that exclude specific repositories.
func (f *GitHubFixtures) ExcludeOnlyFilters() *github.RepositoryFilters {
	return builders.NewRepositoryFiltersBuilder().
		WithExcludeNames([]string{"test-repo", "temp-repo", "archive-repo"}).
		Build()
}

// EmptyFilters returns empty repository filters.
func (f *GitHubFixtures) EmptyFilters() *github.RepositoryFilters {
	return builders.NewRepositoryFiltersBuilder().Build()
}

// TestRepositoryList returns a list of test repository names.
func (f *GitHubFixtures) TestRepositoryList() []string {
	return []string{
		"repo1",
		"repo2",
		"test-repo",
		"another-repo",
		"important-repo",
		"temp-repo",
		"archive-repo",
	}
}

// LargeRepositoryList returns a large list of repository names for performance testing.
func (f *GitHubFixtures) LargeRepositoryList() []string {
	repos := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		repos[i] = "repo" + string(rune(i))
	}

	return repos
}
