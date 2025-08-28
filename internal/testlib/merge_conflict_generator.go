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

// MergeConflictGenerator creates repositories with various merge conflict scenarios
// for testing synclone behavior in complex Git states.
type MergeConflictGenerator struct {
	timeout time.Duration
}

// ConflictScenario defines a merge conflict scenario
type ConflictScenario struct {
	Name              string
	Description       string
	BaseBranch        string
	ConflictingBranch string
	ConflictFiles     []ConflictingFile
	PreMergeSetup     []PreMergeAction
}

// ConflictingFile represents a file that will have conflicts
type ConflictingFile struct {
	FilePath       string
	BaseContent    string
	BranchAContent string
	BranchBContent string
	ConflictType   ConflictType
}

// PreMergeAction represents actions to take before creating conflicts
type PreMergeAction struct {
	Branch    string
	Action    string // "commit", "modify", "delete", "rename"
	FilePath  string
	Content   string
	CommitMsg string
}

// ConflictType represents the type of merge conflict
type ConflictType int

const (
	ConflictTypeContent      ConflictType = iota // Both sides modified same lines
	ConflictTypeAddAdd                           // Both sides added same file
	ConflictTypeModifyDelete                     // One modified, one deleted
	ConflictTypeRename                           // Rename conflicts
)

// String returns the string representation of ConflictType
func (ct ConflictType) String() string {
	switch ct {
	case ConflictTypeContent:
		return "content"
	case ConflictTypeAddAdd:
		return "add-add"
	case ConflictTypeModifyDelete:
		return "modify-delete"
	case ConflictTypeRename:
		return "rename"
	default:
		return "unknown"
	}
}

// NewMergeConflictGenerator creates a new MergeConflictGenerator instance
func NewMergeConflictGenerator() *MergeConflictGenerator {
	return &MergeConflictGenerator{
		timeout: 60 * time.Second,
	}
}

// CreateConflictScenario creates a repository with the specified conflict scenario
func (mcg *MergeConflictGenerator) CreateConflictScenario(ctx context.Context, repoPath string, scenario ConflictScenario) error {
	// Initialize repository if needed
	if !mcg.isGitRepo(repoPath) {
		if err := mcg.initializeRepo(ctx, repoPath); err != nil {
			return fmt.Errorf("failed to initialize repository: %w", err)
		}
	}

	// Create base content for all conflict files
	for _, file := range scenario.ConflictFiles {
		if err := mcg.createBaseFile(ctx, repoPath, file); err != nil {
			return fmt.Errorf("failed to create base file %s: %w", file.FilePath, err)
		}
	}

	// Execute pre-merge actions
	for _, action := range scenario.PreMergeSetup {
		if err := mcg.executePreMergeAction(ctx, repoPath, action); err != nil {
			return fmt.Errorf("failed to execute pre-merge action: %w", err)
		}
	}

	// Create conflicting branches and content
	if err := mcg.createConflictingBranches(ctx, repoPath, scenario); err != nil {
		return fmt.Errorf("failed to create conflicting branches: %w", err)
	}

	// Attempt merge to create conflicts
	if err := mcg.createMergeConflicts(ctx, repoPath, scenario); err != nil {
		return fmt.Errorf("failed to create merge conflicts: %w", err)
	}

	return nil
}

// CreateSimpleContentConflict creates a basic content conflict scenario
func (mcg *MergeConflictGenerator) CreateSimpleContentConflict(ctx context.Context, repoPath string) error {
	scenario := ConflictScenario{
		Name:              "Simple Content Conflict",
		Description:       "Two branches modify the same lines in a file",
		BaseBranch:        "main",
		ConflictingBranch: "feature",
		ConflictFiles: []ConflictingFile{
			{
				FilePath:       "shared.txt",
				BaseContent:    "Line 1: Original\nLine 2: Original\nLine 3: Original\n",
				BranchAContent: "Line 1: Modified in main\nLine 2: Original\nLine 3: Modified in main\n",
				BranchBContent: "Line 1: Original\nLine 2: Modified in feature\nLine 3: Modified in feature\n",
				ConflictType:   ConflictTypeContent,
			},
		},
	}

	return mcg.CreateConflictScenario(ctx, repoPath, scenario)
}

// CreateComplexConflict creates a complex conflict scenario with multiple types
func (mcg *MergeConflictGenerator) CreateComplexConflict(ctx context.Context, repoPath string) error {
	scenario := ConflictScenario{
		Name:              "Complex Multi-type Conflict",
		Description:       "Multiple conflict types in the same repository",
		BaseBranch:        "main",
		ConflictingBranch: "develop",
		ConflictFiles: []ConflictingFile{
			{
				FilePath:       "config.json",
				BaseContent:    `{"version": "1.0", "name": "app", "debug": false}`,
				BranchAContent: `{"version": "1.1", "name": "app", "debug": false, "feature_a": true}`,
				BranchBContent: `{"version": "1.0", "name": "myapp", "debug": true, "feature_b": true}`,
				ConflictType:   ConflictTypeContent,
			},
			{
				FilePath:       "new-feature.go",
				BaseContent:    "", // File doesn't exist in base
				BranchAContent: "package main\n\n// Feature A implementation\nfunc FeatureA() {}\n",
				BranchBContent: "package main\n\n// Feature B implementation\nfunc FeatureB() {}\n",
				ConflictType:   ConflictTypeAddAdd,
			},
		},
		PreMergeSetup: []PreMergeAction{
			{
				Branch:    "main",
				Action:    "modify",
				FilePath:  "docs.md",
				Content:   "# Documentation\n\nUpdated from main branch\n",
				CommitMsg: "Update documentation in main",
			},
			{
				Branch:    "develop",
				Action:    "delete",
				FilePath:  "docs.md",
				CommitMsg: "Remove old documentation",
			},
		},
	}

	return mcg.CreateConflictScenario(ctx, repoPath, scenario)
}

// CreateDivergedBranchesConflict creates a scenario where branches have diverged significantly
func (mcg *MergeConflictGenerator) CreateDivergedBranchesConflict(ctx context.Context, repoPath string) error {
	if err := mcg.initializeRepo(ctx, repoPath); err != nil {
		return fmt.Errorf("failed to initialize repository: %w", err)
	}

	// Create multiple commits on main
	mainCommits := []struct {
		file    string
		content string
		message string
	}{
		{"main-1.txt", "Main branch commit 1\n", "Main: Add feature 1"},
		{"main-2.txt", "Main branch commit 2\n", "Main: Add feature 2"},
		{"shared.txt", "Shared file modified in main\n", "Main: Update shared file"},
	}

	for _, commit := range mainCommits {
		if err := mcg.createAndCommitFile(ctx, repoPath, commit.file, commit.content, commit.message); err != nil {
			return fmt.Errorf("failed to create main commit: %w", err)
		}
	}

	// Create feature branch from an earlier commit
	if err := mcg.runGitCommand(ctx, repoPath, "checkout", "HEAD~2"); err != nil {
		return fmt.Errorf("failed to checkout earlier commit: %w", err)
	}

	if err := mcg.runGitCommand(ctx, repoPath, "checkout", "-b", "feature"); err != nil {
		return fmt.Errorf("failed to create feature branch: %w", err)
	}

	// Create multiple commits on feature branch
	featureCommits := []struct {
		file    string
		content string
		message string
	}{
		{"feature-1.txt", "Feature branch commit 1\n", "Feature: Add component 1"},
		{"feature-2.txt", "Feature branch commit 2\n", "Feature: Add component 2"},
		{"feature-3.txt", "Feature branch commit 3\n", "Feature: Add component 3"},
		{"shared.txt", "Shared file modified in feature\nWith additional content\n", "Feature: Update shared file"},
	}

	for _, commit := range featureCommits {
		if err := mcg.createAndCommitFile(ctx, repoPath, commit.file, commit.content, commit.message); err != nil {
			return fmt.Errorf("failed to create feature commit: %w", err)
		}
	}

	// Switch back to main and attempt merge
	if err := mcg.runGitCommand(ctx, repoPath, "checkout", "main"); err != nil {
		return fmt.Errorf("failed to checkout main: %w", err)
	}

	// This merge should create conflicts
	err := mcg.runGitCommand(ctx, repoPath, "merge", "feature")
	if err != nil && (strings.Contains(err.Error(), "conflict") || strings.Contains(err.Error(), "CONFLICT")) {
		// Expected conflict, this is success
		return nil
	} else if err != nil {
		return fmt.Errorf("unexpected error during merge: %w", err)
	}

	// If no conflict occurred, create an explicit conflict
	return mcg.CreateSimpleContentConflict(ctx, repoPath+"-fallback")
}

// GetConflictStatus returns information about current merge conflicts
func (mcg *MergeConflictGenerator) GetConflictStatus(ctx context.Context, repoPath string) (ConflictStatus, error) {
	cmd := exec.CommandContext(ctx, "git", "status", "--porcelain")
	cmd.Dir = repoPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return ConflictStatus{}, fmt.Errorf("failed to get conflict status: %w", err)
	}

	return mcg.parseConflictStatus(string(output)), nil
}

// ConflictStatus represents the current conflict state
type ConflictStatus struct {
	HasConflicts    bool
	ConflictedFiles []string
	UnmergedPaths   []string
	BothModified    []string
	BothAdded       []string
	DeletedByThem   []string
	DeletedByUs     []string
}

// IsInMergeState returns true if the repository is in a merge state
func (cs ConflictStatus) IsInMergeState() bool {
	return cs.HasConflicts || len(cs.UnmergedPaths) > 0
}

// parseConflictStatus parses git status output to identify conflicts
func (mcg *MergeConflictGenerator) parseConflictStatus(output string) ConflictStatus {
	status := ConflictStatus{}
	lines := strings.Split(strings.TrimSpace(output), "\n")

	for _, line := range lines {
		if len(line) < 3 {
			continue
		}

		indexStatus := line[0]
		workTreeStatus := line[1]
		filename := line[3:]

		switch {
		case indexStatus == 'U' && workTreeStatus == 'U':
			status.ConflictedFiles = append(status.ConflictedFiles, filename)
			status.BothModified = append(status.BothModified, filename)
			status.HasConflicts = true
		case indexStatus == 'A' && workTreeStatus == 'A':
			status.ConflictedFiles = append(status.ConflictedFiles, filename)
			status.BothAdded = append(status.BothAdded, filename)
			status.HasConflicts = true
		case indexStatus == 'D' && workTreeStatus == 'U':
			status.ConflictedFiles = append(status.ConflictedFiles, filename)
			status.DeletedByUs = append(status.DeletedByUs, filename)
			status.HasConflicts = true
		case indexStatus == 'U' && workTreeStatus == 'D':
			status.ConflictedFiles = append(status.ConflictedFiles, filename)
			status.DeletedByThem = append(status.DeletedByThem, filename)
			status.HasConflicts = true
		}

		if indexStatus == 'U' || workTreeStatus == 'U' {
			status.UnmergedPaths = append(status.UnmergedPaths, filename)
		}
	}

	return status
}

// createBaseFile creates the initial version of a file that will have conflicts
func (mcg *MergeConflictGenerator) createBaseFile(ctx context.Context, repoPath string, file ConflictingFile) error {
	if file.BaseContent == "" {
		return nil // File doesn't exist in base, skip
	}

	filePath := filepath.Join(repoPath, file.FilePath)

	// Create directory if needed
	if dir := filepath.Dir(filePath); dir != repoPath {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	if err := os.WriteFile(filePath, []byte(file.BaseContent), 0o644); err != nil {
		return fmt.Errorf("failed to write base file: %w", err)
	}

	if err := mcg.runGitCommand(ctx, repoPath, "add", file.FilePath); err != nil {
		return fmt.Errorf("failed to add base file: %w", err)
	}

	if err := mcg.runGitCommand(ctx, repoPath, "commit", "-m", fmt.Sprintf("Add base version of %s", file.FilePath)); err != nil {
		return fmt.Errorf("failed to commit base file: %w", err)
	}

	return nil
}

// executePreMergeAction executes a pre-merge setup action
func (mcg *MergeConflictGenerator) executePreMergeAction(ctx context.Context, repoPath string, action PreMergeAction) error {
	// Switch to the specified branch
	if err := mcg.runGitCommand(ctx, repoPath, "checkout", action.Branch); err != nil {
		return fmt.Errorf("failed to checkout branch %s: %w", action.Branch, err)
	}

	switch action.Action {
	case "commit":
		// Just commit current state
		return mcg.runGitCommand(ctx, repoPath, "commit", "-m", action.CommitMsg)

	case "modify":
		if err := mcg.createAndCommitFile(ctx, repoPath, action.FilePath, action.Content, action.CommitMsg); err != nil {
			return fmt.Errorf("failed to modify file: %w", err)
		}

	case "delete":
		filePath := filepath.Join(repoPath, action.FilePath)
		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to delete file: %w", err)
		}

		if err := mcg.runGitCommand(ctx, repoPath, "add", action.FilePath); err != nil {
			return fmt.Errorf("failed to stage deletion: %w", err)
		}

		if err := mcg.runGitCommand(ctx, repoPath, "commit", "-m", action.CommitMsg); err != nil {
			return fmt.Errorf("failed to commit deletion: %w", err)
		}

	case "rename":
		// Implementation for rename would go here
		return fmt.Errorf("rename action not implemented yet")
	}

	return nil
}

// createConflictingBranches creates branches with conflicting content
func (mcg *MergeConflictGenerator) createConflictingBranches(ctx context.Context, repoPath string, scenario ConflictScenario) error {
	// Ensure we're on the base branch
	if err := mcg.runGitCommand(ctx, repoPath, "checkout", scenario.BaseBranch); err != nil {
		return fmt.Errorf("failed to checkout base branch: %w", err)
	}

	// Modify files on base branch
	for _, file := range scenario.ConflictFiles {
		if file.BranchAContent != "" && file.BranchAContent != file.BaseContent {
			if err := mcg.createAndCommitFile(ctx, repoPath, file.FilePath, file.BranchAContent,
				fmt.Sprintf("Modify %s in %s", file.FilePath, scenario.BaseBranch)); err != nil {
				return fmt.Errorf("failed to modify file in base branch: %w", err)
			}
		}
	}

	// Create conflicting branch
	if err := mcg.runGitCommand(ctx, repoPath, "checkout", "-b", scenario.ConflictingBranch); err != nil {
		return fmt.Errorf("failed to create conflicting branch: %w", err)
	}

	// Reset to before the base branch changes
	if err := mcg.runGitCommand(ctx, repoPath, "reset", "--hard", "HEAD~1"); err != nil {
		// If this fails, continue anyway
	}

	// Modify files on conflicting branch
	for _, file := range scenario.ConflictFiles {
		if file.BranchBContent != "" {
			if err := mcg.createAndCommitFile(ctx, repoPath, file.FilePath, file.BranchBContent,
				fmt.Sprintf("Modify %s in %s", file.FilePath, scenario.ConflictingBranch)); err != nil {
				return fmt.Errorf("failed to modify file in conflicting branch: %w", err)
			}
		}
	}

	return nil
}

// createMergeConflicts attempts to merge branches to create conflicts
func (mcg *MergeConflictGenerator) createMergeConflicts(ctx context.Context, repoPath string, scenario ConflictScenario) error {
	// Switch to base branch
	if err := mcg.runGitCommand(ctx, repoPath, "checkout", scenario.BaseBranch); err != nil {
		return fmt.Errorf("failed to checkout base branch: %w", err)
	}

	// Attempt merge (this should create conflicts)
	err := mcg.runGitCommand(ctx, repoPath, "merge", scenario.ConflictingBranch)
	if err != nil {
		// Check if it's a merge conflict (expected)
		if strings.Contains(err.Error(), "conflict") ||
			strings.Contains(err.Error(), "CONFLICT") ||
			strings.Contains(err.Error(), "merge") {
			// This is the expected result - conflicts were created
			return nil
		}
		return fmt.Errorf("unexpected merge error: %w", err)
	}

	// If merge succeeded without conflicts, this might be unexpected
	// but we'll allow it for scenarios where conflicts aren't guaranteed
	return nil
}

// Helper methods

// createAndCommitFile creates a file and commits it
func (mcg *MergeConflictGenerator) createAndCommitFile(ctx context.Context, repoPath, filePath, content, commitMessage string) error {
	fullPath := filepath.Join(repoPath, filePath)

	// Create directory if needed
	if dir := filepath.Dir(fullPath); dir != repoPath {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	if err := mcg.runGitCommand(ctx, repoPath, "add", filePath); err != nil {
		return fmt.Errorf("failed to add file: %w", err)
	}

	if err := mcg.runGitCommand(ctx, repoPath, "commit", "-m", commitMessage); err != nil {
		return fmt.Errorf("failed to commit file: %w", err)
	}

	return nil
}

// isGitRepo checks if a directory is a Git repository
func (mcg *MergeConflictGenerator) isGitRepo(repoPath string) bool {
	gitDir := filepath.Join(repoPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return false
	}
	return true
}

// initializeRepo initializes a Git repository
func (mcg *MergeConflictGenerator) initializeRepo(ctx context.Context, repoPath string) error {
	if err := os.MkdirAll(repoPath, 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := mcg.runGitCommand(ctx, repoPath, "init"); err != nil {
		return fmt.Errorf("failed to initialize git repository: %w", err)
	}

	// Configure git user
	if err := mcg.runGitCommand(ctx, repoPath, "config", "user.name", "Test User"); err != nil {
		return fmt.Errorf("failed to set git user name: %w", err)
	}
	if err := mcg.runGitCommand(ctx, repoPath, "config", "user.email", "test@example.com"); err != nil {
		return fmt.Errorf("failed to set git user email: %w", err)
	}

	// Create initial commit
	readmePath := filepath.Join(repoPath, "README.md")
	content := fmt.Sprintf("# %s\n\nTest repository for merge conflict testing.\n", filepath.Base(repoPath))
	if err := os.WriteFile(readmePath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("failed to create README.md: %w", err)
	}

	if err := mcg.runGitCommand(ctx, repoPath, "add", "README.md"); err != nil {
		return fmt.Errorf("failed to add README.md: %w", err)
	}

	if err := mcg.runGitCommand(ctx, repoPath, "commit", "-m", "Initial commit"); err != nil {
		return fmt.Errorf("failed to create initial commit: %w", err)
	}

	return nil
}

// runGitCommand executes a git command with timeout
func (mcg *MergeConflictGenerator) runGitCommand(ctx context.Context, dir string, args ...string) error {
	ctx, cancel := context.WithTimeout(ctx, mcg.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git %v failed: %w, output: %s", args, err, string(output))
	}

	return nil
}
