// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package netenv

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ContainerRuntime represents different container runtime types.
type ContainerRuntime string

const (
	Docker     ContainerRuntime = "docker"
	Podman     ContainerRuntime = "podman"
	Containerd ContainerRuntime = "containerd"
	Nerdctl    ContainerRuntime = "nerdctl"

	kubernetesDefaultNamespace = "default"
)

// ContainerEnvironment represents the detected container environment.
type ContainerEnvironment struct {
	AvailableRuntimes      []RuntimeInfo           `json:"availableRuntimes"`
	PrimaryRuntime         ContainerRuntime        `json:"primaryRuntime"`
	OrchestrationPlatform  string                  `json:"orchestrationPlatform"` // docker-swarm, kubernetes, standalone
	RunningContainers      []DetectedContainer     `json:"runningContainers"`
	Networks               []DetectedNetwork       `json:"networks"`
	ComposeProjects        []ComposeProject        `json:"composeProjects"`
	KubernetesInfo         *KubernetesClusterInfo  `json:"kubernetesInfo,omitempty"`
	ResourceUsage          *ContainerResourceUsage `json:"resourceUsage"`
	DetectedAt             time.Time               `json:"detectedAt"`
	EnvironmentFingerprint string                  `json:"environmentFingerprint"`
}

// RuntimeInfo represents information about a container runtime.
// nolint:tagliatelle // External API format - must match container runtime JSON output
type RuntimeInfo struct {
	Runtime    ContainerRuntime `json:"runtime"`
	Version    string           `json:"version"`
	Available  bool             `json:"available"`
	Executable string           `json:"executable"`
	ServerInfo *ServerInfo      `json:"server_info,omitempty"`
}

// ServerInfo represents container runtime server information.
// nolint:tagliatelle // External API format - must match container runtime JSON output
type ServerInfo struct {
	Version       string            `json:"version"`
	OS            string            `json:"os"`
	Architecture  string            `json:"architecture"`
	KernelVersion string            `json:"kernel_version"`
	TotalMemory   string            `json:"total_memory"`
	CPUs          int               `json:"cpus"`
	StorageDriver string            `json:"storage_driver"`
	LoggingDriver string            `json:"logging_driver"`
	CgroupDriver  string            `json:"cgroup_driver"`
	RuntimeConfig map[string]string `json:"runtime_config"`
}

// DetectedContainer represents a detected running container.
// nolint:tagliatelle // External API format - must match container runtime JSON output
type DetectedContainer struct {
	ID             string                  `json:"id"`
	Name           string                  `json:"name"`
	Image          string                  `json:"image"`
	ImageID        string                  `json:"image_id"`
	Status         string                  `json:"status"`
	State          string                  `json:"state"`
	Runtime        ContainerRuntime        `json:"runtime"`
	Created        time.Time               `json:"created"`
	StartedAt      time.Time               `json:"started_at"`
	Ports          []DetectedPortMapping   `json:"ports"`
	Networks       []DetectedNetworkInfo   `json:"networks"`
	Labels         map[string]string       `json:"labels"`
	Environment    []string                `json:"environment"`
	Mounts         []DetectedMount         `json:"mounts"`
	ResourceLimits *DetectedResourceLimits `json:"resource_limits,omitempty"`
	HealthStatus   string                  `json:"health_status"`
	RestartPolicy  string                  `json:"restart_policy"`
	WorkingDir     string                  `json:"working_dir"`
	Command        []string                `json:"command"`
	Args           []string                `json:"args"`
}

// DetectedPortMapping represents container port mappings.
type DetectedPortMapping struct {
	ContainerPort int32  `json:"containerPort"`
	HostPort      int32  `json:"hostPort"`
	HostIP        string `json:"hostIp"`
	Protocol      string `json:"protocol"`
}

// DetectedNetworkInfo represents container network information.
type DetectedNetworkInfo struct {
	NetworkName string `json:"networkName"`
	NetworkID   string `json:"networkId"`
	IPAddress   string `json:"ipAddress"`
	MacAddress  string `json:"macAddress"`
	Gateway     string `json:"gateway"`
	Subnet      string `json:"subnet"`
}

// DetectedMount represents container mount information.
type DetectedMount struct {
	Type        string `json:"type"` // bind, volume, tmpfs
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Mode        string `json:"mode"`
	ReadWrite   bool   `json:"readWrite"`
}

// DetectedResourceLimits represents container resource limits.
type DetectedResourceLimits struct {
	CPUShares       int64 `json:"cpuShares"`
	CPUQuota        int64 `json:"cpuQuota"`
	CPUPeriod       int64 `json:"cpuPeriod"`
	Memory          int64 `json:"memory"`
	MemorySwap      int64 `json:"memorySwap"`
	BlkioWeight     int   `json:"blkioWeight"`
	OomKillDisabled bool  `json:"oomKillDisabled"`
}

// DetectedNetwork represents detected container network.
type DetectedNetwork struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Driver   string            `json:"driver"`
	Scope    string            `json:"scope"`
	Internal bool              `json:"internal"`
	IPAM     NetworkIPAM       `json:"ipam"`
	Options  map[string]string `json:"options"`
	Labels   map[string]string `json:"labels"`
	Created  time.Time         `json:"created"`
	Runtime  ContainerRuntime  `json:"runtime"`
}

// NetworkIPAM represents network IP address management.
type NetworkIPAM struct {
	Driver  string                `json:"driver"`
	Config  []ContainerIPAMConfig `json:"config"`
	Options map[string]string     `json:"options"`
}

// ContainerIPAMConfig represents IPAM configuration for container detection.
type ContainerIPAMConfig struct {
	Subnet  string `json:"subnet"`
	Gateway string `json:"gateway"`
	IPRange string `json:"ipRange,omitempty"`
}

// ComposeProject represents a detected Docker Compose project.
type ComposeProject struct {
	Name        string            `json:"name"`
	Services    []ComposeService  `json:"services"`
	Networks    []string          `json:"networks"`
	Volumes     []string          `json:"volumes"`
	ConfigPath  string            `json:"configPath,omitempty"`
	Environment map[string]string `json:"environment"`
	Runtime     ContainerRuntime  `json:"runtime"`
}

// ComposeService represents a service in a Compose project.
type ComposeService struct {
	Name       string                `json:"name"`
	Image      string                `json:"image"`
	Containers []string              `json:"containers"`
	Replicas   int                   `json:"replicas"`
	Ports      []DetectedPortMapping `json:"ports"`
	Labels     map[string]string     `json:"labels"`
}

// KubernetesClusterInfo represents Kubernetes cluster information.
type KubernetesClusterInfo struct {
	Available          bool             `json:"available"`
	Version            string           `json:"version"`
	Context            string           `json:"context"`
	Namespace          string           `json:"namespace"`
	Nodes              []KubernetesNode `json:"nodes"`
	Namespaces         []string         `json:"namespaces"`
	ServiceMesh        *ServiceMeshInfo `json:"serviceMesh,omitempty"`
	IngressControllers []string         `json:"ingressControllers"`
}

// KubernetesNode represents a Kubernetes node.
type KubernetesNode struct {
	Name      string            `json:"name"`
	Status    string            `json:"status"`
	Roles     []string          `json:"roles"`
	Version   string            `json:"version"`
	OS        string            `json:"os"`
	Arch      string            `json:"arch"`
	Addresses map[string]string `json:"addresses"`
}

// ServiceMeshInfo represents service mesh information.
type ServiceMeshInfo struct {
	Type      string `json:"type"` // istio, linkerd, consul-connect
	Version   string `json:"version"`
	Namespace string `json:"namespace"`
	Enabled   bool   `json:"enabled"`
}

// ContainerResourceUsage represents overall container resource usage.
type ContainerResourceUsage struct {
	TotalContainers   int             `json:"totalContainers"`
	RunningContainers int             `json:"runningContainers"`
	StoppedContainers int             `json:"stoppedContainers"`
	Images            int             `json:"images"`
	Networks          int             `json:"networks"`
	Volumes           int             `json:"volumes"`
	ResourceSummary   ResourceSummary `json:"resourceSummary"`
}

// ResourceSummary represents resource usage summary.
type ResourceSummary struct {
	CPUUsage    float64 `json:"cpuUsagePercent"`
	MemoryUsage int64   `json:"memoryUsageBytes"`
	MemoryLimit int64   `json:"memoryLimitBytes"`
	NetworkRx   int64   `json:"networkRxBytes"`
	NetworkTx   int64   `json:"networkTxBytes"`
	BlockRead   int64   `json:"blockReadBytes"`
	BlockWrite  int64   `json:"blockWriteBytes"`
}

// ContainerDetector detects and analyzes container environments.
type ContainerDetector struct {
	logger            *zap.Logger
	cachedEnvironment *ContainerEnvironment
	cacheMutex        sync.RWMutex
	cacheExpiry       time.Duration
	lastDetection     time.Time
}

// NewContainerDetector creates a new container detector.
func NewContainerDetector(logger *zap.Logger) *ContainerDetector {
	return &ContainerDetector{
		logger:      logger,
		cacheExpiry: 30 * time.Second, // Cache results for 30 seconds
	}
}

// DetectContainerEnvironment detects the current container environment.
func (cd *ContainerDetector) DetectContainerEnvironment(ctx context.Context) (*ContainerEnvironment, error) {
	cd.cacheMutex.RLock()

	if cd.cachedEnvironment != nil && time.Since(cd.lastDetection) < cd.cacheExpiry {
		cd.cacheMutex.RUnlock()
		return cd.cachedEnvironment, nil
	}

	cd.cacheMutex.RUnlock()

	cd.logger.Info("Detecting container environment")

	env := &ContainerEnvironment{
		DetectedAt: time.Now(),
	}

	// Detect available container runtimes
	runtimes, err := cd.detectContainerRuntimes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to detect container runtimes: %w", err)
	}

	env.AvailableRuntimes = runtimes

	// Determine primary runtime
	env.PrimaryRuntime = cd.determinePrimaryRuntime(runtimes)

	// Detect orchestration platform
	env.OrchestrationPlatform = cd.detectOrchestrationPlatform(ctx)

	// Detect running containers using primary runtime
	if env.PrimaryRuntime != "" {
		containers, err := cd.detectRunningContainers(ctx, env.PrimaryRuntime)
		if err != nil {
			cd.logger.Warn("Failed to detect running containers", zap.Error(err))
		} else {
			env.RunningContainers = containers
		}

		// Detect networks
		networks, err := cd.detectNetworks(ctx, env.PrimaryRuntime)
		if err != nil {
			cd.logger.Warn("Failed to detect networks", zap.Error(err))
		} else {
			env.Networks = networks
		}

		// Detect Compose projects
		composeProjects, err := cd.detectComposeProjects(ctx, env.PrimaryRuntime)
		if err != nil {
			cd.logger.Warn("Failed to detect Compose projects", zap.Error(err))
		} else {
			env.ComposeProjects = composeProjects
		}
	}

	// Detect Kubernetes if available
	k8sInfo, err := cd.detectKubernetesInfo(ctx)
	if err != nil {
		cd.logger.Debug("Kubernetes not available", zap.Error(err))
	} else {
		env.KubernetesInfo = k8sInfo
	}

	// Calculate resource usage
	resourceUsage := cd.calculateResourceUsage(ctx, env)
	env.ResourceUsage = resourceUsage

	// Generate environment fingerprint
	env.EnvironmentFingerprint = cd.generateEnvironmentFingerprint(env)

	// Cache the result
	cd.cacheMutex.Lock()
	cd.cachedEnvironment = env
	cd.lastDetection = time.Now()
	cd.cacheMutex.Unlock()

	cd.logger.Info("Container environment detection completed",
		zap.String("primary_runtime", string(env.PrimaryRuntime)),
		zap.String("orchestration", env.OrchestrationPlatform),
		zap.Int("containers", len(env.RunningContainers)),
		zap.Int("networks", len(env.Networks)))

	return env, nil
}

// detectContainerRuntimes detects available container runtimes.
func (cd *ContainerDetector) detectContainerRuntimes(ctx context.Context) ([]RuntimeInfo, error) {
	runtimes := []RuntimeInfo{}
	candidates := []ContainerRuntime{Docker, Podman, Nerdctl}

	for _, runtime := range candidates {
		info := RuntimeInfo{
			Runtime: runtime,
		}

		// Check if runtime executable is available
		executable, err := exec.LookPath(string(runtime))
		if err != nil {
			cd.logger.Debug("Runtime not found", zap.String("runtime", string(runtime)))

			info.Available = false
			runtimes = append(runtimes, info)

			continue
		}

		info.Executable = executable
		info.Available = true

		// Get version information
		version, err := cd.getRuntimeVersion(ctx, runtime)
		if err != nil {
			cd.logger.Warn("Failed to get runtime version",
				zap.String("runtime", string(runtime)),
				zap.Error(err))
		} else {
			info.Version = version
		}

		// Get server info if available
		serverInfo, err := cd.getRuntimeServerInfo(ctx, runtime)
		if err != nil {
			cd.logger.Debug("Failed to get runtime server info",
				zap.String("runtime", string(runtime)),
				zap.Error(err))
		} else {
			info.ServerInfo = serverInfo
		}

		runtimes = append(runtimes, info)

		cd.logger.Info("Detected container runtime",
			zap.String("runtime", string(runtime)),
			zap.String("version", version))
	}

	return runtimes, nil
}

// getRuntimeVersion gets the version of a container runtime.
func (cd *ContainerDetector) getRuntimeVersion(ctx context.Context, runtime ContainerRuntime) (string, error) {
	cmd := exec.CommandContext(ctx, string(runtime), "version", "--format", "{{.Client.Version}}")

	output, err := cmd.Output()
	if err != nil {
		// Try alternative version command
		cmd = exec.CommandContext(ctx, string(runtime), "--version")

		output, err = cmd.Output()
		if err != nil {
			return "", err
		}
	}

	version := strings.TrimSpace(string(output))
	// Extract version number from output like "Docker version 20.10.8, build 3967b7d"
	if strings.Contains(version, "version") {
		parts := strings.Fields(version)
		for i, part := range parts {
			if part == "version" && i+1 < len(parts) {
				version = strings.TrimSuffix(parts[i+1], ",")
				break
			}
		}
	}

	return version, nil
}

// getRuntimeServerInfo gets server information from container runtime.
func (cd *ContainerDetector) getRuntimeServerInfo(ctx context.Context, runtime ContainerRuntime) (*ServerInfo, error) {
	var cmd *exec.Cmd

	switch runtime {
	case Docker, Podman:
		cmd = exec.CommandContext(ctx, string(runtime), "system", "info", "--format", "json")
	case Nerdctl:
		cmd = exec.CommandContext(ctx, string(runtime), "system", "info")
	case Containerd:
		cmd = exec.CommandContext(ctx, "ctr", "version")
	default:
		return nil, fmt.Errorf("unsupported runtime: %s", runtime)
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// Parse JSON output for Docker/Podman
	if runtime == Docker || runtime == Podman {
		var info map[string]interface{}
		if err := json.Unmarshal(output, &info); err != nil {
			return nil, err
		}

		serverInfo := &ServerInfo{
			RuntimeConfig: make(map[string]string),
		}

		// Extract common fields
		if version, ok := info["ServerVersion"].(string); ok {
			serverInfo.Version = version
		}

		if os, ok := info["OperatingSystem"].(string); ok {
			serverInfo.OS = os
		}

		if arch, ok := info["Architecture"].(string); ok {
			serverInfo.Architecture = arch
		}

		if kernel, ok := info["KernelVersion"].(string); ok {
			serverInfo.KernelVersion = kernel
		}

		if cpus, ok := info["NCPU"].(float64); ok {
			serverInfo.CPUs = int(cpus)
		}

		if memory, ok := info["MemTotal"].(float64); ok {
			serverInfo.TotalMemory = fmt.Sprintf("%.0f", memory)
		}

		if storage, ok := info["Driver"].(string); ok {
			serverInfo.StorageDriver = storage
		}

		if logging, ok := info["LoggingDriver"].(string); ok {
			serverInfo.LoggingDriver = logging
		}

		if cgroup, ok := info["CgroupDriver"].(string); ok {
			serverInfo.CgroupDriver = cgroup
		}

		return serverInfo, nil
	}

	// For nerdctl, parse text output
	lines := strings.Split(string(output), "\n")
	serverInfo := &ServerInfo{
		RuntimeConfig: make(map[string]string),
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				serverInfo.RuntimeConfig[key] = value
			}
		}
	}

	return serverInfo, nil
}

// determinePrimaryRuntime determines the primary container runtime to use.
func (cd *ContainerDetector) determinePrimaryRuntime(runtimes []RuntimeInfo) ContainerRuntime {
	// Priority order: Docker > Podman > nerdctl
	priorities := []ContainerRuntime{Docker, Podman, Nerdctl}

	for _, priority := range priorities {
		for _, runtime := range runtimes {
			if runtime.Runtime == priority && runtime.Available && runtime.ServerInfo != nil {
				return runtime.Runtime
			}
		}
	}

	// Fallback to any available runtime
	for _, runtime := range runtimes {
		if runtime.Available {
			return runtime.Runtime
		}
	}

	return ""
}

// detectOrchestrationPlatform detects the container orchestration platform.
func (cd *ContainerDetector) detectOrchestrationPlatform(ctx context.Context) string {
	// Check for Kubernetes
	if cd.isKubernetesAvailable(ctx) {
		return "kubernetes"
	}

	// Check for Docker Swarm
	if cd.isDockerSwarmAvailable(ctx) {
		return "docker-swarm"
	}

	return "standalone"
}

// isKubernetesAvailable checks if Kubernetes is available.
func (cd *ContainerDetector) isKubernetesAvailable(ctx context.Context) bool {
	cmd := exec.CommandContext(ctx, "kubectl", "cluster-info")
	return cmd.Run() == nil
}

// isDockerSwarmAvailable checks if Docker Swarm is available.
func (cd *ContainerDetector) isDockerSwarmAvailable(ctx context.Context) bool {
	cmd := exec.CommandContext(ctx, "docker", "node", "ls")
	return cmd.Run() == nil
}

// detectRunningContainers detects running containers using the specified runtime.
func (cd *ContainerDetector) detectRunningContainers(ctx context.Context, runtime ContainerRuntime) ([]DetectedContainer, error) {
	var cmd *exec.Cmd

	switch runtime {
	case Docker, Podman:
		cmd = exec.CommandContext(ctx, string(runtime), "ps", "--format", "json")
	case Nerdctl:
		cmd = exec.CommandContext(ctx, string(runtime), "ps", "--format", "json")
	case Containerd:
		cmd = exec.CommandContext(ctx, "ctr", "containers", "ls", "--format", "json")
	default:
		return nil, fmt.Errorf("unsupported runtime: %s", runtime)
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	containers := make([]DetectedContainer, 0, len(lines))

	for _, line := range lines {
		if line == "" {
			continue
		}

		var psContainer map[string]interface{}
		if err := json.Unmarshal([]byte(line), &psContainer); err != nil {
			cd.logger.Warn("Failed to parse container JSON", zap.Error(err))
			continue
		}

		container := DetectedContainer{
			Runtime: runtime,
		}

		// Extract common fields from ps output
		if id, ok := psContainer["ID"].(string); ok {
			container.ID = id
		}

		if name, ok := psContainer["Names"].(string); ok {
			container.Name = name
		}

		if image, ok := psContainer["Image"].(string); ok {
			container.Image = image
		}

		if status, ok := psContainer["Status"].(string); ok {
			container.Status = status
		}

		if state, ok := psContainer["State"].(string); ok {
			container.State = state
		}

		// Get detailed information using inspect
		detailed, err := cd.inspectContainer(ctx, runtime, container.ID)
		if err != nil {
			cd.logger.Warn("Failed to inspect container",
				zap.String("id", container.ID),
				zap.Error(err))
		} else {
			// Merge detailed information
			container = cd.mergeContainerInfo(container, detailed)
		}

		containers = append(containers, container)
	}

	return containers, nil
}

// inspectContainer gets detailed container information.
func (cd *ContainerDetector) inspectContainer(ctx context.Context, runtime ContainerRuntime, containerID string) (*DetectedContainer, error) { //nolint:gocognit // Complex container inspection - requires architectural refactoring
	cmd := exec.CommandContext(ctx, string(runtime), "inspect", containerID)

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var inspectData []map[string]interface{}
	if err := json.Unmarshal(output, &inspectData); err != nil {
		return nil, err
	}

	if len(inspectData) == 0 {
		return nil, fmt.Errorf("no inspect data for container %s", containerID)
	}

	data := inspectData[0]
	container := &DetectedContainer{
		ID:      containerID,
		Runtime: runtime,
	}

	// Extract detailed information from inspect output
	if config, ok := data["Config"].(map[string]interface{}); ok {
		if image, ok := config["Image"].(string); ok {
			container.Image = image
		}

		if workingDir, ok := config["WorkingDir"].(string); ok {
			container.WorkingDir = workingDir
		}

		if cmd, ok := config["Cmd"].([]interface{}); ok {
			for _, c := range cmd {
				if cmdStr, ok := c.(string); ok {
					container.Command = append(container.Command, cmdStr)
				}
			}
		}

		if env, ok := config["Env"].([]interface{}); ok {
			for _, e := range env {
				if envStr, ok := e.(string); ok {
					container.Environment = append(container.Environment, envStr)
				}
			}
		}

		if labels, ok := config["Labels"].(map[string]interface{}); ok {
			container.Labels = make(map[string]string)
			for k, v := range labels {
				if vStr, ok := v.(string); ok {
					container.Labels[k] = vStr
				}
			}
		}
	}

	// Extract state information
	if state, ok := data["State"].(map[string]interface{}); ok {
		if status, ok := state["Status"].(string); ok {
			container.State = status
		}

		if health, ok := state["Health"].(map[string]interface{}); ok {
			if healthStatus, ok := health["Status"].(string); ok {
				container.HealthStatus = healthStatus
			}
		}

		if startedAt, ok := state["StartedAt"].(string); ok {
			if t, err := time.Parse(time.RFC3339Nano, startedAt); err == nil {
				container.StartedAt = t
			}
		}
	}

	// Extract host config
	if hostConfig, ok := data["HostConfig"].(map[string]interface{}); ok {
		if restartPolicy, ok := hostConfig["RestartPolicy"].(map[string]interface{}); ok {
			if name, ok := restartPolicy["Name"].(string); ok {
				container.RestartPolicy = name
			}
		}

		// Extract resource limits
		container.ResourceLimits = &DetectedResourceLimits{}
		if memory, ok := hostConfig["Memory"].(float64); ok {
			container.ResourceLimits.Memory = int64(memory)
		}

		if cpuShares, ok := hostConfig["CpuShares"].(float64); ok {
			container.ResourceLimits.CPUShares = int64(cpuShares)
		}

		if cpuQuota, ok := hostConfig["CpuQuota"].(float64); ok {
			container.ResourceLimits.CPUQuota = int64(cpuQuota)
		}

		if cpuPeriod, ok := hostConfig["CpuPeriod"].(float64); ok {
			container.ResourceLimits.CPUPeriod = int64(cpuPeriod)
		}
	}

	// Extract network settings
	if networkSettings, ok := data["NetworkSettings"].(map[string]interface{}); ok {
		if networks, ok := networkSettings["Networks"].(map[string]interface{}); ok {
			for networkName, networkData := range networks {
				netInfo, ok := networkData.(map[string]interface{})
				if !ok {
					continue
				}
				detectedNet := DetectedNetworkInfo{
					NetworkName: networkName,
				}
				if networkID, ok := netInfo["NetworkID"].(string); ok {
					detectedNet.NetworkID = networkID
				}

				if ipAddress, ok := netInfo["IPAddress"].(string); ok {
					detectedNet.IPAddress = ipAddress
				}

				if macAddress, ok := netInfo["MacAddress"].(string); ok {
					detectedNet.MacAddress = macAddress
				}

				if gateway, ok := netInfo["Gateway"].(string); ok {
					detectedNet.Gateway = gateway
				}

				container.Networks = append(container.Networks, detectedNet)
			}
		}

		// Extract port mappings
		if ports, ok := networkSettings["Ports"].(map[string]interface{}); ok {
			for containerPort, hostPorts := range ports {
				hostPortList, ok := hostPorts.([]interface{})
				if !ok {
					continue
				}
				for _, hostPortData := range hostPortList {
					hostPort, ok := hostPortData.(map[string]interface{})
					if !ok {
						continue
					}
					portMapping := DetectedPortMapping{}

					// Parse container port
					portParts := strings.Split(containerPort, "/")
					if len(portParts) == 2 {
						if port, err := fmt.Sscanf(portParts[0], "%d", &portMapping.ContainerPort); err == nil && port == 1 {
							portMapping.Protocol = portParts[1]
						}
					}

					if hostIP, ok := hostPort["HostIp"].(string); ok {
						portMapping.HostIP = hostIP
					}

					if hostPortStr, ok := hostPort["HostPort"].(string); ok {
						if port, err := fmt.Sscanf(hostPortStr, "%d", &portMapping.HostPort); err == nil && port == 1 {
							container.Ports = append(container.Ports, portMapping)
						}
					}
				}
			}
		}
	}

	// Extract mount information
	if mounts, ok := data["Mounts"].([]interface{}); ok {
		for _, mountData := range mounts {
			mount, ok := mountData.(map[string]interface{})
			if !ok {
				continue
			}
			detectedMount := DetectedMount{}
			if mountType, ok := mount["Type"].(string); ok {
				detectedMount.Type = mountType
			}

			if source, ok := mount["Source"].(string); ok {
				detectedMount.Source = source
			}

			if destination, ok := mount["Destination"].(string); ok {
				detectedMount.Destination = destination
			}

			if mode, ok := mount["Mode"].(string); ok {
				detectedMount.Mode = mode
			}

			if rw, ok := mount["RW"].(bool); ok {
				detectedMount.ReadWrite = rw
			}

			container.Mounts = append(container.Mounts, detectedMount)
		}
	}

	// Extract creation time
	if created, ok := data["Created"].(string); ok {
		if t, err := time.Parse(time.RFC3339Nano, created); err == nil {
			container.Created = t
		}
	}

	return container, nil
}

// mergeContainerInfo merges basic and detailed container information.
func (cd *ContainerDetector) mergeContainerInfo(basic DetectedContainer, detailed *DetectedContainer) DetectedContainer {
	if detailed == nil {
		return basic
	}

	// Prefer detailed information when available
	result := basic
	if detailed.Name != "" {
		result.Name = detailed.Name
	}

	if detailed.Image != "" {
		result.Image = detailed.Image
	}

	if detailed.State != "" {
		result.State = detailed.State
	}

	if detailed.WorkingDir != "" {
		result.WorkingDir = detailed.WorkingDir
	}

	if len(detailed.Command) > 0 {
		result.Command = detailed.Command
	}

	if len(detailed.Environment) > 0 {
		result.Environment = detailed.Environment
	}

	if len(detailed.Labels) > 0 {
		result.Labels = detailed.Labels
	}

	if len(detailed.Networks) > 0 {
		result.Networks = detailed.Networks
	}

	if len(detailed.Ports) > 0 {
		result.Ports = detailed.Ports
	}

	if len(detailed.Mounts) > 0 {
		result.Mounts = detailed.Mounts
	}

	if detailed.ResourceLimits != nil {
		result.ResourceLimits = detailed.ResourceLimits
	}

	if detailed.HealthStatus != "" {
		result.HealthStatus = detailed.HealthStatus
	}

	if detailed.RestartPolicy != "" {
		result.RestartPolicy = detailed.RestartPolicy
	}

	if !detailed.Created.IsZero() {
		result.Created = detailed.Created
	}

	if !detailed.StartedAt.IsZero() {
		result.StartedAt = detailed.StartedAt
	}

	return result
}

// detectNetworks detects container networks.
func (cd *ContainerDetector) detectNetworks(ctx context.Context, runtime ContainerRuntime) ([]DetectedNetwork, error) {
	cmd := exec.CommandContext(ctx, string(runtime), "network", "ls", "--format", "json")

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	networks := make([]DetectedNetwork, 0, len(lines))

	for _, line := range lines {
		if line == "" {
			continue
		}

		var networkData map[string]interface{}
		if err := json.Unmarshal([]byte(line), &networkData); err != nil {
			cd.logger.Warn("Failed to parse network JSON", zap.Error(err))
			continue
		}

		network := DetectedNetwork{
			Runtime: runtime,
		}

		if id, ok := networkData["ID"].(string); ok {
			network.ID = id
		}

		if name, ok := networkData["Name"].(string); ok {
			network.Name = name
		}

		if driver, ok := networkData["Driver"].(string); ok {
			network.Driver = driver
		}

		if scope, ok := networkData["Scope"].(string); ok {
			network.Scope = scope
		}

		// Get detailed network information
		detailed, err := cd.inspectNetwork(ctx, runtime, network.ID)
		if err != nil {
			cd.logger.Warn("Failed to inspect network",
				zap.String("id", network.ID),
				zap.Error(err))
		} else {
			network = cd.mergeNetworkInfo(network, detailed)
		}

		networks = append(networks, network)
	}

	return networks, nil
}

// inspectNetwork gets detailed network information.
func (cd *ContainerDetector) inspectNetwork(ctx context.Context, runtime ContainerRuntime, networkID string) (*DetectedNetwork, error) { //nolint:gocognit // Complex network inspection with multiple runtime support
	cmd := exec.CommandContext(ctx, string(runtime), "network", "inspect", networkID)

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var inspectData []map[string]interface{}
	if err := json.Unmarshal(output, &inspectData); err != nil {
		return nil, err
	}

	if len(inspectData) == 0 {
		return nil, fmt.Errorf("no inspect data for network %s", networkID)
	}

	data := inspectData[0]
	network := &DetectedNetwork{
		ID:      networkID,
		Runtime: runtime,
	}

	if name, ok := data["Name"].(string); ok {
		network.Name = name
	}

	if driver, ok := data["Driver"].(string); ok {
		network.Driver = driver
	}

	if scope, ok := data["Scope"].(string); ok {
		network.Scope = scope
	}

	if internal, ok := data["Internal"].(bool); ok {
		network.Internal = internal
	}

	if created, ok := data["Created"].(string); ok {
		if t, err := time.Parse(time.RFC3339Nano, created); err == nil {
			network.Created = t
		}
	}

	// Extract IPAM information
	if ipam, ok := data["IPAM"].(map[string]interface{}); ok {
		network.IPAM = NetworkIPAM{}
		if driver, ok := ipam["Driver"].(string); ok {
			network.IPAM.Driver = driver
		}

		if config, ok := ipam["Config"].([]interface{}); ok {
			for _, configData := range config {
				configMap, ok := configData.(map[string]interface{})
				if !ok {
					continue
				}
				ipamConfig := ContainerIPAMConfig{}
				if subnet, ok := configMap["Subnet"].(string); ok {
					ipamConfig.Subnet = subnet
				}

				if gateway, ok := configMap["Gateway"].(string); ok {
					ipamConfig.Gateway = gateway
				}

				if ipRange, ok := configMap["IPRange"].(string); ok {
					ipamConfig.IPRange = ipRange
				}

				network.IPAM.Config = append(network.IPAM.Config, ipamConfig)
			}
		}

		if options, ok := ipam["Options"].(map[string]interface{}); ok {
			network.IPAM.Options = make(map[string]string)
			for k, v := range options {
				if vStr, ok := v.(string); ok {
					network.IPAM.Options[k] = vStr
				}
			}
		}
	}

	// Extract options
	if options, ok := data["Options"].(map[string]interface{}); ok {
		network.Options = make(map[string]string)
		for k, v := range options {
			if vStr, ok := v.(string); ok {
				network.Options[k] = vStr
			}
		}
	}

	// Extract labels
	if labels, ok := data["Labels"].(map[string]interface{}); ok {
		network.Labels = make(map[string]string)
		for k, v := range labels {
			if vStr, ok := v.(string); ok {
				network.Labels[k] = vStr
			}
		}
	}

	return network, nil
}

// mergeNetworkInfo merges basic and detailed network information.
func (cd *ContainerDetector) mergeNetworkInfo(basic DetectedNetwork, detailed *DetectedNetwork) DetectedNetwork {
	if detailed == nil {
		return basic
	}

	result := basic
	if detailed.Name != "" {
		result.Name = detailed.Name
	}

	if detailed.Driver != "" {
		result.Driver = detailed.Driver
	}

	if detailed.Scope != "" {
		result.Scope = detailed.Scope
	}

	result.Internal = detailed.Internal
	if !detailed.Created.IsZero() {
		result.Created = detailed.Created
	}

	if len(detailed.IPAM.Config) > 0 {
		result.IPAM = detailed.IPAM
	}

	if len(detailed.Options) > 0 {
		result.Options = detailed.Options
	}

	if len(detailed.Labels) > 0 {
		result.Labels = detailed.Labels
	}

	return result
}

// detectComposeProjects detects Docker Compose projects.
func (cd *ContainerDetector) detectComposeProjects(ctx context.Context, runtime ContainerRuntime) ([]ComposeProject, error) { //nolint:gocognit // Complex compose project detection logic
	// Only supported for Docker and Podman with compose
	if runtime != Docker && runtime != Podman {
		return []ComposeProject{}, nil
	}

	containers, err := cd.detectRunningContainers(ctx, runtime)
	if err != nil {
		return nil, err
	}

	// Group containers by compose project
	projects := make(map[string]*ComposeProject)

	for _, container := range containers {
		// Check for compose labels
		var (
			projectName string
			serviceName string
		)

		if container.Labels != nil {
			// Docker Compose v2 labels
			if project, ok := container.Labels["com.docker.compose.project"]; ok {
				projectName = project
			}

			if service, ok := container.Labels["com.docker.compose.service"]; ok {
				serviceName = service
			}

			// Docker Compose v1 labels (fallback)
			if projectName == "" {
				if project, ok := container.Labels["com.docker.compose.project.name"]; ok {
					projectName = project
				}
			}
		}

		if projectName == "" {
			continue // Not a compose container
		}

		// Initialize project if not exists
		if _, exists := projects[projectName]; !exists {
			projects[projectName] = &ComposeProject{
				Name:        projectName,
				Services:    []ComposeService{},
				Networks:    []string{},
				Volumes:     []string{},
				Environment: make(map[string]string),
				Runtime:     runtime,
			}
		}

		project := projects[projectName]

		// Find or create service
		var service *ComposeService

		for i := range project.Services {
			if project.Services[i].Name == serviceName {
				service = &project.Services[i]
				break
			}
		}

		if service == nil {
			project.Services = append(project.Services, ComposeService{
				Name:       serviceName,
				Image:      container.Image,
				Containers: []string{},
				Replicas:   0,
				Ports:      []DetectedPortMapping{},
				Labels:     make(map[string]string),
			})
			service = &project.Services[len(project.Services)-1]
		}

		// Add container to service
		service.Containers = append(service.Containers, container.ID)
		service.Replicas++

		// Merge ports
		for _, port := range container.Ports {
			found := false

			for _, existingPort := range service.Ports {
				if existingPort.ContainerPort == port.ContainerPort &&
					existingPort.Protocol == port.Protocol {
					found = true
					break
				}
			}

			if !found {
				service.Ports = append(service.Ports, port)
			}
		}

		// Merge labels
		for k, v := range container.Labels {
			if strings.HasPrefix(k, "com.docker.compose.") {
				continue // Skip compose-specific labels
			}

			service.Labels[k] = v
		}

		// Collect networks
		for _, network := range container.Networks {
			found := false

			for _, existingNetwork := range project.Networks {
				if existingNetwork == network.NetworkName {
					found = true
					break
				}
			}

			if !found {
				project.Networks = append(project.Networks, network.NetworkName)
			}
		}

		// Collect volumes
		for _, mount := range container.Mounts {
			if mount.Type == "volume" {
				found := false

				for _, existingVolume := range project.Volumes {
					if existingVolume == mount.Source {
						found = true
						break
					}
				}

				if !found {
					project.Volumes = append(project.Volumes, mount.Source)
				}
			}
		}
	}

	// Convert map to slice
	result := make([]ComposeProject, 0, len(projects))
	for _, project := range projects {
		result = append(result, *project)
	}

	return result, nil
}

// detectKubernetesInfo detects Kubernetes cluster information.
func (cd *ContainerDetector) detectKubernetesInfo(ctx context.Context) (*KubernetesClusterInfo, error) {
	// Check if kubectl is available
	if _, err := exec.LookPath("kubectl"); err != nil {
		return nil, fmt.Errorf("kubectl not found")
	}

	info := &KubernetesClusterInfo{}

	// Get cluster info
	cmd := exec.CommandContext(ctx, "kubectl", "cluster-info", "--request-timeout=5s")
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("cluster not accessible")
	}

	info.Available = true

	// Get version
	cmd = exec.CommandContext(ctx, "kubectl", "version", "--client", "--output=json")

	output, err := cmd.Output()
	if err == nil {
		var versionData map[string]interface{}
		if err := json.Unmarshal(output, &versionData); err == nil {
			if clientVersion, ok := versionData["clientVersion"].(map[string]interface{}); ok {
				if gitVersion, ok := clientVersion["gitVersion"].(string); ok {
					info.Version = gitVersion
				}
			}
		}
	}

	// Get current context
	cmd = exec.CommandContext(ctx, "kubectl", "config", "current-context")

	output, err = cmd.Output()
	if err == nil {
		info.Context = strings.TrimSpace(string(output))
	}

	// Get current namespace
	cmd = exec.CommandContext(ctx, "kubectl", "config", "view", "--minify", "--output=jsonpath={..namespace}")

	output, err = cmd.Output()
	if err == nil {
		info.Namespace = strings.TrimSpace(string(output))
	}

	if info.Namespace == "" {
		info.Namespace = kubernetesDefaultNamespace
	}

	// Get nodes
	nodes, err := cd.getKubernetesNodes(ctx)
	if err == nil {
		info.Nodes = nodes
	}

	// Get namespaces
	namespaces, err := cd.getKubernetesNamespaces(ctx)
	if err == nil {
		info.Namespaces = namespaces
	}

	// Detect service mesh
	serviceMesh, err := cd.detectServiceMesh(ctx)
	if err == nil {
		info.ServiceMesh = serviceMesh
	}

	// Detect ingress controllers
	ingressControllers := cd.detectIngressControllers(ctx)
	info.IngressControllers = ingressControllers

	return info, nil
}

// getKubernetesNodes gets Kubernetes node information.
func (cd *ContainerDetector) getKubernetesNodes(ctx context.Context) ([]KubernetesNode, error) { //nolint:gocognit // Complex Kubernetes node detection with multiple API calls
	cmd := exec.CommandContext(ctx, "kubectl", "get", "nodes", "-o", "json")

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var nodeList map[string]interface{}
	if err := json.Unmarshal(output, &nodeList); err != nil {
		return nil, err
	}

	items, ok := nodeList["items"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid node list format")
	}

	nodes := make([]KubernetesNode, 0, len(items))

	for _, item := range items {
		nodeData, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		node := KubernetesNode{
			Addresses: make(map[string]string),
		}

		// Extract metadata
		if metadata, ok := nodeData["metadata"].(map[string]interface{}); ok {
			if name, ok := metadata["name"].(string); ok {
				node.Name = name
			}
		}

		// Extract status
		if status, ok := nodeData["status"].(map[string]interface{}); ok {
			if conditions, ok := status["conditions"].([]interface{}); ok {
				for _, condition := range conditions {
					if condMap, ok := condition.(map[string]interface{}); ok {
						if condType, ok := condMap["type"].(string); ok && condType == "Ready" {
							if condStatus, ok := condMap["status"].(string); ok {
								node.Status = condStatus
							}
						}
					}
				}
			}

			if nodeInfo, ok := status["nodeInfo"].(map[string]interface{}); ok {
				if kubeletVersion, ok := nodeInfo["kubeletVersion"].(string); ok {
					node.Version = kubeletVersion
				}

				if osImage, ok := nodeInfo["osImage"].(string); ok {
					node.OS = osImage
				}

				if arch, ok := nodeInfo["architecture"].(string); ok {
					node.Arch = arch
				}
			}

			if addresses, ok := status["addresses"].([]interface{}); ok {
				for _, address := range addresses {
					if addrMap, ok := address.(map[string]interface{}); ok {
						if addrType, ok := addrMap["type"].(string); ok {
							if addr, ok := addrMap["address"].(string); ok {
								node.Addresses[addrType] = addr
							}
						}
					}
				}
			}
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}

// getKubernetesNamespaces gets Kubernetes namespaces.
func (cd *ContainerDetector) getKubernetesNamespaces(ctx context.Context) ([]string, error) {
	cmd := exec.CommandContext(ctx, "kubectl", "get", "namespaces", "-o", "jsonpath={.items[*].metadata.name}")

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	namespaces := strings.Fields(strings.TrimSpace(string(output)))

	return namespaces, nil
}

// detectServiceMesh detects service mesh installation.
func (cd *ContainerDetector) detectServiceMesh(ctx context.Context) (*ServiceMeshInfo, error) {
	// Check for Istio
	cmd := exec.CommandContext(ctx, "kubectl", "get", "namespace", "istio-system")
	if cmd.Run() == nil {
		// Istio namespace exists, check for istiod
		cmd = exec.CommandContext(ctx, "kubectl", "get", "deployment", "-n", "istio-system", "istiod")
		if cmd.Run() == nil {
			return &ServiceMeshInfo{
				Type:      "istio",
				Namespace: "istio-system",
				Enabled:   true,
			}, nil
		}
	}

	// Check for Linkerd
	cmd = exec.CommandContext(ctx, "kubectl", "get", "namespace", "linkerd")
	if cmd.Run() == nil {
		// Linkerd namespace exists, check for linkerd-controller
		cmd = exec.CommandContext(ctx, "kubectl", "get", "deployment", "-n", "linkerd", "linkerd-controller")
		if cmd.Run() == nil {
			return &ServiceMeshInfo{
				Type:      "linkerd",
				Namespace: "linkerd",
				Enabled:   true,
			}, nil
		}
	}

	// Check for Consul Connect
	cmd = exec.CommandContext(ctx, "kubectl", "get", "pods", "-l", "app=consul")
	if cmd.Run() == nil {
		return &ServiceMeshInfo{
			Type:    "consul-connect",
			Enabled: true,
		}, nil
	}

	return nil, fmt.Errorf("no service mesh detected")
}

// detectIngressControllers detects ingress controllers.
func (cd *ContainerDetector) detectIngressControllers(ctx context.Context) []string {
	var controllers []string

	// Check for common ingress controllers
	candidates := map[string]string{
		"nginx":   "app.kubernetes.io/name=ingress-nginx",
		"traefik": "app.kubernetes.io/name=traefik",
		"istio":   "app=istio-ingressgateway",
		"haproxy": "app.kubernetes.io/name=haproxy-ingress",
	}

	for name, selector := range candidates {
		cmd := exec.CommandContext(ctx, "kubectl", "get", "pods", "-l", selector, "--all-namespaces")
		if cmd.Run() == nil {
			controllers = append(controllers, name)
		}
	}

	return controllers
}

// calculateResourceUsage calculates overall resource usage.
func (cd *ContainerDetector) calculateResourceUsage(ctx context.Context, env *ContainerEnvironment) *ContainerResourceUsage {
	usage := &ContainerResourceUsage{
		TotalContainers: len(env.RunningContainers),
		Networks:        len(env.Networks),
		ResourceSummary: ResourceSummary{},
	}

	// Count running/stopped containers
	for _, container := range env.RunningContainers {
		if container.State == "running" {
			usage.RunningContainers++
		} else {
			usage.StoppedContainers++
		}
	}

	// Calculate resource usage for primary runtime
	if env.PrimaryRuntime != "" {
		stats, err := cd.getContainerStats(ctx, env.PrimaryRuntime)
		if err != nil {
			cd.logger.Debug("Failed to get container stats", zap.Error(err))
		} else {
			usage.ResourceSummary = stats
		}

		// Get image count
		imageCount, err := cd.getImageCount(ctx, env.PrimaryRuntime)
		if err != nil {
			cd.logger.Debug("Failed to get image count", zap.Error(err))
		} else {
			usage.Images = imageCount
		}

		// Get volume count
		volumeCount, err := cd.getVolumeCount(ctx, env.PrimaryRuntime)
		if err != nil {
			cd.logger.Debug("Failed to get volume count", zap.Error(err))
		} else {
			usage.Volumes = volumeCount
		}
	}

	return usage
}

// getContainerStats gets aggregated container statistics.
func (cd *ContainerDetector) getContainerStats(ctx context.Context, runtime ContainerRuntime) (ResourceSummary, error) {
	var summary ResourceSummary

	cmd := exec.CommandContext(ctx, string(runtime), "stats", "--no-stream", "--format", "json")

	output, err := cmd.Output()
	if err != nil {
		return summary, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		var stats map[string]interface{}
		if err := json.Unmarshal([]byte(line), &stats); err != nil {
			continue
		}

		// Aggregate CPU usage
		if cpuPercent, ok := stats["CPUPerc"].(string); ok {
			cpuPercent = strings.TrimSuffix(cpuPercent, "%")

			var cpu float64
			if n, err := fmt.Sscanf(cpuPercent, "%f", &cpu); err == nil && n == 1 {
				summary.CPUUsage += cpu
			}
		}

		// Aggregate memory usage
		if memUsage, ok := stats["MemUsage"].(string); ok {
			// Parse format like "1.5GiB / 2GiB"
			parts := strings.Split(memUsage, " / ")
			if len(parts) == 2 {
				usage := cd.parseMemorySize(parts[0])
				limit := cd.parseMemorySize(parts[1])
				summary.MemoryUsage += usage
				summary.MemoryLimit += limit
			}
		}

		// Aggregate network I/O
		if netIO, ok := stats["NetIO"].(string); ok {
			// Parse format like "1.2kB / 3.4kB"
			parts := strings.Split(netIO, " / ")
			if len(parts) == 2 {
				rx := cd.parseNetworkSize(parts[0])
				tx := cd.parseNetworkSize(parts[1])
				summary.NetworkRx += rx
				summary.NetworkTx += tx
			}
		}

		// Aggregate block I/O
		if blockIO, ok := stats["BlockIO"].(string); ok {
			// Parse format like "1.2MB / 3.4MB"
			parts := strings.Split(blockIO, " / ")
			if len(parts) == 2 {
				read := cd.parseBlockSize(parts[0])
				write := cd.parseBlockSize(parts[1])
				summary.BlockRead += read
				summary.BlockWrite += write
			}
		}
	}

	return summary, nil
}

// parseMemorySize parses memory size strings like "1.5GiB".
func (cd *ContainerDetector) parseMemorySize(sizeStr string) int64 {
	sizeStr = strings.TrimSpace(sizeStr)
	if sizeStr == "" {
		return 0
	}

	var (
		size float64
		unit string
	)

	if n, err := fmt.Sscanf(sizeStr, "%f%s", &size, &unit); err != nil || n != 2 {
		return 0
	}

	multiplier := int64(1)

	switch strings.ToLower(unit) {
	case "kib", "k":
		multiplier = 1024
	case "mib", "m":
		multiplier = 1024 * 1024
	case "gib", "g":
		multiplier = 1024 * 1024 * 1024
	case "tib", "t":
		multiplier = 1024 * 1024 * 1024 * 1024
	case "kb":
		multiplier = 1000
	case "mb":
		multiplier = 1000 * 1000
	case "gb":
		multiplier = 1000 * 1000 * 1000
	case "tb":
		multiplier = 1000 * 1000 * 1000 * 1000
	}

	return int64(size * float64(multiplier))
}

// parseNetworkSize parses network size strings like "1.2kB".
func (cd *ContainerDetector) parseNetworkSize(sizeStr string) int64 {
	return cd.parseMemorySize(sizeStr)
}

// parseBlockSize parses block size strings like "1.2MB".
func (cd *ContainerDetector) parseBlockSize(sizeStr string) int64 {
	return cd.parseMemorySize(sizeStr)
}

// getImageCount gets the number of images.
func (cd *ContainerDetector) getImageCount(ctx context.Context, runtime ContainerRuntime) (int, error) {
	cmd := exec.CommandContext(ctx, string(runtime), "images", "-q")

	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return 0, nil
	}

	return len(lines), nil
}

// getVolumeCount gets the number of volumes.
func (cd *ContainerDetector) getVolumeCount(ctx context.Context, runtime ContainerRuntime) (int, error) {
	cmd := exec.CommandContext(ctx, string(runtime), "volume", "ls", "-q")

	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return 0, nil
	}

	return len(lines), nil
}

// generateEnvironmentFingerprint generates a unique fingerprint for the environment.
func (cd *ContainerDetector) generateEnvironmentFingerprint(env *ContainerEnvironment) string {
	var elements []string

	// Include runtime information and counts
	elements = append(elements,
		string(env.PrimaryRuntime),
		env.OrchestrationPlatform,
		fmt.Sprintf("containers:%d", len(env.RunningContainers)),
		fmt.Sprintf("networks:%d", len(env.Networks)),
		fmt.Sprintf("compose:%d", len(env.ComposeProjects)))

	// Include Kubernetes info if available
	if env.KubernetesInfo != nil && env.KubernetesInfo.Available {
		elements = append(elements, fmt.Sprintf("k8s:%s", env.KubernetesInfo.Context))
		if env.KubernetesInfo.ServiceMesh != nil {
			elements = append(elements, fmt.Sprintf("mesh:%s", env.KubernetesInfo.ServiceMesh.Type))
		}
	}

	// Create fingerprint from sorted elements
	fingerprint := strings.Join(elements, "|")

	return fmt.Sprintf("%x", fingerprint)[:16] // Return first 16 characters of hash
}

// InvalidateCache invalidates the cached environment detection.
func (cd *ContainerDetector) InvalidateCache() {
	cd.cacheMutex.Lock()
	defer cd.cacheMutex.Unlock()

	cd.cachedEnvironment = nil
	cd.lastDetection = time.Time{}
}

// GetCachedEnvironment returns the cached environment if available.
func (cd *ContainerDetector) GetCachedEnvironment() *ContainerEnvironment {
	cd.cacheMutex.RLock()
	defer cd.cacheMutex.RUnlock()

	if cd.cachedEnvironment != nil && time.Since(cd.lastDetection) < cd.cacheExpiry {
		return cd.cachedEnvironment
	}

	return nil
}

// Public API methods for container detection

// DetectAvailableRuntimes detects available container runtimes.
func (cd *ContainerDetector) DetectAvailableRuntimes(ctx context.Context) ([]RuntimeInfo, error) {
	return cd.detectContainerRuntimes(ctx)
}

// DeterminePrimaryRuntime determines the primary container runtime.
func (cd *ContainerDetector) DeterminePrimaryRuntime(runtimes []RuntimeInfo) ContainerRuntime {
	return cd.determinePrimaryRuntime(runtimes)
}

// DetectOrchestrationPlatform detects the orchestration platform.
func (cd *ContainerDetector) DetectOrchestrationPlatform(ctx context.Context) (string, error) {
	return cd.detectOrchestrationPlatform(ctx), nil
}

// GetRunningContainers gets running containers for a runtime.
func (cd *ContainerDetector) GetRunningContainers(ctx context.Context, runtime ContainerRuntime) ([]DetectedContainer, error) {
	return cd.detectRunningContainers(ctx, runtime)
}

// GetContainerNetworks gets container networks for a runtime.
func (cd *ContainerDetector) GetContainerNetworks(ctx context.Context, runtime ContainerRuntime) ([]DetectedNetwork, error) {
	return cd.detectNetworks(ctx, runtime)
}

// ParseDockerPsOutput parses docker ps output.
func (cd *ContainerDetector) ParseDockerPsOutput(output string) ([]DetectedContainer, error) {
	var containers []DetectedContainer

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) <= 1 {
		return containers, nil
	}

	// Skip header line
	for i := 1; i < len(lines); i++ {
		fields := strings.Split(lines[i], ",")
		if len(fields) >= 7 {
			container := DetectedContainer{
				ID:      fields[0],
				Image:   fields[1],
				Status:  fields[4],
				Name:    fields[6],
				Runtime: Docker,
			}
			containers = append(containers, container)
		}
	}

	return containers, nil
}

// ParseDockerNetworkOutput parses docker network ls output.
func (cd *ContainerDetector) ParseDockerNetworkOutput(output string) ([]DetectedNetwork, error) {
	var networks []DetectedNetwork

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) <= 1 {
		return networks, nil
	}

	// Skip header line
	for i := 1; i < len(lines); i++ {
		fields := strings.Split(lines[i], ",")
		if len(fields) >= 4 {
			network := DetectedNetwork{
				ID:      fields[0],
				Name:    fields[1],
				Driver:  fields[2],
				Scope:   fields[3],
				Runtime: Docker,
			}
			networks = append(networks, network)
		}
	}

	return networks, nil
}

// CalculateEnvironmentFingerprint calculates environment fingerprint.
func (cd *ContainerDetector) CalculateEnvironmentFingerprint(env *ContainerEnvironment) string {
	return cd.generateEnvironmentFingerprint(env)
}
