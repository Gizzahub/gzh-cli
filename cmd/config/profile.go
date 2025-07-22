// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	configpkg "github.com/gizzahub/gzh-manager-go/pkg/config"
	"github.com/spf13/cobra"
)

// newProfileCmd creates the config profile subcommand.
func newProfileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Manage configuration profiles",
		Long: `Manage configuration profiles for different environments.

Profiles allow you to maintain different configurations for different environments
(development, staging, production) and switch between them easily.

Configuration profiles are stored as:
- gzh.dev.yaml (development profile)
- gzh.staging.yaml (staging profile)
- gzh.prod.yaml (production profile)

Examples:
  gz config profile list                # List available profiles
  gz config profile create dev         # Create development profile
  gz config profile use prod           # Use production profile
  gz config profile current           # Show current profile`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	// Add subcommands
	cmd.AddCommand(newProfileListCmd())
	cmd.AddCommand(newProfileCreateCmd())
	cmd.AddCommand(newProfileUseCmd())
	cmd.AddCommand(newProfileCurrentCmd())
	cmd.AddCommand(newProfileDeleteCmd())

	return cmd
}

// newProfileListCmd lists available profiles.
func newProfileListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available configuration profiles",
		RunE: func(_ *cobra.Command, _ []string) error {
			return listProfiles()
		},
	}
}

// newProfileCreateCmd creates a new profile.
func newProfileCreateCmd() *cobra.Command {
	var (
		fromProfile string
		interactive bool
	)

	cmd := &cobra.Command{
		Use:   "create <profile-name>",
		Short: "Create a new configuration profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			profileName := args[0]
			return createProfile(profileName, fromProfile, interactive)
		},
	}

	cmd.Flags().StringVar(&fromProfile, "from", "", "Copy from existing profile")
	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Create profile interactively")

	return cmd
}

// newProfileUseCmd switches to a profile.
func newProfileUseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "use <profile-name>",
		Short: "Switch to a configuration profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			profileName := args[0]
			return useProfile(profileName)
		},
	}
}

// newProfileCurrentCmd shows current profile.
func newProfileCurrentCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "current",
		Short: "Show current configuration profile",
		RunE: func(_ *cobra.Command, _ []string) error {
			return showCurrentProfile()
		},
	}
}

// newProfileDeleteCmd deletes a profile.
func newProfileDeleteCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <profile-name>",
		Short: "Delete a configuration profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := args[0]
			return deleteProfile(profileName, force)
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Force delete without confirmation")

	return cmd
}

// Profile management functions

// listProfiles lists all available configuration profiles.
func listProfiles() error {
	profiles, err := getAvailableProfiles()
	if err != nil {
		return fmt.Errorf("failed to get profiles: %w", err)
	}

	currentProfile, _ := getCurrentProfile()

	fmt.Println("Available configuration profiles:")
	fmt.Println()

	if len(profiles) == 0 {
		fmt.Println("No profiles found. Create one with: gz config profile create <name>")
		return nil
	}

	for _, profile := range profiles {
		marker := "  "
		if profile == currentProfile {
			marker = "* "
		}

		fmt.Printf("%s%s", marker, profile)

		// Try to load profile info
		if info := getProfileInfo(profile); info != "" {
			fmt.Printf(" %s", info)
		}

		fmt.Println()
	}

	fmt.Println()

	if currentProfile != "" {
		fmt.Printf("Current profile: %s\n", currentProfile)
	} else {
		fmt.Println("No active profile (using default gzh.yaml)")
	}

	return nil
}

// createProfile creates a new configuration profile.
func createProfile(profileName, fromProfile string, interactive bool) error {
	// Validate profile name
	if err := validateProfileName(profileName); err != nil {
		return err
	}

	profileFile := getProfilePath(profileName)

	// Check if profile already exists
	if _, err := os.Stat(profileFile); err == nil {
		return fmt.Errorf("profile '%s' already exists", profileName)
	}

	var content string

	if fromProfile != "" { //nolint:gocritic // Simple if-else chain, switch would not improve readability
		// Copy from existing profile
		sourceFile := getProfilePath(fromProfile)
		if _, err := os.Stat(sourceFile); os.IsNotExist(err) {
			return fmt.Errorf("source profile '%s' does not exist", fromProfile)
		}

		sourceContent, err := os.ReadFile(sourceFile)
		if err != nil {
			return fmt.Errorf("failed to read source profile: %w", err)
		}

		content = string(sourceContent)
	} else if interactive {
		// Create interactively
		return createInteractiveConfig(profileFile)
	} else {
		// Create minimal profile
		content = generateProfileTemplate(profileName)
	}

	// Write profile file
	if err := os.WriteFile(profileFile, []byte(content), 0o644); err != nil {
		return fmt.Errorf("failed to create profile: %w", err)
	}

	fmt.Printf("✓ Profile '%s' created: %s\n", profileName, profileFile)

	return nil
}

// useProfile switches to a configuration profile.
func useProfile(profileName string) error {
	// Validate profile exists
	profileFile := getProfilePath(profileName)
	if _, err := os.Stat(profileFile); os.IsNotExist(err) {
		return fmt.Errorf("profile '%s' does not exist", profileName)
	}

	// Validate profile before switching
	if _, err := configpkg.ParseYAMLFile(profileFile); err != nil {
		return fmt.Errorf("profile '%s' is invalid: %w", profileName, err)
	}

	// Create symlink to active profile
	activeLink := "gzh.yaml"

	// Remove existing link/file
	if _, err := os.Lstat(activeLink); err == nil {
		if err := os.Remove(activeLink); err != nil {
			return fmt.Errorf("failed to remove existing config: %w", err)
		}
	}

	// Create symlink to profile
	if err := os.Symlink(profileFile, activeLink); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	fmt.Printf("✓ Switched to profile: %s\n", profileName)

	return nil
}

// showCurrentProfile shows the current active profile.
func showCurrentProfile() error {
	profile, err := getCurrentProfile()
	if err != nil {
		return fmt.Errorf("failed to get current profile: %w", err)
	}

	if profile == "" {
		fmt.Println("No active profile (using default gzh.yaml)")
	} else {
		fmt.Printf("Current profile: %s\n", profile)

		// Show profile info
		if info := getProfileInfo(profile); info != "" {
			fmt.Printf("Configuration: %s\n", info)
		}
	}

	return nil
}

// deleteProfile deletes a configuration profile.
func deleteProfile(profileName string, force bool) error {
	// Validate profile exists
	profileFile := getProfilePath(profileName)
	if _, err := os.Stat(profileFile); os.IsNotExist(err) {
		return fmt.Errorf("profile '%s' does not exist", profileName)
	}

	// Check if it's the current profile
	currentProfile, _ := getCurrentProfile()
	if profileName == currentProfile {
		return fmt.Errorf("cannot delete active profile '%s'. Switch to another profile first", profileName)
	}

	// Confirm deletion
	if !force {
		fmt.Printf("Delete profile '%s'? (y/N): ", profileName)

		var response string
		if _, err := fmt.Scanln(&response); err != nil {
			// Handle scan error but continue with empty response (defaults to "no")
			response = ""
		}

		if !strings.EqualFold(response, "y") && !strings.EqualFold(response, "yes") {
			fmt.Println("Profile deletion canceled")
			return nil
		}
	}

	// Delete profile file
	if err := os.Remove(profileFile); err != nil {
		return fmt.Errorf("failed to delete profile: %w", err)
	}

	fmt.Printf("✓ Profile '%s' deleted\n", profileName)

	return nil
}

// Helper functions

// getAvailableProfiles returns list of available profiles.
func getAvailableProfiles() ([]string, error) {
	var profiles []string

	// Look for gzh.*.yaml files
	matches, err := filepath.Glob("gzh.*.yaml")
	if err != nil {
		return nil, err
	}

	for _, match := range matches {
		// Extract profile name from filename
		// gzh.dev.yaml -> dev
		base := filepath.Base(match)
		if strings.HasPrefix(base, "gzh.") && strings.HasSuffix(base, ".yaml") {
			profileName := strings.TrimPrefix(base, "gzh.")
			profileName = strings.TrimSuffix(profileName, ".yaml")
			profiles = append(profiles, profileName)
		}
	}

	return profiles, nil
}

// getCurrentProfile returns the name of the current active profile.
func getCurrentProfile() (string, error) {
	// Check if gzh.yaml is a symlink
	linkTarget, err := os.Readlink("gzh.yaml")
	if err != nil {
		// Not a symlink or doesn't exist
		return "", err
	}

	// Extract profile name from link target
	// gzh.dev.yaml -> dev
	base := filepath.Base(linkTarget)
	if strings.HasPrefix(base, "gzh.") && strings.HasSuffix(base, ".yaml") {
		profileName := strings.TrimPrefix(base, "gzh.")
		profileName = strings.TrimSuffix(profileName, ".yaml")

		return profileName, nil
	}

	return "", nil
}

// getProfilePath returns the file path for a profile.
func getProfilePath(profileName string) string {
	return fmt.Sprintf("gzh.%s.yaml", profileName)
}

// validateProfileName validates a profile name.
func validateProfileName(profileName string) error {
	if profileName == "" {
		return fmt.Errorf("profile name cannot be empty")
	}

	// Check for invalid characters
	if strings.ContainsAny(profileName, "/\\:*?\"<>|") {
		return fmt.Errorf("profile name contains invalid characters")
	}

	// Check for reserved names
	reserved := []string{"yaml", "yml"}
	for _, r := range reserved {
		if profileName == r {
			return fmt.Errorf("'%s' is a reserved profile name", profileName)
		}
	}

	return nil
}

// getProfileInfo returns basic info about a profile.
func getProfileInfo(profileName string) string {
	profileFile := getProfilePath(profileName)

	config, err := configpkg.ParseYAMLFile(profileFile)
	if err != nil {
		return "(invalid)"
	}

	providerCount := len(config.Providers)

	targetCount := 0
	for _, provider := range config.Providers {
		targetCount += len(provider.Orgs) + len(provider.Groups)
	}

	return fmt.Sprintf("(%d providers, %d targets)", providerCount, targetCount)
}

// generateProfileTemplate generates a basic profile template.
func generateProfileTemplate(profileName string) string {
	template := fmt.Sprintf(`# gzh.yaml - %s profile
# Generated by: gz config profile create %s
# Edit this file to customize your %s configuration

version: "1.0.0"
default_provider: github

providers:
  github:
    # Set your GitHub token for %s environment
    token: "${GITHUB_TOKEN_%s}"
    orgs:
      - name: "your-org-name"
        visibility: "all"
        clone_dir: "${HOME}/repos/%s/github"
        strategy: "reset"

# Profile-specific notes:
# - Use environment-specific tokens (GITHUB_TOKEN_%s)
# - Adjust clone directories for environment separation
# - Consider different strategies per environment (reset for prod, pull for dev)
`, strings.ToUpper(profileName[:1])+profileName[1:], profileName, profileName, profileName, strings.ToUpper(profileName), profileName, strings.ToUpper(profileName))

	return template
}
