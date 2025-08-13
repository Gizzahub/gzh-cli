// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package event

import (
	"context"
	"fmt"

	"github.com/Gizzahub/gzh-manager-go/pkg/github"
)

// MockStorage implements a mock event storage for testing.
type MockStorage struct{}

// NewMockStorage creates a new mock storage.
func NewMockStorage() *MockStorage {
	return &MockStorage{}
}

// StoreEvent stores an event (mock implementation).
func (m *MockStorage) StoreEvent(_ context.Context, _ *github.GitHubEvent) error {
	return nil
}

// GetEvent retrieves an event by ID (mock implementation).
func (m *MockStorage) GetEvent(_ context.Context, eventID string) (*github.GitHubEvent, error) {
	return nil, fmt.Errorf("event not found: %s", eventID)
}

// ListEvents lists events with filtering (mock implementation).
func (m *MockStorage) ListEvents(_ context.Context, _ *github.EventFilter, _, _ int) ([]*github.GitHubEvent, error) {
	return []*github.GitHubEvent{}, nil
}

// DeleteEvent deletes an event (mock implementation).
func (m *MockStorage) DeleteEvent(_ context.Context, _ string) error {
	return nil
}

// CountEvents counts events matching filter (mock implementation).
func (m *MockStorage) CountEvents(_ context.Context, _ *github.EventFilter) (int, error) {
	return 0, nil
}
