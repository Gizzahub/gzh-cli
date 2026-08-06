// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gizzahub/gzh-cli/internal/env"
)

// ConfigFactory provides a unified interface for loading all configuration types
// with dependency injection support. It consolidates duplicate configuration
// loading logic across the codebase.
type ConfigFactory struct {
	environment   env.Environment
	logger        Logger
	searchPaths   []string
	autoMigrate   bool
	preferUnified bool
	createBackup  bool
}

// NewConfigFactory creates a new configuration factory with default settings.
func NewConfigFactory() *ConfigFactory {
	return &ConfigFactory{
		environment: env.NewOSEnvironment(),
		logger:      &NoOpLogger{},
		// ConfigSearchPaths를 읽는다 -- 공개 변수를 바꾸면 탐색에 반영되어야 한다.
		searchPaths:   append([]string(nil), ConfigSearchPaths...),
		autoMigrate:   true,
		preferUnified: true,
		createBackup:  true,
	}
}

// ConfigFactoryOptions provides options for customizing the config factory.
type ConfigFactoryOptions struct {
	Environment   env.Environment
	Logger        Logger
	SearchPaths   []string
	AutoMigrate   bool
	PreferUnified bool
	CreateBackup  bool
}

// NewConfigFactoryWithOptions creates a new configuration factory with custom options.
func NewConfigFactoryWithOptions(opts *ConfigFactoryOptions) *ConfigFactory {
	factory := NewConfigFactory()

	if opts != nil {
		if opts.Environment != nil {
			factory.environment = opts.Environment
		}
		if opts.Logger != nil {
			factory.logger = opts.Logger
		}
		if len(opts.SearchPaths) > 0 {
			factory.searchPaths = opts.SearchPaths
		}
		factory.autoMigrate = opts.AutoMigrate
		factory.preferUnified = opts.PreferUnified
		factory.createBackup = opts.CreateBackup
	}

	return factory
}

// LoadConfig loads configuration from the first available file using search paths.
//
// GZH_CONFIG_PATH가 설정되어 있으면 그 파일만 쓴다. 이전에는 곧바로
// LoadConfigFromPath("")로 넘겨 UnifiedLoader의 고정 경로 목록만 훑었기 때문에,
// 환경변수 분기가 FindConfigFile 안에만 있고 로드 경로에서는 통째로 죽어 있었다
// -- 문서에 명시된 변수인데 gz가 조용히 무시했다.
func (f *ConfigFactory) LoadConfig() (*UnifiedConfig, error) {
	envPath, err := f.envConfigPath()
	if err != nil {
		return nil, err
	}

	return f.LoadConfigFromPath(envPath)
}

// envConfigPath returns the path from GZH_CONFIG_PATH, or "" when it is unset.
//
// 변수가 가리키는 파일이 없으면 조용히 탐색으로 넘어가지 않고 에러를 낸다.
// 사용자가 경로를 지정했는데 엉뚱한 설정이 로드되는 편이 더 나쁘다.
func (f *ConfigFactory) envConfigPath() (string, error) {
	configPath := f.environment.Get(env.CommonEnvironmentKeys.GZHConfigPath)
	if configPath == "" {
		return "", nil
	}

	expandedPath := f.expandPath(configPath)
	if !f.fileExists(expandedPath) {
		return "", fmt.Errorf("config file specified in GZH_CONFIG_PATH not found: %s", expandedPath)
	}

	return expandedPath, nil
}

// LoadConfigFromPath loads configuration from a specific path or searches if empty.
func (f *ConfigFactory) LoadConfigFromPath(configPath string) (*UnifiedConfig, error) {
	f.logger.Debug("Loading configuration", "path", configPath)

	// Use unified loader for configuration loading
	loader := &UnifiedLoader{
		ConfigPaths:   f.searchPaths,
		AutoMigrate:   f.autoMigrate,
		PreferUnified: f.preferUnified,
		CreateBackup:  f.createBackup,
	}

	result, err := loader.LoadConfigFromPath(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Log warnings and required actions
	for _, warning := range result.Warnings {
		f.logger.Warn("Configuration warning", "message", warning)
	}

	for _, action := range result.RequiredActions {
		f.logger.Info("Required action", "message", action)
	}

	if result.WasMigrated {
		f.logger.Info("Configuration was migrated", "from", result.ConfigPath, "to", result.MigrationPath)
	}

	return result.Config, nil
}

// CreateProviderFactory creates a provider factory using this config factory's dependencies.
func (f *ConfigFactory) CreateProviderFactory() ProviderFactory {
	return NewProviderFactory(f.environment, f.logger)
}

// CreateProviderCloner creates a provider cloner for the specified provider.
func (f *ConfigFactory) CreateProviderCloner(ctx context.Context, providerName, token string) (ProviderCloner, error) {
	factory := f.CreateProviderFactory()
	return factory.CreateCloner(ctx, providerName, token)
}

// FindConfigFile finds the first available configuration file in search paths.
func (f *ConfigFactory) FindConfigFile() (string, error) {
	// Check environment variable first
	envPath, err := f.envConfigPath()
	if err != nil {
		return "", err
	}

	if envPath != "" {
		return envPath, nil
	}

	// Search in predefined paths
	for _, path := range f.searchPaths {
		expandedPath := f.expandPath(path)
		if f.fileExists(expandedPath) {
			return expandedPath, nil
		}
	}

	return "", fmt.Errorf("no configuration file found in search paths: %v", f.searchPaths)
}

// CreateDefaultConfig creates a default configuration file at the specified path.
func (f *ConfigFactory) CreateDefaultConfig(filename string) error {
	defaultConfig := `version: "1.0.0"
default_provider: github

providers:
  github:
    token: "${GITHUB_TOKEN}"
    organizations:
      - name: "your-org-name"
        visibility: all
        clone_dir: "./github"

  gitlab:
    token: "${GITLAB_TOKEN}"
    organizations:
      - name: "your-group-name"
        visibility: public
        clone_dir: "./gitlab"
`

	// Ensure directory exists
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write config file
	if err := os.WriteFile(filename, []byte(defaultConfig), 0o600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	f.logger.Info("Created default configuration", "path", filename)
	return nil
}

// GetDefaultConfigPath returns the default path for creating new config files.
//
// 주입된 environment만 본다. 예전에는 여기서 값이 비면 os.UserHomeDir()로
// 넘어갔는데, 그건 Unix에서 결국 같은 $HOME을 읽으므로 실제 동작에는 보태는
// 것이 없으면서 주입된 environment를 무력화했다 -- 테스트가 빈 MockEnvironment를
// 넣어도 개발자 머신의 홈 디렉토리가 튀어나와, 결과가 머신마다 달라졌다.
func (f *ConfigFactory) GetDefaultConfigPath() string {
	homeDir := f.environment.Get(env.CommonEnvironmentKeys.HomeDir)
	if homeDir == "" {
		return "./gzh.yaml" // Fallback to current directory
	}
	return filepath.Join(homeDir, ".config", "gzh.yaml")
}

// SetSearchPaths updates the search paths for configuration files.
func (f *ConfigFactory) SetSearchPaths(paths []string) {
	f.searchPaths = paths
}

// GetSearchPaths returns the current search paths with variables expanded.
func (f *ConfigFactory) GetSearchPaths() []string {
	paths := make([]string, len(f.searchPaths))
	for i, path := range f.searchPaths {
		paths[i] = f.expandPath(path)
	}
	return paths
}

// expandPath expands ~ to home directory and resolves relative paths.
func (f *ConfigFactory) expandPath(path string) string {
	if path != "" && path[0] == '~' {
		// GetDefaultConfigPath와 같은 이유로 주입된 environment만 본다.
		homeDir := f.environment.Get(env.CommonEnvironmentKeys.HomeDir)
		if homeDir == "" {
			return path // Return original if we can't get home dir
		}
		return filepath.Join(homeDir, path[1:])
	}

	// Expand environment variables
	path = f.environment.Expand(path)

	// Convert to absolute path if relative
	if !filepath.IsAbs(path) {
		if abs, err := filepath.Abs(path); err == nil {
			return abs
		}
	}

	return path
}

// fileExists checks if a file exists and is readable.
func (f *ConfigFactory) fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// getDefaultSearchPaths returns the default search paths for configuration files.
func getDefaultSearchPaths() []string {
	return []string{
		// Current directory (unified format preferred)
		"./gzh.yaml",
		"./gzh.yml",
		"./config.yaml",
		"./config.yml",
		// Legacy format files
		"./bulk-clone.yaml",
		"./bulk-clone.yml",
		// Home directory
		"~/.config/gzh.yaml",
		"~/.config/gzh.yml",
		"~/.config/gzh-manager/gzh.yaml",
		"~/.config/gzh-manager/gzh.yml",
		"~/.config/gzh-manager/bulk-clone.yaml",
		"~/.config/gzh-manager/bulk-clone.yml",
		// System-wide
		"/etc/gzh-manager/gzh.yaml",
		"/etc/gzh-manager/gzh.yml",
		"/etc/gzh-manager/bulk-clone.yaml",
		"/etc/gzh-manager/bulk-clone.yml",
	}
}

// NoOpLogger provides a no-operation logger implementation.
type NoOpLogger struct{}

// Debug implements Logger.Debug.
func (l *NoOpLogger) Debug(_ string, _ ...any) {}

// Info implements Logger.Info.
func (l *NoOpLogger) Info(_ string, _ ...any) {}

// Warn implements Logger.Warn.
func (l *NoOpLogger) Warn(_ string, _ ...any) {}

// Error implements Logger.Error.
func (l *NoOpLogger) Error(_ string, _ ...any) {}
