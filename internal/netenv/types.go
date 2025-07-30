// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package netenv

import (
	"time"
)

// NetworkProfile represents a complete network environment configuration
type NetworkProfile struct {
	Name        string                 `yaml:"name" json:"name"`
	Description string                 `yaml:"description,omitempty" json:"description,omitempty"`
	Auto        bool                   `yaml:"auto,omitempty" json:"auto,omitempty"`
	Priority    int                    `yaml:"priority,omitempty" json:"priority,omitempty"`
	Conditions  []NetworkCondition     `yaml:"conditions,omitempty" json:"conditions,omitempty"`
	Components  NetworkComponents      `yaml:"components" json:"components"`
	Metadata    map[string]interface{} `yaml:"metadata,omitempty" json:"metadata,omitempty"`
	CreatedAt   time.Time              `yaml:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time              `yaml:"updatedAt" json:"updatedAt"`
}

// NetworkCondition defines when a profile should be automatically activated
type NetworkCondition struct {
	Type     string                 `yaml:"type" json:"type"` // wifi_ssid, ip_range, hostname, etc.
	Value    string                 `yaml:"value" json:"value"`
	Operator string                 `yaml:"operator,omitempty" json:"operator,omitempty"` // equals, contains, matches
	Metadata map[string]interface{} `yaml:"metadata,omitempty" json:"metadata,omitempty"`
}

// NetworkComponents contains all network component configurations
type NetworkComponents struct {
	WiFi       *WiFiConfig       `yaml:"wifi,omitempty" json:"wifi,omitempty"`
	VPN        *VPNConfig        `yaml:"vpn,omitempty" json:"vpn,omitempty"`
	DNS        *DNSConfig        `yaml:"dns,omitempty" json:"dns,omitempty"`
	Proxy      *ProxyConfig      `yaml:"proxy,omitempty" json:"proxy,omitempty"`
	Docker     *DockerConfig     `yaml:"docker,omitempty" json:"docker,omitempty"`
	Kubernetes *KubernetesConfig `yaml:"kubernetes,omitempty" json:"kubernetes,omitempty"`
	Hosts      *HostsConfig      `yaml:"hosts,omitempty" json:"hosts,omitempty"`
}

// WiFiConfig represents WiFi network configuration
type WiFiConfig struct {
	SSID     string `yaml:"ssid" json:"ssid"`
	Security string `yaml:"security,omitempty" json:"security,omitempty"`
	Priority int    `yaml:"priority,omitempty" json:"priority,omitempty"`
}

// VPNConfig represents VPN configuration
type VPNConfig struct {
	Name        string            `yaml:"name" json:"name"`
	Type        string            `yaml:"type" json:"type"` // openvpn, wireguard, cisco, etc.
	AutoConnect bool              `yaml:"autoConnect,omitempty" json:"autoConnect,omitempty"`
	Priority    int               `yaml:"priority,omitempty" json:"priority,omitempty"`
	Config      map[string]string `yaml:"config,omitempty" json:"config,omitempty"`
	Failover    []string          `yaml:"failover,omitempty" json:"failover,omitempty"`
	HealthCheck *HealthCheck      `yaml:"healthCheck,omitempty" json:"healthCheck,omitempty"`
}

// DNSConfig represents DNS configuration
type DNSConfig struct {
	Servers  []string          `yaml:"servers" json:"servers"`
	Domains  []string          `yaml:"domains,omitempty" json:"domains,omitempty"`
	Override bool              `yaml:"override,omitempty" json:"override,omitempty"`
	Fallback []string          `yaml:"fallback,omitempty" json:"fallback,omitempty"`
	Config   map[string]string `yaml:"config,omitempty" json:"config,omitempty"`
}

// ProxyConfig represents proxy configuration
type ProxyConfig struct {
	HTTP    string            `yaml:"http,omitempty" json:"http,omitempty"`
	HTTPS   string            `yaml:"https,omitempty" json:"https,omitempty"`
	FTP     string            `yaml:"ftp,omitempty" json:"ftp,omitempty"`
	NoProxy []string          `yaml:"noProxy,omitempty" json:"noProxy,omitempty"`
	Auth    *ProxyAuth        `yaml:"auth,omitempty" json:"auth,omitempty"`
	Config  map[string]string `yaml:"config,omitempty" json:"config,omitempty"`
}

// ProxyAuth represents proxy authentication
type ProxyAuth struct {
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
}

// DockerConfig represents Docker network configuration
type DockerConfig struct {
	Context  string            `yaml:"context,omitempty" json:"context,omitempty"`
	Networks []string          `yaml:"networks,omitempty" json:"networks,omitempty"`
	Config   map[string]string `yaml:"config,omitempty" json:"config,omitempty"`
}

// KubernetesConfig represents Kubernetes network configuration
type KubernetesConfig struct {
	Context   string            `yaml:"context,omitempty" json:"context,omitempty"`
	Namespace string            `yaml:"namespace,omitempty" json:"namespace,omitempty"`
	Policies  []string          `yaml:"policies,omitempty" json:"policies,omitempty"`
	Config    map[string]string `yaml:"config,omitempty" json:"config,omitempty"`
}

// HostsConfig represents hosts file configuration
type HostsConfig struct {
	Entries []HostEntry       `yaml:"entries,omitempty" json:"entries,omitempty"`
	Config  map[string]string `yaml:"config,omitempty" json:"config,omitempty"`
}

// HostEntry represents a single hosts file entry
type HostEntry struct {
	IP        string   `yaml:"ip" json:"ip"`
	Hostnames []string `yaml:"hostnames" json:"hostnames"`
	Comment   string   `yaml:"comment,omitempty" json:"comment,omitempty"`
}

// HealthCheck represents health check configuration
type HealthCheck struct {
	URL      string        `yaml:"url" json:"url"`
	Interval time.Duration `yaml:"interval" json:"interval"`
	Timeout  time.Duration `yaml:"timeout" json:"timeout"`
	Retries  int           `yaml:"retries,omitempty" json:"retries,omitempty"`
}

// NetworkStatus represents the current network status
type NetworkStatus struct {
	Profile    *NetworkProfile   `json:"profile,omitempty"`
	Components ComponentStatuses `json:"components"`
	Health     HealthStatus      `json:"health"`
	Metrics    *NetworkMetrics   `json:"metrics,omitempty"`
	LastSwitch time.Time         `json:"lastSwitch"`
}

// ComponentStatuses contains status for each network component
type ComponentStatuses struct {
	WiFi       *ComponentStatus `json:"wifi,omitempty"`
	VPN        *ComponentStatus `json:"vpn,omitempty"`
	DNS        *ComponentStatus `json:"dns,omitempty"`
	Proxy      *ComponentStatus `json:"proxy,omitempty"`
	Docker     *ComponentStatus `json:"docker,omitempty"`
	Kubernetes *ComponentStatus `json:"kubernetes,omitempty"`
}

// ComponentStatus represents the status of a network component
type ComponentStatus struct {
	Active    bool                   `json:"active"`
	Status    string                 `json:"status"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Error     string                 `json:"error,omitempty"`
	LastCheck time.Time              `json:"lastCheck"`
}

// HealthStatus represents overall network health
type HealthStatus struct {
	Status  string        `json:"status"` // excellent, good, poor, critical
	Score   int           `json:"score"`  // 0-100
	Issues  []string      `json:"issues,omitempty"`
	Latency time.Duration `json:"latency,omitempty"`
}

// NetworkMetrics represents network performance metrics
type NetworkMetrics struct {
	Latency    time.Duration `json:"latency"`
	Bandwidth  *Bandwidth    `json:"bandwidth,omitempty"`
	PacketLoss float64       `json:"packetLoss"`
	Jitter     time.Duration `json:"jitter,omitempty"`
}

// Bandwidth represents bandwidth measurements
type Bandwidth struct {
	Download float64 `json:"download"` // Mbps
	Upload   float64 `json:"upload"`   // Mbps
}

// QuickAction represents a quick network action
type QuickAction struct {
	Name        string            `yaml:"name" json:"name"`
	Description string            `yaml:"description,omitempty" json:"description,omitempty"`
	Component   string            `yaml:"component" json:"component"` // vpn, dns, proxy, etc.
	Action      string            `yaml:"action" json:"action"`       // on, off, toggle, reset, etc.
	Config      map[string]string `yaml:"config,omitempty" json:"config,omitempty"`
}
