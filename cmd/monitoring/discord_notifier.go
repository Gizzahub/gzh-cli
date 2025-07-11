package monitoring

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// DiscordNotifier handles Discord webhook notifications
type DiscordNotifier struct {
	webhookURL string
	username   string
	avatarURL  string
	httpClient *http.Client
	logger     *zap.Logger
}

// DiscordConfig represents Discord notification configuration
type DiscordConfig struct {
	WebhookURL string `json:"webhook_url"`
	Username   string `json:"username"`
	AvatarURL  string `json:"avatar_url"`
	Enabled    bool   `json:"enabled"`
}

// DiscordMessage represents a Discord message payload
type DiscordMessage struct {
	Username  string          `json:"username,omitempty"`
	AvatarURL string          `json:"avatar_url,omitempty"`
	Content   string          `json:"content,omitempty"`
	Embeds    []DiscordEmbed  `json:"embeds,omitempty"`
}

// DiscordEmbed represents a Discord embed message
type DiscordEmbed struct {
	Title       string             `json:"title,omitempty"`
	Type        string             `json:"type,omitempty"`
	Description string             `json:"description,omitempty"`
	URL         string             `json:"url,omitempty"`
	Color       int                `json:"color,omitempty"`
	Footer      *DiscordFooter     `json:"footer,omitempty"`
	Thumbnail   *DiscordThumbnail  `json:"thumbnail,omitempty"`
	Fields      []DiscordField     `json:"fields,omitempty"`
	Timestamp   string             `json:"timestamp,omitempty"`
}

// DiscordFooter represents embed footer
type DiscordFooter struct {
	Text    string `json:"text"`
	IconURL string `json:"icon_url,omitempty"`
}

// DiscordThumbnail represents embed thumbnail
type DiscordThumbnail struct {
	URL string `json:"url"`
}

// DiscordField represents embed field
type DiscordField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

// NewDiscordNotifier creates a new Discord notifier
func NewDiscordNotifier(config *DiscordConfig, logger *zap.Logger) *DiscordNotifier {
	avatarURL := config.AvatarURL
	if avatarURL == "" {
		avatarURL = "https://cdn-icons-png.flaticon.com/512/3131/3131636.png"
	}

	return &DiscordNotifier{
		webhookURL: config.WebhookURL,
		username:   config.Username,
		avatarURL:  avatarURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

// SendAlert sends an alert notification to Discord
func (d *DiscordNotifier) SendAlert(ctx context.Context, alert *AlertInstance) error {
	if d.webhookURL == "" {
		return fmt.Errorf("Discord webhook URL not configured")
	}

	message := d.formatAlertMessage(alert)
	return d.sendMessage(ctx, message)
}

// SendSystemStatus sends a system status notification to Discord
func (d *DiscordNotifier) SendSystemStatus(ctx context.Context, status *SystemStatus) error {
	if d.webhookURL == "" {
		return fmt.Errorf("Discord webhook URL not configured")
	}

	message := d.formatSystemStatusMessage(status)
	return d.sendMessage(ctx, message)
}

// SendCustomMessage sends a custom message to Discord
func (d *DiscordNotifier) SendCustomMessage(ctx context.Context, title, text string, severity AlertSeverity) error {
	if d.webhookURL == "" {
		return fmt.Errorf("Discord webhook URL not configured")
	}

	color := d.getSeverityColor(severity)

	message := &DiscordMessage{
		Username:  d.username,
		AvatarURL: d.avatarURL,
		Embeds: []DiscordEmbed{
			{
				Title:       title,
				Description: text,
				Color:       color,
				Footer: &DiscordFooter{
					Text:    "GZH Monitoring",
					IconURL: "https://cdn-icons-png.flaticon.com/512/3131/3131636.png",
				},
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			},
		},
	}

	return d.sendMessage(ctx, message)
}

// TestConnection tests the Discord webhook connection
func (d *DiscordNotifier) TestConnection(ctx context.Context) error {
	message := &DiscordMessage{
		Username:  d.username,
		AvatarURL: d.avatarURL,
		Content:   "🧪 Test message from GZH Monitoring System",
		Embeds: []DiscordEmbed{
			{
				Title:       "Connection Test",
				Description: "If you see this message, the Discord integration is working correctly!",
				Color:       0x00FF00, // Green color
				Fields: []DiscordField{
					{
						Name:   "Test Time",
						Value:  time.Now().Format("2006-01-02 15:04:05 MST"),
						Inline: true,
					},
					{
						Name:   "Status",
						Value:  "✅ Success",
						Inline: true,
					},
				},
				Footer: &DiscordFooter{
					Text:    "GZH Monitoring",
					IconURL: "https://cdn-icons-png.flaticon.com/512/3131/3131636.png",
				},
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			},
		},
	}

	return d.sendMessage(ctx, message)
}

// formatAlertMessage formats an alert as a Discord message
func (d *DiscordNotifier) formatAlertMessage(alert *AlertInstance) *DiscordMessage {
	color := d.getSeverityColor(AlertSeverity(alert.Severity))
	statusEmoji := d.getStatusEmoji(alert.Status)

	title := fmt.Sprintf("%s %s", statusEmoji, alert.RuleName)

	var description string
	if alert.Status == AlertStatusFiring {
		description = fmt.Sprintf("🚨 **Alert is firing**\n%s", alert.Message)
	} else if alert.Status == AlertStatusResolved {
		description = fmt.Sprintf("✅ **Alert resolved**\n%s", alert.Message)
	}

	fields := []DiscordField{
		{
			Name:   "Severity",
			Value:  string(alert.Severity),
			Inline: true,
		},
		{
			Name:   "Status",
			Value:  string(alert.Status),
			Inline: true,
		},
	}

	if alert.FiredAt != nil {
		fields = append(fields, DiscordField{
			Name:   "Fired At",
			Value:  alert.FiredAt.Format("2006-01-02 15:04:05 MST"),
			Inline: true,
		})
	}

	if alert.ResolvedAt != nil {
		fields = append(fields, DiscordField{
			Name:   "Resolved At",
			Value:  alert.ResolvedAt.Format("2006-01-02 15:04:05 MST"),
			Inline: true,
		})
	}

	// Add labels as fields
	for key, value := range alert.Labels {
		fields = append(fields, DiscordField{
			Name:   key,
			Value:  value,
			Inline: true,
		})
	}

	return &DiscordMessage{
		Username:  d.username,
		AvatarURL: d.avatarURL,
		Embeds: []DiscordEmbed{
			{
				Title:       title,
				Description: description,
				Color:       color,
				Fields:      fields,
				Footer: &DiscordFooter{
					Text:    "GZH Monitoring",
					IconURL: "https://cdn-icons-png.flaticon.com/512/3131/3131636.png",
				},
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			},
		},
	}
}

// formatSystemStatusMessage formats system status as a Discord message
func (d *DiscordNotifier) formatSystemStatusMessage(status *SystemStatus) *DiscordMessage {
	var color int
	var statusEmoji string

	switch status.Status {
	case "healthy":
		color = 0x00FF00 // Green
		statusEmoji = "✅"
	case "warning":
		color = 0xFFFF00 // Yellow
		statusEmoji = "⚠️"
	case "critical":
		color = 0xFF0000 // Red
		statusEmoji = "🚨"
	default:
		color = 0x0099FF // Blue
		statusEmoji = "ℹ️"
	}

	title := fmt.Sprintf("%s System Status: %s", statusEmoji, status.Status)

	fields := []DiscordField{
		{
			Name:   "Uptime",
			Value:  status.Uptime,
			Inline: true,
		},
		{
			Name:   "Active Tasks",
			Value:  fmt.Sprintf("%d", status.ActiveTasks),
			Inline: true,
		},
		{
			Name:   "Memory Usage",
			Value:  formatBytesForDiscord(status.MemoryUsage),
			Inline: true,
		},
		{
			Name:   "CPU Usage",
			Value:  fmt.Sprintf("%.1f%%", status.CPUUsage),
			Inline: true,
		},
		{
			Name:   "Total Requests",
			Value:  fmt.Sprintf("%d", status.TotalRequests),
			Inline: true,
		},
	}

	if status.DiskUsage > 0 {
		fields = append(fields, DiscordField{
			Name:   "Disk Usage",
			Value:  fmt.Sprintf("%.1f%%", status.DiskUsage),
			Inline: true,
		})
	}

	return &DiscordMessage{
		Username:  d.username,
		AvatarURL: d.avatarURL,
		Embeds: []DiscordEmbed{
			{
				Title:  title,
				Color:  color,
				Fields: fields,
				Footer: &DiscordFooter{
					Text:    "GZH Monitoring",
					IconURL: "https://cdn-icons-png.flaticon.com/512/3131/3131636.png",
				},
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			},
		},
	}
}

// sendMessage sends a message to Discord webhook
func (d *DiscordNotifier) sendMessage(ctx context.Context, message *DiscordMessage) error {
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal Discord message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", d.webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Discord returns 204 No Content on success
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Discord webhook returned status code: %d", resp.StatusCode)
	}

	d.logger.Info("Discord message sent successfully",
		zap.String("username", d.username),
		zap.Int("status_code", resp.StatusCode))

	return nil
}

// getSeverityColor returns the color for alert severity (Discord uses decimal color values)
func (d *DiscordNotifier) getSeverityColor(severity AlertSeverity) int {
	switch severity {
	case AlertSeverityCritical:
		return 0xFF0000 // Red
	case AlertSeverityHigh:
		return 0xFF6600 // Orange
	case AlertSeverityMedium:
		return 0xFFFF00 // Yellow
	case AlertSeverityLow:
		return 0x00FF00 // Green
	case AlertSeverityInfo:
		return 0x0099FF // Blue
	default:
		return 0x808080 // Gray
	}
}

// getStatusEmoji returns emoji for alert status
func (d *DiscordNotifier) getStatusEmoji(status AlertStatus) string {
	switch status {
	case AlertStatusFiring:
		return "🚨"
	case AlertStatusResolved:
		return "✅"
	case AlertStatusSilenced:
		return "🔇"
	default:
		return "ℹ️"
	}
}

// formatBytesForDiscord formats bytes to human readable format
func formatBytesForDiscord(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}