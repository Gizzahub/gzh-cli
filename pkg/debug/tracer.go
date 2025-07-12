package debug

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// TraceEvent represents a single trace event
type TraceEvent struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Category  string                 `json:"category"`
	Phase     string                 `json:"phase"` // "B" (begin), "E" (end), "I" (instant)
	Timestamp int64                  `json:"ts"`    // microseconds
	PID       int                    `json:"pid"`
	TID       int                    `json:"tid"`
	Duration  int64                  `json:"dur,omitempty"` // microseconds
	Args      map[string]interface{} `json:"args,omitempty"`
	Stack     []string               `json:"stack,omitempty"`
}

// TracerConfig holds tracer configuration
type TracerConfig struct {
	Enabled       bool          `json:"enabled"`
	OutputFile    string        `json:"output_file"`
	MaxEvents     int           `json:"max_events"`
	IncludeStack  bool          `json:"include_stack"`
	StackDepth    int           `json:"stack_depth"`
	BufferSize    int           `json:"buffer_size"`
	FlushInterval time.Duration `json:"flush_interval"`
	Categories    []string      `json:"categories"`
}

// DefaultTracerConfig returns a default tracer configuration
func DefaultTracerConfig() *TracerConfig {
	return &TracerConfig{
		Enabled:       false,
		OutputFile:    "./debug-trace.json",
		MaxEvents:     100000,
		IncludeStack:  false,
		StackDepth:    10,
		BufferSize:    1000,
		FlushInterval: 5 * time.Second,
		Categories:    []string{"default"},
	}
}

// Tracer provides execution tracing capabilities
type Tracer struct {
	config    *TracerConfig
	events    []TraceEvent
	mu        sync.RWMutex
	file      *os.File
	active    bool
	cancel    context.CancelFunc
	eventID   int64
	startTime time.Time
}

// NewTracer creates a new tracer instance
func NewTracer(config *TracerConfig) *Tracer {
	if config == nil {
		config = DefaultTracerConfig()
	}

	return &Tracer{
		config:    config,
		events:    make([]TraceEvent, 0, config.BufferSize),
		active:    false,
		startTime: time.Now(),
	}
}

// Start starts the tracer
func (t *Tracer) Start(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.active {
		return fmt.Errorf("tracer is already active")
	}

	if !t.config.Enabled {
		return fmt.Errorf("tracer is disabled")
	}

	// Create output directory
	dir := filepath.Dir(t.config.OutputFile)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Open output file
	file, err := os.Create(t.config.OutputFile)
	if err != nil {
		return fmt.Errorf("failed to create trace file: %w", err)
	}
	t.file = file

	// Write Chrome trace format header
	if _, err := t.file.WriteString("[\n"); err != nil {
		return fmt.Errorf("failed to write trace header: %w", err)
	}

	// Set up context for cancellation
	ctx, cancel := context.WithCancel(ctx)
	t.cancel = cancel
	t.startTime = time.Now()
	t.active = true

	Info("Tracer started", map[string]interface{}{
		"output_file": t.config.OutputFile,
		"max_events":  t.config.MaxEvents,
	})

	// Start flush goroutine
	go t.flushLoop(ctx)

	return nil
}

// Stop stops the tracer
func (t *Tracer) Stop() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.active {
		return fmt.Errorf("tracer is not active")
	}

	if t.cancel != nil {
		t.cancel()
	}

	// Final flush
	t.flushEventsUnsafe()

	// Write Chrome trace format footer
	if t.file != nil {
		if len(t.events) > 0 {
			// Remove trailing comma if there are events
			t.file.Seek(-2, 2) // Go back 2 bytes to overwrite ",\n"
			t.file.WriteString("\n")
		}
		t.file.WriteString("]\n")
		t.file.Close()
	}

	t.active = false
	duration := time.Since(t.startTime)

	Info("Tracer stopped", map[string]interface{}{
		"duration":     duration.String(),
		"total_events": len(t.events),
		"output_file":  t.config.OutputFile,
	})

	return nil
}

// IsActive returns true if the tracer is currently active
func (t *Tracer) IsActive() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.active
}

// TraceSpan represents a traced span
type TraceSpan struct {
	tracer    *Tracer
	event     TraceEvent
	startTime time.Time
}

// Begin starts a new trace span
func (t *Tracer) Begin(name, category string, args ...map[string]interface{}) *TraceSpan {
	if !t.IsActive() {
		return &TraceSpan{} // Return empty span if not active
	}

	t.mu.Lock()
	t.eventID++
	eventID := t.eventID
	t.mu.Unlock()

	startTime := time.Now()
	event := TraceEvent{
		ID:        fmt.Sprintf("span_%d", eventID),
		Name:      name,
		Category:  category,
		Phase:     "B",
		Timestamp: t.timeToMicros(startTime),
		PID:       os.Getpid(),
		TID:       t.getGoroutineID(),
	}

	// Add arguments
	if len(args) > 0 {
		event.Args = make(map[string]interface{})
		for _, argMap := range args {
			for k, v := range argMap {
				event.Args[k] = v
			}
		}
	}

	// Add stack trace if enabled
	if t.config.IncludeStack {
		event.Stack = t.captureStack()
	}

	t.addEvent(event)

	return &TraceSpan{
		tracer:    t,
		event:     event,
		startTime: startTime,
	}
}

// End ends the trace span
func (span *TraceSpan) End(args ...map[string]interface{}) {
	if span.tracer == nil || !span.tracer.IsActive() {
		return
	}

	endTime := time.Now()
	duration := endTime.Sub(span.startTime)

	endEvent := TraceEvent{
		ID:        span.event.ID,
		Name:      span.event.Name,
		Category:  span.event.Category,
		Phase:     "E",
		Timestamp: span.tracer.timeToMicros(endTime),
		PID:       span.event.PID,
		TID:       span.event.TID,
		Duration:  int64(duration.Nanoseconds() / 1000), // Convert to microseconds
	}

	// Add arguments
	if len(args) > 0 {
		endEvent.Args = make(map[string]interface{})
		for _, argMap := range args {
			for k, v := range argMap {
				endEvent.Args[k] = v
			}
		}
	}

	span.tracer.addEvent(endEvent)
}

// Instant creates an instant trace event
func (t *Tracer) Instant(name, category string, args ...map[string]interface{}) {
	if !t.IsActive() {
		return
	}

	t.mu.Lock()
	t.eventID++
	eventID := t.eventID
	t.mu.Unlock()

	event := TraceEvent{
		ID:        fmt.Sprintf("instant_%d", eventID),
		Name:      name,
		Category:  category,
		Phase:     "I",
		Timestamp: t.timeToMicros(time.Now()),
		PID:       os.Getpid(),
		TID:       t.getGoroutineID(),
	}

	// Add arguments
	if len(args) > 0 {
		event.Args = make(map[string]interface{})
		for _, argMap := range args {
			for k, v := range argMap {
				event.Args[k] = v
			}
		}
	}

	// Add stack trace if enabled
	if t.config.IncludeStack {
		event.Stack = t.captureStack()
	}

	t.addEvent(event)
}

// addEvent adds an event to the tracer
func (t *Tracer) addEvent(event TraceEvent) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.active {
		return
	}

	// Check if we've reached the maximum number of events
	if len(t.events) >= t.config.MaxEvents {
		// Remove oldest event (simple circular buffer)
		copy(t.events, t.events[1:])
		t.events = t.events[:len(t.events)-1]
	}

	t.events = append(t.events, event)
}

// flushLoop periodically flushes events to the file
func (t *Tracer) flushLoop(ctx context.Context) {
	ticker := time.NewTicker(t.config.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			t.mu.Lock()
			t.flushEventsUnsafe()
			t.mu.Unlock()
		}
	}
}

// flushEventsUnsafe flushes events to file (must be called with lock held)
func (t *Tracer) flushEventsUnsafe() {
	if t.file == nil || len(t.events) == 0 {
		return
	}

	for _, event := range t.events {
		data, err := json.Marshal(event)
		if err != nil {
			Error("Failed to marshal trace event", map[string]interface{}{"error": err})
			continue
		}

		if _, err := t.file.Write(data); err != nil {
			Error("Failed to write trace event", map[string]interface{}{"error": err})
			continue
		}

		if _, err := t.file.WriteString(",\n"); err != nil {
			Error("Failed to write trace delimiter", map[string]interface{}{"error": err})
			continue
		}
	}

	// Clear events after writing
	t.events = t.events[:0]

	// Sync to disk
	if err := t.file.Sync(); err != nil {
		Error("Failed to sync trace file", map[string]interface{}{"error": err})
	}
}

// timeToMicros converts time to microseconds since start
func (t *Tracer) timeToMicros(tm time.Time) int64 {
	return tm.Sub(t.startTime).Nanoseconds() / 1000
}

// getGoroutineID gets the current goroutine ID (approximation)
func (t *Tracer) getGoroutineID() int {
	// This is a hack to get goroutine ID for tracing purposes
	// In production, you might want to use a more robust method
	return runtime.NumGoroutine()
}

// captureStack captures the current stack trace
func (t *Tracer) captureStack() []string {
	stack := make([]uintptr, t.config.StackDepth)
	n := runtime.Callers(3, stack) // Skip captureStack, addEvent, and Begin/Instant
	stack = stack[:n]

	frames := runtime.CallersFrames(stack)
	var result []string

	for {
		frame, more := frames.Next()
		result = append(result, fmt.Sprintf("%s:%d %s",
			filepath.Base(frame.File), frame.Line, frame.Function))

		if !more {
			break
		}
	}

	return result
}

// GetStats returns tracer statistics
func (t *Tracer) GetStats() map[string]interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()

	stats := map[string]interface{}{
		"active":       t.active,
		"total_events": len(t.events),
		"max_events":   t.config.MaxEvents,
		"output_file":  t.config.OutputFile,
		"buffer_usage": float64(len(t.events)) / float64(t.config.MaxEvents) * 100,
	}

	if t.active {
		stats["uptime"] = time.Since(t.startTime).String()
		stats["start_time"] = t.startTime
	}

	return stats
}

// Global tracer instance
var (
	globalTracer   *Tracer
	globalTracerMu sync.RWMutex
)

// InitGlobalTracer initializes the global tracer
func InitGlobalTracer(config *TracerConfig) error {
	tracer := NewTracer(config)

	globalTracerMu.Lock()
	if globalTracer != nil && globalTracer.IsActive() {
		globalTracer.Stop()
	}
	globalTracer = tracer
	globalTracerMu.Unlock()

	return nil
}

// GetGlobalTracer returns the global tracer
func GetGlobalTracer() *Tracer {
	globalTracerMu.RLock()
	defer globalTracerMu.RUnlock()
	return globalTracer
}

// TraceFunction traces a function call
func TraceFunction(name string, fn func()) {
	if tracer := GetGlobalTracer(); tracer != nil {
		span := tracer.Begin(name, "function")
		defer span.End()
		fn()
	} else {
		fn()
	}
}

// TraceFunctionWithResult traces a function call that returns a result
func TraceFunctionWithResult[T any](name string, fn func() T) T {
	if tracer := GetGlobalTracer(); tracer != nil {
		span := tracer.Begin(name, "function")
		defer span.End()
		return fn()
	}
	return fn()
}

// TraceFunctionWithError traces a function call that returns an error
func TraceFunctionWithError(name string, fn func() error) error {
	if tracer := GetGlobalTracer(); tracer != nil {
		span := tracer.Begin(name, "function")
		defer func() {
			// Add error information if function returned an error
			if err := recover(); err != nil {
				span.End(map[string]interface{}{"error": err})
				panic(err)
			}
		}()

		err := fn()
		if err != nil {
			span.End(map[string]interface{}{"error": err.Error()})
		} else {
			span.End()
		}
		return err
	}
	return fn()
}
