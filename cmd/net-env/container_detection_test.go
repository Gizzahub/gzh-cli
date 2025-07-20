//nolint:testpackage // White-box testing needed for internal function access
package netenv

import (
	"context"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
	
	"github.com/gizzahub/gzh-manager-go/internal/env"
)

func TestContainerDetector_DetectContainerEnvironment(t *testing.T) {
	logger := zap.NewNop()
	cd := NewContainerDetector(logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	env, err := cd.DetectContainerEnvironment(ctx)
	if err != nil {
		t.Fatalf("DetectContainerEnvironment failed: %v", err)
	}

	// Basic validations
	if env == nil {
		t.Fatal("Environment should not be nil")
	}

	if env.DetectedAt.IsZero() {
		t.Error("DetectedAt should be set")
	}

	if env.EnvironmentFingerprint == "" {
		t.Error("EnvironmentFingerprint should not be empty")
	}

	// At least one runtime should be detected (even if unavailable)
	if len(env.AvailableRuntimes) == 0 {
		t.Error("Should detect at least one runtime")
	}

	// Primary runtime should be set
	if env.PrimaryRuntime == "" {
		t.Error("PrimaryRuntime should be set")
	}

	// Orchestration platform should be detected
	validPlatforms := []string{"standalone", "docker-swarm", "kubernetes", "unknown"}
	found := false

	for _, platform := range validPlatforms {
		if env.OrchestrationPlatform == platform {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Invalid orchestration platform: %s", env.OrchestrationPlatform)
	}

	t.Logf("Detected environment:")
	t.Logf("  Primary Runtime: %s", env.PrimaryRuntime)
	t.Logf("  Orchestration: %s", env.OrchestrationPlatform)
	t.Logf("  Available Runtimes: %d", len(env.AvailableRuntimes))
	t.Logf("  Running Containers: %d", len(env.RunningContainers))
	t.Logf("  Networks: %d", len(env.Networks))
	t.Logf("  Compose Projects: %d", len(env.ComposeProjects))
}

func TestContainerDetector_DetectAvailableRuntimes(t *testing.T) {
	logger := zap.NewNop()
	cd := NewContainerDetector(logger)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	runtimes, err := cd.DetectAvailableRuntimes(ctx)
	if err != nil {
		t.Fatalf("DetectAvailableRuntimes failed: %v", err)
	}

	// Should detect common runtimes
	expectedRuntimes := []ContainerRuntime{Docker, Podman, Containerd, Nerdctl}
	if len(runtimes) != len(expectedRuntimes) {
		t.Errorf("Expected %d runtimes, got %d", len(expectedRuntimes), len(runtimes))
	}

	// Validate runtime information
	for _, runtime := range runtimes {
		if runtime.Runtime == "" {
			t.Error("Runtime name should not be empty")
		}

		if runtime.Executable == "" {
			t.Error("Runtime executable should not be empty")
		}

		// Available runtimes should have version information
		if runtime.Available && runtime.Version == "" {
			t.Errorf("Available runtime %s should have version information", runtime.Runtime)
		}

		t.Logf("Runtime %s: available=%t, version=%s, executable=%s",
			runtime.Runtime, runtime.Available, runtime.Version, runtime.Executable)
	}
}

func TestContainerDetector_DeterminePrimaryRuntime(t *testing.T) {
	logger := zap.NewNop()
	cd := NewContainerDetector(logger)

	// Test with available runtimes
	runtimes := []RuntimeInfo{
		{Runtime: Docker, Available: true, Version: "20.10.0"},
		{Runtime: Podman, Available: false},
		{Runtime: Containerd, Available: true, Version: "1.6.0"},
		{Runtime: Nerdctl, Available: false},
	}

	primary := cd.DeterminePrimaryRuntime(runtimes)

	// Docker should be primary if available
	if primary != Docker {
		t.Errorf("Expected Docker to be primary runtime, got %s", primary)
	}

	// Test with no available runtimes
	noRuntimes := []RuntimeInfo{
		{Runtime: Docker, Available: false},
		{Runtime: Podman, Available: false},
		{Runtime: Containerd, Available: false},
		{Runtime: Nerdctl, Available: false},
	}

	primary = cd.DeterminePrimaryRuntime(noRuntimes)
	if primary == "" {
		t.Error("Should return a runtime even if none are available")
	}

	// Test with only Podman available
	podmanOnly := []RuntimeInfo{
		{Runtime: Docker, Available: false},
		{Runtime: Podman, Available: true, Version: "3.4.0"},
		{Runtime: Containerd, Available: false},
		{Runtime: Nerdctl, Available: false},
	}

	primary = cd.DeterminePrimaryRuntime(podmanOnly)
	if primary != Podman {
		t.Errorf("Expected Podman to be primary runtime, got %s", primary)
	}
}

func TestContainerDetector_CalculateEnvironmentFingerprint(t *testing.T) {
	logger := zap.NewNop()
	cd := NewContainerDetector(logger)

	env := &ContainerEnvironment{
		PrimaryRuntime:        Docker,
		OrchestrationPlatform: "standalone",
		RunningContainers: []DetectedContainer{
			{ID: "container1", Name: "test1", Image: "nginx:latest"},
			{ID: "container2", Name: "test2", Image: "redis:alpine"},
		},
		Networks: []DetectedNetwork{
			{ID: "net1", Name: "bridge"},
			{ID: "net2", Name: "custom"},
		},
	}

	fingerprint := cd.CalculateEnvironmentFingerprint(env)

	if fingerprint == "" {
		t.Error("Fingerprint should not be empty")
	}

	// Fingerprint should be consistent
	fingerprint2 := cd.CalculateEnvironmentFingerprint(env)
	if fingerprint != fingerprint2 {
		t.Error("Fingerprint should be consistent for same environment")
	}

	// Fingerprint should change when environment changes
	env.RunningContainers = append(env.RunningContainers, DetectedContainer{
		ID: "container3", Name: "test3", Image: "postgres:13",
	})

	fingerprint3 := cd.CalculateEnvironmentFingerprint(env)
	if fingerprint == fingerprint3 {
		t.Error("Fingerprint should change when environment changes")
	}

	t.Logf("Original fingerprint: %s", fingerprint)
	t.Logf("Changed fingerprint: %s", fingerprint3)
}

func TestContainerDetector_DetectOrchestrationPlatform(t *testing.T) {
	logger := zap.NewNop()
	cd := NewContainerDetector(logger)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	platform, err := cd.DetectOrchestrationPlatform(ctx)
	if err != nil {
		t.Fatalf("DetectOrchestrationPlatform failed: %v", err)
	}

	validPlatforms := []string{"standalone", "docker-swarm", "kubernetes", "unknown"}
	found := false

	for _, validPlatform := range validPlatforms {
		if platform == validPlatform {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Invalid orchestration platform: %s", platform)
	}

	t.Logf("Detected orchestration platform: %s", platform)
}

func TestContainerDetector_ParseDockerPsOutput(t *testing.T) {
	logger := zap.NewNop()
	cd := NewContainerDetector(logger)

	sampleOutput := `CONTAINER ID,IMAGE,COMMAND,CREATED,STATUS,PORTS,NAMES
abc123456789,nginx:latest,nginx -g daemon off;,2 hours ago,Up 2 hours,0.0.0.0:80->80/tcp,webserver
def987654321,redis:alpine,redis-server,1 hour ago,Up 1 hour,6379/tcp,cache`

	containers, err := cd.ParseDockerPsOutput(sampleOutput)
	if err != nil {
		t.Fatalf("ParseDockerPsOutput failed: %v", err)
	}

	if len(containers) != 2 {
		t.Errorf("Expected 2 containers, got %d", len(containers))
	}

	// Validate first container
	if containers[0].ID != env.TestContainerID {
		t.Errorf("Expected container ID 'abc123456789', got '%s'", containers[0].ID)
	}

	if containers[0].Name != "webserver" {
		t.Errorf("Expected container name 'webserver', got '%s'", containers[0].Name)
	}

	if containers[0].Image != "nginx:latest" {
		t.Errorf("Expected image 'nginx:latest', got '%s'", containers[0].Image)
	}

	if containers[0].Status != "Up 2 hours" {
		t.Errorf("Expected status 'Up 2 hours', got '%s'", containers[0].Status)
	}

	// Validate second container
	if containers[1].ID != "def987654321" {
		t.Errorf("Expected container ID 'def987654321', got '%s'", containers[1].ID)
	}

	if containers[1].Name != "cache" {
		t.Errorf("Expected container name 'cache', got '%s'", containers[1].Name)
	}
}

func TestContainerDetector_ParseDockerNetworkOutput(t *testing.T) {
	logger := zap.NewNop()
	cd := NewContainerDetector(logger)

	sampleOutput := `NETWORK ID,NAME,DRIVER,SCOPE
abc123456789,bridge,bridge,local
def987654321,host,host,local
ghi123456789,mynetwork,bridge,local`

	networks, err := cd.ParseDockerNetworkOutput(sampleOutput)
	if err != nil {
		t.Fatalf("ParseDockerNetworkOutput failed: %v", err)
	}

	if len(networks) != 3 {
		t.Errorf("Expected 3 networks, got %d", len(networks))
	}

	// Validate first network
	if networks[0].ID != env.TestContainerID {
		t.Errorf("Expected network ID 'abc123456789', got '%s'", networks[0].ID)
	}

	if networks[0].Name != "bridge" {
		t.Errorf("Expected network name 'bridge', got '%s'", networks[0].Name)
	}

	if networks[0].Driver != "bridge" {
		t.Errorf("Expected driver 'bridge', got '%s'", networks[0].Driver)
	}

	if networks[0].Scope != "local" {
		t.Errorf("Expected scope 'local', got '%s'", networks[0].Scope)
	}
}

func TestRuntimeInfo_Validation(t *testing.T) {
	// Test valid runtime
	runtime := RuntimeInfo{
		Runtime:    Docker,
		Version:    "20.10.0",
		Available:  true,
		Executable: "/usr/bin/docker",
	}

	if runtime.Runtime != Docker {
		t.Error("Runtime should be Docker")
	}

	if !runtime.Available {
		t.Error("Runtime should be available")
	}

	// Test runtime with server info
	runtime.ServerInfo = &ServerInfo{
		Version:       "20.10.0",
		OS:            "linux",
		Architecture:  "amd64",
		KernelVersion: "5.4.0",
		TotalMemory:   "8GB",
		CPUs:          4,
		StorageDriver: "overlay2",
		LoggingDriver: "json-file",
		CgroupDriver:  "cgroupfs",
		RuntimeConfig: map[string]string{
			"runc": "1.0.0",
		},
	}

	if runtime.ServerInfo.OS != "linux" {
		t.Error("Server OS should be linux")
	}

	if runtime.ServerInfo.CPUs != 4 {
		t.Error("Server should have 4 CPUs")
	}
}

func TestDetectedContainer_Validation(t *testing.T) {
	container := DetectedContainer{
		ID:        env.TestContainerID,
		Name:      "test-container",
		Image:     "nginx:latest",
		ImageID:   "sha256:def987654321",
		Status:    "Up 2 hours",
		State:     "running",
		Runtime:   Docker,
		Created:   time.Now().Add(-2 * time.Hour),
		StartedAt: time.Now().Add(-2 * time.Hour),
		Ports: []DetectedPortMapping{
			{ContainerPort: 80, HostPort: 8080, HostIP: "0.0.0.0", Protocol: "tcp"},
		},
		Networks: []DetectedNetworkInfo{
			{NetworkName: "bridge", NetworkID: "net123", IPAddress: "172.17.0.2"},
		},
		Labels: map[string]string{
			"app": "nginx",
		},
		Environment: []string{
			"PATH=/usr/local/sbin:/usr/local/bin",
			"NGINX_VERSION=1.21.0",
		},
		HealthStatus:  "healthy",
		RestartPolicy: "always",
		WorkingDir:    "/usr/share/nginx/html",
		Command:       []string{"nginx", "-g", "daemon off;"},
	}

	if container.ID != env.TestContainerID {
		t.Error("Container ID mismatch")
	}

	if container.Runtime != Docker {
		t.Error("Container runtime should be Docker")
	}

	if len(container.Ports) != 1 {
		t.Error("Container should have 1 port mapping")
	}

	if container.Ports[0].ContainerPort != 80 {
		t.Error("Container port should be 80")
	}

	if len(container.Networks) != 1 {
		t.Error("Container should be connected to 1 network")
	}

	if container.Networks[0].IPAddress != "172.17.0.2" {
		t.Error("Container IP should be 172.17.0.2")
	}

	if len(container.Labels) != 1 {
		t.Error("Container should have 1 label")
	}

	if container.Labels["app"] != "nginx" {
		t.Error("Container label 'app' should be 'nginx'")
	}
}

// Benchmark tests for performance.
func BenchmarkContainerDetector_DetectAvailableRuntimes(b *testing.B) {
	logger := zap.NewNop()
	cd := NewContainerDetector(logger)
	ctx := context.Background()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := cd.DetectAvailableRuntimes(ctx)
		if err != nil {
			b.Fatalf("DetectAvailableRuntimes failed: %v", err)
		}
	}
}

func BenchmarkContainerDetector_CalculateEnvironmentFingerprint(b *testing.B) {
	logger := zap.NewNop()
	cd := NewContainerDetector(logger)

	env := &ContainerEnvironment{
		PrimaryRuntime:        Docker,
		OrchestrationPlatform: "standalone",
		RunningContainers: []DetectedContainer{
			{ID: "container1", Name: "test1", Image: "nginx:latest"},
			{ID: "container2", Name: "test2", Image: "redis:alpine"},
		},
		Networks: []DetectedNetwork{
			{ID: "net1", Name: "bridge"},
			{ID: "net2", Name: "custom"},
		},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = cd.CalculateEnvironmentFingerprint(env)
	}
}

// Integration tests that require Docker to be running.
func TestContainerDetector_Integration_DockerRunning(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := zap.NewNop()
	cd := NewContainerDetector(logger)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if Docker is available
	runtimes, err := cd.DetectAvailableRuntimes(ctx)
	if err != nil {
		t.Skipf("Cannot detect runtimes: %v", err)
	}

	dockerAvailable := false

	for _, runtime := range runtimes {
		if runtime.Runtime == Docker && runtime.Available {
			dockerAvailable = true
			break
		}
	}

	if !dockerAvailable {
		t.Skip("Docker is not available, skipping integration test")
	}

	// Test getting running containers
	containers, err := cd.GetRunningContainers(ctx, Docker)
	if err != nil {
		t.Fatalf("GetRunningContainers failed: %v", err)
	}

	t.Logf("Found %d running Docker containers", len(containers))

	// Test getting networks
	networks, err := cd.GetContainerNetworks(ctx, Docker)
	if err != nil {
		t.Fatalf("GetContainerNetworks failed: %v", err)
	}

	t.Logf("Found %d Docker networks", len(networks))

	// At least the default networks should be present
	foundBridge := false

	for _, network := range networks {
		if network.Name == "bridge" {
			foundBridge = true
			break
		}
	}

	if !foundBridge {
		t.Error("Should find the default 'bridge' network")
	}
}

// Test error conditions.
func TestContainerDetector_ErrorConditions(t *testing.T) {
	logger := zap.NewNop()
	cd := NewContainerDetector(logger)

	// Test with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := cd.DetectContainerEnvironment(ctx)
	if err == nil {
		t.Error("Should return error for cancelled context")
	}

	if !strings.Contains(err.Error(), "context") {
		t.Errorf("Error should mention context cancellation, got: %v", err)
	}
}
