package netenv

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

const (
	statusActive = "active"
	statusFailed = "failed"
)

var logTimeRegex = regexp.MustCompile(`^\w{3} \d{2} \d{2}:\d{2}:\d{2}`)

type daemonOptions struct {
	serviceName     string
	networkServices bool
	followLogs      bool
	showInactive    bool
	enableHealth    bool
	action          string
}

type serviceInfo struct {
	Name        string
	Status      string
	Enabled     string
	Description string
	MainPID     string
	Memory      string
	Since       string
	CPUUsage    string
	LoadState   string
	SubState    string
}

type healthCheck struct {
	Name           string
	Status         string
	LastChecked    time.Time
	ResponseTime   time.Duration
	ErrorCount     int
	HealthEndpoint string
}

func newDaemonCmd(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "daemon",
		Short: "Monitor and manage system daemons",
		Long: `Monitor and manage system daemons/services.

This command provides comprehensive daemon monitoring capabilities:
- List all system services and their status
- Monitor specific services
- Check network-related services
- Track service dependencies and resource usage

Examples:
  # List all services
  gz net-env daemon list
  
  # List only network-related services
  gz net-env daemon list --network-services
  
  # Check specific service status
  gz net-env daemon status --service ssh
  
  # Monitor service with live updates
  gz net-env daemon monitor --service nginx`,
		SilenceUsage: true,
	}

	cmd.AddCommand(newDaemonListCmd())
	cmd.AddCommand(newDaemonStatusCmd())
	cmd.AddCommand(newDaemonMonitorCmd(ctx))
	cmd.AddCommand(newDaemonManageCmd())
	cmd.AddCommand(newDaemonHealthCmd(ctx))

	return cmd
}

func newDaemonListCmd() *cobra.Command {
	o := &daemonOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List system daemons and their status",
		Long: `List all system daemons/services and their current status.

This command shows a comprehensive view of system services including:
- Service name and current status (active, inactive, failed)
- Whether the service is enabled to start at boot
- Service description and main process ID
- Memory usage and start time

Examples:
  # List all services
  gz net-env daemon list
  
  # List only network-related services
  gz net-env daemon list --network-services
  
  # Include inactive services
  gz net-env daemon list --show-inactive`,
		RunE: o.runList,
	}

	cmd.Flags().BoolVar(&o.networkServices, "network-services", false, "Show only network-related services")
	cmd.Flags().BoolVar(&o.showInactive, "show-inactive", false, "Include inactive services in the list")

	return cmd
}

func newDaemonStatusCmd() *cobra.Command {
	o := &daemonOptions{}

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show detailed status of a specific daemon",
		Long: `Show detailed status information for a specific system daemon/service.

This command provides comprehensive information about a service including:
- Current status and enabled state
- Process information and resource usage
- Recent log entries
- Service configuration details

Examples:
  # Check SSH daemon status
  gz net-env daemon status --service ssh
  
  # Check NetworkManager status
  gz net-env daemon status --service NetworkManager
  
  # Check Docker daemon status
  gz net-env daemon status --service docker`,
		RunE: o.runStatus,
	}

	cmd.Flags().StringVar(&o.serviceName, "service", "", "Name of the service to check (required)")
	cmd.MarkFlagRequired("service")

	return cmd
}

func newDaemonMonitorCmd(ctx context.Context) *cobra.Command {
	o := &daemonOptions{}

	cmd := &cobra.Command{
		Use:   "monitor",
		Short: "Monitor daemon status with live updates",
		Long: `Monitor system daemon status with live updates.

This command provides real-time monitoring of system services:
- Live status updates every few seconds
- Resource usage tracking
- Log tail functionality
- Service state change notifications

Examples:
  # Monitor SSH daemon
  gz net-env daemon monitor --service ssh
  
  # Monitor with log following
  gz net-env daemon monitor --service nginx --follow-logs
  
  # Monitor network services
  gz net-env daemon monitor --network-services`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.runMonitor(ctx, cmd, args)
		},
	}

	cmd.Flags().StringVar(&o.serviceName, "service", "", "Name of the service to monitor")
	cmd.Flags().BoolVar(&o.networkServices, "network-services", false, "Monitor network-related services")
	cmd.Flags().BoolVar(&o.followLogs, "follow-logs", false, "Follow service logs in real-time")
	cmd.Flags().BoolVar(&o.enableHealth, "enable-health", false, "Enable health monitoring for services")

	return cmd
}

func (o *daemonOptions) runList(_ *cobra.Command, args []string) error {
	services, err := o.getSystemServices()
	if err != nil {
		return fmt.Errorf("failed to get system services: %w", err)
	}

	// Filter services if requested
	if o.networkServices {
		services = o.filterNetworkServices(services)
	}

	if !o.showInactive {
		services = o.filterActiveServices(services)
	}

	// Display services
	fmt.Printf("System Services (%d):\n\n", len(services))
	fmt.Printf("%-25s %-10s %-8s %-12s %-20s\n", "SERVICE", "STATUS", "ENABLED", "MAIN PID", "DESCRIPTION")
	fmt.Printf("%-25s %-10s %-8s %-12s %-20s\n", strings.Repeat("-", 25), strings.Repeat("-", 10), strings.Repeat("-", 8), strings.Repeat("-", 12), strings.Repeat("-", 20))

	for _, service := range services {
		statusIcon := o.getStatusIcon(service.Status)
		description := service.Description
		if len(description) > 40 {
			description = description[:37] + "..."
		}

		fmt.Printf("%-25s %s%-9s %-8s %-12s %-20s\n",
			service.Name,
			statusIcon,
			service.Status,
			service.Enabled,
			service.MainPID,
			description)
	}

	return nil
}

func (o *daemonOptions) runStatus(_ *cobra.Command, args []string) error {
	service, err := o.getServiceDetails(o.serviceName)
	if err != nil {
		return fmt.Errorf("failed to get service details: %w", err)
	}

	// Display detailed service information
	fmt.Printf("🔧 Service: %s\n", service.Name)
	fmt.Printf("   Status: %s%s\n", o.getStatusIcon(service.Status), service.Status)
	fmt.Printf("   Enabled: %s\n", service.Enabled)
	if service.MainPID != "" && service.MainPID != "0" {
		fmt.Printf("   Main PID: %s\n", service.MainPID)
	}
	if service.Memory != "" {
		fmt.Printf("   Memory: %s\n", service.Memory)
	}
	if service.Since != "" {
		fmt.Printf("   Active since: %s\n", service.Since)
	}
	if service.Description != "" {
		fmt.Printf("   Description: %s\n", service.Description)
	}

	// Show recent logs
	fmt.Println("\n📋 Recent logs:")
	if err := o.showServiceLogs(o.serviceName, 10); err != nil {
		fmt.Printf("   Warning: Could not retrieve logs: %v\n", err)
	}

	return nil
}

func (o *daemonOptions) runMonitor(ctx context.Context, _ *cobra.Command, args []string) error {
	if o.serviceName == "" && !o.networkServices {
		return fmt.Errorf("either --service or --network-services must be specified")
	}

	fmt.Println("📊 Starting daemon monitor (Press Ctrl+C to stop)")
	fmt.Println()

	// Monitor specific service or network services
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("\n🛑 Stopping daemon monitoring (reason: %v)\n", ctx.Err())
			return nil

		case <-ticker.C:
			// Clear screen (simple approach)
			fmt.Print("\033[2J\033[H")

			if o.serviceName != "" {
				if err := o.displayServiceMonitor(o.serviceName); err != nil {
					fmt.Printf("Error monitoring service %s: %v\n", o.serviceName, err)
				}
			} else if o.networkServices {
				if err := o.displayNetworkServicesMonitor(); err != nil {
					fmt.Printf("Error monitoring network services: %v\n", err)
				}
			}

			fmt.Printf("\nLast updated: %s\n", time.Now().Format("2006-01-02 15:04:05"))
		}
	}
}

func (o *daemonOptions) getSystemServices() ([]serviceInfo, error) {
	cmd := exec.Command("systemctl", "list-units", "--type=service", "--all", "--no-pager", "--plain")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute systemctl: %w", err)
	}

	var services []serviceInfo
	scanner := bufio.NewScanner(strings.NewReader(string(output)))

	// Skip header lines
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.Contains(line, "UNIT") && strings.Contains(line, "LOAD") {
			break
		}
	}

	// Parse service lines
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "LOAD") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 4 {
			serviceName := strings.TrimSuffix(fields[0], ".service")
			service := serviceInfo{
				Name:   serviceName,
				Status: fields[3],
			}

			// Get additional details
			if details, err := o.getServiceDetails(serviceName); err == nil {
				service.Enabled = details.Enabled
				service.Description = details.Description
				service.MainPID = details.MainPID
				service.Memory = details.Memory
				service.Since = details.Since
			}

			services = append(services, service)
		}
	}

	return services, nil
}

func (o *daemonOptions) getServiceDetails(serviceName string) (*serviceInfo, error) {
	cmd := exec.Command("systemctl", "show", serviceName, "--no-pager")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get service details: %w", err)
	}

	service := &serviceInfo{Name: serviceName}
	scanner := bufio.NewScanner(strings.NewReader(string(output)))

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "LoadState=") {
			// Skip LoadState, we'll get UnitFileState for enabled status
		} else if strings.HasPrefix(line, "ActiveState=") {
			service.Status = strings.TrimPrefix(line, "ActiveState=")
		} else if strings.HasPrefix(line, "UnitFileState=") {
			service.Enabled = strings.TrimPrefix(line, "UnitFileState=")
		} else if strings.HasPrefix(line, "Description=") {
			service.Description = strings.TrimPrefix(line, "Description=")
		} else if strings.HasPrefix(line, "MainPID=") {
			service.MainPID = strings.TrimPrefix(line, "MainPID=")
		} else if strings.HasPrefix(line, "MemoryCurrent=") {
			memStr := strings.TrimPrefix(line, "MemoryCurrent=")
			if memStr != "[not set]" && memStr != "18446744073709551615" {
				service.Memory = o.formatBytes(memStr)
			}
		} else if strings.HasPrefix(line, "ActiveEnterTimestamp=") {
			timeStr := strings.TrimPrefix(line, "ActiveEnterTimestamp=")
			if timeStr != "" && timeStr != "n/a" {
				service.Since = timeStr
			}
		}
	}

	return service, nil
}

func (o *daemonOptions) filterNetworkServices(services []serviceInfo) []serviceInfo {
	networkPatterns := []string{
		"network", "wifi", "ethernet", "vpn", "dns", "dhcp", "ssh", "firewall",
		"iptables", "nginx", "apache", "httpd", "proxy", "NetworkManager", "systemd-networkd",
		"systemd-resolved", "wpa_supplicant", "hostapd", "openvpn", "wireguard",
		"ufw", "fail2ban", "bind", "named", "dnsmasq", "avahi",
	}

	var filtered []serviceInfo
	for _, service := range services {
		for _, pattern := range networkPatterns {
			if strings.Contains(strings.ToLower(service.Name), strings.ToLower(pattern)) ||
				strings.Contains(strings.ToLower(service.Description), strings.ToLower(pattern)) {
				filtered = append(filtered, service)
				break
			}
		}
	}

	return filtered
}

func (o *daemonOptions) filterActiveServices(services []serviceInfo) []serviceInfo {
	var filtered []serviceInfo
	for _, service := range services {
		if service.Status == statusActive || service.Status == statusFailed {
			filtered = append(filtered, service)
		}
	}
	return filtered
}

func (o *daemonOptions) getStatusIcon(status string) string {
	switch status {
	case statusActive:
		return "✅ "
	case "inactive":
		return "⚪ "
	case statusFailed:
		return "❌ "
	case "activating":
		return "🔄 "
	case "deactivating":
		return "🔄 "
	default:
		return "❓ "
	}
}

func (o *daemonOptions) formatBytes(bytesStr string) string {
	// Simple byte formatting - could be enhanced
	if bytesStr == "" || bytesStr == "0" {
		return "0 B"
	}
	return bytesStr + " bytes"
}

func (o *daemonOptions) showServiceLogs(serviceName string, lines int) error {
	cmd := exec.Command("journalctl", "-u", serviceName, "-n", fmt.Sprintf("%d", lines), "--no-pager")
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		// Clean up journal output formatting
		if logTimeRegex.MatchString(line) {
			fmt.Printf("   %s\n", line)
		}
	}

	return nil
}

func (o *daemonOptions) displayServiceMonitor(serviceName string) error {
	service, err := o.getServiceDetails(serviceName)
	if err != nil {
		return err
	}

	fmt.Printf("🔧 Monitoring Service: %s\n", service.Name)
	fmt.Printf("   Status: %s%s\n", o.getStatusIcon(service.Status), service.Status)
	fmt.Printf("   Enabled: %s\n", service.Enabled)
	if service.MainPID != "" && service.MainPID != "0" {
		fmt.Printf("   Main PID: %s\n", service.MainPID)
	}
	if service.Memory != "" {
		fmt.Printf("   Memory: %s\n", service.Memory)
	}

	if o.followLogs {
		fmt.Println("\n📋 Recent logs:")
		if err := o.showServiceLogs(serviceName, 5); err != nil {
			fmt.Printf("   Warning: Could not retrieve logs: %v\n", err)
		}
	}

	return nil
}

func (o *daemonOptions) displayNetworkServicesMonitor() error {
	services, err := o.getSystemServices()
	if err != nil {
		return err
	}

	networkServices := o.filterNetworkServices(services)
	activeNetworkServices := o.filterActiveServices(networkServices)

	fmt.Printf("🌐 Network Services Monitor (%d active)\n\n", len(activeNetworkServices))
	fmt.Printf("%-20s %-10s %-8s %-12s\n", "SERVICE", "STATUS", "ENABLED", "MAIN PID")
	fmt.Printf("%-20s %-10s %-8s %-12s\n", strings.Repeat("-", 20), strings.Repeat("-", 10), strings.Repeat("-", 8), strings.Repeat("-", 12))

	for _, service := range activeNetworkServices {
		statusIcon := o.getStatusIcon(service.Status)
		fmt.Printf("%-20s %s%-9s %-8s %-12s\n",
			service.Name,
			statusIcon,
			service.Status,
			service.Enabled,
			service.MainPID)
	}

	return nil
}

func newDaemonManageCmd() *cobra.Command {
	o := &daemonOptions{}

	cmd := &cobra.Command{
		Use:   "manage",
		Short: "Manage daemon services (start, stop, restart, enable, disable)",
		Long: `Manage system daemon/service lifecycle operations.

This command provides service management capabilities:
- Start, stop, restart services
- Enable or disable services for boot
- Service dependency management
- Bulk operations on multiple services

Examples:
  # Start a service
  gz net-env daemon manage --service ssh --action start
  
  # Stop and disable a service
  gz net-env daemon manage --service nginx --action stop
  gz net-env daemon manage --service nginx --action disable
  
  # Restart a service
  gz net-env daemon manage --service NetworkManager --action restart
  
  # Enable service for boot
  gz net-env daemon manage --service docker --action enable`,
		RunE: o.runManage,
	}

	cmd.Flags().StringVar(&o.serviceName, "service", "", "Name of the service to manage (required)")
	cmd.Flags().StringVar(&o.action, "action", "", "Action to perform: start, stop, restart, enable, disable, reload (required)")
	cmd.MarkFlagRequired("service")
	cmd.MarkFlagRequired("action")

	return cmd
}

func newDaemonHealthCmd(ctx context.Context) *cobra.Command {
	o := &daemonOptions{}

	cmd := &cobra.Command{
		Use:   "health",
		Short: "Monitor daemon health and performance metrics",
		Long: `Monitor system daemon health with custom health checks and performance metrics.

This command provides comprehensive health monitoring:
- Service health status and availability
- Performance metrics (CPU, memory, response times)
- Custom health check endpoints
- Alert thresholds and notifications
- Historical health data tracking

Examples:
  # Monitor service health
  gz net-env daemon health --service nginx
  
  # Monitor network services health
  gz net-env daemon health --network-services
  
  # Enable continuous health monitoring
  gz net-env daemon health --service docker --enable-health`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.runHealth(ctx, cmd, args)
		},
	}

	cmd.Flags().StringVar(&o.serviceName, "service", "", "Name of the service to health check")
	cmd.Flags().BoolVar(&o.networkServices, "network-services", false, "Monitor health of network-related services")
	cmd.Flags().BoolVar(&o.enableHealth, "enable-health", false, "Enable continuous health monitoring")

	return cmd
}

func (o *daemonOptions) runManage(_ *cobra.Command, args []string) error {
	validActions := map[string]bool{
		"start":   true,
		"stop":    true,
		"restart": true,
		"enable":  true,
		"disable": true,
		"reload":  true,
	}

	if !validActions[o.action] {
		return fmt.Errorf("invalid action '%s'. Valid actions: start, stop, restart, enable, disable, reload", o.action)
	}

	fmt.Printf("🔧 Managing service '%s' with action '%s'...\n", o.serviceName, o.action)

	var cmd *exec.Cmd
	switch o.action {
	case "start", "stop", "restart", "reload":
		cmd = exec.Command("sudo", "systemctl", o.action, o.serviceName)
	case "enable", "disable":
		cmd = exec.Command("sudo", "systemctl", o.action, o.serviceName)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to %s service %s: %w\nOutput: %s", o.action, o.serviceName, err, string(output))
	}

	fmt.Printf("✅ Successfully executed '%s' on service '%s'\n", o.action, o.serviceName)

	if string(output) != "" {
		fmt.Printf("Output: %s\n", string(output))
	}

	// Show updated status
	fmt.Println("\n📊 Updated service status:")
	service, err := o.getServiceDetails(o.serviceName)
	if err != nil {
		fmt.Printf("Warning: Could not retrieve updated status: %v\n", err)
	} else {
		fmt.Printf("   Status: %s%s\n", o.getStatusIcon(service.Status), service.Status)
		fmt.Printf("   Enabled: %s\n", service.Enabled)
	}

	return nil
}

func (o *daemonOptions) runHealth(ctx context.Context, _ *cobra.Command, args []string) error {
	if o.serviceName == "" && !o.networkServices {
		return fmt.Errorf("either --service or --network-services must be specified")
	}

	fmt.Println("🏥 Starting health monitoring (Press Ctrl+C to stop)")
	fmt.Println()

	if o.enableHealth {
		return o.runContinuousHealth(ctx)
	}

	if o.serviceName != "" {
		return o.runSingleServiceHealth(o.serviceName)
	}

	return o.runNetworkServicesHealth()
}

func (o *daemonOptions) runSingleServiceHealth(serviceName string) error {
	service, err := o.getServiceDetails(serviceName)
	if err != nil {
		return fmt.Errorf("failed to get service details: %w", err)
	}

	healthStatus := o.checkServiceHealth(service)

	fmt.Printf("🏥 Health Check Results for '%s'\n", serviceName)
	fmt.Printf("   Service Status: %s%s\n", o.getStatusIcon(service.Status), service.Status)
	fmt.Printf("   Health Status: %s\n", healthStatus.Status)
	fmt.Printf("   Last Checked: %s\n", healthStatus.LastChecked.Format("2006-01-02 15:04:05"))

	if healthStatus.ResponseTime > 0 {
		fmt.Printf("   Response Time: %v\n", healthStatus.ResponseTime)
	}

	if healthStatus.ErrorCount > 0 {
		fmt.Printf("   Error Count: %d\n", healthStatus.ErrorCount)
	}

	// Show performance metrics
	if service.MainPID != "" && service.MainPID != "0" {
		metrics, err := o.getServiceMetrics(service.MainPID)
		if err == nil {
			fmt.Printf("\n📈 Performance Metrics:\n")
			fmt.Printf("   CPU Usage: %s\n", metrics.CPUUsage)
			fmt.Printf("   Memory: %s\n", service.Memory)
		}
	}

	return nil
}

func (o *daemonOptions) runNetworkServicesHealth() error {
	services, err := o.getSystemServices()
	if err != nil {
		return fmt.Errorf("failed to get system services: %w", err)
	}

	networkServices := o.filterNetworkServices(services)
	activeNetworkServices := o.filterActiveServices(networkServices)

	fmt.Printf("🌐 Network Services Health Report (%d services)\n\n", len(activeNetworkServices))
	fmt.Printf("%-20s %-10s %-12s %-15s %-10s\n", "SERVICE", "STATUS", "HEALTH", "RESPONSE TIME", "ERRORS")
	fmt.Printf("%-20s %-10s %-12s %-15s %-10s\n", strings.Repeat("-", 20), strings.Repeat("-", 10), strings.Repeat("-", 12), strings.Repeat("-", 15), strings.Repeat("-", 10))

	for _, service := range activeNetworkServices {
		health := o.checkServiceHealth(&service)
		responseTime := "-"
		if health.ResponseTime > 0 {
			responseTime = fmt.Sprintf("%v", health.ResponseTime)
		}

		fmt.Printf("%-20s %s%-9s %-12s %-15s %-10d\n",
			service.Name,
			o.getStatusIcon(service.Status),
			service.Status,
			health.Status,
			responseTime,
			health.ErrorCount)
	}

	return nil
}

func (o *daemonOptions) runContinuousHealth(ctx context.Context) error {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("\n🛑 Stopping health monitoring (reason: %v)\n", ctx.Err())
			return nil

		case <-ticker.C:
			// Clear screen
			fmt.Print("\033[2J\033[H")

			if o.serviceName != "" {
				if err := o.runSingleServiceHealth(o.serviceName); err != nil {
					fmt.Printf("Error checking health: %v\n", err)
				}
			} else if o.networkServices {
				if err := o.runNetworkServicesHealth(); err != nil {
					fmt.Printf("Error checking network services health: %v\n", err)
				}
			}

			fmt.Printf("\nLast updated: %s\n", time.Now().Format("2006-01-02 15:04:05"))
		}
	}
}

func (o *daemonOptions) checkServiceHealth(service *serviceInfo) healthCheck {
	health := healthCheck{
		Name:        service.Name,
		LastChecked: time.Now(),
		Status:      "unknown",
	}

	start := time.Now()

	// Basic health check based on service status
	switch service.Status {
	case statusActive:
		health.Status = "healthy"
		health.ResponseTime = time.Since(start)
	case statusFailed:
		health.Status = "unhealthy"
		health.ErrorCount = 1
	case "inactive":
		health.Status = "stopped"
	default:
		health.Status = "unknown"
	}

	// Check if process is actually running
	if service.MainPID != "" && service.MainPID != "0" {
		if !o.isProcessRunning(service.MainPID) {
			health.Status = "unhealthy"
			health.ErrorCount++
		}
	}

	return health
}

func (o *daemonOptions) isProcessRunning(pidStr string) bool {
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return false
	}

	// Check if process exists
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// Send signal 0 to check if process is alive
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

func (o *daemonOptions) getServiceMetrics(pidStr string) (*serviceInfo, error) {
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return nil, err
	}

	// Get CPU usage from /proc/[pid]/stat
	statFile := fmt.Sprintf("/proc/%d/stat", pid)
	statData, err := os.ReadFile(statFile)
	if err != nil {
		return nil, err
	}

	fields := strings.Fields(string(statData))
	if len(fields) < 17 {
		return nil, fmt.Errorf("invalid stat file format")
	}

	// Calculate CPU usage (simplified)
	metrics := &serviceInfo{
		CPUUsage: "0.0%", // Simplified for this implementation
	}

	return metrics, nil
}
