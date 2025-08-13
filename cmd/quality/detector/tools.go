// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package detector

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Gizzahub/gzh-manager-go/cmd/quality/tools"
)

// SystemToolDetector detects available quality tools on the system.
type SystemToolDetector struct {
	// pathCache caches tool availability results
	pathCache map[string]bool
}

// NewSystemToolDetector creates a new system tool detector.
func NewSystemToolDetector() *SystemToolDetector {
	return &SystemToolDetector{
		pathCache: make(map[string]bool),
	}
}

// IsToolAvailable checks if a tool is available on the system.
func (d *SystemToolDetector) IsToolAvailable(toolName string) bool {
	// Check cache first
	if available, cached := d.pathCache[toolName]; cached {
		return available
	}

	// Check system PATH
	_, err := exec.LookPath(toolName)
	available := err == nil

	// Check common installation locations if not found in PATH
	if !available {
		available = d.checkCommonLocations(toolName)
	}

	// Cache the result
	d.pathCache[toolName] = available
	return available
}

// checkCommonLocations checks for tools in common installation locations.
func (d *SystemToolDetector) checkCommonLocations(toolName string) bool {
	commonPaths := []string{
		"/usr/local/bin",
		"/opt/homebrew/bin",
		filepath.Join(os.Getenv("HOME"), "go", "bin"),
		filepath.Join(os.Getenv("HOME"), ".local", "bin"),
		filepath.Join(os.Getenv("HOME"), ".cargo", "bin"),
		filepath.Join(os.Getenv("HOME"), ".npm-global", "bin"),
		filepath.Join(os.Getenv("GOPATH"), "bin"),
	}

	for _, path := range commonPaths {
		if path == "" {
			continue
		}
		toolPath := filepath.Join(path, toolName)
		if _, err := os.Stat(toolPath); err == nil {
			return true
		}
	}

	return false
}

// GetToolVersion returns the version of a tool if available.
func (d *SystemToolDetector) GetToolVersion(toolName string) string {
	if !d.IsToolAvailable(toolName) {
		return ""
	}

	// Common version flags to try
	versionFlags := []string{"--version", "-version", "-V", "-v"}

	for _, flag := range versionFlags {
		cmd := exec.Command(toolName, flag)
		output, err := cmd.CombinedOutput()
		if err == nil {
			version := strings.TrimSpace(string(output))
			if version != "" {
				return strings.Split(version, "\n")[0] // Return first line
			}
		}
	}

	return "unknown"
}

// ConfigFileDetector finds configuration files for quality tools.
type ConfigFileDetector struct{}

// NewConfigFileDetector creates a new configuration file detector.
func NewConfigFileDetector() *ConfigFileDetector {
	return &ConfigFileDetector{}
}

// FindConfigs searches for tool configuration files.
func (d *ConfigFileDetector) FindConfigs(projectRoot string, toolList []tools.QualityTool) map[string]string {
	configs := make(map[string]string)

	for _, tool := range toolList {
		configFiles := tool.FindConfigFiles(projectRoot)
		for _, configFile := range configFiles {
			if _, err := os.Stat(configFile); err == nil {
				configs[tool.Name()] = configFile
				break // Use first found config
			}
		}
	}

	return configs
}

// ValidateConfig checks if a configuration file is valid.
func (d *ConfigFileDetector) ValidateConfig(toolName, configPath string) error {
	// Basic validation - check if file exists and is readable
	if _, err := os.Stat(configPath); err != nil {
		return err
	}

	file, err := os.Open(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Tool-specific validation could be added here
	// For now, we just verify the file is accessible
	return nil
}

// ProjectAnalyzer analyzes a project to determine quality tool setup.
type ProjectAnalyzer struct {
	langDetector   *FileTypeDetector
	toolDetector   *SystemToolDetector
	configDetector *ConfigFileDetector
}

// NewProjectAnalyzer creates a new project analyzer.
func NewProjectAnalyzer() *ProjectAnalyzer {
	return &ProjectAnalyzer{
		langDetector:   NewFileTypeDetector(),
		toolDetector:   NewSystemToolDetector(),
		configDetector: NewConfigFileDetector(),
	}
}

// AnalysisResult contains the results of project analysis.
type AnalysisResult struct {
	// ProjectRoot is the root directory of the project
	ProjectRoot string

	// Languages contains detected languages with their files
	Languages map[string][]string

	// AvailableTools lists tools that are installed and available
	AvailableTools []string

	// RecommendedTools suggests tools for the detected languages
	RecommendedTools map[string][]string

	// ConfigFiles maps tool names to their configuration files
	ConfigFiles map[string]string

	// Issues contains any problems detected during analysis
	Issues []string
}

// AnalyzeProject performs comprehensive project analysis.
func (a *ProjectAnalyzer) AnalyzeProject(projectRoot string, registry tools.ToolRegistry) (*AnalysisResult, error) {
	result := &AnalysisResult{
		ProjectRoot:      projectRoot,
		Languages:        make(map[string][]string),
		AvailableTools:   make([]string, 0),
		RecommendedTools: make(map[string][]string),
		ConfigFiles:      make(map[string]string),
		Issues:           make([]string, 0),
	}

	// Detect languages
	languages, err := a.langDetector.DetectLanguages(projectRoot)
	if err != nil {
		return nil, err
	}

	// Get files by language
	result.Languages, err = a.langDetector.GetFilesByLanguage(projectRoot, languages)
	if err != nil {
		return nil, err
	}

	// Get all tools from registry
	allTools := registry.GetTools()

	// Check tool availability and build recommendations
	for _, lang := range languages {
		langTools := registry.GetToolsByLanguage(lang)
		recommendations := make([]string, 0)

		for _, tool := range langTools {
			if a.toolDetector.IsToolAvailable(tool.Name()) {
				result.AvailableTools = append(result.AvailableTools, tool.Name())
				recommendations = append(recommendations, tool.Name())
			}
		}

		if len(recommendations) > 0 {
			result.RecommendedTools[lang] = recommendations
		} else {
			result.Issues = append(result.Issues, "No quality tools available for "+lang)
		}
	}

	// Find configuration files
	result.ConfigFiles = a.configDetector.FindConfigs(projectRoot, allTools)

	// Remove duplicates from available tools
	result.AvailableTools = removeDuplicates(result.AvailableTools)

	return result, nil
}

// GetOptimalToolSelection returns the best tools for each language.
func (a *ProjectAnalyzer) GetOptimalToolSelection(result *AnalysisResult, registry tools.ToolRegistry) map[string][]tools.QualityTool {
	selection := make(map[string][]tools.QualityTool)

	for lang, recommendedNames := range result.RecommendedTools {
		selectedTools := make([]tools.QualityTool, 0)

		// Prefer tools with configuration files
		for _, toolName := range recommendedNames {
			tool := registry.FindTool(toolName)
			if tool == nil {
				continue
			}

			// Prioritize tools that have configuration files
			if _, hasConfig := result.ConfigFiles[toolName]; hasConfig {
				selectedTools = append(selectedTools, tool)
			}
		}

		// Add remaining tools without configs
		for _, toolName := range recommendedNames {
			tool := registry.FindTool(toolName)
			if tool == nil {
				continue
			}

			// Skip if already added
			found := false
			for _, selected := range selectedTools {
				if selected.Name() == tool.Name() {
					found = true
					break
				}
			}

			if !found {
				selectedTools = append(selectedTools, tool)
			}
		}

		if len(selectedTools) > 0 {
			selection[lang] = selectedTools
		}
	}

	return selection
}

// removeDuplicates removes duplicate strings from a slice.
func removeDuplicates(slice []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0)

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

// Ensure detectors implement their interfaces.
var (
	_ tools.LanguageDetector = (*FileTypeDetector)(nil)
	_ tools.ConfigDetector   = (*ConfigFileDetector)(nil)
)
