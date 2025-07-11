package monitoring

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// LoggingAPIServer provides HTTP API for log ingestion and management
type LoggingAPIServer struct {
	logger     *CentralizedLogger
	config     *CentralizedLoggingConfig
	router     *gin.Engine
	httpServer *http.Server
	zapLogger  *zap.Logger
}

// NewLoggingAPIServer creates a new logging API server
func NewLoggingAPIServer(logger *CentralizedLogger, config *CentralizedLoggingConfig) *LoggingAPIServer {
	zapLogger, _ := zap.NewProduction()

	server := &LoggingAPIServer{
		logger:    logger,
		config:    config,
		router:    gin.New(),
		zapLogger: zapLogger,
	}

	server.setupRoutes()
	return server
}

// setupRoutes configures the API routes
func (s *LoggingAPIServer) setupRoutes() {
	// Middleware
	s.router.Use(gin.Recovery())
	s.router.Use(gin.Logger())
	s.router.Use(s.corsMiddleware())

	// API routes
	api := s.router.Group("/api/v1/logging")
	{
		// Log ingestion endpoints
		api.POST("/ingest", s.ingestLogs)
		api.POST("/ingest/batch", s.ingestLogsBatch)

		// Log streaming endpoints
		api.GET("/stream", s.streamLogs)

		// Configuration endpoints
		api.GET("/config", s.getLoggingConfig)
		api.PUT("/config", s.updateLoggingConfig)

		// Statistics endpoints
		api.GET("/stats", s.getLoggingStats)
		api.GET("/health", s.getLoggingHealth)

		// Output management
		api.GET("/outputs", s.listOutputs)
		api.POST("/outputs/:name/flush", s.flushOutput)

		// Shipper management
		api.GET("/shippers", s.listShippers)
		api.POST("/shippers/:name/test", s.testShipper)

		// Search endpoints
		api.POST("/search", s.searchLogs)
		api.GET("/fields", s.getFields)
		api.GET("/indices", s.getIndices)
		api.GET("/indices/:name/stats", s.getIndexStats)
	}

	// Health endpoint
	s.router.GET("/health", s.getHealth)
}

// Start starts the logging API server
func (s *LoggingAPIServer) Start(addr string) error {
	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: s.router,
	}

	s.zapLogger.Info("Starting logging API server", zap.String("address", addr))
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *LoggingAPIServer) Shutdown(ctx context.Context) error {
	s.zapLogger.Info("Shutting down logging API server")
	return s.httpServer.Shutdown(ctx)
}

// API handler functions

func (s *LoggingAPIServer) ingestLogs(c *gin.Context) {
	var entry LogEntry
	if err := c.ShouldBindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set timestamp if not provided
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	// Process the log entry
	if err := s.logger.Log(&entry); err != nil {
		s.zapLogger.Error("Failed to process log entry", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process log entry"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "log entry processed successfully"})
}

func (s *LoggingAPIServer) ingestLogsBatch(c *gin.Context) {
	var entries []LogEntry
	if err := c.ShouldBindJSON(&entries); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	processed := 0
	failed := 0

	for _, entry := range entries {
		// Set timestamp if not provided
		if entry.Timestamp.IsZero() {
			entry.Timestamp = time.Now()
		}

		// Process the log entry
		if err := s.logger.Log(&entry); err != nil {
			s.zapLogger.Error("Failed to process log entry in batch", zap.Error(err))
			failed++
		} else {
			processed++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "batch processing completed",
		"processed": processed,
		"failed":    failed,
		"total":     len(entries),
	})
}

func (s *LoggingAPIServer) streamLogs(c *gin.Context) {
	wsManager := s.logger.GetWebSocketManager()
	if wsManager == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "real-time streaming not enabled"})
		return
	}

	// Upgrade to WebSocket connection
	wsManager.HandleWebSocket(c.Writer, c.Request)
}

func (s *LoggingAPIServer) getLoggingConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"config": s.config})
}

func (s *LoggingAPIServer) updateLoggingConfig(c *gin.Context) {
	var newConfig CentralizedLoggingConfig
	if err := c.ShouldBindJSON(&newConfig); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate configuration
	if err := validateLoggingConfig(&newConfig); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Note: In a real implementation, you would need to restart/reconfigure the logger
	c.JSON(http.StatusOK, gin.H{"message": "configuration updated (restart required)"})
}

func (s *LoggingAPIServer) getLoggingStats(c *gin.Context) {
	stats := s.logger.GetStats()
	c.JSON(http.StatusOK, gin.H{"stats": stats})
}

func (s *LoggingAPIServer) getLoggingHealth(c *gin.Context) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"outputs":   len(s.logger.outputs),
		"shippers":  len(s.logger.shippers),
	}

	c.JSON(http.StatusOK, gin.H{"health": health})
}

func (s *LoggingAPIServer) listOutputs(c *gin.Context) {
	s.logger.mutex.RLock()
	defer s.logger.mutex.RUnlock()

	outputs := make([]string, 0, len(s.logger.outputs))
	for name := range s.logger.outputs {
		outputs = append(outputs, name)
	}

	c.JSON(http.StatusOK, gin.H{"outputs": outputs})
}

func (s *LoggingAPIServer) flushOutput(c *gin.Context) {
	outputName := c.Param("name")

	s.logger.mutex.RLock()
	output, exists := s.logger.outputs[outputName]
	s.logger.mutex.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "output not found"})
		return
	}

	if err := output.Flush(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "output flushed successfully"})
}

func (s *LoggingAPIServer) listShippers(c *gin.Context) {
	s.logger.mutex.RLock()
	defer s.logger.mutex.RUnlock()

	shippers := make([]string, 0, len(s.logger.shippers))
	for name := range s.logger.shippers {
		shippers = append(shippers, name)
	}

	c.JSON(http.StatusOK, gin.H{"shippers": shippers})
}

func (s *LoggingAPIServer) testShipper(c *gin.Context) {
	shipperName := c.Param("name")

	s.logger.mutex.RLock()
	shipper, exists := s.logger.shippers[shipperName]
	s.logger.mutex.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "shipper not found"})
		return
	}

	// Create test log entries
	testEntries := []*LogEntry{
		{
			Timestamp: time.Now(),
			Level:     "info",
			Message:   fmt.Sprintf("Test message from %s shipper", shipperName),
			Logger:    "api-test",
			Fields: map[string]interface{}{
				"test":      true,
				"shipper":   shipperName,
				"timestamp": time.Now().Unix(),
			},
		},
	}

	if err := shipper.Ship(testEntries); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "test entries shipped successfully"})
}

// Search endpoint handlers

func (s *LoggingAPIServer) searchLogs(c *gin.Context) {
	if s.logger.indexer == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "search indexing not enabled"})
		return
	}

	var query SearchQuery
	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Apply default limits
	if query.Limit == 0 {
		query.Limit = 50
	}
	if query.Limit > 1000 {
		query.Limit = 1000 // Maximum safety limit
	}

	result, err := s.logger.indexer.Search(&query)
	if err != nil {
		s.zapLogger.Error("Search error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "search failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": result})
}

func (s *LoggingAPIServer) getFields(c *gin.Context) {
	if s.logger.indexer == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "search indexing not enabled"})
		return
	}

	// Get available fields from the indexer
	stats := s.logger.indexer.GetStats()
	fields := []string{
		"timestamp", "level", "logger", "message", "trace_id", "span_id",
	}

	c.JSON(http.StatusOK, gin.H{
		"fields": fields,
		"stats":  stats,
	})
}

func (s *LoggingAPIServer) getIndices(c *gin.Context) {
	if s.logger.indexer == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "search indexing not enabled"})
		return
	}

	stats := s.logger.indexer.GetStats()
	indices := []map[string]interface{}{
		{
			"name":   stats.Name,
			"health": stats.Health,
			"docs":   stats.DocCount,
			"size":   stats.Size,
		},
	}

	c.JSON(http.StatusOK, gin.H{"indices": indices})
}

func (s *LoggingAPIServer) getIndexStats(c *gin.Context) {
	if s.logger.indexer == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "search indexing not enabled"})
		return
	}

	indexName := c.Param("name")
	stats := s.logger.indexer.GetStats()

	if stats.Name != indexName {
		c.JSON(http.StatusNotFound, gin.H{"error": "index not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"stats": stats})
}

func (s *LoggingAPIServer) getHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now(),
		"service":   "logging-api",
	})
}

// Middleware functions

func (s *LoggingAPIServer) corsMiddleware() gin.HandlerFunc {
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
