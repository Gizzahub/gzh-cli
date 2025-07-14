// Package utils provides common utility functions and helpers
// for the GZH Manager system.
//
// This package implements reusable utility functions including
// string manipulation, file operations, data conversion, and
// general-purpose helpers used throughout the system.
//
// Key Components:
//
// String Utils:
//   - String formatting and manipulation
//   - Template processing and rendering
//   - Text transformation utilities
//   - String validation and sanitization
//
// File Utils:
//   - File and directory operations
//   - Path manipulation and resolution
//   - File content processing
//   - Permission and attribute management
//
// Data Utils:
//   - Data structure conversion
//   - JSON and YAML processing
//   - Configuration parsing
//   - Data validation and sanitization
//
// System Utils:
//   - Environment variable handling
//   - Process management
//   - System information gathering
//   - Cross-platform compatibility
//
// Features:
//   - Cross-platform compatibility
//   - Error handling and recovery
//   - Performance-optimized operations
//   - Thread-safe implementations
//   - Comprehensive validation
//
// Example usage:
//
//	// String operations
//	formatted := utils.FormatTemplate(template, data)
//	sanitized := utils.SanitizeString(input)
//
//	// File operations
//	exists := utils.FileExists(path)
//	content, err := utils.ReadFileContents(path)
//
//	// Data conversion
//	jsonData := utils.ToJSON(struct)
//	config := utils.ParseYAML(yamlData)
//
// The package provides essential utility functions that simplify
// common operations and promote code reuse across the system.
package utils
