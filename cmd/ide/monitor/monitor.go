// Copyright (c) 2025 Gizzahub
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

	"github.com/gizzahub/gzh-cli/internal/env"
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

// NewCmd creates the IDE monitor subcommand.
//
// ctx를 받지 않는다. 예전에는 받았지만 쓰지 않았고, 준 쪽(register.go)은
// context.Background()를 넣고 있었다. RunE가 cmd.Context()를 쓰므로 받을
// 까닭이 없다. 남겨 두면 취소되는 맥락처럼 보여서 다시 붙잡아 쓰기 쉽다.
func NewCmd() *cobra.Command {
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
			// 붙잡아 둔 ctx가 아니라 cmd.Context()를 쓴다. 이 명령을 짜
			// 넣는 register.go가 NewIDECmd에 context.Background()를 준다
			// -- registry.Provider의 Command()에 ctx 자리가 없어서 다섯
			// 제공자가 전부 그렇게 한다. 그것을 붙잡아 두면 취소가 오지
			// 않아 감시 고리에서 빠져나올 수 없다. cmd.Context()는 root의
			// ExecuteContext가 넣어 준 것이라 SIGINT/SIGTERM에 반응한다.
			return o.runMonitor(cmd.Context(), cmd, args)
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
			// --watch-dir로 콕 집어 준 것이 안 되면 바로 실패한다. 쓰는
			// 사람이 그 디렉토리를 보라고 한 것이지 아무거나 보라고 한 것이
			// 아니다. 경고만 흘리고 계속 가면 없는 경로를 지켜보는 줄 알고
			// 기다리게 된다.
			if o.watchDir != "" {
				return fmt.Errorf("cannot watch %s: %w", dir, err)
			}

			// 자동으로 찾은 것은 하나가 안 돼도 나머지가 남아 있을 수 있다.
			fmt.Printf("⚠️  Warning: Could not watch %s: %v\n", dir, err)
		}
	}

	// 한 곳도 못 붙였으면 감시 고리에 들어갈 까닭이 없다. 예전에는
	// "Watching 0 paths for changes"를 찍고 그대로 눌러앉았다 -- 아무 일도
	// 일어날 수 없는데 일어나기를 기다리는 상태였다.
	watchedPaths := len(watcher.WatchList())
	if watchedPaths == 0 {
		return fmt.Errorf("no directories could be watched (tried %d)", len(watchDirs))
	}

	fmt.Printf("📁 Watching %d paths for changes\n", watchedPaths)
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
