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

const (
	metadataKeyDescription = "description"
	metadataKeySavedAt     = "saved_at"
	metadataKeySourcePath  = "source_path"
)

// BaseOptions represents common options for dev-env commands.
type BaseOptions struct {
	Name        string
	Description string
	ConfigPath  string
	StorePath   string
	Force       bool
	ListAll     bool
}

// ConfigMetadata represents metadata for saved configurations.
type ConfigMetadata struct {
	Description string    `json:"description"`
	SavedAt     time.Time `json:"saved_at"`
	SourcePath  string    `json:"source_path"`
}

// BaseCommand provides common functionality for dev-env commands.
type BaseCommand struct {
	serviceName    string
	configFileName string
	defaultConfig  string
	description    string
	examples       []string
}

// NewBaseCommand creates a new base command instance.
func NewBaseCommand(serviceName, configFileName, defaultConfig, description string, examples []string) *BaseCommand {
	return &BaseCommand{
		serviceName:    serviceName,
		configFileName: configFileName,
		defaultConfig:  defaultConfig,
		description:    description,
		examples:       examples,
	}
}

// DefaultOptions returns default options for the service.
func (bc *BaseCommand) DefaultOptions() *BaseOptions {
	homeDir, _ := os.UserHomeDir()

	return &BaseOptions{
		ConfigPath: filepath.Join(homeDir, bc.defaultConfig),
		StorePath:  filepath.Join(homeDir, ".gz", bc.serviceName+"-configs"),
	}
}

// CreateMainCommand creates the main command for the service.
func (bc *BaseCommand) CreateMainCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          bc.serviceName,
		Short:        fmt.Sprintf("Manage %s configuration files", strings.Title(bc.serviceName)),
		Long:         bc.buildLongDescription(),
		SilenceUsage: true,
	}

	cmd.AddCommand(bc.CreateSaveCommand())
	cmd.AddCommand(bc.CreateLoadCommand())
	cmd.AddCommand(bc.CreateListCommand())

	return cmd
}

// buildLongDescription builds the long description for the command.
func (bc *BaseCommand) buildLongDescription() string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("Save and load %s configuration files.\n\n", strings.Title(bc.serviceName)))
	builder.WriteString(bc.description)
	builder.WriteString("\n\n")
	builder.WriteString(fmt.Sprintf("The configurations are saved to ~/.gz/%s-configs/ by default.\n\n", bc.serviceName))
	builder.WriteString("Examples:\n")

	for _, example := range bc.examples {
		builder.WriteString("  ")
		builder.WriteString(example)
		builder.WriteString("\n")
	}

	return builder.String()
}

// CreateSaveCommand creates the save subcommand.
func (bc *BaseCommand) CreateSaveCommand() *cobra.Command {
	opts := bc.DefaultOptions()

	cmd := &cobra.Command{
		Use:   "save",
		Short: fmt.Sprintf("Save current %s configuration", bc.serviceName),
		RunE: func(cmd *cobra.Command, args []string) error {
			return bc.SaveConfig(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Name, "name", "n", "", "Name for the saved configuration (required)")
	cmd.Flags().StringVarP(&opts.Description, "description", "d", "", "Description for the saved configuration")
	cmd.Flags().StringVar(&opts.ConfigPath, "config-path", opts.ConfigPath, fmt.Sprintf("Path to %s config file", bc.serviceName))
	cmd.Flags().StringVar(&opts.StorePath, "store-path", opts.StorePath, "Path to store saved configurations")
	cmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "Overwrite existing configuration")

	cmd.MarkFlagRequired("name")

	return cmd
}

// CreateLoadCommand creates the load subcommand.
func (bc *BaseCommand) CreateLoadCommand() *cobra.Command {
	opts := bc.DefaultOptions()

	cmd := &cobra.Command{
		Use:   "load",
		Short: fmt.Sprintf("Load saved %s configuration", bc.serviceName),
		RunE: func(cmd *cobra.Command, args []string) error {
			return bc.LoadConfig(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Name, "name", "n", "", "Name of the configuration to load (required)")
	cmd.Flags().StringVar(&opts.ConfigPath, "config-path", opts.ConfigPath, fmt.Sprintf("Path to %s config file", bc.serviceName))
	cmd.Flags().StringVar(&opts.StorePath, "store-path", opts.StorePath, "Path to stored configurations")
	cmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "Overwrite existing config file")

	cmd.MarkFlagRequired("name")

	return cmd
}

// CreateListCommand creates the list subcommand.
func (bc *BaseCommand) CreateListCommand() *cobra.Command {
	opts := bc.DefaultOptions()

	cmd := &cobra.Command{
		Use:   "list",
		Short: fmt.Sprintf("List saved %s configurations", bc.serviceName),
		RunE: func(cmd *cobra.Command, args []string) error {
			return bc.ListConfigs(opts)
		},
	}

	cmd.Flags().StringVar(&opts.StorePath, "store-path", opts.StorePath, "Path to stored configurations")
	cmd.Flags().BoolVarP(&opts.ListAll, "all", "a", false, "Show detailed information for all configurations")

	return cmd
}

// SaveConfig saves the current configuration.
func (bc *BaseCommand) SaveConfig(opts *BaseOptions) error {
	// Validate inputs
	if opts.Name == "" {
		return fmt.Errorf("configuration name is required")
	}

	// Check if source config exists
	if _, err := os.Stat(opts.ConfigPath); os.IsNotExist(err) {
		return fmt.Errorf("%s config file not found at %s", bc.serviceName, opts.ConfigPath)
	}

	// Create store directory if it doesn't exist
	if err := os.MkdirAll(opts.StorePath, 0o755); err != nil {
		return fmt.Errorf("failed to create store directory: %w", err)
	}

	// Check if config already exists
	configFile := filepath.Join(opts.StorePath, opts.Name+"."+bc.configFileName)
	if _, err := os.Stat(configFile); err == nil && !opts.Force {
		return fmt.Errorf("configuration '%s' already exists (use --force to overwrite)", opts.Name)
	}

	// Copy config file
	if err := bc.copyFile(opts.ConfigPath, configFile); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	// Save metadata
	metadata := ConfigMetadata{
		Description: opts.Description,
		SavedAt:     time.Now(),
		SourcePath:  opts.ConfigPath,
	}

	metadataFile := filepath.Join(opts.StorePath, opts.Name+".metadata.json")
	if err := bc.saveMetadata(metadataFile, metadata); err != nil {
		// Don't fail if metadata save fails, just warn
		fmt.Printf("Warning: failed to save metadata: %v\n", err)
	}

	fmt.Printf("✅ %s configuration '%s' saved successfully\n", strings.Title(bc.serviceName), opts.Name)
	if opts.Description != "" {
		fmt.Printf("   Description: %s\n", opts.Description)
	}
	fmt.Printf("   Saved to: %s\n", configFile)

	return nil
}

// LoadConfig loads a saved configuration.
func (bc *BaseCommand) LoadConfig(opts *BaseOptions) error {
	// Validate inputs
	if opts.Name == "" {
		return fmt.Errorf("configuration name is required")
	}

	// Check if saved config exists
	configFile := filepath.Join(opts.StorePath, opts.Name+"."+bc.configFileName)
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return fmt.Errorf("configuration '%s' not found", opts.Name)
	}

	// Check if target config already exists
	if _, err := os.Stat(opts.ConfigPath); err == nil && !opts.Force {
		return fmt.Errorf("config file already exists at %s (use --force to overwrite)", opts.ConfigPath)
	}

	// Create target directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(opts.ConfigPath), 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Copy config file
	if err := bc.copyFile(configFile, opts.ConfigPath); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Load and display metadata if available
	metadataFile := filepath.Join(opts.StorePath, opts.Name+".metadata.json")
	if metadata, err := bc.loadMetadata(metadataFile); err == nil {
		fmt.Printf("✅ %s configuration '%s' loaded successfully\n", strings.Title(bc.serviceName), opts.Name)
		if metadata.Description != "" {
			fmt.Printf("   Description: %s\n", metadata.Description)
		}
		fmt.Printf("   Originally saved: %s\n", metadata.SavedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("   Loaded to: %s\n", opts.ConfigPath)
	} else {
		fmt.Printf("✅ %s configuration '%s' loaded successfully to %s\n", strings.Title(bc.serviceName), opts.Name, opts.ConfigPath)
	}

	return nil
}

// ListConfigs lists saved configurations.
func (bc *BaseCommand) ListConfigs(opts *BaseOptions) error {
	// Check if store directory exists
	if _, err := os.Stat(opts.StorePath); os.IsNotExist(err) {
		fmt.Printf("No %s configurations found (store directory doesn't exist)\n", bc.serviceName)
		return nil
	}

	// Read directory contents
	entries, err := os.ReadDir(opts.StorePath)
	if err != nil {
		return fmt.Errorf("failed to read store directory: %w", err)
	}

	// Filter for config files
	configExtension := "." + bc.configFileName
	var configs []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), configExtension) {
			configName := strings.TrimSuffix(entry.Name(), configExtension)
			configs = append(configs, configName)
		}
	}

	if len(configs) == 0 {
		fmt.Printf("No %s configurations found\n", bc.serviceName)
		return nil
	}

	fmt.Printf("Saved %s configurations:\n\n", bc.serviceName)

	for _, configName := range configs {
		if opts.ListAll {
			bc.printDetailedConfig(opts.StorePath, configName)
		} else {
			fmt.Printf("  • %s\n", configName)
		}
	}

	if !opts.ListAll {
		fmt.Printf("\nUse --all to show detailed information\n")
	}

	return nil
}

// printDetailedConfig prints detailed information about a configuration.
func (bc *BaseCommand) printDetailedConfig(storePath, configName string) {
	fmt.Printf("  📄 %s\n", configName)

	metadataFile := filepath.Join(storePath, configName+".metadata.json")
	if metadata, err := bc.loadMetadata(metadataFile); err == nil {
		if metadata.Description != "" {
			fmt.Printf("     Description: %s\n", metadata.Description)
		}
		fmt.Printf("     Saved: %s\n", metadata.SavedAt.Format("2006-01-02 15:04:05"))
		if metadata.SourcePath != "" {
			fmt.Printf("     Source: %s\n", metadata.SourcePath)
		}
	}

	configFile := filepath.Join(storePath, configName+"."+bc.configFileName)
	if stat, err := os.Stat(configFile); err == nil {
		fmt.Printf("     Size: %d bytes\n", stat.Size())
	}

	fmt.Println()
}

// copyFile copies a file from src to dst.
func (bc *BaseCommand) copyFile(src, dst string) error {
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

// saveMetadata saves metadata to a JSON file.
func (bc *BaseCommand) saveMetadata(filename string, metadata ConfigMetadata) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(metadata)
}

// loadMetadata loads metadata from a JSON file.
func (bc *BaseCommand) loadMetadata(filename string) (*ConfigMetadata, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var metadata ConfigMetadata
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&metadata)
	return &metadata, err
}
