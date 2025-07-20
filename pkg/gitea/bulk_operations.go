package gitea

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/gizzahub/gzh-manager-go/internal/workerpool"
)

// RefreshAllWithWorkerPool performs bulk repository refresh using worker pools.
func RefreshAllWithWorkerPool(ctx context.Context, targetPath, org string) error {
	config := workerpool.DefaultRepositoryPoolConfig()

	pool := workerpool.NewRepositoryWorkerPool(config)
	if err := pool.Start(); err != nil { //nolint:contextcheck // Worker pool manages its own context lifecycle
		return fmt.Errorf("failed to start worker pool: %w", err)
	}
	defer pool.Stop()

	// Get repository list
	repos, err := List(ctx, org)
	if err != nil {
		return fmt.Errorf("failed to list repositories: %w", err)
	}

	if len(repos) == 0 {
		fmt.Printf("No repositories found for organization: %s\n", org)
		return nil
	}

	// Create jobs for each repository - Gitea RefreshAll currently only supports cloning
	jobs := make([]workerpool.RepositoryJob, 0, len(repos))
	for _, repo := range repos {
		repoPath := filepath.Join(targetPath, repo)

		jobs = append(jobs, workerpool.RepositoryJob{
			Repository: repo,
			Operation:  workerpool.OperationClone,
			Path:       repoPath,
		})
	}

	// Process repositories using worker pool
	processFn := func(ctx context.Context, job workerpool.RepositoryJob) error {
		return processGiteaRepositoryJob(ctx, job, org)
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

	fmt.Printf("Gitea bulk operation completed: %d/%d successful\n", successCount, len(results))

	return nil
}

// processGiteaRepositoryJob processes a single Gitea repository job.
func processGiteaRepositoryJob(ctx context.Context, job workerpool.RepositoryJob, org string) error {
	switch job.Operation {
	case workerpool.OperationClone:
		return Clone(ctx, job.Path, org, job.Repository, "")

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
	cmd := exec.CommandContext(ctx, "git", gitArgs...) //nolint:gosec // Git command with controlled arguments

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git %s failed: %w", args[0], err)
	}

	return nil
}
