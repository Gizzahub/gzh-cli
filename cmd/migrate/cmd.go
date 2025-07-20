// Package migrate provides migration utilities for configuration and data transformations.
package migrate

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gizzahub/gzh-manager-go/pkg/config"
	"github.com/spf13/cobra"
)

// Options holds the configuration for the migrate command.
type Options struct {
	SourceFile string
	TargetFile string
	DryRun     bool
	Backup     bool
	Force      bool
	Verbose    bool
	Format     string
}

// NewMigrateCmd creates a new migrate command.
func NewMigrateCmd() *cobra.Command {
	opts := &Options{}

	cmd := &cobra.Command{
		Use:   "migrate [source-file] [target-file]",
		Short: "Migrate configuration files to unified format",
		Long: `Migrate configuration files from legacy formats to the unified gzh.yaml format.

The migrate command can convert legacy bulk-clone.yaml files to the new unified
gzh.yaml format. It supports dry-run mode for testing migrations, automatic
backup creation, and detailed migration reporting.

Examples:
  # Migrate a specific file
  gz migrate ./bulk-clone.yaml ./gzh.yaml

  # Migrate with dry-run mode
  gz migrate ./bulk-clone.yaml ./gzh.yaml --dry-run

  # Migrate with backup
  gz migrate ./bulk-clone.yaml ./gzh.yaml --backup

  # Auto-detect source and target files
  gz migrate --auto

  # Migrate all legacy files in current directory
  gz migrate --batch`,
		Args: cobra.RangeArgs(0, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMigrate(cmd, opts, args)
		},
	}

	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "Preview migration without making changes")
	cmd.Flags().BoolVar(&opts.Backup, "backup", true, "Create backup before migration")
	cmd.Flags().BoolVar(&opts.Force, "force", false, "Force migration even if target exists")
	cmd.Flags().BoolVar(&opts.Verbose, "verbose", false, "Enable verbose output")
	cmd.Flags().StringVar(&opts.Format, "format", "yaml", "Output format (yaml, json)")

	// Add batch and auto-detect flags
	cmd.Flags().Bool("batch", false, "Migrate all legacy files in current directory")
	cmd.Flags().Bool("auto", false, "Auto-detect source and target files")

	return cmd
}

func runMigrate(cmd *cobra.Command, opts *Options, args []string) error {
	ctx := cmd.Context()

	// Handle batch migration
	if batchMode, _ := cmd.Flags().GetBool("batch"); batchMode {
		return runBatchMigration(ctx, opts)
	}

	// Handle auto-detection
	if autoMode, _ := cmd.Flags().GetBool("auto"); autoMode {
		return runAutoMigration(ctx, opts)
	}

	// Handle specific file migration
	if len(args) < 2 {
		return fmt.Errorf("source and target files are required")
	}

	opts.SourceFile = args[0]
	opts.TargetFile = args[1]

	return runSingleMigration(ctx, opts)
}

func runSingleMigration(ctx context.Context, opts *Options) error {
	fmt.Printf("🔄 Migrating configuration: %s → %s\n", opts.SourceFile, opts.TargetFile)

	// Check if source file exists
	if _, err := os.Stat(opts.SourceFile); os.IsNotExist(err) {
		return fmt.Errorf("source file does not exist: %s", opts.SourceFile)
	}

	// Check if target file exists and handle accordingly
	if _, err := os.Stat(opts.TargetFile); err == nil && !opts.Force {
		return fmt.Errorf("target file already exists: %s (use --force to overwrite)", opts.TargetFile)
	}

	// Detect source format
	isLegacy, err := detectLegacyFormat(opts.SourceFile)
	if err != nil {
		return fmt.Errorf("failed to detect source format: %w", err)
	}

	if !isLegacy {
		fmt.Printf("✅ Source file is already in unified format: %s\n", opts.SourceFile)
		return nil
	}

	// Perform migration
	result, err := performMigration(ctx, opts.SourceFile, opts.TargetFile, opts)
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	// Display results
	displayMigrationResult(result, opts)

	return nil
}

func runBatchMigration(ctx context.Context, opts *Options) error {
	fmt.Printf("🔄 Running batch migration in current directory\n")

	// Find all legacy configuration files
	legacyFiles, err := findLegacyFiles(".")
	if err != nil {
		return fmt.Errorf("failed to find legacy files: %w", err)
	}

	if len(legacyFiles) == 0 {
		fmt.Printf("✅ No legacy configuration files found\n")
		return nil
	}

	fmt.Printf("Found %d legacy configuration files:\n", len(legacyFiles))

	for _, file := range legacyFiles {
		fmt.Printf("  - %s\n", file)
	}

	// Process each file
	successCount := 0
	failureCount := 0

	for _, sourceFile := range legacyFiles {
		targetFile := generateTargetFilename(sourceFile)
		fmt.Printf("\n🔄 Migrating: %s → %s\n", sourceFile, targetFile)

		migrateOpts := *opts
		migrateOpts.SourceFile = sourceFile
		migrateOpts.TargetFile = targetFile

		if err := runSingleMigration(ctx, &migrateOpts); err != nil {
			fmt.Printf("❌ Migration failed: %v\n", err)

			failureCount++
		} else {
			fmt.Printf("✅ Migration successful\n")

			successCount++
		}
	}

	fmt.Printf("\n📊 Batch migration completed:\n")
	fmt.Printf("  ✅ Successful: %d\n", successCount)
	fmt.Printf("  ❌ Failed: %d\n", failureCount)

	return nil
}

func runAutoMigration(ctx context.Context, opts *Options) error {
	fmt.Printf("🔄 Auto-detecting configuration files\n")

	// Look for legacy files in standard locations
	legacyFiles := []string{
		"./bulk-clone.yaml",
		"./bulk-clone.yml",
		filepath.Join(os.Getenv("HOME"), ".config/gzh-manager/bulk-clone.yaml"),
		"/etc/gzh-manager/bulk-clone.yaml",
	}

	var sourceFile string

	for _, file := range legacyFiles {
		if _, err := os.Stat(file); err == nil {
			sourceFile = file
			break
		}
	}

	if sourceFile == "" {
		fmt.Printf("✅ No legacy configuration files found for auto-migration\n")
		return nil
	}

	// Generate target filename
	targetFile := generateTargetFilename(sourceFile)

	opts.SourceFile = sourceFile
	opts.TargetFile = targetFile

	fmt.Printf("📝 Auto-detected migration: %s → %s\n", sourceFile, targetFile)

	return runSingleMigration(ctx, opts)
}

func performMigration(_ context.Context, sourceFile, targetFile string, opts *Options) (*config.MigrationResult, error) {
	if opts.DryRun {
		fmt.Printf("🧪 Dry-run mode: previewing migration\n")
	}

	// Create backup if requested
	if opts.Backup && !opts.DryRun {
		backupFile := createBackupFilename(sourceFile)
		if err := copyFile(sourceFile, backupFile); err != nil {
			return nil, fmt.Errorf("failed to create backup: %w", err)
		}

		fmt.Printf("💾 Backup created: %s\n", backupFile)
	}

	// Use the existing migration functionality
	result, err := config.MigrateConfigFile(sourceFile, targetFile, opts.DryRun)
	if err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	return result, nil
}

func displayMigrationResult(result *config.MigrationResult, opts *Options) {
	fmt.Printf("\n📊 Migration Results:\n")
	fmt.Printf("  📁 Source: %s\n", result.SourcePath)
	fmt.Printf("  📁 Target: %s\n", result.TargetPath)
	fmt.Printf("  ✅ Success: %v\n", result.Success)

	if result.BackupPath != "" {
		fmt.Printf("  💾 Backup: %s\n", result.BackupPath)
	}

	if len(result.Warnings) > 0 {
		fmt.Printf("  ⚠️  Warnings:\n")

		for _, warning := range result.Warnings {
			fmt.Printf("    - %s\n", warning)
		}
	}

	if len(result.RequiredActions) > 0 {
		fmt.Printf("  🔧 Required Actions:\n")

		for _, action := range result.RequiredActions {
			fmt.Printf("    - %s\n", action)
		}
	}

	if opts.Verbose && result.MigrationReport != "" {
		fmt.Printf("  📊 Migration Statistics:\n")
		fmt.Printf("    - Migrated targets: %d\n", result.MigratedTargets)
		fmt.Printf("    - Migration report:\n%s\n", result.MigrationReport)
	}
}

func detectLegacyFormat(filename string) (bool, error) {
	// Use the existing detection logic from config package
	return config.DetectLegacyFormat(filename)
}

func findLegacyFiles(dir string) ([]string, error) {
	var legacyFiles []string

	patterns := []string{
		"bulk-clone.yaml",
		"bulk-clone.yml",
	}

	for _, pattern := range patterns {
		matches, err := filepath.Glob(filepath.Join(dir, pattern))
		if err != nil {
			return nil, err
		}

		legacyFiles = append(legacyFiles, matches...)
	}

	return legacyFiles, nil
}

func generateTargetFilename(sourceFile string) string {
	dir := filepath.Dir(sourceFile)

	// Convert bulk-clone.yaml to gzh.yaml
	if strings.Contains(sourceFile, "bulk-clone") {
		return filepath.Join(dir, "gzh.yaml")
	}

	// Default transformation
	ext := filepath.Ext(sourceFile)
	base := strings.TrimSuffix(filepath.Base(sourceFile), ext)

	return filepath.Join(dir, base+"-unified"+ext)
}

func createBackupFilename(sourceFile string) string {
	timestamp := time.Now().Format("20060102-150405")
	ext := filepath.Ext(sourceFile)
	base := strings.TrimSuffix(sourceFile, ext)

	return fmt.Sprintf("%s.backup.%s%s", base, timestamp, ext)
}

func copyFile(src, dst string) error {
	input, err := os.ReadFile(src) //nolint:gosec // File paths are validated by caller
	if err != nil {
		return err
	}

	return os.WriteFile(dst, input, 0o600) // More secure file permissions
}
