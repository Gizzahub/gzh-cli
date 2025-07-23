// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

// Package errors provides standardized error handling patterns and utilities
// for consistent error management across the application.
package errors

import (
	"errors"
	"fmt"
	"runtime"
	"time"
)

// ErrorCode represents standardized error codes for consistent error classification.
type ErrorCode string

const (
	// Configuration errors..
	ErrorCodeInvalidConfig  ErrorCode = "INVALID_CONFIG"
	ErrorCodeMissingConfig  ErrorCode = "MISSING_CONFIG"
	ErrorCodeConfigNotFound ErrorCode = "CONFIG_NOT_FOUND"

	// Authentication errors..
	ErrorCodeInvalidToken      ErrorCode = "INVALID_TOKEN"
	ErrorCodeTokenExpired      ErrorCode = "TOKEN_EXPIRED"
	ErrorCodeInsufficientPerms ErrorCode = "INSUFFICIENT_PERMISSIONS"
	ErrorCodeAuthFailed        ErrorCode = "AUTHENTICATION_FAILED"

	// Network errors..
	ErrorCodeNetworkTimeout    ErrorCode = "NETWORK_TIMEOUT"
	ErrorCodeConnectionFailed  ErrorCode = "CONNECTION_FAILED"
	ErrorCodeRateLimitExceeded ErrorCode = "RATE_LIMIT_EXCEEDED"
	ErrorCodeAPIUnavailable    ErrorCode = "API_UNAVAILABLE"

	// Repository errors..
	ErrorCodeRepoNotFound       ErrorCode = "REPOSITORY_NOT_FOUND"
	ErrorCodeCloneFailed        ErrorCode = "CLONE_FAILED"
	ErrorCodeGitOperationFailed ErrorCode = "GIT_OPERATION_FAILED"
	ErrorCodePermissionDenied   ErrorCode = "PERMISSION_DENIED"

	// Validation errors..
	ErrorCodeInvalidInput     ErrorCode = "INVALID_INPUT"
	ErrorCodeValidationFailed ErrorCode = "VALIDATION_FAILED"
	ErrorCodeInvalidFormat    ErrorCode = "INVALID_FORMAT"

	// System errors.
	ErrorCodeInternalError     ErrorCode = "INTERNAL_ERROR"
	ErrorCodeResourceExhausted ErrorCode = "RESOURCE_EXHAUSTED"
	ErrorCodeOperationFailed   ErrorCode = "OPERATION_FAILED"
	ErrorCodeTimeout           ErrorCode = "TIMEOUT"

	// File system errors.
	ErrorCodeFileNotFound ErrorCode = "FILE_NOT_FOUND"
	ErrorCodeAccessDenied ErrorCode = "ACCESS_DENIED"
	ErrorCodeDiskFull     ErrorCode = "DISK_FULL"
	ErrorCodeIOError      ErrorCode = "IO_ERROR"
)

// ErrorSeverity represents the severity level of an error.
type ErrorSeverity string

const (
	SeverityLow      ErrorSeverity = "low"
	SeverityMedium   ErrorSeverity = "medium"
	SeverityHigh     ErrorSeverity = "high"
	SeverityCritical ErrorSeverity = "critical"
)

// StandardError represents a standardized error with context and metadata.
type StandardError struct {
	Code        ErrorCode              `json:"code"`
	Message     string                 `json:"message"`
	Severity    ErrorSeverity          `json:"severity"`
	Timestamp   time.Time              `json:"timestamp"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Cause       error                  `json:"cause,omitempty"`
	StackTrace  string                 `json:"stackTrace,omitempty"`
	Suggestions []string               `json:"suggestions,omitempty"`
	Retryable   bool                   `json:"retryable"`
}

// Error implements the error interface.
func (e *StandardError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause for error unwrapping.
func (e *StandardError) Unwrap() error {
	return e.Cause
}

// Is implements error comparison for errors.Is().
func (e *StandardError) Is(target error) bool {
	var stdErr *StandardError
	if errors.As(target, &stdErr) {
		return e.Code == stdErr.Code
	}
	return false
}

// NewStandardError creates a new standardized error.
func NewStandardError(code ErrorCode, message string, severity ErrorSeverity) *StandardError {
	return &StandardError{
		Code:        code,
		Message:     message,
		Severity:    severity,
		Timestamp:   time.Now(),
		Context:     make(map[string]interface{}),
		Suggestions: make([]string, 0),
		Retryable:   determineRetryability(code),
		StackTrace:  captureStackTrace(),
	}
}

// WrapError wraps an existing error with standardized error context.
func WrapError(cause error, code ErrorCode, message string, severity ErrorSeverity) *StandardError {
	return &StandardError{
		Code:        code,
		Message:     message,
		Severity:    severity,
		Timestamp:   time.Now(),
		Context:     make(map[string]interface{}),
		Cause:       cause,
		Suggestions: make([]string, 0),
		Retryable:   determineRetryability(code),
		StackTrace:  captureStackTrace(),
	}
}

// WithContext adds context information to the error.
func (e *StandardError) WithContext(key string, value interface{}) *StandardError {
	e.Context[key] = value
	return e
}

// WithSuggestion adds a suggestion for resolving the error.
func (e *StandardError) WithSuggestion(suggestion string) *StandardError {
	e.Suggestions = append(e.Suggestions, suggestion)
	return e
}

// WithRetryable sets whether the error should be retried.
func (e *StandardError) WithRetryable(retryable bool) *StandardError {
	e.Retryable = retryable
	return e
}

// determineRetryability determines if an error should be retryable based on its code.
func determineRetryability(code ErrorCode) bool {
	retryableErrors := map[ErrorCode]bool{
		ErrorCodeNetworkTimeout:    true,
		ErrorCodeConnectionFailed:  true,
		ErrorCodeRateLimitExceeded: true,
		ErrorCodeAPIUnavailable:    true,
		ErrorCodeTimeout:           true,
		ErrorCodeResourceExhausted: true,
		ErrorCodeIOError:           true,
	}
	return retryableErrors[code]
}

// captureStackTrace captures the current stack trace.
func captureStackTrace() string {
	// Capture stack trace (simplified implementation)
	buf := make([]byte, 2048)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

// Common error constructors for frequently used errors

// NewConfigError creates a configuration-related error.
func NewConfigError(message string, cause error) *StandardError {
	return WrapError(cause, ErrorCodeInvalidConfig, message, SeverityHigh).
		WithSuggestion("Check configuration file syntax and values").
		WithSuggestion("Refer to documentation for valid configuration options")
}

// NewAuthError creates an authentication-related error.
func NewAuthError(message string, cause error) *StandardError {
	return WrapError(cause, ErrorCodeAuthFailed, message, SeverityHigh).
		WithSuggestion("Verify authentication credentials").
		WithSuggestion("Check token permissions and validity")
}

// NewNetworkError creates a network-related error.
func NewNetworkError(message string, cause error) *StandardError {
	return WrapError(cause, ErrorCodeConnectionFailed, message, SeverityMedium).
		WithSuggestion("Check network connectivity").
		WithSuggestion("Verify firewall and proxy settings").
		WithRetryable(true)
}

// NewValidationError creates a validation-related error.
func NewValidationError(message string, field string) *StandardError {
	return NewStandardError(ErrorCodeValidationFailed, message, SeverityMedium).
		WithContext("field", field).
		WithSuggestion("Check input format and constraints")
}

// NewRepositoryError creates a repository-related error.
func NewRepositoryError(message string, repository string, cause error) *StandardError {
	return WrapError(cause, ErrorCodeRepoNotFound, message, SeverityMedium).
		WithContext("repository", repository).
		WithSuggestion("Verify repository exists and is accessible").
		WithSuggestion("Check authentication permissions for repository")
}

// NewFileSystemError creates a file system-related error.
func NewFileSystemError(message string, path string, cause error) *StandardError {
	return WrapError(cause, ErrorCodeIOError, message, SeverityMedium).
		WithContext("path", path).
		WithSuggestion("Check file permissions and disk space").
		WithSuggestion("Verify path exists and is accessible")
}

// ErrorCollector collects multiple errors for batch processing.
type ErrorCollector struct {
	errors   []*StandardError
	warnings []*StandardError
}

// NewErrorCollector creates a new error collector.
func NewErrorCollector() *ErrorCollector {
	return &ErrorCollector{
		errors:   make([]*StandardError, 0),
		warnings: make([]*StandardError, 0),
	}
}

// AddError adds an error to the collector.
func (ec *ErrorCollector) AddError(err *StandardError) {
	if err.Severity == SeverityLow || err.Severity == SeverityMedium {
		ec.warnings = append(ec.warnings, err)
	} else {
		ec.errors = append(ec.errors, err)
	}
}

// HasErrors returns true if there are any errors.
func (ec *ErrorCollector) HasErrors() bool {
	return len(ec.errors) > 0
}

// HasWarnings returns true if there are any warnings.
func (ec *ErrorCollector) HasWarnings() bool {
	return len(ec.warnings) > 0
}

// GetErrors returns all collected errors.
func (ec *ErrorCollector) GetErrors() []*StandardError {
	return ec.errors
}

// GetWarnings returns all collected warnings.
func (ec *ErrorCollector) GetWarnings() []*StandardError {
	return ec.warnings
}

// ToError converts the collector to a single error if there are errors.
func (ec *ErrorCollector) ToError() error {
	if len(ec.errors) == 0 {
		return nil
	}

	if len(ec.errors) == 1 {
		return ec.errors[0]
	}

	// Create a multi-error
	return NewStandardError(
		ErrorCodeOperationFailed,
		fmt.Sprintf("%d errors occurred during operation", len(ec.errors)),
		SeverityHigh,
	).WithContext("error_count", len(ec.errors)).
		WithContext("warning_count", len(ec.warnings))
}

// ErrorFilter provides filtering capabilities for errors.
type ErrorFilter struct {
	codes      []ErrorCode
	severities []ErrorSeverity
	retryable  *bool
}

// NewErrorFilter creates a new error filter.
func NewErrorFilter() *ErrorFilter {
	return &ErrorFilter{}
}

// WithCodes filters by error codes.
func (ef *ErrorFilter) WithCodes(codes ...ErrorCode) *ErrorFilter {
	ef.codes = codes
	return ef
}

// WithSeverities filters by severities.
func (ef *ErrorFilter) WithSeverities(severities ...ErrorSeverity) *ErrorFilter {
	ef.severities = severities
	return ef
}

// WithRetryable filters by retryable status.
func (ef *ErrorFilter) WithRetryable(retryable bool) *ErrorFilter {
	ef.retryable = &retryable
	return ef
}

// Matches returns true if the error matches the filter criteria.
func (ef *ErrorFilter) Matches(err *StandardError) bool {
	if len(ef.codes) > 0 {
		found := false
		for _, code := range ef.codes {
			if err.Code == code {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	if len(ef.severities) > 0 {
		found := false
		for _, severity := range ef.severities {
			if err.Severity == severity {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	if ef.retryable != nil && err.Retryable != *ef.retryable {
		return false
	}

	return true
}
