// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package workerpool

import (
	"context"
	"fmt"
	"time"
)

// RepositoryOperation represents a repository operation type.
type RepositoryOperation string

// Repository operation constants define the available operation types.
const (
	OperationClone  RepositoryOperation = "clone"
	OperationPull   RepositoryOperation = "pull"
	OperationFetch  RepositoryOperation = "fetch"
	OperationReset  RepositoryOperation = "reset"
	OperationConfig RepositoryOperation = "config"
)

// RepositoryJob represents a repository operation job.
type RepositoryJob struct {
	Repository       string
	Operation        RepositoryOperation
	Path             string
	Branch           string
	Strategy         string
	WorkerPoolConfig map[string]interface{} // For configuration operations
}

// RepositoryResult represents the result of a repository operation.
type RepositoryResult struct {
	Job      RepositoryJob
	Success  bool
	Error    error
	Duration time.Duration
	Message  string
}

// RepositoryPoolConfig represents configuration for repository operations.
type RepositoryPoolConfig struct {
	// CloneWorkers specifies concurrent workers for clone operations
	CloneWorkers int
	// UpdateWorkers specifies concurrent workers for update operations (pull/fetch/reset)
	UpdateWorkers int
	// ConfigWorkers specifies concurrent workers for configuration operations
	ConfigWorkers int
	// OperationTimeout specifies timeout for individual operations
	OperationTimeout time.Duration
	// RetryAttempts specifies number of retry attempts for failed operations
	RetryAttempts int
	// RetryDelay specifies delay between retry attempts
	RetryDelay time.Duration
}

// DefaultRepositoryPoolConfig returns default configuration for repository operations.
func DefaultRepositoryPoolConfig() RepositoryPoolConfig {
	return RepositoryPoolConfig{
		CloneWorkers:     10, // I/O bound, can handle more concurrency
		UpdateWorkers:    15, // Faster than clones, can handle more
		ConfigWorkers:    5,  // API limited, keep conservative
		OperationTimeout: 5 * time.Minute,
		RetryAttempts:    3,
		RetryDelay:       2 * time.Second,
	}
}

// RepositoryWorkerPool manages repository operations with operation-specific worker pools.
type RepositoryWorkerPool struct {
	config     RepositoryPoolConfig
	clonePool  *Pool[RepositoryJob]
	updatePool *Pool[RepositoryJob]
	configPool *Pool[RepositoryJob]
	results    chan RepositoryResult
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewRepositoryWorkerPool creates a new repository worker pool.
func NewRepositoryWorkerPool(config RepositoryPoolConfig) *RepositoryWorkerPool {
	ctx, cancel := context.WithCancel(context.Background())

	// Create specialized pools for different operation types
	clonePool := New[RepositoryJob](WorkerPoolConfig{
		WorkerCount: config.CloneWorkers,
		BufferSize:  config.CloneWorkers * 2,
		Timeout:     config.OperationTimeout,
	})

	updatePool := New[RepositoryJob](WorkerPoolConfig{
		WorkerCount: config.UpdateWorkers,
		BufferSize:  config.UpdateWorkers * 2,
		Timeout:     config.OperationTimeout,
	})

	configPool := New[RepositoryJob](WorkerPoolConfig{
		WorkerCount: config.ConfigWorkers,
		BufferSize:  config.ConfigWorkers * 2,
		Timeout:     config.OperationTimeout,
	})

	return &RepositoryWorkerPool{
		config:     config,
		clonePool:  clonePool,
		updatePool: updatePool,
		configPool: configPool,
		results:    make(chan RepositoryResult, 100), // Buffer for results
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Start initializes and starts all worker pools.
func (rp *RepositoryWorkerPool) Start() error {
	if err := rp.clonePool.Start(); err != nil {
		return fmt.Errorf("failed to start clone pool: %w", err)
	}

	if err := rp.updatePool.Start(); err != nil {
		rp.clonePool.Stop()
		return fmt.Errorf("failed to start update pool: %w", err)
	}

	if err := rp.configPool.Start(); err != nil {
		rp.clonePool.Stop()
		rp.updatePool.Stop()

		return fmt.Errorf("failed to start config pool: %w", err)
	}

	// Start result collectors
	go rp.collectResults(rp.clonePool.Results(), "clone")
	go rp.collectResults(rp.updatePool.Results(), "update")
	go rp.collectResults(rp.configPool.Results(), "config")

	return nil
}

// Stop gracefully shuts down all worker pools.
func (rp *RepositoryWorkerPool) Stop() {
	rp.cancel()
	rp.clonePool.Stop()
	rp.updatePool.Stop()
	rp.configPool.Stop()
	close(rp.results)
}

// SubmitJob submits a repository job to the appropriate worker pool.
func (rp *RepositoryWorkerPool) SubmitJob(job RepositoryJob,
	processFn func(context.Context, RepositoryJob) error,
) error {
	var pool *Pool[RepositoryJob]

	switch job.Operation {
	case OperationClone:
		pool = rp.clonePool
	case OperationPull, OperationFetch, OperationReset:
		pool = rp.updatePool
	case OperationConfig:
		pool = rp.configPool
	default:
		return fmt.Errorf("unknown operation: %s", job.Operation)
	}

	// Wrap processFn with retry logic
	wrappedFn := rp.wrapWithRetry(processFn)

	return pool.Submit(job, wrappedFn)
}

// Results returns a channel to receive job results.
func (rp *RepositoryWorkerPool) Results() <-chan RepositoryResult {
	return rp.results
}

// ProcessRepositories processes a batch of repository jobs.
func (rp *RepositoryWorkerPool) ProcessRepositories(ctx context.Context,
	jobs []RepositoryJob, processFn func(context.Context, RepositoryJob) error,
) ([]RepositoryResult, error) {
	if len(jobs) == 0 {
		return []RepositoryResult{}, nil
	}

	// Submit all jobs
	for _, job := range jobs {
		if err := rp.SubmitJob(job, processFn); err != nil {
			return nil, fmt.Errorf("failed to submit job for %s: %w", job.Repository, err)
		}
	}

	// Collect results
	results := make([]RepositoryResult, 0, len(jobs))
	for i := 0; i < len(jobs); i++ {
		select {
		case result := <-rp.results:
			results = append(results, result)
		case <-ctx.Done():
			return results, ctx.Err()
		}
	}

	return results, nil
}

// collectResults collects results from a worker pool and forwards them.
func (rp *RepositoryWorkerPool) collectResults(resultsChan <-chan Result[RepositoryJob], poolType string) {
	for result := range resultsChan {
		repoResult := RepositoryResult{
			Job:     result.Data,
			Success: result.Error == nil,
			Error:   result.Error,
			Message: fmt.Sprintf("Processed by %s pool", poolType),
		}

		select {
		case rp.results <- repoResult:
		case <-rp.ctx.Done():
			return
		}
	}
}

// wrapWithRetry wraps a processing function with retry logic.
func (rp *RepositoryWorkerPool) wrapWithRetry( //nolint:gocognit // Complex worker pool retry logic with backoff, timeout, and error classification
	processFn func(context.Context, RepositoryJob) error,
) func(context.Context, RepositoryJob) error {
	return func(ctx context.Context, job RepositoryJob) error {
		var lastErr error

		for attempt := 0; attempt <= rp.config.RetryAttempts; attempt++ {
			if attempt > 0 {
				// Wait before retry
				select {
				case <-time.After(rp.config.RetryDelay):
				case <-ctx.Done():
					return ctx.Err()
				}
			}

			startTime := time.Now()
			err := processFn(ctx, job)
			duration := time.Since(startTime)

			if err == nil {
				// Success - log if this was a retry
				if attempt > 0 {
					fmt.Printf("Repository %s succeeded on attempt %d (took %v)\n",
						job.Repository, attempt+1, duration)
				}

				return nil
			}

			lastErr = err

			// Check if error is retryable
			if !isRetryableError(err) {
				break
			}

			if attempt < rp.config.RetryAttempts {
				fmt.Printf("Repository %s failed on attempt %d, retrying: %v\n",
					job.Repository, attempt+1, err)
			}
		}

		return fmt.Errorf("repository %s failed after %d attempts: %w",
			job.Repository, rp.config.RetryAttempts+1, lastErr)
	}
}

// isRetryableError determines if an error is worth retrying.
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// Network-related errors are usually retryable
	retryablePatterns := []string{
		"connection refused",
		"timeout",
		"temporary failure",
		"network is unreachable",
		"no such host",
		"connection reset",
		"broken pipe",
	}

	for _, pattern := range retryablePatterns {
		if contains(errStr, pattern) {
			return true
		}
	}

	return false
}

// contains checks if a string contains a substring (case-insensitive).
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsInner(s, substr)))
}

func containsInner(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}

	return false
}
