// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package cloud

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"time"
)

// DefaultSyncManager implements SyncManager interface.
type DefaultSyncManager struct {
	config      *Config
	syncHistory []SyncStatus
	statePath   string
}

// NewSyncManager creates a new sync manager instance.
func NewSyncManager(config *Config) SyncManager {
	statePath := getSyncStatePath()
	manager := &DefaultSyncManager{
		config:    config,
		statePath: statePath,
	}

	// Load existing sync history
	manager.loadSyncHistory()

	return manager
}

// SyncProfiles synchronizes specific profiles between providers.
func (sm *DefaultSyncManager) SyncProfiles(ctx context.Context, source, target Provider, profileNames []string) error {
	if len(profileNames) == 0 {
		return fmt.Errorf("no profiles specified for sync")
	}

	var allConflicts []SyncConflict

	syncResults := make([]SyncStatus, 0, len(profileNames))

	for _, profileName := range profileNames {
		status := SyncStatus{
			ProfileName: profileName,
			Source:      source.Name(),
			Target:      target.Name(),
			Status:      "pending",
			LastSync:    time.Now(),
		}

		// Get source profile
		sourceProfile, err := source.GetProfile(ctx, profileName)
		if err != nil {
			status.Status = "error"
			status.Error = fmt.Sprintf("failed to get source profile: %v", err)
			syncResults = append(syncResults, status)

			continue
		}

		// Get target profile (if exists)
		targetProfile, err := target.GetProfile(ctx, profileName)
		if err != nil {
			// Profile doesn't exist in target - create new
			targetProfile = &Profile{
				Name:        profileName,
				Provider:    target.Name(),
				Environment: sourceProfile.Environment,
				Region:      sourceProfile.Region,
			}
		}

		// Detect conflicts
		conflicts := sm.detectConflicts(profileName, sourceProfile, targetProfile)
		if len(conflicts) > 0 {
			allConflicts = append(allConflicts, conflicts...)

			// Apply conflict resolution strategy
			strategy := sm.config.Sync.ConflictMode
			if strategy == "" {
				strategy = ConflictStrategyAsk
			}

			if err := sm.ResolveSyncConflicts(conflicts, strategy); err != nil {
				status.Status = "conflict"
				status.Error = fmt.Sprintf("conflict resolution failed: %v", err)
				syncResults = append(syncResults, status)

				continue
			}
		}

		// Merge source profile into target
		mergedProfile := sm.mergeProfiles(sourceProfile, targetProfile, target.Name())

		// Sync to target provider
		if err := target.SyncProfile(ctx, mergedProfile); err != nil {
			status.Status = "error"
			status.Error = fmt.Sprintf("failed to sync to target: %v", err)
		} else {
			status.Status = "synced"
		}

		syncResults = append(syncResults, status)
	}

	// Update sync history
	sm.updateSyncHistory(syncResults)
	sm.saveSyncHistory()

	// Return error if any sync failed
	for _, result := range syncResults {
		if result.Status == "error" || result.Status == "conflict" {
			return fmt.Errorf("sync completed with errors - check sync status for details")
		}
	}

	return nil
}

// SyncAll synchronizes all profiles between providers.
func (sm *DefaultSyncManager) SyncAll(ctx context.Context, source, target Provider) error {
	// Get all profiles from source provider
	sourceProfiles, err := source.ListProfiles(ctx)
	if err != nil {
		return fmt.Errorf("failed to list source profiles: %w", err)
	}

	if len(sourceProfiles) == 0 {
		return fmt.Errorf("no profiles found in source provider")
	}

	// Extract profile names
	profileNames := make([]string, len(sourceProfiles))
	for i, profile := range sourceProfiles {
		profileNames[i] = profile.Name
	}

	return sm.SyncProfiles(ctx, source, target, profileNames)
}

// GetSyncStatus returns sync status for profiles.
func (sm *DefaultSyncManager) GetSyncStatus(ctx context.Context) ([]SyncStatus, error) {
	return sm.syncHistory, nil
}

// ResolveSyncConflicts resolves conflicts during sync.
func (sm *DefaultSyncManager) ResolveSyncConflicts(conflicts []SyncConflict, strategy ConflictStrategy) error {
	for i := range conflicts {
		conflict := &conflicts[i]

		switch strategy {
		case ConflictStrategySourceWins:
			// Source value takes precedence - no action needed as merge will use source
			continue

		case ConflictStrategyTargetWins:
			// Target value takes precedence - swap values
			conflict.SourceValue = conflict.TargetValue

		case ConflictStrategyMerge:
			// Attempt to merge values
			merged, err := sm.mergeValues(conflict.SourceValue, conflict.TargetValue)
			if err != nil {
				return fmt.Errorf("failed to merge conflict for field %s: %w", conflict.Field, err)
			}

			conflict.SourceValue = merged

		case ConflictStrategyAsk:
			// For now, default to source wins in automated mode
			// In interactive mode, this would prompt the user
			fmt.Printf("Conflict detected for profile %s, field %s:\n", conflict.ProfileName, conflict.Field)
			fmt.Printf("  Source: %v\n", conflict.SourceValue)
			fmt.Printf("  Target: %v\n", conflict.TargetValue)
			fmt.Printf("  Resolution: Using source value (automated mode)\n")

			continue

		default:
			return fmt.Errorf("unsupported conflict strategy: %s", strategy)
		}
	}

	return nil
}

// detectConflicts compares source and target profiles to detect conflicts.
func (sm *DefaultSyncManager) detectConflicts(profileName string, source, target *Profile) []SyncConflict {
	var conflicts []SyncConflict

	// Compare basic fields
	if source.Environment != target.Environment && target.Environment != "" {
		conflicts = append(conflicts, SyncConflict{
			ProfileName: profileName,
			Field:       "environment",
			SourceValue: source.Environment,
			TargetValue: target.Environment,
		})
	}

	if source.Region != target.Region && target.Region != "" {
		conflicts = append(conflicts, SyncConflict{
			ProfileName: profileName,
			Field:       "region",
			SourceValue: source.Region,
			TargetValue: target.Region,
		})
	}

	// Compare network configuration
	if source.Network.VPCId != target.Network.VPCId && target.Network.VPCId != "" {
		conflicts = append(conflicts, SyncConflict{
			ProfileName: profileName,
			Field:       "network.vpc_id",
			SourceValue: source.Network.VPCId,
			TargetValue: target.Network.VPCId,
		})
	}

	// Compare DNS servers
	if !equalStringSlices(source.Network.DNSServers, target.Network.DNSServers) && len(target.Network.DNSServers) > 0 {
		conflicts = append(conflicts, SyncConflict{
			ProfileName: profileName,
			Field:       "network.dns_servers",
			SourceValue: source.Network.DNSServers,
			TargetValue: target.Network.DNSServers,
		})
	}

	// Compare proxy configuration
	if !equalProxyConfig(source.Network.Proxy, target.Network.Proxy) && target.Network.Proxy != nil {
		conflicts = append(conflicts, SyncConflict{
			ProfileName: profileName,
			Field:       "network.proxy",
			SourceValue: source.Network.Proxy,
			TargetValue: target.Network.Proxy,
		})
	}

	// Compare tags
	if !equalStringMaps(source.Tags, target.Tags) && len(target.Tags) > 0 {
		conflicts = append(conflicts, SyncConflict{
			ProfileName: profileName,
			Field:       "tags",
			SourceValue: source.Tags,
			TargetValue: target.Tags,
		})
	}

	return conflicts
}

// mergeProfiles merges source profile into target with provider-specific adjustments.
func (sm *DefaultSyncManager) mergeProfiles(source, target *Profile, targetProvider string) *Profile {
	merged := &Profile{
		Name:        source.Name,
		Provider:    targetProvider,
		Environment: source.Environment,
		Region:      source.Region,
		Network:     source.Network,
		Services:    make(map[string]ServiceConfig),
		Tags:        make(map[string]string),
		LastSync:    time.Now(),
	}

	// Merge services
	for k, v := range source.Services {
		merged.Services[k] = v
	}

	for k, v := range target.Services {
		if _, exists := merged.Services[k]; !exists {
			merged.Services[k] = v
		}
	}

	// Merge tags
	for k, v := range source.Tags {
		merged.Tags[k] = v
	}

	for k, v := range target.Tags {
		if _, exists := merged.Tags[k]; !exists {
			merged.Tags[k] = v
		}
	}

	// Add sync metadata
	merged.Tags["sync_source"] = source.Provider
	merged.Tags["sync_timestamp"] = time.Now().Format(time.RFC3339)

	return merged
}

// mergeValues attempts to merge two values intelligently.
func (sm *DefaultSyncManager) mergeValues(source, target interface{}) (interface{}, error) {
	sourceType := reflect.TypeOf(source)
	targetType := reflect.TypeOf(target)

	if sourceType != targetType {
		return source, nil // Use source if types don't match
	}

	switch s := source.(type) {
	case []string:
		t := target.([]string)
		// Merge string slices, removing duplicates
		merged := make([]string, 0, len(s)+len(t))
		seen := make(map[string]bool)

		for _, v := range s {
			if !seen[v] {
				merged = append(merged, v)
				seen[v] = true
			}
		}

		for _, v := range t {
			if !seen[v] {
				merged = append(merged, v)
				seen[v] = true
			}
		}

		return merged, nil

	case map[string]string:
		t := target.(map[string]string)
		// Merge maps, source takes precedence
		merged := make(map[string]string)
		for k, v := range t {
			merged[k] = v
		}

		for k, v := range s {
			merged[k] = v
		}

		return merged, nil

	default:
		// For other types, use source value
		return source, nil
	}
}

// Helper functions.
func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func equalStringMaps(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}

	for k, v := range a {
		if b[k] != v {
			return false
		}
	}

	return true
}

func equalProxyConfig(a, b *ProxyConfig) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	return a.HTTP == b.HTTP && a.HTTPS == b.HTTPS && equalStringSlices(a.NoProxy, b.NoProxy)
}

// updateSyncHistory updates the sync history with new results.
func (sm *DefaultSyncManager) updateSyncHistory(results []SyncStatus) {
	for _, result := range results {
		// Find existing entry or add new one
		found := false

		for i := range sm.syncHistory {
			if sm.syncHistory[i].ProfileName == result.ProfileName &&
				sm.syncHistory[i].Source == result.Source &&
				sm.syncHistory[i].Target == result.Target {
				sm.syncHistory[i] = result
				found = true

				break
			}
		}

		if !found {
			sm.syncHistory = append(sm.syncHistory, result)
		}
	}
}

// loadSyncHistory loads sync history from disk.
func (sm *DefaultSyncManager) loadSyncHistory() {
	if sm.statePath == "" {
		return
	}

	data, err := os.ReadFile(sm.statePath)
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Printf("Warning: failed to load sync history: %v\n", err)
		}

		return
	}

	var history []SyncStatus
	if err := json.Unmarshal(data, &history); err != nil {
		fmt.Printf("Warning: failed to parse sync history: %v\n", err)
		return
	}

	sm.syncHistory = history
}

// saveSyncHistory saves sync history to disk.
func (sm *DefaultSyncManager) saveSyncHistory() {
	if sm.statePath == "" {
		return
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(sm.statePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		fmt.Printf("Warning: failed to create sync state directory: %v\n", err)
		return
	}

	data, err := json.MarshalIndent(sm.syncHistory, "", "  ")
	if err != nil {
		fmt.Printf("Warning: failed to marshal sync history: %v\n", err)
		return
	}

	if err := os.WriteFile(sm.statePath, data, 0o644); err != nil {
		fmt.Printf("Warning: failed to save sync history: %v\n", err)
	}
}

// getSyncStatePath returns the path for sync state storage.
func getSyncStatePath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}

	return filepath.Join(configDir, "gzh-manager", "cloud-sync-state.json")
}

// ValidateSyncConfig validates sync configuration.
func ValidateSyncConfig(config *Config) error {
	if !config.Sync.Enabled {
		return nil
	}

	// Validate conflict strategy
	validStrategies := []ConflictStrategy{
		ConflictStrategySourceWins,
		ConflictStrategyTargetWins,
		ConflictStrategyMerge,
		ConflictStrategyAsk,
	}

	valid := false

	for _, strategy := range validStrategies {
		if config.Sync.ConflictMode == strategy {
			valid = true
			break
		}
	}

	if !valid && config.Sync.ConflictMode != "" {
		return fmt.Errorf("invalid conflict strategy: %s", config.Sync.ConflictMode)
	}

	// Validate sync targets
	for i, target := range config.Sync.Targets {
		if target.Source == "" {
			return fmt.Errorf("sync target %d: source provider is required", i)
		}

		if target.Target == "" {
			return fmt.Errorf("sync target %d: target provider is required", i)
		}

		if target.Source == target.Target {
			return fmt.Errorf("sync target %d: source and target cannot be the same", i)
		}

		// Check if providers exist
		if _, exists := config.Providers[target.Source]; !exists {
			return fmt.Errorf("sync target %d: source provider %s not found", i, target.Source)
		}

		if _, exists := config.Providers[target.Target]; !exists {
			return fmt.Errorf("sync target %d: target provider %s not found", i, target.Target)
		}
	}

	return nil
}

// GetSyncRecommendations analyzes profiles and suggests sync targets.
func GetSyncRecommendations(config *Config) ([]SyncTarget, error) {
	var recommendations []SyncTarget

	// Group profiles by environment
	envProfiles := make(map[string][]string)

	for name, profile := range config.Profiles {
		env := profile.Environment
		if env == "" {
			env = "default"
		}

		envProfiles[env] = append(envProfiles[env], name)
	}

	// Suggest sync targets for environments with multiple providers
	providersByEnv := make(map[string]map[string]bool)

	for _, profile := range config.Profiles {
		env := profile.Environment
		if env == "" {
			env = "default"
		}

		if providersByEnv[env] == nil {
			providersByEnv[env] = make(map[string]bool)
		}

		providersByEnv[env][profile.Provider] = true
	}

	for env, providers := range providersByEnv {
		if len(providers) > 1 {
			// Multiple providers for same environment - suggest sync
			var providerList []string
			for provider := range providers {
				providerList = append(providerList, provider)
			}

			// Create bidirectional sync recommendations
			for i := 0; i < len(providerList); i++ {
				for j := i + 1; j < len(providerList); j++ {
					recommendations = append(recommendations, SyncTarget{
						Source:   providerList[i],
						Target:   providerList[j],
						Profiles: envProfiles[env],
					})
				}
			}
		}
	}

	return recommendations, nil
}
