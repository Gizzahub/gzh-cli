package reposync

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"
)

// JavaScriptDependencyParser implements dependency parsing for JavaScript
type JavaScriptDependencyParser struct {
	logger *zap.Logger
}

// TypeScriptDependencyParser implements dependency parsing for TypeScript
type TypeScriptDependencyParser struct {
	*JavaScriptDependencyParser
}

// PackageJSON represents the structure of package.json
type PackageJSON struct {
	Name             string            `json:"name"`
	Version          string            `json:"version"`
	Description      string            `json:"description"`
	Dependencies     map[string]string `json:"dependencies"`
	DevDependencies  map[string]string `json:"devDependencies"`
	PeerDependencies map[string]string `json:"peerDependencies"`
}

// NewJavaScriptDependencyParser creates a new JavaScript dependency parser
func NewJavaScriptDependencyParser(logger *zap.Logger) *JavaScriptDependencyParser {
	return &JavaScriptDependencyParser{logger: logger}
}

// NewTypeScriptDependencyParser creates a new TypeScript dependency parser
func NewTypeScriptDependencyParser(logger *zap.Logger) *TypeScriptDependencyParser {
	return &TypeScriptDependencyParser{
		JavaScriptDependencyParser: NewJavaScriptDependencyParser(logger),
	}
}

func (j *JavaScriptDependencyParser) Name() string     { return "javascript-parser" }
func (j *JavaScriptDependencyParser) Language() string { return "javascript" }

func (j *JavaScriptDependencyParser) FilePatterns() []string {
	return []string{"*.js", "*.jsx", "*.mjs"}
}

func (t *TypeScriptDependencyParser) Name() string     { return "typescript-parser" }
func (t *TypeScriptDependencyParser) Language() string { return "typescript" }

func (t *TypeScriptDependencyParser) FilePatterns() []string {
	return []string{"*.ts", "*.tsx"}
}

// ParseFile parses dependencies from a JavaScript/TypeScript source file
func (j *JavaScriptDependencyParser) ParseFile(filePath string) (*FileDependencies, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	fileDeps := &FileDependencies{
		FilePath:     filePath,
		Language:     j.Language(),
		Dependencies: make([]*Dependency, 0),
		Exports:      make([]string, 0),
		LastModified: fileInfo.ModTime(),
	}

	// Determine module path
	fileDeps.Module = j.extractModulePath(filePath)

	scanner := bufio.NewScanner(file)
	lineNum := 0

	// Regular expressions for different import/require patterns
	importRegexes := []*regexp.Regexp{
		// ES6 imports
		regexp.MustCompile(`^import\s+.*\s+from\s+['"]([^'"]+)['"]`),
		regexp.MustCompile(`^import\s+['"]([^'"]+)['"]`),

		// CommonJS requires
		regexp.MustCompile(`require\s*\(\s*['"]([^'"]+)['"]\s*\)`),

		// Dynamic imports
		regexp.MustCompile(`import\s*\(\s*['"]([^'"]+)['"]\s*\)`),

		// TypeScript imports (for TypeScript parser)
		regexp.MustCompile(`^import\s+.*\s+=\s+require\s*\(\s*['"]([^'"]+)['"]\s*\)`),
	}

	// Regular expressions for exports
	exportRegexes := []*regexp.Regexp{
		regexp.MustCompile(`^export\s+(?:default\s+)?(?:function|class|const|let|var)\s+(\w+)`),
		regexp.MustCompile(`^export\s+\{\s*([^}]+)\s*\}`),
		regexp.MustCompile(`^export\s+\*\s+from\s+['"]([^'"]+)['"]`),
	}

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if strings.HasPrefix(line, "//") || strings.HasPrefix(line, "/*") || line == "" {
			continue
		}

		// Parse imports/requires
		for _, regex := range importRegexes {
			matches := regex.FindStringSubmatch(line)
			if len(matches) > 1 {
				importPath := matches[1]

				dep := &Dependency{
					From:     fileDeps.Module,
					To:       importPath,
					Type:     j.getImportType(line),
					Language: j.Language(),
					Strength: j.getImportStrength(line),
					Location: SourceLocation{
						File:   filePath,
						Line:   lineNum,
						Column: strings.Index(line, importPath),
					},
					External: j.isExternalImport(importPath),
					Resolved: true,
				}

				fileDeps.Dependencies = append(fileDeps.Dependencies, dep)
				break
			}
		}

		// Parse exports
		for _, regex := range exportRegexes {
			matches := regex.FindStringSubmatch(line)
			if len(matches) > 1 {
				exports := j.parseExports(matches[1])
				fileDeps.Exports = append(fileDeps.Exports, exports...)
				break
			}
		}
	}

	// Count lines of code
	fileDeps.LinesOfCode = j.countLinesOfCode(filePath)

	return fileDeps, nil
}

// ParseModule parses dependencies for a JavaScript/TypeScript module
func (j *JavaScriptDependencyParser) ParseModule(modulePath string) (*ModuleDependencies, error) {
	moduleDeps := &ModuleDependencies{
		ModulePath:   modulePath,
		Language:     j.Language(),
		Files:        make([]string, 0),
		Dependencies: make([]*Dependency, 0),
		InternalDeps: make([]*Dependency, 0),
		ExternalDeps: make([]*Dependency, 0),
		Exports:      make([]string, 0),
	}

	// Read package.json if exists
	packageJSONPath := filepath.Join(modulePath, "package.json")
	if _, err := os.Stat(packageJSONPath); err == nil {
		if err := j.parsePackageJSON(packageJSONPath, moduleDeps); err != nil {
			j.logger.Warn("Failed to parse package.json", zap.Error(err))
		}
	}

	// Find all JavaScript/TypeScript files in the module
	patterns := j.FilePatterns()
	if j.Language() == "typescript" {
		patterns = append(patterns, "*.ts", "*.tsx")
	}

	err := filepath.Walk(modulePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		for _, pattern := range patterns {
			if matched, _ := filepath.Match(pattern, info.Name()); matched {
				// Skip node_modules and test files
				if !strings.Contains(path, "node_modules") && !strings.Contains(path, ".test.") && !strings.Contains(path, ".spec.") {
					moduleDeps.Files = append(moduleDeps.Files, path)
				}
				break
			}
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk module directory: %w", err)
	}

	return moduleDeps, nil
}

// parsePackageJSON parses package.json to extract module information
func (j *JavaScriptDependencyParser) parsePackageJSON(packageJSONPath string, moduleDeps *ModuleDependencies) error {
	data, err := os.ReadFile(packageJSONPath)
	if err != nil {
		return err
	}

	var pkg PackageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return err
	}

	moduleDeps.ModulePath = pkg.Name
	moduleDeps.Version = pkg.Version
	moduleDeps.Description = pkg.Description

	// Parse dependencies
	for name, version := range pkg.Dependencies {
		dep := &Dependency{
			From:     pkg.Name,
			To:       name,
			Type:     DependencyTypeRequire,
			Language: j.Language(),
			Strength: DependencyStrengthStrong,
			External: true,
			Resolved: true,
			Version:  version,
		}
		moduleDeps.ExternalDeps = append(moduleDeps.ExternalDeps, dep)
		moduleDeps.Dependencies = append(moduleDeps.Dependencies, dep)
	}

	// Parse dev dependencies
	for name, version := range pkg.DevDependencies {
		dep := &Dependency{
			From:     pkg.Name,
			To:       name,
			Type:     DependencyTypeRequire,
			Language: j.Language(),
			Strength: DependencyStrengthWeak,
			External: true,
			Resolved: true,
			Version:  version,
		}
		moduleDeps.ExternalDeps = append(moduleDeps.ExternalDeps, dep)
		moduleDeps.Dependencies = append(moduleDeps.Dependencies, dep)
	}

	// Parse peer dependencies
	for name, version := range pkg.PeerDependencies {
		dep := &Dependency{
			From:     pkg.Name,
			To:       name,
			Type:     DependencyTypeRequire,
			Language: j.Language(),
			Strength: DependencyStrengthOptional,
			External: true,
			Resolved: true,
			Version:  version,
		}
		moduleDeps.ExternalDeps = append(moduleDeps.ExternalDeps, dep)
		moduleDeps.Dependencies = append(moduleDeps.Dependencies, dep)
	}

	return nil
}

// extractModulePath extracts the module path for a JavaScript/TypeScript file
func (j *JavaScriptDependencyParser) extractModulePath(filePath string) string {
	// Look for package.json in parent directories
	dir := filepath.Dir(filePath)
	for {
		packageJSONPath := filepath.Join(dir, "package.json")
		if _, err := os.Stat(packageJSONPath); err == nil {
			// Found package.json, extract package name
			if packageName := j.getPackageNameFromJSON(packageJSONPath); packageName != "" {
				// Calculate relative path from package root
				relPath, err := filepath.Rel(dir, filepath.Dir(filePath))
				if err == nil && relPath != "." {
					return filepath.Join(packageName, relPath)
				}
				return packageName
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break // Reached root
		}
		dir = parent
	}

	// Fallback to directory-based module path
	return filepath.Dir(filePath)
}

// getPackageNameFromJSON extracts package name from package.json
func (j *JavaScriptDependencyParser) getPackageNameFromJSON(packageJSONPath string) string {
	data, err := os.ReadFile(packageJSONPath)
	if err != nil {
		return ""
	}

	var pkg PackageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return ""
	}

	return pkg.Name
}

// getImportType determines the type of import statement
func (j *JavaScriptDependencyParser) getImportType(line string) DependencyType {
	if strings.Contains(line, "require(") {
		return DependencyTypeRequire
	}
	if strings.Contains(line, "import(") {
		return DependencyTypeRequire // Dynamic import
	}
	return DependencyTypeImport
}

// getImportStrength determines the strength of the import
func (j *JavaScriptDependencyParser) getImportStrength(line string) DependencyStrength {
	// Dynamic imports are typically weaker
	if strings.Contains(line, "import(") {
		return DependencyStrengthWeak
	}

	// Type-only imports (TypeScript)
	if strings.Contains(line, "import type") {
		return DependencyStrengthWeak
	}

	return DependencyStrengthStrong
}

// isExternalImport determines if an import is external
func (j *JavaScriptDependencyParser) isExternalImport(importPath string) bool {
	// Relative imports (starting with . or /) are internal
	if strings.HasPrefix(importPath, ".") || strings.HasPrefix(importPath, "/") {
		return false
	}

	// Node.js built-in modules
	builtinModules := []string{
		"assert", "buffer", "child_process", "cluster", "crypto", "dgram",
		"dns", "domain", "events", "fs", "http", "https", "net", "os",
		"path", "punycode", "querystring", "readline", "repl", "stream",
		"string_decoder", "tls", "tty", "url", "util", "vm", "zlib",
	}

	for _, builtin := range builtinModules {
		if importPath == builtin {
			return false // Built-in module, not external
		}
	}

	// If it doesn't start with a relative path, it's likely external
	return true
}

// parseExports parses export statements to extract exported names
func (j *JavaScriptDependencyParser) parseExports(exportStr string) []string {
	exports := make([]string, 0)

	// Handle export lists: { foo, bar as baz }
	if strings.Contains(exportStr, ",") {
		parts := strings.Split(exportStr, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			// Handle "as" aliases
			if strings.Contains(part, " as ") {
				aliasParts := strings.Split(part, " as ")
				if len(aliasParts) > 1 {
					exports = append(exports, strings.TrimSpace(aliasParts[1]))
				}
			} else {
				exports = append(exports, part)
			}
		}
	} else {
		exports = append(exports, strings.TrimSpace(exportStr))
	}

	return exports
}

// countLinesOfCode counts non-empty, non-comment lines
func (j *JavaScriptDependencyParser) countLinesOfCode(filePath string) int {
	file, err := os.Open(filePath)
	if err != nil {
		return 0
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	count := 0
	inBlockComment := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Handle block comments
		if strings.Contains(line, "/*") && !strings.Contains(line, "*/") {
			inBlockComment = true
			continue
		}
		if strings.Contains(line, "*/") {
			inBlockComment = false
			continue
		}
		if inBlockComment {
			continue
		}

		// Skip empty lines and single-line comments
		if line != "" && !strings.HasPrefix(line, "//") {
			count++
		}
	}

	return count
}
