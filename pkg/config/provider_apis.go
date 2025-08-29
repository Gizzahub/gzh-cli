// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package config

import (
	"context"

	"github.com/Gizzahub/gzh-cli/pkg/gitea"
	"github.com/Gizzahub/gzh-cli/pkg/github"
	"github.com/Gizzahub/gzh-cli/pkg/gitlab"
)

// GitHubAPI implements ProviderAPI for GitHub.
type GitHubAPI struct{}

// NewGitHubAPI creates a new GitHub API implementation.
func NewGitHubAPI() *GitHubAPI {
	return &GitHubAPI{}
}

// List lists repositories for a GitHub owner.
func (g *GitHubAPI) List(ctx context.Context, owner string) ([]string, error) {
	return github.List(ctx, owner)
}

// GetDefaultBranch gets the default branch for a GitHub repository.
func (g *GitHubAPI) GetDefaultBranch(ctx context.Context, owner, repository string) (string, error) {
	return github.GetDefaultBranch(ctx, owner, repository)
}

// Clone clones a GitHub repository.
func (g *GitHubAPI) Clone(ctx context.Context, targetPath, owner, repository string, _ ...string) error {
	return github.Clone(ctx, targetPath, owner, repository)
}

// RefreshAll refreshes all GitHub repositories.
func (g *GitHubAPI) RefreshAll(ctx context.Context, targetPath, owner string, extraParams ...string) error {
	strategy := ""
	if len(extraParams) > 0 {
		strategy = extraParams[0]
	}
	return github.RefreshAll(ctx, targetPath, owner, strategy)
}

// GitLabAPI implements ProviderAPI for GitLab.
type GitLabAPI struct{}

// NewGitLabAPI creates a new GitLab API implementation.
func NewGitLabAPI() *GitLabAPI {
	return &GitLabAPI{}
}

// List lists repositories for a GitLab owner.
func (g *GitLabAPI) List(ctx context.Context, owner string) ([]string, error) {
	return gitlab.List(ctx, owner)
}

// GetDefaultBranch gets the default branch for a GitLab repository.
func (g *GitLabAPI) GetDefaultBranch(ctx context.Context, owner, repository string) (string, error) {
	return gitlab.GetDefaultBranch(ctx, owner, repository)
}

// Clone clones a GitLab repository.
func (g *GitLabAPI) Clone(ctx context.Context, targetPath, owner, repository string, extraParams ...string) error {
	extraParam := ""
	if len(extraParams) > 0 {
		extraParam = extraParams[0]
	}
	return gitlab.Clone(ctx, targetPath, owner, repository, extraParam)
}

// RefreshAll refreshes all GitLab repositories.
func (g *GitLabAPI) RefreshAll(ctx context.Context, targetPath, owner string, extraParams ...string) error {
	strategy := ""
	if len(extraParams) > 0 {
		strategy = extraParams[0]
	}
	return gitlab.RefreshAll(ctx, targetPath, owner, strategy)
}

// GiteaAPI implements ProviderAPI for Gitea.
type GiteaAPI struct{}

// NewGiteaAPI creates a new Gitea API implementation.
func NewGiteaAPI() *GiteaAPI {
	return &GiteaAPI{}
}

// List lists repositories for a Gitea owner.
func (g *GiteaAPI) List(ctx context.Context, owner string) ([]string, error) {
	return gitea.List(ctx, owner)
}

// GetDefaultBranch gets the default branch for a Gitea repository.
func (g *GiteaAPI) GetDefaultBranch(ctx context.Context, owner, repository string) (string, error) {
	return gitea.GetDefaultBranch(ctx, owner, repository)
}

// Clone clones a Gitea repository.
func (g *GiteaAPI) Clone(ctx context.Context, targetPath, owner, repository string, extraParams ...string) error {
	extraParam := ""
	if len(extraParams) > 0 {
		extraParam = extraParams[0]
	}
	return gitea.Clone(ctx, targetPath, owner, repository, extraParam)
}

// RefreshAll refreshes all Gitea repositories.
func (g *GiteaAPI) RefreshAll(ctx context.Context, targetPath, owner string, _ ...string) error {
	// Note: gitea.RefreshAll doesn't support strategy parameter
	return gitea.RefreshAll(ctx, targetPath, owner)
}
