// Package workerpool provides configurable worker pool implementations for bulk operations.
package workerpool

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// Config represents worker pool configuration.
type Config struct {
	// WorkerCount specifies the number of workers. If 0, defaults to runtime.NumCPU()
	WorkerCount int
	// BufferSize specifies the job queue buffer size. If 0, defaults to WorkerCount * 2
	BufferSize int
	// Timeout specifies the maximum time to wait for a job to complete
	Timeout time.Duration
}

// DefaultConfig returns a sensible default configuration.
func DefaultConfig() Config {
	return Config{
		WorkerCount: runtime.NumCPU(),
		BufferSize:  runtime.NumCPU() * 2,
		Timeout:     30 * time.Second,
	}
}

// Job represents a unit of work to be processed by the worker pool.
type Job[T any] struct {
	Data T
	Fn   func(context.Context, T) error
}

// Result represents the result of processing a job.
type Result[T any] struct {
	Data  T
	Error error
}

// Pool represents a generic worker pool.
type Pool[T any] struct {
	config  Config
	jobs    chan Job[T]
	results chan Result[T]
	wg      sync.WaitGroup
	ctx     context.Context
	cancel  context.CancelFunc
	started bool
	mu      sync.RWMutex
}

// New creates a new worker pool with the given configuration.
func New[T any](config Config) *Pool[T] {
	if config.WorkerCount <= 0 {
		config.WorkerCount = runtime.NumCPU()
	}

	if config.BufferSize <= 0 {
		config.BufferSize = config.WorkerCount * 2
	}

	if config.Timeout <= 0 {
		config.Timeout = 30 * time.Second
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Pool[T]{
		config:  config,
		jobs:    make(chan Job[T], config.BufferSize),
		results: make(chan Result[T], config.BufferSize),
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Start initializes and starts the worker pool.
func (p *Pool[T]) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.started {
		return fmt.Errorf("worker pool already started")
	}

	// Start workers
	for i := 0; i < p.config.WorkerCount; i++ {
		p.wg.Add(1)

		go p.worker(i)
	}

	p.started = true

	return nil
}

// Submit submits a job to the worker pool.
func (p *Pool[T]) Submit(data T, fn func(context.Context, T) error) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.started {
		return fmt.Errorf("worker pool not started")
	}

	select {
	case p.jobs <- Job[T]{Data: data, Fn: fn}:
		return nil
	case <-p.ctx.Done():
		return p.ctx.Err()
	default:
		return fmt.Errorf("job queue is full")
	}
}

// Results returns a channel to receive job results.
func (p *Pool[T]) Results() <-chan Result[T] {
	return p.results
}

// Stop gracefully shuts down the worker pool.
func (p *Pool[T]) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.started {
		return
	}

	// Signal shutdown and close job channel
	close(p.jobs)

	// Wait for all workers to finish
	p.wg.Wait()

	// Close results channel
	close(p.results)

	// Cancel context
	p.cancel()

	p.started = false
}

// StopWithTimeout stops the worker pool with a timeout.
func (p *Pool[T]) StopWithTimeout(timeout time.Duration) error {
	done := make(chan struct{})

	go func() {
		p.Stop()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-time.After(timeout):
		p.cancel() // Force cancellation
		return fmt.Errorf("worker pool shutdown timed out after %v", timeout)
	}
}

// worker is the main worker goroutine.
func (p *Pool[T]) worker(id int) {
	defer p.wg.Done()

	for job := range p.jobs {
		// Create a context with timeout for this job
		jobCtx, jobCancel := context.WithTimeout(p.ctx, p.config.Timeout)

		// Execute the job
		err := job.Fn(jobCtx, job.Data)

		// Send result
		result := Result[T]{
			Data:  job.Data,
			Error: err,
		}

		select {
		case p.results <- result:
		case <-p.ctx.Done():
			jobCancel()
			return
		}

		jobCancel()

		// Check if we should stop
		select {
		case <-p.ctx.Done():
			return
		default:
		}
	}
}

// ProcessBatch processes a batch of items using the worker pool.
func ProcessBatch[T any](ctx context.Context, items []T, config Config,
	processFn func(context.Context, T) error,
) ([]Result[T], error) {
	if len(items) == 0 {
		return []Result[T]{}, nil
	}

	// Ensure buffer size is large enough for all items
	if config.BufferSize < len(items) {
		config.BufferSize = len(items)
	}

	pool := New[T](config)
	if err := pool.Start(); err != nil {
		return nil, fmt.Errorf("failed to start worker pool: %w", err)
	}
	defer pool.Stop()

	// Submit all jobs
	for _, item := range items {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		if err := pool.Submit(item, processFn); err != nil {
			return nil, fmt.Errorf("failed to submit job: %w", err)
		}
	}

	// Collect results
	results := make([]Result[T], 0, len(items))
	for i := 0; i < len(items); i++ {
		select {
		case result := <-pool.Results():
			results = append(results, result)
		case <-ctx.Done():
			return results, ctx.Err()
		}
	}

	return results, nil
}
