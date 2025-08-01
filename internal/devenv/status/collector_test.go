// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package status

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockServiceChecker is a mock implementation of ServiceChecker.
type MockServiceChecker struct {
	mock.Mock
}

func (m *MockServiceChecker) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockServiceChecker) CheckStatus(ctx context.Context) (*ServiceStatus, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	status, ok := args.Get(0).(*ServiceStatus)
	if !ok {
		return nil, args.Error(1)
	}
	return status, args.Error(1)
}

func (m *MockServiceChecker) CheckHealth(ctx context.Context) (*HealthStatus, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	health, ok := args.Get(0).(*HealthStatus)
	if !ok {
		return nil, args.Error(1)
	}
	return health, args.Error(1)
}

func TestStatusCollector_CollectAll(t *testing.T) {
	tests := []struct {
		name     string
		parallel bool
		wantLen  int
	}{
		{
			name:     "sequential collection",
			parallel: false,
			wantLen:  2,
		},
		{
			name:     "parallel collection",
			parallel: true,
			wantLen:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock checkers
			mockChecker1 := &MockServiceChecker{}
			mockChecker2 := &MockServiceChecker{}

			mockChecker1.On("Name").Return("aws").Maybe()
			mockChecker1.On("CheckStatus", mock.Anything).Return(&ServiceStatus{
				Name:   "aws",
				Status: StatusActive,
			}, nil)

			mockChecker2.On("Name").Return("gcp").Maybe()
			mockChecker2.On("CheckStatus", mock.Anything).Return(&ServiceStatus{
				Name:   "gcp",
				Status: StatusInactive,
			}, nil)

			checkers := []ServiceChecker{mockChecker1, mockChecker2}
			collector := NewStatusCollector(checkers, 30*time.Second)

			options := StatusOptions{
				Parallel: tt.parallel,
			}

			statuses, err := collector.CollectAll(context.Background(), options)

			require.NoError(t, err)
			assert.Len(t, statuses, tt.wantLen)

			// Verify mock expectations
			mockChecker1.AssertExpectations(t)
			mockChecker2.AssertExpectations(t)
		})
	}
}

func TestStatusCollector_CollectAll_WithHealthCheck(t *testing.T) {
	mockChecker := &MockServiceChecker{}

	mockChecker.On("Name").Return("aws").Maybe()
	mockChecker.On("CheckStatus", mock.Anything).Return(&ServiceStatus{
		Name:   "aws",
		Status: StatusActive,
	}, nil)
	mockChecker.On("CheckHealth", mock.Anything).Return(&HealthStatus{
		Status:  StatusActive,
		Message: "All good",
	}, nil)

	checkers := []ServiceChecker{mockChecker}
	collector := NewStatusCollector(checkers, 30*time.Second)

	options := StatusOptions{
		CheckHealth: true,
		Parallel:    false,
	}

	statuses, err := collector.CollectAll(context.Background(), options)

	require.NoError(t, err)
	assert.Len(t, statuses, 1)
	assert.NotNil(t, statuses[0].HealthCheck)
	assert.Equal(t, "All good", statuses[0].HealthCheck.Message)

	mockChecker.AssertExpectations(t)
}

func TestStatusCollector_filterCheckers(t *testing.T) {
	mockChecker1 := &MockServiceChecker{}
	mockChecker2 := &MockServiceChecker{}

	mockChecker1.On("Name").Return("aws")
	mockChecker2.On("Name").Return("gcp")

	checkers := []ServiceChecker{mockChecker1, mockChecker2}
	collector := NewStatusCollector(checkers, 30*time.Second)

	tests := []struct {
		name              string
		requestedServices []string
		expectedLen       int
	}{
		{
			name:              "all services",
			requestedServices: nil,
			expectedLen:       2,
		},
		{
			name:              "specific service",
			requestedServices: []string{"aws"},
			expectedLen:       1,
		},
		{
			name:              "nonexistent service",
			requestedServices: []string{"nonexistent"},
			expectedLen:       0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := collector.filterCheckers(tt.requestedServices)
			assert.Len(t, filtered, tt.expectedLen)
		})
	}
}
