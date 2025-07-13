package errors

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"syscall"
	"time"
)

// ErrorAdapter provides utilities to convert standard errors to UserErrors
type ErrorAdapter struct {
	defaultDomain string
	requestID     string
	context       map[string]interface{}
}

// NewErrorAdapter creates a new error adapter with optional default domain
func NewErrorAdapter(domain string) *ErrorAdapter {
	return &ErrorAdapter{
		defaultDomain: domain,
		context:       make(map[string]interface{}),
	}
}

// WithRequestID sets the request ID for all adapted errors
func (a *ErrorAdapter) WithRequestID(requestID string) *ErrorAdapter {
	a.requestID = requestID
	return a
}

// WithContext adds context that will be included in all adapted errors
func (a *ErrorAdapter) WithContext(key string, value interface{}) *ErrorAdapter {
	a.context[key] = value
	return a
}

// FromHTTPResponse converts HTTP response errors to UserErrors
func (a *ErrorAdapter) FromHTTPResponse(resp *http.Response, operation string) *UserError {
	if resp == nil {
		return a.fromUnknownError(fmt.Errorf("nil HTTP response"), operation)
	}

	domain := a.getProviderFromURL(resp.Request.URL.String())

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return a.fromAuthError(resp, domain, operation)
	case http.StatusForbidden:
		return a.fromPermissionError(resp, domain, operation)
	case http.StatusNotFound:
		return a.fromNotFoundError(resp, domain, operation)
	case http.StatusTooManyRequests:
		return a.fromRateLimitError(resp, domain, operation)
	case http.StatusRequestTimeout, http.StatusGatewayTimeout:
		return a.fromTimeoutError(resp, domain, operation)
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
		return a.fromServerError(resp, domain, operation)
	default:
		return a.fromHTTPError(resp, domain, operation)
	}
}

// FromOSError converts OS/filesystem errors to UserErrors
func (a *ErrorAdapter) FromOSError(err error, path string, operation string) *UserError {
	if err == nil {
		return nil
	}

	// Handle specific OS errors
	if os.IsNotExist(err) {
		return NewError(DomainFile, CategoryNotFound, "FILE_NOT_FOUND").
			Message(fmt.Sprintf("File not found: %s", path)).
			Description(fmt.Sprintf("The file '%s' does not exist", path)).
			Context("path", path).
			Context("operation", operation).
			Suggest("Check the file path and ensure it exists").
			Suggest("Verify you have the correct permissions").
			Cause(err).
			RequestID(a.requestID).
			Build()
	}

	if os.IsPermission(err) {
		permErr := FilePermissionError(path, operation)
		permErr.RequestID = a.requestID
		permErr.Cause = err
		return permErr
	}

	// Check for specific syscall errors
	if pathErr, ok := err.(*os.PathError); ok {
		if errno, ok := pathErr.Err.(syscall.Errno); ok {
			switch errno {
			case syscall.ENOSPC:
				return a.fromDiskSpaceError(path, operation, err)
			case syscall.ENOENT:
				return a.fromFileNotFoundError(path, operation, err)
			case syscall.EACCES:
				permErr := FilePermissionError(path, operation)
				permErr.Cause = err
				return permErr
			}
		}
	}

	return a.fromUnknownError(err, fmt.Sprintf("file operation: %s", operation))
}

// FromNetworkError converts network errors to UserErrors
func (a *ErrorAdapter) FromNetworkError(err error, operation string, host string) *UserError {
	if err == nil {
		return nil
	}

	errStr := err.Error()

	// Check for timeout
	if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "deadline exceeded") {
		timeoutErr := NetworkTimeoutError(operation, 30*time.Second)
		timeoutErr.RequestID = a.requestID
		timeoutErr.Context["host"] = host
		timeoutErr.Cause = err
		return timeoutErr
	}

	// Check for connection refused
	if strings.Contains(errStr, "connection refused") {
		return NewError(DomainNetwork, CategoryNetwork, "CONNECTION_REFUSED").
			Message("Connection refused").
			Description(fmt.Sprintf("Unable to connect to %s", host)).
			Context("host", host).
			Context("operation", operation).
			Suggest("Check if the service is running").
			Suggest("Verify the hostname and port").
			Suggest("Check firewall settings").
			Cause(err).
			RequestID(a.requestID).
			Build()
	}

	// Check for DNS resolution errors
	if strings.Contains(errStr, "no such host") || strings.Contains(errStr, "cannot resolve") {
		return NewError(DomainNetwork, CategoryNetwork, "DNS_RESOLUTION_FAILED").
			Message("DNS resolution failed").
			Description(fmt.Sprintf("Cannot resolve hostname: %s", host)).
			Context("host", host).
			Context("operation", operation).
			Suggest("Check the hostname spelling").
			Suggest("Verify DNS configuration").
			Suggest("Try using IP address instead").
			Cause(err).
			RequestID(a.requestID).
			Build()
	}

	return a.fromUnknownError(err, fmt.Sprintf("network operation: %s", operation))
}

// FromValidationError converts validation errors to UserErrors
func (a *ErrorAdapter) FromValidationError(field string, value interface{}, reason string) *UserError {
	return NewError(a.defaultDomain, CategoryValidation, "VALIDATION_FAILED").
		Message(fmt.Sprintf("Validation failed for field: %s", field)).
		Description(fmt.Sprintf("Field '%s' has invalid value '%v': %s", field, value, reason)).
		Context("field", field).
		Context("value", value).
		Context("reason", reason).
		Suggest("Check the field value and format").
		Suggest("Refer to the documentation for valid values").
		RequestID(a.requestID).
		Build()
}

// FromConfigError converts configuration errors to UserErrors
func (a *ErrorAdapter) FromConfigError(configType string, err error) *UserError {
	return NewError(DomainConfig, CategoryValidation, "CONFIG_ERROR").
		Message(fmt.Sprintf("Configuration error: %s", configType)).
		Description(fmt.Sprintf("Failed to load or validate %s configuration", configType)).
		Context("config_type", configType).
		Suggest("Check the configuration file format").
		Suggest("Validate configuration syntax").
		Suggest("Ensure all required fields are present").
		Cause(err).
		RequestID(a.requestID).
		Build()
}

// Helper methods for specific error types
func (a *ErrorAdapter) fromAuthError(resp *http.Response, domain, operation string) *UserError {
	return NewError(domain, CategoryAuth, "AUTHENTICATION_FAILED").
		Message("Authentication failed").
		Description("The request failed due to invalid or missing authentication").
		Context("status_code", resp.StatusCode).
		Context("operation", operation).
		Context("url", resp.Request.URL.String()).
		Suggest("Check your authentication token").
		Suggest("Ensure the token has required permissions").
		Suggest("Verify the token hasn't expired").
		RequestID(a.requestID).
		Build()
}

func (a *ErrorAdapter) fromPermissionError(resp *http.Response, domain, operation string) *UserError {
	return NewError(domain, CategoryPermission, "ACCESS_FORBIDDEN").
		Message("Access forbidden").
		Description("You don't have permission to perform this operation").
		Context("status_code", resp.StatusCode).
		Context("operation", operation).
		Context("url", resp.Request.URL.String()).
		Suggest("Check your access permissions").
		Suggest("Contact an administrator for access").
		Suggest("Verify you're using the correct account").
		RequestID(a.requestID).
		Build()
}

func (a *ErrorAdapter) fromNotFoundError(resp *http.Response, domain, operation string) *UserError {
	return NewError(domain, CategoryNotFound, "RESOURCE_NOT_FOUND").
		Message("Resource not found").
		Description("The requested resource could not be found").
		Context("status_code", resp.StatusCode).
		Context("operation", operation).
		Context("url", resp.Request.URL.String()).
		Suggest("Check the resource identifier").
		Suggest("Verify the resource exists").
		Suggest("Ensure you have access to the resource").
		RequestID(a.requestID).
		Build()
}

func (a *ErrorAdapter) fromRateLimitError(resp *http.Response, domain, operation string) *UserError {
	resetTime := time.Now().Add(time.Hour) // Default 1 hour

	// Try to parse rate limit reset header
	if resetHeader := resp.Header.Get("X-RateLimit-Reset"); resetHeader != "" {
		// Implementation would parse the reset time
	}

	rateLimitErr := APIRateLimitError(domain, resetTime)
	rateLimitErr.RequestID = a.requestID
	rateLimitErr.Context["operation"] = operation
	rateLimitErr.Context["url"] = resp.Request.URL.String()
	return rateLimitErr
}

func (a *ErrorAdapter) fromTimeoutError(resp *http.Response, domain, operation string) *UserError {
	timeoutErr := NetworkTimeoutError(operation, 30*time.Second)
	timeoutErr.RequestID = a.requestID
	timeoutErr.Context["status_code"] = resp.StatusCode
	timeoutErr.Context["url"] = resp.Request.URL.String()
	return timeoutErr
}

func (a *ErrorAdapter) fromServerError(resp *http.Response, domain, operation string) *UserError {
	return NewError(domain, CategoryNetwork, "SERVER_ERROR").
		Message("Server error").
		Description("The server encountered an error while processing the request").
		Context("status_code", resp.StatusCode).
		Context("operation", operation).
		Context("url", resp.Request.URL.String()).
		Suggest("Try the request again later").
		Suggest("Check the service status").
		Suggest("Contact support if the problem persists").
		RequestID(a.requestID).
		Build()
}

func (a *ErrorAdapter) fromHTTPError(resp *http.Response, domain, operation string) *UserError {
	return NewError(domain, CategoryNetwork, "HTTP_ERROR").
		Message(fmt.Sprintf("HTTP %d error", resp.StatusCode)).
		Description(fmt.Sprintf("Request failed with status code %d", resp.StatusCode)).
		Context("status_code", resp.StatusCode).
		Context("operation", operation).
		Context("url", resp.Request.URL.String()).
		Suggest("Check the request parameters").
		Suggest("Review the API documentation").
		RequestID(a.requestID).
		Build()
}

func (a *ErrorAdapter) fromDiskSpaceError(path, operation string, err error) *UserError {
	return NewError(DomainFile, CategoryResource, "DISK_SPACE_FULL").
		Message("Insufficient disk space").
		Description("The operation failed because there is not enough disk space").
		Context("path", path).
		Context("operation", operation).
		Suggest("Free up disk space").
		Suggest("Choose a different location with more space").
		Suggest("Clean up temporary files").
		Cause(err).
		RequestID(a.requestID).
		Build()
}

func (a *ErrorAdapter) fromFileNotFoundError(path, operation string, err error) *UserError {
	return NewError(DomainFile, CategoryNotFound, "FILE_NOT_FOUND").
		Message("File not found").
		Description(fmt.Sprintf("The file '%s' does not exist", path)).
		Context("path", path).
		Context("operation", operation).
		Suggest("Check the file path").
		Suggest("Ensure the file exists").
		Suggest("Verify directory permissions").
		Cause(err).
		RequestID(a.requestID).
		Build()
}

func (a *ErrorAdapter) fromUnknownError(err error, operation string) *UserError {
	domain := a.defaultDomain
	if domain == "" {
		domain = "unknown"
	}

	return NewError(domain, "unknown", "UNKNOWN_ERROR").
		Message("An unexpected error occurred").
		Description(fmt.Sprintf("Operation '%s' failed with an unknown error", operation)).
		Context("operation", operation).
		Suggest("Try the operation again").
		Suggest("Check the logs for more details").
		Suggest("Contact support if the problem persists").
		Cause(err).
		RequestID(a.requestID).
		Build()
}

func (a *ErrorAdapter) getProviderFromURL(url string) string {
	if strings.Contains(url, "github.com") || strings.Contains(url, "api.github.com") {
		return DomainGitHub
	}
	if strings.Contains(url, "gitlab.com") {
		return DomainGitLab
	}
	if strings.Contains(url, "gitea") {
		return DomainGitea
	}
	return a.defaultDomain
}

// Global convenience functions
func AdaptHTTPError(resp *http.Response, operation string) *UserError {
	return NewErrorAdapter("api").FromHTTPResponse(resp, operation)
}

func AdaptOSError(err error, path string, operation string) *UserError {
	return NewErrorAdapter("file").FromOSError(err, path, operation)
}

func AdaptNetworkError(err error, operation string, host string) *UserError {
	return NewErrorAdapter("network").FromNetworkError(err, operation, host)
}

func AdaptWithContext(ctx context.Context, err error, operation string) *UserError {
	if err == nil {
		return nil
	}

	adapter := NewErrorAdapter("app")
	if requestID := GetRequestIDFromContext(ctx); requestID != "" {
		adapter = adapter.WithRequestID(requestID)
	}

	// Try to determine error type and convert appropriately
	return adapter.fromUnknownError(err, operation)
}
