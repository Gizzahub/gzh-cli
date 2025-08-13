// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package bootstrap

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Gizzahub/gzh-manager-go/internal/logger"
)

// mockBootstrapper is a mock implementation of PackageManagerBootstrapper for testing
type mockBootstrapper struct {
	name         string
	isSupported  bool
	isInstalled  bool
	dependencies []string
	installError error
	validateError error
}

func (m *mockBootstrapper) GetName() string { return m.name }
func (m *mockBootstrapper) IsSupported() bool { return m.isSupported }
func (m *mockBootstrapper) GetDependencies() []string { return m.dependencies }

func (m *mockBootstrapper) CheckInstallation(ctx context.Context) (*BootstrapStatus, error) {
	return &BootstrapStatus{
		Manager:   m.name,
		Installed: m.isInstalled,
		Version:   "1.0.0",
	}, nil
}

func (m *mockBootstrapper) Install(ctx context.Context, force bool) error {
	if m.installError != nil {
		return m.installError
	}
	m.isInstalled = true
	return nil
}

func (m *mockBootstrapper) Configure(ctx context.Context) error {
	return nil
}

func (m *mockBootstrapper) GetInstallScript() (string, error) {
	return "echo 'mock install'", nil
}

func (m *mockBootstrapper) Validate(ctx context.Context) error {
	return m.validateError
}

func TestBootstrapManager_GetAvailableManagers(t *testing.T) {
	logger := logger.NewSimpleLogger("test")
	manager := &BootstrapManager{
		bootstrappers: make(map[string]PackageManagerBootstrapper),
		logger:        logger,
		resolver:      NewDependencyResolver(),
	}

	// Add mock bootstrappers
	manager.bootstrappers["brew"] = &mockBootstrapper{name: "brew", isSupported: true}
	manager.bootstrappers["asdf"] = &mockBootstrapper{name: "asdf", isSupported: true}
	manager.bootstrappers["nvm"] = &mockBootstrapper{name: "nvm", isSupported: true}

	managers := manager.GetAvailableManagers()
	expected := []string{"asdf", "brew", "nvm"}
	assert.ElementsMatch(t, expected, managers)
}

func TestBootstrapManager_CheckAll(t *testing.T) {
	logger := logger.NewSimpleLogger("test")
	manager := &BootstrapManager{
		bootstrappers: make(map[string]PackageManagerBootstrapper),
		logger:        logger,
		resolver:      NewDependencyResolver(),
	}

	// Add mock bootstrappers
	manager.bootstrappers["brew"] = &mockBootstrapper{
		name: "brew", isSupported: true, isInstalled: true,
	}
	manager.bootstrappers["nvm"] = &mockBootstrapper{
		name: "nvm", isSupported: true, isInstalled: false,
	}

	ctx := context.Background()
	report, err := manager.CheckAll(ctx)
	
	require.NoError(t, err)
	assert.Equal(t, 2, report.Summary.Total)
	assert.Equal(t, 1, report.Summary.Installed)
	assert.Equal(t, 1, report.Summary.Missing)
	assert.Len(t, report.Managers, 2)
}

func TestBootstrapManager_InstallManagers(t *testing.T) {
	logger := logger.NewSimpleLogger("test")
	manager := &BootstrapManager{
		bootstrappers: make(map[string]PackageManagerBootstrapper),
		logger:        logger,
		resolver:      NewDependencyResolver(),
	}

	// Add mock bootstrappers with dependencies
	brewMock := &mockBootstrapper{
		name: "brew", isSupported: true, isInstalled: false, dependencies: []string{},
	}
	asdfMock := &mockBootstrapper{
		name: "asdf", isSupported: true, isInstalled: false, dependencies: []string{"brew"},
	}

	manager.bootstrappers["brew"] = brewMock
	manager.bootstrappers["asdf"] = asdfMock
	manager.resolver.AddDependency("asdf", []string{"brew"})

	ctx := context.Background()
	opts := BootstrapOptions{
		Force:   false,
		DryRun:  false,
		Timeout: Duration{Duration: 5 * time.Minute},
	}

	managerNames := []string{"asdf"} // This should also install brew due to dependency
	report, err := manager.InstallManagers(ctx, managerNames, opts)
	
	require.NoError(t, err)
	// Only asdf should be in the report (as it was requested)
	assert.True(t, asdfMock.isInstalled) // Should be installed as requested
	assert.True(t, report.Summary.Total >= 1)
	
	// Find asdf in the report
	var asdfStatus *BootstrapStatus
	for _, status := range report.Managers {
		if status.Manager == "asdf" {
			asdfStatus = &status
			break
		}
	}
	require.NotNil(t, asdfStatus, "asdf should be in report")
	assert.True(t, asdfStatus.Installed)
}

func TestBootstrapManager_GetInstallationOrder(t *testing.T) {
	logger := logger.NewSimpleLogger("test")
	manager := &BootstrapManager{
		bootstrappers: make(map[string]PackageManagerBootstrapper),
		logger:        logger,
		resolver:      NewDependencyResolver(),
	}

	// Setup dependencies
	manager.resolver.AddDependency("asdf", []string{"brew"})
	manager.resolver.AddDependency("rbenv", []string{"brew"})

	order, err := manager.GetInstallationOrder([]string{"asdf", "rbenv", "nvm"})
	require.NoError(t, err)

	// brew should come before asdf and rbenv
	brewIndex := -1
	asdfIndex := -1
	rbenvIndex := -1

	for i, name := range order {
		switch name {
		case "brew":
			brewIndex = i
		case "asdf":
			asdfIndex = i
		case "rbenv":
			rbenvIndex = i
		}
	}

	// Dependencies should be resolved correctly
	if brewIndex != -1 && asdfIndex != -1 {
		assert.Less(t, brewIndex, asdfIndex, "brew should come before asdf")
	}
	if brewIndex != -1 && rbenvIndex != -1 {
		assert.Less(t, brewIndex, rbenvIndex, "brew should come before rbenv")
	}
	
	// nvm has no dependencies, so it should be included
	assert.Contains(t, order, "nvm")
}

func TestBootstrapManager_FormatReport(t *testing.T) {
	logger := logger.NewSimpleLogger("test")
	manager := &BootstrapManager{
		bootstrappers: make(map[string]PackageManagerBootstrapper),
		logger:        logger,
		resolver:      NewDependencyResolver(),
	}
	
	// Add some mock bootstrappers to prevent nil pointer
	manager.bootstrappers["brew"] = &mockBootstrapper{name: "brew", isSupported: true}
	manager.bootstrappers["asdf"] = &mockBootstrapper{name: "asdf", isSupported: true}

	report := &BootstrapReport{
		Platform:  "darwin/arm64",
		Timestamp: time.Now(),
		Managers: []BootstrapStatus{
			{
				Manager:   "brew",
				Installed: true,
				Version:   "4.1.0",
				ConfigPath: "/opt/homebrew/bin/brew",
			},
			{
				Manager:   "asdf",
				Installed: false,
				Issues:    []string{"not found in PATH"},
			},
		},
		Summary: BootstrapSummary{
			Total:     2,
			Installed: 1,
			Missing:   1,
		},
	}

	output := manager.FormatReport(report, false)
	
	assert.Contains(t, output, "Package Manager Bootstrap Status")
	assert.Contains(t, output, "darwin/arm64")
	assert.Contains(t, output, "✅ brew")
	assert.Contains(t, output, "❌ asdf")
	assert.Contains(t, output, "1/2 installed, 1 missing")
}

func TestBootstrapOptions_JSONMarshal(t *testing.T) {
	// Test Duration JSON marshaling
	d := Duration{Duration: 5 * time.Minute}
	
	data, err := d.MarshalJSON()
	require.NoError(t, err)
	assert.Equal(t, `"5m0s"`, string(data))
	
	// Test unmarshaling
	var d2 Duration
	err = d2.UnmarshalJSON([]byte(`"10m30s"`))
	require.NoError(t, err)
	assert.Equal(t, 10*time.Minute+30*time.Second, d2.Duration)
}