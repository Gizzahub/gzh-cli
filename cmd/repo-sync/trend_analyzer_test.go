package reposync

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestTrendAnalyzer(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tempDir := t.TempDir()

	t.Run("NewTrendAnalyzer", func(t *testing.T) {
		analyzer := NewTrendAnalyzer(logger, tempDir)
		assert.NotNil(t, analyzer)
		assert.Equal(t, tempDir, analyzer.dataDir)
		assert.Equal(t, 10.0, analyzer.alertThresholds.QualityDropThreshold)
	})

	t.Run("SetThresholds", func(t *testing.T) {
		analyzer := NewTrendAnalyzer(logger, tempDir)

		customThresholds := AlertThresholds{
			QualityDropThreshold: 15.0,
			MinimumQualityScore:  70.0,
		}

		analyzer.SetThresholds(customThresholds)
		assert.Equal(t, 15.0, analyzer.alertThresholds.QualityDropThreshold)
		assert.Equal(t, 70.0, analyzer.alertThresholds.MinimumQualityScore)
	})

	t.Run("CheckQualityDrop", func(t *testing.T) {
		analyzer := NewTrendAnalyzer(logger, tempDir)

		// Test with insufficient data
		history := []*QualityResult{}
		alert := analyzer.checkQualityDrop(history)
		assert.Nil(t, alert)

		// Test with single data point
		history = []*QualityResult{
			{OverallScore: 85.0},
		}
		alert = analyzer.checkQualityDrop(history)
		assert.Nil(t, alert)

		// Test below minimum threshold
		history = []*QualityResult{
			{OverallScore: 80.0, Repository: "test-repo"},
			{OverallScore: 45.0, Repository: "test-repo"},
		}
		alert = analyzer.checkQualityDrop(history)
		assert.NotNil(t, alert)
		assert.Equal(t, AlertTypeQualityDrop, alert.Type)
		assert.Equal(t, AlertSeverityCritical, alert.Severity)

		// Test percentage drop
		history = []*QualityResult{
			{OverallScore: 90.0, Repository: "test-repo"},
			{OverallScore: 75.0, Repository: "test-repo"}, // 16.7% drop
		}
		alert = analyzer.checkQualityDrop(history)
		assert.NotNil(t, alert)
		assert.Equal(t, AlertSeverityHigh, alert.Severity)

		// Test no significant drop
		history = []*QualityResult{
			{OverallScore: 90.0, Repository: "test-repo"},
			{OverallScore: 88.0, Repository: "test-repo"}, // 2.2% drop
		}
		alert = analyzer.checkQualityDrop(history)
		assert.Nil(t, alert)
	})

	t.Run("CheckComplexityTrend", func(t *testing.T) {
		analyzer := NewTrendAnalyzer(logger, tempDir)

		// Test high complexity
		history := []*QualityResult{
			{
				Repository: "test-repo",
				Metrics: QualityMetrics{
					AvgComplexity: 20.0, // Above threshold of 15.0
				},
			},
		}
		alert := analyzer.checkComplexityTrend(history)
		assert.NotNil(t, alert)
		assert.Equal(t, AlertTypeComplexityHigh, alert.Type)
		assert.Equal(t, AlertSeverityHigh, alert.Severity)

		// Test increasing complexity trend
		now := time.Now()
		history = make([]*QualityResult, 5)
		for i := 0; i < 5; i++ {
			history[i] = &QualityResult{
				Repository: "test-repo",
				Timestamp:  now.Add(time.Duration(i) * time.Hour),
				Metrics: QualityMetrics{
					AvgComplexity: 5.0 + float64(i)*2.0, // Increasing complexity
				},
			}
		}
		alert = analyzer.checkComplexityTrend(history)
		assert.NotNil(t, alert)
		assert.Equal(t, AlertTypeComplexityHigh, alert.Type)
	})

	t.Run("CheckSecurityIssues", func(t *testing.T) {
		analyzer := NewTrendAnalyzer(logger, tempDir)

		// Test low security score
		result := &QualityResult{
			Repository: "test-repo",
			Metrics: QualityMetrics{
				SecurityScore: 70.0, // Below threshold of 80.0
			},
		}
		alert := analyzer.checkSecurityIssues(result)
		assert.NotNil(t, alert)
		assert.Equal(t, AlertTypeSecurityIssue, alert.Type)
		assert.Equal(t, AlertSeverityCritical, alert.Severity)

		// Test critical security issues
		result = &QualityResult{
			Repository: "test-repo",
			Metrics: QualityMetrics{
				SecurityScore: 95.0,
			},
			SecurityIssues: []SecurityIssue{
				{Severity: "critical"},
			},
		}
		alert = analyzer.checkSecurityIssues(result)
		assert.NotNil(t, alert)
		assert.Equal(t, AlertSeverityCritical, alert.Severity)
	})

	t.Run("DetectAnomalies", func(t *testing.T) {
		analyzer := NewTrendAnalyzer(logger, tempDir)

		// Test with insufficient data
		history := make([]*QualityResult, 5)
		alerts := analyzer.detectAnomalies(history)
		assert.Empty(t, alerts)

		// Test with anomaly (outlier)
		history = make([]*QualityResult, 15)
		for i := 0; i < 14; i++ {
			history[i] = &QualityResult{
				OverallScore: 85.0 + float64(i)*0.5, // Normal range 85-92
				Metrics: QualityMetrics{
					AvgComplexity: 5.0,
				},
			}
		}
		// Add outlier
		history[14] = &QualityResult{
			Repository:   "test-repo",
			OverallScore: 50.0, // Significant outlier
			Metrics: QualityMetrics{
				AvgComplexity: 5.0,
			},
		}

		alerts := analyzer.detectAnomalies(history)
		assert.NotEmpty(t, alerts)

		foundAnomalyAlert := false
		for _, alert := range alerts {
			if alert.Type == AlertTypeTrendAnomaly {
				foundAnomalyAlert = true
				break
			}
		}
		assert.True(t, foundAnomalyAlert)
	})
}

func TestAlertHandlers(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	t.Run("ConsoleAlertHandler", func(t *testing.T) {
		handler := NewConsoleAlertHandler(logger)

		alert := &QualityAlert{
			ID:         "test-alert",
			Type:       AlertTypeQualityDrop,
			Severity:   AlertSeverityHigh,
			Timestamp:  time.Now(),
			Repository: "test-repo",
			Message:    "Test alert message",
			Details: map[string]interface{}{
				"previous_score": 90.0,
				"current_score":  75.0,
			},
			Suggestions: []string{
				"Review recent changes",
				"Run code analysis",
			},
		}

		ctx := context.Background()
		err := handler.HandleAlert(ctx, alert)
		assert.NoError(t, err)
	})

	t.Run("FileAlertHandler", func(t *testing.T) {
		tempDir := t.TempDir()
		handler := NewFileAlertHandler(logger, tempDir)

		alert := &QualityAlert{
			ID:         "test-alert",
			Type:       AlertTypeQualityDrop,
			Severity:   AlertSeverityHigh,
			Timestamp:  time.Now(),
			Repository: "test-repo",
			Message:    "Test alert message",
		}

		ctx := context.Background()
		err := handler.HandleAlert(ctx, alert)
		assert.NoError(t, err)

		// Check if file was created
		alertsDir := filepath.Join(tempDir, "alerts", "test-repo")
		files, err := os.ReadDir(alertsDir)
		assert.NoError(t, err)
		assert.Len(t, files, 1)
	})

	t.Run("CompositeAlertHandler", func(t *testing.T) {
		tempDir := t.TempDir()

		consoleHandler := NewConsoleAlertHandler(logger)
		fileHandler := NewFileAlertHandler(logger, tempDir)

		composite := NewCompositeAlertHandler(consoleHandler, fileHandler)

		alert := &QualityAlert{
			ID:         "test-alert",
			Type:       AlertTypeQualityDrop,
			Severity:   AlertSeverityHigh,
			Timestamp:  time.Now(),
			Repository: "test-repo",
			Message:    "Test alert message",
		}

		ctx := context.Background()
		err := composite.HandleAlert(ctx, alert)
		assert.NoError(t, err)

		// Check if file was created by file handler
		alertsDir := filepath.Join(tempDir, "alerts", "test-repo")
		files, err := os.ReadDir(alertsDir)
		assert.NoError(t, err)
		assert.Len(t, files, 1)
	})
}

func TestCalculateStats(t *testing.T) {
	t.Run("EmptyValues", func(t *testing.T) {
		mean, stdDev := calculateStats([]float64{})
		assert.Equal(t, 0.0, mean)
		assert.Equal(t, 0.0, stdDev)
	})

	t.Run("SingleValue", func(t *testing.T) {
		mean, stdDev := calculateStats([]float64{10.0})
		assert.Equal(t, 10.0, mean)
		assert.Equal(t, 0.0, stdDev)
	})

	t.Run("MultipleValues", func(t *testing.T) {
		values := []float64{10.0, 20.0, 30.0, 40.0, 50.0}
		mean, stdDev := calculateStats(values)

		assert.InDelta(t, 30.0, mean, 0.001)
		assert.InDelta(t, 14.142, stdDev, 0.001)
	})
}

func TestGenerateAlertID(t *testing.T) {
	id1 := generateAlertID()
	id2 := generateAlertID()

	assert.NotEqual(t, id1, id2)
	assert.Contains(t, id1, "alert-")
	assert.Contains(t, id2, "alert-")
}

func TestCountCriticalSecurityIssues(t *testing.T) {
	result := &QualityResult{
		SecurityIssues: []SecurityIssue{
			{Severity: "critical"},
			{Severity: "major"},
			{Severity: "critical"},
			{CVSS: 9.5}, // High CVSS score
			{CVSS: 5.0}, // Medium CVSS score
		},
	}

	count := countCriticalSecurityIssues(result)
	assert.Equal(t, 3, count) // 2 critical severity + 1 high CVSS
}
