// Package api provides REST API server functionality for GZH Manager
package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gizzahub/gzh-manager-go/pkg/gzhclient"
	"github.com/gizzahub/gzh-manager-go/pkg/i18n"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/swagger"
)

// Server represents the REST API server
type Server struct {
	app    *fiber.App
	config *Config
	client *gzhclient.Client
}

// Config holds server configuration
type Config struct {
	Host          string `json:"host"`
	Port          int    `json:"port"`
	Environment   string `json:"environment"`
	CORSOrigins   string `json:"cors_origins"`
	EnableSwagger bool   `json:"enable_swagger"`
	EnableAuth    bool   `json:"enable_auth"`
	RateLimit     int    `json:"rate_limit"`
	ReadTimeout   int    `json:"read_timeout"`
	WriteTimeout  int    `json:"write_timeout"`
	LogLevel      string `json:"log_level"`
}

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Error   string                 `json:"error"`
	Message string                 `json:"message"`
	Code    int                    `json:"code"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// SuccessResponse represents a standard success response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// HealthResponse represents health check response
type HealthResponse struct {
	Status    string                   `json:"status"`
	Version   string                   `json:"version"`
	Timestamp time.Time                `json:"timestamp"`
	Uptime    string                   `json:"uptime"`
	Services  map[string]ServiceHealth `json:"services"`
}

// ServiceHealth represents individual service health
type ServiceHealth struct {
	Status      string    `json:"status"`
	LastChecked time.Time `json:"last_checked"`
	Message     string    `json:"message,omitempty"`
}

// PaginationRequest represents pagination parameters
type PaginationRequest struct {
	Page    int    `query:"page" json:"page"`
	PerPage int    `query:"per_page" json:"per_page"`
	Sort    string `query:"sort" json:"sort"`
	Order   string `query:"order" json:"order"`
}

// PaginationResponse represents pagination metadata
type PaginationResponse struct {
	Page       int  `json:"page"`
	PerPage    int  `json:"per_page"`
	Total      int  `json:"total"`
	TotalPages int  `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}

// DefaultConfig returns a default server configuration
func DefaultConfig() *Config {
	return &Config{
		Host:          "localhost",
		Port:          8080,
		Environment:   "development",
		CORSOrigins:   "*",
		EnableSwagger: true,
		EnableAuth:    false,
		RateLimit:     100,
		ReadTimeout:   30,
		WriteTimeout:  30,
		LogLevel:      "info",
	}
}

// NewServer creates a new API server instance
func NewServer(config *Config, client *gzhclient.Client) *Server {
	if config == nil {
		config = DefaultConfig()
	}

	// Create Fiber app with configuration
	app := fiber.New(fiber.Config{
		AppName:           "GZH Manager API",
		ReadTimeout:       time.Duration(config.ReadTimeout) * time.Second,
		WriteTimeout:      time.Duration(config.WriteTimeout) * time.Second,
		EnablePrintRoutes: config.Environment == "development",
		ErrorHandler:      errorHandler,
	})

	server := &Server{
		app:    app,
		config: config,
		client: client,
	}

	// Setup middleware
	server.setupMiddleware()

	// Setup routes
	server.setupRoutes()

	return server
}

// setupMiddleware configures all middleware
func (s *Server) setupMiddleware() {
	// Request ID middleware
	s.app.Use(requestid.New())

	// Logger middleware
	if s.config.LogLevel != "silent" {
		s.app.Use(logger.New(logger.Config{
			Format: "[${time}] ${status} - ${method} ${path} (${latency})\n",
		}))
	}

	// Recovery middleware
	s.app.Use(recover.New())

	// CORS middleware
	s.app.Use(cors.New(cors.Config{
		AllowOrigins:     s.config.CORSOrigins,
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
	}))

	// Custom middleware for i18n
	s.app.Use(s.i18nMiddleware)

	// Authentication middleware (if enabled)
	if s.config.EnableAuth {
		s.app.Use(s.authMiddleware)
	}
}

// setupRoutes configures all API routes
func (s *Server) setupRoutes() {
	// Health check endpoint (no auth required)
	s.app.Get("/health", s.healthHandler)

	// API v1 routes
	api := s.app.Group("/api/v1")

	// Bulk clone endpoints
	api.Post("/bulk-clone", s.bulkCloneHandler)
	api.Get("/bulk-clone/status/:id", s.bulkCloneStatusHandler)

	// Plugin management endpoints
	plugins := api.Group("/plugins")
	plugins.Get("/", s.listPluginsHandler)
	plugins.Post("/:name/execute", s.executePluginHandler)
	plugins.Get("/:name", s.getPluginHandler)
	plugins.Put("/:name/enable", s.enablePluginHandler)
	plugins.Put("/:name/disable", s.disablePluginHandler)

	// Configuration endpoints
	config := api.Group("/config")
	config.Get("/", s.getConfigHandler)
	config.Put("/", s.updateConfigHandler)
	config.Post("/validate", s.validateConfigHandler)

	// System information endpoints
	system := api.Group("/system")
	system.Get("/info", s.getSystemInfoHandler)
	system.Get("/metrics", s.getMetricsHandler)
	system.Get("/logs", s.getLogsHandler)

	// i18n endpoints
	i18nGroup := api.Group("/i18n")
	i18nGroup.Get("/languages", s.getLanguagesHandler)
	i18nGroup.Put("/language/:lang", s.setLanguageHandler)
	i18nGroup.Get("/messages/:lang", s.getMessagesHandler)

	// Swagger documentation
	if s.config.EnableSwagger {
		s.app.Get("/swagger/*", swagger.HandlerDefault)
	}

	// Static files (if needed)
	s.app.Static("/static", "./web/static")

	// 404 handler
	s.app.Use(s.notFoundHandler)
}

// Start starts the API server
func (s *Server) Start() error {
	// Setup graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
		log.Printf("🚀 GZH Manager API server starting on %s", addr)

		if err := s.app.Listen(addr); err != nil {
			log.Printf("❌ Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-c
	log.Println("🛑 Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return s.app.ShutdownWithContext(ctx)
}

// Stop stops the API server
func (s *Server) Stop() error {
	return s.app.Shutdown()
}

// GetApp returns the underlying Fiber app
func (s *Server) GetApp() *fiber.App {
	return s.app
}

// Middleware implementations

// i18nMiddleware handles internationalization
func (s *Server) i18nMiddleware(c *fiber.Ctx) error {
	// Get language from header, query param, or default
	lang := c.Get("Accept-Language", "en")
	if queryLang := c.Query("lang"); queryLang != "" {
		lang = queryLang
	}

	// Parse Accept-Language header
	if strings.Contains(lang, ",") {
		langs := strings.Split(lang, ",")
		if len(langs) > 0 {
			lang = strings.TrimSpace(strings.Split(langs[0], ";")[0])
		}
	}

	// Set language in context
	c.Locals("language", lang)

	// Initialize i18n if not already done
	if i18n.GetManager() == nil {
		config := i18n.DefaultConfig()
		i18n.Init(config)
	}

	// Set current language
	i18n.SetLanguage(lang)

	return c.Next()
}

// authMiddleware handles authentication
func (s *Server) authMiddleware(c *fiber.Ctx) error {
	// Skip auth for certain paths
	skipPaths := []string{"/health", "/swagger", "/static"}
	for _, path := range skipPaths {
		if strings.HasPrefix(c.Path(), path) {
			return c.Next()
		}
	}

	// Get authorization header
	auth := c.Get("Authorization")
	if auth == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{
			Error:   "unauthorized",
			Message: "Authorization header required",
			Code:    fiber.StatusUnauthorized,
		})
	}

	// Validate token (implement your auth logic here)
	if !s.validateAuthToken(auth) {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{
			Error:   "unauthorized",
			Message: "Invalid authorization token",
			Code:    fiber.StatusUnauthorized,
		})
	}

	return c.Next()
}

// validateAuthToken validates the authentication token
func (s *Server) validateAuthToken(token string) bool {
	// Implement your token validation logic here
	// For now, accept any non-empty token in development
	if s.config.Environment == "development" {
		return strings.HasPrefix(token, "Bearer ")
	}

	// In production, implement proper JWT validation or API key checking
	return false
}

// errorHandler handles global errors
func errorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "Internal Server Error"

	// Check if it's a Fiber error
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	// Log error
	log.Printf("❌ API Error [%d]: %v", code, err)

	return c.Status(code).JSON(ErrorResponse{
		Error:   http.StatusText(code),
		Message: message,
		Code:    code,
	})
}

// Helper functions

// validatePagination validates and sets default pagination parameters
func validatePagination(req *PaginationRequest) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PerPage < 1 || req.PerPage > 100 {
		req.PerPage = 20
	}
	if req.Sort == "" {
		req.Sort = "created_at"
	}
	if req.Order != "asc" && req.Order != "desc" {
		req.Order = "desc"
	}
}

// calculatePagination calculates pagination metadata
func calculatePagination(page, perPage, total int) PaginationResponse {
	totalPages := (total + perPage - 1) / perPage

	return PaginationResponse{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}

// parseIntQuery parses integer query parameter with default
func parseIntQuery(c *fiber.Ctx, key string, defaultValue int) int {
	if value := c.Query(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// respondWithError sends a standardized error response
func respondWithError(c *fiber.Ctx, code int, err string, message string) error {
	return c.Status(code).JSON(ErrorResponse{
		Error:   err,
		Message: message,
		Code:    code,
	})
}

// respondWithSuccess sends a standardized success response
func respondWithSuccess(c *fiber.Ctx, data interface{}, message string) error {
	return c.JSON(SuccessResponse{
		Success: true,
		Data:    data,
		Message: message,
	})
}
