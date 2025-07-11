package monitoring

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// LogRetentionManager manages log retention policies
type LogRetentionManager struct {
	config   *LogRetentionConfig
	logger   *zap.Logger
	policies map[string]*RetentionPolicy
	metrics  *RetentionMetrics
	mutex    sync.RWMutex

	// Context and cancellation
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// LogRetentionConfig represents the log retention configuration
type LogRetentionConfig struct {
	Enabled               bool                        `yaml:"enabled" json:"enabled"`
	CheckInterval         time.Duration               `yaml:"check_interval" json:"check_interval"`
	DefaultPolicy         *RetentionPolicy            `yaml:"default_policy" json:"default_policy"`
	Policies              map[string]*RetentionPolicy `yaml:"policies" json:"policies"`
	CompressionEnabled    bool                        `yaml:"compression_enabled" json:"compression_enabled"`
	CompressionDelay      time.Duration               `yaml:"compression_delay" json:"compression_delay"`
	ArchiveLocation       string                      `yaml:"archive_location" json:"archive_location"`
	BackupBeforeDelete    bool                        `yaml:"backup_before_delete" json:"backup_before_delete"`
	BackupLocation        string                      `yaml:"backup_location" json:"backup_location"`
	DryRun                bool                        `yaml:"dry_run" json:"dry_run"`
	MaxConcurrentCleanup  int                         `yaml:"max_concurrent_cleanup" json:"max_concurrent_cleanup"`
	NotificationThreshold int64                       `yaml:"notification_threshold" json:"notification_threshold"`
}

// ExtendedRetentionPolicy extends the basic RetentionPolicy with more options
type ExtendedRetentionPolicy struct {
	*RetentionPolicy
	Name           string        `yaml:"name" json:"name"`
	LogLevels      []string      `yaml:"log_levels" json:"log_levels"`           // Only apply to specific log levels
	LogSources     []string      `yaml:"log_sources" json:"log_sources"`         // Only apply to specific sources
	CompressionAge time.Duration `yaml:"compression_age" json:"compression_age"` // Compress logs older than this
	ArchiveAge     time.Duration `yaml:"archive_age" json:"archive_age"`         // Archive logs older than this
	Priority       int           `yaml:"priority" json:"priority"`               // Policy priority (higher = more important)
	Enabled        bool          `yaml:"enabled" json:"enabled"`
	CreatedAt      time.Time     `yaml:"created_at" json:"created_at"`
	UpdatedAt      time.Time     `yaml:"updated_at" json:"updated_at"`
	LastExecuted   time.Time     `yaml:"last_executed" json:"last_executed"`
	ExecutionCount int64         `yaml:"execution_count" json:"execution_count"`
	FilesProcessed int64         `yaml:"files_processed" json:"files_processed"`
	BytesReclaimed int64         `yaml:"bytes_reclaimed" json:"bytes_reclaimed"`
}

// RetentionAction represents what action to take on logs
type RetentionAction string

const (
	ActionDelete   RetentionAction = "delete"
	ActionCompress RetentionAction = "compress"
	ActionArchive  RetentionAction = "archive"
	ActionNothing  RetentionAction = "nothing"
)

// RetentionResult represents the result of a retention operation
type RetentionResult struct {
	Action         RetentionAction `json:"action"`
	FilePath       string          `json:"file_path"`
	OriginalSize   int64           `json:"original_size"`
	FinalSize      int64           `json:"final_size"`
	ProcessingTime time.Duration   `json:"processing_time"`
	Success        bool            `json:"success"`
	Error          error           `json:"error,omitempty"`
	Timestamp      time.Time       `json:"timestamp"`
}

// RetentionMetrics represents metrics for retention operations
type RetentionMetrics struct {
	TotalFilesProcessed  int64                     `json:"total_files_processed"`
	TotalBytesReclaimed  int64                     `json:"total_bytes_reclaimed"`
	TotalFilesDeleted    int64                     `json:"total_files_deleted"`
	TotalFilesCompressed int64                     `json:"total_files_compressed"`
	TotalFilesArchived   int64                     `json:"total_files_archived"`
	LastCleanupTime      time.Time                 `json:"last_cleanup_time"`
	CleanupDuration      time.Duration             `json:"cleanup_duration"`
	ErrorCount           int64                     `json:"error_count"`
	LastError            string                    `json:"last_error,omitempty"`
	PolicyExecutions     map[string]int64          `json:"policy_executions"`
	ActionCounts         map[RetentionAction]int64 `json:"action_counts"`
}

// FileInfo represents information about a log file
type FileInfo struct {
	Path         string    `json:"path"`
	Size         int64     `json:"size"`
	ModTime      time.Time `json:"mod_time"`
	IsCompressed bool      `json:"is_compressed"`
	LogLevel     string    `json:"log_level,omitempty"`
	LogSource    string    `json:"log_source,omitempty"`
}

// NewLogRetentionManager creates a new log retention manager
func NewLogRetentionManager(config *LogRetentionConfig, logger *zap.Logger) *LogRetentionManager {
	ctx, cancel := context.WithCancel(context.Background())

	// Set defaults
	if config.CheckInterval == 0 {
		config.CheckInterval = time.Hour * 6 // Check every 6 hours
	}
	if config.MaxConcurrentCleanup == 0 {
		config.MaxConcurrentCleanup = 5
	}
	if config.CompressionDelay == 0 {
		config.CompressionDelay = time.Hour * 24 // Compress after 1 day
	}

	rm := &LogRetentionManager{
		config:   config,
		logger:   logger,
		policies: make(map[string]*RetentionPolicy),
		metrics: &RetentionMetrics{
			PolicyExecutions: make(map[string]int64),
			ActionCounts:     make(map[RetentionAction]int64),
		},
		ctx:    ctx,
		cancel: cancel,
	}

	// Load policies
	rm.loadPolicies()

	return rm
}

// Start starts the retention manager
func (rm *LogRetentionManager) Start() {
	if !rm.config.Enabled {
		rm.logger.Info("Log retention manager disabled")
		return
	}

	rm.logger.Info("Starting log retention manager",
		zap.Duration("check_interval", rm.config.CheckInterval),
		zap.Int("policies", len(rm.policies)))

	rm.wg.Add(1)
	go rm.retentionLoop()
}

// Stop stops the retention manager
func (rm *LogRetentionManager) Stop() {
	rm.logger.Info("Stopping log retention manager")
	rm.cancel()
	rm.wg.Wait()
}

// loadPolicies loads retention policies from configuration
func (rm *LogRetentionManager) loadPolicies() {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	// Load default policy
	if rm.config.DefaultPolicy != nil {
		rm.policies["default"] = rm.config.DefaultPolicy
	}

	// Load named policies
	for name, policy := range rm.config.Policies {
		rm.policies[name] = policy
	}

	rm.logger.Info("Loaded retention policies", zap.Int("count", len(rm.policies)))
}

// retentionLoop runs the periodic retention cleanup
func (rm *LogRetentionManager) retentionLoop() {
	defer rm.wg.Done()

	ticker := time.NewTicker(rm.config.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-rm.ctx.Done():
			return
		case <-ticker.C:
			rm.executeRetentionPolicies()
		}
	}
}

// executeRetentionPolicies executes all configured retention policies
func (rm *LogRetentionManager) executeRetentionPolicies() {
	start := time.Now()
	rm.logger.Info("Starting retention policy execution")

	rm.mutex.RLock()
	policies := make(map[string]*RetentionPolicy)
	for k, v := range rm.policies {
		policies[k] = v
	}
	rm.mutex.RUnlock()

	totalResults := []*RetentionResult{}

	for name, policy := range policies {
		results := rm.executePolicyForDirectories(name, policy)
		totalResults = append(totalResults, results...)

		rm.metrics.PolicyExecutions[name]++
	}

	// Update metrics
	rm.updateMetrics(totalResults, time.Since(start))

	rm.logger.Info("Completed retention policy execution",
		zap.Duration("duration", time.Since(start)),
		zap.Int("files_processed", len(totalResults)))
}

// executePolicyForDirectories executes a policy for configured log directories
func (rm *LogRetentionManager) executePolicyForDirectories(policyName string, policy *RetentionPolicy) []*RetentionResult {
	var results []*RetentionResult

	// Find log directories to process
	logDirs := rm.getLogDirectories()

	for _, dir := range logDirs {
		files, err := rm.getLogFiles(dir)
		if err != nil {
			rm.logger.Error("Failed to get log files",
				zap.String("policy", policyName),
				zap.String("directory", dir),
				zap.Error(err))
			continue
		}

		policyResults := rm.applyPolicyToFiles(policyName, policy, files)
		results = append(results, policyResults...)
	}

	return results
}

// getLogDirectories returns directories that contain log files
func (rm *LogRetentionManager) getLogDirectories() []string {
	var dirs []string

	// Default log directories
	commonDirs := []string{
		"/var/log",
		"./logs",
		"./log",
		"/tmp/gzh-manager-logs",
	}

	for _, dir := range commonDirs {
		if _, err := os.Stat(dir); err == nil {
			dirs = append(dirs, dir)
		}
	}

	// Add archive location if configured
	if rm.config.ArchiveLocation != "" {
		if _, err := os.Stat(rm.config.ArchiveLocation); err == nil {
			dirs = append(dirs, rm.config.ArchiveLocation)
		}
	}

	return dirs
}

// getLogFiles returns log files in a directory
func (rm *LogRetentionManager) getLogFiles(dir string) ([]*FileInfo, error) {
	var files []*FileInfo

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files with errors
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if it's a log file
		if rm.isLogFile(path) {
			fileInfo := &FileInfo{
				Path:         path,
				Size:         info.Size(),
				ModTime:      info.ModTime(),
				IsCompressed: rm.isCompressedFile(path),
			}

			// Extract log level and source from filename if possible
			fileInfo.LogLevel, fileInfo.LogSource = rm.extractLogMetadata(path)

			files = append(files, fileInfo)
		}

		return nil
	})

	return files, err
}

// isLogFile checks if a file is a log file based on extension and naming patterns
func (rm *LogRetentionManager) isLogFile(path string) bool {
	filename := filepath.Base(path)
	ext := filepath.Ext(filename)

	// Common log file extensions
	logExtensions := []string{".log", ".out", ".err", ".trace"}
	for _, logExt := range logExtensions {
		if ext == logExt {
			return true
		}
	}

	// Check for compressed log files
	compressedExtensions := []string{".gz", ".bz2", ".xz", ".zip"}
	for _, compExt := range compressedExtensions {
		if ext == compExt {
			// Check if the file without compression extension is a log file
			nameWithoutComp := strings.TrimSuffix(filename, compExt)
			return rm.isLogFile(nameWithoutComp)
		}
	}

	// Check for log file naming patterns (but be more specific)
	lowerFilename := strings.ToLower(filename)

	// Must contain log-specific patterns AND have appropriate extensions
	logPatterns := []string{"access", "error", "audit"}
	for _, pattern := range logPatterns {
		if strings.Contains(lowerFilename, pattern) {
			return true
		}
	}

	// Special handling for debug and trace - only if they have log extensions
	debugTracePatterns := []string{"debug", "trace"}
	for _, pattern := range debugTracePatterns {
		if strings.Contains(lowerFilename, pattern) && (ext == ".log" || ext == ".out" || ext == ".err") {
			return true
		}
	}

	// Check if filename contains "log" but exclude plain text files
	if strings.Contains(lowerFilename, "log") && ext != ".txt" {
		return true
	}

	return false
}

// isCompressedFile checks if a file is compressed
func (rm *LogRetentionManager) isCompressedFile(path string) bool {
	ext := filepath.Ext(path)
	compressedExtensions := []string{".gz", ".bz2", ".xz", ".zip"}
	for _, compExt := range compressedExtensions {
		if ext == compExt {
			return true
		}
	}
	return false
}

// extractLogMetadata extracts log level and source from filename
func (rm *LogRetentionManager) extractLogMetadata(path string) (level string, source string) {
	filename := filepath.Base(path)

	// Extract log level
	levels := []string{"debug", "info", "warn", "warning", "error", "fatal", "trace"}
	for _, lvl := range levels {
		if strings.Contains(strings.ToLower(filename), lvl) {
			level = lvl
			break
		}
	}

	// Extract source/service name (simple heuristic)
	parts := strings.Split(filename, ".")
	if len(parts) > 0 {
		source = parts[0]
	}

	return level, source
}

// applyPolicyToFiles applies retention policy to a list of files
func (rm *LogRetentionManager) applyPolicyToFiles(policyName string, policy *RetentionPolicy, files []*FileInfo) []*RetentionResult {
	var results []*RetentionResult

	// Sort files by modification time (oldest first)
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime.Before(files[j].ModTime)
	})

	for _, file := range files {
		action := rm.determineAction(policy, file)
		if action == ActionNothing {
			continue
		}

		result := rm.executeAction(action, file)
		results = append(results, result)

		// Check if we should stop processing (e.g., reached limits)
		if rm.shouldStopProcessing(policy, results) {
			break
		}
	}

	return results
}

// determineAction determines what action to take for a file based on policy
func (rm *LogRetentionManager) determineAction(policy *RetentionPolicy, file *FileInfo) RetentionAction {
	now := time.Now()

	// Parse max age
	maxAge, err := rm.parseDuration(policy.MaxAge)
	if err != nil {
		rm.logger.Warn("Invalid max age in policy", zap.String("max_age", policy.MaxAge), zap.Error(err))
		return ActionNothing
	}

	fileAge := now.Sub(file.ModTime)

	// Check if file should be deleted (highest priority)
	if maxAge > 0 && fileAge > maxAge {
		return ActionDelete
	}

	// Check if file should be archived (before compression)
	archiveAge := maxAge / 2 // Archive at half the max age
	if rm.config.ArchiveLocation != "" && maxAge > 0 && fileAge > archiveAge {
		return ActionArchive
	}

	// Check if file should be compressed (lowest priority)
	if rm.config.CompressionEnabled && !file.IsCompressed {
		if fileAge > rm.config.CompressionDelay {
			return ActionCompress
		}
	}

	return ActionNothing
}

// executeAction executes the determined action on a file
func (rm *LogRetentionManager) executeAction(action RetentionAction, file *FileInfo) *RetentionResult {
	start := time.Now()
	result := &RetentionResult{
		Action:         action,
		FilePath:       file.Path,
		OriginalSize:   file.Size,
		Timestamp:      start,
		ProcessingTime: 0,
		Success:        false,
	}

	if rm.config.DryRun {
		rm.logger.Info("DRY RUN: Would execute action",
			zap.String("action", string(action)),
			zap.String("file", file.Path))
		result.Success = true
		result.ProcessingTime = time.Since(start)
		return result
	}

	switch action {
	case ActionDelete:
		result.Error = rm.deleteFile(file.Path)
		result.FinalSize = 0
	case ActionCompress:
		newSize, err := rm.compressFile(file.Path)
		result.Error = err
		result.FinalSize = newSize
	case ActionArchive:
		result.Error = rm.archiveFile(file.Path)
		result.FinalSize = file.Size // Size unchanged for archive
	}

	result.Success = result.Error == nil
	result.ProcessingTime = time.Since(start)

	if result.Error != nil {
		rm.logger.Error("Failed to execute retention action",
			zap.String("action", string(action)),
			zap.String("file", file.Path),
			zap.Error(result.Error))
	} else {
		rm.logger.Info("Successfully executed retention action",
			zap.String("action", string(action)),
			zap.String("file", file.Path),
			zap.Int64("original_size", result.OriginalSize),
			zap.Int64("final_size", result.FinalSize))
	}

	return result
}

// deleteFile deletes a file with optional backup
func (rm *LogRetentionManager) deleteFile(filePath string) error {
	if rm.config.BackupBeforeDelete && rm.config.BackupLocation != "" {
		if err := rm.backupFile(filePath); err != nil {
			return fmt.Errorf("failed to backup file before deletion: %w", err)
		}
	}

	return os.Remove(filePath)
}

// compressFile compresses a file (simplified implementation)
func (rm *LogRetentionManager) compressFile(filePath string) (int64, error) {
	// This is a simplified implementation
	// In a real scenario, you would use proper compression libraries
	compressedPath := filePath + ".gz"

	// Read original file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to read file: %w", err)
	}

	// Write compressed file (simplified - would use gzip in reality)
	err = os.WriteFile(compressedPath, data, 0o644)
	if err != nil {
		return 0, fmt.Errorf("failed to write compressed file: %w", err)
	}

	// Remove original file
	if err := os.Remove(filePath); err != nil {
		return 0, fmt.Errorf("failed to remove original file: %w", err)
	}

	// Get compressed file size
	stat, err := os.Stat(compressedPath)
	if err != nil {
		return 0, fmt.Errorf("failed to get compressed file size: %w", err)
	}

	return stat.Size(), nil
}

// archiveFile moves a file to the archive location
func (rm *LogRetentionManager) archiveFile(filePath string) error {
	if rm.config.ArchiveLocation == "" {
		return fmt.Errorf("archive location not configured")
	}

	// Create archive directory if it doesn't exist
	if err := os.MkdirAll(rm.config.ArchiveLocation, 0o755); err != nil {
		return fmt.Errorf("failed to create archive directory: %w", err)
	}

	// Generate archive path
	filename := filepath.Base(filePath)
	archivePath := filepath.Join(rm.config.ArchiveLocation, filename)

	// Move file to archive
	return os.Rename(filePath, archivePath)
}

// backupFile creates a backup of a file
func (rm *LogRetentionManager) backupFile(filePath string) error {
	if rm.config.BackupLocation == "" {
		return fmt.Errorf("backup location not configured")
	}

	// Create backup directory if it doesn't exist
	if err := os.MkdirAll(rm.config.BackupLocation, 0o755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Generate backup path with timestamp
	filename := filepath.Base(filePath)
	timestamp := time.Now().Format("20060102_150405")
	backupFilename := fmt.Sprintf("%s.%s.backup", filename, timestamp)
	backupPath := filepath.Join(rm.config.BackupLocation, backupFilename)

	// Copy file to backup location
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	return os.WriteFile(backupPath, data, 0o644)
}

// shouldStopProcessing checks if we should stop processing files
func (rm *LogRetentionManager) shouldStopProcessing(policy *RetentionPolicy, results []*RetentionResult) bool {
	// Check if we've processed enough files
	if rm.config.NotificationThreshold > 0 && int64(len(results)) >= rm.config.NotificationThreshold {
		return true
	}

	// Check max docs limit from policy
	if policy.MaxDocs > 0 && int64(len(results)) >= policy.MaxDocs {
		return true
	}

	return false
}

// updateMetrics updates retention metrics
func (rm *LogRetentionManager) updateMetrics(results []*RetentionResult, duration time.Duration) {
	rm.metrics.LastCleanupTime = time.Now()
	rm.metrics.CleanupDuration = duration
	rm.metrics.TotalFilesProcessed += int64(len(results))

	for _, result := range results {
		if result.Success {
			switch result.Action {
			case ActionDelete:
				rm.metrics.TotalFilesDeleted++
				rm.metrics.TotalBytesReclaimed += result.OriginalSize
			case ActionCompress:
				rm.metrics.TotalFilesCompressed++
				rm.metrics.TotalBytesReclaimed += (result.OriginalSize - result.FinalSize)
			case ActionArchive:
				rm.metrics.TotalFilesArchived++
			}
		} else {
			rm.metrics.ErrorCount++
			if result.Error != nil {
				rm.metrics.LastError = result.Error.Error()
			}
		}

		rm.metrics.ActionCounts[result.Action]++
	}
}

// parseDuration parses duration strings like "30d", "1y", "24h"
func (rm *LogRetentionManager) parseDuration(duration string) (time.Duration, error) {
	if duration == "" {
		return 0, nil
	}

	// Handle common suffixes
	if strings.HasSuffix(duration, "d") {
		daysStr := strings.TrimSuffix(duration, "d")
		days, err := time.ParseDuration(daysStr + "h")
		if err != nil {
			return 0, err
		}
		return days * 24, nil // Convert hours to days
	}
	if strings.HasSuffix(duration, "w") {
		weeksStr := strings.TrimSuffix(duration, "w")
		weeks, err := time.ParseDuration(weeksStr + "h")
		if err != nil {
			return 0, err
		}
		return weeks * 24 * 7, nil // Convert to weeks
	}
	if strings.HasSuffix(duration, "y") {
		yearsStr := strings.TrimSuffix(duration, "y")
		years, err := time.ParseDuration(yearsStr + "h")
		if err != nil {
			return 0, err
		}
		return years * 24 * 365, nil // Convert to years
	}

	// Standard Go duration parsing
	return time.ParseDuration(duration)
}

// GetMetrics returns current retention metrics
func (rm *LogRetentionManager) GetMetrics() *RetentionMetrics {
	return rm.metrics
}

// AddPolicy adds a new retention policy
func (rm *LogRetentionManager) AddPolicy(name string, policy *RetentionPolicy) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	rm.policies[name] = policy
	rm.logger.Info("Added retention policy", zap.String("name", name))
}

// RemovePolicy removes a retention policy
func (rm *LogRetentionManager) RemovePolicy(name string) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	delete(rm.policies, name)
	rm.logger.Info("Removed retention policy", zap.String("name", name))
}

// GetPolicies returns all configured policies
func (rm *LogRetentionManager) GetPolicies() map[string]*RetentionPolicy {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	policies := make(map[string]*RetentionPolicy)
	for k, v := range rm.policies {
		policies[k] = v
	}
	return policies
}

// ExecutePolicy manually executes a specific policy
func (rm *LogRetentionManager) ExecutePolicy(policyName string) ([]*RetentionResult, error) {
	rm.mutex.RLock()
	policy, exists := rm.policies[policyName]
	rm.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("policy %s not found", policyName)
	}

	rm.logger.Info("Manually executing retention policy", zap.String("policy", policyName))
	results := rm.executePolicyForDirectories(policyName, policy)

	rm.metrics.PolicyExecutions[policyName]++
	rm.updateMetrics(results, 0)

	return results, nil
}
