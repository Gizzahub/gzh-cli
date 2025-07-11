package monitoring

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// AdvancedAlertRule represents an enhanced alert rule with complex conditions and scheduling
type AdvancedAlertRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Priority    int                    `json:"priority"`
	Enabled     bool                   `json:"enabled"`
	Conditions  *AlertCondition        `json:"conditions"`
	Actions     []AlertAction          `json:"actions"`
	Schedule    *AlertSchedule         `json:"schedule,omitempty"`
	Throttle    *AlertThrottle         `json:"throttle,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// AlertCondition represents complex condition logic for alert evaluation
type AlertCondition struct {
	Type      string                 `json:"type"`               // "simple", "composite", "time_based"
	Operator  string                 `json:"operator,omitempty"` // "and", "or", "not"
	Metric    string                 `json:"metric,omitempty"`
	Threshold *ThresholdConfig       `json:"threshold,omitempty"`
	TimeFrame *TimeFrameConfig       `json:"time_frame,omitempty"`
	Children  []*AlertCondition      `json:"children,omitempty"`
	Custom    map[string]interface{} `json:"custom,omitempty"`
}

// ThresholdConfig represents threshold configuration for metric evaluation
type ThresholdConfig struct {
	Operator    string  `json:"operator"` // "gt", "gte", "lt", "lte", "eq", "ne", "between", "outside"
	Value       float64 `json:"value"`
	SecondValue float64 `json:"second_value,omitempty"` // For "between" and "outside" operators
	Unit        string  `json:"unit,omitempty"`
}

// TimeFrameConfig represents time-based evaluation configuration
type TimeFrameConfig struct {
	Duration    time.Duration `json:"duration"`
	Aggregation string        `json:"aggregation"` // "avg", "max", "min", "sum", "count"
	WindowType  string        `json:"window_type"` // "sliding", "tumbling"
}

// AlertAction represents an action to execute when alert fires
type AlertAction struct {
	Type     string                 `json:"type"` // "notification", "webhook", "script", "escalation"
	Target   string                 `json:"target"`
	Config   map[string]interface{} `json:"config,omitempty"`
	Delay    time.Duration          `json:"delay,omitempty"`
	Retry    *RetryConfig           `json:"retry,omitempty"`
	Template string                 `json:"template,omitempty"`
}

// RetryConfig represents retry configuration for actions
type RetryConfig struct {
	MaxAttempts int           `json:"max_attempts"`
	Interval    time.Duration `json:"interval"`
	Backoff     string        `json:"backoff"` // "linear", "exponential"
}

// AlertSchedule represents scheduling configuration for alert rules
type AlertSchedule struct {
	CronExpression     string              `json:"cron_expression,omitempty"`
	TimeZone           string              `json:"timezone"`
	ActivePeriods      []TimePeriod        `json:"active_periods,omitempty"`
	ExcludePeriods     []TimePeriod        `json:"exclude_periods,omitempty"`
	MaintenanceWindows []MaintenanceWindow `json:"maintenance_windows,omitempty"`
}

// TimePeriod represents a time period for scheduling
type TimePeriod struct {
	Start    string   `json:"start"` // "HH:MM" format
	End      string   `json:"end"`   // "HH:MM" format
	Days     []string `json:"days"`  // ["monday", "tuesday", ...]
	TimeZone string   `json:"timezone"`
}

// MaintenanceWindow represents a maintenance window where alerts are suppressed
type MaintenanceWindow struct {
	Name      string    `json:"name"`
	Start     time.Time `json:"start"`
	End       time.Time `json:"end"`
	Recurring bool      `json:"recurring"`
	Pattern   string    `json:"pattern,omitempty"` // Cron expression for recurring windows
}

// AlertThrottle represents throttling configuration to prevent alert spam
type AlertThrottle struct {
	MaxAlerts      int           `json:"max_alerts"`
	TimeWindow     time.Duration `json:"time_window"`
	CooldownPeriod time.Duration `json:"cooldown_period"`
	GroupBy        []string      `json:"group_by,omitempty"` // Fields to group throttling by
}

// AdvancedAlertManager manages enhanced alert rules with complex conditions
type AdvancedAlertManager struct {
	rules          map[string]*AdvancedAlertRule
	evaluator      *ConditionEvaluator
	scheduler      *AlertScheduler
	throttler      *AlertThrottler
	actionExecutor *AlertActionExecutor
	logger         *zap.Logger
	mutex          sync.RWMutex
}

// ConditionEvaluator handles complex condition evaluation
type ConditionEvaluator struct {
	metricProvider MetricProvider
	logger         *zap.Logger
}

// MetricProvider interface for retrieving metrics
type MetricProvider interface {
	GetMetric(ctx context.Context, metric string, timeframe *TimeFrameConfig) (float64, error)
	GetMetricHistory(ctx context.Context, metric string, duration time.Duration) ([]MetricPoint, error)
}

// MetricPoint represents a metric data point
type MetricPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// AlertScheduler handles rule scheduling and timing
type AlertScheduler struct {
	logger *zap.Logger
}

// AlertThrottler handles alert throttling and rate limiting
type AlertThrottler struct {
	alertCounts map[string][]time.Time
	mutex       sync.RWMutex
	logger      *zap.Logger
}

// AlertActionExecutor handles execution of alert actions
type AlertActionExecutor struct {
	logger *zap.Logger
}

// NewAdvancedAlertManager creates a new advanced alert manager
func NewAdvancedAlertManager(metricProvider MetricProvider, logger *zap.Logger) *AdvancedAlertManager {
	return &AdvancedAlertManager{
		rules: make(map[string]*AdvancedAlertRule),
		evaluator: &ConditionEvaluator{
			metricProvider: metricProvider,
			logger:         logger,
		},
		scheduler: &AlertScheduler{
			logger: logger,
		},
		throttler: &AlertThrottler{
			alertCounts: make(map[string][]time.Time),
			logger:      logger,
		},
		actionExecutor: &AlertActionExecutor{
			logger: logger,
		},
		logger: logger,
	}
}

// AddRule adds a new advanced alert rule
func (aam *AdvancedAlertManager) AddRule(rule *AdvancedAlertRule) error {
	aam.mutex.Lock()
	defer aam.mutex.Unlock()

	// Validate rule
	if err := aam.validateRule(rule); err != nil {
		return fmt.Errorf("invalid rule: %w", err)
	}

	// Set timestamps
	now := time.Now()
	if rule.CreatedAt.IsZero() {
		rule.CreatedAt = now
	}
	rule.UpdatedAt = now

	// Store rule
	aam.rules[rule.ID] = rule

	aam.logger.Info("Advanced alert rule added",
		zap.String("id", rule.ID),
		zap.String("name", rule.Name),
		zap.Int("priority", rule.Priority))

	return nil
}

// EvaluateRule evaluates a single rule's conditions
func (aam *AdvancedAlertManager) EvaluateRule(ctx context.Context, rule *AdvancedAlertRule) (bool, error) {
	// Check if rule is enabled
	if !rule.Enabled {
		return false, nil
	}

	// Check schedule
	if rule.Schedule != nil {
		active, err := aam.scheduler.IsRuleActive(rule.Schedule)
		if err != nil {
			return false, fmt.Errorf("schedule check failed: %w", err)
		}
		if !active {
			return false, nil
		}
	}

	// Evaluate conditions
	if rule.Conditions == nil {
		return false, fmt.Errorf("rule has no conditions")
	}

	result, err := aam.evaluator.EvaluateCondition(ctx, rule.Conditions)
	if err != nil {
		return false, fmt.Errorf("condition evaluation failed: %w", err)
	}

	// Check throttling
	if result && rule.Throttle != nil {
		throttled, err := aam.throttler.ShouldThrottle(rule.ID, rule.Throttle)
		if err != nil {
			aam.logger.Error("Throttle check failed", zap.Error(err))
		}
		if throttled {
			aam.logger.Debug("Alert throttled", zap.String("rule_id", rule.ID))
			return false, nil
		}
	}

	return result, nil
}

// ExecuteActions executes the actions for a triggered rule
func (aam *AdvancedAlertManager) ExecuteActions(ctx context.Context, rule *AdvancedAlertRule, alertData map[string]interface{}) error {
	for _, action := range rule.Actions {
		if err := aam.actionExecutor.ExecuteAction(ctx, &action, alertData); err != nil {
			aam.logger.Error("Action execution failed",
				zap.String("rule_id", rule.ID),
				zap.String("action_type", action.Type),
				zap.Error(err))
			// Continue with other actions even if one fails
		}
	}
	return nil
}

// validateRule validates an advanced alert rule
func (aam *AdvancedAlertManager) validateRule(rule *AdvancedAlertRule) error {
	if rule.ID == "" {
		return fmt.Errorf("rule ID is required")
	}
	if rule.Name == "" {
		return fmt.Errorf("rule name is required")
	}
	if rule.Conditions == nil {
		return fmt.Errorf("rule conditions are required")
	}
	if len(rule.Actions) == 0 {
		return fmt.Errorf("rule must have at least one action")
	}

	// Validate conditions
	if err := aam.validateCondition(rule.Conditions); err != nil {
		return fmt.Errorf("invalid conditions: %w", err)
	}

	// Validate actions
	for i, action := range rule.Actions {
		if err := aam.validateAction(&action); err != nil {
			return fmt.Errorf("invalid action %d: %w", i, err)
		}
	}

	return nil
}

// validateCondition validates alert conditions recursively
func (aam *AdvancedAlertManager) validateCondition(condition *AlertCondition) error {
	switch condition.Type {
	case "simple":
		if condition.Metric == "" {
			return fmt.Errorf("metric is required for simple condition")
		}
		if condition.Threshold == nil {
			return fmt.Errorf("threshold is required for simple condition")
		}
	case "composite":
		if condition.Operator == "" {
			return fmt.Errorf("operator is required for composite condition")
		}
		if len(condition.Children) == 0 {
			return fmt.Errorf("children are required for composite condition")
		}
		for _, child := range condition.Children {
			if err := aam.validateCondition(child); err != nil {
				return err
			}
		}
	case "time_based":
		if condition.TimeFrame == nil {
			return fmt.Errorf("time frame is required for time-based condition")
		}
	default:
		return fmt.Errorf("unknown condition type: %s", condition.Type)
	}
	return nil
}

// validateAction validates alert actions
func (aam *AdvancedAlertManager) validateAction(action *AlertAction) error {
	if action.Type == "" {
		return fmt.Errorf("action type is required")
	}
	if action.Target == "" {
		return fmt.Errorf("action target is required")
	}
	return nil
}

// EvaluateCondition evaluates a condition tree
func (ce *ConditionEvaluator) EvaluateCondition(ctx context.Context, condition *AlertCondition) (bool, error) {
	switch condition.Type {
	case "simple":
		return ce.evaluateSimpleCondition(ctx, condition)
	case "composite":
		return ce.evaluateCompositeCondition(ctx, condition)
	case "time_based":
		return ce.evaluateTimeBasedCondition(ctx, condition)
	default:
		return false, fmt.Errorf("unknown condition type: %s", condition.Type)
	}
}

// evaluateSimpleCondition evaluates a simple metric threshold condition
func (ce *ConditionEvaluator) evaluateSimpleCondition(ctx context.Context, condition *AlertCondition) (bool, error) {
	if condition.Metric == "" || condition.Threshold == nil {
		return false, fmt.Errorf("metric and threshold required for simple condition")
	}

	value, err := ce.metricProvider.GetMetric(ctx, condition.Metric, condition.TimeFrame)
	if err != nil {
		return false, fmt.Errorf("failed to get metric %s: %w", condition.Metric, err)
	}

	return ce.evaluateThreshold(value, condition.Threshold), nil
}

// evaluateCompositeCondition evaluates composite conditions with logical operators
func (ce *ConditionEvaluator) evaluateCompositeCondition(ctx context.Context, condition *AlertCondition) (bool, error) {
	if len(condition.Children) == 0 {
		return false, fmt.Errorf("composite condition requires children")
	}

	results := make([]bool, len(condition.Children))
	for i, child := range condition.Children {
		result, err := ce.EvaluateCondition(ctx, child)
		if err != nil {
			return false, err
		}
		results[i] = result
	}

	switch strings.ToLower(condition.Operator) {
	case "and":
		for _, result := range results {
			if !result {
				return false, nil
			}
		}
		return true, nil
	case "or":
		for _, result := range results {
			if result {
				return true, nil
			}
		}
		return false, nil
	case "not":
		if len(results) != 1 {
			return false, fmt.Errorf("NOT operator requires exactly one child")
		}
		return !results[0], nil
	default:
		return false, fmt.Errorf("unknown operator: %s", condition.Operator)
	}
}

// evaluateTimeBasedCondition evaluates time-based conditions with historical data
func (ce *ConditionEvaluator) evaluateTimeBasedCondition(ctx context.Context, condition *AlertCondition) (bool, error) {
	if condition.TimeFrame == nil {
		return false, fmt.Errorf("time frame required for time-based condition")
	}

	history, err := ce.metricProvider.GetMetricHistory(ctx, condition.Metric, condition.TimeFrame.Duration)
	if err != nil {
		return false, fmt.Errorf("failed to get metric history: %w", err)
	}

	aggregatedValue := ce.aggregateMetrics(history, condition.TimeFrame.Aggregation)
	return ce.evaluateThreshold(aggregatedValue, condition.Threshold), nil
}

// evaluateThreshold evaluates a value against threshold configuration
func (ce *ConditionEvaluator) evaluateThreshold(value float64, threshold *ThresholdConfig) bool {
	switch strings.ToLower(threshold.Operator) {
	case "gt", ">":
		return value > threshold.Value
	case "gte", ">=":
		return value >= threshold.Value
	case "lt", "<":
		return value < threshold.Value
	case "lte", "<=":
		return value <= threshold.Value
	case "eq", "==":
		return value == threshold.Value
	case "ne", "!=":
		return value != threshold.Value
	case "between":
		return value >= threshold.Value && value <= threshold.SecondValue
	case "outside":
		return value < threshold.Value || value > threshold.SecondValue
	default:
		return false
	}
}

// aggregateMetrics aggregates metric history based on aggregation type
func (ce *ConditionEvaluator) aggregateMetrics(history []MetricPoint, aggregation string) float64 {
	if len(history) == 0 {
		return 0
	}

	switch strings.ToLower(aggregation) {
	case "avg", "average":
		sum := 0.0
		for _, point := range history {
			sum += point.Value
		}
		return sum / float64(len(history))
	case "max", "maximum":
		max := history[0].Value
		for _, point := range history {
			if point.Value > max {
				max = point.Value
			}
		}
		return max
	case "min", "minimum":
		min := history[0].Value
		for _, point := range history {
			if point.Value < min {
				min = point.Value
			}
		}
		return min
	case "sum":
		sum := 0.0
		for _, point := range history {
			sum += point.Value
		}
		return sum
	case "count":
		return float64(len(history))
	default:
		return history[len(history)-1].Value // Latest value
	}
}

// IsRuleActive checks if a rule should be active based on its schedule
func (as *AlertScheduler) IsRuleActive(schedule *AlertSchedule) (bool, error) {
	now := time.Now()

	// Check maintenance windows
	for _, window := range schedule.MaintenanceWindows {
		if now.After(window.Start) && now.Before(window.End) {
			return false, nil
		}
	}

	// Check active periods
	if len(schedule.ActivePeriods) > 0 {
		active := false
		for _, period := range schedule.ActivePeriods {
			if as.isTimeInPeriod(now, period) {
				active = true
				break
			}
		}
		if !active {
			return false, nil
		}
	}

	// Check exclude periods
	for _, period := range schedule.ExcludePeriods {
		if as.isTimeInPeriod(now, period) {
			return false, nil
		}
	}

	return true, nil
}

// isTimeInPeriod checks if the current time falls within a time period
func (as *AlertScheduler) isTimeInPeriod(now time.Time, period TimePeriod) bool {
	// Simple implementation - would need proper timezone handling in production
	weekday := strings.ToLower(now.Weekday().String())

	// Check if current day is in the allowed days
	dayAllowed := false
	for _, day := range period.Days {
		if strings.ToLower(day) == weekday {
			dayAllowed = true
			break
		}
	}
	if !dayAllowed {
		return false
	}

	// Check time range
	currentTime := now.Format("15:04")
	return currentTime >= period.Start && currentTime <= period.End
}

// ShouldThrottle checks if an alert should be throttled based on throttling configuration
func (at *AlertThrottler) ShouldThrottle(ruleID string, throttle *AlertThrottle) (bool, error) {
	at.mutex.Lock()
	defer at.mutex.Unlock()

	now := time.Now()

	// Get or create alert count list for this rule
	if _, exists := at.alertCounts[ruleID]; !exists {
		at.alertCounts[ruleID] = make([]time.Time, 0)
	}

	// Clean old entries outside the time window
	cutoff := now.Add(-throttle.TimeWindow)
	var validCounts []time.Time
	for _, timestamp := range at.alertCounts[ruleID] {
		if timestamp.After(cutoff) {
			validCounts = append(validCounts, timestamp)
		}
	}
	at.alertCounts[ruleID] = validCounts

	// Check if we've exceeded the limit
	if len(at.alertCounts[ruleID]) >= throttle.MaxAlerts {
		return true, nil
	}

	// Add current alert to count
	at.alertCounts[ruleID] = append(at.alertCounts[ruleID], now)
	return false, nil
}

// ExecuteAction executes a single alert action
func (aae *AlertActionExecutor) ExecuteAction(ctx context.Context, action *AlertAction, alertData map[string]interface{}) error {
	// Add delay if configured
	if action.Delay > 0 {
		time.Sleep(action.Delay)
	}

	var err error
	switch action.Type {
	case "notification":
		err = aae.executeNotificationAction(ctx, action, alertData)
	case "webhook":
		err = aae.executeWebhookAction(ctx, action, alertData)
	case "script":
		err = aae.executeScriptAction(ctx, action, alertData)
	case "escalation":
		err = aae.executeEscalationAction(ctx, action, alertData)
	default:
		return fmt.Errorf("unknown action type: %s", action.Type)
	}

	// Handle retries if configured
	if err != nil && action.Retry != nil {
		return aae.executeWithRetry(ctx, action, alertData)
	}

	return err
}

// executeNotificationAction executes notification actions
func (aae *AlertActionExecutor) executeNotificationAction(ctx context.Context, action *AlertAction, alertData map[string]interface{}) error {
	aae.logger.Info("Executing notification action",
		zap.String("target", action.Target),
		zap.Any("data", alertData))

	// This would integrate with the existing notification system
	// Implementation would depend on the target (slack, email, etc.)
	return nil
}

// executeWebhookAction executes webhook actions
func (aae *AlertActionExecutor) executeWebhookAction(ctx context.Context, action *AlertAction, alertData map[string]interface{}) error {
	aae.logger.Info("Executing webhook action",
		zap.String("target", action.Target),
		zap.Any("data", alertData))

	// Implementation would make HTTP POST to the webhook URL
	return nil
}

// executeScriptAction executes script actions
func (aae *AlertActionExecutor) executeScriptAction(ctx context.Context, action *AlertAction, alertData map[string]interface{}) error {
	aae.logger.Info("Executing script action",
		zap.String("target", action.Target),
		zap.Any("data", alertData))

	// Implementation would execute the specified script
	return nil
}

// executeEscalationAction executes escalation actions
func (aae *AlertActionExecutor) executeEscalationAction(ctx context.Context, action *AlertAction, alertData map[string]interface{}) error {
	aae.logger.Info("Executing escalation action",
		zap.String("target", action.Target),
		zap.Any("data", alertData))

	// Implementation would trigger escalation to higher-level alerts
	return nil
}

// executeWithRetry executes an action with retry logic
func (aae *AlertActionExecutor) executeWithRetry(ctx context.Context, action *AlertAction, alertData map[string]interface{}) error {
	var lastErr error

	for attempt := 1; attempt <= action.Retry.MaxAttempts; attempt++ {
		if attempt > 1 {
			// Apply backoff
			delay := action.Retry.Interval
			if action.Retry.Backoff == "exponential" {
				delay = time.Duration(attempt*attempt) * action.Retry.Interval
			}
			time.Sleep(delay)
		}

		lastErr = aae.ExecuteAction(ctx, action, alertData)
		if lastErr == nil {
			return nil
		}

		aae.logger.Warn("Action retry failed",
			zap.Int("attempt", attempt),
			zap.Error(lastErr))
	}

	return fmt.Errorf("action failed after %d attempts: %w", action.Retry.MaxAttempts, lastErr)
}
