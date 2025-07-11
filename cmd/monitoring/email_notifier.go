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
	"sync"
	"time"

	"go.uber.org/zap"
)

// EmailNotifier handles email notifications
type EmailNotifier struct {
	config    *EmailConfig
	logger    *zap.Logger
	templates map[string]*template.Template
	digest    *DigestCollector
}

// EmailConfig represents email notification configuration
type EmailConfig struct {
	SMTPHost          string        `json:"smtp_host"`
	SMTPPort          int           `json:"smtp_port"`
	Username          string        `json:"username"`
	Password          string        `json:"password"`
	From              string        `json:"from"`
	Recipients        []string      `json:"recipients"`
	UseTLS            bool          `json:"use_tls"`
	Enabled           bool          `json:"enabled"`
	DigestEnabled     bool          `json:"digest_enabled"`
	DigestInterval    time.Duration `json:"digest_interval"`
	DigestMaxAlerts   int           `json:"digest_max_alerts"`
	ImmediateSeverity AlertSeverity `json:"immediate_severity"`
}

// DigestCollector collects alerts for digest email
type DigestCollector struct {
	alerts      []*AlertInstance
	systemStats *SystemStatus
	mutex       sync.RWMutex
	lastSent    time.Time
}

// DigestSummary contains aggregated digest information
type DigestSummary struct {
	Alerts       []*AlertInstance
	AlertCounts  map[AlertSeverity]int
	StatusCounts map[AlertStatus]int
	TimeRange    string
	TotalAlerts  int
	SystemStatus *SystemStatus
	GeneratedAt  time.Time
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
		digest: &DigestCollector{
			alerts:   make([]*AlertInstance, 0),
			lastSent: time.Now(),
		},
	}

	// Set default digest configuration if not specified
	if config.DigestInterval == 0 {
		config.DigestInterval = time.Hour // Default: hourly digest
	}
	if config.DigestMaxAlerts == 0 {
		config.DigestMaxAlerts = 50 // Default: max 50 alerts per digest
	}
	if config.ImmediateSeverity == "" {
		config.ImmediateSeverity = AlertSeverityCritical // Default: only critical alerts sent immediately
	}

	// Initialize email templates
	notifier.initializeTemplates()

	// Start digest scheduler if digest is enabled
	if config.DigestEnabled {
		go notifier.startDigestScheduler()
	}

	return notifier
}

// SendAlert sends an alert notification via email
func (e *EmailNotifier) SendAlert(ctx context.Context, alert *AlertInstance) error {
	if !e.config.Enabled || e.config.SMTPHost == "" {
		return fmt.Errorf("email notifications not configured")
	}

	// Check if digest is enabled and if this alert should be sent immediately
	if e.config.DigestEnabled && !e.shouldSendImmediately(alert) {
		e.addToDigest(alert)
		e.logger.Debug("Alert added to digest queue",
			zap.String("alert_id", alert.ID),
			zap.String("severity", string(alert.Severity)))
		return nil
	}

	// Send immediate alert
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

	// Digest template
	digestTemplate := `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; margin: 0; padding: 0; background-color: #f5f5f5; }
        .container { max-width: 700px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #2196F3, #21CBF3); color: white; padding: 30px; text-align: center; border-radius: 8px 8px 0 0; }
        .summary { background-color: white; padding: 20px; border-radius: 8px; margin: 20px 0; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .stats { display: flex; justify-content: space-around; margin: 20px 0; flex-wrap: wrap; }
        .stat { text-align: center; padding: 15px; margin: 5px; background-color: #f8f9fa; border-radius: 8px; min-width: 100px; }
        .stat-number { font-size: 24px; font-weight: bold; color: #2196F3; }
        .stat-label { font-size: 12px; color: #666; text-transform: uppercase; }
        .alerts-section { background-color: white; padding: 20px; border-radius: 8px; margin: 20px 0; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .alert-item { border-left: 4px solid #ddd; padding: 15px; margin: 10px 0; background-color: #fafafa; border-radius: 0 4px 4px 0; }
        .alert-critical { border-left-color: #f44336; }
        .alert-high { border-left-color: #FF5722; }
        .alert-medium { border-left-color: #FF9800; }
        .alert-low { border-left-color: #4CAF50; }
        .alert-info { border-left-color: #2196F3; }
        .alert-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; }
        .alert-title { font-weight: bold; font-size: 16px; }
        .alert-severity { padding: 4px 8px; border-radius: 12px; font-size: 12px; font-weight: bold; color: white; }
        .severity-critical { background-color: #f44336; }
        .severity-high { background-color: #FF5722; }
        .severity-medium { background-color: #FF9800; }
        .severity-low { background-color: #4CAF50; }
        .severity-info { background-color: #2196F3; }
        .alert-meta { font-size: 14px; color: #666; margin-top: 8px; }
        .footer { text-align: center; padding: 20px; color: #666; font-size: 12px; }
        .time-range { background-color: #e3f2fd; padding: 10px; border-radius: 4px; margin: 10px 0; }
        .no-alerts { text-align: center; padding: 40px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📊 GZH Monitoring Digest</h1>
            <p>Alert Summary Report</p>
        </div>

        <div class="time-range">
            <strong>Time Range:</strong> {{.TimeRange}}
        </div>

        <div class="summary">
            <h2>Summary</h2>
            <div class="stats">
                <div class="stat">
                    <div class="stat-number">{{.TotalAlerts}}</div>
                    <div class="stat-label">Total Alerts</div>
                </div>
                {{range $severity, $count := .AlertCounts}}
                {{if gt $count 0}}
                <div class="stat">
                    <div class="stat-number" style="color: {{if eq $severity "critical"}}#f44336{{else if eq $severity "high"}}#FF5722{{else if eq $severity "medium"}}#FF9800{{else if eq $severity "low"}}#4CAF50{{else}}#2196F3{{end}};">{{$count}}</div>
                    <div class="stat-label">{{if eq $severity "critical"}}Critical{{else if eq $severity "high"}}High{{else if eq $severity "medium"}}Medium{{else if eq $severity "low"}}Low{{else}}Info{{end}}</div>
                </div>
                {{end}}
                {{end}}
            </div>
        </div>

        {{if .Alerts}}
        <div class="alerts-section">
            <h2>Alert Details</h2>
            {{range .Alerts}}
            <div class="alert-item alert-{{.Severity}}">
                <div class="alert-header">
                    <div class="alert-title">{{.RuleName}}</div>
                    <span class="alert-severity severity-{{.Severity}}">{{.Severity}}</span>
                </div>
                <div>{{.Message}}</div>
                <div class="alert-meta">
                    Status: {{.Status}} | 
                    {{if .FiredAt}}Fired: {{.FiredAt.Format "2006-01-02 15:04"}}{{end}}
                    {{if .ResolvedAt}} | Resolved: {{.ResolvedAt.Format "2006-01-02 15:04"}}{{end}}
                </div>
            </div>
            {{end}}
        </div>
        {{else}}
        <div class="alerts-section">
            <div class="no-alerts">
                ✅ No alerts during this period
            </div>
        </div>
        {{end}}

        {{if .SystemStatus}}
        <div class="summary">
            <h2>System Status</h2>
            <p><strong>Overall Status:</strong> {{.SystemStatus.Status}}</p>
            <p><strong>Uptime:</strong> {{.SystemStatus.Uptime}}</p>
            {{if .SystemStatus.Metrics}}
            <div style="display: flex; justify-content: space-around; flex-wrap: wrap;">
                {{range $key, $value := .SystemStatus.Metrics}}
                <div class="stat">
                    <div class="stat-number">{{$value}}</div>
                    <div class="stat-label">{{$key}}</div>
                </div>
                {{end}}
            </div>
            {{end}}
        </div>
        {{end}}

        <div class="footer">
            <p>Generated at: {{.GeneratedAt.Format "2006-01-02 15:04:05 MST"}}</p>
            <p>This is an automated digest from GZH Monitoring System</p>
        </div>
    </div>
</body>
</html>
`
	e.templates["digest"] = template.Must(template.New("digest").Parse(digestTemplate))
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

// Digest-related methods

// shouldSendImmediately determines if an alert should be sent immediately instead of added to digest
func (e *EmailNotifier) shouldSendImmediately(alert *AlertInstance) bool {
	// Send immediately if severity is at or above the configured threshold
	switch e.config.ImmediateSeverity {
	case AlertSeverityCritical:
		return alert.Severity == AlertSeverityCritical
	case AlertSeverityHigh:
		return alert.Severity == AlertSeverityCritical || alert.Severity == AlertSeverityHigh
	case AlertSeverityMedium:
		return alert.Severity == AlertSeverityCritical || alert.Severity == AlertSeverityHigh || alert.Severity == AlertSeverityMedium
	default:
		return false
	}
}

// addToDigest adds an alert to the digest collection
func (e *EmailNotifier) addToDigest(alert *AlertInstance) {
	e.digest.mutex.Lock()
	defer e.digest.mutex.Unlock()

	// Add alert to digest
	e.digest.alerts = append(e.digest.alerts, alert)

	// Limit the number of alerts in digest to prevent memory issues
	if len(e.digest.alerts) > e.config.DigestMaxAlerts {
		// Remove oldest alerts, keep most recent ones
		e.digest.alerts = e.digest.alerts[len(e.digest.alerts)-e.config.DigestMaxAlerts:]
	}
}

// updateSystemStatus updates the system status in digest
func (e *EmailNotifier) updateSystemStatus(status *SystemStatus) {
	e.digest.mutex.Lock()
	defer e.digest.mutex.Unlock()
	e.digest.systemStats = status
}

// startDigestScheduler starts the digest email scheduler
func (e *EmailNotifier) startDigestScheduler() {
	ticker := time.NewTicker(e.config.DigestInterval)
	defer ticker.Stop()

	for range ticker.C {
		if err := e.sendDigest(context.Background()); err != nil {
			e.logger.Error("Failed to send digest email", zap.Error(err))
		}
	}
}

// sendDigest sends a digest email with collected alerts
func (e *EmailNotifier) sendDigest(ctx context.Context) error {
	e.digest.mutex.Lock()
	defer e.digest.mutex.Unlock()

	// Skip if no alerts to send
	if len(e.digest.alerts) == 0 {
		e.logger.Debug("No alerts in digest queue, skipping digest email")
		return nil
	}

	// Create digest summary
	summary := e.createDigestSummary()

	// Format digest email
	subject := e.formatDigestSubject(summary)
	body, err := e.formatDigestBody(summary)
	if err != nil {
		return fmt.Errorf("failed to format digest body: %w", err)
	}

	message := &EmailMessage{
		To:      e.config.Recipients,
		Subject: subject,
		Body:    body,
	}

	// Send digest email
	if err := e.sendEmail(ctx, message); err != nil {
		return fmt.Errorf("failed to send digest email: %w", err)
	}

	// Clear digest after successful send
	e.digest.alerts = make([]*AlertInstance, 0)
	e.digest.lastSent = time.Now()

	e.logger.Info("Digest email sent successfully",
		zap.Int("alert_count", summary.TotalAlerts),
		zap.String("time_range", summary.TimeRange))

	return nil
}

// createDigestSummary creates a summary of collected alerts
func (e *EmailNotifier) createDigestSummary() *DigestSummary {
	now := time.Now()

	summary := &DigestSummary{
		Alerts:       make([]*AlertInstance, len(e.digest.alerts)),
		AlertCounts:  make(map[AlertSeverity]int),
		StatusCounts: make(map[AlertStatus]int),
		GeneratedAt:  now,
		TotalAlerts:  len(e.digest.alerts),
		SystemStatus: e.digest.systemStats,
	}

	// Copy alerts
	copy(summary.Alerts, e.digest.alerts)

	// Calculate time range
	if e.digest.lastSent.IsZero() {
		summary.TimeRange = fmt.Sprintf("Since %s", now.Add(-e.config.DigestInterval).Format("2006-01-02 15:04"))
	} else {
		summary.TimeRange = fmt.Sprintf("%s - %s",
			e.digest.lastSent.Format("2006-01-02 15:04"),
			now.Format("2006-01-02 15:04"))
	}

	// Count alerts by severity and status
	for _, alert := range e.digest.alerts {
		summary.AlertCounts[alert.Severity]++
		summary.StatusCounts[alert.Status]++
	}

	return summary
}

// formatDigestSubject formats the subject line for digest emails
func (e *EmailNotifier) formatDigestSubject(summary *DigestSummary) string {
	if summary.TotalAlerts == 0 {
		return "GZH Monitoring - No Alerts Digest"
	}

	criticalCount := summary.AlertCounts[AlertSeverityCritical]
	highCount := summary.AlertCounts[AlertSeverityHigh]

	if criticalCount > 0 {
		return fmt.Sprintf("🚨 GZH Monitoring - %d Alerts (%d Critical, %d High)",
			summary.TotalAlerts, criticalCount, highCount)
	} else if highCount > 0 {
		return fmt.Sprintf("⚠️ GZH Monitoring - %d Alerts (%d High Priority)",
			summary.TotalAlerts, highCount)
	}

	return fmt.Sprintf("📊 GZH Monitoring - %d Alerts Digest", summary.TotalAlerts)
}

// formatDigestBody formats the HTML body for digest emails
func (e *EmailNotifier) formatDigestBody(summary *DigestSummary) (string, error) {
	tmpl, exists := e.templates["digest"]
	if !exists {
		return "", fmt.Errorf("digest template not found")
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, summary); err != nil {
		return "", fmt.Errorf("failed to execute digest template: %w", err)
	}

	return buf.String(), nil
}

// SendDigestNow forces sending a digest email immediately
func (e *EmailNotifier) SendDigestNow(ctx context.Context) error {
	return e.sendDigest(ctx)
}

// GetDigestStats returns current digest statistics
func (e *EmailNotifier) GetDigestStats() map[string]interface{} {
	e.digest.mutex.RLock()
	defer e.digest.mutex.RUnlock()

	stats := make(map[string]interface{})
	stats["total_alerts"] = len(e.digest.alerts)
	stats["last_sent"] = e.digest.lastSent
	stats["next_send"] = e.digest.lastSent.Add(e.config.DigestInterval)
	stats["digest_enabled"] = e.config.DigestEnabled
	stats["digest_interval"] = e.config.DigestInterval.String()

	// Count by severity
	severityCounts := make(map[string]int)
	for _, alert := range e.digest.alerts {
		severityCounts[string(alert.Severity)]++
	}
	stats["severity_counts"] = severityCounts

	return stats
}
