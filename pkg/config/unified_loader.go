package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gizzahub/gzh-manager-go/pkg/bulk-clone"
	"gopkg.in/yaml.v3"
)

const defaultConfigVersion = "1.0.0"

// UnifiedConfigLoader loads configuration from both legacy and unified formats.
type UnifiedConfigLoader struct {
	ConfigPaths   []string
	AutoMigrate   bool
	PreferUnified bool
	CreateBackup  bool
}

// NewUnifiedConfigLoader creates a new configuration loader.
func NewUnifiedConfigLoader() *UnifiedConfigLoader {
	return &UnifiedConfigLoader{
		ConfigPaths: []string{
			// Unified format files (preferred)
			"./gzh.yaml",
			"./gzh.yml",
			"./config.yaml",
			"./config.yml",
			// Legacy format files
			"./bulk-clone.yaml",
			"./bulk-clone.yml",
		},
		AutoMigrate:   true,
		PreferUnified: true,
		CreateBackup:  true,
	}
}

// LoadResult contains the result of loading a configuration.
type LoadResult struct {
	Config          *UnifiedConfig
	ConfigPath      string
	IsLegacy        bool
	WasMigrated     bool
	MigrationPath   string
	Warnings        []string
	RequiredActions []string
}

// LoadConfig loads configuration from available files.
func (l *UnifiedConfigLoader) LoadConfig() (*LoadResult, error) {
	return l.LoadConfigFromPath("")
}

// LoadConfigFromPath loads configuration from a specific path.
func (l *UnifiedConfigLoader) LoadConfigFromPath(configPath string) (*LoadResult, error) {
	// If specific path provided, use it
	if configPath != "" {
		return l.loadFromSpecificPath(configPath)
	}

	// Add system paths to search list
	searchPaths := l.getSearchPaths()

	// Find the first available config file
	for _, path := range searchPaths {
		if FileExists(path) {
			return l.loadFromSpecificPath(path)
		}
	}

	return nil, fmt.Errorf("no configuration file found in search paths: %v", searchPaths)
}

// loadFromSpecificPath loads configuration from a specific file path.
func (l *UnifiedConfigLoader) loadFromSpecificPath(configPath string) (*LoadResult, error) {
	result := &LoadResult{
		ConfigPath:      configPath,
		Warnings:        []string{},
		RequiredActions: []string{},
	}

	// Check if file exists
	if !FileExists(configPath) {
		return nil, fmt.Errorf("configuration file not found: %s", configPath)
	}

	// Detect format
	isLegacy, err := DetectLegacyFormat(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to detect configuration format: %w", err)
	}

	result.IsLegacy = isLegacy

	if isLegacy {
		return l.loadLegacyConfig(configPath, result)
	}

	return l.loadUnifiedConfig(configPath, result)
}

// loadUnifiedConfig loads a unified format configuration.
func (l *UnifiedConfigLoader) loadUnifiedConfig(configPath string, result *LoadResult) (*LoadResult, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config UnifiedConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal unified config: %w", err)
	}

	// Validate configuration
	if err := l.validateUnifiedConfig(&config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	result.Config = &config

	return result, nil
}

// loadLegacyConfig loads a legacy format configuration.
func (l *UnifiedConfigLoader) loadLegacyConfig(configPath string, result *LoadResult) (*LoadResult, error) {
	// Load legacy configuration
	legacyConfig, err := bulkclone.LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load legacy configuration: %w", err)
	}

	if l.AutoMigrate {
		// Auto-migrate to unified format
		migrationResult, err := l.performAutoMigration(configPath, legacyConfig)
		if err != nil {
			return nil, fmt.Errorf("auto-migration failed: %w", err)
		}

		result.Config = migrationResult.UnifiedConfig
		result.WasMigrated = true
		result.MigrationPath = migrationResult.TargetPath
		result.Warnings = append(result.Warnings, migrationResult.Warnings...)
		result.RequiredActions = append(result.RequiredActions, migrationResult.RequiredActions...)
		result.RequiredActions = append(result.RequiredActions,
			fmt.Sprintf("Consider removing legacy configuration file: %s", configPath))
	} else {
		// Convert legacy config to unified format in memory
		unifiedConfig := l.convertLegacyToUnified(legacyConfig)
		result.Config = unifiedConfig
		result.RequiredActions = append(result.RequiredActions,
			"Consider migrating to unified configuration format for better features")
	}

	return result, nil
}

// performAutoMigration performs automatic migration from legacy to unified format.
func (l *UnifiedConfigLoader) performAutoMigration(configPath string, legacyConfig *bulkclone.BulkCloneConfig) (*MigrationResult, error) {
	// Determine target path
	dir := filepath.Dir(configPath)
	targetPath := filepath.Join(dir, "gzh.yaml")

	// If target already exists, create a versioned name
	if FileExists(targetPath) {
		targetPath = filepath.Join(dir, fmt.Sprintf("gzh.migrated.%s.yaml",
			generateTimestamp()))
	}

	// Create migrator
	migrator := NewConfigMigrator(configPath, targetPath)
	migrator.CreateBackup = l.CreateBackup

	// Perform migration
	return migrator.MigrateFromBulkClone()
}

// convertLegacyToUnified converts legacy configuration to unified format in memory.
func (l *UnifiedConfigLoader) convertLegacyToUnified(legacyConfig *bulkclone.BulkCloneConfig) *UnifiedConfig {
	migrator := NewConfigMigrator("", "")
	unifiedConfig, _, _ := migrator.convertBulkCloneToUnified(legacyConfig)

	return unifiedConfig
}

// validateUnifiedConfig validates a unified configuration.
func (l *UnifiedConfigLoader) validateUnifiedConfig(config *UnifiedConfig) error {
	if config.Version == "" {
		return fmt.Errorf("version is required")
	}

	if config.Version != defaultConfigVersion {
		return fmt.Errorf("unsupported version: %s (expected: 1.0.0)", config.Version)
	}

	if len(config.Providers) == 0 {
		return fmt.Errorf("at least one provider must be configured")
	}

	// Validate each provider
	for providerName, provider := range config.Providers {
		if err := l.validateProvider(providerName, provider); err != nil {
			return fmt.Errorf("provider %s validation failed: %w", providerName, err)
		}
	}

	return nil
}

// validateProvider validates a provider configuration.
func (l *UnifiedConfigLoader) validateProvider(providerName string, provider *ProviderConfig) error {
	if provider.Token == "" {
		return fmt.Errorf("token is required for provider %s", providerName)
	}

	if len(provider.Organizations) == 0 {
		return fmt.Errorf("at least one organization must be configured for provider %s", providerName)
	}

	// Validate each organization
	for _, org := range provider.Organizations {
		if err := l.validateOrganization(org); err != nil {
			return fmt.Errorf("organization %s validation failed: %w", org.Name, err)
		}
	}

	return nil
}

// validateOrganization validates an organization configuration.
func (l *UnifiedConfigLoader) validateOrganization(org *OrganizationConfig) error {
	if org.Name == "" {
		return fmt.Errorf("organization name is required")
	}

	if org.CloneDir == "" {
		return fmt.Errorf("clone directory is required for organization %s", org.Name)
	}

	// Validate visibility
	if org.Visibility != "" && !isValidVisibility(org.Visibility) {
		return fmt.Errorf("invalid visibility %s for organization %s", org.Visibility, org.Name)
	}

	// Validate strategy
	if org.Strategy != "" && !isValidStrategy(org.Strategy) {
		return fmt.Errorf("invalid strategy %s for organization %s", org.Strategy, org.Name)
	}

	// Validate regex pattern
	if org.Include != "" {
		if _, err := CompileRegex(org.Include); err != nil {
			return fmt.Errorf("invalid include pattern for organization %s: %w", org.Name, err)
		}
	}

	// Validate exclude patterns
	for _, pattern := range org.Exclude {
		if _, err := CompileRegex(pattern); err != nil {
			return fmt.Errorf("invalid exclude pattern %s for organization %s: %w", pattern, org.Name, err)
		}
	}

	return nil
}

// getSearchPaths returns all possible configuration file paths.
func (l *UnifiedConfigLoader) getSearchPaths() []string {
	var paths []string

	// Add configured paths
	paths = append(paths, l.ConfigPaths...)

	// Add home directory paths
	homeDir, err := os.UserHomeDir()
	if err == nil {
		configDir := filepath.Join(homeDir, ".config", "gzh-manager")
		paths = append(paths,
			filepath.Join(configDir, "gzh.yaml"),
			filepath.Join(configDir, "gzh.yml"),
			filepath.Join(configDir, "config.yaml"),
			filepath.Join(configDir, "config.yml"),
			filepath.Join(configDir, "bulk-clone.yaml"),
			filepath.Join(configDir, "bulk-clone.yml"),
		)
	}

	// Add system paths
	paths = append(paths,
		"/etc/gzh-manager/gzh.yaml",
		"/etc/gzh-manager/gzh.yml",
		"/etc/gzh-manager/config.yaml",
		"/etc/gzh-manager/config.yml",
		"/etc/gzh-manager/bulk-clone.yaml",
		"/etc/gzh-manager/bulk-clone.yml",
	)

	// Sort paths by preference (unified format first if preferred)
	if l.PreferUnified {
		return l.sortPathsByPreference(paths)
	}

	return paths
}

// sortPathsByPreference sorts paths to prefer unified format files.
func (l *UnifiedConfigLoader) sortPathsByPreference(paths []string) []string {
	var (
		unifiedPaths []string
		legacyPaths  []string
	)

	for _, path := range paths {
		if l.isUnifiedFormatPath(path) {
			unifiedPaths = append(unifiedPaths, path)
		} else {
			legacyPaths = append(legacyPaths, path)
		}
	}

	// Return unified paths first, then legacy paths
	return append(unifiedPaths, legacyPaths...)
}

// isUnifiedFormatPath checks if a path is likely to be unified format.
func (l *UnifiedConfigLoader) isUnifiedFormatPath(path string) bool {
	base := filepath.Base(path)

	return base == "gzh.yaml" || base == "gzh.yml" ||
		base == "config.yaml" || base == "config.yml"
}

// generateTimestamp generates a timestamp string for file naming.
func generateTimestamp() string {
	return GenerateTimestamp()
}
