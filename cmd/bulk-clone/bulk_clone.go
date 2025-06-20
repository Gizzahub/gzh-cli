package bulk_clone

import "github.com/spf13/cobra"

func NewBulkCloneCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "bulk-clone",
		Short:        "bulk-clone subcommand to manage GitHub org repositories",
		SilenceUsage: true,
		Long: `Clone multiple repositories from various Git hosting services.
		
You can use a configuration file (bulk-clone.yaml) to define multiple organizations
and their settings, or use command-line flags for single organization operations.`,
	}

	cmd.AddCommand(newBulkCloneGiteaCmd())
	cmd.AddCommand(newBulkCloneGithubCmd())
	cmd.AddCommand(newBulkCloneGitlabCmd())
	cmd.AddCommand(newBulkCloneGogsCmd())

	return cmd
}
