// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package clone

import (
	"errors"
	"fmt"
)

// Common clone errors
var (
	ErrMissingProvider         = errors.New("provider is required")
	ErrMissingOrganization     = errors.New("organization is required")
	ErrInvalidStrategy         = errors.New("invalid clone strategy")
	ErrInvalidFormat           = errors.New("invalid output format")
	ErrInvalidProtocol         = errors.New("invalid protocol, must be 'https' or 'ssh'")
	ErrInvalidVisibility       = errors.New("invalid visibility, must be 'all', 'public', or 'private'")
	ErrInvalidMatchPattern     = errors.New("invalid match pattern")
	ErrInvalidExcludePattern   = errors.New("invalid exclude pattern")
	ErrSessionNotFound         = errors.New("session not found")
	ErrSessionInvalid          = errors.New("session is invalid")
	ErrCloneInProgress         = errors.New("clone operation already in progress")
	ErrRepositoryExists        = errors.New("repository already exists")
	ErrRepositoryNotFound      = errors.New("repository not found")
	ErrGitCommandFailed        = errors.New("git command failed")
	ErrNetworkError            = errors.New("network error")
	ErrAuthenticationFailed    = errors.New("authentication failed")
	ErrRateLimitExceeded       = errors.New("rate limit exceeded")
	ErrInsufficientPermissions = errors.New("insufficient permissions")
	ErrTargetPathInvalid       = errors.New("target path is invalid")
	ErrDiskSpaceInsufficient   = errors.New("insufficient disk space")
)

// CloneError represents an error that occurred during cloning operations.
type CloneError struct {
	Repository string `json:"repository"`
	Operation  string `json:"operation"`
	Message    string `json:"message"`
	Cause      error  `json:"-"`
	Retryable  bool   `json:"retryable"`
	Code       string `json:"code"`
}

// Error implements the error interface.
func (e *CloneError) Error() string {
	if e.Repository != "" {
		return fmt.Sprintf("clone error for %s: %s: %s", e.Repository, e.Operation, e.Message)
	}
	return fmt.Sprintf("clone error: %s: %s", e.Operation, e.Message)
}

// Unwrap returns the underlying error.
func (e *CloneError) Unwrap() error {
	return e.Cause
}

// Is allows error comparison using errors.Is.
func (e *CloneError) Is(target error) bool {
	if e.Cause != nil && errors.Is(e.Cause, target) {
		return true
	}
	return false
}

// NewCloneError creates a new clone error.
func NewCloneError(repository, operation, message string, cause error) *CloneError {
	return &CloneError{
		Repository: repository,
		Operation:  operation,
		Message:    message,
		Cause:      cause,
		Retryable:  isRetryableError(cause),
	}
}

// NewCloneErrorWithCode creates a new clone error with a specific error code.
func NewCloneErrorWithCode(repository, operation, message, code string, cause error) *CloneError {
	return &CloneError{
		Repository: repository,
		Operation:  operation,
		Message:    message,
		Code:       code,
		Cause:      cause,
		Retryable:  isRetryableError(cause),
	}
}

// SessionError represents an error related to session management.
type SessionError struct {
	SessionID string `json:"session_id"`
	Operation string `json:"operation"`
	Message   string `json:"message"`
	Cause     error  `json:"-"`
}

// Error implements the error interface.
func (e *SessionError) Error() string {
	return fmt.Sprintf("session error [%s]: %s: %s", e.SessionID, e.Operation, e.Message)
}

// Unwrap returns the underlying error.
func (e *SessionError) Unwrap() error {
	return e.Cause
}

// NewSessionError creates a new session error.
func NewSessionError(sessionID, operation, message string, cause error) *SessionError {
	return &SessionError{
		SessionID: sessionID,
		Operation: operation,
		Message:   message,
		Cause:     cause,
	}
}

// ValidationError represents a validation error.
type ValidationError struct {
	Field   string `json:"field"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s': %s", e.Field, e.Message)
}

// NewValidationError creates a new validation error.
func NewValidationError(field, value, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Value:   value,
		Message: message,
	}
}

// ErrorSummary represents a summary of errors from a clone operation.
type ErrorSummary struct {
	TotalErrors     int                 `json:"total_errors"`
	RetryableErrors int                 `json:"retryable_errors"`
	FatalErrors     int                 `json:"fatal_errors"`
	ErrorsByType    map[string]int      `json:"errors_by_type"`
	ErrorsByRepo    map[string][]string `json:"errors_by_repo"`
	CommonErrors    []string            `json:"common_errors"`
	Recommendations []string            `json:"recommendations"`
}

// NewErrorSummary creates a new error summary.
func NewErrorSummary() *ErrorSummary {
	return &ErrorSummary{
		ErrorsByType:    make(map[string]int),
		ErrorsByRepo:    make(map[string][]string),
		CommonErrors:    make([]string, 0),
		Recommendations: make([]string, 0),
	}
}

// AddError adds an error to the summary.
func (s *ErrorSummary) AddError(repository string, err error) {
	s.TotalErrors++

	// Classify error type
	errorType := classifyError(err)
	s.ErrorsByType[errorType]++

	// Check if retryable
	if isRetryableError(err) {
		s.RetryableErrors++
	} else {
		s.FatalErrors++
	}

	// Add to repository-specific errors
	if repository != "" {
		s.ErrorsByRepo[repository] = append(s.ErrorsByRepo[repository], err.Error())
	}

	// Track common errors
	errorMessage := err.Error()
	found := false
	for _, commonError := range s.CommonErrors {
		if commonError == errorMessage {
			found = true
			break
		}
	}
	if !found && s.ErrorsByType[errorType] > 1 {
		s.CommonErrors = append(s.CommonErrors, errorMessage)
	}
}

// GenerateRecommendations generates recommendations based on the error patterns.
func (s *ErrorSummary) GenerateRecommendations() {
	s.Recommendations = make([]string, 0)

	if s.ErrorsByType["authentication"] > 0 {
		s.Recommendations = append(s.Recommendations,
			"Check your authentication credentials (token, username/password)")
	}

	if s.ErrorsByType["network"] > 0 {
		s.Recommendations = append(s.Recommendations,
			"Check your network connection and try again")
	}

	if s.ErrorsByType["rate_limit"] > 0 {
		s.Recommendations = append(s.Recommendations,
			"Rate limit exceeded. Wait before retrying or reduce parallel workers")
	}

	if s.ErrorsByType["disk_space"] > 0 {
		s.Recommendations = append(s.Recommendations,
			"Free up disk space before continuing")
	}

	if s.ErrorsByType["permissions"] > 0 {
		s.Recommendations = append(s.Recommendations,
			"Check repository permissions and access rights")
	}

	if s.RetryableErrors > s.FatalErrors {
		s.Recommendations = append(s.Recommendations,
			"Most errors are retryable. Consider using --resume to retry failed operations")
	}

	if len(s.ErrorsByRepo) > len(s.ErrorsByRepo)/2 {
		s.Recommendations = append(s.Recommendations,
			"High failure rate detected. Check provider connectivity and credentials")
	}
}

// Helper functions

// isRetryableError determines if an error is retryable.
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check known retryable errors
	switch {
	case errors.Is(err, ErrNetworkError):
		return true
	case errors.Is(err, ErrRateLimitExceeded):
		return true
	case errors.Is(err, ErrGitCommandFailed):
		return true // Git commands can often be retried
	default:
		return false
	}
}

// classifyError classifies an error into a category.
func classifyError(err error) string {
	if err == nil {
		return "unknown"
	}

	switch {
	case errors.Is(err, ErrAuthenticationFailed):
		return "authentication"
	case errors.Is(err, ErrNetworkError):
		return "network"
	case errors.Is(err, ErrRateLimitExceeded):
		return "rate_limit"
	case errors.Is(err, ErrInsufficientPermissions):
		return "permissions"
	case errors.Is(err, ErrDiskSpaceInsufficient):
		return "disk_space"
	case errors.Is(err, ErrGitCommandFailed):
		return "git"
	case errors.Is(err, ErrRepositoryExists):
		return "conflict"
	case errors.Is(err, ErrRepositoryNotFound):
		return "not_found"
	case errors.Is(err, ErrTargetPathInvalid):
		return "validation"
	default:
		return "unknown"
	}
}

// WrapGitError wraps a git command error with additional context.
func WrapGitError(repository, command string, err error, output []byte) error {
	message := fmt.Sprintf("git %s failed", command)
	if len(output) > 0 {
		message += fmt.Sprintf(": %s", string(output))
	}

	return NewCloneError(repository, "git_command", message, err)
}

// WrapNetworkError wraps a network error with clone context.
func WrapNetworkError(repository, operation string, err error) error {
	return NewCloneError(repository, operation, "network operation failed", err)
}

// WrapAuthError wraps an authentication error with clone context.
func WrapAuthError(repository, provider string, err error) error {
	message := fmt.Sprintf("authentication failed for %s", provider)
	return NewCloneError(repository, "authentication", message, err)
}
