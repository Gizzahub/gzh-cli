//nolint:testpackage // White-box testing needed for internal function access
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

func (m *mockAPIClient) SetToken(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
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

func (m *mockAPIClient) CreateRepository(ctx context.Context, owner string, opts *CreateRepositoryOptions) (*RepositoryInfo, error) {
	args := m.Called(ctx, owner, opts)
	if info, ok := args.Get(0).(*RepositoryInfo); ok {
		return info, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAPIClient) DeleteRepository(ctx context.Context, owner, repo string) error {
	args := m.Called(ctx, owner, repo)
	return args.Error(0)
}

func (m *mockAPIClient) ArchiveRepository(ctx context.Context, owner, repo string) error {
	args := m.Called(ctx, owner, repo)
	return args.Error(0)
}

func (m *mockAPIClient) UnarchiveRepository(ctx context.Context, owner, repo string) error {
	args := m.Called(ctx, owner, repo)
	return args.Error(0)
}

func (m *mockAPIClient) SearchRepositories(ctx context.Context, query string, opts *SearchRepositoriesOptions) (*RepositorySearchResult, error) {
	args := m.Called(ctx, query, opts)
	if result, ok := args.Get(0).(*RepositorySearchResult); ok {
		return result, args.Error(1)
	}
	return nil, args.Error(1)
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

// mockLogger는 로그를 버리는 시험용 구현이다.
//
// 예전에는 mock.Mock을 심어 두고 모든 메서드가 m.Called()를 부르게 했다.
// testify의 mock은 기대를 걸지 않은 호출을 받으면 패닉하므로, 검사 대상
// 코드가 로그를 한 줄만 남겨도 테스트가 패닉으로 끝났다 -- 패닉은 그
// 테스트만이 아니라 패키지 전체 실행을 중단시켜서 뒤따르는 테스트는
// 아예 돌지 않았다.
//
// 이 패키지에서 mockLogger는 59곳에서 만들어지지만 기대를 거는 곳
// (.On("Info", ...))도, 호출을 확인하는 곳도 하나도 없다. 로그는 검증
// 대상이 아니라 의존성을 채우기 위한 자리였다는 뜻이므로, 확인 기능을
// 흉내내는 대신 실제로 아무것도 하지 않게 둔다.
type mockLogger struct{}

func (m *mockLogger) Debug(_ string, _ ...any) {}

func (m *mockLogger) Info(_ string, _ ...any) {}

func (m *mockLogger) Warn(_ string, _ ...any) {}

func (m *mockLogger) Error(_ string, _ ...any) {}

func (m *mockLogger) Fatal(_ string, _ ...any) {}

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

func (m *mockConditionEvaluator) EvaluatePayloadMatcher(ctx context.Context, matcher *PayloadMatcher, payload map[string]any) (bool, error) {
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

type mockRuleManager struct {
	mock.Mock
}

func (m *mockRuleManager) CreateRule(ctx context.Context, rule *AutomationRule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *mockRuleManager) GetRule(ctx context.Context, org, ruleID string) (*AutomationRule, error) {
	args := m.Called(ctx, org, ruleID)
	if rule, ok := args.Get(0).(*AutomationRule); ok {
		return rule, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRuleManager) ListRules(ctx context.Context, org string, filter *RuleFilter) ([]*AutomationRule, error) {
	args := m.Called(ctx, org, filter)
	if rules, ok := args.Get(0).([]*AutomationRule); ok {
		return rules, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRuleManager) UpdateRule(ctx context.Context, rule *AutomationRule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *mockRuleManager) DeleteRule(ctx context.Context, org, ruleID string) error {
	args := m.Called(ctx, org, ruleID)
	return args.Error(0)
}

func (m *mockRuleManager) EnableRule(ctx context.Context, org, ruleID string) error {
	args := m.Called(ctx, org, ruleID)
	return args.Error(0)
}

func (m *mockRuleManager) DisableRule(ctx context.Context, org, ruleID string) error {
	args := m.Called(ctx, org, ruleID)
	return args.Error(0)
}

func (m *mockRuleManager) EvaluateConditions(ctx context.Context, rule *AutomationRule, event *GitHubEvent) (bool, error) {
	args := m.Called(ctx, rule, event)
	return args.Bool(0), args.Error(1)
}

func (m *mockRuleManager) ExecuteRule(ctx context.Context, rule *AutomationRule, context *AutomationExecutionContext) (*AutomationRuleExecution, error) {
	args := m.Called(ctx, rule, context)
	if execution, ok := args.Get(0).(*AutomationRuleExecution); ok {
		return execution, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRuleManager) CreateRuleSet(ctx context.Context, ruleSet *AutomationRuleSet) error {
	args := m.Called(ctx, ruleSet)
	return args.Error(0)
}

func (m *mockRuleManager) GetRuleSet(ctx context.Context, org, setID string) (*AutomationRuleSet, error) {
	args := m.Called(ctx, org, setID)
	if ruleSet, ok := args.Get(0).(*AutomationRuleSet); ok {
		return ruleSet, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRuleManager) ListRuleSets(ctx context.Context, org string) ([]*AutomationRuleSet, error) {
	args := m.Called(ctx, org)
	if ruleSets, ok := args.Get(0).([]*AutomationRuleSet); ok {
		return ruleSets, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRuleManager) UpdateRuleSet(ctx context.Context, ruleSet *AutomationRuleSet) error {
	args := m.Called(ctx, ruleSet)
	return args.Error(0)
}

func (m *mockRuleManager) DeleteRuleSet(ctx context.Context, org, setID string) error {
	args := m.Called(ctx, org, setID)
	return args.Error(0)
}

func (m *mockRuleManager) CreateTemplate(ctx context.Context, template *AutomationRuleTemplate) error {
	args := m.Called(ctx, template)
	return args.Error(0)
}

func (m *mockRuleManager) GetTemplate(ctx context.Context, templateID string) (*AutomationRuleTemplate, error) {
	args := m.Called(ctx, templateID)
	if template, ok := args.Get(0).(*AutomationRuleTemplate); ok {
		return template, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRuleManager) ListTemplates(ctx context.Context, category string) ([]*AutomationRuleTemplate, error) {
	args := m.Called(ctx, category)
	if templates, ok := args.Get(0).([]*AutomationRuleTemplate); ok {
		return templates, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRuleManager) UpdateTemplate(ctx context.Context, template *AutomationRuleTemplate) error {
	args := m.Called(ctx, template)
	return args.Error(0)
}

func (m *mockRuleManager) DeleteTemplate(ctx context.Context, templateID string) error {
	args := m.Called(ctx, templateID)
	return args.Error(0)
}

func (m *mockRuleManager) InstantiateTemplate(ctx context.Context, templateID string, variables map[string]any) (*AutomationRule, error) {
	args := m.Called(ctx, templateID, variables)
	if rule, ok := args.Get(0).(*AutomationRule); ok {
		return rule, args.Error(1)
	}
	return nil, args.Error(1)
}

// Test helper functions

func createTestAutomationEngine() (*AutomationEngine, *mockEventProcessor, *mockRuleManager) {
	logger := &mockLogger{}
	apiClient := &mockAPIClient{}

	// Create a mock RuleManager for testing
	mockRM := &mockRuleManager{}

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

	// 예전에는 값이 비어 있는 &RuleManager{}를 엔진에 넘기고 mockRM은
	// 호출자에게만 돌려줬다. 그래서 테스트가 mockRM에 건 ListRules 기대는
	// 아무도 부르지 않았고, 실제로 불린 &RuleManager{}는 storage가 nil이라
	// 워커 고루틴에서 SIGSEGV로 죽었다. 이제 엔진이 인터페이스를 받으므로
	// 돌려주는 대역과 엔진이 쓰는 대역이 같다.
	engine := NewAutomationEngine(
		logger,
		apiClient,
		mockRM,
		conditionEvaluator,
		actionExecutor,
		eventProcessor,
		config,
	)

	return engine, eventProcessor, mockRM
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
		Payload: map[string]any{
			"action": "opened",
			"pull_request": map[string]any{
				"title":  "Test PR",
				"number": 1,
			},
		},
	}
}

func createTestEngineRule() *AutomationRule {
	return &AutomationRule{
		ID:           "test-rule-001",
		Name:         "Test Automation Rule",
		Organization: "testorg",
		Enabled:      true,
		Conditions: AutomationConditions{
			EventTypes: []EventType{"push"},
		},
		Actions: []AutomationAction{
			{
				ID:      "action-001",
				Type:    ActionTypeWebhook,
				Enabled: true,
				Parameters: map[string]any{
					"url": "https://example.com/webhook",
				},
			},
		},
		Priority: 1,
	}
}

// Test Cases

func TestNewAutomationEngine(t *testing.T) {
	engine, _, _ := createTestAutomationEngine()

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

	// 엔진이 인터페이스를 받으므로 storage가 nil인 실제 RuleManager 대신
	// 대역을 그대로 넣는다.
	ruleManager := &mockRuleManager{}

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
	engine, _, _ := createTestAutomationEngine()
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
	engine, _, _ := createTestAutomationEngine()
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
	engine, _, _ := createTestAutomationEngine()
	event := createTestEngineEvent()

	err := engine.ProcessEvent(context.Background(), event)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")
}

func TestAutomationEngine_ProcessEvent_ValidationFailed(t *testing.T) {
	engine, eventProcessor, _ := createTestAutomationEngine()
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
	engine, eventProcessor, _ := createTestAutomationEngine()
	event := createTestEngineEvent()
	ctx := context.Background()

	eventProcessor.On("ValidateEvent", mock.Anything, event).Return(nil)
	eventProcessor.On("FilterEvent", mock.Anything, event, mock.Anything).Return(false, nil) // Event is filtered out

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

	// 걸러진 이벤트는 처리 건수에 잡히지 않는다. EventsProcessed는
	// LastProcessedEvent, EventTypeDistribution과 한 묶음으로 이벤트 채널에
	// 실제로 실린 시점에만 올라가므로 "규칙 실행으로 넘어간 이벤트"를 뜻한다.
	// 예전 기대값 1은 이 테스트의 이름이 말하는 것과 정반대였다.
	metrics := engine.GetMetrics()
	assert.Equal(t, int64(0), metrics.EventsProcessed)

	eventProcessor.AssertExpectations(t)
}

func TestAutomationEngine_ProcessEvent_ExcludedEventType(t *testing.T) {
	engine, eventProcessor, _ := createTestAutomationEngine()
	engine.config.ExcludedEventTypes = []EventType{EventTypePullRequest}
	event := createTestEngineEvent()
	ctx := context.Background()

	eventProcessor.On("ValidateEvent", mock.Anything, event).Return(nil)
	// 설정의 EnableRuleFiltering이 켜져 있으므로 제외 타입 검사에 닿기 전에
	// FilterEvent가 먼저 불린다. 기대를 걸어 두지 않으면 통과가 아니라
	// 패닉으로 끝난다.
	eventProcessor.On("FilterEvent", mock.Anything, event, mock.Anything).Return(true, nil)

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

	// 제외된 타입도 걸러진 이벤트와 같이 채널에 실리지 않으므로 0이다.
	metrics := engine.GetMetrics()
	assert.Equal(t, int64(0), metrics.EventsProcessed)

	eventProcessor.AssertExpectations(t)
}

func TestAutomationEngine_ProcessEvent_Success(t *testing.T) {
	engine, eventProcessor, ruleManager := createTestAutomationEngine()
	event := createTestEngineEvent()
	rule := createTestEngineRule()
	ctx := context.Background()

	eventProcessor.On("ValidateEvent", mock.Anything, event).Return(nil)
	eventProcessor.On("FilterEvent", mock.Anything, event, mock.Anything).Return(true, nil)
	ruleManager.On("ListRules", mock.Anything, "testorg", mock.AnythingOfType("*github.RuleFilter")).Return([]*AutomationRule{rule}, nil)
	ruleManager.On("EvaluateConditions", mock.Anything, rule, event).Return(true, nil)

	execution := &AutomationRuleExecution{
		ID:     "exec-001",
		RuleID: rule.ID,
		Status: ExecutionStatusCompleted,
	}
	ruleManager.On("ExecuteRule", mock.Anything, rule, mock.AnythingOfType("*github.AutomationExecutionContext")).Return(execution, nil)

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
	ruleManager.AssertExpectations(t)
}

func TestAutomationEngine_ProcessEvent_NoMatchingRules(t *testing.T) {
	engine, eventProcessor, ruleManager := createTestAutomationEngine()
	event := createTestEngineEvent()
	rule := createTestEngineRule()
	ctx := context.Background()

	eventProcessor.On("ValidateEvent", mock.Anything, event).Return(nil)
	eventProcessor.On("FilterEvent", mock.Anything, event, mock.Anything).Return(true, nil)
	ruleManager.On("ListRules", mock.Anything, "testorg", mock.AnythingOfType("*github.RuleFilter")).Return([]*AutomationRule{rule}, nil)
	ruleManager.On("EvaluateConditions", mock.Anything, rule, event).Return(false, nil) // Conditions don't match

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
	ruleManager.AssertExpectations(t)
}

func TestAutomationEngine_ProcessEvent_ExecutionFailure(t *testing.T) {
	engine, eventProcessor, ruleManager := createTestAutomationEngine()
	event := createTestEngineEvent()
	rule := createTestEngineRule()
	ctx := context.Background()

	eventProcessor.On("ValidateEvent", mock.Anything, event).Return(nil)
	eventProcessor.On("FilterEvent", mock.Anything, event, mock.Anything).Return(true, nil)
	ruleManager.On("ListRules", mock.Anything, "testorg", mock.AnythingOfType("*github.RuleFilter")).Return([]*AutomationRule{rule}, nil)
	ruleManager.On("EvaluateConditions", mock.Anything, rule, event).Return(true, nil)
	ruleManager.On("ExecuteRule", mock.Anything, rule, mock.AnythingOfType("*github.AutomationExecutionContext")).
		Return((*AutomationRuleExecution)(nil), assert.AnError)

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
	ruleManager.AssertExpectations(t)
}

func TestAutomationEngine_GetMetrics(t *testing.T) {
	engine, _, _ := createTestAutomationEngine()

	metrics := engine.GetMetrics()
	assert.NotNil(t, metrics)
	assert.Equal(t, int64(0), metrics.EventsProcessed)
	assert.NotNil(t, metrics.EventTypeDistribution)
	assert.NotNil(t, metrics.ExecutionsByStatus)
}

func TestAutomationEngine_GetActiveExecutions(t *testing.T) {
	engine, _, _ := createTestAutomationEngine()

	executions := engine.GetActiveExecutions()
	assert.NotNil(t, executions)
	assert.Len(t, executions, 0)
}

func TestAutomationEngine_SyncExecution(t *testing.T) {
	engine, eventProcessor, ruleManager := createTestAutomationEngine()
	engine.config.EnableAsyncExecution = false // Disable async execution

	event := createTestEngineEvent()
	rule := createTestEngineRule()
	ctx := context.Background()

	eventProcessor.On("ValidateEvent", mock.Anything, event).Return(nil)
	eventProcessor.On("FilterEvent", mock.Anything, event, mock.Anything).Return(true, nil)
	ruleManager.On("ListRules", mock.Anything, "testorg", mock.AnythingOfType("*github.RuleFilter")).Return([]*AutomationRule{rule}, nil)
	ruleManager.On("EvaluateConditions", mock.Anything, rule, event).Return(true, nil)

	execution := &AutomationRuleExecution{
		ID:     "exec-001",
		RuleID: rule.ID,
		Status: ExecutionStatusCompleted,
	}
	ruleManager.On("ExecuteRule", mock.Anything, rule, mock.AnythingOfType("*github.AutomationExecutionContext")).Return(execution, nil)

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
	ruleManager.AssertExpectations(t)
}

func TestAutomationEngine_EventChannelFull(t *testing.T) {
	// 이 시험은 버퍼가 찼을 때 이벤트가 거절되는지를 본다. 그러려면 두 가지가
	// 필요하다.
	//
	// 첫째, 버퍼 크기를 생성 시점에 정해야 한다. 예전에는 엔진을 만든 뒤
	// engine.config.EventBufferSize = 1로 바꿨는데, 채널은 이미 생성자에서
	// make(chan, 10)으로 만들어진 뒤라 이 대입은 아무 효과가 없었다.
	//
	// 둘째, 채널을 비우는 쪽이 없어야 한다. MaxWorkers가 2면 이벤트 워커가
	// 곧바로 채널을 비우므로 "가득 참"은 경합에 따라 나기도 하고 안 나기도
	// 한다. MaxWorkers를 0으로 두면 이벤트 워커가 뜨지 않아 결과가 확정된다.
	logger := &mockLogger{}
	apiClient := &mockAPIClient{}
	ruleManager := &mockRuleManager{}
	conditionEvaluator := &mockConditionEvaluator{}
	actionExecutor := &mockActionExecutor{}
	eventProcessor := &mockEventProcessor{}

	engine := NewAutomationEngine(
		logger,
		apiClient,
		ruleManager,
		conditionEvaluator,
		actionExecutor,
		eventProcessor,
		&AutomationEngineConfig{
			MaxWorkers:           0, // 이벤트 워커 없음 -- 채널을 비우는 쪽이 없다
			EventBufferSize:      1, // 한 건만 담긴다
			ExecutionTimeout:     30 * time.Second,
			EnableAsyncExecution: true,
			EnableRuleFiltering:  true,
			EnableMetrics:        true,
			MaxRetries:           2,
			RetryBackoffFactor:   1.5,
		},
	)

	event := createTestEngineEvent()
	ctx := context.Background()

	eventProcessor.On("ValidateEvent", mock.Anything, event).Return(nil)
	// EnableRuleFiltering이 켜져 있으므로 채널에 닿기 전에 FilterEvent가 먼저
	// 불린다. 기대를 걸어 두지 않으면 통과가 아니라 패닉으로 끝난다.
	eventProcessor.On("FilterEvent", mock.Anything, event, mock.Anything).Return(true, nil)

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
	engine, _, _ := createTestAutomationEngine()

	// Simulate concurrent access to metrics
	done := make(chan bool)

	// Writer goroutine
	go func() {
		for range 100 {
			engine.updateMetrics(func(m *EngineMetrics) {
				m.EventsProcessed++
				m.RulesEvaluated++
			})
		}

		done <- true
	}()

	// Reader goroutine
	go func() {
		for range 100 {
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
	engine, _, _ := createTestAutomationEngine()

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

	// 엔진이 인터페이스를 받으므로 storage가 nil인 실제 RuleManager 대신
	// 대역을 그대로 넣는다.
	ruleManager := &mockRuleManager{}

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
	rule := createTestEngineRule()
	ctx := context.Background()

	eventProcessor.On("ValidateEvent", mock.Anything, event).Return(nil)
	eventProcessor.On("FilterEvent", mock.Anything, event, mock.Anything).Return(true, nil)
	ruleManager.On("ListRules", mock.Anything, "testorg", mock.AnythingOfType("*github.RuleFilter")).Return([]*AutomationRule{rule}, nil)
	ruleManager.On("EvaluateConditions", mock.Anything, rule, event).Return(false, nil) // Don't execute for benchmark

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
	engine, _, _ := createTestAutomationEngine()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		engine.updateMetrics(func(m *EngineMetrics) {
			m.EventsProcessed++
		})
	}
}

// Integration test

func TestAutomationEngine_Integration(t *testing.T) {
	engine, eventProcessor, ruleManager := createTestAutomationEngine()

	// Test complete flow: start -> process event -> execute rule -> stop
	event := createTestEngineEvent()
	rule := createTestEngineRule()
	ctx := context.Background()

	// Set up mocks for complete flow
	eventProcessor.On("ValidateEvent", mock.Anything, event).Return(nil)
	eventProcessor.On("FilterEvent", mock.Anything, event, mock.Anything).Return(true, nil)
	ruleManager.On("ListRules", mock.Anything, "testorg", mock.AnythingOfType("*github.RuleFilter")).Return([]*AutomationRule{rule}, nil)
	ruleManager.On("EvaluateConditions", mock.Anything, rule, event).Return(true, nil)

	execution := &AutomationRuleExecution{
		ID:     "exec-001",
		RuleID: rule.ID,
		Status: ExecutionStatusCompleted,
		Actions: []ActionExecutionResult{
			{
				ActionID:   "action-001",
				ActionType: ActionTypeAddLabel,
				Status:     ExecutionStatusCompleted,
			},
		},
	}
	ruleManager.On("ExecuteRule", mock.Anything, rule, mock.AnythingOfType("*github.AutomationExecutionContext")).Return(execution, nil)

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
	ruleManager.AssertExpectations(t)
}
