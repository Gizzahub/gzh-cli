// Package debug provides structured logging capabilities with RFC 5424 compliance
package debug

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/otel/trace"
)

// RFC5424Severity represents syslog severity levels
type RFC5424Severity int

const (
	SeverityEmergency RFC5424Severity = iota // 0: Emergency: system is unusable
	SeverityAlert                            // 1: Alert: action must be taken immediately
	SeverityCritical                         // 2: Critical: critical conditions
	SeverityError                            // 3: Error: error conditions
	SeverityWarning                          // 4: Warning: warning conditions
	SeverityNotice                           // 5: Notice: normal but significant condition
	SeverityInfo                             // 6: Informational: informational messages
	SeverityDebug                            // 7: Debug: debug-level messages
)

var rfc5424Names = map[RFC5424Severity]string{
	SeverityEmergency: "emergency",
	SeverityAlert:     "alert",
	SeverityCritical:  "critical",
	SeverityError:     "error",
	SeverityWarning:   "warning",
	SeverityNotice:    "notice",
	SeverityInfo:      "info",
	SeverityDebug:     "debug",
}

// StructuredLogEntry represents a standardized log entry following RFC 5424
type StructuredLogEntry struct {
	// RFC 5424 Required Fields
	Timestamp time.Time       `json:"@timestamp"`      // ISO 8601 timestamp
	Version   int             `json:"@version"`        // Log format version
	Level     string          `json:"level"`           // Log level
	Severity  RFC5424Severity `json:"severity"`        // Numeric severity
	Hostname  string          `json:"hostname"`        // Hostname
	AppName   string          `json:"appname"`         // Application name
	ProcID    string          `json:"procid"`          // Process ID
	MsgID     string          `json:"msgid,omitempty"` // Message ID
	Message   string          `json:"message"`         // Log message

	// Distributed Tracing Fields
	TraceID string `json:"trace_id,omitempty"` // Distributed trace ID
	SpanID  string `json:"span_id,omitempty"`  // Span ID

	// Source Code Fields
	Caller struct {
		File     string `json:"file,omitempty"`     // Source file
		Line     int    `json:"line,omitempty"`     // Line number
		Function string `json:"function,omitempty"` // Function name
	} `json:"caller,omitempty"`

	// Structured Data
	Fields map[string]interface{} `json:"fields,omitempty"` // Additional structured data

	// Performance Fields
	Duration  *time.Duration `json:"duration,omitempty"`   // Operation duration
	Latency   *time.Duration `json:"latency,omitempty"`    // Request latency
	BytesRead *int64         `json:"bytes_read,omitempty"` // Bytes read
	BytesOut  *int64         `json:"bytes_out,omitempty"`  // Bytes written
}

// StructuredLoggerConfig holds structured logger configuration
type StructuredLoggerConfig struct {
	// Basic Configuration
	Level       RFC5424Severity `json:"level"`
	Format      string          `json:"format"` // "json", "logfmt", "console"
	Output      string          `json:"output"` // "stdout", "stderr", or file path
	AppName     string          `json:"app_name"`
	Version     string          `json:"version"`
	Environment string          `json:"environment"` // "development", "staging", "production"

	// Trace Configuration
	EnableTracing bool `json:"enable_tracing"`
	EnableCaller  bool `json:"enable_caller"`
	CallerSkip    int  `json:"caller_skip"`

	// Sampling Configuration
	EnableSampling  bool    `json:"enable_sampling"`
	SampleRate      float64 `json:"sample_rate"`      // 0.0 to 1.0
	SampleThreshold int     `json:"sample_threshold"` // Log every Nth message

	// Performance Configuration
	AsyncLogging  bool          `json:"async_logging"`
	BufferSize    int           `json:"buffer_size"`
	FlushInterval time.Duration `json:"flush_interval"`
	MaxFileSize   int64         `json:"max_file_size"`
	MaxBackups    int           `json:"max_backups"`
	Compress      bool          `json:"compress"`

	// Filter Configuration
	ModuleLevels map[string]RFC5424Severity `json:"module_levels,omitempty"` // Per-module log levels
	IgnoreFields []string                   `json:"ignore_fields,omitempty"` // Fields to exclude
}

// DefaultStructuredLoggerConfig returns a default structured logger configuration
func DefaultStructuredLoggerConfig() *StructuredLoggerConfig {
	return &StructuredLoggerConfig{
		Level:           SeverityInfo,
		Format:          "json",
		Output:          "stderr",
		AppName:         "gzh-manager",
		Version:         "1.0.0",
		Environment:     "development",
		EnableTracing:   true,
		EnableCaller:    true,
		CallerSkip:      2,
		EnableSampling:  false,
		SampleRate:      1.0,
		SampleThreshold: 1,
		AsyncLogging:    false,
		BufferSize:      1000,
		FlushInterval:   time.Second,
		MaxFileSize:     100 * 1024 * 1024, // 100MB
		MaxBackups:      5,
		Compress:        true,
		ModuleLevels:    make(map[string]RFC5424Severity),
	}
}

// StructuredLogger provides RFC 5424 compliant structured logging
type StructuredLogger struct {
	config   *StructuredLoggerConfig
	writer   io.Writer
	file     *os.File
	hostname string
	mutex    sync.RWMutex

	// Async logging
	logChan chan *StructuredLogEntry
	done    chan struct{}
	wg      sync.WaitGroup

	// Sampling
	sampleCounter uint64
	sampleMutex   sync.Mutex
}

// NewStructuredLogger creates a new structured logger
func NewStructuredLogger(config *StructuredLoggerConfig) (*StructuredLogger, error) {
	if config == nil {
		config = DefaultStructuredLoggerConfig()
	}

	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}

	logger := &StructuredLogger{
		config:   config,
		hostname: hostname,
	}

	// Setup output writer
	if err := logger.setupOutput(); err != nil {
		return nil, fmt.Errorf("failed to setup output: %w", err)
	}

	// Setup async logging if enabled
	if config.AsyncLogging {
		logger.logChan = make(chan *StructuredLogEntry, config.BufferSize)
		logger.done = make(chan struct{})
		logger.startAsyncWorker()
	}

	return logger, nil
}

// Emergency logs an emergency message
func (sl *StructuredLogger) Emergency(ctx context.Context, msg string, fields ...map[string]interface{}) {
	sl.log(ctx, SeverityEmergency, msg, fields...)
}

// Alert logs an alert message
func (sl *StructuredLogger) Alert(ctx context.Context, msg string, fields ...map[string]interface{}) {
	sl.log(ctx, SeverityAlert, msg, fields...)
}

// Critical logs a critical message
func (sl *StructuredLogger) Critical(ctx context.Context, msg string, fields ...map[string]interface{}) {
	sl.log(ctx, SeverityCritical, msg, fields...)
}

// ErrorLevel logs an error message (renamed to avoid conflict)
func (sl *StructuredLogger) ErrorLevel(ctx context.Context, msg string, fields ...map[string]interface{}) {
	sl.log(ctx, SeverityError, msg, fields...)
}

// Warning logs a warning message
func (sl *StructuredLogger) Warning(ctx context.Context, msg string, fields ...map[string]interface{}) {
	sl.log(ctx, SeverityWarning, msg, fields...)
}

// Notice logs a notice message
func (sl *StructuredLogger) Notice(ctx context.Context, msg string, fields ...map[string]interface{}) {
	sl.log(ctx, SeverityNotice, msg, fields...)
}

// InfoLevel logs an info message (renamed to avoid conflict)
func (sl *StructuredLogger) InfoLevel(ctx context.Context, msg string, fields ...map[string]interface{}) {
	sl.log(ctx, SeverityInfo, msg, fields...)
}

// DebugLevel logs a debug message (renamed to avoid conflict)
func (sl *StructuredLogger) DebugLevel(ctx context.Context, msg string, fields ...map[string]interface{}) {
	sl.log(ctx, SeverityDebug, msg, fields...)
}

// WithModule returns a logger with module-specific configuration
func (sl *StructuredLogger) WithModule(module string) *ModuleStructuredLogger {
	return &ModuleStructuredLogger{
		logger: sl,
		module: module,
	}
}

// WithFields returns a logger with pre-set fields
func (sl *StructuredLogger) WithFields(fields map[string]interface{}) *FieldStructuredLogger {
	return &FieldStructuredLogger{
		logger: sl,
		fields: fields,
	}
}

// SetLevel sets the logging level dynamically
func (sl *StructuredLogger) SetLevel(level RFC5424Severity) {
	sl.mutex.Lock()
	defer sl.mutex.Unlock()
	sl.config.Level = level
}

// SetModuleLevel sets the logging level for a specific module
func (sl *StructuredLogger) SetModuleLevel(module string, level RFC5424Severity) {
	sl.mutex.Lock()
	defer sl.mutex.Unlock()
	sl.config.ModuleLevels[module] = level
}

// GetLevel returns the current logging level
func (sl *StructuredLogger) GetLevel() RFC5424Severity {
	sl.mutex.RLock()
	defer sl.mutex.RUnlock()
	return sl.config.Level
}

// log writes a structured log entry
func (sl *StructuredLogger) log(ctx context.Context, level RFC5424Severity, msg string, fields ...map[string]interface{}) {
	sl.mutex.RLock()
	currentLevel := sl.config.Level
	enableTracing := sl.config.EnableTracing
	enableCaller := sl.config.EnableCaller
	callerSkip := sl.config.CallerSkip
	enableSampling := sl.config.EnableSampling
	sampleRate := sl.config.SampleRate
	sampleThreshold := sl.config.SampleThreshold
	asyncLogging := sl.config.AsyncLogging
	sl.mutex.RUnlock()

	// Check level
	if level > currentLevel {
		return
	}

	// Apply sampling
	if enableSampling && !sl.shouldSample(level, sampleRate, sampleThreshold) {
		return
	}

	// Create log entry
	entry := &StructuredLogEntry{
		Timestamp: time.Now().UTC(),
		Version:   1,
		Level:     rfc5424Names[level],
		Severity:  level,
		Hostname:  sl.hostname,
		AppName:   sl.config.AppName,
		ProcID:    fmt.Sprintf("%d", os.Getpid()),
		Message:   msg,
	}

	// Add tracing information
	if enableTracing && ctx != nil {
		if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
			entry.TraceID = span.SpanContext().TraceID().String()
			entry.SpanID = span.SpanContext().SpanID().String()
		}
	}

	// Add caller information
	if enableCaller {
		if pc, file, line, ok := runtime.Caller(callerSkip); ok {
			entry.Caller.File = filepath.Base(file)
			entry.Caller.Line = line
			if fn := runtime.FuncForPC(pc); fn != nil {
				entry.Caller.Function = fn.Name()
			}
		}
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

	// Write log entry
	if asyncLogging {
		select {
		case sl.logChan <- entry:
		default:
			// Channel is full, write synchronously
			sl.writeEntry(entry)
		}
	} else {
		sl.writeEntry(entry)
	}
}

// shouldSample determines if a log entry should be sampled
func (sl *StructuredLogger) shouldSample(level RFC5424Severity, sampleRate float64, sampleThreshold int) bool {
	// Always log high severity messages
	if level <= SeverityError {
		return true
	}

	sl.sampleMutex.Lock()
	defer sl.sampleMutex.Unlock()

	sl.sampleCounter++

	// Threshold-based sampling
	if sampleThreshold > 1 && sl.sampleCounter%uint64(sampleThreshold) != 0 {
		return false
	}

	// Rate-based sampling (simplified)
	if sampleRate < 1.0 {
		return sl.sampleCounter%uint64(1.0/sampleRate) == 0
	}

	return true
}

// writeEntry writes a log entry to the output
func (sl *StructuredLogger) writeEntry(entry *StructuredLogEntry) {
	sl.mutex.RLock()
	format := sl.config.Format
	writer := sl.writer
	sl.mutex.RUnlock()

	var output string
	switch format {
	case "json":
		output = sl.formatJSON(entry)
	case "logfmt":
		output = sl.formatLogfmt(entry)
	case "console":
		output = sl.formatConsole(entry)
	default:
		output = sl.formatJSON(entry)
	}

	fmt.Fprint(writer, output)
}

// formatJSON formats log entry as JSON
func (sl *StructuredLogger) formatJSON(entry *StructuredLogEntry) string {
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Sprintf(`{"error":"failed to marshal log entry: %v","@timestamp":"%s"}%s`,
			err, time.Now().Format(time.RFC3339), "\n")
	}
	return string(data) + "\n"
}

// formatLogfmt formats log entry as logfmt
func (sl *StructuredLogger) formatLogfmt(entry *StructuredLogEntry) string {
	parts := []string{
		fmt.Sprintf(`@timestamp="%s"`, entry.Timestamp.Format(time.RFC3339)),
		fmt.Sprintf(`level="%s"`, entry.Level),
		fmt.Sprintf(`severity=%d`, entry.Severity),
		fmt.Sprintf(`hostname="%s"`, entry.Hostname),
		fmt.Sprintf(`appname="%s"`, entry.AppName),
		fmt.Sprintf(`procid="%s"`, entry.ProcID),
		fmt.Sprintf(`message="%s"`, strings.ReplaceAll(entry.Message, `"`, `\"`)),
	}

	if entry.TraceID != "" {
		parts = append(parts, fmt.Sprintf(`trace_id="%s"`, entry.TraceID))
	}
	if entry.SpanID != "" {
		parts = append(parts, fmt.Sprintf(`span_id="%s"`, entry.SpanID))
	}

	if entry.Caller.File != "" {
		parts = append(parts, fmt.Sprintf(`caller="%s:%d"`, entry.Caller.File, entry.Caller.Line))
	}

	if entry.Fields != nil {
		for k, v := range entry.Fields {
			parts = append(parts, fmt.Sprintf(`%s="%v"`, k, v))
		}
	}

	return strings.Join(parts, " ") + "\n"
}

// formatConsole formats log entry for console output
func (sl *StructuredLogger) formatConsole(entry *StructuredLogEntry) string {
	timestamp := entry.Timestamp.Format("15:04:05.000")
	level := strings.ToUpper(entry.Level)

	message := fmt.Sprintf("[%s] %-5s %s", timestamp, level, entry.Message)

	if entry.Caller.File != "" {
		message += fmt.Sprintf(" (%s:%d)", entry.Caller.File, entry.Caller.Line)
	}

	if entry.TraceID != "" {
		message += fmt.Sprintf(" trace=%s", entry.TraceID[:8])
	}

	if entry.Fields != nil && len(entry.Fields) > 0 {
		fieldStrs := make([]string, 0, len(entry.Fields))
		for k, v := range entry.Fields {
			fieldStrs = append(fieldStrs, fmt.Sprintf("%s=%v", k, v))
		}
		message += " " + strings.Join(fieldStrs, " ")
	}

	return message + "\n"
}

// setupOutput sets up the output writer
func (sl *StructuredLogger) setupOutput() error {
	switch sl.config.Output {
	case "stdout":
		sl.writer = os.Stdout
	case "stderr":
		sl.writer = os.Stderr
	default:
		// File output
		dir := filepath.Dir(sl.config.Output)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}

		file, err := os.OpenFile(sl.config.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			return err
		}

		sl.file = file
		sl.writer = file
	}
	return nil
}

// startAsyncWorker starts the async logging worker
func (sl *StructuredLogger) startAsyncWorker() {
	sl.wg.Add(1)
	go func() {
		defer sl.wg.Done()
		ticker := time.NewTicker(sl.config.FlushInterval)
		defer ticker.Stop()

		for {
			select {
			case entry := <-sl.logChan:
				sl.writeEntry(entry)
			case <-ticker.C:
				// Flush any buffered data
				if f, ok := sl.writer.(*os.File); ok {
					f.Sync()
				}
			case <-sl.done:
				// Drain remaining entries
				for {
					select {
					case entry := <-sl.logChan:
						sl.writeEntry(entry)
					default:
						return
					}
				}
			}
		}
	}()
}

// Close closes the logger and flushes any remaining logs
func (sl *StructuredLogger) Close() error {
	if sl.config.AsyncLogging {
		close(sl.done)
		sl.wg.Wait()
	}

	if sl.file != nil {
		return sl.file.Close()
	}
	return nil
}

// ModuleStructuredLogger provides module-specific logging
type ModuleStructuredLogger struct {
	logger *StructuredLogger
	module string
}

// ErrorLevel logs an error message with module context
func (ml *ModuleStructuredLogger) ErrorLevel(ctx context.Context, msg string, fields ...map[string]interface{}) {
	fields = ml.addModuleField(fields...)
	ml.logger.ErrorLevel(ctx, msg, fields...)
}

// Warning logs a warning message with module context
func (ml *ModuleStructuredLogger) Warning(ctx context.Context, msg string, fields ...map[string]interface{}) {
	fields = ml.addModuleField(fields...)
	ml.logger.Warning(ctx, msg, fields...)
}

// InfoLevel logs an info message with module context
func (ml *ModuleStructuredLogger) InfoLevel(ctx context.Context, msg string, fields ...map[string]interface{}) {
	fields = ml.addModuleField(fields...)
	ml.logger.InfoLevel(ctx, msg, fields...)
}

// DebugLevel logs a debug message with module context
func (ml *ModuleStructuredLogger) DebugLevel(ctx context.Context, msg string, fields ...map[string]interface{}) {
	fields = ml.addModuleField(fields...)
	ml.logger.DebugLevel(ctx, msg, fields...)
}

// addModuleField adds module information to fields
func (ml *ModuleStructuredLogger) addModuleField(fields ...map[string]interface{}) []map[string]interface{} {
	moduleField := map[string]interface{}{"module": ml.module}
	return append([]map[string]interface{}{moduleField}, fields...)
}

// FieldStructuredLogger provides structured logging with pre-set fields
type FieldStructuredLogger struct {
	logger *StructuredLogger
	fields map[string]interface{}
}

// ErrorLevel logs an error message with pre-set fields
func (fsl *FieldStructuredLogger) ErrorLevel(ctx context.Context, msg string, additionalFields ...map[string]interface{}) {
	fields := fsl.mergeFields(additionalFields...)
	fsl.logger.ErrorLevel(ctx, msg, fields)
}

// Warning logs a warning message with pre-set fields
func (fsl *FieldStructuredLogger) Warning(ctx context.Context, msg string, additionalFields ...map[string]interface{}) {
	fields := fsl.mergeFields(additionalFields...)
	fsl.logger.Warning(ctx, msg, fields)
}

// InfoLevel logs an info message with pre-set fields
func (fsl *FieldStructuredLogger) InfoLevel(ctx context.Context, msg string, additionalFields ...map[string]interface{}) {
	fields := fsl.mergeFields(additionalFields...)
	fsl.logger.InfoLevel(ctx, msg, fields)
}

// DebugLevel logs a debug message with pre-set fields
func (fsl *FieldStructuredLogger) DebugLevel(ctx context.Context, msg string, additionalFields ...map[string]interface{}) {
	fields := fsl.mergeFields(additionalFields...)
	fsl.logger.DebugLevel(ctx, msg, fields)
}

// mergeFields merges pre-set fields with additional fields
func (fsl *FieldStructuredLogger) mergeFields(additionalFields ...map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})

	// Copy base fields
	for k, v := range fsl.fields {
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

// ParseRFC5424Severity parses a string into RFC5424Severity
func ParseRFC5424Severity(level string) (RFC5424Severity, error) {
	switch strings.ToLower(level) {
	case "emergency", "emerg":
		return SeverityEmergency, nil
	case "alert":
		return SeverityAlert, nil
	case "critical", "crit":
		return SeverityCritical, nil
	case "error", "err":
		return SeverityError, nil
	case "warning", "warn":
		return SeverityWarning, nil
	case "notice":
		return SeverityNotice, nil
	case "info", "informational":
		return SeverityInfo, nil
	case "debug":
		return SeverityDebug, nil
	default:
		return SeverityInfo, fmt.Errorf("unknown log level: %s", level)
	}
}

// Global structured logger instance
var (
	globalStructuredLogger   *StructuredLogger
	globalStructuredLoggerMu sync.RWMutex
)

// InitGlobalStructuredLogger initializes the global structured logger
func InitGlobalStructuredLogger(config *StructuredLoggerConfig) error {
	logger, err := NewStructuredLogger(config)
	if err != nil {
		return err
	}

	globalStructuredLoggerMu.Lock()
	if globalStructuredLogger != nil {
		globalStructuredLogger.Close()
	}
	globalStructuredLogger = logger
	globalStructuredLoggerMu.Unlock()

	return nil
}

// GetGlobalStructuredLogger returns the global structured logger
func GetGlobalStructuredLogger() *StructuredLogger {
	globalStructuredLoggerMu.RLock()
	defer globalStructuredLoggerMu.RUnlock()
	return globalStructuredLogger
}
