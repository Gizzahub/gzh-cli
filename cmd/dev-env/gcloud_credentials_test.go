//nolint:testpackage // White-box testing needed for internal function access
package devenv

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultGcloudCredentialsOptions(t *testing.T) {
	opts := defaultGcloudCredentialsOptions()

	assert.NotEmpty(t, opts.configPath)
	assert.NotEmpty(t, opts.storePath)
	assert.False(t, opts.force)
	assert.False(t, opts.listAll)
}

func TestNewGcloudCredentialsCmd(t *testing.T) {
	cmd := newGcloudCredentialsCmd()

	assert.Equal(t, "gcloud-credentials", cmd.Use)
	assert.Equal(t, "Manage Google Cloud credentials", cmd.Short)
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

func TestGcloudCredentialsSaveCmd(t *testing.T) {
	cmd := newGcloudCredentialsSaveCmd()

	assert.Equal(t, "save", cmd.Use)
	assert.Equal(t, "Save current gcloud credentials", cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Test flags
	assert.NotNil(t, cmd.Flags().Lookup("name"))
	assert.NotNil(t, cmd.Flags().Lookup("description"))
	assert.NotNil(t, cmd.Flags().Lookup("config-path"))
	assert.NotNil(t, cmd.Flags().Lookup("store-path"))
	assert.NotNil(t, cmd.Flags().Lookup("force"))
}

func TestGcloudCredentialsLoadCmd(t *testing.T) {
	cmd := newGcloudCredentialsLoadCmd()

	assert.Equal(t, "load", cmd.Use)
	assert.Equal(t, "Load a saved gcloud credentials set", cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Test flags
	assert.NotNil(t, cmd.Flags().Lookup("name"))
	assert.NotNil(t, cmd.Flags().Lookup("config-path"))
	assert.NotNil(t, cmd.Flags().Lookup("store-path"))
	assert.NotNil(t, cmd.Flags().Lookup("force"))
}

func TestGcloudCredentialsListCmd(t *testing.T) {
	cmd := newGcloudCredentialsListCmd()

	assert.Equal(t, "list", cmd.Use)
	assert.Equal(t, "List saved gcloud credentials", cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Test flags
	assert.NotNil(t, cmd.Flags().Lookup("store-path"))
}

func TestGcloudCredentialsSaveLoad(t *testing.T) {
	// Create temporary directories for testing
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config")
	storeDir := filepath.Join(tempDir, "store")

	// Create a test gcloud credentials directory structure
	err := os.MkdirAll(configDir, 0o755)
	require.NoError(t, err)

	// Create application_default_credentials.json
	adcContent := map[string]interface{}{
		"client_id":     "123456789.apps.googleusercontent.com",
		"client_secret": "test-secret",
		"refresh_token": "test-refresh-token",
		"type":          "authorized_user",
		"client_email":  "test-service@my-project.iam.gserviceaccount.com",
	}
	adcData, err := json.MarshalIndent(adcContent, "", "  ")
	require.NoError(t, err)

	adcPath := filepath.Join(configDir, "application_default_credentials.json")
	err = os.WriteFile(adcPath, adcData, 0o600)
	require.NoError(t, err)

	// Create legacy_credentials directory
	legacyDir := filepath.Join(configDir, "legacy_credentials")
	err = os.MkdirAll(legacyDir, 0o700)
	require.NoError(t, err)

	legacyFile := filepath.Join(legacyDir, "test@example.com")
	legacyContent := `{
  "access_token": "test-access-token",
  "refresh_token": "test-refresh-token",
  "id_token": null,
  "token_expiry": "2024-01-01T00:00:00Z",
  "token_uri": "https://oauth2.googleapis.com/token",
  "client_id": "123456789.apps.googleusercontent.com",
  "client_secret": "test-secret",
  "scopes": ["https://www.googleapis.com/auth/cloud-platform"]
}`
	err = os.WriteFile(legacyFile, []byte(legacyContent), 0o600)
	require.NoError(t, err)

	// Create credentials.db (mock file)
	credDbPath := filepath.Join(configDir, "credentials.db")
	err = os.WriteFile(credDbPath, []byte("mock credentials database"), 0o600)
	require.NoError(t, err)

	// Create access_tokens.db (mock file)
	tokensDbPath := filepath.Join(configDir, "access_tokens.db")
	err = os.WriteFile(tokensDbPath, []byte("mock access tokens database"), 0o600)
	require.NoError(t, err)

	t.Run("save gcloud credentials", func(t *testing.T) {
		opts := &gcloudCredentialsOptions{
			name:        "test-credentials",
			description: "Test gcloud credentials",
			configPath:  configDir,
			storePath:   storeDir,
			force:       false,
		}

		err := opts.runSave(nil, nil)
		assert.NoError(t, err)

		// Check if directory was saved
		savedPath := filepath.Join(storeDir, "test-credentials")
		assert.DirExists(t, savedPath)

		// Check if metadata was saved
		metadataPath := filepath.Join(storeDir, "test-credentials.meta")
		assert.FileExists(t, metadataPath)

		// Check metadata file permissions (should be 0600 for security)
		info, err := os.Stat(metadataPath)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())

		// Verify credential files were copied
		savedAdcPath := filepath.Join(savedPath, "application_default_credentials.json")
		assert.FileExists(t, savedAdcPath)

		savedLegacyDir := filepath.Join(savedPath, "legacy_credentials")
		assert.DirExists(t, savedLegacyDir)

		savedCredDbPath := filepath.Join(savedPath, "credentials.db")
		assert.FileExists(t, savedCredDbPath)

		// Check file permissions (should be 0600 for security)
		adcInfo, err := os.Stat(savedAdcPath)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o600), adcInfo.Mode().Perm())
	})

	t.Run("load gcloud credentials", func(t *testing.T) {
		// Create a different target path
		targetPath := filepath.Join(tempDir, "loaded")

		opts := &gcloudCredentialsOptions{
			name:       "test-credentials",
			configPath: targetPath,
			storePath:  storeDir,
			force:      true, // Skip backup for test
		}

		err := opts.runLoad(nil, nil)
		assert.NoError(t, err)

		// Check if directory was created
		assert.DirExists(t, targetPath)

		// Check if credential files were loaded
		loadedAdcPath := filepath.Join(targetPath, "application_default_credentials.json")
		assert.FileExists(t, loadedAdcPath)

		loadedLegacyDir := filepath.Join(targetPath, "legacy_credentials")
		assert.DirExists(t, loadedLegacyDir)

		// Check file permissions (should be 0600 for security)
		adcInfo, err := os.Stat(loadedAdcPath)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o600), adcInfo.Mode().Perm())

		// Check directory permissions (should be 0700 for security)
		dirInfo, err := os.Stat(targetPath)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o700), dirInfo.Mode().Perm())
	})

	t.Run("list gcloud credentials", func(t *testing.T) {
		opts := &gcloudCredentialsOptions{
			storePath: storeDir,
		}

		err := opts.runList(nil, nil)
		assert.NoError(t, err)
	})
}

func TestGcloudCredentialsMetadata(t *testing.T) {
	tempDir := t.TempDir()

	opts := &gcloudCredentialsOptions{
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

func TestGcloudCredentialsCopyFile(t *testing.T) {
	tempDir := t.TempDir()

	// Create source file
	srcPath := filepath.Join(tempDir, "source.json")
	content := `{"test": "content", "secret": "value"}`
	err := os.WriteFile(srcPath, []byte(content), 0o600)
	require.NoError(t, err)

	// Test copy
	opts := &gcloudCredentialsOptions{}
	dstPath := filepath.Join(tempDir, "destination.json")

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

func TestGcloudCredentialsCopyDir(t *testing.T) {
	tempDir := t.TempDir()

	// Create source directory with files
	srcDir := filepath.Join(tempDir, "source")
	err := os.MkdirAll(filepath.Join(srcDir, "subdir"), 0o700)
	require.NoError(t, err)

	// Create credential files
	err = os.WriteFile(filepath.Join(srcDir, "cred1.json"), []byte("credential1"), 0o600)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(srcDir, "subdir", "cred2.json"), []byte("credential2"), 0o600)
	require.NoError(t, err)

	// Test copy directory
	opts := &gcloudCredentialsOptions{}
	dstDir := filepath.Join(tempDir, "destination")

	err = opts.copyDir(srcDir, dstDir)
	assert.NoError(t, err)

	// Check copied files
	assert.FileExists(t, filepath.Join(dstDir, "cred1.json"))
	assert.FileExists(t, filepath.Join(dstDir, "subdir", "cred2.json"))

	// Check content
	content1, err := os.ReadFile(filepath.Join(dstDir, "cred1.json"))
	require.NoError(t, err)
	assert.Equal(t, "credential1", string(content1))

	content2, err := os.ReadFile(filepath.Join(dstDir, "subdir", "cred2.json"))
	require.NoError(t, err)
	assert.Equal(t, "credential2", string(content2))

	// Check directory permissions (should be 0700 for security)
	dirInfo, err := os.Stat(dstDir)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o700), dirInfo.Mode().Perm())

	// Check file permissions (should be 0600 for security)
	fileInfo, err := os.Stat(filepath.Join(dstDir, "cred1.json"))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), fileInfo.Mode().Perm())
}

func TestGcloudCredentialsCopyCredentials(t *testing.T) {
	tempDir := t.TempDir()

	// Create source directory with mixed files
	srcDir := filepath.Join(tempDir, "source")
	err := os.MkdirAll(srcDir, 0o755)
	require.NoError(t, err)

	// Create credential files (should be copied)
	credentialFiles := map[string]string{
		"application_default_credentials.json": `{"type": "authorized_user"}`,
		"credentials.db":                       "credentials database",
		"access_tokens.db":                     "access tokens database",
	}

	for filename, content := range credentialFiles {
		err = os.WriteFile(filepath.Join(srcDir, filename), []byte(content), 0o600)
		require.NoError(t, err)
	}

	// Create legacy_credentials directory
	legacyDir := filepath.Join(srcDir, "legacy_credentials")
	err = os.MkdirAll(legacyDir, 0o700)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(legacyDir, "test@example.com"), []byte("legacy cred"), 0o600)
	require.NoError(t, err)

	// Create non-credential files (should NOT be copied)
	nonCredentialFiles := []string{
		"active_config",
		"configurations",
		"logs",
	}

	for _, filename := range nonCredentialFiles {
		err = os.WriteFile(filepath.Join(srcDir, filename), []byte("non-credential"), 0o644)
		require.NoError(t, err)
	}

	// Test copy credentials
	opts := &gcloudCredentialsOptions{}
	dstDir := filepath.Join(tempDir, "destination")

	err = opts.copyCredentials(srcDir, dstDir)
	assert.NoError(t, err)

	// Check that credential files were copied
	for filename := range credentialFiles {
		assert.FileExists(t, filepath.Join(dstDir, filename))
	}

	assert.DirExists(t, filepath.Join(dstDir, "legacy_credentials"))

	// Check that non-credential files were NOT copied
	for _, filename := range nonCredentialFiles {
		assert.NoFileExists(t, filepath.Join(dstDir, filename))
	}

	// Check directory permissions (should be 0700 for security)
	dirInfo, err := os.Stat(dstDir)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o700), dirInfo.Mode().Perm())
}

func TestGcloudCredentialsDisplayInfo(t *testing.T) {
	tempDir := t.TempDir()

	// Create test credentials structure
	credentialsDir := filepath.Join(tempDir, "credentials")
	err := os.MkdirAll(credentialsDir, 0o700)
	require.NoError(t, err)

	// Create application_default_credentials.json with service account info
	adcContent := map[string]interface{}{
		"type":         "service_account",
		"client_email": "test-service@my-project.iam.gserviceaccount.com",
		"private_key":  "-----BEGIN PRIVATE KEY-----\ntest-key\n-----END PRIVATE KEY-----\n",
	}
	adcData, err := json.MarshalIndent(adcContent, "", "  ")
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(credentialsDir, "application_default_credentials.json"), adcData, 0o600)
	require.NoError(t, err)

	// Create other credential files
	err = os.WriteFile(filepath.Join(credentialsDir, "credentials.db"), []byte("mock db"), 0o600)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(credentialsDir, "access_tokens.db"), []byte("mock tokens"), 0o600)
	require.NoError(t, err)

	// Create legacy_credentials directory
	legacyDir := filepath.Join(credentialsDir, "legacy_credentials")
	err = os.MkdirAll(legacyDir, 0o700)
	require.NoError(t, err)

	opts := &gcloudCredentialsOptions{}
	opts.displayCredentialsInfo(credentialsDir)
}

func TestGcloudCredentialsErrorCases(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("save non-existent credentials", func(t *testing.T) {
		opts := &gcloudCredentialsOptions{
			name:       "test",
			configPath: "/non/existent/path",
			storePath:  tempDir,
		}

		err := opts.runSave(nil, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "gcloud config directory not found")
	})

	t.Run("load non-existent credentials", func(t *testing.T) {
		opts := &gcloudCredentialsOptions{
			name:      "non-existent",
			storePath: tempDir,
		}

		err := opts.runLoad(nil, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("save duplicate without force", func(t *testing.T) {
		// Create a test credentials directory
		credentialsPath := filepath.Join(tempDir, "gcloud")
		err := os.MkdirAll(credentialsPath, 0o700)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(credentialsPath, "application_default_credentials.json"), []byte("{}"), 0o600)
		require.NoError(t, err)

		opts := &gcloudCredentialsOptions{
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

	t.Run("getDirSize", func(t *testing.T) {
		// Create test directory with files
		testDir := filepath.Join(tempDir, "size-test")
		err := os.MkdirAll(testDir, 0o755)
		require.NoError(t, err)

		err = os.WriteFile(filepath.Join(testDir, "file1.txt"), []byte("hello"), 0o644)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(testDir, "file2.txt"), []byte("world"), 0o644)
		require.NoError(t, err)

		opts := &gcloudCredentialsOptions{}
		size, err := opts.getDirSize(testDir)
		assert.NoError(t, err)
		assert.Equal(t, int64(10), size) // "hello" (5) + "world" (5)
	})

	t.Run("mergeCredentials", func(t *testing.T) {
		// Create source and destination directories
		srcDir := filepath.Join(tempDir, "merge-src")
		dstDir := filepath.Join(tempDir, "merge-dst")

		err := os.MkdirAll(srcDir, 0o700)
		require.NoError(t, err)
		err = os.MkdirAll(dstDir, 0o700)
		require.NoError(t, err)

		// Create source credential file
		err = os.WriteFile(filepath.Join(srcDir, "application_default_credentials.json"), []byte("new creds"), 0o600)
		require.NoError(t, err)

		// Create existing file in destination
		err = os.WriteFile(filepath.Join(dstDir, "existing.txt"), []byte("existing"), 0o644)
		require.NoError(t, err)

		opts := &gcloudCredentialsOptions{}
		err = opts.mergeCredentials(srcDir, dstDir)
		assert.NoError(t, err)

		// Check that source file was copied
		assert.FileExists(t, filepath.Join(dstDir, "application_default_credentials.json"))

		// Check that existing file is still there
		assert.FileExists(t, filepath.Join(dstDir, "existing.txt"))

		// Check content
		content, err := os.ReadFile(filepath.Join(dstDir, "application_default_credentials.json"))
		require.NoError(t, err)
		assert.Equal(t, "new creds", string(content))
	})
}
