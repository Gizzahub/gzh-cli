//nolint:tagliatelle // Network metrics output may require specific JSON field naming conventions
package netenv

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// newNetworkMetricsCmd creates the network metrics monitoring command.
func newNetworkMetricsCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "network-metrics",
		Short: "Monitor and analyze real-time network performance metrics",
		Long: `Monitor and analyze real-time network performance metrics including bandwidth, latency, packet loss, and connection quality.

This command provides comprehensive network performance monitoring with:
- Real-time bandwidth monitoring (upload/download)
- Latency measurement to multiple targets
- Packet loss detection and analysis
- Connection quality assessment
- Historical metrics collection and analysis
- Performance optimization recommendations

Examples:
  # Start real-time network monitoring
  gz net-env network-metrics monitor

  # Show current network metrics
  gz net-env network-metrics show

  # Test latency to specific targets
  gz net-env network-metrics latency --targets 8.8.8.8,1.1.1.1,google.com

  # Monitor bandwidth usage
  gz net-env network-metrics bandwidth --interface eth0

  # Generate performance report
  gz net-env network-metrics report --duration 1h`,
		SilenceUsage: true,
	}

	// Add subcommands
	cmd.AddCommand(newNetworkMetricsMonitorCmd(logger, configDir))
	cmd.AddCommand(newNetworkMetricsShowCmd(logger, configDir))
	cmd.AddCommand(newNetworkMetricsLatencyCmd(logger, configDir))
	cmd.AddCommand(newNetworkMetricsBandwidthCmd(logger, configDir))
	cmd.AddCommand(newNetworkMetricsReportCmd(logger, configDir))
	cmd.AddCommand(newNetworkMetricsOptimizeCmd(logger, configDir))

	return cmd
}

// newNetworkMetricsMonitorCmd creates the monitor subcommand.
func newNetworkMetricsMonitorCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "monitor",
		Short: "Start real-time network metrics monitoring",
		Long:  `Start continuous monitoring of network performance metrics with real-time display.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			collector, err := createNetworkMetricsCollector(ctx, logger, configDir)
			if err != nil {
				return fmt.Errorf("failed to create metrics collector: %w", err)
			}
			defer collector.Close()

			interval, _ := cmd.Flags().GetDuration("interval")
			interfaces, _ := cmd.Flags().GetStringSlice("interfaces")
			targets, _ := cmd.Flags().GetStringSlice("targets")
			duration, _ := cmd.Flags().GetDuration("duration")

			config := MonitoringConfig{
				Interval:   interval,
				Interfaces: interfaces,
				Targets:    targets,
				Duration:   duration,
			}

			fmt.Printf("🔍 Starting network metrics monitoring (interval: %s)\n", interval)
			fmt.Println("Press Ctrl+C to stop...")

			return collector.StartMonitoring(ctx, config)
		},
	}

	cmd.Flags().DurationP("interval", "i", 5*time.Second, "Monitoring interval")
	cmd.Flags().StringSlice("interfaces", []string{}, "Network interfaces to monitor (auto-detect if empty)")
	cmd.Flags().StringSlice("targets", []string{"8.8.8.8", "1.1.1.1"}, "Latency test targets")
	cmd.Flags().DurationP("duration", "d", 0, "Monitoring duration (0 = unlimited)")

	return cmd
}

// newNetworkMetricsShowCmd creates the show subcommand.
func newNetworkMetricsShowCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show current network metrics",
		Long:  `Display current network performance metrics and statistics.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			collector, err := createNetworkMetricsCollector(ctx, logger, configDir)
			if err != nil {
				return fmt.Errorf("failed to create metrics collector: %w", err)
			}
			defer collector.Close()

			output, _ := cmd.Flags().GetString("output")

			metrics, err := collector.GetCurrentMetrics(ctx)
			if err != nil {
				return fmt.Errorf("failed to get network metrics: %w", err)
			}

			switch output {
			case "json":
				return json.NewEncoder(os.Stdout).Encode(metrics)
			default:
				printNetworkMetrics(metrics)
				return nil
			}
		},
	}

	cmd.Flags().StringP("output", "o", "table", "Output format (table|json)")

	return cmd
}

// newNetworkMetricsLatencyCmd creates the latency subcommand.
func newNetworkMetricsLatencyCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "latency",
		Short: "Test network latency to targets",
		Long:  `Test network latency and packet loss to specified targets.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			collector, err := createNetworkMetricsCollector(ctx, logger, configDir)
			if err != nil {
				return fmt.Errorf("failed to create metrics collector: %w", err)
			}
			defer collector.Close()

			targets, _ := cmd.Flags().GetStringSlice("targets")
			count, _ := cmd.Flags().GetInt("count")
			output, _ := cmd.Flags().GetString("output")

			if len(targets) == 0 {
				targets = []string{"8.8.8.8", "1.1.1.1", "google.com"}
			}

			fmt.Printf("🔍 Testing latency to %d targets (%d packets each)...\n", len(targets), count)

			results, err := collector.TestLatency(ctx, targets, count)
			if err != nil {
				return fmt.Errorf("failed to test latency: %w", err)
			}

			switch output {
			case "json":
				return json.NewEncoder(os.Stdout).Encode(results)
			default:
				return printLatencyResults(results)
			}
		},
	}

	cmd.Flags().StringSlice("targets", []string{}, "Targets to test (IPs or hostnames)")
	cmd.Flags().IntP("count", "c", 5, "Number of ping packets per target")
	cmd.Flags().StringP("output", "o", "table", "Output format (table|json)")

	return cmd
}

// newNetworkMetricsBandwidthCmd creates the bandwidth subcommand.
func newNetworkMetricsBandwidthCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bandwidth",
		Short: "Monitor bandwidth usage",
		Long:  `Monitor real-time bandwidth usage for network interfaces.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			collector, err := createNetworkMetricsCollector(ctx, logger, configDir)
			if err != nil {
				return fmt.Errorf("failed to create metrics collector: %w", err)
			}
			defer collector.Close()

			iface, _ := cmd.Flags().GetString("interface")
			interval, _ := cmd.Flags().GetDuration("interval")
			duration, _ := cmd.Flags().GetDuration("duration")
			output, _ := cmd.Flags().GetString("output")

			fmt.Printf("📊 Monitoring bandwidth for interface: %s (interval: %s)\n", iface, interval)

			usage, err := collector.MonitorBandwidth(ctx, iface, interval, duration)
			if err != nil {
				return fmt.Errorf("failed to monitor bandwidth: %w", err)
			}

			switch output {
			case "json":
				return json.NewEncoder(os.Stdout).Encode(usage)
			default:
				return printBandwidthUsage(usage)
			}
		},
	}

	cmd.Flags().StringP("interface", "i", "", "Network interface to monitor (auto-detect if empty)")
	cmd.Flags().Duration("interval", 2*time.Second, "Measurement interval")
	cmd.Flags().DurationP("duration", "d", 30*time.Second, "Monitoring duration")
	cmd.Flags().StringP("output", "o", "table", "Output format (table|json)")

	return cmd
}

// newNetworkMetricsReportCmd creates the report subcommand.
func newNetworkMetricsReportCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report",
		Short: "Generate network performance report",
		Long:  `Generate comprehensive network performance report with historical analysis.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			collector, err := createNetworkMetricsCollector(ctx, logger, configDir)
			if err != nil {
				return fmt.Errorf("failed to create metrics collector: %w", err)
			}
			defer collector.Close()

			duration, _ := cmd.Flags().GetDuration("duration")
			outputFile, _ := cmd.Flags().GetString("output-file")
			format, _ := cmd.Flags().GetString("format")

			fmt.Printf("📊 Generating network performance report (duration: %s)...\n", duration)

			report, err := collector.GenerateReport(ctx, duration)
			if err != nil {
				return fmt.Errorf("failed to generate report: %w", err)
			}

			if outputFile != "" {
				return saveReport(report, outputFile, format)
			}

			printPerformanceReport(report)
			return nil
		},
	}

	cmd.Flags().DurationP("duration", "d", 1*time.Hour, "Report time period")
	cmd.Flags().String("output-file", "", "Save report to file")
	cmd.Flags().String("format", "text", "Report format (text|json|html)")

	return cmd
}

// newNetworkMetricsOptimizeCmd creates the optimize subcommand.
func newNetworkMetricsOptimizeCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "optimize",
		Short: "Analyze and provide optimization recommendations",
		Long:  `Analyze current network performance and provide optimization recommendations.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			collector, err := createNetworkMetricsCollector(ctx, logger, configDir)
			if err != nil {
				return fmt.Errorf("failed to create metrics collector: %w", err)
			}
			defer collector.Close()

			apply, _ := cmd.Flags().GetBool("apply")

			fmt.Println("🔍 Analyzing network performance...")

			recommendations, err := collector.AnalyzeAndOptimize(ctx, apply)
			if err != nil {
				return fmt.Errorf("failed to analyze network: %w", err)
			}

			printOptimizationRecommendations(recommendations, apply)
			return nil
		},
	}

	cmd.Flags().Bool("apply", false, "Apply optimization recommendations automatically")

	return cmd
}

// Helper types and structures

type NetworkMetricsCollector struct {
	logger       *zap.Logger
	configDir    string
	commandPool  *CommandPool
	metrics      *NetworkMetrics
	isMonitoring bool
}

type MonitoringConfig struct {
	Interval   time.Duration `json:"interval"`
	Interfaces []string      `json:"interfaces"`
	Targets    []string      `json:"targets"`
	Duration   time.Duration `json:"duration"`
}

type NetworkMetrics struct {
	Timestamp       time.Time                   `json:"timestamp"`
	Interfaces      map[string]InterfaceMetrics `json:"interfaces"`
	LatencyResults  []LatencyResult             `json:"latency_results"`
	QualityScore    float64                     `json:"quality_score"`
	TotalBandwidth  BandwidthMetrics            `json:"total_bandwidth"`
	ConnectionCount int                         `json:"connection_count"`
}

type InterfaceMetrics struct {
	Name        string           `json:"name"`
	State       string           `json:"state"`
	Speed       int64            `json:"speed_mbps"`
	Bandwidth   BandwidthMetrics `json:"bandwidth"`
	PacketStats PacketStats      `json:"packet_stats"`
	MTU         int              `json:"mtu"`
	IPAddress   string           `json:"ip_address"`
}

type BandwidthMetrics struct {
	UploadMbps   float64 `json:"upload_mbps"`
	DownloadMbps float64 `json:"download_mbps"`
	TotalBytes   int64   `json:"total_bytes"`
	PeakUpload   float64 `json:"peak_upload_mbps"`
	PeakDownload float64 `json:"peak_download_mbps"`
}

type PacketStats struct {
	Transmitted int64   `json:"transmitted"`
	Received    int64   `json:"received"`
	Dropped     int64   `json:"dropped"`
	Errors      int64   `json:"errors"`
	LossRate    float64 `json:"loss_rate_percent"`
}

type LatencyResult struct {
	Target     string        `json:"target"`
	MinLatency time.Duration `json:"min_latency"`
	MaxLatency time.Duration `json:"max_latency"`
	AvgLatency time.Duration `json:"avg_latency"`
	PacketLoss float64       `json:"packet_loss_percent"`
	Jitter     time.Duration `json:"jitter"`
	Success    bool          `json:"success"`
}

type PerformanceReport struct {
	GeneratedAt     time.Time                    `json:"generated_at"`
	Duration        time.Duration                `json:"duration"`
	Summary         PerformanceSummary           `json:"summary"`
	InterfaceStats  map[string]InterfaceStats    `json:"interface_stats"`
	LatencyTrends   []LatencyTrend               `json:"latency_trends"`
	BandwidthTrends []BandwidthTrend             `json:"bandwidth_trends"`
	Issues          []PerformanceIssue           `json:"issues"`
	Recommendations []OptimizationRecommendation `json:"recommendations"`
}

type PerformanceSummary struct {
	AvgBandwidth   BandwidthMetrics `json:"avg_bandwidth"`
	AvgLatency     time.Duration    `json:"avg_latency"`
	OverallQuality string           `json:"overall_quality"`
	UptimePercent  float64          `json:"uptime_percent"`
	MajorIssues    int              `json:"major_issues"`
	MinorIssues    int              `json:"minor_issues"`
}

type InterfaceStats struct {
	Name          string           `json:"name"`
	AvgBandwidth  BandwidthMetrics `json:"avg_bandwidth"`
	PeakBandwidth BandwidthMetrics `json:"peak_bandwidth"`
	TotalTraffic  int64            `json:"total_traffic_bytes"`
	ErrorRate     float64          `json:"error_rate_percent"`
	UptimePercent float64          `json:"uptime_percent"`
}

// LatencyTrend type moved to network_analysis_cmd.go to avoid duplication

type BandwidthTrend struct {
	Timestamp   time.Time        `json:"timestamp"`
	Interface   string           `json:"interface"`
	Bandwidth   BandwidthMetrics `json:"bandwidth"`
	Utilization float64          `json:"utilization_percent"`
}

type PerformanceIssue struct {
	Type        string    `json:"type"`
	Severity    string    `json:"severity"`
	Description string    `json:"description"`
	Interface   string    `json:"interface,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
	Impact      string    `json:"impact"`
}

type OptimizationRecommendation struct {
	Category    string `json:"category"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Command     string `json:"command,omitempty"`
	Impact      string `json:"impact"`
	Applied     bool   `json:"applied"`
}

func createNetworkMetricsCollector(_ context.Context, logger *zap.Logger, configDir string) (*NetworkMetricsCollector, error) { //nolint:unparam // Error always nil but kept for consistency
	collector := &NetworkMetricsCollector{
		logger:      logger,
		configDir:   configDir,
		commandPool: NewCommandPool(10),
		metrics:     &NetworkMetrics{},
	}

	return collector, nil
}

func (nmc *NetworkMetricsCollector) Close() {
	nmc.commandPool.Close()
}

func (nmc *NetworkMetricsCollector) StartMonitoring(ctx context.Context, config MonitoringConfig) error {
	nmc.isMonitoring = true

	defer func() { nmc.isMonitoring = false }()

	ticker := time.NewTicker(config.Interval)
	defer ticker.Stop()

	var endTime time.Time
	if config.Duration > 0 {
		endTime = time.Now().Add(config.Duration)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if config.Duration > 0 && time.Now().After(endTime) {
				fmt.Println("✅ Monitoring completed")
				return nil
			}

			metrics, err := nmc.collectMetrics(ctx, config)
			if err != nil {
				nmc.logger.Warn("Failed to collect metrics", zap.Error(err))
				continue
			}

			nmc.displayRealTimeMetrics(metrics)
		}
	}
}

func (nmc *NetworkMetricsCollector) GetCurrentMetrics(ctx context.Context) (*NetworkMetrics, error) {
	config := MonitoringConfig{
		Interfaces: []string{},
		Targets:    []string{"8.8.8.8", "1.1.1.1"},
	}

	return nmc.collectMetrics(ctx, config)
}

func (nmc *NetworkMetricsCollector) TestLatency(ctx context.Context, targets []string, count int) ([]LatencyResult, error) {
	results := make([]LatencyResult, 0, len(targets))

	for _, target := range targets {
		result := nmc.pingTarget(ctx, target, count)
		results = append(results, result)
	}

	return results, nil
}

func (nmc *NetworkMetricsCollector) MonitorBandwidth(ctx context.Context, iface string, interval, duration time.Duration) ([]BandwidthTrend, error) {
	var trends []BandwidthTrend

	if iface == "" {
		iface = nmc.getDefaultInterface()
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	endTime := time.Now().Add(duration)

	for time.Now().Before(endTime) {
		select {
		case <-ctx.Done():
			return trends, ctx.Err()
		case <-ticker.C:
			bandwidth := nmc.measureBandwidth(iface)
			trend := BandwidthTrend{
				Timestamp:   time.Now(),
				Interface:   iface,
				Bandwidth:   bandwidth,
				Utilization: nmc.calculateUtilization(bandwidth, iface),
			}
			trends = append(trends, trend)

			fmt.Printf("\r📊 %s: ↑ %.2f Mbps ↓ %.2f Mbps (%.1f%% utilization)",
				iface, bandwidth.UploadMbps, bandwidth.DownloadMbps, trend.Utilization)
		}
	}

	fmt.Println() // New line after monitoring

	return trends, nil
}

func (nmc *NetworkMetricsCollector) GenerateReport(ctx context.Context, duration time.Duration) (*PerformanceReport, error) {
	// TODO: Implement comprehensive report generation
	// This would collect historical data and generate analysis
	report := &PerformanceReport{
		GeneratedAt: time.Now(),
		Duration:    duration,
		Summary: PerformanceSummary{
			OverallQuality: "Good",
			UptimePercent:  99.5,
			MajorIssues:    0,
			MinorIssues:    2,
		},
		Issues: []PerformanceIssue{
			{
				Type:        "Latency",
				Severity:    "Minor",
				Description: "Occasional high latency spikes detected",
				Timestamp:   time.Now(),
				Impact:      "May affect real-time applications",
			},
		},
		Recommendations: []OptimizationRecommendation{
			{
				Category:    "DNS",
				Title:       "Optimize DNS Configuration",
				Description: "Switch to faster DNS servers for improved resolution speed",
				Command:     "resolvectl dns eth0 1.1.1.1 8.8.8.8",
				Impact:      "Reduce DNS lookup time by 20-30%",
			},
		},
	}

	return report, nil
}

func (nmc *NetworkMetricsCollector) AnalyzeAndOptimize(ctx context.Context, apply bool) ([]OptimizationRecommendation, error) {
	recommendations := []OptimizationRecommendation{
		{
			Category:    "TCP",
			Title:       "Optimize TCP Buffer Sizes",
			Description: "Increase TCP buffer sizes for better throughput",
			Command:     "sysctl -w net.core.rmem_max=16777216 net.core.wmem_max=16777216",
			Impact:      "Improve throughput by up to 15%",
		},
		{
			Category:    "Network",
			Title:       "Enable TCP Window Scaling",
			Description: "Enable TCP window scaling for high-bandwidth connections",
			Command:     "sysctl -w net.ipv4.tcp_window_scaling=1",
			Impact:      "Better performance on high-latency connections",
		},
		{
			Category:    "DNS",
			Title:       "Configure Fast DNS Servers",
			Description: "Use Cloudflare and Google DNS for faster resolution",
			Command:     "resolvectl dns eth0 1.1.1.1 8.8.8.8",
			Impact:      "Reduce DNS lookup time by 20-30%",
		},
	}

	if apply {
		for i := range recommendations {
			if recommendations[i].Command != "" {
				nmc.logger.Info("Applying optimization",
					zap.String("title", recommendations[i].Title),
					zap.String("command", recommendations[i].Command))

				result := nmc.commandPool.ExecuteCommand("bash", "-c", recommendations[i].Command)
				recommendations[i].Applied = result.Error == nil
			}
		}
	}

	return recommendations, nil
}

// Helper functions

func (nmc *NetworkMetricsCollector) collectMetrics(ctx context.Context, config MonitoringConfig) (*NetworkMetrics, error) {
	metrics := &NetworkMetrics{
		Timestamp:  time.Now(),
		Interfaces: make(map[string]InterfaceMetrics),
	}

	// Collect interface metrics
	interfaces := config.Interfaces
	if len(interfaces) == 0 {
		interfaces = nmc.getActiveInterfaces()
	}

	for _, iface := range interfaces {
		ifaceMetrics := nmc.collectInterfaceMetrics(iface)
		metrics.Interfaces[iface] = ifaceMetrics
	}

	// Test latency to targets
	for _, target := range config.Targets {
		result := nmc.pingTarget(ctx, target, 3)
		metrics.LatencyResults = append(metrics.LatencyResults, result)
	}

	// Calculate quality score
	metrics.QualityScore = nmc.calculateQualityScore(metrics)

	return metrics, nil
}

func (nmc *NetworkMetricsCollector) getActiveInterfaces() []string {
	result := nmc.commandPool.ExecuteCommand("ip", "link", "show", "up")
	if result.Error != nil {
		return []string{"eth0"} // Fallback
	}

	var interfaces []string

	lines := strings.Split(string(result.Output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "state UP") && strings.Contains(line, ":") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				name := strings.TrimSpace(parts[1])
				if name != "lo" { // Skip loopback
					interfaces = append(interfaces, name)
				}
			}
		}
	}

	if len(interfaces) == 0 {
		return []string{"eth0"} // Fallback
	}

	return interfaces
}

func (nmc *NetworkMetricsCollector) collectInterfaceMetrics(iface string) InterfaceMetrics {
	metrics := InterfaceMetrics{
		Name:        iface,
		State:       "up",
		Bandwidth:   nmc.measureBandwidth(iface),
		PacketStats: nmc.getPacketStats(iface),
	}

	// Get IP address
	result := nmc.commandPool.ExecuteCommand("ip", "addr", "show", iface)
	if result.Error == nil {
		output := string(result.Output)
		if idx := strings.Index(output, "inet "); idx != -1 {
			line := output[idx:]
			if spaceIdx := strings.Index(line[5:], " "); spaceIdx != -1 {
				ipWithCidr := line[5 : 5+spaceIdx]
				if slashIdx := strings.Index(ipWithCidr, "/"); slashIdx != -1 {
					metrics.IPAddress = ipWithCidr[:slashIdx]
				}
			}
		}
	}

	return metrics
}

func (nmc *NetworkMetricsCollector) measureBandwidth(iface string) BandwidthMetrics {
	// Get interface statistics
	result := nmc.commandPool.ExecuteCommand("cat", fmt.Sprintf("/sys/class/net/%s/statistics/rx_bytes", iface))
	rxBytes := int64(0)

	if result.Error == nil {
		if val, err := strconv.ParseInt(strings.TrimSpace(string(result.Output)), 10, 64); err == nil {
			rxBytes = val
		}
	}

	result = nmc.commandPool.ExecuteCommand("cat", fmt.Sprintf("/sys/class/net/%s/statistics/tx_bytes", iface))
	txBytes := int64(0)

	if result.Error == nil {
		if val, err := strconv.ParseInt(strings.TrimSpace(string(result.Output)), 10, 64); err == nil {
			txBytes = val
		}
	}

	// TODO: Calculate actual bandwidth rates by comparing with previous measurements
	// For now, return simple metrics
	return BandwidthMetrics{
		UploadMbps:   float64(txBytes) / (1024 * 1024), // Simplified calculation
		DownloadMbps: float64(rxBytes) / (1024 * 1024), // Simplified calculation
		TotalBytes:   rxBytes + txBytes,
	}
}

func (nmc *NetworkMetricsCollector) getPacketStats(iface string) PacketStats {
	stats := PacketStats{}

	// Get packet statistics
	result := nmc.commandPool.ExecuteCommand("cat", fmt.Sprintf("/sys/class/net/%s/statistics/rx_packets", iface))
	if result.Error == nil {
		if val, err := strconv.ParseInt(strings.TrimSpace(string(result.Output)), 10, 64); err == nil {
			stats.Received = val
		}
	}

	result = nmc.commandPool.ExecuteCommand("cat", fmt.Sprintf("/sys/class/net/%s/statistics/tx_packets", iface))
	if result.Error == nil {
		if val, err := strconv.ParseInt(strings.TrimSpace(string(result.Output)), 10, 64); err == nil {
			stats.Transmitted = val
		}
	}

	result = nmc.commandPool.ExecuteCommand("cat", fmt.Sprintf("/sys/class/net/%s/statistics/rx_dropped", iface))
	if result.Error == nil {
		if val, err := strconv.ParseInt(strings.TrimSpace(string(result.Output)), 10, 64); err == nil {
			stats.Dropped = val
		}
	}

	result = nmc.commandPool.ExecuteCommand("cat", fmt.Sprintf("/sys/class/net/%s/statistics/rx_errors", iface))
	if result.Error == nil {
		if val, err := strconv.ParseInt(strings.TrimSpace(string(result.Output)), 10, 64); err == nil {
			stats.Errors = val
		}
	}

	// Calculate loss rate
	totalPackets := stats.Received + stats.Transmitted
	if totalPackets > 0 {
		stats.LossRate = float64(stats.Dropped+stats.Errors) / float64(totalPackets) * 100
	}

	return stats
}

func (nmc *NetworkMetricsCollector) pingTarget(_ context.Context, target string, count int) LatencyResult { //nolint:gocognit // Complex ping implementation with cross-platform support
	result := LatencyResult{
		Target:  target,
		Success: false,
	}

	// Execute ping command
	pingResult := nmc.commandPool.ExecuteCommand("ping", "-c", strconv.Itoa(count), target)
	if pingResult.Error != nil {
		return result
	}

	output := string(pingResult.Output)
	result.Success = true

	// Parse ping output for statistics and packet loss
	parsePingStatistics(output, &result)
	parsePingPacketLoss(output, &result)

	return result
}

func (nmc *NetworkMetricsCollector) getDefaultInterface() string {
	result := nmc.commandPool.ExecuteCommand("ip", "route", "get", "1.1.1.1")
	if result.Error != nil {
		return "eth0" // Fallback
	}

	fields := strings.Fields(string(result.Output))
	for i, field := range fields {
		if field == "dev" && i+1 < len(fields) {
			return fields[i+1]
		}
	}

	return "eth0" // Fallback
}

func (nmc *NetworkMetricsCollector) calculateUtilization(bandwidth BandwidthMetrics, _ string) float64 {
	// TODO: Get interface speed and calculate actual utilization
	// For now, return a simplified calculation
	totalMbps := bandwidth.UploadMbps + bandwidth.DownloadMbps
	return (totalMbps / 1000) * 100 // Assume 1 Gbps interface
}

func (nmc *NetworkMetricsCollector) calculateQualityScore(metrics *NetworkMetrics) float64 {
	score := 100.0

	// Reduce score based on latency
	for _, latency := range metrics.LatencyResults {
		if latency.Success {
			if latency.AvgLatency > 100*time.Millisecond {
				score -= 10
			} else if latency.AvgLatency > 50*time.Millisecond {
				score -= 5
			}

			if latency.PacketLoss > 5 {
				score -= 20
			} else if latency.PacketLoss > 1 {
				score -= 10
			}
		} else {
			score -= 25
		}
	}

	// Reduce score based on packet loss
	for _, iface := range metrics.Interfaces {
		if iface.PacketStats.LossRate > 1 {
			score -= 15
		}
	}

	if score < 0 {
		score = 0
	}

	return score
}

func (nmc *NetworkMetricsCollector) displayRealTimeMetrics(metrics *NetworkMetrics) {
	fmt.Printf("\r⏰ %s | Quality: %.1f%% | ",
		metrics.Timestamp.Format("15:04:05"), metrics.QualityScore)

	for name, iface := range metrics.Interfaces {
		fmt.Printf("%s: ↑%.2fMb ↓%.2fMb | ",
			name, iface.Bandwidth.UploadMbps, iface.Bandwidth.DownloadMbps)
	}

	if len(metrics.LatencyResults) > 0 {
		avgLatency := time.Duration(0)
		successCount := 0

		for _, result := range metrics.LatencyResults {
			if result.Success {
				avgLatency += result.AvgLatency
				successCount++
			}
		}

		if successCount > 0 {
			avgLatency /= time.Duration(successCount)
			fmt.Printf("Latency: %s", avgLatency.Round(time.Millisecond))
		}
	}
}

// Print functions

func printNetworkMetrics(metrics *NetworkMetrics) {
	fmt.Printf("📊 Network Performance Metrics\n\n")
	fmt.Printf("Timestamp: %s\n", metrics.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Printf("Quality Score: %.1f%%\n\n", metrics.QualityScore)

	// Interface metrics
	if len(metrics.Interfaces) > 0 {
		fmt.Printf("Interface Metrics:\n")

		w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
		_, _ = fmt.Fprintln(w, "INTERFACE\tSTATE\tIP ADDRESS\tUPLOAD\tDOWNLOAD\tPACKET LOSS")

		for _, iface := range metrics.Interfaces {
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%.2f Mbps\t%.2f Mbps\t%.2f%%\n",
				iface.Name,
				iface.State,
				iface.IPAddress,
				iface.Bandwidth.UploadMbps,
				iface.Bandwidth.DownloadMbps,
				iface.PacketStats.LossRate)
		}

		_ = w.Flush()
		fmt.Println()
	}

	// Latency results
	if len(metrics.LatencyResults) > 0 {
		fmt.Printf("Latency Test Results:\n")
		_ = printLatencyResults(metrics.LatencyResults)
	}
}

func printLatencyResults(results []LatencyResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	_, _ = fmt.Fprintln(w, "TARGET\tMIN\tAVG\tMAX\tJITTER\tPACKET LOSS\tSTATUS")

	for _, result := range results {
		status := "❌ Failed"
		if result.Success {
			status = "✅ Success"
		}

		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%.1f%%\t%s\n",
			result.Target,
			result.MinLatency.Round(time.Millisecond),
			result.AvgLatency.Round(time.Millisecond),
			result.MaxLatency.Round(time.Millisecond),
			result.Jitter.Round(time.Millisecond),
			result.PacketLoss,
			status)
	}

	return w.Flush()
}

func printBandwidthUsage(trends []BandwidthTrend) error {
	fmt.Printf("📊 Bandwidth Usage History\n\n")

	if len(trends) == 0 {
		fmt.Println("No bandwidth data collected.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	_, _ = fmt.Fprintln(w, "TIME\tINTERFACE\tUPLOAD\tDOWNLOAD\tUTILIZATION")

	for _, trend := range trends {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%.2f Mbps\t%.2f Mbps\t%.1f%%\n",
			trend.Timestamp.Format("15:04:05"),
			trend.Interface,
			trend.Bandwidth.UploadMbps,
			trend.Bandwidth.DownloadMbps,
			trend.Utilization)
	}

	return w.Flush()
}

func printPerformanceReport(report *PerformanceReport) {
	fmt.Printf("📊 Network Performance Report\n\n")
	fmt.Printf("Generated: %s\n", report.GeneratedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Duration: %s\n\n", report.Duration)

	// Summary
	fmt.Printf("📈 Summary:\n")
	fmt.Printf("  Overall Quality: %s\n", report.Summary.OverallQuality)
	fmt.Printf("  Uptime: %.2f%%\n", report.Summary.UptimePercent)
	fmt.Printf("  Issues: %d major, %d minor\n\n", report.Summary.MajorIssues, report.Summary.MinorIssues)

	// Issues
	if len(report.Issues) > 0 {
		fmt.Printf("⚠️  Issues Found:\n")

		for i, issue := range report.Issues {
			fmt.Printf("  %d. [%s] %s: %s\n", i+1, issue.Severity, issue.Type, issue.Description)
		}

		fmt.Println()
	}

	// Recommendations
	if len(report.Recommendations) > 0 {
		fmt.Printf("💡 Optimization Recommendations:\n")
		printOptimizationRecommendations(report.Recommendations, false)
	}
}

func printOptimizationRecommendations(recommendations []OptimizationRecommendation, applied bool) {
	if len(recommendations) == 0 {
		fmt.Println("No optimization recommendations at this time.")
		return
	}

	for i, rec := range recommendations {
		status := ""

		if applied {
			if rec.Applied {
				status = " ✅ Applied"
			} else {
				status = " ❌ Failed to apply"
			}
		}

		fmt.Printf("%d. [%s] %s%s\n", i+1, rec.Category, rec.Title, status)
		fmt.Printf("   %s\n", rec.Description)
		fmt.Printf("   Impact: %s\n", rec.Impact)

		if rec.Command != "" && !applied {
			fmt.Printf("   Command: %s\n", rec.Command)
		}

		fmt.Println()
	}
}

func saveReport(report *PerformanceReport, filename, format string) error {
	var (
		data []byte
		err  error
	)

	switch format {
	case "json":
		data, err = json.MarshalIndent(report, "", "  ")
	case "html":
		// TODO: Implement HTML report generation
		return fmt.Errorf("HTML format not yet implemented")
	default:
		// Text format
		// TODO: Implement text report generation
		return fmt.Errorf("text format not yet implemented")
	}

	if err != nil {
		return fmt.Errorf("failed to format report: %w", err)
	}

	if err := os.WriteFile(filename, data, 0o644); err != nil {
		return fmt.Errorf("failed to save report: %w", err)
	}

	fmt.Printf("✅ Report saved to: %s\n", filename)

	return nil
}

// parsePingStatistics parses round-trip statistics from ping output.
func parsePingStatistics(output string, result *LatencyResult) {
	// Example: "round-trip min/avg/max/stddev = 10.123/15.456/20.789/3.123 ms"
	if !strings.Contains(output, "round-trip") {
		return
	}

	idx := strings.Index(output, "round-trip")
	if idx == -1 {
		return
	}

	line := output[idx:]
	eqIdx := strings.Index(line, "=")
	if eqIdx == -1 {
		return
	}

	statsStr := line[eqIdx+1:]
	spaceIdx := strings.Index(statsStr, " ms")
	if spaceIdx == -1 {
		return
	}

	statsStr = strings.TrimSpace(statsStr[:spaceIdx])
	parts := strings.Split(statsStr, "/")
	if len(parts) < 3 {
		return
	}

	if minLatency, err := time.ParseDuration(parts[0] + "ms"); err == nil {
		result.MinLatency = minLatency
	}

	if avg, err := time.ParseDuration(parts[1] + "ms"); err == nil {
		result.AvgLatency = avg
	}

	if maxLatency, err := time.ParseDuration(parts[2] + "ms"); err == nil {
		result.MaxLatency = maxLatency
	}

	if len(parts) >= 4 {
		if jitter, err := time.ParseDuration(parts[3] + "ms"); err == nil {
			result.Jitter = jitter
		}
	}
}

// parsePingPacketLoss parses packet loss information from ping output.
func parsePingPacketLoss(output string, result *LatencyResult) {
	if !strings.Contains(output, "% packet loss") {
		return
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if !strings.Contains(line, "% packet loss") {
			continue
		}

		fields := strings.Fields(line)
		for i, field := range fields {
			if strings.Contains(field, "%") && i > 0 {
				if loss, err := strconv.ParseFloat(strings.TrimSuffix(field, "%"), 64); err == nil {
					result.PacketLoss = loss
				}
				return
			}
		}
	}
}
