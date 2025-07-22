// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	alwayslatest "github.com/gizzahub/gzh-manager-go/cmd/always-latest"
	bulkclone "github.com/gizzahub/gzh-manager-go/cmd/bulk-clone"
	"github.com/gizzahub/gzh-manager-go/cmd/config"
	devenv "github.com/gizzahub/gzh-manager-go/cmd/dev-env"
	"github.com/gizzahub/gzh-manager-go/cmd/docker"
	doctorcmd "github.com/gizzahub/gzh-manager-go/cmd/doctor"
	genconfig "github.com/gizzahub/gzh-manager-go/cmd/gen-config"
	"github.com/gizzahub/gzh-manager-go/cmd/ide"
	"github.com/gizzahub/gzh-manager-go/cmd/migrate"
	netenv "github.com/gizzahub/gzh-manager-go/cmd/net-env"
	repoconfig "github.com/gizzahub/gzh-manager-go/cmd/repo-config"
	reposync "github.com/gizzahub/gzh-manager-go/cmd/repo-sync"
	"github.com/gizzahub/gzh-manager-go/cmd/shell"
	sshconfig "github.com/gizzahub/gzh-manager-go/cmd/ssh-config"
)

func newRootCmd(ctx context.Context, version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gz",
		Short: "Cli 종합 Manager by Gizzahub",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(newVersionCmd(version))
	cmd.AddCommand(alwayslatest.NewAlwaysLatestCmd(ctx))
	cmd.AddCommand(bulkclone.NewBulkCloneCmd(ctx))
	cmd.AddCommand(config.NewConfigCmd()) //nolint:contextcheck // Command setup doesn't require context propagation
	cmd.AddCommand(doctorcmd.DoctorCmd)
	cmd.AddCommand(devenv.NewDevEnvCmd()) //nolint:contextcheck // Command setup doesn't require context propagation
	cmd.AddCommand(docker.DockerCmd)
	cmd.AddCommand(genconfig.NewGenConfigCmd(ctx))
	cmd.AddCommand(ide.NewIDECmd(ctx))
	cmd.AddCommand(migrate.NewMigrateCmd()) //nolint:contextcheck // Command setup doesn't require context propagation
	cmd.AddCommand(netenv.NewNetEnvCmd(ctx))
	cmd.AddCommand(repoconfig.NewRepoConfigCmd()) //nolint:contextcheck // Command setup doesn't require context propagation
	cmd.AddCommand(reposync.NewRepoSyncCmd(ctx))
	cmd.AddCommand(shell.ShellCmd)
	cmd.AddCommand(sshconfig.NewSSHConfigCmd())
	cmd.AddCommand(NewTaskRunnerCmd())
	cmd.AddCommand(NewWebhookCmd())
	cmd.AddCommand(NewEventCmd()) //nolint:contextcheck // Command setup doesn't require context propagation

	return cmd
}

// Execute invokes the command.
func Execute(ctx context.Context, version string) error {
	if err := newRootCmd(ctx, version).Execute(); err != nil {
		return fmt.Errorf("error executing root command: %w", err)
	}

	return nil
}
