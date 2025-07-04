package sshconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	bulkclone "github.com/gizzahub/gzh-manager-go/pkg/bulk-clone"
	"github.com/spf13/cobra"
)

type sshConfigValidateOptions struct {
	configFile string
	useConfig  bool
	sshConfig  string
	keyDir     string
}

func defaultSSHConfigValidateOptions() *sshConfigValidateOptions {
	homeDir, _ := os.UserHomeDir()
	return &sshConfigValidateOptions{
		sshConfig: filepath.Join(homeDir, ".ssh", "config"),
		keyDir:    filepath.Join(homeDir, ".ssh"),
	}
}

func newSSHConfigValidateCmd() *cobra.Command {
	o := defaultSSHConfigValidateOptions()

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate SSH configuration for Git hosting services",
		Long: `Validate that SSH configuration is properly set up for your Git hosting services.

This command checks:
- SSH config file exists and is readable
- Required SSH keys exist and have correct permissions
- SSH hosts are properly configured for each organization
- Git protocol settings match SSH configuration

Examples:
  # Validate SSH config based on bulk-clone.yaml
  gz ssh-config validate --config bulk-clone.yaml
  
  # Validate with custom SSH config file
  gz ssh-config validate --config bulk-clone.yaml --ssh-config ~/custom-ssh-config`,
		RunE: o.run,
	}

	cmd.Flags().StringVarP(&o.configFile, "config", "c", o.configFile, "Path to bulk-clone config file")
	cmd.Flags().BoolVar(&o.useConfig, "use-config", false, "Use config file from standard locations")
	cmd.Flags().StringVar(&o.sshConfig, "ssh-config", o.sshConfig, "Path to SSH config file (default: ~/.ssh/config)")
	cmd.Flags().StringVar(&o.keyDir, "key-dir", o.keyDir, "Directory containing SSH keys (default: ~/.ssh)")

	// Mark flags as mutually exclusive
	cmd.MarkFlagsMutuallyExclusive("config", "use-config")
	cmd.MarkFlagsOneRequired("config", "use-config")

	return cmd
}

func (o *sshConfigValidateOptions) run(_ *cobra.Command, args []string) error {
	// Load configuration
	var configPath string
	if o.configFile != "" {
		configPath = o.configFile
	}

	cfg, err := bulkclone.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Validate SSH configuration
	issues := o.validateSSHConfig(cfg)

	if len(issues) == 0 {
		fmt.Println("✅ SSH configuration validation passed!")
		return nil
	}

	fmt.Println("❌ SSH configuration validation failed:")
	for _, issue := range issues {
		fmt.Printf("  • %s\n", issue)
	}

	return fmt.Errorf("SSH configuration validation failed with %d issues", len(issues))
}

func (o *sshConfigValidateOptions) validateSSHConfig(cfg *bulkclone.BulkCloneConfig) []string {
	var issues []string

	// Check if SSH config file exists
	if _, err := os.Stat(o.sshConfig); os.IsNotExist(err) {
		issues = append(issues, fmt.Sprintf("SSH config file does not exist: %s", o.sshConfig))
		return issues // Can't validate further without the file
	}

	// Read SSH config file
	sshConfigContent, err := os.ReadFile(o.sshConfig)
	if err != nil {
		issues = append(issues, fmt.Sprintf("Cannot read SSH config file: %v", err))
		return issues
	}

	sshConfigText := string(sshConfigContent)

	// Validate SSH directory permissions
	if info, err := os.Stat(filepath.Dir(o.sshConfig)); err == nil {
		if info.Mode().Perm() != 0o700 {
			issues = append(issues, fmt.Sprintf("SSH directory has incorrect permissions: %o (should be 700)", info.Mode().Perm()))
		}
	}

	// Validate SSH config file permissions
	if info, err := os.Stat(o.sshConfig); err == nil {
		if info.Mode().Perm() != 0o600 {
			issues = append(issues, fmt.Sprintf("SSH config file has incorrect permissions: %o (should be 600)", info.Mode().Perm()))
		}
	}

	// Check configurations for each repository root
	for _, repoRoot := range cfg.RepoRoots {
		if repoRoot.Protocol != "ssh" {
			continue // Skip non-SSH configurations
		}

		issues = append(issues, o.validateRepoRootSSH(repoRoot, sshConfigText)...)
	}

	// Check default configurations
	if cfg.Default.Protocol == "ssh" {
		if cfg.Default.Github.OrgName != "" {
			issues = append(issues, o.validateProviderSSH("github", cfg.Default.Github.OrgName, sshConfigText)...)
		}
		if cfg.Default.Gitlab.GroupName != "" {
			issues = append(issues, o.validateProviderSSH("gitlab", cfg.Default.Gitlab.GroupName, sshConfigText)...)
		}
	}

	return issues
}

func (o *sshConfigValidateOptions) validateRepoRootSSH(repoRoot bulkclone.BulkCloneGithub, sshConfigText string) []string {
	return o.validateProviderSSH(repoRoot.Provider, repoRoot.OrgName, sshConfigText)
}

func (o *sshConfigValidateOptions) validateProviderSSH(provider, orgName, sshConfigText string) []string {
	var issues []string

	// Generate expected host alias
	hostAlias := fmt.Sprintf("%s-%s", provider, orgName)

	// Check if host is configured in SSH config
	if !strings.Contains(sshConfigText, fmt.Sprintf("Host %s", hostAlias)) {
		issues = append(issues, fmt.Sprintf("SSH host not configured: %s (for %s/%s)", hostAlias, provider, orgName))
		return issues // Can't validate key without host config
	}

	// Find and validate SSH key
	expectedKeyPaths := o.getExpectedKeyPaths(provider, orgName)
	keyFound := false

	for _, keyPath := range expectedKeyPaths {
		if _, err := os.Stat(keyPath); err == nil {
			keyFound = true

			// Validate key permissions
			if info, err := os.Stat(keyPath); err == nil {
				if info.Mode().Perm() != 0o600 {
					issues = append(issues, fmt.Sprintf("SSH key has incorrect permissions: %s (%o, should be 600)", keyPath, info.Mode().Perm()))
				}
			}

			// Check if public key exists
			pubKeyPath := keyPath + ".pub"
			if _, err := os.Stat(pubKeyPath); os.IsNotExist(err) {
				issues = append(issues, fmt.Sprintf("SSH public key missing: %s", pubKeyPath))
			} else {
				// Validate public key permissions
				if info, err := os.Stat(pubKeyPath); err == nil {
					if info.Mode().Perm() != 0o644 {
						issues = append(issues, fmt.Sprintf("SSH public key has incorrect permissions: %s (%o, should be 644)", pubKeyPath, info.Mode().Perm()))
					}
				}
			}
			break
		}
	}

	if !keyFound {
		issues = append(issues, fmt.Sprintf("SSH key not found for %s/%s. Expected locations: %v", provider, orgName, expectedKeyPaths))
	}

	return issues
}

func (o *sshConfigValidateOptions) getExpectedKeyPaths(provider, orgName string) []string {
	possibleKeys := []string{
		fmt.Sprintf("%s_%s", provider, orgName),
		fmt.Sprintf("%s-%s", provider, orgName),
		fmt.Sprintf("id_%s_%s", provider, orgName),
		fmt.Sprintf("id_rsa_%s_%s", provider, orgName),
		provider,
		fmt.Sprintf("id_%s", provider),
		"id_rsa",
	}

	var keyPaths []string
	for _, keyName := range possibleKeys {
		keyPaths = append(keyPaths, filepath.Join(o.keyDir, keyName))
	}

	return keyPaths
}
