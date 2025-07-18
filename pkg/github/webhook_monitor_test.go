package github

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mock implementations are defined in automation_engine_test.go
// Using shared mockAPIClient to avoid redeclaration

// Test helper functions

func createTestWebhookMonitor() (*WebhookMonitor, *mockAPIClient) {
	logger := &mockLogger{}
	apiClient := &mockAPIClient{}

	config := &WebhookMonitorConfig{
		CheckInterval:       1 * time.Second,
		HealthCheckTimeout:  5 * time.Second,
		RetentionPeriod:     1 * time.Hour,
		EnableNotifications: true,
		MaxHistorySize:      10,
		AlertThresholds: AlertThresholds{
			ErrorRate:          15.0,
			ResponseTime:       2 * time.Second,
			FailureCount:       3,
			DeliveryFailureAge: 30 * time.Minute,
		},
	}

	monitor := NewWebhookMonitor(logger, apiClient, config)

	return monitor, apiClient
}

func createTestWebhookStatus() *WebhookStatus {
	return &WebhookStatus{
		ID:           "webhook-001",
		URL:          "https://example.com/webhook",
		Organization: "testorg",
		Repository:   "test-repo",
		Events:       []string{"push", "pull_request"},
		Active:       true,
		CreatedAt:    time.Now().Add(-24 * time.Hour),
		UpdatedAt:    time.Now().Add(-1 * time.Hour),
		LastChecked:  time.Now().Add(-5 * time.Minute),
		Status:       WebhookStatusHealthy,
		Metrics: WebhookStatusMetrics{
			TotalDeliveries:      1000,
			SuccessfulDeliveries: 950,
			FailedDeliveries:     50,
			AverageResponseTime:  500 * time.Millisecond,
			LastDeliveryTime:     time.Now().Add(-1 * time.Minute),
			LastSuccessTime:      time.Now().Add(-1 * time.Minute),
			LastFailureTime:      time.Now().Add(-10 * time.Minute),
			ConsecutiveFailures:  0,
			ErrorRate:            5.0,
			Uptime:               95.0,
		},
		Config: map[string]interface{}{
			"content_type": "application/json",
			"secret":       "webhook-secret",
		},
		Alerts:  []WebhookAlert{},
		History: []WebhookHealthCheck{},
	}
}

func createTestWebhookAlert() WebhookAlert {
	return WebhookAlert{
		ID:        "alert-001",
		WebhookID: "webhook-001",
		Type:      AlertTypeHighErrorRate,
		Severity:  AlertSeverityWarning,
		Message:   "High error rate detected",
		CreatedAt: time.Now().Add(-30 * time.Minute),
		Details: map[string]interface{}{
			"error_rate": 15.5,
			"threshold":  15.0,
		},
	}
}

// Test Cases

func TestNewWebhookMonitor(t *testing.T) {
	logger := &mockLogger{}
	apiClient := &mockAPIClient{}

	monitor := NewWebhookMonitor(logger, apiClient, nil)

	assert.NotNil(t, monitor)
	assert.NotNil(t, monitor.config)
	assert.NotNil(t, monitor.metrics)
	assert.NotNil(t, monitor.webhooks)
	assert.False(t, monitor.running)
}

func TestWebhookMonitor_StartStop(t *testing.T) {
	monitor, _ := createTestWebhookMonitor()
	ctx := context.Background()

	// Test start
	err := monitor.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, monitor.running)

	// Test start when already running
	err = monitor.Start(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already running")

	// Test stop
	err = monitor.Stop(ctx)
	assert.NoError(t, err)
	assert.False(t, monitor.running)

	// Test stop when not running
	err = monitor.Stop(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")
}

func TestWebhookMonitor_GetWebhookStatus(t *testing.T) {
	monitor, _ := createTestWebhookMonitor()

	// Test webhook not found
	_, err := monitor.GetWebhookStatus("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "webhook not found")

	// Add test webhook
	testWebhook := createTestWebhookStatus()
	monitor.webhooks[testWebhook.ID] = testWebhook

	// Test successful retrieval
	status, err := monitor.GetWebhookStatus(testWebhook.ID)
	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.Equal(t, testWebhook.ID, status.ID)
	assert.Equal(t, testWebhook.URL, status.URL)
	assert.Equal(t, testWebhook.Status, status.Status)
}

func TestWebhookMonitor_GetAllWebhookStatuses(t *testing.T) {
	monitor, _ := createTestWebhookMonitor()

	// Test empty webhooks
	statuses := monitor.GetAllWebhookStatuses()
	assert.NotNil(t, statuses)
	assert.Len(t, statuses, 0)

	// Add test webhooks
	webhook1 := createTestWebhookStatus()
	webhook1.ID = "webhook-001"
	webhook2 := createTestWebhookStatus()
	webhook2.ID = "webhook-002"
	webhook2.Organization = "another-org"

	monitor.webhooks[webhook1.ID] = webhook1
	monitor.webhooks[webhook2.ID] = webhook2

	// Test retrieval with webhooks
	statuses = monitor.GetAllWebhookStatuses()
	assert.Len(t, statuses, 2)
	assert.Contains(t, statuses, webhook1.ID)
	assert.Contains(t, statuses, webhook2.ID)
}

func TestWebhookMonitor_GetMetrics(t *testing.T) {
	monitor, _ := createTestWebhookMonitor()

	// Test initial metrics
	metrics := monitor.GetMetrics()
	assert.NotNil(t, metrics)
	assert.Equal(t, int64(0), metrics.TotalWebhooks)
	assert.NotNil(t, metrics.StatusDistribution)
	assert.NotNil(t, metrics.OrganizationMetrics)

	// Add test webhooks and update metrics
	webhook1 := createTestWebhookStatus()
	webhook1.Status = WebhookStatusHealthy
	webhook2 := createTestWebhookStatus()
	webhook2.ID = "webhook-002"
	webhook2.Status = WebhookStatusUnhealthy

	monitor.webhooks[webhook1.ID] = webhook1
	monitor.webhooks[webhook2.ID] = webhook2
	monitor.updateMetrics()

	// Test updated metrics
	metrics = monitor.GetMetrics()
	assert.Equal(t, int64(2), metrics.TotalWebhooks)
	assert.Equal(t, int64(1), metrics.HealthyWebhooks)
	assert.Equal(t, int64(1), metrics.UnhealthyWebhooks)
	assert.Equal(t, int64(1), metrics.StatusDistribution[WebhookStatusHealthy])
	assert.Equal(t, int64(1), metrics.StatusDistribution[WebhookStatusUnhealthy])
}

func TestWebhookMonitor_GetActiveAlerts(t *testing.T) {
	monitor, _ := createTestWebhookMonitor()

	// Test no alerts
	alerts := monitor.GetActiveAlerts()
	assert.NotNil(t, alerts)
	assert.Len(t, alerts, 0)

	// Add webhook with alerts
	webhook := createTestWebhookStatus()
	activeAlert := createTestWebhookAlert()
	resolvedAlert := createTestWebhookAlert()
	resolvedAlert.ID = "alert-002"
	resolvedTime := time.Now()
	resolvedAlert.ResolvedAt = &resolvedTime

	webhook.Alerts = []WebhookAlert{activeAlert, resolvedAlert}
	monitor.webhooks[webhook.ID] = webhook

	// Test active alerts only
	alerts = monitor.GetActiveAlerts()
	assert.Len(t, alerts, 1)
	assert.Equal(t, activeAlert.ID, alerts[0].ID)
	assert.Nil(t, alerts[0].ResolvedAt)
}

func TestWebhookMonitor_AcknowledgeAlert(t *testing.T) {
	monitor, _ := createTestWebhookMonitor()

	// Test alert not found
	err := monitor.AcknowledgeAlert("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "alert not found")

	// Add webhook with alert
	webhook := createTestWebhookStatus()
	alert := createTestWebhookAlert()
	webhook.Alerts = []WebhookAlert{alert}
	monitor.webhooks[webhook.ID] = webhook

	// Test successful acknowledgment
	err = monitor.AcknowledgeAlert(alert.ID)
	assert.NoError(t, err)

	// Verify alert was acknowledged
	updatedWebhook := monitor.webhooks[webhook.ID]
	assert.True(t, updatedWebhook.Alerts[0].Acknowledged)
}

func TestWebhookMonitor_UpdateMetrics(t *testing.T) {
	monitor, _ := createTestWebhookMonitor()

	// Add test webhooks with different statuses
	webhooks := []*WebhookStatus{
		{
			ID:           "webhook-001",
			Organization: "org1",
			Active:       true,
			Status:       WebhookStatusHealthy,
			Alerts:       []WebhookAlert{createTestWebhookAlert()},
		},
		{
			ID:           "webhook-002",
			Organization: "org1",
			Active:       true,
			Status:       WebhookStatusUnhealthy,
			Alerts:       []WebhookAlert{},
		},
		{
			ID:           "webhook-003",
			Organization: "org2",
			Active:       false,
			Status:       WebhookStatusDisabled,
			Alerts:       []WebhookAlert{},
		},
	}

	for _, webhook := range webhooks {
		monitor.webhooks[webhook.ID] = webhook
	}

	// Update metrics
	monitor.updateMetrics()

	metrics := monitor.GetMetrics()

	// Verify global metrics
	assert.Equal(t, int64(3), metrics.TotalWebhooks)
	assert.Equal(t, int64(2), metrics.ActiveWebhooks)
	assert.Equal(t, int64(1), metrics.HealthyWebhooks)
	assert.Equal(t, int64(1), metrics.UnhealthyWebhooks)
	assert.Equal(t, int64(1), metrics.ActiveAlerts)

	// Verify status distribution
	assert.Equal(t, int64(1), metrics.StatusDistribution[WebhookStatusHealthy])
	assert.Equal(t, int64(1), metrics.StatusDistribution[WebhookStatusUnhealthy])
	assert.Equal(t, int64(1), metrics.StatusDistribution[WebhookStatusDisabled])

	// Verify organization metrics
	assert.Contains(t, metrics.OrganizationMetrics, "org1")
	assert.Contains(t, metrics.OrganizationMetrics, "org2")

	org1Metrics := metrics.OrganizationMetrics["org1"]
	assert.Equal(t, int64(2), org1Metrics.TotalWebhooks)
	assert.Equal(t, int64(1), org1Metrics.HealthyWebhooks)
	assert.Equal(t, int64(1), org1Metrics.UnhealthyWebhooks)

	org2Metrics := metrics.OrganizationMetrics["org2"]
	assert.Equal(t, int64(1), org2Metrics.TotalWebhooks)
	assert.Equal(t, int64(0), org2Metrics.HealthyWebhooks)
	assert.Equal(t, int64(1), org2Metrics.UnhealthyWebhooks)
}

func TestWebhookStatus_String(t *testing.T) {
	statuses := []WebhookHealthStatus{
		WebhookStatusHealthy,
		WebhookStatusDegraded,
		WebhookStatusUnhealthy,
		WebhookStatusUnknown,
		WebhookStatusDisabled,
	}

	expectedStrings := []string{
		"healthy",
		"degraded",
		"unhealthy",
		"unknown",
		"disabled",
	}

	for i, status := range statuses {
		assert.Equal(t, expectedStrings[i], string(status))
	}
}

func TestAlertTypes_String(t *testing.T) {
	alertTypes := []WebhookAlertType{
		AlertTypeHighErrorRate,
		AlertTypeSlowResponse,
		AlertTypeConsecutiveFailures,
		AlertTypeConfigurationIssue,
		AlertTypeDeliveryFailure,
		AlertTypeEndpointDown,
	}

	expectedStrings := []string{
		"high_error_rate",
		"slow_response",
		"consecutive_failures",
		"configuration_issue",
		"delivery_failure",
		"endpoint_down",
	}

	for i, alertType := range alertTypes {
		assert.Equal(t, expectedStrings[i], string(alertType))
	}
}

func TestAlertSeverities_String(t *testing.T) {
	severities := []WebhookAlertSeverity{
		AlertSeverityInfo,
		AlertSeverityWarning,
		AlertSeverityError,
		AlertSeverityCritical,
	}

	expectedStrings := []string{
		"info",
		"warning",
		"error",
		"critical",
	}

	for i, severity := range severities {
		assert.Equal(t, expectedStrings[i], string(severity))
	}
}

func TestWebhookMonitorConfig_Defaults(t *testing.T) {
	config := getDefaultWebhookMonitorConfig()

	assert.Equal(t, 5*time.Minute, config.CheckInterval)
	assert.Equal(t, 30*time.Second, config.HealthCheckTimeout)
	assert.Equal(t, 24*time.Hour, config.RetentionPeriod)
	assert.True(t, config.EnableNotifications)
	assert.Equal(t, 100, config.MaxHistorySize)
	assert.Equal(t, 10.0, config.AlertThresholds.ErrorRate)
	assert.Equal(t, 5*time.Second, config.AlertThresholds.ResponseTime)
	assert.Equal(t, 5, config.AlertThresholds.FailureCount)
	assert.Equal(t, 1*time.Hour, config.AlertThresholds.DeliveryFailureAge)
}

func TestWebhookMonitor_Integration(t *testing.T) {
	monitor, apiClient := createTestWebhookMonitor()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Set up API client expectations (if needed)
	apiClient.On("ListOrganizationWebhooks", mock.Anything, mock.Anything).Return([]WebhookInfo{}, nil).Maybe()

	// Start monitor
	err := monitor.Start(ctx)
	require.NoError(t, err)

	// Add test webhook
	webhook := createTestWebhookStatus()
	monitor.webhooks[webhook.ID] = webhook

	// Wait for a monitoring cycle
	time.Sleep(2 * time.Second)

	// Verify metrics were updated
	metrics := monitor.GetMetrics()
	assert.NotNil(t, metrics)
	assert.True(t, metrics.LastUpdated.After(time.Now().Add(-10*time.Second)))

	// Stop monitor
	err = monitor.Stop(ctx)
	require.NoError(t, err)
}

// Benchmark tests

func BenchmarkWebhookMonitor_UpdateMetrics(b *testing.B) {
	monitor, _ := createTestWebhookMonitor()

	// Add many test webhooks
	for i := 0; i < 1000; i++ {
		webhook := createTestWebhookStatus()
		webhook.ID = fmt.Sprintf("webhook-%d", i)
		webhook.Organization = fmt.Sprintf("org-%d", i%10)
		monitor.webhooks[webhook.ID] = webhook
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		monitor.updateMetrics()
	}
}

func BenchmarkWebhookMonitor_GetAllWebhookStatuses(b *testing.B) {
	monitor, _ := createTestWebhookMonitor()

	// Add many test webhooks
	for i := 0; i < 1000; i++ {
		webhook := createTestWebhookStatus()
		webhook.ID = fmt.Sprintf("webhook-%d", i)
		monitor.webhooks[webhook.ID] = webhook
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = monitor.GetAllWebhookStatuses()
	}
}
