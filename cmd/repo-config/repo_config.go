package repoconfig

import (
	"github.com/spf13/cobra"
)

// NewRepoConfigCmd creates the repo-config command with subcommands
func NewRepoConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repo-config",
		Short: "GitHub repository configuration management",
		Long: `Manage GitHub repository configurations across organizations.

This command provides tools for managing repository settings, security policies,
and compliance across entire GitHub organizations using infrastructure-as-code
principles.

Key Features:
- Apply consistent configuration across repositories
- Manage security policies and branch protection rules
- Template-based configuration management
- Compliance auditing and reporting
- Dry-run mode for safe changes

Examples:
  gz repo-config list                    # List repositories with current settings
  gz repo-config apply                   # Apply configuration to repositories
  gz repo-config validate               # Validate configuration files
  gz repo-config diff                   # Show differences between current and target
  gz repo-config audit                  # Generate compliance audit report
  gz repo-config webhook                # Manage repository webhooks
  gz repo-config dashboard              # Start real-time compliance dashboard
  gz repo-config risk-assessment        # Perform CVSS-based risk assessment`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	// Add subcommands
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newApplyCmd())
	cmd.AddCommand(newValidateCmd())
	cmd.AddCommand(newDiffCmd())
	cmd.AddCommand(newAuditCmd())
	cmd.AddCommand(newTemplateCmd())
	cmd.AddCommand(newWebhookCmd())
	cmd.AddCommand(newDashboardCmd())
	cmd.AddCommand(newRiskAssessmentCmd())

	return cmd
}

// Global flags for all repo-config commands
type GlobalFlags struct {
	Organization string
	ConfigFile   string
	Token        string
	DryRun       bool
	Verbose      bool
	Parallel     int
	Timeout      string
}

// addGlobalFlags adds common flags to a command
func addGlobalFlags(cmd *cobra.Command, flags *GlobalFlags) {
	cmd.Flags().StringVarP(&flags.Organization, "org", "o", "", "GitHub organization name")
	cmd.Flags().StringVarP(&flags.ConfigFile, "config", "c", "", "Configuration file path")
	cmd.Flags().StringVarP(&flags.Token, "token", "t", "", "GitHub personal access token")
	cmd.Flags().BoolVar(&flags.DryRun, "dry-run", false, "Preview changes without applying")
	cmd.Flags().BoolVarP(&flags.Verbose, "verbose", "v", false, "Verbose output")
	cmd.Flags().IntVar(&flags.Parallel, "parallel", 5, "Number of parallel operations")
	cmd.Flags().StringVar(&flags.Timeout, "timeout", "30s", "API timeout duration")
}

// newDashboardCmd creates the dashboard subcommand
func newDashboardCmd() *cobra.Command {
	var flags GlobalFlags
	var (
		port        int
		autoRefresh bool
		refreshRate int
	)

	cmd := &cobra.Command{
		Use:   "dashboard",
		Short: "Start real-time compliance dashboard",
		Long: `Start a web-based dashboard for real-time repository compliance monitoring.

The dashboard provides:
- Real-time compliance status across repositories
- Configuration drift detection
- Security policy violations
- Interactive configuration management
- Historical compliance trends

Features:
- Live repository status updates
- Configurable auto-refresh intervals
- Filter and search capabilities
- Export compliance reports
- Visual configuration comparison

Examples:
  gz repo-config dashboard --org myorg                    # Start dashboard
  gz repo-config dashboard --port 8080                    # Custom port
  gz repo-config dashboard --auto-refresh                 # Auto refresh enabled
  gz repo-config dashboard --refresh-rate 30              # Custom refresh rate`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDashboardCommand(flags, port, autoRefresh, refreshRate)
		},
	}

	// Add global flags
	addGlobalFlags(cmd, &flags)

	// Add dashboard-specific flags
	cmd.Flags().IntVar(&port, "port", 8080, "Dashboard server port")
	cmd.Flags().BoolVar(&autoRefresh, "auto-refresh", false, "Enable automatic refresh")
	cmd.Flags().IntVar(&refreshRate, "refresh-rate", 60, "Auto refresh rate in seconds")

	return cmd
}

// newRiskAssessmentCmd creates the risk-assessment subcommand
func newRiskAssessmentCmd() *cobra.Command {
	var flags GlobalFlags
	var (
		format          string
		includeArchived bool
		severityFilter  string
		outputFile      string
		riskThreshold   float64
	)

	cmd := &cobra.Command{
		Use:   "risk-assessment",
		Short: "Perform CVSS-based risk assessment",
		Long: `Perform comprehensive risk assessment using CVSS scoring methodology.

This command analyzes repository configurations and security settings to
provide risk scores and recommendations based on industry standards.

Risk Assessment Features:
- CVSS-based vulnerability scoring
- Configuration weakness detection
- Security policy compliance analysis
- Risk trend analysis over time
- Actionable remediation recommendations

Assessment Categories:
- Access Control (IAM, permissions, branch protection)
- Data Protection (encryption, secrets, visibility)
- Infrastructure Security (webhooks, integrations)
- Operational Security (monitoring, logging, auditing)

Output Formats:
- table: Human-readable risk assessment table
- json: Structured data for integration
- csv: Spreadsheet-compatible format
- html: Detailed HTML report

Examples:
  gz repo-config risk-assessment --org myorg              # Full assessment
  gz repo-config risk-assessment --severity high          # High severity only
  gz repo-config risk-assessment --format html            # HTML report
  gz repo-config risk-assessment --output report.html     # Save to file`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRiskAssessmentCommand(flags, format, includeArchived, severityFilter, outputFile, riskThreshold)
		},
	}

	// Add global flags
	addGlobalFlags(cmd, &flags)

	// Add risk assessment-specific flags
	cmd.Flags().StringVar(&format, "format", "table", "Output format (table, json, csv, html)")
	cmd.Flags().BoolVar(&includeArchived, "include-archived", false, "Include archived repositories")
	cmd.Flags().StringVar(&severityFilter, "severity", "", "Filter by severity (low, medium, high, critical)")
	cmd.Flags().StringVar(&outputFile, "output", "", "Output file path")
	cmd.Flags().Float64Var(&riskThreshold, "risk-threshold", 7.0, "Risk score threshold for critical issues")

	return cmd
}
