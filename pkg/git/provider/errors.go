// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package provider

import (
	"errors"
	"fmt"
	"net/http"
)

// Common provider errors that can be returned by any provider.
var (
	ErrNotFound              = errors.New("resource not found")
	ErrUnauthorized          = errors.New("unauthorized access")
	ErrForbidden             = errors.New("forbidden access")
	ErrRateLimitExceeded     = errors.New("rate limit exceeded")
	ErrNotSupported          = errors.New("operation not supported")
	ErrInvalidInput          = errors.New("invalid input")
	ErrConflict              = errors.New("resource conflict")
	ErrServiceUnavailable    = errors.New("service unavailable")
	ErrInternalError         = errors.New("internal server error")
	ErrBadRequest            = errors.New("bad request")
	ErrNetworkError          = errors.New("network error")
	ErrTimeoutError          = errors.New("timeout error")
	ErrProviderNotConfigured = errors.New("provider not configured")
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrTokenExpired          = errors.New("token expired")
	ErrQuotaExceeded         = errors.New("quota exceeded")
)

// ProviderError wraps provider-specific errors with additional context.
type ProviderError struct {
	Provider   string                 `json:"provider"`
	Operation  string                 `json:"operation"`
	Resource   string                 `json:"resource,omitempty"`
	StatusCode int                    `json:"status_code,omitempty"`
	ErrorCode  string                 `json:"error_code,omitempty"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details,omitempty"`
	Cause      error                  `json:"-"`
	Retryable  bool                   `json:"retryable"`
	RetryAfter int                    `json:"retry_after,omitempty"` // seconds
}

// Error implements the error interface.
func (e *ProviderError) Error() string {
	if e.Resource != "" {
		return fmt.Sprintf("%s: %s %s: %s", e.Provider, e.Operation, e.Resource, e.Message)
	}
	return fmt.Sprintf("%s: %s: %s", e.Provider, e.Operation, e.Message)
}

// Unwrap returns the underlying error for error inspection.
func (e *ProviderError) Unwrap() error {
	return e.Cause
}

// Is allows error comparison using errors.Is.
func (e *ProviderError) Is(target error) bool {
	if e.Cause != nil && errors.Is(e.Cause, target) {
		return true
	}

	// Check against common provider errors
	switch target {
	case ErrNotFound:
		return e.StatusCode == http.StatusNotFound
	case ErrUnauthorized:
		return e.StatusCode == http.StatusUnauthorized
	case ErrForbidden:
		return e.StatusCode == http.StatusForbidden
	case ErrRateLimitExceeded:
		return e.StatusCode == http.StatusTooManyRequests
	case ErrConflict:
		return e.StatusCode == http.StatusConflict
	case ErrServiceUnavailable:
		return e.StatusCode == http.StatusServiceUnavailable
	case ErrInternalError:
		return e.StatusCode == http.StatusInternalServerError
	case ErrBadRequest:
		return e.StatusCode == http.StatusBadRequest
	default:
		return false
	}
}

// NewProviderError creates a new provider error with the given details.
func NewProviderError(provider, operation string, err error) *ProviderError {
	pe := &ProviderError{
		Provider:  provider,
		Operation: operation,
		Cause:     err,
		Message:   err.Error(),
	}

	// Determine retryability based on error type
	pe.Retryable = isRetryableError(err)

	return pe
}

// NewProviderErrorWithDetails creates a new provider error with additional details.
func NewProviderErrorWithDetails(provider, operation, resource string, statusCode int, message string, details map[string]interface{}) *ProviderError {
	pe := &ProviderError{
		Provider:   provider,
		Operation:  operation,
		Resource:   resource,
		StatusCode: statusCode,
		Message:    message,
		Details:    details,
	}

	// Determine retryability based on status code
	pe.Retryable = isRetryableStatusCode(statusCode)

	return pe
}

// WithRetryAfter sets the retry after duration for the error.
func (e *ProviderError) WithRetryAfter(seconds int) *ProviderError {
	e.RetryAfter = seconds
	return e
}

// WithErrorCode sets the provider-specific error code.
func (e *ProviderError) WithErrorCode(code string) *ProviderError {
	e.ErrorCode = code
	return e
}

// WithResource sets the resource name for the error.
func (e *ProviderError) WithResource(resource string) *ProviderError {
	e.Resource = resource
	return e
}

// WithDetails adds additional details to the error.
func (e *ProviderError) WithDetails(details map[string]interface{}) *ProviderError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	for k, v := range details {
		e.Details[k] = v
	}
	return e
}

// ValidationError represents validation errors with field-specific details.
type ValidationError struct {
	Provider string                 `json:"provider"`
	Fields   map[string][]string    `json:"fields"`
	General  []string               `json:"general,omitempty"`
	Details  map[string]interface{} `json:"details,omitempty"`
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	if len(e.General) > 0 {
		return fmt.Sprintf("%s: validation error: %s", e.Provider, e.General[0])
	}
	if len(e.Fields) > 0 {
		for field, errors := range e.Fields {
			if len(errors) > 0 {
				return fmt.Sprintf("%s: validation error in field '%s': %s", e.Provider, field, errors[0])
			}
		}
	}
	return fmt.Sprintf("%s: validation error", e.Provider)
}

// AddFieldError adds a field-specific validation error.
func (e *ValidationError) AddFieldError(field, message string) {
	if e.Fields == nil {
		e.Fields = make(map[string][]string)
	}
	e.Fields[field] = append(e.Fields[field], message)
}

// AddGeneralError adds a general validation error.
func (e *ValidationError) AddGeneralError(message string) {
	e.General = append(e.General, message)
}

// HasErrors returns true if there are any validation errors.
func (e *ValidationError) HasErrors() bool {
	return len(e.Fields) > 0 || len(e.General) > 0
}

// NewValidationError creates a new validation error for the given provider.
func NewValidationError(provider string) *ValidationError {
	return &ValidationError{
		Provider: provider,
		Fields:   make(map[string][]string),
		General:  []string{},
	}
}

// ErrorClassifier helps classify and handle different types of errors.
type ErrorClassifier struct{}

// ClassifyError classifies an error and returns appropriate handling information.
func (ec *ErrorClassifier) ClassifyError(err error) ErrorClassification {
	if err == nil {
		return ErrorClassification{Type: ErrorTypeNone}
	}

	var pe *ProviderError
	if errors.As(err, &pe) {
		return ec.classifyProviderError(pe)
	}

	var ve *ValidationError
	if errors.As(err, &ve) {
		return ErrorClassification{
			Type:      ErrorTypeValidation,
			Retryable: false,
			Severity:  SeverityMedium,
		}
	}

	// Classify based on error content
	switch {
	case errors.Is(err, ErrNotFound):
		return ErrorClassification{Type: ErrorTypeNotFound, Retryable: false, Severity: SeverityLow}
	case errors.Is(err, ErrUnauthorized):
		return ErrorClassification{Type: ErrorTypeAuth, Retryable: false, Severity: SeverityHigh}
	case errors.Is(err, ErrRateLimitExceeded):
		return ErrorClassification{Type: ErrorTypeRateLimit, Retryable: true, Severity: SeverityMedium}
	case errors.Is(err, ErrNetworkError):
		return ErrorClassification{Type: ErrorTypeNetwork, Retryable: true, Severity: SeverityMedium}
	case errors.Is(err, ErrTimeoutError):
		return ErrorClassification{Type: ErrorTypeTimeout, Retryable: true, Severity: SeverityMedium}
	default:
		return ErrorClassification{Type: ErrorTypeUnknown, Retryable: false, Severity: SeverityHigh}
	}
}

func (ec *ErrorClassifier) classifyProviderError(pe *ProviderError) ErrorClassification {
	classification := ErrorClassification{
		Retryable: pe.Retryable,
	}

	switch pe.StatusCode {
	case http.StatusNotFound:
		classification.Type = ErrorTypeNotFound
		classification.Severity = SeverityLow
	case http.StatusUnauthorized:
		classification.Type = ErrorTypeAuth
		classification.Severity = SeverityHigh
	case http.StatusForbidden:
		classification.Type = ErrorTypeAuth
		classification.Severity = SeverityHigh
	case http.StatusTooManyRequests:
		classification.Type = ErrorTypeRateLimit
		classification.Severity = SeverityMedium
		classification.Retryable = true
	case http.StatusBadRequest:
		classification.Type = ErrorTypeValidation
		classification.Severity = SeverityMedium
	case http.StatusConflict:
		classification.Type = ErrorTypeConflict
		classification.Severity = SeverityMedium
	case http.StatusServiceUnavailable:
		classification.Type = ErrorTypeServiceUnavailable
		classification.Severity = SeverityHigh
		classification.Retryable = true
	case http.StatusInternalServerError:
		classification.Type = ErrorTypeInternal
		classification.Severity = SeverityHigh
		classification.Retryable = true
	default:
		classification.Type = ErrorTypeUnknown
		classification.Severity = SeverityMedium
	}

	return classification
}

// ErrorClassification represents the classification of an error.
type ErrorClassification struct {
	Type      ErrorType     `json:"type"`
	Retryable bool          `json:"retryable"`
	Severity  ErrorSeverity `json:"severity"`
}

// ErrorType represents different types of errors.
type ErrorType string

const (
	ErrorTypeNone               ErrorType = "none"
	ErrorTypeValidation         ErrorType = "validation"
	ErrorTypeAuth               ErrorType = "authentication"
	ErrorTypeNotFound           ErrorType = "not_found"
	ErrorTypeConflict           ErrorType = "conflict"
	ErrorTypeRateLimit          ErrorType = "rate_limit"
	ErrorTypeNetwork            ErrorType = "network"
	ErrorTypeTimeout            ErrorType = "timeout"
	ErrorTypeServiceUnavailable ErrorType = "service_unavailable"
	ErrorTypeInternal           ErrorType = "internal"
	ErrorTypeUnknown            ErrorType = "unknown"
)

// ErrorSeverity represents the severity level of an error.
type ErrorSeverity string

const (
	SeverityLow    ErrorSeverity = "low"
	SeverityMedium ErrorSeverity = "medium"
	SeverityHigh   ErrorSeverity = "high"
)

// Helper functions

// isRetryableError determines if an error is retryable based on its type.
func isRetryableError(err error) bool {
	switch {
	case errors.Is(err, ErrRateLimitExceeded):
		return true
	case errors.Is(err, ErrServiceUnavailable):
		return true
	case errors.Is(err, ErrNetworkError):
		return true
	case errors.Is(err, ErrTimeoutError):
		return true
	case errors.Is(err, ErrInternalError):
		return true
	default:
		return false
	}
}

// isRetryableStatusCode determines if a status code indicates a retryable error.
func isRetryableStatusCode(statusCode int) bool {
	switch statusCode {
	case http.StatusTooManyRequests:
		return true
	case http.StatusServiceUnavailable:
		return true
	case http.StatusBadGateway:
		return true
	case http.StatusGatewayTimeout:
		return true
	case http.StatusInternalServerError:
		return true
	default:
		return statusCode >= 500 && statusCode < 600
	}
}

// WrapError wraps an error with provider context.
func WrapError(provider, operation string, err error) error {
	if err == nil {
		return nil
	}

	// If it's already a provider error, don't wrap again
	var pe *ProviderError
	if errors.As(err, &pe) {
		return err
	}

	return NewProviderError(provider, operation, err)
}

// WrapHTTPError wraps an HTTP error with provider context and status code.
func WrapHTTPError(provider, operation, resource string, statusCode int, message string) error {
	return NewProviderErrorWithDetails(provider, operation, resource, statusCode, message, nil)
}
