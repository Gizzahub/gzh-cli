// Package fixtures provides pre-built test fixtures for common testing scenarios.
//
// This package complements the builders package by providing ready-to-use fixtures
// for common test scenarios. While builders provide flexibility for custom test data,
// fixtures provide convenience for standard test cases.
//
// The fixtures are organized by domain:
//
// Configuration Fixtures:
// - ConfigFixtures: Common configuration objects
// - ConfigYAMLFixtures: YAML configuration strings
//
// GitHub Fixtures:
// - GitHubFixtures: GitHub API objects and test data
//
// Usage Example:
//
//	func TestSomething(t *testing.T) {
//		fixtures := fixtures.NewConfigFixtures()
//
//		// Use a pre-built configuration
//		config := fixtures.SimpleGitHubConfig()
//
//		// Use a pre-built YAML configuration
//		yamlFixtures := fixtures.NewConfigYAMLFixtures()
//		yaml := yamlFixtures.SimpleGitHubYAML()
//
//		// Test with the fixtures
//		// ...
//	}
//
// Fixtures are designed to be deterministic and self-contained, making tests
// reliable and reproducible. They cover common scenarios like:
// - Valid configurations
// - Invalid configurations for error testing
// - Complex multi-provider setups
// - Large datasets for performance testing
//
// For custom test data that doesn't fit the standard fixtures, use the
// builders package instead.
package fixtures
