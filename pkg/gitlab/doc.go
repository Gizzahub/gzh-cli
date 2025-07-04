// Package gitlab provides GitLab API integration for group and repository management.
//
// This package handles:
//   - GitLab API client implementation
//   - Group and subgroup repository listing
//   - Repository cloning with branch support
//   - Bulk refresh operations with multiple strategies
//   - Default branch detection
//   - Recursive subgroup discovery
//   - Repository synchronization with cleanup
//
// The package implements direct HTTP API calls to GitLab instances (primarily gitlab.com)
// and provides Git operations through command-line execution. It supports both
// individual repository operations and bulk group-wide operations with recursive
// subgroup support.
//
// Main functions:
//   - List: List all repositories in a GitLab group (including subgroups)
//   - Clone: Clone a specific repository with branch support
//   - RefreshAll: Bulk refresh with multiple strategies (reset, pull, fetch)
//   - GetDefaultBranch: Get the default branch name for a repository
//
// Key features:
//   - Recursive subgroup discovery and repository listing
//   - Multiple refresh strategies for different use cases:
//   - "reset": Hard reset + pull (discards local changes)
//   - "pull": Merge remote changes with local changes
//   - "fetch": Update remote tracking without changing working directory
//   - Automatic repository cleanup (removes repos not in group)
//   - Error handling with detailed error messages
//   - Directory synchronization utilities
//
// The package uses the GitLab v4 API and supports standard Git operations
// through subprocess execution. It's designed to work with public GitLab
// instances and follows GitLab's API conventions for groups and projects.
package gitlab
