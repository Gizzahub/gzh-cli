package monitoring

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ServiceDiscovery manages service discovery for Prometheus targets
type ServiceDiscovery struct {
	logger   *zap.Logger
	config   *ServiceDiscoveryConfig
	targets  map[string]*DiscoveryTarget
	mutex    sync.RWMutex
	stopCh   chan struct{}
	watchers []TargetWatcher
}

// DiscoveryTarget represents a discovered target
type DiscoveryTarget struct {
	Address     string                 `json:"address"`
	Port        int                    `json:"port"`
	Labels      map[string]string      `json:"labels"`
	Health      string                 `json:"health"` // "healthy", "unhealthy", "unknown"
	LastSeen    time.Time              `json:"last_seen"`
	Metadata    map[string]interface{} `json:"metadata"`
	ServiceType string                 `json:"service_type"`
}

// TargetWatcher interface for different service discovery mechanisms
type TargetWatcher interface {
	Start(ctx context.Context) error
	Stop() error
	GetTargets() ([]*DiscoveryTarget, error)
	Watch(callback func([]*DiscoveryTarget)) error
}

// StaticTargetWatcher implements static target discovery
type StaticTargetWatcher struct {
	targets []*DiscoveryTarget
	logger  *zap.Logger
}

// KubernetesTargetWatcher implements Kubernetes service discovery
type KubernetesTargetWatcher struct {
	logger    *zap.Logger
	namespace string
	selector  map[string]string
	targets   map[string]*DiscoveryTarget
	mutex     sync.RWMutex
}

// ConsulTargetWatcher implements Consul service discovery
type ConsulTargetWatcher struct {
	logger      *zap.Logger
	consulAddr  string
	serviceName string
	targets     map[string]*DiscoveryTarget
	mutex       sync.RWMutex
}

// DNSTargetWatcher implements DNS-based service discovery
type DNSTargetWatcher struct {
	logger   *zap.Logger
	dnsName  string
	port     int
	interval time.Duration
	targets  map[string]*DiscoveryTarget
	mutex    sync.RWMutex
}

// NewServiceDiscovery creates a new service discovery instance
func NewServiceDiscovery(logger *zap.Logger, config *ServiceDiscoveryConfig) *ServiceDiscovery {
	return &ServiceDiscovery{
		logger:   logger,
		config:   config,
		targets:  make(map[string]*DiscoveryTarget),
		stopCh:   make(chan struct{}),
		watchers: []TargetWatcher{},
	}
}

// Start starts the service discovery
func (sd *ServiceDiscovery) Start(ctx context.Context) error {
	if !sd.config.Enabled {
		sd.logger.Info("Service discovery disabled")
		return nil
	}

	sd.logger.Info("Starting service discovery",
		zap.String("type", sd.config.Type))

	// Initialize appropriate watcher based on type
	watcher, err := sd.createWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}

	sd.watchers = append(sd.watchers, watcher)

	// Start watcher
	if err := watcher.Start(ctx); err != nil {
		return fmt.Errorf("failed to start watcher: %w", err)
	}

	// Setup target watching
	go sd.watchTargets(ctx, watcher)

	// Start periodic refresh
	go sd.periodicRefresh(ctx)

	return nil
}

// Stop stops the service discovery
func (sd *ServiceDiscovery) Stop() error {
	sd.logger.Info("Stopping service discovery")

	close(sd.stopCh)

	for _, watcher := range sd.watchers {
		if err := watcher.Stop(); err != nil {
			sd.logger.Error("Failed to stop watcher", zap.Error(err))
		}
	}

	return nil
}

// createWatcher creates the appropriate target watcher
func (sd *ServiceDiscovery) createWatcher() (TargetWatcher, error) {
	switch sd.config.Type {
	case "static":
		return sd.createStaticWatcher()
	case "kubernetes":
		return sd.createKubernetesWatcher()
	case "consul":
		return sd.createConsulWatcher()
	case "dns":
		return sd.createDNSWatcher()
	default:
		return nil, fmt.Errorf("unsupported discovery type: %s", sd.config.Type)
	}
}

// createStaticWatcher creates a static target watcher
func (sd *ServiceDiscovery) createStaticWatcher() (TargetWatcher, error) {
	targetsConfig, ok := sd.config.Config["targets"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid static targets configuration")
	}

	var targets []*DiscoveryTarget
	for _, targetConfig := range targetsConfig {
		targetMap, ok := targetConfig.(map[string]interface{})
		if !ok {
			continue
		}

		address, _ := targetMap["address"].(string)
		port, _ := targetMap["port"].(float64)
		labels, _ := targetMap["labels"].(map[string]interface{})

		labelMap := make(map[string]string)
		for k, v := range labels {
			labelMap[k] = fmt.Sprintf("%v", v)
		}

		target := &DiscoveryTarget{
			Address:     address,
			Port:        int(port),
			Labels:      labelMap,
			Health:      "unknown",
			LastSeen:    time.Now(),
			ServiceType: "static",
		}
		targets = append(targets, target)
	}

	return &StaticTargetWatcher{
		targets: targets,
		logger:  sd.logger,
	}, nil
}

// createKubernetesWatcher creates a Kubernetes target watcher
func (sd *ServiceDiscovery) createKubernetesWatcher() (TargetWatcher, error) {
	namespace, _ := sd.config.Config["namespace"].(string)
	if namespace == "" {
		namespace = "default"
	}

	selector, _ := sd.config.Config["selector"].(map[string]interface{})
	selectorMap := make(map[string]string)
	for k, v := range selector {
		selectorMap[k] = fmt.Sprintf("%v", v)
	}

	return &KubernetesTargetWatcher{
		logger:    sd.logger,
		namespace: namespace,
		selector:  selectorMap,
		targets:   make(map[string]*DiscoveryTarget),
	}, nil
}

// createConsulWatcher creates a Consul target watcher
func (sd *ServiceDiscovery) createConsulWatcher() (TargetWatcher, error) {
	consulAddr, _ := sd.config.Config["address"].(string)
	if consulAddr == "" {
		consulAddr = "localhost:8500"
	}

	serviceName, _ := sd.config.Config["service"].(string)
	if serviceName == "" {
		serviceName = "gzh-manager"
	}

	return &ConsulTargetWatcher{
		logger:      sd.logger,
		consulAddr:  consulAddr,
		serviceName: serviceName,
		targets:     make(map[string]*DiscoveryTarget),
	}, nil
}

// createDNSWatcher creates a DNS target watcher
func (sd *ServiceDiscovery) createDNSWatcher() (TargetWatcher, error) {
	dnsName, _ := sd.config.Config["name"].(string)
	if dnsName == "" {
		return nil, fmt.Errorf("DNS name is required for DNS discovery")
	}

	port, _ := sd.config.Config["port"].(float64)
	if port == 0 {
		port = 8080
	}

	interval, _ := sd.config.Config["interval"].(string)
	refreshInterval, err := time.ParseDuration(interval)
	if err != nil {
		refreshInterval = 30 * time.Second
	}

	return &DNSTargetWatcher{
		logger:   sd.logger,
		dnsName:  dnsName,
		port:     int(port),
		interval: refreshInterval,
		targets:  make(map[string]*DiscoveryTarget),
	}, nil
}

// watchTargets watches for target changes
func (sd *ServiceDiscovery) watchTargets(ctx context.Context, watcher TargetWatcher) {
	err := watcher.Watch(func(targets []*DiscoveryTarget) {
		sd.updateTargets(targets)
	})
	if err != nil {
		sd.logger.Error("Failed to watch targets", zap.Error(err))
	}
}

// periodicRefresh periodically refreshes targets
func (sd *ServiceDiscovery) periodicRefresh(ctx context.Context) {
	interval := sd.config.ScrapeInterval
	if interval == 0 {
		interval = 30 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-sd.stopCh:
			return
		case <-ticker.C:
			sd.refreshTargets()
		}
	}
}

// refreshTargets refreshes all targets
func (sd *ServiceDiscovery) refreshTargets() {
	for _, watcher := range sd.watchers {
		targets, err := watcher.GetTargets()
		if err != nil {
			sd.logger.Error("Failed to get targets", zap.Error(err))
			continue
		}
		sd.updateTargets(targets)
	}
}

// updateTargets updates the target list
func (sd *ServiceDiscovery) updateTargets(targets []*DiscoveryTarget) {
	sd.mutex.Lock()
	defer sd.mutex.Unlock()

	// Clear existing targets
	sd.targets = make(map[string]*DiscoveryTarget)

	// Add new targets
	for _, target := range targets {
		key := fmt.Sprintf("%s:%d", target.Address, target.Port)
		target.LastSeen = time.Now()
		sd.targets[key] = target
	}

	sd.logger.Debug("Updated service discovery targets",
		zap.Int("count", len(targets)))
}

// GetTargets returns current targets
func (sd *ServiceDiscovery) GetTargets() []*DiscoveryTarget {
	sd.mutex.RLock()
	defer sd.mutex.RUnlock()

	targets := make([]*DiscoveryTarget, 0, len(sd.targets))
	for _, target := range sd.targets {
		targets = append(targets, target)
	}

	return targets
}

// GetPrometheusTargets returns targets in Prometheus format
func (sd *ServiceDiscovery) GetPrometheusTargets() []map[string]interface{} {
	targets := sd.GetTargets()
	prometheusTargets := make([]map[string]interface{}, 0, len(targets))

	for _, target := range targets {
		prometheusTarget := map[string]interface{}{
			"targets": []string{fmt.Sprintf("%s:%d", target.Address, target.Port)},
			"labels":  target.Labels,
		}
		prometheusTargets = append(prometheusTargets, prometheusTarget)
	}

	return prometheusTargets
}

// Static Target Watcher Implementation

func (stw *StaticTargetWatcher) Start(ctx context.Context) error {
	stw.logger.Info("Starting static target watcher",
		zap.Int("targets", len(stw.targets)))
	return nil
}

func (stw *StaticTargetWatcher) Stop() error {
	stw.logger.Info("Stopping static target watcher")
	return nil
}

func (stw *StaticTargetWatcher) GetTargets() ([]*DiscoveryTarget, error) {
	return stw.targets, nil
}

func (stw *StaticTargetWatcher) Watch(callback func([]*DiscoveryTarget)) error {
	// Static targets don't change, so call callback once
	callback(stw.targets)
	return nil
}

// DNS Target Watcher Implementation

func (dtw *DNSTargetWatcher) Start(ctx context.Context) error {
	dtw.logger.Info("Starting DNS target watcher",
		zap.String("dns_name", dtw.dnsName),
		zap.Int("port", dtw.port))

	// Initial discovery
	go dtw.discover()

	return nil
}

func (dtw *DNSTargetWatcher) Stop() error {
	dtw.logger.Info("Stopping DNS target watcher")
	return nil
}

func (dtw *DNSTargetWatcher) GetTargets() ([]*DiscoveryTarget, error) {
	dtw.mutex.RLock()
	defer dtw.mutex.RUnlock()

	targets := make([]*DiscoveryTarget, 0, len(dtw.targets))
	for _, target := range dtw.targets {
		targets = append(targets, target)
	}

	return targets, nil
}

func (dtw *DNSTargetWatcher) Watch(callback func([]*DiscoveryTarget)) error {
	go func() {
		ticker := time.NewTicker(dtw.interval)
		defer ticker.Stop()

		for range ticker.C {
			dtw.discover()
			targets, _ := dtw.GetTargets()
			callback(targets)
		}
	}()

	return nil
}

func (dtw *DNSTargetWatcher) discover() {
	ips, err := net.LookupIP(dtw.dnsName)
	if err != nil {
		dtw.logger.Error("DNS lookup failed",
			zap.String("dns_name", dtw.dnsName),
			zap.Error(err))
		return
	}

	dtw.mutex.Lock()
	defer dtw.mutex.Unlock()

	// Clear existing targets
	dtw.targets = make(map[string]*DiscoveryTarget)

	// Add discovered IPs
	for _, ip := range ips {
		target := &DiscoveryTarget{
			Address:     ip.String(),
			Port:        dtw.port,
			Labels:      map[string]string{"dns_name": dtw.dnsName},
			Health:      "unknown",
			LastSeen:    time.Now(),
			ServiceType: "dns",
		}
		dtw.targets[ip.String()] = target
	}

	dtw.logger.Debug("DNS discovery completed",
		zap.String("dns_name", dtw.dnsName),
		zap.Int("targets", len(dtw.targets)))
}

// Kubernetes Target Watcher Implementation (Simplified)

func (ktw *KubernetesTargetWatcher) Start(ctx context.Context) error {
	ktw.logger.Info("Starting Kubernetes target watcher",
		zap.String("namespace", ktw.namespace))

	// In a real implementation, this would connect to Kubernetes API
	// For now, we'll just log that it's not implemented
	ktw.logger.Warn("Kubernetes service discovery not fully implemented")

	return nil
}

func (ktw *KubernetesTargetWatcher) Stop() error {
	ktw.logger.Info("Stopping Kubernetes target watcher")
	return nil
}

func (ktw *KubernetesTargetWatcher) GetTargets() ([]*DiscoveryTarget, error) {
	ktw.mutex.RLock()
	defer ktw.mutex.RUnlock()

	targets := make([]*DiscoveryTarget, 0, len(ktw.targets))
	for _, target := range ktw.targets {
		targets = append(targets, target)
	}

	return targets, nil
}

func (ktw *KubernetesTargetWatcher) Watch(callback func([]*DiscoveryTarget)) error {
	// In a real implementation, this would watch Kubernetes API events
	ktw.logger.Info("Kubernetes target watching not implemented")
	return nil
}

// Consul Target Watcher Implementation (Simplified)

func (ctw *ConsulTargetWatcher) Start(ctx context.Context) error {
	ctw.logger.Info("Starting Consul target watcher",
		zap.String("consul_addr", ctw.consulAddr),
		zap.String("service", ctw.serviceName))

	// In a real implementation, this would connect to Consul API
	// For now, we'll just log that it's not implemented
	ctw.logger.Warn("Consul service discovery not fully implemented")

	return nil
}

func (ctw *ConsulTargetWatcher) Stop() error {
	ctw.logger.Info("Stopping Consul target watcher")
	return nil
}

func (ctw *ConsulTargetWatcher) GetTargets() ([]*DiscoveryTarget, error) {
	ctw.mutex.RLock()
	defer ctw.mutex.RUnlock()

	targets := make([]*DiscoveryTarget, 0, len(ctw.targets))
	for _, target := range ctw.targets {
		targets = append(targets, target)
	}

	return targets, nil
}

func (ctw *ConsulTargetWatcher) Watch(callback func([]*DiscoveryTarget)) error {
	// In a real implementation, this would watch Consul service changes
	ctw.logger.Info("Consul target watching not implemented")
	return nil
}
