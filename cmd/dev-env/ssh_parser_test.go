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

func TestSSHConfigParser_Parse(t *testing.T) {
	tests := []struct {
		name             string
		config           string
		includeFiles     map[string]string
		keyFiles         map[string]string
		expectedIncludes int
		expectedKeys     int
		expectedPubKeys  int
	}{
		{
			name: "simple config without includes",
			config: `Host example.com
    HostName example.com
    User myuser
    IdentityFile id_rsa`,
			keyFiles: map[string]string{
				"id_rsa":     "-----BEGIN PRIVATE KEY-----\ntest\n-----END PRIVATE KEY-----",
				"id_rsa.pub": "ssh-rsa AAAAB3NzaC1yc2E test@example.com",
			},
			expectedIncludes: 0,
			expectedKeys:     1,
			expectedPubKeys:  1,
		},
		{
			name: "config with includes",
			config: `Include config.d/*
Include personal.conf

Host example.com
    HostName example.com
    User myuser`,
			includeFiles: map[string]string{
				"config.d/work.conf": `Host work.com
    HostName work.internal.com
    User workuser`,
				"config.d/personal.conf": `Host personal.com
    HostName personal.com
    User me`,
				"personal.conf": `Host direct.com
    HostName direct.com
    User direct`,
			},
			expectedIncludes: 3,
			expectedKeys:     0,
			expectedPubKeys:  0,
		},
		{
			name: "config with includes and keys",
			config: `Include config.d/*

Host server1
    HostName server1.com
    User admin
    IdentityFile server1_key

Host server2
    HostName server2.com
    User root
    IdentityFile server2_key`,
			includeFiles: map[string]string{
				"config.d/extra.conf": `Host extra.com
    HostName extra.com
    User extra
    IdentityFile extra_key`,
			},
			keyFiles: map[string]string{
				"server1_key":     "private key 1",
				"server1_key.pub": "public key 1",
				"server2_key":     "private key 2",
				"extra_key":       "extra private key",
				"extra_key.pub":   "extra public key",
			},
			expectedIncludes: 1,
			expectedKeys:     3, // server1_key, server2_key, extra_key
			expectedPubKeys:  2, // server1_key.pub, extra_key.pub (server2_key.pub doesn't exist)
		},
		{
			name: "case insensitive directives",
			config: `INCLUDE config.d/*
include personal.conf

Host example
    Hostname example.com
    IDENTITYFILE test_key`,
			includeFiles: map[string]string{
				"config.d/test.conf": "Host test\n    HostName test.com",
				"personal.conf":      "Host personal\n    HostName personal.com",
			},
			keyFiles: map[string]string{
				"test_key": "test private key",
			},
			expectedIncludes: 2,
			expectedKeys:     1,
			expectedPubKeys:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory structure
			tempDir := t.TempDir()
			sshDir := filepath.Join(tempDir, ".ssh")
			require.NoError(t, os.MkdirAll(sshDir, 0755))

			// Create main config file
			configPath := filepath.Join(sshDir, "config")
			require.NoError(t, os.WriteFile(configPath, []byte(tt.config), 0644))

			// Create include files
			for relPath, content := range tt.includeFiles {
				fullPath := filepath.Join(sshDir, relPath)
				require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0755))
				require.NoError(t, os.WriteFile(fullPath, []byte(content), 0644))
			}

			// Create key files
			for keyName, content := range tt.keyFiles {
				keyPath := filepath.Join(sshDir, keyName)
				require.NoError(t, os.WriteFile(keyPath, []byte(content), 0600))
			}

			// Parse SSH config
			parser := NewSSHConfigParser(configPath)
			result, err := parser.Parse()

			// Assertions
			require.NoError(t, err)
			assert.Equal(t, configPath, result.MainConfigPath)
			assert.Len(t, result.IncludeFiles, tt.expectedIncludes)
			assert.Len(t, result.PrivateKeys, tt.expectedKeys)
			assert.Len(t, result.PublicKeys, tt.expectedPubKeys)

			// Verify all returned paths exist
			for _, includePath := range result.IncludeFiles {
				assert.FileExists(t, includePath)
			}
			for _, keyPath := range result.PrivateKeys {
				assert.FileExists(t, keyPath)
			}
			for _, pubKeyPath := range result.PublicKeys {
				assert.FileExists(t, pubKeyPath)
			}
		})
	}
}

func TestSSHConfigParser_ParseIncludeLine(t *testing.T) {
	tempDir := t.TempDir()
	sshDir := filepath.Join(tempDir, ".ssh")
	require.NoError(t, os.MkdirAll(sshDir, 0755))

	// Create test files
	configDir := filepath.Join(sshDir, "config.d")
	require.NoError(t, os.MkdirAll(configDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "file1.conf"), []byte("test"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "file2.conf"), []byte("test"), 0644))

	parser := NewSSHConfigParser(filepath.Join(sshDir, "config"))
	result := &ParsedSSHConfig{}

	tests := []struct {
		name           string
		line           string
		expectedCount  int
	}{
		{
			name:          "glob pattern",
			line:          "Include config.d/*",
			expectedCount: 2,
		},
		{
			name:          "case insensitive",
			line:          "INCLUDE config.d/*",
			expectedCount: 2,
		},
		{
			name:          "with extra spaces",
			line:          "  Include   config.d/*  ",
			expectedCount: 2,
		},
		{
			name:          "not an include line",
			line:          "Host example.com",
			expectedCount: 0,
		},
		{
			name:          "commented include",
			line:          "# Include config.d/*",
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result.IncludeFiles = []string{} // Reset
			err := parser.parseIncludeLine(tt.line, result)
			assert.NoError(t, err)
			assert.Len(t, result.IncludeFiles, tt.expectedCount)
		})
	}
}

func TestSSHConfigParser_ParseIdentityFileLine(t *testing.T) {
	tempDir := t.TempDir()
	sshDir := filepath.Join(tempDir, ".ssh")
	require.NoError(t, os.MkdirAll(sshDir, 0755))

	// Create test key files
	require.NoError(t, os.WriteFile(filepath.Join(sshDir, "test_key"), []byte("private"), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(sshDir, "test_key.pub"), []byte("public"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(sshDir, "no_pub_key"), []byte("private"), 0600))

	parser := NewSSHConfigParser(filepath.Join(sshDir, "config"))
	
	tests := []struct {
		name               string
		line               string
		expectedPrivateKeys int
		expectedPublicKeys  int
	}{
		{
			name:               "identity file with pub key relative path",
			line:               "IdentityFile test_key",
			expectedPrivateKeys: 1,
			expectedPublicKeys:  1,
		},
		{
			name:               "identity file without pub key relative path",
			line:               "IdentityFile no_pub_key",
			expectedPrivateKeys: 1,
			expectedPublicKeys:  0,
		},
		{
			name:               "case insensitive",
			line:               "IDENTITYFILE test_key",
			expectedPrivateKeys: 1,
			expectedPublicKeys:  1,
		},
		{
			name:               "with extra spaces",
			line:               "  IdentityFile   test_key  ",
			expectedPrivateKeys: 1,
			expectedPublicKeys:  1,
		},
		{
			name:               "not an identity file line",
			line:               "HostName example.com",
			expectedPrivateKeys: 0,
			expectedPublicKeys:  0,
		},
		{
			name:               "non-existent key",
			line:               "IdentityFile nonexistent",
			expectedPrivateKeys: 0,
			expectedPublicKeys:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &ParsedSSHConfig{}
			err := parser.parseIdentityFileLine(tt.line, result)
			assert.NoError(t, err)
			assert.Len(t, result.PrivateKeys, tt.expectedPrivateKeys)
			assert.Len(t, result.PublicKeys, tt.expectedPublicKeys)
		})
	}
}

func TestRemoveDuplicates(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "no duplicates",
			input:    []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "with duplicates",
			input:    []string{"a", "b", "a", "c", "b"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "empty slice",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "all duplicates",
			input:    []string{"a", "a", "a"},
			expected: []string{"a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeDuplicates(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}