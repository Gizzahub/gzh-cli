package audit

import (
	"fmt"
	"math"
	"sort"
	"time"
)

// TrendReport contains the trend analysis results
type TrendReport struct {
	Organization     string                 `json:"organization"`
	Period           time.Duration          `json:"period"`
	StartDate        time.Time              `json:"start_date"`
	EndDate          time.Time              `json:"end_date"`
	OverallTrend     TrendDirection         `json:"overall_trend"`
	ComplianceChange float64                `json:"compliance_change"`
	PolicyTrends     map[string]PolicyTrend `json:"policy_trends"`
	Anomalies        []Anomaly              `json:"anomalies"`
	Predictions      []Prediction           `json:"predictions"`
	DailyCompliance  []DailyCompliancePoint `json:"daily_compliance"`
}

// TrendDirection indicates the direction of a trend
type TrendDirection string

const (
	TrendImproving TrendDirection = "improving"
	TrendDeclining TrendDirection = "declining"
	TrendStable    TrendDirection = "stable"
)

// PolicyTrend represents trend information for a specific policy
type PolicyTrend struct {
	PolicyName        string         `json:"policy_name"`
	TrendDirection    TrendDirection `json:"trend_direction"`
	ChangeRate        float64        `json:"change_rate"` // Percentage change per day
	CurrentViolations int            `json:"current_violations"`
	AverageViolations float64        `json:"average_violations"`
	PeakViolations    int            `json:"peak_violations"`
}

// Anomaly represents an unusual pattern in the data
type Anomaly struct {
	Date        time.Time `json:"date"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Severity    string    `json:"severity"`
	Value       float64   `json:"value"`
}

// Prediction represents a future trend prediction
type Prediction struct {
	Date                 time.Time `json:"date"`
	CompliancePercentage float64   `json:"compliance_percentage"`
	Confidence           float64   `json:"confidence"`
}

// DailyCompliancePoint represents compliance data for a specific day
type DailyCompliancePoint struct {
	Date                 time.Time `json:"date"`
	CompliancePercentage float64   `json:"compliance_percentage"`
	TotalRepositories    int       `json:"total_repositories"`
	CompliantRepos       int       `json:"compliant_repos"`
	ViolationCount       int       `json:"violation_count"`
}

// TrendAnalyzer analyzes audit trends over time
type TrendAnalyzer struct {
	store AuditStore
}

// NewTrendAnalyzer creates a new trend analyzer
func NewTrendAnalyzer(store AuditStore) *TrendAnalyzer {
	return &TrendAnalyzer{
		store: store,
	}
}

// AnalyzeTrends performs comprehensive trend analysis for an organization
func (ta *TrendAnalyzer) AnalyzeTrends(org string, duration time.Duration) (*TrendReport, error) {
	// Get historical data
	history, err := ta.store.GetHistoricalData(org, duration)
	if err != nil {
		return nil, fmt.Errorf("failed to get historical data: %w", err)
	}

	if len(history) == 0 {
		return nil, fmt.Errorf("no historical data available for organization %s", org)
	}

	// Sort by timestamp
	sort.Slice(history, func(i, j int) bool {
		return history[i].Timestamp.Before(history[j].Timestamp)
	})

	report := &TrendReport{
		Organization: org,
		Period:       duration,
		StartDate:    history[0].Timestamp,
		EndDate:      history[len(history)-1].Timestamp,
		PolicyTrends: make(map[string]PolicyTrend),
	}

	// Calculate overall trend
	report.OverallTrend, report.ComplianceChange = ta.calculateOverallTrend(history)

	// Analyze policy trends
	report.PolicyTrends = ta.analyzePolicyTrends(history)

	// Detect anomalies
	report.Anomalies = ta.detectAnomalies(history)

	// Generate predictions
	report.Predictions = ta.generatePredictions(history)

	// Generate daily compliance points
	report.DailyCompliance = ta.generateDailyCompliance(history)

	return report, nil
}

// calculateOverallTrend determines the overall compliance trend
func (ta *TrendAnalyzer) calculateOverallTrend(history []AuditHistory) (TrendDirection, float64) {
	if len(history) < 2 {
		return TrendStable, 0
	}

	// Calculate linear regression slope
	var sumX, sumY, sumXY, sumX2 float64
	n := float64(len(history))

	for i, h := range history {
		x := float64(i)
		y := h.Summary.CompliancePercentage
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	// Calculate slope (change per data point)
	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)

	// Calculate total change
	firstCompliance := history[0].Summary.CompliancePercentage
	lastCompliance := history[len(history)-1].Summary.CompliancePercentage
	totalChange := lastCompliance - firstCompliance

	// Determine trend direction
	var trend TrendDirection
	if math.Abs(slope) < 0.1 { // Less than 0.1% change per data point
		trend = TrendStable
	} else if slope > 0 {
		trend = TrendImproving
	} else {
		trend = TrendDeclining
	}

	return trend, totalChange
}

// analyzePolicyTrends analyzes trends for each policy
func (ta *TrendAnalyzer) analyzePolicyTrends(history []AuditHistory) map[string]PolicyTrend {
	policyData := make(map[string][]PolicyStatistics)

	// Group by policy
	for _, h := range history {
		for policyName, stats := range h.PolicyStats {
			policyData[policyName] = append(policyData[policyName], stats)
		}
	}

	trends := make(map[string]PolicyTrend)
	for policyName, stats := range policyData {
		if len(stats) == 0 {
			continue
		}

		// Calculate trend for this policy
		var totalViolations float64
		peakViolations := 0

		for _, s := range stats {
			totalViolations += float64(s.ViolationCount)
			if s.ViolationCount > peakViolations {
				peakViolations = s.ViolationCount
			}
		}

		avgViolations := totalViolations / float64(len(stats))
		currentViolations := stats[len(stats)-1].ViolationCount

		// Simple trend calculation based on first and last data points
		var changeRate float64
		var direction TrendDirection

		if len(stats) >= 2 {
			firstViolations := float64(stats[0].ViolationCount)
			lastViolations := float64(stats[len(stats)-1].ViolationCount)

			if firstViolations > 0 {
				changeRate = ((lastViolations - firstViolations) / firstViolations) * 100
			}

			if math.Abs(changeRate) < 5 {
				direction = TrendStable
			} else if changeRate < 0 {
				direction = TrendImproving // Fewer violations is improving
			} else {
				direction = TrendDeclining
			}
		} else {
			direction = TrendStable
		}

		trends[policyName] = PolicyTrend{
			PolicyName:        policyName,
			TrendDirection:    direction,
			ChangeRate:        changeRate,
			CurrentViolations: currentViolations,
			AverageViolations: avgViolations,
			PeakViolations:    peakViolations,
		}
	}

	return trends
}

// detectAnomalies identifies unusual patterns in the data
func (ta *TrendAnalyzer) detectAnomalies(history []AuditHistory) []Anomaly {
	var anomalies []Anomaly

	if len(history) < 3 {
		return anomalies
	}

	// Calculate moving average and standard deviation
	windowSize := 7 // 7-day moving average
	if windowSize > len(history) {
		windowSize = len(history)
	}

	for i := windowSize - 1; i < len(history); i++ {
		// Calculate window statistics
		var sum, sumSq float64
		for j := i - windowSize + 1; j <= i; j++ {
			compliance := history[j].Summary.CompliancePercentage
			sum += compliance
			sumSq += compliance * compliance
		}

		mean := sum / float64(windowSize)
		variance := (sumSq / float64(windowSize)) - (mean * mean)
		stdDev := math.Sqrt(variance)

		currentCompliance := history[i].Summary.CompliancePercentage

		// Check for anomalies (2 standard deviations from mean)
		if math.Abs(currentCompliance-mean) > 2*stdDev && stdDev > 0 {
			severity := "medium"
			if math.Abs(currentCompliance-mean) > 3*stdDev {
				severity = "high"
			}

			anomalyType := "sudden_drop"
			if currentCompliance > mean {
				anomalyType = "sudden_improvement"
			}

			anomalies = append(anomalies, Anomaly{
				Date:        history[i].Timestamp,
				Type:        anomalyType,
				Description: fmt.Sprintf("Compliance %.1f%% deviates significantly from average %.1f%%", currentCompliance, mean),
				Severity:    severity,
				Value:       currentCompliance,
			})
		}

		// Check for sudden violation spikes
		if i > 0 {
			prevViolations := history[i-1].Summary.TotalViolations
			currViolations := history[i].Summary.TotalViolations

			if prevViolations > 0 {
				changePercent := float64(currViolations-prevViolations) / float64(prevViolations) * 100
				if changePercent > 50 { // 50% increase in violations
					anomalies = append(anomalies, Anomaly{
						Date:        history[i].Timestamp,
						Type:        "violation_spike",
						Description: fmt.Sprintf("Violations increased by %.0f%% from %d to %d", changePercent, prevViolations, currViolations),
						Severity:    "high",
						Value:       float64(currViolations),
					})
				}
			}
		}
	}

	return anomalies
}

// generatePredictions creates future compliance predictions
func (ta *TrendAnalyzer) generatePredictions(history []AuditHistory) []Prediction {
	var predictions []Prediction

	if len(history) < 7 { // Need at least a week of data
		return predictions
	}

	// Simple linear regression for prediction
	var sumX, sumY, sumXY, sumX2 float64
	n := float64(len(history))

	for i, h := range history {
		x := float64(i)
		y := h.Summary.CompliancePercentage
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	// Calculate regression coefficients
	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
	intercept := (sumY - slope*sumX) / n

	// Generate predictions for next 7 days
	for i := 1; i <= 7; i++ {
		predictedDate := history[len(history)-1].Timestamp.AddDate(0, 0, i)
		x := float64(len(history) - 1 + i)
		predictedCompliance := slope*x + intercept

		// Ensure compliance is within valid range
		if predictedCompliance < 0 {
			predictedCompliance = 0
		} else if predictedCompliance > 100 {
			predictedCompliance = 100
		}

		// Calculate confidence based on data consistency
		confidence := ta.calculatePredictionConfidence(history, slope)

		predictions = append(predictions, Prediction{
			Date:                 predictedDate,
			CompliancePercentage: predictedCompliance,
			Confidence:           confidence,
		})
	}

	return predictions
}

// calculatePredictionConfidence calculates confidence in predictions
func (ta *TrendAnalyzer) calculatePredictionConfidence(history []AuditHistory, slope float64) float64 {
	// Calculate R-squared for linear regression
	var sumY, ssTotal, ssResidual float64
	n := float64(len(history))

	// Calculate mean
	for _, h := range history {
		sumY += h.Summary.CompliancePercentage
	}
	mean := sumY / n

	// Calculate sum of squares
	for i, h := range history {
		y := h.Summary.CompliancePercentage
		yPred := slope*float64(i) + mean

		ssTotal += (y - mean) * (y - mean)
		ssResidual += (y - yPred) * (y - yPred)
	}

	// R-squared value (0 to 1)
	rSquared := 1 - (ssResidual / ssTotal)
	if rSquared < 0 {
		rSquared = 0
	}

	// Convert to percentage confidence
	confidence := rSquared * 100

	// Reduce confidence for volatile data
	if len(history) > 2 {
		var volatility float64
		for i := 1; i < len(history); i++ {
			change := math.Abs(history[i].Summary.CompliancePercentage - history[i-1].Summary.CompliancePercentage)
			volatility += change
		}
		avgVolatility := volatility / float64(len(history)-1)

		// High volatility reduces confidence
		if avgVolatility > 10 {
			confidence *= 0.7
		} else if avgVolatility > 5 {
			confidence *= 0.85
		}
	}

	return confidence
}

// generateDailyCompliance creates daily compliance data points
func (ta *TrendAnalyzer) generateDailyCompliance(history []AuditHistory) []DailyCompliancePoint {
	dailyMap := make(map[string][]AuditHistory)

	// Group by date
	for _, h := range history {
		dateKey := h.Timestamp.Format("2006-01-02")
		dailyMap[dateKey] = append(dailyMap[dateKey], h)
	}

	var points []DailyCompliancePoint
	for date, records := range dailyMap {
		// Use the last record of each day
		lastRecord := records[len(records)-1]

		parsedDate, _ := time.Parse("2006-01-02", date)
		points = append(points, DailyCompliancePoint{
			Date:                 parsedDate,
			CompliancePercentage: lastRecord.Summary.CompliancePercentage,
			TotalRepositories:    lastRecord.Summary.TotalRepositories,
			CompliantRepos:       lastRecord.Summary.CompliantRepositories,
			ViolationCount:       lastRecord.Summary.TotalViolations,
		})
	}

	// Sort by date
	sort.Slice(points, func(i, j int) bool {
		return points[i].Date.Before(points[j].Date)
	})

	return points
}
