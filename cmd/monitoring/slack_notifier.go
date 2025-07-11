package monitoring

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

// SlackNotifier handles Slack webhook notifications
type SlackNotifier struct {
	webhookURL string
	channel    string
	username   string
	iconEmoji  string
	httpClient *http.Client
	logger     *zap.Logger
}

// SlackConfig represents Slack notification configuration
type SlackConfig struct {
	WebhookURL string `json:"webhook_url"`
	Channel    string `json:"channel"`
	Username   string `json:"username"`
	IconEmoji  string `json:"icon_emoji"`
	Enabled    bool   `json:"enabled"`
}

// SlackMessage represents a Slack message payload
type SlackMessage struct {
	Channel     string            `json:"channel,omitempty"`
	Username    string            `json:"username,omitempty"`
	IconEmoji   string            `json:"icon_emoji,omitempty"`
	Text        string            `json:"text,omitempty"`
	Attachments []SlackAttachment `json:"attachments,omitempty"`
}

// SlackAttachment represents a Slack message attachment
type SlackAttachment struct {
	Color      string        `json:"color,omitempty"`
	Title      string        `json:"title,omitempty"`
	TitleLink  string        `json:"title_link,omitempty"`
	Text       string        `json:"text,omitempty"`
	Fields     []SlackField  `json:"fields,omitempty"`
	Actions    []SlackAction `json:"actions,omitempty"`
	Footer     string        `json:"footer,omitempty"`
	FooterIcon string        `json:"footer_icon,omitempty"`
	Timestamp  int64         `json:"ts,omitempty"`
	CallbackID string        `json:"callback_id,omitempty"`
}

// SlackField represents a field in Slack attachment
type SlackField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

// SlackAction represents an interactive action in Slack message
type SlackAction struct {
	Type  string `json:"type"`
	Text  string `json:"text"`
	Value string `json:"value"`
	Name  string `json:"name"`
	Style string `json:"style,omitempty"` // primary, danger, default
	Url   string `json:"url,omitempty"`
}

// SlackInteractionPayload represents payload from Slack interactive components
type SlackInteractionPayload struct {
	Type            string                `json:"type"`
	Actions         []SlackActionResponse `json:"actions"`
	CallbackID      string                `json:"callback_id"`
	Team            SlackTeam             `json:"team"`
	Channel         SlackChannel          `json:"channel"`
	User            SlackUser             `json:"user"`
	ActionTS        string                `json:"action_ts"`
	MessageTS       string                `json:"message_ts"`
	AttachmentID    string                `json:"attachment_id"`
	Token           string                `json:"token"`
	OriginalMessage SlackMessage          `json:"original_message"`
	ResponseURL     string                `json:"response_url"`
}

// SlackActionResponse represents action response from user interaction
type SlackActionResponse struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

// SlackTeam represents Slack team info
type SlackTeam struct {
	ID     string `json:"id"`
	Domain string `json:"domain"`
}

// SlackChannel represents Slack channel info
type SlackChannel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// SlackUser represents Slack user info
type SlackUser struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// NewSlackNotifier creates a new Slack notifier
func NewSlackNotifier(config *SlackConfig, logger *zap.Logger) *SlackNotifier {
	return &SlackNotifier{
		webhookURL: config.WebhookURL,
		channel:    config.Channel,
		username:   config.Username,
		iconEmoji:  config.IconEmoji,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

// SendAlert sends an alert notification to Slack
func (s *SlackNotifier) SendAlert(ctx context.Context, alert *AlertInstance) error {
	if s.webhookURL == "" {
		return fmt.Errorf("Slack webhook URL not configured")
	}

	message := s.formatAlertMessage(alert)
	return s.sendMessage(ctx, message)
}

// SendInteractiveAlert sends an interactive alert notification to Slack
func (s *SlackNotifier) SendInteractiveAlert(ctx context.Context, alert *AlertInstance) error {
	if s.webhookURL == "" {
		return fmt.Errorf("Slack webhook URL not configured")
	}

	message := s.formatInteractiveAlertMessage(alert)
	return s.sendMessage(ctx, message)
}

// SendSystemStatus sends a system status notification to Slack
func (s *SlackNotifier) SendSystemStatus(ctx context.Context, status *SystemStatus) error {
	if s.webhookURL == "" {
		return fmt.Errorf("Slack webhook URL not configured")
	}

	message := s.formatSystemStatusMessage(status)
	return s.sendMessage(ctx, message)
}

// SendCustomMessage sends a custom message to Slack
func (s *SlackNotifier) SendCustomMessage(ctx context.Context, title, text string, severity AlertSeverity) error {
	if s.webhookURL == "" {
		return fmt.Errorf("Slack webhook URL not configured")
	}

	color := s.getSeverityColor(severity)

	message := &SlackMessage{
		Channel:   s.channel,
		Username:  s.username,
		IconEmoji: s.iconEmoji,
		Attachments: []SlackAttachment{
			{
				Color:      color,
				Title:      title,
				Text:       text,
				Footer:     "GZH Monitoring",
				FooterIcon: "https://cdn-icons-png.flaticon.com/512/3131/3131636.png",
				Timestamp:  time.Now().Unix(),
			},
		},
	}

	return s.sendMessage(ctx, message)
}

// TestConnection tests the Slack webhook connection
func (s *SlackNotifier) TestConnection(ctx context.Context) error {
	message := &SlackMessage{
		Channel:   s.channel,
		Username:  s.username,
		IconEmoji: s.iconEmoji,
		Text:      "🧪 Test message from GZH Monitoring System",
		Attachments: []SlackAttachment{
			{
				Color: "good",
				Title: "Connection Test",
				Text:  "If you see this message, the Slack integration is working correctly!",
				Fields: []SlackField{
					{
						Title: "Test Time",
						Value: time.Now().Format("2006-01-02 15:04:05 MST"),
						Short: true,
					},
					{
						Title: "Status",
						Value: "✅ Success",
						Short: true,
					},
				},
				Footer:     "GZH Monitoring",
				FooterIcon: "https://cdn-icons-png.flaticon.com/512/3131/3131636.png",
				Timestamp:  time.Now().Unix(),
			},
		},
	}

	return s.sendMessage(ctx, message)
}

// formatAlertMessage formats an alert as a Slack message
func (s *SlackNotifier) formatAlertMessage(alert *AlertInstance) *SlackMessage {
	color := s.getSeverityColor(AlertSeverity(alert.Severity))
	statusEmoji := s.getStatusEmoji(alert.Status)

	title := fmt.Sprintf("%s %s", statusEmoji, alert.RuleName)

	var text string
	if alert.Status == AlertStatusFiring {
		text = fmt.Sprintf("🚨 *Alert is firing*\n%s", alert.Message)
	} else if alert.Status == AlertStatusResolved {
		text = fmt.Sprintf("✅ *Alert resolved*\n%s", alert.Message)
	}

	fields := []SlackField{
		{
			Title: "Severity",
			Value: string(alert.Severity),
			Short: true,
		},
		{
			Title: "Status",
			Value: string(alert.Status),
			Short: true,
		},
	}

	if alert.FiredAt != nil {
		fields = append(fields, SlackField{
			Title: "Fired At",
			Value: alert.FiredAt.Format("2006-01-02 15:04:05 MST"),
			Short: true,
		})
	}

	if alert.ResolvedAt != nil {
		fields = append(fields, SlackField{
			Title: "Resolved At",
			Value: alert.ResolvedAt.Format("2006-01-02 15:04:05 MST"),
			Short: true,
		})
	}

	// Add labels as fields
	for key, value := range alert.Labels {
		fields = append(fields, SlackField{
			Title: key,
			Value: value,
			Short: true,
		})
	}

	return &SlackMessage{
		Channel:   s.channel,
		Username:  s.username,
		IconEmoji: s.iconEmoji,
		Attachments: []SlackAttachment{
			{
				Color:      color,
				Title:      title,
				Text:       text,
				Fields:     fields,
				Footer:     "GZH Monitoring",
				FooterIcon: "https://cdn-icons-png.flaticon.com/512/3131/3131636.png",
				Timestamp:  time.Now().Unix(),
			},
		},
	}
}

// formatSystemStatusMessage formats system status as a Slack message
func (s *SlackNotifier) formatSystemStatusMessage(status *SystemStatus) *SlackMessage {
	var color string
	var statusEmoji string

	switch status.Status {
	case "healthy":
		color = "good"
		statusEmoji = "✅"
	case "warning":
		color = "warning"
		statusEmoji = "⚠️"
	case "critical":
		color = "danger"
		statusEmoji = "🚨"
	default:
		color = "#439FE0"
		statusEmoji = "ℹ️"
	}

	title := fmt.Sprintf("%s System Status: %s", statusEmoji, strings.ToUpper(status.Status))

	fields := []SlackField{
		{
			Title: "Uptime",
			Value: status.Uptime,
			Short: true,
		},
		{
			Title: "Active Tasks",
			Value: fmt.Sprintf("%d", status.ActiveTasks),
			Short: true,
		},
		{
			Title: "Memory Usage",
			Value: formatBytes(status.MemoryUsage),
			Short: true,
		},
		{
			Title: "CPU Usage",
			Value: fmt.Sprintf("%.1f%%", status.CPUUsage),
			Short: true,
		},
		{
			Title: "Total Requests",
			Value: fmt.Sprintf("%d", status.TotalRequests),
			Short: true,
		},
	}

	if status.DiskUsage > 0 {
		fields = append(fields, SlackField{
			Title: "Disk Usage",
			Value: fmt.Sprintf("%.1f%%", status.DiskUsage),
			Short: true,
		})
	}

	return &SlackMessage{
		Channel:   s.channel,
		Username:  s.username,
		IconEmoji: s.iconEmoji,
		Attachments: []SlackAttachment{
			{
				Color:      color,
				Title:      title,
				Fields:     fields,
				Footer:     "GZH Monitoring",
				FooterIcon: "https://cdn-icons-png.flaticon.com/512/3131/3131636.png",
				Timestamp:  time.Now().Unix(),
			},
		},
	}
}

// sendMessage sends a message to Slack webhook
func (s *SlackNotifier) sendMessage(ctx context.Context, message *SlackMessage) error {
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal Slack message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Slack webhook returned status code: %d", resp.StatusCode)
	}

	s.logger.Info("Slack message sent successfully",
		zap.String("channel", s.channel),
		zap.Int("status_code", resp.StatusCode))

	return nil
}

// getSeverityColor returns the color for alert severity
func (s *SlackNotifier) getSeverityColor(severity AlertSeverity) string {
	switch severity {
	case AlertSeverityCritical:
		return "danger"
	case AlertSeverityHigh:
		return "danger"
	case AlertSeverityMedium:
		return "warning"
	case AlertSeverityLow:
		return "good"
	case AlertSeverityInfo:
		return "#439FE0"
	default:
		return "#D3D3D3"
	}
}

// getStatusEmoji returns emoji for alert status
func (s *SlackNotifier) getStatusEmoji(status AlertStatus) string {
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

// formatBytes formats bytes to human readable format
func formatBytesForSlack(bytes uint64) string {
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

// formatInteractiveAlertMessage formats an alert as an interactive Slack message
func (s *SlackNotifier) formatInteractiveAlertMessage(alert *AlertInstance) *SlackMessage {
	color := s.getSeverityColor(AlertSeverity(alert.Severity))
	statusEmoji := s.getStatusEmoji(alert.Status)

	title := fmt.Sprintf("%s %s", statusEmoji, alert.RuleName)

	var text string
	if alert.Status == AlertStatusFiring {
		text = fmt.Sprintf("🚨 *Alert is firing*\n%s", alert.Message)
	} else if alert.Status == AlertStatusResolved {
		text = fmt.Sprintf("✅ *Alert resolved*\n%s", alert.Message)
	}

	fields := []SlackField{
		{
			Title: "Severity",
			Value: string(alert.Severity),
			Short: true,
		},
		{
			Title: "Status",
			Value: string(alert.Status),
			Short: true,
		},
	}

	if alert.FiredAt != nil {
		fields = append(fields, SlackField{
			Title: "Fired At",
			Value: alert.FiredAt.Format("2006-01-02 15:04:05 MST"),
			Short: true,
		})
	}

	// Add labels as fields
	for key, value := range alert.Labels {
		fields = append(fields, SlackField{
			Title: key,
			Value: value,
			Short: true,
		})
	}

	// Create interactive actions based on alert status
	var actions []SlackAction
	callbackID := fmt.Sprintf("alert_%s", alert.ID)

	if alert.Status == AlertStatusFiring {
		actions = []SlackAction{
			{
				Type:  "button",
				Text:  "🔇 Silence Alert",
				Name:  "silence",
				Value: alert.ID,
				Style: "default",
			},
			{
				Type:  "button",
				Text:  "✅ Resolve",
				Name:  "resolve",
				Value: alert.ID,
				Style: "primary",
			},
			{
				Type:  "button",
				Text:  "📊 View Details",
				Name:  "details",
				Value: alert.ID,
				Style: "default",
				Url:   fmt.Sprintf("http://localhost:8080/alerts/%s", alert.ID),
			},
		}
	} else if alert.Status == AlertStatusSilenced {
		actions = []SlackAction{
			{
				Type:  "button",
				Text:  "🔊 Unsilence",
				Name:  "unsilence",
				Value: alert.ID,
				Style: "default",
			},
			{
				Type:  "button",
				Text:  "✅ Resolve",
				Name:  "resolve",
				Value: alert.ID,
				Style: "primary",
			},
		}
	} else if alert.Status == AlertStatusResolved {
		actions = []SlackAction{
			{
				Type:  "button",
				Text:  "📊 View Details",
				Name:  "details",
				Value: alert.ID,
				Style: "default",
				Url:   fmt.Sprintf("http://localhost:8080/alerts/%s", alert.ID),
			},
		}
	}

	return &SlackMessage{
		Channel:   s.channel,
		Username:  s.username,
		IconEmoji: s.iconEmoji,
		Attachments: []SlackAttachment{
			{
				Color:      color,
				Title:      title,
				Text:       text,
				Fields:     fields,
				Actions:    actions,
				CallbackID: callbackID,
				Footer:     "GZH Monitoring",
				FooterIcon: "https://cdn-icons-png.flaticon.com/512/3131/3131636.png",
				Timestamp:  time.Now().Unix(),
			},
		},
	}
}

// formatInteractiveSystemStatusMessage formats system status as an interactive Slack message
func (s *SlackNotifier) formatInteractiveSystemStatusMessage(status *SystemStatus) *SlackMessage {
	var color string
	var statusEmoji string

	switch status.Status {
	case "healthy":
		color = "good"
		statusEmoji = "✅"
	case "warning":
		color = "warning"
		statusEmoji = "⚠️"
	case "critical":
		color = "danger"
		statusEmoji = "🚨"
	default:
		color = "#439FE0"
		statusEmoji = "ℹ️"
	}

	title := fmt.Sprintf("%s System Status: %s", statusEmoji, strings.ToUpper(status.Status))

	fields := []SlackField{
		{
			Title: "Uptime",
			Value: status.Uptime,
			Short: true,
		},
		{
			Title: "Active Tasks",
			Value: fmt.Sprintf("%d", status.ActiveTasks),
			Short: true,
		},
		{
			Title: "Memory Usage",
			Value: formatBytes(status.MemoryUsage),
			Short: true,
		},
		{
			Title: "CPU Usage",
			Value: fmt.Sprintf("%.1f%%", status.CPUUsage),
			Short: true,
		},
	}

	// Create interactive actions for system status
	actions := []SlackAction{
		{
			Type:  "button",
			Text:  "📊 Full Dashboard",
			Name:  "dashboard",
			Value: "view_dashboard",
			Style: "primary",
			Url:   "http://localhost:8080/dashboard",
		},
		{
			Type:  "button",
			Text:  "🔄 Refresh Status",
			Name:  "refresh",
			Value: "refresh_status",
			Style: "default",
		},
		{
			Type:  "button",
			Text:  "📈 View Metrics",
			Name:  "metrics",
			Value: "view_metrics",
			Style: "default",
			Url:   "http://localhost:8080/metrics",
		},
	}

	return &SlackMessage{
		Channel:   s.channel,
		Username:  s.username,
		IconEmoji: s.iconEmoji,
		Attachments: []SlackAttachment{
			{
				Color:      color,
				Title:      title,
				Fields:     fields,
				Actions:    actions,
				CallbackID: "system_status",
				Footer:     "GZH Monitoring",
				FooterIcon: "https://cdn-icons-png.flaticon.com/512/3131/3131636.png",
				Timestamp:  time.Now().Unix(),
			},
		},
	}
}

// SendInteractiveSystemStatus sends an interactive system status notification to Slack
func (s *SlackNotifier) SendInteractiveSystemStatus(ctx context.Context, status *SystemStatus) error {
	if s.webhookURL == "" {
		return fmt.Errorf("Slack webhook URL not configured")
	}

	message := s.formatInteractiveSystemStatusMessage(status)
	return s.sendMessage(ctx, message)
}

// ProcessInteraction processes Slack interactive message responses
func (s *SlackNotifier) ProcessInteraction(payload *SlackInteractionPayload) (*SlackMessage, error) {
	if len(payload.Actions) == 0 {
		return nil, fmt.Errorf("no actions in payload")
	}

	action := payload.Actions[0]

	switch action.Name {
	case "silence":
		return s.handleSilenceAction(payload, action)
	case "resolve":
		return s.handleResolveAction(payload, action)
	case "unsilence":
		return s.handleUnsilenceAction(payload, action)
	case "refresh":
		return s.handleRefreshAction(payload, action)
	default:
		s.logger.Warn("Unknown action received", zap.String("action", action.Name))
		return nil, fmt.Errorf("unknown action: %s", action.Name)
	}
}

// handleSilenceAction handles alert silence action
func (s *SlackNotifier) handleSilenceAction(payload *SlackInteractionPayload, action SlackActionResponse) (*SlackMessage, error) {
	alertID := action.Value

	// Create response message
	response := &SlackMessage{
		Text: fmt.Sprintf("🔇 Alert `%s` has been silenced for 1 hour by <@%s>", alertID, payload.User.ID),
		Attachments: []SlackAttachment{
			{
				Color:     "warning",
				Text:      "The alert will not trigger notifications until the silence expires or is manually removed.",
				Footer:    "GZH Monitoring",
				Timestamp: time.Now().Unix(),
			},
		},
	}

	s.logger.Info("Alert silenced via Slack interaction",
		zap.String("alert_id", alertID),
		zap.String("user", payload.User.Name))

	return response, nil
}

// handleResolveAction handles alert resolve action
func (s *SlackNotifier) handleResolveAction(payload *SlackInteractionPayload, action SlackActionResponse) (*SlackMessage, error) {
	alertID := action.Value

	// Create response message
	response := &SlackMessage{
		Text: fmt.Sprintf("✅ Alert `%s` has been resolved by <@%s>", alertID, payload.User.ID),
		Attachments: []SlackAttachment{
			{
				Color:     "good",
				Text:      "The alert has been marked as resolved and will no longer trigger notifications.",
				Footer:    "GZH Monitoring",
				Timestamp: time.Now().Unix(),
			},
		},
	}

	s.logger.Info("Alert resolved via Slack interaction",
		zap.String("alert_id", alertID),
		zap.String("user", payload.User.Name))

	return response, nil
}

// handleUnsilenceAction handles alert unsilence action
func (s *SlackNotifier) handleUnsilenceAction(payload *SlackInteractionPayload, action SlackActionResponse) (*SlackMessage, error) {
	alertID := action.Value

	// Create response message
	response := &SlackMessage{
		Text: fmt.Sprintf("🔊 Alert `%s` has been unsilenced by <@%s>", alertID, payload.User.ID),
		Attachments: []SlackAttachment{
			{
				Color:     "#439FE0",
				Text:      "The alert is now active and will trigger notifications if conditions are met.",
				Footer:    "GZH Monitoring",
				Timestamp: time.Now().Unix(),
			},
		},
	}

	s.logger.Info("Alert unsilenced via Slack interaction",
		zap.String("alert_id", alertID),
		zap.String("user", payload.User.Name))

	return response, nil
}

// handleRefreshAction handles system status refresh action
func (s *SlackNotifier) handleRefreshAction(payload *SlackInteractionPayload, action SlackActionResponse) (*SlackMessage, error) {
	// Create response message indicating refresh
	response := &SlackMessage{
		Text: fmt.Sprintf("🔄 System status refresh requested by <@%s>", payload.User.ID),
		Attachments: []SlackAttachment{
			{
				Color:     "#439FE0",
				Text:      "Fetching latest system status...",
				Footer:    "GZH Monitoring",
				Timestamp: time.Now().Unix(),
			},
		},
	}

	s.logger.Info("System status refresh requested via Slack interaction",
		zap.String("user", payload.User.Name))

	return response, nil
}
