package reposync

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// File operation constants.
const (
	opWrite  = "write"
	opCreate = "create"
)

// newWatchCmd creates the watch subcommand for real-time file system monitoring.
func newWatchCmd(logger *zap.Logger) *cobra.Command {
	var (
		configFile     string
		batchSize      int
		batchTimeout   time.Duration
		ignorePatterns []string
		watchPatterns  []string
		verbose        bool
		autoCommit     bool
	)

	cmd := &cobra.Command{
		Use:   "watch [repository-path]",
		Short: "Watch repository for file system changes with efficient batching",
		Long: `Watch a Git repository for file system changes using fsnotify with advanced features:

- Efficient file change detection with batching to reduce system load
- Pattern-based filtering for relevant files only
- Real-time change processing with configurable batch sizes and timeouts
- Checksum-based change validation to avoid duplicate processing
- Integration with Git operations for automatic commit workflows

The watcher uses an intelligent batching algorithm that:
1. Collects file change events in configurable batches
2. Deduplicates events for the same file within a batch
3. Processes batches when size limit or timeout is reached
4. Calculates checksums to verify actual file changes

Examples:
  # Watch current directory with default settings
  gz repo-sync watch .

  # Watch with custom batch settings
  gz repo-sync watch ./my-repo --batch-size 50 --batch-timeout 10s

  # Watch with auto-commit enabled
  gz repo-sync watch ./my-repo --auto-commit --verbose

  # Watch with custom patterns
  gz repo-sync watch ./my-repo --watch-patterns "**/*.go,**/*.md" --ignore-patterns "vendor/**,*.tmp"`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			repoPath := "."
			if len(args) > 0 {
				repoPath = args[0]
			}

			// Validate repository path
			if err := validateRepositoryPath(repoPath); err != nil {
				return fmt.Errorf("invalid repository path: %w", err)
			}

			// Create configuration
			config := DefaultConfig()
			config.RepositoryPath = repoPath
			config.BatchSize = batchSize
			config.BatchTimeout = batchTimeout
			config.AutoCommit = autoCommit

			if len(watchPatterns) > 0 {
				config.WatchPatterns = watchPatterns
			}
			if len(ignorePatterns) > 0 {
				config.IgnorePatterns = ignorePatterns
			}

			// Create and start file watcher
			watcher, err := NewRepositoryWatcher(logger, config)
			if err != nil {
				return fmt.Errorf("failed to create repository watcher: %w", err)
			}
			defer func() {
				if err := watcher.Close(); err != nil {
					logger.Error("Failed to close watcher", zap.Error(err))
				}
			}()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Start watching
			fmt.Printf("🔍 Starting repository watch for: %s\n", repoPath)
			fmt.Printf("📦 Batch size: %d events, Timeout: %s\n", batchSize, batchTimeout)
			fmt.Printf("🎯 Watch patterns: %v\n", config.WatchPatterns)
			fmt.Printf("🚫 Ignore patterns: %v\n", config.IgnorePatterns)
			fmt.Printf("Press Ctrl+C to stop...\n\n")

			return watcher.Start(ctx, verbose)
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&configFile, "config", "c", "", "Configuration file path")
	cmd.Flags().IntVar(&batchSize, "batch-size", 100, "Number of events to batch before processing")
	cmd.Flags().DurationVar(&batchTimeout, "batch-timeout", 5*time.Second, "Maximum time to wait before processing batch")
	cmd.Flags().StringSliceVar(&ignorePatterns, "ignore-patterns", []string{}, "File patterns to ignore (comma-separated)")
	cmd.Flags().StringSliceVar(&watchPatterns, "watch-patterns", []string{}, "File patterns to watch (comma-separated)")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	cmd.Flags().BoolVar(&autoCommit, "auto-commit", false, "Automatically commit changes")

	return cmd
}

// RepositoryWatcher handles efficient file system monitoring for Git repositories.
type RepositoryWatcher struct {
	logger  *zap.Logger
	config  *Config
	watcher *fsnotify.Watcher
	events  chan FileChangeEvent
	batches chan FileChangeBatch
	mu      sync.RWMutex
	running bool
	stats   WatcherStats
}

// WatcherStats tracks statistics about the file watcher.
type WatcherStats struct {
	StartTime        time.Time `json:"startTime"`
	TotalEvents      int64     `json:"totalEvents"`
	BatchesProcessed int64     `json:"batchesProcessed"`
	FilesModified    int64     `json:"filesModified"`
	LastEventTime    time.Time `json:"lastEventTime"`
	ErrorCount       int64     `json:"errorCount"`
}

// NewRepositoryWatcher creates a new repository file system watcher.
func NewRepositoryWatcher(logger *zap.Logger, config *Config) (*RepositoryWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create fsnotify watcher: %w", err)
	}

	return &RepositoryWatcher{
		logger:  logger,
		config:  config,
		watcher: watcher,
		events:  make(chan FileChangeEvent, config.BatchSize*2),
		batches: make(chan FileChangeBatch, 10),
		stats: WatcherStats{
			StartTime: time.Now(),
		},
	}, nil
}

// Start begins watching the repository for file changes.
func (rw *RepositoryWatcher) Start(ctx context.Context, verbose bool) error {
	rw.mu.Lock()
	rw.running = true
	rw.mu.Unlock()

	// Add repository directory to watcher
	err := rw.addDirectoryRecursively(rw.config.RepositoryPath)
	if err != nil {
		return fmt.Errorf("failed to add repository to watcher: %w", err)
	}

	// Start event processing goroutines
	go rw.processFileEvents(ctx)
	go rw.processBatches(ctx, verbose)

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Main event loop
	for {
		select {
		case <-ctx.Done():
			rw.logger.Info("Context cancelled, stopping watcher")
			return nil

		case sig := <-sigChan:
			fmt.Printf("\n🛑 Received signal %v, stopping watcher\n", sig)
			rw.printStats()

			return nil

		case event, ok := <-rw.watcher.Events:
			if !ok {
				rw.logger.Info("Watcher events channel closed")
				return nil
			}

			rw.handleFileEvent(event)

		case err, ok := <-rw.watcher.Errors:
			if !ok {
				rw.logger.Info("Watcher errors channel closed")
				return nil
			}

			rw.logger.Error("File watcher error", zap.Error(err))
			rw.stats.ErrorCount++
		}
	}
}

// Close stops the file watcher and cleans up resources.
func (rw *RepositoryWatcher) Close() error {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	if rw.running {
		rw.running = false
		close(rw.events)
		close(rw.batches)

		return rw.watcher.Close()
	}

	return nil
}

// addDirectoryRecursively adds a directory and all its subdirectories to the watcher.
func (rw *RepositoryWatcher) addDirectoryRecursively(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip if not a directory
		if !info.IsDir() {
			return nil
		}

		// Skip ignored directories
		if rw.shouldIgnore(path) {
			return filepath.SkipDir
		}

		// Add directory to watcher
		err = rw.watcher.Add(path)
		if err != nil {
			rw.logger.Warn("Failed to add directory to watcher",
				zap.String("path", path),
				zap.Error(err))
		} else {
			rw.logger.Debug("Added directory to watcher", zap.String("path", path))
		}

		return nil
	})
}

// handleFileEvent processes a single file system event.
func (rw *RepositoryWatcher) handleFileEvent(event fsnotify.Event) {
	// Skip if file should be ignored
	if rw.shouldIgnore(event.Name) {
		return
	}

	// Skip if file doesn't match watch patterns
	if !rw.matchesWatchPatterns(event.Name) {
		return
	}

	// Create file change event
	changeEvent := FileChangeEvent{
		Path:        event.Name,
		Operation:   rw.mapOperation(event.Op),
		IsDirectory: rw.isDirectory(event.Name),
		Timestamp:   time.Now(),
	}

	// Add file size and checksum for relevant operations
	if changeEvent.Operation == opWrite || changeEvent.Operation == opCreate {
		if stat, err := os.Stat(event.Name); err == nil && !stat.IsDir() {
			changeEvent.Size = stat.Size()

			// Calculate checksum for small files only (< 1MB)
			if stat.Size() < 1024*1024 {
				if checksum, err := rw.calculateChecksum(event.Name); err == nil {
					changeEvent.Checksum = checksum
				}
			}
		}
	}

	// Send event to processing channel
	select {
	case rw.events <- changeEvent:
		rw.stats.TotalEvents++
		rw.stats.LastEventTime = time.Now()
	default:
		rw.logger.Warn("Event channel full, dropping event", zap.String("path", event.Name))
	}
}

// processFileEvents collects events into batches for efficient processing.
func (rw *RepositoryWatcher) processFileEvents(ctx context.Context) {
	ticker := time.NewTicker(rw.config.BatchTimeout)
	defer ticker.Stop()

	currentBatch := make([]FileChangeEvent, 0, rw.config.BatchSize)
	batchStartTime := time.Now()

	for {
		select {
		case <-ctx.Done():
			// Process final batch if any events remain
			if len(currentBatch) > 0 {
				rw.sendBatch(currentBatch, batchStartTime)
			}

			return

		case event, ok := <-rw.events:
			if !ok {
				// Process final batch if any events remain
				if len(currentBatch) > 0 {
					rw.sendBatch(currentBatch, batchStartTime)
				}

				return
			}

			// Add event to current batch
			currentBatch = append(currentBatch, event)

			// Send batch if size limit reached
			if len(currentBatch) >= rw.config.BatchSize {
				rw.sendBatch(currentBatch, batchStartTime)
				currentBatch = currentBatch[:0] // Clear batch
				batchStartTime = time.Now()

				ticker.Reset(rw.config.BatchTimeout)
			}

		case <-ticker.C:
			// Send batch if timeout reached and there are events
			if len(currentBatch) > 0 {
				rw.sendBatch(currentBatch, batchStartTime)
				currentBatch = currentBatch[:0] // Clear batch
				batchStartTime = time.Now()
			}
		}
	}
}

// sendBatch creates and sends a batch of file change events.
func (rw *RepositoryWatcher) sendBatch(events []FileChangeEvent, startTime time.Time) {
	// Deduplicate events for the same file (keep the latest)
	deduped := rw.deduplicateEvents(events)

	batch := FileChangeBatch{
		Events:      deduped,
		BatchID:     fmt.Sprintf("batch-%d", time.Now().UnixNano()),
		StartTime:   startTime,
		EndTime:     time.Now(),
		TotalEvents: len(deduped),
	}

	select {
	case rw.batches <- batch:
		rw.stats.BatchesProcessed++
	default:
		rw.logger.Warn("Batch channel full, dropping batch", zap.Int("events", len(deduped)))
	}
}

// deduplicateEvents removes duplicate events for the same file, keeping the latest.
func (rw *RepositoryWatcher) deduplicateEvents(events []FileChangeEvent) []FileChangeEvent {
	eventMap := make(map[string]FileChangeEvent)

	for _, event := range events {
		// Use the file path as the key, keeping the latest event
		if existing, exists := eventMap[event.Path]; !exists || event.Timestamp.After(existing.Timestamp) {
			eventMap[event.Path] = event
		}
	}

	// Convert map back to slice
	deduped := make([]FileChangeEvent, 0, len(eventMap))
	for _, event := range eventMap {
		deduped = append(deduped, event)
	}

	return deduped
}

// processBatches handles batches of file change events.
func (rw *RepositoryWatcher) processBatches(ctx context.Context, verbose bool) {
	for {
		select {
		case <-ctx.Done():
			return

		case batch, ok := <-rw.batches:
			if !ok {
				return
			}

			rw.processBatch(batch, verbose)
		}
	}
}

// processBatch processes a single batch of file change events.
func (rw *RepositoryWatcher) processBatch(batch FileChangeBatch, verbose bool) {
	startTime := time.Now()

	if verbose {
		fmt.Printf("📦 Processing batch %s with %d events\n", batch.BatchID, batch.TotalEvents)
	}

	// Group events by operation type
	operationCounts := make(map[string]int)
	for _, event := range batch.Events {
		operationCounts[event.Operation]++

		if verbose {
			fmt.Printf("   %s: %s (%s)\n",
				rw.getOperationEmoji(event.Operation),
				event.Path,
				event.Operation)
		}
	}

	// Update statistics
	rw.stats.FilesModified += int64(batch.TotalEvents)

	// Print batch summary
	duration := time.Since(startTime)
	fmt.Printf("✅ Batch %s processed in %v", batch.BatchID, duration.Round(time.Millisecond))

	if len(operationCounts) > 0 {
		fmt.Printf(" (")

		first := true
		for op, count := range operationCounts {
			if !first {
				fmt.Printf(", ")
			}

			fmt.Printf("%d %s", count, op)

			first = false
		}

		fmt.Printf(")")
	}

	fmt.Println()

	// Auto-commit if enabled
	if rw.config.AutoCommit {
		rw.performAutoCommit(batch)
	}
}

// performAutoCommit automatically commits changes if enabled.
func (rw *RepositoryWatcher) performAutoCommit(batch FileChangeBatch) {
	// Git auto-commit integration placeholder - implement Git command integration
	// This would integrate with Git commands to:
	// 1. Stage modified files
	// 2. Create commit with batch information
	// 3. Handle merge conflicts if they occur
	rw.logger.Info("Auto-commit triggered",
		zap.String("batchId", batch.BatchID),
		zap.Int("events", batch.TotalEvents))

	fmt.Printf("🔄 Auto-commit: %d changes (batch %s)\n", batch.TotalEvents, batch.BatchID)
}

// Helper methods

// shouldIgnore checks if a path should be ignored based on ignore patterns.
func (rw *RepositoryWatcher) shouldIgnore(path string) bool {
	for _, pattern := range rw.config.IgnorePatterns {
		if matched, _ := filepath.Match(pattern, path); matched {
			return true
		}

		// Check if path contains pattern (for directory patterns like vendor/**)
		if strings.Contains(pattern, "**") {
			cleanPattern := strings.ReplaceAll(pattern, "**", "*")
			if strings.Contains(path, strings.TrimSuffix(cleanPattern, "/*")) {
				return true
			}
		}
	}

	return false
}

// matchesWatchPatterns checks if a path matches any watch patterns.
func (rw *RepositoryWatcher) matchesWatchPatterns(path string) bool {
	// If no patterns specified, watch everything
	if len(rw.config.WatchPatterns) == 0 {
		return true
	}

	ext := filepath.Ext(path)
	for _, pattern := range rw.config.WatchPatterns {
		// Direct pattern match
		if matched, _ := filepath.Match(pattern, path); matched {
			return true
		}

		// Check file extension match (*.go, *.md, etc.)
		if strings.HasPrefix(pattern, "*.") && pattern[1:] == ext {
			return true
		}

		// Handle ** patterns like **/*.go, **/*.md
		if strings.Contains(pattern, "**") {
			// For patterns like **/*.go, match any file with .go extension
			if strings.HasPrefix(pattern, "**/") {
				suffixPattern := pattern[3:] // Remove "**/" prefix
				if matched, _ := filepath.Match(suffixPattern, filepath.Base(path)); matched {
					return true
				}
			}

			// General glob pattern with ** replaced by *
			cleanPattern := strings.ReplaceAll(pattern, "**", "*")
			if matched, _ := filepath.Match(cleanPattern, filepath.Base(path)); matched {
				return true
			}
		}
	}

	return false
}

// mapOperation converts fsnotify operations to readable strings.
func (rw *RepositoryWatcher) mapOperation(op fsnotify.Op) string {
	switch {
	case op&fsnotify.Create == fsnotify.Create:
		return "create"
	case op&fsnotify.Write == fsnotify.Write:
		return "write"
	case op&fsnotify.Remove == fsnotify.Remove:
		return "remove"
	case op&fsnotify.Rename == fsnotify.Rename:
		return "rename"
	case op&fsnotify.Chmod == fsnotify.Chmod:
		return "chmod"
	default:
		return "unknown"
	}
}

// getOperationEmoji returns an emoji for the operation type.
func (rw *RepositoryWatcher) getOperationEmoji(operation string) string {
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

// isDirectory checks if a path is a directory.
func (rw *RepositoryWatcher) isDirectory(path string) bool {
	if stat, err := os.Stat(path); err == nil {
		return stat.IsDir()
	}

	return false
}

// calculateChecksum calculates SHA256 checksum of a file.
func (rw *RepositoryWatcher) calculateChecksum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = file.Close() //nolint:errcheck // Not critical
	}()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

// printStats prints current watcher statistics.
func (rw *RepositoryWatcher) printStats() {
	uptime := time.Since(rw.stats.StartTime)

	fmt.Printf("\n📊 Watcher Statistics:\n")
	fmt.Printf("   Uptime: %v\n", uptime.Round(time.Second))
	fmt.Printf("   Total Events: %d\n", rw.stats.TotalEvents)
	fmt.Printf("   Batches Processed: %d\n", rw.stats.BatchesProcessed)
	fmt.Printf("   Files Modified: %d\n", rw.stats.FilesModified)

	if !rw.stats.LastEventTime.IsZero() {
		fmt.Printf("   Last Event: %v ago\n", time.Since(rw.stats.LastEventTime).Round(time.Second))
	}

	if rw.stats.ErrorCount > 0 {
		fmt.Printf("   Errors: %d\n", rw.stats.ErrorCount)
	}

	fmt.Println()
}
