package monitoring

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log/syslog"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

// FileOutput implements file-based log output with rotation
type FileOutput struct {
	name    string
	config  *OutputConfig
	writer  *lumberjack.Logger
	encoder LogEncoder
	mutex   sync.Mutex
}

// ConsoleOutput implements console-based log output
type ConsoleOutput struct {
	name    string
	config  *OutputConfig
	encoder LogEncoder
	writer  *bufio.Writer
	mutex   sync.Mutex
}

// SyslogOutput implements syslog-based log output
type SyslogOutput struct {
	name    string
	config  *OutputConfig
	writer  *syslog.Writer
	encoder LogEncoder
	mutex   sync.Mutex
}

// HTTPOutput implements HTTP-based log output for log aggregation services
type HTTPOutput struct {
	name       string
	config     *OutputConfig
	client     *http.Client
	endpoint   string
	headers    map[string]string
	buffer     []*LogEntry
	bufferSize int
	encoder    LogEncoder
	mutex      sync.Mutex
}

// LogEncoder interface for encoding log entries
type LogEncoder interface {
	Encode(entry *LogEntry) ([]byte, error)
}

// JSONEncoder encodes log entries as JSON
type JSONEncoder struct{}

// TextEncoder encodes log entries as plain text
type TextEncoder struct {
	template string
}

// StructuredEncoder encodes log entries with structured format
type StructuredEncoder struct {
	includeTimestamp bool
	includeSource    bool
}

// NewFileOutput creates a new file output
func NewFileOutput(name string, config *OutputConfig) (*FileOutput, error) {
	settings := config.Settings

	filename, ok := settings["filename"].(string)
	if !ok {
		return nil, fmt.Errorf("filename is required for file output")
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Rotation settings
	maxSize := 10 // Default 10MB
	if ms, ok := settings["max_size_mb"].(int); ok {
		maxSize = ms
	}

	maxFiles := 5 // Default 5 files
	if mf, ok := settings["max_files"].(int); ok {
		maxFiles = mf
	}

	maxAge := 30 // Default 30 days
	if ma, ok := settings["max_age_days"].(int); ok {
		maxAge = ma
	}

	compress := true
	if c, ok := settings["compress"].(bool); ok {
		compress = c
	}

	writer := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    maxSize,
		MaxBackups: maxFiles,
		MaxAge:     maxAge,
		Compress:   compress,
		LocalTime:  true,
	}

	encoder, err := createEncoder(config.Format)
	if err != nil {
		return nil, fmt.Errorf("failed to create encoder: %w", err)
	}

	return &FileOutput{
		name:    name,
		config:  config,
		writer:  writer,
		encoder: encoder,
	}, nil
}

func (fo *FileOutput) Write(entry *LogEntry) error {
	fo.mutex.Lock()
	defer fo.mutex.Unlock()

	data, err := fo.encoder.Encode(entry)
	if err != nil {
		return fmt.Errorf("failed to encode log entry: %w", err)
	}

	data = append(data, '\n')

	if _, err := fo.writer.Write(data); err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}

func (fo *FileOutput) Flush() error {
	// lumberjack doesn't have explicit flush, but file writes are immediate
	return nil
}

func (fo *FileOutput) Close() error {
	fo.mutex.Lock()
	defer fo.mutex.Unlock()

	return fo.writer.Close()
}

func (fo *FileOutput) Name() string {
	return fo.name
}

// NewConsoleOutput creates a new console output
func NewConsoleOutput(name string, config *OutputConfig) (*ConsoleOutput, error) {
	encoder, err := createEncoder(config.Format)
	if err != nil {
		return nil, fmt.Errorf("failed to create encoder: %w", err)
	}

	writer := bufio.NewWriter(os.Stdout)
	if target, ok := config.Settings["target"].(string); ok && target == "stderr" {
		writer = bufio.NewWriter(os.Stderr)
	}

	return &ConsoleOutput{
		name:    name,
		config:  config,
		encoder: encoder,
		writer:  writer,
	}, nil
}

func (co *ConsoleOutput) Write(entry *LogEntry) error {
	co.mutex.Lock()
	defer co.mutex.Unlock()

	data, err := co.encoder.Encode(entry)
	if err != nil {
		return fmt.Errorf("failed to encode log entry: %w", err)
	}

	data = append(data, '\n')

	if _, err := co.writer.Write(data); err != nil {
		return fmt.Errorf("failed to write to console: %w", err)
	}

	return co.writer.Flush()
}

func (co *ConsoleOutput) Flush() error {
	co.mutex.Lock()
	defer co.mutex.Unlock()

	return co.writer.Flush()
}

func (co *ConsoleOutput) Close() error {
	return co.Flush()
}

func (co *ConsoleOutput) Name() string {
	return co.name
}

// NewSyslogOutput creates a new syslog output
func NewSyslogOutput(name string, config *OutputConfig) (*SyslogOutput, error) {
	network := "tcp"
	if n, ok := config.Settings["network"].(string); ok {
		network = n
	}

	address := "localhost:514"
	if a, ok := config.Settings["address"].(string); ok {
		address = a
	}

	tag := "gzh-manager"
	if t, ok := config.Settings["tag"].(string); ok {
		tag = t
	}

	priority := syslog.LOG_INFO | syslog.LOG_DAEMON
	if p, ok := config.Settings["priority"].(int); ok {
		priority = syslog.Priority(p)
	}

	writer, err := syslog.Dial(network, address, priority, tag)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to syslog: %w", err)
	}

	encoder, err := createEncoder(config.Format)
	if err != nil {
		return nil, fmt.Errorf("failed to create encoder: %w", err)
	}

	return &SyslogOutput{
		name:    name,
		config:  config,
		writer:  writer,
		encoder: encoder,
	}, nil
}

func (so *SyslogOutput) Write(entry *LogEntry) error {
	so.mutex.Lock()
	defer so.mutex.Unlock()

	data, err := so.encoder.Encode(entry)
	if err != nil {
		return fmt.Errorf("failed to encode log entry: %w", err)
	}

	// Send to syslog based on level
	switch entry.Level {
	case "debug":
		return so.writer.Debug(string(data))
	case "info":
		return so.writer.Info(string(data))
	case "warn":
		return so.writer.Warning(string(data))
	case "error":
		return so.writer.Err(string(data))
	default:
		return so.writer.Info(string(data))
	}
}

func (so *SyslogOutput) Flush() error {
	// Syslog writes are immediate
	return nil
}

func (so *SyslogOutput) Close() error {
	so.mutex.Lock()
	defer so.mutex.Unlock()

	return so.writer.Close()
}

func (so *SyslogOutput) Name() string {
	return so.name
}

// NewHTTPOutput creates a new HTTP output
func NewHTTPOutput(name string, config *OutputConfig) (*HTTPOutput, error) {
	endpoint, ok := config.Settings["endpoint"].(string)
	if !ok {
		return nil, fmt.Errorf("endpoint is required for HTTP output")
	}

	bufferSize := 100
	if bs, ok := config.Settings["buffer_size"].(int); ok {
		bufferSize = bs
	}

	timeout := 30 * time.Second
	if t, ok := config.Settings["timeout"].(string); ok {
		if duration, err := time.ParseDuration(t); err == nil {
			timeout = duration
		}
	}

	client := &http.Client{
		Timeout: timeout,
	}

	headers := make(map[string]string)
	if h, ok := config.Settings["headers"].(map[string]interface{}); ok {
		for k, v := range h {
			if str, ok := v.(string); ok {
				headers[k] = str
			}
		}
	}

	encoder, err := createEncoder(config.Format)
	if err != nil {
		return nil, fmt.Errorf("failed to create encoder: %w", err)
	}

	return &HTTPOutput{
		name:       name,
		config:     config,
		client:     client,
		endpoint:   endpoint,
		headers:    headers,
		buffer:     make([]*LogEntry, 0, bufferSize),
		bufferSize: bufferSize,
		encoder:    encoder,
	}, nil
}

func (ho *HTTPOutput) Write(entry *LogEntry) error {
	ho.mutex.Lock()
	defer ho.mutex.Unlock()

	ho.buffer = append(ho.buffer, entry)

	// Flush if buffer is full
	if len(ho.buffer) >= ho.bufferSize {
		return ho.flushBuffer()
	}

	return nil
}

func (ho *HTTPOutput) Flush() error {
	ho.mutex.Lock()
	defer ho.mutex.Unlock()

	return ho.flushBuffer()
}

func (ho *HTTPOutput) flushBuffer() error {
	if len(ho.buffer) == 0 {
		return nil
	}

	// Encode all entries
	var entries []json.RawMessage
	for _, entry := range ho.buffer {
		data, err := ho.encoder.Encode(entry)
		if err != nil {
			continue
		}
		entries = append(entries, data)
	}

	// Create request body
	requestBody := map[string]interface{}{
		"entries":   entries,
		"timestamp": time.Now(),
		"source":    "gzh-manager",
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", ho.endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range ho.headers {
		req.Header.Set(k, v)
	}

	// Send request
	resp, err := ho.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP request failed with status: %d", resp.StatusCode)
	}

	// Clear buffer on successful send
	ho.buffer = ho.buffer[:0]

	return nil
}

func (ho *HTTPOutput) Close() error {
	return ho.Flush()
}

func (ho *HTTPOutput) Name() string {
	return ho.name
}

// Encoder implementations

func createEncoder(format string) (LogEncoder, error) {
	switch format {
	case "json":
		return &JSONEncoder{}, nil
	case "text":
		return &TextEncoder{
			template: "{{.Timestamp}} [{{.Level}}] {{.Logger}}: {{.Message}}",
		}, nil
	case "structured":
		return &StructuredEncoder{
			includeTimestamp: true,
			includeSource:    true,
		}, nil
	default:
		return &JSONEncoder{}, nil
	}
}

func (je *JSONEncoder) Encode(entry *LogEntry) ([]byte, error) {
	return json.Marshal(entry)
}

func (te *TextEncoder) Encode(entry *LogEntry) ([]byte, error) {
	// Simple text format implementation
	timestamp := entry.Timestamp.Format("2006-01-02 15:04:05")
	text := fmt.Sprintf("%s [%s] %s: %s",
		timestamp,
		entry.Level,
		entry.Logger,
		entry.Message)

	// Add fields if present
	if len(entry.Fields) > 0 {
		fieldStr := ""
		for k, v := range entry.Fields {
			if fieldStr != "" {
				fieldStr += ", "
			}
			fieldStr += fmt.Sprintf("%s=%v", k, v)
		}
		text += fmt.Sprintf(" {%s}", fieldStr)
	}

	return []byte(text), nil
}

func (se *StructuredEncoder) Encode(entry *LogEntry) ([]byte, error) {
	// Structured format with key-value pairs
	var buf bytes.Buffer

	if se.includeTimestamp {
		buf.WriteString(fmt.Sprintf("time=%s ", entry.Timestamp.Format(time.RFC3339)))
	}

	buf.WriteString(fmt.Sprintf("level=%s ", entry.Level))
	buf.WriteString(fmt.Sprintf("logger=%s ", entry.Logger))
	buf.WriteString(fmt.Sprintf("msg=\"%s\"", entry.Message))

	// Add fields
	for k, v := range entry.Fields {
		buf.WriteString(fmt.Sprintf(" %s=%v", k, v))
	}

	// Add labels
	for k, v := range entry.Labels {
		buf.WriteString(fmt.Sprintf(" label_%s=%s", k, v))
	}

	// Add source information if available and requested
	if se.includeSource && entry.Source != nil {
		if entry.Source.File != "" {
			buf.WriteString(fmt.Sprintf(" file=%s", entry.Source.File))
		}
		if entry.Source.Line != 0 {
			buf.WriteString(fmt.Sprintf(" line=%d", entry.Source.Line))
		}
		if entry.Source.Function != "" {
			buf.WriteString(fmt.Sprintf(" func=%s", entry.Source.Function))
		}
	}

	// Add trace information if available
	if entry.TraceID != "" {
		buf.WriteString(fmt.Sprintf(" trace_id=%s", entry.TraceID))
	}
	if entry.SpanID != "" {
		buf.WriteString(fmt.Sprintf(" span_id=%s", entry.SpanID))
	}

	return buf.Bytes(), nil
}
