package debug

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// LogLevel represents logging levels
type LogLevel int

const (
	LevelSilent LogLevel = iota
	LevelError
	LevelWarn
	LevelInfo
	LevelDebug
	LevelTrace
)

var levelNames = map[LogLevel]string{
	LevelSilent: "SILENT",
	LevelError:  "ERROR",
	LevelWarn:   "WARN",
	LevelInfo:   "INFO",
	LevelDebug:  "DEBUG",
	LevelTrace:  "TRACE",
}

var levelColors = map[LogLevel]string{
	LevelError: "\033[31m", // Red
	LevelWarn:  "\033[33m", // Yellow
	LevelInfo:  "\033[36m", // Cyan
	LevelDebug: "\033[32m", // Green
	LevelTrace: "\033[35m", // Magenta
}

const colorReset = "\033[0m"

// LoggerConfig holds logger configuration
type LoggerConfig struct {
	Level       LogLevel `json:"level"`
	File        string   `json:"file"`
	EnableColor bool     `json:"enable_color"`
	EnableTrace bool     `json:"enable_trace"`
	MaxFileSize int64    `json:"max_file_size"`
	MaxBackups  int      `json:"max_backups"`
	Compress    bool     `json:"compress"`
	Format      string   `json:"format"` // "text" or "json"
}

// DefaultLoggerConfig returns a default logger configuration
func DefaultLoggerConfig() *LoggerConfig {
	return &LoggerConfig{
		Level:       LevelInfo,
		File:        "",
		EnableColor: true,
		EnableTrace: false,
		MaxFileSize: 100 * 1024 * 1024, // 100MB
		MaxBackups:  5,
		Compress:    true,
		Format:      "text",
	}
}

// Logger provides enhanced logging capabilities
type Logger struct {
	config *LoggerConfig
	mu     sync.RWMutex
	file   *os.File
	writer io.Writer
}

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	File      string                 `json:"file,omitempty"`
	Line      int                    `json:"line,omitempty"`
	Function  string                 `json:"function,omitempty"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// NewLogger creates a new logger instance
func NewLogger(config *LoggerConfig) (*Logger, error) {
	if config == nil {
		config = DefaultLoggerConfig()
	}

	logger := &Logger{
		config: config,
		writer: os.Stderr,
	}

	// Setup file output if specified
	if config.File != "" {
		if err := logger.setupFileOutput(); err != nil {
			return nil, fmt.Errorf("failed to setup file output: %w", err)
		}
	}

	return logger, nil
}

// SetLevel sets the logging level
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.config.Level = level
}

// GetLevel returns the current logging level
func (l *Logger) GetLevel() LogLevel {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.config.Level
}

// SetFormat sets the logging format
func (l *Logger) SetFormat(format string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.config.Format = format
}

// Error logs an error message
func (l *Logger) Error(msg string, fields ...map[string]interface{}) {
	l.log(LevelError, msg, fields...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, fields ...map[string]interface{}) {
	l.log(LevelWarn, msg, fields...)
}

// Info logs an info message
func (l *Logger) Info(msg string, fields ...map[string]interface{}) {
	l.log(LevelInfo, msg, fields...)
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, fields ...map[string]interface{}) {
	l.log(LevelDebug, msg, fields...)
}

// Trace logs a trace message
func (l *Logger) Trace(msg string, fields ...map[string]interface{}) {
	l.log(LevelTrace, msg, fields...)
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.log(LevelError, fmt.Sprintf(format, args...))
}

// Warnf logs a formatted warning message
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.log(LevelWarn, fmt.Sprintf(format, args...))
}

// Infof logs a formatted info message
func (l *Logger) Infof(format string, args ...interface{}) {
	l.log(LevelInfo, fmt.Sprintf(format, args...))
}

// Debugf logs a formatted debug message
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.log(LevelDebug, fmt.Sprintf(format, args...))
}

// Tracef logs a formatted trace message
func (l *Logger) Tracef(format string, args ...interface{}) {
	l.log(LevelTrace, fmt.Sprintf(format, args...))
}

// WithFields creates a new logger with additional fields
func (l *Logger) WithFields(fields map[string]interface{}) *FieldLogger {
	return &FieldLogger{
		logger: l,
		fields: fields,
	}
}

// log writes a log entry
func (l *Logger) log(level LogLevel, msg string, fields ...map[string]interface{}) {
	l.mu.RLock()
	currentLevel := l.config.Level
	format := l.config.Format
	enableTrace := l.config.EnableTrace
	writer := l.writer
	l.mu.RUnlock()

	if level > currentLevel {
		return
	}

	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     levelNames[level],
		Message:   msg,
	}

	// Add fields
	if len(fields) > 0 {
		entry.Fields = make(map[string]interface{})
		for _, fieldMap := range fields {
			for k, v := range fieldMap {
				entry.Fields[k] = v
			}
		}
	}

	// Add trace information
	if enableTrace || level <= LevelDebug {
		if pc, file, line, ok := runtime.Caller(2); ok {
			entry.File = filepath.Base(file)
			entry.Line = line
			if fn := runtime.FuncForPC(pc); fn != nil {
				entry.Function = fn.Name()
			}
		}
	}

	// Format and write
	var output string
	if format == "json" {
		output = l.formatJSON(entry)
	} else {
		output = l.formatText(entry, level)
	}

	fmt.Fprint(writer, output)
}

// formatText formats log entry as text
func (l *Logger) formatText(entry LogEntry, level LogLevel) string {
	timestamp := entry.Timestamp.Format("2006-01-02 15:04:05.000")
	levelStr := entry.Level

	// Add color if enabled
	if l.config.EnableColor {
		if color, ok := levelColors[level]; ok {
			levelStr = color + levelStr + colorReset
		}
	}

	// Build message
	parts := []string{
		fmt.Sprintf("[%s]", timestamp),
		fmt.Sprintf("%-5s", levelStr),
		entry.Message,
	}

	// Add trace info
	if entry.File != "" {
		traceInfo := fmt.Sprintf("(%s:%d)", entry.File, entry.Line)
		if entry.Function != "" {
			funcName := filepath.Base(entry.Function)
			traceInfo = fmt.Sprintf("%s in %s", traceInfo, funcName)
		}
		parts = append(parts, traceInfo)
	}

	message := strings.Join(parts, " ")

	// Add fields
	if entry.Fields != nil && len(entry.Fields) > 0 {
		fieldStrs := make([]string, 0, len(entry.Fields))
		for k, v := range entry.Fields {
			fieldStrs = append(fieldStrs, fmt.Sprintf("%s=%v", k, v))
		}
		message += " " + strings.Join(fieldStrs, " ")
	}

	return message + "\n"
}

// formatJSON formats log entry as JSON
func (l *Logger) formatJSON(entry LogEntry) string {
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Sprintf(`{"error":"failed to marshal log entry: %v"}%s`, err, "\n")
	}
	return string(data) + "\n"
}

// setupFileOutput sets up file output
func (l *Logger) setupFileOutput() error {
	dir := filepath.Dir(l.config.File)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	file, err := os.OpenFile(l.config.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}

	l.file = file
	l.writer = file
	return nil
}

// Close closes the logger and any open files
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// FieldLogger is a logger with pre-set fields
type FieldLogger struct {
	logger *Logger
	fields map[string]interface{}
}

// Error logs an error message with fields
func (fl *FieldLogger) Error(msg string, additionalFields ...map[string]interface{}) {
	fields := fl.mergeFields(additionalFields...)
	fl.logger.log(LevelError, msg, fields)
}

// Warn logs a warning message with fields
func (fl *FieldLogger) Warn(msg string, additionalFields ...map[string]interface{}) {
	fields := fl.mergeFields(additionalFields...)
	fl.logger.log(LevelWarn, msg, fields)
}

// Info logs an info message with fields
func (fl *FieldLogger) Info(msg string, additionalFields ...map[string]interface{}) {
	fields := fl.mergeFields(additionalFields...)
	fl.logger.log(LevelInfo, msg, fields)
}

// Debug logs a debug message with fields
func (fl *FieldLogger) Debug(msg string, additionalFields ...map[string]interface{}) {
	fields := fl.mergeFields(additionalFields...)
	fl.logger.log(LevelDebug, msg, fields)
}

// Trace logs a trace message with fields
func (fl *FieldLogger) Trace(msg string, additionalFields ...map[string]interface{}) {
	fields := fl.mergeFields(additionalFields...)
	fl.logger.log(LevelTrace, msg, fields)
}

// mergeFields merges the field logger's fields with additional fields
func (fl *FieldLogger) mergeFields(additionalFields ...map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})

	// Copy base fields
	for k, v := range fl.fields {
		merged[k] = v
	}

	// Add additional fields
	for _, fieldMap := range additionalFields {
		for k, v := range fieldMap {
			merged[k] = v
		}
	}

	return merged
}

// ParseLogLevel parses a string into LogLevel
func ParseLogLevel(level string) (LogLevel, error) {
	switch strings.ToUpper(level) {
	case "SILENT":
		return LevelSilent, nil
	case "ERROR":
		return LevelError, nil
	case "WARN", "WARNING":
		return LevelWarn, nil
	case "INFO":
		return LevelInfo, nil
	case "DEBUG":
		return LevelDebug, nil
	case "TRACE":
		return LevelTrace, nil
	default:
		return LevelInfo, fmt.Errorf("unknown log level: %s", level)
	}
}

// Global logger instance
var (
	globalLogger   *Logger
	globalLoggerMu sync.RWMutex
)

// InitGlobalLogger initializes the global logger
func InitGlobalLogger(config *LoggerConfig) error {
	logger, err := NewLogger(config)
	if err != nil {
		return err
	}

	globalLoggerMu.Lock()
	if globalLogger != nil {
		globalLogger.Close()
	}
	globalLogger = logger
	globalLoggerMu.Unlock()

	return nil
}

// GetGlobalLogger returns the global logger
func GetGlobalLogger() *Logger {
	globalLoggerMu.RLock()
	defer globalLoggerMu.RUnlock()
	return globalLogger
}

// Global logging functions

// Error logs an error using the global logger
func Error(msg string, fields ...map[string]interface{}) {
	if logger := GetGlobalLogger(); logger != nil {
		logger.Error(msg, fields...)
	} else {
		log.Printf("ERROR: %s", msg)
	}
}

// Warn logs a warning using the global logger
func Warn(msg string, fields ...map[string]interface{}) {
	if logger := GetGlobalLogger(); logger != nil {
		logger.Warn(msg, fields...)
	} else {
		log.Printf("WARN: %s", msg)
	}
}

// Info logs an info using the global logger
func Info(msg string, fields ...map[string]interface{}) {
	if logger := GetGlobalLogger(); logger != nil {
		logger.Info(msg, fields...)
	} else {
		log.Printf("INFO: %s", msg)
	}
}

// Debug logs a debug using the global logger
func Debug(msg string, fields ...map[string]interface{}) {
	if logger := GetGlobalLogger(); logger != nil {
		logger.Debug(msg, fields...)
	} else {
		log.Printf("DEBUG: %s", msg)
	}
}

// Trace logs a trace using the global logger
func Trace(msg string, fields ...map[string]interface{}) {
	if logger := GetGlobalLogger(); logger != nil {
		logger.Trace(msg, fields...)
	} else {
		log.Printf("TRACE: %s", msg)
	}
}
