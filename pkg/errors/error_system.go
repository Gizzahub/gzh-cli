package errors

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"time"
)

// ErrorCode represents a structured error code with domain and category
type ErrorCode struct {
	Domain   string `json:"domain"`   // e.g., "git", "github", "config", "network"
	Category string `json:"category"` // e.g., "validation", "network", "auth", "timeout"
	Code     string `json:"code"`     // e.g., "INVALID_TOKEN", "REPO_NOT_FOUND"
}

// String returns the full error code in format: DOMAIN_CATEGORY_CODE
func (ec ErrorCode) String() string {
	return fmt.Sprintf("%s_%s_%s",
		strings.ToUpper(ec.Domain),
		strings.ToUpper(ec.Category),
		strings.ToUpper(ec.Code))
}

// UserError represents a user-friendly error with context
type UserError struct {
	Code        ErrorCode              `json:"code"`
	Message     string                 `json:"message"`
	Description string                 `json:"description"`
	Suggestions []string               `json:"suggestions"`
	Context     map[string]interface{} `json:"context"`
	Timestamp   time.Time              `json:"timestamp"`
	RequestID   string                 `json:"request_id,omitempty"`
	StackTrace  []string               `json:"stack_trace,omitempty"`
	Cause       error                  `json:"-"` // Original error
	i18nKey     string                 // For localization
}

// Error implements the error interface
func (e *UserError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("[%s] %s", e.Code.String(), e.Description)
}

// Unwrap returns the underlying error
func (e *UserError) Unwrap() error {
	return e.Cause
}

// WithContext adds context information to the error
func (e *UserError) WithContext(key string, value interface{}) *UserError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithRequestID adds a request ID for tracing
func (e *UserError) WithRequestID(requestID string) *UserError {
	e.RequestID = requestID
	return e
}

// WithSuggestion adds a helpful suggestion
func (e *UserError) WithSuggestion(suggestion string) *UserError {
	e.Suggestions = append(e.Suggestions, suggestion)
	return e
}

// WithStackTrace captures the current stack trace
func (e *UserError) WithStackTrace() *UserError {
	if e.StackTrace == nil {
		e.StackTrace = captureStackTrace()
	}
	return e
}

// JSON returns the error as JSON string
func (e *UserError) JSON() string {
	data, _ := json.MarshalIndent(e, "", "  ")
	return string(data)
}

// ErrorBuilder helps construct user-friendly errors
type ErrorBuilder struct {
	code        ErrorCode
	message     string
	description string
	suggestions []string
	context     map[string]interface{}
	requestID   string
	cause       error
	i18nKey     string
}

// NewError creates a new error builder
func NewError(domain, category, code string) *ErrorBuilder {
	return &ErrorBuilder{
		code: ErrorCode{
			Domain:   domain,
			Category: category,
			Code:     code,
		},
		context: make(map[string]interface{}),
	}
}

// Message sets the primary error message
func (b *ErrorBuilder) Message(msg string) *ErrorBuilder {
	b.message = msg
	return b
}

// Messagef sets the primary error message with formatting
func (b *ErrorBuilder) Messagef(format string, args ...interface{}) *ErrorBuilder {
	b.message = fmt.Sprintf(format, args...)
	return b
}

// Description sets the detailed error description
func (b *ErrorBuilder) Description(desc string) *ErrorBuilder {
	b.description = desc
	return b
}

// Descriptionf sets the detailed error description with formatting
func (b *ErrorBuilder) Descriptionf(format string, args ...interface{}) *ErrorBuilder {
	b.description = fmt.Sprintf(format, args...)
	return b
}

// Suggest adds a helpful suggestion
func (b *ErrorBuilder) Suggest(suggestion string) *ErrorBuilder {
	b.suggestions = append(b.suggestions, suggestion)
	return b
}

// Context adds context information
func (b *ErrorBuilder) Context(key string, value interface{}) *ErrorBuilder {
	if b.context == nil {
		b.context = make(map[string]interface{})
	}
	b.context[key] = value
	return b
}

// RequestID sets the request ID for tracing
func (b *ErrorBuilder) RequestID(id string) *ErrorBuilder {
	b.requestID = id
	return b
}

// Cause sets the underlying error
func (b *ErrorBuilder) Cause(err error) *ErrorBuilder {
	b.cause = err
	return b
}

// I18nKey sets the internationalization key
func (b *ErrorBuilder) I18nKey(key string) *ErrorBuilder {
	b.i18nKey = key
	return b
}

// Build creates the final UserError
func (b *ErrorBuilder) Build() *UserError {
	err := &UserError{
		Code:        b.code,
		Message:     b.message,
		Description: b.description,
		Suggestions: b.suggestions,
		Context:     b.context,
		Timestamp:   time.Now(),
		RequestID:   b.requestID,
		Cause:       b.cause,
		i18nKey:     b.i18nKey,
	}

	// Capture stack trace
	err.StackTrace = captureStackTrace()

	return err
}

// Predefined error domains and categories
const (
	// Domains
	DomainGit     = "git"
	DomainGitHub  = "github"
	DomainGitLab  = "gitlab"
	DomainGitea   = "gitea"
	DomainConfig  = "config"
	DomainNetwork = "network"
	DomainAuth    = "auth"
	DomainCLI     = "cli"
	DomainFile    = "file"
	DomainAPI     = "api"

	// Categories
	CategoryValidation = "validation"
	CategoryNetwork    = "network"
	CategoryAuth       = "auth"
	CategoryTimeout    = "timeout"
	CategoryNotFound   = "not_found"
	CategoryPermission = "permission"
	CategoryFormat     = "format"
	CategoryState      = "state"
	CategoryResource   = "resource"
	CategoryConfig     = "config"
)

// Common error constructors
func ConfigValidationError(field string, value interface{}) *UserError {
	return NewError(DomainConfig, CategoryValidation, "INVALID_FIELD").
		Message("Configuration validation failed").
		Descriptionf("Invalid value for field '%s': %v", field, value).
		Context("field", field).
		Context("value", value).
		Suggest("Check the configuration documentation for valid values").
		Suggest("Use 'gz config validate' to verify your configuration").
		Build()
}

func GitHubTokenError(err error) *UserError {
	return NewError(DomainGitHub, CategoryAuth, "INVALID_TOKEN").
		Message("GitHub authentication failed").
		Description("The provided GitHub token is invalid or has insufficient permissions").
		Cause(err).
		Suggest("Check your GitHub token in GITHUB_TOKEN environment variable").
		Suggest("Ensure the token has the required permissions (repo, admin:org)").
		Suggest("Generate a new token at https://github.com/settings/tokens").
		Build()
}

func NetworkTimeoutError(operation string, duration time.Duration) *UserError {
	return NewError(DomainNetwork, CategoryTimeout, "OPERATION_TIMEOUT").
		Message("Network operation timed out").
		Descriptionf("The %s operation timed out after %v", operation, duration).
		Context("operation", operation).
		Context("timeout", duration.String()).
		Suggest("Check your internet connection").
		Suggest("Try increasing the timeout value").
		Suggest("Check if the remote service is available").
		Build()
}

func RepositoryNotFoundError(repo string, provider string) *UserError {
	return NewError(provider, CategoryNotFound, "REPO_NOT_FOUND").
		Message("Repository not found").
		Descriptionf("The repository '%s' was not found on %s", repo, provider).
		Context("repository", repo).
		Context("provider", provider).
		Suggest("Check the repository name and owner").
		Suggest("Ensure you have access to the repository").
		Suggest("Verify the repository exists and is not private").
		Build()
}

func FilePermissionError(path string, operation string) *UserError {
	return NewError(DomainFile, CategoryPermission, "ACCESS_DENIED").
		Message("File permission denied").
		Descriptionf("Permission denied when trying to %s file: %s", operation, path).
		Context("path", path).
		Context("operation", operation).
		Suggest("Check file permissions and ownership").
		Suggest("Run with appropriate privileges if needed").
		Suggest("Ensure the directory exists and is writable").
		Build()
}

func APIRateLimitError(provider string, resetTime time.Time) *UserError {
	return NewError(provider, CategoryResource, "RATE_LIMIT_EXCEEDED").
		Message("API rate limit exceeded").
		Descriptionf("You have exceeded the API rate limit for %s", provider).
		Context("provider", provider).
		Context("reset_time", resetTime.Format(time.RFC3339)).
		Suggest("Wait until the rate limit resets").
		Suggest("Use a token with higher rate limits").
		Suggest("Implement request batching or caching").
		Build()
}

// Helper functions
func captureStackTrace() []string {
	var stack []string
	pc := make([]uintptr, 10)
	n := runtime.Callers(3, pc) // Skip captureStackTrace, WithStackTrace/Build, and caller

	for i := 0; i < n; i++ {
		fn := runtime.FuncForPC(pc[i])
		if fn != nil {
			file, line := fn.FileLine(pc[i])
			stack = append(stack, fmt.Sprintf("%s:%d %s", file, line, fn.Name()))
		}
	}

	return stack
}

// GetRequestIDFromContext extracts request ID from context
func GetRequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	if id := ctx.Value("request_id"); id != nil {
		if requestID, ok := id.(string); ok {
			return requestID
		}
	}

	if id := ctx.Value("requestId"); id != nil {
		if requestID, ok := id.(string); ok {
			return requestID
		}
	}

	return ""
}

// Is checks if the error is of a specific type
func Is(err error, code ErrorCode) bool {
	var userErr *UserError
	if As(err, &userErr) {
		return userErr.Code == code
	}
	return false
}

// As finds the first error in err's chain that matches target
func As(err error, target interface{}) bool {
	if target == nil {
		panic("errors: target cannot be nil")
	}

	for err != nil {
		if userErr, ok := err.(*UserError); ok {
			if targetPtr, ok := target.(**UserError); ok {
				*targetPtr = userErr
				return true
			}
		}

		// Check if error implements Unwrap
		if unwrapper, ok := err.(interface{ Unwrap() error }); ok {
			err = unwrapper.Unwrap()
		} else {
			break
		}
	}

	return false
}

// Wrap wraps an existing error with user-friendly information
func Wrap(err error, domain, category, code string) *ErrorBuilder {
	return NewError(domain, category, code).Cause(err)
}

// WrapWithMessage wraps an error with a message
func WrapWithMessage(err error, message string) error {
	return &UserError{
		Message:   message,
		Cause:     err,
		Timestamp: time.Now(),
	}
}
