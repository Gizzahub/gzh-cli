// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package testlib

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// MockRepoFactory provides interface for creating test repositories
// with various Git states for synclone testing.
type MockRepoFactory interface {
	CreateBasicRepos(ctx context.Context, opts BasicRepoOptions) error
	CreateConflictRepos(ctx context.Context, opts ConflictRepoOptions) error
	CreateSpecialRepos(ctx context.Context, opts SpecialRepoOptions) error
}

// BasicRepoOptions defines options for creating basic test repositories.
type BasicRepoOptions struct {
	BaseDir     string
	RepoName    string
	InitialData bool
	Branches    []string
}

// ConflictRepoOptions defines options for creating conflict scenario repositories.
type ConflictRepoOptions struct {
	BaseDir      string
	RepoName     string
	ConflictType string // "merge", "rebase", "diverged"
	LocalChanges bool
}

// SpecialRepoOptions defines options for creating special scenario repositories.
type SpecialRepoOptions struct {
	BaseDir     string
	RepoName    string
	SpecialType string // "lfs", "submodule", "large"
	Size        int64  // for large repos
}

// DefaultMockRepoFactory implements MockRepoFactory interface.
type DefaultMockRepoFactory struct {
	timeout time.Duration
}

// NewMockRepoFactory creates a new MockRepoFactory instance.
func NewMockRepoFactory() MockRepoFactory {
	return &DefaultMockRepoFactory{
		timeout: 30 * time.Second,
	}
}

// CreateBasicRepos creates basic test repositories with standard Git structures.
func (f *DefaultMockRepoFactory) CreateBasicRepos(ctx context.Context, opts BasicRepoOptions) error {
	if opts.BaseDir == "" {
		return fmt.Errorf("base directory is required")
	}
	if opts.RepoName == "" {
		return fmt.Errorf("repository name is required")
	}

	repoPath := filepath.Join(opts.BaseDir, opts.RepoName)

	// Create directory
	if err := os.MkdirAll(repoPath, 0o755); err != nil {
		return fmt.Errorf("failed to create repository directory: %w", err)
	}

	// Initialize Git repository
	if err := f.runGitCommand(ctx, repoPath, "init"); err != nil {
		return fmt.Errorf("failed to initialize git repository: %w", err)
	}

	// Configure git user for testing
	if err := f.runGitCommand(ctx, repoPath, "config", "user.name", "Test User"); err != nil {
		return fmt.Errorf("failed to set git user name: %w", err)
	}
	if err := f.runGitCommand(ctx, repoPath, "config", "user.email", "test@example.com"); err != nil {
		return fmt.Errorf("failed to set git user email: %w", err)
	}

	// Add initial data if requested
	if opts.InitialData {
		readmePath := filepath.Join(repoPath, "README.md")
		content := fmt.Sprintf("# %s\n\nTest repository for synclone testing.\n", opts.RepoName)
		if err := os.WriteFile(readmePath, []byte(content), 0o644); err != nil {
			return fmt.Errorf("failed to create README.md: %w", err)
		}

		if err := f.runGitCommand(ctx, repoPath, "add", "README.md"); err != nil {
			return fmt.Errorf("failed to add README.md: %w", err)
		}

		if err := f.runGitCommand(ctx, repoPath, "commit", "-m", "Initial commit"); err != nil {
			return fmt.Errorf("failed to create initial commit: %w", err)
		}
	}

	// Create additional branches if specified
	for _, branch := range opts.Branches {
		if branch != "main" && branch != "master" {
			if err := f.runGitCommand(ctx, repoPath, "checkout", "-b", branch); err != nil {
				return fmt.Errorf("failed to create branch %s: %w", branch, err)
			}
			// Switch back to main branch
			if err := f.runGitCommand(ctx, repoPath, "checkout", "main"); err != nil {
				// Try master if main doesn't exist
				if err := f.runGitCommand(ctx, repoPath, "checkout", "master"); err != nil {
					return fmt.Errorf("failed to checkout main or master branch in %s: %w", repoPath, err)
				}
			}
		}
	}

	return nil
}

// CreateConflictRepos creates repositories with conflict scenarios.
func (f *DefaultMockRepoFactory) CreateConflictRepos(ctx context.Context, opts ConflictRepoOptions) error {
	// First create a basic repo
	basicOpts := BasicRepoOptions{
		BaseDir:     opts.BaseDir,
		RepoName:    opts.RepoName,
		InitialData: true,
	}

	if err := f.CreateBasicRepos(ctx, basicOpts); err != nil {
		return fmt.Errorf("failed to create basic repository: %w", err)
	}

	repoPath := filepath.Join(opts.BaseDir, opts.RepoName)

	switch opts.ConflictType {
	case "merge":
		return f.createMergeConflict(ctx, repoPath)
	case "rebase":
		return f.createRebaseConflict(ctx, repoPath)
	case "diverged":
		return f.createDivergedState(ctx, repoPath)
	default:
		return fmt.Errorf("unknown conflict type: %s", opts.ConflictType)
	}
}

// CreateSpecialRepos creates repositories with special scenarios.
func (f *DefaultMockRepoFactory) CreateSpecialRepos(ctx context.Context, opts SpecialRepoOptions) error {
	// Implementation will be added in Phase 1C
	return fmt.Errorf("special repositories not implemented yet")
}

// runGitCommand executes a git command with timeout.
func (f *DefaultMockRepoFactory) runGitCommand(ctx context.Context, dir string, args ...string) error {
	ctx, cancel := context.WithTimeout(ctx, f.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git %v failed: %w, output: %s", args, err, string(output))
	}

	return nil
}

// createMergeConflict creates a merge conflict scenario.
func (f *DefaultMockRepoFactory) createMergeConflict(ctx context.Context, repoPath string) error {
	// Create feature branch
	if err := f.runGitCommand(ctx, repoPath, "checkout", "-b", "feature"); err != nil {
		return fmt.Errorf("failed to create feature branch: %w", err)
	}

	// Modify file in feature branch
	conflictFile := filepath.Join(repoPath, "conflict.txt")
	if err := os.WriteFile(conflictFile, []byte("feature content\n"), 0o644); err != nil {
		return fmt.Errorf("failed to write conflict file: %w", err)
	}

	if err := f.runGitCommand(ctx, repoPath, "add", "conflict.txt"); err != nil {
		return fmt.Errorf("failed to add conflict file: %w", err)
	}

	if err := f.runGitCommand(ctx, repoPath, "commit", "-m", "Add feature content"); err != nil {
		return fmt.Errorf("failed to commit feature content: %w", err)
	}

	// Switch to main and modify same file
	if err := f.runGitCommand(ctx, repoPath, "checkout", "main"); err != nil {
		// Fallback to master branch
		if err := f.runGitCommand(ctx, repoPath, "checkout", "master"); err != nil {
			return fmt.Errorf("failed to checkout main or master branch: %w", err)
		}
	}

	if err := os.WriteFile(conflictFile, []byte("main content\n"), 0o644); err != nil {
		return fmt.Errorf("failed to write main content: %w", err)
	}

	if err := f.runGitCommand(ctx, repoPath, "add", "conflict.txt"); err != nil {
		return fmt.Errorf("failed to add main content: %w", err)
	}

	if err := f.runGitCommand(ctx, repoPath, "commit", "-m", "Add main content"); err != nil {
		return fmt.Errorf("failed to commit main content: %w", err)
	}

	return nil
}

// createRebaseConflict creates a rebase conflict scenario.
func (f *DefaultMockRepoFactory) createRebaseConflict(ctx context.Context, repoPath string) error {
	// Similar to merge conflict but with rebase scenario setup
	return f.createMergeConflict(ctx, repoPath)
}

// createDivergedState creates a diverged branch scenario.
func (f *DefaultMockRepoFactory) createDivergedState(ctx context.Context, repoPath string) error {
	// Create multiple commits on different branches
	if err := f.runGitCommand(ctx, repoPath, "checkout", "-b", "diverged"); err != nil {
		return fmt.Errorf("failed to create diverged branch: %w", err)
	}

	// Add commits to diverged branch
	for i := 1; i <= 3; i++ {
		fileName := fmt.Sprintf("diverged-%d.txt", i)
		filePath := filepath.Join(repoPath, fileName)
		content := fmt.Sprintf("Diverged commit %d\n", i)

		if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
			return fmt.Errorf("failed to write diverged file %d: %w", i, err)
		}

		if err := f.runGitCommand(ctx, repoPath, "add", fileName); err != nil {
			return fmt.Errorf("failed to add diverged file %d: %w", i, err)
		}

		if err := f.runGitCommand(ctx, repoPath, "commit", "-m", fmt.Sprintf("Diverged commit %d", i)); err != nil {
			return fmt.Errorf("failed to commit diverged file %d: %w", i, err)
		}
	}

	return nil
}
