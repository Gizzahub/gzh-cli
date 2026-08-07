// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package github

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"time"
)

// ErrEventNotFound is returned when a requested event ID is not in storage.
var ErrEventNotFound = fmt.Errorf("event not found")

// MemoryEventStorage is a process-local, thread-safe EventStorage implementation.
// Events are retained only for the lifetime of the process.
type MemoryEventStorage struct {
	mu     sync.RWMutex
	events map[string]*GitHubEvent
	order  []string // insertion order for stable listing
}

// NewMemoryEventStorage creates an empty in-memory event store.
func NewMemoryEventStorage() *MemoryEventStorage {
	return &MemoryEventStorage{
		events: make(map[string]*GitHubEvent),
		order:  make([]string, 0),
	}
}

// StoreEvent persists (or overwrites) an event by ID.
func (s *MemoryEventStorage) StoreEvent(_ context.Context, event *GitHubEvent) error {
	if event == nil {
		return fmt.Errorf("event is nil")
	}
	if event.ID == "" {
		return fmt.Errorf("event ID is empty")
	}

	// Shallow copy so callers cannot mutate stored state via the same pointer.
	stored := *event
	if event.Payload != nil {
		stored.Payload = copyAnyMap(event.Payload)
	}
	if event.Headers != nil {
		stored.Headers = copyStringMap(event.Headers)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.events[event.ID]; !exists {
		s.order = append(s.order, event.ID)
	}
	s.events[event.ID] = &stored
	return nil
}

// GetEvent returns a stored event by ID.
func (s *MemoryEventStorage) GetEvent(_ context.Context, eventID string) (*GitHubEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	event, ok := s.events[eventID]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrEventNotFound, eventID)
	}
	return cloneEvent(event), nil
}

// ListEvents returns events matching filter (best-effort), with limit/offset.
// limit <= 0 means no limit. offset < 0 is treated as 0.
func (s *MemoryEventStorage) ListEvents(_ context.Context, filter *EventFilter, limit, offset int) ([]*GitHubEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if offset < 0 {
		offset = 0
	}

	matched := make([]*GitHubEvent, 0, len(s.order))
	for _, id := range s.order {
		event := s.events[id]
		if event == nil {
			continue
		}
		if matchesEventFilter(event, filter) {
			matched = append(matched, cloneEvent(event))
		}
	}

	if offset >= len(matched) {
		return []*GitHubEvent{}, nil
	}
	matched = matched[offset:]
	if limit > 0 && len(matched) > limit {
		matched = matched[:limit]
	}
	return matched, nil
}

// DeleteEvent removes an event by ID. Missing IDs are a no-op success.
func (s *MemoryEventStorage) DeleteEvent(_ context.Context, eventID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.events[eventID]; !ok {
		return nil
	}
	delete(s.events, eventID)
	s.order = slices.DeleteFunc(s.order, func(id string) bool { return id == eventID })
	return nil
}

// CountEvents returns the number of events matching filter (best-effort).
func (s *MemoryEventStorage) CountEvents(_ context.Context, filter *EventFilter) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	for _, id := range s.order {
		event := s.events[id]
		if event != nil && matchesEventFilter(event, filter) {
			count++
		}
	}
	return count, nil
}

// AggregateMetrics builds EventMetrics from currently stored events.
// Processing-time fields are zeroed (storage has no processor timing).
func (s *MemoryEventStorage) AggregateMetrics(_ context.Context) *EventMetrics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	metrics := &EventMetrics{
		EventsByType:         make(map[string]int64),
		EventsByOrganization: make(map[string]int64),
		HandlersStatus:       make(map[string]string),
	}

	var lastAt time.Time
	for _, id := range s.order {
		event := s.events[id]
		if event == nil {
			continue
		}
		metrics.TotalEventsReceived++
		metrics.TotalEventsProcessed++
		metrics.EventsByType[event.Type]++
		if event.Organization != "" {
			metrics.EventsByOrganization[event.Organization]++
		}
		if event.Timestamp.After(lastAt) {
			lastAt = event.Timestamp
		}
	}
	metrics.LastEventAt = lastAt
	return metrics
}

// Ensure MemoryEventStorage implements EventStorage.
var _ EventStorage = (*MemoryEventStorage)(nil)

func matchesEventFilter(event *GitHubEvent, filter *EventFilter) bool {
	if filter == nil {
		return true
	}
	if filter.Organization != "" && event.Organization != filter.Organization {
		return false
	}
	if filter.Repository != "" && event.Repository != filter.Repository {
		return false
	}
	if len(filter.EventTypes) > 0 {
		found := false
		for _, t := range filter.EventTypes {
			if string(t) == event.Type {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if len(filter.Actions) > 0 && event.Action != "" {
		found := false
		for _, a := range filter.Actions {
			if string(a) == event.Action {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if filter.Sender != "" && event.Sender != filter.Sender {
		return false
	}
	if filter.TimeRange != nil {
		if !filter.TimeRange.Start.IsZero() && event.Timestamp.Before(filter.TimeRange.Start) {
			return false
		}
		if !filter.TimeRange.End.IsZero() && event.Timestamp.After(filter.TimeRange.End) {
			return false
		}
	}
	return true
}

func cloneEvent(event *GitHubEvent) *GitHubEvent {
	if event == nil {
		return nil
	}
	out := *event
	if event.Payload != nil {
		out.Payload = copyAnyMap(event.Payload)
	}
	if event.Headers != nil {
		out.Headers = copyStringMap(event.Headers)
	}
	return &out
}

func copyStringMap(in map[string]string) map[string]string {
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func copyAnyMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
