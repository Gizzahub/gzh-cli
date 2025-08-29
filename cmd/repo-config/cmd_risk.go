// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package repoconfig

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// RiskAssessment represents a repository risk assessment.
type RiskAssessment struct {
	Repository      string              `json:"repository"`
	OverallScore    float64             `json:"overallScore"`
	Severity        string              `json:"severity"`
	Categories      RiskCategories      `json:"categories"`
	Vulnerabilities []RiskVulnerability `json:"vulnerabilities"`
	Recommendations []string            `json:"recommendations"`
	LastAssessed    time.Time           `json:"lastAssessed"`
}

// RiskCategories represents different risk category scores.
type RiskCategories struct {
	AccessControl       float64 `json:"accessControl"`
	DataProtection      float64 `json:"dataProtection"`
	InfrastructureSec   float64 `json:"infrastructureSecurity"`
	OperationalSecurity float64 `json:"operationalSecurity"`
}

// RiskVulnerability represents a specific vulnerability.
type RiskVulnerability struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Severity    string  `json:"severity"`
	CVSSScore   float64 `json:"cvssScore"`
	Category    string  `json:"category"`
	Remediation string  `json:"remediation"`
}

// runRiskAssessmentCommand executes the risk assessment command.
//
//nolint:unused // 중복 구현으로 현재 사용되지 않음
func runRiskAssessmentCommand(flags GlobalFlags, format string, includeArchived bool, severityFilter, outputFile string, riskThreshold float64) error {
	if flags.Organization == "" {
		return fmt.Errorf("organization is required (use --org flag)")
	}

	if flags.Verbose {
		fmt.Printf("🔍 Performing risk assessment for organization: %s\n", flags.Organization)
		fmt.Printf("Format: %s\n", format)
		fmt.Printf("Include archived: %t\n", includeArchived)

		if severityFilter != "" {
			fmt.Printf("Severity filter: %s\n", severityFilter)
		}

		fmt.Printf("Risk threshold: %.1f\n", riskThreshold)
		fmt.Println()
	}

	fmt.Printf("🛡️  Security Risk Assessment\n")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("Organization: %s\n", flags.Organization)
	fmt.Printf("Assessment Date: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println()

	// Perform risk assessments
	assessments, err := performRiskAssessments(flags.Organization, flags.Token, includeArchived)
	if err != nil {
		return fmt.Errorf("failed to perform risk assessments: %w", err)
	}

	// Apply severity filter if specified
	if severityFilter != "" {
		assessments = filterBySeverity(assessments, severityFilter)
	}

	// Filter by risk threshold
	assessments = filterByRiskThreshold(assessments, riskThreshold)

	if len(assessments) == 0 {
		fmt.Println("✅ No repositories match the specified criteria")
		return nil
	}

	// Generate output based on format
	switch format {
	case "table":
		displayRiskAssessmentTable(assessments)
	case "json":
		err = displayRiskAssessmentJSON(assessments, outputFile)
	case "csv":
		err = displayRiskAssessmentCSV(assessments, outputFile)
	case "html":
		err = displayRiskAssessmentHTML(assessments, outputFile, flags.Organization)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	if err != nil {
		return fmt.Errorf("failed to generate output: %w", err)
	}

	// Display summary
	displayRiskSummary(assessments, riskThreshold)

	return nil
}

// performRiskAssessments performs risk assessments for repositories.
//
//nolint:unused // 중복 구현으로 현재 사용되지 않음
func performRiskAssessments(organization, _ string, includeArchived bool) ([]RiskAssessment, error) { //nolint:unparam // Token unused in current implementation
	_ = organization // organization unused in mock implementation
	// This is a mock implementation. In reality, this would:
	// 1. Fetch repository configurations from GitHub API
	// 2. Analyze security settings and configurations
	// 3. Calculate CVSS scores based on vulnerabilities
	// 4. Generate recommendations
	mockAssessments := []RiskAssessment{
		{
			Repository:   "api-server",
			OverallScore: 8.5,
			Severity:     "high",
			Categories: RiskCategories{
				AccessControl:       7.0,
				DataProtection:      9.0,
				InfrastructureSec:   8.5,
				OperationalSecurity: 9.0,
			},
			Vulnerabilities: []RiskVulnerability{
				{
					ID:          "CVE-2024-0001",
					Title:       "Missing Branch Protection",
					Description: "Main branch lacks required status checks",
					Severity:    "high",
					CVSSScore:   8.5,
					Category:    "access_control",
					Remediation: "Enable required status checks for main branch",
				},
			},
			Recommendations: []string{
				"Enable branch protection rules for main branch",
				"Require signed commits",
				"Add CODEOWNERS file",
			},
			LastAssessed: time.Now(),
		},
		{
			Repository:   "web-frontend",
			OverallScore: 6.2,
			Severity:     "medium",
			Categories: RiskCategories{
				AccessControl:       5.5,
				DataProtection:      6.0,
				InfrastructureSec:   7.0,
				OperationalSecurity: 6.0,
			},
			Vulnerabilities: []RiskVulnerability{
				{
					ID:          "CVE-2024-0002",
					Title:       "Insufficient Access Controls",
					Description: "Repository allows force pushes to protected branches",
					Severity:    "medium",
					CVSSScore:   6.2,
					Category:    "access_control",
					Remediation: "Disable force pushes on protected branches",
				},
			},
			Recommendations: []string{
				"Disable force pushes on protected branches",
				"Enable secret scanning",
				"Add dependency scanning",
			},
			LastAssessed: time.Now(),
		},
		{
			Repository:   "legacy-service",
			OverallScore: 9.1,
			Severity:     "critical",
			Categories: RiskCategories{
				AccessControl:       9.5,
				DataProtection:      8.0,
				InfrastructureSec:   9.0,
				OperationalSecurity: 10.0,
			},
			Vulnerabilities: []RiskVulnerability{
				{
					ID:          "CVE-2024-0003",
					Title:       "Public Repository with Secrets",
					Description: "Public repository contains hardcoded credentials",
					Severity:    "critical",
					CVSSScore:   9.1,
					Category:    "data_protection",
					Remediation: "Remove secrets and make repository private",
				},
			},
			Recommendations: []string{
				"Immediately make repository private",
				"Remove all hardcoded secrets",
				"Implement proper secret management",
				"Rotate exposed credentials",
			},
			LastAssessed: time.Now(),
		},
	}

	return mockAssessments, nil
}

// displayRiskAssessmentTable displays risk assessments in table format.
//
//nolint:unused // 중복 구현으로 현재 사용되지 않음
func displayRiskAssessmentTable(assessments []RiskAssessment) {
	fmt.Printf("%-20s %-10s %-8s %-12s %-15s %s\n",
		"REPOSITORY", "SEVERITY", "SCORE", "VULNS", "TOP CATEGORY", "LAST ASSESSED")
	fmt.Println(strings.Repeat("─", 90))

	for _, assessment := range assessments {
		topCategory := getTopRiskCategory(assessment.Categories)
		vulnerabilityCount := len(assessment.Vulnerabilities)

		severitySymbol := getSeveritySymbol(assessment.Severity)

		fmt.Printf("%-20s %-10s %-8.1f %-12d %-15s %s\n",
			truncateString(assessment.Repository, 20),
			severitySymbol,
			assessment.OverallScore,
			vulnerabilityCount,
			topCategory,
			assessment.LastAssessed.Format("2006-01-02"),
		)
	}
}

// displayRiskAssessmentJSON displays risk assessments in JSON format.
func displayRiskAssessmentJSON(assessments []RiskAssessment, outputFile string) error {
	data := map[string]interface{}{
		"assessments":  assessments,
		"summary":      generateAssessmentSummary(assessments),
		"generated_at": time.Now().Format(time.RFC3339),
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	if outputFile != "" {
		return os.WriteFile(outputFile, jsonBytes, 0o600)
	}

	fmt.Println(string(jsonBytes))

	return nil
}

// displayRiskAssessmentCSV displays risk assessments in CSV format.
func displayRiskAssessmentCSV(assessments []RiskAssessment, outputFile string) error {
	var (
		writer *csv.Writer
		file   *os.File
		err    error
	)

	if outputFile != "" {
		file, err = os.Create(outputFile) //nolint:gosec // Safe file path construction
		if err != nil {
			return err
		}
		defer func() {
			if err := file.Close(); err != nil {
				fmt.Printf("Warning: failed to close file: %v\n", err)
			}
		}()

		writer = csv.NewWriter(file)
	} else {
		writer = csv.NewWriter(os.Stdout)
	}

	defer writer.Flush()

	// Write header
	header := []string{
		"Repository", "Overall Score", "Severity", "Access Control",
		"Data Protection", "Infrastructure Security", "Operational Security",
		"Vulnerability Count", "Last Assessed",
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write data
	for _, assessment := range assessments {
		record := []string{
			assessment.Repository,
			fmt.Sprintf("%.1f", assessment.OverallScore),
			assessment.Severity,
			fmt.Sprintf("%.1f", assessment.Categories.AccessControl),
			fmt.Sprintf("%.1f", assessment.Categories.DataProtection),
			fmt.Sprintf("%.1f", assessment.Categories.InfrastructureSec),
			fmt.Sprintf("%.1f", assessment.Categories.OperationalSecurity),
			fmt.Sprintf("%d", len(assessment.Vulnerabilities)),
			assessment.LastAssessed.Format("2006-01-02"),
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

// displayRiskAssessmentHTML displays risk assessments in HTML format.
func displayRiskAssessmentHTML(assessments []RiskAssessment, outputFile, organization string) error {
	html := generateRiskAssessmentHTML(assessments, organization)

	if outputFile != "" {
		return os.WriteFile(outputFile, []byte(html), 0o600)
	}

	fmt.Println(html)

	return nil
}

// Helper functions

//nolint:unused // 중복 구현으로 현재 사용되지 않음
func filterBySeverity(assessments []RiskAssessment, severity string) []RiskAssessment {
	var filtered []RiskAssessment

	for _, assessment := range assessments {
		if assessment.Severity == severity {
			filtered = append(filtered, assessment)
		}
	}

	return filtered
}

//nolint:unused // 중복 구현으로 현재 사용되지 않음
func filterByRiskThreshold(assessments []RiskAssessment, threshold float64) []RiskAssessment {
	var filtered []RiskAssessment

	for _, assessment := range assessments {
		if assessment.OverallScore >= threshold {
			filtered = append(filtered, assessment)
		}
	}

	return filtered
}

func getSeveritySymbol(severity string) string {
	switch severity {
	case "critical":
		return "🔴 Critical"
	case "high":
		return "🟠 High"
	case "medium":
		return "🟡 Medium"
	case "low":
		return "🟢 Low"
	default:
		return "❓ Unknown"
	}
}

func getTopRiskCategory(categories RiskCategories) string {
	maxValue := categories.AccessControl
	category := "Access Control"

	if categories.DataProtection > maxValue {
		maxValue = categories.DataProtection
		category = "Data Protection"
	}

	if categories.InfrastructureSec > maxValue {
		maxValue = categories.InfrastructureSec
		category = "Infrastructure"
	}

	if categories.OperationalSecurity > maxValue {
		category = "Operational"
	}

	return category
}

func generateAssessmentSummary(assessments []RiskAssessment) map[string]interface{} {
	total := len(assessments)
	severityCounts := make(map[string]int)

	var totalScore float64

	for _, assessment := range assessments {
		severityCounts[assessment.Severity]++
		totalScore += assessment.OverallScore
	}

	avgScore := 0.0
	if total > 0 {
		avgScore = totalScore / float64(total)
	}

	return map[string]interface{}{
		"total_repositories":    total,
		"average_score":         avgScore,
		"severity_distribution": severityCounts,
	}
}

func displayRiskSummary(assessments []RiskAssessment, threshold float64) {
	fmt.Println()
	fmt.Printf("📊 Risk Assessment Summary\n")
	fmt.Printf("Total repositories assessed: %d\n", len(assessments))

	severityCounts := make(map[string]int)

	var totalScore float64

	aboveThreshold := 0

	for _, assessment := range assessments {
		severityCounts[assessment.Severity]++

		totalScore += assessment.OverallScore
		if assessment.OverallScore >= threshold {
			aboveThreshold++
		}
	}

	avgScore := 0.0
	if len(assessments) > 0 {
		avgScore = totalScore / float64(len(assessments))
	}

	fmt.Printf("Average risk score: %.1f\n", avgScore)
	fmt.Printf("Above threshold (%.1f): %d\n", threshold, aboveThreshold)
	fmt.Println()

	fmt.Printf("Severity distribution:\n")
	fmt.Printf("  🔴 Critical: %d\n", severityCounts["critical"])
	fmt.Printf("  🟠 High: %d\n", severityCounts["high"])
	fmt.Printf("  🟡 Medium: %d\n", severityCounts["medium"])
	fmt.Printf("  🟢 Low: %d\n", severityCounts["low"])
}

func generateRiskAssessmentHTML(assessments []RiskAssessment, organization string) string {
	summary := generateAssessmentSummary(assessments)

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Risk Assessment Report - %s</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 0; padding: 20px; background-color: #f5f5f5; }
        .header { background: linear-gradient(135deg, #ff6b6b 0%%, #ee5a24 100%%); color: white; padding: 20px; border-radius: 8px; margin-bottom: 20px; }
        .header h1 { margin: 0; font-size: 2em; }
        .summary { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 15px; margin-bottom: 20px; }
        .card { background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .metric { text-align: center; }
        .metric-value { font-size: 2em; font-weight: bold; color: #ee5a24; }
        .assessment-table { width: 100%%; border-collapse: collapse; background: white; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .assessment-table th, .assessment-table td { padding: 12px; text-align: left; border-bottom: 1px solid #eee; }
        .assessment-table th { background: #f8f9fa; font-weight: bold; }
        .severity-critical { color: #dc3545; font-weight: bold; }
        .severity-high { color: #fd7e14; font-weight: bold; }
        .severity-medium { color: #ffc107; font-weight: bold; }
        .severity-low { color: #28a745; font-weight: bold; }
    </style>
</head>
<body>
    <div class="header">
        <h1>🛡️ Security Risk Assessment Report</h1>
        <p>Organization: %s | Generated: %s</p>
    </div>

    <div class="summary">
        <div class="card">
            <div class="metric">
                <div class="metric-value">%d</div>
                <div>Total Repositories</div>
            </div>
        </div>
        <div class="card">
            <div class="metric">
                <div class="metric-value">%.1f</div>
                <div>Average Risk Score</div>
            </div>
        </div>
        <div class="card">
            <div class="metric">
                <div class="metric-value">%d</div>
                <div>Critical Issues</div>
            </div>
        </div>
    </div>

    <div class="card">
        <h2>📋 Detailed Assessment Results</h2>
        <table class="assessment-table">
            <thead>
                <tr>
                    <th>Repository</th>
                    <th>Overall Score</th>
                    <th>Severity</th>
                    <th>Vulnerabilities</th>
                    <th>Top Risk Category</th>
                    <th>Last Assessed</th>
                </tr>
            </thead>
            <tbody>
                %s
            </tbody>
        </table>
    </div>
</body>
</html>`, organization, organization, time.Now().Format("2006-01-02 15:04:05"),
		len(assessments), summary["average_score"],
		func() int {
			if dist, ok := summary["severity_distribution"].(map[string]int); ok {
				return dist["critical"]
			}
			return 0
		}(),
		generateTableRows(assessments))
}

func generateTableRows(assessments []RiskAssessment) string {
	rows := make([]string, 0, len(assessments))

	for _, assessment := range assessments {
		severityClass := fmt.Sprintf("severity-%s", assessment.Severity)
		topCategory := getTopRiskCategory(assessment.Categories)

		row := fmt.Sprintf(`<tr>
			<td>%s</td>
			<td>%.1f</td>
			<td class="%s">%s</td>
			<td>%d</td>
			<td>%s</td>
			<td>%s</td>
		</tr>`,
			assessment.Repository,
			assessment.OverallScore,
			severityClass,
			strings.ToUpper(assessment.Severity[:1])+assessment.Severity[1:],
			len(assessment.Vulnerabilities),
			topCategory,
			assessment.LastAssessed.Format("2006-01-02"))

		rows = append(rows, row)
	}

	return strings.Join(rows, "")
}
