package github

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Test helper functions
func createTestAutomationRule() *AutomationRule {
	return &AutomationRule{
		ID:           "test-rule-001",
		Name:         "Test Automation Rule",
		Description:  "Test rule for unit testing",
		Organization: "testorg",
		Enabled:      true,
		Priority:     100,
		Conditions: AutomationConditions{
			EventTypes:         []EventType{EventTypePush, EventTypePullRequest},
			Actions:            []EventAction{ActionOpened, ActionClosed},
			Organization:       "testorg",
			Repository:         "testrepo",
			RepositoryPatterns: []string{"^test-.*"},
			Languages:          []string{"go", "javascript"},
			Topics:             []string{"api", "backend"},
			Visibility:         []string{"public"},
			LogicalOperator:    ConditionOperatorAND,
			PayloadMatch: []PayloadMatcher{
				{
					Path:          "$.pull_request.title",
					Operator:      MatchOperatorContains,
					Value:         "fix",
					CaseSensitive: false,
				},
			},
		},
		Actions: []AutomationAction{
			{
				ID:          "test-action-001",
				Type:        ActionTypeAddLabel,
				Name:        "Add Test Label",
				Description: "Add a test label to the PR",
				Enabled:     true,
				Parameters: map[string]interface{}{
					"labels": []string{"test", "automated"},
				},
				Timeout: 30 * time.Second,
				RetryPolicy: &ActionRetryPolicy{
					MaxRetries:    3,
					RetryInterval: 10 * time.Second,
					BackoffFactor: 2.0,
					MaxInterval:   60 * time.Second,
				},
				OnFailure: ActionFailurePolicyContinue,
			},
		},
		Metadata: AutomationRuleMetadata{
			Version:     "1.0",
			Category:    "testing",
			Environment: "test",
			Owner:       "test-team",
			Team:        "platform",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		CreatedBy: "test-user",
		Tags: map[string]string{
			"environment": "test",
			"team":        "platform",
		},
	}
}

func createTestRuleExecution() *AutomationRuleExecution {
	return &AutomationRuleExecution{
		ID:             "test-execution-001",
		RuleID:         "test-rule-001",
		TriggerEventID: "test-event-001",
		StartedAt:      time.Now(),
		Status:         ExecutionStatusRunning,
		TriggerType:    ExecutionTriggerTypeEvent,
		Context: AutomationExecutionContext{
			Event: &GitHubEvent{
				ID:           "test-event-001",
				Type:         "pull_request",
				Action:       "opened",
				Organization: "testorg",
				Repository:   "testrepo",
				Sender:       "testuser",
				Timestamp:    time.Now(),
			},
			Organization: "testorg",
			User:         "testuser",
			Variables: map[string]interface{}{
				"pr_number": 123,
				"branch":    "feature/test",
			},
		},
		Actions: []ActionExecutionResult{
			{
				ActionID:   "test-action-001",
				ActionType: ActionTypeAddLabel,
				Status:     ExecutionStatusCompleted,
				StartedAt:  time.Now().Add(-5 * time.Minute),
				Duration:   2 * time.Second,
				Result: map[string]interface{}{
					"labels_added": []string{"test", "automated"},
				},
			},
		},
	}
}

func createTestRuleTemplate() *AutomationRuleTemplate {
	return &AutomationRuleTemplate{
		ID:          "test-template-001",
		Name:        "Test Rule Template",
		Description: "Template for testing automation rules",
		Category:    "testing",
		Template:    *createTestAutomationRule(),
		Variables: []TemplateVariable{
			{
				Name:        "organization",
				Type:        "string",
				Description: "GitHub organization name",
				Required:    true,
			},
			{
				Name:         "labels",
				Type:         "array",
				Description:  "Labels to add",
				Required:     false,
				DefaultValue: []string{"automated"},
			},
			{
				Name:         "priority",
				Type:         "number",
				Description:  "Rule priority",
				Required:     false,
				DefaultValue: 100,
				Validation:   "min:1,max:1000",
			},
		},
		Examples: []TemplateExample{
			{
				Name:        "Basic Setup",
				Description: "Basic automation rule setup",
				Variables: map[string]interface{}{
					"organization": "myorg",
					"labels":       []string{"review-required", "automated"},
					"priority":     150,
				},
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		CreatedBy: "template-creator",
		Tags: map[string]string{
			"category": "testing",
			"type":     "template",
		},
	}
}

func TestAutomationRule_BasicStructure(t *testing.T) {
	rule := createTestAutomationRule()

	assert.Equal(t, "test-rule-001", rule.ID)
	assert.Equal(t, "Test Automation Rule", rule.Name)
	assert.Equal(t, "testorg", rule.Organization)
	assert.True(t, rule.Enabled)
	assert.Equal(t, 100, rule.Priority)
	assert.Equal(t, ConditionOperatorAND, rule.Conditions.LogicalOperator)
	assert.Len(t, rule.Actions, 1)
	assert.Equal(t, ActionTypeAddLabel, rule.Actions[0].Type)
}

func TestAutomationConditions_EventMatching(t *testing.T) {
	conditions := AutomationConditions{
		EventTypes:      []EventType{EventTypePush, EventTypePullRequest},
		Actions:         []EventAction{ActionOpened, ActionClosed},
		Organization:    "testorg",
		Repository:      "testrepo",
		LogicalOperator: ConditionOperatorAND,
	}

	assert.Contains(t, conditions.EventTypes, EventTypePush)
	assert.Contains(t, conditions.EventTypes, EventTypePullRequest)
	assert.Contains(t, conditions.Actions, ActionOpened)
	assert.Contains(t, conditions.Actions, ActionClosed)
	assert.Equal(t, "testorg", conditions.Organization)
}

func TestPayloadMatcher_Structure(t *testing.T) {
	matcher := PayloadMatcher{
		Path:          "$.pull_request.title",
		Operator:      MatchOperatorContains,
		Value:         "fix",
		CaseSensitive: false,
	}

	assert.Equal(t, "$.pull_request.title", matcher.Path)
	assert.Equal(t, MatchOperatorContains, matcher.Operator)
	assert.Equal(t, "fix", matcher.Value)
	assert.False(t, matcher.CaseSensitive)
}

func TestAutomationAction_RetryPolicy(t *testing.T) {
	action := AutomationAction{
		ID:      "test-action",
		Type:    ActionTypeWebhook,
		Enabled: true,
		RetryPolicy: &ActionRetryPolicy{
			MaxRetries:    3,
			RetryInterval: 10 * time.Second,
			BackoffFactor: 2.0,
			MaxInterval:   60 * time.Second,
		},
		OnFailure: ActionFailurePolicyContinue,
	}

	assert.NotNil(t, action.RetryPolicy)
	assert.Equal(t, 3, action.RetryPolicy.MaxRetries)
	assert.Equal(t, 10*time.Second, action.RetryPolicy.RetryInterval)
	assert.Equal(t, 2.0, action.RetryPolicy.BackoffFactor)
	assert.Equal(t, ActionFailurePolicyContinue, action.OnFailure)
}

func TestAutomationRuleExecution_Structure(t *testing.T) {
	execution := createTestRuleExecution()

	assert.Equal(t, "test-execution-001", execution.ID)
	assert.Equal(t, "test-rule-001", execution.RuleID)
	assert.Equal(t, ExecutionStatusRunning, execution.Status)
	assert.Equal(t, ExecutionTriggerTypeEvent, execution.TriggerType)
	assert.NotNil(t, execution.Context.Event)
	assert.Len(t, execution.Actions, 1)
	assert.Equal(t, ExecutionStatusCompleted, execution.Actions[0].Status)
}

func TestActionExecutionResult_Structure(t *testing.T) {
	result := ActionExecutionResult{
		ActionID:   "test-action-001",
		ActionType: ActionTypeAddLabel,
		Status:     ExecutionStatusCompleted,
		StartedAt:  time.Now(),
		Duration:   2 * time.Second,
		Result: map[string]interface{}{
			"labels_added": []string{"test", "automated"},
		},
		RetryCount: 0,
	}

	assert.Equal(t, "test-action-001", result.ActionID)
	assert.Equal(t, ActionTypeAddLabel, result.ActionType)
	assert.Equal(t, ExecutionStatusCompleted, result.Status)
	assert.Equal(t, 2*time.Second, result.Duration)
	assert.Equal(t, 0, result.RetryCount)
	assert.Contains(t, result.Result, "labels_added")
}

func TestAutomationRuleTemplate_Variables(t *testing.T) {
	template := createTestRuleTemplate()

	assert.Equal(t, "test-template-001", template.ID)
	assert.Equal(t, "testing", template.Category)
	assert.Len(t, template.Variables, 3)
	assert.Len(t, template.Examples, 1)

	// Test variables
	orgVar := template.Variables[0]
	assert.Equal(t, "organization", orgVar.Name)
	assert.Equal(t, "string", orgVar.Type)
	assert.True(t, orgVar.Required)

	labelsVar := template.Variables[1]
	assert.Equal(t, "labels", labelsVar.Name)
	assert.Equal(t, "array", labelsVar.Type)
	assert.False(t, labelsVar.Required)
	assert.NotNil(t, labelsVar.DefaultValue)

	priorityVar := template.Variables[2]
	assert.Equal(t, "priority", priorityVar.Name)
	assert.Equal(t, "number", priorityVar.Type)
	assert.Equal(t, "min:1,max:1000", priorityVar.Validation)
}

func TestTemplateExample_Structure(t *testing.T) {
	template := createTestRuleTemplate()
	example := template.Examples[0]

	assert.Equal(t, "Basic Setup", example.Name)
	assert.Equal(t, "Basic automation rule setup", example.Description)
	assert.Contains(t, example.Variables, "organization")
	assert.Contains(t, example.Variables, "labels")
	assert.Contains(t, example.Variables, "priority")
	assert.Equal(t, "myorg", example.Variables["organization"])
	assert.Equal(t, 150, example.Variables["priority"])
}

func TestRuleFilter_Structure(t *testing.T) {
	filter := RuleFilter{
		Organization: "testorg",
		Enabled:      boolPtr(true),
		Tags:         []string{"test", "automation"},
		Category:     "testing",
		EventTypes:   []EventType{EventTypePush, EventTypePullRequest},
		CreatedBy:    "test-user",
	}

	assert.Equal(t, "testorg", filter.Organization)
	assert.NotNil(t, filter.Enabled)
	assert.True(t, *filter.Enabled)
	assert.Contains(t, filter.Tags, "test")
	assert.Contains(t, filter.EventTypes, EventTypePush)
}

func TestExecutionFilter_Structure(t *testing.T) {
	now := time.Now()
	filter := ExecutionFilter{
		RuleID:       "test-rule-001",
		Status:       ExecutionStatusCompleted,
		TriggerType:  ExecutionTriggerTypeEvent,
		StartedAfter: &now,
	}

	assert.Equal(t, "test-rule-001", filter.RuleID)
	assert.Equal(t, ExecutionStatusCompleted, filter.Status)
	assert.Equal(t, ExecutionTriggerTypeEvent, filter.TriggerType)
	assert.NotNil(t, filter.StartedAfter)
}

func TestRuleValidationResult_Structure(t *testing.T) {
	result := RuleValidationResult{
		Valid: true,
		Errors: []RuleValidationError{
			{
				Field:      "conditions.organization",
				Message:    "Organization is required",
				Severity:   "error",
				Suggestion: "Please specify an organization name",
			},
		},
		Warnings: []RuleValidationWarning{
			{
				Field:      "actions[0].timeout",
				Message:    "Timeout is very long",
				Suggestion: "Consider reducing timeout value",
			},
		},
		Score: 85,
	}

	assert.True(t, result.Valid)
	assert.Len(t, result.Errors, 1)
	assert.Len(t, result.Warnings, 1)
	assert.Equal(t, 85, result.Score)

	error := result.Errors[0]
	assert.Equal(t, "conditions.organization", error.Field)
	assert.Equal(t, "error", error.Severity)

	warning := result.Warnings[0]
	assert.Equal(t, "actions[0].timeout", warning.Field)
	assert.Contains(t, warning.Message, "Timeout")
}

func TestRuleTestResult_Structure(t *testing.T) {
	result := RuleTestResult{
		RuleID:            "test-rule-001",
		ConditionsMatched: true,
		ActionsExecuted: []ActionExecutionResult{
			{
				ActionID:   "test-action-001",
				ActionType: ActionTypeAddLabel,
				Status:     ExecutionStatusCompleted,
				Duration:   1 * time.Second,
			},
		},
		ExecutionTime: 2 * time.Second,
		Errors:        []string{},
		Context: AutomationExecutionContext{
			Organization: "testorg",
			User:         "testuser",
		},
	}

	assert.Equal(t, "test-rule-001", result.RuleID)
	assert.True(t, result.ConditionsMatched)
	assert.Len(t, result.ActionsExecuted, 1)
	assert.Equal(t, 2*time.Second, result.ExecutionTime)
	assert.Empty(t, result.Errors)
}

func TestMatchOperators_Constants(t *testing.T) {
	operators := []MatchOperator{
		MatchOperatorEquals,
		MatchOperatorNotEquals,
		MatchOperatorContains,
		MatchOperatorNotContains,
		MatchOperatorStartsWith,
		MatchOperatorEndsWith,
		MatchOperatorRegex,
		MatchOperatorGreaterThan,
		MatchOperatorLessThan,
		MatchOperatorExists,
		MatchOperatorNotExists,
		MatchOperatorEmpty,
		MatchOperatorNotEmpty,
	}

	assert.Len(t, operators, 13)
	assert.Equal(t, "equals", string(MatchOperatorEquals))
	assert.Equal(t, "regex", string(MatchOperatorRegex))
	assert.Equal(t, "not_exists", string(MatchOperatorNotExists))
}

func TestActionTypes_Constants(t *testing.T) {
	actionTypes := []ActionType{
		ActionTypeWebhook,
		ActionTypeHTTPRequest,
		ActionTypeCreateIssue,
		ActionTypeCreatePR,
		ActionTypeAddLabel,
		ActionTypeRemoveLabel,
		ActionTypeAssignReviewer,
		ActionTypeMergePR,
		ActionTypeClosePR,
		ActionTypeCloseIssue,
		ActionTypeCreateBranch,
		ActionTypeDeleteBranch,
		ActionTypeProtectBranch,
		ActionTypeCreateTag,
		ActionTypeCreateRelease,
		ActionTypeSlackMessage,
		ActionTypeTeamsMessage,
		ActionTypeEmail,
		ActionTypeSMS,
		ActionTypeTriggerWorkflow,
		ActionTypeRunScript,
		ActionTypeDeployment,
		ActionTypeCustom,
	}

	assert.Len(t, actionTypes, 23)
	assert.Equal(t, "webhook", string(ActionTypeWebhook))
	assert.Equal(t, "create_issue", string(ActionTypeCreateIssue))
	assert.Equal(t, "slack_message", string(ActionTypeSlackMessage))
	assert.Equal(t, "trigger_workflow", string(ActionTypeTriggerWorkflow))
}

func TestExecutionStatus_Constants(t *testing.T) {
	statuses := []ExecutionStatus{
		ExecutionStatusPending,
		ExecutionStatusRunning,
		ExecutionStatusCompleted,
		ExecutionStatusFailed,
		ExecutionStatusCancelled,
		ExecutionStatusTimeout,
	}

	assert.Len(t, statuses, 6)
	assert.Equal(t, "pending", string(ExecutionStatusPending))
	assert.Equal(t, "completed", string(ExecutionStatusCompleted))
	assert.Equal(t, "timeout", string(ExecutionStatusTimeout))
}

func TestConditionOperators_Constants(t *testing.T) {
	operators := []ConditionOperator{
		ConditionOperatorAND,
		ConditionOperatorOR,
		ConditionOperatorNOT,
	}

	assert.Len(t, operators, 3)
	assert.Equal(t, "AND", string(ConditionOperatorAND))
	assert.Equal(t, "OR", string(ConditionOperatorOR))
	assert.Equal(t, "NOT", string(ConditionOperatorNOT))
}

// Helper functions
// boolPtr is defined in automation_engine.go

// Benchmark tests
func BenchmarkAutomationRule_Creation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		rule := createTestAutomationRule()
		_ = rule
	}
}

func BenchmarkPayloadMatcher_Creation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		matcher := PayloadMatcher{
			Path:          "$.pull_request.title",
			Operator:      MatchOperatorContains,
			Value:         "fix",
			CaseSensitive: false,
		}
		_ = matcher
	}
}

func BenchmarkRuleExecution_Creation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		execution := createTestRuleExecution()
		_ = execution
	}
}
