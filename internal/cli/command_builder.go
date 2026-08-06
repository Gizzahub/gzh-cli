// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

// Package cli provides common CLI patterns and utilities for building consistent
// command-line interfaces across the gzh-cli application.
package cli

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/gizzahub/gzh-cli/internal/env"
)

// CommonFlags represents flags that are commonly used across commands.
type CommonFlags struct {
	Organization string
	Token        string
	ConfigFile   string
	Verbose      bool
	DryRun       bool
	Format       string
	Filter       string
	Limit        int
}

// CommandBuilder provides a fluent interface for building cobra commands with common patterns.
type CommandBuilder struct {
	cmd   *cobra.Command
	flags *CommonFlags
}

// NewCommandBuilder creates a new command builder.
//
// ctx를 받지 않는다. 예전에는 받아서 WithRunFuncE가 그것을 넘겼는데,
// 부르는 아홉 곳이 전부 context.Background()를 주고 있었다. 명령을 "만드는"
// 시점의 맥락은 명령을 "실행하는" 시점의 맥락이 아니다 -- 후자는 cobra가
// RunE에 직접 준다.
func NewCommandBuilder(use, short string) *CommandBuilder {
	return &CommandBuilder{
		cmd: &cobra.Command{
			Use:   use,
			Short: short,
		},
		flags: &CommonFlags{},
	}
}

// WithLongDescription adds a long description to the command.
func (b *CommandBuilder) WithLongDescription(long string) *CommandBuilder {
	b.cmd.Long = long
	return b
}

// WithExample adds an example to the command.
func (b *CommandBuilder) WithExample(example string) *CommandBuilder {
	b.cmd.Example = example
	return b
}

// WithOrganizationFlag adds the organization flag.
func (b *CommandBuilder) WithOrganizationFlag(required bool) *CommandBuilder {
	b.cmd.Flags().StringVar(&b.flags.Organization, "org", "", "Organization name")
	if required {
		b.cmd.MarkFlagRequired("org")
	}
	return b
}

// WithTokenFlag adds the token flag.
func (b *CommandBuilder) WithTokenFlag() *CommandBuilder {
	b.cmd.Flags().StringVar(&b.flags.Token, "token", "", "Authentication token (overrides environment)")
	return b
}

// WithConfigFileFlag adds the config file flag.
func (b *CommandBuilder) WithConfigFileFlag() *CommandBuilder {
	b.cmd.Flags().StringVar(&b.flags.ConfigFile, "config", "", "Configuration file path")
	return b
}

// WithVerboseFlag adds the verbose flag.
func (b *CommandBuilder) WithVerboseFlag() *CommandBuilder {
	b.cmd.Flags().BoolVar(&b.flags.Verbose, "verbose", false, "Enable verbose output")
	return b
}

// WithDryRunFlag adds the dry-run flag.
func (b *CommandBuilder) WithDryRunFlag() *CommandBuilder {
	b.cmd.Flags().BoolVar(&b.flags.DryRun, "dry-run", false, "Show what would be done without making changes")
	return b
}

// WithFormatFlag adds the format flag with specified options.
func (b *CommandBuilder) WithFormatFlag(defaultFormat string, validFormats []string) *CommandBuilder {
	validFormatsStr := fmt.Sprintf("(%s)", joinStrings(validFormats, ", "))
	help := fmt.Sprintf("Output format %s", validFormatsStr)
	b.cmd.Flags().StringVar(&b.flags.Format, "format", defaultFormat, help)
	return b
}

// WithFilterFlag adds the filter flag for pattern matching.
func (b *CommandBuilder) WithFilterFlag() *CommandBuilder {
	b.cmd.Flags().StringVar(&b.flags.Filter, "filter", "", "Filter results by name pattern (regex)")
	return b
}

// WithLimitFlag adds the limit flag for result pagination.
func (b *CommandBuilder) WithLimitFlag(defaultLimit int) *CommandBuilder {
	help := "Limit number of results (0 = no limit)"
	b.cmd.Flags().IntVar(&b.flags.Limit, "limit", defaultLimit, help)
	return b
}

// WithCustomFlag adds a custom flag to the command.
func (b *CommandBuilder) WithCustomFlag(name, defaultValue, usage string, target *string) *CommandBuilder {
	b.cmd.Flags().StringVar(target, name, defaultValue, usage)
	return b
}

// WithCustomBoolFlag adds a custom boolean flag to the command.
func (b *CommandBuilder) WithCustomBoolFlag(name string, defaultValue bool, usage string, target *bool) *CommandBuilder {
	b.cmd.Flags().BoolVar(target, name, defaultValue, usage)
	return b
}

// WithCustomIntFlag adds a custom integer flag to the command.
func (b *CommandBuilder) WithCustomIntFlag(name string, defaultValue int, usage string, target *int) *CommandBuilder {
	b.cmd.Flags().IntVar(target, name, defaultValue, usage)
	return b
}

// WithRunFunc sets the run function for the command.
func (b *CommandBuilder) WithRunFunc(runFunc func(cmd *cobra.Command, args []string) error) *CommandBuilder {
	b.cmd.RunE = runFunc
	return b
}

// WithRunFuncE sets the run function that uses the execution context and flags.
func (b *CommandBuilder) WithRunFuncE(runFunc func(ctx context.Context, flags *CommonFlags, args []string) error) *CommandBuilder {
	b.cmd.RunE = func(cmd *cobra.Command, args []string) error {
		// cobra가 넣어 준 맥락을 그대로 넘긴다. 예전에는 이 자리에서 cmd를
		// _로 버리고 빌더가 지니고 있던 맥락을 넘겼다. 좋은 맥락이 바로
		// 손에 들어온 자리에서 죽은 맥락으로 바꿔치기한 셈이다. 그래서
		// `gz doctor health --continuous`처럼 ctx.Done()을 기다리는 명령이
		// Ctrl+C에 반응하지 않았다.
		ctx := cmd.Context()
		if ctx == nil {
			// Execute를 거치지 않고 RunE를 직접 부른 경우다(주로 시험).
			// cobra의 ExecuteC가 같은 상황에서 넣는 값을 쓴다.
			ctx = context.Background()
		}

		return runFunc(ctx, b.flags, args)
	}

	return b
}

// AddSubcommand adds a subcommand to this command.
func (b *CommandBuilder) AddSubcommand(subCmd *cobra.Command) *CommandBuilder {
	b.cmd.AddCommand(subCmd)
	return b
}

// Build returns the built cobra command.
func (b *CommandBuilder) Build() *cobra.Command {
	return b.cmd
}

// GetFlags returns the common flags.
func (b *CommandBuilder) GetFlags() *CommonFlags {
	return b.flags
}

// CommandValidator provides validation utilities for commands.
type CommandValidator struct {
	environment env.Environment
}

// NewCommandValidator creates a new command validator.
func NewCommandValidator() *CommandValidator {
	return &CommandValidator{
		environment: env.NewOSEnvironment(),
	}
}

// ValidateOrganization validates that an organization is provided.
func (v *CommandValidator) ValidateOrganization(org string) error {
	if org == "" {
		return fmt.Errorf("organization is required (use --org flag)")
	}
	return nil
}

// ValidateFormat validates the output format against allowed formats.
func (v *CommandValidator) ValidateFormat(format string, validFormats []string) error {
	if slices.Contains(validFormats, format) {
		return nil
	}
	return fmt.Errorf("invalid format '%s', valid formats: %s", format, joinStrings(validFormats, ", "))
}

// ValidateFilter validates the filter regex pattern.
func (v *CommandValidator) ValidateFilter(filter string) error {
	if filter == "" {
		return nil // Empty filter is valid
	}
	_, err := regexp.Compile(filter)
	if err != nil {
		return fmt.Errorf("invalid filter pattern: %w", err)
	}
	return nil
}

// ValidateLimit validates the limit parameter.
func (v *CommandValidator) ValidateLimit(limit int) error {
	if limit < 0 {
		return fmt.Errorf("limit must be non-negative")
	}
	return nil
}

// GetToken gets the authentication token from flags or environment.
func (v *CommandValidator) GetToken(flagToken, envKey string) string {
	if flagToken != "" {
		return flagToken
	}
	return v.environment.Get(envKey)
}

// joinStrings joins strings with a separator.
func joinStrings(strs []string, separator string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}

	var result strings.Builder
	result.WriteString(strs[0])
	for i := 1; i < len(strs); i++ {
		result.WriteString(separator + strs[i])
	}
	return result.String()
}
