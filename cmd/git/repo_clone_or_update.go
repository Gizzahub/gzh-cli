// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Gizzahub/gzh-cli/internal/git"
	"github.com/Gizzahub/gzh-cli/internal/logger"
)

// CloneOrUpdateStrategy defines the strategy to use when a repository already exists
type CloneOrUpdateStrategy string

const (
	// StrategyRebase rebases local changes on top of remote changes
	StrategyRebase CloneOrUpdateStrategy = "rebase"
	// StrategyReset performs a hard reset to match remote state (discards local changes)
	StrategyReset CloneOrUpdateStrategy = "reset"
	// StrategyClone removes existing directory and performs fresh clone
	StrategyClone CloneOrUpdateStrategy = "clone"
	// StrategySkip leaves the existing repository unchanged
	StrategySkip CloneOrUpdateStrategy = "skip"
	// StrategyPull performs a standard git pull (merge remote changes)
	StrategyPull CloneOrUpdateStrategy = "pull"
	// StrategyFetch only fetches remote changes without updating working directory
	StrategyFetch CloneOrUpdateStrategy = "fetch"
)

// cloneOrUpdateOptions holds the configuration for clone-or-update operations
type cloneOrUpdateOptions struct {
	repoURL    string
	targetPath string
	strategy   CloneOrUpdateStrategy
	branch     string
	depth      int
	force      bool
	verbose    bool
}

// newRepoCloneOrUpdateCmd creates the clone-or-update command for single repository operations
func newRepoCloneOrUpdateCmd() *cobra.Command {
	opts := &cloneOrUpdateOptions{
		strategy: StrategyRebase,
		depth:    0, // Full history by default
	}

	cmd := &cobra.Command{
		Use:   "clone-or-update <repository-url> [target-path]",
		Short: "Clone repository or update existing one with configurable strategies",
		Long: `Clone a repository if it doesn't exist, or update it using the specified strategy.

This command provides intelligent repository management by automatically detecting
whether a repository exists at the target path and taking the appropriate action.

If target-path is not provided, the repository name will be extracted from the URL
and used as the directory name (similar to 'git clone' behavior).

Available Strategies:
  rebase  - Rebase local changes on top of remote changes (default)
  reset   - Hard reset to match remote state (discards local changes)
  clone   - Remove existing directory and perform fresh clone
  skip    - Leave existing repository unchanged
  pull    - Standard git pull (merge remote changes)
  fetch   - Only fetch remote changes without updating working directory

Examples:
  # Clone into directory named from repository (e.g., 'repo')
  gz git repo clone-or-update https://github.com/user/repo.git

  # Clone or rebase existing repository with explicit path
  gz git repo clone-or-update https://github.com/user/repo.git ./my-repo

  # Force fresh clone by removing existing directory
  gz git repo clone-or-update --strategy clone https://github.com/user/repo.git

  # Update existing repository with hard reset (discard local changes)
  gz git repo clone-or-update --strategy reset https://github.com/user/repo.git ./repo

  # Skip existing repositories (useful for bulk operations)
  gz git repo clone-or-update --strategy skip https://github.com/user/repo.git

  # Clone specific branch with shallow history
  gz git repo clone-or-update --branch develop --depth 1 https://github.com/user/repo.git`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.repoURL = args[0]

			// If target path is not provided, extract repository name from URL
			if len(args) > 1 {
				opts.targetPath = args[1]
			} else {
				repoName, err := extractRepoNameFromURL(opts.repoURL)
				if err != nil {
					return fmt.Errorf("failed to extract repository name from URL: %w", err)
				}
				opts.targetPath = repoName
			}

			ctx := cmd.Context()
			return opts.run(ctx)
		},
	}

	// Add flags
	cmd.Flags().StringVarP((*string)(&opts.strategy), "strategy", "s", string(opts.strategy),
		"Strategy when repository exists: rebase, reset, clone, skip, pull, fetch")
	cmd.Flags().StringVarP(&opts.branch, "branch", "b", "",
		"Specific branch to clone/checkout (default: repository default branch)")
	cmd.Flags().IntVarP(&opts.depth, "depth", "d", opts.depth,
		"Create shallow clone with specified depth (0 for full history)")
	cmd.Flags().BoolVarP(&opts.force, "force", "f", opts.force,
		"Force operation even if it might be destructive")
	cmd.Flags().BoolVarP(&opts.verbose, "verbose", "v", opts.verbose,
		"Enable verbose logging")

	return cmd
}

// run executes the clone-or-update operation
func (opts *cloneOrUpdateOptions) run(ctx context.Context) error {
	// Validate strategy
	if !opts.isValidStrategy() {
		return fmt.Errorf("invalid strategy '%s'. Valid strategies: rebase, reset, clone, skip, pull, fetch", opts.strategy)
	}

	// Create absolute path
	absPath, err := filepath.Abs(opts.targetPath)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path: %w", err)
	}
	opts.targetPath = absPath

	if opts.verbose {
		fmt.Printf("Repository URL: %s\n", opts.repoURL)
		fmt.Printf("Target Path: %s\n", opts.targetPath)
		fmt.Printf("Strategy: %s\n", opts.strategy)
		if opts.branch != "" {
			fmt.Printf("Branch: %s\n", opts.branch)
		}
		if opts.depth > 0 {
			fmt.Printf("Depth: %d\n", opts.depth)
		}
	}

	// Check if target directory exists and contains a git repository
	exists, isGitRepo, err := opts.checkTargetDirectory()
	if err != nil {
		return fmt.Errorf("failed to check target directory: %w", err)
	}

	if opts.verbose {
		if exists {
			if isGitRepo {
				fmt.Printf("Target directory exists and is a git repository\n")
			} else {
				fmt.Printf("Target directory exists but is not a git repository\n")
			}
		} else {
			fmt.Printf("Target directory does not exist\n")
		}
	}

	// Execute appropriate action based on existence and strategy
	switch {
	case !exists:
		// Directory doesn't exist - perform clone
		return opts.performClone(ctx)

	case exists && !isGitRepo:
		// Directory exists but is not a git repo
		if opts.strategy == StrategyClone || opts.force {
			// Remove directory and clone
			if opts.verbose {
				fmt.Printf("Removing non-git directory and cloning...\n")
			}
			if err := os.RemoveAll(opts.targetPath); err != nil {
				return fmt.Errorf("failed to remove existing directory: %w", err)
			}
			return opts.performClone(ctx)
		}
		return fmt.Errorf("target directory exists but is not a git repository. Use --strategy=clone or --force to replace it")

	case exists && isGitRepo:
		// Directory exists and is a git repo - apply strategy
		return opts.applyUpdateStrategy(ctx)

	default:
		return fmt.Errorf("unexpected state in target directory analysis")
	}
}

// checkTargetDirectory checks if target directory exists and is a git repository
func (opts *cloneOrUpdateOptions) checkTargetDirectory() (exists bool, isGitRepo bool, err error) {
	// Check if directory exists
	info, err := os.Stat(opts.targetPath)
	if os.IsNotExist(err) {
		return false, false, nil
	}
	if err != nil {
		return false, false, err
	}
	if !info.IsDir() {
		return false, false, fmt.Errorf("target path exists but is not a directory")
	}

	// Check if it's a git repository
	gitDir := filepath.Join(opts.targetPath, ".git")
	_, err = os.Stat(gitDir)
	if os.IsNotExist(err) {
		return true, false, nil
	}
	if err != nil {
		return true, false, err
	}

	return true, true, nil
}

// performClone executes a fresh git clone operation
func (opts *cloneOrUpdateOptions) performClone(ctx context.Context) error {
	if opts.verbose {
		fmt.Printf("Cloning repository...\n")
	}

	// Create git client
	simpleLogger := logger.NewSimpleLogger("clone-or-update")
	executor := &simpleCommandExecutor{}
	gitClient := git.NewClient(nil, executor, simpleLogger)

	// Create clone options
	cloneOpts := git.CloneOptions{
		URL:    opts.repoURL,
		Path:   opts.targetPath,
		Branch: opts.branch,
		Depth:  opts.depth,
	}

	result, err := gitClient.Clone(ctx, cloneOpts)
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("clone operation failed: %s", result.Error)
	}

	fmt.Printf("✅ Successfully cloned %s to %s\n", opts.repoURL, opts.targetPath)
	return nil
}

// applyUpdateStrategy applies the specified update strategy to existing git repository
func (opts *cloneOrUpdateOptions) applyUpdateStrategy(ctx context.Context) error {
	if opts.verbose {
		fmt.Printf("Applying %s strategy to existing repository...\n", opts.strategy)
	}

	// Create git client
	simpleLogger := logger.NewSimpleLogger("clone-or-update")
	executor := &simpleCommandExecutor{}
	gitClient := git.NewClient(nil, executor, simpleLogger)

	switch opts.strategy {
	case StrategySkip:
		fmt.Printf("⏭️  Skipping existing repository at %s\n", opts.targetPath)
		return nil

	case StrategyClone:
		// Remove and re-clone
		if opts.verbose {
			fmt.Printf("Removing existing repository and cloning fresh...\n")
		}
		if err := os.RemoveAll(opts.targetPath); err != nil {
			return fmt.Errorf("failed to remove existing repository: %w", err)
		}
		return opts.performClone(ctx)

	case StrategyFetch:
		// Only fetch remote changes
		result, err := gitClient.Fetch(ctx, opts.targetPath, "origin")
		if err != nil {
			return fmt.Errorf("failed to fetch: %w", err)
		}
		if !result.Success {
			return fmt.Errorf("fetch operation failed: %s", result.Error)
		}
		fmt.Printf("📥 Successfully fetched updates for %s\n", opts.targetPath)
		return nil

	case StrategyPull:
		// Standard git pull
		pullOpts := git.PullOptions{
			Remote: "origin",
			Branch: opts.branch,
		}
		result, err := gitClient.Pull(ctx, opts.targetPath, pullOpts)
		if err != nil {
			return fmt.Errorf("failed to pull: %w", err)
		}
		if !result.Success {
			return fmt.Errorf("pull operation failed: %s", result.Error)
		}
		fmt.Printf("🔄 Successfully pulled updates for %s\n", opts.targetPath)
		return nil

	case StrategyReset:
		// Hard reset to remote
		// First fetch to get latest remote state
		result, err := gitClient.Fetch(ctx, opts.targetPath, "origin")
		if err != nil {
			return fmt.Errorf("failed to fetch before reset: %w", err)
		}
		if !result.Success {
			return fmt.Errorf("fetch operation failed: %s", result.Error)
		}

		// Determine target branch for reset
		resetTarget := "origin/HEAD"
		if opts.branch != "" {
			resetTarget = fmt.Sprintf("origin/%s", opts.branch)
		}

		resetOpts := git.ResetOptions{
			Mode:   "hard",
			Target: resetTarget,
		}
		result, err = gitClient.Reset(ctx, opts.targetPath, resetOpts)
		if err != nil {
			return fmt.Errorf("failed to reset: %w", err)
		}
		if !result.Success {
			return fmt.Errorf("reset operation failed: %s", result.Error)
		}
		fmt.Printf("🔄 Successfully reset %s to %s\n", opts.targetPath, resetTarget)
		return nil

	case StrategyRebase:
		// For rebase, we'll use the strategy executor as it's simpler
		strategyExecutor := git.NewStrategyExecutor(gitClient, simpleLogger)
		result, err := strategyExecutor.ExecuteStrategy(ctx, opts.targetPath, "pull")
		if err != nil {
			return fmt.Errorf("failed to rebase: %w", err)
		}
		if !result.Success {
			return fmt.Errorf("rebase operation failed: %s", result.Error)
		}
		fmt.Printf("🔄 Successfully rebased %s\n", opts.targetPath)
		return nil

	default:
		return fmt.Errorf("unsupported strategy: %s", opts.strategy)
	}
}

// isValidStrategy validates if the provided strategy is supported
func (opts *cloneOrUpdateOptions) isValidStrategy() bool {
	switch opts.strategy {
	case StrategyRebase, StrategyReset, StrategyClone, StrategySkip, StrategyPull, StrategyFetch:
		return true
	default:
		return false
	}
}

// simpleCommandExecutor implements git.CommandExecutor
type simpleCommandExecutor struct{}

// Execute executes a command in the current directory
func (e *simpleCommandExecutor) Execute(ctx context.Context, command string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, command, args...)
	return cmd.CombinedOutput()
}

// ExecuteInDir executes a command in the specified directory
func (e *simpleCommandExecutor) ExecuteInDir(ctx context.Context, dir, command string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = dir
	return cmd.CombinedOutput()
}

// extractRepoNameFromURL extracts the repository name from a Git URL
func extractRepoNameFromURL(repoURL string) (string, error) {
	if repoURL == "" {
		return "", fmt.Errorf("repository URL cannot be empty")
	}

	// Remove common Git URL prefixes and suffixes
	url := strings.TrimSpace(repoURL)

	// Handle different URL formats:
	// https://github.com/user/repo.git
	// git@github.com:user/repo.git
	// https://gitlab.com/user/repo
	// ssh://git@server.com/user/repo.git

	// Remove .git suffix if present
	if strings.HasSuffix(url, ".git") {
		url = strings.TrimSuffix(url, ".git")
	}

	var repoPath string

	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		// HTTP/HTTPS URLs: https://github.com/user/repo
		parts := strings.Split(url, "/")
		if len(parts) < 2 {
			return "", fmt.Errorf("invalid HTTP/HTTPS URL format: %s", repoURL)
		}
		repoPath = parts[len(parts)-1]
	} else if strings.Contains(url, "@") && strings.Contains(url, ":") {
		// SSH URLs: git@github.com:user/repo
		if strings.HasPrefix(url, "ssh://") {
			// ssh://git@server.com/user/repo
			parts := strings.Split(url, "/")
			if len(parts) < 2 {
				return "", fmt.Errorf("invalid SSH URL format: %s", repoURL)
			}
			repoPath = parts[len(parts)-1]
		} else {
			// git@github.com:user/repo
			colonIndex := strings.LastIndex(url, ":")
			if colonIndex == -1 {
				return "", fmt.Errorf("invalid SSH URL format: %s", repoURL)
			}
			pathPart := url[colonIndex+1:]
			parts := strings.Split(pathPart, "/")
			repoPath = parts[len(parts)-1]
		}
	} else {
		// Fallback: try to extract from the last part of the path
		parts := strings.Split(url, "/")
		if len(parts) < 1 {
			return "", fmt.Errorf("unable to extract repository name from URL: %s", repoURL)
		}
		repoPath = parts[len(parts)-1]
	}

	// Clean up the repository name
	repoPath = strings.TrimSpace(repoPath)
	if repoPath == "" {
		return "", fmt.Errorf("extracted repository name is empty from URL: %s", repoURL)
	}

	// Validate repository name (basic validation)
	if strings.Contains(repoPath, " ") || strings.Contains(repoPath, "\t") {
		return "", fmt.Errorf("invalid repository name extracted: %s", repoPath)
	}

	return repoPath, nil
}
