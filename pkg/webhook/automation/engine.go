package automation

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/go-github/v66/github"
	"go.uber.org/zap"
)

// Engine represents the event-based automation rules engine
type Engine struct {
	rules        []Rule
	handlers     map[string]ActionHandler
	logger       *zap.Logger
	githubClient *github.Client
	mu           sync.RWMutex
	metrics      *Metrics
}

// Rule represents an automation rule
type Rule struct {
	ID          string                 `json:"id" yaml:"id"`
	Name        string                 `json:"name" yaml:"name"`
	Description string                 `json:"description" yaml:"description"`
	Enabled     bool                   `json:"enabled" yaml:"enabled"`
	Priority    int                    `json:"priority" yaml:"priority"`
	Conditions  []Condition            `json:"conditions" yaml:"conditions"`
	Actions     []Action               `json:"actions" yaml:"actions"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// Condition represents a rule condition
type Condition struct {
	Type       string                 `json:"type" yaml:"type"`
	Field      string                 `json:"field" yaml:"field"`
	Operator   string                 `json:"operator" yaml:"operator"`
	Value      interface{}            `json:"value" yaml:"value"`
	Parameters map[string]interface{} `json:"parameters,omitempty" yaml:"parameters,omitempty"`
}

// Action represents an action to be executed
type Action struct {
	Type       string                 `json:"type" yaml:"type"`
	Parameters map[string]interface{} `json:"parameters" yaml:"parameters"`
	Async      bool                   `json:"async,omitempty" yaml:"async,omitempty"`
	Timeout    string                 `json:"timeout,omitempty" yaml:"timeout,omitempty"`
}

// Event represents a webhook event
type Event struct {
	ID         string             `json:"id"`
	Type       string             `json:"type"`
	Action     string             `json:"action,omitempty"`
	Repository *github.Repository `json:"repository,omitempty"`
	Sender     *github.User       `json:"sender,omitempty"`
	Payload    interface{}        `json:"payload"`
	ReceivedAt time.Time          `json:"received_at"`
	Headers    map[string]string  `json:"headers,omitempty"`
	RawPayload json.RawMessage    `json:"raw_payload,omitempty"`
}

// ActionHandler defines the interface for action handlers
type ActionHandler interface {
	Execute(ctx context.Context, event *Event, action Action) error
	ValidateParameters(params map[string]interface{}) error
}

// Metrics tracks engine performance
type Metrics struct {
	mu              sync.RWMutex
	EventsProcessed int64
	RulesEvaluated  int64
	ActionsExecuted int64
	Errors          int64
	ProcessingTime  time.Duration
}

// NewEngine creates a new automation engine
func NewEngine(githubClient *github.Client, logger *zap.Logger) *Engine {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &Engine{
		rules:        []Rule{},
		handlers:     make(map[string]ActionHandler),
		logger:       logger,
		githubClient: githubClient,
		metrics:      &Metrics{},
	}
}

// RegisterHandler registers an action handler
func (e *Engine) RegisterHandler(actionType string, handler ActionHandler) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, exists := e.handlers[actionType]; exists {
		return fmt.Errorf("handler for action type %s already registered", actionType)
	}

	e.handlers[actionType] = handler
	e.logger.Info("Registered action handler", zap.String("type", actionType))
	return nil
}

// AddRule adds a new rule to the engine
func (e *Engine) AddRule(rule Rule) error {
	if err := e.validateRule(rule); err != nil {
		return fmt.Errorf("invalid rule: %w", err)
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	e.rules = append(e.rules, rule)
	e.sortRules()

	e.logger.Info("Added rule",
		zap.String("id", rule.ID),
		zap.String("name", rule.Name),
		zap.Int("priority", rule.Priority))

	return nil
}

// ProcessEvent processes a webhook event through the rules engine
func (e *Engine) ProcessEvent(ctx context.Context, event *Event) error {
	start := time.Now()
	defer func() {
		e.metrics.mu.Lock()
		e.metrics.EventsProcessed++
		e.metrics.ProcessingTime += time.Since(start)
		e.metrics.mu.Unlock()
	}()

	e.logger.Info("Processing event",
		zap.String("id", event.ID),
		zap.String("type", event.Type),
		zap.String("action", event.Action))

	e.mu.RLock()
	rules := make([]Rule, len(e.rules))
	copy(rules, e.rules)
	e.mu.RUnlock()

	var matchedRules []Rule
	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		e.metrics.mu.Lock()
		e.metrics.RulesEvaluated++
		e.metrics.mu.Unlock()

		if e.evaluateRule(ctx, rule, event) {
			matchedRules = append(matchedRules, rule)
			e.logger.Debug("Rule matched",
				zap.String("rule_id", rule.ID),
				zap.String("rule_name", rule.Name))
		}
	}

	// Execute actions for matched rules
	for _, rule := range matchedRules {
		if err := e.executeRuleActions(ctx, rule, event); err != nil {
			e.metrics.mu.Lock()
			e.metrics.Errors++
			e.metrics.mu.Unlock()

			e.logger.Error("Failed to execute rule actions",
				zap.String("rule_id", rule.ID),
				zap.Error(err))
			// Continue processing other rules even if one fails
		}
	}

	return nil
}

// evaluateRule evaluates if a rule matches the event
func (e *Engine) evaluateRule(ctx context.Context, rule Rule, event *Event) bool {
	// All conditions must match (AND logic)
	for _, condition := range rule.Conditions {
		if !e.evaluateCondition(ctx, condition, event) {
			return false
		}
	}
	return true
}

// evaluateCondition evaluates a single condition
func (e *Engine) evaluateCondition(ctx context.Context, condition Condition, event *Event) bool {
	switch condition.Type {
	case "event_type":
		return e.evaluateEventTypeCondition(condition, event)
	case "repository":
		return e.evaluateRepositoryCondition(condition, event)
	case "sender":
		return e.evaluateSenderCondition(condition, event)
	case "payload":
		return e.evaluatePayloadCondition(condition, event)
	case "time":
		return e.evaluateTimeCondition(condition, event)
	default:
		e.logger.Warn("Unknown condition type", zap.String("type", condition.Type))
		return false
	}
}

// evaluateEventTypeCondition checks event type conditions
func (e *Engine) evaluateEventTypeCondition(condition Condition, event *Event) bool {
	eventType := event.Type
	if event.Action != "" {
		eventType = fmt.Sprintf("%s.%s", event.Type, event.Action)
	}

	switch condition.Operator {
	case "equals", "==":
		return eventType == condition.Value.(string)
	case "not_equals", "!=":
		return eventType != condition.Value.(string)
	case "in":
		values, ok := condition.Value.([]interface{})
		if !ok {
			return false
		}
		for _, v := range values {
			if str, ok := v.(string); ok && eventType == str {
				return true
			}
		}
		return false
	case "matches":
		pattern, ok := condition.Value.(string)
		if !ok {
			return false
		}
		matched, _ := regexp.MatchString(pattern, eventType)
		return matched
	default:
		return false
	}
}

// evaluateRepositoryCondition checks repository conditions
func (e *Engine) evaluateRepositoryCondition(condition Condition, event *Event) bool {
	if event.Repository == nil {
		return false
	}

	var value string
	switch condition.Field {
	case "name":
		value = event.Repository.GetName()
	case "full_name":
		value = event.Repository.GetFullName()
	case "private":
		return e.evaluateBoolCondition(condition, event.Repository.GetPrivate())
	case "language":
		value = event.Repository.GetLanguage()
	case "default_branch":
		value = event.Repository.GetDefaultBranch()
	default:
		return false
	}

	return e.evaluateStringCondition(condition, value)
}

// evaluateSenderCondition checks sender conditions
func (e *Engine) evaluateSenderCondition(condition Condition, event *Event) bool {
	if event.Sender == nil {
		return false
	}

	var value string
	switch condition.Field {
	case "login":
		value = event.Sender.GetLogin()
	case "type":
		value = event.Sender.GetType()
	case "site_admin":
		return e.evaluateBoolCondition(condition, event.Sender.GetSiteAdmin())
	default:
		return false
	}

	return e.evaluateStringCondition(condition, value)
}

// evaluatePayloadCondition checks payload field conditions
func (e *Engine) evaluatePayloadCondition(condition Condition, event *Event) bool {
	// This would need to be implemented based on specific payload structures
	// For now, return false
	return false
}

// evaluateTimeCondition checks time-based conditions
func (e *Engine) evaluateTimeCondition(condition Condition, event *Event) bool {
	// Implement time-based conditions (business hours, days of week, etc.)
	return true
}

// evaluateStringCondition is a helper for string comparisons
func (e *Engine) evaluateStringCondition(condition Condition, value string) bool {
	switch condition.Operator {
	case "equals", "==":
		return value == condition.Value.(string)
	case "not_equals", "!=":
		return value != condition.Value.(string)
	case "contains":
		return containsString(value, condition.Value.(string))
	case "starts_with":
		return startsWithString(value, condition.Value.(string))
	case "ends_with":
		return endsWithString(value, condition.Value.(string))
	case "matches":
		pattern, ok := condition.Value.(string)
		if !ok {
			return false
		}
		matched, _ := regexp.MatchString(pattern, value)
		return matched
	default:
		return false
	}
}

// evaluateBoolCondition is a helper for boolean comparisons
func (e *Engine) evaluateBoolCondition(condition Condition, value bool) bool {
	expected, ok := condition.Value.(bool)
	if !ok {
		return false
	}

	switch condition.Operator {
	case "equals", "==":
		return value == expected
	case "not_equals", "!=":
		return value != expected
	default:
		return false
	}
}

// executeRuleActions executes all actions for a matched rule
func (e *Engine) executeRuleActions(ctx context.Context, rule Rule, event *Event) error {
	for _, action := range rule.Actions {
		e.metrics.mu.Lock()
		e.metrics.ActionsExecuted++
		e.metrics.mu.Unlock()

		if action.Async {
			// Execute asynchronously
			go func(act Action) {
				actCtx := context.Background()
				if action.Timeout != "" {
					duration, err := time.ParseDuration(action.Timeout)
					if err == nil {
						var cancel context.CancelFunc
						actCtx, cancel = context.WithTimeout(actCtx, duration)
						defer cancel()
					}
				}

				if err := e.executeAction(actCtx, act, event); err != nil {
					e.logger.Error("Failed to execute async action",
						zap.String("rule_id", rule.ID),
						zap.String("action_type", act.Type),
						zap.Error(err))
				}
			}(action)
		} else {
			// Execute synchronously
			actionCtx := ctx
			if action.Timeout != "" {
				duration, err := time.ParseDuration(action.Timeout)
				if err == nil {
					var cancel context.CancelFunc
					actionCtx, cancel = context.WithTimeout(ctx, duration)
					defer cancel()
				}
			}

			if err := e.executeAction(actionCtx, action, event); err != nil {
				return fmt.Errorf("failed to execute action %s: %w", action.Type, err)
			}
		}
	}

	return nil
}

// executeAction executes a single action
func (e *Engine) executeAction(ctx context.Context, action Action, event *Event) error {
	e.mu.RLock()
	handler, exists := e.handlers[action.Type]
	e.mu.RUnlock()

	if !exists {
		return fmt.Errorf("no handler registered for action type: %s", action.Type)
	}

	e.logger.Info("Executing action",
		zap.String("type", action.Type),
		zap.String("event_id", event.ID))

	return handler.Execute(ctx, event, action)
}

// validateRule validates a rule configuration
func (e *Engine) validateRule(rule Rule) error {
	if rule.ID == "" {
		return fmt.Errorf("rule ID is required")
	}
	if rule.Name == "" {
		return fmt.Errorf("rule name is required")
	}
	if len(rule.Conditions) == 0 {
		return fmt.Errorf("at least one condition is required")
	}
	if len(rule.Actions) == 0 {
		return fmt.Errorf("at least one action is required")
	}

	// Validate conditions
	for i, condition := range rule.Conditions {
		if condition.Type == "" {
			return fmt.Errorf("condition[%d] type is required", i)
		}
		if condition.Operator == "" {
			return fmt.Errorf("condition[%d] operator is required", i)
		}
	}

	// Validate actions
	for i, action := range rule.Actions {
		if action.Type == "" {
			return fmt.Errorf("action[%d] type is required", i)
		}

		// Validate action parameters if handler is registered
		e.mu.RLock()
		handler, exists := e.handlers[action.Type]
		e.mu.RUnlock()

		if exists {
			if err := handler.ValidateParameters(action.Parameters); err != nil {
				return fmt.Errorf("action[%d] parameter validation failed: %w", i, err)
			}
		}
	}

	return nil
}

// sortRules sorts rules by priority (higher priority first)
func (e *Engine) sortRules() {
	// Implement sorting logic based on priority
	// Higher priority rules are evaluated first
}

// GetMetrics returns the current metrics
func (e *Engine) GetMetrics() Metrics {
	e.metrics.mu.RLock()
	defer e.metrics.mu.RUnlock()
	return *e.metrics
}

// Helper functions
func containsString(s, substr string) bool {
	return strings.Contains(s, substr)
}

func startsWithString(s, prefix string) bool {
	return len(s) >= len(prefix) && s[0:len(prefix)] == prefix
}

func endsWithString(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}
