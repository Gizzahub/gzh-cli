// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package synclone

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Gizzahub/gzh-cli/internal/app"
)

func TestConfigCommand(t *testing.T) {
	cmd := newSyncCloneConfigCmd(app.NewTestAppContext())

	// Test that config command exists
	assert.NotNil(t, cmd)
	assert.Equal(t, "config", cmd.Use)

	// Test that subcommands exist
	subcommands := []string{"generate", "validate", "convert"}
	for _, sub := range subcommands {
		found := false
		for _, child := range cmd.Commands() {
			if child.Use == sub || child.Name() == sub {
				found = true
				break
			}
		}
		assert.True(t, found, "Subcommand %s not found", sub)
	}
}

func TestConfigValidate(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "test-config.yaml")

	validConfig := `version: "2.0"
synclone:
  providers:
    - name: test-github
      type: github
      organization: test-org
      target_dir: ./repos
`

	err := os.WriteFile(configFile, []byte(validConfig), 0o644)
	require.NoError(t, err)

	// Test validation
	cmd := newConfigValidateCmd()
	cmd.SetArgs([]string{"--file", configFile})

	// Capture output
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err = cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "valid")
}

func TestConfigConvert(t *testing.T) {
	// Create a temporary v1 config file
	tmpDir := t.TempDir()
	v1File := filepath.Join(tmpDir, "v1-config.yaml")
	v2File := filepath.Join(tmpDir, "v2-config.yaml")

	v1Config := `bulk_clone:
  repository_roots:
    - name: test-org
      provider: github
      organization: test-org
      target_dir: ./repos
      token: ${GITHUB_TOKEN}
`

	err := os.WriteFile(v1File, []byte(v1Config), 0o644)
	require.NoError(t, err)

	// Test conversion
	cmd := newConfigConvertCmd()
	cmd.SetArgs([]string{
		"--file", v1File,
		"--output", v2File,
		"--from", "v1",
		"--to", "v2",
		"--backup=false",
	})

	err = cmd.Execute()
	assert.NoError(t, err)

	// Check that v2 file was created
	_, err = os.Stat(v2File)
	assert.NoError(t, err)

	// Read and verify v2 content
	v2Content, err := os.ReadFile(v2File)
	require.NoError(t, err)
	assert.Contains(t, string(v2Content), "version:")
	assert.Contains(t, string(v2Content), "synclone:")
}
