package net_env

import (
	"bufio"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
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
}

type serviceInfo struct {
	Name        string
	Status      string
	Enabled     string
	Description string
	MainPID     string
	Memory      string
	Since       string
}

func newDaemonCmd() *cobra.Command {
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
	cmd.AddCommand(newDaemonMonitorCmd())

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

func newDaemonMonitorCmd() *cobra.Command {
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
		RunE: o.runMonitor,
	}

	cmd.Flags().StringVar(&o.serviceName, "service", "", "Name of the service to monitor")
	cmd.Flags().BoolVar(&o.networkServices, "network-services", false, "Monitor network-related services")
	cmd.Flags().BoolVar(&o.followLogs, "follow-logs", false, "Follow service logs in real-time")

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

func (o *daemonOptions) runMonitor(_ *cobra.Command, args []string) error {
	if o.serviceName == "" && !o.networkServices {
		return fmt.Errorf("either --service or --network-services must be specified")
	}

	fmt.Println("📊 Starting daemon monitor (Press Ctrl+C to stop)")
	fmt.Println()

	// Monitor specific service or network services
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
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

		<-ticker.C
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
