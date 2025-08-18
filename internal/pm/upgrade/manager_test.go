package upgrade

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockLogger implements logger.CommonLogger for testing
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *MockLogger) Info(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *MockLogger) Warn(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *MockLogger) Error(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *MockLogger) ErrorWithStack(err error, msg string, args ...interface{}) {
	m.Called(err, msg, args)
}

// MockUpgrader implements PackageManagerUpgrader for testing
type MockUpgrader struct {
	mock.Mock
}

func (m *MockUpgrader) CheckUpdate(ctx context.Context) (*UpgradeStatus, error) {
	args := m.Called(ctx)
	return args.Get(0).(*UpgradeStatus), args.Error(1)
}

func (m *MockUpgrader) Upgrade(ctx context.Context, options UpgradeOptions) error {
	args := m.Called(ctx, options)
	return args.Error(0)
}

func (m *MockUpgrader) Backup(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

func (m *MockUpgrader) Rollback(ctx context.Context, backupPath string) error {
	args := m.Called(ctx, backupPath)
	return args.Error(0)
}

func (m *MockUpgrader) GetUpdateMethod() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockUpgrader) ValidateUpgrade(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestNewUpgradeCoordinator(t *testing.T) {
	mockLogger := &MockLogger{}
	backupDir := "/tmp/test-backups"

	coordinator := NewUpgradeCoordinator(mockLogger, backupDir)

	assert.NotNil(t, coordinator)
	assert.Equal(t, backupDir, coordinator.backupDir)
	assert.NotEmpty(t, coordinator.upgraders)

	// Check that default upgraders are registered
	upgraders := coordinator.ListUpgraders()
	expectedUpgraders := []string{"homebrew", "brew", "asdf", "nvm", "rbenv", "pyenv", "sdkman"}

	for _, expected := range expectedUpgraders {
		assert.Contains(t, upgraders, expected)
	}
}

func TestUpgradeCoordinator_RegisterUpgrader(t *testing.T) {
	mockLogger := &MockLogger{}
	coordinator := NewUpgradeCoordinator(mockLogger, "/tmp")

	mockUpgrader := &MockUpgrader{}
	coordinator.RegisterUpgrader("test-manager", mockUpgrader)

	upgrader, exists := coordinator.GetUpgrader("test-manager")
	assert.True(t, exists)
	assert.Equal(t, mockUpgrader, upgrader)
}

func TestUpgradeCoordinator_CheckAll(t *testing.T) {
	mockLogger := &MockLogger{}
	coordinator := NewUpgradeCoordinator(mockLogger, "/tmp")

	// Clear default upgraders for controlled testing
	coordinator.upgraders = make(map[string]PackageManagerUpgrader)

	// Setup mock upgraders
	mockUpgrader1 := &MockUpgrader{}
	mockUpgrader2 := &MockUpgrader{}

	coordinator.RegisterUpgrader("manager1", mockUpgrader1)
	coordinator.RegisterUpgrader("manager2", mockUpgrader2)

	// Setup expectations
	status1 := &UpgradeStatus{
		Manager:         "manager1",
		CurrentVersion:  "1.0.0",
		LatestVersion:   "1.1.0",
		UpdateAvailable: true,
		UpdateMethod:    "test-update-1",
	}

	status2 := &UpgradeStatus{
		Manager:         "manager2",
		CurrentVersion:  "2.0.0",
		LatestVersion:   "2.0.0",
		UpdateAvailable: false,
		UpdateMethod:    "test-update-2",
	}

	mockUpgrader1.On("CheckUpdate", mock.Anything).Return(status1, nil)
	mockUpgrader2.On("CheckUpdate", mock.Anything).Return(status2, nil)

	mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
	mockLogger.On("Debug", mock.AnythingOfType("string"), mock.Anything).Return()

	// Execute test
	ctx := context.Background()
	report, err := coordinator.CheckAll(ctx)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, 2, report.TotalManagers)
	assert.Equal(t, 1, report.UpdatesNeeded) // Only manager1 has updates
	assert.Len(t, report.Managers, 2)

	mockUpgrader1.AssertExpectations(t)
	mockUpgrader2.AssertExpectations(t)
}

func TestUpgradeCoordinator_CheckManagers(t *testing.T) {
	mockLogger := &MockLogger{}
	coordinator := NewUpgradeCoordinator(mockLogger, "/tmp")

	// Clear default upgraders
	coordinator.upgraders = make(map[string]PackageManagerUpgrader)

	mockUpgrader := &MockUpgrader{}
	coordinator.RegisterUpgrader("test-manager", mockUpgrader)

	status := &UpgradeStatus{
		Manager:         "test-manager",
		CurrentVersion:  "1.0.0",
		LatestVersion:   "1.1.0",
		UpdateAvailable: true,
		UpdateMethod:    "test-update",
	}

	mockUpgrader.On("CheckUpdate", mock.Anything).Return(status, nil)
	mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
	mockLogger.On("Debug", mock.AnythingOfType("string"), mock.Anything).Return()

	ctx := context.Background()
	report, err := coordinator.CheckManagers(ctx, []string{"test-manager"})

	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, 1, report.TotalManagers)
	assert.Equal(t, 1, report.UpdatesNeeded)
	assert.Len(t, report.Managers, 1)
	assert.Equal(t, "test-manager", report.Managers[0].Manager)

	mockUpgrader.AssertExpectations(t)
}

func TestUpgradeCoordinator_CheckManagers_UnknownManager(t *testing.T) {
	mockLogger := &MockLogger{}
	coordinator := NewUpgradeCoordinator(mockLogger, "/tmp")

	mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
	mockLogger.On("Warn", mock.AnythingOfType("string"), mock.Anything).Return()

	ctx := context.Background()
	report, err := coordinator.CheckManagers(ctx, []string{"unknown-manager"})

	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, 1, report.TotalManagers)
	assert.Equal(t, 0, report.UpdatesNeeded)
	assert.Len(t, report.Managers, 1)
	assert.Equal(t, "unknown-manager", report.Managers[0].Manager)
	assert.Equal(t, "unknown", report.Managers[0].CurrentVersion)
	assert.False(t, report.Managers[0].UpdateAvailable)
}

func TestUpgradeCoordinator_UpgradeManagers_Success(t *testing.T) {
	mockLogger := &MockLogger{}
	coordinator := NewUpgradeCoordinator(mockLogger, "/tmp")

	// Clear default upgraders
	coordinator.upgraders = make(map[string]PackageManagerUpgrader)

	mockUpgrader := &MockUpgrader{}
	coordinator.RegisterUpgrader("test-manager", mockUpgrader)

	preStatus := &UpgradeStatus{
		Manager:         "test-manager",
		CurrentVersion:  "1.0.0",
		LatestVersion:   "1.1.0",
		UpdateAvailable: true,
		UpdateMethod:    "test-update",
	}

	postStatus := &UpgradeStatus{
		Manager:         "test-manager",
		CurrentVersion:  "1.1.0",
		LatestVersion:   "1.1.0",
		UpdateAvailable: false,
		UpdateMethod:    "test-update",
	}

	options := UpgradeOptions{
		Force:          false,
		BackupEnabled:  true,
		SkipValidation: false,
		Timeout:        5 * time.Minute,
	}

	mockUpgrader.On("CheckUpdate", mock.Anything).Return(preStatus, nil).Once()
	mockUpgrader.On("Upgrade", mock.Anything, options).Return(nil)
	mockUpgrader.On("CheckUpdate", mock.Anything).Return(postStatus, nil).Once()

	mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
	mockLogger.On("Debug", mock.AnythingOfType("string"), mock.Anything).Return()

	ctx := context.Background()
	report, err := coordinator.UpgradeManagers(ctx, []string{"test-manager"}, options)

	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, 1, report.TotalManagers)
	assert.Len(t, report.Managers, 1)

	mockUpgrader.AssertExpectations(t)
}

func TestUpgradeCoordinator_FormatReport(t *testing.T) {
	mockLogger := &MockLogger{}
	coordinator := NewUpgradeCoordinator(mockLogger, "/tmp")

	report := &UpgradeReport{
		Platform:      "linux",
		TotalManagers: 2,
		UpdatesNeeded: 1,
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
				LatestVersion:   "0.10.0",
				UpdateAvailable: false,
				UpdateMethod:    "git pull",
			},
		},
		Timestamp: time.Now(),
	}

	output := coordinator.FormatReport(report, false)

	assert.Contains(t, output, "Package Manager Upgrade Report")
	assert.Contains(t, output, "Platform: linux")
	assert.Contains(t, output, "Total Managers: 2")
	assert.Contains(t, output, "Updates Available: 1")
	assert.Contains(t, output, "homebrew: 3.0.0 → 3.1.0")
	assert.Contains(t, output, "asdf: 0.10.0")
}

func TestUpgradeCoordinator_FormatReport_Verbose(t *testing.T) {
	mockLogger := &MockLogger{}
	coordinator := NewUpgradeCoordinator(mockLogger, "/tmp")

	report := &UpgradeReport{
		Platform:      "linux",
		TotalManagers: 1,
		UpdatesNeeded: 1,
		Managers: []UpgradeStatus{
			{
				Manager:         "homebrew",
				CurrentVersion:  "3.0.0",
				LatestVersion:   "3.1.0",
				UpdateAvailable: true,
				UpdateMethod:    "brew update",
			},
		},
		Timestamp: time.Now(),
	}

	output := coordinator.FormatReport(report, true)

	assert.Contains(t, output, "homebrew: 3.0.0 → 3.1.0 (brew update)")
}

func TestUpgradeCoordinator_FormatReport_Nil(t *testing.T) {
	mockLogger := &MockLogger{}
	coordinator := NewUpgradeCoordinator(mockLogger, "/tmp")

	output := coordinator.FormatReport(nil, false)

	assert.Equal(t, "No upgrade report available\n", output)
}
