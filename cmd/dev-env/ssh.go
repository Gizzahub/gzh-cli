// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package devenv

import (
	"github.com/spf13/cobra"
)

func newSshCmd() *cobra.Command {
	baseCmd := NewBaseCommand(
		"ssh",
		"config",
		".ssh/config",
		`This command helps you backup and restore SSH configuration files, which contain:
- SSH host configurations and aliases
- SSH key paths and authentication settings
- Connection parameters and options
- Proxy and jump host configurations

This is useful when:
- Setting up new development machines
- Switching between different SSH environments
- Backing up SSH configurations before changes
- Managing multiple SSH configurations for different projects`,
		[]string{
			"# Save current SSH config with a name",
			"gz dev-env ssh save --name production",
			"",
			"# Save with description",
			"gz dev-env ssh save --name staging --description \"Staging SSH config\"",
			"",
			"# Load a saved SSH config",
			"gz dev-env ssh load --name production",
			"",
			"# List all saved configurations",
			"gz dev-env ssh list",
			"",
			"# Save from specific path",
			"gz dev-env ssh save --name custom --config-path /path/to/ssh/config",
		},
	)

	return baseCmd.CreateMainCommand()
}
