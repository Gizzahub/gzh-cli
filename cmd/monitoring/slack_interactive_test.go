package monitoring

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestSlackNotifier_SendInteractiveAlert(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &SlackConfig{
		WebhookURL: "https://hooks.slack.com/services/test",
		Channel:    "#test-channel",
		Username:   "Test Bot",
		IconEmoji:  ":test:",
		Enabled:    true,
	}

	notifier := NewSlackNotifier(config, logger)

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

	// This would fail in test environment without actual webhook, but we can test the message formatting
	message := notifier.formatInteractiveAlertMessage(alert)

	assert.NotNil(t, message)
	assert.Equal(t, "#test-channel", message.Channel)
	assert.Equal(t, "Test Bot", message.Username)
	assert.Len(t, message.Attachments, 1)

	attachment := message.Attachments[0]
	assert.NotEmpty(t, attachment.Title)
	assert.NotEmpty(t, attachment.Text)
	assert.NotEmpty(t, attachment.Actions)
	assert.Equal(t, "alert_test-alert-1", attachment.CallbackID)

	// Verify actions for firing alert
	assert.Len(t, attachment.Actions, 3) // Silence, Resolve, View Details
	assert.Equal(t, "silence", attachment.Actions[0].Name)
	assert.Equal(t, "resolve", attachment.Actions[1].Name)
	assert.Equal(t, "details", attachment.Actions[2].Name)
}

func TestSlackNotifier_SendInteractiveSystemStatus(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &SlackConfig{
		WebhookURL: "https://hooks.slack.com/services/test",
		Channel:    "#test-channel",
		Username:   "Test Bot",
		IconEmoji:  ":test:",
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

	message := notifier.formatInteractiveSystemStatusMessage(status)

	assert.NotNil(t, message)
	assert.Equal(t, "#test-channel", message.Channel)
	assert.Len(t, message.Attachments, 1)

	attachment := message.Attachments[0]
	assert.Contains(t, attachment.Title, "System Status")
	assert.NotEmpty(t, attachment.Actions)
	assert.Equal(t, "system_status", attachment.CallbackID)

	// Verify system status actions
	assert.Len(t, attachment.Actions, 3) // Dashboard, Refresh, Metrics
	assert.Equal(t, "dashboard", attachment.Actions[0].Name)
	assert.Equal(t, "refresh", attachment.Actions[1].Name)
	assert.Equal(t, "metrics", attachment.Actions[2].Name)
}

func TestSlackNotifier_ProcessInteraction(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &SlackConfig{
		WebhookURL: "https://hooks.slack.com/services/test",
		Channel:    "#test-channel",
		Username:   "Test Bot",
		IconEmoji:  ":test:",
		Enabled:    true,
	}

	notifier := NewSlackNotifier(config, logger)

	testCases := []struct {
		name          string
		action        SlackActionResponse
		expectedText  string
		expectedError bool
	}{
		{
			name: "Silence action",
			action: SlackActionResponse{
				Name:  "silence",
				Type:  "button",
				Value: "alert-123",
			},
			expectedText:  "silenced",
			expectedError: false,
		},
		{
			name: "Resolve action",
			action: SlackActionResponse{
				Name:  "resolve",
				Type:  "button",
				Value: "alert-123",
			},
			expectedText:  "resolved",
			expectedError: false,
		},
		{
			name: "Unsilence action",
			action: SlackActionResponse{
				Name:  "unsilence",
				Type:  "button",
				Value: "alert-123",
			},
			expectedText:  "unsilenced",
			expectedError: false,
		},
		{
			name: "Refresh action",
			action: SlackActionResponse{
				Name:  "refresh",
				Type:  "button",
				Value: "refresh_status",
			},
			expectedText:  "refresh",
			expectedError: false,
		},
		{
			name: "Unknown action",
			action: SlackActionResponse{
				Name:  "unknown",
				Type:  "button",
				Value: "test",
			},
			expectedText:  "",
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			payload := &SlackInteractionPayload{
				Type:       "interactive_message",
				Actions:    []SlackActionResponse{tc.action},
				CallbackID: "test_callback",
				User: SlackUser{
					ID:   "U123456",
					Name: "testuser",
				},
				Channel: SlackChannel{
					ID:   "C123456",
					Name: "test-channel",
				},
				Team: SlackTeam{
					ID:     "T123456",
					Domain: "test-team",
				},
			}

			response, err := notifier.ProcessInteraction(payload)

			if tc.expectedError {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Contains(t, response.Text, tc.expectedText)
				assert.Len(t, response.Attachments, 1)
			}
		})
	}
}

func TestSlackNotifier_InteractiveAlertStates(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &SlackConfig{
		WebhookURL: "https://hooks.slack.com/services/test",
		Channel:    "#test-channel",
		Username:   "Test Bot",
		IconEmoji:  ":test:",
		Enabled:    true,
	}

	notifier := NewSlackNotifier(config, logger)

	testCases := []struct {
		status          AlertStatus
		expectedActions int
		actionNames     []string
	}{
		{
			status:          AlertStatusFiring,
			expectedActions: 3,
			actionNames:     []string{"silence", "resolve", "details"},
		},
		{
			status:          AlertStatusSilenced,
			expectedActions: 2,
			actionNames:     []string{"unsilence", "resolve"},
		},
		{
			status:          AlertStatusResolved,
			expectedActions: 1,
			actionNames:     []string{"details"},
		},
	}

	for _, tc := range testCases {
		t.Run(string(tc.status), func(t *testing.T) {
			alert := &AlertInstance{
				ID:       "test-alert",
				RuleName: "Test Alert",
				Status:   tc.status,
				Severity: AlertSeverityHigh,
				Message:  "Test message",
				Labels:   map[string]string{"test": "value"},
			}

			message := notifier.formatInteractiveAlertMessage(alert)
			attachment := message.Attachments[0]

			assert.Len(t, attachment.Actions, tc.expectedActions)

			for i, expectedName := range tc.actionNames {
				assert.Equal(t, expectedName, attachment.Actions[i].Name)
			}
		})
	}
}

func TestSlackAction_Structure(t *testing.T) {
	action := SlackAction{
		Type:  "button",
		Text:  "Test Button",
		Name:  "test_action",
		Value: "test_value",
		Style: "primary",
		Url:   "https://example.com",
	}

	assert.Equal(t, "button", action.Type)
	assert.Equal(t, "Test Button", action.Text)
	assert.Equal(t, "test_action", action.Name)
	assert.Equal(t, "test_value", action.Value)
	assert.Equal(t, "primary", action.Style)
	assert.Equal(t, "https://example.com", action.Url)
}

func TestSlackInteractionPayload_Structure(t *testing.T) {
	payload := SlackInteractionPayload{
		Type:       "interactive_message",
		CallbackID: "test_callback",
		Actions: []SlackActionResponse{
			{
				Name:  "test_action",
				Type:  "button",
				Value: "test_value",
			},
		},
		User: SlackUser{
			ID:   "U123456",
			Name: "testuser",
		},
		Channel: SlackChannel{
			ID:   "C123456",
			Name: "test-channel",
		},
		Team: SlackTeam{
			ID:     "T123456",
			Domain: "test-team",
		},
		ActionTS:  "1234567890.123456",
		MessageTS: "1234567890.123456",
		Token:     "test_token",
	}

	assert.Equal(t, "interactive_message", payload.Type)
	assert.Equal(t, "test_callback", payload.CallbackID)
	assert.Len(t, payload.Actions, 1)
	assert.Equal(t, "test_action", payload.Actions[0].Name)
	assert.Equal(t, "U123456", payload.User.ID)
	assert.Equal(t, "C123456", payload.Channel.ID)
	assert.Equal(t, "T123456", payload.Team.ID)
}
