package reposync

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestCalculateTechnicalDebt(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	config := &QualityCheckConfig{}
	analyzer, err := NewCodeQualityAnalyzer(logger, config)
	require.NoError(t, err)

	t.Run("EmptyResult", func(t *testing.T) {
		result := &QualityResult{
			Issues: []QualityIssue{},
			Metrics: QualityMetrics{
				TotalLinesOfCode: 1000,
			},
		}

		debt := analyzer.calculateTechnicalDebt(result)
		assert.Equal(t, 0, debt.TotalMinutes)
		assert.Equal(t, 0.0, debt.DebtRatio)
		assert.Greater(t, debt.MaintenanceIndex, 0.0)
	})

	t.Run("WithIssues", func(t *testing.T) {
		result := &QualityResult{
			Issues: []QualityIssue{
				{Severity: "critical"}, // 60 min
				{Severity: "major"},    // 30 min
				{Severity: "major"},    // 30 min
				{Severity: "minor"},    // 15 min
				{Severity: "minor"},    // 15 min
				{Severity: "info"},     // 5 min
			},
			Metrics: QualityMetrics{
				TotalLinesOfCode: 1000,
				AvgComplexity:    10.0,
			},
		}

		debt := analyzer.calculateTechnicalDebt(result)
		assert.Equal(t, 155, debt.TotalMinutes) // 60 + 30 + 30 + 15 + 15 + 5
		assert.Equal(t, 155.0, debt.DebtRatio)  // 155 / 1000 * 1000
		assert.Greater(t, debt.MaintenanceIndex, 0.0)
		assert.Less(t, debt.MaintenanceIndex, 100.0)
	})
}

func TestCalculateSecurityScore(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	config := &QualityCheckConfig{}
	analyzer, err := NewCodeQualityAnalyzer(logger, config)
	require.NoError(t, err)

	t.Run("NoSecurityIssues", func(t *testing.T) {
		result := &QualityResult{
			Issues: []QualityIssue{
				{Type: "style", Severity: "minor"},
				{Type: "bug", Severity: "major"},
			},
			SecurityIssues: []SecurityIssue{},
		}

		score := analyzer.calculateSecurityScore(result)
		assert.Equal(t, 100.0, score)
	})

	t.Run("WithSecurityIssues", func(t *testing.T) {
		result := &QualityResult{
			Issues: []QualityIssue{
				{Type: "security", Severity: "critical"}, // -20
				{Type: "security", Severity: "major"},    // -10
				{Type: "security", Severity: "minor"},    // -5
			},
			SecurityIssues: []SecurityIssue{
				{CVSS: 8.5}, // -15
				{CVSS: 5.5}, // -10
				{CVSS: 2.0}, // -0
			},
		}

		score := analyzer.calculateSecurityScore(result)
		assert.Equal(t, 40.0, score) // 100 - 20 - 10 - 5 - 15 - 10
	})
}

func TestCalculateMaintainability(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	config := &QualityCheckConfig{}
	analyzer, err := NewCodeQualityAnalyzer(logger, config)
	require.NoError(t, err)

	tests := []struct {
		name        string
		quality     float64
		complexity  float64
		coverage    float64
		minExpected float64
		maxExpected float64
	}{
		{
			name:        "HighQuality",
			quality:     90.0,
			complexity:  5.0,
			coverage:    85.0,
			minExpected: 80.0,
			maxExpected: 100.0,
		},
		{
			name:        "LowQuality",
			quality:     50.0,
			complexity:  20.0,
			coverage:    30.0,
			minExpected: 0.0,
			maxExpected: 50.0,
		},
		{
			name:        "MediumQuality",
			quality:     70.0,
			complexity:  10.0,
			coverage:    60.0,
			minExpected: 50.0,
			maxExpected: 80.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			maintainability := analyzer.calculateMaintainability(tt.quality, tt.complexity, tt.coverage)
			assert.GreaterOrEqual(t, maintainability, tt.minExpected)
			assert.LessOrEqual(t, maintainability, tt.maxExpected)
		})
	}
}

func TestGenerateRecommendations(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	config := &QualityCheckConfig{}
	analyzer, err := NewCodeQualityAnalyzer(logger, config)
	require.NoError(t, err)

	t.Run("CriticalQuality", func(t *testing.T) {
		result := &QualityResult{
			OverallScore:    45.0,
			LanguageResults: map[string]*LanguageQuality{},
		}

		analyzer.generateRecommendations(result)
		assert.Contains(t, result.Recommendations[0], "Critical")
	})

	t.Run("HighComplexity", func(t *testing.T) {
		result := &QualityResult{
			OverallScore: 85.0,
			LanguageResults: map[string]*LanguageQuality{
				"go": {
					ComplexityScore: 15.0,
					DuplicationRate: 2.0,
					TestCoverage:    90.0,
				},
			},
		}

		analyzer.generateRecommendations(result)
		found := false
		for _, rec := range result.Recommendations {
			if contains(rec, "complexity") {
				found = true
				break
			}
		}
		assert.True(t, found, "Should have complexity recommendation")
	})

	t.Run("LowTestCoverage", func(t *testing.T) {
		result := &QualityResult{
			OverallScore: 85.0,
			LanguageResults: map[string]*LanguageQuality{
				"python": {
					ComplexityScore: 5.0,
					DuplicationRate: 2.0,
					TestCoverage:    60.0,
				},
			},
		}

		analyzer.generateRecommendations(result)
		found := false
		for _, rec := range result.Recommendations {
			if contains(rec, "test coverage") {
				found = true
				break
			}
		}
		assert.True(t, found, "Should have test coverage recommendation")
	})
}

func TestDetectLanguages(t *testing.T) {
	// This test would require a test directory structure
	// Skipping for now as it requires file system setup
}

func TestLanguageAnalyzers(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	t.Run("GoAnalyzer", func(t *testing.T) {
		analyzer := NewGoQualityAnalyzer(logger)
		assert.Equal(t, "golangci-lint", analyzer.Name())
		assert.Equal(t, "go", analyzer.Language())

		// Check if golangci-lint is available
		ctx := context.Background()
		if analyzer.IsAvailable(ctx) {
			t.Log("golangci-lint is available for testing")
		} else {
			t.Skip("golangci-lint not available")
		}
	})

	t.Run("PythonAnalyzer", func(t *testing.T) {
		analyzer := NewPythonQualityAnalyzer(logger)
		assert.Equal(t, "pylint", analyzer.Name())
		assert.Equal(t, "python", analyzer.Language())

		// Check if pylint is available
		ctx := context.Background()
		if analyzer.IsAvailable(ctx) {
			t.Log("pylint is available for testing")
		} else {
			t.Skip("pylint not available")
		}
	})
}

// Helper function
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || len(s) > len(substr) && containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 1; i < len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
