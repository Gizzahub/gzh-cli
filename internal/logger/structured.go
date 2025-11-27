// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

// Package logger provides structured logging capabilities with JSON formatting and context support.
package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// StructuredLogger provides advanced logging capabilities.
type StructuredLogger struct {
	logger    *slog.Logger
	level     slog.Level
	context   map[string]interface{}
	sessionID string
	component string
}

// LogLevel represents logging levels.
type LogLevel string

const (
	// LevelDebug represents debug log level.
	LevelDebug LogLevel = "debug"
	// LevelInfo represents info log level.
	LevelInfo LogLevel = "info"
	// LevelWarn represents warning log level.
	LevelWarn LogLevel = "warn"
	// LevelError represents error log level.
	LevelError LogLevel = "error"
)

// LogEntry represents a structured log entry.
type LogEntry struct {
	Timestamp   time.Time              `json:"timestamp"`
	Level       string                 `json:"level"`
	Message     string                 `json:"message"`
	Component   string                 `json:"component"`
	SessionID   string                 `json:"sessionId"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Caller      *CallerInfo            `json:"caller,omitempty"`
	Error       *ErrorInfo             `json:"error,omitempty"`
	Performance *PerformanceInfo       `json:"performance,omitempty"`
}

// CallerInfo represents caller information.
type CallerInfo struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Function string `json:"function"`
}

// ErrorInfo represents error information.
type ErrorInfo struct {
	Type       string `json:"type"`
	Message    string `json:"message"`
	StackTrace string `json:"stackTrace,omitempty"`
	Code       string `json:"code,omitempty"`
}

// PerformanceInfo represents performance metrics.
type PerformanceInfo struct {
	Duration    time.Duration          `json:"duration"`
	MemoryUsage int64                  `json:"memoryUsage"`
	Operation   string                 `json:"operation"`
	Metrics     map[string]interface{} `json:"metrics,omitempty"`
}

// NewStructuredLogger creates a new structured logger with dual output.
func NewStructuredLogger(component string, level LogLevel) *StructuredLogger {
	// Use dual logger that outputs to both console and file
	logger, err := NewDualLogger(component, level)
	if err != nil {
		// Fallback to console-only logger if dual logger fails
		return NewConsoleOnlyLogger(component, level)
	}
	return logger
}

// WithContext adds context to the logger.
func (l *StructuredLogger) WithContext(key string, value interface{}) *StructuredLogger {
	newLogger := *l
	// Preallocate with appropriate capacity
	newLogger.context = make(map[string]interface{}, len(l.context)+1)
	for k, v := range l.context {
		newLogger.context[k] = v
	}

	newLogger.context[key] = value

	return &newLogger
}

// WithSession sets a session ID.
func (l *StructuredLogger) WithSession(sessionID string) *StructuredLogger {
	newLogger := *l
	newLogger.sessionID = sessionID

	return &newLogger
}

// Debug logs a debug message.
func (l *StructuredLogger) Debug(msg string, args ...interface{}) {
	l.log(slog.LevelDebug, msg, args...)
}

// Info logs an info message.
func (l *StructuredLogger) Info(msg string, args ...interface{}) {
	l.log(slog.LevelInfo, msg, args...)
}

// Warn logs a warning message.
func (l *StructuredLogger) Warn(msg string, args ...interface{}) {
	l.log(slog.LevelWarn, msg, args...)
}

// Error logs an error message.
func (l *StructuredLogger) Error(msg string, args ...interface{}) {
	l.log(slog.LevelError, msg, args...)
}

// ErrorWithStack logs an error with stack trace.
func (l *StructuredLogger) ErrorWithStack(err error, msg string, args ...interface{}) {
	l.logWithError(slog.LevelError, err, msg, args...)
}

// LogPerformance logs performance metrics.
func (l *StructuredLogger) LogPerformance(operation string, duration time.Duration, metrics map[string]interface{}) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	memoryMB := float64(m.Alloc) / 1024 / 1024
	msg := fmt.Sprintf("Operation '%s' completed in %v (Memory: %.2f MB)", operation, duration, memoryMB)

	if len(metrics) > 0 {
		var metricParts []string
		for k, v := range metrics {
			metricParts = append(metricParts, fmt.Sprintf("%s=%v", k, v))
		}
		msg += fmt.Sprintf(" [%s]", strings.Join(metricParts, " "))
	}

	// Use simple log format instead of JSON structured format for performance logs
	l.Info(msg)
}

// log writes a log message with context.
func (l *StructuredLogger) log(level slog.Level, msg string, args ...interface{}) {
	if !l.logger.Enabled(context.Background(), level) {
		return
	}

	// Check global logging flags - only show logs in debug/verbose mode or for errors
	if !l.shouldShowLog(level) {
		return
	}

	caller := getCaller(2)

	// Preallocate slice with estimated capacity
	capacity := 5 + len(l.context) + len(args)/2
	attrs := make([]slog.Attr, 0, capacity)

	attrs = append(attrs,
		slog.String("component", l.component),
		slog.String("sessionId", l.sessionID),
		slog.String("callerFile", caller.File),
		slog.Int("callerLine", caller.Line),
		slog.String("callerFunction", caller.Function),
	)

	// Add context attributes
	for k, v := range l.context {
		attrs = append(attrs, slog.Any(k, v))
	}

	// Add additional args
	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			if key, ok := args[i].(string); ok {
				attrs = append(attrs, slog.Any(key, args[i+1]))
			}
		}
	}

	l.logger.LogAttrs(context.Background(), level, msg, attrs...)
}

// logWithError logs a message with error information.
func (l *StructuredLogger) logWithError(level slog.Level, err error, msg string, _ ...interface{}) {
	if !l.logger.Enabled(context.Background(), level) {
		return
	}

	caller := getCaller(2)
	errorInfo := &ErrorInfo{
		Type:    fmt.Sprintf("%T", err),
		Message: err.Error(),
	}

	entry := &LogEntry{
		Timestamp: time.Now(),
		Level:     level.String(),
		Message:   msg,
		Component: l.component,
		SessionID: l.sessionID,
		Context:   l.context,
		Caller:    caller,
		Error:     errorInfo,
	}

	l.writeStructuredLog(entry)
}

// writeStructuredLog writes a structured log entry.
func (l *StructuredLogger) writeStructuredLog(entry *LogEntry) {
	// Only output JSON logs in debug mode
	// For normal users, errors should be handled with user-friendly messages
	if !IsDebugEnabled() {
		return
	}

	data, err := json.Marshal(entry)
	if err != nil {
		// Fallback to simple logging
		ctx := context.Background()
		l.logger.ErrorContext(ctx, "Failed to marshal log entry", "error", err, "message", entry.Message)
		return
	}

	// Write to stdout
	if _, writeErr := fmt.Println(string(data)); writeErr != nil {
		// Silent fallback - avoid recursive logging
		ctx := context.Background()
		l.logger.ErrorContext(ctx, "Failed to write log entry", "error", writeErr)
	}
}

// shouldShowLog determines if a log message should be shown based on global flags.
func (l *StructuredLogger) shouldShowLog(level slog.Level) bool {
	// Always show errors
	if level == slog.LevelError {
		return true
	}

	// Show all logs in debug mode
	if IsDebugEnabled() {
		return true
	}

	// Show info and warn logs in verbose mode
	if IsVerboseEnabled() {
		return level >= slog.LevelInfo
	}

	// Default: don't show logs (only console messages should appear)
	return false
}

// getCaller gets caller information.
func getCaller(skip int) *CallerInfo {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return &CallerInfo{
			File:     "unknown",
			Line:     0,
			Function: "unknown",
		}
	}

	fn := runtime.FuncForPC(pc)

	fnName := "unknown"
	if fn != nil {
		fnName = fn.Name()
	}

	return &CallerInfo{
		File:     filepath.Base(file),
		Line:     line,
		Function: fnName,
	}
}

// generateSessionID generates a unique session ID.
func generateSessionID(component string) string {
	return fmt.Sprintf("%s_%d_%d", component, time.Now().Unix(), time.Now().Nanosecond()%1000000)
}

// LoggerMiddleware provides logging middleware functionality.
func (l *StructuredLogger) LoggerMiddleware(next func() error) error {
	start := time.Now()

	l.Info("Operation started")

	err := next()

	duration := time.Since(start)
	if err != nil {
		l.LogPerformance("operation_failed", duration, map[string]interface{}{
			"success": false,
		})
		l.ErrorWithStack(err, "Operation failed")

		return err
	}

	l.LogPerformance("operation_completed", duration, map[string]interface{}{
		"success": true,
	})
	l.Info("Operation completed successfully")

	return nil
}

// SetGlobalLogger sets a global logger instance.
var globalLogger *StructuredLogger

// SetGlobalLogger sets a global logger instance.
func SetGlobalLogger(logger *StructuredLogger) {
	globalLogger = logger
}

// GetGlobalLogger returns the global logger instance.
func GetGlobalLogger() *StructuredLogger {
	if globalLogger == nil {
		globalLogger = NewStructuredLogger("global", LevelInfo)
	}

	return globalLogger
}

// Debug logs a debug message using the global logger.
func Debug(msg string, args ...interface{}) {
	GetGlobalLogger().Debug(msg, args...)
}

// Info logs an info message using the global logger.
func Info(msg string, args ...interface{}) {
	GetGlobalLogger().Info(msg, args...)
}

// Warn logs a warning message using the global logger.
func Warn(msg string, args ...interface{}) {
	GetGlobalLogger().Warn(msg, args...)
}

// Error logs an error message using the global logger.
func Error(msg string, args ...interface{}) {
	GetGlobalLogger().Error(msg, args...)
}

// ErrorWithStack logs an error message with stack trace using the global logger.
func ErrorWithStack(err error, msg string, args ...interface{}) {
	GetGlobalLogger().ErrorWithStack(err, msg, args...)
}
