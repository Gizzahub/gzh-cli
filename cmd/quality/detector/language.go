// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package detector

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/Gizzahub/gzh-cli/cmd/quality/tools"
)

// LanguageInfo contains information about a detected language.
type LanguageInfo struct {
	Name       string            // Language name (e.g., "Go", "Python")
	Extensions []string          // File extensions (e.g., [".go", ".mod"])
	Files      []string          // Detected files
	Indicators []string          // Project indicators (e.g., ["go.mod", "main.go"])
	Confidence float64           // Detection confidence (0.0 - 1.0)
	Metadata   map[string]string // Additional metadata
}

// FileTypeDetector implements language detection based on file types and project indicators.
type FileTypeDetector struct {
	// Map of language name to detection rules
	rules map[string]*LanguageRule
}

// LanguageRule defines how to detect a specific language.
type LanguageRule struct {
	Name       string            // Language name
	Extensions []string          // File extensions to look for
	Indicators []string          // Project files that indicate this language
	Keywords   []string          // Keywords to look for in files
	Patterns   []string          // File name patterns
	MinFiles   int               // Minimum files needed for detection
	Weight     float64           // Base weight for confidence calculation
	Metadata   map[string]string // Additional metadata
}

// NewFileTypeDetector creates a new language detector with default rules.
func NewFileTypeDetector() *FileTypeDetector {
	detector := &FileTypeDetector{
		rules: make(map[string]*LanguageRule),
	}

	// Register default language detection rules
	detector.registerDefaultRules()
	return detector
}

// DetectLanguages scans a directory and returns detected languages.
func (d *FileTypeDetector) DetectLanguages(projectRoot string) ([]string, error) {
	languages, err := d.detectLanguagesWithInfo(projectRoot)
	if err != nil {
		return nil, err
	}

	// Extract language names
	result := make([]string, 0, len(languages))
	for _, lang := range languages {
		result = append(result, lang.Name)
	}

	return result, nil
}

// DetectLanguagesWithInfo returns detailed language detection information.
func (d *FileTypeDetector) detectLanguagesWithInfo(projectRoot string) ([]*LanguageInfo, error) {
	detected := make(map[string]*LanguageInfo)

	err := filepath.Walk(projectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}

		// Skip directories and hidden files/directories
		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip hidden files and common ignore patterns
		if d.shouldIgnoreFile(path, info.Name()) {
			return nil
		}

		// Check each language rule
		for _, rule := range d.rules {
			if d.matchesRule(path, info.Name(), rule) {
				if detected[rule.Name] == nil {
					detected[rule.Name] = &LanguageInfo{
						Name:       rule.Name,
						Extensions: rule.Extensions,
						Files:      make([]string, 0),
						Indicators: make([]string, 0),
						Confidence: 0.0,
						Metadata:   make(map[string]string),
					}
					// Copy metadata
					for k, v := range rule.Metadata {
						detected[rule.Name].Metadata[k] = v
					}
				}

				detected[rule.Name].Files = append(detected[rule.Name].Files, path)

				// Check if this is a project indicator
				for _, indicator := range rule.Indicators {
					if strings.HasSuffix(strings.ToLower(info.Name()), strings.ToLower(indicator)) {
						detected[rule.Name].Indicators = append(detected[rule.Name].Indicators, path)
						break
					}
				}
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Calculate confidence scores and filter out low-confidence detections
	result := make([]*LanguageInfo, 0, len(detected))
	for _, lang := range detected {
		lang.Confidence = d.calculateConfidence(lang)
		if lang.Confidence > 0.1 { // Only include languages with reasonable confidence
			result = append(result, lang)
		}
	}

	return result, nil
}

// GetFilesByLanguage returns files grouped by language.
func (d *FileTypeDetector) GetFilesByLanguage(projectRoot string, languages []string) (map[string][]string, error) {
	detected, err := d.detectLanguagesWithInfo(projectRoot)
	if err != nil {
		return nil, err
	}

	result := make(map[string][]string)
	langSet := make(map[string]bool)
	for _, lang := range languages {
		langSet[lang] = true
	}

	for _, lang := range detected {
		if len(languages) == 0 || langSet[lang.Name] {
			result[lang.Name] = lang.Files
		}
	}

	return result, nil
}

// matchesRule checks if a file matches a language rule.
func (d *FileTypeDetector) matchesRule(path, filename string, rule *LanguageRule) bool {
	// Check file extensions
	ext := strings.ToLower(filepath.Ext(filename))
	for _, ruleExt := range rule.Extensions {
		if strings.EqualFold(ext, ruleExt) {
			return true
		}
	}

	// Check file name patterns
	lowerFilename := strings.ToLower(filename)
	for _, pattern := range rule.Patterns {
		if strings.Contains(lowerFilename, strings.ToLower(pattern)) {
			return true
		}
	}

	return false
}

// shouldIgnoreFile determines if a file should be ignored during detection.
func (d *FileTypeDetector) shouldIgnoreFile(path, filename string) bool {
	// Common ignore patterns
	ignorePatterns := []string{
		"node_modules/", "vendor/", ".git/", ".svn/", ".hg/",
		"__pycache__/", ".pytest_cache/", ".mypy_cache/",
		"target/", "dist/", "build/", ".next/", ".nuxt/",
		".vscode/", ".idea/", "*.tmp", "*.temp", "*.log",
	}

	lowerPath := strings.ToLower(path)
	lowerFilename := strings.ToLower(filename)

	for _, pattern := range ignorePatterns {
		if strings.Contains(lowerPath, pattern) || strings.Contains(lowerFilename, pattern) {
			return true
		}
	}

	// Skip hidden files
	if strings.HasPrefix(filename, ".") && filename != "." {
		return true
	}

	return false
}

// calculateConfidence calculates the confidence score for a detected language.
func (d *FileTypeDetector) calculateConfidence(lang *LanguageInfo) float64 {
	rule := d.rules[lang.Name]
	if rule == nil {
		return 0.0
	}

	confidence := rule.Weight

	// Boost confidence based on number of files
	fileCount := float64(len(lang.Files))
	if fileCount >= float64(rule.MinFiles) {
		confidence += fileCount * 0.1
	} else {
		confidence *= fileCount / float64(rule.MinFiles)
	}

	// Boost confidence for project indicators
	if len(lang.Indicators) > 0 {
		confidence += float64(len(lang.Indicators)) * 0.3
	}

	// Cap confidence at 1.0
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// registerDefaultRules registers the default language detection rules.
func (d *FileTypeDetector) registerDefaultRules() {
	rules := []*LanguageRule{
		{
			Name:       "Go",
			Extensions: []string{".go"},
			Indicators: []string{"go.mod", "go.sum", "main.go"},
			Patterns:   []string{"main.go", "_test.go"},
			MinFiles:   1,
			Weight:     0.8,
			Metadata: map[string]string{
				"package_manager": "go mod",
				"build_tool":      "go build",
			},
		},
		{
			Name:       "Python",
			Extensions: []string{".py", ".pyx", ".pyi"},
			Indicators: []string{"requirements.txt", "pyproject.toml", "setup.py", "Pipfile", "poetry.lock"},
			Patterns:   []string{"__init__.py", "main.py"},
			MinFiles:   1,
			Weight:     0.8,
			Metadata: map[string]string{
				"package_manager": "pip",
				"build_tool":      "setuptools",
			},
		},
		{
			Name:       "JavaScript",
			Extensions: []string{".js", ".jsx", ".mjs", ".cjs"},
			Indicators: []string{"package.json", "package-lock.json", "yarn.lock"},
			Patterns:   []string{"index.js", "main.js", "app.js"},
			MinFiles:   1,
			Weight:     0.8,
			Metadata: map[string]string{
				"package_manager": "npm",
				"build_tool":      "webpack",
			},
		},
		{
			Name:       "TypeScript",
			Extensions: []string{".ts", ".tsx", ".d.ts"},
			Indicators: []string{"tsconfig.json", "package.json"},
			Patterns:   []string{"index.ts", "main.ts", "app.ts"},
			MinFiles:   1,
			Weight:     0.8,
			Metadata: map[string]string{
				"package_manager": "npm",
				"build_tool":      "tsc",
			},
		},
		{
			Name:       "Rust",
			Extensions: []string{".rs"},
			Indicators: []string{"Cargo.toml", "Cargo.lock"},
			Patterns:   []string{"main.rs", "lib.rs"},
			MinFiles:   1,
			Weight:     0.9,
			Metadata: map[string]string{
				"package_manager": "cargo",
				"build_tool":      "cargo",
			},
		},
		{
			Name:       "Java",
			Extensions: []string{".java"},
			Indicators: []string{"pom.xml", "build.gradle", "build.gradle.kts"},
			Patterns:   []string{"Main.java", "Application.java"},
			MinFiles:   1,
			Weight:     0.7,
			Metadata: map[string]string{
				"package_manager": "maven",
				"build_tool":      "maven",
			},
		},
		{
			Name:       "C",
			Extensions: []string{".c", ".h"},
			Indicators: []string{"CMakeLists.txt", "Makefile", "configure.ac"},
			Patterns:   []string{"main.c"},
			MinFiles:   1,
			Weight:     0.6,
			Metadata: map[string]string{
				"build_tool": "make",
			},
		},
		{
			Name:       "C++",
			Extensions: []string{".cpp", ".cxx", ".cc", ".hpp", ".hxx", ".hh"},
			Indicators: []string{"CMakeLists.txt", "Makefile"},
			Patterns:   []string{"main.cpp", "main.cxx"},
			MinFiles:   1,
			Weight:     0.6,
			Metadata: map[string]string{
				"build_tool": "make",
			},
		},
		{
			Name:       "Shell",
			Extensions: []string{".sh", ".bash", ".zsh", ".fish"},
			Indicators: []string{},
			Patterns:   []string{"install.sh", "build.sh", "deploy.sh"},
			MinFiles:   1,
			Weight:     0.4,
			Metadata: map[string]string{
				"interpreter": "bash",
			},
		},
		{
			Name:       "YAML",
			Extensions: []string{".yml", ".yaml"},
			Indicators: []string{".github/workflows", "docker-compose.yml"},
			Patterns:   []string{"config.yml", "docker-compose.yml"},
			MinFiles:   1,
			Weight:     0.3,
			Metadata: map[string]string{
				"config_format": "yaml",
			},
		},
		{
			Name:       "JSON",
			Extensions: []string{".json"},
			Indicators: []string{"package.json", "tsconfig.json"},
			Patterns:   []string{"package.json", "config.json"},
			MinFiles:   1,
			Weight:     0.3,
			Metadata: map[string]string{
				"config_format": "json",
			},
		},
		{
			Name:       "Markdown",
			Extensions: []string{".md", ".markdown"},
			Indicators: []string{"README.md", "CHANGELOG.md"},
			Patterns:   []string{"README.md", "docs/"},
			MinFiles:   1,
			Weight:     0.2,
			Metadata: map[string]string{
				"doc_format": "markdown",
			},
		},
	}

	for _, rule := range rules {
		d.rules[rule.Name] = rule
	}
}

// Ensure FileTypeDetector implements LanguageDetector.
var _ tools.LanguageDetector = (*FileTypeDetector)(nil)
