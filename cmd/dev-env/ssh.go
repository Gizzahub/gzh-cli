// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package devenv

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

type sshOptions struct {
	name        string
	description string
	configPath  string
	storePath   string
	force       bool
	listAll     bool
}

func defaultSshOptions() *sshOptions {
	homeDir, _ := os.UserHomeDir()

	return &sshOptions{
		configPath: filepath.Join(homeDir, ".ssh", "config"),
		storePath:  filepath.Join(homeDir, ".gz", "ssh-configs"),
	}
}

func newSshCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ssh",
		Short: "Manage SSH configuration files",
		Long: `Save and load SSH configuration files.

This command helps you backup and restore SSH configuration files, which contain:
- SSH host configurations
- Connection settings (hostname, port, user)
- Key file specifications
- Proxy and tunnel settings
- Other SSH client options

This is useful when:
- Setting up new development machines
- Switching between different SSH environments
- Backing up SSH configurations before changes
- Managing multiple SSH configurations for different projects

The configurations are saved to ~/.gz/ssh-configs/ by default.

Examples:
  # Save current SSH config with a name
  gz dev-env ssh save --name production
  
  # Save with description
  gz dev-env ssh save --name work --description "Work SSH config"
  
  # Load a saved SSH config
  gz dev-env ssh load --name production
  
  # List all saved configurations
  gz dev-env ssh list
  
  # Save from specific path
  gz dev-env ssh save --name custom --config-path /path/to/ssh_config`,
		SilenceUsage: true,
	}

	cmd.AddCommand(newSshSaveCmd())
	cmd.AddCommand(newSshLoadCmd())
	cmd.AddCommand(newSshListCmd())

	return cmd
}

func newSshSaveCmd() *cobra.Command {
	o := defaultSshOptions()

	cmd := &cobra.Command{
		Use:   "save",
		Short: "Save current SSH config",
		Long: `Save the current SSH configuration file with a given name.

This creates a backup of your current SSH configuration that can be
restored later using the 'load' command. The configuration includes
host definitions, connection settings, and SSH client options.

Examples:
  # Save current SSH config as "production"
  gz dev-env ssh save --name production
  
  # Save with description
  gz dev-env ssh save --name work --description "Work environment"
  
  # Save from specific path
  gz dev-env ssh save --name custom --config-path /path/to/ssh_config`,
		RunE: o.runSave,
	}

	cmd.Flags().StringVarP(&o.name, "name", "n", "", "Name for the saved configuration (required)")
	cmd.Flags().StringVarP(&o.description, "description", "d", "", "Description for the configuration")
	cmd.Flags().StringVar(&o.configPath, "config-path", o.configPath, "Path to SSH config file to save")
	cmd.Flags().StringVar(&o.storePath, "store-path", o.storePath, "Directory to store saved configurations")
	cmd.Flags().BoolVarP(&o.force, "force", "f", false, "Overwrite existing saved configuration")

	cmd.MarkFlagRequired("name")

	return cmd
}

func newSshLoadCmd() *cobra.Command {
	o := defaultSshOptions()

	cmd := &cobra.Command{
		Use:   "load",
		Short: "Load a saved SSH config",
		Long: `Load a previously saved SSH configuration file.

This restores an SSH configuration that was previously saved using the 'save' command.
The current SSH config will be backed up before loading the new one.

Examples:
  # Load the "production" configuration
  gz dev-env ssh load --name production
  
  # Load to specific path
  gz dev-env ssh load --name work --config-path /path/to/ssh_config`,
		RunE: o.runLoad,
	}

	cmd.Flags().StringVarP(&o.name, "name", "n", "", "Name of the saved configuration to load (required)")
	cmd.Flags().StringVar(&o.configPath, "config-path", o.configPath, "Path where to load the SSH config")
	cmd.Flags().StringVar(&o.storePath, "store-path", o.storePath, "Directory where saved configurations are stored")
	cmd.Flags().BoolVarP(&o.force, "force", "f", false, "Skip backup of current configuration")

	cmd.MarkFlagRequired("name")

	return cmd
}

func newSshListCmd() *cobra.Command {
	o := defaultSshOptions()

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List saved SSH configs",
		Long: `List all saved SSH configuration files.

This shows all the SSH configuration files that have been saved using the 'save' command,
along with their descriptions, save dates, and host information.

Examples:
  # List all saved configurations
  gz dev-env ssh list`,
		RunE: o.runList,
	}

	cmd.Flags().StringVar(&o.storePath, "store-path", o.storePath, "Directory where saved configurations are stored")

	return cmd
}

func (o *sshOptions) runSave(_ *cobra.Command, args []string) error {
	// Check if source config exists
	if _, err := os.Stat(o.configPath); os.IsNotExist(err) {
		return fmt.Errorf("SSH config file not found at %s", o.configPath)
	}

	// Create store directory if it doesn't exist
	if err := os.MkdirAll(o.storePath, 0o755); err != nil {
		return fmt.Errorf("failed to create store directory: %w", err)
	}

	// Check if name already exists
	savedPath := filepath.Join(o.storePath, o.name+".config")
	if _, err := os.Stat(savedPath); err == nil && !o.force {
		return fmt.Errorf("configuration '%s' already exists. Use --force to overwrite", o.name)
	}

	// Copy the SSH config file
	if err := o.copyFile(o.configPath, savedPath); err != nil {
		return fmt.Errorf("failed to save SSH config: %w", err)
	}

	// Save metadata
	if err := o.saveMetadata(); err != nil {
		fmt.Printf("Warning: failed to save metadata: %v\n", err)
	}

	// Display config information
	if err := o.displayConfigInfo(savedPath); err != nil {
		fmt.Printf("Warning: failed to read config info: %v\n", err)
	}

	fmt.Printf("✅ SSH config saved as '%s'\n", o.name)

	if o.description != "" {
		fmt.Printf("   Description: %s\n", o.description)
	}

	fmt.Printf("   Saved to: %s\n", savedPath)

	return nil
}

func (o *sshOptions) runLoad(_ *cobra.Command, args []string) error {
	// Check if saved config exists
	savedPath := filepath.Join(o.storePath, o.name+".config")
	if _, err := os.Stat(savedPath); os.IsNotExist(err) {
		return fmt.Errorf("saved configuration '%s' not found", o.name)
	}

	// Backup current config if it exists and force is not set
	if !o.force {
		if _, err := os.Stat(o.configPath); err == nil {
			backupPath := o.configPath + ".backup." + time.Now().Format("20060102-150405")
			if err := o.copyFile(o.configPath, backupPath); err != nil {
				return fmt.Errorf("failed to backup current SSH config: %w", err)
			}

			fmt.Printf("📦 Current SSH config backed up to: %s\n", backupPath)
		}
	}

	// Create target directory if it doesn't exist
	targetDir := filepath.Dir(o.configPath)
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Copy the saved config to target location
	if err := o.copyFile(savedPath, o.configPath); err != nil {
		return fmt.Errorf("failed to load SSH config: %w", err)
	}

	// Display config information
	if err := o.displayConfigInfo(o.configPath); err != nil {
		fmt.Printf("Warning: failed to read config info: %v\n", err)
	}

	fmt.Printf("✅ SSH config '%s' loaded successfully\n", o.name)
	fmt.Printf("   Loaded to: %s\n", o.configPath)

	return nil
}

func (o *sshOptions) runList(_ *cobra.Command, args []string) error {
	// Check if store directory exists
	if _, err := os.Stat(o.storePath); os.IsNotExist(err) {
		fmt.Println("No saved SSH configs found.")
		return nil
	}

	// Read directory contents
	entries, err := os.ReadDir(o.storePath)
	if err != nil {
		return fmt.Errorf("failed to read store directory: %w", err)
	}

	// Filter for .config files
	var configs []string

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".config") {
			name := strings.TrimSuffix(entry.Name(), ".config")
			configs = append(configs, name)
		}
	}

	if len(configs) == 0 {
		fmt.Println("No saved SSH configs found.")
		return nil
	}

	fmt.Printf("Saved SSH configs (%d):\n\n", len(configs))

	// Read and display metadata for each config
	for _, name := range configs {
		metadata := o.loadMetadata(name)
		fmt.Printf("🔐 %s\n", name)

		if metadata.Description != "" {
			fmt.Printf("   Description: %s\n", metadata.Description)
		}

		if !metadata.SavedAt.IsZero() {
			fmt.Printf("   Saved: %s\n", metadata.SavedAt.Format("2006-01-02 15:04:05"))
		}

		configPath := filepath.Join(o.storePath, name+".config")
		if info, err := os.Stat(configPath); err == nil {
			fmt.Printf("   Size: %d bytes\n", info.Size())
		}

		// Display SSH host information
		if err := o.displayConfigInfo(configPath); err == nil {
			// Already displayed in displayConfigInfo
		}

		fmt.Println()
	}

	return nil
}

type sshMetadata struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	SavedAt     time.Time `json:"saved_at"`
	SourcePath  string    `json:"source_path"`
}

func (o *sshOptions) saveMetadata() error {
	metadata := sshMetadata{
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

func (o *sshOptions) loadMetadata(name string) sshMetadata {
	metadata := sshMetadata{Name: name}

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

func (o *sshOptions) copyFile(src, dst string) error {
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

func (o *sshOptions) displayConfigInfo(configPath string) error {
	// Parse SSH config file to extract host information
	hosts := o.parseSshConfig(configPath)

	// Display host information
	if len(hosts) > 0 {
		fmt.Printf("   Hosts: %d configured\n", len(hosts))

		for _, host := range hosts {
			fmt.Printf("     - %s", host.Name)

			if host.Hostname != "" && host.Hostname != host.Name {
				fmt.Printf(" -> %s", host.Hostname)
			}

			if host.User != "" {
				fmt.Printf(" (user: %s)", host.User)
			}

			if host.Port != "" && host.Port != "22" {
				fmt.Printf(" (port: %s)", host.Port)
			}

			fmt.Println()
		}
	}

	return nil
}

type sshHost struct {
	Name     string
	Hostname string
	User     string
	Port     string
	KeyFile  string
}

func (o *sshOptions) parseSshConfig(configPath string) []sshHost {
	content, err := os.ReadFile(configPath)
	if err != nil {
		return nil
	}

	var hosts []sshHost

	var currentHost *sshHost

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse SSH config directives
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		directive := strings.ToLower(parts[0])
		value := strings.Join(parts[1:], " ")

		switch directive {
		case "host":
			// Save previous host if exists
			if currentHost != nil {
				hosts = append(hosts, *currentHost)
			}

			// Start new host
			// Handle multiple host patterns
			hostPatterns := strings.Fields(value)
			if len(hostPatterns) > 0 {
				// Use the first pattern as the main name
				currentHost = &sshHost{Name: hostPatterns[0]}
			}

		case "hostname":
			if currentHost != nil {
				currentHost.Hostname = value
			}

		case "user":
			if currentHost != nil {
				currentHost.User = value
			}

		case "port":
			if currentHost != nil {
				currentHost.Port = value
			}

		case "identityfile":
			if currentHost != nil {
				// Clean up the path (remove quotes if present)
				keyFile := strings.Trim(value, "\"'")
				currentHost.KeyFile = keyFile
			}
		}
	}

	// Don't forget the last host
	if currentHost != nil {
		hosts = append(hosts, *currentHost)
	}

	// Filter out wildcard hosts for display
	var filteredHosts []sshHost

	wildcardPattern := regexp.MustCompile(`[*?]`)
	for _, host := range hosts {
		if !wildcardPattern.MatchString(host.Name) {
			filteredHosts = append(filteredHosts, host)
		}
	}

	return filteredHosts
}
