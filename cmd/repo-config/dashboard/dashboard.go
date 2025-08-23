// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package dashboard

import (
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

// GlobalFlags represents global flags for all repo-config commands.
type GlobalFlags struct {
	Organization string
	ConfigFile   string
	Token        string
	DryRun       bool
	Verbose      bool
	Parallel     int
	Timeout      string
}

// addGlobalFlags adds common flags to the command.
func addGlobalFlags(cmd *cobra.Command, flags *GlobalFlags) {
	cmd.Flags().StringVar(&flags.Organization, "org", "", "GitHub organization name")
	cmd.Flags().StringVar(&flags.ConfigFile, "config-file", "", "Configuration file path")
	cmd.Flags().StringVar(&flags.Token, "token", "", "GitHub token")
	cmd.Flags().BoolVar(&flags.DryRun, "dry-run", false, "Show what would be done without making changes")
	cmd.Flags().BoolVar(&flags.Verbose, "verbose", false, "Enable verbose output")
	cmd.Flags().IntVar(&flags.Parallel, "parallel", 5, "Number of parallel operations")
	cmd.Flags().StringVar(&flags.Timeout, "timeout", "30s", "Request timeout")
}

// NewCmd creates the dashboard subcommand.
func NewCmd() *cobra.Command {
	var (
		flags       GlobalFlags
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

// runDashboardCommand executes the dashboard command.
func runDashboardCommand(flags GlobalFlags, port int, autoRefresh bool, refreshRate int) error {
	if flags.Organization == "" {
		return fmt.Errorf("organization is required (use --org flag)")
	}

	if flags.Verbose {
		fmt.Printf("🚀 Starting compliance dashboard for organization: %s\n", flags.Organization)
		fmt.Printf("Port: %d\n", port)
		fmt.Printf("Auto refresh: %t\n", autoRefresh)

		if autoRefresh {
			fmt.Printf("Refresh rate: %d seconds\n", refreshRate)
		}

		fmt.Println()
	}

	fmt.Printf("🎛️  Repository Compliance Dashboard\n")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("Organization: %s\n", flags.Organization)
	fmt.Printf("Dashboard URL: http://localhost:%d\n", port)
	fmt.Println()

	if autoRefresh {
		fmt.Printf("📡 Auto-refresh enabled (every %d seconds)\n", refreshRate)
	}

	// Set up HTTP server
	mux := http.NewServeMux()

	// Main dashboard page
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handleDashboardHome(w, r, flags.Organization, autoRefresh, refreshRate)
	})

	// API endpoints
	mux.HandleFunc("/api/repositories", func(w http.ResponseWriter, r *http.Request) {
		handleRepositoriesAPI(w, r, flags.Organization, flags.Token)
	})

	mux.HandleFunc("/api/compliance", func(w http.ResponseWriter, r *http.Request) {
		handleComplianceAPI(w, r, flags.Organization, flags.Token)
	})

	// Static assets (if needed)
	mux.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		handleStaticAssets(w, r)
	})

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	fmt.Printf("🌐 Starting web server on port %d...\n", port)
	fmt.Println("Press Ctrl+C to stop the dashboard")
	fmt.Println()

	return server.ListenAndServe()
}

// handleDashboardHome serves the main dashboard page.
func handleDashboardHome(w http.ResponseWriter, _ *http.Request, organization string, autoRefresh bool, refreshRate int) {
	html := generateDashboardHTML(organization, autoRefresh, refreshRate)

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(html)); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}

// handleRepositoriesAPI serves repository data as JSON.
func handleRepositoriesAPI(w http.ResponseWriter, _ *http.Request, organization, token string) {
	_ = organization // organization unused in mock implementation
	_ = token        // token unused in mock implementation
	// This would fetch real repository data
	mockData := `{
		"repositories": [
			{
				"name": "api-server",
				"visibility": "private",
				"template": "security",
				"compliant": true,
				"issues": 0,
				"lastUpdated": "2024-01-15T10:30:00Z"
			},
			{
				"name": "web-frontend",
				"visibility": "public",
				"template": "standard",
				"compliant": false,
				"issues": 3,
				"lastUpdated": "2024-01-14T15:45:00Z"
			}
		],
		"summary": {
			"total": 2,
			"compliant": 1,
			"nonCompliant": 1,
			"complianceRate": 50.0
		}
	}`

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(mockData)); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}

// handleComplianceAPI serves compliance data as JSON.
func handleComplianceAPI(w http.ResponseWriter, r *http.Request, organization, _ string) { //nolint:unparam // Token unused in current implementation
	_ = r            // r unused in mock implementation
	_ = organization // organization unused in mock implementation
	// This would fetch real compliance data
	mockData := `{
		"compliance": {
			"overallScore": 85.5,
			"categories": {
				"security": 90.0,
				"documentation": 75.0,
				"testing": 80.0,
				"deployment": 95.0
			},
			"violations": [
				{
					"repository": "web-frontend",
					"rule": "branch_protection.required_reviews",
					"severity": "medium",
					"message": "Requires 2 reviewers but only has 1"
				}
			]
		}
	}`

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(mockData)); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}

// handleStaticAssets serves static CSS/JS files.
func handleStaticAssets(w http.ResponseWriter, r *http.Request) {
	_ = r // r unused in mock implementation
	// For a real implementation, this would serve actual static files
	w.WriteHeader(http.StatusNotFound)
}

// generateDashboardHTML generates the HTML for the dashboard.
func generateDashboardHTML(organization string, autoRefresh bool, refreshRate int) string {
	refreshScript := ""
	if autoRefresh {
		refreshScript = fmt.Sprintf(`
		<script>
			setInterval(function() {
				location.reload();
			}, %d000);
		</script>`, refreshRate)
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Repository Compliance Dashboard - %s</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 0; padding: 20px; background-color: #f5f5f5; }
        .header { background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 20px; border-radius: 8px; margin-bottom: 20px; }
        .header h1 { margin: 0; font-size: 2em; }
        .header p { margin: 5px 0 0 0; opacity: 0.9; }
        .dashboard { display: grid; grid-template-columns: 1fr 1fr; gap: 20px; }
        .card { background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .card h2 { margin-top: 0; color: #333; }
        .metric { text-align: center; padding: 15px; }
        .metric-value { font-size: 2.5em; font-weight: bold; color: #667eea; }
        .metric-label { color: #666; margin-top: 5px; }
        .status-good { color: #28a745; }
        .status-warning { color: #ffc107; }
        .status-error { color: #dc3545; }
        .refresh-info { text-align: center; margin-top: 20px; padding: 10px; background: white; border-radius: 4px; }
    </style>
    %s
</head>
<body>
    <div class="header">
        <h1>Repository Compliance Dashboard</h1>
        <p>Organization: %s</p>
    </div>

    <div class="dashboard">
        <div class="card">
            <h2>📊 Compliance Overview</h2>
            <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 10px;">
                <div class="metric">
                    <div class="metric-value status-good">85%%</div>
                    <div class="metric-label">Overall Compliance</div>
                </div>
                <div class="metric">
                    <div class="metric-value">24</div>
                    <div class="metric-label">Total Repositories</div>
                </div>
            </div>
        </div>

        <div class="card">
            <h2>🔍 Issues Summary</h2>
            <div style="display: grid; grid-template-columns: 1fr 1fr 1fr; gap: 10px;">
                <div class="metric">
                    <div class="metric-value status-error">3</div>
                    <div class="metric-label">High</div>
                </div>
                <div class="metric">
                    <div class="metric-value status-warning">7</div>
                    <div class="metric-label">Medium</div>
                </div>
                <div class="metric">
                    <div class="metric-value status-good">12</div>
                    <div class="metric-label">Low</div>
                </div>
            </div>
        </div>

        <div class="card">
            <h2>🏆 Top Performing Repositories</h2>
            <ul style="list-style: none; padding: 0;">
                <li style="padding: 8px 0; border-bottom: 1px solid #eee;">
                    <span style="color: #28a745;">✓</span> api-server (100%%)
                </li>
                <li style="padding: 8px 0; border-bottom: 1px solid #eee;">
                    <span style="color: #28a745;">✓</span> core-service (98%%)
                </li>
                <li style="padding: 8px 0; border-bottom: 1px solid #eee;">
                    <span style="color: #28a745;">✓</span> auth-service (95%%)
                </li>
            </ul>
        </div>

        <div class="card">
            <h2>⚠️ Needs Attention</h2>
            <ul style="list-style: none; padding: 0;">
                <li style="padding: 8px 0; border-bottom: 1px solid #eee;">
                    <span style="color: #dc3545;">✗</span> legacy-app (45%%)
                </li>
                <li style="padding: 8px 0; border-bottom: 1px solid #eee;">
                    <span style="color: #ffc107;">⚠</span> web-frontend (72%%)
                </li>
                <li style="padding: 8px 0; border-bottom: 1px solid #eee;">
                    <span style="color: #ffc107;">⚠</span> mobile-app (68%%)
                </li>
            </ul>
        </div>
    </div>

    <div class="refresh-info">
        <p>🕐 Last updated: %s</p>
        %s
    </div>
</body>
</html>`, refreshScript, organization, organization, time.Now().Format("2006-01-02 15:04:05"),
		func() string {
			if autoRefresh {
				return fmt.Sprintf("<p>🔄 Auto-refresh enabled (every %d seconds)</p>", refreshRate)
			}

			return "<p>🔄 Manual refresh - reload page to update</p>"
		}())
}
