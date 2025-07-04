// Package bulkclone provides configuration loading and management for bulk repository cloning operations.
//
// This package handles:
//   - Loading and parsing bulk-clone configuration files (YAML format)
//   - Configuration file discovery across multiple standard locations
//   - Overlay configuration support for environment-specific settings
//   - Path expansion and environment variable resolution
//   - URL building for different Git service providers
//   - JSON Schema validation for configuration files
//
// The package supports a hierarchical configuration system where base configurations
// can be overridden by environment-specific overlay files. Configuration files are
// searched in the following order:
//  1. Current directory (./bulk-clone.yaml, ./bulk-clone.yml)
//  2. User home directory (~/.config/gzh-manager/bulk-clone.yaml)
//  3. System-wide directory (/etc/gzh-manager/bulk-clone.yaml)
//
// Main types:
//   - BulkCloneConfig: Primary configuration structure
//   - BulkCloneGithub: GitHub-specific configuration
//   - BulkCloneGitlab: GitLab-specific configuration
//
// Key functions:
//   - LoadConfig: Load configuration with automatic discovery
//   - FindConfigFile: Locate configuration files in standard paths
//   - ValidateConfig: Validate configuration against JSON schema
//   - BuildURL: Build repository URLs for different providers
package bulkclone
