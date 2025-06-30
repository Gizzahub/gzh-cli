package net_env

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultActionsOptions(t *testing.T) {
	opts := defaultActionsOptions()

	assert.NotEmpty(t, opts.configPath)
	assert.True(t, opts.backup)
	assert.False(t, opts.dryRun)
	assert.False(t, opts.verbose)
}

func TestNewActionsCmd(t *testing.T) {
	cmd := newActionsCmd()

	assert.Equal(t, "actions", cmd.Use)
	assert.Equal(t, "Execute network configuration actions", cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Check subcommands
	subcommands := cmd.Commands()
	assert.Len(t, subcommands, 6)

	expectedCommands := map[string]bool{
		"run":    false,
		"config": false,
		"vpn":    false,
		"dns":    false,
		"proxy":  false,
		"hosts":  false,
	}

	for _, subcmd := range subcommands {
		if _, exists := expectedCommands[subcmd.Use]; exists {
			expectedCommands[subcmd.Use] = true
		}
	}

	for cmdName, found := range expectedCommands {
		assert.True(t, found, "%s subcommand should exist", cmdName)
	}
}

func TestNewActionsRunCmd(t *testing.T) {
	cmd := newActionsRunCmd()

	assert.Equal(t, "run", cmd.Use)
	assert.Equal(t, "Run network actions from configuration file", cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Test flags
	assert.NotNil(t, cmd.Flags().Lookup("config"))
	assert.NotNil(t, cmd.Flags().Lookup("dry-run"))
	assert.NotNil(t, cmd.Flags().Lookup("verbose"))
	assert.NotNil(t, cmd.Flags().Lookup("backup"))
}

func TestNewActionsConfigCmd(t *testing.T) {
	cmd := newActionsConfigCmd()

	assert.Equal(t, "config", cmd.Use)
	assert.Equal(t, "Manage network actions configuration", cmd.Short)

	// Check subcommands
	subcommands := cmd.Commands()
	assert.Len(t, subcommands, 2)

	var initCmd, validateCmd bool
	for _, subcmd := range subcommands {
		switch subcmd.Use {
		case "init":
			initCmd = true
		case "validate":
			validateCmd = true
		}
	}

	assert.True(t, initCmd, "init subcommand should exist")
	assert.True(t, validateCmd, "validate subcommand should exist")
}

func TestActionsOptions(t *testing.T) {
	opts := &actionsOptions{
		configPath: "/test/config.yaml",
		dryRun:     true,
		verbose:    true,
		backup:     false,
	}

	assert.Equal(t, "/test/config.yaml", opts.configPath)
	assert.True(t, opts.dryRun)
	assert.True(t, opts.verbose)
	assert.False(t, opts.backup)
}

func TestNetworkActions(t *testing.T) {
	actions := networkActions{
		VPN: vpnActions{
			Connect: []vpnConfig{
				{Name: "office", Type: "networkmanager"},
				{Name: "home", Type: "openvpn", ConfigFile: "/etc/openvpn/home.conf"},
			},
			Disconnect: []string{"office", "home"},
		},
		DNS: dnsActions{
			Servers:   []string{"1.1.1.1", "1.0.0.1"},
			Interface: "wlan0",
			Method:    "resolvectl",
		},
		Proxy: proxyActions{
			HTTP:    "http://proxy.company.com:8080",
			HTTPS:   "http://proxy.company.com:8080",
			NoProxy: []string{"localhost", "*.local"},
		},
		Hosts: hostsActions{
			Add: []hostEntry{
				{IP: "192.168.1.100", Host: "printer.local"},
				{IP: "10.0.0.50", Host: "dev-server.local"},
			},
			Remove: []string{"old-server.local"},
		},
	}

	// Test VPN configuration
	assert.Len(t, actions.VPN.Connect, 2)
	assert.Equal(t, "office", actions.VPN.Connect[0].Name)
	assert.Equal(t, "networkmanager", actions.VPN.Connect[0].Type)
	assert.Equal(t, "home", actions.VPN.Connect[1].Name)
	assert.Equal(t, "openvpn", actions.VPN.Connect[1].Type)
	assert.Equal(t, "/etc/openvpn/home.conf", actions.VPN.Connect[1].ConfigFile)

	// Test DNS configuration
	assert.Equal(t, []string{"1.1.1.1", "1.0.0.1"}, actions.DNS.Servers)
	assert.Equal(t, "wlan0", actions.DNS.Interface)
	assert.Equal(t, "resolvectl", actions.DNS.Method)

	// Test Proxy configuration
	assert.Equal(t, "http://proxy.company.com:8080", actions.Proxy.HTTP)
	assert.Equal(t, "http://proxy.company.com:8080", actions.Proxy.HTTPS)
	assert.Contains(t, actions.Proxy.NoProxy, "localhost")
	assert.Contains(t, actions.Proxy.NoProxy, "*.local")

	// Test Hosts configuration
	assert.Len(t, actions.Hosts.Add, 2)
	assert.Equal(t, "192.168.1.100", actions.Hosts.Add[0].IP)
	assert.Equal(t, "printer.local", actions.Hosts.Add[0].Host)
	assert.Contains(t, actions.Hosts.Remove, "old-server.local")
}

func TestVPNConfig(t *testing.T) {
	config := vpnConfig{
		Name:       "office",
		Type:       "networkmanager",
		ConfigFile: "/etc/vpn/office.conf",
		Service:    "office-vpn",
		Command:    "custom-vpn-command",
	}

	assert.Equal(t, "office", config.Name)
	assert.Equal(t, "networkmanager", config.Type)
	assert.Equal(t, "/etc/vpn/office.conf", config.ConfigFile)
	assert.Equal(t, "office-vpn", config.Service)
	assert.Equal(t, "custom-vpn-command", config.Command)
}

func TestDNSActions(t *testing.T) {
	dns := dnsActions{
		Servers:   []string{"8.8.8.8", "8.8.4.4"},
		Interface: "eth0",
		Method:    "networkmanager",
	}

	assert.Equal(t, []string{"8.8.8.8", "8.8.4.4"}, dns.Servers)
	assert.Equal(t, "eth0", dns.Interface)
	assert.Equal(t, "networkmanager", dns.Method)
}

func TestProxyActions(t *testing.T) {
	proxy := proxyActions{
		HTTP:    "http://proxy.example.com:8080",
		HTTPS:   "https://proxy.example.com:8443",
		FTP:     "ftp://proxy.example.com:21",
		SOCKS:   "socks5://proxy.example.com:1080",
		NoProxy: []string{"localhost", "127.0.0.1", "*.local"},
		Clear:   false,
	}

	assert.Equal(t, "http://proxy.example.com:8080", proxy.HTTP)
	assert.Equal(t, "https://proxy.example.com:8443", proxy.HTTPS)
	assert.Equal(t, "ftp://proxy.example.com:21", proxy.FTP)
	assert.Equal(t, "socks5://proxy.example.com:1080", proxy.SOCKS)
	assert.Len(t, proxy.NoProxy, 3)
	assert.False(t, proxy.Clear)
}

func TestHostsActions(t *testing.T) {
	hosts := hostsActions{
		Add: []hostEntry{
			{IP: "192.168.1.1", Host: "router.local"},
			{IP: "10.0.0.1", Host: "gateway.local"},
		},
		Remove: []string{"old.local", "deprecated.local"},
		Clear:  false,
	}

	assert.Len(t, hosts.Add, 2)
	assert.Equal(t, "192.168.1.1", hosts.Add[0].IP)
	assert.Equal(t, "router.local", hosts.Add[0].Host)
	assert.Len(t, hosts.Remove, 2)
	assert.Contains(t, hosts.Remove, "old.local")
	assert.False(t, hosts.Clear)
}

func TestHostEntry(t *testing.T) {
	entry := hostEntry{
		IP:   "192.168.1.100",
		Host: "myserver.local",
	}

	assert.Equal(t, "192.168.1.100", entry.IP)
	assert.Equal(t, "myserver.local", entry.Host)
}

func TestRunActionsWithDryRun(t *testing.T) {
	opts := &actionsOptions{
		dryRun:  true,
		verbose: true,
	}

	err := opts.runActions(nil, nil)
	assert.NoError(t, err)
}

func TestRunActionsWithoutDryRun(t *testing.T) {
	opts := &actionsOptions{
		dryRun:  false,
		verbose: false,
	}

	err := opts.runActions(nil, nil)
	assert.NoError(t, err)
}

func TestRunActionsConfigInit(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "actions-config.yaml")

	opts := &actionsOptions{
		configPath: configPath,
	}

	err := opts.runConfigInit(nil, nil)
	assert.NoError(t, err)

	// Check if file was created
	assert.FileExists(t, configPath)

	// Check file content
	content, err := os.ReadFile(configPath)
	require.NoError(t, err)

	contentStr := string(content)
	assert.Contains(t, contentStr, "vpn:")
	assert.Contains(t, contentStr, "dns:")
	assert.Contains(t, contentStr, "proxy:")
	assert.Contains(t, contentStr, "hosts:")
	assert.Contains(t, contentStr, "networkmanager")
	assert.Contains(t, contentStr, "resolvectl")
}

func TestRunActionsConfigInitExistingFile(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "existing-config.yaml")

	// Create existing file
	err := os.WriteFile(configPath, []byte("existing content"), 0o644)
	require.NoError(t, err)

	opts := &actionsOptions{
		configPath: configPath,
	}

	err = opts.runConfigInit(nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestRunActionsConfigValidate(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "valid-config.yaml")

	opts := &actionsOptions{
		configPath: configPath,
	}

	// Test with non-existent file
	err := opts.runConfigValidate(nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Test with existing file
	err = os.WriteFile(configPath, []byte("test content"), 0o644)
	require.NoError(t, err)

	err = opts.runConfigValidate(nil, nil)
	assert.NoError(t, err)
}

func TestVPNCommands(t *testing.T) {
	// Test VPN connect command
	connectCmd := newVPNConnectCmd()
	assert.Equal(t, "connect", connectCmd.Use)
	assert.Equal(t, "Connect to VPN", connectCmd.Short)
	assert.NotNil(t, connectCmd.Flags().Lookup("name"))
	assert.NotNil(t, connectCmd.Flags().Lookup("type"))
	assert.NotNil(t, connectCmd.Flags().Lookup("config"))

	// Test VPN disconnect command
	disconnectCmd := newVPNDisconnectCmd()
	assert.Equal(t, "disconnect", disconnectCmd.Use)
	assert.Equal(t, "Disconnect from VPN", disconnectCmd.Short)
	assert.NotNil(t, disconnectCmd.Flags().Lookup("name"))

	// Test VPN status command
	statusCmd := newVPNStatusCmd()
	assert.Equal(t, "status", statusCmd.Use)
	assert.Equal(t, "Show VPN status", statusCmd.Short)
}

func TestDNSCommands(t *testing.T) {
	// Test DNS set command
	setCmd := newDNSSetCmd()
	assert.Equal(t, "set", setCmd.Use)
	assert.Equal(t, "Set DNS servers", setCmd.Short)
	assert.NotNil(t, setCmd.Flags().Lookup("servers"))
	assert.NotNil(t, setCmd.Flags().Lookup("interface"))

	// Test DNS status command
	statusCmd := newDNSStatusCmd()
	assert.Equal(t, "status", statusCmd.Use)
	assert.Equal(t, "Show current DNS configuration", statusCmd.Short)

	// Test DNS reset command
	resetCmd := newDNSResetCmd()
	assert.Equal(t, "reset", resetCmd.Use)
	assert.Equal(t, "Reset DNS to default configuration", resetCmd.Short)
}

func TestProxyCommands(t *testing.T) {
	// Test Proxy set command
	setCmd := newProxySetCmd()
	assert.Equal(t, "set", setCmd.Use)
	assert.Equal(t, "Set proxy configuration", setCmd.Short)
	assert.NotNil(t, setCmd.Flags().Lookup("http"))
	assert.NotNil(t, setCmd.Flags().Lookup("https"))
	assert.NotNil(t, setCmd.Flags().Lookup("socks"))

	// Test Proxy clear command
	clearCmd := newProxyClearCmd()
	assert.Equal(t, "clear", clearCmd.Use)
	assert.Equal(t, "Clear proxy configuration", clearCmd.Short)

	// Test Proxy status command
	statusCmd := newProxyStatusCmd()
	assert.Equal(t, "status", statusCmd.Use)
	assert.Equal(t, "Show current proxy configuration", statusCmd.Short)
}

func TestHostsCommands(t *testing.T) {
	// Test Hosts add command
	addCmd := newHostsAddCmd()
	assert.Equal(t, "add", addCmd.Use)
	assert.Equal(t, "Add entry to hosts file", addCmd.Short)
	assert.NotNil(t, addCmd.Flags().Lookup("ip"))
	assert.NotNil(t, addCmd.Flags().Lookup("host"))

	// Test Hosts remove command
	removeCmd := newHostsRemoveCmd()
	assert.Equal(t, "remove", removeCmd.Use)
	assert.Equal(t, "Remove entry from hosts file", removeCmd.Short)
	assert.NotNil(t, removeCmd.Flags().Lookup("host"))

	// Test Hosts show command
	showCmd := newHostsShowCmd()
	assert.Equal(t, "show", showCmd.Use)
	assert.Equal(t, "Show hosts file entries", showCmd.Short)
}

func TestNewActionsVPNCmd(t *testing.T) {
	cmd := newActionsVPNCmd()

	assert.Equal(t, "vpn", cmd.Use)
	assert.Equal(t, "Manage VPN connections", cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Check subcommands
	subcommands := cmd.Commands()
	assert.Len(t, subcommands, 3)

	expectedCommands := []string{"connect", "disconnect", "status"}
	foundCommands := make(map[string]bool)

	for _, subcmd := range subcommands {
		foundCommands[subcmd.Use] = true
	}

	for _, expected := range expectedCommands {
		assert.True(t, foundCommands[expected], "%s subcommand should exist", expected)
	}
}

func TestNewActionsDNSCmd(t *testing.T) {
	cmd := newActionsDNSCmd()

	assert.Equal(t, "dns", cmd.Use)
	assert.Equal(t, "Manage DNS configuration", cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Check subcommands
	subcommands := cmd.Commands()
	assert.Len(t, subcommands, 3)

	expectedCommands := []string{"set", "status", "reset"}
	foundCommands := make(map[string]bool)

	for _, subcmd := range subcommands {
		foundCommands[subcmd.Use] = true
	}

	for _, expected := range expectedCommands {
		assert.True(t, foundCommands[expected], "%s subcommand should exist", expected)
	}
}

func TestNewActionsProxyCmd(t *testing.T) {
	cmd := newActionsProxyCmd()

	assert.Equal(t, "proxy", cmd.Use)
	assert.Equal(t, "Manage proxy configuration", cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Check subcommands
	subcommands := cmd.Commands()
	assert.Len(t, subcommands, 3)

	expectedCommands := []string{"set", "clear", "status"}
	foundCommands := make(map[string]bool)

	for _, subcmd := range subcommands {
		foundCommands[subcmd.Use] = true
	}

	for _, expected := range expectedCommands {
		assert.True(t, foundCommands[expected], "%s subcommand should exist", expected)
	}
}

func TestNewActionsHostsCmd(t *testing.T) {
	cmd := newActionsHostsCmd()

	assert.Equal(t, "hosts", cmd.Use)
	assert.Equal(t, "Manage hosts file entries", cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Check subcommands
	subcommands := cmd.Commands()
	assert.Len(t, subcommands, 3)

	expectedCommands := []string{"add", "remove", "show"}
	foundCommands := make(map[string]bool)

	for _, subcmd := range subcommands {
		foundCommands[subcmd.Use] = true
	}

	for _, expected := range expectedCommands {
		assert.True(t, foundCommands[expected], "%s subcommand should exist", expected)
	}
}

func TestActionsCmdStructure(t *testing.T) {
	cmd := newActionsCmd()

	// Test that the command has proper structure
	assert.NotNil(t, cmd.Use)
	assert.NotNil(t, cmd.Short)
	assert.NotNil(t, cmd.Long)
	assert.True(t, cmd.SilenceUsage)

	// Test that examples are included in Long description
	assert.Contains(t, cmd.Long, "Examples:")
	assert.Contains(t, cmd.Long, "gz net-env actions run")
	assert.Contains(t, cmd.Long, "gz net-env actions vpn connect")
	assert.Contains(t, cmd.Long, "gz net-env actions dns set")
}

func TestActionsOptionsDefaults(t *testing.T) {
	opts := &actionsOptions{}

	assert.Empty(t, opts.configPath)
	assert.False(t, opts.dryRun)
	assert.False(t, opts.verbose)
	assert.False(t, opts.backup)
}
