// Package gitea provides Gitea API integration for repository management and cloning operations.
//
// This package handles:
//   - Gitea API client implementation
//   - Organization repository listing
//   - Repository cloning with branch support
//   - Bulk refresh operations for existing repositories
//   - Default branch detection
//   - Error handling for Gitea-specific operations
//
// The package implements direct HTTP API calls to Gitea instances (primarily gitea.com)
// and provides Git operations through command-line execution. It supports both
// individual repository operations and bulk organization-wide operations.
//
// Main functions:
//   - List: List all repositories in a Gitea organization
//   - Clone: Clone a specific repository with branch support
//   - RefreshAll: Bulk refresh all repositories in an organization
//   - GetDefaultBranch: Get the default branch name for a repository
//
// Key features:
//   - Automatic default branch detection
//   - Organization-wide repository discovery
//   - Git command-line integration for cloning operations
//   - Error handling with detailed error messages
//
// The package uses the Gitea v1 API and supports standard Git operations
// through subprocess execution. It's designed to work with public Gitea
// instances and follows Gitea's API conventions.
package gitea
