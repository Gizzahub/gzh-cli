// Package testlib provides testing utilities and infrastructure
// for the GZH Manager system.
//
// This package implements comprehensive testing support including
// test fixtures, mock services, assertion helpers, and testing
// infrastructure for unit, integration, and end-to-end tests.
//
// Key Components:
//
// Test Fixtures:
//   - Sample data generation
//   - Test environment setup
//   - Reproducible test scenarios
//   - Data seeding and cleanup
//
// Mock Services:
//   - HTTP service mocking
//   - External API simulation
//   - File system mocking
//   - Database mocking
//
// Assertion Helpers:
//   - Custom assertion functions
//   - Error validation helpers
//   - Response validation utilities
//   - Test result comparison
//
// Test Infrastructure:
//   - Test harness and runners
//   - Environment isolation
//   - Parallel test execution
//   - Test result reporting
//
// Features:
//   - Cross-platform test support
//   - Performance benchmarking
//   - Test coverage analysis
//   - Flaky test detection
//   - Test result analytics
//
// Example usage:
//
//	testEnv := testlib.NewTestEnvironment()
//	defer testEnv.Cleanup()
//
//	mockAPI := testlib.NewMockAPI()
//	fixture := testlib.LoadFixture("sample-data.json")
//
//	testlib.AssertNoError(t, err)
//	testlib.AssertEqual(t, expected, actual)
//
// The package provides robust testing infrastructure that ensures
// code quality and reliability across the entire system.
package testlib
