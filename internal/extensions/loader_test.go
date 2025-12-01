// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package extensions

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLoader(t *testing.T) {
	loader := NewLoader()
	assert.NotNil(t, loader)
	assert.NotEmpty(t, loader.configPath)
}

func TestLoadConfig_NoFile(t *testing.T) {
	loader := &Loader{
		configPath: "/nonexistent/path/extensions.yaml",
	}

	cfg, err := loader.LoadConfig()
	require.NoError(t, err, "Missing config file should not be an error")
	assert.NotNil(t, cfg)
	assert.Empty(t, cfg.Aliases)
	assert.Empty(t, cfg.External)
}

func TestLoadConfig_ValidFile(t *testing.T) {
	// 임시 설정 파일 생성
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "extensions.yaml")

	configContent := `
aliases:
  update-all:
    command: "pm update --all"
    description: "Update all package managers"
  pull-all:
    command: "git repo pull-all"
    description: "Pull all repositories"

external:
  - name: terraform
    command: /usr/local/bin/terraform
    description: "Terraform infrastructure management"
    passthrough: true
`

	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	loader := &Loader{configPath: configPath}
	cfg, err := loader.LoadConfig()
	require.NoError(t, err)

	// 별칭 확인
	assert.Len(t, cfg.Aliases, 2)
	assert.Equal(t, "pm update --all", cfg.Aliases["update-all"].Command)
	assert.Equal(t, "Update all package managers", cfg.Aliases["update-all"].Description)

	// 외부 명령어 확인
	assert.Len(t, cfg.External, 1)
	assert.Equal(t, "terraform", cfg.External[0].Name)
	assert.Equal(t, "/usr/local/bin/terraform", cfg.External[0].Command)
	assert.True(t, cfg.External[0].Passthrough)
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "extensions.yaml")

	// 잘못된 YAML
	err := os.WriteFile(configPath, []byte("invalid: yaml: content:"), 0o644)
	require.NoError(t, err)

	loader := &Loader{configPath: configPath}
	_, err = loader.LoadConfig()
	assert.Error(t, err, "Invalid YAML should return error")
}

func TestRegisterAlias(t *testing.T) {
	rootCmd := &cobra.Command{Use: "test"}
	loader := NewLoader()

	alias := AliasConfig{
		Command:     "git repo pull-all",
		Description: "Pull all repositories",
	}

	err := loader.registerAlias(rootCmd, "pull-all", alias)
	require.NoError(t, err)

	// 명령어가 등록되었는지 확인
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "pull-all" {
			found = true
			assert.Equal(t, "Pull all repositories", cmd.Short)
			break
		}
	}
	assert.True(t, found, "Alias command should be registered")
}

func TestRegisterAlias_EmptyCommand(t *testing.T) {
	rootCmd := &cobra.Command{Use: "test"}
	loader := NewLoader()

	alias := AliasConfig{
		Command:     "", // 빈 명령어
		Description: "Test alias",
	}

	err := loader.registerAlias(rootCmd, "test-alias", alias)
	assert.Error(t, err, "Empty command should return error")
}

func TestRegisterExternal_CommandNotFound(t *testing.T) {
	rootCmd := &cobra.Command{Use: "test"}
	loader := NewLoader()

	ext := ExternalCommandConfig{
		Name:        "nonexistent",
		Command:     "/nonexistent/command",
		Description: "Nonexistent command",
	}

	err := loader.registerExternal(rootCmd, ext)
	assert.Error(t, err, "Nonexistent command should return error")
}

func TestRegisterExternal_ValidCommand(t *testing.T) {
	rootCmd := &cobra.Command{Use: "test"}
	loader := NewLoader()

	// ls 명령어는 거의 모든 시스템에 존재
	ext := ExternalCommandConfig{
		Name:        "list",
		Command:     "ls",
		Description: "List files",
		Passthrough: true,
	}

	err := loader.registerExternal(rootCmd, ext)
	require.NoError(t, err)

	// 명령어가 등록되었는지 확인
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "list" {
			found = true
			assert.Contains(t, cmd.Short, "[EXTERNAL]")
			assert.True(t, cmd.DisableFlagParsing)
			break
		}
	}
	assert.True(t, found, "External command should be registered")
}

func TestRegisterAll_IntegrationTest(t *testing.T) {
	// 임시 설정 파일 생성
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "extensions.yaml")

	configContent := `
aliases:
  test-alias:
    command: "version"
    description: "Show version"

external:
  - name: echo-test
    command: echo
    description: "Echo command"
    args: ["hello"]
`

	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	rootCmd := &cobra.Command{Use: "test"}
	loader := &Loader{configPath: configPath}

	err = loader.RegisterAll(rootCmd)
	require.NoError(t, err)

	// 등록된 명령어 확인
	commands := make(map[string]bool)
	for _, cmd := range rootCmd.Commands() {
		commands[cmd.Name()] = true
	}

	assert.True(t, commands["test-alias"], "Alias should be registered")
	assert.True(t, commands["echo-test"], "External command should be registered")
}

func TestRegisterAll_NoConfigFile(t *testing.T) {
	rootCmd := &cobra.Command{Use: "test"}
	loader := &Loader{
		configPath: "/nonexistent/path/extensions.yaml",
	}

	err := loader.RegisterAll(rootCmd)
	require.NoError(t, err, "Missing config should not cause error")
}

func TestRegisterAll_PartialFailure(t *testing.T) {
	// 일부 명령어는 성공하고 일부는 실패하는 경우
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "extensions.yaml")

	configContent := `
aliases:
  valid-alias:
    command: "version"
    description: "Valid alias"
  invalid-alias:
    command: ""
    description: "Invalid alias"

external:
  - name: valid-external
    command: echo
    description: "Valid external"
  - name: invalid-external
    command: /nonexistent/command
    description: "Invalid external"
`

	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	rootCmd := &cobra.Command{Use: "test"}
	loader := &Loader{configPath: configPath}

	// 부분 실패해도 전체 등록은 성공해야 함
	err = loader.RegisterAll(rootCmd)
	require.NoError(t, err)

	// 유효한 명령어만 등록되었는지 확인
	commands := make(map[string]bool)
	for _, cmd := range rootCmd.Commands() {
		commands[cmd.Name()] = true
	}

	assert.True(t, commands["valid-alias"], "Valid alias should be registered")
	assert.False(t, commands["invalid-alias"], "Invalid alias should not be registered")
	assert.True(t, commands["valid-external"], "Valid external should be registered")
	assert.False(t, commands["invalid-external"], "Invalid external should not be registered")
}
