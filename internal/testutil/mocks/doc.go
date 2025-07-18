// Package mocks provides pre-configured mock implementations for testing the gzh-manager-go project.
// It includes commonly used mocks for HTTP clients, file systems, Git operations, and external services,
// making it easier to write isolated unit tests.
//
// Available mocks:
//   - HTTP client with configurable responses
//   - File system operations with in-memory implementation
//   - Git command execution with scripted outputs
//   - GitHub/GitLab/Gitea API clients
//   - Configuration loaders
//   - Logger implementations
//
// The mocks are designed to be:
//   - Easy to configure with expected behaviors
//   - Thread-safe for parallel test execution
//   - Provide useful debugging output
//   - Support both success and error scenarios
//
// Example:
//
//	mockHTTP := mocks.NewHTTPClient()
//	mockHTTP.ExpectGet("/api/repos").Return(200, `{"name": "test"}`)
package mocks