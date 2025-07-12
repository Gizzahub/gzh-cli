package i18n

import (
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// LocaleManager handles locale-specific formatting and operations
type LocaleManager struct {
	currentLocale language.Tag
	printer       *message.Printer
	timeFormats   map[string]string
	dateFormats   map[string]string
	numberFormats map[string]string
}

// LocaleConfig holds locale-specific configuration
type LocaleConfig struct {
	Language     language.Tag
	TimeFormat   string
	DateFormat   string
	NumberFormat string
	CurrencyCode string
	Timezone     string
}

// SupportedLocales defines all supported locales with their configurations
var SupportedLocales = map[string]LocaleConfig{
	"en": {
		Language:     language.English,
		TimeFormat:   "15:04:05",
		DateFormat:   "2006-01-02",
		NumberFormat: "en",
		CurrencyCode: "USD",
		Timezone:     "UTC",
	},
	"en-US": {
		Language:     language.AmericanEnglish,
		TimeFormat:   "3:04:05 PM",
		DateFormat:   "01/02/2006",
		NumberFormat: "en-US",
		CurrencyCode: "USD",
		Timezone:     "America/New_York",
	},
	"ko": {
		Language:     language.Korean,
		TimeFormat:   "15:04:05",
		DateFormat:   "2006년 01월 02일",
		NumberFormat: "ko",
		CurrencyCode: "KRW",
		Timezone:     "Asia/Seoul",
	},
	"ko-KR": {
		Language:     language.Korean,
		TimeFormat:   "오후 3:04:05",
		DateFormat:   "2006년 1월 2일",
		NumberFormat: "ko-KR",
		CurrencyCode: "KRW",
		Timezone:     "Asia/Seoul",
	},
	"ja": {
		Language:     language.Japanese,
		TimeFormat:   "15:04:05",
		DateFormat:   "2006年01月02日",
		NumberFormat: "ja",
		CurrencyCode: "JPY",
		Timezone:     "Asia/Tokyo",
	},
	"ja-JP": {
		Language:     language.Japanese,
		TimeFormat:   "午後3:04:05",
		DateFormat:   "2006年1月2日",
		NumberFormat: "ja-JP",
		CurrencyCode: "JPY",
		Timezone:     "Asia/Tokyo",
	},
	"zh": {
		Language:     language.Chinese,
		TimeFormat:   "15:04:05",
		DateFormat:   "2006年01月02日",
		NumberFormat: "zh",
		CurrencyCode: "CNY",
		Timezone:     "Asia/Shanghai",
	},
	"zh-CN": {
		Language:     language.SimplifiedChinese,
		TimeFormat:   "下午3:04:05",
		DateFormat:   "2006年1月2日",
		NumberFormat: "zh-CN",
		CurrencyCode: "CNY",
		Timezone:     "Asia/Shanghai",
	},
	"zh-TW": {
		Language:     language.TraditionalChinese,
		TimeFormat:   "下午3:04:05",
		DateFormat:   "2006年1月2日",
		NumberFormat: "zh-TW",
		CurrencyCode: "TWD",
		Timezone:     "Asia/Taipei",
	},
	"es": {
		Language:     language.Spanish,
		TimeFormat:   "15:04:05",
		DateFormat:   "02/01/2006",
		NumberFormat: "es",
		CurrencyCode: "EUR",
		Timezone:     "Europe/Madrid",
	},
	"fr": {
		Language:     language.French,
		TimeFormat:   "15:04:05",
		DateFormat:   "02/01/2006",
		NumberFormat: "fr",
		CurrencyCode: "EUR",
		Timezone:     "Europe/Paris",
	},
	"de": {
		Language:     language.German,
		TimeFormat:   "15:04:05",
		DateFormat:   "02.01.2006",
		NumberFormat: "de",
		CurrencyCode: "EUR",
		Timezone:     "Europe/Berlin",
	},
	"it": {
		Language:     language.Italian,
		TimeFormat:   "15:04:05",
		DateFormat:   "02/01/2006",
		NumberFormat: "it",
		CurrencyCode: "EUR",
		Timezone:     "Europe/Rome",
	},
	"pt": {
		Language:     language.Portuguese,
		TimeFormat:   "15:04:05",
		DateFormat:   "02/01/2006",
		NumberFormat: "pt",
		CurrencyCode: "EUR",
		Timezone:     "Europe/Lisbon",
	},
	"ru": {
		Language:     language.Russian,
		TimeFormat:   "15:04:05",
		DateFormat:   "02.01.2006",
		NumberFormat: "ru",
		CurrencyCode: "RUB",
		Timezone:     "Europe/Moscow",
	},
}

// NewLocaleManager creates a new locale manager
func NewLocaleManager(locale string) (*LocaleManager, error) {
	config, exists := SupportedLocales[locale]
	if !exists {
		// Try to find base language
		baseLang := strings.Split(locale, "-")[0]
		if baseConfig, baseExists := SupportedLocales[baseLang]; baseExists {
			config = baseConfig
		} else {
			// Default to English
			config = SupportedLocales["en"]
		}
	}

	printer := message.NewPrinter(config.Language)

	return &LocaleManager{
		currentLocale: config.Language,
		printer:       printer,
		timeFormats: map[string]string{
			"short": config.TimeFormat,
			"long":  "15:04:05 MST",
			"full":  "15:04:05 MST 2006-01-02",
		},
		dateFormats: map[string]string{
			"short": config.DateFormat,
			"long":  "January 2, 2006",
			"full":  "Monday, January 2, 2006",
		},
		numberFormats: map[string]string{
			"decimal":  "#,##0.##",
			"currency": "¤#,##0.00",
			"percent":  "#0.00%",
		},
	}, nil
}

// GetCurrentLocale returns the current locale
func (lm *LocaleManager) GetCurrentLocale() language.Tag {
	return lm.currentLocale
}

// FormatTime formats a time according to the current locale
func (lm *LocaleManager) FormatTime(t time.Time, format string) string {
	if formatStr, exists := lm.timeFormats[format]; exists {
		return t.Format(formatStr)
	}
	return t.Format(format)
}

// FormatDate formats a date according to the current locale
func (lm *LocaleManager) FormatDate(t time.Time, format string) string {
	if formatStr, exists := lm.dateFormats[format]; exists {
		return t.Format(formatStr)
	}
	return t.Format(format)
}

// FormatDateTime formats a date and time according to the current locale
func (lm *LocaleManager) FormatDateTime(t time.Time, dateFormat, timeFormat string) string {
	date := lm.FormatDate(t, dateFormat)
	time := lm.FormatTime(t, timeFormat)
	return fmt.Sprintf("%s %s", date, time)
}

// FormatNumber formats a number according to the current locale
func (lm *LocaleManager) FormatNumber(number interface{}) string {
	switch v := number.(type) {
	case int:
		return lm.printer.Sprintf("%d", v)
	case int64:
		return lm.printer.Sprintf("%d", v)
	case float64:
		return lm.printer.Sprintf("%.2f", v)
	case float32:
		return lm.printer.Sprintf("%.2f", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// FormatCurrency formats a currency amount according to the current locale
func (lm *LocaleManager) FormatCurrency(amount float64, currency string) string {
	return lm.printer.Sprintf("%.2f %s", amount, currency)
}

// FormatPercent formats a percentage according to the current locale
func (lm *LocaleManager) FormatPercent(value float64) string {
	return lm.printer.Sprintf("%.2f%%", value*100)
}

// DetectSystemLocale detects the system locale from environment variables
func DetectSystemLocale() string {
	// Check environment variables in order of preference
	envVars := []string{"LC_ALL", "LC_MESSAGES", "LANG", "LANGUAGE"}

	for _, envVar := range envVars {
		if value := os.Getenv(envVar); value != "" {
			// Parse locale string (e.g., "en_US.UTF-8" -> "en-US")
			locale := parseLocaleString(value)
			if locale != "" {
				return locale
			}
		}
	}

	// Default to English if no locale found
	return "en"
}

// parseLocaleString parses a locale string and returns a standardized format
func parseLocaleString(localeStr string) string {
	// Remove encoding and modifiers (e.g., "en_US.UTF-8@euro" -> "en_US")
	parts := strings.Split(localeStr, ".")
	if len(parts) > 0 {
		localeStr = parts[0]
	}

	parts = strings.Split(localeStr, "@")
	if len(parts) > 0 {
		localeStr = parts[0]
	}

	// Convert underscore to hyphen (e.g., "en_US" -> "en-US")
	localeStr = strings.Replace(localeStr, "_", "-", -1)

	// Normalize case
	if strings.Contains(localeStr, "-") {
		parts := strings.Split(localeStr, "-")
		if len(parts) == 2 {
			return strings.ToLower(parts[0]) + "-" + strings.ToUpper(parts[1])
		}
	}

	return strings.ToLower(localeStr)
}

// GetSupportedLocales returns a list of all supported locales
func GetSupportedLocales() []string {
	var locales []string
	for locale := range SupportedLocales {
		locales = append(locales, locale)
	}
	return locales
}

// IsLocaleSupported checks if a locale is supported
func IsLocaleSupported(locale string) bool {
	_, exists := SupportedLocales[locale]
	if !exists {
		// Check base language
		baseLang := strings.Split(locale, "-")[0]
		_, exists = SupportedLocales[baseLang]
	}
	return exists
}

// GetLocaleInfo returns information about a locale
func GetLocaleInfo(locale string) (LocaleConfig, error) {
	config, exists := SupportedLocales[locale]
	if !exists {
		baseLang := strings.Split(locale, "-")[0]
		if baseConfig, baseExists := SupportedLocales[baseLang]; baseExists {
			return baseConfig, nil
		}
		return LocaleConfig{}, fmt.Errorf("unsupported locale: %s", locale)
	}
	return config, nil
}

// GetTimeZoneOffset returns the timezone offset for a locale
func GetTimeZoneOffset(locale string) (int, error) {
	config, err := GetLocaleInfo(locale)
	if err != nil {
		return 0, err
	}

	location, err := time.LoadLocation(config.Timezone)
	if err != nil {
		return 0, fmt.Errorf("failed to load timezone %s: %w", config.Timezone, err)
	}

	_, offset := time.Now().In(location).Zone()
	return offset, nil
}

// FormatRelativeTime formats a relative time (e.g., "2 hours ago", "in 3 days")
func (lm *LocaleManager) FormatRelativeTime(t time.Time) string {
	now := time.Now()
	duration := now.Sub(t)

	if duration < 0 {
		// Future time
		duration = -duration
		return lm.formatFutureTime(duration)
	}

	// Past time
	return lm.formatPastTime(duration)
}

// formatPastTime formats past relative time
func (lm *LocaleManager) formatPastTime(duration time.Duration) string {
	seconds := int(duration.Seconds())
	minutes := seconds / 60
	hours := minutes / 60
	days := hours / 24
	weeks := days / 7
	months := days / 30
	years := days / 365

	switch lm.currentLocale {
	case language.Korean:
		switch {
		case years > 0:
			return fmt.Sprintf("%d년 전", years)
		case months > 0:
			return fmt.Sprintf("%d개월 전", months)
		case weeks > 0:
			return fmt.Sprintf("%d주 전", weeks)
		case days > 0:
			return fmt.Sprintf("%d일 전", days)
		case hours > 0:
			return fmt.Sprintf("%d시간 전", hours)
		case minutes > 0:
			return fmt.Sprintf("%d분 전", minutes)
		default:
			return "방금 전"
		}
	case language.Japanese:
		switch {
		case years > 0:
			return fmt.Sprintf("%d年前", years)
		case months > 0:
			return fmt.Sprintf("%d月前", months)
		case weeks > 0:
			return fmt.Sprintf("%d週間前", weeks)
		case days > 0:
			return fmt.Sprintf("%d日前", days)
		case hours > 0:
			return fmt.Sprintf("%d時間前", hours)
		case minutes > 0:
			return fmt.Sprintf("%d分前", minutes)
		default:
			return "たった今"
		}
	case language.SimplifiedChinese, language.TraditionalChinese:
		switch {
		case years > 0:
			return fmt.Sprintf("%d年前", years)
		case months > 0:
			return fmt.Sprintf("%d个月前", months)
		case weeks > 0:
			return fmt.Sprintf("%d周前", weeks)
		case days > 0:
			return fmt.Sprintf("%d天前", days)
		case hours > 0:
			return fmt.Sprintf("%d小时前", hours)
		case minutes > 0:
			return fmt.Sprintf("%d分钟前", minutes)
		default:
			return "刚才"
		}
	default: // English and others
		switch {
		case years > 0:
			if years == 1 {
				return "1 year ago"
			}
			return fmt.Sprintf("%d years ago", years)
		case months > 0:
			if months == 1 {
				return "1 month ago"
			}
			return fmt.Sprintf("%d months ago", months)
		case weeks > 0:
			if weeks == 1 {
				return "1 week ago"
			}
			return fmt.Sprintf("%d weeks ago", weeks)
		case days > 0:
			if days == 1 {
				return "1 day ago"
			}
			return fmt.Sprintf("%d days ago", days)
		case hours > 0:
			if hours == 1 {
				return "1 hour ago"
			}
			return fmt.Sprintf("%d hours ago", hours)
		case minutes > 0:
			if minutes == 1 {
				return "1 minute ago"
			}
			return fmt.Sprintf("%d minutes ago", minutes)
		default:
			return "just now"
		}
	}
}

// formatFutureTime formats future relative time
func (lm *LocaleManager) formatFutureTime(duration time.Duration) string {
	seconds := int(duration.Seconds())
	minutes := seconds / 60
	hours := minutes / 60
	days := hours / 24
	weeks := days / 7
	months := days / 30
	years := days / 365

	switch lm.currentLocale {
	case language.Korean:
		switch {
		case years > 0:
			return fmt.Sprintf("%d년 후", years)
		case months > 0:
			return fmt.Sprintf("%d개월 후", months)
		case weeks > 0:
			return fmt.Sprintf("%d주 후", weeks)
		case days > 0:
			return fmt.Sprintf("%d일 후", days)
		case hours > 0:
			return fmt.Sprintf("%d시간 후", hours)
		case minutes > 0:
			return fmt.Sprintf("%d분 후", minutes)
		default:
			return "곧"
		}
	case language.Japanese:
		switch {
		case years > 0:
			return fmt.Sprintf("%d年後", years)
		case months > 0:
			return fmt.Sprintf("%d月後", months)
		case weeks > 0:
			return fmt.Sprintf("%d週間後", weeks)
		case days > 0:
			return fmt.Sprintf("%d日後", days)
		case hours > 0:
			return fmt.Sprintf("%d時間後", hours)
		case minutes > 0:
			return fmt.Sprintf("%d分後", minutes)
		default:
			return "すぐに"
		}
	case language.SimplifiedChinese, language.TraditionalChinese:
		switch {
		case years > 0:
			return fmt.Sprintf("%d年后", years)
		case months > 0:
			return fmt.Sprintf("%d个月后", months)
		case weeks > 0:
			return fmt.Sprintf("%d周后", weeks)
		case days > 0:
			return fmt.Sprintf("%d天后", days)
		case hours > 0:
			return fmt.Sprintf("%d小时后", hours)
		case minutes > 0:
			return fmt.Sprintf("%d分钟后", minutes)
		default:
			return "很快"
		}
	default: // English and others
		switch {
		case years > 0:
			if years == 1 {
				return "in 1 year"
			}
			return fmt.Sprintf("in %d years", years)
		case months > 0:
			if months == 1 {
				return "in 1 month"
			}
			return fmt.Sprintf("in %d months", months)
		case weeks > 0:
			if weeks == 1 {
				return "in 1 week"
			}
			return fmt.Sprintf("in %d weeks", weeks)
		case days > 0:
			if days == 1 {
				return "in 1 day"
			}
			return fmt.Sprintf("in %d days", days)
		case hours > 0:
			if hours == 1 {
				return "in 1 hour"
			}
			return fmt.Sprintf("in %d hours", hours)
		case minutes > 0:
			if minutes == 1 {
				return "in 1 minute"
			}
			return fmt.Sprintf("in %d minutes", minutes)
		default:
			return "soon"
		}
	}
}

// Global locale manager instance
var globalLocaleManager *LocaleManager

// InitLocale initializes the global locale manager
func InitLocale(locale string) error {
	var err error
	globalLocaleManager, err = NewLocaleManager(locale)
	return err
}

// FormatTime formats time using the global locale manager
func FormatTime(t time.Time, format string) string {
	if globalLocaleManager == nil {
		return t.Format(format)
	}
	return globalLocaleManager.FormatTime(t, format)
}

// FormatDate formats date using the global locale manager
func FormatDate(t time.Time, format string) string {
	if globalLocaleManager == nil {
		return t.Format(format)
	}
	return globalLocaleManager.FormatDate(t, format)
}

// FormatRelativeTime formats relative time using the global locale manager
func FormatRelativeTime(t time.Time) string {
	if globalLocaleManager == nil {
		return t.Format(time.RFC3339)
	}
	return globalLocaleManager.FormatRelativeTime(t)
}
