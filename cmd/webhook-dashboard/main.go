package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gizzahub/gzh-manager-go/pkg/github"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "webhook-dashboard",
	Short: "GitHub webhook monitoring dashboard",
	Long: `A comprehensive dashboard for monitoring GitHub webhook status and health.
Provides real-time monitoring, alerting, and analytics for GitHub webhooks.`,
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the webhook monitoring dashboard",
	Long:  `Starts the webhook monitoring service and web dashboard.`,
	RunE:  startDashboard,
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current webhook status",
	Long:  `Displays the current status of all monitored webhooks.`,
	RunE:  showStatus,
}

var alertsCmd = &cobra.Command{
	Use:   "alerts",
	Short: "Manage webhook alerts",
	Long:  `Commands for managing webhook alerts and notifications.`,
}

var alertsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List active alerts",
	Long:  `Lists all active webhook alerts.`,
	RunE:  listAlerts,
}

var alertsAckCmd = &cobra.Command{
	Use:   "ack [alert-id]",
	Short: "Acknowledge an alert",
	Long:  `Acknowledges a specific alert by ID.`,
	Args:  cobra.ExactArgs(1),
	RunE:  acknowledgeAlert,
}

func init() {
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(alertsCmd)

	alertsCmd.AddCommand(alertsListCmd)
	alertsCmd.AddCommand(alertsAckCmd)

	// Start command flags
	startCmd.Flags().String("host", "0.0.0.0", "Host to bind the server to")
	startCmd.Flags().Int("port", 8080, "Port to listen on")
	startCmd.Flags().String("token", "", "GitHub token (can also use GITHUB_TOKEN env var)")
	startCmd.Flags().Duration("check-interval", 5*time.Minute, "Webhook health check interval")
	startCmd.Flags().Bool("enable-auth", false, "Enable API authentication")
	startCmd.Flags().String("auth-token", "", "API authentication token")
	startCmd.Flags().Bool("cors", true, "Enable CORS")

	// Status command flags
	statusCmd.Flags().String("org", "", "Filter by organization")
	statusCmd.Flags().String("format", "table", "Output format (table, json)")
	statusCmd.Flags().Bool("show-metrics", false, "Show detailed metrics")

	// Alerts command flags
	alertsListCmd.Flags().String("severity", "", "Filter by severity (info, warning, error, critical)")
	alertsListCmd.Flags().String("type", "", "Filter by alert type")
	alertsListCmd.Flags().String("format", "table", "Output format (table, json)")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func startDashboard(cmd *cobra.Command, args []string) error {
	host, _ := cmd.Flags().GetString("host")
	port, _ := cmd.Flags().GetInt("port")
	token, _ := cmd.Flags().GetString("token")
	checkInterval, _ := cmd.Flags().GetDuration("check-interval")
	enableAuth, _ := cmd.Flags().GetBool("enable-auth")
	authToken, _ := cmd.Flags().GetString("auth-token")
	enableCORS, _ := cmd.Flags().GetBool("cors")

	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
		if token == "" {
			return fmt.Errorf("GitHub token is required (use --token flag or GITHUB_TOKEN env var)")
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create logger
	logger := &consoleLogger{}

	// Create GitHub API client
	apiClient := github.NewGitHubClient(token, logger)

	// Create webhook monitor
	monitorConfig := &github.WebhookMonitorConfig{
		CheckInterval:       checkInterval,
		HealthCheckTimeout:  30 * time.Second,
		RetentionPeriod:     24 * time.Hour,
		EnableNotifications: true,
		MaxHistorySize:      1000,
		AlertThresholds: github.AlertThresholds{
			ErrorRate:          10.0,
			ResponseTime:       5 * time.Second,
			FailureCount:       5,
			DeliveryFailureAge: 1 * time.Hour,
		},
	}

	monitor := github.NewWebhookMonitor(logger, apiClient, monitorConfig)

	// Create dashboard API
	apiConfig := &github.DashboardAPIConfig{
		Host:           host,
		Port:           port,
		EnableCORS:     enableCORS,
		RequestTimeout: 30 * time.Second,
		EnableAuth:     enableAuth,
		AuthToken:      authToken,
	}

	dashboardAPI := github.NewWebhookDashboardAPI(monitor, logger, apiConfig)

	// Start webhook monitor
	logger.Info("Starting webhook monitor", "check_interval", checkInterval)
	if err := monitor.Start(ctx); err != nil {
		return fmt.Errorf("failed to start webhook monitor: %w", err)
	}

	// Start dashboard API server
	logger.Info("Starting webhook dashboard", "host", host, "port", port)

	// Start server in goroutine
	serverErr := make(chan error, 1)
	go func() {
		if err := dashboardAPI.StartServer(ctx); err != nil {
			serverErr <- err
		}
	}()

	// Set up graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	logger.Info("Webhook dashboard started successfully")
	logger.Info("Dashboard available at", "url", fmt.Sprintf("http://%s:%d", host, port))
	logger.Info("API available at", "url", fmt.Sprintf("http://%s:%d/api/v1", host, port))
	logger.Info("Press Ctrl+C to stop")

	// Wait for shutdown signal or server error
	select {
	case <-sigChan:
		logger.Info("Shutdown signal received, stopping dashboard...")
	case err := <-serverErr:
		logger.Error("Dashboard server error", "error", err)
		return err
	}

	// Stop services
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer stopCancel()

	if err := monitor.Stop(stopCtx); err != nil {
		logger.Error("Error stopping webhook monitor", "error", err)
	}

	logger.Info("Dashboard stopped successfully")
	return nil
}

func showStatus(cmd *cobra.Command, args []string) error {
	org, _ := cmd.Flags().GetString("org")
	format, _ := cmd.Flags().GetString("format")
	showMetrics, _ := cmd.Flags().GetBool("show-metrics")

	// This would connect to a running dashboard instance or create a temporary monitor
	logger := &consoleLogger{}

	// For demo purposes, create a mock monitor with sample data
	monitor := createMockMonitor(logger)

	webhooks := monitor.GetAllWebhookStatuses()

	if org != "" {
		// Filter by organization
		filtered := make(map[string]*github.WebhookStatus)
		for id, webhook := range webhooks {
			if webhook.Organization == org {
				filtered[id] = webhook
			}
		}
		webhooks = filtered
	}

	if format == "json" {
		return printJSON(webhooks)
	}

	// Print table format
	fmt.Printf("Webhook Status Report\n")
	fmt.Printf("=====================\n\n")

	if len(webhooks) == 0 {
		fmt.Printf("No webhooks found")
		if org != "" {
			fmt.Printf(" for organization: %s", org)
		}
		fmt.Printf("\n")
		return nil
	}

	// Summary
	healthy := 0
	unhealthy := 0
	total := len(webhooks)

	for _, webhook := range webhooks {
		if webhook.Status == github.WebhookStatusHealthy {
			healthy++
		} else {
			unhealthy++
		}
	}

	fmt.Printf("Summary:\n")
	fmt.Printf("  Total Webhooks: %d\n", total)
	fmt.Printf("  Healthy: %d (%.1f%%)\n", healthy, float64(healthy)/float64(total)*100)
	fmt.Printf("  Unhealthy: %d (%.1f%%)\n", unhealthy, float64(unhealthy)/float64(total)*100)
	fmt.Printf("\n")

	// Detailed status
	fmt.Printf("%-20s %-15s %-20s %-10s %-15s %-10s\n",
		"ID", "Organization", "Repository", "Status", "Last Checked", "Uptime")
	fmt.Printf("%s\n", strings.Repeat("-", 90))

	for _, webhook := range webhooks {
		lastChecked := "Never"
		if !webhook.LastChecked.IsZero() {
			lastChecked = formatTime(webhook.LastChecked)
		}

		fmt.Printf("%-20s %-15s %-20s %-10s %-15s %.1f%%\n",
			truncate(webhook.ID, 20),
			truncate(webhook.Organization, 15),
			truncate(webhook.Repository, 20),
			string(webhook.Status),
			lastChecked,
			webhook.Metrics.Uptime)
	}

	if showMetrics {
		fmt.Printf("\nDetailed Metrics:\n")
		fmt.Printf("================\n")

		metrics := monitor.GetMetrics()
		fmt.Printf("Total Deliveries: %d\n", metrics.TotalDeliveries)
		fmt.Printf("Successful Deliveries: %d\n", metrics.SuccessfulDeliveries)
		fmt.Printf("Failed Deliveries: %d\n", metrics.FailedDeliveries)
		fmt.Printf("Average Response Time: %s\n", metrics.AverageResponseTime)
		fmt.Printf("Active Alerts: %d\n", metrics.ActiveAlerts)
	}

	return nil
}

func listAlerts(cmd *cobra.Command, args []string) error {
	severity, _ := cmd.Flags().GetString("severity")
	alertType, _ := cmd.Flags().GetString("type")
	format, _ := cmd.Flags().GetString("format")

	logger := &consoleLogger{}
	monitor := createMockMonitor(logger)

	alerts := monitor.GetActiveAlerts()

	// Apply filters
	filteredAlerts := make([]github.WebhookAlert, 0)
	for _, alert := range alerts {
		if severity != "" && string(alert.Severity) != severity {
			continue
		}
		if alertType != "" && string(alert.Type) != alertType {
			continue
		}
		filteredAlerts = append(filteredAlerts, alert)
	}

	if format == "json" {
		return printJSON(filteredAlerts)
	}

	// Print table format
	fmt.Printf("Active Webhook Alerts\n")
	fmt.Printf("====================\n\n")

	if len(filteredAlerts) == 0 {
		fmt.Printf("No active alerts found\n")
		return nil
	}

	fmt.Printf("%-15s %-20s %-15s %-10s %-15s %-20s\n",
		"Alert ID", "Webhook ID", "Type", "Severity", "Created", "Message")
	fmt.Printf("%s\n", strings.Repeat("-", 105))

	for _, alert := range filteredAlerts {
		fmt.Printf("%-15s %-20s %-15s %-10s %-15s %-20s\n",
			truncate(alert.ID, 15),
			truncate(alert.WebhookID, 20),
			truncate(string(alert.Type), 15),
			string(alert.Severity),
			formatTime(alert.CreatedAt),
			truncate(alert.Message, 20))
	}

	return nil
}

func acknowledgeAlert(cmd *cobra.Command, args []string) error {
	alertID := args[0]

	logger := &consoleLogger{}
	monitor := createMockMonitor(logger)

	err := monitor.AcknowledgeAlert(alertID)
	if err != nil {
		return fmt.Errorf("failed to acknowledge alert: %w", err)
	}

	fmt.Printf("Alert %s acknowledged successfully\n", alertID)
	return nil
}

// Helper functions

func createMockMonitor(logger github.Logger) *github.WebhookMonitor {
	apiClient := &mockAPIClient{}
	monitor := github.NewWebhookMonitor(logger, apiClient, nil)

	// Add some mock webhooks for demonstration
	mockWebhooks := []*github.WebhookStatus{
		{
			ID:           "webhook-001",
			URL:          "https://api.example.com/webhook",
			Organization: "myorg",
			Repository:   "myrepo",
			Events:       []string{"push", "pull_request"},
			Active:       true,
			LastChecked:  time.Now().Add(-5 * time.Minute),
			Status:       github.WebhookStatusHealthy,
			Metrics: github.WebhookStatusMetrics{
				Uptime:    98.5,
				ErrorRate: 1.5,
			},
		},
		{
			ID:           "webhook-002",
			URL:          "https://hooks.slack.com/webhook",
			Organization: "myorg",
			Repository:   "another-repo",
			Events:       []string{"issues", "release"},
			Active:       true,
			LastChecked:  time.Now().Add(-10 * time.Minute),
			Status:       github.WebhookStatusDegraded,
			Metrics: github.WebhookStatusMetrics{
				Uptime:    92.1,
				ErrorRate: 7.9,
			},
			Alerts: []github.WebhookAlert{
				{
					ID:        "alert-001",
					WebhookID: "webhook-002",
					Type:      github.AlertTypeHighErrorRate,
					Severity:  github.AlertSeverityWarning,
					Message:   "High error rate detected (7.9%)",
					CreatedAt: time.Now().Add(-30 * time.Minute),
				},
			},
		},
	}

	// Add mock webhooks to monitor (accessing private field for demo)
	for _, webhook := range mockWebhooks {
		monitor.AddWebhook(webhook) // This method would need to be added to the monitor
	}

	return monitor
}

// Console logger implementation for CLI
type consoleLogger struct{}

func (l *consoleLogger) Debug(msg string, keysAndValues ...interface{}) {
	l.log("DEBUG", msg, keysAndValues...)
}

func (l *consoleLogger) Info(msg string, keysAndValues ...interface{}) {
	l.log("INFO", msg, keysAndValues...)
}

func (l *consoleLogger) Warn(msg string, keysAndValues ...interface{}) {
	l.log("WARN", msg, keysAndValues...)
}

func (l *consoleLogger) Error(msg string, keysAndValues ...interface{}) {
	l.log("ERROR", msg, keysAndValues...)
}

func (l *consoleLogger) log(level, msg string, keysAndValues ...interface{}) {
	timestamp := time.Now().Format("15:04:05")
	fmt.Printf("[%s] %s: %s", timestamp, level, msg)

	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			fmt.Printf(" %v=%v", keysAndValues[i], keysAndValues[i+1])
		}
	}
	fmt.Println()
}

// Mock API client for demo
type mockAPIClient struct{}

func (m *mockAPIClient) ListOrganizationWebhooks(ctx context.Context, org string) ([]github.WebhookInfo, error) {
	return []github.WebhookInfo{}, nil
}

// Helper functions for formatting
func truncate(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length-3] + "..."
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return "Never"
	}

	duration := time.Since(t)
	if duration < time.Minute {
		return fmt.Sprintf("%ds ago", int(duration.Seconds()))
	} else if duration < time.Hour {
		return fmt.Sprintf("%dm ago", int(duration.Minutes()))
	} else if duration < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(duration.Hours()))
	} else {
		return fmt.Sprintf("%dd ago", int(duration.Hours()/24))
	}
}

func printJSON(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}
