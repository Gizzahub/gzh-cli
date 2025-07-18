package github

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// WebhookMonitor monitors webhook status and health.
type WebhookMonitor struct {
	logger      Logger
	apiClient   APIClient
	config      *WebhookMonitorConfig
	metrics     *WebhookMetrics
	webhooks    map[string]*WebhookStatus
	mu          sync.RWMutex
	stopChannel chan struct{}
	running     bool
}

// WebhookMonitorConfig holds configuration for webhook monitoring.
type WebhookMonitorConfig struct {
	CheckInterval       time.Duration   `json:"check_interval" yaml:"check_interval"`
	HealthCheckTimeout  time.Duration   `json:"health_check_timeout" yaml:"health_check_timeout"`
	RetentionPeriod     time.Duration   `json:"retention_period" yaml:"retention_period"`
	AlertThresholds     AlertThresholds `json:"alert_thresholds" yaml:"alert_thresholds"`
	EnableNotifications bool            `json:"enable_notifications" yaml:"enable_notifications"`
	MaxHistorySize      int             `json:"max_history_size" yaml:"max_history_size"`
}

// AlertThresholds defines thresholds for different alert levels.
type AlertThresholds struct {
	ErrorRate          float64       `json:"error_rate" yaml:"error_rate"`                     // Percentage
	ResponseTime       time.Duration `json:"response_time" yaml:"response_time"`               // Maximum acceptable response time
	FailureCount       int           `json:"failure_count" yaml:"failure_count"`               // Consecutive failures
	DeliveryFailureAge time.Duration `json:"delivery_failure_age" yaml:"delivery_failure_age"` // Age of oldest delivery failure
}

// WebhookStatus represents the current status of a webhook.
type WebhookStatus struct {
	ID           string                 `json:"id"`
	URL          string                 `json:"url"`
	Organization string                 `json:"organization"`
	Repository   string                 `json:"repository,omitempty"`
	Events       []string               `json:"events"`
	Active       bool                   `json:"active"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	LastChecked  time.Time              `json:"last_checked"`
	Status       WebhookHealthStatus    `json:"status"`
	Metrics      WebhookStatusMetrics   `json:"metrics"`
	Config       map[string]interface{} `json:"config"`
	Alerts       []WebhookAlert         `json:"alerts"`
	History      []WebhookHealthCheck   `json:"history"`
}

// WebhookHealthStatus represents the health status of a webhook.
type WebhookHealthStatus string

const (
	WebhookStatusHealthy   WebhookHealthStatus = "healthy"
	WebhookStatusDegraded  WebhookHealthStatus = "degraded"
	WebhookStatusUnhealthy WebhookHealthStatus = "unhealthy"
	WebhookStatusUnknown   WebhookHealthStatus = "unknown"
	WebhookStatusDisabled  WebhookHealthStatus = "disabled"
)

// WebhookStatusMetrics holds metrics for a specific webhook.
type WebhookStatusMetrics struct {
	TotalDeliveries      int64         `json:"total_deliveries"`
	SuccessfulDeliveries int64         `json:"successful_deliveries"`
	FailedDeliveries     int64         `json:"failed_deliveries"`
	AverageResponseTime  time.Duration `json:"average_response_time"`
	LastDeliveryTime     time.Time     `json:"last_delivery_time"`
	LastSuccessTime      time.Time     `json:"last_success_time"`
	LastFailureTime      time.Time     `json:"last_failure_time"`
	ConsecutiveFailures  int           `json:"consecutive_failures"`
	ErrorRate            float64       `json:"error_rate"`
	Uptime               float64       `json:"uptime"`
}

// WebhookAlert represents an alert for a webhook.
type WebhookAlert struct {
	ID           string                 `json:"id"`
	WebhookID    string                 `json:"webhook_id"`
	Type         WebhookAlertType       `json:"type"`
	Severity     WebhookAlertSeverity   `json:"severity"`
	Message      string                 `json:"message"`
	CreatedAt    time.Time              `json:"created_at"`
	ResolvedAt   *time.Time             `json:"resolved_at,omitempty"`
	Acknowledged bool                   `json:"acknowledged"`
	Details      map[string]interface{} `json:"details,omitempty"`
}

// WebhookAlertType defines types of webhook alerts.
type WebhookAlertType string

const (
	AlertTypeHighErrorRate       WebhookAlertType = "high_error_rate"
	AlertTypeSlowResponse        WebhookAlertType = "slow_response"
	AlertTypeConsecutiveFailures WebhookAlertType = "consecutive_failures"
	AlertTypeConfigurationIssue  WebhookAlertType = "configuration_issue"
	AlertTypeDeliveryFailure     WebhookAlertType = "delivery_failure"
	AlertTypeEndpointDown        WebhookAlertType = "endpoint_down"
)

// WebhookAlertSeverity defines severity levels for alerts.
type WebhookAlertSeverity string

const (
	AlertSeverityInfo     WebhookAlertSeverity = "info"
	AlertSeverityWarning  WebhookAlertSeverity = "warning"
	AlertSeverityError    WebhookAlertSeverity = "error"
	AlertSeverityCritical WebhookAlertSeverity = "critical"
)

// WebhookHealthCheck represents a health check result.
type WebhookHealthCheck struct {
	Timestamp    time.Time              `json:"timestamp"`
	Status       WebhookHealthStatus    `json:"status"`
	ResponseTime time.Duration          `json:"response_time"`
	StatusCode   int                    `json:"status_code,omitempty"`
	Error        string                 `json:"error,omitempty"`
	Details      map[string]interface{} `json:"details,omitempty"`
}

// WebhookMetrics holds global webhook metrics.
type WebhookMetrics struct {
	mu                   sync.RWMutex
	TotalWebhooks        int64                           `json:"total_webhooks"`
	ActiveWebhooks       int64                           `json:"active_webhooks"`
	HealthyWebhooks      int64                           `json:"healthy_webhooks"`
	UnhealthyWebhooks    int64                           `json:"unhealthy_webhooks"`
	TotalDeliveries      int64                           `json:"total_deliveries"`
	SuccessfulDeliveries int64                           `json:"successful_deliveries"`
	FailedDeliveries     int64                           `json:"failed_deliveries"`
	AverageResponseTime  time.Duration                   `json:"average_response_time"`
	ActiveAlerts         int64                           `json:"active_alerts"`
	StatusDistribution   map[WebhookHealthStatus]int64   `json:"status_distribution"`
	OrganizationMetrics  map[string]*OrganizationMetrics `json:"organization_metrics"`
	LastUpdated          time.Time                       `json:"last_updated"`
}

// OrganizationMetrics holds metrics for a specific organization.
type OrganizationMetrics struct {
	TotalWebhooks       int64         `json:"total_webhooks"`
	HealthyWebhooks     int64         `json:"healthy_webhooks"`
	UnhealthyWebhooks   int64         `json:"unhealthy_webhooks"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	ErrorRate           float64       `json:"error_rate"`
	ActiveAlerts        int64         `json:"active_alerts"`
}

// NewWebhookMonitor creates a new webhook monitor.
func NewWebhookMonitor(logger Logger, apiClient APIClient, config *WebhookMonitorConfig) *WebhookMonitor {
	if config == nil {
		config = getDefaultWebhookMonitorConfig()
	}

	return &WebhookMonitor{
		logger:      logger,
		apiClient:   apiClient,
		config:      config,
		metrics:     newWebhookMetrics(),
		webhooks:    make(map[string]*WebhookStatus),
		stopChannel: make(chan struct{}),
		running:     false,
	}
}

// Start starts the webhook monitoring service.
func (wm *WebhookMonitor) Start(ctx context.Context) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	if wm.running {
		return fmt.Errorf("webhook monitor is already running")
	}

	wm.logger.Info("Starting webhook monitor", "check_interval", wm.config.CheckInterval)
	wm.running = true

	// Start monitoring goroutine
	go wm.monitorLoop(ctx)

	// Start metrics collection goroutine
	go wm.metricsCollector(ctx)

	// Start alert processor goroutine
	go wm.alertProcessor(ctx)

	wm.logger.Info("Webhook monitor started successfully")

	return nil
}

// Stop stops the webhook monitoring service.
func (wm *WebhookMonitor) Stop(ctx context.Context) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	if !wm.running {
		return fmt.Errorf("webhook monitor is not running")
	}

	wm.logger.Info("Stopping webhook monitor")
	close(wm.stopChannel)
	wm.running = false

	wm.logger.Info("Webhook monitor stopped successfully")

	return nil
}

// GetWebhookStatus returns the status of a specific webhook.
func (wm *WebhookMonitor) GetWebhookStatus(webhookID string) (*WebhookStatus, error) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	status, exists := wm.webhooks[webhookID]
	if !exists {
		return nil, fmt.Errorf("webhook not found: %s", webhookID)
	}

	// Return a copy to avoid data races
	statusCopy := *status

	return &statusCopy, nil
}

// GetAllWebhookStatuses returns the status of all monitored webhooks.
func (wm *WebhookMonitor) GetAllWebhookStatuses() map[string]*WebhookStatus {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	statuses := make(map[string]*WebhookStatus)

	for id, status := range wm.webhooks {
		statusCopy := *status
		statuses[id] = &statusCopy
	}

	return statuses
}

// GetMetrics returns current webhook metrics.
func (wm *WebhookMonitor) GetMetrics() *WebhookMetrics {
	wm.metrics.mu.RLock()
	defer wm.metrics.mu.RUnlock()

	// Create a copy to avoid data races
	metricsCopy := *wm.metrics
	metricsCopy.StatusDistribution = make(map[WebhookHealthStatus]int64)
	metricsCopy.OrganizationMetrics = make(map[string]*OrganizationMetrics)

	for k, v := range wm.metrics.StatusDistribution {
		metricsCopy.StatusDistribution[k] = v
	}

	for k, v := range wm.metrics.OrganizationMetrics {
		orgMetricsCopy := *v
		metricsCopy.OrganizationMetrics[k] = &orgMetricsCopy
	}

	return &metricsCopy
}

// GetActiveAlerts returns all active alerts.
func (wm *WebhookMonitor) GetActiveAlerts() []WebhookAlert {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	var alerts []WebhookAlert

	for _, webhook := range wm.webhooks {
		for _, alert := range webhook.Alerts {
			if alert.ResolvedAt == nil {
				alerts = append(alerts, alert)
			}
		}
	}

	return alerts
}

// AddWebhook adds a webhook to the monitor (for testing/demo purposes).
func (wm *WebhookMonitor) AddWebhook(webhook *WebhookStatus) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	wm.webhooks[webhook.ID] = webhook
}

// AcknowledgeAlert marks an alert as acknowledged.
func (wm *WebhookMonitor) AcknowledgeAlert(alertID string) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	for _, webhook := range wm.webhooks {
		for i := range webhook.Alerts {
			if webhook.Alerts[i].ID == alertID {
				webhook.Alerts[i].Acknowledged = true
				wm.logger.Info("Alert acknowledged", "alert_id", alertID, "webhook_id", webhook.ID)

				return nil
			}
		}
	}

	return fmt.Errorf("alert not found: %s", alertID)
}

// Private methods

func (wm *WebhookMonitor) monitorLoop(ctx context.Context) {
	ticker := time.NewTicker(wm.config.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			wm.performHealthChecks(ctx)
		case <-wm.stopChannel:
			return
		case <-ctx.Done():
			return
		}
	}
}

func (wm *WebhookMonitor) performHealthChecks(ctx context.Context) {
	wm.logger.Debug("Performing webhook health checks")

	// This would integrate with the existing webhook service
	// For now, we'll implement a basic health check simulation
	wm.mu.Lock()
	defer wm.mu.Unlock()

	// Update metrics
	wm.updateMetrics()

	wm.logger.Debug("Health checks completed", "webhook_count", len(wm.webhooks))
}

func (wm *WebhookMonitor) metricsCollector(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			wm.collectMetrics()
		case <-wm.stopChannel:
			return
		case <-ctx.Done():
			return
		}
	}
}

func (wm *WebhookMonitor) alertProcessor(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			wm.processAlerts()
		case <-wm.stopChannel:
			return
		case <-ctx.Done():
			return
		}
	}
}

func (wm *WebhookMonitor) updateMetrics() {
	wm.metrics.mu.Lock()
	defer wm.metrics.mu.Unlock()

	wm.metrics.TotalWebhooks = int64(len(wm.webhooks))
	wm.metrics.LastUpdated = time.Now()

	// Reset counters
	wm.metrics.ActiveWebhooks = 0
	wm.metrics.HealthyWebhooks = 0
	wm.metrics.UnhealthyWebhooks = 0
	wm.metrics.ActiveAlerts = 0

	if wm.metrics.StatusDistribution == nil {
		wm.metrics.StatusDistribution = make(map[WebhookHealthStatus]int64)
	}

	if wm.metrics.OrganizationMetrics == nil {
		wm.metrics.OrganizationMetrics = make(map[string]*OrganizationMetrics)
	}

	// Clear previous distribution
	for k := range wm.metrics.StatusDistribution {
		wm.metrics.StatusDistribution[k] = 0
	}

	// Count webhooks by status
	for _, webhook := range wm.webhooks {
		if webhook.Active {
			wm.metrics.ActiveWebhooks++
		}

		wm.metrics.StatusDistribution[webhook.Status]++

		switch webhook.Status {
		case WebhookStatusHealthy:
			wm.metrics.HealthyWebhooks++
		case WebhookStatusUnhealthy, WebhookStatusDegraded:
			wm.metrics.UnhealthyWebhooks++
		}

		// Count active alerts
		for _, alert := range webhook.Alerts {
			if alert.ResolvedAt == nil {
				wm.metrics.ActiveAlerts++
			}
		}

		// Update organization metrics
		orgMetrics, exists := wm.metrics.OrganizationMetrics[webhook.Organization]
		if !exists {
			orgMetrics = &OrganizationMetrics{}
			wm.metrics.OrganizationMetrics[webhook.Organization] = orgMetrics
		}

		orgMetrics.TotalWebhooks++

		if webhook.Status == WebhookStatusHealthy {
			orgMetrics.HealthyWebhooks++
		} else {
			orgMetrics.UnhealthyWebhooks++
		}
	}
}

func (wm *WebhookMonitor) collectMetrics() {
	// This would collect detailed metrics from webhook deliveries
	wm.logger.Debug("Collecting webhook metrics")
}

func (wm *WebhookMonitor) processAlerts() {
	// This would process and generate alerts based on thresholds
	wm.logger.Debug("Processing webhook alerts")
}

func newWebhookMetrics() *WebhookMetrics {
	return &WebhookMetrics{
		StatusDistribution:  make(map[WebhookHealthStatus]int64),
		OrganizationMetrics: make(map[string]*OrganizationMetrics),
		LastUpdated:         time.Now(),
	}
}

func getDefaultWebhookMonitorConfig() *WebhookMonitorConfig {
	return &WebhookMonitorConfig{
		CheckInterval:       5 * time.Minute,
		HealthCheckTimeout:  30 * time.Second,
		RetentionPeriod:     24 * time.Hour,
		EnableNotifications: true,
		MaxHistorySize:      100,
		AlertThresholds: AlertThresholds{
			ErrorRate:          10.0, // 10% error rate
			ResponseTime:       5 * time.Second,
			FailureCount:       5,
			DeliveryFailureAge: 1 * time.Hour,
		},
	}
}
