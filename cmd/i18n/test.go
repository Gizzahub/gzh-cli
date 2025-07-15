package i18n

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gizzahub/gzh-manager-go/pkg/i18n"
	"github.com/spf13/cobra"
)

// TestCmd represents the test command
var TestCmd = &cobra.Command{
	Use:   "test",
	Short: "Run localization tests and generate reports",
	Long: `Run comprehensive localization tests to validate translation files,
locale formatting, and message consistency across languages.

This command performs various tests including:
- Message existence and completeness
- Template syntax validation
- Pluralization rules checking
- Locale-specific formatting tests
- Cross-language consistency validation

Examples:
  gz i18n test --locales ./locales
  gz i18n test --languages en,ko,ja --report-dir ./reports
  gz i18n test --base-language en --strict`,
	Run: runTest,
}

var (
	testLocalesDir   string
	testLanguages    []string
	testBaseLanguage string
	reportDir        string
	testStrictMode   bool
	generateHTML     bool
	generateJUnit    bool
	testTimeout      int
)

func init() {
	TestCmd.Flags().StringVar(&testLocalesDir, "locales", "locales", "Directory containing locale files")
	TestCmd.Flags().StringSliceVar(&testLanguages, "languages", nil, "Languages to test (default: all)")
	TestCmd.Flags().StringVar(&testBaseLanguage, "base-language", "en", "Base language for comparisons")
	TestCmd.Flags().StringVar(&reportDir, "report-dir", "test-reports", "Directory for test reports")
	TestCmd.Flags().BoolVar(&testStrictMode, "strict", false, "Strict mode - fail on any warning")
	TestCmd.Flags().BoolVar(&generateHTML, "html", true, "Generate HTML report")
	TestCmd.Flags().BoolVar(&generateJUnit, "junit", false, "Generate JUnit XML report")
	TestCmd.Flags().IntVar(&testTimeout, "timeout", 300, "Test timeout in seconds")
}

func runTest(cmd *cobra.Command, args []string) {
	fmt.Println("🧪 Running localization tests...")

	// Check if locales directory exists
	if _, err := os.Stat(testLocalesDir); os.IsNotExist(err) {
		fmt.Printf("❌ Locales directory does not exist: %s\n", testLocalesDir)
		os.Exit(1)
	}

	// Create i18n configuration
	config := &i18n.Config{
		LocalesDir:       testLocalesDir,
		DefaultLanguage:  testBaseLanguage,
		FallbackLanguage: testBaseLanguage,
	}

	// Set supported languages
	if len(testLanguages) > 0 {
		config.SupportedLanguages = testLanguages
	} else {
		// Auto-detect languages from locale files
		languages, err := detectLanguagesFromLocales(testLocalesDir)
		if err != nil {
			fmt.Printf("❌ Failed to detect languages: %v\n", err)
			os.Exit(1)
		}
		config.SupportedLanguages = languages
	}

	fmt.Printf("📁 Locales directory: %s\n", testLocalesDir)
	fmt.Printf("🌐 Testing languages: %v\n", config.SupportedLanguages)
	fmt.Printf("🏠 Base language: %s\n", testBaseLanguage)
	fmt.Printf("📊 Report directory: %s\n", reportDir)

	// Create test suite
	testSuite := i18n.NewTestSuite(config)
	testSuite.SetBaseLanguage(testBaseLanguage)
	testSuite.SetTestLanguages(config.SupportedLanguages)
	testSuite.SetReportDir(reportDir)

	// Add common test messages
	addCommonTestMessages(testSuite)

	// Run tests
	fmt.Println("🚀 Starting test execution...")
	report, err := testSuite.RunAllTests()
	if err != nil {
		fmt.Printf("❌ Failed to run tests: %v\n", err)
		os.Exit(1)
	}

	// Print summary
	fmt.Println("\n📊 Test Results Summary:")
	fmt.Printf("  📁 Total tests: %d\n", report.TotalTests)
	fmt.Printf("  ✅ Passed: %d\n", report.PassedTests)
	fmt.Printf("  ❌ Failed: %d\n", report.FailedTests)
	fmt.Printf("  ⏭️  Skipped: %d\n", report.SkippedTests)
	fmt.Printf("  ⏱️  Duration: %v\n", report.Duration)

	// Print language statistics
	fmt.Println("\n🌐 Language Coverage:")
	for lang, stat := range report.LanguageStats {
		status := "✅"
		if stat.TestsFailed > 0 {
			status = "❌"
		} else if stat.TestsSkipped > 0 {
			status = "⚠️"
		}

		fmt.Printf("  %s %s: %.1f%% coverage (%d/%d messages translated, %d tests passed)\n",
			status, lang, stat.CoveragePercent, stat.TranslatedMessages,
			stat.TotalMessages, stat.TestsPassed)
	}

	// Show failed tests
	if report.FailedTests > 0 {
		fmt.Println("\n❌ Failed Tests:")
		for _, result := range report.Results {
			if result.Status == "fail" {
				fmt.Printf("  [%s] %s: %s\n", result.Language, result.TestName, result.Message)
			}
		}
	}

	// Generate additional reports if requested
	if generateJUnit {
		if err := generateJUnitReport(report, reportDir); err != nil {
			fmt.Printf("⚠️  Failed to generate JUnit report: %v\n", err)
		} else {
			fmt.Println("✅ JUnit XML report generated")
		}
	}

	// Print recommendations
	printRecommendations(report)

	fmt.Printf("\n%s\n", report.Summary)

	// Exit with error code if tests failed and strict mode is enabled
	if testStrictMode && (report.FailedTests > 0 || report.SkippedTests > 0) {
		fmt.Println("\n💥 Exiting with error code due to strict mode")
		os.Exit(1)
	} else if report.FailedTests > 0 {
		fmt.Println("\n💥 Some tests failed")
		os.Exit(1)
	} else {
		fmt.Println("\n🎉 All tests passed!")
	}
}

// detectLanguagesFromLocales detects available languages from locale files
func detectLanguagesFromLocales(localesDir string) ([]string, error) {
	var languages []string

	entries, err := os.ReadDir(localesDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			// Extract language code from filename
			lang := strings.TrimSuffix(entry.Name(), ".json")
			if lang != "config" && lang != "messages" {
				languages = append(languages, lang)
			}
		}
	}

	return languages, nil
}

// addCommonTestMessages adds common test messages to the test suite
func addCommonTestMessages(testSuite *i18n.TestSuite) {
	// Add some common message IDs that should be tested
	commonMessages := map[string]string{
		i18n.MsgWelcome:        "Welcome message should be translated",
		i18n.MsgError:          "Error message should be translated",
		i18n.MsgSuccess:        "Success message should be translated",
		i18n.MsgCloneStarting:  "Clone starting message should be translated",
		i18n.MsgCloneCompleted: "Clone completed message should be translated",
		i18n.MsgCloneStats:     "Clone stats message should be translated",
	}

	for messageID, description := range commonMessages {
		testSuite.AddTestMessage(messageID, description)
	}
}

// generateJUnitReport generates a JUnit XML report
func generateJUnitReport(report *i18n.TestReport, outputDir string) error {
	filename := filepath.Join(outputDir, fmt.Sprintf("junit-i18n-test-%s.xml",
		report.Timestamp.Format("20060102-150405")))

	xml := `<?xml version="1.0" encoding="UTF-8"?>
<testsuite tests="` + fmt.Sprintf("%d", report.TotalTests) + `" 
           failures="` + fmt.Sprintf("%d", report.FailedTests) + `" 
           errors="0" 
           time="` + fmt.Sprintf("%.3f", report.Duration.Seconds()) + `" 
           name="i18n-localization-tests">
`

	// Group results by language and test name
	testCases := make(map[string][]i18n.TestResult)
	for _, result := range report.Results {
		key := fmt.Sprintf("%s.%s", result.Language, result.TestName)
		testCases[key] = append(testCases[key], result)
	}

	for testName, results := range testCases {
		for i, result := range results {
			testCaseName := fmt.Sprintf("%s[%d]", testName, i)
			xml += fmt.Sprintf(`  <testcase classname="%s" name="%s" time="%.3f"`,
				result.Language, testCaseName, result.Duration.Seconds())

			if result.Status == "fail" {
				xml += ">\n"
				xml += fmt.Sprintf(`    <failure message="%s">%s</failure>`,
					result.Message, result.Message)
				xml += "\n  </testcase>\n"
			} else if result.Status == "skip" {
				xml += ">\n"
				xml += fmt.Sprintf(`    <skipped message="%s"/>`, result.Message)
				xml += "\n  </testcase>\n"
			} else {
				xml += "/>\n"
			}
		}
	}

	xml += "</testsuite>\n"

	return os.WriteFile(filename, []byte(xml), 0o644)
}

// printRecommendations prints recommendations based on test results
func printRecommendations(report *i18n.TestReport) {
	fmt.Println("\n💡 Recommendations:")

	if report.FailedTests > 0 {
		fmt.Println("  🔧 Fix failed tests by:")
		fmt.Println("    - Adding missing translations")
		fmt.Println("    - Fixing template syntax errors")
		fmt.Println("    - Completing empty messages")
		fmt.Println("    - Adding required plural forms")
	}

	// Check coverage
	var lowCoverageLanguages []string
	for lang, stat := range report.LanguageStats {
		if stat.CoveragePercent < 80.0 {
			lowCoverageLanguages = append(lowCoverageLanguages,
				fmt.Sprintf("%s (%.1f%%)", lang, stat.CoveragePercent))
		}
	}

	if len(lowCoverageLanguages) > 0 {
		fmt.Println("  📈 Improve translation coverage for:")
		for _, lang := range lowCoverageLanguages {
			fmt.Printf("    - %s\n", lang)
		}
	}

	if report.SkippedTests > 0 {
		fmt.Println("  ⚠️  Review skipped tests - they may indicate missing files or configuration issues")
	}

	fmt.Println("  📚 Consider:")
	fmt.Println("    - Regular translation updates")
	fmt.Println("    - Automated testing in CI/CD pipeline")
	fmt.Println("    - Translator review process")
}
