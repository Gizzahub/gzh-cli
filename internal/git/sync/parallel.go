// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package sync

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Gizzahub/gzh-cli/pkg/git/provider"
)

// executeSyncPlan executes the synchronization plan either sequentially or in parallel.
func (e *SyncEngine) executeSyncPlan(ctx context.Context, plan SyncPlan) error {
	if e.options.Parallel <= 1 {
		return e.executeSequential(ctx, plan)
	}

	return e.executeParallel(ctx, plan)
}

// executeSequential executes synchronization tasks sequentially.
func (e *SyncEngine) executeSequential(ctx context.Context, plan SyncPlan) error {
	totalTasks := len(plan.Create) + len(plan.Update)
	if totalTasks == 0 {
		fmt.Println("No repositories to synchronize")
		return nil
	}

	fmt.Printf("🔄 Executing synchronization plan (%d repositories)...\n\n", totalTasks)

	var errors []error

	// Execute create tasks
	for i, repoSync := range plan.Create {
		fmt.Printf("[%d/%d] Creating %s...\n", i+1, totalTasks, repoSync.Source.FullName)
		if err := e.syncRepository(ctx, repoSync); err != nil {
			fmt.Printf("❌ Failed: %v\n", err)
			errors = append(errors, fmt.Errorf("%s: %w", repoSync.Source.FullName, err))
		} else {
			fmt.Printf("✅ Success\n")
		}
		fmt.Println()
	}

	// Execute update tasks
	for i, repoSync := range plan.Update {
		fmt.Printf("[%d/%d] Updating %s...\n", len(plan.Create)+i+1, totalTasks, repoSync.Source.FullName)
		if err := e.syncRepository(ctx, repoSync); err != nil {
			fmt.Printf("❌ Failed: %v\n", err)
			errors = append(errors, fmt.Errorf("%s: %w", repoSync.Source.FullName, err))
		} else {
			fmt.Printf("✅ Success\n")
		}
		fmt.Println()
	}

	if len(errors) > 0 {
		fmt.Printf("❌ Synchronization completed with %d errors\n", len(errors))
		for _, err := range errors {
			fmt.Printf("  - %v\n", err)
		}
		return fmt.Errorf("synchronization failed for %d repositories", len(errors))
	}

	fmt.Printf("✅ Synchronization completed successfully!\n")
	return nil
}

// executeParallel executes synchronization tasks in parallel.
func (e *SyncEngine) executeParallel(ctx context.Context, plan SyncPlan) error {
	totalTasks := len(plan.Create) + len(plan.Update)
	if totalTasks == 0 {
		fmt.Println("No repositories to synchronize")
		return nil
	}

	fmt.Printf("🔄 Executing synchronization plan (%d repositories, %d workers)...\n\n",
		totalTasks, e.options.Parallel)

	// Create task queue
	tasks := make(chan RepoSync, totalTasks)
	results := make(chan SyncResult, totalTasks)

	// Add all tasks to queue
	for _, repoSync := range plan.Create {
		tasks <- repoSync
	}
	for _, repoSync := range plan.Update {
		tasks <- repoSync
	}
	close(tasks)

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < e.options.Parallel; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			e.worker(ctx, workerID, tasks, results)
		}(i)
	}

	// Wait for workers to finish
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var errors []error
	completed := 0
	for result := range results {
		completed++
		if result.Error != nil {
			fmt.Printf("❌ [%d/%d] %s: %v\n", completed, totalTasks, result.Repository, result.Error)
			errors = append(errors, fmt.Errorf("%s: %w", result.Repository, result.Error))
		} else {
			fmt.Printf("✅ [%d/%d] %s: completed in %v\n",
				completed, totalTasks, result.Repository, result.Duration)
		}
	}

	if len(errors) > 0 {
		fmt.Printf("\n❌ Synchronization completed with %d errors\n", len(errors))
		for _, err := range errors {
			fmt.Printf("  - %v\n", err)
		}
		return fmt.Errorf("synchronization failed for %d repositories", len(errors))
	}

	fmt.Printf("\n✅ Synchronization completed successfully!\n")
	return nil
}

// worker processes synchronization tasks from the queue.
func (e *SyncEngine) worker(ctx context.Context, workerID int, tasks <-chan RepoSync, results chan<- SyncResult) {
	for repoSync := range tasks {
		start := time.Now()

		if e.options.Verbose {
			fmt.Printf("🔧 Worker %d: Starting %s\n", workerID, repoSync.Source.FullName)
		}

		err := e.syncRepository(ctx, repoSync)
		duration := time.Since(start)

		results <- SyncResult{
			Repository: repoSync.Source.FullName,
			Duration:   duration,
			Error:      err,
			WorkerID:   workerID,
		}
	}
}

// syncRepository synchronizes a single repository.
func (e *SyncEngine) syncRepository(ctx context.Context, repoSync RepoSync) error {
	// If creating a new repository, create it first
	if repoSync.Destination == nil {
		destTarget, err := e.options.GetDestinationTarget()
		if err != nil {
			return fmt.Errorf("failed to get destination target: %w", err)
		}

		if destTarget.IsRepository() {
			// Create single repository
			newRepo, err := e.destination.CreateRepository(ctx, provider.CreateRepoRequest{
				Name:        repoSync.Source.Name,
				Description: repoSync.Source.Description,
				Private:     repoSync.Source.Private,
				HasIssues:   true, // Default to true, can be configured later
				HasWiki:     true, // Default to true, can be configured later
				Topics:      repoSync.Source.Topics,
			})
			if err != nil {
				return fmt.Errorf("failed to create repository: %w", err)
			}
			repoSync.Destination = newRepo
		} else {
			// Create repository in organization
			// TODO: Implement organization-level repository creation
			return fmt.Errorf("organization-level repository creation not implemented")
		}
	}

	// Execute all sync actions
	for _, action := range repoSync.Actions {
		if e.options.Verbose {
			fmt.Printf("  Executing %s: %s\n", action.Type, action.Description)
		}

		if err := action.Handler(ctx); err != nil {
			return fmt.Errorf("action %s failed: %w", action.Type, err)
		}
	}

	return nil
}

// SyncResult represents the result of a synchronization task.
type SyncResult struct {
	Repository string
	Duration   time.Duration
	Error      error
	WorkerID   int
}

// ParallelSyncStats tracks statistics for parallel synchronization.
type ParallelSyncStats struct {
	TotalTasks      int           `json:"total_tasks"`
	CompletedTasks  int           `json:"completed_tasks"`
	FailedTasks     int           `json:"failed_tasks"`
	Workers         int           `json:"workers"`
	TotalDuration   time.Duration `json:"total_duration"`
	AverageDuration time.Duration `json:"average_duration"`
}

// SyncProgress tracks the progress of synchronization operations.
type SyncProgress struct {
	Total     int       `json:"total"`
	Completed int       `json:"completed"`
	Failed    int       `json:"failed"`
	Progress  float64   `json:"progress"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GetProgress returns the current synchronization progress.
func (s *SyncProgress) GetProgress() float64 {
	if s.Total == 0 {
		return 0
	}
	return float64(s.Completed+s.Failed) / float64(s.Total) * 100
}

// UpdateProgress updates the progress counters.
func (s *SyncProgress) UpdateProgress(completed, failed int) {
	s.Completed = completed
	s.Failed = failed
	s.Progress = s.GetProgress()
}

// RateLimiter manages rate limiting for API calls during synchronization.
type RateLimiter struct {
	tokens     chan struct{}
	refillRate time.Duration
}

// NewRateLimiter creates a new rate limiter.
func NewRateLimiter(maxConcurrent int, refillRate time.Duration) *RateLimiter {
	rl := &RateLimiter{
		tokens:     make(chan struct{}, maxConcurrent),
		refillRate: refillRate,
	}

	// Fill initial tokens
	for i := 0; i < maxConcurrent; i++ {
		rl.tokens <- struct{}{}
	}

	// Start token refill goroutine
	go rl.refillTokens()

	return rl
}

// Acquire acquires a token for rate limiting.
func (rl *RateLimiter) Acquire(ctx context.Context) error {
	select {
	case <-rl.tokens:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Release releases a token back to the pool.
func (rl *RateLimiter) Release() {
	select {
	case rl.tokens <- struct{}{}:
	default:
		// Channel is full, ignore
	}
}

// refillTokens periodically refills tokens.
func (rl *RateLimiter) refillTokens() {
	ticker := time.NewTicker(rl.refillRate)
	defer ticker.Stop()

	for range ticker.C {
		select {
		case rl.tokens <- struct{}{}:
		default:
			// Channel is full, skip
		}
	}
}
