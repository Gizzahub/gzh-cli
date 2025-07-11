package netenv

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// DockerNetworkManager manages Docker network profiles and configurations
type DockerNetworkManager struct {
	logger      *zap.Logger
	profilesDir string
	mutex       sync.RWMutex
	cache       map[string]*DockerNetworkProfile
	executor    *DockerCommandExecutor
}

// DockerNetworkProfile represents a Docker network configuration profile
type DockerNetworkProfile struct {
	Name        string                       `yaml:"name" json:"name"`
	Description string                       `yaml:"description" json:"description"`
	Networks    map[string]*DockerNetwork    `yaml:"networks" json:"networks"`
	Containers  map[string]*ContainerNetwork `yaml:"containers" json:"containers"`
	Compose     *DockerComposeConfig         `yaml:"compose,omitempty" json:"compose,omitempty"`
	CreatedAt   time.Time                    `yaml:"created_at" json:"created_at"`
	UpdatedAt   time.Time                    `yaml:"updated_at" json:"updated_at"`
	Active      bool                         `yaml:"active" json:"active"`
	Metadata    map[string]string            `yaml:"metadata,omitempty" json:"metadata,omitempty"`
}

// DockerNetwork represents a Docker network configuration
type DockerNetwork struct {
	Name       string            `yaml:"name" json:"name"`
	Driver     string            `yaml:"driver" json:"driver"`
	Subnet     string            `yaml:"subnet,omitempty" json:"subnet,omitempty"`
	Gateway    string            `yaml:"gateway,omitempty" json:"gateway,omitempty"`
	Options    map[string]string `yaml:"options,omitempty" json:"options,omitempty"`
	Labels     map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
	External   bool              `yaml:"external,omitempty" json:"external,omitempty"`
	Attachable bool              `yaml:"attachable,omitempty" json:"attachable,omitempty"`
}

// ContainerNetwork represents container-specific network configuration
type ContainerNetwork struct {
	Image        string            `yaml:"image" json:"image"`
	NetworkMode  string            `yaml:"network_mode,omitempty" json:"network_mode,omitempty"`
	Networks     []string          `yaml:"networks,omitempty" json:"networks,omitempty"`
	Ports        []string          `yaml:"ports,omitempty" json:"ports,omitempty"`
	Environment  map[string]string `yaml:"environment,omitempty" json:"environment,omitempty"`
	DNSServers   []string          `yaml:"dns,omitempty" json:"dns,omitempty"`
	DNSSearch    []string          `yaml:"dns_search,omitempty" json:"dns_search,omitempty"`
	ExtraHosts   []string          `yaml:"extra_hosts,omitempty" json:"extra_hosts,omitempty"`
	Hostname     string            `yaml:"hostname,omitempty" json:"hostname,omitempty"`
	Domainname   string            `yaml:"domainname,omitempty" json:"domainname,omitempty"`
	NetworkAlias []string          `yaml:"network_alias,omitempty" json:"network_alias,omitempty"`
}

// DockerComposeConfig represents Docker Compose integration settings
type DockerComposeConfig struct {
	File        string            `yaml:"file,omitempty" json:"file,omitempty"`
	Project     string            `yaml:"project,omitempty" json:"project,omitempty"`
	Environment map[string]string `yaml:"environment,omitempty" json:"environment,omitempty"`
	Overrides   []string          `yaml:"overrides,omitempty" json:"overrides,omitempty"`
	AutoApply   bool              `yaml:"auto_apply" json:"auto_apply"`
}

// DockerNetworkStatus represents the current status of Docker networks
type DockerNetworkStatus struct {
	NetworkID  string            `json:"network_id"`
	Name       string            `json:"name"`
	Driver     string            `json:"driver"`
	Scope      string            `json:"scope"`
	Created    time.Time         `json:"created"`
	Containers map[string]string `json:"containers"` // container_id -> container_name
	Labels     map[string]string `json:"labels"`
	Options    map[string]string `json:"options"`
	IPAM       *IPAMConfig       `json:"ipam"`
}

// IPAMConfig represents IPAM (IP Address Management) configuration
type IPAMConfig struct {
	Driver  string            `json:"driver"`
	Config  []IPAMEntry       `json:"config"`
	Options map[string]string `json:"options"`
}

// IPAMEntry represents a single IPAM configuration entry
type IPAMEntry struct {
	Subnet     string            `json:"subnet"`
	Gateway    string            `json:"gateway"`
	IPRange    string            `json:"ip_range,omitempty"`
	AuxAddress map[string]string `json:"aux_address,omitempty"`
}

// ContainerNetworkInfo represents network information for a running container
type ContainerNetworkInfo struct {
	ContainerID string                      `json:"container_id"`
	Name        string                      `json:"name"`
	Image       string                      `json:"image"`
	State       string                      `json:"state"`
	Networks    map[string]*NetworkEndpoint `json:"networks"`
	Ports       []PortMapping               `json:"ports"`
	Created     time.Time                   `json:"created"`
	Labels      map[string]string           `json:"labels"`
}

// NetworkEndpoint represents a container's connection to a network
type NetworkEndpoint struct {
	NetworkID           string            `json:"network_id"`
	EndpointID          string            `json:"endpoint_id"`
	Gateway             string            `json:"gateway"`
	IPAddress           string            `json:"ip_address"`
	IPPrefixLen         int               `json:"ip_prefix_len"`
	IPv6Gateway         string            `json:"ipv6_gateway"`
	GlobalIPv6Address   string            `json:"global_ipv6_address"`
	GlobalIPv6PrefixLen int               `json:"global_ipv6_prefix_len"`
	MacAddress          string            `json:"mac_address"`
	DriverOpts          map[string]string `json:"driver_opts"`
	Aliases             []string          `json:"aliases"`
}

// PortMapping represents container port mapping
type PortMapping struct {
	PrivatePort int    `json:"private_port"`
	PublicPort  int    `json:"public_port,omitempty"`
	Type        string `json:"type"`
	IP          string `json:"ip,omitempty"`
}

// DockerCommandExecutor executes Docker commands with timeout and error handling
type DockerCommandExecutor struct {
	logger *zap.Logger
	cache  map[string]*DockerCommandResult
	mutex  sync.RWMutex
}

// DockerCommandResult represents the result of a Docker command execution
type DockerCommandResult struct {
	Output   string
	Error    string
	ExitCode int
	Duration time.Duration
	CachedAt time.Time
}

// NewDockerCommandExecutor creates a new Docker command executor
func NewDockerCommandExecutor(logger *zap.Logger) *DockerCommandExecutor {
	return &DockerCommandExecutor{
		logger: logger,
		cache:  make(map[string]*DockerCommandResult),
	}
}

// ExecuteWithTimeout executes a Docker command with timeout
func (dce *DockerCommandExecutor) ExecuteWithTimeout(ctx context.Context, command string, timeout time.Duration) (*DockerCommandResult, error) {
	// Check cache first (for read-only commands)
	if strings.HasPrefix(command, "docker inspect") || strings.HasPrefix(command, "docker network ls") || strings.HasPrefix(command, "docker ps") {
		if cached := dce.getCachedResult(command); cached != nil {
			return cached, nil
		}
	}

	// Create context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Parse command
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty command")
	}

	start := time.Now()
	cmd := exec.CommandContext(timeoutCtx, parts[0], parts[1:]...)
	output, err := cmd.CombinedOutput()
	duration := time.Since(start)

	result := &DockerCommandResult{
		Output:   string(output),
		Duration: duration,
		CachedAt: time.Now(),
	}

	if err != nil {
		result.Error = err.Error()
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		} else {
			result.ExitCode = 1
		}
	}

	// Cache read-only command results for 30 seconds
	if strings.HasPrefix(command, "docker inspect") || strings.HasPrefix(command, "docker network ls") || strings.HasPrefix(command, "docker ps") {
		dce.setCachedResult(command, result)
	}

	return result, nil
}

// getCachedResult retrieves a cached command result if still valid
func (dce *DockerCommandExecutor) getCachedResult(command string) *DockerCommandResult {
	dce.mutex.RLock()
	defer dce.mutex.RUnlock()

	if cached, exists := dce.cache[command]; exists {
		// Check if cache is still valid (30 seconds)
		if time.Since(cached.CachedAt) < 30*time.Second {
			return cached
		}
	}

	return nil
}

// setCachedResult stores a command result in cache
func (dce *DockerCommandExecutor) setCachedResult(command string, result *DockerCommandResult) {
	dce.mutex.Lock()
	defer dce.mutex.Unlock()

	dce.cache[command] = result
}

// NewDockerNetworkManager creates a new Docker network manager
func NewDockerNetworkManager(logger *zap.Logger, configDir string) *DockerNetworkManager {
	profilesDir := filepath.Join(configDir, "docker", "network_profiles")
	if err := os.MkdirAll(profilesDir, 0o755); err != nil {
		logger.Error("Failed to create Docker network profiles directory", zap.Error(err))
	}

	executor := NewDockerCommandExecutor(logger)

	return &DockerNetworkManager{
		logger:      logger,
		profilesDir: profilesDir,
		cache:       make(map[string]*DockerNetworkProfile),
		executor:    executor,
	}
}

// CreateProfile creates a new Docker network profile
func (dm *DockerNetworkManager) CreateProfile(profile *DockerNetworkProfile) error {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	if profile.Name == "" {
		return fmt.Errorf("profile name cannot be empty")
	}

	// Set timestamps
	now := time.Now()
	profile.CreatedAt = now
	profile.UpdatedAt = now

	// Validate networks
	if err := dm.validateNetworks(profile.Networks); err != nil {
		return fmt.Errorf("invalid network configuration: %w", err)
	}

	// Save to file
	profilePath := filepath.Join(dm.profilesDir, profile.Name+".yaml")
	data, err := yaml.Marshal(profile)
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}

	if err := os.WriteFile(profilePath, data, 0o644); err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	// Update cache
	dm.cache[profile.Name] = profile

	dm.logger.Info("Created Docker network profile",
		zap.String("name", profile.Name),
		zap.String("path", profilePath))

	return nil
}

// LoadProfile loads a Docker network profile
func (dm *DockerNetworkManager) LoadProfile(name string) (*DockerNetworkProfile, error) {
	dm.mutex.RLock()
	if cached, exists := dm.cache[name]; exists {
		dm.mutex.RUnlock()
		return cached, nil
	}
	dm.mutex.RUnlock()

	profilePath := filepath.Join(dm.profilesDir, name+".yaml")
	data, err := os.ReadFile(profilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read profile: %w", err)
	}

	var profile DockerNetworkProfile
	if err := yaml.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("failed to unmarshal profile: %w", err)
	}

	// Update cache
	dm.mutex.Lock()
	dm.cache[name] = &profile
	dm.mutex.Unlock()

	return &profile, nil
}

// ListProfiles lists all available Docker network profiles
func (dm *DockerNetworkManager) ListProfiles() ([]*DockerNetworkProfile, error) {
	files, err := filepath.Glob(filepath.Join(dm.profilesDir, "*.yaml"))
	if err != nil {
		return nil, fmt.Errorf("failed to list profile files: %w", err)
	}

	var profiles []*DockerNetworkProfile
	for _, file := range files {
		name := strings.TrimSuffix(filepath.Base(file), ".yaml")
		profile, err := dm.LoadProfile(name)
		if err != nil {
			dm.logger.Warn("Failed to load profile", zap.String("file", file), zap.Error(err))
			continue
		}
		profiles = append(profiles, profile)
	}

	return profiles, nil
}

// ApplyProfile applies a Docker network profile
func (dm *DockerNetworkManager) ApplyProfile(name string) error {
	profile, err := dm.LoadProfile(name)
	if err != nil {
		return fmt.Errorf("failed to load profile: %w", err)
	}

	dm.logger.Info("Applying Docker network profile", zap.String("name", name))

	// Create networks
	for networkName, network := range profile.Networks {
		if err := dm.createNetwork(networkName, network); err != nil {
			return fmt.Errorf("failed to create network %s: %w", networkName, err)
		}
	}

	// Apply container configurations
	for containerName, container := range profile.Containers {
		if err := dm.applyContainerNetworkConfig(containerName, container); err != nil {
			dm.logger.Warn("Failed to apply container network config",
				zap.String("container", containerName), zap.Error(err))
		}
	}

	// Apply Docker Compose configuration if present
	if profile.Compose != nil && profile.Compose.AutoApply {
		if err := dm.applyComposeConfig(profile.Compose); err != nil {
			return fmt.Errorf("failed to apply Docker Compose config: %w", err)
		}
	}

	// Mark profile as active
	profile.Active = true
	profile.UpdatedAt = time.Now()
	if err := dm.saveProfile(profile); err != nil {
		dm.logger.Warn("Failed to update profile status", zap.Error(err))
	}

	dm.logger.Info("Successfully applied Docker network profile", zap.String("name", name))
	return nil
}

// GetNetworkStatus returns the current status of Docker networks
func (dm *DockerNetworkManager) GetNetworkStatus() ([]*DockerNetworkStatus, error) {
	result, err := dm.executor.ExecuteWithTimeout(context.Background(), "docker network ls --format json", 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to list Docker networks: %w", err)
	}

	var networks []*DockerNetworkStatus
	lines := strings.Split(strings.TrimSpace(result.Output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		var network DockerNetworkStatus
		if err := json.Unmarshal([]byte(line), &network); err != nil {
			dm.logger.Warn("Failed to parse network info", zap.String("line", line), zap.Error(err))
			continue
		}

		// Get detailed network information
		if detailed, err := dm.getDetailedNetworkInfo(network.Name); err == nil {
			network = *detailed
		}

		networks = append(networks, &network)
	}

	return networks, nil
}

// GetContainerNetworkInfo returns network information for running containers
func (dm *DockerNetworkManager) GetContainerNetworkInfo() ([]*ContainerNetworkInfo, error) {
	result, err := dm.executor.ExecuteWithTimeout(context.Background(), "docker ps --format json", 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to list Docker containers: %w", err)
	}

	var containers []*ContainerNetworkInfo
	lines := strings.Split(strings.TrimSpace(result.Output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		var container ContainerNetworkInfo
		if err := json.Unmarshal([]byte(line), &container); err != nil {
			dm.logger.Warn("Failed to parse container info", zap.String("line", line), zap.Error(err))
			continue
		}

		// Get detailed container network information
		if detailed, err := dm.getDetailedContainerNetworkInfo(container.ContainerID); err == nil {
			container = *detailed
		}

		containers = append(containers, &container)
	}

	return containers, nil
}

// DetectDockerComposeProjects detects running Docker Compose projects
func (dm *DockerNetworkManager) DetectDockerComposeProjects() ([]string, error) {
	result, err := dm.executor.ExecuteWithTimeout(context.Background(), "docker ps --filter label=com.docker.compose.project --format '{{.Label \"com.docker.compose.project\"}}' | sort | uniq", 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to detect Docker Compose projects: %w", err)
	}

	projects := strings.Split(strings.TrimSpace(result.Output), "\n")
	var validProjects []string
	for _, project := range projects {
		if project != "" {
			validProjects = append(validProjects, project)
		}
	}

	return validProjects, nil
}

// CreateProfileFromCompose creates a network profile from an existing Docker Compose file
func (dm *DockerNetworkManager) CreateProfileFromCompose(composePath, profileName string) error {
	if _, err := os.Stat(composePath); os.IsNotExist(err) {
		return fmt.Errorf("Docker Compose file not found: %s", composePath)
	}

	// Parse Docker Compose file
	composeData, err := os.ReadFile(composePath)
	if err != nil {
		return fmt.Errorf("failed to read Docker Compose file: %w", err)
	}

	var compose map[string]interface{}
	if err := yaml.Unmarshal(composeData, &compose); err != nil {
		return fmt.Errorf("failed to parse Docker Compose file: %w", err)
	}

	// Extract network and service information
	profile := &DockerNetworkProfile{
		Name:        profileName,
		Description: fmt.Sprintf("Generated from Docker Compose file: %s", composePath),
		Networks:    make(map[string]*DockerNetwork),
		Containers:  make(map[string]*ContainerNetwork),
		Compose: &DockerComposeConfig{
			File:      composePath,
			AutoApply: false,
		},
	}

	// Extract networks
	if networks, ok := compose["networks"].(map[string]interface{}); ok {
		for name, netConfig := range networks {
			if netMap, ok := netConfig.(map[string]interface{}); ok {
				dockerNet := &DockerNetwork{
					Name:   name,
					Driver: "bridge", // Default driver
				}

				if driver, ok := netMap["driver"].(string); ok {
					dockerNet.Driver = driver
				}

				if ipam, ok := netMap["ipam"].(map[string]interface{}); ok {
					if config, ok := ipam["config"].([]interface{}); ok && len(config) > 0 {
						if configMap, ok := config[0].(map[string]interface{}); ok {
							if subnet, ok := configMap["subnet"].(string); ok {
								dockerNet.Subnet = subnet
							}
							if gateway, ok := configMap["gateway"].(string); ok {
								dockerNet.Gateway = gateway
							}
						}
					}
				}

				profile.Networks[name] = dockerNet
			}
		}
	}

	// Extract services
	if services, ok := compose["services"].(map[string]interface{}); ok {
		for name, serviceConfig := range services {
			if serviceMap, ok := serviceConfig.(map[string]interface{}); ok {
				container := &ContainerNetwork{}

				if image, ok := serviceMap["image"].(string); ok {
					container.Image = image
				}

				if networks, ok := serviceMap["networks"].([]interface{}); ok {
					for _, net := range networks {
						if netName, ok := net.(string); ok {
							container.Networks = append(container.Networks, netName)
						}
					}
				}

				if ports, ok := serviceMap["ports"].([]interface{}); ok {
					for _, port := range ports {
						if portStr, ok := port.(string); ok {
							container.Ports = append(container.Ports, portStr)
						}
					}
				}

				profile.Containers[name] = container
			}
		}
	}

	return dm.CreateProfile(profile)
}

// validateNetworks validates network configurations
func (dm *DockerNetworkManager) validateNetworks(networks map[string]*DockerNetwork) error {
	for name, network := range networks {
		if network.Name == "" {
			network.Name = name
		}

		if network.Driver == "" {
			network.Driver = "bridge"
		}

		// Validate driver
		validDrivers := []string{"bridge", "host", "overlay", "macvlan", "none"}
		valid := false
		for _, driver := range validDrivers {
			if network.Driver == driver {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid network driver: %s", network.Driver)
		}
	}

	return nil
}

// createNetwork creates a Docker network
func (dm *DockerNetworkManager) createNetwork(name string, network *DockerNetwork) error {
	// Check if network already exists
	if result, err := dm.executor.ExecuteWithTimeout(context.Background(), fmt.Sprintf("docker network inspect %s", name), 5*time.Second); err == nil && result.ExitCode == 0 {
		dm.logger.Info("Docker network already exists", zap.String("name", name))
		return nil
	}

	// Build create command
	createCmd := fmt.Sprintf("docker network create")

	if network.Driver != "" {
		createCmd += fmt.Sprintf(" --driver %s", network.Driver)
	}

	if network.Subnet != "" {
		createCmd += fmt.Sprintf(" --subnet %s", network.Subnet)
	}

	if network.Gateway != "" {
		createCmd += fmt.Sprintf(" --gateway %s", network.Gateway)
	}

	if network.Attachable {
		createCmd += " --attachable"
	}

	for key, value := range network.Options {
		createCmd += fmt.Sprintf(" --opt %s=%s", key, value)
	}

	for key, value := range network.Labels {
		createCmd += fmt.Sprintf(" --label %s=%s", key, value)
	}

	createCmd += fmt.Sprintf(" %s", name)

	// Execute create command
	result, err := dm.executor.ExecuteWithTimeout(context.Background(), createCmd, 30*time.Second)
	if err != nil {
		return fmt.Errorf("failed to create network: %w", err)
	}

	if result.ExitCode != 0 {
		return fmt.Errorf("docker network create failed: %s", result.Error)
	}

	dm.logger.Info("Created Docker network", zap.String("name", name))
	return nil
}

// applyContainerNetworkConfig applies network configuration to a container
func (dm *DockerNetworkManager) applyContainerNetworkConfig(containerName string, config *ContainerNetwork) error {
	// This is a placeholder for container network configuration
	// In a real implementation, this would modify running containers or prepare configurations
	dm.logger.Info("Applied container network configuration",
		zap.String("container", containerName),
		zap.String("image", config.Image))
	return nil
}

// applyComposeConfig applies Docker Compose configuration
func (dm *DockerNetworkManager) applyComposeConfig(config *DockerComposeConfig) error {
	if config.File == "" {
		return fmt.Errorf("Docker Compose file not specified")
	}

	composeCmd := fmt.Sprintf("docker-compose -f %s", config.File)

	if config.Project != "" {
		composeCmd += fmt.Sprintf(" -p %s", config.Project)
	}

	// Set environment variables
	env := os.Environ()
	for key, value := range config.Environment {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}

	composeCmd += " up -d"

	result, err := dm.executor.ExecuteWithTimeout(context.Background(), composeCmd, 60*time.Second)
	if err != nil {
		return fmt.Errorf("failed to execute docker-compose: %w", err)
	}

	if result.ExitCode != 0 {
		return fmt.Errorf("docker-compose failed: %s", result.Error)
	}

	dm.logger.Info("Applied Docker Compose configuration", zap.String("file", config.File))
	return nil
}

// getDetailedNetworkInfo gets detailed information about a specific network
func (dm *DockerNetworkManager) getDetailedNetworkInfo(networkName string) (*DockerNetworkStatus, error) {
	result, err := dm.executor.ExecuteWithTimeout(context.Background(), fmt.Sprintf("docker network inspect %s", networkName), 10*time.Second)
	if err != nil {
		return nil, err
	}

	var networkDetails []map[string]interface{}
	if err := json.Unmarshal([]byte(result.Output), &networkDetails); err != nil {
		return nil, err
	}

	if len(networkDetails) == 0 {
		return nil, fmt.Errorf("network not found")
	}

	detail := networkDetails[0]
	status := &DockerNetworkStatus{
		Name:       networkName,
		Labels:     make(map[string]string),
		Options:    make(map[string]string),
		Containers: make(map[string]string),
	}

	if id, ok := detail["Id"].(string); ok {
		status.NetworkID = id
	}

	if driver, ok := detail["Driver"].(string); ok {
		status.Driver = driver
	}

	if scope, ok := detail["Scope"].(string); ok {
		status.Scope = scope
	}

	if created, ok := detail["Created"].(string); ok {
		if t, err := time.Parse(time.RFC3339, created); err == nil {
			status.Created = t
		}
	}

	return status, nil
}

// getDetailedContainerNetworkInfo gets detailed network information for a container
func (dm *DockerNetworkManager) getDetailedContainerNetworkInfo(containerID string) (*ContainerNetworkInfo, error) {
	result, err := dm.executor.ExecuteWithTimeout(context.Background(), fmt.Sprintf("docker inspect %s", containerID), 10*time.Second)
	if err != nil {
		return nil, err
	}

	var containerDetails []map[string]interface{}
	if err := json.Unmarshal([]byte(result.Output), &containerDetails); err != nil {
		return nil, err
	}

	if len(containerDetails) == 0 {
		return nil, fmt.Errorf("container not found")
	}

	detail := containerDetails[0]
	info := &ContainerNetworkInfo{
		ContainerID: containerID,
		Networks:    make(map[string]*NetworkEndpoint),
		Labels:      make(map[string]string),
	}

	if name, ok := detail["Name"].(string); ok {
		info.Name = strings.TrimPrefix(name, "/")
	}

	if config, ok := detail["Config"].(map[string]interface{}); ok {
		if image, ok := config["Image"].(string); ok {
			info.Image = image
		}
	}

	if state, ok := detail["State"].(map[string]interface{}); ok {
		if status, ok := state["Status"].(string); ok {
			info.State = status
		}
	}

	if created, ok := detail["Created"].(string); ok {
		if t, err := time.Parse(time.RFC3339, created); err == nil {
			info.Created = t
		}
	}

	return info, nil
}

// saveProfile saves a profile to disk
func (dm *DockerNetworkManager) saveProfile(profile *DockerNetworkProfile) error {
	profilePath := filepath.Join(dm.profilesDir, profile.Name+".yaml")
	data, err := yaml.Marshal(profile)
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}

	if err := os.WriteFile(profilePath, data, 0o644); err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	// Update cache
	dm.mutex.Lock()
	dm.cache[profile.Name] = profile
	dm.mutex.Unlock()

	return nil
}

// DeleteProfile deletes a Docker network profile
func (dm *DockerNetworkManager) DeleteProfile(name string) error {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	profilePath := filepath.Join(dm.profilesDir, name+".yaml")
	if err := os.Remove(profilePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete profile file: %w", err)
	}

	delete(dm.cache, name)
	dm.logger.Info("Deleted Docker network profile", zap.String("name", name))
	return nil
}
