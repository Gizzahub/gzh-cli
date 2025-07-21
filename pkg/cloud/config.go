// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package cloud

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the cloud configuration.
type Config struct {
	Version   string                    `yaml:"version"`
	Providers map[string]ProviderConfig `yaml:"providers"`
	Profiles  map[string]Profile        `yaml:"profiles"`
	Policies  []NetworkPolicy           `yaml:"policies,omitempty"`
	VPNs      map[string]VPNConnection  `yaml:"vpns,omitempty"`
	Sync      SyncConfig                `yaml:"sync,omitempty"`
}

// SyncConfig represents synchronization configuration.
type SyncConfig struct {
	Enabled      bool             `yaml:"enabled"`
	Interval     time.Duration    `yaml:"interval,omitempty"`
	Targets      []SyncTarget     `yaml:"targets,omitempty"`
	ConflictMode ConflictStrategy `yaml:"conflict_mode,omitempty"`
}

// SyncTarget represents a sync target configuration.
type SyncTarget struct {
	Source   string   `yaml:"source"`
	Target   string   `yaml:"target"`
	Profiles []string `yaml:"profiles,omitempty"` // empty means all profiles
}

// LoadConfig loads cloud configuration from file.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Set defaults
	if config.Version == "" {
		config.Version = "1.0"
	}

	if config.Sync.ConflictMode == "" {
		config.Sync.ConflictMode = ConflictStrategyAsk
	}

	if config.Sync.Interval == 0 {
		config.Sync.Interval = 1 * time.Hour
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// SaveConfig saves cloud configuration to file.
func SaveConfig(config *Config, path string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if len(c.Providers) == 0 {
		return fmt.Errorf("no providers configured")
	}

	// Validate providers
	for name, provider := range c.Providers {
		if provider.Type == "" {
			return fmt.Errorf("provider %s: type is required", name)
		}

		if !IsProviderSupported(provider.Type) {
			return fmt.Errorf("provider %s: unsupported type %s", name, provider.Type)
		}

		if provider.Region == "" {
			return fmt.Errorf("provider %s: region is required", name)
		}

		if provider.Auth.Method == "" {
			return fmt.Errorf("provider %s: auth method is required", name)
		}
	}

	// Validate profiles
	for name, profile := range c.Profiles {
		if profile.Provider == "" {
			return fmt.Errorf("profile %s: provider is required", name)
		}

		if _, exists := c.Providers[profile.Provider]; !exists {
			return fmt.Errorf("profile %s: unknown provider %s", name, profile.Provider)
		}

		if profile.Environment == "" {
			return fmt.Errorf("profile %s: environment is required", name)
		}
	}

	// Validate sync targets
	for i, target := range c.Sync.Targets {
		if target.Source == "" || target.Target == "" {
			return fmt.Errorf("sync target %d: source and target are required", i)
		}

		if _, exists := c.Providers[target.Source]; !exists {
			return fmt.Errorf("sync target %d: unknown source provider %s", i, target.Source)
		}

		if _, exists := c.Providers[target.Target]; !exists {
			return fmt.Errorf("sync target %d: unknown target provider %s", i, target.Target)
		}
	}

	return nil
}

// GetProvider returns provider configuration by name.
func (c *Config) GetProvider(name string) (ProviderConfig, bool) {
	provider, exists := c.Providers[name]
	return provider, exists
}

// GetProfile returns profile by name.
func (c *Config) GetProfile(name string) (Profile, bool) {
	profile, exists := c.Profiles[name]
	return profile, exists
}

// GetProfilesForProvider returns all profiles for a provider.
func (c *Config) GetProfilesForProvider(providerName string) []Profile {
	var profiles []Profile

	for _, profile := range c.Profiles {
		if profile.Provider == providerName {
			profiles = append(profiles, profile)
		}
	}

	return profiles
}

// GetVPNConnection returns VPN connection by name.
func (c *Config) GetVPNConnection(name string) (VPNConnection, bool) {
	if c.VPNs == nil {
		return VPNConnection{}, false
	}

	vpn, exists := c.VPNs[name]

	return vpn, exists
}

// AddVPNConnection adds a VPN connection to the configuration.
func (c *Config) AddVPNConnection(conn VPNConnection) {
	if c.VPNs == nil {
		c.VPNs = make(map[string]VPNConnection)
	}

	c.VPNs[conn.Name] = conn
}

// RemoveVPNConnection removes a VPN connection from the configuration.
func (c *Config) RemoveVPNConnection(name string) {
	if c.VPNs != nil {
		delete(c.VPNs, name)
	}
}

// GetVPNConnections returns all VPN connections.
func (c *Config) GetVPNConnections() map[string]VPNConnection {
	if c.VPNs == nil {
		return make(map[string]VPNConnection)
	}

	return c.VPNs
}

// GetDefaultConfigPath returns the default config file path.
func GetDefaultConfigPath() string {
	// Check environment variable first
	if path := os.Getenv("GZH_CLOUD_CONFIG"); path != "" {
		return path
	}

	// Check current directory
	if _, err := os.Stat("cloud-config.yaml"); err == nil {
		return "cloud-config.yaml"
	}

	// Use user config directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}

	return filepath.Join(configDir, "gzh-manager", "cloud-config.yaml")
}
