package async

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"
)

// Event represents a generic event in the system
type Event interface {
	Type() string
	Timestamp() time.Time
	Source() string
	Data() interface{}
}

// BaseEvent provides a basic implementation of Event
type BaseEvent struct {
	EventType   string      `json:"type"`
	EventTime   time.Time   `json:"timestamp"`
	EventSource string      `json:"source"`
	EventData   interface{} `json:"data"`
}

func (e BaseEvent) Type() string         { return e.EventType }
func (e BaseEvent) Timestamp() time.Time { return e.EventTime }
func (e BaseEvent) Source() string       { return e.EventSource }
func (e BaseEvent) Data() interface{}    { return e.EventData }

// EventHandler defines the interface for event handlers
type EventHandler interface {
	Handle(ctx context.Context, event Event) error
}

// EventHandlerFunc is an adapter to allow the use of ordinary functions as EventHandlers
type EventHandlerFunc func(ctx context.Context, event Event) error

func (f EventHandlerFunc) Handle(ctx context.Context, event Event) error {
	return f(ctx, event)
}

// EventBus provides event-driven architecture capabilities
type EventBus struct {
	mu            sync.RWMutex
	handlers      map[string][]EventHandler
	asyncHandlers map[string][]EventHandler
	middleware    []MiddlewareFunc
	stats         EventStats
	bufferSize    int
	workerPool    chan struct{}
	stopCh        chan struct{}
	wg            sync.WaitGroup
}

// EventStats tracks event bus performance metrics
type EventStats struct {
	TotalEvents     int64
	ProcessedEvents int64
	FailedEvents    int64
	HandlerErrors   int64
	AverageLatency  time.Duration
	ActiveHandlers  int
}

// MiddlewareFunc defines middleware for event processing
type MiddlewareFunc func(next EventHandler) EventHandler

// EventBusConfig configures the event bus
type EventBusConfig struct {
	BufferSize     int
	MaxWorkers     int
	EnableMetrics  bool
	DefaultTimeout time.Duration
}

// DefaultEventBusConfig returns sensible defaults
func DefaultEventBusConfig() EventBusConfig {
	return EventBusConfig{
		BufferSize:     1000,
		MaxWorkers:     10,
		EnableMetrics:  true,
		DefaultTimeout: 30 * time.Second,
	}
}

// NewEventBus creates a new event bus
func NewEventBus(config EventBusConfig) *EventBus {
	eb := &EventBus{
		handlers:      make(map[string][]EventHandler),
		asyncHandlers: make(map[string][]EventHandler),
		bufferSize:    config.BufferSize,
		workerPool:    make(chan struct{}, config.MaxWorkers),
		stopCh:        make(chan struct{}),
	}

	// Initialize worker pool
	for i := 0; i < config.MaxWorkers; i++ {
		eb.workerPool <- struct{}{}
	}

	return eb
}

// Subscribe adds a synchronous event handler for a specific event type
func (eb *EventBus) Subscribe(eventType string, handler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.handlers[eventType] = append(eb.handlers[eventType], handler)
}

// SubscribeAsync adds an asynchronous event handler for a specific event type
func (eb *EventBus) SubscribeAsync(eventType string, handler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.asyncHandlers[eventType] = append(eb.asyncHandlers[eventType], handler)
}

// SubscribeFunc is a convenience method for subscribing function handlers
func (eb *EventBus) SubscribeFunc(eventType string, handler func(ctx context.Context, event Event) error) {
	eb.Subscribe(eventType, EventHandlerFunc(handler))
}

// SubscribeAsyncFunc is a convenience method for subscribing async function handlers
func (eb *EventBus) SubscribeAsyncFunc(eventType string, handler func(ctx context.Context, event Event) error) {
	eb.SubscribeAsync(eventType, EventHandlerFunc(handler))
}

// Publish synchronously publishes an event to all subscribers
func (eb *EventBus) Publish(ctx context.Context, event Event) error {
	start := time.Now()
	eb.updateStats(true, false)

	// Apply middleware chain
	handler := eb.createHandlerChain(event.Type())

	err := handler.Handle(ctx, event)
	duration := time.Since(start)

	eb.updateLatency(duration)
	if err != nil {
		eb.updateStats(false, true)
		return fmt.Errorf("event handling failed: %w", err)
	}

	eb.updateStats(false, false)
	return nil
}

// PublishAsync asynchronously publishes an event to all subscribers
func (eb *EventBus) PublishAsync(ctx context.Context, event Event) {
	eb.wg.Add(1)
	go func() {
		defer eb.wg.Done()

		// Wait for available worker
		select {
		case <-eb.workerPool:
			defer func() { eb.workerPool <- struct{}{} }()
		case <-ctx.Done():
			return
		case <-eb.stopCh:
			return
		}

		if err := eb.Publish(ctx, event); err != nil {
			// Log error (in real implementation, use proper logger)
			fmt.Printf("Async event handling error: %v\n", err)
		}
	}()
}

// createHandlerChain creates a handler chain with middleware
func (eb *EventBus) createHandlerChain(eventType string) EventHandler {
	eb.mu.RLock()
	syncHandlers := eb.handlers[eventType]
	asyncHandlers := eb.asyncHandlers[eventType]
	eb.mu.RUnlock()

	return EventHandlerFunc(func(ctx context.Context, event Event) error {
		// Execute synchronous handlers first
		for _, handler := range syncHandlers {
			wrappedHandler := eb.applyMiddleware(handler)
			if err := wrappedHandler.Handle(ctx, event); err != nil {
				eb.updateStats(false, true)
				return err
			}
		}

		// Execute asynchronous handlers
		if len(asyncHandlers) > 0 {
			var wg sync.WaitGroup
			errCh := make(chan error, len(asyncHandlers))

			for _, handler := range asyncHandlers {
				wg.Add(1)
				go func(h EventHandler) {
					defer wg.Done()
					wrappedHandler := eb.applyMiddleware(h)
					if err := wrappedHandler.Handle(ctx, event); err != nil {
						errCh <- err
					}
				}(handler)
			}

			wg.Wait()
			close(errCh)

			// Collect any errors
			var errors []error
			for err := range errCh {
				errors = append(errors, err)
				eb.updateStats(false, true)
			}

			if len(errors) > 0 {
				return fmt.Errorf("async handler errors: %v", errors)
			}
		}

		return nil
	})
}

// applyMiddleware applies all registered middleware to a handler
func (eb *EventBus) applyMiddleware(handler EventHandler) EventHandler {
	result := handler
	for i := len(eb.middleware) - 1; i >= 0; i-- {
		result = eb.middleware[i](result)
	}
	return result
}

// Use adds middleware to the event bus
func (eb *EventBus) Use(middleware MiddlewareFunc) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.middleware = append(eb.middleware, middleware)
}

// Unsubscribe removes a handler for a specific event type
func (eb *EventBus) Unsubscribe(eventType string, handler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	// Remove from synchronous handlers
	if handlers, exists := eb.handlers[eventType]; exists {
		eb.handlers[eventType] = eb.removeHandler(handlers, handler)
	}

	// Remove from asynchronous handlers
	if handlers, exists := eb.asyncHandlers[eventType]; exists {
		eb.asyncHandlers[eventType] = eb.removeHandler(handlers, handler)
	}
}

// removeHandler removes a handler from a slice
func (eb *EventBus) removeHandler(handlers []EventHandler, target EventHandler) []EventHandler {
	targetPtr := reflect.ValueOf(target).Pointer()
	result := make([]EventHandler, 0, len(handlers))

	for _, handler := range handlers {
		if reflect.ValueOf(handler).Pointer() != targetPtr {
			result = append(result, handler)
		}
	}

	return result
}

// updateStats updates event bus statistics
func (eb *EventBus) updateStats(starting, error bool) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if starting {
		eb.stats.TotalEvents++
		eb.stats.ActiveHandlers++
	} else {
		eb.stats.ActiveHandlers--
		if error {
			eb.stats.FailedEvents++
			eb.stats.HandlerErrors++
		} else {
			eb.stats.ProcessedEvents++
		}
	}
}

// updateLatency updates average latency
func (eb *EventBus) updateLatency(duration time.Duration) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if eb.stats.AverageLatency == 0 {
		eb.stats.AverageLatency = duration
	} else {
		alpha := 0.1
		eb.stats.AverageLatency = time.Duration(
			alpha*float64(duration) + (1-alpha)*float64(eb.stats.AverageLatency),
		)
	}
}

// GetStats returns current event bus statistics
func (eb *EventBus) GetStats() EventStats {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	return eb.stats
}

// PrintStats prints detailed event bus statistics
func (eb *EventBus) PrintStats() {
	stats := eb.GetStats()

	fmt.Printf("=== Event Bus Statistics ===\n")
	fmt.Printf("Total Events: %d\n", stats.TotalEvents)
	fmt.Printf("Processed: %d\n", stats.ProcessedEvents)
	fmt.Printf("Failed: %d\n", stats.FailedEvents)
	fmt.Printf("Handler Errors: %d\n", stats.HandlerErrors)
	fmt.Printf("Active Handlers: %d\n", stats.ActiveHandlers)
	fmt.Printf("Average Latency: %v\n", stats.AverageLatency)

	if stats.TotalEvents > 0 {
		successRate := float64(stats.ProcessedEvents) / float64(stats.TotalEvents) * 100
		fmt.Printf("Success Rate: %.2f%%\n", successRate)
	}
}

// Close gracefully shuts down the event bus
func (eb *EventBus) Close() error {
	close(eb.stopCh)

	// Wait for all async operations to complete with timeout
	done := make(chan struct{})
	go func() {
		eb.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-time.After(10 * time.Second):
		return fmt.Errorf("timeout waiting for event handlers to complete")
	}
}

// Common event types
const (
	EventTypeRepositoryCloned    = "repository.cloned"
	EventTypeRepositoryUpdated   = "repository.updated"
	EventTypeAPIRequestCompleted = "api.request.completed"
	EventTypeFileProcessed       = "file.processed"
	EventTypeErrorOccurred       = "error.occurred"
	EventTypeTaskCompleted       = "task.completed"
)

// Common event factories
func NewRepositoryClonedEvent(source, repoURL string, data interface{}) Event {
	return BaseEvent{
		EventType:   EventTypeRepositoryCloned,
		EventTime:   time.Now(),
		EventSource: source,
		EventData:   data,
	}
}

func NewFileProcessedEvent(source, filePath string, data interface{}) Event {
	return BaseEvent{
		EventType:   EventTypeFileProcessed,
		EventTime:   time.Now(),
		EventSource: source,
		EventData:   data,
	}
}

func NewErrorEvent(source string, err error) Event {
	return BaseEvent{
		EventType:   EventTypeErrorOccurred,
		EventTime:   time.Now(),
		EventSource: source,
		EventData:   map[string]interface{}{"error": err.Error()},
	}
}
