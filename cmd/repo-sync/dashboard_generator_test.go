package reposync

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestDashboardGenerator(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tempDir := t.TempDir()

	t.Run("GenerateDashboard", func(t *testing.T) {
		generator := NewDashboardGenerator(logger, tempDir)

		// Create sample quality result
		result := &QualityResult{
			Repository:   "test-repo",
			Timestamp:    time.Now(),
			OverallScore: 85.5,
			LanguageResults: map[string]*LanguageQuality{
				"go": {
					Language:        "go",
					FilesAnalyzed:   50,
					LinesOfCode:     5000,
					ComplexityScore: 6.2,
					DuplicationRate: 3.5,
					TestCoverage:    82.3,
					QualityScore:    88.0,
					Issues: []QualityIssue{
						{
							Type:     "style",
							Severity: "minor",
							File:     "main.go",
							Line:     42,
							Message:  "Function has too many statements",
							Rule:     "funlen",
							Tool:     "golangci-lint",
						},
					},
				},
				"python": {
					Language:        "python",
					FilesAnalyzed:   30,
					LinesOfCode:     3000,
					ComplexityScore: 8.5,
					DuplicationRate: 5.2,
					TestCoverage:    75.0,
					QualityScore:    82.0,
					Issues: []QualityIssue{
						{
							Type:     "security",
							Severity: "major",
							File:     "app.py",
							Line:     156,
							Message:  "Possible SQL injection",
							Rule:     "B608",
							Tool:     "bandit",
						},
					},
				},
			},
			Metrics: QualityMetrics{
				TotalFiles:         80,
				TotalLinesOfCode:   8000,
				AvgComplexity:      7.1,
				DuplicationRate:    4.2,
				TestCoverage:       79.5,
				TechnicalDebtRatio: 15.3,
				SecurityScore:      90.0,
				Maintainability:    85.0,
			},
			Issues: []QualityIssue{
				{
					Type:     "critical",
					Severity: "critical",
					File:     "security.go",
					Line:     234,
					Message:  "Hard-coded credentials",
					Rule:     "G101",
					Tool:     "gosec",
				},
			},
			Recommendations: []string{
				"Improve test coverage for Python code (current: 75.0%)",
				"Address security vulnerabilities in security.go",
				"Reduce code duplication rate below 5%",
			},
			TechnicalDebt: TechnicalDebtInfo{
				TotalMinutes:     450,
				DebtRatio:        15.3,
				MaintenanceIndex: 85.0,
			},
		}

		// Generate dashboard
		err := generator.GenerateDashboard(result, []*QualityResult{})
		require.NoError(t, err)

		// Check if dashboard file was created
		dashboardPath := filepath.Join(tempDir, "quality-dashboard.html")
		assert.FileExists(t, dashboardPath)

		// Check if assets were created
		cssPath := filepath.Join(tempDir, "dashboard.css")
		assert.FileExists(t, cssPath)

		jsPath := filepath.Join(tempDir, "dashboard.js")
		assert.FileExists(t, jsPath)

		// Read and verify dashboard content
		content, err := os.ReadFile(dashboardPath)
		require.NoError(t, err)

		// Check for key elements
		assert.Contains(t, string(content), "Code Quality Dashboard")
		assert.Contains(t, string(content), "test-repo")
		assert.Contains(t, string(content), "85.5%") // Overall score
		assert.Contains(t, string(content), "82.3%") // Go test coverage
		assert.Contains(t, string(content), "Hard-coded credentials")
		assert.Contains(t, string(content), "Improve test coverage for Python code")
	})

	t.Run("PrepareChartsData", func(t *testing.T) {
		generator := NewDashboardGenerator(logger, tempDir)

		result := &QualityResult{
			Issues: []QualityIssue{
				{Type: "bug", Severity: "critical"},
				{Type: "bug", Severity: "major"},
				{Type: "style", Severity: "minor"},
				{Type: "style", Severity: "minor"},
				{Type: "security", Severity: "critical"},
			},
			LanguageResults: map[string]*LanguageQuality{
				"go": {
					FilesAnalyzed:   50,
					ComplexityScore: 5.0,
				},
				"python": {
					FilesAnalyzed:   30,
					ComplexityScore: 15.0,
				},
			},
		}

		chartsData := generator.prepareChartsData(result)

		// Check issues by type
		assert.Equal(t, 2, chartsData.IssuesByType["bug"])
		assert.Equal(t, 2, chartsData.IssuesByType["style"])
		assert.Equal(t, 1, chartsData.IssuesByType["security"])

		// Check issues by severity
		assert.Equal(t, 2, chartsData.IssuesBySeverity["critical"])
		assert.Equal(t, 1, chartsData.IssuesBySeverity["major"])
		assert.Equal(t, 2, chartsData.IssuesBySeverity["minor"])

		// Check files by language
		assert.Equal(t, 50, chartsData.FilesByLanguage["go"])
		assert.Equal(t, 30, chartsData.FilesByLanguage["python"])

		// Check complexity distribution
		assert.Equal(t, 1, chartsData.ComplexityDist["Low (1-5)"])
		assert.Equal(t, 1, chartsData.ComplexityDist["High (10-20)"])
	})

	t.Run("ExtractTrendData", func(t *testing.T) {
		generator := NewDashboardGenerator(logger, tempDir)

		// Create historical data
		now := time.Now()
		historicalData := make([]*QualityResult, 5)
		for i := 0; i < 5; i++ {
			historicalData[i] = &QualityResult{
				Timestamp:    now.Add(time.Duration(-i) * 24 * time.Hour),
				OverallScore: 80.0 + float64(i),
				Metrics: QualityMetrics{
					AvgComplexity:      5.0 + float64(i)*0.5,
					TestCoverage:       70.0 + float64(i)*2,
					TechnicalDebtRatio: 20.0 - float64(i),
				},
				Issues: make([]QualityIssue, 10-i),
			}
		}

		trendData := generator.extractTrendData(historicalData)

		// Check trend data
		assert.Len(t, trendData.Dates, 5)
		assert.Len(t, trendData.Scores, 5)
		assert.Len(t, trendData.Complexity, 5)
		assert.Len(t, trendData.Coverage, 5)
		assert.Len(t, trendData.IssuesCount, 5)
		assert.Len(t, trendData.TechnicalDebt, 5)

		// Verify data is in correct order (oldest to newest)
		assert.Equal(t, 84.0, trendData.Scores[0])
		assert.Equal(t, 80.0, trendData.Scores[4])
	})
}

func TestScoreColorMapping(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	generator := NewDashboardGenerator(logger, "")

	tests := []struct {
		score    float64
		expected string
	}{
		{95.0, "success"},
		{90.0, "success"},
		{85.0, "info"},
		{80.0, "info"},
		{65.0, "warning"},
		{60.0, "warning"},
		{50.0, "danger"},
		{30.0, "danger"},
	}

	for _, tt := range tests {
		t.Run(string(rune(int(tt.score))), func(t *testing.T) {
			color := generator.getScoreColor(tt.score)
			assert.Equal(t, tt.expected, color)
		})
	}
}

func TestComplexityRangeMapping(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	generator := NewDashboardGenerator(logger, "")

	tests := []struct {
		complexity float64
		expected   string
	}{
		{2.0, "Low (1-5)"},
		{4.9, "Low (1-5)"},
		{5.0, "Medium (5-10)"},
		{9.9, "Medium (5-10)"},
		{10.0, "High (10-20)"},
		{19.9, "High (10-20)"},
		{20.0, "Very High (20+)"},
		{50.0, "Very High (20+)"},
	}

	for _, tt := range tests {
		t.Run(string(rune(int(tt.complexity))), func(t *testing.T) {
			rangeStr := generator.getComplexityRange(tt.complexity)
			assert.Equal(t, tt.expected, rangeStr)
		})
	}
}
