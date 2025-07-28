// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

// Package git provides repository lifecycle management commands.
//
// This package implements the `gz git repo` command structure, providing
// comprehensive repository management capabilities across multiple Git platforms
// including GitHub, GitLab, and Gitea.
//
// The package includes subcommands for:
//   - clone: Advanced repository cloning with parallel execution
//   - list: Repository listing with filtering and formatting
//   - create: Repository creation with templates and settings
//   - delete: Safe repository deletion with backups
//   - archive: Repository archival management
//   - sync: Cross-platform repository synchronization
//   - migrate: Repository migration between platforms
//   - search: Advanced repository search capabilities
//
// Usage:
//
//	gz git repo clone --provider github --org myorg --target ./repos
//	gz git repo list --provider gitlab --org mygroup --format json
//	gz git repo create --provider github --org myorg --name newrepo --private
//
// The package integrates with existing synclone functionality for enhanced
// cloning capabilities and provides a unified interface for repository
// lifecycle management across different Git hosting platforms.
package git
