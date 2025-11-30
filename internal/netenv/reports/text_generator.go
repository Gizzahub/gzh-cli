// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package reports

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// TextReportGenerator generates text-based network reports.
type TextReportGenerator struct {
	outputDir string
}

// NewTextReportGenerator creates a new text report generator.
func NewTextReportGenerator(outputDir string) *TextReportGenerator {
	return &TextReportGenerator{
		outputDir: outputDir,
	}
}

// GenerateReport generates a text report from comprehensive network data.
func (trg *TextReportGenerator) GenerateReport(report *ComprehensiveNetworkReport, filename string) error {
	// Ensure output directory exists
	if err := os.MkdirAll(trg.outputDir, 0o750); err != nil {
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

// writeHeader writes the report header.
func (trg *TextReportGenerator) writeHeader(buf *strings.Builder, report *ComprehensiveNetworkReport) {
	buf.WriteString("📊 NETWORK METRICS REPORT\n")
	buf.WriteString(strings.Repeat("=", 80) + "\n\n")
	fmt.Fprintf(buf, "Generated: %s\n", report.Timestamp.Format("2006-01-02 15:04:05 MST"))
	fmt.Fprintf(buf, "Duration: %v\n", report.Duration)
	fmt.Fprintf(buf, "Host: %s (%s)\n", report.SystemInfo.Hostname, report.SystemInfo.Platform)
	if report.SystemInfo.KernelVersion != "" {
		fmt.Fprintf(buf, "Kernel: %s\n", report.SystemInfo.KernelVersion)
	}
	buf.WriteString("\n")
}

// writeSummary writes the network summary section.
func (trg *TextReportGenerator) writeSummary(buf *strings.Builder, summary NetworkSummary) {
	buf.WriteString("📈 NETWORK SUMMARY\n")
	buf.WriteString(strings.Repeat("-", 40) + "\n")

	// Interface summary
	fmt.Fprintf(buf, "Interfaces: %d total, %d active\n",
		summary.TotalInterfaces, summary.ActiveInterfaces)

	// Bandwidth summary
	fmt.Fprintf(buf, "Total Bandwidth: %s/s\n", trg.formatBytes(summary.TotalBandwidth))
	fmt.Fprintf(buf, "Used Bandwidth: %s/s (%.1f%% utilization)\n",
		trg.formatBytes(summary.UsedBandwidth), summary.UtilizationPercent)

	// Quality metrics
	fmt.Fprintf(buf, "Average Latency: %.2f ms\n", summary.AverageLatency)
	fmt.Fprintf(buf, "Packet Loss: %.2f%%\n", summary.PacketLossPercent)

	if summary.TopInterface != "" {
		fmt.Fprintf(buf, "Top Interface: %s\n", summary.TopInterface)
	}

	buf.WriteString("\n")
}

// writeInterfaces writes the interface details section.
func (trg *TextReportGenerator) writeInterfaces(buf *strings.Builder, interfaces []InterfaceReport) {
	buf.WriteString("🔌 INTERFACE DETAILS\n")
	buf.WriteString(strings.Repeat("-", 40) + "\n")

	// Header
	fmt.Fprintf(buf, "%-12s %-8s %-10s %-12s %-12s %-12s %-8s\n",
		"Interface", "Status", "Type", "Speed", "RX Rate", "TX Rate", "Usage%")
	buf.WriteString(strings.Repeat("-", 80) + "\n")

	// Interface rows
	for _, iface := range interfaces {
		fmt.Fprintf(buf, "%-12s %-8s %-10s %-12s %-12s %-12s %-8.1f\n",
			iface.Name,
			iface.Status,
			iface.Type,
			iface.MaxSpeedStr,
			trg.formatBytes(iface.CurrentRxRate)+"/s",
			trg.formatBytes(iface.CurrentTxRate)+"/s",
			iface.Utilization)
	}
	buf.WriteString("\n")
}

// writeBandwidthTrends writes the bandwidth trends section.
func (trg *TextReportGenerator) writeBandwidthTrends(buf *strings.Builder, trends map[string]BandwidthTrend) {
	buf.WriteString("📊 BANDWIDTH TRENDS\n")
	buf.WriteString(strings.Repeat("-", 40) + "\n")

	for interfaceName, trend := range trends {
		fmt.Fprintf(buf, "Interface: %s\n", interfaceName)
		fmt.Fprintf(buf, "  Average RX: %s/s\n", trg.formatBytes(trend.AverageRxRate))
		fmt.Fprintf(buf, "  Average TX: %s/s\n", trg.formatBytes(trend.AverageTxRate))
		fmt.Fprintf(buf, "  Peak RX: %s/s\n", trg.formatBytes(trend.PeakRxRate))
		fmt.Fprintf(buf, "  Peak TX: %s/s\n", trg.formatBytes(trend.PeakTxRate))
		fmt.Fprintf(buf, "  Average Utilization: %.1f%%\n", trend.AverageUtil)
		fmt.Fprintf(buf, "  Peak Utilization: %.1f%%\n", trend.PeakUtil)
		fmt.Fprintf(buf, "  Measurements: %d\n", len(trend.Measurements))
		buf.WriteString("\n")
	}
}

// writeLatencyMetrics writes the latency metrics section.
func (trg *TextReportGenerator) writeLatencyMetrics(buf *strings.Builder, latency *LatencyReport) {
	buf.WriteString("📡 LATENCY METRICS\n")
	buf.WriteString(strings.Repeat("-", 40) + "\n")

	// Overall metrics
	fmt.Fprintf(buf, "Average Latency: %.2f ms\n", latency.AverageLatency)
	fmt.Fprintf(buf, "Min Latency: %.2f ms\n", latency.MinLatency)
	fmt.Fprintf(buf, "Max Latency: %.2f ms\n", latency.MaxLatency)
	fmt.Fprintf(buf, "Jitter: %.2f ms\n", latency.JitterAverage)
	fmt.Fprintf(buf, "Packet Loss: %.2f%%\n", latency.PacketLoss)
	fmt.Fprintf(buf, "Reachability: %.1f%%\n", latency.ReachabilityRate)
	buf.WriteString("\n")

	// Individual targets
	buf.WriteString("Target Details:\n")
	fmt.Fprintf(buf, "%-20s %-15s %-10s %-12s %-10s\n",
		"Host", "IP", "Status", "Latency", "Loss%")
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

		fmt.Fprintf(buf, "%-20s %-15s %-10s %-12s %-10s\n",
			trg.truncateString(target.Host, 20),
			target.IP,
			status,
			latencyStr,
			lossStr)
	}
	buf.WriteString("\n")
}

// writeSystemInfo writes the system information section.
func (trg *TextReportGenerator) writeSystemInfo(buf *strings.Builder, sysInfo *SystemInfo) {
	buf.WriteString("🖥️ SYSTEM INFORMATION\n")
	buf.WriteString(strings.Repeat("-", 40) + "\n")

	fmt.Fprintf(buf, "Hostname: %s\n", sysInfo.Hostname)
	fmt.Fprintf(buf, "Platform: %s\n", sysInfo.Platform)
	if sysInfo.KernelVersion != "" {
		fmt.Fprintf(buf, "Kernel Version: %s\n", sysInfo.KernelVersion)
	}
	fmt.Fprintf(buf, "Default Gateway: %s\n", sysInfo.DefaultGateway)

	if len(sysInfo.DNSServers) > 0 {
		fmt.Fprintf(buf, "DNS Servers: %s\n", strings.Join(sysInfo.DNSServers, ", "))
	}

	if sysInfo.FirewallStatus != "" {
		fmt.Fprintf(buf, "Firewall: %s\n", sysInfo.FirewallStatus)
	}

	if len(sysInfo.NetworkNamespaces) > 0 {
		fmt.Fprintf(buf, "Network Namespaces: %s\n", strings.Join(sysInfo.NetworkNamespaces, ", "))
	}

	// Routing table summary
	if len(sysInfo.RoutingTable) > 0 {
		buf.WriteString("\nRouting Table (Top 10):\n")
		fmt.Fprintf(buf, "%-18s %-15s %-10s %-8s\n", "Destination", "Gateway", "Interface", "Metric")
		buf.WriteString(strings.Repeat("-", 55) + "\n")

		count := 0
		for _, route := range sysInfo.RoutingTable {
			if count >= 10 { // Limit to top 10 routes
				break
			}
			fmt.Fprintf(buf, "%-18s %-15s %-10s %-8d\n",
				trg.truncateString(route.Destination, 18),
				trg.truncateString(route.Gateway, 15),
				trg.truncateString(route.Interface, 10),
				route.Metric)
			count++
		}

		if len(sysInfo.RoutingTable) > 10 {
			fmt.Fprintf(buf, "... and %d more routes\n", len(sysInfo.RoutingTable)-10)
		}
	}

	buf.WriteString("\n")
}

// writeRecommendations writes the recommendations section.
func (trg *TextReportGenerator) writeRecommendations(buf *strings.Builder, recommendations []string) {
	buf.WriteString("💡 OPTIMIZATION RECOMMENDATIONS\n")
	buf.WriteString(strings.Repeat("-", 40) + "\n")

	for i, rec := range recommendations {
		fmt.Fprintf(buf, "%d. %s\n", i+1, rec)
	}
	buf.WriteString("\n")
}

// writeFooter writes the report footer.
func (trg *TextReportGenerator) writeFooter(buf *strings.Builder, report *ComprehensiveNetworkReport) {
	buf.WriteString(strings.Repeat("=", 80) + "\n")
	buf.WriteString("🤖 Generated by gzh-cli Network Metrics\n")
	fmt.Fprintf(buf, "Report completed at %s\n", time.Now().Format("2006-01-02 15:04:05 MST"))
	fmt.Fprintf(buf, "Generation took: %v\n", report.Duration)
	buf.WriteString(strings.Repeat("=", 80) + "\n")
}

// formatBytes formats bytes in human-readable format.
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

// truncateString truncates a string to the specified length.
func (trg *TextReportGenerator) truncateString(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length-3] + "..."
}
