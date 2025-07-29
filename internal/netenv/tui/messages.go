// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package tui

import (
	"time"
)

// Message types for the net-env TUI application
type (
	// TickMsg represents a timer tick for periodic updates
	TickMsg struct {
		Time time.Time
	}

	// NetworkStatusMsg represents an update to network status
	NetworkStatusMsg struct {
		Status NetworkStatus
	}

	// ErrorMsg represents an error
	ErrorMsg struct {
		Error error
	}

	// LoadingMsg represents loading state changes
	LoadingMsg struct {
		Loading bool
		Message string
	}

	// NavigationMsg represents navigation between views
	NavigationMsg struct {
		View ViewType
		Data interface{}
	}

	// ProfileSelectedMsg represents a network profile being selected
	ProfileSelectedMsg struct {
		ProfileName string
		Profile     *NetworkProfile
	}

	// VPNActionMsg represents VPN actions (connect/disconnect)
	VPNActionMsg struct {
		Action    string // "connect", "disconnect", "reconnect"
		VPNName   string
		Success   bool
		Error     error
		NewStatus *VPNStatus
	}

	// ProfileSwitchMsg represents profile switching
	ProfileSwitchMsg struct {
		ProfileName string
		Success     bool
		Error       error
		NewStatus   *NetworkStatus
	}

	// RefreshMsg represents a manual refresh request
	RefreshMsg struct{}

	// QuitMsg represents a quit request
	QuitMsg struct{}

	// WindowSizeMsg represents terminal window size changes
	WindowSizeMsg struct {
		Width  int
		Height int
	}

	// HelpToggleMsg represents help display toggle
	HelpToggleMsg struct{}

	// SearchMsg represents search functionality
	SearchMsg struct {
		Query   string
		Results []SearchResult
	}

	// FilterMsg represents filter functionality
	FilterMsg struct {
		Filter string
		Active bool
	}

	// MonitorUpdateMsg represents monitoring data updates
	MonitorUpdateMsg struct {
		Metrics NetworkMetrics
	}

	// AlertMsg represents network alerts
	AlertMsg struct {
		Alert NetworkAlert
	}
)

// SearchResult represents a search result item
type SearchResult struct {
	Type        string // "profile", "vpn", "action", "setting"
	Name        string
	Description string
	Action      func() error
}

// ViewType represents different views in the TUI
type ViewType int

const (
	ViewDashboard ViewType = iota
	ViewProfileSwitch
	ViewVPNManager
	ViewMonitoring
	ViewSettings
	ViewHelp
	ViewSearch
)

// String returns the string representation of a ViewType
func (v ViewType) String() string {
	switch v {
	case ViewDashboard:
		return "Dashboard"
	case ViewProfileSwitch:
		return "Profile Switch"
	case ViewVPNManager:
		return "VPN Manager"
	case ViewMonitoring:
		return "Network Monitoring"
	case ViewSettings:
		return "Settings"
	case ViewHelp:
		return "Help"
	case ViewSearch:
		return "Search"
	default:
		return "Unknown"
	}
}

// AppState represents the overall application state
type AppState int

const (
	StateLoading AppState = iota
	StateDashboard
	StateProfileSwitch
	StateVPNManager
	StateMonitoring
	StateSettings
	StateError
	StateHelp
	StateSearch
)

// String returns the string representation of an AppState
func (s AppState) String() string {
	switch s {
	case StateLoading:
		return "Loading"
	case StateDashboard:
		return "Dashboard"
	case StateProfileSwitch:
		return "Profile Switch"
	case StateVPNManager:
		return "VPN Manager"
	case StateMonitoring:
		return "Network Monitoring"
	case StateSettings:
		return "Settings"
	case StateError:
		return "Error"
	case StateHelp:
		return "Help"
	case StateSearch:
		return "Search"
	default:
		return "Unknown"
	}
}

// NetworkStatus represents the complete network status
type NetworkStatus struct {
	WiFi         WiFiStatus         `json:"wifi"`
	VPN          VPNStatus          `json:"vpn"`
	DNS          DNSStatus          `json:"dns"`
	Proxy        ProxyStatus        `json:"proxy"`
	Docker       DockerStatus       `json:"docker"`
	Connectivity ConnectivityStatus `json:"connectivity"`
	Timestamp    time.Time          `json:"timestamp"`
}

// WiFiStatus represents WiFi connection status
type WiFiStatus struct {
	SSID           string `json:"ssid"`
	SignalStrength int    `json:"signal_strength"`
	Frequency      string `json:"frequency"`
	Security       string `json:"security"`
	Connected      bool   `json:"connected"`
	IP             string `json:"ip,omitempty"`
	Gateway        string `json:"gateway,omitempty"`
}

// VPNStatus represents VPN connection status
type VPNStatus struct {
	Name        string        `json:"name"`
	Connected   bool          `json:"connected"`
	ServerIP    string        `json:"server_ip"`
	ClientIP    string        `json:"client_ip,omitempty"`
	Latency     time.Duration `json:"latency"`
	BytesUp     int64         `json:"bytes_up"`
	BytesDown   int64         `json:"bytes_down"`
	ConnectedAt time.Time     `json:"connected_at,omitempty"`
	Protocol    string        `json:"protocol,omitempty"`
}

// DNSStatus represents DNS configuration status
type DNSStatus struct {
	Servers   []string      `json:"servers"`
	Custom    bool          `json:"custom"`
	DoH       bool          `json:"doh"`       // DNS over HTTPS
	DoT       bool          `json:"dot"`       // DNS over TLS
	Response  time.Duration `json:"response"`  // Average response time
	Resolving bool          `json:"resolving"` // Whether DNS is resolving properly
}

// ProxyStatus represents proxy configuration status
type ProxyStatus struct {
	Enabled bool   `json:"enabled"`
	Type    string `json:"type"` // "http", "socks5", etc.
	Host    string `json:"host"`
	Port    int    `json:"port"`
	Auth    bool   `json:"auth"`    // Whether authentication is required
	Bypass  string `json:"bypass"`  // Bypass rules
	Working bool   `json:"working"` // Whether proxy is working
}

// DockerStatus represents Docker network status
type DockerStatus struct {
	Context   string            `json:"context"`
	Networks  []DockerNetwork   `json:"networks"`
	Connected bool              `json:"connected"`
	Details   map[string]string `json:"details,omitempty"`
}

// DockerNetwork represents a Docker network
type DockerNetwork struct {
	Name   string `json:"name"`
	Driver string `json:"driver"`
	Scope  string `json:"scope"`
	Active bool   `json:"active"`
}

// ConnectivityStatus represents overall connectivity status
type ConnectivityStatus struct {
	Internet   bool          `json:"internet"`
	Latency    time.Duration `json:"latency"`
	Bandwidth  BandwidthInfo `json:"bandwidth"`
	PacketLoss float64       `json:"packet_loss"`
	Quality    string        `json:"quality"` // "excellent", "good", "poor", "disconnected"
}

// BandwidthInfo represents bandwidth information
type BandwidthInfo struct {
	Download float64 `json:"download"` // Mbps
	Upload   float64 `json:"upload"`   // Mbps
}

// NetworkProfile represents a network configuration profile
type NetworkProfile struct {
	Name        string            `yaml:"name" json:"name"`
	Description string            `yaml:"description" json:"description"`
	WiFi        *WiFiConfig       `yaml:"wifi,omitempty" json:"wifi,omitempty"`
	VPN         *VPNConfig        `yaml:"vpn,omitempty" json:"vpn,omitempty"`
	DNS         *DNSConfig        `yaml:"dns,omitempty" json:"dns,omitempty"`
	Proxy       *ProxyConfig      `yaml:"proxy,omitempty" json:"proxy,omitempty"`
	Docker      *DockerConfig     `yaml:"docker,omitempty" json:"docker,omitempty"`
	Priority    int               `yaml:"priority" json:"priority"`
	AutoDetect  DetectionRule     `yaml:"auto_detect" json:"auto_detect"`
	CreatedAt   time.Time         `yaml:"created_at" json:"created_at"`
	UpdatedAt   time.Time         `yaml:"updated_at" json:"updated_at"`
	Tags        []string          `yaml:"tags,omitempty" json:"tags,omitempty"`
	Metadata    map[string]string `yaml:"metadata,omitempty" json:"metadata,omitempty"`
}

// WiFiConfig represents WiFi configuration
type WiFiConfig struct {
	SSID     string `yaml:"ssid" json:"ssid"`
	Password string `yaml:"password,omitempty" json:"password,omitempty"`
	Security string `yaml:"security,omitempty" json:"security,omitempty"`
	Hidden   bool   `yaml:"hidden,omitempty" json:"hidden,omitempty"`
}

// VPNConfig represents VPN configuration
type VPNConfig struct {
	Name       string            `yaml:"name" json:"name"`
	Type       string            `yaml:"type" json:"type"` // "openvpn", "wireguard", "ipsec"
	Server     string            `yaml:"server" json:"server"`
	Username   string            `yaml:"username,omitempty" json:"username,omitempty"`
	ConfigFile string            `yaml:"config_file,omitempty" json:"config_file,omitempty"`
	AutoStart  bool              `yaml:"auto_start,omitempty" json:"auto_start,omitempty"`
	Options    map[string]string `yaml:"options,omitempty" json:"options,omitempty"`
}

// DNSConfig represents DNS configuration
type DNSConfig struct {
	Servers []string `yaml:"servers" json:"servers"`
	DoH     bool     `yaml:"doh,omitempty" json:"doh,omitempty"`
	DoT     bool     `yaml:"dot,omitempty" json:"dot,omitempty"`
	Secure  bool     `yaml:"secure,omitempty" json:"secure,omitempty"`
}

// ProxyConfig represents proxy configuration
type ProxyConfig struct {
	Type     string `yaml:"type" json:"type"`
	Host     string `yaml:"host" json:"host"`
	Port     int    `yaml:"port" json:"port"`
	Username string `yaml:"username,omitempty" json:"username,omitempty"`
	Password string `yaml:"password,omitempty" json:"password,omitempty"`
	Bypass   string `yaml:"bypass,omitempty" json:"bypass,omitempty"`
}

// DockerConfig represents Docker configuration
type DockerConfig struct {
	Context  string            `yaml:"context" json:"context"`
	Networks []string          `yaml:"networks,omitempty" json:"networks,omitempty"`
	Options  map[string]string `yaml:"options,omitempty" json:"options,omitempty"`
}

// DetectionRule represents automatic detection rules
type DetectionRule struct {
	Conditions []Condition `yaml:"conditions" json:"conditions"`
	Priority   int         `yaml:"priority" json:"priority"`
}

// Condition represents a detection condition
type Condition struct {
	Type     string `yaml:"type" json:"type"`         // "wifi_ssid", "ip_range", "gateway", "dns"
	Value    string `yaml:"value" json:"value"`       // The value to match
	Operator string `yaml:"operator" json:"operator"` // "equals", "contains", "matches", "in_range"
}

// NetworkMetrics represents network performance metrics
type NetworkMetrics struct {
	Timestamp   time.Time     `json:"timestamp"`
	Latency     time.Duration `json:"latency"`
	PacketLoss  float64       `json:"packet_loss"`
	Bandwidth   BandwidthInfo `json:"bandwidth"`
	Connections int           `json:"connections"`
	Throughput  float64       `json:"throughput"` // Current throughput in Mbps
}

// NetworkAlert represents a network alert
type NetworkAlert struct {
	ID           string    `json:"id"`
	Timestamp    time.Time `json:"timestamp"`
	Type         string    `json:"type"`      // "warning", "error", "info"
	Component    string    `json:"component"` // "wifi", "vpn", "dns", "proxy"
	Message      string    `json:"message"`
	Severity     int       `json:"severity"` // 1-5, 5 being most severe
	Acknowledged bool      `json:"acknowledged"`
}
