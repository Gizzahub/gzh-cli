package github

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mock implementations for testing

type mockAPIClient struct {
	mock.Mock
}

func (m *mockAPIClient) GetRepository(ctx context.Context, owner, repo string) (*RepositoryInfo, error) {
	args := m.Called(ctx, owner, repo)
	if repo, ok := args.Get(0).(*RepositoryInfo); ok {
		return repo, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAPIClient) ListOrganizationRepositories(ctx context.Context, org string) ([]RepositoryInfo, error) {
	args := m.Called(ctx, org)
	if repos, ok := args.Get(0).([]RepositoryInfo); ok {
		return repos, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAPIClient) GetDefaultBranch(ctx context.Context, owner, repo string) (string, error) {
	args := m.Called(ctx, owner, repo)
	return args.String(0), args.Error(1)
}

func (m *mockAPIClient) SetToken(token string) {
	m.Called(token)
}

func (m *mockAPIClient) GetRateLimit(ctx context.Context) (*RateLimit, error) {
	args := m.Called(ctx)
	if rateLimit, ok := args.Get(0).(*RateLimit); ok {
		return rateLimit, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAPIClient) GetRepositoryConfiguration(ctx context.Context, owner, repo string) (*RepositoryConfig, error) {
	args := m.Called(ctx, owner, repo)
	if config, ok := args.Get(0).(*RepositoryConfig); ok {
		return config, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAPIClient) UpdateRepositoryConfiguration(ctx context.Context, owner, repo string, config *RepositoryConfig) error {
	args := m.Called(ctx, owner, repo, config)
	return args.Error(0)
}

type mockEventProcessor struct {
	mock.Mock
}

func (m *mockEventProcessor) ProcessEvent(ctx context.Context, event *GitHubEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *mockEventProcessor) FilterEvent(ctx context.Context, event *GitHubEvent, filter *EventFilter) (bool, error) {
	args := m.Called(ctx, event, filter)
	return args.Bool(0), args.Error(1)
}

func (m *mockEventProcessor) ValidateEvent(ctx context.Context, event *GitHubEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *mockEventProcessor) ValidateSignature(payload []byte, signature, secret string) bool {
	args := m.Called(payload, signature, secret)
	return args.Bool(0)
}

func (m *mockEventProcessor) ParseWebhookEvent(r *http.Request) (*GitHubEvent, error) {
	args := m.Called(r)
	if event, ok := args.Get(0).(*GitHubEvent); ok {
		return event, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockEventProcessor) RegisterEventHandler(eventType EventType, handler EventHandler) error {
	args := m.Called(eventType, handler)
	return args.Error(0)
}

func (m *mockEventProcessor) UnregisterEventHandler(eventType EventType) error {
	args := m.Called(eventType)
	return args.Error(0)
}

func (m *mockEventProcessor) GetMetrics() *EventMetrics {
	args := m.Called()
	if metrics, ok := args.Get(0).(*EventMetrics); ok {
		return metrics
	}
	return nil
}

type mockLogger struct {
	mock.Mock
}

func (m *mockLogger) Debug(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *mockLogger) Info(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *mockLogger) Warn(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *mockLogger) Error(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *mockLogger) Fatal(msg string, args ...interface{}) {
	m.Called(msg, args)
}

type mockConditionEvaluator struct {
	mock.Mock
}

func (m *mockConditionEvaluator) EvaluateConditions(ctx context.Context, conditions *AutomationConditions, event *GitHubEvent, context *EvaluationContext) (*EvaluationResult, error) {
	args := m.Called(ctx, conditions, event, context)
	if result := args.Get(0); result != nil {
		if evalResult, ok := result.(*EvaluationResult); ok {
			return evalResult, args.Error(1)
		}
	}

	return nil, args.Error(1)
}

func (m *mockConditionEvaluator) EvaluatePayloadMatcher(ctx context.Context, matcher *PayloadMatcher, payload map[string]interface{}) (bool, error) {
	args := m.Called(ctx, matcher, payload)
	return args.Bool(0), args.Error(1)
}

func (m *mockConditionEvaluator) EvaluateEventConditions(event *GitHubEvent, conditions *AutomationConditions) (bool, error) {
	args := m.Called(event, conditions)
	return args.Bool(0), args.Error(1)
}

func (m *mockConditionEvaluator) EvaluateRepositoryConditions(ctx context.Context, repoInfo *RepositoryInfo, conditions *AutomationConditions) (bool, error) {
	args := m.Called(ctx, repoInfo, conditions)
	return args.Bool(0), args.Error(1)
}

func (m *mockConditionEvaluator) EvaluateTimeConditions(timestamp time.Time, conditions *AutomationConditions) (bool, error) {
	args := m.Called(timestamp, conditions)
	return args.Bool(0), args.Error(1)
}

func (m *mockConditionEvaluator) EvaluateContentConditions(ctx context.Context, event *GitHubEvent, conditions *AutomationConditions) (bool, error) {
	args := m.Called(ctx, event, conditions)
	return args.Bool(0), args.Error(1)
}

func (m *mockConditionEvaluator) ValidateConditions(conditions *AutomationConditions) (*ConditionValidationResult, error) {
	args := m.Called(conditions)
	if result := args.Get(0); result != nil {
		if validationResult, ok := result.(*ConditionValidationResult); ok {
			return validationResult, args.Error(1)
		}
	}

	return nil, args.Error(1)
}

func (m *mockConditionEvaluator) ExplainEvaluation(ctx context.Context, conditions *AutomationConditions, event *GitHubEvent) (*EvaluationExplanation, error) {
	args := m.Called(ctx, conditions, event)
	if result := args.Get(0); result != nil {
		if explanation, ok := result.(*EvaluationExplanation); ok {
			return explanation, args.Error(1)
		}
	}

	return nil, args.Error(1)
}

type mockActionExecutor struct {
	mock.Mock
}

func (m *mockActionExecutor) ExecuteAction(ctx context.Context, action *AutomationAction, execContext *AutomationExecutionContext) (*ActionExecutionResult, error) {
	args := m.Called(ctx, action, execContext)
	if result := args.Get(0); result != nil {
		if actionResult, ok := result.(*ActionExecutionResult); ok {
			return actionResult, args.Error(1)
		}
	}

	return nil, args.Error(1)
}

func (m *mockActionExecutor) ValidateAction(ctx context.Context, action *AutomationAction) error {
	args := m.Called(ctx, action)
	return args.Error(0)
}

func (m *mockActionExecutor) GetSupportedActions() []ActionType {
	args := m.Called()
	if actions, ok := args.Get(0).([]ActionType); ok {
		return actions
	}
	return nil
}

// Test helper functions

func createTestAutomationEngine() (*AutomationEngine, *mockEventProcessor) {
	logger := &mockLogger{}
	apiClient := &mockAPIClient{}

	// Create a real RuleManager for testing since it's expected as a struct pointer
	ruleManager := &RuleManager{}

	conditionEvaluator := &mockConditionEvaluator{}
	actionExecutor := &mockActionExecutor{}
	eventProcessor := &mockEventProcessor{}

	config := &AutomationEngineConfig{
		MaxWorkers:           2,
		EventBufferSize:      10,
		ExecutionTimeout:     30 * time.Second,
		EnableAsyncExecution: true,
		EnableRuleFiltering:  true,
		EnableMetrics:        true,
		MaxRetries:           2,
		RetryBackoffFactor:   1.5,
	}

	engine := NewAutomationEngine(
		logger,
		apiClient,
		ruleManager,
		conditionEvaluator,
		actionExecutor,
		eventProcessor,
		config,
	)

	return engine, eventProcessor
}

func createTestEngineEvent() *GitHubEvent {
	return &GitHubEvent{
		ID:           "test-event-001",
		Type:         "pull_request",
		Action:       "opened",
		Organization: "testorg",
		Repository:   "test-repo",
		Sender:       "test-user",
		Timestamp:    time.Now(),
		Payload: map[string]interface{}{
			"action": "opened",
			"pull_request": map[string]interface{}{
				"title":  "Test PR",
				"number": 1,
			},
		},
	}
}

// Test Cases

func TestNewAutomationEngine(t *testing.T) {
	engine, _ := createTestAutomationEngine()

	assert.NotNil(t, engine)
	assert.NotNil(t, engine.config)
	assert.NotNil(t, engine.metrics)
	assert.NotNil(t, engine.eventChannel)
	assert.NotNil(t, engine.executionChannel)
	assert.False(t, engine.running)
}

func TestNewAutomationEngine_WithNilConfig(t *testing.T) {
	logger := &mockLogger{}
	apiClient := &mockAPIClient{}
	conditionEvaluator := &mockConditionEvaluator{}
	actionExecutor := &mockActionExecutor{}
	eventProcessor := &mockEventProcessor{}

	// Create a real RuleManager since NewAutomationEngine expects *RuleManager
	ruleManager := &RuleManager{
		logger:         logger,
		apiClient:      apiClient,
		evaluator:      conditionEvaluator,
		actionExecutor: actionExecutor,
	}

	engine := NewAutomationEngine(
		logger,
		apiClient,
		ruleManager,
		conditionEvaluator,
		actionExecutor,
		eventProcessor,
		nil, // nil config should use defaults
	)

	assert.NotNil(t, engine)
	assert.NotNil(t, engine.config)
	assert.Equal(t, 10, engine.config.MaxWorkers)
	assert.Equal(t, 1000, engine.config.EventBufferSize)
}

func TestAutomationEngine_Start(t *testing.T) {
	engine, _ := createTestAutomationEngine()
	ctx := context.Background()

	err := engine.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, engine.isRunning())
	assert.False(t, engine.metrics.StartTime.IsZero())

	// Try to start again - should fail
	err = engine.Start(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already running")

	// Stop the engine
	err = engine.Stop(ctx)
	assert.NoError(t, err)
}

func TestAutomationEngine_Stop(t *testing.T) {
	engine, _ := createTestAutomationEngine()
	ctx := context.Background()

	// Try to stop when not running - should fail
	err := engine.Stop(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")

	// Start and then stop
	err = engine.Start(ctx)
	assert.NoError(t, err)

	err = engine.Stop(ctx)
	assert.NoError(t, err)
	assert.False(t, engine.isRunning())
}

func TestAutomationEngine_ProcessEvent_NotRunning(t *testing.T) {
	engine, _ := createTestAutomationEngine()
	event := createTestEngineEvent()

	err := engine.ProcessEvent(context.Background(), event)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")
}

func TestAutomationEngine_ProcessEvent_ValidationFailed(t *testing.T) {
	engine, eventProcessor := createTestAutomationEngine()
	event := createTestEngineEvent()
	ctx := context.Background()

	eventProcessor.On("ValidateEvent", mock.Anything, event).Return(assert.AnError)

	err := engine.Start(ctx)
	require.NoError(t, err)

	defer func() {
		if err := engine.Stop(ctx); err != nil {
			t.Logf("Failed to stop engine: %v", err)
		}
	}()

	err = engine.ProcessEvent(ctx, event)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "event validation failed")

	eventProcessor.AssertExpectations(t)
}

func TestAutomationEngine_ProcessEvent_Filtered(t *testing.T) {
	engine, eventProcessor := createTestAutomationEngine()
	event := createTestEngineEvent()
	ctx := context.Background()

	eventProcessor.On("ValidateEvent", mock.Anything, event).Return(nil)
	eventProcessor.On("FilterEvent", event).Return(false) // Event is filtered out

	err := engine.Start(ctx)
	require.NoError(t, err)

	defer func() {
		if err := engine.Stop(ctx); err != nil {
			t.Logf("Failed to stop engine: %v", err)
		}
	}()

	err = engine.ProcessEvent(ctx, event)
	assert.NoError(t, err)

	// Wait a bit to ensure processing
	time.Sleep(100 * time.Millisecond)

	metrics := engine.GetMetrics()
	assert.Equal(t, int64(1), metrics.EventsProcessed)

	eventProcessor.AssertExpectations(t)
}

func TestAutomationEngine_ProcessEvent_ExcludedEventType(t *testing.T) {
	engine, eventProcessor := createTestAutomationEngine()
	engine.config.ExcludedEventTypes = []EventType{EventTypePullRequest}
	event := createTestEngineEvent()
	ctx := context.Background()

	eventProcessor.On("ValidateEvent", mock.Anything, event).Return(nil)

	err := engine.Start(ctx)
	require.NoError(t, err)

	defer func() {
		if err := engine.Stop(ctx); err != nil {
			t.Logf("Failed to stop engine: %v", err)
		}
	}()

	err = engine.ProcessEvent(ctx, event)
	assert.NoError(t, err)

	// Wait a bit to ensure processing
	time.Sleep(100 * time.Millisecond)

	metrics := engine.GetMetrics()
	assert.Equal(t, int64(1), metrics.EventsProcessed)

	eventProcessor.AssertExpectations(t)
}

func TestAutomationEngine_ProcessEvent_Success(t *testing.T) {
	engine, eventProcessor := createTestAutomationEngine()
	event := createTestEngineEvent()
	// rule := createTestEngineRule() // TODO: Fix RuleManager mocking
	ctx := context.Background()

	eventProcessor.On("ValidateEvent", mock.Anything, event).Return(nil)
	eventProcessor.On("FilterEvent", event).Return(true)
	// TODO: Fix RuleManager mocking - ruleManager.On("ListRules", ctx, "testorg", mock.AnythingOfType("*github.RuleFilter")).Return([]*AutomationRule{rule}, nil)
	// TODO: Fix RuleManager mocking - ruleManager.On("EvaluateConditions", ctx, rule, event).Return(true, nil)

	// 	execution := &AutomationRuleExecution{
	// 		ID:     "exec-001",
	// 		RuleID: rule.ID,
	// 		Status: ExecutionStatusCompleted,
	// 	}
	// 	// TODO: Fix RuleManager mocking - ruleManager.On("ExecuteRule", ctx, rule, mock.AnythingOfType("*github.AutomationExecutionContext")).Return(execution, nil)

	err := engine.Start(ctx)
	require.NoError(t, err)

	defer func() {
		if err := engine.Stop(ctx); err != nil {
			t.Logf("Failed to stop engine: %v", err)
		}
	}()

	err = engine.ProcessEvent(ctx, event)
	assert.NoError(t, err)

	// Wait for processing to complete
	time.Sleep(500 * time.Millisecond)

	metrics := engine.GetMetrics()
	assert.Equal(t, int64(1), metrics.EventsProcessed)
	assert.Greater(t, metrics.RulesEvaluated, int64(0))

	eventProcessor.AssertExpectations(t)
	// TODO: Fix RuleManager mocking - ruleManager.AssertExpectations(t)
}

func TestAutomationEngine_ProcessEvent_NoMatchingRules(t *testing.T) {
	engine, eventProcessor := createTestAutomationEngine()
	event := createTestEngineEvent()
	// 	rule := createTestEngineRule()
	ctx := context.Background()

	eventProcessor.On("ValidateEvent", mock.Anything, event).Return(nil)
	eventProcessor.On("FilterEvent", event).Return(true)
	// TODO: Fix RuleManager mocking - ruleManager.On("ListRules", ctx, "testorg", mock.AnythingOfType("*github.RuleFilter")).Return([]*AutomationRule{rule}, nil)
	// TODO: Fix RuleManager mocking - ruleManager.On("EvaluateConditions", ctx, rule, event).Return(false, nil) // Conditions don't match

	err := engine.Start(ctx)
	require.NoError(t, err)

	defer func() {
		if err := engine.Stop(ctx); err != nil {
			t.Logf("Failed to stop engine: %v", err)
		}
	}()

	err = engine.ProcessEvent(ctx, event)
	assert.NoError(t, err)

	// Wait for processing
	time.Sleep(200 * time.Millisecond)

	metrics := engine.GetMetrics()
	assert.Equal(t, int64(1), metrics.EventsProcessed)
	assert.Greater(t, metrics.RulesEvaluated, int64(0))
	assert.Equal(t, int64(0), metrics.RulesExecuted) // No rules executed

	eventProcessor.AssertExpectations(t)
	// TODO: Fix RuleManager mocking - ruleManager.AssertExpectations(t)
}

func TestAutomationEngine_ProcessEvent_ExecutionFailure(t *testing.T) {
	engine, eventProcessor := createTestAutomationEngine()
	event := createTestEngineEvent()
	// 	rule := createTestEngineRule()
	ctx := context.Background()

	eventProcessor.On("ValidateEvent", mock.Anything, event).Return(nil)
	eventProcessor.On("FilterEvent", event).Return(true)
	// TODO: Fix RuleManager mocking - ruleManager.On("ListRules", ctx, "testorg", mock.AnythingOfType("*github.RuleFilter")).Return([]*AutomationRule{rule}, nil)
	// TODO: Fix RuleManager mocking - ruleManager.On("EvaluateConditions", ctx, rule, event).Return(true, nil)
	// TODO: Fix RuleManager mocking - ruleManager.On("ExecuteRule", ctx, rule, mock.AnythingOfType("*github.AutomationExecutionContext")).Return((*AutomationRuleExecution)(nil), assert.AnError).Times(3) // Will retry

	err := engine.Start(ctx)
	require.NoError(t, err)

	defer func() {
		if err := engine.Stop(ctx); err != nil {
			t.Logf("Failed to stop engine: %v", err)
		}
	}()

	err = engine.ProcessEvent(ctx, event)
	assert.NoError(t, err)

	// Wait for processing and retries
	time.Sleep(1 * time.Second)

	metrics := engine.GetMetrics()
	assert.Greater(t, metrics.ExecutionErrors, int64(0))

	eventProcessor.AssertExpectations(t)
	// TODO: Fix RuleManager mocking - ruleManager.AssertExpectations(t)
}

func TestAutomationEngine_GetMetrics(t *testing.T) {
	engine, _ := createTestAutomationEngine()

	metrics := engine.GetMetrics()
	assert.NotNil(t, metrics)
	assert.Equal(t, int64(0), metrics.EventsProcessed)
	assert.NotNil(t, metrics.EventTypeDistribution)
	assert.NotNil(t, metrics.ExecutionsByStatus)
}

func TestAutomationEngine_GetActiveExecutions(t *testing.T) {
	engine, _ := createTestAutomationEngine()

	executions := engine.GetActiveExecutions()
	assert.NotNil(t, executions)
	assert.Len(t, executions, 0)
}

func TestAutomationEngine_SyncExecution(t *testing.T) {
	engine, eventProcessor := createTestAutomationEngine()
	engine.config.EnableAsyncExecution = false // Disable async execution

	event := createTestEngineEvent()
	// 	rule := createTestEngineRule()
	ctx := context.Background()

	eventProcessor.On("ValidateEvent", mock.Anything, event).Return(nil)
	eventProcessor.On("FilterEvent", event).Return(true)
	// TODO: Fix RuleManager mocking - ruleManager.On("ListRules", ctx, "testorg", mock.AnythingOfType("*github.RuleFilter")).Return([]*AutomationRule{rule}, nil)
	// TODO: Fix RuleManager mocking - ruleManager.On("EvaluateConditions", ctx, rule, event).Return(true, nil)

	// 	execution := &AutomationRuleExecution{
	// 		ID:     "exec-001",
	// 		RuleID: rule.ID,
	// 		Status: ExecutionStatusCompleted,
	// 	}
	// TODO: Fix RuleManager mocking - ruleManager.On("ExecuteRule", ctx, rule, mock.AnythingOfType("*github.AutomationExecutionContext")).Return(execution, nil)

	err := engine.Start(ctx)
	require.NoError(t, err)

	defer func() {
		if err := engine.Stop(ctx); err != nil {
			t.Logf("Failed to stop engine: %v", err)
		}
	}()

	err = engine.ProcessEvent(ctx, event)
	assert.NoError(t, err)

	// Wait for processing
	time.Sleep(200 * time.Millisecond)

	metrics := engine.GetMetrics()
	assert.Greater(t, metrics.RulesExecuted, int64(0))

	eventProcessor.AssertExpectations(t)
	// TODO: Fix RuleManager mocking - ruleManager.AssertExpectations(t)
}

func TestAutomationEngine_EventChannelFull(t *testing.T) {
	engine, eventProcessor := createTestAutomationEngine()
	engine.config.EventBufferSize = 1 // Very small buffer

	event := createTestEngineEvent()
	ctx := context.Background()

	eventProcessor.On("ValidateEvent", mock.Anything, event).Return(nil)

	err := engine.Start(ctx)
	require.NoError(t, err)

	defer func() {
		if err := engine.Stop(ctx); err != nil {
			t.Logf("Failed to stop engine: %v", err)
		}
	}()

	// Fill the channel
	err = engine.ProcessEvent(ctx, event)
	assert.NoError(t, err)

	// This should fail because channel is full
	err = engine.ProcessEvent(ctx, event)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "event channel is full")

	eventProcessor.AssertExpectations(t)
}

func TestAutomationEngineConfig_Defaults(t *testing.T) {
	config := getDefaultConfig()

	assert.Equal(t, 10, config.MaxWorkers)
	assert.Equal(t, 1000, config.EventBufferSize)
	assert.Equal(t, 5*time.Minute, config.ExecutionTimeout)
	assert.True(t, config.EnableAsyncExecution)
	assert.True(t, config.EnableRuleFiltering)
	assert.True(t, config.EnableMetrics)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 2.0, config.RetryBackoffFactor)
}

func TestEngineMetrics_ThreadSafety(t *testing.T) {
	engine, _ := createTestAutomationEngine()

	// Simulate concurrent access to metrics
	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := 0; i < 100; i++ {
			engine.updateMetrics(func(m *EngineMetrics) {
				m.EventsProcessed++
				m.RulesEvaluated++
			})
		}

		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 100; i++ {
			_ = engine.GetMetrics()
		}

		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done

	metrics := engine.GetMetrics()
	assert.Equal(t, int64(100), metrics.EventsProcessed)
	assert.Equal(t, int64(100), metrics.RulesEvaluated)
}

func TestAutomationEngine_ContextCancellation(t *testing.T) {
	engine, _ := createTestAutomationEngine()

	ctx, cancel := context.WithCancel(context.Background())

	err := engine.Start(ctx)
	require.NoError(t, err)

	// Cancel context
	cancel()

	// Stop should work even with cancelled context
	stopCtx := context.Background()
	err = engine.Stop(stopCtx)
	assert.NoError(t, err)
}

// Benchmark tests

func BenchmarkAutomationEngine_ProcessEvent(b *testing.B) {
	// Create a proper benchmark setup with mocks
	logger := &mockLogger{}
	apiClient := &mockAPIClient{}
	conditionEvaluator := &mockConditionEvaluator{}
	actionExecutor := &mockActionExecutor{}
	eventProcessor := &mockEventProcessor{}

	// Create a real RuleManager since NewAutomationEngine expects *RuleManager
	ruleManager := &RuleManager{
		logger:         logger,
		apiClient:      apiClient,
		evaluator:      conditionEvaluator,
		actionExecutor: actionExecutor,
	}

	config := &AutomationEngineConfig{
		MaxWorkers:           2,
		EventBufferSize:      10,
		ExecutionTimeout:     30 * time.Second,
		EnableAsyncExecution: true,
		EnableRuleFiltering:  true,
		EnableMetrics:        true,
		MaxRetries:           2,
		RetryBackoffFactor:   1.5,
	}

	engine := NewAutomationEngine(
		logger,
		apiClient,
		ruleManager,
		conditionEvaluator,
		actionExecutor,
		eventProcessor,
		config,
	)

	event := createTestEngineEvent()
	// 	rule := createTestEngineRule()
	ctx := context.Background()

	eventProcessor.On("ValidateEvent", mock.Anything, event).Return(nil)
	eventProcessor.On("FilterEvent", event).Return(true)
	// TODO: Fix RuleManager mocking - ruleManager.On("ListRules", ctx, "testorg", mock.AnythingOfType("*github.RuleFilter")).Return([]*AutomationRule{rule}, nil)
	// TODO: Fix RuleManager mocking - ruleManager.On("EvaluateConditions", ctx, rule, event).Return(false, nil) // Don't execute for benchmark

	if err := engine.Start(ctx); err != nil {
		b.Errorf("Failed to start engine: %v", err)
	}
	defer func() {
		if err := engine.Stop(ctx); err != nil {
			b.Logf("Failed to stop engine: %v", err)
		}
	}()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := engine.ProcessEvent(ctx, event); err != nil {
			// Ignore errors in benchmark
		}
	}
}

func BenchmarkAutomationEngine_UpdateMetrics(b *testing.B) {
	engine, _ := createTestAutomationEngine()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		engine.updateMetrics(func(m *EngineMetrics) {
			m.EventsProcessed++
		})
	}
}

// Integration test

func TestAutomationEngine_Integration(t *testing.T) {
	engine, eventProcessor := createTestAutomationEngine()

	// Test complete flow: start -> process event -> execute rule -> stop
	event := createTestEngineEvent()
	// rule := createTestEngineRule()
	ctx := context.Background()

	// Set up mocks for complete flow
	eventProcessor.On("ValidateEvent", mock.Anything, event).Return(nil)
	eventProcessor.On("FilterEvent", event).Return(true)
	// TODO: Fix RuleManager mocking - ruleManager.On("ListRules", ctx, "testorg", mock.AnythingOfType("*github.RuleFilter")).Return([]*AutomationRule{rule}, nil)
	// TODO: Fix RuleManager mocking - ruleManager.On("EvaluateConditions", ctx, rule, event).Return(true, nil)

	// 	execution := &AutomationRuleExecution{
	// 		ID:     "exec-001",
	// 		RuleID: rule.ID,
	// 		Status: ExecutionStatusCompleted,
	// 		Actions: []ActionExecutionResult{
	// 			{
	// 				ActionID:   "action-001",
	// 				ActionType: ActionTypeAddLabel,
	// 				Status:     ExecutionStatusCompleted,
	// 			},
	// 		},
	// 	}
	// 	// TODO: Fix RuleManager mocking - ruleManager.On("ExecuteRule", ctx, rule, mock.AnythingOfType("*github.AutomationExecutionContext")).Return(execution, nil)

	// Start engine
	err := engine.Start(ctx)
	require.NoError(t, err)

	// Process event
	err = engine.ProcessEvent(ctx, event)
	require.NoError(t, err)

	// Wait for processing
	time.Sleep(500 * time.Millisecond)

	// Verify metrics
	metrics := engine.GetMetrics()
	assert.Equal(t, int64(1), metrics.EventsProcessed)
	assert.Greater(t, metrics.RulesEvaluated, int64(0))
	assert.Greater(t, metrics.RulesExecuted, int64(0))

	// Stop engine
	err = engine.Stop(ctx)
	require.NoError(t, err)

	// Verify all mocks were called
	eventProcessor.AssertExpectations(t)
	// TODO: Fix RuleManager mocking - ruleManager.AssertExpectations(t)
}
