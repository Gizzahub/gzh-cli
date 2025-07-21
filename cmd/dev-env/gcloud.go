// Copyright (c) 2025 Archmagece
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

type gcloudOptions struct {
	name        string
	description string
	configPath  string
	storePath   string
	force       bool
	listAll     bool
}

func defaultGcloudOptions() *gcloudOptions {
	homeDir, _ := os.UserHomeDir()

	return &gcloudOptions{
		configPath: filepath.Join(homeDir, ".config", "gcloud"),
		storePath:  filepath.Join(homeDir, ".gz", "gcloud-configs"),
	}
}

func newGcloudCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gcloud",
		Short: "Manage Google Cloud configuration",
		Long: `Save and load Google Cloud configuration.

This command helps you backup and restore Google Cloud configuration, which contains:
- Active configuration name
- Project settings
- Account information
- Default regions and zones
- Other Google Cloud SDK settings

This is useful when:
- Setting up new development machines
- Switching between different GCP environments
- Backing up Google Cloud configurations before changes
- Managing multiple GCP configurations for different projects

The configurations are saved to ~/.gz/gcloud-configs/ by default.

Examples:
  # Save current gcloud config with a name
  gz dev-env gcloud save --name production
  
  # Save with description
  gz dev-env gcloud save --name staging --description "Staging GCP config"
  
  # Load a saved gcloud config
  gz dev-env gcloud load --name production
  
  # List all saved configurations
  gz dev-env gcloud list
  
  # Save from specific path
  gz dev-env gcloud save --name custom --config-path /path/to/gcloud`,
		SilenceUsage: true,
	}

	cmd.AddCommand(newGcloudSaveCmd())
	cmd.AddCommand(newGcloudLoadCmd())
	cmd.AddCommand(newGcloudListCmd())

	return cmd
}

func newGcloudSaveCmd() *cobra.Command {
	o := defaultGcloudOptions()

	cmd := &cobra.Command{
		Use:   "save",
		Short: "Save current gcloud config",
		Long: `Save the current Google Cloud configuration with a given name.

This creates a backup of your current Google Cloud configuration that can be
restored later using the 'load' command. The configuration includes
project settings, account information, and other GCP SDK settings.

Examples:
  # Save current gcloud config as "production"
  gz dev-env gcloud save --name production
  
  # Save with description
  gz dev-env gcloud save --name staging --description "Staging environment"
  
  # Save from specific path
  gz dev-env gcloud save --name custom --config-path /path/to/gcloud`,
		RunE: o.runSave,
	}

	cmd.Flags().StringVarP(&o.name, "name", "n", "", "Name for the saved configuration (required)")
	cmd.Flags().StringVarP(&o.description, "description", "d", "", "Description for the configuration")
	cmd.Flags().StringVar(&o.configPath, "config-path", o.configPath, "Path to gcloud config directory to save")
	cmd.Flags().StringVar(&o.storePath, "store-path", o.storePath, "Directory to store saved configurations")
	cmd.Flags().BoolVarP(&o.force, "force", "f", false, "Overwrite existing saved configuration")

	cmd.MarkFlagRequired("name")

	return cmd
}

func newGcloudLoadCmd() *cobra.Command {
	o := defaultGcloudOptions()

	cmd := &cobra.Command{
		Use:   "load",
		Short: "Load a saved gcloud config",
		Long: `Load a previously saved Google Cloud configuration.

This restores a Google Cloud configuration that was previously saved using the 'save' command.
The current gcloud config will be backed up before loading the new one.

Examples:
  # Load the "production" configuration
  gz dev-env gcloud load --name production
  
  # Load to specific path
  gz dev-env gcloud load --name staging --config-path /path/to/gcloud`,
		RunE: o.runLoad,
	}

	cmd.Flags().StringVarP(&o.name, "name", "n", "", "Name of the saved configuration to load (required)")
	cmd.Flags().StringVar(&o.configPath, "config-path", o.configPath, "Path where to load the gcloud config")
	cmd.Flags().StringVar(&o.storePath, "store-path", o.storePath, "Directory where saved configurations are stored")
	cmd.Flags().BoolVarP(&o.force, "force", "f", false, "Skip backup of current configuration")

	cmd.MarkFlagRequired("name")

	return cmd
}

func newGcloudListCmd() *cobra.Command {
	o := defaultGcloudOptions()

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List saved gcloud configs",
		Long: `List all saved Google Cloud configuration directories.

This shows all the gcloud configurations that have been saved using the 'save' command,
along with their descriptions, save dates, and configuration information.

Examples:
  # List all saved configurations
  gz dev-env gcloud list`,
		RunE: o.runList,
	}

	cmd.Flags().StringVar(&o.storePath, "store-path", o.storePath, "Directory where saved configurations are stored")

	return cmd
}

func (o *gcloudOptions) runSave(_ *cobra.Command, args []string) error {
	// Check if source config directory exists
	if _, err := os.Stat(o.configPath); os.IsNotExist(err) {
		return fmt.Errorf("gcloud config directory not found at %s", o.configPath)
	}

	// Create store directory if it doesn't exist
	if err := os.MkdirAll(o.storePath, 0o755); err != nil {
		return fmt.Errorf("failed to create store directory: %w", err)
	}

	// Check if name already exists
	savedPath := filepath.Join(o.storePath, o.name)
	if _, err := os.Stat(savedPath); err == nil && !o.force {
		return fmt.Errorf("configuration '%s' already exists. Use --force to overwrite", o.name)
	}

	// Remove existing directory if force is enabled
	if o.force {
		if err := os.RemoveAll(savedPath); err != nil {
			return fmt.Errorf("failed to remove existing configuration: %w", err)
		}
	}

	// Copy the gcloud config directory
	if err := o.copyDir(o.configPath, savedPath); err != nil {
		return fmt.Errorf("failed to save gcloud config: %w", err)
	}

	// Save metadata
	if err := o.saveMetadata(); err != nil {
		fmt.Printf("Warning: failed to save metadata: %v\n", err)
	}

	// Display config information
	if err := o.displayConfigInfo(savedPath); err != nil {
		fmt.Printf("Warning: failed to read config info: %v\n", err)
	}

	fmt.Printf("✅ Gcloud config saved as '%s'\n", o.name)

	if o.description != "" {
		fmt.Printf("   Description: %s\n", o.description)
	}

	fmt.Printf("   Saved to: %s\n", savedPath)

	return nil
}

func (o *gcloudOptions) runLoad(_ *cobra.Command, args []string) error {
	// Check if saved config exists
	savedPath := filepath.Join(o.storePath, o.name)
	if _, err := os.Stat(savedPath); os.IsNotExist(err) {
		return fmt.Errorf("saved configuration '%s' not found", o.name)
	}

	// Backup current config if it exists and force is not set
	if !o.force {
		if _, err := os.Stat(o.configPath); err == nil {
			backupPath := o.configPath + ".backup." + time.Now().Format("20060102-150405")
			if err := o.copyDir(o.configPath, backupPath); err != nil {
				return fmt.Errorf("failed to backup current gcloud config: %w", err)
			}

			fmt.Printf("📦 Current gcloud config backed up to: %s\n", backupPath)
		}
	}

	// Remove current config directory
	if err := os.RemoveAll(o.configPath); err != nil {
		return fmt.Errorf("failed to remove current gcloud config: %w", err)
	}

	// Copy the saved config to target location
	if err := o.copyDir(savedPath, o.configPath); err != nil {
		return fmt.Errorf("failed to load gcloud config: %w", err)
	}

	// Display config information
	if err := o.displayConfigInfo(o.configPath); err != nil {
		fmt.Printf("Warning: failed to read config info: %v\n", err)
	}

	fmt.Printf("✅ Gcloud config '%s' loaded successfully\n", o.name)
	fmt.Printf("   Loaded to: %s\n", o.configPath)

	return nil
}

func (o *gcloudOptions) runList(_ *cobra.Command, args []string) error {
	// Check if store directory exists
	if _, err := os.Stat(o.storePath); os.IsNotExist(err) {
		fmt.Println("No saved gcloud configs found.")
		return nil
	}

	// Read directory contents
	entries, err := os.ReadDir(o.storePath)
	if err != nil {
		return fmt.Errorf("failed to read store directory: %w", err)
	}

	// Filter for directories (excluding .meta files)
	var configs []string

	for _, entry := range entries {
		if entry.IsDir() {
			configs = append(configs, entry.Name())
		}
	}

	if len(configs) == 0 {
		fmt.Println("No saved gcloud configs found.")
		return nil
	}

	fmt.Printf("Saved gcloud configs (%d):\n\n", len(configs))

	// Read and display metadata for each config
	for _, name := range configs {
		metadata := o.loadMetadata(name)
		fmt.Printf("☁️ %s\n", name)

		if metadata.Description != "" {
			fmt.Printf("   Description: %s\n", metadata.Description)
		}

		if !metadata.SavedAt.IsZero() {
			fmt.Printf("   Saved: %s\n", metadata.SavedAt.Format("2006-01-02 15:04:05"))
		}

		configPath := filepath.Join(o.storePath, name)
		if info, err := os.Stat(configPath); err == nil && info.IsDir() {
			// Get directory size
			size, err := o.getDirSize(configPath)
			if err == nil {
				fmt.Printf("   Size: %d bytes\n", size)
			}
		}

		// Display configuration information
		if err := o.displayConfigInfo(configPath); err == nil {
			// Already displayed in displayConfigInfo
		}

		fmt.Println()
	}

	return nil
}

type gcloudMetadata struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	SavedAt     time.Time `json:"saved_at"`
	SourcePath  string    `json:"source_path"`
}

func (o *gcloudOptions) saveMetadata() error {
	metadata := gcloudMetadata{
		Name:        o.name,
		Description: o.description,
		SavedAt:     time.Now(),
		SourcePath:  o.configPath,
	}

	metadataPath := filepath.Join(o.storePath, o.name+".meta")

	file, err := os.Create(metadataPath)
	if err != nil {
		return err
	}

	defer file.Close()

	// Write metadata as simple key-value pairs
	if metadata.Description != "" {
		fmt.Fprintf(file, "description=%s\n", metadata.Description)
	}

	fmt.Fprintf(file, "saved_at=%s\n", metadata.SavedAt.Format(time.RFC3339))
	fmt.Fprintf(file, "source_path=%s\n", metadata.SourcePath)

	return nil
}

func (o *gcloudOptions) loadMetadata(name string) gcloudMetadata {
	metadata := gcloudMetadata{Name: name}

	metadataPath := filepath.Join(o.storePath, name+".meta")

	content, err := os.ReadFile(metadataPath)
	if err != nil {
		return metadata
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key, value := parts[0], parts[1]
		switch key {
		case "description":
			metadata.Description = value
		case "saved_at":
			if t, err := time.Parse(time.RFC3339, value); err == nil {
				metadata.SavedAt = t
			}
		case "source_path":
			metadata.SourcePath = value
		}
	}

	return metadata
}

func (o *gcloudOptions) copyDir(src, dst string) error {
	// Get source directory info
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// Create destination directory
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	// Read source directory entries
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	// Copy each entry
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Recursively copy subdirectory
			if err := o.copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Copy file
			if err := o.copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func (o *gcloudOptions) copyFile(src, dst string) error {
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
	if err != nil {
		return err
	}

	// Copy file permissions
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, sourceInfo.Mode())
}

func (o *gcloudOptions) getDirSize(path string) (int64, error) {
	var size int64

	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			size += info.Size()
		}

		return nil
	})

	return size, err
}

func (o *gcloudOptions) displayConfigInfo(configPath string) error {
	// Try to read gcloud active configuration
	activeConfigPath := filepath.Join(configPath, "active_config")
	if activeConfig, err := os.ReadFile(activeConfigPath); err == nil {
		fmt.Printf("   Active config: %s\n", strings.TrimSpace(string(activeConfig)))
	}

	// Try to read configurations
	configurationsPath := filepath.Join(configPath, "configurations")
	if _, err := os.Stat(configurationsPath); err == nil {
		configs, err := o.parseGcloudConfigurations(configurationsPath)
		if err == nil && len(configs) > 0 {
			fmt.Printf("   Configurations: %d found\n", len(configs))

			for _, config := range configs {
				fmt.Printf("     - %s", config.Name)

				if config.Project != "" {
					fmt.Printf(" (project: %s)", config.Project)
				}

				if config.Account != "" {
					fmt.Printf(" (account: %s)", config.Account)
				}

				if config.Region != "" {
					fmt.Printf(" (region: %s)", config.Region)
				}

				fmt.Println()
			}
		}
	}

	return nil
}

type gcloudConfiguration struct {
	Name    string
	Project string
	Account string
	Region  string
	Zone    string
}

func (o *gcloudOptions) parseGcloudConfigurations(configurationsPath string) ([]gcloudConfiguration, error) {
	var configurations []gcloudConfiguration

	// Read configuration directories
	entries, err := os.ReadDir(configurationsPath)
	if err != nil {
		return configurations, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		configName := entry.Name()
		configPath := filepath.Join(configurationsPath, configName, "properties")

		// Try to read the properties file
		if _, err := os.Stat(configPath); err != nil {
			continue
		}

		config := gcloudConfiguration{Name: configName}

		// Parse properties file
		if err := o.parseGcloudProperties(configPath, &config); err == nil {
			configurations = append(configurations, config)
		}
	}

	return configurations, nil
}

func (o *gcloudOptions) parseGcloudProperties(propertiesPath string, config *gcloudConfiguration) error {
	content, err := os.ReadFile(propertiesPath)
	if err != nil {
		return err
	}

	// Try to parse as JSON (newer gcloud format)
	var properties map[string]interface{}
	if err := json.Unmarshal(content, &properties); err == nil {
		// JSON format
		if core, ok := properties["core"].(map[string]interface{}); ok {
			if project, ok := core["project"].(string); ok {
				config.Project = project
			}

			if account, ok := core["account"].(string); ok {
				config.Account = account
			}
		}

		if compute, ok := properties["compute"].(map[string]interface{}); ok {
			if region, ok := compute["region"].(string); ok {
				config.Region = region
			}

			if zone, ok := compute["zone"].(string); ok {
				config.Zone = zone
			}
		}
	} else {
		// Try INI format (older gcloud format)
		lines := strings.Split(string(content), "\n")

		var currentSection string

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			// Section header
			if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
				currentSection = strings.Trim(line, "[]")
				continue
			}

			// Key-value pair
			if strings.Contains(line, "=") {
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					value := strings.TrimSpace(parts[1])

					switch currentSection + "." + key {
					case "core.project":
						config.Project = value
					case "core.account":
						config.Account = value
					case "compute.region":
						config.Region = value
					case "compute.zone":
						config.Zone = value
					}
				}
			}
		}
	}

	return nil
}
