package reports

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

// InterfaceSpeedDetector detects and caches network interface maximum speeds
type InterfaceSpeedDetector struct {
	cache map[string]uint64 // Interface name -> speed in bits per second
	mutex sync.RWMutex
}

// InterfaceInfo contains comprehensive information about a network interface
type InterfaceInfo struct {
	Name        string `json:"name"`
	MaxSpeed    uint64 `json:"max_speed_bps"`
	MaxSpeedStr string `json:"max_speed_human"`
	Status      string `json:"status"`
	Type        string `json:"type"`
	MTU         int    `json:"mtu"`
	Driver      string `json:"driver,omitempty"`
}

// NewInterfaceSpeedDetector creates a new interface speed detector
func NewInterfaceSpeedDetector() *InterfaceSpeedDetector {
	return &InterfaceSpeedDetector{
		cache: make(map[string]uint64),
	}
}

// GetInterfaceMaxSpeed returns the maximum speed of an interface in bits per second
func (isd *InterfaceSpeedDetector) GetInterfaceMaxSpeed(interfaceName string) (uint64, error) {
	isd.mutex.RLock()
	if speed, exists := isd.cache[interfaceName]; exists {
		isd.mutex.RUnlock()
		return speed, nil
	}
	isd.mutex.RUnlock()

	speed, err := isd.detectInterfaceSpeed(interfaceName)
	if err != nil {
		return 0, err
	}

	// Cache the result
	isd.mutex.Lock()
	isd.cache[interfaceName] = speed
	isd.mutex.Unlock()

	return speed, nil
}

// detectInterfaceSpeed detects interface speed using platform-specific methods
func (isd *InterfaceSpeedDetector) detectInterfaceSpeed(interfaceName string) (uint64, error) {
	switch runtime.GOOS {
	case "linux":
		return isd.detectSpeedLinux(interfaceName)
	case "darwin":
		return isd.detectSpeedDarwin(interfaceName)
	case "windows":
		return isd.detectSpeedWindows(interfaceName)
	default:
		return isd.estimateSpeedByType(interfaceName), nil
	}
}

// detectSpeedLinux detects interface speed on Linux systems
func (isd *InterfaceSpeedDetector) detectSpeedLinux(interfaceName string) (uint64, error) {
	// Method 1: Read from /sys/class/net/{interface}/speed
	speedFile := fmt.Sprintf("/sys/class/net/%s/speed", interfaceName)
	if data, err := os.ReadFile(speedFile); err == nil {
		speedStr := strings.TrimSpace(string(data))
		if speed, err := strconv.ParseUint(speedStr, 10, 64); err == nil {
			// Speed is in Mbps, convert to bps
			return speed * 1000000, nil
		}
	}

	// Method 2: Use ethtool command
	if speed, err := isd.getSpeedFromEthtool(interfaceName); err == nil {
		return speed, nil
	}

	// Method 3: Parse /proc/net/dev and estimate based on interface type
	return isd.estimateSpeedByType(interfaceName), nil
}

// detectSpeedDarwin detects interface speed on macOS systems
func (isd *InterfaceSpeedDetector) detectSpeedDarwin(interfaceName string) (uint64, error) {
	// Use ifconfig to get interface information
	cmd := exec.Command("ifconfig", interfaceName)
	output, err := cmd.Output()
	if err != nil {
		return isd.estimateSpeedByType(interfaceName), nil
	}

	// Parse ifconfig output for speed information
	if speed := isd.parseIfconfigSpeed(string(output)); speed > 0 {
		return speed, nil
	}

	return isd.estimateSpeedByType(interfaceName), nil
}

// detectSpeedWindows detects interface speed on Windows systems
func (isd *InterfaceSpeedDetector) detectSpeedWindows(interfaceName string) (uint64, error) {
	// Use WMI query or netsh command
	cmd := exec.Command("netsh", "interface", "show", "interface", interfaceName)
	output, err := cmd.Output()
	if err != nil {
		return isd.estimateSpeedByType(interfaceName), nil
	}

	// Parse netsh output for speed information
	if speed := isd.parseNetshSpeed(string(output)); speed > 0 {
		return speed, nil
	}

	return isd.estimateSpeedByType(interfaceName), nil
}

// getSpeedFromEthtool uses ethtool command to get interface speed
func (isd *InterfaceSpeedDetector) getSpeedFromEthtool(interfaceName string) (uint64, error) {
	cmd := exec.Command("ethtool", interfaceName)
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	return isd.parseEthtoolOutput(string(output))
}

// parseEthtoolOutput parses ethtool output to extract speed information
func (isd *InterfaceSpeedDetector) parseEthtoolOutput(output string) (uint64, error) {
	// Look for "Speed: XXXXX Mb/s" pattern
	re := regexp.MustCompile(`Speed:\s+(\d+)Mb/s`)
	matches := re.FindStringSubmatch(output)
	if len(matches) >= 2 {
		if speed, err := strconv.ParseUint(matches[1], 10, 64); err == nil {
			return speed * 1000000, nil // Convert Mbps to bps
		}
	}

	return 0, fmt.Errorf("could not parse ethtool output")
}

// parseIfconfigSpeed parses ifconfig output for macOS
func (isd *InterfaceSpeedDetector) parseIfconfigSpeed(output string) uint64 {
	// Look for speed indicators in ifconfig output
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Look for various speed indicators
		if strings.Contains(line, "1000baseT") || strings.Contains(line, "1000BaseTX") {
			return 1000000000 // 1 Gbps
		}
		if strings.Contains(line, "100baseTX") || strings.Contains(line, "100BaseT") {
			return 100000000 // 100 Mbps
		}
		if strings.Contains(line, "10baseT") || strings.Contains(line, "10BaseT") {
			return 10000000 // 10 Mbps
		}
	}

	return 0
}

// parseNetshSpeed parses netsh output for Windows
func (isd *InterfaceSpeedDetector) parseNetshSpeed(output string) uint64 {
	// Parse Windows netsh output for speed information
	// This is a simplified implementation
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "Speed") {
			// Try to extract speed from line
			re := regexp.MustCompile(`(\d+)\s*[Mm]bps`)
			matches := re.FindStringSubmatch(line)
			if len(matches) >= 2 {
				if speed, err := strconv.ParseUint(matches[1], 10, 64); err == nil {
					return speed * 1000000 // Convert Mbps to bps
				}
			}
		}
	}

	return 0
}

// estimateSpeedByType estimates speed based on interface name patterns
func (isd *InterfaceSpeedDetector) estimateSpeedByType(interfaceName string) uint64 {
	name := strings.ToLower(interfaceName)

	// Ethernet interfaces
	if strings.HasPrefix(name, "eth") || strings.HasPrefix(name, "en") {
		return 1000000000 // Default to 1 Gbps for ethernet
	}

	// WiFi interfaces
	if strings.HasPrefix(name, "wlan") || strings.HasPrefix(name, "wifi") ||
		strings.HasPrefix(name, "wl") || strings.Contains(name, "wireless") {
		return 300000000 // Default to 300 Mbps for WiFi
	}

	// Loopback interfaces
	if strings.HasPrefix(name, "lo") {
		return 10000000000 // 10 Gbps for loopback
	}

	// Virtual interfaces
	if strings.Contains(name, "docker") || strings.Contains(name, "veth") ||
		strings.Contains(name, "br-") || strings.Contains(name, "virbr") {
		return 10000000000 // 10 Gbps for virtual interfaces
	}

	// Default fallback
	return 1000000000 // 1 Gbps default
}

// CalculateUtilization calculates interface utilization percentage
func (isd *InterfaceSpeedDetector) CalculateUtilization(interfaceName string, rxRate, txRate uint64) (float64, error) {
	maxSpeed, err := isd.GetInterfaceMaxSpeed(interfaceName)
	if err != nil {
		return 0, err
	}

	if maxSpeed == 0 {
		return 0, fmt.Errorf("unknown interface speed for %s", interfaceName)
	}

	totalRate := rxRate + txRate
	utilization := (float64(totalRate) / float64(maxSpeed)) * 100

	// Cap at 100% to handle measurement variations
	if utilization > 100 {
		utilization = 100
	}

	return utilization, nil
}

// GetInterfaceInfo returns comprehensive information about an interface
func (isd *InterfaceSpeedDetector) GetInterfaceInfo(interfaceName string) (*InterfaceInfo, error) {
	maxSpeed, err := isd.GetInterfaceMaxSpeed(interfaceName)
	if err != nil {
		maxSpeed = isd.estimateSpeedByType(interfaceName)
	}

	info := &InterfaceInfo{
		Name:        interfaceName,
		MaxSpeed:    maxSpeed,
		MaxSpeedStr: isd.formatSpeed(maxSpeed),
		Type:        isd.getInterfaceType(interfaceName),
		Status:      "unknown",
	}

	// Try to get additional information based on OS
	if runtime.GOOS == "linux" {
		isd.enrichLinuxInterfaceInfo(info)
	}

	return info, nil
}

// formatSpeed formats speed in human-readable format
func (isd *InterfaceSpeedDetector) formatSpeed(speedBps uint64) string {
	const (
		Kbps = 1000
		Mbps = 1000 * Kbps
		Gbps = 1000 * Mbps
	)

	switch {
	case speedBps >= Gbps:
		return fmt.Sprintf("%.1f Gbps", float64(speedBps)/Gbps)
	case speedBps >= Mbps:
		return fmt.Sprintf("%.1f Mbps", float64(speedBps)/Mbps)
	case speedBps >= Kbps:
		return fmt.Sprintf("%.1f Kbps", float64(speedBps)/Kbps)
	default:
		return fmt.Sprintf("%d bps", speedBps)
	}
}

// getInterfaceType determines interface type based on name
func (isd *InterfaceSpeedDetector) getInterfaceType(interfaceName string) string {
	name := strings.ToLower(interfaceName)

	if strings.HasPrefix(name, "eth") || strings.HasPrefix(name, "en") {
		return "ethernet"
	}
	if strings.HasPrefix(name, "wlan") || strings.HasPrefix(name, "wifi") ||
		strings.HasPrefix(name, "wl") {
		return "wireless"
	}
	if strings.HasPrefix(name, "lo") {
		return "loopback"
	}
	if strings.Contains(name, "docker") || strings.Contains(name, "veth") ||
		strings.Contains(name, "br-") {
		return "virtual"
	}

	return "unknown"
}

// enrichLinuxInterfaceInfo adds Linux-specific interface information
func (isd *InterfaceSpeedDetector) enrichLinuxInterfaceInfo(info *InterfaceInfo) {
	basePath := fmt.Sprintf("/sys/class/net/%s", info.Name)

	// Get interface status
	if data, err := os.ReadFile(basePath + "/operstate"); err == nil {
		info.Status = strings.TrimSpace(string(data))
	}

	// Get MTU
	if data, err := os.ReadFile(basePath + "/mtu"); err == nil {
		if mtu, err := strconv.Atoi(strings.TrimSpace(string(data))); err == nil {
			info.MTU = mtu
		}
	}
}

// ClearCache clears the speed cache
func (isd *InterfaceSpeedDetector) ClearCache() {
	isd.mutex.Lock()
	defer isd.mutex.Unlock()
	isd.cache = make(map[string]uint64)
}
