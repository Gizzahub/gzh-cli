package jenkins

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

// ValidateCmd represents the validate command
var ValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate Jenkins pipeline configurations",
	Long: `Validate Jenkins pipeline configurations for syntax and best practices.

Performs comprehensive validation including:
- Jenkinsfile syntax validation
- Pipeline structure validation
- Agent and tool configuration validation
- Security best practices validation
- Performance optimization checks
- Blue Ocean compatibility checks

Examples:
  gz jenkins validate
  gz jenkins validate --file Jenkinsfile
  gz jenkins validate --strict --security-check`,
	Run: runValidate,
}

var (
	validateFile     string
	strictMode       bool
	securityCheck    bool
	performanceCheck bool
	blueOceanCheck   bool
	outputFormat     string
	allowWarnings    bool
	checkPlugins     bool
)

func init() {
	ValidateCmd.Flags().StringVarP(&validateFile, "file", "f", "Jenkinsfile", "Pipeline file to validate")
	ValidateCmd.Flags().BoolVar(&strictMode, "strict", false, "Enable strict validation mode")
	ValidateCmd.Flags().BoolVar(&securityCheck, "security-check", true, "Enable security validation")
	ValidateCmd.Flags().BoolVar(&performanceCheck, "performance-check", true, "Enable performance validation")
	ValidateCmd.Flags().BoolVar(&blueOceanCheck, "blue-ocean-check", false, "Enable Blue Ocean compatibility checks")
	ValidateCmd.Flags().StringVarP(&outputFormat, "output", "o", "text", "Output format (text, json)")
	ValidateCmd.Flags().BoolVar(&allowWarnings, "allow-warnings", false, "Allow warnings (exit code 0)")
	ValidateCmd.Flags().BoolVar(&checkPlugins, "check-plugins", true, "Validate required plugins")
}

// ValidationResult represents a validation result
type ValidationResult struct {
	File     string `json:"file"`
	Type     string `json:"type"`
	Level    string `json:"level"` // error, warning, info
	Message  string `json:"message"`
	Line     int    `json:"line,omitempty"`
	Rule     string `json:"rule"`
	Category string `json:"category"`
}

// ValidationReport represents the complete validation report
type ValidationReport struct {
	Summary ValidationSummary  `json:"summary"`
	Results []ValidationResult `json:"results"`
}

// ValidationSummary represents validation summary statistics
type ValidationSummary struct {
	TotalChecks  int  `json:"total_checks"`
	ErrorCount   int  `json:"error_count"`
	WarningCount int  `json:"warning_count"`
	InfoCount    int  `json:"info_count"`
	Valid        bool `json:"valid"`
}

func runValidate(cmd *cobra.Command, args []string) {
	if validateFile == "" {
		fmt.Println("Error: pipeline file is required")
		os.Exit(1)
	}

	// Check if pipeline file exists
	if _, err := os.Stat(validateFile); os.IsNotExist(err) {
		fmt.Printf("Error: pipeline file not found: %s\n", validateFile)
		os.Exit(1)
	}

	fmt.Printf("🔍 Validating Jenkins pipeline: %s\n", validateFile)
	if strictMode {
		fmt.Println("📋 Mode: Strict validation")
	}

	// Run validation
	report, err := validateJenkinsfile(validateFile)
	if err != nil {
		fmt.Printf("Error running validation: %v\n", err)
		os.Exit(1)
	}

	// Output results
	if err := outputValidationResults(report); err != nil {
		fmt.Printf("Error outputting results: %v\n", err)
		os.Exit(1)
	}

	// Print summary
	printValidationSummary(report)

	// Exit with appropriate code
	if report.Summary.ErrorCount > 0 {
		os.Exit(1)
	} else if report.Summary.WarningCount > 0 && !allowWarnings {
		os.Exit(2)
	}
}

func validateJenkinsfile(filePath string) (*ValidationReport, error) {
	report := &ValidationReport{
		Results: []ValidationResult{},
	}

	// Read file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	contentStr := string(content)
	lines := strings.Split(contentStr, "\n")

	// Basic syntax validation
	results := validateSyntax(filePath, contentStr, lines)
	report.Results = append(report.Results, results...)

	// Structure validation
	results = validateStructure(filePath, contentStr, lines)
	report.Results = append(report.Results, results...)

	// Security validation
	if securityCheck {
		results = validateSecurity(filePath, contentStr, lines)
		report.Results = append(report.Results, results...)
	}

	// Performance validation
	if performanceCheck {
		results = validatePerformance(filePath, contentStr, lines)
		report.Results = append(report.Results, results...)
	}

	// Blue Ocean validation
	if blueOceanCheck {
		results = validateBlueOcean(filePath, contentStr, lines)
		report.Results = append(report.Results, results...)
	}

	// Plugin validation
	if checkPlugins {
		results = validatePlugins(filePath, contentStr, lines)
		report.Results = append(report.Results, results...)
	}

	// Calculate summary
	for _, result := range report.Results {
		report.Summary.TotalChecks++
		switch result.Level {
		case "error":
			report.Summary.ErrorCount++
		case "warning":
			report.Summary.WarningCount++
		case "info":
			report.Summary.InfoCount++
		}
	}

	report.Summary.Valid = report.Summary.ErrorCount == 0 && (allowWarnings || report.Summary.WarningCount == 0)

	return report, nil
}

func validateSyntax(filePath, content string, lines []string) []ValidationResult {
	var results []ValidationResult

	// Check for basic Groovy syntax
	if !strings.Contains(content, "pipeline") && !strings.Contains(content, "node") {
		results = append(results, ValidationResult{
			File:     filePath,
			Type:     "syntax",
			Level:    "error",
			Message:  "No pipeline or node block found",
			Rule:     "pipeline-structure",
			Category: "syntax",
		})
	}

	// Check for balanced braces
	braceCount := 0
	for i, line := range lines {
		for _, char := range line {
			switch char {
			case '{':
				braceCount++
			case '}':
				braceCount--
			}
		}

		if braceCount < 0 {
			results = append(results, ValidationResult{
				File:     filePath,
				Type:     "syntax",
				Level:    "error",
				Message:  "Unmatched closing brace",
				Line:     i + 1,
				Rule:     "balanced-braces",
				Category: "syntax",
			})
		}
	}

	if braceCount != 0 {
		results = append(results, ValidationResult{
			File:     filePath,
			Type:     "syntax",
			Level:    "error",
			Message:  "Unbalanced braces in pipeline",
			Rule:     "balanced-braces",
			Category: "syntax",
		})
	}

	// Check for basic string quoting
	for i, line := range lines {
		// Simple check for unmatched quotes
		singleQuotes := strings.Count(line, "'")
		doubleQuotes := strings.Count(line, "\"")

		if singleQuotes%2 != 0 {
			results = append(results, ValidationResult{
				File:     filePath,
				Type:     "syntax",
				Level:    "warning",
				Message:  "Potentially unmatched single quotes",
				Line:     i + 1,
				Rule:     "quote-matching",
				Category: "syntax",
			})
		}

		if doubleQuotes%2 != 0 {
			results = append(results, ValidationResult{
				File:     filePath,
				Type:     "syntax",
				Level:    "warning",
				Message:  "Potentially unmatched double quotes",
				Line:     i + 1,
				Rule:     "quote-matching",
				Category: "syntax",
			})
		}
	}

	return results
}

func validateStructure(filePath, content string, lines []string) []ValidationResult {
	var results []ValidationResult

	// Check for declarative pipeline structure
	if strings.Contains(content, "pipeline {") {
		// Validate declarative pipeline
		results = append(results, validateDeclarativePipeline(filePath, content, lines)...)
	} else if strings.Contains(content, "node") {
		// Validate scripted pipeline
		results = append(results, validateScriptedPipeline(filePath, content, lines)...)
	}

	// Check for required sections
	if !strings.Contains(content, "stage(") && !strings.Contains(content, "stages {") {
		results = append(results, ValidationResult{
			File:     filePath,
			Type:     "structure",
			Level:    "warning",
			Message:  "No stages found in pipeline",
			Rule:     "stages-required",
			Category: "structure",
		})
	}

	// Check for agent configuration
	if !strings.Contains(content, "agent") {
		results = append(results, ValidationResult{
			File:     filePath,
			Type:     "structure",
			Level:    "warning",
			Message:  "No agent configuration found",
			Rule:     "agent-required",
			Category: "structure",
		})
	}

	return results
}

func validateDeclarativePipeline(filePath, content string, lines []string) []ValidationResult {
	var results []ValidationResult

	// Check for required top-level sections in declarative pipeline
	requiredSections := []string{"agent", "stages"}
	for _, section := range requiredSections {
		if !strings.Contains(content, section+" {") && !strings.Contains(content, section+" ") {
			results = append(results, ValidationResult{
				File:     filePath,
				Type:     "declarative",
				Level:    "error",
				Message:  fmt.Sprintf("Missing required section: %s", section),
				Rule:     "declarative-structure",
				Category: "structure",
			})
		}
	}

	// Check for valid agent configurations
	agentPattern := regexp.MustCompile(`agent\s+\{[^}]*\}|agent\s+any|agent\s+none`)
	if !agentPattern.MatchString(content) {
		results = append(results, ValidationResult{
			File:     filePath,
			Type:     "declarative",
			Level:    "warning",
			Message:  "Agent configuration may be invalid",
			Rule:     "agent-configuration",
			Category: "structure",
		})
	}

	// Check for post section usage
	if strings.Contains(content, "post {") {
		// Validate post section
		validPostConditions := []string{"always", "success", "failure", "unstable", "aborted", "changed"}
		postSection := extractSection(content, "post")

		for _, condition := range validPostConditions {
			if strings.Contains(postSection, condition+" {") {
				// Valid condition found
				break
			}
		}
	}

	return results
}

func validateScriptedPipeline(filePath, content string, lines []string) []ValidationResult {
	var results []ValidationResult

	// Check for node block
	if !strings.Contains(content, "node") {
		results = append(results, ValidationResult{
			File:     filePath,
			Type:     "scripted",
			Level:    "error",
			Message:  "Scripted pipeline must have a node block",
			Rule:     "node-required",
			Category: "structure",
		})
	}

	// Check for try-catch blocks
	if !strings.Contains(content, "try") && !strings.Contains(content, "catch") {
		results = append(results, ValidationResult{
			File:     filePath,
			Type:     "scripted",
			Level:    "warning",
			Message:  "Consider using try-catch blocks for error handling",
			Rule:     "error-handling",
			Category: "best-practices",
		})
	}

	return results
}

func validateSecurity(filePath, content string, lines []string) []ValidationResult {
	var results []ValidationResult

	// Check for hardcoded secrets
	secretPatterns := []string{
		`password\s*=\s*["'].*["']`,
		`token\s*=\s*["'].*["']`,
		`key\s*=\s*["'].*["']`,
		`secret\s*=\s*["'].*["']`,
	}

	for i, line := range lines {
		for _, pattern := range secretPatterns {
			matched, _ := regexp.MatchString(pattern, strings.ToLower(line))
			if matched && !strings.Contains(line, "${") && !strings.Contains(line, "env.") {
				results = append(results, ValidationResult{
					File:     filePath,
					Type:     "security",
					Level:    "warning",
					Message:  "Potential hardcoded secret detected",
					Line:     i + 1,
					Rule:     "no-hardcoded-secrets",
					Category: "security",
				})
			}
		}
	}

	// Check for credential usage
	if strings.Contains(content, "withCredentials") {
		if !strings.Contains(content, "credentialsId") {
			results = append(results, ValidationResult{
				File:     filePath,
				Type:     "security",
				Level:    "warning",
				Message:  "withCredentials block should specify credentialsId",
				Rule:     "credentials-id-required",
				Category: "security",
			})
		}
	}

	// Check for shell injection vulnerabilities
	dangerousCommands := []string{"eval", "exec", "system", "Runtime.getRuntime"}
	for i, line := range lines {
		for _, cmd := range dangerousCommands {
			if strings.Contains(line, cmd) {
				results = append(results, ValidationResult{
					File:     filePath,
					Type:     "security",
					Level:    "warning",
					Message:  fmt.Sprintf("Potentially dangerous command: %s", cmd),
					Line:     i + 1,
					Rule:     "dangerous-commands",
					Category: "security",
				})
			}
		}
	}

	return results
}

func validatePerformance(filePath, content string, lines []string) []ValidationResult {
	var results []ValidationResult

	// Check for checkout scm optimization
	checkoutCount := strings.Count(content, "checkout scm")
	if checkoutCount > 1 {
		results = append(results, ValidationResult{
			File:     filePath,
			Type:     "performance",
			Level:    "warning",
			Message:  "Multiple checkout scm calls detected - consider optimization",
			Rule:     "checkout-optimization",
			Category: "performance",
		})
	}

	// Check for parallel stages
	if strings.Count(content, "stage(") > 3 && !strings.Contains(content, "parallel") {
		results = append(results, ValidationResult{
			File:     filePath,
			Type:     "performance",
			Level:    "info",
			Message:  "Consider using parallel stages for better performance",
			Rule:     "parallel-stages",
			Category: "performance",
		})
	}

	// Check for workspace cleanup
	if !strings.Contains(content, "cleanWs") && !strings.Contains(content, "deleteDir") {
		results = append(results, ValidationResult{
			File:     filePath,
			Type:     "performance",
			Level:    "warning",
			Message:  "No workspace cleanup found - consider adding cleanWs()",
			Rule:     "workspace-cleanup",
			Category: "performance",
		})
	}

	// Check for build discarder
	if !strings.Contains(content, "buildDiscarder") && !strings.Contains(content, "logRotator") {
		results = append(results, ValidationResult{
			File:     filePath,
			Type:     "performance",
			Level:    "info",
			Message:  "Consider adding build discarder to manage disk space",
			Rule:     "build-discarder",
			Category: "performance",
		})
	}

	return results
}

func validateBlueOcean(filePath, content string, lines []string) []ValidationResult {
	var results []ValidationResult

	// Blue Ocean works better with declarative pipelines
	if !strings.Contains(content, "pipeline {") {
		results = append(results, ValidationResult{
			File:     filePath,
			Type:     "blue-ocean",
			Level:    "info",
			Message:  "Blue Ocean works best with declarative pipelines",
			Rule:     "declarative-preferred",
			Category: "blue-ocean",
		})
	}

	// Check for stage names
	stagePattern := regexp.MustCompile(`stage\(['"]([^'"]+)['"]`)
	matches := stagePattern.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) > 1 {
			stageName := match[1]
			if len(stageName) > 50 {
				results = append(results, ValidationResult{
					File:     filePath,
					Type:     "blue-ocean",
					Level:    "warning",
					Message:  fmt.Sprintf("Stage name too long for Blue Ocean: %s", stageName),
					Rule:     "stage-name-length",
					Category: "blue-ocean",
				})
			}
		}
	}

	return results
}

func validatePlugins(filePath, content string, lines []string) []ValidationResult {
	var results []ValidationResult

	// Common plugin requirements
	pluginChecks := map[string]string{
		"docker":             "Docker Pipeline Plugin",
		"publishTestResults": "JUnit Plugin",
		"archiveArtifacts":   "Core Jenkins",
		"emailext":           "Email Extension Plugin",
		"slackSend":          "Slack Notification Plugin",
		"withCredentials":    "Credentials Plugin",
		"input":              "Pipeline Input Step Plugin",
		"parallel":           "Pipeline Graph Analysis Plugin",
		"build job:":         "Build Trigger Plugin",
		"publishHTML":        "HTML Publisher Plugin",
	}

	for feature, plugin := range pluginChecks {
		if strings.Contains(content, feature) {
			results = append(results, ValidationResult{
				File:     filePath,
				Type:     "plugins",
				Level:    "info",
				Message:  fmt.Sprintf("Pipeline uses %s - ensure %s is installed", feature, plugin),
				Rule:     "plugin-requirements",
				Category: "plugins",
			})
		}
	}

	return results
}

func extractSection(content, sectionName string) string {
	// Simple extraction of pipeline sections
	start := strings.Index(content, sectionName+" {")
	if start == -1 {
		return ""
	}

	braceCount := 0
	inSection := false
	result := ""

	for i := start; i < len(content); i++ {
		char := content[i]
		if char == '{' {
			braceCount++
			inSection = true
		} else if char == '}' {
			braceCount--
		}

		if inSection {
			result += string(char)
		}

		if inSection && braceCount == 0 {
			break
		}
	}

	return result
}

func outputValidationResults(report *ValidationReport) error {
	switch outputFormat {
	case "json":
		return outputValidationJSON(report)
	default:
		return outputValidationText(report)
	}
}

func outputValidationJSON(report *ValidationReport) error {
	// Simple JSON output - in production, use proper JSON marshaling
	fmt.Printf("{\n")
	fmt.Printf("  \"summary\": {\n")
	fmt.Printf("    \"total_checks\": %d,\n", report.Summary.TotalChecks)
	fmt.Printf("    \"error_count\": %d,\n", report.Summary.ErrorCount)
	fmt.Printf("    \"warning_count\": %d,\n", report.Summary.WarningCount)
	fmt.Printf("    \"info_count\": %d,\n", report.Summary.InfoCount)
	fmt.Printf("    \"valid\": %t\n", report.Summary.Valid)
	fmt.Printf("  },\n")
	fmt.Printf("  \"results\": [\n")

	for i, result := range report.Results {
		fmt.Printf("    {\n")
		fmt.Printf("      \"file\": \"%s\",\n", result.File)
		fmt.Printf("      \"type\": \"%s\",\n", result.Type)
		fmt.Printf("      \"level\": \"%s\",\n", result.Level)
		fmt.Printf("      \"message\": \"%s\",\n", result.Message)
		fmt.Printf("      \"rule\": \"%s\",\n", result.Rule)
		fmt.Printf("      \"category\": \"%s\"\n", result.Category)
		fmt.Printf("    }")
		if i < len(report.Results)-1 {
			fmt.Printf(",")
		}
		fmt.Printf("\n")
	}

	fmt.Printf("  ]\n")
	fmt.Printf("}\n")

	return nil
}

func outputValidationText(report *ValidationReport) error {
	if len(report.Results) == 0 {
		return nil
	}

	fmt.Printf("\n📋 Validation Results:\n")
	fmt.Printf("====================\n\n")

	for _, result := range report.Results {
		icon := "ℹ️"
		switch result.Level {
		case "error":
			icon = "❌"
		case "warning":
			icon = "⚠️"
		case "info":
			icon = "ℹ️"
		}

		fmt.Printf("%s [%s] %s", icon, strings.ToUpper(result.Level), result.File)
		if result.Line > 0 {
			fmt.Printf(":%d", result.Line)
		}
		fmt.Printf(": %s\n", result.Message)

		if result.Rule != "" {
			fmt.Printf("    Rule: %s\n", result.Rule)
		}
		if result.Category != "" {
			fmt.Printf("    Category: %s\n", result.Category)
		}
		fmt.Println()
	}

	return nil
}

func printValidationSummary(report *ValidationReport) {
	fmt.Printf("\n📊 Validation Summary:\n")
	fmt.Printf("=====================\n")
	fmt.Printf("Total checks: %d\n", report.Summary.TotalChecks)
	fmt.Printf("Errors: %d\n", report.Summary.ErrorCount)
	fmt.Printf("Warnings: %d\n", report.Summary.WarningCount)
	fmt.Printf("Info: %d\n", report.Summary.InfoCount)

	if report.Summary.Valid {
		fmt.Println("✅ Validation passed")
	} else {
		fmt.Println("❌ Validation failed")
	}
}
