//nolint:testpackage // White-box testing needed for internal function access
package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnifiedConfigFacade(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.yaml")

	// Create a new facade
	facade := NewUnifiedConfigFacade()
	assert.NotNil(t, facade)

	// Initially, no configuration should be loaded
	assert.Nil(t, facade.GetConfiguration())
	assert.Nil(t, facade.GetLoadResult())

	// Create a default configuration
	err := facade.CreateDefaultConfiguration(configPath)
	require.NoError(t, err)

	// Verify the file was created
	assert.FileExists(t, configPath)

	// Load the configuration
	err = facade.LoadConfigurationFromPath(configPath)
	require.NoError(t, err)

	// Verify configuration is loaded
	config := facade.GetConfiguration()
	require.NotNil(t, config)
	assert.Equal(t, "1.0.0", config.Version)
	assert.Equal(t, "github", config.DefaultProvider)

	// Test load result
	loadResult := facade.GetLoadResult()
	require.NotNil(t, loadResult)
	assert.Equal(t, config, loadResult.Config)
}

func TestUnifiedConfigFacadeGetters(t *testing.T) {
	facade := NewUnifiedConfigFacade()

	// Create a default configuration in memory
	defaultConfig := DefaultUnifiedConfig()
	facade.config = defaultConfig

	// Test IDE configuration getter
	ideConfig := facade.GetIDEConfig()
	assert.NotNil(t, ideConfig)
	assert.Equal(t, defaultConfig.IDE, ideConfig)

	// Test DevEnv configuration getter
	devEnvConfig := facade.GetDevEnvConfig()
	assert.NotNil(t, devEnvConfig)
	assert.Equal(t, defaultConfig.DevEnv, devEnvConfig)

	// Test NetEnv configuration getter
	netEnvConfig := facade.GetNetEnvConfig()
	assert.NotNil(t, netEnvConfig)
	assert.Equal(t, defaultConfig.NetEnv, netEnvConfig)

	// Test SSH configuration getter
	sshConfig := facade.GetSSHConfig()
	assert.NotNil(t, sshConfig)
	assert.Equal(t, defaultConfig.SSHConfig, sshConfig)

	// Test global settings getter
	globalSettings := facade.GetGlobalSettings()
	assert.NotNil(t, globalSettings)
	assert.Equal(t, defaultConfig.Global, globalSettings)

	// Test provider configuration getter
	providerConfig := facade.GetProviderConfig("github")
	assert.Nil(t, providerConfig) // No providers configured in default

	// Add a provider and test
	defaultConfig.Providers["github"] = &ProviderConfig{
		Token: "test-token",
	}
	providerConfig = facade.GetProviderConfig("github")
	assert.NotNil(t, providerConfig)
	assert.Equal(t, "test-token", providerConfig.Token)
}

func TestUnifiedConfigFacadeSetters(t *testing.T) {
	facade := NewUnifiedConfigFacade()
	facade.config = DefaultUnifiedConfig()

	// Test IDE configuration update
	newIDEConfig := &IDEConfig{
		Enabled:     false,
		AutoFixSync: false,
	}
	err := facade.UpdateIDEConfig(newIDEConfig)
	assert.NoError(t, err)
	assert.Equal(t, newIDEConfig, facade.config.IDE)

	// Test DevEnv configuration update
	newDevEnvConfig := &DevEnvConfig{
		Enabled:        false,
		BackupLocation: "/tmp/backups",
	}
	err = facade.UpdateDevEnvConfig(newDevEnvConfig)
	assert.NoError(t, err)
	assert.Equal(t, newDevEnvConfig, facade.config.DevEnv)

	// Test NetEnv configuration update
	newNetEnvConfig := &NetEnvConfig{
		Enabled: false,
	}
	err = facade.UpdateNetEnvConfig(newNetEnvConfig)
	assert.NoError(t, err)
	assert.Equal(t, newNetEnvConfig, facade.config.NetEnv)

	// Test SSH configuration update
	newSSHConfig := &SSHConfigSettings{
		Enabled:    false,
		ConfigFile: "/tmp/ssh-config",
	}
	err = facade.UpdateSSHConfig(newSSHConfig)
	assert.NoError(t, err)
	assert.Equal(t, newSSHConfig, facade.config.SSHConfig)
}

func TestUnifiedConfigFacadeSettersWithoutConfig(t *testing.T) {
	facade := NewUnifiedConfigFacade()
	// No configuration loaded

	// Test that setters fail when no configuration is loaded
	err := facade.UpdateIDEConfig(&IDEConfig{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no configuration loaded")

	err = facade.UpdateDevEnvConfig(&DevEnvConfig{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no configuration loaded")

	err = facade.UpdateNetEnvConfig(&NetEnvConfig{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no configuration loaded")

	err = facade.UpdateSSHConfig(&SSHConfigSettings{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no configuration loaded")
}

func TestUnifiedConfigFacadeFeatureEnabled(t *testing.T) {
	facade := NewUnifiedConfigFacade()
	facade.config = DefaultUnifiedConfig()

	// Test feature enabled checks
	assert.True(t, facade.IsFeatureEnabled("ide"))
	assert.True(t, facade.IsFeatureEnabled("dev-env"))
	assert.True(t, facade.IsFeatureEnabled("net-env"))

	// Test unknown feature
	assert.False(t, facade.IsFeatureEnabled("unknown"))

	// Test with disabled features
	facade.config.IDE.Enabled = false
	assert.False(t, facade.IsFeatureEnabled("ide"))

	facade.config.DevEnv.Enabled = false
	assert.False(t, facade.IsFeatureEnabled("dev-env"))

	facade.config.NetEnv.Enabled = false
	assert.False(t, facade.IsFeatureEnabled("net-env"))
}

func TestUnifiedConfigFacadeFeatureEnabledWithoutConfig(t *testing.T) {
	facade := NewUnifiedConfigFacade()
	// No configuration loaded

	// All features should be disabled when no configuration is loaded
	assert.False(t, facade.IsFeatureEnabled("ide"))
	assert.False(t, facade.IsFeatureEnabled("dev-env"))
	assert.False(t, facade.IsFeatureEnabled("net-env"))
}

func TestUnifiedConfigFacadeConfigurationSummary(t *testing.T) {
	facade := NewUnifiedConfigFacade()
	facade.config = DefaultUnifiedConfig()

	// Add a provider for testing
	facade.config.Providers["github"] = &ProviderConfig{
		Token: "test-token",
	}

	summary := facade.GetConfigurationSummary()
	require.NotNil(t, summary)

	// Test summary fields
	assert.Equal(t, "1.0.0", summary["version"])
	assert.Equal(t, "github", summary["default_provider"])
	assert.Equal(t, 1, summary["providers"])

	// Test global settings in summary
	globalSummary, ok := summary["global"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "$HOME/repos", globalSummary["clone_base_dir"])
	assert.Equal(t, "reset", globalSummary["default_strategy"])
	assert.Equal(t, "all", globalSummary["default_visibility"])
	assert.Equal(t, 10, globalSummary["clone_workers"])
	assert.Equal(t, 15, globalSummary["update_workers"])

	// Test features in summary
	features, ok := summary["features"].(map[string]bool)
	require.True(t, ok)
	assert.True(t, features["ide"])
	assert.True(t, features["dev-env"])
	assert.True(t, features["net-env"])
}

func TestUnifiedConfigFacadeConfigurationSummaryWithoutConfig(t *testing.T) {
	facade := NewUnifiedConfigFacade()
	// No configuration loaded

	summary := facade.GetConfigurationSummary()
	assert.Nil(t, summary)
}

func TestUnifiedConfigFacadeGetExpandedPath(t *testing.T) {
	facade := NewUnifiedConfigFacade()

	// Test path expansion
	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)

	expandedPath := facade.GetExpandedPath("$HOME/repos")
	assert.Equal(t, homeDir+"/repos", expandedPath)

	// Test path without variables
	expandedPath = facade.GetExpandedPath("/absolute/path")
	assert.Equal(t, "/absolute/path", expandedPath)
}

func TestUnifiedConfigFacadeValidateConfiguration(t *testing.T) {
	facade := NewUnifiedConfigFacade()

	// Test validation without configuration
	err := facade.ValidateConfiguration()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no configuration loaded")

	// DefaultUnifiedConfig()는 완성된 설정이 아니라 뼈대다. Providers를 빈
	// 맵으로 두는 것은 의도된 것으로 -- 호출자가 nil 맵 대입 패닉 없이 채워
	// 넣도록 make()로 만든다 -- 실제 호출자인 CreateDefaultConfiguration과
	// 마이그레이션은 모두 그 뒤에 provider를 채운다. 따라서 뼈대 그대로는
	// "at least one provider must be configured"에 걸리는 것이 맞다.
	// 예전에는 이 상태로 NoError를 기대했는데, 주석부터 "일단 패닉만 안 나면
	// 된다"고 적혀 있었다.
	facade.config = DefaultUnifiedConfig()
	err = facade.ValidateConfiguration()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least one provider must be configured")

	// provider까지 채운 완성된 설정은 통과해야 한다.
	facade.config.Providers["github"] = &ProviderConfig{
		Token: "${GITHUB_TOKEN}",
		Organizations: []*OrganizationConfig{
			{Name: "test-org", CloneDir: "~/repos/test-org"},
		},
	}
	assert.NoError(t, facade.ValidateConfiguration())
}

func TestUnifiedConfigFacadeSaveConfiguration(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-save.yaml")

	facade := NewUnifiedConfigFacade()

	// Test save without configuration
	err := facade.SaveConfiguration(configPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no configuration loaded")

	// Test save with configuration
	facade.config = DefaultUnifiedConfig()
	err = facade.SaveConfiguration(configPath)
	assert.NoError(t, err)

	// Verify file was created
	assert.FileExists(t, configPath)

	// Verify content
	content, err := os.ReadFile(configPath)
	require.NoError(t, err)
	// yaml.Marshal은 따옴표가 필요할 때만 붙인다. 1.0.0은 점이 둘이라 숫자로
	// 해석될 수 없으므로 따옴표 없이 나가고, 다시 읽어도 문자열 "1.0.0"이다.
	// 예전 기대값은 직렬화 라이브러리의 인용 스타일을 검증하고 있었다.
	assert.Contains(t, string(content), "version: 1.0.0")
	assert.Contains(t, string(content), "default_provider: github")
	assert.Contains(t, string(content), "# gzh-manager unified configuration")
}

func TestUnifiedConfigFacadeLoadConfiguration(t *testing.T) {
	facade := NewUnifiedConfigFacade()

	// Test load configuration without path
	err := facade.LoadConfiguration()
	// This may fail if no configuration file exists in standard locations
	// We're primarily testing that it doesn't panic
	_ = err

	// Test load configuration with invalid path
	err = facade.LoadConfigurationFromPath("/nonexistent/path/config.yaml")
	assert.Error(t, err)
}
