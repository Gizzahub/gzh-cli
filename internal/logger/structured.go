// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
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
	LevelDebug LogLevel = "debug"
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
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

// NewStructuredLogger creates a new structured logger.
func NewStructuredLogger(component string, level LogLevel) *StructuredLogger {
	var slogLevel slog.Level

	switch level {
	case LevelDebug:
		slogLevel = slog.LevelDebug
	case LevelInfo:
		slogLevel = slog.LevelInfo
	case LevelWarn:
		slogLevel = slog.LevelWarn
	case LevelError:
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level:     slogLevel,
		AddSource: true,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)

	return &StructuredLogger{
		logger:    logger,
		level:     slogLevel,
		context:   make(map[string]interface{}),
		sessionID: generateSessionID(),
		component: component,
	}
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

	perfInfo := &PerformanceInfo{
		Duration: duration,
		MemoryUsage: func() int64 {
			const maxInt64 = uint64(1<<63 - 1) // Max value for int64
			if m.Alloc > maxInt64 {            // Check for overflow
				return 1<<63 - 1 // Return max int64 if overflow would occur
			}
			return int64(m.Alloc)
		}(),
		Operation: operation,
		Metrics:   metrics,
	}

	entry := &LogEntry{
		Timestamp:   time.Now(),
		Level:       "info",
		Message:     fmt.Sprintf("Performance: %s completed in %v", operation, duration),
		Component:   l.component,
		SessionID:   l.sessionID,
		Context:     l.context,
		Performance: perfInfo,
	}

	l.writeStructuredLog(entry)
}

// log writes a log message with context.
func (l *StructuredLogger) log(level slog.Level, msg string, args ...interface{}) {
	if !l.logger.Enabled(context.Background(), level) {
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
	data, err := json.Marshal(entry)
	if err != nil {
		// Fallback to simple logging
		l.logger.Error("Failed to marshal log entry", "error", err, "message", entry.Message)
		return
	}

	// Write to stdout
	if _, writeErr := fmt.Println(string(data)); writeErr != nil {
		// Silent fallback - avoid recursive logging
		l.logger.Error("Failed to write log entry", "error", writeErr)
	}
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
func generateSessionID() string {
	return fmt.Sprintf("sess_%d_%d", time.Now().Unix(), time.Now().Nanosecond()%1000000)
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

func SetGlobalLogger(logger *StructuredLogger) {
	globalLogger = logger
}

func GetGlobalLogger() *StructuredLogger {
	if globalLogger == nil {
		globalLogger = NewStructuredLogger("global", LevelInfo)
	}

	return globalLogger
}

// Convenience functions for global logger.
func Debug(msg string, args ...interface{}) {
	GetGlobalLogger().Debug(msg, args...)
}

func Info(msg string, args ...interface{}) {
	GetGlobalLogger().Info(msg, args...)
}

func Warn(msg string, args ...interface{}) {
	GetGlobalLogger().Warn(msg, args...)
}

func Error(msg string, args ...interface{}) {
	GetGlobalLogger().Error(msg, args...)
}

func ErrorWithStack(err error, msg string, args ...interface{}) {
	GetGlobalLogger().ErrorWithStack(err, msg, args...)
}
