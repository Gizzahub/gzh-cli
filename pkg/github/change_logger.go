package github

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/gizzahub/gzh-cli/internal/env"
)

// ChangeLogger provides comprehensive logging for repository configuration changes.
type ChangeLogger struct {
	changelog *ChangeLog
	options   *LoggerOptions
}

// LoggerOptions configures the change logger behavior.
type LoggerOptions struct {
	// LogDirectory specifies where log files are stored
	LogDirectory string
	// LogFormat specifies the log output format (json, text, csv)
	LogFormat LogFormat
	// LogLevel controls which events are logged
	LogLevel LogLevel
	// MaxLogFileSize specifies maximum size before rotation (in bytes)
	MaxLogFileSize int64
	// MaxLogFiles specifies how many rotated files to keep
	MaxLogFiles int
	// EnableConsoleOutput enables logging to stdout/stderr
	EnableConsoleOutput bool
	// EnableStructuredOutput enables structured JSON output
	EnableStructuredOutput bool
}

// LogFormat represents the output format for logs.
type LogFormat string

const (
	LogFormatJSON LogFormat = "json"
	LogFormatText LogFormat = "text"
	LogFormatCSV  LogFormat = "csv"
)

// LogLevel represents the logging level.
type LogLevel string

const (
	LogLevelTrace LogLevel = "trace"
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

// logEntry represents a structured log entry.
type logEntry struct {
	Timestamp  time.Time              `json:"timestamp"`
	Level      LogLevel               `json:"level"`
	Message    string                 `json:"message"`
	ChangeID   string                 `json:"changeId,omitempty"`
	Operation  string                 `json:"operation"`
	Category   string                 `json:"category"`
	Repository string                 `json:"repository,omitempty"`
	User       string                 `json:"user"`
	Source     string                 `json:"source"`
	RequestID  string                 `json:"requestId,omitempty"`
	Context    map[string]interface{} `json:"context,omitempty"`
	Error      string                 `json:"error,omitempty"`
	Duration   *time.Duration         `json:"duration,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// operationContext provides context for logging operations.
type operationContext struct {
	RequestID    string
	User         string
	Source       string
	Organization string
	Repository   string
	StartTime    time.Time
	Metadata     map[string]interface{}
}

// NewChangeLogger creates a new change logger with the specified options.
func NewChangeLogger(changelog *ChangeLog, options *LoggerOptions) *ChangeLogger {
	if options == nil {
		options = DefaultLoggerOptions()
	}

	// Ensure log directory exists
	if options.LogDirectory != "" {
		if err := os.MkdirAll(options.LogDirectory, 0o750); err != nil {
			// Log directory creation failure is not critical, use temp dir
			options.LogDirectory = os.TempDir()
		}
	}

	return &ChangeLogger{
		changelog: changelog,
		options:   options,
	}
}

// DefaultLoggerOptions returns default logger configuration.
func DefaultLoggerOptions() *LoggerOptions {
	homeDir, _ := os.UserHomeDir()

	return &LoggerOptions{
		LogDirectory:           filepath.Join(homeDir, ".config", "gzh-manager", "logs"),
		LogFormat:              LogFormatJSON,
		LogLevel:               LogLevelInfo,
		MaxLogFileSize:         10 * 1024 * 1024, // 10MB
		MaxLogFiles:            5,
		EnableConsoleOutput:    true,
		EnableStructuredOutput: true,
	}
}

// LogRepositoryChange logs a repository configuration change with full context.
func (cl *ChangeLogger) LogRepositoryChange(ctx context.Context, opCtx *operationContext, changeRecord *ChangeRecord, level LogLevel, message string, err error) error {
	entry := &logEntry{
		Timestamp:  time.Now(),
		Level:      level,
		Message:    message,
		ChangeID:   changeRecord.ID,
		Operation:  changeRecord.Operation,
		Category:   changeRecord.Category,
		Repository: changeRecord.Repository,
		User:       opCtx.User,
		Source:     opCtx.Source,
		RequestID:  opCtx.RequestID,
		Context: map[string]interface{}{
			"organization": changeRecord.Organization,
			"before":       changeRecord.Before,
			"after":        changeRecord.After,
		},
		Metadata: opCtx.Metadata,
	}

	if err != nil {
		entry.Error = err.Error()
	}

	if !opCtx.StartTime.IsZero() {
		duration := time.Since(opCtx.StartTime)
		entry.Duration = &duration
	}

	return cl.writeLogEntry(entry)
}

// LogOperation logs a general operation with context.
func (cl *ChangeLogger) LogOperation(ctx context.Context, opCtx *operationContext, level LogLevel, operation, category, message string, err error) error {
	entry := &logEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Operation: operation,
		Category:  category,
		User:      opCtx.User,
		Source:    opCtx.Source,
		RequestID: opCtx.RequestID,
		Context: map[string]interface{}{
			"organization": opCtx.Organization,
			"repository":   opCtx.Repository,
		},
		Metadata: opCtx.Metadata,
	}

	if err != nil {
		entry.Error = err.Error()
	}

	if !opCtx.StartTime.IsZero() {
		duration := time.Since(opCtx.StartTime)
		entry.Duration = &duration
	}

	return cl.writeLogEntry(entry)
}

// LogBulkOperation logs bulk operations with aggregated statistics.
func (cl *ChangeLogger) LogBulkOperation(ctx context.Context, opCtx *operationContext, level LogLevel, operation string, stats *bulkOperationStats, err error) error {
	message := fmt.Sprintf("Bulk %s completed: %d total, %d success, %d failed, %d skipped",
		operation, stats.Total, stats.Success, stats.Failed, stats.Skipped)

	entry := &logEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Operation: fmt.Sprintf("bulk_%s", operation),
		Category:  "bulk_operation",
		User:      opCtx.User,
		Source:    opCtx.Source,
		RequestID: opCtx.RequestID,
		Context: map[string]interface{}{
			"organization": opCtx.Organization,
			"statistics":   stats,
		},
		Metadata: opCtx.Metadata,
	}

	if err != nil {
		entry.Error = err.Error()
	}

	if !opCtx.StartTime.IsZero() {
		duration := time.Since(opCtx.StartTime)
		entry.Duration = &duration
	}

	return cl.writeLogEntry(entry)
}

// bulkOperationStats contains statistics for bulk operations.
type bulkOperationStats struct {
	Total    int                    `json:"total"`
	Success  int                    `json:"success"`
	Failed   int                    `json:"failed"`
	Skipped  int                    `json:"skipped"`
	Errors   map[string]string      `json:"errors,omitempty"`
	Duration time.Duration          `json:"duration"`
	Details  map[string]interface{} `json:"details,omitempty"`
}

// CreateOperationContext creates a new operation context for logging.
func (cl *ChangeLogger) CreateOperationContext(requestID, operation string) *operationContext {
	user := getSystemUser()

	return &operationContext{
		RequestID: requestID,
		User:      user,
		Source:    "cli",
		StartTime: time.Now(),
		Metadata:  make(map[string]interface{}),
	}
}

// writeLogEntry writes a log entry to the configured outputs.
func (cl *ChangeLogger) writeLogEntry(entry *logEntry) error {
	// Check if we should log this level
	if !cl.shouldLog(entry.Level) {
		return nil
	}

	// Write to console if enabled
	if cl.options.EnableConsoleOutput {
		if err := cl.writeToConsole(entry); err != nil {
			return fmt.Errorf("failed to write to console: %w", err)
		}
	}

	// Write to log file if directory is configured
	if cl.options.LogDirectory != "" {
		if err := cl.writeToFile(entry); err != nil {
			return fmt.Errorf("failed to write to file: %w", err)
		}
	}

	return nil
}

// shouldLog determines if a log entry should be written based on level.
func (cl *ChangeLogger) shouldLog(level LogLevel) bool {
	levels := map[LogLevel]int{
		LogLevelTrace: 0,
		LogLevelDebug: 1,
		LogLevelInfo:  2,
		LogLevelWarn:  3,
		LogLevelError: 4,
	}

	entryLevel, exists := levels[level]
	if !exists {
		return true // Log unknown levels
	}

	configLevel, exists := levels[cl.options.LogLevel]
	if !exists {
		return true // Log if config level is unknown
	}

	return entryLevel >= configLevel
}

// writeToConsole writes log entry to console.
func (cl *ChangeLogger) writeToConsole(entry *logEntry) error {
	var output string

	switch cl.options.LogFormat {
	case LogFormatJSON:
		jsonBytes, err := json.Marshal(entry)
		if err != nil {
			return err
		}

		output = string(jsonBytes)

	case LogFormatText:
		output = cl.formatTextEntry(entry)

	case LogFormatCSV:
		output = cl.formatCSVEntry(entry)

	default:
		output = cl.formatTextEntry(entry)
	}

	// Write to stderr for errors and warnings, stdout for others
	if entry.Level == LogLevelError || entry.Level == LogLevelWarn {
		fmt.Fprintln(os.Stderr, output)
	} else {
		fmt.Println(output)
	}

	return nil
}

// writeToFile writes log entry to file with rotation.
func (cl *ChangeLogger) writeToFile(entry *logEntry) error {
	logFile := cl.getLogFileName()

	// Check if rotation is needed
	if err := cl.rotateLogIfNeeded(logFile); err != nil {
		return err
	}

	// Open or create log file
	file, err := os.OpenFile(filepath.Clean(logFile), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			// File close errors in logging are typically not critical
			// but we should at least log to stderr if possible
			fmt.Fprintf(os.Stderr, "Warning: failed to close log file: %v\n", err)
		}
	}()

	// Format entry based on configured format
	var line string

	switch cl.options.LogFormat {
	case LogFormatJSON:
		jsonBytes, err := json.Marshal(entry)
		if err != nil {
			return err
		}

		line = string(jsonBytes) + "\n"

	case LogFormatText:
		line = cl.formatTextEntry(entry) + "\n"

	case LogFormatCSV:
		line = cl.formatCSVEntry(entry) + "\n"

	default:
		line = cl.formatTextEntry(entry) + "\n"
	}

	_, err = file.WriteString(line)

	return err
}

// getLogFileName returns the current log file name.
func (cl *ChangeLogger) getLogFileName() string {
	date := time.Now().Format("2006-01-02")
	return filepath.Join(cl.options.LogDirectory, fmt.Sprintf("gzh-manager-%s.log", date))
}

// rotateLogIfNeeded checks if log rotation is needed and performs it.
func (cl *ChangeLogger) rotateLogIfNeeded(logFile string) error {
	info, err := os.Stat(logFile)
	if os.IsNotExist(err) {
		return nil // File doesn't exist, no rotation needed
	}

	if err != nil {
		return err
	}

	if info.Size() < cl.options.MaxLogFileSize {
		return nil // File is not large enough for rotation
	}

	// Rotate the log file
	timestamp := time.Now().Format("20060102-150405")
	rotatedFile := fmt.Sprintf("%s.%s", logFile, timestamp)

	if err := os.Rename(logFile, rotatedFile); err != nil {
		return err
	}

	// Clean up old rotated files
	return cl.cleanupOldLogFiles()
}

// cleanupOldLogFiles removes old rotated log files.
func (cl *ChangeLogger) cleanupOldLogFiles() error {
	pattern := filepath.Join(cl.options.LogDirectory, "gzh-manager-*.log.*")

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	if len(matches) <= cl.options.MaxLogFiles {
		return nil
	}

	// Sort files by modification time and remove oldest
	// This is a simplified implementation
	for i := 0; i < len(matches)-cl.options.MaxLogFiles; i++ {
		if err := os.Remove(matches[i]); err != nil {
			// Log cleanup failure is not critical, but log the error
			fmt.Fprintf(os.Stderr, "Warning: failed to remove old log file %s: %v\n", matches[i], err)
		}
	}

	return nil
}

// formatTextEntry formats a log entry as human-readable text.
func (cl *ChangeLogger) formatTextEntry(entry *logEntry) string {
	timestamp := entry.Timestamp.Format("2006-01-02 15:04:05")
	parts := []string{
		fmt.Sprintf("[%s]", strings.ToUpper(string(entry.Level))),
		timestamp,
		entry.Message,
	}

	if entry.Repository != "" {
		parts = append(parts, fmt.Sprintf("repo=%s", entry.Repository))
	}

	if entry.User != "" {
		parts = append(parts, fmt.Sprintf("user=%s", entry.User))
	}

	if entry.Operation != "" {
		parts = append(parts, fmt.Sprintf("op=%s", entry.Operation))
	}

	if entry.Duration != nil {
		parts = append(parts, fmt.Sprintf("duration=%s", entry.Duration.String()))
	}

	if entry.Error != "" {
		parts = append(parts, fmt.Sprintf("error=%s", entry.Error))
	}

	return strings.Join(parts, " ")
}

// formatCSVEntry formats a log entry as CSV.
func (cl *ChangeLogger) formatCSVEntry(entry *logEntry) string {
	timestamp := entry.Timestamp.Format("2006-01-02 15:04:05")
	fields := []string{
		timestamp,
		string(entry.Level),
		entry.Operation,
		entry.Category,
		entry.Repository,
		entry.User,
		entry.Message,
	}

	if entry.Error != "" {
		fields = append(fields, entry.Error)
	} else {
		fields = append(fields, "")
	}

	// Simple CSV formatting - would need proper escaping for production
	return strings.Join(fields, ",")
}

// getSystemUser gets the current system user for logging (enhanced version).
func getSystemUser() string {
	return getSystemUserWithEnv(env.NewOSEnvironment())
}

// getSystemUserWithEnv gets the current system user using the provided environment.
func getSystemUserWithEnv(environment env.Environment) string {
	if u, err := user.Current(); err == nil {
		return u.Username
	}

	if username := environment.Get(env.CommonEnvironmentKeys.User); username != "" {
		return username
	}

	if username := environment.Get(env.CommonEnvironmentKeys.Username); username != "" {
		return username
	}

	return "unknown"
}

// GetLogSummary returns a summary of recent log entries.
func (cl *ChangeLogger) GetLogSummary(ctx context.Context, since time.Time) (*LogSummary, error) {
	filter := ChangeFilter{
		Since: since,
		Limit: 1000, // Reasonable limit for summary
	}

	changes, err := cl.changelog.ListChanges(ctx, filter)
	if err != nil {
		return nil, err
	}

	summary := &LogSummary{
		Period:       fmt.Sprintf("Since %s", since.Format("2006-01-02 15:04:05")),
		TotalChanges: len(changes),
		ByCategory:   make(map[string]int),
		ByOperation:  make(map[string]int),
		ByUser:       make(map[string]int),
		Errors:       make([]string, 0),
	}

	for _, change := range changes {
		summary.ByCategory[change.Category]++
		summary.ByOperation[change.Operation]++
		summary.ByUser[change.User]++
	}

	return summary, nil
}

// LogSummary provides a summary of logging activity.
type LogSummary struct {
	Period       string         `json:"period"`
	TotalChanges int            `json:"totalChanges"`
	ByCategory   map[string]int `json:"byCategory"`
	ByOperation  map[string]int `json:"byOperation"`
	ByUser       map[string]int `json:"byUser"`
	Errors       []string       `json:"errors"`
}
