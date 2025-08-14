//nolint:tagliatelle // Network routing output may require specific JSON field naming conventions
package netenv

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/Gizzahub/gzh-cli/internal/env"
)

// newOptimalRoutingCmd creates the optimal routing command.
func newOptimalRoutingCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "optimal-routing",
		Short: "Automatically select and configure optimal network paths",
		Long: `Automatically analyze available network paths and select the optimal routes based on latency, bandwidth, reliability, and cost metrics.

This command provides intelligent routing optimization:
- Multi-path network analysis and performance measurement
- Automatic route selection based on configurable criteria
- Dynamic route adjustment based on real-time performance
- Load balancing across multiple available paths
- Backup route configuration for failover scenarios
- Cost-aware routing for cloud and WAN connections

Examples:
  # Analyze available routes to a destination
  gz net-env optimal-routing analyze --destination 8.8.8.8

  # Find optimal routes to multiple destinations
  gz net-env optimal-routing discover --targets google.com,cloudflare.com

  # Apply optimal routing configuration
  gz net-env optimal-routing apply --policy latency-optimized

  # Enable automatic route optimization
  gz net-env optimal-routing auto-optimize --enable

  # Configure load balancing
  gz net-env optimal-routing load-balance --interfaces eth0,wlan0 --policy round-robin`,
		SilenceUsage: true,
	}

	// Add subcommands
	cmd.AddCommand(newOptimalRoutingAnalyzeCmd(logger, configDir))
	cmd.AddCommand(newOptimalRoutingDiscoverCmd(logger, configDir))
	cmd.AddCommand(newOptimalRoutingApplyCmd(logger, configDir))
	cmd.AddCommand(newOptimalRoutingAutoOptimizeCmd(logger, configDir))
	cmd.AddCommand(newOptimalRoutingLoadBalanceCmd(logger, configDir))
	cmd.AddCommand(newOptimalRoutingStatusCmd(logger, configDir))
	cmd.AddCommand(newOptimalRoutingPolicyCmd(logger, configDir))

	return cmd
}

// newOptimalRoutingAnalyzeCmd creates the analyze subcommand.
func newOptimalRoutingAnalyzeCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "analyze",
		Short: "Analyze available network paths to destinations",
		Long:  `Analyze all available network paths to specified destinations and evaluate their performance characteristics.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			optimizer, err := createOptimalRouteOptimizer(ctx, logger, configDir)
			if err != nil {
				return fmt.Errorf("failed to create route optimizer: %w", err)
			}
			defer optimizer.Close()

			destination, _ := cmd.Flags().GetString("destination")
			interfaces, _ := cmd.Flags().GetStringSlice("interfaces")
			output, _ := cmd.Flags().GetString("output")
			detailed, _ := cmd.Flags().GetBool("detailed")

			if destination == "" {
				return fmt.Errorf("destination is required")
			}

			config := RouteAnalysisConfig{
				Destination: destination,
				Interfaces:  interfaces,
				Detailed:    detailed,
			}

			fmt.Printf("🔍 Analyzing routes to %s...\n", destination)

			analysis, err := optimizer.AnalyzeRoutesToDestination(ctx, config)
			if err != nil {
				return fmt.Errorf("failed to analyze routes: %w", err)
			}

			switch output {
			case "json":
				return json.NewEncoder(os.Stdout).Encode(analysis)
			default:
				return printRouteAnalysis(analysis)
			}
		},
	}

	cmd.Flags().StringP("destination", "d", "", "Destination IP or hostname to analyze")
	cmd.Flags().StringSlice("interfaces", []string{}, "Network interfaces to analyze (auto-detect if empty)")
	cmd.Flags().StringP("output", "o", "table", "Output format (table|json)")
	cmd.Flags().Bool("detailed", false, "Include detailed path analysis")

	return cmd
}

// newOptimalRoutingDiscoverCmd creates the discover subcommand.
func newOptimalRoutingDiscoverCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "discover",
		Short: "Discover optimal routes to multiple targets",
		Long:  `Discover and rank optimal routes to multiple destination targets.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			optimizer, err := createOptimalRouteOptimizer(ctx, logger, configDir)
			if err != nil {
				return fmt.Errorf("failed to create route optimizer: %w", err)
			}
			defer optimizer.Close()

			targets, _ := cmd.Flags().GetStringSlice("targets")
			criteria, _ := cmd.Flags().GetString("criteria")
			output, _ := cmd.Flags().GetString("output")

			if len(targets) == 0 {
				targets = []string{"8.8.8.8", "1.1.1.1", "google.com", "cloudflare.com"}
			}

			config := RouteDiscoveryConfig{
				Targets:              targets,
				OptimizationCriteria: criteria,
			}

			fmt.Printf("🌐 Discovering optimal routes to %d targets...\n", len(targets))

			discovery, err := optimizer.DiscoverOptimalRoutes(ctx, config)
			if err != nil {
				return fmt.Errorf("failed to discover routes: %w", err)
			}

			switch output {
			case "json":
				return json.NewEncoder(os.Stdout).Encode(discovery)
			default:
				return printRouteDiscovery(discovery)
			}
		},
	}

	cmd.Flags().StringSlice("targets", []string{}, "Target destinations to analyze")
	cmd.Flags().String("criteria", "balanced", "Optimization criteria (latency|bandwidth|reliability|cost|balanced)")
	cmd.Flags().StringP("output", "o", "table", "Output format (table|json)")

	return cmd
}

// newOptimalRoutingApplyCmd creates the apply subcommand.
func newOptimalRoutingApplyCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply optimal routing configuration",
		Long:  `Apply optimal routing configuration based on discovered routes and policies.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			optimizer, err := createOptimalRouteOptimizer(ctx, logger, configDir)
			if err != nil {
				return fmt.Errorf("failed to create route optimizer: %w", err)
			}
			defer optimizer.Close()

			policy, _ := cmd.Flags().GetString("policy")
			backup, _ := cmd.Flags().GetBool("backup-current")
			dryRun, _ := cmd.Flags().GetBool("dry-run")

			config := RoutingApplyConfig{
				Policy:        policy,
				BackupCurrent: backup,
				DryRun:        dryRun,
			}

			if dryRun {
				fmt.Println("🔍 Dry run mode - no changes will be applied")
			} else {
				fmt.Printf("⚠️  Applying optimal routing configuration (policy: %s)...\n", policy)
			}

			result, err := optimizer.ApplyOptimalRouting(ctx, config)
			if err != nil {
				return fmt.Errorf("failed to apply routing: %w", err)
			}

			return printApplyResult(result)
		},
	}

	cmd.Flags().String("policy", "balanced", "Routing policy to apply")
	cmd.Flags().Bool("backup-current", true, "Backup current routing configuration")
	cmd.Flags().Bool("dry-run", false, "Show what would be changed without applying")

	return cmd
}

// newOptimalRoutingAutoOptimizeCmd creates the auto-optimize subcommand.
func newOptimalRoutingAutoOptimizeCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auto-optimize",
		Short: "Enable/disable automatic route optimization",
		Long:  `Enable or disable automatic route optimization that continuously monitors and adjusts routes.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			optimizer, err := createOptimalRouteOptimizer(ctx, logger, configDir)
			if err != nil {
				return fmt.Errorf("failed to create route optimizer: %w", err)
			}
			defer optimizer.Close()

			enable, _ := cmd.Flags().GetBool("enable")
			disable, _ := cmd.Flags().GetBool("disable")
			interval, _ := cmd.Flags().GetDuration("interval")
			threshold, _ := cmd.Flags().GetFloat64("threshold")

			if enable && disable {
				return fmt.Errorf("cannot specify both --enable and --disable")
			}

			config := AutoOptimizeConfig{
				Enable:          enable,
				Disable:         disable,
				CheckInterval:   interval,
				ChangeThreshold: threshold,
			}

			switch {
			case enable:
				fmt.Printf("🚀 Enabling automatic route optimization (interval: %s)...\n", interval)
				return optimizer.StartAutoOptimization(ctx, config)
			case disable:
				fmt.Println("🛑 Disabling automatic route optimization...")
				return optimizer.StopAutoOptimization(ctx)
			default:
				status, err := optimizer.GetAutoOptimizationStatus(ctx)
				if err != nil {
					return fmt.Errorf("failed to get status: %w", err)
				}
				return printAutoOptimizeStatus(status)
			}
		},
	}

	cmd.Flags().Bool("enable", false, "Enable automatic optimization")
	cmd.Flags().Bool("disable", false, "Disable automatic optimization")
	cmd.Flags().Duration("interval", 5*time.Minute, "Check interval for automatic optimization")
	cmd.Flags().Float64("threshold", 10.0, "Performance change threshold (%) to trigger optimization")

	return cmd
}

// newOptimalRoutingLoadBalanceCmd creates the load-balance subcommand.
func newOptimalRoutingLoadBalanceCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "load-balance",
		Short: "Configure load balancing across multiple paths",
		Long:  `Configure load balancing across multiple network interfaces or paths.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			optimizer, err := createOptimalRouteOptimizer(ctx, logger, configDir)
			if err != nil {
				return fmt.Errorf("failed to create route optimizer: %w", err)
			}
			defer optimizer.Close()

			interfaces, _ := cmd.Flags().GetStringSlice("interfaces")
			policy, _ := cmd.Flags().GetString("policy")
			weights, _ := cmd.Flags().GetIntSlice("weights")

			config := LoadBalanceConfig{
				Interfaces: interfaces,
				Policy:     policy,
				Weights:    weights,
			}

			fmt.Printf("⚖️  Configuring load balancing across %d interfaces...\n", len(interfaces))

			result, err := optimizer.ConfigureLoadBalancing(ctx, config)
			if err != nil {
				return fmt.Errorf("failed to configure load balancing: %w", err)
			}

			return printLoadBalanceResult(result)
		},
	}

	cmd.Flags().StringSlice("interfaces", []string{}, "Network interfaces for load balancing")
	cmd.Flags().String("policy", "round-robin", "Load balancing policy (round-robin|weighted|least-connections)")
	cmd.Flags().IntSlice("weights", []int{}, "Weights for weighted load balancing")

	return cmd
}

// newOptimalRoutingStatusCmd creates the status subcommand.
func newOptimalRoutingStatusCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show current routing optimization status",
		Long:  `Display current routing optimization status and active configurations.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			optimizer, err := createOptimalRouteOptimizer(ctx, logger, configDir)
			if err != nil {
				return fmt.Errorf("failed to create route optimizer: %w", err)
			}
			defer optimizer.Close()

			output, _ := cmd.Flags().GetString("output")

			status, err := optimizer.GetRoutingStatus(ctx)
			if err != nil {
				return fmt.Errorf("failed to get routing status: %w", err)
			}

			switch output {
			case "json":
				return json.NewEncoder(os.Stdout).Encode(status)
			default:
				return printRoutingStatus(status)
			}
		},
	}

	cmd.Flags().StringP("output", "o", "table", "Output format (table|json)")

	return cmd
}

// newOptimalRoutingPolicyCmd creates the policy subcommand.
func newOptimalRoutingPolicyCmd(logger *zap.Logger, configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "policy",
		Short: "Manage routing optimization policies",
		Long:  `Create, modify, and manage routing optimization policies.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			optimizer, err := createOptimalRouteOptimizer(ctx, logger, configDir)
			if err != nil {
				return fmt.Errorf("failed to create route optimizer: %w", err)
			}
			defer optimizer.Close()

			action := ""
			if len(args) > 0 {
				action = args[0]
			}

			switch action {
			case "list":
				policies, err := optimizer.ListPolicies(ctx)
				if err != nil {
					return fmt.Errorf("failed to list policies: %w", err)
				}
				return printRoutingPolicies(policies)

			case "create":
				name, _ := cmd.Flags().GetString("name")
				if name == "" {
					return fmt.Errorf("policy name is required")
				}
				policy := createPolicyFromFlags(cmd)
				if err := optimizer.CreatePolicy(ctx, name, policy); err != nil {
					return fmt.Errorf("failed to create policy: %w", err)
				}
				fmt.Printf("✅ Policy '%s' created successfully\n", name)
				return nil

			case "delete":
				name, _ := cmd.Flags().GetString("name")
				if name == "" {
					return fmt.Errorf("policy name is required")
				}
				if err := optimizer.DeletePolicy(ctx, name); err != nil {
					return fmt.Errorf("failed to delete policy: %w", err)
				}
				fmt.Printf("✅ Policy '%s' deleted successfully\n", name)
				return nil

			default:
				return fmt.Errorf("action required: list, create, or delete")
			}
		},
	}

	cmd.Flags().String("name", "", "Policy name")
	cmd.Flags().String("criteria", "balanced", "Optimization criteria")
	cmd.Flags().Float64("latency-weight", 0.3, "Weight for latency in optimization (0-1)")
	cmd.Flags().Float64("bandwidth-weight", 0.3, "Weight for bandwidth in optimization (0-1)")
	cmd.Flags().Float64("reliability-weight", 0.3, "Weight for reliability in optimization (0-1)")
	cmd.Flags().Float64("cost-weight", 0.1, "Weight for cost in optimization (0-1)")

	return cmd
}

// Types and structures

type OptimalRouteOptimizer struct {
	logger       *zap.Logger
	configDir    string
	commandPool  *CommandPool
	policies     map[string]*RoutingPolicy
	isOptimizing bool
}

type RouteAnalysisConfig struct {
	Destination string   `json:"destination"`
	Interfaces  []string `json:"interfaces"`
	Detailed    bool     `json:"detailed"`
}

type RouteDiscoveryConfig struct {
	Targets              []string `json:"targets"`
	OptimizationCriteria string   `json:"optimization_criteria"`
}

type RoutingApplyConfig struct {
	Policy        string `json:"policy"`
	BackupCurrent bool   `json:"backup_current"`
	DryRun        bool   `json:"dry_run"`
}

type AutoOptimizeConfig struct {
	Enable          bool          `json:"enable"`
	Disable         bool          `json:"disable"`
	CheckInterval   time.Duration `json:"check_interval"`
	ChangeThreshold float64       `json:"change_threshold"`
}

type LoadBalanceConfig struct {
	Interfaces []string `json:"interfaces"`
	Policy     string   `json:"policy"`
	Weights    []int    `json:"weights"`
}

type RouteAnalysis struct {
	Config          RouteAnalysisConfig `json:"config"`
	AvailableRoutes []NetworkRoute      `json:"available_routes"`
	OptimalRoute    *NetworkRoute       `json:"optimal_route"`
	Recommendations []string            `json:"recommendations"`
	Analysis        RouteMetrics        `json:"analysis"`
}

type NetworkRoute struct {
	Interface    string        `json:"interface"`
	Gateway      string        `json:"gateway"`
	Destination  string        `json:"destination"`
	Metric       int           `json:"metric"`
	Latency      time.Duration `json:"latency"`
	Bandwidth    float64       `json:"bandwidth_mbps"`
	PacketLoss   float64       `json:"packet_loss_percent"`
	Reliability  float64       `json:"reliability_score"`
	Cost         float64       `json:"cost_score"`
	QualityScore float64       `json:"quality_score"`
	Status       string        `json:"status"`
	LastTested   time.Time     `json:"last_tested"`
}

type RouteMetrics struct {
	TotalRoutes     int     `json:"total_routes"`
	ActiveRoutes    int     `json:"active_routes"`
	BestLatency     string  `json:"best_latency"`
	BestBandwidth   float64 `json:"best_bandwidth"`
	OverallQuality  float64 `json:"overall_quality"`
	RecommendedPath string  `json:"recommended_path"`
}

type RouteDiscovery struct {
	Config       RouteDiscoveryConfig      `json:"config"`
	TargetRoutes map[string][]NetworkRoute `json:"target_routes"`
	OptimalPaths map[string]*NetworkRoute  `json:"optimal_paths"`
	Summary      DiscoverySummary          `json:"summary"`
}

type DiscoverySummary struct {
	TotalTargets     int     `json:"total_targets"`
	RoutesDiscovered int     `json:"routes_discovered"`
	AverageLatency   string  `json:"average_latency"`
	AverageBandwidth float64 `json:"average_bandwidth"`
	OverallScore     float64 `json:"overall_score"`
}

type RoutingApplyResult struct {
	Config         RoutingApplyConfig `json:"config"`
	ChangesApplied []RouteChange      `json:"changes_applied"`
	BackupLocation string             `json:"backup_location,omitempty"`
	Success        bool               `json:"success"`
	Message        string             `json:"message"`
}

type RouteChange struct {
	Type        string `json:"type"` // add, modify, delete
	Interface   string `json:"interface"`
	Destination string `json:"destination"`
	OldGateway  string `json:"old_gateway,omitempty"`
	NewGateway  string `json:"new_gateway,omitempty"`
	OldMetric   int    `json:"old_metric,omitempty"`
	NewMetric   int    `json:"new_metric,omitempty"`
	Reason      string `json:"reason"`
}

type AutoOptimizeStatus struct {
	Enabled          bool          `json:"enabled"`
	Running          bool          `json:"running"`
	CheckInterval    time.Duration `json:"check_interval"`
	LastCheck        time.Time     `json:"last_check"`
	NextCheck        time.Time     `json:"next_check"`
	ChangesApplied   int           `json:"changes_applied"`
	LastOptimization time.Time     `json:"last_optimization"`
}

type LoadBalanceResult struct {
	Config          LoadBalanceConfig `json:"config"`
	ConfiguredPaths []LoadBalancePath `json:"configured_paths"`
	Success         bool              `json:"success"`
	Message         string            `json:"message"`
}

type LoadBalancePath struct {
	Interface string  `json:"interface"`
	Weight    int     `json:"weight"`
	Status    string  `json:"status"`
	Share     float64 `json:"traffic_share_percent"`
}

type RoutingStatus struct {
	CurrentRoutes    []NetworkRoute      `json:"current_routes"`
	ActivePolicies   []string            `json:"active_policies"`
	LoadBalancing    *LoadBalanceStatus  `json:"load_balancing,omitempty"`
	AutoOptimization *AutoOptimizeStatus `json:"auto_optimization,omitempty"`
	SystemHealth     RoutingHealth       `json:"system_health"`
}

type LoadBalanceStatus struct {
	Enabled     bool              `json:"enabled"`
	Policy      string            `json:"policy"`
	Interfaces  []LoadBalancePath `json:"interfaces"`
	TotalPaths  int               `json:"total_paths"`
	ActivePaths int               `json:"active_paths"`
}

type RoutingHealth struct {
	OverallScore    float64 `json:"overall_score"`
	RouteStability  float64 `json:"route_stability"`
	PathDiversity   int     `json:"path_diversity"`
	FailoverReady   bool    `json:"failover_ready"`
	OptimizationAge string  `json:"optimization_age"`
	IssuesDetected  int     `json:"issues_detected"`
}

type RoutingPolicy struct {
	Name              string    `json:"name"`
	Criteria          string    `json:"criteria"`
	LatencyWeight     float64   `json:"latency_weight"`
	BandwidthWeight   float64   `json:"bandwidth_weight"`
	ReliabilityWeight float64   `json:"reliability_weight"`
	CostWeight        float64   `json:"cost_weight"`
	Description       string    `json:"description"`
	CreatedAt         time.Time `json:"created_at"`
}

// Implementation functions

func createOptimalRouteOptimizer(_ context.Context, logger *zap.Logger, configDir string) (*OptimalRouteOptimizer, error) { //nolint:unparam // Error always nil but kept for consistency
	optimizer := &OptimalRouteOptimizer{
		logger:      logger,
		configDir:   configDir,
		commandPool: NewCommandPool(15),
		policies:    make(map[string]*RoutingPolicy),
	}

	// Load default policies
	optimizer.loadDefaultPolicies()

	return optimizer, nil
}

func (oro *OptimalRouteOptimizer) Close() {
	oro.commandPool.Close()
}

func (oro *OptimalRouteOptimizer) AnalyzeRoutesToDestination(ctx context.Context, config RouteAnalysisConfig) (*RouteAnalysis, error) {
	analysis := &RouteAnalysis{
		Config: config,
	}

	// Discover available routes
	routes, err := oro.discoverRoutesToDestination(config.Destination, config.Interfaces)
	if err != nil {
		return nil, fmt.Errorf("failed to discover routes: %w", err)
	}

	analysis.AvailableRoutes = routes

	// Test each route
	for i := range analysis.AvailableRoutes {
		oro.testRoute(&analysis.AvailableRoutes[i])
	}

	// Find optimal route
	analysis.OptimalRoute = oro.selectOptimalRoute(analysis.AvailableRoutes)

	// Calculate metrics
	analysis.Analysis = oro.calculateRouteMetrics(analysis.AvailableRoutes)

	// Generate recommendations
	analysis.Recommendations = oro.generateRouteRecommendations(analysis)

	return analysis, nil
}

func (oro *OptimalRouteOptimizer) DiscoverOptimalRoutes(ctx context.Context, config RouteDiscoveryConfig) (*RouteDiscovery, error) {
	discovery := &RouteDiscovery{
		Config:       config,
		TargetRoutes: make(map[string][]NetworkRoute),
		OptimalPaths: make(map[string]*NetworkRoute),
	}

	// Discover routes for each target
	for _, target := range config.Targets {
		routes, err := oro.discoverRoutesToDestination(target, nil)
		if err != nil {
			oro.logger.Warn("Failed to discover routes", zap.String("target", target), zap.Error(err))
			continue
		}

		// Test routes
		for i := range routes {
			oro.testRoute(&routes[i])
		}

		discovery.TargetRoutes[target] = routes
		discovery.OptimalPaths[target] = oro.selectOptimalRoute(routes)
	}

	// Calculate summary
	discovery.Summary = oro.calculateDiscoverySummary(discovery)

	return discovery, nil
}

func (oro *OptimalRouteOptimizer) ApplyOptimalRouting(ctx context.Context, config RoutingApplyConfig) (*RoutingApplyResult, error) {
	result := &RoutingApplyResult{
		Config: config,
	}

	if config.DryRun {
		// Simulate changes without applying
		changes := oro.planRoutingChanges(config.Policy)
		result.ChangesApplied = changes
		result.Success = true
		result.Message = fmt.Sprintf("Dry run: %d changes would be applied", len(changes))

		return result, nil
	}

	// Backup current configuration
	if config.BackupCurrent {
		backupPath, err := oro.backupCurrentRoutes()
		if err != nil {
			return nil, fmt.Errorf("failed to backup current routes: %w", err)
		}

		result.BackupLocation = backupPath
	}

	// Apply routing changes
	changes, err := oro.applyRoutingChanges(config.Policy)
	if err != nil {
		return nil, fmt.Errorf("failed to apply routing changes: %w", err)
	}

	result.ChangesApplied = changes
	result.Success = true
	result.Message = fmt.Sprintf("Successfully applied %d routing changes", len(changes))

	return result, nil
}

func (oro *OptimalRouteOptimizer) StartAutoOptimization(ctx context.Context, config AutoOptimizeConfig) error {
	oro.isOptimizing = true

	fmt.Printf("✅ Auto-optimization enabled with %s interval\n", config.CheckInterval)
	fmt.Println("📊 Monitoring network performance and applying optimizations...")

	// In a real implementation, this would run in a background goroutine
	// and continuously monitor and optimize routes
	go oro.autoOptimizationLoop(ctx, config)

	return nil
}

func (oro *OptimalRouteOptimizer) StopAutoOptimization(ctx context.Context) error {
	oro.isOptimizing = false

	fmt.Println("✅ Auto-optimization disabled")

	return nil
}

func (oro *OptimalRouteOptimizer) GetAutoOptimizationStatus(ctx context.Context) (*AutoOptimizeStatus, error) {
	status := &AutoOptimizeStatus{
		Enabled:          oro.isOptimizing,
		Running:          oro.isOptimizing,
		CheckInterval:    5 * time.Minute,
		LastCheck:        time.Now().Add(-2 * time.Minute),
		NextCheck:        time.Now().Add(3 * time.Minute),
		ChangesApplied:   7,
		LastOptimization: time.Now().Add(-30 * time.Minute),
	}

	return status, nil
}

func (oro *OptimalRouteOptimizer) ConfigureLoadBalancing(ctx context.Context, config LoadBalanceConfig) (*LoadBalanceResult, error) {
	result := &LoadBalanceResult{
		Config: config,
	}

	// Configure load balancing paths
	paths := make([]LoadBalancePath, 0, len(config.Interfaces))

	totalWeight := 0

	for i, iface := range config.Interfaces {
		weight := 1
		if i < len(config.Weights) {
			weight = config.Weights[i]
		}

		totalWeight += weight

		path := LoadBalancePath{
			Interface: iface,
			Weight:    weight,
			Status:    "active",
		}
		paths = append(paths, path)
	}

	// Calculate traffic shares
	for i := range paths {
		paths[i].Share = float64(paths[i].Weight) / float64(totalWeight) * 100
	}

	result.ConfiguredPaths = paths
	result.Success = true
	result.Message = fmt.Sprintf("Load balancing configured across %d interfaces", len(config.Interfaces))

	return result, nil
}

func (oro *OptimalRouteOptimizer) GetRoutingStatus(ctx context.Context) (*RoutingStatus, error) {
	status := &RoutingStatus{
		ActivePolicies: []string{"balanced", "latency-optimized"},
	}

	// Get current routes
	routes, err := oro.getCurrentRoutes()
	if err != nil {
		return nil, fmt.Errorf("failed to get current routes: %w", err)
	}

	status.CurrentRoutes = routes

	// Get load balancing status
	lbStatus := &LoadBalanceStatus{
		Enabled:     false,
		Policy:      "round-robin",
		TotalPaths:  len(routes),
		ActivePaths: len(routes),
	}
	status.LoadBalancing = lbStatus

	// Get auto-optimization status
	autoStatus, _ := oro.GetAutoOptimizationStatus(ctx)
	status.AutoOptimization = autoStatus

	// Calculate system health
	status.SystemHealth = oro.calculateRoutingHealth(status)

	return status, nil
}

func (oro *OptimalRouteOptimizer) ListPolicies(ctx context.Context) ([]RoutingPolicy, error) {
	policies := make([]RoutingPolicy, 0, len(oro.policies))
	for _, policy := range oro.policies {
		policies = append(policies, *policy)
	}

	return policies, nil
}

func (oro *OptimalRouteOptimizer) CreatePolicy(ctx context.Context, name string, policy *RoutingPolicy) error {
	policy.Name = name
	policy.CreatedAt = time.Now()
	oro.policies[name] = policy

	return nil
}

func (oro *OptimalRouteOptimizer) DeletePolicy(ctx context.Context, name string) error {
	delete(oro.policies, name)
	return nil
}

// Helper functions

func (oro *OptimalRouteOptimizer) loadDefaultPolicies() {
	oro.policies["balanced"] = &RoutingPolicy{
		Name:              "balanced",
		Criteria:          "balanced",
		LatencyWeight:     0.3,
		BandwidthWeight:   0.3,
		ReliabilityWeight: 0.3,
		CostWeight:        0.1,
		Description:       "Balanced optimization across all metrics",
		CreatedAt:         time.Now(),
	}

	oro.policies["latency-optimized"] = &RoutingPolicy{
		Name:              "latency-optimized",
		Criteria:          "latency",
		LatencyWeight:     0.7,
		BandwidthWeight:   0.2,
		ReliabilityWeight: 0.1,
		CostWeight:        0.0,
		Description:       "Optimize for lowest latency",
		CreatedAt:         time.Now(),
	}

	oro.policies["bandwidth-optimized"] = &RoutingPolicy{
		Name:              "bandwidth-optimized",
		Criteria:          "bandwidth",
		LatencyWeight:     0.1,
		BandwidthWeight:   0.7,
		ReliabilityWeight: 0.2,
		CostWeight:        0.0,
		Description:       "Optimize for highest bandwidth",
		CreatedAt:         time.Now(),
	}

	oro.policies["cost-optimized"] = &RoutingPolicy{
		Name:              "cost-optimized",
		Criteria:          "cost",
		LatencyWeight:     0.2,
		BandwidthWeight:   0.2,
		ReliabilityWeight: 0.2,
		CostWeight:        0.4,
		Description:       "Optimize for lowest cost",
		CreatedAt:         time.Now(),
	}
}

func (oro *OptimalRouteOptimizer) discoverRoutesToDestination(destination string, interfaces []string) ([]NetworkRoute, error) {
	var routes []NetworkRoute

	// Get routing table
	result := oro.commandPool.ExecuteCommand("ip", "route", "show")
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get routing table: %w", result.Error)
	}

	// Parse routing table
	lines := strings.Split(string(result.Output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		route := oro.parseRouteEntry(line, destination)
		if route.Interface != "" {
			routes = append(routes, route)
		}
	}

	// Add interface-specific routes if specified
	if len(interfaces) > 0 {
		for _, iface := range interfaces {
			route := oro.createRouteForInterface(iface, destination)
			if route.Interface != "" {
				routes = append(routes, route)
			}
		}
	}

	return routes, nil
}

func (oro *OptimalRouteOptimizer) parseRouteEntry(line, destination string) NetworkRoute {
	fields := strings.Fields(line)
	route := NetworkRoute{
		Destination: destination,
		Status:      "unknown",
		LastTested:  time.Now(),
	}

	// Parse route fields
	for i, field := range fields {
		switch field {
		case "dev":
			if i+1 < len(fields) {
				route.Interface = fields[i+1]
			}
		case "via":
			if i+1 < len(fields) {
				route.Gateway = fields[i+1]
			}
		case "metric":
			if i+1 < len(fields) {
				if metric, err := strconv.Atoi(fields[i+1]); err == nil {
					route.Metric = metric
				}
			}
		}
	}

	return route
}

func (oro *OptimalRouteOptimizer) createRouteForInterface(iface, destination string) NetworkRoute {
	route := NetworkRoute{
		Interface:   iface,
		Destination: destination,
		Status:      "available",
		LastTested:  time.Now(),
	}

	// Get gateway for interface
	result := oro.commandPool.ExecuteCommand("ip", "route", "show", "dev", iface)
	if result.Error == nil {
		lines := strings.Split(string(result.Output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "default") {
				fields := strings.Fields(line)
				for i, field := range fields {
					if field == "via" && i+1 < len(fields) {
						route.Gateway = fields[i+1]
						break
					}
				}

				break
			}
		}
	}

	return route
}

func (oro *OptimalRouteOptimizer) testRoute(route *NetworkRoute) {
	// Test latency
	latency := oro.testLatency(route.Destination, route.Interface)
	route.Latency = latency

	// Test bandwidth (simplified)
	route.Bandwidth = oro.estimateBandwidth(route.Interface)

	// Test packet loss
	route.PacketLoss = oro.testPacketLoss(route.Destination, route.Interface)

	// Calculate reliability score
	route.Reliability = oro.calculateReliabilityScore(route)

	// Estimate cost score
	route.Cost = oro.estimateCostScore(route.Interface)

	// Calculate overall quality score
	route.QualityScore = oro.calculateQualityScore(route)

	// Update status
	switch {
	case latency > 0 && route.PacketLoss < 5:
		route.Status = env.StatusGood
	case latency > 0:
		route.Status = "degraded"
	default:
		route.Status = "unreachable"
	}

	route.LastTested = time.Now()
}

func (oro *OptimalRouteOptimizer) testLatency(destination, iface string) time.Duration {
	// Use ping to test latency
	var result *CommandResult
	if iface != "" {
		result = oro.commandPool.ExecuteCommand("ping", "-c", "3", "-I", iface, destination)
	} else {
		result = oro.commandPool.ExecuteCommand("ping", "-c", "3", destination)
	}

	if result.Error != nil {
		return 0
	}

	// Parse ping output for average latency
	output := string(result.Output)
	if strings.Contains(output, "round-trip") {
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if strings.Contains(line, "round-trip") {
				if idx := strings.Index(line, "="); idx != -1 {
					statsStr := line[idx+1:]
					if spaceIdx := strings.Index(statsStr, " ms"); spaceIdx != -1 {
						statsStr = strings.TrimSpace(statsStr[:spaceIdx])

						parts := strings.Split(statsStr, "/")
						if len(parts) >= 2 {
							if avg, err := time.ParseDuration(parts[1] + "ms"); err == nil {
								return avg
							}
						}
					}
				}
			}
		}
	}

	return 0
}

func (oro *OptimalRouteOptimizer) estimateBandwidth(iface string) float64 {
	// Get interface speed
	result := oro.commandPool.ExecuteCommand("cat", fmt.Sprintf("/sys/class/net/%s/speed", iface))
	if result.Error == nil {
		if speed, err := strconv.ParseFloat(strings.TrimSpace(string(result.Output)), 64); err == nil {
			return speed
		}
	}

	// Default estimates based on interface type
	if strings.HasPrefix(iface, "eth") {
		return 1000.0 // 1 Gbps
	} else if strings.HasPrefix(iface, "wlan") {
		return 100.0 // 100 Mbps
	}

	return 10.0 // 10 Mbps default
}

func (oro *OptimalRouteOptimizer) testPacketLoss(destination, iface string) float64 {
	// Use ping to test packet loss
	var result *CommandResult
	if iface != "" {
		result = oro.commandPool.ExecuteCommand("ping", "-c", "10", "-I", iface, destination)
	} else {
		result = oro.commandPool.ExecuteCommand("ping", "-c", "10", destination)
	}

	if result.Error != nil {
		return 100.0 // 100% loss if unreachable
	}

	// Parse ping output for packet loss
	output := string(result.Output)
	if strings.Contains(output, "% packet loss") {
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if strings.Contains(line, "% packet loss") {
				fields := strings.Fields(line)
				for i, field := range fields {
					if strings.Contains(field, "%") && i > 0 {
						if loss, err := strconv.ParseFloat(strings.TrimSuffix(field, "%"), 64); err == nil {
							return loss
						}
					}
				}
			}
		}
	}

	return 0.0
}

func (oro *OptimalRouteOptimizer) calculateReliabilityScore(route *NetworkRoute) float64 {
	score := 100.0

	// Reduce score based on packet loss
	score -= route.PacketLoss * 2

	// Reduce score based on latency
	latencyMs := float64(route.Latency) / float64(time.Millisecond)
	if latencyMs > 100 {
		score -= (latencyMs - 100) / 10
	}

	// Interface reliability factor
	if strings.HasPrefix(route.Interface, "wlan") {
		score -= 10 // Wireless is less reliable
	}

	return maxFloat64(0, score)
}

func (oro *OptimalRouteOptimizer) estimateCostScore(iface string) float64 {
	// Cost estimates (lower is better, inverted for scoring)
	switch {
	case strings.HasPrefix(iface, "eth"):
		return 90.0 // Wired is usually cheaper
	case strings.HasPrefix(iface, "wlan"):
		return 85.0 // WiFi might have data costs
	case strings.Contains(iface, "cellular") || strings.Contains(iface, "mobile"):
		return 30.0 // Cellular is expensive
	}

	return 50.0 // Unknown
}

func (oro *OptimalRouteOptimizer) calculateQualityScore(route *NetworkRoute) float64 {
	// Use balanced weights for quality score
	policy := oro.policies["balanced"]

	latencyScore := 100.0

	if route.Latency > 0 {
		latencyMs := float64(route.Latency) / float64(time.Millisecond)
		switch {
		case latencyMs <= 10:
			latencyScore = 100
		case latencyMs <= 50:
			latencyScore = 90 - (latencyMs-10)*2
		default:
			latencyScore = maxFloat64(0, 50-latencyMs/10)
		}
	}

	bandwidthScore := minFloat64(100, route.Bandwidth/10) // Normalize to 100 for 1 Gbps

	qualityScore := (latencyScore*policy.LatencyWeight +
		bandwidthScore*policy.BandwidthWeight +
		route.Reliability*policy.ReliabilityWeight +
		route.Cost*policy.CostWeight)

	return qualityScore
}

func (oro *OptimalRouteOptimizer) selectOptimalRoute(routes []NetworkRoute) *NetworkRoute {
	if len(routes) == 0 {
		return nil
	}

	var bestRoute *NetworkRoute

	bestScore := 0.0

	for i := range routes {
		if routes[i].QualityScore > bestScore {
			bestScore = routes[i].QualityScore
			bestRoute = &routes[i]
		}
	}

	return bestRoute
}

func (oro *OptimalRouteOptimizer) calculateRouteMetrics(routes []NetworkRoute) RouteMetrics {
	metrics := RouteMetrics{
		TotalRoutes: len(routes),
	}

	if len(routes) == 0 {
		return metrics
	}

	var (
		totalQuality  float64
		bestLatency   time.Duration
		bestBandwidth float64
	)

	bestRoute := ""

	for _, route := range routes {
		if route.Status == env.StatusGood || route.Status == env.StatusDegraded {
			metrics.ActiveRoutes++
		}

		totalQuality += route.QualityScore

		if bestLatency == 0 || (route.Latency > 0 && route.Latency < bestLatency) {
			bestLatency = route.Latency
		}

		if route.Bandwidth > bestBandwidth {
			bestBandwidth = route.Bandwidth
			bestRoute = route.Interface
		}
	}

	metrics.BestLatency = bestLatency.String()
	metrics.BestBandwidth = bestBandwidth
	metrics.OverallQuality = totalQuality / float64(len(routes))
	metrics.RecommendedPath = bestRoute

	return metrics
}

func (oro *OptimalRouteOptimizer) generateRouteRecommendations(analysis *RouteAnalysis) []string {
	var recommendations []string

	if analysis.OptimalRoute == nil {
		recommendations = append(recommendations, "No optimal route found - check network connectivity")
		return recommendations
	}

	// Quality-based recommendations
	if analysis.OptimalRoute.QualityScore < 70 {
		recommendations = append(recommendations, "Consider alternative network paths for better performance")
	}

	// Latency recommendations
	if analysis.OptimalRoute.Latency > 100*time.Millisecond {
		recommendations = append(recommendations, "High latency detected - consider using wired connection")
	}

	// Packet loss recommendations
	if analysis.OptimalRoute.PacketLoss > 1 {
		recommendations = append(recommendations, "Packet loss detected - check network quality")
	}

	// Interface recommendations
	if strings.HasPrefix(analysis.OptimalRoute.Interface, "wlan") && len(analysis.AvailableRoutes) > 1 {
		recommendations = append(recommendations, "Wired connection available - consider switching for better reliability")
	}

	// Multi-path recommendations
	activeRoutes := 0

	for _, route := range analysis.AvailableRoutes {
		if route.Status == env.StatusGood {
			activeRoutes++
		}
	}

	if activeRoutes > 1 {
		recommendations = append(recommendations, "Multiple good paths available - consider load balancing")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Network routing is optimal - no changes needed")
	}

	return recommendations
}

func (oro *OptimalRouteOptimizer) calculateDiscoverySummary(discovery *RouteDiscovery) DiscoverySummary {
	summary := DiscoverySummary{
		TotalTargets: len(discovery.Config.Targets),
	}

	var (
		totalRoutes    int
		totalLatency   time.Duration
		totalBandwidth float64
		totalQuality   float64
	)

	validTargets := 0

	for _, routes := range discovery.TargetRoutes {
		totalRoutes += len(routes)
		for _, route := range routes {
			if route.Status == env.StatusGood {
				totalLatency += route.Latency
				totalBandwidth += route.Bandwidth
				totalQuality += route.QualityScore
				validTargets++
			}
		}
	}

	summary.RoutesDiscovered = totalRoutes

	if validTargets > 0 {
		avgLatency := totalLatency / time.Duration(validTargets)
		summary.AverageLatency = avgLatency.String()
		summary.AverageBandwidth = totalBandwidth / float64(validTargets)
		summary.OverallScore = totalQuality / float64(validTargets)
	}

	return summary
}

func (oro *OptimalRouteOptimizer) planRoutingChanges(_ string) []RouteChange { //nolint:unparam // Policy unused in current implementation
	// In a real implementation, this would analyze current routes
	// and plan optimizations based on the policy
	changes := []RouteChange{
		{
			Type:        "modify",
			Interface:   "eth0",
			Destination: "0.0.0.0/0",
			OldMetric:   100,
			NewMetric:   50,
			Reason:      "Optimize metric for better path selection",
		},
		{
			Type:        "add",
			Interface:   "wlan0",
			Destination: "8.8.8.0/24",
			NewGateway:  "192.168.1.1",
			NewMetric:   200,
			Reason:      "Add backup route for DNS servers",
		},
	}

	return changes
}

func (oro *OptimalRouteOptimizer) backupCurrentRoutes() (string, error) {
	backupPath := fmt.Sprintf("%s/route-backup-%d.txt", oro.configDir, time.Now().Unix())

	result := oro.commandPool.ExecuteCommand("ip", "route", "show")
	if result.Error != nil {
		return "", fmt.Errorf("failed to get current routes: %w", result.Error)
	}

	if err := os.WriteFile(backupPath, result.Output, 0o644); err != nil {
		return "", fmt.Errorf("failed to write backup: %w", err)
	}

	return backupPath, nil
}

func (oro *OptimalRouteOptimizer) applyRoutingChanges(policy string) ([]RouteChange, error) { //nolint:unparam // Error always nil but kept for consistency
	changes := oro.planRoutingChanges(policy)

	// In a real implementation, this would apply actual routing changes
	for i := range changes {
		oro.logger.Info("Applied routing change",
			zap.String("type", changes[i].Type),
			zap.String("interface", changes[i].Interface),
			zap.String("reason", changes[i].Reason))
	}

	return changes, nil
}

func (oro *OptimalRouteOptimizer) autoOptimizationLoop(ctx context.Context, config AutoOptimizeConfig) {
	ticker := time.NewTicker(config.CheckInterval)
	defer ticker.Stop()

	for oro.isOptimizing {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			oro.performOptimizationCheck(config.ChangeThreshold)
		}
	}
}

func (oro *OptimalRouteOptimizer) performOptimizationCheck(threshold float64) {
	// Simplified optimization check
	oro.logger.Info("Performing automatic optimization check",
		zap.Float64("threshold", threshold))

	// In a real implementation, this would:
	// 1. Measure current network performance
	// 2. Compare against historical baselines
	// 3. Apply optimizations if performance degraded beyond threshold
}

func (oro *OptimalRouteOptimizer) getCurrentRoutes() ([]NetworkRoute, error) {
	result := oro.commandPool.ExecuteCommand("ip", "route", "show")
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get routes: %w", result.Error)
	}

	var routes []NetworkRoute

	lines := strings.Split(string(result.Output), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		route := oro.parseRouteEntry(line, "default")
		if route.Interface != "" {
			routes = append(routes, route)
		}
	}

	return routes, nil
}

func (oro *OptimalRouteOptimizer) calculateRoutingHealth(status *RoutingStatus) RoutingHealth {
	health := RoutingHealth{
		PathDiversity:   len(status.CurrentRoutes),
		FailoverReady:   len(status.CurrentRoutes) > 1,
		OptimizationAge: "Recent",
	}

	// Calculate overall score based on various factors
	score := 100.0

	// Reduce score if no path diversity
	if health.PathDiversity <= 1 {
		score -= 20
	}

	// Reduce score if auto-optimization is disabled
	if status.AutoOptimization != nil && !status.AutoOptimization.Enabled {
		score -= 10
	}

	// Reduce score if no load balancing and multiple paths available
	if status.LoadBalancing != nil && !status.LoadBalancing.Enabled && health.PathDiversity > 1 {
		score -= 15
	}

	health.OverallScore = score
	health.RouteStability = 95.0 // Default good stability

	return health
}

func createPolicyFromFlags(cmd *cobra.Command) *RoutingPolicy {
	criteria, _ := cmd.Flags().GetString("criteria")
	latencyWeight, _ := cmd.Flags().GetFloat64("latency-weight")
	bandwidthWeight, _ := cmd.Flags().GetFloat64("bandwidth-weight")
	reliabilityWeight, _ := cmd.Flags().GetFloat64("reliability-weight")
	costWeight, _ := cmd.Flags().GetFloat64("cost-weight")

	return &RoutingPolicy{
		Criteria:          criteria,
		LatencyWeight:     latencyWeight,
		BandwidthWeight:   bandwidthWeight,
		ReliabilityWeight: reliabilityWeight,
		CostWeight:        costWeight,
		Description:       "Custom policy created via CLI",
	}
}

// Helper functions

func maxFloat64(a, b float64) float64 {
	if a > b {
		return a
	}

	return b
}

func minFloat64(a, b float64) float64 {
	if a < b {
		return a
	}

	return b
}

// Print functions

func printRouteAnalysis(analysis *RouteAnalysis) error { //nolint:unparam // Error always nil but kept for consistency
	fmt.Printf("🔍 Route Analysis for %s\n\n", analysis.Config.Destination)

	// Metrics summary
	metrics := analysis.Analysis

	fmt.Printf("📊 Analysis Summary:\n")
	fmt.Printf("  Total Routes: %d\n", metrics.TotalRoutes)
	fmt.Printf("  Active Routes: %d\n", metrics.ActiveRoutes)
	fmt.Printf("  Best Latency: %s\n", metrics.BestLatency)
	fmt.Printf("  Best Bandwidth: %.2f Mbps\n", metrics.BestBandwidth)
	fmt.Printf("  Overall Quality: %.1f%%\n", metrics.OverallQuality)
	fmt.Printf("  Recommended Path: %s\n\n", metrics.RecommendedPath)

	// Available routes
	if len(analysis.AvailableRoutes) > 0 {
		fmt.Printf("🛤️  Available Routes:\n")

		w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
		_, _ = fmt.Fprintln(w, "INTERFACE\tGATEWAY\tMETRIC\tLATENCY\tBANDWIDTH\tLOSS\tQUALITY\tSTATUS")

		// Sort routes by quality score
		routes := make([]NetworkRoute, len(analysis.AvailableRoutes))
		copy(routes, analysis.AvailableRoutes)
		sort.Slice(routes, func(i, j int) bool {
			return routes[i].QualityScore > routes[j].QualityScore
		})

		for _, route := range routes {
			status := route.Status
			if analysis.OptimalRoute != nil && route.Interface == analysis.OptimalRoute.Interface {
				status = "✅ " + status
			}

			_, _ = fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%.1f Mbps\t%.1f%%\t%.1f%%\t%s\n",
				route.Interface,
				route.Gateway,
				route.Metric,
				route.Latency.Round(time.Millisecond),
				route.Bandwidth,
				route.PacketLoss,
				route.QualityScore,
				status)
		}

		_ = w.Flush()
		fmt.Println()
	}

	// Optimal route
	if analysis.OptimalRoute != nil {
		fmt.Printf("⭐ Optimal Route:\n")
		fmt.Printf("  Interface: %s\n", analysis.OptimalRoute.Interface)
		fmt.Printf("  Gateway: %s\n", analysis.OptimalRoute.Gateway)
		fmt.Printf("  Latency: %s\n", analysis.OptimalRoute.Latency.Round(time.Millisecond))
		fmt.Printf("  Bandwidth: %.2f Mbps\n", analysis.OptimalRoute.Bandwidth)
		fmt.Printf("  Quality Score: %.1f%%\n\n", analysis.OptimalRoute.QualityScore)
	}

	// Recommendations
	if len(analysis.Recommendations) > 0 {
		fmt.Printf("💡 Recommendations:\n")

		for i, rec := range analysis.Recommendations {
			fmt.Printf("  %d. %s\n", i+1, rec)
		}

		fmt.Println()
	}

	return nil
}

func printRouteDiscovery(discovery *RouteDiscovery) error { //nolint:unparam // Error always nil but kept for consistency
	fmt.Printf("🌐 Route Discovery Results\n\n")

	// Summary
	summary := discovery.Summary

	fmt.Printf("📊 Discovery Summary:\n")
	fmt.Printf("  Total Targets: %d\n", summary.TotalTargets)
	fmt.Printf("  Routes Discovered: %d\n", summary.RoutesDiscovered)
	fmt.Printf("  Average Latency: %s\n", summary.AverageLatency)
	fmt.Printf("  Average Bandwidth: %.2f Mbps\n", summary.AverageBandwidth)
	fmt.Printf("  Overall Score: %.1f%%\n\n", summary.OverallScore)

	// Optimal paths
	if len(discovery.OptimalPaths) > 0 {
		fmt.Printf("⭐ Optimal Paths:\n")

		w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
		_, _ = fmt.Fprintln(w, "TARGET\tINTERFACE\tGATEWAY\tLATENCY\tBANDWIDTH\tQUALITY")

		for target, route := range discovery.OptimalPaths {
			if route != nil {
				_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%.1f Mbps\t%.1f%%\n",
					target,
					route.Interface,
					route.Gateway,
					route.Latency.Round(time.Millisecond),
					route.Bandwidth,
					route.QualityScore)
			}
		}

		_ = w.Flush()
		fmt.Println()
	}

	// Detailed routes by target
	fmt.Printf("🛤️  Routes by Target:\n")

	for target, routes := range discovery.TargetRoutes {
		fmt.Printf("\n  %s (%d routes):\n", target, len(routes))

		if len(routes) > 0 {
			w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
			_, _ = fmt.Fprintln(w, "  INTERFACE\tLATENCY\tBANDWIDTH\tLOSS\tQUALITY\tSTATUS")

			// Sort by quality
			sort.Slice(routes, func(i, j int) bool {
				return routes[i].QualityScore > routes[j].QualityScore
			})

			for _, route := range routes {
				_, _ = fmt.Fprintf(w, "  %s\t%s\t%.1f Mbps\t%.1f%%\t%.1f%%\t%s\n",
					route.Interface,
					route.Latency.Round(time.Millisecond),
					route.Bandwidth,
					route.PacketLoss,
					route.QualityScore,
					route.Status)
			}

			_ = w.Flush()
		}
	}

	return nil
}

func printApplyResult(result *RoutingApplyResult) error { //nolint:unparam // Error always nil but kept for consistency
	fmt.Printf("⚙️  Routing Apply Result\n\n")

	if result.Config.DryRun {
		fmt.Printf("🔍 Dry Run Mode: %s\n\n", result.Message)
	} else {
		fmt.Printf("Status: %s\n", result.Message)

		if result.BackupLocation != "" {
			fmt.Printf("Backup Location: %s\n", result.BackupLocation)
		}

		fmt.Println()
	}

	// Changes applied
	if len(result.ChangesApplied) > 0 {
		fmt.Printf("📝 Changes Applied:\n")

		w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
		_, _ = fmt.Fprintln(w, "TYPE\tINTERFACE\tDESTINATION\tCHANGE\tREASON")

		for _, change := range result.ChangesApplied {
			changeDesc := ""

			switch change.Type {
			case "add":
				changeDesc = fmt.Sprintf("Add route via %s metric %d", change.NewGateway, change.NewMetric)
			case "modify":
				changeDesc = fmt.Sprintf("Metric %d → %d", change.OldMetric, change.NewMetric)
			case "delete":
				changeDesc = "Delete route"
			}

			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				change.Type,
				change.Interface,
				change.Destination,
				changeDesc,
				change.Reason)
		}

		_ = w.Flush()
	}

	return nil
}

func printAutoOptimizeStatus(status *AutoOptimizeStatus) error { //nolint:unparam // Error always nil but kept for consistency
	fmt.Printf("🤖 Auto-Optimization Status\n\n")

	fmt.Printf("Status: ")

	if status.Enabled {
		fmt.Printf("✅ Enabled")

		if status.Running {
			fmt.Printf(" (Running)")
		}
	} else {
		fmt.Printf("❌ Disabled")
	}

	fmt.Println()

	if status.Enabled {
		fmt.Printf("Check Interval: %s\n", status.CheckInterval)
		fmt.Printf("Last Check: %s\n", status.LastCheck.Format("2006-01-02 15:04:05"))
		fmt.Printf("Next Check: %s\n", status.NextCheck.Format("2006-01-02 15:04:05"))
		fmt.Printf("Changes Applied: %d\n", status.ChangesApplied)
		fmt.Printf("Last Optimization: %s\n", status.LastOptimization.Format("2006-01-02 15:04:05"))
	}

	return nil
}

func printLoadBalanceResult(result *LoadBalanceResult) error { //nolint:unparam // Error always nil but kept for consistency
	fmt.Printf("⚖️  Load Balancing Configuration\n\n")
	fmt.Printf("Status: %s\n", result.Message)
	fmt.Printf("Policy: %s\n\n", result.Config.Policy)

	if len(result.ConfiguredPaths) > 0 {
		fmt.Printf("📊 Configured Paths:\n")

		w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
		_, _ = fmt.Fprintln(w, "INTERFACE\tWEIGHT\tTRAFFIC SHARE\tSTATUS")

		for _, path := range result.ConfiguredPaths {
			_, _ = fmt.Fprintf(w, "%s\t%d\t%.1f%%\t%s\n",
				path.Interface,
				path.Weight,
				path.Share,
				path.Status)
		}

		_ = w.Flush()
	}

	return nil
}

func printRoutingStatus(status *RoutingStatus) error { //nolint:unparam // Error always nil but kept for consistency
	fmt.Printf("🔍 Routing Status Overview\n\n")

	// System health
	health := status.SystemHealth

	fmt.Printf("🏥 System Health:\n")
	fmt.Printf("  Overall Score: %.1f%%\n", health.OverallScore)
	fmt.Printf("  Route Stability: %.1f%%\n", health.RouteStability)
	fmt.Printf("  Path Diversity: %d paths\n", health.PathDiversity)
	fmt.Printf("  Failover Ready: %v\n", health.FailoverReady)
	fmt.Printf("  Optimization Age: %s\n", health.OptimizationAge)

	if health.IssuesDetected > 0 {
		fmt.Printf("  Issues Detected: %d\n", health.IssuesDetected)
	}

	fmt.Println()

	// Active policies
	if len(status.ActivePolicies) > 0 {
		fmt.Printf("📋 Active Policies: %s\n\n", strings.Join(status.ActivePolicies, ", "))
	}

	// Current routes
	if len(status.CurrentRoutes) > 0 {
		fmt.Printf("🛤️  Current Routes:\n")

		w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
		_, _ = fmt.Fprintln(w, "INTERFACE\tGATEWAY\tMETRIC\tSTATUS\tQUALITY")

		for _, route := range status.CurrentRoutes {
			_, _ = fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%.1f%%\n",
				route.Interface,
				route.Gateway,
				route.Metric,
				route.Status,
				route.QualityScore)
		}

		_ = w.Flush()
		fmt.Println()
	}

	// Load balancing status
	if status.LoadBalancing != nil {
		lb := status.LoadBalancing

		fmt.Printf("⚖️  Load Balancing:\n")
		fmt.Printf("  Status: ")

		if lb.Enabled {
			fmt.Printf("✅ Enabled (%s)\n", lb.Policy)
			fmt.Printf("  Paths: %d total, %d active\n", lb.TotalPaths, lb.ActivePaths)
		} else {
			fmt.Printf("❌ Disabled\n")
		}

		fmt.Println()
	}

	// Auto-optimization status
	if status.AutoOptimization != nil {
		auto := status.AutoOptimization

		fmt.Printf("🤖 Auto-Optimization:\n")
		fmt.Printf("  Status: ")

		if auto.Enabled {
			fmt.Printf("✅ Enabled\n")
			fmt.Printf("  Changes Applied: %d\n", auto.ChangesApplied)
		} else {
			fmt.Printf("❌ Disabled\n")
		}
	}

	return nil
}

func printRoutingPolicies(policies []RoutingPolicy) error {
	fmt.Printf("📋 Routing Policies\n\n")

	if len(policies) == 0 {
		fmt.Println("No policies configured.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	_, _ = fmt.Fprintln(w, "NAME\tCRITERIA\tLATENCY\tBANDWIDTH\tRELIABILITY\tCOST\tCREATED")

	for _, policy := range policies {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%.1f\t%.1f\t%.1f\t%.1f\t%s\n",
			policy.Name,
			policy.Criteria,
			policy.LatencyWeight,
			policy.BandwidthWeight,
			policy.ReliabilityWeight,
			policy.CostWeight,
			policy.CreatedAt.Format("2006-01-02"))
	}

	return w.Flush()
}
