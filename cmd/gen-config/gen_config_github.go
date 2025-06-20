package gen_config

import (
	"fmt"

	"github.com/gizzahub/gzh-manager-go/pkg/github"
	"github.com/spf13/cobra"
)

type genConfigGithubOptions struct {
	targetPath string
	orgName    string
}

func defaultGenConfigGithubOptions() *genConfigGithubOptions {
	return &genConfigGithubOptions{}
}

func newGenConfigGithubCmd() *cobra.Command {
	o := defaultGenConfigGithubOptions()

	cmd := &cobra.Command{
		Use:   "github",
		Short: "Clone repositories from a GitHub organization (legacy)",
		Long: `Clone repositories from a GitHub organization.

This is a legacy command that directly performs Git operations. For generating
configuration files, use:
  gz gen-config init        # Interactive wizard
  gz gen-config template    # Predefined templates
  gz gen-config discover    # Auto-discover existing repos`,
		Args: cobra.NoArgs,
		RunE: o.run,
	}

	cmd.Flags().StringVarP(&o.targetPath, "targetPath", "t", o.targetPath, "targetPath")
	cmd.Flags().StringVarP(&o.orgName, "orgName", "o", o.orgName, "orgName")

	cmd.MarkFlagRequired("targetPath")
	cmd.MarkFlagRequired("orgName")

	return cmd
}

func (o *genConfigGithubOptions) run(_ *cobra.Command, args []string) error {
	if o.targetPath == "" || o.orgName == "" {
		return fmt.Errorf("both targetPath and orgName must be specified")
	}

	err := github.RefreshAll(o.targetPath, o.orgName, "reset")
	if err != nil {
		// return err
		// return fmt.Errorf("failed to refresh repositories: %w", err)
		return nil
	}

	return nil
}
