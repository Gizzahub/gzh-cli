// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package devenv

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// SimpleSSHInstaller uses system SSH command for installation
type SimpleSSHInstaller struct{}

// NewSimpleSSHInstaller creates a new simple SSH installer
func NewSimpleSSHInstaller() *SimpleSSHInstaller {
	return &SimpleSSHInstaller{}
}

// InstallPublicKeySimple installs a public key using system SSH
func (installer *SimpleSSHInstaller) InstallPublicKeySimple(host, user, publicKeyPath string) error {
	// Read public key
	keyContent, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read public key: %w", err)
	}

	key := strings.TrimSpace(string(keyContent))
	if key == "" {
		return fmt.Errorf("public key file is empty")
	}

	// Basic validation
	if !strings.HasPrefix(key, "ssh-") {
		return fmt.Errorf("invalid public key format")
	}

	fmt.Printf("Installing public key to %s@%s...\n", user, host)

	// Use ssh to install the key
	commands := []string{
		"mkdir -p ~/.ssh",
		"chmod 700 ~/.ssh",
		fmt.Sprintf("echo '%s' >> ~/.ssh/authorized_keys", key),
		"chmod 600 ~/.ssh/authorized_keys",
		// Remove duplicates
		"sort ~/.ssh/authorized_keys | uniq > ~/.ssh/authorized_keys.tmp && mv ~/.ssh/authorized_keys.tmp ~/.ssh/authorized_keys",
		"echo 'Public key installed successfully'",
	}

	cmdStr := strings.Join(commands, " && ")
	cmd := exec.Command("ssh", "-o", "StrictHostKeyChecking=no", fmt.Sprintf("%s@%s", user, host), cmdStr)

	// Connect stdin/stdout/stderr to allow password input
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
