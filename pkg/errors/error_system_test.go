package errors

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorCode(t *testing.T) {
	t.Run("String formatting", func(t *testing.T) {
		code := ErrorCode{
			Domain:   "github",
			Category: "auth",
			Code:     "invalid_token",
		}

		expected := "GITHUB_AUTH_INVALID_TOKEN"
		assert.Equal(t, expected, code.String())
	})
}

func TestUserError(t *testing.T) {
	t.Run("Basic error creation", func(t *testing.T) {
		err := NewError(DomainGitHub, CategoryAuth, "INVALID_TOKEN").
			Message("Authentication failed").
			Description("The provided token is invalid").
			Suggest("Check your token").
			Context("token", "ghp_***").
			Build()

		assert.Equal(t, "Authentication failed", err.Error())
		assert.Equal(t, "GITHUB_AUTH_INVALID_TOKEN", err.Code.String())
		assert.Len(t, err.Suggestions, 1)
		assert.Equal(t, "ghp_***", err.Context["token"])
		assert.WithinDuration(t, time.Now(), err.Timestamp, time.Second)
	})

	t.Run("Error with cause", func(t *testing.T) {
		originalErr := fmt.Errorf("network timeout")

		err := NewError(DomainNetwork, CategoryTimeout, "TIMEOUT").
			Message("Operation timed out").
			Cause(originalErr).
			Build()

		assert.Equal(t, originalErr, err.Unwrap())
	})

	t.Run("Error with request ID", func(t *testing.T) {
		requestID := "req-123"

		err := NewError(DomainAPI, CategoryValidation, "INVALID_INPUT").
			Message("Invalid input").
			RequestID(requestID).
			Build()

		assert.Equal(t, requestID, err.RequestID)
	})

	t.Run("Error with stack trace", func(t *testing.T) {
		err := NewError(DomainFile, CategoryPermission, "ACCESS_DENIED").
			Message("Permission denied").
			Build().
			WithStackTrace()

		assert.NotEmpty(t, err.StackTrace)
		assert.Contains(t, err.StackTrace[0], "error_system_test.go")
	})

	t.Run("JSON serialization", func(t *testing.T) {
		err := NewError(DomainConfig, CategoryValidation, "INVALID_FIELD").
			Message("Invalid config").
			Description("Field validation failed").
			Context("field", "database.host").
			Context("value", "").
			Build()

		jsonStr := err.JSON()
		assert.Contains(t, jsonStr, "Invalid config")
		assert.Contains(t, jsonStr, "INVALID_FIELD")

		// Test deserialization
		var restored UserError
		require.NoError(t, json.Unmarshal([]byte(jsonStr), &restored))
		assert.Equal(t, err.Message, restored.Message)
		assert.Equal(t, err.Code.String(), restored.Code.String())
	})
}

func TestPredefinedErrors(t *testing.T) {
	t.Run("ConfigValidationError", func(t *testing.T) {
		err := ConfigValidationError("database.port", "invalid")

		assert.Equal(t, DomainConfig, err.Code.Domain)
		assert.Equal(t, CategoryValidation, err.Code.Category)
		assert.Contains(t, err.Description, "database.port")
		assert.Contains(t, err.Description, "invalid")
		assert.NotEmpty(t, err.Suggestions)
	})

	t.Run("GitHubTokenError", func(t *testing.T) {
		originalErr := fmt.Errorf("401 Unauthorized")
		err := GitHubTokenError(originalErr)

		assert.Equal(t, DomainGitHub, err.Code.Domain)
		assert.Equal(t, CategoryAuth, err.Code.Category)
		assert.Equal(t, originalErr, err.Unwrap())
		assert.Contains(t, err.Suggestions[0], "GITHUB_TOKEN")
	})

	t.Run("NetworkTimeoutError", func(t *testing.T) {
		duration := 30 * time.Second
		err := NetworkTimeoutError("clone", duration)

		assert.Equal(t, DomainNetwork, err.Code.Domain)
		assert.Equal(t, CategoryTimeout, err.Code.Category)
		assert.Contains(t, err.Description, "clone")
		assert.Contains(t, err.Description, "30s")
		assert.Equal(t, "clone", err.Context["operation"])
	})

	t.Run("RepositoryNotFoundError", func(t *testing.T) {
		err := RepositoryNotFoundError("owner/repo", "github")

		assert.Equal(t, "github", err.Code.Domain)
		assert.Equal(t, CategoryNotFound, err.Code.Category)
		assert.Contains(t, err.Description, "owner/repo")
		assert.Equal(t, "owner/repo", err.Context["repository"])
	})

	t.Run("FilePermissionError", func(t *testing.T) {
		err := FilePermissionError("/etc/config", "write")

		assert.Equal(t, DomainFile, err.Code.Domain)
		assert.Equal(t, CategoryPermission, err.Code.Category)
		assert.Contains(t, err.Description, "/etc/config")
		assert.Contains(t, err.Description, "write")
	})

	t.Run("APIRateLimitError", func(t *testing.T) {
		resetTime := time.Now().Add(time.Hour)
		err := APIRateLimitError("github", resetTime)

		assert.Equal(t, "github", err.Code.Domain)
		assert.Equal(t, CategoryResource, err.Code.Category)
		assert.Contains(t, err.Description, "github")
		assert.NotEmpty(t, err.Context["reset_time"])
	})
}

func TestErrorHelpers(t *testing.T) {
	t.Run("Is function", func(t *testing.T) {
		code := ErrorCode{Domain: DomainGitHub, Category: CategoryAuth, Code: "INVALID_TOKEN"}
		err := NewError(code.Domain, code.Category, code.Code).
			Message("Auth failed").
			Build()

		assert.True(t, Is(err, code))

		otherCode := ErrorCode{Domain: DomainConfig, Category: CategoryValidation, Code: "INVALID_FIELD"}
		assert.False(t, Is(err, otherCode))

		// Test with non-UserError
		regularErr := fmt.Errorf("regular error")
		assert.False(t, Is(regularErr, code))
	})

	t.Run("As function", func(t *testing.T) {
		originalErr := NewError(DomainNetwork, CategoryTimeout, "TIMEOUT").
			Message("Timeout").
			Build()

		var userErr *UserError
		assert.True(t, As(originalErr, &userErr))
		assert.Equal(t, "Timeout", userErr.Message)

		// Test with wrapped error
		wrappedErr := fmt.Errorf("wrapped: %w", originalErr)
		var userErr2 *UserError
		assert.True(t, As(wrappedErr, &userErr2))
		assert.Equal(t, "Timeout", userErr2.Message)

		// Test with non-UserError
		regularErr := fmt.Errorf("regular error")
		var userErr3 *UserError
		assert.False(t, As(regularErr, &userErr3))
	})

	t.Run("Wrap function", func(t *testing.T) {
		originalErr := fmt.Errorf("original error")

		wrappedErr := Wrap(originalErr, DomainFile, CategoryPermission, "ACCESS_DENIED").
			Message("Wrapped error").
			Build()

		assert.Equal(t, "Wrapped error", wrappedErr.Error())
		assert.Equal(t, originalErr, wrappedErr.Unwrap())
	})

	t.Run("WrapWithMessage function", func(t *testing.T) {
		originalErr := fmt.Errorf("original error")
		wrappedErr := WrapWithMessage(originalErr, "Operation failed")

		assert.Equal(t, "Operation failed", wrappedErr.Error())

		var userErr *UserError
		require.True(t, As(wrappedErr, &userErr))
		assert.Equal(t, originalErr, userErr.Unwrap())
	})
}

func TestGetRequestIDFromContext(t *testing.T) {
	t.Run("With request_id key", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "request_id", "req-123")
		requestID := GetRequestIDFromContext(ctx)
		assert.Equal(t, "req-123", requestID)
	})

	t.Run("With requestId key", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "requestId", "req-456")
		requestID := GetRequestIDFromContext(ctx)
		assert.Equal(t, "req-456", requestID)
	})

	t.Run("With nil context", func(t *testing.T) {
		requestID := GetRequestIDFromContext(nil)
		assert.Empty(t, requestID)
	})

	t.Run("With no request ID", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "other_key", "value")
		requestID := GetRequestIDFromContext(ctx)
		assert.Empty(t, requestID)
	})
}

func TestI18nManager(t *testing.T) {
	t.Run("Basic message retrieval", func(t *testing.T) {
		manager := NewI18nManager(LocaleEnglish)

		catalog := MessageCatalog{
			"test.message": {
				Text:        "Test message",
				Description: "Test description",
				Suggestions: []string{"Test suggestion"},
			},
		}

		manager.AddCatalog(LocaleEnglish, catalog)

		msg, found := manager.GetMessage(LocaleEnglish, "test.message")
		assert.True(t, found)
		assert.Equal(t, "Test message", msg.Text)
		assert.Equal(t, "Test description", msg.Description)
		assert.Len(t, msg.Suggestions, 1)
	})

	t.Run("Fallback to default locale", func(t *testing.T) {
		manager := NewI18nManager(LocaleEnglish)

		englishCatalog := MessageCatalog{
			"test.message": {Text: "English message"},
		}

		manager.AddCatalog(LocaleEnglish, englishCatalog)

		// Request Korean but get English fallback
		msg, found := manager.GetMessage(LocaleKorean, "test.message")
		assert.True(t, found)
		assert.Equal(t, "English message", msg.Text)
	})

	t.Run("Message not found", func(t *testing.T) {
		manager := NewI18nManager(LocaleEnglish)

		_, found := manager.GetMessage(LocaleEnglish, "nonexistent.message")
		assert.False(t, found)
	})

	t.Run("Load from file", func(t *testing.T) {
		manager := NewI18nManager(LocaleEnglish)

		// Create temporary file
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "en.json")

		catalog := MessageCatalog{
			"file.test": {Text: "From file"},
		}

		data, err := json.Marshal(catalog)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(filePath, data, 0o644))

		// Load from file
		err = manager.LoadFromFile(LocaleEnglish, filePath)
		require.NoError(t, err)

		msg, found := manager.GetMessage(LocaleEnglish, "file.test")
		assert.True(t, found)
		assert.Equal(t, "From file", msg.Text)
	})

	t.Run("Load from directory", func(t *testing.T) {
		manager := NewI18nManager(LocaleEnglish)

		tempDir := t.TempDir()

		// Create English file
		enCatalog := MessageCatalog{"en.test": {Text: "English"}}
		enData, _ := json.Marshal(enCatalog)
		require.NoError(t, os.WriteFile(filepath.Join(tempDir, "en.json"), enData, 0o644))

		// Create Korean file
		koCatalog := MessageCatalog{"ko.test": {Text: "한국어"}}
		koData, _ := json.Marshal(koCatalog)
		require.NoError(t, os.WriteFile(filepath.Join(tempDir, "ko.json"), koData, 0o644))

		// Load directory
		err := manager.LoadFromDirectory(tempDir)
		require.NoError(t, err)

		// Test English
		msg, found := manager.GetMessage(LocaleEnglish, "en.test")
		assert.True(t, found)
		assert.Equal(t, "English", msg.Text)

		// Test Korean
		msg, found = manager.GetMessage(LocaleKorean, "ko.test")
		assert.True(t, found)
		assert.Equal(t, "한국어", msg.Text)
	})

	t.Run("Interpolate message", func(t *testing.T) {
		manager := NewI18nManager(LocaleEnglish)

		context := map[string]interface{}{
			"name":  "John",
			"count": 42,
		}

		result := manager.interpolateMessage("Hello {{name}}, you have {{count}} items", context)
		assert.Equal(t, "Hello John, you have 42 items", result)
	})

	t.Run("Get locale from context", func(t *testing.T) {
		manager := NewI18nManager(LocaleEnglish)

		// Test with locale in context
		ctx := context.WithValue(context.Background(), "locale", LocaleKorean)
		locale := manager.GetLocaleFromContext(ctx)
		assert.Equal(t, LocaleKorean, locale)

		// Test with string locale
		ctx = context.WithValue(context.Background(), "locale", "ko")
		locale = manager.GetLocaleFromContext(ctx)
		assert.Equal(t, LocaleKorean, locale)

		// Test fallback
		ctx = context.Background()
		locale = manager.GetLocaleFromContext(ctx)
		assert.Equal(t, LocaleEnglish, locale)
	})

	t.Run("Localize error", func(t *testing.T) {
		manager := NewI18nManager(LocaleEnglish)

		catalog := MessageCatalog{
			"test.error": {
				Text:        "Error: {{field}} is invalid",
				Description: "The field {{field}} has value {{value}}",
				Suggestions: []string{"Check {{field}} configuration"},
			},
		}
		manager.AddCatalog(LocaleEnglish, catalog)

		err := NewError("test", "validation", "error").
			I18nKey("test.error").
			Context("field", "username").
			Context("value", "").
			Build()

		localized := manager.LocalizeError(err, LocaleEnglish)
		assert.Equal(t, "Error: username is invalid", localized.Message)
		assert.Contains(t, localized.Description, "username")
		assert.Contains(t, localized.Suggestions[0], "username")
	})
}

func TestGlobalI18n(t *testing.T) {
	t.Run("Global instance", func(t *testing.T) {
		i18n1 := GetGlobalI18n()
		i18n2 := GetGlobalI18n()

		// Should be the same instance
		assert.Same(t, i18n1, i18n2)

		// Should have default locales
		locales := i18n1.SupportedLocales()
		assert.Contains(t, locales, LocaleEnglish)
		assert.Contains(t, locales, LocaleKorean)
	})

	t.Run("Localize with context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "locale", LocaleKorean)

		err := NewLocalizedError(DomainGitHub, CategoryAuth, "INVALID_TOKEN").
			Context("provider", "GitHub").
			Build()

		localized := LocalizeErrorWithContext(ctx, err)
		assert.Contains(t, localized.Message, "GitHub")
		// Should be in Korean
		assert.Contains(t, localized.Message, "인증")
	})
}

func TestNewLocalizedError(t *testing.T) {
	t.Run("Creates error with i18n key", func(t *testing.T) {
		err := NewLocalizedError(DomainConfig, CategoryValidation, "INVALID_FIELD").
			Context("field", "database.host").
			Build()

		assert.Equal(t, "config.validation.invalid_field", err.i18nKey)
		assert.Equal(t, "CONFIG_VALIDATION_INVALID_FIELD", err.Code.String())
	})
}
