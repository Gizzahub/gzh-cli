package reports

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// TextReportGenerator generates text-based network reports
type TextReportGenerator struct {
	outputDir string
}

// NewTextReportGenerator creates a new text report generator
func NewTextReportGenerator(outputDir string) *TextReportGenerator {
	return &TextReportGenerator{
		outputDir: outputDir,
	}
}

// GenerateReport generates a text report from comprehensive network data
func (trg *TextReportGenerator) GenerateReport(report *ComprehensiveNetworkReport, filename string) error {
	// Ensure output directory exists
	if err := os.MkdirAll(trg.outputDir, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	var buf strings.Builder

	// Header
	trg.writeHeader(&buf, report)

	// Summary section
	trg.writeSummary(&buf, report.Summary)

	// Interface details
	trg.writeInterfaces(&buf, report.Interfaces)

	// Bandwidth trends
	if len(report.BandwidthTrends) > 0 {
		trg.writeBandwidthTrends(&buf, report.BandwidthTrends)
	}

	// Latency metrics
	if len(report.LatencyMetrics.Targets) > 0 {
		trg.writeLatencyMetrics(&buf, &report.LatencyMetrics)
	}

	// System information
	trg.writeSystemInfo(&buf, &report.SystemInfo)

	// Recommendations
	if len(report.Recommendations) > 0 {
		trg.writeRecommendations(&buf, report.Recommendations)
	}

	// Footer
	trg.writeFooter(&buf, report)

	// Write to file
	outputPath := filepath.Join(trg.outputDir, filename)
	return os.WriteFile(outputPath, []byte(buf.String()), 0o600) // G306: 보안 강화된 파일 권한
}

// writeHeader writes the report header
func (trg *TextReportGenerator) writeHeader(buf *strings.Builder, report *ComprehensiveNetworkReport) {
	buf.WriteString("📊 NETWORK METRICS REPORT\n")
	buf.WriteString(strings.Repeat("=", 80) + "\n\n")
	buf.WriteString(fmt.Sprintf("Generated: %s\n", report.Timestamp.Format("2006-01-02 15:04:05 MST")))
	buf.WriteString(fmt.Sprintf("Duration: %v\n", report.Duration))
	buf.WriteString(fmt.Sprintf("Host: %s (%s)\n", report.SystemInfo.Hostname, report.SystemInfo.Platform))
	if report.SystemInfo.KernelVersion != "" {
		buf.WriteString(fmt.Sprintf("Kernel: %s\n", report.SystemInfo.KernelVersion))
	}
	buf.WriteString("\n")
}

// writeSummary writes the network summary section
func (trg *TextReportGenerator) writeSummary(buf *strings.Builder, summary NetworkSummary) {
	buf.WriteString("📈 NETWORK SUMMARY\n")
	buf.WriteString(strings.Repeat("-", 40) + "\n")

	// Interface summary
	buf.WriteString(fmt.Sprintf("Interfaces: %d total, %d active\n",
		summary.TotalInterfaces, summary.ActiveInterfaces))

	// Bandwidth summary
	buf.WriteString(fmt.Sprintf("Total Bandwidth: %s/s\n", trg.formatBytes(summary.TotalBandwidth)))
	buf.WriteString(fmt.Sprintf("Used Bandwidth: %s/s (%.1f%% utilization)\n",
		trg.formatBytes(summary.UsedBandwidth), summary.UtilizationPercent))

	// Quality metrics
	buf.WriteString(fmt.Sprintf("Average Latency: %.2f ms\n", summary.AverageLatency))
	buf.WriteString(fmt.Sprintf("Packet Loss: %.2f%%\n", summary.PacketLossPercent))

	if summary.TopInterface != "" {
		buf.WriteString(fmt.Sprintf("Top Interface: %s\n", summary.TopInterface))
	}

	buf.WriteString("\n")
}

// writeInterfaces writes the interface details section
func (trg *TextReportGenerator) writeInterfaces(buf *strings.Builder, interfaces []InterfaceReport) {
	buf.WriteString("🔌 INTERFACE DETAILS\n")
	buf.WriteString(strings.Repeat("-", 40) + "\n")

	// Header
	buf.WriteString(fmt.Sprintf("%-12s %-8s %-10s %-12s %-12s %-12s %-8s\n",
		"Interface", "Status", "Type", "Speed", "RX Rate", "TX Rate", "Usage%"))
	buf.WriteString(strings.Repeat("-", 80) + "\n")

	// Interface rows
	for _, iface := range interfaces {
		buf.WriteString(fmt.Sprintf("%-12s %-8s %-10s %-12s %-12s %-12s %-8.1f\n",
			iface.Name,
			iface.Status,
			iface.Type,
			iface.MaxSpeedStr,
			trg.formatBytes(iface.CurrentRxRate)+"/s",
			trg.formatBytes(iface.CurrentTxRate)+"/s",
			iface.Utilization))
	}
	buf.WriteString("\n")
}

// writeBandwidthTrends writes the bandwidth trends section
func (trg *TextReportGenerator) writeBandwidthTrends(buf *strings.Builder, trends map[string]BandwidthTrend) {
	buf.WriteString("📊 BANDWIDTH TRENDS\n")
	buf.WriteString(strings.Repeat("-", 40) + "\n")

	for interfaceName, trend := range trends {
		buf.WriteString(fmt.Sprintf("Interface: %s\n", interfaceName))
		buf.WriteString(fmt.Sprintf("  Average RX: %s/s\n", trg.formatBytes(trend.AverageRxRate)))
		buf.WriteString(fmt.Sprintf("  Average TX: %s/s\n", trg.formatBytes(trend.AverageTxRate)))
		buf.WriteString(fmt.Sprintf("  Peak RX: %s/s\n", trg.formatBytes(trend.PeakRxRate)))
		buf.WriteString(fmt.Sprintf("  Peak TX: %s/s\n", trg.formatBytes(trend.PeakTxRate)))
		buf.WriteString(fmt.Sprintf("  Average Utilization: %.1f%%\n", trend.AverageUtil))
		buf.WriteString(fmt.Sprintf("  Peak Utilization: %.1f%%\n", trend.PeakUtil))
		buf.WriteString(fmt.Sprintf("  Measurements: %d\n", len(trend.Measurements)))
		buf.WriteString("\n")
	}
}

// writeLatencyMetrics writes the latency metrics section
func (trg *TextReportGenerator) writeLatencyMetrics(buf *strings.Builder, latency *LatencyReport) {
	buf.WriteString("📡 LATENCY METRICS\n")
	buf.WriteString(strings.Repeat("-", 40) + "\n")

	// Overall metrics
	buf.WriteString(fmt.Sprintf("Average Latency: %.2f ms\n", latency.AverageLatency))
	buf.WriteString(fmt.Sprintf("Min Latency: %.2f ms\n", latency.MinLatency))
	buf.WriteString(fmt.Sprintf("Max Latency: %.2f ms\n", latency.MaxLatency))
	buf.WriteString(fmt.Sprintf("Jitter: %.2f ms\n", latency.JitterAverage))
	buf.WriteString(fmt.Sprintf("Packet Loss: %.2f%%\n", latency.PacketLoss))
	buf.WriteString(fmt.Sprintf("Reachability: %.1f%%\n", latency.ReachabilityRate))
	buf.WriteString("\n")

	// Individual targets
	buf.WriteString("Target Details:\n")
	buf.WriteString(fmt.Sprintf("%-20s %-15s %-10s %-12s %-10s\n",
		"Host", "IP", "Status", "Latency", "Loss%"))
	buf.WriteString(strings.Repeat("-", 70) + "\n")

	for _, target := range latency.Targets {
		status := "Reachable"
		if !target.Reachable {
			status = "Failed"
		}

		latencyStr := "N/A"
		lossStr := "N/A"
		if target.Reachable {
			latencyStr = fmt.Sprintf("%.1f ms", target.LatencyMs)
			lossStr = fmt.Sprintf("%.1f%%", target.PacketLoss)
		}

		buf.WriteString(fmt.Sprintf("%-20s %-15s %-10s %-12s %-10s\n",
			trg.truncateString(target.Host, 20),
			target.IP,
			status,
			latencyStr,
			lossStr))
	}
	buf.WriteString("\n")
}

// writeSystemInfo writes the system information section
func (trg *TextReportGenerator) writeSystemInfo(buf *strings.Builder, sysInfo *SystemInfo) {
	buf.WriteString("🖥️ SYSTEM INFORMATION\n")
	buf.WriteString(strings.Repeat("-", 40) + "\n")

	buf.WriteString(fmt.Sprintf("Hostname: %s\n", sysInfo.Hostname))
	buf.WriteString(fmt.Sprintf("Platform: %s\n", sysInfo.Platform))
	if sysInfo.KernelVersion != "" {
		buf.WriteString(fmt.Sprintf("Kernel Version: %s\n", sysInfo.KernelVersion))
	}
	buf.WriteString(fmt.Sprintf("Default Gateway: %s\n", sysInfo.DefaultGateway))

	if len(sysInfo.DNSServers) > 0 {
		buf.WriteString(fmt.Sprintf("DNS Servers: %s\n", strings.Join(sysInfo.DNSServers, ", ")))
	}

	if sysInfo.FirewallStatus != "" {
		buf.WriteString(fmt.Sprintf("Firewall: %s\n", sysInfo.FirewallStatus))
	}

	if len(sysInfo.NetworkNamespaces) > 0 {
		buf.WriteString(fmt.Sprintf("Network Namespaces: %s\n", strings.Join(sysInfo.NetworkNamespaces, ", ")))
	}

	// Routing table summary
	if len(sysInfo.RoutingTable) > 0 {
		buf.WriteString("\nRouting Table (Top 10):\n")
		buf.WriteString(fmt.Sprintf("%-18s %-15s %-10s %-8s\n", "Destination", "Gateway", "Interface", "Metric"))
		buf.WriteString(strings.Repeat("-", 55) + "\n")

		count := 0
		for _, route := range sysInfo.RoutingTable {
			if count >= 10 { // Limit to top 10 routes
				break
			}
			buf.WriteString(fmt.Sprintf("%-18s %-15s %-10s %-8d\n",
				trg.truncateString(route.Destination, 18),
				trg.truncateString(route.Gateway, 15),
				trg.truncateString(route.Interface, 10),
				route.Metric))
			count++
		}

		if len(sysInfo.RoutingTable) > 10 {
			buf.WriteString(fmt.Sprintf("... and %d more routes\n", len(sysInfo.RoutingTable)-10))
		}
	}

	buf.WriteString("\n")
}

// writeRecommendations writes the recommendations section
func (trg *TextReportGenerator) writeRecommendations(buf *strings.Builder, recommendations []string) {
	buf.WriteString("💡 OPTIMIZATION RECOMMENDATIONS\n")
	buf.WriteString(strings.Repeat("-", 40) + "\n")

	for i, rec := range recommendations {
		buf.WriteString(fmt.Sprintf("%d. %s\n", i+1, rec))
	}
	buf.WriteString("\n")
}

// writeFooter writes the report footer
func (trg *TextReportGenerator) writeFooter(buf *strings.Builder, report *ComprehensiveNetworkReport) {
	buf.WriteString(strings.Repeat("=", 80) + "\n")
	buf.WriteString("🤖 Generated by gzh-cli Network Metrics\n")
	buf.WriteString(fmt.Sprintf("Report completed at %s\n", time.Now().Format("2006-01-02 15:04:05 MST")))
	buf.WriteString(fmt.Sprintf("Generation took: %v\n", report.Duration))
	buf.WriteString(strings.Repeat("=", 80) + "\n")
}

// formatBytes formats bytes in human-readable format
func (trg *TextReportGenerator) formatBytes(bytes uint64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/TB)
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// truncateString truncates a string to the specified length
func (trg *TextReportGenerator) truncateString(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length-3] + "..."
}
