package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// UnifiedConfigFacade provides a unified interface for configuration management
type UnifiedConfigFacade struct {
	loader       *UnifiedConfigLoader
	config       *UnifiedConfig
	loadResult   *LoadResult
	integration  *BulkCloneIntegration
}

// NewUnifiedConfigFacade creates a new unified configuration facade
func NewUnifiedConfigFacade() *UnifiedConfigFacade {
	return &UnifiedConfigFacade{
		loader: NewUnifiedConfigLoader(),
	}
}

// LoadConfiguration loads configuration from available sources
func (f *UnifiedConfigFacade) LoadConfiguration() error {
	return f.LoadConfigurationFromPath("")
}

// LoadConfigurationFromPath loads configuration from a specific path
func (f *UnifiedConfigFacade) LoadConfigurationFromPath(configPath string) error {
	result, err := f.loader.LoadConfigFromPath(configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	f.loadResult = result
	f.config = result.Config

	// Create integration adapter
	legacyConfig := f.convertToLegacyFormat(result.Config)
	f.integration = NewBulkCloneIntegration(legacyConfig)

	return nil
}

// GetConfiguration returns the loaded configuration
func (f *UnifiedConfigFacade) GetConfiguration() *UnifiedConfig {
	return f.config
}

// GetLoadResult returns the load result with migration details
func (f *UnifiedConfigFacade) GetLoadResult() *LoadResult {
	return f.loadResult
}

// GetBulkCloneIntegration returns the bulk clone integration adapter
func (f *UnifiedConfigFacade) GetBulkCloneIntegration() *BulkCloneIntegration {
	return f.integration
}

// SaveConfiguration saves the current configuration to a file
func (f *UnifiedConfigFacade) SaveConfiguration(configPath string) error {
	if f.config == nil {
		return fmt.Errorf("no configuration loaded")
	}

	// Ensure directory exists
	if err := CreateDirectory(filepath.Dir(configPath)); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(f.config)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	// Add header
	header := fmt.Sprintf(`# gzh-manager unified configuration
# Generated: %s
# Version: %s
# Documentation: https://github.com/gizzahub/gzh-manager-go/docs/configuration.md

`, time.Now().Format("2006-01-02 15:04:05"), f.config.Version)

	content := header + string(data)

	// Write to file
	if err := WriteFile(configPath, content); err != nil {
		return fmt.Errorf("failed to write configuration: %w", err)
	}

	return nil
}

// CreateDefaultConfiguration creates a default configuration file
func (f *UnifiedConfigFacade) CreateDefaultConfiguration(configPath string) error {
	defaultConfig := DefaultUnifiedConfig()
	
	// Set up a basic example configuration
	defaultConfig.Providers["github"] = &ProviderConfig{
		Token: "${GITHUB_TOKEN}",
		Organizations: []*OrganizationConfig{
			{
				Name:       "your-org-name",
				CloneDir:   "~/repos/github/your-org-name",
				Visibility: VisibilityAll,
				Strategy:   StrategyReset,
				Exclude:    []string{"test-.*", ".*-archive"},
			},
		},
	}

	f.config = defaultConfig
	return f.SaveConfiguration(configPath)
}

// ValidateConfiguration validates the current configuration
func (f *UnifiedConfigFacade) ValidateConfiguration() error {
	if f.config == nil {
		return fmt.Errorf("no configuration loaded")
	}

	return f.loader.validateUnifiedConfig(f.config)
}

// GetProviderTargets returns all targets for a specific provider
func (f *UnifiedConfigFacade) GetProviderTargets(providerName string) ([]BulkCloneTarget, error) {
	if f.integration == nil {
		return nil, fmt.Errorf("no configuration loaded")
	}

	return f.integration.GetTargetsByProvider(providerName)
}

// GetAllTargets returns all configured targets
func (f *UnifiedConfigFacade) GetAllTargets() ([]BulkCloneTarget, error) {
	if f.integration == nil {
		return nil, fmt.Errorf("no configuration loaded")
	}

	return f.integration.GetAllTargets()
}

// GetConfiguredProviders returns all configured providers
func (f *UnifiedConfigFacade) GetConfiguredProviders() []string {
	if f.integration == nil {
		return []string{}
	}

	return f.integration.GetConfiguredProviders()
}

// MigrateConfiguration migrates a legacy configuration to unified format
func (f *UnifiedConfigFacade) MigrateConfiguration(sourcePath, targetPath string) (*MigrationResult, error) {
	return MigrateConfigFile(sourcePath, targetPath, false)
}

// GenerateConfigurationReport generates a configuration report
func (f *UnifiedConfigFacade) GenerateConfigurationReport() (string, error) {
	if f.config == nil {
		return "", fmt.Errorf("no configuration loaded")
	}

	targets, err := f.GetAllTargets()
	if err != nil {
		return "", fmt.Errorf("failed to get targets: %w", err)
	}

	report := fmt.Sprintf(`# Configuration Report

**Generated:** %s
**Version:** %s
**Configuration Path:** %s
**Format:** %s

## Summary
- **Providers:** %d
- **Total Targets:** %d
- **Migration Status:** %s

## Providers Configuration
`, 
		time.Now().Format("2006-01-02 15:04:05"),
		f.config.Version,
		f.loadResult.ConfigPath,
		map[bool]string{true: "Legacy (auto-converted)", false: "Unified"}[f.loadResult.IsLegacy],
		len(f.config.Providers),
		len(targets),
		map[bool]string{true: "Migrated", false: "Native"}[f.loadResult.WasMigrated],
	)

	// Add provider details
	for providerName, provider := range f.config.Providers {
		report += fmt.Sprintf(`
### %s
- **Token:** %s
- **Organizations:** %d
- **API URL:** %s
`, 
			providerName,
			maskToken(provider.Token),
			len(provider.Organizations),
			provider.APIURL,
		)

		for _, org := range provider.Organizations {
			report += fmt.Sprintf(`  - **%s:** %s (%s, %s)
`, 
				org.Name,
				org.CloneDir,
				org.Visibility,
				org.Strategy,
			)
		}
	}

	// Add warnings and required actions
	if len(f.loadResult.Warnings) > 0 {
		report += "\n## Warnings\n"
		for _, warning := range f.loadResult.Warnings {
			report += fmt.Sprintf("- %s\n", warning)
		}
	}

	if len(f.loadResult.RequiredActions) > 0 {
		report += "\n## Required Actions\n"
		for _, action := range f.loadResult.RequiredActions {
			report += fmt.Sprintf("- [ ] %s\n", action)
		}
	}

	return report, nil
}

// convertToLegacyFormat converts unified configuration to legacy format for compatibility
func (f *UnifiedConfigFacade) convertToLegacyFormat(unifiedConfig *UnifiedConfig) *Config {
	legacyConfig := &Config{
		Version:         unifiedConfig.Version,
		DefaultProvider: unifiedConfig.DefaultProvider,
		Providers:       make(map[string]Provider),
	}

	// Convert providers
	for providerName, provider := range unifiedConfig.Providers {
		legacyProvider := Provider{
			Token:  provider.Token,
			Orgs:   []GitTarget{},
			Groups: []GitTarget{},
		}

		// Convert organizations
		for _, org := range provider.Organizations {
			target := GitTarget{
				Name:       org.Name,
				Visibility: org.Visibility,
				Recursive:  org.Recursive,
				Flatten:    org.Flatten,
				Match:      org.Include,
				CloneDir:   org.CloneDir,
				Exclude:    org.Exclude,
				Strategy:   org.Strategy,
			}

			if providerName == "gitlab" {
				legacyProvider.Groups = append(legacyProvider.Groups, target)
			} else {
				legacyProvider.Orgs = append(legacyProvider.Orgs, target)
			}
		}

		legacyConfig.Providers[providerName] = legacyProvider
	}

	return legacyConfig
}

// maskToken masks sensitive token information for display
func maskToken(token string) string {
	if token == "" {
		return "Not configured"
	}

	// If it's an environment variable, show it as is
	if token[0] == '$' {
		return token
	}

	// Mask actual token values
	if len(token) > 8 {
		return token[:4] + "***" + token[len(token)-4:]
	}

	return "***"
}

// IsConfigurationLoaded checks if a configuration is loaded
func (f *UnifiedConfigFacade) IsConfigurationLoaded() bool {
	return f.config != nil
}

// HasWarnings checks if there are any warnings from configuration loading
func (f *UnifiedConfigFacade) HasWarnings() bool {
	return f.loadResult != nil && len(f.loadResult.Warnings) > 0
}

// HasRequiredActions checks if there are any required actions
func (f *UnifiedConfigFacade) HasRequiredActions() bool {
	return f.loadResult != nil && len(f.loadResult.RequiredActions) > 0
}

// GetWarnings returns all warnings from configuration loading
func (f *UnifiedConfigFacade) GetWarnings() []string {
	if f.loadResult == nil {
		return []string{}
	}
	return f.loadResult.Warnings
}

// GetRequiredActions returns all required actions
func (f *UnifiedConfigFacade) GetRequiredActions() []string {
	if f.loadResult == nil {
		return []string{}
	}
	return f.loadResult.RequiredActions
}

// AutoMigrateIfNeeded automatically migrates configuration if legacy format is detected
func (f *UnifiedConfigFacade) AutoMigrateIfNeeded(configPath string) (*MigrationResult, error) {
	return AutoMigrate(configPath)
}

// SetAutoMigrate configures whether to automatically migrate legacy configurations
func (f *UnifiedConfigFacade) SetAutoMigrate(autoMigrate bool) {
	f.loader.AutoMigrate = autoMigrate
}

// SetPreferUnified configures whether to prefer unified format files
func (f *UnifiedConfigFacade) SetPreferUnified(preferUnified bool) {
	f.loader.PreferUnified = preferUnified
}