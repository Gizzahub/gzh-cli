// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package git

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GitUtils provides Git-related utilities for quality processing
type GitUtils struct {
	projectRoot string
}

// NewGitUtils creates a new GitUtils instance
func NewGitUtils(projectRoot string) *GitUtils {
	return &GitUtils{
		projectRoot: projectRoot,
	}
}

// IsGitRepository checks if the current directory is a Git repository
func (g *GitUtils) IsGitRepository() bool {
	gitDir := filepath.Join(g.projectRoot, ".git")
	if stat, err := os.Stat(gitDir); err == nil {
		return stat.IsDir()
	}

	// Check if it's a git worktree or submodule
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = g.projectRoot
	return cmd.Run() == nil
}

// GetChangedFiles returns files changed since a specific commit
func (g *GitUtils) GetChangedFiles(since string) ([]string, error) {
	if !g.IsGitRepository() {
		return nil, fmt.Errorf("not a git repository")
	}

	var cmd *exec.Cmd
	if since == "" {
		// Default to comparing with HEAD~1
		since = "HEAD~1"
	}

	// Get changed files since the specified commit
	cmd = exec.Command("git", "diff", "--name-only", since)
	cmd.Dir = g.projectRoot

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get git diff: %w", err)
	}

	return g.parseFileList(string(output)), nil
}

// GetStagedFiles returns currently staged files
func (g *GitUtils) GetStagedFiles() ([]string, error) {
	if !g.IsGitRepository() {
		return nil, fmt.Errorf("not a git repository")
	}

	cmd := exec.Command("git", "diff", "--cached", "--name-only")
	cmd.Dir = g.projectRoot

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get staged files: %w", err)
	}

	return g.parseFileList(string(output)), nil
}

// GetModifiedFiles returns modified files in working directory
func (g *GitUtils) GetModifiedFiles() ([]string, error) {
	if !g.IsGitRepository() {
		return nil, fmt.Errorf("not a git repository")
	}

	cmd := exec.Command("git", "diff", "--name-only")
	cmd.Dir = g.projectRoot

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get modified files: %w", err)
	}

	return g.parseFileList(string(output)), nil
}

// GetUntrackedFiles returns untracked files
func (g *GitUtils) GetUntrackedFiles() ([]string, error) {
	if !g.IsGitRepository() {
		return nil, fmt.Errorf("not a git repository")
	}

	cmd := exec.Command("git", "ls-files", "--others", "--exclude-standard")
	cmd.Dir = g.projectRoot

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get untracked files: %w", err)
	}

	return g.parseFileList(string(output)), nil
}

// GetAllChangedFiles returns all changed files (staged + modified + untracked)
func (g *GitUtils) GetAllChangedFiles() ([]string, error) {
	var allFiles []string

	// Get staged files
	staged, err := g.GetStagedFiles()
	if err != nil {
		return nil, err
	}
	allFiles = append(allFiles, staged...)

	// Get modified files
	modified, err := g.GetModifiedFiles()
	if err != nil {
		return nil, err
	}
	allFiles = append(allFiles, modified...)

	// Get untracked files
	untracked, err := g.GetUntrackedFiles()
	if err != nil {
		return nil, err
	}
	allFiles = append(allFiles, untracked...)

	// Remove duplicates and return absolute paths
	return g.deduplicateAndMakeAbsolute(allFiles), nil
}

// GetCurrentBranch returns the current Git branch name
func (g *GitUtils) GetCurrentBranch() (string, error) {
	if !g.IsGitRepository() {
		return "", fmt.Errorf("not a git repository")
	}

	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = g.projectRoot

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// parseFileList parses git command output into file list
func (g *GitUtils) parseFileList(output string) []string {
	var files []string
	scanner := bufio.NewScanner(strings.NewReader(output))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			files = append(files, line)
		}
	}

	return files
}

// deduplicateAndMakeAbsolute removes duplicates and converts to absolute paths
func (g *GitUtils) deduplicateAndMakeAbsolute(files []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, file := range files {
		if seen[file] {
			continue
		}
		seen[file] = true

		// Convert to absolute path
		absPath := filepath.Join(g.projectRoot, file)
		if _, err := os.Stat(absPath); err == nil {
			result = append(result, absPath)
		}
	}

	return result
}

// ValidateCommitish checks if a commit-ish reference is valid
func (g *GitUtils) ValidateCommitish(commitish string) error {
	if !g.IsGitRepository() {
		return fmt.Errorf("not a git repository")
	}

	cmd := exec.Command("git", "rev-parse", "--verify", commitish+"^{commit}")
	cmd.Dir = g.projectRoot

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("invalid commit reference '%s': %w", commitish, err)
	}

	return nil
}
