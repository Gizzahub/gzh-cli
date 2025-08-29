package upgrade

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Gizzahub/gzh-cli/internal/logger"
)

// UpgradeCoordinator coordinates upgrades across multiple package managers
type UpgradeCoordinator struct {
	upgraders map[string]PackageManagerUpgrader
	logger    logger.CommonLogger
	backupDir string
	mu        sync.RWMutex
}

// NewUpgradeCoordinator creates a new upgrade coordinator
func NewUpgradeCoordinator(logger logger.CommonLogger, backupDir string) *UpgradeCoordinator {
	coordinator := &UpgradeCoordinator{
		upgraders: make(map[string]PackageManagerUpgrader),
		logger:    logger,
		backupDir: backupDir,
	}

	// Register default upgraders
	coordinator.registerDefaultUpgraders()

	return coordinator
}

// registerDefaultUpgraders registers all available upgraders
func (uc *UpgradeCoordinator) registerDefaultUpgraders() {
	uc.RegisterUpgrader("homebrew", NewHomebrewUpgrader(uc.logger))
	uc.RegisterUpgrader("brew", NewHomebrewUpgrader(uc.logger)) // Alias
	uc.RegisterUpgrader("asdf", NewAsdfUpgrader(uc.logger))
	uc.RegisterUpgrader("nvm", NewNvmUpgrader(uc.logger))
	uc.RegisterUpgrader("rbenv", NewRbenvUpgrader(uc.logger))
	uc.RegisterUpgrader("pyenv", NewPyenvUpgrader(uc.logger))
	uc.RegisterUpgrader("sdkman", NewSdkmanUpgrader(uc.logger))
}

// RegisterUpgrader registers a package manager upgrader
func (uc *UpgradeCoordinator) RegisterUpgrader(name string, upgrader PackageManagerUpgrader) {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	uc.upgraders[name] = upgrader
}

// GetUpgrader returns an upgrader by name
func (uc *UpgradeCoordinator) GetUpgrader(name string) (PackageManagerUpgrader, bool) {
	uc.mu.RLock()
	defer uc.mu.RUnlock()
	upgrader, exists := uc.upgraders[name]
	return upgrader, exists
}

// ListUpgraders returns all registered upgrader names
func (uc *UpgradeCoordinator) ListUpgraders() []string {
	uc.mu.RLock()
	defer uc.mu.RUnlock()

	names := make([]string, 0, len(uc.upgraders))
	for name := range uc.upgraders {
		names = append(names, name)
	}
	return names
}

// CheckAll generates a comprehensive upgrade report for all registered package managers
func (uc *UpgradeCoordinator) CheckAll(ctx context.Context) (*UpgradeReport, error) {
	uc.logger.Info("Checking upgrade status for all package managers")

	uc.mu.RLock()
	defer uc.mu.RUnlock()

	statuses := make([]UpgradeStatus, 0, len(uc.upgraders))
	updatesNeeded := 0

	for name, upgrader := range uc.upgraders {
		uc.logger.Debug("Checking upgrade status for: %s", name)

		status, err := upgrader.CheckUpdate(ctx)
		if err != nil {
			uc.logger.Warn("Failed to check update for %s: %v", name, err)
			// Create a status with error information
			status = &UpgradeStatus{
				Manager:         name,
				CurrentVersion:  "unknown",
				LatestVersion:   "unknown",
				UpdateAvailable: false,
				UpdateMethod:    upgrader.GetUpdateMethod(),
			}
		}

		statuses = append(statuses, *status)
		if status.UpdateAvailable {
			updatesNeeded++
		}
	}

	return &UpgradeReport{
		Platform:      detectPlatform(),
		TotalManagers: len(uc.upgraders),
		UpdatesNeeded: updatesNeeded,
		Managers:      statuses,
		Timestamp:     time.Now(),
	}, nil
}

// CheckManagers checks upgrade status for specific package managers
func (uc *UpgradeCoordinator) CheckManagers(ctx context.Context, names []string) (*UpgradeReport, error) {
	uc.logger.Info("Checking upgrade status for managers: %v", names)

	statuses := make([]UpgradeStatus, 0, len(names))
	updatesNeeded := 0

	for _, name := range names {
		upgrader, exists := uc.GetUpgrader(name)
		if !exists {
			uc.logger.Warn("Unknown package manager: %s", name)
			status := UpgradeStatus{
				Manager:         name,
				CurrentVersion:  "unknown",
				LatestVersion:   "unknown",
				UpdateAvailable: false,
				UpdateMethod:    "unknown",
			}
			statuses = append(statuses, status)
			continue
		}

		uc.logger.Debug("Checking upgrade status for: %s", name)
		status, err := upgrader.CheckUpdate(ctx)
		if err != nil {
			uc.logger.Warn("Failed to check update for %s: %v", name, err)
			status = &UpgradeStatus{
				Manager:         name,
				CurrentVersion:  "unknown",
				LatestVersion:   "unknown",
				UpdateAvailable: false,
				UpdateMethod:    upgrader.GetUpdateMethod(),
			}
		}

		statuses = append(statuses, *status)
		if status.UpdateAvailable {
			updatesNeeded++
		}
	}

	return &UpgradeReport{
		Platform:      detectPlatform(),
		TotalManagers: len(names),
		UpdatesNeeded: updatesNeeded,
		Managers:      statuses,
		Timestamp:     time.Now(),
	}, nil
}

// UpgradeAll upgrades all package managers
func (uc *UpgradeCoordinator) UpgradeAll(ctx context.Context, options UpgradeOptions) (*UpgradeReport, error) {
	uc.logger.Info("Starting upgrade for all package managers")

	uc.mu.RLock()
	names := make([]string, 0, len(uc.upgraders))
	for name := range uc.upgraders {
		names = append(names, name)
	}
	uc.mu.RUnlock()

	return uc.UpgradeManagers(ctx, names, options)
}

// UpgradeManagers upgrades specific package managers
func (uc *UpgradeCoordinator) UpgradeManagers(ctx context.Context, names []string, options UpgradeOptions) (*UpgradeReport, error) {
	uc.logger.Info("Starting upgrade for managers: %v", names)

	startTime := time.Now()
	statuses := make([]UpgradeStatus, 0, len(names))
	successCount := 0
	failureCount := 0

	for _, name := range names {
		upgrader, exists := uc.GetUpgrader(name)
		if !exists {
			uc.logger.Error("Unknown package manager: %s", name)
			status := UpgradeStatus{
				Manager:         name,
				CurrentVersion:  "unknown",
				LatestVersion:   "unknown",
				UpdateAvailable: false,
				UpdateMethod:    "unknown",
			}
			statuses = append(statuses, status)
			failureCount++
			continue
		}

		uc.logger.Info("Upgrading %s...", name)

		// Check current status before upgrade
		preStatus, err := upgrader.CheckUpdate(ctx)
		if err != nil {
			uc.logger.Warn("Failed to check pre-upgrade status for %s: %v", name, err)
			preStatus = &UpgradeStatus{
				Manager:         name,
				CurrentVersion:  "unknown",
				LatestVersion:   "unknown",
				UpdateAvailable: true,
				UpdateMethod:    upgrader.GetUpdateMethod(),
			}
		}

		// Perform upgrade
		if err := upgrader.Upgrade(ctx, options); err != nil {
			uc.logger.Error("Failed to upgrade %s: %v", name, err)
			failureCount++

			// Add failed status
			failedStatus := *preStatus
			failedStatus.UpdateAvailable = false // Mark as failed
			statuses = append(statuses, failedStatus)
			continue
		}

		// Check post-upgrade status
		postStatus, err := upgrader.CheckUpdate(ctx)
		if err != nil {
			uc.logger.Warn("Failed to check post-upgrade status for %s: %v", name, err)
			postStatus = preStatus
		}

		statuses = append(statuses, *postStatus)
		successCount++
		uc.logger.Info("Successfully upgraded %s", name)
	}

	duration := time.Since(startTime)
	uc.logger.Info("Upgrade completed in %v. Success: %d, Failed: %d", duration, successCount, failureCount)

	return &UpgradeReport{
		Platform:      detectPlatform(),
		TotalManagers: len(names),
		UpdatesNeeded: failureCount, // Reuse field to indicate failures
		Managers:      statuses,
		Timestamp:     time.Now(),
	}, nil
}

// GetAvailableManagers returns a list of all available package managers
func (uc *UpgradeCoordinator) GetAvailableManagers() []string {
	return uc.ListUpgraders()
}

// detectPlatform detects the current platform
func detectPlatform() string {
	// This is a simple implementation - could be enhanced
	return "linux" // Default for now
}

// FormatReport formats an upgrade report for display
func (uc *UpgradeCoordinator) FormatReport(report *UpgradeReport, verbose bool) string {
	if report == nil {
		return "No upgrade report available\n"
	}

	result := "📊 Package Manager Upgrade Report\n" // S1039 수정: 불필요한 fmt.Sprintf 제거
	result += fmt.Sprintf("Platform: %s\n", report.Platform)
	result += fmt.Sprintf("Timestamp: %s\n", report.Timestamp.Format("2006-01-02 15:04:05"))
	result += fmt.Sprintf("Total Managers: %d\n", report.TotalManagers)
	result += fmt.Sprintf("Updates Available: %d\n\n", report.UpdatesNeeded)

	for _, status := range report.Managers {
		icon := "✅"
		if status.UpdateAvailable {
			icon = "🆙"
		}

		result += fmt.Sprintf("%s %s: %s", icon, status.Manager, status.CurrentVersion)

		if status.UpdateAvailable && status.LatestVersion != "" {
			result += fmt.Sprintf(" → %s", status.LatestVersion)
		}

		if verbose {
			result += fmt.Sprintf(" (%s)", status.UpdateMethod)
		}

		result += "\n"
	}

	return result
}
