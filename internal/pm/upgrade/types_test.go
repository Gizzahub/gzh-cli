package upgrade

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUpgradeStatus_Validation(t *testing.T) {
	status := UpgradeStatus{
		Manager:         "homebrew",
		CurrentVersion:  "3.0.0",
		LatestVersion:   "3.1.0",
		UpdateAvailable: true,
		UpdateMethod:    "brew update",
		ReleaseDate:     time.Now(),
		ChangelogURL:    "https://github.com/Homebrew/brew/releases",
		Size:            1024 * 1024, // 1MB
	}

	assert.Equal(t, "homebrew", status.Manager)
	assert.Equal(t, "3.0.0", status.CurrentVersion)
	assert.Equal(t, "3.1.0", status.LatestVersion)
	assert.True(t, status.UpdateAvailable)
	assert.Equal(t, "brew update", status.UpdateMethod)
	assert.False(t, status.ReleaseDate.IsZero())
	assert.NotEmpty(t, status.ChangelogURL)
	assert.Greater(t, status.Size, int64(0))
}

func TestUpgradeReport_Validation(t *testing.T) {
	now := time.Now()

	report := UpgradeReport{
		Platform:      "linux",
		TotalManagers: 3,
		UpdatesNeeded: 2,
		Managers: []UpgradeStatus{
			{
				Manager:         "homebrew",
				CurrentVersion:  "3.0.0",
				LatestVersion:   "3.1.0",
				UpdateAvailable: true,
				UpdateMethod:    "brew update",
			},
			{
				Manager:         "asdf",
				CurrentVersion:  "0.10.0",
				LatestVersion:   "0.11.0",
				UpdateAvailable: true,
				UpdateMethod:    "git pull",
			},
			{
				Manager:         "nvm",
				CurrentVersion:  "0.39.0",
				LatestVersion:   "0.39.0",
				UpdateAvailable: false,
				UpdateMethod:    "curl install script",
			},
		},
		Timestamp: now,
	}

	assert.Equal(t, "linux", report.Platform)
	assert.Equal(t, 3, report.TotalManagers)
	assert.Equal(t, 2, report.UpdatesNeeded)
	assert.Len(t, report.Managers, 3)
	assert.Equal(t, now, report.Timestamp)

	// Verify specific managers
	assert.Equal(t, "homebrew", report.Managers[0].Manager)
	assert.True(t, report.Managers[0].UpdateAvailable)
	assert.Equal(t, "asdf", report.Managers[1].Manager)
	assert.True(t, report.Managers[1].UpdateAvailable)
	assert.Equal(t, "nvm", report.Managers[2].Manager)
	assert.False(t, report.Managers[2].UpdateAvailable)
}

func TestUpgradeOptions_DefaultValues(t *testing.T) {
	options := UpgradeOptions{}

	assert.False(t, options.Force)
	assert.False(t, options.PreRelease)
	assert.False(t, options.BackupEnabled)
	assert.False(t, options.SkipValidation)
	assert.Equal(t, time.Duration(0), options.Timeout)
}

func TestUpgradeOptions_Configuration(t *testing.T) {
	options := UpgradeOptions{
		Force:          true,
		PreRelease:     true,
		BackupEnabled:  true,
		SkipValidation: false,
		Timeout:        5 * time.Minute,
	}

	assert.True(t, options.Force)
	assert.True(t, options.PreRelease)
	assert.True(t, options.BackupEnabled)
	assert.False(t, options.SkipValidation)
	assert.Equal(t, 5*time.Minute, options.Timeout)
}

func TestNewUpgradeManager(t *testing.T) {
	mockLogger := &MockLogger{}
	backupDir := "/tmp/test-backups"

	manager := NewUpgradeManager(mockLogger, backupDir)

	assert.NotNil(t, manager)
	assert.Equal(t, mockLogger, manager.logger)
	assert.Equal(t, backupDir, manager.backupDir)
	assert.NotNil(t, manager.upgraders)
	assert.Empty(t, manager.upgraders) // NewUpgradeManager doesn't register upgraders
}

func TestUpgradeManager_RegisterUpgrader(t *testing.T) {
	mockLogger := &MockLogger{}
	manager := NewUpgradeManager(mockLogger, "/tmp")

	mockUpgrader := &MockUpgrader{}
	manager.RegisterUpgrader("test-manager", mockUpgrader)

	upgrader, exists := manager.GetUpgrader("test-manager")
	assert.True(t, exists)
	assert.Equal(t, mockUpgrader, upgrader)
}

func TestUpgradeManager_GetUpgrader_NotFound(t *testing.T) {
	mockLogger := &MockLogger{}
	manager := NewUpgradeManager(mockLogger, "/tmp")

	upgrader, exists := manager.GetUpgrader("non-existent")
	assert.False(t, exists)
	assert.Nil(t, upgrader)
}

func TestUpgradeManager_ListUpgraders(t *testing.T) {
	mockLogger := &MockLogger{}
	manager := NewUpgradeManager(mockLogger, "/tmp")

	// Initially empty
	upgraders := manager.ListUpgraders()
	assert.Empty(t, upgraders)

	// Add some upgraders
	mockUpgrader1 := &MockUpgrader{}
	mockUpgrader2 := &MockUpgrader{}
	manager.RegisterUpgrader("manager1", mockUpgrader1)
	manager.RegisterUpgrader("manager2", mockUpgrader2)

	upgraders = manager.ListUpgraders()
	assert.Len(t, upgraders, 2)
	assert.Contains(t, upgraders, "manager1")
	assert.Contains(t, upgraders, "manager2")
}

func TestDetectPlatform(t *testing.T) {
	platform := detectPlatform()

	// Should return a non-empty string
	assert.NotEmpty(t, platform)
	// Currently hardcoded to "linux" in the implementation
	assert.Equal(t, "linux", platform)
}
