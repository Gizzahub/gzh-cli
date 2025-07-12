package plugins

import (
	"sync"
	"time"
)

// EventBus handles event distribution between plugins and the host
type EventBus struct {
	subscribers map[string][]EventHandler
	mu          sync.RWMutex
	buffer      []Event
	bufferSize  int
}

// NewEventBus creates a new event bus
func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string][]EventHandler),
		buffer:      make([]Event, 0),
		bufferSize:  1000,
	}
}

// Subscribe adds an event handler for a specific event type
func (eb *EventBus) Subscribe(eventType string, handler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.subscribers[eventType] = append(eb.subscribers[eventType], handler)
}

// Unsubscribe removes an event handler (simplified implementation)
func (eb *EventBus) Unsubscribe(eventType string) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	delete(eb.subscribers, eventType)
}

// Emit sends an event to all subscribers
func (eb *EventBus) Emit(event Event) error {
	eb.mu.RLock()
	handlers := eb.subscribers[event.Type]
	allHandlers := eb.subscribers["*"] // Wildcard subscribers
	eb.mu.RUnlock()

	// Add to buffer for replay
	eb.addToBuffer(event)

	// Send to specific event type handlers
	for _, handler := range handlers {
		go func(h EventHandler) {
			defer func() {
				if r := recover(); r != nil {
					// Log handler panic but don't crash
				}
			}()
			h(event)
		}(handler)
	}

	// Send to wildcard handlers
	for _, handler := range allHandlers {
		go func(h EventHandler) {
			defer func() {
				if r := recover(); r != nil {
					// Log handler panic but don't crash
				}
			}()
			h(event)
		}(handler)
	}

	return nil
}

// addToBuffer adds an event to the circular buffer
func (eb *EventBus) addToBuffer(event Event) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if len(eb.buffer) >= eb.bufferSize {
		// Remove oldest event
		eb.buffer = eb.buffer[1:]
	}

	eb.buffer = append(eb.buffer, event)
}

// GetRecentEvents returns recent events from the buffer
func (eb *EventBus) GetRecentEvents(count int) []Event {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	if count <= 0 || count > len(eb.buffer) {
		count = len(eb.buffer)
	}

	start := len(eb.buffer) - count
	return append([]Event(nil), eb.buffer[start:]...)
}

// GetEventsSince returns events since a specific timestamp
func (eb *EventBus) GetEventsSince(since time.Time) []Event {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	var events []Event
	for _, event := range eb.buffer {
		if event.Timestamp.After(since) {
			events = append(events, event)
		}
	}

	return events
}
