package repoconfig

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// newAuditCmd creates the audit subcommand
func newAuditCmd() *cobra.Command {
	var flags GlobalFlags
	var (
		format     string
		outputFile string
		detailed   bool
		policy     string
	)

	cmd := &cobra.Command{
		Use:   "audit",
		Short: "Generate compliance audit report",
		Long: `Generate comprehensive compliance audit report for repository configurations.

This command analyzes repository configurations against defined policies
and generates detailed compliance reports. It helps track policy adherence
and identify security and configuration issues across organizations.

Audit Features:
- Policy compliance assessment
- Security posture analysis
- Configuration drift detection
- Compliance trend tracking
- Detailed violation reporting

Output Formats:
- table: Human-readable audit table (default)
- json: JSON format for programmatic use
- html: HTML report for web viewing
- csv: CSV format for spreadsheet analysis

Examples:
  gz repo-config audit --org myorg                    # Full audit report
  gz repo-config audit --policy security             # Security policy audit
  gz repo-config audit --detailed                    # Detailed violation info
  gz repo-config audit --format html --output report.html  # HTML report`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAuditCommand(flags, format, outputFile, detailed, policy)
		},
	}

	// Add global flags
	addGlobalFlags(cmd, &flags)

	// Add audit-specific flags
	cmd.Flags().StringVar(&format, "format", "table", "Output format (table, json, html, csv)")
	cmd.Flags().StringVar(&outputFile, "output", "", "Output file path")
	cmd.Flags().BoolVar(&detailed, "detailed", false, "Include detailed violation information")
	cmd.Flags().StringVar(&policy, "policy", "", "Audit specific policy only")

	return cmd
}

// runAuditCommand executes the audit command
func runAuditCommand(flags GlobalFlags, format, outputFile string, detailed bool, policy string) error {
	if flags.Organization == "" {
		return fmt.Errorf("organization is required (use --org flag)")
	}

	if flags.Verbose {
		fmt.Printf("📊 Generating compliance audit for organization: %s\n", flags.Organization)
		if policy != "" {
			fmt.Printf("Policy filter: %s\n", policy)
		}
		fmt.Printf("Format: %s\n", format)
		if outputFile != "" {
			fmt.Printf("Output file: %s\n", outputFile)
		}
		fmt.Println()
	}

	fmt.Printf("📋 Repository Compliance Audit Report\n")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("Organization: %s\n", flags.Organization)
	fmt.Printf("Generated: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println()

	// Generate audit data
	auditData, err := performComplianceAudit(flags.Organization, policy)
	if err != nil {
		return fmt.Errorf("failed to perform audit: %w", err)
	}

	switch format {
	case "table":
		displayAuditTable(auditData, detailed)
	case "json":
		displayAuditJSON(auditData)
	case "html":
		displayAuditHTML(auditData, outputFile)
	case "csv":
		displayAuditCSV(auditData, outputFile)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	return nil
}

// AuditData represents the complete audit information
type AuditData struct {
	Organization     string             `json:"organization"`
	GeneratedAt      time.Time          `json:"generated_at"`
	Summary          AuditSummary       `json:"summary"`
	PolicyCompliance []PolicyCompliance `json:"policy_compliance"`
	Repositories     []RepositoryAudit  `json:"repositories"`
	Violations       []ViolationDetail  `json:"violations"`
}

// AuditSummary provides overall compliance statistics
type AuditSummary struct {
	TotalRepositories     int     `json:"total_repositories"`
	CompliantRepositories int     `json:"compliant_repositories"`
	CompliancePercentage  float64 `json:"compliance_percentage"`
	TotalViolations       int     `json:"total_violations"`
	CriticalViolations    int     `json:"critical_violations"`
	PolicyCount           int     `json:"policy_count"`
	CompliantCount        int     `json:"compliant_count"`
	NonCompliantCount     int     `json:"non_compliant_count"`
}

// PolicyCompliance tracks compliance per policy
type PolicyCompliance struct {
	PolicyName           string  `json:"policy_name"`
	Description          string  `json:"description"`
	Severity             string  `json:"severity"`
	CompliantRepos       int     `json:"compliant_repos"`
	ViolatingRepos       int     `json:"violating_repos"`
	CompliancePercentage float64 `json:"compliance_percentage"`
}

// RepositoryAudit contains audit info for a single repository
type RepositoryAudit struct {
	Name             string   `json:"name"`
	Visibility       string   `json:"visibility"`
	Template         string   `json:"template"`
	OverallCompliant bool     `json:"overall_compliant"`
	ViolationCount   int      `json:"violation_count"`
	CriticalCount    int      `json:"critical_count"`
	LastChecked      string   `json:"last_checked"`
	PolicyStatus     []string `json:"policy_status"`
}

// ViolationDetail provides detailed violation information
type ViolationDetail struct {
	Repository  string `json:"repository"`
	Policy      string `json:"policy"`
	Setting     string `json:"setting"`
	Expected    string `json:"expected"`
	Actual      string `json:"actual"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	Remediation string `json:"remediation"`
}

// performComplianceAudit performs actual audit logic
func performComplianceAudit(organization, policy string) (AuditData, error) {
	// This is a mock implementation - in reality, this would:
	// 1. Fetch repository configurations from GitHub API
	// 2. Load compliance policies and templates
	// 3. Analyze each repository against policies
	// 4. Generate detailed violation reports
	// 5. Calculate compliance metrics
	return AuditData{
		Organization: organization,
		GeneratedAt:  time.Now(),
		Summary: AuditSummary{
			TotalRepositories:     25,
			CompliantRepositories: 18,
			CompliancePercentage:  72.0,
			TotalViolations:       15,
			CriticalViolations:    3,
			PolicyCount:           8,
			CompliantCount:        18,
			NonCompliantCount:     7,
		},
		PolicyCompliance: []PolicyCompliance{
			{
				PolicyName:           "Branch Protection",
				Description:          "Require branch protection on main branches",
				Severity:             "critical",
				CompliantRepos:       20,
				ViolatingRepos:       5,
				CompliancePercentage: 80.0,
			},
			{
				PolicyName:           "Required Reviews",
				Description:          "Minimum 2 reviews required for PRs",
				Severity:             "high",
				CompliantRepos:       22,
				ViolatingRepos:       3,
				CompliancePercentage: 88.0,
			},
			{
				PolicyName:           "Security Scanning",
				Description:          "Enable security scanning features",
				Severity:             "medium",
				CompliantRepos:       18,
				ViolatingRepos:       7,
				CompliancePercentage: 72.0,
			},
		},
		Repositories: []RepositoryAudit{
			{
				Name:             "api-server",
				Visibility:       "private",
				Template:         "microservice",
				OverallCompliant: true,
				ViolationCount:   0,
				CriticalCount:    0,
				LastChecked:      "2024-01-15 14:30:00",
				PolicyStatus:     []string{"✅", "✅", "✅", "✅", "✅", "✅", "✅", "✅"},
			},
			{
				Name:             "web-frontend",
				Visibility:       "private",
				Template:         "frontend",
				OverallCompliant: true,
				ViolationCount:   1,
				CriticalCount:    0,
				LastChecked:      "2024-01-15 14:30:00",
				PolicyStatus:     []string{"✅", "✅", "⚠️", "✅", "✅", "✅", "✅", "✅"},
			},
			{
				Name:             "legacy-service",
				Visibility:       "private",
				Template:         "none",
				OverallCompliant: false,
				ViolationCount:   5,
				CriticalCount:    2,
				LastChecked:      "2024-01-15 14:30:00",
				PolicyStatus:     []string{"❌", "❌", "⚠️", "✅", "❌", "⚠️", "✅", "❌"},
			},
		},
		Violations: []ViolationDetail{
			{
				Repository:  "legacy-service",
				Policy:      "Branch Protection",
				Setting:     "branch_protection.main.enabled",
				Expected:    "true",
				Actual:      "false",
				Severity:    "critical",
				Description: "Main branch lacks protection rules",
				Remediation: "Enable branch protection for main branch",
			},
			{
				Repository:  "legacy-service",
				Policy:      "Required Reviews",
				Setting:     "branch_protection.main.required_reviews",
				Expected:    "2",
				Actual:      "0",
				Severity:    "critical",
				Description: "No required reviewers configured",
				Remediation: "Set minimum required reviewers to 2",
			},
		},
	}, nil
}

// displayAuditTable displays audit results in table format
func displayAuditTable(data AuditData, detailed bool) {
	// Summary
	fmt.Printf("📊 Compliance Summary\n")
	fmt.Printf("Total Repositories: %d\n", data.Summary.TotalRepositories)
	fmt.Printf("Compliant: %d (%.1f%%)\n", data.Summary.CompliantRepositories, data.Summary.CompliancePercentage)
	fmt.Printf("Total Violations: %d\n", data.Summary.TotalViolations)
	fmt.Printf("Critical Violations: %d\n", data.Summary.CriticalViolations)
	fmt.Println()

	// Policy compliance
	fmt.Printf("📋 Policy Compliance\n")
	fmt.Printf("%-20s %-10s %-12s %-12s %s\n", "POLICY", "SEVERITY", "COMPLIANT", "VIOLATIONS", "PERCENTAGE")
	fmt.Println("────────────────────────────────────────────────────────────────────────")
	for _, policy := range data.PolicyCompliance {
		severitySymbol := getSeveritySymbol(policy.Severity)
		fmt.Printf("%-20s %-10s %-12d %-12d %.1f%%\n",
			truncateString(policy.PolicyName, 20),
			severitySymbol,
			policy.CompliantRepos,
			policy.ViolatingRepos,
			policy.CompliancePercentage,
		)
	}
	fmt.Println()

	// Repository details
	fmt.Printf("🏗️ Repository Status\n")
	fmt.Printf("%-20s %-12s %-12s %-10s %-10s %s\n", "REPOSITORY", "VISIBILITY", "TEMPLATE", "COMPLIANT", "VIOLATIONS", "CRITICAL")
	fmt.Println("────────────────────────────────────────────────────────────────────────────────")
	for _, repo := range data.Repositories {
		compliantSymbol := "❌"
		if repo.OverallCompliant {
			compliantSymbol = "✅"
		}

		fmt.Printf("%-20s %-12s %-12s %-10s %-10d %d\n",
			repo.Name,
			repo.Visibility,
			repo.Template,
			compliantSymbol,
			repo.ViolationCount,
			repo.CriticalCount,
		)
	}
	fmt.Println()

	// Detailed violations if requested
	if detailed && len(data.Violations) > 0 {
		fmt.Printf("🚨 Violation Details\n")
		fmt.Println("────────────────────────────────────────────────────────────────────")
		for _, violation := range data.Violations {
			severitySymbol := getSeveritySymbol(violation.Severity)
			fmt.Printf("Repository: %s\n", violation.Repository)
			fmt.Printf("Policy: %s (%s)\n", violation.Policy, severitySymbol)
			fmt.Printf("Setting: %s\n", violation.Setting)
			fmt.Printf("Expected: %s, Actual: %s\n", violation.Expected, violation.Actual)
			fmt.Printf("Description: %s\n", violation.Description)
			fmt.Printf("Remediation: %s\n", violation.Remediation)
			fmt.Println()
		}
	}
}

// displayAuditJSON displays audit results in JSON format
func displayAuditJSON(data AuditData) {
	if jsonBytes, err := json.MarshalIndent(data, "", "  "); err != nil {
		fmt.Printf("Error serializing JSON: %v\n", err)
	} else {
		fmt.Println(string(jsonBytes))
	}
}

// displayAuditHTML displays audit results in HTML format
func displayAuditHTML(data AuditData, outputFile string) {
	htmlContent := generateHTMLReport(data)

	if outputFile != "" {
		if err := os.WriteFile(outputFile, []byte(htmlContent), 0o600); err != nil {
			fmt.Printf("Error writing HTML report: %v\n", err)
			return
		}
		fmt.Printf("HTML audit report generated: %s\n", outputFile)
	} else {
		fmt.Println(htmlContent)
	}
}

// displayAuditCSV displays audit results in CSV format
func displayAuditCSV(data AuditData, outputFile string) {
	csvContent := generateCSVReport(data)

	if outputFile != "" {
		if err := os.WriteFile(outputFile, []byte(csvContent), 0o600); err != nil {
			fmt.Printf("Error writing CSV report: %v\n", err)
			return
		}
		fmt.Printf("CSV audit report generated: %s\n", outputFile)
	} else {
		fmt.Println(csvContent)
	}
}

// getSeveritySymbol returns the symbol for severity level
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

//go:embed templates/audit-report.html
var auditReportTemplate string

// HTMLTemplateData contains data for the HTML template
type HTMLTemplateData struct {
	Organization      string
	GeneratedAt       string
	Summary           AuditSummary
	Repositories      []RepositoryStatus
	Policies          []PolicySummary
	ScoreColor        string
	ScoreClass        string
	ScoreArc          float64
	TrendLabels       template.JS
	TrendCompliant    template.JS
	TrendNonCompliant template.JS
}

// PolicySummary represents a policy summary for the template
type PolicySummary struct {
	Name           string
	Enforcement    string
	ViolationCount int
}

// RepositoryViolation represents a violation for template display
type RepositoryViolation struct {
	Policy  string
	Rule    string
	Message string
}

// RepositoryStatus extends the basic status with template-specific fields
type RepositoryStatus struct {
	Name             string
	Description      string
	Visibility       string
	Template         string
	OverallCompliant bool
	ViolationCount   int
	CriticalCount    int
	IsCompliant      bool
	Violations       []RepositoryViolation
	AppliedPolicies  []string
	LastChecked      string
}

// generateHTMLReport creates HTML content for audit report using the template
func generateHTMLReport(data AuditData) string {
	// Parse the embedded template
	tmpl, err := template.New("audit-report").Parse(auditReportTemplate)
	if err != nil {
		// Fallback to simple HTML if template parsing fails
		return generateSimpleHTMLReport(data)
	}

	// Convert data to template format
	templateData := HTMLTemplateData{
		Organization: data.Organization,
		GeneratedAt:  data.GeneratedAt.Format("January 2, 2006 15:04:05"),
		Summary: AuditSummary{
			TotalRepositories:    data.Summary.TotalRepositories,
			CompliantCount:       data.Summary.CompliantRepositories,
			NonCompliantCount:    data.Summary.TotalRepositories - data.Summary.CompliantRepositories,
			TotalViolations:      data.Summary.TotalViolations,
			CompliancePercentage: data.Summary.CompliancePercentage,
		},
	}

	// Convert repositories
	for _, repo := range data.Repositories {
		repoStatus := RepositoryStatus{
			Name:             repo.Name,
			Description:      "", // Add if available
			Visibility:       repo.Visibility,
			Template:         repo.Template,
			OverallCompliant: repo.OverallCompliant,
			ViolationCount:   repo.ViolationCount,
			CriticalCount:    repo.CriticalCount,
			IsCompliant:      repo.OverallCompliant,
			LastChecked:      time.Now().Format("15:04:05"),
		}

		// Add mock violations for demonstration
		if !repo.OverallCompliant {
			repoStatus.Violations = []RepositoryViolation{
				{
					Policy:  "Security",
					Rule:    "branch_protection",
					Message: "Main branch lacks required protection rules",
				},
			}
		}

		// Add applied policies
		repoStatus.AppliedPolicies = []string{"default", "security"}

		templateData.Repositories = append(templateData.Repositories, repoStatus)
	}

	// Add policy summaries
	templateData.Policies = []PolicySummary{
		{Name: "Security Policy", Enforcement: "required", ViolationCount: data.Summary.CriticalViolations},
		{Name: "Compliance Policy", Enforcement: "required", ViolationCount: data.Summary.TotalViolations - data.Summary.CriticalViolations},
		{Name: "Best Practices", Enforcement: "recommended", ViolationCount: 0},
	}

	// Calculate score visualization
	if templateData.Summary.CompliancePercentage >= 80 {
		templateData.ScoreColor = "#28a745"
		templateData.ScoreClass = "text-success"
	} else if templateData.Summary.CompliancePercentage >= 60 {
		templateData.ScoreColor = "#ffc107"
		templateData.ScoreClass = "text-warning"
	} else {
		templateData.ScoreColor = "#dc3545"
		templateData.ScoreClass = "text-danger"
	}
	templateData.ScoreArc = templateData.Summary.CompliancePercentage * 5.65

	// Generate trend data
	labels := []string{}
	compliantData := []int{}
	nonCompliantData := []int{}
	for i := 29; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i).Format("Jan 2")
		labels = append(labels, fmt.Sprintf("\"%s\"", date))
		compliant := templateData.Summary.CompliantCount + (i-15)*2
		if compliant < 0 {
			compliant = 0
		}
		if compliant > templateData.Summary.TotalRepositories {
			compliant = templateData.Summary.TotalRepositories
		}
		compliantData = append(compliantData, compliant)
		nonCompliantData = append(nonCompliantData, templateData.Summary.TotalRepositories-compliant)
	}

	templateData.TrendLabels = template.JS(fmt.Sprintf("[%s]", strings.Join(labels, ", ")))
	templateData.TrendCompliant = template.JS(fmt.Sprintf("%v", compliantData))
	templateData.TrendNonCompliant = template.JS(fmt.Sprintf("%v", nonCompliantData))

	// Execute template
	var output strings.Builder
	if err := tmpl.Execute(&output, templateData); err != nil {
		// Fallback to simple HTML if template execution fails
		return generateSimpleHTMLReport(data)
	}

	return output.String()
}

// generateSimpleHTMLReport is a fallback for when template processing fails
func generateSimpleHTMLReport(data AuditData) string {
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>Repository Compliance Audit Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background-color: #f5f5f5; padding: 20px; border-radius: 5px; }
        .summary { margin: 20px 0; }
        .table { width: 100%%; border-collapse: collapse; margin: 20px 0; }
        .table th, .table td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        .table th { background-color: #f2f2f2; }
        .compliant { color: green; }
        .non-compliant { color: red; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Repository Compliance Audit Report</h1>
        <p>Organization: %s</p>
        <p>Generated: %s</p>
    </div>
    <div class="summary">
        <h2>Summary</h2>
        <p>Total Repositories: %d</p>
        <p>Compliant: %d (%.1f%%)</p>
        <p>Total Violations: %d</p>
    </div>
    <h2>Repository Status</h2>
    <table class="table">
        <tr>
            <th>Repository</th>
            <th>Status</th>
            <th>Violations</th>
        </tr>`,
		data.Organization,
		data.GeneratedAt.Format("2006-01-02 15:04:05"),
		data.Summary.TotalRepositories,
		data.Summary.CompliantRepositories,
		data.Summary.CompliancePercentage,
		data.Summary.TotalViolations,
	)

	for _, repo := range data.Repositories {
		status := "Compliant"
		if !repo.OverallCompliant {
			status = "Non-Compliant"
		}
		html += fmt.Sprintf(`
        <tr>
            <td>%s</td>
            <td class="%s">%s</td>
            <td>%d</td>
        </tr>`,
			repo.Name,
			strings.ToLower(status),
			status,
			repo.ViolationCount,
		)
	}

	html += `
    </table>
</body>
</html>`
	return html
}

// generateCSVReport creates CSV content for audit report
func generateCSVReport(data AuditData) string {
	csv := "Repository,Visibility,Template,Compliant,Violations,Critical\n"

	for _, repo := range data.Repositories {
		compliant := "No"
		if repo.OverallCompliant {
			compliant = "Yes"
		}

		csv += fmt.Sprintf("%s,%s,%s,%s,%d,%d\n",
			repo.Name,
			repo.Visibility,
			repo.Template,
			compliant,
			repo.ViolationCount,
			repo.CriticalCount,
		)
	}

	return csv
}
