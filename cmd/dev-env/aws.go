// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package devenv

import (
	"github.com/spf13/cobra"
)

func newAwsCmd() *cobra.Command {
	baseCmd := NewBaseCommand(
		"aws",
		"config",
		".aws/config",
		`This command helps you backup and restore AWS configuration files, which contain:
- AWS profiles and regions
- Default output format settings
- Role and credential configurations
- SSO configurations
- Other AWS CLI settings

This is useful when:
- Setting up new development machines
- Switching between different AWS environments
- Backing up AWS configurations before changes
- Managing multiple AWS configurations for different projects`,
		[]string{
			"# Save current AWS config with a name",
			"gz dev-env aws save --name production",
			"",
			"# Save with description",
			"gz dev-env aws save --name staging --description \"Staging AWS config\"",
			"",
			"# Load a saved AWS config",
			"gz dev-env aws load --name production",
			"",
			"# List all saved configurations",
			"gz dev-env aws list",
			"",
			"# Save from specific path",
			"gz dev-env aws save --name custom --config-path /path/to/config",
		},
	)

	return baseCmd.CreateMainCommand()
}
