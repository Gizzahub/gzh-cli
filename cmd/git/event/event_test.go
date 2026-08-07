// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

//nolint:testpackage // white-box tests for command helpers
package event

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gizzahub/gzh-cli/pkg/github"
)

func withIsolatedStorage(t *testing.T) *github.MemoryEventStorage {
	t.Helper()
	store := github.NewMemoryEventStorage()
	prev := DefaultEventStorage()
	SetDefaultEventStorage(store)
	t.Cleanup(func() {
		SetDefaultEventStorage(prev)
	})
	return store
}

func captureStdout(t *testing.T, fn func() error) (string, error) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	runErr := fn()
	_ = w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	_ = r.Close()
	return buf.String(), runErr
}

func TestRunEventListEmpty(t *testing.T) {
	_ = withIsolatedStorage(t)

	out, err := captureStdout(t, func() error {
		return runEventList(nil, nil, "", "", "", "", "", "", "", 50, 0, "table")
	})
	if err != nil {
		t.Fatalf("runEventList: %v", err)
	}
	if !strings.Contains(out, "No events stored") {
		t.Fatalf("expected empty message, got %q", out)
	}
	// Must not fabricate demo data.
	if strings.Contains(out, "testorg") || strings.Contains(out, "event-1") {
		t.Fatalf("fabricated data present: %q", out)
	}
}

func TestRunEventListAndGetWithStorage(t *testing.T) {
	store := withIsolatedStorage(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	_ = store.StoreEvent(ctx, &github.GitHubEvent{
		ID: "evt-push-1", Type: "push", Action: "created",
		Organization: "acme", Repository: "api", Sender: "alice",
		Timestamp: now,
	})
	_ = store.StoreEvent(ctx, &github.GitHubEvent{
		ID: "evt-pr-1", Type: "pull_request", Action: "opened",
		Organization: "acme", Repository: "web", Sender: "bob",
		Timestamp: now.Add(-time.Hour),
	})

	out, err := captureStdout(t, func() error {
		return runEventList(nil, nil, "acme", "", "push", "", "", "", "", 50, 0, "table")
	})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if !strings.Contains(out, "evt-push-1") {
		t.Fatalf("expected evt-push-1 in output: %q", out)
	}
	if strings.Contains(out, "evt-pr-1") {
		t.Fatalf("filter failed, pr event present: %q", out)
	}

	out, err = captureStdout(t, func() error {
		return runEventGet(nil, []string{"evt-push-1"}, "json")
	})
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if !strings.Contains(out, `"id": "evt-push-1"`) {
		t.Fatalf("get output missing id: %q", out)
	}

	err = runEventGet(nil, []string{"missing-id"}, "json")
	if !errors.Is(err, github.ErrEventNotFound) {
		t.Fatalf("expected ErrEventNotFound, got %v", err)
	}
}

func TestRunEventMetricsFromStorage(t *testing.T) {
	store := withIsolatedStorage(t)
	ctx := context.Background()
	_ = store.StoreEvent(ctx, &github.GitHubEvent{ID: "1", Type: "push", Organization: "acme", Timestamp: time.Now()})
	_ = store.StoreEvent(ctx, &github.GitHubEvent{ID: "2", Type: "issues", Organization: "acme", Timestamp: time.Now()})

	out, err := captureStdout(t, func() error {
		return runEventMetrics(nil, nil, "table")
	})
	if err != nil {
		t.Fatalf("metrics: %v", err)
	}
	if !strings.Contains(out, "Total Events Received:  2") {
		t.Fatalf("unexpected metrics: %q", out)
	}
	// Must not show fabricated 1250 sample metrics.
	if strings.Contains(out, "1250") {
		t.Fatalf("fabricated metrics present: %q", out)
	}
}

func TestBuildEventFilterTimeValidation(t *testing.T) {
	t.Parallel()
	_, err := buildEventFilter("", "", "", "", "", "not-a-time", "")
	if err == nil {
		t.Fatal("expected since parse error")
	}
	_, err = buildEventFilter("", "", "", "", "", "", "also-bad")
	if err == nil {
		t.Fatal("expected until parse error")
	}

	f, err := buildEventFilter("o", "r", "push", "opened", "s",
		"2020-01-01T00:00:00Z", "2020-01-02T00:00:00Z")
	if err != nil {
		t.Fatalf("build filter: %v", err)
	}
	if f.Organization != "o" || len(f.EventTypes) != 1 || f.TimeRange == nil {
		t.Fatalf("unexpected filter: %+v", f)
	}
}

func TestServerUsesDefaultStorage(t *testing.T) {
	// Smoke: processor path stores into default storage without network.
	store := withIsolatedStorage(t)
	logger := getLogger()
	processor := github.NewEventProcessor(DefaultEventStorage(), logger)

	event := &github.GitHubEvent{
		ID: "server-evt", Type: "push", Organization: "acme", Repository: "r",
		Sender: "u", Timestamp: time.Now().UTC(),
	}
	if err := processor.ProcessEvent(context.Background(), event); err != nil {
		t.Fatalf("ProcessEvent: %v", err)
	}

	got, err := store.GetEvent(context.Background(), "server-evt")
	if err != nil {
		t.Fatalf("GetEvent after process: %v", err)
	}
	if got.ID != "server-evt" {
		t.Fatalf("unexpected event: %+v", got)
	}
}
