package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

// TestSuite provides comprehensive testing for internationalization
type TestSuite struct {
	config        *Config
	baseLanguage  string
	testLanguages []string
	localesDir    string
	testMessages  map[string]string
	reportDir     string
}

// TestResult holds the result of a localization test
type TestResult struct {
	Language      string            `json:"language"`
	TestName      string            `json:"test_name"`
	Status        string            `json:"status"` // "pass", "fail", "skip"
	Message       string            `json:"message,omitempty"`
	ExpectedValue interface{}       `json:"expected_value,omitempty"`
	ActualValue   interface{}       `json:"actual_value,omitempty"`
	Duration      time.Duration     `json:"duration"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

// TestReport holds a comprehensive test report
type TestReport struct {
	Timestamp     time.Time               `json:"timestamp"`
	TotalTests    int                     `json:"total_tests"`
	PassedTests   int                     `json:"passed_tests"`
	FailedTests   int                     `json:"failed_tests"`
	SkippedTests  int                     `json:"skipped_tests"`
	Duration      time.Duration           `json:"duration"`
	Results       []TestResult            `json:"results"`
	LanguageStats map[string]LanguageStat `json:"language_stats"`
	Summary       string                  `json:"summary"`
}

// LanguageStat holds statistics for a specific language
type LanguageStat struct {
	Language           string  `json:"language"`
	TotalMessages      int     `json:"total_messages"`
	TranslatedMessages int     `json:"translated_messages"`
	CoveragePercent    float64 `json:"coverage_percent"`
	TestsPassed        int     `json:"tests_passed"`
	TestsFailed        int     `json:"tests_failed"`
	TestsSkipped       int     `json:"tests_skipped"`
}

// NewTestSuite creates a new localization test suite
func NewTestSuite(config *Config) *TestSuite {
	return &TestSuite{
		config:        config,
		baseLanguage:  config.DefaultLanguage,
		testLanguages: config.SupportedLanguages,
		localesDir:    config.LocalesDir,
		testMessages:  make(map[string]string),
		reportDir:     "test-reports",
	}
}

// SetBaseLanguage sets the base language for comparisons
func (ts *TestSuite) SetBaseLanguage(lang string) {
	ts.baseLanguage = lang
}

// SetTestLanguages sets the languages to test
func (ts *TestSuite) SetTestLanguages(languages []string) {
	ts.testLanguages = languages
}

// SetReportDir sets the directory for test reports
func (ts *TestSuite) SetReportDir(dir string) {
	ts.reportDir = dir
}

// AddTestMessage adds a message to test
func (ts *TestSuite) AddTestMessage(messageID, expectedTranslation string) {
	ts.testMessages[messageID] = expectedTranslation
}

// RunAllTests runs all localization tests
func (ts *TestSuite) RunAllTests() (*TestReport, error) {
	startTime := time.Now()

	report := &TestReport{
		Timestamp:     startTime,
		Results:       []TestResult{},
		LanguageStats: make(map[string]LanguageStat),
	}

	// Create report directory
	if err := os.MkdirAll(ts.reportDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create report directory: %w", err)
	}

	// Run tests for each language
	for _, lang := range ts.testLanguages {
		langResults, err := ts.runLanguageTests(lang)
		if err != nil {
			fmt.Printf("Warning: Failed to run tests for language %s: %v\n", lang, err)
			continue
		}

		report.Results = append(report.Results, langResults...)
		report.LanguageStats[lang] = ts.calculateLanguageStats(lang, langResults)
	}

	// Calculate totals
	report.Duration = time.Since(startTime)
	for _, result := range report.Results {
		report.TotalTests++
		switch result.Status {
		case "pass":
			report.PassedTests++
		case "fail":
			report.FailedTests++
		case "skip":
			report.SkippedTests++
		}
	}

	// Generate summary
	report.Summary = ts.generateSummary(report)

	// Save report
	if err := ts.saveReport(report); err != nil {
		return report, fmt.Errorf("failed to save report: %w", err)
	}

	return report, nil
}

// runLanguageTests runs tests for a specific language
func (ts *TestSuite) runLanguageTests(lang string) ([]TestResult, error) {
	var results []TestResult

	// Load language bundle
	bundle, err := ts.loadLanguageBundle(lang)
	if err != nil {
		return nil, fmt.Errorf("failed to load bundle for %s: %w", lang, err)
	}

	// Test message existence
	results = append(results, ts.testMessageExistence(lang, bundle)...)

	// Test message completeness
	results = append(results, ts.testMessageCompleteness(lang, bundle)...)

	// Test template syntax
	results = append(results, ts.testTemplateSyntax(lang, bundle)...)

	// Test pluralization
	results = append(results, ts.testPluralization(lang, bundle)...)

	// Test locale formatting
	results = append(results, ts.testLocaleFormatting(lang)...)

	// Test message consistency
	results = append(results, ts.testMessageConsistency(lang, bundle)...)

	return results, nil
}

// loadLanguageBundle loads a language bundle for testing
func (ts *TestSuite) loadLanguageBundle(lang string) (*LocalizationBundle, error) {
	filename := filepath.Join(ts.localesDir, fmt.Sprintf("%s.json", lang))

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var bundle LocalizationBundle
	if err := json.Unmarshal(data, &bundle); err != nil {
		return nil, err
	}

	return &bundle, nil
}

// testMessageExistence tests if all required messages exist
func (ts *TestSuite) testMessageExistence(lang string, bundle *LocalizationBundle) []TestResult {
	var results []TestResult

	// Get base language messages for comparison
	baseBundle, err := ts.loadLanguageBundle(ts.baseLanguage)
	if err != nil {
		results = append(results, TestResult{
			Language: lang,
			TestName: "message_existence",
			Status:   "fail",
			Message:  fmt.Sprintf("Failed to load base language %s: %v", ts.baseLanguage, err),
			Duration: 0,
		})
		return results
	}

	// Check if all base messages exist in target language
	for messageID := range baseBundle.Messages {
		start := time.Now()

		if _, exists := bundle.Messages[messageID]; exists {
			results = append(results, TestResult{
				Language: lang,
				TestName: "message_existence",
				Status:   "pass",
				Message:  fmt.Sprintf("Message %s exists", messageID),
				Duration: time.Since(start),
				Metadata: map[string]string{"message_id": messageID},
			})
		} else {
			results = append(results, TestResult{
				Language: lang,
				TestName: "message_existence",
				Status:   "fail",
				Message:  fmt.Sprintf("Missing message: %s", messageID),
				Duration: time.Since(start),
				Metadata: map[string]string{"message_id": messageID},
			})
		}
	}

	return results
}

// testMessageCompleteness tests if messages are not empty
func (ts *TestSuite) testMessageCompleteness(lang string, bundle *LocalizationBundle) []TestResult {
	var results []TestResult

	for messageID, msg := range bundle.Messages {
		start := time.Now()

		if strings.TrimSpace(msg.Message) == "" {
			results = append(results, TestResult{
				Language: lang,
				TestName: "message_completeness",
				Status:   "fail",
				Message:  fmt.Sprintf("Empty message: %s", messageID),
				Duration: time.Since(start),
				Metadata: map[string]string{"message_id": messageID},
			})
		} else {
			results = append(results, TestResult{
				Language: lang,
				TestName: "message_completeness",
				Status:   "pass",
				Message:  fmt.Sprintf("Message %s is complete", messageID),
				Duration: time.Since(start),
				Metadata: map[string]string{"message_id": messageID},
			})
		}
	}

	return results
}

// testTemplateSyntax tests template syntax validity
func (ts *TestSuite) testTemplateSyntax(lang string, bundle *LocalizationBundle) []TestResult {
	var results []TestResult

	for messageID, msg := range bundle.Messages {
		start := time.Now()

		// Check for balanced braces
		openCount := strings.Count(msg.Message, "{{")
		closeCount := strings.Count(msg.Message, "}}")

		if openCount != closeCount {
			results = append(results, TestResult{
				Language: lang,
				TestName: "template_syntax",
				Status:   "fail",
				Message:  fmt.Sprintf("Unbalanced template braces in message: %s", messageID),
				Duration: time.Since(start),
				Metadata: map[string]string{
					"message_id":   messageID,
					"open_braces":  fmt.Sprintf("%d", openCount),
					"close_braces": fmt.Sprintf("%d", closeCount),
				},
			})
		} else {
			results = append(results, TestResult{
				Language: lang,
				TestName: "template_syntax",
				Status:   "pass",
				Message:  fmt.Sprintf("Template syntax valid for message: %s", messageID),
				Duration: time.Since(start),
				Metadata: map[string]string{"message_id": messageID},
			})
		}
	}

	return results
}

// testPluralization tests pluralization rules
func (ts *TestSuite) testPluralization(lang string, bundle *LocalizationBundle) []TestResult {
	var results []TestResult

	// Get language-specific pluralization requirements
	requirements := ts.getPluralizationRequirements(lang)

	for messageID, msg := range bundle.Messages {
		start := time.Now()

		// Check if message has plural forms
		hasPluralForms := msg.Zero != "" || msg.One != "" || msg.Two != "" ||
			msg.Few != "" || msg.Many != "" || msg.Other != ""

		if hasPluralForms {
			// Validate required plural forms for this language
			missing := ts.checkRequiredPluralForms(lang, msg, requirements)

			if len(missing) > 0 {
				results = append(results, TestResult{
					Language: lang,
					TestName: "pluralization",
					Status:   "fail",
					Message:  fmt.Sprintf("Missing plural forms for %s: %s", messageID, strings.Join(missing, ", ")),
					Duration: time.Since(start),
					Metadata: map[string]string{
						"message_id":    messageID,
						"missing_forms": strings.Join(missing, ","),
					},
				})
			} else {
				results = append(results, TestResult{
					Language: lang,
					TestName: "pluralization",
					Status:   "pass",
					Message:  fmt.Sprintf("Pluralization valid for message: %s", messageID),
					Duration: time.Since(start),
					Metadata: map[string]string{"message_id": messageID},
				})
			}
		}
	}

	return results
}

// testLocaleFormatting tests locale-specific formatting
func (ts *TestSuite) testLocaleFormatting(lang string) []TestResult {
	var results []TestResult

	// Test locale manager creation
	start := time.Now()
	localeManager, err := NewLocaleManager(lang)
	if err != nil {
		results = append(results, TestResult{
			Language: lang,
			TestName: "locale_formatting",
			Status:   "fail",
			Message:  fmt.Sprintf("Failed to create locale manager: %v", err),
			Duration: time.Since(start),
		})
		return results
	}

	// Test time formatting
	testTime := time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC)

	timeFormats := []string{"short", "long", "full"}
	for _, format := range timeFormats {
		start := time.Now()
		formatted := localeManager.FormatTime(testTime, format)

		if formatted != "" {
			results = append(results, TestResult{
				Language:    lang,
				TestName:    "locale_formatting",
				Status:      "pass",
				Message:     fmt.Sprintf("Time formatting (%s) works", format),
				ActualValue: formatted,
				Duration:    time.Since(start),
				Metadata:    map[string]string{"format_type": "time", "format": format},
			})
		} else {
			results = append(results, TestResult{
				Language: lang,
				TestName: "locale_formatting",
				Status:   "fail",
				Message:  fmt.Sprintf("Time formatting (%s) returned empty", format),
				Duration: time.Since(start),
				Metadata: map[string]string{"format_type": "time", "format": format},
			})
		}
	}

	// Test number formatting
	start = time.Now()
	formattedNumber := localeManager.FormatNumber(1234567.89)
	if formattedNumber != "" {
		results = append(results, TestResult{
			Language:    lang,
			TestName:    "locale_formatting",
			Status:      "pass",
			Message:     "Number formatting works",
			ActualValue: formattedNumber,
			Duration:    time.Since(start),
			Metadata:    map[string]string{"format_type": "number"},
		})
	} else {
		results = append(results, TestResult{
			Language: lang,
			TestName: "locale_formatting",
			Status:   "fail",
			Message:  "Number formatting returned empty",
			Duration: time.Since(start),
			Metadata: map[string]string{"format_type": "number"},
		})
	}

	return results
}

// testMessageConsistency tests consistency between languages
func (ts *TestSuite) testMessageConsistency(lang string, bundle *LocalizationBundle) []TestResult {
	var results []TestResult

	if lang == ts.baseLanguage {
		return results // Skip consistency check for base language
	}

	baseBundle, err := ts.loadLanguageBundle(ts.baseLanguage)
	if err != nil {
		return results
	}

	for messageID, baseMsg := range baseBundle.Messages {
		start := time.Now()

		if targetMsg, exists := bundle.Messages[messageID]; exists {
			// Check template variable consistency
			baseVars := extractTemplateVariables(baseMsg.Message)
			targetVars := extractTemplateVariables(targetMsg.Message)

			if !areVariablesConsistent(baseVars, targetVars) {
				results = append(results, TestResult{
					Language:      lang,
					TestName:      "message_consistency",
					Status:        "fail",
					Message:       fmt.Sprintf("Template variables mismatch for %s", messageID),
					ExpectedValue: baseVars,
					ActualValue:   targetVars,
					Duration:      time.Since(start),
					Metadata:      map[string]string{"message_id": messageID},
				})
			} else {
				results = append(results, TestResult{
					Language: lang,
					TestName: "message_consistency",
					Status:   "pass",
					Message:  fmt.Sprintf("Template variables consistent for %s", messageID),
					Duration: time.Since(start),
					Metadata: map[string]string{"message_id": messageID},
				})
			}
		}
	}

	return results
}

// Helper functions

func (ts *TestSuite) getPluralizationRequirements(lang string) []string {
	switch lang {
	case "en", "en-US":
		return []string{"one", "other"}
	case "ko", "ja", "zh", "zh-CN", "zh-TW":
		return []string{} // No plural forms typically required
	case "ru":
		return []string{"one", "few", "many", "other"}
	case "pl":
		return []string{"one", "few", "many", "other"}
	default:
		return []string{"other"}
	}
}

func (ts *TestSuite) checkRequiredPluralForms(lang string, msg MessageConfig, required []string) []string {
	var missing []string

	pluralForms := map[string]string{
		"zero":  msg.Zero,
		"one":   msg.One,
		"two":   msg.Two,
		"few":   msg.Few,
		"many":  msg.Many,
		"other": msg.Other,
	}

	for _, form := range required {
		if pluralForms[form] == "" {
			missing = append(missing, form)
		}
	}

	return missing
}

func extractTemplateVariables(message string) []string {
	var variables []string
	parts := strings.Split(message, "{{.")

	for i := 1; i < len(parts); i++ {
		end := strings.Index(parts[i], "}}")
		if end > 0 {
			variable := strings.TrimSpace(parts[i][:end])
			variables = append(variables, variable)
		}
	}

	sort.Strings(variables)
	return variables
}

func areVariablesConsistent(expected, actual []string) bool {
	if len(expected) != len(actual) {
		return false
	}

	for i, v := range expected {
		if v != actual[i] {
			return false
		}
	}

	return true
}

func (ts *TestSuite) calculateLanguageStats(lang string, results []TestResult) LanguageStat {
	stat := LanguageStat{
		Language: lang,
	}

	for _, result := range results {
		switch result.Status {
		case "pass":
			stat.TestsPassed++
		case "fail":
			stat.TestsFailed++
		case "skip":
			stat.TestsSkipped++
		}
	}

	// Calculate coverage if possible
	bundle, err := ts.loadLanguageBundle(lang)
	if err == nil {
		stat.TotalMessages = len(bundle.Messages)

		translatedCount := 0
		for _, msg := range bundle.Messages {
			if strings.TrimSpace(msg.Message) != "" {
				translatedCount++
			}
		}

		stat.TranslatedMessages = translatedCount
		if stat.TotalMessages > 0 {
			stat.CoveragePercent = float64(translatedCount) / float64(stat.TotalMessages) * 100
		}
	}

	return stat
}

func (ts *TestSuite) generateSummary(report *TestReport) string {
	successRate := float64(report.PassedTests) / float64(report.TotalTests) * 100

	summary := fmt.Sprintf("Localization test completed in %v. ", report.Duration)
	summary += fmt.Sprintf("Success rate: %.1f%% (%d/%d tests passed). ",
		successRate, report.PassedTests, report.TotalTests)

	if report.FailedTests > 0 {
		summary += fmt.Sprintf("%d tests failed. ", report.FailedTests)
	}

	if report.SkippedTests > 0 {
		summary += fmt.Sprintf("%d tests skipped. ", report.SkippedTests)
	}

	// Add language-specific summary
	var languageSummaries []string
	for lang, stat := range report.LanguageStats {
		langSummary := fmt.Sprintf("%s: %.1f%% coverage", lang, stat.CoveragePercent)
		languageSummaries = append(languageSummaries, langSummary)
	}

	if len(languageSummaries) > 0 {
		summary += "Languages: " + strings.Join(languageSummaries, ", ")
	}

	return summary
}

func (ts *TestSuite) saveReport(report *TestReport) error {
	// Save JSON report
	jsonFile := filepath.Join(ts.reportDir, fmt.Sprintf("i18n-test-report-%s.json",
		report.Timestamp.Format("20060102-150405")))

	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(jsonFile, jsonData, 0o644); err != nil {
		return err
	}

	// Save HTML report
	htmlFile := filepath.Join(ts.reportDir, fmt.Sprintf("i18n-test-report-%s.html",
		report.Timestamp.Format("20060102-150405")))

	htmlContent := ts.generateHTMLReport(report)
	if err := os.WriteFile(htmlFile, []byte(htmlContent), 0o644); err != nil {
		return err
	}

	fmt.Printf("Test reports saved:\n")
	fmt.Printf("  JSON: %s\n", jsonFile)
	fmt.Printf("  HTML: %s\n", htmlFile)

	return nil
}

func (ts *TestSuite) generateHTMLReport(report *TestReport) string {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Localization Test Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background: #f5f5f5; padding: 20px; border-radius: 5px; }
        .summary { margin: 20px 0; }
        .stats { display: flex; gap: 20px; margin: 20px 0; }
        .stat-box { background: #e7f3ff; padding: 15px; border-radius: 5px; text-align: center; }
        .pass { color: green; }
        .fail { color: red; }
        .skip { color: orange; }
        table { width: 100%; border-collapse: collapse; margin: 20px 0; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
        .language-section { margin: 30px 0; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Localization Test Report</h1>
        <p><strong>Generated:</strong> ` + report.Timestamp.Format("2006-01-02 15:04:05") + `</p>
        <p><strong>Duration:</strong> ` + report.Duration.String() + `</p>
    </div>

    <div class="summary">
        <h2>Summary</h2>
        <p>` + report.Summary + `</p>
    </div>

    <div class="stats">
        <div class="stat-box">
            <h3>` + fmt.Sprintf("%d", report.TotalTests) + `</h3>
            <p>Total Tests</p>
        </div>
        <div class="stat-box pass">
            <h3>` + fmt.Sprintf("%d", report.PassedTests) + `</h3>
            <p>Passed</p>
        </div>
        <div class="stat-box fail">
            <h3>` + fmt.Sprintf("%d", report.FailedTests) + `</h3>
            <p>Failed</p>
        </div>
        <div class="stat-box skip">
            <h3>` + fmt.Sprintf("%d", report.SkippedTests) + `</h3>
            <p>Skipped</p>
        </div>
    </div>`

	// Add language statistics
	html += `<div class="language-section">
        <h2>Language Statistics</h2>
        <table>
            <tr>
                <th>Language</th>
                <th>Total Messages</th>
                <th>Translated</th>
                <th>Coverage</th>
                <th>Tests Passed</th>
                <th>Tests Failed</th>
            </tr>`

	for lang, stat := range report.LanguageStats {
		html += fmt.Sprintf(`
            <tr>
                <td>%s</td>
                <td>%d</td>
                <td>%d</td>
                <td>%.1f%%</td>
                <td class="pass">%d</td>
                <td class="fail">%d</td>
            </tr>`, lang, stat.TotalMessages, stat.TranslatedMessages,
			stat.CoveragePercent, stat.TestsPassed, stat.TestsFailed)
	}

	html += `    </table>
    </div>

    <div class="language-section">
        <h2>Test Results</h2>
        <table>
            <tr>
                <th>Language</th>
                <th>Test Name</th>
                <th>Status</th>
                <th>Message</th>
                <th>Duration</th>
            </tr>`

	for _, result := range report.Results {
		statusClass := result.Status
		html += fmt.Sprintf(`
            <tr>
                <td>%s</td>
                <td>%s</td>
                <td class="%s">%s</td>
                <td>%s</td>
                <td>%s</td>
            </tr>`, result.Language, result.TestName, statusClass,
			strings.ToUpper(result.Status), result.Message, result.Duration)
	}

	html += `    </table>
    </div>
</body>
</html>`

	return html
}

// RunTestsForTesting is a helper function for Go testing framework
func RunTestsForTesting(t *testing.T, config *Config) {
	testSuite := NewTestSuite(config)

	report, err := testSuite.RunAllTests()
	if err != nil {
		t.Fatalf("Failed to run localization tests: %v", err)
	}

	// Assert test results
	if report.FailedTests > 0 {
		t.Errorf("Localization tests failed: %d out of %d tests failed",
			report.FailedTests, report.TotalTests)

		// Log failed tests
		for _, result := range report.Results {
			if result.Status == "fail" {
				t.Logf("FAIL [%s] %s: %s", result.Language, result.TestName, result.Message)
			}
		}
	}

	t.Logf("Localization test summary: %s", report.Summary)
}
