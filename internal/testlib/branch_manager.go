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

// BranchManager provides functionality for managing Git branches
// and creating various branching scenarios for synclone testing.
type BranchManager struct {
	timeout time.Duration
}

// BranchScenario defines a branching scenario for testing
type BranchScenario struct {
	Name        string
	Branches    []BranchConfig
	Merges      []MergeConfig
	Description string
}

// BranchConfig defines configuration for a single branch
type BranchConfig struct {
	Name         string
	StartingFrom string // Branch to branch from (empty for main/master)
	Commits      []CommitConfig
}

// CommitConfig defines a commit to be made on a branch
type CommitConfig struct {
	Message string
	Files   map[string]string // filename -> content
	IsTag   bool
	TagName string
}

// MergeConfig defines a merge operation
type MergeConfig struct {
	FromBranch string
	ToBranch   string
	Message    string
	Strategy   string // "merge", "rebase", "squash"
}

// NewBranchManager creates a new BranchManager instance
func NewBranchManager() *BranchManager {
	return &BranchManager{
		timeout: 60 * time.Second,
	}
}

// CreateBranchScenario creates a specific branching scenario
func (bm *BranchManager) CreateBranchScenario(ctx context.Context, repoPath string, scenario BranchScenario) error {
	if repoPath == "" {
		return fmt.Errorf("repository path is required")
	}

	// Ensure we have a base repository
	if !bm.isGitRepo(repoPath) {
		if err := bm.initializeRepo(ctx, repoPath); err != nil {
			return fmt.Errorf("failed to initialize repository: %w", err)
		}
	}

	// Create branches and their commits
	for _, branchConfig := range scenario.Branches {
		if err := bm.createBranchWithCommits(ctx, repoPath, branchConfig); err != nil {
			return fmt.Errorf("failed to create branch %s: %w", branchConfig.Name, err)
		}
	}

	// Perform merges
	for _, mergeConfig := range scenario.Merges {
		if err := bm.performMerge(ctx, repoPath, mergeConfig); err != nil {
			return fmt.Errorf("failed to perform merge from %s to %s: %w",
				mergeConfig.FromBranch, mergeConfig.ToBranch, err)
		}
	}

	return nil
}

// CreateGitFlowScenario creates a Git Flow branching model scenario
func (bm *BranchManager) CreateGitFlowScenario(ctx context.Context, repoPath string) error {
	scenario := BranchScenario{
		Name:        "Git Flow",
		Description: "Standard Git Flow with main, develop, feature, and hotfix branches",
		Branches: []BranchConfig{
			{
				Name: "develop",
				Commits: []CommitConfig{
					{
						Message: "Start develop branch",
						Files:   map[string]string{"develop.txt": "Development branch started\n"},
					},
				},
			},
			{
				Name:         "feature/user-authentication",
				StartingFrom: "develop",
				Commits: []CommitConfig{
					{
						Message: "Add user model",
						Files:   map[string]string{"src/user.go": "package main\n\ntype User struct {\n\tID string\n\tName string\n}\n"},
					},
					{
						Message: "Add authentication logic",
						Files:   map[string]string{"src/auth.go": "package main\n\nfunc Authenticate(user User) bool {\n\treturn true\n}\n"},
					},
				},
			},
			{
				Name:         "feature/payment-system",
				StartingFrom: "develop",
				Commits: []CommitConfig{
					{
						Message: "Add payment model",
						Files:   map[string]string{"src/payment.go": "package main\n\ntype Payment struct {\n\tAmount float64\n}\n"},
					},
				},
			},
			{
				Name:         "release/v1.0.0",
				StartingFrom: "develop",
				Commits: []CommitConfig{
					{
						Message: "Prepare release v1.0.0",
						Files:   map[string]string{"VERSION": "1.0.0\n"},
						IsTag:   true,
						TagName: "v1.0.0",
					},
				},
			},
			{
				Name:         "hotfix/critical-security-fix",
				StartingFrom: "main",
				Commits: []CommitConfig{
					{
						Message: "Fix critical security vulnerability",
						Files:   map[string]string{"src/security.go": "package main\n\nfunc SecureFunction() {\n\t// Fixed security issue\n}\n"},
						IsTag:   true,
						TagName: "v1.0.1",
					},
				},
			},
		},
		Merges: []MergeConfig{
			{
				FromBranch: "feature/user-authentication",
				ToBranch:   "develop",
				Message:    "Merge user authentication feature",
				Strategy:   "merge",
			},
			{
				FromBranch: "feature/payment-system",
				ToBranch:   "develop",
				Message:    "Merge payment system feature",
				Strategy:   "merge",
			},
			{
				FromBranch: "release/v1.0.0",
				ToBranch:   "main",
				Message:    "Merge release v1.0.0",
				Strategy:   "merge",
			},
			{
				FromBranch: "hotfix/critical-security-fix",
				ToBranch:   "main",
				Message:    "Merge hotfix",
				Strategy:   "merge",
			},
			{
				FromBranch: "hotfix/critical-security-fix",
				ToBranch:   "develop",
				Message:    "Merge hotfix into develop",
				Strategy:   "merge",
			},
		},
	}

	return bm.CreateBranchScenario(ctx, repoPath, scenario)
}

// CreateDivergentBranchesScenario creates branches that have diverged significantly
func (bm *BranchManager) CreateDivergentBranchesScenario(ctx context.Context, repoPath string) error {
	scenario := BranchScenario{
		Name:        "Divergent Branches",
		Description: "Branches with significant divergence for testing conflict resolution",
		Branches: []BranchConfig{
			{
				Name: "experimental",
				Commits: []CommitConfig{
					{
						Message: "Experimental feature A",
						Files:   map[string]string{"experimental.txt": "Feature A implementation\n"},
					},
					{
						Message: "Experimental feature B",
						Files:   map[string]string{"experimental.txt": "Feature A implementation\nFeature B implementation\n"},
					},
					{
						Message: "Experimental refactoring",
						Files:   map[string]string{"refactored.txt": "Major refactoring\n"},
					},
				},
			},
			{
				Name: "stable",
				Commits: []CommitConfig{
					{
						Message: "Stable feature 1",
						Files:   map[string]string{"stable.txt": "Stable feature 1\n"},
					},
					{
						Message: "Bug fixes",
						Files:   map[string]string{"bugfixes.txt": "Various bug fixes\n"},
					},
					{
						Message: "Performance improvements",
						Files:   map[string]string{"performance.txt": "Performance optimizations\n"},
					},
				},
			},
		},
		// No merges to keep branches divergent
		Merges: []MergeConfig{},
	}

	// First create the scenario
	if err := bm.CreateBranchScenario(ctx, repoPath, scenario); err != nil {
		return err
	}

	// Add more commits to main to create further divergence
	mainCommits := []CommitConfig{
		{
			Message: "Main branch update 1",
			Files:   map[string]string{"main-update-1.txt": "Update 1 on main\n"},
		},
		{
			Message: "Main branch update 2",
			Files:   map[string]string{"main-update-2.txt": "Update 2 on main\n"},
		},
	}

	// Switch to main and add commits
	if err := bm.runGitCommand(ctx, repoPath, "checkout", "main"); err != nil {
		bm.runGitCommand(ctx, repoPath, "checkout", "master")
	}

	for _, commit := range mainCommits {
		if err := bm.createCommit(ctx, repoPath, commit); err != nil {
			return fmt.Errorf("failed to create commit on main: %w", err)
		}
	}

	return nil
}

// GetBranchInfo returns information about all branches in the repository
func (bm *BranchManager) GetBranchInfo(ctx context.Context, repoPath string) ([]string, error) {
	cmd := exec.CommandContext(ctx, "git", "branch", "-a")
	cmd.Dir = repoPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get branch info: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	var branches []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "*") {
			// Remove leading "* " for current branch
			line = strings.TrimPrefix(line, "* ")
			branches = append(branches, line)
		}
	}

	return branches, nil
}

// SwitchToBranch switches to a specified branch
func (bm *BranchManager) SwitchToBranch(ctx context.Context, repoPath, branchName string) error {
	return bm.runGitCommand(ctx, repoPath, "checkout", branchName)
}

// createBranchWithCommits creates a branch and adds the specified commits
func (bm *BranchManager) createBranchWithCommits(ctx context.Context, repoPath string, branchConfig BranchConfig) error {
	// Determine starting branch
	startBranch := "main"
	if branchConfig.StartingFrom != "" {
		startBranch = branchConfig.StartingFrom
	}

	// Switch to starting branch
	if err := bm.runGitCommand(ctx, repoPath, "checkout", startBranch); err != nil {
		// Try master if main doesn't exist
		if err := bm.runGitCommand(ctx, repoPath, "checkout", "master"); err != nil {
			return fmt.Errorf("failed to checkout starting branch %s: %w", startBranch, err)
		}
	}

	// Create new branch
	if err := bm.runGitCommand(ctx, repoPath, "checkout", "-b", branchConfig.Name); err != nil {
		return fmt.Errorf("failed to create branch %s: %w", branchConfig.Name, err)
	}

	// Add commits
	for _, commit := range branchConfig.Commits {
		if err := bm.createCommit(ctx, repoPath, commit); err != nil {
			return fmt.Errorf("failed to create commit on branch %s: %w", branchConfig.Name, err)
		}
	}

	return nil
}

// createCommit creates a single commit with the specified files and message
func (bm *BranchManager) createCommit(ctx context.Context, repoPath string, commit CommitConfig) error {
	// Create files
	for filename, content := range commit.Files {
		filePath := filepath.Join(repoPath, filename)

		// Create directory if needed
		if dir := filepath.Dir(filePath); dir != "." && dir != repoPath {
			if err := os.MkdirAll(filepath.Join(repoPath, filepath.Dir(filename)), 0755); err != nil {
				return fmt.Errorf("failed to create directory for %s: %w", filename, err)
			}
		}

		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create file %s: %w", filename, err)
		}

		if err := bm.runGitCommand(ctx, repoPath, "add", filename); err != nil {
			return fmt.Errorf("failed to add file %s: %w", filename, err)
		}
	}

	// Create commit
	if err := bm.runGitCommand(ctx, repoPath, "commit", "-m", commit.Message); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	// Create tag if specified
	if commit.IsTag && commit.TagName != "" {
		if err := bm.runGitCommand(ctx, repoPath, "tag", "-a", commit.TagName, "-m", commit.Message); err != nil {
			return fmt.Errorf("failed to create tag %s: %w", commit.TagName, err)
		}
	}

	return nil
}

// performMerge performs a merge operation
func (bm *BranchManager) performMerge(ctx context.Context, repoPath string, mergeConfig MergeConfig) error {
	// Switch to target branch
	if err := bm.runGitCommand(ctx, repoPath, "checkout", mergeConfig.ToBranch); err != nil {
		return fmt.Errorf("failed to checkout target branch %s: %w", mergeConfig.ToBranch, err)
	}

	// Perform merge based on strategy
	var args []string
	switch mergeConfig.Strategy {
	case "rebase":
		args = []string{"rebase", mergeConfig.FromBranch}
	case "squash":
		args = []string{"merge", "--squash", mergeConfig.FromBranch}
	default: // "merge"
		args = []string{"merge", "--no-ff", mergeConfig.FromBranch, "-m", mergeConfig.Message}
	}

	if err := bm.runGitCommand(ctx, repoPath, args...); err != nil {
		return fmt.Errorf("failed to %s %s into %s: %w", mergeConfig.Strategy, mergeConfig.FromBranch, mergeConfig.ToBranch, err)
	}

	// If squash, need to commit
	if mergeConfig.Strategy == "squash" {
		if err := bm.runGitCommand(ctx, repoPath, "commit", "-m", mergeConfig.Message); err != nil {
			return fmt.Errorf("failed to commit squash merge: %w", err)
		}
	}

	return nil
}

// isGitRepo checks if a directory is a Git repository
func (bm *BranchManager) isGitRepo(repoPath string) bool {
	gitDir := filepath.Join(repoPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return false
	}
	return true
}

// initializeRepo initializes a Git repository
func (bm *BranchManager) initializeRepo(ctx context.Context, repoPath string) error {
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := bm.runGitCommand(ctx, repoPath, "init"); err != nil {
		return fmt.Errorf("failed to initialize git repository: %w", err)
	}

	// Configure git user
	if err := bm.runGitCommand(ctx, repoPath, "config", "user.name", "Test User"); err != nil {
		return fmt.Errorf("failed to set git user name: %w", err)
	}
	if err := bm.runGitCommand(ctx, repoPath, "config", "user.email", "test@example.com"); err != nil {
		return fmt.Errorf("failed to set git user email: %w", err)
	}

	// Create initial commit
	readmePath := filepath.Join(repoPath, "README.md")
	content := fmt.Sprintf("# %s\n\nTest repository for branch management testing.\n", filepath.Base(repoPath))
	if err := os.WriteFile(readmePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to create README.md: %w", err)
	}

	if err := bm.runGitCommand(ctx, repoPath, "add", "README.md"); err != nil {
		return fmt.Errorf("failed to add README.md: %w", err)
	}

	if err := bm.runGitCommand(ctx, repoPath, "commit", "-m", "Initial commit"); err != nil {
		return fmt.Errorf("failed to create initial commit: %w", err)
	}

	return nil
}

// runGitCommand executes a git command with timeout
func (bm *BranchManager) runGitCommand(ctx context.Context, dir string, args ...string) error {
	ctx, cancel := context.WithTimeout(ctx, bm.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git %v failed: %w, output: %s", args, err, string(output))
	}

	return nil
}
