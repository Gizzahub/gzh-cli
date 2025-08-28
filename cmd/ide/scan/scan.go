// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package scan

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// IDE represents an integrated development environment.
type IDE struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"` // jetbrains, vscode, other
	Executable  string   `json:"executable"`
	Version     string   `json:"version"`
	Aliases     []string `json:"aliases"`
	ProductCode string   `json:"product_code,omitempty"`
}

// IDEDetector interface for detecting IDEs.
type IDEDetector interface {
	DetectIDEs(useCache bool) ([]IDE, error)
}

// NewIDEDetector creates a new IDE detector - placeholder implementation.
func NewIDEDetector() IDEDetector {
	return &mockDetector{}
}

// mockDetector is a simple mock implementation for now.
type mockDetector struct{}

func (d *mockDetector) DetectIDEs(useCache bool) ([]IDE, error) {
	// Mock implementation - in reality this would detect actual IDEs
	return []IDE{
		{
			Name:       "VS Code",
			Type:       "vscode",
			Executable: "/usr/local/bin/code",
			Version:    "1.85.0",
			Aliases:    []string{"code", "vscode"},
		},
		{
			Name:       "PyCharm Professional",
			Type:       "jetbrains",
			Executable: "/usr/local/bin/pycharm",
			Version:    "2023.3.2",
			Aliases:    []string{"pycharm"},
		},
		{
			Name:       "IntelliJ IDEA Ultimate",
			Type:       "jetbrains",
			Executable: "/usr/local/bin/idea",
			Version:    "2023.3",
			Aliases:    []string{"idea", "intellij"},
		},
		{
			Name:       "Cursor",
			Type:       "vscode",
			Executable: "/usr/local/bin/cursor",
			Version:    "0.17.3",
			Aliases:    []string{"cursor"},
		},
		{
			Name:       "Vim",
			Type:       "other",
			Executable: "/usr/bin/vim",
			Version:    "9.0",
			Aliases:    []string{"vim"},
		},
	}, nil
}

type scanOptions struct {
	refresh bool
	verbose bool
}

// NewCmd creates the IDE scan subcommand
func NewCmd() *cobra.Command {
	o := &scanOptions{}

	cmd := &cobra.Command{
		Use:   "scan",
		Short: "Scan for installed IDEs",
		Long: `Scan the system for installed IDE applications.

This command searches for IDE installations across different locations:
- JetBrains Toolbox and system installations
- VS Code family (VS Code, VS Code Insiders, Cursor, VSCodium)
- Other popular editors (Sublime Text, Vim, Neovim, Emacs)

Results are cached for 24 hours to improve performance. Use --refresh to force a new scan.

Examples:
  # Scan for IDEs with caching
  gz ide scan

  # Force refresh scan
  gz ide scan --refresh

  # Verbose output with details
  gz ide scan --verbose`,
		RunE: o.runScan,
	}

	cmd.Flags().BoolVar(&o.refresh, "refresh", false, "Force refresh scan (ignore cache)")
	cmd.Flags().BoolVar(&o.verbose, "verbose", false, "Show detailed scan information")

	return cmd
}

func (o *scanOptions) runScan(cmd *cobra.Command, args []string) error {
	detector := NewIDEDetector()

	fmt.Printf("🔍 Scanning for installed IDEs...\n")
	if o.refresh {
		fmt.Printf("   (forced refresh, ignoring cache)\n")
	}
	fmt.Println()

	// Detect IDEs
	ides, err := detector.DetectIDEs(!o.refresh)
	if err != nil {
		return fmt.Errorf("failed to detect IDEs: %w", err)
	}

	if len(ides) == 0 {
		fmt.Printf("❌ No IDEs found on this system\n")
		fmt.Printf("\nConsider installing:\n")
		fmt.Printf("  • VS Code: https://code.visualstudio.com/\n")
		fmt.Printf("  • JetBrains Toolbox: https://www.jetbrains.com/toolbox-app/\n")
		fmt.Printf("  • Cursor: https://cursor.sh/\n")
		return nil
	}

	// Group IDEs by type
	jetbrainsIDEs := []IDE{}
	vscodeIDEs := []IDE{}
	otherIDEs := []IDE{}

	for _, ide := range ides {
		switch ide.Type {
		case "jetbrains":
			jetbrainsIDEs = append(jetbrainsIDEs, ide)
		case "vscode":
			vscodeIDEs = append(vscodeIDEs, ide)
		default:
			otherIDEs = append(otherIDEs, ide)
		}
	}

	// Display results
	fmt.Printf("✅ Found %d IDEs:\n\n", len(ides))

	if len(jetbrainsIDEs) > 0 {
		o.displayIDEGroup("JetBrains IDEs", jetbrainsIDEs)
	}

	if len(vscodeIDEs) > 0 {
		o.displayIDEGroup("VS Code Family", vscodeIDEs)
	}

	if len(otherIDEs) > 0 {
		o.displayIDEGroup("Other Editors", otherIDEs)
	}

	// Show usage hints
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  • Run 'gz ide status' to see detailed information\n")
	fmt.Printf("  • Run 'gz ide open <name>' to open an IDE\n")
	fmt.Printf("  • Available aliases: %s\n", o.getAvailableAliases(ides))

	return nil
}

func (o *scanOptions) displayIDEGroup(groupName string, ides []IDE) {
	fmt.Printf("📦 %s (%d found):\n", groupName, len(ides))

	for _, ide := range ides {
		fmt.Printf("   ✓ %s", ide.Name)

		if ide.Version != "unknown" && ide.Version != "" {
			fmt.Printf(" (v%s)", ide.Version)
		}

		if o.verbose {
			fmt.Printf("\n     Path: %s", ide.Executable)
			if len(ide.Aliases) > 0 {
				fmt.Printf("\n     Aliases: %s", fmt.Sprintf("[%s]", strings.Join(ide.Aliases, ", ")))
			}
		}

		fmt.Println()
	}

	fmt.Println()
}

func (o *scanOptions) getAvailableAliases(ides []IDE) string {
	var aliases []string
	seen := make(map[string]bool)

	for _, ide := range ides {
		for _, alias := range ide.Aliases {
			if !seen[alias] {
				aliases = append(aliases, alias)
				seen[alias] = true
			}
		}
	}

	if len(aliases) == 0 {
		return "none"
	}

	return strings.Join(aliases, ", ")
}
