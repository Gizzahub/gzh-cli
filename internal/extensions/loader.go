// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package extensions

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// Loader loads extension configuration and registers commands.
type Loader struct {
	configPath string
}

// NewLoader creates a new Loader.
func NewLoader() *Loader {
	// 기본 설정 경로: ~/.config/gzh-manager/extensions.yaml
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return &Loader{configPath: ""}
	}

	return &Loader{
		configPath: filepath.Join(homeDir, ".config", "gzh-manager", "extensions.yaml"),
	}
}

// LoadConfig loads extension configuration.
func (l *Loader) LoadConfig() (*Config, error) {
	// 설정 파일이 없으면 빈 설정 반환 (에러 아님)
	if _, err := os.Stat(l.configPath); os.IsNotExist(err) {
		return &Config{
			Aliases:  make(map[string]AliasConfig),
			External: []ExternalCommandConfig{},
		}, nil
	}

	data, err := os.ReadFile(l.configPath)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	return &cfg, nil
}

// RegisterAll registers all extension commands to rootCmd.
func (l *Loader) RegisterAll(rootCmd *cobra.Command) error {
	cfg, err := l.LoadConfig()
	if err != nil {
		return err
	}

	// 별칭 등록
	for name, alias := range cfg.Aliases {
		if err := l.registerAlias(rootCmd, name, alias); err != nil {
			// 개별 별칭 등록 실패는 경고만 출력하고 계속 진행
			fmt.Fprintf(os.Stderr, "⚠️  Failed to register alias '%s': %v\n", name, err)
		}
	}

	// 외부 명령어 등록
	for _, ext := range cfg.External {
		if err := l.registerExternal(rootCmd, ext); err != nil {
			// 개별 외부 명령어 등록 실패는 경고만 출력하고 계속 진행
			fmt.Fprintf(os.Stderr, "⚠️  Failed to register external command '%s': %v\n", ext.Name, err)
		}
	}

	return nil
}

// registerAlias registers an alias command.
func (l *Loader) registerAlias(parent *cobra.Command, name string, alias AliasConfig) error {
	// Multi-step workflow or single command
	if len(alias.Steps) > 0 {
		return l.registerWorkflowAlias(parent, name, alias)
	}

	// Parameterized alias or simple alias
	if len(alias.Params) > 0 {
		return l.registerParameterizedAlias(parent, name, alias)
	}

	// Simple alias
	if alias.Command == "" {
		return fmt.Errorf("alias command is empty")
	}

	cmd := &cobra.Command{
		Use:   name,
		Short: alias.Description,
		Long:  fmt.Sprintf("%s\n\n[ALIAS] This is a user-defined alias command.", alias.Description),
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeAlias(alias.Command, args)
		},
		// 별칭은 숨김 처리하지 않음 (사용자가 추가한 것이므로)
	}

	parent.AddCommand(cmd)
	return nil
}

// registerWorkflowAlias registers a multi-step workflow alias.
func (l *Loader) registerWorkflowAlias(parent *cobra.Command, name string, alias AliasConfig) error {
	if len(alias.Steps) == 0 {
		return fmt.Errorf("workflow has no steps")
	}

	cmd := &cobra.Command{
		Use:   name,
		Short: alias.Description,
		Long:  fmt.Sprintf("%s\n\n[WORKFLOW] This executes multiple commands in sequence:\n%s", alias.Description, formatSteps(alias.Steps)),
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeWorkflow(alias.Steps, args)
		},
	}

	parent.AddCommand(cmd)
	return nil
}

// registerParameterizedAlias registers a parameterized alias.
func (l *Loader) registerParameterizedAlias(parent *cobra.Command, name string, alias AliasConfig) error {
	if alias.Command == "" {
		return fmt.Errorf("parameterized alias command is empty")
	}

	// Use 문자열에 파라미터 추가
	use := name
	for _, param := range alias.Params {
		if param.Required {
			use += fmt.Sprintf(" <%s>", param.Name)
		} else {
			use += fmt.Sprintf(" [%s]", param.Name)
		}
	}

	cmd := &cobra.Command{
		Use:   use,
		Short: alias.Description,
		Long:  fmt.Sprintf("%s\n\n[PARAMETERIZED] Parameters:\n%s", alias.Description, formatParams(alias.Params)),
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeParameterizedAlias(alias.Command, alias.Params, args)
		},
	}

	parent.AddCommand(cmd)
	return nil
}

// registerExternal registers an external command.
func (l *Loader) registerExternal(parent *cobra.Command, ext ExternalCommandConfig) error {
	// 외부 명령어가 존재하는지 확인
	if _, err := exec.LookPath(ext.Command); err != nil {
		// 명령어가 없으면 경고만 출력하고 등록하지 않음
		return fmt.Errorf("command not found: %s", ext.Command)
	}

	cmd := &cobra.Command{
		Use:   ext.Name,
		Short: fmt.Sprintf("[EXTERNAL] %s", ext.Description),
		Long:  fmt.Sprintf("%s\n\nThis command is integrated from external source: %s", ext.Description, ext.Command),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdArgs := make([]string, 0, len(ext.Args)+len(args))
			cmdArgs = append(cmdArgs, ext.Args...)
			cmdArgs = append(cmdArgs, args...)
			execCmd := exec.Command(ext.Command, cmdArgs...)
			execCmd.Stdin = os.Stdin
			execCmd.Stdout = os.Stdout
			execCmd.Stderr = os.Stderr
			return execCmd.Run()
		},
		DisableFlagParsing: ext.Passthrough,
	}

	parent.AddCommand(cmd)
	return nil
}

// executeAlias executes an alias command.
func executeAlias(aliasCmd string, args []string) error {
	// 별칭 명령어 파싱
	parts := strings.Fields(aliasCmd)
	if len(parts) == 0 {
		return fmt.Errorf("empty alias command")
	}

	// gz 명령어로 실행
	// 예: "git repo pull-all" -> gz git repo pull-all [args...]
	gzPath, err := exec.LookPath("gz")
	if err != nil {
		// gz를 찾을 수 없으면 현재 실행 파일 사용
		gzPath = os.Args[0]
	}

	cmd := exec.Command(gzPath, append(parts, args...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// executeWorkflow executes multiple commands in sequence.
func executeWorkflow(steps, args []string) error {
	gzPath, err := exec.LookPath("gz")
	if err != nil {
		gzPath = os.Args[0]
	}

	for i, step := range steps {
		fmt.Fprintf(os.Stderr, "🔄 Step %d/%d: %s\n", i+1, len(steps), step)

		parts := strings.Fields(step)
		if len(parts) == 0 {
			return fmt.Errorf("empty step at index %d", i)
		}

		// Execute each step
		cmd := exec.Command(gzPath, parts...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("step %d failed: %w", i+1, err)
		}

		fmt.Fprintf(os.Stderr, "✅ Step %d/%d completed\n\n", i+1, len(steps))
	}

	fmt.Fprintf(os.Stderr, "🎉 All steps completed successfully!\n")
	return nil
}

// executeParameterizedAlias executes an alias with parameter substitution.
func executeParameterizedAlias(aliasCmd string, params []Param, args []string) error {
	// Validate required parameters
	requiredCount := 0
	for _, p := range params {
		if p.Required {
			requiredCount++
		}
	}

	if len(args) < requiredCount {
		return fmt.Errorf("missing required parameters: expected at least %d, got %d", requiredCount, len(args))
	}

	// Build substitution map
	substitutions := make(map[string]string)
	for i, param := range params {
		if i < len(args) {
			substitutions[fmt.Sprintf("${%s}", param.Name)] = args[i]
			substitutions[fmt.Sprintf("$%s", param.Name)] = args[i]
		}
	}

	// Perform substitution
	expandedCmd := aliasCmd
	for placeholder, value := range substitutions {
		expandedCmd = strings.ReplaceAll(expandedCmd, placeholder, value)
	}

	// Execute the expanded command
	parts := strings.Fields(expandedCmd)
	if len(parts) == 0 {
		return fmt.Errorf("empty expanded command")
	}

	gzPath, err := exec.LookPath("gz")
	if err != nil {
		gzPath = os.Args[0]
	}

	// Append remaining args after parameters
	remainingArgs := args[len(params):]
	cmd := exec.Command(gzPath, append(parts, remainingArgs...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// formatSteps formats workflow steps for display.
func formatSteps(steps []string) string {
	var sb strings.Builder
	for i, step := range steps {
		sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, step))
	}
	return sb.String()
}

// formatParams formats parameters for display.
func formatParams(params []Param) string {
	var sb strings.Builder
	for _, param := range params {
		required := ""
		if param.Required {
			required = " (required)"
		}
		sb.WriteString(fmt.Sprintf("  - %s%s: %s\n", param.Name, required, param.Description))
	}
	return sb.String()
}
