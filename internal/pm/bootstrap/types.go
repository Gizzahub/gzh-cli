// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

// Package bootstrap provides automatic installation and configuration of package managers.
package bootstrap

import (
	"context"
	"encoding/json"
	"time"
)

// BootstrapStatus represents the installation status of a package manager.
type BootstrapStatus struct {
	Manager      string            `json:"manager"`
	Installed    bool              `json:"installed"`
	Version      string            `json:"version,omitempty"`
	ConfigPath   string            `json:"configPath,omitempty"`
	Issues       []string          `json:"issues,omitempty"`
	Dependencies []string          `json:"dependencies,omitempty"`
	Details      map[string]string `json:"details,omitempty"`
}

// BootstrapSummary provides overview statistics for bootstrap operations.
type BootstrapSummary struct {
	Total      int `json:"total"`
	Installed  int `json:"installed"`
	Missing    int `json:"missing"`
	Configured int `json:"configured"`
	Failed     int `json:"failed"`
}

// BootstrapReport contains the complete bootstrap analysis and results.
type BootstrapReport struct {
	Platform  string            `json:"platform"`
	Summary   BootstrapSummary  `json:"summary"`
	Managers  []BootstrapStatus `json:"managers"`
	Timestamp time.Time         `json:"timestamp"`
	Duration  time.Duration     `json:"duration,omitempty"`
}

// PackageManagerBootstrapper defines the interface for package manager installation and configuration.
type PackageManagerBootstrapper interface {
	// CheckInstallation verifies if the package manager is installed and configured
	CheckInstallation(ctx context.Context) (*BootstrapStatus, error)

	// Install installs the package manager
	Install(ctx context.Context, force bool) error

	// Configure sets up the package manager with appropriate configuration
	Configure(ctx context.Context) error

	// GetDependencies returns list of other managers this one depends on
	GetDependencies() []string

	// GetInstallScript returns the installation script or command
	GetInstallScript() (string, error)

	// Validate ensures the installation is working correctly
	Validate(ctx context.Context) error

	// GetName returns the name of this package manager
	GetName() string

	// IsSupported checks if this manager is supported on current platform
	IsSupported() bool
}

// BootstrapOptions configures bootstrap behavior.
type BootstrapOptions struct {
	Managers          []string `json:"managers,omitempty"` // Specific managers to process (empty = all)
	Force             bool     `json:"force"`              // Force reinstall even if already installed
	SkipConfiguration bool     `json:"skipConfiguration"`  // Skip post-install configuration
	DryRun            bool     `json:"dryRun"`             // Only simulate, don't actually install
	Timeout           Duration `json:"timeout"`            // Timeout for installation operations
	Verbose           bool     `json:"verbose"`            // Enable verbose output
}

// Duration is a wrapper for time.Duration to support JSON marshaling.
type Duration struct {
	time.Duration
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.String() + `"`), nil
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	var err error
	d.Duration, err = time.ParseDuration(s)
	return err
}
