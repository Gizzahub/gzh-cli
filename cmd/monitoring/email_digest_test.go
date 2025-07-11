package monitoring

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestEmailDigest_Configuration(t *testing.T) {
	logger := zap.NewNop()

	t.Run("Default digest configuration", func(t *testing.T) {
		config := &EmailConfig{
			SMTPHost:      "smtp.example.com",
			SMTPPort:      587,
			From:          "test@example.com",
			Recipients:    []string{"admin@example.com"},
			Enabled:       true,
			DigestEnabled: true,
		}

		notifier := NewEmailNotifier(config, logger)
		require.NotNil(t, notifier)

		assert.Equal(t, time.Hour, config.DigestInterval)
		assert.Equal(t, 50, config.DigestMaxAlerts)
		assert.Equal(t, AlertSeverityCritical, config.ImmediateSeverity)
		assert.NotNil(t, notifier.digest)
	})

	t.Run("Custom digest configuration", func(t *testing.T) {
		config := &EmailConfig{
			SMTPHost:          "smtp.example.com",
			SMTPPort:          587,
			From:              "test@example.com",
			Recipients:        []string{"admin@example.com"},
			Enabled:           true,
			DigestEnabled:     true,
			DigestInterval:    30 * time.Minute,
			DigestMaxAlerts:   25,
			ImmediateSeverity: AlertSeverityHigh,
		}

		notifier := NewEmailNotifier(config, logger)
		require.NotNil(t, notifier)

		assert.Equal(t, 30*time.Minute, config.DigestInterval)
		assert.Equal(t, 25, config.DigestMaxAlerts)
		assert.Equal(t, AlertSeverityHigh, config.ImmediateSeverity)
	})
}

func TestEmailDigest_ShouldSendImmediately(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name              string
		immediateSeverity AlertSeverity
		alertSeverity     AlertSeverity
		expected          bool
	}{
		{
			name:              "Critical threshold - Critical alert",
			immediateSeverity: AlertSeverityCritical,
			alertSeverity:     AlertSeverityCritical,
			expected:          true,
		},
		{
			name:              "Critical threshold - High alert",
			immediateSeverity: AlertSeverityCritical,
			alertSeverity:     AlertSeverityHigh,
			expected:          false,
		},
		{
			name:              "High threshold - Critical alert",
			immediateSeverity: AlertSeverityHigh,
			alertSeverity:     AlertSeverityCritical,
			expected:          true,
		},
		{
			name:              "High threshold - High alert",
			immediateSeverity: AlertSeverityHigh,
			alertSeverity:     AlertSeverityHigh,
			expected:          true,
		},
		{
			name:              "High threshold - Medium alert",
			immediateSeverity: AlertSeverityHigh,
			alertSeverity:     AlertSeverityMedium,
			expected:          false,
		},
		{
			name:              "Medium threshold - Low alert",
			immediateSeverity: AlertSeverityMedium,
			alertSeverity:     AlertSeverityLow,
			expected:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &EmailConfig{
				SMTPHost:          "smtp.example.com",
				SMTPPort:          587,
				From:              "test@example.com",
				Recipients:        []string{"admin@example.com"},
				Enabled:           true,
				DigestEnabled:     true,
				ImmediateSeverity: tt.immediateSeverity,
			}

			notifier := NewEmailNotifier(config, logger)

			alert := &AlertInstance{
				ID:       "test-alert-1",
				Severity: tt.alertSeverity,
				Status:   AlertStatusFiring,
			}

			result := notifier.shouldSendImmediately(alert)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEmailDigest_AddToDigest(t *testing.T) {
	logger := zap.NewNop()
	config := &EmailConfig{
		SMTPHost:        "smtp.example.com",
		SMTPPort:        587,
		From:            "test@example.com",
		Recipients:      []string{"admin@example.com"},
		Enabled:         true,
		DigestEnabled:   true,
		DigestMaxAlerts: 3, // Small limit for testing
	}

	notifier := NewEmailNotifier(config, logger)

	t.Run("Add alerts to digest", func(t *testing.T) {
		alert1 := &AlertInstance{ID: "alert-1", Severity: AlertSeverityMedium}
		alert2 := &AlertInstance{ID: "alert-2", Severity: AlertSeverityLow}

		notifier.addToDigest(alert1)
		notifier.addToDigest(alert2)

		stats := notifier.GetDigestStats()
		assert.Equal(t, 2, stats["total_alerts"].(int))
	})

	t.Run("Respect max alerts limit", func(t *testing.T) {
		// Add 5 alerts when max is 3
		for i := 0; i < 5; i++ {
			alert := &AlertInstance{
				ID:       string(rune('A' + i)),
				Severity: AlertSeverityMedium,
			}
			notifier.addToDigest(alert)
		}

		stats := notifier.GetDigestStats()
		assert.Equal(t, 3, stats["total_alerts"].(int)) // Should be limited to 3

		// Verify we kept the most recent alerts (C, D, E)
		notifier.digest.mutex.RLock()
		assert.Equal(t, "C", notifier.digest.alerts[0].ID)
		assert.Equal(t, "D", notifier.digest.alerts[1].ID)
		assert.Equal(t, "E", notifier.digest.alerts[2].ID)
		notifier.digest.mutex.RUnlock()
	})
}

func TestEmailDigest_CreateDigestSummary(t *testing.T) {
	logger := zap.NewNop()
	config := &EmailConfig{
		SMTPHost:       "smtp.example.com",
		SMTPPort:       587,
		From:           "test@example.com",
		Recipients:     []string{"admin@example.com"},
		Enabled:        true,
		DigestEnabled:  true,
		DigestInterval: time.Hour,
	}

	notifier := NewEmailNotifier(config, logger)

	// Add test alerts
	alerts := []*AlertInstance{
		{ID: "alert-1", Severity: AlertSeverityCritical, Status: AlertStatusFiring, RuleName: "High CPU Usage"},
		{ID: "alert-2", Severity: AlertSeverityHigh, Status: AlertStatusFiring, RuleName: "Memory Usage"},
		{ID: "alert-3", Severity: AlertSeverityMedium, Status: AlertStatusResolved, RuleName: "Disk Space"},
		{ID: "alert-4", Severity: AlertSeverityCritical, Status: AlertStatusFiring, RuleName: "Service Down"},
	}

	for _, alert := range alerts {
		notifier.addToDigest(alert)
	}

	summary := notifier.createDigestSummary()

	t.Run("Basic summary properties", func(t *testing.T) {
		assert.Equal(t, 4, summary.TotalAlerts)
		assert.Len(t, summary.Alerts, 4)
		assert.NotEmpty(t, summary.TimeRange)
		assert.False(t, summary.GeneratedAt.IsZero())
	})

	t.Run("Alert counts by severity", func(t *testing.T) {
		assert.Equal(t, 2, summary.AlertCounts[AlertSeverityCritical])
		assert.Equal(t, 1, summary.AlertCounts[AlertSeverityHigh])
		assert.Equal(t, 1, summary.AlertCounts[AlertSeverityMedium])
		assert.Equal(t, 0, summary.AlertCounts[AlertSeverityLow])
	})

	t.Run("Alert counts by status", func(t *testing.T) {
		assert.Equal(t, 3, summary.StatusCounts[AlertStatusFiring])
		assert.Equal(t, 1, summary.StatusCounts[AlertStatusResolved])
	})
}

func TestEmailDigest_FormatDigestSubject(t *testing.T) {
	logger := zap.NewNop()
	config := &EmailConfig{
		SMTPHost:      "smtp.example.com",
		SMTPPort:      587,
		From:          "test@example.com",
		Recipients:    []string{"admin@example.com"},
		Enabled:       true,
		DigestEnabled: true,
	}

	notifier := NewEmailNotifier(config, logger)

	tests := []struct {
		name     string
		summary  *DigestSummary
		expected string
	}{
		{
			name: "No alerts",
			summary: &DigestSummary{
				TotalAlerts: 0,
				AlertCounts: make(map[AlertSeverity]int),
			},
			expected: "GZH Monitoring - No Alerts Digest",
		},
		{
			name: "Critical alerts present",
			summary: &DigestSummary{
				TotalAlerts: 5,
				AlertCounts: map[AlertSeverity]int{
					AlertSeverityCritical: 2,
					AlertSeverityHigh:     1,
					AlertSeverityMedium:   2,
				},
			},
			expected: "🚨 GZH Monitoring - 5 Alerts (2 Critical, 1 High)",
		},
		{
			name: "High priority alerts without critical",
			summary: &DigestSummary{
				TotalAlerts: 3,
				AlertCounts: map[AlertSeverity]int{
					AlertSeverityHigh:   2,
					AlertSeverityMedium: 1,
				},
			},
			expected: "⚠️ GZH Monitoring - 3 Alerts (2 High Priority)",
		},
		{
			name: "Low priority alerts only",
			summary: &DigestSummary{
				TotalAlerts: 2,
				AlertCounts: map[AlertSeverity]int{
					AlertSeverityMedium: 1,
					AlertSeverityLow:    1,
				},
			},
			expected: "📊 GZH Monitoring - 2 Alerts Digest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := notifier.formatDigestSubject(tt.summary)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEmailDigest_GetDigestStats(t *testing.T) {
	logger := zap.NewNop()
	config := &EmailConfig{
		SMTPHost:       "smtp.example.com",
		SMTPPort:       587,
		From:           "test@example.com",
		Recipients:     []string{"admin@example.com"},
		Enabled:        true,
		DigestEnabled:  true,
		DigestInterval: 30 * time.Minute,
	}

	notifier := NewEmailNotifier(config, logger)

	// Add test alerts
	alerts := []*AlertInstance{
		{ID: "alert-1", Severity: AlertSeverityCritical},
		{ID: "alert-2", Severity: AlertSeverityHigh},
		{ID: "alert-3", Severity: AlertSeverityMedium},
	}

	for _, alert := range alerts {
		notifier.addToDigest(alert)
	}

	stats := notifier.GetDigestStats()

	t.Run("Basic stats", func(t *testing.T) {
		assert.Equal(t, 3, stats["total_alerts"].(int))
		assert.Equal(t, true, stats["digest_enabled"].(bool))
		assert.Equal(t, "30m0s", stats["digest_interval"].(string))
		assert.NotNil(t, stats["last_sent"])
		assert.NotNil(t, stats["next_send"])
	})

	t.Run("Severity counts", func(t *testing.T) {
		severityCounts := stats["severity_counts"].(map[string]int)
		assert.Equal(t, 1, severityCounts["critical"])
		assert.Equal(t, 1, severityCounts["high"])
		assert.Equal(t, 1, severityCounts["medium"])
		assert.Equal(t, 0, severityCounts["low"])
	})
}

func TestEmailDigest_SendAlert_DigestMode(t *testing.T) {
	logger := zap.NewNop()

	t.Run("Alert added to digest when digest enabled", func(t *testing.T) {
		config := &EmailConfig{
			SMTPHost:          "smtp.example.com",
			SMTPPort:          587,
			From:              "test@example.com",
			Recipients:        []string{"admin@example.com"},
			Enabled:           true,
			DigestEnabled:     true,
			ImmediateSeverity: AlertSeverityCritical,
		}

		notifier := NewEmailNotifier(config, logger)

		// Medium severity alert should go to digest
		alert := &AlertInstance{
			ID:       "test-alert",
			Severity: AlertSeverityMedium,
			Status:   AlertStatusFiring,
			RuleName: "Test Alert",
			Message:  "Test message",
		}

		err := notifier.SendAlert(context.Background(), alert)

		// Should not error (alert added to digest)
		assert.NoError(t, err)

		// Check alert was added to digest
		stats := notifier.GetDigestStats()
		assert.Equal(t, 1, stats["total_alerts"].(int))
	})

	t.Run("Critical alert sent immediately even with digest enabled", func(t *testing.T) {
		config := &EmailConfig{
			SMTPHost:          "localhost", // Will fail but that's OK for this test
			SMTPPort:          587,
			From:              "test@example.com",
			Recipients:        []string{"admin@example.com"},
			Enabled:           true,
			DigestEnabled:     true,
			ImmediateSeverity: AlertSeverityCritical,
		}

		notifier := NewEmailNotifier(config, logger)

		// Critical alert should be sent immediately
		alert := &AlertInstance{
			ID:       "critical-alert",
			Severity: AlertSeverityCritical,
			Status:   AlertStatusFiring,
			RuleName: "Critical Alert",
			Message:  "Critical message",
		}

		err := notifier.SendAlert(context.Background(), alert)

		// Should error because SMTP will fail, but this proves it tried to send immediately
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to send email")

		// Check alert was NOT added to digest
		stats := notifier.GetDigestStats()
		assert.Equal(t, 0, stats["total_alerts"].(int))
	})
}

func TestEmailDigest_Template(t *testing.T) {
	logger := zap.NewNop()
	config := &EmailConfig{
		SMTPHost:      "smtp.example.com",
		SMTPPort:      587,
		From:          "test@example.com",
		Recipients:    []string{"admin@example.com"},
		Enabled:       true,
		DigestEnabled: true,
	}

	notifier := NewEmailNotifier(config, logger)

	t.Run("Digest template exists", func(t *testing.T) {
		_, exists := notifier.templates["digest"]
		assert.True(t, exists, "Digest template should be initialized")
	})

	t.Run("Format digest body", func(t *testing.T) {
		firedTime := time.Now().Add(-30 * time.Minute)
		summary := &DigestSummary{
			TotalAlerts: 2,
			TimeRange:   "2024-01-01 10:00 - 2024-01-01 11:00",
			GeneratedAt: time.Now(),
			Alerts: []*AlertInstance{
				{
					ID:       "alert-1",
					RuleName: "High CPU",
					Severity: AlertSeverityCritical,
					Status:   AlertStatusFiring,
					Message:  "CPU usage above 90%",
					FiredAt:  &firedTime,
				},
			},
			AlertCounts: map[AlertSeverity]int{
				AlertSeverityCritical: 1,
				AlertSeverityHigh:     1,
			},
			StatusCounts: map[AlertStatus]int{
				AlertStatusFiring: 2,
			},
		}

		body, err := notifier.formatDigestBody(summary)

		require.NoError(t, err)
		assert.Contains(t, body, "GZH Monitoring Digest")
		assert.Contains(t, body, "2024-01-01 10:00 - 2024-01-01 11:00")
		assert.Contains(t, body, "High CPU")
		assert.Contains(t, body, "CPU usage above 90%")
		assert.Contains(t, body, "Total Alerts")
	})
}
