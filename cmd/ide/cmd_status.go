// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package ide

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

type statusOptions struct {
	output  string
	verbose bool
}

// newIDEStatusCmd creates the IDE status subcommand
func newIDEStatusCmd() *cobra.Command {
	o := &statusOptions{}

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show status of installed IDEs",
		Long: `Show detailed status information for all detected IDEs.

This command displays:
- List of installed IDEs with versions
- Last update time for each IDE
- Installation paths and executable locations
- Available aliases for opening IDEs

Supports multiple output formats for integration with other tools.

Examples:
  # Show IDE status in table format
  gz ide status

  # Show verbose information
  gz ide status --verbose

  # Output as JSON for scripting
  gz ide status --output json

  # Output as YAML
  gz ide status --output yaml`,
		RunE: o.runStatus,
	}

	cmd.Flags().StringVarP(&o.output, "output", "o", "table", "Output format (table, json, yaml)")
	cmd.Flags().BoolVar(&o.verbose, "verbose", false, "Show detailed information")

	return cmd
}

func (o *statusOptions) runStatus(cmd *cobra.Command, args []string) error {
	detector := NewIDEDetector()

	// Get IDEs (prefer cache for status command)
	ides, err := detector.DetectIDEs(true)
	if err != nil {
		return fmt.Errorf("failed to detect IDEs: %w", err)
	}

	if len(ides) == 0 {
		fmt.Printf("No IDEs found. Run 'gz ide scan' to detect IDEs.\n")
		return nil
	}

	switch o.output {
	case "json":
		return o.outputJSON(ides)
	case "yaml":
		return o.outputYAML(ides)
	case "table":
		fallthrough
	default:
		return o.outputTable(ides)
	}
}

func (o *statusOptions) outputTable(ides []IDE) error {
	fmt.Printf("🖥️  IDE Status (%d found):\n\n", len(ides))

	// Calculate column widths
	maxNameWidth := 12   // "IDE"
	maxVersionWidth := 8 // "Version"
	maxPathWidth := 20   // "Path"

	for _, ide := range ides {
		if len(ide.Name) > maxNameWidth {
			maxNameWidth = len(ide.Name)
		}
		if len(ide.Version) > maxVersionWidth {
			maxVersionWidth = len(ide.Version)
		}

		path := o.formatPath(ide.Executable)
		if len(path) > maxPathWidth {
			maxPathWidth = len(path)
		}
	}

	// Limit column widths
	if maxNameWidth > 25 {
		maxNameWidth = 25
	}
	if maxVersionWidth > 15 {
		maxVersionWidth = 15
	}
	if maxPathWidth > 50 {
		maxPathWidth = 50
	}

	// Print header
	headerFormat := fmt.Sprintf("┌─%%-%ds─┬─%%-%ds─┬─%%-%ds─┬─%%s─┐\n", maxNameWidth, maxVersionWidth, maxPathWidth)
	fmt.Printf(headerFormat,
		strings.Repeat("─", maxNameWidth),
		strings.Repeat("─", maxVersionWidth),
		strings.Repeat("─", maxPathWidth),
		strings.Repeat("─", 15))

	rowFormat := fmt.Sprintf("│ %%-%ds │ %%-%ds │ %%-%ds │ %%s │\n", maxNameWidth, maxVersionWidth, maxPathWidth)
	fmt.Printf(rowFormat, "IDE", "Version", "Path", "Last Updated")

	separatorFormat := fmt.Sprintf("├─%%-%ds─┼─%%-%ds─┼─%%-%ds─┼─%%s─┤\n", maxNameWidth, maxVersionWidth, maxPathWidth)
	fmt.Printf(separatorFormat,
		strings.Repeat("─", maxNameWidth),
		strings.Repeat("─", maxVersionWidth),
		strings.Repeat("─", maxPathWidth),
		strings.Repeat("─", 15))

	// Print rows
	for _, ide := range ides {
		name := o.truncateString(ide.Name, maxNameWidth)
		version := o.truncateString(ide.Version, maxVersionWidth)
		path := o.truncateString(o.formatPath(ide.Executable), maxPathWidth)
		lastUpdated := o.formatLastUpdated(ide.LastUpdated)

		fmt.Printf(rowFormat, name, version, path, lastUpdated)
	}

	// Print footer
	footerFormat := fmt.Sprintf("└─%%-%ds─┴─%%-%ds─┴─%%-%ds─┴─%%s─┘\n", maxNameWidth, maxVersionWidth, maxPathWidth)
	fmt.Printf(footerFormat,
		strings.Repeat("─", maxNameWidth),
		strings.Repeat("─", maxVersionWidth),
		strings.Repeat("─", maxPathWidth),
		strings.Repeat("─", 15))

	// Show verbose information
	if o.verbose {
		fmt.Printf("\nDetailed Information:\n\n")
		for _, ide := range ides {
			fmt.Printf("🔹 %s\n", ide.Name)
			fmt.Printf("   Version: %s\n", ide.Version)
			fmt.Printf("   Type: %s\n", ide.Type)
			fmt.Printf("   Install Method: %s\n", o.formatInstallMethod(ide.InstallMethod, ide.InstallPath))
			fmt.Printf("   Executable: %s\n", ide.Executable)
			fmt.Printf("   Last Updated: %s\n", o.formatDetailedTime(ide.LastUpdated))
			if len(ide.Aliases) > 0 {
				fmt.Printf("   Aliases: %s\n", strings.Join(ide.Aliases, ", "))
			}
			fmt.Println()
		}
	}

	// Show usage hints
	fmt.Printf("\nUsage:\n")
	fmt.Printf("  gz ide open <name>     # Open IDE in current directory\n")
	fmt.Printf("  gz ide open <name> .   # Open IDE in current directory\n")
	fmt.Printf("  gz ide open <name> dir # Open IDE in specified directory\n")

	return nil
}

func (o *statusOptions) outputJSON(ides []IDE) error {
	// Simple JSON output without external dependencies
	fmt.Printf("{\n")
	fmt.Printf("  \"ides\": [\n")

	for i, ide := range ides {
		fmt.Printf("    {\n")
		fmt.Printf("      \"name\": \"%s\",\n", o.escapeJSON(ide.Name))
		fmt.Printf("      \"executable\": \"%s\",\n", o.escapeJSON(ide.Executable))
		fmt.Printf("      \"version\": \"%s\",\n", o.escapeJSON(ide.Version))
		fmt.Printf("      \"type\": \"%s\",\n", o.escapeJSON(ide.Type))
		fmt.Printf("      \"install_method\": \"%s\",\n", o.escapeJSON(ide.InstallMethod))
		fmt.Printf("      \"install_path\": \"%s\",\n", o.escapeJSON(ide.InstallPath))
		fmt.Printf("      \"last_updated\": \"%s\",\n", ide.LastUpdated.Format(time.RFC3339))
		fmt.Printf("      \"aliases\": [")

		for j, alias := range ide.Aliases {
			fmt.Printf("\"%s\"", o.escapeJSON(alias))
			if j < len(ide.Aliases)-1 {
				fmt.Printf(", ")
			}
		}
		fmt.Printf("]\n")
		fmt.Printf("    }")

		if i < len(ides)-1 {
			fmt.Printf(",")
		}
		fmt.Printf("\n")
	}

	fmt.Printf("  ],\n")
	fmt.Printf("  \"count\": %d\n", len(ides))
	fmt.Printf("}\n")

	return nil
}

func (o *statusOptions) outputYAML(ides []IDE) error {
	// Simple YAML output without external dependencies
	fmt.Printf("ides:\n")

	for _, ide := range ides {
		fmt.Printf("  - name: \"%s\"\n", o.escapeYAML(ide.Name))
		fmt.Printf("    executable: \"%s\"\n", o.escapeYAML(ide.Executable))
		fmt.Printf("    version: \"%s\"\n", o.escapeYAML(ide.Version))
		fmt.Printf("    type: \"%s\"\n", o.escapeYAML(ide.Type))
		fmt.Printf("    install_method: \"%s\"\n", o.escapeYAML(ide.InstallMethod))
		fmt.Printf("    install_path: \"%s\"\n", o.escapeYAML(ide.InstallPath))
		fmt.Printf("    last_updated: \"%s\"\n", ide.LastUpdated.Format(time.RFC3339))
		fmt.Printf("    aliases:\n")

		for _, alias := range ide.Aliases {
			fmt.Printf("      - \"%s\"\n", o.escapeYAML(alias))
		}
	}

	fmt.Printf("count: %d\n", len(ides))

	return nil
}

func (o *statusOptions) formatPath(path string) string {
	// Show path relative to home if possible
	if strings.HasPrefix(path, "/home/") {
		parts := strings.Split(path, "/")
		if len(parts) >= 3 {
			return "~/" + strings.Join(parts[3:], "/")
		}
	}
	return path
}

func (o *statusOptions) formatLastUpdated(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}

	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		return fmt.Sprintf("%dm ago", int(diff.Minutes()))
	case diff < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(diff.Hours()))
	case diff < 7*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(diff.Hours()/24))
	case diff < 30*24*time.Hour:
		return fmt.Sprintf("%dw ago", int(diff.Hours()/(24*7)))
	default:
		return fmt.Sprintf("%dm ago", int(diff.Hours()/(24*30)))
	}
}

func (o *statusOptions) formatDetailedTime(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}
	return t.Format("2006-01-02 15:04:05")
}

func (o *statusOptions) truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func (o *statusOptions) escapeJSON(s string) string {
	// Basic JSON string escaping
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
}

func (o *statusOptions) escapeYAML(s string) string {
	// Basic YAML string escaping
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	return s
}

func (o *statusOptions) formatInstallMethod(method, path string) string {
	switch method {
	case "appimage":
		return "AppImage"
	case "pacman":
		return "Pacman (Arch Linux)"
	case "snap":
		return "Snap"
	case "flatpak":
		return "Flatpak"
	case "toolbox":
		return "JetBrains Toolbox"
	case "direct":
		return "Direct Installation"
	default:
		if method != "" {
			return method
		}
		return "Unknown"
	}
}
