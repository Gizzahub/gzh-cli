package bulkclone

import (
	"fmt"

	bulkclonepkg "github.com/gizzahub/gzh-manager-go/pkg/bulk-clone"
	"github.com/gizzahub/gzh-manager-go/pkg/github"
	"github.com/spf13/cobra"
)

type bulkCloneGithubOptions struct {
	targetPath string
	orgName    string
	strategy   string
	configFile string
	useConfig  bool
}

func defaultBulkCloneGithubOptions() *bulkCloneGithubOptions {
	return &bulkCloneGithubOptions{
		strategy: "reset",
	}
}

func newBulkCloneGithubCmd() *cobra.Command {
	o := defaultBulkCloneGithubOptions()

	cmd := &cobra.Command{
		Use:   "github",
		Short: "Clone repositories from a GitHub organization",
		Args:  cobra.NoArgs,
		RunE:  o.run,
	}

	cmd.Flags().StringVarP(&o.targetPath, "targetPath", "t", o.targetPath, "targetPath")
	cmd.Flags().StringVarP(&o.orgName, "orgName", "o", o.orgName, "orgName")
	cmd.Flags().StringVarP(&o.strategy, "strategy", "s", o.strategy, "Sync strategy: reset, pull, or fetch")
	cmd.Flags().StringVarP(&o.configFile, "config", "c", o.configFile, "Path to config file")
	cmd.Flags().BoolVar(&o.useConfig, "use-config", false, "Use config file from standard locations")

	// Mark flags as required only if not using config
	cmd.MarkFlagsMutuallyExclusive("config", "use-config")
	cmd.MarkFlagsOneRequired("targetPath", "config", "use-config")
	cmd.MarkFlagsOneRequired("orgName", "config", "use-config")

	return cmd
}

func (o *bulkCloneGithubOptions) run(_ *cobra.Command, args []string) error {
	// Load config if specified
	if o.configFile != "" || o.useConfig {
		err := o.loadFromConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
	}

	if o.targetPath == "" || o.orgName == "" {
		return fmt.Errorf("both targetPath and orgName must be specified")
	}

	// Validate strategy
	if o.strategy != "reset" && o.strategy != "pull" && o.strategy != "fetch" {
		return fmt.Errorf("invalid strategy: %s. Must be one of: reset, pull, fetch", o.strategy)
	}

	err := github.RefreshAll(o.targetPath, o.orgName, o.strategy)
	if err != nil {
		// return err
		// return fmt.Errorf("failed to refresh repositories: %w", err)
		return nil
	}

	return nil
}

func (o *bulkCloneGithubOptions) loadFromConfig() error {
	var configPath string
	if o.configFile != "" {
		configPath = o.configFile
	}

	cfg, err := bulkclonepkg.LoadConfig(configPath)
	if err != nil {
		return err
	}

	// If orgName is specified via CLI, use it; otherwise get from config
	if o.orgName == "" {
		// Handle multiple organizations from config
		// First try to get from RepoRoots, then fall back to defaults
		if len(cfg.RepoRoots) > 0 {
			// Use the first organization from repo roots
			o.orgName = cfg.RepoRoots[0].OrgName
		} else if cfg.Default.Github.OrgName != "" {
			o.orgName = cfg.Default.Github.OrgName
		} else {
			return fmt.Errorf("no organization found in config")
		}
	}

	// Get config for the specific organization
	orgConfig, err := cfg.GetGithubOrgConfig(o.orgName)
	if err != nil {
		return err
	}

	// Apply config values (CLI flags take precedence)
	if o.targetPath == "" && orgConfig.RootPath != "" {
		o.targetPath = bulkclonepkg.ExpandPath(orgConfig.RootPath)
	}

	// Apply protocol from config if available
	if orgConfig.Protocol != "" {
		// Protocol support would be implemented here
		// For now, this is documented for future implementation
	}

	// Apply auth configuration if available
	// Authentication support would be implemented here using
	// environment variables or config-based token management
	// For now, this is documented for future implementation

	return nil
}
