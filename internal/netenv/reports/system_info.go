// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package reports

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// SystemInfoCollector gathers system-level network information.
type SystemInfoCollector struct {
	platform string
}

// NewSystemInfoCollector creates a new system info collector.
func NewSystemInfoCollector() *SystemInfoCollector {
	return &SystemInfoCollector{
		platform: runtime.GOOS,
	}
}

// CollectSystemInfo gathers comprehensive system network information.
func (sic *SystemInfoCollector) CollectSystemInfo() (*SystemInfo, error) {
	info := &SystemInfo{
		Platform: sic.platform,
	}

	// Get hostname
	if hostname, err := os.Hostname(); err == nil {
		info.Hostname = hostname
	}

	// Get kernel version (Linux/Unix)
	if sic.platform != osWindows {
		if version, err := sic.getKernelVersion(); err == nil {
			info.KernelVersion = version
		}
	}

	// Get default gateway
	if gateway, err := sic.getDefaultGateway(); err == nil {
		info.DefaultGateway = gateway
	}

	// Get DNS servers
	if dnsServers, err := sic.getDNSServers(); err == nil {
		info.DNSServers = dnsServers
	}

	// Get routing table
	if routes, err := sic.getRoutingTable(); err == nil {
		info.RoutingTable = routes
	}

	// Get network namespaces (Linux only)
	if sic.platform == osLinux {
		if namespaces, err := sic.getNetworkNamespaces(); err == nil {
			info.NetworkNamespaces = namespaces
		}

		// Get firewall status
		if status, err := sic.getFirewallStatus(); err == nil {
			info.FirewallStatus = status
		}
	}

	return info, nil
}

// getKernelVersion retrieves the kernel version.
func (sic *SystemInfoCollector) getKernelVersion() (string, error) {
	cmd := exec.CommandContext(context.Background(), "uname", "-r")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// getDefaultGateway finds the default gateway.
func (sic *SystemInfoCollector) getDefaultGateway() (string, error) {
	switch sic.platform {
	case osLinux:
		return sic.getDefaultGatewayLinux()
	case osDarwin:
		return sic.getDefaultGatewayDarwin()
	case osWindows:
		return sic.getDefaultGatewayWindows()
	default:
		return "", fmt.Errorf("unsupported platform: %s", sic.platform)
	}
}

// getDefaultGatewayLinux gets default gateway on Linux.
func (sic *SystemInfoCollector) getDefaultGatewayLinux() (string, error) {
	// Try reading from /proc/net/route
	file, err := os.Open("/proc/net/route")
	if err != nil {
		return sic.getDefaultGatewayFromRoute()
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan() // Skip header

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 3 && fields[1] == "00000000" {
			// Default route found, parse gateway
			gatewayHex := fields[2]
			if gateway, err := sic.parseHexIP(gatewayHex); err == nil {
				return gateway, nil
			}
		}
	}

	return sic.getDefaultGatewayFromRoute()
}

// parseHexIP converts hex IP address to dotted decimal.
func (sic *SystemInfoCollector) parseHexIP(hexIP string) (string, error) {
	if len(hexIP) != 8 {
		return "", fmt.Errorf("invalid hex IP length")
	}

	var octets []string
	for i := 6; i >= 0; i -= 2 {
		octet, err := strconv.ParseUint(hexIP[i:i+2], 16, 8)
		if err != nil {
			return "", err
		}
		octets = append(octets, strconv.Itoa(int(octet)))
	}

	return strings.Join(octets, "."), nil
}

// getDefaultGatewayFromRoute uses route command as fallback.
func (sic *SystemInfoCollector) getDefaultGatewayFromRoute() (string, error) {
	cmd := exec.CommandContext(context.Background(), "route", "-n")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "0.0.0.0") && strings.Contains(line, "UG") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				return fields[1], nil
			}
		}
	}

	return "", fmt.Errorf("default gateway not found")
}

// getDefaultGatewayDarwin gets default gateway on macOS.
func (sic *SystemInfoCollector) getDefaultGatewayDarwin() (string, error) {
	cmd := exec.Command("route", "-n", "get", "default")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "gateway:") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				return strings.TrimSpace(parts[1]), nil
			}
		}
	}

	return "", fmt.Errorf("default gateway not found")
}

// getDefaultGatewayWindows gets default gateway on Windows.
func (sic *SystemInfoCollector) getDefaultGatewayWindows() (string, error) {
	cmd := exec.Command("route", "print", "0.0.0.0")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "0.0.0.0") {
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				return fields[2], nil
			}
		}
	}

	return "", fmt.Errorf("default gateway not found")
}

// getDNSServers retrieves configured DNS servers.
func (sic *SystemInfoCollector) getDNSServers() ([]string, error) {
	switch sic.platform {
	case "linux", "darwin":
		return sic.getDNSServersUnix()
	case "windows":
		return sic.getDNSServersWindows()
	default:
		return nil, fmt.Errorf("unsupported platform: %s", sic.platform)
	}
}

// getDNSServersUnix gets DNS servers on Unix systems.
func (sic *SystemInfoCollector) getDNSServersUnix() ([]string, error) {
	file, err := os.Open("/etc/resolv.conf")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var dnsServers []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "nameserver ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				dnsServers = append(dnsServers, parts[1])
			}
		}
	}

	return dnsServers, scanner.Err()
}

// getDNSServersWindows gets DNS servers on Windows.
func (sic *SystemInfoCollector) getDNSServersWindows() ([]string, error) {
	cmd := exec.Command("nslookup", "localhost")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var dnsServers []string
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "Server:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				dnsServers = append(dnsServers, parts[1])
			}
		}
	}

	return dnsServers, nil
}

// getRoutingTable retrieves the routing table.
func (sic *SystemInfoCollector) getRoutingTable() ([]RouteEntry, error) {
	switch sic.platform {
	case "linux":
		return sic.getRoutingTableLinux()
	case "darwin":
		return sic.getRoutingTableDarwin()
	case "windows":
		return sic.getRoutingTableWindows()
	default:
		return nil, fmt.Errorf("unsupported platform: %s", sic.platform)
	}
}

// getRoutingTableLinux gets routing table on Linux.
func (sic *SystemInfoCollector) getRoutingTableLinux() ([]RouteEntry, error) {
	cmd := exec.Command("route", "-n")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var routes []RouteEntry
	lines := strings.Split(string(output), "\n")

	// Skip header lines
	for i, line := range lines {
		if i <= 1 || strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 8 {
			metric := 0
			if m, err := strconv.Atoi(fields[4]); err == nil {
				metric = m
			}

			routes = append(routes, RouteEntry{
				Destination: fields[0],
				Gateway:     fields[1],
				Interface:   fields[7],
				Metric:      metric,
			})
		}
	}

	return routes, nil
}

// getRoutingTableDarwin gets routing table on macOS.
func (sic *SystemInfoCollector) getRoutingTableDarwin() ([]RouteEntry, error) {
	cmd := exec.Command("netstat", "-rn", "-f", "inet")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var routes []RouteEntry
	lines := strings.Split(string(output), "\n")

	for i, line := range lines {
		if i <= 2 || strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 6 {
			routes = append(routes, RouteEntry{
				Destination: fields[0],
				Gateway:     fields[1],
				Interface:   fields[3],
				Metric:      0, // macOS doesn't show metric in netstat -rn
			})
		}
	}

	return routes, nil
}

// getRoutingTableWindows gets routing table on Windows.
func (sic *SystemInfoCollector) getRoutingTableWindows() ([]RouteEntry, error) {
	cmd := exec.Command("route", "print")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var routes []RouteEntry
	lines := strings.Split(string(output), "\n")
	inIPv4Section := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.Contains(line, "IPv4 Route Table") {
			inIPv4Section = true
			continue
		}

		if strings.Contains(line, "IPv6 Route Table") {
			inIPv4Section = false
			continue
		}

		if inIPv4Section && len(line) > 0 && !strings.HasPrefix(line, "=") && !strings.HasPrefix(line, "Network") {
			fields := strings.Fields(line)
			if len(fields) >= 5 {
				metric := 0
				if m, err := strconv.Atoi(fields[4]); err == nil {
					metric = m
				}

				routes = append(routes, RouteEntry{
					Destination: fields[0],
					Gateway:     fields[2],
					Interface:   fields[3],
					Metric:      metric,
				})
			}
		}
	}

	return routes, nil
}

// getNetworkNamespaces gets network namespaces (Linux only).
func (sic *SystemInfoCollector) getNetworkNamespaces() ([]string, error) {
	cmd := exec.Command("ip", "netns", "list")
	output, err := cmd.Output()
	if err != nil {
		// ip command might not be available or no namespaces
		return []string{}, err
	}

	var namespaces []string
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			// Extract namespace name (first word)
			parts := strings.Fields(line)
			if len(parts) > 0 {
				namespaces = append(namespaces, parts[0])
			}
		}
	}

	return namespaces, nil
}

// getFirewallStatus gets firewall status (Linux only).
func (sic *SystemInfoCollector) getFirewallStatus() (string, error) {
	// Try iptables first
	if status, err := sic.getIptablesStatus(); err == nil {
		return status, nil
	}

	// Try ufw (Ubuntu)
	if status, err := sic.getUfwStatus(); err == nil {
		return status, nil
	}

	// Try firewalld (CentOS/RHEL)
	if status, err := sic.getFirewalldStatus(); err == nil {
		return status, nil
	}

	return "unknown", nil
}

// getIptablesStatus checks iptables status.
func (sic *SystemInfoCollector) getIptablesStatus() (string, error) {
	cmd := exec.Command("iptables", "-L", "-n")
	_, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return "iptables active", nil
}

// getUfwStatus checks ufw status.
func (sic *SystemInfoCollector) getUfwStatus() (string, error) {
	cmd := exec.Command("ufw", "status")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	if strings.Contains(string(output), "Status: active") {
		return "ufw active", nil
	} else if strings.Contains(string(output), "Status: inactive") {
		return "ufw inactive", nil
	}

	return "ufw unknown", nil
}

// getFirewalldStatus checks firewalld status.
func (sic *SystemInfoCollector) getFirewalldStatus() (string, error) {
	cmd := exec.Command("firewall-cmd", "--state")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	if strings.Contains(string(output), "running") {
		return "firewalld active", nil
	}

	return "firewalld inactive", nil
}
