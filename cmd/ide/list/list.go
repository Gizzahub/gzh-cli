// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package list

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Gizzahub/gzh-cli/internal/env"
)

type listOptions struct {
	verbose bool
}

type jetbrainsProduct struct {
	Name     string
	DirName  string
	BasePath string
}

// NewCmd creates the IDE list subcommand
func NewCmd() *cobra.Command {
	o := &listOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List detected JetBrains IDE installations",
		Long: `List all detected JetBrains IDE installations and their settings directories.

This command scans the system for JetBrains IDE installations and displays
their configuration directories, versions, and status information.

Examples:
  # List all detected installations
  gz ide list

  # List with verbose information
  gz ide list --verbose`,
		RunE: o.runList,
	}

	cmd.Flags().BoolVar(&o.verbose, "verbose", false, "Show detailed information")

	return cmd
}

func (o *listOptions) runList(_ *cobra.Command, _ []string) error {
	products := o.detectJetBrainsProducts()

	if len(products) == 0 {
		fmt.Println("No JetBrains IDE installations found")
		return nil
	}

	fmt.Printf("🚀 JetBrains IDE Installations (%d found):\n\n", len(products))

	for _, product := range products {
		fmt.Printf("📦 %s\n", product.Name)
		fmt.Printf("   Path: %s\n", product.BasePath)

		if o.verbose {
			// Check if directory exists and get info
			if info, err := os.Stat(product.BasePath); err == nil {
				fmt.Printf("   Size: %s\n", o.formatSize(o.getDirSize(product.BasePath)))
				fmt.Printf("   Modified: %s\n", info.ModTime().Format("2006-01-02 15:04:05"))

				// Count configuration files
				configFiles := o.countConfigFiles(product.BasePath)
				fmt.Printf("   Config files: %d\n", configFiles)
			} else {
				fmt.Printf("   Status: Directory not accessible\n")
			}
		}

		fmt.Println()
	}

	return nil
}

func (o *listOptions) detectJetBrainsProducts() []jetbrainsProduct {
	var products []jetbrainsProduct

	basePaths := o.getJetBrainsBasePaths()

	for _, basePath := range basePaths {
		if _, err := os.Stat(basePath); os.IsNotExist(err) {
			continue
		}

		entries, err := os.ReadDir(basePath)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			name := entry.Name()
			if o.isJetBrainsProduct(name) {
				product := jetbrainsProduct{
					Name:     o.formatProductName(name),
					DirName:  name,
					BasePath: filepath.Join(basePath, name),
				}
				products = append(products, product)
			}
		}
	}

	return products
}

func (o *listOptions) getJetBrainsBasePaths() []string {
	return o.getJetBrainsBasePathsWithEnv(env.NewOSEnvironment())
}

func (o *listOptions) getJetBrainsBasePathsWithEnv(environment env.Environment) []string {
	switch runtime.GOOS {
	case "linux":
		homeDir, _ := os.UserHomeDir()

		return []string{
			filepath.Join(homeDir, ".config", "JetBrains"),
		}
	case "darwin":
		homeDir, _ := os.UserHomeDir()

		return []string{
			filepath.Join(homeDir, "Library", "Application Support", "JetBrains"),
		}
	case "windows":
		appData := environment.Get("APPDATA")
		if appData == "" {
			homeDir, _ := os.UserHomeDir()
			appData = filepath.Join(homeDir, "AppData", "Roaming")
		}

		return []string{
			filepath.Join(appData, "JetBrains"),
		}
	default:
		return []string{}
	}
}

func (o *listOptions) isJetBrainsProduct(name string) bool {
	jetbrainsProducts := []string{
		"IntelliJIdea", "PyCharm", "WebStorm", "PhpStorm", "RubyMine",
		"CLion", "GoLand", "DataGrip", "Rider", "AndroidStudio",
	}

	for _, product := range jetbrainsProducts {
		if strings.HasPrefix(name, product) {
			return true
		}
	}

	return false
}

func (o *listOptions) formatProductName(dirName string) string {
	// Extract product name and version
	for i, char := range dirName {
		if char >= '0' && char <= '9' {
			product := dirName[:i]
			version := dirName[i:]

			return fmt.Sprintf("%s %s", product, version)
		}
	}

	return dirName
}

func (o *listOptions) getDirSize(path string) int64 {
	var size int64

	_ = filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			size += info.Size()
		}

		return nil
	})

	return size
}

func (o *listOptions) formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func (o *listOptions) countConfigFiles(path string) int {
	count := 0
	_ = filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && (strings.HasSuffix(info.Name(), ".xml") || strings.HasSuffix(info.Name(), ".json")) {
			count++
		}

		return nil
	})

	return count
}
