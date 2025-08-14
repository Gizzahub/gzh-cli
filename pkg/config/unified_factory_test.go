// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Gizzahub/gzh-cli/internal/env"
)

func TestNewConfigFactory(t *testing.T) {
	factory := NewConfigFactory()

	assert.NotNil(t, factory)
	assert.NotNil(t, factory.environment)
	assert.NotNil(t, factory.logger)
	assert.True(t, len(factory.searchPaths) > 0)
	assert.True(t, factory.autoMigrate)
	assert.True(t, factory.preferUnified)
	assert.True(t, factory.createBackup)
}

func TestNewConfigFactoryWithOptions(t *testing.T) {
	mockEnv := env.NewMockEnvironment(map[string]string{})
	mockLogger := &NoOpLogger{}
	customPaths := []string{"./custom.yaml"}

	opts := &ConfigFactoryOptions{
		Environment:   mockEnv,
		Logger:        mockLogger,
		SearchPaths:   customPaths,
		AutoMigrate:   false,
		PreferUnified: false,
		CreateBackup:  false,
	}

	factory := NewConfigFactoryWithOptions(opts)

	assert.Equal(t, mockEnv, factory.environment)
	assert.Equal(t, mockLogger, factory.logger)
	assert.Equal(t, customPaths, factory.searchPaths)
	assert.False(t, factory.autoMigrate)
	assert.False(t, factory.preferUnified)
	assert.False(t, factory.createBackup)
}

func TestConfigFactory_FindConfigFile(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		envVar      string
		setupFiles  []string
		expected    string
		expectError bool
	}{
		{
			name:       "finds config from environment variable",
			envVar:     filepath.Join(tmpDir, "env-config.yaml"),
			setupFiles: []string{filepath.Join(tmpDir, "env-config.yaml")},
			expected:   filepath.Join(tmpDir, "env-config.yaml"),
		},
		{
			name:       "finds config from search paths",
			setupFiles: []string{filepath.Join(tmpDir, "gzh.yaml")},
			expected:   filepath.Join(tmpDir, "gzh.yaml"),
		},
		{
			name:        "returns error when no config found",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment
			mockEnv := env.NewMockEnvironment(map[string]string{})
			if tt.envVar != "" {
				mockEnv.Set(env.CommonEnvironmentKeys.GZHConfigPath, tt.envVar)
			}

			// Setup test files
			for _, file := range tt.setupFiles {
				require.NoError(t, os.WriteFile(file, []byte("version: 1.0.0"), 0o600))
			}

			// Create factory with custom search paths
			factory := NewConfigFactory()
			factory.environment = mockEnv
			factory.searchPaths = []string{filepath.Join(tmpDir, "gzh.yaml")}

			result, err := factory.FindConfigFile()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestConfigFactory_CreateDefaultConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.yaml")

	factory := NewConfigFactory()
	err := factory.CreateDefaultConfig(configPath)

	require.NoError(t, err)
	assert.FileExists(t, configPath)

	// Verify content
	content, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), `version: "1.0.0"`)
	assert.Contains(t, string(content), "providers:")
	assert.Contains(t, string(content), "github:")
	assert.Contains(t, string(content), "gitlab:")
}

func TestConfigFactory_GetDefaultConfigPath(t *testing.T) {
	tests := []struct {
		name     string
		homeDir  string
		expected string
	}{
		{
			name:     "with home directory",
			homeDir:  "/home/user",
			expected: "/home/user/.config/gzh.yaml",
		},
		{
			name:     "without home directory",
			homeDir:  "",
			expected: "./gzh.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEnv := env.NewMockEnvironment(map[string]string{})
			if tt.homeDir != "" {
				mockEnv.Set(env.CommonEnvironmentKeys.HomeDir, tt.homeDir)
			}

			factory := NewConfigFactory()
			factory.environment = mockEnv

			result := factory.GetDefaultConfigPath()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfigFactory_CreateProviderFactory(t *testing.T) {
	factory := NewConfigFactory()
	providerFactory := factory.CreateProviderFactory()

	assert.NotNil(t, providerFactory)
	assert.True(t, len(providerFactory.GetSupportedProviders()) > 0)
}

func TestConfigFactory_CreateProviderCloner(t *testing.T) {
	factory := NewConfigFactory()
	ctx := context.Background()

	// Test with supported provider
	cloner, err := factory.CreateProviderCloner(ctx, "github", "test-token")
	assert.NoError(t, err)
	assert.NotNil(t, cloner)

	// Test with unsupported provider
	_, err = factory.CreateProviderCloner(ctx, "unsupported", "test-token")
	assert.Error(t, err)
}

func TestConfigFactory_ExpandPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		homeDir  string
		expected string
	}{
		{
			name:     "expands tilde",
			path:     "~/config.yaml",
			homeDir:  "/home/user",
			expected: "/home/user/config.yaml",
		},
		{
			name:     "absolute path unchanged",
			path:     "/etc/config.yaml",
			homeDir:  "/home/user",
			expected: "/etc/config.yaml",
		},
		{
			name:     "relative path converted to absolute",
			path:     "config.yaml",
			homeDir:  "/home/user",
			expected: filepath.Join(getCurrentDir(), "config.yaml"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEnv := env.NewMockEnvironment(map[string]string{})
			mockEnv.Set(env.CommonEnvironmentKeys.HomeDir, tt.homeDir)

			factory := NewConfigFactory()
			factory.environment = mockEnv

			result := factory.expandPath(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfigFactory_SetAndGetSearchPaths(t *testing.T) {
	factory := NewConfigFactory()
	customPaths := []string{"./custom1.yaml", "./custom2.yaml"}

	factory.SetSearchPaths(customPaths)
	result := factory.GetSearchPaths()

	// Paths should be expanded to absolute paths
	assert.Len(t, result, 2)
	for _, path := range result {
		assert.True(t, filepath.IsAbs(path))
	}
}

func TestNoOpLogger(t *testing.T) {
	logger := &NoOpLogger{}

	// These should not panic
	logger.Debug("test")
	logger.Info("test")
	logger.Warn("test")
	logger.Error("test")
}

func getCurrentDir() string {
	dir, _ := os.Getwd()
	return dir
}
