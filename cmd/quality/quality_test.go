//nolint:testpackage // White-box testing needed for internal function access
package quality

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Gizzahub/gzh-cli/cmd/quality/detector"
	"github.com/Gizzahub/gzh-cli/cmd/quality/tools"
)

func TestNewQualityManager(t *testing.T) {
	manager := NewQualityManager()

	assert.NotNil(t, manager)
	assert.NotNil(t, manager.registry)
	assert.NotNil(t, manager.analyzer)
	assert.NotNil(t, manager.executor)
	assert.NotNil(t, manager.planner)
}

func TestNewQualityCmd(t *testing.T) {
	cmd := NewQualityCmd()

	assert.Equal(t, "quality", cmd.Use)
	assert.Contains(t, cmd.Short, "통합 코드 품질 도구")
	assert.NotEmpty(t, cmd.Long)
	assert.NotNil(t, cmd.RunE)

	// Check that subcommands are added
	subcommands := cmd.Commands()
	assert.Greater(t, len(subcommands), 5)

	// Verify expected subcommands exist (match actual cobra Use field)
	subcommandNames := make(map[string]bool)
	for _, subcmd := range subcommands {
		// Extract the command name (first word from Use field)
		cmdName := subcmd.Name()
		subcommandNames[cmdName] = true
	}

	expectedSubcommands := []string{"run", "check", "init", "analyze", "install", "upgrade", "version", "list", "tool"}
	for _, expected := range expectedSubcommands {
		assert.True(t, subcommandNames[expected], "Subcommand %s should exist", expected)
	}
}

func TestQualityManagerRunCmd(t *testing.T) {
	manager := NewQualityManager()
	cmd := manager.newRunCmd()

	assert.Equal(t, "run", cmd.Use)
	assert.Contains(t, cmd.Short, "모든 포매팅 및 린팅 도구 실행")
	assert.NotEmpty(t, cmd.Long)
	assert.NotNil(t, cmd.RunE)

	// Check flags exist
	flags := []string{"files", "fix", "format-only", "lint-only", "workers", "extra-args", "dry-run", "verbose", "report", "output", "since", "staged", "changed"}
	for _, flagName := range flags {
		flag := cmd.Flags().Lookup(flagName)
		assert.NotNil(t, flag, "Flag %s should exist", flagName)
	}
}

func TestQualityManagerCheckCmd(t *testing.T) {
	manager := NewQualityManager()
	cmd := manager.newCheckCmd()

	assert.Equal(t, "check", cmd.Use)
	assert.Contains(t, cmd.Short, "린팅만 실행")
	assert.NotEmpty(t, cmd.Long)
	assert.NotNil(t, cmd.RunE)

	// Check flags exist
	flags := []string{"files", "workers", "extra-args", "dry-run", "verbose", "report", "output", "since", "staged", "changed"}
	for _, flagName := range flags {
		flag := cmd.Flags().Lookup(flagName)
		assert.NotNil(t, flag, "Flag %s should exist", flagName)
	}
}

func TestQualityManagerAnalyzeCmd(t *testing.T) {
	manager := NewQualityManager()
	cmd := manager.newAnalyzeCmd()

	assert.Equal(t, "analyze", cmd.Use)
	assert.Contains(t, cmd.Short, "프로젝트 분석")
	assert.NotNil(t, cmd.RunE)
}

func TestQualityManagerInstallCmd(t *testing.T) {
	manager := NewQualityManager()
	cmd := manager.newInstallCmd()

	assert.Equal(t, "install [tool-name...]", cmd.Use)
	assert.Contains(t, cmd.Short, "품질 도구 설치")
	assert.NotEmpty(t, cmd.Long)
	assert.NotNil(t, cmd.RunE)
}

func TestQualityManagerUpgradeCmd(t *testing.T) {
	manager := NewQualityManager()
	cmd := manager.newUpgradeCmd()

	assert.Equal(t, "upgrade [tool-name...]", cmd.Use)
	assert.Contains(t, cmd.Short, "품질 도구 업그레이드")
	assert.NotEmpty(t, cmd.Long)
	assert.NotNil(t, cmd.RunE)
}

func TestQualityManagerVersionCmd(t *testing.T) {
	manager := NewQualityManager()
	cmd := manager.newVersionCmd()

	assert.Equal(t, "version [tool-name...]", cmd.Use)
	assert.Contains(t, cmd.Short, "품질 도구 버전 확인")
	assert.NotEmpty(t, cmd.Long)
	assert.NotNil(t, cmd.RunE)
}

func TestQualityManagerListCmd(t *testing.T) {
	manager := NewQualityManager()
	cmd := manager.newListCmd()

	assert.Equal(t, "list", cmd.Use)
	assert.Contains(t, cmd.Short, "사용 가능한 품질 도구 목록 표시")
	assert.NotNil(t, cmd.RunE)
}

func TestQualityManagerInitCmd(t *testing.T) {
	manager := NewQualityManager()
	cmd := manager.newInitCmd()

	assert.Equal(t, "init", cmd.Use)
	assert.Contains(t, cmd.Short, "프로젝트 설정 파일 자동 생성")
	assert.NotEmpty(t, cmd.Long)
	assert.NotNil(t, cmd.RunE)
}

func TestQualityManagerToolCmd(t *testing.T) {
	manager := NewQualityManager()
	cmd := manager.newToolCmd()

	assert.Equal(t, "tool [tool-name]", cmd.Use)
	assert.Contains(t, cmd.Short, "개별 도구 직접 실행")
	assert.NotEmpty(t, cmd.Long)
	assert.NotNil(t, cmd.RunE)

	// Should have tool-specific subcommands
	subcommands := cmd.Commands()
	assert.Greater(t, len(subcommands), 0, "Should have tool-specific subcommands")
}

func TestDisplayPlan(t *testing.T) {
	manager := NewQualityManager()

	plan := &tools.ExecutionPlan{
		Tasks: []tools.Task{
			{
				Tool:  &mockTool{name: "test-tool", language: "Go", toolType: tools.FORMAT},
				Files: []string{"test.go"},
			},
		},
		TotalFiles:        1,
		EstimatedDuration: "1s",
	}

	// This should not panic
	assert.NotPanics(t, func() {
		manager.displayPlan(plan, false)
	})

	assert.NotPanics(t, func() {
		manager.displayPlan(plan, true)
	})
}

func TestDisplayResults(t *testing.T) {
	manager := NewQualityManager()

	results := []*tools.Result{
		{
			Tool:           "test-tool",
			Language:       "Go",
			Success:        true,
			FilesProcessed: 1,
			Duration:       "100ms",
			Issues:         []tools.Issue{},
			Error:          nil,
		},
		{
			Tool:           "test-tool-2",
			Language:       "Python",
			Success:        false,
			FilesProcessed: 1,
			Duration:       "200ms",
			Issues: []tools.Issue{
				{
					File:    "test.py",
					Line:    1,
					Column:  1,
					Message: "Test issue",
					Rule:    "test-rule",
				},
			},
			Error: assert.AnError,
		},
	}

	// This should not panic
	assert.NotPanics(t, func() {
		manager.displayResults(results, time.Second, false)
	})

	assert.NotPanics(t, func() {
		manager.displayResults(results, time.Second, true)
	})
}

func TestValidateGitFlags(t *testing.T) {
	manager := NewQualityManager()

	tests := []struct {
		name     string
		since    string
		staged   bool
		changed  bool
		hasError bool
	}{
		{"no flags", "", false, false, false},
		{"only since", "HEAD~1", false, false, false},
		{"only staged", "", true, false, false},
		{"only changed", "", false, true, false},
		{"since and staged", "HEAD~1", true, false, true},
		{"staged and changed", "", true, true, true},
		{"since and changed", "HEAD~1", false, true, true},
		{"all flags", "HEAD~1", true, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.validateGitFlags(tt.since, tt.staged, tt.changed)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigGeneration(t *testing.T) {
	manager := NewQualityManager()

	// Create a real analysis result
	analysis := &detector.AnalysisResult{
		Languages: map[string][]string{
			"Go":     {"main.go", "test.go"},
			"Python": {"script.py"},
		},
	}

	config := manager.generateConfig(analysis)

	assert.NotNil(t, config)
	assert.True(t, config.Enabled)
	assert.NotNil(t, config.Languages)

	// Check Go language config
	if goConfig, exists := config.Languages["Go"]; exists {
		assert.True(t, goConfig.Enabled)
		assert.NotNil(t, goConfig.Tools)
	}

	// Check Python language config
	if pythonConfig, exists := config.Languages["Python"]; exists {
		assert.True(t, pythonConfig.Enabled)
		assert.NotNil(t, pythonConfig.Tools)
	}
}

func TestConfigToYAML(t *testing.T) {
	config := &Config{
		Enabled: true,
		Languages: map[string]*LanguageConfig{
			"Go": {
				Enabled: true,
				Tools: map[string]*ToolConfig{
					"gofumpt": {Enabled: true},
				},
			},
		},
	}

	yaml, err := config.ToYAML()
	require.NoError(t, err)
	assert.NotEmpty(t, yaml)
	assert.Contains(t, yaml, "enabled: true")
	assert.Contains(t, yaml, "Go:")
	assert.Contains(t, yaml, "gofumpt:")
}

func TestGetLanguageList(t *testing.T) {
	languages := map[string][]string{
		"Go":     {"main.go"},
		"Python": {"script.py"},
		"Rust":   {"main.rs"},
	}

	list := getLanguageList(languages)
	assert.Len(t, list, 3)
	assert.Contains(t, list, "Go")
	assert.Contains(t, list, "Python")
	assert.Contains(t, list, "Rust")
}

func TestContains(t *testing.T) {
	languages := map[string][]string{
		"Go":     {"main.go"},
		"Python": {"script.py"},
	}

	assert.True(t, contains(languages, "Go"))
	assert.True(t, contains(languages, "Python"))
	assert.False(t, contains(languages, "Rust"))
	assert.False(t, contains(languages, "JavaScript"))
}

func TestProjectAnalyzerAdapter(t *testing.T) {
	// Skip test since it requires complex setup with detector package
	t.Skip("ProjectAnalyzerAdapter requires complex setup, skipping for basic coverage")
}

// Mock implementations for testing

type mockTool struct {
	name     string
	language string
	toolType tools.ToolType
}

func (m *mockTool) Name() string                { return m.name }
func (m *mockTool) Language() string            { return m.language }
func (m *mockTool) Type() tools.ToolType        { return m.toolType }
func (m *mockTool) IsAvailable() bool           { return true }
func (m *mockTool) Install() error              { return nil }
func (m *mockTool) Upgrade() error              { return nil }
func (m *mockTool) GetVersion() (string, error) { return "1.0.0", nil }
func (m *mockTool) Execute(ctx context.Context, files []string, options tools.ExecuteOptions) (*tools.Result, error) {
	return &tools.Result{
		Tool:           m.name,
		Language:       m.language,
		Success:        true,
		FilesProcessed: 1,
		Duration:       "1ms",
		Issues:         []tools.Issue{},
	}, nil
}

func (m *mockTool) FindConfigFiles(projectRoot string) []string {
	return []string{}
}
