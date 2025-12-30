// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package sync

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/gizzahub/gzh-cli/internal/logger"
)

const (
	statusUnknown = "unknown"
)

// NewSyncManager creates a new synchronization manager.
func NewSyncManager(logger logger.CommonLogger) *SyncManager {
	manager := &SyncManager{
		synchronizers: make(map[string]VersionSynchronizer),
		policy: SyncPolicy{
			Strategy:      "vm_priority",
			AutoFix:       false,
			BackupEnabled: true,
			PromptUser:    true,
		},
		logger: logger,
	}

	// Register default synchronizers
	manager.registerDefaultSynchronizers()

	return manager
}

// registerDefaultSynchronizers registers all available synchronizers.
func (sm *SyncManager) registerDefaultSynchronizers() {
	sm.RegisterSynchronizer("nvm-npm", NewNvmNpmSynchronizer(sm.logger))
	sm.RegisterSynchronizer("rbenv-gem", NewRbenvGemSynchronizer(sm.logger))
	sm.RegisterSynchronizer("pyenv-pip", NewPyenvPipSynchronizer(sm.logger))
	sm.RegisterSynchronizer("asdf-multi", NewAsdfMultiSynchronizer(sm.logger))
}

// RegisterSynchronizer registers a version synchronizer.
func (sm *SyncManager) RegisterSynchronizer(name string, synchronizer VersionSynchronizer) {
	sm.synchronizers[name] = synchronizer
}

// GetSynchronizer returns a synchronizer by name.
func (sm *SyncManager) GetSynchronizer(name string) (VersionSynchronizer, bool) {
	synchronizer, exists := sm.synchronizers[name]
	return synchronizer, exists
}

// ListSynchronizers returns all registered synchronizer names.
func (sm *SyncManager) ListSynchronizers() []string {
	names := make([]string, 0, len(sm.synchronizers))
	for name := range sm.synchronizers {
		names = append(names, name)
	}
	return names
}

// CheckAll generates a comprehensive sync report for all registered synchronizers.
func (sm *SyncManager) CheckAll(ctx context.Context) (*SyncReport, error) {
	sm.logger.Info("Checking version synchronization for all manager pairs")

	statuses := make([]VersionSyncStatus, 0, len(sm.synchronizers))
	inSyncCount := 0

	var wg sync.WaitGroup
	var mu sync.Mutex
	statusChan := make(chan VersionSyncStatus, len(sm.synchronizers))
	errorChan := make(chan error, len(sm.synchronizers))

	// Check all synchronizers concurrently
	for name, synchronizer := range sm.synchronizers {
		wg.Add(1)
		go func(name string, sync VersionSynchronizer) {
			defer wg.Done()

			sm.logger.Debug("Checking sync status for: %s", name)
			status, err := sync.CheckSync(ctx)
			if err != nil {
				sm.logger.Warn("Failed to check sync for %s: %v", name, err)
				// Create a status with error information
				vmName, pmName := sync.GetManagerPair()
				status = &VersionSyncStatus{
					VersionManager:    vmName,
					PackageManager:    pmName,
					VMVersion:         statusUnknown,
					PMVersion:         statusUnknown,
					ExpectedPMVersion: statusUnknown,
					InSync:            false,
					SyncAction:        "check_failed",
					Issues:            []string{err.Error()},
				}
			}

			statusChan <- *status
		}(name, synchronizer)
	}

	// Wait for all checks to complete
	go func() {
		wg.Wait()
		close(statusChan)
		close(errorChan)
	}()

	// Collect results
	for status := range statusChan {
		mu.Lock()
		statuses = append(statuses, status)
		if status.InSync {
			inSyncCount++
		}
		mu.Unlock()
	}

	return &SyncReport{
		Platform:       detectPlatform(),
		TotalPairs:     len(statuses),
		InSyncCount:    inSyncCount,
		OutOfSyncCount: len(statuses) - inSyncCount,
		SyncStatuses:   statuses,
		Timestamp:      time.Now(),
	}, nil
}

// CheckPairs checks synchronization status for specific manager pairs.
func (sm *SyncManager) CheckPairs(ctx context.Context, pairs []string) (*SyncReport, error) {
	sm.logger.Info("Checking sync status for pairs: %v", pairs)

	statuses := make([]VersionSyncStatus, 0, len(pairs))
	inSyncCount := 0

	for _, pair := range pairs {
		synchronizer, exists := sm.GetSynchronizer(pair)
		if !exists {
			sm.logger.Warn("Unknown synchronizer pair: %s", pair)
			vmName, pmName := parsePairName(pair)
			status := VersionSyncStatus{
				VersionManager:    vmName,
				PackageManager:    pmName,
				VMVersion:         statusUnknown,
				PMVersion:         statusUnknown,
				ExpectedPMVersion: statusUnknown,
				InSync:            false,
				SyncAction:        "unknown_pair",
				Issues:            []string{fmt.Sprintf("Unknown synchronizer pair: %s", pair)},
			}
			statuses = append(statuses, status)
			continue
		}

		sm.logger.Debug("Checking sync status for: %s", pair)
		status, err := synchronizer.CheckSync(ctx)
		if err != nil {
			sm.logger.Warn("Failed to check sync for %s: %v", pair, err)
			vmName, pmName := synchronizer.GetManagerPair()
			status = &VersionSyncStatus{
				VersionManager:    vmName,
				PackageManager:    pmName,
				VMVersion:         statusUnknown,
				PMVersion:         statusUnknown,
				ExpectedPMVersion: statusUnknown,
				InSync:            false,
				SyncAction:        "check_failed",
				Issues:            []string{err.Error()},
			}
		}

		statuses = append(statuses, *status)
		if status.InSync {
			inSyncCount++
		}
	}

	return &SyncReport{
		Platform:       detectPlatform(),
		TotalPairs:     len(statuses),
		InSyncCount:    inSyncCount,
		OutOfSyncCount: len(statuses) - inSyncCount,
		SyncStatuses:   statuses,
		Timestamp:      time.Now(),
	}, nil
}

// FixSynchronization fixes synchronization issues for specified pairs.
func (sm *SyncManager) FixSynchronization(ctx context.Context, pairs []string, policy SyncPolicy) (*SyncReport, error) {
	sm.logger.Info("Starting synchronization fix for pairs: %v", pairs)

	startTime := time.Now()
	statuses := make([]VersionSyncStatus, 0, len(pairs))
	successCount := 0
	failureCount := 0

	for _, pair := range pairs {
		synchronizer, exists := sm.GetSynchronizer(pair)
		if !exists {
			sm.logger.Error("Unknown synchronizer pair: %s", pair)
			vmName, pmName := parsePairName(pair)
			status := VersionSyncStatus{
				VersionManager:    vmName,
				PackageManager:    pmName,
				VMVersion:         statusUnknown,
				PMVersion:         statusUnknown,
				ExpectedPMVersion: statusUnknown,
				InSync:            false,
				SyncAction:        "failed",
				Issues:            []string{fmt.Sprintf("Unknown synchronizer pair: %s", pair)},
			}
			statuses = append(statuses, status)
			failureCount++
			continue
		}

		sm.logger.Info("Synchronizing %s...", pair)

		// Check current status before sync
		preStatus, err := synchronizer.CheckSync(ctx)
		if err != nil {
			sm.logger.Warn("Failed to check pre-sync status for %s: %v", pair, err)
			vmName, pmName := synchronizer.GetManagerPair()
			preStatus = &VersionSyncStatus{
				VersionManager:    vmName,
				PackageManager:    pmName,
				VMVersion:         statusUnknown,
				PMVersion:         statusUnknown,
				ExpectedPMVersion: statusUnknown,
				InSync:            false,
				SyncAction:        "pre_check_failed",
			}
		}

		// Perform synchronization
		if err := synchronizer.Synchronize(ctx, policy); err != nil {
			sm.logger.Error("Failed to synchronize %s: %v", pair, err)
			failureCount++

			// Add failed status
			failedStatus := *preStatus
			failedStatus.SyncAction = "failed"
			failedStatus.Issues = append(failedStatus.Issues, err.Error())
			statuses = append(statuses, failedStatus)
			continue
		}

		// Check post-sync status
		postStatus, err := synchronizer.CheckSync(ctx)
		if err != nil {
			sm.logger.Warn("Failed to check post-sync status for %s: %v", pair, err)
			postStatus = preStatus
			postStatus.SyncAction = "post_check_failed"
		} else {
			postStatus.SyncAction = "synchronized"
		}

		statuses = append(statuses, *postStatus)
		successCount++
		sm.logger.Info("Successfully synchronized %s", pair)
	}

	duration := time.Since(startTime)
	sm.logger.Info("Synchronization completed in %v. Success: %d, Failed: %d", duration, successCount, failureCount)

	return &SyncReport{
		Platform:       detectPlatform(),
		TotalPairs:     len(statuses),
		InSyncCount:    successCount,
		OutOfSyncCount: failureCount,
		SyncStatuses:   statuses,
		Timestamp:      time.Now(),
	}, nil
}

// GetAvailablePairs returns a list of all available synchronizer pairs.
func (sm *SyncManager) GetAvailablePairs() []string {
	return sm.ListSynchronizers()
}

// SetPolicy sets the default synchronization policy.
func (sm *SyncManager) SetPolicy(policy SyncPolicy) {
	sm.policy = policy
}

// GetPolicy returns the current default policy.
func (sm *SyncManager) GetPolicy() SyncPolicy {
	return sm.policy
}

// FormatReport formats a sync report for display.
func (sm *SyncManager) FormatReport(report *SyncReport, verbose bool) string {
	if report == nil {
		return "No synchronization report available\n"
	}

	result := "🔄 Package Manager Version Synchronization Status\n" // S1039 수정: 불필요한 fmt.Sprintf 제거
	result += fmt.Sprintf("Platform: %s\n", report.Platform)
	result += fmt.Sprintf("Timestamp: %s\n", report.Timestamp.Format("2006-01-02 15:04:05"))
	result += fmt.Sprintf("Total Pairs: %d\n", report.TotalPairs)
	result += fmt.Sprintf("In Sync: %d, Out of Sync: %d\n\n", report.InSyncCount, report.OutOfSyncCount)

	result += "Version Manager Pairs:\n"
	for _, status := range report.SyncStatuses {
		icon := "✅"
		if !status.InSync {
			icon = "❌"
		}

		result += fmt.Sprintf("  %s %s ↔ %s", icon, status.VersionManager, status.PackageManager)

		if status.VMVersion != statusUnknown {
			result += fmt.Sprintf("      %s v%s ↔ %s v%s", status.VersionManager, status.VMVersion, status.PackageManager, status.PMVersion)
		}

		if status.InSync {
			result += "     (in sync)\n"
		} else {
			result += "       (out of sync)\n"
			if status.ExpectedPMVersion != statusUnknown && status.ExpectedPMVersion != "" {
				result += fmt.Sprintf("     Expected %s version: v%s\n", status.PackageManager, status.ExpectedPMVersion)
			}
			if status.SyncAction != "" && status.SyncAction != statusUnknown {
				result += fmt.Sprintf("     Action needed: %s\n", status.SyncAction)
			}
		}

		if verbose && len(status.Issues) > 0 {
			result += fmt.Sprintf("     Issues: %s\n", strings.Join(status.Issues, ", "))
		}

		result += "\n"
	}

	if report.OutOfSyncCount > 0 {
		result += "Synchronization strategies:\n"
		result += "  --strategy vm_priority    Update package managers to match version managers\n"
		result += "  --strategy pm_priority    Update version managers to match package managers\n"
		result += "  --strategy latest         Update both to latest compatible versions\n\n"
	}

	return result
}

// detectPlatform detects the current platform.
func detectPlatform() string {
	return runtime.GOOS
}

// parsePairName parses a pair name like "nvm-npm" into ("nvm", "npm").
func parsePairName(pair string) (string, string) {
	parts := strings.Split(pair, "-")
	if len(parts) >= 2 {
		return parts[0], parts[1]
	}
	return pair, statusUnknown
}
