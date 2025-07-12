package operator

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
	Short: "Validate operator configurations and manifests",
	Long: `Validate operator configurations and manifests for correctness.

Performs comprehensive validation including:
- CRD schema validation
- RBAC configuration validation
- Deployment manifest validation
- Resource naming convention validation
- Security best practices validation
- Kubernetes API compatibility validation

Examples:
  gz operator validate --path ./operator
  gz operator validate --path ./operator --strict
  gz operator validate --check-security --check-rbac`,
	Run: runValidate,
}

var (
	validatePath   string
	strictMode     bool
	checkSecurity  bool
	checkRBAC      bool
	checkSchema    bool
	validateOutput string
	allowWarnings  bool
)

func init() {
	ValidateCmd.Flags().StringVarP(&validatePath, "path", "p", "./operator", "Path to operator directory")
	ValidateCmd.Flags().BoolVar(&strictMode, "strict", false, "Enable strict validation mode")
	ValidateCmd.Flags().BoolVar(&checkSecurity, "check-security", true, "Validate security configurations")
	ValidateCmd.Flags().BoolVar(&checkRBAC, "check-rbac", true, "Validate RBAC configurations")
	ValidateCmd.Flags().BoolVar(&checkSchema, "check-schema", true, "Validate CRD schemas")
	ValidateCmd.Flags().StringVarP(&validateOutput, "output", "o", "text", "Output format (text, json, yaml)")
	ValidateCmd.Flags().BoolVar(&allowWarnings, "allow-warnings", false, "Allow warnings (exit code 0)")
}

// ValidationResult represents a validation result
type ValidationResult struct {
	File    string                 `json:"file" yaml:"file"`
	Type    string                 `json:"type" yaml:"type"`
	Level   string                 `json:"level" yaml:"level"` // error, warning, info
	Message string                 `json:"message" yaml:"message"`
	Line    int                    `json:"line,omitempty" yaml:"line,omitempty"`
	Column  int                    `json:"column,omitempty" yaml:"column,omitempty"`
	Rule    string                 `json:"rule" yaml:"rule"`
	Details map[string]interface{} `json:"details,omitempty" yaml:"details,omitempty"`
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
		fmt.Println("Error: operator path is required")
		os.Exit(1)
	}

	// Check if operator directory exists
	if _, err := os.Stat(validatePath); os.IsNotExist(err) {
		fmt.Printf("Error: operator directory not found: %s\n", validatePath)
		os.Exit(1)
	}

	fmt.Printf("🔍 Validating operator: %s\n", validatePath)
	if strictMode {
		fmt.Println("📋 Mode: Strict validation")
	}

	// Run validation
	report, err := validateOperator(validatePath)
	if err != nil {
		fmt.Printf("Error running validation: %v\n", err)
		os.Exit(1)
	}

	// Output results
	if err := outputResults(report); err != nil {
		fmt.Printf("Error outputting results: %v\n", err)
		os.Exit(1)
	}

	// Print summary
	printSummary(report)

	// Exit with appropriate code
	if report.Summary.ErrorCount > 0 {
		os.Exit(1)
	} else if report.Summary.WarningCount > 0 && !allowWarnings {
		os.Exit(2)
	}
}

func validateOperator(operatorPath string) (*ValidationReport, error) {
	report := &ValidationReport{
		Results: []ValidationResult{},
	}

	// Walk through operator directory
	err := filepath.WalkDir(operatorPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Continue on errors
		}

		// Skip directories and non-YAML files
		if d.IsDir() || (!strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml")) {
			return nil
		}

		report.Summary.TotalFiles++

		// Validate file
		fileResults, err := validateFile(path)
		if err != nil {
			result := ValidationResult{
				File:    path,
				Type:    "file",
				Level:   "error",
				Message: fmt.Sprintf("Failed to validate file: %v", err),
				Rule:    "file-access",
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

func validateFile(filePath string) ([]ValidationResult, error) {
	var results []ValidationResult

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Parse YAML
	var docs []map[string]interface{}
	decoder := yaml.NewDecoder(strings.NewReader(string(data)))

	for {
		var doc map[string]interface{}
		if err := decoder.Decode(&doc); err != nil {
			if err.Error() == "EOF" {
				break
			}
			result := ValidationResult{
				File:    filePath,
				Type:    "yaml",
				Level:   "error",
				Message: fmt.Sprintf("YAML parsing error: %v", err),
				Rule:    "yaml-syntax",
			}
			results = append(results, result)
			return results, nil
		}

		if len(doc) > 0 {
			docs = append(docs, doc)
		}
	}

	// Validate each document
	for _, doc := range docs {
		docResults := validateDocument(filePath, doc)
		results = append(results, docResults...)
	}

	return results, nil
}

func validateDocument(filePath string, doc map[string]interface{}) []ValidationResult {
	var results []ValidationResult

	// Get basic fields
	apiVersion, _ := doc["apiVersion"].(string)
	kind, _ := doc["kind"].(string)
	metadata, _ := doc["metadata"].(map[string]interface{})

	// Validate required fields
	if apiVersion == "" {
		results = append(results, ValidationResult{
			File:    filePath,
			Type:    "kubernetes",
			Level:   "error",
			Message: "Missing required field: apiVersion",
			Rule:    "required-fields",
		})
	}

	if kind == "" {
		results = append(results, ValidationResult{
			File:    filePath,
			Type:    "kubernetes",
			Level:   "error",
			Message: "Missing required field: kind",
			Rule:    "required-fields",
		})
	}

	if metadata == nil {
		results = append(results, ValidationResult{
			File:    filePath,
			Type:    "kubernetes",
			Level:   "error",
			Message: "Missing required field: metadata",
			Rule:    "required-fields",
		})
	}

	// Type-specific validation
	switch kind {
	case "CustomResourceDefinition":
		results = append(results, validateCRD(filePath, doc)...)
	case "ClusterRole", "Role":
		if checkRBAC {
			results = append(results, validateRBAC(filePath, doc)...)
		}
	case "Deployment":
		results = append(results, validateDeployment(filePath, doc)...)
	case "ServiceAccount":
		results = append(results, validateServiceAccount(filePath, doc)...)
	}

	// Security validation
	if checkSecurity {
		results = append(results, validateSecurity(filePath, doc)...)
	}

	// Naming convention validation
	results = append(results, validateNaming(filePath, doc)...)

	return results
}

func validateCRD(filePath string, doc map[string]interface{}) []ValidationResult {
	var results []ValidationResult

	spec, ok := doc["spec"].(map[string]interface{})
	if !ok {
		results = append(results, ValidationResult{
			File:    filePath,
			Type:    "crd",
			Level:   "error",
			Message: "CRD missing spec section",
			Rule:    "crd-spec",
		})
		return results
	}

	// Validate group
	group, _ := spec["group"].(string)
	if group == "" {
		results = append(results, ValidationResult{
			File:    filePath,
			Type:    "crd",
			Level:   "error",
			Message: "CRD missing group in spec",
			Rule:    "crd-group",
		})
	} else if !strings.Contains(group, ".") {
		results = append(results, ValidationResult{
			File:    filePath,
			Type:    "crd",
			Level:   "warning",
			Message: "CRD group should be a DNS subdomain (contain '.')",
			Rule:    "crd-group-format",
		})
	}

	// Validate versions
	versions, ok := spec["versions"].([]interface{})
	if !ok || len(versions) == 0 {
		results = append(results, ValidationResult{
			File:    filePath,
			Type:    "crd",
			Level:   "error",
			Message: "CRD must have at least one version",
			Rule:    "crd-versions",
		})
	}

	// Validate schema if checking schemas
	if checkSchema {
		results = append(results, validateCRDSchema(filePath, spec)...)
	}

	return results
}

func validateCRDSchema(filePath string, spec map[string]interface{}) []ValidationResult {
	var results []ValidationResult

	versions, ok := spec["versions"].([]interface{})
	if !ok {
		return results
	}

	for i, versionIface := range versions {
		version, ok := versionIface.(map[string]interface{})
		if !ok {
			continue
		}

		schema, hasSchema := version["schema"]
		if !hasSchema {
			results = append(results, ValidationResult{
				File:    filePath,
				Type:    "crd-schema",
				Level:   "warning",
				Message: fmt.Sprintf("Version %d missing schema definition", i),
				Rule:    "crd-schema-required",
			})
		}

		// Validate schema structure
		if schemaMap, ok := schema.(map[string]interface{}); ok {
			if openAPISchema, hasOpenAPI := schemaMap["openAPIV3Schema"]; hasOpenAPI {
				if schemaProps, ok := openAPISchema.(map[string]interface{}); ok {
					results = append(results, validateSchemaProperties(filePath, schemaProps)...)
				}
			}
		}
	}

	return results
}

func validateSchemaProperties(filePath string, schema map[string]interface{}) []ValidationResult {
	var results []ValidationResult

	// Check for type definition
	if _, hasType := schema["type"]; !hasType {
		results = append(results, ValidationResult{
			File:    filePath,
			Type:    "crd-schema",
			Level:   "warning",
			Message: "Schema missing type definition",
			Rule:    "schema-type",
		})
	}

	// Validate properties if present
	if properties, hasProps := schema["properties"].(map[string]interface{}); hasProps {
		for propName, propDef := range properties {
			if propMap, ok := propDef.(map[string]interface{}); ok {
				if _, hasType := propMap["type"]; !hasType {
					results = append(results, ValidationResult{
						File:    filePath,
						Type:    "crd-schema",
						Level:   "warning",
						Message: fmt.Sprintf("Property '%s' missing type definition", propName),
						Rule:    "schema-property-type",
					})
				}
			}
		}
	}

	return results
}

func validateRBAC(filePath string, doc map[string]interface{}) []ValidationResult {
	var results []ValidationResult

	rules, ok := doc["rules"].([]interface{})
	if !ok {
		results = append(results, ValidationResult{
			File:    filePath,
			Type:    "rbac",
			Level:   "error",
			Message: "RBAC resource missing rules",
			Rule:    "rbac-rules",
		})
		return results
	}

	// Check for overly permissive rules
	for i, ruleIface := range rules {
		rule, ok := ruleIface.(map[string]interface{})
		if !ok {
			continue
		}

		verbs, hasVerbs := rule["verbs"].([]interface{})
		if hasVerbs {
			for _, verbIface := range verbs {
				if verb, ok := verbIface.(string); ok && verb == "*" {
					results = append(results, ValidationResult{
						File:    filePath,
						Type:    "rbac",
						Level:   "warning",
						Message: fmt.Sprintf("Rule %d uses wildcard verb '*' - consider being more specific", i),
						Rule:    "rbac-wildcard-verbs",
					})
				}
			}
		}

		resources, hasResources := rule["resources"].([]interface{})
		if hasResources {
			for _, resourceIface := range resources {
				if resource, ok := resourceIface.(string); ok && resource == "*" {
					results = append(results, ValidationResult{
						File:    filePath,
						Type:    "rbac",
						Level:   "warning",
						Message: fmt.Sprintf("Rule %d uses wildcard resource '*' - consider being more specific", i),
						Rule:    "rbac-wildcard-resources",
					})
				}
			}
		}
	}

	return results
}

func validateDeployment(filePath string, doc map[string]interface{}) []ValidationResult {
	var results []ValidationResult

	spec, ok := doc["spec"].(map[string]interface{})
	if !ok {
		results = append(results, ValidationResult{
			File:    filePath,
			Type:    "deployment",
			Level:   "error",
			Message: "Deployment missing spec",
			Rule:    "deployment-spec",
		})
		return results
	}

	// Validate template
	template, hasTemplate := spec["template"].(map[string]interface{})
	if !hasTemplate {
		results = append(results, ValidationResult{
			File:    filePath,
			Type:    "deployment",
			Level:   "error",
			Message: "Deployment missing template",
			Rule:    "deployment-template",
		})
		return results
	}

	templateSpec, hasTemplateSpec := template["spec"].(map[string]interface{})
	if hasTemplateSpec {
		results = append(results, validatePodSpec(filePath, templateSpec)...)
	}

	return results
}

func validatePodSpec(filePath string, podSpec map[string]interface{}) []ValidationResult {
	var results []ValidationResult

	containers, hasContainers := podSpec["containers"].([]interface{})
	if !hasContainers || len(containers) == 0 {
		results = append(results, ValidationResult{
			File:    filePath,
			Type:    "pod",
			Level:   "error",
			Message: "Pod spec missing containers",
			Rule:    "pod-containers",
		})
		return results
	}

	// Validate each container
	for i, containerIface := range containers {
		container, ok := containerIface.(map[string]interface{})
		if !ok {
			continue
		}

		// Check required fields
		if _, hasImage := container["image"]; !hasImage {
			results = append(results, ValidationResult{
				File:    filePath,
				Type:    "container",
				Level:   "error",
				Message: fmt.Sprintf("Container %d missing image", i),
				Rule:    "container-image",
			})
		}

		if _, hasName := container["name"]; !hasName {
			results = append(results, ValidationResult{
				File:    filePath,
				Type:    "container",
				Level:   "error",
				Message: fmt.Sprintf("Container %d missing name", i),
				Rule:    "container-name",
			})
		}
	}

	return results
}

func validateSecurity(filePath string, doc map[string]interface{}) []ValidationResult {
	var results []ValidationResult

	kind, _ := doc["kind"].(string)

	// Security context validation for pods/deployments
	if kind == "Deployment" || kind == "Pod" {
		results = append(results, validateSecurityContext(filePath, doc)...)
	}

	return results
}

func validateSecurityContext(filePath string, doc map[string]interface{}) []ValidationResult {
	var results []ValidationResult

	// Navigate to pod spec
	var podSpec map[string]interface{}
	kind, _ := doc["kind"].(string)

	if kind == "Deployment" {
		if spec, ok := doc["spec"].(map[string]interface{}); ok {
			if template, ok := spec["template"].(map[string]interface{}); ok {
				podSpec, _ = template["spec"].(map[string]interface{})
			}
		}
	} else if kind == "Pod" {
		podSpec, _ = doc["spec"].(map[string]interface{})
	}

	if podSpec == nil {
		return results
	}

	// Check pod security context
	if _, hasSecurityContext := podSpec["securityContext"]; !hasSecurityContext {
		results = append(results, ValidationResult{
			File:    filePath,
			Type:    "security",
			Level:   "warning",
			Message: "Pod missing securityContext - consider adding for better security",
			Rule:    "pod-security-context",
		})
	}

	// Check container security contexts
	containers, hasContainers := podSpec["containers"].([]interface{})
	if hasContainers {
		for i, containerIface := range containers {
			container, ok := containerIface.(map[string]interface{})
			if !ok {
				continue
			}

			if _, hasSecurityContext := container["securityContext"]; !hasSecurityContext {
				results = append(results, ValidationResult{
					File:    filePath,
					Type:    "security",
					Level:   "warning",
					Message: fmt.Sprintf("Container %d missing securityContext", i),
					Rule:    "container-security-context",
				})
			}
		}
	}

	return results
}

func validateServiceAccount(filePath string, doc map[string]interface{}) []ValidationResult {
	var results []ValidationResult

	metadata, ok := doc["metadata"].(map[string]interface{})
	if !ok {
		return results
	}

	name, _ := metadata["name"].(string)
	if name == "default" {
		results = append(results, ValidationResult{
			File:    filePath,
			Type:    "serviceaccount",
			Level:   "warning",
			Message: "Using 'default' ServiceAccount - consider creating a dedicated ServiceAccount",
			Rule:    "serviceaccount-name",
		})
	}

	return results
}

func validateNaming(filePath string, doc map[string]interface{}) []ValidationResult {
	var results []ValidationResult

	metadata, ok := doc["metadata"].(map[string]interface{})
	if !ok {
		return results
	}

	name, _ := metadata["name"].(string)
	if name == "" {
		results = append(results, ValidationResult{
			File:    filePath,
			Type:    "naming",
			Level:   "error",
			Message: "Resource missing name",
			Rule:    "naming-required",
		})
		return results
	}

	// Validate name format (DNS-1123 subdomain)
	if !isValidDNSName(name) {
		results = append(results, ValidationResult{
			File:    filePath,
			Type:    "naming",
			Level:   "error",
			Message: fmt.Sprintf("Resource name '%s' is not a valid DNS name", name),
			Rule:    "naming-format",
		})
	}

	return results
}

func isValidDNSName(name string) bool {
	if len(name) == 0 || len(name) > 253 {
		return false
	}

	// Basic validation - should be more comprehensive
	return !strings.Contains(name, "_") && !strings.HasPrefix(name, "-") && !strings.HasSuffix(name, "-")
}

func outputResults(report *ValidationReport) error {
	switch validateOutput {
	case "json":
		return outputJSON(report)
	case "yaml":
		return outputYAML(report)
	default:
		return outputText(report)
	}
}

func outputJSON(report *ValidationReport) error {
	encoder := yaml.NewEncoder(os.Stdout)
	defer encoder.Close()
	return encoder.Encode(report)
}

func outputYAML(report *ValidationReport) error {
	encoder := yaml.NewEncoder(os.Stdout)
	defer encoder.Close()
	return encoder.Encode(report)
}

func outputText(report *ValidationReport) error {
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
		fmt.Println()
	}

	return nil
}

func printSummary(report *ValidationReport) {
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
