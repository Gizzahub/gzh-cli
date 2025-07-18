// Package gitplatform provides a common interface for different Git hosting platforms.
package gitplatform

import (
	"context"
)

// Repository represents a git repository with common fields across platforms.
type Repository struct {
	Name          string
	FullName      string
	URL           string
	SSHURL        string
	DefaultBranch string
	Private       bool
	Archived      bool
	Description   string
}

// GitPlatformClient defines the common interface for all git hosting platforms.
type GitPlatformClient interface {
	// GetDefaultBranch retrieves the default branch name for a repository
	GetDefaultBranch(ctx context.Context, owner, repo string) (string, error)

	// ListRepositories lists all repositories for an organization/user/group
	ListRepositories(ctx context.Context, owner string) ([]Repository, error)

	// Clone clones a single repository to the specified path using the given strategy
	Clone(ctx context.Context, repo Repository, targetPath, strategy string) error

	// RefreshAll refreshes all repositories in the target path using the given strategy
	RefreshAll(ctx context.Context, targetPath, strategy string) error

	// SetAuthentication sets the authentication token for API calls
	SetAuthentication(token string)

	// GetPlatformName returns the name of the platform (github, gitlab, gitea)
	GetPlatformName() string
}

// CloneStrategy represents the strategy for cloning/updating repositories.
type CloneStrategy string

const (
	// StrategyReset performs hard reset and pull (discards local changes).
	StrategyReset CloneStrategy = "reset"
	// StrategyPull performs git pull (merges changes).
	StrategyPull CloneStrategy = "pull"
	// StrategyFetch only fetches without changing working directory.
	StrategyFetch CloneStrategy = "fetch"
)

// ProviderConfig represents common configuration for a git provider.
type ProviderConfig struct {
	Name         string
	BaseURL      string
	Token        string
	Protocol     string // https or ssh
	TargetPath   string
	Strategy     CloneStrategy
	MatchPattern string // Optional regex pattern for filtering repositories
}
