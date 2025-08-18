// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package devenv

import (
	"github.com/spf13/cobra"
)

func newDockerCmd() *cobra.Command {
	baseCmd := NewBaseCommand(
		"docker",
		"json",
		".docker/config.json",
		`This command helps you backup and restore Docker configuration files, which contain:
- Docker registry authentication credentials
- Docker daemon configuration settings
- Registry mirrors and insecure registries
- Other Docker client settings

This is useful when:
- Setting up new development machines
- Switching between different Docker environments
- Backing up Docker credentials before changes
- Managing multiple Docker configurations for different projects`,
		[]string{
			"# Save current Docker config with a name",
			"gz dev-env docker save --name production",
			"",
			"# Save with description",
			"gz dev-env docker save --name staging --description \"Staging Docker config\"",
			"",
			"# Load a saved Docker config",
			"gz dev-env docker load --name production",
			"",
			"# List all saved configurations",
			"gz dev-env docker list",
			"",
			"# Save from specific path",
			"gz dev-env docker save --name custom --config-path /path/to/config.json",
		},
	)

	return baseCmd.CreateMainCommand()
}
