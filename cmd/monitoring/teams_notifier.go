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

// TeamsNotifier handles Microsoft Teams webhook notifications
type TeamsNotifier struct {
	webhookURL   string
	httpClient   *http.Client
	logger       *zap.Logger
	graphConfig  *TeamsGraphConfig
	channelRules []ChannelRule
}

// TeamsConfig represents Teams notification configuration
type TeamsConfig struct {
	WebhookURL   string            `json:"webhook_url"`
	Enabled      bool              `json:"enabled"`
	GraphConfig  *TeamsGraphConfig `json:"graph_config,omitempty"`
	ChannelRules []ChannelRule     `json:"channel_rules,omitempty"`
}

// TeamsGraphConfig represents Microsoft Graph API configuration
type TeamsGraphConfig struct {
	TenantID     string `json:"tenant_id"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	DefaultTeam  string `json:"default_team_id"`
}

// ChannelRule defines routing rules for different notification types
type ChannelRule struct {
	Severity    AlertSeverity `json:"severity"`
	Type        string        `json:"type"` // "alert", "system", "custom"
	TeamID      string        `json:"team_id"`
	ChannelID   string        `json:"channel_id"`
	ChannelName string        `json:"channel_name"`
}

// TeamsMessage represents a Teams message payload with Adaptive Cards
type TeamsMessage struct {
	Type            string                `json:"type"`
	Attachments     []TeamsAdaptiveCard   `json:"attachments"`
	Summary         string                `json:"summary,omitempty"`
	ThemeColor      string                `json:"themeColor,omitempty"`
	Sections        []TeamsMessageSection `json:"sections,omitempty"`
	PotentialAction []TeamsAction         `json:"potentialAction,omitempty"`
}

// TeamsAdaptiveCard represents an Adaptive Card attachment
type TeamsAdaptiveCard struct {
	ContentType string           `json:"contentType"`
	Content     TeamsCardContent `json:"content"`
}

// TeamsCardContent represents the content of an Adaptive Card
type TeamsCardContent struct {
	Schema  string            `json:"$schema"`
	Type    string            `json:"type"`
	Version string            `json:"version"`
	Body    []TeamsCardBody   `json:"body"`
	Actions []TeamsCardAction `json:"actions,omitempty"`
}

// TeamsCardBody represents body elements in an Adaptive Card
type TeamsCardBody struct {
	Type    string            `json:"type"`
	Text    string            `json:"text,omitempty"`
	Size    string            `json:"size,omitempty"`
	Weight  string            `json:"weight,omitempty"`
	Color   string            `json:"color,omitempty"`
	Wrap    bool              `json:"wrap,omitempty"`
	Spacing string            `json:"spacing,omitempty"`
	Columns []TeamsCardColumn `json:"columns,omitempty"`
	Items   []TeamsCardBody   `json:"items,omitempty"`
	Facts   []TeamsCardFact   `json:"facts,omitempty"`
}

// TeamsCardColumn represents a column in an Adaptive Card
type TeamsCardColumn struct {
	Type  string          `json:"type"`
	Width string          `json:"width,omitempty"`
	Items []TeamsCardBody `json:"items"`
}

// TeamsCardFact represents a fact in an Adaptive Card FactSet
type TeamsCardFact struct {
	Title string `json:"title"`
	Value string `json:"value"`
}

// TeamsCardAction represents an action in an Adaptive Card
type TeamsCardAction struct {
	Type  string `json:"type"`
	Title string `json:"title"`
	URL   string `json:"url,omitempty"`
	Data  string `json:"data,omitempty"`
}

// TeamsMessageSection represents a section in a Teams message (legacy format)
type TeamsMessageSection struct {
	ActivityTitle    string        `json:"activityTitle,omitempty"`
	ActivitySubtitle string        `json:"activitySubtitle,omitempty"`
	ActivityImage    string        `json:"activityImage,omitempty"`
	Text             string        `json:"text,omitempty"`
	Facts            []TeamsFact   `json:"facts,omitempty"`
	PotentialAction  []TeamsAction `json:"potentialAction,omitempty"`
}

// TeamsFact represents a fact in a Teams message section
type TeamsFact struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// TeamsAction represents an action in a Teams message
type TeamsAction struct {
	Type    string              `json:"@type"`
	Name    string              `json:"name"`
	Targets []TeamsActionTarget `json:"targets,omitempty"`
}

// TeamsActionTarget represents a target for a Teams action
type TeamsActionTarget struct {
	OS  string `json:"os"`
	URI string `json:"uri"`
}

// Microsoft Graph API structures for channel integration

// GraphToken represents OAuth token for Graph API
type GraphToken struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	ExpiresAt   time.Time
}

// GraphTeam represents a Microsoft Teams team
type GraphTeam struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
}

// GraphChannel represents a Microsoft Teams channel
type GraphChannel struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
	WebURL      string `json:"webUrl"`
}

// GraphChannelMessage represents a message to be sent to a channel
type GraphChannelMessage struct {
	Body struct {
		ContentType string `json:"contentType"`
		Content     string `json:"content"`
	} `json:"body"`
	Attachments []GraphMessageAttachment `json:"attachments,omitempty"`
}

// GraphMessageAttachment represents an attachment in a Graph API message
type GraphMessageAttachment struct {
	ID          string      `json:"id"`
	ContentType string      `json:"contentType"`
	ContentURL  string      `json:"contentUrl,omitempty"`
	Content     interface{} `json:"content,omitempty"`
	Name        string      `json:"name,omitempty"`
}

// NewTeamsNotifier creates a new Teams notifier
func NewTeamsNotifier(config *TeamsConfig, logger *zap.Logger) *TeamsNotifier {
	return &TeamsNotifier{
		webhookURL: config.WebhookURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger:       logger,
		graphConfig:  config.GraphConfig,
		channelRules: config.ChannelRules,
	}
}

// SendAlert sends an alert notification to Teams
func (t *TeamsNotifier) SendAlert(ctx context.Context, alert *AlertInstance) error {
	if t.webhookURL == "" {
		return fmt.Errorf("Teams webhook URL not configured")
	}

	message := t.formatAlertMessage(alert)
	return t.sendMessage(ctx, message)
}

// SendSystemStatus sends a system status notification to Teams
func (t *TeamsNotifier) SendSystemStatus(ctx context.Context, status *SystemStatus) error {
	if t.webhookURL == "" {
		return fmt.Errorf("Teams webhook URL not configured")
	}

	message := t.formatSystemStatusMessage(status)
	return t.sendMessage(ctx, message)
}

// SendCustomMessage sends a custom message to Teams
func (t *TeamsNotifier) SendCustomMessage(ctx context.Context, title, text string, severity AlertSeverity) error {
	if t.webhookURL == "" {
		return fmt.Errorf("Teams webhook URL not configured")
	}

	message := t.formatCustomMessage(title, text, severity)
	return t.sendMessage(ctx, message)
}

// TestConnection tests the Teams webhook connection
func (t *TeamsNotifier) TestConnection(ctx context.Context) error {
	if t.webhookURL == "" {
		return fmt.Errorf("Teams webhook URL not configured")
	}

	message := t.formatTestMessage()
	return t.sendMessage(ctx, message)
}

// formatAlertMessage formats an alert as a Teams Adaptive Card message
func (t *TeamsNotifier) formatAlertMessage(alert *AlertInstance) *TeamsMessage {
	themeColor := t.getSeverityColor(AlertSeverity(alert.Severity))
	statusEmoji := t.getStatusEmoji(alert.Status)

	title := fmt.Sprintf("%s %s", statusEmoji, alert.RuleName)

	var statusText string
	if alert.Status == AlertStatusFiring {
		statusText = "🚨 **Alert is firing**"
	} else if alert.Status == AlertStatusResolved {
		statusText = "✅ **Alert resolved**"
	}

	// Create facts for the alert details
	facts := []TeamsCardFact{
		{Title: "Severity", Value: string(alert.Severity)},
		{Title: "Status", Value: string(alert.Status)},
	}

	if alert.FiredAt != nil {
		facts = append(facts, TeamsCardFact{
			Title: "Fired At",
			Value: alert.FiredAt.Format("2006-01-02 15:04:05 MST"),
		})
	}

	if alert.ResolvedAt != nil {
		facts = append(facts, TeamsCardFact{
			Title: "Resolved At",
			Value: alert.ResolvedAt.Format("2006-01-02 15:04:05 MST"),
		})
	}

	// Add labels as facts
	for key, value := range alert.Labels {
		facts = append(facts, TeamsCardFact{
			Title: strings.Title(key),
			Value: value,
		})
	}

	cardContent := TeamsCardContent{
		Schema:  "http://adaptivecards.io/schemas/adaptive-card.json",
		Type:    "AdaptiveCard",
		Version: "1.2",
		Body: []TeamsCardBody{
			{
				Type:   "TextBlock",
				Text:   title,
				Size:   "Medium",
				Weight: "Bolder",
				Color:  t.getCardTextColor(alert.Status),
				Wrap:   true,
			},
			{
				Type:    "TextBlock",
				Text:    statusText,
				Wrap:    true,
				Spacing: "Small",
			},
			{
				Type:    "TextBlock",
				Text:    alert.Message,
				Wrap:    true,
				Spacing: "Medium",
			},
			{
				Type:    "FactSet",
				Facts:   facts,
				Spacing: "Medium",
			},
		},
		Actions: []TeamsCardAction{
			{
				Type:  "Action.OpenUrl",
				Title: "View Details",
				URL:   fmt.Sprintf("http://localhost:8080/alerts/%s", alert.ID),
			},
			{
				Type:  "Action.OpenUrl",
				Title: "Dashboard",
				URL:   "http://localhost:8080/dashboard",
			},
		},
	}

	return &TeamsMessage{
		Type: "message",
		Attachments: []TeamsAdaptiveCard{
			{
				ContentType: "application/vnd.microsoft.card.adaptive",
				Content:     cardContent,
			},
		},
		Summary:    fmt.Sprintf("Alert: %s", alert.RuleName),
		ThemeColor: themeColor,
	}
}

// formatSystemStatusMessage formats system status as a Teams Adaptive Card message
func (t *TeamsNotifier) formatSystemStatusMessage(status *SystemStatus) *TeamsMessage {
	themeColor := t.getSystemStatusColor(status.Status)
	statusEmoji := t.getSystemStatusEmoji(status.Status)

	title := fmt.Sprintf("%s System Status: %s", statusEmoji, strings.ToUpper(status.Status))

	facts := []TeamsCardFact{
		{Title: "Uptime", Value: status.Uptime},
		{Title: "Active Tasks", Value: fmt.Sprintf("%d", status.ActiveTasks)},
		{Title: "Memory Usage", Value: formatBytes(status.MemoryUsage)},
		{Title: "CPU Usage", Value: fmt.Sprintf("%.1f%%", status.CPUUsage)},
		{Title: "Total Requests", Value: fmt.Sprintf("%d", status.TotalRequests)},
	}

	if status.DiskUsage > 0 {
		facts = append(facts, TeamsCardFact{
			Title: "Disk Usage",
			Value: fmt.Sprintf("%.1f%%", status.DiskUsage),
		})
	}

	cardContent := TeamsCardContent{
		Schema:  "http://adaptivecards.io/schemas/adaptive-card.json",
		Type:    "AdaptiveCard",
		Version: "1.2",
		Body: []TeamsCardBody{
			{
				Type:   "TextBlock",
				Text:   title,
				Size:   "Medium",
				Weight: "Bolder",
				Color:  t.getSystemStatusTextColor(status.Status),
				Wrap:   true,
			},
			{
				Type:    "FactSet",
				Facts:   facts,
				Spacing: "Medium",
			},
		},
		Actions: []TeamsCardAction{
			{
				Type:  "Action.OpenUrl",
				Title: "Full Dashboard",
				URL:   "http://localhost:8080/dashboard",
			},
			{
				Type:  "Action.OpenUrl",
				Title: "View Metrics",
				URL:   "http://localhost:8080/metrics",
			},
		},
	}

	return &TeamsMessage{
		Type: "message",
		Attachments: []TeamsAdaptiveCard{
			{
				ContentType: "application/vnd.microsoft.card.adaptive",
				Content:     cardContent,
			},
		},
		Summary:    "System Status Update",
		ThemeColor: themeColor,
	}
}

// formatCustomMessage formats a custom message as a Teams Adaptive Card
func (t *TeamsNotifier) formatCustomMessage(title, text string, severity AlertSeverity) *TeamsMessage {
	themeColor := t.getSeverityColor(severity)

	cardContent := TeamsCardContent{
		Schema:  "http://adaptivecards.io/schemas/adaptive-card.json",
		Type:    "AdaptiveCard",
		Version: "1.2",
		Body: []TeamsCardBody{
			{
				Type:   "TextBlock",
				Text:   title,
				Size:   "Medium",
				Weight: "Bolder",
				Wrap:   true,
			},
			{
				Type:    "TextBlock",
				Text:    text,
				Wrap:    true,
				Spacing: "Medium",
			},
		},
		Actions: []TeamsCardAction{
			{
				Type:  "Action.OpenUrl",
				Title: "Dashboard",
				URL:   "http://localhost:8080/dashboard",
			},
		},
	}

	return &TeamsMessage{
		Type: "message",
		Attachments: []TeamsAdaptiveCard{
			{
				ContentType: "application/vnd.microsoft.card.adaptive",
				Content:     cardContent,
			},
		},
		Summary:    title,
		ThemeColor: themeColor,
	}
}

// formatTestMessage formats a test message for Teams
func (t *TeamsNotifier) formatTestMessage() *TeamsMessage {
	cardContent := TeamsCardContent{
		Schema:  "http://adaptivecards.io/schemas/adaptive-card.json",
		Type:    "AdaptiveCard",
		Version: "1.2",
		Body: []TeamsCardBody{
			{
				Type:   "TextBlock",
				Text:   "🧪 Test Message from GZH Monitoring",
				Size:   "Medium",
				Weight: "Bolder",
				Color:  "Good",
				Wrap:   true,
			},
			{
				Type:    "TextBlock",
				Text:    "If you can see this message, the Teams integration is working correctly!",
				Wrap:    true,
				Spacing: "Medium",
			},
			{
				Type: "FactSet",
				Facts: []TeamsCardFact{
					{Title: "Test Time", Value: time.Now().Format("2006-01-02 15:04:05 MST")},
					{Title: "Status", Value: "✅ Success"},
				},
				Spacing: "Medium",
			},
		},
		Actions: []TeamsCardAction{
			{
				Type:  "Action.OpenUrl",
				Title: "View Dashboard",
				URL:   "http://localhost:8080/dashboard",
			},
		},
	}

	return &TeamsMessage{
		Type: "message",
		Attachments: []TeamsAdaptiveCard{
			{
				ContentType: "application/vnd.microsoft.card.adaptive",
				Content:     cardContent,
			},
		},
		Summary:    "GZH Monitoring Test",
		ThemeColor: "0078D4", // Teams blue
	}
}

// sendMessage sends a message to Teams webhook
func (t *TeamsNotifier) sendMessage(ctx context.Context, message *TeamsMessage) error {
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal Teams message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", t.webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Teams webhook returned status code: %d", resp.StatusCode)
	}

	t.logger.Info("Teams message sent successfully",
		zap.Int("status_code", resp.StatusCode))

	return nil
}

// Helper methods for colors and emojis

func (t *TeamsNotifier) getSeverityColor(severity AlertSeverity) string {
	switch severity {
	case AlertSeverityCritical:
		return "FF0000" // Red
	case AlertSeverityHigh:
		return "FF6600" // Orange
	case AlertSeverityMedium:
		return "FFFF00" // Yellow
	case AlertSeverityLow:
		return "00FF00" // Green
	case AlertSeverityInfo:
		return "0078D4" // Teams blue
	default:
		return "808080" // Gray
	}
}

func (t *TeamsNotifier) getSystemStatusColor(status string) string {
	switch status {
	case "healthy":
		return "00FF00" // Green
	case "warning":
		return "FFFF00" // Yellow
	case "critical":
		return "FF0000" // Red
	default:
		return "0078D4" // Teams blue
	}
}

func (t *TeamsNotifier) getStatusEmoji(status AlertStatus) string {
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

func (t *TeamsNotifier) getSystemStatusEmoji(status string) string {
	switch status {
	case "healthy":
		return "✅"
	case "warning":
		return "⚠️"
	case "critical":
		return "🚨"
	default:
		return "ℹ️"
	}
}

func (t *TeamsNotifier) getCardTextColor(status AlertStatus) string {
	switch status {
	case AlertStatusFiring:
		return "Attention"
	case AlertStatusResolved:
		return "Good"
	case AlertStatusSilenced:
		return "Warning"
	default:
		return "Default"
	}
}

func (t *TeamsNotifier) getSystemStatusTextColor(status string) string {
	switch status {
	case "healthy":
		return "Good"
	case "warning":
		return "Warning"
	case "critical":
		return "Attention"
	default:
		return "Default"
	}
}

// Microsoft Graph API integration for Teams channel management

// getGraphToken obtains an OAuth token for Microsoft Graph API
func (t *TeamsNotifier) getGraphToken(ctx context.Context) (*GraphToken, error) {
	if t.graphConfig == nil {
		return nil, fmt.Errorf("Graph API configuration not provided")
	}

	tokenURL := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", t.graphConfig.TenantID)

	data := fmt.Sprintf("client_id=%s&scope=https://graph.microsoft.com/.default&client_secret=%s&grant_type=client_credentials",
		t.graphConfig.ClientID, t.graphConfig.ClientSecret)

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to request token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token request failed with status: %d", resp.StatusCode)
	}

	var token GraphToken
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	token.ExpiresAt = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
	return &token, nil
}

// ListTeams retrieves all Teams that the application has access to
func (t *TeamsNotifier) ListTeams(ctx context.Context) ([]GraphTeam, error) {
	token, err := t.getGraphToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", "https://graph.microsoft.com/v1.0/groups?$filter=resourceProvisioningOptions/Any(x:x eq 'Team')", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list teams: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("teams request failed with status: %d", resp.StatusCode)
	}

	var result struct {
		Value []GraphTeam `json:"value"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode teams response: %w", err)
	}

	return result.Value, nil
}

// ListChannels retrieves all channels for a specific team
func (t *TeamsNotifier) ListChannels(ctx context.Context, teamID string) ([]GraphChannel, error) {
	token, err := t.getGraphToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	url := fmt.Sprintf("https://graph.microsoft.com/v1.0/teams/%s/channels", teamID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list channels: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("channels request failed with status: %d", resp.StatusCode)
	}

	var result struct {
		Value []GraphChannel `json:"value"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode channels response: %w", err)
	}

	return result.Value, nil
}

// SendToChannel sends a message with Adaptive Card to a specific Teams channel
func (t *TeamsNotifier) SendToChannel(ctx context.Context, teamID, channelID string, card *TeamsCardContent) error {
	token, err := t.getGraphToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to get access token: %w", err)
	}

	// Create Graph message with Adaptive Card attachment
	message := GraphChannelMessage{
		Body: struct {
			ContentType string `json:"contentType"`
			Content     string `json:"content"`
		}{
			ContentType: "html",
			Content:     "<attachment id=\"adaptive-card\"></attachment>",
		},
		Attachments: []GraphMessageAttachment{
			{
				ID:          "adaptive-card",
				ContentType: "application/vnd.microsoft.card.adaptive",
				Content:     card,
			},
		},
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	url := fmt.Sprintf("https://graph.microsoft.com/v1.0/teams/%s/channels/%s/messages", teamID, channelID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send message to channel: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("channel message request failed with status: %d", resp.StatusCode)
	}

	t.logger.Info("Message sent to Teams channel successfully",
		zap.String("team_id", teamID),
		zap.String("channel_id", channelID))

	return nil
}

// SendAlertToChannel sends an alert to appropriate channel based on routing rules
func (t *TeamsNotifier) SendAlertToChannel(ctx context.Context, alert *AlertInstance) error {
	if t.graphConfig == nil {
		return fmt.Errorf("Graph API not configured for channel integration")
	}

	// Find matching channel rule
	rule := t.findChannelRule("alert", AlertSeverity(alert.Severity))
	if rule == nil {
		// Use default team if no rule matches
		if t.graphConfig.DefaultTeam == "" {
			return fmt.Errorf("no channel rule found and no default team configured")
		}
		// Send to general channel of default team
		channels, err := t.ListChannels(ctx, t.graphConfig.DefaultTeam)
		if err != nil {
			return fmt.Errorf("failed to list channels: %w", err)
		}

		var generalChannel *GraphChannel
		for _, ch := range channels {
			if strings.ToLower(ch.DisplayName) == "general" {
				generalChannel = &ch
				break
			}
		}

		if generalChannel == nil {
			return fmt.Errorf("general channel not found in default team")
		}

		rule = &ChannelRule{
			TeamID:    t.graphConfig.DefaultTeam,
			ChannelID: generalChannel.ID,
		}
	}

	// Format alert as adaptive card
	message := t.formatAlertMessage(alert)
	cardContent := &message.Attachments[0].Content

	return t.SendToChannel(ctx, rule.TeamID, rule.ChannelID, cardContent)
}

// SendSystemStatusToChannel sends system status to appropriate channel
func (t *TeamsNotifier) SendSystemStatusToChannel(ctx context.Context, status *SystemStatus) error {
	if t.graphConfig == nil {
		return fmt.Errorf("Graph API not configured for channel integration")
	}

	// Find matching channel rule for system status
	rule := t.findChannelRule("system", AlertSeverityInfo)
	if rule == nil && t.graphConfig.DefaultTeam != "" {
		// Use default team if no rule matches
		channels, err := t.ListChannels(ctx, t.graphConfig.DefaultTeam)
		if err != nil {
			return fmt.Errorf("failed to list channels: %w", err)
		}

		var generalChannel *GraphChannel
		for _, ch := range channels {
			if strings.ToLower(ch.DisplayName) == "general" {
				generalChannel = &ch
				break
			}
		}

		if generalChannel != nil {
			rule = &ChannelRule{
				TeamID:    t.graphConfig.DefaultTeam,
				ChannelID: generalChannel.ID,
			}
		}
	}

	if rule == nil {
		return fmt.Errorf("no channel rule found for system status")
	}

	// Format system status as adaptive card
	message := t.formatSystemStatusMessage(status)
	cardContent := &message.Attachments[0].Content

	return t.SendToChannel(ctx, rule.TeamID, rule.ChannelID, cardContent)
}

// findChannelRule finds the most appropriate channel rule for given type and severity
func (t *TeamsNotifier) findChannelRule(msgType string, severity AlertSeverity) *ChannelRule {
	var bestRule *ChannelRule

	for _, rule := range t.channelRules {
		if rule.Type == msgType {
			if rule.Severity == severity {
				// Exact match - return immediately
				return &rule
			}
			if bestRule == nil {
				// First matching type
				bestRule = &rule
			}
		}
	}

	return bestRule
}

// GetChannelConfiguration returns current channel configuration for management
func (t *TeamsNotifier) GetChannelConfiguration() []ChannelRule {
	return t.channelRules
}

// AddChannelRule adds a new channel routing rule
func (t *TeamsNotifier) AddChannelRule(rule ChannelRule) {
	t.channelRules = append(t.channelRules, rule)
}

// RemoveChannelRule removes a channel routing rule
func (t *TeamsNotifier) RemoveChannelRule(teamID, channelID string) {
	for i, rule := range t.channelRules {
		if rule.TeamID == teamID && rule.ChannelID == channelID {
			t.channelRules = append(t.channelRules[:i], t.channelRules[i+1:]...)
			break
		}
	}
}
