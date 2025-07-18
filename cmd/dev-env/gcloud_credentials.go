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

type gcloudCredentialsOptions struct {
	name        string
	description string
	configPath  string
	storePath   string
	force       bool
	listAll     bool
}

func defaultGcloudCredentialsOptions() *gcloudCredentialsOptions {
	homeDir, _ := os.UserHomeDir()

	return &gcloudCredentialsOptions{
		// gcloud credentials are typically stored in ~/.config/gcloud or ~/.gcloud
		configPath: filepath.Join(homeDir, ".config", "gcloud"),
		storePath:  filepath.Join(homeDir, ".gz", "gcloud-credentials"),
	}
}

func newGcloudCredentialsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gcloud-credentials",
		Short: "Manage Google Cloud credentials",
		Long: `Save and load Google Cloud credentials.

This command helps you backup and restore Google Cloud credentials, which contain:
- Service account key files
- Application default credentials
- User access tokens (when available)
- OAuth2 refresh tokens
- Other authentication information

This is useful when:
- Setting up new development machines
- Switching between different GCP credential sets
- Backing up GCP credentials before changes
- Managing multiple GCP credential sets for different projects/environments

The credentials are saved to ~/.gz/gcloud-credentials/ by default.

SECURITY WARNING: This stores sensitive credential information. Ensure your
storage location is properly secured and consider encrypting the stored files.

Examples:
  # Save current gcloud credentials with a name
  gz dev-env gcloud-credentials save --name production
  
  # Save with description
  gz dev-env gcloud-credentials save --name staging --description "Staging GCP credentials"
  
  # Load a saved gcloud credentials set
  gz dev-env gcloud-credentials load --name production
  
  # List all saved credentials
  gz dev-env gcloud-credentials list
  
  # Save from specific path
  gz dev-env gcloud-credentials save --name custom --config-path /path/to/gcloud`,
		SilenceUsage: true,
	}

	cmd.AddCommand(newGcloudCredentialsSaveCmd())
	cmd.AddCommand(newGcloudCredentialsLoadCmd())
	cmd.AddCommand(newGcloudCredentialsListCmd())

	return cmd
}

func newGcloudCredentialsSaveCmd() *cobra.Command {
	o := defaultGcloudCredentialsOptions()

	cmd := &cobra.Command{
		Use:   "save",
		Short: "Save current gcloud credentials",
		Long: `Save the current Google Cloud credentials with a given name.

This creates a backup of your current Google Cloud credentials that can be
restored later using the 'load' command. The credentials include service
account keys, application default credentials, and other authentication data.

SECURITY WARNING: This stores sensitive credential information.

Examples:
  # Save current gcloud credentials as "production"
  gz dev-env gcloud-credentials save --name production
  
  # Save with description
  gz dev-env gcloud-credentials save --name staging --description "Staging environment"
  
  # Save from specific path
  gz dev-env gcloud-credentials save --name custom --config-path /path/to/gcloud`,
		RunE: o.runSave,
	}

	cmd.Flags().StringVarP(&o.name, "name", "n", "", "Name for the saved credentials (required)")
	cmd.Flags().StringVarP(&o.description, "description", "d", "", "Description for the credentials")
	cmd.Flags().StringVar(&o.configPath, "config-path", o.configPath, "Path to gcloud config directory to save")
	cmd.Flags().StringVar(&o.storePath, "store-path", o.storePath, "Directory to store saved credentials")
	cmd.Flags().BoolVarP(&o.force, "force", "f", false, "Overwrite existing saved credentials")

	cmd.MarkFlagRequired("name")

	return cmd
}

func newGcloudCredentialsLoadCmd() *cobra.Command {
	o := defaultGcloudCredentialsOptions()

	cmd := &cobra.Command{
		Use:   "load",
		Short: "Load a saved gcloud credentials set",
		Long: `Load a previously saved Google Cloud credentials set.

This restores Google Cloud credentials that were previously saved using the 'save' command.
The current gcloud credentials will be backed up before loading the new ones.

Examples:
  # Load the "production" credentials
  gz dev-env gcloud-credentials load --name production
  
  # Load to specific path
  gz dev-env gcloud-credentials load --name staging --config-path /path/to/gcloud`,
		RunE: o.runLoad,
	}

	cmd.Flags().StringVarP(&o.name, "name", "n", "", "Name of the saved credentials to load (required)")
	cmd.Flags().StringVar(&o.configPath, "config-path", o.configPath, "Path where to load the gcloud credentials")
	cmd.Flags().StringVar(&o.storePath, "store-path", o.storePath, "Directory where saved credentials are stored")
	cmd.Flags().BoolVarP(&o.force, "force", "f", false, "Skip backup of current credentials")

	cmd.MarkFlagRequired("name")

	return cmd
}

func newGcloudCredentialsListCmd() *cobra.Command {
	o := defaultGcloudCredentialsOptions()

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List saved gcloud credentials",
		Long: `List all saved Google Cloud credentials.

This shows all the gcloud credentials that have been saved using the 'save' command,
along with their descriptions, save dates, and credential information.

Examples:
  # List all saved credentials
  gz dev-env gcloud-credentials list`,
		RunE: o.runList,
	}

	cmd.Flags().StringVar(&o.storePath, "store-path", o.storePath, "Directory where saved credentials are stored")

	return cmd
}

func (o *gcloudCredentialsOptions) runSave(_ *cobra.Command, args []string) error {
	// Check if source credentials directory exists
	if _, err := os.Stat(o.configPath); os.IsNotExist(err) {
		return fmt.Errorf("gcloud config directory not found at %s", o.configPath)
	}

	// Create store directory if it doesn't exist with secure permissions
	if err := os.MkdirAll(o.storePath, 0o700); err != nil {
		return fmt.Errorf("failed to create store directory: %w", err)
	}

	// Check if name already exists
	savedPath := filepath.Join(o.storePath, o.name)
	if _, err := os.Stat(savedPath); err == nil && !o.force {
		return fmt.Errorf("credentials '%s' already exists. Use --force to overwrite", o.name)
	}

	// Remove existing directory if force is enabled
	if o.force {
		if err := os.RemoveAll(savedPath); err != nil {
			return fmt.Errorf("failed to remove existing credentials: %w", err)
		}
	}

	// Copy the gcloud credentials (only credential-related files)
	if err := o.copyCredentials(o.configPath, savedPath); err != nil {
		return fmt.Errorf("failed to save gcloud credentials: %w", err)
	}

	// Save metadata
	if err := o.saveMetadata(); err != nil {
		fmt.Printf("Warning: failed to save metadata: %v\n", err)
	}

	// Display credentials information (without showing sensitive data)
	if err := o.displayCredentialsInfo(savedPath); err != nil {
		fmt.Printf("Warning: failed to read credentials info: %v\n", err)
	}

	fmt.Printf("✅ Gcloud credentials saved as '%s'\n", o.name)

	if o.description != "" {
		fmt.Printf("   Description: %s\n", o.description)
	}

	fmt.Printf("   Saved to: %s\n", savedPath)

	return nil
}

func (o *gcloudCredentialsOptions) runLoad(_ *cobra.Command, args []string) error {
	// Check if saved credentials exist
	savedPath := filepath.Join(o.storePath, o.name)
	if _, err := os.Stat(savedPath); os.IsNotExist(err) {
		return fmt.Errorf("saved credentials '%s' not found", o.name)
	}

	// Backup current credentials if they exist and force is not set
	if !o.force {
		if _, err := os.Stat(o.configPath); err == nil {
			backupPath := o.configPath + ".backup." + time.Now().Format("20060102-150405")
			if err := o.copyCredentials(o.configPath, backupPath); err != nil {
				return fmt.Errorf("failed to backup current gcloud credentials: %w", err)
			}

			fmt.Printf("📦 Current gcloud credentials backed up to: %s\n", backupPath)
		}
	}

	// Create target directory if it doesn't exist
	if err := os.MkdirAll(o.configPath, 0o700); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Copy the saved credentials to target location (merge into existing gcloud config)
	if err := o.mergeCredentials(savedPath, o.configPath); err != nil {
		return fmt.Errorf("failed to load gcloud credentials: %w", err)
	}

	// Display credentials information
	if err := o.displayCredentialsInfo(o.configPath); err != nil {
		fmt.Printf("Warning: failed to read credentials info: %v\n", err)
	}

	fmt.Printf("✅ Gcloud credentials '%s' loaded successfully\n", o.name)
	fmt.Printf("   Loaded to: %s\n", o.configPath)

	return nil
}

func (o *gcloudCredentialsOptions) runList(_ *cobra.Command, args []string) error {
	// Check if store directory exists
	if _, err := os.Stat(o.storePath); os.IsNotExist(err) {
		fmt.Println("No saved gcloud credentials found.")
		return nil
	}

	// Read directory contents
	entries, err := os.ReadDir(o.storePath)
	if err != nil {
		return fmt.Errorf("failed to read store directory: %w", err)
	}

	// Filter for directories (excluding .meta files)
	var credentials []string

	for _, entry := range entries {
		if entry.IsDir() {
			credentials = append(credentials, entry.Name())
		}
	}

	if len(credentials) == 0 {
		fmt.Println("No saved gcloud credentials found.")
		return nil
	}

	fmt.Printf("Saved gcloud credentials (%d):\n\n", len(credentials))

	// Read and display metadata for each credential set
	for _, name := range credentials {
		metadata := o.loadMetadata(name)
		fmt.Printf("🔐 %s\n", name)

		if metadata.Description != "" {
			fmt.Printf("   Description: %s\n", metadata.Description)
		}

		if !metadata.SavedAt.IsZero() {
			fmt.Printf("   Saved: %s\n", metadata.SavedAt.Format("2006-01-02 15:04:05"))
		}

		credentialsPath := filepath.Join(o.storePath, name)
		if info, err := os.Stat(credentialsPath); err == nil && info.IsDir() {
			// Get directory size
			size, err := o.getDirSize(credentialsPath)
			if err == nil {
				fmt.Printf("   Size: %d bytes\n", size)
			}
		}

		// Display credentials information (without sensitive data)
		if err := o.displayCredentialsInfo(credentialsPath); err == nil {
			// Already displayed in displayCredentialsInfo
		}

		fmt.Println()
	}

	return nil
}

type gcloudCredentialsMetadata struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	SavedAt     time.Time `json:"saved_at"`
	SourcePath  string    `json:"source_path"`
}

func (o *gcloudCredentialsOptions) saveMetadata() error {
	metadata := gcloudCredentialsMetadata{
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

	// Set secure permissions for metadata
	if err := os.Chmod(metadataPath, 0o600); err != nil {
		return err
	}

	// Write metadata as simple key-value pairs
	if metadata.Description != "" {
		fmt.Fprintf(file, "description=%s\n", metadata.Description)
	}

	fmt.Fprintf(file, "saved_at=%s\n", metadata.SavedAt.Format(time.RFC3339))
	fmt.Fprintf(file, "source_path=%s\n", metadata.SourcePath)

	return nil
}

func (o *gcloudCredentialsOptions) loadMetadata(name string) gcloudCredentialsMetadata {
	metadata := gcloudCredentialsMetadata{Name: name}

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

// copyCredentials copies only credential-related files, not configuration.
func (o *gcloudCredentialsOptions) copyCredentials(src, dst string) error {
	// Create destination directory with secure permissions
	if err := os.MkdirAll(dst, 0o700); err != nil {
		return err
	}

	// List of credential-related files and directories to copy
	credentialPaths := []string{
		"application_default_credentials.json",
		"legacy_credentials",
		"credentials.db",
		"access_tokens.db",
		"gce",
	}

	for _, credPath := range credentialPaths {
		srcPath := filepath.Join(src, credPath)
		dstPath := filepath.Join(dst, credPath)

		// Check if source exists
		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			continue // Skip if doesn't exist
		}

		// Copy file or directory
		if info, err := os.Stat(srcPath); err == nil {
			if info.IsDir() {
				if err := o.copyDir(srcPath, dstPath); err != nil {
					return fmt.Errorf("failed to copy credential directory %s: %w", credPath, err)
				}
			} else {
				if err := o.copyFile(srcPath, dstPath); err != nil {
					return fmt.Errorf("failed to copy credential file %s: %w", credPath, err)
				}
			}
		}
	}

	return nil
}

// mergeCredentials merges saved credentials into existing gcloud config.
func (o *gcloudCredentialsOptions) mergeCredentials(src, dst string) error {
	// Read source directory
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	// Copy each credential file/directory
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Remove existing directory and copy new one
			os.RemoveAll(dstPath)

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

func (o *gcloudCredentialsOptions) copyDir(src, dst string) error {
	// Get source directory info
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// Create destination directory with secure permissions
	if err := os.MkdirAll(dst, 0o700); err != nil {
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

	// Set directory permissions
	return os.Chmod(dst, srcInfo.Mode())
}

func (o *gcloudCredentialsOptions) copyFile(src, dst string) error {
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

	// Set secure permissions for credential files
	if err := os.Chmod(dst, 0o600); err != nil {
		return err
	}

	_, err = io.Copy(destFile, sourceFile)

	return err
}

func (o *gcloudCredentialsOptions) getDirSize(path string) (int64, error) {
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

func (o *gcloudCredentialsOptions) displayCredentialsInfo(credentialsPath string) error {
	var (
		credentialFiles []string
		serviceAccounts []string
	)

	// Check for application default credentials
	adcPath := filepath.Join(credentialsPath, "application_default_credentials.json")
	if _, err := os.Stat(adcPath); err == nil {
		credentialFiles = append(credentialFiles, "Application Default Credentials")

		// Try to read ADC for account info (without exposing sensitive data)
		if content, err := os.ReadFile(adcPath); err == nil {
			var adc map[string]interface{}
			if json.Unmarshal(content, &adc) == nil {
				if clientEmail, ok := adc["client_email"].(string); ok {
					serviceAccounts = append(serviceAccounts, clientEmail)
				}
			}
		}
	}

	// Check for legacy credentials
	legacyPath := filepath.Join(credentialsPath, "legacy_credentials")
	if _, err := os.Stat(legacyPath); err == nil {
		credentialFiles = append(credentialFiles, "Legacy Credentials")
	}

	// Check for credentials database
	credDbPath := filepath.Join(credentialsPath, "credentials.db")
	if _, err := os.Stat(credDbPath); err == nil {
		credentialFiles = append(credentialFiles, "Credentials Database")
	}

	// Check for access tokens database
	tokensDbPath := filepath.Join(credentialsPath, "access_tokens.db")
	if _, err := os.Stat(tokensDbPath); err == nil {
		credentialFiles = append(credentialFiles, "Access Tokens Database")
	}

	// Check for GCE credentials
	gcePath := filepath.Join(credentialsPath, "gce")
	if _, err := os.Stat(gcePath); err == nil {
		credentialFiles = append(credentialFiles, "GCE Credentials")
	}

	// Display information
	if len(credentialFiles) > 0 {
		fmt.Printf("   Credential types: %d found\n", len(credentialFiles))

		for _, credFile := range credentialFiles {
			fmt.Printf("     - %s\n", credFile)
		}
	}

	if len(serviceAccounts) > 0 {
		fmt.Printf("   Service accounts: %d found\n", len(serviceAccounts))

		for _, sa := range serviceAccounts {
			fmt.Printf("     - %s\n", sa)
		}
	}

	return nil
}
