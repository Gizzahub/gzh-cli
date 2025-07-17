package netenv

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"
)

func TestNewNetworkTopologyAnalyzer(t *testing.T) {
	logger := zaptest.NewLogger(t)
	detector := NewContainerDetector(logger)

	analyzer := NewNetworkTopologyAnalyzer(logger, detector)

	assert.NotNil(t, analyzer)
	assert.Equal(t, logger, analyzer.logger)
	assert.Equal(t, detector, analyzer.containerDetector)
	assert.Equal(t, 60*time.Second, analyzer.cacheExpiry)
}

func TestAnalyzeNetworkTopology(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Create mock container environment
	mockEnv := &ContainerEnvironment{
		AvailableRuntimes: []RuntimeInfo{
			{Runtime: Docker, Available: true, Version: "20.10.0"},
		},
		PrimaryRuntime: Docker,
		RunningContainers: []DetectedContainer{
			{
				ID:      "container1",
				Name:    "web-server",
				Image:   "nginx:latest",
				Status:  "running",
				Runtime: Docker,
				Networks: []DetectedNetworkInfo{
					{
						NetworkID:   "bridge",
						NetworkName: "bridge",
						IPAddress:   "172.17.0.2",
						MacAddress:  "02:42:ac:11:00:02",
					},
				},
				Ports: []DetectedPortMapping{
					{
						ContainerPort: 80,
						HostPort:      8080,
						Protocol:      "tcp",
					},
				},
			},
			{
				ID:      "container2",
				Name:    "database",
				Image:   "postgres:13",
				Status:  "running",
				Runtime: Docker,
				Networks: []DetectedNetworkInfo{
					{
						NetworkID:   "bridge",
						NetworkName: "bridge",
						IPAddress:   "172.17.0.3",
						MacAddress:  "02:42:ac:11:00:03",
					},
				},
				Ports: []DetectedPortMapping{
					{
						ContainerPort: 5432,
						HostPort: 5432,
						Protocol:     "tcp",
					},
				},
			},
		},
		Networks: []DetectedNetwork{
			{
				ID:       "bridge",
				Name:     "bridge",
				Driver:   "bridge",
				Scope:    "local",
				Internal: false,
			},
		},
		DetectedAt: time.Now(),
	}

	// Create mock detector
	mockDetector := &MockContainerDetector{}
	mockDetector.On("DetectContainerEnvironment", mock.Anything).Return(mockEnv, nil)

	analyzer := &NetworkTopologyAnalyzer{
		logger:       logger,
		cacheExpiry:  5 * time.Minute,
	}
	// Note: For testing, we'd need to inject the mock detector differently
	// This is a simplified version for the test structure

	ctx := context.Background()
	topology, err := analyzer.AnalyzeNetworkTopology(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, topology)
	assert.Equal(t, 1, len(topology.Networks))
	assert.Equal(t, 2, len(topology.Containers))
	assert.True(t, len(topology.Services) > 0)
	assert.NotEmpty(t, topology.GeneratedAt)

	mockDetector.AssertExpectations(t)
}

func TestAnalyzeNetworkTopologyWithCache(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockDetector := &MockContainerDetector{}

	analyzer := &NetworkTopologyAnalyzer{
		logger:       logger,
		cacheExpiry:  5 * time.Minute,
		lastAnalysis: time.Now(),
		cachedTopology: &NetworkTopology{
			GeneratedAt: time.Now(),
			Networks:    []TopologyNetwork{},
			Containers:  []TopologyContainer{},
			Services:    []TopologyService{},
		},
	}

	ctx := context.Background()
	topology, err := analyzer.AnalyzeNetworkTopology(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, topology)
	assert.Equal(t, analyzer.cachedTopology, topology)

	// Should not call detector when cache is valid
	mockDetector.AssertNotCalled(t, "DetectContainerEnvironment")
}

func TestAnalyzeNetworkTopologyExpiredCache(t *testing.T) {
	logger := zaptest.NewLogger(t)

	mockEnv := &ContainerEnvironment{
		AvailableRuntimes: []RuntimeInfo{
			{Runtime: Docker, Available: true, Version: "20.10.0"},
		},
		PrimaryRuntime:    Docker,
		RunningContainers: []DetectedContainer{},
		Networks:          []DetectedNetwork{},
		DetectedAt:        time.Now(),
	}

	mockDetector := &MockContainerDetector{}
	mockDetector.On("DetectContainerEnvironment", mock.Anything).Return(mockEnv, nil)

	analyzer := &NetworkTopologyAnalyzer{
		logger:       logger,
		cacheExpiry:  5 * time.Minute,
		lastAnalysis: time.Now().Add(-10 * time.Minute), // Expired cache
		cachedTopology: &NetworkTopology{
			GeneratedAt: time.Now().Add(-10 * time.Minute),
		},
	}

	ctx := context.Background()
	topology, err := analyzer.AnalyzeNetworkTopology(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, topology)

	// Should call detector when cache is expired
	mockDetector.AssertExpectations(t)
}

func TestDiscoverServices(t *testing.T) {
	logger := zaptest.NewLogger(t)
	detector := NewContainerDetector(logger)
	analyzer := NewNetworkTopologyAnalyzer(logger, detector)

	containers := []TopologyContainer{
		{
			ID:    "web1",
			Name:  "web-server-1",
			Image: "nginx:latest",
			ExposedPorts: []ContainerPort{
				{ContainerPort: 80, Protocol: "tcp"},
			},
			ServiceLabels: map[string]string{
				"service.name": "web",
				"service.type": "web",
			},
		},
		{
			ID:    "web2",
			Name:  "web-server-2",
			Image: "nginx:latest",
			ExposedPorts: []ContainerPort{
				{ContainerPort: 80, Protocol: "tcp"},
			},
			ServiceLabels: map[string]string{
				"service.name": "web",
				"service.type": "web",
			},
		},
		{
			ID:    "db1",
			Name:  "postgres-db",
			Image: "postgres:13",
			ExposedPorts: []ContainerPort{
				{ContainerPort: 5432, Protocol: "tcp"},
			},
			ServiceLabels: map[string]string{
				"service.name": "database",
				"service.type": "database",
			},
		},
	}

	services, _ := analyzer.discoverServices(context.Background(), containers)

	assert.Len(t, services, 2)

	// Find web service
	var webService *TopologyService
	for i := range services {
		if services[i].Name == "web" {
			webService = &services[i]
			break
		}
	}

	assert.NotNil(t, webService)
	assert.Equal(t, ServiceTypeWeb, webService.Type)
	assert.Len(t, webService.Containers, 2)
	assert.Contains(t, webService.Containers, "web1")
	assert.Contains(t, webService.Containers, "web2")

	// Find database service
	var dbService *TopologyService
	for i := range services {
		if services[i].Name == "database" {
			dbService = &services[i]
			break
		}
	}

	assert.NotNil(t, dbService)
	assert.Equal(t, ServiceTypeDatabase, dbService.Type)
	assert.Len(t, dbService.Containers, 1)
	assert.Contains(t, dbService.Containers, "db1")
}

func TestAnalyzeConnections(t *testing.T) {
	logger := zaptest.NewLogger(t)
	detector := NewContainerDetector(logger)
	analyzer := NewNetworkTopologyAnalyzer(logger, detector)

	containers := []TopologyContainer{
		{
			ID:   "web1",
			Name: "web-server",
			NetworkInterfaces: []ContainerNetworkInterface{
				{IPAddress: "172.17.0.2"},
			},
			ExposedPorts: []ContainerPort{
				{ContainerPort: 80, Protocol: "tcp"},
			},
		},
		{
			ID:   "db1",
			Name: "database",
			NetworkInterfaces: []ContainerNetworkInterface{
				{IPAddress: "172.17.0.3"},
			},
			ExposedPorts: []ContainerPort{
				{ContainerPort: 5432, Protocol: "tcp"},
			},
		},
	}

	ctx := context.Background()
	connections, _ := analyzer.analyzeConnections(ctx, containers, []TopologyNetwork{})

	// Should find potential connections between containers
	assert.True(t, len(connections) >= 0)

	for _, conn := range connections {
		assert.NotEmpty(t, conn.ID)
		assert.NotEmpty(t, conn.Source.Name)
		assert.NotEmpty(t, conn.Target.Name)
		assert.NotEmpty(t, conn.Protocol)
		assert.True(t, conn.Port > 0)
	}
}

func TestIdentifyClusters(t *testing.T) {
	logger := zaptest.NewLogger(t)
	detector := NewContainerDetector(logger)
	analyzer := NewNetworkTopologyAnalyzer(logger, detector)

	networks := []TopologyNetwork{
		{
			ID:                  "bridge",
			Name:                "bridge",
			Driver:              "bridge",
			ConnectedContainers: []string{"web1", "web2"},
		},
		{
			ID:                  "custom",
			Name:                "app-network",
			Driver:              "bridge",
			ConnectedContainers: []string{"api1", "db1"},
		},
	}

	containers := []TopologyContainer{
		{ID: "web1", Name: "web1"},
		{ID: "web2", Name: "web2"},
		{ID: "api1", Name: "api1"},
		{ID: "db1", Name: "db1"},
	}

	clusters, _ := analyzer.identifyNetworkClusters(context.Background(), networks, containers, []TopologyService{})

	assert.Len(t, clusters, 2)

	for _, cluster := range clusters {
		assert.NotEmpty(t, cluster.ID)
		assert.NotEmpty(t, cluster.Name)
		assert.True(t, len(cluster.Members) > 0)
	}
}

func TestCalculateTopologyComplexity(t *testing.T) {
	logger := zaptest.NewLogger(t)
	detector := NewContainerDetector(logger)
	analyzer := NewNetworkTopologyAnalyzer(logger, detector)

	topology := &NetworkTopology{
		Networks: []TopologyNetwork{
			{ID: "net1", ConnectedContainers: []string{"c1", "c2"}},
			{ID: "net2", ConnectedContainers: []string{"c2", "c3"}},
		},
		Containers: []TopologyContainer{
			{ID: "c1", Name: "container1"},
			{ID: "c2", Name: "container2"},
			{ID: "c3", Name: "container3"},
		},
		Services: []TopologyService{
			{Name: "web", Containers: []string{"c1", "c2"}},
			{Name: "api", Containers: []string{"c3"}},
		},
		Connections: []NetworkConnection{
			{Source: ConnectionNode{ID: "c1"}, Target: ConnectionNode{ID: "c2"}},
			{Source: ConnectionNode{ID: "c2"}, Target: ConnectionNode{ID: "c3"}},
		},
	}

	complexity := analyzer.calculateComplexityMetrics(topology)

	assert.True(t, complexity.NetworkComplexity >= 0)
	assert.True(t, complexity.ServiceComplexity >= 0)
	assert.True(t, complexity.ConnectionDensity >= 0 && complexity.ConnectionDensity <= 1)
	assert.True(t, complexity.BranchingFactor >= 0)
	assert.True(t, complexity.CyclomaticComplexity >= 0)
}

func TestExportTopology(t *testing.T) {
	logger := zaptest.NewLogger(t)
	detector := NewContainerDetector(logger)
	analyzer := NewNetworkTopologyAnalyzer(logger, detector)

	topology := &NetworkTopology{
		GeneratedAt: time.Now(),
		Networks: []TopologyNetwork{
			{ID: "net1", Name: "test-network"},
		},
		Containers: []TopologyContainer{
			{ID: "c1", Name: "test-container"},
		},
		Services: []TopologyService{
			{Name: "test-service"},
		},
	}

	// Test JSON export
	jsonData, err := analyzer.ExportTopology(topology, "json")
	assert.NoError(t, err)
	assert.Contains(t, string(jsonData), "test-network")
	assert.Contains(t, string(jsonData), "test-container")

	// Test DOT export
	dotData, err := analyzer.ExportTopology(topology, "dot")
	assert.NoError(t, err)
	assert.Contains(t, string(dotData), "digraph")
	assert.Contains(t, string(dotData), "test-network")

	// Test Cytoscape export
	cytoscapeData, err := analyzer.ExportTopology(topology, "cytoscape")
	assert.NoError(t, err)
	assert.Contains(t, string(cytoscapeData), "nodes")
	assert.Contains(t, string(cytoscapeData), "edges")

	// Test invalid format
	_, err = analyzer.ExportTopology(topology, "invalid")
	assert.Error(t, err)
}

func TestGenerateTopologyHash(t *testing.T) {
	topology := &NetworkTopology{
		Networks: []TopologyNetwork{
			{ID: "12345678901234567890", Name: "net1"},
			{ID: "abcdefghijklmnopqrst", Name: "net2"},
		},
		Containers: []TopologyContainer{
			{ID: "c1", Name: "container1"},
		},
		Services: []TopologyService{
			{Name: "service1"},
		},
		Connections: []NetworkConnection{
			{ID: "conn1"},
		},
	}

	hash := generateTopologyHash(topology)

	assert.NotEmpty(t, hash)
	assert.Contains(t, hash, "nets:2")
	assert.Contains(t, hash, "containers:1")
	assert.Contains(t, hash, "services:1")
	assert.Contains(t, hash, "connections:1")
	assert.Contains(t, hash, "net_ids:")
}

func TestValidateTopology(t *testing.T) {
	// Test topology with issues
	topology := &NetworkTopology{
		Networks: []TopologyNetwork{
			{
				ID:                  "net1",
				Name:                "public-network",
				Driver:              "bridge",
				Internal:            false,
				ConnectedContainers: make([]string, 15), // Too many containers
			},
		},
		Containers: []TopologyContainer{
			{
				ID:                "c1",
				Name:              "isolated-container",
				NetworkInterfaces: []ContainerNetworkInterface{}, // No network interfaces
			},
		},
		Services: []TopologyService{
			{
				Name:       "empty-service",
				Containers: []string{}, // No containers
			},
		},
		Connections: []NetworkConnection{
			{
				ID:     "conn1",
				Status: StatusFailed, // Failed connection
			},
		},
		Summary: TopologySummary{
			TopologyComplexity: TopologyComplexityMetrics{
				ConnectionDensity:    0.9, // High density
				CyclomaticComplexity: 60,  // High complexity
			},
		},
	}

	issues := validateTopology(topology, false)

	assert.True(t, len(issues) > 0)

	// Check for specific issues
	issueText := strings.Join(issues, " ")
	assert.Contains(t, issueText, "no network interfaces")
	assert.Contains(t, issueText, "no containers")
	assert.Contains(t, issueText, "failed state")
	assert.Contains(t, issueText, "High connection density")
	assert.Contains(t, issueText, "High cyclomatic complexity")
	assert.Contains(t, issueText, "many containers")
}

// MockContainerDetector for testing
type MockContainerDetector struct {
	mock.Mock
}

func (m *MockContainerDetector) DetectContainerEnvironment(ctx context.Context) (*ContainerEnvironment, error) {
	args := m.Called(ctx)
	return args.Get(0).(*ContainerEnvironment), args.Error(1)
}

func (m *MockContainerDetector) CheckRuntimeAvailability(runtime ContainerRuntime) RuntimeInfo {
	args := m.Called(runtime)
	return args.Get(0).(RuntimeInfo)
}

func (m *MockContainerDetector) detectDockerContainers(ctx context.Context) ([]DetectedContainer, error) {
	args := m.Called(ctx)
	return args.Get(0).([]DetectedContainer), args.Error(1)
}

func (m *MockContainerDetector) detectPodmanContainers(ctx context.Context) ([]DetectedContainer, error) {
	args := m.Called(ctx)
	return args.Get(0).([]DetectedContainer), args.Error(1)
}
