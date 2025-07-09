package audit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuditStore is a mock implementation of AuditStore for testing
type MockAuditStore struct {
	mock.Mock
}

func (m *MockAuditStore) SaveAuditResult(history *AuditHistory) error {
	args := m.Called(history)
	return args.Error(0)
}

func (m *MockAuditStore) GetHistoricalData(org string, duration time.Duration) ([]AuditHistory, error) {
	args := m.Called(org, duration)
	return args.Get(0).([]AuditHistory), args.Error(1)
}

func (m *MockAuditStore) GetPolicyTrends(org, policy string, duration time.Duration) ([]PolicyStatistics, error) {
	args := m.Called(org, policy, duration)
	return args.Get(0).([]PolicyStatistics), args.Error(1)
}

func TestTrendAnalyzer(t *testing.T) {
	t.Run("AnalyzeTrends_NoData", func(t *testing.T) {
		store := &MockAuditStore{}
		store.On("GetHistoricalData", "test-org", time.Duration(30*24*time.Hour)).Return([]AuditHistory{}, nil)

		analyzer := NewTrendAnalyzer(store)

		_, err := analyzer.AnalyzeTrends("test-org", 30*24*time.Hour)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no historical data available")
	})

	t.Run("AnalyzeTrends_WithData", func(t *testing.T) {
		store := &MockAuditStore{}

		// Create mock historical data
		history := []AuditHistory{
			{
				Timestamp:    time.Now().Add(-3 * 24 * time.Hour),
				Organization: "test-org",
				Summary: AuditSummary{
					TotalRepositories:     10,
					CompliantRepositories: 6,
					CompliancePercentage:  60.0,
					TotalViolations:       8,
					CriticalViolations:    2,
				},
				PolicyStats: map[string]PolicyStatistics{
					"security": {
						PolicyName:           "security",
						ViolationCount:       5,
						CompliantRepos:       5,
						ViolatingRepos:       5,
						CompliancePercentage: 50.0,
					},
				},
			},
			{
				Timestamp:    time.Now().Add(-2 * 24 * time.Hour),
				Organization: "test-org",
				Summary: AuditSummary{
					TotalRepositories:     10,
					CompliantRepositories: 7,
					CompliancePercentage:  70.0,
					TotalViolations:       6,
					CriticalViolations:    1,
				},
				PolicyStats: map[string]PolicyStatistics{
					"security": {
						PolicyName:           "security",
						ViolationCount:       4,
						CompliantRepos:       6,
						ViolatingRepos:       4,
						CompliancePercentage: 60.0,
					},
				},
			},
			{
				Timestamp:    time.Now().Add(-1 * 24 * time.Hour),
				Organization: "test-org",
				Summary: AuditSummary{
					TotalRepositories:     10,
					CompliantRepositories: 8,
					CompliancePercentage:  80.0,
					TotalViolations:       4,
					CriticalViolations:    1,
				},
				PolicyStats: map[string]PolicyStatistics{
					"security": {
						PolicyName:           "security",
						ViolationCount:       3,
						CompliantRepos:       7,
						ViolatingRepos:       3,
						CompliancePercentage: 70.0,
					},
				},
			},
		}

		store.On("GetHistoricalData", "test-org", time.Duration(30*24*time.Hour)).Return(history, nil)

		analyzer := NewTrendAnalyzer(store)
		report, err := analyzer.AnalyzeTrends("test-org", 30*24*time.Hour)

		assert.NoError(t, err)
		assert.NotNil(t, report)
		assert.Equal(t, "test-org", report.Organization)
		assert.Equal(t, time.Duration(30*24*time.Hour), report.Period)
		assert.Equal(t, TrendImproving, report.OverallTrend)
		assert.Equal(t, 20.0, report.ComplianceChange) // 80% - 60% = 20%

		// Check policy trends
		assert.Contains(t, report.PolicyTrends, "security")
		securityTrend := report.PolicyTrends["security"]
		assert.Equal(t, TrendImproving, securityTrend.TrendDirection)
		assert.Equal(t, 3, securityTrend.CurrentViolations)

		// Check daily compliance
		assert.Len(t, report.DailyCompliance, 3)
		assert.Equal(t, 80.0, report.DailyCompliance[2].CompliancePercentage) // Latest data

		// Check that predictions are generated (need at least 7 days of data)
		// With only 3 days of data, no predictions should be generated
		assert.Empty(t, report.Predictions)
	})

	t.Run("CalculateOverallTrend", func(t *testing.T) {
		store := &MockAuditStore{}
		analyzer := NewTrendAnalyzer(store)

		// Test improving trend
		improvingHistory := []AuditHistory{
			{Summary: AuditSummary{CompliancePercentage: 60.0}},
			{Summary: AuditSummary{CompliancePercentage: 70.0}},
			{Summary: AuditSummary{CompliancePercentage: 80.0}},
		}
		trend, change := analyzer.calculateOverallTrend(improvingHistory)
		assert.Equal(t, TrendImproving, trend)
		assert.Equal(t, 20.0, change)

		// Test declining trend
		decliningHistory := []AuditHistory{
			{Summary: AuditSummary{CompliancePercentage: 80.0}},
			{Summary: AuditSummary{CompliancePercentage: 70.0}},
			{Summary: AuditSummary{CompliancePercentage: 60.0}},
		}
		trend, change = analyzer.calculateOverallTrend(decliningHistory)
		assert.Equal(t, TrendDeclining, trend)
		assert.Equal(t, -20.0, change)

		// Test stable trend
		stableHistory := []AuditHistory{
			{Summary: AuditSummary{CompliancePercentage: 70.0}},
			{Summary: AuditSummary{CompliancePercentage: 70.1}},
			{Summary: AuditSummary{CompliancePercentage: 69.9}},
		}
		trend, change = analyzer.calculateOverallTrend(stableHistory)
		assert.Equal(t, TrendStable, trend)
		assert.InDelta(t, -0.1, change, 0.01) // Use delta for floating point comparison
	})

	t.Run("DetectAnomalies", func(t *testing.T) {
		store := &MockAuditStore{}
		analyzer := NewTrendAnalyzer(store)

		// Create history with anomaly
		history := []AuditHistory{
			{Timestamp: time.Now().Add(-10 * time.Hour), Summary: AuditSummary{CompliancePercentage: 70.0, TotalViolations: 5}},
			{Timestamp: time.Now().Add(-9 * time.Hour), Summary: AuditSummary{CompliancePercentage: 72.0, TotalViolations: 5}},
			{Timestamp: time.Now().Add(-8 * time.Hour), Summary: AuditSummary{CompliancePercentage: 71.0, TotalViolations: 5}},
			{Timestamp: time.Now().Add(-7 * time.Hour), Summary: AuditSummary{CompliancePercentage: 69.0, TotalViolations: 5}},
			{Timestamp: time.Now().Add(-6 * time.Hour), Summary: AuditSummary{CompliancePercentage: 70.0, TotalViolations: 5}},
			{Timestamp: time.Now().Add(-5 * time.Hour), Summary: AuditSummary{CompliancePercentage: 71.0, TotalViolations: 5}},
			{Timestamp: time.Now().Add(-4 * time.Hour), Summary: AuditSummary{CompliancePercentage: 72.0, TotalViolations: 5}},
			{Timestamp: time.Now().Add(-3 * time.Hour), Summary: AuditSummary{CompliancePercentage: 35.0, TotalViolations: 5}}, // Anomaly
			{Timestamp: time.Now().Add(-2 * time.Hour), Summary: AuditSummary{CompliancePercentage: 70.0, TotalViolations: 5}},
			{Timestamp: time.Now().Add(-1 * time.Hour), Summary: AuditSummary{CompliancePercentage: 71.0, TotalViolations: 5}},
		}

		anomalies := analyzer.detectAnomalies(history)
		assert.NotEmpty(t, anomalies)

		// Check that anomaly was detected
		found := false
		for _, anomaly := range anomalies {
			if anomaly.Type == "sudden_drop" {
				found = true
				assert.Equal(t, 35.0, anomaly.Value)
				break
			}
		}
		assert.True(t, found, "Expected to find sudden_drop anomaly")
	})

	t.Run("GeneratePredictions", func(t *testing.T) {
		store := &MockAuditStore{}
		analyzer := NewTrendAnalyzer(store)

		// Create history with clear trend
		history := []AuditHistory{}
		for i := 0; i < 10; i++ {
			history = append(history, AuditHistory{
				Timestamp: time.Now().Add(time.Duration(-i) * 24 * time.Hour),
				Summary: AuditSummary{
					CompliancePercentage: float64(60 + i*2), // Improving trend
				},
			})
		}

		predictions := analyzer.generatePredictions(history)
		assert.Len(t, predictions, 7) // Should generate 7 days of predictions

		// Check that predictions are reasonable
		for _, pred := range predictions {
			assert.True(t, pred.CompliancePercentage >= 0 && pred.CompliancePercentage <= 100)
			assert.True(t, pred.Confidence >= 0 && pred.Confidence <= 100)
		}
	})

	t.Run("AnalyzePolicyTrends", func(t *testing.T) {
		store := &MockAuditStore{}
		analyzer := NewTrendAnalyzer(store)

		history := []AuditHistory{
			{
				PolicyStats: map[string]PolicyStatistics{
					"security": {
						PolicyName:           "security",
						ViolationCount:       10,
						CompliantRepos:       5,
						ViolatingRepos:       5,
						CompliancePercentage: 50.0,
					},
				},
			},
			{
				PolicyStats: map[string]PolicyStatistics{
					"security": {
						PolicyName:           "security",
						ViolationCount:       8,
						CompliantRepos:       6,
						ViolatingRepos:       4,
						CompliancePercentage: 60.0,
					},
				},
			},
			{
				PolicyStats: map[string]PolicyStatistics{
					"security": {
						PolicyName:           "security",
						ViolationCount:       6,
						CompliantRepos:       7,
						ViolatingRepos:       3,
						CompliancePercentage: 70.0,
					},
				},
			},
		}

		trends := analyzer.analyzePolicyTrends(history)
		assert.Contains(t, trends, "security")

		securityTrend := trends["security"]
		assert.Equal(t, "security", securityTrend.PolicyName)
		assert.Equal(t, TrendImproving, securityTrend.TrendDirection) // Violations decreasing
		assert.Equal(t, 6, securityTrend.CurrentViolations)
		assert.Equal(t, 10, securityTrend.PeakViolations)
		assert.Equal(t, 8.0, securityTrend.AverageViolations)
	})
}

func TestTrendDirection(t *testing.T) {
	tests := []struct {
		name     string
		trend    TrendDirection
		expected string
	}{
		{"Improving", TrendImproving, "improving"},
		{"Declining", TrendDeclining, "declining"},
		{"Stable", TrendStable, "stable"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.trend))
		})
	}
}
