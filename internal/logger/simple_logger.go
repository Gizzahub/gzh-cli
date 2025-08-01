// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

// Package logger provides simple terminal output logging capabilities.
package logger

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/gizzahub/gzh-manager-go/internal/config"
)

// Log level constants for simple logger (string format).
const (
	SimpleLevelDebug = "DEBUG"
	SimpleLevelInfo  = "INFO"
	SimpleLevelWarn  = "WARN"
	SimpleLevelError = "ERROR"
)

// Global flags for CLI logging control.
var (
	globalVerbose bool
	globalDebug   bool
	globalQuiet   bool
)

// CommonLogger defines the common interface for both structured and simple loggers.
type CommonLogger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	ErrorWithStack(err error, msg string, args ...interface{})
}

// SimpleLogger provides straightforward terminal output for better readability.
type SimpleLogger struct {
	component string
	context   map[string]interface{}
	sessionID string
	config    *config.CLILoggingConfig
}

// Ensure SimpleLogger implements CommonLogger interface.
var _ CommonLogger = (*SimpleLogger)(nil)

// NewSimpleLogger creates a new simple terminal logger.
func NewSimpleLogger(component string) *SimpleLogger {
	globalConfig, err := config.LoadGlobalConfig()
	if err != nil {
		// Use default config if loading fails
		globalConfig = config.DefaultGlobalConfig()
	}
	cliConfig := &globalConfig.Logging.CLILogging

	return &SimpleLogger{
		component: component,
		context:   make(map[string]interface{}),
		sessionID: generateSimpleSessionID(component),
		config:    cliConfig,
	}
}

// WithContext adds context to the logger.
func (l *SimpleLogger) WithContext(key string, value interface{}) *SimpleLogger {
	newLogger := *l
	newLogger.context = make(map[string]interface{}, len(l.context)+1)
	for k, v := range l.context {
		newLogger.context[k] = v
	}
	newLogger.context[key] = value
	return &newLogger
}

// WithSession sets a session ID.
func (l *SimpleLogger) WithSession(sessionID string) *SimpleLogger {
	newLogger := *l
	newLogger.sessionID = sessionID
	return &newLogger
}

// Debug prints a debug message.
func (l *SimpleLogger) Debug(msg string, args ...interface{}) {
	if l.shouldLog(SimpleLevelDebug) {
		l.print(SimpleLevelDebug, msg, args...)
	}
}

// Info prints an info message.
func (l *SimpleLogger) Info(msg string, args ...interface{}) {
	if l.shouldLog(SimpleLevelInfo) {
		l.print(SimpleLevelInfo, msg, args...)
	}
}

// Warn prints a warning message.
func (l *SimpleLogger) Warn(msg string, args ...interface{}) {
	if l.shouldLog(SimpleLevelWarn) {
		l.print(SimpleLevelWarn, msg, args...)
	}
}

// Error prints an error message.
func (l *SimpleLogger) Error(msg string, args ...interface{}) {
	if l.shouldLog(SimpleLevelError) {
		l.print(SimpleLevelError, msg, args...)
	}
}

// ErrorWithStack prints an error message with error details.
func (l *SimpleLogger) ErrorWithStack(err error, msg string, args ...interface{}) {
	if l.shouldLog(SimpleLevelError) {
		fullMsg := fmt.Sprintf("%s: %v", msg, err)
		l.print(SimpleLevelError, fullMsg, args...)
	}
}

// LogPerformance prints performance information.
func (l *SimpleLogger) LogPerformance(operation string, duration time.Duration, metrics map[string]interface{}) {
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

	if l.shouldLog(SimpleLevelInfo) { // Performance logs are considered INFO level
		l.print("PERF", msg)
	}
}

// print outputs a formatted message to the terminal.
func (l *SimpleLogger) print(level string, msg string, args ...interface{}) {
	timestamp := time.Now().Format("15:04:05")

	// Build context string with only essential information
	var contextParts []string

	// Add component name (shortened)
	if l.component != "" {
		shortComponent := l.shortenComponent(l.component)
		contextParts = append(contextParts, shortComponent)
	}

	// Add only critical context values
	for k, v := range l.context {
		if k == "org_name" {
			contextParts = append(contextParts, fmt.Sprintf("%v", v))
		}
	}

	// Add args based on log level and importance
	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			if key, ok := args[i].(string); ok {
				shouldShow := false
				if level == SimpleLevelDebug {
					// DEBUG shows all args except those explicitly marked as noise
					shouldShow = !l.isDebugArg(key) || l.isImportantArg(key)
				} else {
					// Other levels show only important args
					shouldShow = l.isImportantArg(key)
				}

				if shouldShow {
					contextParts = append(contextParts, fmt.Sprintf("%s=%v", key, args[i+1]))
				}
			}
		}
	}

	contextStr := ""
	if len(contextParts) > 0 {
		contextStr = fmt.Sprintf(" [%s]", strings.Join(contextParts, " "))
	}

	// Format: TIME LEVEL[CONTEXT] MESSAGE
	output := fmt.Sprintf("%s %s%s %s", timestamp, level, contextStr, msg)
	fmt.Println(output)
}

// shortenComponent shortens component names for better readability.
func (l *SimpleLogger) shortenComponent(component string) string {
	switch component {
	case "bulk-clone-github":
		return "github"
	case "bulk-clone-gitlab":
		return "gitlab"
	case "bulk-clone-gitea":
		return "gitea"
	case "doctor":
		return "doctor"
	default:
		return component
	}
}

// isImportantArg determines if an argument should be shown in logs.
func (l *SimpleLogger) isImportantArg(key string) bool {
	importantArgs := map[string]bool{
		"attempt":      true,
		"max_retries":  true,
		"error_count":  true,
		"progress":     true,
		"repos_count":  true,
		"success_rate": true,
		"duration":     true,
		"repo_name":    true,
		"has_token":    true,
	}
	return importantArgs[key]
}

// shouldLog determines if a message should be logged based on configuration.
func (l *SimpleLogger) shouldLog(level string) bool {
	// Global flags override all config settings
	if globalQuiet {
		return level == SimpleLevelError
	}

	if globalDebug {
		return true // Show all levels
	}

	if globalVerbose {
		return level != SimpleLevelDebug // Show all except debug
	}

	// Fall back to config-based logic
	if l.config == nil {
		return level == SimpleLevelError || level == SimpleLevelWarn // Default: only errors and warnings
	}

	// If CLI logging is disabled, don't log anything except critical errors
	if !l.config.Enabled {
		return level == SimpleLevelError
	}

	// If quiet mode is enabled, only show critical errors
	if l.config.Quiet {
		return level == SimpleLevelError
	}

	// If only errors mode, show only errors and warnings
	if l.config.OnlyErrors {
		return level == SimpleLevelError || level == SimpleLevelWarn
	}

	// Check level hierarchy: DEBUG < INFO < WARN < ERROR
	configLevel := strings.ToUpper(l.config.Level)
	switch configLevel {
	case SimpleLevelDebug:
		return true // Show all levels
	case SimpleLevelInfo:
		return level != SimpleLevelDebug
	case SimpleLevelWarn:
		return level == SimpleLevelWarn || level == SimpleLevelError
	case SimpleLevelError:
		return level == SimpleLevelError
	default:
		return level == SimpleLevelError || level == SimpleLevelWarn // Default to errors and warnings
	}
}

// isDebugArg determines if an argument should be shown only in DEBUG level.
func (l *SimpleLogger) isDebugArg(key string) bool {
	debugArgs := map[string]bool{
		"optimized":     true,
		"streaming":     true,
		"enable_cache":  true,
		"resume":        true,
		"progress_mode": true,
	}
	return debugArgs[key]
}

// LoggerMiddleware provides logging middleware functionality.
func (l *SimpleLogger) LoggerMiddleware(next func() error) error {
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

// generateSimpleSessionID generates a simple session ID.
func generateSimpleSessionID(component string) string {
	return fmt.Sprintf("%s_%d", component, time.Now().Unix())
}

// Global simple logger instance.
var globalSimpleLogger *SimpleLogger

// SetGlobalSimpleLogger sets a global simple logger instance.
func SetGlobalSimpleLogger(logger *SimpleLogger) {
	globalSimpleLogger = logger
}

// GetGlobalSimpleLogger returns the global simple logger instance.
func GetGlobalSimpleLogger() *SimpleLogger {
	if globalSimpleLogger == nil {
		globalSimpleLogger = NewSimpleLogger("global")
	}
	return globalSimpleLogger
}

// SimpleDebug logs a debug message using the global logger.
func SimpleDebug(msg string, args ...interface{}) {
	GetGlobalSimpleLogger().Debug(msg, args...)
}

// SimpleInfo logs an info message using the global logger.
func SimpleInfo(msg string, args ...interface{}) {
	GetGlobalSimpleLogger().Info(msg, args...)
}

// SimpleWarn logs a warning message using the global logger.
func SimpleWarn(msg string, args ...interface{}) {
	GetGlobalSimpleLogger().Warn(msg, args...)
}

// SimpleError logs an error message using the global logger.
func SimpleError(msg string, args ...interface{}) {
	GetGlobalSimpleLogger().Error(msg, args...)
}

// SimpleErrorWithStack logs an error with stack trace using the global logger.
func SimpleErrorWithStack(err error, msg string, args ...interface{}) {
	GetGlobalSimpleLogger().ErrorWithStack(err, msg, args...)
}

// SetGlobalLoggingFlags sets global logging flags that override config settings.
func SetGlobalLoggingFlags(verbose, debug, quiet bool) {
	globalVerbose = verbose
	globalDebug = debug
	globalQuiet = quiet
}
