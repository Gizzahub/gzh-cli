package github

import (
	"context"
)

// GitHubManager provides a high-level facade for GitHub operations
type GitHubManager interface {
	// Repository Operations
	ListOrganizationRepositories(ctx context.Context, organization string) ([]string, error)
	CloneRepository(ctx context.Context, organization, repository, targetPath string) error
	GetRepositoryDefaultBranch(ctx context.Context, organization, repository string) (string, error)

	// Bulk Operations
	RefreshAllRepositories(ctx context.Context, targetPath, organization, strategy string) error
	BulkCloneRepositories(ctx context.Context, request *BulkCloneRequest) (*BulkCloneResult, error)

	// Repository Management
	GetRepositoryInfo(ctx context.Context, organization, repository string) (*RepositoryInfo, error)
	ValidateRepositoryAccess(ctx context.Context, organization, repository string) error
}

// BulkCloneRequest represents a request for bulk repository operations
type BulkCloneRequest struct {
	Organization string
	TargetPath   string
	Strategy     string
	Repositories []string // if empty, clone all repositories
	Filters      *RepositoryFilters
	Concurrency  int
}

// BulkCloneResult represents the result of bulk operations
type BulkCloneResult struct {
	TotalRepositories    int
	SuccessfulOperations int
	FailedOperations     int
	SkippedRepositories  int
	OperationResults     []RepositoryOperationResult
	ExecutionTime        string
}

// RepositoryOperationResult represents the result of a single repository operation
type RepositoryOperationResult struct {
	Repository string
	Operation  string
	Success    bool
	Error      string
	Duration   string
}

// Note: RepositoryInfo is defined in interfaces.go

// RepositoryFilters contains filtering criteria for repositories
type RepositoryFilters struct {
	IncludeNames    []string
	ExcludeNames    []string
	IncludePrivate  bool
	IncludePublic   bool
	Languages       []string
	SizeLimit       int64
	LastUpdatedDays int
}

// gitHubManagerImpl implements the GitHubManager interface
type gitHubManagerImpl struct {
	factory GitHubProviderFactory
	logger  Logger
}

// NewGitHubManager creates a new GitHub manager facade
func NewGitHubManager(factory GitHubProviderFactory, logger Logger) GitHubManager {
	return &gitHubManagerImpl{
		factory: factory,
		logger:  logger,
	}
}

// ListOrganizationRepositories lists all repositories in an organization
func (g *gitHubManagerImpl) ListOrganizationRepositories(ctx context.Context, organization string) ([]string, error) {
	g.logger.Debug("Listing repositories for organization", "org", organization)

	// Use existing List function with context support
	return List(ctx, organization)
}

// CloneRepository clones a single repository
func (g *gitHubManagerImpl) CloneRepository(ctx context.Context, organization, repository, targetPath string) error {
	g.logger.Debug("Cloning repository", "org", organization, "repo", repository, "path", targetPath)

	// Use existing Clone function with context support
	return Clone(ctx, targetPath, organization, repository)
}

// GetRepositoryDefaultBranch gets the default branch for a repository
func (g *gitHubManagerImpl) GetRepositoryDefaultBranch(ctx context.Context, organization, repository string) (string, error) {
	g.logger.Debug("Getting default branch", "org", organization, "repo", repository)

	// Use existing GetDefaultBranch function
	return GetDefaultBranch(ctx, organization, repository)
}

// RefreshAllRepositories refreshes all repositories in an organization
func (g *gitHubManagerImpl) RefreshAllRepositories(ctx context.Context, targetPath, organization, strategy string) error {
	g.logger.Info("Refreshing all repositories", "org", organization, "strategy", strategy)

	// Use existing RefreshAll function
	return RefreshAll(ctx, targetPath, organization, strategy)
}

// BulkCloneRepositories performs bulk repository operations
func (g *gitHubManagerImpl) BulkCloneRepositories(ctx context.Context, request *BulkCloneRequest) (*BulkCloneResult, error) {
	g.logger.Info("Starting bulk clone operation", "org", request.Organization)

	result := &BulkCloneResult{
		OperationResults: make([]RepositoryOperationResult, 0),
	}

	// Get list of repositories to clone
	repositories := request.Repositories
	if len(repositories) == 0 {
		repos, err := g.ListOrganizationRepositories(ctx, request.Organization)
		if err != nil {
			return nil, err
		}
		repositories = repos
	}

	// Apply filters if provided
	if request.Filters != nil {
		repositories = g.applyFilters(repositories, request.Filters)
	}

	result.TotalRepositories = len(repositories)

	// Clone repositories (simplified implementation)
	for _, repo := range repositories {
		opResult := RepositoryOperationResult{
			Repository: repo,
			Operation:  "clone",
		}

		err := g.CloneRepository(ctx, request.Organization, repo, request.TargetPath)
		if err != nil {
			opResult.Success = false
			opResult.Error = err.Error()
			result.FailedOperations++
		} else {
			opResult.Success = true
			result.SuccessfulOperations++
		}

		result.OperationResults = append(result.OperationResults, opResult)
	}

	return result, nil
}

// GetRepositoryInfo gets detailed information about a repository
func (g *gitHubManagerImpl) GetRepositoryInfo(ctx context.Context, organization, repository string) (*RepositoryInfo, error) {
	g.logger.Debug("Getting repository info", "org", organization, "repo", repository)

	// Implementation would make API call to get detailed repository information
	// For now, return basic info
	defaultBranch, err := g.GetRepositoryDefaultBranch(ctx, organization, repository)
	if err != nil {
		return nil, err
	}

	return &RepositoryInfo{
		Name:          repository,
		FullName:      organization + "/" + repository,
		DefaultBranch: defaultBranch,
		CloneURL:      "https://github.com/" + organization + "/" + repository + ".git",
		SSHURL:        "git@github.com:" + organization + "/" + repository + ".git",
	}, nil
}

// ValidateRepositoryAccess validates that the user has access to a repository
func (g *gitHubManagerImpl) ValidateRepositoryAccess(ctx context.Context, organization, repository string) error {
	g.logger.Debug("Validating repository access", "org", organization, "repo", repository)

	// Implementation would check access permissions
	// For now, just try to get repository info
	_, err := g.GetRepositoryInfo(ctx, organization, repository)
	return err
}

// applyFilters applies repository filters to a list of repositories
func (g *gitHubManagerImpl) applyFilters(repositories []string, filters *RepositoryFilters) []string {
	if filters == nil {
		return repositories
	}

	filtered := make([]string, 0, len(repositories))

	for _, repo := range repositories {
		// Apply include/exclude name filters
		if len(filters.IncludeNames) > 0 {
			found := false
			for _, include := range filters.IncludeNames {
				if repo == include {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		excluded := false
		for _, exclude := range filters.ExcludeNames {
			if repo == exclude {
				excluded = true
				break
			}
		}
		if excluded {
			continue
		}

		filtered = append(filtered, repo)
	}

	return filtered
}
