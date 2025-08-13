package gitlab

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Gizzahub/gzh-manager-go/internal/git"
	"github.com/Gizzahub/gzh-manager-go/internal/workerpool"
)

// RefreshAllWithWorkerPool performs bulk repository refresh using worker pools.
func RefreshAllWithWorkerPool(ctx context.Context, targetPath, group, strategy string, parallel int, maxRetries int) error {
	config := workerpool.DefaultRepositoryPoolConfig()

	// Override defaults with user-specified values
	if parallel > 0 {
		config.CloneWorkers = parallel
		config.UpdateWorkers = parallel + (parallel / 2) // 50% more for updates

		config.ConfigWorkers = parallel / 2 // 50% less for config operations
		if config.ConfigWorkers < 1 {
			config.ConfigWorkers = 1
		}
	}

	if maxRetries > 0 {
		config.RetryAttempts = maxRetries
	}

	pool := workerpool.NewRepositoryWorkerPool(config)
	if err := pool.Start(); err != nil { //nolint:contextcheck // Worker pool start manages its own context
		return fmt.Errorf("failed to start worker pool: %w", err)
	}
	defer pool.Stop()

	// Get repository list
	repos, err := List(ctx, group)
	if err != nil {
		return fmt.Errorf("failed to list repositories: %w", err)
	}

	if len(repos) == 0 {
		fmt.Printf("No repositories found for group: %s\n", group)
		return nil
	}

	// Create jobs for each repository
	jobs := make([]workerpool.RepositoryJob, 0, len(repos))
	for _, repo := range repos {
		repoPath := filepath.Join(targetPath, repo)

		// Determine operation type based on whether repo exists
		var operation workerpool.RepositoryOperation
		if _, err := os.Stat(repoPath); os.IsNotExist(err) {
			operation = workerpool.OperationClone
		} else {
			switch strategy {
			case git.StrategyReset:
				operation = workerpool.OperationReset
			case git.StrategyPull:
				operation = workerpool.OperationPull
			case git.StrategyFetch:
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

	// Process repositories using worker pool
	processFn := func(ctx context.Context, job workerpool.RepositoryJob) error {
		return processGitLabRepositoryJob(ctx, job, group)
	}

	results, err := pool.ProcessRepositories(ctx, jobs, processFn)
	if err != nil {
		return fmt.Errorf("failed to process repositories: %w", err)
	}

	// Report results
	successCount := 0

	for _, result := range results {
		if result.Success {
			successCount++
		} else {
			fmt.Printf("Error processing %s: %v\n", result.Job.Repository, result.Error)
		}
	}

	fmt.Printf("GitLab bulk operation completed: %d/%d successful\n", successCount, len(results))

	return nil
}

// processGitLabRepositoryJob processes a single GitLab repository job.
func processGitLabRepositoryJob(ctx context.Context, job workerpool.RepositoryJob, group string) error {
	switch job.Operation {
	case workerpool.OperationClone:
		return Clone(ctx, job.Path, group, job.Repository, "")

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

// executeGitOperation executes a git command in the repository path with security validation.
func executeGitOperation(ctx context.Context, repoPath string, args ...string) error {
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
