// Package repoconfig defines types and utilities for repository configuration management.
// It provides data structures for representing repository settings, validation logic,
// and serialization/deserialization capabilities for various configuration formats.
//
// The package supports:
//   - Repository metadata (name, URL, description)
//   - Clone and sync strategies
//   - Authentication configuration
//   - Hook and workflow settings
//   - Platform-specific configurations
//   - Configuration inheritance and overrides
//
// Configuration can be loaded from:
//   - YAML/JSON files
//   - Environment variables
//   - API responses
//   - Command-line arguments
//
// All configuration types implement validation to ensure correctness
// before being used by other parts of the system.
package repoconfig
