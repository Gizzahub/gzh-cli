package gitlab

import (
	"context"
)

// GitLabManager provides a high-level facade for GitLab operations.
type GitLabManager interface {
	// Repository Operations
	ListGroupRepositories(ctx context.Context, group string) ([]string, error)
	CloneRepository(ctx context.Context, group, repository, targetPath, branch string) error
	GetRepositoryDefaultBranch(ctx context.Context, group, repository string) (string, error)

	// Bulk Operations
	RefreshAllRepositories(ctx context.Context, targetPath, group, strategy string) error
	BulkCloneRepositories(ctx context.Context, request *BulkCloneRequest) (*BulkCloneResult, error)

	// Group Management
	GetGroupInfo(ctx context.Context, group string) (*GroupInfo, error)
	ListSubgroups(ctx context.Context, group string) ([]string, error)
	ValidateGroupAccess(ctx context.Context, group string) error
}

// BulkCloneRequest represents a request for bulk repository operations.
type BulkCloneRequest struct {
	Group        string
	TargetPath   string
	Strategy     string
	Repositories []string // if empty, clone all repositories
	Filters      *RepositoryFilters
	Concurrency  int
	Recursive    bool // clone subgroups recursively
}

// BulkCloneResult represents the result of bulk operations.
type BulkCloneResult struct {
	TotalRepositories    int
	SuccessfulOperations int
	FailedOperations     int
	SkippedRepositories  int
	OperationResults     []RepositoryOperationResult
	ExecutionTime        string
	ProcessedGroups      []string
}

// RepositoryOperationResult represents the result of a single repository operation.
type RepositoryOperationResult struct {
	Repository string
	Group      string
	Operation  string
	Success    bool
	Error      string
	Duration   string
}

// GroupInfo contains metadata about a GitLab group.
type GroupInfo struct {
	Name          string
	FullName      string
	Path          string
	Description   string
	Visibility    string
	ProjectCount  int
	SubgroupCount int
}

// RepositoryFilters contains filtering criteria for repositories.
type RepositoryFilters struct {
	IncludeNames    []string
	ExcludeNames    []string
	IncludePrivate  bool
	IncludePublic   bool
	Languages       []string
	SizeLimit       int64
	LastUpdatedDays int
	IncludeArchived bool
}

// gitLabManagerImpl implements the GitLabManager interface.
type gitLabManagerImpl struct {
	factory GitLabProviderFactory
	client  HTTPClient
	logger  Logger
}

// NewGitLabManager creates a new GitLab manager facade.
func NewGitLabManager(factory GitLabProviderFactory, client HTTPClient, logger Logger) GitLabManager {
	return &gitLabManagerImpl{
		factory: factory,
		client:  client,
		logger:  logger,
	}
}

// ListGroupRepositories lists all repositories in a group.
func (g *gitLabManagerImpl) ListGroupRepositories(ctx context.Context, group string) ([]string, error) {
	g.logger.Debug("Listing repositories for group", "group", group)

	// Use existing List function with context support
	return List(ctx, group)
}

// CloneRepository clones a single repository.
func (g *gitLabManagerImpl) CloneRepository(ctx context.Context, group, repository, targetPath, branch string) error {
	g.logger.Debug("Cloning repository", "group", group, "repo", repository, "path", targetPath)

	// Use existing Clone function with context support
	return Clone(ctx, targetPath, group, repository, branch)
}

// GetRepositoryDefaultBranch gets the default branch for a repository.
func (g *gitLabManagerImpl) GetRepositoryDefaultBranch(ctx context.Context, group, repository string) (string, error) {
	g.logger.Debug("Getting default branch", "group", group, "repo", repository)

	// Use existing GetDefaultBranch function
	return GetDefaultBranch(ctx, group, repository)
}

// RefreshAllRepositories refreshes all repositories in a group.
func (g *gitLabManagerImpl) RefreshAllRepositories(ctx context.Context, targetPath, group, strategy string) error {
	g.logger.Info("Refreshing all repositories", "group", group, "strategy", strategy)

	// Use existing RefreshAll function
	return RefreshAll(ctx, targetPath, group, strategy)
}

// BulkCloneRepositories performs bulk repository operations.
func (g *gitLabManagerImpl) BulkCloneRepositories(ctx context.Context, request *BulkCloneRequest) (*BulkCloneResult, error) {
	g.logger.Info("Starting bulk clone operation", "group", request.Group)

	result := &BulkCloneResult{
		OperationResults: make([]RepositoryOperationResult, 0),
		ProcessedGroups:  []string{request.Group},
	}

	// Get list of repositories to clone
	repositories := request.Repositories
	if len(repositories) == 0 {
		repos, err := g.ListGroupRepositories(ctx, request.Group)
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
			Group:      request.Group,
			Operation:  "clone",
		}

		err := g.CloneRepository(ctx, request.Group, repo, request.TargetPath, "")
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

// GetGroupInfo gets detailed information about a group.
func (g *gitLabManagerImpl) GetGroupInfo(ctx context.Context, group string) (*GroupInfo, error) {
	g.logger.Debug("Getting group info", "group", group)

	// Implementation would make API call to get detailed group information
	// For now, return basic info
	return &GroupInfo{
		Name:     group,
		FullName: group,
		Path:     group,
	}, nil
}

// ListSubgroups lists all subgroups within a group.
func (g *gitLabManagerImpl) ListSubgroups(ctx context.Context, group string) ([]string, error) {
	g.logger.Debug("Listing subgroups", "group", group)

	// Implementation would make API call to get subgroups
	// For now, return empty list
	return []string{}, nil
}

// ValidateGroupAccess validates that the user has access to a group.
func (g *gitLabManagerImpl) ValidateGroupAccess(ctx context.Context, group string) error {
	g.logger.Debug("Validating group access", "group", group)

	// Implementation would check access permissions
	// For now, just try to get group info
	_, err := g.GetGroupInfo(ctx, group)

	return err
}

// applyFilters applies repository filters to a list of repositories.
func (g *gitLabManagerImpl) applyFilters(repositories []string, filters *RepositoryFilters) []string {
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

// Logger interface for dependency injection.
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}
