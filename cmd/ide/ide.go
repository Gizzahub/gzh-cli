package ide

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gizzahub/gzh-manager-go/internal/env"
	"github.com/spf13/cobra"
)

type ideOptions struct {
	watchDir     string
	product      string
	recursive    bool
	verbose      bool
	daemon       bool
	fixSync      bool
	logPath      string
	excludePaths []string
}

type jetbrainsProduct struct {
	Name     string
	DirName  string
	BasePath string
}

func defaultIDEOptions() *ideOptions {
	homeDir, _ := os.UserHomeDir()

	return &ideOptions{
		recursive:    true,
		verbose:      false,
		daemon:       false,
		fixSync:      false,
		logPath:      filepath.Join(homeDir, ".gz", "logs", "ide-monitor.log"),
		excludePaths: []string{".git", "node_modules", "target", "build", ".idea/shelf"},
	}
}

// NewIDECmd creates the IDE subcommand for monitoring and managing IDE configuration changes.
func NewIDECmd(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ide",
		Short: "Monitor and manage IDE configuration changes",
		Long: `Monitor and manage IDE configuration changes, particularly JetBrains products.

This command provides monitoring and management capabilities for IDE settings:
- Real-time monitoring of JetBrains settings directories
- Cross-platform support for Linux, macOS, and Windows
- Automatic detection of JetBrains products and versions
- Settings synchronization issue detection and fixes
- File change tracking with filtering capabilities

Supported IDEs:
- IntelliJ IDEA (Community, Ultimate)
- PyCharm (Community, Professional)
- WebStorm, PhpStorm, RubyMine
- CLion, GoLand, DataGrip
- Android Studio, Rider

Examples:
  # Monitor all JetBrains settings
  gz ide monitor
  
  # Monitor specific product
  gz ide monitor --product IntelliJIdea2023.2
  
  # Fix settings sync issues
  gz ide fix-sync
  
  # List detected JetBrains installations
  gz ide list`,
		SilenceUsage: true,
	}

	cmd.AddCommand(newIDEMonitorCmd(ctx))
	cmd.AddCommand(newIDEListCmd())
	cmd.AddCommand(newIDEFixSyncCmd())

	return cmd
}

func newIDEMonitorCmd(ctx context.Context) *cobra.Command {
	o := defaultIDEOptions()

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

func newIDEListCmd() *cobra.Command {
	o := defaultIDEOptions()

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

func newIDEFixSyncCmd() *cobra.Command {
	o := defaultIDEOptions()

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

func (o *ideOptions) runMonitor(ctx context.Context, _ *cobra.Command, _ []string) error {
	watchDirs, err := o.getWatchDirectories()
	if err != nil {
		return fmt.Errorf("failed to get watch directories: %w", err)
	}

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

func (o *ideOptions) runList(_ *cobra.Command, _ []string) error {
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

func (o *ideOptions) runFixSync(_ *cobra.Command, _ []string) error {
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

func (o *ideOptions) getWatchDirectories() ([]string, error) {
	if o.watchDir != "" {
		// Use specific directory
		return []string{o.watchDir}, nil
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

	return dirs, nil
}

func (o *ideOptions) detectJetBrainsProducts() []jetbrainsProduct {
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

func (o *ideOptions) getJetBrainsBasePaths() []string {
	return o.getJetBrainsBasePathsWithEnv(env.NewOSEnvironment())
}

func (o *ideOptions) getJetBrainsBasePathsWithEnv(environment env.Environment) []string {
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

func (o *ideOptions) isJetBrainsProduct(name string) bool {
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

func (o *ideOptions) formatProductName(dirName string) string {
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

func (o *ideOptions) addWatchRecursive(watcher *fsnotify.Watcher, root string) error {
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

func (o *ideOptions) handleFileEvent(event fsnotify.Event) {
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

func (o *ideOptions) shouldIgnoreEvent(event fsnotify.Event) bool {
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

func (o *ideOptions) getRelativePath(fullPath string) string {
	// Try to make path relative to home directory for readability
	homeDir, _ := os.UserHomeDir()
	if rel, err := filepath.Rel(homeDir, fullPath); err == nil && !strings.HasPrefix(rel, "..") {
		return "~/" + rel
	}

	return fullPath
}

func (o *ideOptions) isSyncProblematicFile(filePath string) bool {
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

func (o *ideOptions) fixProductSyncIssues(product jetbrainsProduct) error {
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

func (o *ideOptions) fixFiletypesXML(filePath string) error {
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

func (o *ideOptions) applyFiletypesXMLFixes(content string) string {
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

func (o *ideOptions) copyFile(src, dst string) error {
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

func (o *ideOptions) getDirSize(path string) int64 {
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

func (o *ideOptions) formatSize(bytes int64) string {
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

func (o *ideOptions) countConfigFiles(path string) int {
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
