package errors_test

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gizzahub/gzh-manager-go/pkg/errors"
)

// ExampleGitHubTokenError demonstrates creating user-friendly GitHub authentication errors.
func ExampleGitHubTokenError() {
	// Simulate a GitHub API authentication failure
	originalErr := fmt.Errorf("401 Unauthorized: Bad credentials")

	// Create a user-friendly error
	userErr := errors.GitHubTokenError(originalErr)

	fmt.Println("User-friendly error message:")
	fmt.Printf("Title: %s\n", userErr.GetMessage())
	fmt.Printf("Description: %s\n", userErr.GetDescription())

	fmt.Println("\nSuggested solutions:")
	for _, suggestion := range userErr.GetSuggestions() {
		fmt.Printf("- %s\n", suggestion)
	}

	// Output: User-friendly GitHub authentication error with actionable guidance
}

// ExampleNetworkTimeoutError demonstrates creating user-friendly network timeout errors.
func ExampleNetworkTimeoutError() {
	operation := "repository clone"
	timeout := 30 * time.Second

	// Create a user-friendly timeout error
	userErr := errors.NetworkTimeoutError(operation, timeout)

	fmt.Printf("Operation: %s\n", operation)
	fmt.Printf("Timeout: %v\n", timeout)
	fmt.Printf("Error message: %s\n", userErr.GetMessage())
	fmt.Printf("Description: %s\n", userErr.GetDescription())

	fmt.Println("\nSuggested solutions:")
	for _, suggestion := range userErr.GetSuggestions() {
		fmt.Printf("- %s\n", suggestion)
	}

	// Output: User-friendly network timeout error with retry guidance
}

// ExampleUserErrorBuilder demonstrates building custom user-friendly errors.
func ExampleUserErrorBuilder() {
	// Build a custom error with detailed context
	userErr := errors.NewError(errors.DomainGit, errors.CategoryValidation, "INVALID_REPOSITORY_URL").
		Message("Invalid repository URL format").
		Description("The provided repository URL does not match expected Git URL patterns").
		Suggest("Ensure the URL follows the format: https://github.com/owner/repo.git").
		Suggest("Check for typos in the organization and repository names").
		Suggest("Verify the repository exists and is accessible").
		AddDetail("provided_url", "https://invalid-url").
		AddDetail("expected_format", "https://host/owner/repo.git").
		Build()

	fmt.Printf("Domain: %s\n", userErr.GetDomain())
	fmt.Printf("Category: %s\n", userErr.GetCategory())
	fmt.Printf("Code: %s\n", userErr.GetCode())
	fmt.Printf("Message: %s\n", userErr.GetMessage())
	fmt.Printf("Description: %s\n", userErr.GetDescription())

	fmt.Println("\nDetails:")
	for key, value := range userErr.GetDetails() {
		fmt.Printf("- %s: %v\n", key, value)
	}

	fmt.Println("\nSuggestions:")
	for _, suggestion := range userErr.GetSuggestions() {
		fmt.Printf("- %s\n", suggestion)
	}

	// Output: Custom user error with comprehensive context and guidance
}

// ExampleErrorAdapters demonstrates converting system errors to user-friendly errors.
func ExampleErrorAdapters() {
	// Simulate different types of system errors
	fmt.Println("=== OS Error Adaptation ===")
	osErr := &os.PathError{
		Op:   "open",
		Path: "/nonexistent/config.yaml",
		Err:  os.ErrNotExist,
	}

	userErr := errors.AdaptOSError(osErr, "/nonexistent/config.yaml", "configuration file access")
	fmt.Printf("Original: %v\n", osErr)
	fmt.Printf("User-friendly: %s\n", userErr.GetMessage())
	fmt.Printf("Description: %s\n", userErr.GetDescription())

	fmt.Println("\n=== Network Error Adaptation ===")
	networkErr := fmt.Errorf("dial tcp: lookup github.com: no such host")

	userNetErr := errors.AdaptNetworkError(networkErr, "repository clone", "github.com")
	fmt.Printf("Original: %v\n", networkErr)
	fmt.Printf("User-friendly: %s\n", userNetErr.GetMessage())
	fmt.Printf("Description: %s\n", userNetErr.GetDescription())

	// Output: System errors converted to user-friendly messages with context
}

// ExampleErrorContext demonstrates adding context to errors for better debugging.
func ExampleErrorContext() {
	// Create an error with rich context
	userErr := errors.NewError(errors.DomainGitLab, errors.CategoryAuth, "TOKEN_INSUFFICIENT_PERMISSIONS").
		Message("GitLab token has insufficient permissions").
		Description("The provided GitLab token does not have the required scopes to access private repositories").
		Suggest("Regenerate the token with 'read_repository' and 'read_user' scopes").
		Suggest("Check token permissions at https://gitlab.com/-/profile/personal_access_tokens").
		AddDetail("required_scopes", []string{"read_repository", "read_user"}).
		AddDetail("token_prefix", "glpat-").
		AddDetail("gitlab_instance", "https://gitlab.com").
		Build()

	// Demonstrate accessing error information
	fmt.Printf("Error Classification:\n")
	fmt.Printf("- Domain: %s\n", userErr.GetDomain())
	fmt.Printf("- Category: %s\n", userErr.GetCategory())
	fmt.Printf("- Code: %s\n", userErr.GetCode())

	fmt.Printf("\nUser Information:\n")
	fmt.Printf("- Message: %s\n", userErr.GetMessage())
	fmt.Printf("- Description: %s\n", userErr.GetDescription())

	fmt.Printf("\nContextual Details:\n")
	for key, value := range userErr.GetDetails() {
		fmt.Printf("- %s: %v\n", key, value)
	}

	fmt.Printf("\nRemediation Steps:\n")
	for i, suggestion := range userErr.GetSuggestions() {
		fmt.Printf("%d. %s\n", i+1, suggestion)
	}

	// Output: Comprehensive error context for effective troubleshooting
}

// ExampleErrorWorkflow demonstrates a complete error handling workflow
// in a typical application scenario.
func ExampleErrorWorkflow() {
	ctx := context.Background()

	// Simulate a complex operation that might fail at different stages
	fmt.Println("=== Repository Clone Workflow ===")

	// Stage 1: Configuration validation
	configErr := validateConfiguration()
	if configErr != nil {
		fmt.Printf("Configuration Error: %s\n", configErr.GetMessage())
		fmt.Printf("Resolution: %s\n", configErr.GetSuggestions()[0])
		fmt.Println()
	}

	// Stage 2: Authentication
	authErr := authenticateWithProvider(ctx)
	if authErr != nil {
		fmt.Printf("Authentication Error: %s\n", authErr.GetMessage())
		fmt.Printf("Resolution: %s\n", authErr.GetSuggestions()[0])
		fmt.Println()
	}

	// Stage 3: Network operation
	networkErr := performNetworkOperation(ctx)
	if networkErr != nil {
		fmt.Printf("Network Error: %s\n", networkErr.GetMessage())
		fmt.Printf("Resolution: %s\n", networkErr.GetSuggestions()[0])
		fmt.Println()
	}

	fmt.Println("Error workflow demonstrates progressive error handling")
	// Output: Complete error handling workflow with staged error management
}

// Helper functions for the workflow example
func validateConfiguration() *errors.UserError {
	return errors.NewError(errors.DomainConfig, errors.CategoryValidation, "MISSING_REQUIRED_FIELD").
		Message("Configuration validation failed").
		Description("Required field 'organizations' is missing from the configuration").
		Suggest("Add at least one organization to your bulk-clone.yaml configuration").
		AddDetail("missing_field", "organizations").
		Build()
}

func authenticateWithProvider(ctx context.Context) *errors.UserError {
	return errors.GitHubTokenError(fmt.Errorf("invalid token"))
}

func performNetworkOperation(ctx context.Context) *errors.UserError {
	return errors.NetworkTimeoutError("repository listing", 30*time.Second)
}

// ExampleErrorRecovery demonstrates error recovery and retry mechanisms.
func ExampleErrorRecovery() {
	maxRetries := 3
	currentAttempt := 1

	fmt.Printf("=== Error Recovery Example ===\n")

	for currentAttempt <= maxRetries {
		fmt.Printf("Attempt %d/%d: ", currentAttempt, maxRetries)

		// Simulate an operation that might fail
		if currentAttempt < 3 {
			// Create a recoverable error
			err := errors.NetworkTimeoutError("API request", 10*time.Second)
			fmt.Printf("Failed - %s\n", err.GetMessage())

			if currentAttempt < maxRetries {
				fmt.Printf("  Retrying in %d seconds...\n", currentAttempt*2)
				// In real code: time.Sleep(time.Duration(currentAttempt*2) * time.Second)
			}
		} else {
			fmt.Printf("Success!\n")
			break
		}

		currentAttempt++
	}

	if currentAttempt > maxRetries {
		finalErr := errors.NewError(errors.DomainNetwork, errors.CategoryTimeout, "MAX_RETRIES_EXCEEDED").
			Message("Operation failed after maximum retry attempts").
			Descriptionf("Failed to complete operation after %d attempts", maxRetries).
			Suggest("Check your network connection and try again later").
			Suggest("Contact support if the problem persists").
			AddDetail("max_retries", maxRetries).
			AddDetail("total_attempts", currentAttempt-1).
			Build()

		fmt.Printf("\nFinal Error: %s\n", finalErr.GetMessage())
	}

	// Output: Error recovery demonstrates retry logic with progressive backoff
}
