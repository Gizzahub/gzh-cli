// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package netenv

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// NetworkDetector handles automatic network environment detection
type NetworkDetector struct {
	profiles []NetworkProfile
}

// NewNetworkDetector creates a new network detector
func NewNetworkDetector(profiles []NetworkProfile) *NetworkDetector {
	return &NetworkDetector{
		profiles: profiles,
	}
}

// DetectEnvironment automatically detects the current network environment
func (nd *NetworkDetector) DetectEnvironment(ctx context.Context) (*NetworkProfile, error) {
	// Get current network information
	networkInfo, err := nd.getCurrentNetworkInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get network info: %w", err)
	}

	// Score each profile based on how well it matches current conditions
	bestProfile := nd.findBestMatchingProfile(networkInfo)

	return bestProfile, nil
}

// getCurrentNetworkInfo gathers current network environment information
func (nd *NetworkDetector) getCurrentNetworkInfo(ctx context.Context) (*NetworkInfo, error) {
	info := &NetworkInfo{
		Timestamp: time.Now(),
	}

	// Get WiFi SSID
	if ssid, err := nd.getWiFiSSID(ctx); err == nil {
		info.WiFiSSID = ssid
	}

	// Get IP addresses
	if ips, err := nd.getLocalIPs(); err == nil {
		info.LocalIPs = ips
	}

	// Get hostname
	if hostname, err := os.Hostname(); err == nil {
		info.Hostname = hostname
	}

	// Get default gateway
	if gateway, err := nd.getDefaultGateway(ctx); err == nil {
		info.DefaultGateway = gateway
	}

	// Get DNS servers
	if dns, err := nd.getDNSServers(); err == nil {
		info.DNSServers = dns
	}

	return info, nil
}

// getWiFiSSID gets the current WiFi SSID
func (nd *NetworkDetector) getWiFiSSID(ctx context.Context) (string, error) {
	switch runtime.GOOS {
	case "darwin":
		return nd.getWiFiSSIDMacOS(ctx)
	case "linux":
		return nd.getWiFiSSIDLinux(ctx)
	case "windows":
		return nd.getWiFiSSIDWindows(ctx)
	default:
		return "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// getWiFiSSIDMacOS gets WiFi SSID on macOS
func (nd *NetworkDetector) getWiFiSSIDMacOS(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "/System/Library/PrivateFrameworks/Apple80211.framework/Versions/Current/Resources/airport", "-I")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, " SSID:") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				return strings.TrimSpace(parts[1]), nil
			}
		}
	}

	return "", fmt.Errorf("SSID not found")
}

// getWiFiSSIDLinux gets WiFi SSID on Linux
func (nd *NetworkDetector) getWiFiSSIDLinux(ctx context.Context) (string, error) {
	// Try iwgetid first
	cmd := exec.CommandContext(ctx, "iwgetid", "-r")
	if output, err := cmd.Output(); err == nil {
		ssid := strings.TrimSpace(string(output))
		if ssid != "" {
			return ssid, nil
		}
	}

	// Try nmcli as fallback
	cmd = exec.CommandContext(ctx, "nmcli", "-t", "-f", "active,ssid", "dev", "wifi")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "yes:") {
			return strings.TrimPrefix(line, "yes:"), nil
		}
	}

	return "", fmt.Errorf("SSID not found")
}

// getWiFiSSIDWindows gets WiFi SSID on Windows
func (nd *NetworkDetector) getWiFiSSIDWindows(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "netsh", "wlan", "show", "profiles")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// This is a simplified implementation - would need more parsing for Windows
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Profile") && strings.Contains(line, ":") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				return strings.TrimSpace(parts[1]), nil
			}
		}
	}

	return "", fmt.Errorf("SSID not found")
}

// getLocalIPs gets all local IP addresses
func (nd *NetworkDetector) getLocalIPs() ([]string, error) {
	var ips []string

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP.String())
			}
		}
	}

	return ips, nil
}

// getDefaultGateway gets the default gateway IP
func (nd *NetworkDetector) getDefaultGateway(ctx context.Context) (string, error) {
	switch runtime.GOOS {
	case "darwin", "linux":
		return nd.getDefaultGatewayUnix(ctx)
	case "windows":
		return nd.getDefaultGatewayWindows(ctx)
	default:
		return "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// getDefaultGatewayUnix gets default gateway on Unix systems
func (nd *NetworkDetector) getDefaultGatewayUnix(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "route", "-n", "get", "default")
	output, err := cmd.Output()
	if err != nil {
		// Try alternative command
		cmd = exec.CommandContext(ctx, "ip", "route", "show", "default")
		output, err = cmd.Output()
		if err != nil {
			return "", err
		}
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "gateway") || strings.Contains(line, "via") {
			fields := strings.Fields(line)
			for i, field := range fields {
				if (field == "gateway" || field == "via") && i+1 < len(fields) {
					return fields[i+1], nil
				}
			}
		}
	}

	return "", fmt.Errorf("gateway not found")
}

// getDefaultGatewayWindows gets default gateway on Windows
func (nd *NetworkDetector) getDefaultGatewayWindows(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "ipconfig")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Default Gateway") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				gateway := strings.TrimSpace(parts[1])
				if gateway != "" {
					return gateway, nil
				}
			}
		}
	}

	return "", fmt.Errorf("gateway not found")
}

// getDNSServers gets current DNS servers
func (nd *NetworkDetector) getDNSServers() ([]string, error) {
	// Read resolv.conf on Unix systems
	if runtime.GOOS != "windows" {
		return nd.getDNSServersUnix()
	}

	return nd.getDNSServersWindows()
}

// getDNSServersUnix gets DNS servers from resolv.conf
func (nd *NetworkDetector) getDNSServersUnix() ([]string, error) {
	content, err := os.ReadFile("/etc/resolv.conf")
	if err != nil {
		return nil, err
	}

	var servers []string
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "nameserver") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				servers = append(servers, fields[1])
			}
		}
	}

	return servers, nil
}

// getDNSServersWindows gets DNS servers on Windows
func (nd *NetworkDetector) getDNSServersWindows() ([]string, error) {
	// This would require more complex Windows API calls or registry access
	// For now, return empty slice
	return []string{}, nil
}

// findBestMatchingProfile finds the profile that best matches current conditions
func (nd *NetworkDetector) findBestMatchingProfile(networkInfo *NetworkInfo) *NetworkProfile {
	var bestProfile *NetworkProfile
	bestScore := 0

	for _, profile := range nd.profiles {
		score := nd.scoreProfile(&profile, networkInfo)
		if score > bestScore {
			bestScore = score
			bestProfile = &profile
		}
	}

	return bestProfile
}

// scoreProfile calculates how well a profile matches current network conditions
func (nd *NetworkDetector) scoreProfile(profile *NetworkProfile, networkInfo *NetworkInfo) int {
	score := 0

	for _, condition := range profile.Conditions {
		switch condition.Type {
		case "wifi_ssid":
			if nd.matchCondition(condition, networkInfo.WiFiSSID) {
				score += 100
			}
		case "ip_range":
			for _, ip := range networkInfo.LocalIPs {
				if nd.matchIPRange(condition.Value, ip) {
					score += 50
				}
			}
		case "hostname":
			if nd.matchCondition(condition, networkInfo.Hostname) {
				score += 30
			}
		case "gateway":
			if nd.matchCondition(condition, networkInfo.DefaultGateway) {
				score += 70
			}
		}
	}

	// Add priority bonus
	score += profile.Priority

	return score
}

// matchCondition checks if a condition matches a value
func (nd *NetworkDetector) matchCondition(condition NetworkCondition, value string) bool {
	switch condition.Operator {
	case "contains":
		return strings.Contains(value, condition.Value)
	case "matches":
		// Could implement regex matching here
		return value == condition.Value
	case "equals", "":
		return value == condition.Value
	default:
		return false
	}
}

// matchIPRange checks if an IP address is within a specified range
func (nd *NetworkDetector) matchIPRange(cidr, ip string) bool {
	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		// If not CIDR, try exact match
		return cidr == ip
	}

	ipAddr := net.ParseIP(ip)
	if ipAddr == nil {
		return false
	}

	return network.Contains(ipAddr)
}

// NetworkInfo contains current network environment information
type NetworkInfo struct {
	WiFiSSID       string    `json:"wifi_ssid,omitempty"`
	LocalIPs       []string  `json:"local_ips,omitempty"`
	Hostname       string    `json:"hostname,omitempty"`
	DefaultGateway string    `json:"default_gateway,omitempty"`
	DNSServers     []string  `json:"dns_servers,omitempty"`
	Timestamp      time.Time `json:"timestamp"`
}
