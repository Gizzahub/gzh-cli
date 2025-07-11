package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
)

// InstanceInfo represents information about a monitoring instance
type InstanceInfo struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Host        string            `json:"host"`
	Port        int               `json:"port"`
	Status      string            `json:"status"`
	LastSeen    time.Time         `json:"last_seen"`
	Version     string            `json:"version"`
	Metrics     *SystemStatus     `json:"metrics,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`
	Environment string            `json:"environment"`
}

// InstanceManager manages multiple monitoring instances
type InstanceManager struct {
	mu            sync.RWMutex
	instances     map[string]*InstanceInfo
	localInstance *InstanceInfo
	logger        *zap.Logger
	syncInterval  time.Duration
	ctx           context.Context
	cancel        context.CancelFunc
}

// NewInstanceManager creates a new instance manager
func NewInstanceManager(logger *zap.Logger) *InstanceManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &InstanceManager{
		instances:    make(map[string]*InstanceInfo),
		logger:       logger,
		syncInterval: 30 * time.Second,
		ctx:          ctx,
		cancel:       cancel,
	}
}

// RegisterLocalInstance registers the local monitoring instance
func (im *InstanceManager) RegisterLocalInstance(host string, port int, name string) {
	im.mu.Lock()
	defer im.mu.Unlock()

	instanceID := fmt.Sprintf("%s:%d", host, port)

	im.localInstance = &InstanceInfo{
		ID:          instanceID,
		Name:        name,
		Host:        host,
		Port:        port,
		Status:      "running",
		LastSeen:    time.Now(),
		Version:     "1.0.0", // This should come from build info
		Environment: getEnvironment(),
		Tags: map[string]string{
			"type": "local",
		},
	}

	im.instances[instanceID] = im.localInstance
	im.logger.Info("Local instance registered",
		zap.String("id", instanceID),
		zap.String("name", name))
}

// DiscoverInstance adds a remote instance for monitoring
func (im *InstanceManager) DiscoverInstance(host string, port int) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	instanceID := fmt.Sprintf("%s:%d", host, port)

	// Check if instance already exists
	if _, exists := im.instances[instanceID]; exists {
		return nil
	}

	// Try to connect and get instance info
	info, err := im.fetchInstanceInfo(host, port)
	if err != nil {
		return fmt.Errorf("failed to discover instance %s: %w", instanceID, err)
	}

	info.Tags = map[string]string{
		"type": "remote",
	}

	im.instances[instanceID] = info
	im.logger.Info("Remote instance discovered",
		zap.String("id", instanceID),
		zap.String("name", info.Name))

	return nil
}

// GetInstances returns all registered instances
func (im *InstanceManager) GetInstances() []*InstanceInfo {
	im.mu.RLock()
	defer im.mu.RUnlock()

	instances := make([]*InstanceInfo, 0, len(im.instances))
	for _, instance := range im.instances {
		instances = append(instances, instance)
	}

	return instances
}

// GetInstance returns a specific instance by ID
func (im *InstanceManager) GetInstance(id string) (*InstanceInfo, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	instance, exists := im.instances[id]
	if !exists {
		return nil, fmt.Errorf("instance %s not found", id)
	}

	return instance, nil
}

// UpdateLocalMetrics updates the local instance metrics
func (im *InstanceManager) UpdateLocalMetrics(metrics *SystemStatus) {
	im.mu.Lock()
	defer im.mu.Unlock()

	if im.localInstance != nil {
		im.localInstance.Metrics = metrics
		im.localInstance.LastSeen = time.Now()
		im.localInstance.Status = "running"
	}
}

// Start begins the instance management background tasks
func (im *InstanceManager) Start() {
	go im.syncLoop()
	im.logger.Info("Instance manager started")
}

// Stop stops the instance manager
func (im *InstanceManager) Stop() {
	im.cancel()
	im.logger.Info("Instance manager stopped")
}

// RemoveInstance removes an instance from the registry
func (im *InstanceManager) RemoveInstance(id string) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	if im.localInstance != nil && im.localInstance.ID == id {
		return fmt.Errorf("cannot remove local instance")
	}

	delete(im.instances, id)
	im.logger.Info("Instance removed", zap.String("id", id))

	return nil
}

// GetClusterStatus returns overall cluster status
func (im *InstanceManager) GetClusterStatus() *ClusterStatus {
	im.mu.RLock()
	defer im.mu.RUnlock()

	status := &ClusterStatus{
		TotalInstances:     len(im.instances),
		RunningInstances:   0,
		UnhealthyInstances: 0,
		LastUpdate:         time.Now(),
		Instances:          make([]*InstanceInfo, 0, len(im.instances)),
	}

	for _, instance := range im.instances {
		status.Instances = append(status.Instances, instance)

		switch instance.Status {
		case "running":
			status.RunningInstances++
		case "unhealthy", "error":
			status.UnhealthyInstances++
		}

		// Check if instance is stale
		if time.Since(instance.LastSeen) > 2*time.Minute {
			status.UnhealthyInstances++
		}
	}

	return status
}

// syncLoop periodically syncs with remote instances
func (im *InstanceManager) syncLoop() {
	ticker := time.NewTicker(im.syncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-im.ctx.Done():
			return
		case <-ticker.C:
			im.syncRemoteInstances()
		}
	}
}

// syncRemoteInstances updates status of remote instances
func (im *InstanceManager) syncRemoteInstances() {
	im.mu.Lock()
	instancesToSync := make([]*InstanceInfo, 0)
	for _, instance := range im.instances {
		if instance.Tags["type"] == "remote" {
			instancesToSync = append(instancesToSync, instance)
		}
	}
	im.mu.Unlock()

	for _, instance := range instancesToSync {
		go func(inst *InstanceInfo) {
			info, err := im.fetchInstanceInfo(inst.Host, inst.Port)
			if err != nil {
				im.markInstanceUnhealthy(inst.ID, err)
				return
			}

			im.mu.Lock()
			if existing, exists := im.instances[inst.ID]; exists {
				existing.Metrics = info.Metrics
				existing.LastSeen = time.Now()
				existing.Status = "running"
			}
			im.mu.Unlock()
		}(instance)
	}
}

// fetchInstanceInfo retrieves instance information from a remote host
func (im *InstanceManager) fetchInstanceInfo(host string, port int) (*InstanceInfo, error) {
	url := fmt.Sprintf("http://%s:%d/api/v1/status", host, port)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received status code %d", resp.StatusCode)
	}

	var metrics SystemStatus
	if err := json.NewDecoder(resp.Body).Decode(&metrics); err != nil {
		return nil, err
	}

	info := &InstanceInfo{
		ID:       fmt.Sprintf("%s:%d", host, port),
		Host:     host,
		Port:     port,
		Status:   "running",
		LastSeen: time.Now(),
		Metrics:  &metrics,
	}

	return info, nil
}

// markInstanceUnhealthy marks an instance as unhealthy
func (im *InstanceManager) markInstanceUnhealthy(id string, err error) {
	im.mu.Lock()
	defer im.mu.Unlock()

	if instance, exists := im.instances[id]; exists {
		instance.Status = "unhealthy"
		instance.LastSeen = time.Now()
		im.logger.Warn("Instance marked as unhealthy",
			zap.String("id", id),
			zap.Error(err))
	}
}

// ClusterStatus represents the overall status of the monitoring cluster
type ClusterStatus struct {
	TotalInstances     int             `json:"total_instances"`
	RunningInstances   int             `json:"running_instances"`
	UnhealthyInstances int             `json:"unhealthy_instances"`
	LastUpdate         time.Time       `json:"last_update"`
	Instances          []*InstanceInfo `json:"instances"`
}

// getEnvironment determines the current environment
func getEnvironment() string {
	if env := os.Getenv("GZH_ENV"); env != "" {
		return env
	}
	if env := os.Getenv("ENVIRONMENT"); env != "" {
		return env
	}
	return "development"
}
