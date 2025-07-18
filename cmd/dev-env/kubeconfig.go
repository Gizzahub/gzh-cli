package devenv

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

type kubeconfigOptions struct {
	name        string
	description string
	configPath  string
	storePath   string
	force       bool
	listAll     bool
}

func defaultKubeconfigOptions() *kubeconfigOptions {
	homeDir, _ := os.UserHomeDir()

	return &kubeconfigOptions{
		configPath: filepath.Join(homeDir, ".kube", "config"),
		storePath:  filepath.Join(homeDir, ".gz", "kubeconfigs"),
	}
}

func newKubeconfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kubeconfig",
		Short: "Manage Kubernetes configuration files",
		Long: `Save and load Kubernetes configuration files (kubeconfig).

This command helps you backup and restore kubeconfig files, which is useful when:
- Setting up new development machines
- Switching between different Kubernetes clusters
- Maintaining multiple cluster configurations
- Backing up access credentials before changes

The configurations are saved to ~/.gz/kubeconfigs/ by default.

Examples:
  # Save current kubeconfig with a name
  gz dev-env kubeconfig save --name production
  
  # Save with description
  gz dev-env kubeconfig save --name staging --description "Staging cluster config"
  
  # Load a saved kubeconfig
  gz dev-env kubeconfig load --name production
  
  # List all saved configurations
  gz dev-env kubeconfig list
  
  # Save from specific path
  gz dev-env kubeconfig save --name custom --config-path /path/to/config`,
		SilenceUsage: true,
	}

	cmd.AddCommand(newKubeconfigSaveCmd())
	cmd.AddCommand(newKubeconfigLoadCmd())
	cmd.AddCommand(newKubeconfigListCmd())

	return cmd
}

func newKubeconfigSaveCmd() *cobra.Command {
	o := defaultKubeconfigOptions()

	cmd := &cobra.Command{
		Use:   "save",
		Short: "Save current kubeconfig",
		Long: `Save the current kubeconfig file with a given name.

This creates a backup of your current Kubernetes configuration that can be
restored later using the 'load' command.

Examples:
  # Save current kubeconfig as "production"
  gz dev-env kubeconfig save --name production
  
  # Save with description
  gz dev-env kubeconfig save --name staging --description "Staging environment"
  
  # Save from specific path
  gz dev-env kubeconfig save --name custom --config-path /path/to/kubeconfig`,
		RunE: o.runSave,
	}

	cmd.Flags().StringVarP(&o.name, "name", "n", "", "Name for the saved configuration (required)")
	cmd.Flags().StringVarP(&o.description, "description", "d", "", "Description for the configuration")
	cmd.Flags().StringVar(&o.configPath, "config-path", o.configPath, "Path to kubeconfig file to save")
	cmd.Flags().StringVar(&o.storePath, "store-path", o.storePath, "Directory to store saved configurations")
	cmd.Flags().BoolVarP(&o.force, "force", "f", false, "Overwrite existing saved configuration")

	cmd.MarkFlagRequired("name")

	return cmd
}

func newKubeconfigLoadCmd() *cobra.Command {
	o := defaultKubeconfigOptions()

	cmd := &cobra.Command{
		Use:   "load",
		Short: "Load a saved kubeconfig",
		Long: `Load a previously saved kubeconfig file.

This restores a kubeconfig that was previously saved using the 'save' command.
The current kubeconfig will be backed up before loading the new one.

Examples:
  # Load the "production" configuration
  gz dev-env kubeconfig load --name production
  
  # Load to specific path
  gz dev-env kubeconfig load --name staging --config-path /path/to/kubeconfig`,
		RunE: o.runLoad,
	}

	cmd.Flags().StringVarP(&o.name, "name", "n", "", "Name of the saved configuration to load (required)")
	cmd.Flags().StringVar(&o.configPath, "config-path", o.configPath, "Path where to load the kubeconfig")
	cmd.Flags().StringVar(&o.storePath, "store-path", o.storePath, "Directory where saved configurations are stored")
	cmd.Flags().BoolVarP(&o.force, "force", "f", false, "Skip backup of current configuration")

	cmd.MarkFlagRequired("name")

	return cmd
}

func newKubeconfigListCmd() *cobra.Command {
	o := defaultKubeconfigOptions()

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List saved kubeconfigs",
		Long: `List all saved kubeconfig files.

This shows all the kubeconfig files that have been saved using the 'save' command,
along with their descriptions and save dates.

Examples:
  # List all saved configurations
  gz dev-env kubeconfig list`,
		RunE: o.runList,
	}

	cmd.Flags().StringVar(&o.storePath, "store-path", o.storePath, "Directory where saved configurations are stored")

	return cmd
}

func (o *kubeconfigOptions) runSave(_ *cobra.Command, args []string) error {
	// Check if source config exists
	if _, err := os.Stat(o.configPath); os.IsNotExist(err) {
		return fmt.Errorf("kubeconfig file not found at %s", o.configPath)
	}

	// Create store directory if it doesn't exist
	if err := os.MkdirAll(o.storePath, 0o755); err != nil {
		return fmt.Errorf("failed to create store directory: %w", err)
	}

	// Check if name already exists
	savedPath := filepath.Join(o.storePath, o.name+".yaml")
	if _, err := os.Stat(savedPath); err == nil && !o.force {
		return fmt.Errorf("configuration '%s' already exists. Use --force to overwrite", o.name)
	}

	// Copy the kubeconfig file
	if err := o.copyFile(o.configPath, savedPath); err != nil {
		return fmt.Errorf("failed to save kubeconfig: %w", err)
	}

	// Save metadata
	if err := o.saveMetadata(); err != nil {
		fmt.Printf("Warning: failed to save metadata: %v\n", err)
	}

	fmt.Printf("✅ Kubeconfig saved as '%s'\n", o.name)

	if o.description != "" {
		fmt.Printf("   Description: %s\n", o.description)
	}

	fmt.Printf("   Saved to: %s\n", savedPath)

	return nil
}

func (o *kubeconfigOptions) runLoad(_ *cobra.Command, args []string) error {
	// Check if saved config exists
	savedPath := filepath.Join(o.storePath, o.name+".yaml")
	if _, err := os.Stat(savedPath); os.IsNotExist(err) {
		return fmt.Errorf("saved configuration '%s' not found", o.name)
	}

	// Backup current config if it exists and force is not set
	if !o.force {
		if _, err := os.Stat(o.configPath); err == nil {
			backupPath := o.configPath + ".backup." + time.Now().Format("20060102-150405")
			if err := o.copyFile(o.configPath, backupPath); err != nil {
				return fmt.Errorf("failed to backup current kubeconfig: %w", err)
			}

			fmt.Printf("📦 Current kubeconfig backed up to: %s\n", backupPath)
		}
	}

	// Create target directory if it doesn't exist
	targetDir := filepath.Dir(o.configPath)
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Copy the saved config to target location
	if err := o.copyFile(savedPath, o.configPath); err != nil {
		return fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	fmt.Printf("✅ Kubeconfig '%s' loaded successfully\n", o.name)
	fmt.Printf("   Loaded to: %s\n", o.configPath)

	return nil
}

func (o *kubeconfigOptions) runList(_ *cobra.Command, args []string) error {
	// Check if store directory exists
	if _, err := os.Stat(o.storePath); os.IsNotExist(err) {
		fmt.Println("No saved kubeconfigs found.")
		return nil
	}

	// Read directory contents
	entries, err := os.ReadDir(o.storePath)
	if err != nil {
		return fmt.Errorf("failed to read store directory: %w", err)
	}

	// Filter for .yaml files
	var configs []string

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".yaml") {
			name := strings.TrimSuffix(entry.Name(), ".yaml")
			configs = append(configs, name)
		}
	}

	if len(configs) == 0 {
		fmt.Println("No saved kubeconfigs found.")
		return nil
	}

	fmt.Printf("Saved kubeconfigs (%d):\n\n", len(configs))

	// Read and display metadata for each config
	for _, name := range configs {
		metadata := o.loadMetadata(name)
		fmt.Printf("📋 %s\n", name)

		if metadata.Description != "" {
			fmt.Printf("   Description: %s\n", metadata.Description)
		}

		if !metadata.SavedAt.IsZero() {
			fmt.Printf("   Saved: %s\n", metadata.SavedAt.Format("2006-01-02 15:04:05"))
		}

		configPath := filepath.Join(o.storePath, name+".yaml")
		if info, err := os.Stat(configPath); err == nil {
			fmt.Printf("   Size: %d bytes\n", info.Size())
		}

		fmt.Println()
	}

	return nil
}

type kubeconfigMetadata struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	SavedAt     time.Time `json:"saved_at"`
	SourcePath  string    `json:"source_path"`
}

func (o *kubeconfigOptions) saveMetadata() error {
	metadata := kubeconfigMetadata{
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

func (o *kubeconfigOptions) loadMetadata(name string) kubeconfigMetadata {
	metadata := kubeconfigMetadata{Name: name}

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

func (o *kubeconfigOptions) copyFile(src, dst string) error {
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
