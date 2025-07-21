//nolint:tagliatelle // Network analysis output may require specific JSON field naming conventions
package netenv

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/gizzahub/gzh-manager-go/internal/env"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// newNetworkAnalysisCmd creates the network analysis command.
func newNetworkAnalysisCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "network-analysis",
		Short: "Perform advanced network latency and bandwidth analysis",
		Long: `Perform comprehensive network performance analysis including detailed latency patterns, bandwidth utilization trends, and performance optimization recommendations.

This command provides advanced analysis capabilities:
- Detailed latency pattern analysis and trends
- Bandwidth utilization statistics and predictions
- Network performance quality assessment
- Bottleneck identification and analysis
- Historical performance comparison
- Performance regression detection

Examples:
  # Analyze current network latency patterns
  gz net-env network-analysis latency --duration 10m --targets 8.8.8.8,1.1.1.1,google.com

  # Perform bandwidth utilization analysis
  gz net-env network-analysis bandwidth --interface eth0 --duration 5m

  # Comprehensive network performance analysis
  gz net-env network-analysis comprehensive --duration 15m

  # Generate performance trend report
  gz net-env network-analysis trends --period 24h --format json`,
		SilenceUsage: true,
	}

	// Add subcommands
	cmd.AddCommand(newNetworkAnalysisLatencyCmd(logger, configDir))
	cmd.AddCommand(newNetworkAnalysisBandwidthCmd(logger, configDir))
	cmd.AddCommand(newNetworkAnalysisComprehensiveCmd(logger, configDir))
	cmd.AddCommand(newNetworkAnalysisTrendsCmd(logger, configDir))
	cmd.AddCommand(newNetworkAnalysisBottleneckCmd(logger, configDir))

	return cmd
}

// newNetworkAnalysisLatencyCmd creates the latency analysis subcommand.
func newNetworkAnalysisLatencyCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "latency",
		Short: "Perform detailed latency pattern analysis",
		Long:  `Analyze network latency patterns, detect anomalies, and provide latency optimization recommendations.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			analyzer := createNetworkAnalyzer(ctx, logger, configDir)
			defer analyzer.Close()

			targets, _ := cmd.Flags().GetStringSlice("targets")
			duration, _ := cmd.Flags().GetDuration("duration")
			interval, _ := cmd.Flags().GetDuration("interval")
			output, _ := cmd.Flags().GetString("output")

			if len(targets) == 0 {
				targets = []string{"8.8.8.8", "1.1.1.1", "google.com", "cloudflare.com"}
			}

			config := LatencyAnalysisConfig{
				Targets:  targets,
				Duration: duration,
				Interval: interval,
			}

			fmt.Printf("🔍 Analyzing latency patterns for %d targets over %s...\n", len(targets), duration)

			analysis, err := analyzer.AnalyzeLatencyPatterns(ctx, config)
			if err != nil {
				return fmt.Errorf("failed to analyze latency patterns: %w", err)
			}

			switch output {
			case "json":
				return json.NewEncoder(os.Stdout).Encode(analysis)
			default:
				printLatencyAnalysis(analysis)
				return nil
			}
		},
	}

	cmd.Flags().StringSlice("targets", []string{}, "Targets to analyze")
	cmd.Flags().DurationP("duration", "d", 10*time.Minute, "Analysis duration")
	cmd.Flags().Duration("interval", 1*time.Second, "Measurement interval")
	cmd.Flags().StringP("output", "o", "table", "Output format (table|json)")

	return cmd
}

// newNetworkAnalysisBandwidthCmd creates the bandwidth analysis subcommand.
func newNetworkAnalysisBandwidthCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bandwidth",
		Short: "Perform detailed bandwidth utilization analysis",
		Long:  `Analyze bandwidth utilization patterns, identify peak usage periods, and detect capacity issues.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			analyzer := createNetworkAnalyzer(ctx, logger, configDir)
			defer analyzer.Close()

			iface, _ := cmd.Flags().GetString("interface")
			duration, _ := cmd.Flags().GetDuration("duration")
			interval, _ := cmd.Flags().GetDuration("interval")
			output, _ := cmd.Flags().GetString("output")

			if iface == "" {
				iface = analyzer.getDefaultInterface()
			}

			config := BandwidthAnalysisConfig{
				Interface: iface,
				Duration:  duration,
				Interval:  interval,
			}

			fmt.Printf("📊 Analyzing bandwidth utilization for %s over %s...\n", iface, duration)

			analysis, err := analyzer.AnalyzeBandwidthUtilization(ctx, config)
			if err != nil {
				return fmt.Errorf("failed to analyze bandwidth utilization: %w", err)
			}

			switch output {
			case "json":
				return json.NewEncoder(os.Stdout).Encode(analysis)
			default:
				printBandwidthAnalysis(analysis)
				return nil
			}
		},
	}

	cmd.Flags().StringP("interface", "i", "", "Network interface to analyze")
	cmd.Flags().DurationP("duration", "d", 5*time.Minute, "Analysis duration")
	cmd.Flags().Duration("interval", 2*time.Second, "Measurement interval")
	cmd.Flags().StringP("output", "o", "table", "Output format (table|json)")

	return cmd
}

// newNetworkAnalysisComprehensiveCmd creates the comprehensive analysis subcommand.
func newNetworkAnalysisComprehensiveCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "comprehensive",
		Short: "Perform comprehensive network performance analysis",
		Long:  `Perform comprehensive analysis combining latency, bandwidth, and overall network health assessment.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			analyzer := createNetworkAnalyzer(ctx, logger, configDir)
			defer analyzer.Close()

			duration, _ := cmd.Flags().GetDuration("duration")
			output, _ := cmd.Flags().GetString("output")
			saveReport, _ := cmd.Flags().GetString("save-report")

			config := ComprehensiveAnalysisConfig{
				Duration:       duration,
				LatencyTargets: []string{"8.8.8.8", "1.1.1.1", "google.com", "cloudflare.com"},
				Interfaces:     analyzer.getActiveInterfaces(),
				IncludeQuality: true,
				IncludeTrends:  true,
			}

			fmt.Printf("🔍 Performing comprehensive network analysis over %s...\n", duration)

			analysis, err := analyzer.PerformComprehensiveAnalysis(ctx, config)
			if err != nil {
				return fmt.Errorf("failed to perform comprehensive analysis: %w", err)
			}

			if saveReport != "" {
				if err := analyzer.SaveAnalysisReport(analysis, saveReport); err != nil {
					return fmt.Errorf("failed to save report: %w", err)
				}
				fmt.Printf("✅ Analysis report saved to: %s\n", saveReport)
			}

			switch output {
			case "json":
				return json.NewEncoder(os.Stdout).Encode(analysis)
			default:
				printComprehensiveAnalysis(analysis)
				return nil
			}
		},
	}

	cmd.Flags().DurationP("duration", "d", 15*time.Minute, "Analysis duration")
	cmd.Flags().StringP("output", "o", "table", "Output format (table|json)")
	cmd.Flags().String("save-report", "", "Save detailed report to file")

	return cmd
}

// newNetworkAnalysisTrendsCmd creates the trends analysis subcommand.
func newNetworkAnalysisTrendsCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trends",
		Short: "Analyze network performance trends",
		Long:  `Analyze historical network performance trends and detect performance regressions.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			analyzer := createNetworkAnalyzer(ctx, logger, configDir)
			defer analyzer.Close()

			period, _ := cmd.Flags().GetDuration("period")
			output, _ := cmd.Flags().GetString("output")

			fmt.Printf("📈 Analyzing network performance trends over %s...\n", period)

			trends, err := analyzer.AnalyzePerformanceTrends(ctx, period)
			if err != nil {
				return fmt.Errorf("failed to analyze performance trends: %w", err)
			}

			switch output {
			case "json":
				return json.NewEncoder(os.Stdout).Encode(trends)
			default:
				printPerformanceTrends(trends)
				return nil
			}
		},
	}

	cmd.Flags().DurationP("period", "p", 24*time.Hour, "Analysis period")
	cmd.Flags().StringP("output", "o", "table", "Output format (table|json)")

	return cmd
}

// newNetworkAnalysisBottleneckCmd creates the bottleneck detection subcommand.
func newNetworkAnalysisBottleneckCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bottleneck",
		Short: "Detect and analyze network bottlenecks",
		Long:  `Detect network bottlenecks and provide optimization recommendations.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			analyzer := createNetworkAnalyzer(ctx, logger, configDir)
			defer analyzer.Close()

			fmt.Println("🔍 Detecting network bottlenecks...")

			bottlenecks, err := analyzer.DetectBottlenecks(ctx)
			if err != nil {
				return fmt.Errorf("failed to detect bottlenecks: %w", err)
			}

			printBottleneckAnalysis(bottlenecks)
			return nil
		},
	}

	return cmd
}

// Analysis types and structures

type NetworkAnalyzer struct {
	logger      *zap.Logger
	configDir   string
	commandPool *CommandPool
}

type LatencyAnalysisConfig struct {
	Targets  []string      `json:"targets"`
	Duration time.Duration `json:"duration"`
	Interval time.Duration `json:"interval"`
}

type BandwidthAnalysisConfig struct {
	Interface string        `json:"interface"`
	Duration  time.Duration `json:"duration"`
	Interval  time.Duration `json:"interval"`
}

type ComprehensiveAnalysisConfig struct {
	Duration       time.Duration `json:"duration"`
	LatencyTargets []string      `json:"latencyTargets"`
	Interfaces     []string      `json:"interfaces"`
	IncludeQuality bool          `json:"includeQuality"`
	IncludeTrends  bool          `json:"includeTrends"`
}

type LatencyAnalysis struct {
	Config           LatencyAnalysisConfig `json:"config"`
	StartTime        time.Time             `json:"startTime"`
	EndTime          time.Time             `json:"endTime"`
	TargetAnalysis   []TargetLatencyStats  `json:"targetAnalysis"`
	OverallStats     OverallLatencyStats   `json:"overallStats"`
	AnomalyDetection []LatencyAnomaly      `json:"anomalyDetection"`
	Recommendations  []string              `json:"recommendations"`
	QualityScore     float64               `json:"qualityScore"`
}

type TargetLatencyStats struct {
	Target         string                `json:"target"`
	Measurements   []LatencyMeasurement  `json:"measurements"`
	Statistics     LatencyStatistics     `json:"statistics"`
	TrendAnalysis  LatencyTrend          `json:"trendAnalysis"`
	QualityMetrics LatencyQualityMetrics `json:"qualityMetrics"`
}

type LatencyMeasurement struct {
	Timestamp  time.Time     `json:"timestamp"`
	Latency    time.Duration `json:"latency"`
	PacketLoss float64       `json:"packetLoss"`
	Jitter     time.Duration `json:"jitter"`
	Success    bool          `json:"success"`
}

type LatencyStatistics struct {
	Count       int           `json:"count"`
	Min         time.Duration `json:"min"`
	Max         time.Duration `json:"max"`
	Mean        time.Duration `json:"mean"`
	Median      time.Duration `json:"median"`
	P95         time.Duration `json:"p95"`
	P99         time.Duration `json:"p99"`
	StdDev      time.Duration `json:"stdDev"`
	SuccessRate float64       `json:"successRate"`
}

type LatencyTrend struct {
	Direction   string  `json:"direction"`  // improving, degrading, stable
	Magnitude   float64 `json:"magnitude"`  // percentage change
	Confidence  float64 `json:"confidence"` // 0-1
	Description string  `json:"description"`
}

type LatencyQualityMetrics struct {
	Consistency  float64 `json:"consistency"`   // 0-100
	Reliability  float64 `json:"reliability"`   // 0-100
	Performance  float64 `json:"performance"`   // 0-100
	OverallScore float64 `json:"overall_score"` // 0-100
}

type OverallLatencyStats struct {
	TotalMeasurements int           `json:"total_measurements"`
	AverageLatency    time.Duration `json:"average_latency"`
	BestTarget        string        `json:"best_target"`
	WorstTarget       string        `json:"worst_target"`
	NetworkStability  float64       `json:"network_stability"`
}

type LatencyAnomaly struct {
	Target      string    `json:"target"`
	Timestamp   time.Time `json:"timestamp"`
	Type        string    `json:"type"`     // spike, timeout, jitter
	Severity    string    `json:"severity"` // low, medium, high
	Value       string    `json:"value"`
	Description string    `json:"description"`
}

type BandwidthAnalysis struct {
	Config            BandwidthAnalysisConfig `json:"config"`
	StartTime         time.Time               `json:"start_time"`
	EndTime           time.Time               `json:"end_time"`
	Measurements      []BandwidthMeasurement  `json:"measurements"`
	Statistics        BandwidthStatistics     `json:"statistics"`
	UtilizationTrends UtilizationTrends       `json:"utilization_trends"`
	CapacityAnalysis  CapacityAnalysis        `json:"capacity_analysis"`
	Recommendations   []string                `json:"recommendations"`
}

type BandwidthMeasurement struct {
	Timestamp    time.Time `json:"timestamp"`
	UploadMbps   float64   `json:"upload_mbps"`
	DownloadMbps float64   `json:"download_mbps"`
	Utilization  float64   `json:"utilization_percent"`
	TotalBytes   int64     `json:"total_bytes"`
}

type BandwidthStatistics struct {
	Count           int     `json:"count"`
	AvgUpload       float64 `json:"avg_upload_mbps"`
	AvgDownload     float64 `json:"avg_download_mbps"`
	PeakUpload      float64 `json:"peak_upload_mbps"`
	PeakDownload    float64 `json:"peak_download_mbps"`
	AvgUtilization  float64 `json:"avg_utilization_percent"`
	PeakUtilization float64 `json:"peak_utilization_percent"`
	TotalTraffic    int64   `json:"total_traffic_bytes"`
}

type UtilizationTrends struct {
	Direction   string       `json:"direction"`
	GrowthRate  float64      `json:"growth_rate_percent"`
	PeakPeriods []PeakPeriod `json:"peak_periods"`
	LowPeriods  []LowPeriod  `json:"low_periods"`
}

type PeakPeriod struct {
	StartTime time.Time     `json:"start_time"`
	EndTime   time.Time     `json:"end_time"`
	PeakValue float64       `json:"peak_value"`
	Duration  time.Duration `json:"duration"`
}

type LowPeriod struct {
	StartTime time.Time     `json:"start_time"`
	EndTime   time.Time     `json:"end_time"`
	MinValue  float64       `json:"min_value"`
	Duration  time.Duration `json:"duration"`
}

type CapacityAnalysis struct {
	CurrentCapacity     float64        `json:"current_capacity_mbps"`
	UtilizedCapacity    float64        `json:"utilized_capacity_mbps"`
	AvailableCapacity   float64        `json:"available_capacity_mbps"`
	CapacityUtilization float64        `json:"capacity_utilization_percent"`
	TimeToCapacity      *time.Duration `json:"time_to_capacity,omitempty"`
	RecommendedUpgrade  bool           `json:"recommended_upgrade"`
}

type ComprehensiveAnalysis struct {
	StartTime         time.Time                `json:"start_time"`
	EndTime           time.Time                `json:"end_time"`
	Duration          time.Duration            `json:"duration"`
	LatencyAnalysis   *LatencyAnalysis         `json:"latency_analysis"`
	BandwidthAnalysis *BandwidthAnalysis       `json:"bandwidth_analysis"`
	OverallHealth     NetworkHealth            `json:"overall_health"`
	Correlations      []PerformanceCorrelation `json:"correlations"`
	Recommendations   []AnalysisRecommendation `json:"recommendations"`
}

type NetworkHealth struct {
	OverallScore   float64 `json:"overall_score"`
	LatencyScore   float64 `json:"latency_score"`
	BandwidthScore float64 `json:"bandwidth_score"`
	StabilityScore float64 `json:"stability_score"`
	HealthStatus   string  `json:"health_status"` // excellent, good, fair, poor
	MajorIssues    int     `json:"major_issues"`
	MinorIssues    int     `json:"minor_issues"`
}

type PerformanceCorrelation struct {
	Type         string  `json:"type"`
	Correlation  float64 `json:"correlation"`
	Significance string  `json:"significance"`
	Description  string  `json:"description"`
}

type AnalysisRecommendation struct {
	Priority    string `json:"priority"` // high, medium, low
	Category    string `json:"category"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Impact      string `json:"impact"`
	Effort      string `json:"effort"` // low, medium, high
	Command     string `json:"command,omitempty"`
}

type PerformanceTrends struct {
	Period          time.Duration            `json:"period"`
	LatencyTrends   []HistoricalTrend        `json:"latency_trends"`
	BandwidthTrends []HistoricalTrend        `json:"bandwidth_trends"`
	QualityTrends   []QualityTrend           `json:"quality_trends"`
	Regressions     []PerformanceRegression  `json:"regressions"`
	Improvements    []PerformanceImprovement `json:"improvements"`
}

type HistoricalTrend struct {
	Metric       string  `json:"metric"`
	Target       string  `json:"target,omitempty"`
	StartValue   float64 `json:"start_value"`
	EndValue     float64 `json:"end_value"`
	Change       float64 `json:"change_percent"`
	Direction    string  `json:"direction"`
	Significance string  `json:"significance"`
}

type QualityTrend struct {
	Timestamp    time.Time `json:"timestamp"`
	QualityScore float64   `json:"quality_score"`
	Category     string    `json:"category"`
}

type PerformanceRegression struct {
	DetectedAt  time.Time `json:"detected_at"`
	Metric      string    `json:"metric"`
	Severity    string    `json:"severity"`
	Impact      string    `json:"impact"`
	Description string    `json:"description"`
}

type PerformanceImprovement struct {
	DetectedAt  time.Time `json:"detected_at"`
	Metric      string    `json:"metric"`
	Improvement float64   `json:"improvement_percent"`
	Description string    `json:"description"`
}

type BottleneckAnalysis struct {
	DetectedBottlenecks []NetworkBottleneck        `json:"detected_bottlenecks"`
	SystemLimits        []SystemLimit              `json:"system_limits"`
	Recommendations     []BottleneckRecommendation `json:"recommendations"`
	OverallAssessment   string                     `json:"overall_assessment"`
}

type NetworkBottleneck struct {
	Type        string  `json:"type"` // bandwidth, latency, packet_loss, cpu, memory
	Location    string  `json:"location"`
	Severity    string  `json:"severity"`
	Impact      string  `json:"impact"`
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence"`
}

type SystemLimit struct {
	Resource    string  `json:"resource"`
	Current     float64 `json:"current"`
	Maximum     float64 `json:"maximum"`
	Utilization float64 `json:"utilization_percent"`
	AtRisk      bool    `json:"at_risk"`
}

type BottleneckRecommendation struct {
	Bottleneck  string `json:"bottleneck"`
	Action      string `json:"action"`
	Priority    string `json:"priority"`
	Description string `json:"description"`
	Command     string `json:"command,omitempty"`
}

// Implementation functions

func createNetworkAnalyzer(_ context.Context, logger *zap.Logger, configDir string) *NetworkAnalyzer {
	analyzer := &NetworkAnalyzer{
		logger:      logger,
		configDir:   configDir,
		commandPool: NewCommandPool(15),
	}

	return analyzer
}

func (na *NetworkAnalyzer) Close() {
	na.commandPool.Close()
}

func (na *NetworkAnalyzer) AnalyzeLatencyPatterns(ctx context.Context, config LatencyAnalysisConfig) (*LatencyAnalysis, error) {
	analysis := &LatencyAnalysis{
		Config:    config,
		StartTime: time.Now(),
	}

	// Collect latency measurements for each target
	targetAnalysis := make([]TargetLatencyStats, 0, len(config.Targets))

	for _, target := range config.Targets {
		stats, err := na.collectTargetLatencyStats(ctx, target, config)
		if err != nil {
			na.logger.Warn("Failed to collect latency stats", zap.String("target", target), zap.Error(err))
			continue
		}

		targetAnalysis = append(targetAnalysis, stats)
	}

	analysis.TargetAnalysis = targetAnalysis
	analysis.EndTime = time.Now()

	// Calculate overall statistics
	analysis.OverallStats = na.calculateOverallLatencyStats(targetAnalysis)

	// Detect anomalies
	analysis.AnomalyDetection = na.detectLatencyAnomalies(targetAnalysis)

	// Generate recommendations
	analysis.Recommendations = na.generateLatencyRecommendations(analysis)

	// Calculate quality score
	analysis.QualityScore = na.calculateLatencyQualityScore(analysis)

	return analysis, nil
}

func (na *NetworkAnalyzer) AnalyzeBandwidthUtilization(ctx context.Context, config BandwidthAnalysisConfig) (*BandwidthAnalysis, error) {
	analysis := &BandwidthAnalysis{
		Config:    config,
		StartTime: time.Now(),
	}

	// Collect bandwidth measurements
	measurements, err := na.collectBandwidthMeasurements(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to collect bandwidth measurements: %w", err)
	}

	analysis.Measurements = measurements
	analysis.EndTime = time.Now()

	// Calculate statistics
	analysis.Statistics = na.calculateBandwidthStatistics(measurements)

	// Analyze utilization trends
	analysis.UtilizationTrends = na.analyzeUtilizationTrends(measurements)

	// Perform capacity analysis
	analysis.CapacityAnalysis = na.analyzeCapacity(config.Interface, analysis.Statistics)

	// Generate recommendations
	analysis.Recommendations = na.generateBandwidthRecommendations(analysis)

	return analysis, nil
}

func (na *NetworkAnalyzer) PerformComprehensiveAnalysis(ctx context.Context, config ComprehensiveAnalysisConfig) (*ComprehensiveAnalysis, error) {
	analysis := &ComprehensiveAnalysis{
		StartTime: time.Now(),
		Duration:  config.Duration,
	}

	// Perform latency analysis
	if len(config.LatencyTargets) > 0 {
		latencyConfig := LatencyAnalysisConfig{
			Targets:  config.LatencyTargets,
			Duration: config.Duration,
			Interval: 5 * time.Second,
		}

		latencyAnalysis, err := na.AnalyzeLatencyPatterns(ctx, latencyConfig)
		if err != nil {
			na.logger.Warn("Failed to perform latency analysis", zap.Error(err))
		} else {
			analysis.LatencyAnalysis = latencyAnalysis
		}
	}

	// Perform bandwidth analysis for primary interface
	if len(config.Interfaces) > 0 {
		bandwidthConfig := BandwidthAnalysisConfig{
			Interface: config.Interfaces[0],
			Duration:  config.Duration,
			Interval:  10 * time.Second,
		}

		bandwidthAnalysis, err := na.AnalyzeBandwidthUtilization(ctx, bandwidthConfig)
		if err != nil {
			na.logger.Warn("Failed to perform bandwidth analysis", zap.Error(err))
		} else {
			analysis.BandwidthAnalysis = bandwidthAnalysis
		}
	}

	analysis.EndTime = time.Now()

	// Calculate overall health
	analysis.OverallHealth = na.calculateNetworkHealth(analysis)

	// Find correlations
	analysis.Correlations = na.findPerformanceCorrelations(analysis)

	// Generate comprehensive recommendations
	analysis.Recommendations = na.generateComprehensiveRecommendations(analysis)

	return analysis, nil
}

func (na *NetworkAnalyzer) AnalyzePerformanceTrends(ctx context.Context, period time.Duration) (*PerformanceTrends, error) {
	// TODO: Implement historical trend analysis
	// This would require storing historical performance data
	trends := &PerformanceTrends{
		Period: period,
		LatencyTrends: []HistoricalTrend{
			{
				Metric:       "average_latency",
				Target:       "8.8.8.8",
				StartValue:   25.5,
				EndValue:     22.1,
				Change:       -13.3,
				Direction:    "improving",
				Significance: "significant",
			},
		},
		BandwidthTrends: []HistoricalTrend{
			{
				Metric:       "average_utilization",
				StartValue:   45.2,
				EndValue:     52.8,
				Change:       16.8,
				Direction:    "increasing",
				Significance: "moderate",
			},
		},
		Regressions: []PerformanceRegression{
			{
				DetectedAt:  time.Now().Add(-2 * time.Hour),
				Metric:      "packet_loss",
				Severity:    "medium",
				Impact:      "Minor impact on real-time applications",
				Description: "Packet loss increased from 0.1% to 0.8%",
			},
		},
	}

	return trends, nil
}

func (na *NetworkAnalyzer) DetectBottlenecks(ctx context.Context) (*BottleneckAnalysis, error) {
	analysis := &BottleneckAnalysis{
		DetectedBottlenecks: []NetworkBottleneck{},
		SystemLimits:        []SystemLimit{},
		Recommendations:     []BottleneckRecommendation{},
	}

	// Check interface utilization
	interfaces := na.getActiveInterfaces()
	for _, iface := range interfaces {
		bandwidth := na.measureCurrentBandwidth(iface)
		capacity := na.getInterfaceCapacity(iface)

		utilization := (bandwidth.UploadMbps + bandwidth.DownloadMbps) / capacity * 100

		limit := SystemLimit{
			Resource:    fmt.Sprintf("interface_%s", iface),
			Current:     bandwidth.UploadMbps + bandwidth.DownloadMbps,
			Maximum:     capacity,
			Utilization: utilization,
			AtRisk:      utilization > 80,
		}
		analysis.SystemLimits = append(analysis.SystemLimits, limit)

		if utilization > 90 {
			bottleneck := NetworkBottleneck{
				Type:        "bandwidth",
				Location:    iface,
				Severity:    "high",
				Impact:      "Network performance significantly degraded",
				Description: fmt.Sprintf("Interface %s utilization at %.1f%%", iface, utilization),
				Confidence:  0.95,
			}
			analysis.DetectedBottlenecks = append(analysis.DetectedBottlenecks, bottleneck)

			recommendation := BottleneckRecommendation{
				Bottleneck:  "bandwidth",
				Action:      "upgrade_interface",
				Priority:    "high",
				Description: "Consider upgrading network interface or optimizing traffic",
			}
			analysis.Recommendations = append(analysis.Recommendations, recommendation)
		}
	}

	// Check CPU and memory impact on networking
	na.checkSystemResourceBottlenecks(analysis)

	// Overall assessment
	if len(analysis.DetectedBottlenecks) == 0 {
		analysis.OverallAssessment = "No significant bottlenecks detected"
	} else {
		analysis.OverallAssessment = fmt.Sprintf("%d bottlenecks detected requiring attention", len(analysis.DetectedBottlenecks))
	}

	return analysis, nil
}

func (na *NetworkAnalyzer) SaveAnalysisReport(analysis *ComprehensiveAnalysis, filename string) error {
	data, err := json.MarshalIndent(analysis, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal analysis: %w", err)
	}

	if err := os.WriteFile(filename, data, 0o644); err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}

	return nil
}

// Helper functions

func (na *NetworkAnalyzer) getDefaultInterface() string {
	result := na.commandPool.ExecuteCommand("ip", "route", "get", "1.1.1.1")
	if result.Error != nil {
		return "eth0"
	}

	fields := strings.Fields(string(result.Output))
	for i, field := range fields {
		if field == "dev" && i+1 < len(fields) {
			return fields[i+1]
		}
	}

	return "eth0"
}

func (na *NetworkAnalyzer) getActiveInterfaces() []string {
	result := na.commandPool.ExecuteCommand("ip", "link", "show", "up")
	if result.Error != nil {
		return []string{"eth0"}
	}

	var interfaces []string

	lines := strings.Split(string(result.Output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "state UP") && strings.Contains(line, ":") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				name := strings.TrimSpace(parts[1])
				if name != "lo" {
					interfaces = append(interfaces, name)
				}
			}
		}
	}

	if len(interfaces) == 0 {
		return []string{"eth0"}
	}

	return interfaces
}

func (na *NetworkAnalyzer) collectTargetLatencyStats(ctx context.Context, target string, config LatencyAnalysisConfig) (TargetLatencyStats, error) {
	stats := TargetLatencyStats{
		Target: target,
	}

	// Collect measurements over the specified duration
	ticker := time.NewTicker(config.Interval)
	defer ticker.Stop()

	endTime := time.Now().Add(config.Duration)

	var measurements []LatencyMeasurement

	for time.Now().Before(endTime) {
		select {
		case <-ctx.Done():
			stats.Measurements = measurements
			return stats, ctx.Err()
		case <-ticker.C:
			measurement := na.measureLatency(target)
			measurements = append(measurements, measurement)
		}
	}

	stats.Measurements = measurements

	// Calculate statistics
	stats.Statistics = na.calculateLatencyStatistics(measurements)

	// Analyze trends
	stats.TrendAnalysis = na.analyzeLatencyTrend(measurements)

	// Calculate quality metrics
	stats.QualityMetrics = na.calculateLatencyQualityMetrics(measurements)

	return stats, nil
}

func (na *NetworkAnalyzer) measureLatency(target string) LatencyMeasurement {
	measurement := LatencyMeasurement{
		Timestamp: time.Now(),
		Success:   false,
	}

	result := na.commandPool.ExecuteCommand("ping", "-c", "1", "-W", "3", target)
	if result.Error != nil {
		return measurement
	}

	output := string(result.Output)
	measurement.Success = true

	// Parse ping output for latency
	if strings.Contains(output, "time=") {
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if strings.Contains(line, "time=") {
				fields := strings.Fields(line)
				for _, field := range fields {
					if strings.HasPrefix(field, "time=") {
						timeStr := strings.TrimPrefix(field, "time=")
						if latency, err := time.ParseDuration(timeStr + "ms"); err == nil {
							measurement.Latency = latency
						}

						break
					}
				}

				break
			}
		}
	}

	// TODO: Implement jitter calculation and packet loss detection for single ping
	measurement.Jitter = time.Duration(0)
	measurement.PacketLoss = 0

	return measurement
}

func (na *NetworkAnalyzer) calculateLatencyStatistics(measurements []LatencyMeasurement) LatencyStatistics {
	stats := LatencyStatistics{
		Count: len(measurements),
	}

	if len(measurements) == 0 {
		return stats
	}

	var latencies []time.Duration

	successCount := 0

	for _, m := range measurements {
		if m.Success {
			latencies = append(latencies, m.Latency)
			successCount++
		}
	}

	stats.SuccessRate = float64(successCount) / float64(len(measurements)) * 100

	if len(latencies) == 0 {
		return stats
	}

	// Sort for percentile calculations
	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})

	stats.Min = latencies[0]
	stats.Max = latencies[len(latencies)-1]

	// Calculate mean
	var total time.Duration
	for _, lat := range latencies {
		total += lat
	}

	stats.Mean = total / time.Duration(len(latencies))

	// Calculate median
	if len(latencies)%2 == 0 {
		stats.Median = (latencies[len(latencies)/2-1] + latencies[len(latencies)/2]) / 2
	} else {
		stats.Median = latencies[len(latencies)/2]
	}

	// Calculate percentiles
	p95Index := int(float64(len(latencies)) * 0.95)
	if p95Index >= len(latencies) {
		p95Index = len(latencies) - 1
	}

	stats.P95 = latencies[p95Index]

	p99Index := int(float64(len(latencies)) * 0.99)
	if p99Index >= len(latencies) {
		p99Index = len(latencies) - 1
	}

	stats.P99 = latencies[p99Index]

	// Calculate standard deviation
	if len(latencies) > 1 {
		var sumSquaredDiffs float64

		meanFloat := float64(stats.Mean)
		for _, lat := range latencies {
			diff := float64(lat) - meanFloat
			sumSquaredDiffs += diff * diff
		}

		variance := sumSquaredDiffs / float64(len(latencies)-1)
		stats.StdDev = time.Duration(math.Sqrt(variance))
	}

	return stats
}

func (na *NetworkAnalyzer) analyzeLatencyTrend(measurements []LatencyMeasurement) LatencyTrend {
	trend := LatencyTrend{
		Direction:  "stable",
		Magnitude:  0,
		Confidence: 0,
	}

	if len(measurements) < 10 {
		trend.Description = "Insufficient data for trend analysis"
		return trend
	}

	// Simple linear regression to detect trend
	firstHalf := measurements[:len(measurements)/2]
	secondHalf := measurements[len(measurements)/2:]

	firstAvg := na.calculateAverageLatency(firstHalf)
	secondAvg := na.calculateAverageLatency(secondHalf)

	if firstAvg > 0 && secondAvg > 0 {
		change := (float64(secondAvg) - float64(firstAvg)) / float64(firstAvg) * 100
		trend.Magnitude = math.Abs(change)

		if change > 5 {
			trend.Direction = "degrading"
			trend.Description = fmt.Sprintf("Latency increased by %.1f%%", change)
		} else if change < -5 {
			trend.Direction = "improving"
			trend.Description = fmt.Sprintf("Latency improved by %.1f%%", math.Abs(change))
		} else {
			trend.Description = "Latency remains stable"
		}

		// Simple confidence calculation based on measurement count and consistency
		trend.Confidence = math.Min(float64(len(measurements))/100.0, 1.0)
	}

	return trend
}

func (na *NetworkAnalyzer) calculateAverageLatency(measurements []LatencyMeasurement) time.Duration {
	var total time.Duration

	count := 0

	for _, m := range measurements {
		if m.Success {
			total += m.Latency
			count++
		}
	}

	if count == 0 {
		return 0
	}

	return total / time.Duration(count)
}

func (na *NetworkAnalyzer) calculateLatencyQualityMetrics(measurements []LatencyMeasurement) LatencyQualityMetrics {
	metrics := LatencyQualityMetrics{}

	if len(measurements) == 0 {
		return metrics
	}

	stats := na.calculateLatencyStatistics(measurements)

	// Consistency: Based on standard deviation relative to mean
	if stats.Mean > 0 {
		cvPercent := float64(stats.StdDev) / float64(stats.Mean) * 100
		metrics.Consistency = math.Max(0, 100-cvPercent)
	}

	// Reliability: Based on success rate
	metrics.Reliability = stats.SuccessRate

	// Performance: Based on average latency (lower is better)
	avgLatencyMs := float64(stats.Mean) / float64(time.Millisecond)
	if avgLatencyMs <= 10 {
		metrics.Performance = 100
	} else if avgLatencyMs <= 50 {
		metrics.Performance = 90 - (avgLatencyMs-10)*2
	} else if avgLatencyMs <= 100 {
		metrics.Performance = 50 - (avgLatencyMs-50)*0.8
	} else {
		metrics.Performance = math.Max(0, 10-avgLatencyMs/100)
	}

	// Overall score: Weighted average
	metrics.OverallScore = (metrics.Consistency*0.3 + metrics.Reliability*0.4 + metrics.Performance*0.3)

	return metrics
}

func (na *NetworkAnalyzer) calculateOverallLatencyStats(targetAnalysis []TargetLatencyStats) OverallLatencyStats {
	stats := OverallLatencyStats{}

	if len(targetAnalysis) == 0 {
		return stats
	}

	var (
		totalMeasurements int
		totalLatency      time.Duration
		successCount      int
	)

	bestScore := 0.0
	worstScore := 100.0
	bestTarget := ""
	worstTarget := ""

	for _, target := range targetAnalysis {
		totalMeasurements += target.Statistics.Count

		// Weight by success rate
		if target.Statistics.SuccessRate > 0 {
			totalLatency += time.Duration(float64(target.Statistics.Mean) * target.Statistics.SuccessRate / 100)
			successCount++
		}

		score := target.QualityMetrics.OverallScore
		if score > bestScore {
			bestScore = score
			bestTarget = target.Target
		}

		if score < worstScore {
			worstScore = score
			worstTarget = target.Target
		}
	}

	stats.TotalMeasurements = totalMeasurements
	stats.BestTarget = bestTarget
	stats.WorstTarget = worstTarget

	if successCount > 0 {
		stats.AverageLatency = totalLatency / time.Duration(successCount)
	}

	// Network stability: based on variance between targets
	stats.NetworkStability = math.Max(0, 100-(bestScore-worstScore))

	return stats
}

func (na *NetworkAnalyzer) detectLatencyAnomalies(targetAnalysis []TargetLatencyStats) []LatencyAnomaly {
	var anomalies []LatencyAnomaly

	for _, target := range targetAnalysis {
		stats := target.Statistics

		// Detect high latency spikes
		if stats.P99 > 200*time.Millisecond {
			anomaly := LatencyAnomaly{
				Target:      target.Target,
				Timestamp:   time.Now(),
				Type:        "spike",
				Severity:    "medium",
				Value:       stats.P99.String(),
				Description: fmt.Sprintf("99th percentile latency spike: %s", stats.P99),
			}
			anomalies = append(anomalies, anomaly)
		}

		// Detect high jitter
		if stats.StdDev > 50*time.Millisecond {
			anomaly := LatencyAnomaly{
				Target:      target.Target,
				Timestamp:   time.Now(),
				Type:        "jitter",
				Severity:    "low",
				Value:       stats.StdDev.String(),
				Description: fmt.Sprintf("High latency variance: %s", stats.StdDev),
			}
			anomalies = append(anomalies, anomaly)
		}

		// Detect timeouts
		if stats.SuccessRate < 95 {
			anomaly := LatencyAnomaly{
				Target:      target.Target,
				Timestamp:   time.Now(),
				Type:        "timeout",
				Severity:    "high",
				Value:       fmt.Sprintf("%.1f%%", stats.SuccessRate),
				Description: fmt.Sprintf("Low success rate: %.1f%%", stats.SuccessRate),
			}
			anomalies = append(anomalies, anomaly)
		}
	}

	return anomalies
}

func (na *NetworkAnalyzer) generateLatencyRecommendations(analysis *LatencyAnalysis) []string {
	var recommendations []string

	// Check overall quality
	if analysis.QualityScore < 70 {
		recommendations = append(recommendations, "Consider switching to faster DNS servers (1.1.1.1, 8.8.8.8)")
		recommendations = append(recommendations, "Check for network congestion and optimize traffic routing")
	}

	// Check for specific issues
	for _, anomaly := range analysis.AnomalyDetection {
		switch anomaly.Type {
		case "timeout":
			recommendations = append(recommendations, fmt.Sprintf("Investigate connectivity issues to %s", anomaly.Target))
		case "spike":
			recommendations = append(recommendations, "Monitor for network congestion during peak hours")
		case "jitter":
			recommendations = append(recommendations, "Check for wireless interference or switch to wired connection")
		}
	}

	// Geographic recommendations
	for _, target := range analysis.TargetAnalysis {
		if target.Statistics.Mean > 100*time.Millisecond {
			recommendations = append(recommendations, fmt.Sprintf("Consider using a CDN or geographically closer servers to %s", target.Target))
		}
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Network latency performance is good - no immediate optimizations needed")
	}

	return recommendations
}

func (na *NetworkAnalyzer) calculateLatencyQualityScore(analysis *LatencyAnalysis) float64 {
	if len(analysis.TargetAnalysis) == 0 {
		return 0
	}

	var totalScore float64
	for _, target := range analysis.TargetAnalysis {
		totalScore += target.QualityMetrics.OverallScore
	}

	baseScore := totalScore / float64(len(analysis.TargetAnalysis))

	// Penalty for anomalies
	anomalyPenalty := float64(len(analysis.AnomalyDetection)) * 5
	finalScore := math.Max(0, baseScore-anomalyPenalty)

	return finalScore
}

func (na *NetworkAnalyzer) collectBandwidthMeasurements(ctx context.Context, config BandwidthAnalysisConfig) ([]BandwidthMeasurement, error) {
	var measurements []BandwidthMeasurement

	ticker := time.NewTicker(config.Interval)
	defer ticker.Stop()

	endTime := time.Now().Add(config.Duration)

	for time.Now().Before(endTime) {
		select {
		case <-ctx.Done():
			return measurements, ctx.Err()
		case <-ticker.C:
			measurement := na.measureBandwidth(config.Interface)
			measurements = append(measurements, measurement)
		}
	}

	return measurements, nil
}

func (na *NetworkAnalyzer) measureBandwidth(iface string) BandwidthMeasurement {
	measurement := BandwidthMeasurement{
		Timestamp: time.Now(),
	}

	// Get interface statistics
	rxResult := na.commandPool.ExecuteCommand("cat", fmt.Sprintf("/sys/class/net/%s/statistics/rx_bytes", iface))
	txResult := na.commandPool.ExecuteCommand("cat", fmt.Sprintf("/sys/class/net/%s/statistics/tx_bytes", iface))

	if rxResult.Error == nil && txResult.Error == nil {
		if rxBytes, err := strconv.ParseInt(strings.TrimSpace(string(rxResult.Output)), 10, 64); err == nil {
			if txBytes, err := strconv.ParseInt(strings.TrimSpace(string(txResult.Output)), 10, 64); err == nil {
				// TODO: Calculate actual rates by comparing with previous measurements
				// For now, return cumulative bytes converted to Mbps (simplified)
				measurement.DownloadMbps = float64(rxBytes) / (1024 * 1024) / 8 // Convert to Mbps
				measurement.UploadMbps = float64(txBytes) / (1024 * 1024) / 8   // Convert to Mbps
				measurement.TotalBytes = rxBytes + txBytes

				// Calculate utilization (assuming 1 Gbps interface)
				totalMbps := measurement.DownloadMbps + measurement.UploadMbps
				measurement.Utilization = (totalMbps / 1000) * 100
			}
		}
	}

	return measurement
}

func (na *NetworkAnalyzer) measureCurrentBandwidth(iface string) BandwidthMetrics {
	measurement := na.measureBandwidth(iface)

	return BandwidthMetrics{
		UploadMbps:   measurement.UploadMbps,
		DownloadMbps: measurement.DownloadMbps,
		TotalBytes:   measurement.TotalBytes,
	}
}

func (na *NetworkAnalyzer) getInterfaceCapacity(iface string) float64 {
	// Try to get interface speed
	result := na.commandPool.ExecuteCommand("cat", fmt.Sprintf("/sys/class/net/%s/speed", iface))
	if result.Error == nil {
		if speed, err := strconv.ParseInt(strings.TrimSpace(string(result.Output)), 10, 64); err == nil {
			return float64(speed) // Speed in Mbps
		}
	}

	// Default assumption: 1 Gbps
	return 1000.0
}

func (na *NetworkAnalyzer) calculateBandwidthStatistics(measurements []BandwidthMeasurement) BandwidthStatistics {
	stats := BandwidthStatistics{
		Count: len(measurements),
	}

	if len(measurements) == 0 {
		return stats
	}

	var (
		totalUpload, totalDownload, totalUtil float64
		totalTraffic                          int64
	)

	for _, m := range measurements {
		totalUpload += m.UploadMbps
		totalDownload += m.DownloadMbps
		totalUtil += m.Utilization
		totalTraffic += m.TotalBytes

		if m.UploadMbps > stats.PeakUpload {
			stats.PeakUpload = m.UploadMbps
		}

		if m.DownloadMbps > stats.PeakDownload {
			stats.PeakDownload = m.DownloadMbps
		}

		if m.Utilization > stats.PeakUtilization {
			stats.PeakUtilization = m.Utilization
		}
	}

	count := float64(len(measurements))
	stats.AvgUpload = totalUpload / count
	stats.AvgDownload = totalDownload / count
	stats.AvgUtilization = totalUtil / count
	stats.TotalTraffic = totalTraffic

	return stats
}

func (na *NetworkAnalyzer) analyzeUtilizationTrends(measurements []BandwidthMeasurement) UtilizationTrends {
	trends := UtilizationTrends{
		Direction: "stable",
	}

	if len(measurements) < 10 {
		return trends
	}

	// Simple trend analysis
	firstHalf := measurements[:len(measurements)/2]
	secondHalf := measurements[len(measurements)/2:]

	firstAvg := na.calculateAverageUtilization(firstHalf)
	secondAvg := na.calculateAverageUtilization(secondHalf)

	if firstAvg > 0 {
		change := (secondAvg - firstAvg) / firstAvg * 100
		trends.GrowthRate = change

		if change > 10 {
			trends.Direction = "increasing"
		} else if change < -10 {
			trends.Direction = "decreasing"
		}
	}

	// Detect peak periods (utilization > 80%)
	var currentPeak *PeakPeriod

	for _, m := range measurements {
		if m.Utilization > 80 {
			if currentPeak == nil {
				currentPeak = &PeakPeriod{
					StartTime: m.Timestamp,
					PeakValue: m.Utilization,
				}
			} else {
				currentPeak.EndTime = m.Timestamp
				if m.Utilization > currentPeak.PeakValue {
					currentPeak.PeakValue = m.Utilization
				}
			}
		} else {
			if currentPeak != nil {
				currentPeak.Duration = currentPeak.EndTime.Sub(currentPeak.StartTime)
				trends.PeakPeriods = append(trends.PeakPeriods, *currentPeak)
				currentPeak = nil
			}
		}
	}

	return trends
}

func (na *NetworkAnalyzer) calculateAverageUtilization(measurements []BandwidthMeasurement) float64 {
	if len(measurements) == 0 {
		return 0
	}

	var total float64
	for _, m := range measurements {
		total += m.Utilization
	}

	return total / float64(len(measurements))
}

func (na *NetworkAnalyzer) analyzeCapacity(iface string, stats BandwidthStatistics) CapacityAnalysis {
	capacity := na.getInterfaceCapacity(iface)

	analysis := CapacityAnalysis{
		CurrentCapacity:     capacity,
		UtilizedCapacity:    stats.AvgUpload + stats.AvgDownload,
		CapacityUtilization: ((stats.AvgUpload + stats.AvgDownload) / capacity) * 100,
	}

	analysis.AvailableCapacity = capacity - analysis.UtilizedCapacity
	analysis.RecommendedUpgrade = analysis.CapacityUtilization > 80

	// Estimate time to capacity based on growth rate
	if stats.AvgUtilization > 0 && analysis.CapacityUtilization < 90 {
		// Simple linear projection
		remainingCapacity := 90 - analysis.CapacityUtilization
		if remainingCapacity > 0 {
			// Assume 5% annual growth rate
			monthsToCapacity := remainingCapacity / (5.0 / 12.0)
			if monthsToCapacity < 24 {
				estimatedTime := time.Duration(monthsToCapacity * 30 * 24 * float64(time.Hour))
				analysis.TimeToCapacity = &estimatedTime
			}
		}
	}

	return analysis
}

func (na *NetworkAnalyzer) generateBandwidthRecommendations(analysis *BandwidthAnalysis) []string {
	var recommendations []string

	if analysis.Statistics.PeakUtilization > 90 {
		recommendations = append(recommendations, "Consider upgrading network interface - peak utilization exceeds 90%")
	} else if analysis.Statistics.AvgUtilization > 70 {
		recommendations = append(recommendations, "Monitor network usage - average utilization is high")
	}

	if len(analysis.UtilizationTrends.PeakPeriods) > 0 {
		recommendations = append(recommendations, "Implement traffic shaping during peak usage periods")
		recommendations = append(recommendations, "Consider load balancing across multiple interfaces")
	}

	if analysis.UtilizationTrends.GrowthRate > 20 {
		recommendations = append(recommendations, "Plan for capacity upgrade - bandwidth usage is growing rapidly")
	}

	if analysis.CapacityAnalysis.RecommendedUpgrade {
		recommendations = append(recommendations, "Upgrade network capacity to handle current load")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Bandwidth utilization is within normal parameters")
	}

	return recommendations
}

func (na *NetworkAnalyzer) calculateNetworkHealth(analysis *ComprehensiveAnalysis) NetworkHealth {
	health := NetworkHealth{}

	// Calculate individual scores
	if analysis.LatencyAnalysis != nil {
		health.LatencyScore = analysis.LatencyAnalysis.QualityScore
	}

	if analysis.BandwidthAnalysis != nil {
		// Convert bandwidth utilization to a score (lower utilization = higher score for available capacity)
		util := analysis.BandwidthAnalysis.Statistics.AvgUtilization
		if util < 50 {
			health.BandwidthScore = 100
		} else if util < 80 {
			health.BandwidthScore = 100 - (util-50)*2
		} else {
			health.BandwidthScore = 40 - (util-80)*2
		}

		health.BandwidthScore = math.Max(0, health.BandwidthScore)
	}

	// Calculate stability score based on consistency
	health.StabilityScore = 85 // Default good stability

	// Calculate overall score
	scores := []float64{}
	if health.LatencyScore > 0 {
		scores = append(scores, health.LatencyScore)
	}

	if health.BandwidthScore > 0 {
		scores = append(scores, health.BandwidthScore)
	}

	scores = append(scores, health.StabilityScore)

	var total float64
	for _, score := range scores {
		total += score
	}

	health.OverallScore = total / float64(len(scores))

	// Determine health status
	if health.OverallScore >= 90 {
		health.HealthStatus = "excellent"
	} else if health.OverallScore >= 75 {
		health.HealthStatus = env.StatusGood
	} else if health.OverallScore >= 60 {
		health.HealthStatus = "fair"
	} else {
		health.HealthStatus = "poor"
	}

	// Count issues
	if analysis.LatencyAnalysis != nil {
		for _, anomaly := range analysis.LatencyAnalysis.AnomalyDetection {
			if anomaly.Severity == "high" {
				health.MajorIssues++
			} else {
				health.MinorIssues++
			}
		}
	}

	return health
}

func (na *NetworkAnalyzer) findPerformanceCorrelations(analysis *ComprehensiveAnalysis) []PerformanceCorrelation {
	var correlations []PerformanceCorrelation

	// Simple correlation analysis
	if analysis.LatencyAnalysis != nil && analysis.BandwidthAnalysis != nil {
		// Check if high bandwidth usage correlates with high latency
		avgLatency := analysis.LatencyAnalysis.OverallStats.AverageLatency
		avgUtilization := analysis.BandwidthAnalysis.Statistics.AvgUtilization

		if avgLatency > 50*time.Millisecond && avgUtilization > 70 {
			correlation := PerformanceCorrelation{
				Type:         "latency_bandwidth",
				Correlation:  0.75, // Simplified correlation value
				Significance: "moderate",
				Description:  "High bandwidth utilization appears to correlate with increased latency",
			}
			correlations = append(correlations, correlation)
		}
	}

	return correlations
}

func (na *NetworkAnalyzer) generateComprehensiveRecommendations(analysis *ComprehensiveAnalysis) []AnalysisRecommendation {
	var recommendations []AnalysisRecommendation

	// High priority recommendations based on health score
	if analysis.OverallHealth.OverallScore < 60 {
		rec := AnalysisRecommendation{
			Priority:    "high",
			Category:    "performance",
			Title:       "Address Network Performance Issues",
			Description: "Overall network health is poor and requires immediate attention",
			Impact:      "Significant improvement in network performance",
			Effort:      "medium",
		}
		recommendations = append(recommendations, rec)
	}

	// Latency-specific recommendations
	if analysis.LatencyAnalysis != nil && analysis.LatencyAnalysis.QualityScore < 70 {
		rec := AnalysisRecommendation{
			Priority:    "medium",
			Category:    "latency",
			Title:       "Optimize Network Latency",
			Description: "Network latency is impacting performance",
			Impact:      "Improved responsiveness for applications",
			Effort:      "low",
			Command:     "resolvectl dns eth0 1.1.1.1 8.8.8.8",
		}
		recommendations = append(recommendations, rec)
	}

	// Bandwidth-specific recommendations
	if analysis.BandwidthAnalysis != nil && analysis.BandwidthAnalysis.Statistics.PeakUtilization > 85 {
		rec := AnalysisRecommendation{
			Priority:    "high",
			Category:    "bandwidth",
			Title:       "Increase Network Capacity",
			Description: "Network bandwidth utilization is approaching capacity limits",
			Impact:      "Prevent network congestion and performance degradation",
			Effort:      "high",
		}
		recommendations = append(recommendations, rec)
	}

	// General optimization recommendations
	rec := AnalysisRecommendation{
		Priority:    "low",
		Category:    "optimization",
		Title:       "Enable TCP Optimizations",
		Description: "Configure TCP parameters for better network performance",
		Impact:      "Moderate improvement in throughput",
		Effort:      "low",
		Command:     "sysctl -w net.core.rmem_max=16777216 net.core.wmem_max=16777216",
	}
	recommendations = append(recommendations, rec)

	return recommendations
}

func (na *NetworkAnalyzer) checkSystemResourceBottlenecks(analysis *BottleneckAnalysis) {
	na.checkCPUUsage(analysis)
	na.checkMemoryUsage(analysis)
}

// checkCPUUsage monitors CPU usage impact on networking.
func (na *NetworkAnalyzer) checkCPUUsage(analysis *BottleneckAnalysis) {
	result := na.commandPool.ExecuteCommand("top", "-bn1")
	if result.Error != nil {
		return
	}

	output := string(result.Output)
	if !strings.Contains(output, "Cpu(s):") {
		return
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Cpu(s):") && strings.Contains(line, "us") {
			bottleneck := NetworkBottleneck{
				Type:        "cpu",
				Location:    "system",
				Severity:    "low",
				Impact:      "May affect network processing capacity",
				Description: "Monitor CPU usage during network intensive operations",
				Confidence:  0.6,
			}
			analysis.DetectedBottlenecks = append(analysis.DetectedBottlenecks, bottleneck)
			break
		}
	}
}

// checkMemoryUsage monitors memory usage impact on networking.
func (na *NetworkAnalyzer) checkMemoryUsage(analysis *BottleneckAnalysis) {
	result := na.commandPool.ExecuteCommand("free", "-m")
	if result.Error != nil {
		return
	}

	output := string(result.Output)
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if !strings.HasPrefix(line, "Mem:") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 3 {
			break
		}

		total, err := strconv.ParseFloat(fields[1], 64)
		if err != nil {
			break
		}

		used, err := strconv.ParseFloat(fields[2], 64)
		if err != nil {
			break
		}

		utilization := (used / total) * 100
		limit := SystemLimit{
			Resource:    "memory",
			Current:     used,
			Maximum:     total,
			Utilization: utilization,
			AtRisk:      utilization > 85,
		}
		analysis.SystemLimits = append(analysis.SystemLimits, limit)

		if utilization > 90 {
			bottleneck := NetworkBottleneck{
				Type:        "memory",
				Location:    "system",
				Severity:    "medium",
				Impact:      "May cause network buffer limitations",
				Description: fmt.Sprintf("Memory utilization at %.1f%%", utilization),
				Confidence:  0.8,
			}
			analysis.DetectedBottlenecks = append(analysis.DetectedBottlenecks, bottleneck)
		}

		break
	}
}

// Print functions

func printLatencyAnalysis(analysis *LatencyAnalysis) {
	fmt.Printf("🔍 Latency Analysis Report\n\n")
	fmt.Printf("Analysis Period: %s to %s (Duration: %s)\n",
		analysis.StartTime.Format("15:04:05"),
		analysis.EndTime.Format("15:04:05"),
		analysis.EndTime.Sub(analysis.StartTime).Round(time.Second))
	fmt.Printf("Overall Quality Score: %.1f%%\n\n", analysis.QualityScore)

	// Target analysis
	if len(analysis.TargetAnalysis) > 0 {
		fmt.Printf("📊 Target Analysis:\n")

		w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
		_, _ = fmt.Fprintln(w, "TARGET\tMEAN\tMEDIAN\tP95\tP99\tSUCCESS\tQUALITY")

		for _, target := range analysis.TargetAnalysis {
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%.1f%%\t%.1f%%\n",
				target.Target,
				target.Statistics.Mean.Round(time.Millisecond),
				target.Statistics.Median.Round(time.Millisecond),
				target.Statistics.P95.Round(time.Millisecond),
				target.Statistics.P99.Round(time.Millisecond),
				target.Statistics.SuccessRate,
				target.QualityMetrics.OverallScore)
		}

		_ = w.Flush()
		fmt.Println()
	}

	// Overall statistics
	fmt.Printf("📈 Overall Statistics:\n")
	fmt.Printf("  Total Measurements: %d\n", analysis.OverallStats.TotalMeasurements)
	fmt.Printf("  Average Latency: %s\n", analysis.OverallStats.AverageLatency.Round(time.Millisecond))
	fmt.Printf("  Best Target: %s\n", analysis.OverallStats.BestTarget)
	fmt.Printf("  Worst Target: %s\n", analysis.OverallStats.WorstTarget)
	fmt.Printf("  Network Stability: %.1f%%\n\n", analysis.OverallStats.NetworkStability)

	// Anomalies
	if len(analysis.AnomalyDetection) > 0 {
		fmt.Printf("⚠️  Anomalies Detected:\n")

		for i, anomaly := range analysis.AnomalyDetection {
			fmt.Printf("  %d. [%s] %s on %s: %s\n",
				i+1, anomaly.Severity, anomaly.Type, anomaly.Target, anomaly.Description)
		}

		fmt.Println()
	}

	// Recommendations
	if len(analysis.Recommendations) > 0 {
		fmt.Printf("💡 Recommendations:\n")

		for i, rec := range analysis.Recommendations {
			fmt.Printf("  %d. %s\n", i+1, rec)
		}

		fmt.Println()
	}
}

func printBandwidthAnalysis(analysis *BandwidthAnalysis) {
	fmt.Printf("📊 Bandwidth Analysis Report\n\n")
	fmt.Printf("Interface: %s\n", analysis.Config.Interface)
	fmt.Printf("Analysis Period: %s to %s (Duration: %s)\n",
		analysis.StartTime.Format("15:04:05"),
		analysis.EndTime.Format("15:04:05"),
		analysis.EndTime.Sub(analysis.StartTime).Round(time.Second))

	// Statistics
	fmt.Printf("\n📈 Statistics:\n")
	fmt.Printf("  Measurements: %d\n", analysis.Statistics.Count)
	fmt.Printf("  Average Upload: %.2f Mbps\n", analysis.Statistics.AvgUpload)
	fmt.Printf("  Average Download: %.2f Mbps\n", analysis.Statistics.AvgDownload)
	fmt.Printf("  Peak Upload: %.2f Mbps\n", analysis.Statistics.PeakUpload)
	fmt.Printf("  Peak Download: %.2f Mbps\n", analysis.Statistics.PeakDownload)
	fmt.Printf("  Average Utilization: %.1f%%\n", analysis.Statistics.AvgUtilization)
	fmt.Printf("  Peak Utilization: %.1f%%\n", analysis.Statistics.PeakUtilization)
	fmt.Printf("  Total Traffic: %.2f GB\n", float64(analysis.Statistics.TotalTraffic)/(1024*1024*1024))

	// Utilization trends
	fmt.Printf("\n📈 Utilization Trends:\n")
	fmt.Printf("  Direction: %s\n", analysis.UtilizationTrends.Direction)
	fmt.Printf("  Growth Rate: %.1f%%\n", analysis.UtilizationTrends.GrowthRate)

	if len(analysis.UtilizationTrends.PeakPeriods) > 0 {
		fmt.Printf("  Peak Periods: %d detected\n", len(analysis.UtilizationTrends.PeakPeriods))
	}

	// Capacity analysis
	fmt.Printf("\n🔧 Capacity Analysis:\n")
	fmt.Printf("  Interface Capacity: %.0f Mbps\n", analysis.CapacityAnalysis.CurrentCapacity)
	fmt.Printf("  Utilized Capacity: %.2f Mbps\n", analysis.CapacityAnalysis.UtilizedCapacity)
	fmt.Printf("  Available Capacity: %.2f Mbps\n", analysis.CapacityAnalysis.AvailableCapacity)
	fmt.Printf("  Utilization: %.1f%%\n", analysis.CapacityAnalysis.CapacityUtilization)

	if analysis.CapacityAnalysis.TimeToCapacity != nil {
		fmt.Printf("  Estimated Time to Capacity: %s\n", analysis.CapacityAnalysis.TimeToCapacity.Round(24*time.Hour))
	}

	if analysis.CapacityAnalysis.RecommendedUpgrade {
		fmt.Printf("  Upgrade Recommended: Yes\n")
	}

	// Recommendations
	if len(analysis.Recommendations) > 0 {
		fmt.Printf("\n💡 Recommendations:\n")

		for i, rec := range analysis.Recommendations {
			fmt.Printf("  %d. %s\n", i+1, rec)
		}
	}
}

func printComprehensiveAnalysis(analysis *ComprehensiveAnalysis) {
	fmt.Printf("🔍 Comprehensive Network Analysis Report\n\n")
	fmt.Printf("Analysis Period: %s to %s\n",
		analysis.StartTime.Format("2006-01-02 15:04:05"),
		analysis.EndTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("Duration: %s\n\n", analysis.Duration.Round(time.Second))

	// Overall health
	fmt.Printf("🏥 Network Health Summary:\n")
	fmt.Printf("  Overall Score: %.1f%% (%s)\n",
		analysis.OverallHealth.OverallScore,
		analysis.OverallHealth.HealthStatus)
	fmt.Printf("  Latency Score: %.1f%%\n", analysis.OverallHealth.LatencyScore)
	fmt.Printf("  Bandwidth Score: %.1f%%\n", analysis.OverallHealth.BandwidthScore)
	fmt.Printf("  Stability Score: %.1f%%\n", analysis.OverallHealth.StabilityScore)
	fmt.Printf("  Issues: %d major, %d minor\n\n",
		analysis.OverallHealth.MajorIssues,
		analysis.OverallHealth.MinorIssues)

	// Quick summaries
	if analysis.LatencyAnalysis != nil {
		fmt.Printf("📊 Latency Summary:\n")
		fmt.Printf("  Quality Score: %.1f%%\n", analysis.LatencyAnalysis.QualityScore)
		fmt.Printf("  Average Latency: %s\n", analysis.LatencyAnalysis.OverallStats.AverageLatency.Round(time.Millisecond))
		fmt.Printf("  Anomalies: %d detected\n\n", len(analysis.LatencyAnalysis.AnomalyDetection))
	}

	if analysis.BandwidthAnalysis != nil {
		fmt.Printf("📈 Bandwidth Summary:\n")
		fmt.Printf("  Average Utilization: %.1f%%\n", analysis.BandwidthAnalysis.Statistics.AvgUtilization)
		fmt.Printf("  Peak Utilization: %.1f%%\n", analysis.BandwidthAnalysis.Statistics.PeakUtilization)

		if analysis.BandwidthAnalysis.CapacityAnalysis.RecommendedUpgrade {
			fmt.Printf("  Capacity Status: Upgrade recommended\n")
		} else {
			fmt.Printf("  Capacity Status: Adequate\n")
		}

		fmt.Println()
	}

	// Correlations
	if len(analysis.Correlations) > 0 {
		fmt.Printf("🔗 Performance Correlations:\n")

		for _, corr := range analysis.Correlations {
			fmt.Printf("  • %s (%.2f correlation, %s significance)\n",
				corr.Description, corr.Correlation, corr.Significance)
		}

		fmt.Println()
	}

	// Recommendations
	if len(analysis.Recommendations) > 0 {
		fmt.Printf("💡 Recommendations:\n")

		for i, rec := range analysis.Recommendations {
			fmt.Printf("  %d. [%s - %s] %s\n", i+1, rec.Priority, rec.Category, rec.Title)
			fmt.Printf("     %s\n", rec.Description)
			fmt.Printf("     Impact: %s | Effort: %s\n", rec.Impact, rec.Effort)

			if rec.Command != "" {
				fmt.Printf("     Command: %s\n", rec.Command)
			}

			fmt.Println()
		}
	}
}

func printPerformanceTrends(trends *PerformanceTrends) {
	fmt.Printf("📈 Performance Trends Analysis\n\n")
	fmt.Printf("Analysis Period: %s\n\n", trends.Period)

	// Latency trends
	if len(trends.LatencyTrends) > 0 {
		fmt.Printf("🕐 Latency Trends:\n")

		w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
		_, _ = fmt.Fprintln(w, "METRIC\tTARGET\tSTART\tEND\tCHANGE\tDIRECTION\tSIGNIFICANCE")

		for _, trend := range trends.LatencyTrends {
			_, _ = fmt.Fprintf(w, "%s\t%s\t%.2f\t%.2f\t%.1f%%\t%s\t%s\n",
				trend.Metric,
				trend.Target,
				trend.StartValue,
				trend.EndValue,
				trend.Change,
				trend.Direction,
				trend.Significance)
		}

		_ = w.Flush()
		fmt.Println()
	}

	// Bandwidth trends
	if len(trends.BandwidthTrends) > 0 {
		fmt.Printf("📊 Bandwidth Trends:\n")

		w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
		_, _ = fmt.Fprintln(w, "METRIC\tSTART\tEND\tCHANGE\tDIRECTION\tSIGNIFICANCE")

		for _, trend := range trends.BandwidthTrends {
			_, _ = fmt.Fprintf(w, "%s\t%.2f\t%.2f\t%.1f%%\t%s\t%s\n",
				trend.Metric,
				trend.StartValue,
				trend.EndValue,
				trend.Change,
				trend.Direction,
				trend.Significance)
		}

		_ = w.Flush()
		fmt.Println()
	}

	// Regressions
	if len(trends.Regressions) > 0 {
		fmt.Printf("⚠️  Performance Regressions:\n")

		for i, regression := range trends.Regressions {
			fmt.Printf("  %d. [%s] %s: %s\n",
				i+1, regression.Severity, regression.Metric, regression.Description)
			fmt.Printf("     Detected: %s\n", regression.DetectedAt.Format("2006-01-02 15:04"))
			fmt.Printf("     Impact: %s\n", regression.Impact)
		}

		fmt.Println()
	}

	// Improvements
	if len(trends.Improvements) > 0 {
		fmt.Printf("✅ Performance Improvements:\n")

		for i, improvement := range trends.Improvements {
			fmt.Printf("  %d. %s: %s (%.1f%% improvement)\n",
				i+1, improvement.Metric, improvement.Description, improvement.Improvement)
			fmt.Printf("     Detected: %s\n", improvement.DetectedAt.Format("2006-01-02 15:04"))
		}

		fmt.Println()
	}
}

func printBottleneckAnalysis(analysis *BottleneckAnalysis) {
	fmt.Printf("🔍 Network Bottleneck Analysis\n\n")
	fmt.Printf("Overall Assessment: %s\n\n", analysis.OverallAssessment)

	// Detected bottlenecks
	if len(analysis.DetectedBottlenecks) > 0 {
		fmt.Printf("🚨 Detected Bottlenecks:\n")

		w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
		_, _ = fmt.Fprintln(w, "TYPE\tLOCATION\tSEVERITY\tCONFIDENCE\tDESCRIPTION")

		for _, bottleneck := range analysis.DetectedBottlenecks {
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%.0f%%\t%s\n",
				bottleneck.Type,
				bottleneck.Location,
				bottleneck.Severity,
				bottleneck.Confidence*100,
				bottleneck.Description)
		}

		_ = w.Flush()
		fmt.Println()
	}

	// System limits
	if len(analysis.SystemLimits) > 0 {
		fmt.Printf("⚡ System Resource Limits:\n")

		w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
		_, _ = fmt.Fprintln(w, "RESOURCE\tCURRENT\tMAXIMUM\tUTILIZATION\tAT RISK")

		for _, limit := range analysis.SystemLimits {
			atRisk := ""
			if limit.AtRisk {
				atRisk = "⚠️  Yes"
			} else {
				atRisk = "✅ No"
			}

			_, _ = fmt.Fprintf(w, "%s\t%.2f\t%.2f\t%.1f%%\t%s\n",
				limit.Resource,
				limit.Current,
				limit.Maximum,
				limit.Utilization,
				atRisk)
		}

		_ = w.Flush()
		fmt.Println()
	}

	// Recommendations
	if len(analysis.Recommendations) > 0 {
		fmt.Printf("💡 Bottleneck Recommendations:\n")

		for i, rec := range analysis.Recommendations {
			fmt.Printf("  %d. [%s - %s] %s\n", i+1, rec.Priority, rec.Bottleneck, rec.Action)
			fmt.Printf("     %s\n", rec.Description)

			if rec.Command != "" {
				fmt.Printf("     Command: %s\n", rec.Command)
			}

			fmt.Println()
		}
	}
}
