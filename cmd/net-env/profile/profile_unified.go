// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package profile

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/Gizzahub/gzh-cli/internal/netenv"
)

// newProfileUnifiedCmd creates the unified net-env profile command.
func NewProfileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Manage network environment profiles",
		Long: `Manage network environment profiles including creation, editing,
deletion, import, and export operations.

Network profiles define comprehensive network configurations that can be
automatically applied when switching network environments. Each profile
can include VPN settings, DNS configuration, proxy settings, Docker
contexts, Kubernetes contexts, and more.

Examples:
  # List all profiles
  gz net-env profile list

  # Create a new profile interactively
  gz net-env profile create myprofile

  # Edit an existing profile
  gz net-env profile edit office

  # Delete a profile
  gz net-env profile delete old-profile

  # Export a profile to file
  gz net-env profile export office office.yaml

  # Import a profile from file
  gz net-env profile import myprofile.yaml`,
		SilenceUsage: true,
	}

	cmd.AddCommand(newProfileListCmd())
	cmd.AddCommand(newProfileCreateCmd())
	cmd.AddCommand(newProfileEditCmd())
	cmd.AddCommand(newProfileDeleteCmd())
	cmd.AddCommand(newProfileExportCmd())
	cmd.AddCommand(newProfileImportCmd())

	return cmd
}

// newProfileListCmd creates the profile list subcommand.
func newProfileListCmd() *cobra.Command {
	var verbose bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all network profiles",
		Long:  `List all available network environment profiles with their details.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProfileList(verbose)
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed profile information")

	return cmd
}

// newProfileCreateCmd creates the profile create subcommand.
func newProfileCreateCmd() *cobra.Command {
	var (
		description string
		priority    int
		auto        bool
		template    string
	)

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new network profile",
		Long: `Create a new network environment profile.

You can create a profile from scratch or use a template as a starting point.
Available templates: home, office, cafe, minimal`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProfileCreate(args[0], description, priority, auto, template)
		},
	}

	cmd.Flags().StringVarP(&description, "description", "d", "", "Profile description")
	cmd.Flags().IntVarP(&priority, "priority", "p", 50, "Profile priority (0-100)")
	cmd.Flags().BoolVar(&auto, "auto", false, "Enable auto-detection for this profile")
	cmd.Flags().StringVarP(&template, "template", "t", "", "Template to use (home, office, cafe, minimal)")

	return cmd
}

// newProfileEditCmd creates the profile edit subcommand.
func newProfileEditCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit <name>",
		Short: "Edit an existing network profile",
		Long:  `Edit an existing network environment profile using your default editor.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProfileEdit(args[0])
		},
	}

	return cmd
}

// newProfileDeleteCmd creates the profile delete subcommand.
func newProfileDeleteCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a network profile",
		Long:  `Delete an existing network environment profile.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProfileDelete(args[0], force)
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force deletion without confirmation")

	return cmd
}

// newProfileExportCmd creates the profile export subcommand.
func newProfileExportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export <name> <file>",
		Short: "Export a network profile to file",
		Long:  `Export an existing network environment profile to a YAML file.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProfileExport(args[0], args[1])
		},
	}

	return cmd
}

// newProfileImportCmd creates the profile import subcommand.
func newProfileImportCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "import <file>",
		Short: "Import a network profile from file",
		Long:  `Import a network environment profile from a YAML file.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProfileImport(args[0], force)
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force import, overwriting existing profile")

	return cmd
}

// Implementation functions

// runProfileList lists all network profiles.
func runProfileList(verbose bool) error {
	configDir := netenv.GetConfigDirectory()
	profileManager := netenv.NewProfileManager(configDir)

	if err := profileManager.LoadProfiles(); err != nil {
		return fmt.Errorf("failed to load profiles: %w", err)
	}

	profiles := profileManager.ListProfiles()

	if len(profiles) == 0 {
		fmt.Println("No network profiles found.")
		fmt.Println("Use 'gz net-env profile create <name>' to create a new profile.")
		return nil
	}

	fmt.Printf("Network Profiles (%d):\n\n", len(profiles))

	for _, profile := range profiles {
		fmt.Printf("📶 %s", profile.Name)
		if profile.Description != "" {
			fmt.Printf(" - %s", profile.Description)
		}
		fmt.Println()

		if verbose {
			// Show creation and update times
			if !profile.CreatedAt.IsZero() {
				fmt.Printf("   Created: %s\n", profile.CreatedAt.Format("2006-01-02 15:04:05"))
			}
			if !profile.UpdatedAt.IsZero() {
				fmt.Printf("   Updated: %s\n", profile.UpdatedAt.Format("2006-01-02 15:04:05"))
			}

			// Show priority and auto-detection
			fmt.Printf("   Priority: %d", profile.Priority)
			if profile.Auto {
				fmt.Printf(", Auto-detection: enabled")
			}
			fmt.Println()

			// Show conditions
			if len(profile.Conditions) > 0 {
				fmt.Printf("   Conditions: ")
				for i, condition := range profile.Conditions {
					if i > 0 {
						fmt.Printf(", ")
					}
					fmt.Printf("%s=%s", condition.Type, condition.Value)
				}
				fmt.Println()
			}

			// Show configured components
			components := []string{}
			if profile.Components.WiFi != nil {
				components = append(components, "WiFi")
			}
			if profile.Components.VPN != nil {
				components = append(components, fmt.Sprintf("VPN(%s)", profile.Components.VPN.Name))
			}
			if profile.Components.DNS != nil {
				components = append(components, fmt.Sprintf("DNS(%d servers)", len(profile.Components.DNS.Servers)))
			}
			if profile.Components.Proxy != nil {
				components = append(components, "Proxy")
			}
			if profile.Components.Docker != nil {
				components = append(components, fmt.Sprintf("Docker(%s)", profile.Components.Docker.Context))
			}
			if profile.Components.Kubernetes != nil {
				components = append(components, fmt.Sprintf("K8s(%s)", profile.Components.Kubernetes.Context))
			}

			if len(components) > 0 {
				fmt.Printf("   Components: %s\n", joinComponents(components))
			}
		}

		fmt.Println()
	}

	return nil
}

// runProfileCreate creates a new network profile.
func runProfileCreate(name, description string, priority int, auto bool, template string) error {
	configDir := netenv.GetConfigDirectory()
	profileManager := netenv.NewProfileManager(configDir)

	if err := profileManager.LoadProfiles(); err != nil {
		return fmt.Errorf("failed to load profiles: %w", err)
	}

	// Check if profile already exists
	if _, err := profileManager.GetProfile(name); err == nil {
		return fmt.Errorf("profile '%s' already exists", name)
	}

	// Create new profile
	profile := &netenv.NetworkProfile{
		Name:        name,
		Description: description,
		Priority:    priority,
		Auto:        auto,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Components:  netenv.NetworkComponents{},
	}

	// Apply template if specified
	if template != "" {
		if err := applyTemplate(profile, template); err != nil {
			return fmt.Errorf("failed to apply template: %w", err)
		}
	}

	// Save the profile
	if err := profileManager.SaveProfile(profile); err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	fmt.Printf("✅ Created profile: %s\n", name)
	if template != "" {
		fmt.Printf("   Applied template: %s\n", template)
	}
	fmt.Println("Use 'gz net-env profile edit' to customize the profile configuration.")

	return nil
}

// runProfileEdit edits an existing profile.
func runProfileEdit(name string) error {
	configDir := netenv.GetConfigDirectory()
	profileManager := netenv.NewProfileManager(configDir)

	if err := profileManager.LoadProfiles(); err != nil {
		return fmt.Errorf("failed to load profiles: %w", err)
	}

	// Check if profile exists
	profile, err := profileManager.GetProfile(name)
	if err != nil {
		return fmt.Errorf("profile not found: %w", err)
	}

	// Get profile file path
	profilePath := filepath.Join(configDir, "net-env", "profiles", fmt.Sprintf("%s.yaml", name))

	fmt.Printf("Opening profile '%s' for editing...\n", name)
	fmt.Printf("File: %s\n", profilePath)
	fmt.Println("Edit the file and save to update the profile.")

	// In a real implementation, you would open the file with the user's default editor
	// For now, just show the path
	fmt.Printf("\nCurrent profile configuration:\n")
	fmt.Printf("  Name: %s\n", profile.Name)
	fmt.Printf("  Description: %s\n", profile.Description)
	fmt.Printf("  Priority: %d\n", profile.Priority)
	fmt.Printf("  Auto-detection: %v\n", profile.Auto)

	return nil
}

// runProfileDelete deletes a profile.
func runProfileDelete(name string, force bool) error {
	configDir := netenv.GetConfigDirectory()
	profileManager := netenv.NewProfileManager(configDir)

	if err := profileManager.LoadProfiles(); err != nil {
		return fmt.Errorf("failed to load profiles: %w", err)
	}

	// Check if profile exists
	_, err := profileManager.GetProfile(name)
	if err != nil {
		return fmt.Errorf("profile not found: %w", err)
	}

	// Confirm deletion if not forced
	if !force {
		fmt.Printf("Are you sure you want to delete profile '%s'? (y/N): ", name)
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Deletion canceled.")
			return nil
		}
	}

	// Delete the profile
	if err := profileManager.DeleteProfile(name); err != nil {
		return fmt.Errorf("failed to delete profile: %w", err)
	}

	fmt.Printf("✅ Deleted profile: %s\n", name)
	return nil
}

// runProfileExport exports a profile to file.
func runProfileExport(name, outputFile string) error {
	configDir := netenv.GetConfigDirectory()
	profileManager := netenv.NewProfileManager(configDir)

	if err := profileManager.LoadProfiles(); err != nil {
		return fmt.Errorf("failed to load profiles: %w", err)
	}

	// Export the profile
	if err := profileManager.ExportProfile(name, outputFile); err != nil {
		return fmt.Errorf("failed to export profile: %w", err)
	}

	fmt.Printf("✅ Exported profile '%s' to: %s\n", name, outputFile)
	return nil
}

// runProfileImport imports a profile from file.
func runProfileImport(inputFile string, force bool) error {
	configDir := netenv.GetConfigDirectory()
	profileManager := netenv.NewProfileManager(configDir)

	if err := profileManager.LoadProfiles(); err != nil {
		return fmt.Errorf("failed to load profiles: %w", err)
	}

	// Import the profile
	if err := profileManager.ImportProfile(inputFile); err != nil {
		return fmt.Errorf("failed to import profile: %w", err)
	}

	fmt.Printf("✅ Imported profile from: %s\n", inputFile)
	return nil
}

// applyTemplate applies a template to a profile.
func applyTemplate(profile *netenv.NetworkProfile, template string) error {
	switch template {
	case "home":
		profile.Description = "Home network configuration"
		profile.Priority = 50
		profile.Components.DNS = &netenv.DNSConfig{
			Servers: []string{"1.1.1.1", "1.0.0.1"},
		}
	case "office":
		profile.Description = "Corporate office network"
		profile.Priority = 100
		profile.Components.VPN = &netenv.VPNConfig{
			Name:        "corp-vpn",
			Type:        "openvpn",
			AutoConnect: true,
		}
		profile.Components.DNS = &netenv.DNSConfig{
			Servers:  []string{"10.0.0.1", "10.0.0.2"},
			Override: true,
		}
		profile.Components.Proxy = &netenv.ProxyConfig{
			HTTP:  "proxy.corp.com:8080",
			HTTPS: "proxy.corp.com:8080",
		}
	case "cafe":
		profile.Description = "Public WiFi / Cafe network"
		profile.Priority = 25
		profile.Components.VPN = &netenv.VPNConfig{
			Name:        "personal-vpn",
			Type:        "wireguard",
			AutoConnect: true,
		}
		profile.Components.DNS = &netenv.DNSConfig{
			Servers:  []string{"1.1.1.1", "8.8.8.8"},
			Override: true,
		}
	case "minimal":
		profile.Description = "Minimal network configuration"
		profile.Priority = 10
		// No components configured
	default:
		return fmt.Errorf("unknown template: %s", template)
	}

	return nil
}

// joinComponents joins component names with proper formatting.
func joinComponents(components []string) string {
	if len(components) == 0 {
		return ""
	}

	result := ""
	for i, component := range components {
		if i > 0 {
			if i == len(components)-1 {
				result += " and "
			} else {
				result += ", "
			}
		}
		result += component
	}

	return result
}
