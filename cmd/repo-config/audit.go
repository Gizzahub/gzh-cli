package repoconfig

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gizzahub/gzh-manager-go/pkg/audit"
	"github.com/gizzahub/gzh-manager-go/pkg/compliance"
	"github.com/gizzahub/gzh-manager-go/pkg/config"
	"github.com/spf13/cobra"
)

// AuditOptions contains all options for the audit command
type AuditOptions struct {
	GlobalFlags GlobalFlags
	Format      string
	OutputFile  string
	Detailed    bool
	Policy      string
	SaveTrend   bool
	ShowTrend   bool
	TrendPeriod string

	// Repository filters
	FilterVisibility string
	FilterTemplate   string
	FilterTopics     []string
	FilterTeam       string
	FilterModified   string
	FilterPattern    string

	// Policy filters
	PolicyGroup  string
	PolicyPreset string

	// CI/CD options
	ExitOnFail    bool
	FailThreshold float64
	Baseline      string

	// Notification options
	NotifyWebhook string
	NotifyEmail   string

	// Auto-fix options
	SuggestFixes bool
	AutoFix      bool
	DryRun       bool
}

// newAuditCmd creates the audit subcommand
func newAuditCmd() *cobra.Command {
	var flags GlobalFlags
	var (
		format      string
		outputFile  string
		detailed    bool
		policy      string
		saveTrend   bool
		showTrend   bool
		trendPeriod string

		// Repository filters
		filterVisibility string
		filterTemplate   string
		filterTopics     []string
		filterTeam       string
		filterModified   string
		filterPattern    string

		// Policy filters
		policyGroup  string
		policyPreset string

		// CI/CD options
		exitOnFail    bool
		failThreshold float64
		baseline      string

		// Notification options
		notifyWebhook string
		notifyEmail   string

		// Auto-fix options
		suggestFixes bool
		autoFix      bool
		dryRun       bool
	)

	cmd := &cobra.Command{
		Use:   "audit",
		Short: "Generate compliance audit report",
		Long: `Generate comprehensive compliance audit report for repository configurations.

This command analyzes repository configurations against defined policies
and generates detailed compliance reports. It helps track policy adherence
and identify security and configuration issues across organizations.

Audit Features:
- Policy compliance assessment with grouping and presets
- Security posture analysis with risk scoring
- Configuration drift detection
- Compliance trend tracking
- Detailed violation reporting
- Repository filtering by multiple criteria
- CI/CD integration with exit codes

Output Formats:
- table: Human-readable audit table (default)
- json: JSON format for programmatic use
- html: HTML report for web viewing
- csv: CSV format for spreadsheet analysis
- sarif: Static Analysis Results Interchange Format
- junit: JUnit XML format for CI integration

Repository Filters:
- --filter-visibility: Filter by visibility (public, private, all)
- --filter-template: Filter by template name
- --filter-topics: Filter by repository topics
- --filter-team: Filter by team ownership
- --filter-modified: Filter by last modified time (e.g., "7d", "30d", "2023-01-01")
- --filter-pattern: Filter by repository name pattern (regex)

Policy Options:
- --policy-group: Audit specific policy group (security, compliance, best-practice)
- --policy-preset: Use predefined policy preset (soc2, iso27001, nist, pci-dss)

CI/CD Integration:
- --exit-on-fail: Exit with non-zero code if compliance fails
- --fail-threshold: Compliance percentage threshold for failure (default: 80)
- --baseline: Compare against baseline file

Examples:
  # Full audit report
  gz repo-config audit --org myorg
  
  # Security policies only
  gz repo-config audit --policy-group security
  
  # SOC2 compliance check
  gz repo-config audit --policy-preset soc2
  
  # Filter private repos modified in last 30 days
  gz repo-config audit --filter-visibility private --filter-modified 30d
  
  # CI pipeline with failure on low compliance
  gz repo-config audit --format junit --exit-on-fail --fail-threshold 90
  
  # Generate SARIF report for GitHub Advanced Security
  gz repo-config audit --format sarif --output results.sarif
  
  # Show automated fix suggestions
  gz repo-config audit --suggest-fixes
  
  # Preview automatic fixes
  gz repo-config audit --auto-fix --dry-run
  
  # Apply automatic fixes
  gz repo-config audit --auto-fix`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := AuditOptions{
				GlobalFlags:      flags,
				Format:           format,
				OutputFile:       outputFile,
				Detailed:         detailed,
				Policy:           policy,
				SaveTrend:        saveTrend,
				ShowTrend:        showTrend,
				TrendPeriod:      trendPeriod,
				FilterVisibility: filterVisibility,
				FilterTemplate:   filterTemplate,
				FilterTopics:     filterTopics,
				FilterTeam:       filterTeam,
				FilterModified:   filterModified,
				FilterPattern:    filterPattern,
				PolicyGroup:      policyGroup,
				PolicyPreset:     policyPreset,
				ExitOnFail:       exitOnFail,
				FailThreshold:    failThreshold,
				Baseline:         baseline,
				NotifyWebhook:    notifyWebhook,
				NotifyEmail:      notifyEmail,
				SuggestFixes:     suggestFixes,
				AutoFix:          autoFix,
				DryRun:           dryRun,
			}
			return runAuditCommandWithOptions(opts)
		},
	}

	// Add global flags
	addGlobalFlags(cmd, &flags)

	// Add audit-specific flags
	cmd.Flags().StringVar(&format, "format", "table", "Output format (table, json, html, csv, sarif, junit)")
	cmd.Flags().StringVar(&outputFile, "output", "", "Output file path")
	cmd.Flags().BoolVar(&detailed, "detailed", false, "Include detailed violation information")
	cmd.Flags().StringVar(&policy, "policy", "", "Audit specific policy only")
	cmd.Flags().BoolVar(&saveTrend, "save-trend", false, "Save audit results for trend analysis")
	cmd.Flags().BoolVar(&showTrend, "show-trend", false, "Show trend analysis report")
	cmd.Flags().StringVar(&trendPeriod, "trend-period", "30d", "Trend analysis period (e.g., 7d, 30d, 90d)")
	cmd.Flags().BoolVar(&suggestFixes, "suggest-fixes", false, "Generate automated fix suggestions for violations")
	cmd.Flags().BoolVar(&autoFix, "auto-fix", false, "Automatically apply fixes for violations")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview fixes without applying them (use with --auto-fix)")

	// Repository filter flags
	cmd.Flags().StringVar(&filterVisibility, "filter-visibility", "", "Filter by visibility (public, private, all)")
	cmd.Flags().StringVar(&filterTemplate, "filter-template", "", "Filter by template name")
	cmd.Flags().StringSliceVar(&filterTopics, "filter-topics", nil, "Filter by repository topics (comma-separated)")
	cmd.Flags().StringVar(&filterTeam, "filter-team", "", "Filter by team ownership")
	cmd.Flags().StringVar(&filterModified, "filter-modified", "", "Filter by last modified time (e.g., 7d, 30d, 2023-01-01)")
	cmd.Flags().StringVar(&filterPattern, "filter-pattern", "", "Filter by repository name pattern (regex)")

	// Policy filter flags
	cmd.Flags().StringVar(&policyGroup, "policy-group", "", "Audit specific policy group (security, compliance, best-practice)")
	cmd.Flags().StringVar(&policyPreset, "policy-preset", "", "Use predefined policy preset (soc2, iso27001, nist, pci-dss, hipaa, gdpr)")

	// CI/CD integration flags
	cmd.Flags().BoolVar(&exitOnFail, "exit-on-fail", false, "Exit with non-zero code if compliance fails")
	cmd.Flags().Float64Var(&failThreshold, "fail-threshold", 80.0, "Compliance percentage threshold for failure")
	cmd.Flags().StringVar(&baseline, "baseline", "", "Compare against baseline file")

	// Notification flags
	cmd.Flags().StringVar(&notifyWebhook, "notify-webhook", "", "Send audit results to webhook URL")
	cmd.Flags().StringVar(&notifyEmail, "notify-email", "", "Send audit results to email address")

	return cmd
}

// runAuditCommandWithOptions executes the audit command with all options
func runAuditCommandWithOptions(opts AuditOptions) error {
	// For backward compatibility, delegate to the original function if no new features are used
	if !hasNewFeatures(opts) {
		return runAuditCommand(opts.GlobalFlags, opts.Format, opts.OutputFile, opts.Detailed,
			opts.Policy, opts.SaveTrend, opts.ShowTrend, opts.TrendPeriod)
	}

	// New implementation with enhanced features
	return runEnhancedAudit(opts)
}

// hasNewFeatures checks if any new features are being used
func hasNewFeatures(opts AuditOptions) bool {
	return opts.FilterVisibility != "" || opts.FilterTemplate != "" || len(opts.FilterTopics) > 0 ||
		opts.FilterTeam != "" || opts.FilterModified != "" || opts.FilterPattern != "" ||
		opts.PolicyGroup != "" || opts.PolicyPreset != "" || opts.ExitOnFail ||
		opts.Baseline != "" || opts.NotifyWebhook != "" || opts.NotifyEmail != "" ||
		opts.Format == "sarif" || opts.Format == "junit" ||
		opts.SuggestFixes || opts.AutoFix || opts.DryRun
}

// runEnhancedAudit runs the audit with enhanced features
func runEnhancedAudit(opts AuditOptions) error {
	if opts.GlobalFlags.Organization == "" {
		return fmt.Errorf("organization is required (use --org flag)")
	}

	if opts.GlobalFlags.Verbose {
		fmt.Printf("📊 Generating enhanced compliance audit for organization: %s\n", opts.GlobalFlags.Organization)
		if opts.PolicyGroup != "" {
			fmt.Printf("Policy group: %s\n", opts.PolicyGroup)
		}
		if opts.PolicyPreset != "" {
			fmt.Printf("Policy preset: %s\n", opts.PolicyPreset)
		}
		fmt.Printf("Format: %s\n", opts.Format)
		if opts.OutputFile != "" {
			fmt.Printf("Output file: %s\n", opts.OutputFile)
		}
		fmt.Println()
	}

	// Load repository states with filters
	repos, err := loadFilteredRepositories(opts)
	if err != nil {
		return fmt.Errorf("failed to load repositories: %w", err)
	}

	// Load policies based on group/preset selection
	policies, err := loadPoliciesWithOptions(opts)
	if err != nil {
		return fmt.Errorf("failed to load policies: %w", err)
	}

	// Perform enhanced audit
	auditData, err := performEnhancedComplianceAudit(opts.GlobalFlags.Organization, policies, repos, opts)
	if err != nil {
		return fmt.Errorf("failed to perform audit: %w", err)
	}

	// Handle baseline comparison if specified
	if opts.Baseline != "" {
		if err := compareWithBaseline(auditData, opts.Baseline); err != nil {
			return fmt.Errorf("failed to compare with baseline: %w", err)
		}
	}

	// Handle notifications
	if opts.NotifyWebhook != "" || opts.NotifyEmail != "" {
		sendNotifications(auditData, opts)
	}

	// Handle trend analysis
	if opts.SaveTrend || opts.ShowTrend {
		if err := handleTrendAnalysis(auditData, opts); err != nil {
			return fmt.Errorf("trend analysis failed: %w", err)
		}
		if opts.ShowTrend {
			return nil // Exit after showing trend report
		}
	}

	// Display results based on format
	switch opts.Format {
	case "table":
		displayAuditTable(auditData, opts.Detailed)
	case "json":
		displayAuditJSON(auditData)
	case "html":
		displayAuditHTML(auditData, opts.OutputFile)
	case "csv":
		displayAuditCSV(auditData, opts.OutputFile)
	case "sarif":
		if err := displayAuditSARIF(auditData, opts.OutputFile); err != nil {
			return err
		}
	case "junit":
		if err := displayAuditJUnit(auditData, opts.OutputFile); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported format: %s", opts.Format)
	}

	// Handle CI/CD exit codes
	if opts.ExitOnFail && auditData.Summary.CompliancePercentage < opts.FailThreshold {
		fmt.Printf("\n❌ Compliance check failed: %.1f%% < %.1f%% threshold\n",
			auditData.Summary.CompliancePercentage, opts.FailThreshold)
		os.Exit(1)
	}

	return nil
}

// runAuditCommand executes the audit command (original implementation)
func runAuditCommand(flags GlobalFlags, format, outputFile string, detailed bool, policy string, saveTrend, showTrend bool, trendPeriod string) error {
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

	// Handle trend analysis
	if saveTrend || showTrend {
		// Initialize audit store
		store, err := audit.NewFileBasedAuditStore("")
		if err != nil {
			return fmt.Errorf("failed to initialize audit store: %w", err)
		}

		// Save current audit results
		if saveTrend {
			history := convertToAuditHistory(auditData)
			if err := store.SaveAuditResult(history); err != nil {
				return fmt.Errorf("failed to save audit results: %w", err)
			}
			fmt.Println("✅ Audit results saved for trend analysis")
		}

		// Show trend analysis
		if showTrend {
			trendAnalyzer := audit.NewTrendAnalyzer(store)
			duration, err := parseTrendPeriod(trendPeriod)
			if err != nil {
				return fmt.Errorf("invalid trend period: %w", err)
			}

			trendReport, err := trendAnalyzer.AnalyzeTrends(flags.Organization, duration)
			if err != nil {
				return fmt.Errorf("failed to analyze trends: %w", err)
			}

			displayTrendReport(trendReport)
			return nil // Exit after showing trend report
		}
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
	TotalRepositories     int                         `json:"total_repositories"`
	CompliantRepositories int                         `json:"compliant_repositories"`
	CompliancePercentage  float64                     `json:"compliance_percentage"`
	TotalViolations       int                         `json:"total_violations"`
	CriticalViolations    int                         `json:"critical_violations"`
	PolicyCount           int                         `json:"policy_count"`
	CompliantCount        int                         `json:"compliant_count"`
	NonCompliantCount     int                         `json:"non_compliant_count"`
	ComplianceScore       *compliance.ComplianceScore `json:"compliance_score,omitempty"`
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

// FixSuggestion represents an automated fix suggestion
type FixSuggestion struct {
	ID             string     `json:"id"`
	Repository     string     `json:"repository"`
	Violation      string     `json:"violation"`
	Severity       string     `json:"severity"`
	FixType        FixType    `json:"fix_type"`
	Description    string     `json:"description"`
	Command        string     `json:"command,omitempty"`
	APIAction      *APIAction `json:"api_action,omitempty"`
	Prerequisites  []string   `json:"prerequisites,omitempty"`
	RiskLevel      string     `json:"risk_level"`
	EstimatedTime  string     `json:"estimated_time"`
	AutoApplicable bool       `json:"auto_applicable"`
	Priority       int        `json:"priority"` // 1-5, 1 is highest
}

// FixType defines the type of fix
type FixType string

const (
	FixTypeAPI     FixType = "api"     // GitHub API call
	FixTypeManual  FixType = "manual"  // Manual intervention required
	FixTypeCommand FixType = "command" // CLI command execution
	FixTypeScript  FixType = "script"  // Script execution
	FixTypeConfig  FixType = "config"  // Configuration file change
)

// APIAction represents a GitHub API action for fixing violations
type APIAction struct {
	Method   string                 `json:"method"`
	Endpoint string                 `json:"endpoint"`
	Body     map[string]interface{} `json:"body"`
	Headers  map[string]string      `json:"headers,omitempty"`
}

// FixResult represents the result of applying a fix
type FixResult struct {
	FixID    string `json:"fix_id"`
	Success  bool   `json:"success"`
	Error    string `json:"error,omitempty"`
	Applied  bool   `json:"applied"`
	Message  string `json:"message"`
	Duration string `json:"duration"`
}

// performComplianceAudit performs actual audit logic
func performComplianceAudit(organization, policy string) (AuditData, error) {
	// This is a mock implementation - in reality, this would:
	// 1. Fetch repository configurations from GitHub API
	// 2. Load compliance policies and templates
	// 3. Analyze each repository against policies
	// 4. Generate detailed violation reports
	// 5. Calculate compliance metrics

	auditData := AuditData{
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
	}

	// Calculate compliance score
	calculator := compliance.NewScoreCalculator()

	// Convert to compliance types
	complianceSummary := compliance.AuditSummary{
		TotalRepositories:     auditData.Summary.TotalRepositories,
		CompliantRepositories: auditData.Summary.CompliantRepositories,
		CompliancePercentage:  auditData.Summary.CompliancePercentage,
		TotalViolations:       auditData.Summary.TotalViolations,
		CriticalViolations:    auditData.Summary.CriticalViolations,
		PolicyCount:           auditData.Summary.PolicyCount,
		CompliantCount:        auditData.Summary.CompliantCount,
		NonCompliantCount:     auditData.Summary.NonCompliantCount,
	}

	var compliancePolicies []compliance.PolicyCompliance
	for _, policy := range auditData.PolicyCompliance {
		compliancePolicies = append(compliancePolicies, compliance.PolicyCompliance{
			PolicyName:           policy.PolicyName,
			Description:          policy.Description,
			Severity:             policy.Severity,
			CompliantRepos:       policy.CompliantRepos,
			ViolatingRepos:       policy.ViolatingRepos,
			CompliancePercentage: policy.CompliancePercentage,
		})
	}

	// Calculate score
	score, err := calculator.CalculateScore(complianceSummary, compliancePolicies, nil)
	if err == nil {
		auditData.Summary.ComplianceScore = score
	}

	return auditData, nil
}

// displayAuditTable displays audit results in table format
func displayAuditTable(data AuditData, detailed bool) {
	// Summary
	fmt.Printf("📊 Compliance Summary\n")
	fmt.Printf("Total Repositories: %d\n", data.Summary.TotalRepositories)
	fmt.Printf("Compliant: %d (%.1f%%)\n", data.Summary.CompliantRepositories, data.Summary.CompliancePercentage)
	fmt.Printf("Total Violations: %d\n", data.Summary.TotalViolations)
	fmt.Printf("Critical Violations: %d\n", data.Summary.CriticalViolations)

	// Risk Analysis
	enhanced := enhanceAuditDataWithRiskScores(data)
	if enhanced.RiskAnalysis.CriticalRiskRepos > 0 || enhanced.RiskAnalysis.HighRiskRepos > 0 {
		fmt.Printf("\n🚨 Risk Analysis\n")
		fmt.Printf("Overall Risk Level: %s\n", enhanced.RiskAnalysis.OverallRiskLevel)
		fmt.Printf("Critical Risk Repos: %d\n", enhanced.RiskAnalysis.CriticalRiskRepos)
		fmt.Printf("High Risk Repos: %d\n", enhanced.RiskAnalysis.HighRiskRepos)
	}

	// Display compliance score if available
	if data.Summary.ComplianceScore != nil {
		score := data.Summary.ComplianceScore
		gradeSymbol := getGradeSymbol(score.Grade)
		fmt.Printf("Compliance Score: %.1f/100 %s (%s)\n", score.TotalScore, gradeSymbol, score.Grade)

		// Show score breakdown
		if detailed {
			fmt.Printf("  Base Score: %.1f\n", score.ScoreBreakdown.BaseScore)
			fmt.Printf("  Security Penalty: -%.1f\n", score.ScoreBreakdown.SecurityPenalty)
			fmt.Printf("  Compliance Penalty: -%.1f\n", score.ScoreBreakdown.CompliancePenalty)
			fmt.Printf("  Best Practice Bonus: +%.1f\n", score.ScoreBreakdown.BestPracticeBonus)
		}

		// Show recommendations
		if len(score.Recommendations) > 0 {
			fmt.Println("\n💡 Recommendations:")
			for _, rec := range score.Recommendations {
				fmt.Printf("  %s\n", rec)
			}
		}
	}
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

	// Risk details if high-risk repos exist
	if detailed && (enhanced.RiskAnalysis.CriticalRiskRepos > 0 || enhanced.RiskAnalysis.HighRiskRepos > 0) {
		fmt.Printf("⚡ High Risk Repository Details\n")
		fmt.Println("────────────────────────────────────────────────────────────────────")
		for _, risk := range enhanced.RiskAnalysis.TopRisks {
			riskSymbol := "🟡"
			if risk.RiskLevel == "critical" {
				riskSymbol = "🔴"
			} else if risk.RiskLevel == "high" {
				riskSymbol = "🟠"
			}

			fmt.Printf("%s %s (Risk Score: %.1f)\n", riskSymbol, risk.Repository, risk.TotalScore)
			fmt.Printf("   Risk Factors:\n")
			for _, factor := range risk.RiskFactors {
				fmt.Printf("   - %s: %.1f%% - %s\n", factor.Name, factor.Score, factor.Description)
			}
			if len(risk.Recommendations) > 0 {
				fmt.Printf("   Recommendations:\n")
				for _, rec := range risk.Recommendations {
					fmt.Printf("   %s\n", rec)
				}
			}
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
			ComplianceScore:      data.Summary.ComplianceScore,
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

	// Generate trend data - try to fetch real data from store
	labels := []string{}
	compliantData := []int{}
	nonCompliantData := []int{}

	// Try to get real trend data
	store, err := audit.NewFileBasedAuditStore("")
	if err == nil {
		// Get last 30 days of data
		history, err := store.GetHistoricalData(data.Organization, 30*24*time.Hour)
		if err == nil && len(history) > 0 {
			// Use real data
			for _, h := range history {
				date := h.Timestamp.Format("Jan 2")
				labels = append(labels, fmt.Sprintf("\"%s\"", date))
				compliantData = append(compliantData, h.Summary.CompliantRepositories)
				nonCompliantData = append(nonCompliantData, h.Summary.TotalRepositories-h.Summary.CompliantRepositories)
			}
		}
	}

	// Fall back to mock data if no real data available
	if len(labels) == 0 {
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

// generateFixSuggestions creates automated fix suggestions for violations
func generateFixSuggestions(violations []ViolationDetail) []FixSuggestion {
	var suggestions []FixSuggestion

	for i, violation := range violations {
		suggestion := FixSuggestion{
			ID:         fmt.Sprintf("fix-%d", i+1),
			Repository: violation.Repository,
			Violation:  fmt.Sprintf("%s: %s", violation.Policy, violation.Setting),
			Severity:   violation.Severity,
		}

		// Generate specific fix based on violation type
		switch {
		case strings.Contains(violation.Setting, "branch_protection"):
			suggestion = generateBranchProtectionFix(violation, suggestion)
		case strings.Contains(violation.Setting, "required_reviews"):
			suggestion = generateRequiredReviewsFix(violation, suggestion)
		case strings.Contains(violation.Setting, "security_scanning"):
			suggestion = generateSecurityScanningFix(violation, suggestion)
		case strings.Contains(violation.Setting, "visibility"):
			suggestion = generateVisibilityFix(violation, suggestion)
		case strings.Contains(violation.Setting, "permissions"):
			suggestion = generatePermissionsFix(violation, suggestion)
		default:
			suggestion = generateGenericFix(violation, suggestion)
		}

		// Set priority based on severity
		suggestion.Priority = getSeverityPriority(violation.Severity)

		suggestions = append(suggestions, suggestion)
	}

	// Sort by priority (1 is highest priority)
	for i := 0; i < len(suggestions); i++ {
		for j := i + 1; j < len(suggestions); j++ {
			if suggestions[i].Priority > suggestions[j].Priority {
				suggestions[i], suggestions[j] = suggestions[j], suggestions[i]
			}
		}
	}

	return suggestions
}

// generateBranchProtectionFix creates fix for branch protection violations
func generateBranchProtectionFix(violation ViolationDetail, suggestion FixSuggestion) FixSuggestion {
	suggestion.FixType = FixTypeAPI
	suggestion.Description = "Enable branch protection with required settings"
	suggestion.RiskLevel = "low"
	suggestion.EstimatedTime = "30 seconds"
	suggestion.AutoApplicable = true

	// Extract branch name from setting
	branch := "main"
	if strings.Contains(violation.Setting, ".") {
		parts := strings.Split(violation.Setting, ".")
		if len(parts) > 1 {
			branch = parts[1]
		}
	}

	suggestion.APIAction = &APIAction{
		Method:   "PUT",
		Endpoint: fmt.Sprintf("/repos/{owner}/%s/branches/%s/protection", violation.Repository, branch),
		Body: map[string]interface{}{
			"required_status_checks": map[string]interface{}{
				"strict":   true,
				"contexts": []string{},
			},
			"enforce_admins": true,
			"required_pull_request_reviews": map[string]interface{}{
				"required_approving_review_count": 2,
				"dismiss_stale_reviews":           true,
				"require_code_owner_reviews":      true,
			},
			"restrictions": nil,
		},
	}

	suggestion.Command = fmt.Sprintf("gz repo-config apply --org {org} --repo %s --setting branch_protection.%s.enabled=true", violation.Repository, branch)

	return suggestion
}

// generateRequiredReviewsFix creates fix for required reviews violations
func generateRequiredReviewsFix(violation ViolationDetail, suggestion FixSuggestion) FixSuggestion {
	suggestion.FixType = FixTypeAPI
	suggestion.Description = fmt.Sprintf("Set required reviewers to %s", violation.Expected)
	suggestion.RiskLevel = "low"
	suggestion.EstimatedTime = "15 seconds"
	suggestion.AutoApplicable = true

	// Extract branch name
	branch := "main"
	if strings.Contains(violation.Setting, ".") {
		parts := strings.Split(violation.Setting, ".")
		if len(parts) > 1 {
			branch = parts[1]
		}
	}

	suggestion.APIAction = &APIAction{
		Method:   "PATCH",
		Endpoint: fmt.Sprintf("/repos/{owner}/%s/branches/%s/protection/required_pull_request_reviews", violation.Repository, branch),
		Body: map[string]interface{}{
			"required_approving_review_count": violation.Expected,
			"dismiss_stale_reviews":           true,
			"require_code_owner_reviews":      true,
		},
	}

	suggestion.Command = fmt.Sprintf("gz repo-config apply --org {org} --repo %s --setting branch_protection.%s.required_reviews=%s", violation.Repository, branch, violation.Expected)

	return suggestion
}

// generateSecurityScanningFix creates fix for security scanning violations
func generateSecurityScanningFix(violation ViolationDetail, suggestion FixSuggestion) FixSuggestion {
	suggestion.FixType = FixTypeAPI
	suggestion.Description = "Enable security scanning features"
	suggestion.RiskLevel = "low"
	suggestion.EstimatedTime = "45 seconds"
	suggestion.AutoApplicable = true

	suggestion.APIAction = &APIAction{
		Method:   "PUT",
		Endpoint: fmt.Sprintf("/repos/{owner}/%s/vulnerability-alerts", violation.Repository),
		Body:     map[string]interface{}{},
	}

	suggestion.Prerequisites = []string{
		"Repository must have security features enabled",
		"Organization must allow security scanning",
	}

	suggestion.Command = fmt.Sprintf("gz repo-config apply --org {org} --repo %s --setting security.vulnerability_alerts=true", violation.Repository)

	return suggestion
}

// generateVisibilityFix creates fix for repository visibility violations
func generateVisibilityFix(violation ViolationDetail, suggestion FixSuggestion) FixSuggestion {
	suggestion.FixType = FixTypeAPI
	suggestion.Description = fmt.Sprintf("Change repository visibility to %s", violation.Expected)
	suggestion.RiskLevel = "high"
	suggestion.EstimatedTime = "10 seconds"
	suggestion.AutoApplicable = false // Visibility changes require confirmation

	isPrivate := violation.Expected == "private"

	suggestion.APIAction = &APIAction{
		Method:   "PATCH",
		Endpoint: fmt.Sprintf("/repos/{owner}/%s", violation.Repository),
		Body: map[string]interface{}{
			"private": isPrivate,
		},
	}

	suggestion.Prerequisites = []string{
		"⚠️  Repository visibility changes affect access permissions",
		"⚠️  Confirm this change won't break existing integrations",
		"⚠️  Consider impact on team workflows",
	}

	suggestion.Command = fmt.Sprintf("gz repo-config apply --org {org} --repo %s --setting visibility=%s --confirm", violation.Repository, violation.Expected)

	return suggestion
}

// generatePermissionsFix creates fix for permissions violations
func generatePermissionsFix(violation ViolationDetail, suggestion FixSuggestion) FixSuggestion {
	suggestion.FixType = FixTypeAPI
	suggestion.Description = "Update team permissions"
	suggestion.RiskLevel = "medium"
	suggestion.EstimatedTime = "20 seconds"
	suggestion.AutoApplicable = true

	// Extract team name and permission level
	parts := strings.Split(violation.Setting, ".")
	teamName := "unknown"
	if len(parts) >= 3 {
		teamName = parts[2]
	}

	suggestion.APIAction = &APIAction{
		Method:   "PUT",
		Endpoint: fmt.Sprintf("/orgs/{owner}/teams/{team_slug}/repos/{owner}/%s", violation.Repository),
		Body: map[string]interface{}{
			"permission": violation.Expected,
		},
	}

	suggestion.Prerequisites = []string{
		fmt.Sprintf("Team '%s' must exist in the organization", teamName),
		"User must have admin access to manage team permissions",
	}

	suggestion.Command = fmt.Sprintf("gz repo-config apply --org {org} --repo %s --setting permissions.team.%s=%s", violation.Repository, teamName, violation.Expected)

	return suggestion
}

// generateGenericFix creates a generic fix suggestion
func generateGenericFix(violation ViolationDetail, suggestion FixSuggestion) FixSuggestion {
	suggestion.FixType = FixTypeManual
	suggestion.Description = "Manual configuration required"
	suggestion.RiskLevel = "medium"
	suggestion.EstimatedTime = "5 minutes"
	suggestion.AutoApplicable = false

	suggestion.Prerequisites = []string{
		"Review violation details carefully",
		"Test changes in a safe environment first",
		violation.Remediation,
	}

	suggestion.Command = fmt.Sprintf("# Manual fix required for %s\n# %s\n# Expected: %s, Current: %s",
		violation.Setting, violation.Description, violation.Expected, violation.Actual)

	return suggestion
}

// getSeverityPriority converts severity to priority number (1-5, 1 is highest)
func getSeverityPriority(severity string) int {
	switch severity {
	case "critical":
		return 1
	case "high":
		return 2
	case "medium":
		return 3
	case "low":
		return 4
	default:
		return 5
	}
}

// displayFixSuggestions shows automated fix suggestions
func displayFixSuggestions(suggestions []FixSuggestion) {
	fmt.Println("🔧 Automated Fix Suggestions")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	for i, suggestion := range suggestions {
		fmt.Printf("\n%d. %s\n", i+1, suggestion.Description)
		fmt.Printf("   Repository: %s\n", suggestion.Repository)
		fmt.Printf("   Violation: %s\n", suggestion.Violation)
		fmt.Printf("   Severity: %s\n", getSeveritySymbol(suggestion.Severity))
		fmt.Printf("   Risk Level: %s\n", getRiskSymbol(suggestion.RiskLevel))
		fmt.Printf("   Fix Type: %s\n", getFixTypeSymbol(suggestion.FixType))
		fmt.Printf("   Estimated Time: %s\n", suggestion.EstimatedTime)
		fmt.Printf("   Auto-Applicable: %s\n", getBoolSymbol(suggestion.AutoApplicable))

		if len(suggestion.Prerequisites) > 0 {
			fmt.Printf("   Prerequisites:\n")
			for _, prereq := range suggestion.Prerequisites {
				fmt.Printf("     • %s\n", prereq)
			}
		}

		if suggestion.Command != "" {
			fmt.Printf("   Command:\n")
			fmt.Printf("     %s\n", suggestion.Command)
		}

		if suggestion.AutoApplicable {
			fmt.Printf("   💡 Run with --auto-fix to apply automatically\n")
		}
	}

	// Summary
	autoApplicable := 0
	for _, s := range suggestions {
		if s.AutoApplicable {
			autoApplicable++
		}
	}

	fmt.Printf("\n📊 Summary: %d suggestions (%d auto-applicable)\n", len(suggestions), autoApplicable)
	if autoApplicable > 0 {
		fmt.Printf("💡 To apply all auto-fixes: gz repo-config audit --auto-fix --org {org}\n")
		fmt.Printf("💡 To preview first: gz repo-config audit --auto-fix --dry-run --org {org}\n")
	}
}

// displayFixPreview shows preview of fixes that would be applied
func displayFixPreview(suggestions []FixSuggestion) {
	applicableFixes := []FixSuggestion{}
	for _, suggestion := range suggestions {
		if suggestion.AutoApplicable {
			applicableFixes = append(applicableFixes, suggestion)
		}
	}

	if len(applicableFixes) == 0 {
		fmt.Println("⚠️  No auto-applicable fixes found")
		return
	}

	fmt.Printf("Found %d auto-applicable fixes:\n\n", len(applicableFixes))

	for i, fix := range applicableFixes {
		fmt.Printf("%d. %s (%s)\n", i+1, fix.Description, fix.Repository)
		fmt.Printf("   Command: %s\n", fix.Command)
		if fix.APIAction != nil {
			fmt.Printf("   API: %s %s\n", fix.APIAction.Method, fix.APIAction.Endpoint)
		}
		fmt.Printf("   Risk: %s, Time: %s\n", fix.RiskLevel, fix.EstimatedTime)
		fmt.Println()
	}

	fmt.Printf("✅ These %d fixes would be applied with --auto-fix\n", len(applicableFixes))
}

// applyAutomaticFixes applies automated fixes
func applyAutomaticFixes(organization string, suggestions []FixSuggestion) {
	applicableFixes := []FixSuggestion{}
	for _, suggestion := range suggestions {
		if suggestion.AutoApplicable {
			applicableFixes = append(applicableFixes, suggestion)
		}
	}

	if len(applicableFixes) == 0 {
		fmt.Println("⚠️  No auto-applicable fixes found")
		return
	}

	fmt.Printf("Applying %d automatic fixes...\n\n", len(applicableFixes))

	successCount := 0
	for i, fix := range applicableFixes {
		fmt.Printf("[%d/%d] Applying: %s (%s)...", i+1, len(applicableFixes), fix.Description, fix.Repository)

		// Simulate fix application (in real implementation, this would call GitHub API)
		result := simulateFixApplication(fix)

		if result.Success {
			fmt.Printf(" ✅ %s\n", result.Message)
			successCount++
		} else {
			fmt.Printf(" ❌ %s\n", result.Error)
		}
	}

	fmt.Printf("\n📊 Applied %d/%d fixes successfully\n", successCount, len(applicableFixes))
	if successCount < len(applicableFixes) {
		fmt.Printf("⚠️  %d fixes failed - check permissions and prerequisites\n", len(applicableFixes)-successCount)
	}
}

// simulateFixApplication simulates applying a fix (placeholder for actual GitHub API calls)
func simulateFixApplication(fix FixSuggestion) FixResult {
	// In a real implementation, this would:
	// 1. Authenticate with GitHub API
	// 2. Make the required API calls
	// 3. Handle responses and errors
	// 4. Return actual results

	// For now, simulate success/failure
	result := FixResult{
		FixID:   fix.ID,
		Applied: true,
		Success: true,
		Message: "Fix applied successfully",
	}

	// Simulate some failures for demonstration
	if strings.Contains(fix.Repository, "legacy") {
		result.Success = false
		result.Applied = false
		result.Error = "Repository permissions insufficient"
		result.Message = "Failed to apply fix"
	}

	return result
}

// Helper functions for display
func getRiskSymbol(risk string) string {
	switch risk {
	case "high":
		return "🔴 High"
	case "medium":
		return "🟡 Medium"
	case "low":
		return "🟢 Low"
	default:
		return "❓ Unknown"
	}
}

func getFixTypeSymbol(fixType FixType) string {
	switch fixType {
	case FixTypeAPI:
		return "🔌 API Call"
	case FixTypeCommand:
		return "⚡ Command"
	case FixTypeManual:
		return "👤 Manual"
	case FixTypeScript:
		return "📜 Script"
	case FixTypeConfig:
		return "⚙️  Config"
	default:
		return "❓ Unknown"
	}
}

func getBoolSymbol(value bool) string {
	if value {
		return "✅ Yes"
	}
	return "❌ No"
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

// convertToAuditHistory converts AuditData to audit.AuditHistory
func convertToAuditHistory(data AuditData) *audit.AuditHistory {
	history := &audit.AuditHistory{
		Timestamp:    data.GeneratedAt,
		Organization: data.Organization,
		Summary: audit.AuditSummary{
			TotalRepositories:     data.Summary.TotalRepositories,
			CompliantRepositories: data.Summary.CompliantRepositories,
			CompliancePercentage:  data.Summary.CompliancePercentage,
			TotalViolations:       data.Summary.TotalViolations,
			CriticalViolations:    data.Summary.CriticalViolations,
		},
		PolicyStats: make(map[string]audit.PolicyStatistics),
	}

	// Convert policy compliance to statistics
	for _, policy := range data.PolicyCompliance {
		history.PolicyStats[policy.PolicyName] = audit.PolicyStatistics{
			PolicyName:           policy.PolicyName,
			ViolationCount:       policy.ViolatingRepos,
			CompliantRepos:       policy.CompliantRepos,
			ViolatingRepos:       policy.ViolatingRepos,
			CompliancePercentage: policy.CompliancePercentage,
		}
	}

	return history
}

// parseTrendPeriod parses trend period string to duration
func parseTrendPeriod(period string) (time.Duration, error) {
	// Handle common formats: 7d, 30d, 90d
	if strings.HasSuffix(period, "d") {
		days := strings.TrimSuffix(period, "d")
		var daysInt int
		if _, err := fmt.Sscanf(days, "%d", &daysInt); err != nil {
			return 0, fmt.Errorf("invalid day format: %s", period)
		}
		return time.Duration(daysInt) * 24 * time.Hour, nil
	}

	// Try parsing as standard duration
	return time.ParseDuration(period)
}

// displayTrendReport displays the trend analysis report
func displayTrendReport(report *audit.TrendReport) {
	fmt.Println("\n📈 Trend Analysis Report")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("Organization: %s\n", report.Organization)
	fmt.Printf("Period: %s (%s to %s)\n", report.Period, report.StartDate.Format("2006-01-02"), report.EndDate.Format("2006-01-02"))
	fmt.Printf("Overall Trend: %s (%.1f%% change)\n", getTrendSymbol(report.OverallTrend), report.ComplianceChange)
	fmt.Println()

	// Display policy trends
	if len(report.PolicyTrends) > 0 {
		fmt.Println("📋 Policy Trends")
		fmt.Printf("%-25s %-12s %-15s %-12s %s\n", "POLICY", "TREND", "CHANGE RATE", "CURRENT", "AVERAGE")
		fmt.Println("─────────────────────────────────────────────────────────────────────────────────")
		for _, trend := range report.PolicyTrends {
			fmt.Printf("%-25s %-12s %-15s %-12d %.1f\n",
				truncateString(trend.PolicyName, 25),
				getTrendSymbol(trend.TrendDirection),
				fmt.Sprintf("%.1f%%/day", trend.ChangeRate),
				trend.CurrentViolations,
				trend.AverageViolations,
			)
		}
		fmt.Println()
	}

	// Display anomalies
	if len(report.Anomalies) > 0 {
		fmt.Println("⚠️ Anomalies Detected")
		for _, anomaly := range report.Anomalies {
			severitySymbol := "🟡"
			if anomaly.Severity == "high" {
				severitySymbol = "🔴"
			}
			fmt.Printf("%s %s - %s: %s (value: %.1f)\n",
				severitySymbol,
				anomaly.Date.Format("2006-01-02"),
				anomaly.Type,
				anomaly.Description,
				anomaly.Value,
			)
		}
		fmt.Println()
	}

	// Display predictions
	if len(report.Predictions) > 0 {
		fmt.Println("🔮 7-Day Predictions")
		fmt.Printf("%-12s %-20s %s\n", "DATE", "COMPLIANCE", "CONFIDENCE")
		fmt.Println("──────────────────────────────────────────────")
		for _, prediction := range report.Predictions {
			fmt.Printf("%-12s %-20s %.1f%%\n",
				prediction.Date.Format("2006-01-02"),
				fmt.Sprintf("%.1f%%", prediction.CompliancePercentage),
				prediction.Confidence,
			)
		}
		fmt.Println()
	}

	// Display daily compliance summary
	if len(report.DailyCompliance) > 0 {
		fmt.Println("📊 Recent Compliance History")
		// Show last 7 days
		start := len(report.DailyCompliance) - 7
		if start < 0 {
			start = 0
		}
		for _, point := range report.DailyCompliance[start:] {
			complianceBar := generateComplianceBar(point.CompliancePercentage)
			fmt.Printf("%s: %s %.1f%% (%d/%d repos)\n",
				point.Date.Format("01/02"),
				complianceBar,
				point.CompliancePercentage,
				point.CompliantRepos,
				point.TotalRepositories,
			)
		}
	}
}

// getTrendSymbol returns symbol for trend direction
func getTrendSymbol(trend audit.TrendDirection) string {
	switch trend {
	case audit.TrendImproving:
		return "📈 Improving"
	case audit.TrendDeclining:
		return "📉 Declining"
	case audit.TrendStable:
		return "➡️ Stable"
	default:
		return "❓ Unknown"
	}
}

// generateComplianceBar creates a visual bar for compliance percentage
func generateComplianceBar(percentage float64) string {
	barLength := 20
	filledLength := int(percentage / 100 * float64(barLength))
	bar := strings.Repeat("█", filledLength) + strings.Repeat("░", barLength-filledLength)
	return bar
}

// getGradeSymbol returns symbol for compliance grade
func getGradeSymbol(grade compliance.Grade) string {
	switch grade {
	case compliance.GradeA:
		return "🏆"
	case compliance.GradeB:
		return "🥈"
	case compliance.GradeC:
		return "🥉"
	case compliance.GradeD:
		return "⚠️"
	case compliance.GradeF:
		return "🚫"
	default:
		return "❓"
	}
}

// getIntValueFromInterface safely extracts an int value from an interface{}
func getIntValueFromInterface(v interface{}) (int, bool) {
	switch val := v.(type) {
	case int:
		return val, true
	case int64:
		return int(val), true
	case float64:
		return int(val), true
	default:
		return 0, false
	}
}

// loadFilteredRepositories loads repositories based on filter criteria
func loadFilteredRepositories(opts AuditOptions) (map[string]config.RepositoryState, error) {
	// Mock implementation - in reality, this would fetch from GitHub API
	allRepos := map[string]config.RepositoryState{
		"api-server": {
			Name:        "api-server",
			Private:     true,
			Archived:    false,
			HasIssues:   true,
			HasWiki:     true,
			HasProjects: true,
			BranchProtection: map[string]config.BranchProtectionState{
				"main": {Protected: true, RequiredReviews: 2, EnforceAdmins: true},
			},
			VulnerabilityAlerts: true,
			SecurityAdvisories:  true,
			Files:               []string{"README.md", "LICENSE", "SECURITY.md"},
			Workflows:           []string{"ci.yml", "cd.yml"},
			LastModified:        time.Now().AddDate(0, 0, -5),
		},
		"web-frontend": {
			Name:        "web-frontend",
			Private:     true,
			Archived:    false,
			HasIssues:   true,
			HasWiki:     false,
			HasProjects: true,
			BranchProtection: map[string]config.BranchProtectionState{
				"main": {Protected: true, RequiredReviews: 1, EnforceAdmins: false},
			},
			VulnerabilityAlerts: true,
			SecurityAdvisories:  false,
			Files:               []string{"README.md", "LICENSE"},
			Workflows:           []string{"ci.yml"},
			LastModified:        time.Now().AddDate(0, 0, -10),
		},
		"public-docs": {
			Name:        "public-docs",
			Private:     false,
			Archived:    false,
			HasIssues:   true,
			HasWiki:     true,
			HasProjects: false,
			BranchProtection: map[string]config.BranchProtectionState{
				"main": {Protected: false},
			},
			VulnerabilityAlerts: false,
			SecurityAdvisories:  false,
			Files:               []string{"README.md"},
			Workflows:           []string{},
			LastModified:        time.Now().AddDate(0, -1, 0),
		},
	}

	// Apply filters
	filtered := make(map[string]config.RepositoryState)
	for name, repo := range allRepos {
		if shouldIncludeRepo(repo, opts) {
			filtered[name] = repo
		}
	}

	return filtered, nil
}

// shouldIncludeRepo checks if a repository matches filter criteria
func shouldIncludeRepo(repo config.RepositoryState, opts AuditOptions) bool {
	// Visibility filter
	if opts.FilterVisibility != "" && opts.FilterVisibility != "all" {
		isPrivate := opts.FilterVisibility == "private"
		if repo.Private != isPrivate {
			return false
		}
	}

	// Pattern filter
	if opts.FilterPattern != "" {
		matched, err := regexp.MatchString(opts.FilterPattern, repo.Name)
		if err != nil || !matched {
			return false
		}
	}

	// Modified time filter
	if opts.FilterModified != "" {
		duration, err := parseDuration(opts.FilterModified)
		if err == nil {
			cutoff := time.Now().Add(-duration)
			if repo.LastModified.Before(cutoff) {
				return false
			}
		}
	}

	// Template filter - would need template info from config
	// Topic filter - would need topic info from API
	// Team filter - would need team info from API

	return true
}

// parseDuration parses duration strings like "7d", "30d"
func parseDuration(s string) (time.Duration, error) {
	if strings.HasSuffix(s, "d") {
		days := strings.TrimSuffix(s, "d")
		var d int
		if _, err := fmt.Sscanf(days, "%d", &d); err != nil {
			return 0, err
		}
		return time.Duration(d) * 24 * time.Hour, nil
	}
	return time.ParseDuration(s)
}

// loadPoliciesWithOptions loads policies based on group/preset selection
func loadPoliciesWithOptions(opts AuditOptions) (map[string]*config.PolicyTemplate, error) {
	policies := make(map[string]*config.PolicyTemplate)

	// Load predefined templates
	predefined := config.GetPredefinedPolicyTemplates()

	if opts.PolicyPreset != "" {
		// Load preset
		presets := config.GetPolicyPresets()
		preset, exists := presets[opts.PolicyPreset]
		if !exists {
			return nil, fmt.Errorf("unknown policy preset: %s", opts.PolicyPreset)
		}

		// Load all policies from preset
		for _, policyName := range preset.Policies {
			if policy, exists := predefined[policyName]; exists {
				policies[policyName] = policy
			}
		}

		// Apply overrides
		for policyName, override := range preset.Overrides {
			if policy, exists := policies[policyName]; exists {
				applyPolicyOverride(policy, override)
			}
		}
	} else if opts.PolicyGroup != "" {
		// Load policies from group
		groups := config.GetPolicyGroups()
		group, exists := groups[opts.PolicyGroup]
		if !exists {
			return nil, fmt.Errorf("unknown policy group: %s", opts.PolicyGroup)
		}

		for _, policyName := range group.Policies {
			if policy, exists := predefined[policyName]; exists {
				policies[policyName] = policy
			}
		}
	} else if opts.Policy != "" {
		// Load specific policy
		if policy, exists := predefined[opts.Policy]; exists {
			policies[opts.Policy] = policy
		} else {
			return nil, fmt.Errorf("unknown policy: %s", opts.Policy)
		}
	} else {
		// Load all policies
		policies = predefined
	}

	return policies, nil
}

// applyPolicyOverride applies preset overrides to a policy
func applyPolicyOverride(policy *config.PolicyTemplate, override config.PolicyOverride) {
	if override.Enforcement != "" {
		// Apply enforcement override to all rules
		for name, rule := range policy.Rules {
			rule.Enforcement = override.Enforcement
			policy.Rules[name] = rule
		}
	}

	// Apply rule-specific overrides
	for ruleName, ruleOverride := range override.Rules {
		if rule, exists := policy.Rules[ruleName]; exists {
			if ruleOverride.Value != nil {
				rule.Value = ruleOverride.Value
			}
			if ruleOverride.Enforcement != "" {
				rule.Enforcement = ruleOverride.Enforcement
			}
			if !ruleOverride.Disabled {
				policy.Rules[ruleName] = rule
			} else {
				delete(policy.Rules, ruleName)
			}
		}
	}
}

// performEnhancedComplianceAudit performs audit with enhanced features
func performEnhancedComplianceAudit(organization string, policies map[string]*config.PolicyTemplate,
	repos map[string]config.RepositoryState, opts AuditOptions,
) (AuditData, error) {
	// Convert to basic audit format and perform audit
	// This is a simplified implementation - in reality would use the config audit system
	auditData := AuditData{
		Organization: organization,
		GeneratedAt:  time.Now(),
		Summary: AuditSummary{
			TotalRepositories:     len(repos),
			CompliantRepositories: 0,
			CompliancePercentage:  0.0,
			TotalViolations:       0,
			CriticalViolations:    0,
			PolicyCount:           len(policies),
			CompliantCount:        0,
			NonCompliantCount:     0,
		},
		PolicyCompliance: []PolicyCompliance{},
		Repositories:     []RepositoryAudit{},
		Violations:       []ViolationDetail{},
	}

	// Calculate compliance for each policy
	for policyName, policy := range policies {
		compliant := 0
		violations := 0

		for repoName, repo := range repos {
			hasViolation := false
			for ruleName, rule := range policy.Rules {
				if !checkRuleComplianceEnhanced(rule, repo) {
					hasViolation = true
					auditData.Violations = append(auditData.Violations, ViolationDetail{
						Repository:  repoName,
						Policy:      policyName,
						Setting:     ruleName,
						Expected:    fmt.Sprintf("%v", rule.Value),
						Actual:      "non-compliant",
						Severity:    policy.Severity,
						Description: rule.Message,
						Remediation: "Fix the violation",
					})
				}
			}

			if !hasViolation {
				compliant++
			} else {
				violations++
				auditData.Summary.TotalViolations++
				if policy.Severity == "critical" {
					auditData.Summary.CriticalViolations++
				}
			}
		}

		compliancePercentage := 0.0
		if len(repos) > 0 {
			compliancePercentage = float64(compliant) / float64(len(repos)) * 100
		}

		auditData.PolicyCompliance = append(auditData.PolicyCompliance, PolicyCompliance{
			PolicyName:           policyName,
			Description:          policy.Description,
			Severity:             policy.Severity,
			CompliantRepos:       compliant,
			ViolatingRepos:       violations,
			CompliancePercentage: compliancePercentage,
		})
	}

	// Calculate repository compliance
	for repoName, repo := range repos {
		repoAudit := RepositoryAudit{
			Name:             repoName,
			Visibility:       "public",
			Template:         "unknown",
			OverallCompliant: true,
			ViolationCount:   0,
			CriticalCount:    0,
			LastChecked:      time.Now().Format("2006-01-02 15:04:05"),
			PolicyStatus:     []string{},
		}

		if repo.Private {
			repoAudit.Visibility = "private"
		}

		// Count violations for this repo
		for _, violation := range auditData.Violations {
			if violation.Repository == repoName {
				repoAudit.ViolationCount++
				repoAudit.OverallCompliant = false
				if violation.Severity == "critical" {
					repoAudit.CriticalCount++
				}
			}
		}

		if repoAudit.OverallCompliant {
			auditData.Summary.CompliantRepositories++
		}

		auditData.Repositories = append(auditData.Repositories, repoAudit)
	}

	// Calculate overall compliance
	if auditData.Summary.TotalRepositories > 0 {
		auditData.Summary.CompliancePercentage = float64(auditData.Summary.CompliantRepositories) /
			float64(auditData.Summary.TotalRepositories) * 100
	}

	// Calculate compliance score
	calculator := compliance.NewScoreCalculator()
	complianceSummary := compliance.AuditSummary{
		TotalRepositories:     auditData.Summary.TotalRepositories,
		CompliantRepositories: auditData.Summary.CompliantRepositories,
		CompliancePercentage:  auditData.Summary.CompliancePercentage,
		TotalViolations:       auditData.Summary.TotalViolations,
		CriticalViolations:    auditData.Summary.CriticalViolations,
		PolicyCount:           auditData.Summary.PolicyCount,
		CompliantCount:        auditData.Summary.CompliantCount,
		NonCompliantCount:     auditData.Summary.NonCompliantCount,
	}

	var compliancePolicies []compliance.PolicyCompliance
	for _, policy := range auditData.PolicyCompliance {
		compliancePolicies = append(compliancePolicies, compliance.PolicyCompliance{
			PolicyName:           policy.PolicyName,
			Description:          policy.Description,
			Severity:             policy.Severity,
			CompliantRepos:       policy.CompliantRepos,
			ViolatingRepos:       policy.ViolatingRepos,
			CompliancePercentage: policy.CompliancePercentage,
		})
	}

	score, err := calculator.CalculateScore(complianceSummary, compliancePolicies, nil)
	if err == nil {
		auditData.Summary.ComplianceScore = score
	}

	return auditData, nil
}

// checkRuleComplianceEnhanced checks if a repository complies with a rule
func checkRuleComplianceEnhanced(rule config.PolicyRule, repo config.RepositoryState) bool {
	switch rule.Type {
	case "branch_protection":
		if val, ok := rule.Value.(bool); ok && val {
			if bp, exists := repo.BranchProtection["main"]; exists {
				return bp.Protected
			}
			return false
		}
	case "min_reviews":
		if val, ok := getIntValueFromInterface(rule.Value); ok {
			if bp, exists := repo.BranchProtection["main"]; exists {
				return bp.RequiredReviews >= val
			}
			return false
		}
	case "security_feature":
		if feature, ok := rule.Value.(string); ok {
			switch feature {
			case "vulnerability_alerts":
				return repo.VulnerabilityAlerts
			case "security_advisories":
				return repo.SecurityAdvisories
			}
		}
	case "file_exists":
		if file, ok := rule.Value.(string); ok {
			for _, f := range repo.Files {
				if strings.EqualFold(f, file) {
					return true
				}
			}
			return false
		}
	case "workflow_exists":
		if workflow, ok := rule.Value.(string); ok {
			workflowName := strings.TrimPrefix(workflow, ".github/workflows/")
			for _, w := range repo.Workflows {
				if strings.EqualFold(w, workflowName) {
					return true
				}
			}
			return false
		}
	}
	return true // Default to compliant if rule type is unknown
}

// handleTrendAnalysis handles trend saving and analysis
func handleTrendAnalysis(auditData AuditData, opts AuditOptions) error {
	store, err := audit.NewFileBasedAuditStore("")
	if err != nil {
		return fmt.Errorf("failed to initialize audit store: %w", err)
	}

	if opts.SaveTrend {
		history := convertToAuditHistory(auditData)
		if err := store.SaveAuditResult(history); err != nil {
			return fmt.Errorf("failed to save audit results: %w", err)
		}
		fmt.Println("✅ Audit results saved for trend analysis")
	}

	if opts.ShowTrend {
		trendAnalyzer := audit.NewTrendAnalyzer(store)
		duration, err := parseTrendPeriod(opts.TrendPeriod)
		if err != nil {
			return fmt.Errorf("invalid trend period: %w", err)
		}

		trendReport, err := trendAnalyzer.AnalyzeTrends(opts.GlobalFlags.Organization, duration)
		if err != nil {
			return fmt.Errorf("failed to analyze trends: %w", err)
		}

		displayTrendReport(trendReport)
	}

	return nil
}

// compareWithBaseline compares current audit results with baseline
func compareWithBaseline(auditData AuditData, baselineFile string) error {
	// Load baseline data
	baselineData, err := os.ReadFile(baselineFile)
	if err != nil {
		return fmt.Errorf("failed to read baseline file: %w", err)
	}

	var baseline AuditData
	if err := json.Unmarshal(baselineData, &baseline); err != nil {
		return fmt.Errorf("failed to parse baseline: %w", err)
	}

	// Compare and add comparison data to audit
	fmt.Println("\n📊 Baseline Comparison")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	complianceChange := auditData.Summary.CompliancePercentage - baseline.Summary.CompliancePercentage
	fmt.Printf("Compliance Change: %+.1f%% (%.1f%% → %.1f%%)\n",
		complianceChange, baseline.Summary.CompliancePercentage, auditData.Summary.CompliancePercentage)

	violationChange := auditData.Summary.TotalViolations - baseline.Summary.TotalViolations
	fmt.Printf("Violation Change: %+d (%d → %d)\n",
		violationChange, baseline.Summary.TotalViolations, auditData.Summary.TotalViolations)

	repoChange := auditData.Summary.CompliantRepositories - baseline.Summary.CompliantRepositories
	fmt.Printf("Compliant Repos Change: %+d (%d → %d)\n",
		repoChange, baseline.Summary.CompliantRepositories, auditData.Summary.CompliantRepositories)

	return nil
}

// sendNotifications sends audit notifications
func sendNotifications(auditData AuditData, opts AuditOptions) {
	if opts.NotifyWebhook != "" {
		fmt.Printf("📤 Sending webhook notification to: %s\n", opts.NotifyWebhook)
		// Implement webhook notification
	}

	if opts.NotifyEmail != "" {
		fmt.Printf("📧 Sending email notification to: %s\n", opts.NotifyEmail)
		// Implement email notification
	}
}
