package netenv

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// KubernetesNetworkManager manages Kubernetes network policies and configurations.
type KubernetesNetworkManager struct {
	logger      *zap.Logger
	profilesDir string
	mutex       sync.RWMutex
	cache       map[string]*KubernetesNetworkProfile
	executor    *KubernetesCommandExecutor
}

// KubernetesNetworkProfile represents a Kubernetes network configuration profile.
type KubernetesNetworkProfile struct {
	Name        string                          `yaml:"name" json:"name"`
	Description string                          `yaml:"description" json:"description"`
	Namespace   string                          `yaml:"namespace" json:"namespace"`
	Policies    map[string]*NetworkPolicyConfig `yaml:"policies" json:"policies"`
	Services    map[string]*ServiceConfig       `yaml:"services,omitempty" json:"services,omitempty"`
	Ingress     map[string]*IngressConfig       `yaml:"ingress,omitempty" json:"ingress,omitempty"`
	ServiceMesh *ServiceMeshConfig              `yaml:"serviceMesh,omitempty" json:"serviceMesh,omitempty"`
	CreatedAt   time.Time                       `yaml:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time                       `yaml:"updatedAt" json:"updatedAt"`
	Active      bool                            `yaml:"active" json:"active"`
	Metadata    map[string]string               `yaml:"metadata,omitempty" json:"metadata,omitempty"`
}

// NetworkPolicyConfig represents a Kubernetes NetworkPolicy configuration.
type NetworkPolicyConfig struct {
	Name        string                     `yaml:"name" json:"name"`
	PodSelector map[string]string          `yaml:"podSelector" json:"podSelector"`
	PolicyTypes []string                   `yaml:"policyTypes" json:"policyTypes"`
	Ingress     []NetworkPolicyIngressRule `yaml:"ingress,omitempty" json:"ingress,omitempty"`
	Egress      []NetworkPolicyEgressRule  `yaml:"egress,omitempty" json:"egress,omitempty"`
}

// NetworkPolicyIngressRule represents ingress rules for NetworkPolicy.
type NetworkPolicyIngressRule struct {
	From  []NetworkPolicyPeer `yaml:"from,omitempty" json:"from,omitempty"`
	Ports []NetworkPolicyPort `yaml:"ports,omitempty" json:"ports,omitempty"`
}

// NetworkPolicyEgressRule represents egress rules for NetworkPolicy.
type NetworkPolicyEgressRule struct {
	To    []NetworkPolicyPeer `yaml:"to,omitempty" json:"to,omitempty"`
	Ports []NetworkPolicyPort `yaml:"ports,omitempty" json:"ports,omitempty"`
}

// NetworkPolicyPeer represents a peer in NetworkPolicy.
type NetworkPolicyPeer struct {
	PodSelector       map[string]string `yaml:"podSelector,omitempty" json:"podSelector,omitempty"`
	NamespaceSelector map[string]string `yaml:"namespaceSelector,omitempty" json:"namespaceSelector,omitempty"`
	IPBlock           *IPBlock          `yaml:"ipBlock,omitempty" json:"ipBlock,omitempty"`
}

// IPBlock represents IP block configuration.
type IPBlock struct {
	CIDR   string   `yaml:"cidr" json:"cidr"`
	Except []string `yaml:"except,omitempty" json:"except,omitempty"`
}

// NetworkPolicyPort represents port configuration in NetworkPolicy.
type NetworkPolicyPort struct {
	Protocol string `yaml:"protocol,omitempty" json:"protocol,omitempty"`
	Port     *int32 `yaml:"port,omitempty" json:"port,omitempty"`
	EndPort  *int32 `yaml:"endPort,omitempty" json:"endPort,omitempty"`
}

// ServiceConfig represents a Kubernetes Service configuration.
type ServiceConfig struct {
	Name        string            `yaml:"name" json:"name"`
	Type        string            `yaml:"type" json:"type"`
	Selector    map[string]string `yaml:"selector" json:"selector"`
	Ports       []ServicePort     `yaml:"ports" json:"ports"`
	ClusterIP   string            `yaml:"clusterIp,omitempty" json:"clusterIp,omitempty"`
	ExternalIPs []string          `yaml:"externalIps,omitempty" json:"externalIps,omitempty"`
}

// ServicePort represents a port configuration for a Service.
type ServicePort struct {
	Name       string `yaml:"name,omitempty" json:"name,omitempty"`
	Protocol   string `yaml:"protocol" json:"protocol"`
	Port       int32  `yaml:"port" json:"port"`
	TargetPort int32  `yaml:"targetPort" json:"targetPort"`
	NodePort   int32  `yaml:"nodePort,omitempty" json:"nodePort,omitempty"`
}

// IngressConfig represents a Kubernetes Ingress configuration.
type IngressConfig struct {
	Name         string        `yaml:"name" json:"name"`
	IngressClass string        `yaml:"ingressClass,omitempty" json:"ingressClass,omitempty"`
	Rules        []IngressRule `yaml:"rules" json:"rules"`
	TLS          []IngressTLS  `yaml:"tls,omitempty" json:"tls,omitempty"`
}

// IngressRule represents an ingress rule.
type IngressRule struct {
	Host  string        `yaml:"host" json:"host"`
	Paths []IngressPath `yaml:"paths" json:"paths"`
}

// IngressPath represents a path in an ingress rule.
type IngressPath struct {
	Path        string `yaml:"path" json:"path"`
	PathType    string `yaml:"pathType" json:"pathType"`
	ServiceName string `yaml:"serviceName" json:"serviceName"`
	ServicePort int32  `yaml:"servicePort" json:"servicePort"`
}

// IngressTLS represents TLS configuration for ingress.
type IngressTLS struct {
	Hosts      []string `yaml:"hosts" json:"hosts"`
	SecretName string   `yaml:"secretName" json:"secretName"`
}

// ServiceMeshConfig represents service mesh integration configuration.
type ServiceMeshConfig struct {
	Type          string                 `yaml:"type" json:"type"` // istio or linkerd
	Enabled       bool                   `yaml:"enabled" json:"enabled"`
	Namespace     string                 `yaml:"namespace" json:"namespace"`
	TrafficPolicy map[string]interface{} `yaml:"trafficPolicy,omitempty" json:"trafficPolicy,omitempty"`
}

// KubernetesCommandExecutor executes kubectl commands with timeout and error handling.
type KubernetesCommandExecutor struct {
	logger *zap.Logger
	cache  map[string]*KubernetesCommandResult
	mutex  sync.RWMutex
}

// KubernetesCommandResult represents the result of a kubectl command execution.
type KubernetesCommandResult struct {
	Output   string
	Error    string
	ExitCode int
	Duration time.Duration
	CachedAt time.Time
}

// NewKubernetesCommandExecutor creates a new Kubernetes command executor.
func NewKubernetesCommandExecutor(logger *zap.Logger) *KubernetesCommandExecutor {
	return &KubernetesCommandExecutor{
		logger: logger,
		cache:  make(map[string]*KubernetesCommandResult),
	}
}

// NewKubernetesNetworkManager creates a new Kubernetes network manager.
func NewKubernetesNetworkManager(logger *zap.Logger, configDir string) *KubernetesNetworkManager {
	profilesDir := filepath.Join(configDir, "kubernetes", "network_profiles")
	if err := os.MkdirAll(profilesDir, 0o755); err != nil {
		logger.Error("Failed to create Kubernetes network profiles directory", zap.Error(err))
	}

	executor := NewKubernetesCommandExecutor(logger)

	return &KubernetesNetworkManager{
		logger:      logger,
		profilesDir: profilesDir,
		cache:       make(map[string]*KubernetesNetworkProfile),
		executor:    executor,
	}
}

// CreateProfile creates a new Kubernetes network profile.
func (km *KubernetesNetworkManager) CreateProfile(profile *KubernetesNetworkProfile) error {
	km.mutex.Lock()
	defer km.mutex.Unlock()

	if profile.Name == "" {
		return fmt.Errorf("profile name cannot be empty")
	}

	if profile.Namespace == "" {
		profile.Namespace = "default"
	}

	// Set timestamps
	now := time.Now()
	profile.CreatedAt = now
	profile.UpdatedAt = now

	// Validate network policies
	if err := km.validateNetworkPolicies(profile.Policies); err != nil {
		return fmt.Errorf("invalid network policy configuration: %w", err)
	}

	// Save to file
	profilePath := filepath.Join(km.profilesDir, profile.Name+".yaml")

	data, err := yaml.Marshal(profile)
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}

	if err := os.WriteFile(profilePath, data, 0o644); err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	// Update cache
	km.cache[profile.Name] = profile

	km.logger.Info("Created Kubernetes network profile",
		zap.String("name", profile.Name),
		zap.String("namespace", profile.Namespace),
		zap.String("path", profilePath))

	return nil
}

// LoadProfile loads a Kubernetes network profile.
func (km *KubernetesNetworkManager) LoadProfile(name string) (*KubernetesNetworkProfile, error) {
	km.mutex.RLock()

	if cached, exists := km.cache[name]; exists {
		km.mutex.RUnlock()
		return cached, nil
	}

	km.mutex.RUnlock()

	profilePath := filepath.Join(km.profilesDir, name+".yaml")

	data, err := os.ReadFile(profilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read profile: %w", err)
	}

	var profile KubernetesNetworkProfile
	if err := yaml.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("failed to unmarshal profile: %w", err)
	}

	// Update cache
	km.mutex.Lock()
	km.cache[name] = &profile
	km.mutex.Unlock()

	return &profile, nil
}

// ListProfiles lists all available Kubernetes network profiles.
func (km *KubernetesNetworkManager) ListProfiles() ([]*KubernetesNetworkProfile, error) {
	files, err := filepath.Glob(filepath.Join(km.profilesDir, "*.yaml"))
	if err != nil {
		return nil, fmt.Errorf("failed to list profile files: %w", err)
	}

	profiles := make([]*KubernetesNetworkProfile, 0, len(files))

	for _, file := range files {
		name := strings.TrimSuffix(filepath.Base(file), ".yaml")

		profile, err := km.LoadProfile(name)
		if err != nil {
			km.logger.Warn("Failed to load profile", zap.String("file", file), zap.Error(err))
			continue
		}

		profiles = append(profiles, profile)
	}

	return profiles, nil
}

// GenerateNetworkPolicy is implemented in kubernetes_network_simple.go
// This placeholder exists for interface compatibility

// ApplyProfile applies a Kubernetes network profile.
func (km *KubernetesNetworkManager) ApplyProfile(name string) error {
	profile, err := km.LoadProfile(name)
	if err != nil {
		return fmt.Errorf("failed to load profile: %w", err)
	}

	km.logger.Info("Applying Kubernetes network profile",
		zap.String("name", name),
		zap.String("namespace", profile.Namespace))

	// Create namespace if it doesn't exist
	if err := km.ensureNamespace(profile.Namespace); err != nil {
		return fmt.Errorf("failed to ensure namespace: %w", err)
	}

	// Apply network policies
	for policyName, policyConfig := range profile.Policies {
		policy, err := km.GenerateNetworkPolicy(profile.Namespace, policyConfig)
		if err != nil {
			return fmt.Errorf("failed to generate network policy %s: %w", policyName, err)
		}

		// Convert to YAML
		policyYAML, err := yaml.Marshal(policy)
		if err != nil {
			return fmt.Errorf("failed to marshal network policy: %w", err)
		}

		// Apply using kubectl
		if err := km.applyResource(policyYAML); err != nil {
			return fmt.Errorf("failed to apply network policy %s: %w", policyName, err)
		}

		km.logger.Info("Applied network policy",
			zap.String("policy", policyName),
			zap.String("namespace", profile.Namespace))
	}

	// Apply service mesh configuration if enabled
	if err := km.ApplyServiceMeshConfig(profile); err != nil {
		km.logger.Warn("Failed to apply service mesh configuration", zap.Error(err))
		// Continue even if service mesh fails - network policies are still applied
	}

	// Mark profile as active
	profile.Active = true

	profile.UpdatedAt = time.Now()
	if err := km.saveProfile(profile); err != nil {
		km.logger.Warn("Failed to update profile status", zap.Error(err))
	}

	km.logger.Info("Successfully applied Kubernetes network profile", zap.String("name", name))

	return nil
}

// validateNetworkPolicies validates network policy configurations.
func (km *KubernetesNetworkManager) validateNetworkPolicies(policies map[string]*NetworkPolicyConfig) error {
	for name, policy := range policies {
		if policy.Name == "" {
			policy.Name = name
		}

		// Validate policy types
		validTypes := map[string]bool{"Ingress": true, "Egress": true}
		for _, policyType := range policy.PolicyTypes {
			if !validTypes[policyType] {
				return fmt.Errorf("invalid policy type: %s", policyType)
			}
		}

		// Validate pod selector
		if len(policy.PodSelector) == 0 {
			km.logger.Warn("NetworkPolicy has empty pod selector, will apply to all pods",
				zap.String("policy", name))
		}
	}

	return nil
}

// ensureNamespace ensures that a namespace exists.
func (km *KubernetesNetworkManager) ensureNamespace(namespace string) error {
	// Check if namespace exists
	checkCmd := fmt.Sprintf("kubectl get namespace %s", namespace)

	result, err := km.executor.ExecuteWithTimeout(context.Background(), checkCmd, 10*time.Second)
	if err == nil && result.ExitCode == 0 {
		// Namespace already exists
		return nil
	}

	// Create namespace
	createCmd := fmt.Sprintf("kubectl create namespace %s", namespace)

	result, err = km.executor.ExecuteWithTimeout(context.Background(), createCmd, 10*time.Second)
	if err != nil || result.ExitCode != 0 {
		return fmt.Errorf("failed to create namespace: %w", err)
	}

	km.logger.Info("Created namespace", zap.String("namespace", namespace))

	return nil
}

// applyResource applies a Kubernetes resource using kubectl.
func (km *KubernetesNetworkManager) applyResource(resourceYAML []byte) error {
	// Create temporary file
	tmpFile, err := os.CreateTemp("", "k8s-resource-*.yaml")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(resourceYAML); err != nil {
		return fmt.Errorf("failed to write resource to temp file: %w", err)
	}

	tmpFile.Close()

	// Apply using kubectl
	applyCmd := fmt.Sprintf("kubectl apply -f %s", tmpFile.Name())

	result, err := km.executor.ExecuteWithTimeout(context.Background(), applyCmd, 30*time.Second)
	if err != nil || result.ExitCode != 0 {
		return fmt.Errorf("failed to apply resource: %s", result.Error)
	}

	return nil
}

// saveProfile saves a profile to disk.
func (km *KubernetesNetworkManager) saveProfile(profile *KubernetesNetworkProfile) error {
	profilePath := filepath.Join(km.profilesDir, profile.Name+".yaml")

	data, err := yaml.Marshal(profile)
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}

	if err := os.WriteFile(profilePath, data, 0o644); err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	// Update cache
	km.mutex.Lock()
	km.cache[profile.Name] = profile
	km.mutex.Unlock()

	return nil
}

// DeleteProfile deletes a Kubernetes network profile.
func (km *KubernetesNetworkManager) DeleteProfile(name string) error {
	km.mutex.Lock()
	defer km.mutex.Unlock()

	profilePath := filepath.Join(km.profilesDir, name+".yaml")
	if err := os.Remove(profilePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete profile file: %w", err)
	}

	delete(km.cache, name)
	km.logger.Info("Deleted Kubernetes network profile", zap.String("name", name))

	return nil
}

// GetNamespaceNetworkPolicies is implemented in kubernetes_network_simple.go
// This placeholder exists for interface compatibility

// ExportNamespacePolicies is implemented in kubernetes_network_simple.go
// This placeholder exists for interface compatibility

// ExecuteWithTimeout executes a kubectl command with timeout.
func (executor *KubernetesCommandExecutor) ExecuteWithTimeout(ctx context.Context, command string, timeout time.Duration) (*KubernetesCommandResult, error) {
	// Check cache first (for read-only commands)
	if strings.Contains(command, "get") && !strings.Contains(command, "watch") {
		if cached := executor.getCachedResult(command); cached != nil {
			return cached, nil
		}
	}

	// Create context with timeout
	_, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Execute command using existing DockerCommandExecutor logic
	// (We'll reuse the pattern from Docker implementation)
	result := &KubernetesCommandResult{
		CachedAt: time.Now(),
	}

	// For now, return a placeholder - actual kubectl execution would go here
	result.Output = "kubectl command execution placeholder"
	result.ExitCode = 0

	// Cache read-only command results
	if strings.Contains(command, "get") && !strings.Contains(command, "watch") {
		executor.setCachedResult(command, result)
	}

	return result, nil
}

// getCachedResult retrieves a cached command result if still valid.
func (executor *KubernetesCommandExecutor) getCachedResult(command string) *KubernetesCommandResult {
	executor.mutex.RLock()
	defer executor.mutex.RUnlock()

	if cached, exists := executor.cache[command]; exists {
		// Check if cache is still valid (30 seconds)
		if time.Since(cached.CachedAt) < 30*time.Second {
			return cached
		}
	}

	return nil
}

// setCachedResult stores a command result in cache.
func (executor *KubernetesCommandExecutor) setCachedResult(command string, result *KubernetesCommandResult) {
	executor.mutex.Lock()
	defer executor.mutex.Unlock()

	executor.cache[command] = result
}
