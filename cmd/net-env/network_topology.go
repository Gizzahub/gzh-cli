package netenv

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// NetworkTopologyAnalyzer analyzes network topology and relationships
type NetworkTopologyAnalyzer struct {
	logger            *zap.Logger
	containerDetector *ContainerDetector
	cachedTopology    *NetworkTopology
	cacheMutex        sync.RWMutex
	cacheExpiry       time.Duration
	lastAnalysis      time.Time
}

// NetworkTopology represents the complete network topology
type NetworkTopology struct {
	GeneratedAt     time.Time               `json:"generated_at"`
	Networks        []TopologyNetwork       `json:"networks"`
	Containers      []TopologyContainer     `json:"containers"`
	Services        []TopologyService       `json:"services"`
	Connections     []NetworkConnection     `json:"connections"`
	Dependencies    []ServiceDependency     `json:"dependencies"`
	Clusters        []NetworkCluster        `json:"clusters"`
	Summary         TopologySummary         `json:"summary"`
	AnalysisMetrics TopologyAnalysisMetrics `json:"analysis_metrics"`
}

// TopologyNetwork represents a network in the topology
type TopologyNetwork struct {
	ID                  string            `json:"id"`
	Name                string            `json:"name"`
	Driver              string            `json:"driver"`
	Scope               string            `json:"scope"`
	Subnet              string            `json:"subnet"`
	Gateway             string            `json:"gateway"`
	Internal            bool              `json:"internal"`
	Attachable          bool              `json:"attachable"`
	ConnectedContainers []string          `json:"connected_containers"`
	NetworkType         NetworkType       `json:"network_type"`
	Labels              map[string]string `json:"labels"`
	Options             map[string]string `json:"options"`
	Runtime             ContainerRuntime  `json:"runtime"`
}

// TopologyContainer represents a container in the topology
type TopologyContainer struct {
	ID                string                      `json:"id"`
	Name              string                      `json:"name"`
	Image             string                      `json:"image"`
	State             string                      `json:"state"`
	Runtime           ContainerRuntime            `json:"runtime"`
	NetworkInterfaces []ContainerNetworkInterface `json:"network_interfaces"`
	ExposedPorts      []ContainerPort             `json:"exposed_ports"`
	ServiceLabels     map[string]string           `json:"service_labels"`
	DiscoveryInfo     ServiceDiscoveryInfo        `json:"discovery_info"`
	HealthStatus      string                      `json:"health_status"`
	ResourceLimits    *ContainerResourceLimits    `json:"resource_limits,omitempty"`
}

// TopologyService represents a logical service in the topology
type TopologyService struct {
	Name          string               `json:"name"`
	Type          ServiceType          `json:"type"`
	Containers    []string             `json:"containers"`
	Endpoints     []ServiceEndpoint    `json:"endpoints"`
	LoadBalancer  *LoadBalancerConfig  `json:"load_balancer,omitempty"`
	ServiceMesh   *ServiceMeshConfig   `json:"service_mesh,omitempty"`
	HealthChecks  []HealthCheckConfig  `json:"health_checks"`
	TrafficPolicy *TrafficPolicyConfig `json:"traffic_policy,omitempty"`
	Labels        map[string]string    `json:"labels"`
	Annotations   map[string]string    `json:"annotations"`
}

// NetworkConnection represents a connection between network entities
type NetworkConnection struct {
	ID         string              `json:"id"`
	Source     ConnectionNode      `json:"source"`
	Target     ConnectionNode      `json:"target"`
	Protocol   string              `json:"protocol"`
	Port       int                 `json:"port"`
	Direction  ConnectionDirection `json:"direction"`
	Status     ConnectionStatus    `json:"status"`
	Bandwidth  int64               `json:"bandwidth,omitempty"`
	Latency    time.Duration       `json:"latency,omitempty"`
	PacketLoss float64             `json:"packet_loss,omitempty"`
	LastSeen   time.Time           `json:"last_seen"`
}

// ServiceDependency represents a dependency between services
type ServiceDependency struct {
	SourceService  string                        `json:"source_service"`
	TargetService  string                        `json:"target_service"`
	DependencyType DependencyType                `json:"dependency_type"`
	Protocol       string                        `json:"protocol"`
	Ports          []int                         `json:"ports"`
	Required       bool                          `json:"required"`
	HealthImpact   HealthImpactLevel             `json:"health_impact"`
	CircuitBreaker *TopologyCircuitBreakerConfig `json:"circuit_breaker,omitempty"`
}

// NetworkCluster represents a logical grouping of network entities
type NetworkCluster struct {
	ID        string              `json:"id"`
	Name      string              `json:"name"`
	Type      ClusterType         `json:"type"`
	Members   []ClusterMember     `json:"members"`
	Subnets   []string            `json:"subnets"`
	Isolation IsolationLevel      `json:"isolation"`
	Policies  []NetworkPolicyRule `json:"policies"`
	Gateway   *ClusterGateway     `json:"gateway,omitempty"`
	Metadata  map[string]string   `json:"metadata"`
}

// TopologySummary provides high-level topology statistics
type TopologySummary struct {
	TotalNetworks      int                       `json:"total_networks"`
	TotalContainers    int                       `json:"total_containers"`
	TotalServices      int                       `json:"total_services"`
	TotalConnections   int                       `json:"total_connections"`
	TotalClusters      int                       `json:"total_clusters"`
	NetworksByDriver   map[string]int            `json:"networks_by_driver"`
	ContainersByState  map[string]int            `json:"containers_by_state"`
	ServicesByType     map[string]int            `json:"services_by_type"`
	TopologyComplexity TopologyComplexityMetrics `json:"topology_complexity"`
}

// TopologyAnalysisMetrics contains analysis performance metrics
type TopologyAnalysisMetrics struct {
	AnalysisDuration  time.Duration  `json:"analysis_duration"`
	DiscoveryDuration time.Duration  `json:"discovery_duration"`
	MappingDuration   time.Duration  `json:"mapping_duration"`
	ConnectionTests   int            `json:"connection_tests"`
	SuccessfulTests   int            `json:"successful_tests"`
	FailedTests       int            `json:"failed_tests"`
	CacheHitRate      float64        `json:"cache_hit_rate"`
	DataSourceCounts  map[string]int `json:"data_source_counts"`
}

// Supporting types
type (
	NetworkType         string
	ServiceType         string
	ConnectionDirection string
	ConnectionStatus    string
	DependencyType      string
	HealthImpactLevel   string
	ClusterType         string
	IsolationLevel      string
)

const (
	// NetworkType constants
	NetworkTypeBridge  NetworkType = "bridge"
	NetworkTypeHost    NetworkType = "host"
	NetworkTypeOverlay NetworkType = "overlay"
	NetworkTypeMacvlan NetworkType = "macvlan"
	NetworkTypeCustom  NetworkType = "custom"

	// ServiceType constants
	ServiceTypeWeb      ServiceType = "web"
	ServiceTypeAPI      ServiceType = "api"
	ServiceTypeDatabase ServiceType = "database"
	ServiceTypeCache    ServiceType = "cache"
	ServiceTypeQueue    ServiceType = "queue"
	ServiceTypeWorker   ServiceType = "worker"
	ServiceTypeProxy    ServiceType = "proxy"
	ServiceTypeOther    ServiceType = "other"

	// ConnectionDirection constants
	DirectionInbound       ConnectionDirection = "inbound"
	DirectionOutbound      ConnectionDirection = "outbound"
	DirectionBidirectional ConnectionDirection = "bidirectional"

	// ConnectionStatus constants
	StatusActive  ConnectionStatus = "active"
	StatusIdle    ConnectionStatus = "idle"
	StatusFailed  ConnectionStatus = "failed"
	StatusUnknown ConnectionStatus = "unknown"

	// DependencyType constants
	DependencySynchronous  DependencyType = "synchronous"
	DependencyAsynchronous DependencyType = "asynchronous"
	DependencyOptional     DependencyType = "optional"

	// HealthImpactLevel constants
	HealthImpactCritical HealthImpactLevel = "critical"
	HealthImpactHigh     HealthImpactLevel = "high"
	HealthImpactMedium   HealthImpactLevel = "medium"
	HealthImpactLow      HealthImpactLevel = "low"

	// ClusterType constants
	ClusterTypeNamespace   ClusterType = "namespace"
	ClusterTypeProject     ClusterType = "project"
	ClusterTypeEnvironment ClusterType = "environment"
	ClusterTypeLogical     ClusterType = "logical"

	// IsolationLevel constants
	IsolationStrict     IsolationLevel = "strict"
	IsolationModerate   IsolationLevel = "moderate"
	IsolationPermissive IsolationLevel = "permissive"
)

// Supporting data structures
type ContainerNetworkInterface struct {
	NetworkID   string `json:"network_id"`
	NetworkName string `json:"network_name"`
	IPAddress   string `json:"ip_address"`
	MacAddress  string `json:"mac_address"`
	Gateway     string `json:"gateway"`
	Subnet      string `json:"subnet"`
	MTU         int    `json:"mtu"`
}

type ContainerPort struct {
	ContainerPort int32  `json:"container_port"`
	HostPort      int32  `json:"host_port"`
	HostIP        string `json:"host_ip"`
	Protocol      string `json:"protocol"`
	ServiceName   string `json:"service_name,omitempty"`
}

type ContainerResourceLimits struct {
	MemoryLimit string `json:"memory_limit,omitempty"`
	CPULimit    string `json:"cpu_limit,omitempty"`
	MemoryUsage string `json:"memory_usage,omitempty"`
	CPUUsage    string `json:"cpu_usage,omitempty"`
}

type ServiceDiscoveryInfo struct {
	DNSName       string             `json:"dns_name,omitempty"`
	ServiceTags   []string           `json:"service_tags"`
	ConsulService *ConsulServiceInfo `json:"consul_service,omitempty"`
	EurekaService *EurekaServiceInfo `json:"eureka_service,omitempty"`
}

type ServiceEndpoint struct {
	Address     string             `json:"address"`
	Port        int                `json:"port"`
	Protocol    string             `json:"protocol"`
	Path        string             `json:"path,omitempty"`
	HealthCheck *HealthCheckConfig `json:"health_check,omitempty"`
	Metadata    map[string]string  `json:"metadata"`
}

type LoadBalancerConfig struct {
	Type            string             `json:"type"`
	Algorithm       string             `json:"algorithm"`
	HealthCheck     *HealthCheckConfig `json:"health_check,omitempty"`
	SessionAffinity bool               `json:"session_affinity"`
	Options         map[string]string  `json:"options"`
}

type HealthCheckConfig struct {
	Type             string        `json:"type"`
	Endpoint         string        `json:"endpoint"`
	Interval         time.Duration `json:"interval"`
	Timeout          time.Duration `json:"timeout"`
	Retries          int           `json:"retries"`
	SuccessThreshold int           `json:"success_threshold"`
	FailureThreshold int           `json:"failure_threshold"`
}

type TrafficPolicyConfig struct {
	LoadBalancing  *LoadBalancingPolicy          `json:"load_balancing,omitempty"`
	Retries        *RetryPolicy                  `json:"retries,omitempty"`
	Timeout        *TimeoutPolicy                `json:"timeout,omitempty"`
	CircuitBreaker *TopologyCircuitBreakerConfig `json:"circuit_breaker,omitempty"`
}

type ConnectionNode struct {
	ID       string            `json:"id"`
	Type     string            `json:"type"` // container, service, external
	Name     string            `json:"name"`
	Address  string            `json:"address"`
	Port     int               `json:"port,omitempty"`
	Metadata map[string]string `json:"metadata"`
}

type ClusterMember struct {
	ID       string            `json:"id"`
	Type     string            `json:"type"` // container, service, network
	Name     string            `json:"name"`
	Role     string            `json:"role,omitempty"`
	Metadata map[string]string `json:"metadata"`
}

type NetworkPolicyRule struct {
	Action      string         `json:"action"`    // allow, deny
	Direction   string         `json:"direction"` // ingress, egress
	Source      PolicySelector `json:"source"`
	Destination PolicySelector `json:"destination"`
	Ports       []int          `json:"ports,omitempty"`
	Protocols   []string       `json:"protocols,omitempty"`
}

type ClusterGateway struct {
	Type         string              `json:"type"`
	Address      string              `json:"address"`
	Ports        []int               `json:"ports"`
	Protocols    []string            `json:"protocols"`
	LoadBalancer *LoadBalancerConfig `json:"load_balancer,omitempty"`
}

type PolicySelector struct {
	Type      string            `json:"type"`
	Labels    map[string]string `json:"labels,omitempty"`
	Addresses []string          `json:"addresses,omitempty"`
}

type TopologyComplexityMetrics struct {
	NetworkComplexity    float64 `json:"network_complexity"`
	ServiceComplexity    float64 `json:"service_complexity"`
	ConnectionDensity    float64 `json:"connection_density"`
	CyclomaticComplexity int     `json:"cyclomatic_complexity"`
	MaxDepth             int     `json:"max_depth"`
	BranchingFactor      float64 `json:"branching_factor"`
}

type LoadBalancingPolicy struct {
	Algorithm string            `json:"algorithm"`
	Options   map[string]string `json:"options"`
}

type RetryPolicy struct {
	MaxRetries      int           `json:"max_retries"`
	InitialInterval time.Duration `json:"initial_interval"`
	MaxInterval     time.Duration `json:"max_interval"`
	Multiplier      float64       `json:"multiplier"`
}

type TimeoutPolicy struct {
	ConnectTimeout time.Duration `json:"connect_timeout"`
	RequestTimeout time.Duration `json:"request_timeout"`
	IdleTimeout    time.Duration `json:"idle_timeout"`
}

type TopologyCircuitBreakerConfig struct {
	FailureThreshold int           `json:"failure_threshold"`
	RecoveryTimeout  time.Duration `json:"recovery_timeout"`
	SampleSize       int           `json:"sample_size"`
	HalfOpenRequests int           `json:"half_open_requests"`
}

type ConsulServiceInfo struct {
	ServiceID string            `json:"service_id"`
	Tags      []string          `json:"tags"`
	Meta      map[string]string `json:"meta"`
	Address   string            `json:"address"`
	Port      int               `json:"port"`
}

type EurekaServiceInfo struct {
	AppName    string            `json:"app_name"`
	InstanceID string            `json:"instance_id"`
	Status     string            `json:"status"`
	VIPAddress string            `json:"vip_address"`
	Metadata   map[string]string `json:"metadata"`
}

// NewNetworkTopologyAnalyzer creates a new network topology analyzer
func NewNetworkTopologyAnalyzer(logger *zap.Logger, containerDetector *ContainerDetector) *NetworkTopologyAnalyzer {
	return &NetworkTopologyAnalyzer{
		logger:            logger,
		containerDetector: containerDetector,
		cacheExpiry:       60 * time.Second, // Cache results for 60 seconds
	}
}

// AnalyzeNetworkTopology performs comprehensive network topology analysis
func (nta *NetworkTopologyAnalyzer) AnalyzeNetworkTopology(ctx context.Context) (*NetworkTopology, error) {
	nta.cacheMutex.RLock()
	if nta.cachedTopology != nil && time.Since(nta.lastAnalysis) < nta.cacheExpiry {
		nta.cacheMutex.RUnlock()
		return nta.cachedTopology, nil
	}
	nta.cacheMutex.RUnlock()

	startTime := time.Now()
	nta.logger.Info("Starting network topology analysis")

	topology := &NetworkTopology{
		GeneratedAt: time.Now(),
		AnalysisMetrics: TopologyAnalysisMetrics{
			DataSourceCounts: make(map[string]int),
		},
	}

	// Step 1: Discover container environment
	discoveryStart := time.Now()
	containerEnv, err := nta.containerDetector.DetectContainerEnvironment(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to detect container environment: %w", err)
	}
	topology.AnalysisMetrics.DiscoveryDuration = time.Since(discoveryStart)

	// Step 2: Map networks
	networks, err := nta.mapNetworks(ctx, containerEnv)
	if err != nil {
		nta.logger.Warn("Failed to map networks", zap.Error(err))
	} else {
		topology.Networks = networks
		topology.AnalysisMetrics.DataSourceCounts["networks"] = len(networks)
	}

	// Step 3: Map containers with network interfaces
	mappingStart := time.Now()
	containers, err := nta.mapContainers(ctx, containerEnv)
	if err != nil {
		nta.logger.Warn("Failed to map containers", zap.Error(err))
	} else {
		topology.Containers = containers
		topology.AnalysisMetrics.DataSourceCounts["containers"] = len(containers)
	}

	// Step 4: Discover and map services
	services, err := nta.discoverServices(ctx, containers)
	if err != nil {
		nta.logger.Warn("Failed to discover services", zap.Error(err))
	} else {
		topology.Services = services
		topology.AnalysisMetrics.DataSourceCounts["services"] = len(services)
	}

	// Step 5: Analyze network connections
	connections, err := nta.analyzeConnections(ctx, containers, networks)
	if err != nil {
		nta.logger.Warn("Failed to analyze connections", zap.Error(err))
	} else {
		topology.Connections = connections
		topology.AnalysisMetrics.DataSourceCounts["connections"] = len(connections)
	}

	// Step 6: Map service dependencies
	dependencies, err := nta.mapServiceDependencies(ctx, services, connections)
	if err != nil {
		nta.logger.Warn("Failed to map service dependencies", zap.Error(err))
	} else {
		topology.Dependencies = dependencies
		topology.AnalysisMetrics.DataSourceCounts["dependencies"] = len(dependencies)
	}

	// Step 7: Identify network clusters
	clusters, err := nta.identifyNetworkClusters(ctx, networks, containers, services)
	if err != nil {
		nta.logger.Warn("Failed to identify network clusters", zap.Error(err))
	} else {
		topology.Clusters = clusters
		topology.AnalysisMetrics.DataSourceCounts["clusters"] = len(clusters)
	}

	topology.AnalysisMetrics.MappingDuration = time.Since(mappingStart)

	// Step 8: Generate summary and complexity metrics
	topology.Summary = nta.generateTopologySummary(topology)
	topology.AnalysisMetrics.AnalysisDuration = time.Since(startTime)

	// Cache the result
	nta.cacheMutex.Lock()
	nta.cachedTopology = topology
	nta.lastAnalysis = time.Now()
	nta.cacheMutex.Unlock()

	nta.logger.Info("Network topology analysis completed",
		zap.Duration("duration", topology.AnalysisMetrics.AnalysisDuration),
		zap.Int("networks", len(topology.Networks)),
		zap.Int("containers", len(topology.Containers)),
		zap.Int("services", len(topology.Services)),
		zap.Int("connections", len(topology.Connections)))

	return topology, nil
}

// mapNetworks maps detected networks to topology networks
func (nta *NetworkTopologyAnalyzer) mapNetworks(ctx context.Context, containerEnv *ContainerEnvironment) ([]TopologyNetwork, error) {
	var topologyNetworks []TopologyNetwork

	for _, network := range containerEnv.Networks {
		topologyNetwork := TopologyNetwork{
			ID:                  network.ID,
			Name:                network.Name,
			Driver:              network.Driver,
			Scope:               network.Scope,
			Internal:            network.Internal,
			NetworkType:         nta.determineNetworkType(network.Driver),
			Labels:              network.Labels,
			Options:             network.Options,
			Runtime:             network.Runtime,
			ConnectedContainers: make([]string, 0),
		}

		// Extract subnet and gateway from IPAM config
		if len(network.IPAM.Config) > 0 {
			topologyNetwork.Subnet = network.IPAM.Config[0].Subnet
			topologyNetwork.Gateway = network.IPAM.Config[0].Gateway
		}

		// Find connected containers
		for _, container := range containerEnv.RunningContainers {
			for _, netInfo := range container.Networks {
				if netInfo.NetworkID == network.ID || netInfo.NetworkName == network.Name {
					topologyNetwork.ConnectedContainers = append(topologyNetwork.ConnectedContainers, container.ID)
					break
				}
			}
		}

		topologyNetworks = append(topologyNetworks, topologyNetwork)
	}

	return topologyNetworks, nil
}

// mapContainers maps detected containers to topology containers
func (nta *NetworkTopologyAnalyzer) mapContainers(ctx context.Context, containerEnv *ContainerEnvironment) ([]TopologyContainer, error) {
	var topologyContainers []TopologyContainer

	for _, container := range containerEnv.RunningContainers {
		topologyContainer := TopologyContainer{
			ID:                container.ID,
			Name:              container.Name,
			Image:             container.Image,
			State:             container.State,
			Runtime:           container.Runtime,
			ServiceLabels:     container.Labels,
			HealthStatus:      container.HealthStatus,
			ResourceLimits:    nta.convertResourceLimits(container.ResourceLimits),
			NetworkInterfaces: make([]ContainerNetworkInterface, 0),
			ExposedPorts:      make([]ContainerPort, 0),
		}

		// Map network interfaces
		for _, netInfo := range container.Networks {
			networkInterface := ContainerNetworkInterface{
				NetworkID:   netInfo.NetworkID,
				NetworkName: netInfo.NetworkName,
				IPAddress:   netInfo.IPAddress,
				MacAddress:  netInfo.MacAddress,
				Gateway:     netInfo.Gateway,
				Subnet:      netInfo.Subnet,
			}
			topologyContainer.NetworkInterfaces = append(topologyContainer.NetworkInterfaces, networkInterface)
		}

		// Map exposed ports
		for _, port := range container.Ports {
			containerPort := ContainerPort{
				ContainerPort: port.ContainerPort,
				HostPort:      port.HostPort,
				HostIP:        port.HostIP,
				Protocol:      port.Protocol,
			}
			topologyContainer.ExposedPorts = append(topologyContainer.ExposedPorts, containerPort)
		}

		// Extract service discovery information
		topologyContainer.DiscoveryInfo = nta.extractServiceDiscoveryInfo(container)

		topologyContainers = append(topologyContainers, topologyContainer)
	}

	return topologyContainers, nil
}

// discoverServices discovers logical services from containers
func (nta *NetworkTopologyAnalyzer) discoverServices(ctx context.Context, containers []TopologyContainer) ([]TopologyService, error) {
	serviceMap := make(map[string]*TopologyService)

	for _, container := range containers {
		serviceName := nta.extractServiceName(container)
		if serviceName == "" {
			continue
		}

		service, exists := serviceMap[serviceName]
		if !exists {
			service = &TopologyService{
				Name:         serviceName,
				Type:         nta.determineServiceType(container),
				Containers:   make([]string, 0),
				Endpoints:    make([]ServiceEndpoint, 0),
				HealthChecks: make([]HealthCheckConfig, 0),
				Labels:       make(map[string]string),
				Annotations:  make(map[string]string),
			}
			serviceMap[serviceName] = service
		}

		service.Containers = append(service.Containers, container.ID)

		// Extract endpoints from exposed ports
		for _, port := range container.ExposedPorts {
			endpoint := ServiceEndpoint{
				Address:  nta.getContainerIPAddress(container),
				Port:     int(port.ContainerPort),
				Protocol: port.Protocol,
				Metadata: make(map[string]string),
			}
			service.Endpoints = append(service.Endpoints, endpoint)
		}

		// Merge labels and annotations
		for k, v := range container.ServiceLabels {
			if strings.HasPrefix(k, "service.") {
				service.Labels[strings.TrimPrefix(k, "service.")] = v
			} else if strings.HasPrefix(k, "annotation.") {
				service.Annotations[strings.TrimPrefix(k, "annotation.")] = v
			}
		}
	}

	// Convert map to slice
	var services []TopologyService
	for _, service := range serviceMap {
		services = append(services, *service)
	}

	return services, nil
}

// analyzeConnections analyzes network connections between containers
func (nta *NetworkTopologyAnalyzer) analyzeConnections(ctx context.Context, containers []TopologyContainer, networks []TopologyNetwork) ([]NetworkConnection, error) {
	var connections []NetworkConnection

	// Analyze container-to-container connections within same networks
	for i, container1 := range containers {
		for j, container2 := range containers {
			if i >= j {
				continue // Avoid duplicates and self-connections
			}

			// Check if containers are on the same network
			sharedNetworks := nta.findSharedNetworks(container1, container2)
			if len(sharedNetworks) == 0 {
				continue
			}

			// Test connectivity for each exposed port
			for _, port := range container2.ExposedPorts {
				connectionID := fmt.Sprintf("%s-%s-%d", container1.ID[:12], container2.ID[:12], port.ContainerPort)

				connection := NetworkConnection{
					ID: connectionID,
					Source: ConnectionNode{
						ID:      container1.ID,
						Type:    "container",
						Name:    container1.Name,
						Address: nta.getContainerIPAddress(container1),
					},
					Target: ConnectionNode{
						ID:      container2.ID,
						Type:    "container",
						Name:    container2.Name,
						Address: nta.getContainerIPAddress(container2),
						Port:    int(port.ContainerPort),
					},
					Protocol:  port.Protocol,
					Port:      int(port.ContainerPort),
					Direction: DirectionOutbound,
					Status:    nta.testConnection(ctx, nta.getContainerIPAddress(container2), int(port.ContainerPort)),
					LastSeen:  time.Now(),
				}

				connections = append(connections, connection)
			}
		}
	}

	return connections, nil
}

// mapServiceDependencies maps dependencies between services
func (nta *NetworkTopologyAnalyzer) mapServiceDependencies(ctx context.Context, services []TopologyService, connections []NetworkConnection) ([]ServiceDependency, error) {
	var dependencies []ServiceDependency
	serviceContainerMap := nta.buildServiceContainerMap(services)

	// Analyze dependencies based on connections
	for _, connection := range connections {
		sourceService := nta.findServiceByContainer(serviceContainerMap, connection.Source.ID)
		targetService := nta.findServiceByContainer(serviceContainerMap, connection.Target.ID)

		if sourceService == "" || targetService == "" || sourceService == targetService {
			continue
		}

		// Check if dependency already exists
		dependencyKey := fmt.Sprintf("%s->%s", sourceService, targetService)
		existing := false
		for _, dep := range dependencies {
			if fmt.Sprintf("%s->%s", dep.SourceService, dep.TargetService) == dependencyKey {
				existing = true
				break
			}
		}

		if !existing {
			dependency := ServiceDependency{
				SourceService:  sourceService,
				TargetService:  targetService,
				DependencyType: nta.determineDependencyType(connection),
				Protocol:       connection.Protocol,
				Ports:          []int{connection.Port},
				Required:       nta.isDependencyRequired(connection),
				HealthImpact:   nta.determineHealthImpact(connection),
			}

			dependencies = append(dependencies, dependency)
		}
	}

	return dependencies, nil
}

// identifyNetworkClusters identifies logical network clusters
func (nta *NetworkTopologyAnalyzer) identifyNetworkClusters(ctx context.Context, networks []TopologyNetwork, containers []TopologyContainer, services []TopologyService) ([]NetworkCluster, error) {
	var clusters []NetworkCluster

	// Cluster by network
	for _, network := range networks {
		if len(network.ConnectedContainers) < 2 {
			continue // Skip networks with less than 2 containers
		}

		cluster := NetworkCluster{
			ID:        fmt.Sprintf("network-%s", network.ID[:12]),
			Name:      fmt.Sprintf("Network %s", network.Name),
			Type:      ClusterTypeLogical,
			Members:   make([]ClusterMember, 0),
			Subnets:   []string{network.Subnet},
			Isolation: nta.determineIsolationLevel(network),
			Policies:  make([]NetworkPolicyRule, 0),
			Metadata: map[string]string{
				"network_id":     network.ID,
				"network_driver": network.Driver,
				"network_scope":  network.Scope,
			},
		}

		// Add containers as cluster members
		for _, containerID := range network.ConnectedContainers {
			for _, container := range containers {
				if container.ID == containerID {
					member := ClusterMember{
						ID:   container.ID,
						Type: "container",
						Name: container.Name,
						Metadata: map[string]string{
							"image": container.Image,
							"state": container.State,
						},
					}
					cluster.Members = append(cluster.Members, member)
					break
				}
			}
		}

		clusters = append(clusters, cluster)
	}

	// Cluster by Compose project
	composeProjects := nta.groupContainersByComposeProject(containers)
	for projectName, projectContainers := range composeProjects {
		if len(projectContainers) < 2 {
			continue
		}

		cluster := NetworkCluster{
			ID:        fmt.Sprintf("compose-%s", projectName),
			Name:      fmt.Sprintf("Compose Project %s", projectName),
			Type:      ClusterTypeProject,
			Members:   make([]ClusterMember, 0),
			Subnets:   make([]string, 0),
			Isolation: IsolationModerate,
			Policies:  make([]NetworkPolicyRule, 0),
			Metadata: map[string]string{
				"compose_project": projectName,
				"cluster_type":    "docker_compose",
			},
		}

		for _, container := range projectContainers {
			member := ClusterMember{
				ID:   container.ID,
				Type: "container",
				Name: container.Name,
				Role: nta.extractComposeServiceName(container),
				Metadata: map[string]string{
					"compose_service": nta.extractComposeServiceName(container),
				},
			}
			cluster.Members = append(cluster.Members, member)
		}

		clusters = append(clusters, cluster)
	}

	return clusters, nil
}

// generateTopologySummary generates summary statistics
func (nta *NetworkTopologyAnalyzer) generateTopologySummary(topology *NetworkTopology) TopologySummary {
	summary := TopologySummary{
		TotalNetworks:     len(topology.Networks),
		TotalContainers:   len(topology.Containers),
		TotalServices:     len(topology.Services),
		TotalConnections:  len(topology.Connections),
		TotalClusters:     len(topology.Clusters),
		NetworksByDriver:  make(map[string]int),
		ContainersByState: make(map[string]int),
		ServicesByType:    make(map[string]int),
	}

	// Count networks by driver
	for _, network := range topology.Networks {
		summary.NetworksByDriver[network.Driver]++
	}

	// Count containers by state
	for _, container := range topology.Containers {
		summary.ContainersByState[container.State]++
	}

	// Count services by type
	for _, service := range topology.Services {
		summary.ServicesByType[string(service.Type)]++
	}

	// Calculate complexity metrics
	summary.TopologyComplexity = nta.calculateComplexityMetrics(topology)

	return summary
}

// Helper methods

func (nta *NetworkTopologyAnalyzer) determineNetworkType(driver string) NetworkType {
	switch driver {
	case "bridge":
		return NetworkTypeBridge
	case "host":
		return NetworkTypeHost
	case "overlay":
		return NetworkTypeOverlay
	case "macvlan":
		return NetworkTypeMacvlan
	default:
		return NetworkTypeCustom
	}
}

func (nta *NetworkTopologyAnalyzer) extractServiceName(container TopologyContainer) string {
	// Try compose service label first
	if serviceName, ok := container.ServiceLabels["com.docker.compose.service"]; ok {
		return serviceName
	}

	// Try custom service label
	if serviceName, ok := container.ServiceLabels["service.name"]; ok {
		return serviceName
	}

	// Extract from container name
	parts := strings.Split(container.Name, "_")
	if len(parts) > 1 {
		return parts[len(parts)-2] // Usually service name is second to last
	}

	return ""
}

func (nta *NetworkTopologyAnalyzer) determineServiceType(container TopologyContainer) ServiceType {
	// Check labels for explicit type
	if serviceType, ok := container.ServiceLabels["service.type"]; ok {
		switch serviceType {
		case "web", "http", "frontend":
			return ServiceTypeWeb
		case "api", "rest", "grpc":
			return ServiceTypeAPI
		case "database", "db", "mysql", "postgres", "mongodb":
			return ServiceTypeDatabase
		case "cache", "redis", "memcached":
			return ServiceTypeCache
		case "queue", "kafka", "rabbitmq":
			return ServiceTypeQueue
		case "worker", "job":
			return ServiceTypeWorker
		case "proxy", "nginx", "haproxy":
			return ServiceTypeProxy
		}
	}

	// Infer from image name
	imageName := strings.ToLower(container.Image)
	if strings.Contains(imageName, "nginx") || strings.Contains(imageName, "apache") || strings.Contains(imageName, "frontend") {
		return ServiceTypeWeb
	}
	if strings.Contains(imageName, "postgres") || strings.Contains(imageName, "mysql") || strings.Contains(imageName, "mongo") {
		return ServiceTypeDatabase
	}
	if strings.Contains(imageName, "redis") || strings.Contains(imageName, "memcached") {
		return ServiceTypeCache
	}
	if strings.Contains(imageName, "kafka") || strings.Contains(imageName, "rabbitmq") {
		return ServiceTypeQueue
	}

	// Infer from exposed ports
	for _, port := range container.ExposedPorts {
		switch port.ContainerPort {
		case 80, 8080, 3000, 4200:
			return ServiceTypeWeb
		case 3306, 5432, 27017:
			return ServiceTypeDatabase
		case 6379:
			return ServiceTypeCache
		case 5672, 9092:
			return ServiceTypeQueue
		}
	}

	return ServiceTypeOther
}

func (nta *NetworkTopologyAnalyzer) getContainerIPAddress(container TopologyContainer) string {
	if len(container.NetworkInterfaces) > 0 {
		return container.NetworkInterfaces[0].IPAddress
	}
	return ""
}

func (nta *NetworkTopologyAnalyzer) extractServiceDiscoveryInfo(container DetectedContainer) ServiceDiscoveryInfo {
	info := ServiceDiscoveryInfo{
		ServiceTags: make([]string, 0),
	}

	// Extract DNS name from labels
	if dnsName, ok := container.Labels["service.dns.name"]; ok {
		info.DNSName = dnsName
	}

	// Extract service tags
	if tags, ok := container.Labels["service.tags"]; ok {
		info.ServiceTags = strings.Split(tags, ",")
	}

	return info
}

func (nta *NetworkTopologyAnalyzer) findSharedNetworks(container1, container2 TopologyContainer) []string {
	var shared []string

	for _, net1 := range container1.NetworkInterfaces {
		for _, net2 := range container2.NetworkInterfaces {
			if net1.NetworkID == net2.NetworkID {
				shared = append(shared, net1.NetworkID)
				break
			}
		}
	}

	return shared
}

func (nta *NetworkTopologyAnalyzer) testConnection(ctx context.Context, address string, port int) ConnectionStatus {
	if address == "" {
		return StatusUnknown
	}

	timeout := 3 * time.Second
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", address, port), timeout)
	if err != nil {
		return StatusFailed
	}
	defer conn.Close()

	return StatusActive
}

func (nta *NetworkTopologyAnalyzer) buildServiceContainerMap(services []TopologyService) map[string]string {
	serviceMap := make(map[string]string)
	for _, service := range services {
		for _, containerID := range service.Containers {
			serviceMap[containerID] = service.Name
		}
	}
	return serviceMap
}

func (nta *NetworkTopologyAnalyzer) findServiceByContainer(serviceMap map[string]string, containerID string) string {
	return serviceMap[containerID]
}

func (nta *NetworkTopologyAnalyzer) determineDependencyType(connection NetworkConnection) DependencyType {
	// Simple heuristic based on port and protocol
	if strings.ToLower(connection.Protocol) == "tcp" {
		return DependencySynchronous
	}
	return DependencyAsynchronous
}

func (nta *NetworkTopologyAnalyzer) isDependencyRequired(connection NetworkConnection) bool {
	// Consider connection as required if it's to a database or cache
	return connection.Port == 3306 || connection.Port == 5432 || connection.Port == 6379
}

func (nta *NetworkTopologyAnalyzer) determineHealthImpact(connection NetworkConnection) HealthImpactLevel {
	// Determine impact based on target port
	switch connection.Port {
	case 3306, 5432, 27017: // Databases
		return HealthImpactCritical
	case 6379: // Cache
		return HealthImpactHigh
	case 80, 443, 8080: // Web services
		return HealthImpactMedium
	default:
		return HealthImpactLow
	}
}

func (nta *NetworkTopologyAnalyzer) determineIsolationLevel(network TopologyNetwork) IsolationLevel {
	if network.Internal {
		return IsolationStrict
	}
	if network.Driver == "bridge" {
		return IsolationModerate
	}
	return IsolationPermissive
}

func (nta *NetworkTopologyAnalyzer) groupContainersByComposeProject(containers []TopologyContainer) map[string][]TopologyContainer {
	projects := make(map[string][]TopologyContainer)

	for _, container := range containers {
		projectName := ""
		if project, ok := container.ServiceLabels["com.docker.compose.project"]; ok {
			projectName = project
		}

		if projectName != "" {
			projects[projectName] = append(projects[projectName], container)
		}
	}

	return projects
}

func (nta *NetworkTopologyAnalyzer) extractComposeServiceName(container TopologyContainer) string {
	if serviceName, ok := container.ServiceLabels["com.docker.compose.service"]; ok {
		return serviceName
	}
	return ""
}

func (nta *NetworkTopologyAnalyzer) calculateComplexityMetrics(topology *NetworkTopology) TopologyComplexityMetrics {
	metrics := TopologyComplexityMetrics{}

	totalNodes := len(topology.Containers) + len(topology.Services)
	totalConnections := len(topology.Connections)

	if totalNodes > 0 {
		// Connection density: ratio of actual connections to possible connections
		maxPossibleConnections := totalNodes * (totalNodes - 1) / 2
		if maxPossibleConnections > 0 {
			metrics.ConnectionDensity = float64(totalConnections) / float64(maxPossibleConnections)
		}

		// Average branching factor
		if totalConnections > 0 {
			metrics.BranchingFactor = float64(totalConnections) / float64(totalNodes)
		}
	}

	// Network complexity: consider number of networks and their types
	metrics.NetworkComplexity = float64(len(topology.Networks))
	for _, network := range topology.Networks {
		if network.Driver != "bridge" {
			metrics.NetworkComplexity += 0.5 // More complex networks add weight
		}
	}

	// Service complexity: consider service types and dependencies
	metrics.ServiceComplexity = float64(len(topology.Services))
	for _, service := range topology.Services {
		if service.Type == ServiceTypeAPI || service.Type == ServiceTypeProxy {
			metrics.ServiceComplexity += 0.5 // Complex service types add weight
		}
	}

	// Calculate cyclomatic complexity (simplified)
	metrics.CyclomaticComplexity = totalConnections - totalNodes + 2

	return metrics
}

// InvalidateCache invalidates the cached topology
func (nta *NetworkTopologyAnalyzer) InvalidateCache() {
	nta.cacheMutex.Lock()
	defer nta.cacheMutex.Unlock()
	nta.cachedTopology = nil
	nta.lastAnalysis = time.Time{}
}

// GetCachedTopology returns the cached topology if available
func (nta *NetworkTopologyAnalyzer) GetCachedTopology() *NetworkTopology {
	nta.cacheMutex.RLock()
	defer nta.cacheMutex.RUnlock()
	if nta.cachedTopology != nil && time.Since(nta.lastAnalysis) < nta.cacheExpiry {
		return nta.cachedTopology
	}
	return nil
}

// ExportTopology exports topology to various formats
func (nta *NetworkTopologyAnalyzer) ExportTopology(topology *NetworkTopology, format string) ([]byte, error) {
	switch format {
	case "json":
		return json.MarshalIndent(topology, "", "  ")
	case "dot":
		return nta.exportToDOT(topology), nil
	case "cytoscape":
		return nta.exportToCytoscape(topology)
	default:
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}
}

// exportToDOT exports topology to Graphviz DOT format
func (nta *NetworkTopologyAnalyzer) exportToDOT(topology *NetworkTopology) []byte {
	var dot strings.Builder

	dot.WriteString("digraph NetworkTopology {\n")
	dot.WriteString("  rankdir=TB;\n")
	dot.WriteString("  node [shape=box];\n\n")

	// Add containers as nodes
	for _, container := range topology.Containers {
		dot.WriteString(fmt.Sprintf("  \"%s\" [label=\"%s\\n%s\" shape=box style=filled fillcolor=lightblue];\n",
			container.ID[:12], container.Name, container.Image))
	}

	// Add services as nodes
	for _, service := range topology.Services {
		dot.WriteString(fmt.Sprintf("  \"%s\" [label=\"%s\\n%s\" shape=ellipse style=filled fillcolor=lightgreen];\n",
			service.Name, service.Name, service.Type))
	}

	// Add connections as edges
	for _, conn := range topology.Connections {
		dot.WriteString(fmt.Sprintf("  \"%s\" -> \"%s\" [label=\"%s:%d\"];\n",
			conn.Source.ID[:12], conn.Target.ID[:12], conn.Protocol, conn.Port))
	}

	dot.WriteString("}\n")
	return []byte(dot.String())
}

// exportToCytoscape exports topology to Cytoscape.js format
func (nta *NetworkTopologyAnalyzer) exportToCytoscape(topology *NetworkTopology) ([]byte, error) {
	cytoscapeData := map[string]interface{}{
		"elements": map[string]interface{}{
			"nodes": make([]map[string]interface{}, 0),
			"edges": make([]map[string]interface{}, 0),
		},
	}

	nodes := cytoscapeData["elements"].(map[string]interface{})["nodes"].([]map[string]interface{})
	edges := cytoscapeData["elements"].(map[string]interface{})["edges"].([]map[string]interface{})

	// Add container nodes
	for _, container := range topology.Containers {
		node := map[string]interface{}{
			"data": map[string]interface{}{
				"id":    container.ID[:12],
				"label": container.Name,
				"type":  "container",
				"image": container.Image,
				"state": container.State,
			},
		}
		nodes = append(nodes, node)
	}

	// Add service nodes
	for _, service := range topology.Services {
		node := map[string]interface{}{
			"data": map[string]interface{}{
				"id":          service.Name,
				"label":       service.Name,
				"type":        "service",
				"serviceType": service.Type,
			},
		}
		nodes = append(nodes, node)
	}

	// Add connection edges
	for i, conn := range topology.Connections {
		edge := map[string]interface{}{
			"data": map[string]interface{}{
				"id":       fmt.Sprintf("edge-%d", i),
				"source":   conn.Source.ID[:12],
				"target":   conn.Target.ID[:12],
				"protocol": conn.Protocol,
				"port":     conn.Port,
				"status":   conn.Status,
			},
		}
		edges = append(edges, edge)
	}

	cytoscapeData["elements"].(map[string]interface{})["nodes"] = nodes
	cytoscapeData["elements"].(map[string]interface{})["edges"] = edges

	return json.MarshalIndent(cytoscapeData, "", "  ")
}

// convertResourceLimits converts DetectedResourceLimits to ContainerResourceLimits
func (nta *NetworkTopologyAnalyzer) convertResourceLimits(detected *DetectedResourceLimits) *ContainerResourceLimits {
	if detected == nil {
		return nil
	}

	limits := &ContainerResourceLimits{}

	if detected.Memory > 0 {
		limits.MemoryLimit = fmt.Sprintf("%d", detected.Memory)
	}

	if detected.CPUQuota > 0 && detected.CPUPeriod > 0 {
		cpuLimit := float64(detected.CPUQuota) / float64(detected.CPUPeriod)
		limits.CPULimit = fmt.Sprintf("%.2f", cpuLimit)
	}

	return limits
}
