package debug

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gizzahub/gzh-manager-go/cmd/monitoring"
)

// CentralizedLoggerBridge bridges StructuredLogger with CentralizedLogger
type CentralizedLoggerBridge struct {
	structuredLogger  *StructuredLogger
	centralizedLogger *monitoring.CentralizedLogger
	enabled           bool
	bufferSize        int
	logChannel        chan *monitoring.LogEntry
	ctx               context.Context
	cancel            context.CancelFunc
}

// CentralizedBridgeConfig configures the bridge between structured and centralized logging
type CentralizedBridgeConfig struct {
	Enabled            bool              `json:"enabled"`
	BufferSize         int               `json:"buffer_size"`
	CentralizedConfig  string            `json:"centralized_config_path"`
	ForwardLevels      []RFC5424Severity `json:"forward_levels"`
	AddStructuredData  bool              `json:"add_structured_data"`
	PreserveOriginalID bool              `json:"preserve_original_id"`
}

// NewCentralizedLoggerBridge creates a bridge between structured and centralized logging
func NewCentralizedLoggerBridge(
	structuredLogger *StructuredLogger,
	centralizedLogger *monitoring.CentralizedLogger,
	config *CentralizedBridgeConfig,
) (*CentralizedLoggerBridge, error) {
	if config == nil {
		config = &CentralizedBridgeConfig{
			Enabled:           true,
			BufferSize:        1000,
			AddStructuredData: true,
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	bridge := &CentralizedLoggerBridge{
		structuredLogger:  structuredLogger,
		centralizedLogger: centralizedLogger,
		enabled:           config.Enabled,
		bufferSize:        config.BufferSize,
		logChannel:        make(chan *monitoring.LogEntry, config.BufferSize),
		ctx:               ctx,
		cancel:            cancel,
	}

	if config.Enabled {
		bridge.startForwardingWorker()
	}

	return bridge, nil
}

// ForwardLogEntry forwards a structured log entry to the centralized logging system
func (clb *CentralizedLoggerBridge) ForwardLogEntry(entry *StructuredLogEntry) error {
	if !clb.enabled || clb.centralizedLogger == nil {
		return nil
	}

	// Convert StructuredLogEntry to CentralizedLogger LogEntry
	centralizedEntry := &monitoring.LogEntry{
		Timestamp: entry.Timestamp,
		Level:     entry.Level,
		Message:   entry.Message,
		Logger:    entry.AppName,
		Fields:    make(map[string]interface{}),
		Labels:    make(map[string]string),
		TraceID:   entry.TraceID,
		SpanID:    entry.SpanID,
	}

	// Add source information
	if entry.Caller.File != "" {
		centralizedEntry.Source = &monitoring.LogSource{
			File:     entry.Caller.File,
			Line:     entry.Caller.Line,
			Function: entry.Caller.Function,
		}
	}

	// Copy fields
	if entry.Fields != nil {
		for k, v := range entry.Fields {
			centralizedEntry.Fields[k] = v
		}
	}

	// Add structured logging metadata
	centralizedEntry.Fields["structured_logger"] = true
	centralizedEntry.Fields["rfc5424_severity"] = int(entry.Severity)
	centralizedEntry.Fields["hostname"] = entry.Hostname
	centralizedEntry.Fields["app_name"] = entry.AppName
	centralizedEntry.Fields["proc_id"] = entry.ProcID
	centralizedEntry.Fields["version"] = entry.Version

	if entry.MsgID != "" {
		centralizedEntry.Fields["msg_id"] = entry.MsgID
	}

	// Add performance metrics if available
	if entry.Duration != nil {
		centralizedEntry.Fields["duration_ms"] = entry.Duration.Milliseconds()
	}
	if entry.Latency != nil {
		centralizedEntry.Fields["latency_ms"] = entry.Latency.Milliseconds()
	}
	if entry.BytesRead != nil {
		centralizedEntry.Fields["bytes_read"] = *entry.BytesRead
	}
	if entry.BytesOut != nil {
		centralizedEntry.Fields["bytes_out"] = *entry.BytesOut
	}

	// Add default labels
	centralizedEntry.Labels["service"] = entry.AppName
	centralizedEntry.Labels["host"] = entry.Hostname
	centralizedEntry.Labels["log_format"] = "structured"

	// Forward to centralized logger
	select {
	case clb.logChannel <- centralizedEntry:
		return nil
	default:
		// Channel is full, log directly to avoid blocking
		return clb.centralizedLogger.Log(centralizedEntry)
	}
}

// startForwardingWorker starts the worker goroutine for forwarding log entries
func (clb *CentralizedLoggerBridge) startForwardingWorker() {
	go func() {
		for {
			select {
			case entry := <-clb.logChannel:
				if err := clb.centralizedLogger.Log(entry); err != nil {
					// If centralized logging fails, fall back to structured logger
					if clb.structuredLogger != nil {
						clb.structuredLogger.ErrorLevel(clb.ctx,
							"Failed to forward log to centralized system",
							map[string]interface{}{
								"error":            err.Error(),
								"original_message": entry.Message,
								"original_level":   entry.Level,
							})
					}
				}
			case <-clb.ctx.Done():
				// Drain remaining entries
				for {
					select {
					case entry := <-clb.logChannel:
						clb.centralizedLogger.Log(entry)
					default:
						return
					}
				}
			}
		}
	}()
}

// EnableForwarding enables log forwarding to centralized system
func (clb *CentralizedLoggerBridge) EnableForwarding() {
	clb.enabled = true
}

// DisableForwarding disables log forwarding to centralized system
func (clb *CentralizedLoggerBridge) DisableForwarding() {
	clb.enabled = false
}

// IsForwardingEnabled returns whether forwarding is enabled
func (clb *CentralizedLoggerBridge) IsForwardingEnabled() bool {
	return clb.enabled
}

// GetStats returns bridge statistics
func (clb *CentralizedLoggerBridge) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"enabled":            clb.enabled,
		"buffer_size":        clb.bufferSize,
		"channel_length":     len(clb.logChannel),
		"buffer_utilization": float64(len(clb.logChannel)) / float64(clb.bufferSize) * 100,
	}
}

// Shutdown gracefully shuts down the bridge
func (clb *CentralizedLoggerBridge) Shutdown() error {
	if clb.cancel != nil {
		clb.cancel()
	}

	// Wait a moment for worker to drain
	time.Sleep(100 * time.Millisecond)

	return nil
}

// EnhancedStructuredLogger extends StructuredLogger with centralized logging integration
type EnhancedStructuredLogger struct {
	*StructuredLogger
	bridge *CentralizedLoggerBridge
}

// NewEnhancedStructuredLogger creates a structured logger with centralized logging integration
func NewEnhancedStructuredLogger(
	config *StructuredLoggerConfig,
	centralizedLogger *monitoring.CentralizedLogger,
	bridgeConfig *CentralizedBridgeConfig,
) (*EnhancedStructuredLogger, error) {
	structuredLogger, err := NewStructuredLogger(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create structured logger: %w", err)
	}

	var bridge *CentralizedLoggerBridge
	if centralizedLogger != nil {
		bridge, err = NewCentralizedLoggerBridge(structuredLogger, centralizedLogger, bridgeConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create bridge: %w", err)
		}
	}

	return &EnhancedStructuredLogger{
		StructuredLogger: structuredLogger,
		bridge:           bridge,
	}, nil
}

// log overrides the base log method to include centralized forwarding
func (esl *EnhancedStructuredLogger) log(ctx context.Context, level RFC5424Severity, msg string, fields ...map[string]interface{}) {
	// First, log through the structured logger
	esl.StructuredLogger.log(ctx, level, msg, fields...)

	// Then forward to centralized system if bridge is enabled
	if esl.bridge != nil && esl.bridge.IsForwardingEnabled() {
		// Create the log entry similar to how StructuredLogger does it
		entry := &StructuredLogEntry{
			Timestamp: time.Now().UTC(),
			Version:   1,
			Level:     rfc5424Names[level],
			Severity:  level,
			Hostname:  esl.StructuredLogger.hostname,
			AppName:   esl.StructuredLogger.config.AppName,
			ProcID:    fmt.Sprintf("%d", os.Getpid()),
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

		// Forward to centralized logger
		if err := esl.bridge.ForwardLogEntry(entry); err != nil {
			// Log the forwarding error using structured logger only (avoid recursion)
			esl.StructuredLogger.log(ctx, SeverityWarning,
				"Failed to forward log entry to centralized system",
				map[string]interface{}{"error": err.Error()})
		}
	}
}

// Override all the logging methods to use the enhanced log function

// Emergency logs an emergency message with centralized forwarding
func (esl *EnhancedStructuredLogger) Emergency(ctx context.Context, msg string, fields ...map[string]interface{}) {
	esl.log(ctx, SeverityEmergency, msg, fields...)
}

// Alert logs an alert message with centralized forwarding
func (esl *EnhancedStructuredLogger) Alert(ctx context.Context, msg string, fields ...map[string]interface{}) {
	esl.log(ctx, SeverityAlert, msg, fields...)
}

// Critical logs a critical message with centralized forwarding
func (esl *EnhancedStructuredLogger) Critical(ctx context.Context, msg string, fields ...map[string]interface{}) {
	esl.log(ctx, SeverityCritical, msg, fields...)
}

// ErrorLevel logs an error message with centralized forwarding
func (esl *EnhancedStructuredLogger) ErrorLevel(ctx context.Context, msg string, fields ...map[string]interface{}) {
	esl.log(ctx, SeverityError, msg, fields...)
}

// Warning logs a warning message with centralized forwarding
func (esl *EnhancedStructuredLogger) Warning(ctx context.Context, msg string, fields ...map[string]interface{}) {
	esl.log(ctx, SeverityWarning, msg, fields...)
}

// Notice logs a notice message with centralized forwarding
func (esl *EnhancedStructuredLogger) Notice(ctx context.Context, msg string, fields ...map[string]interface{}) {
	esl.log(ctx, SeverityNotice, msg, fields...)
}

// InfoLevel logs an info message with centralized forwarding
func (esl *EnhancedStructuredLogger) InfoLevel(ctx context.Context, msg string, fields ...map[string]interface{}) {
	esl.log(ctx, SeverityInfo, msg, fields...)
}

// DebugLevel logs a debug message with centralized forwarding
func (esl *EnhancedStructuredLogger) DebugLevel(ctx context.Context, msg string, fields ...map[string]interface{}) {
	esl.log(ctx, SeverityDebug, msg, fields...)
}

// GetBridge returns the centralized logging bridge
func (esl *EnhancedStructuredLogger) GetBridge() *CentralizedLoggerBridge {
	return esl.bridge
}

// GetCentralizedLogger returns the centralized logger instance
func (esl *EnhancedStructuredLogger) GetCentralizedLogger() *monitoring.CentralizedLogger {
	if esl.bridge != nil {
		return esl.bridge.centralizedLogger
	}
	return nil
}

// Close closes both the structured logger and the bridge
func (esl *EnhancedStructuredLogger) Close() error {
	var err error

	if esl.bridge != nil {
		if bridgeErr := esl.bridge.Shutdown(); bridgeErr != nil {
			err = bridgeErr
		}
	}

	if structuredErr := esl.StructuredLogger.Close(); structuredErr != nil {
		if err != nil {
			return fmt.Errorf("multiple errors: bridge: %v, structured: %v", err, structuredErr)
		}
		err = structuredErr
	}

	return err
}
