// Package testutil provides testing utilities, fixtures, and builders for the gzh-manager-go project.
//
// This package is organized into two main sub-packages:
//
// builders/ - Provides fluent builder interfaces for creating test objects
// fixtures/ - Provides pre-built test fixtures for common scenarios
//
// The testutil package follows the builder pattern and fixture pattern to make
// tests more readable, maintainable, and reduce code duplication.
//
// Key Design Principles:
// - Fluent interfaces for readable test setup
// - Builder pattern for flexible test data construction
// - Fixtures for common test scenarios
// - Mock objects with call tracking for verification
// - Deterministic and reproducible test data
//
// Example Usage:
//
//	import (
//		"github.com/gizzahub/gzh-manager-go/internal/testutil/builders"
//		"github.com/gizzahub/gzh-manager-go/internal/testutil/fixtures"
//	)
//
//	func TestExample(t *testing.T) {
//		// Use builders for custom test data
//		config := builders.NewConfigBuilder().
//			WithGitHubProvider("token").
//			WithOrganization("github", "org", "path").
//			Build()
//
//		// Use fixtures for common scenarios
//		fixtures := fixtures.NewConfigFixtures()
//		standardConfig := fixtures.SimpleGitHubConfig()
//
//		// Use mock builders for testing
//		mockLogger := builders.NewMockLoggerBuilder().Build()
//		mockEnv := builders.NewEnvironmentBuilder().
//			WithGitHubToken("test-token").
//			Build()
//	}
//
// The package is designed to be internal to the project and should not be
// imported by external packages.
package testutil
