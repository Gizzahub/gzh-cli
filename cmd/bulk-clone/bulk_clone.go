package bulk_clone

import "github.com/spf13/cobra"

func NewBulkCloneCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "bulk-clone",
		Short:        "bulk-clone subcommand to manage GitHub org repositories",
		SilenceUsage: true,
	}

	cmd.AddCommand(newBulkCloneGiteaCmd())
	cmd.AddCommand(newBulkCloneGithubCmd())
	cmd.AddCommand(newBulkCloneGitlabCmd())
	cmd.AddCommand(newBulkCloneGogsCmd())

	return cmd
}
