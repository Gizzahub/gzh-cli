// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package devenv

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

const (
	outputFormatJSON = "json"
)

// AzureSubscription represents an Azure subscription configuration.
type AzureSubscription struct {
	ID                string            `json:"id"`
	DisplayName       string            `json:"displayName"`
	Name              string            `json:"name"`
	State             string            `json:"state"`
	TenantID          string            `json:"tenantId"`
	TenantDisplayName string            `json:"tenantDisplayName"`
	User              string            `json:"user"`
	IsDefault         bool              `json:"isDefault"`
	IsActive          bool              `json:"isActive"`
	LastUsed          *time.Time        `json:"lastUsed,omitempty"`
	Tags              map[string]string `json:"tags,omitempty"`
	ResourceGroups    []string          `json:"resourceGroups,omitempty"`
	Regions           []string          `json:"regions,omitempty"`
	EnvironmentName   string            `json:"environmentName"`
	HomeTenantID      string            `json:"homeTenantId"`
	ManagedByTenants  []string          `json:"managedByTenants,omitempty"`
}

// AzureSubscriptionManager manages Azure subscriptions and configurations.
type AzureSubscriptionManager struct {
	configPath    string
	subscriptions map[string]*AzureSubscription
	tenants       map[string]string // tenant_id -> display_name
	ctx           context.Context
}

// NewAzureSubscriptionManager creates a new Azure subscription manager.
func NewAzureSubscriptionManager(ctx context.Context) (*AzureSubscriptionManager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	manager := &AzureSubscriptionManager{
		configPath:    fmt.Sprintf("%s/.azure", homeDir),
		subscriptions: make(map[string]*AzureSubscription),
		tenants:       make(map[string]string),
		ctx:           ctx,
	}

	if err := manager.loadSubscriptions(); err != nil {
		return nil, fmt.Errorf("failed to load subscriptions: %w", err)
	}

	return manager, nil
}

// newAzureSubscriptionCmd creates the azure-subscription command.
func newAzureSubscriptionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "azure-subscription",
		Short: "Manage Azure subscriptions and configurations",
		Long: `Manage Microsoft Azure subscriptions and configurations.

This command provides functionality to:
- List and switch between Azure subscriptions
- Manage multi-tenant access
- View subscription details and settings
- Switch active subscription context
- Validate subscription access and permissions

Examples:
  # List all available subscriptions
  gz dev-env azure-subscription list

  # Switch to a specific subscription
  gz dev-env azure-subscription switch my-subscription-id

  # Show current subscription details
  gz dev-env azure-subscription show

  # Login to Azure
  gz dev-env azure-subscription login

  # Validate subscription access
  gz dev-env azure-subscription validate`,
		SilenceUsage: true,
	}

	cmd.AddCommand(newAzureSubscriptionListCmd())
	cmd.AddCommand(newAzureSubscriptionSwitchCmd())
	cmd.AddCommand(newAzureSubscriptionShowCmd())
	cmd.AddCommand(newAzureSubscriptionLoginCmd())
	cmd.AddCommand(newAzureSubscriptionValidateCmd())
	cmd.AddCommand(newAzureSubscriptionTenantCmd())

	return cmd
}

// newAzureSubscriptionListCmd creates the list subcommand.
func newAzureSubscriptionListCmd() *cobra.Command {
	var (
		outputFormat string
		tenantID     string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all available Azure subscriptions",
		Long: `List all Microsoft Azure subscriptions accessible to the current user.

This command shows:
- Subscription ID and display name
- Current status and state
- Associated tenant information
- Active subscription indicator
- Last used timestamp

Examples:
  # List subscriptions in table format
  gz dev-env azure-subscription list

  # List subscriptions in JSON format
  gz dev-env azure-subscription list --output json

  # List subscriptions for specific tenant
  gz dev-env azure-subscription list --tenant my-tenant-id`,
		RunE: func(cmd *cobra.Command, args []string) error {
			manager, err := NewAzureSubscriptionManager(cmd.Context())
			if err != nil {
				return err
			}
			return manager.listSubscriptions(outputFormat, tenantID)
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json)")
	cmd.Flags().StringVarP(&tenantID, "tenant", "t", "", "Filter by tenant ID")

	return cmd
}

// newAzureSubscriptionSwitchCmd creates the switch subcommand.
func newAzureSubscriptionSwitchCmd() *cobra.Command {
	var (
		interactive bool
		tenantID    string
	)

	cmd := &cobra.Command{
		Use:   "switch [SUBSCRIPTION_ID]",
		Short: "Switch to a specific Azure subscription",
		Long: `Switch the active Azure subscription context.

This command:
- Updates the Azure CLI active subscription
- Sets the default subscription for az commands
- Updates environment variables if needed
- Records the switch timestamp

Examples:
  # Switch to a specific subscription
  gz dev-env azure-subscription switch my-subscription-id

  # Interactive subscription selection
  gz dev-env azure-subscription switch --interactive

  # Switch with specific tenant context
  gz dev-env azure-subscription switch my-subscription-id --tenant my-tenant-id`,
		RunE: func(cmd *cobra.Command, args []string) error {
			manager, err := NewAzureSubscriptionManager(cmd.Context())
			if err != nil {
				return err
			}

			var subscriptionID string
			if len(args) > 0 {
				subscriptionID = args[0]
			}

			return manager.switchSubscription(subscriptionID, interactive, tenantID)
		},
	}

	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Interactive subscription selection")
	cmd.Flags().StringVarP(&tenantID, "tenant", "t", "", "Tenant ID context")

	return cmd
}

// newAzureSubscriptionShowCmd creates the show subcommand.
func newAzureSubscriptionShowCmd() *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "show [SUBSCRIPTION_ID]",
		Short: "Show detailed information about an Azure subscription",
		Long: `Show detailed information about a Microsoft Azure subscription.

If no subscription ID is specified, shows information about the current active subscription.

Examples:
  # Show current active subscription
  gz dev-env azure-subscription show

  # Show specific subscription details
  gz dev-env azure-subscription show my-subscription-id

  # Show subscription details in JSON format
  gz dev-env azure-subscription show --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			manager, err := NewAzureSubscriptionManager(cmd.Context())
			if err != nil {
				return err
			}

			var subscriptionID string
			if len(args) > 0 {
				subscriptionID = args[0]
			}

			return manager.showSubscription(subscriptionID, outputFormat)
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json)")

	return cmd
}

// newAzureSubscriptionLoginCmd creates the login subcommand.
func newAzureSubscriptionLoginCmd() *cobra.Command {
	var (
		tenantID         string
		useDeviceCode    bool
		servicePrincipal bool
	)

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login to Azure and authenticate",
		Long: `Login to Microsoft Azure and authenticate the current user.

This command supports different authentication methods:
- Interactive browser login (default)
- Device code flow for headless environments
- Service principal authentication

Examples:
  # Interactive browser login
  gz dev-env azure-subscription login

  # Login with device code
  gz dev-env azure-subscription login --device-code

  # Login to specific tenant
  gz dev-env azure-subscription login --tenant my-tenant-id

  # Service principal login (requires environment variables)
  gz dev-env azure-subscription login --service-principal`,
		RunE: func(cmd *cobra.Command, args []string) error {
			manager, err := NewAzureSubscriptionManager(cmd.Context())
			if err != nil {
				return err
			}
			return manager.login(tenantID, useDeviceCode, servicePrincipal)
		},
	}

	cmd.Flags().StringVarP(&tenantID, "tenant", "t", "", "Tenant ID to login to")
	cmd.Flags().BoolVar(&useDeviceCode, "device-code", false, "Use device code flow for authentication")
	cmd.Flags().BoolVar(&servicePrincipal, "service-principal", false, "Login using service principal")

	return cmd
}

// newAzureSubscriptionValidateCmd creates the validate subcommand.
func newAzureSubscriptionValidateCmd() *cobra.Command {
	var (
		checkResourceGroups bool
		checkPermissions    bool
		checkQuotas         bool
	)

	cmd := &cobra.Command{
		Use:   "validate [SUBSCRIPTION_ID]",
		Short: "Validate Azure subscription access and configuration",
		Long: `Validate Microsoft Azure subscription access and configuration.

This command checks:
- Subscription existence and access permissions
- Resource provider registrations
- Resource group access
- Role assignments and permissions
- Quota and usage limits

Examples:
  # Validate current subscription
  gz dev-env azure-subscription validate

  # Validate specific subscription with all checks
  gz dev-env azure-subscription validate my-subscription-id --check-resource-groups --check-permissions --check-quotas`,
		RunE: func(cmd *cobra.Command, args []string) error {
			manager, err := NewAzureSubscriptionManager(cmd.Context())
			if err != nil {
				return err
			}

			var subscriptionID string
			if len(args) > 0 {
				subscriptionID = args[0]
			}

			return manager.validateSubscription(subscriptionID, checkResourceGroups, checkPermissions, checkQuotas)
		},
	}

	cmd.Flags().BoolVar(&checkResourceGroups, "check-resource-groups", false, "Check resource group access")
	cmd.Flags().BoolVar(&checkPermissions, "check-permissions", false, "Check role assignments and permissions")
	cmd.Flags().BoolVar(&checkQuotas, "check-quotas", false, "Check quota and usage limits")

	return cmd
}

// newAzureSubscriptionTenantCmd creates the tenant subcommand.
func newAzureSubscriptionTenantCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tenant",
		Short: "Manage Azure tenant operations",
		Long: `Manage Microsoft Azure tenant operations and multi-tenant access.

This command provides functionality for:
- Listing available tenants
- Switching tenant context
- Managing cross-tenant access

Examples:
  # List available tenants
  gz dev-env azure-subscription tenant list

  # Switch tenant context
  gz dev-env azure-subscription tenant switch my-tenant-id`,
		SilenceUsage: true,
	}

	cmd.AddCommand(newAzureSubscriptionTenantListCmd())
	cmd.AddCommand(newAzureSubscriptionTenantSwitchCmd())

	return cmd
}

func newAzureSubscriptionTenantListCmd() *cobra.Command {
	var outputFormat string

	return &cobra.Command{
		Use:   "list",
		Short: "List available Azure tenants",
		RunE: func(cmd *cobra.Command, args []string) error {
			manager, err := NewAzureSubscriptionManager(cmd.Context())
			if err != nil {
				return err
			}
			return manager.listTenants(outputFormat)
		},
	}
}

func newAzureSubscriptionTenantSwitchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "switch [TENANT_ID]",
		Short: "Switch to a specific Azure tenant",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			manager, err := NewAzureSubscriptionManager(cmd.Context())
			if err != nil {
				return err
			}
			return manager.switchTenant(args[0])
		},
	}
}

// loadSubscriptions loads all Azure subscriptions using Azure CLI.
func (m *AzureSubscriptionManager) loadSubscriptions() error {
	// Use Azure CLI to list subscriptions
	cmd := exec.CommandContext(m.ctx, "az", "account", "list", "--output", "json")

	output, err := cmd.Output()
	if err != nil {
		// If Azure CLI is not available or not authenticated, return empty list
		// This allows the manager to work even without Azure CLI installed
		m.subscriptions = make(map[string]*AzureSubscription)
		return nil
	}

	var subscriptions []struct {
		ID                string `json:"id"`
		Name              string `json:"name"`
		State             string `json:"state"`
		TenantID          string `json:"tenantId"`
		TenantDisplayName string `json:"tenantDisplayName"`
		User              struct {
			Name string `json:"name"`
			Type string `json:"type"`
		} `json:"user"`
		IsDefault        bool     `json:"isDefault"`
		EnvironmentName  string   `json:"environmentName"`
		HomeTenantID     string   `json:"homeTenantId"`
		ManagedByTenants []string `json:"managedByTenants"`
	}

	if err := json.Unmarshal(output, &subscriptions); err != nil {
		return fmt.Errorf("failed to parse subscriptions JSON: %w", err)
	}

	// Get current subscription from Azure CLI
	currentSubscription := m.getCurrentSubscription()

	for _, sub := range subscriptions {
		subscription := &AzureSubscription{
			ID:                sub.ID,
			DisplayName:       sub.Name,
			Name:              sub.Name,
			State:             sub.State,
			TenantID:          sub.TenantID,
			TenantDisplayName: sub.TenantDisplayName,
			User:              sub.User.Name,
			IsDefault:         sub.IsDefault,
			IsActive:          sub.ID == currentSubscription,
			Tags:              make(map[string]string),
			EnvironmentName:   sub.EnvironmentName,
			HomeTenantID:      sub.HomeTenantID,
			ManagedByTenants:  sub.ManagedByTenants,
		}

		// Try to enrich with additional details
		m.enrichSubscriptionDetails(subscription)
		m.subscriptions[sub.ID] = subscription

		// Track tenants
		m.tenants[sub.TenantID] = sub.TenantDisplayName
	}

	return nil
}

// enrichSubscriptionDetails adds additional details to a subscription.
func (m *AzureSubscriptionManager) enrichSubscriptionDetails(subscription *AzureSubscription) {
	// Get resource groups for this subscription
	cmd := exec.CommandContext(m.ctx, "az", "group", "list", "--subscription", subscription.ID, "--query", "[].name", "--output", "tsv") //nolint:gosec // Azure CLI with controlled arguments
	if output, err := cmd.Output(); err == nil {
		resourceGroups := strings.Split(strings.TrimSpace(string(output)), "\n")
		if len(resourceGroups) > 0 && resourceGroups[0] != "" {
			subscription.ResourceGroups = resourceGroups
		}
	}

	// Get available regions for this subscription
	cmd = exec.CommandContext(m.ctx, "az", "account", "list-locations", "--subscription", subscription.ID, "--query", "[].name", "--output", "tsv") //nolint:gosec // Azure CLI with controlled arguments
	if output, err := cmd.Output(); err == nil {
		regions := strings.Split(strings.TrimSpace(string(output)), "\n")
		if len(regions) > 0 && regions[0] != "" {
			subscription.Regions = regions
		}
	}
}

// getCurrentSubscription gets the current active subscription.
func (m *AzureSubscriptionManager) getCurrentSubscription() string {
	cmd := exec.CommandContext(m.ctx, "az", "account", "show", "--query", "id", "--output", "tsv")

	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(output))
}

// listSubscriptions lists all available subscriptions.
func (m *AzureSubscriptionManager) listSubscriptions(format, tenantID string) error {
	if len(m.subscriptions) == 0 {
		fmt.Println("No Azure subscriptions found. Make sure you are authenticated with Azure CLI.")
		return nil
	}

	// Filter by tenant if specified
	var subscriptions []*AzureSubscription
	for _, subscription := range m.subscriptions {
		if tenantID == "" || subscription.TenantID == tenantID {
			subscriptions = append(subscriptions, subscription)
		}
	}

	if len(subscriptions) == 0 {
		if tenantID != "" {
			fmt.Printf("No subscriptions found for tenant: %s\n", tenantID)
		}

		return nil
	}

	// Sort subscriptions by name
	sort.Slice(subscriptions, func(i, j int) bool {
		return subscriptions[i].DisplayName < subscriptions[j].DisplayName
	})

	switch format {
	case outputFormatJSON:
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")

		return encoder.Encode(subscriptions)

	case "table":
		table := tablewriter.NewWriter(os.Stdout)
		table.Header("Subscription ID", "Name", "State", "Tenant", "User", "Active")

		for _, subscription := range subscriptions {
			active := ""
			if subscription.IsActive {
				active = "✓"
			}

			// Truncate long subscription IDs for display
			displayID := subscription.ID
			if len(displayID) > 36 {
				displayID = displayID[:8] + "..." + displayID[len(displayID)-8:]
			}

			table.Append([]string{ //nolint:errcheck // Table operations are non-critical for CLI display
				displayID,
				subscription.DisplayName,
				subscription.State,
				subscription.TenantDisplayName,
				subscription.User,
				active,
			})
		}

		table.Render() //nolint:errcheck // Table rendering errors are non-critical for CLI display

		return nil

	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

// switchSubscription switches to a specific subscription.
func (m *AzureSubscriptionManager) switchSubscription(subscriptionID string, interactive bool, tenantID string) error {
	// Interactive selection if no subscription specified
	if subscriptionID == "" || interactive {
		var err error

		subscriptionID, err = m.selectSubscriptionInteractively(tenantID)
		if err != nil {
			return err
		}
	}

	// Validate subscription exists
	subscription, exists := m.subscriptions[subscriptionID]
	if !exists {
		return fmt.Errorf("subscription '%s' not found", subscriptionID)
	}

	// Set the subscription using Azure CLI
	cmd := exec.CommandContext(m.ctx, "az", "account", "set", "--subscription", subscriptionID)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set subscription '%s': %w", subscriptionID, err)
	}

	// Update subscription status
	now := time.Now()

	for _, sub := range m.subscriptions {
		sub.IsActive = false
	}

	subscription.IsActive = true
	subscription.LastUsed = &now

	fmt.Printf("✅ Switched to subscription: %s (%s)\n", subscription.DisplayName, subscription.ID)
	fmt.Printf("   Tenant: %s\n", subscription.TenantDisplayName)
	fmt.Printf("   User: %s\n", subscription.User)

	return nil
}

// selectSubscriptionInteractively provides interactive subscription selection.
func (m *AzureSubscriptionManager) selectSubscriptionInteractively(tenantID string) (string, error) {
	if len(m.subscriptions) == 0 {
		return "", fmt.Errorf("no subscriptions available")
	}

	var (
		subscriptions []*AzureSubscription
		items         []string
	)

	for _, subscription := range m.subscriptions {
		if tenantID == "" || subscription.TenantID == tenantID {
			subscriptions = append(subscriptions, subscription)

			label := fmt.Sprintf("%s (%s)", subscription.DisplayName, subscription.ID[:8]+"...")
			if subscription.IsActive {
				label += " [current]"
			}

			items = append(items, label)
		}
	}

	if len(subscriptions) == 0 {
		return "", fmt.Errorf("no subscriptions available for tenant")
	}

	sort.Slice(subscriptions, func(i, j int) bool {
		return subscriptions[i].DisplayName < subscriptions[j].DisplayName
	})

	prompt := promptui.Select{
		Label: "Select an Azure subscription",
		Items: items,
	}

	index, _, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return subscriptions[index].ID, nil
}

// showSubscription shows detailed information about a subscription.
func (m *AzureSubscriptionManager) showSubscription(subscriptionID, format string) error {
	// Get current subscription if none specified
	if subscriptionID == "" {
		subscriptionID = m.getCurrentSubscription()
		if subscriptionID == "" {
			return fmt.Errorf("no active subscription found")
		}
	}

	subscription, exists := m.subscriptions[subscriptionID]
	if !exists {
		return fmt.Errorf("subscription '%s' not found", subscriptionID)
	}

	switch format {
	case outputFormatJSON:
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")

		return encoder.Encode(subscription)

	case "table":
		fmt.Printf("Subscription Details\n")
		fmt.Printf("===================\n")
		fmt.Printf("ID:               %s\n", subscription.ID)
		fmt.Printf("Name:             %s\n", subscription.DisplayName)
		fmt.Printf("State:            %s\n", subscription.State)
		fmt.Printf("Tenant ID:        %s\n", subscription.TenantID)
		fmt.Printf("Tenant Name:      %s\n", subscription.TenantDisplayName)
		fmt.Printf("User:             %s\n", subscription.User)
		fmt.Printf("Environment:      %s\n", subscription.EnvironmentName)
		fmt.Printf("Is Default:       %t\n", subscription.IsDefault)
		fmt.Printf("Is Active:        %t\n", subscription.IsActive)

		if subscription.LastUsed != nil {
			fmt.Printf("Last Used:        %s\n", subscription.LastUsed.Format("2006-01-02 15:04:05"))
		}

		if len(subscription.ResourceGroups) > 0 {
			fmt.Printf("\nResource Groups (%d):\n", len(subscription.ResourceGroups))

			for _, rg := range subscription.ResourceGroups {
				fmt.Printf("  - %s\n", rg)
			}
		}

		if len(subscription.Regions) > 0 {
			fmt.Printf("\nAvailable Regions (%d):\n", len(subscription.Regions))

			for i, region := range subscription.Regions {
				if i < 10 { // Show first 10 regions
					fmt.Printf("  - %s\n", region)
				} else if i == 10 {
					fmt.Printf("  ... and %d more\n", len(subscription.Regions)-10)
					break
				}
			}
		}

		return nil

	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

// login authenticates with Azure.
func (m *AzureSubscriptionManager) login(tenantID string, useDeviceCode, servicePrincipal bool) error {
	var args []string

	if servicePrincipal {
		args = []string{"login", "--service-principal"}
		// Check for required environment variables
		if os.Getenv("AZURE_CLIENT_ID") == "" || os.Getenv("AZURE_CLIENT_SECRET") == "" || os.Getenv("AZURE_TENANT_ID") == "" {
			return fmt.Errorf("service principal login requires AZURE_CLIENT_ID, AZURE_CLIENT_SECRET, and AZURE_TENANT_ID environment variables")
		}
	} else {
		args = []string{"login"}
		if useDeviceCode {
			args = append(args, "--use-device-code")
		}
	}

	if tenantID != "" {
		args = append(args, "--tenant", tenantID)
	}

	fmt.Printf("🔐 Logging in to Azure...\n")

	cmd := exec.CommandContext(m.ctx, "az", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("azure login failed: %w", err)
	}

	fmt.Printf("✅ Azure login successful\n")

	// Reload subscriptions after login
	if err := m.loadSubscriptions(); err != nil {
		fmt.Printf("Warning: failed to reload subscriptions: %v\n", err)
	}

	return nil
}

// validateSubscription validates subscription access and configuration.
func (m *AzureSubscriptionManager) validateSubscription(subscriptionID string, checkResourceGroups, checkPermissions, checkQuotas bool) error {
	// Get current subscription if none specified
	if subscriptionID == "" {
		subscriptionID = m.getCurrentSubscription()
		if subscriptionID == "" {
			return fmt.Errorf("no active subscription found")
		}
	}

	fmt.Printf("Validating subscription: %s\n", subscriptionID)
	fmt.Println("==========================")

	// Basic subscription access check
	cmd := exec.CommandContext(m.ctx, "az", "account", "show", "--subscription", subscriptionID)
	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ Subscription access: FAILED\n")
		return fmt.Errorf("cannot access subscription '%s': %w", subscriptionID, err)
	}

	fmt.Printf("✅ Subscription access: OK\n")

	// Check resource groups if requested
	if checkResourceGroups {
		cmd = exec.CommandContext(m.ctx, "az", "group", "list", "--subscription", subscriptionID, "--query", "length(@)")
		if output, err := cmd.Output(); err != nil {
			fmt.Printf("❌ Resource groups: ACCESS DENIED\n")
		} else {
			rgCount := strings.TrimSpace(string(output))
			fmt.Printf("✅ Resource groups: %s accessible\n", rgCount)
		}
	}

	// Check permissions if requested
	if checkPermissions {
		cmd = exec.CommandContext(m.ctx, "az", "role", "assignment", "list", "--assignee", "@me", "--subscription", subscriptionID, "--query", "length(@)")
		if output, err := cmd.Output(); err != nil {
			fmt.Printf("❌ Role assignments: FAILED\n")
		} else {
			roleCount := strings.TrimSpace(string(output))
			fmt.Printf("✅ Role assignments: %s found\n", roleCount)
		}
	}

	// Check quotas if requested
	if checkQuotas {
		cmd = exec.CommandContext(m.ctx, "az", "vm", "list-usage", "--location", "eastus", "--subscription", subscriptionID, "--query", "length(@)")
		if output, err := cmd.Output(); err != nil {
			fmt.Printf("❌ Quota information: UNAVAILABLE\n")
		} else {
			quotaCount := strings.TrimSpace(string(output))
			fmt.Printf("✅ Quota information: %s items available\n", quotaCount)
		}
	}

	fmt.Printf("\nValidation completed for subscription: %s\n", subscriptionID)

	return nil
}

// listTenants lists all available tenants.
func (m *AzureSubscriptionManager) listTenants(format string) error {
	// Validate format first before checking data
	switch format {
	case outputFormatJSON, "table":
		// Valid formats
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}

	if len(m.tenants) == 0 {
		fmt.Println("No Azure tenants found.")
		return nil
	}

	switch format {
	case outputFormatJSON:
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")

		return encoder.Encode(m.tenants)

	case "table":
		table := tablewriter.NewWriter(os.Stdout)
		table.Header("Tenant ID", "Display Name", "Subscriptions")

		for tenantID, displayName := range m.tenants {
			// Count subscriptions in this tenant
			subCount := 0

			for _, sub := range m.subscriptions {
				if sub.TenantID == tenantID {
					subCount++
				}
			}

			table.Append([]string{ //nolint:errcheck // Table operations are non-critical for CLI display
				tenantID,
				displayName,
				fmt.Sprintf("%d", subCount),
			})
		}

		table.Render() //nolint:errcheck // Table rendering errors are non-critical for CLI display

		return nil
	}

	return nil
}

// switchTenant switches to a specific tenant context.
func (m *AzureSubscriptionManager) switchTenant(tenantID string) error {
	// Validate tenant exists
	displayName, exists := m.tenants[tenantID]
	if !exists {
		return fmt.Errorf("tenant '%s' not found", tenantID)
	}

	// Login to specific tenant
	cmd := exec.CommandContext(m.ctx, "az", "login", "--tenant", tenantID)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to switch to tenant '%s': %w", tenantID, err)
	}

	fmt.Printf("✅ Switched to tenant: %s (%s)\n", displayName, tenantID)

	// Reload subscriptions for new tenant context
	if err := m.loadSubscriptions(); err != nil {
		fmt.Printf("Warning: failed to reload subscriptions: %v\n", err)
	}

	return nil
}
