// Package i18n provides internationalization support for GZH Manager
package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

// Manager handles all internationalization operations
type Manager struct {
	bundle       *i18n.Bundle
	localizer    *i18n.Localizer
	currentLang  language.Tag
	fallbackLang language.Tag
	localesDir   string
	mutex        sync.RWMutex
}

// Config holds i18n configuration
type Config struct {
	// LocalesDir is the directory containing locale files
	LocalesDir string
	// DefaultLanguage is the default language to use
	DefaultLanguage string
	// FallbackLanguage is used when a message is not found in the default language
	FallbackLanguage string
	// SupportedLanguages is a list of supported language codes
	SupportedLanguages []string
}

// MessageConfig holds message template configuration
type MessageConfig struct {
	ID           string                 `json:"id"`
	Description  string                 `json:"description,omitempty"`
	Message      string                 `json:"message"`
	Zero         string                 `json:"zero,omitempty"`
	One          string                 `json:"one,omitempty"`
	Two          string                 `json:"two,omitempty"`
	Few          string                 `json:"few,omitempty"`
	Many         string                 `json:"many,omitempty"`
	Other        string                 `json:"other,omitempty"`
	TemplateData map[string]interface{} `json:"template_data,omitempty"`
}

// LocalizationBundle represents a complete localization package
type LocalizationBundle struct {
	Language string                   `json:"language"`
	Version  string                   `json:"version"`
	Messages map[string]MessageConfig `json:"messages"`
}

// MessageKey constants for commonly used messages
const (
	// General messages
	MsgWelcome   = "welcome"
	MsgError     = "error"
	MsgSuccess   = "success"
	MsgWarning   = "warning"
	MsgInfo      = "info"
	MsgConfirm   = "confirm"
	MsgCancel    = "cancel"
	MsgContinue  = "continue"
	MsgRetry     = "retry"
	MsgHelp      = "help"
	MsgVersion   = "version"
	MsgLoading   = "loading"
	MsgCompleted = "completed"
	MsgFailed    = "failed"
	MsgSkipped   = "skipped"

	// Command messages
	MsgCmdBulkClone    = "cmd.bulk_clone"
	MsgCmdAlwaysLatest = "cmd.always_latest"
	MsgCmdDevEnv       = "cmd.dev_env"
	MsgCmdNetEnv       = "cmd.net_env"
	MsgCmdIDE          = "cmd.ide"
	MsgCmdGenConfig    = "cmd.gen_config"
	MsgCmdSSHConfig    = "cmd.ssh_config"
	MsgCmdDocker       = "cmd.docker"
	MsgCmdPlugin       = "cmd.plugin"

	// Clone operation messages
	MsgCloneStarting    = "clone.starting"
	MsgCloneDiscovering = "clone.discovering"
	MsgCloneFiltering   = "clone.filtering"
	MsgCloneProcessing  = "clone.processing"
	MsgCloneCompleted   = "clone.completed"
	MsgCloneFailed      = "clone.failed"
	MsgCloneStats       = "clone.stats"
	MsgCloneProgress    = "clone.progress"

	// Plugin messages
	MsgPluginLoading   = "plugin.loading"
	MsgPluginLoaded    = "plugin.loaded"
	MsgPluginFailed    = "plugin.failed"
	MsgPluginExecuting = "plugin.executing"
	MsgPluginCompleted = "plugin.completed"
	MsgPluginNotFound  = "plugin.not_found"
	MsgPluginInvalid   = "plugin.invalid"

	// Error messages
	MsgErrInvalidConfig    = "error.invalid_config"
	MsgErrFileNotFound     = "error.file_not_found"
	MsgErrPermissionDenied = "error.permission_denied"
	MsgErrNetworkError     = "error.network_error"
	MsgErrAuthFailed       = "error.auth_failed"
	MsgErrTimeout          = "error.timeout"
	MsgErrUnknown          = "error.unknown"

	// Docker messages
	MsgDockerGenerating = "docker.generating"
	MsgDockerGenerated  = "docker.generated"
	MsgDockerOptimizing = "docker.optimizing"
	MsgDockerScanning   = "docker.scanning"
	MsgDockerLanguage   = "docker.language"
	MsgDockerFramework  = "docker.framework"
)

// DefaultConfig returns a default i18n configuration
func DefaultConfig() *Config {
	return &Config{
		LocalesDir:       "locales",
		DefaultLanguage:  "en",
		FallbackLanguage: "en",
		SupportedLanguages: []string{
			"en", "ko", "ja", "zh", "zh-CN", "zh-TW",
			"es", "fr", "de", "it", "pt", "ru",
		},
	}
}

// NewManager creates a new i18n manager
func NewManager(config *Config) (*Manager, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Parse language tags
	defaultLang, err := language.Parse(config.DefaultLanguage)
	if err != nil {
		return nil, fmt.Errorf("invalid default language %s: %w", config.DefaultLanguage, err)
	}

	fallbackLang, err := language.Parse(config.FallbackLanguage)
	if err != nil {
		return nil, fmt.Errorf("invalid fallback language %s: %w", config.FallbackLanguage, err)
	}

	// Create bundle
	bundle := i18n.NewBundle(defaultLang)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
	bundle.RegisterUnmarshalFunc("toml", nil) // Add TOML support if needed

	manager := &Manager{
		bundle:       bundle,
		currentLang:  defaultLang,
		fallbackLang: fallbackLang,
		localesDir:   config.LocalesDir,
	}

	// Load locale files
	if err := manager.LoadLocales(); err != nil {
		return nil, fmt.Errorf("failed to load locales: %w", err)
	}

	// Create localizer
	manager.localizer = i18n.NewLocalizer(bundle, config.DefaultLanguage, config.FallbackLanguage)

	return manager, nil
}

// LoadLocales loads all locale files from the locales directory
func (m *Manager) LoadLocales() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, err := os.Stat(m.localesDir); os.IsNotExist(err) {
		// Create locales directory if it doesn't exist
		if err := os.MkdirAll(m.localesDir, 0o755); err != nil {
			return fmt.Errorf("failed to create locales directory: %w", err)
		}
		// Generate default locale files
		return m.generateDefaultLocales()
	}

	return filepath.Walk(m.localesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Load JSON and TOML files
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".json" || ext == ".toml" {
			if _, err := m.bundle.LoadMessageFile(path); err != nil {
				return fmt.Errorf("failed to load locale file %s: %w", path, err)
			}
		}

		return nil
	})
}

// SetLanguage sets the current language
func (m *Manager) SetLanguage(lang string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	langTag, err := language.Parse(lang)
	if err != nil {
		return fmt.Errorf("invalid language %s: %w", lang, err)
	}

	m.currentLang = langTag
	m.localizer = i18n.NewLocalizer(m.bundle, lang, m.fallbackLang.String())

	return nil
}

// GetCurrentLanguage returns the current language
func (m *Manager) GetCurrentLanguage() string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.currentLang.String()
}

// T translates a message with the given key and optional template data
func (m *Manager) T(messageID string, templateData ...map[string]interface{}) string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	localizeConfig := &i18n.LocalizeConfig{
		MessageID: messageID,
	}

	if len(templateData) > 0 && templateData[0] != nil {
		localizeConfig.TemplateData = templateData[0]
	}

	message, err := m.localizer.Localize(localizeConfig)
	if err != nil {
		// Return the message ID if translation fails
		return messageID
	}

	return message
}

// Tn translates a message with pluralization support
func (m *Manager) Tn(messageID string, count int, templateData ...map[string]interface{}) string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	localizeConfig := &i18n.LocalizeConfig{
		MessageID:    messageID,
		PluralCount:  count,
		TemplateData: map[string]interface{}{"Count": count},
	}

	if len(templateData) > 0 && templateData[0] != nil {
		if templateMap, ok := localizeConfig.TemplateData.(map[string]interface{}); ok {
			for k, v := range templateData[0] {
				templateMap[k] = v
			}
		}
	}

	message, err := m.localizer.Localize(localizeConfig)
	if err != nil {
		return messageID
	}

	return message
}

// Tf translates a message with formatted template data
func (m *Manager) Tf(messageID string, args ...interface{}) string {
	translated := m.T(messageID)
	if len(args) > 0 {
		return fmt.Sprintf(translated, args...)
	}
	return translated
}

// MustT translates a message and panics if translation fails
func (m *Manager) MustT(messageID string, templateData ...map[string]interface{}) string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	localizeConfig := &i18n.LocalizeConfig{
		MessageID: messageID,
	}

	if len(templateData) > 0 && templateData[0] != nil {
		localizeConfig.TemplateData = templateData[0]
	}

	message, err := m.localizer.Localize(localizeConfig)
	if err != nil {
		panic(fmt.Sprintf("failed to translate message %s: %v", messageID, err))
	}

	return message
}

// GetAvailableLanguages returns a list of available languages
func (m *Manager) GetAvailableLanguages() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var languages []string
	for _, tag := range m.bundle.LanguageTags() {
		languages = append(languages, tag.String())
	}
	return languages
}

// AddMessageBundle adds a message bundle for a specific language
func (m *Manager) AddMessageBundle(lang string, bundle *LocalizationBundle) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	langTag, err := language.Parse(lang)
	if err != nil {
		return fmt.Errorf("invalid language %s: %w", lang, err)
	}

	// Convert our bundle format to go-i18n format
	messages := make(map[string]*i18n.Message)
	for id, config := range bundle.Messages {
		msg := &i18n.Message{
			ID:          id,
			Description: config.Description,
			Other:       config.Message,
		}

		// Add plural forms if present
		if config.Zero != "" {
			msg.Zero = config.Zero
		}
		if config.One != "" {
			msg.One = config.One
		}
		if config.Two != "" {
			msg.Two = config.Two
		}
		if config.Few != "" {
			msg.Few = config.Few
		}
		if config.Many != "" {
			msg.Many = config.Many
		}
		if config.Other != "" {
			msg.Other = config.Other
		}

		messages[id] = msg
	}

	for _, msg := range messages {
		m.bundle.AddMessages(langTag, msg)
	}
	return nil
}

// ExportMessages exports all messages for a specific language
func (m *Manager) ExportMessages(lang string) (*LocalizationBundle, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	_, err := language.Parse(lang)
	if err != nil {
		return nil, fmt.Errorf("invalid language %s: %w", lang, err)
	}

	// This is a simplified export - in a real implementation,
	// you'd need to access the bundle's internal message store
	return &LocalizationBundle{
		Language: lang,
		Version:  "1.0.0",
		Messages: make(map[string]MessageConfig),
	}, nil
}

// ValidateMessages validates all messages for completeness
func (m *Manager) ValidateMessages() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var issues []string
	// Implementation would check for missing translations, invalid syntax, etc.
	return issues
}

// generateDefaultLocales creates default locale files
func (m *Manager) generateDefaultLocales() error {
	// Generate English locale
	enBundle := &LocalizationBundle{
		Language: "en",
		Version:  "1.0.0",
		Messages: map[string]MessageConfig{
			MsgWelcome: {
				ID:      MsgWelcome,
				Message: "Welcome to GZH Manager",
			},
			MsgError: {
				ID:      MsgError,
				Message: "Error",
			},
			MsgSuccess: {
				ID:      MsgSuccess,
				Message: "Success",
			},
			MsgCloneStarting: {
				ID:      MsgCloneStarting,
				Message: "Starting bulk clone operation",
			},
			MsgCloneCompleted: {
				ID:      MsgCloneCompleted,
				Message: "Clone operation completed successfully",
			},
			MsgCloneStats: {
				ID:      MsgCloneStats,
				Message: "Cloned {{.Cloned}} repositories, {{.Failed}} failed, {{.Skipped}} skipped",
			},
		},
	}

	if err := m.saveLocaleFile("en.json", enBundle); err != nil {
		return err
	}

	// Generate Korean locale
	koBundle := &LocalizationBundle{
		Language: "ko",
		Version:  "1.0.0",
		Messages: map[string]MessageConfig{
			MsgWelcome: {
				ID:      MsgWelcome,
				Message: "GZH Manager에 오신 것을 환영합니다",
			},
			MsgError: {
				ID:      MsgError,
				Message: "오류",
			},
			MsgSuccess: {
				ID:      MsgSuccess,
				Message: "성공",
			},
			MsgCloneStarting: {
				ID:      MsgCloneStarting,
				Message: "대량 복제 작업을 시작합니다",
			},
			MsgCloneCompleted: {
				ID:      MsgCloneCompleted,
				Message: "복제 작업이 성공적으로 완료되었습니다",
			},
			MsgCloneStats: {
				ID:      MsgCloneStats,
				Message: "{{.Cloned}}개 저장소 복제됨, {{.Failed}}개 실패, {{.Skipped}}개 건너뜀",
			},
		},
	}

	return m.saveLocaleFile("ko.json", koBundle)
}

// saveLocaleFile saves a locale bundle to a file
func (m *Manager) saveLocaleFile(filename string, bundle *LocalizationBundle) error {
	filePath := filepath.Join(m.localesDir, filename)

	data, err := json.MarshalIndent(bundle, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal locale bundle: %w", err)
	}

	return os.WriteFile(filePath, data, 0o644)
}

// Global manager instance
var globalManager *Manager

// Init initializes the global i18n manager
func Init(config *Config) error {
	var err error
	globalManager, err = NewManager(config)
	return err
}

// SetLanguage sets the language for the global manager
func SetLanguage(lang string) error {
	if globalManager == nil {
		return fmt.Errorf("i18n manager not initialized")
	}
	return globalManager.SetLanguage(lang)
}

// T translates using the global manager
func T(messageID string, templateData ...map[string]interface{}) string {
	if globalManager == nil {
		return messageID
	}
	return globalManager.T(messageID, templateData...)
}

// Tn translates with pluralization using the global manager
func Tn(messageID string, count int, templateData ...map[string]interface{}) string {
	if globalManager == nil {
		return messageID
	}
	return globalManager.Tn(messageID, count, templateData...)
}

// Tf translates with formatting using the global manager
func Tf(messageID string, args ...interface{}) string {
	if globalManager == nil {
		return messageID
	}
	return globalManager.Tf(messageID, args...)
}

// GetManager returns the global manager
func GetManager() *Manager {
	return globalManager
}

// GetSupportedLanguages returns list of supported languages
func (m *Manager) GetSupportedLanguages() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Return languages from configuration or scan available locale files
	supportedLangs := []string{}

	// Scan locale directory for available language files
	if m.localesDir != "" {
		if files, err := filepath.Glob(filepath.Join(m.localesDir, "*.json")); err == nil {
			for _, file := range files {
				basename := filepath.Base(file)
				lang := strings.TrimSuffix(basename, filepath.Ext(basename))
				if lang != "" {
					supportedLangs = append(supportedLangs, lang)
				}
			}
		}
	}

	// If no files found, return default supported languages
	if len(supportedLangs) == 0 {
		supportedLangs = []string{"en", "ko"}
	}

	return supportedLangs
}

// IsLanguageSupported checks if a language is supported
func (m *Manager) IsLanguageSupported(lang string) bool {
	supported := m.GetSupportedLanguages()
	for _, supportedLang := range supported {
		if supportedLang == lang {
			return true
		}
	}
	return false
}

// Localize translates a message (alias for T method)
func (m *Manager) Localize(messageID string, templateData ...map[string]interface{}) string {
	return m.T(messageID, templateData...)
}
