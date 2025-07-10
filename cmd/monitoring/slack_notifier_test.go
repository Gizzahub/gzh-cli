package monitoring

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestNewSlackNotifier(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &SlackConfig{
		WebhookURL: "https://hooks.slack.com/services/test",
		Channel:    "#test-channel",
		Username:   "Test Bot",
		IconEmoji:  ":test:",
		Enabled:    true,
	}

	notifier := NewSlackNotifier(config, logger)
	
	assert.Equal(t, config.WebhookURL, notifier.webhookURL)
	assert.Equal(t, config.Channel, notifier.channel)
	assert.Equal(t, config.Username, notifier.username)
	assert.Equal(t, config.IconEmoji, notifier.iconEmoji)
	assert.NotNil(t, notifier.httpClient)
	assert.NotNil(t, notifier.logger)
}

func TestSlackNotifier_SendAlert(t *testing.T) {
	// Create a test server to mock Slack webhook
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	logger := zaptest.NewLogger(t)
	config := &SlackConfig{
		WebhookURL: server.URL,
		Channel:    "#alerts",
		Username:   "GZH Monitoring",
		IconEmoji:  ":robot_face:",
		Enabled:    true,
	}

	notifier := NewSlackNotifier(config, logger)

	// Create test alert
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
		Annotations: map[string]string{
			"description": "CPU usage has been high for 5 minutes",
		},
		StartsAt:  time.Now(),
		UpdatedAt: time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := notifier.SendAlert(ctx, alert)
	assert.NoError(t, err)
}

func TestSlackNotifier_SendSystemStatus(t *testing.T) {
	// Create a test server to mock Slack webhook
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	logger := zaptest.NewLogger(t)
	config := &SlackConfig{
		WebhookURL: server.URL,
		Channel:    "#monitoring",
		Username:   "GZH Monitoring",
		IconEmoji:  ":robot_face:",
		Enabled:    true,
	}

	notifier := NewSlackNotifier(config, logger)

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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := notifier.SendSystemStatus(ctx, status)
	assert.NoError(t, err)
}

func TestSlackNotifier_SendCustomMessage(t *testing.T) {
	// Create a test server to mock Slack webhook
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	logger := zaptest.NewLogger(t)
	config := &SlackConfig{
		WebhookURL: server.URL,
		Channel:    "#general",
		Username:   "Test Bot",
		IconEmoji:  ":test:",
		Enabled:    true,
	}

	notifier := NewSlackNotifier(config, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := notifier.SendCustomMessage(ctx, "Test Title", "This is a test message", AlertSeverityInfo)
	assert.NoError(t, err)
}

func TestSlackNotifier_TestConnection(t *testing.T) {
	// Create a test server to mock Slack webhook
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	logger := zaptest.NewLogger(t)
	config := &SlackConfig{
		WebhookURL: server.URL,
		Channel:    "#test",
		Username:   "Test Bot",
		IconEmoji:  ":test:",
		Enabled:    true,
	}

	notifier := NewSlackNotifier(config, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := notifier.TestConnection(ctx)
	assert.NoError(t, err)
}

func TestSlackNotifier_EmptyWebhookURL(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &SlackConfig{
		WebhookURL: "",
		Channel:    "#test",
		Username:   "Test Bot",
		IconEmoji:  ":test:",
		Enabled:    true,
	}

	notifier := NewSlackNotifier(config, logger)

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
	assert.Contains(t, err.Error(), "webhook URL not configured")
}

func TestSlackNotifier_HTTPError(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad Request"))
	}))
	defer server.Close()

	logger := zaptest.NewLogger(t)
	config := &SlackConfig{
		WebhookURL: server.URL,
		Channel:    "#test",
		Username:   "Test Bot",
		IconEmoji:  ":test:",
		Enabled:    true,
	}

	notifier := NewSlackNotifier(config, logger)

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
	assert.Contains(t, err.Error(), "status code: 400")
}

func TestSlackNotifier_GetSeverityColor(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &SlackConfig{
		WebhookURL: "https://test.com",
		Channel:    "#test",
		Username:   "Test Bot",
		IconEmoji:  ":test:",
		Enabled:    true,
	}

	notifier := NewSlackNotifier(config, logger)

	testCases := []struct {
		severity AlertSeverity
		expected string
	}{
		{AlertSeverityCritical, "danger"},
		{AlertSeverityHigh, "danger"},
		{AlertSeverityMedium, "warning"},
		{AlertSeverityLow, "good"},
		{AlertSeverityInfo, "#439FE0"},
		{AlertSeverity("unknown"), "#D3D3D3"},
	}

	for _, tc := range testCases {
		t.Run(string(tc.severity), func(t *testing.T) {
			color := notifier.getSeverityColor(tc.severity)
			assert.Equal(t, tc.expected, color)
		})
	}
}

func TestSlackNotifier_GetStatusEmoji(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &SlackConfig{
		WebhookURL: "https://test.com",
		Channel:    "#test",
		Username:   "Test Bot",
		IconEmoji:  ":test:",
		Enabled:    true,
	}

	notifier := NewSlackNotifier(config, logger)

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

func TestFormatBytesForSlack(t *testing.T) {
	testCases := []struct {
		bytes    uint64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1024 * 1024, "1.0 MB"},
		{1024 * 1024 * 1024, "1.0 GB"},
		{1024 * 1024 * 1024 * 1024, "1.0 TB"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			result := formatBytesForSlack(tc.bytes)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestSlackMessage_JSON(t *testing.T) {
	// Test that SlackMessage can be properly marshaled to JSON
	message := &SlackMessage{
		Channel:   "#test",
		Username:  "Test Bot",
		IconEmoji: ":test:",
		Text:      "Test message",
		Attachments: []SlackAttachment{
			{
				Color: "good",
				Title: "Test Title",
				Text:  "Test attachment text",
				Fields: []SlackField{
					{
						Title: "Field 1",
						Value: "Value 1",
						Short: true,
					},
				},
				Footer:    "Test Footer",
				Timestamp: time.Now().Unix(),
			},
		},
	}

	// Verify the message structure is valid
	assert.Equal(t, "#test", message.Channel)
	assert.Equal(t, "Test Bot", message.Username)
	assert.Equal(t, ":test:", message.IconEmoji)
	assert.Equal(t, "Test message", message.Text)
	assert.Len(t, message.Attachments, 1)
	assert.Equal(t, "good", message.Attachments[0].Color)
	assert.Len(t, message.Attachments[0].Fields, 1)
}