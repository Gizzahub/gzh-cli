// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// CacheInfo holds information about cache for a specific package manager.
type CacheInfo struct {
	Name      string `json:"name"`
	Path      string `json:"path,omitempty"`
	Size      int64  `json:"size"`
	SizeHuman string `json:"sizeHuman"`
	Exists    bool   `json:"exists"`
	Error     string `json:"error,omitempty"`
}

// CacheStatus holds the overall cache status.
type CacheStatus struct {
	TotalSize      int64       `json:"totalSize"`
	TotalSizeHuman string      `json:"totalSizeHuman"`
	Caches         []CacheInfo `json:"caches"`
	Timestamp      time.Time   `json:"timestamp"`
}

// CleanResult holds the result of cache cleaning operation.
type CleanResult struct {
	Manager        string `json:"manager"`
	Success        bool   `json:"success"`
	SizeBefore     int64  `json:"sizeBefore"`
	SizeAfter      int64  `json:"sizeAfter"`
	SizeFreed      int64  `json:"sizeFreed"`
	SizeFreedHuman string `json:"sizeFreedHuman"`
	Command        string `json:"command,omitempty"`
	Error          string `json:"error,omitempty"`
}

// PackageManagerCache defines cache operations for a package manager.
type PackageManagerCache struct {
	Name         string
	CheckCmd     []string
	CleanCmd     []string
	CachePaths   []string
	RequiresSudo bool
	OSSupport    []string // linux, darwin, windows
}

// getSupportedCacheManagers returns list of supported package managers for cache operations
func getSupportedCacheManagers() []PackageManagerCache {
	homeDir, _ := os.UserHomeDir()

	return []PackageManagerCache{
		{
			Name:     "go",
			CheckCmd: []string{"go", "version"},
			CleanCmd: []string{"go", "clean", "-modcache", "-cache", "-testcache"},
			CachePaths: []string{
				filepath.Join(homeDir, "go", "pkg", "mod"),
				filepath.Join(homeDir, ".cache", "go-build"),
			},
			OSSupport: []string{"linux", "darwin", "windows"},
		},
		{
			Name:     "npm",
			CheckCmd: []string{"npm", "--version"},
			CleanCmd: []string{"npm", "cache", "clean", "--force"},
			CachePaths: []string{
				filepath.Join(homeDir, ".npm"),
			},
			OSSupport: []string{"linux", "darwin", "windows"},
		},
		{
			Name:     "yarn",
			CheckCmd: []string{"yarn", "--version"},
			CleanCmd: []string{"yarn", "cache", "clean"},
			CachePaths: []string{
				filepath.Join(homeDir, ".yarn", "cache"),
				filepath.Join(homeDir, ".cache", "yarn"),
			},
			OSSupport: []string{"linux", "darwin", "windows"},
		},
		{
			Name:     "pnpm",
			CheckCmd: []string{"pnpm", "--version"},
			CleanCmd: []string{"pnpm", "store", "prune"},
			CachePaths: []string{
				filepath.Join(homeDir, ".pnpm-store"),
			},
			OSSupport: []string{"linux", "darwin", "windows"},
		},
		{
			Name:     "pip",
			CheckCmd: []string{"pip", "--version"},
			CleanCmd: []string{"pip", "cache", "purge"},
			CachePaths: []string{
				filepath.Join(homeDir, ".cache", "pip"),
			},
			OSSupport: []string{"linux", "darwin", "windows"},
		},
		{
			Name:     "poetry",
			CheckCmd: []string{"poetry", "--version"},
			CleanCmd: []string{"poetry", "cache", "clear", "--all", "."},
			CachePaths: []string{
				filepath.Join(homeDir, ".cache", "pypoetry"),
			},
			OSSupport: []string{"linux", "darwin", "windows"},
		},
		{
			Name:     "cargo",
			CheckCmd: []string{"cargo", "--version"},
			CleanCmd: []string{"cargo", "cache", "-a"},
			CachePaths: []string{
				filepath.Join(homeDir, ".cargo", "registry"),
				filepath.Join(homeDir, ".cargo", "git"),
			},
			OSSupport: []string{"linux", "darwin", "windows"},
		},
		{
			Name:     "brew",
			CheckCmd: []string{"brew", "--version"},
			CleanCmd: []string{"brew", "cleanup", "-s"},
			CachePaths: []string{
				"/usr/local/var/homebrew/cache",
				"/opt/homebrew/var/homebrew/cache",
			},
			OSSupport: []string{"darwin"},
		},
		{
			Name:         "apt",
			CheckCmd:     []string{"apt-get", "--version"},
			CleanCmd:     []string{"sh", "-c", "apt-get clean && apt-get autoclean"},
			CachePaths:   []string{"/var/cache/apt"},
			RequiresSudo: true,
			OSSupport:    []string{"linux"},
		},
		{
			Name:         "dnf",
			CheckCmd:     []string{"dnf", "--version"},
			CleanCmd:     []string{"dnf", "clean", "all"},
			CachePaths:   []string{"/var/cache/dnf"},
			RequiresSudo: true,
			OSSupport:    []string{"linux"},
		},
		{
			Name:         "yum",
			CheckCmd:     []string{"yum", "--version"},
			CleanCmd:     []string{"yum", "clean", "all"},
			CachePaths:   []string{"/var/cache/yum"},
			RequiresSudo: true,
			OSSupport:    []string{"linux"},
		},
		{
			Name:         "pacman",
			CheckCmd:     []string{"pacman", "--version"},
			CleanCmd:     []string{"pacman", "-Sc", "--noconfirm"},
			CachePaths:   []string{"/var/cache/pacman/pkg"},
			RequiresSudo: true,
			OSSupport:    []string{"linux"},
		},
	}
}

// isManagerAvailable checks if a package manager is available on the system
func isManagerAvailable(manager PackageManagerCache) bool {
	// Check OS support
	currentOS := runtime.GOOS
	supported := false
	for _, os := range manager.OSSupport {
		if os == currentOS {
			supported = true
			break
		}
	}
	if !supported {
		return false
	}

	// Check if command exists
	if len(manager.CheckCmd) == 0 {
		return false
	}

	cmd := exec.Command(manager.CheckCmd[0], manager.CheckCmd[1:]...)
	err := cmd.Run()
	return err == nil
}

// getCacheSize calculates the total size of cache directories
func getCacheSize(paths []string) int64 {
	var totalSize int64

	for _, path := range paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}

		err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Skip errors
			}
			if !info.IsDir() {
				totalSize += info.Size()
			}
			return nil
		})
		if err != nil {
			continue // Skip on error
		}
	}

	return totalSize
}

// formatSize converts bytes to human readable format.
func formatSize(bytes int64) string {
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

func newCacheCleanCmd(ctx context.Context) *cobra.Command {
	var (
		dryRun     bool
		force      bool
		jsonOutput bool
		managers   struct {
			go_    bool
			npm    bool
			yarn   bool
			pnpm   bool
			pip    bool
			poetry bool
			cargo  bool
			brew   bool
			apt    bool
			dnf    bool
			yum    bool
			pacman bool
		}
	)

	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Clean package manager caches",
		Long: `Clean caches for specified package managers or all available ones.

By default, this command will prompt for confirmation before cleaning caches.
Use --force to skip confirmation prompts.

Examples:
  # Clean all available caches (with confirmation)
  gz pm cache clean

  # Clean specific manager caches
  gz pm cache clean --go --npm

  # Dry-run to see what would be cleaned
  gz pm cache clean --dry-run

  # Force clean without confirmation
  gz pm cache clean --force

  # Clean with JSON output
  gz pm cache clean --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			availableManagers := getSupportedCacheManagers()
			var selectedManagers []PackageManagerCache

			// Determine which managers to clean
			if managers.go_ || managers.npm || managers.yarn || managers.pnpm ||
				managers.pip || managers.poetry || managers.cargo || managers.brew ||
				managers.apt || managers.dnf || managers.yum || managers.pacman {
				// Clean only selected managers
				managerFlags := map[string]bool{
					"go":     managers.go_,
					"npm":    managers.npm,
					"yarn":   managers.yarn,
					"pnpm":   managers.pnpm,
					"pip":    managers.pip,
					"poetry": managers.poetry,
					"cargo":  managers.cargo,
					"brew":   managers.brew,
					"apt":    managers.apt,
					"dnf":    managers.dnf,
					"yum":    managers.yum,
					"pacman": managers.pacman,
				}

				for _, mgr := range availableManagers {
					if managerFlags[mgr.Name] && isManagerAvailable(mgr) {
						selectedManagers = append(selectedManagers, mgr)
					}
				}
			} else {
				// Clean all available managers
				for _, mgr := range availableManagers {
					if isManagerAvailable(mgr) {
						selectedManagers = append(selectedManagers, mgr)
					}
				}
			}

			if len(selectedManagers) == 0 {
				return fmt.Errorf("no package managers available for cache cleaning")
			}

			// Show what will be cleaned
			if !jsonOutput {
				fmt.Printf("🧹 Cache cleaning plan:\n\n")
				for _, mgr := range selectedManagers {
					size := getCacheSize(mgr.CachePaths)
					fmt.Printf("  %-10s %s", mgr.Name, formatSize(size))
					if mgr.RequiresSudo {
						fmt.Printf(" (requires sudo)")
					}
					fmt.Println()
				}
				fmt.Println()
			}

			// Confirmation prompt (unless force or dry-run)
			if !force && !dryRun && !jsonOutput {
				fmt.Print("Do you want to proceed with cache cleaning? [y/N]: ")
				var response string
				fmt.Scanln(&response)
				if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
					fmt.Println("Cache cleaning cancelled.")
					return nil
				}
			}

			var results []CleanResult

			for _, mgr := range selectedManagers {
				result := CleanResult{
					Manager: mgr.Name,
					Command: strings.Join(mgr.CleanCmd, " "),
				}

				// Get size before cleaning
				result.SizeBefore = getCacheSize(mgr.CachePaths)

				if dryRun {
					result.Success = true
					result.SizeAfter = result.SizeBefore
					result.SizeFreed = 0
					if !jsonOutput {
						fmt.Printf("🔍 [DRY-RUN] Would clean %s cache (%s)\n",
							mgr.Name, formatSize(result.SizeBefore))
					}
				} else {
					// Execute clean command
					var cmd *exec.Cmd
					if mgr.RequiresSudo {
						// Prepend sudo to the command
						cmdArgs := append([]string{"sudo"}, mgr.CleanCmd...)
						cmd = exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)
					} else {
						cmd = exec.CommandContext(ctx, mgr.CleanCmd[0], mgr.CleanCmd[1:]...)
					}

					if !jsonOutput {
						fmt.Printf("🧹 Cleaning %s cache...", mgr.Name)
					}

					err := cmd.Run()
					if err != nil {
						result.Success = false
						result.Error = err.Error()
						if !jsonOutput {
							fmt.Printf(" ❌ Failed: %v\n", err)
						}
					} else {
						result.Success = true
						result.SizeAfter = getCacheSize(mgr.CachePaths)
						result.SizeFreed = result.SizeBefore - result.SizeAfter
						result.SizeFreedHuman = formatSize(result.SizeFreed)

						if !jsonOutput {
							if result.SizeFreed > 0 {
								fmt.Printf(" ✅ Freed %s\n", result.SizeFreedHuman)
							} else {
								fmt.Printf(" ✅ Already clean\n")
							}
						}
					}
				}

				results = append(results, result)
			}

			if jsonOutput {
				output := map[string]interface{}{
					"timestamp": time.Now(),
					"dryRun":    dryRun,
					"results":   results,
				}

				jsonData, err := json.MarshalIndent(output, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to generate JSON output: %w", err)
				}
				fmt.Println(string(jsonData))
			} else {
				// Summary
				totalFreed := int64(0)
				successCount := 0
				for _, result := range results {
					if result.Success {
						successCount++
						totalFreed += result.SizeFreed
					}
				}

				fmt.Printf("\n📊 Summary:\n")
				fmt.Printf("  Managers processed: %d\n", len(results))
				fmt.Printf("  Successful: %d\n", successCount)
				if !dryRun {
					fmt.Printf("  Total space freed: %s\n", formatSize(totalFreed))
				}
			}

			return nil
		},
	}

	// Add flags for specific package managers
	cmd.Flags().BoolVar(&managers.go_, "go", false, "Clean Go module and build cache")
	cmd.Flags().BoolVar(&managers.npm, "npm", false, "Clean npm cache")
	cmd.Flags().BoolVar(&managers.yarn, "yarn", false, "Clean yarn cache")
	cmd.Flags().BoolVar(&managers.pnpm, "pnpm", false, "Clean pnpm store")
	cmd.Flags().BoolVar(&managers.pip, "pip", false, "Clean pip cache")
	cmd.Flags().BoolVar(&managers.poetry, "poetry", false, "Clean poetry cache")
	cmd.Flags().BoolVar(&managers.cargo, "cargo", false, "Clean cargo cache")
	cmd.Flags().BoolVar(&managers.brew, "brew", false, "Clean Homebrew cache")
	cmd.Flags().BoolVar(&managers.apt, "apt", false, "Clean APT cache")
	cmd.Flags().BoolVar(&managers.dnf, "dnf", false, "Clean DNF cache")
	cmd.Flags().BoolVar(&managers.yum, "yum", false, "Clean YUM cache")
	cmd.Flags().BoolVar(&managers.pacman, "pacman", false, "Clean Pacman cache")

	// Add global flags
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be cleaned without actually cleaning")
	cmd.Flags().BoolVar(&force, "force", false, "Clean without confirmation prompts")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output results in JSON format")

	return cmd
}

func newCacheStatusCmd(_ context.Context) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show cache status for all package managers",
		Long: `Display cache information for all available package managers including:
- Cache existence and location
- Cache size in human-readable format
- Total cache size across all managers

Examples:
  # Show cache status
  gz pm cache status

  # Show status in JSON format
  gz pm cache status --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			availableManagers := getSupportedCacheManagers()
			var caches []CacheInfo
			var totalSize int64

			for _, mgr := range availableManagers {
				if !isManagerAvailable(mgr) {
					continue
				}

				cacheInfo := CacheInfo{
					Name:   mgr.Name,
					Exists: false,
				}

				// Check cache paths and calculate total size
				for _, path := range mgr.CachePaths {
					if _, err := os.Stat(path); err == nil {
						cacheInfo.Exists = true
						if cacheInfo.Path == "" {
							cacheInfo.Path = path
						}
					}
				}

				cacheInfo.Size = getCacheSize(mgr.CachePaths)
				cacheInfo.SizeHuman = formatSize(cacheInfo.Size)
				totalSize += cacheInfo.Size

				caches = append(caches, cacheInfo)
			}

			status := CacheStatus{
				TotalSize:      totalSize,
				TotalSizeHuman: formatSize(totalSize),
				Caches:         caches,
				Timestamp:      time.Now(),
			}

			if jsonOutput {
				jsonData, err := json.MarshalIndent(status, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to generate JSON output: %w", err)
				}
				fmt.Println(string(jsonData))
			} else {
				fmt.Printf("📊 Package Manager Cache Status\n\n")
				fmt.Printf("%-12s %-10s %-8s %s\n", "Manager", "Size", "Status", "Path")
				fmt.Printf("%-12s %-10s %-8s %s\n", "-------", "----", "------", "----")

				for _, cache := range caches {
					status := "❌ Not found"
					if cache.Exists {
						status = "✅ Found"
					}
					fmt.Printf("%-12s %-10s %-8s %s\n",
						cache.Name, cache.SizeHuman, status, cache.Path)
				}

				fmt.Printf("\n📈 Total cache size: %s\n", status.TotalSizeHuman)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output status in JSON format")

	return cmd
}

func newCacheSizeCmd(_ context.Context) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "size",
		Short: "Show cache sizes for all package managers",
		Long: `Display cache sizes for all available package managers.
This is a quick way to see which caches are taking up the most space.

Examples:
  # Show cache sizes
  gz pm cache size

  # Show sizes in JSON format
  gz pm cache size --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			availableManagers := getSupportedCacheManagers()
			var caches []CacheInfo
			var totalSize int64

			for _, mgr := range availableManagers {
				if !isManagerAvailable(mgr) {
					continue
				}

				size := getCacheSize(mgr.CachePaths)
				if size > 0 {
					cacheInfo := CacheInfo{
						Name:      mgr.Name,
						Size:      size,
						SizeHuman: formatSize(size),
						Exists:    true,
					}
					caches = append(caches, cacheInfo)
					totalSize += size
				}
			}

			if jsonOutput {
				output := map[string]interface{}{
					"totalSize":      totalSize,
					"totalSizeHuman": formatSize(totalSize),
					"caches":         caches,
					"timestamp":      time.Now(),
				}

				jsonData, err := json.MarshalIndent(output, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to generate JSON output: %w", err)
				}
				fmt.Println(string(jsonData))
			} else {
				fmt.Printf("💾 Package Manager Cache Sizes\n\n")

				if len(caches) == 0 {
					fmt.Println("No caches found or all caches are empty.")
					return nil
				}

				fmt.Printf("%-12s %s\n", "Manager", "Size")
				fmt.Printf("%-12s %s\n", "-------", "----")

				for _, cache := range caches {
					fmt.Printf("%-12s %s\n", cache.Name, cache.SizeHuman)
				}

				fmt.Printf("\n📊 Total: %s\n", formatSize(totalSize))
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output sizes in JSON format")

	return cmd
}

func NewCacheCmd(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cache",
		Short: "Manage package manager caches",
		Long: `Manage caches for various package managers including Go, npm, yarn, pip, and more.

This command provides centralized cache management for:
- Language caches: Go modules, npm, yarn, pnpm, pip, poetry, cargo
- System caches: brew, apt, dnf, yum, pacman
- Build caches: gradle, maven

Examples:
  # Show cache status for all managers
  gz pm cache status

  # Clean all caches
  gz pm cache clean

  # Clean specific manager caches
  gz pm cache clean --go --npm

  # Show cache sizes
  gz pm cache size

  # Dry-run cache cleaning
  gz pm cache clean --dry-run`,
	}

	// Register subcommands
	cmd.AddCommand(newCacheStatusCmd(ctx))
	cmd.AddCommand(newCacheCleanCmd(ctx))
	cmd.AddCommand(newCacheSizeCmd(ctx))

	return cmd
}
