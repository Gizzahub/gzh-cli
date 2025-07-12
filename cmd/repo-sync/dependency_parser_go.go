package reposync

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"
)

// GoDependencyParser implements dependency parsing for Go language
type GoDependencyParser struct {
	logger *zap.Logger
}

// NewGoDependencyParser creates a new Go dependency parser
func NewGoDependencyParser(logger *zap.Logger) *GoDependencyParser {
	return &GoDependencyParser{logger: logger}
}

func (g *GoDependencyParser) Name() string     { return "go-parser" }
func (g *GoDependencyParser) Language() string { return "go" }

func (g *GoDependencyParser) FilePatterns() []string {
	return []string{"*.go"}
}

// ParseFile parses dependencies from a Go source file
func (g *GoDependencyParser) ParseFile(filePath string) (*FileDependencies, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// Parse the Go source file
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Go file: %w", err)
	}

	fileDeps := &FileDependencies{
		FilePath:     filePath,
		Language:     "go",
		Dependencies: make([]*Dependency, 0),
		Exports:      make([]string, 0),
		LastModified: fileInfo.ModTime(),
	}

	// Count lines of code
	fileDeps.LinesOfCode = g.countLinesOfCode(filePath)

	// Extract package name and determine module
	fileDeps.Module = g.extractModulePath(filePath, node.Name.Name)

	// Parse imports
	for _, imp := range node.Imports {
		importPath := strings.Trim(imp.Path.Value, "\"")

		dep := &Dependency{
			From:     fileDeps.Module,
			To:       importPath,
			Type:     DependencyTypeImport,
			Language: "go",
			Strength: DependencyStrengthStrong,
			Location: SourceLocation{
				File:   filePath,
				Line:   fset.Position(imp.Pos()).Line,
				Column: fset.Position(imp.Pos()).Column,
			},
			External: g.isExternalImport(importPath),
			Resolved: true,
		}

		// Check if it's a dot import (special case)
		if imp.Name != nil && imp.Name.Name == "." {
			dep.Strength = DependencyStrengthWeak
		}

		fileDeps.Dependencies = append(fileDeps.Dependencies, dep)
	}

	// Extract exported symbols
	for _, decl := range node.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if d.Name.IsExported() {
				fileDeps.Exports = append(fileDeps.Exports, d.Name.Name)
			}
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					if s.Name.IsExported() {
						fileDeps.Exports = append(fileDeps.Exports, s.Name.Name)
					}
				case *ast.ValueSpec:
					for _, name := range s.Names {
						if name.IsExported() {
							fileDeps.Exports = append(fileDeps.Exports, name.Name)
						}
					}
				}
			}
		}
	}

	return fileDeps, nil
}

// ParseModule parses dependencies for a Go module
func (g *GoDependencyParser) ParseModule(modulePath string) (*ModuleDependencies, error) {
	moduleDeps := &ModuleDependencies{
		ModulePath:   modulePath,
		Language:     "go",
		Files:        make([]string, 0),
		Dependencies: make([]*Dependency, 0),
		InternalDeps: make([]*Dependency, 0),
		ExternalDeps: make([]*Dependency, 0),
		Exports:      make([]string, 0),
	}

	// Read go.mod file if exists
	goModPath := filepath.Join(modulePath, "go.mod")
	if _, err := os.Stat(goModPath); err == nil {
		if err := g.parseGoMod(goModPath, moduleDeps); err != nil {
			g.logger.Warn("Failed to parse go.mod", zap.Error(err))
		}
	}

	// Find all Go files in the module
	err := filepath.Walk(modulePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			moduleDeps.Files = append(moduleDeps.Files, path)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk module directory: %w", err)
	}

	return moduleDeps, nil
}

// parseGoMod parses go.mod file to extract module information
func (g *GoDependencyParser) parseGoMod(goModPath string, moduleDeps *ModuleDependencies) error {
	file, err := os.Open(goModPath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	inRequireBlock := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if strings.HasPrefix(line, "//") || line == "" {
			continue
		}

		// Parse module name
		if strings.HasPrefix(line, "module ") {
			moduleName := strings.TrimSpace(strings.TrimPrefix(line, "module"))
			moduleDeps.ModulePath = moduleName
		}

		// Parse go version
		if strings.HasPrefix(line, "go ") {
			version := strings.TrimSpace(strings.TrimPrefix(line, "go"))
			moduleDeps.Version = version
		}

		// Handle require block
		if strings.HasPrefix(line, "require (") {
			inRequireBlock = true
			continue
		}

		if inRequireBlock && line == ")" {
			inRequireBlock = false
			continue
		}

		// Parse require statements
		if strings.HasPrefix(line, "require ") || inRequireBlock {
			g.parseRequireLine(line, moduleDeps)
		}
	}

	return scanner.Err()
}

// parseRequireLine parses a require line from go.mod
func (g *GoDependencyParser) parseRequireLine(line string, moduleDeps *ModuleDependencies) {
	// Remove "require " prefix
	line = strings.TrimSpace(strings.TrimPrefix(line, "require"))

	// Skip empty lines
	if line == "" {
		return
	}

	// Parse dependency line: module version [// indirect]
	parts := strings.Fields(line)
	if len(parts) >= 2 {
		module := parts[0]
		version := parts[1]

		dep := &Dependency{
			From:     moduleDeps.ModulePath,
			To:       module,
			Type:     DependencyTypeRequire,
			Language: "go",
			Strength: DependencyStrengthStrong,
			External: true,
			Resolved: true,
			Version:  version,
		}

		// Check if it's an indirect dependency
		if len(parts) > 2 && strings.Contains(strings.Join(parts[2:], " "), "indirect") {
			dep.Strength = DependencyStrengthWeak
		}

		moduleDeps.ExternalDeps = append(moduleDeps.ExternalDeps, dep)
		moduleDeps.Dependencies = append(moduleDeps.Dependencies, dep)
	}
}

// extractModulePath extracts the module path for a Go file
func (g *GoDependencyParser) extractModulePath(filePath, packageName string) string {
	// Look for go.mod in parent directories
	dir := filepath.Dir(filePath)
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			// Found go.mod, extract module name
			if moduleName := g.getModuleNameFromGoMod(goModPath); moduleName != "" {
				// Calculate relative path from module root
				relPath, err := filepath.Rel(dir, filepath.Dir(filePath))
				if err == nil && relPath != "." {
					return filepath.Join(moduleName, relPath)
				}
				return moduleName
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

// getModuleNameFromGoMod extracts module name from go.mod file
func (g *GoDependencyParser) getModuleNameFromGoMod(goModPath string) string {
	file, err := os.Open(goModPath)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module"))
		}
	}

	return ""
}

// isExternalImport determines if an import is external (not part of current module)
func (g *GoDependencyParser) isExternalImport(importPath string) bool {
	// Standard library packages (no dots in path and common patterns)
	stdLibPatterns := []string{
		"^(bufio|bytes|context|crypto|database|encoding|errors|fmt|hash|html|image|io|log|math|net|os|path|reflect|regexp|runtime|sort|strconv|strings|sync|syscall|testing|text|time|unicode|unsafe)(/.*)?$",
		"^compress/.*",
		"^container/.*",
		"^debug/.*",
		"^go/.*",
		"^index/.*",
		"^mime/.*",
	}

	for _, pattern := range stdLibPatterns {
		if matched, _ := regexp.MatchString(pattern, importPath); matched {
			return false // Standard library, not external
		}
	}

	// If it contains a dot, it's likely external
	return strings.Contains(importPath, ".")
}

// countLinesOfCode counts non-empty, non-comment lines
func (g *GoDependencyParser) countLinesOfCode(filePath string) int {
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
		if strings.Contains(line, "/*") {
			inBlockComment = true
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
