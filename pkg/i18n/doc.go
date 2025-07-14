// Package i18n provides internationalization and localization support
// for the GZH Manager system.
//
// This package implements comprehensive multi-language support, enabling
// the GZH Manager to provide localized user experiences across different
// languages, regions, and cultural contexts.
//
// Key Components:
//
// Locale Manager:
//   - Language detection and selection
//   - Regional preference handling
//   - Fallback language support
//   - Dynamic locale switching
//
// Message Extraction:
//   - Automatic message extraction from source code
//   - Template-based message identification
//   - Pluralization rule handling
//   - Context-aware message extraction
//
// Translation Management:
//   - Translation file management (JSON, YAML, PO)
//   - Translation validation and verification
//   - Missing translation detection
//   - Translation quality assurance
//
// Runtime Localization:
//   - Real-time message localization
//   - Template variable substitution
//   - Number and date formatting
//   - Cultural adaptation
//
// Features:
//   - Support for 50+ languages
//   - Pluralization rules for all languages
//   - RTL (Right-to-Left) language support
//   - Currency and number formatting
//   - Date and time localization
//   - Cultural context awareness
//
// Example usage:
//
//	manager := i18n.NewManager()
//	manager.LoadTranslations("locales/")
//
//	localizer := manager.GetLocalizer("en-US")
//	message := localizer.Localize("welcome.message", data)
//
//	extractor := i18n.NewExtractor()
//	messages := extractor.ExtractFromFiles("*.go")
//
// The package enables global accessibility and user experience
// optimization across different languages and cultures.
package i18n
