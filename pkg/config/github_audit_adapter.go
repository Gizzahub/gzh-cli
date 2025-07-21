// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package config

import (
	"context"
	"fmt"
	"time"

	"github.com/gizzahub/gzh-manager-go/pkg/github"
)

// GitHubAuditAdapter provides audit functionality for GitHub repositories.
type GitHubAuditAdapter struct {
	client *github.RepoConfigClient
}

// NewGitHubAuditAdapter creates a new GitHub audit adapter.
func NewGitHubAuditAdapter(client *github.RepoConfigClient) *GitHubAuditAdapter {
	return &GitHubAuditAdapter{
		client: client,
	}
}

// RunComplianceAudit performs a compliance audit for all repositories in an organization.
func (a *GitHubAuditAdapter) RunComplianceAudit(ctx context.Context, configPath, org string) (*AuditReport, error) {
	// Load the repository configuration
	repoConfig, err := LoadRepoConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Collect repository states from GitHub
	githubStates, err := a.client.CollectRepositoryStates(ctx, org)
	if err != nil {
		return nil, fmt.Errorf("failed to collect repository states: %w", err)
	}

	// Convert GitHub states to config states
	repoStates := make(map[string]RepositoryState)
	for name, ghState := range githubStates {
		repoStates[name] = convertGitHubStateToConfigState(ghState)
	}

	// Run the compliance audit
	return repoConfig.RunComplianceAudit(repoStates)
}

// convertGitHubStateToConfigState converts GitHub state data to config state.
func convertGitHubStateToConfigState(ghState github.RepositoryStateData) RepositoryState {
	state := RepositoryState{
		Name:                ghState.Name,
		Private:             ghState.Private,
		Archived:            ghState.Archived,
		HasIssues:           ghState.HasIssues,
		HasWiki:             ghState.HasWiki,
		HasProjects:         ghState.HasProjects,
		HasDownloads:        ghState.HasDownloads,
		VulnerabilityAlerts: ghState.VulnerabilityAlerts,
		SecurityAdvisories:  ghState.SecurityAdvisories,
		Files:               ghState.Files,
		Workflows:           ghState.Workflows,
		BranchProtection:    make(map[string]BranchProtectionState),
	}

	// Parse last modified time
	if ghState.LastModified != "" {
		if t, err := time.Parse(time.RFC3339, ghState.LastModified); err == nil {
			state.LastModified = t
		}
	}

	// Convert branch protection data
	for branch, ghProtection := range ghState.BranchProtection {
		state.BranchProtection[branch] = BranchProtectionState{
			Protected:       ghProtection.Protected,
			RequiredReviews: ghProtection.RequiredReviews,
			EnforceAdmins:   ghProtection.EnforceAdmins,
		}
	}

	return state
}
