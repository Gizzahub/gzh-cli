// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package tui

import (
	"context"
	"time"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model represents the main net-env TUI application model
type Model struct {
	state       AppState
	currentView ViewType
	keymap      KeyMap
	help        help.Model
	width       int
	height      int

	// View models
	dashboardModel *DashboardModel

	// Network management
	networkStatus  NetworkStatus
	profiles       []NetworkProfile
	currentProfile string
	lastUpdate     time.Time
	updateInterval time.Duration

	// Application state
	ctx      context.Context
	quitting bool
}

// NewModel creates a new net-env TUI model
func NewModel(ctx context.Context) *Model {
	return &Model{
		state:          StateLoading,
		currentView:    ViewDashboard,
		keymap:         DefaultKeyMap,
		help:           help.New(),
		dashboardModel: NewDashboardModel(),
		networkStatus:  NetworkStatus{},
		profiles:       []NetworkProfile{},
		currentProfile: "default",
		updateInterval: 3 * time.Second, // Update network status every 3 seconds
		ctx:            ctx,
	}
}

// Init initializes the TUI application
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.refreshNetworkStatus(),
		m.startUpdateTicker(),
		tea.EnterAltScreen,
	)
}

// Update handles all messages in the TUI
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.handleGlobalKeys(msg) {
			return m, tea.Quit
		}

		// Delegate to current view
		cmd := m.updateCurrentView(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Update all view models with new size
		cmd := m.updateCurrentView(WindowSizeMsg{Width: msg.Width, Height: msg.Height})
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case TickMsg:
		// Periodic network status update
		cmds = append(cmds, m.refreshNetworkStatus())
		cmds = append(cmds, m.startUpdateTicker())

	case NetworkStatusMsg:
		m.networkStatus = msg.Status
		m.lastUpdate = time.Now()
		m.state = StateDashboard

		// Update current view with network status data
		cmd := m.updateCurrentView(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case ErrorMsg:
		m.state = StateError
		cmd := m.updateCurrentView(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case NavigationMsg:
		m.currentView = msg.View
		m.updateStateFromView()

	case ProfileSelectedMsg:
		m.currentProfile = msg.ProfileName
		// Switch to the selected profile
		cmds = append(cmds, m.switchProfile(msg.ProfileName))

	case ProfileSwitchMsg:
		if msg.Success {
			m.currentProfile = msg.ProfileName
			if msg.NewStatus != nil {
				m.networkStatus = *msg.NewStatus
			}
			// Return to dashboard after successful switch
			m.currentView = ViewDashboard
			m.state = StateDashboard
		}

	case VPNActionMsg:
		// Handle VPN action results
		if msg.Success && msg.NewStatus != nil {
			m.networkStatus.VPN = *msg.NewStatus
		}
		// Update current view with VPN action results
		cmd := m.updateCurrentView(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case RefreshMsg:
		cmds = append(cmds, m.refreshNetworkStatus())

	case QuitMsg:
		m.quitting = true
		return m, tea.Quit

	default:
		// Delegate to current view
		cmd := m.updateCurrentView(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

// View renders the current view
func (m *Model) View() string {
	if m.quitting {
		return "Goodbye! 👋\n"
	}

	switch m.currentView {
	case ViewDashboard:
		return m.dashboardModel.View()
	case ViewProfileSwitch:
		return m.renderProfileSwitch()
	case ViewVPNManager:
		return m.renderVPNManager()
	case ViewMonitoring:
		return m.renderMonitoring()
	case ViewSettings:
		return m.renderSettings()
	case ViewHelp:
		return m.renderHelp()
	case ViewSearch:
		return m.renderSearch()
	default:
		return m.dashboardModel.View()
	}
}

// handleGlobalKeys handles global keyboard shortcuts
func (m *Model) handleGlobalKeys(msg tea.KeyMsg) bool {
	switch msg.String() {
	case "ctrl+c", "q", "Q":
		if m.currentView == ViewDashboard {
			return true // Quit
		} else {
			// Navigate back to dashboard
			m.currentView = ViewDashboard
			m.state = StateDashboard
			return false
		}
	case "esc":
		if m.currentView != ViewDashboard {
			m.currentView = ViewDashboard
			m.state = StateDashboard
		}
		return false
	default:
		return false
	}
}

// updateCurrentView updates the current view with a message
func (m *Model) updateCurrentView(msg tea.Msg) tea.Cmd {
	switch m.currentView {
	case ViewDashboard:
		var cmd tea.Cmd
		m.dashboardModel, cmd = m.dashboardModel.Update(msg)
		return cmd
	case ViewProfileSwitch:
		// TODO: Implement profile switch view
		return nil
	case ViewVPNManager:
		// TODO: Implement VPN manager view
		return nil
	case ViewMonitoring:
		// TODO: Implement monitoring view
		return nil
	case ViewSettings:
		// TODO: Implement settings view
		return nil
	case ViewHelp:
		// TODO: Implement help view
		return nil
	case ViewSearch:
		// TODO: Implement search view
		return nil
	default:
		return nil
	}
}

// updateStateFromView updates the app state based on current view
func (m *Model) updateStateFromView() {
	switch m.currentView {
	case ViewDashboard:
		m.state = StateDashboard
	case ViewProfileSwitch:
		m.state = StateProfileSwitch
	case ViewVPNManager:
		m.state = StateVPNManager
	case ViewMonitoring:
		m.state = StateMonitoring
	case ViewSettings:
		m.state = StateSettings
	case ViewHelp:
		m.state = StateHelp
	case ViewSearch:
		m.state = StateSearch
	}
}

// refreshNetworkStatus refreshes the network status
func (m *Model) refreshNetworkStatus() tea.Cmd {
	return func() tea.Msg {
		// Create mock network status for now
		// In a real implementation, this would collect actual network status
		status := NetworkStatus{
			WiFi: WiFiStatus{
				SSID:           "Office WiFi",
				SignalStrength: 85,
				Frequency:      "5GHz",
				Security:       "WPA3",
				Connected:      true,
				IP:             "192.168.1.100",
				Gateway:        "192.168.1.1",
			},
			VPN: VPNStatus{
				Name:        "corp-vpn",
				Connected:   true,
				ServerIP:    "vpn.company.com",
				ClientIP:    "10.0.0.100",
				Latency:     15 * time.Millisecond,
				BytesUp:     1024 * 1024,
				BytesDown:   5 * 1024 * 1024,
				ConnectedAt: time.Now().Add(-45 * time.Minute),
				Protocol:    "OpenVPN",
			},
			DNS: DNSStatus{
				Servers:   []string{"10.0.0.1", "10.0.0.2"},
				Custom:    true,
				DoH:       false,
				DoT:       false,
				Response:  5 * time.Millisecond,
				Resolving: true,
			},
			Proxy: ProxyStatus{
				Enabled: true,
				Type:    "HTTP",
				Host:    "proxy.corp.com",
				Port:    8080,
				Auth:    true,
				Bypass:  "localhost,127.0.0.1,*.local",
				Working: true,
			},
			Docker: DockerStatus{
				Context:   "office",
				Connected: true,
				Networks: []DockerNetwork{
					{Name: "bridge", Driver: "bridge", Scope: "local", Active: true},
					{Name: "host", Driver: "host", Scope: "local", Active: false},
					{Name: "myapp-net", Driver: "bridge", Scope: "local", Active: true},
				},
			},
			Connectivity: ConnectivityStatus{
				Internet:   true,
				Latency:    25 * time.Millisecond,
				Bandwidth:  BandwidthInfo{Download: 100.5, Upload: 25.2},
				PacketLoss: 0.1,
				Quality:    "excellent",
			},
			Timestamp: time.Now(),
		}

		return NetworkStatusMsg{Status: status}
	}
}

// startUpdateTicker starts the periodic update ticker
func (m *Model) startUpdateTicker() tea.Cmd {
	return tea.Tick(m.updateInterval, func(t time.Time) tea.Msg {
		return TickMsg{Time: t}
	})
}

// switchProfile switches to a different network profile
func (m *Model) switchProfile(profileName string) tea.Cmd {
	return func() tea.Msg {
		// In a real implementation, this would:
		// 1. Load the profile configuration
		// 2. Apply network settings (VPN, DNS, proxy, etc.)
		// 3. Verify the changes
		// 4. Return success/failure

		// For now, simulate a successful profile switch
		time.Sleep(100 * time.Millisecond) // Simulate switching time

		return ProfileSwitchMsg{
			ProfileName: profileName,
			Success:     true,
			Error:       nil,
			NewStatus:   nil, // Would contain new status after switch
		}
	}
}

// Placeholder view implementations (to be implemented later)

func (m *Model) renderProfileSwitch() string {
	content := `Profile Switch View

Available Profiles:
> office        Corporate network with VPN and proxy
  home          Home network configuration
  cafe          Public WiFi with VPN protection
  mobile        Mobile hotspot configuration

Profile Details (office):
┌────────────────────────────────────────┐
│ WiFi:  Auto-detect Corporate WiFi      │
│ VPN:   corp-vpn.company.com            │
│ DNS:   10.0.0.1, 10.0.0.2              │
│ Proxy: proxy.corp.com:8080             │
│ Docker: office context                 │
└────────────────────────────────────────┘

[Enter] Apply Profile  [e] Edit  [n] New  [d] Delete  [Esc] Back`

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		InfoStyle.Render(content),
	)
}

func (m *Model) renderVPNManager() string {
	content := `VPN Connection Manager

Active Connection:
┌────────────────────────────────────────┐
│ corp-vpn                   ● Connected │
│ Server: vpn.company.com   Latency: 15ms│
│ IP: 10.0.0.100           Speed: ↑2↓5MB │
└────────────────────────────────────────┘

Available VPN Connections:
> corp-vpn      Company VPN (Active)
  backup-vpn    Backup VPN server
  client-vpn    Client network access

Connection Log:
┌────────────────────────────────────────┐
│ 14:30:15 corp-vpn connected successfully│
│ 14:25:02 Attempting connection         │
│ 14:24:58 backup-vpn disconnected       │
└────────────────────────────────────────┘

[c] Connect  [d] Disconnect  [r] Reconnect  [l] Logs  [Esc] Back`

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		InfoStyle.Render(content),
	)
}

func (m *Model) renderMonitoring() string {
	content := `Network Monitoring View

Real-time Network Metrics:
┌────────────────────────────────────────┐
│ Latency:      25ms (Excellent)         │
│ Bandwidth:    ↓100.5Mb ↑25.2Mb        │
│ Packet Loss:  0.1%                     │
│ Connections:  42 active                │
│ Quality:      Excellent                │
└────────────────────────────────────────┘

Network Traffic (Last 5 minutes):
📊 [████████████████████░░] 85%

Recent Alerts:
⚠️  High latency detected (125ms)
ℹ️  VPN reconnected successfully
✅ Network quality improved

[r] Reset Stats  [e] Export  [s] Settings  [Esc] Back`

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		InfoStyle.Render(content),
	)
}

func (m *Model) renderSettings() string {
	content := `Network Environment Settings

General Settings:
┌────────────────────────────────────────┐
│ □ Auto-detect network changes          │
│ □ Auto-connect VPN                     │
│ ☑ Show network notifications          │
│ ☑ Enable network monitoring           │
│ □ Use secure DNS (DoH/DoT)             │
└────────────────────────────────────────┘

Profile Settings:
┌────────────────────────────────────────┐
│ Default Profile:     [office     ▼]   │
│ Auto-switch:         [enabled    ▼]   │
│ Update Interval:     [3 seconds  ▼]   │
│ Backup Profiles:     [enabled    ▼]   │
└────────────────────────────────────────┘

[s] Save  [r] Reset  [e] Export Config  [Esc] Back`

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		InfoStyle.Render(content),
	)
}

func (m *Model) renderHelp() string {
	helpContent := `GZH Network Environment Manager - Help

Navigation:
  ↑/k, ↓/j     Navigate up/down
  ←/h, →/l     Navigate left/right
  Enter        Select/confirm
  Esc          Go back
  q/Q          Quit (from dashboard)

Network Actions:
  s            Switch profile
  v            VPN toggle
  d            DNS settings
  p            Proxy toggle
  c            Quick connect VPN
  x            Quick disconnect VPN

Views:
  m            Network monitoring
  P            Settings/preferences
  ?            Toggle help
  /            Search

Quick Actions:
  r            Refresh status
  1,2,3        Quick actions
  f            Filter

Press 'esc' to go back to dashboard`

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		HelpHeaderStyle.Render(helpContent),
	)
}

func (m *Model) renderSearch() string {
	content := `Search Network Components

Search: [                    ]

Results:
┌────────────────────────────────────────┐
│ 🔍 VPN Connections                     │
│   → corp-vpn (Connected)               │
│   → backup-vpn (Available)             │
│                                        │
│ 🔍 Profiles                            │
│   → office (Active)                    │
│   → home (Available)                   │
│                                        │
│ 🔍 Settings                            │
│   → DNS Configuration                  │
│   → Proxy Settings                     │
└────────────────────────────────────────┘

[Enter] Select  [Tab] Next  [Esc] Back`

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		InfoStyle.Render(content),
	)
}
