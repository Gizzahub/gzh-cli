package github

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/gizzahub/gzh-cli/internal/workerpool"
	synclonepkg "github.com/gizzahub/gzh-cli/pkg/synclone"
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
	// 상태파일을 타겟 디렉토리 하위에 저장하도록 상태 매니저 경로를 설정
	rcm.stateManager = synclonepkg.NewStateManager(filepath.Join(targetPath, ".gzh", "state"))
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
		if ctx.Err() != nil {
			return rcm.saveCancellation(state, ctx.Err())
		}
		return rcm.finalizeOperation(state, org)
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

		config.PoolConfig.ConfigWorkers = max(parallel/2, 1)
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

	// Show initial progress (0/total or existing progress if resuming)
	fmt.Printf("\r\033[K%s", progressTracker.RenderProgress())

	// Update progress tracker with existing state
	for _, completed := range state.CompletedRepos {
		progressTracker.CompleteRepository(completed.Name, completed.Message)
	}

	for _, failed := range state.FailedRepos {
		progressTracker.SetRepositoryError(failed.Name, failed.Error)
	}

	// Show updated progress after loading existing state
	fmt.Printf("\r\033[K%s", progressTracker.RenderProgress())

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

	// Start progress tracking for pending jobs (don't trigger initial display)
	for _, job := range jobs {
		progressTracker.UpdateRepository(job.Repository, getProgressStatusFromOperation(job.Operation), "Starting...", 0.0)
	}

	// Process jobs using the correct pattern
	processFn := func(ctx context.Context, job workerpool.RepositoryJob) error {
		return processRepositoryJob(ctx, job, org)
	}

	// 큐가 차면 Submit은 즉시 거절한다. 결과 하나를 받은 뒤에 같은 작업을 다시 넣는다.
	if err := pool.SubmitDraining(ctx, jobs, processFn, workerpool.DrainConfig{
		ProgressInterval: 2 * time.Second,
		SaveInterval:     30 * time.Second,
		OnResult: func(result workerpool.RepositoryResult) {
			if result.Error != nil {
				state.AddFailedRepository(result.Job.Repository, result.Job.Path, string(result.Job.Operation), result.Error.Error(), 1)
				progressTracker.SetRepositoryError(result.Job.Repository, result.Error.Error())
				return
			}
			state.AddCompletedRepository(result.Job.Repository, result.Job.Path, string(result.Job.Operation), result.Message)
			progressTracker.CompleteRepository(result.Job.Repository, result.Message)
		},
		OnProgress: func() {
			fmt.Printf("\r\033[K%s", progressTracker.RenderProgress())
		},
		OnSave: func() {
			if saveErr := rcm.stateManager.SaveState(state); saveErr != nil {
				fmt.Printf("\n⚠️  Warning: failed to save state: %v\n", saveErr)
			}
		},
	}); err != nil {
		if ctx.Err() != nil {
			return rcm.saveCancellation(state, ctx.Err())
		}
		if errors.Is(err, workerpool.ErrQueueFull) {
			if saveErr := rcm.stateManager.SaveState(state); saveErr != nil {
				return errors.Join(err, fmt.Errorf("failed to save state: %w", saveErr))
			}
		}
		return err
	}
	// Final progress update
	fmt.Printf("\r\033[K%s\n", progressTracker.RenderProgress())
	if ctx.Err() != nil {
		return rcm.saveCancellation(state, ctx.Err())
	}
	return rcm.finalizeOperation(state, org)
}

// 완료는 미처리 항목과 보존된 실패가 모두 없을 때만 확정한다.
func (rcm *ResumableCloneManager) finalizeOperation(state *synclonepkg.CloneState, org string) error {
	var operationErr error
	if len(state.FailedRepos) > 0 {
		failures := make([]error, 0, len(state.FailedRepos)+1)
		failures = append(failures, fmt.Errorf("%d operations failed", len(state.FailedRepos)))
		for _, failed := range state.FailedRepos {
			failures = append(failures, fmt.Errorf("repository %s: %s", failed.Name, failed.Error))
		}
		operationErr = errors.Join(failures...)
	} else if len(state.GetRemainingRepositories()) > 0 {
		operationErr = fmt.Errorf("clone operation incomplete: %d repositories pending", len(state.GetRemainingRepositories()))
	}
	if len(state.GetRemainingRepositories()) == 0 && len(state.FailedRepos) == 0 {
		state.MarkCompleted()
	} else {
		state.MarkFailed()
	}

	// Save final state
	if err := rcm.stateManager.SaveState(state); err != nil {
		return errors.Join(operationErr, fmt.Errorf("failed to save final state: %w", err))
	}

	// Skip detailed summary to avoid duplication

	// Clean up state file if completed successfully
	if state.Status == "completed" {
		if err := rcm.stateManager.DeleteState("github", org); err != nil {
			return fmt.Errorf("failed to delete completed state: %w", err)
		}
		fmt.Printf("✅ Clone operation completed successfully\n")
	} else if len(state.GetRemainingRepositories()) > 0 {
		fmt.Printf("⚠️  Clone operation incomplete. Use --resume to continue\n")
	} else {
		fmt.Printf("⚠️  Recorded failures remain. GitHub --resume does not retry failed repositories; review and back up the saved snapshot and resolve failures before starting fresh.\n")
	}

	if len(state.FailedRepos) > 0 {
		// Print detailed failure information
		rcm.printFailureDetails(state)
	}

	return operationErr
}

// 취소 원인과 상태 저장 실패를 함께 반환하여 복구 정보 유실을 드러낸다.
func (rcm *ResumableCloneManager) saveCancellation(state *synclonepkg.CloneState, cause error) error {
	state.MarkCancelled()
	if err := rcm.stateManager.SaveState(state); err != nil {
		return errors.Join(fmt.Errorf("operation cancelled: %w", cause), fmt.Errorf("failed to save cancelled state: %w", err))
	}
	return fmt.Errorf("operation cancelled: %w", cause)
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
		if err != nil {
			return nil, fmt.Errorf("failed to load existing state: %w", err)
		}
		if existingState.TargetPath == targetPath {
			// If it's the same target path, suggest using --resume
			path, err := rcm.stateManager.ResolveStateFilePath("github", org)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve existing state path: %w", err)
			}
			return nil, fmt.Errorf("existing state found for %s at %s. Use --resume to continue; review and back up the existing snapshot at %s. To start fresh after resolving failures, move this snapshot to a backup location to preserve recovery data", org, targetPath, path)
		}
		// Different target path, clean up old state
		if err := rcm.stateManager.DeleteState("github", org); err != nil {
			return nil, fmt.Errorf("failed to remove previous state: %w", err)
		}
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
	}

	// Determine which repositories need to be processed
	var reposToProcess []string
	if resume {
		reposToProcess = rcm.getResumeRepositories(state, allRepos)
		state.SetPendingRepositories(reposToProcess)
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
	return slices.Contains(slice, str)
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

// printFailureDetails prints detailed information about failed repositories.
func (rcm *ResumableCloneManager) printFailureDetails(state *synclonepkg.CloneState) {
	failedRepos := state.FailedRepos
	if len(failedRepos) == 0 {
		return
	}

	fmt.Printf("\n❌ Failed repositories (%d):\n", len(failedRepos))
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	for i, failed := range failedRepos {
		fmt.Printf("\n%d. Repository: %s\n", i+1, failed.Name)
		fmt.Printf("   Path: %s\n", failed.Path)
		fmt.Printf("   Operation: %s\n", failed.Operation)
		fmt.Printf("   Error: %s\n", failed.Error)

		// Provide specific solutions based on error type
		rcm.printSolutionSuggestion(failed.Name, failed.Path, failed.Error)
	}

	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("\n💡 General Solutions:")
	fmt.Println("   • Use --resume for pending repositories; previously failed repositories are not retried automatically")
	fmt.Println("   • Check network connection and GitHub access")
	fmt.Println("   • Verify GitHub token is set: export GITHUB_TOKEN=\"your_token\"")
	fmt.Println("   • Resolve failed repositories, then review and back up the saved snapshot before starting a fresh operation")
	if path, err := rcm.stateManager.ResolveStateFilePath("github", state.Organization); err != nil {
		fmt.Printf("   • Snapshot path could not be verified: %v\n", err)
	} else {
		fmt.Printf("   • Saved snapshot: %s\n", path)
	}
	fmt.Println("   • Preserve this file for recovery before deciding how to start fresh")
}

// printSolutionSuggestion provides specific solution suggestions based on error patterns.
func (rcm *ResumableCloneManager) printSolutionSuggestion(repoName, repoPath, errorMsg string) {
	fmt.Println("   💡 Suggested Solution:")

	// Check for common error patterns
	errorLower := strings.ToLower(errorMsg)

	switch {
	case strings.Contains(errorLower, "already exists") || strings.Contains(errorLower, "destination path"):
		fmt.Printf("      → Clone destination already exists\n")
		fmt.Printf("      → Inspect and back up the existing directory: %s\n", repoPath)
		fmt.Printf("      → Preserve local data before choosing a separate clone destination\n")

	case strings.Contains(errorLower, "authentication") || strings.Contains(errorLower, "403") || strings.Contains(errorLower, "401"):
		fmt.Printf("      → Authentication failed - check your GitHub token\n")
		fmt.Printf("      → Set token: export GITHUB_TOKEN=\"your_personal_access_token\"\n")
		fmt.Printf("      → Create token at: https://github.com/settings/tokens\n")

	case strings.Contains(errorLower, "not found") || strings.Contains(errorLower, "404"):
		fmt.Printf("      → Repository not found or access denied\n")
		fmt.Printf("      → Verify repository exists: https://github.com/%s\n", repoName)
		fmt.Printf("      → Check if you have access permissions\n")

	case strings.Contains(errorLower, "timeout") || strings.Contains(errorLower, "connection"):
		fmt.Printf("      → Network connection issue\n")
		fmt.Printf("      → Check your internet connection\n")
		fmt.Printf("      → Resolve the connection issue, then review and back up the saved snapshot before starting fresh\n")

	case strings.Contains(errorLower, "remote") || strings.Contains(errorLower, "origin"):
		fmt.Printf("      → Git remote configuration issue\n")
		fmt.Printf("      → Inspect the configured remote URL in this repository: %s\n", repoPath)
		fmt.Printf("      → Back up local data before choosing a separate clone destination\n")

	case strings.Contains(errorLower, "permission") || strings.Contains(errorLower, "denied"):
		fmt.Printf("      → Permission denied\n")
		fmt.Printf("      → Inspect directory permissions: %s\n", filepath.Dir(repoPath))
		fmt.Printf("      → Ensure you have write access to the target directory\n")

	default:
		fmt.Printf("      → Check the error message above for specific details\n")
		fmt.Printf("      → Inspect and back up local data before choosing a separate clone destination: %s\n", repoPath)
		fmt.Printf("      → For persistent issues, use --debug flag for detailed logs\n")
	}
}
