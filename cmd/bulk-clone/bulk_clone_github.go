package bulk_clone

import (
	"fmt"

	"github.com/gizzahub/gzh-manager-go/pkg/github"
	"github.com/spf13/cobra"
)

type bulkCloneGithubOptions struct {
	targetPath string
	orgName    string
	strategy   string
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

	cmd.MarkFlagRequired("targetPath")
	cmd.MarkFlagRequired("orgName")

	return cmd
}

func (o *bulkCloneGithubOptions) run(_ *cobra.Command, args []string) error {
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
