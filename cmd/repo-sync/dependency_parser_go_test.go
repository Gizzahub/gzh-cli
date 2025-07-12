package reposync

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestGoDependencyParser_Basic(t *testing.T) {
	logger := zaptest.NewLogger(t)
	parser := NewGoDependencyParser(logger)

	assert.Equal(t, "go-parser", parser.Name())
	assert.Equal(t, "go", parser.Language())
	assert.Equal(t, []string{"*.go"}, parser.FilePatterns())
}

func TestGoDependencyParser_ParseFile(t *testing.T) {
	// Create a temporary Go file
	tempDir, err := os.MkdirTemp("", "go_parser_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	goFile := filepath.Join(tempDir, "test.go")
	goContent := `package main

import (
	"fmt"
	"os"
	"github.com/example/pkg"
	. "github.com/dot/import"
)

// ExportedFunction is an exported function
func ExportedFunction() {
	fmt.Println("Hello")
}

func privateFunction() {
	// private function
}

// ExportedType is an exported type
type ExportedType struct {
	Field string
}

// ExportedVar is an exported variable
var ExportedVar = "test"

var privateVar = "private"
`

	err = os.WriteFile(goFile, []byte(goContent), 0o644)
	require.NoError(t, err)

	logger := zaptest.NewLogger(t)
	parser := NewGoDependencyParser(logger)

	result, err := parser.ParseFile(goFile)
	require.NoError(t, err)

	assert.Equal(t, goFile, result.FilePath)
	assert.Equal(t, "go", result.Language)
	assert.True(t, result.LinesOfCode > 0)

	// Check dependencies
	assert.Len(t, result.Dependencies, 4)

	// Check that standard library imports are not external
	fmtDep := findDependency(result.Dependencies, "fmt")
	assert.NotNil(t, fmtDep)
	assert.False(t, fmtDep.External)

	// Check that third-party imports are external
	exampleDep := findDependency(result.Dependencies, "github.com/example/pkg")
	assert.NotNil(t, exampleDep)
	assert.True(t, exampleDep.External)

	// Check dot import has weak strength
	dotDep := findDependency(result.Dependencies, "github.com/dot/import")
	assert.NotNil(t, dotDep)
	assert.Equal(t, DependencyStrengthWeak, dotDep.Strength)

	// Check exports
	assert.Contains(t, result.Exports, "ExportedFunction")
	assert.Contains(t, result.Exports, "ExportedType")
	assert.Contains(t, result.Exports, "ExportedVar")
	assert.NotContains(t, result.Exports, "privateFunction")
	assert.NotContains(t, result.Exports, "privateVar")
}

func TestGoDependencyParser_ParseModule(t *testing.T) {
	// Create temporary module structure
	tempDir, err := os.MkdirTemp("", "go_module_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create go.mod
	goModContent := `module github.com/test/module

go 1.19

require (
	github.com/stretchr/testify v1.8.0
	go.uber.org/zap v1.24.0 // indirect
)
`
	err = os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goModContent), 0o644)
	require.NoError(t, err)

	// Create some Go files
	mainGo := filepath.Join(tempDir, "main.go")
	err = os.WriteFile(mainGo, []byte("package main\n\nfunc main() {}\n"), 0o644)
	require.NoError(t, err)

	subDir := filepath.Join(tempDir, "pkg")
	err = os.MkdirAll(subDir, 0o755)
	require.NoError(t, err)

	pkgGo := filepath.Join(subDir, "pkg.go")
	err = os.WriteFile(pkgGo, []byte("package pkg\n\nfunc Helper() {}\n"), 0o644)
	require.NoError(t, err)

	// Create test file (should be excluded)
	testGo := filepath.Join(tempDir, "main_test.go")
	err = os.WriteFile(testGo, []byte("package main\n\nfunc TestMain() {}\n"), 0o644)
	require.NoError(t, err)

	logger := zaptest.NewLogger(t)
	parser := NewGoDependencyParser(logger)

	result, err := parser.ParseModule(tempDir)
	require.NoError(t, err)

	assert.Equal(t, "github.com/test/module", result.ModulePath)
	assert.Equal(t, "go", result.Language)
	assert.Contains(t, result.Files, mainGo)
	assert.Contains(t, result.Files, pkgGo)
	assert.NotContains(t, result.Files, testGo) // test files excluded

	// Check external dependencies from go.mod
	testifyDep := findDependency(result.ExternalDeps, "github.com/stretchr/testify")
	assert.NotNil(t, testifyDep)
	assert.Equal(t, "v1.8.0", testifyDep.Version)
	assert.Equal(t, DependencyStrengthStrong, testifyDep.Strength)

	zapDep := findDependency(result.ExternalDeps, "go.uber.org/zap")
	assert.NotNil(t, zapDep)
	assert.Equal(t, "v1.24.0", zapDep.Version)
	assert.Equal(t, DependencyStrengthWeak, zapDep.Strength) // indirect
}

func TestGoDependencyParser_IsExternalImport(t *testing.T) {
	logger := zaptest.NewLogger(t)
	parser := NewGoDependencyParser(logger)

	tests := []struct {
		importPath string
		expected   bool
	}{
		{"fmt", false},
		{"os", false},
		{"path/filepath", false},
		{"net/http", false},
		{"encoding/json", false},
		{"github.com/example/pkg", true},
		{"golang.org/x/tools", true},
		{"go.uber.org/zap", true},
		{"example.com/package", true},
	}

	for _, tt := range tests {
		t.Run(tt.importPath, func(t *testing.T) {
			result := parser.isExternalImport(tt.importPath)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGoDependencyParser_ExtractModulePath(t *testing.T) {
	// Create temporary directory with go.mod
	tempDir, err := os.MkdirTemp("", "go_module_path_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	goModContent := "module github.com/test/project\n\ngo 1.19\n"
	err = os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goModContent), 0o644)
	require.NoError(t, err)

	// Create subdirectory
	subDir := filepath.Join(tempDir, "cmd", "server")
	err = os.MkdirAll(subDir, 0o755)
	require.NoError(t, err)

	logger := zaptest.NewLogger(t)
	parser := NewGoDependencyParser(logger)

	testFile := filepath.Join(subDir, "main.go")
	result := parser.extractModulePath(testFile, "main")

	expected := "github.com/test/project/cmd/server"
	assert.Equal(t, expected, result)
}

func TestGoDependencyParser_GetModuleNameFromGoMod(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "go_mod_name_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	goModFile := filepath.Join(tempDir, "go.mod")
	goModContent := `module github.com/example/project

go 1.19

require github.com/stretchr/testify v1.8.0
`
	err = os.WriteFile(goModFile, []byte(goModContent), 0o644)
	require.NoError(t, err)

	logger := zaptest.NewLogger(t)
	parser := NewGoDependencyParser(logger)

	result := parser.getModuleNameFromGoMod(goModFile)
	assert.Equal(t, "github.com/example/project", result)
}

func TestGoDependencyParser_CountLinesOfCode(t *testing.T) {
	tempFile, err := os.CreateTemp("", "go_loc_test_*.go")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())

	content := `package main

import "fmt"

// This is a comment
func main() {
	// Another comment
	fmt.Println("Hello")
	
	/* Block comment
	   spanning multiple lines */
	
	if true {
		fmt.Println("World")
	}
}
`
	_, err = tempFile.WriteString(content)
	require.NoError(t, err)
	tempFile.Close()

	logger := zaptest.NewLogger(t)
	parser := NewGoDependencyParser(logger)

	count := parser.countLinesOfCode(tempFile.Name())

	// Should count: package, import, func, fmt.Println, if, fmt.Println, }
	// Should not count: comments, empty lines
	expected := 7
	assert.Equal(t, expected, count)
}

// Helper function to find dependency by target
func findDependency(deps []*Dependency, target string) *Dependency {
	for _, dep := range deps {
		if dep.To == target {
			return dep
		}
	}
	return nil
}
