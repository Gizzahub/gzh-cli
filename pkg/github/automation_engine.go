package github

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// AutomationEngine is the main engine that processes GitHub events and executes automation rules
type AutomationEngine struct {
	logger             Logger
	apiClient          APIClient
	ruleManager        *RuleManager
	conditionEvaluator ConditionEvaluator
	actionExecutor     ActionExecutor
	eventProcessor     EventProcessor

	// Configuration
	config *AutomationEngineConfig

	// State management
	mu               sync.RWMutex
	running          bool
	activeExecutions map[string]*AutomationRuleExecution
	executionWorkers int

	// Channels for processing
	eventChannel     chan *GitHubEvent
	executionChannel chan *ExecutionTask
	shutdownChannel  chan struct{}

	// Metrics
	metrics *EngineMetrics
}

// AutomationEngineConfig holds configuration for the automation engine
type AutomationEngineConfig struct {
	// Worker configuration
	MaxWorkers       int           `json:"max_workers" yaml:"max_workers"`
	EventBufferSize  int           `json:"event_buffer_size" yaml:"event_buffer_size"`
	ExecutionTimeout time.Duration `json:"execution_timeout" yaml:"execution_timeout"`

	// Rate limiting
	EventsPerSecond     int `json:"events_per_second" yaml:"events_per_second"`
	ExecutionsPerMinute int `json:"executions_per_minute" yaml:"executions_per_minute"`

	// Feature flags
	EnableAsyncExecution bool `json:"enable_async_execution" yaml:"enable_async_execution"`
	EnableRuleFiltering  bool `json:"enable_rule_filtering" yaml:"enable_rule_filtering"`
	EnableMetrics        bool `json:"enable_metrics" yaml:"enable_metrics"`

	// Error handling
	MaxRetries         int     `json:"max_retries" yaml:"max_retries"`
	RetryBackoffFactor float64 `json:"retry_backoff_factor" yaml:"retry_backoff_factor"`
	ErrorThreshold     int     `json:"error_threshold" yaml:"error_threshold"`

	// Filtering
	ExcludedEventTypes []EventType `json:"excluded_event_types" yaml:"excluded_event_types"`
	IncludedEventTypes []EventType `json:"included_event_types" yaml:"included_event_types"`
	Organizations      []string    `json:"organizations" yaml:"organizations"`
}

// ExecutionTask represents a task to execute a rule
type ExecutionTask struct {
	ID         string
	Rule       *AutomationRule
	Event      *GitHubEvent
	Context    *AutomationExecutionContext
	RetryCount int
	CreatedAt  time.Time
}

// EngineMetrics holds metrics for the automation engine
type EngineMetrics struct {
	mu                    sync.RWMutex
	EventsProcessed       int64                     `json:"events_processed"`
	RulesEvaluated        int64                     `json:"rules_evaluated"`
	RulesExecuted         int64                     `json:"rules_executed"`
	ExecutionErrors       int64                     `json:"execution_errors"`
	AverageExecutionTime  time.Duration             `json:"average_execution_time"`
	EventTypeDistribution map[string]int64          `json:"event_type_distribution"`
	ExecutionsByStatus    map[ExecutionStatus]int64 `json:"executions_by_status"`
	LastProcessedEvent    time.Time                 `json:"last_processed_event"`
	StartTime             time.Time                 `json:"start_time"`
}

// AutomationEventProcessor defines the interface for processing GitHub events in automation
type AutomationEventProcessor interface {
	ProcessEvent(ctx context.Context, event *GitHubEvent) error
	FilterEvent(event *GitHubEvent) bool
	ValidateEvent(event *GitHubEvent) error
}

// NewAutomationEngine creates a new automation engine
func NewAutomationEngine(
	logger Logger,
	apiClient APIClient,
	ruleManager *RuleManager,
	conditionEvaluator ConditionEvaluator,
	actionExecutor ActionExecutor,
	eventProcessor EventProcessor,
	config *AutomationEngineConfig,
) *AutomationEngine {
	if config == nil {
		config = getDefaultConfig()
	}

	engine := &AutomationEngine{
		logger:             logger,
		apiClient:          apiClient,
		ruleManager:        ruleManager,
		conditionEvaluator: conditionEvaluator,
		actionExecutor:     actionExecutor,
		eventProcessor:     eventProcessor,
		config:             config,
		activeExecutions:   make(map[string]*AutomationRuleExecution),
		eventChannel:       make(chan *GitHubEvent, config.EventBufferSize),
		executionChannel:   make(chan *ExecutionTask, config.MaxWorkers*2),
		shutdownChannel:    make(chan struct{}),
		metrics:            newEngineMetrics(),
	}

	return engine
}

// Start starts the automation engine
func (ae *AutomationEngine) Start(ctx context.Context) error {
	ae.mu.Lock()
	defer ae.mu.Unlock()

	if ae.running {
		return fmt.Errorf("automation engine is already running")
	}

	ae.logger.Info("Starting automation engine",
		"max_workers", ae.config.MaxWorkers,
		"event_buffer_size", ae.config.EventBufferSize)

	ae.running = true
	ae.metrics.StartTime = time.Now()

	// Start event processing workers
	for i := 0; i < ae.config.MaxWorkers; i++ {
		go ae.eventWorker(ctx, i)
	}

	// Start execution workers
	ae.executionWorkers = ae.config.MaxWorkers / 2
	if ae.executionWorkers < 1 {
		ae.executionWorkers = 1
	}

	for i := 0; i < ae.executionWorkers; i++ {
		go ae.executionWorker(ctx, i)
	}

	// Start metrics collector if enabled
	if ae.config.EnableMetrics {
		go ae.metricsCollector(ctx)
	}

	ae.logger.Info("Automation engine started successfully")
	return nil
}

// Stop stops the automation engine
func (ae *AutomationEngine) Stop(ctx context.Context) error {
	ae.mu.Lock()
	defer ae.mu.Unlock()

	if !ae.running {
		return fmt.Errorf("automation engine is not running")
	}

	ae.logger.Info("Stopping automation engine")

	close(ae.shutdownChannel)
	ae.running = false

	// Wait for active executions to complete or timeout
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			ae.logger.Warn("Timeout waiting for executions to complete",
				"active_executions", len(ae.activeExecutions))
			return nil
		case <-ticker.C:
			if len(ae.activeExecutions) == 0 {
				ae.logger.Info("All executions completed, automation engine stopped")
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// ProcessEvent processes a GitHub event through the automation engine
func (ae *AutomationEngine) ProcessEvent(ctx context.Context, event *GitHubEvent) error {
	if !ae.isRunning() {
		return fmt.Errorf("automation engine is not running")
	}

	// Validate event
	if ae.eventProcessor != nil {
		if err := ae.eventProcessor.ValidateEvent(ctx, event); err != nil {
			ae.logger.Warn("Event validation failed", "event_id", event.ID, "error", err)
			return fmt.Errorf("event validation failed: %w", err)
		}
	}

	// Filter event if filtering is enabled
	if ae.config.EnableRuleFiltering && ae.eventProcessor != nil {
		if passed, err := ae.eventProcessor.FilterEvent(ctx, event, nil); err != nil || !passed {
			ae.logger.Debug("Event filtered out", "event_id", event.ID, "event_type", event.Type)
			return nil
		}
	}

	// Check if event type is excluded
	if ae.isEventTypeExcluded(event.Type) {
		ae.logger.Debug("Event type excluded", "event_id", event.ID, "event_type", event.Type)
		return nil
	}

	// Send to event channel for processing
	select {
	case ae.eventChannel <- event:
		ae.updateMetrics(func(m *EngineMetrics) {
			m.EventsProcessed++
			m.LastProcessedEvent = time.Now()
			if m.EventTypeDistribution == nil {
				m.EventTypeDistribution = make(map[string]int64)
			}
			m.EventTypeDistribution[event.Type]++
		})
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return fmt.Errorf("event channel is full, dropping event %s", event.ID)
	}
}

// GetMetrics returns current engine metrics
func (ae *AutomationEngine) GetMetrics() *EngineMetrics {
	ae.metrics.mu.RLock()
	defer ae.metrics.mu.RUnlock()

	// Create a copy to avoid data races
	metrics := &EngineMetrics{
		EventsProcessed:       ae.metrics.EventsProcessed,
		RulesEvaluated:        ae.metrics.RulesEvaluated,
		RulesExecuted:         ae.metrics.RulesExecuted,
		ExecutionErrors:       ae.metrics.ExecutionErrors,
		AverageExecutionTime:  ae.metrics.AverageExecutionTime,
		EventTypeDistribution: make(map[string]int64),
		ExecutionsByStatus:    make(map[ExecutionStatus]int64),
		LastProcessedEvent:    ae.metrics.LastProcessedEvent,
		StartTime:             ae.metrics.StartTime,
	}

	for k, v := range ae.metrics.EventTypeDistribution {
		metrics.EventTypeDistribution[k] = v
	}

	for k, v := range ae.metrics.ExecutionsByStatus {
		metrics.ExecutionsByStatus[k] = v
	}

	return metrics
}

// GetActiveExecutions returns currently active executions
func (ae *AutomationEngine) GetActiveExecutions() map[string]*AutomationRuleExecution {
	ae.mu.RLock()
	defer ae.mu.RUnlock()

	executions := make(map[string]*AutomationRuleExecution)
	for k, v := range ae.activeExecutions {
		executions[k] = v
	}

	return executions
}

// Worker functions

func (ae *AutomationEngine) eventWorker(ctx context.Context, workerID int) {
	ae.logger.Debug("Starting event worker", "worker_id", workerID)

	for {
		select {
		case event := <-ae.eventChannel:
			ae.handleEvent(ctx, event, workerID)
		case <-ae.shutdownChannel:
			ae.logger.Debug("Event worker shutting down", "worker_id", workerID)
			return
		case <-ctx.Done():
			ae.logger.Debug("Event worker context cancelled", "worker_id", workerID)
			return
		}
	}
}

func (ae *AutomationEngine) executionWorker(ctx context.Context, workerID int) {
	ae.logger.Debug("Starting execution worker", "worker_id", workerID)

	for {
		select {
		case task := <-ae.executionChannel:
			ae.executeTask(ctx, task, workerID)
		case <-ae.shutdownChannel:
			ae.logger.Debug("Execution worker shutting down", "worker_id", workerID)
			return
		case <-ctx.Done():
			ae.logger.Debug("Execution worker context cancelled", "worker_id", workerID)
			return
		}
	}
}

func (ae *AutomationEngine) handleEvent(ctx context.Context, event *GitHubEvent, workerID int) {
	ae.logger.Debug("Processing event", "event_id", event.ID, "worker_id", workerID)

	// Get applicable rules for the organization
	filter := &RuleFilter{
		Organization: event.Organization,
		Enabled:      boolPtr(true),
	}

	rules, err := ae.ruleManager.ListRules(ctx, event.Organization, filter)
	if err != nil {
		ae.logger.Error("Failed to get rules for organization",
			"organization", event.Organization,
			"error", err)
		return
	}

	ae.logger.Debug("Found rules for evaluation",
		"event_id", event.ID,
		"organization", event.Organization,
		"rule_count", len(rules))

	// Evaluate each rule
	for _, rule := range rules {
		ae.evaluateRule(ctx, rule, event, workerID)
	}
}

func (ae *AutomationEngine) evaluateRule(ctx context.Context, rule *AutomationRule, event *GitHubEvent, workerID int) {
	ae.updateMetrics(func(m *EngineMetrics) {
		m.RulesEvaluated++
	})

	ae.logger.Debug("Evaluating rule",
		"rule_id", rule.ID,
		"event_id", event.ID,
		"worker_id", workerID)

	// Evaluate conditions
	matched, err := ae.ruleManager.EvaluateConditions(ctx, rule, event)
	if err != nil {
		ae.logger.Error("Rule evaluation failed",
			"rule_id", rule.ID,
			"event_id", event.ID,
			"error", err)
		return
	}

	if !matched {
		ae.logger.Debug("Rule conditions not matched",
			"rule_id", rule.ID,
			"event_id", event.ID)
		return
	}

	ae.logger.Info("Rule conditions matched, scheduling execution",
		"rule_id", rule.ID,
		"event_id", event.ID)

	// Create execution context
	execContext := &AutomationExecutionContext{
		Event:        event,
		Organization: event.Organization,
		User:         event.Sender,
		Variables:    make(map[string]interface{}),
		Environment:  rule.Metadata.Environment,
		Metadata:     make(map[string]interface{}),
	}

	// Add event metadata to context
	execContext.Variables["event_id"] = event.ID
	execContext.Variables["event_type"] = event.Type
	execContext.Variables["event_action"] = event.Action
	execContext.Variables["repository"] = event.Repository
	execContext.Variables["sender"] = event.Sender

	// Create execution task
	task := &ExecutionTask{
		ID:        uuid.New().String(),
		Rule:      rule,
		Event:     event,
		Context:   execContext,
		CreatedAt: time.Now(),
	}

	// Schedule execution
	if ae.config.EnableAsyncExecution {
		select {
		case ae.executionChannel <- task:
			ae.logger.Debug("Execution task scheduled", "task_id", task.ID, "rule_id", rule.ID)
		case <-ctx.Done():
			return
		default:
			ae.logger.Warn("Execution queue full, dropping task", "task_id", task.ID, "rule_id", rule.ID)
		}
	} else {
		// Execute synchronously
		ae.executeTask(ctx, task, workerID)
	}
}

func (ae *AutomationEngine) executeTask(ctx context.Context, task *ExecutionTask, workerID int) {
	startTime := time.Now()

	ae.logger.Info("Executing rule task",
		"task_id", task.ID,
		"rule_id", task.Rule.ID,
		"worker_id", workerID)

	// Track active execution
	ae.mu.Lock()
	execution := &AutomationRuleExecution{
		ID:          task.ID,
		RuleID:      task.Rule.ID,
		StartedAt:   startTime,
		Status:      ExecutionStatusRunning,
		TriggerType: ExecutionTriggerTypeEvent,
	}
	ae.activeExecutions[task.ID] = execution
	ae.mu.Unlock()

	// Create execution context with timeout
	execCtx, cancel := context.WithTimeout(ctx, ae.config.ExecutionTimeout)
	defer cancel()

	// Execute the rule
	result, err := ae.ruleManager.ExecuteRule(execCtx, task.Rule, task.Context)

	// Update metrics
	duration := time.Since(startTime)
	ae.updateMetrics(func(m *EngineMetrics) {
		m.RulesExecuted++
		if err != nil {
			m.ExecutionErrors++
		}

		// Update average execution time
		if m.AverageExecutionTime == 0 {
			m.AverageExecutionTime = duration
		} else {
			m.AverageExecutionTime = (m.AverageExecutionTime + duration) / 2
		}

		// Update execution status distribution
		if m.ExecutionsByStatus == nil {
			m.ExecutionsByStatus = make(map[ExecutionStatus]int64)
		}
		if result != nil {
			m.ExecutionsByStatus[result.Status]++
		} else {
			m.ExecutionsByStatus[ExecutionStatusFailed]++
		}
	})

	// Remove from active executions
	ae.mu.Lock()
	delete(ae.activeExecutions, task.ID)
	ae.mu.Unlock()

	if err != nil {
		ae.logger.Error("Rule execution failed",
			"task_id", task.ID,
			"rule_id", task.Rule.ID,
			"duration", duration,
			"error", err)

		// Retry logic
		if task.RetryCount < ae.config.MaxRetries {
			ae.retryTask(ctx, task)
		}
	} else {
		ae.logger.Info("Rule execution completed successfully",
			"task_id", task.ID,
			"rule_id", task.Rule.ID,
			"duration", duration,
			"execution_id", result.ID,
			"status", result.Status)
	}
}

func (ae *AutomationEngine) retryTask(ctx context.Context, task *ExecutionTask) {
	task.RetryCount++

	// Calculate backoff delay
	backoffDelay := time.Duration(float64(time.Second) *
		ae.config.RetryBackoffFactor *
		float64(task.RetryCount))

	ae.logger.Info("Retrying failed task",
		"task_id", task.ID,
		"rule_id", task.Rule.ID,
		"retry_count", task.RetryCount,
		"backoff_delay", backoffDelay)

	// Schedule retry after backoff
	go func() {
		timer := time.NewTimer(backoffDelay)
		defer timer.Stop()

		select {
		case <-timer.C:
			select {
			case ae.executionChannel <- task:
				ae.logger.Debug("Retry task scheduled", "task_id", task.ID)
			case <-ae.shutdownChannel:
				return
			case <-ctx.Done():
				return
			}
		case <-ae.shutdownChannel:
			return
		case <-ctx.Done():
			return
		}
	}()
}

func (ae *AutomationEngine) metricsCollector(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ae.collectMetrics()
		case <-ae.shutdownChannel:
			return
		case <-ctx.Done():
			return
		}
	}
}

func (ae *AutomationEngine) collectMetrics() {
	// This could be extended to collect additional metrics
	// like memory usage, goroutine count, etc.
	ae.logger.Debug("Collecting metrics",
		"events_processed", ae.metrics.EventsProcessed,
		"rules_evaluated", ae.metrics.RulesEvaluated,
		"rules_executed", ae.metrics.RulesExecuted,
		"execution_errors", ae.metrics.ExecutionErrors)
}

// Helper functions

func (ae *AutomationEngine) isRunning() bool {
	ae.mu.RLock()
	defer ae.mu.RUnlock()
	return ae.running
}

func (ae *AutomationEngine) isEventTypeExcluded(eventType string) bool {
	if len(ae.config.ExcludedEventTypes) == 0 {
		return false
	}

	for _, excluded := range ae.config.ExcludedEventTypes {
		if string(excluded) == eventType {
			return true
		}
	}

	return false
}

func (ae *AutomationEngine) updateMetrics(updateFunc func(*EngineMetrics)) {
	ae.metrics.mu.Lock()
	defer ae.metrics.mu.Unlock()
	updateFunc(ae.metrics)
}

func getDefaultConfig() *AutomationEngineConfig {
	return &AutomationEngineConfig{
		MaxWorkers:           10,
		EventBufferSize:      1000,
		ExecutionTimeout:     5 * time.Minute,
		EventsPerSecond:      100,
		ExecutionsPerMinute:  1000,
		EnableAsyncExecution: true,
		EnableRuleFiltering:  true,
		EnableMetrics:        true,
		MaxRetries:           3,
		RetryBackoffFactor:   2.0,
		ErrorThreshold:       100,
		ExcludedEventTypes:   []EventType{},
		IncludedEventTypes:   []EventType{},
		Organizations:        []string{},
	}
}

func newEngineMetrics() *EngineMetrics {
	return &EngineMetrics{
		EventTypeDistribution: make(map[string]int64),
		ExecutionsByStatus:    make(map[ExecutionStatus]int64),
	}
}

func boolPtr(b bool) *bool {
	return &b
}
