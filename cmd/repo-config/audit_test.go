package repoconfig

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGenerateHTMLReport(t *testing.T) {
	// Create test data
	testData := AuditData{
		Organization: "test-org",
		GeneratedAt:  time.Now(),
		Summary: AuditSummary{
			TotalRepositories:     10,
			CompliantRepositories: 7,
			CompliancePercentage:  70.0,
			TotalViolations:       5,
			CriticalViolations:    2,
			PolicyCount:           3,
			CompliantCount:        7,
			NonCompliantCount:     3,
		},
		PolicyCompliance: []PolicyCompliance{
			{
				PolicyName:           "Security Policy",
				Description:          "Security requirements",
				Severity:             "critical",
				CompliantRepos:       8,
				ViolatingRepos:       2,
				CompliancePercentage: 80.0,
			},
		},
		Repositories: []RepositoryAudit{
			{
				Name:             "test-repo-1",
				Visibility:       "private",
				Template:         "default",
				OverallCompliant: true,
				ViolationCount:   0,
				CriticalCount:    0,
				LastChecked:      "2024-01-15 14:30:00",
				PolicyStatus:     []string{"✅", "✅", "✅"},
			},
			{
				Name:             "test-repo-2",
				Visibility:       "public",
				Template:         "frontend",
				OverallCompliant: false,
				ViolationCount:   2,
				CriticalCount:    1,
				LastChecked:      "2024-01-15 14:30:00",
				PolicyStatus:     []string{"❌", "✅", "⚠️"},
			},
		},
		Violations: []ViolationDetail{
			{
				Repository:  "test-repo-2",
				Policy:      "Security Policy",
				Setting:     "branch_protection",
				Expected:    "enabled",
				Actual:      "disabled",
				Severity:    "critical",
				Description: "Branch protection is disabled",
				Remediation: "Enable branch protection",
			},
		},
	}

	// Generate HTML report
	html := generateHTMLReport(testData)

	// Verify the report contains expected content
	assert.Contains(t, html, "<!DOCTYPE html>")
	assert.Contains(t, html, "Repository Compliance Audit Report")
	assert.Contains(t, html, "test-org")
	assert.Contains(t, html, "Total Repositories")
	assert.Contains(t, html, "10") // Total repositories

	// Check for Bootstrap CSS
	assert.Contains(t, html, "bootstrap")

	// Check for compliance score visualization
	assert.Contains(t, html, "compliance-score")
	assert.Contains(t, html, "70%") // Compliance percentage

	// Check for repository details
	assert.Contains(t, html, "test-repo-1")
	assert.Contains(t, html, "test-repo-2")

	// Check for policy information
	assert.Contains(t, html, "Security Policy")

	// Check for charts
	assert.Contains(t, html, "trendChart")
	assert.Contains(t, html, "chart.js")

	// Check for filter functionality
	assert.Contains(t, html, "filterTable")
	assert.Contains(t, html, "searchInput")
}

func TestHTMLTemplateData(t *testing.T) {
	// Test score color calculation
	tests := []struct {
		percentage float64
		expected   string
	}{
		{85.0, "#28a745"}, // Green for high compliance
		{70.0, "#ffc107"}, // Yellow for medium compliance
		{40.0, "#dc3545"}, // Red for low compliance
	}

	for _, tt := range tests {
		data := AuditData{
			Organization: "test",
			GeneratedAt:  time.Now(),
			Summary: AuditSummary{
				CompliancePercentage: tt.percentage,
				CompliantCount:       5,
				NonCompliantCount:    5,
				TotalRepositories:    10,
			},
		}

		html := generateHTMLReport(data)

		// Check that the correct color is used
		assert.Contains(t, html, tt.expected)
	}
}

func TestGenerateSimpleHTMLReport(t *testing.T) {
	// Test the fallback simple HTML report
	testData := AuditData{
		Organization: "fallback-org",
		GeneratedAt:  time.Now(),
		Summary: AuditSummary{
			TotalRepositories:     5,
			CompliantRepositories: 3,
			CompliancePercentage:  60.0,
			TotalViolations:       2,
		},
		Repositories: []RepositoryAudit{
			{
				Name:             "repo1",
				OverallCompliant: true,
				ViolationCount:   0,
			},
			{
				Name:             "repo2",
				OverallCompliant: false,
				ViolationCount:   2,
			},
		},
	}

	html := generateSimpleHTMLReport(testData)

	// Verify basic HTML structure
	assert.Contains(t, html, "<!DOCTYPE html>")
	assert.Contains(t, html, "<title>Repository Compliance Audit Report</title>")
	assert.Contains(t, html, "fallback-org")
	assert.Contains(t, html, "Total Repositories: 5")
	assert.Contains(t, html, "Compliant: 3 (60.0%)")
	assert.Contains(t, html, "repo1")
	assert.Contains(t, html, "repo2")
	assert.Contains(t, html, "Compliant")
	assert.Contains(t, html, "Non-Compliant")
}

func TestPolicySummaryGeneration(t *testing.T) {
	// Test that policy summaries are correctly generated
	testData := AuditData{
		Organization: "policy-test",
		GeneratedAt:  time.Now(),
		Summary: AuditSummary{
			TotalRepositories:    10,
			CriticalViolations:   3,
			TotalViolations:      8,
			CompliancePercentage: 70.0,
			CompliantCount:       7,
			NonCompliantCount:    3,
		},
	}

	html := generateHTMLReport(testData)

	// Check for policy summary data
	assert.Contains(t, html, "Security Policy")
	assert.Contains(t, html, "Compliance Policy")
	assert.Contains(t, html, "Best Practices")
	assert.Contains(t, html, "required")
	assert.Contains(t, html, "recommended")
}

func TestTrendDataGeneration(t *testing.T) {
	testData := AuditData{
		Organization: "trend-test",
		GeneratedAt:  time.Now(),
		Summary: AuditSummary{
			TotalRepositories: 20,
			CompliantCount:    15,
			NonCompliantCount: 5,
		},
	}

	html := generateHTMLReport(testData)

	// Check for trend chart data
	assert.Contains(t, html, "TrendLabels")
	assert.Contains(t, html, "TrendCompliant")
	assert.Contains(t, html, "TrendNonCompliant")

	// Verify that trend data contains dates
	assert.True(t, strings.Contains(html, "Jan"))
}

func TestRepositoryViolations(t *testing.T) {
	testData := AuditData{
		Organization: "violation-test",
		GeneratedAt:  time.Now(),
		Summary: AuditSummary{
			TotalRepositories: 2,
			CompliantCount:    1,
			NonCompliantCount: 1,
		},
		Repositories: []RepositoryAudit{
			{
				Name:             "non-compliant-repo",
				OverallCompliant: false,
				ViolationCount:   1,
			},
		},
	}

	html := generateHTMLReport(testData)

	// Check that violations are displayed
	assert.Contains(t, html, "non-compliant-repo")
	assert.Contains(t, html, "violation-item")
	assert.Contains(t, html, "Main branch lacks required protection rules")
}
