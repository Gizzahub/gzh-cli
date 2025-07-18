package event

import (
	"context"

	"github.com/gizzahub/gzh-manager-go/pkg/github"
)

// MockStorage implements a mock event storage for testing
type MockStorage struct{}

// NewMockStorage creates a new mock storage
func NewMockStorage() *MockStorage {
	return &MockStorage{}
}

// StoreEvent stores an event (mock implementation)
func (m *MockStorage) StoreEvent(ctx context.Context, event *github.GitHubEvent) error {
	return nil
}

// GetEvent retrieves an event by ID (mock implementation)
func (m *MockStorage) GetEvent(ctx context.Context, eventID string) (*github.GitHubEvent, error) {
	return nil, nil
}

// ListEvents lists events with filtering (mock implementation)
func (m *MockStorage) ListEvents(ctx context.Context, filter *github.EventFilter, limit, offset int) ([]*github.GitHubEvent, error) {
	return []*github.GitHubEvent{}, nil
}

// DeleteEvent deletes an event (mock implementation)
func (m *MockStorage) DeleteEvent(ctx context.Context, eventID string) error {
	return nil
}

// CountEvents counts events matching filter (mock implementation)
func (m *MockStorage) CountEvents(ctx context.Context, filter *github.EventFilter) (int, error) {
	return 0, nil
}
