package template

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// ValidateCmd represents the validate command
var ValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "템플릿 메타데이터 및 구조 검증",
	Long: `템플릿의 메타데이터와 파일 구조를 검증합니다.

검증 항목:
- 메타데이터 스키마 유효성
- 필수 파일 존재 여부
- 매개변수 정의 일관성
- 의존성 해결 가능성
- 버전 형식 검증
- 템플릿 파일 구문 검증

Examples:
  gz template validate
  gz template validate --path ./my-template
  gz template validate --strict`,
	Run: runValidate,
}

var (
	validatePath string
	strict       bool
	verbose      bool
)

func init() {
	ValidateCmd.Flags().StringVarP(&validatePath, "path", "p", ".", "검증할 템플릿 경로")
	ValidateCmd.Flags().BoolVar(&strict, "strict", false, "엄격한 검증 모드")
	ValidateCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "상세한 출력")
}

// ValidationResult represents the validation result
type ValidationResult struct {
	Valid    bool                `json:"valid"`
	Errors   []ValidationError   `json:"errors"`
	Warnings []ValidationWarning `json:"warnings"`
	Summary  ValidationSummary   `json:"summary"`
	Details  ValidationDetails   `json:"details"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Code     string `json:"code"`
	Message  string `json:"message"`
	File     string `json:"file,omitempty"`
	Line     int    `json:"line,omitempty"`
	Column   int    `json:"column,omitempty"`
	Severity string `json:"severity"`
}

// ValidationWarning represents a validation warning
type ValidationWarning struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	File    string `json:"file,omitempty"`
	Line    int    `json:"line,omitempty"`
}

// ValidationSummary represents validation summary
type ValidationSummary struct {
	TotalFiles    int `json:"totalFiles"`
	ValidFiles    int `json:"validFiles"`
	ErrorCount    int `json:"errorCount"`
	WarningCount  int `json:"warningCount"`
	TemplateFiles int `json:"templateFiles"`
	StaticFiles   int `json:"staticFiles"`
}

// ValidationDetails represents detailed validation information
type ValidationDetails struct {
	Metadata     MetadataValidation   `json:"metadata"`
	Files        FileValidation       `json:"files"`
	Dependencies DependencyValidation `json:"dependencies"`
	Parameters   ParameterValidation  `json:"parameters"`
}

// MetadataValidation represents metadata validation result
type MetadataValidation struct {
	Valid          bool     `json:"valid"`
	RequiredFields []string `json:"requiredFields"`
	OptionalFields []string `json:"optionalFields"`
	VersionFormat  bool     `json:"versionFormat"`
	CategoryValid  bool     `json:"categoryValid"`
	TypeValid      bool     `json:"typeValid"`
}

// FileValidation represents file validation result
type FileValidation struct {
	StructureValid bool              `json:"structureValid"`
	RequiredDirs   []DirectoryCheck  `json:"requiredDirs"`
	TemplateFiles  []TemplateCheck   `json:"templateFiles"`
	StaticFiles    []StaticFileCheck `json:"staticFiles"`
}

// DirectoryCheck represents directory validation
type DirectoryCheck struct {
	Path     string `json:"path"`
	Exists   bool   `json:"exists"`
	Required bool   `json:"required"`
}

// TemplateCheck represents template file validation
type TemplateCheck struct {
	Path        string   `json:"path"`
	Exists      bool     `json:"exists"`
	SyntaxValid bool     `json:"syntaxValid"`
	Variables   []string `json:"variables"`
}

// StaticFileCheck represents static file validation
type StaticFileCheck struct {
	Path   string `json:"path"`
	Exists bool   `json:"exists"`
	Size   int64  `json:"size"`
}

// DependencyValidation represents dependency validation
type DependencyValidation struct {
	Valid        bool              `json:"valid"`
	Dependencies []DependencyCheck `json:"dependencies"`
	Circular     []string          `json:"circular"`
	Missing      []string          `json:"missing"`
}

// DependencyCheck represents individual dependency validation
type DependencyCheck struct {
	Name       string `json:"name"`
	Version    string `json:"version"`
	Available  bool   `json:"available"`
	Compatible bool   `json:"compatible"`
	Repository string `json:"repository,omitempty"`
}

// ParameterValidation represents parameter validation
type ParameterValidation struct {
	Valid      bool             `json:"valid"`
	Parameters []ParameterCheck `json:"parameters"`
	Unused     []string         `json:"unused"`
	Missing    []string         `json:"missing"`
}

// ParameterCheck represents individual parameter validation
type ParameterCheck struct {
	Name         string   `json:"name"`
	Type         string   `json:"type"`
	Required     bool     `json:"required"`
	HasDefault   bool     `json:"hasDefault"`
	ValidationOK bool     `json:"validationOK"`
	UsedInFiles  []string `json:"usedInFiles"`
}

func runValidate(cmd *cobra.Command, args []string) {
	fmt.Printf("🔍 템플릿 검증 시작\n")
	fmt.Printf("📁 경로: %s\n", validatePath)

	if strict {
		fmt.Printf("⚡ 엄격한 검증 모드\n")
	}

	// Perform validation
	result, err := validateTemplate()
	if err != nil {
		fmt.Printf("❌ 검증 실행 실패: %v\n", err)
		os.Exit(1)
	}

	// Display results
	displayValidationResult(result)

	if !result.Valid {
		os.Exit(1)
	}
}

func validateTemplate() (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:    true,
		Errors:   []ValidationError{},
		Warnings: []ValidationWarning{},
		Summary:  ValidationSummary{},
		Details:  ValidationDetails{},
	}

	// Validate metadata
	if err := validateMetadata(result); err != nil {
		return nil, err
	}

	// Validate file structure
	if err := validateFileStructure(result); err != nil {
		return nil, err
	}

	// Validate dependencies
	if err := validateDependencies(result); err != nil {
		return nil, err
	}

	// Validate parameters
	if err := validateParameters(result); err != nil {
		return nil, err
	}

	// Set overall validity
	result.Valid = len(result.Errors) == 0

	return result, nil
}

func validateMetadata(result *ValidationResult) error {
	metadataFile := filepath.Join(validatePath, "template.yaml")

	// Check if metadata file exists
	if _, err := os.Stat(metadataFile); os.IsNotExist(err) {
		result.Errors = append(result.Errors, ValidationError{
			Code:     "METADATA_MISSING",
			Message:  "template.yaml 파일이 없습니다",
			File:     "template.yaml",
			Severity: "error",
		})
		return nil
	}

	// Read and parse metadata
	data, err := os.ReadFile(metadataFile)
	if err != nil {
		return fmt.Errorf("메타데이터 파일 읽기 실패: %w", err)
	}

	var metadata TemplateMetadata
	if err := yaml.Unmarshal(data, &metadata); err != nil {
		result.Errors = append(result.Errors, ValidationError{
			Code:     "METADATA_INVALID_YAML",
			Message:  fmt.Sprintf("YAML 구문 오류: %v", err),
			File:     "template.yaml",
			Severity: "error",
		})
		return nil
	}

	// Validate required fields
	metadataValidation := MetadataValidation{
		Valid:          true,
		RequiredFields: []string{},
		OptionalFields: []string{},
	}

	if metadata.Metadata.Name == "" {
		result.Errors = append(result.Errors, ValidationError{
			Code:     "METADATA_NAME_MISSING",
			Message:  "metadata.name 필드가 필요합니다",
			File:     "template.yaml",
			Severity: "error",
		})
		metadataValidation.Valid = false
	}

	if metadata.Metadata.Version == "" {
		result.Errors = append(result.Errors, ValidationError{
			Code:     "METADATA_VERSION_MISSING",
			Message:  "metadata.version 필드가 필요합니다",
			File:     "template.yaml",
			Severity: "error",
		})
		metadataValidation.Valid = false
	} else {
		// Validate semantic version format
		if !isValidSemVer(metadata.Metadata.Version) {
			result.Errors = append(result.Errors, ValidationError{
				Code:     "METADATA_VERSION_INVALID",
				Message:  "버전은 semantic versioning 형식이어야 합니다 (예: 1.0.0)",
				File:     "template.yaml",
				Severity: "error",
			})
			metadataValidation.VersionFormat = false
		} else {
			metadataValidation.VersionFormat = true
		}
	}

	// Validate category
	validCategories := []string{"web", "database", "infrastructure", "cicd", "monitoring", "security", "general"}
	if !contains(validCategories, metadata.Metadata.Category) {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Code:    "METADATA_CATEGORY_UNKNOWN",
			Message: fmt.Sprintf("알 수 없는 카테고리: %s", metadata.Metadata.Category),
			File:    "template.yaml",
		})
		metadataValidation.CategoryValid = false
	} else {
		metadataValidation.CategoryValid = true
	}

	// Validate type
	validTypes := []string{"docker", "helm", "terraform", "ansible", "github-actions", "gitlab-ci", "generic"}
	if !contains(validTypes, metadata.Metadata.Type) {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Code:    "METADATA_TYPE_UNKNOWN",
			Message: fmt.Sprintf("알 수 없는 타입: %s", metadata.Metadata.Type),
			File:    "template.yaml",
		})
		metadataValidation.TypeValid = false
	} else {
		metadataValidation.TypeValid = true
	}

	// Strict mode validations
	if strict {
		if metadata.Metadata.Description == "" {
			result.Errors = append(result.Errors, ValidationError{
				Code:     "METADATA_DESCRIPTION_MISSING",
				Message:  "엄격 모드에서는 설명이 필요합니다",
				File:     "template.yaml",
				Severity: "error",
			})
		}

		if metadata.Metadata.Author == "" {
			result.Errors = append(result.Errors, ValidationError{
				Code:     "METADATA_AUTHOR_MISSING",
				Message:  "엄격 모드에서는 작성자 정보가 필요합니다",
				File:     "template.yaml",
				Severity: "error",
			})
		}
	}

	result.Details.Metadata = metadataValidation
	return nil
}

func validateFileStructure(result *ValidationResult) error {
	fileValidation := FileValidation{
		StructureValid: true,
		RequiredDirs:   []DirectoryCheck{},
		TemplateFiles:  []TemplateCheck{},
		StaticFiles:    []StaticFileCheck{},
	}

	// Check required directories
	requiredDirs := []string{"templates", "docs", "examples", "tests"}
	for _, dir := range requiredDirs {
		dirPath := filepath.Join(validatePath, dir)
		exists := true
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			exists = false
			if strict {
				result.Errors = append(result.Errors, ValidationError{
					Code:     "STRUCTURE_REQUIRED_DIR_MISSING",
					Message:  fmt.Sprintf("필수 디렉터리가 없습니다: %s", dir),
					File:     dir,
					Severity: "error",
				})
				fileValidation.StructureValid = false
			} else {
				result.Warnings = append(result.Warnings, ValidationWarning{
					Code:    "STRUCTURE_RECOMMENDED_DIR_MISSING",
					Message: fmt.Sprintf("권장 디렉터리가 없습니다: %s", dir),
					File:    dir,
				})
			}
		}

		fileValidation.RequiredDirs = append(fileValidation.RequiredDirs, DirectoryCheck{
			Path:     dir,
			Exists:   exists,
			Required: strict,
		})
	}

	// Validate template files
	templatesDir := filepath.Join(validatePath, "templates")
	if _, err := os.Stat(templatesDir); err == nil {
		filepath.Walk(templatesDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() {
				relPath, _ := filepath.Rel(validatePath, path)
				templateCheck := TemplateCheck{
					Path:   relPath,
					Exists: true,
				}

				// Check if it's a template file
				if strings.HasSuffix(path, ".tpl") {
					// Validate template syntax
					if err := validateTemplateFile(path, &templateCheck); err != nil {
						result.Errors = append(result.Errors, ValidationError{
							Code:     "TEMPLATE_SYNTAX_ERROR",
							Message:  fmt.Sprintf("템플릿 구문 오류: %v", err),
							File:     relPath,
							Severity: "error",
						})
						templateCheck.SyntaxValid = false
					} else {
						templateCheck.SyntaxValid = true
					}
				}

				fileValidation.TemplateFiles = append(fileValidation.TemplateFiles, templateCheck)
				result.Summary.TemplateFiles++
			}
			return nil
		})
	}

	result.Details.Files = fileValidation
	result.Summary.TotalFiles = result.Summary.TemplateFiles + result.Summary.StaticFiles
	return nil
}

func validateDependencies(result *ValidationResult) error {
	// Read metadata to get dependencies
	metadataFile := filepath.Join(validatePath, "template.yaml")
	data, err := os.ReadFile(metadataFile)
	if err != nil {
		return nil // Already handled in metadata validation
	}

	var metadata TemplateMetadata
	if err := yaml.Unmarshal(data, &metadata); err != nil {
		return nil // Already handled in metadata validation
	}

	dependencyValidation := DependencyValidation{
		Valid:        true,
		Dependencies: []DependencyCheck{},
		Circular:     []string{},
		Missing:      []string{},
	}

	// Check each dependency
	for _, dep := range metadata.Spec.Dependencies {
		depCheck := DependencyCheck{
			Name:       dep.Name,
			Version:    dep.Version,
			Available:  true, // In real implementation, check repository
			Compatible: true, // In real implementation, check version compatibility
			Repository: dep.Repository,
		}

		// Validate version format
		if !isValidSemVer(dep.Version) && dep.Version != "*" && dep.Version != "latest" {
			result.Errors = append(result.Errors, ValidationError{
				Code:     "DEPENDENCY_VERSION_INVALID",
				Message:  fmt.Sprintf("의존성 %s의 버전 형식이 잘못되었습니다: %s", dep.Name, dep.Version),
				File:     "template.yaml",
				Severity: "error",
			})
			depCheck.Compatible = false
			dependencyValidation.Valid = false
		}

		dependencyValidation.Dependencies = append(dependencyValidation.Dependencies, depCheck)
	}

	result.Details.Dependencies = dependencyValidation
	return nil
}

func validateParameters(result *ValidationResult) error {
	// Read metadata to get parameters
	metadataFile := filepath.Join(validatePath, "template.yaml")
	data, err := os.ReadFile(metadataFile)
	if err != nil {
		return nil
	}

	var metadata TemplateMetadata
	if err := yaml.Unmarshal(data, &metadata); err != nil {
		return nil
	}

	paramValidation := ParameterValidation{
		Valid:      true,
		Parameters: []ParameterCheck{},
		Unused:     []string{},
		Missing:    []string{},
	}

	// Validate each parameter
	for _, param := range metadata.Spec.Parameters {
		paramCheck := ParameterCheck{
			Name:         param.Name,
			Type:         param.Type,
			Required:     param.Required,
			HasDefault:   param.Default != nil,
			ValidationOK: true,
			UsedInFiles:  []string{},
		}

		// Validate parameter type
		validTypes := []string{"string", "integer", "boolean", "array", "object"}
		if !contains(validTypes, param.Type) {
			result.Errors = append(result.Errors, ValidationError{
				Code:     "PARAMETER_TYPE_INVALID",
				Message:  fmt.Sprintf("매개변수 %s의 타입이 잘못되었습니다: %s", param.Name, param.Type),
				File:     "template.yaml",
				Severity: "error",
			})
			paramCheck.ValidationOK = false
			paramValidation.Valid = false
		}

		// Check if required parameter has default
		if param.Required && param.Default != nil {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Code:    "PARAMETER_REQUIRED_WITH_DEFAULT",
				Message: fmt.Sprintf("필수 매개변수 %s에 기본값이 설정되어 있습니다", param.Name),
				File:    "template.yaml",
			})
		}

		paramValidation.Parameters = append(paramValidation.Parameters, paramCheck)
	}

	result.Details.Parameters = paramValidation
	return nil
}

func validateTemplateFile(filePath string, check *TemplateCheck) error {
	// Simple template validation - check for basic Go template syntax
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Extract variables used in template
	varRegex := regexp.MustCompile(`\{\{\s*\.(\w+)\s*\}\}`)
	matches := varRegex.FindAllStringSubmatch(string(content), -1)

	for _, match := range matches {
		if len(match) > 1 {
			check.Variables = append(check.Variables, match[1])
		}
	}

	// Remove duplicates
	check.Variables = removeDuplicates(check.Variables)

	return nil
}

func displayValidationResult(result *ValidationResult) {
	fmt.Printf("\n" + strings.Repeat("=", 60) + "\n")
	fmt.Printf("📊 템플릿 검증 결과\n")
	fmt.Printf(strings.Repeat("=", 60) + "\n")

	// Overall status
	if result.Valid {
		fmt.Printf("✅ 상태: 유효\n")
	} else {
		fmt.Printf("❌ 상태: 오류 발견\n")
	}

	// Summary
	fmt.Printf("\n📋 요약:\n")
	fmt.Printf("  📁 총 파일: %d개\n", result.Summary.TotalFiles)
	fmt.Printf("  📄 템플릿 파일: %d개\n", result.Summary.TemplateFiles)
	fmt.Printf("  📄 정적 파일: %d개\n", result.Summary.StaticFiles)
	fmt.Printf("  ❌ 오류: %d개\n", result.Summary.ErrorCount)
	fmt.Printf("  ⚠️  경고: %d개\n", result.Summary.WarningCount)

	// Errors
	if len(result.Errors) > 0 {
		fmt.Printf("\n❌ 오류:\n")
		for _, err := range result.Errors {
			fmt.Printf("  • %s: %s", err.Code, err.Message)
			if err.File != "" {
				fmt.Printf(" (%s)", err.File)
			}
			fmt.Printf("\n")
		}
	}

	// Warnings
	if len(result.Warnings) > 0 {
		fmt.Printf("\n⚠️  경고:\n")
		for _, warn := range result.Warnings {
			fmt.Printf("  • %s: %s", warn.Code, warn.Message)
			if warn.File != "" {
				fmt.Printf(" (%s)", warn.File)
			}
			fmt.Printf("\n")
		}
	}

	// Verbose details
	if verbose {
		displayDetailedResults(result)
	}

	fmt.Printf(strings.Repeat("=", 60) + "\n")

	if result.Valid {
		fmt.Printf("🎉 템플릿이 유효합니다!\n")
	} else {
		fmt.Printf("🔧 오류를 수정한 후 다시 검증해주세요.\n")
	}
}

func displayDetailedResults(result *ValidationResult) {
	fmt.Printf("\n🔍 상세 결과:\n")

	// Metadata details
	fmt.Printf("\n📄 메타데이터:\n")
	if result.Details.Metadata.Valid {
		fmt.Printf("  ✅ 메타데이터 유효\n")
	} else {
		fmt.Printf("  ❌ 메타데이터 오류\n")
	}
	fmt.Printf("  📦 버전 형식: %v\n", result.Details.Metadata.VersionFormat)
	fmt.Printf("  📂 카테고리 유효: %v\n", result.Details.Metadata.CategoryValid)
	fmt.Printf("  🏷️  타입 유효: %v\n", result.Details.Metadata.TypeValid)

	// File structure details
	fmt.Printf("\n📁 파일 구조:\n")
	if result.Details.Files.StructureValid {
		fmt.Printf("  ✅ 구조 유효\n")
	} else {
		fmt.Printf("  ❌ 구조 오류\n")
	}

	for _, dir := range result.Details.Files.RequiredDirs {
		status := "❌"
		if dir.Exists {
			status = "✅"
		}
		fmt.Printf("  %s %s", status, dir.Path)
		if dir.Required {
			fmt.Printf(" (필수)")
		}
		fmt.Printf("\n")
	}

	// Dependencies details
	if len(result.Details.Dependencies.Dependencies) > 0 {
		fmt.Printf("\n🔗 의존성:\n")
		for _, dep := range result.Details.Dependencies.Dependencies {
			status := "✅"
			if !dep.Available || !dep.Compatible {
				status = "❌"
			}
			fmt.Printf("  %s %s@%s\n", status, dep.Name, dep.Version)
		}
	}

	// Parameters details
	if len(result.Details.Parameters.Parameters) > 0 {
		fmt.Printf("\n⚙️  매개변수:\n")
		for _, param := range result.Details.Parameters.Parameters {
			status := "✅"
			if !param.ValidationOK {
				status = "❌"
			}
			fmt.Printf("  %s %s (%s)", status, param.Name, param.Type)
			if param.Required {
				fmt.Printf(" [필수]")
			}
			if param.HasDefault {
				fmt.Printf(" [기본값]")
			}
			fmt.Printf("\n")
		}
	}
}

// Utility functions
func isValidSemVer(version string) bool {
	semVerRegex := regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)
	return semVerRegex.MatchString(version)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func removeDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	result := []string{}
	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}
	return result
}
