// Package config provides centralized configuration management for gzh-manager.
//
// This package implements a unified configuration service that handles:
//   - Loading configuration from multiple sources (files, environment variables)
//   - Configuration validation and migration
//   - Hot-reloading and file watching
//   - Integration with Viper for flexible configuration management
//
// Key Components:
//
// ConfigService Interface:
// The main interface for configuration management operations.
// Provides methods for loading, reloading, saving, and watching configuration files.
//
// DefaultConfigService:
// The primary implementation of ConfigService using Viper as the underlying
// configuration management library. Supports both unified gzh.yaml format
// and legacy bulk-clone.yaml format with automatic migration.
//
// ServiceFactory:
// Factory for creating ConfigService instances with different options.
// Supports dependency injection and testing scenarios.
//
// Usage Example:
//
//	// Create a configuration service
//	service, err := config.CreateDefaultConfigService()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Load configuration
//	ctx := context.Background()
//	cfg, err := service.LoadConfiguration(ctx, "")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Use configuration
//	providers := service.GetConfiguredProviders()
//	targets, err := service.GetBulkCloneTargets(ctx, "github")
//
//	// Watch for configuration changes
//	err = service.WatchConfiguration(ctx, func(cfg *config.UnifiedConfig) {
//	    log.Println("Configuration reloaded")
//	})
//
// Configuration Sources:
//
// The service searches for configuration files in the following order:
//   1. Explicitly provided path
//   2. Current directory (./gzh.yaml, ./gzh.yml)
//   3. User config directory (~/.config/gzh-manager/gzh.yaml)
//   4. System config directory (/etc/gzh-manager/gzh.yaml)
//
// Environment Variables:
//
// Configuration values can be overridden using environment variables with
// the prefix "GZH_". For example:
//   - GZH_DEFAULT_PROVIDER=gitlab
//   - GZH_PROVIDERS_GITHUB_TOKEN=secret
//
// Hot Reloading:
//
// The service supports watching configuration files for changes and
// automatically reloading them. This is useful for long-running processes
// that need to adapt to configuration changes without restart.
//
// Migration Support:
//
// The service automatically detects and migrates legacy bulk-clone.yaml
// configurations to the new unified gzh.yaml format. Migration information
// is preserved and can be accessed through the service API.
package config