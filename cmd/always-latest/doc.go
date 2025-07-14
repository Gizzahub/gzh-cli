// Package always_latest provides automated package manager update functionality
// for keeping development tools and dependencies up to date.
//
// This package implements the always-latest command that automatically updates
// packages across multiple package managers to ensure developers always have
// the latest versions of their tools.
//
// Supported Package Managers:
//   - asdf - Version manager for multiple languages
//   - Homebrew - macOS package manager
//   - MacPorts - macOS package manager alternative
//   - SDKMAN - Java ecosystem version manager
//   - rbenv - Ruby version manager
//   - APT - Debian/Ubuntu package manager
//
// Key Features:
//   - Multi-platform support (macOS, Linux, Windows)
//   - Selective package updates with filtering
//   - Dry-run mode for preview
//   - Rollback functionality
//   - Update scheduling and automation
//   - Cross-platform compatibility checks
//
// Update Strategies:
//   - Conservative: Only patch updates
//   - Moderate: Minor version updates
//   - Aggressive: All available updates
//   - Custom: User-defined update rules
//
// Example usage:
//
//	gz always-latest --all
//	gz always-latest --package-manager homebrew
//	gz always-latest --dry-run --strategy conservative
//	gz always-latest rollback --last-update
//
// The package automatically detects available package managers on the system
// and provides intelligent update recommendations based on the user's
// development environment and preferences.
package alwayslatest
