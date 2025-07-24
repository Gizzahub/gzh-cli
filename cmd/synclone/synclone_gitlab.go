// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package synclone

import (
	"fmt"

	"github.com/spf13/cobra"

	gitlabpkg "github.com/gizzahub/gzh-manager-go/pkg/gitlab"
	synclonepkg "github.com/gizzahub/gzh-manager-go/pkg/synclone"
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
}

func defaultSyncCloneGitlabOptions() *syncCloneGitlabOptions {
	return &syncCloneGitlabOptions{
		strategy:     "reset",
		parallel:     10,
		maxRetries:   3,
		progressMode: "bar",
	}
}

func newSyncCloneGitlabCmd() *cobra.Command {
	o := defaultSyncCloneGitlabOptions()

	cmd := &cobra.Command{
		Use:   "gitlab",
		Short: "Clone repositories from a GitLab group",
		Args:  cobra.NoArgs,
		RunE:  o.run,
	}

	cmd.Flags().StringVarP(&o.targetPath, "targetPath", "t", o.targetPath, "targetPath")
	cmd.Flags().StringVarP(&o.groupName, "groupName", "g", o.groupName, "groupName")
	cmd.Flags().BoolVarP(&o.recursively, "recursively", "r", o.recursively, "recursively")
	cmd.Flags().StringVarP(&o.strategy, "strategy", "s", o.strategy, "Sync strategy: reset, pull, or fetch")
	cmd.Flags().StringVarP(&o.configFile, "config", "c", o.configFile, "Path to config file")
	cmd.Flags().BoolVar(&o.useConfig, "use-config", false, "Use config file from standard locations")
	cmd.Flags().IntVarP(&o.parallel, "parallel", "p", o.parallel, "Number of parallel workers for cloning")
	cmd.Flags().IntVar(&o.maxRetries, "max-retries", o.maxRetries, "Maximum retry attempts for failed operations")
	cmd.Flags().BoolVar(&o.resume, "resume", false, "Resume interrupted clone operation from saved state")
	cmd.Flags().StringVar(&o.progressMode, "progress-mode", o.progressMode, "Progress display mode: bar, dots, spinner, quiet")

	// Mark flags as required only if not using config
	cmd.MarkFlagsMutuallyExclusive("config", "use-config")
	cmd.MarkFlagsOneRequired("targetPath", "config", "use-config")
	cmd.MarkFlagsOneRequired("groupName", "config", "use-config")

	return cmd
}

func (o *syncCloneGitlabOptions) run(cmd *cobra.Command, args []string) error {
	// Load config if specified
	if o.configFile != "" || o.useConfig {
		err := o.loadFromConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
	}

	if o.targetPath == "" || o.groupName == "" {
		return fmt.Errorf("both targetPath and groupName must be specified")
	}

	// Validate strategy
	if o.strategy != "reset" && o.strategy != "pull" && o.strategy != "fetch" {
		return fmt.Errorf("invalid strategy: %s. Must be one of: reset, pull, fetch", o.strategy)
	}

	// Use resumable clone if requested or if parallel/worker pool is enabled
	ctx := cmd.Context()

	var err error
	if o.resume || o.parallel > 1 {
		err = gitlabpkg.RefreshAllResumable(ctx, o.targetPath, o.groupName, o.strategy, o.parallel, o.maxRetries, o.resume, o.progressMode)
	} else {
		err = gitlabpkg.RefreshAll(ctx, o.targetPath, o.groupName, o.strategy)
	}

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
