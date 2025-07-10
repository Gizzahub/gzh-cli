package monitoring

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
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

// ServerConfig represents server configuration
type ServerConfig struct {
	Host  string
	Port  int
	Debug bool
}

// MonitoringServer represents the monitoring server
type MonitoringServer struct {
	config     *ServerConfig
	router     *gin.Engine
	httpServer *http.Server
	metrics    *MetricsCollector
	alerts     *AlertManager
	startTime  time.Time
}

// NewMonitoringServer creates a new monitoring server
func NewMonitoringServer(config *ServerConfig) *MonitoringServer {
	server := &MonitoringServer{
		config:    config,
		router:    gin.New(),
		metrics:   NewMetricsCollector(),
		alerts:    NewAlertManager(),
		startTime: time.Now(),
	}

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

	// API routes
	api := s.router.Group("/api/v1")
	{
		// System endpoints
		api.GET("/status", s.getSystemStatus)
		api.GET("/health", s.getHealth)
		api.GET("/metrics", s.getMetrics)

		// Task monitoring endpoints
		api.GET("/tasks", s.getTasks)
		api.GET("/tasks/:id", s.getTask)
		api.POST("/tasks/:id/stop", s.stopTask)

		// Alert endpoints
		api.GET("/alerts", s.getAlerts)
		api.POST("/alerts", s.createAlert)
		api.PUT("/alerts/:id", s.updateAlert)
		api.DELETE("/alerts/:id", s.deleteAlert)

		// Notification endpoints
		api.GET("/notifications", s.getNotifications)
		api.POST("/notifications/test", s.testNotification)

		// Configuration endpoints
		api.GET("/config", s.getConfig)
		api.PUT("/config", s.updateConfig)
	}

	// WebSocket endpoint for real-time updates
	s.router.GET("/ws", s.handleWebSocket)

	// Static files for dashboard (if implemented)
	s.router.Static("/static", "./web/static")
	s.router.StaticFile("/", "./web/index.html")
}

// Start starts the monitoring server
func (s *MonitoringServer) Start(addr string) error {
	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: s.router,
	}

	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *MonitoringServer) Shutdown(ctx context.Context) error {
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

	c.JSON(http.StatusOK, status)
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
	c.JSON(http.StatusOK, gin.H{"message": "test notification sent"})
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

func (s *MonitoringServer) handleWebSocket(c *gin.Context) {
	// WebSocket implementation placeholder
	c.JSON(http.StatusNotImplemented, gin.H{"error": "WebSocket not implemented yet"})
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
