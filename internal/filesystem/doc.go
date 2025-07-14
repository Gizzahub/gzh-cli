// Package filesystem provides file system abstraction interfaces and
// implementations for the GZH Manager system.
//
// This package defines the FileSystem interface that abstracts file system
// operations, enabling easy testing through mocking and providing consistent
// file operations across different platforms and environments.
//
// Key Components:
//
// FileSystem Interface:
//   - File and directory operations (Create, Read, Write, Delete)
//   - Path manipulation and validation
//   - Permission and ownership management
//   - Atomic operations and transactions
//
// Implementations:
//   - OSFileSystem: Standard operating system file operations
//   - MemoryFileSystem: In-memory file system for testing
//   - MockFileSystem: Generated mock for unit testing
//
// Features:
//   - Cross-platform path handling
//   - Safe file operations with error handling
//   - Directory traversal and pattern matching
//   - Temporary file and directory management
//   - File system event monitoring
//
// Testing Support:
//   - Complete mock implementation using gomock
//   - Test utilities for file system scenarios
//   - Temporary directory helpers
//   - File content validation utilities
//
// Example usage:
//
//	fs := filesystem.NewOSFileSystem()
//	err := fs.WriteFile("config.yaml", content, 0644)
//	content, err := fs.ReadFile("config.yaml")
//
// The abstraction enables consistent file operations throughout the
// application while supporting comprehensive testing through mocking.
package filesystem
