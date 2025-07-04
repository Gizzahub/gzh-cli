package sshconfig

import "github.com/spf13/cobra"

func NewSSHConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "ssh-config",
		Short:        "SSH configuration management for Git operations",
		SilenceUsage: true,
		Long: `Manage SSH configuration for Git operations across multiple hosting services.

This command helps generate and manage SSH configurations for GitHub, GitLab,
Gitea services. It can generate ~/.ssh/config entries and manage
SSH keys for different organizations and services.`,
	}

	cmd.AddCommand(newSSHConfigGenerateCmd())
	cmd.AddCommand(newSSHConfigValidateCmd())

	return cmd
}
