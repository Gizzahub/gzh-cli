// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package repo

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Gizzahub/gzh-cli/internal/app"

	gitrepo "github.com/gizzahub/gzh-cli-git/pkg/repository"
)

// newRepoCloneOrUpdateCmd creates the clone-or-update command using gzh-cli-git library
// This delegates repository synchronization functionality to the external gzh-cli-git package,
// avoiding code duplication and ensuring consistency with the standalone git CLI.
//
// NOTE: This function replaces the old implementation in repo_clone_or_update.go
func newRepoCloneOrUpdateCmd() *cobra.Command {
	// Get appCtx from context or use nil (library handles nil gracefully)
	// For now, we'll create a minimal appCtx inline
	appCtx := &app.AppContext{}

	return newRepoCloneOrUpdateCmdWithContext(appCtx)
}

// newRepoCloneOrUpdateCmdWithContext is the internal implementation with app context
func newRepoCloneOrUpdateCmdWithContext(appCtx *app.AppContext) *cobra.Command {
	opts := &cloneOrUpdateCmdOptions{
		strategy: "rebase",
		depth:    0,
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
			return opts.run(cmd.Context(), args, appCtx)
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&opts.strategy, "strategy", "s", opts.strategy,
		"Strategy when repository exists: rebase, reset, clone, skip, pull, fetch")
	cmd.Flags().StringVarP(&opts.branch, "branch", "b", "",
		"Specific branch to clone/checkout (default: repository default branch)")
	cmd.Flags().IntVarP(&opts.depth, "depth", "d", opts.depth,
		"Create shallow clone with specified depth (0 for full history)")
	cmd.Flags().BoolVarP(&opts.force, "force", "f", false,
		"Force operation even if it might be destructive")
	cmd.Flags().BoolVarP(&opts.verbose, "verbose", "v", false,
		"Enable verbose logging")

	return cmd
}

// cloneOrUpdateCmdOptions holds the configuration for clone-or-update command
type cloneOrUpdateCmdOptions struct {
	strategy string
	branch   string
	depth    int
	force    bool
	verbose  bool
}

// run executes the clone-or-update operation using gzh-cli-git library
func (opts *cloneOrUpdateCmdOptions) run(ctx context.Context, args []string, appCtx *app.AppContext) error {
	repoURL := args[0]
	var targetPath string

	// If target path is not provided, extract repository name from URL
	if len(args) > 1 {
		targetPath = args[1]
	} else {
		repoName, err := gitrepo.ExtractRepoNameFromURL(repoURL)
		if err != nil {
			return fmt.Errorf("failed to extract repository name from URL: %w", err)
		}
		targetPath = repoName
	}

	// Create logger adapter
	var logger gitrepo.Logger
	if opts.verbose && appCtx != nil && appCtx.Logger != nil {
		logger = &appContextLoggerAdapter{logger: appCtx.Logger}
	} else {
		logger = gitrepo.NewNoopLogger()
	}

	// Create git repository client
	client := gitrepo.NewClient(gitrepo.WithClientLogger(logger))

	// Prepare clone-or-update options
	updateOpts := gitrepo.CloneOrUpdateOptions{
		URL:         repoURL,
		Destination: targetPath,
		Strategy:    gitrepo.UpdateStrategy(opts.strategy),
		Branch:      opts.branch,
		Depth:       opts.depth,
		Force:       opts.force,
		Logger:      logger,
	}

	if opts.verbose {
		fmt.Printf("Repository URL: %s\n", updateOpts.URL)
		fmt.Printf("Target Path: %s\n", updateOpts.Destination)
		fmt.Printf("Strategy: %s\n", updateOpts.Strategy)
		if updateOpts.Branch != "" {
			fmt.Printf("Branch: %s\n", updateOpts.Branch)
		}
		if updateOpts.Depth > 0 {
			fmt.Printf("Depth: %d\n", updateOpts.Depth)
		}
	}

	// Execute clone-or-update
	result, err := client.CloneOrUpdate(ctx, updateOpts)
	if err != nil {
		return fmt.Errorf("failed to clone or update repository: %w", err)
	}

	// Print result based on action
	switch result.Action {
	case "cloned":
		fmt.Printf("✅ %s\n", result.Message)
	case "skipped":
		fmt.Printf("⏭️  %s\n", result.Message)
	case "fetched":
		fmt.Printf("📥 %s\n", result.Message)
	case "pulled":
		fmt.Printf("🔄 %s\n", result.Message)
	case "reset":
		fmt.Printf("🔄 %s\n", result.Message)
	case "rebased":
		fmt.Printf("🔄 %s\n", result.Message)
	default:
		fmt.Printf("✅ %s\n", result.Message)
	}

	return nil
}

// appContextLoggerAdapter adapts app.AppContext.Logger to gitrepo.Logger interface
type appContextLoggerAdapter struct {
	logger interface {
		Debug(msg string, args ...interface{})
		Info(msg string, args ...interface{})
		Warn(msg string, args ...interface{})
		Error(msg string, args ...interface{})
	}
}

func (a *appContextLoggerAdapter) Debug(msg string, args ...interface{}) {
	a.logger.Debug(msg, args...)
}

func (a *appContextLoggerAdapter) Info(msg string, args ...interface{}) {
	a.logger.Info(msg, args...)
}

func (a *appContextLoggerAdapter) Warn(msg string, args ...interface{}) {
	a.logger.Warn(msg, args...)
}

func (a *appContextLoggerAdapter) Error(msg string, args ...interface{}) {
	a.logger.Error(msg, args...)
}
