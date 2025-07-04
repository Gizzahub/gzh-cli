package alwayslatest

import "github.com/spf13/cobra"

func NewAlwaysLatestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "always-latest",
		Short:        "Keep development tools and package managers up to date",
		SilenceUsage: true,
		Long: `Keep development tools and package managers updated to their latest versions.

This command provides automated updating for various development environment tools:
- asdf: Universal version manager for programming language runtimes
- brew: macOS package manager (Homebrew)
- sdkman: Software Development Kit Manager (Java ecosystem)
- port: MacPorts package manager for macOS
- apt: Debian/Ubuntu package manager (APT)
- rbenv: Ruby version manager

Supports two update strategies:
- minor: Update to latest patch/minor version (safer)
- major: Update to absolute latest version (includes breaking changes)

Examples:
  # Update asdf and its tools
  gz always-latest asdf
  
  # Update with major version strategy
  gz always-latest asdf --strategy major
  
  # Update all supported package managers
  gz always-latest --all`,
	}

	cmd.AddCommand(newAlwaysLatestAsdfCmd())
	cmd.AddCommand(newAlwaysLatestBrewCmd())
	cmd.AddCommand(newAlwaysLatestSdkmanCmd())
	cmd.AddCommand(newAlwaysLatestPortCmd())
	cmd.AddCommand(newAlwaysLatestAptCmd())
	cmd.AddCommand(newAlwaysLatestRbenvCmd())

	return cmd
}
