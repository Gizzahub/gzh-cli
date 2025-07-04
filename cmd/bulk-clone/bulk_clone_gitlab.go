package bulkclone

import (
	"fmt"

	bulkclonepkg "github.com/gizzahub/gzh-manager-go/pkg/bulk-clone"
	gitlabpkg "github.com/gizzahub/gzh-manager-go/pkg/gitlab"
	"github.com/spf13/cobra"
)

type bulkCloneGitlabOptions struct {
	targetPath  string
	groupName   string
	recursively bool
	strategy    string
	configFile  string
	useConfig   bool
}

func defaultBulkCloneGitlabOptions() *bulkCloneGitlabOptions {
	return &bulkCloneGitlabOptions{
		strategy: "reset",
	}
}

func newBulkCloneGitlabCmd() *cobra.Command {
	o := defaultBulkCloneGitlabOptions()

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

	// Mark flags as required only if not using config
	cmd.MarkFlagsMutuallyExclusive("config", "use-config")
	cmd.MarkFlagsOneRequired("targetPath", "config", "use-config")
	cmd.MarkFlagsOneRequired("groupName", "config", "use-config")

	return cmd
}

func (o *bulkCloneGitlabOptions) run(_ *cobra.Command, args []string) error {
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

	err := gitlabpkg.RefreshAll(o.targetPath, o.groupName, o.strategy)
	if err != nil {
		return err
	}

	return nil
}

func (o *bulkCloneGitlabOptions) loadFromConfig() error {
	var configPath string
	if o.configFile != "" {
		configPath = o.configFile
	}

	cfg, err := bulkclonepkg.LoadConfig(configPath)
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
		o.targetPath = bulkclonepkg.ExpandPath(groupConfig.RootPath)
	}

	if !o.recursively && groupConfig.Recursive {
		o.recursively = groupConfig.Recursive
	}

	return nil
}
