// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package devenv

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// GCPServiceAccount represents a GCP service account.
type GCPServiceAccount struct {
	Email          string    `json:"email"`
	Name           string    `json:"name"`
	DisplayName    string    `json:"displayName"`
	ProjectID      string    `json:"projectId"`
	UniqueID       string    `json:"uniqueId"`
	Description    string    `json:"description"`
	Disabled       bool      `json:"disabled"`
	OAuth2ClientID string    `json:"oauth2ClientId"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	KeyCount       int       `json:"keyCount"`
	IsActive       bool      `json:"isActive"`
}

// GCPServiceAccountKey represents a service account key.
type GCPServiceAccountKey struct {
	Name            string    `json:"name"`
	PrivateKeyType  string    `json:"privateKeyType"`
	KeyAlgorithm    string    `json:"keyAlgorithm"`
	PrivateKeyData  string    `json:"privateKeyData"`
	ValidAfterTime  time.Time `json:"validAfterTime"`
	ValidBeforeTime time.Time `json:"validBeforeTime"`
	KeyOrigin       string    `json:"keyOrigin"`
	KeyType         string    `json:"keyType"`
}

// GCPServiceAccountManager manages service accounts.
type GCPServiceAccountManager struct {
	projectID       string
	serviceAccounts map[string]*GCPServiceAccount
	ctx             context.Context
}

// NewGCPServiceAccountManager creates a new service account manager.
func NewGCPServiceAccountManager(ctx context.Context, projectID string) (*GCPServiceAccountManager, error) {
	if projectID == "" {
		// Get current project
		cmd := exec.CommandContext(ctx, "gcloud", "config", "get-value", "project")

		output, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("failed to get current project: %w", err)
		}

		projectID = strings.TrimSpace(string(output))
	}

	manager := &GCPServiceAccountManager{
		projectID:       projectID,
		serviceAccounts: make(map[string]*GCPServiceAccount),
		ctx:             ctx,
	}

	if err := manager.loadServiceAccounts(); err != nil {
		return nil, fmt.Errorf("failed to load service accounts: %w", err)
	}

	return manager, nil
}

// newGCPServiceAccountCmd creates the service account management command.
func newGCPServiceAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "service-account",
		Short: "Manage GCP service accounts",
		Long: `Manage Google Cloud Platform service accounts.

This command provides functionality to:
- List service accounts in a project
- Create and delete service accounts
- Manage service account keys
- Set service account permissions
- Download and manage credentials

Examples:
  # List service accounts
  gz dev-env gcp-project service-account list

  # Create a new service account
  gz dev-env gcp-project service-account create --name my-service --display-name "My Service Account"

  # Create a key for a service account
  gz dev-env gcp-project service-account create-key my-service@project.iam.gserviceaccount.com

  # Set active service account
  gz dev-env gcp-project service-account activate my-service@project.iam.gserviceaccount.com`,
		SilenceUsage: true,
	}

	cmd.AddCommand(newServiceAccountListCmd())
	cmd.AddCommand(newServiceAccountCreateCmd())
	cmd.AddCommand(newServiceAccountDeleteCmd())
	cmd.AddCommand(newServiceAccountCreateKeyCmd())
	cmd.AddCommand(newServiceAccountDeleteKeyCmd())
	cmd.AddCommand(newServiceAccountActivateCmd())
	cmd.AddCommand(newServiceAccountShowCmd())

	return cmd
}

func newServiceAccountListCmd() *cobra.Command {
	var (
		outputFormat string
		projectID    string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List service accounts in the project",
		RunE: func(cmd *cobra.Command, args []string) error {
			manager, err := NewGCPServiceAccountManager(cmd.Context(), projectID)
			if err != nil {
				return err
			}
			return manager.listServiceAccounts(outputFormat)
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json)")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Project ID (defaults to current project)")

	return cmd
}

func newServiceAccountCreateCmd() *cobra.Command {
	var name, displayName, description, projectID string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new service account",
		RunE: func(cmd *cobra.Command, args []string) error {
			manager, err := NewGCPServiceAccountManager(cmd.Context(), projectID)
			if err != nil {
				return err
			}
			return manager.createServiceAccount(name, displayName, description)
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Service account name (required)")
	cmd.Flags().StringVarP(&displayName, "display-name", "d", "", "Display name for the service account")
	cmd.Flags().StringVar(&description, "description", "", "Description for the service account")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Project ID (defaults to current project)")

	_ = cmd.MarkFlagRequired("name") //nolint:errcheck // Required flag setup

	return cmd
}

func newServiceAccountDeleteCmd() *cobra.Command {
	var (
		projectID string
		force     bool
	)

	cmd := &cobra.Command{
		Use:   "delete [SERVICE_ACCOUNT_EMAIL]",
		Short: "Delete a service account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			manager, err := NewGCPServiceAccountManager(cmd.Context(), projectID)
			if err != nil {
				return err
			}
			return manager.deleteServiceAccount(args[0], force)
		},
	}

	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Project ID (defaults to current project)")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")

	return cmd
}

func newServiceAccountCreateKeyCmd() *cobra.Command {
	var projectID, keyType, outputFile string

	cmd := &cobra.Command{
		Use:   "create-key [SERVICE_ACCOUNT_EMAIL]",
		Short: "Create a key for a service account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			manager, err := NewGCPServiceAccountManager(cmd.Context(), projectID)
			if err != nil {
				return err
			}
			return manager.createServiceAccountKey(args[0], keyType, outputFile)
		},
	}

	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Project ID (defaults to current project)")
	cmd.Flags().StringVarP(&keyType, "key-type", "t", "json", "Key file type (json, p12)")
	cmd.Flags().StringVarP(&outputFile, "output-file", "o", "", "Output file path (defaults to service account name)")

	return cmd
}

func newServiceAccountDeleteKeyCmd() *cobra.Command {
	var (
		projectID string
		force     bool
	)

	cmd := &cobra.Command{
		Use:   "delete-key [SERVICE_ACCOUNT_EMAIL] [KEY_ID]",
		Short: "Delete a service account key",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			manager, err := NewGCPServiceAccountManager(cmd.Context(), projectID)
			if err != nil {
				return err
			}
			return manager.deleteServiceAccountKey(args[0], args[1], force)
		},
	}

	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Project ID (defaults to current project)")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")

	return cmd
}

func newServiceAccountActivateCmd() *cobra.Command {
	var keyFile string

	cmd := &cobra.Command{
		Use:   "activate [SERVICE_ACCOUNT_EMAIL]",
		Short: "Set the active service account for authentication",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			manager, err := NewGCPServiceAccountManager(cmd.Context(), "")
			if err != nil {
				return err
			}
			return manager.activateServiceAccount(args[0], keyFile)
		},
	}

	cmd.Flags().StringVarP(&keyFile, "key-file", "k", "", "Path to service account key file")

	return cmd
}

func newServiceAccountShowCmd() *cobra.Command {
	var projectID, outputFormat string

	cmd := &cobra.Command{
		Use:   "show [SERVICE_ACCOUNT_EMAIL]",
		Short: "Show detailed information about a service account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			manager, err := NewGCPServiceAccountManager(cmd.Context(), projectID)
			if err != nil {
				return err
			}
			return manager.showServiceAccount(args[0], outputFormat)
		},
	}

	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Project ID (defaults to current project)")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json)")

	return cmd
}

func (m *GCPServiceAccountManager) loadServiceAccounts() error {
	cmd := exec.CommandContext(m.ctx, "gcloud", "iam", "service-accounts", "list",
		"--project", m.projectID, "--format=json")

	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to list service accounts: %w", err)
	}

	var accounts []struct {
		Email          string `json:"email"`
		Name           string `json:"name"`
		DisplayName    string `json:"displayName"`
		ProjectID      string `json:"projectId"`
		UniqueID       string `json:"uniqueId"`
		Description    string `json:"description"`
		Disabled       bool   `json:"disabled"`
		OAuth2ClientID string `json:"oauth2ClientId"`
	}

	if err := json.Unmarshal(output, &accounts); err != nil {
		return fmt.Errorf("failed to parse service accounts JSON: %w", err)
	}

	// Get currently active service account
	activeAccount := m.getActiveServiceAccount()

	for _, acc := range accounts {
		serviceAccount := &GCPServiceAccount{
			Email:          acc.Email,
			Name:           acc.Name,
			DisplayName:    acc.DisplayName,
			ProjectID:      acc.ProjectID,
			UniqueID:       acc.UniqueID,
			Description:    acc.Description,
			Disabled:       acc.Disabled,
			OAuth2ClientID: acc.OAuth2ClientID,
			IsActive:       acc.Email == activeAccount,
		}

		// Get key count for each service account
		m.enrichServiceAccountDetails(serviceAccount)
		m.serviceAccounts[acc.Email] = serviceAccount
	}

	return nil
}

func (m *GCPServiceAccountManager) enrichServiceAccountDetails(account *GCPServiceAccount) {
	// Get service account keys
	cmd := exec.CommandContext(m.ctx, "gcloud", "iam", "service-accounts", "keys", "list",
		"--iam-account", account.Email, "--project", m.projectID, "--format=json")
	if output, err := cmd.Output(); err == nil {
		var keys []interface{}
		if err := json.Unmarshal(output, &keys); err == nil {
			account.KeyCount = len(keys)
		}
	}
}

func (m *GCPServiceAccountManager) getActiveServiceAccount() string {
	cmd := exec.CommandContext(m.ctx, "gcloud", "auth", "list", "--filter=status:ACTIVE", "--format=value(account)")
	if output, err := cmd.Output(); err == nil {
		return strings.TrimSpace(string(output))
	}

	return ""
}

func (m *GCPServiceAccountManager) listServiceAccounts(format string) error {
	if len(m.serviceAccounts) == 0 {
		fmt.Println("No service accounts found in project:", m.projectID)
		return nil
	}

	switch format {
	case "json":
		var accounts []*GCPServiceAccount
		for _, account := range m.serviceAccounts {
			accounts = append(accounts, account)
		}

		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")

		return encoder.Encode(accounts)

	case "table":
		table := tablewriter.NewWriter(os.Stdout)
		table.Header("Email", "Display Name", "Keys", "Disabled", "Active")

		for _, account := range m.serviceAccounts {
			disabled := ""
			if account.Disabled {
				disabled = "✓"
			}

			active := ""
			if account.IsActive {
				active = "✓"
			}

			_ = table.Append([]string{ //nolint:errcheck // Table display errors are non-critical
				account.Email,
				account.DisplayName,
				fmt.Sprintf("%d", account.KeyCount),
				disabled,
				active,
			})
		}

		_ = table.Render() //nolint:errcheck // Table display errors are non-critical

		return nil

	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

func (m *GCPServiceAccountManager) createServiceAccount(name, displayName, description string) error {
	args := []string{"iam", "service-accounts", "create", name, "--project", m.projectID}

	if displayName != "" {
		args = append(args, "--display-name", displayName)
	}

	if description != "" {
		args = append(args, "--description", description)
	}

	cmd := exec.CommandContext(m.ctx, "gcloud", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create service account: %w", err)
	}

	email := fmt.Sprintf("%s@%s.iam.gserviceaccount.com", name, m.projectID)
	fmt.Printf("✅ Service account created: %s\n", email)

	// Reload service accounts
	if err := m.loadServiceAccounts(); err != nil {
		fmt.Printf("Warning: failed to reload service accounts: %v\n", err)
	}

	return nil
}

func (m *GCPServiceAccountManager) deleteServiceAccount(email string, force bool) error {
	if !force {
		prompt := promptui.Prompt{
			Label:     fmt.Sprintf("Delete service account '%s'", email),
			IsConfirm: true,
		}

		if _, err := prompt.Run(); err != nil {
			return fmt.Errorf("operation canceled")
		}
	}

	cmd := exec.CommandContext(m.ctx, "gcloud", "iam", "service-accounts", "delete",
		email, "--project", m.projectID, "--quiet")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete service account: %w", err)
	}

	fmt.Printf("✅ Service account deleted: %s\n", email)

	return nil
}

func (m *GCPServiceAccountManager) createServiceAccountKey(email, keyType, outputFile string) error {
	if outputFile == "" {
		// Generate default filename
		parts := strings.Split(email, "@")
		if len(parts) > 0 {
			outputFile = fmt.Sprintf("%s-key.%s", parts[0], keyType)
		} else {
			outputFile = fmt.Sprintf("service-account-key.%s", keyType)
		}
	}

	// Check if file already exists
	if _, err := os.Stat(outputFile); err == nil {
		prompt := promptui.Prompt{
			Label:     fmt.Sprintf("File '%s' already exists. Overwrite", outputFile),
			IsConfirm: true,
		}

		if _, err := prompt.Run(); err != nil {
			return fmt.Errorf("operation canceled")
		}
	}

	cmd := exec.CommandContext(m.ctx, "gcloud", "iam", "service-accounts", "keys", "create",
		outputFile, "--iam-account", email, "--project", m.projectID, "--key-file-type", keyType)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create service account key: %w", err)
	}

	// Set secure permissions on the key file
	if err := os.Chmod(outputFile, 0o600); err != nil {
		fmt.Printf("Warning: failed to set secure permissions on key file: %v\n", err)
	}

	fmt.Printf("✅ Service account key created: %s\n", outputFile)
	fmt.Printf("   Service account: %s\n", email)
	fmt.Printf("   Key type: %s\n", keyType)
	fmt.Printf("   ⚠️  Keep this file secure and do not commit to version control\n")

	return nil
}

func (m *GCPServiceAccountManager) deleteServiceAccountKey(email, keyID string, force bool) error {
	if !force {
		prompt := promptui.Prompt{
			Label:     fmt.Sprintf("Delete key '%s' for service account '%s'", keyID, email),
			IsConfirm: true,
		}

		if _, err := prompt.Run(); err != nil {
			return fmt.Errorf("operation canceled")
		}
	}

	cmd := exec.CommandContext(m.ctx, "gcloud", "iam", "service-accounts", "keys", "delete",
		keyID, "--iam-account", email, "--project", m.projectID, "--quiet")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete service account key: %w", err)
	}

	fmt.Printf("✅ Service account key deleted: %s\n", keyID)

	return nil
}

func (m *GCPServiceAccountManager) activateServiceAccount(email, keyFile string) error {
	if keyFile != "" {
		// Activate using key file
		cmd := exec.CommandContext(m.ctx, "gcloud", "auth", "activate-service-account",
			email, "--key-file", keyFile)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to activate service account with key file: %w", err)
		}
	} else {
		// Try to activate without key file (for existing authentication)
		cmd := exec.CommandContext(m.ctx, "gcloud", "config", "set", "account", email)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to set active account: %w", err)
		}
	}

	fmt.Printf("✅ Service account activated: %s\n", email)

	// Update active status in our cache
	for _, account := range m.serviceAccounts {
		account.IsActive = account.Email == email
	}

	return nil
}

func (m *GCPServiceAccountManager) showServiceAccount(email, format string) error {
	account, exists := m.serviceAccounts[email]
	if !exists {
		return fmt.Errorf("service account '%s' not found", email)
	}

	switch format {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")

		return encoder.Encode(account)

	case "table":
		fmt.Printf("Service Account Details\n")
		fmt.Printf("======================\n")
		fmt.Printf("Email:        %s\n", account.Email)
		fmt.Printf("Display Name: %s\n", account.DisplayName)
		fmt.Printf("Project ID:   %s\n", account.ProjectID)
		fmt.Printf("Unique ID:    %s\n", account.UniqueID)
		fmt.Printf("Description:  %s\n", account.Description)
		fmt.Printf("Disabled:     %t\n", account.Disabled)
		fmt.Printf("Key Count:    %d\n", account.KeyCount)
		fmt.Printf("Active:       %t\n", account.IsActive)

		if account.OAuth2ClientID != "" {
			fmt.Printf("OAuth2 Client ID: %s\n", account.OAuth2ClientID)
		}

		// Show keys if any exist
		if account.KeyCount > 0 {
			fmt.Printf("\nService Account Keys:\n")
			_ = m.listServiceAccountKeys(email) //nolint:errcheck // Display function errors are non-critical
		}

		return nil

	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

func (m *GCPServiceAccountManager) listServiceAccountKeys(email string) error {
	cmd := exec.CommandContext(m.ctx, "gcloud", "iam", "service-accounts", "keys", "list",
		"--iam-account", email, "--project", m.projectID, "--format=json")

	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to list service account keys: %w", err)
	}

	var keys []struct {
		Name            string `json:"name"`
		PrivateKeyType  string `json:"privateKeyType"`
		KeyAlgorithm    string `json:"keyAlgorithm"`
		ValidAfterTime  string `json:"validAfterTime"`
		ValidBeforeTime string `json:"validBeforeTime"`
		KeyOrigin       string `json:"keyOrigin"`
		KeyType         string `json:"keyType"`
	}

	if err := json.Unmarshal(output, &keys); err != nil {
		return fmt.Errorf("failed to parse keys JSON: %w", err)
	}

	if len(keys) == 0 {
		fmt.Println("No keys found for this service account")
		return nil
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.Header("Key ID", "Type", "Algorithm", "Origin", "Valid After", "Valid Before")

	for _, key := range keys {
		// Extract key ID from name (format: projects/PROJECT/serviceAccounts/EMAIL/keys/KEYID)
		parts := strings.Split(key.Name, "/")
		keyID := parts[len(parts)-1]

		_ = table.Append([]string{ //nolint:errcheck // Table display errors are non-critical
			keyID,
			key.PrivateKeyType,
			key.KeyAlgorithm,
			key.KeyOrigin,
			key.ValidAfterTime,
			key.ValidBeforeTime,
		})
	}

	_ = table.Render() //nolint:errcheck // Table display errors are non-critical

	return nil
}
