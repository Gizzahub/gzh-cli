// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package synclone

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newConfigGenerateCmd creates the generate subcommand for config.
func newConfigGenerateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate configuration from existing repositories",
		Long: `Generate synclone configuration files from existing repositories.

This command provides various ways to create configuration files:
- Interactive wizard for step-by-step configuration creation
- Predefined templates for common use cases
- Auto-discovery from existing repository directories
- GitHub organization cloning (legacy functionality)

Examples:
  # Interactive configuration creation
  gz synclone config generate init

  # Generate from template
  gz synclone config generate template simple

  # Auto-discover from existing repositories
  gz synclone config generate discover ~/projects --recursive`,
		// RunE를 두지 않는다. 하위 명령만 있고 제 할 일이 없는 묶음 명령은
		// cobra가 알아서 도움말을 찍고 0으로 끝낸다. 어미 명령인 `config`와
		// `gz git`, `gz repo-config` 등이 전부 이 모양이다.
		//
		// 예전 RunE는 인자가 없으면 오류를 냈고(도움말 없이 종료 코드 1),
		// 인자가 있으면 nil을 돌려줘서 `gz synclone config generate foo`가
		// 아무것도 찍지 않고 성공했다. 이제는 도움말이 나온다.
	}

	// Add subcommands
	cmd.AddCommand(newConfigGenerateInitCmd())
	cmd.AddCommand(newConfigGenerateTemplateCmd())
	cmd.AddCommand(newConfigGenerateDiscoverCmd())
	cmd.AddCommand(newConfigGenerateGithubCmd())

	return cmd
}

func newConfigGenerateInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize configuration with interactive wizard",
		Long:  `Create a new synclone configuration file using an interactive wizard.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement interactive wizard
			return fmt.Errorf("config generate init: not yet implemented")
		},
	}
}

// newConfigGenerateTemplateCmd is implemented in config_generate_template.go

// newConfigGenerateDiscoverCmd is implemented in config_generate_discover.go

func newConfigGenerateGithubCmd() *cobra.Command {
	var (
		outputFile string
		token      string
		targetDir  string
	)

	cmd := &cobra.Command{
		Use:        "github [organization]",
		Short:      "Generate configuration from GitHub organization (legacy)",
		Long:       `Generate configuration by fetching repository list from a GitHub organization.`,
		Deprecated: "Use 'gz synclone github' directly for GitHub operations",
		Args:       cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement GitHub org scanning
			org := args[0]
			return fmt.Errorf("config generate github %s: not yet implemented", org)
		},
	}

	cmd.Flags().StringVarP(&outputFile, "output", "o", "synclone.yaml", "Output file path")
	cmd.Flags().StringVar(&token, "token", "", "GitHub personal access token")
	cmd.Flags().StringVar(&targetDir, "target-dir", ".", "Target directory for organization")

	return cmd
}
