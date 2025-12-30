// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package wrapper

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gizzahub/gzh-cli/internal/app"
)

func TestNewDocGenerator(t *testing.T) {
	config := DocGeneratorConfig{
		OutputDir: "test-output",
	}

	gen, err := NewDocGenerator(config)
	require.NoError(t, err)
	assert.NotNil(t, gen)
	assert.Equal(t, "test-output", gen.config.OutputDir)
}

func TestNewDocGenerator_DefaultOutputDir(t *testing.T) {
	config := DocGeneratorConfig{}

	gen, err := NewDocGenerator(config)
	require.NoError(t, err)
	assert.Equal(t, "docs/integration", gen.config.OutputDir)
}

func TestDocGenerator_Generate(t *testing.T) {
	tmpDir := t.TempDir()

	config := DocGeneratorConfig{
		OutputDir: tmpDir,
	}

	gen, err := NewDocGenerator(config)
	require.NoError(t, err)

	// 테스트용 라이브러리
	base := NewBaseWrapper(
		"quality",
		"1.0.0",
		"https://github.com/gizzahub/gzh-cli-quality",
		[]string{"go", "npm"},
	)
	lib := &mockLibrary{BaseWrapper: base}

	// 문서 생성
	err = gen.Generate(lib)
	require.NoError(t, err)

	// 파일이 생성되었는지 확인
	outputPath := filepath.Join(tmpDir, "quality-wrapper.md")
	_, err = os.Stat(outputPath)
	require.NoError(t, err, "Generated file should exist")

	// 파일 내용 확인
	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	contentStr := string(content)
	assert.Contains(t, contentStr, "# quality Wrapper Documentation")
	assert.Contains(t, contentStr, "**Version**: 1.0.0")
	assert.Contains(t, contentStr, "https://github.com/gizzahub/gzh-cli-quality")
	assert.Contains(t, contentStr, "- go")
	assert.Contains(t, contentStr, "- npm")
	assert.Contains(t, contentStr, "## Local Development")
}

func TestDocGenerator_Generate_NoDependencies(t *testing.T) {
	tmpDir := t.TempDir()

	config := DocGeneratorConfig{
		OutputDir: tmpDir,
	}

	gen, err := NewDocGenerator(config)
	require.NoError(t, err)

	// 의존성 없는 라이브러리
	base := NewBaseWrapper(
		"simple",
		"1.0.0",
		"https://github.com/test/simple",
		[]string{},
	)
	lib := &mockLibrary{BaseWrapper: base}

	err = gen.Generate(lib)
	require.NoError(t, err)

	outputPath := filepath.Join(tmpDir, "simple-wrapper.md")
	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	contentStr := string(content)
	assert.Contains(t, contentStr, "No external dependencies required")
}

func TestExtractLibraryName(t *testing.T) {
	tests := []struct {
		name     string
		repoURL  string
		expected string
	}{
		{
			name:     "GitHub URL",
			repoURL:  "https://github.com/gizzahub/gzh-cli-quality",
			expected: "gzh-cli-quality",
		},
		{
			name:     "GitHub URL with .git",
			repoURL:  "https://github.com/gizzahub/gzh-cli-quality.git",
			expected: "gzh-cli-quality",
		},
		{
			name:     "Simple name",
			repoURL:  "mylib",
			expected: "mylib",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractLibraryName(tt.repoURL)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractModulePath(t *testing.T) {
	tests := []struct {
		name     string
		repoURL  string
		expected string
	}{
		{
			name:     "HTTPS GitHub URL",
			repoURL:  "https://github.com/gizzahub/gzh-cli-quality",
			expected: "github.com/gizzahub/gzh-cli-quality",
		},
		{
			name:     "HTTPS GitHub URL with .git",
			repoURL:  "https://github.com/gizzahub/gzh-cli-quality.git",
			expected: "github.com/gizzahub/gzh-cli-quality",
		},
		{
			name:     "HTTP GitHub URL",
			repoURL:  "http://github.com/gizzahub/gzh-cli-quality",
			expected: "github.com/gizzahub/gzh-cli-quality",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractModulePath(tt.repoURL)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDocGenerator_Generate_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	outputDir := filepath.Join(tmpDir, "nested", "output", "dir")

	config := DocGeneratorConfig{
		OutputDir: outputDir,
	}

	gen, err := NewDocGenerator(config)
	require.NoError(t, err)

	base := NewBaseWrapper("test", "1.0.0", "https://github.com/test/test", []string{})
	lib := &mockLibrary{BaseWrapper: base}

	err = gen.Generate(lib)
	require.NoError(t, err)

	// 디렉토리가 생성되었는지 확인
	_, err = os.Stat(outputDir)
	require.NoError(t, err, "Output directory should be created")
}

func TestDocGenerator_Generate_ValidMarkdown(t *testing.T) {
	tmpDir := t.TempDir()

	config := DocGeneratorConfig{
		OutputDir: tmpDir,
	}

	gen, err := NewDocGenerator(config)
	require.NoError(t, err)

	base := NewBaseWrapper(
		"quality",
		"1.2.3",
		"https://github.com/gizzahub/gzh-cli-quality",
		[]string{"go"},
	)
	lib := &mockLibrary{BaseWrapper: base}

	err = gen.Generate(lib)
	require.NoError(t, err)

	outputPath := filepath.Join(tmpDir, "quality-wrapper.md")
	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	contentStr := string(content)

	// 마크다운 구조 검증
	assert.True(t, strings.HasPrefix(contentStr, "# quality"), "Should start with h1 header")
	assert.Contains(t, contentStr, "## Overview")
	assert.Contains(t, contentStr, "## Dependencies")
	assert.Contains(t, contentStr, "## Local Development")
	assert.Contains(t, contentStr, "## Testing")
	assert.Contains(t, contentStr, "## Version Updates")
	assert.Contains(t, contentStr, "## References")

	// 코드 블록 검증
	assert.Contains(t, contentStr, "```bash")
	assert.Contains(t, contentStr, "```go")
}

// testLibrary is a mock library for testing.
type testLibrary struct {
	name         string
	version      string
	repository   string
	dependencies []string
}

func (m *testLibrary) Name() string           { return m.name }
func (m *testLibrary) Version() string        { return m.version }
func (m *testLibrary) Repository() string     { return m.repository }
func (m *testLibrary) Dependencies() []string { return m.dependencies }
func (m *testLibrary) Validate() error        { return nil }
func (m *testLibrary) CreateCommand(appCtx *app.AppContext) (*cobra.Command, error) {
	return &cobra.Command{Use: m.name}, nil
}
