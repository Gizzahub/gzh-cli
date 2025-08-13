// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package bootstrap

import (
	"context"
	"fmt"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/Gizzahub/gzh-manager-go/internal/logger"
)

// BootstrapManager manages the installation and configuration of package managers.
type BootstrapManager struct {
	platform      string
	bootstrappers map[string]PackageManagerBootstrapper
	logger        logger.CommonLogger
	resolver      *DependencyResolver
}

// NewBootstrapManager creates a new bootstrap manager instance.
func NewBootstrapManager(logger logger.CommonLogger) *BootstrapManager {
	manager := &BootstrapManager{
		platform:      fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		bootstrappers: make(map[string]PackageManagerBootstrapper),
		logger:        logger,
		resolver:      NewDependencyResolver(),
	}

	// Register all available bootstrappers
	manager.registerBootstrappers()

	return manager
}

// registerBootstrappers registers all supported package manager bootstrappers.
func (bm *BootstrapManager) registerBootstrappers() {
	bootstrappers := []PackageManagerBootstrapper{
		NewHomebrewBootstrapper(bm.logger),
		NewAsdfBootstrapper(bm.logger),
		NewNvmBootstrapper(bm.logger),
		NewRbenvBootstrapper(bm.logger),
		NewPyenvBootstrapper(bm.logger),
		NewSdkmanBootstrapper(bm.logger),
	}

	for _, bootstrapper := range bootstrappers {
		if bootstrapper.IsSupported() {
			bm.bootstrappers[bootstrapper.GetName()] = bootstrapper

			// Register dependencies for resolver
			deps := bootstrapper.GetDependencies()
			if len(deps) > 0 {
				bm.resolver.AddDependency(bootstrapper.GetName(), deps)
			}
		}
	}
}

// CheckAll checks the installation status of all supported package managers.
func (bm *BootstrapManager) CheckAll(ctx context.Context) (*BootstrapReport, error) {
	startTime := time.Now()

	report := &BootstrapReport{
		Platform:  bm.platform,
		Timestamp: startTime,
		Managers:  make([]BootstrapStatus, 0),
	}

	// Check each registered manager
	for name, bootstrapper := range bm.bootstrappers {
		bm.logger.Debug("Checking manager", "name", name)

		status, err := bootstrapper.CheckInstallation(ctx)
		if err != nil {
			bm.logger.Warn("Failed to check manager", "name", name, "error", err)
			status = &BootstrapStatus{
				Manager:   name,
				Installed: false,
				Issues:    []string{fmt.Sprintf("Check failed: %v", err)},
			}
		}

		report.Managers = append(report.Managers, *status)
	}

	// Sort managers by name for consistent output
	sort.Slice(report.Managers, func(i, j int) bool {
		return report.Managers[i].Manager < report.Managers[j].Manager
	})

	// Calculate summary
	bm.calculateSummary(report)

	report.Duration = time.Since(startTime)

	return report, nil
}

// InstallManagers installs the specified package managers.
func (bm *BootstrapManager) InstallManagers(ctx context.Context, managerNames []string, opts BootstrapOptions) (*BootstrapReport, error) {
	startTime := time.Now()

	// If no specific managers specified, install all missing ones
	if len(managerNames) == 0 {
		checkReport, err := bm.CheckAll(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to check current status: %w", err)
		}

		for _, status := range checkReport.Managers {
			if !status.Installed {
				managerNames = append(managerNames, status.Manager)
			}
		}
	}

	// Resolve installation order based on dependencies
	installOrder, err := bm.resolver.ResolveDependencies(managerNames)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve dependencies: %w", err)
	}

	bm.logger.Info("Installation order determined", "order", installOrder)

	report := &BootstrapReport{
		Platform:  bm.platform,
		Timestamp: startTime,
		Managers:  make([]BootstrapStatus, 0),
	}

	// Install each manager in dependency order
	for _, managerName := range installOrder {
		bootstrapper, exists := bm.bootstrappers[managerName]
		if !exists {
			bm.logger.Warn("Unknown manager requested", "name", managerName)
			continue
		}

		bm.logger.Info("Installing manager", "name", managerName)

		status, err := bm.installSingleManager(ctx, bootstrapper, opts)
		if err != nil {
			bm.logger.Error("Failed to install manager", "name", managerName, "error", err)
			status.Issues = append(status.Issues, fmt.Sprintf("Installation failed: %v", err))
		}

		report.Managers = append(report.Managers, *status)
	}

	// Calculate final summary
	bm.calculateSummary(report)

	report.Duration = time.Since(startTime)

	return report, nil
}

// installSingleManager installs and configures a single package manager.
func (bm *BootstrapManager) installSingleManager(ctx context.Context, bootstrapper PackageManagerBootstrapper, opts BootstrapOptions) (*BootstrapStatus, error) {
	managerName := bootstrapper.GetName()

	// Check current status first
	status, err := bootstrapper.CheckInstallation(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check %s installation: %w", managerName, err)
	}

	// Skip if already installed and not forcing
	if status.Installed && !opts.Force {
		bm.logger.Info("Manager already installed, skipping", "name", managerName)
		return status, nil
	}

	if opts.DryRun {
		bm.logger.Info("DRY RUN: Would install manager", "name", managerName)
		status.Issues = append(status.Issues, "Dry run - not actually installed")
		return status, nil
	}

	// Install the manager
	bm.logger.Info("Installing manager", "name", managerName)
	if err := bootstrapper.Install(ctx, opts.Force); err != nil {
		return status, fmt.Errorf("installation failed: %w", err)
	}

	// Configure if not skipping configuration
	if !opts.SkipConfiguration {
		bm.logger.Info("Configuring manager", "name", managerName)
		if err := bootstrapper.Configure(ctx); err != nil {
			bm.logger.Warn("Configuration failed", "name", managerName, "error", err)
			status.Issues = append(status.Issues, fmt.Sprintf("Configuration failed: %v", err))
		}
	}

	// Validate installation
	if err := bootstrapper.Validate(ctx); err != nil {
		bm.logger.Warn("Validation failed", "name", managerName, "error", err)
		status.Issues = append(status.Issues, fmt.Sprintf("Validation failed: %v", err))
	}

	// Re-check status after installation
	newStatus, err := bootstrapper.CheckInstallation(ctx)
	if err != nil {
		bm.logger.Warn("Failed to check status after installation", "name", managerName, "error", err)
		status.Issues = append(status.Issues, fmt.Sprintf("Post-install check failed: %v", err))
		return status, nil
	}

	return newStatus, nil
}

// GetAvailableManagers returns list of all available package managers.
func (bm *BootstrapManager) GetAvailableManagers() []string {
	managers := make([]string, 0, len(bm.bootstrappers))
	for name := range bm.bootstrappers {
		managers = append(managers, name)
	}
	sort.Strings(managers)
	return managers
}

// GetInstallationOrder returns the recommended installation order for given managers.
func (bm *BootstrapManager) GetInstallationOrder(managers []string) ([]string, error) {
	return bm.resolver.ResolveDependencies(managers)
}

// calculateSummary calculates the summary statistics for a bootstrap report.
func (bm *BootstrapManager) calculateSummary(report *BootstrapReport) {
	summary := BootstrapSummary{
		Total: len(report.Managers),
	}

	for _, status := range report.Managers {
		if status.Installed {
			summary.Installed++
			if len(status.Issues) == 0 {
				summary.Configured++
			}
		} else {
			summary.Missing++
		}

		if len(status.Issues) > 0 {
			summary.Failed++
		}
	}

	report.Summary = summary
}

// FormatReport formats a bootstrap report for human-readable output.
func (bm *BootstrapManager) FormatReport(report *BootstrapReport, verbose bool) string {
	var builder strings.Builder

	builder.WriteString("📦 Package Manager Bootstrap Status\n\n")
	builder.WriteString(fmt.Sprintf("Platform: %s\n", report.Platform))
	builder.WriteString(fmt.Sprintf("Checked: %s\n\n", report.Timestamp.Format("2006-01-02 15:04:05")))

	builder.WriteString("Manager Status:\n")
	for _, status := range report.Managers {
		icon := "❌"
		if status.Installed {
			icon = "✅"
		}

		line := fmt.Sprintf("  %s %-12s", icon, status.Manager)
		if status.Version != "" {
			line += fmt.Sprintf(" %-12s", status.Version)
		} else {
			line += " missing     "
		}

		if status.ConfigPath != "" {
			line += fmt.Sprintf(" %s", status.ConfigPath)
		} else if len(status.Dependencies) > 0 {
			line += fmt.Sprintf(" Will install via %s", strings.Join(status.Dependencies, ", "))
		}

		builder.WriteString(line + "\n")

		if verbose && len(status.Issues) > 0 {
			for _, issue := range status.Issues {
				builder.WriteString(fmt.Sprintf("    ⚠️  %s\n", issue))
			}
		}
	}

	builder.WriteString(fmt.Sprintf("\nSummary: %d/%d installed, %d missing",
		report.Summary.Installed, report.Summary.Total, report.Summary.Missing))

	if report.Summary.Failed > 0 {
		builder.WriteString(fmt.Sprintf(", %d with issues", report.Summary.Failed))
	}

	builder.WriteString("\n")

	// Show recommended installation order for missing managers
	missing := make([]string, 0)
	for _, status := range report.Managers {
		if !status.Installed {
			missing = append(missing, status.Manager)
		}
	}

	if len(missing) > 0 {
		if order, err := bm.resolver.ResolveDependencies(missing); err == nil {
			builder.WriteString("\nRecommended installation order:\n")
			for i, manager := range order {
				deps := bm.bootstrappers[manager].GetDependencies()
				depInfo := ""
				if len(deps) > 0 {
					depInfo = fmt.Sprintf(" (depends on: %s)", strings.Join(deps, ", "))
				}
				builder.WriteString(fmt.Sprintf("  %d. %s%s\n", i+1, manager, depInfo))
			}
		}
	}

	return builder.String()
}
