package legacy

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGzhError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *GzhError
		expected string
	}{
		{
			name: "error with details",
			err: &GzhError{
				Code:    ErrorCodeVPNConnection,
				Message: "Failed to connect to VPN",
				Details: "Connection timeout after 30 seconds",
			},
			expected: "[VPN_CONNECTION] Failed to connect to VPN: Connection timeout after 30 seconds",
		},
		{
			name: "error without details",
			err: &GzhError{
				Code:    ErrorCodeConfigNotFound,
				Message: "Configuration file not found",
			},
			expected: "[CONFIG_NOT_FOUND] Configuration file not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestGzhError_WithContext(t *testing.T) {
	err := &GzhError{
		Code:    ErrorCodeVPNConnection,
		Message: "VPN connection failed",
	}

	_ = err.WithContext("server", "vpn.example.com")
	_ = err.WithContext("port", 1194)

	assert.Len(t, err.Context, 2)
	assert.Equal(t, "vpn.example.com", err.Context["server"])
	assert.Equal(t, 1194, err.Context["port"])
}

func TestGzhError_WithSuggestion(t *testing.T) {
	err := &GzhError{
		Code:    ErrorCodeNetworkConnection,
		Message: "Network connection failed",
	}

	if retErr := err.WithSuggestion("Check your internet connection"); retErr != nil {
		t.Logf("Warning: failed to add suggestion: %v", retErr)
	}
	if retErr := err.WithSuggestion("Try using a different network"); retErr != nil {
		t.Logf("Warning: failed to add suggestion: %v", retErr)
	}

	assert.Len(t, err.Suggestions, 2)
	assert.Equal(t, "Check your internet connection", err.Suggestions[0])
	assert.Equal(t, "Try using a different network", err.Suggestions[1])
}

func TestGzhError_WithCause(t *testing.T) {
	originalErr := errors.New("original error")
	err := &GzhError{
		Code:    ErrorCodeSystemInternal,
		Message: "Internal error occurred",
	}

	_ = err.WithCause(originalErr)

	assert.Equal(t, originalErr, err.Cause)
	assert.Equal(t, originalErr, err.Unwrap())
}

func TestGzhError_Is(t *testing.T) {
	originalErr := errors.New("original error")
	err1 := &GzhError{
		Code:    ErrorCodeVPNConnection,
		Message: "VPN error",
		Cause:   originalErr,
	}

	err2 := &GzhError{
		Code: ErrorCodeVPNConnection,
	}

	err3 := &GzhError{
		Code: ErrorCodeNetworkConnection,
	}

	// Test matching error codes
	assert.True(t, err1.Is(err2))
	assert.False(t, err1.Is(err3))

	// Test matching underlying cause
	assert.True(t, err1.Is(originalErr))
}

func TestGzhError_GetFormattedError(t *testing.T) {
	err := &GzhError{
		Code:        ErrorCodeVPNConnection,
		Message:     "VPN connection failed",
		Details:     "Connection timeout after 30 seconds",
		Suggestions: []string{"Check internet connection", "Try different server"},
		Context: map[string]interface{}{
			"server": "vpn.example.com",
			"port":   1194,
		},
	}

	formatted := err.GetFormattedError()

	assert.Contains(t, formatted, "❌ Error: VPN connection failed")
	assert.Contains(t, formatted, "📝 Details: Connection timeout after 30 seconds")
	assert.Contains(t, formatted, "🔍 Context:")
	assert.Contains(t, formatted, "server: vpn.example.com")
	assert.Contains(t, formatted, "port: 1194")
	assert.Contains(t, formatted, "💡 Suggestions:")
	assert.Contains(t, formatted, "1. Check internet connection")
	assert.Contains(t, formatted, "2. Try different server")
}

func TestNewVPNError(t *testing.T) {
	tests := []struct {
		name                string
		code                ErrorCode
		message             string
		expectedSuggestions int
	}{
		{
			name:                "VPN connection error",
			code:                ErrorCodeVPNConnection,
			message:             "Failed to connect",
			expectedSuggestions: 3,
		},
		{
			name:                "VPN authentication error",
			code:                ErrorCodeVPNAuthentication,
			message:             "Authentication failed",
			expectedSuggestions: 3,
		},
		{
			name:                "VPN configuration error",
			code:                ErrorCodeVPNConfiguration,
			message:             "Invalid configuration",
			expectedSuggestions: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewVPNError(tt.code, tt.message)

			assert.Equal(t, tt.code, err.Code)
			assert.Equal(t, tt.message, err.Message)
			assert.Len(t, err.Suggestions, tt.expectedSuggestions)
		})
	}
}

func TestNewConfigError(t *testing.T) {
	tests := []struct {
		name                string
		code                ErrorCode
		message             string
		expectedSuggestions int
	}{
		{
			name:                "Config not found",
			code:                ErrorCodeConfigNotFound,
			message:             "Configuration file not found",
			expectedSuggestions: 3,
		},
		{
			name:                "Config invalid",
			code:                ErrorCodeConfigInvalid,
			message:             "Invalid configuration",
			expectedSuggestions: 3,
		},
		{
			name:                "Config syntax error",
			code:                ErrorCodeConfigSyntax,
			message:             "Syntax error in configuration",
			expectedSuggestions: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewConfigError(tt.code, tt.message)

			assert.Equal(t, tt.code, err.Code)
			assert.Equal(t, tt.message, err.Message)
			assert.Len(t, err.Suggestions, tt.expectedSuggestions)
		})
	}
}

func TestNewAuthError(t *testing.T) {
	tests := []struct {
		name                string
		code                ErrorCode
		message             string
		expectedSuggestions int
	}{
		{
			name:                "Auth missing",
			code:                ErrorCodeAuthMissing,
			message:             "Authentication required",
			expectedSuggestions: 3,
		},
		{
			name:                "Auth expired",
			code:                ErrorCodeAuthExpired,
			message:             "Authentication expired",
			expectedSuggestions: 3,
		},
		{
			name:                "Auth failed",
			code:                ErrorCodeAuthFailed,
			message:             "Authentication failed",
			expectedSuggestions: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewAuthError(tt.code, tt.message)

			assert.Equal(t, tt.code, err.Code)
			assert.Equal(t, tt.message, err.Message)
			assert.Len(t, err.Suggestions, tt.expectedSuggestions)
		})
	}
}

func TestNewSystemError(t *testing.T) {
	tests := []struct {
		name                string
		code                ErrorCode
		message             string
		expectedSuggestions int
	}{
		{
			name:                "System resource error",
			code:                ErrorCodeSystemResource,
			message:             "Insufficient resources",
			expectedSuggestions: 3,
		},
		{
			name:                "System timeout",
			code:                ErrorCodeSystemTimeout,
			message:             "Operation timed out",
			expectedSuggestions: 3,
		},
		{
			name:                "Permission denied",
			code:                ErrorCodePermissionDenied,
			message:             "Permission denied",
			expectedSuggestions: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewSystemError(tt.code, tt.message)

			assert.Equal(t, tt.code, err.Code)
			assert.Equal(t, tt.message, err.Message)
			assert.Len(t, err.Suggestions, tt.expectedSuggestions)
		})
	}
}

func TestWrapError(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := WrapError(originalErr, ErrorCodeVPNConnection, "VPN connection failed")

	assert.Equal(t, ErrorCodeVPNConnection, wrappedErr.Code)
	assert.Equal(t, "VPN connection failed", wrappedErr.Message)
	assert.Equal(t, originalErr, wrappedErr.Cause)
	assert.Equal(t, originalErr, wrappedErr.Unwrap())
}

func TestFromError(t *testing.T) {
	tests := []struct {
		name     string
		input    error
		expected *GzhError
	}{
		{
			name:     "nil error",
			input:    nil,
			expected: nil,
		},
		{
			name: "already GzhError",
			input: &GzhError{
				Code:    ErrorCodeVPNConnection,
				Message: "VPN error",
			},
			expected: &GzhError{
				Code:    ErrorCodeVPNConnection,
				Message: "VPN error",
			},
		},
		{
			name:  "standard error",
			input: errors.New("standard error"),
			expected: &GzhError{
				Code:    ErrorCodeSystemInternal,
				Message: "An internal error occurred",
				Details: "standard error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FromError(tt.input)
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.Equal(t, tt.expected.Code, result.Code)
				assert.Equal(t, tt.expected.Message, result.Message)

				if tt.expected.Details != "" {
					assert.Equal(t, tt.expected.Details, result.Details)
				}
			}
		})
	}
}

func TestIsCode(t *testing.T) {
	gzhErr := &GzhError{Code: ErrorCodeVPNConnection}
	standardErr := errors.New("standard error")

	assert.True(t, IsCode(gzhErr, ErrorCodeVPNConnection))
	assert.False(t, IsCode(gzhErr, ErrorCodeNetworkConnection))
	assert.False(t, IsCode(standardErr, ErrorCodeVPNConnection))
}

func TestExtractCode(t *testing.T) {
	gzhErr := &GzhError{Code: ErrorCodeVPNConnection}
	standardErr := errors.New("standard error")

	assert.Equal(t, ErrorCodeVPNConnection, ExtractCode(gzhErr))
	assert.Equal(t, ErrorCodeSystemInternal, ExtractCode(standardErr))
}

func TestRecoveryManager(t *testing.T) {
	rm := NewRecoveryManager()

	// Register a mock recovery strategy
	callCount := 0
	mockStrategy := func(_ *GzhError) error {
		callCount++
		return nil
	}

	rm.RegisterStrategy(ErrorCodeVPNConnection, mockStrategy)

	// Test recovery for registered error code
	err := &GzhError{Code: ErrorCodeVPNConnection}
	result := rm.Recover(err)

	assert.NoError(t, result)
	assert.Equal(t, 1, callCount)

	// Test recovery for unregistered error code
	err2 := &GzhError{Code: ErrorCodeNetworkConnection}
	result2 := rm.Recover(err2)

	assert.Equal(t, err2, result2)
	assert.Equal(t, 1, callCount) // Should not increment
}

func TestGetSuggestions(t *testing.T) {
	tests := []struct {
		name        string
		suggestions []string
		expected    string
	}{
		{
			name:        "no suggestions",
			suggestions: []string{},
			expected:    "",
		},
		{
			name:        "single suggestion",
			suggestions: []string{"Check internet connection"},
			expected:    "Suggestions:\n  1. Check internet connection\n",
		},
		{
			name:        "multiple suggestions",
			suggestions: []string{"Check internet connection", "Try different server"},
			expected:    "Suggestions:\n  1. Check internet connection\n  2. Try different server\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &GzhError{
				Code:        ErrorCodeVPNConnection,
				Message:     "Test error",
				Suggestions: tt.suggestions,
			}

			result := err.GetSuggestions()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestErrorCodes(t *testing.T) {
	// Test that all error codes are defined and unique
	codes := []ErrorCode{
		ErrorCodeNetworkConnection,
		ErrorCodeNetworkTimeout,
		ErrorCodeNetworkDNS,
		ErrorCodeNetworkUnreachable,
		ErrorCodeVPNConnection,
		ErrorCodeVPNAuthentication,
		ErrorCodeVPNConfiguration,
		ErrorCodeVPNHierarchy,
		ErrorCodeConfigInvalid,
		ErrorCodeConfigNotFound,
		ErrorCodeConfigSyntax,
		ErrorCodeConfigValidation,
		ErrorCodeAuthFailed,
		ErrorCodeAuthExpired,
		ErrorCodeAuthMissing,
		ErrorCodeAuthInvalid,
		ErrorCodePermissionDenied,
		ErrorCodeResourceNotFound,
		ErrorCodeResourceExists,
		ErrorCodeSystemInternal,
		ErrorCodeSystemTimeout,
		ErrorCodeSystemResource,
	}

	// Check that no error code is empty
	for _, code := range codes {
		assert.NotEmpty(t, string(code))
	}

	// Check for uniqueness
	seen := make(map[ErrorCode]bool)
	for _, code := range codes {
		assert.False(t, seen[code], "Duplicate error code: %s", code)
		seen[code] = true
	}
}

func TestErrorChaining(t *testing.T) {
	// Test error chaining with multiple levels
	originalErr := errors.New("original network error")

	level1 := WrapError(originalErr, ErrorCodeNetworkConnection, "Network connection failed")
	level2 := WrapError(level1, ErrorCodeVPNConnection, "VPN connection failed due to network issues")

	// Test unwrapping
	assert.Equal(t, level1, level2.Unwrap())
	assert.Equal(t, originalErr, level1.Unwrap())

	// Test error chain matching
	assert.True(t, errors.Is(level2, level1))
	assert.True(t, errors.Is(level2, originalErr))
	assert.True(t, errors.Is(level1, originalErr))
}
