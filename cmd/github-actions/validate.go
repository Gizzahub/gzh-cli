package githubactions

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// ValidateCmd represents the validate command
var ValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate GitHub Actions workflows",
	Long: `Validate GitHub Actions workflows for syntax and best practices.

Performs comprehensive validation including:
- YAML syntax validation
- Workflow structure validation
- Job and step validation
- Security best practices validation
- Performance optimization checks
- Dependency validation

Examples:
  gz github-actions validate
  gz github-actions validate --path .github/workflows
  gz github-actions validate --strict --security-check`,
	Run: runValidate,
}

var (
	validatePath     string
	strictValidation bool
	securityCheck    bool
	performanceCheck bool
	outputFormat     string
	allowWarnings    bool
)

func init() {
	ValidateCmd.Flags().StringVarP(&validatePath, "path", "p", ".github/workflows", "Path to workflows directory")
	ValidateCmd.Flags().BoolVar(&strictValidation, "strict", false, "Enable strict validation mode")
	ValidateCmd.Flags().BoolVar(&securityCheck, "security-check", true, "Enable security validation")
	ValidateCmd.Flags().BoolVar(&performanceCheck, "performance-check", true, "Enable performance validation")
	ValidateCmd.Flags().StringVarP(&outputFormat, "output", "o", "text", "Output format (text, json, yaml)")
	ValidateCmd.Flags().BoolVar(&allowWarnings, "allow-warnings", false, "Allow warnings (exit code 0)")
}

// ValidationResult represents a validation result
type ValidationResult struct {
	File     string                 `json:"file" yaml:"file"`
	Type     string                 `json:"type" yaml:"type"`
	Level    string                 `json:"level" yaml:"level"` // error, warning, info
	Message  string                 `json:"message" yaml:"message"`
	Line     int                    `json:"line,omitempty" yaml:"line,omitempty"`
	Column   int                    `json:"column,omitempty" yaml:"column,omitempty"`
	Rule     string                 `json:"rule" yaml:"rule"`
	Category string                 `json:"category" yaml:"category"`
	Details  map[string]interface{} `json:"details,omitempty" yaml:"details,omitempty"`
}

// ValidationReport represents the complete validation report
type ValidationReport struct {
	Summary ValidationSummary  `json:"summary" yaml:"summary"`
	Results []ValidationResult `json:"results" yaml:"results"`
}

// ValidationSummary represents validation summary statistics
type ValidationSummary struct {
	TotalFiles   int  `json:"total_files" yaml:"total_files"`
	TotalChecks  int  `json:"total_checks" yaml:"total_checks"`
	ErrorCount   int  `json:"error_count" yaml:"error_count"`
	WarningCount int  `json:"warning_count" yaml:"warning_count"`
	InfoCount    int  `json:"info_count" yaml:"info_count"`
	Valid        bool `json:"valid" yaml:"valid"`
}

func runValidate(cmd *cobra.Command, args []string) {
	if validatePath == "" {
		fmt.Println("Error: workflows path is required")
		os.Exit(1)
	}

	// Check if workflows directory exists
	if _, err := os.Stat(validatePath); os.IsNotExist(err) {
		fmt.Printf("Error: workflows directory not found: %s\n", validatePath)
		os.Exit(1)
	}

	fmt.Printf("🔍 Validating GitHub Actions workflows: %s\n", validatePath)
	if strictValidation {
		fmt.Println("📋 Mode: Strict validation")
	}

	// Run validation
	report, err := validateWorkflows(validatePath)
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

func validateWorkflows(workflowsPath string) (*ValidationReport, error) {
	report := &ValidationReport{
		Results: []ValidationResult{},
	}

	// Walk through workflows directory
	err := filepath.WalkDir(workflowsPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Continue on errors
		}

		// Skip directories and non-YAML files
		if d.IsDir() || (!strings.HasSuffix(path, ".yml") && !strings.HasSuffix(path, ".yaml")) {
			return nil
		}

		report.Summary.TotalFiles++

		// Validate workflow file
		fileResults, err := validateWorkflowFile(path)
		if err != nil {
			result := ValidationResult{
				File:     path,
				Type:     "file",
				Level:    "error",
				Message:  fmt.Sprintf("Failed to validate file: %v", err),
				Rule:     "file-access",
				Category: "syntax",
			}
			report.Results = append(report.Results, result)
			report.Summary.ErrorCount++
			return nil
		}

		// Add results
		for _, result := range fileResults {
			report.Results = append(report.Results, result)
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

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Determine overall validity
	report.Summary.Valid = report.Summary.ErrorCount == 0 && (allowWarnings || report.Summary.WarningCount == 0)

	return report, nil
}

func validateWorkflowFile(filePath string) ([]ValidationResult, error) {
	var results []ValidationResult

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Parse YAML
	var workflow map[string]interface{}
	if err := yaml.Unmarshal(data, &workflow); err != nil {
		result := ValidationResult{
			File:     filePath,
			Type:     "yaml",
			Level:    "error",
			Message:  fmt.Sprintf("YAML parsing error: %v", err),
			Rule:     "yaml-syntax",
			Category: "syntax",
		}
		results = append(results, result)
		return results, nil
	}

	// Validate workflow structure
	results = append(results, validateWorkflowStructure(filePath, workflow)...)

	// Validate jobs
	if jobs, ok := workflow["jobs"].(map[string]interface{}); ok {
		for jobName, jobData := range jobs {
			if job, ok := jobData.(map[string]interface{}); ok {
				results = append(results, validateJob(filePath, jobName, job)...)
			}
		}
	}

	// Security validation
	if securityCheck {
		results = append(results, validateWorkflowSecurity(filePath, workflow)...)
	}

	// Performance validation
	if performanceCheck {
		results = append(results, validateWorkflowPerformance(filePath, workflow)...)
	}

	return results, nil
}

func validateWorkflowStructure(filePath string, workflow map[string]interface{}) []ValidationResult {
	var results []ValidationResult

	// Check required fields
	if _, hasName := workflow["name"]; !hasName {
		results = append(results, ValidationResult{
			File:     filePath,
			Type:     "workflow",
			Level:    "error",
			Message:  "Workflow missing required field: name",
			Rule:     "required-fields",
			Category: "structure",
		})
	}

	if _, hasOn := workflow["on"]; !hasOn {
		results = append(results, ValidationResult{
			File:     filePath,
			Type:     "workflow",
			Level:    "error",
			Message:  "Workflow missing required field: on",
			Rule:     "required-fields",
			Category: "structure",
		})
	}

	if _, hasJobs := workflow["jobs"]; !hasJobs {
		results = append(results, ValidationResult{
			File:     filePath,
			Type:     "workflow",
			Level:    "error",
			Message:  "Workflow missing required field: jobs",
			Rule:     "required-fields",
			Category: "structure",
		})
	}

	// Validate triggers
	if onTriggers, ok := workflow["on"]; ok {
		results = append(results, validateTriggers(filePath, onTriggers)...)
	}

	return results
}

func validateTriggers(filePath string, triggers interface{}) []ValidationResult {
	var results []ValidationResult

	// Handle different trigger formats
	switch t := triggers.(type) {
	case string:
		// Simple trigger like "push"
		if !isValidTriggerType(t) {
			results = append(results, ValidationResult{
				File:     filePath,
				Type:     "trigger",
				Level:    "error",
				Message:  fmt.Sprintf("Invalid trigger type: %s", t),
				Rule:     "valid-triggers",
				Category: "structure",
			})
		}
	case []interface{}:
		// Array of triggers
		for _, trigger := range t {
			if triggerStr, ok := trigger.(string); ok {
				if !isValidTriggerType(triggerStr) {
					results = append(results, ValidationResult{
						File:     filePath,
						Type:     "trigger",
						Level:    "error",
						Message:  fmt.Sprintf("Invalid trigger type: %s", triggerStr),
						Rule:     "valid-triggers",
						Category: "structure",
					})
				}
			}
		}
	case map[string]interface{}:
		// Complex triggers with configuration
		for triggerType := range t {
			if !isValidTriggerType(triggerType) {
				results = append(results, ValidationResult{
					File:     filePath,
					Type:     "trigger",
					Level:    "error",
					Message:  fmt.Sprintf("Invalid trigger type: %s", triggerType),
					Rule:     "valid-triggers",
					Category: "structure",
				})
			}
		}
	}

	return results
}

func isValidTriggerType(triggerType string) bool {
	validTriggers := map[string]bool{
		"push":                true,
		"pull_request":        true,
		"pull_request_target": true,
		"schedule":            true,
		"workflow_dispatch":   true,
		"workflow_call":       true,
		"release":             true,
		"issues":              true,
		"issue_comment":       true,
		"fork":                true,
		"watch":               true,
		"create":              true,
		"delete":              true,
		"deployment":          true,
		"deployment_status":   true,
		"page_build":          true,
		"public":              true,
		"gollum":              true,
		"project":             true,
		"project_card":        true,
		"project_column":      true,
		"milestone":           true,
		"label":               true,
	}
	return validTriggers[triggerType]
}

func validateJob(filePath, jobName string, job map[string]interface{}) []ValidationResult {
	var results []ValidationResult

	// Check required fields
	if _, hasRunsOn := job["runs-on"]; !hasRunsOn {
		results = append(results, ValidationResult{
			File:     filePath,
			Type:     "job",
			Level:    "error",
			Message:  fmt.Sprintf("Job '%s' missing required field: runs-on", jobName),
			Rule:     "job-required-fields",
			Category: "structure",
		})
	}

	if _, hasSteps := job["steps"]; !hasSteps {
		results = append(results, ValidationResult{
			File:     filePath,
			Type:     "job",
			Level:    "error",
			Message:  fmt.Sprintf("Job '%s' missing required field: steps", jobName),
			Rule:     "job-required-fields",
			Category: "structure",
		})
	}

	// Validate steps
	if steps, ok := job["steps"].([]interface{}); ok {
		for i, stepData := range steps {
			if step, ok := stepData.(map[string]interface{}); ok {
				results = append(results, validateStep(filePath, jobName, i, step)...)
			}
		}
	}

	// Validate runner
	if runsOn, ok := job["runs-on"]; ok {
		results = append(results, validateRunner(filePath, jobName, runsOn)...)
	}

	return results
}

func validateStep(filePath, jobName string, stepIndex int, step map[string]interface{}) []ValidationResult {
	var results []ValidationResult

	stepID := fmt.Sprintf("%s[%d]", jobName, stepIndex)

	// A step must have either 'uses' or 'run'
	hasUses := false
	hasRun := false

	if _, ok := step["uses"]; ok {
		hasUses = true
	}
	if _, ok := step["run"]; ok {
		hasRun = true
	}

	if !hasUses && !hasRun {
		results = append(results, ValidationResult{
			File:     filePath,
			Type:     "step",
			Level:    "error",
			Message:  fmt.Sprintf("Step %s must have either 'uses' or 'run' field", stepID),
			Rule:     "step-action-required",
			Category: "structure",
		})
	}

	if hasUses && hasRun {
		results = append(results, ValidationResult{
			File:     filePath,
			Type:     "step",
			Level:    "error",
			Message:  fmt.Sprintf("Step %s cannot have both 'uses' and 'run' fields", stepID),
			Rule:     "step-action-exclusive",
			Category: "structure",
		})
	}

	// Validate action versions
	if uses, ok := step["uses"].(string); ok {
		results = append(results, validateActionVersion(filePath, stepID, uses)...)
	}

	// Validate shell commands
	if run, ok := step["run"].(string); ok {
		results = append(results, validateShellCommand(filePath, stepID, run)...)
	}

	return results
}

func validateActionVersion(filePath, stepID, uses string) []ValidationResult {
	var results []ValidationResult

	// Check for version pinning
	if strings.Contains(uses, "@") {
		parts := strings.Split(uses, "@")
		if len(parts) >= 2 {
			version := parts[1]

			// Warn about using @main or @master
			if version == "main" || version == "master" {
				results = append(results, ValidationResult{
					File:     filePath,
					Type:     "action",
					Level:    "warning",
					Message:  fmt.Sprintf("Step %s uses unstable version '%s' - consider pinning to a specific version", stepID, version),
					Rule:     "action-version-pinning",
					Category: "security",
				})
			}
		}
	} else {
		results = append(results, ValidationResult{
			File:     filePath,
			Type:     "action",
			Level:    "warning",
			Message:  fmt.Sprintf("Step %s action '%s' is not version pinned", stepID, uses),
			Rule:     "action-version-required",
			Category: "security",
		})
	}

	return results
}

func validateShellCommand(filePath, stepID, command string) []ValidationResult {
	var results []ValidationResult

	// Check for potentially dangerous commands
	dangerousCommands := []string{"sudo", "rm -rf", "chmod 777", "curl | sh", "wget | sh"}

	for _, dangerous := range dangerousCommands {
		if strings.Contains(strings.ToLower(command), dangerous) {
			results = append(results, ValidationResult{
				File:     filePath,
				Type:     "shell",
				Level:    "warning",
				Message:  fmt.Sprintf("Step %s contains potentially dangerous command: %s", stepID, dangerous),
				Rule:     "shell-command-safety",
				Category: "security",
			})
		}
	}

	return results
}

func validateRunner(filePath, jobName string, runsOn interface{}) []ValidationResult {
	var results []ValidationResult

	switch runner := runsOn.(type) {
	case string:
		if !isValidRunner(runner) {
			results = append(results, ValidationResult{
				File:     filePath,
				Type:     "runner",
				Level:    "warning",
				Message:  fmt.Sprintf("Job '%s' uses potentially invalid runner: %s", jobName, runner),
				Rule:     "runner-validation",
				Category: "performance",
			})
		}
	case []interface{}:
		for _, r := range runner {
			if runnerStr, ok := r.(string); ok {
				if !isValidRunner(runnerStr) {
					results = append(results, ValidationResult{
						File:     filePath,
						Type:     "runner",
						Level:    "warning",
						Message:  fmt.Sprintf("Job '%s' uses potentially invalid runner: %s", jobName, runnerStr),
						Rule:     "runner-validation",
						Category: "performance",
					})
				}
			}
		}
	}

	return results
}

func isValidRunner(runner string) bool {
	// Check for GitHub-hosted runners
	githubRunners := []string{
		"ubuntu-latest", "ubuntu-22.04", "ubuntu-20.04",
		"windows-latest", "windows-2022", "windows-2019",
		"macos-latest", "macos-12", "macos-11",
	}

	for _, validRunner := range githubRunners {
		if runner == validRunner {
			return true
		}
	}

	// Could be self-hosted runner or matrix variable
	return strings.HasPrefix(runner, "self-hosted") || strings.Contains(runner, "${{")
}

func validateWorkflowSecurity(filePath string, workflow map[string]interface{}) []ValidationResult {
	var results []ValidationResult

	// Check for hardcoded secrets
	workflowStr := fmt.Sprintf("%v", workflow)
	if strings.Contains(workflowStr, "password") || strings.Contains(workflowStr, "token") {
		// More sophisticated secret detection would be needed
		if !strings.Contains(workflowStr, "secrets.") {
			results = append(results, ValidationResult{
				File:     filePath,
				Type:     "security",
				Level:    "warning",
				Message:  "Workflow may contain hardcoded secrets - use GitHub secrets instead",
				Rule:     "no-hardcoded-secrets",
				Category: "security",
			})
		}
	}

	// Check for pull_request_target usage
	if onTriggers, ok := workflow["on"]; ok {
		if triggersMap, ok := onTriggers.(map[string]interface{}); ok {
			if _, hasPRTarget := triggersMap["pull_request_target"]; hasPRTarget {
				results = append(results, ValidationResult{
					File:     filePath,
					Type:     "security",
					Level:    "warning",
					Message:  "Using pull_request_target trigger - ensure proper security measures",
					Rule:     "pull-request-target-security",
					Category: "security",
				})
			}
		}
	}

	return results
}

func validateWorkflowPerformance(filePath string, workflow map[string]interface{}) []ValidationResult {
	var results []ValidationResult

	// Check for caching usage
	if jobs, ok := workflow["jobs"].(map[string]interface{}); ok {
		for jobName, jobData := range jobs {
			if job, ok := jobData.(map[string]interface{}); ok {
				if steps, ok := job["steps"].([]interface{}); ok {
					hasCaching := false
					for _, stepData := range steps {
						if step, ok := stepData.(map[string]interface{}); ok {
							if uses, ok := step["uses"].(string); ok {
								if strings.Contains(uses, "actions/cache") {
									hasCaching = true
									break
								}
							}
						}
					}

					// Check if job should use caching
					if shouldUseCaching(steps) && !hasCaching {
						results = append(results, ValidationResult{
							File:     filePath,
							Type:     "performance",
							Level:    "info",
							Message:  fmt.Sprintf("Job '%s' might benefit from dependency caching", jobName),
							Rule:     "dependency-caching",
							Category: "performance",
						})
					}
				}
			}
		}
	}

	return results
}

func shouldUseCaching(steps []interface{}) bool {
	// Check if workflow installs dependencies that could benefit from caching
	for _, stepData := range steps {
		if step, ok := stepData.(map[string]interface{}); ok {
			if run, ok := step["run"].(string); ok {
				cacheableCommands := []string{"npm install", "yarn install", "pip install", "go mod download", "mvn install"}
				for _, cmd := range cacheableCommands {
					if strings.Contains(strings.ToLower(run), cmd) {
						return true
					}
				}
			}
		}
	}
	return false
}

func outputValidationResults(report *ValidationReport) error {
	switch outputFormat {
	case "json":
		return outputValidationJSON(report)
	case "yaml":
		return outputValidationYAML(report)
	default:
		return outputValidationText(report)
	}
}

func outputValidationJSON(report *ValidationReport) error {
	encoder := yaml.NewEncoder(os.Stdout)
	defer encoder.Close()
	return encoder.Encode(report)
}

func outputValidationYAML(report *ValidationReport) error {
	encoder := yaml.NewEncoder(os.Stdout)
	defer encoder.Close()
	return encoder.Encode(report)
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

		fmt.Printf("%s [%s] %s: %s\n", icon, strings.ToUpper(result.Level), result.File, result.Message)
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
	fmt.Printf("Files validated: %d\n", report.Summary.TotalFiles)
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
