package cloud

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockVPNManager for testing
type MockVPNManager struct {
	mock.Mock
	connections map[string]*VPNConnection
	status      map[string]*VPNStatus
}

func NewMockVPNManager() *MockVPNManager {
	return &MockVPNManager{
		connections: make(map[string]*VPNConnection),
		status:      make(map[string]*VPNStatus),
	}
}

func (m *MockVPNManager) AddVPNConnection(conn *VPNConnection) error {
	args := m.Called(conn)
	if args.Error(0) == nil {
		m.connections[conn.Name] = conn
	}
	return args.Error(0)
}

func (m *MockVPNManager) RemoveVPNConnection(name string) error {
	args := m.Called(name)
	if args.Error(0) == nil {
		delete(m.connections, name)
	}
	return args.Error(0)
}

func (m *MockVPNManager) ConnectVPN(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	if args.Error(0) == nil {
		m.status[name] = &VPNStatus{
			Name:        name,
			State:       VPNStateConnected,
			ConnectedAt: time.Now(),
		}
	}
	return args.Error(0)
}

func (m *MockVPNManager) DisconnectVPN(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	if args.Error(0) == nil {
		if status, exists := m.status[name]; exists {
			status.State = VPNStateDisconnected
			status.DisconnectedAt = time.Now()
		}
	}
	return args.Error(0)
}

func (m *MockVPNManager) ConnectByPriority(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockVPNManager) GetConnectionStatus() map[string]*VPNStatus {
	args := m.Called()
	if args.Get(0) != nil {
		return args.Get(0).(map[string]*VPNStatus)
	}
	return m.status
}

func (m *MockVPNManager) StartFailoverMonitoring(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockVPNManager) StopFailoverMonitoring() {
	m.Called()
}

func (m *MockVPNManager) GetActiveConnections() []*VPNConnection {
	args := m.Called()
	if args.Get(0) != nil {
		return args.Get(0).([]*VPNConnection)
	}

	var active []*VPNConnection
	for name, status := range m.status {
		if status.State == VPNStateConnected {
			if conn, exists := m.connections[name]; exists {
				active = append(active, conn)
			}
		}
	}
	return active
}

func (m *MockVPNManager) ValidateConnection(conn *VPNConnection) error {
	args := m.Called(conn)
	return args.Error(0)
}

// Hierarchical VPN Management Methods for Mock

func (m *MockVPNManager) ConnectHierarchical(ctx context.Context, rootConnection string) error {
	args := m.Called(ctx, rootConnection)
	return args.Error(0)
}

func (m *MockVPNManager) DisconnectHierarchical(ctx context.Context, rootConnection string) error {
	args := m.Called(ctx, rootConnection)
	return args.Error(0)
}

func (m *MockVPNManager) GetVPNHierarchy() map[string]*VPNHierarchyNode {
	args := m.Called()
	if args.Get(0) != nil {
		return args.Get(0).(map[string]*VPNHierarchyNode)
	}
	return make(map[string]*VPNHierarchyNode)
}

func (m *MockVPNManager) ValidateHierarchy() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockVPNManager) GetConnectionsByLayer() map[int][]*VPNConnection {
	args := m.Called()
	if args.Get(0) != nil {
		return args.Get(0).(map[int][]*VPNConnection)
	}
	return make(map[int][]*VPNConnection)
}

func (m *MockVPNManager) GetConnectionsByEnvironment(env NetworkEnvironment) []*VPNConnection {
	args := m.Called(env)
	if args.Get(0) != nil {
		return args.Get(0).([]*VPNConnection)
	}
	return []*VPNConnection{}
}

func (m *MockVPNManager) AutoConnectForEnvironment(ctx context.Context, env NetworkEnvironment) error {
	args := m.Called(ctx, env)
	return args.Error(0)
}

func (m *MockVPNManager) UpdateHierarchicalRouting(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestHierarchicalVPNManager_AddVPNConnection(t *testing.T) {
	mockManager := NewMockVPNManager()
	mockManager.On("AddVPNConnection", mock.Anything).Return(nil)

	hierarchicalManager := NewHierarchicalVPNManager(mockManager)

	conn := &VPNConnection{
		Name:     "test-vpn",
		Type:     "openvpn",
		Server:   "vpn.example.com",
		Priority: 100,
		Hierarchy: &VPNHierarchy{
			Layer:    0,
			Mode:     VPNHierarchyModeStandalone,
			SiteType: VPNSiteTypePersonal,
		},
	}

	err := hierarchicalManager.AddVPNConnection(conn)
	assert.NoError(t, err)

	hierarchy := hierarchicalManager.GetVPNHierarchy()
	assert.Len(t, hierarchy, 1)
	assert.Contains(t, hierarchy, "test-vpn")

	node := hierarchy["test-vpn"]
	assert.Equal(t, "test-vpn", node.Connection.Name)
	assert.Equal(t, 0, node.Layer)
	assert.Equal(t, VPNSiteTypePersonal, node.SiteType)
}

func TestHierarchicalVPNManager_HierarchyConstruction(t *testing.T) {
	mockManager := NewMockVPNManager()
	mockManager.On("AddVPNConnection", mock.Anything).Return(nil)

	hierarchicalManager := NewHierarchicalVPNManager(mockManager)

	// Add parent VPN
	parentConn := &VPNConnection{
		Name:     "corporate-vpn",
		Type:     "openvpn",
		Server:   "corp.example.com",
		Priority: 200,
		Hierarchy: &VPNHierarchy{
			Layer:            0,
			Mode:             VPNHierarchyModeStandalone,
			SiteType:         VPNSiteTypeCorporate,
			ChildConnections: []string{"personal-vpn"},
		},
	}

	// Add child VPN
	childConn := &VPNConnection{
		Name:     "personal-vpn",
		Type:     "wireguard",
		Server:   "personal.example.com",
		Priority: 100,
		Hierarchy: &VPNHierarchy{
			Layer:            1,
			Mode:             VPNHierarchyModeChained,
			SiteType:         VPNSiteTypePersonal,
			ParentConnection: "corporate-vpn",
		},
	}

	err := hierarchicalManager.AddVPNConnection(parentConn)
	assert.NoError(t, err)

	err = hierarchicalManager.AddVPNConnection(childConn)
	assert.NoError(t, err)

	hierarchy := hierarchicalManager.GetVPNHierarchy()
	assert.Len(t, hierarchy, 2)

	parentNode := hierarchy["corporate-vpn"]
	childNode := hierarchy["personal-vpn"]

	assert.NotNil(t, parentNode)
	assert.NotNil(t, childNode)

	// Check parent-child relationships
	assert.Len(t, parentNode.Children, 1)
	assert.Equal(t, "personal-vpn", parentNode.Children[0].Connection.Name)
	assert.Equal(t, parentNode, childNode.Parent)
}

func TestHierarchicalVPNManager_ConnectHierarchical(t *testing.T) {
	mockManager := NewMockVPNManager()
	mockManager.On("AddVPNConnection", mock.Anything).Return(nil)
	mockManager.On("ConnectVPN", mock.Anything, "corporate-vpn").Return(nil)
	mockManager.On("ConnectVPN", mock.Anything, "personal-vpn").Return(nil)
	mockManager.On("GetConnectionStatus").Return(map[string]*VPNStatus{
		"corporate-vpn": {Name: "corporate-vpn", State: VPNStateConnected},
	})

	hierarchicalManager := NewHierarchicalVPNManager(mockManager)

	// Setup hierarchy
	parentConn := &VPNConnection{
		Name:     "corporate-vpn",
		Type:     "openvpn",
		Server:   "corp.example.com",
		Priority: 200,
		Hierarchy: &VPNHierarchy{
			Layer:            0,
			Mode:             VPNHierarchyModeStandalone,
			SiteType:         VPNSiteTypeCorporate,
			ChildConnections: []string{"personal-vpn"},
		},
	}

	childConn := &VPNConnection{
		Name:     "personal-vpn",
		Type:     "wireguard",
		Server:   "personal.example.com",
		Priority: 100,
		Hierarchy: &VPNHierarchy{
			Layer:            1,
			Mode:             VPNHierarchyModeChained,
			SiteType:         VPNSiteTypePersonal,
			ParentConnection: "corporate-vpn",
		},
	}

	hierarchicalManager.AddVPNConnection(parentConn)
	hierarchicalManager.AddVPNConnection(childConn)

	ctx := context.Background()
	err := hierarchicalManager.ConnectHierarchical(ctx, "corporate-vpn")
	assert.NoError(t, err)

	// Verify both connections were attempted
	mockManager.AssertCalled(t, "ConnectVPN", ctx, "corporate-vpn")
	mockManager.AssertCalled(t, "ConnectVPN", ctx, "personal-vpn")
}

func TestHierarchicalVPNManager_GetConnectionsByLayer(t *testing.T) {
	mockManager := NewMockVPNManager()
	mockManager.On("AddVPNConnection", mock.Anything).Return(nil)

	hierarchicalManager := NewHierarchicalVPNManager(mockManager)

	// Add connections at different layers
	layer0Conn := &VPNConnection{
		Name:     "layer0-vpn",
		Type:     "openvpn",
		Server:   "layer0.example.com",
		Priority: 200,
		Hierarchy: &VPNHierarchy{
			Layer:    0,
			Mode:     VPNHierarchyModeStandalone,
			SiteType: VPNSiteTypeCorporate,
		},
	}

	layer1Conn := &VPNConnection{
		Name:     "layer1-vpn",
		Type:     "wireguard",
		Server:   "layer1.example.com",
		Priority: 100,
		Hierarchy: &VPNHierarchy{
			Layer:    1,
			Mode:     VPNHierarchyModeChained,
			SiteType: VPNSiteTypePersonal,
		},
	}

	hierarchicalManager.AddVPNConnection(layer0Conn)
	hierarchicalManager.AddVPNConnection(layer1Conn)

	layers := hierarchicalManager.GetConnectionsByLayer()

	assert.Len(t, layers, 2)
	assert.Len(t, layers[0], 1)
	assert.Len(t, layers[1], 1)
	assert.Equal(t, "layer0-vpn", layers[0][0].Name)
	assert.Equal(t, "layer1-vpn", layers[1][0].Name)
}

func TestHierarchicalVPNManager_GetConnectionsByEnvironment(t *testing.T) {
	mockManager := NewMockVPNManager()
	mockManager.On("AddVPNConnection", mock.Anything).Return(nil)

	hierarchicalManager := NewHierarchicalVPNManager(mockManager)

	// Add connection with environment restrictions
	homeConn := &VPNConnection{
		Name:     "home-vpn",
		Type:     "wireguard",
		Server:   "home.example.com",
		Priority: 100,
		Hierarchy: &VPNHierarchy{
			Layer:                0,
			Mode:                 VPNHierarchyModeStandalone,
			SiteType:             VPNSiteTypePersonal,
			RequiredEnvironments: []NetworkEnvironment{NetworkEnvironmentHome},
		},
	}

	officeConn := &VPNConnection{
		Name:     "office-vpn",
		Type:     "openvpn",
		Server:   "office.example.com",
		Priority: 200,
		Hierarchy: &VPNHierarchy{
			Layer:                0,
			Mode:                 VPNHierarchyModeStandalone,
			SiteType:             VPNSiteTypeCorporate,
			RequiredEnvironments: []NetworkEnvironment{NetworkEnvironmentOffice},
		},
	}

	hierarchicalManager.AddVPNConnection(homeConn)
	hierarchicalManager.AddVPNConnection(officeConn)

	homeConnections := hierarchicalManager.GetConnectionsByEnvironment(NetworkEnvironmentHome)
	officeConnections := hierarchicalManager.GetConnectionsByEnvironment(NetworkEnvironmentOffice)

	assert.Len(t, homeConnections, 1)
	assert.Len(t, officeConnections, 1)
	assert.Equal(t, "home-vpn", homeConnections[0].Name)
	assert.Equal(t, "office-vpn", officeConnections[0].Name)
}

func TestHierarchicalVPNManager_ValidateHierarchy(t *testing.T) {
	mockManager := NewMockVPNManager()
	mockManager.On("AddVPNConnection", mock.Anything).Return(nil)

	hierarchicalManager := NewHierarchicalVPNManager(mockManager)

	// Add valid hierarchy
	parentConn := &VPNConnection{
		Name:     "parent-vpn",
		Type:     "openvpn",
		Server:   "parent.example.com",
		Priority: 200,
		Hierarchy: &VPNHierarchy{
			Layer:            0,
			Mode:             VPNHierarchyModeStandalone,
			SiteType:         VPNSiteTypeCorporate,
			ChildConnections: []string{"child-vpn"},
		},
	}

	childConn := &VPNConnection{
		Name:     "child-vpn",
		Type:     "wireguard",
		Server:   "child.example.com",
		Priority: 100,
		Hierarchy: &VPNHierarchy{
			Layer:            1,
			Mode:             VPNHierarchyModeChained,
			SiteType:         VPNSiteTypePersonal,
			ParentConnection: "parent-vpn",
		},
	}

	hierarchicalManager.AddVPNConnection(parentConn)
	hierarchicalManager.AddVPNConnection(childConn)

	err := hierarchicalManager.ValidateHierarchy()
	assert.NoError(t, err)
}

func TestHierarchicalVPNManager_ValidateHierarchy_CircularDependency(t *testing.T) {
	mockManager := NewMockVPNManager()
	mockManager.On("AddVPNConnection", mock.Anything).Return(nil)

	hierarchicalManager := NewHierarchicalVPNManager(mockManager)

	// Add circular dependency
	conn1 := &VPNConnection{
		Name:     "vpn1",
		Type:     "openvpn",
		Server:   "vpn1.example.com",
		Priority: 100,
		Hierarchy: &VPNHierarchy{
			Layer:            0,
			Mode:             VPNHierarchyModeChained,
			SiteType:         VPNSiteTypePersonal,
			ParentConnection: "vpn2",
			ChildConnections: []string{"vpn2"},
		},
	}

	conn2 := &VPNConnection{
		Name:     "vpn2",
		Type:     "wireguard",
		Server:   "vpn2.example.com",
		Priority: 100,
		Hierarchy: &VPNHierarchy{
			Layer:            1,
			Mode:             VPNHierarchyModeChained,
			SiteType:         VPNSiteTypePersonal,
			ParentConnection: "vpn1",
			ChildConnections: []string{"vpn1"},
		},
	}

	hierarchicalManager.AddVPNConnection(conn1)
	hierarchicalManager.AddVPNConnection(conn2)

	err := hierarchicalManager.ValidateHierarchy()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circular dependency")
}

func TestHierarchicalVPNManager_RemoveVPNConnection(t *testing.T) {
	mockManager := NewMockVPNManager()
	mockManager.On("AddVPNConnection", mock.Anything).Return(nil)
	mockManager.On("RemoveVPNConnection", "test-vpn").Return(nil)

	hierarchicalManager := NewHierarchicalVPNManager(mockManager)

	conn := &VPNConnection{
		Name:     "test-vpn",
		Type:     "openvpn",
		Server:   "test.example.com",
		Priority: 100,
		Hierarchy: &VPNHierarchy{
			Layer:    0,
			Mode:     VPNHierarchyModeStandalone,
			SiteType: VPNSiteTypePersonal,
		},
	}

	// Add then remove
	hierarchicalManager.AddVPNConnection(conn)

	hierarchy := hierarchicalManager.GetVPNHierarchy()
	assert.Len(t, hierarchy, 1)

	err := hierarchicalManager.RemoveVPNConnection("test-vpn")
	assert.NoError(t, err)

	hierarchy = hierarchicalManager.GetVPNHierarchy()
	assert.Len(t, hierarchy, 0)
}
