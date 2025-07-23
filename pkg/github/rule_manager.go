package github

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// RuleManager implements the AutomationRuleService interface.
type RuleManager struct {
	logger          Logger
	apiClient       APIClient
	evaluator       ConditionEvaluator
	actionExecutor  ActionExecutor
	storage         RuleStorage
	templateStorage TemplateStorage
	mu              sync.RWMutex
	ruleCache       map[string]*AutomationRule
	enabledRules    map[string]bool
}

// RuleStorage defines the interface for persisting automation rules.
type RuleStorage interface {
	// Rule operations
	CreateRule(ctx context.Context, rule *AutomationRule) error
	GetRule(ctx context.Context, org, ruleID string) (*AutomationRule, error)
	ListRules(ctx context.Context, org string, filter *RuleFilter) ([]*AutomationRule, error)
	UpdateRule(ctx context.Context, rule *AutomationRule) error
	DeleteRule(ctx context.Context, org, ruleID string) error

	// Rule Set operations
	CreateRuleSet(ctx context.Context, ruleSet *AutomationRuleSet) error
	GetRuleSet(ctx context.Context, org, setID string) (*AutomationRuleSet, error)
	ListRuleSets(ctx context.Context, org string) ([]*AutomationRuleSet, error)
	UpdateRuleSet(ctx context.Context, ruleSet *AutomationRuleSet) error
	DeleteRuleSet(ctx context.Context, org, setID string) error

	// Execution history
	SaveExecution(ctx context.Context, execution *AutomationRuleExecution) error
	GetExecution(ctx context.Context, executionID string) (*AutomationRuleExecution, error)
	ListExecutions(ctx context.Context, org string, filter *ExecutionFilter) ([]*AutomationRuleExecution, error)
}

// TemplateStorage defines the interface for persisting rule templates.
type TemplateStorage interface {
	CreateTemplate(ctx context.Context, template *AutomationRuleTemplate) error
	GetTemplate(ctx context.Context, templateID string) (*AutomationRuleTemplate, error)
	ListTemplates(ctx context.Context, category string) ([]*AutomationRuleTemplate, error)
	UpdateTemplate(ctx context.Context, template *AutomationRuleTemplate) error
	DeleteTemplate(ctx context.Context, templateID string) error
}

// ActionExecutor defines the interface for executing automation actions.
type ActionExecutor interface {
	ExecuteAction(ctx context.Context, action *AutomationAction, context *AutomationExecutionContext) (*ActionExecutionResult, error)
	ValidateAction(ctx context.Context, action *AutomationAction) error
	GetSupportedActions() []ActionType
}

// NewRuleManager creates a new rule manager instance.
func NewRuleManager(logger Logger, apiClient APIClient, evaluator ConditionEvaluator, actionExecutor ActionExecutor, storage RuleStorage, templateStorage TemplateStorage) *RuleManager {
	return &RuleManager{
		logger:          logger,
		apiClient:       apiClient,
		evaluator:       evaluator,
		actionExecutor:  actionExecutor,
		storage:         storage,
		templateStorage: templateStorage,
		ruleCache:       make(map[string]*AutomationRule),
		enabledRules:    make(map[string]bool),
	}
}

// CreateRule creates a new automation rule.
func (rm *RuleManager) CreateRule(ctx context.Context, rule *AutomationRule) error {
	rm.logger.Info("Creating automation rule", "rule_id", rule.ID, "organization", rule.Organization)

	// Validate rule
	if err := rm.validateRule(rule); err != nil {
		return fmt.Errorf("rule validation failed: %w", err)
	}

	// Set metadata
	now := time.Now()

	if rule.ID == "" {
		rule.ID = uuid.New().String()
	}

	rule.CreatedAt = now
	rule.UpdatedAt = now

	// Set default priority if not specified
	if rule.Priority == 0 {
		rule.Priority = 100
	}

	// Validate conditions
	validationResult, err := rm.evaluator.ValidateConditions(&rule.Conditions)
	if err != nil {
		return fmt.Errorf("failed to validate conditions: %w", err)
	}

	if !validationResult.Valid {
		return fmt.Errorf("rule conditions are invalid: %v", validationResult.Errors)
	}

	// Validate actions
	for i, action := range rule.Actions {
		if err := rm.actionExecutor.ValidateAction(ctx, &action); err != nil {
			return fmt.Errorf("action %d validation failed: %w", i, err)
		}
	}

	// Save to storage
	if err := rm.storage.CreateRule(ctx, rule); err != nil {
		return fmt.Errorf("failed to save rule: %w", err)
	}

	// Update cache
	rm.mu.Lock()
	rm.ruleCache[rm.cacheKey(rule.Organization, rule.ID)] = rule
	rm.enabledRules[rm.cacheKey(rule.Organization, rule.ID)] = rule.Enabled
	rm.mu.Unlock()

	rm.logger.Info("Automation rule created successfully", "rule_id", rule.ID)

	return nil
}

// GetRule retrieves an automation rule by ID.
func (rm *RuleManager) GetRule(ctx context.Context, org, ruleID string) (*AutomationRule, error) {
	cacheKey := rm.cacheKey(org, ruleID)

	// Check cache first
	rm.mu.RLock()

	if rule, exists := rm.ruleCache[cacheKey]; exists {
		rm.mu.RUnlock()
		return rule, nil
	}

	rm.mu.RUnlock()

	// Load from storage
	rule, err := rm.storage.GetRule(ctx, org, ruleID)
	if err != nil {
		return nil, err
	}

	// Update cache
	rm.mu.Lock()
	rm.ruleCache[cacheKey] = rule
	rm.enabledRules[cacheKey] = rule.Enabled
	rm.mu.Unlock()

	return rule, nil
}

// ListRules lists automation rules with optional filtering.
func (rm *RuleManager) ListRules(ctx context.Context, org string, filter *RuleFilter) ([]*AutomationRule, error) {
	rules, err := rm.storage.ListRules(ctx, org, filter)
	if err != nil {
		return nil, err
	}

	// Sort by priority (higher priority first) and creation time
	sort.Slice(rules, func(i, j int) bool {
		if rules[i].Priority != rules[j].Priority {
			return rules[i].Priority > rules[j].Priority
		}

		return rules[i].CreatedAt.Before(rules[j].CreatedAt)
	})

	// Update cache for retrieved rules
	rm.mu.Lock()

	for _, rule := range rules {
		cacheKey := rm.cacheKey(rule.Organization, rule.ID)
		rm.ruleCache[cacheKey] = rule
		rm.enabledRules[cacheKey] = rule.Enabled
	}

	rm.mu.Unlock()

	return rules, nil
}

// UpdateRule updates an existing automation rule.
func (rm *RuleManager) UpdateRule(ctx context.Context, rule *AutomationRule) error {
	rm.logger.Info("Updating automation rule", "rule_id", rule.ID, "organization", rule.Organization)

	// Validate rule
	if err := rm.validateRule(rule); err != nil {
		return fmt.Errorf("rule validation failed: %w", err)
	}

	// Check if rule exists
	existing, err := rm.storage.GetRule(ctx, rule.Organization, rule.ID)
	if err != nil {
		return fmt.Errorf("failed to get existing rule: %w", err)
	}

	// Preserve creation metadata
	rule.CreatedAt = existing.CreatedAt
	rule.CreatedBy = existing.CreatedBy
	rule.UpdatedAt = time.Now()

	// Validate conditions
	validationResult, err := rm.evaluator.ValidateConditions(&rule.Conditions)
	if err != nil {
		return fmt.Errorf("failed to validate conditions: %w", err)
	}

	if !validationResult.Valid {
		return fmt.Errorf("rule conditions are invalid: %v", validationResult.Errors)
	}

	// Validate actions
	for i, action := range rule.Actions {
		if err := rm.actionExecutor.ValidateAction(ctx, &action); err != nil {
			return fmt.Errorf("action %d validation failed: %w", i, err)
		}
	}

	// Save to storage
	if err := rm.storage.UpdateRule(ctx, rule); err != nil {
		return fmt.Errorf("failed to update rule: %w", err)
	}

	// Update cache
	cacheKey := rm.cacheKey(rule.Organization, rule.ID)
	rm.mu.Lock()
	rm.ruleCache[cacheKey] = rule
	rm.enabledRules[cacheKey] = rule.Enabled
	rm.mu.Unlock()

	rm.logger.Info("Automation rule updated successfully", "rule_id", rule.ID)

	return nil
}

// DeleteRule deletes an automation rule.
func (rm *RuleManager) DeleteRule(ctx context.Context, org, ruleID string) error {
	rm.logger.Info("Deleting automation rule", "rule_id", ruleID, "organization", org)

	// Delete from storage
	if err := rm.storage.DeleteRule(ctx, org, ruleID); err != nil {
		return fmt.Errorf("failed to delete rule: %w", err)
	}

	// Remove from cache
	cacheKey := rm.cacheKey(org, ruleID)
	rm.mu.Lock()
	delete(rm.ruleCache, cacheKey)
	delete(rm.enabledRules, cacheKey)
	rm.mu.Unlock()

	rm.logger.Info("Automation rule deleted successfully", "rule_id", ruleID)

	return nil
}

// EnableRule enables an automation rule.
func (rm *RuleManager) EnableRule(ctx context.Context, org, ruleID string) error {
	return rm.setRuleEnabled(ctx, org, ruleID, true)
}

// DisableRule disables an automation rule.
func (rm *RuleManager) DisableRule(ctx context.Context, org, ruleID string) error {
	return rm.setRuleEnabled(ctx, org, ruleID, false)
}

// setRuleEnabled sets the enabled status of a rule.
func (rm *RuleManager) setRuleEnabled(ctx context.Context, org, ruleID string, enabled bool) error {
	rule, err := rm.GetRule(ctx, org, ruleID)
	if err != nil {
		return err
	}

	rule.Enabled = enabled
	rule.UpdatedAt = time.Now()

	return rm.UpdateRule(ctx, rule)
}

// EvaluateConditions evaluates conditions for a rule against an event.
func (rm *RuleManager) EvaluateConditions(ctx context.Context, rule *AutomationRule, event *GitHubEvent) (bool, error) {
	// Check if rule is enabled
	if !rule.Enabled {
		return false, nil
	}

	// Create evaluation context
	evalContext := &EvaluationContext{
		Environment: rule.Metadata.Environment,
		Variables:   make(map[string]interface{}),
		Timezone:    time.UTC,
	}

	// Add repository info if available
	if event.Repository != "" {
		repoInfo, err := rm.getRepositoryInfo(ctx, event.Organization, event.Repository)
		if err != nil {
			rm.logger.Warn("Failed to get repository info", "repository", event.Repository, "error", err)
		} else {
			evalContext.Repository = repoInfo
		}
	}

	// Add organization info if available
	if event.Organization != "" {
		orgInfo, err := rm.getOrganizationInfo(ctx, event.Organization)
		if err != nil {
			rm.logger.Warn("Failed to get organization info", "organization", event.Organization, "error", err)
		} else {
			evalContext.Organization = orgInfo
		}
	}

	// Evaluate conditions
	result, err := rm.evaluator.EvaluateConditions(ctx, &rule.Conditions, event, evalContext)
	if err != nil {
		return false, fmt.Errorf("condition evaluation failed: %w", err)
	}

	rm.logger.Debug("Rule condition evaluation completed",
		"rule_id", rule.ID,
		"event_id", event.ID,
		"matched", result.Matched,
		"evaluation_time", result.EvaluationTime)

	return result.Matched, nil
}

// ExecuteRule executes a rule if conditions are met.
func (rm *RuleManager) ExecuteRule(ctx context.Context, rule *AutomationRule, execContext *AutomationExecutionContext) (*AutomationRuleExecution, error) {
	rm.logger.Info("Executing automation rule", "rule_id", rule.ID, "event_id", execContext.Event.ID)

	startTime := time.Now()
	execution := &AutomationRuleExecution{
		ID:          uuid.New().String(),
		RuleID:      rule.ID,
		StartedAt:   startTime,
		Status:      ExecutionStatusRunning,
		TriggerType: ExecutionTriggerTypeEvent,
		Context:     *execContext,
		Actions:     []ActionExecutionResult{},
		Metadata:    make(map[string]interface{}),
	}

	if execContext.Event != nil {
		execution.TriggerEventID = execContext.Event.ID
	}

	// Save initial execution state
	if err := rm.storage.SaveExecution(ctx, execution); err != nil {
		rm.logger.Error("Failed to save execution state", "execution_id", execution.ID, "error", err)
	}

	// Execute actions
	var executionErr error

	for _, action := range rule.Actions {
		if !action.Enabled {
			rm.logger.Debug("Skipping disabled action", "action_id", action.ID, "rule_id", rule.ID)
			continue
		}

		actionResult := ActionExecutionResult{
			ActionID:   action.ID,
			ActionType: action.Type,
			Status:     ExecutionStatusRunning,
			StartedAt:  time.Now(),
		}

		rm.logger.Debug("Executing action", "action_id", action.ID, "action_type", action.Type, "rule_id", rule.ID)

		// Execute the action
		result, err := rm.actionExecutor.ExecuteAction(ctx, &action, execContext)
		if err != nil {
			actionResult.Status = ExecutionStatusFailed
			actionResult.Error = err.Error()
			executionErr = err

			rm.logger.Error("Action execution failed",
				"action_id", action.ID,
				"rule_id", rule.ID,
				"error", err)

			// Handle failure policy
			switch action.OnFailure {
			case ActionFailurePolicyStop:
				rm.logger.Info("Stopping rule execution due to action failure", "rule_id", rule.ID)
				execution.Status = ExecutionStatusFailed
				endTime := time.Now()
				execution.CompletedAt = &endTime
				execution.Duration = endTime.Sub(startTime)
				return execution, fmt.Errorf("rule execution stopped due to action failure")
			case ActionFailurePolicyContinue:
				rm.logger.Info("Continuing rule execution despite action failure", "rule_id", rule.ID)
			case ActionFailurePolicyRetry:
				// Implement retry logic with exponential backoff
				if rm.shouldRetryAction(&action, &actionResult) {
					retryErr := rm.retryActionWithBackoff(ctx, &action, &actionResult, rule.ID)
					if retryErr != nil {
						rm.logger.Error("All retry attempts failed",
							"action_id", action.ID,
							"rule_id", rule.ID,
							"retry_count", actionResult.RetryCount,
							"error", retryErr)
						actionResult.Status = ExecutionStatusFailed
						actionResult.Error = fmt.Sprintf("Failed after %d retries: %v", actionResult.RetryCount, retryErr)
					} else {
						rm.logger.Info("Action succeeded after retry",
							"action_id", action.ID,
							"rule_id", rule.ID,
							"retry_count", actionResult.RetryCount)
						actionResult.Status = ExecutionStatusCompleted
						actionResult.Error = ""
					}
				} else {
					rm.logger.Info("Max retries exceeded, marking as failed",
						"action_id", action.ID,
						"rule_id", rule.ID,
						"retry_count", actionResult.RetryCount)
					actionResult.Status = ExecutionStatusFailed
				}
			case ActionFailurePolicySkip:
				rm.logger.Info("Skipping action and marking as failed", "rule_id", rule.ID)
				actionResult.Status = ExecutionStatusFailed
			default:
				rm.logger.Warn("Unknown failure policy, continuing", "policy", action.OnFailure)
			}
		} else {
			actionResult.Status = ExecutionStatusCompleted
			if result != nil {
				actionResult.Result = result.Result
				actionResult.RetryCount = result.RetryCount
			}
		}

		completedAt := time.Now()
		actionResult.CompletedAt = &completedAt
		actionResult.Duration = completedAt.Sub(actionResult.StartedAt)

		execution.Actions = append(execution.Actions, actionResult)

		// Break on stop policy
		if err != nil && action.OnFailure == ActionFailurePolicyStop {
			break
		}
	}

	// Update execution status
	completedAt := time.Now()
	execution.CompletedAt = &completedAt
	execution.Duration = completedAt.Sub(startTime)

	if executionErr != nil {
		execution.Status = ExecutionStatusFailed
		execution.Error = executionErr.Error()
	} else {
		execution.Status = ExecutionStatusCompleted
	}

	// Save final execution state
	if err := rm.storage.SaveExecution(ctx, execution); err != nil {
		rm.logger.Error("Failed to save final execution state", "execution_id", execution.ID, "error", err)
	}

	rm.logger.Info("Rule execution completed",
		"rule_id", rule.ID,
		"execution_id", execution.ID,
		"status", execution.Status,
		"duration", execution.Duration,
		"actions_executed", len(execution.Actions))

	return execution, executionErr
}

// Rule Set Management

// CreateRuleSet creates a new rule set.
func (rm *RuleManager) CreateRuleSet(ctx context.Context, ruleSet *AutomationRuleSet) error {
	rm.logger.Info("Creating rule set", "set_id", ruleSet.ID, "organization", ruleSet.Organization)

	// Validate rule set
	if err := rm.validateRuleSet(ruleSet); err != nil {
		return fmt.Errorf("rule set validation failed: %w", err)
	}

	// Set metadata
	now := time.Now()

	if ruleSet.ID == "" {
		ruleSet.ID = uuid.New().String()
	}

	ruleSet.CreatedAt = now
	ruleSet.UpdatedAt = now

	// Validate all rules in the set
	for i, rule := range ruleSet.Rules {
		if err := rm.validateRule(&rule); err != nil {
			return fmt.Errorf("rule %d validation failed: %w", i, err)
		}
	}

	return rm.storage.CreateRuleSet(ctx, ruleSet)
}

// GetRuleSet retrieves a rule set by ID.
func (rm *RuleManager) GetRuleSet(ctx context.Context, org, setID string) (*AutomationRuleSet, error) {
	return rm.storage.GetRuleSet(ctx, org, setID)
}

// ListRuleSets lists all rule sets for an organization.
func (rm *RuleManager) ListRuleSets(ctx context.Context, org string) ([]*AutomationRuleSet, error) {
	return rm.storage.ListRuleSets(ctx, org)
}

// UpdateRuleSet updates an existing rule set.
func (rm *RuleManager) UpdateRuleSet(ctx context.Context, ruleSet *AutomationRuleSet) error {
	rm.logger.Info("Updating rule set", "set_id", ruleSet.ID, "organization", ruleSet.Organization)

	// Validate rule set
	if err := rm.validateRuleSet(ruleSet); err != nil {
		return fmt.Errorf("rule set validation failed: %w", err)
	}

	// Get existing rule set to preserve metadata
	existing, err := rm.storage.GetRuleSet(ctx, ruleSet.Organization, ruleSet.ID)
	if err != nil {
		return fmt.Errorf("failed to get existing rule set: %w", err)
	}

	ruleSet.CreatedAt = existing.CreatedAt
	ruleSet.CreatedBy = existing.CreatedBy
	ruleSet.UpdatedAt = time.Now()

	// Validate all rules in the set
	for i, rule := range ruleSet.Rules {
		if err := rm.validateRule(&rule); err != nil {
			return fmt.Errorf("rule %d validation failed: %w", i, err)
		}
	}

	return rm.storage.UpdateRuleSet(ctx, ruleSet)
}

// DeleteRuleSet deletes a rule set.
func (rm *RuleManager) DeleteRuleSet(ctx context.Context, org, setID string) error {
	rm.logger.Info("Deleting rule set", "set_id", setID, "organization", org)
	return rm.storage.DeleteRuleSet(ctx, org, setID)
}

// Template Management

// CreateTemplate creates a new rule template.
func (rm *RuleManager) CreateTemplate(ctx context.Context, template *AutomationRuleTemplate) error {
	rm.logger.Info("Creating rule template", "template_id", template.ID)

	// Validate template
	if err := rm.validateTemplate(template); err != nil {
		return fmt.Errorf("template validation failed: %w", err)
	}

	// Set metadata
	now := time.Now()

	if template.ID == "" {
		template.ID = uuid.New().String()
	}

	template.CreatedAt = now
	template.UpdatedAt = now

	return rm.templateStorage.CreateTemplate(ctx, template)
}

// GetTemplate retrieves a template by ID.
func (rm *RuleManager) GetTemplate(ctx context.Context, templateID string) (*AutomationRuleTemplate, error) {
	return rm.templateStorage.GetTemplate(ctx, templateID)
}

// ListTemplates lists templates by category.
func (rm *RuleManager) ListTemplates(ctx context.Context, category string) ([]*AutomationRuleTemplate, error) {
	return rm.templateStorage.ListTemplates(ctx, category)
}

// UpdateTemplate updates an existing template.
func (rm *RuleManager) UpdateTemplate(ctx context.Context, template *AutomationRuleTemplate) error {
	rm.logger.Info("Updating rule template", "template_id", template.ID)

	// Validate template
	if err := rm.validateTemplate(template); err != nil {
		return fmt.Errorf("template validation failed: %w", err)
	}

	// Get existing template to preserve metadata
	existing, err := rm.templateStorage.GetTemplate(ctx, template.ID)
	if err != nil {
		return fmt.Errorf("failed to get existing template: %w", err)
	}

	template.CreatedAt = existing.CreatedAt
	template.CreatedBy = existing.CreatedBy
	template.UpdatedAt = time.Now()

	return rm.templateStorage.UpdateTemplate(ctx, template)
}

// DeleteTemplate deletes a template.
func (rm *RuleManager) DeleteTemplate(ctx context.Context, templateID string) error {
	rm.logger.Info("Deleting rule template", "template_id", templateID)
	return rm.templateStorage.DeleteTemplate(ctx, templateID)
}

// InstantiateTemplate creates a rule from a template with variable substitution.
func (rm *RuleManager) InstantiateTemplate(ctx context.Context, templateID string, variables map[string]interface{}) (*AutomationRule, error) {
	template, err := rm.GetTemplate(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	// Validate required variables
	for _, variable := range template.Variables {
		if variable.Required {
			if _, exists := variables[variable.Name]; !exists {
				return nil, fmt.Errorf("required variable '%s' not provided", variable.Name)
			}
		}
	}

	// Create rule from template
	rule := template.Template
	rule.ID = uuid.New().String()
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()

	// Apply variable substitution
	if err := rm.applyVariableSubstitution(&rule, variables, template.Variables); err != nil {
		return nil, fmt.Errorf("variable substitution failed: %w", err)
	}

	return &rule, nil
}

// Execution History

// GetExecution retrieves an execution by ID.
func (rm *RuleManager) GetExecution(ctx context.Context, executionID string) (*AutomationRuleExecution, error) {
	return rm.storage.GetExecution(ctx, executionID)
}

// ListExecutions lists executions with optional filtering.
func (rm *RuleManager) ListExecutions(ctx context.Context, org string, filter *ExecutionFilter) ([]*AutomationRuleExecution, error) {
	return rm.storage.ListExecutions(ctx, org, filter)
}

// CancelExecution cancels a running execution.
func (rm *RuleManager) CancelExecution(ctx context.Context, executionID string) error {
	execution, err := rm.storage.GetExecution(ctx, executionID)
	if err != nil {
		return err
	}

	if execution.Status != ExecutionStatusRunning && execution.Status != ExecutionStatusPending {
		return fmt.Errorf("cannot cancel execution in status: %s", execution.Status)
	}

	execution.Status = ExecutionStatusCancelled
	completedAt := time.Now()
	execution.CompletedAt = &completedAt
	execution.Duration = completedAt.Sub(execution.StartedAt)

	return rm.storage.SaveExecution(ctx, execution)
}

// Validation and Testing

// ValidateRule validates a rule structure and configuration.
func (rm *RuleManager) ValidateRule(ctx context.Context, rule *AutomationRule) (*RuleValidationResult, error) {
	result := &RuleValidationResult{
		Valid:    true,
		Errors:   []RuleValidationError{},
		Warnings: []RuleValidationWarning{},
		Score:    100,
	}

	// Basic validation
	if err := rm.validateRule(rule); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, RuleValidationError{
			Field:    "rule",
			Message:  err.Error(),
			Severity: "error",
		})
		result.Score -= 20
	}

	// Validate conditions
	conditionResult, err := rm.evaluator.ValidateConditions(&rule.Conditions)
	if err != nil {
		result.Errors = append(result.Errors, RuleValidationError{
			Field:    "conditions",
			Message:  fmt.Sprintf("Condition validation failed: %v", err),
			Severity: "error",
		})
		result.Score -= 30
	} else if !conditionResult.Valid {
		result.Valid = false
		for _, condErr := range conditionResult.Errors {
			result.Errors = append(result.Errors, RuleValidationError{
				Field:      fmt.Sprintf("conditions.%s", condErr.Field),
				Message:    condErr.Message,
				Severity:   "error",
				Suggestion: condErr.Suggestion,
			})
		}

		result.Score -= 25
	}

	// Validate actions
	for i, action := range rule.Actions {
		if err := rm.actionExecutor.ValidateAction(ctx, &action); err != nil {
			result.Errors = append(result.Errors, RuleValidationError{
				Field:    fmt.Sprintf("actions[%d]", i),
				Message:  err.Error(),
				Severity: "error",
			})
			result.Score -= 15
		}
	}

	// Performance warnings
	if len(rule.Actions) > 10 {
		result.Warnings = append(result.Warnings, RuleValidationWarning{
			Field:      "actions",
			Message:    "Large number of actions may impact performance",
			Suggestion: "Consider consolidating actions or splitting into multiple rules",
		})
		result.Score -= 5
	}

	return result, nil
}

// TestRule tests a rule against a sample event.
func (rm *RuleManager) TestRule(ctx context.Context, rule *AutomationRule, testEvent *GitHubEvent) (*RuleTestResult, error) {
	startTime := time.Now()

	// Create test execution context
	execContext := &AutomationExecutionContext{
		Event:        testEvent,
		Organization: testEvent.Organization,
		User:         testEvent.Sender,
		Variables:    make(map[string]interface{}),
		Environment:  "test",
	}

	result := &RuleTestResult{
		RuleID:          rule.ID,
		ActionsExecuted: []ActionExecutionResult{},
		Context:         *execContext,
		Errors:          []string{},
	}

	// Test condition matching
	matched, err := rm.EvaluateConditions(ctx, rule, testEvent)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Condition evaluation failed: %v", err))
		result.ConditionsMatched = false
	} else {
		result.ConditionsMatched = matched
	}

	// If conditions match, simulate action execution
	if matched {
		for _, action := range rule.Actions {
			if !action.Enabled {
				continue
			}

			actionResult := ActionExecutionResult{
				ActionID:   action.ID,
				ActionType: action.Type,
				Status:     ExecutionStatusCompleted,
				StartedAt:  time.Now(),
				Duration:   50 * time.Millisecond, // Simulated duration
				Result: map[string]interface{}{
					"simulated": true,
					"test_mode": true,
				},
			}

			result.ActionsExecuted = append(result.ActionsExecuted, actionResult)
		}
	}

	result.ExecutionTime = time.Since(startTime)

	return result, nil
}

// DryRunRule performs a dry run of a rule against an event without executing actions.
func (rm *RuleManager) DryRunRule(ctx context.Context, ruleID string, event *GitHubEvent) (*RuleTestResult, error) {
	rule, err := rm.GetRule(ctx, event.Organization, ruleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get rule: %w", err)
	}

	return rm.TestRule(ctx, rule, event)
}

// Helper methods

func (rm *RuleManager) validateRule(rule *AutomationRule) error {
	if rule.Name == "" {
		return fmt.Errorf("rule name is required")
	}

	if rule.Organization == "" {
		return fmt.Errorf("organization is required")
	}

	if len(rule.Actions) == 0 {
		return fmt.Errorf("at least one action is required")
	}

	return nil
}

func (rm *RuleManager) validateRuleSet(ruleSet *AutomationRuleSet) error {
	if ruleSet.Name == "" {
		return fmt.Errorf("rule set name is required")
	}

	if ruleSet.Organization == "" {
		return fmt.Errorf("organization is required")
	}

	if len(ruleSet.Rules) == 0 {
		return fmt.Errorf("at least one rule is required")
	}

	return nil
}

func (rm *RuleManager) validateTemplate(template *AutomationRuleTemplate) error {
	if template.Name == "" {
		return fmt.Errorf("template name is required")
	}

	if template.Category == "" {
		return fmt.Errorf("template category is required")
	}

	return rm.validateRule(&template.Template)
}

func (rm *RuleManager) cacheKey(org, ruleID string) string {
	return fmt.Sprintf("%s:%s", org, ruleID)
}

func (rm *RuleManager) getRepositoryInfo(ctx context.Context, org, repo string) (*RepositoryInfo, error) {
	// This would typically call the GitHub API to get repository information
	// For now, return a basic structure
	return &RepositoryInfo{
		Name:       repo,
		Language:   "Unknown",
		Topics:     []string{},
		Visibility: "unknown",
		Archived:   false,
		IsTemplate: false,
	}, nil
}

func (rm *RuleManager) getOrganizationInfo(ctx context.Context, org string) (*OrganizationInfo, error) {
	// This would typically call the GitHub API to get organization information
	// For now, return a basic structure
	return &OrganizationInfo{
		Login:             org,
		Name:              org,
		Type:              "Organization",
		Plan:              "unknown",
		TwoFactorRequired: false,
		MemberCount:       0,
		RepoCount:         0,
	}, nil
}

func (rm *RuleManager) applyVariableSubstitution(rule *AutomationRule, variables map[string]interface{}, templateVars []TemplateVariable) error {
	// Create variable map with defaults
	varMap := make(map[string]interface{})

	for _, tv := range templateVars {
		if tv.DefaultValue != nil {
			varMap[tv.Name] = tv.DefaultValue
		}
	}

	// Override with provided variables
	for k, v := range variables {
		varMap[k] = v
	}

	// Convert rule to JSON for string replacement
	ruleJSON, err := json.Marshal(rule)
	if err != nil {
		return err
	}

	ruleStr := string(ruleJSON)

	// Replace template variables
	for k, v := range varMap {
		placeholder := fmt.Sprintf("{{%s}}", k)
		replacement := fmt.Sprintf("%v", v)
		ruleStr = strings.ReplaceAll(ruleStr, placeholder, replacement)
	}

	// Convert back to rule
	return json.Unmarshal([]byte(ruleStr), rule)
}

// shouldRetryAction determines if an action should be retried based on its retry policy.
func (rm *RuleManager) shouldRetryAction(action *AutomationAction, result *ActionExecutionResult) bool {
	if action.RetryPolicy == nil {
		return false
	}

	return result.RetryCount < action.RetryPolicy.MaxRetries
}

// retryActionWithBackoff retries an action with exponential backoff.
func (rm *RuleManager) retryActionWithBackoff(ctx context.Context, action *AutomationAction, result *ActionExecutionResult, ruleID string) error {
	if action.RetryPolicy == nil {
		return fmt.Errorf("no retry policy configured for action %s", action.ID)
	}

	policy := action.RetryPolicy
	interval := policy.RetryInterval
	if interval == 0 {
		interval = 1 * time.Second // Default retry interval
	}

	backoffFactor := policy.BackoffFactor
	if backoffFactor <= 0 {
		backoffFactor = 2.0 // Default exponential backoff factor
	}

	maxInterval := policy.MaxInterval
	if maxInterval == 0 {
		maxInterval = 60 * time.Second // Default max interval
	}

	var lastErr error
	for attempt := result.RetryCount; attempt < policy.MaxRetries; attempt++ {
		// Calculate retry delay with exponential backoff
		retryDelay := time.Duration(float64(interval) * pow(backoffFactor, float64(attempt)))
		if retryDelay > maxInterval {
			retryDelay = maxInterval
		}

		rm.logger.Info("Retrying action",
			"action_id", action.ID,
			"rule_id", ruleID,
			"attempt", attempt+1,
			"max_retries", policy.MaxRetries,
			"delay", retryDelay)

		// Wait before retry
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(retryDelay):
		}

		// Increment retry count
		result.RetryCount = attempt + 1

		// Execute the action again
		_, executionErr := rm.actionExecutor.ExecuteAction(ctx, action, nil)
		if executionErr == nil {
			rm.logger.Info("Action retry succeeded",
				"action_id", action.ID,
				"rule_id", ruleID,
				"attempt", attempt+1)
			return nil
		}

		lastErr = executionErr
		rm.logger.Warn("Action retry failed",
			"action_id", action.ID,
			"rule_id", ruleID,
			"attempt", attempt+1,
			"error", executionErr)
	}

	return fmt.Errorf("action failed after %d retry attempts: %w", policy.MaxRetries, lastErr)
}

// pow is a simple integer power function for exponential backoff calculation.
func pow(base float64, exp float64) float64 {
	result := 1.0
	for i := 0; i < int(exp); i++ {
		result *= base
	}
	return result
}
