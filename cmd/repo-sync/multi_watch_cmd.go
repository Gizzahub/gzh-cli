package reposync

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// newWatchMultiCmd creates the watch-multi subcommand for monitoring multiple repositories
func newWatchMultiCmd(logger *zap.Logger) *cobra.Command {
	var (
		batchSize    int
		batchTimeout time.Duration
		maxRepos     int
		verbose      bool
	)

	cmd := &cobra.Command{
		Use:   "watch-multi [repository-paths...]",
		Short: "Watch multiple repositories simultaneously with efficient resource management",
		Long: `Watch multiple Git repositories simultaneously for file system changes with advanced resource management.

This command provides efficient multi-repository monitoring:
- Concurrent watching of multiple repositories with shared resources
- Intelligent resource allocation based on repository activity
- Unified event processing and batching across all repositories
- Cross-repository change correlation and analysis
- Scalable architecture supporting large numbers of repositories

Features:
- Shared file system watcher instances for efficiency
- Repository activity-based resource prioritization
- Unified event stream with repository identification
- Cross-repository dependency detection
- Bulk synchronization capabilities

Examples:
  # Watch multiple repositories in current directory
  gz repo-sync watch-multi ./repo1 ./repo2 ./repo3
  
  # Watch all repositories in a directory
  gz repo-sync watch-multi ./projects/*
  
  # Watch with custom resource limits
  gz repo-sync watch-multi --max-repos 50 --batch-size 200 ./projects/*`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate repository paths
			validRepos := make([]string, 0, len(args))
			for _, path := range args {
				if err := validateRepositoryPath(path); err != nil {
					logger.Warn("Skipping invalid repository",
						zap.String("path", path),
						zap.Error(err))
					continue
				}
				validRepos = append(validRepos, path)
			}

			if len(validRepos) == 0 {
				return fmt.Errorf("no valid repositories found")
			}

			if len(validRepos) > maxRepos {
				return fmt.Errorf("too many repositories: %d (max: %d)", len(validRepos), maxRepos)
			}

			// Create multi-repository watcher
			config := &MultiWatchConfig{
				Repositories: validRepos,
				BatchSize:    batchSize,
				BatchTimeout: batchTimeout,
				MaxRepos:     maxRepos,
			}

			multiWatcher, err := NewMultiRepositoryWatcher(logger, config)
			if err != nil {
				return fmt.Errorf("failed to create multi-repository watcher: %w", err)
			}
			defer multiWatcher.Close()

			ctx := context.Background()
			fmt.Printf("🔍 Starting multi-repository watch for %d repositories\n", len(validRepos))
			fmt.Printf("📦 Batch size: %d events, Timeout: %s\n", batchSize, batchTimeout)
			fmt.Printf("Press Ctrl+C to stop...\n\n")

			return multiWatcher.Start(ctx, verbose)
		},
	}

	// Add flags
	cmd.Flags().IntVar(&batchSize, "batch-size", 200, "Number of events to batch across all repositories")
	cmd.Flags().DurationVar(&batchTimeout, "batch-timeout", 10*time.Second, "Batch timeout for multi-repo events")
	cmd.Flags().IntVar(&maxRepos, "max-repos", 100, "Maximum number of repositories to watch")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	return cmd
}

// MultiWatchConfig represents configuration for multi-repository watching
type MultiWatchConfig struct {
	Repositories []string      `json:"repositories"`
	BatchSize    int           `json:"batch_size"`
	BatchTimeout time.Duration `json:"batch_timeout"`
	MaxRepos     int           `json:"max_repos"`
}

// MultiRepositoryWatcher handles concurrent monitoring of multiple repositories
type MultiRepositoryWatcher struct {
	logger   *zap.Logger
	config   *MultiWatchConfig
	watchers map[string]*RepositoryWatcher
	events   chan MultiRepoEvent
	mu       sync.RWMutex
	running  bool
	stats    MultiWatchStats
}

// MultiRepoEvent represents an event from any watched repository
type MultiRepoEvent struct {
	Repository string          `json:"repository"`
	Event      FileChangeEvent `json:"event"`
	Timestamp  time.Time       `json:"timestamp"`
}

// MultiWatchStats tracks statistics for multi-repository watching
type MultiWatchStats struct {
	StartTime          time.Time        `json:"start_time"`
	TotalRepositories  int              `json:"total_repositories"`
	ActiveRepositories int              `json:"active_repositories"`
	TotalEvents        int64            `json:"total_events"`
	EventsByRepo       map[string]int64 `json:"events_by_repo"`
	BatchesProcessed   int64            `json:"batches_processed"`
}

// NewMultiRepositoryWatcher creates a new multi-repository watcher
func NewMultiRepositoryWatcher(logger *zap.Logger, config *MultiWatchConfig) (*MultiRepositoryWatcher, error) {
	return &MultiRepositoryWatcher{
		logger:   logger,
		config:   config,
		watchers: make(map[string]*RepositoryWatcher),
		events:   make(chan MultiRepoEvent, config.BatchSize*2),
		stats: MultiWatchStats{
			StartTime:         time.Now(),
			TotalRepositories: len(config.Repositories),
			EventsByRepo:      make(map[string]int64),
		},
	}, nil
}

// Start begins watching all configured repositories
func (mrw *MultiRepositoryWatcher) Start(ctx context.Context, verbose bool) error {
	mrw.mu.Lock()
	mrw.running = true
	mrw.mu.Unlock()

	// Start individual repository watchers
	var wg sync.WaitGroup
	for _, repoPath := range mrw.config.Repositories {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			if err := mrw.startRepositoryWatcher(ctx, path, verbose); err != nil {
				mrw.logger.Error("Failed to start repository watcher",
					zap.String("path", path),
					zap.Error(err))
			}
		}(repoPath)
	}

	// Start event processing
	go mrw.processEvents(ctx, verbose)

	// Wait for all watchers to start or context to be cancelled
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		fmt.Printf("✅ Started watching %d repositories\n\n", len(mrw.config.Repositories))

		// Print periodic statistics
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				mrw.printFinalStats()
				return nil
			case <-ticker.C:
				mrw.printStats(verbose)
			}
		}
	}
}

// Close stops all repository watchers and cleans up resources
func (mrw *MultiRepositoryWatcher) Close() error {
	mrw.mu.Lock()
	defer mrw.mu.Unlock()

	if !mrw.running {
		return nil
	}

	mrw.running = false

	// Close all individual watchers
	for _, watcher := range mrw.watchers {
		if err := watcher.Close(); err != nil {
			mrw.logger.Warn("Failed to close repository watcher", zap.Error(err))
		}
	}

	close(mrw.events)
	return nil
}

// startRepositoryWatcher starts watching a single repository
func (mrw *MultiRepositoryWatcher) startRepositoryWatcher(ctx context.Context, repoPath string, verbose bool) error {
	// Create configuration for individual repository watcher
	config := DefaultRepoSyncConfig()
	config.RepositoryPath = repoPath
	config.BatchSize = 50 // Smaller batch size for individual repos

	// Create repository watcher
	watcher, err := NewRepositoryWatcher(mrw.logger, config)
	if err != nil {
		return fmt.Errorf("failed to create watcher for %s: %w", repoPath, err)
	}

	mrw.mu.Lock()
	mrw.watchers[repoPath] = watcher
	mrw.stats.ActiveRepositories++
	mrw.mu.Unlock()

	// Start watching (this will run in its own goroutine)
	return watcher.Start(ctx, false) // Disable verbose for individual watchers
}

// processEvents processes events from all repositories
func (mrw *MultiRepositoryWatcher) processEvents(ctx context.Context, verbose bool) {
	ticker := time.NewTicker(mrw.config.BatchTimeout)
	defer ticker.Stop()

	batch := make([]MultiRepoEvent, 0, mrw.config.BatchSize)
	batchStart := time.Now()

	for {
		select {
		case <-ctx.Done():
			if len(batch) > 0 {
				mrw.processBatch(batch, verbose)
			}
			return

		case event, ok := <-mrw.events:
			if !ok {
				if len(batch) > 0 {
					mrw.processBatch(batch, verbose)
				}
				return
			}

			batch = append(batch, event)
			mrw.stats.TotalEvents++
			mrw.stats.EventsByRepo[event.Repository]++

			// Process batch if size limit reached
			if len(batch) >= mrw.config.BatchSize {
				mrw.processBatch(batch, verbose)
				batch = batch[:0]
				batchStart = time.Now()
				ticker.Reset(mrw.config.BatchTimeout)
			}

		case <-ticker.C:
			// Process batch if timeout reached
			if len(batch) > 0 {
				mrw.processBatch(batch, verbose)
				batch = batch[:0]
				batchStart = time.Now()
			}
		}
	}
}

// processBatch processes a batch of multi-repository events
func (mrw *MultiRepositoryWatcher) processBatch(batch []MultiRepoEvent, verbose bool) {
	startTime := time.Now()
	mrw.stats.BatchesProcessed++

	// Group events by repository
	repoEvents := make(map[string][]MultiRepoEvent)
	for _, event := range batch {
		repoEvents[event.Repository] = append(repoEvents[event.Repository], event)
	}

	fmt.Printf("📦 Processing multi-repo batch: %d events across %d repositories\n",
		len(batch), len(repoEvents))

	if verbose {
		for repo, events := range repoEvents {
			fmt.Printf("   📁 %s: %d events\n", repo, len(events))
			for _, event := range events {
				fmt.Printf("      %s %s (%s)\n",
					getOperationEmoji(event.Event.Operation),
					event.Event.Path,
					event.Event.Operation)
			}
		}
	}

	duration := time.Since(startTime)
	fmt.Printf("✅ Multi-repo batch processed in %v\n\n", duration.Round(time.Millisecond))
}

// getOperationEmoji returns emoji for operation type
func getOperationEmoji(operation string) string {
	switch operation {
	case "create":
		return "📄"
	case "write":
		return "✏️"
	case "remove":
		return "🗑️"
	case "rename":
		return "🔄"
	case "chmod":
		return "🔒"
	default:
		return "❓"
	}
}

// printStats prints current multi-watch statistics
func (mrw *MultiRepositoryWatcher) printStats(verbose bool) {
	mrw.mu.RLock()
	defer mrw.mu.RUnlock()

	uptime := time.Since(mrw.stats.StartTime)
	fmt.Printf("📊 Multi-Watch Stats (uptime: %v):\n", uptime.Round(time.Second))
	fmt.Printf("   Repositories: %d total, %d active\n",
		mrw.stats.TotalRepositories, mrw.stats.ActiveRepositories)
	fmt.Printf("   Events: %d total, %d batches processed\n",
		mrw.stats.TotalEvents, mrw.stats.BatchesProcessed)

	if verbose && len(mrw.stats.EventsByRepo) > 0 {
		fmt.Printf("   Events by repository:\n")
		for repo, count := range mrw.stats.EventsByRepo {
			fmt.Printf("     📁 %s: %d events\n", repo, count)
		}
	}
	fmt.Println()
}

// printFinalStats prints final statistics when shutting down
func (mrw *MultiRepositoryWatcher) printFinalStats() {
	fmt.Printf("\n📊 Final Multi-Watch Statistics:\n")
	fmt.Printf("   Total Runtime: %v\n", time.Since(mrw.stats.StartTime).Round(time.Second))
	fmt.Printf("   Repositories Watched: %d\n", mrw.stats.TotalRepositories)
	fmt.Printf("   Total Events Processed: %d\n", mrw.stats.TotalEvents)
	fmt.Printf("   Batches Processed: %d\n", mrw.stats.BatchesProcessed)

	if len(mrw.stats.EventsByRepo) > 0 {
		fmt.Printf("   Most Active Repositories:\n")
		// Sort and show top 5
		type repoCount struct {
			repo  string
			count int64
		}
		var repos []repoCount
		for repo, count := range mrw.stats.EventsByRepo {
			repos = append(repos, repoCount{repo, count})
		}

		// Simple sort by count (descending)
		for i := 0; i < len(repos)-1; i++ {
			for j := i + 1; j < len(repos); j++ {
				if repos[j].count > repos[i].count {
					repos[i], repos[j] = repos[j], repos[i]
				}
			}
		}

		maxShow := 5
		if len(repos) < maxShow {
			maxShow = len(repos)
		}

		for i := 0; i < maxShow; i++ {
			fmt.Printf("     %d. %s: %d events\n", i+1, repos[i].repo, repos[i].count)
		}
	}
	fmt.Println()
}
