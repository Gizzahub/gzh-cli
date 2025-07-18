package github

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewChangeLogger(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "change_logger_test")
	require.NoError(t, err)

	defer os.RemoveAll(tempDir)

	store, err := NewFileStore(tempDir)
	require.NoError(t, err)

	changelog := NewChangeLog(nil, store)

	options := &LoggerOptions{
		LogDirectory:        tempDir,
		LogFormat:           LogFormatJSON,
		LogLevel:            LogLevelInfo,
		EnableConsoleOutput: false,
	}

	logger := NewChangeLogger(changelog, options)
	assert.NotNil(t, logger)
	assert.Equal(t, changelog, logger.changelog)
	assert.Equal(t, options, logger.options)
}

func TestDefaultLoggerOptions(t *testing.T) {
	options := DefaultLoggerOptions()
	assert.NotNil(t, options)
	assert.NotEmpty(t, options.LogDirectory)
	assert.Equal(t, LogFormatJSON, options.LogFormat)
	assert.Equal(t, LogLevelInfo, options.LogLevel)
	assert.True(t, options.EnableConsoleOutput)
	assert.Equal(t, int64(10*1024*1024), options.MaxLogFileSize)
	assert.Equal(t, 5, options.MaxLogFiles)
}

func TestLogRepositoryChange(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "change_logger_test")
	require.NoError(t, err)

	defer os.RemoveAll(tempDir)

	store, err := NewFileStore(tempDir)
	require.NoError(t, err)

	changelog := NewChangeLog(nil, store)

	options := &LoggerOptions{
		LogDirectory:        tempDir,
		LogFormat:           LogFormatJSON,
		LogLevel:            LogLevelInfo,
		EnableConsoleOutput: false,
	}

	logger := NewChangeLogger(changelog, options)
	ctx := context.Background()

	opCtx := &operationContext{
		RequestID:    "test-123",
		User:         "testuser",
		Source:       "cli",
		Organization: "testorg",
		StartTime:    time.Now().Add(-5 * time.Second),
		Metadata:     map[string]interface{}{"test": true},
	}

	changeRecord := &ChangeRecord{
		ID:           "change-123",
		Timestamp:    time.Now(),
		User:         "testuser",
		Organization: "testorg",
		Repository:   "testorg/testrepo",
		Operation:    "update",
		Category:     "settings",
		Before:       map[string]interface{}{"private": false},
		After:        map[string]interface{}{"private": true},
		Description:  "Made repository private",
		Source:       "cli",
	}

	err = logger.LogRepositoryChange(ctx, opCtx, changeRecord, LogLevelInfo, "Repository updated successfully", nil)
	require.NoError(t, err)

	// Verify log file was created
	logFiles, err := filepath.Glob(filepath.Join(tempDir, "gzh-manager-*.log"))
	require.NoError(t, err)
	assert.Len(t, logFiles, 1)

	// Read and verify log content
	content, err := os.ReadFile(logFiles[0])
	require.NoError(t, err)

	var entry logEntry

	err = json.Unmarshal(content, &entry)
	require.NoError(t, err)

	assert.Equal(t, LogLevelInfo, entry.Level)
	assert.Equal(t, "Repository updated successfully", entry.Message)
	assert.Equal(t, "change-123", entry.ChangeID)
	assert.Equal(t, "update", entry.Operation)
	assert.Equal(t, "settings", entry.Category)
	assert.Equal(t, "testorg/testrepo", entry.Repository)
	assert.Equal(t, "testuser", entry.User)
	assert.Equal(t, "test-123", entry.RequestID)
	assert.NotNil(t, entry.Duration)
	assert.Greater(t, *entry.Duration, time.Duration(0))
}

func TestLogOperation(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "change_logger_test")
	require.NoError(t, err)

	defer os.RemoveAll(tempDir)

	store, err := NewFileStore(tempDir)
	require.NoError(t, err)

	changelog := NewChangeLog(nil, store)

	options := &LoggerOptions{
		LogDirectory:        tempDir,
		LogFormat:           LogFormatText,
		LogLevel:            LogLevelDebug,
		EnableConsoleOutput: false,
	}

	logger := NewChangeLogger(changelog, options)
	ctx := context.Background()

	opCtx := &operationContext{
		RequestID:    "op-456",
		User:         "testuser",
		Source:       "api",
		Organization: "testorg",
		Repository:   "testorg/repo",
		StartTime:    time.Now().Add(-2 * time.Second),
	}

	err = logger.LogOperation(ctx, opCtx, LogLevelWarn, "validate", "configuration", "Validation completed with warnings", nil)
	require.NoError(t, err)

	// Verify log file content
	logFiles, err := filepath.Glob(filepath.Join(tempDir, "gzh-manager-*.log"))
	require.NoError(t, err)
	assert.Len(t, logFiles, 1)

	content, err := os.ReadFile(logFiles[0])
	require.NoError(t, err)

	logLine := string(content)
	assert.Contains(t, logLine, "[WARN]")
	assert.Contains(t, logLine, "Validation completed with warnings")
	assert.Contains(t, logLine, "user=testuser")
	assert.Contains(t, logLine, "op=validate")
	assert.Contains(t, logLine, "duration=")
}

func TestLogBulkOperation(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "change_logger_test")
	require.NoError(t, err)

	defer os.RemoveAll(tempDir)

	store, err := NewFileStore(tempDir)
	require.NoError(t, err)

	changelog := NewChangeLog(nil, store)

	options := &LoggerOptions{
		LogDirectory:        tempDir,
		LogFormat:           LogFormatJSON,
		LogLevel:            LogLevelInfo,
		EnableConsoleOutput: false,
	}

	logger := NewChangeLogger(changelog, options)
	ctx := context.Background()

	opCtx := &operationContext{
		RequestID:    "bulk-789",
		User:         "testuser",
		Source:       "cli",
		Organization: "testorg",
		StartTime:    time.Now().Add(-30 * time.Second),
	}

	stats := &bulkOperationStats{
		Total:    10,
		Success:  8,
		Failed:   1,
		Skipped:  1,
		Duration: 30 * time.Second,
		Errors:   map[string]string{"repo1": "permission denied"},
	}

	err = logger.LogBulkOperation(ctx, opCtx, LogLevelInfo, "update", stats, nil)
	require.NoError(t, err)

	// Verify log content
	logFiles, err := filepath.Glob(filepath.Join(tempDir, "gzh-manager-*.log"))
	require.NoError(t, err)
	assert.Len(t, logFiles, 1)

	content, err := os.ReadFile(logFiles[0])
	require.NoError(t, err)

	var entry logEntry

	err = json.Unmarshal(content, &entry)
	require.NoError(t, err)

	assert.Equal(t, "bulk_update", entry.Operation)
	assert.Equal(t, "bulk_operation", entry.Category)
	assert.Contains(t, entry.Message, "10 total, 8 success, 1 failed, 1 skipped")
	assert.Contains(t, entry.Context, "statistics")
}

func TestCreateOperationContext(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "change_logger_test")
	require.NoError(t, err)

	defer os.RemoveAll(tempDir)

	store, err := NewFileStore(tempDir)
	require.NoError(t, err)

	changelog := NewChangeLog(nil, store)

	logger := NewChangeLogger(changelog, nil)

	opCtx := logger.CreateOperationContext("req-123", "test_operation")

	assert.Equal(t, "req-123", opCtx.RequestID)
	assert.Equal(t, "cli", opCtx.Source)
	assert.NotEmpty(t, opCtx.User)
	assert.False(t, opCtx.StartTime.IsZero())
	assert.NotNil(t, opCtx.Metadata)
}

func TestShouldLog(t *testing.T) {
	logger := &ChangeLogger{
		options: &LoggerOptions{
			LogLevel: LogLevelWarn,
		},
	}

	tests := []struct {
		level    LogLevel
		expected bool
	}{
		{LogLevelTrace, false},
		{LogLevelDebug, false},
		{LogLevelInfo, false},
		{LogLevelWarn, true},
		{LogLevelError, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.level), func(t *testing.T) {
			result := logger.shouldLog(tt.level)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatTextEntry(t *testing.T) {
	logger := &ChangeLogger{}
	duration := 5 * time.Second

	entry := &logEntry{
		Timestamp:  time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
		Level:      LogLevelInfo,
		Message:    "Test message",
		Repository: "test/repo",
		User:       "testuser",
		Operation:  "update",
		Duration:   &duration,
		Error:      "test error",
	}

	result := logger.formatTextEntry(entry)

	assert.Contains(t, result, "[INFO]")
	assert.Contains(t, result, "2023-01-01 12:00:00")
	assert.Contains(t, result, "Test message")
	assert.Contains(t, result, "repo=test/repo")
	assert.Contains(t, result, "user=testuser")
	assert.Contains(t, result, "op=update")
	assert.Contains(t, result, "duration=5s")
	assert.Contains(t, result, "error=test error")
}

func TestFormatCSVEntry(t *testing.T) {
	logger := &ChangeLogger{}

	entry := &logEntry{
		Timestamp:  time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
		Level:      LogLevelError,
		Operation:  "delete",
		Category:   "permissions",
		Repository: "org/repo",
		User:       "admin",
		Message:    "Permission removed",
		Error:      "access denied",
	}

	result := logger.formatCSVEntry(entry)

	expected := "2023-01-01 12:00:00,error,delete,permissions,org/repo,admin,Permission removed,access denied"
	assert.Equal(t, expected, result)
}

func TestLogRotation(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "change_logger_test")
	require.NoError(t, err)

	defer os.RemoveAll(tempDir)

	store, err := NewFileStore(tempDir)
	require.NoError(t, err)

	changelog := NewChangeLog(nil, store)

	options := &LoggerOptions{
		LogDirectory:        tempDir,
		LogFormat:           LogFormatText,
		LogLevel:            LogLevelInfo,
		MaxLogFileSize:      100, // Very small size to trigger rotation
		MaxLogFiles:         2,
		EnableConsoleOutput: false,
	}

	logger := NewChangeLogger(changelog, options)
	logFile := logger.getLogFileName()

	// Create a large log file that needs rotation
	largeContent := strings.Repeat("This is a test log line\n", 10)
	err = os.WriteFile(logFile, []byte(largeContent), 0o644)
	require.NoError(t, err)

	ctx := context.Background()
	opCtx := &operationContext{
		RequestID: "rotation-test",
		User:      "testuser",
		Source:    "cli",
	}

	// This should trigger rotation
	err = logger.LogOperation(ctx, opCtx, LogLevelInfo, "test", "rotation", "Test rotation", nil)
	require.NoError(t, err)

	// Check that rotated file exists
	rotatedFiles, err := filepath.Glob(filepath.Join(tempDir, "gzh-manager-*.log.*"))
	require.NoError(t, err)
	assert.Greater(t, len(rotatedFiles), 0)
}

func TestGetLogSummary(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "change_logger_test")
	require.NoError(t, err)

	defer os.RemoveAll(tempDir)

	store, err := NewFileStore(tempDir)
	require.NoError(t, err)

	changelog := NewChangeLog(nil, store)

	logger := NewChangeLogger(changelog, nil)
	ctx := context.Background()

	// Add some test changes
	changes := []*ChangeRecord{
		{
			ID:          "1",
			Timestamp:   time.Now(),
			User:        "user1",
			Operation:   "create",
			Category:    "settings",
			Description: "Test change 1",
		},
		{
			ID:          "2",
			Timestamp:   time.Now(),
			User:        "user2",
			Operation:   "update",
			Category:    "permissions",
			Description: "Test change 2",
		},
		{
			ID:          "3",
			Timestamp:   time.Now(),
			User:        "user1",
			Operation:   "delete",
			Category:    "settings",
			Description: "Test change 3",
		},
	}

	for _, change := range changes {
		err = changelog.RecordChange(ctx, change)
		require.NoError(t, err)
	}

	// Get summary
	since := time.Now().Add(-1 * time.Hour)
	summary, err := logger.GetLogSummary(ctx, since)
	require.NoError(t, err)

	assert.Equal(t, 3, summary.TotalChanges)
	assert.Equal(t, 2, summary.ByCategory["settings"])
	assert.Equal(t, 1, summary.ByCategory["permissions"])
	assert.Equal(t, 1, summary.ByOperation["create"])
	assert.Equal(t, 1, summary.ByOperation["update"])
	assert.Equal(t, 1, summary.ByOperation["delete"])
	assert.Equal(t, 2, summary.ByUser["user1"])
	assert.Equal(t, 1, summary.ByUser["user2"])
}

func TestLogFormats(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "change_logger_test")
	require.NoError(t, err)

	defer os.RemoveAll(tempDir)

	store, err := NewFileStore(tempDir)
	require.NoError(t, err)

	changelog := NewChangeLog(nil, store)

	formats := []LogFormat{LogFormatJSON, LogFormatText, LogFormatCSV}

	for _, format := range formats {
		t.Run(string(format), func(t *testing.T) {
			formatDir := filepath.Join(tempDir, string(format))
			err := os.MkdirAll(formatDir, 0o755)
			require.NoError(t, err)

			options := &LoggerOptions{
				LogDirectory:        formatDir,
				LogFormat:           format,
				LogLevel:            LogLevelInfo,
				EnableConsoleOutput: false,
			}

			logger := NewChangeLogger(changelog, options)
			ctx := context.Background()

			opCtx := &operationContext{
				RequestID: "format-test",
				User:      "testuser",
				Source:    "cli",
			}

			err = logger.LogOperation(ctx, opCtx, LogLevelInfo, "test", "format", "Testing format", nil)
			require.NoError(t, err)

			// Verify file was created
			logFiles, err := filepath.Glob(filepath.Join(formatDir, "gzh-manager-*.log"))
			require.NoError(t, err)
			assert.Len(t, logFiles, 1)

			// Verify content format
			content, err := os.ReadFile(logFiles[0])
			require.NoError(t, err)

			switch format {
			case LogFormatJSON:
				var entry logEntry

				err = json.Unmarshal(content, &entry)
				assert.NoError(t, err)
			case LogFormatText:
				assert.Contains(t, string(content), "[INFO]")
				assert.Contains(t, string(content), "Testing format")
			case LogFormatCSV:
				assert.Contains(t, string(content), "info,test,format")
			}
		})
	}
}
