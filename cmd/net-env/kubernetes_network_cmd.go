package netenv

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// newKubernetesNetworkCmd creates the kubernetes-network command
func newKubernetesNetworkCmd(logger *zap.Logger, configDir string) *cobra.Command {
	km := NewKubernetesNetworkManager(logger, configDir)

	cmd := &cobra.Command{
		Use:     "kubernetes-network",
		Short:   "Manage Kubernetes network policies and namespace configurations",
		Long:    `Manage Kubernetes network policies, namespace-specific configurations, and service mesh integration.`,
		Aliases: []string{"k8s-network", "k8s-net"},
	}

	// Add subcommands
	cmd.AddCommand(newK8sNetworkCreateCmd(km))
	cmd.AddCommand(newK8sNetworkListCmd(km))
	cmd.AddCommand(newK8sNetworkApplyCmd(km))
	cmd.AddCommand(newK8sNetworkDeleteCmd(km))
	cmd.AddCommand(newK8sNetworkExportCmd(km))
	cmd.AddCommand(newK8sNetworkPolicyCmd(km))
	cmd.AddCommand(newK8sNetworkGenerateCmd(km))
	cmd.AddCommand(newK8sNetworkStatusCmd(km))
	cmd.AddCommand(newServiceMeshCmd(km))

	return cmd
}

// newK8sNetworkCreateCmd creates the create subcommand
func newK8sNetworkCreateCmd(km *KubernetesNetworkManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [profile-name]",
		Short: "Create a new Kubernetes network profile",
		Long:  `Create a new Kubernetes network profile with network policies and configurations.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := args[0]

			// Get flags
			description, _ := cmd.Flags().GetString("description")
			namespace, _ := cmd.Flags().GetString("namespace")
			interactive, _ := cmd.Flags().GetBool("interactive")

			profile := &KubernetesNetworkProfile{
				Name:        profileName,
				Description: description,
				Namespace:   namespace,
				Policies:    make(map[string]*NetworkPolicyConfig),
				Services:    make(map[string]*ServiceConfig),
				Ingress:     make(map[string]*IngressConfig),
				Metadata:    make(map[string]string),
			}

			if interactive {
				return createK8sProfileInteractively(km, profile)
			}

			if err := km.CreateProfile(profile); err != nil {
				return fmt.Errorf("failed to create profile: %w", err)
			}

			fmt.Printf("✅ Created Kubernetes network profile: %s\n", profileName)
			if description != "" {
				fmt.Printf("   Description: %s\n", description)
			}
			fmt.Printf("   Namespace: %s\n", profile.Namespace)

			return nil
		},
	}

	cmd.Flags().String("description", "", "Profile description")
	cmd.Flags().StringP("namespace", "n", "default", "Target namespace")
	cmd.Flags().BoolP("interactive", "i", false, "Create profile interactively")

	return cmd
}

// newK8sNetworkListCmd creates the list subcommand
func newK8sNetworkListCmd(km *KubernetesNetworkManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List Kubernetes network profiles",
		Long:    `List all available Kubernetes network profiles with their configurations.`,
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			profiles, err := km.ListProfiles()
			if err != nil {
				return fmt.Errorf("failed to list profiles: %w", err)
			}

			if len(profiles) == 0 {
				fmt.Println("No Kubernetes network profiles found.")
				return nil
			}

			// Get output format
			output, _ := cmd.Flags().GetString("output")

			switch output {
			case "json":
				return printK8sProfilesJSON(profiles)
			case "yaml":
				return printK8sProfilesYAML(profiles)
			default:
				return printK8sProfilesTable(profiles)
			}
		},
	}

	cmd.Flags().StringP("output", "o", "table", "Output format (table|json|yaml)")

	return cmd
}

// newK8sNetworkApplyCmd creates the apply subcommand
func newK8sNetworkApplyCmd(km *KubernetesNetworkManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply [profile-name]",
		Short: "Apply a Kubernetes network profile",
		Long:  `Apply a Kubernetes network profile to create network policies and configurations.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := args[0]

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				fmt.Printf("🔍 Dry run mode - would apply profile: %s\n", profileName)
				profile, err := km.LoadProfile(profileName)
				if err != nil {
					return fmt.Errorf("failed to load profile: %w", err)
				}
				return printK8sProfileDetails(profile)
			}

			fmt.Printf("⏳ Applying Kubernetes network profile: %s\n", profileName)

			if err := km.ApplyProfile(profileName); err != nil {
				return fmt.Errorf("failed to apply profile: %w", err)
			}

			fmt.Printf("✅ Successfully applied Kubernetes network profile: %s\n", profileName)
			return nil
		},
	}

	cmd.Flags().Bool("dry-run", false, "Show what would be applied without making changes")

	return cmd
}

// newK8sNetworkDeleteCmd creates the delete subcommand
func newK8sNetworkDeleteCmd(km *KubernetesNetworkManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete [profile-name]",
		Short:   "Delete a Kubernetes network profile",
		Long:    `Delete a Kubernetes network profile. This does not affect existing resources.`,
		Aliases: []string{"rm"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := args[0]

			force, _ := cmd.Flags().GetBool("force")
			if !force {
				fmt.Printf("⚠️  Are you sure you want to delete profile '%s'? (y/N): ", profileName)
				var response string
				fmt.Scanln(&response)
				if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
					fmt.Println("❌ Deletion cancelled.")
					return nil
				}
			}

			if err := km.DeleteProfile(profileName); err != nil {
				return fmt.Errorf("failed to delete profile: %w", err)
			}

			fmt.Printf("✅ Deleted Kubernetes network profile: %s\n", profileName)
			return nil
		},
	}

	cmd.Flags().BoolP("force", "f", false, "Force deletion without confirmation")

	return cmd
}

// newK8sNetworkExportCmd creates the export subcommand
func newK8sNetworkExportCmd(km *KubernetesNetworkManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export namespace policies to a profile",
		Long:  `Export existing Kubernetes network policies from a namespace to a profile.`,
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "namespace [namespace] [profile-name]",
		Short: "Export all network policies from a namespace",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			namespace := args[0]
			profileName := args[1]

			fmt.Printf("📥 Exporting network policies from namespace: %s\n", namespace)

			if err := km.ExportNamespacePolicies(namespace, profileName); err != nil {
				return fmt.Errorf("failed to export namespace policies: %w", err)
			}

			fmt.Printf("✅ Exported namespace policies to profile '%s'\n", profileName)
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "profile [profile-name] [output-file]",
		Short: "Export a profile to a file",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := args[0]
			outputFile := args[1]

			profile, err := km.LoadProfile(profileName)
			if err != nil {
				return fmt.Errorf("failed to load profile: %w", err)
			}

			format := strings.ToLower(filepath.Ext(outputFile))
			var data []byte

			switch format {
			case ".json":
				data, err = json.MarshalIndent(profile, "", "  ")
			case ".yaml", ".yml":
				data, err = yaml.Marshal(profile)
			default:
				return fmt.Errorf("unsupported output format: %s (use .json, .yaml, or .yml)", format)
			}

			if err != nil {
				return fmt.Errorf("failed to marshal profile: %w", err)
			}

			if err := os.WriteFile(outputFile, data, 0o644); err != nil {
				return fmt.Errorf("failed to write output file: %w", err)
			}

			fmt.Printf("✅ Exported profile '%s' to %s\n", profileName, outputFile)
			return nil
		},
	})

	return cmd
}

// newK8sNetworkPolicyCmd creates the policy subcommand
func newK8sNetworkPolicyCmd(km *KubernetesNetworkManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "policy",
		Short: "Manage network policies within profiles",
		Long:  `Add, update, remove, and show network policies within Kubernetes network profiles.`,
	}

	// Add policy subcommands
	cmd.AddCommand(newPolicyAddCmd(km))
	cmd.AddCommand(newPolicyRemoveCmd(km))
	cmd.AddCommand(newPolicyShowCmd(km))

	return cmd
}

// newPolicyAddCmd creates the policy add subcommand
func newPolicyAddCmd(km *KubernetesNetworkManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [profile-name] [policy-name]",
		Short: "Add a network policy to a profile",
		Long:  `Add a network policy with specific ingress and egress rules to a Kubernetes network profile.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := args[0]
			policyName := args[1]

			profile, err := km.LoadProfile(profileName)
			if err != nil {
				return fmt.Errorf("failed to load profile: %w", err)
			}

			// Get flags
			podSelector, _ := cmd.Flags().GetStringToString("pod-selector")
			policyTypes, _ := cmd.Flags().GetStringSlice("policy-types")
			allowFrom, _ := cmd.Flags().GetStringSlice("allow-from")
			allowTo, _ := cmd.Flags().GetStringSlice("allow-to")
			ports, _ := cmd.Flags().GetStringSlice("ports")

			// Build policy configuration
			config := &NetworkPolicyConfig{
				Name:        policyName,
				PodSelector: podSelector,
				PolicyTypes: policyTypes,
			}

			// Parse ingress rules if specified
			if len(allowFrom) > 0 {
				ingressRule := NetworkPolicyIngressRule{
					From: make([]NetworkPolicyPeer, 0),
				}

				for _, from := range allowFrom {
					peer := NetworkPolicyPeer{}
					if strings.HasPrefix(from, "pod:") {
						// Pod selector format: pod:app=frontend
						selector := strings.TrimPrefix(from, "pod:")
						peer.PodSelector = parseSelector(selector)
					} else if strings.HasPrefix(from, "ns:") {
						// Namespace selector format: ns:name=production
						selector := strings.TrimPrefix(from, "ns:")
						peer.NamespaceSelector = parseSelector(selector)
					} else if strings.Contains(from, "/") {
						// IP block format: 10.0.0.0/8
						peer.IPBlock = &IPBlock{CIDR: from}
					}
					ingressRule.From = append(ingressRule.From, peer)
				}

				// Parse ports
				ingressRule.Ports = parsePorts(ports)
				config.Ingress = []NetworkPolicyIngressRule{ingressRule}
			}

			// Parse egress rules if specified
			if len(allowTo) > 0 {
				egressRule := NetworkPolicyEgressRule{
					To: make([]NetworkPolicyPeer, 0),
				}

				for _, to := range allowTo {
					peer := NetworkPolicyPeer{}
					if strings.HasPrefix(to, "pod:") {
						selector := strings.TrimPrefix(to, "pod:")
						peer.PodSelector = parseSelector(selector)
					} else if strings.HasPrefix(to, "ns:") {
						selector := strings.TrimPrefix(to, "ns:")
						peer.NamespaceSelector = parseSelector(selector)
					} else if strings.Contains(to, "/") {
						peer.IPBlock = &IPBlock{CIDR: to}
					}
					egressRule.To = append(egressRule.To, peer)
				}

				egressRule.Ports = parsePorts(ports)
				config.Egress = []NetworkPolicyEgressRule{egressRule}
			}

			// Add policy to profile
			if profile.Policies == nil {
				profile.Policies = make(map[string]*NetworkPolicyConfig)
			}
			profile.Policies[policyName] = config
			profile.UpdatedAt = time.Now()

			if err := km.saveProfile(profile); err != nil {
				return fmt.Errorf("failed to save profile: %w", err)
			}

			fmt.Printf("✅ Added network policy '%s' to profile '%s'\n", policyName, profileName)
			return nil
		},
	}

	cmd.Flags().StringToString("pod-selector", map[string]string{}, "Pod selector labels (e.g., app=frontend,tier=web)")
	cmd.Flags().StringSlice("policy-types", []string{"Ingress", "Egress"}, "Policy types")
	cmd.Flags().StringSlice("allow-from", []string{}, "Allowed ingress sources (e.g., pod:app=backend, ns:name=prod, 10.0.0.0/8)")
	cmd.Flags().StringSlice("allow-to", []string{}, "Allowed egress destinations")
	cmd.Flags().StringSlice("ports", []string{}, "Allowed ports (e.g., TCP:80, UDP:53, 8080-8090)")

	return cmd
}

// newPolicyRemoveCmd creates the policy remove subcommand
func newPolicyRemoveCmd(km *KubernetesNetworkManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove [profile-name] [policy-name]",
		Short:   "Remove a network policy from a profile",
		Aliases: []string{"rm"},
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := args[0]
			policyName := args[1]

			profile, err := km.LoadProfile(profileName)
			if err != nil {
				return fmt.Errorf("failed to load profile: %w", err)
			}

			if profile.Policies == nil || profile.Policies[policyName] == nil {
				return fmt.Errorf("policy '%s' not found in profile '%s'", policyName, profileName)
			}

			delete(profile.Policies, policyName)
			profile.UpdatedAt = time.Now()

			if err := km.saveProfile(profile); err != nil {
				return fmt.Errorf("failed to save profile: %w", err)
			}

			fmt.Printf("✅ Removed network policy '%s' from profile '%s'\n", policyName, profileName)
			return nil
		},
	}

	return cmd
}

// newPolicyShowCmd creates the policy show subcommand
func newPolicyShowCmd(km *KubernetesNetworkManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show [profile-name] [policy-name]",
		Short: "Show network policy details",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := args[0]
			policyName := args[1]

			profile, err := km.LoadProfile(profileName)
			if err != nil {
				return fmt.Errorf("failed to load profile: %w", err)
			}

			policy, exists := profile.Policies[policyName]
			if !exists {
				return fmt.Errorf("policy '%s' not found in profile '%s'", policyName, profileName)
			}

			output, _ := cmd.Flags().GetString("output")
			switch output {
			case "json":
				data, err := json.MarshalIndent(policy, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
			case "yaml":
				data, err := yaml.Marshal(policy)
				if err != nil {
					return err
				}
				fmt.Print(string(data))
			default:
				printPolicyDetails(policy)
			}

			return nil
		},
	}

	cmd.Flags().StringP("output", "o", "table", "Output format (table|json|yaml)")

	return cmd
}

// newK8sNetworkGenerateCmd creates the generate subcommand
func newK8sNetworkGenerateCmd(km *KubernetesNetworkManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate Kubernetes manifests from profiles",
		Long:  `Generate Kubernetes YAML manifests from network profiles for manual review or version control.`,
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "policies [profile-name] [output-dir]",
		Short: "Generate NetworkPolicy manifests from a profile",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := args[0]
			outputDir := args[1]

			profile, err := km.LoadProfile(profileName)
			if err != nil {
				return fmt.Errorf("failed to load profile: %w", err)
			}

			// Create output directory
			if err := os.MkdirAll(outputDir, 0o755); err != nil {
				return fmt.Errorf("failed to create output directory: %w", err)
			}

			// Generate policies
			for policyName, policyConfig := range profile.Policies {
				policy, err := km.GenerateNetworkPolicy(profile.Namespace, policyConfig)
				if err != nil {
					return fmt.Errorf("failed to generate policy %s: %w", policyName, err)
				}

				// Convert to YAML
				data, err := yaml.Marshal(policy)
				if err != nil {
					return fmt.Errorf("failed to marshal policy: %w", err)
				}

				// Write to file
				filename := filepath.Join(outputDir, fmt.Sprintf("networkpolicy-%s.yaml", policyName))
				if err := os.WriteFile(filename, data, 0o644); err != nil {
					return fmt.Errorf("failed to write policy file: %w", err)
				}

				fmt.Printf("✅ Generated NetworkPolicy: %s\n", filename)
			}

			fmt.Printf("\n📁 Generated %d NetworkPolicy manifests in %s\n", len(profile.Policies), outputDir)
			return nil
		},
	})

	return cmd
}

// newK8sNetworkStatusCmd creates the status subcommand
func newK8sNetworkStatusCmd(km *KubernetesNetworkManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show Kubernetes network policy status",
		Long:  `Show the current status of network policies in namespaces.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			namespace, _ := cmd.Flags().GetString("namespace")
			output, _ := cmd.Flags().GetString("output")
			allNamespaces, _ := cmd.Flags().GetBool("all-namespaces")

			if allNamespaces {
				// Get policies from all namespaces
				namespaces := []string{"default", "kube-system", "kube-public"}
				// In real implementation, we would list all namespaces

				for _, ns := range namespaces {
					fmt.Printf("\n🌐 Namespace: %s\n", ns)
					policies, err := km.GetNamespaceNetworkPolicies(ns)
					if err != nil {
						fmt.Printf("  ❌ Error: %v\n", err)
						continue
					}

					if len(policies) == 0 {
						fmt.Println("  No network policies found.")
					} else {
						printNetworkPoliciesTable(policies)
					}
				}
			} else {
				policies, err := km.GetNamespaceNetworkPolicies(namespace)
				if err != nil {
					return fmt.Errorf("failed to get network policies: %w", err)
				}

				if output == "json" {
					return json.NewEncoder(os.Stdout).Encode(policies)
				}

				fmt.Printf("🌐 Network Policies in namespace: %s\n", namespace)
				if len(policies) == 0 {
					fmt.Println("No network policies found.")
				} else {
					printNetworkPoliciesTable(policies)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringP("namespace", "n", "default", "Target namespace")
	cmd.Flags().BoolP("all-namespaces", "A", false, "Show policies from all namespaces")
	cmd.Flags().StringP("output", "o", "table", "Output format (table|json)")

	return cmd
}

// Helper functions

func createK8sProfileInteractively(km *KubernetesNetworkManager, profile *KubernetesNetworkProfile) error {
	fmt.Printf("Creating Kubernetes network profile interactively...\n\n")

	// Get namespace
	if profile.Namespace == "" {
		fmt.Print("Enter namespace [default]: ")
		fmt.Scanln(&profile.Namespace)
		if profile.Namespace == "" {
			profile.Namespace = "default"
		}
	}

	// Get description
	if profile.Description == "" {
		fmt.Print("Enter description (optional): ")
		fmt.Scanln(&profile.Description)
	}

	// Ask if user wants to add a network policy
	fmt.Print("Add a network policy? (y/N): ")
	var addPolicy string
	fmt.Scanln(&addPolicy)

	if strings.ToLower(addPolicy) == "y" || strings.ToLower(addPolicy) == "yes" {
		// Simple deny-all policy as starting point
		fmt.Println("\nCreating a default deny-all policy...")

		denyAllPolicy := &NetworkPolicyConfig{
			Name:        "deny-all",
			PodSelector: map[string]string{}, // Empty selector applies to all pods
			PolicyTypes: []string{"Ingress", "Egress"},
		}

		profile.Policies["deny-all"] = denyAllPolicy
		fmt.Println("✅ Added deny-all network policy")
		fmt.Println("💡 You can add more specific policies using 'gz net-env kubernetes-network policy add'")
	}

	return km.CreateProfile(profile)
}

func printK8sProfilesJSON(profiles []*KubernetesNetworkProfile) error {
	return json.NewEncoder(os.Stdout).Encode(profiles)
}

func printK8sProfilesYAML(profiles []*KubernetesNetworkProfile) error {
	data, err := yaml.Marshal(profiles)
	if err != nil {
		return err
	}
	fmt.Print(string(data))
	return nil
}

func printK8sProfilesTable(profiles []*KubernetesNetworkProfile) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	fmt.Fprintln(w, "NAME\tNAMESPACE\tDESCRIPTION\tPOLICIES\tACTIVE\tCREATED")

	for _, profile := range profiles {
		active := "No"
		if profile.Active {
			active = "Yes"
		}

		created := profile.CreatedAt.Format("2006-01-02")
		if profile.CreatedAt.IsZero() {
			created = "-"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\t%s\n",
			profile.Name,
			profile.Namespace,
			truncateString(profile.Description, 40),
			len(profile.Policies),
			active,
			created)
	}

	return w.Flush()
}

func printK8sProfileDetails(profile *KubernetesNetworkProfile) error {
	fmt.Printf("Profile: %s\n", profile.Name)
	fmt.Printf("Namespace: %s\n", profile.Namespace)
	if profile.Description != "" {
		fmt.Printf("Description: %s\n", profile.Description)
	}

	fmt.Printf("\nNetwork Policies (%d):\n", len(profile.Policies))
	for name, policy := range profile.Policies {
		fmt.Printf("  • %s\n", name)
		if len(policy.PodSelector) > 0 {
			fmt.Printf("    Pod Selector: %v\n", policy.PodSelector)
		}
		if len(policy.PolicyTypes) > 0 {
			fmt.Printf("    Policy Types: %s\n", strings.Join(policy.PolicyTypes, ", "))
		}
	}

	if len(profile.Services) > 0 {
		fmt.Printf("\nServices (%d):\n", len(profile.Services))
		for name, service := range profile.Services {
			fmt.Printf("  • %s (Type: %s)\n", name, service.Type)
		}
	}

	if len(profile.Ingress) > 0 {
		fmt.Printf("\nIngress Rules (%d):\n", len(profile.Ingress))
		for name, ingress := range profile.Ingress {
			fmt.Printf("  • %s\n", name)
			for _, rule := range ingress.Rules {
				fmt.Printf("    Host: %s\n", rule.Host)
			}
		}
	}

	return nil
}

func printPolicyDetails(policy *NetworkPolicyConfig) {
	fmt.Printf("Policy: %s\n", policy.Name)

	fmt.Println("Pod Selector:")
	if len(policy.PodSelector) == 0 {
		fmt.Println("  <all pods>")
	} else {
		for k, v := range policy.PodSelector {
			fmt.Printf("  %s: %s\n", k, v)
		}
	}

	fmt.Printf("Policy Types: %s\n", strings.Join(policy.PolicyTypes, ", "))

	if len(policy.Ingress) > 0 {
		fmt.Printf("\nIngress Rules (%d):\n", len(policy.Ingress))
		for i, rule := range policy.Ingress {
			fmt.Printf("  Rule %d:\n", i+1)
			if len(rule.From) > 0 {
				fmt.Println("    From:")
				for _, peer := range rule.From {
					printPolicyPeer(&peer, "      ")
				}
			}
			if len(rule.Ports) > 0 {
				fmt.Println("    Ports:")
				for _, port := range rule.Ports {
					printPolicyPort(&port, "      ")
				}
			}
		}
	}

	if len(policy.Egress) > 0 {
		fmt.Printf("\nEgress Rules (%d):\n", len(policy.Egress))
		for i, rule := range policy.Egress {
			fmt.Printf("  Rule %d:\n", i+1)
			if len(rule.To) > 0 {
				fmt.Println("    To:")
				for _, peer := range rule.To {
					printPolicyPeer(&peer, "      ")
				}
			}
			if len(rule.Ports) > 0 {
				fmt.Println("    Ports:")
				for _, port := range rule.Ports {
					printPolicyPort(&port, "      ")
				}
			}
		}
	}
}

func printPolicyPeer(peer *NetworkPolicyPeer, indent string) {
	if len(peer.PodSelector) > 0 {
		fmt.Printf("%sPod Selector: %v\n", indent, peer.PodSelector)
	}
	if len(peer.NamespaceSelector) > 0 {
		fmt.Printf("%sNamespace Selector: %v\n", indent, peer.NamespaceSelector)
	}
	if peer.IPBlock != nil {
		fmt.Printf("%sIP Block: %s\n", indent, peer.IPBlock.CIDR)
		if len(peer.IPBlock.Except) > 0 {
			fmt.Printf("%s  Except: %v\n", indent, peer.IPBlock.Except)
		}
	}
}

func printPolicyPort(port *NetworkPolicyPort, indent string) {
	portInfo := indent
	if port.Protocol != "" {
		portInfo += port.Protocol + " "
	}
	if port.Port != nil {
		portInfo += fmt.Sprintf("Port %d", *port.Port)
	}
	if port.EndPort != nil {
		portInfo += fmt.Sprintf("-%d", *port.EndPort)
	}
	fmt.Println(portInfo)
}

func printNetworkPoliciesTable(policies []interface{}) {
	// This would print actual NetworkPolicy objects
	// For now, just print a placeholder
	fmt.Println("  NetworkPolicy list would be displayed here")
}

func parseSelector(selector string) map[string]string {
	result := make(map[string]string)
	parts := strings.Split(selector, ",")
	for _, part := range parts {
		kv := strings.Split(strings.TrimSpace(part), "=")
		if len(kv) == 2 {
			result[kv[0]] = kv[1]
		}
	}
	return result
}

func parsePorts(portSpecs []string) []NetworkPolicyPort {
	var ports []NetworkPolicyPort

	for _, spec := range portSpecs {
		port := NetworkPolicyPort{}

		// Parse protocol and port
		parts := strings.Split(spec, ":")
		if len(parts) == 2 {
			port.Protocol = parts[0]
			spec = parts[1]
		} else {
			port.Protocol = "TCP" // Default protocol
		}

		// Parse port range
		if strings.Contains(spec, "-") {
			rangeParts := strings.Split(spec, "-")
			if len(rangeParts) == 2 {
				if startPort, err := strconv.ParseInt(rangeParts[0], 10, 32); err == nil {
					startPort32 := int32(startPort)
					port.Port = &startPort32
				}
				if endPort, err := strconv.ParseInt(rangeParts[1], 10, 32); err == nil {
					endPort32 := int32(endPort)
					port.EndPort = &endPort32
				}
			}
		} else {
			// Single port
			if portNum, err := strconv.ParseInt(spec, 10, 32); err == nil {
				portNum32 := int32(portNum)
				port.Port = &portNum32
			}
		}

		ports = append(ports, port)
	}

	return ports
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
