// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package testlib

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// LocalChangesSimulator simulates various local working directory states
// for testing synclone behavior with uncommitted and staged changes.
type LocalChangesSimulator struct {
	timeout time.Duration
}

// WorkingDirectoryState represents different states of a working directory.
type WorkingDirectoryState struct {
	UncommittedChanges []LocalFileChange
	StagedChanges      []LocalFileChange
	UntrackedFiles     []LocalFileChange
	DeletedFiles       []string
	RenamedFiles       []FileRename
}

// LocalFileChange represents a change to a local file.
type LocalFileChange struct {
	FilePath     string
	Content      string
	ChangeType   ChangeType
	IsConflicted bool
}

// FileRename represents a file rename operation.
type FileRename struct {
	OldPath string
	NewPath string
	Content string
}

// ChangeType represents the type of file change.
type ChangeType int

const (
	ChangeTypeModify ChangeType = iota
	ChangeTypeAdd
	ChangeTypeDelete
	ChangeTypeRename
)

// String returns the string representation of ChangeType.
func (ct ChangeType) String() string {
	switch ct {
	case ChangeTypeModify:
		return "modify"
	case ChangeTypeAdd:
		return "add"
	case ChangeTypeDelete:
		return "delete"
	case ChangeTypeRename:
		return "rename" //nolint:goconst // string representation of enum
	default:
		return "unknown" //nolint:goconst // string representation of enum
	}
}

// NewLocalChangesSimulator creates a new LocalChangesSimulator instance.
func NewLocalChangesSimulator() *LocalChangesSimulator {
	return &LocalChangesSimulator{
		timeout: 30 * time.Second,
	}
}

// CreateWorkingDirectoryState creates a repository with the specified working directory state.
func (lcs *LocalChangesSimulator) CreateWorkingDirectoryState(ctx context.Context, repoPath string, state WorkingDirectoryState) error {
	// Ensure we have a Git repository
	if !lcs.isGitRepo(repoPath) {
		if err := lcs.initializeRepo(ctx, repoPath); err != nil {
			return fmt.Errorf("failed to initialize repository: %w", err)
		}
	}

	// Apply untracked files first
	for _, change := range state.UntrackedFiles {
		if err := lcs.applyFileChange(ctx, repoPath, change, false); err != nil {
			return fmt.Errorf("failed to create untracked file %s: %w", change.FilePath, err)
		}
	}

	// Apply file renames
	for _, rename := range state.RenamedFiles {
		if err := lcs.applyFileRename(ctx, repoPath, rename); err != nil {
			return fmt.Errorf("failed to rename file %s to %s: %w", rename.OldPath, rename.NewPath, err)
		}
	}

	// Apply staged changes
	for _, change := range state.StagedChanges {
		if err := lcs.applyFileChange(ctx, repoPath, change, true); err != nil {
			return fmt.Errorf("failed to create staged change %s: %w", change.FilePath, err)
		}
	}

	// Apply uncommitted changes
	for _, change := range state.UncommittedChanges {
		if err := lcs.applyFileChange(ctx, repoPath, change, false); err != nil {
			return fmt.Errorf("failed to create uncommitted change %s: %w", change.FilePath, err)
		}
	}

	// Delete specified files
	for _, filePath := range state.DeletedFiles {
		fullPath := filepath.Join(repoPath, filePath)
		if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to delete file %s: %w", filePath, err)
		}
	}

	return nil
}

// CreateDirtyWorkingDirectory creates a repository with various uncommitted changes.
func (lcs *LocalChangesSimulator) CreateDirtyWorkingDirectory(ctx context.Context, repoPath string) error {
	state := WorkingDirectoryState{
		UncommittedChanges: []LocalFileChange{
			{
				FilePath:   "modified-file.txt",
				Content:    "This file has been modified but not staged\n",
				ChangeType: ChangeTypeModify,
			},
			{
				FilePath:   "new-file.txt",
				Content:    "This is a new file\n",
				ChangeType: ChangeTypeAdd,
			},
		},
		StagedChanges: []LocalFileChange{
			{
				FilePath:   "staged-file.txt",
				Content:    "This file is staged for commit\n",
				ChangeType: ChangeTypeAdd,
			},
		},
		UntrackedFiles: []LocalFileChange{
			{
				FilePath:   "untracked.txt",
				Content:    "This file is untracked\n",
				ChangeType: ChangeTypeAdd,
			},
		},
	}

	return lcs.CreateWorkingDirectoryState(ctx, repoPath, state)
}

// CreateConflictedWorkingDirectory creates a repository with conflicted files.
func (lcs *LocalChangesSimulator) CreateConflictedWorkingDirectory(ctx context.Context, repoPath string) error {
	// First create a basic repository with some content
	if err := lcs.initializeRepo(ctx, repoPath); err != nil {
		return fmt.Errorf("failed to initialize repository: %w", err)
	}

	// Create initial file
	conflictFile := "conflict-file.txt"
	initialContent := "Line 1: Original content\nLine 2: Original content\nLine 3: Original content\n"

	if err := lcs.createAndCommitFile(ctx, repoPath, conflictFile, initialContent, "Initial commit"); err != nil {
		return fmt.Errorf("failed to create initial file: %w", err)
	}

	// Create a branch and make changes
	if err := lcs.runGitCommand(ctx, repoPath, "checkout", "-b", "branch1"); err != nil {
		return fmt.Errorf("failed to create branch1: %w", err)
	}

	branch1Content := "Line 1: Modified in branch1\nLine 2: Original content\nLine 3: Modified in branch1\n"
	if err := lcs.createAndCommitFile(ctx, repoPath, conflictFile, branch1Content, "Changes from branch1"); err != nil {
		return fmt.Errorf("failed to commit branch1 changes: %w", err)
	}

	// Switch back to main and make conflicting changes
	if err := lcs.runGitCommand(ctx, repoPath, "checkout", "main"); err != nil {
		// Fallback to master branch
		if err := lcs.runGitCommand(ctx, repoPath, "checkout", "master"); err != nil {
			return fmt.Errorf("failed to checkout main or master branch: %w", err)
		}
	}

	mainContent := "Line 1: Modified in main\nLine 2: Modified in main\nLine 3: Original content\n"
	if err := lcs.createAndCommitFile(ctx, repoPath, conflictFile, mainContent, "Changes from main"); err != nil {
		return fmt.Errorf("failed to commit main changes: %w", err)
	}

	// Try to merge branch1 (this will create conflicts)
	err := lcs.runGitCommand(ctx, repoPath, "merge", "branch1")
	if err != nil {
		// Expected to fail with conflicts, check if it's actually a conflict
		if !strings.Contains(err.Error(), "conflict") && !strings.Contains(err.Error(), "merge") {
			return fmt.Errorf("unexpected error during merge: %w", err)
		}
		// Conflict created successfully
	}

	return nil
}

// CreateMixedStateRepository creates a repository with a combination of different states.
func (lcs *LocalChangesSimulator) CreateMixedStateRepository(ctx context.Context, repoPath string) error {
	state := WorkingDirectoryState{
		UncommittedChanges: []LocalFileChange{
			{
				FilePath:   "work-in-progress.go",
				Content:    "package main\n\n// TODO: implement this function\nfunc main() {\n\t// work in progress\n}\n",
				ChangeType: ChangeTypeModify,
			},
		},
		StagedChanges: []LocalFileChange{
			{
				FilePath:   "feature.go",
				Content:    "package main\n\n// New feature implementation\nfunc NewFeature() {\n\t// implemented\n}\n",
				ChangeType: ChangeTypeAdd,
			},
		},
		UntrackedFiles: []LocalFileChange{
			{
				FilePath:   "temp.log",
				Content:    "Temporary log file\n",
				ChangeType: ChangeTypeAdd,
			},
			{
				FilePath:   ".DS_Store",
				Content:    "Mac system file\n",
				ChangeType: ChangeTypeAdd,
			},
		},
		RenamedFiles: []FileRename{
			{
				OldPath: "old-config.yaml",
				NewPath: "config.yaml",
				Content: "version: 2.0\nconfig: updated\n",
			},
		},
		DeletedFiles: []string{"obsolete-file.txt"},
	}

	return lcs.CreateWorkingDirectoryState(ctx, repoPath, state)
}

// GetWorkingDirectoryStatus returns the current status of the working directory.
func (lcs *LocalChangesSimulator) GetWorkingDirectoryStatus(ctx context.Context, repoPath string) (WorkingDirectoryStatus, error) {
	cmd := exec.CommandContext(ctx, "git", "status", "--porcelain")
	cmd.Dir = repoPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return WorkingDirectoryStatus{}, fmt.Errorf("failed to get git status: %w", err)
	}

	return lcs.parseGitStatus(string(output)), nil
}

// WorkingDirectoryStatus represents the parsed status of a working directory.
type WorkingDirectoryStatus struct {
	ModifiedFiles   []string
	AddedFiles      []string
	DeletedFiles    []string
	RenamedFiles    []string
	UntrackedFiles  []string
	ConflictedFiles []string
}

// IsClean returns true if the working directory has no changes.
func (wds WorkingDirectoryStatus) IsClean() bool {
	return len(wds.ModifiedFiles) == 0 &&
		len(wds.AddedFiles) == 0 &&
		len(wds.DeletedFiles) == 0 &&
		len(wds.RenamedFiles) == 0 &&
		len(wds.UntrackedFiles) == 0 &&
		len(wds.ConflictedFiles) == 0
}

// HasUncommittedChanges returns true if there are uncommitted changes.
func (wds WorkingDirectoryStatus) HasUncommittedChanges() bool {
	return len(wds.ModifiedFiles) > 0 ||
		len(wds.AddedFiles) > 0 ||
		len(wds.DeletedFiles) > 0 ||
		len(wds.RenamedFiles) > 0
}

// HasConflicts returns true if there are conflicted files.
func (wds WorkingDirectoryStatus) HasConflicts() bool {
	return len(wds.ConflictedFiles) > 0
}

// parseGitStatus parses the output of `git status --porcelain`.
func (lcs *LocalChangesSimulator) parseGitStatus(output string) WorkingDirectoryStatus {
	status := WorkingDirectoryStatus{}
	lines := strings.Split(strings.TrimSpace(output), "\n")

	for _, line := range lines {
		if len(line) < 3 {
			continue
		}

		indexStatus := line[0]
		workTreeStatus := line[1]
		filename := line[3:]

		switch {
		case indexStatus == 'U' || workTreeStatus == 'U':
			status.ConflictedFiles = append(status.ConflictedFiles, filename)
		case workTreeStatus == 'M':
			status.ModifiedFiles = append(status.ModifiedFiles, filename)
		case workTreeStatus == 'D':
			status.DeletedFiles = append(status.DeletedFiles, filename)
		case indexStatus == 'A':
			status.AddedFiles = append(status.AddedFiles, filename)
		case indexStatus == 'R':
			status.RenamedFiles = append(status.RenamedFiles, filename)
		case indexStatus == '?' && workTreeStatus == '?':
			status.UntrackedFiles = append(status.UntrackedFiles, filename)
		}
	}

	return status
}

// applyFileChange applies a single file change to the repository.
func (lcs *LocalChangesSimulator) applyFileChange(ctx context.Context, repoPath string, change LocalFileChange, staged bool) error {
	filePath := filepath.Join(repoPath, change.FilePath)

	// Create directory if needed
	if dir := filepath.Dir(filePath); dir != repoPath {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	switch change.ChangeType {
	case ChangeTypeAdd, ChangeTypeModify:
		if err := os.WriteFile(filePath, []byte(change.Content), 0o644); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}

		if staged {
			if err := lcs.runGitCommand(ctx, repoPath, "add", change.FilePath); err != nil {
				return fmt.Errorf("failed to stage file: %w", err)
			}
		}

	case ChangeTypeDelete:
		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to delete file: %w", err)
		}

		if staged {
			if err := lcs.runGitCommand(ctx, repoPath, "rm", change.FilePath); err != nil {
				return fmt.Errorf("failed to stage deletion: %w", err)
			}
		}

	case ChangeTypeRename:
		// Rename is typically handled separately through FileRename struct
		// This case exists for exhaustive switch checking
		return fmt.Errorf("rename operations should use FileRename struct")
	}

	return nil
}

// applyFileRename applies a file rename operation.
func (lcs *LocalChangesSimulator) applyFileRename(ctx context.Context, repoPath string, rename FileRename) error {
	oldPath := filepath.Join(repoPath, rename.OldPath)
	newPath := filepath.Join(repoPath, rename.NewPath)

	// Create the old file first if it doesn't exist
	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		if err := os.WriteFile(oldPath, []byte("original content\n"), 0o644); err != nil {
			return fmt.Errorf("failed to create original file: %w", err)
		}
		if err := lcs.runGitCommand(ctx, repoPath, "add", rename.OldPath); err != nil {
			return fmt.Errorf("failed to add original file: %w", err)
		}
		if err := lcs.runGitCommand(ctx, repoPath, "commit", "-m", "Add original file for rename"); err != nil {
			return fmt.Errorf("failed to commit original file: %w", err)
		}
	}

	// Create the new file with updated content
	if err := os.WriteFile(newPath, []byte(rename.Content), 0o644); err != nil {
		return fmt.Errorf("failed to create new file: %w", err)
	}

	// Remove the old file
	if err := os.Remove(oldPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove old file: %w", err)
	}

	// Stage both operations
	if err := lcs.runGitCommand(ctx, repoPath, "rm", rename.OldPath); err != nil {
		return fmt.Errorf("failed to stage file removal: %w", err)
	}
	if err := lcs.runGitCommand(ctx, repoPath, "add", rename.NewPath); err != nil {
		return fmt.Errorf("failed to stage new file: %w", err)
	}

	return nil
}

// createAndCommitFile creates a file and commits it.
func (lcs *LocalChangesSimulator) createAndCommitFile(ctx context.Context, repoPath, filePath, content, commitMessage string) error {
	fullPath := filepath.Join(repoPath, filePath)

	if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	if err := lcs.runGitCommand(ctx, repoPath, "add", filePath); err != nil {
		return fmt.Errorf("failed to add file: %w", err)
	}

	if err := lcs.runGitCommand(ctx, repoPath, "commit", "-m", commitMessage); err != nil {
		return fmt.Errorf("failed to commit file: %w", err)
	}

	return nil
}

// isGitRepo checks if a directory is a Git repository.
func (lcs *LocalChangesSimulator) isGitRepo(repoPath string) bool {
	gitDir := filepath.Join(repoPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return false
	}
	return true
}

// initializeRepo initializes a Git repository with an initial commit.
func (lcs *LocalChangesSimulator) initializeRepo(ctx context.Context, repoPath string) error {
	if err := os.MkdirAll(repoPath, 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := lcs.runGitCommand(ctx, repoPath, "init"); err != nil {
		return fmt.Errorf("failed to initialize git repository: %w", err)
	}

	// Configure git user
	if err := lcs.runGitCommand(ctx, repoPath, "config", "user.name", "Test User"); err != nil {
		return fmt.Errorf("failed to set git user name: %w", err)
	}
	if err := lcs.runGitCommand(ctx, repoPath, "config", "user.email", "test@example.com"); err != nil {
		return fmt.Errorf("failed to set git user email: %w", err)
	}

	// Create initial commit
	readmePath := filepath.Join(repoPath, "README.md")
	content := fmt.Sprintf("# %s\n\nTest repository for local changes simulation.\n", filepath.Base(repoPath))
	if err := os.WriteFile(readmePath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("failed to create README.md: %w", err)
	}

	if err := lcs.runGitCommand(ctx, repoPath, "add", "README.md"); err != nil {
		return fmt.Errorf("failed to add README.md: %w", err)
	}

	if err := lcs.runGitCommand(ctx, repoPath, "commit", "-m", "Initial commit"); err != nil {
		return fmt.Errorf("failed to create initial commit: %w", err)
	}

	// Create some files that can be modified
	existingFiles := map[string]string{
		"modified-file.txt": "Original content\nThis file will be modified\n",
		"existing.txt":      "Existing file content\n",
		"old-config.yaml":   "version: 1.0\nconfig: original\n",
		"obsolete-file.txt": "This file will be deleted\n",
	}

	for filename, content := range existingFiles {
		if err := lcs.createAndCommitFile(ctx, repoPath, filename, content, fmt.Sprintf("Add %s", filename)); err != nil {
			return fmt.Errorf("failed to create existing file %s: %w", filename, err)
		}
	}

	return nil
}

// runGitCommand executes a git command with timeout.
func (lcs *LocalChangesSimulator) runGitCommand(ctx context.Context, dir string, args ...string) error {
	ctx, cancel := context.WithTimeout(ctx, lcs.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git %v failed: %w, output: %s", args, err, string(output))
	}

	return nil
}
