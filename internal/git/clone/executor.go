// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package clone

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/gizzahub/gzh-manager-go/pkg/git/provider"
)

// CloneExecutor handles the execution of repository cloning operations.
type CloneExecutor struct {
	provider provider.GitProvider
	options  *CloneOptions
	session  *Session
	progress *ProgressReporter
}

// NewCloneExecutor creates a new clone executor with the given provider and options.
func NewCloneExecutor(p provider.GitProvider, opts *CloneOptions) (*CloneExecutor, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("invalid options: %w", err)
	}

	session := NewSession(opts)
	progress := NewProgressReporter(opts.Format, opts.Quiet, opts.Verbose)

	return &CloneExecutor{
		provider: p,
		options:  opts,
		session:  session,
		progress: progress,
	}, nil
}

// Execute performs the clone operation based on the configured options.
func (e *CloneExecutor) Execute(ctx context.Context) error {
	// 1. Initialize or restore session
	if e.options.Resume != "" {
		if err := e.session.Load(e.options.Resume); err != nil {
			return NewSessionError(e.options.Resume, "load", "failed to resume session", err)
		}
		e.progress.ResumeSession(e.session)
	} else {
		if err := e.session.Initialize(); err != nil {
			return NewSessionError(e.session.ID, "initialize", "failed to initialize session", err)
		}
	}

	// 2. List repositories from provider
	repos, err := e.listRepositories(ctx)
	if err != nil {
		return fmt.Errorf("failed to list repositories: %w", err)
	}

	// 3. Apply filtering
	filtered := e.filterRepositories(repos)
	if len(filtered) == 0 {
		e.progress.Info("No repositories match the specified filters")
		return nil
	}

	// 4. Initialize progress tracking
	e.progress.Start(len(filtered))
	defer e.progress.Finish()

	// 5. Handle dry run
	if e.options.DryRun {
		return e.printDryRun(filtered)
	}

	// 6. Execute clone operations
	summary, err := e.cloneRepositories(ctx, filtered)
	if err != nil {
		e.progress.Error("Clone operation failed: %v", err)
		return err
	}

	// 7. Print summary
	e.printSummary(summary)

	return nil
}

// listRepositories retrieves repositories from the provider.
func (e *CloneExecutor) listRepositories(ctx context.Context) ([]provider.Repository, error) {
	// Convert visibility string to VisibilityType
	var visibility provider.VisibilityType
	switch e.options.Visibility {
	case "public":
		visibility = provider.VisibilityPublic
	case "private":
		visibility = provider.VisibilityPrivate
	default:
		visibility = "" // All
	}

	listOpts := provider.ListOptions{
		Organization: e.options.Org,
		Visibility:   visibility,
		Type:         "all",
		Sort:         "updated",
		Direction:    "desc",
		PerPage:      100,
	}

	// Set archived filter
	if !e.options.IncludeArchived {
		archived := false
		listOpts.Archived = &archived
	}

	repositoryList, err := e.provider.ListRepositories(ctx, listOpts)
	if err != nil {
		return nil, WrapNetworkError("", "list_repositories", err)
	}

	return repositoryList.Repositories, nil
}

// filterRepositories applies filtering options to the repository list.
func (e *CloneExecutor) filterRepositories(repos []provider.Repository) []RepositoryInfo {
	var filtered []RepositoryInfo

	for _, repo := range repos {
		repoInfo := RepositoryInfo{
			ID:            repo.ID,
			Name:          repo.Name,
			FullName:      repo.FullName,
			CloneURL:      repo.CloneURL,
			SSHURL:        repo.SSHURL,
			Private:       repo.Private,
			Archived:      repo.Archived,
			Fork:          repo.Fork,
			Language:      repo.Language,
			Topics:        repo.Topics,
			Stars:         repo.Stars,
			Forks:         repo.Forks,
			UpdatedAt:     repo.UpdatedAt,
			DefaultBranch: repo.DefaultBranch,
		}

		if repoInfo.Matches(e.options) {
			filtered = append(filtered, repoInfo)
		}
	}

	return filtered
}

// printDryRun prints what would be cloned without actually cloning.
func (e *CloneExecutor) printDryRun(repos []RepositoryInfo) error {
	e.progress.Info("Dry run - repositories that would be cloned:")
	for _, repo := range repos {
		targetPath := filepath.Join(e.options.Target, repo.FullName)
		e.progress.Info("  %s -> %s", repo.FullName, targetPath)
	}
	e.progress.Info("Total: %d repositories", len(repos))
	return nil
}

// cloneRepositories performs the actual cloning with parallel execution.
func (e *CloneExecutor) cloneRepositories(ctx context.Context, repos []RepositoryInfo) (*CloneSummary, error) {
	// Initialize summary
	summary := &CloneSummary{
		Total:     len(repos),
		StartTime: time.Now(),
	}

	// Create worker pool
	sem := make(chan struct{}, e.options.Parallel)
	resultChan := make(chan CloneResult, len(repos))
	var wg sync.WaitGroup

	// Process repositories
	for _, repo := range repos {
		// Skip if already completed in this session
		if e.session.IsCompleted(repo.FullName) {
			summary.Skipped++
			e.progress.Skip(repo.FullName, "already completed")
			continue
		}

		wg.Add(1)
		go func(r RepositoryInfo) {
			defer wg.Done()

			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			// Create clone request
			request := &CloneRequest{
				Repository: r,
				TargetPath: filepath.Join(e.options.Target, r.FullName),
				Options:    e.options,
				SessionID:  e.session.ID,
				StartedAt:  time.Now(),
			}

			// Execute clone with retries
			result := e.cloneWithRetries(ctx, request)
			resultChan <- result

			// Update session
			if result.Error != nil {
				e.session.MarkFailed(r.FullName, result.Error)
			} else {
				e.session.MarkCompleted(r.FullName)
			}

			// Save session progress
			if err := e.session.Save(); err != nil {
				e.progress.Warning("Failed to save session: %v", err)
			}
		}(repo)
	}

	// Wait for all workers to complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	for result := range resultChan {
		if result.Error != nil {
			summary.Failed++
			summary.Errors = append(summary.Errors, result.Error)
			e.progress.Fail(result.Repository.FullName, result.Error)
		} else {
			summary.Succeeded++
			e.progress.Success(result.Repository.FullName)
		}
	}

	summary.EndTime = time.Now()
	summary.Duration = summary.EndTime.Sub(summary.StartTime)

	return summary, nil
}

// cloneWithRetries performs clone operation with retry logic.
func (e *CloneExecutor) cloneWithRetries(ctx context.Context, request *CloneRequest) CloneResult {
	var lastErr error

	for attempt := 1; attempt <= e.options.MaxRetries+1; attempt++ {
		request.Attempt = attempt

		err := e.cloneRepository(ctx, request)
		if err == nil {
			request.CompletedAt = time.Now()
			return CloneResult{
				Repository: request.Repository,
				Success:    true,
			}
		}

		lastErr = err
		request.Error = err.Error()

		// Check if error is retryable
		if cloneErr, ok := err.(*CloneError); ok && !cloneErr.Retryable {
			break
		}

		// Don't retry on last attempt
		if attempt <= e.options.MaxRetries {
			e.progress.Retry(request.Repository.FullName, attempt, err)

			// Wait before retry
			select {
			case <-ctx.Done():
				return CloneResult{
					Repository: request.Repository,
					Error:      ctx.Err(),
				}
			case <-time.After(e.options.RetryDelay):
				// Continue to next attempt
			}
		}
	}

	return CloneResult{
		Repository: request.Repository,
		Error:      lastErr,
	}
}

// cloneRepository performs the actual clone operation for a single repository.
func (e *CloneExecutor) cloneRepository(ctx context.Context, request *CloneRequest) error {
	repo := request.Repository
	targetPath := request.TargetPath

	// Check if repository already exists
	if exists, err := e.pathExists(targetPath); err != nil {
		return NewCloneError(repo.FullName, "path_check", "failed to check target path", err)
	} else if exists {
		return e.handleExistingRepository(ctx, targetPath, repo)
	}

	// Clone new repository
	return e.cloneNewRepository(ctx, targetPath, repo)
}

// cloneNewRepository clones a repository to a new location.
func (e *CloneExecutor) cloneNewRepository(ctx context.Context, targetPath string, repo RepositoryInfo) error {
	// Create parent directory
	parentDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(parentDir, 0o755); err != nil {
		return NewCloneError(repo.FullName, "mkdir", "failed to create parent directory", err)
	}

	// Get clone URL based on protocol
	cloneURL := repo.GetCloneURL(e.options.Protocol)

	// Build git clone command
	args := []string{"clone"}

	if e.options.Depth > 0 {
		args = append(args, "--depth", fmt.Sprintf("%d", e.options.Depth))
	}

	if e.options.SingleBranch {
		args = append(args, "--single-branch")
	}

	if e.options.Branch != "" {
		args = append(args, "--branch", e.options.Branch)
	}

	args = append(args, cloneURL, targetPath)

	// Execute git clone
	cmd := exec.CommandContext(ctx, "git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Clean up partially cloned directory
		os.RemoveAll(targetPath)
		return WrapGitError(repo.FullName, "clone", err, output)
	}

	// Create GZH file if requested
	if e.options.CreateGZHFile {
		if err := e.createGZHFile(targetPath, repo); err != nil {
			e.progress.Warning("Failed to create .gzh file for %s: %v", repo.FullName, err)
		}
	}

	return nil
}

// handleExistingRepository handles operations on existing repositories.
func (e *CloneExecutor) handleExistingRepository(ctx context.Context, targetPath string, repo RepositoryInfo) error {
	switch e.options.Strategy {
	case StrategyReset:
		return e.resetAndPull(ctx, targetPath, repo)
	case StrategyPull:
		return e.pull(ctx, targetPath, repo)
	case StrategyFetch:
		return e.fetch(ctx, targetPath, repo)
	default:
		return NewCloneError(repo.FullName, "strategy", "unknown strategy", ErrInvalidStrategy)
	}
}

// resetAndPull performs git reset --hard and git pull.
func (e *CloneExecutor) resetAndPull(ctx context.Context, targetPath string, repo RepositoryInfo) error {
	// Change to repository directory
	originalDir, err := os.Getwd()
	if err != nil {
		return NewCloneError(repo.FullName, "getwd", "failed to get working directory", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(targetPath); err != nil {
		return NewCloneError(repo.FullName, "chdir", "failed to change to repository directory", err)
	}

	// Git reset --hard
	cmd := exec.CommandContext(ctx, "git", "reset", "--hard")
	if output, err := cmd.CombinedOutput(); err != nil {
		return WrapGitError(repo.FullName, "reset", err, output)
	}

	// Git pull
	cmd = exec.CommandContext(ctx, "git", "pull")
	if output, err := cmd.CombinedOutput(); err != nil {
		return WrapGitError(repo.FullName, "pull", err, output)
	}

	return nil
}

// pull performs git pull.
func (e *CloneExecutor) pull(ctx context.Context, targetPath string, repo RepositoryInfo) error {
	return e.runGitCommand(ctx, targetPath, repo, "pull", []string{"pull"})
}

// fetch performs git fetch.
func (e *CloneExecutor) fetch(ctx context.Context, targetPath string, repo RepositoryInfo) error {
	return e.runGitCommand(ctx, targetPath, repo, "fetch", []string{"fetch"})
}

// runGitCommand runs a git command in the specified directory.
func (e *CloneExecutor) runGitCommand(ctx context.Context, targetPath string, repo RepositoryInfo, operation string, args []string) error {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = targetPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return WrapGitError(repo.FullName, operation, err, output)
	}

	return nil
}

// pathExists checks if a path exists.
func (e *CloneExecutor) pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// createGZHFile creates a .gzh file with repository metadata.
func (e *CloneExecutor) createGZHFile(targetPath string, repo RepositoryInfo) error {
	gzhPath := filepath.Join(targetPath, ".gzh")
	content := fmt.Sprintf(`# GZH Repository Metadata
repository: %s
clone_url: %s
cloned_at: %s
provider: %s
`, repo.FullName, repo.CloneURL, time.Now().Format(time.RFC3339), e.options.Provider)

	return os.WriteFile(gzhPath, []byte(content), 0o644)
}

// printSummary prints the final operation summary.
func (e *CloneExecutor) printSummary(summary *CloneSummary) {
	e.progress.Info("\nClone Summary:")
	e.progress.Info("  Total:     %d", summary.Total)
	e.progress.Info("  Succeeded: %d", summary.Succeeded)
	e.progress.Info("  Failed:    %d", summary.Failed)
	e.progress.Info("  Skipped:   %d", summary.Skipped)
	e.progress.Info("  Duration:  %v", summary.Duration)

	if summary.Failed > 0 {
		e.progress.Info("\nFailed repositories:")
		for _, err := range summary.Errors {
			e.progress.Info("  %v", err)
		}
	}
}

// CloneResult represents the result of a single clone operation.
type CloneResult struct {
	Repository RepositoryInfo
	Success    bool
	Error      error
}

// CloneSummary contains summary information about a clone operation.
type CloneSummary struct {
	Total     int
	Succeeded int
	Failed    int
	Skipped   int
	Errors    []error
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
}
