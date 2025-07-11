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

func TestNewDiscordNotifier(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &DiscordConfig{
		WebhookURL: "https://discord.com/api/webhooks/123456/test",
		Username:   "Test Bot",
		AvatarURL:  "https://example.com/avatar.png",
		Enabled:    true,
	}

	notifier := NewDiscordNotifier(config, logger)

	assert.Equal(t, config.WebhookURL, notifier.webhookURL)
	assert.Equal(t, config.Username, notifier.username)
	assert.Equal(t, config.AvatarURL, notifier.avatarURL)
	assert.NotNil(t, notifier.httpClient)
	assert.NotNil(t, notifier.logger)
}

func TestNewDiscordNotifier_DefaultAvatar(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &DiscordConfig{
		WebhookURL: "https://discord.com/api/webhooks/123456/test",
		Username:   "Test Bot",
		AvatarURL:  "", // Empty avatar URL
		Enabled:    true,
	}

	notifier := NewDiscordNotifier(config, logger)

	assert.Equal(t, "https://cdn-icons-png.flaticon.com/512/3131/3131636.png", notifier.avatarURL)
}

func TestDiscordNotifier_SendAlert(t *testing.T) {
	// Create a test server to mock Discord webhook
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusNoContent) // Discord returns 204 on success
	}))
	defer server.Close()

	logger := zaptest.NewLogger(t)
	config := &DiscordConfig{
		WebhookURL: server.URL,
		Username:   "GZH Monitoring",
		AvatarURL:  "",
		Enabled:    true,
	}

	notifier := NewDiscordNotifier(config, logger)

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

func TestDiscordNotifier_SendSystemStatus(t *testing.T) {
	// Create a test server to mock Discord webhook
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	logger := zaptest.NewLogger(t)
	config := &DiscordConfig{
		WebhookURL: server.URL,
		Username:   "GZH Monitoring",
		AvatarURL:  "",
		Enabled:    true,
	}

	notifier := NewDiscordNotifier(config, logger)

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

func TestDiscordNotifier_SendCustomMessage(t *testing.T) {
	// Create a test server to mock Discord webhook
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	logger := zaptest.NewLogger(t)
	config := &DiscordConfig{
		WebhookURL: server.URL,
		Username:   "Test Bot",
		AvatarURL:  "",
		Enabled:    true,
	}

	notifier := NewDiscordNotifier(config, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := notifier.SendCustomMessage(ctx, "Test Title", "This is a test message", AlertSeverityInfo)
	assert.NoError(t, err)
}

func TestDiscordNotifier_TestConnection(t *testing.T) {
	// Create a test server to mock Discord webhook
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	logger := zaptest.NewLogger(t)
	config := &DiscordConfig{
		WebhookURL: server.URL,
		Username:   "Test Bot",
		AvatarURL:  "",
		Enabled:    true,
	}

	notifier := NewDiscordNotifier(config, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := notifier.TestConnection(ctx)
	assert.NoError(t, err)
}

func TestDiscordNotifier_EmptyWebhookURL(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &DiscordConfig{
		WebhookURL: "",
		Username:   "Test Bot",
		AvatarURL:  "",
		Enabled:    true,
	}

	notifier := NewDiscordNotifier(config, logger)

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

func TestDiscordNotifier_HTTPError(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad Request"))
	}))
	defer server.Close()

	logger := zaptest.NewLogger(t)
	config := &DiscordConfig{
		WebhookURL: server.URL,
		Username:   "Test Bot",
		AvatarURL:  "",
		Enabled:    true,
	}

	notifier := NewDiscordNotifier(config, logger)

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

func TestDiscordNotifier_GetSeverityColor(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &DiscordConfig{
		WebhookURL: "https://test.com",
		Username:   "Test Bot",
		AvatarURL:  "",
		Enabled:    true,
	}

	notifier := NewDiscordNotifier(config, logger)

	testCases := []struct {
		severity AlertSeverity
		expected int
	}{
		{AlertSeverityCritical, 0xFF0000}, // Red
		{AlertSeverityHigh, 0xFF6600},     // Orange
		{AlertSeverityMedium, 0xFFFF00},   // Yellow
		{AlertSeverityLow, 0x00FF00},      // Green
		{AlertSeverityInfo, 0x0099FF},     // Blue
		{AlertSeverity("unknown"), 0x808080}, // Gray
	}

	for _, tc := range testCases {
		t.Run(string(tc.severity), func(t *testing.T) {
			color := notifier.getSeverityColor(tc.severity)
			assert.Equal(t, tc.expected, color)
		})
	}
}

func TestDiscordNotifier_GetStatusEmoji(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &DiscordConfig{
		WebhookURL: "https://test.com",
		Username:   "Test Bot",
		AvatarURL:  "",
		Enabled:    true,
	}

	notifier := NewDiscordNotifier(config, logger)

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

func TestFormatBytesForDiscord(t *testing.T) {
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
			result := formatBytesForDiscord(tc.bytes)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestDiscordMessage_Structure(t *testing.T) {
	// Test that DiscordMessage can be properly structured
	message := &DiscordMessage{
		Username:  "Test Bot",
		AvatarURL: "https://example.com/avatar.png",
		Content:   "Test content",
		Embeds: []DiscordEmbed{
			{
				Title:       "Test Title",
				Description: "Test description",
				Color:       0xFF0000,
				Fields: []DiscordField{
					{
						Name:   "Field 1",
						Value:  "Value 1",
						Inline: true,
					},
				},
				Footer: &DiscordFooter{
					Text:    "Test Footer",
					IconURL: "https://example.com/icon.png",
				},
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			},
		},
	}

	// Verify the message structure is valid
	assert.Equal(t, "Test Bot", message.Username)
	assert.Equal(t, "https://example.com/avatar.png", message.AvatarURL)
	assert.Equal(t, "Test content", message.Content)
	assert.Len(t, message.Embeds, 1)
	assert.Equal(t, "Test Title", message.Embeds[0].Title)
	assert.Equal(t, 0xFF0000, message.Embeds[0].Color)
	assert.Len(t, message.Embeds[0].Fields, 1)
}

func TestDiscordEmbed_CompleteFields(t *testing.T) {
	// Test that all Discord embed fields work correctly
	embed := DiscordEmbed{
		Title:       "Test Title",
		Type:        "rich",
		Description: "Test Description",
		URL:         "https://example.com",
		Color:       0x00FF00,
		Footer: &DiscordFooter{
			Text:    "Footer Text",
			IconURL: "https://example.com/footer.png",
		},
		Thumbnail: &DiscordThumbnail{
			URL: "https://example.com/thumb.png",
		},
		Fields: []DiscordField{
			{
				Name:   "Field Name",
				Value:  "Field Value",
				Inline: true,
			},
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	assert.Equal(t, "Test Title", embed.Title)
	assert.Equal(t, "rich", embed.Type)
	assert.Equal(t, "Test Description", embed.Description)
	assert.Equal(t, "https://example.com", embed.URL)
	assert.Equal(t, 0x00FF00, embed.Color)
	assert.NotNil(t, embed.Footer)
	assert.NotNil(t, embed.Thumbnail)
	assert.Len(t, embed.Fields, 1)
	assert.NotEmpty(t, embed.Timestamp)
}