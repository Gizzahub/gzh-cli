// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	bulkclone "github.com/gizzahub/gzh-manager-go/pkg/bulk-clone"
	"gopkg.in/yaml.v3"
)

// MigrationResult contains the results of a configuration migration.
type MigrationResult struct {
	Success         bool
	SourcePath      string
	TargetPath      string
	BackupPath      string
	MigratedTargets int
	Warnings        []string
	RequiredActions []string
	UnifiedConfig   *UnifiedConfig
	LegacyConfig    *bulkclone.BulkCloneConfig
	MigrationReport string
}

// Migrator handles migration from legacy bulk-clone.yaml to unified format.
type Migrator struct {
	SourcePath   string
	TargetPath   string
	CreateBackup bool
	DryRun       bool
}

// NewMigrator creates a new configuration migrator.
func NewMigrator(sourcePath, targetPath string) *Migrator {
	return &Migrator{
		SourcePath:   sourcePath,
		TargetPath:   targetPath,
		CreateBackup: true,
		DryRun:       false,
	}
}

// MigrateFromBulkClone migrates from bulk-clone.yaml to unified gzh.yaml format.
func (m *Migrator) MigrateFromBulkClone() (*MigrationResult, error) {
	result := &MigrationResult{
		SourcePath:      m.SourcePath,
		TargetPath:      m.TargetPath,
		Warnings:        []string{},
		RequiredActions: []string{},
	}

	// Load legacy configuration
	legacyConfig, err := bulkclone.LoadConfig(m.SourcePath)
	if err != nil {
		return result, fmt.Errorf("failed to load legacy configuration: %w", err)
	}

	result.LegacyConfig = legacyConfig

	// Convert to unified configuration
	unifiedConfig, warnings, actions := m.convertBulkCloneToUnified(legacyConfig)
	result.UnifiedConfig = unifiedConfig
	result.Warnings = warnings
	result.RequiredActions = actions
	result.MigratedTargets = m.countTargets(unifiedConfig)

	// Create backup if requested
	if m.CreateBackup && !m.DryRun {
		backupPath := fmt.Sprintf("%s.backup.%s", m.SourcePath, time.Now().Format("20060102-150405"))
		if err := CopyFile(m.SourcePath, backupPath); err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Failed to create backup: %v", err))
		} else {
			result.BackupPath = backupPath
		}
	}

	// Save unified configuration
	if !m.DryRun {
		if err := m.saveUnifiedConfig(unifiedConfig); err != nil {
			return result, fmt.Errorf("failed to save unified configuration: %w", err)
		}
	}

	// Generate migration report
	result.MigrationReport = m.generateMigrationReport(result)
	result.Success = true

	return result, nil
}

// convertBulkCloneToUnified converts bulk-clone.yaml format to unified format.
func (m *Migrator) convertBulkCloneToUnified(legacy *bulkclone.BulkCloneConfig) (config *UnifiedConfig, warnings, actions []string) {
	config = DefaultUnifiedConfig()

	// Set migration information
	config.Migration = &MigrationInfo{
		SourceFormat:  "bulk-clone.yaml",
		MigrationDate: time.Now(),
		SourcePath:    m.SourcePath,
		ToolVersion:   "gzh-manager-go",
	}

	// Convert version
	if legacy.Version != "" {
		config.Version = "1.0.0" // Always use new version format
		if legacy.Version != "0.1" {
			warnings = append(warnings, fmt.Sprintf("Legacy version %s converted to 1.0.0", legacy.Version))
		}
	}

	// Convert default provider (infer from configured providers)
	if legacy.Default.Github.RootPath != "" || len(legacy.RepoRoots) > 0 {
		config.DefaultProvider = ProviderGitHub
	} else if legacy.Default.Gitlab.RootPath != "" {
		config.DefaultProvider = ProviderGitLab
	}

	// Convert global ignore patterns
	if len(legacy.IgnoreNameRegexes) > 0 {
		config.Global.GlobalIgnores = legacy.IgnoreNameRegexes

		actions = append(actions, "Global ignore patterns moved to per-organization exclude patterns")
	}

	// Convert GitHub configurations
	if err := m.convertGitHubConfigurations(legacy, config, &warnings, &actions); err != nil {
		warnings = append(warnings, fmt.Sprintf("GitHub conversion errors: %v", err))
	}

	// Convert GitLab configurations
	if err := m.convertGitLabConfigurations(legacy, config, &warnings, &actions); err != nil {
		warnings = append(warnings, fmt.Sprintf("GitLab conversion errors: %v", err))
	}

	// Set authentication requirements
	if _, hasGitHub := config.Providers["github"]; hasGitHub {
		actions = append(actions, "Set GITHUB_TOKEN environment variable or update token field")
	}

	if _, hasGitLab := config.Providers["gitlab"]; hasGitLab {
		actions = append(actions, "Set GITLAB_TOKEN environment variable or update token field")
	}

	return config, warnings, actions
}

// convertGitHubConfigurations converts GitHub-specific configurations.
// Note: Currently always returns nil, but maintains error interface for future validation.
func (m *Migrator) convertGitHubConfigurations(legacy *bulkclone.BulkCloneConfig, config *UnifiedConfig, warnings, actions *[]string) error { //nolint:unparam // Always returns nil currently, but interface preserved for future validation
	githubProvider := &ProviderConfig{
		Token:         "${GITHUB_TOKEN}",
		Organizations: []*OrganizationConfig{},
	}

	// Track if we're processing deprecated configurations
	_ = warnings // Parameter reserved for future warning messages

	hasGitHubConfig := false

	// Convert from repo_roots
	for _, repoRoot := range legacy.RepoRoots {
		if repoRoot.Provider == "github" {
			org := &OrganizationConfig{
				Name:       repoRoot.OrgName,
				CloneDir:   repoRoot.RootPath,
				Visibility: VisibilityAll,
				Strategy:   StrategyReset,
				Exclude:    legacy.IgnoreNameRegexes,
			}

			// Convert protocol to strategy hints
			if repoRoot.Protocol == "ssh" {
				*actions = append(*actions, fmt.Sprintf("Organization %s used SSH protocol - configure SSH keys or use tokens", org.Name))
			}

			githubProvider.Organizations = append(githubProvider.Organizations, org)
			hasGitHubConfig = true
		}
	}

	// Convert from defaults if no specific repos configured
	if !hasGitHubConfig && legacy.Default.Github.RootPath != "" {
		org := &OrganizationConfig{
			Name:       legacy.Default.Github.OrgName,
			CloneDir:   legacy.Default.Github.RootPath,
			Visibility: VisibilityAll,
			Strategy:   StrategyReset,
			Exclude:    legacy.IgnoreNameRegexes,
		}

		if legacy.Default.Github.OrgName == "" {
			org.Name = "your-org-name"

			*actions = append(*actions, "Update organization name in GitHub configuration")
		}

		githubProvider.Organizations = append(githubProvider.Organizations, org)
		hasGitHubConfig = true
	}

	// Add GitHub provider if we have configuration
	if hasGitHubConfig {
		config.Providers["github"] = githubProvider
	}

	return nil
}

// convertGitLabConfigurations converts GitLab-specific configurations.
// Note: Currently always returns nil, but maintains error interface for future validation.
func (m *Migrator) convertGitLabConfigurations(legacy *bulkclone.BulkCloneConfig, config *UnifiedConfig, warnings, actions *[]string) error { //nolint:unparam // Always returns nil currently, but interface preserved for future validation
	if legacy.Default.Gitlab.RootPath == "" {
		return nil // No GitLab configuration
	}

	gitlabProvider := &ProviderConfig{
		Token:         "${GITLAB_TOKEN}",
		Organizations: []*OrganizationConfig{},
	}

	// Convert GitLab configuration
	org := &OrganizationConfig{
		Name:       legacy.Default.Gitlab.GroupName,
		CloneDir:   legacy.Default.Gitlab.RootPath,
		Visibility: VisibilityAll,
		Strategy:   StrategyReset,
		Exclude:    legacy.IgnoreNameRegexes,
		Recursive:  legacy.Default.Gitlab.Recursive,
	}

	if legacy.Default.Gitlab.GroupName == "" {
		org.Name = "your-group-name"

		*actions = append(*actions, "Update group name in GitLab configuration")
	}

	// Handle custom GitLab URL
	if legacy.Default.Gitlab.URL != "" && legacy.Default.Gitlab.URL != "https://gitlab.com" {
		gitlabProvider.APIURL = legacy.Default.Gitlab.URL
		*warnings = append(*warnings, fmt.Sprintf("Custom GitLab URL configured: %s", legacy.Default.Gitlab.URL))
	}

	gitlabProvider.Organizations = append(gitlabProvider.Organizations, org)
	config.Providers["gitlab"] = gitlabProvider

	return nil
}

// countTargets counts the total number of targets in the unified configuration.
func (m *Migrator) countTargets(config *UnifiedConfig) int {
	count := 0
	for _, provider := range config.Providers {
		count += len(provider.Organizations)
	}

	return count
}

// saveUnifiedConfig saves the unified configuration to a file.
func (m *Migrator) saveUnifiedConfig(config *UnifiedConfig) error {
	// Ensure target directory exists
	if err := CreateDirectory(filepath.Dir(m.TargetPath)); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	// Add header comment
	header := fmt.Sprintf(`# gzh-manager unified configuration
# Generated by migration from: %s
# Migration date: %s
# This file replaces the legacy bulk-clone.yaml format
# See https://github.com/gizzahub/gzh-manager-go/docs/configuration.md for documentation

`, m.SourcePath, time.Now().Format("2006-01-02 15:04:05"))

	content := header + string(data)

	// Write to file
	if err := WriteFile(m.TargetPath, content); err != nil {
		return fmt.Errorf("failed to write configuration file: %w", err)
	}

	return nil
}

// generateMigrationReport generates a detailed migration report.
func (m *Migrator) generateMigrationReport(result *MigrationResult) string {
	var report strings.Builder

	report.WriteString("# Configuration Migration Report\n\n")
	report.WriteString(fmt.Sprintf("**Migration Date:** %s\n", time.Now().Format("2006-01-02 15:04:05")))
	report.WriteString(fmt.Sprintf("**Source:** %s\n", result.SourcePath))
	report.WriteString(fmt.Sprintf("**Target:** %s\n", result.TargetPath))
	report.WriteString(fmt.Sprintf("**Status:** %s\n", map[bool]string{true: "SUCCESS", false: "FAILED"}[result.Success]))
	report.WriteString(fmt.Sprintf("**Migrated Targets:** %d\n\n", result.MigratedTargets))

	if result.BackupPath != "" {
		report.WriteString(fmt.Sprintf("**Backup Created:** %s\n\n", result.BackupPath))
	}

	if len(result.Warnings) > 0 {
		report.WriteString("## Warnings\n\n")

		for _, warning := range result.Warnings {
			report.WriteString(fmt.Sprintf("- %s\n", warning))
		}

		report.WriteString("\n")
	}

	if len(result.RequiredActions) > 0 {
		report.WriteString("## Required Actions\n\n")

		for _, action := range result.RequiredActions {
			report.WriteString(fmt.Sprintf("- [ ] %s\n", action))
		}

		report.WriteString("\n")
	}

	report.WriteString("## Migration Details\n\n")
	report.WriteString("The following changes were made:\n\n")
	report.WriteString("1. **Version updated:** Legacy version converted to 1.0.0\n")
	report.WriteString("2. **Authentication:** Protocol-based auth converted to token-based\n")
	report.WriteString("3. **Structure:** Reorganized into provider-specific sections\n")
	report.WriteString("4. **Ignore patterns:** Moved to per-organization exclude patterns\n")
	report.WriteString("5. **New features:** Added support for visibility filtering, strategies, and advanced options\n\n")

	report.WriteString("## Next Steps\n\n")
	report.WriteString("1. Review the generated configuration file\n")
	report.WriteString("2. Configure authentication tokens (see Required Actions above)\n")
	report.WriteString("3. Test the configuration with `gz bulk-clone --dry-run`\n")
	report.WriteString("4. Remove the legacy configuration file when satisfied\n")

	return report.String()
}

// DetectLegacyFormat detects if a file is in legacy bulk-clone.yaml format.
func DetectLegacyFormat(configPath string) (bool, error) {
	if !FileExists(configPath) {
		return false, fmt.Errorf("config file not found: %s", configPath)
	}

	data, err := os.ReadFile(configPath) //nolint:gosec // configPath is user-provided configuration file path
	if err != nil {
		return false, fmt.Errorf("failed to read config file: %w", err)
	}

	// Check for legacy format markers
	content := string(data)

	// Legacy format has repo_roots and version 0.1
	hasRepoRoots := strings.Contains(content, "repo_roots:")
	hasLegacyVersion := strings.Contains(content, "version: \"0.1\"") || strings.Contains(content, "version: '0.1'")
	hasIgnoreNames := strings.Contains(content, "ignore_names:")

	// Unified format has providers and version 1.0.0
	hasProviders := strings.Contains(content, "providers:")
	hasUnifiedVersion := strings.Contains(content, "version: \"1.0.0\"") || strings.Contains(content, "version: '1.0.0'")

	// If it has the old markers and not the new ones, it's legacy
	if (hasRepoRoots || hasLegacyVersion || hasIgnoreNames) && !hasProviders && !hasUnifiedVersion {
		return true, nil
	}

	return false, nil
}

// MigrateConfigFile migrates a configuration file from legacy to unified format.
func MigrateConfigFile(sourcePath, targetPath string, dryRun bool) (*MigrationResult, error) {
	migrator := NewMigrator(sourcePath, targetPath)
	migrator.DryRun = dryRun

	return migrator.MigrateFromBulkClone()
}

// AutoMigrate automatically migrates configuration if legacy format is detected.
func AutoMigrate(configPath string) (*MigrationResult, error) {
	isLegacy, err := DetectLegacyFormat(configPath)
	if err != nil {
		return nil, err
	}

	if !isLegacy {
		return &MigrationResult{
			Success:    true,
			SourcePath: configPath,
			TargetPath: configPath,
		}, nil // No migration needed
	}

	// Determine target path (same directory, different name)
	dir := filepath.Dir(configPath)
	targetPath := filepath.Join(dir, "gzh.yaml")

	// If target already exists, create a versioned name
	if FileExists(targetPath) {
		targetPath = filepath.Join(dir, fmt.Sprintf("gzh.migrated.%s.yaml", time.Now().Format("20060102-150405")))
	}

	return MigrateConfigFile(configPath, targetPath, false)
}
