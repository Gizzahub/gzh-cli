// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

// Package wrapper provides standard interface and utilities for integrating external libraries.
package wrapper

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DocGeneratorConfig contains configuration for document generation.
type DocGeneratorConfig struct {
	OutputDir string // Document output directory (default: docs/integration/)
}

// DocGenerator automatically generates wrapper documentation.
type DocGenerator struct {
	config DocGeneratorConfig
}

// NewDocGenerator creates a new document generator.
func NewDocGenerator(config DocGeneratorConfig) (*DocGenerator, error) {
	if config.OutputDir == "" {
		config.OutputDir = "docs/integration"
	}

	return &DocGenerator{
		config: config,
	}, nil
}

// Generate generates documentation for the wrapper.
func (g *DocGenerator) Generate(lib ExternalLibrary) error {
	meta := GetMetadata(lib)

	// 출력 파일명: {name}-wrapper.md
	filename := fmt.Sprintf("%s-wrapper.md", meta.Name)
	outputPath := filepath.Join(g.config.OutputDir, filename)

	// 출력 디렉토리 생성
	if err := os.MkdirAll(g.config.OutputDir, 0o750); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	// 파일 생성
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	// 문서 내용 생성
	content := g.generateContent(meta)

	// 파일에 쓰기
	if _, err := f.WriteString(content); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

// generateContent generates document content.
func (g *DocGenerator) generateContent(meta Metadata) string {
	libraryName := extractLibraryName(meta.Repository)
	modulePath := extractModulePath(meta.Repository)

	var sb strings.Builder

	// Header
	sb.WriteString(fmt.Sprintf("# %s Wrapper Documentation\n\n", meta.Name))
	sb.WriteString(fmt.Sprintf("**Status**: %s\n", meta.Status))
	sb.WriteString(fmt.Sprintf("**Version**: %s\n", meta.Version))
	sb.WriteString(fmt.Sprintf("**External Repository**: [%s](%s)\n\n", meta.Repository, meta.Repository))

	// Overview
	sb.WriteString("## Overview\n\n")
	sb.WriteString("This command is implemented in an external library and integrated into gz via wrapper pattern.\n")
	sb.WriteString("The wrapper provides seamless integration while maintaining separation of concerns.\n\n")

	// Dependencies
	sb.WriteString("## Dependencies\n\n")
	if len(meta.Dependencies) > 0 {
		sb.WriteString("Required external tools:\n\n")
		for _, dep := range meta.Dependencies {
			sb.WriteString(fmt.Sprintf("- %s\n", dep))
		}
	} else {
		sb.WriteString("No external dependencies required.\n")
	}
	sb.WriteString("\n")

	// Local Development
	sb.WriteString("## Local Development\n\n")
	sb.WriteString(fmt.Sprintf("To work on the %s command:\n\n", meta.Name))
	sb.WriteString("### 1. Clone External Library\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString(fmt.Sprintf("git clone %s ../%s\n", meta.Repository, libraryName))
	sb.WriteString("```\n\n")

	sb.WriteString("### 2. Update go.mod with Local Replace\n\n")
	sb.WriteString("```go\n")
	sb.WriteString("// Add to go.mod\n")
	sb.WriteString(fmt.Sprintf("replace %s => ../%s\n", modulePath, libraryName))
	sb.WriteString("```\n\n")

	sb.WriteString("### 3. Make Changes\n\n")
	sb.WriteString("Edit code in the external library repository.\n\n")

	sb.WriteString("### 4. Test Integration\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString("# Build gz with local library\n")
	sb.WriteString("go build -o gz ./cmd/gz\n\n")
	sb.WriteString("# Test the command\n")
	sb.WriteString(fmt.Sprintf("./gz %s --help\n", meta.Name))
	sb.WriteString("```\n\n")

	// Testing
	sb.WriteString("## Testing\n\n")
	sb.WriteString("### Test Wrapper\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString(fmt.Sprintf("go test ./cmd/%s_wrapper_test.go -v\n", meta.Name))
	sb.WriteString("```\n\n")

	sb.WriteString("### Test External Library\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString(fmt.Sprintf("cd ../%s\n", libraryName))
	sb.WriteString("go test ./... -v\n")
	sb.WriteString("```\n\n")

	// Version Updates
	sb.WriteString("## Version Updates\n\n")
	sb.WriteString("To update the external library version:\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString("# Update to specific version\n")
	sb.WriteString(fmt.Sprintf("go get %s@v1.2.3\n\n", modulePath))
	sb.WriteString("# Update to latest\n")
	sb.WriteString(fmt.Sprintf("go get %s@latest\n\n", modulePath))
	sb.WriteString("# Verify update\n")
	sb.WriteString("go mod tidy\n")
	sb.WriteString("make build\n")
	sb.WriteString("make test\n")
	sb.WriteString("```\n\n")

	// References
	sb.WriteString("## References\n\n")
	sb.WriteString(fmt.Sprintf("- External Library: %s\n", meta.Repository))
	sb.WriteString(fmt.Sprintf("- Wrapper Code: [cmd/%s_wrapper.go](../../cmd/%s_wrapper.go)\n", meta.Name, meta.Name))
	sb.WriteString("- Registry Integration: [cmd/registry/registry.go](../../cmd/registry/registry.go)\n\n")

	sb.WriteString("---\n\n")
	sb.WriteString("**Generated**: Automatically generated wrapper documentation\n")
	sb.WriteString("**Maintainer**: See external library repository\n")

	return sb.String()
}

// extractLibraryName extracts library name from repository URL.
// Example: https://github.com/gizzahub/gzh-cli-quality -> gzh-cli-quality.
func extractLibraryName(repoURL string) string {
	parts := strings.Split(repoURL, "/")
	if len(parts) > 0 {
		name := parts[len(parts)-1]
		// .git 제거
		return strings.TrimSuffix(name, ".git")
	}
	return "unknown"
}

// extractModulePath extracts Go module path from repository URL.
// Example: https://github.com/gizzahub/gzh-cli-quality -> github.com/gizzahub/gzh-cli-quality.
func extractModulePath(repoURL string) string {
	// https:// 또는 http:// 제거
	path := strings.TrimPrefix(repoURL, "https://")
	path = strings.TrimPrefix(path, "http://")
	// .git 제거
	path = strings.TrimSuffix(path, ".git")
	return path
}
