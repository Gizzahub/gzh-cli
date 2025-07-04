// Package config provides comprehensive configuration management for gzh-manager.
//
// This package handles:
//   - Loading and parsing gzh configuration files (YAML format)
//   - Multi-provider configuration support (GitHub, GitLab, Gitea)
//   - Configuration validation and schema enforcement
//   - Environment variable integration and expansion
//   - Repository filtering and visibility controls
//   - Provider-specific configuration management
//
// The package implements a unified configuration system that supports multiple
// Git service providers through a common interface. Configuration files are
// searched in standard locations with support for both system-wide and
// user-specific settings.
//
// Configuration search order:
//  1. Environment variable (GZH_CONFIG_PATH)
//  2. Current directory (./gzh.yaml, ./gzh.yml)
//  3. User config directory (~/.config/gzh.yaml)
//  4. System config directory (/etc/gzh-manager/gzh.yaml)
//
// Main types:
//   - Config: Root configuration structure
//   - ProviderConfig: Provider-specific settings
//   - RepositoryFilter: Repository filtering criteria
//   - DirectoryResolver: Path resolution utilities
//
// Key interfaces:
//   - ConfigLoader: Configuration loading abstraction
//   - ConfigValidator: Configuration validation interface
//   - ProviderConfigManager: Provider-specific config management
//
// The package supports advanced features like regex-based filtering,
// visibility controls, and dynamic configuration resolution.
package config
