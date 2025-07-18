package devenv

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssooidc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/ini.v1"
)

func TestAWSProfileManager(t *testing.T) {
	// Create temporary AWS config directory
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".aws")
	require.NoError(t, os.MkdirAll(configDir, 0o755))

	// Create test config file
	configPath := filepath.Join(configDir, "config")
	cfg := ini.Empty()

	// Add default profile
	defaultSection, err := cfg.NewSection("default")
	require.NoError(t, err)
	defaultSection.NewKey("region", "us-east-1")
	defaultSection.NewKey("output", "json")

	// Add SSO profile
	ssoSection, err := cfg.NewSection("profile production")
	require.NoError(t, err)
	ssoSection.NewKey("sso_start_url", "https://mycompany.awsapps.com/start")
	ssoSection.NewKey("sso_region", "us-east-1")
	ssoSection.NewKey("sso_account_id", "123456789012")
	ssoSection.NewKey("sso_role_name", "AdministratorAccess")
	ssoSection.NewKey("region", "us-west-2")
	ssoSection.NewKey("output", "json")

	// Add assume role profile
	roleSection, err := cfg.NewSection("profile staging")
	require.NoError(t, err)
	roleSection.NewKey("role_arn", "arn:aws:iam::123456789012:role/StagingRole")
	roleSection.NewKey("source_profile", "default")
	roleSection.NewKey("region", "eu-west-1")
	roleSection.NewKey("mfa_serial", "arn:aws:iam::123456789012:mfa/user")

	// Save config file
	require.NoError(t, cfg.SaveTo(configPath))

	// Create credentials file
	credentialsPath := filepath.Join(configDir, "credentials")
	credCfg := ini.Empty()

	defaultCred, err := credCfg.NewSection("default")
	require.NoError(t, err)
	defaultCred.NewKey("aws_access_key_id", "AKIAIOSFODNN7EXAMPLE")
	defaultCred.NewKey("aws_secret_access_key", "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY")

	require.NoError(t, credCfg.SaveTo(credentialsPath))

	// Override home directory for testing
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	t.Run("LoadProfiles", func(t *testing.T) {
		ctx := context.Background()
		manager, err := NewAWSProfileManager(ctx)
		require.NoError(t, err)

		profiles := manager.ListProfiles()
		assert.Len(t, profiles, 3) // default, production, staging

		// Check default profile
		defaultProfile, err := manager.GetProfile("default")
		require.NoError(t, err)
		assert.Equal(t, "default", defaultProfile.Name)
		assert.Equal(t, "us-east-1", defaultProfile.Region)
		assert.Equal(t, "json", defaultProfile.Output)

		// Check SSO profile
		ssoProfile, err := manager.GetProfile("production")
		require.NoError(t, err)
		assert.Equal(t, "production", ssoProfile.Name)
		assert.Equal(t, "https://mycompany.awsapps.com/start", ssoProfile.SSOStartURL)
		assert.Equal(t, "123456789012", ssoProfile.SSOAccountID)
		assert.Equal(t, "AdministratorAccess", ssoProfile.SSORoleName)
		assert.Equal(t, "us-west-2", ssoProfile.Region)

		// Check assume role profile
		roleProfile, err := manager.GetProfile("staging")
		require.NoError(t, err)
		assert.Equal(t, "staging", roleProfile.Name)
		assert.Equal(t, "arn:aws:iam::123456789012:role/StagingRole", roleProfile.RoleArn)
		assert.Equal(t, "default", roleProfile.SourceProfile)
		assert.Equal(t, "arn:aws:iam::123456789012:mfa/user", roleProfile.MFASerial)
		assert.Equal(t, "eu-west-1", roleProfile.Region)
	})

	t.Run("GetProfile_NotFound", func(t *testing.T) {
		ctx := context.Background()
		manager, err := NewAWSProfileManager(ctx)
		require.NoError(t, err)

		_, err = manager.GetProfile("nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "profile 'nonexistent' not found")
	})

	t.Run("ListProfiles_Sorted", func(t *testing.T) {
		ctx := context.Background()
		manager, err := NewAWSProfileManager(ctx)
		require.NoError(t, err)

		profiles := manager.ListProfiles()
		require.Len(t, profiles, 3)

		// Check that profiles are sorted alphabetically
		assert.Equal(t, "default", profiles[0].Name)
		assert.Equal(t, "production", profiles[1].Name)
		assert.Equal(t, "staging", profiles[2].Name)
	})

	t.Run("SwitchProfile", func(t *testing.T) {
		ctx := context.Background()
		manager, err := NewAWSProfileManager(ctx)
		require.NoError(t, err)

		// Switch to production profile
		err = manager.SwitchProfile("production")
		require.NoError(t, err)

		// Check that AWS_PROFILE environment variable is set
		assert.Equal(t, "production", os.Getenv("AWS_PROFILE"))
	})

	t.Run("ActiveProfile", func(t *testing.T) {
		// Set AWS_PROFILE environment variable
		os.Setenv("AWS_PROFILE", "staging")
		defer os.Unsetenv("AWS_PROFILE")

		ctx := context.Background()
		manager, err := NewAWSProfileManager(ctx)
		require.NoError(t, err)

		// Check that staging profile is marked as active
		stagingProfile, err := manager.GetProfile("staging")
		require.NoError(t, err)
		assert.True(t, stagingProfile.IsActive)

		// Check that other profiles are not active
		defaultProfile, err := manager.GetProfile("default")
		require.NoError(t, err)
		assert.False(t, defaultProfile.IsActive)

		productionProfile, err := manager.GetProfile("production")
		require.NoError(t, err)
		assert.False(t, productionProfile.IsActive)
	})

	t.Run("SaveSSOToken", func(t *testing.T) {
		ctx := context.Background()
		manager, err := NewAWSProfileManager(ctx)
		require.NoError(t, err)

		profile, err := manager.GetProfile("production")
		require.NoError(t, err)

		// Mock SSO token response
		accessToken := "test-access-token"
		expiresIn := int32(3600)
		refreshToken := "test-refresh-token"

		// Create a mock token response
		tokenResp := &ssooidc.CreateTokenOutput{
			AccessToken:  aws.String(accessToken),
			ExpiresIn:    expiresIn,
			RefreshToken: aws.String(refreshToken),
		}

		// Save token
		err = manager.saveSSOToken(profile, tokenResp)
		require.NoError(t, err)

		// Verify token was saved
		cacheDir := filepath.Join(tmpDir, ".aws", "sso", "cache")
		files, err := os.ReadDir(cacheDir)
		require.NoError(t, err)
		assert.Len(t, files, 1)
	})
}

func TestAWSProfileCommands(t *testing.T) {
	t.Run("ListCommand", func(t *testing.T) {
		cmd := newAWSProfileListCmd()
		assert.NotNil(t, cmd)
		assert.Equal(t, "list", cmd.Use)
		assert.NotEmpty(t, cmd.Short)
	})

	t.Run("SwitchCommand", func(t *testing.T) {
		cmd := newAWSProfileSwitchCmd()
		assert.NotNil(t, cmd)
		assert.Equal(t, "switch [profile-name]", cmd.Use)
		assert.NotEmpty(t, cmd.Short)

		// Check flags
		interactive := cmd.Flags().Lookup("interactive")
		assert.NotNil(t, interactive)
		assert.Equal(t, "false", interactive.DefValue)
	})

	t.Run("LoginCommand", func(t *testing.T) {
		cmd := newAWSProfileLoginCmd()
		assert.NotNil(t, cmd)
		assert.Equal(t, "login [profile-name]", cmd.Use)
		assert.NotEmpty(t, cmd.Short)
	})

	t.Run("ShowCommand", func(t *testing.T) {
		cmd := newAWSProfileShowCmd()
		assert.NotNil(t, cmd)
		assert.Equal(t, "show [profile-name]", cmd.Use)
		assert.NotEmpty(t, cmd.Short)
	})

	t.Run("ValidateCommand", func(t *testing.T) {
		cmd := newAWSProfileValidateCmd()
		assert.NotNil(t, cmd)
		assert.Equal(t, "validate [profile-name]", cmd.Use)
		assert.NotEmpty(t, cmd.Short)

		// Check flags
		all := cmd.Flags().Lookup("all")
		assert.NotNil(t, all)
		assert.Equal(t, "false", all.DefValue)
	})

	t.Run("MainCommand", func(t *testing.T) {
		cmd := newAWSProfileCmd()
		assert.NotNil(t, cmd)
		assert.Equal(t, "aws-profile", cmd.Use)
		assert.NotEmpty(t, cmd.Short)
		assert.NotEmpty(t, cmd.Long)

		// Check subcommands
		subcommands := cmd.Commands()
		assert.Len(t, subcommands, 5) // list, switch, login, show, validate
	})
}

func TestAWSProfile_Serialization(t *testing.T) {
	now := time.Now()
	profile := &AWSProfile{
		Name:          "test-profile",
		Region:        "us-west-2",
		Output:        "json",
		SourceProfile: "default",
		RoleArn:       "arn:aws:iam::123456789012:role/TestRole",
		MFASerial:     "arn:aws:iam::123456789012:mfa/user",
		SSOStartURL:   "https://test.awsapps.com/start",
		SSORoleName:   "TestRole",
		SSOAccountID:  "123456789012",
		SSORegion:     "us-east-1",
		Tags: map[string]string{
			"environment": "test",
			"team":        "platform",
		},
		LastUsed: &now,
		IsActive: true,
	}

	// Test JSON serialization
	data, err := json.Marshal(profile)
	require.NoError(t, err)

	// Test JSON deserialization
	var decoded AWSProfile
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, profile.Name, decoded.Name)
	assert.Equal(t, profile.Region, decoded.Region)
	assert.Equal(t, profile.SSOStartURL, decoded.SSOStartURL)
	assert.Equal(t, profile.Tags, decoded.Tags)
	assert.Equal(t, profile.IsActive, decoded.IsActive)
}

func TestUpdateShellConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test shell config files
	bashrcPath := filepath.Join(tmpDir, ".bashrc")
	err := os.WriteFile(bashrcPath, []byte("# Test bashrc\nexport PATH=/usr/local/bin:$PATH\n"), 0o644)
	require.NoError(t, err)

	zshrcPath := filepath.Join(tmpDir, ".zshrc")
	err = os.WriteFile(zshrcPath, []byte("# Test zshrc\nexport AWS_PROFILE=old-profile\n"), 0o644)
	require.NoError(t, err)

	// Override home directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	ctx := context.Background()
	manager := &AWSProfileManager{
		configPath:      filepath.Join(tmpDir, ".aws", "config"),
		credentialsPath: filepath.Join(tmpDir, ".aws", "credentials"),
		profiles:        make(map[string]*AWSProfile),
		ctx:             ctx,
	}

	// Update shell config
	err = manager.updateShellConfig("new-profile")
	require.NoError(t, err)

	// Check bashrc - should have new export added
	bashrcContent, err := os.ReadFile(bashrcPath)
	require.NoError(t, err)
	assert.Contains(t, string(bashrcContent), "export AWS_PROFILE=new-profile")

	// Check zshrc - should have existing export updated
	zshrcContent, err := os.ReadFile(zshrcPath)
	require.NoError(t, err)
	assert.Contains(t, string(zshrcContent), "export AWS_PROFILE=new-profile")
	assert.NotContains(t, string(zshrcContent), "export AWS_PROFILE=old-profile")
}
