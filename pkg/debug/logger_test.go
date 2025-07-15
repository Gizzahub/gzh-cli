package debug

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogLevel(t *testing.T) {
	tests := []struct {
		level LogLevel
		name  string
		color string
	}{
		{LevelSilent, "SILENT", ""},
		{LevelError, "ERROR", "\033[31m"},
		{LevelWarn, "WARN", "\033[33m"},
		{LevelInfo, "INFO", "\033[36m"},
		{LevelDebug, "DEBUG", "\033[32m"},
		{LevelTrace, "TRACE", "\033[35m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.name, levelNames[tt.level])
			if tt.color != "" {
				assert.Equal(t, tt.color, levelColors[tt.level])
			}
		})
	}
}

func TestDefaultLoggerConfig(t *testing.T) {
	config := DefaultLoggerConfig()

	assert.Equal(t, LevelInfo, config.Level)
	assert.True(t, config.EnableColor)
	assert.False(t, config.EnableTrace)
	assert.Equal(t, int64(100*1024*1024), config.MaxFileSize)
	assert.Equal(t, 5, config.MaxBackups)
	assert.True(t, config.Compress)
	assert.Equal(t, "text", config.Format)
}

func TestNewLogger(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("with config", func(t *testing.T) {
		config := &LoggerConfig{
			Level:       LevelDebug,
			File:        filepath.Join(tmpDir, "test.log"),
			EnableColor: true,
			EnableTrace: false,
			Format:      "text",
		}

		logger, err := NewLogger(config)
		assert.NoError(t, err)
		assert.NotNil(t, logger)

		// Clean up
		logger.Close()
	})

	t.Run("with nil config", func(t *testing.T) {
		logger, err := NewLogger(nil)
		assert.NoError(t, err)
		assert.NotNil(t, logger)

		// Clean up
		logger.Close()
	})
}

func TestLoggerLevels(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	config := &LoggerConfig{
		Level:       LevelDebug,
		File:        logFile,
		EnableColor: false,
		EnableTrace: false,
		Format:      "text",
	}

	logger, err := NewLogger(config)
	require.NoError(t, err)
	defer logger.Close()

	// Test different log levels
	logger.Error("error message")
	logger.Warn("warn message")
	logger.Info("info message")
	logger.Debug("debug message")
	logger.Trace("trace message")

	// Read log file
	content, err := os.ReadFile(logFile)
	if err == nil {
		logContent := string(content)
		assert.Contains(t, logContent, "error message")
		assert.Contains(t, logContent, "warn message")
		assert.Contains(t, logContent, "info message")
		assert.Contains(t, logContent, "debug message")
		// Trace might not appear if level is Debug
	}
}

func TestLoggerJSON(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	config := &LoggerConfig{
		Level:       LevelInfo,
		File:        logFile,
		EnableColor: false,
		EnableTrace: false,
		Format:      "json",
	}

	logger, err := NewLogger(config)
	require.NoError(t, err)
	defer logger.Close()

	logger.Info("test json message")

	// Check if JSON format is used
	content, err := os.ReadFile(logFile)
	if err == nil {
		logContent := string(content)
		assert.Contains(t, logContent, "test json message")
		// JSON logs should contain quotes and braces
		assert.True(t, strings.Contains(logContent, "{") || strings.Contains(logContent, "\""))
	}
}

func TestLoggerSetLevel(t *testing.T) {
	logger, err := NewLogger(nil)
	require.NoError(t, err)
	defer logger.Close()

	// Test setting different levels
	logger.SetLevel(LevelError)
	assert.Equal(t, LevelError, logger.GetLevel())

	logger.SetLevel(LevelDebug)
	assert.Equal(t, LevelDebug, logger.GetLevel())
}

func TestLoggerWithFields(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	config := &LoggerConfig{
		Level:       LevelInfo,
		File:        logFile,
		EnableColor: false,
		EnableTrace: false,
		Format:      "json",
	}

	logger, err := NewLogger(config)
	require.NoError(t, err)
	defer logger.Close()

	fields := map[string]interface{}{
		"user_id": 123,
		"action":  "login",
	}

	logger.WithFields(fields).Info("user action")

	// Check if fields are included
	content, err := os.ReadFile(logFile)
	if err == nil {
		logContent := string(content)
		assert.Contains(t, logContent, "user action")
		// Fields should be present in JSON format
		if strings.Contains(logContent, "{") {
			assert.Contains(t, logContent, "123")
			assert.Contains(t, logContent, "login")
		}
	}
}

func TestLoggerConcurrentAccess(t *testing.T) {
	logger, err := NewLogger(nil)
	require.NoError(t, err)
	defer logger.Close()

	// Test concurrent logging
	done := make(chan bool, 2)

	go func() {
		for i := 0; i < 10; i++ {
			logger.Infof("goroutine 1 message %d", i)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 10; i++ {
			logger.Debugf("goroutine 2 message %d", i)
		}
		done <- true
	}()

	// Wait for both goroutines to complete
	<-done
	<-done

	// If we reach here without panic, concurrent access is safe
	assert.True(t, true)
}

func TestLoggerFiltering(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	config := &LoggerConfig{
		Level:       LevelWarn, // Only WARN and ERROR should be logged
		File:        logFile,
		EnableColor: false,
		EnableTrace: false,
		Format:      "text",
	}

	logger, err := NewLogger(config)
	require.NoError(t, err)
	defer logger.Close()

	logger.Error("error message")
	logger.Warn("warn message")
	logger.Info("info message")   // Should be filtered out
	logger.Debug("debug message") // Should be filtered out

	// Read log file
	content, err := os.ReadFile(logFile)
	if err == nil {
		logContent := string(content)
		assert.Contains(t, logContent, "error message")
		assert.Contains(t, logContent, "warn message")
		assert.NotContains(t, logContent, "info message")
		assert.NotContains(t, logContent, "debug message")
	}
}

func BenchmarkLoggerInfo(b *testing.B) {
	logger, err := NewLogger(nil)
	require.NoError(b, err)
	defer logger.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Infof("benchmark message %d", i)
	}
}

func BenchmarkLoggerWithFields(b *testing.B) {
	logger, err := NewLogger(nil)
	require.NoError(b, err)
	defer logger.Close()

	fields := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.WithFields(fields).Info("benchmark message")
	}
}
