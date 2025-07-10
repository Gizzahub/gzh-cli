package monitoring

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

// AlertSeverity represents alert severity levels
type AlertSeverity string

const (
	AlertSeverityCritical AlertSeverity = "critical"
	AlertSeverityHigh     AlertSeverity = "high"
	AlertSeverityMedium   AlertSeverity = "medium"
	AlertSeverityLow      AlertSeverity = "low"
	AlertSeverityInfo     AlertSeverity = "info"
)

// AlertStatus represents alert status
type AlertStatus string

const (
	AlertStatusFiring   AlertStatus = "firing"
	AlertStatusResolved AlertStatus = "resolved"
	AlertStatusSilenced AlertStatus = "silenced"
)

// AlertRule represents an alert rule configuration
type AlertRule struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Query       string        `json:"query"`
	Threshold   float64       `json:"threshold"`
	Duration    time.Duration `json:"duration"`
	Severity    AlertSeverity `json:"severity"`
	Enabled     bool          `json:"enabled"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

// AlertInstance represents a fired alert instance
type AlertInstance struct {
	ID          string            `json:"id"`
	RuleID      string            `json:"rule_id"`
	RuleName    string            `json:"rule_name"`
	Status      AlertStatus       `json:"status"`
	Severity    AlertSeverity     `json:"severity"`
	Message     string            `json:"message"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	Value       float64           `json:"value"`
	Threshold   float64           `json:"threshold"`
	StartsAt    time.Time         `json:"starts_at"`
	EndsAt      *time.Time        `json:"ends_at,omitempty"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// AlertManager manages alerts and alert rules
type AlertManager struct {
	mu        sync.RWMutex
	rules     map[string]*AlertRule
	alerts    map[string]*AlertInstance
	silences  map[string]time.Time
	evaluator *AlertEvaluator
}

// AlertEvaluator evaluates alert rules
type AlertEvaluator struct {
	metrics *MetricsCollector
}

// NewAlertManager creates a new alert manager
func NewAlertManager() *AlertManager {
	return &AlertManager{
		rules:     make(map[string]*AlertRule),
		alerts:    make(map[string]*AlertInstance),
		silences:  make(map[string]time.Time),
		evaluator: &AlertEvaluator{},
	}
}

// SetMetrics sets the metrics collector for alert evaluation
func (am *AlertManager) SetMetrics(metrics *MetricsCollector) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.evaluator.metrics = metrics
}

// CreateRule creates a new alert rule
func (am *AlertManager) CreateRule(rule *AlertRule) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if rule.ID == "" {
		rule.ID = generateID()
	}

	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()

	am.rules[rule.ID] = rule
	return nil
}

// UpdateRule updates an existing alert rule
func (am *AlertManager) UpdateRule(rule *AlertRule) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	existing, exists := am.rules[rule.ID]
	if !exists {
		return fmt.Errorf("rule not found: %s", rule.ID)
	}

	rule.CreatedAt = existing.CreatedAt
	rule.UpdatedAt = time.Now()
	am.rules[rule.ID] = rule

	return nil
}

// DeleteRule deletes an alert rule
func (am *AlertManager) DeleteRule(ruleID string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if _, exists := am.rules[ruleID]; !exists {
		return fmt.Errorf("rule not found: %s", ruleID)
	}

	delete(am.rules, ruleID)

	// Also resolve any active alerts for this rule
	for _, alert := range am.alerts {
		if alert.RuleID == ruleID && alert.Status == AlertStatusFiring {
			alert.Status = AlertStatusResolved
			now := time.Now()
			alert.EndsAt = &now
			alert.UpdatedAt = now
		}
	}

	return nil
}

// GetRule gets an alert rule by ID
func (am *AlertManager) GetRule(ruleID string) (*AlertRule, error) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	rule, exists := am.rules[ruleID]
	if !exists {
		return nil, fmt.Errorf("rule not found: %s", ruleID)
	}

	return rule, nil
}

// ListRules lists all alert rules
func (am *AlertManager) ListRules() []*AlertRule {
	am.mu.RLock()
	defer am.mu.RUnlock()

	rules := make([]*AlertRule, 0, len(am.rules))
	for _, rule := range am.rules {
		rules = append(rules, rule)
	}

	// Sort by creation time
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].CreatedAt.Before(rules[j].CreatedAt)
	})

	return rules
}

// CreateAlert creates a new alert instance
func (am *AlertManager) CreateAlert(alert *Alert) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if alert.ID == "" {
		alert.ID = generateID()
	}

	alert.CreatedAt = time.Now()
	alert.UpdatedAt = time.Now()

	// Convert Alert to AlertInstance for internal storage
	instance := &AlertInstance{
		ID:          alert.ID,
		RuleID:      alert.Name, // Using name as rule reference for standalone alerts
		RuleName:    alert.Name,
		Status:      AlertStatusFiring,
		Severity:    AlertSeverity(alert.Severity),
		Message:     alert.Description,
		Labels:      make(map[string]string),
		Annotations: make(map[string]string),
		StartsAt:    alert.CreatedAt,
		UpdatedAt:   alert.UpdatedAt,
	}

	am.alerts[alert.ID] = instance
	return nil
}

// UpdateAlert updates an existing alert
func (am *AlertManager) UpdateAlert(alert *Alert) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	existing, exists := am.alerts[alert.ID]
	if !exists {
		return fmt.Errorf("alert not found: %s", alert.ID)
	}

	// Update fields
	existing.Status = AlertStatus(alert.Status)
	existing.Severity = AlertSeverity(alert.Severity)
	existing.Message = alert.Description
	existing.UpdatedAt = time.Now()

	if alert.Status == "resolved" && existing.EndsAt == nil {
		now := time.Now()
		existing.EndsAt = &now
	}

	return nil
}

// DeleteAlert deletes an alert
func (am *AlertManager) DeleteAlert(alertID string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if _, exists := am.alerts[alertID]; !exists {
		return fmt.Errorf("alert not found: %s", alertID)
	}

	delete(am.alerts, alertID)
	return nil
}

// GetAlerts gets all alerts
func (am *AlertManager) GetAlerts() ([]*AlertInstance, error) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	alerts := make([]*AlertInstance, 0, len(am.alerts))
	for _, alert := range am.alerts {
		alerts = append(alerts, alert)
	}

	// Sort by start time (newest first)
	sort.Slice(alerts, func(i, j int) bool {
		return alerts[i].StartsAt.After(alerts[j].StartsAt)
	})

	return alerts, nil
}

// GetAlert gets a specific alert by ID
func (am *AlertManager) GetAlert(alertID string) (*AlertInstance, error) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	alert, exists := am.alerts[alertID]
	if !exists {
		return nil, fmt.Errorf("alert not found: %s", alertID)
	}

	return alert, nil
}

// SilenceAlert silences an alert for a specified duration
func (am *AlertManager) SilenceAlert(alertID string, duration time.Duration) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	alert, exists := am.alerts[alertID]
	if !exists {
		return fmt.Errorf("alert not found: %s", alertID)
	}

	alert.Status = AlertStatusSilenced
	alert.UpdatedAt = time.Now()
	am.silences[alertID] = time.Now().Add(duration)

	return nil
}

// ResolveAlert resolves an alert
func (am *AlertManager) ResolveAlert(alertID string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	alert, exists := am.alerts[alertID]
	if !exists {
		return fmt.Errorf("alert not found: %s", alertID)
	}

	alert.Status = AlertStatusResolved
	now := time.Now()
	alert.EndsAt = &now
	alert.UpdatedAt = now

	// Remove from silences if it was silenced
	delete(am.silences, alertID)

	return nil
}

// EvaluateRules evaluates all alert rules
func (am *AlertManager) EvaluateRules(ctx context.Context) error {
	am.mu.RLock()
	rules := make([]*AlertRule, 0, len(am.rules))
	for _, rule := range am.rules {
		if rule.Enabled {
			rules = append(rules, rule)
		}
	}
	am.mu.RUnlock()

	for _, rule := range rules {
		if err := am.evaluateRule(ctx, rule); err != nil {
			// Log error but continue with other rules
			fmt.Printf("Error evaluating rule %s: %v\n", rule.Name, err)
		}
	}

	// Check for silences that have expired
	am.checkExpiredSilences()

	return nil
}

// evaluateRule evaluates a single alert rule
func (am *AlertManager) evaluateRule(ctx context.Context, rule *AlertRule) error {
	if am.evaluator.metrics == nil {
		return fmt.Errorf("metrics collector not set")
	}

	// Simple rule evaluation based on query
	value := am.evaluateQuery(rule.Query)

	if value > rule.Threshold {
		// Check if alert already exists
		alertID := fmt.Sprintf("%s_alert", rule.ID)

		am.mu.Lock()
		existing, exists := am.alerts[alertID]

		if !exists {
			// Create new alert
			alert := &AlertInstance{
				ID:       alertID,
				RuleID:   rule.ID,
				RuleName: rule.Name,
				Status:   AlertStatusFiring,
				Severity: rule.Severity,
				Message:  fmt.Sprintf("%s: %.2f > %.2f", rule.Description, value, rule.Threshold),
				Labels: map[string]string{
					"rule_id":   rule.ID,
					"rule_name": rule.Name,
				},
				Annotations: map[string]string{
					"description": rule.Description,
					"threshold":   fmt.Sprintf("%.2f", rule.Threshold),
					"value":       fmt.Sprintf("%.2f", value),
				},
				Value:     value,
				Threshold: rule.Threshold,
				StartsAt:  time.Now(),
				UpdatedAt: time.Now(),
			}
			am.alerts[alertID] = alert
		} else if existing.Status == AlertStatusResolved {
			// Re-fire resolved alert
			existing.Status = AlertStatusFiring
			existing.StartsAt = time.Now()
			existing.UpdatedAt = time.Now()
			existing.EndsAt = nil
			existing.Value = value
		}
		am.mu.Unlock()
	} else {
		// Check if we should resolve an existing alert
		alertID := fmt.Sprintf("%s_alert", rule.ID)

		am.mu.Lock()
		if existing, exists := am.alerts[alertID]; exists && existing.Status == AlertStatusFiring {
			existing.Status = AlertStatusResolved
			now := time.Now()
			existing.EndsAt = &now
			existing.UpdatedAt = now
		}
		am.mu.Unlock()
	}

	return nil
}

// evaluateQuery evaluates a simple query against metrics
func (am *AlertManager) evaluateQuery(query string) float64 {
	if am.evaluator.metrics == nil {
		return 0
	}

	// Simple query evaluation - in production this would be more sophisticated
	switch query {
	case "memory_usage_percent":
		return float64(am.evaluator.metrics.GetMemoryUsage()) / 1024 / 1024 // MB
	case "cpu_usage_percent":
		return am.evaluator.metrics.GetCPUUsage()
	case "error_rate_percent":
		return am.evaluator.metrics.GetErrorRate()
	case "active_tasks":
		return float64(am.evaluator.metrics.GetActiveTasks())
	case "response_time_ms":
		return float64(am.evaluator.metrics.GetAverageResponseTime().Nanoseconds()) / 1000000
	default:
		return 0
	}
}

// checkExpiredSilences checks for expired silences and updates alert status
func (am *AlertManager) checkExpiredSilences() {
	am.mu.Lock()
	defer am.mu.Unlock()

	now := time.Now()
	for alertID, expiry := range am.silences {
		if now.After(expiry) {
			// Silence has expired
			if alert, exists := am.alerts[alertID]; exists && alert.Status == AlertStatusSilenced {
				alert.Status = AlertStatusFiring
				alert.UpdatedAt = now
			}
			delete(am.silences, alertID)
		}
	}
}

// StartEvaluationLoop starts the alert rule evaluation loop
func (am *AlertManager) StartEvaluationLoop(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			am.EvaluateRules(ctx)
		}
	}
}

// GetAlertStats returns statistics about alerts
func (am *AlertManager) GetAlertStats() map[string]interface{} {
	am.mu.RLock()
	defer am.mu.RUnlock()

	stats := map[string]interface{}{
		"total_rules":     len(am.rules),
		"total_alerts":    len(am.alerts),
		"firing_alerts":   0,
		"resolved_alerts": 0,
		"silenced_alerts": 0,
	}

	for _, alert := range am.alerts {
		switch alert.Status {
		case AlertStatusFiring:
			stats["firing_alerts"] = stats["firing_alerts"].(int) + 1
		case AlertStatusResolved:
			stats["resolved_alerts"] = stats["resolved_alerts"].(int) + 1
		case AlertStatusSilenced:
			stats["silenced_alerts"] = stats["silenced_alerts"].(int) + 1
		}
	}

	return stats
}

// generateID generates a unique ID for alerts and rules
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
