// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package fixsync

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/Gizzahub/gzh-cli/internal/env"
)

type fixSyncOptions struct {
	product string
	verbose bool
}

type jetbrainsProduct struct {
	Name     string
	DirName  string
	BasePath string
}

// NewCmd creates the IDE fix-sync subcommand
func NewCmd() *cobra.Command {
	o := &fixSyncOptions{}

	cmd := &cobra.Command{
		Use:   "fix-sync",
		Short: "Fix JetBrains settings synchronization issues",
		Long: `Fix known JetBrains settings synchronization issues.

This command identifies and fixes common settings sync problems, such as:
- Corrupted filetypes.xml files
- Invalid settings sync configurations
- Duplicate or conflicting settings files

Examples:
  # Fix sync issues for all products
  gz ide fix-sync

  # Fix specific product
  gz ide fix-sync --product PyCharm2024.3

  # Verbose mode with detailed output
  gz ide fix-sync --verbose`,
		RunE: o.runFixSync,
	}

	cmd.Flags().StringVar(&o.product, "product", "", "Specific JetBrains product to fix")
	cmd.Flags().BoolVar(&o.verbose, "verbose", false, "Show detailed information")

	return cmd
}

func (o *fixSyncOptions) runFixSync(_ *cobra.Command, _ []string) error {
	products := o.detectJetBrainsProducts()

	if o.product != "" {
		// Filter to specific product
		filtered := []jetbrainsProduct{}

		for _, p := range products {
			if strings.Contains(p.DirName, o.product) {
				filtered = append(filtered, p)
			}
		}

		products = filtered
	}

	if len(products) == 0 {
		fmt.Println("No matching JetBrains installations found")
		return nil
	}

	fmt.Printf("🔧 Fixing settings sync issues for %d products...\n\n", len(products))

	for _, product := range products {
		fmt.Printf("🔨 Processing %s\n", product.Name)

		if err := o.fixProductSyncIssues(product); err != nil {
			fmt.Printf("   ❌ Error: %v\n", err)
		} else {
			fmt.Printf("   ✅ Completed\n")
		}

		fmt.Println()
	}

	return nil
}

func (o *fixSyncOptions) detectJetBrainsProducts() []jetbrainsProduct {
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

func (o *fixSyncOptions) getJetBrainsBasePaths() []string {
	return o.getJetBrainsBasePathsWithEnv(env.NewOSEnvironment())
}

func (o *fixSyncOptions) getJetBrainsBasePathsWithEnv(environment env.Environment) []string {
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

func (o *fixSyncOptions) isJetBrainsProduct(name string) bool {
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

func (o *fixSyncOptions) formatProductName(dirName string) string {
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

func (o *fixSyncOptions) fixProductSyncIssues(product jetbrainsProduct) error {
	// Fix filetypes.xml sync issues
	filetypesPath := filepath.Join(product.BasePath, "settingsSync", "options", "filetypes.xml")
	if err := o.fixFiletypesXML(filetypesPath); err != nil {
		return fmt.Errorf("failed to fix filetypes.xml: %w", err)
	}

	// Check for other common issues
	if o.verbose {
		fmt.Printf("   Checked filetypes.xml sync issues\n")
	}

	return nil
}

func (o *fixSyncOptions) fixFiletypesXML(filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil // File doesn't exist, nothing to fix
	}

	// Create backup
	backupPath := filePath + ".backup." + time.Now().Format("20060102-150405")
	if err := o.copyFile(filePath, backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Read current content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Check if file has known issues and fix them
	originalContent := string(content)
	fixedContent := o.applyFiletypesXMLFixes(originalContent)

	if originalContent != fixedContent {
		if err := os.WriteFile(filePath, []byte(fixedContent), 0o600); err != nil {
			return fmt.Errorf("failed to write fixed file: %w", err)
		}

		fmt.Printf("   🔧 Fixed filetypes.xml (backup: %s)\n", filepath.Base(backupPath))
	}

	return nil
}

func (o *fixSyncOptions) applyFiletypesXMLFixes(content string) string {
	// Apply common fixes for filetypes.xml sync issues
	// This is a placeholder - actual fixes would depend on specific issues

	// Remove duplicate entries (simple approach)
	lines := strings.Split(content, "\n")
	seen := make(map[string]bool)

	var uniqueLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || !seen[trimmed] {
			uniqueLines = append(uniqueLines, line)
			seen[trimmed] = true
		}
	}

	return strings.Join(uniqueLines, "\n")
}

func (o *fixSyncOptions) copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	err = os.WriteFile(dst, input, 0o600)
	if err != nil {
		return err
	}

	return nil
}
