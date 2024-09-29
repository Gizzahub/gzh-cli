package cmd

import (
	"fmt"
	"github.com/gizzahub/gzh-manager-go/pkg/github_org"

	"github.com/spf13/cobra"
)

type githubRepoOptions struct {
	targetPath string
	orgName    string
}

func defaultGithubRepoOptions() *githubRepoOptions {
	return &githubRepoOptions{}
}

func newGithubRepoCmd() *cobra.Command {
	o := defaultGithubRepoOptions()

	cmd := &cobra.Command{
		Use:          "githubOrg",
		Short:        "githubOrg subcommand to manage GitHub repositories",
		SilenceUsage: true,
	}

	cloneCmd := &cobra.Command{
		Use:   "clone",
		Short: "Clone repositories from a GitHub organization",
		Args:  cobra.NoArgs,
		RunE:  o.runClone,
	}
	cloneCmd.Flags().StringVarP(&o.targetPath, "targetPath", "t", o.targetPath, "targetPath")
	cloneCmd.Flags().StringVarP(&o.orgName, "orgName", "o", o.orgName, "orgName")

	cmd.AddCommand(cloneCmd)

	return cmd
}

func (o *githubRepoOptions) runClone(cmd *cobra.Command, args []string) error {
	if o.targetPath == "" || o.orgName == "" {
		return fmt.Errorf("both targetPath and orgName must be specified")
	}

	err := github_org.RefreshAll(o.targetPath, o.orgName)
	if err != nil {
		return err
	}

	return nil
}
