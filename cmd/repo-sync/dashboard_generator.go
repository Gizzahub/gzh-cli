package reposync

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"
)

// DashboardGenerator generates HTML dashboard for quality metrics
type DashboardGenerator struct {
	logger       *zap.Logger
	templatePath string
	outputDir    string
}

// NewDashboardGenerator creates a new dashboard generator
func NewDashboardGenerator(logger *zap.Logger, outputDir string) *DashboardGenerator {
	return &DashboardGenerator{
		logger:    logger,
		outputDir: outputDir,
	}
}

// GenerateDashboard generates an HTML dashboard from quality results
func (dg *DashboardGenerator) GenerateDashboard(result *RepoQualityResult, historicalData []*RepoQualityResult) error {
	// Ensure output directory exists
	if err := os.MkdirAll(dg.outputDir, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate dashboard data
	data := dg.prepareDashboardData(result, historicalData)

	// Generate HTML
	html, err := dg.generateHTML(data)
	if err != nil {
		return fmt.Errorf("failed to generate HTML: %w", err)
	}

	// Write HTML file
	outputPath := filepath.Join(dg.outputDir, "quality-dashboard.html")
	if err := os.WriteFile(outputPath, []byte(html), 0o644); err != nil {
		return fmt.Errorf("failed to write dashboard file: %w", err)
	}

	// Copy assets (CSS, JS)
	if err := dg.generateAssets(); err != nil {
		return fmt.Errorf("failed to generate assets: %w", err)
	}

	dg.logger.Info("Dashboard generated successfully",
		zap.String("path", outputPath))

	fmt.Printf("📊 Quality dashboard generated: %s\n", outputPath)
	return nil
}

// DashboardData represents data for the dashboard template
type DashboardData struct {
	Title           string               `json:"title"`
	GeneratedAt     string               `json:"generated_at"`
	Repository      string               `json:"repository"`
	OverallScore    float64              `json:"overall_score"`
	ScoreColor      string               `json:"score_color"`
	Metrics         QualityMetrics       `json:"metrics"`
	LanguageResults []LanguageResultData `json:"language_results"`
	IssuesSummary   map[string]int       `json:"issues_summary"`
	TopIssues       []QualityIssue       `json:"top_issues"`
	TrendData       TrendData            `json:"trend_data"`
	Recommendations []string             `json:"recommendations"`
	ChartsData      ChartsData           `json:"charts_data"`
}

// LanguageResultData represents language-specific results for dashboard
type LanguageResultData struct {
	Language        string  `json:"language"`
	Score           float64 `json:"score"`
	ScoreColor      string  `json:"score_color"`
	FilesAnalyzed   int     `json:"files_analyzed"`
	LinesOfCode     int     `json:"lines_of_code"`
	IssuesCount     int     `json:"issues_count"`
	ComplexityScore float64 `json:"complexity_score"`
	TestCoverage    float64 `json:"test_coverage"`
	DuplicationRate float64 `json:"duplication_rate"`
}

// TrendData represents historical trend data
type TrendData struct {
	Dates         []string  `json:"dates"`
	Scores        []float64 `json:"scores"`
	Complexity    []float64 `json:"complexity"`
	Coverage      []float64 `json:"coverage"`
	IssuesCount   []int     `json:"issues_count"`
	TechnicalDebt []float64 `json:"technical_debt"`
}

// ChartsData represents data for various charts
type ChartsData struct {
	IssuesByType     map[string]int `json:"issues_by_type"`
	IssuesBySeverity map[string]int `json:"issues_by_severity"`
	FilesByLanguage  map[string]int `json:"files_by_language"`
	ComplexityDist   map[string]int `json:"complexity_distribution"`
}

func (dg *DashboardGenerator) prepareDashboardData(result *RepoQualityResult, historicalData []*RepoQualityResult) *DashboardData {
	data := &DashboardData{
		Title:           "Code Quality Dashboard",
		GeneratedAt:     time.Now().Format("2006-01-02 15:04:05"),
		Repository:      result.Repository,
		OverallScore:    result.OverallScore,
		ScoreColor:      dg.getScoreColor(result.OverallScore),
		Metrics:         result.Metrics,
		IssuesSummary:   dg.summarizeIssues(result.Issues),
		TopIssues:       dg.getTopIssues(result.Issues, 10),
		Recommendations: result.Recommendations,
		TrendData:       dg.extractTrendData(historicalData),
		ChartsData:      dg.prepareChartsData(result),
	}

	// Prepare language results
	for lang, langResult := range result.LanguageResults {
		data.LanguageResults = append(data.LanguageResults, LanguageResultData{
			Language:        lang,
			Score:           langResult.QualityScore,
			ScoreColor:      dg.getScoreColor(langResult.QualityScore),
			FilesAnalyzed:   langResult.FilesAnalyzed,
			LinesOfCode:     langResult.LinesOfCode,
			IssuesCount:     len(langResult.Issues),
			ComplexityScore: langResult.ComplexityScore,
			TestCoverage:    langResult.TestCoverage,
			DuplicationRate: langResult.DuplicationRate,
		})
	}

	// Sort language results by score
	sort.Slice(data.LanguageResults, func(i, j int) bool {
		return data.LanguageResults[i].Score < data.LanguageResults[j].Score
	})

	return data
}

func (dg *DashboardGenerator) generateHTML(data *DashboardData) (string, error) {
	tmplStr := dg.getDashboardTemplate()

	tmpl, err := template.New("dashboard").Funcs(template.FuncMap{
		"json": func(v interface{}) template.JS {
			b, _ := json.Marshal(v)
			return template.JS(b)
		},
		"formatFloat": func(f float64) string {
			return fmt.Sprintf("%.1f", f)
		},
		"formatPercent": func(f float64) string {
			return fmt.Sprintf("%.1f%%", f)
		},
	}).Parse(tmplStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (dg *DashboardGenerator) getScoreColor(score float64) string {
	switch {
	case score >= 90:
		return "success"
	case score >= 80:
		return "info"
	case score >= 60:
		return "warning"
	default:
		return "danger"
	}
}

func (dg *DashboardGenerator) summarizeIssues(issues []QualityIssue) map[string]int {
	summary := make(map[string]int)
	for _, issue := range issues {
		summary[issue.Severity]++
	}
	return summary
}

func (dg *DashboardGenerator) getTopIssues(issues []QualityIssue, limit int) []QualityIssue {
	// Sort by severity
	severityOrder := map[string]int{
		"critical": 0,
		"major":    1,
		"minor":    2,
		"info":     3,
	}

	sort.Slice(issues, func(i, j int) bool {
		return severityOrder[issues[i].Severity] < severityOrder[issues[j].Severity]
	})

	if len(issues) > limit {
		return issues[:limit]
	}
	return issues
}

func (dg *DashboardGenerator) extractTrendData(historicalData []*RepoQualityResult) TrendData {
	trend := TrendData{
		Dates:         make([]string, 0),
		Scores:        make([]float64, 0),
		Complexity:    make([]float64, 0),
		Coverage:      make([]float64, 0),
		IssuesCount:   make([]int, 0),
		TechnicalDebt: make([]float64, 0),
	}

	// Sort historical data by timestamp
	sort.Slice(historicalData, func(i, j int) bool {
		return historicalData[i].Timestamp.Before(historicalData[j].Timestamp)
	})

	// Extract last 30 data points
	start := 0
	if len(historicalData) > 30 {
		start = len(historicalData) - 30
	}

	for i := start; i < len(historicalData); i++ {
		result := historicalData[i]
		trend.Dates = append(trend.Dates, result.Timestamp.Format("2006-01-02"))
		trend.Scores = append(trend.Scores, result.OverallScore)
		trend.Complexity = append(trend.Complexity, result.Metrics.AvgComplexity)
		trend.Coverage = append(trend.Coverage, result.Metrics.TestCoverage)
		trend.IssuesCount = append(trend.IssuesCount, len(result.Issues))
		trend.TechnicalDebt = append(trend.TechnicalDebt, result.TechnicalDebt.DebtRatio)
	}

	return trend
}

func (dg *DashboardGenerator) prepareChartsData(result *RepoQualityResult) ChartsData {
	charts := ChartsData{
		IssuesByType:     make(map[string]int),
		IssuesBySeverity: make(map[string]int),
		FilesByLanguage:  make(map[string]int),
		ComplexityDist:   make(map[string]int),
	}

	// Count issues by type and severity
	for _, issue := range result.Issues {
		charts.IssuesByType[issue.Type]++
		charts.IssuesBySeverity[issue.Severity]++
	}

	// Count files by language
	for lang, langResult := range result.LanguageResults {
		charts.FilesByLanguage[lang] = langResult.FilesAnalyzed
	}

	// Complexity distribution
	for _, langResult := range result.LanguageResults {
		complexityRange := dg.getComplexityRange(langResult.ComplexityScore)
		charts.ComplexityDist[complexityRange]++
	}

	return charts
}

func (dg *DashboardGenerator) getComplexityRange(complexity float64) string {
	switch {
	case complexity < 5:
		return "Low (1-5)"
	case complexity < 10:
		return "Medium (5-10)"
	case complexity < 20:
		return "High (10-20)"
	default:
		return "Very High (20+)"
	}
}

func (dg *DashboardGenerator) getDashboardTemplate() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}} - {{.Repository}}</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/css/bootstrap.min.css" rel="stylesheet">
    <script src="https://cdn.jsdelivr.net/npm/chart.js@3.7.0/dist/chart.min.js"></script>
    <style>
        .metric-card {
            transition: transform 0.2s;
        }
        .metric-card:hover {
            transform: translateY(-5px);
        }
        .score-badge {
            font-size: 2rem;
            font-weight: bold;
        }
        .trend-chart {
            height: 300px;
        }
        .issue-item {
            border-left: 4px solid;
            padding-left: 10px;
            margin-bottom: 10px;
        }
        .issue-critical { border-left-color: #dc3545; }
        .issue-major { border-left-color: #fd7e14; }
        .issue-minor { border-left-color: #ffc107; }
        .issue-info { border-left-color: #0dcaf0; }
    </style>
</head>
<body>
    <nav class="navbar navbar-dark bg-dark">
        <div class="container-fluid">
            <span class="navbar-brand mb-0 h1">{{.Title}}</span>
            <span class="text-light">Generated: {{.GeneratedAt}}</span>
        </div>
    </nav>

    <div class="container-fluid mt-4">
        <!-- Overview Section -->
        <div class="row mb-4">
            <div class="col-12">
                <h2>{{.Repository}}</h2>
            </div>
        </div>

        <!-- Main Metrics -->
        <div class="row mb-4">
            <div class="col-md-3">
                <div class="card metric-card">
                    <div class="card-body text-center">
                        <h5 class="card-title">Overall Score</h5>
                        <div class="score-badge text-{{.ScoreColor}}">{{formatFloat .OverallScore}}%</div>
                    </div>
                </div>
            </div>
            <div class="col-md-3">
                <div class="card metric-card">
                    <div class="card-body text-center">
                        <h5 class="card-title">Code Coverage</h5>
                        <div class="score-badge text-info">{{formatFloat .Metrics.TestCoverage}}%</div>
                    </div>
                </div>
            </div>
            <div class="col-md-3">
                <div class="card metric-card">
                    <div class="card-body text-center">
                        <h5 class="card-title">Avg Complexity</h5>
                        <div class="score-badge">{{formatFloat .Metrics.AvgComplexity}}</div>
                    </div>
                </div>
            </div>
            <div class="col-md-3">
                <div class="card metric-card">
                    <div class="card-body text-center">
                        <h5 class="card-title">Total Issues</h5>
                        <div class="score-badge">{{len .TopIssues}}</div>
                    </div>
                </div>
            </div>
        </div>

        <!-- Language Breakdown -->
        <div class="row mb-4">
            <div class="col-12">
                <h3>Language Analysis</h3>
                <div class="table-responsive">
                    <table class="table table-striped">
                        <thead>
                            <tr>
                                <th>Language</th>
                                <th>Score</th>
                                <th>Files</th>
                                <th>Lines</th>
                                <th>Issues</th>
                                <th>Coverage</th>
                                <th>Complexity</th>
                                <th>Duplication</th>
                            </tr>
                        </thead>
                        <tbody>
                            {{range .LanguageResults}}
                            <tr>
                                <td><strong>{{.Language}}</strong></td>
                                <td><span class="badge bg-{{.ScoreColor}}">{{formatFloat .Score}}%</span></td>
                                <td>{{.FilesAnalyzed}}</td>
                                <td>{{.LinesOfCode}}</td>
                                <td>{{.IssuesCount}}</td>
                                <td>{{formatPercent .TestCoverage}}</td>
                                <td>{{formatFloat .ComplexityScore}}</td>
                                <td>{{formatPercent .DuplicationRate}}</td>
                            </tr>
                            {{end}}
                        </tbody>
                    </table>
                </div>
            </div>
        </div>

        <!-- Charts Row -->
        <div class="row mb-4">
            <div class="col-md-6">
                <div class="card">
                    <div class="card-header">
                        <h5 class="mb-0">Quality Score Trend</h5>
                    </div>
                    <div class="card-body">
                        <canvas id="trendChart" class="trend-chart"></canvas>
                    </div>
                </div>
            </div>
            <div class="col-md-6">
                <div class="card">
                    <div class="card-header">
                        <h5 class="mb-0">Issues by Severity</h5>
                    </div>
                    <div class="card-body">
                        <canvas id="severityChart" class="trend-chart"></canvas>
                    </div>
                </div>
            </div>
        </div>

        <!-- Additional Charts Row -->
        <div class="row mb-4">
            <div class="col-md-6">
                <div class="card">
                    <div class="card-header">
                        <h5 class="mb-0">Code Complexity Distribution</h5>
                    </div>
                    <div class="card-body">
                        <canvas id="complexityChart" class="trend-chart"></canvas>
                    </div>
                </div>
            </div>
            <div class="col-md-6">
                <div class="card">
                    <div class="card-header">
                        <h5 class="mb-0">Technical Debt Trend</h5>
                    </div>
                    <div class="card-body">
                        <canvas id="debtChart" class="trend-chart"></canvas>
                    </div>
                </div>
            </div>
        </div>

        <!-- Test Coverage and Metrics -->
        <div class="row mb-4">
            <div class="col-12">
                <div class="card">
                    <div class="card-header">
                        <h5 class="mb-0">Test Coverage & Quality Metrics</h5>
                    </div>
                    <div class="card-body">
                        <div class="row">
                            <div class="col-md-3 text-center">
                                <h6>Line Coverage</h6>
                                <div class="progress" style="height: 25px;">
                                    <div class="progress-bar bg-success" role="progressbar" 
                                         style="width: {{.Metrics.TestCoverage}}%">
                                        {{formatFloat .Metrics.TestCoverage}}%
                                    </div>
                                </div>
                            </div>
                            <div class="col-md-3 text-center">
                                <h6>Technical Debt Ratio</h6>
                                <div class="score-badge text-warning">{{formatFloat .Metrics.TechnicalDebtRatio}}</div>
                                <small>minutes/1000 LOC</small>
                            </div>
                            <div class="col-md-3 text-center">
                                <h6>Maintainability Index</h6>
                                <div class="score-badge text-info">{{formatFloat .Metrics.Maintainability}}</div>
                                <small>0-100 scale</small>
                            </div>
                            <div class="col-md-3 text-center">
                                <h6>Security Score</h6>
                                <div class="score-badge text-{{if ge .Metrics.SecurityScore 90}}success{{else if ge .Metrics.SecurityScore 70}}warning{{else}}danger{{end}}">
                                    {{formatFloat .Metrics.SecurityScore}}%
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <!-- Top Issues -->
        <div class="row mb-4">
            <div class="col-12">
                <h3>Top Issues</h3>
                <div class="card">
                    <div class="card-body">
                        {{range .TopIssues}}
                        <div class="issue-item issue-{{.Severity}}">
                            <div class="d-flex justify-content-between">
                                <div>
                                    <span class="badge bg-{{if eq .Severity "critical"}}danger{{else if eq .Severity "major"}}warning{{else if eq .Severity "minor"}}info{{else}}secondary{{end}}">{{.Severity}}</span>
                                    <strong>{{.File}}:{{.Line}}</strong>
                                    <span class="text-muted">[{{.Rule}}]</span>
                                </div>
                                <small class="text-muted">{{.Tool}}</small>
                            </div>
                            <div>{{.Message}}</div>
                            {{if .Suggestion}}<div class="text-muted"><em>Suggestion: {{.Suggestion}}</em></div>{{end}}
                        </div>
                        {{end}}
                    </div>
                </div>
            </div>
        </div>

        <!-- Recommendations -->
        {{if .Recommendations}}
        <div class="row mb-4">
            <div class="col-12">
                <h3>Recommendations</h3>
                <div class="card">
                    <div class="card-body">
                        <ul>
                            {{range .Recommendations}}
                            <li>{{.}}</li>
                            {{end}}
                        </ul>
                    </div>
                </div>
            </div>
        </div>
        {{end}}
    </div>

    <script>
        // Trend Chart
        const trendCtx = document.getElementById('trendChart').getContext('2d');
        const trendData = {{json .TrendData}};
        new Chart(trendCtx, {
            type: 'line',
            data: {
                labels: trendData.dates,
                datasets: [{
                    label: 'Quality Score',
                    data: trendData.scores,
                    borderColor: 'rgb(75, 192, 192)',
                    tension: 0.1
                }, {
                    label: 'Test Coverage',
                    data: trendData.coverage,
                    borderColor: 'rgb(54, 162, 235)',
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

        // Severity Chart
        const severityCtx = document.getElementById('severityChart').getContext('2d');
        const severityData = {{json .ChartsData.IssuesBySeverity}};
        new Chart(severityCtx, {
            type: 'doughnut',
            data: {
                labels: Object.keys(severityData),
                datasets: [{
                    data: Object.values(severityData),
                    backgroundColor: [
                        'rgb(220, 53, 69)',
                        'rgb(255, 193, 7)',
                        'rgb(13, 202, 240)',
                        'rgb(108, 117, 125)'
                    ]
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false
            }
        });

        // Complexity Distribution Chart
        const complexityCtx = document.getElementById('complexityChart').getContext('2d');
        const complexityData = {{json .ChartsData.ComplexityDist}};
        new Chart(complexityCtx, {
            type: 'bar',
            data: {
                labels: Object.keys(complexityData),
                datasets: [{
                    label: 'Number of Functions',
                    data: Object.values(complexityData),
                    backgroundColor: [
                        'rgb(40, 167, 69)',
                        'rgb(255, 193, 7)',
                        'rgb(253, 126, 20)',
                        'rgb(220, 53, 69)'
                    ]
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    y: {
                        beginAtZero: true
                    }
                }
            }
        });

        // Technical Debt Trend Chart
        const debtCtx = document.getElementById('debtChart').getContext('2d');
        new Chart(debtCtx, {
            type: 'line',
            data: {
                labels: trendData.dates,
                datasets: [{
                    label: 'Technical Debt',
                    data: trendData.technical_debt,
                    borderColor: 'rgb(255, 99, 132)',
                    tension: 0.1
                }, {
                    label: 'Complexity',
                    data: trendData.complexity,
                    borderColor: 'rgb(255, 159, 64)',
                    tension: 0.1
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    y: {
                        beginAtZero: true
                    }
                }
            }
        });
    </script>
</body>
</html>`
}

func (dg *DashboardGenerator) generateAssets() error {
	// Generate custom CSS if needed
	cssContent := dg.getCustomCSS()
	cssPath := filepath.Join(dg.outputDir, "dashboard.css")
	if err := os.WriteFile(cssPath, []byte(cssContent), 0o644); err != nil {
		return fmt.Errorf("failed to write CSS file: %w", err)
	}

	// Generate data export functionality
	jsContent := dg.getExportJS()
	jsPath := filepath.Join(dg.outputDir, "dashboard.js")
	if err := os.WriteFile(jsPath, []byte(jsContent), 0o644); err != nil {
		return fmt.Errorf("failed to write JS file: %w", err)
	}

	return nil
}

func (dg *DashboardGenerator) getCustomCSS() string {
	return `/* Custom styles for quality dashboard */
.dark-mode {
    background-color: #1a1a1a;
    color: #e0e0e0;
}

.dark-mode .card {
    background-color: #2a2a2a;
    border-color: #3a3a3a;
}

.export-buttons {
    position: fixed;
    bottom: 20px;
    right: 20px;
    z-index: 1000;
}

.complexity-heat-map {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(100px, 1fr));
    gap: 5px;
}

.complexity-cell {
    padding: 10px;
    text-align: center;
    border-radius: 4px;
    font-size: 0.8rem;
}

.complexity-low { background-color: #28a745; }
.complexity-medium { background-color: #ffc107; }
.complexity-high { background-color: #fd7e14; }
.complexity-very-high { background-color: #dc3545; }`
}

func (dg *DashboardGenerator) getExportJS() string {
	return `// Dashboard export functionality
function exportToJSON() {
    const data = {
        generated: new Date().toISOString(),
        repository: document.querySelector('h2').textContent,
        metrics: gatherMetrics()
    };
    
    const blob = new Blob([JSON.stringify(data, null, 2)], {type: 'application/json'});
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'quality-metrics.json';
    a.click();
}

function gatherMetrics() {
    // Gather metrics from dashboard
    return {
        score: document.querySelector('.score-badge').textContent,
        timestamp: new Date().toISOString()
    };
}`
}
