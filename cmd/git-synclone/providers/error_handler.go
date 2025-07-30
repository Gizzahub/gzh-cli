// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package providers

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// ProviderError represents provider-specific errors with additional context.
type ProviderError struct {
	Provider    string
	Operation   string
	StatusCode  int
	Message     string
	Details     map[string]interface{}
	Retryable   bool
	RetryAfter  time.Duration
	OriginalErr error
}

func (e *ProviderError) Error() string {
	if e.Details != nil && len(e.Details) > 0 {
		return fmt.Sprintf("%s %s failed: %s (details: %v)", e.Provider, e.Operation, e.Message, e.Details)
	}
	return fmt.Sprintf("%s %s failed: %s", e.Provider, e.Operation, e.Message)
}

func (e *ProviderError) Unwrap() error {
	return e.OriginalErr
}

// RateLimitError represents rate limit errors with reset information.
type RateLimitError struct {
	Provider   string
	ResetAt    time.Time
	Remaining  int
	Limit      int
	RetryAfter time.Duration
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limit exceeded for %s (limit: %d, remaining: %d, resets at: %s)",
		e.Provider, e.Limit, e.Remaining, e.ResetAt.Format("15:04:05"))
}

// AuthenticationError represents authentication failures.
type AuthenticationError struct {
	Provider string
	Message  string
	Hint     string
}

func (e *AuthenticationError) Error() string {
	if e.Hint != "" {
		return fmt.Sprintf("authentication failed for %s: %s (hint: %s)", e.Provider, e.Message, e.Hint)
	}
	return fmt.Sprintf("authentication failed for %s: %s", e.Provider, e.Message)
}

// NetworkError represents network connectivity issues.
type NetworkError struct {
	Provider  string
	Operation string
	URL       string
	Message   string
	Timeout   bool
}

func (e *NetworkError) Error() string {
	if e.Timeout {
		return fmt.Sprintf("network timeout for %s %s: %s", e.Provider, e.Operation, e.Message)
	}
	return fmt.Sprintf("network error for %s %s: %s", e.Provider, e.Operation, e.Message)
}

// ValidationError represents input validation failures.
type ValidationError struct {
	Field   string
	Value   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s' with value '%s': %s", e.Field, e.Value, e.Message)
}

// RepositoryError represents repository-specific errors.
type RepositoryError struct {
	Repository string
	Operation  string
	Message    string
	Permanent  bool
}

func (e *RepositoryError) Error() string {
	return fmt.Sprintf("repository error for '%s' during %s: %s", e.Repository, e.Operation, e.Message)
}

// ErrorHandler provides centralized error handling for git-synclone operations.
type ErrorHandler struct {
	formatter *OutputFormatter
}

// NewErrorHandler creates a new error handler.
func NewErrorHandler(formatter *OutputFormatter) *ErrorHandler {
	return &ErrorHandler{
		formatter: formatter,
	}
}

// HandleError processes and formats errors appropriately.
func (eh *ErrorHandler) HandleError(err error, context map[string]string) error {
	if err == nil {
		return nil
	}

	// Classify and handle different error types
	switch e := err.(type) {
	case *RateLimitError:
		return eh.handleRateLimitError(e, context)
	case *AuthenticationError:
		return eh.handleAuthenticationError(e, context)
	case *NetworkError:
		return eh.handleNetworkError(e, context)
	case *ValidationError:
		return eh.handleValidationError(e, context)
	case *RepositoryError:
		return eh.handleRepositoryError(e, context)
	case *ProviderError:
		return eh.handleProviderError(e, context)
	default:
		return eh.handleGenericError(err, context)
	}
}

// handleRateLimitError handles rate limit errors with retry suggestions.
func (eh *ErrorHandler) handleRateLimitError(err *RateLimitError, context map[string]string) error {
	message := fmt.Sprintf("Rate limit exceeded for %s", err.Provider)

	if err.RetryAfter > 0 {
		message += fmt.Sprintf("\nRetry after: %s", err.RetryAfter)
	} else if !err.ResetAt.IsZero() {
		untilReset := time.Until(err.ResetAt)
		if untilReset > 0 {
			message += fmt.Sprintf("\nRate limit resets in: %s", untilReset.Truncate(time.Second))
		}
	}

	if err.Remaining == 0 {
		message += "\nSuggestion: Wait for rate limit reset or use a different authentication token"
	}

	return eh.formatter.PrintError(fmt.Errorf("%s", message))
}

// handleAuthenticationError handles authentication failures with helpful hints.
func (eh *ErrorHandler) handleAuthenticationError(err *AuthenticationError, context map[string]string) error {
	message := fmt.Sprintf("Authentication failed for %s: %s", err.Provider, err.Message)

	suggestions := eh.getAuthenticationSuggestions(err.Provider)
	if len(suggestions) > 0 {
		message += "\n\nSuggestions:"
		for _, suggestion := range suggestions {
			message += fmt.Sprintf("\n  • %s", suggestion)
		}
	}

	if err.Hint != "" {
		message += fmt.Sprintf("\n\nHint: %s", err.Hint)
	}

	return eh.formatter.PrintError(fmt.Errorf("%s", message))
}

// handleNetworkError handles network connectivity issues.
func (eh *ErrorHandler) handleNetworkError(err *NetworkError, context map[string]string) error {
	message := fmt.Sprintf("Network error for %s", err.Provider)

	if err.Timeout {
		message += " (timeout)"
		message += "\nSuggestions:"
		message += "\n  • Check your internet connection"
		message += "\n  • Try increasing timeout settings"
		message += "\n  • Verify the service is not experiencing outages"
	} else {
		message += fmt.Sprintf(": %s", err.Message)
		message += "\nSuggestions:"
		message += "\n  • Check your internet connection"
		message += "\n  • Verify the service URL is correct"
		message += "\n  • Check if you're behind a proxy or firewall"
	}

	return eh.formatter.PrintError(fmt.Errorf("%s", message))
}

// handleValidationError handles input validation failures.
func (eh *ErrorHandler) handleValidationError(err *ValidationError, context map[string]string) error {
	message := fmt.Sprintf("Invalid input for %s: %s", err.Field, err.Message)

	// Provide field-specific suggestions
	suggestions := eh.getValidationSuggestions(err.Field, err.Value)
	if len(suggestions) > 0 {
		message += "\n\nSuggestions:"
		for _, suggestion := range suggestions {
			message += fmt.Sprintf("\n  • %s", suggestion)
		}
	}

	return eh.formatter.PrintError(fmt.Errorf("%s", message))
}

// handleRepositoryError handles repository-specific errors.
func (eh *ErrorHandler) handleRepositoryError(err *RepositoryError, context map[string]string) error {
	message := fmt.Sprintf("Repository error for '%s': %s", err.Repository, err.Message)

	if err.Permanent {
		message += "\nThis error is likely permanent and the repository will be skipped."
	} else {
		message += "\nThis error may be temporary. The operation will be retried."
	}

	return eh.formatter.PrintError(fmt.Errorf("%s", message))
}

// handleProviderError handles generic provider errors.
func (eh *ErrorHandler) handleProviderError(err *ProviderError, context map[string]string) error {
	message := fmt.Sprintf("%s error during %s", err.Provider, err.Operation)

	if err.StatusCode > 0 {
		message += fmt.Sprintf(" (HTTP %d)", err.StatusCode)
	}

	message += fmt.Sprintf(": %s", err.Message)

	if err.Retryable {
		message += "\nThis error is retryable."
		if err.RetryAfter > 0 {
			message += fmt.Sprintf(" Retry after: %s", err.RetryAfter)
		}
	}

	// Add HTTP status code specific suggestions
	if err.StatusCode > 0 {
		suggestions := eh.getHTTPStatusSuggestions(err.StatusCode)
		if len(suggestions) > 0 {
			message += "\n\nSuggestions:"
			for _, suggestion := range suggestions {
				message += fmt.Sprintf("\n  • %s", suggestion)
			}
		}
	}

	return eh.formatter.PrintError(fmt.Errorf("%s", message))
}

// handleGenericError handles unclassified errors.
func (eh *ErrorHandler) handleGenericError(err error, context map[string]string) error {
	message := fmt.Sprintf("Unexpected error: %s", err.Error())

	// Try to extract more context from the error message
	if strings.Contains(err.Error(), "timeout") {
		message += "\n\nThis appears to be a timeout error."
		message += "\nSuggestions:"
		message += "\n  • Check your internet connection"
		message += "\n  • Try reducing parallel workers (--parallel flag)"
		message += "\n  • Increase timeout settings if available"
	} else if strings.Contains(err.Error(), "permission") || strings.Contains(err.Error(), "forbidden") {
		message += "\n\nThis appears to be a permission error."
		message += "\nSuggestions:"
		message += "\n  • Verify your authentication token has the required permissions"
		message += "\n  • Check if the organization/repository exists and is accessible"
	}

	return eh.formatter.PrintError(fmt.Errorf("%s", message))
}

// Helper functions for providing contextual suggestions

func (eh *ErrorHandler) getAuthenticationSuggestions(provider string) []string {
	switch strings.ToLower(provider) {
	case "github":
		return []string{
			"Verify your GitHub token is set: export GITHUB_TOKEN=your_token",
			"Ensure the token has the required scopes (repo, read:org)",
			"Check if the token is still valid and not expired",
			"For organizations, ensure you have access to the organization",
		}
	case "gitlab":
		return []string{
			"Verify your GitLab token is set: export GITLAB_TOKEN=your_token",
			"Ensure the token has the required scopes (api, read_repository)",
			"Check if the token is still valid and not expired",
			"Verify the GitLab instance URL is correct",
		}
	case "gitea":
		return []string{
			"Verify your Gitea token is set: export GITEA_TOKEN=your_token",
			"Ensure the token has the required permissions",
			"Check if the token is still valid and not expired",
			"Verify the Gitea instance URL is correct",
		}
	default:
		return []string{
			"Verify your authentication token is set correctly",
			"Check if the token is still valid and has required permissions",
		}
	}
}

func (eh *ErrorHandler) getValidationSuggestions(field, value string) []string {
	switch field {
	case "organization", "org":
		return []string{
			"Organization name should not contain spaces or special characters",
			"Use the exact organization name as it appears on the platform",
			"For personal repositories, use your username",
		}
	case "strategy":
		return []string{
			"Valid strategies are: reset, pull, fetch",
			"reset: Hard reset and pull (discards local changes)",
			"pull: Merge remote changes with local changes",
			"fetch: Update remote tracking without changing working directory",
		}
	case "parallel":
		return []string{
			"Parallel workers must be a positive integer",
			"Recommended range: 1-20 workers",
			"Higher values may trigger rate limits",
		}
	case "protocol":
		return []string{
			"Valid protocols are: https, ssh",
			"Use https for token-based authentication",
			"Use ssh if you have SSH keys configured",
		}
	default:
		return []string{}
	}
}

func (eh *ErrorHandler) getHTTPStatusSuggestions(statusCode int) []string {
	switch statusCode {
	case http.StatusUnauthorized:
		return []string{
			"Check your authentication token",
			"Ensure the token has not expired",
			"Verify the token has the required permissions",
		}
	case http.StatusForbidden:
		return []string{
			"Your token does not have permission to access this resource",
			"Check if you have access to the organization/repository",
			"Verify the token scopes include the necessary permissions",
		}
	case http.StatusNotFound:
		return []string{
			"The organization or repository does not exist",
			"Check if the name is spelled correctly",
			"Verify you have access to view the resource",
		}
	case http.StatusTooManyRequests:
		return []string{
			"You have exceeded the rate limit",
			"Wait for the rate limit to reset",
			"Consider using fewer parallel workers",
			"Use a different authentication token if available",
		}
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
		return []string{
			"The service is experiencing issues",
			"Try again in a few minutes",
			"Check the service status page for known issues",
		}
	default:
		return []string{}
	}
}

// IsRetryable determines if an error should be retried.
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	switch e := err.(type) {
	case *RateLimitError:
		return true
	case *NetworkError:
		return e.Timeout || strings.Contains(e.Message, "connection")
	case *ProviderError:
		return e.Retryable || e.StatusCode >= 500
	case *RepositoryError:
		return !e.Permanent
	default:
		// Check for common retryable error patterns
		errMsg := strings.ToLower(err.Error())
		return strings.Contains(errMsg, "timeout") ||
			strings.Contains(errMsg, "connection") ||
			strings.Contains(errMsg, "temporary") ||
			strings.Contains(errMsg, "retry")
	}
}

// GetRetryDelay calculates the appropriate retry delay for an error.
func GetRetryDelay(err error, attempt int) time.Duration {
	baseDelay := time.Second
	maxDelay := 5 * time.Minute

	switch e := err.(type) {
	case *RateLimitError:
		if e.RetryAfter > 0 {
			return e.RetryAfter
		}
		if !e.ResetAt.IsZero() {
			untilReset := time.Until(e.ResetAt)
			if untilReset > 0 && untilReset < maxDelay {
				return untilReset
			}
		}
		return maxDelay
	case *ProviderError:
		if e.RetryAfter > 0 {
			return e.RetryAfter
		}
	}

	// Exponential backoff with jitter
	delay := baseDelay * time.Duration(1<<uint(attempt))
	if delay > maxDelay {
		delay = maxDelay
	}

	return delay
}

// WrapProviderError wraps an error with provider context.
func WrapProviderError(provider, operation string, err error) error {
	if err == nil {
		return nil
	}

	// Don't double-wrap provider errors
	if _, ok := err.(*ProviderError); ok {
		return err
	}

	return &ProviderError{
		Provider:    provider,
		Operation:   operation,
		Message:     err.Error(),
		OriginalErr: err,
		Retryable:   IsRetryable(err),
	}
}

// WrapRepositoryError wraps an error with repository context.
func WrapRepositoryError(repository, operation string, err error, permanent bool) error {
	if err == nil {
		return nil
	}

	return &RepositoryError{
		Repository: repository,
		Operation:  operation,
		Message:    err.Error(),
		Permanent:  permanent,
	}
}
