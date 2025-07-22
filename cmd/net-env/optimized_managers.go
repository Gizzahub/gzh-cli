// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package netenv

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
)

// OptimizedVPNManager provides batch VPN operations for better performance.
type OptimizedVPNManager struct {
	cache map[string]string
	mutex sync.RWMutex
}

// OptimizedDNSManager provides batch DNS operations for better performance.
type OptimizedDNSManager struct {
	cache map[string][]string
	mutex sync.RWMutex
}

// DNSConfig represents DNS configuration for batch operations.
type DNSConfig struct {
	Servers   []string
	Interface string
	Method    string
}

// NewOptimizedVPNManager creates a new optimized VPN manager.
func NewOptimizedVPNManager() *OptimizedVPNManager {
	return &OptimizedVPNManager{
		cache: make(map[string]string),
	}
}

// NewOptimizedDNSManager creates a new optimized DNS manager.
func NewOptimizedDNSManager() *OptimizedDNSManager {
	return &OptimizedDNSManager{
		cache: make(map[string][]string),
	}
}

// ConnectVPNBatch connects multiple VPNs efficiently.
func (m *OptimizedVPNManager) ConnectVPNBatch(configs []vpnConfig) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, config := range configs {
		if err := m.connectSingleVPN(config); err != nil {
			return fmt.Errorf("failed to connect VPN %s: %w", config.Name, err)
		}

		m.cache[config.Name] = "connected"
	}

	return nil
}

// GetVPNStatusBatch gets status for multiple VPNs efficiently.
func (m *OptimizedVPNManager) GetVPNStatusBatch(names []string) (map[string]string, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	status := make(map[string]string)

	// Check NetworkManager connections
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "nmcli", "-t", "-f", "NAME,STATE", "connection", "show", "--active")
	if output, err := cmd.Output(); err == nil {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		for _, line := range lines {
			fields := strings.Split(line, ":")
			if len(fields) >= 2 {
				name := fields[0]
				state := fields[1]

				for _, requestedName := range names {
					if name == requestedName {
						status[name] = state
					}
				}
			}
		}
	}

	// Fill in any missing with "disconnected"
	for _, name := range names {
		if _, exists := status[name]; !exists {
			status[name] = "disconnected"
		}
	}

	return status, nil
}

// connectSingleVPN connects a single VPN connection.
func (m *OptimizedVPNManager) connectSingleVPN(config vpnConfig) error {
	switch config.Type {
	case "networkmanager":
		ctx := context.Background()
		return exec.CommandContext(ctx, "nmcli", "connection", "up", config.Name).Run()
	case "openvpn":
		if config.Service != "" {
			ctx := context.Background()
			return exec.CommandContext(ctx, "systemctl", "start", config.Service).Run()
		}

		if config.ConfigFile != "" {
			ctx := context.Background()
			cmd := exec.CommandContext(ctx, "openvpn", "--config", config.ConfigFile, "--daemon")
			return cmd.Run()
		}

		return fmt.Errorf("openvpn requires either service or config file")
	case "wireguard":
		if config.ConfigFile != "" {
			ctx := context.Background()
			return exec.CommandContext(ctx, "wg-quick", "up", config.ConfigFile).Run()
		}

		ctx := context.Background()
		return exec.CommandContext(ctx, "wg-quick", "up", config.Name).Run()
	default:
		return fmt.Errorf("unsupported VPN type: %s", config.Type)
	}
}

// SetDNSServersBatch sets DNS servers for multiple configurations efficiently.
func (m *OptimizedDNSManager) SetDNSServersBatch(configs []DNSConfig) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, config := range configs {
		if err := m.setSingleDNS(config); err != nil {
			return fmt.Errorf("failed to set DNS servers %v: %w", config.Servers, err)
		}

		m.cache[config.Interface] = config.Servers
	}

	return nil
}

// setSingleDNS sets DNS servers for a single interface.
func (m *OptimizedDNSManager) setSingleDNS(config DNSConfig) error {
	iface := config.Interface
	if iface == "" {
		// Auto-detect primary interface
		if detectedIface, err := m.detectPrimaryInterface(); err == nil {
			iface = detectedIface
		} else {
			iface = "wlan0" // fallback
		}
	}

	switch config.Method {
	case "resolvectl", "":
		args := []string{"dns", iface}
		args = append(args, config.Servers...)

		ctx := context.Background()
		return exec.CommandContext(ctx, "resolvectl", args...).Run()
	case "networkmanager":
		servers := strings.Join(config.Servers, ",")
		ctx := context.Background()
		return exec.CommandContext(ctx, "nmcli", "connection", "modify", iface, "ipv4.dns", servers).Run()
	default:
		return fmt.Errorf("unsupported DNS method: %s", config.Method)
	}
}

// detectPrimaryInterface detects the primary network interface.
func (m *OptimizedDNSManager) detectPrimaryInterface() (string, error) {
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "ip", "route", "show", "default")

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// Parse output to find default interface
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "default via") {
			fields := strings.Fields(line)
			for i, field := range fields {
				if field == "dev" && i+1 < len(fields) {
					return fields[i+1], nil
				}
			}
		}
	}

	return "", fmt.Errorf("could not detect primary interface")
}

// executeShellCommand executes a shell command.
func executeShellCommand(command string) error {
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	return cmd.Run()
}
