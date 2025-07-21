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

const (
	metadataKeyDescription = "description"
	metadataKeySavedAt     = "saved_at"
	metadataKeySourcePath  = "source_path"
)

type awsOptions struct {
	name        string
	description string
	configPath  string
	storePath   string
	force       bool
	listAll     bool
}

func defaultAwsOptions() *awsOptions {
	homeDir, _ := os.UserHomeDir()

	return &awsOptions{
		configPath: filepath.Join(homeDir, ".aws", "config"),
		storePath:  filepath.Join(homeDir, ".gz", "aws-configs"),
	}
}

func newAwsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "aws",
		Short: "Manage AWS configuration files",
		Long: `Save and load AWS configuration files.

This command helps you backup and restore AWS configuration files, which contain:
- AWS profiles and regions
- Default output format settings
- Role and credential configurations
- SSO configurations
- Other AWS CLI settings

This is useful when:
- Setting up new development machines
- Switching between different AWS environments
- Backing up AWS configurations before changes
- Managing multiple AWS configurations for different projects

The configurations are saved to ~/.gz/aws-configs/ by default.

Examples:
  # Save current AWS config with a name
  gz dev-env aws save --name production

  # Save with description
  gz dev-env aws save --name staging --description "Staging AWS config"

  # Load a saved AWS config
  gz dev-env aws load --name production

  # List all saved configurations
  gz dev-env aws list

  # Save from specific path
  gz dev-env aws save --name custom --config-path /path/to/config`,
		SilenceUsage: true,
	}

	cmd.AddCommand(newAwsSaveCmd())
	cmd.AddCommand(newAwsLoadCmd())
	cmd.AddCommand(newAwsListCmd())

	return cmd
}

func newAwsSaveCmd() *cobra.Command {
	o := defaultAwsOptions()

	cmd := &cobra.Command{
		Use:   "save",
		Short: "Save current AWS config",
		Long: `Save the current AWS configuration file with a given name.

This creates a backup of your current AWS configuration that can be
restored later using the 'load' command. The configuration includes
profiles, regions, output formats, and other AWS CLI settings.

Examples:
  # Save current AWS config as "production"
  gz dev-env aws save --name production

  # Save with description
  gz dev-env aws save --name staging --description "Staging environment"

  # Save from specific path
  gz dev-env aws save --name custom --config-path /path/to/config`,
		RunE: o.runSave,
	}

	cmd.Flags().StringVarP(&o.name, "name", "n", "", "Name for the saved configuration (required)")
	cmd.Flags().StringVarP(&o.description, "description", "d", "", "Description for the configuration")
	cmd.Flags().StringVar(&o.configPath, "config-path", o.configPath, "Path to AWS config file to save")
	cmd.Flags().StringVar(&o.storePath, "store-path", o.storePath, "Directory to store saved configurations")
	cmd.Flags().BoolVarP(&o.force, "force", "f", false, "Overwrite existing saved configuration")

	if err := cmd.MarkFlagRequired("name"); err != nil {
		// This should not happen with a valid flag name
		panic(fmt.Sprintf("Failed to mark flag as required: %v", err))
	}

	return cmd
}

func newAwsLoadCmd() *cobra.Command {
	o := defaultAwsOptions()

	cmd := &cobra.Command{
		Use:   "load",
		Short: "Load a saved AWS config",
		Long: `Load a previously saved AWS configuration file.

This restores an AWS configuration that was previously saved using the 'save' command.
The current AWS config will be backed up before loading the new one.

Examples:
  # Load the "production" configuration
  gz dev-env aws load --name production

  # Load to specific path
  gz dev-env aws load --name staging --config-path /path/to/config`,
		RunE: o.runLoad,
	}

	cmd.Flags().StringVarP(&o.name, "name", "n", "", "Name of the saved configuration to load (required)")
	cmd.Flags().StringVar(&o.configPath, "config-path", o.configPath, "Path where to load the AWS config")
	cmd.Flags().StringVar(&o.storePath, "store-path", o.storePath, "Directory where saved configurations are stored")
	cmd.Flags().BoolVarP(&o.force, "force", "f", false, "Skip backup of current configuration")

	if err := cmd.MarkFlagRequired("name"); err != nil {
		// This should not happen with a valid flag name
		panic(fmt.Sprintf("Failed to mark flag as required: %v", err))
	}

	return cmd
}

func newAwsListCmd() *cobra.Command {
	o := defaultAwsOptions()

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List saved AWS configs",
		Long: `List all saved AWS configuration files.

This shows all the AWS configuration files that have been saved using the 'save' command,
along with their descriptions, save dates, and profile information.

Examples:
  # List all saved configurations
  gz dev-env aws list`,
		RunE: o.runList,
	}

	cmd.Flags().StringVar(&o.storePath, "store-path", o.storePath, "Directory where saved configurations are stored")

	return cmd
}

func (o *awsOptions) runSave(_ *cobra.Command, args []string) error {
	// Check if source config exists
	if _, err := os.Stat(o.configPath); os.IsNotExist(err) {
		return fmt.Errorf("AWS config file not found at %s", o.configPath)
	}

	// Create store directory if it doesn't exist
	if err := os.MkdirAll(o.storePath, 0o750); err != nil {
		return fmt.Errorf("failed to create store directory: %w", err)
	}

	// Check if name already exists
	savedPath := filepath.Join(o.storePath, o.name+".config")
	if _, err := os.Stat(savedPath); err == nil && !o.force {
		return fmt.Errorf("configuration '%s' already exists. Use --force to overwrite", o.name)
	}

	// Copy the AWS config file
	if err := o.copyFile(o.configPath, savedPath); err != nil {
		return fmt.Errorf("failed to save AWS config: %w", err)
	}

	// Save metadata
	if err := o.saveMetadata(); err != nil {
		fmt.Printf("Warning: failed to save metadata: %v\n", err)
	}

	// Display config information
	if err := o.displayConfigInfo(savedPath); err != nil {
		fmt.Printf("Warning: failed to read config info: %v\n", err)
	}

	fmt.Printf("✅ AWS config saved as '%s'\n", o.name)

	if o.description != "" {
		fmt.Printf("   Description: %s\n", o.description)
	}

	fmt.Printf("   Saved to: %s\n", savedPath)

	return nil
}

func (o *awsOptions) runLoad(_ *cobra.Command, args []string) error {
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
				return fmt.Errorf("failed to backup current AWS config: %w", err)
			}

			fmt.Printf("📦 Current AWS config backed up to: %s\n", backupPath)
		}
	}

	// Create target directory if it doesn't exist
	targetDir := filepath.Dir(o.configPath)
	if err := os.MkdirAll(targetDir, 0o750); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Copy the saved config to target location
	if err := o.copyFile(savedPath, o.configPath); err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Display config information
	if err := o.displayConfigInfo(o.configPath); err != nil {
		fmt.Printf("Warning: failed to read config info: %v\n", err)
	}

	fmt.Printf("✅ AWS config '%s' loaded successfully\n", o.name)
	fmt.Printf("   Loaded to: %s\n", o.configPath)

	return nil
}

func (o *awsOptions) runList(_ *cobra.Command, args []string) error {
	// Check if store directory exists
	if _, err := os.Stat(o.storePath); os.IsNotExist(err) {
		fmt.Println("No saved AWS configs found.")
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
		fmt.Println("No saved AWS configs found.")
		return nil
	}

	fmt.Printf("Saved AWS configs (%d):\n\n", len(configs))

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

		configPath := filepath.Join(o.storePath, name+".config")
		if info, err := os.Stat(configPath); err == nil {
			fmt.Printf("   Size: %d bytes\n", info.Size())
		}

		// Display profile information
		if err := o.displayConfigInfo(configPath); err != nil {
			// Log error but don't fail the operation
			fmt.Printf("Warning: Could not display config info: %v\n", err)
		}

		fmt.Println()
	}

	return nil
}

type awsMetadata struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	SavedAt     time.Time `json:"savedAt"`
	SourcePath  string    `json:"sourcePath"`
}

func (o *awsOptions) saveMetadata() error {
	metadata := awsMetadata{
		Name:        o.name,
		Description: o.description,
		SavedAt:     time.Now(),
		SourcePath:  o.configPath,
	}

	metadataPath := filepath.Join(o.storePath, o.name+".meta")

	file, err := os.OpenFile(metadataPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600) //nolint:gosec // Safe file path construction
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Printf("Warning: Failed to close file: %v\n", err)
		}
	}()

	// Write metadata as simple key-value pairs
	if metadata.Description != "" {
		if _, err := fmt.Fprintf(file, "description=%s\n", metadata.Description); err != nil {
			return fmt.Errorf("failed to write description: %w", err)
		}
	}

	if _, err := fmt.Fprintf(file, "saved_at=%s\n", metadata.SavedAt.Format(time.RFC3339)); err != nil {
		return fmt.Errorf("failed to write saved_at: %w", err)
	}
	if _, err := fmt.Fprintf(file, "source_path=%s\n", metadata.SourcePath); err != nil {
		return fmt.Errorf("failed to write source_path: %w", err)
	}

	return nil
}

func (o *awsOptions) loadMetadata(name string) awsMetadata {
	metadata := awsMetadata{Name: name}

	metadataPath := filepath.Join(o.storePath, name+".meta")

	// Validate metadataPath to prevent directory traversal
	if !filepath.IsAbs(metadataPath) {
		metadataPath = filepath.Join(o.storePath, filepath.Base(name)+".meta")
	}
	content, err := os.ReadFile(filepath.Clean(metadataPath))
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
		case metadataKeyDescription:
			metadata.Description = value
		case metadataKeySavedAt:
			if t, err := time.Parse(time.RFC3339, value); err == nil {
				metadata.SavedAt = t
			}
		case metadataKeySourcePath:
			metadata.SourcePath = value
		}
	}

	return metadata
}

func (o *awsOptions) copyFile(src, dst string) error {
	// Validate source file path to prevent directory traversal
	if !filepath.IsAbs(src) {
		return fmt.Errorf("source path must be absolute: %s", src)
	}
	sourceFile, err := os.Open(filepath.Clean(src))
	if err != nil {
		return err
	}
	defer func() {
		if err := sourceFile.Close(); err != nil {
			// Log error but don't override main error
			fmt.Printf("Warning: failed to close source file: %v\n", err)
		}
	}()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		if err := destFile.Close(); err != nil {
			// Log error but don't override main error
			fmt.Printf("Warning: failed to close destination file: %v\n", err)
		}
	}()

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

func (o *awsOptions) displayConfigInfo(configPath string) error {
	// Read and parse AWS config
	content, err := os.ReadFile(configPath) //nolint:gosec // Safe file path construction
	if err != nil {
		return err
	}

	// Parse AWS config file to extract profile information
	profiles := o.parseAwsConfig(string(content))

	// Display profile information
	if len(profiles) > 0 {
		fmt.Printf("   Profiles: %d configured\n", len(profiles))

		for _, profile := range profiles {
			fmt.Printf("     - %s", profile.Name)

			if profile.Region != "" {
				fmt.Printf(" (region: %s)", profile.Region)
			}

			if profile.Output != "" {
				fmt.Printf(" (output: %s)", profile.Output)
			}

			fmt.Println()
		}
	}

	return nil
}

type awsProfile struct {
	Name   string
	Region string
	Output string
}

func (o *awsOptions) parseAwsConfig(content string) []awsProfile {
	var (
		profiles       []awsProfile
		currentProfile *awsProfile
	)

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Check for profile section header
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			// Save previous profile if exists
			if currentProfile != nil {
				profiles = append(profiles, *currentProfile)
			}

			// Start new profile
			profileName := strings.Trim(line, "[]")
			// Handle both [default] and [profile name] formats
			profileName = strings.TrimPrefix(profileName, "profile ")

			currentProfile = &awsProfile{Name: profileName}

			continue
		}

		// Parse key-value pairs
		if currentProfile != nil && strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				switch key {
				case "region":
					currentProfile.Region = value
				case "output":
					currentProfile.Output = value
				}
			}
		}
	}

	// Don't forget the last profile
	if currentProfile != nil {
		profiles = append(profiles, *currentProfile)
	}

	return profiles
}
