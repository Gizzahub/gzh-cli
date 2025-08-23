// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package monitor

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"

	"github.com/Gizzahub/gzh-cli/internal/env"
)

type monitorOptions struct {
	watchDir     string
	product      string
	recursive    bool
	verbose      bool
	daemon       bool
	logPath      string
	excludePaths []string
}

type jetbrainsProduct struct {
	Name     string
	DirName  string
	BasePath string
}

func defaultMonitorOptions() *monitorOptions {
	homeDir, _ := os.UserHomeDir()

	return &monitorOptions{
		recursive:    true,
		verbose:      false,
		daemon:       false,
		logPath:      filepath.Join(homeDir, ".gz", "logs", "ide-monitor.log"),
		excludePaths: []string{".git", "node_modules", "target", "build", ".idea/shelf"},
	}
}

// NewCmd creates the IDE monitor subcommand
func NewCmd(ctx context.Context) *cobra.Command {
	o := defaultMonitorOptions()

	cmd := &cobra.Command{
		Use:   "monitor",
		Short: "Monitor JetBrains settings for changes",
		Long: `Monitor JetBrains IDE settings directories for file changes.

This command watches JetBrains settings directories and reports any changes
in real-time. It can help track settings modifications, detect sync issues,
and monitor configuration changes across different IDE installations.

Examples:
  # Monitor all JetBrains products
  gz ide monitor

  # Monitor specific product with verbose output
  gz ide monitor --product PyCharm2024.3 --verbose

  # Run as daemon with logging
  gz ide monitor --daemon --log /var/log/ide-monitor.log

  # Monitor with custom directory
  gz ide monitor --watch-dir ~/.config/JetBrains/IntelliJIdea2023.2`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.runMonitor(ctx, cmd, args)
		},
	}

	cmd.Flags().StringVar(&o.watchDir, "watch-dir", "", "Specific directory to monitor (auto-detect if not specified)")
	cmd.Flags().StringVar(&o.product, "product", "", "Specific JetBrains product to monitor")
	cmd.Flags().BoolVar(&o.recursive, "recursive", true, "Monitor subdirectories recursively")
	cmd.Flags().BoolVar(&o.verbose, "verbose", false, "Enable verbose logging")
	cmd.Flags().BoolVar(&o.daemon, "daemon", false, "Run as background daemon")
	cmd.Flags().StringVar(&o.logPath, "log", o.logPath, "Log file path (used when running as daemon)")
	cmd.Flags().StringSliceVar(&o.excludePaths, "exclude", o.excludePaths, "Paths to exclude from monitoring")

	return cmd
}

func (o *monitorOptions) runMonitor(ctx context.Context, _ *cobra.Command, _ []string) error {
	watchDirs := o.getWatchDirectories()

	if len(watchDirs) == 0 {
		fmt.Println("⚠️  No JetBrains IDE installations found")
		return nil
	}

	fmt.Printf("🔍 Starting IDE settings monitor\n")
	fmt.Printf("   Monitoring %d directories\n", len(watchDirs))

	if o.verbose {
		for _, dir := range watchDirs {
			fmt.Printf("   - %s\n", dir)
		}
	}

	fmt.Printf("   Recursive: %v\n", o.recursive)
	fmt.Printf("   Excludes: %s\n\n", strings.Join(o.excludePaths, ", "))

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}
	defer func() {
		if err := watcher.Close(); err != nil {
			fmt.Printf("Warning: Failed to close file watcher: %v\n", err)
		}
	}()

	// Add directories to watcher
	for _, dir := range watchDirs {
		if err := o.addWatchRecursive(watcher, dir); err != nil {
			fmt.Printf("⚠️  Warning: Could not watch %s: %v\n", dir, err)
		}
	}

	fmt.Printf("📁 Watching %d paths for changes\n", len(watcher.WatchList()))
	fmt.Printf("🎯 Press Ctrl+C to stop monitoring\n\n")

	// Start monitoring with graceful shutdown support
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("\n🛑 Stopping IDE monitoring (reason: %v)\n", ctx.Err())
			return nil

		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}

			o.handleFileEvent(event)

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}

			fmt.Printf("❌ Watcher error: %v\n", err)
		}
	}
}

func (o *monitorOptions) getWatchDirectories() []string {
	if o.watchDir != "" {
		// Use specific directory
		return []string{o.watchDir}
	}

	// Auto-detect JetBrains directories
	products := o.detectJetBrainsProducts()

	// Filter by specific product if specified
	if o.product != "" {
		filtered := []jetbrainsProduct{}

		for _, p := range products {
			if strings.Contains(p.DirName, o.product) {
				filtered = append(filtered, p)
			}
		}

		products = filtered
	}

	dirs := make([]string, 0, len(products))
	for _, product := range products {
		dirs = append(dirs, product.BasePath)
	}

	return dirs
}

func (o *monitorOptions) detectJetBrainsProducts() []jetbrainsProduct {
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

func (o *monitorOptions) getJetBrainsBasePaths() []string {
	return o.getJetBrainsBasePathsWithEnv(env.NewOSEnvironment())
}

func (o *monitorOptions) getJetBrainsBasePathsWithEnv(environment env.Environment) []string {
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

func (o *monitorOptions) isJetBrainsProduct(name string) bool {
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

func (o *monitorOptions) formatProductName(dirName string) string {
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

func (o *monitorOptions) addWatchRecursive(watcher *fsnotify.Watcher, root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err // Skip directories we can't access
		}

		if !info.IsDir() {
			return nil
		}

		// Check if path should be excluded
		for _, exclude := range o.excludePaths {
			if strings.Contains(path, exclude) {
				if info.IsDir() {
					return filepath.SkipDir
				}

				return nil
			}
		}

		return watcher.Add(path)
	})
}

func (o *monitorOptions) handleFileEvent(event fsnotify.Event) {
	// Filter out certain events we don't care about
	if o.shouldIgnoreEvent(event) {
		return
	}

	timestamp := time.Now().Format("15:04:05")
	relativePath := o.getRelativePath(event.Name)

	var icon string

	switch {
	case event.Op&fsnotify.Create == fsnotify.Create:
		icon = "📝"
	case event.Op&fsnotify.Write == fsnotify.Write:
		icon = "✏️"
	case event.Op&fsnotify.Remove == fsnotify.Remove:
		icon = "🗑️"
	case event.Op&fsnotify.Rename == fsnotify.Rename:
		icon = "📝"
	case event.Op&fsnotify.Chmod == fsnotify.Chmod:
		icon = "🔧"
	default:
		icon = "📁"
	}

	fmt.Printf("[%s] %s %s %s\n", timestamp, icon, event.Op.String(), relativePath)

	// Check for sync issues
	if o.isSyncProblematicFile(event.Name) {
		fmt.Printf("   ⚠️  Potential sync issue detected in: %s\n", relativePath)
	}

	if o.verbose {
		if info, err := os.Stat(event.Name); err == nil && !info.IsDir() {
			fmt.Printf("   Size: %d bytes\n", info.Size())
		}
	}
}

func (o *monitorOptions) shouldIgnoreEvent(event fsnotify.Event) bool {
	// Ignore temporary files and certain patterns
	name := filepath.Base(event.Name)

	ignorePatterns := []string{
		".tmp", "~", ".swp", ".DS_Store", "Thumbs.db",
		".lock", ".log", "___jb_", // JetBrains temp files
	}

	for _, pattern := range ignorePatterns {
		if strings.Contains(name, pattern) {
			return true
		}
	}

	// Ignore chmod events on directories (too noisy)
	if event.Op&fsnotify.Chmod == fsnotify.Chmod {
		if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
			return true
		}
	}

	return false
}

func (o *monitorOptions) getRelativePath(fullPath string) string {
	// Try to make path relative to home directory for readability
	homeDir, _ := os.UserHomeDir()
	if rel, err := filepath.Rel(homeDir, fullPath); err == nil && !strings.HasPrefix(rel, "..") {
		return "~/" + rel
	}

	return fullPath
}

func (o *monitorOptions) isSyncProblematicFile(filePath string) bool {
	problematicFiles := []string{
		"filetypes.xml",
		"settingsSync/options/filetypes.xml",
		"workspace.xml",
	}

	for _, problematic := range problematicFiles {
		if strings.Contains(filePath, problematic) {
			return true
		}
	}

	return false
}
