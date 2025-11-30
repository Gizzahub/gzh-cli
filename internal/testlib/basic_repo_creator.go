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

// BasicRepoCreator provides functionality for creating basic Git repositories
// with various initial states for testing synclone operations.
type BasicRepoCreator struct {
	timeout time.Duration
}

// NewBasicRepoCreator creates a new BasicRepoCreator instance.
func NewBasicRepoCreator() *BasicRepoCreator {
	return &BasicRepoCreator{
		timeout: 30 * time.Second,
	}
}

// CreateEmptyRepo creates a Git repository with just git init
// This simulates the most basic repository state.
func (c *BasicRepoCreator) CreateEmptyRepo(ctx context.Context, repoPath string) error {
	if repoPath == "" {
		return fmt.Errorf("repository path is required")
	}

	// Create directory
	if err := os.MkdirAll(repoPath, 0o755); err != nil {
		return fmt.Errorf("failed to create repository directory: %w", err)
	}

	// Initialize Git repository
	if err := c.runGitCommand(ctx, repoPath, "init"); err != nil {
		return fmt.Errorf("failed to initialize git repository: %w", err)
	}

	// Configure git user for testing
	if err := c.configureGitUser(ctx, repoPath); err != nil {
		return fmt.Errorf("failed to configure git user: %w", err)
	}

	return nil
}

// CreateMinimalRepo creates a repository with a single initial commit
// This simulates the most common starting state for repositories.
func (c *BasicRepoCreator) CreateMinimalRepo(ctx context.Context, repoPath string) error {
	if err := c.CreateEmptyRepo(ctx, repoPath); err != nil {
		return fmt.Errorf("failed to create empty repository: %w", err)
	}

	// Add a README file
	readmePath := filepath.Join(repoPath, "README.md")
	content := fmt.Sprintf("# %s\n\nMinimal test repository for synclone testing.\n", filepath.Base(repoPath))
	if err := os.WriteFile(readmePath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("failed to create README.md: %w", err)
	}

	// Stage and commit the file
	if err := c.runGitCommand(ctx, repoPath, "add", "README.md"); err != nil {
		return fmt.Errorf("failed to add README.md: %w", err)
	}

	if err := c.runGitCommand(ctx, repoPath, "commit", "-m", "Initial commit"); err != nil {
		return fmt.Errorf("failed to create initial commit: %w", err)
	}

	return nil
}

// CreateRepoWithFiles creates a repository with multiple files
// This simulates repositories with actual content.
func (c *BasicRepoCreator) CreateRepoWithFiles(ctx context.Context, repoPath string, files map[string]string) error {
	if err := c.CreateEmptyRepo(ctx, repoPath); err != nil {
		return fmt.Errorf("failed to create empty repository: %w", err)
	}

	// Create and add files
	for filename, content := range files {
		filePath := filepath.Join(repoPath, filename)

		// Create directory if needed
		if dir := filepath.Dir(filePath); dir != "." {
			if err := os.MkdirAll(filepath.Join(repoPath, dir), 0o755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", dir, err)
			}
		}

		if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
			return fmt.Errorf("failed to create file %s: %w", filename, err)
		}

		if err := c.runGitCommand(ctx, repoPath, "add", filename); err != nil {
			return fmt.Errorf("failed to add file %s: %w", filename, err)
		}
	}

	// Commit all files
	if err := c.runGitCommand(ctx, repoPath, "commit", "-m", "Add initial files"); err != nil {
		return fmt.Errorf("failed to commit files: %w", err)
	}

	return nil
}

// configureGitUser sets up git user configuration for testing.
func (c *BasicRepoCreator) configureGitUser(ctx context.Context, repoPath string) error {
	if err := c.runGitCommand(ctx, repoPath, "config", "user.name", "Test User"); err != nil {
		return fmt.Errorf("failed to set git user name: %w", err)
	}
	if err := c.runGitCommand(ctx, repoPath, "config", "user.email", "test@example.com"); err != nil {
		return fmt.Errorf("failed to set git user email: %w", err)
	}
	return nil
}

// runGitCommand executes a git command with timeout.
func (c *BasicRepoCreator) runGitCommand(ctx context.Context, dir string, args ...string) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git %v failed: %w, output: %s", args, err, string(output))
	}

	return nil
}
