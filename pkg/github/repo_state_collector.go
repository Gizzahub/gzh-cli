package github

import (
	"context"
)

// RepositoryStateData represents the raw state data collected from GitHub
// This is a simple data structure with no dependencies on other packages.
type RepositoryStateData struct {
	Name         string
	Private      bool
	Archived     bool
	HasIssues    bool
	HasWiki      bool
	HasProjects  bool
	HasDownloads bool

	// Branch protection
	BranchProtection map[string]BranchProtectionData

	// Security features
	VulnerabilityAlerts bool
	SecurityAdvisories  bool

	// Files present
	Files []string

	// Workflows
	Workflows []string

	// Last modified
	LastModified string // ISO 8601 format
}

// BranchProtectionData represents raw branch protection data.
type BranchProtectionData struct {
	Protected       bool
	RequiredReviews int
	EnforceAdmins   bool
}

// CollectRepositoryStates collects state data for all repositories in the organization.
func (c *RepoConfigClient) CollectRepositoryStates(ctx context.Context, org string) (map[string]RepositoryStateData, error) {
	repos, err := c.ListRepositories(ctx, org, nil)
	if err != nil {
		return nil, err
	}

	states := make(map[string]RepositoryStateData)

	for _, repo := range repos {
		state, err := c.collectRepositoryState(ctx, org, repo)
		if err != nil {
			// Log error but continue with other repos
			continue
		}

		states[repo.Name] = state
	}

	return states, nil
}

// collectRepositoryState collects the current state of a repository.
func (c *RepoConfigClient) collectRepositoryState(ctx context.Context, org string, repo *Repository) (RepositoryStateData, error) {
	state := RepositoryStateData{
		Name:             repo.Name,
		Private:          repo.Private,
		Archived:         repo.Archived,
		HasIssues:        repo.HasIssues,
		HasWiki:          repo.HasWiki,
		HasProjects:      repo.HasProjects,
		HasDownloads:     repo.HasDownloads,
		LastModified:     repo.UpdatedAt, // Already a string in the Repository struct
		BranchProtection: make(map[string]BranchProtectionData),
	}

	// Get security features - simplified for now
	// Note: Security features would need additional API calls to GitHub
	// For now, we'll set defaults
	state.VulnerabilityAlerts = false
	state.SecurityAdvisories = false

	// Get branch protection for default branch
	if repo.DefaultBranch != "" {
		protection, err := c.GetBranchProtection(ctx, org, repo.Name, repo.DefaultBranch)
		if err == nil && protection != nil {
			state.BranchProtection[repo.DefaultBranch] = BranchProtectionData{
				Protected:       true,
				RequiredReviews: getBranchProtectionRequiredReviews(protection),
				EnforceAdmins:   protection.EnforceAdmins,
			}
		}
	}

	// Check for specific files
	state.Files = c.checkForFiles(ctx, org, repo.Name)

	// Check for workflows
	state.Workflows = c.listRepoWorkflows(ctx, org, repo.Name)

	return state, nil
}

// getBranchProtectionRequiredReviews extracts the required review count from branch protection.
func getBranchProtectionRequiredReviews(protection *BranchProtection) int {
	if protection.RequiredPullRequestReviews == nil {
		return 0
	}

	return protection.RequiredPullRequestReviews.RequiredApprovingReviewCount
}

// checkForFiles checks for the existence of specific files in the repository.
func (c *RepoConfigClient) checkForFiles(ctx context.Context, org, repoName string) []string {
	var foundFiles []string

	// List of files to check
	filesToCheck := []string{
		"README.md",
		"LICENSE",
		"SECURITY.md",
		"CONTRIBUTING.md",
		"CODE_OF_CONDUCT.md",
		".github/CODEOWNERS",
		"COMPLIANCE.md",
	}

	for _, file := range filesToCheck {
		// Skip file checking for now since we'd need to implement GetContents
		// or use the existing HTTP client directly
		// This is a placeholder that should be implemented with proper API calls
		_ = file
	}

	return foundFiles
}

// listRepoWorkflows lists GitHub Actions workflows in the repository.
func (c *RepoConfigClient) listRepoWorkflows(ctx context.Context, org, repoName string) []string {
	var workflows []string

	// Skip workflow listing for now since we'd need to implement GetContents
	// or use the existing HTTP client directly
	// This is a placeholder that should be implemented with proper API calls
	_ = org
	_ = repoName

	return workflows
}
