package bulk_clone

import (
	"fmt"

	bulkclone "github.com/gizzahub/gzh-manager-go/pkg/bulk-clone"
	"github.com/gizzahub/gzh-manager-go/pkg/github"
	"github.com/gizzahub/gzh-manager-go/pkg/gitlab"
	"github.com/spf13/cobra"
)

type bulkCloneOptions struct {
	configFile string
	useConfig  bool
	strategy   string
}

func defaultBulkCloneOptions() *bulkCloneOptions {
	return &bulkCloneOptions{
		strategy: "reset",
	}
}

func NewBulkCloneCmd() *cobra.Command {
	o := defaultBulkCloneOptions()

	cmd := &cobra.Command{
		Use:          "bulk-clone",
		Short:        "Clone repositories from multiple Git hosting services",
		SilenceUsage: true,
		Long: `Clone multiple repositories from various Git hosting services.
		
You can use a configuration file (bulk-clone.yaml) to define multiple organizations
and their settings. This command will process all repository roots defined in the
configuration file regardless of the provider (GitHub, GitLab, Gitea, Gogs).

For provider-specific operations, use the subcommands (github, gitlab, etc.).`,
		RunE: o.run,
	}

	cmd.Flags().StringVarP(&o.configFile, "config", "c", o.configFile, "Path to config file")
	cmd.Flags().BoolVar(&o.useConfig, "use-config", false, "Use config file from standard locations")
	cmd.Flags().StringVarP(&o.strategy, "strategy", "s", o.strategy, "Sync strategy: reset, pull, or fetch")

	// Mark flags as mutually exclusive
	cmd.MarkFlagsMutuallyExclusive("config", "use-config")
	cmd.MarkFlagsOneRequired("config", "use-config")

	cmd.AddCommand(newBulkCloneGiteaCmd())
	cmd.AddCommand(newBulkCloneGithubCmd())
	cmd.AddCommand(newBulkCloneGitlabCmd())
	cmd.AddCommand(newBulkCloneGogsCmd())
	cmd.AddCommand(newBulkCloneValidateCmd())

	return cmd
}

func (o *bulkCloneOptions) run(_ *cobra.Command, args []string) error {
	// Load configuration
	var configPath string
	if o.configFile != "" {
		configPath = o.configFile
	}

	cfg, err := bulkclone.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Validate strategy
	if o.strategy != "reset" && o.strategy != "pull" && o.strategy != "fetch" {
		return fmt.Errorf("invalid strategy: %s. Must be one of: reset, pull, fetch", o.strategy)
	}

	// Process all repository roots
	for _, repoRoot := range cfg.RepoRoots {
		fmt.Printf("Processing %s organization: %s -> %s\n", repoRoot.Provider, repoRoot.OrgName, repoRoot.RootPath)

		// Expand the root path
		targetPath := bulkclone.ExpandPath(repoRoot.RootPath)

		switch repoRoot.Provider {
		case "github":
			err = github.RefreshAll(targetPath, repoRoot.OrgName, o.strategy)
			if err != nil {
				fmt.Printf("Error processing GitHub org %s: %v\n", repoRoot.OrgName, err)
				continue
			}
		case "gitlab":
			// For GitLab, we need to use the group name from the repoRoot
			// Since RepoRoots currently only has GitHub structure, we'll use OrgName as GroupName
			err = gitlab.RefreshAll(targetPath, repoRoot.OrgName, o.strategy)
			if err != nil {
				fmt.Printf("Error processing GitLab group %s: %v\n", repoRoot.OrgName, err)
				continue
			}
		default:
			fmt.Printf("Unsupported provider: %s (skipping %s)\n", repoRoot.Provider, repoRoot.OrgName)
			continue
		}

		fmt.Printf("✓ Successfully processed %s/%s\n", repoRoot.Provider, repoRoot.OrgName)
	}

	// Also process default GitHub and GitLab configurations if they have org/group names
	if cfg.Default.Github.OrgName != "" {
		fmt.Printf("Processing default GitHub organization: %s\n", cfg.Default.Github.OrgName)
		targetPath := bulkclone.ExpandPath(cfg.Default.Github.RootPath)
		err = github.RefreshAll(targetPath, cfg.Default.Github.OrgName, o.strategy)
		if err != nil {
			fmt.Printf("Error processing default GitHub org: %v\n", err)
		} else {
			fmt.Printf("✓ Successfully processed default GitHub org: %s\n", cfg.Default.Github.OrgName)
		}
	}

	if cfg.Default.Gitlab.GroupName != "" {
		fmt.Printf("Processing default GitLab group: %s\n", cfg.Default.Gitlab.GroupName)
		targetPath := bulkclone.ExpandPath(cfg.Default.Gitlab.RootPath)
		err = gitlab.RefreshAll(targetPath, cfg.Default.Gitlab.GroupName, o.strategy)
		if err != nil {
			fmt.Printf("Error processing default GitLab group: %v\n", err)
		} else {
			fmt.Printf("✓ Successfully processed default GitLab group: %s\n", cfg.Default.Gitlab.GroupName)
		}
	}

	fmt.Println("Bulk clone operation completed!")
	return nil
}
