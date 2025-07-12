package reposync

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"
)

// TrendAnalyzer analyzes quality trends and detects anomalies
type TrendAnalyzer struct {
	logger          *zap.Logger
	dataDir         string
	alertThresholds AlertThresholds
	alertHandlers   []AlertHandler
}

// AlertThresholds defines thresholds for various alerts
type AlertThresholds struct {
	// Quality score thresholds
	QualityDropThreshold float64 // Percentage drop to trigger alert
	MinimumQualityScore  float64 // Absolute minimum quality score

	// Complexity thresholds
	MaxComplexity          float64 // Maximum average complexity
	ComplexityIncreaseRate float64 // Maximum allowed increase rate

	// Coverage thresholds
	MinimumCoverage       float64 // Minimum test coverage
	CoverageDropThreshold float64 // Coverage drop to trigger alert

	// Technical debt thresholds
	MaxDebtRatio     float64 // Maximum debt ratio
	DebtIncreaseRate float64 // Maximum debt increase rate

	// Security thresholds
	MinSecurityScore  float64 // Minimum security score
	MaxCriticalIssues int     // Maximum critical issues allowed
}

// AlertHandler handles quality alerts
type AlertHandler interface {
	HandleAlert(ctx context.Context, alert *QualityAlert) error
}

// QualityAlert represents a quality alert
type QualityAlert struct {
	ID          string                 `json:"id"`
	Type        AlertType              `json:"type"`
	Severity    AlertSeverity          `json:"severity"`
	Timestamp   time.Time              `json:"timestamp"`
	Repository  string                 `json:"repository"`
	Message     string                 `json:"message"`
	Details     map[string]interface{} `json:"details"`
	Suggestions []string               `json:"suggestions"`
}

// AlertType represents the type of alert
type AlertType string

const (
	AlertTypeQualityDrop    AlertType = "quality_drop"
	AlertTypeComplexityHigh AlertType = "complexity_high"
	AlertTypeCoverageLow    AlertType = "coverage_low"
	AlertTypeDebtIncrease   AlertType = "debt_increase"
	AlertTypeSecurityIssue  AlertType = "security_issue"
	AlertTypeTrendAnomaly   AlertType = "trend_anomaly"
)

// AlertSeverity represents alert severity
type AlertSeverity string

const (
	AlertSeverityCritical AlertSeverity = "critical"
	AlertSeverityHigh     AlertSeverity = "high"
	AlertSeverityMedium   AlertSeverity = "medium"
	AlertSeverityLow      AlertSeverity = "low"
)

// NewTrendAnalyzer creates a new trend analyzer
func NewTrendAnalyzer(logger *zap.Logger, dataDir string) *TrendAnalyzer {
	return &TrendAnalyzer{
		logger:  logger,
		dataDir: dataDir,
		alertThresholds: AlertThresholds{
			QualityDropThreshold:   10.0, // 10% drop
			MinimumQualityScore:    60.0,
			MaxComplexity:          15.0,
			ComplexityIncreaseRate: 20.0, // 20% increase
			MinimumCoverage:        70.0,
			CoverageDropThreshold:  10.0, // 10% drop
			MaxDebtRatio:           50.0,
			DebtIncreaseRate:       30.0, // 30% increase
			MinSecurityScore:       80.0,
			MaxCriticalIssues:      0,
		},
		alertHandlers: make([]AlertHandler, 0),
	}
}

// SetThresholds sets custom alert thresholds
func (ta *TrendAnalyzer) SetThresholds(thresholds AlertThresholds) {
	ta.alertThresholds = thresholds
}

// AddAlertHandler adds an alert handler
func (ta *TrendAnalyzer) AddAlertHandler(handler AlertHandler) {
	ta.alertHandlers = append(ta.alertHandlers, handler)
}

// AnalyzeTrends analyzes quality trends and generates alerts
func (ta *TrendAnalyzer) AnalyzeTrends(ctx context.Context, currentResult *QualityResult) ([]*QualityAlert, error) {
	// Load historical data
	historicalData, err := ta.loadHistoricalData(currentResult.Repository)
	if err != nil {
		return nil, fmt.Errorf("failed to load historical data: %w", err)
	}

	// Add current result to history
	historicalData = append(historicalData, currentResult)

	// Sort by timestamp
	sort.Slice(historicalData, func(i, j int) bool {
		return historicalData[i].Timestamp.Before(historicalData[j].Timestamp)
	})

	alerts := make([]*QualityAlert, 0)

	// Check for quality drops
	if alert := ta.checkQualityDrop(historicalData); alert != nil {
		alerts = append(alerts, alert)
	}

	// Check complexity trends
	if alert := ta.checkComplexityTrend(historicalData); alert != nil {
		alerts = append(alerts, alert)
	}

	// Check coverage trends
	if alert := ta.checkCoverageTrend(historicalData); alert != nil {
		alerts = append(alerts, alert)
	}

	// Check technical debt trends
	if alert := ta.checkDebtTrend(historicalData); alert != nil {
		alerts = append(alerts, alert)
	}

	// Check security issues
	if alert := ta.checkSecurityIssues(currentResult); alert != nil {
		alerts = append(alerts, alert)
	}

	// Detect anomalies using statistical analysis
	anomalyAlerts := ta.detectAnomalies(historicalData)
	alerts = append(alerts, anomalyAlerts...)

	// Process alerts through handlers
	for _, alert := range alerts {
		for _, handler := range ta.alertHandlers {
			if err := handler.HandleAlert(ctx, alert); err != nil {
				ta.logger.Warn("Alert handler failed",
					zap.String("alert_id", alert.ID),
					zap.Error(err))
			}
		}
	}

	// Save alerts
	if err := ta.saveAlerts(currentResult.Repository, alerts); err != nil {
		ta.logger.Warn("Failed to save alerts", zap.Error(err))
	}

	return alerts, nil
}

// checkQualityDrop checks for significant quality score drops
func (ta *TrendAnalyzer) checkQualityDrop(history []*QualityResult) *QualityAlert {
	if len(history) < 2 {
		return nil
	}

	current := history[len(history)-1]
	previous := history[len(history)-2]

	// Check absolute minimum
	if current.OverallScore < ta.alertThresholds.MinimumQualityScore {
		return &QualityAlert{
			ID:         generateAlertID(),
			Type:       AlertTypeQualityDrop,
			Severity:   AlertSeverityCritical,
			Timestamp:  time.Now(),
			Repository: current.Repository,
			Message: fmt.Sprintf("Quality score critically low: %.1f%% (minimum: %.1f%%)",
				current.OverallScore, ta.alertThresholds.MinimumQualityScore),
			Details: map[string]interface{}{
				"current_score": current.OverallScore,
				"minimum_score": ta.alertThresholds.MinimumQualityScore,
				"issues_count":  len(current.Issues),
			},
			Suggestions: []string{
				"Immediately address critical and major issues",
				"Focus on reducing code complexity",
				"Improve test coverage for critical components",
			},
		}
	}

	// Check percentage drop
	drop := previous.OverallScore - current.OverallScore
	dropPercentage := (drop / previous.OverallScore) * 100

	if dropPercentage > ta.alertThresholds.QualityDropThreshold {
		return &QualityAlert{
			ID:         generateAlertID(),
			Type:       AlertTypeQualityDrop,
			Severity:   AlertSeverityHigh,
			Timestamp:  time.Now(),
			Repository: current.Repository,
			Message: fmt.Sprintf("Quality score dropped by %.1f%% (from %.1f%% to %.1f%%)",
				dropPercentage, previous.OverallScore, current.OverallScore),
			Details: map[string]interface{}{
				"previous_score":  previous.OverallScore,
				"current_score":   current.OverallScore,
				"drop_percentage": dropPercentage,
			},
			Suggestions: []string{
				"Review recent code changes for quality issues",
				"Run full code review on recent commits",
				"Check if new technical debt was introduced",
			},
		}
	}

	return nil
}

// checkComplexityTrend checks for complexity issues
func (ta *TrendAnalyzer) checkComplexityTrend(history []*QualityResult) *QualityAlert {
	if len(history) < 1 {
		return nil
	}

	current := history[len(history)-1]

	// Check absolute maximum
	if current.Metrics.AvgComplexity > ta.alertThresholds.MaxComplexity {
		return &QualityAlert{
			ID:         generateAlertID(),
			Type:       AlertTypeComplexityHigh,
			Severity:   AlertSeverityHigh,
			Timestamp:  time.Now(),
			Repository: current.Repository,
			Message: fmt.Sprintf("Code complexity too high: %.1f (maximum: %.1f)",
				current.Metrics.AvgComplexity, ta.alertThresholds.MaxComplexity),
			Details: map[string]interface{}{
				"current_complexity": current.Metrics.AvgComplexity,
				"max_complexity":     ta.alertThresholds.MaxComplexity,
			},
			Suggestions: []string{
				"Refactor complex functions into smaller units",
				"Apply SOLID principles to reduce complexity",
				"Consider extracting complex logic into separate modules",
			},
		}
	}

	// Check increase rate
	if len(history) >= 5 {
		// Calculate trend over last 5 data points
		startIdx := len(history) - 5
		startComplexity := history[startIdx].Metrics.AvgComplexity

		if startComplexity > 0 {
			increaseRate := ((current.Metrics.AvgComplexity - startComplexity) / startComplexity) * 100
			if increaseRate > ta.alertThresholds.ComplexityIncreaseRate {
				return &QualityAlert{
					ID:         generateAlertID(),
					Type:       AlertTypeComplexityHigh,
					Severity:   AlertSeverityMedium,
					Timestamp:  time.Now(),
					Repository: current.Repository,
					Message:    fmt.Sprintf("Complexity increasing rapidly: %.1f%% increase", increaseRate),
					Details: map[string]interface{}{
						"start_complexity":   startComplexity,
						"current_complexity": current.Metrics.AvgComplexity,
						"increase_rate":      increaseRate,
					},
					Suggestions: []string{
						"Review architectural decisions",
						"Consider code simplification sprint",
						"Implement complexity budget for new features",
					},
				}
			}
		}
	}

	return nil
}

// checkCoverageTrend checks for test coverage issues
func (ta *TrendAnalyzer) checkCoverageTrend(history []*QualityResult) *QualityAlert {
	if len(history) < 1 {
		return nil
	}

	current := history[len(history)-1]

	// Check absolute minimum
	if current.Metrics.TestCoverage < ta.alertThresholds.MinimumCoverage {
		severity := AlertSeverityMedium
		if current.Metrics.TestCoverage < 50 {
			severity = AlertSeverityHigh
		}

		return &QualityAlert{
			ID:         generateAlertID(),
			Type:       AlertTypeCoverageLow,
			Severity:   severity,
			Timestamp:  time.Now(),
			Repository: current.Repository,
			Message: fmt.Sprintf("Test coverage below minimum: %.1f%% (minimum: %.1f%%)",
				current.Metrics.TestCoverage, ta.alertThresholds.MinimumCoverage),
			Details: map[string]interface{}{
				"current_coverage": current.Metrics.TestCoverage,
				"minimum_coverage": ta.alertThresholds.MinimumCoverage,
			},
			Suggestions: []string{
				"Add unit tests for uncovered code",
				"Focus on testing critical business logic",
				"Set up coverage gates in CI/CD pipeline",
			},
		}
	}

	// Check coverage drop
	if len(history) >= 2 {
		previous := history[len(history)-2]
		drop := previous.Metrics.TestCoverage - current.Metrics.TestCoverage

		if drop > ta.alertThresholds.CoverageDropThreshold {
			return &QualityAlert{
				ID:         generateAlertID(),
				Type:       AlertTypeCoverageLow,
				Severity:   AlertSeverityMedium,
				Timestamp:  time.Now(),
				Repository: current.Repository,
				Message:    fmt.Sprintf("Test coverage dropped by %.1f%%", drop),
				Details: map[string]interface{}{
					"previous_coverage": previous.Metrics.TestCoverage,
					"current_coverage":  current.Metrics.TestCoverage,
					"coverage_drop":     drop,
				},
				Suggestions: []string{
					"Ensure new code includes tests",
					"Review if tests were accidentally removed",
					"Add tests for recently added features",
				},
			}
		}
	}

	return nil
}

// checkDebtTrend checks for technical debt issues
func (ta *TrendAnalyzer) checkDebtTrend(history []*QualityResult) *QualityAlert {
	if len(history) < 1 {
		return nil
	}

	current := history[len(history)-1]

	// Check absolute maximum
	if current.Metrics.TechnicalDebtRatio > ta.alertThresholds.MaxDebtRatio {
		return &QualityAlert{
			ID:         generateAlertID(),
			Type:       AlertTypeDebtIncrease,
			Severity:   AlertSeverityHigh,
			Timestamp:  time.Now(),
			Repository: current.Repository,
			Message: fmt.Sprintf("Technical debt ratio too high: %.1f (maximum: %.1f)",
				current.Metrics.TechnicalDebtRatio, ta.alertThresholds.MaxDebtRatio),
			Details: map[string]interface{}{
				"current_debt_ratio": current.Metrics.TechnicalDebtRatio,
				"max_debt_ratio":     ta.alertThresholds.MaxDebtRatio,
				"total_debt_minutes": current.TechnicalDebt.TotalMinutes,
			},
			Suggestions: []string{
				"Schedule technical debt reduction sprint",
				"Prioritize fixing high-severity issues",
				"Allocate time for refactoring in each sprint",
			},
		}
	}

	// Check increase rate
	if len(history) >= 3 {
		// Calculate average over last 3 data points
		avgDebt := 0.0
		for i := len(history) - 3; i < len(history)-1; i++ {
			avgDebt += history[i].Metrics.TechnicalDebtRatio
		}
		avgDebt /= 2.0

		if avgDebt > 0 {
			increaseRate := ((current.Metrics.TechnicalDebtRatio - avgDebt) / avgDebt) * 100
			if increaseRate > ta.alertThresholds.DebtIncreaseRate {
				return &QualityAlert{
					ID:         generateAlertID(),
					Type:       AlertTypeDebtIncrease,
					Severity:   AlertSeverityMedium,
					Timestamp:  time.Now(),
					Repository: current.Repository,
					Message:    fmt.Sprintf("Technical debt increasing rapidly: %.1f%% increase", increaseRate),
					Details: map[string]interface{}{
						"average_debt":  avgDebt,
						"current_debt":  current.Metrics.TechnicalDebtRatio,
						"increase_rate": increaseRate,
					},
					Suggestions: []string{
						"Review development practices",
						"Implement stricter code review process",
						"Consider debt ceiling policy",
					},
				}
			}
		}
	}

	return nil
}

// checkSecurityIssues checks for security-related alerts
func (ta *TrendAnalyzer) checkSecurityIssues(result *QualityResult) *QualityAlert {
	// Check security score
	if result.Metrics.SecurityScore < ta.alertThresholds.MinSecurityScore {
		return &QualityAlert{
			ID:         generateAlertID(),
			Type:       AlertTypeSecurityIssue,
			Severity:   AlertSeverityCritical,
			Timestamp:  time.Now(),
			Repository: result.Repository,
			Message: fmt.Sprintf("Security score below threshold: %.1f%% (minimum: %.1f%%)",
				result.Metrics.SecurityScore, ta.alertThresholds.MinSecurityScore),
			Details: map[string]interface{}{
				"security_score":  result.Metrics.SecurityScore,
				"security_issues": len(result.SecurityIssues),
				"critical_issues": countCriticalSecurityIssues(result),
			},
			Suggestions: []string{
				"Immediately fix critical security vulnerabilities",
				"Run security audit on the codebase",
				"Update dependencies with known vulnerabilities",
			},
		}
	}

	// Check critical issues count
	criticalCount := countCriticalSecurityIssues(result)
	if criticalCount > ta.alertThresholds.MaxCriticalIssues {
		return &QualityAlert{
			ID:         generateAlertID(),
			Type:       AlertTypeSecurityIssue,
			Severity:   AlertSeverityCritical,
			Timestamp:  time.Now(),
			Repository: result.Repository,
			Message:    fmt.Sprintf("Critical security issues found: %d", criticalCount),
			Details: map[string]interface{}{
				"critical_count": criticalCount,
				"total_security": len(result.SecurityIssues),
			},
			Suggestions: []string{
				"Fix critical security issues immediately",
				"Review security best practices",
				"Consider security-focused code review",
			},
		}
	}

	return nil
}

// detectAnomalies uses statistical analysis to detect anomalies
func (ta *TrendAnalyzer) detectAnomalies(history []*QualityResult) []*QualityAlert {
	alerts := make([]*QualityAlert, 0)

	if len(history) < 10 {
		return alerts // Need sufficient data for statistical analysis
	}

	// Analyze various metrics for anomalies
	metrics := []struct {
		name      string
		extractor func(*QualityResult) float64
	}{
		{"quality_score", func(r *QualityResult) float64 { return r.OverallScore }},
		{"complexity", func(r *QualityResult) float64 { return r.Metrics.AvgComplexity }},
		{"coverage", func(r *QualityResult) float64 { return r.Metrics.TestCoverage }},
		{"debt_ratio", func(r *QualityResult) float64 { return r.Metrics.TechnicalDebtRatio }},
	}

	for _, metric := range metrics {
		values := make([]float64, len(history))
		for i, result := range history {
			values[i] = metric.extractor(result)
		}

		// Calculate statistics
		mean, stdDev := calculateStats(values)
		current := values[len(values)-1]

		// Check if current value is an outlier (3 standard deviations)
		zScore := math.Abs(current-mean) / stdDev
		if zScore > 3 {
			alert := &QualityAlert{
				ID:         generateAlertID(),
				Type:       AlertTypeTrendAnomaly,
				Severity:   AlertSeverityMedium,
				Timestamp:  time.Now(),
				Repository: history[len(history)-1].Repository,
				Message:    fmt.Sprintf("Anomaly detected in %s: %.2f (z-score: %.2f)", metric.name, current, zScore),
				Details: map[string]interface{}{
					"metric":        metric.name,
					"current_value": current,
					"mean":          mean,
					"std_dev":       stdDev,
					"z_score":       zScore,
				},
				Suggestions: []string{
					"Investigate recent changes that might have caused the anomaly",
					"Review if this is an expected change or an error",
					"Check for data collection issues",
				},
			}
			alerts = append(alerts, alert)
		}
	}

	return alerts
}

// loadHistoricalData loads historical quality data
func (ta *TrendAnalyzer) loadHistoricalData(repository string) ([]*QualityResult, error) {
	historyDir := filepath.Join(ta.dataDir, "history")
	results := make([]*QualityResult, 0)

	files, err := os.ReadDir(historyDir)
	if err != nil {
		if os.IsNotExist(err) {
			return results, nil
		}
		return nil, err
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(historyDir, file.Name()))
		if err != nil {
			ta.logger.Warn("Failed to read history file",
				zap.String("file", file.Name()),
				zap.Error(err))
			continue
		}

		var result QualityResult
		if err := json.Unmarshal(data, &result); err != nil {
			ta.logger.Warn("Failed to parse history file",
				zap.String("file", file.Name()),
				zap.Error(err))
			continue
		}

		if result.Repository == repository {
			results = append(results, &result)
		}
	}

	return results, nil
}

// saveAlerts saves alerts to file
func (ta *TrendAnalyzer) saveAlerts(repository string, alerts []*QualityAlert) error {
	if len(alerts) == 0 {
		return nil
	}

	alertsDir := filepath.Join(ta.dataDir, "alerts")
	if err := os.MkdirAll(alertsDir, 0o755); err != nil {
		return err
	}

	filename := fmt.Sprintf("alerts-%s-%s.json",
		sanitizeFilename(repository),
		time.Now().Format("20060102-150405"))

	data, err := json.MarshalIndent(alerts, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(alertsDir, filename), data, 0o644)
}

// Helper functions

func generateAlertID() string {
	return fmt.Sprintf("alert-%d", time.Now().UnixNano())
}

func countCriticalSecurityIssues(result *QualityResult) int {
	count := 0
	for _, issue := range result.SecurityIssues {
		if issue.Severity == "critical" || issue.CVSS >= 9.0 {
			count++
		}
	}
	return count
}

func calculateStats(values []float64) (mean, stdDev float64) {
	if len(values) == 0 {
		return 0, 0
	}

	// Calculate mean
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean = sum / float64(len(values))

	// Calculate standard deviation
	variance := 0.0
	for _, v := range values {
		variance += math.Pow(v-mean, 2)
	}
	variance /= float64(len(values))
	stdDev = math.Sqrt(variance)

	return mean, stdDev
}

func sanitizeFilename(s string) string {
	// Replace problematic characters
	replacer := strings.NewReplacer(
		"/", "-",
		"\\", "-",
		":", "-",
		"*", "-",
		"?", "-",
		"\"", "-",
		"<", "-",
		">", "-",
		"|", "-",
		" ", "_",
	)
	return replacer.Replace(s)
}
