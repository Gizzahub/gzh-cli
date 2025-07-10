package cmd

import (
	"context"
	"fmt"

	alwayslatest "github.com/gizzahub/gzh-manager-go/cmd/always-latest"
	bulkclone "github.com/gizzahub/gzh-manager-go/cmd/bulk-clone"
	"github.com/gizzahub/gzh-manager-go/cmd/config"
	devenv "github.com/gizzahub/gzh-manager-go/cmd/dev-env"
	genconfig "github.com/gizzahub/gzh-manager-go/cmd/gen-config"
	"github.com/gizzahub/gzh-manager-go/cmd/ide"
	"github.com/gizzahub/gzh-manager-go/cmd/migrate"
	netenv "github.com/gizzahub/gzh-manager-go/cmd/net-env"
	repoconfig "github.com/gizzahub/gzh-manager-go/cmd/repo-config"
	sshconfig "github.com/gizzahub/gzh-manager-go/cmd/ssh-config"
	"github.com/spf13/cobra"
)

func newRootCmd(ctx context.Context, version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gz",
		Short: "Cli 종합 Manager by Gizzahub",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(newVersionCmd(version))
	cmd.AddCommand(alwayslatest.NewAlwaysLatestCmd(ctx))
	cmd.AddCommand(bulkclone.NewBulkCloneCmd(ctx))
	cmd.AddCommand(config.NewConfigCmd())
	cmd.AddCommand(devenv.NewDevEnvCmd())
	cmd.AddCommand(genconfig.NewGenConfigCmd(ctx))
	cmd.AddCommand(ide.NewIDECmd(ctx))
	cmd.AddCommand(migrate.NewMigrateCmd())
	cmd.AddCommand(netenv.NewNetEnvCmd(ctx))
	cmd.AddCommand(repoconfig.NewRepoConfigCmd())
	cmd.AddCommand(sshconfig.NewSSHConfigCmd())
	cmd.AddCommand(NewTaskRunnerCmd())
	cmd.AddCommand(NewWebhookCmd())

	return cmd
}

// Execute invokes the command.
func Execute(ctx context.Context, version string) error {
	if err := newRootCmd(ctx, version).Execute(); err != nil {
		return fmt.Errorf("error executing root command: %w", err)
	}

	return nil
}
