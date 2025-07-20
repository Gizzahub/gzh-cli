//nolint:testpackage // White-box testing needed for internal function access
package netenv

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSwitchOptionsDefaults(t *testing.T) {
	opts := defaultSwitchOptions()

	assert.NotEmpty(t, opts.configPath)
	assert.Contains(t, opts.configPath, "network-profiles.yaml")
	assert.False(t, opts.dryRun)
	assert.False(t, opts.verbose)
	assert.False(t, opts.force)
	assert.Empty(t, opts.profileName)
}

func TestNewSwitchCmd(t *testing.T) {
	cmd := newSwitchCmd()

	assert.Equal(t, "switch [profile-name]", cmd.Use)
	assert.Equal(t, "Switch network environment to specified profile", cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.NotNil(t, cmd.RunE)

	// Check flags
	flags := cmd.Flags()

	profile, err := flags.GetString("profile")
	assert.NoError(t, err)
	assert.Empty(t, profile)

	dryRun, err := flags.GetBool("dry-run")
	assert.NoError(t, err)
	assert.False(t, dryRun)

	verbose, err := flags.GetBool("verbose")
	assert.NoError(t, err)
	assert.False(t, verbose)

	force, err := flags.GetBool("force")
	assert.NoError(t, err)
	assert.False(t, force)

	list, err := flags.GetBool("list")
	assert.NoError(t, err)
	assert.False(t, list)

	init, err := flags.GetBool("init")
	assert.NoError(t, err)
	assert.False(t, init)
}

func TestLoadProfiles(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "network-profiles.yaml")

	testConfig := `default: "test"

profiles:
  - name: "test"
    description: "Test profile"
    dns:
      servers:
        - "1.1.1.1"
        - "1.0.0.1"
      method: "resolvectl"
    proxy:
      clear: true
`

	err := os.WriteFile(configPath, []byte(testConfig), 0o600)
	require.NoError(t, err)

	opts := &switchOptions{
		configPath: configPath,
	}

	profiles, err := opts.loadProfiles()
	require.NoError(t, err)
	assert.NotNil(t, profiles)
	assert.Equal(t, "test", profiles.Default)
	assert.Len(t, profiles.Profiles, 1)
	assert.Equal(t, "test", profiles.Profiles[0].Name)
	assert.Equal(t, "Test profile", profiles.Profiles[0].Description)
}

func TestLoadProfilesFileNotExists(t *testing.T) {
	opts := &switchOptions{
		configPath: "/nonexistent/path/network-profiles.yaml",
	}

	profiles, err := opts.loadProfiles()
	assert.Error(t, err)
	assert.Nil(t, profiles)
	assert.Contains(t, err.Error(), "configuration file not found")
}

func TestInitConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "network-profiles.yaml")

	opts := &switchOptions{
		configPath: configPath,
	}

	err := opts.initConfig()
	require.NoError(t, err)

	// Check if file was created
	_, err = os.Stat(configPath)
	assert.NoError(t, err)

	// Check file contents
	content, err := os.ReadFile(configPath)
	require.NoError(t, err)

	configStr := string(content)
	assert.Contains(t, configStr, `default: "home"`)
	assert.Contains(t, configStr, "profiles:")
	assert.Contains(t, configStr, `name: "home"`)
	assert.Contains(t, configStr, `name: "office"`)
	assert.Contains(t, configStr, `name: "cafe"`)
	assert.Contains(t, configStr, `name: "travel"`)
}

func TestInitConfigFileExists(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "network-profiles.yaml")

	// Create existing file
	err := os.WriteFile(configPath, []byte("existing"), 0o600)
	require.NoError(t, err)

	opts := &switchOptions{
		configPath: configPath,
	}

	err = opts.initConfig()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestCheckConditions(t *testing.T) {
	opts := &switchOptions{}

	// Profile with no conditions should return true
	profile := &networkProfile{
		Name: "test",
	}
	assert.True(t, opts.checkConditions(profile))

	// Profile with conditions should return true for now (TODO implementation)
	profile.Conditions = &profileConditions{
		SSID: []string{"TestWiFi"},
	}
	assert.True(t, opts.checkConditions(profile))
}

func TestExecuteScripts(t *testing.T) {
	opts := &switchOptions{
		dryRun: true, // Use dry-run to avoid actual script execution
	}

	scripts := []string{
		"echo 'test1'",
		"echo 'test2'",
	}

	err := opts.executeScripts(scripts, "test")
	assert.NoError(t, err)
}

func TestExecuteScriptsEmpty(t *testing.T) {
	opts := &switchOptions{}

	err := opts.executeScripts([]string{}, "test")
	assert.NoError(t, err)
}

func TestApplyVPNConfig(t *testing.T) {
	opts := &switchOptions{
		dryRun: true,
	}

	vpnConfig := &vpnActions{
		Connect: []vpnConfig{
			{
				Name: "test-vpn",
				Type: "networkmanager",
			},
		},
		Disconnect: []string{"old-vpn"},
	}

	err := opts.applyVPNConfig(vpnConfig)
	assert.NoError(t, err)
}

func TestApplyDNSConfig(t *testing.T) {
	opts := &switchOptions{
		dryRun: true,
	}

	dnsConfig := &dnsActions{
		Servers:   []string{"1.1.1.1", "1.0.0.1"},
		Interface: "wlan0",
	}

	err := opts.applyDNSConfig(dnsConfig)
	assert.NoError(t, err)
}

func TestApplyProxyConfig(t *testing.T) {
	opts := &switchOptions{
		dryRun: true,
	}

	// Test clearing proxy
	proxyConfig := &proxyActions{
		Clear: true,
	}

	err := opts.applyProxyConfig(proxyConfig)
	assert.NoError(t, err)

	// Test setting proxy
	proxyConfig = &proxyActions{
		HTTP:  "http://proxy.example.com:8080",
		HTTPS: "https://proxy.example.com:8080",
		SOCKS: "socks5://proxy.example.com:1080",
	}

	err = opts.applyProxyConfig(proxyConfig)
	assert.NoError(t, err)
}

func TestApplyHostsConfig(t *testing.T) {
	opts := &switchOptions{
		dryRun: true,
	}

	hostsConfig := &hostsActions{
		Add: []hostEntry{
			{IP: "192.168.1.100", Host: "test.local"},
		},
		Remove: []string{"old.local"},
	}

	err := opts.applyHostsConfig(hostsConfig)
	assert.NoError(t, err)
}
