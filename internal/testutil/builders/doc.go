// Package builders provides fluent builder interfaces for creating test fixtures
// and mock objects in a consistent and readable way.
//
// This package implements the builder pattern to reduce repetitive test data creation
// and make tests more maintainable. It includes builders for:
//
// Configuration Objects:
// - ConfigBuilder: For building UnifiedConfig test objects
// - EnvironmentBuilder: For building mock environments with variables
//
// Mock Objects:
// - MockLoggerBuilder: For building mock loggers with call tracking
// - MockHTTPClientBuilder: For building mock HTTP clients with response configuration
// - MockGitHubProviderFactoryBuilder: For building mock GitHub provider factories
//
// GitHub API Objects:
// - BulkCloneRequestBuilder: For building bulk clone requests
// - BulkCloneResultBuilder: For building bulk clone results
// - RepositoryInfoBuilder: For building repository information
// - RepositoryFiltersBuilder: For building repository filters
//
// Usage Example:
//
//	// Create a test configuration
//	config := builders.NewConfigBuilder().
//		WithVersion("1.0.0").
//		WithDefaultProvider("github").
//		WithGitHubProvider("${GITHUB_TOKEN}").
//		WithOrganization("github", "test-org", "~/repos/test").
//		Build()
//
//	// Create a mock environment
//	env := builders.NewEnvironmentBuilder().
//		WithGitHubToken("test-token").
//		WithHome("/home/user").
//		Build()
//
//	// Create a mock logger
//	logger := builders.NewMockLoggerBuilder().Build()
//
// All builders follow the fluent interface pattern and are designed to be
// chained together for readable test setup.
package builders
