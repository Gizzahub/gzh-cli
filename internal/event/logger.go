package event

import (
	"fmt"
	"log"
)

// SimpleLogger implements a basic logger.
type SimpleLogger struct{}

// NewSimpleLogger creates a new simple logger.
func NewSimpleLogger() *SimpleLogger {
	return &SimpleLogger{}
}

// Info logs an info message.
func (l *SimpleLogger) Info(msg string, keysAndValues ...interface{}) {
	log.Printf("[INFO] %s %v", msg, keysAndValues)
}

// Error logs an error message.
func (l *SimpleLogger) Error(msg string, err error, keysAndValues ...interface{}) {
	log.Printf("[ERROR] %s: %v %v", msg, err, keysAndValues)
}

// Debug logs a debug message.
func (l *SimpleLogger) Debug(msg string, keysAndValues ...interface{}) {
	log.Printf("[DEBUG] %s %v", msg, keysAndValues)
}

// Warn logs a warning message.
func (l *SimpleLogger) Warn(msg string, keysAndValues ...interface{}) {
	log.Printf("[WARN] %s %v", msg, keysAndValues)
}

// With returns a logger with additional context.
func (l *SimpleLogger) With(keysAndValues ...interface{}) interface{} {
	return l
}

// getLogger returns an EventLogger implementation.
func GetLogger() interface{} {
	return NewSimpleLogger()
}

// FormatEvents formats events for output.
func FormatEvents(events []*interface{}, format string) error {
	switch format {
	case "json":
		// JSON formatting implementation
		fmt.Println("[JSON output]")
	case "yaml":
		// YAML formatting implementation
		fmt.Println("[YAML output]")
	default:
		// Table formatting implementation
		fmt.Println("[Table output]")
	}

	return nil
}
