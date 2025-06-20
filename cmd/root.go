package cmd

import (
	"fmt"

	always_latest "github.com/gizzahub/gzh-manager-go/cmd/always-latest"
	bulk_clone "github.com/gizzahub/gzh-manager-go/cmd/bulk-clone"
	gen_config "github.com/gizzahub/gzh-manager-go/cmd/gen-config"
	ssh_config "github.com/gizzahub/gzh-manager-go/cmd/ssh-config"
	"github.com/spf13/cobra"
)

func newRootCmd(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gz",
		Short: "Cli 종합 Manager by Gizzahub",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(newVersionCmd(version))
	cmd.AddCommand(always_latest.NewAlwaysLatestCmd())
	cmd.AddCommand(bulk_clone.NewBulkCloneCmd())
	cmd.AddCommand(gen_config.NewGenConfigCmd())
	cmd.AddCommand(ssh_config.NewSSHConfigCmd())

	return cmd
}

// Execute invokes the command.
func Execute(version string) error {
	if err := newRootCmd(version).Execute(); err != nil {
		return fmt.Errorf("error executing root command: %w", err)
	}

	return nil
}
