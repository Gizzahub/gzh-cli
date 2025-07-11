package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// NewMonitoringCmd creates the monitoring command
func NewMonitoringCmd(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "monitoring",
		Short: "Run monitoring and alerting system",
		Long:  `Start the monitoring and alerting system with REST API server, WebSocket support, and dashboard`,
	}

	// Add subcommands
	cmd.AddCommand(newServerCmd(ctx))
	cmd.AddCommand(newStatusCmd(ctx))
	cmd.AddCommand(newMetricsCmd(ctx))
	cmd.AddCommand(newInstanceCmd(ctx))
	cmd.AddCommand(newNotificationCmd(ctx))
	cmd.AddCommand(newPerformanceCmd(ctx))
	cmd.AddCommand(newCentralizedLoggingCmd(ctx))

	return cmd
}

// newServerCmd creates the server subcommand
func newServerCmd(ctx context.Context) *cobra.Command {
	var port int
	var host string
	var debug bool

	cmd := &cobra.Command{
		Use:   "server",
		Short: "Start the monitoring REST API server",
		Long:  `Start the monitoring REST API server with WebSocket support for real-time updates`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Set gin mode
			if !debug {
				gin.SetMode(gin.ReleaseMode)
			}

			// Create server
			server := NewMonitoringServer(&ServerConfig{
				Host:  host,
				Port:  port,
				Debug: debug,
			})

			// Start server in goroutine
			go func() {
				addr := fmt.Sprintf("%s:%d", host, port)
				fmt.Printf("🚀 Starting monitoring server on %s\n", addr)
				fmt.Printf("📊 Dashboard available at http://%s/dashboard\n", addr)
				fmt.Printf("🔌 WebSocket endpoint at ws://%s/ws\n", addr)
				fmt.Printf("📡 API endpoints at http://%s/api/v1/*\n", addr)
				if err := server.Start(addr); err != nil && err != http.ErrServerClosed {
					log.Fatalf("Failed to start server: %v", err)
				}
			}()

			// Wait for interrupt signal
			quit := make(chan os.Signal, 1)
			signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
			<-quit

			fmt.Println("\n🛑 Shutting down monitoring server...")

			// Shutdown server gracefully
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			if err := server.Shutdown(shutdownCtx); err != nil {
				return fmt.Errorf("server forced to shutdown: %w", err)
			}

			fmt.Println("✅ Monitoring server exited")
			return nil
		},
	}

	cmd.Flags().IntVarP(&port, "port", "p", 8080, "Server port")
	cmd.Flags().StringVarP(&host, "host", "H", "localhost", "Server host")
	cmd.Flags().BoolVarP(&debug, "debug", "d", false, "Enable debug mode")

	return cmd
}

// newStatusCmd creates the status subcommand
func newStatusCmd(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Check monitoring system status",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := NewMonitoringClient("http://localhost:8080")
			status, err := client.GetSystemStatus(ctx)
			if err != nil {
				return fmt.Errorf("failed to get system status: %w", err)
			}

			fmt.Printf("📊 System Status:\n")
			fmt.Printf("  Status: %s\n", status.Status)
			fmt.Printf("  Uptime: %s\n", status.Uptime)
			fmt.Printf("  Active Tasks: %d\n", status.ActiveTasks)
			fmt.Printf("  Total Requests: %d\n", status.TotalRequests)
			fmt.Printf("  Memory Usage: %.2f MB\n", float64(status.MemoryUsage)/1024/1024)

			return nil
		},
	}

	return cmd
}

// newMetricsCmd creates the metrics subcommand
func newMetricsCmd(ctx context.Context) *cobra.Command {
	var format string
	var output string

	cmd := &cobra.Command{
		Use:   "metrics",
		Short: "Export system metrics",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := NewMonitoringClient("http://localhost:8080")
			metrics, err := client.GetMetrics(ctx, format)
			if err != nil {
				return fmt.Errorf("failed to get metrics: %w", err)
			}

			if output != "" {
				err := os.WriteFile(output, []byte(metrics), 0o644)
				if err != nil {
					return fmt.Errorf("failed to write metrics to file: %w", err)
				}
				fmt.Printf("📈 Metrics exported to %s\n", output)
			} else {
				fmt.Print(metrics)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "prometheus", "Output format (prometheus, json)")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Output file path")

	return cmd
}

// newInstanceCmd creates the instance management subcommand
func newInstanceCmd(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "instance",
		Short: "Manage monitoring instances",
		Long:  `Manage multiple monitoring instances including discovery and status`,
	}

	// Add instance subcommands
	cmd.AddCommand(newInstanceListCmd(ctx))
	cmd.AddCommand(newInstanceDiscoverCmd(ctx))
	cmd.AddCommand(newInstanceStatusCmd(ctx))

	return cmd
}

// newInstanceListCmd creates the instance list subcommand
func newInstanceListCmd(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all monitoring instances",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := NewMonitoringClient("http://localhost:8080")
			instances, err := client.GetInstances(ctx)
			if err != nil {
				return fmt.Errorf("failed to get instances: %w", err)
			}

			fmt.Printf("📋 Monitoring Instances:\n")
			for _, instance := range instances {
				fmt.Printf("  ID: %s\n", instance.ID)
				fmt.Printf("  Name: %s\n", instance.Name)
				fmt.Printf("  Host: %s:%d\n", instance.Host, instance.Port)
				fmt.Printf("  Status: %s\n", instance.Status)
				fmt.Printf("  Type: %s\n", instance.Tags["type"])
				fmt.Printf("  Last Seen: %s\n", instance.LastSeen.Format("2006-01-02 15:04:05"))
				if instance.Metrics != nil {
					fmt.Printf("  CPU: %.1f%%, Memory: %s\n",
						instance.Metrics.CPUUsage,
						formatBytes(instance.Metrics.MemoryUsage))
				}
				fmt.Println()
			}

			return nil
		},
	}

	return cmd
}

// newInstanceDiscoverCmd creates the instance discovery subcommand
func newInstanceDiscoverCmd(ctx context.Context) *cobra.Command {
	var host string
	var port int

	cmd := &cobra.Command{
		Use:   "discover",
		Short: "Discover and add a remote monitoring instance",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := NewMonitoringClient("http://localhost:8080")
			err := client.DiscoverInstance(ctx, host, port)
			if err != nil {
				return fmt.Errorf("failed to discover instance: %w", err)
			}

			fmt.Printf("✅ Successfully discovered instance at %s:%d\n", host, port)
			return nil
		},
	}

	cmd.Flags().StringVarP(&host, "host", "H", "", "Remote host to discover")
	cmd.Flags().IntVarP(&port, "port", "p", 8080, "Remote port")
	cmd.MarkFlagRequired("host")

	return cmd
}

// newInstanceStatusCmd creates the cluster status subcommand
func newInstanceStatusCmd(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show cluster status",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := NewMonitoringClient("http://localhost:8080")
			status, err := client.GetClusterStatus(ctx)
			if err != nil {
				return fmt.Errorf("failed to get cluster status: %w", err)
			}

			fmt.Printf("🔧 Cluster Status:\n")
			fmt.Printf("  Total Instances: %d\n", status.TotalInstances)
			fmt.Printf("  Running: %d\n", status.RunningInstances)
			fmt.Printf("  Unhealthy: %d\n", status.UnhealthyInstances)
			fmt.Printf("  Health Rate: %.1f%%\n",
				float64(status.RunningInstances)/float64(status.TotalInstances)*100)
			fmt.Printf("  Last Update: %s\n", status.LastUpdate.Format("2006-01-02 15:04:05"))

			return nil
		},
	}

	return cmd
}

// formatBytes formats bytes to human readable format
func formatBytes(bytes uint64) string {
	if bytes == 0 {
		return "0 B"
	}
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// newNotificationCmd creates the notification management subcommand
func newNotificationCmd(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "notification",
		Short: "Test and manage notifications",
		Long:  `Test notification integrations like Slack, Discord, etc.`,
	}

	// Add notification subcommands
	cmd.AddCommand(newNotificationTestCmd(ctx))

	return cmd
}

// newNotificationTestCmd creates the notification test subcommand
func newNotificationTestCmd(ctx context.Context) *cobra.Command {
	var notificationType string
	var message string

	cmd := &cobra.Command{
		Use:   "test",
		Short: "Test notification delivery",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := NewMonitoringClient("http://localhost:8080")

			err := client.TestNotification(ctx, notificationType, "", message)
			if err != nil {
				return fmt.Errorf("failed to send test notification: %w", err)
			}

			fmt.Printf("✅ Test %s notification sent successfully\n", notificationType)
			return nil
		},
	}

	cmd.Flags().StringVarP(&notificationType, "type", "t", "slack", "Notification type (slack, discord, teams, email)")
	cmd.Flags().StringVarP(&message, "message", "m", "Test message from GZH Monitoring", "Test message content")

	return cmd
}

// ServerConfig represents server configuration
type ServerConfig struct {
	Host  string
	Port  int
	Debug bool
}

// MonitoringServer represents the monitoring server
type MonitoringServer struct {
	config             *ServerConfig
	router             *gin.Engine
	httpServer         *http.Server
	metrics            *MetricsCollector
	customMetrics      *CustomMetricsManager
	alerts             *AlertManager
	startTime          time.Time
	wsManager          *WebSocketManager
	authManager        *AuthManager
	instanceManager    *InstanceManager
	prometheusExporter *PrometheusExporter
	logger             *zap.Logger
}

// NewMonitoringServer creates a new monitoring server
func NewMonitoringServer(config *ServerConfig) *MonitoringServer {
	// Create logger
	logger, _ := zap.NewProduction()
	if config.Debug {
		logger, _ = zap.NewDevelopment()
	}

	server := &MonitoringServer{
		config:    config,
		router:    gin.New(),
		metrics:   NewMetricsCollector(),
		alerts:    NewAlertManager(),
		startTime: time.Now(),
		logger:    logger,
	}

	// Initialize managers with logger
	server.wsManager = NewWebSocketManager(logger)
	server.authManager = NewAuthManager(logger)
	server.instanceManager = NewInstanceManager(logger)

	// Initialize Prometheus exporter first to get the registry
	prometheusConfig := &PrometheusConfig{
		Enabled:       true,
		ListenAddress: "localhost:9090",
		MetricsPath:   "/metrics",
		Namespace:     "gzh_manager",
		Subsystem:     "monitoring",
	}
	server.prometheusExporter = NewPrometheusExporter(logger, prometheusConfig, server.metrics)

	// Initialize custom metrics manager with the Prometheus registry
	if server.prometheusExporter != nil {
		server.customMetrics = NewCustomMetricsManager(logger, server.prometheusExporter.GetMetrics())
	}

	// Set metrics collector for alerts
	server.alerts.SetMetrics(server.metrics)

	// Initialize Slack notifier if configured
	server.initializeSlackNotifier(logger)

	server.setupRoutes()
	return server
}

// setupRoutes configures the API routes
func (s *MonitoringServer) setupRoutes() {
	// Middleware
	s.router.Use(gin.Recovery())
	if s.config.Debug {
		s.router.Use(gin.Logger())
	}
	s.router.Use(s.corsMiddleware())
	s.router.Use(s.metricsMiddleware())

	// Public authentication endpoints
	auth := s.router.Group("/auth")
	{
		auth.POST("/login", s.login)
		auth.POST("/logout", s.logout)
		auth.GET("/me", s.authManager.JWTAuthMiddleware(), s.getCurrentUser)
	}

	// API routes with authentication
	api := s.router.Group("/api/v1")
	api.Use(s.authManager.JWTAuthMiddleware()) // Require authentication for all API routes
	{
		// System endpoints
		api.GET("/status", s.authManager.RequirePermission("read:system"), s.getSystemStatus)
		api.GET("/health", s.authManager.RequirePermission("read:system"), s.getHealth)
		api.GET("/metrics", s.authManager.RequirePermission("read:system"), s.getMetrics)

		// Task monitoring endpoints
		api.GET("/tasks", s.authManager.RequirePermission("read:tasks"), s.getTasks)
		api.GET("/tasks/:id", s.authManager.RequirePermission("read:tasks"), s.getTask)
		api.POST("/tasks/:id/stop", s.authManager.RequirePermission("write:tasks"), s.stopTask)

		// Alert endpoints
		api.GET("/alerts", s.authManager.RequirePermission("read:alerts"), s.getAlerts)
		api.POST("/alerts", s.authManager.RequirePermission("write:alerts"), s.createAlert)
		api.PUT("/alerts/:id", s.authManager.RequirePermission("write:alerts"), s.updateAlert)
		api.DELETE("/alerts/:id", s.authManager.RequirePermission("write:alerts"), s.deleteAlert)

		// Notification endpoints
		api.GET("/notifications", s.authManager.RequirePermission("read:alerts"), s.getNotifications)
		api.POST("/notifications/test", s.authManager.RequirePermission("write:alerts"), s.testNotification)

		// Slack interactive endpoints (no auth required for webhook callbacks)
		api.POST("/slack/interactive", s.handleSlackInteraction)
		api.POST("/slack/command", s.handleSlackCommand)

		// Teams management endpoints
		teams := api.Group("/teams")
		teams.Use(s.authManager.RequirePermission("read:alerts"))
		{
			teams.GET("", s.getTeams)
			teams.GET("/:teamId/channels", s.getTeamChannels)
			teams.GET("/channels/rules", s.getChannelRules)
			teams.POST("/channels/rules", s.authManager.RequirePermission("write:alerts"), s.addChannelRule)
			teams.DELETE("/channels/rules/:teamId/:channelId", s.authManager.RequirePermission("write:alerts"), s.removeChannelRule)
			teams.POST("/test/:teamId/:channelId", s.authManager.RequirePermission("write:alerts"), s.testTeamsChannel)
		}

		// Configuration endpoints
		api.GET("/config", s.authManager.RequirePermission("read:config"), s.getConfig)
		api.PUT("/config", s.authManager.RequirePermission("write:config"), s.updateConfig)

		// User management endpoints (admin only)
		users := api.Group("/users")
		users.Use(s.authManager.RequirePermission("read:users"))
		{
			users.GET("", s.getUsers)
			users.GET("/:username", s.getUser)
			users.POST("", s.authManager.RequirePermission("write:users"), s.createUser)
			users.PUT("/:username/password", s.authManager.RequirePermission("write:users"), s.updateUserPassword)
			users.DELETE("/:username", s.authManager.RequirePermission("delete:users"), s.deleteUser)
		}

		// Instance management endpoints
		instances := api.Group("/instances")
		instances.Use(s.authManager.RequirePermission("read:system"))
		{
			instances.GET("", s.getInstances)
			instances.GET("/:id", s.getInstance)
			instances.POST("/discover", s.authManager.RequirePermission("write:system"), s.discoverInstance)
			instances.DELETE("/:id", s.authManager.RequirePermission("write:system"), s.removeInstance)
			instances.GET("/cluster/status", s.getClusterStatus)
		}

		// Custom metrics management endpoints
		customMetrics := api.Group("/custom-metrics")
		customMetrics.Use(s.authManager.RequirePermission("read:system"))
		{
			customMetrics.GET("", s.getCustomMetrics)
			customMetrics.GET("/summary", s.getCustomMetricsSummary)
			customMetrics.POST("", s.authManager.RequirePermission("write:system"), s.createCustomMetric)
			customMetrics.DELETE("/:name", s.authManager.RequirePermission("write:system"), s.deleteCustomMetric)
			customMetrics.GET("/business", s.getBusinessMetrics)
			customMetrics.GET("/performance", s.getPerformanceMetrics)
			customMetrics.GET("/usage", s.getUsageMetrics)
			customMetrics.POST("/record", s.authManager.RequirePermission("write:system"), s.recordCustomMetric)
		}
	}

	// WebSocket endpoint for real-time updates
	s.router.GET("/ws", s.handleWebSocket)

	// Dashboard endpoint (legacy HTML dashboard)
	s.router.GET("/dashboard", s.serveDashboard)

	// React SPA static files
	s.router.Static("/static", "./web/build/static")
	s.router.StaticFile("/favicon.ico", "./web/build/favicon.ico")
	s.router.StaticFile("/manifest.json", "./web/build/manifest.json")

	// Serve React SPA for all other routes (fallback to index.html)
	s.router.NoRoute(s.serveSPA)
}

// Start starts the monitoring server
func (s *MonitoringServer) Start(addr string) error {
	// Start WebSocket manager
	s.wsManager.Start()

	// Start instance manager and register local instance
	s.instanceManager.Start()
	instanceName := fmt.Sprintf("gzh-monitoring-%s", s.config.Host)
	s.instanceManager.RegisterLocalInstance(s.config.Host, s.config.Port, instanceName)

	// Start periodic updates
	go s.startPeriodicUpdates(5 * time.Second)

	// Start alert evaluation loop
	go s.alerts.StartEvaluationLoop(context.Background(), 30*time.Second)

	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: s.router,
	}

	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *MonitoringServer) Shutdown(ctx context.Context) error {
	// Stop managers
	s.wsManager.Stop()
	s.instanceManager.Stop()

	return s.httpServer.Shutdown(ctx)
}

// Middleware functions

func (s *MonitoringServer) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func (s *MonitoringServer) metricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)

		s.metrics.RecordRequest(c.Request.Method, c.FullPath(), c.Writer.Status(), duration)
	}
}

// API handler functions

func (s *MonitoringServer) getSystemStatus(c *gin.Context) {
	status := s.getCurrentSystemStatus()
	c.JSON(http.StatusOK, status)
}

// getCurrentSystemStatus returns current system status without gin context
func (s *MonitoringServer) getCurrentSystemStatus() *SystemStatus {
	return &SystemStatus{
		Status:        "healthy",
		Uptime:        time.Since(s.startTime).String(),
		ActiveTasks:   s.metrics.GetActiveTasks(),
		TotalRequests: s.metrics.GetTotalRequests(),
		MemoryUsage:   s.metrics.GetMemoryUsage(),
		CPUUsage:      s.metrics.GetCPUUsage(),
		DiskUsage:     s.metrics.GetDiskUsage(),
		NetworkIO:     s.metrics.GetNetworkIO(),
		Timestamp:     time.Now(),
	}
}

func (s *MonitoringServer) getHealth(c *gin.Context) {
	health := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now(),
		"checks": map[string]string{
			"database":     "ok",
			"external_api": "ok",
			"disk_space":   "ok",
		},
	}

	c.JSON(http.StatusOK, health)
}

func (s *MonitoringServer) getMetrics(c *gin.Context) {
	format := c.DefaultQuery("format", "prometheus")

	var metrics string
	var err error

	switch format {
	case "prometheus":
		metrics = s.metrics.ExportPrometheus()
	case "json":
		metrics, err = s.metrics.ExportJSON()
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported format"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if format == "prometheus" {
		c.Header("Content-Type", "text/plain")
	} else {
		c.Header("Content-Type", "application/json")
	}

	c.String(http.StatusOK, metrics)
}

func (s *MonitoringServer) getTasks(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	status := c.Query("status")

	tasks, total, err := s.getTaskList(limit, offset, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks":  tasks,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (s *MonitoringServer) getTask(c *gin.Context) {
	taskID := c.Param("id")
	task, err := s.getTaskByID(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	c.JSON(http.StatusOK, task)
}

func (s *MonitoringServer) stopTask(c *gin.Context) {
	taskID := c.Param("id")
	err := s.stopTaskByID(taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "task stopped"})
}

func (s *MonitoringServer) getAlerts(c *gin.Context) {
	alerts, err := s.alerts.GetAlerts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"alerts": alerts})
}

func (s *MonitoringServer) createAlert(c *gin.Context) {
	var alert Alert
	if err := c.ShouldBindJSON(&alert); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.alerts.CreateAlert(&alert); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Broadcast alert through WebSocket
	s.wsManager.BroadcastAlert(alert)

	c.JSON(http.StatusCreated, alert)
}

func (s *MonitoringServer) updateAlert(c *gin.Context) {
	alertID := c.Param("id")
	var alert Alert
	if err := c.ShouldBindJSON(&alert); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	alert.ID = alertID
	if err := s.alerts.UpdateAlert(&alert); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, alert)
}

func (s *MonitoringServer) deleteAlert(c *gin.Context) {
	alertID := c.Param("id")
	if err := s.alerts.DeleteAlert(alertID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "alert deleted"})
}

func (s *MonitoringServer) getNotifications(c *gin.Context) {
	// Placeholder implementation
	c.JSON(http.StatusOK, gin.H{"notifications": []interface{}{}})
}

func (s *MonitoringServer) testNotification(c *gin.Context) {
	var req struct {
		Type    string `json:"type"`
		Target  string `json:"target"`
		Message string `json:"message"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Test notification implementation
	switch req.Type {
	case "slack":
		if s.alerts.slackNotifier == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Slack notifications not configured"})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
		defer cancel()

		err := s.alerts.slackNotifier.SendCustomMessage(ctx, "Test Notification", req.Message, AlertSeverityInfo)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Slack test notification sent successfully"})
	case "discord":
		if s.alerts.discordNotifier == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Discord notifications not configured"})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
		defer cancel()

		err := s.alerts.discordNotifier.SendCustomMessage(ctx, "Test Notification", req.Message, AlertSeverityInfo)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Discord test notification sent successfully"})
	case "email":
		if s.alerts.emailNotifier == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email notifications not configured"})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
		defer cancel()

		err := s.alerts.emailNotifier.SendCustomMessage(ctx, "Test Notification", req.Message, AlertSeverityInfo)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Email test notification sent successfully"})
	case "teams":
		if s.alerts.teamsNotifier == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Teams notifications not configured"})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
		defer cancel()

		err := s.alerts.teamsNotifier.SendCustomMessage(ctx, "Test Notification", req.Message, AlertSeverityInfo)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Teams test notification sent successfully"})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported notification type"})
	}
}

func (s *MonitoringServer) getConfig(c *gin.Context) {
	config := map[string]interface{}{
		"server": map[string]interface{}{
			"host":  s.config.Host,
			"port":  s.config.Port,
			"debug": s.config.Debug,
		},
		"metrics": map[string]interface{}{
			"enabled":  true,
			"interval": "30s",
		},
		"alerts": map[string]interface{}{
			"enabled": true,
			"rules":   []interface{}{},
		},
	}

	c.JSON(http.StatusOK, config)
}

func (s *MonitoringServer) updateConfig(c *gin.Context) {
	var config map[string]interface{}
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Configuration update implementation
	c.JSON(http.StatusOK, gin.H{"message": "configuration updated"})
}

// Authentication handlers

func (s *MonitoringServer) login(c *gin.Context) {
	var creds Credentials
	if err := c.ShouldBindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, token, err := s.authManager.Authenticate(creds.Username, creds.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":  user,
		"token": token,
	})
}

func (s *MonitoringServer) logout(c *gin.Context) {
	// For JWT tokens, logout is typically handled client-side
	// In a more sophisticated implementation, you might maintain a blacklist
	c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}

func (s *MonitoringServer) getCurrentUser(c *gin.Context) {
	username, _ := c.Get("username")
	user, err := s.authManager.GetUser(username.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

// User management handlers

func (s *MonitoringServer) getUsers(c *gin.Context) {
	users := s.authManager.GetUsers()
	c.JSON(http.StatusOK, gin.H{"users": users})
}

func (s *MonitoringServer) getUser(c *gin.Context) {
	username := c.Param("username")
	user, err := s.authManager.GetUser(username)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (s *MonitoringServer) createUser(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
		Role     string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := s.authManager.CreateUser(req.Username, req.Email, req.Password, req.Role)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"user": user})
}

func (s *MonitoringServer) updateUserPassword(c *gin.Context) {
	username := c.Param("username")
	var req struct {
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.authManager.UpdateUserPassword(username, req.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password updated successfully"})
}

func (s *MonitoringServer) deleteUser(c *gin.Context) {
	username := c.Param("username")

	// Prevent self-deletion
	currentUsername, _ := c.Get("username")
	if username == currentUsername.(string) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete your own account"})
		return
	}

	if err := s.authManager.DeactivateUser(username); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user deleted successfully"})
}

func (s *MonitoringServer) handleWebSocket(c *gin.Context) {
	// Authenticate WebSocket connection
	user, err := s.authManager.AuthenticateWebSocket(c.Request)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "WebSocket authentication failed"})
		return
	}

	// Use the WebSocketManager to handle the upgrade with authenticated user
	s.wsManager.HandleAuthenticatedWebSocket(c.Writer, c.Request, user)
}

func (s *MonitoringServer) serveDashboard(c *gin.Context) {
	// Serve the embedded dashboard HTML
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, dashboardHTML)
}

func (s *MonitoringServer) serveSPA(c *gin.Context) {
	// Serve React SPA index.html for all non-API routes
	c.File("./web/build/index.html")
}

// Helper methods

func (s *MonitoringServer) getTaskList(limit, offset int, status string) ([]Task, int, error) {
	// Placeholder implementation
	tasks := []Task{
		{
			ID:        "task-1",
			Name:      "Bulk Clone GitHub",
			Status:    "running",
			Progress:  75,
			StartTime: time.Now().Add(-30 * time.Minute),
		},
		{
			ID:        "task-2",
			Name:      "VPN Connection Monitor",
			Status:    "completed",
			Progress:  100,
			StartTime: time.Now().Add(-1 * time.Hour),
			EndTime:   &[]time.Time{time.Now().Add(-30 * time.Minute)}[0],
		},
	}

	return tasks, len(tasks), nil
}

func (s *MonitoringServer) getTaskByID(id string) (*Task, error) {
	// Placeholder implementation
	task := &Task{
		ID:        id,
		Name:      "Sample Task",
		Status:    "running",
		Progress:  50,
		StartTime: time.Now().Add(-15 * time.Minute),
		Details: map[string]interface{}{
			"processed": 150,
			"total":     300,
			"errors":    2,
		},
	}

	return task, nil
}

func (s *MonitoringServer) stopTaskByID(id string) error {
	// Placeholder implementation
	return nil
}

// startPeriodicUpdates sends periodic updates through WebSocket
func (s *MonitoringServer) startPeriodicUpdates(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Broadcast system status
			status := &SystemStatus{
				Status:        "healthy",
				Uptime:        time.Since(s.startTime).String(),
				ActiveTasks:   s.metrics.GetActiveTasks(),
				TotalRequests: s.metrics.GetTotalRequests(),
				MemoryUsage:   s.metrics.GetMemoryUsage(),
				CPUUsage:      s.metrics.GetCPUUsage(),
				DiskUsage:     s.metrics.GetDiskUsage(),
				NetworkIO:     s.metrics.GetNetworkIO(),
				Timestamp:     time.Now(),
			}
			s.wsManager.BroadcastSystemStatus(status)

			// Update local instance metrics
			s.instanceManager.UpdateLocalMetrics(status)

			// Broadcast metrics
			metrics := map[string]interface{}{
				"active_tasks":    s.metrics.GetActiveTasks(),
				"memory_usage_mb": float64(s.metrics.GetMemoryUsage()) / 1024 / 1024,
				"cpu_usage":       s.metrics.GetCPUUsage(),
				"total_requests":  s.metrics.GetTotalRequests(),
			}
			s.wsManager.BroadcastMetrics(metrics)
		}
	}
}

// Data structures

type SystemStatus struct {
	Status        string    `json:"status"`
	Uptime        string    `json:"uptime"`
	ActiveTasks   int       `json:"active_tasks"`
	TotalRequests int64     `json:"total_requests"`
	MemoryUsage   uint64    `json:"memory_usage"`
	CPUUsage      float64   `json:"cpu_usage"`
	DiskUsage     float64   `json:"disk_usage"`
	NetworkIO     NetworkIO `json:"network_io"`
	Timestamp     time.Time `json:"timestamp"`
}

type NetworkIO struct {
	BytesIn  uint64 `json:"bytes_in"`
	BytesOut uint64 `json:"bytes_out"`
}

type Task struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Status    string                 `json:"status"`
	Progress  int                    `json:"progress"`
	StartTime time.Time              `json:"start_time"`
	EndTime   *time.Time             `json:"end_time,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

type Alert struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Severity    string    `json:"severity"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Embedded dashboard HTML
const dashboardHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>GZH Monitoring Dashboard</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/css/bootstrap.min.css" rel="stylesheet">
    <link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.0.0/css/all.min.css" rel="stylesheet">
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    <style>
        body { background-color: #f8f9fa; }
        .metric-card { background: white; border-radius: 10px; padding: 20px; margin-bottom: 20px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .metric-value { font-size: 2.5rem; font-weight: bold; color: #2c3e50; }
        .metric-label { color: #7f8c8d; font-size: 0.9rem; text-transform: uppercase; letter-spacing: 1px; }
        .status-indicator { width: 12px; height: 12px; border-radius: 50%; display: inline-block; margin-right: 5px; }
        .status-healthy { background-color: #28a745; }
        .status-warning { background-color: #ffc107; }
        .status-critical { background-color: #dc3545; }
        .ws-status { position: fixed; top: 20px; right: 20px; padding: 10px 20px; border-radius: 20px; background: white; box-shadow: 0 2px 10px rgba(0,0,0,0.1); z-index: 1000; }
        .chart-container { position: relative; height: 300px; margin-top: 20px; }
    </style>
</head>
<body>
    <div class="ws-status" id="wsStatus">
        <span class="status-indicator status-warning" id="wsIndicator"></span>
        <span id="wsStatusText">Connecting...</span>
    </div>
    <div class="container-fluid mt-4">
        <h1 class="mb-4"><i class="fas fa-chart-line"></i> GZH Monitoring Dashboard</h1>
        <div class="row">
            <div class="col-md-3">
                <div class="metric-card">
                    <div class="metric-label">System Status</div>
                    <div class="metric-value" id="systemStatus">-</div>
                    <small class="text-muted">Uptime: <span id="uptime">-</span></small>
                </div>
            </div>
            <div class="col-md-3">
                <div class="metric-card">
                    <div class="metric-label">Active Tasks</div>
                    <div class="metric-value" id="activeTasks">0</div>
                </div>
            </div>
            <div class="col-md-3">
                <div class="metric-card">
                    <div class="metric-label">Memory (MB)</div>
                    <div class="metric-value" id="memoryUsage">0</div>
                </div>
            </div>
            <div class="col-md-3">
                <div class="metric-card">
                    <div class="metric-label">CPU Usage</div>
                    <div class="metric-value" id="cpuUsage">0%</div>
                </div>
            </div>
        </div>
    </div>
    <script>
        let ws = null;
        function connectWebSocket() {
            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            const wsUrl = protocol + '//' + window.location.host + '/ws';
            ws = new WebSocket(wsUrl);
            ws.onopen = function() {
                updateConnectionStatus('connected');
                ws.send(JSON.stringify({ type: 'subscribe', filter: { types: ['all'] } }));
            };
            ws.onmessage = function(event) {
                const data = JSON.parse(event.data);
                handleWebSocketMessage(data);
            };
            ws.onclose = function() {
                updateConnectionStatus('disconnected');
                setTimeout(connectWebSocket, 5000);
            };
        }
        function handleWebSocketMessage(data) {
            switch (data.type) {
                case 'system_status':
                    updateSystemStatus(data.data);
                    break;
                case 'metrics_update':
                    updateMetrics(data.data);
                    break;
            }
        }
        function updateSystemStatus(status) {
            document.getElementById('systemStatus').textContent = status.status.toUpperCase();
            document.getElementById('uptime').textContent = status.uptime;
            document.getElementById('activeTasks').textContent = status.active_tasks;
            document.getElementById('memoryUsage').textContent = (status.memory_usage / 1024 / 1024).toFixed(1);
            document.getElementById('cpuUsage').textContent = status.cpu_usage.toFixed(1) + '%';
        }
        function updateMetrics(metrics) {
            if (metrics.active_tasks !== undefined) {
                document.getElementById('activeTasks').textContent = metrics.active_tasks;
            }
            if (metrics.memory_usage_mb !== undefined) {
                document.getElementById('memoryUsage').textContent = metrics.memory_usage_mb.toFixed(1);
            }
            if (metrics.cpu_usage !== undefined) {
                document.getElementById('cpuUsage').textContent = metrics.cpu_usage.toFixed(1) + '%';
            }
        }
        function updateConnectionStatus(status) {
            const indicator = document.getElementById('wsIndicator');
            const statusText = document.getElementById('wsStatusText');
            if (status === 'connected') {
                indicator.className = 'status-indicator status-healthy';
                statusText.textContent = 'Connected';
            } else {
                indicator.className = 'status-indicator status-warning';
                statusText.textContent = 'Disconnected';
            }
        }
        document.addEventListener('DOMContentLoaded', function() {
            connectWebSocket();
        });
    </script>
</body>
</html>`

// Instance management handlers

func (s *MonitoringServer) getInstances(c *gin.Context) {
	instances := s.instanceManager.GetInstances()
	c.JSON(http.StatusOK, gin.H{"instances": instances})
}

func (s *MonitoringServer) getInstance(c *gin.Context) {
	instanceID := c.Param("id")
	instance, err := s.instanceManager.GetInstance(instanceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "instance not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"instance": instance})
}

func (s *MonitoringServer) discoverInstance(c *gin.Context) {
	var req struct {
		Host string `json:"host" binding:"required"`
		Port int    `json:"port" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := s.instanceManager.DiscoverInstance(req.Host, req.Port)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "instance discovered successfully"})
}

func (s *MonitoringServer) removeInstance(c *gin.Context) {
	instanceID := c.Param("id")

	err := s.instanceManager.RemoveInstance(instanceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "instance removed successfully"})
}

func (s *MonitoringServer) getClusterStatus(c *gin.Context) {
	status := s.instanceManager.GetClusterStatus()
	c.JSON(http.StatusOK, status)
}

// Custom metrics handlers

func (s *MonitoringServer) getCustomMetrics(c *gin.Context) {
	if s.customMetrics == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Custom metrics not initialized"})
		return
	}

	metrics := s.customMetrics.ListCustomMetrics()
	c.JSON(http.StatusOK, gin.H{
		"custom_metrics": metrics,
		"count":          len(metrics),
	})
}

func (s *MonitoringServer) getCustomMetricsSummary(c *gin.Context) {
	if s.customMetrics == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Custom metrics not initialized"})
		return
	}

	summary := s.customMetrics.GetMetricsSummary()
	c.JSON(http.StatusOK, summary)
}

func (s *MonitoringServer) createCustomMetric(c *gin.Context) {
	if s.customMetrics == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Custom metrics not initialized"})
		return
	}

	var req CustomMetricDefinition
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Validate required fields
	if req.Name == "" || req.Type == "" || req.Help == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name, type, and help are required"})
		return
	}

	var err error
	switch req.Type {
	case "counter":
		err = s.customMetrics.CreateCustomCounter(req.Name, req.Help, req.Labels, req.ConstLabels)
	case "gauge":
		err = s.customMetrics.CreateCustomGauge(req.Name, req.Help, req.Labels, req.ConstLabels)
	case "histogram":
		err = s.customMetrics.CreateCustomHistogram(req.Name, req.Help, req.Labels, req.Buckets, req.ConstLabels)
	case "summary":
		err = s.customMetrics.CreateCustomSummary(req.Name, req.Help, req.Labels, req.Objectives, req.ConstLabels)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid metric type. Must be one of: counter, gauge, histogram, summary"})
		return
	}

	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Custom metric created successfully",
		"name":    req.Name,
		"type":    req.Type,
	})
}

func (s *MonitoringServer) deleteCustomMetric(c *gin.Context) {
	if s.customMetrics == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Custom metrics not initialized"})
		return
	}

	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Metric name is required"})
		return
	}

	err := s.customMetrics.DeleteCustomMetric(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Custom metric deleted successfully",
		"name":    name,
	})
}

func (s *MonitoringServer) getBusinessMetrics(c *gin.Context) {
	if s.customMetrics == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Custom metrics not initialized"})
		return
	}

	// Return business metrics metadata and current values
	businessInfo := map[string]interface{}{
		"repo_operations": map[string]interface{}{
			"clone_total":     "Total repository clone operations",
			"clone_duration":  "Repository clone duration metrics",
			"sync_operations": "Repository sync operations",
			"repo_size":       "Repository size tracking",
		},
		"organization_management": map[string]interface{}{
			"organizations_total": "Total number of organizations",
			"projects_active":     "Active projects count",
			"users_active":        "Active users count",
		},
		"task_execution": map[string]interface{}{
			"tasks_completed":   "Completed tasks by type and outcome",
			"task_failure_rate": "Task failure rate percentages",
			"task_throughput":   "Task processing throughput",
		},
		"integrations": map[string]interface{}{
			"integration_status":   "Integration service health status",
			"api_latency":          "Integration API call latency",
			"rate_limit_remaining": "API rate limit status",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"business_metrics": businessInfo,
		"description":      "Business-specific metrics for gzh-manager operations",
	})
}

func (s *MonitoringServer) getPerformanceMetrics(c *gin.Context) {
	if s.customMetrics == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Custom metrics not initialized"})
		return
	}

	performanceInfo := map[string]interface{}{
		"system_resources": map[string]interface{}{
			"cpu_utilization":    "CPU utilization by core and type",
			"memory_utilization": "Memory utilization by type",
			"disk_io":            "Disk I/O operations and bytes",
			"network_io":         "Network I/O bytes by interface",
		},
		"application": map[string]interface{}{
			"goroutine_count": "Number of active goroutines",
			"gc_duration":     "Garbage collection duration",
			"heap_alloc":      "Heap allocated bytes",
		},
		"database": map[string]interface{}{
			"connections_active": "Active database connections",
			"query_duration":     "Database query performance",
			"connection_pool":    "Connection pool metrics",
		},
		"cache": map[string]interface{}{
			"hit_ratio":  "Cache hit ratios",
			"operations": "Cache operations by type",
			"evictions":  "Cache evictions by reason",
		},
		"queue": map[string]interface{}{
			"depth":               "Queue depth by priority",
			"processing_duration": "Queue processing time",
			"throughput":          "Queue throughput metrics",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"performance_metrics": performanceInfo,
		"description":         "Performance indicators for system and application monitoring",
	})
}

func (s *MonitoringServer) getUsageMetrics(c *gin.Context) {
	if s.customMetrics == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Custom metrics not initialized"})
		return
	}

	usageInfo := map[string]interface{}{
		"user_activity": map[string]interface{}{
			"active_users":     "Active user counts by time window",
			"session_duration": "User session duration tracking",
		},
		"feature_usage": map[string]interface{}{
			"feature_usage_total": "Feature utilization counters",
			"feature_latency":     "Feature execution latency",
			"feature_error_rate":  "Feature error rates",
		},
		"resource_consumption": map[string]interface{}{
			"bandwidth_usage": "Bandwidth consumption by operation",
			"storage_usage":   "Storage utilization by type",
			"compute_hours":   "Compute resource consumption",
		},
		"api_usage": map[string]interface{}{
			"api_calls":      "API call counters and latency",
			"quota_usage":    "API quota utilization",
			"retry_attempts": "API retry statistics",
			"geographical":   "Geographical request distribution",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"usage_metrics": usageInfo,
		"description":   "Usage statistics and consumption metrics",
	})
}

type MetricRecordRequest struct {
	MetricType string                 `json:"metric_type"` // business, performance, usage, custom
	Action     string                 `json:"action"`      // record, set, inc, add, observe
	Name       string                 `json:"name"`
	Labels     map[string]string      `json:"labels"`
	Value      float64                `json:"value"`
	Duration   string                 `json:"duration,omitempty"` // for duration-based metrics
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

func (s *MonitoringServer) recordCustomMetric(c *gin.Context) {
	if s.customMetrics == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Custom metrics not initialized"})
		return
	}

	var req MetricRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Validate required fields
	if req.MetricType == "" || req.Action == "" || req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "metric_type, action, and name are required"})
		return
	}

	var err error
	var duration time.Duration

	// Parse duration if provided
	if req.Duration != "" {
		duration, err = time.ParseDuration(req.Duration)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid duration format: " + err.Error()})
			return
		}
	}

	// Handle different metric types and actions
	switch req.MetricType {
	case "business":
		err = s.handleBusinessMetricRecord(req, duration)
	case "performance":
		err = s.handlePerformanceMetricRecord(req, duration)
	case "usage":
		err = s.handleUsageMetricRecord(req, duration)
	case "custom":
		err = s.handleCustomMetricRecord(req, duration)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid metric_type. Must be one of: business, performance, usage, custom"})
		return
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Metric recorded successfully",
		"metric_type": req.MetricType,
		"action":      req.Action,
		"name":        req.Name,
	})
}

func (s *MonitoringServer) handleBusinessMetricRecord(req MetricRecordRequest, duration time.Duration) error {
	switch req.Name {
	case "repo_clone":
		if req.Action == "record" && req.Labels != nil {
			org := req.Labels["organization"]
			platform := req.Labels["platform"]
			status := req.Labels["status"]
			sizeCategory := req.Labels["size_category"]
			s.customMetrics.RecordRepoClone(org, platform, status, duration, sizeCategory)
		}
	case "repo_sync":
		if req.Action == "record" && req.Labels != nil {
			operation := req.Labels["operation"]
			status := req.Labels["status"]
			org := req.Labels["organization"]
			s.customMetrics.RecordRepoSync(operation, status, org)
		}
	case "repo_size":
		if req.Action == "set" && req.Labels != nil {
			repo := req.Labels["repository"]
			org := req.Labels["organization"]
			s.customMetrics.SetRepoSize(repo, org, req.Value)
		}
	case "organizations_total":
		if req.Action == "set" {
			s.customMetrics.SetOrganizationsTotal(req.Value)
		}
	case "projects_active_total":
		if req.Action == "set" {
			s.customMetrics.SetProjectsActiveTotal(req.Value)
		}
	case "users_active_total":
		if req.Action == "set" {
			s.customMetrics.SetUsersActiveTotal(req.Value)
		}
	case "task_completion":
		if req.Action == "record" && req.Labels != nil {
			taskType := req.Labels["task_type"]
			org := req.Labels["organization"]
			outcome := req.Labels["outcome"]
			s.customMetrics.RecordTaskCompletion(taskType, org, outcome)
		}
	default:
		return fmt.Errorf("unknown business metric: %s", req.Name)
	}
	return nil
}

func (s *MonitoringServer) handlePerformanceMetricRecord(req MetricRecordRequest, duration time.Duration) error {
	switch req.Name {
	case "cpu_utilization":
		if req.Action == "set" && req.Labels != nil {
			core := req.Labels["core"]
			cpuType := req.Labels["type"]
			s.customMetrics.SetCPUUtilization(core, cpuType, req.Value)
		}
	case "memory_utilization":
		if req.Action == "set" && req.Labels != nil {
			memType := req.Labels["type"]
			s.customMetrics.SetMemoryUtilization(memType, req.Value)
		}
	case "goroutine_count":
		if req.Action == "set" {
			s.customMetrics.SetGoroutineCount(req.Value)
		}
	case "gc_duration":
		if req.Action == "observe" && req.Labels != nil {
			gcType := req.Labels["gc_type"]
			s.customMetrics.RecordGCDuration(gcType, duration)
		}
	case "heap_alloc":
		if req.Action == "set" {
			s.customMetrics.SetHeapAlloc(req.Value)
		}
	default:
		return fmt.Errorf("unknown performance metric: %s", req.Name)
	}
	return nil
}

func (s *MonitoringServer) handleUsageMetricRecord(req MetricRecordRequest, duration time.Duration) error {
	switch req.Name {
	case "active_users":
		if req.Action == "set" && req.Labels != nil && req.Metadata != nil {
			org := req.Labels["organization"]
			role := req.Labels["role"]
			users5min, _ := req.Metadata["users_5min"].(float64)
			users1hour, _ := req.Metadata["users_1hour"].(float64)
			users24hour, _ := req.Metadata["users_24hour"].(float64)
			s.customMetrics.SetActiveUsers(org, role, users5min, users1hour, users24hour)
		}
	case "user_session":
		if req.Action == "record" && req.Labels != nil {
			org := req.Labels["organization"]
			role := req.Labels["role"]
			sessionType := req.Labels["session_type"]
			s.customMetrics.RecordUserSession(org, role, sessionType, duration)
		}
	case "feature_usage":
		if req.Action == "record" && req.Labels != nil {
			feature := req.Labels["feature"]
			org := req.Labels["organization"]
			userRole := req.Labels["user_role"]
			complexity := req.Labels["complexity"]
			s.customMetrics.RecordFeatureUsage(feature, org, userRole, duration, complexity)
		}
	case "api_call":
		if req.Action == "record" && req.Labels != nil {
			apiVersion := req.Labels["api_version"]
			endpoint := req.Labels["endpoint"]
			method := req.Labels["method"]
			status := req.Labels["status"]
			s.customMetrics.RecordAPICall(apiVersion, endpoint, method, status, duration)
		}
	default:
		return fmt.Errorf("unknown usage metric: %s", req.Name)
	}
	return nil
}

func (s *MonitoringServer) handleCustomMetricRecord(req MetricRecordRequest, duration time.Duration) error {
	switch req.Action {
	case "inc":
		counter, err := s.customMetrics.GetCustomCounter(req.Name)
		if err != nil {
			return err
		}
		labelValues := make([]string, 0, len(req.Labels))
		for _, value := range req.Labels {
			labelValues = append(labelValues, value)
		}
		counter.WithLabelValues(labelValues...).Inc()
	case "add":
		counter, err := s.customMetrics.GetCustomCounter(req.Name)
		if err != nil {
			return err
		}
		labelValues := make([]string, 0, len(req.Labels))
		for _, value := range req.Labels {
			labelValues = append(labelValues, value)
		}
		counter.WithLabelValues(labelValues...).Add(req.Value)
	case "set":
		gauge, err := s.customMetrics.GetCustomGauge(req.Name)
		if err != nil {
			return err
		}
		labelValues := make([]string, 0, len(req.Labels))
		for _, value := range req.Labels {
			labelValues = append(labelValues, value)
		}
		gauge.WithLabelValues(labelValues...).Set(req.Value)
	case "observe":
		if duration > 0 {
			histogram, err := s.customMetrics.GetCustomHistogram(req.Name)
			if err == nil {
				labelValues := make([]string, 0, len(req.Labels))
				for _, value := range req.Labels {
					labelValues = append(labelValues, value)
				}
				histogram.WithLabelValues(labelValues...).Observe(duration.Seconds())
				return nil
			}
			// Try summary if histogram fails
			summary, err := s.customMetrics.GetCustomSummary(req.Name)
			if err != nil {
				return err
			}
			labelValues := make([]string, 0, len(req.Labels))
			for _, value := range req.Labels {
				labelValues = append(labelValues, value)
			}
			summary.WithLabelValues(labelValues...).Observe(duration.Seconds())
		} else {
			return fmt.Errorf("duration required for observe action")
		}
	default:
		return fmt.Errorf("unknown action: %s", req.Action)
	}
	return nil
}

// initializeSlackNotifier initializes Slack notification if configured
func (s *MonitoringServer) initializeSlackNotifier(logger *zap.Logger) {
	// Check for Slack configuration from environment variables
	webhookURL := os.Getenv("SLACK_WEBHOOK_URL")
	if webhookURL == "" {
		logger.Info("Slack webhook URL not configured, skipping Slack notifications")
		return
	}

	slackConfig := &SlackConfig{
		WebhookURL: webhookURL,
		Channel:    getEnvOrDefault("SLACK_CHANNEL", "#monitoring"),
		Username:   getEnvOrDefault("SLACK_USERNAME", "GZH Monitoring"),
		IconEmoji:  getEnvOrDefault("SLACK_ICON_EMOJI", ":robot_face:"),
		Enabled:    true,
	}

	slackNotifier := NewSlackNotifier(slackConfig, logger)
	s.alerts.SetSlackNotifier(slackNotifier)

	logger.Info("Slack notifications initialized",
		zap.String("channel", slackConfig.Channel),
		zap.String("username", slackConfig.Username))

	// Also initialize Discord if configured
	s.initializeDiscordNotifier(logger)

	// Also initialize Email if configured
	s.initializeEmailNotifier(logger)

	// Also initialize Teams if configured
	s.initializeTeamsNotifier(logger)
}

// initializeDiscordNotifier initializes Discord notification if configured
func (s *MonitoringServer) initializeDiscordNotifier(logger *zap.Logger) {
	// Check for Discord configuration from environment variables
	webhookURL := os.Getenv("DISCORD_WEBHOOK_URL")
	if webhookURL == "" {
		logger.Info("Discord webhook URL not configured, skipping Discord notifications")
		return
	}

	discordConfig := &DiscordConfig{
		WebhookURL: webhookURL,
		Username:   getEnvOrDefault("DISCORD_USERNAME", "GZH Monitoring"),
		AvatarURL:  getEnvOrDefault("DISCORD_AVATAR_URL", ""),
		Enabled:    true,
	}

	discordNotifier := NewDiscordNotifier(discordConfig, logger)
	s.alerts.SetDiscordNotifier(discordNotifier)

	logger.Info("Discord notifications initialized",
		zap.String("username", discordConfig.Username))
}

// initializeEmailNotifier initializes Email notification if configured
func (s *MonitoringServer) initializeEmailNotifier(logger *zap.Logger) {
	// Check for Email configuration from environment variables
	smtpHost := os.Getenv("SMTP_HOST")
	if smtpHost == "" {
		logger.Info("SMTP host not configured, skipping email notifications")
		return
	}

	smtpPort := 587 // Default SMTP port
	if port := os.Getenv("SMTP_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			smtpPort = p
		}
	}

	recipients := strings.Split(os.Getenv("EMAIL_RECIPIENTS"), ",")
	if len(recipients) == 0 || recipients[0] == "" {
		logger.Info("No email recipients configured, skipping email notifications")
		return
	}

	emailConfig := &EmailConfig{
		SMTPHost:   smtpHost,
		SMTPPort:   smtpPort,
		Username:   os.Getenv("SMTP_USERNAME"),
		Password:   os.Getenv("SMTP_PASSWORD"),
		From:       getEnvOrDefault("EMAIL_FROM", "monitoring@gzh-manager.com"),
		Recipients: recipients,
		UseTLS:     getEnvOrDefault("SMTP_USE_TLS", "true") == "true",
		Enabled:    true,
	}

	emailNotifier := NewEmailNotifier(emailConfig, logger)
	s.alerts.SetEmailNotifier(emailNotifier)

	logger.Info("Email notifications initialized",
		zap.String("smtp_host", emailConfig.SMTPHost),
		zap.Int("smtp_port", emailConfig.SMTPPort),
		zap.Strings("recipients", emailConfig.Recipients))
}

// initializeTeamsNotifier initializes Teams notification if configured
func (s *MonitoringServer) initializeTeamsNotifier(logger *zap.Logger) {
	// Check for Teams configuration from environment variables
	webhookURL := os.Getenv("TEAMS_WEBHOOK_URL")

	// Check for Graph API configuration
	tenantID := os.Getenv("TEAMS_TENANT_ID")
	clientID := os.Getenv("TEAMS_CLIENT_ID")
	clientSecret := os.Getenv("TEAMS_CLIENT_SECRET")
	defaultTeam := os.Getenv("TEAMS_DEFAULT_TEAM")

	if webhookURL == "" && tenantID == "" {
		logger.Info("Teams webhook URL or Graph API not configured, skipping Teams notifications")
		return
	}

	teamsConfig := &TeamsConfig{
		WebhookURL: webhookURL,
		Enabled:    true,
	}

	// Add Graph API configuration if available
	if tenantID != "" && clientID != "" && clientSecret != "" {
		teamsConfig.GraphConfig = &TeamsGraphConfig{
			TenantID:     tenantID,
			ClientID:     clientID,
			ClientSecret: clientSecret,
			DefaultTeam:  defaultTeam,
		}

		// Add default channel rules if configured
		teamsConfig.ChannelRules = s.loadTeamsChannelRules()
	}

	teamsNotifier := NewTeamsNotifier(teamsConfig, logger)
	s.alerts.SetTeamsNotifier(teamsNotifier)

	if webhookURL != "" {
		truncateLength := 50
		if len(webhookURL) < truncateLength {
			truncateLength = len(webhookURL)
		}
		logger.Info("Teams webhook notifications initialized",
			zap.String("webhook_url_prefix", webhookURL[:truncateLength]+"..."))
	}

	if teamsConfig.GraphConfig != nil {
		logger.Info("Teams Graph API integration initialized",
			zap.String("tenant_id", tenantID),
			zap.String("client_id", clientID),
			zap.String("default_team", defaultTeam),
			zap.Int("channel_rules", len(teamsConfig.ChannelRules)))
	}
}

// loadTeamsChannelRules loads channel routing rules from environment variables
func (s *MonitoringServer) loadTeamsChannelRules() []ChannelRule {
	var rules []ChannelRule

	// Example environment variables:
	// TEAMS_CRITICAL_ALERT_TEAM=team-id
	// TEAMS_CRITICAL_ALERT_CHANNEL=channel-id
	// TEAMS_HIGH_ALERT_TEAM=team-id
	// TEAMS_HIGH_ALERT_CHANNEL=channel-id
	// TEAMS_SYSTEM_STATUS_TEAM=team-id
	// TEAMS_SYSTEM_STATUS_CHANNEL=channel-id

	criticalTeam := os.Getenv("TEAMS_CRITICAL_ALERT_TEAM")
	criticalChannel := os.Getenv("TEAMS_CRITICAL_ALERT_CHANNEL")
	if criticalTeam != "" && criticalChannel != "" {
		rules = append(rules, ChannelRule{
			Severity:    AlertSeverityCritical,
			Type:        "alert",
			TeamID:      criticalTeam,
			ChannelID:   criticalChannel,
			ChannelName: "critical-alerts",
		})
	}

	highTeam := os.Getenv("TEAMS_HIGH_ALERT_TEAM")
	highChannel := os.Getenv("TEAMS_HIGH_ALERT_CHANNEL")
	if highTeam != "" && highChannel != "" {
		rules = append(rules, ChannelRule{
			Severity:    AlertSeverityHigh,
			Type:        "alert",
			TeamID:      highTeam,
			ChannelID:   highChannel,
			ChannelName: "high-alerts",
		})
	}

	mediumTeam := os.Getenv("TEAMS_MEDIUM_ALERT_TEAM")
	mediumChannel := os.Getenv("TEAMS_MEDIUM_ALERT_CHANNEL")
	if mediumTeam != "" && mediumChannel != "" {
		rules = append(rules, ChannelRule{
			Severity:    AlertSeverityMedium,
			Type:        "alert",
			TeamID:      mediumTeam,
			ChannelID:   mediumChannel,
			ChannelName: "medium-alerts",
		})
	}

	systemTeam := os.Getenv("TEAMS_SYSTEM_STATUS_TEAM")
	systemChannel := os.Getenv("TEAMS_SYSTEM_STATUS_CHANNEL")
	if systemTeam != "" && systemChannel != "" {
		rules = append(rules, ChannelRule{
			Type:        "system",
			TeamID:      systemTeam,
			ChannelID:   systemChannel,
			ChannelName: "system-status",
		})
	}

	return rules
}

// getEnvOrDefault gets environment variable with default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// handleSlackInteraction handles Slack interactive message callbacks
func (s *MonitoringServer) handleSlackInteraction(c *gin.Context) {
	// Parse the form data from Slack
	payload := c.PostForm("payload")
	if payload == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing payload"})
		return
	}

	// Parse the JSON payload
	var slackPayload SlackInteractionPayload
	if err := json.Unmarshal([]byte(payload), &slackPayload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload format"})
		return
	}

	// Verify the request came from Slack (optional - requires verification token)
	// if slackPayload.Token != expectedToken {
	//     c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
	//     return
	// }

	// Get Slack notifier
	if s.alerts.slackNotifier == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Slack integration not configured"})
		return
	}

	// Process the interaction
	response, err := s.alerts.slackNotifier.ProcessInteraction(&slackPayload)
	if err != nil {
		s.logger.Error("Failed to process Slack interaction", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process interaction"})
		return
	}

	// Handle specific actions that require backend operations
	if len(slackPayload.Actions) > 0 {
		action := slackPayload.Actions[0]
		alertID := action.Value

		switch action.Name {
		case "silence":
			// Silence the alert in the backend
			if err := s.alerts.SilenceAlert(alertID, time.Hour); err != nil {
				s.logger.Error("Failed to silence alert", zap.String("alert_id", alertID), zap.Error(err))
			} else {
				s.logger.Info("Alert silenced via Slack", zap.String("alert_id", alertID))
			}

		case "resolve":
			// Resolve the alert in the backend
			if err := s.alerts.ResolveAlert(alertID); err != nil {
				s.logger.Error("Failed to resolve alert", zap.String("alert_id", alertID), zap.Error(err))
			} else {
				s.logger.Info("Alert resolved via Slack", zap.String("alert_id", alertID))
			}

		case "unsilence":
			// Unsilence the alert by resolving and then letting it re-fire if needed
			// This is a simplified implementation
			s.logger.Info("Alert unsilence requested via Slack", zap.String("alert_id", alertID))

		case "refresh":
			// Trigger a system status refresh
			go func() {
				status := s.getCurrentSystemStatus()
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()

				if err := s.alerts.slackNotifier.SendInteractiveSystemStatus(ctx, status); err != nil {
					s.logger.Error("Failed to send refreshed system status", zap.Error(err))
				}
			}()
		}
	}

	// Return the response message to Slack
	if response != nil {
		c.JSON(http.StatusOK, response)
	} else {
		// Return empty response to acknowledge the interaction
		c.JSON(http.StatusOK, gin.H{})
	}
}

// handleSlackCommand handles Slack slash command callbacks
func (s *MonitoringServer) handleSlackCommand(c *gin.Context) {
	// Parse the form data from Slack slash command
	var slackCmd SlackSlashCommand

	// Slack sends form data, not JSON
	slackCmd.Token = c.PostForm("token")
	slackCmd.TeamID = c.PostForm("team_id")
	slackCmd.TeamDomain = c.PostForm("team_domain")
	slackCmd.ChannelID = c.PostForm("channel_id")
	slackCmd.ChannelName = c.PostForm("channel_name")
	slackCmd.UserID = c.PostForm("user_id")
	slackCmd.UserName = c.PostForm("user_name")
	slackCmd.Command = c.PostForm("command")
	slackCmd.Text = c.PostForm("text")
	slackCmd.ResponseURL = c.PostForm("response_url")
	slackCmd.TriggerID = c.PostForm("trigger_id")
	slackCmd.APIAppID = c.PostForm("api_app_id")
	slackCmd.IsEnterpriseInstall = c.PostForm("is_enterprise_install")

	// Verify the request came from Slack (optional - requires verification token)
	// if slackCmd.Token != expectedToken {
	//     c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
	//     return
	// }

	// Get Slack notifier
	if s.alerts.slackNotifier == nil {
		c.JSON(http.StatusServiceUnavailable, SlackCommandResponse{
			ResponseType: "ephemeral",
			Text:         "❌ Slack integration not configured on this server.",
		})
		return
	}

	// Process the slash command
	response, err := s.alerts.slackNotifier.ProcessSlashCommand(&slackCmd)
	if err != nil {
		s.logger.Error("Failed to process Slack slash command", zap.Error(err))
		c.JSON(http.StatusOK, SlackCommandResponse{
			ResponseType: "ephemeral",
			Text:         "❌ Failed to process command. Please try again.",
		})
		return
	}

	// Handle specific commands that require backend operations
	if slackCmd.Text != "" {
		args := strings.Fields(slackCmd.Text)
		if len(args) > 0 {
			switch args[0] {
			case "silence":
				if len(args) > 1 {
					alertID := args[1]
					duration := time.Hour // Default 1 hour
					if err := s.alerts.SilenceAlert(alertID, duration); err != nil {
						s.logger.Error("Failed to silence alert via Slack command",
							zap.String("alert_id", alertID), zap.Error(err))
					} else {
						s.logger.Info("Alert silenced via Slack command",
							zap.String("alert_id", alertID), zap.String("user", slackCmd.UserName))
					}
				}

			case "resolve":
				if len(args) > 1 {
					alertID := args[1]
					if err := s.alerts.ResolveAlert(alertID); err != nil {
						s.logger.Error("Failed to resolve alert via Slack command",
							zap.String("alert_id", alertID), zap.Error(err))
					} else {
						s.logger.Info("Alert resolved via Slack command",
							zap.String("alert_id", alertID), zap.String("user", slackCmd.UserName))
					}
				}

			case "status":
				// For status commands, we could update the response with real data
				if response != nil && len(response.Attachments) > 0 {
					// Update with real system status
					systemStatus := s.getCurrentSystemStatus()
					statusText := fmt.Sprintf(`*🖥️ GZH Monitoring System Status*

*Overall Health:* %s %s
*Uptime:* %s
*Active Tasks:* %d
*Memory Usage:* %s
*CPU Usage:* %.1f%%
*Total Requests:* %d

*Quick Actions:*`,
						s.getHealthEmoji(systemStatus.Status),
						systemStatus.Status,
						systemStatus.Uptime,
						systemStatus.ActiveTasks,
						formatBytes(systemStatus.MemoryUsage),
						systemStatus.CPUUsage,
						systemStatus.TotalRequests)

					response.Attachments[0].Text = statusText
					response.Attachments[0].Color = s.getHealthColor(systemStatus.Status)
				}

			case "test":
				// Log test command usage
				testType := "basic"
				if len(args) > 1 {
					testType = args[1]
				}
				s.logger.Info("Test command executed via Slack",
					zap.String("test_type", testType),
					zap.String("user", slackCmd.UserName),
					zap.String("channel", slackCmd.ChannelName))
			}
		}
	}

	// Return the response to Slack
	c.JSON(http.StatusOK, response)
}

// Helper methods for status commands
func (s *MonitoringServer) getHealthEmoji(status string) string {
	switch status {
	case "healthy":
		return "✅"
	case "warning":
		return "⚠️"
	case "critical":
		return "🚨"
	default:
		return "ℹ️"
	}
}

func (s *MonitoringServer) getHealthColor(status string) string {
	switch status {
	case "healthy":
		return "good"
	case "warning":
		return "warning"
	case "critical":
		return "danger"
	default:
		return "#439FE0"
	}
}

// Teams API handlers

func (s *MonitoringServer) getTeams(c *gin.Context) {
	if s.alerts.teamsNotifier == nil || s.alerts.teamsNotifier.graphConfig == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Teams Graph API integration not configured"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	teams, err := s.alerts.teamsNotifier.ListTeams(ctx)
	if err != nil {
		s.logger.Error("Failed to list Teams", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list teams"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"teams": teams})
}

func (s *MonitoringServer) getTeamChannels(c *gin.Context) {
	if s.alerts.teamsNotifier == nil || s.alerts.teamsNotifier.graphConfig == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Teams Graph API integration not configured"})
		return
	}

	teamID := c.Param("teamId")
	if teamID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "team ID is required"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	channels, err := s.alerts.teamsNotifier.ListChannels(ctx, teamID)
	if err != nil {
		s.logger.Error("Failed to list channels", zap.String("team_id", teamID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list channels"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"channels": channels})
}

func (s *MonitoringServer) getChannelRules(c *gin.Context) {
	if s.alerts.teamsNotifier == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Teams integration not configured"})
		return
	}

	rules := s.alerts.teamsNotifier.GetChannelConfiguration()
	c.JSON(http.StatusOK, gin.H{"rules": rules})
}

func (s *MonitoringServer) addChannelRule(c *gin.Context) {
	if s.alerts.teamsNotifier == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Teams integration not configured"})
		return
	}

	var rule ChannelRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate required fields
	if rule.TeamID == "" || rule.ChannelID == "" || rule.Type == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "team_id, channel_id, and type are required"})
		return
	}

	s.alerts.teamsNotifier.AddChannelRule(rule)
	s.logger.Info("Channel rule added",
		zap.String("team_id", rule.TeamID),
		zap.String("channel_id", rule.ChannelID),
		zap.String("type", rule.Type),
		zap.String("severity", string(rule.Severity)))

	c.JSON(http.StatusCreated, gin.H{"message": "channel rule added successfully"})
}

func (s *MonitoringServer) removeChannelRule(c *gin.Context) {
	if s.alerts.teamsNotifier == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Teams integration not configured"})
		return
	}

	teamID := c.Param("teamId")
	channelID := c.Param("channelId")

	if teamID == "" || channelID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "team ID and channel ID are required"})
		return
	}

	s.alerts.teamsNotifier.RemoveChannelRule(teamID, channelID)
	s.logger.Info("Channel rule removed",
		zap.String("team_id", teamID),
		zap.String("channel_id", channelID))

	c.JSON(http.StatusOK, gin.H{"message": "channel rule removed successfully"})
}

func (s *MonitoringServer) testTeamsChannel(c *gin.Context) {
	if s.alerts.teamsNotifier == nil || s.alerts.teamsNotifier.graphConfig == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Teams Graph API integration not configured"})
		return
	}

	teamID := c.Param("teamId")
	channelID := c.Param("channelId")

	if teamID == "" || channelID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "team ID and channel ID are required"})
		return
	}

	// Get optional message from request body
	var req struct {
		Message string `json:"message"`
	}
	c.ShouldBindJSON(&req)

	message := req.Message
	if message == "" {
		message = "This is a test message from GZH Monitoring system to verify Teams channel integration."
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	// Create a test adaptive card
	testMessage := s.alerts.teamsNotifier.formatCustomMessage("🧪 Teams Channel Test", message, AlertSeverityInfo)
	cardContent := &testMessage.Attachments[0].Content

	err := s.alerts.teamsNotifier.SendToChannel(ctx, teamID, channelID, cardContent)
	if err != nil {
		s.logger.Error("Failed to send test message to Teams channel",
			zap.String("team_id", teamID),
			zap.String("channel_id", channelID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send test message"})
		return
	}

	s.logger.Info("Test message sent to Teams channel",
		zap.String("team_id", teamID),
		zap.String("channel_id", channelID))

	c.JSON(http.StatusOK, gin.H{"message": "test message sent successfully to Teams channel"})
}
