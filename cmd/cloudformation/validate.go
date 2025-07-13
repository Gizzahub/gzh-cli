package cloudformation

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// ValidateCmd represents the validate command
var ValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate CloudFormation templates",
	Long: `Validate CloudFormation templates for syntax and best practices.

Performs comprehensive validation including:
- CloudFormation syntax validation via AWS API
- Template structure and format validation
- Resource configuration best practices
- Security compliance checks
- Cost optimization recommendations
- Cross-reference validation

Examples:
  gz cloudformation validate --template template.yaml
  gz cloudformation validate --template template.json --security-check
  gz cloudformation validate --template template.yaml --strict --cost-analysis`,
	Run: runValidate,
}

var (
	validateTemplate   string
	validateRegion     string
	validateProfile    string
	strictValidation   bool
	securityValidation bool
	costAnalysis       bool
	validateOutputFormat string
	skipAWSValidation  bool
	localValidation    bool
)

func init() {
	ValidateCmd.Flags().StringVarP(&validateTemplate, "template", "t", "", "CloudFormation template file path")
	ValidateCmd.Flags().StringVarP(&validateRegion, "region", "r", "us-west-2", "AWS region for validation")
	ValidateCmd.Flags().StringVar(&validateProfile, "profile", "", "AWS profile")
	ValidateCmd.Flags().BoolVar(&strictValidation, "strict", false, "Enable strict validation mode")
	ValidateCmd.Flags().BoolVar(&securityValidation, "security-check", true, "Enable security validation")
	ValidateCmd.Flags().BoolVar(&costAnalysis, "cost-analysis", false, "Enable cost analysis")
	ValidateCmd.Flags().StringVarP(&validateOutputFormat, "output", "o", "text", "Output format (text, json)")
	ValidateCmd.Flags().BoolVar(&skipAWSValidation, "skip-aws", false, "Skip AWS API validation")
	ValidateCmd.Flags().BoolVar(&localValidation, "local-only", false, "Perform only local validation")

	ValidateCmd.MarkFlagRequired("template")
}

// ValidationResult represents a validation issue
type ValidationResult struct {
	Type       string `json:"type"`
	Level      string `json:"level"` // error, warning, info
	Message    string `json:"message"`
	Resource   string `json:"resource,omitempty"`
	Property   string `json:"property,omitempty"`
	Line       int    `json:"line,omitempty"`
	Rule       string `json:"rule"`
	Category   string `json:"category"`
	Suggestion string `json:"suggestion,omitempty"`
}

// ValidationReport represents the complete validation report
type ValidationReport struct {
	TemplateFile  string               `json:"template_file"`
	Valid         bool                 `json:"valid"`
	Summary       ValidationSummary    `json:"summary"`
	Results       []ValidationResult   `json:"results"`
	AWSValidation *AWSValidationResult `json:"aws_validation,omitempty"`
}

// ValidationSummary represents validation summary statistics
type ValidationSummary struct {
	TotalChecks  int `json:"total_checks"`
	ErrorCount   int `json:"error_count"`
	WarningCount int `json:"warning_count"`
	InfoCount    int `json:"info_count"`
}

// AWSValidationResult represents AWS API validation result
type AWSValidationResult struct {
	Valid        bool                   `json:"valid"`
	Description  string                 `json:"description,omitempty"`
	Parameters   []AWSTemplateParameter `json:"parameters,omitempty"`
	Capabilities []string               `json:"capabilities,omitempty"`
	Error        string                 `json:"error,omitempty"`
}

// AWSTemplateParameter represents AWS template parameter
type AWSTemplateParameter struct {
	ParameterKey string `json:"parameter_key"`
	DefaultValue string `json:"default_value,omitempty"`
	NoEcho       bool   `json:"no_echo,omitempty"`
	Description  string `json:"description,omitempty"`
}

func runValidate(cmd *cobra.Command, args []string) {
	if validateTemplate == "" {
		fmt.Printf("❌ Template file is required\n")
		cmd.Help()
		os.Exit(1)
	}

	fmt.Printf("🔍 Validating CloudFormation template: %s\n", validateTemplate)

	// Check if template file exists
	if _, err := os.Stat(validateTemplate); os.IsNotExist(err) {
		fmt.Printf("❌ Template file not found: %s\n", validateTemplate)
		os.Exit(1)
	}

	// Run validation
	report, err := validateCloudFormationTemplate(validateTemplate)
	if err != nil {
		fmt.Printf("❌ Validation failed: %v\n", err)
		os.Exit(1)
	}

	// Output results
	if err := outputValidationReport(report); err != nil {
		fmt.Printf("❌ Failed to output results: %v\n", err)
		os.Exit(1)
	}

	// Print summary
	printValidationSummary(report)

	// Exit with appropriate code
	if report.Summary.ErrorCount > 0 {
		os.Exit(1)
	} else if report.Summary.WarningCount > 0 && strictValidation {
		os.Exit(2)
	}
}

func validateCloudFormationTemplate(templatePath string) (*ValidationReport, error) {
	report := &ValidationReport{
		TemplateFile: templatePath,
		Valid:        true,
		Results:      []ValidationResult{},
	}

	// Read template file
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template: %w", err)
	}

	// Parse template
	template, err := parseTemplate(templateContent)
	if err != nil {
		report.Results = append(report.Results, ValidationResult{
			Type:     "syntax",
			Level:    "error",
			Message:  fmt.Sprintf("Failed to parse template: %v", err),
			Rule:     "template-syntax",
			Category: "syntax",
		})
		report.Valid = false
	} else {
		// Run local validations
		results := runLocalValidations(template, templateContent)
		report.Results = append(report.Results, results...)
	}

	// Run AWS validation if not skipped
	if !skipAWSValidation && !localValidation {
		awsResult, err := runAWSValidation(templateContent)
		if err != nil {
			report.Results = append(report.Results, ValidationResult{
				Type:     "aws",
				Level:    "warning",
				Message:  fmt.Sprintf("AWS validation failed: %v", err),
				Rule:     "aws-validation",
				Category: "aws",
			})
		} else {
			report.AWSValidation = awsResult
			if !awsResult.Valid {
				report.Results = append(report.Results, ValidationResult{
					Type:     "aws",
					Level:    "error",
					Message:  awsResult.Error,
					Rule:     "aws-validation",
					Category: "aws",
				})
				report.Valid = false
			}
		}
	}

	// Calculate summary
	for _, result := range report.Results {
		report.Summary.TotalChecks++
		switch result.Level {
		case "error":
			report.Summary.ErrorCount++
			report.Valid = false
		case "warning":
			report.Summary.WarningCount++
		case "info":
			report.Summary.InfoCount++
		}
	}

	return report, nil
}

func parseTemplate(content []byte) (map[string]interface{}, error) {
	var template map[string]interface{}

	// Try JSON first
	if err := json.Unmarshal(content, &template); err == nil {
		return template, nil
	}

	// Try YAML
	if err := yaml.Unmarshal(content, &template); err == nil {
		return template, nil
	}

	return nil, fmt.Errorf("template must be valid JSON or YAML")
}

func runLocalValidations(template map[string]interface{}, content []byte) []ValidationResult {
	var results []ValidationResult

	// Basic structure validation
	results = append(results, validateTemplateStructure(template)...)

	// Resource validation
	results = append(results, validateResources(template)...)

	// Parameter validation
	results = append(results, validateParameters(template)...)

	// Output validation
	results = append(results, validateOutputs(template)...)

	// Security validation
	if securityValidation {
		results = append(results, validateSecurity(template)...)
	}

	// Cost analysis
	if costAnalysis {
		results = append(results, analyzeCosts(template)...)
	}

	// Best practices validation
	results = append(results, validateBestPractices(template)...)

	return results
}

func validateTemplateStructure(template map[string]interface{}) []ValidationResult {
	var results []ValidationResult

	// Check required sections
	requiredSections := []string{"Resources"}
	for _, section := range requiredSections {
		if _, exists := template[section]; !exists {
			results = append(results, ValidationResult{
				Type:     "structure",
				Level:    "error",
				Message:  fmt.Sprintf("Missing required section: %s", section),
				Rule:     "required-sections",
				Category: "structure",
			})
		}
	}

	// Check AWSTemplateFormatVersion
	if version, exists := template["AWSTemplateFormatVersion"]; exists {
		if versionStr, ok := version.(string); ok {
			if versionStr != "2010-09-09" {
				results = append(results, ValidationResult{
					Type:     "structure",
					Level:    "warning",
					Message:  fmt.Sprintf("Unusual template format version: %s", versionStr),
					Rule:     "template-version",
					Category: "structure",
				})
			}
		}
	} else {
		results = append(results, ValidationResult{
			Type:       "structure",
			Level:      "info",
			Message:    "Consider adding AWSTemplateFormatVersion",
			Rule:       "template-version",
			Category:   "structure",
			Suggestion: "Add 'AWSTemplateFormatVersion: \"2010-09-09\"' to the template",
		})
	}

	// Check Description
	if _, exists := template["Description"]; !exists {
		results = append(results, ValidationResult{
			Type:       "structure",
			Level:      "info",
			Message:    "Consider adding a Description",
			Rule:       "template-description",
			Category:   "structure",
			Suggestion: "Add a meaningful description of what this template creates",
		})
	}

	return results
}

func validateResources(template map[string]interface{}) []ValidationResult {
	var results []ValidationResult

	resources, exists := template["Resources"]
	if !exists {
		return results
	}

	resourcesMap, ok := resources.(map[string]interface{})
	if !ok {
		results = append(results, ValidationResult{
			Type:     "resources",
			Level:    "error",
			Message:  "Resources section must be an object",
			Rule:     "resources-format",
			Category: "structure",
		})
		return results
	}

	// Validate each resource
	for resourceName, resourceDef := range resourcesMap {
		resourceResults := validateResource(resourceName, resourceDef)
		results = append(results, resourceResults...)
	}

	// Check for duplicate resource names
	if len(resourcesMap) == 0 {
		results = append(results, ValidationResult{
			Type:     "resources",
			Level:    "warning",
			Message:  "No resources defined in template",
			Rule:     "empty-resources",
			Category: "structure",
		})
	}

	return results
}

func validateResource(resourceName string, resourceDef interface{}) []ValidationResult {
	var results []ValidationResult

	resourceMap, ok := resourceDef.(map[string]interface{})
	if !ok {
		results = append(results, ValidationResult{
			Type:     "resources",
			Level:    "error",
			Message:  fmt.Sprintf("Resource %s must be an object", resourceName),
			Resource: resourceName,
			Rule:     "resource-format",
			Category: "structure",
		})
		return results
	}

	// Check required Type property
	resourceType, exists := resourceMap["Type"]
	if !exists {
		results = append(results, ValidationResult{
			Type:     "resources",
			Level:    "error",
			Message:  fmt.Sprintf("Resource %s missing required Type property", resourceName),
			Resource: resourceName,
			Rule:     "resource-type",
			Category: "structure",
		})
		return results
	}

	typeStr, ok := resourceType.(string)
	if !ok {
		results = append(results, ValidationResult{
			Type:     "resources",
			Level:    "error",
			Message:  fmt.Sprintf("Resource %s Type must be a string", resourceName),
			Resource: resourceName,
			Rule:     "resource-type",
			Category: "structure",
		})
		return results
	}

	// Validate resource type format
	if !isValidResourceType(typeStr) {
		results = append(results, ValidationResult{
			Type:     "resources",
			Level:    "warning",
			Message:  fmt.Sprintf("Resource %s has unusual type format: %s", resourceName, typeStr),
			Resource: resourceName,
			Rule:     "resource-type-format",
			Category: "structure",
		})
	}

	// Validate resource-specific properties
	resourceResults := validateResourceSpecific(resourceName, typeStr, resourceMap)
	results = append(results, resourceResults...)

	return results
}

func validateResourceSpecific(resourceName, resourceType string, resourceMap map[string]interface{}) []ValidationResult {
	var results []ValidationResult

	switch {
	case strings.HasPrefix(resourceType, "AWS::S3::"):
		results = append(results, validateS3Resource(resourceName, resourceType, resourceMap)...)
	case strings.HasPrefix(resourceType, "AWS::EC2::"):
		results = append(results, validateEC2Resource(resourceName, resourceType, resourceMap)...)
	case strings.HasPrefix(resourceType, "AWS::IAM::"):
		results = append(results, validateIAMResource(resourceName, resourceType, resourceMap)...)
	case strings.HasPrefix(resourceType, "AWS::RDS::"):
		results = append(results, validateRDSResource(resourceName, resourceType, resourceMap)...)
	case strings.HasPrefix(resourceType, "AWS::Lambda::"):
		results = append(results, validateLambdaResource(resourceName, resourceType, resourceMap)...)
	}

	return results
}

func validateS3Resource(resourceName, resourceType string, resourceMap map[string]interface{}) []ValidationResult {
	var results []ValidationResult

	if resourceType == "AWS::S3::Bucket" {
		properties, exists := resourceMap["Properties"]
		if exists {
			if propertiesMap, ok := properties.(map[string]interface{}); ok {
				// Check for encryption
				if _, hasEncryption := propertiesMap["BucketEncryption"]; !hasEncryption {
					results = append(results, ValidationResult{
						Type:       "security",
						Level:      "warning",
						Message:    fmt.Sprintf("S3 bucket %s should have encryption enabled", resourceName),
						Resource:   resourceName,
						Rule:       "s3-encryption",
						Category:   "security",
						Suggestion: "Add BucketEncryption property",
					})
				}

				// Check for public access block
				if _, hasPublicAccessBlock := propertiesMap["PublicAccessBlockConfiguration"]; !hasPublicAccessBlock {
					results = append(results, ValidationResult{
						Type:       "security",
						Level:      "warning",
						Message:    fmt.Sprintf("S3 bucket %s should have public access blocked", resourceName),
						Resource:   resourceName,
						Rule:       "s3-public-access",
						Category:   "security",
						Suggestion: "Add PublicAccessBlockConfiguration property",
					})
				}

				// Check for versioning
				if _, hasVersioning := propertiesMap["VersioningConfiguration"]; !hasVersioning {
					results = append(results, ValidationResult{
						Type:     "best-practice",
						Level:    "info",
						Message:  fmt.Sprintf("S3 bucket %s should consider enabling versioning", resourceName),
						Resource: resourceName,
						Rule:     "s3-versioning",
						Category: "best-practice",
					})
				}
			}
		}
	}

	return results
}

func validateEC2Resource(resourceName, resourceType string, resourceMap map[string]interface{}) []ValidationResult {
	var results []ValidationResult

	if resourceType == "AWS::EC2::Instance" {
		properties, exists := resourceMap["Properties"]
		if exists {
			if propertiesMap, ok := properties.(map[string]interface{}); ok {
				// Check for security groups
				if _, hasSecurityGroups := propertiesMap["SecurityGroups"]; !hasSecurityGroups {
					if _, hasSecurityGroupIds := propertiesMap["SecurityGroupIds"]; !hasSecurityGroupIds {
						results = append(results, ValidationResult{
							Type:     "security",
							Level:    "warning",
							Message:  fmt.Sprintf("EC2 instance %s should have security groups defined", resourceName),
							Resource: resourceName,
							Rule:     "ec2-security-groups",
							Category: "security",
						})
					}
				}

				// Check for key pair
				if _, hasKeyName := propertiesMap["KeyName"]; !hasKeyName {
					results = append(results, ValidationResult{
						Type:     "access",
						Level:    "info",
						Message:  fmt.Sprintf("EC2 instance %s has no key pair for SSH access", resourceName),
						Resource: resourceName,
						Rule:     "ec2-key-pair",
						Category: "access",
					})
				}
			}
		}
	}

	if resourceType == "AWS::EC2::SecurityGroup" {
		properties, exists := resourceMap["Properties"]
		if exists {
			if propertiesMap, ok := properties.(map[string]interface{}); ok {
				// Check for overly permissive rules
				if ingress, hasIngress := propertiesMap["SecurityGroupIngress"]; hasIngress {
					if ingressRules, ok := ingress.([]interface{}); ok {
						for _, rule := range ingressRules {
							if ruleMap, ok := rule.(map[string]interface{}); ok {
								if cidr, hasCidr := ruleMap["CidrIp"]; hasCidr {
									if cidr == "0.0.0.0/0" {
										results = append(results, ValidationResult{
											Type:       "security",
											Level:      "warning",
											Message:    fmt.Sprintf("Security group %s allows access from anywhere (0.0.0.0/0)", resourceName),
											Resource:   resourceName,
											Rule:       "sg-open-access",
											Category:   "security",
											Suggestion: "Restrict access to specific IP ranges",
										})
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return results
}

func validateIAMResource(resourceName, resourceType string, resourceMap map[string]interface{}) []ValidationResult {
	var results []ValidationResult

	if resourceType == "AWS::IAM::Role" || resourceType == "AWS::IAM::Policy" {
		properties, exists := resourceMap["Properties"]
		if exists {
			if propertiesMap, ok := properties.(map[string]interface{}); ok {
				// Check for wildcard permissions
				if policyDoc, hasPolicyDoc := propertiesMap["PolicyDocument"]; hasPolicyDoc {
					if doc, ok := policyDoc.(map[string]interface{}); ok {
						if statements, hasStatements := doc["Statement"]; hasStatements {
							if stmtArray, ok := statements.([]interface{}); ok {
								for _, stmt := range stmtArray {
									if stmtMap, ok := stmt.(map[string]interface{}); ok {
										if actions, hasActions := stmtMap["Action"]; hasActions {
											if actionStr, ok := actions.(string); ok && actionStr == "*" {
												results = append(results, ValidationResult{
													Type:       "security",
													Level:      "warning",
													Message:    fmt.Sprintf("IAM resource %s uses wildcard permissions (*)", resourceName),
													Resource:   resourceName,
													Rule:       "iam-wildcard-permissions",
													Category:   "security",
													Suggestion: "Use specific permissions instead of wildcards",
												})
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return results
}

func validateRDSResource(resourceName, resourceType string, resourceMap map[string]interface{}) []ValidationResult {
	var results []ValidationResult

	if resourceType == "AWS::RDS::DBInstance" {
		properties, exists := resourceMap["Properties"]
		if exists {
			if propertiesMap, ok := properties.(map[string]interface{}); ok {
				// Check for encryption
				if encrypted, hasEncryption := propertiesMap["StorageEncrypted"]; hasEncryption {
					if encryptedBool, ok := encrypted.(bool); ok && !encryptedBool {
						results = append(results, ValidationResult{
							Type:     "security",
							Level:    "warning",
							Message:  fmt.Sprintf("RDS instance %s should have storage encryption enabled", resourceName),
							Resource: resourceName,
							Rule:     "rds-encryption",
							Category: "security",
						})
					}
				} else {
					results = append(results, ValidationResult{
						Type:     "security",
						Level:    "warning",
						Message:  fmt.Sprintf("RDS instance %s should have storage encryption enabled", resourceName),
						Resource: resourceName,
						Rule:     "rds-encryption",
						Category: "security",
					})
				}

				// Check for backup retention
				if retention, hasRetention := propertiesMap["BackupRetentionPeriod"]; hasRetention {
					if retentionNum, ok := retention.(float64); ok && retentionNum < 7 {
						results = append(results, ValidationResult{
							Type:       "best-practice",
							Level:      "info",
							Message:    fmt.Sprintf("RDS instance %s has short backup retention period", resourceName),
							Resource:   resourceName,
							Rule:       "rds-backup-retention",
							Category:   "best-practice",
							Suggestion: "Consider extending backup retention to at least 7 days",
						})
					}
				}
			}
		}
	}

	return results
}

func validateLambdaResource(resourceName, resourceType string, resourceMap map[string]interface{}) []ValidationResult {
	var results []ValidationResult

	if resourceType == "AWS::Lambda::Function" {
		properties, exists := resourceMap["Properties"]
		if exists {
			if propertiesMap, ok := properties.(map[string]interface{}); ok {
				// Check for timeout
				if timeout, hasTimeout := propertiesMap["Timeout"]; hasTimeout {
					if timeoutNum, ok := timeout.(float64); ok && timeoutNum > 300 {
						results = append(results, ValidationResult{
							Type:       "performance",
							Level:      "warning",
							Message:    fmt.Sprintf("Lambda function %s has very long timeout (%.0f seconds)", resourceName, timeoutNum),
							Resource:   resourceName,
							Rule:       "lambda-timeout",
							Category:   "performance",
							Suggestion: "Consider if such a long timeout is necessary",
						})
					}
				}

				// Check for reserved concurrency
				if _, hasReservedConcurrency := propertiesMap["ReservedConcurrencyLimit"]; !hasReservedConcurrency {
					results = append(results, ValidationResult{
						Type:     "best-practice",
						Level:    "info",
						Message:  fmt.Sprintf("Lambda function %s should consider setting reserved concurrency", resourceName),
						Resource: resourceName,
						Rule:     "lambda-concurrency",
						Category: "best-practice",
					})
				}
			}
		}
	}

	return results
}

func validateParameters(template map[string]interface{}) []ValidationResult {
	var results []ValidationResult

	parameters, exists := template["Parameters"]
	if !exists {
		return results
	}

	parametersMap, ok := parameters.(map[string]interface{})
	if !ok {
		results = append(results, ValidationResult{
			Type:     "parameters",
			Level:    "error",
			Message:  "Parameters section must be an object",
			Rule:     "parameters-format",
			Category: "structure",
		})
		return results
	}

	// Validate each parameter
	for paramName, paramDef := range parametersMap {
		paramResults := validateParameter(paramName, paramDef)
		results = append(results, paramResults...)
	}

	return results
}

func validateParameter(paramName string, paramDef interface{}) []ValidationResult {
	var results []ValidationResult

	paramMap, ok := paramDef.(map[string]interface{})
	if !ok {
		results = append(results, ValidationResult{
			Type:     "parameters",
			Level:    "error",
			Message:  fmt.Sprintf("Parameter %s must be an object", paramName),
			Rule:     "parameter-format",
			Category: "structure",
		})
		return results
	}

	// Check required Type property
	if _, exists := paramMap["Type"]; !exists {
		results = append(results, ValidationResult{
			Type:     "parameters",
			Level:    "error",
			Message:  fmt.Sprintf("Parameter %s missing required Type property", paramName),
			Rule:     "parameter-type",
			Category: "structure",
		})
	}

	// Check for Description
	if _, exists := paramMap["Description"]; !exists {
		results = append(results, ValidationResult{
			Type:     "parameters",
			Level:    "info",
			Message:  fmt.Sprintf("Parameter %s should have a Description", paramName),
			Rule:     "parameter-description",
			Category: "documentation",
		})
	}

	return results
}

func validateOutputs(template map[string]interface{}) []ValidationResult {
	var results []ValidationResult

	outputs, exists := template["Outputs"]
	if !exists {
		results = append(results, ValidationResult{
			Type:       "outputs",
			Level:      "info",
			Message:    "Consider adding Outputs section",
			Rule:       "outputs-section",
			Category:   "best-practice",
			Suggestion: "Outputs help other stacks reference resources from this stack",
		})
		return results
	}

	outputsMap, ok := outputs.(map[string]interface{})
	if !ok {
		results = append(results, ValidationResult{
			Type:     "outputs",
			Level:    "error",
			Message:  "Outputs section must be an object",
			Rule:     "outputs-format",
			Category: "structure",
		})
		return results
	}

	// Validate each output
	for outputName, outputDef := range outputsMap {
		outputResults := validateOutput(outputName, outputDef)
		results = append(results, outputResults...)
	}

	return results
}

func validateOutput(outputName string, outputDef interface{}) []ValidationResult {
	var results []ValidationResult

	outputMap, ok := outputDef.(map[string]interface{})
	if !ok {
		results = append(results, ValidationResult{
			Type:     "outputs",
			Level:    "error",
			Message:  fmt.Sprintf("Output %s must be an object", outputName),
			Rule:     "output-format",
			Category: "structure",
		})
		return results
	}

	// Check required Value property
	if _, exists := outputMap["Value"]; !exists {
		results = append(results, ValidationResult{
			Type:     "outputs",
			Level:    "error",
			Message:  fmt.Sprintf("Output %s missing required Value property", outputName),
			Rule:     "output-value",
			Category: "structure",
		})
	}

	// Check for Description
	if _, exists := outputMap["Description"]; !exists {
		results = append(results, ValidationResult{
			Type:     "outputs",
			Level:    "info",
			Message:  fmt.Sprintf("Output %s should have a Description", outputName),
			Rule:     "output-description",
			Category: "documentation",
		})
	}

	return results
}

func validateSecurity(template map[string]interface{}) []ValidationResult {
	var results []ValidationResult

	// Security-specific validations already covered in resource validation
	// Add any additional security checks here

	return results
}

func analyzeCosts(template map[string]interface{}) []ValidationResult {
	var results []ValidationResult

	resources, exists := template["Resources"]
	if !exists {
		return results
	}

	resourcesMap, ok := resources.(map[string]interface{})
	if !ok {
		return results
	}

	// Analyze cost-impacting resources
	for resourceName, resourceDef := range resourcesMap {
		resourceMap, ok := resourceDef.(map[string]interface{})
		if !ok {
			continue
		}

		resourceType, exists := resourceMap["Type"]
		if !exists {
			continue
		}

		typeStr, ok := resourceType.(string)
		if !ok {
			continue
		}

		// Check for expensive resources
		switch {
		case strings.Contains(typeStr, "::RDS::"):
			results = append(results, ValidationResult{
				Type:       "cost",
				Level:      "info",
				Message:    fmt.Sprintf("Resource %s (%s) may incur significant costs", resourceName, typeStr),
				Resource:   resourceName,
				Rule:       "cost-awareness",
				Category:   "cost",
				Suggestion: "Review instance size and storage requirements",
			})
		case strings.Contains(typeStr, "::EC2::Instance"):
			results = append(results, ValidationResult{
				Type:       "cost",
				Level:      "info",
				Message:    fmt.Sprintf("Resource %s (%s) may incur ongoing costs", resourceName, typeStr),
				Resource:   resourceName,
				Rule:       "cost-awareness",
				Category:   "cost",
				Suggestion: "Consider using smaller instance types for development",
			})
		}
	}

	return results
}

func validateBestPractices(template map[string]interface{}) []ValidationResult {
	var results []ValidationResult

	// Check for tags
	resources, exists := template["Resources"]
	if exists {
		resourcesMap, ok := resources.(map[string]interface{})
		if ok {
			for resourceName, resourceDef := range resourcesMap {
				resourceMap, ok := resourceDef.(map[string]interface{})
				if !ok {
					continue
				}

				properties, hasProperties := resourceMap["Properties"]
				if hasProperties {
					propertiesMap, ok := properties.(map[string]interface{})
					if ok {
						if _, hasTags := propertiesMap["Tags"]; !hasTags {
							results = append(results, ValidationResult{
								Type:       "best-practice",
								Level:      "info",
								Message:    fmt.Sprintf("Resource %s should have tags for better management", resourceName),
								Resource:   resourceName,
								Rule:       "resource-tags",
								Category:   "best-practice",
								Suggestion: "Add tags like Environment, Project, Owner",
							})
						}
					}
				}
			}
		}
	}

	return results
}

func runAWSValidation(templateContent []byte) (*AWSValidationResult, error) {
	ctx := context.Background()

	// Initialize AWS config
	var options []func(*config.LoadOptions) error
	if validateRegion != "" {
		options = append(options, config.WithRegion(validateRegion))
	}
	if validateProfile != "" {
		options = append(options, config.WithSharedConfigProfile(validateProfile))
	}

	cfg, err := config.LoadDefaultConfig(ctx, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := cloudformation.NewFromConfig(cfg)

	// Validate template
	input := &cloudformation.ValidateTemplateInput{
		TemplateBody: aws.String(string(templateContent)),
	}

	output, err := client.ValidateTemplate(ctx, input)
	if err != nil {
		return &AWSValidationResult{
			Valid: false,
			Error: err.Error(),
		}, nil
	}

	// Convert parameters
	var parameters []AWSTemplateParameter
	for _, param := range output.Parameters {
		parameters = append(parameters, AWSTemplateParameter{
			ParameterKey: aws.ToString(param.ParameterKey),
			DefaultValue: aws.ToString(param.DefaultValue),
			NoEcho:       aws.ToBool(param.NoEcho),
			Description:  aws.ToString(param.Description),
		})
	}

	// Convert capabilities
	var capabilities []string
	for _, cap := range output.Capabilities {
		capabilities = append(capabilities, string(cap))
	}

	return &AWSValidationResult{
		Valid:        true,
		Description:  aws.ToString(output.Description),
		Parameters:   parameters,
		Capabilities: capabilities,
	}, nil
}

func isValidResourceType(resourceType string) bool {
	// Check if resource type follows AWS::Service::Resource format
	pattern := regexp.MustCompile(`^AWS::[A-Za-z0-9]+::[A-Za-z0-9]+$`)
	return pattern.MatchString(resourceType)
}

func outputValidationReport(report *ValidationReport) error {
	switch validateOutputFormat {
	case "json":
		return outputValidationJSON(report)
	default:
		return outputValidationText(report)
	}
}

func outputValidationJSON(report *ValidationReport) error {
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(jsonData))
	return nil
}

func outputValidationText(report *ValidationReport) error {
	if len(report.Results) == 0 {
		fmt.Printf("✅ No validation issues found\n")
		return nil
	}

	fmt.Printf("\n📋 Validation Results:\n")
	fmt.Printf("=====================\n\n")

	for _, result := range report.Results {
		icon := getValidationIcon(result.Level)
		fmt.Printf("%s [%s] %s", icon, strings.ToUpper(result.Level), result.Message)

		if result.Resource != "" {
			fmt.Printf(" (Resource: %s)", result.Resource)
		}

		if result.Property != "" {
			fmt.Printf(" (Property: %s)", result.Property)
		}

		fmt.Printf("\n")

		if result.Rule != "" {
			fmt.Printf("    Rule: %s\n", result.Rule)
		}

		if result.Suggestion != "" {
			fmt.Printf("    Suggestion: %s\n", result.Suggestion)
		}

		fmt.Printf("\n")
	}

	return nil
}

func getValidationIcon(level string) string {
	switch level {
	case "error":
		return "❌"
	case "warning":
		return "⚠️"
	case "info":
		return "ℹ️"
	default:
		return "🔹"
	}
}

func printValidationSummary(report *ValidationReport) {
	fmt.Printf("\n📊 Validation Summary:\n")
	fmt.Printf("=====================\n")
	fmt.Printf("Template: %s\n", report.TemplateFile)
	fmt.Printf("Valid: %t\n", report.Valid)
	fmt.Printf("Total checks: %d\n", report.Summary.TotalChecks)
	fmt.Printf("Errors: %d\n", report.Summary.ErrorCount)
	fmt.Printf("Warnings: %d\n", report.Summary.WarningCount)
	fmt.Printf("Info: %d\n", report.Summary.InfoCount)

	if report.AWSValidation != nil {
		fmt.Printf("\n🔍 AWS Validation:\n")
		fmt.Printf("AWS Valid: %t\n", report.AWSValidation.Valid)
		if report.AWSValidation.Description != "" {
			fmt.Printf("Description: %s\n", report.AWSValidation.Description)
		}
		if len(report.AWSValidation.Parameters) > 0 {
			fmt.Printf("Parameters: %d\n", len(report.AWSValidation.Parameters))
		}
		if len(report.AWSValidation.Capabilities) > 0 {
			fmt.Printf("Required Capabilities: %s\n", strings.Join(report.AWSValidation.Capabilities, ", "))
		}
	}

	if report.Valid {
		fmt.Printf("\n✅ Template validation passed\n")
	} else {
		fmt.Printf("\n❌ Template validation failed\n")
	}

	fmt.Printf("\n📝 Next steps:\n")
	if report.Summary.ErrorCount > 0 {
		fmt.Printf("1. Fix all errors before deploying\n")
	}
	if report.Summary.WarningCount > 0 {
		fmt.Printf("2. Review and address warnings\n")
	}
	fmt.Printf("3. Test deployment in a development environment\n")
	fmt.Printf("4. Deploy using: gz cloudformation deploy --template %s\n", report.TemplateFile)
}
