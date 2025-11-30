// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package devenv

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

// SSHKeyInstaller handles installing public keys to remote servers.
type SSHKeyInstaller struct {
	timeout time.Duration
	verbose bool
}

// NewSSHKeyInstaller creates a new SSH key installer.
func NewSSHKeyInstaller() *SSHKeyInstaller {
	return &SSHKeyInstaller{
		timeout: 30 * time.Second,
		verbose: false,
	}
}

// SetVerbose sets the verbose mode for the installer.
func (installer *SSHKeyInstaller) SetVerbose(verbose bool) {
	installer.verbose = verbose
}

// InstallOptions represents options for installing SSH keys.
type InstallOptions struct {
	Host           string
	Port           string
	User           string
	PublicKeyPath  string
	PrivateKeyPath string
	Password       string
	Force          bool
	DryRun         bool
}

// InstallResult represents the result of a key installation.
type InstallResult struct {
	Host      string
	Success   bool
	Message   string
	KeyAdded  bool
	KeyExists bool
}

// InstallPublicKey installs a public key to a remote server.
func (installer *SSHKeyInstaller) InstallPublicKey(opts *InstallOptions) (*InstallResult, error) {
	result := &InstallResult{
		Host: opts.Host,
	}

	// Validate inputs
	if err := installer.validateOptions(opts); err != nil {
		result.Message = err.Error()
		return result, err
	}

	// Read public key
	publicKey, err := installer.readPublicKey(opts.PublicKeyPath)
	if err != nil {
		result.Message = fmt.Sprintf("Failed to read public key: %v", err)
		return result, err
	}

	if opts.DryRun {
		result.Success = true
		result.Message = fmt.Sprintf("DRY RUN: Would install key from %s to %s@%s",
			opts.PublicKeyPath, opts.User, opts.Host)
		return result, nil
	}

	// Create SSH client
	client, err := installer.createSSHClient(opts)
	if err != nil {
		result.Message = fmt.Sprintf("Failed to connect: %v", err)
		return result, err
	}
	defer client.Close()

	// Check if key already exists
	if installer.verbose {
		fmt.Printf("Checking if key already exists on remote server...\n")
	}
	exists, err := installer.keyExists(client, publicKey)
	if err != nil {
		result.Message = fmt.Sprintf("Failed to check existing keys: %v", err)
		if installer.verbose {
			fmt.Printf("Key existence check error: %v\n", err)
		}
		return result, err
	}
	if installer.verbose {
		fmt.Printf("Key existence check completed. Key exists: %v\n", exists)
	}

	result.KeyExists = exists
	if exists && !opts.Force {
		result.Success = true
		result.Message = "Public key already exists on remote server"
		return result, nil
	}

	// Install the key
	if installer.verbose {
		fmt.Printf("Installing key to remote server...\n")
	}
	if err := installer.installKey(client, publicKey); err != nil {
		result.Message = fmt.Sprintf("Failed to install key: %v", err)
		if installer.verbose {
			fmt.Printf("Install key error: %v\n", err)
		}
		return result, err
	}

	result.Success = true
	result.KeyAdded = true
	if exists {
		result.Message = "Public key updated on remote server"
	} else {
		result.Message = "Public key installed successfully on remote server"
	}

	return result, nil
}

// InstallKeysFromConfig installs all keys from a saved SSH configuration.
func (installer *SSHKeyInstaller) InstallKeysFromConfig(configName, host, user string, opts *InstallOptions) ([]*InstallResult, error) {
	// Load SSH configuration
	enhancedCmd := NewEnhancedSSHCommand()

	// Allow override of store directory for testing
	storeDir := ""
	if opts.Host != host { // Hack: if Host differs from host param, use it as store dir
		storeDir = opts.Host
		opts.Host = host // Restore original host
	} else {
		homeDir, _ := os.UserHomeDir()
		storeDir = filepath.Join(homeDir, ".gz", "ssh-configs")
	}

	metadataFile := filepath.Join(storeDir, configName, "metadata.json")
	metadata, err := enhancedCmd.loadEnhancedMetadata(metadataFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration '%s': %w", configName, err)
	}

	if len(metadata.PublicKeys) == 0 {
		return nil, fmt.Errorf("no public keys found in configuration '%s'", configName)
	}

	var results []*InstallResult
	keysDir := filepath.Join(storeDir, configName, "keys")

	for _, originalKeyPath := range metadata.PublicKeys {
		keyName := filepath.Base(originalKeyPath)
		publicKeyPath := filepath.Join(keysDir, keyName)
		privateKeyPath := strings.TrimSuffix(publicKeyPath, ".pub")

		installOpts := &InstallOptions{
			Host:           host,
			Port:           opts.Port,
			User:           user,
			PublicKeyPath:  publicKeyPath,
			PrivateKeyPath: privateKeyPath,
			Password:       opts.Password,
			Force:          opts.Force,
			DryRun:         opts.DryRun,
		}

		result, err := installer.InstallPublicKey(installOpts)
		if err != nil {
			result = &InstallResult{
				Host:    host,
				Success: false,
				Message: err.Error(),
			}
		}

		// Add key information to result
		result.Message = fmt.Sprintf("[%s] %s", keyName, result.Message)
		results = append(results, result)
	}

	return results, nil
}

// validateOptions validates installation options.
func (installer *SSHKeyInstaller) validateOptions(opts *InstallOptions) error {
	if opts.Host == "" {
		return fmt.Errorf("host is required")
	}
	if opts.User == "" {
		return fmt.Errorf("user is required")
	}
	if opts.PublicKeyPath == "" {
		return fmt.Errorf("public key path is required")
	}
	if _, err := os.Stat(opts.PublicKeyPath); os.IsNotExist(err) {
		return fmt.Errorf("public key file not found: %s", opts.PublicKeyPath)
	}
	if opts.Port == "" {
		opts.Port = "22"
	}
	return nil
}

// readPublicKey reads and validates a public key file.
func (installer *SSHKeyInstaller) readPublicKey(keyPath string) (string, error) {
	content, err := os.ReadFile(keyPath)
	if err != nil {
		return "", err
	}

	key := strings.TrimSpace(string(content))
	if key == "" {
		return "", fmt.Errorf("public key file is empty")
	}

	// Basic validation - should start with ssh-rsa, ssh-ed25519, etc.
	if !strings.HasPrefix(key, "ssh-") {
		return "", fmt.Errorf("invalid public key format")
	}

	return key, nil
}

// createSSHClient creates an SSH client connection.
func (installer *SSHKeyInstaller) createSSHClient(opts *InstallOptions) (*ssh.Client, error) {
	var authMethods []ssh.AuthMethod
	var password string

	// Try private key authentication first if available
	if opts.PrivateKeyPath != "" {
		if _, err := os.Stat(opts.PrivateKeyPath); err == nil {
			if keyAuth, err := installer.createKeyAuth(opts.PrivateKeyPath); err == nil {
				authMethods = append(authMethods, keyAuth)
			}
		}
	}

	// Get password if not provided
	if opts.Password != "" {
		password = opts.Password
		if installer.verbose {
			fmt.Printf("Using provided password\n")
		}
	} else {
		fmt.Printf("Password for %s@%s: ", opts.User, opts.Host)
		var err error
		password, err = installer.readPassword()
		if err != nil {
			return nil, fmt.Errorf("failed to read password: %w", err)
		}
		if password == "" {
			return nil, fmt.Errorf("empty password entered")
		}
		if installer.verbose {
			fmt.Printf("Password read successfully (length: %d)\n", len(password))
		}
	}

	// Add password authentication
	authMethods = append(authMethods, ssh.Password(password))

	// Add keyboard interactive authentication
	authMethods = append(authMethods, ssh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) (answers []string, err error) {
		answers = make([]string, len(questions))
		for i, question := range questions {
			if strings.Contains(strings.ToLower(question), "password") {
				answers[i] = password
			} else {
				fmt.Printf("%s", question)
				var answer string
				fmt.Scanln(&answer)
				answers[i] = answer
			}
		}
		return answers, nil
	}))

	config := &ssh.ClientConfig{
		User:            opts.User,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // In production, use proper host key checking
		Timeout:         installer.timeout,
	}

	address := net.JoinHostPort(opts.Host, opts.Port)
	if installer.verbose {
		fmt.Printf("Attempting to connect to %s...\n", address)
	}
	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		if installer.verbose {
			// More detailed error information
			fmt.Printf("SSH connection failed: %v\n", err)
			fmt.Printf("Tried authentication methods: ")
			if len(authMethods) > 0 {
				for i := range authMethods {
					if i > 0 {
						fmt.Printf(", ")
					}
					fmt.Printf("method-%d", i+1)
				}
			}
			fmt.Printf("\n")
		}
		return nil, fmt.Errorf("failed to connect to %s@%s:%s - %w", opts.User, opts.Host, opts.Port, err)
	}
	if installer.verbose {
		fmt.Printf("SSH connection established successfully\n")
	}
	return client, nil
}

// createKeyAuth creates SSH key authentication method.
func (installer *SSHKeyInstaller) createKeyAuth(privateKeyPath string) (ssh.AuthMethod, error) {
	key, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, err
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		// Try with passphrase if parsing fails
		fmt.Printf("Enter passphrase for key '%s': ", privateKeyPath)
		passphrase, err := installer.readPassword()
		if err != nil {
			return nil, err
		}
		signer, err = ssh.ParsePrivateKeyWithPassphrase(key, []byte(passphrase))
		if err != nil {
			return nil, err
		}
	}

	return ssh.PublicKeys(signer), nil
}

// readPassword reads password from stdin without echoing.
func (installer *SSHKeyInstaller) readPassword() (string, error) {
	// Check if stdin is a terminal
	fd := syscall.Stdin
	if term.IsTerminal(fd) {
		// Read password without echoing
		password, err := term.ReadPassword(fd)
		if err != nil {
			return "", err
		}
		fmt.Println() // Print newline after hidden password input
		return string(password), nil
	}

	// Fallback for non-terminal input (e.g., pipes, tests)
	reader := bufio.NewReader(os.Stdin)
	password, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(password), nil
}

// keyExists checks if the public key already exists in authorized_keys using SFTP.
func (installer *SSHKeyInstaller) keyExists(client *ssh.Client, publicKey string) (bool, error) {
	if installer.verbose {
		fmt.Printf("Creating SFTP client for key existence check...\n")
	}

	// Create SFTP client
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		if installer.verbose {
			fmt.Printf("Failed to create SFTP client: %v\n", err)
		}
		return false, fmt.Errorf("failed to create SFTP client: %w", err)
	}
	defer sftpClient.Close()

	// Extract the key part (without comment) for comparison
	keyParts := strings.Fields(publicKey)
	if len(keyParts) < 2 {
		return false, fmt.Errorf("invalid public key format")
	}
	keyToCheck := strings.Join(keyParts[:2], " ") // type + key, without comment
	if installer.verbose {
		fmt.Printf("Looking for key type: %s\n", keyParts[0])
	}

	// Read authorized_keys file
	authorizedKeysPath := ".ssh/authorized_keys"
	if installer.verbose {
		fmt.Printf("Reading %s file via SFTP...\n", authorizedKeysPath)
	}

	file, err := sftpClient.Open(authorizedKeysPath)
	if err != nil {
		// File doesn't exist, which is normal
		if os.IsNotExist(err) {
			if installer.verbose {
				fmt.Printf("authorized_keys file doesn't exist (normal for new setup)\n")
			}
			return false, nil
		}
		if installer.verbose {
			fmt.Printf("Error opening authorized_keys: %v\n", err)
		}
		return false, fmt.Errorf("failed to open authorized_keys: %w", err)
	}
	defer file.Close()

	// Read file content
	content, err := io.ReadAll(file)
	if err != nil {
		if installer.verbose {
			fmt.Printf("Error reading authorized_keys: %v\n", err)
		}
		return false, fmt.Errorf("failed to read authorized_keys: %w", err)
	}

	// Check if key exists in file content
	fileContent := string(content)
	if installer.verbose {
		fmt.Printf("Read %d bytes from authorized_keys\n", len(content))
	}

	// Check each line for the key
	lines := strings.Split(fileContent, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Extract type+key from each line (ignore comments)
		lineParts := strings.Fields(line)
		if len(lineParts) >= 2 {
			lineKey := strings.Join(lineParts[:2], " ")
			if lineKey == keyToCheck {
				if installer.verbose {
					fmt.Printf("Key found in authorized_keys\n")
				}
				return true, nil
			}
		}
	}

	if installer.verbose {
		fmt.Printf("Key not found in authorized_keys\n")
	}
	return false, nil
}

// installKey installs the public key to authorized_keys using SFTP.
func (installer *SSHKeyInstaller) installKey(client *ssh.Client, publicKey string) error {
	if installer.verbose {
		fmt.Printf("Creating SFTP client for key installation...\n")
	}

	// Create SFTP client
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		if installer.verbose {
			fmt.Printf("Failed to create SFTP client: %v\n", err)
		}
		return fmt.Errorf("failed to create SFTP client: %w", err)
	}
	defer sftpClient.Close()

	// Create .ssh directory if it doesn't exist
	if installer.verbose {
		fmt.Printf("Ensuring .ssh directory exists...\n")
	}
	sshDir := ".ssh"
	if err := sftpClient.MkdirAll(sshDir); err != nil {
		if installer.verbose {
			fmt.Printf("Failed to create .ssh directory: %v\n", err)
		}
		return fmt.Errorf("failed to create .ssh directory: %w", err)
	}

	// Set .ssh directory permissions (0700)
	if installer.verbose {
		fmt.Printf("Setting .ssh directory permissions to 0700...\n")
	}
	if err := sftpClient.Chmod(sshDir, 0o700); err != nil {
		if installer.verbose {
			fmt.Printf("Failed to set .ssh permissions: %v\n", err)
		}
		return fmt.Errorf("failed to set .ssh permissions: %w", err)
	}

	authorizedKeysPath := ".ssh/authorized_keys"
	var existingContent string

	// Read existing authorized_keys content
	if installer.verbose {
		fmt.Printf("Reading existing authorized_keys file...\n")
	}
	file, err := sftpClient.Open(authorizedKeysPath)
	if err != nil {
		if os.IsNotExist(err) {
			if installer.verbose {
				fmt.Printf("authorized_keys file doesn't exist, will create new one\n")
			}
			existingContent = ""
		} else {
			if installer.verbose {
				fmt.Printf("Error opening authorized_keys: %v\n", err)
			}
			return fmt.Errorf("failed to open authorized_keys: %w", err)
		}
	} else {
		defer file.Close()
		content, err := io.ReadAll(file)
		if err != nil {
			if installer.verbose {
				fmt.Printf("Error reading authorized_keys: %v\n", err)
			}
			return fmt.Errorf("failed to read authorized_keys: %w", err)
		}
		existingContent = string(content)
		file.Close()
	}

	// Parse existing keys and add new key
	if installer.verbose {
		fmt.Printf("Processing keys...\n")
	}
	keyLines := make(map[string]bool)

	// Add existing keys to map (for deduplication)
	if existingContent != "" {
		lines := strings.Split(existingContent, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "#") {
				keyLines[line] = true
			}
		}
	}

	// Add new key
	newKey := strings.TrimSpace(publicKey)
	if newKey != "" {
		keyLines[newKey] = true
		if installer.verbose {
			fmt.Printf("Added new key to collection\n")
		}
	}

	// Build final content
	var finalLines []string
	for line := range keyLines {
		finalLines = append(finalLines, line)
	}
	finalContent := strings.Join(finalLines, "\n")
	if len(finalLines) > 0 {
		finalContent += "\n" // Ensure file ends with newline
	}

	// Write the updated authorized_keys file
	if installer.verbose {
		fmt.Printf("Writing updated authorized_keys file...\n")
	}
	outputFile, err := sftpClient.Create(authorizedKeysPath)
	if err != nil {
		if installer.verbose {
			fmt.Printf("Failed to create authorized_keys file: %v\n", err)
		}
		return fmt.Errorf("failed to create authorized_keys file: %w", err)
	}
	defer outputFile.Close()

	if _, err := outputFile.Write([]byte(finalContent)); err != nil {
		if installer.verbose {
			fmt.Printf("Failed to write authorized_keys content: %v\n", err)
		}
		return fmt.Errorf("failed to write authorized_keys content: %w", err)
	}

	// Set authorized_keys file permissions (0600)
	if installer.verbose {
		fmt.Printf("Setting authorized_keys permissions to 0600...\n")
	}
	if err := sftpClient.Chmod(authorizedKeysPath, 0o600); err != nil {
		if installer.verbose {
			fmt.Printf("Failed to set authorized_keys permissions: %v\n", err)
		}
		return fmt.Errorf("failed to set authorized_keys permissions: %w", err)
	}

	fmt.Printf("✅ SSH key installed successfully\n")
	if installer.verbose {
		fmt.Printf("Total keys in authorized_keys: %d\n", len(finalLines))
	}
	return nil
}
