//nolint:testpackage // White-box testing needed for internal function access
package devenv

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testCredentialsContent = `[default]
aws_access_key_id = AKIAIOSFODNN7EXAMPLE
aws_secret_access_key = wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY

[dev]
aws_access_key_id = AKIAI44QH8DHBEXAMPLE
aws_secret_access_key = je7MtGbClwBF/2Zp9Utk/h3yCo8nvbEXAMPLEKEY
aws_session_token = FwoGZXIvYXdzEDwaDFwXHpI4s4m+k1ppyiLABTJuCM7p5xOuJkzlXy...
mfa_serial = arn:aws:iam::123456789012:mfa/dev-user

[production]
aws_access_key_id = AKIAIOSFODNN7PRODEXAMPLE
aws_secret_access_key = wJalrXUtnFEMI/K7MDENG/bPxRfiCYPRODEXAMPLE
`

func TestDefaultAwsCredentialsOptions(t *testing.T) {
	opts := defaultAwsCredentialsOptions()

	assert.NotEmpty(t, opts.configPath)
	assert.NotEmpty(t, opts.storePath)
	assert.False(t, opts.force)
	assert.False(t, opts.listAll)
}

func TestNewAwsCredentialsCmd(t *testing.T) {
	cmd := newAwsCredentialsCmd()

	assert.Equal(t, "aws-credentials", cmd.Use)
	assert.Equal(t, "Manage AWS credentials files", cmd.Short)
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

func TestAwsCredentialsSaveCmd(t *testing.T) {
	cmd := newAwsCredentialsSaveCmd()

	assert.Equal(t, "save", cmd.Use)
	assert.Equal(t, "Save current AWS credentials", cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Test flags
	assert.NotNil(t, cmd.Flags().Lookup("name"))
	assert.NotNil(t, cmd.Flags().Lookup("description"))
	assert.NotNil(t, cmd.Flags().Lookup("config-path"))
	assert.NotNil(t, cmd.Flags().Lookup("store-path"))
	assert.NotNil(t, cmd.Flags().Lookup("force"))
}

func TestAwsCredentialsLoadCmd(t *testing.T) {
	cmd := newAwsCredentialsLoadCmd()

	assert.Equal(t, "load", cmd.Use)
	assert.Equal(t, "Load a saved AWS credentials file", cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Test flags
	assert.NotNil(t, cmd.Flags().Lookup("name"))
	assert.NotNil(t, cmd.Flags().Lookup("config-path"))
	assert.NotNil(t, cmd.Flags().Lookup("store-path"))
	assert.NotNil(t, cmd.Flags().Lookup("force"))
}

func TestAwsCredentialsListCmd(t *testing.T) {
	cmd := newAwsCredentialsListCmd()

	assert.Equal(t, "list", cmd.Use)
	assert.Equal(t, "List saved AWS credentials", cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Test flags
	assert.NotNil(t, cmd.Flags().Lookup("store-path"))
}

func TestAwsCredentialsSaveLoad(t *testing.T) {
	// Create temporary directories for testing
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config")
	storeDir := filepath.Join(tempDir, "store")

	err := os.MkdirAll(configDir, 0o755)
	require.NoError(t, err)

	// Create a test AWS credentials file

	configPath := filepath.Join(configDir, "credentials")
	err = os.WriteFile(configPath, []byte(testCredentialsContent), 0o600)
	require.NoError(t, err)

	t.Run("save AWS credentials", func(t *testing.T) {
		opts := &awsCredentialsOptions{
			name:        "test-credentials",
			description: "Test AWS credentials",
			configPath:  configPath,
			storePath:   storeDir,
			force:       false,
		}

		err := opts.runSave(nil, nil)
		assert.NoError(t, err)

		// Check if file was saved
		savedPath := filepath.Join(storeDir, "test-credentials.credentials")
		assert.FileExists(t, savedPath)

		// Check if metadata was saved
		metadataPath := filepath.Join(storeDir, "test-credentials.meta")
		assert.FileExists(t, metadataPath)

		// Verify saved content
		savedContent, err := os.ReadFile(savedPath)
		require.NoError(t, err)
		assert.Equal(t, testCredentialsContent, string(savedContent))

		// Check file permissions (should be 0600 for security)
		info, err := os.Stat(savedPath)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
	})

	t.Run("load AWS credentials", func(t *testing.T) {
		// Create a different target path
		targetPath := filepath.Join(tempDir, "loaded", "credentials")

		opts := &awsCredentialsOptions{
			name:       "test-credentials",
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
		assert.Equal(t, testCredentialsContent, string(loadedContent))

		// Check file permissions (should be 0600 for security)
		info, err := os.Stat(targetPath)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
	})

	t.Run("list AWS credentials", func(t *testing.T) {
		opts := &awsCredentialsOptions{
			storePath: storeDir,
		}

		err := opts.runList(nil, nil)
		assert.NoError(t, err)
	})
}

func TestAwsCredentialsMetadata(t *testing.T) {
	tempDir := t.TempDir()

	opts := &awsCredentialsOptions{
		name:        "test-metadata",
		description: "Test description",
		storePath:   tempDir,
	}

	// Test save metadata
	err := opts.saveMetadata()
	assert.NoError(t, err)

	metadataPath := filepath.Join(tempDir, "test-metadata.meta")
	assert.FileExists(t, metadataPath)

	// Check metadata file permissions (should be 0600 for security)
	info, err := os.Stat(metadataPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())

	// Test load metadata
	metadata := opts.loadMetadata("test-metadata")
	assert.Equal(t, "test-metadata", metadata.Name)
	assert.Equal(t, "Test description", metadata.Description)
	assert.False(t, metadata.SavedAt.IsZero())
}

func TestAwsCredentialsCopyFile(t *testing.T) {
	tempDir := t.TempDir()

	// Create source file
	srcPath := filepath.Join(tempDir, "source.credentials")
	content := `[default]
aws_access_key_id = AKIAIOSFODNN7EXAMPLE
aws_secret_access_key = wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY`
	err := os.WriteFile(srcPath, []byte(content), 0o600)
	require.NoError(t, err)

	// Test copy
	opts := &awsCredentialsOptions{}
	dstPath := filepath.Join(tempDir, "destination.credentials")

	err = opts.copyFile(srcPath, dstPath)
	assert.NoError(t, err)

	// Check content
	copiedContent, err := os.ReadFile(dstPath)
	require.NoError(t, err)
	assert.Equal(t, content, string(copiedContent))

	// Check permissions (should be 0600 for security)
	info, err := os.Stat(dstPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
}

func TestAwsCredentialsDisplayInfo(t *testing.T) {
	tempDir := t.TempDir()

	// Create test credentials

	credentialsPath := filepath.Join(tempDir, "credentials")
	err := os.WriteFile(credentialsPath, []byte(testCredentialsContent), 0o600)
	require.NoError(t, err)

	opts := &awsCredentialsOptions{}
	err = opts.displayCredentialsInfo(credentialsPath)
	assert.NoError(t, err)
}

func TestAwsCredentialsParseCredentials(t *testing.T) {
	opts := &awsCredentialsOptions{}

	testCredentials := testCredentialsContent

	profiles := opts.parseAwsCredentials(testCredentials)

	assert.Len(t, profiles, 3)

	// Check default profile
	defaultProfile := profiles[0]
	assert.Equal(t, "default", defaultProfile.Name)
	assert.True(t, defaultProfile.HasAccessKey)
	assert.True(t, defaultProfile.HasSecretKey)
	assert.False(t, defaultProfile.HasSessionToken)
	assert.False(t, defaultProfile.HasMfaSerial)

	// Check dev profile
	devProfile := profiles[1]
	assert.Equal(t, "dev", devProfile.Name)
	assert.True(t, devProfile.HasAccessKey)
	assert.True(t, devProfile.HasSecretKey)
	assert.True(t, devProfile.HasSessionToken)
	assert.True(t, devProfile.HasMfaSerial)

	// Check production profile
	prodProfile := profiles[2]
	assert.Equal(t, "production", prodProfile.Name)
	assert.True(t, prodProfile.HasAccessKey)
	assert.True(t, prodProfile.HasSecretKey)
	assert.False(t, prodProfile.HasSessionToken)
	assert.False(t, prodProfile.HasMfaSerial)
}

func TestAwsCredentialsErrorCases(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("save non-existent credentials", func(t *testing.T) {
		opts := &awsCredentialsOptions{
			name:       "test",
			configPath: "/non/existent/path",
			storePath:  tempDir,
		}

		err := opts.runSave(nil, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "AWS credentials file not found")
	})

	t.Run("load non-existent credentials", func(t *testing.T) {
		opts := &awsCredentialsOptions{
			name:      "non-existent",
			storePath: tempDir,
		}

		err := opts.runLoad(nil, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("save duplicate without force", func(t *testing.T) {
		// Create a test credentials file
		credentialsPath := filepath.Join(tempDir, "credentials")
		err := os.WriteFile(credentialsPath, []byte("[default]\naws_access_key_id=test"), 0o600)
		require.NoError(t, err)

		opts := &awsCredentialsOptions{
			name:       "duplicate-test",
			configPath: credentialsPath,
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
}
