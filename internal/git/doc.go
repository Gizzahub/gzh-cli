// Package git provides Git operations abstraction and utilities for
// repository management within the GZH Manager system.
//
// This package defines interfaces and implementations for Git operations,
// enabling consistent repository management across different Git platforms
// and providing testable abstractions for Git functionality.
//
// Key Components:
//
// Git Interface:
//   - Repository cloning and initialization
//   - Branch and tag management
//   - Commit and history operations
//   - Remote repository management
//   - Status and diff operations
//
// Implementations:
//   - LibGit2Repository: Git operations using libgit2
//   - CommandLineGit: Git operations using git command
//   - MockGitRepository: Generated mock for unit testing
//
// Features:
//   - Cross-platform Git operations
//   - Authentication handling (SSH, HTTPS, tokens)
//   - Progress tracking for long operations
//   - Error handling and recovery
//   - Repository state validation
//
// Clone Strategies:
//   - reset: Hard reset to remote state
//   - pull: Merge remote changes with local
//   - fetch: Update remote tracking only
//
// Authentication Support:
//   - SSH key authentication
//   - Personal access tokens
//   - Username/password authentication
//   - Credential helper integration
//
// Example usage:
//
//	repo := git.NewRepository(path)
//	err := repo.Clone(url, options)
//	status, err := repo.Status()
//	err = repo.Pull(strategy)
//
// The abstraction enables consistent Git operations throughout the
// application while supporting comprehensive testing and different
// Git backend implementations.
package git
