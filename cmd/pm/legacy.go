// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package pm

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// Legacy package manager specific commands
// These provide compatibility with the old always-latest command structure

func newAsdfCmd(ctx context.Context) *cobra.Command {
	return &cobra.Command{
		Use:     "asdf",
		Aliases: []string{"update-asdf"},
		Short:   "Manage asdf version manager",
		Long: `Manage asdf plugins and tool versions.

This is a compatibility command for the legacy 'gz always-latest asdf' command.
For new features, use the unified commands like 'gz pm update --manager asdf'.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Use 'gz pm update --manager asdf' for the new unified interface")
			return updateAsdf(ctx, "stable", false)
		},
	}
}

func newBrewCmd(ctx context.Context) *cobra.Command {
	return &cobra.Command{
		Use:     "brew",
		Aliases: []string{"update-brew"},
		Short:   "Manage Homebrew packages",
		Long: `Manage Homebrew formulae and casks.

This is a compatibility command for the legacy 'gz always-latest brew' command.
For new features, use the unified commands like 'gz pm update --manager brew'.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Use 'gz pm update --manager brew' for the new unified interface")
			return updateBrew(ctx, "stable", false)
		},
	}
}

func newSdkmanCmd(ctx context.Context) *cobra.Command {
	return &cobra.Command{
		Use:     "sdkman",
		Aliases: []string{"update-sdkman"},
		Short:   "Manage SDKMAN candidates",
		Long: `Manage SDKMAN candidates for JVM ecosystem.

This is a compatibility command for the legacy 'gz always-latest sdkman' command.
For new features, use the unified commands like 'gz pm update --manager sdkman'.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Use 'gz pm update --manager sdkman' for the new unified interface")
			return updateSdkman(ctx, "stable", false)
		},
	}
}

func newAptCmd(ctx context.Context) *cobra.Command {
	return &cobra.Command{
		Use:     "apt",
		Aliases: []string{"update-apt"},
		Short:   "Manage APT packages",
		Long: `Manage APT packages on Debian/Ubuntu systems.

This is a compatibility command for the legacy 'gz always-latest apt' command.
For new features, use the unified commands like 'gz pm update --manager apt'.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Use 'gz pm update --manager apt' for the new unified interface")
			return updateApt(ctx, "stable", false)
		},
	}
}

func newPortCmd(ctx context.Context) *cobra.Command {
	return &cobra.Command{
		Use:     "port",
		Aliases: []string{"update-port"},
		Short:   "Manage MacPorts packages",
		Long: `Manage MacPorts packages on macOS.

This is a compatibility command for the legacy 'gz always-latest port' command.
For new features, use the unified commands like 'gz pm update --manager port'.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Use 'gz pm update --manager port' for the new unified interface")
			return fmt.Errorf("port update not yet implemented")
		},
	}
}

func newRbenvCmd(ctx context.Context) *cobra.Command {
	return &cobra.Command{
		Use:     "rbenv",
		Aliases: []string{"update-rbenv"},
		Short:   "Manage rbenv Ruby versions",
		Long: `Manage rbenv and Ruby versions.

This is a compatibility command for the legacy 'gz always-latest rbenv' command.
For new features, use the unified commands like 'gz pm update --manager rbenv'.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Use 'gz pm update --manager rbenv' for the new unified interface")
			return fmt.Errorf("rbenv update not yet implemented")
		},
	}
}
