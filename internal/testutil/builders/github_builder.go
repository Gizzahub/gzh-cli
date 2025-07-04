package builders

import (
	"github.com/gizzahub/gzh-manager-go/pkg/github"
)

// BulkCloneRequestBuilder provides a fluent interface for building GitHub bulk clone requests
type BulkCloneRequestBuilder struct {
	request *github.BulkCloneRequest
}

// NewBulkCloneRequestBuilder creates a new BulkCloneRequestBuilder with default values
func NewBulkCloneRequestBuilder() *BulkCloneRequestBuilder {
	return &BulkCloneRequestBuilder{
		request: &github.BulkCloneRequest{
			Organization: "test-org",
			TargetPath:   "/tmp/test",
			Strategy:     "reset",
			Concurrency:  1,
			Repositories: []string{},
		},
	}
}

// WithOrganization sets the organization name
func (b *BulkCloneRequestBuilder) WithOrganization(org string) *BulkCloneRequestBuilder {
	b.request.Organization = org
	return b
}

// WithTargetPath sets the target path
func (b *BulkCloneRequestBuilder) WithTargetPath(path string) *BulkCloneRequestBuilder {
	b.request.TargetPath = path
	return b
}

// WithStrategy sets the clone strategy
func (b *BulkCloneRequestBuilder) WithStrategy(strategy string) *BulkCloneRequestBuilder {
	b.request.Strategy = strategy
	return b
}

// WithConcurrency sets the concurrency level
func (b *BulkCloneRequestBuilder) WithConcurrency(concurrency int) *BulkCloneRequestBuilder {
	b.request.Concurrency = concurrency
	return b
}

// WithRepositories sets the list of repositories
func (b *BulkCloneRequestBuilder) WithRepositories(repos []string) *BulkCloneRequestBuilder {
	b.request.Repositories = repos
	return b
}

// WithRepository adds a single repository
func (b *BulkCloneRequestBuilder) WithRepository(repo string) *BulkCloneRequestBuilder {
	b.request.Repositories = append(b.request.Repositories, repo)
	return b
}

// Build returns the constructed bulk clone request
func (b *BulkCloneRequestBuilder) Build() *github.BulkCloneRequest {
	return b.request
}

// BulkCloneResultBuilder provides a fluent interface for building GitHub bulk clone results
type BulkCloneResultBuilder struct {
	result *github.BulkCloneResult
}

// NewBulkCloneResultBuilder creates a new BulkCloneResultBuilder with default values
func NewBulkCloneResultBuilder() *BulkCloneResultBuilder {
	return &BulkCloneResultBuilder{
		result: &github.BulkCloneResult{
			TotalRepositories:    0,
			SuccessfulOperations: 0,
			FailedOperations:     0,
			SkippedRepositories:  0,
			OperationResults:     []github.RepositoryOperationResult{},
		},
	}
}

// WithTotalRepositories sets the total number of repositories
func (b *BulkCloneResultBuilder) WithTotalRepositories(total int) *BulkCloneResultBuilder {
	b.result.TotalRepositories = total
	return b
}

// WithSuccessfulOperations sets the number of successful operations
func (b *BulkCloneResultBuilder) WithSuccessfulOperations(successful int) *BulkCloneResultBuilder {
	b.result.SuccessfulOperations = successful
	return b
}

// WithFailedOperations sets the number of failed operations
func (b *BulkCloneResultBuilder) WithFailedOperations(failed int) *BulkCloneResultBuilder {
	b.result.FailedOperations = failed
	return b
}

// WithSkippedRepositories sets the number of skipped repositories
func (b *BulkCloneResultBuilder) WithSkippedRepositories(skipped int) *BulkCloneResultBuilder {
	b.result.SkippedRepositories = skipped
	return b
}

// WithOperationResult adds a single operation result
func (b *BulkCloneResultBuilder) WithOperationResult(repo, operation string, success bool, errorMsg string) *BulkCloneResultBuilder {
	result := github.RepositoryOperationResult{
		Repository: repo,
		Operation:  operation,
		Success:    success,
	}
	if errorMsg != "" {
		result.Error = errorMsg
	}

	b.result.OperationResults = append(b.result.OperationResults, result)
	return b
}

// WithSuccessfulClone adds a successful clone operation result
func (b *BulkCloneResultBuilder) WithSuccessfulClone(repo string) *BulkCloneResultBuilder {
	return b.WithOperationResult(repo, "clone", true, "")
}

// WithFailedClone adds a failed clone operation result
func (b *BulkCloneResultBuilder) WithFailedClone(repo, errorMsg string) *BulkCloneResultBuilder {
	return b.WithOperationResult(repo, "clone", false, errorMsg)
}

// WithSkippedRepository adds a skipped repository operation result
func (b *BulkCloneResultBuilder) WithSkippedRepository(repo, reason string) *BulkCloneResultBuilder {
	return b.WithOperationResult(repo, "skip", false, reason)
}

// Build returns the constructed bulk clone result
func (b *BulkCloneResultBuilder) Build() *github.BulkCloneResult {
	return b.result
}

// RepositoryInfoBuilder provides a fluent interface for building repository information
type RepositoryInfoBuilder struct {
	info *github.RepositoryInfo
}

// NewRepositoryInfoBuilder creates a new RepositoryInfoBuilder with default values
func NewRepositoryInfoBuilder() *RepositoryInfoBuilder {
	return &RepositoryInfoBuilder{
		info: &github.RepositoryInfo{
			Name:          "test-repo",
			FullName:      "test-org/test-repo",
			DefaultBranch: "main",
			CloneURL:      "https://github.com/test-org/test-repo.git",
			SSHURL:        "git@github.com:test-org/test-repo.git",
			Private:       false,
			Description:   "Test repository",
		},
	}
}

// WithName sets the repository name
func (b *RepositoryInfoBuilder) WithName(name string) *RepositoryInfoBuilder {
	b.info.Name = name
	return b
}

// WithFullName sets the full repository name
func (b *RepositoryInfoBuilder) WithFullName(fullName string) *RepositoryInfoBuilder {
	b.info.FullName = fullName
	return b
}

// WithDefaultBranch sets the default branch
func (b *RepositoryInfoBuilder) WithDefaultBranch(branch string) *RepositoryInfoBuilder {
	b.info.DefaultBranch = branch
	return b
}

// WithCloneURL sets the clone URL
func (b *RepositoryInfoBuilder) WithCloneURL(url string) *RepositoryInfoBuilder {
	b.info.CloneURL = url
	return b
}

// WithSSHURL sets the SSH URL
func (b *RepositoryInfoBuilder) WithSSHURL(url string) *RepositoryInfoBuilder {
	b.info.SSHURL = url
	return b
}

// WithPrivate sets the private flag
func (b *RepositoryInfoBuilder) WithPrivate(private bool) *RepositoryInfoBuilder {
	b.info.Private = private
	return b
}

// WithDescription sets the description
func (b *RepositoryInfoBuilder) WithDescription(description string) *RepositoryInfoBuilder {
	b.info.Description = description
	return b
}

// WithOrganization sets the organization and updates URLs accordingly
func (b *RepositoryInfoBuilder) WithOrganization(org string) *RepositoryInfoBuilder {
	b.info.FullName = org + "/" + b.info.Name
	b.info.CloneURL = "https://github.com/" + org + "/" + b.info.Name + ".git"
	b.info.SSHURL = "git@github.com:" + org + "/" + b.info.Name + ".git"
	return b
}

// Build returns the constructed repository information
func (b *RepositoryInfoBuilder) Build() *github.RepositoryInfo {
	return b.info
}

// RepositoryFiltersBuilder provides a fluent interface for building repository filters
type RepositoryFiltersBuilder struct {
	filters *github.RepositoryFilters
}

// NewRepositoryFiltersBuilder creates a new RepositoryFiltersBuilder
func NewRepositoryFiltersBuilder() *RepositoryFiltersBuilder {
	return &RepositoryFiltersBuilder{
		filters: &github.RepositoryFilters{
			IncludeNames: []string{},
			ExcludeNames: []string{},
		},
	}
}

// WithIncludeNames sets the include names filter
func (b *RepositoryFiltersBuilder) WithIncludeNames(names []string) *RepositoryFiltersBuilder {
	b.filters.IncludeNames = names
	return b
}

// WithExcludeNames sets the exclude names filter
func (b *RepositoryFiltersBuilder) WithExcludeNames(names []string) *RepositoryFiltersBuilder {
	b.filters.ExcludeNames = names
	return b
}

// WithIncludeName adds a single include name
func (b *RepositoryFiltersBuilder) WithIncludeName(name string) *RepositoryFiltersBuilder {
	b.filters.IncludeNames = append(b.filters.IncludeNames, name)
	return b
}

// WithExcludeName adds a single exclude name
func (b *RepositoryFiltersBuilder) WithExcludeName(name string) *RepositoryFiltersBuilder {
	b.filters.ExcludeNames = append(b.filters.ExcludeNames, name)
	return b
}

// Build returns the constructed repository filters
func (b *RepositoryFiltersBuilder) Build() *github.RepositoryFilters {
	return b.filters
}
