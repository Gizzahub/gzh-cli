package github

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mock implementations for testing

type mockRuleStorage struct {
	mock.Mock
}

func (m *mockRuleStorage) CreateRule(ctx context.Context, rule *AutomationRule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *mockRuleStorage) GetRule(ctx context.Context, org, ruleID string) (*AutomationRule, error) {
	args := m.Called(ctx, org, ruleID)
	if rule, ok := args.Get(0).(*AutomationRule); ok {
		return rule, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRuleStorage) ListRules(ctx context.Context, org string, filter *RuleFilter) ([]*AutomationRule, error) {
	args := m.Called(ctx, org, filter)
	if rules, ok := args.Get(0).([]*AutomationRule); ok {
		return rules, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRuleStorage) UpdateRule(ctx context.Context, rule *AutomationRule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *mockRuleStorage) DeleteRule(ctx context.Context, org, ruleID string) error {
	args := m.Called(ctx, org, ruleID)
	return args.Error(0)
}

func (m *mockRuleStorage) CreateRuleSet(ctx context.Context, ruleSet *AutomationRuleSet) error {
	args := m.Called(ctx, ruleSet)
	return args.Error(0)
}

func (m *mockRuleStorage) GetRuleSet(ctx context.Context, org, setID string) (*AutomationRuleSet, error) {
	args := m.Called(ctx, org, setID)
	if ruleSet, ok := args.Get(0).(*AutomationRuleSet); ok {
		return ruleSet, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRuleStorage) ListRuleSets(ctx context.Context, org string) ([]*AutomationRuleSet, error) {
	args := m.Called(ctx, org)
	if ruleSets, ok := args.Get(0).([]*AutomationRuleSet); ok {
		return ruleSets, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRuleStorage) UpdateRuleSet(ctx context.Context, ruleSet *AutomationRuleSet) error {
	args := m.Called(ctx, ruleSet)
	return args.Error(0)
}

func (m *mockRuleStorage) DeleteRuleSet(ctx context.Context, org, setID string) error {
	args := m.Called(ctx, org, setID)
	return args.Error(0)
}

func (m *mockRuleStorage) SaveExecution(ctx context.Context, execution *AutomationRuleExecution) error {
	args := m.Called(ctx, execution)
	return args.Error(0)
}

func (m *mockRuleStorage) GetExecution(ctx context.Context, executionID string) (*AutomationRuleExecution, error) {
	args := m.Called(ctx, executionID)
	if execution, ok := args.Get(0).(*AutomationRuleExecution); ok {
		return execution, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRuleStorage) ListExecutions(ctx context.Context, org string, filter *ExecutionFilter) ([]*AutomationRuleExecution, error) {
	args := m.Called(ctx, org, filter)
	if executions, ok := args.Get(0).([]*AutomationRuleExecution); ok {
		return executions, args.Error(1)
	}
	return nil, args.Error(1)
}

type mockTemplateStorage struct {
	mock.Mock
}

func (m *mockTemplateStorage) CreateTemplate(ctx context.Context, template *AutomationRuleTemplate) error {
	args := m.Called(ctx, template)
	return args.Error(0)
}

func (m *mockTemplateStorage) GetTemplate(ctx context.Context, templateID string) (*AutomationRuleTemplate, error) {
	args := m.Called(ctx, templateID)
	if template, ok := args.Get(0).(*AutomationRuleTemplate); ok {
		return template, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockTemplateStorage) ListTemplates(ctx context.Context, category string) ([]*AutomationRuleTemplate, error) {
	args := m.Called(ctx, category)
	if templates, ok := args.Get(0).([]*AutomationRuleTemplate); ok {
		return templates, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockTemplateStorage) UpdateTemplate(ctx context.Context, template *AutomationRuleTemplate) error {
	args := m.Called(ctx, template)
	return args.Error(0)
}

func (m *mockTemplateStorage) DeleteTemplate(ctx context.Context, templateID string) error {
	args := m.Called(ctx, templateID)
	return args.Error(0)
}

type mockRuleActionExecutor struct {
	mock.Mock
}

func (m *mockRuleActionExecutor) ExecuteAction(ctx context.Context, action *AutomationAction, context *AutomationExecutionContext) (*ActionExecutionResult, error) {
	args := m.Called(ctx, action, context)
	if result, ok := args.Get(0).(*ActionExecutionResult); ok {
		return result, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRuleActionExecutor) ValidateAction(ctx context.Context, action *AutomationAction) error {
	args := m.Called(ctx, action)
	return args.Error(0)
}

func (m *mockRuleActionExecutor) GetSupportedActions() []ActionType {
	args := m.Called()
	if actions, ok := args.Get(0).([]ActionType); ok {
		return actions
	}
	return nil
}

type mockRuleConditionEvaluator struct {
	mock.Mock
}

func (m *mockRuleConditionEvaluator) EvaluateConditions(ctx context.Context, conditions *AutomationConditions, event *GitHubEvent, context *EvaluationContext) (*EvaluationResult, error) {
	args := m.Called(ctx, conditions, event, context)
	if result, ok := args.Get(0).(*EvaluationResult); ok {
		return result, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRuleConditionEvaluator) EvaluatePayloadMatcher(ctx context.Context, matcher *PayloadMatcher, payload map[string]interface{}) (bool, error) {
	args := m.Called(ctx, matcher, payload)
	return args.Bool(0), args.Error(1)
}

func (m *mockRuleConditionEvaluator) EvaluateEventConditions(event *GitHubEvent, conditions *AutomationConditions) (bool, error) {
	args := m.Called(event, conditions)
	return args.Bool(0), args.Error(1)
}

func (m *mockRuleConditionEvaluator) EvaluateRepositoryConditions(ctx context.Context, repoInfo *RepositoryInfo, conditions *AutomationConditions) (bool, error) {
	args := m.Called(ctx, repoInfo, conditions)
	return args.Bool(0), args.Error(1)
}

func (m *mockRuleConditionEvaluator) EvaluateTimeConditions(timestamp time.Time, conditions *AutomationConditions) (bool, error) {
	args := m.Called(timestamp, conditions)
	return args.Bool(0), args.Error(1)
}

func (m *mockRuleConditionEvaluator) EvaluateContentConditions(ctx context.Context, event *GitHubEvent, conditions *AutomationConditions) (bool, error) {
	args := m.Called(ctx, event, conditions)
	return args.Bool(0), args.Error(1)
}

func (m *mockRuleConditionEvaluator) ValidateConditions(conditions *AutomationConditions) (*ConditionValidationResult, error) {
	args := m.Called(conditions)
	if result, ok := args.Get(0).(*ConditionValidationResult); ok {
		return result, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRuleConditionEvaluator) ExplainEvaluation(ctx context.Context, conditions *AutomationConditions, event *GitHubEvent) (*EvaluationExplanation, error) {
	args := m.Called(ctx, conditions, event)
	if explanation, ok := args.Get(0).(*EvaluationExplanation); ok {
		return explanation, args.Error(1)
	}
	return nil, args.Error(1)
}

// Test helper functions

func createTestRuleManager() (*RuleManager, *mockRuleStorage, *mockTemplateStorage, *mockRuleActionExecutor, *mockRuleConditionEvaluator) {
	logger := &mockLogger{}
	apiClient := &mockAPIClient{}
	storage := &mockRuleStorage{}
	templateStorage := &mockTemplateStorage{}
	actionExecutor := &mockRuleActionExecutor{}
	evaluator := &mockRuleConditionEvaluator{}

	rm := NewRuleManager(logger, apiClient, evaluator, actionExecutor, storage, templateStorage)

	return rm, storage, templateStorage, actionExecutor, evaluator
}

func createTestRule() *AutomationRule {
	return &AutomationRule{
		ID:           "test-rule-001",
		Name:         "Test Rule",
		Description:  "Test automation rule",
		Organization: "testorg",
		Enabled:      true,
		Priority:     100,
		Conditions: AutomationConditions{
			EventTypes:      []EventType{EventTypePullRequest},
			Actions:         []EventAction{ActionOpened},
			LogicalOperator: ConditionOperatorAND,
		},
		Actions: []AutomationAction{
			{
				ID:      "test-action-001",
				Type:    ActionTypeAddLabel,
				Name:    "Add Label",
				Enabled: true,
				Parameters: map[string]interface{}{
					"labels": []string{"automated"},
				},
			},
		},
		Metadata: AutomationRuleMetadata{
			Version:     "1.0",
			Environment: "test",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		CreatedBy: "test-user",
	}
}

func createTestEventForRuleManager() *GitHubEvent {
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

func TestRuleManager_CreateRule(t *testing.T) {
	rm, storage, _, actionExecutor, evaluator := createTestRuleManager()
	rule := createTestRule()

	// Set up mocks
	evaluator.On("ValidateConditions", &rule.Conditions).Return(&ConditionValidationResult{Valid: true}, nil)
	actionExecutor.On("ValidateAction", mock.Anything, &rule.Actions[0]).Return(nil)
	storage.On("CreateRule", mock.Anything, mock.AnythingOfType("*github.AutomationRule")).Return(nil)

	err := rm.CreateRule(context.Background(), rule)

	assert.NoError(t, err)
	assert.NotEmpty(t, rule.ID)
	assert.False(t, rule.CreatedAt.IsZero())
	assert.False(t, rule.UpdatedAt.IsZero())

	storage.AssertExpectations(t)
	actionExecutor.AssertExpectations(t)
	evaluator.AssertExpectations(t)
}

func TestRuleManager_CreateRule_ValidationFailure(t *testing.T) {
	rm, _, _, _, _ := createTestRuleManager()
	rule := createTestRule()
	rule.Name = "" // Invalid rule

	err := rm.CreateRule(context.Background(), rule)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rule validation failed")
}

func TestRuleManager_GetRule(t *testing.T) {
	rm, storage, _, _, _ := createTestRuleManager()
	rule := createTestRule()

	storage.On("GetRule", mock.Anything, "testorg", "test-rule-001").Return(rule, nil)

	result, err := rm.GetRule(context.Background(), "testorg", "test-rule-001")

	assert.NoError(t, err)
	assert.Equal(t, rule.ID, result.ID)
	assert.Equal(t, rule.Name, result.Name)

	storage.AssertExpectations(t)
}

func TestRuleManager_ListRules(t *testing.T) {
	rm, storage, _, _, _ := createTestRuleManager()
	rules := []*AutomationRule{createTestRule()}

	storage.On("ListRules", mock.Anything, "testorg", mock.AnythingOfType("*github.RuleFilter")).Return(rules, nil)

	result, err := rm.ListRules(context.Background(), "testorg", &RuleFilter{})

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, rules[0].ID, result[0].ID)

	storage.AssertExpectations(t)
}

func TestRuleManager_UpdateRule(t *testing.T) {
	rm, storage, _, actionExecutor, evaluator := createTestRuleManager()
	rule := createTestRule()
	existingRule := createTestRule()

	storage.On("GetRule", mock.Anything, "testorg", "test-rule-001").Return(existingRule, nil)
	evaluator.On("ValidateConditions", &rule.Conditions).Return(&ConditionValidationResult{Valid: true}, nil)
	actionExecutor.On("ValidateAction", mock.Anything, &rule.Actions[0]).Return(nil)
	storage.On("UpdateRule", mock.Anything, mock.AnythingOfType("*github.AutomationRule")).Return(nil)

	err := rm.UpdateRule(context.Background(), rule)

	assert.NoError(t, err)
	assert.Equal(t, existingRule.CreatedAt, rule.CreatedAt)
	assert.Equal(t, existingRule.CreatedBy, rule.CreatedBy)

	storage.AssertExpectations(t)
	actionExecutor.AssertExpectations(t)
	evaluator.AssertExpectations(t)
}

func TestRuleManager_DeleteRule(t *testing.T) {
	rm, storage, _, _, _ := createTestRuleManager()

	storage.On("DeleteRule", mock.Anything, "testorg", "test-rule-001").Return(nil)

	err := rm.DeleteRule(context.Background(), "testorg", "test-rule-001")

	assert.NoError(t, err)
	storage.AssertExpectations(t)
}

func TestRuleManager_EnableDisableRule(t *testing.T) {
	rm, storage, _, actionExecutor, evaluator := createTestRuleManager()
	rule := createTestRule()
	rule.Enabled = false

	// Test Enable
	storage.On("GetRule", mock.Anything, "testorg", "test-rule-001").Return(rule, nil)
	evaluator.On("ValidateConditions", &rule.Conditions).Return(&ConditionValidationResult{Valid: true}, nil)
	actionExecutor.On("ValidateAction", mock.Anything, &rule.Actions[0]).Return(nil)
	storage.On("UpdateRule", mock.Anything, mock.MatchedBy(func(r *AutomationRule) bool {
		return r.Enabled == true
	})).Return(nil)

	err := rm.EnableRule(context.Background(), "testorg", "test-rule-001")
	assert.NoError(t, err)

	// Test Disable
	rule.Enabled = true
	storage.On("GetRule", mock.Anything, "testorg", "test-rule-001").Return(rule, nil)
	evaluator.On("ValidateConditions", &rule.Conditions).Return(&ConditionValidationResult{Valid: true}, nil)
	actionExecutor.On("ValidateAction", mock.Anything, &rule.Actions[0]).Return(nil)
	storage.On("UpdateRule", mock.Anything, mock.MatchedBy(func(r *AutomationRule) bool {
		return r.Enabled == false
	})).Return(nil)

	err = rm.DisableRule(context.Background(), "testorg", "test-rule-001")
	assert.NoError(t, err)

	storage.AssertExpectations(t)
}

func TestRuleManager_EvaluateConditions(t *testing.T) {
	rm, _, _, _, evaluator := createTestRuleManager()
	rule := createTestRule()
	event := createTestEventForRuleManager()

	evaluationResult := &EvaluationResult{
		Matched:           true,
		MatchedConditions: []string{"event_conditions"},
		EvaluationTime:    10 * time.Millisecond,
	}

	evaluator.On("EvaluateConditions", mock.Anything, &rule.Conditions, event, mock.AnythingOfType("*github.EvaluationContext")).Return(evaluationResult, nil)

	result, err := rm.EvaluateConditions(context.Background(), rule, event)

	assert.NoError(t, err)
	assert.True(t, result)
	evaluator.AssertExpectations(t)
}

func TestRuleManager_EvaluateConditions_DisabledRule(t *testing.T) {
	rm, _, _, _, _ := createTestRuleManager()
	rule := createTestRule()
	rule.Enabled = false
	event := createTestEventForRuleManager()

	result, err := rm.EvaluateConditions(context.Background(), rule, event)

	assert.NoError(t, err)
	assert.False(t, result)
}

func TestRuleManager_ExecuteRule(t *testing.T) {
	rm, storage, _, actionExecutor, _ := createTestRuleManager()
	rule := createTestRule()
	event := createTestEventForRuleManager()

	execContext := &AutomationExecutionContext{
		Event:        event,
		Organization: event.Organization,
		User:         event.Sender,
		Variables:    make(map[string]interface{}),
	}

	actionResult := &ActionExecutionResult{
		Result:     map[string]interface{}{"success": true},
		RetryCount: 0,
	}

	actionExecutor.On("ExecuteAction", mock.Anything, &rule.Actions[0], execContext).Return(actionResult, nil)
	storage.On("SaveExecution", mock.Anything, mock.AnythingOfType("*github.AutomationRuleExecution")).Return(nil).Twice()

	execution, err := rm.ExecuteRule(context.Background(), rule, execContext)

	assert.NoError(t, err)
	assert.NotNil(t, execution)
	assert.Equal(t, ExecutionStatusCompleted, execution.Status)
	assert.Len(t, execution.Actions, 1)
	assert.Equal(t, ExecutionStatusCompleted, execution.Actions[0].Status)

	storage.AssertExpectations(t)
	actionExecutor.AssertExpectations(t)
}

func TestRuleManager_CreateRuleSet(t *testing.T) {
	rm, storage, _, actionExecutor, evaluator := createTestRuleManager()
	ruleSet := &AutomationRuleSet{
		ID:           "test-set-001",
		Name:         "Test Rule Set",
		Organization: "testorg",
		Rules:        []AutomationRule{*createTestRule()},
		Enabled:      true,
	}

	evaluator.On("ValidateConditions", mock.AnythingOfType("*github.AutomationConditions")).Return(&ConditionValidationResult{Valid: true}, nil)
	actionExecutor.On("ValidateAction", mock.Anything, mock.AnythingOfType("*github.AutomationAction")).Return(nil)
	storage.On("CreateRuleSet", mock.Anything, mock.AnythingOfType("*github.AutomationRuleSet")).Return(nil)

	err := rm.CreateRuleSet(context.Background(), ruleSet)

	assert.NoError(t, err)
	assert.False(t, ruleSet.CreatedAt.IsZero())
	storage.AssertExpectations(t)
}

func TestRuleManager_CreateTemplate(t *testing.T) {
	rm, _, templateStorage, _, _ := createTestRuleManager()
	template := &AutomationRuleTemplate{
		ID:        "test-template-001",
		Name:      "Test Template",
		Category:  "testing",
		Template:  *createTestRule(),
		Variables: []TemplateVariable{},
		CreatedBy: "test-user",
	}

	templateStorage.On("CreateTemplate", mock.Anything, mock.AnythingOfType("*github.AutomationRuleTemplate")).Return(nil)

	err := rm.CreateTemplate(context.Background(), template)

	assert.NoError(t, err)
	assert.False(t, template.CreatedAt.IsZero())
	templateStorage.AssertExpectations(t)
}

func TestRuleManager_InstantiateTemplate(t *testing.T) {
	rm, _, templateStorage, _, _ := createTestRuleManager()
	template := &AutomationRuleTemplate{
		ID:       "test-template-001",
		Name:     "Test Template",
		Category: "testing",
		Template: *createTestRule(),
		Variables: []TemplateVariable{
			{
				Name:     "organization",
				Type:     "string",
				Required: true,
			},
			{
				Name:         "priority",
				Type:         "number",
				Required:     false,
				DefaultValue: 100,
			},
		},
	}

	templateStorage.On("GetTemplate", mock.Anything, "test-template-001").Return(template, nil)

	variables := map[string]interface{}{
		"organization": "myorg",
		"priority":     150,
	}

	rule, err := rm.InstantiateTemplate(context.Background(), "test-template-001", variables)

	assert.NoError(t, err)
	assert.NotNil(t, rule)
	assert.NotEqual(t, template.Template.ID, rule.ID) // Should have new ID
	templateStorage.AssertExpectations(t)
}

func TestRuleManager_InstantiateTemplate_MissingRequiredVariable(t *testing.T) {
	rm, _, templateStorage, _, _ := createTestRuleManager()
	template := &AutomationRuleTemplate{
		ID:       "test-template-001",
		Name:     "Test Template",
		Category: "testing",
		Template: *createTestRule(),
		Variables: []TemplateVariable{
			{
				Name:     "organization",
				Type:     "string",
				Required: true,
			},
		},
	}

	templateStorage.On("GetTemplate", mock.Anything, "test-template-001").Return(template, nil)

	// Missing required variable
	variables := map[string]interface{}{}

	_, err := rm.InstantiateTemplate(context.Background(), "test-template-001", variables)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required variable 'organization' not provided")
	templateStorage.AssertExpectations(t)
}

func TestRuleManager_ValidateRule(t *testing.T) {
	rm, _, _, actionExecutor, evaluator := createTestRuleManager()
	rule := createTestRule()

	evaluator.On("ValidateConditions", &rule.Conditions).Return(&ConditionValidationResult{Valid: true}, nil)
	actionExecutor.On("ValidateAction", mock.Anything, &rule.Actions[0]).Return(nil)

	result, err := rm.ValidateRule(context.Background(), rule)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Valid)
	assert.Equal(t, 100, result.Score)
	assert.Len(t, result.Errors, 0)

	evaluator.AssertExpectations(t)
	actionExecutor.AssertExpectations(t)
}

func TestRuleManager_TestRule(t *testing.T) {
	rm, _, _, _, evaluator := createTestRuleManager()
	rule := createTestRule()
	event := createTestEventForRuleManager()

	evaluationResult := &EvaluationResult{
		Matched:           true,
		MatchedConditions: []string{"event_conditions"},
		EvaluationTime:    10 * time.Millisecond,
	}

	evaluator.On("EvaluateConditions", mock.Anything, &rule.Conditions, event, mock.AnythingOfType("*github.EvaluationContext")).Return(evaluationResult, nil)

	result, err := rm.TestRule(context.Background(), rule, event)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.ConditionsMatched)
	assert.Len(t, result.ActionsExecuted, 1)
	if simulated, ok := result.ActionsExecuted[0].Result["simulated"].(bool); ok {
		assert.True(t, simulated)
	} else {
		t.Errorf("Expected simulated to be a bool, got %T", result.ActionsExecuted[0].Result["simulated"])
	}

	evaluator.AssertExpectations(t)
}

func TestRuleManager_CancelExecution(t *testing.T) {
	rm, storage, _, _, _ := createTestRuleManager()
	execution := &AutomationRuleExecution{
		ID:        "test-execution-001",
		RuleID:    "test-rule-001",
		Status:    ExecutionStatusRunning,
		StartedAt: time.Now(),
	}

	storage.On("GetExecution", mock.Anything, "test-execution-001").Return(execution, nil)
	storage.On("SaveExecution", mock.Anything, mock.MatchedBy(func(e *AutomationRuleExecution) bool {
		return e.Status == ExecutionStatusCancelled
	})).Return(nil)

	err := rm.CancelExecution(context.Background(), "test-execution-001")

	assert.NoError(t, err)
	storage.AssertExpectations(t)
}

func TestRuleManager_CancelExecution_InvalidStatus(t *testing.T) {
	rm, storage, _, _, _ := createTestRuleManager()
	execution := &AutomationRuleExecution{
		ID:     "test-execution-001",
		RuleID: "test-rule-001",
		Status: ExecutionStatusCompleted,
	}

	storage.On("GetExecution", mock.Anything, "test-execution-001").Return(execution, nil)

	err := rm.CancelExecution(context.Background(), "test-execution-001")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot cancel execution in status")
	storage.AssertExpectations(t)
}

// Benchmark tests

func BenchmarkRuleManager_CreateRule(b *testing.B) {
	rm, storage, _, actionExecutor, evaluator := createTestRuleManager()

	evaluator.On("ValidateConditions", mock.AnythingOfType("*github.AutomationConditions")).Return(&ConditionValidationResult{Valid: true}, nil)
	actionExecutor.On("ValidateAction", mock.Anything, mock.AnythingOfType("*github.AutomationAction")).Return(nil)
	storage.On("CreateRule", mock.Anything, mock.AnythingOfType("*github.AutomationRule")).Return(nil)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rule := createTestRule()
		rule.ID = "" // Reset ID for creation
		rm.CreateRule(context.Background(), rule)
	}
}

func BenchmarkRuleManager_EvaluateConditions(b *testing.B) {
	rm, _, _, _, evaluator := createTestRuleManager()
	rule := createTestRule()
	event := createTestEventForRuleManager()

	evaluationResult := &EvaluationResult{
		Matched:           true,
		MatchedConditions: []string{"event_conditions"},
		EvaluationTime:    10 * time.Millisecond,
	}

	evaluator.On("EvaluateConditions", mock.Anything, mock.AnythingOfType("*github.AutomationConditions"), mock.AnythingOfType("*github.GitHubEvent"), mock.AnythingOfType("*github.EvaluationContext")).Return(evaluationResult, nil)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rm.EvaluateConditions(context.Background(), rule, event)
	}
}

// Integration test

func TestRuleManager_Integration(t *testing.T) {
	rm, storage, _, actionExecutor, evaluator := createTestRuleManager()

	// Create a rule
	rule := createTestRule()
	rule.ID = ""

	evaluator.On("ValidateConditions", &rule.Conditions).Return(&ConditionValidationResult{Valid: true}, nil)
	actionExecutor.On("ValidateAction", mock.Anything, &rule.Actions[0]).Return(nil)
	storage.On("CreateRule", mock.Anything, mock.AnythingOfType("*github.AutomationRule")).Return(nil)

	err := rm.CreateRule(context.Background(), rule)
	require.NoError(t, err)

	// Get the rule
	storage.On("GetRule", mock.Anything, rule.Organization, rule.ID).Return(rule, nil)
	retrievedRule, err := rm.GetRule(context.Background(), rule.Organization, rule.ID)
	require.NoError(t, err)
	assert.Equal(t, rule.ID, retrievedRule.ID)

	// Test the rule
	event := createTestEventForRuleManager()
	evaluationResult := &EvaluationResult{
		Matched:           true,
		MatchedConditions: []string{"event_conditions"},
		EvaluationTime:    10 * time.Millisecond,
	}

	evaluator.On("EvaluateConditions", mock.Anything, &rule.Conditions, event, mock.AnythingOfType("*github.EvaluationContext")).Return(evaluationResult, nil)

	testResult, err := rm.TestRule(context.Background(), rule, event)
	require.NoError(t, err)
	assert.True(t, testResult.ConditionsMatched)

	// Delete the rule
	storage.On("DeleteRule", mock.Anything, rule.Organization, rule.ID).Return(nil)
	err = rm.DeleteRule(context.Background(), rule.Organization, rule.ID)
	require.NoError(t, err)

	storage.AssertExpectations(t)
	actionExecutor.AssertExpectations(t)
	evaluator.AssertExpectations(t)
}
