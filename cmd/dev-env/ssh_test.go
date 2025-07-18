package devenv

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultSshOptions(t *testing.T) {
	opts := defaultSshOptions()

	assert.NotEmpty(t, opts.configPath)
	assert.NotEmpty(t, opts.storePath)
	assert.False(t, opts.force)
	assert.False(t, opts.listAll)
}

func TestNewSshCmd(t *testing.T) {
	cmd := newSshCmd()

	assert.Equal(t, "ssh", cmd.Use)
	assert.Equal(t, "Manage SSH configuration files", cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Check subcommands
	subcommands := cmd.Commands()
	assert.Len(t, subcommands, 3)

	var saveCmd, loadCmd, listCmd bool

	for _, subcmd := range subcommands {
		switch subcmd.Use {
		case "save":
			saveCmd = true
		case "load":
			loadCmd = true
		case "list":
			listCmd = true
		}
	}

	assert.True(t, saveCmd, "save subcommand should exist")
	assert.True(t, loadCmd, "load subcommand should exist")
	assert.True(t, listCmd, "list subcommand should exist")
}

func TestSshSaveCmd(t *testing.T) {
	cmd := newSshSaveCmd()

	assert.Equal(t, "save", cmd.Use)
	assert.Equal(t, "Save current SSH config", cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Test flags
	assert.NotNil(t, cmd.Flags().Lookup("name"))
	assert.NotNil(t, cmd.Flags().Lookup("description"))
	assert.NotNil(t, cmd.Flags().Lookup("config-path"))
	assert.NotNil(t, cmd.Flags().Lookup("store-path"))
	assert.NotNil(t, cmd.Flags().Lookup("force"))
}

func TestSshLoadCmd(t *testing.T) {
	cmd := newSshLoadCmd()

	assert.Equal(t, "load", cmd.Use)
	assert.Equal(t, "Load a saved SSH config", cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Test flags
	assert.NotNil(t, cmd.Flags().Lookup("name"))
	assert.NotNil(t, cmd.Flags().Lookup("config-path"))
	assert.NotNil(t, cmd.Flags().Lookup("store-path"))
	assert.NotNil(t, cmd.Flags().Lookup("force"))
}

func TestSshListCmd(t *testing.T) {
	cmd := newSshListCmd()

	assert.Equal(t, "list", cmd.Use)
	assert.Equal(t, "List saved SSH configs", cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Test flags
	assert.NotNil(t, cmd.Flags().Lookup("store-path"))
}

func TestSshSaveLoad(t *testing.T) {
	// Create temporary directories for testing
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config")
	storeDir := filepath.Join(tempDir, "store")

	err := os.MkdirAll(configDir, 0o755)
	require.NoError(t, err)

	// Create a test SSH config file
	testSshConfigContent := `# SSH Config for testing

Host production-server
    HostName prod.example.com
    User deploy
    Port 2222
    IdentityFile ~/.ssh/id_rsa_prod

Host staging-server
    HostName staging.example.com
    User ubuntu
    IdentityFile ~/.ssh/id_rsa_staging

Host dev-*
    User developer
    Port 22
    IdentityFile ~/.ssh/id_rsa_dev

Host github.com
    HostName github.com
    User git
    IdentityFile ~/.ssh/id_rsa

Host *
    ServerAliveInterval 60
    ServerAliveCountMax 3
    TCPKeepAlive yes
`

	configPath := filepath.Join(configDir, "config")
	err = os.WriteFile(configPath, []byte(testSshConfigContent), 0o644)
	require.NoError(t, err)

	t.Run("save SSH config", func(t *testing.T) {
		opts := &sshOptions{
			name:        "test-config",
			description: "Test SSH configuration",
			configPath:  configPath,
			storePath:   storeDir,
			force:       false,
		}

		err := opts.runSave(nil, nil)
		assert.NoError(t, err)

		// Check if file was saved
		savedPath := filepath.Join(storeDir, "test-config.config")
		assert.FileExists(t, savedPath)

		// Check if metadata was saved
		metadataPath := filepath.Join(storeDir, "test-config.meta")
		assert.FileExists(t, metadataPath)

		// Verify saved content
		savedContent, err := os.ReadFile(savedPath)
		require.NoError(t, err)
		assert.Equal(t, testSshConfigContent, string(savedContent))
	})

	t.Run("load SSH config", func(t *testing.T) {
		// Create a different target path
		targetPath := filepath.Join(tempDir, "loaded", "config")

		opts := &sshOptions{
			name:       "test-config",
			configPath: targetPath,
			storePath:  storeDir,
			force:      true, // Skip backup for test
		}

		err := opts.runLoad(nil, nil)
		assert.NoError(t, err)

		// Check if file was loaded
		assert.FileExists(t, targetPath)

		// Check content matches
		loadedContent, err := os.ReadFile(targetPath)
		require.NoError(t, err)
		assert.Equal(t, testSshConfigContent, string(loadedContent))
	})

	t.Run("list SSH configs", func(t *testing.T) {
		opts := &sshOptions{
			storePath: storeDir,
		}

		err := opts.runList(nil, nil)
		assert.NoError(t, err)
	})
}

func TestSshMetadata(t *testing.T) {
	tempDir := t.TempDir()

	opts := &sshOptions{
		name:        "test-metadata",
		description: "Test description",
		storePath:   tempDir,
	}

	// Test save metadata
	err := opts.saveMetadata()
	assert.NoError(t, err)

	metadataPath := filepath.Join(tempDir, "test-metadata.meta")
	assert.FileExists(t, metadataPath)

	// Test load metadata
	metadata := opts.loadMetadata("test-metadata")
	assert.Equal(t, "test-metadata", metadata.Name)
	assert.Equal(t, "Test description", metadata.Description)
	assert.False(t, metadata.SavedAt.IsZero())
}

func TestSshCopyFile(t *testing.T) {
	tempDir := t.TempDir()

	// Create source file
	srcPath := filepath.Join(tempDir, "source.config")
	content := `Host test
    HostName test.example.com
    User testuser`
	err := os.WriteFile(srcPath, []byte(content), 0o644)
	require.NoError(t, err)

	// Test copy
	opts := &sshOptions{}
	dstPath := filepath.Join(tempDir, "destination.config")

	err = opts.copyFile(srcPath, dstPath)
	assert.NoError(t, err)

	// Check content
	copiedContent, err := os.ReadFile(dstPath)
	require.NoError(t, err)
	assert.Equal(t, content, string(copiedContent))

	// Check permissions
	srcInfo, err := os.Stat(srcPath)
	require.NoError(t, err)
	dstInfo, err := os.Stat(dstPath)
	require.NoError(t, err)
	assert.Equal(t, srcInfo.Mode(), dstInfo.Mode())
}

func TestSshDisplayConfigInfo(t *testing.T) {
	tempDir := t.TempDir()

	// Create test SSH config
	testSshConfigContent := `Host web-server
    HostName web.example.com
    User admin
    Port 2222
    IdentityFile ~/.ssh/id_rsa_web

Host database-server
    HostName db.example.com
    User postgres
    IdentityFile ~/.ssh/id_rsa_db

Host github.com
    HostName github.com
    User git
    IdentityFile ~/.ssh/id_github

Host dev-*
    User developer
    Port 22

Host *
    ServerAliveInterval 60
`

	configPath := filepath.Join(tempDir, "config")
	err := os.WriteFile(configPath, []byte(testSshConfigContent), 0o644)
	require.NoError(t, err)

	opts := &sshOptions{}
	err = opts.displayConfigInfo(configPath)
	assert.NoError(t, err)
}

func TestSshParseSshConfig(t *testing.T) {
	opts := &sshOptions{}

	testConfig := `# Test SSH Config

Host production-web
    HostName web.prod.example.com
    User deploy
    Port 2222
    IdentityFile ~/.ssh/id_rsa_prod

Host staging-web
    HostName web.staging.example.com
    User ubuntu
    IdentityFile ~/.ssh/id_rsa_staging

Host database
    HostName db.example.com
    User postgres
    Port 5432
    IdentityFile ~/.ssh/id_rsa_db

Host github.com
    HostName github.com
    User git
    IdentityFile ~/.ssh/id_github

# Wildcard patterns (should be filtered out in display)
Host dev-*
    User developer
    Port 22

Host *.internal
    User admin
    IdentityFile ~/.ssh/id_internal

Host *
    ServerAliveInterval 60
    ServerAliveCountMax 3
`

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config")
	err := os.WriteFile(configPath, []byte(testConfig), 0o644)
	require.NoError(t, err)

	hosts := opts.parseSshConfig(configPath)

	// Should have 4 hosts (excluding wildcard patterns)
	assert.Len(t, hosts, 4)

	// Check production-web host
	var prodHost *sshHost

	for i := range hosts {
		if hosts[i].Name == "production-web" {
			prodHost = &hosts[i]
			break
		}
	}

	require.NotNil(t, prodHost)
	assert.Equal(t, "web.prod.example.com", prodHost.Hostname)
	assert.Equal(t, "deploy", prodHost.User)
	assert.Equal(t, "2222", prodHost.Port)
	assert.Equal(t, "~/.ssh/id_rsa_prod", prodHost.KeyFile)

	// Check staging-web host
	var stagingHost *sshHost

	for i := range hosts {
		if hosts[i].Name == "staging-web" {
			stagingHost = &hosts[i]
			break
		}
	}

	require.NotNil(t, stagingHost)
	assert.Equal(t, "web.staging.example.com", stagingHost.Hostname)
	assert.Equal(t, "ubuntu", stagingHost.User)
	assert.Equal(t, "~/.ssh/id_rsa_staging", stagingHost.KeyFile)

	// Check database host
	var dbHost *sshHost

	for i := range hosts {
		if hosts[i].Name == "database" {
			dbHost = &hosts[i]
			break
		}
	}

	require.NotNil(t, dbHost)
	assert.Equal(t, "db.example.com", dbHost.Hostname)
	assert.Equal(t, "postgres", dbHost.User)
	assert.Equal(t, "5432", dbHost.Port)

	// Check github.com host
	var githubHost *sshHost

	for i := range hosts {
		if hosts[i].Name == "github.com" {
			githubHost = &hosts[i]
			break
		}
	}

	require.NotNil(t, githubHost)
	assert.Equal(t, "github.com", githubHost.Hostname)
	assert.Equal(t, "git", githubHost.User)
	assert.Equal(t, "~/.ssh/id_github", githubHost.KeyFile)

	// Verify wildcard patterns are filtered out
	for _, host := range hosts {
		assert.NotEqual(t, "dev-*", host.Name)
		assert.NotEqual(t, "*.internal", host.Name)
		assert.NotEqual(t, "*", host.Name)
	}
}

func TestSshErrorCases(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("save non-existent config", func(t *testing.T) {
		opts := &sshOptions{
			name:       "test",
			configPath: "/non/existent/path",
			storePath:  tempDir,
		}

		err := opts.runSave(nil, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "SSH config file not found")
	})

	t.Run("load non-existent config", func(t *testing.T) {
		opts := &sshOptions{
			name:      "non-existent",
			storePath: tempDir,
		}

		err := opts.runLoad(nil, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("save duplicate without force", func(t *testing.T) {
		// Create a test config
		configPath := filepath.Join(tempDir, "config")
		err := os.WriteFile(configPath, []byte("Host test\n  HostName test.com"), 0o644)
		require.NoError(t, err)

		opts := &sshOptions{
			name:       "duplicate-test",
			configPath: configPath,
			storePath:  tempDir,
			force:      false,
		}

		// Save first time
		err = opts.runSave(nil, nil)
		assert.NoError(t, err)

		// Try to save again without force
		err = opts.runSave(nil, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("parse invalid SSH config", func(t *testing.T) {
		tempDir := t.TempDir()
		invalidConfigPath := filepath.Join(tempDir, "invalid_config")
		err := os.WriteFile(invalidConfigPath, []byte("invalid ssh config content"), 0o644)
		require.NoError(t, err)

		opts := &sshOptions{}
		hosts := opts.parseSshConfig(invalidConfigPath)

		// Should return empty list for invalid config
		assert.Empty(t, hosts)
	})

	t.Run("parse non-existent SSH config", func(t *testing.T) {
		opts := &sshOptions{}
		hosts := opts.parseSshConfig("/non/existent/config")

		// Should return nil for non-existent file
		assert.Nil(t, hosts)
	})
}

func TestSshConfigWithQuotes(t *testing.T) {
	opts := &sshOptions{}

	testConfig := `Host quoted-host
    HostName "hostname.example.com"
    User "myuser"
    IdentityFile "~/.ssh/my key file with spaces"
    Port "2222"

Host single-quoted
    HostName 'single.example.com'
    IdentityFile '~/.ssh/single key'
`

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "quoted_config")
	err := os.WriteFile(configPath, []byte(testConfig), 0o644)
	require.NoError(t, err)

	hosts := opts.parseSshConfig(configPath)

	assert.Len(t, hosts, 2)

	// Check quoted host (should handle quotes properly)
	var quotedHost *sshHost

	for i := range hosts {
		if hosts[i].Name == "quoted-host" {
			quotedHost = &hosts[i]
			break
		}
	}

	require.NotNil(t, quotedHost)
	assert.Equal(t, "\"hostname.example.com\"", quotedHost.Hostname)
	assert.Equal(t, "\"myuser\"", quotedHost.User)
	assert.Equal(t, "\"2222\"", quotedHost.Port)
	assert.Equal(t, "~/.ssh/my key file with spaces", quotedHost.KeyFile) // Quotes should be stripped

	// Check single quoted host
	var singleQuotedHost *sshHost

	for i := range hosts {
		if hosts[i].Name == "single-quoted" {
			singleQuotedHost = &hosts[i]
			break
		}
	}

	require.NotNil(t, singleQuotedHost)
	assert.Equal(t, "'single.example.com'", singleQuotedHost.Hostname)
	assert.Equal(t, "~/.ssh/single key", singleQuotedHost.KeyFile) // Quotes should be stripped
}
