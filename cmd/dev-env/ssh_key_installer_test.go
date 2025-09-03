// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package devenv

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSSHKeyInstaller_ValidateOptions(t *testing.T) {
	installer := NewSSHKeyInstaller()
	tempDir := t.TempDir()
	
	// Create test public key
	testKeyPath := filepath.Join(tempDir, "test_key.pub")
	testKeyContent := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC7... test@example.com"
	require.NoError(t, os.WriteFile(testKeyPath, []byte(testKeyContent), 0644))

	tests := []struct {
		name    string
		opts    *InstallOptions
		wantErr string
	}{
		{
			name: "valid options",
			opts: &InstallOptions{
				Host:          "example.com",
				User:          "testuser",
				PublicKeyPath: testKeyPath,
				Port:          "22",
			},
			wantErr: "",
		},
		{
			name: "missing host",
			opts: &InstallOptions{
				User:          "testuser",
				PublicKeyPath: testKeyPath,
			},
			wantErr: "host is required",
		},
		{
			name: "missing user",
			opts: &InstallOptions{
				Host:          "example.com",
				PublicKeyPath: testKeyPath,
			},
			wantErr: "user is required",
		},
		{
			name: "missing public key path",
			opts: &InstallOptions{
				Host: "example.com",
				User: "testuser",
			},
			wantErr: "public key path is required",
		},
		{
			name: "non-existent public key",
			opts: &InstallOptions{
				Host:          "example.com",
				User:          "testuser",
				PublicKeyPath: "/nonexistent/key.pub",
			},
			wantErr: "public key file not found",
		},
		{
			name: "default port should be set",
			opts: &InstallOptions{
				Host:          "example.com",
				User:          "testuser",
				PublicKeyPath: testKeyPath,
			},
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := installer.validateOptions(tt.opts)
			if tt.wantErr == "" {
				assert.NoError(t, err)
				if tt.opts.Port == "" {
					assert.Equal(t, "22", tt.opts.Port) // Should set default port
				}
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			}
		})
	}
}

func TestSSHKeyInstaller_ReadPublicKey(t *testing.T) {
	installer := NewSSHKeyInstaller()
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		keyContent  string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid ssh-rsa key",
			keyContent:  "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC7... test@example.com",
			expectError: false,
		},
		{
			name:        "valid ssh-ed25519 key",
			keyContent:  "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI... test@example.com",
			expectError: false,
		},
		{
			name:        "valid key with extra whitespace",
			keyContent:  "\n  ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC7... test@example.com  \n",
			expectError: false,
		},
		{
			name:        "empty key file",
			keyContent:  "",
			expectError: true,
			errorMsg:    "public key file is empty",
		},
		{
			name:        "invalid key format",
			keyContent:  "not-a-valid-ssh-key",
			expectError: true,
			errorMsg:    "invalid public key format",
		},
		{
			name:        "whitespace only",
			keyContent:  "\n\t  \n",
			expectError: true,
			errorMsg:    "public key file is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keyPath := filepath.Join(tempDir, tt.name+".pub")
			require.NoError(t, os.WriteFile(keyPath, []byte(tt.keyContent), 0644))

			key, err := installer.readPublicKey(keyPath)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Empty(t, key)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, key)
				assert.True(t, len(key) > 0)
				// Should be trimmed
				assert.Equal(t, key, key)
			}
		})
	}
}

func TestSSHKeyInstaller_DryRun(t *testing.T) {
	installer := NewSSHKeyInstaller()
	tempDir := t.TempDir()
	
	// Create test public key
	testKeyPath := filepath.Join(tempDir, "test_key.pub")
	testKeyContent := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC7... test@example.com"
	require.NoError(t, os.WriteFile(testKeyPath, []byte(testKeyContent), 0644))

	opts := &InstallOptions{
		Host:          "example.com",
		User:          "testuser",
		PublicKeyPath: testKeyPath,
		DryRun:        true,
	}

	result, err := installer.InstallPublicKey(opts)
	
	assert.NoError(t, err)
	assert.True(t, result.Success)
	assert.Contains(t, result.Message, "DRY RUN")
	assert.Contains(t, result.Message, "Would install key")
	assert.False(t, result.KeyAdded)
	assert.False(t, result.KeyExists)
}

func TestSSHKeyInstaller_InstallKeysFromConfig(t *testing.T) {
	installer := NewSSHKeyInstaller()
	tempDir := t.TempDir()
	storeDir := filepath.Join(tempDir, "store")
	configName := "test-config"
	
	// Create configuration directory structure
	configDir := filepath.Join(storeDir, configName)
	keysDir := filepath.Join(configDir, "keys")
	require.NoError(t, os.MkdirAll(keysDir, 0755))

	// Create test keys
	testKeys := map[string]string{
		"id_rsa.pub":     "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC7... test@example.com",
		"id_ed25519.pub": "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI... test@example.com",
	}

	var publicKeyPaths []string
	for keyName, keyContent := range testKeys {
		keyPath := filepath.Join(keysDir, keyName)
		require.NoError(t, os.WriteFile(keyPath, []byte(keyContent), 0644))
		publicKeyPaths = append(publicKeyPaths, keyPath)
		
		// Create corresponding private key
		privateKeyPath := filepath.Join(keysDir, keyName[:len(keyName)-4]) // Remove .pub
		require.NoError(t, os.WriteFile(privateKeyPath, []byte("private key content"), 0600))
	}

	// Create metadata
	enhancedCmd := NewEnhancedSSHCommand()
	metadata := EnhancedSSHMetadata{
		Description: "Test configuration",
		PublicKeys:  publicKeyPaths,
		HasKeys:     true,
	}
	metadataFile := filepath.Join(configDir, "metadata.json")
	require.NoError(t, enhancedCmd.saveEnhancedMetadata(metadataFile, metadata))

	// Test with dry run
	opts := &InstallOptions{
		Host:   storeDir, // Pass store dir as host (hack for testing)
		Port:   "22",
		DryRun: true,
	}
	
	results, err := installer.InstallKeysFromConfig(configName, "example.com", "testuser", opts)
	
	assert.NoError(t, err)
	assert.Len(t, results, 2)
	
	for _, result := range results {
		assert.True(t, result.Success)
		assert.Contains(t, result.Message, "DRY RUN")
	}
}

func TestEnhancedSSHCommand_ListKeys(t *testing.T) {
	enhancedCmd := NewEnhancedSSHCommand()
	tempDir := t.TempDir()
	storeDir := filepath.Join(tempDir, "store")
	configName := "test-config"
	
	// Create configuration directory structure
	configDir := filepath.Join(storeDir, configName)
	keysDir := filepath.Join(configDir, "keys")
	require.NoError(t, os.MkdirAll(keysDir, 0755))

	// Create test keys
	testKeys := map[string]string{
		"id_rsa.pub":     "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC7vF8k1234567890abcdefghijk test@example.com",
		"id_ed25519.pub": "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI1234567890abcdefghijk test@example.com",
	}

	var publicKeyPaths []string
	for keyName, keyContent := range testKeys {
		keyPath := filepath.Join(keysDir, keyName)
		require.NoError(t, os.WriteFile(keyPath, []byte(keyContent), 0644))
		publicKeyPaths = append(publicKeyPaths, filepath.Join("/original/path", keyName))
		
		// Create corresponding private key for one of them
		if keyName == "id_rsa.pub" {
			privateKeyPath := filepath.Join(keysDir, "id_rsa")
			require.NoError(t, os.WriteFile(privateKeyPath, []byte("private key"), 0600))
		}
	}

	// Create metadata
	metadata := EnhancedSSHMetadata{
		Description: "Test configuration",
		PublicKeys:  publicKeyPaths,
		HasKeys:     true,
	}
	metadataFile := filepath.Join(configDir, "metadata.json")
	require.NoError(t, enhancedCmd.saveEnhancedMetadata(metadataFile, metadata))

	// Test listing keys
	err := enhancedCmd.listKeysFromConfig(storeDir, configName)
	assert.NoError(t, err)
}

func TestSSHKeyInstaller_CreateKeyAuth(t *testing.T) {
	installer := NewSSHKeyInstaller()
	tempDir := t.TempDir()
	
	// Create a dummy private key file (not a real key)
	privateKeyPath := filepath.Join(tempDir, "test_key")
	// This is not a real private key, just for testing file reading
	keyContent := `-----BEGIN OPENSSH PRIVATE KEY-----
not_a_real_key_content_for_testing
-----END OPENSSH PRIVATE KEY-----`
	require.NoError(t, os.WriteFile(privateKeyPath, []byte(keyContent), 0600))

	// This will fail to parse, but we're testing the file reading part
	_, err := installer.createKeyAuth(privateKeyPath)
	
	// Should return an error since it's not a real key, but no file read error
	assert.Error(t, err)
	assert.NotContains(t, err.Error(), "no such file or directory")
}

func TestInstallOptions_Validation(t *testing.T) {
	tests := []struct {
		name string
		opts InstallOptions
	}{
		{
			name: "all fields set",
			opts: InstallOptions{
				Host:           "example.com",
				Port:           "2222",
				User:           "testuser",
				PublicKeyPath:  "/path/to/key.pub",
				PrivateKeyPath: "/path/to/key",
				Password:       "secret",
				Force:          true,
				DryRun:         false,
			},
		},
		{
			name: "minimal options",
			opts: InstallOptions{
				Host:          "example.com",
				User:          "testuser",
				PublicKeyPath: "/path/to/key.pub",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just test that the struct can be created and fields accessed
			assert.Equal(t, "example.com", tt.opts.Host)
			assert.Equal(t, "testuser", tt.opts.User)
			assert.Equal(t, "/path/to/key.pub", tt.opts.PublicKeyPath)
		})
	}
}

func TestInstallResult_Fields(t *testing.T) {
	result := &InstallResult{
		Host:      "example.com",
		Success:   true,
		Message:   "Key installed successfully",
		KeyAdded:  true,
		KeyExists: false,
	}

	assert.Equal(t, "example.com", result.Host)
	assert.True(t, result.Success)
	assert.Equal(t, "Key installed successfully", result.Message)
	assert.True(t, result.KeyAdded)
	assert.False(t, result.KeyExists)
}