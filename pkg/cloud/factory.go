// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package cloud

import (
	"context"
	"fmt"
	"time"
)

// ProviderType represents supported cloud provider types.
type ProviderType string

const (
	// ProviderTypeAWS represents Amazon Web Services.
	ProviderTypeAWS ProviderType = "aws"

	// ProviderTypeGCP represents Google Cloud Platform.
	ProviderTypeGCP ProviderType = "gcp"

	// ProviderTypeAzure represents Microsoft Azure.
	ProviderTypeAzure ProviderType = "azure"
)

// ProviderFactory is a factory function for creating providers.
type ProviderFactory func(ctx context.Context, config ProviderConfig) (Provider, error)

// Registry holds registered provider factories.
type Registry struct {
	factories map[ProviderType]ProviderFactory
}

// globalRegistry is the global provider registry.
var globalRegistry = &Registry{
	factories: make(map[ProviderType]ProviderFactory),
}

// Register registers a provider factory.
func Register(providerType ProviderType, factory ProviderFactory) {
	globalRegistry.factories[providerType] = factory
}

// NewProvider creates a new provider instance.
func NewProvider(ctx context.Context, config ProviderConfig) (Provider, error) {
	providerType := ProviderType(config.Type)

	factory, exists := globalRegistry.factories[providerType]
	if !exists {
		return nil, fmt.Errorf("unsupported provider type: %s", config.Type)
	}

	return factory(ctx, config)
}

// GetSupportedProviders returns list of supported provider types.
func GetSupportedProviders() []ProviderType {
	providers := make([]ProviderType, 0, len(globalRegistry.factories))
	for p := range globalRegistry.factories {
		providers = append(providers, p)
	}

	return providers
}

// IsProviderSupported checks if a provider type is supported.
func IsProviderSupported(providerType string) bool {
	_, exists := globalRegistry.factories[ProviderType(providerType)]
	return exists
}

// GetRegisteredProviders returns list of registered provider type names.
func GetRegisteredProviders() []string {
	providers := make([]string, 0, len(globalRegistry.factories))
	for p := range globalRegistry.factories {
		providers = append(providers, string(p))
	}

	return providers
}

// NewVPNManager creates a new VPN manager instance.
func NewVPNManager() VPNManager {
	return &defaultVPNManager{
		connections: make(map[string]*VPNConnection),
		statuses:    make(map[string]*VPNStatus),
	}
}

// NewHierarchicalVPNManager creates a new hierarchical VPN manager instance.
func NewHierarchicalVPNManager(baseManager VPNManager) HierarchicalVPNManager {
	return &defaultHierarchicalVPNManager{
		baseManager: baseManager,
		hierarchies: make(map[string]*VPNHierarchy),
	}
}

// NewPolicyManager creates a new policy manager instance.
func NewPolicyManager(config *Config) PolicyManager {
	return &defaultPolicyManager{
		policies: make(map[string]*NetworkPolicy),
		config:   config,
	}
}

// defaultVPNManager is the default implementation of VPNManager.
type defaultVPNManager struct {
	connections map[string]*VPNConnection
	statuses    map[string]*VPNStatus
}

// AddVPNConnection adds a VPN connection.
func (m *defaultVPNManager) AddVPNConnection(conn *VPNConnection) error {
	if conn == nil {
		return fmt.Errorf("VPN connection cannot be nil")
	}

	if conn.Name == "" {
		return fmt.Errorf("VPN connection name cannot be empty")
	}

	m.connections[conn.Name] = conn

	return nil
}

// RemoveVPNConnection removes a VPN connection.
func (m *defaultVPNManager) RemoveVPNConnection(name string) error {
	if name == "" {
		return fmt.Errorf("VPN connection name cannot be empty")
	}

	delete(m.connections, name)
	delete(m.statuses, name)

	return nil
}

// GetVPNConnection retrieves a VPN connection by name.
func (m *defaultVPNManager) GetVPNConnection(name string) (*VPNConnection, error) {
	if name == "" {
		return nil, fmt.Errorf("VPN connection name cannot be empty")
	}

	conn, exists := m.connections[name]
	if !exists {
		return nil, fmt.Errorf("VPN connection not found: %s", name)
	}

	return conn, nil
}

// ListVPNConnections lists all VPN connections.
func (m *defaultVPNManager) ListVPNConnections() ([]*VPNConnection, error) {
	connections := make([]*VPNConnection, 0, len(m.connections))
	for _, conn := range m.connections {
		connections = append(connections, conn)
	}

	return connections, nil
}

// ConnectVPN connects to a VPN (mock implementation).
func (m *defaultVPNManager) ConnectVPN(ctx context.Context, name string) error {
	conn, err := m.GetVPNConnection(name)
	if err != nil {
		return err
	}
	// Mock implementation - in real implementation, this would connect to the VPN
	m.statuses[name] = &VPNStatus{
		Name:   conn.Name,
		Status: "connected",
	}

	return nil
}

// DisconnectVPN disconnects from a VPN (mock implementation).
func (m *defaultVPNManager) DisconnectVPN(ctx context.Context, name string) error {
	_, err := m.GetVPNConnection(name)
	if err != nil {
		return err
	}
	// Mock implementation - in real implementation, this would disconnect from the VPN
	if status, exists := m.statuses[name]; exists {
		status.Status = "disconnected"
	}

	return nil
}

// GetVPNStatus returns the status of a VPN connection.
func (m *defaultVPNManager) GetVPNStatus(ctx context.Context, name string) (*VPNStatus, error) {
	status, exists := m.statuses[name]
	if !exists {
		return &VPNStatus{
			Name:   name,
			Status: "disconnected",
		}, nil
	}

	return status, nil
}

// GetAllVPNStatuses returns statuses of all VPN connections.
func (m *defaultVPNManager) GetAllVPNStatuses(ctx context.Context) (map[string]*VPNStatus, error) {
	statuses := make(map[string]*VPNStatus)

	for name := range m.connections {
		status, _ := m.GetVPNStatus(ctx, name)
		statuses[name] = status
	}

	return statuses, nil
}

// GetConnectionStatus returns the status of a VPN connection (alias for GetVPNStatus).
func (m *defaultVPNManager) GetConnectionStatus(ctx context.Context, name string) (*VPNStatus, error) {
	return m.GetVPNStatus(ctx, name)
}

// GetActiveConnections returns all active VPN connections.
func (m *defaultVPNManager) GetActiveConnections(ctx context.Context) (map[string]*VPNStatus, error) {
	activeConnections := make(map[string]*VPNStatus)

	for name := range m.connections {
		status, err := m.GetVPNStatus(ctx, name)
		if err != nil {
			continue
		}
		// Only include connected connections
		if status.Status == VPNStateConnected {
			activeConnections[name] = status
		}
	}

	return activeConnections, nil
}

// ConnectByPriority connects VPN connections by priority order.
func (m *defaultVPNManager) ConnectByPriority(ctx context.Context, connectionNames []string) error {
	// Mock implementation - in real implementation, this would connect VPNs by priority
	for _, name := range connectionNames {
		if err := m.ConnectVPN(ctx, name); err != nil {
			return fmt.Errorf("failed to connect VPN %s: %w", name, err)
		}
	}

	return nil
}

// defaultHierarchicalVPNManager is the default implementation of HierarchicalVPNManager.
type defaultHierarchicalVPNManager struct {
	baseManager VPNManager
	hierarchies map[string]*VPNHierarchy
}

// AddVPNConnection adds a VPN connection.
func (m *defaultHierarchicalVPNManager) AddVPNConnection(conn *VPNConnection) error {
	return m.baseManager.AddVPNConnection(conn)
}

// RemoveVPNConnection removes a VPN connection.
func (m *defaultHierarchicalVPNManager) RemoveVPNConnection(name string) error {
	return m.baseManager.RemoveVPNConnection(name)
}

// GetVPNConnection retrieves a VPN connection by name.
func (m *defaultHierarchicalVPNManager) GetVPNConnection(name string) (*VPNConnection, error) {
	return m.baseManager.GetVPNConnection(name)
}

// ListVPNConnections lists all VPN connections.
func (m *defaultHierarchicalVPNManager) ListVPNConnections() ([]*VPNConnection, error) {
	return m.baseManager.ListVPNConnections()
}

// ConnectVPN connects to a VPN.
func (m *defaultHierarchicalVPNManager) ConnectVPN(ctx context.Context, name string) error {
	return m.baseManager.ConnectVPN(ctx, name)
}

// DisconnectVPN disconnects from a VPN.
func (m *defaultHierarchicalVPNManager) DisconnectVPN(ctx context.Context, name string) error {
	return m.baseManager.DisconnectVPN(ctx, name)
}

// GetVPNStatus returns the status of a VPN connection.
func (m *defaultHierarchicalVPNManager) GetVPNStatus(ctx context.Context, name string) (*VPNStatus, error) {
	return m.baseManager.GetVPNStatus(ctx, name)
}

// GetAllVPNStatuses returns statuses of all VPN connections.
func (m *defaultHierarchicalVPNManager) GetAllVPNStatuses(ctx context.Context) (map[string]*VPNStatus, error) {
	return m.baseManager.GetAllVPNStatuses(ctx)
}

// GetConnectionStatus returns the status of a VPN connection (alias for GetVPNStatus).
func (m *defaultHierarchicalVPNManager) GetConnectionStatus(ctx context.Context, name string) (*VPNStatus, error) {
	return m.baseManager.GetConnectionStatus(ctx, name)
}

// GetActiveConnections returns all active VPN connections.
func (m *defaultHierarchicalVPNManager) GetActiveConnections(ctx context.Context) (map[string]*VPNStatus, error) {
	return m.baseManager.GetActiveConnections(ctx)
}

// ConnectByPriority connects VPN connections by priority order.
func (m *defaultHierarchicalVPNManager) ConnectByPriority(ctx context.Context, connectionNames []string) error {
	return m.baseManager.ConnectByPriority(ctx, connectionNames)
}

// AddVPNHierarchy adds a VPN hierarchy.
func (m *defaultHierarchicalVPNManager) AddVPNHierarchy(hierarchy *VPNHierarchy) error {
	if hierarchy == nil {
		return fmt.Errorf("VPN hierarchy cannot be nil")
	}

	if hierarchy.Name == "" {
		return fmt.Errorf("VPN hierarchy name cannot be empty")
	}

	m.hierarchies[hierarchy.Name] = hierarchy

	return nil
}

// RemoveVPNHierarchy removes a VPN hierarchy.
func (m *defaultHierarchicalVPNManager) RemoveVPNHierarchy(name string) error {
	if name == "" {
		return fmt.Errorf("VPN hierarchy name cannot be empty")
	}

	delete(m.hierarchies, name)

	return nil
}

// GetVPNHierarchy retrieves a VPN hierarchy by name.
func (m *defaultHierarchicalVPNManager) GetVPNHierarchy(name string) (*VPNHierarchy, error) {
	if name == "" {
		return nil, fmt.Errorf("VPN hierarchy name cannot be empty")
	}

	hierarchy, exists := m.hierarchies[name]
	if !exists {
		return nil, fmt.Errorf("VPN hierarchy not found: %s", name)
	}

	return hierarchy, nil
}

// ListVPNHierarchies lists all VPN hierarchies.
func (m *defaultHierarchicalVPNManager) ListVPNHierarchies() ([]*VPNHierarchy, error) {
	hierarchies := make([]*VPNHierarchy, 0, len(m.hierarchies))
	for _, hierarchy := range m.hierarchies {
		hierarchies = append(hierarchies, hierarchy)
	}

	return hierarchies, nil
}

// ConnectVPNHierarchy connects to a VPN hierarchy (mock implementation).
func (m *defaultHierarchicalVPNManager) ConnectVPNHierarchy(ctx context.Context, name string) error {
	hierarchy, err := m.GetVPNHierarchy(name)
	if err != nil {
		return err
	}
	// Mock implementation - in real implementation, this would connect to the VPN hierarchy
	for _, layer := range hierarchy.Layers {
		for _, node := range layer {
			if node.Connection != nil {
				_ = m.baseManager.ConnectVPN(ctx, node.Connection.Name)
			}
		}
	}

	return nil
}

// DisconnectVPNHierarchy disconnects from a VPN hierarchy (mock implementation).
func (m *defaultHierarchicalVPNManager) DisconnectVPNHierarchy(ctx context.Context, name string) error {
	hierarchy, err := m.GetVPNHierarchy(name)
	if err != nil {
		return err
	}
	// Mock implementation - in real implementation, this would disconnect from the VPN hierarchy
	for _, layer := range hierarchy.Layers {
		for _, node := range layer {
			if node.Connection != nil {
				_ = m.baseManager.DisconnectVPN(ctx, node.Connection.Name)
			}
		}
	}

	return nil
}

// GetVPNHierarchyStatus returns the status of a VPN hierarchy (mock implementation).
func (m *defaultHierarchicalVPNManager) GetVPNHierarchyStatus(ctx context.Context, name string) (*VPNHierarchyStatus, error) {
	hierarchy, err := m.GetVPNHierarchy(name)
	if err != nil {
		return nil, err
	}

	// Mock implementation - in real implementation, this would get actual status
	return &VPNHierarchyStatus{
		Name:   hierarchy.Name,
		Status: "disconnected",
	}, nil
}

// defaultPolicyManager is the default implementation of PolicyManager.
type defaultPolicyManager struct {
	policies map[string]*NetworkPolicy
	config   *Config
}

// AddPolicy adds a network policy.
func (m *defaultPolicyManager) AddPolicy(policy *NetworkPolicy) error {
	if policy == nil {
		return fmt.Errorf("network policy cannot be nil")
	}

	if policy.Name == "" {
		return fmt.Errorf("network policy name cannot be empty")
	}

	m.policies[policy.Name] = policy

	return nil
}

// RemovePolicy removes a network policy.
func (m *defaultPolicyManager) RemovePolicy(name string) error {
	if name == "" {
		return fmt.Errorf("network policy name cannot be empty")
	}

	delete(m.policies, name)

	return nil
}

// GetPolicy retrieves a network policy by name.
func (m *defaultPolicyManager) GetPolicy(name string) (*NetworkPolicy, error) {
	if name == "" {
		return nil, fmt.Errorf("network policy name cannot be empty")
	}

	policy, exists := m.policies[name]
	if !exists {
		return nil, fmt.Errorf("network policy not found: %s", name)
	}

	return policy, nil
}

// ListPolicies lists all network policies.
func (m *defaultPolicyManager) ListPolicies() ([]*NetworkPolicy, error) {
	policies := make([]*NetworkPolicy, 0, len(m.policies))
	for _, policy := range m.policies {
		policies = append(policies, policy)
	}

	return policies, nil
}

// ApplyPolicy applies a network policy (mock implementation).
func (m *defaultPolicyManager) ApplyPolicy(ctx context.Context, name string) error {
	_, err := m.GetPolicy(name)
	if err != nil {
		return err
	}
	// Mock implementation - in real implementation, this would apply the policy
	return nil
}

// RemoveAppliedPolicy removes a network policy (mock implementation).
func (m *defaultPolicyManager) RemoveAppliedPolicy(ctx context.Context, name string) error {
	_, err := m.GetPolicy(name)
	if err != nil {
		return err
	}
	// Mock implementation - in real implementation, this would remove the applied policy
	return nil
}

// ApplyEnvironmentPolicies applies policies for an environment (mock implementation).
func (m *defaultPolicyManager) ApplyEnvironmentPolicies(ctx context.Context, environment string) error {
	// Mock implementation - in real implementation, this would apply policies for the environment
	return nil
}

// GetApplicablePolicies gets applicable policies for a profile (mock implementation).
func (m *defaultPolicyManager) GetApplicablePolicies(ctx context.Context, profileName string) ([]*NetworkPolicy, error) {
	// Mock implementation - in real implementation, this would return applicable policies for the profile
	var applicablePolicies []*NetworkPolicy

	// Get policies from config if available
	if m.config != nil {
		for _, policy := range m.config.Policies {
			// Simple matching - in real implementation, this would be more sophisticated
			if policy.ProfileName == profileName || policy.ProfileName == "" {
				p := policy // Create copy to avoid memory aliasing
				applicablePolicies = append(applicablePolicies, &p)
			}
		}
	}

	// Also check manager's policies
	for _, policy := range m.policies {
		if policy.ProfileName == profileName || policy.ProfileName == "" {
			applicablePolicies = append(applicablePolicies, policy)
		}
	}

	return applicablePolicies, nil
}

// ApplyPoliciesForProfile applies policies for a specific profile (mock implementation).
func (m *defaultPolicyManager) ApplyPoliciesForProfile(ctx context.Context, profileName string) error {
	// Mock implementation - in real implementation, this would apply policies for the profile
	return nil
}

// GetPolicyStatus gets the status of applied policies (mock implementation).
func (m *defaultPolicyManager) GetPolicyStatus(ctx context.Context) ([]*PolicyStatus, error) {
	// Mock implementation - in real implementation, this would return the status of applied policies
	statuses := make([]*PolicyStatus, 0, len(m.policies))

	// Get all policies and create status entries
	for _, policy := range m.policies {
		status := &PolicyStatus{
			PolicyName:  policy.Name,
			ProfileName: policy.ProfileName,
			Provider:    policy.Provider,
			Status:      "applied", // Mock status
			Applied:     time.Now(),
			Error:       "",
		}
		statuses = append(statuses, status)
	}

	// Also include policies from config
	if m.config != nil {
		for _, policy := range m.config.Policies {
			status := &PolicyStatus{
				PolicyName:  policy.Name,
				ProfileName: policy.ProfileName,
				Provider:    policy.Provider,
				Status:      "applied", // Mock status
				Applied:     time.Now(),
				Error:       "",
			}
			statuses = append(statuses, status)
		}
	}

	return statuses, nil
}

// GetPolicyStatusForProfile gets the status of applied policies for a specific profile (mock implementation).
func (m *defaultPolicyManager) GetPolicyStatusForProfile(ctx context.Context, profileName string) (map[string]string, error) {
	// Mock implementation - in real implementation, this would return the status of applied policies
	status := make(map[string]string)

	// Get applicable policies and set their status
	policies, err := m.GetApplicablePolicies(ctx, profileName)
	if err != nil {
		return nil, err
	}

	for _, policy := range policies {
		status[policy.Name] = "applied" // Mock status
	}

	return status, nil
}

// ValidatePolicy validates a network policy (mock implementation).
func (m *defaultPolicyManager) ValidatePolicy(ctx context.Context, policy *NetworkPolicy) error {
	// Mock implementation - in real implementation, this would validate the policy
	if policy == nil {
		return fmt.Errorf("policy cannot be nil")
	}

	if policy.Name == "" {
		return fmt.Errorf("policy name cannot be empty")
	}

	return nil
}
