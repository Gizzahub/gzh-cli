package automation

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-github/v66/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// MockActionHandler for testing
type MockActionHandler struct {
	ExecuteFunc            func(ctx context.Context, event *Event, action Action) error
	ValidateParametersFunc func(params map[string]interface{}) error
	ExecutedActions        []Action
}

func (m *MockActionHandler) Execute(ctx context.Context, event *Event, action Action) error {
	m.ExecutedActions = append(m.ExecutedActions, action)
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, event, action)
	}
	return nil
}

func (m *MockActionHandler) ValidateParameters(params map[string]interface{}) error {
	if m.ValidateParametersFunc != nil {
		return m.ValidateParametersFunc(params)
	}
	return nil
}

func TestNewEngine(t *testing.T) {
	logger := zaptest.NewLogger(t)
	engine := NewEngine(nil, logger)

	assert.NotNil(t, engine)
	assert.Equal(t, logger, engine.logger)
	assert.NotNil(t, engine.handlers)
	assert.NotNil(t, engine.metrics)
	assert.Empty(t, engine.rules)
}

func TestEngineRegisterHandler(t *testing.T) {
	logger := zaptest.NewLogger(t)
	engine := NewEngine(nil, logger)
	handler := &MockActionHandler{}

	// Test successful registration
	err := engine.RegisterHandler("test_action", handler)
	assert.NoError(t, err)

	// Test duplicate registration
	err = engine.RegisterHandler("test_action", handler)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")
}

func TestEngineAddRule(t *testing.T) {
	logger := zaptest.NewLogger(t)
	engine := NewEngine(nil, logger)

	tests := []struct {
		name    string
		rule    Rule
		wantErr bool
	}{
		{
			name: "valid rule",
			rule: Rule{
				ID:          "test-rule",
				Name:        "Test Rule",
				Description: "A test rule",
				Enabled:     true,
				Priority:    100,
				Conditions: []Condition{
					{
						Type:     "event_type",
						Operator: "equals",
						Value:    "push",
					},
				},
				Actions: []Action{
					{
						Type:       "test_action",
						Parameters: map[string]interface{}{"key": "value"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			rule: Rule{
				Name:        "Test Rule",
				Description: "A test rule",
				Enabled:     true,
				Priority:    100,
				Conditions: []Condition{
					{
						Type:     "event_type",
						Operator: "equals",
						Value:    "push",
					},
				},
				Actions: []Action{
					{
						Type:       "test_action",
						Parameters: map[string]interface{}{"key": "value"},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "missing conditions",
			rule: Rule{
				ID:          "test-rule-2",
				Name:        "Test Rule 2",
				Description: "A test rule",
				Enabled:     true,
				Priority:    100,
				Conditions:  []Condition{},
				Actions: []Action{
					{
						Type:       "test_action",
						Parameters: map[string]interface{}{"key": "value"},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "missing actions",
			rule: Rule{
				ID:          "test-rule-3",
				Name:        "Test Rule 3",
				Description: "A test rule",
				Enabled:     true,
				Priority:    100,
				Conditions: []Condition{
					{
						Type:     "event_type",
						Operator: "equals",
						Value:    "push",
					},
				},
				Actions: []Action{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := engine.AddRule(tt.rule)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEngineProcessEvent(t *testing.T) {
	logger := zaptest.NewLogger(t)
	engine := NewEngine(nil, logger)

	// Register a mock handler
	handler := &MockActionHandler{}
	err := engine.RegisterHandler("test_action", handler)
	require.NoError(t, err)

	// Add a rule
	rule := Rule{
		ID:          "test-rule",
		Name:        "Test Rule",
		Description: "A test rule",
		Enabled:     true,
		Priority:    100,
		Conditions: []Condition{
			{
				Type:     "event_type",
				Operator: "equals",
				Value:    "push",
			},
		},
		Actions: []Action{
			{
				Type:       "test_action",
				Parameters: map[string]interface{}{"key": "value"},
			},
		},
	}
	err = engine.AddRule(rule)
	require.NoError(t, err)

	// Create test event
	repo := &github.Repository{
		Name:     github.String("test-repo"),
		FullName: github.String("owner/test-repo"),
	}
	sender := &github.User{
		Login: github.String("testuser"),
	}

	event := &Event{
		ID:         "test-event-1",
		Type:       "push",
		Repository: repo,
		Sender:     sender,
		ReceivedAt: time.Now(),
	}

	// Process event
	ctx := context.Background()
	err = engine.ProcessEvent(ctx, event)
	assert.NoError(t, err)

	// Verify action was executed
	assert.Len(t, handler.ExecutedActions, 1)
	assert.Equal(t, "test_action", handler.ExecutedActions[0].Type)

	// Check metrics
	metrics := engine.GetMetrics()
	assert.Equal(t, int64(1), metrics.EventsProcessed)
	assert.Equal(t, int64(1), metrics.RulesEvaluated)
	assert.Equal(t, int64(1), metrics.ActionsExecuted)
	assert.Equal(t, int64(0), metrics.Errors)
}

func TestEvaluateEventTypeCondition(t *testing.T) {
	logger := zaptest.NewLogger(t)
	engine := NewEngine(nil, logger)

	tests := []struct {
		name      string
		condition Condition
		event     *Event
		expected  bool
	}{
		{
			name: "equals - match",
			condition: Condition{
				Type:     "event_type",
				Operator: "equals",
				Value:    "push",
			},
			event: &Event{
				Type: "push",
			},
			expected: true,
		},
		{
			name: "equals - no match",
			condition: Condition{
				Type:     "event_type",
				Operator: "equals",
				Value:    "push",
			},
			event: &Event{
				Type: "pull_request",
			},
			expected: false,
		},
		{
			name: "equals with action - match",
			condition: Condition{
				Type:     "event_type",
				Operator: "equals",
				Value:    "pull_request.opened",
			},
			event: &Event{
				Type:   "pull_request",
				Action: "opened",
			},
			expected: true,
		},
		{
			name: "in operator - match",
			condition: Condition{
				Type:     "event_type",
				Operator: "in",
				Value:    []interface{}{"push", "pull_request", "issues"},
			},
			event: &Event{
				Type: "push",
			},
			expected: true,
		},
		{
			name: "in operator - no match",
			condition: Condition{
				Type:     "event_type",
				Operator: "in",
				Value:    []interface{}{"push", "pull_request"},
			},
			event: &Event{
				Type: "issues",
			},
			expected: false,
		},
		{
			name: "matches operator - match",
			condition: Condition{
				Type:     "event_type",
				Operator: "matches",
				Value:    "pull_request.*",
			},
			event: &Event{
				Type:   "pull_request",
				Action: "opened",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.evaluateEventTypeCondition(tt.condition, tt.event)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEvaluateRepositoryCondition(t *testing.T) {
	logger := zaptest.NewLogger(t)
	engine := NewEngine(nil, logger)

	repo := &github.Repository{
		Name:     github.String("test-repo"),
		FullName: github.String("owner/test-repo"),
		Private:  github.Bool(false),
		Language: github.String("Go"),
	}

	tests := []struct {
		name      string
		condition Condition
		event     *Event
		expected  bool
	}{
		{
			name: "repository name equals - match",
			condition: Condition{
				Type:     "repository",
				Field:    "name",
				Operator: "equals",
				Value:    "test-repo",
			},
			event: &Event{
				Repository: repo,
			},
			expected: true,
		},
		{
			name: "repository name equals - no match",
			condition: Condition{
				Type:     "repository",
				Field:    "name",
				Operator: "equals",
				Value:    "other-repo",
			},
			event: &Event{
				Repository: repo,
			},
			expected: false,
		},
		{
			name: "repository private - match",
			condition: Condition{
				Type:     "repository",
				Field:    "private",
				Operator: "equals",
				Value:    false,
			},
			event: &Event{
				Repository: repo,
			},
			expected: true,
		},
		{
			name: "repository language - match",
			condition: Condition{
				Type:     "repository",
				Field:    "language",
				Operator: "equals",
				Value:    "Go",
			},
			event: &Event{
				Repository: repo,
			},
			expected: true,
		},
		{
			name: "no repository - no match",
			condition: Condition{
				Type:     "repository",
				Field:    "name",
				Operator: "equals",
				Value:    "test-repo",
			},
			event: &Event{
				Repository: nil,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.evaluateRepositoryCondition(tt.condition, tt.event)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEvaluateSenderCondition(t *testing.T) {
	logger := zaptest.NewLogger(t)
	engine := NewEngine(nil, logger)

	sender := &github.User{
		Login:     github.String("testuser"),
		Type:      github.String("User"),
		SiteAdmin: github.Bool(false),
	}

	tests := []struct {
		name      string
		condition Condition
		event     *Event
		expected  bool
	}{
		{
			name: "sender login equals - match",
			condition: Condition{
				Type:     "sender",
				Field:    "login",
				Operator: "equals",
				Value:    "testuser",
			},
			event: &Event{
				Sender: sender,
			},
			expected: true,
		},
		{
			name: "sender type equals - match",
			condition: Condition{
				Type:     "sender",
				Field:    "type",
				Operator: "equals",
				Value:    "User",
			},
			event: &Event{
				Sender: sender,
			},
			expected: true,
		},
		{
			name: "sender site_admin - match",
			condition: Condition{
				Type:     "sender",
				Field:    "site_admin",
				Operator: "equals",
				Value:    false,
			},
			event: &Event{
				Sender: sender,
			},
			expected: true,
		},
		{
			name: "no sender - no match",
			condition: Condition{
				Type:     "sender",
				Field:    "login",
				Operator: "equals",
				Value:    "testuser",
			},
			event: &Event{
				Sender: nil,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.evaluateSenderCondition(tt.condition, tt.event)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEvaluateStringCondition(t *testing.T) {
	logger := zaptest.NewLogger(t)
	engine := NewEngine(nil, logger)

	tests := []struct {
		name      string
		condition Condition
		value     string
		expected  bool
	}{
		{
			name: "equals - match",
			condition: Condition{
				Operator: "equals",
				Value:    "test",
			},
			value:    "test",
			expected: true,
		},
		{
			name: "not_equals - match",
			condition: Condition{
				Operator: "not_equals",
				Value:    "test",
			},
			value:    "other",
			expected: true,
		},
		{
			name: "contains - match",
			condition: Condition{
				Operator: "contains",
				Value:    "est",
			},
			value:    "test",
			expected: true,
		},
		{
			name: "starts_with - match",
			condition: Condition{
				Operator: "starts_with",
				Value:    "te",
			},
			value:    "test",
			expected: true,
		},
		{
			name: "ends_with - match",
			condition: Condition{
				Operator: "ends_with",
				Value:    "st",
			},
			value:    "test",
			expected: true,
		},
		{
			name: "matches - match",
			condition: Condition{
				Operator: "matches",
				Value:    "t.*t",
			},
			value:    "test",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.evaluateStringCondition(tt.condition, tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAsyncActionExecution(t *testing.T) {
	logger := zaptest.NewLogger(t)
	engine := NewEngine(nil, logger)

	// Register a mock handler that takes some time
	handler := &MockActionHandler{
		ExecuteFunc: func(ctx context.Context, event *Event, action Action) error {
			time.Sleep(50 * time.Millisecond)
			return nil
		},
	}
	err := engine.RegisterHandler("slow_action", handler)
	require.NoError(t, err)

	// Add a rule with async action
	rule := Rule{
		ID:          "async-test-rule",
		Name:        "Async Test Rule",
		Description: "A test rule with async action",
		Enabled:     true,
		Priority:    100,
		Conditions: []Condition{
			{
				Type:     "event_type",
				Operator: "equals",
				Value:    "push",
			},
		},
		Actions: []Action{
			{
				Type:       "slow_action",
				Parameters: map[string]interface{}{"key": "value"},
				Async:      true,
			},
		},
	}
	err = engine.AddRule(rule)
	require.NoError(t, err)

	// Create test event
	event := &Event{
		ID:         "async-test-event",
		Type:       "push",
		ReceivedAt: time.Now(),
	}

	// Process event (should return quickly due to async execution)
	ctx := context.Background()
	start := time.Now()
	err = engine.ProcessEvent(ctx, event)
	duration := time.Since(start)

	assert.NoError(t, err)
	assert.Less(t, duration, 40*time.Millisecond, "Async action should not block")

	// Wait for async action to complete
	time.Sleep(100 * time.Millisecond)

	// Check metrics
	metrics := engine.GetMetrics()
	assert.Equal(t, int64(1), metrics.EventsProcessed)
	assert.Equal(t, int64(1), metrics.ActionsExecuted)
}

func TestEngineMetrics(t *testing.T) {
	logger := zaptest.NewLogger(t)
	engine := NewEngine(nil, logger)

	// Get initial metrics
	metrics := engine.GetMetrics()
	assert.Equal(t, int64(0), metrics.EventsProcessed)
	assert.Equal(t, int64(0), metrics.RulesEvaluated)
	assert.Equal(t, int64(0), metrics.ActionsExecuted)
	assert.Equal(t, int64(0), metrics.Errors)

	// Register handler and add rule
	handler := &MockActionHandler{}
	err := engine.RegisterHandler("test_action", handler)
	require.NoError(t, err)

	rule := Rule{
		ID:          "metrics-test-rule",
		Name:        "Metrics Test Rule",
		Description: "A rule for testing metrics",
		Enabled:     true,
		Priority:    100,
		Conditions: []Condition{
			{
				Type:     "event_type",
				Operator: "equals",
				Value:    "push",
			},
		},
		Actions: []Action{
			{
				Type:       "test_action",
				Parameters: map[string]interface{}{"key": "value"},
			},
		},
	}
	err = engine.AddRule(rule)
	require.NoError(t, err)

	// Process an event
	event := &Event{
		ID:         "metrics-test-event",
		Type:       "push",
		ReceivedAt: time.Now(),
	}

	ctx := context.Background()
	err = engine.ProcessEvent(ctx, event)
	assert.NoError(t, err)

	// Check updated metrics
	metrics = engine.GetMetrics()
	assert.Equal(t, int64(1), metrics.EventsProcessed)
	assert.Equal(t, int64(1), metrics.RulesEvaluated)
	assert.Equal(t, int64(1), metrics.ActionsExecuted)
	assert.Equal(t, int64(0), metrics.Errors)
	assert.Greater(t, metrics.ProcessingTime, time.Duration(0))
}
