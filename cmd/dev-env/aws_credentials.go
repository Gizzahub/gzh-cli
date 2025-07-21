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

type awsCredentialsOptions struct {
	name        string
	description string
	configPath  string
	storePath   string
	force       bool
	listAll     bool
}

func defaultAwsCredentialsOptions() *awsCredentialsOptions {
	homeDir, _ := os.UserHomeDir()

	return &awsCredentialsOptions{
		configPath: filepath.Join(homeDir, ".aws", "credentials"),
		storePath:  filepath.Join(homeDir, ".gz", "aws-credentials"),
	}
}

func newAwsCredentialsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "aws-credentials",
		Short: "Manage AWS credentials files",
		Long: `Save and load AWS credentials files.

This command helps you backup and restore AWS credentials files, which contain:
- AWS access keys and secret keys
- Session tokens for temporary credentials
- Profile-specific credentials
- MFA and role assumption credentials
- Other AWS authentication settings

This is useful when:
- Setting up new development machines
- Switching between different AWS credential sets
- Backing up AWS credentials before changes
- Managing multiple AWS credential sets for different projects/environments

The credentials are saved to ~/.gz/aws-credentials/ by default.

SECURITY WARNING: This stores sensitive credential information. Ensure your
storage location is properly secured and consider encrypting the stored files.

Examples:
  # Save current AWS credentials with a name
  gz dev-env aws-credentials save --name production

  # Save with description
  gz dev-env aws-credentials save --name staging --description "Staging AWS credentials"

  # Load a saved AWS credentials file
  gz dev-env aws-credentials load --name production

  # List all saved credentials
  gz dev-env aws-credentials list

  # Save from specific path
  gz dev-env aws-credentials save --name custom --config-path /path/to/credentials`,
		SilenceUsage: true,
	}

	cmd.AddCommand(newAwsCredentialsSaveCmd())
	cmd.AddCommand(newAwsCredentialsLoadCmd())
	cmd.AddCommand(newAwsCredentialsListCmd())

	return cmd
}

func newAwsCredentialsSaveCmd() *cobra.Command {
	o := defaultAwsCredentialsOptions()

	cmd := &cobra.Command{
		Use:   "save",
		Short: "Save current AWS credentials",
		Long: `Save the current AWS credentials file with a given name.

This creates a backup of your current AWS credentials that can be
restored later using the 'load' command. The credentials include
access keys, secret keys, session tokens, and profile configurations.

SECURITY WARNING: This stores sensitive credential information.

Examples:
  # Save current AWS credentials as "production"
  gz dev-env aws-credentials save --name production

  # Save with description
  gz dev-env aws-credentials save --name staging --description "Staging environment"

  # Save from specific path
  gz dev-env aws-credentials save --name custom --config-path /path/to/credentials`,
		RunE: o.runSave,
	}

	cmd.Flags().StringVarP(&o.name, "name", "n", "", "Name for the saved credentials (required)")
	cmd.Flags().StringVarP(&o.description, "description", "d", "", "Description for the credentials")
	cmd.Flags().StringVar(&o.configPath, "config-path", o.configPath, "Path to AWS credentials file to save")
	cmd.Flags().StringVar(&o.storePath, "store-path", o.storePath, "Directory to store saved credentials")
	cmd.Flags().BoolVarP(&o.force, "force", "f", false, "Overwrite existing saved credentials")

	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func newAwsCredentialsLoadCmd() *cobra.Command {
	o := defaultAwsCredentialsOptions()

	cmd := &cobra.Command{
		Use:   "load",
		Short: "Load a saved AWS credentials file",
		Long: `Load a previously saved AWS credentials file.

This restores AWS credentials that were previously saved using the 'save' command.
The current AWS credentials will be backed up before loading the new ones.

Examples:
  # Load the "production" credentials
  gz dev-env aws-credentials load --name production

  # Load to specific path
  gz dev-env aws-credentials load --name staging --config-path /path/to/credentials`,
		RunE: o.runLoad,
	}

	cmd.Flags().StringVarP(&o.name, "name", "n", "", "Name of the saved credentials to load (required)")
	cmd.Flags().StringVar(&o.configPath, "config-path", o.configPath, "Path where to load the AWS credentials")
	cmd.Flags().StringVar(&o.storePath, "store-path", o.storePath, "Directory where saved credentials are stored")
	cmd.Flags().BoolVarP(&o.force, "force", "f", false, "Skip backup of current credentials")

	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func newAwsCredentialsListCmd() *cobra.Command {
	o := defaultAwsCredentialsOptions()

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List saved AWS credentials",
		Long: `List all saved AWS credentials files.

This shows all the AWS credentials files that have been saved using the 'save' command,
along with their descriptions, save dates, and profile information.

Examples:
  # List all saved credentials
  gz dev-env aws-credentials list`,
		RunE: o.runList,
	}

	cmd.Flags().StringVar(&o.storePath, "store-path", o.storePath, "Directory where saved credentials are stored")

	return cmd
}

func (o *awsCredentialsOptions) runSave(_ *cobra.Command, _ []string) error {
	// Check if source credentials file exists
	if _, err := os.Stat(o.configPath); os.IsNotExist(err) {
		return fmt.Errorf("AWS credentials file not found at %s", o.configPath)
	}

	// Create store directory if it doesn't exist
	if err := os.MkdirAll(o.storePath, 0o700); err != nil { // Use 0700 for credentials security
		return fmt.Errorf("failed to create store directory: %w", err)
	}

	// Check if name already exists
	savedPath := filepath.Join(o.storePath, o.name+".credentials")
	if _, err := os.Stat(savedPath); err == nil && !o.force {
		return fmt.Errorf("credentials '%s' already exists. Use --force to overwrite", o.name)
	}

	// Copy the AWS credentials file
	if err := o.copyFile(o.configPath, savedPath); err != nil {
		return fmt.Errorf("failed to save AWS credentials: %w", err)
	}

	// Save metadata
	if err := o.saveMetadata(); err != nil {
		fmt.Printf("Warning: failed to save metadata: %v\n", err)
	}

	// Display credentials information (without showing sensitive data)
	if err := o.displayCredentialsInfo(savedPath); err != nil {
		fmt.Printf("Warning: failed to read credentials info: %v\n", err)
	}

	fmt.Printf("✅ AWS credentials saved as '%s'\n", o.name)

	if o.description != "" {
		fmt.Printf("   Description: %s\n", o.description)
	}

	fmt.Printf("   Saved to: %s\n", savedPath)

	return nil
}

func (o *awsCredentialsOptions) runLoad(_ *cobra.Command, _ []string) error {
	// Check if saved credentials exist
	savedPath := filepath.Join(o.storePath, o.name+".credentials")
	if _, err := os.Stat(savedPath); os.IsNotExist(err) {
		return fmt.Errorf("saved credentials '%s' not found", o.name)
	}

	// Backup current credentials if they exist and force is not set
	if !o.force {
		if _, err := os.Stat(o.configPath); err == nil {
			backupPath := o.configPath + ".backup." + time.Now().Format("20060102-150405")
			if err := o.copyFile(o.configPath, backupPath); err != nil {
				return fmt.Errorf("failed to backup current AWS credentials: %w", err)
			}

			fmt.Printf("📦 Current AWS credentials backed up to: %s\n", backupPath)
		}
	}

	// Create target directory if it doesn't exist
	targetDir := filepath.Dir(o.configPath)
	if err := os.MkdirAll(targetDir, 0o700); err != nil { // Use 0700 for credentials security
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Copy the saved credentials to target location
	if err := o.copyFile(savedPath, o.configPath); err != nil {
		return fmt.Errorf("failed to load AWS credentials: %w", err)
	}

	// Display credentials information
	if err := o.displayCredentialsInfo(o.configPath); err != nil {
		fmt.Printf("Warning: failed to read credentials info: %v\n", err)
	}

	fmt.Printf("✅ AWS credentials '%s' loaded successfully\n", o.name)
	fmt.Printf("   Loaded to: %s\n", o.configPath)

	return nil
}

func (o *awsCredentialsOptions) runList(_ *cobra.Command, _ []string) error {
	// Check if store directory exists
	if _, err := os.Stat(o.storePath); os.IsNotExist(err) {
		fmt.Println("No saved AWS credentials found.")
		return nil
	}

	// Read directory contents
	entries, err := os.ReadDir(o.storePath)
	if err != nil {
		return fmt.Errorf("failed to read store directory: %w", err)
	}

	// Filter for .credentials files
	var credentials []string

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".credentials") {
			name := strings.TrimSuffix(entry.Name(), ".credentials")
			credentials = append(credentials, name)
		}
	}

	if len(credentials) == 0 {
		fmt.Println("No saved AWS credentials found.")
		return nil
	}

	fmt.Printf("Saved AWS credentials (%d):\n\n", len(credentials))

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

		credentialsPath := filepath.Join(o.storePath, name+".credentials")
		if info, err := os.Stat(credentialsPath); err == nil {
			fmt.Printf("   Size: %d bytes\n", info.Size())
		}

		// Display profile information (without sensitive data)
		_ = o.displayCredentialsInfo(credentialsPath)

		fmt.Println()
	}

	return nil
}

type awsCredentialsMetadata struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	SavedAt     time.Time `json:"savedAt"`
	SourcePath  string    `json:"sourcePath"`
}

func (o *awsCredentialsOptions) saveMetadata() error {
	metadata := awsCredentialsMetadata{
		Name:        o.name,
		Description: o.description,
		SavedAt:     time.Now(),
		SourcePath:  o.configPath,
	}

	metadataPath := filepath.Join(o.storePath, o.name+".meta")

	file, err := os.Create(metadataPath) //nolint:gosec // Safe file path construction
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			// Log error but don't override main error
		}
	}()

	// Set secure permissions for metadata
	if err := os.Chmod(metadataPath, 0o600); err != nil {
		return err
	}

	// Write metadata as simple key-value pairs
	if metadata.Description != "" {
		if _, err := fmt.Fprintf(file, "description=%s\n", metadata.Description); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintf(file, "saved_at=%s\n", metadata.SavedAt.Format(time.RFC3339)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(file, "source_path=%s\n", metadata.SourcePath); err != nil {
		return err
	}

	return nil
}

func (o *awsCredentialsOptions) loadMetadata(name string) awsCredentialsMetadata {
	metadata := awsCredentialsMetadata{Name: name}

	metadataPath := filepath.Join(o.storePath, name+".meta")

	content, err := os.ReadFile(metadataPath) //nolint:gosec // Safe file path construction
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

func (o *awsCredentialsOptions) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src) //nolint:gosec // Safe file path construction
	if err != nil {
		return err
	}
	defer func() {
		if err := sourceFile.Close(); err != nil {
			// Log error but don't override main error
		}
	}()

	destFile, err := os.Create(dst) //nolint:gosec // dst parameter from controlled path construction
	if err != nil {
		return err
	}
	defer func() {
		if err := destFile.Close(); err != nil {
			// Log error but don't override main error
		}
	}()

	// Set secure permissions for credentials
	if err := os.Chmod(dst, 0o600); err != nil {
		return err
	}

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	return nil
}

func (o *awsCredentialsOptions) displayCredentialsInfo(credentialsPath string) error {
	// Read and parse AWS credentials
	content, err := os.ReadFile(credentialsPath) //nolint:gosec // credentialsPath from controlled path construction
	if err != nil {
		return err
	}

	// Parse AWS credentials file to extract profile information (without showing sensitive data)
	profiles := o.parseAwsCredentials(string(content))

	// Display profile information (safe info only)
	if len(profiles) > 0 {
		fmt.Printf("   Profiles: %d configured\n", len(profiles))

		for _, profile := range profiles {
			fmt.Printf("     - %s", profile.Name)

			if profile.HasSessionToken {
				fmt.Printf(" (with session token)")
			}

			if profile.HasMfaSerial {
				fmt.Printf(" (with MFA)")
			}

			fmt.Println()
		}
	}

	return nil
}

type awsCredentialsProfile struct {
	Name            string
	HasAccessKey    bool
	HasSecretKey    bool
	HasSessionToken bool
	HasMfaSerial    bool
}

func (o *awsCredentialsOptions) parseAwsCredentials(content string) []awsCredentialsProfile {
	var (
		profiles       []awsCredentialsProfile
		currentProfile *awsCredentialsProfile
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

			currentProfile = &awsCredentialsProfile{Name: profileName}

			continue
		}

		// Parse key-value pairs (without storing sensitive values)
		if currentProfile != nil && strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])

				switch key {
				case "aws_access_key_id":
					currentProfile.HasAccessKey = true
				case "aws_secret_access_key":
					currentProfile.HasSecretKey = true
				case "aws_session_token":
					currentProfile.HasSessionToken = true
				case "mfa_serial":
					currentProfile.HasMfaSerial = true
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
