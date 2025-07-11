package monitoring

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestNewTeamsNotifier(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &TeamsConfig{
		WebhookURL: "https://test.webhook.office.com/webhookb2/test",
		Enabled:    true,
	}

	notifier := NewTeamsNotifier(config, logger)

	assert.NotNil(t, notifier)
	assert.Equal(t, config.WebhookURL, notifier.webhookURL)
	assert.NotNil(t, notifier.httpClient)
	assert.Equal(t, logger, notifier.logger)
}

func TestTeamsNotifier_FormatAlertMessage(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &TeamsConfig{
		WebhookURL: "https://test.webhook.office.com/webhookb2/test",
		Enabled:    true,
	}

	notifier := NewTeamsNotifier(config, logger)

	testCases := []struct {
		name     string
		alert    *AlertInstance
		validate func(t *testing.T, message *TeamsMessage)
	}{
		{
			name: "Firing alert",
			alert: &AlertInstance{
				ID:       "alert-123",
				RuleName: "High CPU Usage",
				Severity: "critical",
				Status:   AlertStatusFiring,
				Message:  "CPU usage is above 90%",
				Labels: map[string]string{
					"service":  "web-server",
					"instance": "web-01",
				},
				FiredAt: &[]time.Time{time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)}[0],
			},
			validate: func(t *testing.T, message *TeamsMessage) {
				assert.Equal(t, "message", message.Type)
				assert.Equal(t, "FF0000", message.ThemeColor) // Red for critical
				assert.Equal(t, "Alert: High CPU Usage", message.Summary)
				assert.Len(t, message.Attachments, 1)

				card := message.Attachments[0]
				assert.Equal(t, "application/vnd.microsoft.card.adaptive", card.ContentType)
				assert.Equal(t, "AdaptiveCard", card.Content.Type)
				assert.Equal(t, "1.2", card.Content.Version)
				assert.Contains(t, card.Content.Schema, "adaptivecards.io")

				// Check body elements
				assert.GreaterOrEqual(t, len(card.Content.Body), 3)

				// Title should contain emoji and rule name
				titleBlock := card.Content.Body[0]
				assert.Equal(t, "TextBlock", titleBlock.Type)
				assert.Contains(t, titleBlock.Text, "🚨")
				assert.Contains(t, titleBlock.Text, "High CPU Usage")
				assert.Equal(t, "Attention", titleBlock.Color)

				// Should have FactSet with alert details
				hasFactSet := false
				for _, body := range card.Content.Body {
					if body.Type == "FactSet" {
						hasFactSet = true
						assert.NotEmpty(t, body.Facts)
						// Check for required facts
						factTitles := make(map[string]bool)
						for _, fact := range body.Facts {
							factTitles[fact.Title] = true
						}
						assert.True(t, factTitles["Severity"])
						assert.True(t, factTitles["Status"])
						assert.True(t, factTitles["Fired At"])
						assert.True(t, factTitles["Service"])
						assert.True(t, factTitles["Instance"])
						break
					}
				}
				assert.True(t, hasFactSet)

				// Check actions
				assert.Len(t, card.Content.Actions, 2)
				assert.Equal(t, "Action.OpenUrl", card.Content.Actions[0].Type)
				assert.Equal(t, "View Details", card.Content.Actions[0].Title)
				assert.Contains(t, card.Content.Actions[0].URL, "alert-123")
			},
		},
		{
			name: "Resolved alert",
			alert: &AlertInstance{
				ID:       "alert-456",
				RuleName: "Database Connection",
				Severity: "high",
				Status:   AlertStatusResolved,
				Message:  "Database connection restored",
				Labels: map[string]string{
					"database": "primary",
				},
				FiredAt:    &[]time.Time{time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)}[0],
				ResolvedAt: &[]time.Time{time.Date(2024, 1, 1, 12, 30, 0, 0, time.UTC)}[0],
			},
			validate: func(t *testing.T, message *TeamsMessage) {
				assert.Equal(t, "FF6600", message.ThemeColor) // Orange for high
				assert.Equal(t, "Alert: Database Connection", message.Summary)

				card := message.Attachments[0]

				// Title should contain resolved emoji
				titleBlock := card.Content.Body[0]
				assert.Contains(t, titleBlock.Text, "✅")
				assert.Equal(t, "Good", titleBlock.Color)

				// Should have Resolved At fact
				hasResolvedAt := false
				for _, body := range card.Content.Body {
					if body.Type == "FactSet" {
						for _, fact := range body.Facts {
							if fact.Title == "Resolved At" {
								hasResolvedAt = true
								assert.Contains(t, fact.Value, "2024-01-01")
								break
							}
						}
						break
					}
				}
				assert.True(t, hasResolvedAt)
			},
		},
		{
			name: "Silenced alert",
			alert: &AlertInstance{
				ID:       "alert-789",
				RuleName: "Memory Usage",
				Severity: "medium",
				Status:   AlertStatusSilenced,
				Message:  "Memory usage is high",
				Labels:   map[string]string{},
			},
			validate: func(t *testing.T, message *TeamsMessage) {
				assert.Equal(t, "FFFF00", message.ThemeColor) // Yellow for medium

				card := message.Attachments[0]
				titleBlock := card.Content.Body[0]
				assert.Contains(t, titleBlock.Text, "🔇")
				assert.Equal(t, "Warning", titleBlock.Color)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			message := notifier.formatAlertMessage(tc.alert)
			tc.validate(t, message)
		})
	}
}

func TestTeamsNotifier_FormatSystemStatusMessage(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &TeamsConfig{
		WebhookURL: "https://test.webhook.office.com/webhookb2/test",
		Enabled:    true,
	}

	notifier := NewTeamsNotifier(config, logger)

	testCases := []struct {
		name     string
		status   *SystemStatus
		validate func(t *testing.T, message *TeamsMessage)
	}{
		{
			name: "Healthy system status",
			status: &SystemStatus{
				Status:        "healthy",
				Uptime:        "24h 30m",
				ActiveTasks:   5,
				MemoryUsage:   1024 * 1024 * 512, // 512MB
				CPUUsage:      25.5,
				DiskUsage:     60.2,
				TotalRequests: 10000,
			},
			validate: func(t *testing.T, message *TeamsMessage) {
				assert.Equal(t, "message", message.Type)
				assert.Equal(t, "00FF00", message.ThemeColor) // Green for healthy
				assert.Equal(t, "System Status Update", message.Summary)

				card := message.Attachments[0]
				titleBlock := card.Content.Body[0]
				assert.Contains(t, titleBlock.Text, "✅")
				assert.Contains(t, titleBlock.Text, "HEALTHY")
				assert.Equal(t, "Good", titleBlock.Color)

				// Check FactSet
				hasFactSet := false
				for _, body := range card.Content.Body {
					if body.Type == "FactSet" {
						hasFactSet = true
						assert.NotEmpty(t, body.Facts)

						factValues := make(map[string]string)
						for _, fact := range body.Facts {
							factValues[fact.Title] = fact.Value
						}

						assert.Equal(t, "24h 30m", factValues["Uptime"])
						assert.Equal(t, "5", factValues["Active Tasks"])
						assert.Contains(t, factValues["Memory Usage"], "MB")
						assert.Equal(t, "25.5%", factValues["CPU Usage"])
						assert.Equal(t, "60.2%", factValues["Disk Usage"])
						assert.Equal(t, "10000", factValues["Total Requests"])
						break
					}
				}
				assert.True(t, hasFactSet)

				// Check actions
				assert.Len(t, card.Content.Actions, 2)
				assert.Equal(t, "Full Dashboard", card.Content.Actions[0].Title)
				assert.Equal(t, "View Metrics", card.Content.Actions[1].Title)
			},
		},
		{
			name: "Warning system status",
			status: &SystemStatus{
				Status:        "warning",
				Uptime:        "1h 15m",
				ActiveTasks:   15,
				MemoryUsage:   1024 * 1024 * 1024 * 2, // 2GB
				CPUUsage:      85.3,
				DiskUsage:     0, // No disk usage
				TotalRequests: 50000,
			},
			validate: func(t *testing.T, message *TeamsMessage) {
				assert.Equal(t, "FFFF00", message.ThemeColor) // Yellow for warning

				card := message.Attachments[0]
				titleBlock := card.Content.Body[0]
				assert.Contains(t, titleBlock.Text, "⚠️")
				assert.Contains(t, titleBlock.Text, "WARNING")
				assert.Equal(t, "Warning", titleBlock.Color)

				// Should not have Disk Usage fact when it's 0
				hasDiskUsage := false
				for _, body := range card.Content.Body {
					if body.Type == "FactSet" {
						for _, fact := range body.Facts {
							if fact.Title == "Disk Usage" {
								hasDiskUsage = true
								break
							}
						}
						break
					}
				}
				assert.False(t, hasDiskUsage)
			},
		},
		{
			name: "Critical system status",
			status: &SystemStatus{
				Status:        "critical",
				Uptime:        "5m",
				ActiveTasks:   50,
				MemoryUsage:   1024 * 1024 * 1024 * 8, // 8GB
				CPUUsage:      98.7,
				DiskUsage:     95.0,
				TotalRequests: 1000000,
			},
			validate: func(t *testing.T, message *TeamsMessage) {
				assert.Equal(t, "FF0000", message.ThemeColor) // Red for critical

				card := message.Attachments[0]
				titleBlock := card.Content.Body[0]
				assert.Contains(t, titleBlock.Text, "🚨")
				assert.Contains(t, titleBlock.Text, "CRITICAL")
				assert.Equal(t, "Attention", titleBlock.Color)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			message := notifier.formatSystemStatusMessage(tc.status)
			tc.validate(t, message)
		})
	}
}

func TestTeamsNotifier_FormatCustomMessage(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &TeamsConfig{
		WebhookURL: "https://test.webhook.office.com/webhookb2/test",
		Enabled:    true,
	}

	notifier := NewTeamsNotifier(config, logger)

	testCases := []struct {
		name       string
		title      string
		text       string
		severity   AlertSeverity
		themeColor string
	}{
		{
			name:       "Info message",
			title:      "System Update",
			text:       "System will be updated tonight",
			severity:   AlertSeverityInfo,
			themeColor: "0078D4", // Teams blue
		},
		{
			name:       "Critical message",
			title:      "Emergency Maintenance",
			text:       "Emergency maintenance required",
			severity:   AlertSeverityCritical,
			themeColor: "FF0000", // Red
		},
		{
			name:       "Low severity message",
			title:      "Scheduled Task",
			text:       "Backup completed successfully",
			severity:   AlertSeverityLow,
			themeColor: "00FF00", // Green
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			message := notifier.formatCustomMessage(tc.title, tc.text, tc.severity)

			assert.Equal(t, "message", message.Type)
			assert.Equal(t, tc.themeColor, message.ThemeColor)
			assert.Equal(t, tc.title, message.Summary)
			assert.Len(t, message.Attachments, 1)

			card := message.Attachments[0]
			assert.Equal(t, "application/vnd.microsoft.card.adaptive", card.ContentType)

			// Check title and text in body
			assert.GreaterOrEqual(t, len(card.Content.Body), 2)

			titleBlock := card.Content.Body[0]
			assert.Equal(t, "TextBlock", titleBlock.Type)
			assert.Equal(t, tc.title, titleBlock.Text)
			assert.Equal(t, "Bolder", titleBlock.Weight)

			textBlock := card.Content.Body[1]
			assert.Equal(t, "TextBlock", textBlock.Type)
			assert.Equal(t, tc.text, textBlock.Text)

			// Check action
			assert.Len(t, card.Content.Actions, 1)
			assert.Equal(t, "Dashboard", card.Content.Actions[0].Title)
		})
	}
}

func TestTeamsNotifier_FormatTestMessage(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &TeamsConfig{
		WebhookURL: "https://test.webhook.office.com/webhookb2/test",
		Enabled:    true,
	}

	notifier := NewTeamsNotifier(config, logger)
	message := notifier.formatTestMessage()

	assert.Equal(t, "message", message.Type)
	assert.Equal(t, "0078D4", message.ThemeColor) // Teams blue
	assert.Equal(t, "GZH Monitoring Test", message.Summary)
	assert.Len(t, message.Attachments, 1)

	card := message.Attachments[0]
	assert.Equal(t, "application/vnd.microsoft.card.adaptive", card.ContentType)

	// Check title
	titleBlock := card.Content.Body[0]
	assert.Contains(t, titleBlock.Text, "🧪")
	assert.Contains(t, titleBlock.Text, "Test Message")
	assert.Equal(t, "Good", titleBlock.Color)

	// Check FactSet
	hasFactSet := false
	for _, body := range card.Content.Body {
		if body.Type == "FactSet" {
			hasFactSet = true
			assert.Len(t, body.Facts, 2)
			assert.Equal(t, "Test Time", body.Facts[0].Title)
			assert.Equal(t, "Status", body.Facts[1].Title)
			assert.Equal(t, "✅ Success", body.Facts[1].Value)
			break
		}
	}
	assert.True(t, hasFactSet)

	// Check action
	assert.Len(t, card.Content.Actions, 1)
	assert.Equal(t, "View Dashboard", card.Content.Actions[0].Title)
}

func TestTeamsNotifier_HelperMethods(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &TeamsConfig{
		WebhookURL: "https://test.webhook.office.com/webhookb2/test",
		Enabled:    true,
	}

	notifier := NewTeamsNotifier(config, logger)

	// Test severity colors
	assert.Equal(t, "FF0000", notifier.getSeverityColor(AlertSeverityCritical))
	assert.Equal(t, "FF6600", notifier.getSeverityColor(AlertSeverityHigh))
	assert.Equal(t, "FFFF00", notifier.getSeverityColor(AlertSeverityMedium))
	assert.Equal(t, "00FF00", notifier.getSeverityColor(AlertSeverityLow))
	assert.Equal(t, "0078D4", notifier.getSeverityColor(AlertSeverityInfo))
	assert.Equal(t, "808080", notifier.getSeverityColor("unknown"))

	// Test system status colors
	assert.Equal(t, "00FF00", notifier.getSystemStatusColor("healthy"))
	assert.Equal(t, "FFFF00", notifier.getSystemStatusColor("warning"))
	assert.Equal(t, "FF0000", notifier.getSystemStatusColor("critical"))
	assert.Equal(t, "0078D4", notifier.getSystemStatusColor("unknown"))

	// Test status emojis
	assert.Equal(t, "🚨", notifier.getStatusEmoji(AlertStatusFiring))
	assert.Equal(t, "✅", notifier.getStatusEmoji(AlertStatusResolved))
	assert.Equal(t, "🔇", notifier.getStatusEmoji(AlertStatusSilenced))
	assert.Equal(t, "ℹ️", notifier.getStatusEmoji("unknown"))

	// Test system status emojis
	assert.Equal(t, "✅", notifier.getSystemStatusEmoji("healthy"))
	assert.Equal(t, "⚠️", notifier.getSystemStatusEmoji("warning"))
	assert.Equal(t, "🚨", notifier.getSystemStatusEmoji("critical"))
	assert.Equal(t, "ℹ️", notifier.getSystemStatusEmoji("unknown"))

	// Test card text colors
	assert.Equal(t, "Attention", notifier.getCardTextColor(AlertStatusFiring))
	assert.Equal(t, "Good", notifier.getCardTextColor(AlertStatusResolved))
	assert.Equal(t, "Warning", notifier.getCardTextColor(AlertStatusSilenced))
	assert.Equal(t, "Default", notifier.getCardTextColor("unknown"))

	// Test system status text colors
	assert.Equal(t, "Good", notifier.getSystemStatusTextColor("healthy"))
	assert.Equal(t, "Warning", notifier.getSystemStatusTextColor("warning"))
	assert.Equal(t, "Attention", notifier.getSystemStatusTextColor("critical"))
	assert.Equal(t, "Default", notifier.getSystemStatusTextColor("unknown"))
}

func TestTeamsNotifier_ErrorHandling(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Test with empty webhook URL
	config := &TeamsConfig{
		WebhookURL: "",
		Enabled:    true,
	}

	notifier := NewTeamsNotifier(config, logger)
	ctx := context.Background()

	// Test SendAlert with empty webhook URL
	alert := &AlertInstance{
		ID:       "test-alert",
		RuleName: "Test Rule",
		Severity: "high",
		Status:   AlertStatusFiring,
		Message:  "Test message",
	}

	err := notifier.SendAlert(ctx, alert)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Teams webhook URL not configured")

	// Test SendSystemStatus with empty webhook URL
	status := &SystemStatus{
		Status:      "healthy",
		Uptime:      "1h",
		ActiveTasks: 0,
	}

	err = notifier.SendSystemStatus(ctx, status)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Teams webhook URL not configured")

	// Test SendCustomMessage with empty webhook URL
	err = notifier.SendCustomMessage(ctx, "Test", "Test message", AlertSeverityInfo)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Teams webhook URL not configured")

	// Test TestConnection with empty webhook URL
	err = notifier.TestConnection(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Teams webhook URL not configured")
}

func TestTeamsConfig_Structure(t *testing.T) {
	config := TeamsConfig{
		WebhookURL: "https://test.webhook.office.com/webhookb2/test",
		Enabled:    true,
	}

	assert.Equal(t, "https://test.webhook.office.com/webhookb2/test", config.WebhookURL)
	assert.True(t, config.Enabled)
}

func TestTeamsAdaptiveCard_Structure(t *testing.T) {
	card := TeamsAdaptiveCard{
		ContentType: "application/vnd.microsoft.card.adaptive",
		Content: TeamsCardContent{
			Schema:  "http://adaptivecards.io/schemas/adaptive-card.json",
			Type:    "AdaptiveCard",
			Version: "1.2",
			Body: []TeamsCardBody{
				{
					Type:   "TextBlock",
					Text:   "Test Title",
					Size:   "Medium",
					Weight: "Bolder",
					Color:  "Good",
					Wrap:   true,
				},
			},
			Actions: []TeamsCardAction{
				{
					Type:  "Action.OpenUrl",
					Title: "Test Action",
					URL:   "https://example.com",
				},
			},
		},
	}

	assert.Equal(t, "application/vnd.microsoft.card.adaptive", card.ContentType)
	assert.Equal(t, "AdaptiveCard", card.Content.Type)
	assert.Equal(t, "1.2", card.Content.Version)
	assert.Len(t, card.Content.Body, 1)
	assert.Equal(t, "TextBlock", card.Content.Body[0].Type)
	assert.Equal(t, "Test Title", card.Content.Body[0].Text)
	assert.Len(t, card.Content.Actions, 1)
	assert.Equal(t, "Action.OpenUrl", card.Content.Actions[0].Type)
	assert.Equal(t, "Test Action", card.Content.Actions[0].Title)
}
