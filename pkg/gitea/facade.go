package gitea

import (
	"context"
	"io"
	"net/http"
)

// GiteaManager provides a high-level facade for Gitea operations
type GiteaManager interface {
	// Repository Operations
	ListOrganizationRepositories(ctx context.Context, organization string) ([]string, error)
	CloneRepository(ctx context.Context, organization, repository, targetPath string) error
	GetRepositoryInfo(ctx context.Context, organization, repository string) (*RepositoryInfo, error)
	
	// Bulk Operations
	RefreshAllRepositories(ctx context.Context, targetPath, organization string) error
	BulkCloneRepositories(ctx context.Context, request *BulkCloneRequest) (*BulkCloneResult, error)
	
	// Organization Management
	GetOrganizationInfo(ctx context.Context, organization string) (*OrganizationInfo, error)
	ListUserOrganizations(ctx context.Context, username string) ([]string, error)
	ValidateOrganizationAccess(ctx context.Context, organization string) error
}

// BulkCloneRequest represents a request for bulk repository operations
type BulkCloneRequest struct {
	Organization string
	TargetPath   string
	Repositories []string // if empty, clone all repositories
	Filters      *RepositoryFilters
	Concurrency  int
}

// BulkCloneResult represents the result of bulk operations
type BulkCloneResult struct {
	TotalRepositories     int
	SuccessfulOperations  int
	FailedOperations      int
	SkippedRepositories   int
	OperationResults      []RepositoryOperationResult
	ExecutionTime         string
}

// RepositoryOperationResult represents the result of a single repository operation
type RepositoryOperationResult struct {
	Repository string
	Operation  string
	Success    bool
	Error      string
	Duration   string
}

// RepositoryInfo contains metadata about a repository
type RepositoryInfo struct {
	Name        string
	FullName    string
	IsPrivate   bool
	CloneURL    string
	SSHUrl      string
	LastUpdated string
	Description string
	Language    string
	Size        int64
}

// OrganizationInfo contains metadata about an organization
type OrganizationInfo struct {
	Name         string
	FullName     string
	Description  string
	Website      string
	Location     string
	Visibility   string
	RepoCount    int
	MemberCount  int
}

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

// giteaManagerImpl implements the GiteaManager interface
type giteaManagerImpl struct {
	factory GiteaProviderFactory
	client  HTTPClient
	logger  Logger
}

// NewGiteaManager creates a new Gitea manager facade
func NewGiteaManager(factory GiteaProviderFactory, client HTTPClient, logger Logger) GiteaManager {
	return &giteaManagerImpl{
		factory: factory,
		client:  client,
		logger:  logger,
	}
}

// ListOrganizationRepositories lists all repositories in an organization
func (g *giteaManagerImpl) ListOrganizationRepositories(ctx context.Context, organization string) ([]string, error) {
	g.logger.Debug("Listing repositories for organization", "org", organization)
	
	// Use existing List function with context support
	return List(organization)
}

// CloneRepository clones a single repository
func (g *giteaManagerImpl) CloneRepository(ctx context.Context, organization, repository, targetPath string) error {
	g.logger.Debug("Cloning repository", "org", organization, "repo", repository, "path", targetPath)
	
	// Use existing Clone function with context support
	return Clone(targetPath, organization, repository)
}

// GetRepositoryInfo gets detailed information about a repository
func (g *giteaManagerImpl) GetRepositoryInfo(ctx context.Context, organization, repository string) (*RepositoryInfo, error) {
	g.logger.Debug("Getting repository info", "org", organization, "repo", repository)
	
	// Implementation would make API call to get detailed repository information
	// For now, return basic info
	return &RepositoryInfo{
		Name:     repository,
		FullName: organization + "/" + repository,
		CloneURL: "https://gitea.example.com/" + organization + "/" + repository + ".git",
		SSHUrl:   "git@gitea.example.com:" + organization + "/" + repository + ".git",
	}, nil
}

// RefreshAllRepositories refreshes all repositories in an organization
func (g *giteaManagerImpl) RefreshAllRepositories(ctx context.Context, targetPath, organization string) error {
	g.logger.Info("Refreshing all repositories", "org", organization)
	
	// Use existing RefreshAll function
	return RefreshAll(targetPath, organization)
}

// BulkCloneRepositories performs bulk repository operations
func (g *giteaManagerImpl) BulkCloneRepositories(ctx context.Context, request *BulkCloneRequest) (*BulkCloneResult, error) {
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

// GetOrganizationInfo gets detailed information about an organization
func (g *giteaManagerImpl) GetOrganizationInfo(ctx context.Context, organization string) (*OrganizationInfo, error) {
	g.logger.Debug("Getting organization info", "org", organization)
	
	// Implementation would make API call to get detailed organization information
	// For now, return basic info
	return &OrganizationInfo{
		Name:     organization,
		FullName: organization,
	}, nil
}

// ListUserOrganizations lists all organizations for a user
func (g *giteaManagerImpl) ListUserOrganizations(ctx context.Context, username string) ([]string, error) {
	g.logger.Debug("Listing user organizations", "user", username)
	
	// Implementation would make API call to get user organizations
	// For now, return empty list
	return []string{}, nil
}

// ValidateOrganizationAccess validates that the user has access to an organization
func (g *giteaManagerImpl) ValidateOrganizationAccess(ctx context.Context, organization string) error {
	g.logger.Debug("Validating organization access", "org", organization)
	
	// Implementation would check access permissions
	// For now, just try to get organization info
	_, err := g.GetOrganizationInfo(ctx, organization)
	return err
}

// applyFilters applies repository filters to a list of repositories
func (g *giteaManagerImpl) applyFilters(repositories []string, filters *RepositoryFilters) []string {
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

// Note: HTTPClient and Logger interfaces are defined elsewhere