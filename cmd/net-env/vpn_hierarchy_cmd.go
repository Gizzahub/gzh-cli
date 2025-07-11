package netenv

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/gizzahub/gzh-manager-go/pkg/cloud"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// newVPNHierarchyCmd creates the VPN hierarchy management command
func newVPNHierarchyCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vpn-hierarchy",
		Short: "Manage hierarchical VPN connections",
		Long: `Manage hierarchical VPN connections with layered priorities and dependencies.

This command provides comprehensive management of VPN connections organized in hierarchical
structures, allowing for:
- Site-to-site VPN prioritization
- Personal VPN auxiliary connections
- Automatic failover between hierarchical levels
- Environment-specific connection mappings

Examples:
  # Show VPN hierarchy
  gz net-env vpn-hierarchy show
  
  # Connect hierarchical VPN starting from root
  gz net-env vpn-hierarchy connect --root corp-vpn
  
  # List connections by layer
  gz net-env vpn-hierarchy layers
  
  # Auto-connect for current environment
  gz net-env vpn-hierarchy auto-connect
  
  # Add new hierarchical connection
  gz net-env vpn-hierarchy add --config connection.yaml`,
		SilenceUsage: true,
	}

	// Add subcommands
	cmd.AddCommand(newVPNHierarchyShowCmd(logger, configDir))
	cmd.AddCommand(newVPNHierarchyConnectCmd(logger, configDir))
	cmd.AddCommand(newVPNHierarchyDisconnectCmd(logger, configDir))
	cmd.AddCommand(newVPNHierarchyLayersCmd(logger, configDir))
	cmd.AddCommand(newVPNHierarchyAddCmd(logger, configDir))
	cmd.AddCommand(newVPNHierarchyRemoveCmd(logger, configDir))
	cmd.AddCommand(newVPNHierarchyAutoConnectCmd(logger, configDir))
	cmd.AddCommand(newVPNHierarchyValidateCmd(logger, configDir))
	cmd.AddCommand(newVPNHierarchyStatusCmd(logger, configDir))

	return cmd
}

// newVPNHierarchyShowCmd creates the show subcommand
func newVPNHierarchyShowCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show VPN hierarchy tree",
		Long:  `Display the complete VPN hierarchy tree with connections, layers, and relationships.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			manager, err := createHierarchicalVPNManager(ctx, logger, configDir)
			if err != nil {
				return fmt.Errorf("failed to create VPN manager: %w", err)
			}

			output, _ := cmd.Flags().GetString("output")

			hierarchy := manager.GetVPNHierarchy()

			switch output {
			case "json":
				return json.NewEncoder(os.Stdout).Encode(hierarchy)
			default:
				return printVPNHierarchy(hierarchy)
			}
		},
	}

	cmd.Flags().StringP("output", "o", "table", "Output format (table|json)")

	return cmd
}

// newVPNHierarchyConnectCmd creates the connect subcommand
func newVPNHierarchyConnectCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "connect",
		Short: "Connect hierarchical VPN",
		Long:  `Connect VPN connections in hierarchical order starting from root connection.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			manager, err := createHierarchicalVPNManager(ctx, logger, configDir)
			if err != nil {
				return fmt.Errorf("failed to create VPN manager: %w", err)
			}

			rootConnection, _ := cmd.Flags().GetString("root")
			if rootConnection == "" {
				return fmt.Errorf("root connection name is required")
			}

			fmt.Printf("🔗 Connecting hierarchical VPN starting from %s...\n", rootConnection)

			if err := manager.ConnectHierarchical(ctx, rootConnection); err != nil {
				return fmt.Errorf("failed to connect hierarchical VPN: %w", err)
			}

			fmt.Println("✅ Hierarchical VPN connections established successfully")
			return nil
		},
	}

	cmd.Flags().StringP("root", "r", "", "Root connection name to start hierarchy from")
	cmd.MarkFlagRequired("root")

	return cmd
}

// newVPNHierarchyDisconnectCmd creates the disconnect subcommand
func newVPNHierarchyDisconnectCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disconnect",
		Short: "Disconnect hierarchical VPN",
		Long:  `Disconnect VPN connections in reverse hierarchical order.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			manager, err := createHierarchicalVPNManager(ctx, logger, configDir)
			if err != nil {
				return fmt.Errorf("failed to create VPN manager: %w", err)
			}

			rootConnection, _ := cmd.Flags().GetString("root")
			if rootConnection == "" {
				return fmt.Errorf("root connection name is required")
			}

			fmt.Printf("🔌 Disconnecting hierarchical VPN from %s...\n", rootConnection)

			if err := manager.DisconnectHierarchical(ctx, rootConnection); err != nil {
				return fmt.Errorf("failed to disconnect hierarchical VPN: %w", err)
			}

			fmt.Println("✅ Hierarchical VPN connections disconnected successfully")
			return nil
		},
	}

	cmd.Flags().StringP("root", "r", "", "Root connection name to start disconnection from")
	cmd.MarkFlagRequired("root")

	return cmd
}

// newVPNHierarchyLayersCmd creates the layers subcommand
func newVPNHierarchyLayersCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "layers",
		Short: "Show VPN connections by layer",
		Long:  `Display VPN connections grouped by hierarchical layers.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			manager, err := createHierarchicalVPNManager(ctx, logger, configDir)
			if err != nil {
				return fmt.Errorf("failed to create VPN manager: %w", err)
			}

			output, _ := cmd.Flags().GetString("output")

			layers := manager.GetConnectionsByLayer()

			switch output {
			case "json":
				return json.NewEncoder(os.Stdout).Encode(layers)
			default:
				return printVPNLayers(layers)
			}
		},
	}

	cmd.Flags().StringP("output", "o", "table", "Output format (table|json)")

	return cmd
}

// newVPNHierarchyAutoConnectCmd creates the auto-connect subcommand
func newVPNHierarchyAutoConnectCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auto-connect",
		Short: "Auto-connect VPNs for current environment",
		Long:  `Automatically connect appropriate VPN connections for the current network environment.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
			defer cancel()

			manager, err := createHierarchicalVPNManager(ctx, logger, configDir)
			if err != nil {
				return fmt.Errorf("failed to create VPN manager: %w", err)
			}

			// Detect current network environment
			env, err := detectNetworkEnvironment(ctx)
			if err != nil {
				logger.Warn("Failed to detect network environment, using default", zap.Error(err))
				env = cloud.NetworkEnvironmentOffice // Default
			}

			fmt.Printf("🌐 Auto-connecting VPNs for environment: %s\n", env)

			if err := manager.AutoConnectForEnvironment(ctx, env); err != nil {
				return fmt.Errorf("failed to auto-connect VPNs: %w", err)
			}

			fmt.Println("✅ Auto-connection completed successfully")
			return nil
		},
	}

	return cmd
}

// newVPNHierarchyAddCmd creates the add subcommand
func newVPNHierarchyAddCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add new hierarchical VPN connection",
		Long:  `Add a new VPN connection to the hierarchical management system.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			manager, err := createHierarchicalVPNManager(ctx, logger, configDir)
			if err != nil {
				return fmt.Errorf("failed to create VPN manager: %w", err)
			}

			configFile, _ := cmd.Flags().GetString("config")
			if configFile == "" {
				return fmt.Errorf("config file is required")
			}

			// Load connection configuration
			conn, err := loadVPNConnectionConfig(configFile)
			if err != nil {
				return fmt.Errorf("failed to load connection config: %w", err)
			}

			if err := manager.AddVPNConnection(conn); err != nil {
				return fmt.Errorf("failed to add VPN connection: %w", err)
			}

			fmt.Printf("✅ Added VPN connection: %s\n", conn.Name)
			return nil
		},
	}

	cmd.Flags().StringP("config", "c", "", "Path to VPN connection configuration file")
	cmd.MarkFlagRequired("config")

	return cmd
}

// newVPNHierarchyRemoveCmd creates the remove subcommand
func newVPNHierarchyRemoveCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove VPN connection from hierarchy",
		Long:  `Remove a VPN connection from the hierarchical management system.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			manager, err := createHierarchicalVPNManager(ctx, logger, configDir)
			if err != nil {
				return fmt.Errorf("failed to create VPN manager: %w", err)
			}

			if len(args) == 0 {
				return fmt.Errorf("connection name is required")
			}

			connectionName := args[0]

			if err := manager.RemoveVPNConnection(connectionName); err != nil {
				return fmt.Errorf("failed to remove VPN connection: %w", err)
			}

			fmt.Printf("✅ Removed VPN connection: %s\n", connectionName)
			return nil
		},
	}

	return cmd
}

// newVPNHierarchyValidateCmd creates the validate subcommand
func newVPNHierarchyValidateCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate VPN hierarchy configuration",
		Long:  `Validate the VPN hierarchy configuration for circular dependencies and consistency.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			manager, err := createHierarchicalVPNManager(ctx, logger, configDir)
			if err != nil {
				return fmt.Errorf("failed to create VPN manager: %w", err)
			}

			fmt.Println("🔍 Validating VPN hierarchy configuration...")

			if err := manager.ValidateHierarchy(); err != nil {
				fmt.Printf("❌ Validation failed: %v\n", err)
				return err
			}

			fmt.Println("✅ VPN hierarchy configuration is valid")
			return nil
		},
	}

	return cmd
}

// newVPNHierarchyStatusCmd creates the status subcommand
func newVPNHierarchyStatusCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show VPN connection status",
		Long:  `Show the current status of all VPN connections in the hierarchy.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			manager, err := createHierarchicalVPNManager(ctx, logger, configDir)
			if err != nil {
				return fmt.Errorf("failed to create VPN manager: %w", err)
			}

			output, _ := cmd.Flags().GetString("output")

			status := manager.GetConnectionStatus()

			switch output {
			case "json":
				return json.NewEncoder(os.Stdout).Encode(status)
			default:
				return printVPNStatus(status)
			}
		},
	}

	cmd.Flags().StringP("output", "o", "table", "Output format (table|json)")

	return cmd
}

// Helper functions

func createHierarchicalVPNManager(ctx context.Context, logger *zap.Logger, configDir string) (*cloud.HierarchicalVPNManager, error) {
	// Create base VPN manager
	baseManager, err := cloud.NewVPNManager(logger, configDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create base VPN manager: %w", err)
	}

	// Create hierarchical manager
	hierarchicalManager := cloud.NewHierarchicalVPNManager(baseManager)

	// Load existing VPN configurations
	if err := loadVPNHierarchyConfig(hierarchicalManager, configDir); err != nil {
		logger.Warn("Failed to load VPN hierarchy config", zap.Error(err))
	}

	return hierarchicalManager, nil
}

func loadVPNHierarchyConfig(manager *cloud.HierarchicalVPNManager, configDir string) error {
	configPath := filepath.Join(configDir, "vpn-hierarchy.yaml")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Config file doesn't exist, that's ok
		return nil
	}

	// TODO: Implement YAML configuration loading
	// This would load VPN connections from configuration files
	return nil
}

func loadVPNConnectionConfig(configFile string) (*cloud.VPNConnection, error) {
	// TODO: Implement VPN connection configuration loading
	// This would parse YAML/JSON configuration files
	return &cloud.VPNConnection{
		Name: "example-vpn",
		Type: cloud.VPNTypeOpenVPN,
		Endpoint: cloud.VPNEndpoint{
			Host: "vpn.example.com",
			Port: 1194,
		},
		Priority:    100,
		AutoConnect: true,
	}, nil
}

func detectNetworkEnvironment(ctx context.Context) (cloud.NetworkEnvironment, error) {
	// TODO: Implement network environment detection logic
	// This could check:
	// - Current WiFi SSID
	// - Network IP ranges
	// - Available services
	// - DNS servers

	return cloud.NetworkEnvironmentOffice, nil
}

func printVPNHierarchy(hierarchy map[string]*cloud.VPNHierarchyNode) error {
	fmt.Printf("🌐 VPN Hierarchy Tree\n\n")

	if len(hierarchy) == 0 {
		fmt.Println("  No VPN connections configured.")
		return nil
	}

	// Find root nodes (nodes without parents)
	var roots []*cloud.VPNHierarchyNode
	for _, node := range hierarchy {
		if node.Parent == nil {
			roots = append(roots, node)
		}
	}

	// Print each root tree
	for i, root := range roots {
		if i > 0 {
			fmt.Println()
		}
		printHierarchyNode(root, 0)
	}

	return nil
}

func printHierarchyNode(node *cloud.VPNHierarchyNode, indent int) {
	indentStr := ""
	for i := 0; i < indent; i++ {
		indentStr += "  "
	}

	conn := node.Connection
	fmt.Printf("%s├─ %s (Layer %d, Priority %d, %s)\n",
		indentStr, conn.Name, node.Layer, conn.Priority, node.SiteType)

	// Print children
	for _, child := range node.Children {
		printHierarchyNode(child, indent+1)
	}
}

func printVPNLayers(layers map[int][]*cloud.VPNConnection) error {
	fmt.Printf("📊 VPN Connections by Layer\n\n")

	if len(layers) == 0 {
		fmt.Println("  No VPN connections configured.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	fmt.Fprintln(w, "LAYER\tCONNECTION\tTYPE\tPRIORITY\tAUTO-CONNECT\tENDPOINT")

	// Sort layers by number
	var layerNumbers []int
	for layer := range layers {
		layerNumbers = append(layerNumbers, layer)
	}

	for _, layer := range layerNumbers {
		connections := layers[layer]
		for i, conn := range connections {
			layerStr := ""
			if i == 0 {
				layerStr = strconv.Itoa(layer)
			}

			autoConnect := "No"
			if conn.AutoConnect {
				autoConnect = "Yes"
			}

			endpoint := fmt.Sprintf("%s:%d", conn.Endpoint.Host, conn.Endpoint.Port)

			fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\t%s\n",
				layerStr, conn.Name, conn.Type, conn.Priority, autoConnect, endpoint)
		}
	}

	return w.Flush()
}

func printVPNStatus(status map[string]*cloud.VPNStatus) error {
	fmt.Printf("📡 VPN Connection Status\n\n")

	if len(status) == 0 {
		fmt.Println("  No VPN connections found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	fmt.Fprintln(w, "CONNECTION\tSTATE\tUPTIME\tBYTES IN\tBYTES OUT\tLAST ERROR")

	for name, stat := range status {
		uptime := "-"
		if stat.ConnectedAt != nil && stat.State == cloud.VPNStateConnected {
			uptime = time.Since(*stat.ConnectedAt).Round(time.Second).String()
		}

		bytesIn := formatBytes(stat.BytesReceived)
		bytesOut := formatBytes(stat.BytesSent)

		lastError := "-"
		if stat.LastError != "" {
			lastError = truncateStringUtil(stat.LastError, 30)
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			name, stat.State, uptime, bytesIn, bytesOut, lastError)
	}

	return w.Flush()
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
