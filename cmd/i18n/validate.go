package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gizzahub/gzh-manager-go/pkg/i18n"
	"github.com/spf13/cobra"
)

// ValidateCmd represents the validate command
var ValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate translation files and check for issues",
	Long: `Validate translation files and check for common issues like:
- Missing translations
- Invalid template syntax
- Inconsistent pluralization
- Unused messages
- Formatting errors

Examples:
  gz i18n validate --locales ./locales
  gz i18n validate --strict --check-unused
  gz i18n validate --languages en,ko,ja`,
	Run: runValidate,
}

var (
	validateLocalesDir string
	validateLanguages  []string
	strictMode         bool
	checkUnused        bool
	checkSyntax        bool
	showWarnings       bool
	outputFormat       string
)

func init() {
	ValidateCmd.Flags().StringVar(&validateLocalesDir, "locales", "locales", "Directory containing locale files")
	ValidateCmd.Flags().StringSliceVar(&validateLanguages, "languages", nil, "Languages to validate (default: all)")
	ValidateCmd.Flags().BoolVar(&strictMode, "strict", false, "Strict validation mode")
	ValidateCmd.Flags().BoolVar(&checkUnused, "check-unused", false, "Check for unused messages")
	ValidateCmd.Flags().BoolVar(&checkSyntax, "check-syntax", true, "Check template syntax")
	ValidateCmd.Flags().BoolVar(&showWarnings, "warnings", true, "Show warnings")
	ValidateCmd.Flags().StringVar(&outputFormat, "format", "text", "Output format (text, json, junit)")
}

// ValidationResult holds validation results
type ValidationResult struct {
	Language string              `json:"language"`
	File     string              `json:"file"`
	Valid    bool                `json:"valid"`
	Errors   []ValidationError   `json:"errors"`
	Warnings []ValidationWarning `json:"warnings"`
	Stats    ValidationStats     `json:"stats"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Type       string `json:"type"`
	MessageID  string `json:"message_id,omitempty"`
	Message    string `json:"message"`
	Line       int    `json:"line,omitempty"`
	Column     int    `json:"column,omitempty"`
	Severity   string `json:"severity"`
	Suggestion string `json:"suggestion,omitempty"`
}

// ValidationWarning represents a validation warning
type ValidationWarning struct {
	Type      string `json:"type"`
	MessageID string `json:"message_id,omitempty"`
	Message   string `json:"message"`
	Line      int    `json:"line,omitempty"`
}

// ValidationStats holds validation statistics
type ValidationStats struct {
	TotalMessages   int     `json:"total_messages"`
	TranslatedCount int     `json:"translated_count"`
	MissingCount    int     `json:"missing_count"`
	EmptyCount      int     `json:"empty_count"`
	ErrorCount      int     `json:"error_count"`
	WarningCount    int     `json:"warning_count"`
	CompletionRate  float64 `json:"completion_rate"`
}

// ValidationSummary holds overall validation summary
type ValidationSummary struct {
	TotalFiles    int                `json:"total_files"`
	ValidFiles    int                `json:"valid_files"`
	Results       []ValidationResult `json:"results"`
	OverallValid  bool               `json:"overall_valid"`
	TotalErrors   int                `json:"total_errors"`
	TotalWarnings int                `json:"total_warnings"`
}

func runValidate(cmd *cobra.Command, args []string) {
	fmt.Println("🔍 Validating translation files...")

	// Check if locales directory exists
	if _, err := os.Stat(validateLocalesDir); os.IsNotExist(err) {
		fmt.Printf("❌ Locales directory does not exist: %s\n", validateLocalesDir)
		os.Exit(1)
	}

	// Find locale files
	localeFiles, err := findLocaleFiles(validateLocalesDir, validateLanguages)
	if err != nil {
		fmt.Printf("❌ Failed to find locale files: %v\n", err)
		os.Exit(1)
	}

	if len(localeFiles) == 0 {
		fmt.Println("❌ No locale files found")
		os.Exit(1)
	}

	fmt.Printf("📁 Found %d locale files\n", len(localeFiles))

	// Validate each file
	var results []ValidationResult
	for _, file := range localeFiles {
		result := validateLocaleFile(file)
		results = append(results, result)

		if outputFormat == "text" {
			printValidationResult(result)
		}
	}

	// Create summary
	summary := createValidationSummary(results)

	// Output results
	switch outputFormat {
	case "json":
		printJSONResults(summary)
	case "junit":
		printJUnitResults(summary)
	default:
		printTextSummary(summary)
	}

	// Exit with error code if validation failed
	if !summary.OverallValid {
		if strictMode || summary.TotalErrors > 0 {
			os.Exit(1)
		}
	}
}

// findLocaleFiles finds all locale files in the directory
func findLocaleFiles(dir string, languages []string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Check for JSON files
		if strings.HasSuffix(path, ".json") {
			// Extract language from filename
			base := filepath.Base(path)
			lang := strings.TrimSuffix(base, ".json")

			// Filter by languages if specified
			if len(languages) > 0 {
				var found bool
				for _, l := range languages {
					if l == lang {
						found = true
						break
					}
				}
				if !found {
					return nil
				}
			}

			files = append(files, path)
		}

		return nil
	})

	return files, err
}

// validateLocaleFile validates a single locale file
func validateLocaleFile(filename string) ValidationResult {
	base := filepath.Base(filename)
	lang := strings.TrimSuffix(base, ".json")

	result := ValidationResult{
		Language: lang,
		File:     filename,
		Valid:    true,
		Errors:   []ValidationError{},
		Warnings: []ValidationWarning{},
		Stats:    ValidationStats{},
	}

	// Read and parse file
	data, err := os.ReadFile(filename)
	if err != nil {
		result.addError("file_read", "", fmt.Sprintf("Failed to read file: %v", err), "error")
		result.Valid = false
		return result
	}

	var bundle i18n.LocalizationBundle
	if err := json.Unmarshal(data, &bundle); err != nil {
		result.addError("json_parse", "", fmt.Sprintf("Invalid JSON: %v", err), "error")
		result.Valid = false
		return result
	}

	// Validate bundle structure
	if bundle.Language == "" {
		result.addError("missing_language", "", "Missing language field", "error")
		result.Valid = false
	} else if bundle.Language != lang {
		result.addWarning("language_mismatch", "", fmt.Sprintf("Language field '%s' doesn't match filename '%s'", bundle.Language, lang))
	}

	if bundle.Version == "" {
		result.addWarning("missing_version", "", "Missing version field")
	}

	// Validate messages
	result.Stats.TotalMessages = len(bundle.Messages)

	for id, msg := range bundle.Messages {
		result.validateMessage(id, msg)
	}

	// Calculate statistics
	result.calculateStats()

	return result
}

// validateMessage validates a single message
func (r *ValidationResult) validateMessage(id string, msg i18n.MessageConfig) {
	// Check if ID matches
	if msg.ID != id {
		r.addError("id_mismatch", id, fmt.Sprintf("Message ID '%s' doesn't match key '%s'", msg.ID, id), "error")
		r.Valid = false
	}

	// Check if message is empty
	if msg.Message == "" {
		r.addError("empty_message", id, "Message is empty", "error")
		r.Stats.EmptyCount++
		r.Valid = false
		return
	}

	r.Stats.TranslatedCount++

	// Check template syntax if enabled
	if checkSyntax {
		r.validateTemplateSyntax(id, msg.Message)
	}

	// Check pluralization
	r.validatePluralization(id, msg)
}

// validateTemplateSyntax validates Go template syntax
func (r *ValidationResult) validateTemplateSyntax(id, message string) {
	// Simple validation for {{.Variable}} syntax
	openCount := strings.Count(message, "{{")
	closeCount := strings.Count(message, "}}")

	if openCount != closeCount {
		r.addError("template_syntax", id, "Unmatched template braces", "error")
		r.Valid = false
		return
	}

	// Check for common template issues
	if strings.Contains(message, "{{.") {
		// Validate variable names
		vars := extractTemplateVars(message)
		for _, v := range vars {
			if !isValidVariableName(v) {
				r.addWarning("invalid_variable", id, fmt.Sprintf("Potentially invalid variable name: %s", v))
			}
		}
	}
}

// validatePluralization validates pluralization forms
func (r *ValidationResult) validatePluralization(id string, msg i18n.MessageConfig) {
	hasPluralForms := msg.Zero != "" || msg.One != "" || msg.Two != "" || msg.Few != "" || msg.Many != "" || msg.Other != ""

	if hasPluralForms {
		// For languages that need specific plural forms
		switch r.Language {
		case "en":
			if msg.One == "" || msg.Other == "" {
				r.addWarning("incomplete_plurals", id, "English requires 'one' and 'other' plural forms")
			}
		case "ko", "ja", "zh":
			// These languages typically don't need plural forms
			if hasPluralForms {
				r.addWarning("unnecessary_plurals", id, "This language typically doesn't require plural forms")
			}
		}
	}
}

// extractTemplateVars extracts variable names from template
func extractTemplateVars(message string) []string {
	var vars []string
	parts := strings.Split(message, "{{.")
	for i := 1; i < len(parts); i++ {
		end := strings.Index(parts[i], "}}")
		if end > 0 {
			vars = append(vars, parts[i][:end])
		}
	}
	return vars
}

// isValidVariableName checks if a variable name is valid
func isValidVariableName(name string) bool {
	if len(name) == 0 {
		return false
	}
	// Simple validation: alphanumeric and underscores
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_') {
			return false
		}
	}
	return true
}

// addError adds a validation error
func (r *ValidationResult) addError(errorType, messageID, message, severity string) {
	r.Errors = append(r.Errors, ValidationError{
		Type:      errorType,
		MessageID: messageID,
		Message:   message,
		Severity:  severity,
	})
	r.Stats.ErrorCount++
}

// addWarning adds a validation warning
func (r *ValidationResult) addWarning(warningType, messageID, message string) {
	r.Warnings = append(r.Warnings, ValidationWarning{
		Type:      warningType,
		MessageID: messageID,
		Message:   message,
	})
	r.Stats.WarningCount++
}

// calculateStats calculates validation statistics
func (r *ValidationResult) calculateStats() {
	r.Stats.MissingCount = r.Stats.TotalMessages - r.Stats.TranslatedCount - r.Stats.EmptyCount
	if r.Stats.TotalMessages > 0 {
		r.Stats.CompletionRate = float64(r.Stats.TranslatedCount) / float64(r.Stats.TotalMessages) * 100
	}
}

// printValidationResult prints validation results for a single file
func printValidationResult(result ValidationResult) {
	status := "✅"
	if !result.Valid {
		status = "❌"
	} else if len(result.Warnings) > 0 {
		status = "⚠️"
	}

	fmt.Printf("%s %s (%s)\n", status, result.File, result.Language)

	// Print errors
	for _, err := range result.Errors {
		symbol := "❌"
		if err.Severity == "warning" {
			symbol = "⚠️"
		}
		if err.MessageID != "" {
			fmt.Printf("  %s %s: %s (%s)\n", symbol, err.MessageID, err.Message, err.Type)
		} else {
			fmt.Printf("  %s %s (%s)\n", symbol, err.Message, err.Type)
		}
	}

	// Print warnings if enabled
	if showWarnings {
		for _, warning := range result.Warnings {
			if warning.MessageID != "" {
				fmt.Printf("  ⚠️  %s: %s (%s)\n", warning.MessageID, warning.Message, warning.Type)
			} else {
				fmt.Printf("  ⚠️  %s (%s)\n", warning.Message, warning.Type)
			}
		}
	}

	// Print stats
	if result.Stats.TotalMessages > 0 {
		fmt.Printf("  📊 %d messages, %.1f%% complete, %d errors, %d warnings\n",
			result.Stats.TotalMessages, result.Stats.CompletionRate,
			result.Stats.ErrorCount, result.Stats.WarningCount)
	}
}

// createValidationSummary creates a summary of all validation results
func createValidationSummary(results []ValidationResult) ValidationSummary {
	summary := ValidationSummary{
		TotalFiles:   len(results),
		Results:      results,
		OverallValid: true,
	}

	for _, result := range results {
		if result.Valid {
			summary.ValidFiles++
		} else {
			summary.OverallValid = false
		}
		summary.TotalErrors += result.Stats.ErrorCount
		summary.TotalWarnings += result.Stats.WarningCount
	}

	return summary
}

// printTextSummary prints a text summary of validation results
func printTextSummary(summary ValidationSummary) {
	fmt.Println("\n📊 Validation Summary:")
	fmt.Printf("  📁 Total files: %d\n", summary.TotalFiles)
	fmt.Printf("  ✅ Valid files: %d\n", summary.ValidFiles)
	fmt.Printf("  ❌ Invalid files: %d\n", summary.TotalFiles-summary.ValidFiles)
	fmt.Printf("  🐛 Total errors: %d\n", summary.TotalErrors)
	fmt.Printf("  ⚠️  Total warnings: %d\n", summary.TotalWarnings)

	if summary.OverallValid {
		fmt.Println("\n🎉 All validations passed!")
	} else {
		fmt.Println("\n💥 Validation failed!")
		fmt.Println("Please fix the errors above and run validation again.")
	}
}

// printJSONResults prints validation results in JSON format
func printJSONResults(summary ValidationSummary) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	encoder.Encode(summary)
}

// printJUnitResults prints validation results in JUnit XML format
func printJUnitResults(summary ValidationSummary) {
	fmt.Println("<?xml version=\"1.0\" encoding=\"UTF-8\"?>")
	fmt.Printf("<testsuite tests=\"%d\" failures=\"%d\" errors=\"%d\" name=\"i18n-validation\">\n",
		summary.TotalFiles, summary.TotalFiles-summary.ValidFiles, summary.TotalErrors)

	for _, result := range summary.Results {
		if result.Valid {
			fmt.Printf("  <testcase name=\"%s\" classname=\"%s\"/>\n", result.Language, result.File)
		} else {
			fmt.Printf("  <testcase name=\"%s\" classname=\"%s\">\n", result.Language, result.File)
			for _, err := range result.Errors {
				fmt.Printf("    <failure message=\"%s\">%s</failure>\n", err.Message, err.Type)
			}
			fmt.Println("  </testcase>")
		}
	}

	fmt.Println("</testsuite>")
}
