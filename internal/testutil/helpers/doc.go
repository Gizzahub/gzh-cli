// Package helpers provides common test utilities and helper functions for the gzh-manager-go test suite.
// It includes assertion helpers, temporary directory management, environment variable utilities,
// and other testing conveniences to reduce boilerplate in test code.
//
// Key features:
//   - Enhanced assertions with detailed error messages
//   - Temporary directory creation and cleanup
//   - Environment variable capture and restoration
//   - Test data fixtures and generators
//   - Mock helper utilities
//   - Golden file testing support
//
// Example usage:
//
//	func TestMyFunction(t *testing.T) {
//	    tmpDir := helpers.TempDir(t)
//	    helpers.WithEnvVars(t, map[string]string{"GZH_TOKEN": "test"})
//	    
//	    result := MyFunction()
//	    helpers.AssertEqual(t, expected, result)
//	}
package helpers