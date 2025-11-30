// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package devenv

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// CreateInstallKeySimpleCommand creates a simple install-key command using system SSH.
func (c *EnhancedSSHCommand) CreateInstallKeySimpleCommand() *cobra.Command {
	var (
		host          string
		user          string
		publicKeyPath string
	)

	cmd := &cobra.Command{
		Use:   "install-key-simple",
		Short: "Install SSH public key using system SSH (simple)",
		Long: `Install SSH public key to remote server using system SSH command.

This is a simpler alternative that uses the system's SSH command directly,
which should work with any SSH configuration that allows password authentication.

Examples:
  # Install specific public key (will prompt for password)
  gz dev-env ssh install-key-simple --host server.com --user admin --public-key ~/.ssh/id_rsa.pub`,
		RunE: func(cmd *cobra.Command, args []string) error {
			installer := NewSimpleSSHInstaller()
			return installer.InstallPublicKeySimple(host, user, publicKeyPath)
		},
	}

	cmd.Flags().StringVar(&host, "host", "", "Remote server hostname or IP (required)")
	cmd.Flags().StringVar(&user, "user", "", "Username for SSH connection (required)")
	cmd.Flags().StringVar(&publicKeyPath, "public-key", "", "Path to public key file (required)")

	cmd.MarkFlagRequired("host")
	cmd.MarkFlagRequired("user")
	cmd.MarkFlagRequired("public-key")

	return cmd
}

// CreateInstallKeyCommand creates the install-key command for SSH.
func (c *EnhancedSSHCommand) CreateInstallKeyCommand() *cobra.Command {
	var (
		host           string
		port           string
		user           string
		publicKeyPath  string
		privateKeyPath string
		password       string
		configName     string
		force          bool
		dryRun         bool
	)

	cmd := &cobra.Command{
		Use:   "install-key",
		Short: "Install SSH public key to remote server",
		Long: `Install SSH public key to remote server's authorized_keys file.

This command can install keys in two ways:
1. Install a specific public key file to a remote server
2. Install all keys from a saved SSH configuration to a remote server

The command will:
- Connect to the remote server using SSH
- Add the public key to ~/.ssh/authorized_keys
- Set proper permissions on SSH files
- Avoid duplicate keys (unless --force is used)

Authentication methods tried in order:
1. Private key authentication (if private key is available)
2. Password authentication (if --password is provided)
3. Interactive password prompt

Examples:
  # Install specific public key
  gz dev-env ssh install-key --host server.com --user admin --public-key ~/.ssh/id_rsa.pub

  # Install key with password authentication
  gz dev-env ssh install-key --host server.com --user admin --public-key ~/.ssh/id_rsa.pub --password mypass

  # Install all keys from saved configuration
  gz dev-env ssh install-key --config production --host server.com --user admin

  # Dry run to see what would be installed
  gz dev-env ssh install-key --config production --host server.com --user admin --dry-run`,
		RunE: func(cmd *cobra.Command, args []string) error {
			installer := NewSSHKeyInstaller()

			// Set verbose mode based on global flag
			verbose, _ := cmd.Flags().GetBool("verbose")
			installer.SetVerbose(verbose)

			var err error
			if configName != "" {
				err = c.installKeysFromConfig(installer, configName, host, user, port, password, force, dryRun)
			} else {
				err = c.installSingleKey(installer, host, port, user, publicKeyPath, privateKeyPath, password, force, dryRun)
			}

			if err != nil {
				// Don't show usage on execution errors
				cmd.SilenceUsage = true
			}
			return err
		},
	}

	cmd.Flags().StringVar(&host, "host", "", "Remote server hostname or IP (required)")
	cmd.Flags().StringVar(&port, "port", "22", "SSH port")
	cmd.Flags().StringVar(&user, "user", "", "Username for SSH connection (required)")
	cmd.Flags().StringVar(&publicKeyPath, "public-key", "", "Path to public key file")
	cmd.Flags().StringVar(&privateKeyPath, "private-key", "", "Path to private key file (for authentication)")
	cmd.Flags().StringVar(&password, "password", "", "Password for SSH connection")
	cmd.Flags().StringVar(&configName, "config", "", "Name of saved SSH configuration to install keys from")
	cmd.Flags().BoolVar(&force, "force", false, "Force install even if key already exists")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be done without making changes")

	cmd.MarkFlagRequired("host")
	cmd.MarkFlagRequired("user")

	// Either public-key or config must be specified
	cmd.MarkFlagsOneRequired("public-key", "config")

	return cmd
}

// installSingleKey installs a single public key to a remote server.
func (c *EnhancedSSHCommand) installSingleKey(installer *SSHKeyInstaller, host, port, user, publicKeyPath, privateKeyPath, password string, force, dryRun bool) error {
	// If no private key specified, try to find corresponding private key
	if privateKeyPath == "" && publicKeyPath != "" {
		if strings.HasSuffix(publicKeyPath, ".pub") {
			potentialPrivateKey := strings.TrimSuffix(publicKeyPath, ".pub")
			if _, err := os.Stat(potentialPrivateKey); err == nil {
				privateKeyPath = potentialPrivateKey
			}
		}
	}

	opts := &InstallOptions{
		Host:           host,
		Port:           port,
		User:           user,
		PublicKeyPath:  publicKeyPath,
		PrivateKeyPath: privateKeyPath,
		Password:       password,
		Force:          force,
		DryRun:         dryRun,
	}

	result, err := installer.InstallPublicKey(opts)
	if err != nil {
		return err
	}

	c.printInstallResult(result)
	return nil
}

// installKeysFromConfig installs all keys from a saved SSH configuration.
func (c *EnhancedSSHCommand) installKeysFromConfig(installer *SSHKeyInstaller, configName, host, user, port, password string, force, dryRun bool) error {
	opts := &InstallOptions{
		Port:     port,
		Password: password,
		Force:    force,
		DryRun:   dryRun,
	}

	results, err := installer.InstallKeysFromConfig(configName, host, user, opts)
	if err != nil {
		return err
	}

	fmt.Printf("Installing keys from configuration '%s' to %s@%s:\n\n", configName, user, host)

	successCount := 0
	for _, result := range results {
		c.printInstallResult(result)
		if result.Success {
			successCount++
		}
	}

	fmt.Printf("\n📊 Summary: %d/%d keys processed successfully\n", successCount, len(results))
	return nil
}

// printInstallResult prints the result of a key installation.
func (c *EnhancedSSHCommand) printInstallResult(result *InstallResult) {
	if result.Success {
		if result.KeyAdded {
			fmt.Printf("✅ %s\n", result.Message)
		} else if result.KeyExists {
			fmt.Printf("ℹ️  %s\n", result.Message)
		} else {
			fmt.Printf("✅ %s\n", result.Message)
		}
	} else {
		fmt.Printf("❌ %s\n", result.Message)
	}
}

// CreateListKeysCommand creates the list-keys command to show available keys.
func (c *EnhancedSSHCommand) CreateListKeysCommand() *cobra.Command {
	var configName string

	cmd := &cobra.Command{
		Use:   "list-keys",
		Short: "List available SSH keys from saved configuration",
		Long: `List all public keys available in a saved SSH configuration.

This command shows:
- Key file names and paths
- Key types (RSA, ED25519, etc.)
- Key fingerprints
- Whether corresponding private keys exist

Examples:
  # List keys from a specific configuration
  gz dev-env ssh list-keys --config production

  # List keys from all configurations
  gz dev-env ssh list-keys`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.listKeys(configName)
		},
	}

	cmd.Flags().StringVar(&configName, "config", "", "Name of saved SSH configuration")

	return cmd
}

// listKeys lists available SSH keys from configurations.
func (c *EnhancedSSHCommand) listKeys(configName string) error {
	homeDir, _ := os.UserHomeDir()
	storeDir := filepath.Join(homeDir, ".gz", "ssh-configs")

	if configName != "" {
		return c.listKeysFromConfig(storeDir, configName)
	}

	// List keys from all configurations
	entries, err := os.ReadDir(storeDir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No SSH configurations found")
			return nil
		}
		return fmt.Errorf("failed to read configurations: %w", err)
	}

	fmt.Println("Available SSH keys from all configurations:")
	fmt.Println()

	for _, entry := range entries {
		if entry.IsDir() {
			fmt.Printf("📁 Configuration: %s\n", entry.Name())
			if err := c.listKeysFromConfig(storeDir, entry.Name()); err != nil {
				fmt.Printf("   Error: %v\n", err)
			}
			fmt.Println()
		}
	}

	return nil
}

// listKeysFromConfig lists keys from a specific configuration.
func (c *EnhancedSSHCommand) listKeysFromConfig(storeDir, configName string) error {
	metadataFile := filepath.Join(storeDir, configName, "metadata.json")
	metadata, err := c.loadEnhancedMetadata(metadataFile)
	if err != nil {
		return fmt.Errorf("failed to load configuration '%s': %w", configName, err)
	}

	if len(metadata.PublicKeys) == 0 {
		fmt.Printf("   No public keys found in configuration '%s'\n", configName)
		return nil
	}

	keysDir := filepath.Join(storeDir, configName, "keys")

	fmt.Printf("   Public keys (%d found):\n", len(metadata.PublicKeys))

	for _, originalKeyPath := range metadata.PublicKeys {
		keyName := filepath.Base(originalKeyPath)
		publicKeyPath := filepath.Join(keysDir, keyName)
		privateKeyName := strings.TrimSuffix(keyName, ".pub")
		privateKeyPath := filepath.Join(keysDir, privateKeyName)

		// Read public key for type detection
		keyType := "unknown"
		fingerprint := "unknown"
		if content, err := os.ReadFile(publicKeyPath); err == nil {
			keyStr := strings.TrimSpace(string(content))
			if parts := strings.Fields(keyStr); len(parts) >= 2 {
				keyType = parts[0]
				// Simple fingerprint approximation
				if len(parts[1]) >= 16 {
					fingerprint = parts[1][:16] + "..."
				}
			}
		}

		// Check if private key exists
		hasPrivateKey := ""
		if _, err := os.Stat(privateKeyPath); err == nil {
			hasPrivateKey = " (with private key)"
		}

		fmt.Printf("     🔑 %s - %s%s\n", keyName, keyType, hasPrivateKey)
		fmt.Printf("         Fingerprint: %s\n", fingerprint)
		fmt.Printf("         Original: %s\n", originalKeyPath)
	}

	return nil
}
