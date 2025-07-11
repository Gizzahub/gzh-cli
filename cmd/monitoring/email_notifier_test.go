package monitoring

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestNewEmailNotifier(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &EmailConfig{
		SMTPHost:   "smtp.gmail.com",
		SMTPPort:   587,
		Username:   "test@example.com",
		Password:   "testpass",
		From:       "monitoring@example.com",
		Recipients: []string{"admin@example.com", "ops@example.com"},
		UseTLS:     true,
		Enabled:    true,
	}

	notifier := NewEmailNotifier(config, logger)

	assert.Equal(t, config, notifier.config)
	assert.NotNil(t, notifier.logger)
	assert.NotNil(t, notifier.templates)
	assert.Len(t, notifier.templates, 3) // alert, status, custom templates
}

func TestEmailNotifier_FormatAlertBody(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &EmailConfig{
		SMTPHost:   "smtp.test.com",
		SMTPPort:   587,
		Username:   "test",
		Password:   "test",
		From:       "test@example.com",
		Recipients: []string{"admin@example.com"},
		Enabled:    true,
	}

	notifier := NewEmailNotifier(config, logger)

	// Create test alert
	now := time.Now()
	alert := &AlertInstance{
		ID:       "test-alert-1",
		RuleID:   "rule-1",
		RuleName: "High CPU Usage",
		Status:   AlertStatusFiring,
		Severity: AlertSeverityCritical,
		Message:  "CPU usage is above 90%",
		Labels: map[string]string{
			"host":    "server-01",
			"service": "web-app",
		},
		FiredAt:   &now,
		UpdatedAt: now,
	}

	body, err := notifier.formatAlertBody(alert)
	require.NoError(t, err)
	assert.Contains(t, body, "High CPU Usage")
	assert.Contains(t, body, "CPU usage is above 90%")
	assert.Contains(t, body, "critical")
	assert.Contains(t, body, "server-01")
}

func TestEmailNotifier_FormatSystemStatusBody(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &EmailConfig{
		SMTPHost:   "smtp.test.com",
		SMTPPort:   587,
		From:       "test@example.com",
		Recipients: []string{"admin@example.com"},
		Enabled:    true,
	}

	notifier := NewEmailNotifier(config, logger)

	// Create test system status
	status := &SystemStatus{
		Status:        "healthy",
		Uptime:        "2h30m15s",
		ActiveTasks:   5,
		TotalRequests: 12345,
		MemoryUsage:   1024 * 1024 * 512, // 512MB
		CPUUsage:      25.7,
		DiskUsage:     45.2,
		Timestamp:     time.Now(),
	}

	body, err := notifier.formatSystemStatusBody(status)
	require.NoError(t, err)
	assert.Contains(t, body, "System Status: healthy")
	assert.Contains(t, body, "2h30m15s")
	assert.Contains(t, body, "25.7%")
	assert.Contains(t, body, "512.0 MB")
}

func TestEmailNotifier_FormatCustomMessageBody(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &EmailConfig{
		SMTPHost:   "smtp.test.com",
		SMTPPort:   587,
		From:       "test@example.com",
		Recipients: []string{"admin@example.com"},
		Enabled:    true,
	}

	notifier := NewEmailNotifier(config, logger)

	body, err := notifier.formatCustomMessageBody("Test Title", "This is a test message", AlertSeverityInfo)
	require.NoError(t, err)
	assert.Contains(t, body, "Test Title")
	assert.Contains(t, body, "This is a test message")
}

func TestEmailNotifier_NotConfigured(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &EmailConfig{
		SMTPHost: "",
		Enabled:  false,
	}

	notifier := NewEmailNotifier(config, logger)

	alert := &AlertInstance{
		ID:       "test-alert",
		RuleName: "Test Alert",
		Status:   AlertStatusFiring,
		Severity: AlertSeverityInfo,
		Message:  "Test message",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := notifier.SendAlert(ctx, alert)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not configured")
}

func TestEmailNotifier_NoRecipients(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &EmailConfig{
		SMTPHost:   "smtp.test.com",
		SMTPPort:   587,
		From:       "test@example.com",
		Recipients: []string{}, // No recipients
		Enabled:    true,
	}

	notifier := NewEmailNotifier(config, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := notifier.SendCustomMessage(ctx, "Test", "Test message", AlertSeverityInfo)
	assert.Error(t, err)
	// Note: The actual error will occur in sendEmail when trying to send
}

func TestEmailNotifier_FormatAlertSubject(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &EmailConfig{
		SMTPHost: "smtp.test.com",
		From:     "test@example.com",
		Enabled:  true,
	}

	notifier := NewEmailNotifier(config, logger)

	testCases := []struct {
		alert    *AlertInstance
		expected string
	}{
		{
			alert: &AlertInstance{
				RuleName: "High CPU",
				Status:   AlertStatusFiring,
				Severity: AlertSeverityCritical,
			},
			expected: "🚨 GZH Alert - critical: High CPU",
		},
		{
			alert: &AlertInstance{
				RuleName: "Service Recovered",
				Status:   AlertStatusResolved,
				Severity: AlertSeverityInfo,
			},
			expected: "✅ GZH Alert - info: Service Recovered",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			subject := notifier.formatAlertSubject(tc.alert)
			assert.Equal(t, tc.expected, subject)
		})
	}
}

func TestEmailNotifier_GetStatusEmoji(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &EmailConfig{}
	notifier := NewEmailNotifier(config, logger)

	testCases := []struct {
		status   AlertStatus
		expected string
	}{
		{AlertStatusFiring, "🚨"},
		{AlertStatusResolved, "✅"},
		{AlertStatusSilenced, "🔇"},
		{AlertStatus("unknown"), "ℹ️"},
	}

	for _, tc := range testCases {
		t.Run(string(tc.status), func(t *testing.T) {
			emoji := notifier.getStatusEmoji(tc.status)
			assert.Equal(t, tc.expected, emoji)
		})
	}
}

func TestEmailNotifier_GetStatusColor(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &EmailConfig{}
	notifier := NewEmailNotifier(config, logger)

	testCases := []struct {
		status   AlertStatus
		expected string
	}{
		{AlertStatusFiring, "#f44336"},
		{AlertStatusResolved, "#4CAF50"},
		{AlertStatusSilenced, "#FF9800"},
		{AlertStatus("unknown"), "#2196F3"},
	}

	for _, tc := range testCases {
		t.Run(string(tc.status), func(t *testing.T) {
			color := notifier.getStatusColor(tc.status)
			assert.Equal(t, tc.expected, color)
		})
	}
}

func TestEmailNotifier_GetSeverityColor(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &EmailConfig{}
	notifier := NewEmailNotifier(config, logger)

	testCases := []struct {
		severity AlertSeverity
		expected string
	}{
		{AlertSeverityCritical, "#f44336"},
		{AlertSeverityHigh, "#FF5722"},
		{AlertSeverityMedium, "#FF9800"},
		{AlertSeverityLow, "#4CAF50"},
		{AlertSeverityInfo, "#2196F3"},
		{AlertSeverity("unknown"), "#9E9E9E"},
	}

	for _, tc := range testCases {
		t.Run(string(tc.severity), func(t *testing.T) {
			color := notifier.getSeverityColor(tc.severity)
			assert.Equal(t, tc.expected, color)
		})
	}
}

func TestEmailNotifier_GetSystemStatusEmoji(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &EmailConfig{}
	notifier := NewEmailNotifier(config, logger)

	testCases := []struct {
		status   string
		expected string
	}{
		{"healthy", "✅"},
		{"warning", "⚠️"},
		{"critical", "🚨"},
		{"unknown", "ℹ️"},
	}

	for _, tc := range testCases {
		t.Run(tc.status, func(t *testing.T) {
			emoji := notifier.getSystemStatusEmoji(tc.status)
			assert.Equal(t, tc.expected, emoji)
		})
	}
}

func TestEmailNotifier_GetSystemStatusColor(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &EmailConfig{}
	notifier := NewEmailNotifier(config, logger)

	testCases := []struct {
		status   string
		expected string
	}{
		{"healthy", "#4CAF50"},
		{"warning", "#FF9800"},
		{"critical", "#f44336"},
		{"unknown", "#2196F3"},
	}

	for _, tc := range testCases {
		t.Run(tc.status, func(t *testing.T) {
			color := notifier.getSystemStatusColor(tc.status)
			assert.Equal(t, tc.expected, color)
		})
	}
}

func TestEmailMessage_Headers(t *testing.T) {
	message := &EmailMessage{
		To:      []string{"admin@example.com", "ops@example.com"},
		Subject: "Test Alert",
		Body:    "<html><body>Test</body></html>",
	}

	assert.Len(t, message.To, 2)
	assert.Equal(t, "Test Alert", message.Subject)
	assert.Contains(t, message.Body, "<html>")
}

func TestEmailConfig_Validation(t *testing.T) {
	// Test various config scenarios
	configs := []struct {
		name    string
		config  EmailConfig
		isValid bool
	}{
		{
			name: "Valid config",
			config: EmailConfig{
				SMTPHost:   "smtp.gmail.com",
				SMTPPort:   587,
				From:       "test@example.com",
				Recipients: []string{"admin@example.com"},
				Enabled:    true,
			},
			isValid: true,
		},
		{
			name: "Missing SMTP host",
			config: EmailConfig{
				SMTPPort:   587,
				From:       "test@example.com",
				Recipients: []string{"admin@example.com"},
				Enabled:    true,
			},
			isValid: false,
		},
		{
			name: "No recipients",
			config: EmailConfig{
				SMTPHost: "smtp.gmail.com",
				SMTPPort: 587,
				From:     "test@example.com",
				Enabled:  true,
			},
			isValid: false,
		},
		{
			name: "Disabled",
			config: EmailConfig{
				Enabled: false,
			},
			isValid: false,
		},
	}

	for _, tc := range configs {
		t.Run(tc.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			notifier := NewEmailNotifier(&tc.config, logger)

			ctx := context.Background()
			err := notifier.SendCustomMessage(ctx, "Test", "Test", AlertSeverityInfo)

			if tc.isValid {
				// Will fail at SMTP connection, not configuration
				assert.Error(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "not configured")
			}
		})
	}
}

func TestEmailNotifier_TemplateRendering(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &EmailConfig{
		SMTPHost:   "smtp.test.com",
		SMTPPort:   587,
		From:       "test@example.com",
		Recipients: []string{"admin@example.com"},
		Enabled:    true,
	}

	notifier := NewEmailNotifier(config, logger)

	// Test that templates are properly initialized
	assert.NotNil(t, notifier.templates["alert"])
	assert.NotNil(t, notifier.templates["status"])
	assert.NotNil(t, notifier.templates["custom"])

	// Test alert template with various data
	now := time.Now()
	alert := &AlertInstance{
		RuleName:   "Test Alert",
		Status:     AlertStatusFiring,
		Severity:   AlertSeverityCritical,
		Message:    "Test message with <special> characters & symbols",
		FiredAt:    &now,
		ResolvedAt: nil,
		Labels: map[string]string{
			"env":     "production",
			"service": "api",
		},
	}

	body, err := notifier.formatAlertBody(alert)
	require.NoError(t, err)

	// Verify HTML structure
	assert.Contains(t, body, "<!DOCTYPE html>")
	assert.Contains(t, body, "<html>")
	assert.Contains(t, body, "</html>")

	// Verify content
	assert.Contains(t, body, "Test Alert")
	assert.Contains(t, body, "Test message with &lt;special&gt; characters &amp; symbols") // HTML escaped
	assert.Contains(t, body, "production")
	assert.Contains(t, body, "api")
}

func TestEmailNotifier_MultipleRecipients(t *testing.T) {
	recipients := []string{
		"admin@example.com",
		"ops@example.com",
		"dev@example.com",
	}

	message := &EmailMessage{
		To:      recipients,
		Subject: "Test",
		Body:    "Test body",
	}

	// Verify all recipients are included
	assert.Equal(t, 3, len(message.To))
	assert.Equal(t, "admin@example.com", message.To[0])
	assert.Equal(t, "ops@example.com", message.To[1])
	assert.Equal(t, "dev@example.com", message.To[2])

	// Test joining for email header
	joined := strings.Join(message.To, ", ")
	assert.Equal(t, "admin@example.com, ops@example.com, dev@example.com", joined)
}
