package reports

import (
	"fmt"
	"sort"
	"time"
)

// ComprehensiveNetworkReport represents a complete network analysis
type ComprehensiveNetworkReport struct {
	Summary         NetworkSummary            `json:"summary"`
	Interfaces      []InterfaceReport         `json:"interfaces"`
	BandwidthTrends map[string]BandwidthTrend `json:"bandwidth_trends"`
	LatencyMetrics  LatencyReport             `json:"latency_metrics"`
	Recommendations []string                  `json:"recommendations"`
	Timestamp       time.Time                 `json:"timestamp"`
	Duration        time.Duration             `json:"duration"`
	SystemInfo      SystemInfo                `json:"system_info"`
}

// NetworkSummary provides high-level network statistics
type NetworkSummary struct {
	TotalInterfaces    int       `json:"total_interfaces"`
	ActiveInterfaces   int       `json:"active_interfaces"`
	TotalBandwidth     uint64    `json:"total_bandwidth_bps"`
	UsedBandwidth      uint64    `json:"used_bandwidth_bps"`
	UtilizationPercent float64   `json:"utilization_percent"`
	AverageLatency     float64   `json:"average_latency_ms"`
	PacketLossPercent  float64   `json:"packet_loss_percent"`
	TopInterface       string    `json:"top_interface_by_usage"`
	PeakUsageTimestamp time.Time `json:"peak_usage_timestamp"`
}

// InterfaceReport contains detailed interface information
type InterfaceReport struct {
	Name          string  `json:"name"`
	Status        string  `json:"status"`
	Type          string  `json:"type"`
	MaxSpeed      uint64  `json:"max_speed_bps"`
	MaxSpeedStr   string  `json:"max_speed_human"`
	CurrentRxRate uint64  `json:"current_rx_rate_bps"`
	CurrentTxRate uint64  `json:"current_tx_rate_bps"`
	Utilization   float64 `json:"utilization_percent"`
	PacketsRx     uint64  `json:"packets_rx"`
	PacketsTx     uint64  `json:"packets_tx"`
	ErrorsRx      uint64  `json:"errors_rx"`
	ErrorsTx      uint64  `json:"errors_tx"`
	DroppedRx     uint64  `json:"dropped_rx"`
	DroppedTx     uint64  `json:"dropped_tx"`
	MTU           int     `json:"mtu"`
	Driver        string  `json:"driver,omitempty"`
}

// LatencyReport contains network latency measurements
type LatencyReport struct {
	Targets          []LatencyTarget `json:"targets"`
	AverageLatency   float64         `json:"average_latency_ms"`
	MinLatency       float64         `json:"min_latency_ms"`
	MaxLatency       float64         `json:"max_latency_ms"`
	PacketLoss       float64         `json:"packet_loss_percent"`
	JitterAverage    float64         `json:"jitter_average_ms"`
	ReachabilityRate float64         `json:"reachability_rate_percent"`
}

// LatencyTarget represents a ping target with results
type LatencyTarget struct {
	Host        string    `json:"host"`
	IP          string    `json:"ip"`
	LatencyMs   float64   `json:"latency_ms"`
	PacketLoss  float64   `json:"packet_loss_percent"`
	Reachable   bool      `json:"reachable"`
	LastChecked time.Time `json:"last_checked"`
}

// SystemInfo contains system-level network information
type SystemInfo struct {
	Hostname          string       `json:"hostname"`
	Platform          string       `json:"platform"`
	KernelVersion     string       `json:"kernel_version,omitempty"`
	DefaultGateway    string       `json:"default_gateway"`
	DNSServers        []string     `json:"dns_servers"`
	RoutingTable      []RouteEntry `json:"routing_table"`
	NetworkNamespaces []string     `json:"network_namespaces,omitempty"`
	FirewallStatus    string       `json:"firewall_status,omitempty"`
}

// RouteEntry represents a routing table entry
type RouteEntry struct {
	Destination string `json:"destination"`
	Gateway     string `json:"gateway"`
	Interface   string `json:"interface"`
	Metric      int    `json:"metric"`
}

// ReportGenerator generates comprehensive network reports
type ReportGenerator struct {
	bandwidthCalc   *BandwidthCalculator
	speedDetector   *InterfaceSpeedDetector
	latencyTester   *LatencyTester
	systemCollector *SystemInfoCollector
}

// NewReportGenerator creates a new report generator
func NewReportGenerator() *ReportGenerator {
	return &ReportGenerator{
		bandwidthCalc:   NewBandwidthCalculator(5 * time.Second),
		speedDetector:   NewInterfaceSpeedDetector(),
		latencyTester:   NewLatencyTester(),
		systemCollector: NewSystemInfoCollector(),
	}
}

// GenerateReport creates a comprehensive network report
func (rg *ReportGenerator) GenerateReport(duration time.Duration) (*ComprehensiveNetworkReport, error) {
	startTime := time.Now()

	// Collect system information
	systemInfo, err := rg.systemCollector.CollectSystemInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to collect system info: %w", err)
	}

	// Get interface list and basic info
	interfaces, err := rg.collectInterfaceReports()
	if err != nil {
		return nil, fmt.Errorf("failed to collect interface reports: %w", err)
	}

	// Collect bandwidth trends
	bandwidthTrends := rg.bandwidthCalc.GetBandwidthTrends()

	// Perform latency tests
	latencyReport, err := rg.latencyTester.RunLatencyTests([]string{
		"8.8.8.8",                 // Google DNS
		"1.1.1.1",                 // Cloudflare DNS
		"208.67.222.222",          // OpenDNS
		systemInfo.DefaultGateway, // Local gateway
	})
	if err != nil {
		// Don't fail the entire report for latency issues
		latencyReport = &LatencyReport{}
	}

	// Generate summary
	summary := rg.generateSummary(interfaces, bandwidthTrends, latencyReport)

	// Generate recommendations
	recommendations := rg.generateRecommendations(summary, interfaces, bandwidthTrends, latencyReport)

	report := &ComprehensiveNetworkReport{
		Summary:         summary,
		Interfaces:      interfaces,
		BandwidthTrends: bandwidthTrends,
		LatencyMetrics:  *latencyReport,
		Recommendations: recommendations,
		Timestamp:       startTime,
		Duration:        time.Since(startTime),
		SystemInfo:      *systemInfo,
	}

	return report, nil
}

// collectInterfaceReports gathers detailed information about all network interfaces
func (rg *ReportGenerator) collectInterfaceReports() ([]InterfaceReport, error) {
	// Get list of network interfaces (implementation would depend on OS)
	interfaceNames, err := rg.getNetworkInterfaceNames()
	if err != nil {
		return nil, fmt.Errorf("failed to get interface names: %w", err)
	}

	reports := make([]InterfaceReport, 0, len(interfaceNames))
	recentMeasurements := rg.bandwidthCalc.GetRecentMeasurements()

	for _, name := range interfaceNames {
		// Get interface info
		info, err := rg.speedDetector.GetInterfaceInfo(name)
		if err != nil {
			// Skip interfaces that can't be queried
			continue
		}

		report := InterfaceReport{
			Name:        info.Name,
			Status:      info.Status,
			Type:        info.Type,
			MaxSpeed:    info.MaxSpeed,
			MaxSpeedStr: info.MaxSpeedStr,
			MTU:         info.MTU,
			Driver:      info.Driver,
		}

		// Add current bandwidth measurements if available
		if measurement, exists := recentMeasurements[name]; exists {
			report.CurrentRxRate = measurement.RxBytesPerSec
			report.CurrentTxRate = measurement.TxBytesPerSec
			report.Utilization = measurement.Utilization
		}

		reports = append(reports, report)
	}

	// Sort interfaces by name for consistent output
	sort.Slice(reports, func(i, j int) bool {
		return reports[i].Name < reports[j].Name
	})

	return reports, nil
}

// generateSummary creates a network summary from collected data
func (rg *ReportGenerator) generateSummary(interfaces []InterfaceReport, trends map[string]BandwidthTrend, latency *LatencyReport) NetworkSummary {
	var totalBandwidth, usedBandwidth uint64
	activeInterfaces := 0
	var topInterface string
	var maxUtilization float64
	var peakTime time.Time

	for _, iface := range interfaces {
		totalBandwidth += iface.MaxSpeed
		usedBandwidth += iface.CurrentRxRate + iface.CurrentTxRate

		if iface.Status == "up" {
			activeInterfaces++
		}

		if iface.Utilization > maxUtilization {
			maxUtilization = iface.Utilization
			topInterface = iface.Name
		}
	}

	// Find peak usage time from trends
	for _, trend := range trends {
		for _, measurement := range trend.Measurements {
			if measurement.Utilization > maxUtilization {
				peakTime = measurement.Timestamp
			}
		}
	}

	utilization := float64(0)
	if totalBandwidth > 0 {
		utilization = (float64(usedBandwidth) / float64(totalBandwidth)) * 100
	}

	return NetworkSummary{
		TotalInterfaces:    len(interfaces),
		ActiveInterfaces:   activeInterfaces,
		TotalBandwidth:     totalBandwidth,
		UsedBandwidth:      usedBandwidth,
		UtilizationPercent: utilization,
		AverageLatency:     latency.AverageLatency,
		PacketLossPercent:  latency.PacketLoss,
		TopInterface:       topInterface,
		PeakUsageTimestamp: peakTime,
	}
}

// generateRecommendations creates actionable recommendations based on network analysis
func (rg *ReportGenerator) generateRecommendations(summary NetworkSummary, interfaces []InterfaceReport, trends map[string]BandwidthTrend, latency *LatencyReport) []string {
	var recommendations []string

	// High utilization warning
	if summary.UtilizationPercent > 80 {
		recommendations = append(recommendations,
			fmt.Sprintf("⚠️ High network utilization (%.1f%%). Consider upgrading network capacity or optimizing traffic.", summary.UtilizationPercent))
	}

	// Interface-specific recommendations
	for _, iface := range interfaces {
		if iface.Status != "up" && iface.Type != "loopback" {
			recommendations = append(recommendations,
				fmt.Sprintf("🔧 Interface %s is down. Check physical connection and configuration.", iface.Name))
		}

		if iface.Utilization > 90 {
			recommendations = append(recommendations,
				fmt.Sprintf("⚡ Interface %s is heavily utilized (%.1f%%). Monitor for congestion.", iface.Name, iface.Utilization))
		}

		if iface.ErrorsRx > 0 || iface.ErrorsTx > 0 {
			recommendations = append(recommendations,
				fmt.Sprintf("❌ Interface %s has packet errors (RX: %d, TX: %d). Check for hardware issues.", iface.Name, iface.ErrorsRx, iface.ErrorsTx))
		}
	}

	// Latency recommendations
	if latency.AverageLatency > 100 {
		recommendations = append(recommendations,
			fmt.Sprintf("🐌 High latency detected (%.1fms average). Check network path and DNS resolution.", latency.AverageLatency))
	}

	if latency.PacketLoss > 1 {
		recommendations = append(recommendations,
			fmt.Sprintf("📉 Packet loss detected (%.1f%%). Investigate network stability and routing.", latency.PacketLoss))
	}

	// Bandwidth trend analysis
	for interfaceName, trend := range trends {
		if trend.PeakUtil > 95 {
			recommendations = append(recommendations,
				fmt.Sprintf("📈 Interface %s reached peak utilization of %.1f%%. Consider load balancing or capacity planning.", interfaceName, trend.PeakUtil))
		}

		// Check for consistent high usage
		if trend.AverageUtil > 70 {
			recommendations = append(recommendations,
				fmt.Sprintf("📊 Interface %s shows sustained high usage (%.1f%% average). Monitor for capacity needs.", interfaceName, trend.AverageUtil))
		}
	}

	// General recommendations
	if len(recommendations) == 0 {
		recommendations = append(recommendations, "✅ Network appears to be operating normally. No immediate issues detected.")
	}

	return recommendations
}

// getNetworkInterfaceNames returns a list of network interface names
// This is a placeholder - actual implementation would be OS-specific
func (rg *ReportGenerator) getNetworkInterfaceNames() ([]string, error) {
	// This would typically use net.Interfaces() or system-specific calls
	// For now, return common interface names
	return []string{"eth0", "wlan0", "lo"}, nil
}
