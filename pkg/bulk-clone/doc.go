// Package bulkclone implements the core functionality for cloning multiple repositories
// from various Git hosting platforms. It provides a unified interface for bulk operations
// across GitHub, GitLab, Gitea, and Gogs platforms.
//
// The package features:
//   - Multi-platform support with consistent API
//   - Concurrent cloning with configurable parallelism
//   - Progress tracking and resumable operations
//   - Flexible filtering and selection criteria
//   - Multiple clone strategies (HTTPS, SSH, mirror)
//   - Automatic retry with exponential backoff
//   - Detailed logging and error reporting
//
// Configuration is handled through YAML files with schema validation,
// supporting organization-wide cloning, team-based filtering, and
// custom repository selection rules.
//
// Example usage:
//
//	config := bulkclone.LoadConfig("config.yaml")
//	manager := bulkclone.NewManager(config)
//	results := manager.CloneAll(context.Background())
package bulkclone
