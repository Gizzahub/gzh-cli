// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package devenv

import (
	"github.com/spf13/cobra"
)

func newSshCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ssh",
		Short: "Manage SSH configuration files with includes and keys",
		Long: `Save and load SSH configuration files with advanced features:

ENHANCED FEATURES:
- Automatically parses and includes all files referenced by Include directives
- Backs up private keys referenced by IdentityFile directives  
- Backs up corresponding public keys (.pub files)
- Preserves directory structure and file relationships
- Maintains proper file permissions (600 for private keys, 644 for public keys)

WHAT GETS SAVED:
- Main SSH config file (~/.ssh/config)
- All files referenced by "Include" directives (e.g., config.d/*)
- All private keys referenced by "IdentityFile" directives
- Corresponding public keys (.pub files)
- Metadata with timestamps and descriptions

DIRECTORY STRUCTURE:
Configuration saved as: ~/.gz/ssh-configs/<name>/
├── config              # Main SSH config
├── includes/           # Files from Include directives  
├── keys/               # Private and public keys
└── metadata.json       # Metadata and file listings

This is useful when:
- Setting up new development machines
- Switching between different SSH environments  
- Backing up complete SSH setups before changes
- Managing multiple SSH configurations for different projects
- Ensuring all SSH dependencies are captured

Examples:
  # Save complete SSH setup with a name
  gz dev-env ssh save --name production
  
  # Save with description  
  gz dev-env ssh save --name staging --description "Staging environment SSH setup"
  
  # Save without private keys (config and includes only)
  gz dev-env ssh save --name minimal --include-keys=false
  
  # Load a saved SSH configuration
  gz dev-env ssh load --name production
  
  # List all saved configurations with details
  gz dev-env ssh list --all
  
  # Install public key to remote server
  gz dev-env ssh install-key --host server.com --user admin --public-key ~/.ssh/id_rsa.pub
  
  # Install all keys from saved configuration
  gz dev-env ssh install-key --config production --host server.com --user admin
  
  # List available keys in configurations
  gz dev-env ssh list-keys --config production`,
		SilenceUsage: true,
	}

	// Create enhanced SSH command instance
	enhancedCmd := NewEnhancedSSHCommand()
	
	// Add enhanced subcommands
	cmd.AddCommand(enhancedCmd.CreateEnhancedSaveCommand())
	cmd.AddCommand(enhancedCmd.CreateEnhancedLoadCommand())
	cmd.AddCommand(enhancedCmd.CreateEnhancedListCommand())
	
	// Add key management subcommands
	cmd.AddCommand(enhancedCmd.CreateInstallKeyCommand())
	cmd.AddCommand(enhancedCmd.CreateInstallKeySimpleCommand())
	cmd.AddCommand(enhancedCmd.CreateListKeysCommand())

	return cmd
}
