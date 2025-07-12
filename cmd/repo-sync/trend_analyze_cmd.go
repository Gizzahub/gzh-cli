package reposync

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// newTrendAnalyzeCmd creates the trend-analyze subcommand
func newTrendAnalyzeCmd(logger *zap.Logger) *cobra.Command {
	var (
		outputDir        string
		enableAlerts     bool
		alertChannels    []string
		webhookURL       string
		slackWebhook     string
		slackChannel     string
		emailConfig      string
		thresholdsConfig string
	)

	cmd := &cobra.Command{
		Use:   "trend-analyze [repository-path]",
		Short: "Analyze quality trends and generate alerts",
		Long: `Analyze historical quality metrics to identify trends, anomalies, and generate actionable alerts.

This command provides:
- Trend analysis over time with statistical insights
- Anomaly detection using z-score analysis
- Predictive analytics for future quality metrics
- Automated alert generation for quality issues
- Comprehensive trend reports with visualizations

Alert Channels:
- console: Print alerts to console (default)
- file: Save alerts to files
- webhook: Send alerts to webhook endpoint
- slack: Send alerts to Slack channel
- email: Send alerts via email

Threshold Configuration:
You can customize alert thresholds by providing a JSON file with the following structure:
{
  "quality_drop_threshold": 10.0,
  "minimum_quality_score": 60.0,
  "max_complexity": 15.0,
  "complexity_increase_rate": 20.0,
  "minimum_coverage": 70.0,
  "coverage_drop_threshold": 10.0,
  "max_debt_ratio": 50.0,
  "debt_increase_rate": 30.0,
  "min_security_score": 80.0,
  "max_critical_issues": 0
}

Examples:
  # Analyze trends with console alerts
  gz repo-sync trend-analyze ./my-repo
  
  # Generate alerts to multiple channels
  gz repo-sync trend-analyze ./my-repo --alert-channels console,file,slack
  
  # Use webhook for alerts
  gz repo-sync trend-analyze ./my-repo --webhook-url https://example.com/webhook
  
  # Configure Slack alerts
  gz repo-sync trend-analyze ./my-repo --slack-webhook https://hooks.slack.com/... --slack-channel #quality
  
  # Use custom thresholds
  gz repo-sync trend-analyze ./my-repo --thresholds-config ./alert-thresholds.json`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoPath := "."
			if len(args) > 0 {
				repoPath = args[0]
			}

			// Validate repository path
			if err := validateRepositoryPath(repoPath); err != nil {
				return fmt.Errorf("invalid repository path: %w", err)
			}

			// Create trend analyzer
			analyzer := NewTrendAnalyzer(logger, outputDir)

			// Load custom thresholds if provided
			if thresholdsConfig != "" {
				thresholds, err := loadThresholdsFromFile(thresholdsConfig)
				if err != nil {
					return fmt.Errorf("failed to load thresholds: %w", err)
				}
				analyzer.SetThresholds(*thresholds)
			}

			// Setup alert handlers
			if enableAlerts {
				if err := setupAlertHandlers(analyzer, logger, alertChannels, AlertHandlerConfig{
					OutputDir:    outputDir,
					WebhookURL:   webhookURL,
					SlackWebhook: slackWebhook,
					SlackChannel: slackChannel,
					EmailConfig:  emailConfig,
				}); err != nil {
					return fmt.Errorf("failed to setup alert handlers: %w", err)
				}
			}

			// Run quality analysis to get current result
			config := &QualityCheckConfig{
				RepositoryPath: repoPath,
				Threshold:      80,
				OutputFormat:   "json",
				SaveReport:     true,
			}

			qualityAnalyzer, err := NewCodeQualityAnalyzer(logger, config)
			if err != nil {
				return fmt.Errorf("failed to create quality analyzer: %w", err)
			}

			ctx := context.Background()
			currentResult, err := qualityAnalyzer.AnalyzeQuality(ctx)
			if err != nil {
				return fmt.Errorf("quality analysis failed: %w", err)
			}

			// Save current result
			qualityAnalyzer.saveToHistory(currentResult)

			fmt.Printf("🔍 Analyzing quality trends for: %s\n", repoPath)

			// Analyze trends
			alerts, err := analyzer.AnalyzeTrends(ctx, currentResult)
			if err != nil {
				return fmt.Errorf("trend analysis failed: %w", err)
			}

			// Generate trend report
			reportGen := NewTrendReportGenerator(logger, outputDir)
			historicalData, err := analyzer.loadHistoricalData(currentResult.Repository)
			if err != nil {
				logger.Warn("Failed to load historical data", zap.Error(err))
				historicalData = []*QualityResult{}
			}

			report, err := reportGen.GenerateTrendReport(currentResult.Repository, historicalData, alerts)
			if err != nil {
				return fmt.Errorf("failed to generate trend report: %w", err)
			}

			// Print summary
			printTrendSummary(report, alerts)

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVar(&outputDir, "output-dir", "quality-reports", "Output directory for reports and alerts")
	cmd.Flags().BoolVar(&enableAlerts, "enable-alerts", true, "Enable alert generation")
	cmd.Flags().StringSliceVar(&alertChannels, "alert-channels", []string{"console"}, "Alert channels (console,file,webhook,slack,email)")
	cmd.Flags().StringVar(&webhookURL, "webhook-url", "", "Webhook URL for alerts")
	cmd.Flags().StringVar(&slackWebhook, "slack-webhook", "", "Slack webhook URL")
	cmd.Flags().StringVar(&slackChannel, "slack-channel", "#alerts", "Slack channel for alerts")
	cmd.Flags().StringVar(&emailConfig, "email-config", "", "Email configuration file")
	cmd.Flags().StringVar(&thresholdsConfig, "thresholds-config", "", "Alert thresholds configuration file")

	return cmd
}

// AlertHandlerConfig contains configuration for alert handlers
type AlertHandlerConfig struct {
	OutputDir    string
	WebhookURL   string
	SlackWebhook string
	SlackChannel string
	EmailConfig  string
}

// setupAlertHandlers sets up alert handlers based on configuration
func setupAlertHandlers(analyzer *TrendAnalyzer, logger *zap.Logger, channels []string, config AlertHandlerConfig) error {
	for _, channel := range channels {
		switch strings.ToLower(channel) {
		case "console":
			handler := NewConsoleAlertHandler(logger)
			analyzer.AddAlertHandler(handler)

		case "file":
			handler := NewFileAlertHandler(logger, config.OutputDir)
			analyzer.AddAlertHandler(handler)

		case "webhook":
			if config.WebhookURL == "" {
				return fmt.Errorf("webhook URL required for webhook alerts")
			}
			handler := NewWebhookAlertHandler(logger, config.WebhookURL)
			analyzer.AddAlertHandler(handler)

		case "slack":
			if config.SlackWebhook == "" {
				return fmt.Errorf("Slack webhook required for Slack alerts")
			}
			handler := NewSlackAlertHandler(logger, config.SlackWebhook, config.SlackChannel, "Quality Bot")
			analyzer.AddAlertHandler(handler)

		case "email":
			if config.EmailConfig == "" {
				return fmt.Errorf("email configuration required for email alerts")
			}
			emailCfg, err := loadEmailConfig(config.EmailConfig)
			if err != nil {
				return fmt.Errorf("failed to load email config: %w", err)
			}
			handler := NewEmailAlertHandler(logger, *emailCfg)
			analyzer.AddAlertHandler(handler)

		default:
			return fmt.Errorf("unknown alert channel: %s", channel)
		}
	}

	return nil
}

// loadThresholdsFromFile loads alert thresholds from JSON file
func loadThresholdsFromFile(filename string) (*AlertThresholds, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var thresholds AlertThresholds
	if err := json.Unmarshal(data, &thresholds); err != nil {
		return nil, err
	}

	return &thresholds, nil
}

// loadEmailConfig loads email configuration from JSON file
func loadEmailConfig(filename string) (*EmailConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config EmailConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// printTrendSummary prints trend analysis summary
func printTrendSummary(report *TrendReport, alerts []*QualityAlert) {
	fmt.Printf("\n📊 Trend Analysis Summary\n")
	fmt.Printf("Repository: %s\n", report.Repository)
	fmt.Printf("Period: %s to %s\n",
		report.Period.Start.Format("2006-01-02"),
		report.Period.End.Format("2006-01-02"))
	fmt.Printf("Data Points: %d\n", report.Summary.DataPoints)
	fmt.Printf("Overall Trend: %s\n", report.Summary.OverallTrend)

	// Print metric changes
	fmt.Printf("\n📈 Metric Changes:\n")
	fmt.Printf("  Quality Score: %+.1f%%\n", report.Summary.QualityChange)
	fmt.Printf("  Complexity: %+.1f\n", report.Summary.ComplexityChange)
	fmt.Printf("  Test Coverage: %+.1f%%\n", report.Summary.CoverageChange)
	fmt.Printf("  Technical Debt: %+.1f\n", report.Summary.DebtChange)

	// Print improvements
	if len(report.Improvements) > 0 {
		fmt.Printf("\n✅ Improvements:\n")
		for _, imp := range report.Improvements {
			fmt.Printf("  • %s: %.1f → %.1f (%.1f%% improvement)\n",
				imp.Metric, imp.StartValue, imp.EndValue, imp.ImprovementRate)
		}
	}

	// Print degradations
	if len(report.Degradations) > 0 {
		fmt.Printf("\n❌ Degradations:\n")
		for _, deg := range report.Degradations {
			fmt.Printf("  • %s: %.1f → %.1f (%.1f%% degradation)\n",
				deg.Metric, deg.StartValue, deg.EndValue, deg.DegradationRate)
		}
	}

	// Print alerts
	if len(alerts) > 0 {
		fmt.Printf("\n⚠️  Alerts Generated: %d\n", len(alerts))

		// Count by severity
		severityCounts := make(map[AlertSeverity]int)
		for _, alert := range alerts {
			severityCounts[alert.Severity]++
		}

		for severity, count := range severityCounts {
			fmt.Printf("  %s: %d\n", severity, count)
		}
	}

	// Print predictions
	fmt.Printf("\n🔮 Predictions:\n")
	fmt.Printf("  Quality Score: %.1f (%s, %.1f%% confidence)\n",
		report.Predictions.QualityPrediction.NextValue,
		report.Predictions.QualityPrediction.TrendDirection,
		report.Predictions.QualityPrediction.Confidence)
	fmt.Printf("  Complexity: %.1f (%s)\n",
		report.Predictions.ComplexityPrediction.NextValue,
		report.Predictions.ComplexityPrediction.TrendDirection)
	fmt.Printf("  Technical Debt: %.1f (%s)\n",
		report.Predictions.DebtPrediction.NextValue,
		report.Predictions.DebtPrediction.TrendDirection)

	// Print top recommendations
	if len(report.Recommendations) > 0 {
		fmt.Printf("\n💡 Top Recommendations:\n")
		for i, rec := range report.Recommendations {
			if i >= 3 {
				break // Show only top 3
			}
			fmt.Printf("  %d. [%s] %s\n", i+1, rec.Priority, rec.Title)
		}
	}

	// Print report location
	reportPath := filepath.Join("quality-reports",
		fmt.Sprintf("trend-report-%s-%s.html",
			sanitizeFilename(report.Repository),
			report.GeneratedAt.Format("20060102-150405")))
	fmt.Printf("\n📄 Full report saved to: %s\n", reportPath)
}
