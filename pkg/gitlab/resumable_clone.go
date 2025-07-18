package gitlab

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/gizzahub/gzh-manager-go/internal/workerpool"
	bulkclonepkg "github.com/gizzahub/gzh-manager-go/pkg/bulk-clone"
)

// ResumableCloneManager handles resumable clone operations for GitLab.
type ResumableCloneManager struct {
	stateManager *bulkclonepkg.StateManager
	config       workerpool.RepositoryPoolConfig
}

// NewResumableCloneManager creates a new resumable clone manager for GitLab.
func NewResumableCloneManager(config workerpool.RepositoryPoolConfig) *ResumableCloneManager {
	return &ResumableCloneManager{
		stateManager: bulkclonepkg.NewStateManager(""),
		config:       config,
	}
}

// RefreshAllResumable performs bulk repository refresh with resumable support for GitLab.
func (rcm *ResumableCloneManager) RefreshAllResumable(ctx context.Context, targetPath, group, strategy string, parallel, maxRetries int, resume bool, progressMode string) error {
	var (
		state *bulkclonepkg.CloneState
		err   error
	)

	// Load existing state if resuming

	if resume {
		state, err = rcm.stateManager.LoadState("gitlab", group)
		if err != nil {
			return fmt.Errorf("failed to load state for resume: %w", err)
		}

		// Validate that the resume is compatible
		if state.TargetPath != targetPath {
			return fmt.Errorf("target path mismatch: state has %s, requested %s", state.TargetPath, targetPath)
		}

		if state.Strategy != strategy {
			fmt.Printf("⚠️  Warning: strategy changed from %s to %s\n", state.Strategy, strategy)
			state.Strategy = strategy
		}

		// Update configuration if different
		if state.Parallel != parallel {
			fmt.Printf("⚠️  Warning: parallel count changed from %d to %d\n", state.Parallel, parallel)
			state.Parallel = parallel
		}

		if state.MaxRetries != maxRetries {
			fmt.Printf("⚠️  Warning: max retries changed from %d to %d\n", state.MaxRetries, maxRetries)
			state.MaxRetries = maxRetries
		}

		fmt.Printf("🔄 Resuming clone operation for %s (%.1f%% complete)\n", group, state.GetProgressPercent())
	} else {
		// Create new state
		state = bulkclonepkg.NewCloneState("gitlab", group, targetPath, strategy, parallel, maxRetries)

		// Check if there's already a state file
		if rcm.stateManager.HasState("gitlab", group) {
			return fmt.Errorf("existing state found for %s. Use --resume to continue or delete state file", group)
		}
	}

	// Get all repositories from GitLab
	allRepos, err := List(ctx, group)
	if err != nil {
		return fmt.Errorf("failed to list repositories: %w", err)
	}

	if len(allRepos) == 0 {
		fmt.Printf("No repositories found for group: %s\n", group)
		return nil
	}

	// Determine which repositories need to be processed
	var reposToProcess []string
	if resume {
		// Use remaining repositories from state
		reposToProcess = state.GetRemainingRepositories()

		// Also check if any new repositories were added since last run
		for _, repo := range allRepos {
			if !state.IsCompleted(repo) && !state.IsFailed(repo) {
				found := false

				for _, pending := range reposToProcess {
					if pending == repo {
						found = true
						break
					}
				}

				if !found {
					reposToProcess = append(reposToProcess, repo)
				}
			}
		}
	} else {
		// Process all repositories
		reposToProcess = allRepos
		state.SetPendingRepositories(reposToProcess)
	}

	if len(reposToProcess) == 0 {
		fmt.Printf("✅ All repositories already processed\n")
		state.MarkCompleted()
		rcm.stateManager.SaveState(state)

		return nil
	}

	fmt.Printf("📦 Processing %d repositories (%d remaining)\n", len(allRepos), len(reposToProcess))

	// Save initial state
	if err := rcm.stateManager.SaveState(state); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	// Configure worker pool
	config := rcm.config
	if parallel > 0 {
		config.CloneWorkers = parallel
		config.UpdateWorkers = parallel + (parallel / 2)

		config.ConfigWorkers = parallel / 2
		if config.ConfigWorkers < 1 {
			config.ConfigWorkers = 1
		}
	}

	if maxRetries > 0 {
		config.RetryAttempts = maxRetries
	}

	// Create and start worker pool
	pool := workerpool.NewRepositoryWorkerPool(config)
	if err := pool.Start(); err != nil {
		return fmt.Errorf("failed to start worker pool: %w", err)
	}
	defer pool.Stop()

	// Create progress tracker with specified mode
	displayMode := getDisplayMode(progressMode)
	progressTracker := bulkclonepkg.NewProgressTracker(allRepos, displayMode)

	// Update progress tracker with existing state
	for _, completed := range state.CompletedRepos {
		progressTracker.CompleteRepository(completed.Name, completed.Message)
	}

	for _, failed := range state.FailedRepos {
		progressTracker.SetRepositoryError(failed.Name, failed.Error)
	}

	// Print initial progress
	fmt.Printf("\n%s\n", progressTracker.RenderProgress())

	// Create jobs for repositories to process
	jobs := make([]workerpool.RepositoryJob, 0, len(reposToProcess))
	for _, repo := range reposToProcess {
		repoPath := filepath.Join(targetPath, repo)

		// Determine operation type
		var operation workerpool.RepositoryOperation
		if _, err := os.Stat(repoPath); os.IsNotExist(err) {
			operation = workerpool.OperationClone
		} else {
			switch strategy {
			case "reset":
				operation = workerpool.OperationReset
			case "pull":
				operation = workerpool.OperationPull
			case "fetch":
				operation = workerpool.OperationFetch
			default:
				operation = workerpool.OperationPull
			}
		}

		jobs = append(jobs, workerpool.RepositoryJob{
			Repository: repo,
			Operation:  operation,
			Path:       repoPath,
			Strategy:   strategy,
		})
	}

	// Start progress tracking for pending jobs
	for _, job := range jobs {
		progressTracker.UpdateRepository(job.Repository, getProgressStatusFromOperation(job.Operation), "Starting...", 0.0)
	}

	// Process jobs using the correct pattern
	processFn := func(ctx context.Context, job workerpool.RepositoryJob) error {
		return rcm.processRepositoryJob(ctx, job, group)
	}

	// Submit jobs and collect results
	resultsChan := pool.Results()

	// Submit all jobs
	for _, job := range jobs {
		if err := pool.SubmitJob(job, processFn); err != nil {
			return fmt.Errorf("failed to submit job for %s: %w", job.Repository, err)
		}
	}

	// Track results and update state
	successCount := 0
	failureCount := 0

	// Set up periodic state saving and progress updates
	stateSaveTicker := time.NewTicker(30 * time.Second)
	progressUpdateTicker := time.NewTicker(1 * time.Second)

	defer stateSaveTicker.Stop()
	defer progressUpdateTicker.Stop()

	for i := 0; i < len(jobs); i++ {
		select {
		case result := <-resultsChan:
			if result.Error != nil {
				failureCount++

				state.AddFailedRepository(result.Job.Repository, result.Job.Path, string(result.Job.Operation), result.Error.Error(), 1)
				progressTracker.SetRepositoryError(result.Job.Repository, result.Error.Error())
			} else {
				successCount++

				state.AddCompletedRepository(result.Job.Repository, result.Job.Path, string(result.Job.Operation), result.Message)
				progressTracker.CompleteRepository(result.Job.Repository, result.Message)
			}

		case <-progressUpdateTicker.C:
			// Update progress display
			fmt.Printf("\r\033[K%s", progressTracker.RenderProgress())

		case <-stateSaveTicker.C:
			// Periodically save state
			if err := rcm.stateManager.SaveState(state); err != nil {
				fmt.Printf("\n⚠️  Warning: failed to save state: %v\n", err)
			}

		case <-ctx.Done():
			// Operation cancelled
			state.MarkCancelled()
			rcm.stateManager.SaveState(state)

			return fmt.Errorf("operation cancelled: %w", ctx.Err())
		}
	}

	// Final progress update
	fmt.Printf("\r\033[K%s\n", progressTracker.RenderProgress())

	// Final state update
	if len(state.GetRemainingRepositories()) == 0 {
		state.MarkCompleted()
	} else {
		state.MarkFailed()
	}

	// Save final state
	if err := rcm.stateManager.SaveState(state); err != nil {
		fmt.Printf("⚠️  Warning: failed to save final state: %v\n", err)
	}

	// Show final summary
	fmt.Printf("\n%s\n", progressTracker.GetSummary())

	// Clean up state file if completed successfully
	if state.Status == "completed" {
		rcm.stateManager.DeleteState("gitlab", group)
		fmt.Printf("✅ Clone operation completed successfully\n")
	} else {
		fmt.Printf("⚠️  Clone operation incomplete. Use --resume to continue\n")
	}

	if failureCount > 0 {
		return fmt.Errorf("%d operations failed", failureCount)
	}

	return nil
}

// processRepositoryJob processes a single repository job for GitLab.
func (rcm *ResumableCloneManager) processRepositoryJob(ctx context.Context, job workerpool.RepositoryJob, group string) error {
	switch job.Operation {
	case workerpool.OperationClone:
		return Clone(ctx, job.Path, group, job.Repository, "")

	case workerpool.OperationPull:
		return rcm.executeGitOperation(ctx, job.Path, "pull")

	case workerpool.OperationFetch:
		return rcm.executeGitOperation(ctx, job.Path, "fetch")

	case workerpool.OperationReset:
		// Reset hard HEAD and pull
		if err := rcm.executeGitOperation(ctx, job.Path, "reset", "--hard", "HEAD"); err != nil {
			return fmt.Errorf("git reset failed: %w", err)
		}

		return rcm.executeGitOperation(ctx, job.Path, "pull")

	default:
		return fmt.Errorf("unknown operation: %s", job.Operation)
	}
}

// executeGitOperation executes a git command in the repository path.
func (rcm *ResumableCloneManager) executeGitOperation(ctx context.Context, repoPath string, args ...string) error {
	// Build git command
	gitArgs := append([]string{"-C", repoPath}, args...)
	cmd := exec.CommandContext(ctx, "git", gitArgs...)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git %s failed: %w", args[0], err)
	}

	return nil
}

// RefreshAllResumable is a convenience function for resumable cloning.
func RefreshAllResumable(ctx context.Context, targetPath, group, strategy string, parallel, maxRetries int, resume bool, progressMode string) error {
	config := workerpool.DefaultRepositoryPoolConfig()
	manager := NewResumableCloneManager(config)

	return manager.RefreshAllResumable(ctx, targetPath, group, strategy, parallel, maxRetries, resume, progressMode)
}

// getProgressStatusFromOperation converts worker pool operation to progress status.
func getProgressStatusFromOperation(operation workerpool.RepositoryOperation) bulkclonepkg.ProgressStatus {
	switch operation {
	case workerpool.OperationClone:
		return bulkclonepkg.StatusCloning
	case workerpool.OperationPull:
		return bulkclonepkg.StatusPulling
	case workerpool.OperationFetch:
		return bulkclonepkg.StatusFetching
	case workerpool.OperationReset:
		return bulkclonepkg.StatusResetting
	default:
		return bulkclonepkg.StatusStarted
	}
}

// getDisplayMode converts string to DisplayMode.
func getDisplayMode(mode string) bulkclonepkg.DisplayMode {
	switch mode {
	case "compact":
		return bulkclonepkg.DisplayModeCompact
	case "detailed":
		return bulkclonepkg.DisplayModeDetailed
	case "quiet":
		return bulkclonepkg.DisplayModeQuiet
	default:
		return bulkclonepkg.DisplayModeCompact
	}
}

// GetCloneState returns the current clone state for a group.
func GetCloneState(group string) (*bulkclonepkg.CloneState, error) {
	stateManager := bulkclonepkg.NewStateManager("")
	return stateManager.LoadState("gitlab", group)
}

// DeleteCloneState removes the state file for a group.
func DeleteCloneState(group string) error {
	stateManager := bulkclonepkg.NewStateManager("")
	return stateManager.DeleteState("gitlab", group)
}

// ListCloneStates returns all saved clone states.
func ListCloneStates() ([]bulkclonepkg.CloneState, error) {
	stateManager := bulkclonepkg.NewStateManager("")
	return stateManager.ListStates()
}
