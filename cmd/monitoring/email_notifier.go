package monitoring

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"html/template"
	"net"
	"net/smtp"
	"strings"
	"time"

	"go.uber.org/zap"
)

// EmailNotifier handles email notifications
type EmailNotifier struct {
	config    *EmailConfig
	logger    *zap.Logger
	templates map[string]*template.Template
}

// EmailConfig represents email notification configuration
type EmailConfig struct {
	SMTPHost   string   `json:"smtp_host"`
	SMTPPort   int      `json:"smtp_port"`
	Username   string   `json:"username"`
	Password   string   `json:"password"`
	From       string   `json:"from"`
	Recipients []string `json:"recipients"`
	UseTLS     bool     `json:"use_tls"`
	Enabled    bool     `json:"enabled"`
}

// EmailMessage represents an email message
type EmailMessage struct {
	To      []string
	Subject string
	Body    string
}

// NewEmailNotifier creates a new email notifier
func NewEmailNotifier(config *EmailConfig, logger *zap.Logger) *EmailNotifier {
	notifier := &EmailNotifier{
		config:    config,
		logger:    logger,
		templates: make(map[string]*template.Template),
	}

	// Initialize email templates
	notifier.initializeTemplates()

	return notifier
}

// SendAlert sends an alert notification via email
func (e *EmailNotifier) SendAlert(ctx context.Context, alert *AlertInstance) error {
	if !e.config.Enabled || e.config.SMTPHost == "" {
		return fmt.Errorf("email notifications not configured")
	}

	subject := e.formatAlertSubject(alert)
	body, err := e.formatAlertBody(alert)
	if err != nil {
		return fmt.Errorf("failed to format alert body: %w", err)
	}

	message := &EmailMessage{
		To:      e.config.Recipients,
		Subject: subject,
		Body:    body,
	}

	return e.sendEmail(ctx, message)
}

// SendSystemStatus sends a system status notification via email
func (e *EmailNotifier) SendSystemStatus(ctx context.Context, status *SystemStatus) error {
	if !e.config.Enabled || e.config.SMTPHost == "" {
		return fmt.Errorf("email notifications not configured")
	}

	subject := fmt.Sprintf("GZH Monitoring - System Status: %s", strings.ToUpper(status.Status))
	body, err := e.formatSystemStatusBody(status)
	if err != nil {
		return fmt.Errorf("failed to format system status body: %w", err)
	}

	message := &EmailMessage{
		To:      e.config.Recipients,
		Subject: subject,
		Body:    body,
	}

	return e.sendEmail(ctx, message)
}

// SendCustomMessage sends a custom email message
func (e *EmailNotifier) SendCustomMessage(ctx context.Context, title, text string, severity AlertSeverity) error {
	if !e.config.Enabled || e.config.SMTPHost == "" {
		return fmt.Errorf("email notifications not configured")
	}

	subject := fmt.Sprintf("GZH Monitoring - %s", title)
	body, err := e.formatCustomMessageBody(title, text, severity)
	if err != nil {
		return fmt.Errorf("failed to format custom message body: %w", err)
	}

	message := &EmailMessage{
		To:      e.config.Recipients,
		Subject: subject,
		Body:    body,
	}

	return e.sendEmail(ctx, message)
}

// TestConnection tests the email configuration
func (e *EmailNotifier) TestConnection(ctx context.Context) error {
	subject := "GZH Monitoring - Test Email"
	body := `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #2196F3; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background-color: #f5f5f5; }
        .success { color: #4CAF50; font-weight: bold; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h2>🧪 Test Email from GZH Monitoring</h2>
        </div>
        <div class="content">
            <p>If you're reading this message, your email integration is working correctly!</p>
            <p class="success">✅ Email configuration test successful</p>
            <p><strong>Test Time:</strong> %s</p>
            <p><strong>SMTP Server:</strong> %s:%d</p>
        </div>
    </div>
</body>
</html>
`
	body = fmt.Sprintf(body,
		time.Now().Format("2006-01-02 15:04:05 MST"),
		e.config.SMTPHost,
		e.config.SMTPPort,
	)

	message := &EmailMessage{
		To:      e.config.Recipients,
		Subject: subject,
		Body:    body,
	}

	return e.sendEmail(ctx, message)
}

// sendEmail sends an email message
func (e *EmailNotifier) sendEmail(ctx context.Context, message *EmailMessage) error {
	if len(message.To) == 0 {
		return fmt.Errorf("no recipients configured")
	}

	// Prepare email headers
	headers := make(map[string]string)
	headers["From"] = e.config.From
	headers["To"] = strings.Join(message.To, ", ")
	headers["Subject"] = message.Subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"
	headers["Date"] = time.Now().Format(time.RFC1123Z)

	// Build email message
	var msg bytes.Buffer
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n")
	msg.WriteString(message.Body)

	// Connect to SMTP server
	addr := fmt.Sprintf("%s:%d", e.config.SMTPHost, e.config.SMTPPort)

	var auth smtp.Auth
	if e.config.Username != "" && e.config.Password != "" {
		auth = smtp.PlainAuth("", e.config.Username, e.config.Password, e.config.SMTPHost)
	}

	// Send email
	var err error
	if e.config.UseTLS {
		err = e.sendEmailTLS(addr, auth, e.config.From, message.To, msg.Bytes())
	} else {
		err = smtp.SendMail(addr, auth, e.config.From, message.To, msg.Bytes())
	}

	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	e.logger.Info("Email sent successfully",
		zap.Strings("recipients", message.To),
		zap.String("subject", message.Subject))

	return nil
}

// sendEmailTLS sends email with TLS encryption
func (e *EmailNotifier) sendEmailTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	// Connect to the server
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Create SMTP client
	client, err := smtp.NewClient(conn, e.config.SMTPHost)
	if err != nil {
		return err
	}
	defer client.Close()

	// Start TLS
	tlsConfig := &tls.Config{
		ServerName: e.config.SMTPHost,
	}
	if err = client.StartTLS(tlsConfig); err != nil {
		return err
	}

	// Authenticate
	if auth != nil {
		if err = client.Auth(auth); err != nil {
			return err
		}
	}

	// Set sender and recipients
	if err = client.Mail(from); err != nil {
		return err
	}
	for _, recipient := range to {
		if err = client.Rcpt(recipient); err != nil {
			return err
		}
	}

	// Send the email body
	w, err := client.Data()
	if err != nil {
		return err
	}
	_, err = w.Write(msg)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}

	return client.Quit()
}

// formatAlertSubject formats the email subject for an alert
func (e *EmailNotifier) formatAlertSubject(alert *AlertInstance) string {
	statusEmoji := e.getStatusEmoji(alert.Status)
	return fmt.Sprintf("%s GZH Alert - %s: %s", statusEmoji, alert.Severity, alert.RuleName)
}

// formatAlertBody formats the email body for an alert
func (e *EmailNotifier) formatAlertBody(alert *AlertInstance) (string, error) {
	tmpl, exists := e.templates["alert"]
	if !exists {
		return "", fmt.Errorf("alert template not found")
	}

	var buf bytes.Buffer
	err := tmpl.Execute(&buf, map[string]interface{}{
		"Alert":         alert,
		"StatusEmoji":   e.getStatusEmoji(alert.Status),
		"StatusColor":   e.getStatusColor(alert.Status),
		"SeverityColor": e.getSeverityColor(AlertSeverity(alert.Severity)),
		"Timestamp":     time.Now().Format("2006-01-02 15:04:05 MST"),
	})
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// formatSystemStatusBody formats the email body for system status
func (e *EmailNotifier) formatSystemStatusBody(status *SystemStatus) (string, error) {
	tmpl, exists := e.templates["status"]
	if !exists {
		return "", fmt.Errorf("status template not found")
	}

	var buf bytes.Buffer
	err := tmpl.Execute(&buf, map[string]interface{}{
		"Status":        status,
		"StatusEmoji":   e.getSystemStatusEmoji(status.Status),
		"StatusColor":   e.getSystemStatusColor(status.Status),
		"MemoryUsageMB": float64(status.MemoryUsage) / 1024 / 1024,
		"Timestamp":     time.Now().Format("2006-01-02 15:04:05 MST"),
	})
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// formatCustomMessageBody formats a custom message body
func (e *EmailNotifier) formatCustomMessageBody(title, text string, severity AlertSeverity) (string, error) {
	tmpl, exists := e.templates["custom"]
	if !exists {
		return "", fmt.Errorf("custom template not found")
	}

	var buf bytes.Buffer
	err := tmpl.Execute(&buf, map[string]interface{}{
		"Title":         title,
		"Text":          text,
		"Severity":      severity,
		"SeverityColor": e.getSeverityColor(severity),
		"Timestamp":     time.Now().Format("2006-01-02 15:04:05 MST"),
	})
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// initializeTemplates initializes email templates
func (e *EmailNotifier) initializeTemplates() {
	// Alert template
	alertTemplate := `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; margin: 0; padding: 0; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: {{.SeverityColor}}; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background-color: #f5f5f5; }
        .field { margin: 10px 0; }
        .field-name { font-weight: bold; }
        .labels { margin-top: 20px; }
        .label { display: inline-block; background-color: #e0e0e0; padding: 5px 10px; margin: 5px; border-radius: 3px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h2>{{.StatusEmoji}} {{.Alert.RuleName}}</h2>
        </div>
        <div class="content">
            <div class="field">
                <span class="field-name">Status:</span> {{.Alert.Status}}
            </div>
            <div class="field">
                <span class="field-name">Severity:</span> {{.Alert.Severity}}
            </div>
            <div class="field">
                <span class="field-name">Message:</span> {{.Alert.Message}}
            </div>
            {{if .Alert.FiredAt}}
            <div class="field">
                <span class="field-name">Fired At:</span> {{.Alert.FiredAt.Format "2006-01-02 15:04:05 MST"}}
            </div>
            {{end}}
            {{if .Alert.ResolvedAt}}
            <div class="field">
                <span class="field-name">Resolved At:</span> {{.Alert.ResolvedAt.Format "2006-01-02 15:04:05 MST"}}
            </div>
            {{end}}
            {{if .Alert.Labels}}
            <div class="labels">
                <div class="field-name">Labels:</div>
                {{range $key, $value := .Alert.Labels}}
                <span class="label">{{$key}}: {{$value}}</span>
                {{end}}
            </div>
            {{end}}
            <hr>
            <p style="font-size: 12px; color: #666;">
                Generated by GZH Monitoring at {{.Timestamp}}
            </p>
        </div>
    </div>
</body>
</html>
`
	e.templates["alert"] = template.Must(template.New("alert").Parse(alertTemplate))

	// System status template
	statusTemplate := `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; margin: 0; padding: 0; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: {{.StatusColor}}; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background-color: #f5f5f5; }
        .metric { display: inline-block; width: 45%; margin: 10px 2.5%; text-align: center; background-color: white; padding: 15px; border-radius: 5px; }
        .metric-value { font-size: 24px; font-weight: bold; color: #333; }
        .metric-label { font-size: 14px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h2>{{.StatusEmoji}} System Status: {{.Status.Status}}</h2>
        </div>
        <div class="content">
            <div class="metric">
                <div class="metric-value">{{.Status.Uptime}}</div>
                <div class="metric-label">Uptime</div>
            </div>
            <div class="metric">
                <div class="metric-value">{{.Status.ActiveTasks}}</div>
                <div class="metric-label">Active Tasks</div>
            </div>
            <div class="metric">
                <div class="metric-value">{{printf "%.1f" .Status.CPUUsage}}%</div>
                <div class="metric-label">CPU Usage</div>
            </div>
            <div class="metric">
                <div class="metric-value">{{printf "%.1f" .MemoryUsageMB}} MB</div>
                <div class="metric-label">Memory Usage</div>
            </div>
            <div class="metric">
                <div class="metric-value">{{printf "%.1f" .Status.DiskUsage}}%</div>
                <div class="metric-label">Disk Usage</div>
            </div>
            <div class="metric">
                <div class="metric-value">{{.Status.TotalRequests}}</div>
                <div class="metric-label">Total Requests</div>
            </div>
            <hr>
            <p style="font-size: 12px; color: #666;">
                Generated by GZH Monitoring at {{.Timestamp}}
            </p>
        </div>
    </div>
</body>
</html>
`
	e.templates["status"] = template.Must(template.New("status").Parse(statusTemplate))

	// Custom message template
	customTemplate := `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; margin: 0; padding: 0; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: {{.SeverityColor}}; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background-color: #f5f5f5; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h2>{{.Title}}</h2>
        </div>
        <div class="content">
            <p>{{.Text}}</p>
            <hr>
            <p style="font-size: 12px; color: #666;">
                Generated by GZH Monitoring at {{.Timestamp}}
            </p>
        </div>
    </div>
</body>
</html>
`
	e.templates["custom"] = template.Must(template.New("custom").Parse(customTemplate))
}

// Helper methods

func (e *EmailNotifier) getStatusEmoji(status AlertStatus) string {
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

func (e *EmailNotifier) getStatusColor(status AlertStatus) string {
	switch status {
	case AlertStatusFiring:
		return "#f44336"
	case AlertStatusResolved:
		return "#4CAF50"
	case AlertStatusSilenced:
		return "#FF9800"
	default:
		return "#2196F3"
	}
}

func (e *EmailNotifier) getSeverityColor(severity AlertSeverity) string {
	switch severity {
	case AlertSeverityCritical:
		return "#f44336"
	case AlertSeverityHigh:
		return "#FF5722"
	case AlertSeverityMedium:
		return "#FF9800"
	case AlertSeverityLow:
		return "#4CAF50"
	case AlertSeverityInfo:
		return "#2196F3"
	default:
		return "#9E9E9E"
	}
}

func (e *EmailNotifier) getSystemStatusEmoji(status string) string {
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

func (e *EmailNotifier) getSystemStatusColor(status string) string {
	switch status {
	case "healthy":
		return "#4CAF50"
	case "warning":
		return "#FF9800"
	case "critical":
		return "#f44336"
	default:
		return "#2196F3"
	}
}
