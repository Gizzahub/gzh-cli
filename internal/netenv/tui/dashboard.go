// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DashboardModel represents the main network dashboard view
type DashboardModel struct {
	table          table.Model
	help           help.Model
	keymap         KeyMap
	networkStatus  NetworkStatus
	profiles       []NetworkProfile
	currentProfile string
	lastUpdate     time.Time
	width          int
	height         int
	loading        bool
	errorMsg       string
	alerts         []NetworkAlert
}

// NewDashboardModel creates a new dashboard model
func NewDashboardModel() *DashboardModel {
	// Create table columns for network components
	columns := []table.Column{
		{Title: "Component", Width: 12},
		{Title: "Status", Width: 15},
		{Title: "Details", Width: 30},
		{Title: "Health", Width: 15},
		{Title: "", Width: 3},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(7),
	)

	// Apply table styles
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(ColorBorder).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(ColorBackground).
		Background(ColorHighlight).
		Bold(false)
	t.SetStyles(s)

	return &DashboardModel{
		table:          t,
		help:           help.New(),
		keymap:         DefaultKeyMap,
		networkStatus:  NetworkStatus{},
		profiles:       []NetworkProfile{},
		currentProfile: "default",
		lastUpdate:     time.Now(),
		loading:        true,
		alerts:         []NetworkAlert{},
	}
}

// Init initializes the dashboard model
func (m *DashboardModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the dashboard
func (m *DashboardModel) Update(msg tea.Msg) (*DashboardModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.Up):
			m.table, cmd = m.table.Update(msg)
		case key.Matches(msg, m.keymap.Down):
			m.table, cmd = m.table.Update(msg)
		case key.Matches(msg, m.keymap.Enter):
			return m, m.selectComponent()
		case key.Matches(msg, m.keymap.Refresh):
			return m, m.refreshNetworkStatus()
		case key.Matches(msg, m.keymap.SwitchProfile):
			return m, func() tea.Msg {
				return NavigationMsg{View: ViewProfileSwitch}
			}
		case key.Matches(msg, m.keymap.VPNToggle):
			return m, func() tea.Msg {
				return NavigationMsg{View: ViewVPNManager}
			}
		case key.Matches(msg, m.keymap.Monitor):
			return m, func() tea.Msg {
				return NavigationMsg{View: ViewMonitoring}
			}
		case key.Matches(msg, m.keymap.Settings):
			return m, func() tea.Msg {
				return NavigationMsg{View: ViewSettings}
			}
		case key.Matches(msg, m.keymap.Search):
			return m, func() tea.Msg {
				return NavigationMsg{View: ViewSearch}
			}
		case key.Matches(msg, m.keymap.QuickAction1):
			return m, m.handleQuickAction(1)
		case key.Matches(msg, m.keymap.QuickAction2):
			return m, m.handleQuickAction(2)
		case key.Matches(msg, m.keymap.QuickAction3):
			return m, m.handleQuickAction(3)
		case key.Matches(msg, m.keymap.QuickConnect):
			return m, m.handleQuickConnect()
		case key.Matches(msg, m.keymap.QuickDisconnect):
			return m, m.handleQuickDisconnect()
		default:
			m.table, cmd = m.table.Update(msg)
		}

	case NetworkStatusMsg:
		m.updateNetworkStatus(msg.Status)
		m.loading = false
		m.errorMsg = ""
		m.lastUpdate = time.Now()

	case ErrorMsg:
		m.loading = false
		m.errorMsg = msg.Error.Error()

	case LoadingMsg:
		m.loading = msg.Loading

	case WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateTableSize()

	case AlertMsg:
		m.addAlert(msg.Alert)

	default:
		m.table, cmd = m.table.Update(msg)
	}

	return m, cmd
}

// View renders the dashboard
func (m *DashboardModel) View() string {
	if m.loading {
		return m.renderLoading()
	}

	if m.errorMsg != "" {
		return m.renderError()
	}

	return m.renderDashboard()
}

// renderDashboard renders the main dashboard view
func (m *DashboardModel) renderDashboard() string {
	var b strings.Builder

	// Header
	header := m.renderHeader()
	b.WriteString(header)
	b.WriteString("\n")

	// Network status table
	tableView := m.table.View()
	b.WriteString(tableView)
	b.WriteString("\n")

	// Alerts section (if any)
	if len(m.alerts) > 0 {
		alertsView := m.renderAlerts()
		b.WriteString(alertsView)
		b.WriteString("\n")
	}

	// Quick actions
	quickActions := m.renderQuickActions()
	b.WriteString(quickActions)
	b.WriteString("\n")

	// Help
	helpView := m.help.View(m.keymap)
	b.WriteString(helpView)

	return b.String()
}

// renderHeader renders the dashboard header
func (m *DashboardModel) renderHeader() string {
	title := "GZH Network Environment Manager"
	profile := fmt.Sprintf("Current Profile: %s", m.currentProfile)

	// Get current network name from WiFi status
	networkName := "Unknown Network"
	if m.networkStatus.WiFi.Connected && m.networkStatus.WiFi.SSID != "" {
		networkName = m.networkStatus.WiFi.SSID
	}
	network := fmt.Sprintf("Network: %s", networkName)
	updated := fmt.Sprintf("Updated: %s", m.lastUpdate.Format("15:04:05"))

	titleStyle := TitleStyle.Width(m.width - 2).Align(lipgloss.Center)
	headerStyle := HeaderStyle.Width(m.width - 2)

	// Create header content with profile and network info
	leftContent := fmt.Sprintf("%s     %s", profile, network)
	rightContent := updated
	spacer := strings.Repeat(" ", m.width-len(leftContent)-len(rightContent)-4)

	headerContent := leftContent + spacer + rightContent

	return lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Render(title),
		headerStyle.Render(headerContent),
	)
}

// renderQuickActions renders the quick actions bar
func (m *DashboardModel) renderQuickActions() string {
	actions := []string{
		"[s]witch Profile",
		"[v]pn Toggle",
		"[d]ns Settings",
		"[p]roxy Toggle",
		"[r]efresh",
		"[q]uit",
	}

	secondRow := []string{
		"[c]onnect",
		"[x]disconnect",
		"[m]onitor",
		"[/]search",
		"[?]help",
		"[Enter] Details",
	}

	style := FooterStyle.Width(m.width - 2)

	firstLine := "Quick Actions: " + strings.Join(actions, "  ")
	secondLine := strings.Join(secondRow, "  ")

	return style.Render(lipgloss.JoinVertical(
		lipgloss.Left,
		firstLine,
		secondLine,
	))
}

// renderLoading renders the loading state
func (m *DashboardModel) renderLoading() string {
	loadingText := "Loading network environment status..."
	spinner := "⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏"
	spinnerChar := string(spinner[int(time.Now().UnixNano()/100000000)%len(spinner)])

	content := fmt.Sprintf("%s %s", spinnerChar, loadingText)

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		SpinnerStyle.Render(content),
	)
}

// renderError renders the error state
func (m *DashboardModel) renderError() string {
	errorContent := fmt.Sprintf("Error: %s\n\nPress 'r' to retry or 'q' to quit", m.errorMsg)

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		ErrorStyle.Render(errorContent),
	)
}

// renderAlerts renders the alerts section
func (m *DashboardModel) renderAlerts() string {
	if len(m.alerts) == 0 {
		return ""
	}

	var alertsText strings.Builder
	alertsText.WriteString("🚨 Alerts:\n")

	// Show only the most recent 3 alerts
	maxAlerts := 3
	if len(m.alerts) < maxAlerts {
		maxAlerts = len(m.alerts)
	}

	for i := 0; i < maxAlerts; i++ {
		alert := m.alerts[len(m.alerts)-1-i] // Show newest first
		icon := "⚠️"
		if alert.Type == "error" {
			icon = "❌"
		} else if alert.Type == "info" {
			icon = "ℹ️"
		}
		alertsText.WriteString(fmt.Sprintf("  %s %s: %s\n", icon, alert.Component, alert.Message))
	}

	style := WarningStyle.
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorWarning).
		Padding(0, 1).
		Width(m.width - 4)

	return style.Render(alertsText.String())
}

// updateNetworkStatus updates the network status and table rows
func (m *DashboardModel) updateNetworkStatus(status NetworkStatus) {
	m.networkStatus = status

	// Create table rows for each network component
	rows := []table.Row{
		m.createWiFiRow(status.WiFi),
		m.createVPNRow(status.VPN),
		m.createDNSRow(status.DNS),
		m.createProxyRow(status.Proxy),
		m.createDockerRow(status.Docker),
	}

	m.table.SetRows(rows)
}

// createWiFiRow creates a table row for WiFi status
func (m *DashboardModel) createWiFiRow(wifi WiFiStatus) table.Row {
	component := "WiFi"

	var statusText, details, health string
	if wifi.Connected {
		icon := GetStatusIcon("wifi", "connected")
		statusText = fmt.Sprintf("%s Connected", icon)
		details = fmt.Sprintf("%s (%s)", wifi.SSID, wifi.Frequency)
		if wifi.SignalStrength > 0 {
			signalIcon := GetSignalStrengthIcon(wifi.SignalStrength)
			health = fmt.Sprintf("%s %d%%", signalIcon, wifi.SignalStrength)
		} else {
			health = "Signal OK"
		}
	} else {
		icon := GetStatusIcon("wifi", "disconnected")
		statusText = fmt.Sprintf("%s Disconnected", icon)
		details = "No connection"
		health = "-"
	}

	return table.Row{component, statusText, details, health, "→"}
}

// createVPNRow creates a table row for VPN status
func (m *DashboardModel) createVPNRow(vpn VPNStatus) table.Row {
	component := "VPN"

	var statusText, details, health string
	if vpn.Connected && vpn.Name != "" {
		icon := GetStatusIcon("vpn", "connected")
		statusText = fmt.Sprintf("%s Connected", icon)
		details = fmt.Sprintf("%s (%s)", vpn.Name, vpn.ServerIP)
		if vpn.Latency > 0 {
			latencyMs := int(vpn.Latency.Milliseconds())
			health = fmt.Sprintf("%dms latency", latencyMs)
		} else {
			health = "Connected"
		}
	} else {
		icon := GetStatusIcon("vpn", "disconnected")
		statusText = fmt.Sprintf("%s Disconnected", icon)
		details = "No VPN connection"
		health = "-"
	}

	return table.Row{component, statusText, details, health, "→"}
}

// createDNSRow creates a table row for DNS status
func (m *DashboardModel) createDNSRow(dns DNSStatus) table.Row {
	component := "DNS"

	var statusText, details, health string
	if len(dns.Servers) > 0 {
		icon := GetStatusIcon("dns", "active")
		if dns.Custom {
			statusText = fmt.Sprintf("%s Custom", icon)
		} else {
			statusText = fmt.Sprintf("%s Default", icon)
		}

		// Show first 2 DNS servers
		if len(dns.Servers) > 2 {
			details = fmt.Sprintf("%s, %s (+%d more)", dns.Servers[0], dns.Servers[1], len(dns.Servers)-2)
		} else {
			details = strings.Join(dns.Servers, ", ")
		}

		if dns.Response > 0 {
			responseMs := int(dns.Response.Milliseconds())
			health = fmt.Sprintf("<%dms response", responseMs)
		} else {
			health = "Resolving"
		}
	} else {
		icon := GetStatusIcon("dns", "inactive")
		statusText = fmt.Sprintf("%s No DNS", icon)
		details = "DNS not configured"
		health = "-"
	}

	return table.Row{component, statusText, details, health, "→"}
}

// createProxyRow creates a table row for proxy status
func (m *DashboardModel) createProxyRow(proxy ProxyStatus) table.Row {
	component := "Proxy"

	var statusText, details, health string
	if proxy.Enabled {
		icon := GetStatusIcon("proxy", "enabled")
		statusText = fmt.Sprintf("%s Enabled", icon)
		details = fmt.Sprintf("%s:%d (%s)", proxy.Host, proxy.Port, proxy.Type)
		if proxy.Working {
			health = "Connected"
		} else {
			health = "Not responding"
		}
	} else {
		icon := GetStatusIcon("proxy", "disabled")
		statusText = fmt.Sprintf("%s Disabled", icon)
		details = "Direct connection"
		health = "-"
	}

	return table.Row{component, statusText, details, health, "→"}
}

// createDockerRow creates a table row for Docker status
func (m *DashboardModel) createDockerRow(docker DockerStatus) table.Row {
	component := "Docker"

	var statusText, details, health string
	if docker.Connected {
		icon := GetStatusIcon("docker", "connected")
		statusText = fmt.Sprintf("%s Connected", icon)
		details = fmt.Sprintf("%s context", docker.Context)
		health = fmt.Sprintf("%d networks", len(docker.Networks))
	} else {
		icon := GetStatusIcon("docker", "disconnected")
		statusText = fmt.Sprintf("%s Disconnected", icon)
		details = "Docker not available"
		health = "-"
	}

	return table.Row{component, statusText, details, health, "→"}
}

// updateTableSize updates the table size based on terminal dimensions
func (m *DashboardModel) updateTableSize() {
	if m.width < 80 {
		// Adjust column widths for smaller terminals
		columns := []table.Column{
			{Title: "Component", Width: 10},
			{Title: "Status", Width: 12},
			{Title: "Details", Width: 25},
			{Title: "Health", Width: 12},
			{Title: "", Width: 2},
		}
		m.table.SetColumns(columns)
	}

	// Adjust table height
	availableHeight := m.height - 12 // Reserve space for header, footer, help, alerts
	if availableHeight < 5 {
		availableHeight = 5
	}
	if availableHeight > 10 {
		availableHeight = 10
	}

	m.table.SetHeight(availableHeight)
}

// selectComponent handles component selection
func (m *DashboardModel) selectComponent() tea.Cmd {
	selectedRow := m.table.SelectedRow()
	if selectedRow == nil {
		return nil
	}

	component := selectedRow[0]

	// Navigate to appropriate view based on component
	switch strings.ToLower(component) {
	case "vpn":
		return func() tea.Msg {
			return NavigationMsg{View: ViewVPNManager}
		}
	case ComponentWiFi, ComponentDNS, ComponentProxy, ComponentDocker:
		return func() tea.Msg {
			return NavigationMsg{View: ViewSettings, Data: component}
		}
	}

	return nil
}

// refreshNetworkStatus triggers a network status refresh
func (m *DashboardModel) refreshNetworkStatus() tea.Cmd {
	return func() tea.Msg {
		return RefreshMsg{}
	}
}

// handleQuickAction handles quick action buttons
func (m *DashboardModel) handleQuickAction(action int) tea.Cmd {
	switch action {
	case 1: // Switch Profile
		return func() tea.Msg {
			return NavigationMsg{View: ViewProfileSwitch}
		}
	case 2: // Refresh Status
		return m.refreshNetworkStatus()
	case 3: // Monitor
		return func() tea.Msg {
			return NavigationMsg{View: ViewMonitoring}
		}
	default:
		return nil
	}
}

// handleQuickConnect handles quick connect action
func (m *DashboardModel) handleQuickConnect() tea.Cmd {
	// Find the most appropriate VPN to connect
	return func() tea.Msg {
		// This would trigger VPN connection logic
		return VPNActionMsg{
			Action:  "connect",
			VPNName: "auto", // Auto-select best VPN
		}
	}
}

// handleQuickDisconnect handles quick disconnect action
func (m *DashboardModel) handleQuickDisconnect() tea.Cmd {
	return func() tea.Msg {
		// This would trigger VPN disconnection logic
		return VPNActionMsg{
			Action:  "disconnect",
			VPNName: m.networkStatus.VPN.Name,
		}
	}
}

// addAlert adds a new alert to the alerts list
func (m *DashboardModel) addAlert(alert NetworkAlert) {
	m.alerts = append(m.alerts, alert)

	// Keep only the last 10 alerts
	if len(m.alerts) > 10 {
		m.alerts = m.alerts[len(m.alerts)-10:]
	}
}
