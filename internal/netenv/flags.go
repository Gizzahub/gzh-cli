// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package netenv

import "github.com/spf13/pflag"

// AddCommonFlags adds common flags to a flag set that are shared across net-env commands.
func AddCommonFlags(flags *pflag.FlagSet, opts *CommonOptions) {
	flags.StringVarP(&opts.ConfigFile, "config", "c", "", "Configuration file path")
	flags.BoolVarP(&opts.Verbose, "verbose", "v", false, "Enable verbose output")
	flags.BoolVar(&opts.DryRun, "dry-run", false, "Show what would be done without making changes")
}

// AddConfigFlags adds configuration-related flags.
func AddConfigFlags(flags *pflag.FlagSet, configDir *string) {
	flags.StringVar(configDir, "config-dir", GetConfigDirectory(), "Configuration directory")
}

// AddOutputFlags adds output formatting flags.
func AddOutputFlags(flags *pflag.FlagSet, format *string, output *string) {
	flags.StringVarP(format, "format", "f", "table", "Output format (table, json, yaml)")
	flags.StringVarP(output, "output", "o", "", "Output file (default: stdout)")
}
