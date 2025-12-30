// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

// Package godoc provides API documentation analysis capabilities
package godoc

import (
	"context"
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/gizzahub/gzh-cli/internal/logger"
)

// Analyzer provides GoDoc coverage and quality analysis.
type Analyzer struct {
	logger     *logger.SimpleLogger
	fileSet    *token.FileSet
	workingDir string
}

// PackageInfo contains documentation analysis information for a package.
type PackageInfo struct {
	ImportPath       string         `json:"importPath"`
	Name             string         `json:"name"`
	Dir              string         `json:"dir"`
	GoFiles          []string       `json:"goFiles"`
	PublicFunctions  []FunctionInfo `json:"publicFunctions"`
	PublicTypes      []TypeInfo     `json:"publicTypes"`
	PublicVariables  []VariableInfo `json:"publicVariables"`
	PublicConstants  []ConstantInfo `json:"publicConstants"`
	CoverageStats    CoverageStats  `json:"coverageStats"`
	QualityIssues    []QualityIssue `json:"qualityIssues"`
	ExampleFunctions []ExampleInfo  `json:"example_functions"`
	PackageDoc       string         `json:"package_doc"`
	HasPackageDoc    bool           `json:"has_package_doc"`
	Recommendations  []string       `json:"recommendations"`
}

// FunctionInfo contains information about a documented function.
type FunctionInfo struct {
	Name       string   `json:"name"`
	Signature  string   `json:"signature"`
	Doc        string   `json:"doc"`
	HasDoc     bool     `json:"has_doc"`
	IsExported bool     `json:"is_exported"`
	Line       int      `json:"line"`
	Examples   []string `json:"examples"`
	Complexity int      `json:"complexity"`
}

// TypeInfo contains information about a documented type.
type TypeInfo struct {
	Name       string         `json:"name"`
	Kind       string         `json:"kind"`
	Doc        string         `json:"doc"`
	HasDoc     bool           `json:"has_doc"`
	IsExported bool           `json:"is_exported"`
	Line       int            `json:"line"`
	Methods    []FunctionInfo `json:"methods"`
	Fields     []FieldInfo    `json:"fields,omitempty"`
}

// FieldInfo contains information about struct fields.
type FieldInfo struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Doc        string `json:"doc"`
	HasDoc     bool   `json:"has_doc"`
	IsExported bool   `json:"is_exported"`
	Tag        string `json:"tag,omitempty"`
}

// VariableInfo contains information about package variables.
type VariableInfo struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Doc        string `json:"doc"`
	HasDoc     bool   `json:"has_doc"`
	IsExported bool   `json:"is_exported"`
	Line       int    `json:"line"`
}

// ConstantInfo contains information about package constants.
type ConstantInfo struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Value      string `json:"value"`
	Doc        string `json:"doc"`
	HasDoc     bool   `json:"has_doc"`
	IsExported bool   `json:"is_exported"`
	Line       int    `json:"line"`
}

// ExampleInfo contains information about example functions.
type ExampleInfo struct {
	Name        string `json:"name"`
	ForFunction string `json:"for_function"`
	Code        string `json:"code"`
	Output      string `json:"output"`
	HasOutput   bool   `json:"has_output"`
}

// CoverageStats contains documentation coverage statistics.
type CoverageStats struct {
	TotalPublicSymbols  int     `json:"total_public_symbols"`
	DocumentedSymbols   int     `json:"documented_symbols"`
	UndocumentedSymbols int     `json:"undocumented_symbols"`
	CoveragePercentage  float64 `json:"coverage_percentage"`
	FunctionCoverage    float64 `json:"function_coverage"`
	TypeCoverage        float64 `json:"type_coverage"`
	VariableCoverage    float64 `json:"variable_coverage"`
	ConstantCoverage    float64 `json:"constant_coverage"`
	PackageDocumented   bool    `json:"package_documented"`
	ExampleCount        int     `json:"example_count"`
	ExamplesPerFunction float64 `json:"examples_per_function"`
}

// QualityIssue represents a documentation quality issue.
type QualityIssue struct {
	Type       string `json:"type"`
	Severity   string `json:"severity"`
	Message    string `json:"message"`
	Symbol     string `json:"symbol"`
	File       string `json:"file"`
	Line       int    `json:"line"`
	Suggestion string `json:"suggestion"`
}

// NewAnalyzer creates a new GoDoc analyzer.
func NewAnalyzer(workingDir string) *Analyzer {
	if workingDir == "" {
		workingDir = "."
	}

	return &Analyzer{
		logger:     logger.NewSimpleLogger("godoc-analyzer"),
		fileSet:    token.NewFileSet(),
		workingDir: workingDir,
	}
}

// AnalyzePackage analyzes a Go package for documentation coverage and quality.
func (a *Analyzer) AnalyzePackage(_ context.Context, packagePath string) (*PackageInfo, error) {
	a.logger.Debug("Analyzing package", "path", packagePath)

	// Parse package directory
	var pkgPath string
	if filepath.IsAbs(packagePath) {
		pkgPath = packagePath
	} else {
		pkgPath = filepath.Join(a.workingDir, packagePath)
		var err error
		pkgPath, err = filepath.Abs(pkgPath)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve absolute path: %w", err)
		}
	}

	// Check if directory exists
	if _, err := os.Stat(pkgPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("package directory does not exist: %s", pkgPath)
	}

	// Parse Go files in the package
	pkgs, err := parser.ParseDir(a.fileSet, pkgPath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse package: %w", err)
	}

	// Find the main package (ignore test packages)
	var pkg *ast.Package
	for name, p := range pkgs {
		if !strings.HasSuffix(name, "_test") {
			pkg = p
			break
		}
	}

	if pkg == nil {
		return nil, fmt.Errorf("no non-test packages found in directory: %s", pkgPath)
	}

	// Create doc package
	docPkg := doc.New(pkg, packagePath, doc.AllDecls)

	// Build package info
	pkgInfo := &PackageInfo{
		ImportPath:      packagePath,
		Name:            pkg.Name,
		Dir:             pkgPath,
		GoFiles:         a.getGoFiles(pkgPath),
		PackageDoc:      docPkg.Doc,
		HasPackageDoc:   strings.TrimSpace(docPkg.Doc) != "",
		Recommendations: make([]string, 0),
	}

	// Analyze functions
	pkgInfo.PublicFunctions = a.analyzeFunctions(docPkg.Funcs)

	// Analyze types
	pkgInfo.PublicTypes = a.analyzeTypes(docPkg.Types)

	// Analyze variables
	pkgInfo.PublicVariables = a.analyzeVariables(docPkg.Vars)

	// Analyze constants
	pkgInfo.PublicConstants = a.analyzeConstants(docPkg.Consts)

	// Find examples
	pkgInfo.ExampleFunctions = a.analyzeExamples(pkg)

	// Calculate coverage statistics
	pkgInfo.CoverageStats = a.calculateCoverageStats(pkgInfo)

	// Identify quality issues
	pkgInfo.QualityIssues = a.identifyQualityIssues(pkgInfo)

	// Generate recommendations
	pkgInfo.Recommendations = a.generateRecommendations(pkgInfo)

	a.logger.Info("Package analysis completed",
		"package", packagePath,
		"coverage", fmt.Sprintf("%.1f%%", pkgInfo.CoverageStats.CoveragePercentage),
		"quality_issues", len(pkgInfo.QualityIssues),
	)

	return pkgInfo, nil
}

// getGoFiles returns a list of Go files in the directory.
func (a *Analyzer) getGoFiles(dir string) []string {
	files := make([]string, 0)

	entries, err := os.ReadDir(dir)
	if err != nil {
		a.logger.Warn("Failed to read directory", "dir", dir, "error", err)
		return files
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go") {
			files = append(files, name)
		}
	}

	return files
}

// analyzeFunctions analyzes function documentation.
func (a *Analyzer) analyzeFunctions(funcs []*doc.Func) []FunctionInfo {
	functions := make([]FunctionInfo, 0, len(funcs))

	for _, fn := range funcs {
		funcInfo := FunctionInfo{
			Name:       fn.Name,
			Doc:        fn.Doc,
			HasDoc:     strings.TrimSpace(fn.Doc) != "",
			IsExported: ast.IsExported(fn.Name),
			Examples:   make([]string, 0),
		}

		// Get position information
		if fn.Decl != nil {
			pos := a.fileSet.Position(fn.Decl.Pos())
			funcInfo.Line = pos.Line
		}

		// Build signature
		if fn.Decl != nil {
			funcInfo.Signature = a.buildFunctionSignature(fn.Decl)
			funcInfo.Complexity = a.calculateComplexity(fn.Decl)
		}

		functions = append(functions, funcInfo)
	}

	return functions
}

// analyzeTypes analyzes type documentation.
func (a *Analyzer) analyzeTypes(types []*doc.Type) []TypeInfo {
	typeInfos := make([]TypeInfo, 0, len(types))

	for _, typ := range types {
		typeInfo := TypeInfo{
			Name:       typ.Name,
			Doc:        typ.Doc,
			HasDoc:     strings.TrimSpace(typ.Doc) != "",
			IsExported: ast.IsExported(typ.Name),
			Methods:    make([]FunctionInfo, 0),
			Fields:     make([]FieldInfo, 0),
		}

		// Get position information
		if typ.Decl != nil {
			pos := a.fileSet.Position(typ.Decl.Pos())
			typeInfo.Line = pos.Line
		}

		// Determine type kind
		if typ.Decl != nil && len(typ.Decl.Specs) > 0 {
			if typeSpec, ok := typ.Decl.Specs[0].(*ast.TypeSpec); ok {
				typeInfo.Kind = a.getTypeKind(typeSpec.Type)

				// Analyze struct fields if it's a struct
				if structType, ok := typeSpec.Type.(*ast.StructType); ok {
					typeInfo.Fields = a.analyzeStructFields(structType)
				}
			}
		}

		// Analyze methods
		for _, method := range typ.Funcs {
			methodInfo := FunctionInfo{
				Name:       method.Name,
				Doc:        method.Doc,
				HasDoc:     strings.TrimSpace(method.Doc) != "",
				IsExported: ast.IsExported(method.Name),
				Examples:   make([]string, 0),
			}

			if method.Decl != nil {
				pos := a.fileSet.Position(method.Decl.Pos())
				methodInfo.Line = pos.Line
				methodInfo.Signature = a.buildFunctionSignature(method.Decl)
				methodInfo.Complexity = a.calculateComplexity(method.Decl)
			}

			typeInfo.Methods = append(typeInfo.Methods, methodInfo)
		}

		typeInfos = append(typeInfos, typeInfo)
	}

	return typeInfos
}

// analyzeVariables analyzes variable documentation.
func (a *Analyzer) analyzeVariables(vars []*doc.Value) []VariableInfo {
	variables := make([]VariableInfo, 0)

	for _, v := range vars {
		if v.Decl == nil {
			continue
		}

		pos := a.fileSet.Position(v.Decl.Pos())

		for _, name := range v.Names {
			varInfo := VariableInfo{
				Name:       name,
				Doc:        v.Doc,
				HasDoc:     strings.TrimSpace(v.Doc) != "",
				IsExported: ast.IsExported(name),
				Line:       pos.Line,
			}

			variables = append(variables, varInfo)
		}
	}

	return variables
}

// analyzeConstants analyzes constant documentation.
func (a *Analyzer) analyzeConstants(consts []*doc.Value) []ConstantInfo {
	constants := make([]ConstantInfo, 0)

	for _, c := range consts {
		if c.Decl == nil {
			continue
		}

		pos := a.fileSet.Position(c.Decl.Pos())

		for _, name := range c.Names {
			constInfo := ConstantInfo{
				Name:       name,
				Doc:        c.Doc,
				HasDoc:     strings.TrimSpace(c.Doc) != "",
				IsExported: ast.IsExported(name),
				Line:       pos.Line,
			}

			constants = append(constants, constInfo)
		}
	}

	return constants
}

// analyzeExamples finds and analyzes example functions.
func (a *Analyzer) analyzeExamples(pkg *ast.Package) []ExampleInfo {
	examples := make([]ExampleInfo, 0)

	for _, file := range pkg.Files {
		for _, decl := range file.Decls {
			if funcDecl, ok := decl.(*ast.FuncDecl); ok {
				if strings.HasPrefix(funcDecl.Name.Name, "Example") {
					exampleInfo := ExampleInfo{
						Name: funcDecl.Name.Name,
					}

					// Extract the function this example is for
					if len(funcDecl.Name.Name) > 7 { // len("Example")
						exampleInfo.ForFunction = funcDecl.Name.Name[7:]
					}

					examples = append(examples, exampleInfo)
				}
			}
		}
	}

	return examples
}

// calculateCoverageStats calculates documentation coverage statistics.
func (a *Analyzer) calculateCoverageStats(pkgInfo *PackageInfo) CoverageStats {
	stats := CoverageStats{
		PackageDocumented: pkgInfo.HasPackageDoc,
		ExampleCount:      len(pkgInfo.ExampleFunctions),
	}

	// Count documented symbols by type
	functionCounts := a.countDocumentedFunctions(pkgInfo, &stats)
	typeCounts := a.countDocumentedTypes(pkgInfo, &stats, &functionCounts)
	varCounts := a.countDocumentedVariables(pkgInfo, &stats)
	constCounts := a.countDocumentedConstants(pkgInfo, &stats)

	// Calculate coverage percentages
	a.calculateCoveragePercentages(&stats, functionCounts, typeCounts, varCounts, constCounts)

	return stats
}

type symbolCounts struct {
	documented int
	total      int
}

// countDocumentedFunctions counts documented functions and methods.
func (a *Analyzer) countDocumentedFunctions(pkgInfo *PackageInfo, stats *CoverageStats) symbolCounts {
	counts := symbolCounts{}

	for _, fn := range pkgInfo.PublicFunctions {
		if fn.IsExported {
			stats.TotalPublicSymbols++
			if fn.HasDoc {
				counts.documented++
				stats.DocumentedSymbols++
			} else {
				stats.UndocumentedSymbols++
			}
		}
	}

	return counts
}

// countDocumentedTypes counts documented types and their methods.
func (a *Analyzer) countDocumentedTypes(pkgInfo *PackageInfo, stats *CoverageStats, functionCounts *symbolCounts) symbolCounts {
	counts := symbolCounts{}

	for _, typ := range pkgInfo.PublicTypes {
		if typ.IsExported {
			counts.total++
			stats.TotalPublicSymbols++
			if typ.HasDoc {
				counts.documented++
				stats.DocumentedSymbols++
			} else {
				stats.UndocumentedSymbols++
			}

			// Count methods
			for _, method := range typ.Methods {
				if method.IsExported {
					stats.TotalPublicSymbols++
					if method.HasDoc {
						functionCounts.documented++
						stats.DocumentedSymbols++
					} else {
						stats.UndocumentedSymbols++
					}
				}
			}
		}
	}

	return counts
}

// countDocumentedVariables counts documented variables.
func (a *Analyzer) countDocumentedVariables(pkgInfo *PackageInfo, stats *CoverageStats) symbolCounts {
	counts := symbolCounts{}

	for _, v := range pkgInfo.PublicVariables {
		if v.IsExported {
			counts.total++
			stats.TotalPublicSymbols++
			if v.HasDoc {
				counts.documented++
				stats.DocumentedSymbols++
			} else {
				stats.UndocumentedSymbols++
			}
		}
	}

	return counts
}

// countDocumentedConstants counts documented constants.
func (a *Analyzer) countDocumentedConstants(pkgInfo *PackageInfo, stats *CoverageStats) symbolCounts {
	counts := symbolCounts{}

	for _, c := range pkgInfo.PublicConstants {
		if c.IsExported {
			counts.total++
			stats.TotalPublicSymbols++
			if c.HasDoc {
				counts.documented++
				stats.DocumentedSymbols++
			} else {
				stats.UndocumentedSymbols++
			}
		}
	}

	return counts
}

// calculateCoveragePercentages calculates coverage percentages for all symbol types.
func (a *Analyzer) calculateCoveragePercentages(stats *CoverageStats, functionCounts, typeCounts, varCounts, constCounts symbolCounts) {
	// Overall coverage
	if stats.TotalPublicSymbols > 0 {
		stats.CoveragePercentage = float64(stats.DocumentedSymbols) * 100.0 / float64(stats.TotalPublicSymbols)
	}

	// Function coverage
	totalFunctions := functionCounts.documented + (stats.UndocumentedSymbols - typeCounts.total - varCounts.total - constCounts.total)
	if totalFunctions > 0 {
		stats.FunctionCoverage = float64(functionCounts.documented) * 100.0 / float64(totalFunctions)
	}

	// Type coverage
	if typeCounts.total > 0 {
		stats.TypeCoverage = float64(typeCounts.documented) * 100.0 / float64(typeCounts.total)
	}

	// Variable coverage
	if varCounts.total > 0 {
		stats.VariableCoverage = float64(varCounts.documented) * 100.0 / float64(varCounts.total)
	}

	// Constant coverage
	if constCounts.total > 0 {
		stats.ConstantCoverage = float64(constCounts.documented) * 100.0 / float64(constCounts.total)
	}

	// Examples per function
	if totalFunctions > 0 {
		stats.ExamplesPerFunction = float64(stats.ExampleCount) / float64(totalFunctions)
	}
}

// identifyQualityIssues identifies documentation quality issues.
func (a *Analyzer) identifyQualityIssues(pkgInfo *PackageInfo) []QualityIssue {
	issues := make([]QualityIssue, 0)

	// Check for missing package documentation
	if !pkgInfo.HasPackageDoc {
		issues = append(issues, QualityIssue{
			Type:       "missing_package_doc",
			Severity:   "high",
			Message:    "Package lacks documentation comment",
			Symbol:     pkgInfo.Name,
			Suggestion: fmt.Sprintf("Add a package comment starting with 'Package %s'", pkgInfo.Name),
		})
	}

	// Check undocumented exported functions
	for _, fn := range pkgInfo.PublicFunctions {
		if fn.IsExported && !fn.HasDoc {
			issues = append(issues, QualityIssue{
				Type:       "missing_function_doc",
				Severity:   "medium",
				Message:    fmt.Sprintf("Exported function '%s' lacks documentation", fn.Name),
				Symbol:     fn.Name,
				Line:       fn.Line,
				Suggestion: fmt.Sprintf("Add a comment starting with '%s'", fn.Name),
			})
		}
	}

	// Check undocumented exported types
	for _, typ := range pkgInfo.PublicTypes {
		if typ.IsExported && !typ.HasDoc {
			issues = append(issues, QualityIssue{
				Type:       "missing_type_doc",
				Severity:   "medium",
				Message:    fmt.Sprintf("Exported type '%s' lacks documentation", typ.Name),
				Symbol:     typ.Name,
				Line:       typ.Line,
				Suggestion: fmt.Sprintf("Add a comment starting with '%s'", typ.Name),
			})
		}

		// Check undocumented exported methods
		for _, method := range typ.Methods {
			if method.IsExported && !method.HasDoc {
				issues = append(issues, QualityIssue{
					Type:       "missing_method_doc",
					Severity:   "medium",
					Message:    fmt.Sprintf("Exported method '%s.%s' lacks documentation", typ.Name, method.Name),
					Symbol:     fmt.Sprintf("%s.%s", typ.Name, method.Name),
					Line:       method.Line,
					Suggestion: fmt.Sprintf("Add a comment starting with '%s'", method.Name),
				})
			}
		}
	}

	// Check for complex functions without adequate documentation
	for _, fn := range pkgInfo.PublicFunctions {
		if fn.IsExported && fn.Complexity > 10 && len(strings.Split(fn.Doc, "\n")) < 3 {
			issues = append(issues, QualityIssue{
				Type:       "inadequate_complex_function_doc",
				Severity:   "low",
				Message:    fmt.Sprintf("Complex function '%s' needs more detailed documentation", fn.Name),
				Symbol:     fn.Name,
				Line:       fn.Line,
				Suggestion: "Add parameter descriptions, return value explanations, and usage examples",
			})
		}
	}

	return issues
}

// generateRecommendations generates improvement recommendations.
func (a *Analyzer) generateRecommendations(pkgInfo *PackageInfo) []string {
	recommendations := make([]string, 0)

	// Coverage-based recommendations
	if pkgInfo.CoverageStats.CoveragePercentage < 80 {
		recommendations = append(recommendations,
			fmt.Sprintf("Improve documentation coverage from %.1f%% to at least 80%%", pkgInfo.CoverageStats.CoveragePercentage))
	}

	// Example-based recommendations
	exportedFunctions := 0
	for _, fn := range pkgInfo.PublicFunctions {
		if fn.IsExported {
			exportedFunctions++
		}
	}

	if exportedFunctions > 0 && pkgInfo.CoverageStats.ExamplesPerFunction < 0.5 {
		recommendations = append(recommendations,
			fmt.Sprintf("Add more example functions (current: %d, recommended: %d)",
				pkgInfo.CoverageStats.ExampleCount, exportedFunctions/2))
	}

	// Quality-based recommendations
	highSeverityIssues := 0
	for _, issue := range pkgInfo.QualityIssues {
		if issue.Severity == "high" {
			highSeverityIssues++
		}
	}

	if highSeverityIssues > 0 {
		recommendations = append(recommendations,
			fmt.Sprintf("Address %d high-severity documentation issues", highSeverityIssues))
	}

	// Type-specific recommendations
	if pkgInfo.CoverageStats.TypeCoverage < pkgInfo.CoverageStats.FunctionCoverage {
		recommendations = append(recommendations,
			"Focus on documenting exported types and their methods")
	}

	return recommendations
}

// Helper functions

func (a *Analyzer) buildFunctionSignature(decl *ast.FuncDecl) string {
	// This is a simplified signature builder
	// In a complete implementation, you'd want to fully reconstruct the signature
	if decl.Name != nil {
		return fmt.Sprintf("func %s(...)", decl.Name.Name)
	}
	return "func(...)"
}

func (a *Analyzer) calculateComplexity(decl *ast.FuncDecl) int {
	// Simplified cyclomatic complexity calculation
	// Count branching statements
	complexity := 1

	ast.Inspect(decl, func(n ast.Node) bool {
		switch n.(type) {
		case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt, *ast.SwitchStmt, *ast.TypeSwitchStmt:
			complexity++
		case *ast.CaseClause:
			complexity++
		}
		return true
	})

	return complexity
}

func (a *Analyzer) getTypeKind(expr ast.Expr) string {
	switch expr.(type) {
	case *ast.StructType:
		return "struct"
	case *ast.InterfaceType:
		return "interface"
	case *ast.ArrayType:
		return "array"
	case *ast.MapType:
		return "map"
	case *ast.ChanType:
		return "chan"
	case *ast.FuncType:
		return "func"
	default:
		return "type"
	}
}

func (a *Analyzer) analyzeStructFields(structType *ast.StructType) []FieldInfo {
	fields := make([]FieldInfo, 0)

	if structType.Fields != nil {
		for _, field := range structType.Fields.List {
			for _, name := range field.Names {
				fieldInfo := FieldInfo{
					Name:       name.Name,
					IsExported: ast.IsExported(name.Name),
				}

				if field.Doc != nil {
					fieldInfo.Doc = field.Doc.Text()
					fieldInfo.HasDoc = strings.TrimSpace(fieldInfo.Doc) != ""
				}

				if field.Tag != nil {
					fieldInfo.Tag = field.Tag.Value
				}

				fields = append(fields, fieldInfo)
			}
		}
	}

	return fields
}
