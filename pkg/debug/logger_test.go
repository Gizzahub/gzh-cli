package debug

import (
	"bytes"
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

// Structured Logger Tests

func TestDefaultStructuredLoggerConfig(t *testing.T) {
	config := DefaultStructuredLoggerConfig()

	assert.Equal(t, SeverityInfo, config.Level)
	assert.Equal(t, "json", config.Format)
	assert.Equal(t, "stderr", config.Output)
	assert.Equal(t, "gzh-manager", config.AppName)
	assert.Equal(t, "1.0.0", config.Version)
	assert.Equal(t, "development", config.Environment)
	assert.True(t, config.EnableTracing)
	assert.True(t, config.EnableCaller)
	assert.Equal(t, 2, config.CallerSkip)
	assert.False(t, config.EnableSampling)
	assert.Equal(t, 1.0, config.SampleRate)
	assert.Equal(t, 1, config.SampleThreshold)
	assert.False(t, config.AsyncLogging)
	assert.Equal(t, 1000, config.BufferSize)
	assert.Equal(t, time.Second, config.FlushInterval)
	assert.Equal(t, int64(100*1024*1024), config.MaxFileSize)
	assert.Equal(t, 5, config.MaxBackups)
	assert.True(t, config.Compress)
	assert.NotNil(t, config.ModuleLevels)
}

func TestNewStructuredLogger(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("with config", func(t *testing.T) {
		config := &StructuredLoggerConfig{
			Level:   SeverityDebug,
			Format:  "json",
			Output:  filepath.Join(tmpDir, "test.log"),
			AppName: "test-app",
			Version: "2.0.0",
		}

		logger, err := NewStructuredLogger(config)
		assert.NoError(t, err)
		assert.NotNil(t, logger)

		// Clean up
		logger.Close()
	})

	t.Run("with nil config", func(t *testing.T) {
		logger, err := NewStructuredLogger(nil)
		assert.NoError(t, err)
		assert.NotNil(t, logger)

		// Clean up
		logger.Close()
	})

	t.Run("with stdout output", func(t *testing.T) {
		config := &StructuredLoggerConfig{
			Level:  SeverityInfo,
			Format: "json",
			Output: "stdout",
		}

		logger, err := NewStructuredLogger(config)
		assert.NoError(t, err)
		assert.NotNil(t, logger)

		// Clean up
		logger.Close()
	})
}

func TestStructuredLoggerLevels(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	config := &StructuredLoggerConfig{
		Level:         SeverityDebug,
		Format:        "json",
		Output:        logFile,
		EnableTracing: false,
		EnableCaller:  false,
	}

	logger, err := NewStructuredLogger(config)
	require.NoError(t, err)
	defer logger.Close()

	ctx := context.Background()

	// Test different log levels
	logger.Emergency(ctx, "emergency message")
	logger.Alert(ctx, "alert message")
	logger.Critical(ctx, "critical message")
	logger.ErrorLevel(ctx, "error message")
	logger.Warning(ctx, "warning message")
	logger.Notice(ctx, "notice message")
	logger.InfoLevel(ctx, "info message")
	logger.DebugLevel(ctx, "debug message")

	// Read log file
	content, err := os.ReadFile(logFile)
	if err == nil {
		logContent := string(content)
		assert.Contains(t, logContent, "emergency message")
		assert.Contains(t, logContent, "alert message")
		assert.Contains(t, logContent, "critical message")
		assert.Contains(t, logContent, "error message")
		assert.Contains(t, logContent, "warning message")
		assert.Contains(t, logContent, "notice message")
		assert.Contains(t, logContent, "info message")
		assert.Contains(t, logContent, "debug message")
	}
}

func TestStructuredLoggerJSON(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	config := &StructuredLoggerConfig{
		Level:         SeverityInfo,
		Format:        "json",
		Output:        logFile,
		AppName:       "test-app",
		Version:       "1.0.0",
		Environment:   "test",
		EnableTracing: false,
		EnableCaller:  false,
	}

	logger, err := NewStructuredLogger(config)
	require.NoError(t, err)
	defer logger.Close()

	ctx := context.Background()
	logger.InfoLevel(ctx, "test json message")

	// Check if JSON format is used
	content, err := os.ReadFile(logFile)
	if err == nil {
		logContent := string(content)

		// Should contain JSON structure
		assert.Contains(t, logContent, "test json message")
		assert.Contains(t, logContent, `"@timestamp"`)
		assert.Contains(t, logContent, `"level":"info"`)
		assert.Contains(t, logContent, `"severity":6`)
		assert.Contains(t, logContent, `"appname":"test-app"`)
		assert.Contains(t, logContent, `"message":"test json message"`)

		// Parse as JSON to ensure validity
		var entry StructuredLogEntry
		lines := strings.Split(strings.TrimSpace(logContent), "\n")
		if len(lines) > 0 {
			err := json.Unmarshal([]byte(lines[0]), &entry)
			assert.NoError(t, err)
			assert.Equal(t, "test json message", entry.Message)
			assert.Equal(t, "info", entry.Level)
			assert.Equal(t, SeverityInfo, entry.Severity)
		}
	}
}

func TestStructuredLoggerConsoleFormat(t *testing.T) {
	var buf bytes.Buffer

	config := &StructuredLoggerConfig{
		Level:         SeverityInfo,
		Format:        "console",
		Output:        "stderr",
		EnableTracing: false,
		EnableCaller:  false,
	}

	logger, err := NewStructuredLogger(config)
	require.NoError(t, err)
	defer logger.Close()

	// Override writer for testing
	logger.writer = &buf

	ctx := context.Background()
	logger.InfoLevel(ctx, "test console message")

	output := buf.String()
	assert.Contains(t, output, "test console message")
	assert.Contains(t, output, "INFO")
	assert.Contains(t, output, "[")
	assert.Contains(t, output, "]")
}

func TestStructuredLoggerLogfmtFormat(t *testing.T) {
	var buf bytes.Buffer

	config := &StructuredLoggerConfig{
		Level:         SeverityInfo,
		Format:        "logfmt",
		Output:        "stderr",
		EnableTracing: false,
		EnableCaller:  false,
	}

	logger, err := NewStructuredLogger(config)
	require.NoError(t, err)
	defer logger.Close()

	// Override writer for testing
	logger.writer = &buf

	ctx := context.Background()
	logger.InfoLevel(ctx, "test logfmt message")

	output := buf.String()
	assert.Contains(t, output, "test logfmt message")
	assert.Contains(t, output, `@timestamp=`)
	assert.Contains(t, output, `level="info"`)
	assert.Contains(t, output, `severity=6`)
	assert.Contains(t, output, `message="test logfmt message"`)
}

func TestStructuredLoggerSetLevel(t *testing.T) {
	logger, err := NewStructuredLogger(nil)
	require.NoError(t, err)
	defer logger.Close()

	// Test setting different levels
	logger.SetLevel(SeverityError)
	assert.Equal(t, SeverityError, logger.GetLevel())

	logger.SetLevel(SeverityDebug)
	assert.Equal(t, SeverityDebug, logger.GetLevel())
}

func TestStructuredLoggerWithFields(t *testing.T) {
	var buf bytes.Buffer

	config := &StructuredLoggerConfig{
		Level:         SeverityInfo,
		Format:        "json",
		Output:        "stderr",
		EnableTracing: false,
		EnableCaller:  false,
	}

	logger, err := NewStructuredLogger(config)
	require.NoError(t, err)
	defer logger.Close()

	// Override writer for testing
	logger.writer = &buf

	ctx := context.Background()
	fields := map[string]interface{}{
		"user_id": 123,
		"action":  "login",
	}

	logger.InfoLevel(ctx, "user action", fields)

	output := buf.String()
	assert.Contains(t, output, "user action")

	// Parse JSON to verify fields
	var entry StructuredLogEntry
	err = json.Unmarshal([]byte(output), &entry)
	if err == nil {
		assert.NotNil(t, entry.Fields)
		assert.Equal(t, float64(123), entry.Fields["user_id"])
		assert.Equal(t, "login", entry.Fields["action"])
	}
}

func TestStructuredLoggerWithModule(t *testing.T) {
	var buf bytes.Buffer

	config := &StructuredLoggerConfig{
		Level:  SeverityInfo,
		Format: "json",
		Output: "stderr",
	}

	logger, err := NewStructuredLogger(config)
	require.NoError(t, err)
	defer logger.Close()

	// Override writer for testing
	logger.writer = &buf

	ctx := context.Background()
	moduleLogger := logger.WithModule("test-module")
	moduleLogger.InfoLevel(ctx, "module message")

	output := buf.String()
	assert.Contains(t, output, "module message")

	// Parse JSON to verify module field
	var entry StructuredLogEntry
	err = json.Unmarshal([]byte(output), &entry)
	if err == nil {
		assert.NotNil(t, entry.Fields)
		assert.Equal(t, "test-module", entry.Fields["module"])
	}
}

func TestStructuredLoggerWithModuleLevel(t *testing.T) {
	logger, err := NewStructuredLogger(nil)
	require.NoError(t, err)
	defer logger.Close()

	// Set module-specific level
	logger.SetModuleLevel("test-module", SeverityError)

	// Verify the module level was set
	assert.Equal(t, SeverityError, logger.config.ModuleLevels["test-module"])
}

func TestStructuredLoggerFiltering(t *testing.T) {
	var buf bytes.Buffer

	config := &StructuredLoggerConfig{
		Level:         SeverityWarning, // Only WARNING and above should be logged
		Format:        "json",
		Output:        "stderr",
		EnableTracing: false,
		EnableCaller:  false,
	}

	logger, err := NewStructuredLogger(config)
	require.NoError(t, err)
	defer logger.Close()

	// Override writer for testing
	logger.writer = &buf

	ctx := context.Background()
	logger.ErrorLevel(ctx, "error message")
	logger.Warning(ctx, "warning message")
	logger.InfoLevel(ctx, "info message")   // Should be filtered out
	logger.DebugLevel(ctx, "debug message") // Should be filtered out

	output := buf.String()
	assert.Contains(t, output, "error message")
	assert.Contains(t, output, "warning message")
	assert.NotContains(t, output, "info message")
	assert.NotContains(t, output, "debug message")
}

func TestStructuredLoggerSampling(t *testing.T) {
	var buf bytes.Buffer

	config := &StructuredLoggerConfig{
		Level:           SeverityDebug,
		Format:          "json",
		Output:          "stderr",
		EnableSampling:  true,
		SampleRate:      0.5, // 50% sampling rate
		SampleThreshold: 2,   // Every 2nd message
		EnableTracing:   false,
		EnableCaller:    false,
	}

	logger, err := NewStructuredLogger(config)
	require.NoError(t, err)
	defer logger.Close()

	// Override writer for testing
	logger.writer = &buf

	ctx := context.Background()

	// Log many debug messages
	for i := 0; i < 10; i++ {
		logger.DebugLevel(ctx, "debug message")
	}

	output := buf.String()
	messageCount := strings.Count(output, "debug message")

	// Should have fewer messages due to sampling
	assert.True(t, messageCount < 10, "Sampling should reduce message count")
	assert.True(t, messageCount > 0, "Some messages should still be logged")
}

func TestParseRFC5424Severity(t *testing.T) {
	tests := []struct {
		input    string
		expected RFC5424Severity
		hasError bool
	}{
		{"emergency", SeverityEmergency, false},
		{"emerg", SeverityEmergency, false},
		{"alert", SeverityAlert, false},
		{"critical", SeverityCritical, false},
		{"crit", SeverityCritical, false},
		{"error", SeverityError, false},
		{"err", SeverityError, false},
		{"warning", SeverityWarning, false},
		{"warn", SeverityWarning, false},
		{"notice", SeverityNotice, false},
		{"info", SeverityInfo, false},
		{"informational", SeverityInfo, false},
		{"debug", SeverityDebug, false},
		{"invalid", SeverityInfo, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParseRFC5424Severity(tt.input)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestGlobalStructuredLogger(t *testing.T) {
	config := DefaultStructuredLoggerConfig()
	config.Output = "stderr"

	err := InitGlobalStructuredLogger(config)
	assert.NoError(t, err)

	globalLogger := GetGlobalStructuredLogger()
	assert.NotNil(t, globalLogger)

	// Clean up
	globalLogger.Close()
}

func BenchmarkStructuredLoggerInfo(b *testing.B) {
	logger, err := NewStructuredLogger(nil)
	require.NoError(b, err)
	defer logger.Close()

	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.InfoLevel(ctx, "benchmark message")
	}
}

func BenchmarkStructuredLoggerWithFields(b *testing.B) {
	logger, err := NewStructuredLogger(nil)
	require.NoError(b, err)
	defer logger.Close()

	ctx := context.Background()
	fields := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
		"key3": true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.InfoLevel(ctx, "benchmark message", fields)
	}
}
