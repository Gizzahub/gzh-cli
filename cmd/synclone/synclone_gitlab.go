// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package synclone

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/Gizzahub/gzh-cli/internal/app"
	gitlabpkg "github.com/Gizzahub/gzh-cli/pkg/gitlab"
	synclonepkg "github.com/Gizzahub/gzh-cli/pkg/synclone"
)

type syncCloneGitlabOptions struct {
	targetPath   string
	groupName    string
	recursively  bool
	strategy     string
	configFile   string
	useConfig    bool
	parallel     int
	maxRetries   int
	resume       bool
	progressMode string
	heartbeatSec int
	token        string
}

func defaultSyncCloneGitlabOptions() *syncCloneGitlabOptions {
	return &syncCloneGitlabOptions{
		strategy:     "reset",
		parallel:     10,
		maxRetries:   3,
		progressMode: "bar",
		heartbeatSec: 10,
	}
}

func newSyncCloneGitlabCmd(appCtx *app.AppContext) *cobra.Command {
	_ = appCtx
	o := defaultSyncCloneGitlabOptions()

	cmd := &cobra.Command{
		Use:          "gitlab",
		Short:        "Clone repositories from a GitLab group",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE:         o.run,
	}

	cmd.Flags().StringVarP(&o.targetPath, "target", "t", o.targetPath, "Target directory")
	cmd.Flags().StringVarP(&o.groupName, "group", "g", o.groupName, "Group ID or path")

	// Backward-compatible deprecated flags (hidden)
	cmd.Flags().StringVar(&o.targetPath, "targetPath", o.targetPath, "(deprecated) use --target instead")
	cmd.Flags().StringVar(&o.groupName, "groupName", o.groupName, "(deprecated) use --group instead")
	cmd.Flags().MarkDeprecated("targetPath", "use --target instead")
	cmd.Flags().MarkDeprecated("groupName", "use --group instead")
	cmd.Flags().MarkHidden("targetPath")
	cmd.Flags().MarkHidden("groupName")
	cmd.Flags().BoolVarP(&o.recursively, "recursively", "r", o.recursively, "recursively")
	cmd.Flags().StringVarP(&o.strategy, "strategy", "s", o.strategy, "Sync strategy: reset, pull, or fetch")
	cmd.Flags().StringVarP(&o.configFile, "config", "c", o.configFile, "Path to config file")
	cmd.Flags().BoolVar(&o.useConfig, "use-config", false, "Use config file from standard locations")
	cmd.Flags().IntVarP(&o.parallel, "parallel", "p", o.parallel, "Number of parallel workers for cloning")
	cmd.Flags().IntVar(&o.maxRetries, "max-retries", o.maxRetries, "Maximum retry attempts for failed operations")
	cmd.Flags().BoolVar(&o.resume, "resume", false, "Resume interrupted clone operation from saved state")
	cmd.Flags().StringVar(&o.progressMode, "progress-mode", o.progressMode, "Progress display mode: bar, dots, spinner, quiet")

	// 커스텀 GitLab 인스턴스 URL 지정(예: https://gitlab.company.com)
	var baseURL string
	cmd.Flags().StringVar(&baseURL, "base-url", "", "GitLab instance base URL (e.g., https://gitlab.company.com)")

	// 무출력 방지를 위한 하트비트 간격(초). 0이면 비활성화
	cmd.Flags().IntVar(&o.heartbeatSec, "heartbeat-interval", o.heartbeatSec, "Heartbeat interval in seconds (0 to disable)")
	// 토큰 플래그
	cmd.Flags().StringVar(&o.token, "token", "", "GitLab token for API access (or set GITLAB_TOKEN)")

	// Mark flags as required only if not using config
	cmd.MarkFlagsMutuallyExclusive("config", "use-config")
	// cmd.MarkFlagsOneRequired("targetPath", "config", "use-config")
	// cmd.MarkFlagsOneRequired("groupName", "config", "use-config")

	return cmd
}

func (o *syncCloneGitlabOptions) run(cmd *cobra.Command, args []string) error {
	// 플래그가 전혀 없는 경우: 에러 대신 도움말 출력 후 정상 종료
	if cmd.Flags().NFlag() == 0 {
		_ = cmd.Help()
		return nil
	}

	// base-url이 지정되면 GitLab API 베이스 설정 주입
	if f := cmd.Flags().Lookup("base-url"); f != nil {
		v := f.Value.String()
		if v != "" {
			gitlabpkg.SetBaseURL(v)
		}
	}

	// 토큰 설정: 플래그 우선, 없으면 ENV(GITLAB_TOKEN/GZH_GITLAB_TOKEN)
	token := o.token
	if token == "" {
		if v := os.Getenv("GITLAB_TOKEN"); v != "" {
			token = v
		} else if v := os.Getenv("GZH_GITLAB_TOKEN"); v != "" {
			token = v
		}
	}
	if token != "" {
		gitlabpkg.SetToken(token)
	}

	// 프리플라이트: 그룹 접근 가능 여부 확인 (비공개면 토큰 안내)
	if err := gitlabpkg.PreflightCheckGroupAccess(cmd.Context(), o.groupName); err != nil {
		return fmt.Errorf("preflight failed: %w", err)
	}

	// 그룹명 정규화: 공백 및 트레일링 슬래시 제거
	o.groupName = strings.TrimSpace(o.groupName)
	o.groupName = strings.Trim(o.groupName, "/")

	// Load config if specified
	if o.configFile != "" || o.useConfig {
		err := o.loadFromConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
	}

	if o.targetPath == "" || o.groupName == "" {
		return fmt.Errorf("both --target and --group must be specified")
	}

	// Validate strategy
	if o.strategy != "reset" && o.strategy != "pull" && o.strategy != "fetch" {
		return fmt.Errorf("invalid strategy: %s. Must be one of: reset, pull, fetch", o.strategy)
	}

	// 하트비트 출력: 긴 작업 중 무출력 방지 (한국어 주석)
	stopHeartbeat := make(chan struct{})
	if o.heartbeatSec > 0 {
		go func(intervalSec int) {
			ticker := time.NewTicker(time.Duration(intervalSec) * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-stopHeartbeat:
					return
				case <-ticker.C:
					fmt.Println("진행 중... (heartbeat)")
				}
			}
		}(o.heartbeatSec)
	}

	// Use resumable clone if requested or if parallel/worker pool is enabled
	ctx := cmd.Context()

	var err error
	if o.resume || o.parallel > 1 {
		err = gitlabpkg.RefreshAllResumable(ctx, o.targetPath, o.groupName, o.strategy, o.parallel, o.maxRetries, o.resume, o.progressMode)
	} else {
		err = gitlabpkg.RefreshAll(ctx, o.targetPath, o.groupName, o.strategy)
	}

	close(stopHeartbeat)

	if err != nil {
		return err
	}

	return nil
}

func (o *syncCloneGitlabOptions) loadFromConfig() error {
	var configPath string
	if o.configFile != "" {
		configPath = o.configFile
	}

	cfg, err := synclonepkg.LoadConfig(configPath)
	if err != nil {
		return err
	}

	// If groupName is specified via CLI, use it; otherwise get from config
	if o.groupName == "" {
		if cfg.Default.Gitlab.GroupName != "" {
			o.groupName = cfg.Default.Gitlab.GroupName
		} else {
			return fmt.Errorf("no group found in config")
		}
	}

	// Get config for the specific group
	groupConfig, err := cfg.GetGitlabGroupConfig(o.groupName)
	if err != nil {
		return err
	}

	// Apply config values (CLI flags take precedence)
	if o.targetPath == "" && groupConfig.RootPath != "" {
		o.targetPath = synclonepkg.ExpandPath(groupConfig.RootPath)
	}

	if !o.recursively && groupConfig.Recursive {
		o.recursively = groupConfig.Recursive
	}

	return nil
}
