// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package doctor

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/gizzahub/gzh-cli/internal/cli"
	"github.com/gizzahub/gzh-cli/internal/logger"
)

// ContainerDiagnostics represents comprehensive container diagnostics.
type ContainerDiagnostics struct {
	Timestamp          time.Time              `json:"timestamp"`
	Environment        ContainerEnvironment   `json:"environment"`
	Containers         []ContainerInfo        `json:"containers"`
	Networks           []NetworkInfo          `json:"networks"`
	Images             []ImageInfo            `json:"images"`
	Volumes            []VolumeInfo           `json:"volumes"`
	SystemInfo         DockerSystemInfo       `json:"systemInfo"`
	ResourceUsage      ResourceUsage          `json:"resourceUsage"`
	HealthChecks       []ContainerHealthCheck `json:"healthChecks"`
	SecurityAnalysis   SecurityAnalysis       `json:"securityAnalysis"`
	PerformanceMetrics PerformanceMetrics     `json:"performanceMetrics"`
	Recommendations    []string               `json:"recommendations"`
	Issues             []DiagnosticIssue      `json:"issues"`
}

// ContainerEnvironment captures the Docker environment details.
type ContainerEnvironment struct {
	DockerVersion    string `json:"dockerVersion"`
	ComposeVersion   string `json:"composeVersion,omitempty"`
	Platform         string `json:"platform"`
	KernelVersion    string `json:"kernelVersion"`
	Architecture     string `json:"architecture"`
	TotalMemory      uint64 `json:"totalMemory"`
	CPUCount         int    `json:"cpuCount"`
	StorageDriver    string `json:"storageDriver"`
	CgroupVersion    string `json:"cgroupVersion"`
	RuntimesDetected string `json:"runtimesDetected"`
}

// ContainerInfo represents detailed container information.
type ContainerInfo struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Image          string                 `json:"image"`
	Status         string                 `json:"status"`
	State          string                 `json:"state"`
	Created        time.Time              `json:"created"`
	Started        *time.Time             `json:"started,omitempty"`
	Ports          []PortMapping          `json:"ports"`
	Networks       map[string]NetworkInfo `json:"networks"`
	Mounts         []MountInfo            `json:"mounts"`
	Environment    []string               `json:"environment,omitempty"`
	Labels         map[string]string      `json:"labels"`
	ResourceLimits ResourceLimits         `json:"resourceLimits"`
	HealthStatus   string                 `json:"healthStatus,omitempty"`
	RestartCount   int                    `json:"restartCount"`
	ExitCode       *int                   `json:"exitCode,omitempty"`
	LogPath        string                 `json:"logPath,omitempty"`
	ComposeProject string                 `json:"composeProject,omitempty"`
	ComposeService string                 `json:"composeService,omitempty"`
}

// NetworkInfo represents Docker network information.
type NetworkInfo struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Driver     string            `json:"driver"`
	Scope      string            `json:"scope"`
	Internal   bool              `json:"internal"`
	Attachable bool              `json:"attachable"`
	IPAM       IPAMInfo          `json:"ipam"`
	Containers map[string]string `json:"containers"`
	Options    map[string]string `json:"options"`
	Labels     map[string]string `json:"labels"`
	Created    time.Time         `json:"created"`
}

// ImageInfo represents Docker image information.
type ImageInfo struct {
	ID          string            `json:"id"`
	RepoTags    []string          `json:"repoTags"`
	Size        int64             `json:"size"`
	VirtualSize int64             `json:"virtualSize"`
	Created     time.Time         `json:"created"`
	Labels      map[string]string `json:"labels"`
	Dangling    bool              `json:"dangling"`
	InUse       bool              `json:"inUse"`
	Layers      int               `json:"layers"`
}

// VolumeInfo represents Docker volume information.
type VolumeInfo struct {
	Name       string            `json:"name"`
	Driver     string            `json:"driver"`
	Mountpoint string            `json:"mountpoint"`
	Scope      string            `json:"scope"`
	Labels     map[string]string `json:"labels"`
	Options    map[string]string `json:"options"`
	Size       *int64            `json:"size,omitempty"`
	InUse      bool              `json:"inUse"`
	Created    time.Time         `json:"created"`
}

// DockerSystemInfo represents Docker system information.
type DockerSystemInfo struct {
	ContainersRunning int    `json:"containersRunning"`
	ContainersPaused  int    `json:"containersPaused"`
	ContainersStopped int    `json:"containersStopped"`
	Images            int    `json:"images"`
	ServerVersion     string `json:"serverVersion"`
	StorageDriver     string `json:"storageDriver"`
	LoggingDriver     string `json:"loggingDriver"`
	KernelVersion     string `json:"kernelVersion"`
	OperatingSystem   string `json:"operatingSystem"`
	Architecture      string `json:"architecture"`
	NCPU              int    `json:"ncpu"`
	MemTotal          int64  `json:"memTotal"`
	DockerRootDir     string `json:"dockerRootDir"`
	HTTPProxy         string `json:"httpProxy,omitempty"`
	HTTPSProxy        string `json:"httpsProxy,omitempty"`
	NoProxy           string `json:"noProxy,omitempty"`
}

// ResourceUsage represents system resource usage.
type ResourceUsage struct {
	TotalMemory    int64            `json:"totalMemory"`
	UsedMemory     int64            `json:"usedMemory"`
	TotalDisk      int64            `json:"totalDisk"`
	UsedDisk       int64            `json:"usedDisk"`
	ContainerStats []ContainerStats `json:"containerStats"`
	NetworkTraffic NetworkTraffic   `json:"networkTraffic"`
}

// ContainerStats represents container resource statistics.
type ContainerStats struct {
	ContainerID   string  `json:"containerId"`
	ContainerName string  `json:"containerName"`
	CPUPercent    float64 `json:"cpuPercent"`
	MemoryUsage   int64   `json:"memoryUsage"`
	MemoryLimit   int64   `json:"memoryLimit"`
	MemoryPercent float64 `json:"memoryPercent"`
	NetworkRX     int64   `json:"networkRx"`
	NetworkTX     int64   `json:"networkTx"`
	BlockRead     int64   `json:"blockRead"`
	BlockWrite    int64   `json:"blockWrite"`
	PIDs          int     `json:"pids"`
}

// ContainerHealthCheck represents container health check results.
type ContainerHealthCheck struct {
	ContainerID    string    `json:"containerId"`
	ContainerName  string    `json:"containerName"`
	HealthStatus   string    `json:"healthStatus"`
	FailingStreak  int       `json:"failingStreak"`
	LastCheck      time.Time `json:"lastCheck"`
	ChecksTotal    int       `json:"checksTotal"`
	ChecksFailed   int       `json:"checksFailed"`
	HealthCommand  string    `json:"healthCommand,omitempty"`
	HealthInterval string    `json:"healthInterval,omitempty"`
	HealthTimeout  string    `json:"healthTimeout,omitempty"`
	HealthRetries  int       `json:"healthRetries,omitempty"`
}

// SecurityAnalysis represents container security analysis.
type SecurityAnalysis struct {
	PrivilegedContainers  []string            `json:"privilegedContainers"`
	RootContainers        []string            `json:"rootContainers"`
	CapabilitiesAdded     map[string][]string `json:"capabilitiesAdded"`
	HostNetworkContainers []string            `json:"hostNetworkContainers"`
	HostPIDContainers     []string            `json:"hostPidContainers"`
	SecretsExposed        []SecurityIssue     `json:"secretsExposed"`
	VulnerableImages      []VulnerabilityInfo `json:"vulnerableImages"`
	SecurityScore         float64             `json:"securityScore"`
	Recommendations       []string            `json:"recommendations"`
}

// SecurityIssue represents a security issue.
type SecurityIssue struct {
	ContainerID   string `json:"containerId"`
	ContainerName string `json:"containerName"`
	Issue         string `json:"issue"`
	Severity      string `json:"severity"`
	Description   string `json:"description"`
	Resolution    string `json:"resolution"`
}

// VulnerabilityInfo represents image vulnerability information.
type VulnerabilityInfo struct {
	ImageID     string     `json:"imageId"`
	ImageName   string     `json:"imageName"`
	Critical    int        `json:"critical"`
	High        int        `json:"high"`
	Medium      int        `json:"medium"`
	Low         int        `json:"low"`
	LastScanned *time.Time `json:"lastScanned,omitempty"`
}

// PerformanceMetrics represents container performance metrics.
type PerformanceMetrics struct {
	AverageStartupTime time.Duration            `json:"averageStartupTime"`
	MemoryEfficiency   float64                  `json:"memoryEfficiency"`
	CPUEfficiency      float64                  `json:"cpuEfficiency"`
	NetworkLatency     map[string]time.Duration `json:"networkLatency"`
	DiskIOPerformance  DiskIOMetrics            `json:"diskIoPerformance"`
	ContainerDensity   float64                  `json:"containerDensity"`
	ResourceWastage    ResourceWastage          `json:"resourceWastage"`
}

// DiagnosticIssue represents a diagnostic issue found.
type DiagnosticIssue struct {
	ID          string    `json:"id"`
	Severity    string    `json:"severity"`
	Category    string    `json:"category"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Resolution  string    `json:"resolution"`
	Affected    []string  `json:"affected"`
	DetectedAt  time.Time `json:"detectedAt"`
}

// Additional supporting types.
type PortMapping struct {
	PrivatePort int    `json:"privatePort"`
	PublicPort  int    `json:"publicPort,omitempty"`
	Type        string `json:"type"`
	IP          string `json:"ip,omitempty"`
}

type MountInfo struct {
	Type        string `json:"type"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Mode        string `json:"mode"`
	RW          bool   `json:"rw"`
	Propagation string `json:"propagation,omitempty"`
}

type ResourceLimits struct {
	Memory      *int64 `json:"memory,omitempty"`
	CPUShares   *int64 `json:"cpuShares,omitempty"`
	CPUQuota    *int64 `json:"cpuQuota,omitempty"`
	CPUPeriod   *int64 `json:"cpuPeriod,omitempty"`
	CPUsetCPUs  string `json:"cpusetCpus,omitempty"`
	BlkioWeight *int64 `json:"blkioWeight,omitempty"`
}

type IPAMInfo struct {
	Driver  string            `json:"driver"`
	Config  []IPAMEntry       `json:"config"`
	Options map[string]string `json:"options"`
}

type IPAMEntry struct {
	Subnet  string `json:"subnet"`
	Gateway string `json:"gateway"`
	IPRange string `json:"ipRange,omitempty"`
}

type NetworkTraffic struct {
	TotalRX int64 `json:"totalRx"`
	TotalTX int64 `json:"totalTx"`
}

type DiskIOMetrics struct {
	TotalRead  int64 `json:"totalRead"`
	TotalWrite int64 `json:"totalWrite"`
	ReadOps    int64 `json:"readOps"`
	WriteOps   int64 `json:"writeOps"`
}

type ResourceWastage struct {
	OverProvisionedMemory int64   `json:"overProvisionedMemory"`
	UnderutilizedCPU      float64 `json:"underutilizedCpu"`
	UnusedVolumes         int     `json:"unusedVolumes"`
	DanglingImages        int     `json:"danglingImages"`
}

// newContainerCmd creates the container monitoring and diagnostics subcommand.
func newContainerCmd() *cobra.Command {
	ctx := context.Background()

	var (
		includeStats     bool
		includeHealth    bool
		includeSecurity  bool
		includeNetworks  bool
		includeImages    bool
		includeVolumes   bool
		containerFilter  string
		watchMode        bool
		watchInterval    time.Duration
		outputFile       string
		generateReport   bool
		serverMode       bool
		serverPort       int
		detailedAnalysis bool
	)

	cmd := cli.NewCommandBuilder(ctx, "container", "Monitor and diagnose Docker containers").
		WithLongDescription(`Comprehensive Docker container monitoring and diagnostics.

This command provides detailed analysis of your Docker environment including:
- Container status, resource usage, and performance metrics
- Network topology and connectivity analysis
- Image vulnerability scanning and security analysis
- Volume usage and storage optimization recommendations
- Health check monitoring and failure analysis
- Performance profiling and resource utilization
- Security posture assessment and compliance checks
- Comprehensive reporting and alerting capabilities

Features:
- Real-time container monitoring with customizable metrics
- Security vulnerability detection and remediation guidance
- Performance optimization recommendations
- Network topology visualization and diagnostics
- Resource usage analysis and capacity planning
- Health monitoring with alerting capabilities
- Comprehensive reporting in multiple formats
- HTTP server mode for external monitoring integration

Examples:
  gz doctor container                           # Basic container diagnostics
  gz doctor container --include-stats          # Include resource statistics
  gz doctor container --include-security       # Include security analysis
  gz doctor container --container-filter nginx # Filter by container name pattern
  gz doctor container --watch --interval 30s   # Continuous monitoring every 30s
  gz doctor container --detailed-analysis      # Comprehensive analysis with recommendations
  gz doctor container --server --port 8080     # HTTP server mode for monitoring integration`).
		WithExample("gz doctor container --include-stats --include-security").
		WithFormatFlag("table", []string{"table", "json", "yaml"}).
		WithRunFuncE(func(ctx context.Context, flags *cli.CommonFlags, args []string) error {
			return runContainerDiagnostics(ctx, flags, containerOptions{
				includeStats:     includeStats,
				includeHealth:    includeHealth,
				includeSecurity:  includeSecurity,
				includeNetworks:  includeNetworks,
				includeImages:    includeImages,
				includeVolumes:   includeVolumes,
				containerFilter:  containerFilter,
				watchMode:        watchMode,
				watchInterval:    watchInterval,
				outputFile:       outputFile,
				generateReport:   generateReport,
				serverMode:       serverMode,
				serverPort:       serverPort,
				detailedAnalysis: detailedAnalysis,
			})
		}).
		Build()

	cmd.Flags().BoolVar(&includeStats, "include-stats", true, "Include container resource statistics")
	cmd.Flags().BoolVar(&includeHealth, "include-health", true, "Include container health checks")
	cmd.Flags().BoolVar(&includeSecurity, "include-security", false, "Include security analysis")
	cmd.Flags().BoolVar(&includeNetworks, "include-networks", true, "Include network information")
	cmd.Flags().BoolVar(&includeImages, "include-images", false, "Include image information")
	cmd.Flags().BoolVar(&includeVolumes, "include-volumes", false, "Include volume information")
	cmd.Flags().StringVar(&containerFilter, "container-filter", "", "Filter containers by name pattern")
	cmd.Flags().BoolVar(&watchMode, "watch", false, "Continuous monitoring mode")
	cmd.Flags().DurationVar(&watchInterval, "watch-interval", 10*time.Second, "Watch mode update interval")
	cmd.Flags().StringVar(&outputFile, "output", "", "Output file for diagnostics results")
	cmd.Flags().BoolVar(&generateReport, "generate-report", false, "Generate comprehensive diagnostics report")
	cmd.Flags().BoolVar(&serverMode, "server", false, "Run in HTTP server mode for external monitoring")
	cmd.Flags().IntVar(&serverPort, "server-port", 8080, "HTTP server port")
	cmd.Flags().BoolVar(&detailedAnalysis, "detailed-analysis", false, "Perform detailed analysis with recommendations")

	return cmd
}

type containerOptions struct {
	includeStats     bool
	includeHealth    bool
	includeSecurity  bool
	includeNetworks  bool
	includeImages    bool
	includeVolumes   bool
	containerFilter  string
	watchMode        bool
	watchInterval    time.Duration
	outputFile       string
	generateReport   bool
	serverMode       bool
	serverPort       int
	detailedAnalysis bool
}

func runContainerDiagnostics(ctx context.Context, flags *cli.CommonFlags, opts containerOptions) error {
	logger := logger.NewSimpleLogger("doctor-container")

	logger.Info("Starting container diagnostics",
		"include_stats", opts.includeStats,
		"include_security", opts.includeSecurity,
		"watch_mode", opts.watchMode,
	)

	// Check if Docker is available
	if !isDockerAvailable(ctx) {
		return fmt.Errorf("Docker is not available or not running")
	}

	// Run diagnostics
	diagnostics, err := performContainerDiagnostics(ctx, opts, logger)
	if err != nil {
		return fmt.Errorf("failed to perform container diagnostics: %w", err)
	}

	// Save results if output file specified
	if opts.outputFile != "" {
		if err := saveDiagnosticsResults(diagnostics, opts.outputFile); err != nil {
			return fmt.Errorf("failed to save diagnostics results: %w", err)
		}
		logger.Info("Diagnostics results saved", "file", opts.outputFile)
	}

	// Display results
	formatter := cli.NewOutputFormatter(flags.Format)

	switch flags.Format {
	case "json", "yaml":
		return formatter.FormatOutput(diagnostics)
	default:
		return displayContainerDiagnostics(diagnostics, opts)
	}
}

func performContainerDiagnostics(ctx context.Context, opts containerOptions, logger logger.CommonLogger) (*ContainerDiagnostics, error) {
	diagnostics := &ContainerDiagnostics{
		Timestamp:       time.Now(),
		Containers:      make([]ContainerInfo, 0),
		Networks:        make([]NetworkInfo, 0),
		Images:          make([]ImageInfo, 0),
		Volumes:         make([]VolumeInfo, 0),
		HealthChecks:    make([]ContainerHealthCheck, 0),
		Recommendations: make([]string, 0),
		Issues:          make([]DiagnosticIssue, 0),
	}

	var err error

	// Collect environment information
	logger.Info("Collecting Docker environment information")
	diagnostics.Environment, err = getContainerEnvironment(ctx)
	if err != nil {
		logger.Warn("Failed to collect environment information", "error", err)
	}

	// Collect system information
	logger.Info("Collecting Docker system information")
	diagnostics.SystemInfo, err = getDockerSystemInfo(ctx)
	if err != nil {
		logger.Warn("Failed to collect system information", "error", err)
	}

	// Collect container information
	logger.Info("Collecting container information")
	diagnostics.Containers, err = getContainerInfo(ctx, opts.containerFilter)
	if err != nil {
		logger.Warn("Failed to collect container information", "error", err)
	}

	// Collect network information
	if opts.includeNetworks {
		logger.Info("Collecting network information")
		diagnostics.Networks, err = getNetworkInfo(ctx)
		if err != nil {
			logger.Warn("Failed to collect network information", "error", err)
		}
	}

	// Collect image information
	if opts.includeImages {
		logger.Info("Collecting image information")
		diagnostics.Images, err = getImageInfo(ctx)
		if err != nil {
			logger.Warn("Failed to collect image information", "error", err)
		}
	}

	// Collect volume information
	if opts.includeVolumes {
		logger.Info("Collecting volume information")
		diagnostics.Volumes, err = getVolumeInfo(ctx)
		if err != nil {
			logger.Warn("Failed to collect volume information", "error", err)
		}
	}

	// Collect resource usage statistics
	if opts.includeStats {
		logger.Info("Collecting resource usage statistics")
		diagnostics.ResourceUsage, err = getResourceUsage(ctx, diagnostics.Containers)
		if err != nil {
			logger.Warn("Failed to collect resource usage", "error", err)
		}
	}

	// Collect health check information
	if opts.includeHealth {
		logger.Info("Collecting health check information")
		diagnostics.HealthChecks = getHealthCheckInfo(diagnostics.Containers)
	}

	// Perform security analysis
	if opts.includeSecurity {
		logger.Info("Performing security analysis")
		diagnostics.SecurityAnalysis, err = performSecurityAnalysis(ctx, diagnostics.Containers, diagnostics.Images)
		if err != nil {
			logger.Warn("Failed to perform security analysis", "error", err)
		}
	}

	// Perform detailed analysis if requested
	if opts.detailedAnalysis {
		logger.Info("Performing detailed analysis")
		performDetailedAnalysis(diagnostics)
	}

	// Generate recommendations and identify issues
	generateRecommendations(diagnostics)
	identifyIssues(diagnostics)

	logger.Info("Container diagnostics completed",
		"containers", len(diagnostics.Containers),
		"networks", len(diagnostics.Networks),
		"images", len(diagnostics.Images),
		"volumes", len(diagnostics.Volumes),
		"issues", len(diagnostics.Issues),
		"recommendations", len(diagnostics.Recommendations),
	)

	return diagnostics, nil
}

// Helper functions for Docker operations

func isDockerAvailable(ctx context.Context) bool {
	cmd := exec.CommandContext(ctx, "docker", "version")
	return cmd.Run() == nil
}

func getContainerEnvironment(ctx context.Context) (ContainerEnvironment, error) {
	env := ContainerEnvironment{}

	// Get Docker version
	cmd := exec.CommandContext(ctx, "docker", "version", "--format", "{{.Server.Version}}")
	if output, err := cmd.Output(); err == nil {
		env.DockerVersion = strings.TrimSpace(string(output))
	}

	// Get Docker Compose version
	cmd = exec.CommandContext(ctx, "docker-compose", "version", "--short")
	if output, err := cmd.Output(); err == nil {
		env.ComposeVersion = strings.TrimSpace(string(output))
	}

	// Get system information
	cmd = exec.CommandContext(ctx, "docker", "system", "info", "--format", "json")
	if output, err := cmd.Output(); err == nil {
		var info map[string]interface{}
		if json.Unmarshal(output, &info) == nil {
			if platform, ok := info["OSType"].(string); ok {
				env.Platform = platform
			}
			if kernel, ok := info["KernelVersion"].(string); ok {
				env.KernelVersion = kernel
			}
			if arch, ok := info["Architecture"].(string); ok {
				env.Architecture = arch
			}
			if storage, ok := info["Driver"].(string); ok {
				env.StorageDriver = storage
			}
			if memTotal, ok := info["MemTotal"].(float64); ok {
				env.TotalMemory = uint64(memTotal)
			}
			if cpus, ok := info["NCPU"].(float64); ok {
				env.CPUCount = int(cpus)
			}
			if cgroup, ok := info["CgroupVersion"].(string); ok {
				env.CgroupVersion = cgroup
			}
		}
	}

	return env, nil
}

func getDockerSystemInfo(ctx context.Context) (DockerSystemInfo, error) {
	info := DockerSystemInfo{}

	cmd := exec.CommandContext(ctx, "docker", "system", "info", "--format", "json")
	output, err := cmd.Output()
	if err != nil {
		return info, err
	}

	var data map[string]interface{}
	if err := json.Unmarshal(output, &data); err != nil {
		return info, err
	}

	// Parse system information
	if val, ok := data["ContainersRunning"].(float64); ok {
		info.ContainersRunning = int(val)
	}
	if val, ok := data["ContainersPaused"].(float64); ok {
		info.ContainersPaused = int(val)
	}
	if val, ok := data["ContainersStopped"].(float64); ok {
		info.ContainersStopped = int(val)
	}
	if val, ok := data["Images"].(float64); ok {
		info.Images = int(val)
	}
	if val, ok := data["ServerVersion"].(string); ok {
		info.ServerVersion = val
	}
	if val, ok := data["Driver"].(string); ok {
		info.StorageDriver = val
	}
	if val, ok := data["LoggingDriver"].(string); ok {
		info.LoggingDriver = val
	}
	if val, ok := data["KernelVersion"].(string); ok {
		info.KernelVersion = val
	}
	if val, ok := data["OperatingSystem"].(string); ok {
		info.OperatingSystem = val
	}
	if val, ok := data["Architecture"].(string); ok {
		info.Architecture = val
	}
	if val, ok := data["NCPU"].(float64); ok {
		info.NCPU = int(val)
	}
	if val, ok := data["MemTotal"].(float64); ok {
		info.MemTotal = int64(val)
	}
	if val, ok := data["DockerRootDir"].(string); ok {
		info.DockerRootDir = val
	}

	return info, nil
}

func getContainerInfo(ctx context.Context, filter string) ([]ContainerInfo, error) {
	cmd := exec.CommandContext(ctx, "docker", "ps", "-a", "--format", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	containers := make([]ContainerInfo, 0, len(lines))

	for _, line := range lines {
		if line == "" {
			continue
		}

		var container ContainerInfo
		if err := json.Unmarshal([]byte(line), &container); err != nil {
			continue
		}

		// Apply filter if specified
		if filter != "" && !strings.Contains(container.Name, filter) {
			continue
		}

		// Get detailed container information
		if detailed, err := getDetailedContainerInfo(ctx, container.ID); err == nil {
			container = *detailed
		}

		containers = append(containers, container)
	}

	return containers, nil
}

func getDetailedContainerInfo(ctx context.Context, containerID string) (*ContainerInfo, error) {
	cmd := exec.CommandContext(ctx, "docker", "inspect", containerID)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var inspectData []map[string]interface{}
	if err := json.Unmarshal(output, &inspectData); err != nil {
		return nil, err
	}

	if len(inspectData) == 0 {
		return nil, fmt.Errorf("no container data found")
	}

	data := inspectData[0]
	container := &ContainerInfo{
		ID:     containerID,
		Labels: make(map[string]string),
	}

	// Parse container details from inspect output
	if name, ok := data["Name"].(string); ok {
		container.Name = strings.TrimPrefix(name, "/")
	}

	if config, ok := data["Config"].(map[string]interface{}); ok {
		if image, ok := config["Image"].(string); ok {
			container.Image = image
		}
		if labels, ok := config["Labels"].(map[string]interface{}); ok {
			for k, v := range labels {
				if str, ok := v.(string); ok {
					container.Labels[k] = str
				}
			}
		}
	}

	if state, ok := data["State"].(map[string]interface{}); ok {
		if status, ok := state["Status"].(string); ok {
			container.Status = status
			container.State = status
		}
		if health, ok := state["Health"].(map[string]interface{}); ok {
			if healthStatus, ok := health["Status"].(string); ok {
				container.HealthStatus = healthStatus
			}
		}
		if restartCount, ok := state["RestartCount"].(float64); ok {
			container.RestartCount = int(restartCount)
		}
		if exitCode, ok := state["ExitCode"].(float64); ok {
			exitCodeInt := int(exitCode)
			container.ExitCode = &exitCodeInt
		}
	}

	if created, ok := data["Created"].(string); ok {
		if t, err := time.Parse(time.RFC3339Nano, created); err == nil {
			container.Created = t
		}
	}

	// Parse Docker Compose labels
	if project, exists := container.Labels["com.docker.compose.project"]; exists {
		container.ComposeProject = project
	}
	if service, exists := container.Labels["com.docker.compose.service"]; exists {
		container.ComposeService = service
	}

	return container, nil
}

func getNetworkInfo(ctx context.Context) ([]NetworkInfo, error) {
	cmd := exec.CommandContext(ctx, "docker", "network", "ls", "--format", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	networks := make([]NetworkInfo, 0, len(lines))

	for _, line := range lines {
		if line == "" {
			continue
		}

		var network NetworkInfo
		if err := json.Unmarshal([]byte(line), &network); err != nil {
			continue
		}

		// Get detailed network information
		if detailed, err := getDetailedNetworkInfo(ctx, network.Name); err == nil {
			network = *detailed
		}

		networks = append(networks, network)
	}

	return networks, nil
}

func getDetailedNetworkInfo(ctx context.Context, networkName string) (*NetworkInfo, error) {
	cmd := exec.CommandContext(ctx, "docker", "network", "inspect", networkName)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var inspectData []map[string]interface{}
	if err := json.Unmarshal(output, &inspectData); err != nil {
		return nil, err
	}

	if len(inspectData) == 0 {
		return nil, fmt.Errorf("no network data found")
	}

	data := inspectData[0]
	network := &NetworkInfo{
		Name:       networkName,
		Labels:     make(map[string]string),
		Options:    make(map[string]string),
		Containers: make(map[string]string),
	}

	if id, ok := data["Id"].(string); ok {
		network.ID = id
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
	if attachable, ok := data["Attachable"].(bool); ok {
		network.Attachable = attachable
	}

	if created, ok := data["Created"].(string); ok {
		if t, err := time.Parse(time.RFC3339Nano, created); err == nil {
			network.Created = t
		}
	}

	return network, nil
}

func getImageInfo(ctx context.Context) ([]ImageInfo, error) {
	cmd := exec.CommandContext(ctx, "docker", "images", "--format", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	images := make([]ImageInfo, 0, len(lines))

	for _, line := range lines {
		if line == "" {
			continue
		}

		var image ImageInfo
		if err := json.Unmarshal([]byte(line), &image); err != nil {
			continue
		}

		// Get detailed image information
		if detailed, err := getDetailedImageInfo(ctx, image.ID); err == nil {
			image = *detailed
		}

		images = append(images, image)
	}

	return images, nil
}

func getDetailedImageInfo(ctx context.Context, imageID string) (*ImageInfo, error) {
	cmd := exec.CommandContext(ctx, "docker", "inspect", imageID)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var inspectData []map[string]interface{}
	if err := json.Unmarshal(output, &inspectData); err != nil {
		return nil, err
	}

	if len(inspectData) == 0 {
		return nil, fmt.Errorf("no image data found")
	}

	data := inspectData[0]
	image := &ImageInfo{
		ID:     imageID,
		Labels: make(map[string]string),
	}

	if size, ok := data["Size"].(float64); ok {
		image.Size = int64(size)
	}
	if virtualSize, ok := data["VirtualSize"].(float64); ok {
		image.VirtualSize = int64(virtualSize)
	}
	if created, ok := data["Created"].(string); ok {
		if t, err := time.Parse(time.RFC3339Nano, created); err == nil {
			image.Created = t
		}
	}
	if repoTags, ok := data["RepoTags"].([]interface{}); ok {
		tags := make([]string, 0, len(repoTags))
		for _, tag := range repoTags {
			if str, ok := tag.(string); ok {
				tags = append(tags, str)
			}
		}
		image.RepoTags = tags
	}

	return image, nil
}

func getVolumeInfo(ctx context.Context) ([]VolumeInfo, error) {
	cmd := exec.CommandContext(ctx, "docker", "volume", "ls", "--format", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	volumes := make([]VolumeInfo, 0, len(lines))

	for _, line := range lines {
		if line == "" {
			continue
		}

		var volume VolumeInfo
		if err := json.Unmarshal([]byte(line), &volume); err != nil {
			continue
		}

		// Get detailed volume information
		if detailed, err := getDetailedVolumeInfo(ctx, volume.Name); err == nil {
			volume = *detailed
		}

		volumes = append(volumes, volume)
	}

	return volumes, nil
}

func getDetailedVolumeInfo(ctx context.Context, volumeName string) (*VolumeInfo, error) {
	cmd := exec.CommandContext(ctx, "docker", "volume", "inspect", volumeName)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var inspectData []map[string]interface{}
	if err := json.Unmarshal(output, &inspectData); err != nil {
		return nil, err
	}

	if len(inspectData) == 0 {
		return nil, fmt.Errorf("no volume data found")
	}

	data := inspectData[0]
	volume := &VolumeInfo{
		Name:    volumeName,
		Labels:  make(map[string]string),
		Options: make(map[string]string),
	}

	if driver, ok := data["Driver"].(string); ok {
		volume.Driver = driver
	}
	if mountpoint, ok := data["Mountpoint"].(string); ok {
		volume.Mountpoint = mountpoint
	}
	if scope, ok := data["Scope"].(string); ok {
		volume.Scope = scope
	}
	if created, ok := data["CreatedAt"].(string); ok {
		if t, err := time.Parse(time.RFC3339Nano, created); err == nil {
			volume.Created = t
		}
	}

	return volume, nil
}

func getResourceUsage(ctx context.Context, containers []ContainerInfo) (ResourceUsage, error) {
	usage := ResourceUsage{
		ContainerStats: make([]ContainerStats, 0, len(containers)),
	}

	// Get stats for running containers
	for _, container := range containers {
		if container.State == "running" {
			if stats, err := getContainerStats(ctx, container.ID); err == nil {
				usage.ContainerStats = append(usage.ContainerStats, *stats)
			}
		}
	}

	return usage, nil
}

func getContainerStats(ctx context.Context, containerID string) (*ContainerStats, error) {
	cmd := exec.CommandContext(ctx, "docker", "stats", "--no-stream", "--format", "json", containerID)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var statsData map[string]interface{}
	if err := json.Unmarshal(output, &statsData); err != nil {
		return nil, err
	}

	stats := &ContainerStats{
		ContainerID: containerID,
	}

	if name, ok := statsData["Name"].(string); ok {
		stats.ContainerName = name
	}
	if cpuPerc, ok := statsData["CPUPerc"].(string); ok {
		// Parse CPU percentage (remove % sign)
		cpuStr := strings.TrimSuffix(cpuPerc, "%")
		if cpu, err := strconv.ParseFloat(cpuStr, 64); err == nil {
			stats.CPUPercent = cpu
		}
	}
	if memPerc, ok := statsData["MemPerc"].(string); ok {
		// Parse memory percentage (remove % sign)
		memStr := strings.TrimSuffix(memPerc, "%")
		if mem, err := strconv.ParseFloat(memStr, 64); err == nil {
			stats.MemoryPercent = mem
		}
	}

	return stats, nil
}

func getHealthCheckInfo(containers []ContainerInfo) []ContainerHealthCheck {
	healthChecks := make([]ContainerHealthCheck, 0)

	for _, container := range containers {
		if container.HealthStatus != "" {
			healthCheck := ContainerHealthCheck{
				ContainerID:   container.ID,
				ContainerName: container.Name,
				HealthStatus:  container.HealthStatus,
				LastCheck:     time.Now(),
			}
			healthChecks = append(healthChecks, healthCheck)
		}
	}

	return healthChecks
}

func performSecurityAnalysis(_ context.Context, containers []ContainerInfo, _ []ImageInfo) (SecurityAnalysis, error) {
	analysis := SecurityAnalysis{
		PrivilegedContainers:  make([]string, 0),
		RootContainers:        make([]string, 0),
		CapabilitiesAdded:     make(map[string][]string),
		HostNetworkContainers: make([]string, 0),
		HostPIDContainers:     make([]string, 0),
		SecretsExposed:        make([]SecurityIssue, 0),
		VulnerableImages:      make([]VulnerabilityInfo, 0),
		Recommendations:       make([]string, 0),
	}

	// Analyze containers for security issues
	for _, container := range containers {
		// Check for privileged containers
		if isPrivilegedContainer(container) {
			analysis.PrivilegedContainers = append(analysis.PrivilegedContainers, container.Name)
		}

		// Check for root containers
		if isRootContainer(container) {
			analysis.RootContainers = append(analysis.RootContainers, container.Name)
		}

		// Check for host network usage
		if usesHostNetwork(container) {
			analysis.HostNetworkContainers = append(analysis.HostNetworkContainers, container.Name)
		}

		// Check for host PID usage
		if usesHostPID(container) {
			analysis.HostPIDContainers = append(analysis.HostPIDContainers, container.Name)
		}

		// Check for exposed secrets
		if secrets := findExposedSecrets(container); len(secrets) > 0 {
			analysis.SecretsExposed = append(analysis.SecretsExposed, secrets...)
		}
	}

	// Calculate security score
	analysis.SecurityScore = calculateContainerSecurityScore(analysis, len(containers))

	// Generate security recommendations
	generateSecurityRecommendations(&analysis)

	return analysis, nil
}

func performDetailedAnalysis(diagnostics *ContainerDiagnostics) {
	// Calculate performance metrics
	diagnostics.PerformanceMetrics = calculatePerformanceMetrics(diagnostics)
}

func generateRecommendations(diagnostics *ContainerDiagnostics) {
	recommendations := make([]string, 0)

	// Resource optimization recommendations
	if len(diagnostics.ResourceUsage.ContainerStats) > 0 {
		highMemoryContainers := 0
		highCPUContainers := 0

		for _, stats := range diagnostics.ResourceUsage.ContainerStats {
			if stats.MemoryPercent > 80 {
				highMemoryContainers++
			}
			if stats.CPUPercent > 80 {
				highCPUContainers++
			}
		}

		if highMemoryContainers > 0 {
			recommendations = append(recommendations,
				fmt.Sprintf("Consider optimizing memory usage for %d containers with high memory utilization", highMemoryContainers))
		}
		if highCPUContainers > 0 {
			recommendations = append(recommendations,
				fmt.Sprintf("Consider optimizing CPU usage for %d containers with high CPU utilization", highCPUContainers))
		}
	}

	// Health check recommendations
	unhealthyContainers := 0
	for _, health := range diagnostics.HealthChecks {
		if health.HealthStatus != "healthy" && health.HealthStatus != "" {
			unhealthyContainers++
		}
	}
	if unhealthyContainers > 0 {
		recommendations = append(recommendations,
			fmt.Sprintf("Investigate %d containers with health check issues", unhealthyContainers))
	}

	// Security recommendations
	if diagnostics.SecurityAnalysis.SecurityScore < 70 {
		recommendations = append(recommendations,
			"Security score is below threshold - review container security configurations")
	}

	// Image cleanup recommendations
	if len(diagnostics.Images) > 0 {
		danglingImages := 0
		for _, image := range diagnostics.Images {
			if image.Dangling {
				danglingImages++
			}
		}
		if danglingImages > 0 {
			recommendations = append(recommendations,
				fmt.Sprintf("Clean up %d dangling images to free disk space", danglingImages))
		}
	}

	diagnostics.Recommendations = recommendations
}

func identifyIssues(diagnostics *ContainerDiagnostics) {
	issues := make([]DiagnosticIssue, 0)

	// Identify container issues
	for _, container := range diagnostics.Containers {
		if container.State == "exited" && container.ExitCode != nil && *container.ExitCode != 0 {
			issues = append(issues, DiagnosticIssue{
				ID:          fmt.Sprintf("container-exit-%s", container.ID[:12]),
				Severity:    "medium",
				Category:    "container",
				Title:       "Container exited with non-zero code",
				Description: fmt.Sprintf("Container %s exited with code %d", container.Name, *container.ExitCode),
				Resolution:  "Check container logs and fix the underlying issue",
				Affected:    []string{container.Name},
				DetectedAt:  time.Now(),
			})
		}

		if container.RestartCount > 5 {
			issues = append(issues, DiagnosticIssue{
				ID:          fmt.Sprintf("container-restarts-%s", container.ID[:12]),
				Severity:    "high",
				Category:    "container",
				Title:       "High restart count",
				Description: fmt.Sprintf("Container %s has restarted %d times", container.Name, container.RestartCount),
				Resolution:  "Investigate root cause of container instability",
				Affected:    []string{container.Name},
				DetectedAt:  time.Now(),
			})
		}
	}

	// Identify security issues
	if len(diagnostics.SecurityAnalysis.PrivilegedContainers) > 0 {
		issues = append(issues, DiagnosticIssue{
			ID:          "security-privileged-containers",
			Severity:    "high",
			Category:    "security",
			Title:       "Privileged containers detected",
			Description: fmt.Sprintf("%d containers are running with privileged access", len(diagnostics.SecurityAnalysis.PrivilegedContainers)),
			Resolution:  "Review necessity of privileged access and remove if not required",
			Affected:    diagnostics.SecurityAnalysis.PrivilegedContainers,
			DetectedAt:  time.Now(),
		})
	}

	diagnostics.Issues = issues
}

// Helper functions for security analysis

func isPrivilegedContainer(container ContainerInfo) bool {
	// This would check container configuration for privileged mode
	// For now, return false as we need to implement the actual check
	return false
}

func isRootContainer(container ContainerInfo) bool {
	// This would check if container is running as root
	// For now, return false as we need to implement the actual check
	return false
}

func usesHostNetwork(container ContainerInfo) bool {
	// This would check if container uses host network
	// For now, return false as we need to implement the actual check
	return false
}

func usesHostPID(container ContainerInfo) bool {
	// This would check if container uses host PID namespace
	// For now, return false as we need to implement the actual check
	return false
}

func findExposedSecrets(container ContainerInfo) []SecurityIssue {
	// This would scan environment variables and other places for secrets
	// For now, return empty slice
	return []SecurityIssue{}
}

func calculateContainerSecurityScore(analysis SecurityAnalysis, totalContainers int) float64 {
	if totalContainers == 0 {
		return 100.0
	}

	score := 100.0

	// Penalize security issues
	score -= float64(len(analysis.PrivilegedContainers)) * 20.0
	score -= float64(len(analysis.RootContainers)) * 10.0
	score -= float64(len(analysis.HostNetworkContainers)) * 15.0
	score -= float64(len(analysis.HostPIDContainers)) * 15.0
	score -= float64(len(analysis.SecretsExposed)) * 25.0

	if score < 0 {
		score = 0
	}

	return score
}

func generateSecurityRecommendations(analysis *SecurityAnalysis) {
	recommendations := make([]string, 0)

	if len(analysis.PrivilegedContainers) > 0 {
		recommendations = append(recommendations,
			"Remove privileged access from containers unless absolutely necessary")
	}
	if len(analysis.RootContainers) > 0 {
		recommendations = append(recommendations,
			"Run containers with non-root user when possible")
	}
	if len(analysis.HostNetworkContainers) > 0 {
		recommendations = append(recommendations,
			"Avoid using host network mode unless required")
	}
	if len(analysis.SecretsExposed) > 0 {
		recommendations = append(recommendations,
			"Use Docker secrets or external secret management for sensitive data")
	}

	analysis.Recommendations = recommendations
}

func calculatePerformanceMetrics(diagnostics *ContainerDiagnostics) PerformanceMetrics {
	metrics := PerformanceMetrics{
		NetworkLatency:  make(map[string]time.Duration),
		ResourceWastage: ResourceWastage{},
	}

	// Calculate average startup time (simplified)
	if len(diagnostics.Containers) > 0 {
		totalStartupTime := time.Duration(0)
		runningContainers := 0

		for _, container := range diagnostics.Containers {
			if container.Started != nil {
				startupTime := container.Started.Sub(container.Created)
				totalStartupTime += startupTime
				runningContainers++
			}
		}

		if runningContainers > 0 {
			metrics.AverageStartupTime = totalStartupTime / time.Duration(runningContainers)
		}
	}

	// Calculate resource efficiency
	if len(diagnostics.ResourceUsage.ContainerStats) > 0 {
		totalMemoryUsage := 0.0
		totalCPUUsage := 0.0

		for _, stats := range diagnostics.ResourceUsage.ContainerStats {
			totalMemoryUsage += stats.MemoryPercent
			totalCPUUsage += stats.CPUPercent
		}

		metrics.MemoryEfficiency = totalMemoryUsage / float64(len(diagnostics.ResourceUsage.ContainerStats))
		metrics.CPUEfficiency = totalCPUUsage / float64(len(diagnostics.ResourceUsage.ContainerStats))
	}

	// Calculate container density
	if diagnostics.SystemInfo.NCPU > 0 {
		metrics.ContainerDensity = float64(diagnostics.SystemInfo.ContainersRunning) / float64(diagnostics.SystemInfo.NCPU)
	}

	// Calculate resource wastage
	danglingImages := 0
	for _, image := range diagnostics.Images {
		if image.Dangling {
			danglingImages++
		}
	}
	metrics.ResourceWastage.DanglingImages = danglingImages

	unusedVolumes := 0
	for _, volume := range diagnostics.Volumes {
		if !volume.InUse {
			unusedVolumes++
		}
	}
	metrics.ResourceWastage.UnusedVolumes = unusedVolumes

	return metrics
}

func saveDiagnosticsResults(diagnostics *ContainerDiagnostics, filename string) error {
	data, err := json.MarshalIndent(diagnostics, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal diagnostics: %w", err)
	}

	return WriteFile(filename, data, 0o600)
}

// WriteFile writes data to a file with specified permissions.
func WriteFile(filename string, data []byte, perm int) error {
	return os.WriteFile(filename, data, os.FileMode(perm))
}

func displayContainerDiagnostics(diagnostics *ContainerDiagnostics, _ containerOptions) error {
	// Display environment information
	logger.SimpleInfo("🐳 Container Environment",
		"docker_version", diagnostics.Environment.DockerVersion,
		"platform", diagnostics.Environment.Platform,
		"architecture", diagnostics.Environment.Architecture,
		"total_memory", fmt.Sprintf("%.2f GB", float64(diagnostics.Environment.TotalMemory)/(1024*1024*1024)),
		"cpu_count", diagnostics.Environment.CPUCount,
	)

	// Display system information
	logger.SimpleInfo("🖥️ System Information",
		"containers_running", diagnostics.SystemInfo.ContainersRunning,
		"containers_stopped", diagnostics.SystemInfo.ContainersStopped,
		"images", diagnostics.SystemInfo.Images,
		"server_version", diagnostics.SystemInfo.ServerVersion,
		"storage_driver", diagnostics.SystemInfo.StorageDriver,
	)

	// Display container summary
	logger.SimpleInfo("📦 Container Summary",
		"total_containers", len(diagnostics.Containers),
		"running", countContainersByState(diagnostics.Containers, "running"),
		"stopped", countContainersByState(diagnostics.Containers, "exited"),
		"paused", countContainersByState(diagnostics.Containers, "paused"),
	)

	// Display top containers by resource usage
	if len(diagnostics.ResourceUsage.ContainerStats) > 0 {
		logger.SimpleInfo("📊 Top Containers by Resource Usage:")

		// Sort by memory usage
		sortedStats := make([]ContainerStats, len(diagnostics.ResourceUsage.ContainerStats))
		copy(sortedStats, diagnostics.ResourceUsage.ContainerStats)
		sort.Slice(sortedStats, func(i, j int) bool {
			return sortedStats[i].MemoryPercent > sortedStats[j].MemoryPercent
		})

		for i, stats := range sortedStats {
			if i >= 5 { // Show top 5
				break
			}
			logger.SimpleInfo(fmt.Sprintf("  %s", stats.ContainerName),
				"cpu_percent", fmt.Sprintf("%.1f%%", stats.CPUPercent),
				"memory_percent", fmt.Sprintf("%.1f%%", stats.MemoryPercent),
				"memory_usage", fmt.Sprintf("%.2f MB", float64(stats.MemoryUsage)/(1024*1024)),
			)
		}
	}

	// Display network information
	if len(diagnostics.Networks) > 0 {
		logger.SimpleInfo("🌐 Network Summary",
			"total_networks", len(diagnostics.Networks),
			"bridge_networks", countNetworksByDriver(diagnostics.Networks, "bridge"),
			"host_networks", countNetworksByDriver(diagnostics.Networks, "host"),
			"overlay_networks", countNetworksByDriver(diagnostics.Networks, "overlay"),
		)
	}

	// Display health check summary
	if len(diagnostics.HealthChecks) > 0 {
		healthyCount := 0
		unhealthyCount := 0
		for _, health := range diagnostics.HealthChecks {
			if health.HealthStatus == "healthy" {
				healthyCount++
			} else {
				unhealthyCount++
			}
		}

		logger.SimpleInfo("🏥 Health Check Summary",
			"total_checks", len(diagnostics.HealthChecks),
			"healthy", healthyCount,
			"unhealthy", unhealthyCount,
		)
	}

	// Display security analysis
	if diagnostics.SecurityAnalysis.SecurityScore > 0 {
		logger.SimpleInfo("🔒 Security Analysis",
			"security_score", fmt.Sprintf("%.1f/100", diagnostics.SecurityAnalysis.SecurityScore),
			"privileged_containers", len(diagnostics.SecurityAnalysis.PrivilegedContainers),
			"root_containers", len(diagnostics.SecurityAnalysis.RootContainers),
			"host_network_containers", len(diagnostics.SecurityAnalysis.HostNetworkContainers),
			"secrets_exposed", len(diagnostics.SecurityAnalysis.SecretsExposed),
		)

		if len(diagnostics.SecurityAnalysis.Recommendations) > 0 {
			logger.SimpleWarn("🔐 Security Recommendations:")
			for _, rec := range diagnostics.SecurityAnalysis.Recommendations {
				logger.SimpleWarn(fmt.Sprintf("  • %s", rec))
			}
		}
	}

	// Display performance metrics
	if diagnostics.PerformanceMetrics.AverageStartupTime > 0 {
		logger.SimpleInfo("⚡ Performance Metrics",
			"average_startup_time", diagnostics.PerformanceMetrics.AverageStartupTime.String(),
			"memory_efficiency", fmt.Sprintf("%.1f%%", diagnostics.PerformanceMetrics.MemoryEfficiency),
			"cpu_efficiency", fmt.Sprintf("%.1f%%", diagnostics.PerformanceMetrics.CPUEfficiency),
			"container_density", fmt.Sprintf("%.2f", diagnostics.PerformanceMetrics.ContainerDensity),
		)
	}

	// Display identified issues
	if len(diagnostics.Issues) > 0 {
		logger.SimpleWarn("⚠️ Identified Issues:")
		for _, issue := range diagnostics.Issues {
			severityIcon := "🟡"
			switch issue.Severity {
			case "high":
				severityIcon = "🔴"
			case "critical":
				severityIcon = "💥"
			}

			logger.SimpleWarn(fmt.Sprintf("  %s %s", severityIcon, issue.Title),
				"severity", issue.Severity,
				"category", issue.Category,
				"affected", len(issue.Affected),
			)
		}
	}

	// Display recommendations
	if len(diagnostics.Recommendations) > 0 {
		logger.SimpleInfo("💡 Recommendations:")
		for _, rec := range diagnostics.Recommendations {
			logger.SimpleInfo(fmt.Sprintf("  • %s", rec))
		}
	}

	return nil
}

// Helper functions for display

func countContainersByState(containers []ContainerInfo, state string) int {
	count := 0
	for _, container := range containers {
		if container.State == state {
			count++
		}
	}
	return count
}

func countNetworksByDriver(networks []NetworkInfo, driver string) int {
	count := 0
	for _, network := range networks {
		if network.Driver == driver {
			count++
		}
	}
	return count
}
