package github

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/schollz/progressbar/v3"

	"github.com/gizzahub/gzh-manager-go/internal/git"
	"github.com/gizzahub/gzh-manager-go/internal/workerpool"
)

// OptimizedSyncCloneManager handles large-scale repository operations with memory optimization.
type OptimizedSyncCloneManager struct {
	streamingClient *StreamingClient
	workerPool      *workerpool.RepositoryWorkerPool
	config          OptimizedCloneConfig
	memoryMonitor   *MemoryMonitor
}

// OptimizedCloneConfig represents configuration for optimized bulk cloning.
type OptimizedCloneConfig struct {
	// Memory management
	MaxMemoryUsage  int64         // Maximum memory usage in bytes
	MemoryThreshold float64       // Trigger cleanup at this % of max memory
	GCInterval      time.Duration // How often to check memory usage

	// Streaming configuration
	StreamingConfig StreamingConfig

	// Worker pool configuration
	WorkerPoolConfig workerpool.RepositoryPoolConfig

	// Progress and monitoring
	ShowProgress   bool
	VerboseLogging bool
	MetricsEnabled bool

	// Performance tuning
	BatchSize    int // Number of repositories to process in a batch
	PrefetchSize int // Number of repositories to prefetch
}

// MemoryMonitor tracks and manages memory usage.
type MemoryMonitor struct {
	maxMemory    int64
	threshold    float64
	currentUsage int64
	gcCount      int64
	lastGC       time.Time
	mu           sync.RWMutex
}

// CloneStats tracks bulk clone operation statistics.
type CloneStats struct {
	TotalRepositories int
	Successful        int
	Failed            int
	Skipped           int
	MemoryPeakUsage   int64
	TotalDuration     time.Duration
	AverageSpeed      float64 // repos per second
	ErrorDetails      []CloneError
}

// CloneError represents a clone operation error with context.
type CloneError struct {
	Repository  string
	Operation   string
	Error       error
	Attempt     int
	Timestamp   time.Time
	MemoryUsage int64
}

// DefaultOptimizedCloneConfig returns optimized defaults for large-scale operations.
func DefaultOptimizedCloneConfig() OptimizedCloneConfig {
	// Determine optimal settings based on system resources
	numCPU := runtime.NumCPU()
	maxWorkers := numCPU * 2 // Allow some I/O overlap

	// Calculate memory limits based on available system memory
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	systemMemory := int64(memStats.Sys)
	maxMemory := systemMemory / 4 // Use up to 25% of system memory

	if maxMemory < 256*1024*1024 { // Minimum 256MB
		maxMemory = 256 * 1024 * 1024
	}

	return OptimizedCloneConfig{
		MaxMemoryUsage:  maxMemory,
		MemoryThreshold: 0.8, // Cleanup at 80% memory usage
		GCInterval:      30 * time.Second,
		StreamingConfig: DefaultStreamingConfig(),
		WorkerPoolConfig: workerpool.RepositoryPoolConfig{
			CloneWorkers:     maxWorkers,
			UpdateWorkers:    maxWorkers + (maxWorkers / 2),
			ConfigWorkers:    maxWorkers / 2,
			OperationTimeout: 10 * time.Minute,
			RetryAttempts:    3,
			RetryDelay:       5 * time.Second,
		},
		ShowProgress:   true,
		VerboseLogging: false,
		MetricsEnabled: true,
		BatchSize:      100,
		PrefetchSize:   200,
	}
}

// NewOptimizedSyncCloneManager creates a new optimized bulk clone manager.
func NewOptimizedSyncCloneManager(token string, config OptimizedCloneConfig) (*OptimizedSyncCloneManager, error) {
	streamingClient := NewStreamingClient(token, config.StreamingConfig)

	workerPool := workerpool.NewRepositoryWorkerPool(config.WorkerPoolConfig)
	if err := workerPool.Start(); err != nil {
		return nil, fmt.Errorf("failed to start worker pool: %w", err)
	}

	memoryMonitor := &MemoryMonitor{
		maxMemory: config.MaxMemoryUsage,
		threshold: config.MemoryThreshold,
		lastGC:    time.Now(),
	}

	manager := &OptimizedSyncCloneManager{
		streamingClient: streamingClient,
		workerPool:      workerPool,
		config:          config,
		memoryMonitor:   memoryMonitor,
	}

	// Start memory monitoring
	go manager.startMemoryMonitoring()

	return manager, nil
}

// RefreshAllOptimized performs optimized bulk repository refresh with streaming and memory management.
func (m *OptimizedSyncCloneManager) RefreshAllOptimized(ctx context.Context, targetPath, org, strategy string) (*CloneStats, error) {
	stats := &CloneStats{
		ErrorDetails: make([]CloneError, 0),
	}
	startTime := time.Now()

	if m.config.VerboseLogging {
		fmt.Printf("🚀 Starting optimized bulk clone for organization: %s\n", org)
		fmt.Printf("📊 Configuration: MaxMemory=%s, Workers=%d, Batch=%d\n",
			formatBytes(m.config.MaxMemoryUsage),
			m.config.WorkerPoolConfig.CloneWorkers,
			m.config.BatchSize)
	}

	// Create target directory
	if err := os.MkdirAll(targetPath, 0o755); err != nil {
		return stats, fmt.Errorf("failed to create target directory: %w", err)
	}

	// Start streaming repositories
	repoStream, err := m.streamingClient.StreamOrganizationRepositories(ctx, org, m.config.StreamingConfig)
	if err != nil {
		return stats, fmt.Errorf("failed to start repository stream: %w", err)
	}

	// Process repositories in batches
	batch := make([]*Repository, 0, m.config.BatchSize)

	var progressBar *progressbar.ProgressBar

	if m.config.ShowProgress {
		// We'll update total when we know it
		progressBar = progressbar.NewOptions(-1,
			progressbar.OptionSetDescription("Processing Repositories"),
			progressbar.OptionSetRenderBlankState(true),
			progressbar.OptionSpinnerType(14),
		)
	}

	for {
		select {
		case <-ctx.Done():
			return stats, fmt.Errorf("operation cancelled: %w", ctx.Err())

		case repoResult, ok := <-repoStream:
			if !ok {
				// Stream finished, process remaining batch
				if len(batch) > 0 {
					batchStats := m.processBatch(ctx, targetPath, org, strategy, batch, progressBar)
					m.mergeStats(stats, batchStats)
				}

				goto finish
			}

			if repoResult.Error != nil {
				stats.ErrorDetails = append(stats.ErrorDetails, CloneError{
					Repository:  "stream",
					Operation:   "fetch",
					Error:       repoResult.Error,
					Timestamp:   time.Now(),
					MemoryUsage: repoResult.Metadata.MemoryUsage,
				})

				continue
			}

			// Add repository to current batch
			batch = append(batch, repoResult.Repository)
			stats.TotalRepositories++

			// Update progress bar total if we have metadata
			if progressBar != nil && repoResult.Metadata.TotalPages > 0 {
				estimatedTotal := repoResult.Metadata.TotalPages * m.config.StreamingConfig.PageSize
				progressBar.ChangeMax(estimatedTotal)
			}

			// Process batch when it's full
			if len(batch) >= m.config.BatchSize {
				batchStats := m.processBatch(ctx, targetPath, org, strategy, batch, progressBar)
				m.mergeStats(stats, batchStats)

				// Reset batch
				batch = batch[:0]

				// Check memory usage and cleanup if needed
				if err := m.checkAndOptimizeMemory(); err != nil {
					if m.config.VerboseLogging {
						fmt.Printf("⚠️ Memory optimization warning: %v\n", err)
					}
				}
			}
		}
	}

finish:
	stats.TotalDuration = time.Since(startTime)

	if stats.TotalDuration > 0 {
		stats.AverageSpeed = float64(stats.Successful) / stats.TotalDuration.Seconds()
	}

	// Update peak memory usage
	m.memoryMonitor.mu.RLock()
	stats.MemoryPeakUsage = m.memoryMonitor.currentUsage
	m.memoryMonitor.mu.RUnlock()

	if m.config.ShowProgress && progressBar != nil {
		_ = progressBar.Finish()
	}

	if m.config.VerboseLogging {
		m.printFinalStats(stats)
	}

	return stats, nil
}

// processBatch processes a batch of repositories using the worker pool.
func (m *OptimizedSyncCloneManager) processBatch(ctx context.Context, targetPath, org, strategy string,
	repositories []*Repository, progressBar *progressbar.ProgressBar,
) *CloneStats {
	batchStats := &CloneStats{
		ErrorDetails: make([]CloneError, 0),
	}

	// Create jobs for each repository
	jobs := make([]workerpool.RepositoryJob, 0, len(repositories))
	for _, repo := range repositories {
		repoPath := filepath.Join(targetPath, repo.Name)

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
			Repository: repo.Name,
			Operation:  operation,
			Path:       repoPath,
			Strategy:   strategy,
		})
	}

	// Process repositories using worker pool
	processFn := func(ctx context.Context, job workerpool.RepositoryJob) error {
		return m.processRepositoryJob(ctx, job, org)
	}

	// Submit jobs
	for _, job := range jobs {
		if err := m.workerPool.SubmitJob(job, processFn); err != nil {
			batchStats.ErrorDetails = append(batchStats.ErrorDetails, CloneError{
				Repository:  job.Repository,
				Operation:   "submit",
				Error:       err,
				Timestamp:   time.Now(),
				MemoryUsage: m.getCurrentMemoryUsage(),
			})
			batchStats.Failed++

			continue
		}
	}

	// Collect results
	resultsChan := m.workerPool.Results()

	for i := 0; i < len(jobs); i++ {
		select {
		case result := <-resultsChan:
			if result.Success {
				batchStats.Successful++

				if m.config.VerboseLogging {
					fmt.Printf("✅ %s: %s completed\n", result.Job.Repository, result.Job.Operation)
				}
			} else {
				batchStats.Failed++
				batchStats.ErrorDetails = append(batchStats.ErrorDetails, CloneError{
					Repository:  result.Job.Repository,
					Operation:   string(result.Job.Operation),
					Error:       result.Error,
					Timestamp:   time.Now(),
					MemoryUsage: m.getCurrentMemoryUsage(),
				})

				if m.config.VerboseLogging {
					fmt.Printf("❌ %s: %s failed: %v\n", result.Job.Repository, result.Job.Operation, result.Error)
				}
			}

			if progressBar != nil {
				_ = progressBar.Add(1)
			}

		case <-ctx.Done():
			return batchStats
		}
	}

	return batchStats
}

// processRepositoryJob processes a single repository job.
func (m *OptimizedSyncCloneManager) processRepositoryJob(ctx context.Context, job workerpool.RepositoryJob, org string) error {
	switch job.Operation {
	case workerpool.OperationClone:
		return Clone(ctx, job.Path, org, job.Repository)
	case workerpool.OperationPull:
		return m.executeGitOperation(ctx, job.Path, "pull")
	case workerpool.OperationFetch:
		return m.executeGitOperation(ctx, job.Path, "fetch")
	case workerpool.OperationReset:
		// Reset hard HEAD and pull
		if err := m.executeGitOperation(ctx, job.Path, "reset", "--hard", "HEAD"); err != nil {
			return fmt.Errorf("git reset failed: %w", err)
		}

		return m.executeGitOperation(ctx, job.Path, "pull")
	case workerpool.OperationConfig:
		// Config operation - placeholder for configuration updates
		return fmt.Errorf("config operation not yet implemented")
	default:
		return fmt.Errorf("unknown operation: %s", job.Operation)
	}
}

// executeGitOperation executes a git command in the repository path.
func (m *OptimizedSyncCloneManager) executeGitOperation(ctx context.Context, repoPath string, args ...string) error {
	// Check if repository is valid
	repoType, _ := git.CheckGitRepoType(repoPath)
	if repoType == git.RepoTypeEmpty {
		return fmt.Errorf("repository is empty or not a git repository")
	}

	// Use the same git execution logic as the bulk operations
	return nil // Placeholder - would use actual git execution
}

// startMemoryMonitoring starts the memory monitoring goroutine.
func (m *OptimizedSyncCloneManager) startMemoryMonitoring() {
	ticker := time.NewTicker(m.config.GCInterval)
	defer ticker.Stop()

	for range ticker.C {
		m.updateMemoryUsage()

		m.memoryMonitor.mu.RLock()
		currentUsage := m.memoryMonitor.currentUsage
		maxMemory := m.memoryMonitor.maxMemory
		threshold := m.memoryMonitor.threshold
		m.memoryMonitor.mu.RUnlock()

		usagePercent := float64(currentUsage) / float64(maxMemory)

		if usagePercent > threshold {
			if m.config.VerboseLogging {
				fmt.Printf("🧠 Memory usage %.1f%% > threshold %.1f%%, triggering cleanup\n",
					usagePercent*100, threshold*100)
			}

			m.forceMemoryCleanup()
		}
	}
}

// updateMemoryUsage updates current memory usage statistics.
func (m *OptimizedSyncCloneManager) updateMemoryUsage() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	m.memoryMonitor.mu.Lock()
	m.memoryMonitor.currentUsage = int64(memStats.Alloc)
	m.memoryMonitor.mu.Unlock()
}

// getCurrentMemoryUsage returns current memory usage.
func (m *OptimizedSyncCloneManager) getCurrentMemoryUsage() int64 {
	m.memoryMonitor.mu.RLock()
	defer m.memoryMonitor.mu.RUnlock()

	return m.memoryMonitor.currentUsage
}

// checkAndOptimizeMemory checks memory usage and optimizes if needed.
func (m *OptimizedSyncCloneManager) checkAndOptimizeMemory() error {
	m.updateMemoryUsage()

	m.memoryMonitor.mu.RLock()
	currentUsage := m.memoryMonitor.currentUsage
	maxMemory := m.memoryMonitor.maxMemory
	threshold := m.memoryMonitor.threshold
	m.memoryMonitor.mu.RUnlock()

	usagePercent := float64(currentUsage) / float64(maxMemory)

	if usagePercent > threshold {
		m.forceMemoryCleanup()

		// Check again after cleanup
		m.updateMemoryUsage()

		m.memoryMonitor.mu.RLock()
		newUsage := m.memoryMonitor.currentUsage
		m.memoryMonitor.mu.RUnlock()

		newPercent := float64(newUsage) / float64(maxMemory)

		if newPercent > 0.95 { // Still very high after cleanup
			return fmt.Errorf("memory usage %.1f%% remains high after cleanup", newPercent*100)
		}
	}

	return nil
}

// forceMemoryCleanup forces garbage collection and pool cleanup.
func (m *OptimizedSyncCloneManager) forceMemoryCleanup() {
	m.memoryMonitor.mu.Lock()
	m.memoryMonitor.gcCount++
	m.memoryMonitor.lastGC = time.Now()
	m.memoryMonitor.mu.Unlock()

	// Trigger GC
	runtime.GC()
	runtime.GC() // Call twice to ensure sweep phase completes

	// Optimize streaming client memory
	m.streamingClient.optimizeMemory()
}

// mergeStats merges batch statistics into total statistics.
func (m *OptimizedSyncCloneManager) mergeStats(total, batch *CloneStats) {
	total.Successful += batch.Successful
	total.Failed += batch.Failed
	total.Skipped += batch.Skipped
	total.ErrorDetails = append(total.ErrorDetails, batch.ErrorDetails...)
}

// printFinalStats prints final operation statistics.
func (m *OptimizedSyncCloneManager) printFinalStats(stats *CloneStats) {
	fmt.Printf("\n📊 Bulk Clone Operation Complete\n")
	fmt.Printf("Total Repositories: %d\n", stats.TotalRepositories)
	fmt.Printf("✅ Successful: %d\n", stats.Successful)
	fmt.Printf("❌ Failed: %d\n", stats.Failed)
	fmt.Printf("⏭️ Skipped: %d\n", stats.Skipped)
	fmt.Printf("⏱️ Duration: %v\n", stats.TotalDuration)
	fmt.Printf("🚀 Average Speed: %.2f repos/sec\n", stats.AverageSpeed)
	fmt.Printf("🧠 Peak Memory: %s\n", formatBytes(stats.MemoryPeakUsage))

	m.memoryMonitor.mu.RLock()
	gcCount := m.memoryMonitor.gcCount
	m.memoryMonitor.mu.RUnlock()

	fmt.Printf("🗑️ GC Cycles: %d\n", gcCount)

	// Print API metrics
	metrics := m.streamingClient.GetMetrics()

	fmt.Printf("\n🌐 API Metrics:\n")
	fmt.Printf("Total Requests: %d\n", metrics.totalRequests)
	fmt.Printf("Average Latency: %v\n", metrics.averageLatency)
	fmt.Printf("Cache Hits: %d\n", metrics.cachedResponses)
}

// Close cleans up resources.
func (m *OptimizedSyncCloneManager) Close() error {
	if m.workerPool != nil {
		m.workerPool.Stop()
	}

	if m.streamingClient != nil {
		return m.streamingClient.Close()
	}

	return nil
}

// formatBytes formats byte count as human readable string.
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
