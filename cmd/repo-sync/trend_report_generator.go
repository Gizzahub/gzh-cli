package reposync

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	"time"

	"go.uber.org/zap"
)

// TrendReportGenerator generates trend analysis reports
type TrendReportGenerator struct {
	logger    *zap.Logger
	outputDir string
}

// NewTrendReportGenerator creates a new trend report generator
func NewTrendReportGenerator(logger *zap.Logger, outputDir string) *TrendReportGenerator {
	return &TrendReportGenerator{
		logger:    logger,
		outputDir: outputDir,
	}
}

// TrendReport represents a complete trend analysis report
type TrendReport struct {
	Repository      string                `json:"repository"`
	GeneratedAt     time.Time             `json:"generated_at"`
	Period          ReportPeriod          `json:"period"`
	Summary         TrendSummary          `json:"summary"`
	QualityTrends   QualityTrendData      `json:"quality_trends"`
	Alerts          []*QualityAlert       `json:"alerts"`
	Improvements    []ImprovementItem     `json:"improvements"`
	Degradations    []DegradationItem     `json:"degradations"`
	Predictions     TrendPredictions      `json:"predictions"`
	Recommendations []TrendRecommendation `json:"recommendations"`
}

// ReportPeriod represents the time period for the report
type ReportPeriod struct {
	Start    time.Time `json:"start"`
	End      time.Time `json:"end"`
	Duration string    `json:"duration"`
}

// TrendSummary provides high-level trend summary
type TrendSummary struct {
	DataPoints          int     `json:"data_points"`
	OverallTrend        string  `json:"overall_trend"` // improving, stable, degrading
	QualityChange       float64 `json:"quality_change"`
	ComplexityChange    float64 `json:"complexity_change"`
	CoverageChange      float64 `json:"coverage_change"`
	DebtChange          float64 `json:"debt_change"`
	AlertsGenerated     int     `json:"alerts_generated"`
	CriticalAlertsCount int     `json:"critical_alerts_count"`
}

// QualityTrendData contains detailed trend data
type QualityTrendData struct {
	Timestamps       []time.Time `json:"timestamps"`
	QualityScores    []float64   `json:"quality_scores"`
	ComplexityValues []float64   `json:"complexity_values"`
	CoverageValues   []float64   `json:"coverage_values"`
	DebtRatios       []float64   `json:"debt_ratios"`
	IssuesCounts     []int       `json:"issues_counts"`
}

// ImprovementItem represents an area of improvement
type ImprovementItem struct {
	Metric          string  `json:"metric"`
	StartValue      float64 `json:"start_value"`
	EndValue        float64 `json:"end_value"`
	ImprovementRate float64 `json:"improvement_rate"`
	Description     string  `json:"description"`
}

// DegradationItem represents an area of degradation
type DegradationItem struct {
	Metric          string  `json:"metric"`
	StartValue      float64 `json:"start_value"`
	EndValue        float64 `json:"end_value"`
	DegradationRate float64 `json:"degradation_rate"`
	Description     string  `json:"description"`
}

// TrendPredictions contains predicted future trends
type TrendPredictions struct {
	QualityPrediction    PredictionData `json:"quality_prediction"`
	ComplexityPrediction PredictionData `json:"complexity_prediction"`
	DebtPrediction       PredictionData `json:"debt_prediction"`
}

// PredictionData represents a single prediction
type PredictionData struct {
	NextValue      float64    `json:"next_value"`
	Confidence     float64    `json:"confidence"`
	TrendDirection string     `json:"trend_direction"`
	ReachThreshold *time.Time `json:"reach_threshold,omitempty"`
}

// TrendRecommendation provides actionable recommendations
type TrendRecommendation struct {
	Priority    string   `json:"priority"`
	Category    string   `json:"category"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Actions     []string `json:"actions"`
}

// GenerateTrendReport generates a comprehensive trend report
func (trg *TrendReportGenerator) GenerateTrendReport(repository string, history []*QualityResult, alerts []*QualityAlert) (*TrendReport, error) {
	if len(history) < 2 {
		return nil, fmt.Errorf("insufficient data for trend analysis (need at least 2 data points)")
	}

	// Sort history by timestamp
	sort.Slice(history, func(i, j int) bool {
		return history[i].Timestamp.Before(history[j].Timestamp)
	})

	report := &TrendReport{
		Repository:  repository,
		GeneratedAt: time.Now(),
		Period: ReportPeriod{
			Start:    history[0].Timestamp,
			End:      history[len(history)-1].Timestamp,
			Duration: history[len(history)-1].Timestamp.Sub(history[0].Timestamp).String(),
		},
		Alerts: alerts,
	}

	// Extract trend data
	report.QualityTrends = trg.extractTrendData(history)

	// Calculate summary
	report.Summary = trg.calculateSummary(history, alerts)

	// Identify improvements and degradations
	report.Improvements = trg.identifyImprovements(history)
	report.Degradations = trg.identifyDegradations(history)

	// Generate predictions
	report.Predictions = trg.generatePredictions(history)

	// Generate recommendations
	report.Recommendations = trg.generateRecommendations(report)

	// Save report
	if err := trg.saveReport(report); err != nil {
		return nil, fmt.Errorf("failed to save report: %w", err)
	}

	// Generate HTML report
	if err := trg.generateHTMLReport(report); err != nil {
		trg.logger.Warn("Failed to generate HTML report", zap.Error(err))
	}

	return report, nil
}

// extractTrendData extracts trend data from history
func (trg *TrendReportGenerator) extractTrendData(history []*QualityResult) QualityTrendData {
	data := QualityTrendData{
		Timestamps:       make([]time.Time, len(history)),
		QualityScores:    make([]float64, len(history)),
		ComplexityValues: make([]float64, len(history)),
		CoverageValues:   make([]float64, len(history)),
		DebtRatios:       make([]float64, len(history)),
		IssuesCounts:     make([]int, len(history)),
	}

	for i, result := range history {
		data.Timestamps[i] = result.Timestamp
		data.QualityScores[i] = result.OverallScore
		data.ComplexityValues[i] = result.Metrics.AvgComplexity
		data.CoverageValues[i] = result.Metrics.TestCoverage
		data.DebtRatios[i] = result.Metrics.TechnicalDebtRatio
		data.IssuesCounts[i] = len(result.Issues)
	}

	return data
}

// calculateSummary calculates trend summary
func (trg *TrendReportGenerator) calculateSummary(history []*QualityResult, alerts []*QualityAlert) TrendSummary {
	first := history[0]
	last := history[len(history)-1]

	summary := TrendSummary{
		DataPoints:       len(history),
		QualityChange:    last.OverallScore - first.OverallScore,
		ComplexityChange: last.Metrics.AvgComplexity - first.Metrics.AvgComplexity,
		CoverageChange:   last.Metrics.TestCoverage - first.Metrics.TestCoverage,
		DebtChange:       last.Metrics.TechnicalDebtRatio - first.Metrics.TechnicalDebtRatio,
		AlertsGenerated:  len(alerts),
	}

	// Count critical alerts
	for _, alert := range alerts {
		if alert.Severity == AlertSeverityCritical {
			summary.CriticalAlertsCount++
		}
	}

	// Determine overall trend
	if summary.QualityChange > 5 && summary.ComplexityChange < 0 && summary.DebtChange < 0 {
		summary.OverallTrend = "improving"
	} else if summary.QualityChange < -5 || summary.ComplexityChange > 2 || summary.DebtChange > 10 {
		summary.OverallTrend = "degrading"
	} else {
		summary.OverallTrend = "stable"
	}

	return summary
}

// identifyImprovements identifies areas of improvement
func (trg *TrendReportGenerator) identifyImprovements(history []*QualityResult) []ImprovementItem {
	improvements := make([]ImprovementItem, 0)

	if len(history) < 2 {
		return improvements
	}

	first := history[0]
	last := history[len(history)-1]

	// Check quality score improvement
	if last.OverallScore > first.OverallScore {
		improvements = append(improvements, ImprovementItem{
			Metric:          "Quality Score",
			StartValue:      first.OverallScore,
			EndValue:        last.OverallScore,
			ImprovementRate: ((last.OverallScore - first.OverallScore) / first.OverallScore) * 100,
			Description:     fmt.Sprintf("Code quality improved by %.1f%%", last.OverallScore-first.OverallScore),
		})
	}

	// Check complexity reduction
	if last.Metrics.AvgComplexity < first.Metrics.AvgComplexity {
		improvements = append(improvements, ImprovementItem{
			Metric:          "Code Complexity",
			StartValue:      first.Metrics.AvgComplexity,
			EndValue:        last.Metrics.AvgComplexity,
			ImprovementRate: ((first.Metrics.AvgComplexity - last.Metrics.AvgComplexity) / first.Metrics.AvgComplexity) * 100,
			Description:     "Code complexity reduced, making code easier to maintain",
		})
	}

	// Check coverage improvement
	if last.Metrics.TestCoverage > first.Metrics.TestCoverage {
		improvements = append(improvements, ImprovementItem{
			Metric:          "Test Coverage",
			StartValue:      first.Metrics.TestCoverage,
			EndValue:        last.Metrics.TestCoverage,
			ImprovementRate: ((last.Metrics.TestCoverage - first.Metrics.TestCoverage) / first.Metrics.TestCoverage) * 100,
			Description:     fmt.Sprintf("Test coverage increased by %.1f%%", last.Metrics.TestCoverage-first.Metrics.TestCoverage),
		})
	}

	return improvements
}

// identifyDegradations identifies areas of degradation
func (trg *TrendReportGenerator) identifyDegradations(history []*QualityResult) []DegradationItem {
	degradations := make([]DegradationItem, 0)

	if len(history) < 2 {
		return degradations
	}

	first := history[0]
	last := history[len(history)-1]

	// Check quality score degradation
	if last.OverallScore < first.OverallScore {
		degradations = append(degradations, DegradationItem{
			Metric:          "Quality Score",
			StartValue:      first.OverallScore,
			EndValue:        last.OverallScore,
			DegradationRate: ((first.OverallScore - last.OverallScore) / first.OverallScore) * 100,
			Description:     fmt.Sprintf("Code quality decreased by %.1f%%", first.OverallScore-last.OverallScore),
		})
	}

	// Check complexity increase
	if last.Metrics.AvgComplexity > first.Metrics.AvgComplexity {
		degradations = append(degradations, DegradationItem{
			Metric:          "Code Complexity",
			StartValue:      first.Metrics.AvgComplexity,
			EndValue:        last.Metrics.AvgComplexity,
			DegradationRate: ((last.Metrics.AvgComplexity - first.Metrics.AvgComplexity) / first.Metrics.AvgComplexity) * 100,
			Description:     "Code complexity increased, making code harder to maintain",
		})
	}

	// Check debt increase
	if last.Metrics.TechnicalDebtRatio > first.Metrics.TechnicalDebtRatio {
		degradations = append(degradations, DegradationItem{
			Metric:          "Technical Debt",
			StartValue:      first.Metrics.TechnicalDebtRatio,
			EndValue:        last.Metrics.TechnicalDebtRatio,
			DegradationRate: ((last.Metrics.TechnicalDebtRatio - first.Metrics.TechnicalDebtRatio) / first.Metrics.TechnicalDebtRatio) * 100,
			Description:     "Technical debt increased, requiring more effort to fix issues",
		})
	}

	return degradations
}

// generatePredictions generates trend predictions using simple linear regression
func (trg *TrendReportGenerator) generatePredictions(history []*QualityResult) TrendPredictions {
	predictions := TrendPredictions{}

	// Predict quality score
	qualityValues := make([]float64, len(history))
	for i, result := range history {
		qualityValues[i] = result.OverallScore
	}
	predictions.QualityPrediction = trg.predictNextValue(qualityValues, "quality")

	// Predict complexity
	complexityValues := make([]float64, len(history))
	for i, result := range history {
		complexityValues[i] = result.Metrics.AvgComplexity
	}
	predictions.ComplexityPrediction = trg.predictNextValue(complexityValues, "complexity")

	// Predict debt
	debtValues := make([]float64, len(history))
	for i, result := range history {
		debtValues[i] = result.Metrics.TechnicalDebtRatio
	}
	predictions.DebtPrediction = trg.predictNextValue(debtValues, "debt")

	return predictions
}

// predictNextValue uses simple linear regression to predict next value
func (trg *TrendReportGenerator) predictNextValue(values []float64, metricType string) PredictionData {
	n := float64(len(values))
	if n < 2 {
		return PredictionData{
			NextValue:      values[len(values)-1],
			Confidence:     0.0,
			TrendDirection: "unknown",
		}
	}

	// Calculate linear regression
	sumX, sumY, sumXY, sumX2 := 0.0, 0.0, 0.0, 0.0
	for i, y := range values {
		x := float64(i)
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	// Calculate slope and intercept
	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
	intercept := (sumY - slope*sumX) / n

	// Predict next value
	nextX := n
	nextValue := slope*nextX + intercept

	// Calculate R-squared for confidence
	meanY := sumY / n
	ssTotal, ssRes := 0.0, 0.0
	for i, y := range values {
		predicted := slope*float64(i) + intercept
		ssTotal += (y - meanY) * (y - meanY)
		ssRes += (y - predicted) * (y - predicted)
	}
	rSquared := 1 - (ssRes / ssTotal)
	confidence := rSquared * 100

	// Determine trend direction
	var trendDirection string
	if slope > 0.1 {
		trendDirection = "increasing"
	} else if slope < -0.1 {
		trendDirection = "decreasing"
	} else {
		trendDirection = "stable"
	}

	return PredictionData{
		NextValue:      nextValue,
		Confidence:     confidence,
		TrendDirection: trendDirection,
	}
}

// generateRecommendations generates actionable recommendations
func (trg *TrendReportGenerator) generateRecommendations(report *TrendReport) []TrendRecommendation {
	recommendations := make([]TrendRecommendation, 0)

	// Check overall trend
	if report.Summary.OverallTrend == "degrading" {
		recommendations = append(recommendations, TrendRecommendation{
			Priority:    "high",
			Category:    "quality",
			Title:       "Reverse Quality Degradation",
			Description: "Code quality is showing a degrading trend that needs immediate attention",
			Actions: []string{
				"Conduct thorough code review of recent changes",
				"Allocate dedicated time for refactoring",
				"Implement stricter quality gates in CI/CD",
			},
		})
	}

	// Check critical alerts
	if report.Summary.CriticalAlertsCount > 0 {
		recommendations = append(recommendations, TrendRecommendation{
			Priority:    "critical",
			Category:    "alerts",
			Title:       "Address Critical Alerts",
			Description: fmt.Sprintf("%d critical alerts require immediate action", report.Summary.CriticalAlertsCount),
			Actions: []string{
				"Review and fix all critical security vulnerabilities",
				"Address code sections causing critical quality drops",
				"Establish emergency response procedures for critical alerts",
			},
		})
	}

	// Check predictions
	if report.Predictions.ComplexityPrediction.TrendDirection == "increasing" &&
		report.Predictions.ComplexityPrediction.Confidence > 70 {
		recommendations = append(recommendations, TrendRecommendation{
			Priority:    "medium",
			Category:    "complexity",
			Title:       "Control Complexity Growth",
			Description: "Code complexity is predicted to continue increasing",
			Actions: []string{
				"Review and simplify complex functions",
				"Apply SOLID principles more rigorously",
				"Consider architectural improvements",
			},
		})
	}

	// Check technical debt
	if report.Summary.DebtChange > 20 {
		recommendations = append(recommendations, TrendRecommendation{
			Priority:    "high",
			Category:    "debt",
			Title:       "Manage Technical Debt",
			Description: fmt.Sprintf("Technical debt increased by %.1f%%", report.Summary.DebtChange),
			Actions: []string{
				"Schedule dedicated debt reduction sprints",
				"Prioritize fixing high-impact issues",
				"Implement debt ceiling policies",
			},
		})
	}

	return recommendations
}

// saveReport saves the report to file
func (trg *TrendReportGenerator) saveReport(report *TrendReport) error {
	if err := os.MkdirAll(trg.outputDir, 0o755); err != nil {
		return err
	}

	filename := fmt.Sprintf("trend-report-%s-%s.json",
		sanitizeFilename(report.Repository),
		report.GeneratedAt.Format("20060102-150405"))

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(trg.outputDir, filename), data, 0o644)
}

// generateHTMLReport generates an HTML version of the report
func (trg *TrendReportGenerator) generateHTMLReport(report *TrendReport) error {
	tmpl := trg.getHTMLTemplate()

	t, err := template.New("trend-report").Funcs(template.FuncMap{
		"json": func(v interface{}) template.JS {
			b, _ := json.Marshal(v)
			return template.JS(b)
		},
		"formatFloat": func(f float64) string {
			return fmt.Sprintf("%.1f", f)
		},
		"formatTime": func(t time.Time) string {
			return t.Format("2006-01-02 15:04:05")
		},
		"severityColor": func(severity string) string {
			switch severity {
			case "critical":
				return "danger"
			case "high":
				return "warning"
			case "medium":
				return "info"
			default:
				return "secondary"
			}
		},
	}).Parse(tmpl)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, report); err != nil {
		return err
	}

	filename := fmt.Sprintf("trend-report-%s-%s.html",
		sanitizeFilename(report.Repository),
		report.GeneratedAt.Format("20060102-150405"))

	return os.WriteFile(filepath.Join(trg.outputDir, filename), buf.Bytes(), 0o644)
}

// getHTMLTemplate returns the HTML template for trend reports
func (trg *TrendReportGenerator) getHTMLTemplate() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Quality Trend Report - {{.Repository}}</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/css/bootstrap.min.css" rel="stylesheet">
    <script src="https://cdn.jsdelivr.net/npm/chart.js@3.7.0/dist/chart.min.js"></script>
    <style>
        .trend-card {
            border-left: 4px solid;
            margin-bottom: 1rem;
        }
        .trend-improving { border-left-color: #28a745; }
        .trend-stable { border-left-color: #17a2b8; }
        .trend-degrading { border-left-color: #dc3545; }
        .metric-change {
            font-size: 1.2rem;
            font-weight: bold;
        }
        .change-positive { color: #28a745; }
        .change-negative { color: #dc3545; }
        .prediction-card {
            background-color: #f8f9fa;
            border-radius: 8px;
            padding: 1rem;
        }
    </style>
</head>
<body>
    <div class="container-fluid mt-4">
        <h1>Quality Trend Report</h1>
        <p class="text-muted">
            Repository: <strong>{{.Repository}}</strong> | 
            Generated: {{formatTime .GeneratedAt}} |
            Period: {{formatTime .Period.Start}} to {{formatTime .Period.End}}
        </p>

        <!-- Summary Section -->
        <div class="row mb-4">
            <div class="col-12">
                <div class="card trend-card trend-{{.Summary.OverallTrend}}">
                    <div class="card-body">
                        <h3>Trend Summary</h3>
                        <div class="row">
                            <div class="col-md-3">
                                <h5>Overall Trend</h5>
                                <p class="text-capitalize"><strong>{{.Summary.OverallTrend}}</strong></p>
                            </div>
                            <div class="col-md-3">
                                <h5>Quality Change</h5>
                                <p class="metric-change {{if ge .Summary.QualityChange 0}}change-positive{{else}}change-negative{{end}}">
                                    {{if ge .Summary.QualityChange 0}}+{{end}}{{formatFloat .Summary.QualityChange}}%
                                </p>
                            </div>
                            <div class="col-md-3">
                                <h5>Complexity Change</h5>
                                <p class="metric-change {{if le .Summary.ComplexityChange 0}}change-positive{{else}}change-negative{{end}}">
                                    {{if ge .Summary.ComplexityChange 0}}+{{end}}{{formatFloat .Summary.ComplexityChange}}
                                </p>
                            </div>
                            <div class="col-md-3">
                                <h5>Alerts Generated</h5>
                                <p><strong>{{.Summary.AlertsGenerated}}</strong> 
                                {{if gt .Summary.CriticalAlertsCount 0}}
                                    <span class="text-danger">({{.Summary.CriticalAlertsCount}} critical)</span>
                                {{end}}
                                </p>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <!-- Trend Charts -->
        <div class="row mb-4">
            <div class="col-md-6">
                <div class="card">
                    <div class="card-header">
                        <h5 class="mb-0">Quality Score Trend</h5>
                    </div>
                    <div class="card-body">
                        <canvas id="qualityTrendChart" height="300"></canvas>
                    </div>
                </div>
            </div>
            <div class="col-md-6">
                <div class="card">
                    <div class="card-header">
                        <h5 class="mb-0">Complexity & Debt Trends</h5>
                    </div>
                    <div class="card-body">
                        <canvas id="complexityDebtChart" height="300"></canvas>
                    </div>
                </div>
            </div>
        </div>

        <!-- Improvements and Degradations -->
        <div class="row mb-4">
            {{if .Improvements}}
            <div class="col-md-6">
                <div class="card">
                    <div class="card-header bg-success text-white">
                        <h5 class="mb-0">Improvements</h5>
                    </div>
                    <div class="card-body">
                        {{range .Improvements}}
                        <div class="mb-3">
                            <h6>{{.Metric}}</h6>
                            <p class="mb-1">{{.Description}}</p>
                            <small class="text-muted">
                                {{formatFloat .StartValue}} → {{formatFloat .EndValue}} 
                                ({{formatFloat .ImprovementRate}}% improvement)
                            </small>
                        </div>
                        {{end}}
                    </div>
                </div>
            </div>
            {{end}}
            
            {{if .Degradations}}
            <div class="col-md-6">
                <div class="card">
                    <div class="card-header bg-danger text-white">
                        <h5 class="mb-0">Degradations</h5>
                    </div>
                    <div class="card-body">
                        {{range .Degradations}}
                        <div class="mb-3">
                            <h6>{{.Metric}}</h6>
                            <p class="mb-1">{{.Description}}</p>
                            <small class="text-muted">
                                {{formatFloat .StartValue}} → {{formatFloat .EndValue}} 
                                ({{formatFloat .DegradationRate}}% degradation)
                            </small>
                        </div>
                        {{end}}
                    </div>
                </div>
            </div>
            {{end}}
        </div>

        <!-- Predictions -->
        <div class="row mb-4">
            <div class="col-12">
                <h3>Predictions</h3>
                <div class="row">
                    <div class="col-md-4">
                        <div class="prediction-card">
                            <h5>Quality Score</h5>
                            <p class="mb-1">Next Value: <strong>{{formatFloat .Predictions.QualityPrediction.NextValue}}</strong></p>
                            <p class="mb-1">Trend: <span class="text-capitalize">{{.Predictions.QualityPrediction.TrendDirection}}</span></p>
                            <p class="mb-0">Confidence: {{formatFloat .Predictions.QualityPrediction.Confidence}}%</p>
                        </div>
                    </div>
                    <div class="col-md-4">
                        <div class="prediction-card">
                            <h5>Complexity</h5>
                            <p class="mb-1">Next Value: <strong>{{formatFloat .Predictions.ComplexityPrediction.NextValue}}</strong></p>
                            <p class="mb-1">Trend: <span class="text-capitalize">{{.Predictions.ComplexityPrediction.TrendDirection}}</span></p>
                            <p class="mb-0">Confidence: {{formatFloat .Predictions.ComplexityPrediction.Confidence}}%</p>
                        </div>
                    </div>
                    <div class="col-md-4">
                        <div class="prediction-card">
                            <h5>Technical Debt</h5>
                            <p class="mb-1">Next Value: <strong>{{formatFloat .Predictions.DebtPrediction.NextValue}}</strong></p>
                            <p class="mb-1">Trend: <span class="text-capitalize">{{.Predictions.DebtPrediction.TrendDirection}}</span></p>
                            <p class="mb-0">Confidence: {{formatFloat .Predictions.DebtPrediction.Confidence}}%</p>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <!-- Alerts -->
        {{if .Alerts}}
        <div class="row mb-4">
            <div class="col-12">
                <h3>Active Alerts</h3>
                <div class="table-responsive">
                    <table class="table table-striped">
                        <thead>
                            <tr>
                                <th>Severity</th>
                                <th>Type</th>
                                <th>Message</th>
                                <th>Time</th>
                            </tr>
                        </thead>
                        <tbody>
                            {{range .Alerts}}
                            <tr>
                                <td><span class="badge bg-{{severityColor .Severity}}">{{.Severity}}</span></td>
                                <td>{{.Type}}</td>
                                <td>{{.Message}}</td>
                                <td>{{formatTime .Timestamp}}</td>
                            </tr>
                            {{end}}
                        </tbody>
                    </table>
                </div>
            </div>
        </div>
        {{end}}

        <!-- Recommendations -->
        {{if .Recommendations}}
        <div class="row mb-4">
            <div class="col-12">
                <h3>Recommendations</h3>
                {{range .Recommendations}}
                <div class="card mb-3">
                    <div class="card-header">
                        <span class="badge bg-{{severityColor .Priority}} me-2">{{.Priority}}</span>
                        <strong>{{.Title}}</strong>
                        <span class="badge bg-secondary float-end">{{.Category}}</span>
                    </div>
                    <div class="card-body">
                        <p>{{.Description}}</p>
                        <h6>Recommended Actions:</h6>
                        <ul>
                            {{range .Actions}}
                            <li>{{.}}</li>
                            {{end}}
                        </ul>
                    </div>
                </div>
                {{end}}
            </div>
        </div>
        {{end}}
    </div>

    <script>
        // Quality Trend Chart
        const qualityCtx = document.getElementById('qualityTrendChart').getContext('2d');
        const trendData = {{json .QualityTrends}};
        
        new Chart(qualityCtx, {
            type: 'line',
            data: {
                labels: trendData.timestamps.map(t => new Date(t).toLocaleDateString()),
                datasets: [{
                    label: 'Quality Score',
                    data: trendData.quality_scores,
                    borderColor: 'rgb(75, 192, 192)',
                    backgroundColor: 'rgba(75, 192, 192, 0.2)',
                    tension: 0.1
                }, {
                    label: 'Test Coverage',
                    data: trendData.coverage_values,
                    borderColor: 'rgb(54, 162, 235)',
                    backgroundColor: 'rgba(54, 162, 235, 0.2)',
                    tension: 0.1
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    y: {
                        beginAtZero: true,
                        max: 100
                    }
                }
            }
        });

        // Complexity & Debt Chart
        const complexityCtx = document.getElementById('complexityDebtChart').getContext('2d');
        new Chart(complexityCtx, {
            type: 'line',
            data: {
                labels: trendData.timestamps.map(t => new Date(t).toLocaleDateString()),
                datasets: [{
                    label: 'Complexity',
                    data: trendData.complexity_values,
                    borderColor: 'rgb(255, 159, 64)',
                    backgroundColor: 'rgba(255, 159, 64, 0.2)',
                    tension: 0.1,
                    yAxisID: 'y'
                }, {
                    label: 'Technical Debt Ratio',
                    data: trendData.debt_ratios,
                    borderColor: 'rgb(255, 99, 132)',
                    backgroundColor: 'rgba(255, 99, 132, 0.2)',
                    tension: 0.1,
                    yAxisID: 'y1'
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    y: {
                        type: 'linear',
                        display: true,
                        position: 'left',
                        title: {
                            display: true,
                            text: 'Complexity'
                        }
                    },
                    y1: {
                        type: 'linear',
                        display: true,
                        position: 'right',
                        title: {
                            display: true,
                            text: 'Debt Ratio'
                        },
                        grid: {
                            drawOnChartArea: false
                        }
                    }
                }
            }
        });
    </script>
</body>
</html>`
}
