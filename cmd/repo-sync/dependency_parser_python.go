package reposync

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"
)

// PythonDependencyParser implements dependency parsing for Python
type PythonDependencyParser struct {
	logger *zap.Logger
}

// NewPythonDependencyParser creates a new Python dependency parser
func NewPythonDependencyParser(logger *zap.Logger) *PythonDependencyParser {
	return &PythonDependencyParser{logger: logger}
}

func (p *PythonDependencyParser) Name() string     { return "python-parser" }
func (p *PythonDependencyParser) Language() string { return "python" }

func (p *PythonDependencyParser) FilePatterns() []string {
	return []string{"*.py"}
}

// ParseFile parses dependencies from a Python source file
func (p *PythonDependencyParser) ParseFile(filePath string) (*FileDependencies, error) {
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
		Language:     "python",
		Dependencies: make([]*Dependency, 0),
		Exports:      make([]string, 0),
		LastModified: fileInfo.ModTime(),
	}

	// Determine module path
	fileDeps.Module = p.extractModulePath(filePath)

	scanner := bufio.NewScanner(file)
	lineNum := 0

	// Regular expressions for different import patterns
	importRegexes := []*regexp.Regexp{
		// import module
		regexp.MustCompile(`^import\s+([\w.]+)(?:\s+as\s+\w+)?`),

		// from module import ...
		regexp.MustCompile(`^from\s+([\w.]+)\s+import\s+`),

		// from . import ... (relative imports)
		regexp.MustCompile(`^from\s+(\.[\w.]*)\s+import\s+`),

		// from .. import ... (relative imports)
		regexp.MustCompile(`^from\s+(\.\.[\w.]*)\s+import\s+`),
	}

	// Regular expressions for exports (functions, classes, variables)
	exportRegexes := []*regexp.Regexp{
		regexp.MustCompile(`^def\s+(\w+)\s*\(`),
		regexp.MustCompile(`^class\s+(\w+)\s*[\(:]`),
		regexp.MustCompile(`^(\w+)\s*=`),
	}

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip comments, empty lines, and docstrings
		if strings.HasPrefix(line, "#") || line == "" ||
			strings.HasPrefix(line, "\"\"\"") || strings.HasPrefix(line, "'''") {
			continue
		}

		// Parse imports
		for _, regex := range importRegexes {
			matches := regex.FindStringSubmatch(line)
			if len(matches) > 1 {
				importPath := matches[1]

				dep := &Dependency{
					From:     fileDeps.Module,
					To:       importPath,
					Type:     DependencyTypeImport,
					Language: "python",
					Strength: p.getImportStrength(line),
					Location: SourceLocation{
						File:   filePath,
						Line:   lineNum,
						Column: strings.Index(line, importPath),
					},
					External: p.isExternalImport(importPath, filePath),
					Resolved: true,
				}

				fileDeps.Dependencies = append(fileDeps.Dependencies, dep)
				break
			}
		}

		// Parse exports (functions, classes, variables at module level)
		for _, regex := range exportRegexes {
			matches := regex.FindStringSubmatch(line)
			if len(matches) > 1 {
				name := matches[1]
				// Only include names that don't start with underscore (public API)
				if !strings.HasPrefix(name, "_") {
					fileDeps.Exports = append(fileDeps.Exports, name)
				}
				break
			}
		}
	}

	// Count lines of code
	fileDeps.LinesOfCode = p.countLinesOfCode(filePath)

	return fileDeps, nil
}

// ParseModule parses dependencies for a Python module/package
func (p *PythonDependencyParser) ParseModule(modulePath string) (*ModuleDependencies, error) {
	moduleDeps := &ModuleDependencies{
		ModulePath:   modulePath,
		Language:     "python",
		Files:        make([]string, 0),
		Dependencies: make([]*Dependency, 0),
		InternalDeps: make([]*Dependency, 0),
		ExternalDeps: make([]*Dependency, 0),
		Exports:      make([]string, 0),
	}

	// Read setup.py, requirements.txt, or pyproject.toml if exists
	p.parsePackageFiles(modulePath, moduleDeps)

	// Find all Python files in the module
	err := filepath.Walk(modulePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(path, ".py") {
			// Skip test files and __pycache__
			if !strings.Contains(path, "__pycache__") &&
				!strings.Contains(path, "test_") &&
				!strings.HasSuffix(path, "_test.py") {
				moduleDeps.Files = append(moduleDeps.Files, path)
			}
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk module directory: %w", err)
	}

	return moduleDeps, nil
}

// parsePackageFiles parses Python package configuration files
func (p *PythonDependencyParser) parsePackageFiles(modulePath string, moduleDeps *ModuleDependencies) {
	// Try to parse setup.py
	setupPyPath := filepath.Join(modulePath, "setup.py")
	if _, err := os.Stat(setupPyPath); err == nil {
		p.parseSetupPy(setupPyPath, moduleDeps)
	}

	// Try to parse requirements.txt
	requirementsPath := filepath.Join(modulePath, "requirements.txt")
	if _, err := os.Stat(requirementsPath); err == nil {
		p.parseRequirementsTxt(requirementsPath, moduleDeps)
	}

	// Try to parse pyproject.toml
	pyprojectPath := filepath.Join(modulePath, "pyproject.toml")
	if _, err := os.Stat(pyprojectPath); err == nil {
		p.parsePyprojectToml(pyprojectPath, moduleDeps)
	}
}

// parseSetupPy parses setup.py to extract package information
func (p *PythonDependencyParser) parseSetupPy(setupPyPath string, moduleDeps *ModuleDependencies) {
	file, err := os.Open(setupPyPath)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	inInstallRequires := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Extract package name
		if strings.Contains(line, "name=") {
			name := p.extractQuotedValue(line, "name=")
			if name != "" {
				moduleDeps.ModulePath = name
			}
		}

		// Extract version
		if strings.Contains(line, "version=") {
			version := p.extractQuotedValue(line, "version=")
			if version != "" {
				moduleDeps.Version = version
			}
		}

		// Extract description
		if strings.Contains(line, "description=") {
			description := p.extractQuotedValue(line, "description=")
			if description != "" {
				moduleDeps.Description = description
			}
		}

		// Handle install_requires
		if strings.Contains(line, "install_requires=") {
			inInstallRequires = true
			continue
		}

		if inInstallRequires {
			if strings.Contains(line, "]") {
				inInstallRequires = false
				continue
			}

			// Parse requirement line
			if requirement := p.parseRequirementLine(line); requirement != "" {
				dep := &Dependency{
					From:     moduleDeps.ModulePath,
					To:       requirement,
					Type:     DependencyTypeRequire,
					Language: "python",
					Strength: DependencyStrengthStrong,
					External: true,
					Resolved: true,
				}
				moduleDeps.ExternalDeps = append(moduleDeps.ExternalDeps, dep)
				moduleDeps.Dependencies = append(moduleDeps.Dependencies, dep)
			}
		}
	}
}

// parseRequirementsTxt parses requirements.txt file
func (p *PythonDependencyParser) parseRequirementsTxt(requirementsPath string, moduleDeps *ModuleDependencies) {
	file, err := os.Open(requirementsPath)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		// Parse requirement line
		if requirement := p.parseRequirementLine(line); requirement != "" {
			dep := &Dependency{
				From:     moduleDeps.ModulePath,
				To:       requirement,
				Type:     DependencyTypeRequire,
				Language: "python",
				Strength: DependencyStrengthStrong,
				External: true,
				Resolved: true,
			}
			moduleDeps.ExternalDeps = append(moduleDeps.ExternalDeps, dep)
			moduleDeps.Dependencies = append(moduleDeps.Dependencies, dep)
		}
	}
}

// parsePyprojectToml parses pyproject.toml file (basic parsing)
func (p *PythonDependencyParser) parsePyprojectToml(pyprojectPath string, moduleDeps *ModuleDependencies) {
	file, err := os.Open(pyprojectPath)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	inDependencies := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Extract package name
		if strings.Contains(line, "name =") {
			name := p.extractQuotedValue(line, "name =")
			if name != "" {
				moduleDeps.ModulePath = name
			}
		}

		// Handle dependencies section
		if strings.Contains(line, "dependencies = [") {
			inDependencies = true
			continue
		}

		if inDependencies {
			if strings.Contains(line, "]") {
				inDependencies = false
				continue
			}

			// Parse dependency line
			if requirement := p.parseRequirementLine(line); requirement != "" {
				dep := &Dependency{
					From:     moduleDeps.ModulePath,
					To:       requirement,
					Type:     DependencyTypeRequire,
					Language: "python",
					Strength: DependencyStrengthStrong,
					External: true,
					Resolved: true,
				}
				moduleDeps.ExternalDeps = append(moduleDeps.ExternalDeps, dep)
				moduleDeps.Dependencies = append(moduleDeps.Dependencies, dep)
			}
		}
	}
}

// extractModulePath extracts the module path for a Python file
func (p *PythonDependencyParser) extractModulePath(filePath string) string {
	// Look for setup.py, __init__.py, or pyproject.toml in parent directories
	dir := filepath.Dir(filePath)

	for {
		// Check for package indicators
		setupPyPath := filepath.Join(dir, "setup.py")
		pyprojectPath := filepath.Join(dir, "pyproject.toml")
		initPyPath := filepath.Join(dir, "__init__.py")

		if _, err := os.Stat(setupPyPath); err == nil {
			// Found setup.py, this is likely the package root
			packageName := p.getPackageNameFromSetup(setupPyPath)
			if packageName != "" {
				relPath, err := filepath.Rel(dir, filepath.Dir(filePath))
				if err == nil && relPath != "." {
					return strings.ReplaceAll(filepath.Join(packageName, relPath), string(filepath.Separator), ".")
				}
				return packageName
			}
		}

		if _, err := os.Stat(pyprojectPath); err == nil {
			// Found pyproject.toml
			packageName := p.getPackageNameFromPyproject(pyprojectPath)
			if packageName != "" {
				relPath, err := filepath.Rel(dir, filepath.Dir(filePath))
				if err == nil && relPath != "." {
					return strings.ReplaceAll(filepath.Join(packageName, relPath), string(filepath.Separator), ".")
				}
				return packageName
			}
		}

		if _, err := os.Stat(initPyPath); err == nil {
			// Found __init__.py, this directory is a Python package
			packageName := filepath.Base(dir)
			relPath, err := filepath.Rel(dir, filepath.Dir(filePath))
			if err == nil && relPath != "." {
				return strings.ReplaceAll(filepath.Join(packageName, relPath), string(filepath.Separator), ".")
			}
			return packageName
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break // Reached root
		}
		dir = parent
	}

	// Fallback: convert file path to module path
	relPath, err := filepath.Rel(filepath.Dir(filePath), filePath)
	if err != nil {
		relPath = filepath.Base(filePath)
	}

	// Remove .py extension and convert path separators to dots
	modulePath := strings.TrimSuffix(relPath, ".py")
	return strings.ReplaceAll(modulePath, string(filepath.Separator), ".")
}

// getPackageNameFromSetup extracts package name from setup.py
func (p *PythonDependencyParser) getPackageNameFromSetup(setupPyPath string) string {
	file, err := os.Open(setupPyPath)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.Contains(line, "name=") {
			return p.extractQuotedValue(line, "name=")
		}
	}

	return ""
}

// getPackageNameFromPyproject extracts package name from pyproject.toml
func (p *PythonDependencyParser) getPackageNameFromPyproject(pyprojectPath string) string {
	file, err := os.Open(pyprojectPath)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.Contains(line, "name =") {
			return p.extractQuotedValue(line, "name =")
		}
	}

	return ""
}

// getImportStrength determines the strength of the import
func (p *PythonDependencyParser) getImportStrength(line string) DependencyStrength {
	// Try/except imports are optional
	if strings.Contains(line, "try:") || strings.Contains(line, "except:") {
		return DependencyStrengthOptional
	}

	return DependencyStrengthStrong
}

// isExternalImport determines if an import is external
func (p *PythonDependencyParser) isExternalImport(importPath, filePath string) bool {
	// Relative imports (starting with .) are internal
	if strings.HasPrefix(importPath, ".") {
		return false
	}

	// Standard library modules
	stdLibModules := []string{
		"os", "sys", "re", "json", "time", "datetime", "math", "random",
		"collections", "itertools", "functools", "operator", "pathlib",
		"urllib", "http", "socket", "threading", "multiprocessing",
		"subprocess", "argparse", "logging", "unittest", "sqlite3",
		"csv", "xml", "email", "base64", "hashlib", "pickle", "copy",
		"typing", "enum", "dataclasses", "abc", "contextlib", "warnings",
	}

	mainModule := strings.Split(importPath, ".")[0]
	for _, stdLib := range stdLibModules {
		if mainModule == stdLib {
			return false
		}
	}

	// Check if it's a local module (exists in the project)
	projectRoot := p.findProjectRoot(filePath)
	if projectRoot != "" {
		modulePath := strings.ReplaceAll(importPath, ".", string(filepath.Separator))
		localPath := filepath.Join(projectRoot, modulePath+".py")
		localPackagePath := filepath.Join(projectRoot, modulePath, "__init__.py")

		if _, err := os.Stat(localPath); err == nil {
			return false
		}
		if _, err := os.Stat(localPackagePath); err == nil {
			return false
		}
	}

	return true
}

// findProjectRoot finds the root directory of the Python project
func (p *PythonDependencyParser) findProjectRoot(filePath string) string {
	dir := filepath.Dir(filePath)

	for {
		// Look for project indicators
		if _, err := os.Stat(filepath.Join(dir, "setup.py")); err == nil {
			return dir
		}
		if _, err := os.Stat(filepath.Join(dir, "pyproject.toml")); err == nil {
			return dir
		}
		if _, err := os.Stat(filepath.Join(dir, "requirements.txt")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return ""
}

// extractQuotedValue extracts a quoted value from a line
func (p *PythonDependencyParser) extractQuotedValue(line, prefix string) string {
	start := strings.Index(line, prefix)
	if start == -1 {
		return ""
	}

	valueStart := start + len(prefix)
	remainder := line[valueStart:]

	// Find the quoted value
	if match := regexp.MustCompile(`["']([^"']+)["']`).FindStringSubmatch(remainder); len(match) > 1 {
		return match[1]
	}

	return ""
}

// parseRequirementLine parses a requirement line and extracts the package name
func (p *PythonDependencyParser) parseRequirementLine(line string) string {
	// Remove quotes and whitespace
	line = strings.Trim(strings.TrimSpace(line), `"'`)

	// Skip empty lines and comments
	if line == "" || strings.HasPrefix(line, "#") {
		return ""
	}

	// Split on version specifiers
	versionRegex := regexp.MustCompile(`([a-zA-Z0-9_-]+)([>=<!~].*)`)
	matches := versionRegex.FindStringSubmatch(line)
	if len(matches) > 1 {
		return matches[1]
	}

	// If no version specifier, return the whole line as package name
	if regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(line) {
		return line
	}

	return ""
}

// countLinesOfCode counts non-empty, non-comment lines
func (p *PythonDependencyParser) countLinesOfCode(filePath string) int {
	file, err := os.Open(filePath)
	if err != nil {
		return 0
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	count := 0
	inDocstring := false
	docstringDelim := ""

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Handle docstrings
		if strings.HasPrefix(line, `"""`) || strings.HasPrefix(line, "'''") {
			if !inDocstring {
				inDocstring = true
				docstringDelim = line[:3]
				// Check if docstring ends on the same line
				if strings.Count(line, docstringDelim) >= 2 {
					inDocstring = false
				}
				continue
			} else if strings.HasSuffix(line, docstringDelim) {
				inDocstring = false
				continue
			}
		}

		if inDocstring {
			continue
		}

		// Skip empty lines and comments
		if line != "" && !strings.HasPrefix(line, "#") {
			count++
		}
	}

	return count
}
