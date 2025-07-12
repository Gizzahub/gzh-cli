package reposync

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"
)

// DependencyAnalyzer analyzes dependencies between repositories and modules
type DependencyAnalyzer struct {
	logger  *zap.Logger
	config  *DependencyConfig
	parsers map[string]DependencyParser
}

// DependencyConfig represents configuration for dependency analysis
type DependencyConfig struct {
	RepositoryPath   string            `json:"repository_path"`
	IncludeLanguages []string          `json:"include_languages"`
	ExcludePatterns  []string          `json:"exclude_patterns"`
	CustomParsers    map[string]string `json:"custom_parsers"`
	OutputFormat     string            `json:"output_format"`
	AnalysisDepth    int               `json:"analysis_depth"`
	IncludeExternal  bool              `json:"include_external"`
	IncludeInternal  bool              `json:"include_internal"`
}

// DependencyParser interface for language-specific dependency parsing
type DependencyParser interface {
	Name() string
	Language() string
	FilePatterns() []string
	ParseFile(filePath string) (*FileDependencies, error)
	ParseModule(modulePath string) (*ModuleDependencies, error)
}

// DependencyResult represents the complete dependency analysis result
type DependencyResult struct {
	Repository     string                         `json:"repository"`
	AnalysisTime   time.Time                      `json:"analysis_time"`
	TotalFiles     int                            `json:"total_files"`
	TotalModules   int                            `json:"total_modules"`
	Dependencies   []*Dependency                  `json:"dependencies"`
	Modules        map[string]*ModuleDependencies `json:"modules"`
	Files          map[string]*FileDependencies   `json:"files"`
	Statistics     DependencyStatistics           `json:"statistics"`
	CyclicDeps     []CyclicDependency             `json:"cyclic_dependencies"`
	ExternalDeps   []ExternalDependency           `json:"external_dependencies"`
	UnresolvedDeps []UnresolvedDependency         `json:"unresolved_dependencies"`
}

// Dependency represents a dependency relationship
type Dependency struct {
	From     string             `json:"from"`     // Source module/file
	To       string             `json:"to"`       // Target module/file
	Type     DependencyType     `json:"type"`     // import, require, include, etc.
	Language string             `json:"language"` // go, python, javascript, etc.
	Strength DependencyStrength `json:"strength"` // strong, weak, optional
	Location SourceLocation     `json:"location"` // Where the dependency is declared
	Resolved bool               `json:"resolved"` // Whether the dependency can be resolved
	External bool               `json:"external"` // Whether it's an external dependency
	Version  string             `json:"version,omitempty"`
}

// DependencyType represents the type of dependency
type DependencyType string

const (
	DependencyTypeImport  DependencyType = "import"
	DependencyTypeRequire DependencyType = "require"
	DependencyTypeInclude DependencyType = "include"
	DependencyTypeInherit DependencyType = "inherit"
	DependencyTypeCompose DependencyType = "compose"
	DependencyTypeCall    DependencyType = "call"
)

// DependencyStrength represents the strength of a dependency
type DependencyStrength string

const (
	DependencyStrengthStrong   DependencyStrength = "strong"
	DependencyStrengthWeak     DependencyStrength = "weak"
	DependencyStrengthOptional DependencyStrength = "optional"
)

// SourceLocation represents the location where a dependency is declared
type SourceLocation struct {
	File   string `json:"file"`
	Line   int    `json:"line"`
	Column int    `json:"column"`
}

// FileDependencies represents dependencies for a specific file
type FileDependencies struct {
	FilePath     string        `json:"file_path"`
	Language     string        `json:"language"`
	Module       string        `json:"module"`
	Dependencies []*Dependency `json:"dependencies"`
	Exports      []string      `json:"exports"`
	LinesOfCode  int           `json:"lines_of_code"`
	LastModified time.Time     `json:"last_modified"`
}

// ModuleDependencies represents dependencies for a module
type ModuleDependencies struct {
	ModulePath   string        `json:"module_path"`
	Language     string        `json:"language"`
	Files        []string      `json:"files"`
	Dependencies []*Dependency `json:"dependencies"`
	InternalDeps []*Dependency `json:"internal_dependencies"`
	ExternalDeps []*Dependency `json:"external_dependencies"`
	Dependents   []string      `json:"dependents"`
	Exports      []string      `json:"exports"`
	Version      string        `json:"version,omitempty"`
	Description  string        `json:"description,omitempty"`
}

// DependencyStatistics provides statistical information about dependencies
type DependencyStatistics struct {
	TotalDependencies    int                        `json:"total_dependencies"`
	InternalDependencies int                        `json:"internal_dependencies"`
	ExternalDependencies int                        `json:"external_dependencies"`
	CircularDependencies int                        `json:"circular_dependencies"`
	LanguageBreakdown    map[string]int             `json:"language_breakdown"`
	TypeBreakdown        map[DependencyType]int     `json:"type_breakdown"`
	StrengthBreakdown    map[DependencyStrength]int `json:"strength_breakdown"`
	DependencyDepth      map[string]int             `json:"dependency_depth"`
	ModuleComplexity     map[string]float64         `json:"module_complexity"`
	UnresolvedCount      int                        `json:"unresolved_count"`
}

// CyclicDependency represents a circular dependency
type CyclicDependency struct {
	Cycle       []string `json:"cycle"`
	Length      int      `json:"length"`
	Severity    string   `json:"severity"`
	Description string   `json:"description"`
	Suggestions []string `json:"suggestions"`
}

// ExternalDependency represents an external dependency
type ExternalDependency struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Language    string `json:"language"`
	UsageCount  int    `json:"usage_count"`
	License     string `json:"license,omitempty"`
	Repository  string `json:"repository,omitempty"`
	Description string `json:"description,omitempty"`
}

// UnresolvedDependency represents a dependency that cannot be resolved
type UnresolvedDependency struct {
	From        string         `json:"from"`
	Target      string         `json:"target"`
	Language    string         `json:"language"`
	Location    SourceLocation `json:"location"`
	Reason      string         `json:"reason"`
	Suggestions []string       `json:"suggestions"`
}

// NewDependencyAnalyzer creates a new dependency analyzer
func NewDependencyAnalyzer(logger *zap.Logger, config *DependencyConfig) *DependencyAnalyzer {
	analyzer := &DependencyAnalyzer{
		logger:  logger,
		config:  config,
		parsers: make(map[string]DependencyParser),
	}

	// Initialize default parsers
	analyzer.initializeParsers()

	return analyzer
}

// initializeParsers initializes language-specific dependency parsers
func (da *DependencyAnalyzer) initializeParsers() {
	// Go parser
	da.parsers["go"] = NewGoDependencyParser(da.logger)

	// JavaScript/TypeScript parser
	da.parsers["javascript"] = NewJavaScriptDependencyParser(da.logger)
	da.parsers["typescript"] = NewTypeScriptDependencyParser(da.logger)

	// Python parser
	da.parsers["python"] = NewPythonDependencyParser(da.logger)
}

// AddParser adds a custom dependency parser
func (da *DependencyAnalyzer) AddParser(parser DependencyParser) {
	da.parsers[parser.Language()] = parser
}

// AnalyzeDependencies performs comprehensive dependency analysis
func (da *DependencyAnalyzer) AnalyzeDependencies(ctx context.Context) (*DependencyResult, error) {
	startTime := time.Now()

	result := &DependencyResult{
		Repository:     da.config.RepositoryPath,
		AnalysisTime:   startTime,
		Dependencies:   make([]*Dependency, 0),
		Modules:        make(map[string]*ModuleDependencies),
		Files:          make(map[string]*FileDependencies),
		CyclicDeps:     make([]CyclicDependency, 0),
		ExternalDeps:   make([]ExternalDependency, 0),
		UnresolvedDeps: make([]UnresolvedDependency, 0),
	}

	da.logger.Info("Starting dependency analysis",
		zap.String("repository", da.config.RepositoryPath))

	fmt.Printf("🔍 Analyzing dependencies in: %s\n", da.config.RepositoryPath)

	// Discover all files to analyze
	files, err := da.discoverFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to discover files: %w", err)
	}

	result.TotalFiles = len(files)
	fmt.Printf("📁 Found %d files to analyze\n", len(files))

	// Parse file dependencies
	for _, file := range files {
		if err := da.parseFileDependencies(file, result); err != nil {
			da.logger.Warn("Failed to parse file dependencies",
				zap.String("file", file),
				zap.Error(err))
		}
	}

	// Analyze module-level dependencies
	if err := da.analyzeModuleDependencies(result); err != nil {
		return nil, fmt.Errorf("failed to analyze module dependencies: %w", err)
	}

	// Detect circular dependencies
	cyclicDeps := da.detectCircularDependencies(result)
	result.CyclicDeps = cyclicDeps

	// Analyze external dependencies
	externalDeps := da.analyzeExternalDependencies(result)
	result.ExternalDeps = externalDeps

	// Calculate statistics
	result.Statistics = da.calculateStatistics(result)

	fmt.Printf("✅ Analysis completed in %v\n", time.Since(startTime))
	fmt.Printf("📊 Found %d dependencies (%d internal, %d external)\n",
		result.Statistics.TotalDependencies,
		result.Statistics.InternalDependencies,
		result.Statistics.ExternalDependencies)

	if len(result.CyclicDeps) > 0 {
		fmt.Printf("⚠️  Found %d circular dependencies\n", len(result.CyclicDeps))
	}

	return result, nil
}

// discoverFiles discovers all files to analyze based on configuration
func (da *DependencyAnalyzer) discoverFiles() ([]string, error) {
	var files []string

	err := filepath.Walk(da.config.RepositoryPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			// Skip excluded directories
			if da.shouldSkipDir(info.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if file should be included
		if da.shouldIncludeFile(path) {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

// shouldSkipDir determines if a directory should be skipped
func (da *DependencyAnalyzer) shouldSkipDir(dirname string) bool {
	skipDirs := []string{
		".git", ".svn", ".hg",
		"node_modules", "vendor", "third_party",
		".vscode", ".idea",
		"target", "build", "dist", "out",
		"__pycache__", ".pytest_cache",
		".coverage", "coverage",
	}

	for _, skip := range skipDirs {
		if dirname == skip {
			return true
		}
	}

	// Check custom exclude patterns
	for _, pattern := range da.config.ExcludePatterns {
		if matched, _ := filepath.Match(pattern, dirname); matched {
			return true
		}
	}

	return false
}

// shouldIncludeFile determines if a file should be included in analysis
func (da *DependencyAnalyzer) shouldIncludeFile(filePath string) bool {
	ext := filepath.Ext(filePath)

	// Check if language is supported
	lang := da.getLanguageFromExtension(ext)
	if lang == "" {
		return false
	}

	// Check if language is included in configuration
	if len(da.config.IncludeLanguages) > 0 {
		included := false
		for _, includeLang := range da.config.IncludeLanguages {
			if lang == includeLang {
				included = true
				break
			}
		}
		if !included {
			return false
		}
	}

	// Check exclude patterns
	for _, pattern := range da.config.ExcludePatterns {
		if matched, _ := filepath.Match(pattern, filepath.Base(filePath)); matched {
			return false
		}
	}

	return true
}

// getLanguageFromExtension returns the language for a file extension
func (da *DependencyAnalyzer) getLanguageFromExtension(ext string) string {
	langMap := map[string]string{
		".go":    "go",
		".js":    "javascript",
		".jsx":   "javascript",
		".ts":    "typescript",
		".tsx":   "typescript",
		".py":    "python",
		".java":  "java",
		".kt":    "kotlin",
		".rs":    "rust",
		".cpp":   "cpp",
		".cc":    "cpp",
		".cxx":   "cpp",
		".c":     "c",
		".h":     "c",
		".cs":    "csharp",
		".php":   "php",
		".rb":    "ruby",
		".swift": "swift",
	}

	return langMap[ext]
}

// parseFileDependencies parses dependencies for a specific file
func (da *DependencyAnalyzer) parseFileDependencies(filePath string, result *DependencyResult) error {
	ext := filepath.Ext(filePath)
	lang := da.getLanguageFromExtension(ext)

	parser, exists := da.parsers[lang]
	if !exists {
		return fmt.Errorf("no parser available for language: %s", lang)
	}

	fileDeps, err := parser.ParseFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to parse file %s: %w", filePath, err)
	}

	result.Files[filePath] = fileDeps
	result.Dependencies = append(result.Dependencies, fileDeps.Dependencies...)

	return nil
}

// analyzeModuleDependencies analyzes dependencies at the module level
func (da *DependencyAnalyzer) analyzeModuleDependencies(result *DependencyResult) error {
	// Group files by module
	moduleFiles := make(map[string][]string)

	for filePath, fileDeps := range result.Files {
		module := fileDeps.Module
		if module == "" {
			module = da.inferModuleFromPath(filePath)
		}
		moduleFiles[module] = append(moduleFiles[module], filePath)
	}

	// Analyze each module
	for modulePath, files := range moduleFiles {
		moduleDeps := &ModuleDependencies{
			ModulePath:   modulePath,
			Files:        files,
			Dependencies: make([]*Dependency, 0),
			InternalDeps: make([]*Dependency, 0),
			ExternalDeps: make([]*Dependency, 0),
			Dependents:   make([]string, 0),
			Exports:      make([]string, 0),
		}

		// Collect module dependencies from file dependencies
		for _, file := range files {
			if fileDeps, exists := result.Files[file]; exists {
				moduleDeps.Dependencies = append(moduleDeps.Dependencies, fileDeps.Dependencies...)
				moduleDeps.Exports = append(moduleDeps.Exports, fileDeps.Exports...)

				if moduleDeps.Language == "" {
					moduleDeps.Language = fileDeps.Language
				}
			}
		}

		// Classify dependencies as internal or external
		for _, dep := range moduleDeps.Dependencies {
			if dep.External {
				moduleDeps.ExternalDeps = append(moduleDeps.ExternalDeps, dep)
			} else {
				moduleDeps.InternalDeps = append(moduleDeps.InternalDeps, dep)
			}
		}

		result.Modules[modulePath] = moduleDeps
	}

	result.TotalModules = len(result.Modules)
	return nil
}

// inferModuleFromPath infers module name from file path
func (da *DependencyAnalyzer) inferModuleFromPath(filePath string) string {
	relPath, err := filepath.Rel(da.config.RepositoryPath, filePath)
	if err != nil {
		return filepath.Dir(filePath)
	}

	// For Go, look for go.mod
	dir := filepath.Dir(relPath)
	for dir != "." && dir != "/" {
		goModPath := filepath.Join(da.config.RepositoryPath, dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return dir
		}
		dir = filepath.Dir(dir)
	}

	// For JavaScript/TypeScript, look for package.json
	dir = filepath.Dir(relPath)
	for dir != "." && dir != "/" {
		packagePath := filepath.Join(da.config.RepositoryPath, dir, "package.json")
		if _, err := os.Stat(packagePath); err == nil {
			return dir
		}
		dir = filepath.Dir(dir)
	}

	// For Python, look for __init__.py or setup.py
	dir = filepath.Dir(relPath)
	for dir != "." && dir != "/" {
		initPath := filepath.Join(da.config.RepositoryPath, dir, "__init__.py")
		setupPath := filepath.Join(da.config.RepositoryPath, dir, "setup.py")
		if _, err := os.Stat(initPath); err == nil {
			return dir
		}
		if _, err := os.Stat(setupPath); err == nil {
			return dir
		}
		dir = filepath.Dir(dir)
	}

	// Default to directory containing the file
	return filepath.Dir(relPath)
}

// detectCircularDependencies detects circular dependencies using DFS
func (da *DependencyAnalyzer) detectCircularDependencies(result *DependencyResult) []CyclicDependency {
	cyclicDeps := make([]CyclicDependency, 0)

	// Build dependency graph
	graph := make(map[string][]string)
	for _, dep := range result.Dependencies {
		if !dep.External {
			graph[dep.From] = append(graph[dep.From], dep.To)
		}
	}

	// Find cycles using DFS
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var dfs func(string, []string) []string
	dfs = func(node string, path []string) []string {
		if recStack[node] {
			// Found a cycle
			cycleStart := -1
			for i, p := range path {
				if p == node {
					cycleStart = i
					break
				}
			}
			if cycleStart >= 0 {
				return append(path[cycleStart:], node)
			}
		}

		if visited[node] {
			return nil
		}

		visited[node] = true
		recStack[node] = true

		for _, neighbor := range graph[node] {
			if cycle := dfs(neighbor, append(path, node)); cycle != nil {
				return cycle
			}
		}

		recStack[node] = false
		return nil
	}

	// Check each unvisited node
	for node := range graph {
		if !visited[node] {
			if cycle := dfs(node, []string{}); cycle != nil {
				cyclicDep := CyclicDependency{
					Cycle:       cycle,
					Length:      len(cycle) - 1,
					Severity:    da.calculateCycleSeverity(cycle),
					Description: da.describeCycle(cycle),
					Suggestions: da.suggestCycleFixes(cycle),
				}
				cyclicDeps = append(cyclicDeps, cyclicDep)
			}
		}
	}

	return cyclicDeps
}

// calculateCycleSeverity calculates the severity of a circular dependency
func (da *DependencyAnalyzer) calculateCycleSeverity(cycle []string) string {
	switch {
	case len(cycle) <= 3:
		return "high"
	case len(cycle) <= 5:
		return "medium"
	default:
		return "low"
	}
}

// describeCycle provides a human-readable description of the cycle
func (da *DependencyAnalyzer) describeCycle(cycle []string) string {
	if len(cycle) < 2 {
		return "Invalid cycle"
	}

	return fmt.Sprintf("Circular dependency: %s forms a cycle of length %d",
		strings.Join(cycle, " → "), len(cycle)-1)
}

// suggestCycleFixes suggests ways to fix circular dependencies
func (da *DependencyAnalyzer) suggestCycleFixes(cycle []string) []string {
	suggestions := []string{
		"Extract common functionality into a separate module",
		"Use dependency injection to break direct dependencies",
		"Reorganize code to follow a layered architecture",
		"Consider using interfaces to reduce coupling",
	}

	if len(cycle) == 3 {
		suggestions = append(suggestions, "Consider merging the two modules if they are tightly coupled")
	}

	return suggestions
}

// analyzeExternalDependencies analyzes external dependencies
func (da *DependencyAnalyzer) analyzeExternalDependencies(result *DependencyResult) []ExternalDependency {
	externalMap := make(map[string]*ExternalDependency)

	for _, dep := range result.Dependencies {
		if dep.External {
			key := fmt.Sprintf("%s-%s", dep.To, dep.Language)
			if existing, exists := externalMap[key]; exists {
				existing.UsageCount++
			} else {
				externalMap[key] = &ExternalDependency{
					Name:       dep.To,
					Version:    dep.Version,
					Language:   dep.Language,
					UsageCount: 1,
				}
			}
		}
	}

	externalDeps := make([]ExternalDependency, 0, len(externalMap))
	for _, dep := range externalMap {
		externalDeps = append(externalDeps, *dep)
	}

	// Sort by usage count
	sort.Slice(externalDeps, func(i, j int) bool {
		return externalDeps[i].UsageCount > externalDeps[j].UsageCount
	})

	return externalDeps
}

// calculateStatistics calculates dependency statistics
func (da *DependencyAnalyzer) calculateStatistics(result *DependencyResult) DependencyStatistics {
	stats := DependencyStatistics{
		LanguageBreakdown: make(map[string]int),
		TypeBreakdown:     make(map[DependencyType]int),
		StrengthBreakdown: make(map[DependencyStrength]int),
		DependencyDepth:   make(map[string]int),
		ModuleComplexity:  make(map[string]float64),
	}

	// Count dependencies
	stats.TotalDependencies = len(result.Dependencies)
	stats.CircularDependencies = len(result.CyclicDeps)
	stats.UnresolvedCount = len(result.UnresolvedDeps)

	for _, dep := range result.Dependencies {
		// Language breakdown
		stats.LanguageBreakdown[dep.Language]++

		// Type breakdown
		stats.TypeBreakdown[dep.Type]++

		// Strength breakdown
		stats.StrengthBreakdown[dep.Strength]++

		// Internal vs external
		if dep.External {
			stats.ExternalDependencies++
		} else {
			stats.InternalDependencies++
		}
	}

	// Calculate module complexity (based on number of dependencies)
	for modulePath, module := range result.Modules {
		complexity := float64(len(module.Dependencies)) / 10.0 // Normalize
		if complexity > 10.0 {
			complexity = 10.0
		}
		stats.ModuleComplexity[modulePath] = complexity
	}

	return stats
}

// SaveResult saves the dependency analysis result to file
func (da *DependencyAnalyzer) SaveResult(result *DependencyResult, outputPath string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write result file: %w", err)
	}

	da.logger.Info("Dependency analysis result saved",
		zap.String("output_path", outputPath))

	return nil
}
