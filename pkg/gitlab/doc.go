// Package gitlab provides a client library for interacting with GitLab's API.
// It implements group management, project operations, and repository synchronization
// features required by the gzh-manager-go tool for GitLab instances.
//
// Features:
//   - Project and group cloning
//   - Repository synchronization
//   - Pipeline management
//   - Merge request automation
//   - Issue tracking integration
//   - Container registry operations
//   - CI/CD configuration validation
//
// The package supports:
//   - GitLab.com and self-hosted instances
//   - Personal access tokens and OAuth2
//   - API v4 with pagination support
//   - Concurrent operations with rate limiting
//   - Webhook event processing
//
// All operations are designed to be idempotent and safe for automation,
// with comprehensive error handling and retry logic.
package gitlab
