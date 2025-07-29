// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package legacy

import (
	"errors"
	"fmt"
	"strings"
)

// ErrorCode represents different types of errors in the system.
type ErrorCode string

const (
	// ErrorCodeNetworkConnection represents network connection errors.
	ErrorCodeNetworkConnection ErrorCode = "NETWORK_CONNECTION"
	// ErrorCodeNetworkTimeout represents network timeout errors.
	ErrorCodeNetworkTimeout ErrorCode = "NETWORK_TIMEOUT"
	// ErrorCodeNetworkDNS represents DNS resolution errors.
	ErrorCodeNetworkDNS ErrorCode = "NETWORK_DNS"
	// ErrorCodeNetworkUnreachable represents network unreachable errors.
	ErrorCodeNetworkUnreachable ErrorCode = "NETWORK_UNREACHABLE"

	// ErrorCodeVPNConnection represents VPN connection errors.
	ErrorCodeVPNConnection ErrorCode = "VPN_CONNECTION"
	// ErrorCodeVPNAuthentication represents VPN authentication errors.
	ErrorCodeVPNAuthentication ErrorCode = "VPN_AUTHENTICATION"
	// ErrorCodeVPNConfiguration represents VPN configuration errors.
	ErrorCodeVPNConfiguration ErrorCode = "VPN_CONFIGURATION"
	// ErrorCodeVPNHierarchy represents VPN hierarchy errors.
	ErrorCodeVPNHierarchy ErrorCode = "VPN_HIERARCHY"

	// ErrorCodeConfigInvalid represents invalid configuration errors.
	ErrorCodeConfigInvalid ErrorCode = "CONFIG_INVALID"
	// ErrorCodeConfigNotFound represents configuration not found errors.
	ErrorCodeConfigNotFound ErrorCode = "CONFIG_NOT_FOUND"
	// ErrorCodeConfigSyntax represents configuration syntax errors.
	ErrorCodeConfigSyntax ErrorCode = "CONFIG_SYNTAX"
	// ErrorCodeConfigValidation represents configuration validation errors.
	ErrorCodeConfigValidation ErrorCode = "CONFIG_VALIDATION"

	// ErrorCodeAuthFailed represents authentication failure errors.
	ErrorCodeAuthFailed ErrorCode = "AUTH_FAILED"
	// ErrorCodeAuthExpired represents expired authentication errors.
	ErrorCodeAuthExpired ErrorCode = "AUTH_EXPIRED"
	// ErrorCodeAuthMissing represents missing authentication errors.
	ErrorCodeAuthMissing ErrorCode = "AUTH_MISSING"
	// ErrorCodeAuthInvalid represents invalid authentication errors.
	ErrorCodeAuthInvalid ErrorCode = "AUTH_INVALID"

	// ErrorCodePermissionDenied represents permission denied errors.
	ErrorCodePermissionDenied ErrorCode = "PERMISSION_DENIED"
	// ErrorCodeResourceNotFound represents resource not found errors.
	ErrorCodeResourceNotFound ErrorCode = "RESOURCE_NOT_FOUND"
	// ErrorCodeResourceExists represents resource already exists errors.
	ErrorCodeResourceExists ErrorCode = "RESOURCE_EXISTS"

	// ErrorCodeSystemInternal represents internal system errors.
	ErrorCodeSystemInternal ErrorCode = "SYSTEM_INTERNAL"
	// ErrorCodeSystemTimeout represents system timeout errors.
	ErrorCodeSystemTimeout ErrorCode = "SYSTEM_TIMEOUT"
	// ErrorCodeSystemResource represents system resource errors.
	ErrorCodeSystemResource ErrorCode = "SYSTEM_RESOURCE"
)

// GzhError represents an enhanced error with additional context.
type GzhError struct {
	Code        ErrorCode              `json:"code"`
	Message     string                 `json:"message"`
	Details     string                 `json:"details,omitempty"`
	Suggestions []string               `json:"suggestions,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Cause       error                  `json:"cause,omitempty"`
}

// Error implements the error interface.
func (e *GzhError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Code, e.Message, e.Details)
	}

	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause.
func (e *GzhError) Unwrap() error {
	return e.Cause
}

// Is checks if the error matches the given error.
func (e *GzhError) Is(target error) bool {
	if target == nil {
		return e == nil
	}

	if gzhErr, ok := target.(*GzhError); ok {
		return e.Code == gzhErr.Code
	}

	return errors.Is(e.Cause, target)
}

// WithContext adds context information to the error.
func (e *GzhError) WithContext(key string, value interface{}) *GzhError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}

	e.Context[key] = value

	return e
}

// WithSuggestion adds a suggestion for resolving the error.
func (e *GzhError) WithSuggestion(suggestion string) *GzhError {
	e.Suggestions = append(e.Suggestions, suggestion)
	return e
}

// WithCause sets the underlying cause of the error.
func (e *GzhError) WithCause(cause error) *GzhError {
	e.Cause = cause
	return e
}

// GetSuggestions returns formatted suggestions for resolving the error.
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

// GetFormattedError returns a user-friendly formatted error message.
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

// NewNetworkError creates a new network-related error.
func NewNetworkError(code ErrorCode, message string) *GzhError {
	return &GzhError{
		Code:    code,
		Message: message,
	}
}

// NewVPNError creates a new VPN-related error.
func NewVPNError(code ErrorCode, message string) *GzhError {
	err := &GzhError{
		Code:    code,
		Message: message,
	}

	// Add common VPN troubleshooting suggestions
	switch code {
	case ErrorCodeVPNConnection:
		_ = err.WithSuggestion("Check your internet connection").
			WithSuggestion("Verify VPN server address and port").
			WithSuggestion("Try connecting to a different VPN server")
	case ErrorCodeVPNAuthentication:
		_ = err.WithSuggestion("Verify your username and password").
			WithSuggestion("Check if your VPN subscription is active").
			WithSuggestion("Try regenerating your VPN certificates")
	case ErrorCodeVPNConfiguration:
		_ = err.WithSuggestion("Validate your VPN configuration file syntax").
			WithSuggestion("Check file permissions for VPN configuration").
			WithSuggestion("Ensure all required VPN configuration fields are present")
	case ErrorCodeVPNHierarchy:
		_ = err.WithSuggestion("Check VPN hierarchy configuration").
			WithSuggestion("Verify VPN connection dependencies").
			WithSuggestion("Ensure VPN connections are properly layered")
	case ErrorCodeNetworkConnection:
		_ = err.WithSuggestion("Check network connectivity").
			WithSuggestion("Verify network adapter settings").
			WithSuggestion("Try restarting network services")
	case ErrorCodeNetworkTimeout:
		_ = err.WithSuggestion("Increase network timeout settings").
			WithSuggestion("Check for network congestion").
			WithSuggestion("Verify server responsiveness")
	case ErrorCodeNetworkDNS:
		_ = err.WithSuggestion("Check DNS server configuration").
			WithSuggestion("Try alternative DNS servers").
			WithSuggestion("Flush DNS cache")
	case ErrorCodeNetworkUnreachable:
		_ = err.WithSuggestion("Check network routing configuration").
			WithSuggestion("Verify firewall settings").
			WithSuggestion("Ensure network is available")
	case ErrorCodeConfigInvalid, ErrorCodeConfigNotFound, ErrorCodeConfigSyntax, ErrorCodeConfigValidation:
		_ = err.WithSuggestion("Check configuration file syntax").
			WithSuggestion("Validate configuration against schema").
			WithSuggestion("Ensure all required fields are present")
	case ErrorCodeAuthFailed, ErrorCodeAuthExpired, ErrorCodeAuthMissing, ErrorCodeAuthInvalid:
		_ = err.WithSuggestion("Check authentication credentials").
			WithSuggestion("Verify token or certificate validity").
			WithSuggestion("Re-authenticate if necessary")
	case ErrorCodePermissionDenied, ErrorCodeResourceNotFound, ErrorCodeResourceExists:
		_ = err.WithSuggestion("Check access permissions").
			WithSuggestion("Verify resource availability").
			WithSuggestion("Ensure proper authorization")
	case ErrorCodeSystemInternal, ErrorCodeSystemTimeout, ErrorCodeSystemResource:
		_ = err.WithSuggestion("Check system resources").
			WithSuggestion("Monitor system performance").
			WithSuggestion("Contact system administrator if needed")
	default:
		// No specific suggestions for other error codes
	}

	return err
}

// NewConfigError creates a new configuration-related error.
func NewConfigError(code ErrorCode, message string) *GzhError {
	err := &GzhError{
		Code:    code,
		Message: message,
	}

	// Add common configuration troubleshooting suggestions
	switch code {
	case ErrorCodeConfigNotFound:
		_ = err.WithSuggestion("Create a configuration file using: gz synclone config generate init")
		_ = err.WithSuggestion("Check if the configuration file path is correct")
		_ = err.WithSuggestion("Verify file permissions")
	case ErrorCodeConfigInvalid:
		_ = err.WithSuggestion("Validate configuration syntax using: gz validate")
		_ = err.WithSuggestion("Check for missing required fields")
		_ = err.WithSuggestion("Refer to the configuration schema documentation")
	case ErrorCodeConfigSyntax:
		_ = err.WithSuggestion("Check YAML/JSON syntax for errors")
		_ = err.WithSuggestion("Ensure proper indentation")
		_ = err.WithSuggestion("Remove any trailing commas or invalid characters")
	case ErrorCodeConfigValidation:
		_ = err.WithSuggestion("Check configuration validation rules")
		_ = err.WithSuggestion("Ensure all fields meet validation criteria")
		_ = err.WithSuggestion("Review configuration schema requirements")
	case ErrorCodeNetworkConnection, ErrorCodeNetworkTimeout, ErrorCodeNetworkDNS, ErrorCodeNetworkUnreachable:
		_ = err.WithSuggestion("Check network connectivity")
		_ = err.WithSuggestion("Verify network configuration")
		_ = err.WithSuggestion("Ensure network services are running")
	case ErrorCodeVPNConnection, ErrorCodeVPNAuthentication, ErrorCodeVPNConfiguration, ErrorCodeVPNHierarchy:
		_ = err.WithSuggestion("Check VPN configuration")
		_ = err.WithSuggestion("Verify VPN connectivity")
		_ = err.WithSuggestion("Ensure VPN services are running")
	case ErrorCodeAuthFailed, ErrorCodeAuthExpired, ErrorCodeAuthMissing, ErrorCodeAuthInvalid:
		_ = err.WithSuggestion("Check authentication configuration")
		_ = err.WithSuggestion("Verify authentication credentials")
		_ = err.WithSuggestion("Ensure authentication services are available")
	case ErrorCodePermissionDenied, ErrorCodeResourceNotFound, ErrorCodeResourceExists:
		_ = err.WithSuggestion("Check resource permissions")
		_ = err.WithSuggestion("Verify resource availability")
		_ = err.WithSuggestion("Ensure proper access rights")
	case ErrorCodeSystemInternal, ErrorCodeSystemTimeout, ErrorCodeSystemResource:
		_ = err.WithSuggestion("Check system configuration")
		_ = err.WithSuggestion("Monitor system resources")
		_ = err.WithSuggestion("Contact system administrator if needed")
	default:
		// No specific suggestions for other error codes
	}

	return err
}

// NewAuthError creates a new authentication-related error.
func NewAuthError(code ErrorCode, message string) *GzhError {
	err := &GzhError{
		Code:    code,
		Message: message,
	}

	// Add common authentication troubleshooting suggestions
	switch code {
	case ErrorCodeAuthMissing:
		_ = err.WithSuggestion("Set the required environment variable (e.g., GITHUB_TOKEN)")
		_ = err.WithSuggestion("Configure authentication in your profile settings")
		_ = err.WithSuggestion("Use the login command to authenticate")
	case ErrorCodeAuthExpired:
		_ = err.WithSuggestion("Refresh your authentication token")
		_ = err.WithSuggestion("Re-authenticate using the login command")
		_ = err.WithSuggestion("Check token expiration settings")
	case ErrorCodeAuthFailed:
		_ = err.WithSuggestion("Verify your credentials are correct")
		_ = err.WithSuggestion("Check if two-factor authentication is required")
		_ = err.WithSuggestion("Ensure you have the necessary permissions")
	case ErrorCodeAuthInvalid:
		_ = err.WithSuggestion("Check authentication token format")
		_ = err.WithSuggestion("Verify token is not corrupted")
		_ = err.WithSuggestion("Generate a new authentication token")
	case ErrorCodeNetworkConnection, ErrorCodeNetworkTimeout, ErrorCodeNetworkDNS, ErrorCodeNetworkUnreachable:
		_ = err.WithSuggestion("Check network connectivity for authentication")
		_ = err.WithSuggestion("Verify authentication server availability")
		_ = err.WithSuggestion("Ensure network allows authentication traffic")
	case ErrorCodeVPNConnection, ErrorCodeVPNAuthentication, ErrorCodeVPNConfiguration, ErrorCodeVPNHierarchy:
		_ = err.WithSuggestion("Check VPN authentication configuration")
		_ = err.WithSuggestion("Verify VPN credentials")
		_ = err.WithSuggestion("Ensure VPN authentication servers are accessible")
	case ErrorCodeConfigInvalid, ErrorCodeConfigNotFound, ErrorCodeConfigSyntax, ErrorCodeConfigValidation:
		_ = err.WithSuggestion("Check authentication configuration files")
		_ = err.WithSuggestion("Verify authentication configuration syntax")
		_ = err.WithSuggestion("Ensure authentication configuration is complete")
	case ErrorCodePermissionDenied, ErrorCodeResourceNotFound, ErrorCodeResourceExists:
		_ = err.WithSuggestion("Check authentication permissions")
		_ = err.WithSuggestion("Verify access to authentication resources")
		_ = err.WithSuggestion("Ensure proper authentication scope")
	case ErrorCodeSystemInternal, ErrorCodeSystemTimeout, ErrorCodeSystemResource:
		_ = err.WithSuggestion("Check authentication system resources")
		_ = err.WithSuggestion("Monitor authentication service performance")
		_ = err.WithSuggestion("Contact authentication administrator if needed")
	default:
		// No specific suggestions for other error codes
	}

	return err
}

// NewSystemError creates a new system-related error.
func NewSystemError(code ErrorCode, message string) *GzhError {
	err := &GzhError{
		Code:    code,
		Message: message,
	}

	// Add common system troubleshooting suggestions
	switch code {
	case ErrorCodeSystemResource:
		_ = err.WithSuggestion("Check available disk space")
		_ = err.WithSuggestion("Verify memory availability")
		_ = err.WithSuggestion("Close unnecessary applications")
	case ErrorCodeSystemTimeout:
		_ = err.WithSuggestion("Try again with a longer timeout")
		_ = err.WithSuggestion("Check network connectivity")
		_ = err.WithSuggestion("Verify the remote service is responding")
	case ErrorCodeSystemInternal:
		_ = err.WithSuggestion("Check system logs for details")
		_ = err.WithSuggestion("Restart the service if necessary")
		_ = err.WithSuggestion("Contact system administrator for assistance")
	case ErrorCodePermissionDenied:
		_ = err.WithSuggestion("Run the command with appropriate permissions")
		_ = err.WithSuggestion("Check file/directory ownership")
		_ = err.WithSuggestion("Verify your user has the required access rights")
	case ErrorCodeResourceNotFound:
		_ = err.WithSuggestion("Verify the resource path is correct")
		_ = err.WithSuggestion("Check if the resource exists")
		_ = err.WithSuggestion("Ensure proper resource configuration")
	case ErrorCodeResourceExists:
		_ = err.WithSuggestion("Choose a different resource name")
		_ = err.WithSuggestion("Remove the existing resource first")
		_ = err.WithSuggestion("Use force flag if appropriate")
	case ErrorCodeNetworkConnection, ErrorCodeNetworkTimeout, ErrorCodeNetworkDNS, ErrorCodeNetworkUnreachable:
		_ = err.WithSuggestion("Check system network configuration")
		_ = err.WithSuggestion("Verify network connectivity")
		_ = err.WithSuggestion("Ensure network services are running")
	case ErrorCodeVPNConnection, ErrorCodeVPNAuthentication, ErrorCodeVPNConfiguration, ErrorCodeVPNHierarchy:
		_ = err.WithSuggestion("Check system VPN configuration")
		_ = err.WithSuggestion("Verify VPN connectivity")
		_ = err.WithSuggestion("Ensure VPN services are operational")
	case ErrorCodeConfigInvalid, ErrorCodeConfigNotFound, ErrorCodeConfigSyntax, ErrorCodeConfigValidation:
		_ = err.WithSuggestion("Check system configuration files")
		_ = err.WithSuggestion("Verify configuration syntax")
		_ = err.WithSuggestion("Ensure configuration is complete")
	case ErrorCodeAuthFailed, ErrorCodeAuthExpired, ErrorCodeAuthMissing, ErrorCodeAuthInvalid:
		_ = err.WithSuggestion("Check system authentication")
		_ = err.WithSuggestion("Verify credentials")
		_ = err.WithSuggestion("Ensure authentication services are available")
	default:
		// No specific suggestions for other error codes
	}

	return err
}

// Helper functions for common error scenarios

// WrapError wraps an existing error with additional context.
func WrapError(err error, code ErrorCode, message string) *GzhError {
	return &GzhError{
		Code:    code,
		Message: message,
		Cause:   err,
	}
}

// FromError converts a standard error to a GzhError.
func FromError(err error) *GzhError {
	if err == nil {
		return nil
	}

	gzhErr := &GzhError{}
	if errors.As(err, &gzhErr) {
		return gzhErr
	}

	return &GzhError{
		Code:    ErrorCodeSystemInternal,
		Message: "An internal error occurred",
		Details: err.Error(),
		Cause:   err,
	}
}

// IsCode checks if an error has a specific error code.
func IsCode(err error, code ErrorCode) bool {
	gzhErr := &GzhError{}
	if errors.As(err, &gzhErr) {
		return gzhErr.Code == code
	}

	return false
}

// ExtractCode extracts the error code from an error.
func ExtractCode(err error) ErrorCode {
	gzhErr := &GzhError{}
	if errors.As(err, &gzhErr) {
		return gzhErr.Code
	}

	return ErrorCodeSystemInternal
}

// Recovery functions for automatic error recovery

// RecoveryStrategy represents a strategy for automatic error recovery.
type RecoveryStrategy func(err *GzhError) error

// RecoveryManager manages automatic error recovery.
type RecoveryManager struct {
	strategies map[ErrorCode]RecoveryStrategy
}

// NewRecoveryManager creates a new recovery manager.
func NewRecoveryManager() *RecoveryManager {
	return &RecoveryManager{
		strategies: make(map[ErrorCode]RecoveryStrategy),
	}
}

// RegisterStrategy registers a recovery strategy for an error code.
func (rm *RecoveryManager) RegisterStrategy(code ErrorCode, strategy RecoveryStrategy) {
	rm.strategies[code] = strategy
}

// Recover attempts to recover from an error automatically.
func (rm *RecoveryManager) Recover(err error) error {
	gzhErr := FromError(err)
	if strategy, exists := rm.strategies[gzhErr.Code]; exists {
		return strategy(gzhErr)
	}

	return err
}

// Default recovery strategies

// NetworkRecoveryStrategy attempts to recover from network errors.
func NetworkRecoveryStrategy(err *GzhError) error {
	switch err.Code {
	case ErrorCodeNetworkTimeout:
		// Could implement retry logic here
		return fmt.Errorf("network timeout recovery not implemented yet")
	case ErrorCodeNetworkDNS:
		// Could implement alternative DNS resolution
		return fmt.Errorf("dns recovery not implemented yet")
	case ErrorCodeNetworkConnection:
		// Could implement connection retry with backoff
		return fmt.Errorf("network connection recovery not implemented yet")
	case ErrorCodeNetworkUnreachable:
		// Could implement route discovery
		return fmt.Errorf("network unreachable recovery not implemented yet")
	case ErrorCodeVPNConnection, ErrorCodeVPNAuthentication, ErrorCodeVPNConfiguration, ErrorCodeVPNHierarchy:
		// Network-related VPN issues
		return fmt.Errorf("vpn network recovery not implemented yet")
	case ErrorCodeConfigInvalid, ErrorCodeConfigNotFound, ErrorCodeConfigSyntax, ErrorCodeConfigValidation:
		// Network configuration issues
		return fmt.Errorf("network configuration recovery not implemented yet")
	case ErrorCodeAuthFailed, ErrorCodeAuthExpired, ErrorCodeAuthMissing, ErrorCodeAuthInvalid:
		// Network authentication issues
		return fmt.Errorf("network authentication recovery not implemented yet")
	case ErrorCodePermissionDenied, ErrorCodeResourceNotFound, ErrorCodeResourceExists:
		// Network resource issues
		return fmt.Errorf("network resource recovery not implemented yet")
	case ErrorCodeSystemInternal, ErrorCodeSystemTimeout, ErrorCodeSystemResource:
		// System-level network issues
		return fmt.Errorf("network system recovery not implemented yet")
	default:
		return err
	}
}

// VPNRecoveryStrategy attempts to recover from VPN errors.
func VPNRecoveryStrategy(err *GzhError) error {
	switch err.Code {
	case ErrorCodeVPNConnection:
		// Could implement failover to alternative VPN servers
		return fmt.Errorf("vpn connection recovery not implemented yet")
	case ErrorCodeVPNAuthentication:
		// Could implement credential refresh
		return fmt.Errorf("vpn authentication recovery not implemented yet")
	case ErrorCodeVPNConfiguration:
		// Could implement configuration auto-repair
		return fmt.Errorf("vpn configuration recovery not implemented yet")
	case ErrorCodeVPNHierarchy:
		// Could implement hierarchy repair
		return fmt.Errorf("vpn hierarchy recovery not implemented yet")
	case ErrorCodeNetworkConnection, ErrorCodeNetworkTimeout, ErrorCodeNetworkDNS, ErrorCodeNetworkUnreachable:
		// VPN-related network issues
		return fmt.Errorf("vpn network recovery not implemented yet")
	case ErrorCodeConfigInvalid, ErrorCodeConfigNotFound, ErrorCodeConfigSyntax, ErrorCodeConfigValidation:
		// VPN configuration issues
		return fmt.Errorf("vpn config recovery not implemented yet")
	case ErrorCodeAuthFailed, ErrorCodeAuthExpired, ErrorCodeAuthMissing, ErrorCodeAuthInvalid:
		// VPN authentication issues
		return fmt.Errorf("vpn auth recovery not implemented yet")
	case ErrorCodePermissionDenied, ErrorCodeResourceNotFound, ErrorCodeResourceExists:
		// VPN resource issues
		return fmt.Errorf("vpn resource recovery not implemented yet")
	case ErrorCodeSystemInternal, ErrorCodeSystemTimeout, ErrorCodeSystemResource:
		// System-level VPN issues
		return fmt.Errorf("vpn system recovery not implemented yet")
	default:
		return err
	}
}
