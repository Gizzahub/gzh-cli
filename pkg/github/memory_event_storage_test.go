// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package github

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestMemoryEventStorage_StoreGetDelete(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := NewMemoryEventStorage()

	event := &GitHubEvent{
		ID:           "evt-1",
		Type:         "push",
		Action:       "created",
		Organization: "acme",
		Repository:   "api",
		Sender:       "alice",
		Timestamp:    time.Now().UTC().Truncate(time.Second),
		Payload:      map[string]any{"ref": "refs/heads/main"},
		Headers:      map[string]string{"X-GitHub-Event": "push"},
	}

	if err := store.StoreEvent(ctx, event); err != nil {
		t.Fatalf("StoreEvent: %v", err)
	}

	got, err := store.GetEvent(ctx, "evt-1")
	if err != nil {
		t.Fatalf("GetEvent: %v", err)
	}
	if got.ID != event.ID || got.Organization != "acme" {
		t.Fatalf("unexpected event: %+v", got)
	}

	// Mutation of returned copy must not affect storage.
	got.Organization = "mutated"
	again, err := store.GetEvent(ctx, "evt-1")
	if err != nil {
		t.Fatalf("GetEvent again: %v", err)
	}
	if again.Organization != "acme" {
		t.Fatalf("storage mutated via returned pointer")
	}

	if err := store.DeleteEvent(ctx, "evt-1"); err != nil {
		t.Fatalf("DeleteEvent: %v", err)
	}
	_, err = store.GetEvent(ctx, "evt-1")
	if !errors.Is(err, ErrEventNotFound) {
		t.Fatalf("expected ErrEventNotFound, got %v", err)
	}
}

func TestMemoryEventStorage_GetNotFound(t *testing.T) {
	t.Parallel()
	_, err := NewMemoryEventStorage().GetEvent(context.Background(), "missing")
	if !errors.Is(err, ErrEventNotFound) {
		t.Fatalf("expected ErrEventNotFound, got %v", err)
	}
}

func TestMemoryEventStorage_ListFilterLimitOffset(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := NewMemoryEventStorage()
	now := time.Now().UTC().Truncate(time.Second)

	events := []*GitHubEvent{
		{ID: "1", Type: "push", Organization: "acme", Repository: "a", Sender: "alice", Timestamp: now.Add(-3 * time.Hour)},
		{ID: "2", Type: "pull_request", Action: "opened", Organization: "acme", Repository: "b", Sender: "bob", Timestamp: now.Add(-2 * time.Hour)},
		{ID: "3", Type: "push", Organization: "other", Repository: "c", Sender: "carol", Timestamp: now.Add(-1 * time.Hour)},
		{ID: "4", Type: "issues", Action: "opened", Organization: "acme", Repository: "a", Sender: "alice", Timestamp: now},
	}
	for _, e := range events {
		if err := store.StoreEvent(ctx, e); err != nil {
			t.Fatalf("StoreEvent: %v", err)
		}
	}

	// Org filter
	list, err := store.ListEvents(ctx, &EventFilter{Organization: "acme"}, 0, 0)
	if err != nil {
		t.Fatalf("ListEvents: %v", err)
	}
	if len(list) != 3 {
		t.Fatalf("org filter: got %d want 3", len(list))
	}

	// Type + sender
	list, err = store.ListEvents(ctx, &EventFilter{
		EventTypes: []EventType{EventTypePush},
		Sender:     "alice",
	}, 0, 0)
	if err != nil {
		t.Fatalf("ListEvents type/sender: %v", err)
	}
	if len(list) != 1 || list[0].ID != "1" {
		t.Fatalf("type/sender filter unexpected: %+v", list)
	}

	// Time range (since -90m until now+1s)
	list, err = store.ListEvents(ctx, &EventFilter{
		TimeRange: &TimeRange{Start: now.Add(-90 * time.Minute), End: now.Add(time.Second)},
	}, 0, 0)
	if err != nil {
		t.Fatalf("ListEvents time: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("time filter: got %d want 2", len(list))
	}

	// Limit / offset
	list, err = store.ListEvents(ctx, nil, 2, 1)
	if err != nil {
		t.Fatalf("ListEvents page: %v", err)
	}
	if len(list) != 2 || list[0].ID != "2" || list[1].ID != "3" {
		t.Fatalf("limit/offset unexpected: %+v", list)
	}

	count, err := store.CountEvents(ctx, &EventFilter{Organization: "acme"})
	if err != nil {
		t.Fatalf("CountEvents: %v", err)
	}
	if count != 3 {
		t.Fatalf("CountEvents: got %d want 3", count)
	}
}

func TestMemoryEventStorage_AggregateMetrics(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := NewMemoryEventStorage()

	metrics := store.AggregateMetrics(ctx)
	if metrics.TotalEventsReceived != 0 {
		t.Fatalf("empty metrics should be zero, got %d", metrics.TotalEventsReceived)
	}

	_ = store.StoreEvent(ctx, &GitHubEvent{ID: "a", Type: "push", Organization: "acme", Timestamp: time.Unix(100, 0).UTC()})
	_ = store.StoreEvent(ctx, &GitHubEvent{ID: "b", Type: "push", Organization: "acme", Timestamp: time.Unix(200, 0).UTC()})
	_ = store.StoreEvent(ctx, &GitHubEvent{ID: "c", Type: "issues", Organization: "other", Timestamp: time.Unix(150, 0).UTC()})

	metrics = store.AggregateMetrics(ctx)
	if metrics.TotalEventsReceived != 3 || metrics.TotalEventsProcessed != 3 {
		t.Fatalf("totals: %+v", metrics)
	}
	if metrics.EventsByType["push"] != 2 || metrics.EventsByType["issues"] != 1 {
		t.Fatalf("by type: %+v", metrics.EventsByType)
	}
	if metrics.EventsByOrganization["acme"] != 2 || metrics.EventsByOrganization["other"] != 1 {
		t.Fatalf("by org: %+v", metrics.EventsByOrganization)
	}
	if !metrics.LastEventAt.Equal(time.Unix(200, 0).UTC()) {
		t.Fatalf("LastEventAt: %v", metrics.LastEventAt)
	}
}

func TestMemoryEventStorage_StoreValidation(t *testing.T) {
	t.Parallel()
	store := NewMemoryEventStorage()
	ctx := context.Background()

	if err := store.StoreEvent(ctx, nil); err == nil {
		t.Fatal("expected error for nil event")
	}
	if err := store.StoreEvent(ctx, &GitHubEvent{}); err == nil {
		t.Fatal("expected error for empty ID")
	}
}
