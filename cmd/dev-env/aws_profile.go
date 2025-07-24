// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package devenv

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssooidc"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/manifoldco/promptui"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"gopkg.in/ini.v1"

	"github.com/gizzahub/gzh-manager-go/internal/env"
)

const (
	outputFormatTable = "table"
	statusActive      = "✓"
)

// AWSProfile represents an AWS profile configuration.
type AWSProfile struct {
	Name              string            `json:"name"`
	Region            string            `json:"region"`
	Output            string            `json:"output"`
	SourceProfile     string            `json:"sourceProfile,omitempty"`
	RoleArn           string            `json:"roleArn,omitempty"`
	MFASerial         string            `json:"mfaSerial,omitempty"`
	SSOStartURL       string            `json:"ssoStartUrl,omitempty"`
	SSORoleName       string            `json:"ssoRoleName,omitempty"`
	SSOAccountID      string            `json:"ssoAccountId,omitempty"`
	SSORegion         string            `json:"ssoRegion,omitempty"`
	CredentialProcess string            `json:"credentialProcess,omitempty"`
	ExternalID        string            `json:"externalId,omitempty"`
	DurationSeconds   int               `json:"durationSeconds,omitempty"`
	Tags              map[string]string `json:"tags,omitempty"`
	LastUsed          *time.Time        `json:"lastUsed,omitempty"`
	IsActive          bool              `json:"isActive"`
}

// AWSProfileManager manages AWS profiles.
type AWSProfileManager struct {
	configPath      string
	credentialsPath string
	profiles        map[string]*AWSProfile
	ctx             context.Context
}

// NewAWSProfileManager creates a new AWS profile manager.
func NewAWSProfileManager(ctx context.Context) (*AWSProfileManager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	manager := &AWSProfileManager{
		configPath:      filepath.Join(homeDir, ".aws", "config"),
		credentialsPath: filepath.Join(homeDir, ".aws", "credentials"),
		profiles:        make(map[string]*AWSProfile),
		ctx:             ctx,
	}

	if err := manager.loadProfiles(); err != nil {
		return nil, fmt.Errorf("failed to load profiles: %w", err)
	}

	return manager, nil
}

// loadProfiles loads all AWS profiles from config file.
func (m *AWSProfileManager) loadProfiles() error { //nolint:gocognit,gocyclo // Complex AWS profile parsing with multiple configuration options and branching paths
	cfg, err := ini.Load(m.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Config file doesn't exist, create empty one
			return nil
		}

		return fmt.Errorf("failed to load config file: %w", err)
	}

	// Get current profile from environment
	currentProfile := os.Getenv("AWS_PROFILE")
	if currentProfile == "" {
		currentProfile = "default"
	}

	for _, section := range cfg.Sections() {
		if section.Name() == "DEFAULT" {
			continue
		}

		profileName := section.Name()
		profileName = strings.TrimPrefix(profileName, "profile ")

		profile := &AWSProfile{
			Name:     profileName,
			IsActive: profileName == currentProfile,
		}

		// Load profile settings
		if key, err := section.GetKey("region"); err == nil {
			profile.Region = key.String()
		}

		if key, err := section.GetKey("output"); err == nil {
			profile.Output = key.String()
		}

		if key, err := section.GetKey("source_profile"); err == nil {
			profile.SourceProfile = key.String()
		}

		if key, err := section.GetKey("role_arn"); err == nil {
			profile.RoleArn = key.String()
		}

		if key, err := section.GetKey("mfa_serial"); err == nil {
			profile.MFASerial = key.String()
		}

		if key, err := section.GetKey("sso_start_url"); err == nil {
			profile.SSOStartURL = key.String()
		}

		if key, err := section.GetKey("sso_role_name"); err == nil {
			profile.SSORoleName = key.String()
		}

		if key, err := section.GetKey("sso_account_id"); err == nil {
			profile.SSOAccountID = key.String()
		}

		if key, err := section.GetKey("sso_region"); err == nil {
			profile.SSORegion = key.String()
		}

		if key, err := section.GetKey("credential_process"); err == nil {
			profile.CredentialProcess = key.String()
		}

		if key, err := section.GetKey("external_id"); err == nil {
			profile.ExternalID = key.String()
		}

		if key, err := section.GetKey("duration_seconds"); err == nil {
			profile.DurationSeconds = key.MustInt(3600)
		}

		m.profiles[profileName] = profile
	}

	return nil
}

// ListProfiles returns all AWS profiles.
func (m *AWSProfileManager) ListProfiles() []*AWSProfile {
	profiles := make([]*AWSProfile, 0, len(m.profiles))
	for _, profile := range m.profiles {
		profiles = append(profiles, profile)
	}

	sort.Slice(profiles, func(i, j int) bool {
		return profiles[i].Name < profiles[j].Name
	})

	return profiles
}

// GetProfile returns a specific profile.
func (m *AWSProfileManager) GetProfile(name string) (*AWSProfile, error) {
	profile, exists := m.profiles[name]
	if !exists {
		return nil, fmt.Errorf("profile '%s' not found", name)
	}

	return profile, nil
}

// SwitchProfile switches to a different AWS profile.
func (m *AWSProfileManager) SwitchProfile(profileName string) error {
	profile, err := m.GetProfile(profileName)
	if err != nil {
		return err
	}

	// Set environment variable
	if err := os.Setenv("AWS_PROFILE", profileName); err != nil {
		return fmt.Errorf("failed to set AWS_PROFILE: %w", err)
	}

	// Update shell configuration files
	if err := m.updateShellConfig(profileName); err != nil {
		return fmt.Errorf("failed to update shell config: %w", err)
	}

	// Validate credentials if needed
	if profile.SSOStartURL != "" {
		if err := m.validateSSOCredentials(profile); err != nil {
			fmt.Printf("Warning: SSO credentials may need refresh: %v\n", err)
		}
	}

	fmt.Printf("Switched to AWS profile: %s\n", profileName)

	return nil
}

// updateShellConfig updates shell configuration files with the new profile.
func (m *AWSProfileManager) updateShellConfig(profileName string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// Update common shell configs
	shellConfigs := []string{
		filepath.Join(homeDir, ".bashrc"),
		filepath.Join(homeDir, ".zshrc"),
		filepath.Join(homeDir, ".profile"),
	}

	exportLine := fmt.Sprintf("export AWS_PROFILE=%s", profileName)

	for _, configFile := range shellConfigs {
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			continue
		}

		// Read existing content
		content, err := os.ReadFile(filepath.Clean(configFile))
		if err != nil {
			continue
		}

		lines := strings.Split(string(content), "\n")
		found := false

		for i, line := range lines {
			if strings.HasPrefix(line, "export AWS_PROFILE=") {
				lines[i] = exportLine
				found = true

				break
			}
		}

		if !found {
			lines = append(lines, exportLine)
		}

		// Write back
		if err := os.WriteFile(configFile, []byte(strings.Join(lines, "\n")), 0o600); err != nil {
			return fmt.Errorf("failed to update %s: %w", configFile, err)
		}
	}

	return nil
}

// validateSSOCredentials validates SSO credentials for a profile.
func (m *AWSProfileManager) validateSSOCredentials(profile *AWSProfile) error {
	cfg, err := config.LoadDefaultConfig(m.ctx,
		config.WithRegion(profile.SSORegion),
		config.WithSharedConfigProfile(profile.Name),
	)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Try to get caller identity to validate credentials
	stsClient := sts.NewFromConfig(cfg)

	_, err = stsClient.GetCallerIdentity(m.ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return fmt.Errorf("credentials validation failed: %w", err)
	}

	return nil
}

// LoginSSO performs SSO login for a profile.
func (m *AWSProfileManager) LoginSSO(profileName string) error {
	profile, err := m.GetProfile(profileName)
	if err != nil {
		return err
	}

	if profile.SSOStartURL == "" {
		return fmt.Errorf("profile '%s' is not configured for SSO", profileName)
	}

	// Create SSO OIDC client
	cfg, err := config.LoadDefaultConfig(m.ctx,
		config.WithRegion(profile.SSORegion),
	)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	oidcClient := ssooidc.NewFromConfig(cfg)

	// Register client
	registerResp, err := oidcClient.RegisterClient(m.ctx, &ssooidc.RegisterClientInput{
		ClientName: aws.String("gzh-manager"),
		ClientType: aws.String("public"),
	})
	if err != nil {
		return fmt.Errorf("failed to register client: %w", err)
	}

	// Start device authorization
	startResp, err := oidcClient.StartDeviceAuthorization(m.ctx, &ssooidc.StartDeviceAuthorizationInput{
		ClientId:     registerResp.ClientId,
		ClientSecret: registerResp.ClientSecret,
		StartUrl:     aws.String(profile.SSOStartURL),
	})
	if err != nil {
		return fmt.Errorf("failed to start device authorization: %w", err)
	}

	// Display user code and verification URL
	fmt.Printf("\nSSO Login Required\n")
	fmt.Printf("Please visit: %s\n", *startResp.VerificationUriComplete)
	fmt.Printf("User Code: %s\n", *startResp.UserCode)
	fmt.Printf("\nWaiting for authentication...\n")

	// Poll for token
	var tokenResp *ssooidc.CreateTokenOutput
	for {
		tokenResp, err = oidcClient.CreateToken(m.ctx, &ssooidc.CreateTokenInput{
			ClientId:     registerResp.ClientId,
			ClientSecret: registerResp.ClientSecret,
			DeviceCode:   startResp.DeviceCode,
			GrantType:    aws.String("urn:ietf:params:oauth:grant-type:device_code"),
		})
		if err != nil {
			if strings.Contains(err.Error(), "AuthorizationPending") {
				time.Sleep(time.Duration(startResp.Interval) * time.Second)
				continue
			}

			return fmt.Errorf("failed to create token: %w", err)
		}

		break
	}

	// Save SSO token
	if err := m.saveSSOToken(profile, tokenResp); err != nil {
		return fmt.Errorf("failed to save SSO token: %w", err)
	}

	fmt.Printf("Successfully logged in to SSO for profile: %s\n", profileName)

	return nil
}

// saveSSOToken saves SSO token to cache.
func (m *AWSProfileManager) saveSSOToken(profile *AWSProfile, tokenResp *ssooidc.CreateTokenOutput) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// Create cache directory
	cacheDir := filepath.Join(homeDir, ".aws", "sso", "cache")
	if err := os.MkdirAll(cacheDir, 0o700); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Create token cache entry
	token := map[string]interface{}{
		"accessToken":  *tokenResp.AccessToken,
		"expiresAt":    time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second).Format(time.RFC3339),
		"region":       profile.SSORegion,
		"startUrl":     profile.SSOStartURL,
		"refreshToken": tokenResp.RefreshToken,
	}

	// Save to cache file
	fileName := fmt.Sprintf("%x.json", profile.SSOStartURL)
	cacheFile := filepath.Join(cacheDir, fileName)

	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}

	if err := os.WriteFile(cacheFile, data, 0o600); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// newAWSProfileCmd creates the AWS profile management command.
func newAWSProfileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "aws-profile",
		Short: "Manage AWS profiles with SSO support",
		Long: `Manage AWS profiles including SSO integration, multi-account switching,
and automatic credential renewal.

Features:
- List and switch between AWS profiles
- SSO login and token management
- Automatic credential renewal
- Profile validation and health checks`,
	}

	cmd.AddCommand(
		newAWSProfileListCmd(),
		newAWSProfileSwitchCmd(),
		newAWSProfileLoginCmd(),
		newAWSProfileShowCmd(),
		newAWSProfileValidateCmd(),
	)

	return cmd
}

// newAWSProfileListCmd creates the list command.
func newAWSProfileListCmd() *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all AWS profiles",
		Long:  "List all AWS profiles from ~/.aws/config with their configuration details",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			manager, err := NewAWSProfileManager(ctx)
			if err != nil {
				return err
			}

			profiles := manager.ListProfiles()

			switch outputFormat {
			case env.FormatJSON:
				data, err := json.MarshalIndent(profiles, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
			case outputFormatTable:
				fallthrough //nolint:gocritic // Intended fallthrough to default case
			default:
				table := tablewriter.NewWriter(os.Stdout)
				table.Header("Profile", "Region", "Type", "SSO URL", "Account ID", "Active")

				for _, profile := range profiles {
					var profileType string
					switch {
					case profile.SSOStartURL != "":
						profileType = "SSO"
					case profile.RoleArn != "":
						profileType = "AssumeRole"
					case profile.CredentialProcess != "":
						profileType = "Process"
					default:
						profileType = "Standard"
					}

					active := ""
					if profile.IsActive {
						active = statusActive
					}

					if err := table.Append(
						profile.Name,
						profile.Region,
						profileType,
						profile.SSOStartURL,
						profile.SSOAccountID,
						active,
					); err != nil {
						// Log error but continue processing
						fmt.Fprintf(os.Stderr, "Warning: failed to add profile to table: %v\n", err)
					}
				}

				_ = table.Render() // Ignore render error
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json)")

	return cmd
}

// newAWSProfileSwitchCmd creates the switch command.
func newAWSProfileSwitchCmd() *cobra.Command {
	var interactive bool

	cmd := &cobra.Command{
		Use:   "switch [profile-name]",
		Short: "Switch to a different AWS profile",
		Long:  "Switch to a different AWS profile and update environment variables",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			manager, err := NewAWSProfileManager(ctx)
			if err != nil {
				return err
			}

			var profileName string
			switch {
			case len(args) > 0:
				profileName = args[0]
			case interactive:
				// Interactive profile selection
				profiles := manager.ListProfiles()
				profileNames := make([]string, len(profiles))
				for i, p := range profiles {
					desc := p.Name
					if p.Region != "" {
						desc += fmt.Sprintf(" (%s)", p.Region)
					}
					if p.IsActive {
						desc += env.CurrentSuffix
					}
					profileNames[i] = desc
				}

				prompt := promptui.Select{
					Label: "Select AWS Profile",
					Items: profileNames,
				}

				idx, _, err := prompt.Run()
				if err != nil {
					return err
				}

				profileName = profiles[idx].Name
			default:
				return fmt.Errorf("profile name required or use --interactive")
			}

			return manager.SwitchProfile(profileName)
		},
	}

	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Interactive profile selection")

	return cmd
}

// newAWSProfileLoginCmd creates the login command.
func newAWSProfileLoginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login [profile-name]",
		Short: "Login to AWS SSO for a profile",
		Long:  "Perform AWS SSO login for profiles configured with SSO",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			manager, err := NewAWSProfileManager(ctx)
			if err != nil {
				return err
			}

			return manager.LoginSSO(args[0])
		},
	}

	return cmd
}

// newAWSProfileShowCmd creates the show command.
func newAWSProfileShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show [profile-name]",
		Short: "Show detailed information about a profile",
		Long:  "Display detailed configuration and status of an AWS profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			manager, err := NewAWSProfileManager(ctx)
			if err != nil {
				return err
			}

			profile, err := manager.GetProfile(args[0])
			if err != nil {
				return err
			}

			// Display profile information
			fmt.Printf("Profile: %s\n", profile.Name)
			fmt.Printf("Region: %s\n", profile.Region)
			fmt.Printf("Output Format: %s\n", profile.Output)

			if profile.SSOStartURL != "" {
				fmt.Printf("\nSSO Configuration:\n")
				fmt.Printf("  Start URL: %s\n", profile.SSOStartURL)
				fmt.Printf("  Role Name: %s\n", profile.SSORoleName)
				fmt.Printf("  Account ID: %s\n", profile.SSOAccountID)
				fmt.Printf("  SSO Region: %s\n", profile.SSORegion)
			}

			if profile.RoleArn != "" {
				fmt.Printf("\nAssumeRole Configuration:\n")
				fmt.Printf("  Role ARN: %s\n", profile.RoleArn)
				fmt.Printf("  Source Profile: %s\n", profile.SourceProfile)
				if profile.MFASerial != "" {
					fmt.Printf("  MFA Serial: %s\n", profile.MFASerial)
				}
				if profile.ExternalID != "" {
					fmt.Printf("  External ID: %s\n", profile.ExternalID)
				}
			}

			if profile.CredentialProcess != "" {
				fmt.Printf("\nCredential Process: %s\n", profile.CredentialProcess)
			}

			// Check credentials validity
			fmt.Printf("\nValidating credentials...\n")
			if err := manager.validateSSOCredentials(profile); err != nil {
				fmt.Printf("Status: ❌ Invalid or expired\n")
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Printf("Status: ✅ Valid\n")
			}

			return nil
		},
	}

	return cmd
}

// newAWSProfileValidateCmd creates the validate command.
func newAWSProfileValidateCmd() *cobra.Command {
	var all bool

	cmd := &cobra.Command{
		Use:   "validate [profile-name]",
		Short: "Validate AWS profile credentials",
		Long:  "Check if AWS profile credentials are valid and can be used",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			manager, err := NewAWSProfileManager(ctx)
			if err != nil {
				return err
			}

			var profiles []*AWSProfile
			switch {
			case all:
				profiles = manager.ListProfiles()
			case len(args) > 0:
				profile, err := manager.GetProfile(args[0])
				if err != nil {
					return err
				}
				profiles = []*AWSProfile{profile}
			default:
				// Validate current profile
				currentProfile := os.Getenv("AWS_PROFILE")
				if currentProfile == "" {
					currentProfile = "default"
				}
				profile, err := manager.GetProfile(currentProfile)
				if err != nil {
					return err
				}
				profiles = []*AWSProfile{profile}
			}

			// Validate each profile
			table := tablewriter.NewWriter(os.Stdout)
			table.Header("Profile", "Type", "Status", "Details")

			for _, profile := range profiles {
				profileType := "Standard"
				if profile.SSOStartURL != "" {
					profileType = "SSO"
				} else if profile.RoleArn != "" {
					profileType = "AssumeRole"
				}

				status := "✅ Valid"
				details := "Credentials are valid"

				if err := manager.validateSSOCredentials(profile); err != nil {
					status = "❌ Invalid"
					details = err.Error()
				}

				if err := table.Append(
					profile.Name,
					profileType,
					status,
					details,
				); err != nil {
					// Log error but continue processing
					fmt.Fprintf(os.Stderr, "Warning: failed to add profile to validation table: %v\n", err)
				}
			}

			table.Render() //nolint:errcheck,gosec // Table rendering errors are non-critical for CLI display
			return nil
		},
	}

	cmd.Flags().BoolVarP(&all, "all", "a", false, "Validate all profiles")

	return cmd
}
