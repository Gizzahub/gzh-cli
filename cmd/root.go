package cmd

import (
	"context"
	"fmt"

	alwayslatest "github.com/gizzahub/gzh-manager-go/cmd/always-latest"
	bulkclone "github.com/gizzahub/gzh-manager-go/cmd/bulk-clone"
	"github.com/gizzahub/gzh-manager-go/cmd/config"
	debugcmd "github.com/gizzahub/gzh-manager-go/cmd/debug"
	devenv "github.com/gizzahub/gzh-manager-go/cmd/dev-env"
	"github.com/gizzahub/gzh-manager-go/cmd/docker"
	doctorcmd "github.com/gizzahub/gzh-manager-go/cmd/doctor"
	genconfig "github.com/gizzahub/gzh-manager-go/cmd/gen-config"
	githubactions "github.com/gizzahub/gzh-manager-go/cmd/github-actions"
	gitlabci "github.com/gizzahub/gzh-manager-go/cmd/gitlab-ci"
	"github.com/gizzahub/gzh-manager-go/cmd/helm"
	"github.com/gizzahub/gzh-manager-go/cmd/i18n"
	"github.com/gizzahub/gzh-manager-go/cmd/ide"
	"github.com/gizzahub/gzh-manager-go/cmd/migrate"
	"github.com/gizzahub/gzh-manager-go/cmd/monitoring"
	netenv "github.com/gizzahub/gzh-manager-go/cmd/net-env"
	"github.com/gizzahub/gzh-manager-go/cmd/operator"
	"github.com/gizzahub/gzh-manager-go/cmd/plugin"
	repoconfig "github.com/gizzahub/gzh-manager-go/cmd/repo-config"
	reposync "github.com/gizzahub/gzh-manager-go/cmd/repo-sync"
	"github.com/gizzahub/gzh-manager-go/cmd/serve"
	"github.com/gizzahub/gzh-manager-go/cmd/shell"
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
	cmd.AddCommand(debugcmd.DebugCmd)
	cmd.AddCommand(doctorcmd.DoctorCmd)
	cmd.AddCommand(devenv.NewDevEnvCmd())
	cmd.AddCommand(docker.DockerCmd)
	cmd.AddCommand(genconfig.NewGenConfigCmd(ctx))
	cmd.AddCommand(githubactions.GitHubActionsCmd)
	cmd.AddCommand(gitlabci.GitLabCICmd)
	cmd.AddCommand(helm.HelmCmd)
	cmd.AddCommand(i18n.I18nCmd)
	cmd.AddCommand(ide.NewIDECmd(ctx))
	cmd.AddCommand(migrate.NewMigrateCmd())
	cmd.AddCommand(monitoring.NewMonitoringCmd(ctx))
	cmd.AddCommand(netenv.NewNetEnvCmd(ctx))
	cmd.AddCommand(operator.OperatorCmd)
	cmd.AddCommand(plugin.PluginCmd)
	cmd.AddCommand(repoconfig.NewRepoConfigCmd())
	cmd.AddCommand(reposync.NewRepoSyncCmd(ctx))
	cmd.AddCommand(serve.ServeCmd)
	cmd.AddCommand(shell.ShellCmd)
	cmd.AddCommand(sshconfig.NewSSHConfigCmd())
	cmd.AddCommand(NewTaskRunnerCmd())
	cmd.AddCommand(NewWebhookCmd())
	cmd.AddCommand(NewEventCmd())

	return cmd
}

// Execute invokes the command.
func Execute(ctx context.Context, version string) error {
	if err := newRootCmd(ctx, version).Execute(); err != nil {
		return fmt.Errorf("error executing root command: %w", err)
	}

	return nil
}
