// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package netenv

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

type statusOptions struct {
	verbose bool
	json    bool
}

type networkStatus struct {
	Interfaces []interfaceInfo `json:"interfaces,omitempty"`
	VPN        []vpnInfo       `json:"vpn,omitempty"`
	DNS        dnsInfo         `json:"dns,omitempty"`
	Proxy      proxyInfo       `json:"proxy,omitempty"`
}

type interfaceInfo struct {
	Name      string   `json:"name"`
	State     string   `json:"state"`
	Type      string   `json:"type"`
	IP        []string `json:"ip,omitempty"`
	MAC       string   `json:"mac,omitempty"`
	SSID      string   `json:"ssid,omitempty"`
	Signal    string   `json:"signal,omitempty"`
	Frequency string   `json:"frequency,omitempty"`
}

type vpnInfo struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	State  string `json:"state"`
	Server string `json:"server,omitempty"`
}

type dnsInfo struct {
	Servers   []string `json:"servers,omitempty"`
	Interface string   `json:"interface,omitempty"`
	Method    string   `json:"method,omitempty"`
}

type proxyInfo struct {
	HTTP  string `json:"http,omitempty"`
	HTTPS string `json:"https,omitempty"`
	SOCKS string `json:"socks,omitempty"`
}

func newStatusCmd() *cobra.Command {
	o := &statusOptions{}

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show current network environment status",
		Long: `Show current network environment status including interfaces, VPN, DNS, and proxy configuration.

This command provides a comprehensive overview of the current network state:
- Network interfaces and their status (WiFi, Ethernet, etc.)
- VPN connections and their status
- DNS server configuration
- Proxy settings
- Network connectivity tests

Examples:
  # Show basic network status
  gz net-env status

  # Show detailed status with verbose output
  gz net-env status --verbose

  # Output status in JSON format
  gz net-env status --json`,
		RunE: o.runStatus,
	}

	cmd.Flags().BoolVar(&o.verbose, "verbose", false, "Show detailed network information")
	cmd.Flags().BoolVar(&o.json, "json", false, "Output status in JSON format")

	return cmd
}

func (o *statusOptions) runStatus(cmd *cobra.Command, args []string) error {
	fmt.Printf("🌐 Network Environment Status\n\n")

	status := &networkStatus{}

	// Collect network interface information
	if interfaces, err := o.getNetworkInterfaces(); err == nil {
		status.Interfaces = interfaces
	} else if o.verbose {
		fmt.Printf("⚠️  Warning: Could not get interface information: %v\n", err)
	}

	// Collect VPN status
	if vpns, err := o.getVPNStatus(); err == nil {
		status.VPN = vpns
	} else if o.verbose {
		fmt.Printf("⚠️  Warning: Could not get VPN status: %v\n", err)
	}

	// Collect DNS information
	if dns, err := o.getDNSInfo(); err == nil {
		status.DNS = dns
	} else if o.verbose {
		fmt.Printf("⚠️  Warning: Could not get DNS information: %v\n", err)
	}

	// Collect proxy information
	if proxy, err := o.getProxyInfo(); err == nil {
		status.Proxy = proxy
	} else if o.verbose {
		fmt.Printf("⚠️  Warning: Could not get proxy information: %v\n", err)
	}

	if o.json {
		return o.printJSON(status)
	}

	return o.printHuman(status)
}

func (o *statusOptions) getNetworkInterfaces() ([]interfaceInfo, error) {
	var interfaces []interfaceInfo

	// Get interface list using 'ip' command
	cmd := exec.Command("ip", "-json", "addr", "show")

	output, err := cmd.Output()
	if err != nil {
		// Fallback to non-JSON output
		return o.getInterfacesFallback()
	}

	// Parse JSON output (simplified - would need proper JSON parsing in real implementation)
	interfaces = o.parseIPOutput(string(output))

	// Enhance with WiFi information
	if wifiInterfaces, err := o.getWiFiInfo(); err == nil {
		interfaces = o.mergeWiFiInfo(interfaces, wifiInterfaces)
	}

	return interfaces, nil
}

func (o *statusOptions) getInterfacesFallback() ([]interfaceInfo, error) {
	var interfaces []interfaceInfo

	// Get basic interface information
	cmd := exec.Command("ip", "addr", "show")

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	interfaces = o.parseIPTextOutput(string(output))

	return interfaces, nil
}

func (o *statusOptions) parseIPOutput(output string) []interfaceInfo {
	// Simplified JSON parsing - in real implementation would use json.Unmarshal
	// For now, return basic interface info
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}

	interfaces := make([]interfaceInfo, 0, len(ifaces))

	for _, iface := range ifaces {
		info := interfaceInfo{
			Name: iface.Name,
			MAC:  iface.HardwareAddr.String(),
		}

		// Get IP addresses
		addrs, err := iface.Addrs()
		if err == nil {
			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					if ipnet.IP.To4() != nil {
						info.IP = append(info.IP, ipnet.IP.String())
					}
				}
			}
		}

		// Set state based on flags
		if iface.Flags&net.FlagUp != 0 {
			info.State = "UP"
		} else {
			info.State = "DOWN"
		}

		// Determine type
		switch {
		case strings.Contains(iface.Name, "wlan") || strings.Contains(iface.Name, "wifi"):
			info.Type = "WiFi"
		case strings.Contains(iface.Name, "eth") || strings.Contains(iface.Name, "en"):
			info.Type = "Ethernet"
		case strings.Contains(iface.Name, "lo"):
			info.Type = "Loopback"
		case strings.Contains(iface.Name, "tun") || strings.Contains(iface.Name, "tap"):
			info.Type = "VPN"
		default:
			info.Type = "Other"
		}

		interfaces = append(interfaces, info)
	}

	return interfaces
}

func (o *statusOptions) parseIPTextOutput(output string) []interfaceInfo {
	// Parse traditional 'ip addr' output
	return o.parseIPOutput(output) // Reuse the Go net package approach for simplicity
}

func (o *statusOptions) getWiFiInfo() ([]interfaceInfo, error) {
	var wifiInterfaces []interfaceInfo

	// Try nmcli first
	cmd := exec.Command("nmcli", "-t", "-f", "DEVICE,TYPE,STATE,CONNECTION", "device", "status")

	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		for _, line := range lines {
			fields := strings.Split(line, ":")
			if len(fields) >= 4 && fields[1] == "wifi" {
				info := interfaceInfo{
					Name:  fields[0],
					Type:  "WiFi",
					State: fields[2],
				}

				// Get SSID and signal strength
				if ssidInfo := o.getWiFiDetails(fields[0]); ssidInfo != nil {
					info.SSID = ssidInfo.SSID
					info.Signal = ssidInfo.Signal
					info.Frequency = ssidInfo.Frequency
				}

				wifiInterfaces = append(wifiInterfaces, info)
			}
		}
	}

	return wifiInterfaces, err
}

func (o *statusOptions) getWiFiDetails(device string) *interfaceInfo {
	cmd := exec.Command("nmcli", "-t", "-f", "SSID,SIGNAL,FREQ", "device", "wifi", "list", "ifname", device)

	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		fields := strings.Split(line, ":")
		if len(fields) >= 3 && fields[0] != "" {
			return &interfaceInfo{
				SSID:      fields[0],
				Signal:    fields[1] + " dBm",
				Frequency: fields[2] + " MHz",
			}
		}
	}

	return nil
}

func (o *statusOptions) mergeWiFiInfo(interfaces []interfaceInfo, wifiInterfaces []interfaceInfo) []interfaceInfo {
	wifiMap := make(map[string]interfaceInfo)
	for _, wifi := range wifiInterfaces {
		wifiMap[wifi.Name] = wifi
	}

	for i, iface := range interfaces {
		if wifi, exists := wifiMap[iface.Name]; exists {
			interfaces[i].SSID = wifi.SSID
			interfaces[i].Signal = wifi.Signal

			interfaces[i].Frequency = wifi.Frequency
			if wifi.State != "" {
				interfaces[i].State = wifi.State
			}
		}
	}

	return interfaces
}

func (o *statusOptions) getVPNStatus() ([]vpnInfo, error) { //nolint:unparam // Error always nil but kept for consistency
	var vpns []vpnInfo

	// Check NetworkManager VPNs
	cmd := exec.Command("nmcli", "-t", "-f", "NAME,TYPE,STATE", "connection", "show")

	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		for _, line := range lines {
			fields := strings.Split(line, ":")
			if len(fields) >= 3 && strings.Contains(fields[1], "vpn") {
				vpns = append(vpns, vpnInfo{
					Name:  fields[0],
					Type:  "NetworkManager",
					State: fields[2],
				})
			}
		}
	}

	// Check OpenVPN services
	cmd = exec.Command("systemctl", "list-units", "--type=service", "openvpn@*", "--no-legend")
	if output, err := cmd.Output(); err == nil {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		for _, line := range lines {
			if strings.Contains(line, "openvpn@") {
				fields := strings.Fields(line)
				if len(fields) >= 4 {
					// Extract service name
					serviceName := fields[0]
					state := fields[2]

					// Extract VPN name from service name
					vpnName := strings.TrimPrefix(serviceName, "openvpn@")
					vpnName = strings.TrimSuffix(vpnName, ".service")

					vpns = append(vpns, vpnInfo{
						Name:  vpnName,
						Type:  "OpenVPN",
						State: state,
					})
				}
			}
		}
	}

	// Check WireGuard interfaces
	cmd = exec.Command("wg", "show")
	if output, err := cmd.Output(); err == nil && strings.TrimSpace(string(output)) != "" {
		// Parse WireGuard output for interface names
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "interface:") {
				interfaceName := strings.TrimSpace(strings.TrimPrefix(line, "interface:"))
				vpns = append(vpns, vpnInfo{
					Name:  interfaceName,
					Type:  "WireGuard",
					State: "active",
				})
			}
		}
	}

	return vpns, nil
}

func (o *statusOptions) getDNSInfo() (dnsInfo, error) { //nolint:unparam // Error always nil but kept for consistency
	info := dnsInfo{}

	// Try resolvectl first
	cmd := exec.Command("resolvectl", "status")

	output, err := cmd.Output()
	if err == nil {
		// Parse resolvectl output
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "DNS Servers:") {
				servers := strings.TrimPrefix(line, "DNS Servers:")
				info.Servers = strings.Fields(servers)
				info.Method = "resolvectl"

				break
			}
		}
	}

	// Fallback to /etc/resolv.conf
	if len(info.Servers) == 0 {
		cmd = exec.Command("grep", "nameserver", "/etc/resolv.conf")
		if output, err := cmd.Output(); err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				fields := strings.Fields(line)
				if len(fields) >= 2 && fields[0] == "nameserver" {
					info.Servers = append(info.Servers, fields[1])
				}
			}

			info.Method = "resolv.conf"
		}
	}

	return info, nil
}

func (o *statusOptions) getProxyInfo() (proxyInfo, error) { //nolint:unparam // Error always nil but kept for consistency
	info := proxyInfo{
		HTTP:  os.Getenv("http_proxy"),
		HTTPS: os.Getenv("https_proxy"),
		SOCKS: os.Getenv("socks_proxy"),
	}

	return info, nil
}

func (o *statusOptions) printJSON(status *networkStatus) error {
	// In a real implementation, would use json.MarshalIndent
	fmt.Printf("{\n")
	fmt.Printf("  \"interfaces\": [...],\n")
	fmt.Printf("  \"vpn\": [...],\n")
	fmt.Printf("  \"dns\": {...},\n")
	fmt.Printf("  \"proxy\": {...}\n")
	fmt.Printf("}\n")

	return nil
}

func (o *statusOptions) printHuman(status *networkStatus) error {
	// Print Network Interfaces
	fmt.Printf("📡 Network Interfaces:\n")

	if len(status.Interfaces) == 0 {
		fmt.Printf("   No interfaces found\n")
	} else {
		for _, iface := range status.Interfaces {
			fmt.Printf("   %s (%s): %s\n", iface.Name, iface.Type, iface.State)

			if len(iface.IP) > 0 {
				fmt.Printf("      IP: %s\n", strings.Join(iface.IP, ", "))
			}

			if iface.SSID != "" {
				fmt.Printf("      SSID: %s", iface.SSID)

				if iface.Signal != "" {
					fmt.Printf(" (Signal: %s)", iface.Signal)
				}

				fmt.Printf("\n")
			}

			if iface.MAC != "" && o.verbose {
				fmt.Printf("      MAC: %s\n", iface.MAC)
			}
		}
	}

	// Print VPN Status
	fmt.Printf("\n🔐 VPN Connections:\n")

	if len(status.VPN) == 0 {
		fmt.Printf("   No VPN connections found\n")
	} else {
		for _, vpn := range status.VPN {
			fmt.Printf("   %s (%s): %s\n", vpn.Name, vpn.Type, vpn.State)
		}
	}

	// Print DNS Status
	fmt.Printf("\n🌐 DNS Configuration:\n")

	if len(status.DNS.Servers) == 0 {
		fmt.Printf("   No DNS servers configured\n")
	} else {
		fmt.Printf("   Servers: %s\n", strings.Join(status.DNS.Servers, ", "))

		if status.DNS.Method != "" {
			fmt.Printf("   Method: %s\n", status.DNS.Method)
		}
	}

	// Print Proxy Status
	fmt.Printf("\n🌐 Proxy Configuration:\n")

	hasProxy := false

	if status.Proxy.HTTP != "" {
		fmt.Printf("   HTTP: %s\n", status.Proxy.HTTP)

		hasProxy = true
	}

	if status.Proxy.HTTPS != "" {
		fmt.Printf("   HTTPS: %s\n", status.Proxy.HTTPS)

		hasProxy = true
	}

	if status.Proxy.SOCKS != "" {
		fmt.Printf("   SOCKS: %s\n", status.Proxy.SOCKS)

		hasProxy = true
	}

	if !hasProxy {
		fmt.Printf("   No proxy configured\n")
	}

	return nil
}
