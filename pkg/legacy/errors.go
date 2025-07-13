package legacy

import (
	"errors"
	"fmt"
	"strings"
)

// ErrorCode represents different types of errors in the system
type ErrorCode string

const (
	// Network related errors
	ErrorCodeNetworkConnection  ErrorCode = "NETWORK_CONNECTION"
	ErrorCodeNetworkTimeout     ErrorCode = "NETWORK_TIMEOUT"
	ErrorCodeNetworkDNS         ErrorCode = "NETWORK_DNS"
	ErrorCodeNetworkUnreachable ErrorCode = "NETWORK_UNREACHABLE"

	// VPN related errors
	ErrorCodeVPNConnection     ErrorCode = "VPN_CONNECTION"
	ErrorCodeVPNAuthentication ErrorCode = "VPN_AUTHENTICATION"
	ErrorCodeVPNConfiguration  ErrorCode = "VPN_CONFIGURATION"
	ErrorCodeVPNHierarchy      ErrorCode = "VPN_HIERARCHY"

	// Configuration related errors
	ErrorCodeConfigInvalid    ErrorCode = "CONFIG_INVALID"
	ErrorCodeConfigNotFound   ErrorCode = "CONFIG_NOT_FOUND"
	ErrorCodeConfigSyntax     ErrorCode = "CONFIG_SYNTAX"
	ErrorCodeConfigValidation ErrorCode = "CONFIG_VALIDATION"

	// Authentication related errors
	ErrorCodeAuthFailed  ErrorCode = "AUTH_FAILED"
	ErrorCodeAuthExpired ErrorCode = "AUTH_EXPIRED"
	ErrorCodeAuthMissing ErrorCode = "AUTH_MISSING"
	ErrorCodeAuthInvalid ErrorCode = "AUTH_INVALID"

	// Permission related errors
	ErrorCodePermissionDenied ErrorCode = "PERMISSION_DENIED"
	ErrorCodeResourceNotFound ErrorCode = "RESOURCE_NOT_FOUND"
	ErrorCodeResourceExists   ErrorCode = "RESOURCE_EXISTS"

	// System related errors
	ErrorCodeSystemInternal ErrorCode = "SYSTEM_INTERNAL"
	ErrorCodeSystemTimeout  ErrorCode = "SYSTEM_TIMEOUT"
	ErrorCodeSystemResource ErrorCode = "SYSTEM_RESOURCE"
)

// GzhError represents an enhanced error with additional context
type GzhError struct {
	Code        ErrorCode              `json:"code"`
	Message     string                 `json:"message"`
	Details     string                 `json:"details,omitempty"`
	Suggestions []string               `json:"suggestions,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Cause       error                  `json:"cause,omitempty"`
}

// Error implements the error interface
func (e *GzhError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause
func (e *GzhError) Unwrap() error {
	return e.Cause
}

// Is checks if the error matches the given error
func (e *GzhError) Is(target error) bool {
	if target == nil {
		return e == nil
	}

	if gzhErr, ok := target.(*GzhError); ok {
		return e.Code == gzhErr.Code
	}

	return errors.Is(e.Cause, target)
}

// WithContext adds context information to the error
func (e *GzhError) WithContext(key string, value interface{}) *GzhError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithSuggestion adds a suggestion for resolving the error
func (e *GzhError) WithSuggestion(suggestion string) *GzhError {
	e.Suggestions = append(e.Suggestions, suggestion)
	return e
}

// WithCause sets the underlying cause of the error
func (e *GzhError) WithCause(cause error) *GzhError {
	e.Cause = cause
	return e
}

// GetSuggestions returns formatted suggestions for resolving the error
func (e *GzhError) GetSuggestions() string {
	if len(e.Suggestions) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("Suggestions:\n")
	for i, suggestion := range e.Suggestions {
		sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, suggestion))
	}
	return sb.String()
}

// GetFormattedError returns a user-friendly formatted error message
func (e *GzhError) GetFormattedError() string {
	var sb strings.Builder

	// Main error message
	sb.WriteString(fmt.Sprintf("❌ Error: %s\n", e.Message))

	// Details if available
	if e.Details != "" {
		sb.WriteString(fmt.Sprintf("📝 Details: %s\n", e.Details))
	}

	// Context information
	if len(e.Context) > 0 {
		sb.WriteString("🔍 Context:\n")
		for key, value := range e.Context {
			sb.WriteString(fmt.Sprintf("   %s: %v\n", key, value))
		}
	}

	// Suggestions
	if len(e.Suggestions) > 0 {
		sb.WriteString("💡 Suggestions:\n")
		for i, suggestion := range e.Suggestions {
			sb.WriteString(fmt.Sprintf("   %d. %s\n", i+1, suggestion))
		}
	}

	return sb.String()
}

// Error creation functions

// NewNetworkError creates a new network-related error
func NewNetworkError(code ErrorCode, message string) *GzhError {
	return &GzhError{
		Code:    code,
		Message: message,
	}
}

// NewVPNError creates a new VPN-related error
func NewVPNError(code ErrorCode, message string) *GzhError {
	err := &GzhError{
		Code:    code,
		Message: message,
	}

	// Add common VPN troubleshooting suggestions
	switch code {
	case ErrorCodeVPNConnection:
		err.WithSuggestion("Check your internet connection")
		err.WithSuggestion("Verify VPN server address and port")
		err.WithSuggestion("Try connecting to a different VPN server")
	case ErrorCodeVPNAuthentication:
		err.WithSuggestion("Verify your username and password")
		err.WithSuggestion("Check if your VPN subscription is active")
		err.WithSuggestion("Try regenerating your VPN certificates")
	case ErrorCodeVPNConfiguration:
		err.WithSuggestion("Validate your VPN configuration file syntax")
		err.WithSuggestion("Check file permissions for VPN configuration")
		err.WithSuggestion("Ensure all required VPN configuration fields are present")
	}

	return err
}

// NewConfigError creates a new configuration-related error
func NewConfigError(code ErrorCode, message string) *GzhError {
	err := &GzhError{
		Code:    code,
		Message: message,
	}

	// Add common configuration troubleshooting suggestions
	switch code {
	case ErrorCodeConfigNotFound:
		err.WithSuggestion("Create a configuration file using: gz gen-config")
		err.WithSuggestion("Check if the configuration file path is correct")
		err.WithSuggestion("Verify file permissions")
	case ErrorCodeConfigInvalid:
		err.WithSuggestion("Validate configuration syntax using: gz validate")
		err.WithSuggestion("Check for missing required fields")
		err.WithSuggestion("Refer to the configuration schema documentation")
	case ErrorCodeConfigSyntax:
		err.WithSuggestion("Check YAML/JSON syntax for errors")
		err.WithSuggestion("Ensure proper indentation")
		err.WithSuggestion("Remove any trailing commas or invalid characters")
	}

	return err
}

// NewAuthError creates a new authentication-related error
func NewAuthError(code ErrorCode, message string) *GzhError {
	err := &GzhError{
		Code:    code,
		Message: message,
	}

	// Add common authentication troubleshooting suggestions
	switch code {
	case ErrorCodeAuthMissing:
		err.WithSuggestion("Set the required environment variable (e.g., GITHUB_TOKEN)")
		err.WithSuggestion("Configure authentication in your profile settings")
		err.WithSuggestion("Use the login command to authenticate")
	case ErrorCodeAuthExpired:
		err.WithSuggestion("Refresh your authentication token")
		err.WithSuggestion("Re-authenticate using the login command")
		err.WithSuggestion("Check token expiration settings")
	case ErrorCodeAuthFailed:
		err.WithSuggestion("Verify your credentials are correct")
		err.WithSuggestion("Check if two-factor authentication is required")
		err.WithSuggestion("Ensure you have the necessary permissions")
	}

	return err
}

// NewSystemError creates a new system-related error
func NewSystemError(code ErrorCode, message string) *GzhError {
	err := &GzhError{
		Code:    code,
		Message: message,
	}

	// Add common system troubleshooting suggestions
	switch code {
	case ErrorCodeSystemResource:
		err.WithSuggestion("Check available disk space")
		err.WithSuggestion("Verify memory availability")
		err.WithSuggestion("Close unnecessary applications")
	case ErrorCodeSystemTimeout:
		err.WithSuggestion("Try again with a longer timeout")
		err.WithSuggestion("Check network connectivity")
		err.WithSuggestion("Verify the remote service is responding")
	case ErrorCodePermissionDenied:
		err.WithSuggestion("Run the command with appropriate permissions")
		err.WithSuggestion("Check file/directory ownership")
		err.WithSuggestion("Verify your user has the required access rights")
	}

	return err
}

// Helper functions for common error scenarios

// WrapError wraps an existing error with additional context
func WrapError(err error, code ErrorCode, message string) *GzhError {
	return &GzhError{
		Code:    code,
		Message: message,
		Cause:   err,
	}
}

// FromError converts a standard error to a GzhError
func FromError(err error) *GzhError {
	if err == nil {
		return nil
	}

	if gzhErr, ok := err.(*GzhError); ok {
		return gzhErr
	}

	return &GzhError{
		Code:    ErrorCodeSystemInternal,
		Message: "An internal error occurred",
		Details: err.Error(),
		Cause:   err,
	}
}

// IsCode checks if an error has a specific error code
func IsCode(err error, code ErrorCode) bool {
	if gzhErr, ok := err.(*GzhError); ok {
		return gzhErr.Code == code
	}
	return false
}

// ExtractCode extracts the error code from an error
func ExtractCode(err error) ErrorCode {
	if gzhErr, ok := err.(*GzhError); ok {
		return gzhErr.Code
	}
	return ErrorCodeSystemInternal
}

// Recovery functions for automatic error recovery

// RecoveryStrategy represents a strategy for automatic error recovery
type RecoveryStrategy func(err *GzhError) error

// RecoveryManager manages automatic error recovery
type RecoveryManager struct {
	strategies map[ErrorCode]RecoveryStrategy
}

// NewRecoveryManager creates a new recovery manager
func NewRecoveryManager() *RecoveryManager {
	return &RecoveryManager{
		strategies: make(map[ErrorCode]RecoveryStrategy),
	}
}

// RegisterStrategy registers a recovery strategy for an error code
func (rm *RecoveryManager) RegisterStrategy(code ErrorCode, strategy RecoveryStrategy) {
	rm.strategies[code] = strategy
}

// Recover attempts to recover from an error automatically
func (rm *RecoveryManager) Recover(err error) error {
	gzhErr := FromError(err)
	if strategy, exists := rm.strategies[gzhErr.Code]; exists {
		return strategy(gzhErr)
	}
	return err
}

// Default recovery strategies

// NetworkRecoveryStrategy attempts to recover from network errors
func NetworkRecoveryStrategy(err *GzhError) error {
	switch err.Code {
	case ErrorCodeNetworkTimeout:
		// Could implement retry logic here
		return fmt.Errorf("network timeout recovery not implemented yet")
	case ErrorCodeNetworkDNS:
		// Could implement alternative DNS resolution
		return fmt.Errorf("DNS recovery not implemented yet")
	default:
		return err
	}
}

// VPNRecoveryStrategy attempts to recover from VPN errors
func VPNRecoveryStrategy(err *GzhError) error {
	switch err.Code {
	case ErrorCodeVPNConnection:
		// Could implement failover to alternative VPN servers
		return fmt.Errorf("VPN connection recovery not implemented yet")
	default:
		return err
	}
}
