package reports

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
)

// HTMLReportGenerator generates HTML network reports
type HTMLReportGenerator struct {
	outputDir string
}

// NewHTMLReportGenerator creates a new HTML report generator
func NewHTMLReportGenerator(outputDir string) *HTMLReportGenerator {
	return &HTMLReportGenerator{
		outputDir: outputDir,
	}
}

// GenerateReport generates an HTML report from comprehensive network data
func (hrg *HTMLReportGenerator) GenerateReport(report *ComprehensiveNetworkReport, filename string) error {
	// Ensure output directory exists
	if err := os.MkdirAll(hrg.outputDir, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Parse HTML template
	tmpl, err := template.New("report").Funcs(template.FuncMap{
		"formatBytes":    hrg.formatBytes,
		"formatDuration": hrg.formatDuration,
		"printf":         fmt.Sprintf,
		"join":           strings.Join,
		"truncate":       hrg.truncateString,
	}).Parse(htmlReportTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse HTML template: %w", err)
	}

	// Create output file
	outputPath := filepath.Join(hrg.outputDir, filename)
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create HTML report file: %w", err)
	}
	defer file.Close()

	// Execute template with report data
	if err := tmpl.Execute(file, report); err != nil {
		return fmt.Errorf("failed to execute HTML template: %w", err)
	}

	return nil
}

// formatBytes formats bytes in human-readable format
func (hrg *HTMLReportGenerator) formatBytes(bytes uint64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
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

// formatDuration formats duration in human-readable format
func (hrg *HTMLReportGenerator) formatDuration(d interface{}) string {
	switch dur := d.(type) {
	case int64:
		return fmt.Sprintf("%d ms", dur)
	case float64:
		return fmt.Sprintf("%.1f ms", dur)
	default:
		return fmt.Sprintf("%v", dur)
	}
}

// truncateString truncates string to specified length
func (hrg *HTMLReportGenerator) truncateString(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length] + "..."
}

// htmlReportTemplate is the HTML template for network reports
const htmlReportTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Network Metrics Report - {{.Timestamp.Format "2006-01-02 15:04:05"}}</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            line-height: 1.6;
            color: #333;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
        }

        .container {
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
        }

        .header {
            background: rgba(255, 255, 255, 0.95);
            backdrop-filter: blur(10px);
            border-radius: 15px;
            padding: 30px;
            margin-bottom: 30px;
            box-shadow: 0 8px 32px rgba(31, 38, 135, 0.37);
            text-align: center;
        }

        .header h1 {
            color: #2c3e50;
            margin-bottom: 10px;
            font-size: 2.5em;
        }

        .header .meta {
            color: #7f8c8d;
            font-size: 1.1em;
        }

        .summary-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }

        .summary-card {
            background: rgba(255, 255, 255, 0.9);
            border-radius: 12px;
            padding: 25px;
            box-shadow: 0 4px 15px rgba(0, 0, 0, 0.1);
            transition: transform 0.3s ease;
        }

        .summary-card:hover {
            transform: translateY(-5px);
        }

        .summary-card h3 {
            color: #2c3e50;
            margin-bottom: 15px;
            font-size: 1.3em;
            display: flex;
            align-items: center;
            gap: 10px;
        }

        .summary-card .value {
            font-size: 2em;
            font-weight: bold;
            margin: 10px 0;
        }

        .summary-card .label {
            color: #7f8c8d;
            font-size: 0.9em;
        }

        .good { color: #27ae60; }
        .warning { color: #f39c12; }
        .danger { color: #e74c3c; }

        .section {
            background: rgba(255, 255, 255, 0.95);
            border-radius: 12px;
            padding: 25px;
            margin-bottom: 25px;
            box-shadow: 0 4px 15px rgba(0, 0, 0, 0.1);
        }

        .section h2 {
            color: #2c3e50;
            margin-bottom: 20px;
            font-size: 1.6em;
            border-bottom: 2px solid #3498db;
            padding-bottom: 10px;
        }

        .interface-grid {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
            gap: 20px;
        }

        .interface-card {
            border: 1px solid #ddd;
            border-radius: 8px;
            padding: 20px;
            background: #f8f9fa;
        }

        .interface-header {
            display: flex;
            justify-content: between;
            align-items: center;
            margin-bottom: 15px;
        }

        .interface-name {
            font-size: 1.2em;
            font-weight: bold;
            color: #2c3e50;
        }

        .interface-status {
            padding: 4px 12px;
            border-radius: 20px;
            font-size: 0.8em;
            font-weight: bold;
            text-transform: uppercase;
        }

        .status-up {
            background: #d4edda;
            color: #155724;
        }

        .status-down {
            background: #f8d7da;
            color: #721c24;
        }

        .interface-metrics {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 15px;
        }

        .metric {
            text-align: center;
        }

        .metric-value {
            font-size: 1.4em;
            font-weight: bold;
            margin-bottom: 5px;
        }

        .metric-label {
            color: #6c757d;
            font-size: 0.9em;
        }

        .recommendations {
            background: linear-gradient(135deg, #e8f4f8 0%, #f0f8e8 100%);
            border-left: 4px solid #17a2b8;
            padding: 20px;
            border-radius: 8px;
        }

        .recommendation-item {
            margin-bottom: 15px;
            padding-bottom: 15px;
            border-bottom: 1px solid rgba(0,0,0,0.1);
        }

        .recommendation-item:last-child {
            margin-bottom: 0;
            padding-bottom: 0;
            border-bottom: none;
        }

        .system-info {
            background: #f8f9fa;
            border-radius: 8px;
            padding: 20px;
            margin-top: 20px;
        }

        .system-info-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 15px;
        }

        .system-info-item {
            display: flex;
            flex-direction: column;
        }

        .system-info-label {
            font-weight: bold;
            color: #495057;
            margin-bottom: 5px;
        }

        .system-info-value {
            color: #6c757d;
        }

        .chart-container {
            background: white;
            border-radius: 8px;
            padding: 20px;
            margin: 20px 0;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }

        .progress-bar {
            width: 100%;
            height: 20px;
            background: #e9ecef;
            border-radius: 10px;
            overflow: hidden;
            margin: 10px 0;
        }

        .progress-fill {
            height: 100%;
            background: linear-gradient(90deg, #28a745 0%, #ffc107 70%, #dc3545 100%);
            transition: width 0.3s ease;
        }

        .footer {
            text-align: center;
            color: rgba(255, 255, 255, 0.8);
            margin-top: 30px;
            padding: 20px;
        }

        @media (max-width: 768px) {
            .container {
                padding: 10px;
            }

            .summary-grid {
                grid-template-columns: 1fr;
            }

            .interface-grid {
                grid-template-columns: 1fr;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <!-- Header -->
        <div class="header">
            <h1>🌐 Network Metrics Report</h1>
            <div class="meta">
                <p>Generated: {{.Timestamp.Format "2006-01-02 15:04:05"}}</p>
                <p>Duration: {{.Duration}}</p>
                <p>Host: {{.SystemInfo.Hostname}} ({{.SystemInfo.Platform}})</p>
            </div>
        </div>

        <!-- Summary Cards -->
        <div class="summary-grid">
            <div class="summary-card">
                <h3>📊 Interface Summary</h3>
                <div class="value">{{.Summary.TotalInterfaces}}</div>
                <div class="label">Total Interfaces</div>
                <div class="value {{if lt .Summary.ActiveInterfaces .Summary.TotalInterfaces}}warning{{else}}good{{end}}">
                    {{.Summary.ActiveInterfaces}}
                </div>
                <div class="label">Active Interfaces</div>
            </div>

            <div class="summary-card">
                <h3>📈 Bandwidth Usage</h3>
                <div class="value">{{.Summary.TotalBandwidth | formatBytes}}/s</div>
                <div class="label">Total Capacity</div>
                <div class="value {{if gt .Summary.UtilizationPercent 80}}danger{{else if gt .Summary.UtilizationPercent 60}}warning{{else}}good{{end}}">
                    {{.Summary.UtilizationPercent | printf "%.1f"}}%
                </div>
                <div class="label">Utilization</div>
            </div>

            <div class="summary-card">
                <h3>⚡ Network Quality</h3>
                <div class="value {{if gt .Summary.AverageLatency 100}}danger{{else if gt .Summary.AverageLatency 50}}warning{{else}}good{{end}}">
                    {{.Summary.AverageLatency | printf "%.1f"}} ms
                </div>
                <div class="label">Average Latency</div>
                <div class="value {{if gt .Summary.PacketLossPercent 1}}danger{{else if gt .Summary.PacketLossPercent 0.1}}warning{{else}}good{{end}}">
                    {{.Summary.PacketLossPercent | printf "%.2f"}}%
                </div>
                <div class="label">Packet Loss</div>
            </div>
        </div>

        <!-- Network Interfaces -->
        <div class="section">
            <h2>🔌 Network Interfaces</h2>
            <div class="interface-grid">
                {{range .Interfaces}}
                <div class="interface-card">
                    <div class="interface-header">
                        <div class="interface-name">{{.Name}}</div>
                        <div class="interface-status {{if eq .Status "up"}}status-up{{else}}status-down{{end}}">
                            {{.Status}}
                        </div>
                    </div>
                    <div class="interface-metrics">
                        <div class="metric">
                            <div class="metric-value">{{.MaxSpeedStr}}</div>
                            <div class="metric-label">Max Speed</div>
                        </div>
                        <div class="metric">
                            <div class="metric-value {{if gt .Utilization 80}}danger{{else if gt .Utilization 60}}warning{{else}}good{{end}}">
                                {{.Utilization | printf "%.1f"}}%
                            </div>
                            <div class="metric-label">Utilization</div>
                        </div>
                        <div class="metric">
                            <div class="metric-value">{{.CurrentRxRate | formatBytes}}/s</div>
                            <div class="metric-label">RX Rate</div>
                        </div>
                        <div class="metric">
                            <div class="metric-value">{{.CurrentTxRate | formatBytes}}/s</div>
                            <div class="metric-label">TX Rate</div>
                        </div>
                    </div>
                    {{if gt .Utilization 0}}
                    <div class="progress-bar">
                        <div class="progress-fill" style="width: {{.Utilization | printf "%.0f"}}%"></div>
                    </div>
                    {{end}}
                </div>
                {{end}}
            </div>
        </div>

        <!-- Latency Metrics -->
        {{if .LatencyMetrics.Targets}}
        <div class="section">
            <h2>📡 Latency Test Results</h2>
            <div class="interface-grid">
                {{range .LatencyMetrics.Targets}}
                <div class="interface-card">
                    <div class="interface-header">
                        <div class="interface-name">{{.Host}}</div>
                        <div class="interface-status {{if .Reachable}}status-up{{else}}status-down{{end}}">
                            {{if .Reachable}}Reachable{{else}}Unreachable{{end}}
                        </div>
                    </div>
                    {{if .Reachable}}
                    <div class="interface-metrics">
                        <div class="metric">
                            <div class="metric-value {{if gt .LatencyMs 100}}danger{{else if gt .LatencyMs 50}}warning{{else}}good{{end}}">
                                {{.LatencyMs | printf "%.1f"}} ms
                            </div>
                            <div class="metric-label">Latency</div>
                        </div>
                        <div class="metric">
                            <div class="metric-value {{if gt .PacketLoss 1}}danger{{else if gt .PacketLoss 0}}warning{{else}}good{{end}}">
                                {{.PacketLoss | printf "%.1f"}}%
                            </div>
                            <div class="metric-label">Packet Loss</div>
                        </div>
                    </div>
                    {{end}}
                    <div class="system-info-item">
                        <div class="system-info-label">IP Address</div>
                        <div class="system-info-value">{{.IP}}</div>
                    </div>
                </div>
                {{end}}
            </div>
        </div>
        {{end}}

        <!-- Recommendations -->
        {{if .Recommendations}}
        <div class="section">
            <h2>💡 Optimization Recommendations</h2>
            <div class="recommendations">
                {{range .Recommendations}}
                <div class="recommendation-item">
                    <strong>{{. | truncate 100}}</strong>
                </div>
                {{end}}
            </div>
        </div>
        {{end}}

        <!-- System Information -->
        <div class="section">
            <h2>🖥️ System Information</h2>
            <div class="system-info-grid">
                <div class="system-info-item">
                    <div class="system-info-label">Hostname</div>
                    <div class="system-info-value">{{.SystemInfo.Hostname}}</div>
                </div>
                <div class="system-info-item">
                    <div class="system-info-label">Platform</div>
                    <div class="system-info-value">{{.SystemInfo.Platform}}</div>
                </div>
                {{if .SystemInfo.KernelVersion}}
                <div class="system-info-item">
                    <div class="system-info-label">Kernel Version</div>
                    <div class="system-info-value">{{.SystemInfo.KernelVersion}}</div>
                </div>
                {{end}}
                <div class="system-info-item">
                    <div class="system-info-label">Default Gateway</div>
                    <div class="system-info-value">{{.SystemInfo.DefaultGateway}}</div>
                </div>
                {{if .SystemInfo.DNSServers}}
                <div class="system-info-item">
                    <div class="system-info-label">DNS Servers</div>
                    <div class="system-info-value">{{.SystemInfo.DNSServers | join ", "}}</div>
                </div>
                {{end}}
            </div>
        </div>

        <div class="footer">
            <p>🤖 Generated by gzh-cli Network Metrics</p>
            <p>Report generated at {{.Timestamp.Format "2006-01-02 15:04:05 MST"}}</p>
        </div>
    </div>
</body>
</html>
`
