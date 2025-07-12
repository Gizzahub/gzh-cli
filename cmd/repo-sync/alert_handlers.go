package reposync

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
)

// ConsoleAlertHandler prints alerts to console
type ConsoleAlertHandler struct {
	logger *zap.Logger
}

// NewConsoleAlertHandler creates a new console alert handler
func NewConsoleAlertHandler(logger *zap.Logger) *ConsoleAlertHandler {
	return &ConsoleAlertHandler{logger: logger}
}

// HandleAlert handles alert by printing to console
func (c *ConsoleAlertHandler) HandleAlert(ctx context.Context, alert *QualityAlert) error {
	// Print alert with color coding based on severity
	var severityPrefix string
	switch alert.Severity {
	case AlertSeverityCritical:
		severityPrefix = "🚨 CRITICAL"
	case AlertSeverityHigh:
		severityPrefix = "⚠️  HIGH"
	case AlertSeverityMedium:
		severityPrefix = "📢 MEDIUM"
	case AlertSeverityLow:
		severityPrefix = "ℹ️  LOW"
	}

	fmt.Printf("\n%s ALERT: %s\n", severityPrefix, alert.Type)
	fmt.Printf("Repository: %s\n", alert.Repository)
	fmt.Printf("Message: %s\n", alert.Message)
	fmt.Printf("Time: %s\n", alert.Timestamp.Format("2006-01-02 15:04:05"))

	if len(alert.Details) > 0 {
		fmt.Println("Details:")
		for key, value := range alert.Details {
			fmt.Printf("  %s: %v\n", key, value)
		}
	}

	if len(alert.Suggestions) > 0 {
		fmt.Println("Suggestions:")
		for _, suggestion := range alert.Suggestions {
			fmt.Printf("  • %s\n", suggestion)
		}
	}

	fmt.Println(strings.Repeat("-", 60))
	return nil
}

// FileAlertHandler saves alerts to files
type FileAlertHandler struct {
	logger    *zap.Logger
	outputDir string
}

// NewFileAlertHandler creates a new file alert handler
func NewFileAlertHandler(logger *zap.Logger, outputDir string) *FileAlertHandler {
	return &FileAlertHandler{
		logger:    logger,
		outputDir: outputDir,
	}
}

// HandleAlert handles alert by saving to file
func (f *FileAlertHandler) HandleAlert(ctx context.Context, alert *QualityAlert) error {
	// Ensure output directory exists
	alertDir := filepath.Join(f.outputDir, "alerts", alert.Repository)
	if err := os.MkdirAll(alertDir, 0755); err != nil {
		return fmt.Errorf("failed to create alert directory: %w", err)
	}

	// Create filename with timestamp and severity
	filename := fmt.Sprintf("alert-%s-%s-%s.json",
		alert.Severity,
		alert.Type,
		alert.Timestamp.Format("20060102-150405"))

	// Marshal alert to JSON
	data, err := json.MarshalIndent(alert, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal alert: %w", err)
	}

	// Write to file
	filePath := filepath.Join(alertDir, filename)
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write alert file: %w", err)
	}

	f.logger.Info("Alert saved to file",
		zap.String("path", filePath),
		zap.String("alert_id", alert.ID))

	return nil
}

// WebhookAlertHandler sends alerts to webhook endpoints
type WebhookAlertHandler struct {
	logger     *zap.Logger
	webhookURL string
	httpClient *http.Client
}

// NewWebhookAlertHandler creates a new webhook alert handler
func NewWebhookAlertHandler(logger *zap.Logger, webhookURL string) *WebhookAlertHandler {
	return &WebhookAlertHandler{
		logger:     logger,
		webhookURL: webhookURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// HandleAlert handles alert by sending to webhook
func (w *WebhookAlertHandler) HandleAlert(ctx context.Context, alert *QualityAlert) error {
	// Create webhook payload
	payload := map[string]interface{}{
		"alert":      alert,
		"timestamp":  time.Now().Unix(),
		"event_type": "quality_alert",
	}

	// Marshal payload
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.webhookURL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := w.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned error status: %d", resp.StatusCode)
	}

	w.logger.Info("Alert sent to webhook",
		zap.String("alert_id", alert.ID),
		zap.String("webhook_url", w.webhookURL))

	return nil
}

// EmailAlertHandler sends alerts via email
type EmailAlertHandler struct {
	logger     *zap.Logger
	smtpHost   string
	smtpPort   string
	username   string
	password   string
	from       string
	recipients []string
}

// EmailConfig represents email configuration
type EmailConfig struct {
	SMTPHost   string   `json:"smtp_host"`
	SMTPPort   string   `json:"smtp_port"`
	Username   string   `json:"username"`
	Password   string   `json:"password"`
	From       string   `json:"from"`
	Recipients []string `json:"recipients"`
}

// NewEmailAlertHandler creates a new email alert handler
func NewEmailAlertHandler(logger *zap.Logger, config EmailConfig) *EmailAlertHandler {
	return &EmailAlertHandler{
		logger:     logger,
		smtpHost:   config.SMTPHost,
		smtpPort:   config.SMTPPort,
		username:   config.Username,
		password:   config.Password,
		from:       config.From,
		recipients: config.Recipients,
	}
}

// HandleAlert handles alert by sending email
func (e *EmailAlertHandler) HandleAlert(ctx context.Context, alert *QualityAlert) error {
	// Create email subject
	subject := fmt.Sprintf("[%s] Quality Alert: %s - %s",
		strings.ToUpper(string(alert.Severity)),
		alert.Repository,
		alert.Type)

	// Create email body
	body := e.formatEmailBody(alert)

	// Create message
	message := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s\r\n",
		e.from,
		strings.Join(e.recipients, ","),
		subject,
		body)

	// Send email
	auth := smtp.PlainAuth("", e.username, e.password, e.smtpHost)
	err := smtp.SendMail(
		e.smtpHost+":"+e.smtpPort,
		auth,
		e.from,
		e.recipients,
		[]byte(message),
	)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	e.logger.Info("Alert sent via email",
		zap.String("alert_id", alert.ID),
		zap.Strings("recipients", e.recipients))

	return nil
}

// formatEmailBody formats alert as HTML email
func (e *EmailAlertHandler) formatEmailBody(alert *QualityAlert) string {
	var sb strings.Builder

	// HTML header
	sb.WriteString(`<html><body style="font-family: Arial, sans-serif;">`)

	// Alert header with severity color
	severityColor := e.getSeverityColor(alert.Severity)
	sb.WriteString(fmt.Sprintf(`<div style="background-color: %s; color: white; padding: 20px; border-radius: 5px;">`, severityColor))
	sb.WriteString(fmt.Sprintf(`<h2 style="margin: 0;">%s Quality Alert</h2>`, strings.ToUpper(string(alert.Severity))))
	sb.WriteString(`</div>`)

	// Alert details
	sb.WriteString(`<div style="padding: 20px;">`)
	sb.WriteString(fmt.Sprintf(`<p><strong>Repository:</strong> %s</p>`, alert.Repository))
	sb.WriteString(fmt.Sprintf(`<p><strong>Alert Type:</strong> %s</p>`, alert.Type))
	sb.WriteString(fmt.Sprintf(`<p><strong>Time:</strong> %s</p>`, alert.Timestamp.Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf(`<p><strong>Message:</strong> %s</p>`, alert.Message))

	// Details section
	if len(alert.Details) > 0 {
		sb.WriteString(`<h3>Details:</h3><ul>`)
		for key, value := range alert.Details {
			sb.WriteString(fmt.Sprintf(`<li><strong>%s:</strong> %v</li>`, key, value))
		}
		sb.WriteString(`</ul>`)
	}

	// Suggestions section
	if len(alert.Suggestions) > 0 {
		sb.WriteString(`<h3>Suggestions:</h3><ul>`)
		for _, suggestion := range alert.Suggestions {
			sb.WriteString(fmt.Sprintf(`<li>%s</li>`, suggestion))
		}
		sb.WriteString(`</ul>`)
	}

	sb.WriteString(`</div></body></html>`)
	return sb.String()
}

// getSeverityColor returns color for severity level
func (e *EmailAlertHandler) getSeverityColor(severity AlertSeverity) string {
	switch severity {
	case AlertSeverityCritical:
		return "#dc3545"
	case AlertSeverityHigh:
		return "#fd7e14"
	case AlertSeverityMedium:
		return "#ffc107"
	case AlertSeverityLow:
		return "#17a2b8"
	default:
		return "#6c757d"
	}
}

// SlackAlertHandler sends alerts to Slack
type SlackAlertHandler struct {
	logger     *zap.Logger
	webhookURL string
	channel    string
	username   string
}

// NewSlackAlertHandler creates a new Slack alert handler
func NewSlackAlertHandler(logger *zap.Logger, webhookURL, channel, username string) *SlackAlertHandler {
	return &SlackAlertHandler{
		logger:     logger,
		webhookURL: webhookURL,
		channel:    channel,
		username:   username,
	}
}

// HandleAlert handles alert by sending to Slack
func (s *SlackAlertHandler) HandleAlert(ctx context.Context, alert *QualityAlert) error {
	// Create Slack attachment
	attachment := map[string]interface{}{
		"color":  s.getSlackColor(alert.Severity),
		"title":  fmt.Sprintf("%s Alert: %s", alert.Severity, alert.Type),
		"text":   alert.Message,
		"footer": alert.Repository,
		"ts":     alert.Timestamp.Unix(),
		"fields": s.createSlackFields(alert),
	}

	// Create Slack payload
	payload := map[string]interface{}{
		"channel":     s.channel,
		"username":    s.username,
		"attachments": []interface{}{attachment},
	}

	// Send to Slack
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal Slack payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.webhookURL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create Slack request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send Slack webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Slack webhook returned status: %d", resp.StatusCode)
	}

	s.logger.Info("Alert sent to Slack",
		zap.String("alert_id", alert.ID),
		zap.String("channel", s.channel))

	return nil
}

// getSlackColor returns Slack color for severity
func (s *SlackAlertHandler) getSlackColor(severity AlertSeverity) string {
	switch severity {
	case AlertSeverityCritical:
		return "danger"
	case AlertSeverityHigh:
		return "warning"
	case AlertSeverityMedium:
		return "#FFA500"
	case AlertSeverityLow:
		return "good"
	default:
		return "#808080"
	}
}

// createSlackFields creates fields for Slack attachment
func (s *SlackAlertHandler) createSlackFields(alert *QualityAlert) []map[string]interface{} {
	fields := make([]map[string]interface{}, 0)

	// Add key details as fields
	for key, value := range alert.Details {
		fields = append(fields, map[string]interface{}{
			"title": key,
			"value": fmt.Sprintf("%v", value),
			"short": true,
		})
	}

	// Add suggestions as a single field
	if len(alert.Suggestions) > 0 {
		fields = append(fields, map[string]interface{}{
			"title": "Suggestions",
			"value": strings.Join(alert.Suggestions, "\n• "),
			"short": false,
		})
	}

	return fields
}

// CompositeAlertHandler combines multiple alert handlers
type CompositeAlertHandler struct {
	handlers []AlertHandler
}

// NewCompositeAlertHandler creates a new composite alert handler
func NewCompositeAlertHandler(handlers ...AlertHandler) *CompositeAlertHandler {
	return &CompositeAlertHandler{handlers: handlers}
}

// HandleAlert handles alert by calling all registered handlers
func (c *CompositeAlertHandler) HandleAlert(ctx context.Context, alert *QualityAlert) error {
	var errs []error

	for _, handler := range c.handlers {
		if err := handler.HandleAlert(ctx, alert); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("composite handler errors: %v", errs)
	}

	return nil
}
