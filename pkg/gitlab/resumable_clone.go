package gitlab

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Gizzahub/gzh-cli/internal/git"
	"github.com/Gizzahub/gzh-cli/internal/workerpool"
	synclonepkg "github.com/Gizzahub/gzh-cli/pkg/synclone"
)

// ResumableCloneManager handles resumable clone operations for GitLab.
type ResumableCloneManager struct {
	stateManager *synclonepkg.StateManager
	config       workerpool.RepositoryPoolConfig
}

// NewResumableCloneManager creates a new resumable clone manager for GitLab.
func NewResumableCloneManager(config workerpool.RepositoryPoolConfig) *ResumableCloneManager {
	return &ResumableCloneManager{
		stateManager: synclonepkg.NewStateManager(""),
		config:       config,
	}
}

// RefreshAllResumable performs bulk repository refresh with resumable support for GitLab.
func (rcm *ResumableCloneManager) RefreshAllResumable(ctx context.Context, targetPath, group, strategy string, parallel, maxRetries int, resume bool, progressMode string) error {
	// Initialize or load state
	// 상태파일을 타겟 디렉토리 하위에 저장하도록 상태 매니저 경로를 설정
	rcm.stateManager = synclonepkg.NewStateManager(filepath.Join(targetPath, ".gzh", "state"))
	state, err := rcm.initializeState(group, targetPath, strategy, parallel, maxRetries, resume)
	if err != nil {
		return err
	}

	// Get repositories to process
	allRepos, reposToProcess, err := rcm.determineRepositoriesToProcess(ctx, group, state, resume)
	if err != nil {
		return err
	}

	if len(allRepos) == 0 {
		fmt.Printf("No repositories found for group: %s\n", group)
		return nil
	}

	if len(reposToProcess) == 0 {
		return rcm.handleNoRemainingRepositories(state)
	}

	// Setup and execute clone operation
	return rcm.executeCloneOperation(ctx, state, allRepos, reposToProcess, targetPath, group, strategy, parallel, maxRetries, progressMode)
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

	case workerpool.OperationConfig:
		// Config operation - placeholder for configuration updates
		return fmt.Errorf("config operation not yet implemented")

	default:
		return fmt.Errorf("unknown operation: %s", job.Operation)
	}
}

// executeGitOperation executes a git command in the repository path with security validation.
func (rcm *ResumableCloneManager) executeGitOperation(ctx context.Context, repoPath string, args ...string) error {
	// Use secure git executor to prevent command injection
	executor, err := git.NewSecureGitExecutor()
	if err != nil {
		return fmt.Errorf("failed to create secure git executor: %w", err)
	}

	// Execute with validation
	if err := executor.ExecuteSecure(ctx, repoPath, args...); err != nil {
		return fmt.Errorf("secure git operation failed: %w", err)
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

// initializeState initializes or loads the clone state for resumable operations.
func (rcm *ResumableCloneManager) initializeState(group, targetPath, strategy string, parallel, maxRetries int, resume bool) (*synclonepkg.CloneState, error) {
	if resume {
		return rcm.loadAndValidateState(group, targetPath, strategy, parallel, maxRetries)
	}

	// Create new state
	state := synclonepkg.NewCloneState("gitlab", group, targetPath, strategy, parallel, maxRetries)

	// Check if there's already a state file
	if rcm.stateManager.HasState("gitlab", group) {
		return nil, fmt.Errorf("existing state found for %s. Use --resume to continue or delete state file", group)
	}

	return state, nil
}

// loadAndValidateState loads existing state and validates compatibility.
func (rcm *ResumableCloneManager) loadAndValidateState(group, targetPath, strategy string, parallel, maxRetries int) (*synclonepkg.CloneState, error) {
	state, err := rcm.stateManager.LoadState("gitlab", group)
	if err != nil {
		return nil, fmt.Errorf("failed to load state for resume: %w", err)
	}

	// Validate that the resume is compatible
	if state.TargetPath != targetPath {
		return nil, fmt.Errorf("target path mismatch: state has %s, requested %s", state.TargetPath, targetPath)
	}

	// Update configuration if different with warnings
	rcm.updateStateWithWarnings(state, strategy, parallel, maxRetries, group)

	return state, nil
}

// updateStateWithWarnings updates state configuration and shows warnings for changes.
func (rcm *ResumableCloneManager) updateStateWithWarnings(state *synclonepkg.CloneState, strategy string, parallel, maxRetries int, group string) {
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

	fmt.Printf("🔄 Resuming clone operation for %s (%.1f%% complete)\n", group, state.GetProgressPercent())
}

// determineRepositoriesToProcess determines which repositories need to be processed.
func (rcm *ResumableCloneManager) determineRepositoriesToProcess(ctx context.Context, group string, state *synclonepkg.CloneState, resume bool) ([]string, []string, error) {
	// Get all repositories from GitLab
	allRepos, err := List(ctx, group)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list repositories: %w", err)
	}

	if len(allRepos) == 0 {
		return allRepos, nil, nil
	}

	var reposToProcess []string
	if resume {
		reposToProcess = rcm.getResumableRepositories(state, allRepos)
	} else {
		// Process all repositories
		reposToProcess = allRepos
		state.SetPendingRepositories(reposToProcess)
	}

	return allRepos, reposToProcess, nil
}

// getResumableRepositories gets repositories that need to be processed during resume.
func (rcm *ResumableCloneManager) getResumableRepositories(state *synclonepkg.CloneState, allRepos []string) []string {
	// Use remaining repositories from state
	reposToProcess := state.GetRemainingRepositories()

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

	return reposToProcess
}

// handleNoRemainingRepositories handles the case where all repositories are already processed.
func (rcm *ResumableCloneManager) handleNoRemainingRepositories(state *synclonepkg.CloneState) error {
	fmt.Printf("✅ All repositories already processed\n")
	state.MarkCompleted()
	if err := rcm.stateManager.SaveState(state); err != nil {
		fmt.Printf("Warning: failed to save state: %v\n", err)
	}
	return nil
}

// executeCloneOperation executes the main clone operation with worker pool and progress tracking.
func (rcm *ResumableCloneManager) executeCloneOperation(ctx context.Context, state *synclonepkg.CloneState, allRepos, reposToProcess []string, targetPath, group, strategy string, parallel, maxRetries int, progressMode string) error {
	fmt.Printf("📦 Processing %d repositories (%d remaining)\n", len(allRepos), len(reposToProcess))

	// Save initial state
	if err := rcm.stateManager.SaveState(state); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	// Setup worker pool and progress tracking
	pool, progressTracker, err := rcm.setupWorkerPoolAndProgress(allRepos, reposToProcess, targetPath, strategy, parallel, maxRetries, progressMode, state) //nolint:contextcheck // Setup function manages context internally
	if err != nil {
		return err
	}
	defer pool.Stop()

	// Process repositories and track results
	return rcm.processRepositoriesWithTracking(ctx, pool, reposToProcess, targetPath, strategy, group, state, progressTracker)
}

// setupWorkerPoolAndProgress sets up the worker pool and progress tracker.
func (rcm *ResumableCloneManager) setupWorkerPoolAndProgress(allRepos, reposToProcess []string, targetPath, strategy string, parallel, maxRetries int, progressMode string, state *synclonepkg.CloneState) (*workerpool.RepositoryWorkerPool, *synclonepkg.ProgressTracker, error) {
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
		return nil, nil, fmt.Errorf("failed to start worker pool: %w", err)
	}

	// Create progress tracker with specified mode
	displayMode := getDisplayMode(progressMode)
	progressTracker := synclonepkg.NewProgressTracker(allRepos, displayMode)

	// Update progress tracker with existing state
	rcm.updateProgressTrackerWithState(progressTracker, state)

	// Print initial progress
	fmt.Printf("\r%-120s", progressTracker.RenderProgress())

	return pool, progressTracker, nil
}

// updateProgressTrackerWithState updates progress tracker with existing state data.
func (rcm *ResumableCloneManager) updateProgressTrackerWithState(progressTracker *synclonepkg.ProgressTracker, state *synclonepkg.CloneState) {
	for _, completed := range state.CompletedRepos {
		progressTracker.CompleteRepository(completed.Name, completed.Message)
	}
	for _, failed := range state.FailedRepos {
		progressTracker.SetRepositoryError(failed.Name, failed.Error)
	}
}

// processRepositoriesWithTracking processes all repositories with progress tracking and state management.
func (rcm *ResumableCloneManager) processRepositoriesWithTracking(ctx context.Context, pool *workerpool.RepositoryWorkerPool, reposToProcess []string, targetPath, strategy, group string, state *synclonepkg.CloneState, progressTracker *synclonepkg.ProgressTracker) error {
	// Create and submit jobs
	jobs := rcm.createRepositoryJobs(reposToProcess, targetPath, strategy)

	// Start progress tracking for pending jobs
	for _, job := range jobs {
		progressTracker.UpdateRepository(job.Repository, getProgressStatusFromOperation(job.Operation), "Starting...", 0.0)
	}

	// Submit jobs to worker pool
	if err := rcm.submitJobs(pool, jobs, group); err != nil {
		return err
	}

	// Track results with periodic state saving
	_, failureCount, err := rcm.trackResultsWithPeriodicSaving(ctx, pool, jobs, state, progressTracker)
	if err != nil {
		return err
	}

	// Finalize operation
	return rcm.finalizeOperation(state, progressTracker, group, failureCount)
}

// createRepositoryJobs creates worker pool jobs for repositories to process.
func (rcm *ResumableCloneManager) createRepositoryJobs(reposToProcess []string, targetPath, strategy string) []workerpool.RepositoryJob {
	jobs := make([]workerpool.RepositoryJob, 0, len(reposToProcess))
	for _, repo := range reposToProcess {
		repoPath := filepath.Join(targetPath, repo)
		operation := rcm.determineOperation(repoPath, strategy)

		jobs = append(jobs, workerpool.RepositoryJob{
			Repository: repo,
			Operation:  operation,
			Path:       repoPath,
			Strategy:   strategy,
		})
	}
	return jobs
}

// determineOperation determines the operation type based on repository existence and strategy.
func (rcm *ResumableCloneManager) determineOperation(repoPath, strategy string) workerpool.RepositoryOperation {
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		return workerpool.OperationClone
	}

	switch strategy {
	case "reset":
		return workerpool.OperationReset
	case "pull":
		return workerpool.OperationPull
	case "fetch":
		return workerpool.OperationFetch
	default:
		return workerpool.OperationPull
	}
}

// submitJobs submits all jobs to the worker pool.
func (rcm *ResumableCloneManager) submitJobs(pool *workerpool.RepositoryWorkerPool, jobs []workerpool.RepositoryJob, group string) error {
	processFn := func(ctx context.Context, job workerpool.RepositoryJob) error {
		return rcm.processRepositoryJob(ctx, job, group)
	}

	for _, job := range jobs {
		if err := pool.SubmitJob(job, processFn); err != nil {
			return fmt.Errorf("failed to submit job for %s: %w", job.Repository, err)
		}
	}
	return nil
}

// trackResultsWithPeriodicSaving tracks job results with periodic state saving and progress updates.
func (rcm *ResumableCloneManager) trackResultsWithPeriodicSaving(ctx context.Context, pool *workerpool.RepositoryWorkerPool, jobs []workerpool.RepositoryJob, state *synclonepkg.CloneState, progressTracker *synclonepkg.ProgressTracker) (int, int, error) {
	resultsChan := pool.Results()
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
			rcm.handleJobResult(result, state, progressTracker, &successCount, &failureCount)
		case <-progressUpdateTicker.C:
			// Windows PowerShell에서 진행률 표시 - 줄바꿈 없이 덮어쓰기
			fmt.Printf("\r%-120s", progressTracker.RenderProgress())
		case <-stateSaveTicker.C:
			if err := rcm.stateManager.SaveState(state); err != nil {
				fmt.Printf("\n⚠️  Warning: failed to save state: %v\n", err)
			}
		case <-ctx.Done():
			state.MarkCancelled()
			if err := rcm.stateManager.SaveState(state); err != nil {
				fmt.Printf("Warning: failed to save state: %v\n", err)
			}
			return successCount, failureCount, fmt.Errorf("operation cancelled: %w", ctx.Err())
		}
	}

	return successCount, failureCount, nil
}

// handleJobResult handles individual job results and updates state and progress.
func (rcm *ResumableCloneManager) handleJobResult(result workerpool.RepositoryResult, state *synclonepkg.CloneState, progressTracker *synclonepkg.ProgressTracker, successCount, failureCount *int) {
	if result.Error != nil {
		*failureCount++
		state.AddFailedRepository(result.Job.Repository, result.Job.Path, string(result.Job.Operation), result.Error.Error(), 1)
		progressTracker.SetRepositoryError(result.Job.Repository, result.Error.Error())
	} else {
		*successCount++
		state.AddCompletedRepository(result.Job.Repository, result.Job.Path, string(result.Job.Operation), result.Message)
		progressTracker.CompleteRepository(result.Job.Repository, result.Message)
	}
}

// finalizeOperation finalizes the clone operation with final state updates and cleanup.
func (rcm *ResumableCloneManager) finalizeOperation(state *synclonepkg.CloneState, progressTracker *synclonepkg.ProgressTracker, group string, failureCount int) error {
	// Final progress update - 마지막 줄 정리
	fmt.Print("\r\033[2K") // 전체 줄 지우기
	fmt.Printf("%s\n", progressTracker.RenderProgress())

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
	if state.Status == "completed" && failureCount == 0 {
		if err := rcm.stateManager.DeleteState("gitlab", group); err != nil {
			fmt.Printf("Warning: failed to delete state: %v\n", err)
		}
		fmt.Printf("✅ Clone operation completed successfully\n")
	} else if failureCount > 0 {
		fmt.Printf("❌ Clone operation completed with %d failures\n", failureCount)
	} else {
		fmt.Printf("⚠️  Clone operation incomplete. Use --resume to continue\n")
	}

	if failureCount > 0 {
		return fmt.Errorf("%d operations failed", failureCount)
	}

	return nil
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

// GetCloneState returns the current clone state for a group.
func GetCloneState(group string) (*synclonepkg.CloneState, error) {
	stateManager := synclonepkg.NewStateManager("")
	return stateManager.LoadState("gitlab", group)
}

// DeleteCloneState removes the state file for a group.
func DeleteCloneState(group string) error {
	stateManager := synclonepkg.NewStateManager("")
	return stateManager.DeleteState("gitlab", group)
}

// ListCloneStates returns all saved clone states.
func ListCloneStates() ([]synclonepkg.CloneState, error) {
	stateManager := synclonepkg.NewStateManager("")
	return stateManager.ListStates()
}
