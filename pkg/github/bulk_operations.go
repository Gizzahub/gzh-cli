package github

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gizzahub/gzh-manager-go/internal/helpers"
	"github.com/gizzahub/gzh-manager-go/internal/workerpool"
	"github.com/schollz/progressbar/v3"
)

// BulkOperationsConfig represents configuration for bulk operations.
type BulkOperationsConfig struct {
	// WorkerPool configuration
	PoolConfig workerpool.RepositoryPoolConfig
	// Progress tracking
	ShowProgress bool
	// Verbose output
	Verbose bool
}

// DefaultBulkOperationsConfig returns default configuration for bulk operations.
func DefaultBulkOperationsConfig() BulkOperationsConfig {
	return BulkOperationsConfig{
		PoolConfig:   workerpool.DefaultRepositoryPoolConfig(),
		ShowProgress: true,
		Verbose:      false,
	}
}

// BulkOperationsManager manages bulk repository operations using worker pools.
type BulkOperationsManager struct {
	config BulkOperationsConfig
	pool   *workerpool.RepositoryWorkerPool
}

// NewBulkOperationsManager creates a new bulk operations manager.
func NewBulkOperationsManager(config BulkOperationsConfig) *BulkOperationsManager {
	return &BulkOperationsManager{
		config: config,
		pool:   workerpool.NewRepositoryWorkerPool(config.PoolConfig),
	}
}

// Start initializes the bulk operations manager.
func (b *BulkOperationsManager) Start() error {
	return b.pool.Start()
}

// Stop shuts down the bulk operations manager.
func (b *BulkOperationsManager) Stop() {
	b.pool.Stop()
}

// RefreshAllWithWorkerPool performs bulk repository refresh using worker pools.
func (b *BulkOperationsManager) RefreshAllWithWorkerPool(ctx context.Context,
	targetPath, org, strategy string,
) error {
	// Get repository list
	repos, err := List(ctx, org)
	if err != nil {
		return fmt.Errorf("failed to list repositories: %w", err)
	}

	if len(repos) == 0 {
		fmt.Printf("No repositories found for organization: %s\n", org)
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

	// Create progress bar if enabled
	var bar *progressbar.ProgressBar
	if b.config.ShowProgress {
		bar = progressbar.NewOptions(len(jobs),
			progressbar.OptionSetDescription("Processing Repositories"),
			progressbar.OptionSetRenderBlankState(true),
		)
	}

	// Process repositories using worker pool
	processFn := func(ctx context.Context, job workerpool.RepositoryJob) error {
		return b.processRepositoryJob(ctx, job, org)
	}

	// Submit jobs and collect results
	resultsChan := b.pool.Results()

	// Submit all jobs
	for _, job := range jobs {
		if err := b.pool.SubmitJob(job, processFn); err != nil {
			return fmt.Errorf("failed to submit job for %s: %w", job.Repository, err)
		}
	}

	// Collect results
	successCount := 0
	errorCount := 0

	for i := 0; i < len(jobs); i++ {
		select {
		case result := <-resultsChan:
			if result.Success {
				successCount++

				if b.config.Verbose {
					fmt.Printf("✓ %s: %s completed successfully\n",
						result.Job.Repository, result.Job.Operation)
				}
			} else {
				errorCount++

				fmt.Printf("✗ %s: %s failed: %v\n",
					result.Job.Repository, result.Job.Operation, result.Error)
			}

			if bar != nil {
				_ = bar.Add(1) // Progress bar error is not critical
			}

		case <-ctx.Done():
			return fmt.Errorf("operation cancelled: %w", ctx.Err())
		}
	}

	fmt.Printf("\nBulk operation completed: %d successful, %d failed\n", successCount, errorCount)

	return nil
}

// processRepositoryJob processes a single repository job.
func (b *BulkOperationsManager) processRepositoryJob(ctx context.Context,
	job workerpool.RepositoryJob, org string,
) error {
	switch job.Operation {
	case workerpool.OperationClone:
		return Clone(ctx, job.Path, org, job.Repository)

	case workerpool.OperationPull:
		return b.executeGitOperation(ctx, job.Path, "pull")

	case workerpool.OperationFetch:
		return b.executeGitOperation(ctx, job.Path, "fetch")

	case workerpool.OperationReset:
		// Reset hard HEAD and pull
		if err := b.executeGitOperation(ctx, job.Path, "reset", "--hard", "HEAD"); err != nil {
			return fmt.Errorf("git reset failed: %w", err)
		}

		return b.executeGitOperation(ctx, job.Path, "pull")

	case workerpool.OperationConfig:
		// Config operation - placeholder for configuration updates
		return fmt.Errorf("config operation not yet implemented")

	default:
		return fmt.Errorf("unknown operation: %s", job.Operation)
	}
}

// executeGitOperation executes a git command in the repository path.
func (b *BulkOperationsManager) executeGitOperation(ctx context.Context,
	repoPath string, args ...string,
) error {
	// Check if repository is valid
	repoType, _ := helpers.CheckGitRepoType(repoPath)
	if repoType == helpers.RepoTypeEmpty {
		return fmt.Errorf("repository is empty or not a git repository")
	}

	// Build git command
	gitArgs := append([]string{"-C", repoPath}, args...)
	cmd := exec.CommandContext(ctx, "git", gitArgs...) //nolint:gosec // git command with controlled arguments

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git %s failed: %w", args[0], err)
	}

	return nil
}

// RefreshAllWithWorkerPoolWrapper provides a drop-in replacement for the original RefreshAll.
func RefreshAllWithWorkerPool(ctx context.Context, targetPath, org, strategy string, parallel int, maxRetries int) error {
	config := DefaultBulkOperationsConfig()

	// Override defaults with user-specified values
	if parallel > 0 {
		config.PoolConfig.CloneWorkers = parallel
		config.PoolConfig.UpdateWorkers = parallel + (parallel / 2) // 50% more for updates

		config.PoolConfig.ConfigWorkers = parallel / 2 // 50% less for config operations
		if config.PoolConfig.ConfigWorkers < 1 {
			config.PoolConfig.ConfigWorkers = 1
		}
	}

	if maxRetries > 0 {
		config.PoolConfig.RetryAttempts = maxRetries
	}

	manager := NewBulkOperationsManager(config)
	if err := manager.Start(); err != nil { //nolint:contextcheck // Manager start method manages its own context
		return fmt.Errorf("failed to start bulk operations manager: %w", err)
	}
	defer manager.Stop()

	return manager.RefreshAllWithWorkerPool(ctx, targetPath, org, strategy)
}

// BulkCloneOptions represents options for bulk clone operations.
type BulkCloneOptions struct {
	// WorkerPoolConfig allows customizing worker pool behavior
	WorkerPoolConfig workerpool.RepositoryPoolConfig
	// Organizations to clone
	Organizations []string
	// Strategy for existing repositories ("reset", "pull", "fetch")
	Strategy string
	// ShowProgress enables progress bar
	ShowProgress bool
	// Verbose enables detailed output
	Verbose bool
}

// BulkCloneMultipleOrganizations clones repositories from multiple organizations using worker pools.
func BulkCloneMultipleOrganizations(ctx context.Context, targetBasePath string,
	options BulkCloneOptions,
) error {
	if len(options.Organizations) == 0 {
		return fmt.Errorf("no organizations specified")
	}

	config := BulkOperationsConfig{
		PoolConfig:   options.WorkerPoolConfig,
		ShowProgress: options.ShowProgress,
		Verbose:      options.Verbose,
	}

	manager := NewBulkOperationsManager(config)
	if err := manager.Start(); err != nil { //nolint:contextcheck // Manager start method manages its own context
		return fmt.Errorf("failed to start bulk operations manager: %w", err)
	}
	defer manager.Stop()

	// Process each organization
	for i, org := range options.Organizations {
		fmt.Printf("\n[%d/%d] Processing organization: %s\n", i+1, len(options.Organizations), org)

		orgPath := filepath.Join(targetBasePath, org)
		if err := os.MkdirAll(orgPath, 0o750); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", org, err)
		}

		if err := manager.RefreshAllWithWorkerPool(ctx, orgPath, org, options.Strategy); err != nil {
			fmt.Printf("Error processing organization %s: %v\n", org, err)
			continue
		}
	}

	return nil
}
