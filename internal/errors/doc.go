// Package errors provides enhanced error handling utilities for the gzh-manager-go project.
// It extends the standard library's error handling with additional context, error wrapping,
// and structured error types for better debugging and error reporting.
//
// Features:
//   - Structured error types with consistent formatting
//   - Error wrapping with context preservation
//   - Stack trace capture for debugging
//   - Error categorization (user, system, network, etc.)
//   - HTTP status code mapping for API errors
//   - Multi-error aggregation for batch operations
//
// Usage:
//
//	err := errors.New("operation failed").
//	    WithContext("repository", repoName).
//	    WithCause(originalErr)
package errors
