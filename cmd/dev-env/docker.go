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

type dockerOptions struct {
	name        string
	description string
	configPath  string
	storePath   string
	force       bool
	listAll     bool
}

func defaultDockerOptions() *dockerOptions {
	homeDir, _ := os.UserHomeDir()

	return &dockerOptions{
		configPath: filepath.Join(homeDir, ".docker", "config.json"),
		storePath:  filepath.Join(homeDir, ".gz", "docker-configs"),
	}
}

func newDockerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "docker",
		Short: "Manage Docker configuration files",
		Long: `Save and load Docker configuration files.

This command helps you backup and restore Docker configuration files, which contain:
- Docker registry authentication credentials
- Docker daemon configuration settings
- Registry mirrors and insecure registries
- Other Docker client settings

This is useful when:
- Setting up new development machines
- Switching between different Docker environments
- Backing up Docker credentials before changes
- Managing multiple Docker configurations for different projects

The configurations are saved to ~/.gz/docker-configs/ by default.

Examples:
  # Save current Docker config with a name
  gz dev-env docker save --name production
  
  # Save with description
  gz dev-env docker save --name staging --description "Staging Docker config"
  
  # Load a saved Docker config
  gz dev-env docker load --name production
  
  # List all saved configurations
  gz dev-env docker list
  
  # Save from specific path
  gz dev-env docker save --name custom --config-path /path/to/config.json`,
		SilenceUsage: true,
	}

	cmd.AddCommand(newDockerSaveCmd())
	cmd.AddCommand(newDockerLoadCmd())
	cmd.AddCommand(newDockerListCmd())

	return cmd
}

func newDockerSaveCmd() *cobra.Command {
	o := defaultDockerOptions()

	cmd := &cobra.Command{
		Use:   "save",
		Short: "Save current Docker config",
		Long: `Save the current Docker configuration file with a given name.

This creates a backup of your current Docker configuration that can be
restored later using the 'load' command. The configuration includes
registry authentication, daemon settings, and other Docker client preferences.

Examples:
  # Save current Docker config as "production"
  gz dev-env docker save --name production
  
  # Save with description
  gz dev-env docker save --name staging --description "Staging environment"
  
  # Save from specific path
  gz dev-env docker save --name custom --config-path /path/to/config.json`,
		RunE: o.runSave,
	}

	cmd.Flags().StringVarP(&o.name, "name", "n", "", "Name for the saved configuration (required)")
	cmd.Flags().StringVarP(&o.description, "description", "d", "", "Description for the configuration")
	cmd.Flags().StringVar(&o.configPath, "config-path", o.configPath, "Path to Docker config file to save")
	cmd.Flags().StringVar(&o.storePath, "store-path", o.storePath, "Directory to store saved configurations")
	cmd.Flags().BoolVarP(&o.force, "force", "f", false, "Overwrite existing saved configuration")

	cmd.MarkFlagRequired("name")

	return cmd
}

func newDockerLoadCmd() *cobra.Command {
	o := defaultDockerOptions()

	cmd := &cobra.Command{
		Use:   "load",
		Short: "Load a saved Docker config",
		Long: `Load a previously saved Docker configuration file.

This restores a Docker configuration that was previously saved using the 'save' command.
The current Docker config will be backed up before loading the new one.

Examples:
  # Load the "production" configuration
  gz dev-env docker load --name production
  
  # Load to specific path
  gz dev-env docker load --name staging --config-path /path/to/config.json`,
		RunE: o.runLoad,
	}

	cmd.Flags().StringVarP(&o.name, "name", "n", "", "Name of the saved configuration to load (required)")
	cmd.Flags().StringVar(&o.configPath, "config-path", o.configPath, "Path where to load the Docker config")
	cmd.Flags().StringVar(&o.storePath, "store-path", o.storePath, "Directory where saved configurations are stored")
	cmd.Flags().BoolVarP(&o.force, "force", "f", false, "Skip backup of current configuration")

	cmd.MarkFlagRequired("name")

	return cmd
}

func newDockerListCmd() *cobra.Command {
	o := defaultDockerOptions()

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List saved Docker configs",
		Long: `List all saved Docker configuration files.

This shows all the Docker configuration files that have been saved using the 'save' command,
along with their descriptions, save dates, and registry information.

Examples:
  # List all saved configurations
  gz dev-env docker list`,
		RunE: o.runList,
	}

	cmd.Flags().StringVar(&o.storePath, "store-path", o.storePath, "Directory where saved configurations are stored")

	return cmd
}

func (o *dockerOptions) runSave(_ *cobra.Command, args []string) error {
	// Check if source config exists
	if _, err := os.Stat(o.configPath); os.IsNotExist(err) {
		return fmt.Errorf("Docker config file not found at %s", o.configPath)
	}

	// Create store directory if it doesn't exist
	if err := os.MkdirAll(o.storePath, 0o755); err != nil {
		return fmt.Errorf("failed to create store directory: %w", err)
	}

	// Check if name already exists
	savedPath := filepath.Join(o.storePath, o.name+".json")
	if _, err := os.Stat(savedPath); err == nil && !o.force {
		return fmt.Errorf("configuration '%s' already exists. Use --force to overwrite", o.name)
	}

	// Copy the Docker config file
	if err := o.copyFile(o.configPath, savedPath); err != nil {
		return fmt.Errorf("failed to save Docker config: %w", err)
	}

	// Save metadata
	if err := o.saveMetadata(); err != nil {
		fmt.Printf("Warning: failed to save metadata: %v\n", err)
	}

	// Display registry information
	if err := o.displayConfigInfo(savedPath); err != nil {
		fmt.Printf("Warning: failed to read config info: %v\n", err)
	}

	fmt.Printf("✅ Docker config saved as '%s'\n", o.name)

	if o.description != "" {
		fmt.Printf("   Description: %s\n", o.description)
	}

	fmt.Printf("   Saved to: %s\n", savedPath)

	return nil
}

func (o *dockerOptions) runLoad(_ *cobra.Command, args []string) error {
	// Check if saved config exists
	savedPath := filepath.Join(o.storePath, o.name+".json")
	if _, err := os.Stat(savedPath); os.IsNotExist(err) {
		return fmt.Errorf("saved configuration '%s' not found", o.name)
	}

	// Backup current config if it exists and force is not set
	if !o.force {
		if _, err := os.Stat(o.configPath); err == nil {
			backupPath := o.configPath + ".backup." + time.Now().Format("20060102-150405")
			if err := o.copyFile(o.configPath, backupPath); err != nil {
				return fmt.Errorf("failed to backup current Docker config: %w", err)
			}

			fmt.Printf("📦 Current Docker config backed up to: %s\n", backupPath)
		}
	}

	// Create target directory if it doesn't exist
	targetDir := filepath.Dir(o.configPath)
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Copy the saved config to target location
	if err := o.copyFile(savedPath, o.configPath); err != nil {
		return fmt.Errorf("failed to load Docker config: %w", err)
	}

	// Display config information
	if err := o.displayConfigInfo(o.configPath); err != nil {
		fmt.Printf("Warning: failed to read config info: %v\n", err)
	}

	fmt.Printf("✅ Docker config '%s' loaded successfully\n", o.name)
	fmt.Printf("   Loaded to: %s\n", o.configPath)

	return nil
}

func (o *dockerOptions) runList(_ *cobra.Command, args []string) error {
	// Check if store directory exists
	if _, err := os.Stat(o.storePath); os.IsNotExist(err) {
		fmt.Println("No saved Docker configs found.")
		return nil
	}

	// Read directory contents
	entries, err := os.ReadDir(o.storePath)
	if err != nil {
		return fmt.Errorf("failed to read store directory: %w", err)
	}

	// Filter for .json files
	var configs []string

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			name := strings.TrimSuffix(entry.Name(), ".json")
			configs = append(configs, name)
		}
	}

	if len(configs) == 0 {
		fmt.Println("No saved Docker configs found.")
		return nil
	}

	fmt.Printf("Saved Docker configs (%d):\n\n", len(configs))

	// Read and display metadata for each config
	for _, name := range configs {
		metadata := o.loadMetadata(name)
		fmt.Printf("🐳 %s\n", name)

		if metadata.Description != "" {
			fmt.Printf("   Description: %s\n", metadata.Description)
		}

		if !metadata.SavedAt.IsZero() {
			fmt.Printf("   Saved: %s\n", metadata.SavedAt.Format("2006-01-02 15:04:05"))
		}

		configPath := filepath.Join(o.storePath, name+".json")
		if info, err := os.Stat(configPath); err == nil {
			fmt.Printf("   Size: %d bytes\n", info.Size())
		}

		// Display registry information
		if err := o.displayConfigInfo(configPath); err == nil {
			// Already displayed in displayConfigInfo
		}

		fmt.Println()
	}

	return nil
}

type dockerMetadata struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	SavedAt     time.Time `json:"saved_at"`
	SourcePath  string    `json:"source_path"`
}

func (o *dockerOptions) saveMetadata() error {
	metadata := dockerMetadata{
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

func (o *dockerOptions) loadMetadata(name string) dockerMetadata {
	metadata := dockerMetadata{Name: name}

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

func (o *dockerOptions) copyFile(src, dst string) error {
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

// dockerConfig represents the structure of Docker's config.json.
type dockerConfig struct {
	Auths             map[string]interface{} `json:"auths,omitempty"`
	CredsStore        string                 `json:"credsStore,omitempty"`
	CredHelpers       map[string]string      `json:"credHelpers,omitempty"`
	Experimental      string                 `json:"experimental,omitempty"`
	StackOrchestrator string                 `json:"stackOrchestrator,omitempty"`
}

func (o *dockerOptions) displayConfigInfo(configPath string) error {
	// Read and parse Docker config
	content, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	var config dockerConfig
	if err := json.Unmarshal(content, &config); err != nil {
		return err
	}

	// Display authentication information
	if len(config.Auths) > 0 {
		fmt.Printf("   Registries: %d configured\n", len(config.Auths))

		for registry := range config.Auths {
			fmt.Printf("     - %s\n", registry)
		}
	}

	// Display credential store info
	if config.CredsStore != "" {
		fmt.Printf("   Credential store: %s\n", config.CredsStore)
	}

	// Display credential helpers
	if len(config.CredHelpers) > 0 {
		fmt.Printf("   Credential helpers: %d configured\n", len(config.CredHelpers))
	}

	return nil
}
