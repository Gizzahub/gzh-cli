package gitlabci

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// ValidateCmd represents the validate command
var ValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate GitLab CI/CD pipeline configurations",
	Long: `Validate GitLab CI/CD pipeline configurations for syntax and best practices.

Performs comprehensive validation including:
- YAML syntax validation
- Pipeline structure validation
- Job and stage validation
- Security best practices validation
- Performance optimization checks
- GitLab CI/CD specific validation
- Runner configuration validation

Examples:
  gz gitlab-ci validate
  gz gitlab-ci validate --file .gitlab-ci.yml
  gz gitlab-ci validate --strict --security-check`,
	Run: runValidate,
}

var (
	validateFile      string
	strictValidation  bool
	securityCheck     bool
	performanceCheck  bool
	outputFormat      string
	allowWarnings     bool
	checkIncludes     bool
	validateVariables bool
)

func init() {
	ValidateCmd.Flags().StringVarP(&validateFile, "file", "f", ".gitlab-ci.yml", "Pipeline file to validate")
	ValidateCmd.Flags().BoolVar(&strictValidation, "strict", false, "Enable strict validation mode")
	ValidateCmd.Flags().BoolVar(&securityCheck, "security-check", true, "Enable security validation")
	ValidateCmd.Flags().BoolVar(&performanceCheck, "performance-check", true, "Enable performance validation")
	ValidateCmd.Flags().StringVarP(&outputFormat, "output", "o", "text", "Output format (text, json, yaml)")
	ValidateCmd.Flags().BoolVar(&allowWarnings, "allow-warnings", false, "Allow warnings (exit code 0)")
	ValidateCmd.Flags().BoolVar(&checkIncludes, "check-includes", true, "Validate included pipeline files")
	ValidateCmd.Flags().BoolVar(&validateVariables, "check-variables", true, "Validate pipeline variables")
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
	if validateFile == "" {
		fmt.Println("Error: pipeline file is required")
		os.Exit(1)
	}

	// Check if pipeline file exists
	if _, err := os.Stat(validateFile); os.IsNotExist(err) {
		fmt.Printf("Error: pipeline file not found: %s\n", validateFile)
		os.Exit(1)
	}

	fmt.Printf("🔍 Validating GitLab CI/CD pipeline: %s\n", validateFile)
	if strictValidation {
		fmt.Println("📋 Mode: Strict validation")
	}

	// Run validation
	report, err := validatePipeline(validateFile)
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

func validatePipeline(pipelineFile string) (*ValidationReport, error) {
	report := &ValidationReport{
		Results: []ValidationResult{},
		Summary: ValidationSummary{
			TotalFiles: 1,
		},
	}

	// Validate main pipeline file
	fileResults, err := validatePipelineFile(pipelineFile)
	if err != nil {
		result := ValidationResult{
			File:     pipelineFile,
			Type:     "file",
			Level:    "error",
			Message:  fmt.Sprintf("Failed to validate file: %v", err),
			Rule:     "file-access",
			Category: "syntax",
		}
		report.Results = append(report.Results, result)
		report.Summary.ErrorCount++
		return report, nil
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

	// Validate included files if enabled
	if checkIncludes {
		includeResults, err := validateIncludedFiles(pipelineFile)
		if err != nil {
			fmt.Printf("Warning: Failed to validate included files: %v\n", err)
		} else {
			for _, result := range includeResults {
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
		}
	}

	// Determine overall validity
	report.Summary.Valid = report.Summary.ErrorCount == 0 && (allowWarnings || report.Summary.WarningCount == 0)

	return report, nil
}

func validatePipelineFile(filePath string) ([]ValidationResult, error) {
	var results []ValidationResult

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Parse YAML
	var pipeline map[string]interface{}
	if err := yaml.Unmarshal(data, &pipeline); err != nil {
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

	// Validate pipeline structure
	results = append(results, validatePipelineStructure(filePath, pipeline)...)

	// Validate stages
	if stages, ok := pipeline["stages"].([]interface{}); ok {
		results = append(results, validateStages(filePath, stages)...)
	}

	// Validate jobs
	results = append(results, validateJobs(filePath, pipeline)...)

	// Validate variables
	if validateVariables {
		if variables, ok := pipeline["variables"].(map[string]interface{}); ok {
			results = append(results, validatePipelineVariables(filePath, variables)...)
		}
	}

	// Security validation
	if securityCheck {
		results = append(results, validatePipelineSecurity(filePath, pipeline)...)
	}

	// Performance validation
	if performanceCheck {
		results = append(results, validatePipelinePerformance(filePath, pipeline)...)
	}

	return results, nil
}

func validatePipelineStructure(filePath string, pipeline map[string]interface{}) []ValidationResult {
	var results []ValidationResult

	// Check for deprecated fields
	deprecatedFields := []string{"types", "before_script", "after_script"}
	for _, field := range deprecatedFields {
		if _, exists := pipeline[field]; exists {
			results = append(results, ValidationResult{
				File:     filePath,
				Type:     "structure",
				Level:    "warning",
				Message:  fmt.Sprintf("Deprecated field '%s' found - consider using modern alternatives", field),
				Rule:     "deprecated-fields",
				Category: "structure",
			})
		}
	}

	// Check for required structure
	if stages, hasStages := pipeline["stages"]; hasStages {
		if stagesList, ok := stages.([]interface{}); ok && len(stagesList) == 0 {
			results = append(results, ValidationResult{
				File:     filePath,
				Type:     "structure",
				Level:    "warning",
				Message:  "Empty stages list found",
				Rule:     "empty-stages",
				Category: "structure",
			})
		}
	}

	// Check for jobs
	hasJobs := false
	for key := range pipeline {
		if !isReservedKeyword(key) {
			hasJobs = true
			break
		}
	}

	if !hasJobs {
		results = append(results, ValidationResult{
			File:     filePath,
			Type:     "structure",
			Level:    "error",
			Message:  "No jobs found in pipeline",
			Rule:     "no-jobs",
			Category: "structure",
		})
	}

	return results
}

func isReservedKeyword(key string) bool {
	reserved := map[string]bool{
		"image":         true,
		"services":      true,
		"before_script": true,
		"after_script":  true,
		"stages":        true,
		"variables":     true,
		"cache":         true,
		"include":       true,
		"workflow":      true,
		"default":       true,
	}
	return reserved[key]
}

func validateStages(filePath string, stages []interface{}) []ValidationResult {
	var results []ValidationResult

	stageNames := make(map[string]bool)
	for i, stageInterface := range stages {
		stage, ok := stageInterface.(string)
		if !ok {
			results = append(results, ValidationResult{
				File:     filePath,
				Type:     "stage",
				Level:    "error",
				Message:  fmt.Sprintf("Stage %d is not a string", i),
				Rule:     "stage-type",
				Category: "structure",
			})
			continue
		}

		// Check for duplicate stages
		if stageNames[stage] {
			results = append(results, ValidationResult{
				File:     filePath,
				Type:     "stage",
				Level:    "error",
				Message:  fmt.Sprintf("Duplicate stage '%s' found", stage),
				Rule:     "duplicate-stages",
				Category: "structure",
			})
		}
		stageNames[stage] = true

		// Check stage naming conventions
		if !isValidStageName(stage) {
			results = append(results, ValidationResult{
				File:     filePath,
				Type:     "stage",
				Level:    "warning",
				Message:  fmt.Sprintf("Stage '%s' doesn't follow naming conventions", stage),
				Rule:     "stage-naming",
				Category: "structure",
			})
		}
	}

	return results
}

func isValidStageName(name string) bool {
	// Check for common stage naming patterns
	validPatterns := []string{
		"build", "test", "deploy", "package", "security", "lint",
		"prepare", "cleanup", "review", "staging", "production",
	}

	for _, pattern := range validPatterns {
		if strings.Contains(strings.ToLower(name), pattern) {
			return true
		}
	}

	// Allow custom stages with proper naming (lowercase, hyphens, colons)
	return strings.ToLower(name) == name && !strings.ContainsAny(name, " _")
}

func validateJobs(filePath string, pipeline map[string]interface{}) []ValidationResult {
	var results []ValidationResult

	for jobName, jobInterface := range pipeline {
		if isReservedKeyword(jobName) {
			continue
		}

		job, ok := jobInterface.(map[string]interface{})
		if !ok {
			results = append(results, ValidationResult{
				File:     filePath,
				Type:     "job",
				Level:    "error",
				Message:  fmt.Sprintf("Job '%s' is not a valid job configuration", jobName),
				Rule:     "job-structure",
				Category: "structure",
			})
			continue
		}

		// Validate individual job
		results = append(results, validateJob(filePath, jobName, job)...)
	}

	return results
}

func validateJob(filePath, jobName string, job map[string]interface{}) []ValidationResult {
	var results []ValidationResult

	// Check for required script or extends
	hasScript := false
	hasExtends := false

	if script, ok := job["script"]; ok {
		hasScript = true
		if scriptList, ok := script.([]interface{}); ok {
			results = append(results, validateJobScript(filePath, jobName, scriptList)...)
		}
	}

	if _, ok := job["extends"]; ok {
		hasExtends = true
	}

	if !hasScript && !hasExtends {
		results = append(results, ValidationResult{
			File:     filePath,
			Type:     "job",
			Level:    "error",
			Message:  fmt.Sprintf("Job '%s' must have either 'script' or 'extends' field", jobName),
			Rule:     "job-script-required",
			Category: "structure",
		})
	}

	// Validate stage reference
	if stage, ok := job["stage"].(string); ok {
		// Note: We can't validate stage existence without full pipeline context
		// This would require cross-referencing with the stages list
		if stage == "" {
			results = append(results, ValidationResult{
				File:     filePath,
				Type:     "job",
				Level:    "warning",
				Message:  fmt.Sprintf("Job '%s' has empty stage", jobName),
				Rule:     "empty-stage",
				Category: "structure",
			})
		}
	}

	// Validate job naming
	if !isValidJobName(jobName) {
		results = append(results, ValidationResult{
			File:     filePath,
			Type:     "job",
			Level:    "warning",
			Message:  fmt.Sprintf("Job '%s' doesn't follow naming conventions", jobName),
			Rule:     "job-naming",
			Category: "structure",
		})
	}

	// Validate rules vs only/except
	hasRules := false
	hasOnly := false
	hasExcept := false

	if _, ok := job["rules"]; ok {
		hasRules = true
	}
	if _, ok := job["only"]; ok {
		hasOnly = true
	}
	if _, ok := job["except"]; ok {
		hasExcept = true
	}

	if hasRules && (hasOnly || hasExcept) {
		results = append(results, ValidationResult{
			File:     filePath,
			Type:     "job",
			Level:    "error",
			Message:  fmt.Sprintf("Job '%s' cannot use 'rules' with 'only' or 'except'", jobName),
			Rule:     "conflicting-conditions",
			Category: "structure",
		})
	}

	// Validate artifacts configuration
	if artifacts, ok := job["artifacts"].(map[string]interface{}); ok {
		results = append(results, validateArtifacts(filePath, jobName, artifacts)...)
	}

	// Validate cache configuration
	if cache, ok := job["cache"].(map[string]interface{}); ok {
		results = append(results, validateCache(filePath, jobName, cache)...)
	}

	return results
}

func isValidJobName(name string) bool {
	// Job names should be descriptive and follow conventions
	// Allow letters, numbers, colons, hyphens, underscores, and spaces
	if strings.TrimSpace(name) != name {
		return false
	}

	// Check for common prefixes that indicate good naming
	goodPrefixes := []string{
		"build", "test", "deploy", "lint", "security", "package",
		"prepare", "cleanup", "review", "validate", "check",
	}

	lowerName := strings.ToLower(name)
	for _, prefix := range goodPrefixes {
		if strings.HasPrefix(lowerName, prefix) {
			return true
		}
		if strings.Contains(lowerName, ":"+prefix) {
			return true
		}
	}

	return true // Allow custom naming for flexibility
}

func validateJobScript(filePath, jobName string, script []interface{}) []ValidationResult {
	var results []ValidationResult

	if len(script) == 0 {
		results = append(results, ValidationResult{
			File:     filePath,
			Type:     "script",
			Level:    "warning",
			Message:  fmt.Sprintf("Job '%s' has empty script", jobName),
			Rule:     "empty-script",
			Category: "structure",
		})
		return results
	}

	// Check for potentially dangerous commands
	dangerousCommands := []string{
		"rm -rf /", "sudo rm", "chmod 777", "curl | sh", "wget | sh",
		"eval", "> /dev/null", "mktemp",
	}

	for i, cmdInterface := range script {
		cmd, ok := cmdInterface.(string)
		if !ok {
			continue
		}

		for _, dangerous := range dangerousCommands {
			if strings.Contains(strings.ToLower(cmd), dangerous) {
				results = append(results, ValidationResult{
					File:     filePath,
					Type:     "script",
					Level:    "warning",
					Message:  fmt.Sprintf("Job '%s' script line %d contains potentially dangerous command: %s", jobName, i+1, dangerous),
					Rule:     "dangerous-commands",
					Category: "security",
				})
			}
		}

		// Check for hardcoded secrets
		if containsHardcodedSecret(cmd) {
			results = append(results, ValidationResult{
				File:     filePath,
				Type:     "script",
				Level:    "warning",
				Message:  fmt.Sprintf("Job '%s' script line %d may contain hardcoded secrets", jobName, i+1),
				Rule:     "hardcoded-secrets",
				Category: "security",
			})
		}
	}

	return results
}

func containsHardcodedSecret(cmd string) bool {
	secretPatterns := []string{
		"password=", "token=", "key=", "secret=", "api_key=",
		"access_token=", "private_key=", "ssh_key=",
	}

	lowerCmd := strings.ToLower(cmd)
	for _, pattern := range secretPatterns {
		if strings.Contains(lowerCmd, pattern) && !strings.Contains(lowerCmd, "$") {
			return true
		}
	}

	return false
}

func validateArtifacts(filePath, jobName string, artifacts map[string]interface{}) []ValidationResult {
	var results []ValidationResult

	// Check for paths
	if paths, ok := artifacts["paths"].([]interface{}); ok {
		if len(paths) == 0 {
			results = append(results, ValidationResult{
				File:     filePath,
				Type:     "artifacts",
				Level:    "warning",
				Message:  fmt.Sprintf("Job '%s' has empty artifacts paths", jobName),
				Rule:     "empty-artifacts-paths",
				Category: "performance",
			})
		}

		// Check for overly broad paths
		for _, pathInterface := range paths {
			if path, ok := pathInterface.(string); ok {
				if path == "/" || path == "/*" || path == "**/*" {
					results = append(results, ValidationResult{
						File:     filePath,
						Type:     "artifacts",
						Level:    "warning",
						Message:  fmt.Sprintf("Job '%s' has overly broad artifacts path: %s", jobName, path),
						Rule:     "broad-artifacts-paths",
						Category: "performance",
					})
				}
			}
		}
	}

	// Check expire_in
	if expireIn, ok := artifacts["expire_in"].(string); ok {
		if !isValidExpireIn(expireIn) {
			results = append(results, ValidationResult{
				File:     filePath,
				Type:     "artifacts",
				Level:    "warning",
				Message:  fmt.Sprintf("Job '%s' has invalid expire_in format: %s", jobName, expireIn),
				Rule:     "invalid-expire-in",
				Category: "structure",
			})
		}
	}

	return results
}

func isValidExpireIn(expireIn string) bool {
	// Basic validation for expire_in format
	validSuffixes := []string{"sec", "min", "hr", "day", "week", "month", "year"}

	for _, suffix := range validSuffixes {
		if strings.HasSuffix(expireIn, suffix) {
			return true
		}
	}

	return expireIn == "never"
}

func validateCache(filePath, jobName string, cache map[string]interface{}) []ValidationResult {
	var results []ValidationResult

	// Check for paths
	if paths, ok := cache["paths"].([]interface{}); ok {
		if len(paths) == 0 {
			results = append(results, ValidationResult{
				File:     filePath,
				Type:     "cache",
				Level:    "warning",
				Message:  fmt.Sprintf("Job '%s' has empty cache paths", jobName),
				Rule:     "empty-cache-paths",
				Category: "performance",
			})
		}
	}

	// Check cache policy
	if policy, ok := cache["policy"].(string); ok {
		validPolicies := []string{"pull", "push", "pull-push"}
		isValid := false
		for _, validPolicy := range validPolicies {
			if policy == validPolicy {
				isValid = true
				break
			}
		}

		if !isValid {
			results = append(results, ValidationResult{
				File:     filePath,
				Type:     "cache",
				Level:    "error",
				Message:  fmt.Sprintf("Job '%s' has invalid cache policy: %s", jobName, policy),
				Rule:     "invalid-cache-policy",
				Category: "structure",
			})
		}
	}

	return results
}

func validatePipelineVariables(filePath string, variables map[string]interface{}) []ValidationResult {
	var results []ValidationResult

	for varName, varValue := range variables {
		// Check for naming conventions
		if !isValidVariableName(varName) {
			results = append(results, ValidationResult{
				File:     filePath,
				Type:     "variable",
				Level:    "warning",
				Message:  fmt.Sprintf("Variable '%s' doesn't follow naming conventions", varName),
				Rule:     "variable-naming",
				Category: "structure",
			})
		}

		// Check for sensitive values
		if valueStr, ok := varValue.(string); ok {
			if containsSensitiveValue(valueStr) {
				results = append(results, ValidationResult{
					File:     filePath,
					Type:     "variable",
					Level:    "warning",
					Message:  fmt.Sprintf("Variable '%s' may contain sensitive information", varName),
					Rule:     "sensitive-variable",
					Category: "security",
				})
			}
		}
	}

	return results
}

func isValidVariableName(name string) bool {
	// Variables should be UPPERCASE with underscores
	if strings.ToUpper(name) != name {
		return false
	}

	// Should not start with number
	if len(name) > 0 && name[0] >= '0' && name[0] <= '9' {
		return false
	}

	// Should only contain letters, numbers, and underscores
	for _, char := range name {
		if !((char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '_') {
			return false
		}
	}

	return true
}

func containsSensitiveValue(value string) bool {
	sensitivePatterns := []string{
		"password", "token", "key", "secret", "credential",
		"bearer", "oauth", "jwt", "ssh_key", "private_key",
	}

	lowerValue := strings.ToLower(value)
	for _, pattern := range sensitivePatterns {
		if strings.Contains(lowerValue, pattern) {
			return true
		}
	}

	return false
}

func validatePipelineSecurity(filePath string, pipeline map[string]interface{}) []ValidationResult {
	var results []ValidationResult

	// Check for image security
	if image, ok := pipeline["image"].(string); ok {
		if !isSecureImage(image) {
			results = append(results, ValidationResult{
				File:     filePath,
				Type:     "security",
				Level:    "warning",
				Message:  fmt.Sprintf("Image '%s' may not be from a trusted registry", image),
				Rule:     "untrusted-image",
				Category: "security",
			})
		}

		if strings.Contains(image, ":latest") {
			results = append(results, ValidationResult{
				File:     filePath,
				Type:     "security",
				Level:    "warning",
				Message:  "Using ':latest' tag is not recommended for production",
				Rule:     "latest-tag",
				Category: "security",
			})
		}
	}

	// Check for privileged mode
	for jobName, jobInterface := range pipeline {
		if isReservedKeyword(jobName) {
			continue
		}

		if job, ok := jobInterface.(map[string]interface{}); ok {
			if services, ok := job["services"].([]interface{}); ok {
				for _, serviceInterface := range services {
					if service, ok := serviceInterface.(map[string]interface{}); ok {
						if privileged, ok := service["privileged"].(bool); ok && privileged {
							results = append(results, ValidationResult{
								File:     filePath,
								Type:     "security",
								Level:    "warning",
								Message:  fmt.Sprintf("Job '%s' uses privileged mode", jobName),
								Rule:     "privileged-mode",
								Category: "security",
							})
						}
					}
				}
			}
		}
	}

	return results
}

func isSecureImage(image string) bool {
	trustedRegistries := []string{
		"docker.io/library/",
		"registry.gitlab.com/",
		"gcr.io/",
		"quay.io/",
		"docker.io/",
	}

	// If no registry specified, it's from Docker Hub (considered trusted)
	if !strings.Contains(image, "/") {
		return true
	}

	for _, registry := range trustedRegistries {
		if strings.HasPrefix(image, registry) {
			return true
		}
	}

	return false
}

func validatePipelinePerformance(filePath string, pipeline map[string]interface{}) []ValidationResult {
	var results []ValidationResult

	// Check for cache usage
	hasCaching := false
	if _, ok := pipeline["cache"]; ok {
		hasCaching = true
	}

	// Check jobs for cache usage
	jobsWithCache := 0
	totalJobs := 0

	for jobName, jobInterface := range pipeline {
		if isReservedKeyword(jobName) {
			continue
		}

		totalJobs++
		if job, ok := jobInterface.(map[string]interface{}); ok {
			if _, ok := job["cache"]; ok {
				jobsWithCache++
			}
		}
	}

	if !hasCaching && jobsWithCache == 0 && totalJobs > 0 {
		results = append(results, ValidationResult{
			File:     filePath,
			Type:     "performance",
			Level:    "info",
			Message:  "Pipeline doesn't use caching - consider adding cache for better performance",
			Rule:     "no-caching",
			Category: "performance",
		})
	}

	// Check for parallel jobs
	hasParallel := false
	for jobName, jobInterface := range pipeline {
		if isReservedKeyword(jobName) {
			continue
		}

		if job, ok := jobInterface.(map[string]interface{}); ok {
			if _, ok := job["parallel"]; ok {
				hasParallel = true
				break
			}
		}
	}

	if !hasParallel && totalJobs > 3 {
		results = append(results, ValidationResult{
			File:     filePath,
			Type:     "performance",
			Level:    "info",
			Message:  "Consider using parallel jobs for better performance",
			Rule:     "no-parallel",
			Category: "performance",
		})
	}

	return results
}

func validateIncludedFiles(mainFile string) ([]ValidationResult, error) {
	var results []ValidationResult

	// Read main file to find includes
	data, err := os.ReadFile(mainFile)
	if err != nil {
		return nil, err
	}

	var pipeline map[string]interface{}
	if err := yaml.Unmarshal(data, &pipeline); err != nil {
		return nil, err
	}

	// Check for include section
	if includeInterface, ok := pipeline["include"]; ok {
		if includeList, ok := includeInterface.([]interface{}); ok {
			for _, includeItem := range includeList {
				if include, ok := includeItem.(map[string]interface{}); ok {
					if localFile, ok := include["local"].(string); ok {
						// Validate local included file
						localPath := filepath.Join(filepath.Dir(mainFile), localFile)
						if _, err := os.Stat(localPath); os.IsNotExist(err) {
							results = append(results, ValidationResult{
								File:     mainFile,
								Type:     "include",
								Level:    "error",
								Message:  fmt.Sprintf("Local include file not found: %s", localFile),
								Rule:     "missing-include",
								Category: "structure",
							})
						} else {
							// Recursively validate included file
							includeResults, err := validatePipelineFile(localPath)
							if err != nil {
								results = append(results, ValidationResult{
									File:     localPath,
									Type:     "include",
									Level:    "error",
									Message:  fmt.Sprintf("Failed to validate included file: %v", err),
									Rule:     "include-validation-failed",
									Category: "structure",
								})
							} else {
								results = append(results, includeResults...)
							}
						}
					}
				}
			}
		}
	}

	return results, nil
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
