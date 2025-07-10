package cloud

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

// HierarchicalVPNManager extends VPNManager with hierarchical management capabilities
type HierarchicalVPNManager struct {
	baseManager VPNManager
	connections map[string]*VPNConnection
	hierarchy   map[string]*VPNHierarchyNode
	mu          sync.RWMutex
}

// NewHierarchicalVPNManager creates a new hierarchical VPN manager
func NewHierarchicalVPNManager(baseManager VPNManager) *HierarchicalVPNManager {
	return &HierarchicalVPNManager{
		baseManager: baseManager,
		connections: make(map[string]*VPNConnection),
		hierarchy:   make(map[string]*VPNHierarchyNode),
	}
}

// ConnectHierarchical connects VPN connections in hierarchical order
func (h *HierarchicalVPNManager) ConnectHierarchical(ctx context.Context, rootConnection string) error {
	h.mu.RLock()
	rootNode, exists := h.hierarchy[rootConnection]
	h.mu.RUnlock()

	if !exists {
		return fmt.Errorf("root connection '%s' not found in hierarchy", rootConnection)
	}

	// Build connection order based on hierarchy layers
	connectionOrder := h.buildConnectionOrder(rootNode)

	// Connect in order
	for _, conn := range connectionOrder {
		if err := h.connectWithDependencies(ctx, conn); err != nil {
			return fmt.Errorf("failed to connect %s: %w", conn.Name, err)
		}
	}

	return nil
}

// DisconnectHierarchical disconnects VPN connections in reverse hierarchical order
func (h *HierarchicalVPNManager) DisconnectHierarchical(ctx context.Context, rootConnection string) error {
	h.mu.RLock()
	rootNode, exists := h.hierarchy[rootConnection]
	h.mu.RUnlock()

	if !exists {
		return fmt.Errorf("root connection '%s' not found in hierarchy", rootConnection)
	}

	// Build disconnection order (reverse of connection order)
	connectionOrder := h.buildConnectionOrder(rootNode)

	// Reverse the order for disconnection
	for i := len(connectionOrder) - 1; i >= 0; i-- {
		conn := connectionOrder[i]
		if err := h.baseManager.DisconnectVPN(ctx, conn.Name); err != nil {
			// Log error but continue disconnecting other connections
			fmt.Printf("Warning: failed to disconnect %s: %v\n", conn.Name, err)
		}
	}

	return nil
}

// GetVPNHierarchy returns the VPN connection hierarchy tree
func (h *HierarchicalVPNManager) GetVPNHierarchy() map[string]*VPNHierarchyNode {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Return a copy of the hierarchy to prevent external modification
	result := make(map[string]*VPNHierarchyNode)
	for k, v := range h.hierarchy {
		result[k] = v
	}
	return result
}

// ValidateHierarchy validates VPN hierarchy configuration
func (h *HierarchicalVPNManager) ValidateHierarchy() error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Check for circular dependencies
	visited := make(map[string]bool)
	recursionStack := make(map[string]bool)

	for name, node := range h.hierarchy {
		if !visited[name] {
			if h.hasCircularDependency(node, visited, recursionStack) {
				return fmt.Errorf("circular dependency detected in VPN hierarchy starting from %s", name)
			}
		}
	}

	// Validate parent-child relationships
	for _, node := range h.hierarchy {
		if err := h.validateNodeRelationships(node); err != nil {
			return err
		}
	}

	return nil
}

// GetConnectionsByLayer returns VPN connections grouped by layer
func (h *HierarchicalVPNManager) GetConnectionsByLayer() map[int][]*VPNConnection {
	h.mu.RLock()
	defer h.mu.RUnlock()

	layers := make(map[int][]*VPNConnection)

	for _, node := range h.hierarchy {
		layer := node.Layer
		if layers[layer] == nil {
			layers[layer] = make([]*VPNConnection, 0)
		}
		layers[layer] = append(layers[layer], node.Connection)
	}

	// Sort connections within each layer by priority
	for layer := range layers {
		sort.Slice(layers[layer], func(i, j int) bool {
			return layers[layer][i].Priority > layers[layer][j].Priority
		})
	}

	return layers
}

// GetConnectionsByEnvironment returns VPN connections filtered by network environment
func (h *HierarchicalVPNManager) GetConnectionsByEnvironment(env NetworkEnvironment) []*VPNConnection {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var result []*VPNConnection

	for _, node := range h.hierarchy {
		if h.isConnectionSuitableForEnvironment(node.Connection, env) {
			result = append(result, node.Connection)
		}
	}

	// Sort by priority
	sort.Slice(result, func(i, j int) bool {
		return result[i].Priority > result[j].Priority
	})

	return result
}

// AutoConnectForEnvironment automatically connects appropriate VPNs for network environment
func (h *HierarchicalVPNManager) AutoConnectForEnvironment(ctx context.Context, env NetworkEnvironment) error {
	connections := h.GetConnectionsByEnvironment(env)

	for _, conn := range connections {
		if conn.AutoConnect {
			if conn.Hierarchy != nil && conn.Hierarchy.ParentConnection != "" {
				// For hierarchical connections, connect the entire hierarchy
				if err := h.ConnectHierarchical(ctx, h.findRootConnection(conn.Name)); err != nil {
					return fmt.Errorf("failed to connect hierarchical VPN for %s: %w", conn.Name, err)
				}
			} else {
				// For standalone connections, connect directly
				if err := h.baseManager.ConnectVPN(ctx, conn.Name); err != nil {
					return fmt.Errorf("failed to connect VPN %s: %w", conn.Name, err)
				}
			}
		}
	}

	return nil
}

// UpdateHierarchicalRouting updates routing rules for hierarchical connections
func (h *HierarchicalVPNManager) UpdateHierarchicalRouting(ctx context.Context) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, node := range h.hierarchy {
		if node.Connection.Hierarchy != nil && len(node.Connection.Hierarchy.RoutingRules) > 0 {
			if err := h.applyHierarchicalRoutingRules(ctx, node); err != nil {
				return fmt.Errorf("failed to apply routing rules for %s: %w", node.Connection.Name, err)
			}
		}
	}

	return nil
}

// AddVPNConnection adds a VPN connection and builds hierarchy
func (h *HierarchicalVPNManager) AddVPNConnection(conn *VPNConnection) error {
	if err := h.baseManager.AddVPNConnection(conn); err != nil {
		return err
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	h.connections[conn.Name] = conn
	h.buildHierarchyNode(conn)

	return nil
}

// RemoveVPNConnection removes a VPN connection and updates hierarchy
func (h *HierarchicalVPNManager) RemoveVPNConnection(name string) error {
	if err := h.baseManager.RemoveVPNConnection(name); err != nil {
		return err
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.connections, name)
	h.removeFromHierarchy(name)

	return nil
}

// Delegate other methods to base manager
func (h *HierarchicalVPNManager) ConnectVPN(ctx context.Context, name string) error {
	return h.baseManager.ConnectVPN(ctx, name)
}

func (h *HierarchicalVPNManager) DisconnectVPN(ctx context.Context, name string) error {
	return h.baseManager.DisconnectVPN(ctx, name)
}

func (h *HierarchicalVPNManager) ConnectByPriority(ctx context.Context) error {
	return h.baseManager.ConnectByPriority(ctx)
}

func (h *HierarchicalVPNManager) GetConnectionStatus() map[string]*VPNStatus {
	return h.baseManager.GetConnectionStatus()
}

func (h *HierarchicalVPNManager) StartFailoverMonitoring(ctx context.Context) error {
	return h.baseManager.StartFailoverMonitoring(ctx)
}

func (h *HierarchicalVPNManager) StopFailoverMonitoring() {
	h.baseManager.StopFailoverMonitoring()
}

func (h *HierarchicalVPNManager) GetActiveConnections() []*VPNConnection {
	return h.baseManager.GetActiveConnections()
}

func (h *HierarchicalVPNManager) ValidateConnection(conn *VPNConnection) error {
	return h.baseManager.ValidateConnection(conn)
}

// Helper methods

func (h *HierarchicalVPNManager) buildHierarchyNode(conn *VPNConnection) {
	layer := 0
	siteType := VPNSiteTypePersonal
	var requiredEnvs []NetworkEnvironment

	if conn.Hierarchy != nil {
		layer = conn.Hierarchy.Layer
		siteType = conn.Hierarchy.SiteType
		requiredEnvs = conn.Hierarchy.RequiredEnvironments
	}

	node := &VPNHierarchyNode{
		Connection:           conn,
		Layer:                layer,
		SiteType:             siteType,
		RequiredEnvironments: requiredEnvs,
		Children:             make([]*VPNHierarchyNode, 0),
	}

	h.hierarchy[conn.Name] = node

	// Establish parent-child relationships
	if conn.Hierarchy != nil && conn.Hierarchy.ParentConnection != "" {
		if parentNode, exists := h.hierarchy[conn.Hierarchy.ParentConnection]; exists {
			node.Parent = parentNode
			parentNode.Children = append(parentNode.Children, node)
		}
	}

	// Update child relationships for existing nodes
	for _, childName := range conn.Hierarchy.ChildConnections {
		if childNode, exists := h.hierarchy[childName]; exists {
			childNode.Parent = node
			node.Children = append(node.Children, childNode)
		}
	}
}

func (h *HierarchicalVPNManager) removeFromHierarchy(name string) {
	node, exists := h.hierarchy[name]
	if !exists {
		return
	}

	// Remove from parent's children
	if node.Parent != nil {
		for i, child := range node.Parent.Children {
			if child.Connection.Name == name {
				node.Parent.Children = append(node.Parent.Children[:i], node.Parent.Children[i+1:]...)
				break
			}
		}
	}

	// Update children to remove parent reference
	for _, child := range node.Children {
		child.Parent = nil
	}

	delete(h.hierarchy, name)
}

func (h *HierarchicalVPNManager) buildConnectionOrder(root *VPNHierarchyNode) []*VPNConnection {
	var order []*VPNConnection
	visited := make(map[string]bool)

	h.buildConnectionOrderRecursive(root, &order, visited)
	return order
}

func (h *HierarchicalVPNManager) buildConnectionOrderRecursive(node *VPNHierarchyNode, order *[]*VPNConnection, visited map[string]bool) {
	if visited[node.Connection.Name] {
		return
	}

	visited[node.Connection.Name] = true
	*order = append(*order, node.Connection)

	// Sort children by layer and priority
	children := make([]*VPNHierarchyNode, len(node.Children))
	copy(children, node.Children)
	sort.Slice(children, func(i, j int) bool {
		if children[i].Layer != children[j].Layer {
			return children[i].Layer < children[j].Layer
		}
		return children[i].Connection.Priority > children[j].Connection.Priority
	})

	for _, child := range children {
		h.buildConnectionOrderRecursive(child, order, visited)
	}
}

func (h *HierarchicalVPNManager) connectWithDependencies(ctx context.Context, conn *VPNConnection) error {
	// Check if parent is connected (if hierarchical)
	if conn.Hierarchy != nil && conn.Hierarchy.ParentConnection != "" {
		status := h.baseManager.GetConnectionStatus()
		parentStatus, exists := status[conn.Hierarchy.ParentConnection]
		if !exists || parentStatus.State != VPNStateConnected {
			return fmt.Errorf("parent connection %s is not connected", conn.Hierarchy.ParentConnection)
		}
	}

	return h.baseManager.ConnectVPN(ctx, conn.Name)
}

func (h *HierarchicalVPNManager) hasCircularDependency(node *VPNHierarchyNode, visited, recursionStack map[string]bool) bool {
	visited[node.Connection.Name] = true
	recursionStack[node.Connection.Name] = true

	for _, child := range node.Children {
		if !visited[child.Connection.Name] {
			if h.hasCircularDependency(child, visited, recursionStack) {
				return true
			}
		} else if recursionStack[child.Connection.Name] {
			return true
		}
	}

	recursionStack[node.Connection.Name] = false
	return false
}

func (h *HierarchicalVPNManager) validateNodeRelationships(node *VPNHierarchyNode) error {
	conn := node.Connection

	// Validate parent connection exists
	if conn.Hierarchy != nil && conn.Hierarchy.ParentConnection != "" {
		if _, exists := h.hierarchy[conn.Hierarchy.ParentConnection]; !exists {
			return fmt.Errorf("parent connection %s not found for %s", conn.Hierarchy.ParentConnection, conn.Name)
		}
	}

	// Validate child connections exist
	if conn.Hierarchy != nil {
		for _, childName := range conn.Hierarchy.ChildConnections {
			if _, exists := h.hierarchy[childName]; !exists {
				return fmt.Errorf("child connection %s not found for %s", childName, conn.Name)
			}
		}
	}

	return nil
}

func (h *HierarchicalVPNManager) isConnectionSuitableForEnvironment(conn *VPNConnection, env NetworkEnvironment) bool {
	if conn.Hierarchy == nil || len(conn.Hierarchy.RequiredEnvironments) == 0 {
		return true // No restrictions
	}

	for _, requiredEnv := range conn.Hierarchy.RequiredEnvironments {
		if requiredEnv == env {
			return true
		}
	}

	return false
}

func (h *HierarchicalVPNManager) findRootConnection(connectionName string) string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	node, exists := h.hierarchy[connectionName]
	if !exists {
		return connectionName
	}

	// Traverse up to find root
	for node.Parent != nil {
		node = node.Parent
	}

	return node.Connection.Name
}

func (h *HierarchicalVPNManager) applyHierarchicalRoutingRules(ctx context.Context, node *VPNHierarchyNode) error {
	// This is a placeholder for routing rule implementation
	// In a real implementation, this would interact with the system's routing table

	for _, rule := range node.Connection.Hierarchy.RoutingRules {
		// Apply routing rule logic
		fmt.Printf("Applying routing rule %s for connection %s\n", rule.Name, node.Connection.Name)

		// Example: route specific destinations through parent or directly
		for _, dest := range rule.Destinations {
			if rule.RouteViaParent && node.Parent != nil {
				// Route through parent connection
				fmt.Printf("Routing %s via parent %s\n", dest, node.Parent.Connection.Name)
			} else if rule.RouteDirect {
				// Route directly
				fmt.Printf("Routing %s directly via %s\n", dest, node.Connection.Name)
			}
		}
	}

	return nil
}
