// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package repo

import (
	"github.com/spf13/cobra"
)

// NewCmd creates the repository lifecycle root command.
// 기존 'git.NewGitRepoCmd' 대응 어댑터: 상위에서 래핑하여 호출한다
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repo",
		Short: "Repository lifecycle management",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	// 하위 커맨드는 기존 파일들에서 등록한다
	cmd.AddCommand(newRepoCloneCmd())
	cmd.AddCommand(newRepoCloneOrUpdateCmd())
	cmd.AddCommand(newRepoListCmd())
	cmd.AddCommand(newRepoCreateCmd())
	cmd.AddCommand(newRepoDeleteCmd())
	cmd.AddCommand(newRepoArchiveCmd())
	cmd.AddCommand(newRepoSyncCmd())
	cmd.AddCommand(newRepoMigrateCmd())
	cmd.AddCommand(newRepoSearchCmd())

	return cmd
}
