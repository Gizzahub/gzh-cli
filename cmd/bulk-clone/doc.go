// Package bulk_clone provides the bulk cloning functionality for the GZH Manager CLI.
//
// This package implements the core bulk-clone command that allows users to clone
// entire organizations or groups of repositories from multiple Git platforms including
// GitHub, GitLab, Gitea, and Gogs.
//
// Key Features:
//   - Multi-platform repository cloning (GitHub, GitLab, Gitea, Gogs)
//   - Flexible configuration via YAML files
//   - Dry-run mode for testing configurations
//   - Resume functionality for interrupted operations
//   - Progress tracking and reporting
//   - Concurrent cloning with rate limiting
//   - Comprehensive error handling and recovery
//
// Configuration:
//
// The bulk-clone command is configured via YAML files that specify:
//   - Source platforms and authentication
//   - Target directory structure
//   - Repository filtering and selection
//   - Clone strategies and options
//
// Example usage:
//
//	gz bulk-clone --config bulk-clone.yaml
//	gz bulk-clone --config myconfig.yaml --dry-run
//	gz bulk-clone validate --config bulk-clone.yaml
//
// The package supports multiple clone strategies:
//   - reset: Hard reset and pull (default)
//   - pull: Merge remote changes
//   - fetch: Update remote tracking only
//
// See the samples/ directory for example configurations and the
// docs/bulk-clone-schema.json for the complete configuration schema.
package bulkclone
