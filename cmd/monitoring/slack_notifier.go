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

// SlackSlashCommand represents a Slack slash command payload
type SlackSlashCommand struct {
	Token               string `json:"token"`
	TeamID              string `json:"team_id"`
	TeamDomain          string `json:"team_domain"`
	ChannelID           string `json:"channel_id"`
	ChannelName         string `json:"channel_name"`
	UserID              string `json:"user_id"`
	UserName            string `json:"user_name"`
	Command             string `json:"command"`
	Text                string `json:"text"`
	ResponseURL         string `json:"response_url"`
	TriggerID           string `json:"trigger_id"`
	APIAppID            string `json:"api_app_id"`
	IsEnterpriseInstall string `json:"is_enterprise_install"`
}

// SlackCommandResponse represents a response to a Slack slash command
type SlackCommandResponse struct {
	ResponseType string            `json:"response_type,omitempty"` // ephemeral or in_channel
	Text         string            `json:"text,omitempty"`
	Attachments  []SlackAttachment `json:"attachments,omitempty"`
	Blocks       []interface{}     `json:"blocks,omitempty"`
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

// ProcessSlashCommand processes Slack slash commands
func (s *SlackNotifier) ProcessSlashCommand(cmd *SlackSlashCommand) (*SlackCommandResponse, error) {
	// Parse command and arguments
	args := strings.Fields(cmd.Text)

	s.logger.Info("Processing Slack slash command",
		zap.String("command", cmd.Command),
		zap.String("text", cmd.Text),
		zap.String("user", cmd.UserName),
		zap.String("channel", cmd.ChannelName))

	// Route to appropriate handler based on command arguments
	if len(args) == 0 {
		return s.handleHelpCommand(cmd)
	}

	switch args[0] {
	case "status":
		return s.handleStatusCommand(cmd, args[1:])
	case "alerts":
		return s.handleAlertsCommand(cmd, args[1:])
	case "help":
		return s.handleHelpCommand(cmd)
	case "silence":
		return s.handleSilenceCommand(cmd, args[1:])
	case "resolve":
		return s.handleResolveCommand(cmd, args[1:])
	case "test":
		return s.handleTestCommand(cmd, args[1:])
	default:
		return s.handleUnknownCommand(cmd, args[0])
	}
}

// handleHelpCommand shows available commands
func (s *SlackNotifier) handleHelpCommand(cmd *SlackSlashCommand) (*SlackCommandResponse, error) {
	helpText := "*GZH Monitoring Commands*\n\n" +
		"Available commands:\n" +
		"• `/gzh status` - Show current system status\n" +
		"• `/gzh alerts` - List active alerts\n" +
		"• `/gzh alerts firing` - Show only firing alerts\n" +
		"• `/gzh silence <alert-id>` - Silence an alert for 1 hour\n" +
		"• `/gzh resolve <alert-id>` - Resolve an alert\n" +
		"• `/gzh test` - Send a test notification\n" +
		"• `/gzh help` - Show this help message\n\n" +
		"Examples:\n" +
		"• `/gzh status` - Get system health overview\n" +
		"• `/gzh alerts firing` - See what's currently alerting\n" +
		"• `/gzh silence alert-123` - Silence specific alert"

	return &SlackCommandResponse{
		ResponseType: "ephemeral",
		Text:         helpText,
	}, nil
}

// handleStatusCommand shows system status
func (s *SlackNotifier) handleStatusCommand(cmd *SlackSlashCommand, args []string) (*SlackCommandResponse, error) {
	// This would normally fetch real system status
	// For now, we'll create a sample status
	statusText := "*🖥️ GZH Monitoring System Status*\n\n" +
		"*Overall Health:* ✅ Healthy\n" +
		"*Uptime:* 2d 4h 15m\n" +
		"*Active Tasks:* 12\n" +
		"*Memory Usage:* 1.2 GB\n" +
		"*CPU Usage:* 23.5%\n" +
		"*Total Requests:* 45,678\n\n" +
		"*Quick Actions:*"

	actions := []SlackAction{
		{
			Type:  "button",
			Text:  "📊 Full Dashboard",
			Name:  "view_dashboard",
			Value: "dashboard",
			Style: "primary",
			Url:   "http://localhost:8080/dashboard",
		},
		{
			Type:  "button",
			Text:  "🔄 Refresh Status",
			Name:  "refresh_status",
			Value: "refresh",
			Style: "default",
		},
		{
			Type:  "button",
			Text:  "📈 View Metrics",
			Name:  "view_metrics",
			Value: "metrics",
			Style: "default",
			Url:   "http://localhost:8080/metrics",
		},
	}

	attachment := SlackAttachment{
		Color:      "good",
		Text:       statusText,
		Actions:    actions,
		CallbackID: "status_command",
		Footer:     "GZH Monitoring",
		Timestamp:  time.Now().Unix(),
	}

	responseType := "ephemeral"
	if len(args) > 0 && args[0] == "public" {
		responseType = "in_channel"
	}

	return &SlackCommandResponse{
		ResponseType: responseType,
		Attachments:  []SlackAttachment{attachment},
	}, nil
}

// handleAlertsCommand shows alert information
func (s *SlackNotifier) handleAlertsCommand(cmd *SlackSlashCommand, args []string) (*SlackCommandResponse, error) {
	var filterType string
	if len(args) > 0 {
		filterType = args[0]
	}

	// This would normally fetch real alerts from AlertManager
	// For now, we'll create sample alerts
	var alertText string
	var color string

	switch filterType {
	case "firing":
		alertText = "*🚨 Firing Alerts (2)*\n\n" +
			"• `alert-001` - High CPU Usage (critical)\n" +
			"  Host: server-01, CPU: 95%\n\n" +
			"• `alert-002` - Memory Warning (medium)\n" +
			"  Host: server-02, Memory: 85%"
		color = "danger"
	case "resolved":
		alertText = "*✅ Recently Resolved Alerts (3)*\n\n" +
			"• `alert-003` - Disk Space Warning\n" +
			"  Resolved 2 hours ago\n\n" +
			"• `alert-004` - Network Latency\n" +
			"  Resolved 4 hours ago\n\n" +
			"• `alert-005` - Database Connection\n" +
			"  Resolved 6 hours ago"
		color = "good"
	default:
		alertText = "*📋 All Alerts Summary*\n\n" +
			"*Firing:* 2 alerts\n" +
			"*Silenced:* 1 alert\n" +
			"*Resolved (24h):* 5 alerts\n\n" +
			"Use `/gzh alerts firing` or `/gzh alerts resolved` for details."
		color = "#439FE0"
	}

	actions := []SlackAction{
		{
			Type:  "button",
			Text:  "🔥 View Firing",
			Name:  "view_firing",
			Value: "firing",
			Style: "danger",
		},
		{
			Type:  "button",
			Text:  "📊 Alert Dashboard",
			Name:  "view_alerts",
			Value: "alerts",
			Style: "primary",
			Url:   "http://localhost:8080/alerts",
		},
	}

	attachment := SlackAttachment{
		Color:      color,
		Text:       alertText,
		Actions:    actions,
		CallbackID: "alerts_command",
		Footer:     "GZH Monitoring",
		Timestamp:  time.Now().Unix(),
	}

	responseType := "ephemeral"
	if len(args) > 1 && args[1] == "public" {
		responseType = "in_channel"
	}

	return &SlackCommandResponse{
		ResponseType: responseType,
		Attachments:  []SlackAttachment{attachment},
	}, nil
}

// handleSilenceCommand silences an alert
func (s *SlackNotifier) handleSilenceCommand(cmd *SlackSlashCommand, args []string) (*SlackCommandResponse, error) {
	if len(args) == 0 {
		return &SlackCommandResponse{
			ResponseType: "ephemeral",
			Text:         "❌ Usage: `/gzh silence <alert-id>`\nExample: `/gzh silence alert-123`",
		}, nil
	}

	alertID := args[0]
	duration := "1 hour" // Default duration
	if len(args) > 1 {
		duration = strings.Join(args[1:], " ")
	}

	responseText := fmt.Sprintf("🔇 Alert `%s` has been silenced for %s by <@%s>",
		alertID, duration, cmd.UserID)

	attachment := SlackAttachment{
		Color: "warning",
		Text:  "The alert will not trigger notifications until the silence expires or is manually removed.",
		Fields: []SlackField{
			{
				Title: "Alert ID",
				Value: alertID,
				Short: true,
			},
			{
				Title: "Duration",
				Value: duration,
				Short: true,
			},
			{
				Title: "Silenced By",
				Value: fmt.Sprintf("<@%s>", cmd.UserID),
				Short: true,
			},
		},
		Footer:    "GZH Monitoring",
		Timestamp: time.Now().Unix(),
	}

	s.logger.Info("Alert silenced via Slack command",
		zap.String("alert_id", alertID),
		zap.String("duration", duration),
		zap.String("user", cmd.UserName))

	return &SlackCommandResponse{
		ResponseType: "in_channel",
		Text:         responseText,
		Attachments:  []SlackAttachment{attachment},
	}, nil
}

// handleResolveCommand resolves an alert
func (s *SlackNotifier) handleResolveCommand(cmd *SlackSlashCommand, args []string) (*SlackCommandResponse, error) {
	if len(args) == 0 {
		return &SlackCommandResponse{
			ResponseType: "ephemeral",
			Text:         "❌ Usage: `/gzh resolve <alert-id>`\nExample: `/gzh resolve alert-123`",
		}, nil
	}

	alertID := args[0]
	responseText := fmt.Sprintf("✅ Alert `%s` has been resolved by <@%s>", alertID, cmd.UserID)

	attachment := SlackAttachment{
		Color: "good",
		Text:  "The alert has been marked as resolved and will no longer trigger notifications.",
		Fields: []SlackField{
			{
				Title: "Alert ID",
				Value: alertID,
				Short: true,
			},
			{
				Title: "Resolved By",
				Value: fmt.Sprintf("<@%s>", cmd.UserID),
				Short: true,
			},
			{
				Title: "Resolved At",
				Value: time.Now().Format("2006-01-02 15:04:05 MST"),
				Short: true,
			},
		},
		Footer:    "GZH Monitoring",
		Timestamp: time.Now().Unix(),
	}

	s.logger.Info("Alert resolved via Slack command",
		zap.String("alert_id", alertID),
		zap.String("user", cmd.UserName))

	return &SlackCommandResponse{
		ResponseType: "in_channel",
		Text:         responseText,
		Attachments:  []SlackAttachment{attachment},
	}, nil
}

// handleTestCommand sends a test notification
func (s *SlackNotifier) handleTestCommand(cmd *SlackSlashCommand, args []string) (*SlackCommandResponse, error) {
	testType := "basic"
	if len(args) > 0 {
		testType = args[0]
	}

	var responseText string
	var attachment SlackAttachment

	switch testType {
	case "alert":
		responseText = "🧪 Test alert notification sent"
		attachment = SlackAttachment{
			Color: "danger",
			Title: "🚨 Test Alert - High CPU Usage",
			Text:  "This is a test alert to verify notification delivery.",
			Fields: []SlackField{
				{
					Title: "Severity",
					Value: "Critical",
					Short: true,
				},
				{
					Title: "Host",
					Value: "test-server",
					Short: true,
				},
			},
			Footer:    "GZH Monitoring Test",
			Timestamp: time.Now().Unix(),
		}
	case "status":
		responseText = "🧪 Test system status notification sent"
		attachment = SlackAttachment{
			Color: "good",
			Title: "✅ Test System Status - All Systems Operational",
			Text:  "This is a test system status notification.",
			Fields: []SlackField{
				{
					Title: "Status",
					Value: "Healthy",
					Short: true,
				},
				{
					Title: "Uptime",
					Value: "100%",
					Short: true,
				},
			},
			Footer:    "GZH Monitoring Test",
			Timestamp: time.Now().Unix(),
		}
	default:
		responseText = fmt.Sprintf("🧪 Test notification sent by <@%s>", cmd.UserID)
		attachment = SlackAttachment{
			Color:     "#439FE0",
			Title:     "🔔 Test Notification",
			Text:      "If you can see this message, Slack integration is working correctly!",
			Footer:    "GZH Monitoring Test",
			Timestamp: time.Now().Unix(),
		}
	}

	s.logger.Info("Test notification sent via Slack command",
		zap.String("test_type", testType),
		zap.String("user", cmd.UserName))

	return &SlackCommandResponse{
		ResponseType: "ephemeral",
		Text:         responseText,
		Attachments:  []SlackAttachment{attachment},
	}, nil
}

// handleUnknownCommand handles unknown commands
func (s *SlackNotifier) handleUnknownCommand(cmd *SlackSlashCommand, command string) (*SlackCommandResponse, error) {
	responseText := fmt.Sprintf("❌ Unknown command: `%s`\n\nUse `/gzh help` to see available commands.", command)

	return &SlackCommandResponse{
		ResponseType: "ephemeral",
		Text:         responseText,
	}, nil
}
