//nolint:testpackage // White-box testing needed for internal function access
package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultUnifiedConfig(t *testing.T) {
	config := DefaultUnifiedConfig()

	// Test basic structure
	assert.Equal(t, "1.0.0", config.Version)
	assert.Equal(t, "github", config.DefaultProvider)
	assert.NotNil(t, config.Global)
	assert.NotNil(t, config.Providers)
	assert.NotNil(t, config.IDE)
	assert.NotNil(t, config.DevEnv)
	assert.NotNil(t, config.NetEnv)
	assert.NotNil(t, config.SSHConfig)

	// Test global settings
	assert.Equal(t, "$HOME/repos", config.Global.CloneBaseDir)
	assert.Equal(t, "reset", config.Global.DefaultStrategy)
	assert.Equal(t, "all", config.Global.DefaultVisibility)

	// Test timeouts
	assert.Equal(t, 30*time.Second, config.Global.Timeouts.HTTPTimeout)
	assert.Equal(t, 5*time.Minute, config.Global.Timeouts.GitTimeout)
	assert.Equal(t, 1*time.Hour, config.Global.Timeouts.RateLimitTimeout)

	// Test concurrency settings
	assert.Equal(t, 10, config.Global.Concurrency.CloneWorkers)
	assert.Equal(t, 15, config.Global.Concurrency.UpdateWorkers)
	assert.Equal(t, 5, config.Global.Concurrency.APIWorkers)

	// Test IDE configuration
	assert.True(t, config.IDE.Enabled)
	assert.True(t, config.IDE.AutoFixSync)
	assert.True(t, config.IDE.SyncSettings.Enabled)
	assert.Equal(t, 5*time.Minute, config.IDE.SyncSettings.Interval)
	assert.Contains(t, config.IDE.JetBrainsProducts, "IntelliJ")
	assert.Contains(t, config.IDE.JetBrainsProducts, "PyCharm")
	assert.Contains(t, config.IDE.JetBrainsProducts, "GoLand")

	// Test DevEnv configuration
	assert.True(t, config.DevEnv.Enabled)
	assert.Equal(t, "$HOME/.gz/backups", config.DevEnv.BackupLocation)
	assert.True(t, config.DevEnv.AutoBackup)
	assert.NotNil(t, config.DevEnv.Providers.AWS)
	assert.NotNil(t, config.DevEnv.Providers.GCP)
	assert.NotNil(t, config.DevEnv.Providers.Azure)

	// Test NetEnv configuration
	assert.True(t, config.NetEnv.Enabled)
	assert.True(t, config.NetEnv.WiFiMonitoring.Enabled)
	assert.Equal(t, 5*time.Second, config.NetEnv.WiFiMonitoring.Interval)
	assert.Contains(t, config.NetEnv.DNS.DefaultServers, "1.1.1.1")
	assert.Contains(t, config.NetEnv.DNS.DefaultServers, "1.0.0.1")

	// Test SSH configuration
	assert.True(t, config.SSHConfig.Enabled)
	assert.Equal(t, "$HOME/.ssh/config", config.SSHConfig.ConfigFile)
	assert.True(t, config.SSHConfig.BackupEnabled)
	assert.Equal(t, "$HOME/.ssh/backups", config.SSHConfig.BackupDir)
}

func TestSupportedProviders(t *testing.T) {
	providers := SupportedProviders()
	expected := []string{"github", "gitlab", "gitea", "gogs"}

	assert.Equal(t, expected, providers)
}

func TestValidateProvider(t *testing.T) {
	tests := []struct {
		provider string
		expected bool
	}{
		{"github", true},
		{"gitlab", true},
		{"gitea", true},
		{"gogs", true},
		{"invalid", false},
		{"", false},
		{"GITHUB", false}, // case sensitive
	}

	for _, test := range tests {
		t.Run(test.provider, func(t *testing.T) {
			result := ValidateProvider(test.provider)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestIDEConfig(t *testing.T) {
	config := &IDEConfig{
		Enabled:           true,
		WatchDirectories:  []string{"$HOME/.config", "$HOME/.local/share/JetBrains"},
		ExcludePatterns:   []string{`\.git/.*`, `node_modules/.*`},
		JetBrainsProducts: []string{"IntelliJ", "PyCharm"},
		AutoFixSync:       true,
		SyncSettings: &IDESyncSettings{
			Enabled:          true,
			Interval:         5 * time.Minute,
			SyncTypes:        []string{"keymap", "editor"},
			BackupBeforeSync: true,
		},
		Logging: &IDELoggingConfig{
			Level:    "info",
			FilePath: "$HOME/.local/share/gzh-manager/logs/ide.log",
			Console:  true,
			Rotation: &LogRotationConfig{
				MaxSizeMB:  10,
				MaxBackups: 5,
				MaxAgeDays: 30,
				Compress:   true,
			},
		},
	}

	assert.True(t, config.Enabled)
	assert.Contains(t, config.WatchDirectories, "$HOME/.config")
	assert.Contains(t, config.ExcludePatterns, `\.git/.*`)
	assert.Contains(t, config.JetBrainsProducts, "IntelliJ")
	assert.True(t, config.AutoFixSync)
	assert.True(t, config.SyncSettings.Enabled)
	assert.Equal(t, 5*time.Minute, config.SyncSettings.Interval)
	assert.Equal(t, "info", config.Logging.Level)
	assert.True(t, config.Logging.Console)
	assert.Equal(t, 10, config.Logging.Rotation.MaxSizeMB)
}

func TestDevEnvConfig(t *testing.T) {
	config := &DevEnvConfig{
		Enabled:        true,
		BackupLocation: "$HOME/.gz/backups",
		AutoBackup:     true,
		Providers: &DevEnvProviders{
			AWS: &AWSConfig{
				DefaultProfile:   "default",
				PreferredRegions: []string{"us-west-2", "us-east-1"},
				CredentialsFile:  "$HOME/.aws/credentials",
				ConfigFile:       "$HOME/.aws/config",
				EnableMFA:        false,
			},
			GCP: &GCPConfig{
				DefaultProject:   "my-project",
				PreferredRegions: []string{"us-central1", "us-west1"},
				UseADC:           true,
			},
			Azure: &AzureConfig{
				DefaultSubscription: "my-subscription",
				DefaultTenant:       "my-tenant",
				PreferredRegions:    []string{"westus2", "eastus"},
				UseManagedIdentity:  false,
			},
		},
		Containers: &ContainerConfig{
			DefaultRuntime: "docker",
			Docker: &DockerConfig{
				SocketPath:      "/var/run/docker.sock",
				DefaultRegistry: "docker.io",
				BuildOptions: &DockerBuildOptions{
					DefaultContext: ".",
					EnableBuildKit: true,
				},
			},
		},
		Kubernetes: &KubernetesConfig{
			KubeconfigPath:   "$HOME/.kube/config",
			DefaultNamespace: "default",
			AutoDiscovery:    true,
		},
	}

	assert.True(t, config.Enabled)
	assert.Equal(t, "$HOME/.gz/backups", config.BackupLocation)
	assert.True(t, config.AutoBackup)
	assert.Equal(t, "default", config.Providers.AWS.DefaultProfile)
	assert.Contains(t, config.Providers.AWS.PreferredRegions, "us-west-2")
	assert.Equal(t, "my-project", config.Providers.GCP.DefaultProject)
	assert.True(t, config.Providers.GCP.UseADC)
	assert.Equal(t, "my-subscription", config.Providers.Azure.DefaultSubscription)
	assert.Equal(t, "docker", config.Containers.DefaultRuntime)
	assert.Equal(t, "/var/run/docker.sock", config.Containers.Docker.SocketPath)
	assert.True(t, config.Containers.Docker.BuildOptions.EnableBuildKit)
	assert.Equal(t, "$HOME/.kube/config", config.Kubernetes.KubeconfigPath)
	assert.Equal(t, "default", config.Kubernetes.DefaultNamespace)
	assert.True(t, config.Kubernetes.AutoDiscovery)
}

func TestNetEnvConfig(t *testing.T) {
	config := &NetEnvConfig{
		Enabled: true,
		WiFiMonitoring: &WiFiMonitoringConfig{
			Enabled:  true,
			Interval: 5 * time.Second,
			KnownNetworks: map[string]*NetworkProfile{
				"Home-WiFi": {
					SSID:         "Home-WiFi",
					Type:         "home",
					DNSServers:   []string{"192.168.1.1"},
					OnConnect:    []string{"sync-time"},
					OnDisconnect: []string{"pause-sync"},
				},
			},
		},
		VPN: &VPNConfig{
			Profiles: map[string]*VPNProfile{
				"work-vpn": {
					Type:              "openvpn",
					ConfigFile:        "$HOME/.config/vpn/work.ovpn",
					ConnectCommand:    "openvpn --config $HOME/.config/vpn/work.ovpn",
					DisconnectCommand: "pkill openvpn",
				},
			},
			DefaultProfile: "work-vpn",
			AutoConnect: &VPNAutoConnect{
				Enabled:             true,
				OnUntrustedNetworks: true,
				TrustedNetworks:     []string{"Home-WiFi"},
				RetryAttempts:       3,
				RetryDelay:          5 * time.Second,
			},
		},
		DNS: &DNSConfig{
			DefaultServers: []string{"1.1.1.1", "1.0.0.1"},
			EnableDoH:      false,
			DoHProvider:    "cloudflare",
		},
		Daemon: &DaemonConfig{
			Enabled:            false,
			PIDFile:            "/var/run/gzh-manager-netenv.pid",
			LogFile:            "/var/log/gzh-manager-netenv.log",
			LogLevel:           "info",
			WorkingDir:         "/",
			SystemdIntegration: true,
		},
	}

	assert.True(t, config.Enabled)
	assert.True(t, config.WiFiMonitoring.Enabled)
	assert.Equal(t, 5*time.Second, config.WiFiMonitoring.Interval)
	assert.Contains(t, config.WiFiMonitoring.KnownNetworks, "Home-WiFi")
	assert.Equal(t, "home", config.WiFiMonitoring.KnownNetworks["Home-WiFi"].Type)
	assert.Contains(t, config.VPN.Profiles, "work-vpn")
	assert.Equal(t, "openvpn", config.VPN.Profiles["work-vpn"].Type)
	assert.Equal(t, "work-vpn", config.VPN.DefaultProfile)
	assert.True(t, config.VPN.AutoConnect.Enabled)
	assert.Contains(t, config.DNS.DefaultServers, "1.1.1.1")
	assert.Equal(t, "cloudflare", config.DNS.DoHProvider)
	assert.False(t, config.Daemon.Enabled)
	assert.Equal(t, "info", config.Daemon.LogLevel)
}

func TestSSHConfigSettings(t *testing.T) {
	config := &SSHConfigSettings{
		Enabled:       true,
		ConfigFile:    "$HOME/.ssh/config",
		BackupEnabled: true,
		BackupDir:     "$HOME/.ssh/backups",
		ProviderConfigs: map[string]*SSHProviderConfig{
			"github": {
				Hostname:     "github.com",
				User:         "git",
				Port:         22,
				IdentityFile: "$HOME/.ssh/id_ed25519",
				HostAlias:    "gh",
				Options: map[string]string{
					"StrictHostKeyChecking": "yes",
					"UserKnownHostsFile":    "$HOME/.ssh/known_hosts",
				},
			},
		},
		KeyManagement: &SSHKeyManagement{
			Enabled:        true,
			KeyDir:         "$HOME/.ssh",
			DefaultKeyType: "ed25519",
			UseSSHAgent:    true,
		},
	}

	assert.True(t, config.Enabled)
	assert.Equal(t, "$HOME/.ssh/config", config.ConfigFile)
	assert.True(t, config.BackupEnabled)
	assert.Equal(t, "$HOME/.ssh/backups", config.BackupDir)
	assert.Contains(t, config.ProviderConfigs, "github")
	assert.Equal(t, "github.com", config.ProviderConfigs["github"].Hostname)
	assert.Equal(t, "git", config.ProviderConfigs["github"].User)
	assert.Equal(t, 22, config.ProviderConfigs["github"].Port)
	assert.Equal(t, "gh", config.ProviderConfigs["github"].HostAlias)
	assert.True(t, config.KeyManagement.Enabled)
	assert.Equal(t, "ed25519", config.KeyManagement.DefaultKeyType)
	assert.True(t, config.KeyManagement.UseSSHAgent)
}

func TestValidationTags(t *testing.T) {
	// Test that validation tags are properly set
	config := DefaultUnifiedConfig()

	// Test version validation
	assert.Equal(t, "1.0.0", config.Version)

	// Test provider validation
	assert.Equal(t, "github", config.DefaultProvider)
	assert.True(t, ValidateProvider(config.DefaultProvider))

	// Test timeout values
	assert.True(t, config.Global.Timeouts.HTTPTimeout > 0)
	assert.True(t, config.Global.Timeouts.GitTimeout > 0)
	assert.True(t, config.Global.Timeouts.RateLimitTimeout > 0)

	// Test concurrency values
	assert.True(t, config.Global.Concurrency.CloneWorkers >= 1)
	assert.True(t, config.Global.Concurrency.UpdateWorkers >= 1)
	assert.True(t, config.Global.Concurrency.APIWorkers >= 1)
}

func TestConfigurationStructure(t *testing.T) {
	// Test that all required fields are present in the default configuration
	config := DefaultUnifiedConfig()

	// Required fields
	require.NotEmpty(t, config.Version)
	require.NotEmpty(t, config.DefaultProvider)
	require.NotNil(t, config.Global)
	require.NotNil(t, config.Providers)

	// Optional but default-initialized fields
	require.NotNil(t, config.IDE)
	require.NotNil(t, config.DevEnv)
	require.NotNil(t, config.NetEnv)
	require.NotNil(t, config.SSHConfig)

	// Test nested structures
	require.NotNil(t, config.Global.Timeouts)
	require.NotNil(t, config.Global.Concurrency)
	require.NotNil(t, config.IDE.SyncSettings)
	require.NotNil(t, config.IDE.Logging)
	require.NotNil(t, config.DevEnv.Providers)
	require.NotNil(t, config.DevEnv.Containers)
	require.NotNil(t, config.DevEnv.Kubernetes)
	require.NotNil(t, config.NetEnv.WiFiMonitoring)
	require.NotNil(t, config.NetEnv.VPN)
	require.NotNil(t, config.NetEnv.DNS)
	require.NotNil(t, config.SSHConfig.KeyManagement)
}

func TestTimeoutSettings(t *testing.T) {
	timeouts := &TimeoutSettings{
		HTTPTimeout:      30 * time.Second,
		GitTimeout:       5 * time.Minute,
		RateLimitTimeout: 1 * time.Hour,
	}

	assert.Equal(t, 30*time.Second, timeouts.HTTPTimeout)
	assert.Equal(t, 5*time.Minute, timeouts.GitTimeout)
	assert.Equal(t, 1*time.Hour, timeouts.RateLimitTimeout)
}

func TestConcurrencySettings(t *testing.T) {
	concurrency := &ConcurrencySettings{
		CloneWorkers:  10,
		UpdateWorkers: 15,
		APIWorkers:    5,
	}

	assert.Equal(t, 10, concurrency.CloneWorkers)
	assert.Equal(t, 15, concurrency.UpdateWorkers)
	assert.Equal(t, 5, concurrency.APIWorkers)
}

func TestMigrationInfo(t *testing.T) {
	migration := &MigrationInfo{
		SourceFormat:  "bulk-clone.yaml",
		MigrationDate: time.Now(),
		SourcePath:    "/home/user/bulk-clone.yaml",
		ToolVersion:   "1.0.0",
	}

	assert.Equal(t, "bulk-clone.yaml", migration.SourceFormat)
	assert.Equal(t, "/home/user/bulk-clone.yaml", migration.SourcePath)
	assert.Equal(t, "1.0.0", migration.ToolVersion)
	assert.False(t, migration.MigrationDate.IsZero())
}
