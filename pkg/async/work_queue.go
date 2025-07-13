package async

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Job represents a unit of work to be processed
type Job interface {
	ID() string
	Execute(ctx context.Context) error
	Priority() int
	Retry() bool
	MaxRetries() int
	RetryCount() int
	SetRetryCount(count int)
}

// BaseJob provides a basic implementation of Job
type BaseJob struct {
	JobID       string `json:"id"`
	JobPriority int    `json:"priority"`
	CanRetry    bool   `json:"can_retry"`
	MaxRetry    int    `json:"max_retries"`
	RetryNum    int    `json:"retry_count"`
	Handler     func(ctx context.Context) error
}

func (j *BaseJob) ID() string              { return j.JobID }
func (j *BaseJob) Priority() int           { return j.JobPriority }
func (j *BaseJob) Retry() bool             { return j.CanRetry }
func (j *BaseJob) MaxRetries() int         { return j.MaxRetry }
func (j *BaseJob) RetryCount() int         { return j.RetryNum }
func (j *BaseJob) SetRetryCount(count int) { j.RetryNum = count }
func (j *BaseJob) Execute(ctx context.Context) error {
	if j.Handler != nil {
		return j.Handler(ctx)
	}
	return fmt.Errorf("no handler defined for job %s", j.JobID)
}

// JobResult represents the result of job execution
type JobResult struct {
	Job       Job           `json:"job"`
	Error     error         `json:"error,omitempty"`
	StartTime time.Time     `json:"start_time"`
	EndTime   time.Time     `json:"end_time"`
	Duration  time.Duration `json:"duration"`
	Retried   bool          `json:"retried"`
}

// WorkQueue provides a priority-based job queue with worker pool
type WorkQueue struct {
	name           string
	workers        int
	highPriority   chan Job
	normalPriority chan Job
	lowPriority    chan Job
	results        chan JobResult
	retryQueue     chan Job
	eventBus       *EventBus
	stats          QueueStats
	mu             sync.RWMutex
	stopCh         chan struct{}
	wg             sync.WaitGroup
	backoffFunc    func(attempt int) time.Duration
}

// QueueStats tracks work queue performance metrics
type QueueStats struct {
	TotalJobs       int64
	CompletedJobs   int64
	FailedJobs      int64
	RetriedJobs     int64
	ActiveWorkers   int32
	QueuedJobs      int32
	AverageWaitTime time.Duration
	AverageExecTime time.Duration
}

// WorkQueueConfig configures the work queue
type WorkQueueConfig struct {
	Name                 string
	Workers              int
	HighPriorityBuffer   int
	NormalPriorityBuffer int
	LowPriorityBuffer    int
	EnableRetry          bool
	MaxRetries           int
	RetryBackoff         func(attempt int) time.Duration
	EventBus             *EventBus
}

// DefaultWorkQueueConfig returns sensible defaults
func DefaultWorkQueueConfig(name string) WorkQueueConfig {
	return WorkQueueConfig{
		Name:                 name,
		Workers:              5,
		HighPriorityBuffer:   100,
		NormalPriorityBuffer: 500,
		LowPriorityBuffer:    200,
		EnableRetry:          true,
		MaxRetries:           3,
		RetryBackoff:         DefaultBackoffFunc,
	}
}

// DefaultBackoffFunc provides exponential backoff with jitter
func DefaultBackoffFunc(attempt int) time.Duration {
	base := time.Second
	backoff := time.Duration(1<<uint(attempt)) * base
	if backoff > time.Minute {
		backoff = time.Minute
	}
	// Add jitter (±25%)
	jitter := time.Duration(float64(backoff) * 0.25)
	return backoff + jitter - time.Duration(float64(jitter)*2*0.5)
}

// NewWorkQueue creates a new work queue
func NewWorkQueue(config WorkQueueConfig) *WorkQueue {
	wq := &WorkQueue{
		name:           config.Name,
		workers:        config.Workers,
		highPriority:   make(chan Job, config.HighPriorityBuffer),
		normalPriority: make(chan Job, config.NormalPriorityBuffer),
		lowPriority:    make(chan Job, config.LowPriorityBuffer),
		results:        make(chan JobResult, config.Workers*2),
		retryQueue:     make(chan Job, 100),
		eventBus:       config.EventBus,
		stopCh:         make(chan struct{}),
		backoffFunc:    config.RetryBackoff,
	}

	if wq.backoffFunc == nil {
		wq.backoffFunc = DefaultBackoffFunc
	}

	// Start workers
	for i := 0; i < config.Workers; i++ {
		wq.wg.Add(1)
		go wq.worker(i)
	}

	// Start retry handler
	if config.EnableRetry {
		wq.wg.Add(1)
		go wq.retryHandler()
	}

	return wq
}

// Submit adds a job to the appropriate priority queue
func (wq *WorkQueue) Submit(job Job) error {
	atomic.AddInt64(&wq.stats.TotalJobs, 1)
	atomic.AddInt32(&wq.stats.QueuedJobs, 1)

	select {
	case <-wq.stopCh:
		return fmt.Errorf("work queue is shutting down")
	default:
	}

	// Route to appropriate priority queue
	switch {
	case job.Priority() >= 8: // High priority
		select {
		case wq.highPriority <- job:
			return nil
		default:
			return fmt.Errorf("high priority queue is full")
		}
	case job.Priority() >= 4: // Normal priority
		select {
		case wq.normalPriority <- job:
			return nil
		default:
			return fmt.Errorf("normal priority queue is full")
		}
	default: // Low priority
		select {
		case wq.lowPriority <- job:
			return nil
		default:
			return fmt.Errorf("low priority queue is full")
		}
	}
}

// worker processes jobs from the queue
func (wq *WorkQueue) worker(workerID int) {
	defer wq.wg.Done()

	ctx := context.Background()
	for {
		select {
		case <-wq.stopCh:
			return
		case job := <-wq.selectNextJob():
			wq.processJob(ctx, job, workerID)
		}
	}
}

// selectNextJob selects the next job based on priority
func (wq *WorkQueue) selectNextJob() <-chan Job {
	// Priority: High > Normal > Low
	select {
	case job := <-wq.highPriority:
		return wq.wrapJob(job)
	default:
		select {
		case job := <-wq.highPriority:
			return wq.wrapJob(job)
		case job := <-wq.normalPriority:
			return wq.wrapJob(job)
		default:
			select {
			case job := <-wq.highPriority:
				return wq.wrapJob(job)
			case job := <-wq.normalPriority:
				return wq.wrapJob(job)
			case job := <-wq.lowPriority:
				return wq.wrapJob(job)
			}
		}
	}
}

// wrapJob wraps a job in a channel for consistent return type
func (wq *WorkQueue) wrapJob(job Job) <-chan Job {
	ch := make(chan Job, 1)
	ch <- job
	close(ch)
	return ch
}

// processJob executes a job and handles the result
func (wq *WorkQueue) processJob(ctx context.Context, job Job, workerID int) {
	atomic.AddInt32(&wq.stats.ActiveWorkers, 1)
	atomic.AddInt32(&wq.stats.QueuedJobs, -1)
	defer atomic.AddInt32(&wq.stats.ActiveWorkers, -1)

	start := time.Now()

	// Create job context with timeout
	jobCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	// Execute job
	err := job.Execute(jobCtx)
	duration := time.Since(start)

	result := JobResult{
		Job:       job,
		Error:     err,
		StartTime: start,
		EndTime:   time.Now(),
		Duration:  duration,
		Retried:   job.RetryCount() > 0,
	}

	wq.updateStats(duration, err == nil)

	// Handle job result
	if err != nil {
		atomic.AddInt64(&wq.stats.FailedJobs, 1)

		// Retry if applicable
		if job.Retry() && job.RetryCount() < job.MaxRetries() {
			job.SetRetryCount(job.RetryCount() + 1)
			atomic.AddInt64(&wq.stats.RetriedJobs, 1)

			// Send to retry queue with backoff
			go func() {
				backoff := wq.backoffFunc(job.RetryCount())
				time.Sleep(backoff)

				select {
				case wq.retryQueue <- job:
				case <-wq.stopCh:
				}
			}()
		}

		// Publish error event
		if wq.eventBus != nil {
			errorEvent := NewErrorEvent(fmt.Sprintf("worker-%d", workerID), err)
			wq.eventBus.PublishAsync(ctx, errorEvent)
		}
	} else {
		atomic.AddInt64(&wq.stats.CompletedJobs, 1)

		// Publish completion event
		if wq.eventBus != nil {
			completionEvent := BaseEvent{
				EventType:   EventTypeTaskCompleted,
				EventTime:   time.Now(),
				EventSource: fmt.Sprintf("worker-%d", workerID),
				EventData: map[string]interface{}{
					"job_id":   job.ID(),
					"duration": duration,
					"retried":  result.Retried,
				},
			}
			wq.eventBus.PublishAsync(ctx, completionEvent)
		}
	}

	// Send result
	select {
	case wq.results <- result:
	default:
		// Results channel is full, drop result
	}
}

// retryHandler handles job retries
func (wq *WorkQueue) retryHandler() {
	defer wq.wg.Done()

	for {
		select {
		case <-wq.stopCh:
			return
		case job := <-wq.retryQueue:
			// Re-submit job to appropriate queue
			if err := wq.Submit(job); err != nil {
				// Failed to re-submit, drop job
				fmt.Printf("Failed to re-submit job %s: %v\n", job.ID(), err)
			}
		}
	}
}

// Results returns a channel of job results
func (wq *WorkQueue) Results() <-chan JobResult {
	return wq.results
}

// updateStats updates queue performance statistics
func (wq *WorkQueue) updateStats(duration time.Duration, success bool) {
	wq.mu.Lock()
	defer wq.mu.Unlock()

	// Update average execution time
	if wq.stats.AverageExecTime == 0 {
		wq.stats.AverageExecTime = duration
	} else {
		alpha := 0.1
		wq.stats.AverageExecTime = time.Duration(
			alpha*float64(duration) + (1-alpha)*float64(wq.stats.AverageExecTime),
		)
	}
}

// GetStats returns current queue statistics
func (wq *WorkQueue) GetStats() QueueStats {
	wq.mu.RLock()
	defer wq.mu.RUnlock()

	stats := wq.stats
	stats.ActiveWorkers = atomic.LoadInt32(&wq.stats.ActiveWorkers)
	stats.QueuedJobs = atomic.LoadInt32(&wq.stats.QueuedJobs)
	stats.TotalJobs = atomic.LoadInt64(&wq.stats.TotalJobs)
	stats.CompletedJobs = atomic.LoadInt64(&wq.stats.CompletedJobs)
	stats.FailedJobs = atomic.LoadInt64(&wq.stats.FailedJobs)
	stats.RetriedJobs = atomic.LoadInt64(&wq.stats.RetriedJobs)

	return stats
}

// PrintStats prints detailed queue statistics
func (wq *WorkQueue) PrintStats() {
	stats := wq.GetStats()

	fmt.Printf("=== Work Queue '%s' Statistics ===\n", wq.name)
	fmt.Printf("Total Jobs: %d\n", stats.TotalJobs)
	fmt.Printf("Completed: %d\n", stats.CompletedJobs)
	fmt.Printf("Failed: %d\n", stats.FailedJobs)
	fmt.Printf("Retried: %d\n", stats.RetriedJobs)
	fmt.Printf("Active Workers: %d\n", stats.ActiveWorkers)
	fmt.Printf("Queued Jobs: %d\n", stats.QueuedJobs)
	fmt.Printf("Average Execution Time: %v\n", stats.AverageExecTime)

	if stats.TotalJobs > 0 {
		successRate := float64(stats.CompletedJobs) / float64(stats.TotalJobs) * 100
		fmt.Printf("Success Rate: %.2f%%\n", successRate)
	}
}

// Stop gracefully shuts down the work queue
func (wq *WorkQueue) Stop(timeout time.Duration) error {
	close(wq.stopCh)

	// Wait for workers to finish with timeout
	done := make(chan struct{})
	go func() {
		wq.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		close(wq.results)
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("timeout waiting for workers to stop")
	}
}

// QueueLengths returns the current length of each priority queue
func (wq *WorkQueue) QueueLengths() (high, normal, low int) {
	return len(wq.highPriority), len(wq.normalPriority), len(wq.lowPriority)
}

// Common job factories
func NewSimpleJob(id string, priority int, handler func(ctx context.Context) error) Job {
	return &BaseJob{
		JobID:       id,
		JobPriority: priority,
		CanRetry:    true,
		MaxRetry:    3,
		Handler:     handler,
	}
}

func NewFileProcessingJob(filePath string, processor func(ctx context.Context, path string) error) Job {
	return &BaseJob{
		JobID:       fmt.Sprintf("file:%s", filePath),
		JobPriority: 5,
		CanRetry:    true,
		MaxRetry:    2,
		Handler: func(ctx context.Context) error {
			return processor(ctx, filePath)
		},
	}
}

func NewRepositoryCloneJob(repoURL string, cloner func(ctx context.Context, url string) error) Job {
	return &BaseJob{
		JobID:       fmt.Sprintf("clone:%s", repoURL),
		JobPriority: 7,
		CanRetry:    true,
		MaxRetry:    3,
		Handler: func(ctx context.Context) error {
			return cloner(ctx, repoURL)
		},
	}
}
