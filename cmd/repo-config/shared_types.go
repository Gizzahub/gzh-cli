// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package repoconfig

import (
	"github.com/spf13/cobra"
)

// Shared constants for change types
const (
	changeTypeCreate = "create"
	changeTypeUpdate = "update" 
	changeTypeDelete = "delete"
)

// GlobalFlags represents global flags for all repo-config commands.
type GlobalFlags struct {
	Organization string
	ConfigFile   string
	Token        string
	DryRun       bool
	Verbose      bool
	Parallel     int
	Timeout      string
}

// addGlobalFlags adds common flags to a command.
func addGlobalFlags(cmd *cobra.Command, flags *GlobalFlags) {
	cmd.Flags().StringVarP(&flags.Organization, "org", "o", "", "GitHub organization name")
	cmd.Flags().StringVarP(&flags.ConfigFile, "config", "c", "", "Configuration file path")
	cmd.Flags().StringVarP(&flags.Token, "token", "t", "", "GitHub personal access token")
	cmd.Flags().BoolVar(&flags.DryRun, "dry-run", false, "Preview changes without applying")
	cmd.Flags().BoolVarP(&flags.Verbose, "verbose", "v", false, "Verbose output")
	cmd.Flags().IntVar(&flags.Parallel, "parallel", 5, "Number of parallel operations")
	cmd.Flags().StringVar(&flags.Timeout, "timeout", "30s", "API timeout duration")
}

// getActionSymbol returns the symbol for action type.
func getActionSymbol(changeType string) string {
	switch changeType {
	case changeTypeCreate:
		return "➕"
	case changeTypeUpdate:
		return "🔄"
	case changeTypeDelete:
		return "➖"
	default:
		return "📝"
	}
}

// getActionSymbolWithText returns the symbol with text for action type.
func getActionSymbolWithText(changeType string) string {
	switch changeType {
	case changeTypeCreate:
		return "➕ Create"
	case changeTypeUpdate:
		return "🔄 Update"
	case changeTypeDelete:
		return "➖ Delete"
	default:
		return "❓ Unknown"
	}
}

// truncateString truncates a string to the specified length.
func truncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}

	if maxLength <= 3 {
		return s[:maxLength]
	}

	return s[:maxLength-3] + "..."
}

// Note: getStatusSymbol, formatTable, formatJSON are defined in individual command files