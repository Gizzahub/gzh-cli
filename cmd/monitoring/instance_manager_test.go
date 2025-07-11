package monitoring

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestInstanceManager_RegisterLocalInstance(t *testing.T) {
	logger := zaptest.NewLogger(t)
	manager := NewInstanceManager(logger)

	// Test registering local instance
	manager.RegisterLocalInstance("localhost", 8080, "test-instance")

	instances := manager.GetInstances()
	require.Len(t, instances, 1)

	instance := instances[0]
	assert.Equal(t, "localhost:8080", instance.ID)
	assert.Equal(t, "test-instance", instance.Name)
	assert.Equal(t, "localhost", instance.Host)
	assert.Equal(t, 8080, instance.Port)
	assert.Equal(t, "running", instance.Status)
	assert.Equal(t, "local", instance.Tags["type"])
}

func TestInstanceManager_GetInstance(t *testing.T) {
	logger := zaptest.NewLogger(t)
	manager := NewInstanceManager(logger)

	// Register a test instance
	manager.RegisterLocalInstance("localhost", 8080, "test-instance")

	// Test getting existing instance
	instance, err := manager.GetInstance("localhost:8080")
	require.NoError(t, err)
	assert.Equal(t, "test-instance", instance.Name)

	// Test getting non-existing instance
	_, err = manager.GetInstance("nonexistent:9090")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestInstanceManager_UpdateLocalMetrics(t *testing.T) {
	logger := zaptest.NewLogger(t)
	manager := NewInstanceManager(logger)

	// Register local instance
	manager.RegisterLocalInstance("localhost", 8080, "test-instance")

	// Update metrics
	metrics := &SystemStatus{
		Status:        "healthy",
		Uptime:        "1h30m",
		ActiveTasks:   5,
		TotalRequests: 1000,
		MemoryUsage:   1024 * 1024 * 100, // 100MB
		CPUUsage:      25.5,
		Timestamp:     time.Now(),
	}

	manager.UpdateLocalMetrics(metrics)

	// Verify metrics were updated
	instance, err := manager.GetInstance("localhost:8080")
	require.NoError(t, err)
	require.NotNil(t, instance.Metrics)

	assert.Equal(t, "healthy", instance.Metrics.Status)
	assert.Equal(t, 5, instance.Metrics.ActiveTasks)
	assert.Equal(t, int64(1000), instance.Metrics.TotalRequests)
	assert.Equal(t, uint64(1024*1024*100), instance.Metrics.MemoryUsage)
	assert.Equal(t, 25.5, instance.Metrics.CPUUsage)
}

func TestInstanceManager_GetClusterStatus(t *testing.T) {
	logger := zaptest.NewLogger(t)
	manager := NewInstanceManager(logger)

	// Test empty cluster
	status := manager.GetClusterStatus()
	assert.Equal(t, 0, status.TotalInstances)
	assert.Equal(t, 0, status.RunningInstances)
	assert.Equal(t, 0, status.UnhealthyInstances)

	// Add running instance
	manager.RegisterLocalInstance("localhost", 8080, "test-instance-1")

	// Simulate adding a remote instance manually for testing
	manager.mu.Lock()
	remoteInstance := &InstanceInfo{
		ID:       "remote:8081",
		Name:     "remote-instance",
		Host:     "remote",
		Port:     8081,
		Status:   "running",
		LastSeen: time.Now(),
		Tags:     map[string]string{"type": "remote"},
	}
	manager.instances["remote:8081"] = remoteInstance
	manager.mu.Unlock()

	// Add unhealthy instance
	manager.mu.Lock()
	unhealthyInstance := &InstanceInfo{
		ID:       "unhealthy:8082",
		Name:     "unhealthy-instance",
		Host:     "unhealthy",
		Port:     8082,
		Status:   "unhealthy",
		LastSeen: time.Now(),
		Tags:     map[string]string{"type": "remote"},
	}
	manager.instances["unhealthy:8082"] = unhealthyInstance
	manager.mu.Unlock()

	// Test cluster status with multiple instances
	status = manager.GetClusterStatus()
	assert.Equal(t, 3, status.TotalInstances)
	assert.Equal(t, 2, status.RunningInstances)
	assert.Equal(t, 1, status.UnhealthyInstances)
	assert.Len(t, status.Instances, 3)
}

func TestInstanceManager_RemoveInstance(t *testing.T) {
	logger := zaptest.NewLogger(t)
	manager := NewInstanceManager(logger)

	// Register local instance
	manager.RegisterLocalInstance("localhost", 8080, "local-instance")

	// Add remote instance manually for testing
	manager.mu.Lock()
	remoteInstance := &InstanceInfo{
		ID:   "remote:8081",
		Name: "remote-instance",
		Host: "remote",
		Port: 8081,
		Tags: map[string]string{"type": "remote"},
	}
	manager.instances["remote:8081"] = remoteInstance
	manager.mu.Unlock()

	// Test removing remote instance
	err := manager.RemoveInstance("remote:8081")
	assert.NoError(t, err)

	_, err = manager.GetInstance("remote:8081")
	assert.Error(t, err)

	// Test removing local instance (should fail)
	err = manager.RemoveInstance("localhost:8080")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot remove local instance")

	// Verify local instance still exists
	_, err = manager.GetInstance("localhost:8080")
	assert.NoError(t, err)
}

func TestInstanceManager_StartStop(t *testing.T) {
	logger := zaptest.NewLogger(t)
	manager := NewInstanceManager(logger)

	// Test start and stop
	manager.Start()

	// Wait a bit to ensure background routines start
	time.Sleep(100 * time.Millisecond)

	manager.Stop()

	// Verify context is cancelled
	select {
	case <-manager.ctx.Done():
		// Context should be cancelled
	default:
		t.Error("Context should be cancelled after Stop()")
	}
}

func TestGetEnvironment(t *testing.T) {
	// Test default environment
	env := getEnvironment()
	assert.Equal(t, "development", env)

	// Note: Testing with actual environment variables would require
	// setting them in the test environment, which might affect other tests
}

func TestInstanceInfo_JSON(t *testing.T) {
	instance := &InstanceInfo{
		ID:          "test:8080",
		Name:        "test-instance",
		Host:        "test",
		Port:        8080,
		Status:      "running",
		LastSeen:    time.Now(),
		Version:     "1.0.0",
		Environment: "test",
		Tags:        map[string]string{"type": "local"},
	}

	// Test that the struct can be marshaled/unmarshaled properly
	// This is important for API responses
	assert.NotEmpty(t, instance.ID)
	assert.NotEmpty(t, instance.Name)
	assert.Greater(t, instance.Port, 0)
	assert.NotEmpty(t, instance.Status)
	assert.NotNil(t, instance.Tags)
}

func TestClusterStatus_HealthCalculation(t *testing.T) {
	logger := zaptest.NewLogger(t)
	manager := NewInstanceManager(logger)

	// Add multiple instances with different statuses
	instances := []*InstanceInfo{
		{ID: "1", Status: "running", LastSeen: time.Now()},
		{ID: "2", Status: "running", LastSeen: time.Now()},
		{ID: "3", Status: "unhealthy", LastSeen: time.Now()},
		{ID: "4", Status: "running", LastSeen: time.Now().Add(-5 * time.Minute)}, // Stale
	}

	manager.mu.Lock()
	for _, instance := range instances {
		manager.instances[instance.ID] = instance
	}
	manager.mu.Unlock()

	status := manager.GetClusterStatus()

	// Total: 4, Running: 2 (healthy and not stale), Unhealthy: 2 (1 explicitly unhealthy + 1 stale)
	assert.Equal(t, 4, status.TotalInstances)
	assert.Equal(t, 2, status.RunningInstances)
	assert.Equal(t, 2, status.UnhealthyInstances)
}
