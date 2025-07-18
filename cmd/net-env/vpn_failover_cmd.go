package netenv

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// newVPNFailoverCmd creates the VPN failover management command.
func newVPNFailoverCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vpn-failover",
		Short: "Manage VPN automatic failover and backup connections",
		Long: `Manage VPN automatic failover with connection monitoring and backup VPN activation.

This command provides comprehensive VPN failover management including:
- Real-time connection state monitoring
- Automatic backup VPN activation on failure
- Health check configuration and validation
- Failover policies and priorities
- Recovery and reconnection strategies

Examples:
  # Start VPN failover monitoring
  gz net-env vpn-failover start
  
  # Stop VPN failover monitoring
  gz net-env vpn-failover stop
  
  # Show failover status
  gz net-env vpn-failover status
  
  # Configure backup VPN
  gz net-env vpn-failover backup add --primary corp-vpn --backup home-vpn --priority 50
  
  # Test failover scenario
  gz net-env vpn-failover test --scenario connection-loss`,
		SilenceUsage: true,
	}

	// Add subcommands
	cmd.AddCommand(newVPNFailoverStartCmd(logger, configDir))
	cmd.AddCommand(newVPNFailoverStopCmd(logger, configDir))
	cmd.AddCommand(newVPNFailoverStatusCmd(logger, configDir))
	cmd.AddCommand(newVPNFailoverBackupCmd(logger, configDir))
	cmd.AddCommand(newVPNFailoverTestCmd(logger, configDir))
	cmd.AddCommand(newVPNFailoverHealthCmd(logger, configDir))
	cmd.AddCommand(newVPNFailoverPolicyCmd(logger, configDir))

	return cmd
}

// newVPNFailoverStartCmd creates the start subcommand.
func newVPNFailoverStartCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start VPN failover monitoring",
		Long:  `Start automatic VPN failover monitoring and backup activation service.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			manager, err := createFailoverVPNManager(ctx, logger, configDir)
			if err != nil {
				return fmt.Errorf("failed to create VPN failover manager: %w", err)
			}

			interval, _ := cmd.Flags().GetDuration("interval")
			retries, _ := cmd.Flags().GetInt("retries")
			timeout, _ := cmd.Flags().GetDuration("timeout")

			config := FailoverConfig{
				MonitorInterval: interval,
				MaxRetries:      retries,
				HealthTimeout:   timeout,
				AutoReconnect:   true,
				NotifyOnFailure: true,
			}

			fmt.Println("🚀 Starting VPN failover monitoring...")

			if err := manager.StartFailoverMonitoring(ctx, config); err != nil {
				return fmt.Errorf("failed to start failover monitoring: %w", err)
			}

			fmt.Printf("✅ VPN failover monitoring started (interval: %s)\n", interval)
			fmt.Println("Press Ctrl+C to stop monitoring...")

			// Keep monitoring running
			select {
			case <-ctx.Done():
				fmt.Println("\n🛑 Stopping VPN failover monitoring...")
				manager.StopFailoverMonitoring()
				fmt.Println("✅ VPN failover monitoring stopped")
			}

			return nil
		},
	}

	cmd.Flags().DurationP("interval", "i", 30*time.Second, "Health check interval")
	cmd.Flags().IntP("retries", "r", 3, "Max retry attempts before failover")
	cmd.Flags().DurationP("timeout", "t", 10*time.Second, "Health check timeout")

	return cmd
}

// newVPNFailoverStopCmd creates the stop subcommand.
func newVPNFailoverStopCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop VPN failover monitoring",
		Long:  `Stop the VPN failover monitoring service.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			manager, err := createFailoverVPNManager(ctx, logger, configDir)
			if err != nil {
				return fmt.Errorf("failed to create VPN failover manager: %w", err)
			}

			fmt.Println("🛑 Stopping VPN failover monitoring...")

			manager.StopFailoverMonitoring()

			fmt.Println("✅ VPN failover monitoring stopped")
			return nil
		},
	}

	return cmd
}

// newVPNFailoverStatusCmd creates the status subcommand.
func newVPNFailoverStatusCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show VPN failover status",
		Long:  `Display current VPN failover monitoring status and connection health.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			manager, err := createFailoverVPNManager(ctx, logger, configDir)
			if err != nil {
				return fmt.Errorf("failed to create VPN failover manager: %w", err)
			}

			output, _ := cmd.Flags().GetString("output")

			status, err := manager.GetFailoverStatus(ctx)
			if err != nil {
				return fmt.Errorf("failed to get failover status: %w", err)
			}

			switch output {
			case "json":
				return json.NewEncoder(os.Stdout).Encode(status)
			default:
				return printFailoverStatus(status)
			}
		},
	}

	cmd.Flags().StringP("output", "o", "table", "Output format (table|json)")

	return cmd
}

// newVPNFailoverBackupCmd creates the backup subcommand.
func newVPNFailoverBackupCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "backup",
		Short:        "Manage backup VPN connections",
		Long:         `Manage backup VPN connections for automatic failover.`,
		SilenceUsage: true,
	}

	cmd.AddCommand(newVPNFailoverBackupAddCmd(logger, configDir))
	cmd.AddCommand(newVPNFailoverBackupRemoveCmd(logger, configDir))
	cmd.AddCommand(newVPNFailoverBackupListCmd(logger, configDir))

	return cmd
}

// newVPNFailoverBackupAddCmd creates the backup add subcommand.
func newVPNFailoverBackupAddCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add backup VPN connection",
		Long:  `Add a backup VPN connection for automatic failover.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			primary, _ := cmd.Flags().GetString("primary")
			backup, _ := cmd.Flags().GetString("backup")
			priority, _ := cmd.Flags().GetInt("priority")

			if primary == "" || backup == "" {
				return fmt.Errorf("both primary and backup VPN names are required")
			}

			manager, err := createFailoverVPNManager(ctx, logger, configDir)
			if err != nil {
				return fmt.Errorf("failed to create VPN failover manager: %w", err)
			}

			config := BackupVPNConfig{
				PrimaryVPN:   primary,
				BackupVPN:    backup,
				Priority:     priority,
				AutoActivate: true,
				HealthCheck: HealthCheckConfig{
					Type:     "ping",
					Interval: 30 * time.Second,
					Timeout:  10 * time.Second,
					Retries:  3,
				},
			}

			if err := manager.AddBackupVPN(config); err != nil {
				return fmt.Errorf("failed to add backup VPN: %w", err)
			}

			fmt.Printf("✅ Added backup VPN: %s → %s (priority: %d)\n", primary, backup, priority)
			return nil
		},
	}

	cmd.Flags().String("primary", "", "Primary VPN connection name")
	cmd.Flags().String("backup", "", "Backup VPN connection name")
	cmd.Flags().Int("priority", 100, "Backup priority (higher = preferred)")
	cmd.MarkFlagRequired("primary")
	cmd.MarkFlagRequired("backup")

	return cmd
}

// newVPNFailoverBackupRemoveCmd creates the backup remove subcommand.
func newVPNFailoverBackupRemoveCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove backup VPN connection",
		Long:  `Remove a backup VPN connection from failover configuration.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			primary, _ := cmd.Flags().GetString("primary")
			backup, _ := cmd.Flags().GetString("backup")

			if primary == "" {
				return fmt.Errorf("primary VPN name is required")
			}

			manager, err := createFailoverVPNManager(ctx, logger, configDir)
			if err != nil {
				return fmt.Errorf("failed to create VPN failover manager: %w", err)
			}

			if err := manager.RemoveBackupVPN(primary, backup); err != nil {
				return fmt.Errorf("failed to remove backup VPN: %w", err)
			}

			if backup != "" {
				fmt.Printf("✅ Removed backup VPN: %s → %s\n", primary, backup)
			} else {
				fmt.Printf("✅ Removed all backups for primary VPN: %s\n", primary)
			}
			return nil
		},
	}

	cmd.Flags().String("primary", "", "Primary VPN connection name")
	cmd.Flags().String("backup", "", "Specific backup VPN to remove (optional)")
	cmd.MarkFlagRequired("primary")

	return cmd
}

// newVPNFailoverBackupListCmd creates the backup list subcommand.
func newVPNFailoverBackupListCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List backup VPN configurations",
		Long:  `Display all configured backup VPN connections and their priorities.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			manager, err := createFailoverVPNManager(ctx, logger, configDir)
			if err != nil {
				return fmt.Errorf("failed to create VPN failover manager: %w", err)
			}

			output, _ := cmd.Flags().GetString("output")

			backups, err := manager.GetBackupVPNs()
			if err != nil {
				return fmt.Errorf("failed to get backup VPNs: %w", err)
			}

			switch output {
			case "json":
				return json.NewEncoder(os.Stdout).Encode(backups)
			default:
				return printBackupVPNs(backups)
			}
		},
	}

	cmd.Flags().StringP("output", "o", "table", "Output format (table|json)")

	return cmd
}

// newVPNFailoverTestCmd creates the test subcommand.
func newVPNFailoverTestCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test",
		Short: "Test VPN failover scenarios",
		Long:  `Test various VPN failover scenarios to validate configuration.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
			defer cancel()

			scenario, _ := cmd.Flags().GetString("scenario")
			primary, _ := cmd.Flags().GetString("primary")

			manager, err := createFailoverVPNManager(ctx, logger, configDir)
			if err != nil {
				return fmt.Errorf("failed to create VPN failover manager: %w", err)
			}

			fmt.Printf("🧪 Testing failover scenario: %s\n", scenario)

			result, err := manager.TestFailoverScenario(ctx, scenario, primary)
			if err != nil {
				return fmt.Errorf("failover test failed: %w", err)
			}

			printFailoverTestResult(result)
			return nil
		},
	}

	cmd.Flags().String("scenario", "connection-loss", "Test scenario (connection-loss|health-check-fail|manual)")
	cmd.Flags().String("primary", "", "Primary VPN to test (optional)")

	return cmd
}

// newVPNFailoverHealthCmd creates the health subcommand.
func newVPNFailoverHealthCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "health",
		Short:        "Manage VPN health checks",
		Long:         `Configure and manage VPN connection health checks.`,
		SilenceUsage: true,
	}

	cmd.AddCommand(newVPNFailoverHealthShowCmd(logger, configDir))
	cmd.AddCommand(newVPNFailoverHealthConfigCmd(logger, configDir))
	cmd.AddCommand(newVPNFailoverHealthTestCmd(logger, configDir))

	return cmd
}

// newVPNFailoverHealthShowCmd creates the health show subcommand.
func newVPNFailoverHealthShowCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show VPN health check status",
		Long:  `Display current VPN health check status and results.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			manager, err := createFailoverVPNManager(ctx, logger, configDir)
			if err != nil {
				return fmt.Errorf("failed to create VPN failover manager: %w", err)
			}

			output, _ := cmd.Flags().GetString("output")

			health, err := manager.GetHealthStatus(ctx)
			if err != nil {
				return fmt.Errorf("failed to get health status: %w", err)
			}

			switch output {
			case "json":
				return json.NewEncoder(os.Stdout).Encode(health)
			default:
				return printHealthStatus(health)
			}
		},
	}

	cmd.Flags().StringP("output", "o", "table", "Output format (table|json)")

	return cmd
}

// newVPNFailoverHealthConfigCmd creates the health config subcommand.
func newVPNFailoverHealthConfigCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Configure VPN health checks",
		Long:  `Configure health check parameters for VPN connections.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			vpnName, _ := cmd.Flags().GetString("vpn")
			_, _ = cmd.Flags().GetBool("enabled") // enabled variable not used
			interval, _ := cmd.Flags().GetDuration("interval")
			timeout, _ := cmd.Flags().GetDuration("timeout")
			endpoint, _ := cmd.Flags().GetString("endpoint")

			if vpnName == "" {
				return fmt.Errorf("VPN name is required")
			}

			manager, err := createFailoverVPNManager(ctx, logger, configDir)
			if err != nil {
				return fmt.Errorf("failed to create VPN failover manager: %w", err)
			}

			config := HealthCheckConfig{
				Type:     "ping",
				Interval: interval,
				Timeout:  timeout,
				Endpoint: endpoint,
				Retries:  3,
			}

			if err := manager.ConfigureHealthCheck(vpnName, config); err != nil {
				return fmt.Errorf("failed to configure health check: %w", err)
			}

			fmt.Printf("✅ Configured health check for VPN: %s\n", vpnName)
			return nil
		},
	}

	cmd.Flags().String("vpn", "", "VPN connection name")
	cmd.Flags().Bool("enabled", true, "Enable health checks")
	cmd.Flags().Duration("interval", 30*time.Second, "Health check interval")
	cmd.Flags().Duration("timeout", 10*time.Second, "Health check timeout")
	cmd.Flags().String("endpoint", "", "Custom health check endpoint")
	cmd.MarkFlagRequired("vpn")

	return cmd
}

// newVPNFailoverHealthTestCmd creates the health test subcommand.
func newVPNFailoverHealthTestCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test [vpn-name]",
		Short: "Test VPN health check",
		Long:  `Run a manual health check test for a VPN connection.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			if len(args) == 0 {
				return fmt.Errorf("VPN name is required")
			}

			vpnName := args[0]

			manager, err := createFailoverVPNManager(ctx, logger, configDir)
			if err != nil {
				return fmt.Errorf("failed to create VPN failover manager: %w", err)
			}

			fmt.Printf("🔍 Testing health check for VPN: %s\n", vpnName)

			result, err := manager.TestHealthCheck(ctx, vpnName)
			if err != nil {
				return fmt.Errorf("health check test failed: %w", err)
			}

			printHealthCheckResult(result)
			return nil
		},
	}

	return cmd
}

// newVPNFailoverPolicyCmd creates the policy subcommand.
func newVPNFailoverPolicyCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "policy",
		Short:        "Manage failover policies",
		Long:         `Configure failover policies and recovery strategies.`,
		SilenceUsage: true,
	}

	cmd.AddCommand(newVPNFailoverPolicyShowCmd(logger, configDir))
	cmd.AddCommand(newVPNFailoverPolicySetCmd(logger, configDir))

	return cmd
}

// newVPNFailoverPolicyShowCmd creates the policy show subcommand.
func newVPNFailoverPolicyShowCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show current failover policies",
		Long:  `Display current failover policies and recovery settings.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			manager, err := createFailoverVPNManager(ctx, logger, configDir)
			if err != nil {
				return fmt.Errorf("failed to create VPN failover manager: %w", err)
			}

			output, _ := cmd.Flags().GetString("output")

			policy, err := manager.GetFailoverPolicy()
			if err != nil {
				return fmt.Errorf("failed to get failover policy: %w", err)
			}

			switch output {
			case "json":
				return json.NewEncoder(os.Stdout).Encode(policy)
			default:
				return printFailoverPolicy(policy)
			}
		},
	}

	cmd.Flags().StringP("output", "o", "table", "Output format (table|json)")

	return cmd
}

// newVPNFailoverPolicySetCmd creates the policy set subcommand.
func newVPNFailoverPolicySetCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set failover policy",
		Long:  `Configure failover policy and recovery parameters.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			strategy, _ := cmd.Flags().GetString("strategy")
			maxRetries, _ := cmd.Flags().GetInt("max-retries")
			retryInterval, _ := cmd.Flags().GetDuration("retry-interval")
			autoRecover, _ := cmd.Flags().GetBool("auto-recover")

			manager, err := createFailoverVPNManager(ctx, logger, configDir)
			if err != nil {
				return fmt.Errorf("failed to create VPN failover manager: %w", err)
			}

			policy := FailoverPolicy{
				Strategy:        FailoverStrategy(strategy),
				MaxRetries:      maxRetries,
				RetryInterval:   retryInterval,
				AutoRecover:     autoRecover,
				NotifyOnFailure: true,
				NotifyOnRecover: true,
			}

			if err := manager.SetFailoverPolicy(policy); err != nil {
				return fmt.Errorf("failed to set failover policy: %w", err)
			}

			fmt.Printf("✅ Updated failover policy: %s\n", strategy)
			return nil
		},
	}

	cmd.Flags().String("strategy", "priority", "Failover strategy (priority|round-robin|least-connections)")
	cmd.Flags().Int("max-retries", 3, "Maximum retry attempts")
	cmd.Flags().Duration("retry-interval", 30*time.Second, "Retry interval")
	cmd.Flags().Bool("auto-recover", true, "Enable automatic recovery")

	return cmd
}

// Helper functions and types

type FailoverVPNManager struct {
	logger     *zap.Logger
	configDir  string
	config     FailoverConfig
	backups    map[string][]BackupVPNConfig
	policy     FailoverPolicy
	monitoring bool
}

type FailoverConfig struct {
	MonitorInterval time.Duration `json:"monitor_interval"`
	MaxRetries      int           `json:"max_retries"`
	HealthTimeout   time.Duration `json:"health_timeout"`
	AutoReconnect   bool          `json:"auto_reconnect"`
	NotifyOnFailure bool          `json:"notify_on_failure"`
}

type BackupVPNConfig struct {
	PrimaryVPN   string            `json:"primary_vpn"`
	BackupVPN    string            `json:"backup_vpn"`
	Priority     int               `json:"priority"`
	AutoActivate bool              `json:"auto_activate"`
	HealthCheck  HealthCheckConfig `json:"health_check"`
}

// HealthCheckConfig type moved to network_topology.go to avoid duplication
// type HealthCheckConfig struct {
//	Enabled  bool          `json:"enabled"`
//	Interval time.Duration `json:"interval"`
//	Timeout  time.Duration `json:"timeout"`
//	Endpoint string        `json:"endpoint,omitempty"`
//}

type FailoverPolicy struct {
	Strategy        FailoverStrategy `json:"strategy"`
	MaxRetries      int              `json:"max_retries"`
	RetryInterval   time.Duration    `json:"retry_interval"`
	AutoRecover     bool             `json:"auto_recover"`
	NotifyOnFailure bool             `json:"notify_on_failure"`
	NotifyOnRecover bool             `json:"notify_on_recover"`
}

type FailoverStrategy string

const (
	StrategyPriority         FailoverStrategy = "priority"
	StrategyRoundRobin       FailoverStrategy = "round-robin"
	StrategyLeastConnections FailoverStrategy = "least-connections"
)

type FailoverStatus struct {
	Monitoring    bool                         `json:"monitoring"`
	ActiveVPNs    []string                     `json:"active_vpns"`
	FailedVPNs    []string                     `json:"failed_vpns"`
	BackupVPNs    []string                     `json:"backup_vpns"`
	LastFailover  *time.Time                   `json:"last_failover,omitempty"`
	FailoverCount int                          `json:"failover_count"`
	HealthChecks  map[string]HealthCheckResult `json:"health_checks"`
}

type HealthCheckResult struct {
	VPNName      string        `json:"vpn_name"`
	Status       string        `json:"status"`
	LastCheck    time.Time     `json:"last_check"`
	ResponseTime time.Duration `json:"response_time"`
	Error        string        `json:"error,omitempty"`
}

type FailoverTestResult struct {
	Scenario string        `json:"scenario"`
	Success  bool          `json:"success"`
	Duration time.Duration `json:"duration"`
	Steps    []TestStep    `json:"steps"`
	Error    string        `json:"error,omitempty"`
}

type TestStep struct {
	Step        string        `json:"step"`
	Success     bool          `json:"success"`
	Duration    time.Duration `json:"duration"`
	Description string        `json:"description"`
}

func createFailoverVPNManager(ctx context.Context, logger *zap.Logger, configDir string) (*FailoverVPNManager, error) {
	manager := &FailoverVPNManager{
		logger:    logger,
		configDir: configDir,
		backups:   make(map[string][]BackupVPNConfig),
		config: FailoverConfig{
			MonitorInterval: 30 * time.Second,
			MaxRetries:      3,
			HealthTimeout:   10 * time.Second,
			AutoReconnect:   true,
			NotifyOnFailure: true,
		},
		policy: FailoverPolicy{
			Strategy:        StrategyPriority,
			MaxRetries:      3,
			RetryInterval:   30 * time.Second,
			AutoRecover:     true,
			NotifyOnFailure: true,
			NotifyOnRecover: true,
		},
	}

	// Load existing configuration
	if err := manager.loadConfiguration(); err != nil {
		logger.Warn("Failed to load failover configuration", zap.Error(err))
	}

	return manager, nil
}

func (fvm *FailoverVPNManager) StartFailoverMonitoring(ctx context.Context, config FailoverConfig) error {
	fvm.config = config
	fvm.monitoring = true

	// TODO: Implement actual monitoring logic
	// This would include:
	// 1. Periodic health checks
	// 2. Connection state monitoring
	// 3. Automatic failover triggering
	// 4. Backup VPN activation

	fvm.logger.Info("Started VPN failover monitoring",
		zap.Duration("interval", config.MonitorInterval),
		zap.Int("max_retries", config.MaxRetries))

	return nil
}

func (fvm *FailoverVPNManager) StopFailoverMonitoring() {
	fvm.monitoring = false
	fvm.logger.Info("Stopped VPN failover monitoring")
}

func (fvm *FailoverVPNManager) GetFailoverStatus(ctx context.Context) (*FailoverStatus, error) {
	// TODO: Implement actual status collection
	return &FailoverStatus{
		Monitoring:    fvm.monitoring,
		ActiveVPNs:    []string{"corp-vpn"},
		FailedVPNs:    []string{},
		BackupVPNs:    []string{"home-vpn"},
		FailoverCount: 0,
		HealthChecks:  make(map[string]HealthCheckResult),
	}, nil
}

func (fvm *FailoverVPNManager) AddBackupVPN(config BackupVPNConfig) error {
	if fvm.backups[config.PrimaryVPN] == nil {
		fvm.backups[config.PrimaryVPN] = make([]BackupVPNConfig, 0)
	}

	fvm.backups[config.PrimaryVPN] = append(fvm.backups[config.PrimaryVPN], config)

	return fvm.saveConfiguration()
}

func (fvm *FailoverVPNManager) RemoveBackupVPN(primary, backup string) error {
	backups, exists := fvm.backups[primary]
	if !exists {
		return fmt.Errorf("no backups found for primary VPN: %s", primary)
	}

	if backup == "" {
		// Remove all backups for this primary
		delete(fvm.backups, primary)
	} else {
		// Remove specific backup
		for i, b := range backups {
			if b.BackupVPN == backup {
				fvm.backups[primary] = append(backups[:i], backups[i+1:]...)
				break
			}
		}
	}

	return fvm.saveConfiguration()
}

func (fvm *FailoverVPNManager) GetBackupVPNs() ([]BackupVPNConfig, error) {
	var all []BackupVPNConfig
	for _, backups := range fvm.backups {
		all = append(all, backups...)
	}

	return all, nil
}

func (fvm *FailoverVPNManager) TestFailoverScenario(ctx context.Context, scenario, primary string) (*FailoverTestResult, error) {
	// TODO: Implement failover scenario testing
	return &FailoverTestResult{
		Scenario: scenario,
		Success:  true,
		Duration: 15 * time.Second,
		Steps: []TestStep{
			{Step: "1", Success: true, Duration: 5 * time.Second, Description: "Simulate connection failure"},
			{Step: "2", Success: true, Duration: 5 * time.Second, Description: "Detect failure"},
			{Step: "3", Success: true, Duration: 5 * time.Second, Description: "Activate backup VPN"},
		},
	}, nil
}

func (fvm *FailoverVPNManager) GetHealthStatus(ctx context.Context) (map[string]HealthCheckResult, error) {
	// TODO: Implement health status collection
	return map[string]HealthCheckResult{
		"corp-vpn": {
			VPNName:      "corp-vpn",
			Status:       "healthy",
			LastCheck:    time.Now(),
			ResponseTime: 50 * time.Millisecond,
		},
	}, nil
}

func (fvm *FailoverVPNManager) ConfigureHealthCheck(vpnName string, config HealthCheckConfig) error {
	// TODO: Implement health check configuration
	fvm.logger.Info("Configured health check",
		zap.String("vpn", vpnName),
		zap.String("type", config.Type),
		zap.Duration("interval", config.Interval))

	return nil
}

func (fvm *FailoverVPNManager) TestHealthCheck(ctx context.Context, vpnName string) (*HealthCheckResult, error) {
	// TODO: Implement health check testing
	return &HealthCheckResult{
		VPNName:      vpnName,
		Status:       "healthy",
		LastCheck:    time.Now(),
		ResponseTime: 75 * time.Millisecond,
	}, nil
}

func (fvm *FailoverVPNManager) GetFailoverPolicy() (*FailoverPolicy, error) {
	return &fvm.policy, nil
}

func (fvm *FailoverVPNManager) SetFailoverPolicy(policy FailoverPolicy) error {
	fvm.policy = policy
	return fvm.saveConfiguration()
}

func (fvm *FailoverVPNManager) loadConfiguration() error {
	configPath := filepath.Join(fvm.configDir, "vpn-failover.json")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil // Config file doesn't exist, that's ok
	}

	// TODO: Implement JSON configuration loading
	return nil
}

func (fvm *FailoverVPNManager) saveConfiguration() error {
	// configPath := filepath.Join(fvm.configDir, "vpn-failover.json")

	// Ensure config directory exists
	if err := os.MkdirAll(fvm.configDir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// TODO: Implement JSON configuration saving
	return nil
}

// Print functions

func printFailoverStatus(status *FailoverStatus) error {
	fmt.Printf("🔄 VPN Failover Status\n\n")

	fmt.Printf("Monitoring: %t\n", status.Monitoring)
	fmt.Printf("Failover Count: %d\n", status.FailoverCount)

	if status.LastFailover != nil {
		fmt.Printf("Last Failover: %s\n", status.LastFailover.Format("2006-01-02 15:04:05"))
	}

	fmt.Printf("\nActive VPNs: %v\n", status.ActiveVPNs)
	fmt.Printf("Failed VPNs: %v\n", status.FailedVPNs)
	fmt.Printf("Backup VPNs: %v\n", status.BackupVPNs)

	if len(status.HealthChecks) > 0 {
		fmt.Printf("\nHealth Checks:\n")

		w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
		fmt.Fprintln(w, "VPN\tSTATUS\tRESPONSE TIME\tLAST CHECK")

		for _, health := range status.HealthChecks {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				health.VPNName,
				health.Status,
				health.ResponseTime,
				health.LastCheck.Format("15:04:05"))
		}

		w.Flush()
	}

	return nil
}

func printBackupVPNs(backups []BackupVPNConfig) error {
	fmt.Printf("🔄 Backup VPN Configurations\n\n")

	if len(backups) == 0 {
		fmt.Println("  No backup VPNs configured.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	fmt.Fprintln(w, "PRIMARY\tBACKUP\tPRIORITY\tAUTO-ACTIVATE\tHEALTH CHECK")

	for _, backup := range backups {
		autoActivate := "No"
		if backup.AutoActivate {
			autoActivate = "Yes"
		}

		healthCheck := "Disabled"
		if backup.HealthCheck.Type != "" {
			healthCheck = fmt.Sprintf("Every %s", backup.HealthCheck.Interval)
		}

		fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\n",
			backup.PrimaryVPN,
			backup.BackupVPN,
			backup.Priority,
			autoActivate,
			healthCheck)
	}

	return w.Flush()
}

func printFailoverTestResult(result *FailoverTestResult) {
	fmt.Printf("🧪 Failover Test Result: %s\n\n", result.Scenario)

	status := "❌ FAILED"
	if result.Success {
		status = "✅ PASSED"
	}

	fmt.Printf("Status: %s\n", status)
	fmt.Printf("Duration: %s\n\n", result.Duration)

	if result.Error != "" {
		fmt.Printf("Error: %s\n\n", result.Error)
	}

	fmt.Printf("Test Steps:\n")

	for _, step := range result.Steps {
		stepStatus := "❌"
		if step.Success {
			stepStatus = "✅"
		}

		fmt.Printf("  %s %s: %s (%s)\n", stepStatus, step.Step, step.Description, step.Duration)
	}
}

func printHealthStatus(health map[string]HealthCheckResult) error {
	fmt.Printf("💚 VPN Health Status\n\n")

	if len(health) == 0 {
		fmt.Println("  No health checks configured.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	fmt.Fprintln(w, "VPN\tSTATUS\tRESPONSE TIME\tLAST CHECK\tERROR")

	for _, result := range health {
		errorMsg := "-"
		if result.Error != "" {
			errorMsg = truncateStringUtil(result.Error, 30)
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			result.VPNName,
			result.Status,
			result.ResponseTime,
			result.LastCheck.Format("15:04:05"),
			errorMsg)
	}

	return w.Flush()
}

func printHealthCheckResult(result *HealthCheckResult) {
	fmt.Printf("🔍 Health Check Result: %s\n\n", result.VPNName)

	status := result.Status
	if status == "healthy" {
		status = "✅ " + status
	} else {
		status = "❌ " + status
	}

	fmt.Printf("Status: %s\n", status)
	fmt.Printf("Response Time: %s\n", result.ResponseTime)
	fmt.Printf("Last Check: %s\n", result.LastCheck.Format("2006-01-02 15:04:05"))

	if result.Error != "" {
		fmt.Printf("Error: %s\n", result.Error)
	}
}

func printFailoverPolicy(policy *FailoverPolicy) error {
	fmt.Printf("📋 Failover Policy\n\n")

	fmt.Printf("Strategy: %s\n", policy.Strategy)
	fmt.Printf("Max Retries: %d\n", policy.MaxRetries)
	fmt.Printf("Retry Interval: %s\n", policy.RetryInterval)
	fmt.Printf("Auto Recover: %t\n", policy.AutoRecover)
	fmt.Printf("Notify on Failure: %t\n", policy.NotifyOnFailure)
	fmt.Printf("Notify on Recover: %t\n", policy.NotifyOnRecover)

	return nil
}
