// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package pm

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

// PackageManagerInfo holds information about a package manager.
type PackageManagerInfo struct {
	Name         string
	Command      string
	Version      string
	Installed    bool
	Configured   bool
	PackageCount int
}

func newStatusCmd(ctx context.Context) *cobra.Command {
	var (
		jsonOutput bool
		manager    string
	)

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show status of all configured package managers",
		Long: `Display the current status of all package managers including:
- Installation status
- Version information
- Number of managed packages
- Configuration status

Examples:
  # Show status of all package managers
  gz pm status

  # Show status in JSON format
  gz pm status --json

  # Show status of specific manager
  gz pm status --manager brew`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(ctx, manager, jsonOutput)
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	cmd.Flags().StringVar(&manager, "manager", "", "Show status of specific package manager")

	return cmd
}

func runStatus(ctx context.Context, specificManager string, jsonOutput bool) error {
	managers := []PackageManagerInfo{
		{Name: "brew", Command: "brew"},
		{Name: "asdf", Command: "asdf"},
		{Name: "sdkman", Command: "sdk"},
		{Name: "apt", Command: "apt"},
		{Name: "port", Command: "port"},
		{Name: "rbenv", Command: "rbenv"},
		{Name: "pyenv", Command: "pyenv"},
		{Name: "nvm", Command: "nvm"},
		{Name: "pip", Command: "pip"},
		{Name: "npm", Command: "npm"},
		{Name: "gem", Command: "gem"},
		{Name: "cargo", Command: "cargo"},
		{Name: "go", Command: "go"},
	}

	// Filter if specific manager requested
	if specificManager != "" {
		filtered := []PackageManagerInfo{}
		for _, m := range managers {
			if m.Name == specificManager {
				filtered = append(filtered, m)
				break
			}
		}
		if len(filtered) == 0 {
			return fmt.Errorf("unknown package manager: %s", specificManager)
		}
		managers = filtered
	}

	// Check status of each manager
	for i := range managers {
		checkManagerStatus(ctx, &managers[i])
		checkConfiguration(&managers[i])
	}

	// Output results
	if jsonOutput {
		return outputJSON(managers)
	}
	return outputTable(managers)
}

func checkManagerStatus(ctx context.Context, info *PackageManagerInfo) {
	// Check if installed
	switch info.Name {
	case "brew":
		if out, err := exec.CommandContext(ctx, info.Command, "--version").Output(); err == nil {
			info.Installed = true
			info.Version = strings.TrimSpace(strings.Split(string(out), "\n")[0])
			// Count installed packages
			if out, err := exec.CommandContext(ctx, info.Command, "list", "--formula").Output(); err == nil {
				info.PackageCount = len(strings.Split(strings.TrimSpace(string(out)), "\n"))
			}
		}
	case "asdf":
		if out, err := exec.CommandContext(ctx, info.Command, "version").Output(); err == nil {
			info.Installed = true
			info.Version = strings.TrimSpace(string(out))
			// Count plugins
			if out, err := exec.CommandContext(ctx, info.Command, "plugin", "list").Output(); err == nil {
				info.PackageCount = len(strings.Split(strings.TrimSpace(string(out)), "\n"))
			}
		}
	case "sdkman":
		// SDKMAN requires sourcing, check if directory exists
		sdkmanDir := os.Getenv("SDKMAN_DIR")
		if sdkmanDir == "" {
			sdkmanDir = os.Getenv("HOME") + "/.sdkman"
		}
		if _, err := os.Stat(sdkmanDir); err == nil {
			info.Installed = true
			info.Version = "installed"
			// Count candidates
			candidatesDir := sdkmanDir + "/candidates"
			if entries, err := os.ReadDir(candidatesDir); err == nil {
				info.PackageCount = len(entries)
			}
		}
	case "pip", "npm", "gem", "cargo", "go":
		if out, err := exec.CommandContext(ctx, info.Command, "--version").Output(); err == nil {
			info.Installed = true
			info.Version = strings.TrimSpace(string(out))
		}
	default:
		// Generic check
		if _, err := exec.LookPath(info.Command); err == nil {
			info.Installed = true
			if out, err := exec.CommandContext(ctx, info.Command, "--version").Output(); err == nil {
				info.Version = strings.TrimSpace(string(out))
			}
		}
	}
}

func checkConfiguration(info *PackageManagerInfo) {
	// Check if configuration file exists
	configDir := os.Getenv("HOME") + "/.gzh/pm"
	configFile := fmt.Sprintf("%s/%s.yml", configDir, info.Name)
	if _, err := os.Stat(configFile); err == nil {
		info.Configured = true
	}
}

func outputTable(managers []PackageManagerInfo) error {
	fmt.Println("Package Manager Status")
	fmt.Println("======================")
	fmt.Println()
	fmt.Printf("%-12s %-12s %-30s %-12s %s\n", "MANAGER", "STATUS", "VERSION", "PACKAGES", "CONFIG")
	fmt.Printf("%-12s %-12s %-30s %-12s %s\n", "-------", "------", "-------", "--------", "------")

	for _, m := range managers {
		status := "Not Found"
		if m.Installed {
			status = "Installed"
		}

		config := "No"
		if m.Configured {
			config = "Yes"
		}

		packages := "-"
		if m.PackageCount > 0 {
			packages = fmt.Sprintf("%d", m.PackageCount)
		}

		version := m.Version
		if version == "" {
			version = "-"
		}
		if len(version) > 30 {
			version = version[:27] + "..."
		}

		fmt.Printf("%-12s %-12s %-30s %-12s %s\n", m.Name, status, version, packages, config)
	}

	fmt.Println()
	fmt.Printf("Configuration directory: %s/.gzh/pm/\n", os.Getenv("HOME"))

	return nil
}

func outputJSON(managers []PackageManagerInfo) error {
	// Simple JSON output (in real implementation, use encoding/json)
	fmt.Println("{")
	fmt.Println("  \"managers\": [")
	for i, m := range managers {
		fmt.Printf("    {\n")
		fmt.Printf("      \"name\": \"%s\",\n", m.Name)
		fmt.Printf("      \"installed\": %t,\n", m.Installed)
		fmt.Printf("      \"version\": \"%s\",\n", m.Version)
		fmt.Printf("      \"configured\": %t,\n", m.Configured)
		fmt.Printf("      \"package_count\": %d\n", m.PackageCount)
		fmt.Printf("    }")
		if i < len(managers)-1 {
			fmt.Printf(",")
		}
		fmt.Println()
	}
	fmt.Println("  ]")
	fmt.Println("}")
	return nil
}
