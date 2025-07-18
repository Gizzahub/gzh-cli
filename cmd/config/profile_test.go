package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateProfileName(t *testing.T) {
	tests := []struct {
		name        string
		profileName string
		expectError bool
	}{
		{"valid name", "dev", false},
		{"valid name with dash", "staging-env", false},
		{"empty name", "", true},
		{"invalid characters", "dev/test", true},
		{"reserved name", "yaml", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateProfileName(tt.profileName)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetProfilePath(t *testing.T) {
	tests := []struct {
		profileName  string
		expectedPath string
	}{
		{"dev", "gzh.dev.yaml"},
		{"staging", "gzh.staging.yaml"},
		{"prod", "gzh.prod.yaml"},
	}

	for _, tt := range tests {
		t.Run(tt.profileName, func(t *testing.T) {
			path := getProfilePath(tt.profileName)
			assert.Equal(t, tt.expectedPath, path)
		})
	}
}

func TestGenerateProfileTemplate(t *testing.T) {
	template := generateProfileTemplate("dev")

	assert.Contains(t, template, "# gzh.yaml - Dev profile")
	assert.Contains(t, template, "version: \"1.0.0\"")
	assert.Contains(t, template, "default_provider: github")
	assert.Contains(t, template, "token: \"${GITHUB_TOKEN_DEV}\"")
	assert.Contains(t, template, "clone_dir: \"${HOME}/repos/dev/github\"")
}

func TestCreateProfile(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "profile-test-*")
	require.NoError(t, err)

	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	defer os.Chdir(originalDir)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Test creating a new profile
	err = createProfile("dev", "", false)
	assert.NoError(t, err)

	// Check file was created
	profileFile := getProfilePath("dev")
	assert.FileExists(t, profileFile)

	// Test creating duplicate profile should fail
	err = createProfile("dev", "", false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestCreateProfileFromExisting(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "profile-from-test-*")
	require.NoError(t, err)

	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	defer os.Chdir(originalDir)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create source profile
	sourceContent := `version: "1.0.0"
default_provider: github
providers:
  github:
    token: "${GITHUB_TOKEN}"
    orgs:
      - name: "test-org"
`
	sourceFile := getProfilePath("source")
	err = os.WriteFile(sourceFile, []byte(sourceContent), 0o644)
	require.NoError(t, err)

	// Create profile from existing
	err = createProfile("target", "source", false)
	assert.NoError(t, err)

	// Check target file was created with same content
	targetFile := getProfilePath("target")
	assert.FileExists(t, targetFile)

	targetContent, err := os.ReadFile(targetFile)
	require.NoError(t, err)
	assert.Equal(t, sourceContent, string(targetContent))
}

func TestGetAvailableProfiles(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "profiles-list-test-*")
	require.NoError(t, err)

	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	defer os.Chdir(originalDir)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create test profile files
	profiles := []string{"dev", "staging", "prod"}
	for _, profile := range profiles {
		profileFile := getProfilePath(profile)
		err = os.WriteFile(profileFile, []byte("test content"), 0o644)
		require.NoError(t, err)
	}

	// Create non-profile files (should be ignored)
	err = os.WriteFile("gzh.yaml", []byte("main config"), 0o644)
	require.NoError(t, err)
	err = os.WriteFile("other.yaml", []byte("other file"), 0o644)
	require.NoError(t, err)

	// Get available profiles
	availableProfiles, err := getAvailableProfiles()
	assert.NoError(t, err)
	assert.Len(t, availableProfiles, 3)

	// Check all expected profiles are present
	for _, expected := range profiles {
		assert.Contains(t, availableProfiles, expected)
	}
}

func TestUseProfile(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "use-profile-test-*")
	require.NoError(t, err)

	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	defer os.Chdir(originalDir)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create valid profile
	profileContent := `version: "1.0.0"
default_provider: github
providers:
  github:
    token: "test-token"
    orgs:
      - name: "test-org"
`
	profileFile := getProfilePath("dev")
	err = os.WriteFile(profileFile, []byte(profileContent), 0o644)
	require.NoError(t, err)

	// Use profile
	err = useProfile("dev")
	assert.NoError(t, err)

	// Check symlink was created
	linkTarget, err := os.Readlink("gzh.yaml")
	assert.NoError(t, err)
	assert.Equal(t, profileFile, linkTarget)

	// Test using non-existent profile
	err = useProfile("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

func TestGetCurrentProfile(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "current-profile-test-*")
	require.NoError(t, err)

	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	defer os.Chdir(originalDir)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// No active profile initially
	current, err := getCurrentProfile()
	assert.NoError(t, err)
	assert.Equal(t, "", current)

	// Create and use profile
	profileFile := getProfilePath("dev")
	err = os.WriteFile(profileFile, []byte("test"), 0o644)
	require.NoError(t, err)

	err = os.Symlink(profileFile, "gzh.yaml")
	require.NoError(t, err)

	// Should detect current profile
	current, err = getCurrentProfile()
	assert.NoError(t, err)
	assert.Equal(t, "dev", current)
}

func TestDeleteProfile(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "delete-profile-test-*")
	require.NoError(t, err)

	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	defer os.Chdir(originalDir)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create profile
	profileFile := getProfilePath("test")
	err = os.WriteFile(profileFile, []byte("test content"), 0o644)
	require.NoError(t, err)

	// Delete profile with force
	err = deleteProfile("test", true)
	assert.NoError(t, err)

	// Check file was deleted
	_, err = os.Stat(profileFile)
	assert.True(t, os.IsNotExist(err))

	// Test deleting non-existent profile
	err = deleteProfile("nonexistent", true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}
