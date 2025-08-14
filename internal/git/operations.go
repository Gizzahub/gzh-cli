package git

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Gizzahub/gzh-cli/internal/gitplatform"
)

// Operations provides common git operations.
type Operations struct {
	verbose bool
}

// NewOperations creates a new git operations handler.
func NewOperations(verbose bool) *Operations {
	return &Operations{
		verbose: verbose,
	}
}

// Clone clones a repository to the specified path.
func (o *Operations) Clone(ctx context.Context, cloneURL, targetPath string) error {
	// Ensure parent directory exists
	parentDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(parentDir, 0o750); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	cmd := exec.CommandContext(ctx, "git", "clone", cloneURL, targetPath)

	if o.verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	return nil
}

// ExecuteStrategy executes the specified git strategy in the repository path.
func (o *Operations) ExecuteStrategy(ctx context.Context, repoPath string, strategy gitplatform.CloneStrategy) error {
	switch strategy {
	case gitplatform.StrategyReset:
		return o.resetStrategy(ctx, repoPath)
	case gitplatform.StrategyPull:
		return o.pullStrategy(ctx, repoPath)
	case gitplatform.StrategyFetch:
		return o.fetchStrategy(ctx, repoPath)
	default:
		return fmt.Errorf("unknown strategy: %s", strategy)
	}
}

// resetStrategy performs git reset --hard and git pull.
func (o *Operations) resetStrategy(ctx context.Context, repoPath string) error {
	// First, reset any local changes
	resetCmd := exec.CommandContext(ctx, "git", "reset", "--hard")
	resetCmd.Dir = repoPath

	var resetBuf bytes.Buffer

	resetCmd.Stdout = &resetBuf
	resetCmd.Stderr = &resetBuf

	if err := resetCmd.Run(); err != nil {
		return fmt.Errorf("git reset failed: %w\nOutput: %s", err, resetBuf.String())
	}

	// Then pull latest changes
	pullCmd := exec.CommandContext(ctx, "git", "pull")
	pullCmd.Dir = repoPath

	var pullBuf bytes.Buffer

	pullCmd.Stdout = &pullBuf
	pullCmd.Stderr = &pullBuf

	if err := pullCmd.Run(); err != nil {
		return fmt.Errorf("git pull failed: %w\nOutput: %s", err, pullBuf.String())
	}

	if o.verbose {
		fmt.Printf("Reset strategy completed for %s\n", repoPath)
	}

	return nil
}

// pullStrategy performs git pull.
func (o *Operations) pullStrategy(ctx context.Context, repoPath string) error {
	pullCmd := exec.CommandContext(ctx, "git", "pull")
	pullCmd.Dir = repoPath

	var buf bytes.Buffer

	pullCmd.Stdout = &buf
	pullCmd.Stderr = &buf

	if err := pullCmd.Run(); err != nil {
		return fmt.Errorf("git pull failed: %w\nOutput: %s", err, buf.String())
	}

	if o.verbose {
		fmt.Printf("Pull strategy completed for %s\n", repoPath)
	}

	return nil
}

// fetchStrategy performs git fetch.
func (o *Operations) fetchStrategy(ctx context.Context, repoPath string) error {
	fetchCmd := exec.CommandContext(ctx, "git", "fetch")
	fetchCmd.Dir = repoPath

	var buf bytes.Buffer

	fetchCmd.Stdout = &buf
	fetchCmd.Stderr = &buf

	if err := fetchCmd.Run(); err != nil {
		return fmt.Errorf("git fetch failed: %w\nOutput: %s", err, buf.String())
	}

	if o.verbose {
		fmt.Printf("Fetch strategy completed for %s\n", repoPath)
	}

	return nil
}

// IsGitRepository checks if a directory is a git repository.
func IsGitRepository(path string) bool {
	gitDir := filepath.Join(path, ".git")

	info, err := os.Stat(gitDir)
	if err != nil {
		return false
	}

	return info.IsDir()
}

// GetRemoteURL gets the remote URL for a repository.
func GetRemoteURL(ctx context.Context, repoPath string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "remote", "get-url", "origin")
	cmd.Dir = repoPath

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get remote URL: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// GetCurrentBranch gets the current branch name.
func GetCurrentBranch(ctx context.Context, repoPath string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "branch", "--show-current")
	cmd.Dir = repoPath

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// CheckoutBranch checks out a specific branch.
func CheckoutBranch(ctx context.Context, repoPath, branch string) error {
	cmd := exec.CommandContext(ctx, "git", "checkout", branch)
	cmd.Dir = repoPath

	var buf bytes.Buffer

	cmd.Stdout = &buf
	cmd.Stderr = &buf

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to checkout branch %s: %w\nOutput: %s", branch, err, buf.String())
	}

	return nil
}

// HasUncommittedChanges checks if the repository has uncommitted changes.
func HasUncommittedChanges(ctx context.Context, repoPath string) (bool, error) {
	cmd := exec.CommandContext(ctx, "git", "status", "--porcelain")
	cmd.Dir = repoPath

	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to check git status: %w", err)
	}

	return len(strings.TrimSpace(string(output))) > 0, nil
}
