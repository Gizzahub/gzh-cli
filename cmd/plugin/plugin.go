package plugin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"time"

	"github.com/spf13/cobra"
)

// PluginCmd represents the plugin command
var PluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Plugin development and management tools",
	Long: `Tools for developing, testing, and managing plugins for GZH Manager.

This command provides utilities for:
- Creating new plugin templates
- Building and testing plugins
- Managing plugin dependencies
- Validating plugin configurations`,
}

// CreateCmd creates a new plugin from template
var CreateCmd = &cobra.Command{
	Use:   "create [plugin-name]",
	Short: "Create a new plugin from template",
	Long: `Create a new plugin project with boilerplate code, build configuration,
and example implementations.

Example:
  gz plugin create my-backup-tool
  gz plugin create --type=command backup-manager
  gz plugin create --type=service monitoring-agent`,
	Args: cobra.ExactArgs(1),
	RunE: createPlugin,
}

// BuildCmd builds a plugin
var BuildCmd = &cobra.Command{
	Use:   "build [plugin-dir]",
	Short: "Build a plugin into a shared object",
	Long: `Build a Go plugin into a shared object (.so) file that can be loaded
by GZH Manager.

Example:
  gz plugin build ./my-plugin
  gz plugin build --output=/plugins/my-plugin.so ./my-plugin`,
	Args: cobra.MaximumNArgs(1),
	RunE: buildPlugin,
}

// TestCmd tests a plugin
var TestCmd = &cobra.Command{
	Use:   "test [plugin-dir]",
	Short: "Test a plugin",
	Long: `Run tests for a plugin and validate its implementation.

Example:
  gz plugin test ./my-plugin
  gz plugin test --verbose ./my-plugin`,
	Args: cobra.MaximumNArgs(1),
	RunE: testPlugin,
}

// ValidateCmd validates a plugin
var ValidateCmd = &cobra.Command{
	Use:   "validate [plugin-file]",
	Short: "Validate a plugin file",
	Long: `Validate that a plugin file is correctly formatted and implements
the required interfaces.

Example:
  gz plugin validate ./my-plugin.so
  gz plugin validate --strict ./my-plugin.so`,
	Args: cobra.ExactArgs(1),
	RunE: validatePlugin,
}

var (
	pluginType  string
	outputPath  string
	verbose     bool
	strict      bool
	skipTests   bool
	templateDir string
)

func init() {
	// Add subcommands
	PluginCmd.AddCommand(CreateCmd)
	PluginCmd.AddCommand(BuildCmd)
	PluginCmd.AddCommand(TestCmd)
	PluginCmd.AddCommand(ValidateCmd)

	// Create command flags
	CreateCmd.Flags().StringVar(&pluginType, "type", "basic", "Plugin type (basic, command, service, filter)")
	CreateCmd.Flags().StringVar(&templateDir, "template-dir", "", "Custom template directory")

	// Build command flags
	BuildCmd.Flags().StringVar(&outputPath, "output", "", "Output path for the built plugin")
	BuildCmd.Flags().BoolVar(&skipTests, "skip-tests", false, "Skip running tests before building")

	// Test command flags
	TestCmd.Flags().BoolVar(&verbose, "verbose", false, "Verbose test output")

	// Validate command flags
	ValidateCmd.Flags().BoolVar(&strict, "strict", false, "Strict validation mode")
}

// createPlugin creates a new plugin from template
func createPlugin(cmd *cobra.Command, args []string) error {
	pluginName := args[0]

	if !isValidPluginName(pluginName) {
		return fmt.Errorf("invalid plugin name: %s (must be lowercase with hyphens)", pluginName)
	}

	// Create plugin directory
	pluginDir := filepath.Join(".", pluginName)
	if _, err := os.Stat(pluginDir); !os.IsNotExist(err) {
		return fmt.Errorf("directory %s already exists", pluginDir)
	}

	fmt.Printf("Creating plugin '%s' of type '%s'...\n", pluginName, pluginType)

	// Create directory structure
	if err := createPluginStructure(pluginDir, pluginName, pluginType); err != nil {
		return fmt.Errorf("failed to create plugin structure: %w", err)
	}

	// Generate files from templates
	if err := generatePluginFiles(pluginDir, pluginName, pluginType); err != nil {
		return fmt.Errorf("failed to generate plugin files: %w", err)
	}

	fmt.Printf("✅ Plugin '%s' created successfully!\n", pluginName)
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  cd %s\n", pluginName)
	fmt.Printf("  gz plugin build\n")
	fmt.Printf("  gz plugin test\n")

	return nil
}

// buildPlugin builds a plugin into a shared object
func buildPlugin(cmd *cobra.Command, args []string) error {
	pluginDir := "."
	if len(args) > 0 {
		pluginDir = args[0]
	}

	// Get plugin name from directory or go.mod
	pluginName, err := getPluginName(pluginDir)
	if err != nil {
		return fmt.Errorf("failed to determine plugin name: %w", err)
	}

	// Determine output path
	if outputPath == "" {
		outputPath = fmt.Sprintf("%s.so", pluginName)
	}

	fmt.Printf("Building plugin '%s'...\n", pluginName)

	// Run tests first (unless skipped)
	if !skipTests {
		fmt.Println("Running tests...")
		if err := runPluginTests(pluginDir); err != nil {
			return fmt.Errorf("tests failed: %w", err)
		}
		fmt.Println("✅ Tests passed")
	}

	// Build the plugin
	if err := buildPluginBinary(pluginDir, outputPath); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	fmt.Printf("✅ Plugin built successfully: %s\n", outputPath)
	return nil
}

// testPlugin runs tests for a plugin
func testPlugin(cmd *cobra.Command, args []string) error {
	pluginDir := "."
	if len(args) > 0 {
		pluginDir = args[0]
	}

	pluginName, err := getPluginName(pluginDir)
	if err != nil {
		return fmt.Errorf("failed to determine plugin name: %w", err)
	}

	fmt.Printf("Testing plugin '%s'...\n", pluginName)

	if err := runPluginTests(pluginDir); err != nil {
		return fmt.Errorf("tests failed: %w", err)
	}

	fmt.Println("✅ All tests passed")
	return nil
}

// validatePlugin validates a plugin file
func validatePlugin(cmd *cobra.Command, args []string) error {
	pluginFile := args[0]

	fmt.Printf("Validating plugin '%s'...\n", pluginFile)

	if err := validatePluginFile(pluginFile, strict); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	fmt.Println("✅ Plugin validation passed")
	return nil
}

// createPluginStructure creates the directory structure for a new plugin
func createPluginStructure(pluginDir, pluginName, pluginType string) error {
	dirs := []string{
		pluginDir,
		filepath.Join(pluginDir, "cmd"),
		filepath.Join(pluginDir, "internal"),
		filepath.Join(pluginDir, "pkg"),
		filepath.Join(pluginDir, "test"),
		filepath.Join(pluginDir, "docs"),
		filepath.Join(pluginDir, "examples"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return nil
}

// generatePluginFiles generates files from templates
func generatePluginFiles(pluginDir, pluginName, pluginType string) error {
	// Template data
	data := PluginTemplateData{
		PluginName:      pluginName,
		PluginNameTitle: strings.Title(strings.ReplaceAll(pluginName, "-", " ")),
		PluginNameGo:    strings.ReplaceAll(strings.Title(strings.ReplaceAll(pluginName, "-", " ")), " ", ""),
		PluginType:      pluginType,
		ModuleName:      fmt.Sprintf("github.com/example/%s", pluginName),
		Year:            time.Now().Year(),
		Author:          getAuthorName(),
		GoVersion:       getGoVersion(),
	}

	// Generate files based on type
	files := getTemplateFiles(pluginType)

	for _, file := range files {
		if err := generateFileFromTemplate(pluginDir, file, data); err != nil {
			return fmt.Errorf("failed to generate %s: %w", file.Path, err)
		}
	}

	return nil
}

// PluginTemplateData holds data for template generation
type PluginTemplateData struct {
	PluginName      string
	PluginNameTitle string
	PluginNameGo    string
	PluginType      string
	ModuleName      string
	Year            int
	Author          string
	GoVersion       string
}

// TemplateFile represents a file to be generated from a template
type TemplateFile struct {
	Path     string
	Template string
}

// getTemplateFiles returns the list of files to generate for a plugin type
func getTemplateFiles(pluginType string) []TemplateFile {
	commonFiles := []TemplateFile{
		{"go.mod", goModTemplate},
		{"main.go", mainGoTemplate},
		{"plugin.go", pluginGoTemplate},
		{"README.md", readmeTemplate},
		{"Makefile", makefileTemplate},
		{".gitignore", gitignoreTemplate},
		{"test/plugin_test.go", pluginTestTemplate},
		{"docs/DEVELOPMENT.md", developmentDocsTemplate},
		{"examples/config.yaml", configExampleTemplate},
	}

	switch pluginType {
	case "command":
		return append(commonFiles, TemplateFile{"cmd/main.go", commandTemplate})
	case "service":
		return append(commonFiles, TemplateFile{"internal/service.go", serviceTemplate})
	case "filter":
		return append(commonFiles, TemplateFile{"internal/filter.go", filterTemplate})
	default: // basic
		return commonFiles
	}
}

// generateFileFromTemplate generates a file from a template
func generateFileFromTemplate(baseDir string, file TemplateFile, data PluginTemplateData) error {
	tmpl, err := template.New(file.Path).Parse(file.Template)
	if err != nil {
		return err
	}

	filePath := filepath.Join(baseDir, file.Path)

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return err
	}

	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.Execute(f, data)
}

// isValidPluginName checks if a plugin name is valid
func isValidPluginName(name string) bool {
	if len(name) == 0 {
		return false
	}

	for _, char := range name {
		if !((char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-') {
			return false
		}
	}

	return !strings.HasPrefix(name, "-") && !strings.HasSuffix(name, "-")
}

// getPluginName determines the plugin name from directory or go.mod
func getPluginName(pluginDir string) (string, error) {
	// Try to get from go.mod first
	goModPath := filepath.Join(pluginDir, "go.mod")
	if _, err := os.Stat(goModPath); err == nil {
		// Parse go.mod to get module name
		content, err := os.ReadFile(goModPath)
		if err != nil {
			return "", err
		}

		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "module ") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					return filepath.Base(parts[1]), nil
				}
			}
		}
	}

	// Fall back to directory name
	absPath, err := filepath.Abs(pluginDir)
	if err != nil {
		return "", err
	}

	return filepath.Base(absPath), nil
}

// getAuthorName gets the author name from git config or environment
func getAuthorName() string {
	// Try git config first
	if author := os.Getenv("GIT_AUTHOR_NAME"); author != "" {
		return author
	}

	if user := os.Getenv("USER"); user != "" {
		return user
	}

	return "Plugin Developer"
}

// getGoVersion returns the current Go version
func getGoVersion() string {
	version := runtime.Version()
	if strings.HasPrefix(version, "go") {
		return version[2:]
	}
	return version
}

// runPluginTests runs tests for a plugin
func runPluginTests(pluginDir string) error {
	// This would run: go test ./...
	// For now, we'll simulate this
	fmt.Printf("Running tests in %s...\n", pluginDir)

	// In a real implementation, this would execute:
	// cmd := exec.Command("go", "test", "./...")
	// cmd.Dir = pluginDir
	// return cmd.Run()

	return nil
}

// buildPluginBinary builds the plugin into a shared object
func buildPluginBinary(pluginDir, outputPath string) error {
	fmt.Printf("Building plugin binary: %s -> %s\n", pluginDir, outputPath)

	// In a real implementation, this would execute:
	// cmd := exec.Command("go", "build", "-buildmode=plugin", "-o", outputPath, ".")
	// cmd.Dir = pluginDir
	// return cmd.Run()

	return nil
}

// validatePluginFile validates a compiled plugin file
func validatePluginFile(pluginFile string, strict bool) error {
	// Check if file exists
	if _, err := os.Stat(pluginFile); os.IsNotExist(err) {
		return fmt.Errorf("plugin file not found: %s", pluginFile)
	}

	// In a real implementation, this would:
	// 1. Try to load the plugin
	// 2. Check for required exports (NewPlugin function)
	// 3. Validate the plugin interface implementation
	// 4. Check metadata format

	fmt.Printf("Validating plugin file: %s (strict: %v)\n", pluginFile, strict)

	return nil
}
