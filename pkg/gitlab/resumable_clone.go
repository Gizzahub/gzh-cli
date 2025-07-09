package gitlab

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gizzahub/gzh-manager-go/internal/workerpool"
	bulkclonepkg "github.com/gizzahub/gzh-manager-go/pkg/bulk-clone"
	"github.com/schollz/progressbar/v3"
)

// ResumableCloneManager handles resumable clone operations for GitLab
type ResumableCloneManager struct {
	stateManager *bulkclonepkg.StateManager
	config       workerpool.RepositoryPoolConfig
}

// NewResumableCloneManager creates a new resumable clone manager for GitLab
func NewResumableCloneManager(config workerpool.RepositoryPoolConfig) *ResumableCloneManager {
	return &ResumableCloneManager{
		stateManager: bulkclonepkg.NewStateManager(""),
		config:       config,
	}
}

// RefreshAllResumable performs bulk repository refresh with resumable support for GitLab
func (rcm *ResumableCloneManager) RefreshAllResumable(ctx context.Context, targetPath, group, strategy string, parallel, maxRetries int, resume bool) error {
	var state *bulkclonepkg.CloneState
	var err error

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
		config.MaxRetries = maxRetries
	}

	// Create and start worker pool
	pool := workerpool.NewRepositoryWorkerPool(config)
	if err := pool.Start(); err != nil {
		return fmt.Errorf("failed to start worker pool: %w", err)
	}
	defer pool.Stop()

	// Create progress bar
	completed, failed, pending := state.GetProgress()
	totalProcessed := completed + failed

	bar := progressbar.NewOptions(len(allRepos),
		progressbar.OptionSetDescription(fmt.Sprintf("GitLab %s", group)),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionOnCompletion(func() {
			fmt.Printf("\n")
		}),
	)

	// Set initial progress
	bar.Set(totalProcessed)

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
			CloneURL:   fmt.Sprintf("https://gitlab.com/%s/%s.git", group, repo),
		})
	}

	// Process jobs
	results := pool.ProcessJobs(ctx, jobs)

	// Track results and update state
	successCount := 0
	failureCount := 0

	// Set up periodic state saving
	stateSaveTicker := time.NewTicker(30 * time.Second)
	defer stateSaveTicker.Stop()

	for {
		select {
		case result, ok := <-results:
			if !ok {
				// All results processed
				goto completed
			}

			if result.Error != nil {
				failureCount++
				state.AddFailedRepository(result.Repository, result.Path, string(result.Operation), result.Error.Error(), 1)
				fmt.Printf("\n❌ Failed %s for %s: %v\n", result.Operation, result.Repository, result.Error)
			} else {
				successCount++
				state.AddCompletedRepository(result.Repository, result.Path, string(result.Operation), result.Message)
				if result.Message != "" {
					fmt.Printf("\n✅ %s: %s\n", result.Repository, result.Message)
				}
			}

			bar.Add(1)

		case <-stateSaveTicker.C:
			// Periodically save state
			if err := rcm.stateManager.SaveState(state); err != nil {
				fmt.Printf("⚠️  Warning: failed to save state: %v\n", err)
			}

		case <-ctx.Done():
			// Operation cancelled
			state.MarkCancelled()
			rcm.stateManager.SaveState(state)
			return fmt.Errorf("operation cancelled: %w", ctx.Err())
		}
	}

completed:
	bar.Finish()

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

	fmt.Printf("\nCompleted: %d successful, %d failed\n", successCount, failureCount)

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

// RefreshAllResumable is a convenience function for resumable cloning
func RefreshAllResumable(ctx context.Context, targetPath, group, strategy string, parallel, maxRetries int, resume bool) error {
	config := workerpool.DefaultRepositoryPoolConfig()
	manager := NewResumableCloneManager(config)
	return manager.RefreshAllResumable(ctx, targetPath, group, strategy, parallel, maxRetries, resume)
}

// GetCloneState returns the current clone state for a group
func GetCloneState(group string) (*bulkclonepkg.CloneState, error) {
	stateManager := bulkclonepkg.NewStateManager("")
	return stateManager.LoadState("gitlab", group)
}

// DeleteCloneState removes the state file for a group
func DeleteCloneState(group string) error {
	stateManager := bulkclonepkg.NewStateManager("")
	return stateManager.DeleteState("gitlab", group)
}

// ListCloneStates returns all saved clone states
func ListCloneStates() ([]bulkclonepkg.CloneState, error) {
	stateManager := bulkclonepkg.NewStateManager("")
	return stateManager.ListStates()
}
