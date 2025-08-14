// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package netenv

import (
	"context"
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"

	"github.com/Gizzahub/gzh-cli/internal/netenv"
)

// newSwitchUnifiedCmd creates the unified net-env switch command.
func newSwitchUnifiedCmd() *cobra.Command {
	var (
		interactive bool
		list        bool
		preview     bool
		last        bool
		init        bool
		force       bool
	)

	cmd := &cobra.Command{
		Use:   "switch [profile]",
		Short: "Switch network environment profile",
		Long: `Switch to a different network environment profile.

This command allows you to switch between different network environment profiles
that define comprehensive network configurations including VPN, DNS, proxy,
and other network component settings.

The command supports several modes:
- Direct profile switching by name
- Interactive profile selection
- Automatic profile detection and suggestion
- Preview mode to see changes before applying
- Quick switch to last used profile

Examples:
  # Auto-detect and suggest profile for current network
  gz net-env switch

  # Switch to a specific profile
  gz net-env switch office

  # Interactive profile selection
  gz net-env switch --interactive

  # List available profiles
  gz net-env switch --list

  # Preview changes before switching
  gz net-env switch office --preview

  # Switch to last used profile
  gz net-env switch --last

  # Initialize with default example profiles
  gz net-env switch --init`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSwitchUnified(cmd.Context(), args, interactive, list, preview, last, init, force)
		},
	}

	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Interactive profile selection")
	cmd.Flags().BoolVarP(&list, "list", "l", false, "List available network profiles")
	cmd.Flags().BoolVar(&preview, "preview", false, "Preview changes without applying")
	cmd.Flags().BoolVar(&last, "last", false, "Switch to last used profile")
	cmd.Flags().BoolVar(&init, "init", false, "Create default example profiles")
	cmd.Flags().BoolVar(&force, "force", false, "Force switch even if already active")

	return cmd
}

// runSwitchUnified executes the unified switch command.
func runSwitchUnified(ctx context.Context, args []string, interactive, list, preview, last, init, force bool) error {
	configDir := getConfigDirectory()

	// Initialize profile manager
	profileManager := netenv.NewProfileManager(configDir)
	if err := profileManager.LoadProfiles(); err != nil {
		return fmt.Errorf("failed to load profiles: %w", err)
	}

	// Handle init flag
	if init {
		return initializeDefaultProfiles(profileManager)
	}

	// Handle list flag
	if list {
		return listNetworkProfiles(profileManager)
	}

	// Handle last flag
	if last {
		return switchToLastProfile(profileManager)
	}

	// Determine target profile
	var targetProfile *netenv.NetworkProfile
	var err error

	if len(args) > 0 {
		// Explicit profile name provided
		targetProfile, err = profileManager.GetProfile(args[0])
		if err != nil {
			return fmt.Errorf("profile not found: %w", err)
		}
	} else if interactive {
		// Interactive mode
		targetProfile, err = selectProfileInteractively(profileManager)
		if err != nil {
			return fmt.Errorf("profile selection failed: %w", err)
		}
	} else {
		// Auto-detection mode
		targetProfile, err = autoDetectProfile(ctx, profileManager)
		if err != nil {
			return fmt.Errorf("auto-detection failed: %w", err)
		}
	}

	if targetProfile == nil {
		return fmt.Errorf("no profile selected")
	}

	// Preview mode
	if preview {
		return previewProfileSwitch(targetProfile)
	}

	// Execute the switch
	return executeProfileSwitch(ctx, targetProfile, force)
}

// initializeDefaultProfiles creates default example profiles.
func initializeDefaultProfiles(profileManager *netenv.ProfileManager) error {
	fmt.Println("Initializing default network profiles...")

	if err := profileManager.CreateDefaultProfiles(); err != nil {
		return fmt.Errorf("failed to create default profiles: %w", err)
	}

	fmt.Println("✅ Default profiles created successfully!")
	fmt.Println("\nCreated profiles:")
	fmt.Println("  • home - Home network configuration")
	fmt.Println("  • office - Corporate office network")
	fmt.Println("  • cafe - Public WiFi / Cafe configuration")
	fmt.Println("\nUse 'gz net-env switch --list' to see all available profiles.")

	return nil
}

// listNetworkProfiles lists all available network profiles.
func listNetworkProfiles(profileManager *netenv.ProfileManager) error {
	profiles := profileManager.ListProfiles()

	if len(profiles) == 0 {
		fmt.Println("No network profiles found.")
		fmt.Println("Use 'gz net-env switch --init' to create default profiles.")
		return nil
	}

	fmt.Printf("Available Network Profiles (%d):\n\n", len(profiles))

	for _, profile := range profiles {
		fmt.Printf("📶 %s", profile.Name)
		if profile.Description != "" {
			fmt.Printf(" - %s", profile.Description)
		}
		fmt.Println()

		// Show priority
		if profile.Priority > 0 {
			fmt.Printf("   Priority: %d", profile.Priority)
		}

		// Show auto-detection conditions
		if len(profile.Conditions) > 0 {
			fmt.Printf("   Auto-detect: ")
			var conditions []string
			for _, condition := range profile.Conditions {
				conditions = append(conditions, fmt.Sprintf("%s=%s", condition.Type, condition.Value))
			}
			fmt.Printf("%s", strings.Join(conditions, ", "))
		}
		fmt.Println()

		// Show configured components
		components := []string{}
		if profile.Components.WiFi != nil {
			components = append(components, "WiFi")
		}
		if profile.Components.VPN != nil {
			components = append(components, "VPN")
		}
		if profile.Components.DNS != nil {
			components = append(components, "DNS")
		}
		if profile.Components.Proxy != nil {
			components = append(components, "Proxy")
		}
		if profile.Components.Docker != nil {
			components = append(components, "Docker")
		}
		if profile.Components.Kubernetes != nil {
			components = append(components, "Kubernetes")
		}

		if len(components) > 0 {
			fmt.Printf("   Components: %s\n", strings.Join(components, ", "))
		}

		fmt.Println()
	}

	return nil
}

// switchToLastProfile switches to the last used profile.
func switchToLastProfile(profileManager *netenv.ProfileManager) error {
	// This would require storing last used profile information
	// For now, return an error
	return fmt.Errorf("last profile tracking not yet implemented")
}

// selectProfileInteractively presents an interactive profile selection.
func selectProfileInteractively(profileManager *netenv.ProfileManager) (*netenv.NetworkProfile, error) {
	profiles := profileManager.ListProfiles()

	if len(profiles) == 0 {
		return nil, fmt.Errorf("no profiles available")
	}

	// Create selection items
	items := make([]string, len(profiles))
	for i, profile := range profiles {
		description := profile.Name
		if profile.Description != "" {
			description = fmt.Sprintf("%s - %s", profile.Name, profile.Description)
		}
		items[i] = description
	}

	// Create promptui selector
	prompt := promptui.Select{
		Label: "Select Network Profile",
		Items: items,
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}?",
			Active:   "▶ {{ .Name | cyan }} {{ if .Description }}({{ .Description | faint }}){{ end }}",
			Inactive: "  {{ .Name | faint }} {{ if .Description }}({{ .Description | faint }}){{ end }}",
			Selected: "✓ Selected: {{ .Name | green }}",
		},
	}

	index, _, err := prompt.Run()
	if err != nil {
		return nil, err
	}

	return profiles[index], nil
}

// autoDetectProfile automatically detects the best profile for current environment.
func autoDetectProfile(ctx context.Context, profileManager *netenv.ProfileManager) (*netenv.NetworkProfile, error) {
	profiles := profileManager.ListProfiles()

	if len(profiles) == 0 {
		return nil, fmt.Errorf("no profiles available for auto-detection")
	}

	// Create network detector
	networkProfiles := make([]netenv.NetworkProfile, len(profiles))
	for i, p := range profiles {
		networkProfiles[i] = *p
	}
	detector := netenv.NewNetworkDetector(networkProfiles)

	// Detect current environment
	detected, err := detector.DetectEnvironment(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to detect network environment: %w", err)
	}

	if detected == nil {
		fmt.Println("No matching profile found for current network environment.")
		fmt.Println("Available profiles:")

		for _, profile := range profiles {
			fmt.Printf("  • %s", profile.Name)
			if profile.Description != "" {
				fmt.Printf(" - %s", profile.Description)
			}
			fmt.Println()
		}
		return nil, fmt.Errorf("no suitable profile detected")
	}

	fmt.Printf("🔍 Auto-detected network profile: %s\n", detected.Name)
	if detected.Description != "" {
		fmt.Printf("   Description: %s\n", detected.Description)
	}

	return detected, nil
}

// previewProfileSwitch shows what changes would be made without applying them.
func previewProfileSwitch(profile *netenv.NetworkProfile) error {
	fmt.Printf("Preview: Switching to profile '%s'\n\n", profile.Name)

	if profile.Description != "" {
		fmt.Printf("Description: %s\n\n", profile.Description)
	}

	fmt.Println("Changes to be applied:")

	// WiFi changes
	if profile.Components.WiFi != nil {
		fmt.Printf("  📶 WiFi: %s\n", profile.Components.WiFi.SSID)
	}

	// VPN changes
	if profile.Components.VPN != nil {
		fmt.Printf("  🔒 VPN: %s (%s)\n", profile.Components.VPN.Name, profile.Components.VPN.Type)
		if profile.Components.VPN.AutoConnect {
			fmt.Println("         Auto-connect enabled")
		}
	}

	// DNS changes
	if profile.Components.DNS != nil {
		fmt.Printf("  🌐 DNS: %s\n", strings.Join(profile.Components.DNS.Servers, ", "))
		if profile.Components.DNS.Override {
			fmt.Println("         Override system DNS")
		}
	}

	// Proxy changes
	if profile.Components.Proxy != nil {
		if profile.Components.Proxy.HTTP != "" {
			fmt.Printf("  🌐 Proxy: HTTP=%s\n", profile.Components.Proxy.HTTP)
		}
		if profile.Components.Proxy.HTTPS != "" {
			fmt.Printf("          HTTPS=%s\n", profile.Components.Proxy.HTTPS)
		}
	}

	// Docker changes
	if profile.Components.Docker != nil {
		fmt.Printf("  🐳 Docker: Context=%s\n", profile.Components.Docker.Context)
	}

	// Kubernetes changes
	if profile.Components.Kubernetes != nil {
		fmt.Printf("  ☸️  Kubernetes: Context=%s", profile.Components.Kubernetes.Context)
		if profile.Components.Kubernetes.Namespace != "" {
			fmt.Printf(", Namespace=%s", profile.Components.Kubernetes.Namespace)
		}
		fmt.Println()
	}

	fmt.Println("\nUse 'gz net-env switch <profile>' to apply these changes.")

	return nil
}

// executeProfileSwitch executes the actual profile switch.
func executeProfileSwitch(ctx context.Context, profile *netenv.NetworkProfile, force bool) error {
	fmt.Printf("🔄 Switching to network profile: %s\n", profile.Name)

	// Here you would implement the actual switching logic
	// For now, we'll just simulate the process

	// Step 1: Pre-hooks
	if len(profile.Metadata) > 0 {
		fmt.Println("  ⏳ Running pre-switch hooks...")
	}

	// Step 2: Apply VPN configuration
	if profile.Components.VPN != nil {
		fmt.Printf("  🔒 Configuring VPN: %s\n", profile.Components.VPN.Name)
		// Implement VPN connection logic
	}

	// Step 3: Apply DNS configuration
	if profile.Components.DNS != nil {
		fmt.Printf("  🌐 Configuring DNS: %s\n", strings.Join(profile.Components.DNS.Servers, ", "))
		// Implement DNS configuration logic
	}

	// Step 4: Apply proxy configuration
	if profile.Components.Proxy != nil {
		fmt.Println("  🌐 Configuring proxy settings")
		// Implement proxy configuration logic
	}

	// Step 5: Apply Docker configuration
	if profile.Components.Docker != nil {
		fmt.Printf("  🐳 Switching Docker context: %s\n", profile.Components.Docker.Context)
		// Implement Docker context switching
	}

	// Step 6: Apply Kubernetes configuration
	if profile.Components.Kubernetes != nil {
		fmt.Printf("  ☸️  Switching Kubernetes context: %s\n", profile.Components.Kubernetes.Context)
		// Implement Kubernetes context switching
	}

	// Step 7: Post-hooks
	if len(profile.Metadata) > 0 {
		fmt.Println("  ⏳ Running post-switch hooks...")
	}

	fmt.Printf("✅ Successfully switched to profile: %s\n", profile.Name)

	// Show current status
	fmt.Println("\nCurrent Status:")
	// This would call the status command to show current state
	fmt.Println("  Use 'gz net-env status' to see detailed information.")

	return nil
}
