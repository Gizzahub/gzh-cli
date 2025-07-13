package errors

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Locale represents a language locale (e.g., "en", "ko", "ja")
type Locale string

const (
	LocaleEnglish  Locale = "en"
	LocaleKorean   Locale = "ko"
	LocaleJapanese Locale = "ja"
	LocaleSpanish  Locale = "es"
	LocaleFrench   Locale = "fr"
	LocaleGerman   Locale = "de"
	LocaleChinese  Locale = "zh"
)

// Message represents a localized message with placeholders
type Message struct {
	Text         string            `json:"text"`
	Description  string            `json:"description,omitempty"`
	Suggestions  []string          `json:"suggestions,omitempty"`
	Placeholders map[string]string `json:"placeholders,omitempty"`
}

// MessageCatalog holds all localized messages for a specific locale
type MessageCatalog map[string]Message

// I18nManager manages internationalized error messages
type I18nManager struct {
	mu         sync.RWMutex
	catalogs   map[Locale]MessageCatalog
	fallback   Locale
	contextKey string
}

// NewI18nManager creates a new internationalization manager
func NewI18nManager(fallback Locale) *I18nManager {
	return &I18nManager{
		catalogs:   make(map[Locale]MessageCatalog),
		fallback:   fallback,
		contextKey: "locale",
	}
}

// LoadFromFile loads message catalog from a JSON file
func (m *I18nManager) LoadFromFile(locale Locale, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read locale file %s: %w", filePath, err)
	}

	var catalog MessageCatalog
	if err := json.Unmarshal(data, &catalog); err != nil {
		return fmt.Errorf("failed to parse locale file %s: %w", filePath, err)
	}

	m.mu.Lock()
	m.catalogs[locale] = catalog
	m.mu.Unlock()

	return nil
}

// LoadFromDirectory loads all locale files from a directory
func (m *I18nManager) LoadFromDirectory(dir string) error {
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(path, ".json") {
			return nil
		}

		// Extract locale from filename (e.g., "en.json" -> "en")
		filename := d.Name()
		locale := Locale(strings.TrimSuffix(filename, ".json"))

		return m.LoadFromFile(locale, path)
	})
}

// AddCatalog adds a message catalog for a locale
func (m *I18nManager) AddCatalog(locale Locale, catalog MessageCatalog) {
	m.mu.Lock()
	m.catalogs[locale] = catalog
	m.mu.Unlock()
}

// GetMessage retrieves a localized message by key
func (m *I18nManager) GetMessage(locale Locale, key string) (Message, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Try requested locale first
	if catalog, exists := m.catalogs[locale]; exists {
		if msg, found := catalog[key]; found {
			return msg, true
		}
	}

	// Fall back to default locale
	if locale != m.fallback {
		if catalog, exists := m.catalogs[m.fallback]; exists {
			if msg, found := catalog[key]; found {
				return msg, true
			}
		}
	}

	return Message{}, false
}

// GetLocaleFromContext extracts locale from context
func (m *I18nManager) GetLocaleFromContext(ctx context.Context) Locale {
	if ctx != nil {
		if val := ctx.Value(m.contextKey); val != nil {
			if locale, ok := val.(Locale); ok {
				return locale
			}
			if localeStr, ok := val.(string); ok {
				return Locale(localeStr)
			}
		}
	}
	return m.fallback
}

// LocalizeError localizes a UserError based on context or provided locale
func (m *I18nManager) LocalizeError(err *UserError, locale ...Locale) *UserError {
	if err == nil {
		return nil
	}

	var targetLocale Locale
	if len(locale) > 0 {
		targetLocale = locale[0]
	} else {
		targetLocale = m.fallback
	}

	// If no i18n key, return original error
	if err.i18nKey == "" {
		return err
	}

	msg, found := m.GetMessage(targetLocale, err.i18nKey)
	if !found {
		return err
	}

	// Create localized copy
	localized := &UserError{
		Code:        err.Code,
		Message:     m.interpolateMessage(msg.Text, err.Context),
		Description: m.interpolateMessage(msg.Description, err.Context),
		Suggestions: m.interpolateList(msg.Suggestions, err.Context),
		Context:     err.Context,
		Timestamp:   err.Timestamp,
		RequestID:   err.RequestID,
		StackTrace:  err.StackTrace,
		Cause:       err.Cause,
		i18nKey:     err.i18nKey,
	}

	return localized
}

// interpolateMessage replaces placeholders in message with context values
func (m *I18nManager) interpolateMessage(message string, context map[string]interface{}) string {
	if message == "" || context == nil {
		return message
	}

	result := message
	for key, value := range context {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
	}

	return result
}

// interpolateList replaces placeholders in a list of strings
func (m *I18nManager) interpolateList(list []string, context map[string]interface{}) []string {
	if len(list) == 0 || context == nil {
		return list
	}

	result := make([]string, len(list))
	for i, item := range list {
		result[i] = m.interpolateMessage(item, context)
	}

	return result
}

// SupportedLocales returns a list of supported locales
func (m *I18nManager) SupportedLocales() []Locale {
	m.mu.RLock()
	defer m.mu.RUnlock()

	locales := make([]Locale, 0, len(m.catalogs))
	for locale := range m.catalogs {
		locales = append(locales, locale)
	}

	return locales
}

// Global I18n manager instance
var (
	globalI18n *I18nManager
	i18nOnce   sync.Once
)

// GetGlobalI18n returns the global i18n manager instance
func GetGlobalI18n() *I18nManager {
	i18nOnce.Do(func() {
		globalI18n = NewI18nManager(LocaleEnglish)
		initDefaultMessages()
	})
	return globalI18n
}

// LocalizeErrorWithContext localizes an error using the global i18n manager and context
func LocalizeErrorWithContext(ctx context.Context, err *UserError) *UserError {
	i18n := GetGlobalI18n()
	locale := i18n.GetLocaleFromContext(ctx)
	return i18n.LocalizeError(err, locale)
}

// initDefaultMessages initializes default English and Korean messages
func initDefaultMessages() {
	// English messages
	englishCatalog := MessageCatalog{
		"config.validation.invalid_field": {
			Text:        "Invalid configuration field: {{field}}",
			Description: "The field '{{field}}' has an invalid value: {{value}}",
			Suggestions: []string{
				"Check the configuration documentation for valid values",
				"Use 'gz config validate' to verify your configuration",
			},
		},
		"github.auth.invalid_token": {
			Text:        "GitHub authentication failed",
			Description: "The provided GitHub token is invalid or has insufficient permissions",
			Suggestions: []string{
				"Check your GitHub token in GITHUB_TOKEN environment variable",
				"Ensure the token has the required permissions (repo, admin:org)",
				"Generate a new token at https://github.com/settings/tokens",
			},
		},
		"network.timeout.operation_timeout": {
			Text:        "Network operation timed out",
			Description: "The {{operation}} operation timed out after {{timeout}}",
			Suggestions: []string{
				"Check your internet connection",
				"Try increasing the timeout value",
				"Check if the remote service is available",
			},
		},
		"repository.not_found.repo_not_found": {
			Text:        "Repository not found",
			Description: "The repository '{{repository}}' was not found on {{provider}}",
			Suggestions: []string{
				"Check the repository name and owner",
				"Ensure you have access to the repository",
				"Verify the repository exists and is not private",
			},
		},
		"file.permission.access_denied": {
			Text:        "File permission denied",
			Description: "Permission denied when trying to {{operation}} file: {{path}}",
			Suggestions: []string{
				"Check file permissions and ownership",
				"Run with appropriate privileges if needed",
				"Ensure the directory exists and is writable",
			},
		},
		"api.resource.rate_limit_exceeded": {
			Text:        "API rate limit exceeded",
			Description: "You have exceeded the API rate limit for {{provider}}",
			Suggestions: []string{
				"Wait until the rate limit resets at {{reset_time}}",
				"Use a token with higher rate limits",
				"Implement request batching or caching",
			},
		},
	}

	// Korean messages
	koreanCatalog := MessageCatalog{
		"config.validation.invalid_field": {
			Text:        "잘못된 설정 필드: {{field}}",
			Description: "'{{field}}' 필드에 잘못된 값이 있습니다: {{value}}",
			Suggestions: []string{
				"유효한 값에 대한 설정 문서를 확인하세요",
				"'gz config validate' 명령어로 설정을 검증하세요",
			},
		},
		"github.auth.invalid_token": {
			Text:        "GitHub 인증 실패",
			Description: "제공된 GitHub 토큰이 유효하지 않거나 권한이 부족합니다",
			Suggestions: []string{
				"GITHUB_TOKEN 환경변수의 GitHub 토큰을 확인하세요",
				"토큰에 필요한 권한(repo, admin:org)이 있는지 확인하세요",
				"https://github.com/settings/tokens 에서 새 토큰을 생성하세요",
			},
		},
		"network.timeout.operation_timeout": {
			Text:        "네트워크 작업 시간 초과",
			Description: "{{operation}} 작업이 {{timeout}} 후 시간 초과되었습니다",
			Suggestions: []string{
				"인터넷 연결을 확인하세요",
				"시간 초과 값을 늘려보세요",
				"원격 서비스가 사용 가능한지 확인하세요",
			},
		},
		"repository.not_found.repo_not_found": {
			Text:        "저장소를 찾을 수 없음",
			Description: "{{provider}}에서 저장소 '{{repository}}'를 찾을 수 없습니다",
			Suggestions: []string{
				"저장소 이름과 소유자를 확인하세요",
				"저장소에 대한 접근 권한이 있는지 확인하세요",
				"저장소가 존재하고 비공개가 아닌지 확인하세요",
			},
		},
		"file.permission.access_denied": {
			Text:        "파일 권한 거부",
			Description: "파일 {{operation}} 시 권한이 거부되었습니다: {{path}}",
			Suggestions: []string{
				"파일 권한과 소유권을 확인하세요",
				"필요시 적절한 권한으로 실행하세요",
				"디렉터리가 존재하고 쓰기 가능한지 확인하세요",
			},
		},
		"api.resource.rate_limit_exceeded": {
			Text:        "API 속도 제한 초과",
			Description: "{{provider}}의 API 속도 제한을 초과했습니다",
			Suggestions: []string{
				"{{reset_time}}에 속도 제한이 리셋될 때까지 대기하세요",
				"더 높은 속도 제한의 토큰을 사용하세요",
				"요청 배치 처리나 캐싱을 구현하세요",
			},
		},
	}

	globalI18n.AddCatalog(LocaleEnglish, englishCatalog)
	globalI18n.AddCatalog(LocaleKorean, koreanCatalog)
}

// Helper function to create localized errors
func NewLocalizedError(domain, category, code string) *ErrorBuilder {
	i18nKey := fmt.Sprintf("%s.%s.%s",
		strings.ToLower(domain),
		strings.ToLower(category),
		strings.ToLower(code))

	return NewError(domain, category, code).I18nKey(i18nKey)
}
