// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package upgrade

import (
	"context"
	"time"

	"github.com/gizzahub/gzh-cli/internal/logger"
)

// UpgradeStatus represents the current status of a package manager upgrade.
type UpgradeStatus struct {
	Manager         string    `json:"manager"`
	CurrentVersion  string    `json:"current_version"`
	LatestVersion   string    `json:"latest_version"`
	UpdateAvailable bool      `json:"update_available"`
	UpdateMethod    string    `json:"update_method"`
	ReleaseDate     time.Time `json:"release_date,omitempty"`
	ChangelogURL    string    `json:"changelog_url,omitempty"`
	Size            int64     `json:"size,omitempty"`
}

// UpgradeReport represents a comprehensive report of all package manager upgrade statuses.
type UpgradeReport struct {
	Platform      string          `json:"platform"`
	TotalManagers int             `json:"total_managers"`
	UpdatesNeeded int             `json:"updates_needed"`
	Managers      []UpgradeStatus `json:"managers"`
	Timestamp     time.Time       `json:"timestamp"`
}

// UpgradeOptions configures how an upgrade should be performed.
type UpgradeOptions struct {
	Force          bool          `json:"force"`
	PreRelease     bool          `json:"pre_release"`
	BackupEnabled  bool          `json:"backup_enabled"`
	SkipValidation bool          `json:"skip_validation"`
	Timeout        time.Duration `json:"timeout"`
}

// PackageManagerUpgrader defines the interface that all package manager upgraders must implement.
type PackageManagerUpgrader interface {
	// CheckUpdate checks if an update is available for the package manager
	CheckUpdate(ctx context.Context) (*UpgradeStatus, error)

	// Upgrade performs the actual upgrade of the package manager
	Upgrade(ctx context.Context, options UpgradeOptions) error

	// Backup creates a backup before upgrading, returns backup path
	Backup(ctx context.Context) (string, error)

	// Rollback restores from a backup if upgrade fails
	Rollback(ctx context.Context, backupPath string) error

	// GetUpdateMethod returns the method used to update this package manager
	GetUpdateMethod() string

	// ValidateUpgrade validates that the upgrade was successful
	ValidateUpgrade(ctx context.Context) error
}

// UpgradeManager coordinates upgrades across multiple package managers.
type UpgradeManager struct {
	upgraders map[string]PackageManagerUpgrader
	logger    logger.CommonLogger
	backupDir string
}

// NewUpgradeManager creates a new UpgradeManager instance.
func NewUpgradeManager(logger logger.CommonLogger, backupDir string) *UpgradeManager {
	return &UpgradeManager{
		upgraders: make(map[string]PackageManagerUpgrader),
		logger:    logger,
		backupDir: backupDir,
	}
}

// RegisterUpgrader registers a package manager upgrader.
func (um *UpgradeManager) RegisterUpgrader(name string, upgrader PackageManagerUpgrader) {
	um.upgraders[name] = upgrader
}

// GetUpgrader returns an upgrader by name.
func (um *UpgradeManager) GetUpgrader(name string) (PackageManagerUpgrader, bool) {
	upgrader, exists := um.upgraders[name]
	return upgrader, exists
}

// ListUpgraders returns all registered upgrader names.
func (um *UpgradeManager) ListUpgraders() []string {
	names := make([]string, 0, len(um.upgraders))
	for name := range um.upgraders {
		names = append(names, name)
	}
	return names
}

// GenerateReport generates a comprehensive upgrade report for all registered package managers.
func (um *UpgradeManager) GenerateReport(ctx context.Context) (*UpgradeReport, error) {
	statuses := make([]UpgradeStatus, 0, len(um.upgraders))
	updatesNeeded := 0

	for name, upgrader := range um.upgraders {
		status, err := upgrader.CheckUpdate(ctx)
		if err != nil {
			um.logger.Warn("Failed to check update for %s: %v", name, err)
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
		Platform:      "linux", // TODO: detect actual platform
		TotalManagers: len(um.upgraders),
		UpdatesNeeded: updatesNeeded,
		Managers:      statuses,
		Timestamp:     time.Now(),
	}, nil
}
