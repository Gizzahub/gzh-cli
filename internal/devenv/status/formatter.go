// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package status

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// TableFormatter formats status as a table.
type TableFormatter struct {
	UseColor bool
}

// NewTableFormatter creates a new table formatter.
func NewTableFormatter(useColor bool) *TableFormatter {
	return &TableFormatter{UseColor: useColor}
}

// Format formats the status as a table.
func (t *TableFormatter) Format(statuses []ServiceStatus) (string, error) {
	if len(statuses) == 0 {
		return "No services to display", nil
	}

	var sb strings.Builder

	// Header
	sb.WriteString("Development Environment Status\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	// Table header
	sb.WriteString("Service    │ Status      │ Current              │ Credentials    │ Last Used\n")
	sb.WriteString("───────────┼─────────────┼──────────────────────┼────────────────┼───────────\n")

	activeCount := 0
	hasWarnings := false

	// Table rows
	for _, status := range statuses {
		serviceName := fmt.Sprintf("%-10s", status.Name)
		statusStr := t.formatStatus(status.Status)
		currentStr := t.formatCurrent(status.Current)
		credStr := t.formatCredentials(status.Credentials)
		lastUsedStr := t.formatLastUsed(status.LastUsed)

		if status.Status == StatusActive {
			activeCount++
		}
		if status.Credentials.Warning != "" || status.Status == StatusError {
			hasWarnings = true
		}

		sb.WriteString(fmt.Sprintf("%s │ %s │ %-20s │ %-14s │ %s\n",
			serviceName, statusStr, currentStr, credStr, lastUsedStr))
	}

	// Summary
	sb.WriteString("\n")
	if hasWarnings {
		sb.WriteString(t.colorize("⚠️ Warning", "yellow"))
		sb.WriteString(" (Some services have issues)\n")
	} else {
		sb.WriteString(t.colorize("✅ All Good", "green"))
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("Active Environments: %d/%d\n", activeCount, len(statuses)))

	return sb.String(), nil
}

// formatStatus formats the service status with colors.
func (t *TableFormatter) formatStatus(status Type) string {
	switch status {
	case StatusActive:
		return t.colorize("✅ Active  ", "green")
	case StatusInactive:
		return t.colorize("❌ Inactive", "red")
	case StatusError:
		return t.colorize("⚠️ Error   ", "yellow")
	case StatusUnknown:
		return t.colorize("❓ Unknown ", "gray")
	default:
		return t.colorize("❓ Unknown ", "gray")
	}
}

// formatCurrent formats the current configuration.
func (t *TableFormatter) formatCurrent(current CurrentConfig) string {
	parts := []string{}

	if current.Profile != "" {
		parts = append(parts, current.Profile)
	}
	if current.Project != "" {
		parts = append(parts, current.Project)
	}
	if current.Context != "" {
		parts = append(parts, current.Context)
	}

	if current.Region != "" {
		parts = append(parts, fmt.Sprintf("(%s)", current.Region))
	}
	if current.Namespace != "" && current.Namespace != awsDefaultProfile {
		parts = append(parts, fmt.Sprintf("/%s", current.Namespace))
	}

	if len(parts) == 0 {
		return "-"
	}

	result := strings.Join(parts, " ")
	if len(result) > 20 {
		return result[:17] + "..."
	}
	return result
}

// formatCredentials formats the credential status.
func (t *TableFormatter) formatCredentials(creds CredentialStatus) string {
	if !creds.Valid {
		return t.colorize("❌ Invalid", "red")
	}

	if creds.Warning != "" {
		if strings.Contains(creds.Warning, "expire") {
			return t.colorize("⚠️ Expires", "yellow")
		}
		return t.colorize("⚠️ Warning", "yellow")
	}

	if !creds.ExpiresAt.IsZero() {
		timeUntilExpiry := time.Until(creds.ExpiresAt)
		if timeUntilExpiry < 24*time.Hour {
			return t.colorize(fmt.Sprintf("⚠️ %s", t.formatDuration(timeUntilExpiry)), "yellow")
		}
		return t.colorize(fmt.Sprintf("✅ %s", t.formatDuration(timeUntilExpiry)), "green")
	}

	return t.colorize("✅ Valid", "green")
}

// formatLastUsed formats the last used time.
func (t *TableFormatter) formatLastUsed(lastUsed time.Time) string {
	if lastUsed.IsZero() {
		return "Unknown"
	}

	duration := time.Since(lastUsed)
	return t.formatDuration(duration) + " ago"
}

// formatDuration formats duration in a human-readable way.
func (t *TableFormatter) formatDuration(d time.Duration) string {
	if d < time.Minute {
		return "< 1 min"
	}
	if d < time.Hour {
		return fmt.Sprintf("%d min", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%d hour", int(d.Hours()))
	}
	return fmt.Sprintf("%d days", int(d.Hours()/24))
}

// colorize adds color to text if colors are enabled.
func (t *TableFormatter) colorize(text, color string) string {
	if !t.UseColor {
		return text
	}

	colors := map[string]string{
		"red":    "\033[31m",
		"green":  "\033[32m",
		"yellow": "\033[33m",
		"gray":   "\033[37m",
		"reset":  "\033[0m",
	}

	if colorCode, exists := colors[color]; exists {
		return colorCode + text + colors["reset"]
	}
	return text
}

// JSONFormatter formats status as JSON.
type JSONFormatter struct {
	Pretty bool
}

// NewJSONFormatter creates a new JSON formatter.
func NewJSONFormatter(pretty bool) *JSONFormatter {
	return &JSONFormatter{Pretty: pretty}
}

// Format formats the status as JSON.
func (j *JSONFormatter) Format(statuses []ServiceStatus) (string, error) {
	if j.Pretty {
		bytes, err := json.MarshalIndent(statuses, "", "  ")
		return string(bytes), err
	}
	bytes, err := json.Marshal(statuses)
	return string(bytes), err
}

// YAMLFormatter formats status as YAML.
type YAMLFormatter struct{}

// NewYAMLFormatter creates a new YAML formatter.
func NewYAMLFormatter() *YAMLFormatter {
	return &YAMLFormatter{}
}

// Format formats the status as YAML.
func (y *YAMLFormatter) Format(statuses []ServiceStatus) (string, error) {
	bytes, err := yaml.Marshal(statuses)
	return string(bytes), err
}
