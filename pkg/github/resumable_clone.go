package github

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/gizzahub/gzh-manager-go/internal/workerpool"
	synclonepkg "github.com/gizzahub/gzh-manager-go/pkg/synclone"
)

// ResumableCloneManager handles resumable clone operations.
type ResumableCloneManager struct {
	stateManager *synclonepkg.StateManager
	config       BulkOperationsConfig
}

// NewResumableCloneManager creates a new resumable clone manager.
func NewResumableCloneManager(config BulkOperationsConfig) *ResumableCloneManager {
	return &ResumableCloneManager{
		stateManager: synclonepkg.NewStateManager(""),
		config:       config,
	}
}

// RefreshAllResumable performs bulk repository refresh with resumable support.
func (rcm *ResumableCloneManager) RefreshAllResumable(ctx context.Context, targetPath, org, strategy string, parallel, maxRetries int, resume bool, progressMode string) error {
	// Initialize or load state
	state, err := rcm.initializeOrLoadState(org, targetPath, strategy, parallel, maxRetries, resume)
	if err != nil {
		return err
	}

	// Get repositories and determine processing list
	allRepos, reposToProcess, err := rcm.prepareRepositoryList(ctx, org, state, resume)
	if err != nil {
		return err
	}

	if len(reposToProcess) == 0 {
		fmt.Printf("✅ All repositories already processed\n")
		state.MarkCompleted()
		_ = rcm.stateManager.SaveState(state) //nolint:errcheck // State save is best effort
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
		config.PoolConfig.CloneWorkers = parallel
		config.PoolConfig.UpdateWorkers = parallel + (parallel / 2)

		config.PoolConfig.ConfigWorkers = parallel / 2
		if config.PoolConfig.ConfigWorkers < 1 {
			config.PoolConfig.ConfigWorkers = 1
		}
	}

	if maxRetries > 0 {
		config.PoolConfig.RetryAttempts = maxRetries
	}

	// Create and start worker pool
	pool := workerpool.NewRepositoryWorkerPool(config.PoolConfig)
	if err := pool.Start(); err != nil { //nolint:contextcheck // Worker pool start manages its own context
		return fmt.Errorf("failed to start worker pool: %w", err)
	}
	defer pool.Stop()

	// Create progress tracker with specified mode
	displayMode := getDisplayMode(progressMode)
	progressTracker := synclonepkg.NewProgressTracker(allRepos, displayMode)

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
		return processRepositoryJob(ctx, job, org)
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
			if err := rcm.stateManager.SaveState(state); err != nil {
				fmt.Printf("\n⚠️  Warning: failed to save state: %v\n", err)
			}

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
		_ = rcm.stateManager.DeleteState("github", org)
		fmt.Printf("✅ Clone operation completed successfully\n")
	} else {
		fmt.Printf("⚠️  Clone operation incomplete. Use --resume to continue\n")
	}

	if failureCount > 0 {
		return fmt.Errorf("%d operations failed", failureCount)
	}

	return nil
}

// initializeOrLoadState handles state initialization and loading for resume operations.
func (rcm *ResumableCloneManager) initializeOrLoadState(org, targetPath, strategy string, parallel, maxRetries int, resume bool) (*synclonepkg.CloneState, error) {
	if resume {
		return rcm.loadAndValidateState(org, targetPath, strategy, parallel, maxRetries)
	}
	return rcm.createNewState(org, targetPath, strategy, parallel, maxRetries)
}

// loadAndValidateState loads existing state and validates compatibility.
func (rcm *ResumableCloneManager) loadAndValidateState(org, targetPath, strategy string, parallel, maxRetries int) (*synclonepkg.CloneState, error) {
	state, err := rcm.stateManager.LoadState("github", org)
	if err != nil {
		return nil, fmt.Errorf("failed to load state for resume: %w", err)
	}

	// Validate that the resume is compatible
	if state.TargetPath != targetPath {
		return nil, fmt.Errorf("target path mismatch: state has %s, requested %s", state.TargetPath, targetPath)
	}

	// Update configuration if different with warnings
	rcm.updateStateWithWarnings(state, strategy, parallel, maxRetries)

	fmt.Printf("🔄 Resuming clone operation for %s (%.1f%% complete)\n", org, state.GetProgressPercent())
	return state, nil
}

// createNewState creates a new clone state and validates no existing state exists.
func (rcm *ResumableCloneManager) createNewState(org, targetPath, strategy string, parallel, maxRetries int) (*synclonepkg.CloneState, error) {
	// Check if there's already a state file
	if rcm.stateManager.HasState("github", org) {
		// Load existing state to check if it's for the same target path
		existingState, err := rcm.stateManager.LoadState("github", org)
		if err == nil && existingState.TargetPath == targetPath {
			// If it's the same target path, suggest using --resume
			return nil, fmt.Errorf("existing state found for %s at %s. Use --resume to continue or 'gz synclone state clean --all' to start fresh", org, targetPath)
		}
		// Different target path, clean up old state
		_ = rcm.stateManager.DeleteState("github", org)
	}

	// Create new state
	return synclonepkg.NewCloneState("github", org, targetPath, strategy, parallel, maxRetries), nil
}

// updateStateWithWarnings updates state configuration and prints warnings for changes.
func (rcm *ResumableCloneManager) updateStateWithWarnings(state *synclonepkg.CloneState, strategy string, parallel, maxRetries int) {
	if state.Strategy != strategy {
		fmt.Printf("⚠️  Warning: strategy changed from %s to %s\n", state.Strategy, strategy)
		state.Strategy = strategy
	}

	if state.Parallel != parallel {
		fmt.Printf("⚠️  Warning: parallel count changed from %d to %d\n", state.Parallel, parallel)
		state.Parallel = parallel
	}

	if state.MaxRetries != maxRetries {
		fmt.Printf("⚠️  Warning: max retries changed from %d to %d\n", state.MaxRetries, maxRetries)
		state.MaxRetries = maxRetries
	}
}

// prepareRepositoryList gets all repositories and determines which need processing.
func (rcm *ResumableCloneManager) prepareRepositoryList(ctx context.Context, org string, state *synclonepkg.CloneState, resume bool) ([]string, []string, error) {
	// Get all repositories from GitHub
	allRepos, err := List(ctx, org)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list repositories: %w", err)
	}

	if len(allRepos) == 0 {
		fmt.Printf("No repositories found for organization: %s\n", org)
		return allRepos, []string{}, nil
	}

	// Determine which repositories need to be processed
	var reposToProcess []string
	if resume {
		reposToProcess = rcm.getResumeRepositories(state, allRepos)
	} else {
		reposToProcess = allRepos
		state.SetPendingRepositories(reposToProcess)
	}

	return allRepos, reposToProcess, nil
}

// getResumeRepositories determines which repositories to process on resume.
func (rcm *ResumableCloneManager) getResumeRepositories(state *synclonepkg.CloneState, allRepos []string) []string {
	// Use remaining repositories from state
	reposToProcess := state.GetRemainingRepositories()

	// Check if any new repositories were added since last run
	for _, repo := range allRepos {
		if !state.IsCompleted(repo) && !state.IsFailed(repo) && !rcm.containsString(reposToProcess, repo) {
			reposToProcess = append(reposToProcess, repo)
		}
	}

	return reposToProcess
}

// containsString checks if a string slice contains a specific string.
func (rcm *ResumableCloneManager) containsString(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}

// processRepositoryJob processes a single repository job for GitHub.
func processRepositoryJob(ctx context.Context, job workerpool.RepositoryJob, org string) error {
	switch job.Operation {
	case workerpool.OperationClone:
		return Clone(ctx, job.Path, org, job.Repository)

	case workerpool.OperationPull:
		return executeGitOperation(ctx, job.Path, "pull")

	case workerpool.OperationFetch:
		return executeGitOperation(ctx, job.Path, "fetch")

	case workerpool.OperationReset:
		// Reset hard HEAD and pull
		if err := executeGitOperation(ctx, job.Path, "reset", "--hard", "HEAD"); err != nil {
			return fmt.Errorf("git reset failed: %w", err)
		}

		return executeGitOperation(ctx, job.Path, "pull")

	case workerpool.OperationConfig:
		// Config operation - placeholder for configuration updates
		return fmt.Errorf("config operation not yet implemented")

	default:
		return fmt.Errorf("unknown operation: %s", job.Operation)
	}
}

// executeGitOperation executes a git command in the repository path.
func executeGitOperation(ctx context.Context, repoPath string, args ...string) error {
	// Build git command
	gitArgs := append([]string{"-C", repoPath}, args...)
	cmd := exec.CommandContext(ctx, "git", gitArgs...)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git %s failed: %w", args[0], err)
	}

	return nil
}

// RefreshAllResumable is a convenience function for resumable cloning.
func RefreshAllResumable(ctx context.Context, targetPath, org, strategy string, parallel, maxRetries int, resume bool, progressMode string) error {
	config := DefaultBulkOperationsConfig()
	manager := NewResumableCloneManager(config)

	return manager.RefreshAllResumable(ctx, targetPath, org, strategy, parallel, maxRetries, resume, progressMode)
}

// GetCloneState returns the current clone state for an organization.
func GetCloneState(org string) (*synclonepkg.CloneState, error) {
	stateManager := synclonepkg.NewStateManager("")
	return stateManager.LoadState("github", org)
}

// DeleteCloneState removes the state file for an organization.
func DeleteCloneState(org string) error {
	stateManager := synclonepkg.NewStateManager("")
	return stateManager.DeleteState("github", org)
}

// ListCloneStates returns all saved clone states.
func ListCloneStates() ([]synclonepkg.CloneState, error) {
	stateManager := synclonepkg.NewStateManager("")
	return stateManager.ListStates()
}

// getProgressStatusFromOperation converts worker pool operation to progress status.
func getProgressStatusFromOperation(operation workerpool.RepositoryOperation) synclonepkg.ProgressStatus {
	switch operation {
	case workerpool.OperationClone:
		return synclonepkg.StatusCloning
	case workerpool.OperationPull:
		return synclonepkg.StatusPulling
	case workerpool.OperationFetch:
		return synclonepkg.StatusFetching
	case workerpool.OperationReset:
		return synclonepkg.StatusResetting
	case workerpool.OperationConfig:
		return synclonepkg.StatusStarted // Config operations are quick, just show as started
	default:
		return synclonepkg.StatusStarted
	}
}

// getDisplayMode converts string to DisplayMode.
func getDisplayMode(mode string) synclonepkg.DisplayMode {
	switch mode {
	case "compact":
		return synclonepkg.DisplayModeCompact
	case "detailed":
		return synclonepkg.DisplayModeDetailed
	case "quiet":
		return synclonepkg.DisplayModeQuiet
	default:
		return synclonepkg.DisplayModeCompact
	}
}
