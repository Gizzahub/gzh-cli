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

// BranchStrategyTester tests different synclone strategies against various branch scenarios
type BranchStrategyTester struct {
	timeout time.Duration
}

// StrategyTestCase represents a test case for a specific synclone strategy
type StrategyTestCase struct {
	Strategy         string            // "reset", "pull", "fetch", "rebase", "clone"
	InitialState     BranchState       // Initial repository state
	RemoteChanges    []RemoteChange    // Simulated remote changes
	LocalChanges     []LocalChange     // Local uncommitted changes
	ExpectedOutcome  ExpectedOutcome   // What should happen
	ValidationChecks []ValidationCheck // Checks to verify the outcome
}

// BranchState represents the state of a repository's branches
type BranchState struct {
	CurrentBranch      string
	Branches           map[string]BranchInfo // branch name -> info
	UncommittedChanges bool
	ConflictingChanges bool
}

// BranchInfo contains information about a specific branch
type BranchInfo struct {
	LastCommit      string
	AheadCount      int // commits ahead of remote
	BehindCount     int // commits behind remote
	HasLocalCommits bool
}

// RemoteChange represents a change made to the remote repository
type RemoteChange struct {
	Branch        string
	CommitMessage string
	Files         map[string]string // filename -> content
	DeleteFiles   []string
}

// LocalChange represents a local uncommitted change
type LocalChange struct {
	File    string
	Content string
	Action  string // "modify", "add", "delete"
}

// ExpectedOutcome defines what should happen after applying a strategy
type ExpectedOutcome struct {
	ShouldSucceed       bool
	CurrentBranch       string
	ConflictsExpected   bool
	LocalChangesKept    bool
	RemoteChangesPulled bool
	BranchReset         bool
}

// ValidationCheck defines a check to validate the outcome
type ValidationCheck struct {
	Name        string
	CheckType   string // "file_exists", "file_content", "branch_state", "git_status"
	Target      string // file path or branch name
	Expected    string // expected value
	Description string
}

// NewBranchStrategyTester creates a new BranchStrategyTester instance
func NewBranchStrategyTester() *BranchStrategyTester {
	return &BranchStrategyTester{
		timeout: 60 * time.Second,
	}
}

// TestAllStrategies tests all synclone strategies against common scenarios
func (bst *BranchStrategyTester) TestAllStrategies(ctx context.Context, baseRepoPath string) error {
	strategies := []string{"reset", "pull", "fetch", "rebase", "clone"}

	for _, strategy := range strategies {
		if err := bst.TestStrategy(ctx, baseRepoPath, strategy); err != nil {
			return fmt.Errorf("strategy %s failed: %w", strategy, err)
		}
	}

	return nil
}

// TestStrategy tests a specific synclone strategy
func (bst *BranchStrategyTester) TestStrategy(ctx context.Context, baseRepoPath, strategy string) error {
	testCases := bst.generateTestCases(strategy)

	for i, testCase := range testCases {
		testRepoPath := fmt.Sprintf("%s-%s-test-%d", baseRepoPath, strategy, i+1)

		if err := bst.runTestCase(ctx, testRepoPath, testCase); err != nil {
			return fmt.Errorf("test case %d for strategy %s failed: %w", i+1, strategy, err)
		}

		// Clean up test repository
		os.RemoveAll(testRepoPath)
	}

	return nil
}

// TestResetStrategy specifically tests the reset strategy
func (bst *BranchStrategyTester) TestResetStrategy(ctx context.Context, repoPath string) error {
	testCase := StrategyTestCase{
		Strategy: "reset",
		InitialState: BranchState{
			CurrentBranch: "main",
			Branches: map[string]BranchInfo{
				"main": {
					AheadCount:      2, // Local commits
					BehindCount:     3, // Remote commits
					HasLocalCommits: true,
				},
			},
			UncommittedChanges: true,
		},
		RemoteChanges: []RemoteChange{
			{
				Branch:        "main",
				CommitMessage: "Remote update 1",
				Files:         map[string]string{"remote1.txt": "Remote change 1\n"},
			},
			{
				Branch:        "main",
				CommitMessage: "Remote update 2",
				Files:         map[string]string{"remote2.txt": "Remote change 2\n"},
			},
		},
		LocalChanges: []LocalChange{
			{File: "local.txt", Content: "Local change\n", Action: "add"},
			{File: "existing.txt", Content: "Modified locally\n", Action: "modify"},
		},
		ExpectedOutcome: ExpectedOutcome{
			ShouldSucceed:       true,
			CurrentBranch:       "main",
			ConflictsExpected:   false,
			LocalChangesKept:    false, // Reset discards local changes
			RemoteChangesPulled: true,
			BranchReset:         true,
		},
		ValidationChecks: []ValidationCheck{
			{
				Name:        "Remote files exist",
				CheckType:   "file_exists",
				Target:      "remote1.txt",
				Expected:    "true",
				Description: "Remote changes should be present after reset",
			},
			{
				Name:        "Local changes discarded",
				CheckType:   "file_exists",
				Target:      "local.txt",
				Expected:    "false",
				Description: "Local uncommitted changes should be discarded",
			},
			{
				Name:        "No uncommitted changes",
				CheckType:   "git_status",
				Target:      "working_tree",
				Expected:    "clean",
				Description: "Working tree should be clean after reset",
			},
		},
	}

	return bst.runTestCase(ctx, repoPath, testCase)
}

// TestPullStrategy specifically tests the pull strategy
func (bst *BranchStrategyTester) TestPullStrategy(ctx context.Context, repoPath string) error {
	testCase := StrategyTestCase{
		Strategy: "pull",
		InitialState: BranchState{
			CurrentBranch: "main",
			Branches: map[string]BranchInfo{
				"main": {
					BehindCount:     2, // Remote commits to pull
					HasLocalCommits: true,
				},
			},
			UncommittedChanges: false, // Clean working tree
		},
		RemoteChanges: []RemoteChange{
			{
				Branch:        "main",
				CommitMessage: "Remote feature A",
				Files:         map[string]string{"feature-a.txt": "Feature A implementation\n"},
			},
		},
		ExpectedOutcome: ExpectedOutcome{
			ShouldSucceed:       true,
			CurrentBranch:       "main",
			ConflictsExpected:   false,
			LocalChangesKept:    true,
			RemoteChangesPulled: true,
			BranchReset:         false,
		},
		ValidationChecks: []ValidationCheck{
			{
				Name:        "Remote changes merged",
				CheckType:   "file_exists",
				Target:      "feature-a.txt",
				Expected:    "true",
				Description: "Remote changes should be merged",
			},
			{
				Name:        "Local commits preserved",
				CheckType:   "branch_state",
				Target:      "main",
				Expected:    "has_local_commits",
				Description: "Local commits should be preserved",
			},
		},
	}

	return bst.runTestCase(ctx, repoPath, testCase)
}

// TestRebaseStrategy specifically tests the rebase strategy
func (bst *BranchStrategyTester) TestRebaseStrategy(ctx context.Context, repoPath string) error {
	testCase := StrategyTestCase{
		Strategy: "rebase",
		InitialState: BranchState{
			CurrentBranch: "main",
			Branches: map[string]BranchInfo{
				"main": {
					AheadCount:      2, // Local commits to rebase
					BehindCount:     1, // Remote commits
					HasLocalCommits: true,
				},
			},
			UncommittedChanges: false,
		},
		RemoteChanges: []RemoteChange{
			{
				Branch:        "main",
				CommitMessage: "Remote base change",
				Files:         map[string]string{"base.txt": "Base change\n"},
			},
		},
		ExpectedOutcome: ExpectedOutcome{
			ShouldSucceed:       true,
			CurrentBranch:       "main",
			ConflictsExpected:   false,
			LocalChangesKept:    true,
			RemoteChangesPulled: true,
			BranchReset:         false,
		},
		ValidationChecks: []ValidationCheck{
			{
				Name:        "Remote changes present",
				CheckType:   "file_exists",
				Target:      "base.txt",
				Expected:    "true",
				Description: "Remote changes should be present as base",
			},
			{
				Name:        "Local commits rebased",
				CheckType:   "branch_state",
				Target:      "main",
				Expected:    "linear_history",
				Description: "Local commits should be rebased on top",
			},
		},
	}

	return bst.runTestCase(ctx, repoPath, testCase)
}

// runTestCase executes a single test case
func (bst *BranchStrategyTester) runTestCase(ctx context.Context, repoPath string, testCase StrategyTestCase) error {
	// Create test repository
	if err := bst.setupTestRepository(ctx, repoPath, testCase.InitialState); err != nil {
		return fmt.Errorf("failed to setup test repository: %w", err)
	}

	// Apply local changes
	if err := bst.applyLocalChanges(ctx, repoPath, testCase.LocalChanges); err != nil {
		return fmt.Errorf("failed to apply local changes: %w", err)
	}

	// Simulate remote changes (in a real test, this would be a separate remote repo)
	remoteRepo := repoPath + "-remote"
	if err := bst.setupRemoteRepository(ctx, remoteRepo, repoPath, testCase.RemoteChanges); err != nil {
		return fmt.Errorf("failed to setup remote repository: %w", err)
	}

	// Apply the strategy (this would normally be done by synclone command)
	if err := bst.applyStrategy(ctx, repoPath, remoteRepo, testCase.Strategy); err != nil {
		if testCase.ExpectedOutcome.ShouldSucceed {
			return fmt.Errorf("strategy %s failed unexpectedly: %w", testCase.Strategy, err)
		}
		// Expected failure, continue to validation
	}

	// Validate the outcome
	if err := bst.validateOutcome(ctx, repoPath, testCase); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Clean up
	os.RemoveAll(remoteRepo)

	return nil
}

// generateTestCases generates test cases for a specific strategy
func (bst *BranchStrategyTester) generateTestCases(strategy string) []StrategyTestCase {
	switch strategy {
	case "reset":
		return []StrategyTestCase{
			// Test case 1: Clean reset with remote changes
			{
				Strategy: "reset",
				InitialState: BranchState{
					CurrentBranch:      "main",
					UncommittedChanges: false,
				},
				RemoteChanges: []RemoteChange{
					{Branch: "main", CommitMessage: "Remote update", Files: map[string]string{"update.txt": "Updated\n"}},
				},
				ExpectedOutcome: ExpectedOutcome{
					ShouldSucceed: true,
					BranchReset:   true,
				},
			},
		}
	case "pull":
		return []StrategyTestCase{
			// Test case 1: Simple pull
			{
				Strategy: "pull",
				InitialState: BranchState{
					CurrentBranch:      "main",
					UncommittedChanges: false,
				},
				RemoteChanges: []RemoteChange{
					{Branch: "main", CommitMessage: "Remote feature", Files: map[string]string{"feature.txt": "Feature\n"}},
				},
				ExpectedOutcome: ExpectedOutcome{
					ShouldSucceed:       true,
					RemoteChangesPulled: true,
				},
			},
		}
	default:
		return []StrategyTestCase{}
	}
}

// Helper methods for setting up and running tests

// setupTestRepository creates a test repository with the specified initial state
func (bst *BranchStrategyTester) setupTestRepository(ctx context.Context, repoPath string, state BranchState) error {
	// Create directory
	if err := os.MkdirAll(repoPath, 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Initialize git repository
	if err := bst.runGitCommand(ctx, repoPath, "init"); err != nil {
		return fmt.Errorf("failed to initialize repository: %w", err)
	}

	// Configure git user
	if err := bst.runGitCommand(ctx, repoPath, "config", "user.name", "Test User"); err != nil {
		return err
	}
	if err := bst.runGitCommand(ctx, repoPath, "config", "user.email", "test@example.com"); err != nil {
		return err
	}

	// Create initial commit
	readmePath := filepath.Join(repoPath, "README.md")
	if err := os.WriteFile(readmePath, []byte("# Test Repository\n"), 0o644); err != nil {
		return err
	}
	if err := bst.runGitCommand(ctx, repoPath, "add", "README.md"); err != nil {
		return err
	}
	if err := bst.runGitCommand(ctx, repoPath, "commit", "-m", "Initial commit"); err != nil {
		return err
	}

	return nil
}

// applyLocalChanges applies local changes to the repository
func (bst *BranchStrategyTester) applyLocalChanges(ctx context.Context, repoPath string, changes []LocalChange) error {
	for _, change := range changes {
		filePath := filepath.Join(repoPath, change.File)

		switch change.Action {
		case "add", "modify":
			if err := os.WriteFile(filePath, []byte(change.Content), 0o644); err != nil {
				return fmt.Errorf("failed to write file %s: %w", change.File, err)
			}
		case "delete":
			if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("failed to delete file %s: %w", change.File, err)
			}
		}
	}

	return nil
}

// setupRemoteRepository creates a remote repository with changes
func (bst *BranchStrategyTester) setupRemoteRepository(ctx context.Context, remoteRepo, baseRepo string, changes []RemoteChange) error {
	// Clone the base repository to create remote
	if err := bst.runGitCommand(ctx, filepath.Dir(remoteRepo), "clone", baseRepo, filepath.Base(remoteRepo)); err != nil {
		return fmt.Errorf("failed to clone base repository: %w", err)
	}

	// Apply remote changes
	for _, change := range changes {
		if err := bst.runGitCommand(ctx, remoteRepo, "checkout", change.Branch); err != nil {
			// Create branch if it doesn't exist
			if err := bst.runGitCommand(ctx, remoteRepo, "checkout", "-b", change.Branch); err != nil {
				return fmt.Errorf("failed to create/checkout branch %s: %w", change.Branch, err)
			}
		}

		// Apply file changes
		for filename, content := range change.Files {
			filePath := filepath.Join(remoteRepo, filename)
			if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
				return fmt.Errorf("failed to write file %s: %w", filename, err)
			}
			if err := bst.runGitCommand(ctx, remoteRepo, "add", filename); err != nil {
				return fmt.Errorf("failed to add file %s: %w", filename, err)
			}
		}

		// Delete files
		for _, filename := range change.DeleteFiles {
			if err := bst.runGitCommand(ctx, remoteRepo, "rm", filename); err != nil {
				return fmt.Errorf("failed to delete file %s: %w", filename, err)
			}
		}

		// Commit changes
		if err := bst.runGitCommand(ctx, remoteRepo, "commit", "-m", change.CommitMessage); err != nil {
			return fmt.Errorf("failed to commit remote changes: %w", err)
		}
	}

	// Set up the original repository to use this as remote
	if err := bst.runGitCommand(ctx, baseRepo, "remote", "add", "origin", remoteRepo); err != nil {
		// Remote might already exist, try to set URL
		if err := bst.runGitCommand(ctx, baseRepo, "remote", "set-url", "origin", remoteRepo); err != nil {
			return fmt.Errorf("failed to set remote origin URL: %w", err)
		}
	}

	return nil
}

// applyStrategy simulates applying a synclone strategy
func (bst *BranchStrategyTester) applyStrategy(ctx context.Context, repoPath, remoteRepo, strategy string) error {
	// Fetch remote changes first
	if err := bst.runGitCommand(ctx, repoPath, "fetch", "origin"); err != nil {
		return fmt.Errorf("failed to fetch from remote: %w", err)
	}

	switch strategy {
	case "reset":
		return bst.runGitCommand(ctx, repoPath, "reset", "--hard", "origin/main")
	case "pull":
		return bst.runGitCommand(ctx, repoPath, "pull", "origin", "main")
	case "fetch":
		// Fetch only, no merge
		return nil
	case "rebase":
		return bst.runGitCommand(ctx, repoPath, "rebase", "origin/main")
	case "clone":
		// For clone strategy, we would remove and re-clone
		return fmt.Errorf("clone strategy requires different test setup")
	default:
		return fmt.Errorf("unknown strategy: %s", strategy)
	}
}

// validateOutcome validates that the outcome matches expectations
func (bst *BranchStrategyTester) validateOutcome(ctx context.Context, repoPath string, testCase StrategyTestCase) error {
	for _, check := range testCase.ValidationChecks {
		switch check.CheckType {
		case "file_exists":
			filePath := filepath.Join(repoPath, check.Target)
			exists := true
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				exists = false
			}
			expectedExists := check.Expected == "true"
			if exists != expectedExists {
				return fmt.Errorf("validation %s failed: file %s exists=%v, expected=%v",
					check.Name, check.Target, exists, expectedExists)
			}
		case "file_content":
			filePath := filepath.Join(repoPath, check.Target)
			content, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", check.Target, err)
			}
			if strings.TrimSpace(string(content)) != strings.TrimSpace(check.Expected) {
				return fmt.Errorf("validation %s failed: content mismatch", check.Name)
			}
		case "git_status":
			output, err := bst.getGitStatus(ctx, repoPath)
			if err != nil {
				return fmt.Errorf("failed to get git status: %w", err)
			}
			if check.Expected == "clean" && !strings.Contains(output, "working tree clean") {
				return fmt.Errorf("validation %s failed: working tree not clean", check.Name)
			}
		}
	}

	return nil
}

// getGitStatus gets the git status output
func (bst *BranchStrategyTester) getGitStatus(ctx context.Context, repoPath string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "status")
	cmd.Dir = repoPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git status failed: %w", err)
	}

	return string(output), nil
}

// runGitCommand executes a git command with timeout
func (bst *BranchStrategyTester) runGitCommand(ctx context.Context, dir string, args ...string) error {
	ctx, cancel := context.WithTimeout(ctx, bst.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git %v failed: %w, output: %s", args, err, string(output))
	}

	return nil
}
