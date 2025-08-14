// Package github provides a comprehensive client library for interacting with GitHub's API.
// It implements repository management, organization operations, Actions policy enforcement,
// and various automation features required by the gzh-cli tool.
//
// Key features:
//   - Repository cloning and synchronization
//   - Organization and team management
//   - GitHub Actions policy validation and enforcement
//   - Webhook handling and event processing
//   - Pull request and issue automation
//   - Release management
//   - Dependency version policy enforcement
//
// The package uses GitHub's REST and GraphQL APIs, providing:
//   - Automatic retry with exponential backoff
//   - Rate limit handling
//   - Concurrent operations with worker pools
//   - Comprehensive error handling
//   - Metrics and logging integration
//
// Authentication is handled via personal access tokens or GitHub Apps,
// with support for fine-grained permissions and OAuth scopes.
package github
