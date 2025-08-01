// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package logger

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewSimpleLogger(t *testing.T) {
	logger := NewSimpleLogger("test-component")

	assert.NotNil(t, logger)
	assert.Equal(t, "test-component", logger.component)
	assert.NotNil(t, logger.context)
	assert.NotEmpty(t, logger.sessionID)
	assert.NotNil(t, logger.config)
}

func TestSimpleLogger_WithContext(t *testing.T) {
	logger := NewSimpleLogger("test")

	newLogger := logger.WithContext("key", "value")

	assert.NotNil(t, newLogger)
	assert.Equal(t, "value", newLogger.context["key"])
	assert.Equal(t, logger.component, newLogger.component)
}

func TestSimpleLogger_LogLevels(t *testing.T) {
	// Enable debug mode to show all log levels
	SetGlobalLoggingFlags(false, true, false)
	defer SetGlobalLoggingFlags(false, false, false) // Reset after test

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	logger := NewSimpleLogger("test")

	tests := []struct {
		name     string
		logFunc  func(string, ...interface{})
		message  string
		expected string
	}{
		{
			name:     "Info level",
			logFunc:  logger.Info,
			message:  "test info message",
			expected: "INFO",
		},
		{
			name:     "Warning level",
			logFunc:  logger.Warn,
			message:  "test warning message",
			expected: "WARN",
		},
		{
			name:     "Error level",
			logFunc:  logger.Error,
			message:  "test error message",
			expected: "ERROR",
		},
		{
			name:     "Debug level",
			logFunc:  logger.Debug,
			message:  "test debug message",
			expected: "DEBUG",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset the pipe
			r, w, _ = os.Pipe()
			os.Stdout = w

			tt.logFunc(tt.message)

			w.Close()
			os.Stdout = old

			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			assert.Contains(t, output, tt.expected)
			assert.Contains(t, output, tt.message)
		})
	}
}

func TestGlobalLogLevelFlags(t *testing.T) {
	// Test SetGlobalLoggingFlags
	SetGlobalLoggingFlags(true, false, false)
	assert.True(t, globalVerbose)
	assert.False(t, globalDebug)
	assert.False(t, globalQuiet)

	SetGlobalLoggingFlags(false, true, false)
	assert.False(t, globalVerbose)
	assert.True(t, globalDebug)
	assert.False(t, globalQuiet)

	SetGlobalLoggingFlags(false, false, true)
	assert.False(t, globalVerbose)
	assert.False(t, globalDebug)
	assert.True(t, globalQuiet)
}

// formatMessage is a private method, so we test it indirectly through print output.
func TestSimpleLogger_printFormatting(t *testing.T) {
	// Enable debug mode to show all log levels
	SetGlobalLoggingFlags(false, true, false)
	defer SetGlobalLoggingFlags(false, false, false) // Reset after test

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	logger := NewSimpleLogger("test")
	logger.Info("test message")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	assert.Contains(t, output, "INFO")
	assert.Contains(t, output, "test message")
	assert.Contains(t, output, "test") // component name
}

func TestGenerateSimpleSessionID(t *testing.T) {
	id1 := generateSimpleSessionID("component1")
	id2 := generateSimpleSessionID("component2")

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2) // Should be different
}

func TestSimpleLogger_shouldLog(t *testing.T) {
	logger := NewSimpleLogger("test")

	// Test with default settings (CLI logging disabled by default, only errors)
	assert.False(t, logger.shouldLog("INFO")) // Default config disables INFO
	assert.False(t, logger.shouldLog("WARN")) // Default config disables WARN
	assert.True(t, logger.shouldLog("ERROR")) // ERROR should always show

	// Test with debug mode enabled
	SetGlobalLoggingFlags(false, true, false)
	logger = NewSimpleLogger("test")
	assert.True(t, logger.shouldLog("INFO")) // Debug mode shows all levels
	assert.True(t, logger.shouldLog("WARN"))
	assert.True(t, logger.shouldLog("ERROR"))
	assert.True(t, logger.shouldLog("DEBUG"))

	// Test with quiet mode
	SetGlobalLoggingFlags(false, false, true)
	logger = NewSimpleLogger("test")
	assert.False(t, logger.shouldLog("INFO"))
	assert.True(t, logger.shouldLog("ERROR")) // Errors should still show

	// Reset
	SetGlobalLoggingFlags(false, false, false)
}

func TestLogLevelConstants(t *testing.T) {
	assert.Equal(t, "DEBUG", SimpleLevelDebug)
	assert.Equal(t, "INFO", SimpleLevelInfo)
	assert.Equal(t, "WARN", SimpleLevelWarn)
	assert.Equal(t, "ERROR", SimpleLevelError)
}

func TestSimpleLogger_WithSession(t *testing.T) {
	logger := NewSimpleLogger("test")
	originalSessionID := logger.sessionID

	newLogger := logger.WithSession("custom-session-123")

	assert.Equal(t, "custom-session-123", newLogger.sessionID)
	assert.Equal(t, originalSessionID, logger.sessionID) // Original unchanged
	assert.Equal(t, logger.component, newLogger.component)
}

func TestSimpleLogger_ErrorWithStack(t *testing.T) {
	// Enable debug mode to show all log levels
	SetGlobalLoggingFlags(false, true, false)
	defer SetGlobalLoggingFlags(false, false, false)

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	logger := NewSimpleLogger("test")
	testErr := fmt.Errorf("test error")
	logger.ErrorWithStack(testErr, "operation failed")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	assert.Contains(t, output, "ERROR")
	assert.Contains(t, output, "operation failed: test error")
}

func TestSimpleLogger_LogPerformance(t *testing.T) {
	// Enable debug mode to show all log levels
	SetGlobalLoggingFlags(false, true, false)
	defer SetGlobalLoggingFlags(false, false, false)

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	logger := NewSimpleLogger("test")
	metrics := map[string]interface{}{
		"items_processed": 100,
		"success_rate":    95.5,
	}
	logger.LogPerformance("test-operation", 500*time.Millisecond, metrics)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	assert.Contains(t, output, "PERF")
	assert.Contains(t, output, "test-operation")
	assert.Contains(t, output, "500ms")
	assert.Contains(t, output, "Memory:")
	assert.Contains(t, output, "items_processed=100")
	assert.Contains(t, output, "success_rate=95.5")
}

func TestSimpleLogger_LoggerMiddleware(t *testing.T) {
	// Enable debug mode to show all log levels
	SetGlobalLoggingFlags(false, true, false)
	defer SetGlobalLoggingFlags(false, false, false)

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	logger := NewSimpleLogger("test")

	// Test successful operation
	err := logger.LoggerMiddleware(func() error {
		return nil
	})

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	assert.NoError(t, err)
	assert.Contains(t, output, "Operation started")
	assert.Contains(t, output, "Operation completed successfully")
	assert.Contains(t, output, "PERF")
	assert.Contains(t, output, "operation_completed")
}

func TestSimpleLogger_LoggerMiddleware_WithError(t *testing.T) {
	// Enable debug mode to show all log levels
	SetGlobalLoggingFlags(false, true, false)
	defer SetGlobalLoggingFlags(false, false, false)

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	logger := NewSimpleLogger("test")
	testErr := fmt.Errorf("operation failed")

	// Test failed operation
	err := logger.LoggerMiddleware(func() error {
		return testErr
	})

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	assert.Error(t, err)
	assert.Equal(t, testErr, err)
	assert.Contains(t, output, "Operation started")
	assert.Contains(t, output, "Operation failed: operation failed")
	assert.Contains(t, output, "PERF")
	assert.Contains(t, output, "operation_failed")
}

func TestGlobalSimpleLogger(t *testing.T) {
	// Test GetGlobalSimpleLogger creates a logger if none exists
	globalSimpleLogger = nil // Reset
	logger1 := GetGlobalSimpleLogger()
	assert.NotNil(t, logger1)
	assert.Equal(t, "global", logger1.component)

	// Test GetGlobalSimpleLogger returns the same instance
	logger2 := GetGlobalSimpleLogger()
	assert.Equal(t, logger1, logger2)

	// Test SetGlobalSimpleLogger
	customLogger := NewSimpleLogger("custom")
	SetGlobalSimpleLogger(customLogger)
	logger3 := GetGlobalSimpleLogger()
	assert.Equal(t, customLogger, logger3)
	assert.Equal(t, "custom", logger3.component)
}

func TestGlobalLoggerFunctions(t *testing.T) {
	// Enable debug mode to show all log levels
	SetGlobalLoggingFlags(false, true, false)
	defer SetGlobalLoggingFlags(false, false, false)

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Test global logging functions
	SimpleDebug("debug message")
	SimpleInfo("info message")
	SimpleWarn("warn message")
	SimpleError("error message")
	SimpleErrorWithStack(fmt.Errorf("test error"), "error with stack")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	assert.Contains(t, output, "DEBUG")
	assert.Contains(t, output, "debug message")
	assert.Contains(t, output, "INFO")
	assert.Contains(t, output, "info message")
	assert.Contains(t, output, "WARN")
	assert.Contains(t, output, "warn message")
	assert.Contains(t, output, "ERROR")
	assert.Contains(t, output, "error message")
	assert.Contains(t, output, "error with stack: test error")
}

func TestSimpleLogger_ComponentShortening(t *testing.T) {
	// Enable debug mode to show all log levels
	SetGlobalLoggingFlags(false, true, false)
	defer SetGlobalLoggingFlags(false, false, false)

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	tests := []struct {
		component string
		expected  string
	}{
		{"bulk-clone-github", "github"},
		{"bulk-clone-gitlab", "gitlab"},
		{"bulk-clone-gitea", "gitea"},
		{"doctor", "doctor"},
		{"unknown-component", "unknown-component"},
	}

	for _, test := range tests {
		t.Run(test.component, func(t *testing.T) {
			// Reset the pipe
			r, w, _ = os.Pipe()
			os.Stdout = w

			logger := NewSimpleLogger(test.component)
			logger.Info("test message")

			w.Close()
			os.Stdout = old

			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			assert.Contains(t, output, test.expected)
		})
	}
}

func TestSimpleLogger_ContextHandling(t *testing.T) {
	// Enable debug mode to show all log levels
	SetGlobalLoggingFlags(false, true, false)
	defer SetGlobalLoggingFlags(false, false, false)

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	logger := NewSimpleLogger("test")
	loggerWithContext := logger.WithContext("org_name", "test-org")
	loggerWithContext.Info("test message", "attempt", 1, "max_retries", 3)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	assert.Contains(t, output, "test-org")      // org_name context should be shown
	assert.Contains(t, output, "attempt=1")     // important arg should be shown
	assert.Contains(t, output, "max_retries=3") // important arg should be shown
}

func TestSimpleLogger_ConfigurationScenarios(t *testing.T) {
	tests := []struct {
		name        string
		verbose     bool
		debug       bool
		quiet       bool
		level       string
		expectInfo  bool
		expectDebug bool
		expectWarn  bool
		expectError bool
	}{
		{"default", false, false, false, "ERROR", false, false, false, true},
		{"verbose", true, false, false, "INFO", true, false, true, true},
		{"debug", false, true, false, "DEBUG", true, true, true, true},
		{"quiet", false, false, true, "ERROR", false, false, false, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			SetGlobalLoggingFlags(test.verbose, test.debug, test.quiet)
			defer SetGlobalLoggingFlags(false, false, false)

			logger := NewSimpleLogger("test")

			assert.Equal(t, test.expectInfo, logger.shouldLog("INFO"))
			assert.Equal(t, test.expectDebug, logger.shouldLog("DEBUG"))
			assert.Equal(t, test.expectWarn, logger.shouldLog("WARN"))
			assert.Equal(t, test.expectError, logger.shouldLog("ERROR"))
		})
	}
}

func TestSimpleLogger_DebugArgFiltering(t *testing.T) {
	// Enable debug mode to show all log levels
	SetGlobalLoggingFlags(false, true, false)
	defer SetGlobalLoggingFlags(false, false, false)

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	logger := NewSimpleLogger("test")

	// Test debug level with debug args (should show important ones)
	logger.Debug("debug message", "optimized", true, "attempt", 1, "streaming", false, "repo_name", "test-repo")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	assert.Contains(t, output, "attempt=1")           // important arg should be shown
	assert.Contains(t, output, "repo_name=test-repo") // important arg should be shown
	// Debug-only args like "optimized" and "streaming" may or may not be shown based on importance
}

func TestSimpleLogger_EdgeCases(t *testing.T) {
	// Test with nil config (should fall back to defaults)
	logger := &SimpleLogger{
		component: "test",
		context:   make(map[string]interface{}),
		sessionID: "test-session",
		config:    nil, // nil config
	}

	// Should only log errors and warnings with nil config
	assert.False(t, logger.shouldLog("INFO"))
	assert.False(t, logger.shouldLog("DEBUG"))
	assert.True(t, logger.shouldLog("WARN"))
	assert.True(t, logger.shouldLog("ERROR"))

	// Test empty component name
	emptyLogger := NewSimpleLogger("")
	assert.NotNil(t, emptyLogger)
	assert.Equal(t, "", emptyLogger.component)
}

func TestSimpleLogger_ArgumentParsing(t *testing.T) {
	// Enable debug mode to show all log levels
	SetGlobalLoggingFlags(false, true, false)
	defer SetGlobalLoggingFlags(false, false, false)

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	logger := NewSimpleLogger("test")

	// Test with odd number of arguments (should handle gracefully)
	logger.Info("test message", "key1", "value1", "key2") // missing value for key2

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	assert.Contains(t, output, "test message")
	// Should handle gracefully without panicking
}

func TestSimpleLogger_PerformanceWithoutMetrics(t *testing.T) {
	// Enable debug mode to show all log levels
	SetGlobalLoggingFlags(false, true, false)
	defer SetGlobalLoggingFlags(false, false, false)

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	logger := NewSimpleLogger("test")

	// Test performance logging without metrics
	logger.LogPerformance("test-operation", 100*time.Millisecond, nil)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	assert.Contains(t, output, "PERF")
	assert.Contains(t, output, "test-operation")
	assert.Contains(t, output, "100ms")
	assert.Contains(t, output, "Memory:")
	// Should not contain metrics section when nil
}
