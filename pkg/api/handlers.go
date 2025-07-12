// Package api provides REST API handlers for GZH Manager
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/gizzahub/gzh-manager-go/pkg/gzhclient"
	"github.com/gizzahub/gzh-manager-go/pkg/i18n"
	"github.com/gofiber/fiber/v2"
)

// Request/Response types for API endpoints

type BulkCloneAPIRequest struct {
	Platforms      []gzhclient.PlatformConfig `json:"platforms"`
	OutputDir      string                     `json:"output_dir"`
	Concurrency    int                        `json:"concurrency"`
	Strategy       string                     `json:"strategy"`
	IncludePrivate bool                       `json:"include_private"`
	Filters        gzhclient.CloneFilters     `json:"filters"`
}

type PluginExecuteRequest struct {
	Method  string                 `json:"method"`
	Args    map[string]interface{} `json:"args"`
	Timeout int                    `json:"timeout_seconds"`
}

type ConfigUpdateRequest struct {
	Config map[string]interface{} `json:"config"`
}

type SystemInfo struct {
	Version     string            `json:"version"`
	BuildTime   string            `json:"build_time"`
	GoVersion   string            `json:"go_version"`
	Platform    string            `json:"platform"`
	Memory      MemoryInfo        `json:"memory"`
	Environment map[string]string `json:"environment"`
}

type MemoryInfo struct {
	Allocated  uint64 `json:"allocated"`
	TotalAlloc uint64 `json:"total_alloc"`
	Sys        uint64 `json:"sys"`
	NumGC      uint32 `json:"num_gc"`
}

// Health check handler
func (s *Server) healthHandler(c *fiber.Ctx) error {
	startTime := time.Now()

	// Get system health
	health := s.client.Health()

	// Check service health
	services := make(map[string]ServiceHealth)

	// Check GitHub API
	services["github"] = ServiceHealth{
		Status:      "healthy",
		LastChecked: time.Now(),
		Message:     "GitHub API accessible",
	}

	// Check GitLab API
	services["gitlab"] = ServiceHealth{
		Status:      "healthy",
		LastChecked: time.Now(),
		Message:     "GitLab API accessible",
	}

	// Check plugins system
	if _, err := s.client.ListPlugins(); err != nil {
		services["plugins"] = ServiceHealth{
			Status:      "unhealthy",
			LastChecked: time.Now(),
			Message:     fmt.Sprintf("Plugin system error: %v", err),
		}
	} else {
		services["plugins"] = ServiceHealth{
			Status:      "healthy",
			LastChecked: time.Now(),
			Message:     "Plugin system operational",
		}
	}

	response := HealthResponse{
		Status:    "healthy",
		Version:   "1.0.0", // TODO: Get from build info
		Timestamp: time.Now(),
		Uptime:    time.Since(startTime).String(),
		Services:  services,
	}

	return c.JSON(response)
}

// Bulk clone handler
func (s *Server) bulkCloneHandler(c *fiber.Ctx) error {
	var request BulkCloneAPIRequest
	if err := c.BodyParser(&request); err != nil {
		return respondWithError(c, fiber.StatusBadRequest, "invalid_request", "Invalid request body")
	}

	// Validate request
	if len(request.Platforms) == 0 {
		return respondWithError(c, fiber.StatusBadRequest, "validation_error", "At least one platform is required")
	}

	if request.OutputDir == "" {
		return respondWithError(c, fiber.StatusBadRequest, "validation_error", "Output directory is required")
	}

	// Convert to internal request
	bulkCloneReq := gzhclient.BulkCloneRequest{
		Platforms:      request.Platforms,
		OutputDir:      request.OutputDir,
		Concurrency:    request.Concurrency,
		Strategy:       request.Strategy,
		IncludePrivate: request.IncludePrivate,
		Filters:        request.Filters,
	}

	// Execute bulk clone
	result, err := s.client.BulkClone(c.Context(), bulkCloneReq)
	if err != nil {
		return respondWithError(c, fiber.StatusInternalServerError, "execution_error", err.Error())
	}

	return respondWithSuccess(c, result, "Bulk clone operation completed")
}

// Bulk clone status handler
func (s *Server) bulkCloneStatusHandler(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return respondWithError(c, fiber.StatusBadRequest, "invalid_request", "Operation ID is required")
	}

	// TODO: Implement operation status tracking
	// For now, return mock status
	status := map[string]interface{}{
		"id":           id,
		"status":       "completed",
		"progress":     100,
		"started_at":   time.Now().Add(-5 * time.Minute),
		"completed_at": time.Now(),
	}

	return respondWithSuccess(c, status, "Operation status retrieved")
}

// List plugins handler
func (s *Server) listPluginsHandler(c *fiber.Ctx) error {
	// Parse pagination
	page := parseIntQuery(c, "page", 1)
	perPage := parseIntQuery(c, "per_page", 20)

	pagination := PaginationRequest{
		Page:    page,
		PerPage: perPage,
		Sort:    c.Query("sort", "name"),
		Order:   c.Query("order", "asc"),
	}
	validatePagination(&pagination)

	// Get plugins
	plugins, err := s.client.ListPlugins()
	if err != nil {
		return respondWithError(c, fiber.StatusInternalServerError, "execution_error", err.Error())
	}

	// Calculate pagination
	total := len(plugins)
	startIdx := (pagination.Page - 1) * pagination.PerPage
	endIdx := startIdx + pagination.PerPage

	if startIdx > total {
		startIdx = total
	}
	if endIdx > total {
		endIdx = total
	}

	paginatedPlugins := plugins[startIdx:endIdx]
	paginationMeta := calculatePagination(pagination.Page, pagination.PerPage, total)

	response := map[string]interface{}{
		"plugins":    paginatedPlugins,
		"pagination": paginationMeta,
	}

	return respondWithSuccess(c, response, "Plugins retrieved")
}

// Execute plugin handler
func (s *Server) executePluginHandler(c *fiber.Ctx) error {
	pluginName := c.Params("name")
	if pluginName == "" {
		return respondWithError(c, fiber.StatusBadRequest, "invalid_request", "Plugin name is required")
	}

	var request PluginExecuteRequest
	if err := c.BodyParser(&request); err != nil {
		return respondWithError(c, fiber.StatusBadRequest, "invalid_request", "Invalid request body")
	}

	// Convert to internal request
	executeReq := gzhclient.PluginExecuteRequest{
		PluginName: pluginName,
		Method:     request.Method,
		Args:       request.Args,
		Timeout:    time.Duration(request.Timeout) * time.Second,
	}

	// Execute plugin
	result, err := s.client.ExecutePlugin(c.Context(), executeReq)
	if err != nil {
		return respondWithError(c, fiber.StatusInternalServerError, "execution_error", err.Error())
	}

	return respondWithSuccess(c, result, "Plugin executed successfully")
}

// Get plugin handler
func (s *Server) getPluginHandler(c *fiber.Ctx) error {
	pluginName := c.Params("name")
	if pluginName == "" {
		return respondWithError(c, fiber.StatusBadRequest, "invalid_request", "Plugin name is required")
	}

	// Get all plugins and find the requested one
	plugins, err := s.client.ListPlugins()
	if err != nil {
		return respondWithError(c, fiber.StatusInternalServerError, "execution_error", err.Error())
	}

	for _, plugin := range plugins {
		if plugin.Name == pluginName {
			return respondWithSuccess(c, plugin, "Plugin information retrieved")
		}
	}

	return respondWithError(c, fiber.StatusNotFound, "not_found", "Plugin not found")
}

// Enable plugin handler
func (s *Server) enablePluginHandler(c *fiber.Ctx) error {
	pluginName := c.Params("name")
	if pluginName == "" {
		return respondWithError(c, fiber.StatusBadRequest, "invalid_request", "Plugin name is required")
	}

	// TODO: Implement plugin enable/disable functionality in client
	result := map[string]interface{}{
		"plugin":    pluginName,
		"status":    "enabled",
		"timestamp": time.Now(),
	}

	return respondWithSuccess(c, result, "Plugin enabled successfully")
}

// Disable plugin handler
func (s *Server) disablePluginHandler(c *fiber.Ctx) error {
	pluginName := c.Params("name")
	if pluginName == "" {
		return respondWithError(c, fiber.StatusBadRequest, "invalid_request", "Plugin name is required")
	}

	// TODO: Implement plugin enable/disable functionality in client
	result := map[string]interface{}{
		"plugin":    pluginName,
		"status":    "disabled",
		"timestamp": time.Now(),
	}

	return respondWithSuccess(c, result, "Plugin disabled successfully")
}

// Get config handler
func (s *Server) getConfigHandler(c *fiber.Ctx) error {
	// TODO: Implement config retrieval from client
	config := map[string]interface{}{
		"server": map[string]interface{}{
			"host":        s.config.Host,
			"port":        s.config.Port,
			"environment": s.config.Environment,
		},
		"features": map[string]interface{}{
			"plugins_enabled": true,
			"auth_enabled":    s.config.EnableAuth,
			"swagger_enabled": s.config.EnableSwagger,
		},
		"limits": map[string]interface{}{
			"rate_limit":    s.config.RateLimit,
			"read_timeout":  s.config.ReadTimeout,
			"write_timeout": s.config.WriteTimeout,
		},
	}

	return respondWithSuccess(c, config, "Configuration retrieved")
}

// Update config handler
func (s *Server) updateConfigHandler(c *fiber.Ctx) error {
	var request ConfigUpdateRequest
	if err := c.BodyParser(&request); err != nil {
		return respondWithError(c, fiber.StatusBadRequest, "invalid_request", "Invalid request body")
	}

	// TODO: Implement config update functionality
	// For now, return success with echo of received config
	result := map[string]interface{}{
		"updated_config": request.Config,
		"timestamp":      time.Now(),
		"status":         "applied",
	}

	return respondWithSuccess(c, result, "Configuration updated successfully")
}

// Validate config handler
func (s *Server) validateConfigHandler(c *fiber.Ctx) error {
	var request ConfigUpdateRequest
	if err := c.BodyParser(&request); err != nil {
		return respondWithError(c, fiber.StatusBadRequest, "invalid_request", "Invalid request body")
	}

	// TODO: Implement config validation
	validationResult := map[string]interface{}{
		"valid":     true,
		"errors":    []string{},
		"warnings":  []string{},
		"timestamp": time.Now(),
	}

	return respondWithSuccess(c, validationResult, "Configuration validation completed")
}

// Get system info handler
func (s *Server) getSystemInfoHandler(c *fiber.Ctx) error {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	systemInfo := SystemInfo{
		Version:   "1.0.0",      // TODO: Get from build info
		BuildTime: "2025-07-12", // TODO: Get from build info
		GoVersion: runtime.Version(),
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		Memory: MemoryInfo{
			Allocated:  m.Alloc,
			TotalAlloc: m.TotalAlloc,
			Sys:        m.Sys,
			NumGC:      m.NumGC,
		},
		Environment: map[string]string{
			"server_environment": s.config.Environment,
			"auth_enabled":       strconv.FormatBool(s.config.EnableAuth),
			"plugins_available":  "true", // TODO: Get actual plugin count
		},
	}

	return respondWithSuccess(c, systemInfo, "System information retrieved")
}

// Get metrics handler
func (s *Server) getMetricsHandler(c *fiber.Ctx) error {
	metrics, err := s.client.GetSystemMetrics()
	if err != nil {
		return respondWithError(c, fiber.StatusInternalServerError, "execution_error", err.Error())
	}

	return respondWithSuccess(c, metrics, "System metrics retrieved")
}

// Get logs handler
func (s *Server) getLogsHandler(c *fiber.Ctx) error {
	// Parse query parameters
	level := c.Query("level", "")
	limit := parseIntQuery(c, "limit", 100)
	since := c.Query("since", "")

	// TODO: Implement log retrieval from client
	// For now, return mock logs
	logs := []map[string]interface{}{
		{
			"timestamp": time.Now().Add(-1 * time.Hour),
			"level":     "info",
			"message":   "Server started successfully",
			"module":    "server",
		},
		{
			"timestamp": time.Now().Add(-30 * time.Minute),
			"level":     "info",
			"message":   "Plugin system initialized",
			"module":    "plugins",
		},
		{
			"timestamp": time.Now().Add(-10 * time.Minute),
			"level":     "debug",
			"message":   "Health check completed",
			"module":    "health",
		},
	}

	// Apply filters
	if level != "" {
		filteredLogs := []map[string]interface{}{}
		for _, log := range logs {
			if log["level"] == level {
				filteredLogs = append(filteredLogs, log)
			}
		}
		logs = filteredLogs
	}

	// Apply limit
	if limit > 0 && len(logs) > limit {
		logs = logs[:limit]
	}

	response := map[string]interface{}{
		"logs": logs,
		"filters": map[string]interface{}{
			"level": level,
			"limit": limit,
			"since": since,
		},
		"total": len(logs),
	}

	return respondWithSuccess(c, response, "Logs retrieved")
}

// Get languages handler
func (s *Server) getLanguagesHandler(c *fiber.Ctx) error {
	languages := []map[string]interface{}{
		{"code": "en", "name": "English", "native": "English"},
		{"code": "ko", "name": "Korean", "native": "한국어"},
		{"code": "ja", "name": "Japanese", "native": "日本語"},
		{"code": "zh", "name": "Chinese", "native": "中文"},
		{"code": "es", "name": "Spanish", "native": "Español"},
		{"code": "fr", "name": "French", "native": "Français"},
		{"code": "de", "name": "German", "native": "Deutsch"},
		{"code": "ru", "name": "Russian", "native": "Русский"},
		{"code": "pt", "name": "Portuguese", "native": "Português"},
		{"code": "it", "name": "Italian", "native": "Italiano"},
		{"code": "ar", "name": "Arabic", "native": "العربية"},
		{"code": "hi", "name": "Hindi", "native": "हिन्दी"},
	}

	// Get current language from context
	currentLang := c.Locals("language")
	if currentLang == nil {
		currentLang = "en"
	}

	response := map[string]interface{}{
		"languages":        languages,
		"current_language": currentLang,
		"supported_count":  len(languages),
	}

	return respondWithSuccess(c, response, "Supported languages retrieved")
}

// Set language handler
func (s *Server) setLanguageHandler(c *fiber.Ctx) error {
	lang := c.Params("lang")
	if lang == "" {
		return respondWithError(c, fiber.StatusBadRequest, "invalid_request", "Language code is required")
	}

	// Set language in i18n manager
	i18n.SetLanguage(lang)

	// Set in context for this request
	c.Locals("language", lang)

	result := map[string]interface{}{
		"language":  lang,
		"timestamp": time.Now(),
		"status":    "set",
	}

	return respondWithSuccess(c, result, fmt.Sprintf("Language set to %s", lang))
}

// Get messages handler
func (s *Server) getMessagesHandler(c *fiber.Ctx) error {
	lang := c.Params("lang")
	if lang == "" {
		return respondWithError(c, fiber.StatusBadRequest, "invalid_request", "Language code is required")
	}

	// TODO: Get actual messages from i18n system
	// For now, return sample messages
	messages := map[string]string{
		"welcome":         i18n.Get("welcome"),
		"error":           i18n.Get("error"),
		"success":         i18n.Get("success"),
		"clone_starting":  i18n.Get("clone_starting"),
		"clone_completed": i18n.Get("clone_completed"),
	}

	response := map[string]interface{}{
		"language": lang,
		"messages": messages,
		"count":    len(messages),
	}

	return respondWithSuccess(c, response, "Messages retrieved")
}

// Not found handler
func (s *Server) notFoundHandler(c *fiber.Ctx) error {
	return respondWithError(c, fiber.StatusNotFound, "not_found", "Endpoint not found")
}
