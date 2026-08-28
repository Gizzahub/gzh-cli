// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package devenv

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// EnhancedSSHOptions represents options for enhanced SSH commands.
type EnhancedSSHOptions struct {
	Name          string
	Description   string
	ConfigPath    string
	StorePath     string
	Force         bool
	ListAll       bool
	IncludeKeys   bool
	IncludePublic bool
}

// EnhancedSSHMetadata represents metadata for saved SSH configurations.
type EnhancedSSHMetadata struct {
	Description  string    `json:"description"`
	SavedAt      time.Time `json:"saved_at"`
	SourcePath   string    `json:"source_path"`
	IncludeFiles []string  `json:"include_files"`
	PrivateKeys  []string  `json:"private_keys"`
	PublicKeys   []string  `json:"public_keys"`
	HasIncludes  bool      `json:"has_includes"`
	HasKeys      bool      `json:"has_keys"`
}

// EnhancedSSHCommand provides enhanced SSH configuration management.
type EnhancedSSHCommand struct{}

// NewEnhancedSSHCommand creates a new enhanced SSH command instance.
func NewEnhancedSSHCommand() *EnhancedSSHCommand {
	return &EnhancedSSHCommand{}
}

// DefaultEnhancedOptions returns default options for enhanced SSH commands.
func (c *EnhancedSSHCommand) DefaultEnhancedOptions() *EnhancedSSHOptions {
	homeDir, _ := os.UserHomeDir()

	return &EnhancedSSHOptions{
		ConfigPath:    filepath.Join(homeDir, ".ssh", "config"),
		StorePath:     filepath.Join(homeDir, ".gz", "ssh-configs"),
		IncludeKeys:   true,
		IncludePublic: true,
	}
}

// CreateEnhancedSaveCommand creates the enhanced save command.
func (c *EnhancedSSHCommand) CreateEnhancedSaveCommand() *cobra.Command {
	opts := c.DefaultEnhancedOptions()

	cmd := &cobra.Command{
		Use:   "save",
		Short: "Save current SSH configuration with includes and keys",
		Long: `Save current SSH configuration including:
- Main SSH config file
- All files referenced by Include directives
- All private keys referenced by IdentityFile directives
- Corresponding public keys (optional)

The configuration is saved as a directory structure to preserve
relative paths and file relationships.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.SaveEnhancedConfig(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Name, "name", "n", "", "Name for the saved configuration (required)")
	cmd.Flags().StringVarP(&opts.Description, "description", "d", "", "Description for the saved configuration")
	cmd.Flags().StringVar(&opts.ConfigPath, "config-path", opts.ConfigPath, "Path to SSH config file")
	cmd.Flags().StringVar(&opts.StorePath, "store-path", opts.StorePath, "Path to store saved configurations")
	cmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "Overwrite existing configuration")
	cmd.Flags().BoolVar(&opts.IncludeKeys, "include-keys", opts.IncludeKeys, "Include private keys")
	cmd.Flags().BoolVar(&opts.IncludePublic, "include-public", opts.IncludePublic, "Include public keys")

	cmd.MarkFlagRequired("name")

	return cmd
}

// CreateEnhancedLoadCommand creates the enhanced load command.
func (c *EnhancedSSHCommand) CreateEnhancedLoadCommand() *cobra.Command {
	opts := c.DefaultEnhancedOptions()

	cmd := &cobra.Command{
		Use:   "load",
		Short: "Load saved SSH configuration with includes and keys",
		Long: `Load saved SSH configuration including:
- Main SSH config file
- All included configuration files
- All private and public keys

The configuration is restored maintaining the original directory structure.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.LoadEnhancedConfig(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Name, "name", "n", "", "Name of the configuration to load (required)")
	cmd.Flags().StringVar(&opts.ConfigPath, "config-path", opts.ConfigPath, "Path to SSH config file")
	cmd.Flags().StringVar(&opts.StorePath, "store-path", opts.StorePath, "Path to stored configurations")
	cmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "Replace existing regular files (symlinks and non-regular files are refused)")

	cmd.MarkFlagRequired("name")

	return cmd
}

// CreateEnhancedListCommand creates the enhanced list command.
func (c *EnhancedSSHCommand) CreateEnhancedListCommand() *cobra.Command {
	opts := c.DefaultEnhancedOptions()

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List saved SSH configurations with details",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.ListEnhancedConfigs(opts)
		},
	}

	cmd.Flags().StringVar(&opts.StorePath, "store-path", opts.StorePath, "Path to stored configurations")
	cmd.Flags().BoolVarP(&opts.ListAll, "all", "a", false, "Show detailed information for all configurations")

	return cmd
}

// SaveEnhancedConfig saves the SSH configuration with includes and keys.
func (c *EnhancedSSHCommand) SaveEnhancedConfig(opts *EnhancedSSHOptions) error {
	// Validate inputs
	if opts.Name == "" {
		return fmt.Errorf("configuration name is required")
	}

	// Check if source config exists
	if _, err := os.Stat(opts.ConfigPath); os.IsNotExist(err) {
		return fmt.Errorf("SSH config file not found at %s", opts.ConfigPath)
	}

	// Parse SSH configuration
	parser := NewSSHConfigParser(opts.ConfigPath)
	parsed, err := parser.Parse()
	if err != nil {
		return fmt.Errorf("failed to parse SSH configuration: %w", err)
	}

	// Check if config already exists
	configDir := filepath.Join(opts.StorePath, opts.Name)
	if _, err := os.Stat(configDir); err == nil && !opts.Force {
		return fmt.Errorf("configuration '%s' already exists (use --force to overwrite)", opts.Name)
	}

	// Create store directory structure
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Save main config file
	mainConfigDest := filepath.Join(configDir, "config")
	if err := c.copyFile(parsed.MainConfigPath, mainConfigDest); err != nil {
		return fmt.Errorf("failed to save main config: %w", err)
	}

	// Save include files
	if len(parsed.IncludeFiles) > 0 {
		includeDir := filepath.Join(configDir, "includes")
		if err := os.MkdirAll(includeDir, 0o755); err != nil {
			return fmt.Errorf("failed to create includes directory: %w", err)
		}

		for i, includeFile := range parsed.IncludeFiles {
			destName := fmt.Sprintf("include_%d_%s", i, filepath.Base(includeFile))
			destPath := filepath.Join(includeDir, destName)
			if err := c.copyFile(includeFile, destPath); err != nil {
				fmt.Printf("Warning: failed to copy include file %s: %v\n", includeFile, err)
			}
		}
	}

	// Save private keys
	if opts.IncludeKeys && len(parsed.PrivateKeys) > 0 {
		keysDir := filepath.Join(configDir, "keys")
		if err := os.MkdirAll(keysDir, 0o700); err != nil {
			return fmt.Errorf("failed to create keys directory: %w", err)
		}

		for _, keyFile := range parsed.PrivateKeys {
			destName := filepath.Base(keyFile)
			destPath := filepath.Join(keysDir, destName)
			if err := c.copyFile(keyFile, destPath); err != nil {
				fmt.Printf("Warning: failed to copy private key %s: %v\n", keyFile, err)
			} else {
				// Set proper permissions for private keys
				os.Chmod(destPath, 0o600)
			}
		}
	}

	// Save public keys
	if opts.IncludePublic && len(parsed.PublicKeys) > 0 {
		keysDir := filepath.Join(configDir, "keys")
		if err := os.MkdirAll(keysDir, 0o755); err != nil {
			return fmt.Errorf("failed to create keys directory: %w", err)
		}

		for _, keyFile := range parsed.PublicKeys {
			destName := filepath.Base(keyFile)
			destPath := filepath.Join(keysDir, destName)
			if err := c.copyFile(keyFile, destPath); err != nil {
				fmt.Printf("Warning: failed to copy public key %s: %v\n", keyFile, err)
			}
		}
	}

	// Save metadata
	metadata := EnhancedSSHMetadata{
		Description:  opts.Description,
		SavedAt:      time.Now(),
		SourcePath:   opts.ConfigPath,
		IncludeFiles: parsed.IncludeFiles,
		PrivateKeys:  parsed.PrivateKeys,
		PublicKeys:   parsed.PublicKeys,
		HasIncludes:  len(parsed.IncludeFiles) > 0,
		HasKeys:      len(parsed.PrivateKeys) > 0,
	}

	metadataFile := filepath.Join(configDir, "metadata.json")
	if err := c.saveEnhancedMetadata(metadataFile, metadata); err != nil {
		fmt.Printf("Warning: failed to save metadata: %v\n", err)
	}

	// Print summary
	fmt.Printf("✅ SSH configuration '%s' saved successfully\n", opts.Name)
	if opts.Description != "" {
		fmt.Printf("   Description: %s\n", opts.Description)
	}
	fmt.Printf("   Main config: %s\n", parsed.MainConfigPath)
	fmt.Printf("   Include files: %d\n", len(parsed.IncludeFiles))
	if opts.IncludeKeys {
		fmt.Printf("   Private keys: %d\n", len(parsed.PrivateKeys))
	}
	if opts.IncludePublic {
		fmt.Printf("   Public keys: %d\n", len(parsed.PublicKeys))
	}
	fmt.Printf("   Saved to: %s\n", configDir)

	return nil
}

// LoadEnhancedConfig loads a saved SSH configuration.
func (c *EnhancedSSHCommand) LoadEnhancedConfig(opts *EnhancedSSHOptions) error {
	// Validate inputs
	if opts.Name == "" {
		return fmt.Errorf("configuration name is required")
	}

	// Check if saved config exists
	configDir := filepath.Join(opts.StorePath, opts.Name)
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		return fmt.Errorf("configuration '%s' not found", opts.Name)
	}

	// Load metadata
	metadataFile := filepath.Join(configDir, "metadata.json")
	metadata, err := c.loadEnhancedMetadata(metadataFile)
	if err != nil {
		return fmt.Errorf("failed to load metadata: %w", err)
	}

	// Create target directory
	if err := os.MkdirAll(filepath.Dir(opts.ConfigPath), 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Load main config file
	mainConfigSrc := filepath.Join(configDir, "config")
	if err := c.restoreEnhancedFile(mainConfigSrc, opts.ConfigPath, 0o600, opts.Force); err != nil {
		return fmt.Errorf("failed to load main config: %w", err)
	}

	loadedFiles := 1

	// Load include files (this is tricky - we need to restore them to their original paths)
	includeDir := filepath.Join(configDir, "includes")
	if _, err := os.Stat(includeDir); err == nil {
		entries, err := os.ReadDir(includeDir)
		if err != nil {
			return fmt.Errorf("failed to read includes directory: %w", err)
		}

		// Include files are restored under the target configuration directory.
		configD := filepath.Join(filepath.Dir(opts.ConfigPath), "config.d")
		if err := os.MkdirAll(configD, 0o755); err != nil {
			return fmt.Errorf("failed to create includes directory: %w", err)
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			srcPath := filepath.Join(includeDir, entry.Name())
			// Remove the prefix added during save.
			originalName := strings.TrimPrefix(entry.Name(), "include_")
			if idx := strings.Index(originalName, "_"); idx > 0 {
				originalName = originalName[idx+1:]
			}
			destPath := filepath.Join(configD, originalName)
			if err := c.restoreEnhancedFile(srcPath, destPath, 0o600, opts.Force); err != nil {
				return fmt.Errorf("failed to load include file %s: %w", entry.Name(), err)
			}
			loadedFiles++
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to inspect includes directory: %w", err)
	}

	// Load keys
	keysDir := filepath.Join(configDir, "keys")
	if _, err := os.Stat(keysDir); err == nil {
		sshKeysDir := filepath.Dir(opts.ConfigPath)
		entries, err := os.ReadDir(keysDir)
		if err != nil {
			return fmt.Errorf("failed to read keys directory: %w", err)
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			srcPath := filepath.Join(keysDir, entry.Name())
			destPath := filepath.Join(sshKeysDir, entry.Name())
			mode := os.FileMode(0o600)
			if strings.HasSuffix(entry.Name(), ".pub") {
				mode = 0o644 //nolint:gosec // Public SSH keys must remain readable by non-owner tools.
			}
			if err := c.restoreEnhancedFile(srcPath, destPath, mode, opts.Force); err != nil {
				return fmt.Errorf("failed to load key %s: %w", entry.Name(), err)
			}
			loadedFiles++
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to inspect keys directory: %w", err)
	}

	// Print summary
	fmt.Printf("✅ SSH configuration '%s' loaded successfully\n", opts.Name)
	if metadata.Description != "" {
		fmt.Printf("   Description: %s\n", metadata.Description)
	}
	fmt.Printf("   Originally saved: %s\n", metadata.SavedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("   Files restored: %d\n", loadedFiles)
	fmt.Printf("   Loaded to: %s\n", opts.ConfigPath)

	return nil
}

// restoreEnhancedFile copies one saved SSH file into its final destination.
// It intentionally validates only the final component; ancestor no-following and
// multi-file rollback are outside the compatibility contract of enhanced load.
func (c *EnhancedSSHCommand) restoreEnhancedFile(src, dst string, mode os.FileMode, force bool) (err error) {
	if mode != 0o600 && mode != 0o644 {
		return fmt.Errorf("unsupported restore mode %04o", mode)
	}

	if info, lstatErr := os.Lstat(dst); lstatErr == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("refusing symlink destination %s", dst)
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("refusing non-regular destination %s", dst)
		}
		if !force {
			return fmt.Errorf("destination already exists at %s (use --force to replace a regular file)", dst)
		}
	} else if !os.IsNotExist(lstatErr) {
		return fmt.Errorf("failed to inspect destination %s: %w", dst, lstatErr)
	}

	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source %s: %w", src, err)
	}

	tempFile, err := os.CreateTemp(filepath.Dir(dst), ".gzh-ssh-restore-*")
	if err != nil {
		if closeErr := sourceFile.Close(); closeErr != nil {
			return fmt.Errorf("failed to create restore temp for %s: %w (also failed to close source: %w)", dst, err, closeErr)
		}
		return fmt.Errorf("failed to create restore temp for %s: %w", dst, err)
	}
	tempPath := tempFile.Name()
	published := false
	defer func() {
		if !published {
			if removeErr := os.Remove(tempPath); removeErr != nil && !os.IsNotExist(removeErr) && err == nil {
				err = fmt.Errorf("failed to remove restore temp %s: %w", tempPath, removeErr)
			}
		}
	}()

	if err = tempFile.Chmod(0o600); err != nil {
		_ = sourceFile.Close()
		_ = tempFile.Close()
		return fmt.Errorf("failed to secure restore temp %s: %w", tempPath, err)
	}
	if _, err = io.Copy(tempFile, sourceFile); err != nil {
		_ = sourceFile.Close()
		_ = tempFile.Close()
		return fmt.Errorf("failed to copy %s to restore temp: %w", src, err)
	}
	if closeErr := sourceFile.Close(); closeErr != nil {
		_ = tempFile.Close()
		return fmt.Errorf("failed to close source %s: %w", src, closeErr)
	}
	if err = tempFile.Chmod(mode); err != nil {
		_ = tempFile.Close()
		return fmt.Errorf("failed to set restore mode for %s: %w", dst, err)
	}
	if err = tempFile.Close(); err != nil {
		return fmt.Errorf("failed to close restore temp %s: %w", tempPath, err)
	}

	if force {
		if err = os.Rename(tempPath, dst); err != nil {
			return fmt.Errorf("failed to replace destination %s: %w", dst, err)
		}
		published = true
		return nil
	}

	if err = os.Link(tempPath, dst); err != nil {
		return fmt.Errorf("failed to publish destination %s without replacement: %w", dst, err)
	}
	published = true
	if err = os.Remove(tempPath); err != nil {
		return fmt.Errorf("published destination %s but failed to remove restore temp %s: %w", dst, tempPath, err)
	}

	return nil
}

// ListEnhancedConfigs lists saved SSH configurations.
func (c *EnhancedSSHCommand) ListEnhancedConfigs(opts *EnhancedSSHOptions) error {
	// Check if store directory exists
	if _, err := os.Stat(opts.StorePath); os.IsNotExist(err) {
		fmt.Printf("No SSH configurations found (store directory doesn't exist)\n")
		return nil
	}

	// Read directory contents
	entries, err := os.ReadDir(opts.StorePath)
	if err != nil {
		return fmt.Errorf("failed to read store directory: %w", err)
	}

	// Filter for directories
	var configs []string
	for _, entry := range entries {
		if entry.IsDir() {
			// Check if it has metadata.json
			metadataPath := filepath.Join(opts.StorePath, entry.Name(), "metadata.json")
			if _, err := os.Stat(metadataPath); err == nil {
				configs = append(configs, entry.Name())
			}
		}
	}

	if len(configs) == 0 {
		fmt.Printf("No SSH configurations found\n")
		return nil
	}

	fmt.Printf("Saved SSH configurations:\n\n")

	for _, configName := range configs {
		if opts.ListAll {
			c.printDetailedEnhancedConfig(opts.StorePath, configName)
		} else {
			fmt.Printf("  • %s\n", configName)
		}
	}

	if !opts.ListAll {
		fmt.Printf("\nUse --all to show detailed information\n")
	}

	return nil
}

// printDetailedEnhancedConfig prints detailed information about a configuration.
func (c *EnhancedSSHCommand) printDetailedEnhancedConfig(storePath, configName string) {
	fmt.Printf("  📁 %s\n", configName)

	metadataFile := filepath.Join(storePath, configName, "metadata.json")
	if metadata, err := c.loadEnhancedMetadata(metadataFile); err == nil {
		if metadata.Description != "" {
			fmt.Printf("     Description: %s\n", metadata.Description)
		}
		fmt.Printf("     Saved: %s\n", metadata.SavedAt.Format("2006-01-02 15:04:05"))
		if metadata.SourcePath != "" {
			fmt.Printf("     Source: %s\n", metadata.SourcePath)
		}
		if metadata.HasIncludes {
			fmt.Printf("     Include files: %d\n", len(metadata.IncludeFiles))
		}
		if metadata.HasKeys {
			fmt.Printf("     Private keys: %d\n", len(metadata.PrivateKeys))
			fmt.Printf("     Public keys: %d\n", len(metadata.PublicKeys))
		}
	}

	configDir := filepath.Join(storePath, configName)
	if entries, err := os.ReadDir(configDir); err == nil {
		totalSize := int64(0)
		for _, entry := range entries {
			if entry.Name() != "metadata.json" {
				if info, err := entry.Info(); err == nil {
					totalSize += info.Size()
				}
			}
		}
		fmt.Printf("     Total size: %d bytes\n", totalSize)
	}

	fmt.Println()
}

// copyFile copies a file from src to dst.
func (c *EnhancedSSHCommand) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// saveEnhancedMetadata saves enhanced metadata to a JSON file.
func (c *EnhancedSSHCommand) saveEnhancedMetadata(filename string, metadata EnhancedSSHMetadata) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	//gosec:disable G117 -- PrivateKeys는 키 내용이 아니라 원본 파일 경로만 포함한다.
	return encoder.Encode(metadata)
}

// loadEnhancedMetadata loads enhanced metadata from a JSON file.
func (c *EnhancedSSHCommand) loadEnhancedMetadata(filename string) (*EnhancedSSHMetadata, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var metadata EnhancedSSHMetadata
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&metadata)
	return &metadata, err
}
