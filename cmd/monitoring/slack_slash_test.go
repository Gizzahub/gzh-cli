package monitoring

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestSlackNotifier_ProcessSlashCommand(t *testing.T) {
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
		name             string
		command          SlackSlashCommand
		expectedType     string
		expectedContains string
	}{
		{
			name: "Help command",
			command: SlackSlashCommand{
				Command:     "/gzh",
				Text:        "help",
				UserName:    "testuser",
				ChannelName: "test-channel",
			},
			expectedType:     "ephemeral",
			expectedContains: "GZH Monitoring Commands",
		},
		{
			name: "Status command",
			command: SlackSlashCommand{
				Command:     "/gzh",
				Text:        "status",
				UserName:    "testuser",
				ChannelName: "test-channel",
			},
			expectedType:     "ephemeral",
			expectedContains: "GZH Monitoring System Status",
		},
		{
			name: "Status command public",
			command: SlackSlashCommand{
				Command:     "/gzh",
				Text:        "status public",
				UserName:    "testuser",
				ChannelName: "test-channel",
			},
			expectedType:     "in_channel",
			expectedContains: "GZH Monitoring System Status",
		},
		{
			name: "Alerts command",
			command: SlackSlashCommand{
				Command:     "/gzh",
				Text:        "alerts",
				UserName:    "testuser",
				ChannelName: "test-channel",
			},
			expectedType:     "ephemeral",
			expectedContains: "All Alerts Summary",
		},
		{
			name: "Alerts firing command",
			command: SlackSlashCommand{
				Command:     "/gzh",
				Text:        "alerts firing",
				UserName:    "testuser",
				ChannelName: "test-channel",
			},
			expectedType:     "ephemeral",
			expectedContains: "Firing Alerts",
		},
		{
			name: "Alerts resolved command",
			command: SlackSlashCommand{
				Command:     "/gzh",
				Text:        "alerts resolved",
				UserName:    "testuser",
				ChannelName: "test-channel",
			},
			expectedType:     "ephemeral",
			expectedContains: "Recently Resolved Alerts",
		},
		{
			name: "Silence command",
			command: SlackSlashCommand{
				Command:     "/gzh",
				Text:        "silence alert-123",
				UserName:    "testuser",
				UserID:      "U123456",
				ChannelName: "test-channel",
			},
			expectedType:     "in_channel",
			expectedContains: "has been silenced",
		},
		{
			name: "Resolve command",
			command: SlackSlashCommand{
				Command:     "/gzh",
				Text:        "resolve alert-123",
				UserName:    "testuser",
				UserID:      "U123456",
				ChannelName: "test-channel",
			},
			expectedType:     "in_channel",
			expectedContains: "has been resolved",
		},
		{
			name: "Test command",
			command: SlackSlashCommand{
				Command:     "/gzh",
				Text:        "test",
				UserName:    "testuser",
				UserID:      "U123456",
				ChannelName: "test-channel",
			},
			expectedType:     "ephemeral",
			expectedContains: "Test notification sent",
		},
		{
			name: "Test alert command",
			command: SlackSlashCommand{
				Command:     "/gzh",
				Text:        "test alert",
				UserName:    "testuser",
				UserID:      "U123456",
				ChannelName: "test-channel",
			},
			expectedType:     "ephemeral",
			expectedContains: "Test alert notification sent",
		},
		{
			name: "Unknown command",
			command: SlackSlashCommand{
				Command:     "/gzh",
				Text:        "unknown",
				UserName:    "testuser",
				ChannelName: "test-channel",
			},
			expectedType:     "ephemeral",
			expectedContains: "Unknown command",
		},
		{
			name: "Empty command shows help",
			command: SlackSlashCommand{
				Command:     "/gzh",
				Text:        "",
				UserName:    "testuser",
				ChannelName: "test-channel",
			},
			expectedType:     "ephemeral",
			expectedContains: "GZH Monitoring Commands",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			response, err := notifier.ProcessSlashCommand(&tc.command)

			assert.NoError(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, tc.expectedType, response.ResponseType)

			// Check if expected text is contained in response
			found := false
			if response.Text != "" && contains(response.Text, tc.expectedContains) {
				found = true
			}
			for _, attachment := range response.Attachments {
				if attachment.Text != "" && contains(attachment.Text, tc.expectedContains) {
					found = true
					break
				}
				if attachment.Title != "" && contains(attachment.Title, tc.expectedContains) {
					found = true
					break
				}
			}
			assert.True(t, found, "Expected to find '%s' in response", tc.expectedContains)
		})
	}
}

func TestSlackNotifier_SlashCommandUsageErrors(t *testing.T) {
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
		name         string
		command      SlackSlashCommand
		expectedText string
	}{
		{
			name: "Silence without alert ID",
			command: SlackSlashCommand{
				Command:  "/gzh",
				Text:     "silence",
				UserName: "testuser",
			},
			expectedText: "Usage: `/gzh silence <alert-id>`",
		},
		{
			name: "Resolve without alert ID",
			command: SlackSlashCommand{
				Command:  "/gzh",
				Text:     "resolve",
				UserName: "testuser",
			},
			expectedText: "Usage: `/gzh resolve <alert-id>`",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			response, err := notifier.ProcessSlashCommand(&tc.command)

			assert.NoError(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, "ephemeral", response.ResponseType)
			assert.Contains(t, response.Text, tc.expectedText)
		})
	}
}

func TestSlackNotifier_SlashCommandStructures(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &SlackConfig{
		WebhookURL: "https://hooks.slack.com/services/test",
		Channel:    "#test-channel",
		Username:   "Test Bot",
		IconEmoji:  ":test:",
		Enabled:    true,
	}

	notifier := NewSlackNotifier(config, logger)

	// Test status command with actions
	cmd := &SlackSlashCommand{
		Command:     "/gzh",
		Text:        "status",
		UserName:    "testuser",
		ChannelName: "test-channel",
	}

	response, err := notifier.ProcessSlashCommand(cmd)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Len(t, response.Attachments, 1)

	attachment := response.Attachments[0]
	assert.NotEmpty(t, attachment.Text)
	assert.NotEmpty(t, attachment.Actions)
	assert.Equal(t, "status_command", attachment.CallbackID)

	// Verify actions
	assert.Len(t, attachment.Actions, 3) // Dashboard, Refresh, Metrics
	assert.Equal(t, "view_dashboard", attachment.Actions[0].Name)
	assert.Equal(t, "refresh_status", attachment.Actions[1].Name)
	assert.Equal(t, "view_metrics", attachment.Actions[2].Name)
}

func TestSlackNotifier_AlertCommandColors(t *testing.T) {
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
		command       string
		expectedColor string
	}{
		{"alerts firing", "danger"},
		{"alerts resolved", "good"},
		{"alerts", "#439FE0"},
	}

	for _, tc := range testCases {
		t.Run(tc.command, func(t *testing.T) {
			cmd := &SlackSlashCommand{
				Command:     "/gzh",
				Text:        tc.command,
				UserName:    "testuser",
				ChannelName: "test-channel",
			}

			response, err := notifier.ProcessSlashCommand(cmd)
			assert.NoError(t, err)
			assert.NotNil(t, response)
			assert.Len(t, response.Attachments, 1)
			assert.Equal(t, tc.expectedColor, response.Attachments[0].Color)
		})
	}
}

func TestSlackSlashCommand_Structure(t *testing.T) {
	cmd := SlackSlashCommand{
		Token:               "test_token",
		TeamID:              "T123456",
		TeamDomain:          "test-team",
		ChannelID:           "C123456",
		ChannelName:         "test-channel",
		UserID:              "U123456",
		UserName:            "testuser",
		Command:             "/gzh",
		Text:                "status",
		ResponseURL:         "https://hooks.slack.com/commands/response",
		TriggerID:           "trigger_123",
		APIAppID:            "A123456",
		IsEnterpriseInstall: "false",
	}

	assert.Equal(t, "test_token", cmd.Token)
	assert.Equal(t, "T123456", cmd.TeamID)
	assert.Equal(t, "test-team", cmd.TeamDomain)
	assert.Equal(t, "C123456", cmd.ChannelID)
	assert.Equal(t, "test-channel", cmd.ChannelName)
	assert.Equal(t, "U123456", cmd.UserID)
	assert.Equal(t, "testuser", cmd.UserName)
	assert.Equal(t, "/gzh", cmd.Command)
	assert.Equal(t, "status", cmd.Text)
	assert.Equal(t, "https://hooks.slack.com/commands/response", cmd.ResponseURL)
	assert.Equal(t, "trigger_123", cmd.TriggerID)
	assert.Equal(t, "A123456", cmd.APIAppID)
	assert.Equal(t, "false", cmd.IsEnterpriseInstall)
}

func TestSlackCommandResponse_Structure(t *testing.T) {
	response := SlackCommandResponse{
		ResponseType: "ephemeral",
		Text:         "Test response",
		Attachments: []SlackAttachment{
			{
				Color: "good",
				Title: "Test Title",
				Text:  "Test Text",
				Actions: []SlackAction{
					{
						Type:  "button",
						Text:  "Test Button",
						Name:  "test_action",
						Value: "test_value",
					},
				},
			},
		},
	}

	assert.Equal(t, "ephemeral", response.ResponseType)
	assert.Equal(t, "Test response", response.Text)
	assert.Len(t, response.Attachments, 1)
	assert.Equal(t, "good", response.Attachments[0].Color)
	assert.Equal(t, "Test Title", response.Attachments[0].Title)
	assert.Len(t, response.Attachments[0].Actions, 1)
	assert.Equal(t, "test_action", response.Attachments[0].Actions[0].Name)
}

// Helper function to check if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
