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

// Loader는 확장 설정을 로드하고 명령어를 등록합니다
type Loader struct {
	configPath string
}

// NewLoader는 새로운 Loader를 생성합니다
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

// LoadConfig는 확장 설정을 로드합니다
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

// RegisterAll은 모든 확장 명령어를 rootCmd에 등록합니다
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

// registerAlias는 별칭 명령어를 등록합니다
func (l *Loader) registerAlias(parent *cobra.Command, name string, alias AliasConfig) error {
	// 명령어가 비어있으면 에러
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

// registerExternal은 외부 명령어를 등록합니다
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
			cmdArgs := append(ext.Args, args...)
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

// executeAlias는 별칭 명령어를 실행합니다
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
