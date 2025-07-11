package monitoring

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestLogRetentionManager(t *testing.T) {
	t.Run("New LogRetentionManager", func(t *testing.T) {
		logger, _ := zap.NewDevelopment()
		config := &LogRetentionConfig{
			Enabled:              true,
			CheckInterval:        time.Minute,
			CompressionEnabled:   true,
			CompressionDelay:     time.Hour,
			DryRun:               true,
			MaxConcurrentCleanup: 3,
		}

		rm := NewLogRetentionManager(config, logger)
		require.NotNil(t, rm)
		assert.Equal(t, config, rm.config)
		assert.NotNil(t, rm.metrics)
		assert.NotNil(t, rm.policies)

		// Start and stop to test lifecycle
		rm.Start()
		time.Sleep(100 * time.Millisecond) // Let it start
		rm.Stop()
	})

	t.Run("Policy Management", func(t *testing.T) {
		logger, _ := zap.NewDevelopment()
		config := &LogRetentionConfig{
			Enabled: true,
			DefaultPolicy: &RetentionPolicy{
				MaxAge:  "30d",
				MaxSize: "1GB",
				MaxDocs: 10000,
			},
			Policies: map[string]*RetentionPolicy{
				"debug": {
					MaxAge:  "7d",
					MaxSize: "100MB",
					MaxDocs: 1000,
				},
			},
		}

		rm := NewLogRetentionManager(config, logger)

		// Check initial policies
		policies := rm.GetPolicies()
		assert.Len(t, policies, 2) // default + debug
		assert.NotNil(t, policies["default"])
		assert.NotNil(t, policies["debug"])

		// Add new policy
		newPolicy := &RetentionPolicy{
			MaxAge:  "1d",
			MaxSize: "10MB",
			MaxDocs: 100,
		}
		rm.AddPolicy("error", newPolicy)

		policies = rm.GetPolicies()
		assert.Len(t, policies, 3)
		assert.Equal(t, newPolicy, policies["error"])

		// Remove policy
		rm.RemovePolicy("debug")
		policies = rm.GetPolicies()
		assert.Len(t, policies, 2)
		assert.Nil(t, policies["debug"])
	})

	t.Run("File Detection", func(t *testing.T) {
		logger, _ := zap.NewDevelopment()
		config := &LogRetentionConfig{
			Enabled: true,
			DryRun:  true,
		}

		rm := NewLogRetentionManager(config, logger)

		// Test log file detection
		testCases := []struct {
			filename string
			isLog    bool
		}{
			{"application.log", true},
			{"access.log", true},
			{"error.log", true},
			{"debug.txt", false}, // Changed: .txt files need specific patterns
			{"trace.out", true},
			{"app.log.gz", true},
			{"test.log.bz2", true},
			{"regular.txt", false},
			{"config.json", false},
			{"readme.md", false},
		}

		for _, tc := range testCases {
			result := rm.isLogFile(tc.filename)
			assert.Equal(t, tc.isLog, result, "Failed for file: %s", tc.filename)
		}

		// Test compression detection
		assert.True(t, rm.isCompressedFile("test.log.gz"))
		assert.True(t, rm.isCompressedFile("app.log.bz2"))
		assert.False(t, rm.isCompressedFile("app.log"))
	})

	t.Run("Duration Parsing", func(t *testing.T) {
		logger, _ := zap.NewDevelopment()
		config := &LogRetentionConfig{Enabled: true}
		rm := NewLogRetentionManager(config, logger)

		testCases := []struct {
			input    string
			expected time.Duration
			hasError bool
		}{
			{"1h", time.Hour, false},
			{"24h", 24 * time.Hour, false},
			{"1d", 24 * time.Hour, false},
			{"7d", 7 * 24 * time.Hour, false},
			{"1w", 7 * 24 * time.Hour, false},
			{"1y", 365 * 24 * time.Hour, false},
			{"", 0, false},
			{"invalid", 0, true},
		}

		for _, tc := range testCases {
			result, err := rm.parseDuration(tc.input)
			if tc.hasError {
				assert.Error(t, err, "Expected error for input: %s", tc.input)
			} else {
				assert.NoError(t, err, "Unexpected error for input: %s", tc.input)
				assert.Equal(t, tc.expected, result, "Duration mismatch for input: %s", tc.input)
			}
		}
	})

	t.Run("Action Determination", func(t *testing.T) {
		logger, _ := zap.NewDevelopment()
		config := &LogRetentionConfig{
			Enabled:            true,
			CompressionEnabled: true,
			CompressionDelay:   time.Hour * 24,
			ArchiveLocation:    "/tmp/archive",
		}
		rm := NewLogRetentionManager(config, logger)

		policy := &RetentionPolicy{
			MaxAge:  "7d", // 7 days
			MaxSize: "1GB",
			MaxDocs: 1000,
		}

		now := time.Now()

		// File that should be deleted (older than 7 days)
		oldFile := &FileInfo{
			Path:         "/tmp/old.log",
			Size:         1024,
			ModTime:      now.Add(-8 * 24 * time.Hour), // 8 days old
			IsCompressed: false,
		}
		action := rm.determineAction(policy, oldFile)
		assert.Equal(t, ActionDelete, action)

		// File that should be archived (older than archive age but younger than max age)
		archiveFile := &FileInfo{
			Path:         "/tmp/archive.log",
			Size:         1024,
			ModTime:      now.Add(-4 * 24 * time.Hour), // 4 days old (past archive age of 3.5 days)
			IsCompressed: false,
		}
		action = rm.determineAction(policy, archiveFile)
		assert.Equal(t, ActionArchive, action)

		// Recent file that needs no action
		recentFile := &FileInfo{
			Path:         "/tmp/recent.log",
			Size:         1024,
			ModTime:      now.Add(-1 * time.Hour), // 1 hour old
			IsCompressed: false,
		}
		action = rm.determineAction(policy, recentFile)
		assert.Equal(t, ActionNothing, action)

		// Already compressed file that should be archived
		compressedFile := &FileInfo{
			Path:         "/tmp/compressed.log.gz",
			Size:         512,
			ModTime:      now.Add(-4 * 24 * time.Hour), // 4 days old, past archive age
			IsCompressed: true,
		}
		action = rm.determineAction(policy, compressedFile)
		assert.Equal(t, ActionArchive, action) // Should archive since it's past archive age
	})

	t.Run("Metadata Extraction", func(t *testing.T) {
		logger, _ := zap.NewDevelopment()
		config := &LogRetentionConfig{Enabled: true}
		rm := NewLogRetentionManager(config, logger)

		testCases := []struct {
			filename       string
			expectedLevel  string
			expectedSource string
		}{
			{"application.error.log", "error", "application"},
			{"payment-service.debug.log", "debug", "payment-service"},
			{"access.log", "", "access"},
			{"system.info.2023-01-01.log", "info", "system"},
			{"trace.out", "trace", "trace"},
		}

		for _, tc := range testCases {
			level, source := rm.extractLogMetadata(tc.filename)
			assert.Equal(t, tc.expectedLevel, level, "Level mismatch for: %s", tc.filename)
			assert.Equal(t, tc.expectedSource, source, "Source mismatch for: %s", tc.filename)
		}
	})
}

func TestLogRetentionIntegration(t *testing.T) {
	t.Run("Integration with CentralizedLogger", func(t *testing.T) {
		// Create temporary directory for test logs
		tempDir, err := os.MkdirTemp("", "retention_test")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &CentralizedLoggingConfig{
			Level:         "info",
			Format:        "json",
			Directory:     tempDir,
			BaseFilename:  "test.log",
			BufferSize:    1000,
			FlushInterval: time.Second,
			Retention: &LogRetentionConfig{
				Enabled:            true,
				CheckInterval:      time.Minute * 5,
				CompressionEnabled: true,
				CompressionDelay:   time.Hour,
				DryRun:             true, // Don't actually delete files in test
				DefaultPolicy: &RetentionPolicy{
					MaxAge:  "7d",
					MaxSize: "100MB",
					MaxDocs: 1000,
				},
			},
		}

		registry := prometheus.NewRegistry()
		centralLogger, err := NewCentralizedLogger(config, registry)
		require.NoError(t, err)
		defer centralLogger.Shutdown(context.Background())

		// Verify retention manager is initialized
		retentionManager := centralLogger.GetRetentionManager()
		require.NotNil(t, retentionManager)

		// Test retention manager functionality
		policies := retentionManager.GetPolicies()
		assert.Len(t, policies, 1) // default policy
		assert.NotNil(t, policies["default"])

		// Get metrics
		metrics := retentionManager.GetMetrics()
		assert.NotNil(t, metrics)
		assert.Equal(t, int64(0), metrics.TotalFilesProcessed) // No files processed yet

		// Test stats include retention info
		stats := centralLogger.GetStats()
		assert.Contains(t, stats, "retention")
		assert.Contains(t, stats, "retention_config")
	})

	t.Run("Retention Metrics Update", func(t *testing.T) {
		logger, _ := zap.NewDevelopment()
		config := &LogRetentionConfig{
			Enabled: true,
			DryRun:  true,
		}

		rm := NewLogRetentionManager(config, logger)

		// Simulate some retention results
		results := []*RetentionResult{
			{
				Action:       ActionDelete,
				FilePath:     "/tmp/file1.log",
				OriginalSize: 1024,
				FinalSize:    0,
				Success:      true,
				Timestamp:    time.Now(),
			},
			{
				Action:       ActionCompress,
				FilePath:     "/tmp/file2.log",
				OriginalSize: 2048,
				FinalSize:    512,
				Success:      true,
				Timestamp:    time.Now(),
			},
			{
				Action:       ActionArchive,
				FilePath:     "/tmp/file3.log",
				OriginalSize: 1024,
				FinalSize:    1024,
				Success:      true,
				Timestamp:    time.Now(),
			},
		}

		rm.updateMetrics(results, time.Second)

		metrics := rm.GetMetrics()
		assert.Equal(t, int64(3), metrics.TotalFilesProcessed)
		assert.Equal(t, int64(1), metrics.TotalFilesDeleted)
		assert.Equal(t, int64(1), metrics.TotalFilesCompressed)
		assert.Equal(t, int64(1), metrics.TotalFilesArchived)
		assert.Equal(t, int64(1024+1536), metrics.TotalBytesReclaimed) // 1024 (deleted) + 1536 (compressed savings)
		assert.Equal(t, time.Second, metrics.CleanupDuration)
		assert.Equal(t, int64(0), metrics.ErrorCount)
	})

	t.Run("File Operations", func(t *testing.T) {
		// Create temporary directory and files for testing
		tempDir, err := os.MkdirTemp("", "retention_ops_test")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		logger, _ := zap.NewDevelopment()
		config := &LogRetentionConfig{
			Enabled:         true,
			DryRun:          false, // Actually perform operations
			ArchiveLocation: filepath.Join(tempDir, "archive"),
			BackupLocation:  filepath.Join(tempDir, "backup"),
		}

		rm := NewLogRetentionManager(config, logger)

		// Create test files
		testFiles := []string{
			"test1.log",
			"test2.log",
			"test3.log",
		}

		for _, filename := range testFiles {
			filePath := filepath.Join(tempDir, filename)
			content := []byte("test log content for " + filename)
			err := os.WriteFile(filePath, content, 0o644)
			require.NoError(t, err)
		}

		// Test file discovery
		files, err := rm.getLogFiles(tempDir)
		require.NoError(t, err)
		assert.Len(t, files, 3)

		// Test compression (simplified)
		testFile := files[0]
		_, err = rm.compressFile(testFile.Path)
		assert.NoError(t, err)

		// Verify compressed file exists
		compressedPath := testFile.Path + ".gz"
		_, err = os.Stat(compressedPath)
		assert.NoError(t, err)

		// Test archiving
		testFile2 := files[1]
		err = rm.archiveFile(testFile2.Path)
		assert.NoError(t, err)

		// Verify file was moved to archive
		archivedPath := filepath.Join(config.ArchiveLocation, filepath.Base(testFile2.Path))
		_, err = os.Stat(archivedPath)
		assert.NoError(t, err)

		// Test deletion with backup
		config.BackupBeforeDelete = true
		testFile3 := files[2]
		err = rm.deleteFile(testFile3.Path)
		assert.NoError(t, err)

		// Verify original file is deleted
		_, err = os.Stat(testFile3.Path)
		assert.True(t, os.IsNotExist(err))

		// Verify backup was created
		backupFiles, _ := filepath.Glob(filepath.Join(config.BackupLocation, "*.backup"))
		assert.Len(t, backupFiles, 1)
	})
}

func TestRetentionPolicyExecution(t *testing.T) {
	t.Run("Manual Policy Execution", func(t *testing.T) {
		logger, _ := zap.NewDevelopment()
		config := &LogRetentionConfig{
			Enabled: true,
			DryRun:  true,
			Policies: map[string]*RetentionPolicy{
				"test-policy": {
					MaxAge:  "1d",
					MaxSize: "10MB",
					MaxDocs: 100,
				},
			},
		}

		rm := NewLogRetentionManager(config, logger)

		// Execute non-existent policy
		_, err := rm.ExecutePolicy("non-existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")

		// Execute existing policy
		results, err := rm.ExecutePolicy("test-policy")
		assert.NoError(t, err)
		assert.NotNil(t, results)

		// Check that policy execution count increased
		metrics := rm.GetMetrics()
		assert.Equal(t, int64(1), metrics.PolicyExecutions["test-policy"])
	})

	t.Run("Policy Configuration Validation", func(t *testing.T) {
		logger, _ := zap.NewDevelopment()

		// Test with minimal config
		config := &LogRetentionConfig{
			Enabled: true,
		}

		rm := NewLogRetentionManager(config, logger)
		assert.NotNil(t, rm)

		// Verify defaults were set
		assert.Equal(t, time.Hour*6, rm.config.CheckInterval)
		assert.Equal(t, 5, rm.config.MaxConcurrentCleanup)
		assert.Equal(t, time.Hour*24, rm.config.CompressionDelay)
	})

	t.Run("Stop Processing Conditions", func(t *testing.T) {
		logger, _ := zap.NewDevelopment()
		config := &LogRetentionConfig{
			Enabled:               true,
			NotificationThreshold: 2,
		}

		rm := NewLogRetentionManager(config, logger)

		policy := &RetentionPolicy{MaxDocs: 3}

		// Test notification threshold
		results := []*RetentionResult{{}, {}}
		assert.True(t, rm.shouldStopProcessing(policy, results))

		// Test max docs limit
		policy.MaxDocs = 1
		results = []*RetentionResult{{}}
		assert.True(t, rm.shouldStopProcessing(policy, results))

		// Test no limits reached
		policy.MaxDocs = 5
		config.NotificationThreshold = 5
		results = []*RetentionResult{{}}
		assert.False(t, rm.shouldStopProcessing(policy, results))
	})
}
