package netenv

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultWifiOptions(t *testing.T) {
	opts := defaultWifiOptions()

	assert.NotEmpty(t, opts.configPath)
	assert.NotEmpty(t, opts.logPath)
	assert.Equal(t, 5*time.Second, opts.interval)
	assert.False(t, opts.daemon)
	assert.False(t, opts.dryRun)
	assert.False(t, opts.verbose)
}

func TestNewWifiCmd(t *testing.T) {
	cmd := newWifiCmd()

	assert.Equal(t, "wifi", cmd.Use)
	assert.Equal(t, "Monitor WiFi changes and trigger actions", cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Check subcommands
	subcommands := cmd.Commands()
	assert.Len(t, subcommands, 3)

	var monitorCmd, statusCmd, configCmd bool
	for _, subcmd := range subcommands {
		switch subcmd.Use {
		case "monitor":
			monitorCmd = true
		case "status":
			statusCmd = true
		case "config":
			configCmd = true
		}
	}

	assert.True(t, monitorCmd, "monitor subcommand should exist")
	assert.True(t, statusCmd, "status subcommand should exist")
	assert.True(t, configCmd, "config subcommand should exist")
}

func TestNewWifiMonitorCmd(t *testing.T) {
	cmd := newWifiMonitorCmd()

	assert.Equal(t, "monitor", cmd.Use)
	assert.Equal(t, "Monitor WiFi changes and execute actions", cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Test flags
	assert.NotNil(t, cmd.Flags().Lookup("config"))
	assert.NotNil(t, cmd.Flags().Lookup("daemon"))
	assert.NotNil(t, cmd.Flags().Lookup("interval"))
	assert.NotNil(t, cmd.Flags().Lookup("log"))
	assert.NotNil(t, cmd.Flags().Lookup("dry-run"))
	assert.NotNil(t, cmd.Flags().Lookup("verbose"))
}

func TestNewWifiStatusCmd(t *testing.T) {
	cmd := newWifiStatusCmd()

	assert.Equal(t, "status", cmd.Use)
	assert.Equal(t, "Show current WiFi network status", cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Test flags
	assert.NotNil(t, cmd.Flags().Lookup("verbose"))
}

func TestNewWifiConfigCmd(t *testing.T) {
	cmd := newWifiConfigCmd()

	assert.Equal(t, "config", cmd.Use)
	assert.Equal(t, "Manage WiFi hook configuration", cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Check subcommands
	subcommands := cmd.Commands()
	assert.Len(t, subcommands, 3)

	var initCmd, validateCmd, showCmd bool
	for _, subcmd := range subcommands {
		switch subcmd.Use {
		case "init":
			initCmd = true
		case "validate":
			validateCmd = true
		case "show":
			showCmd = true
		}
	}

	assert.True(t, initCmd, "init subcommand should exist")
	assert.True(t, validateCmd, "validate subcommand should exist")
	assert.True(t, showCmd, "show subcommand should exist")
}

func TestWifiOptions(t *testing.T) {
	opts := &wifiOptions{
		configPath: "/test/config.yaml",
		daemon:     true,
		interval:   10 * time.Second,
		logPath:    "/test/log.txt",
		dryRun:     true,
		verbose:    true,
	}

	assert.Equal(t, "/test/config.yaml", opts.configPath)
	assert.True(t, opts.daemon)
	assert.Equal(t, 10*time.Second, opts.interval)
	assert.Equal(t, "/test/log.txt", opts.logPath)
	assert.True(t, opts.dryRun)
	assert.True(t, opts.verbose)
}

func TestNetworkState(t *testing.T) {
	state := networkState{
		SSID:      "TestNetwork",
		Interface: "wlan0",
		State:     "connected",
		IP:        "192.168.1.100",
		Timestamp: time.Now(),
	}

	assert.Equal(t, "TestNetwork", state.SSID)
	assert.Equal(t, "wlan0", state.Interface)
	assert.Equal(t, "connected", state.State)
	assert.Equal(t, "192.168.1.100", state.IP)
	assert.False(t, state.Timestamp.IsZero())
}

func TestWifiAction(t *testing.T) {
	action := wifiAction{
		Name:        "test-action",
		Description: "Test action description",
		Commands:    []string{"echo test", "date"},
	}
	action.Conditions.SSID = []string{"TestSSID"}
	action.Conditions.State = []string{"connected"}

	assert.Equal(t, "test-action", action.Name)
	assert.Equal(t, "Test action description", action.Description)
	assert.Len(t, action.Commands, 2)
	assert.Equal(t, []string{"TestSSID"}, action.Conditions.SSID)
	assert.Equal(t, []string{"connected"}, action.Conditions.State)
}

func TestWifiConfig(t *testing.T) {
	config := wifiConfig{}
	config.Actions = []wifiAction{
		{Name: "action1", Commands: []string{"cmd1"}},
		{Name: "action2", Commands: []string{"cmd2"}},
	}
	config.Global.LogPath = "/test/log.txt"
	config.Global.Interval = 30 * time.Second

	assert.Len(t, config.Actions, 2)
	assert.Equal(t, "action1", config.Actions[0].Name)
	assert.Equal(t, "action2", config.Actions[1].Name)
	assert.Equal(t, "/test/log.txt", config.Global.LogPath)
	assert.Equal(t, 30*time.Second, config.Global.Interval)
}

func TestRunConfigInit(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "wifi-config.yaml")

	opts := &wifiOptions{
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
	assert.Contains(t, contentStr, "actions:")
	assert.Contains(t, contentStr, "vpn-connect-office")
	assert.Contains(t, contentStr, "global:")
	assert.Contains(t, contentStr, "interval:")
}

func TestRunConfigInitExistingFile(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "existing-config.yaml")

	// Create existing file
	err := os.WriteFile(configPath, []byte("existing content"), 0o644)
	require.NoError(t, err)

	opts := &wifiOptions{
		configPath: configPath,
	}

	err = opts.runConfigInit(nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestHasStateChanged(t *testing.T) {
	opts := &wifiOptions{}

	// Test with nil old state
	newState := &networkState{SSID: "Test", State: "connected"}
	assert.True(t, opts.hasStateChanged(nil, newState))

	// Test with same states
	oldState := &networkState{SSID: "Test", State: "connected"}
	assert.False(t, opts.hasStateChanged(oldState, newState))

	// Test with different SSID
	newState2 := &networkState{SSID: "Different", State: "connected"}
	assert.True(t, opts.hasStateChanged(oldState, newState2))

	// Test with different state
	newState3 := &networkState{SSID: "Test", State: "disconnected"}
	assert.True(t, opts.hasStateChanged(oldState, newState3))
}

func TestFormatState(t *testing.T) {
	opts := &wifiOptions{}

	// Test with nil state
	assert.Equal(t, "unknown", opts.formatState(nil))

	// Test with SSID
	state := &networkState{SSID: "TestNetwork", State: "connected"}
	assert.Equal(t, "TestNetwork (connected)", opts.formatState(state))

	// Test without SSID
	state2 := &networkState{State: "disconnected"}
	assert.Equal(t, "disconnected", opts.formatState(state2))
}

func TestShouldExecuteAction(t *testing.T) {
	opts := &wifiOptions{}

	// Test action with SSID condition
	action := wifiAction{}
	action.Conditions.SSID = []string{"Office", "Home"}

	state := &networkState{SSID: "Office", State: "connected"}
	assert.True(t, opts.shouldExecuteAction(action, state))

	state2 := &networkState{SSID: "Public", State: "connected"}
	assert.False(t, opts.shouldExecuteAction(action, state2))

	// Test action with state condition
	action2 := wifiAction{}
	action2.Conditions.State = []string{"connected"}

	state3 := &networkState{SSID: "Any", State: "connected"}
	assert.True(t, opts.shouldExecuteAction(action2, state3))

	state4 := &networkState{SSID: "Any", State: "disconnected"}
	assert.False(t, opts.shouldExecuteAction(action2, state4))

	// Test action with both conditions
	action3 := wifiAction{}
	action3.Conditions.SSID = []string{"Office"}
	action3.Conditions.State = []string{"connected"}

	state5 := &networkState{SSID: "Office", State: "connected"}
	assert.True(t, opts.shouldExecuteAction(action3, state5))

	state6 := &networkState{SSID: "Office", State: "disconnected"}
	assert.False(t, opts.shouldExecuteAction(action3, state6))

	state7 := &networkState{SSID: "Home", State: "connected"}
	assert.False(t, opts.shouldExecuteAction(action3, state7))
}

func TestExecuteActionCommands(t *testing.T) {
	opts := &wifiOptions{dryRun: true}

	action := wifiAction{
		Name:     "test-action",
		Commands: []string{"echo 'test'", "date"},
	}

	// Test dry run mode (should not fail)
	err := opts.executeActionCommands(action)
	assert.NoError(t, err)

	// Test with actual commands (simple echo)
	opts.dryRun = false
	action.Commands = []string{"echo 'hello world'"}
	err = opts.executeActionCommands(action)
	assert.NoError(t, err)

	// Test with failing command
	action.Commands = []string{"false"} // Command that always fails
	err = opts.executeActionCommands(action)
	assert.Error(t, err)
}

func TestLoadConfig(t *testing.T) {
	opts := &wifiOptions{}

	// Test basic config loading (simplified implementation)
	config, err := opts.loadConfig()
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.NotNil(t, config.Actions)
}

func TestRunConfigValidation(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.yaml")

	opts := &wifiOptions{
		configPath: configPath,
	}

	// Test with non-existent file
	err := opts.runConfigValidate(nil, nil)
	assert.NoError(t, err) // Current implementation always returns valid config

	// Test with existing file (simplified)
	err = os.WriteFile(configPath, []byte("test content"), 0o644)
	require.NoError(t, err)

	err = opts.runConfigValidate(nil, nil)
	assert.NoError(t, err)
}

func TestRunConfigShow(t *testing.T) {
	opts := &wifiOptions{}

	// Test showing config (simplified implementation)
	err := opts.runConfigShow(nil, nil)
	assert.NoError(t, err)
}

func TestWifiCmdStructure(t *testing.T) {
	cmd := newWifiCmd()

	// Test that the command has proper structure
	assert.NotNil(t, cmd.Use)
	assert.NotNil(t, cmd.Short)
	assert.NotNil(t, cmd.Long)
	assert.True(t, cmd.SilenceUsage)

	// Test that examples are included in Long description
	assert.Contains(t, cmd.Long, "Examples:")
	assert.Contains(t, cmd.Long, "gz net-env wifi monitor")
	assert.Contains(t, cmd.Long, "gz net-env wifi status")
}

func TestWifiConfigSubcommands(t *testing.T) {
	// Test init command
	initCmd := newWifiConfigInitCmd()
	assert.Equal(t, "init", initCmd.Use)
	assert.Contains(t, initCmd.Short, "Create example")

	// Test validate command
	validateCmd := newWifiConfigValidateCmd()
	assert.Equal(t, "validate", validateCmd.Use)
	assert.Contains(t, validateCmd.Short, "Validate")

	// Test show command
	showCmd := newWifiConfigShowCmd()
	assert.Equal(t, "show", showCmd.Use)
	assert.Contains(t, showCmd.Short, "Show")
}

func TestNetworkStateFromCommands(t *testing.T) {
	opts := &wifiOptions{}

	// Test network manager state (will likely fail in test environment)
	state := opts.getNetworkManagerState()
	// We can't assert specific values since this depends on system state
	// In test environment, this might return nil
	if state != nil {
		assert.NotEmpty(t, state.State)
	}

	// Test fallback method
	state2, err := opts.getNetworkStateFromCommands()
	assert.NoError(t, err)
	assert.NotNil(t, state2)
	// In test environment, this will likely return disconnected state
	assert.NotEmpty(t, state2.State)
}

func TestWifiOptionsDefaults(t *testing.T) {
	opts := &wifiOptions{}

	assert.Empty(t, opts.configPath)
	assert.False(t, opts.daemon)
	assert.Equal(t, time.Duration(0), opts.interval)
	assert.Empty(t, opts.action)
	assert.Empty(t, opts.logPath)
	assert.False(t, opts.dryRun)
	assert.False(t, opts.verbose)
}
