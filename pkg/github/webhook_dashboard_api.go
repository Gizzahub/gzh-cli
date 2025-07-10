package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// WebhookDashboardAPI provides REST API for webhook monitoring dashboard
type WebhookDashboardAPI struct {
	monitor *WebhookMonitor
	logger  Logger
	config  *DashboardAPIConfig
}

// DashboardAPIConfig holds configuration for the dashboard API
type DashboardAPIConfig struct {
	Port            int           `json:"port" yaml:"port"`
	Host            string        `json:"host" yaml:"host"`
	EnableCORS      bool          `json:"enable_cors" yaml:"enable_cors"`
	RequestTimeout  time.Duration `json:"request_timeout" yaml:"request_timeout"`
	MaxRequestSize  int64         `json:"max_request_size" yaml:"max_request_size"`
	EnableAuth      bool          `json:"enable_auth" yaml:"enable_auth"`
	AuthToken       string        `json:"auth_token" yaml:"auth_token"`
	EnableRateLimit bool          `json:"enable_rate_limit" yaml:"enable_rate_limit"`
	RateLimit       int           `json:"rate_limit" yaml:"rate_limit"` // requests per minute
}

// DashboardResponse represents a standard API response
type DashboardResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	RequestID string      `json:"request_id,omitempty"`
}

// WebhookSummary provides a summary view of webhook status
type WebhookSummary struct {
	ID           string              `json:"id"`
	URL          string              `json:"url"`
	Organization string              `json:"organization"`
	Repository   string              `json:"repository,omitempty"`
	Status       WebhookHealthStatus `json:"status"`
	Active       bool                `json:"active"`
	Events       []string            `json:"events"`
	LastChecked  time.Time           `json:"last_checked"`
	ErrorRate    float64             `json:"error_rate"`
	Uptime       float64             `json:"uptime"`
	ActiveAlerts int                 `json:"active_alerts"`
}

// DashboardOverview provides high-level dashboard data
type DashboardOverview struct {
	Summary      DashboardSummary            `json:"summary"`
	StatusCounts map[WebhookHealthStatus]int `json:"status_counts"`
	RecentAlerts []WebhookAlert              `json:"recent_alerts"`
	TopIssues    []WebhookIssue              `json:"top_issues"`
	Trends       DashboardTrends             `json:"trends"`
}

// DashboardSummary provides summary statistics
type DashboardSummary struct {
	TotalWebhooks       int     `json:"total_webhooks"`
	HealthyWebhooks     int     `json:"healthy_webhooks"`
	UnhealthyWebhooks   int     `json:"unhealthy_webhooks"`
	ActiveAlerts        int     `json:"active_alerts"`
	OverallUptime       float64 `json:"overall_uptime"`
	AverageResponseTime string  `json:"average_response_time"`
}

// WebhookIssue represents common webhook issues
type WebhookIssue struct {
	Type        string    `json:"type"`
	Count       int       `json:"count"`
	Description string    `json:"description"`
	Severity    string    `json:"severity"`
	LastSeen    time.Time `json:"last_seen"`
}

// DashboardTrends provides trend data for the dashboard
type DashboardTrends struct {
	DeliverySuccessRate []TrendPoint `json:"delivery_success_rate"`
	AverageResponseTime []TrendPoint `json:"average_response_time"`
	AlertFrequency      []TrendPoint `json:"alert_frequency"`
}

// TrendPoint represents a data point in a trend chart
type TrendPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// OrganizationDashboard provides organization-specific dashboard data
type OrganizationDashboard struct {
	Organization string               `json:"organization"`
	Summary      DashboardSummary     `json:"summary"`
	Webhooks     []WebhookSummary     `json:"webhooks"`
	Alerts       []WebhookAlert       `json:"alerts"`
	Metrics      *OrganizationMetrics `json:"metrics"`
}

// NewWebhookDashboardAPI creates a new dashboard API
func NewWebhookDashboardAPI(monitor *WebhookMonitor, logger Logger, config *DashboardAPIConfig) *WebhookDashboardAPI {
	if config == nil {
		config = getDefaultDashboardAPIConfig()
	}

	return &WebhookDashboardAPI{
		monitor: monitor,
		logger:  logger,
		config:  config,
	}
}

// StartServer starts the dashboard API server
func (api *WebhookDashboardAPI) StartServer(ctx context.Context) error {
	router := mux.NewRouter()

	// Add middleware
	router.Use(api.corsMiddleware)
	router.Use(api.loggingMiddleware)
	if api.config.EnableAuth {
		router.Use(api.authMiddleware)
	}

	// API routes
	apiRouter := router.PathPrefix("/api/v1").Subrouter()
	api.setupRoutes(apiRouter)

	// Static file serving for the dashboard frontend
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/dashboard/")))

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", api.config.Host, api.config.Port),
		Handler:      router,
		ReadTimeout:  api.config.RequestTimeout,
		WriteTimeout: api.config.RequestTimeout,
	}

	api.logger.Info("Starting webhook dashboard API server",
		"host", api.config.Host,
		"port", api.config.Port)

	go func() {
		<-ctx.Done()
		api.logger.Info("Shutting down dashboard API server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		server.Shutdown(shutdownCtx)
	}()

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("dashboard API server failed: %w", err)
	}

	return nil
}

func (api *WebhookDashboardAPI) setupRoutes(router *mux.Router) {
	// Dashboard overview
	router.HandleFunc("/dashboard", api.handleDashboardOverview).Methods("GET")
	router.HandleFunc("/dashboard/organization/{org}", api.handleOrganizationDashboard).Methods("GET")

	// Webhook management
	router.HandleFunc("/webhooks", api.handleListWebhooks).Methods("GET")
	router.HandleFunc("/webhooks/{id}", api.handleGetWebhook).Methods("GET")
	router.HandleFunc("/webhooks/{id}/status", api.handleGetWebhookStatus).Methods("GET")
	router.HandleFunc("/webhooks/{id}/history", api.handleGetWebhookHistory).Methods("GET")

	// Metrics and analytics
	router.HandleFunc("/metrics", api.handleGetMetrics).Methods("GET")
	router.HandleFunc("/metrics/organization/{org}", api.handleGetOrganizationMetrics).Methods("GET")
	router.HandleFunc("/metrics/trends", api.handleGetTrends).Methods("GET")

	// Alerts management
	router.HandleFunc("/alerts", api.handleListAlerts).Methods("GET")
	router.HandleFunc("/alerts/{id}/acknowledge", api.handleAcknowledgeAlert).Methods("POST")
	router.HandleFunc("/alerts/active", api.handleGetActiveAlerts).Methods("GET")

	// Health and status
	router.HandleFunc("/health", api.handleHealthCheck).Methods("GET")
	router.HandleFunc("/status", api.handleSystemStatus).Methods("GET")
}

// API Handlers

func (api *WebhookDashboardAPI) handleDashboardOverview(w http.ResponseWriter, r *http.Request) {
	metrics := api.monitor.GetMetrics()
	alerts := api.monitor.GetActiveAlerts()

	// Get recent alerts (last 24 hours)
	now := time.Now()
	recentAlerts := make([]WebhookAlert, 0)
	for _, alert := range alerts {
		if now.Sub(alert.CreatedAt) <= 24*time.Hour {
			recentAlerts = append(recentAlerts, alert)
		}
	}

	// Calculate summary
	totalWebhooks := int(metrics.TotalWebhooks)
	healthyWebhooks := int(metrics.HealthyWebhooks)
	unhealthyWebhooks := int(metrics.UnhealthyWebhooks)

	// Calculate overall uptime
	overallUptime := 100.0
	if totalWebhooks > 0 {
		overallUptime = float64(healthyWebhooks) / float64(totalWebhooks) * 100
	}

	overview := DashboardOverview{
		Summary: DashboardSummary{
			TotalWebhooks:       totalWebhooks,
			HealthyWebhooks:     healthyWebhooks,
			UnhealthyWebhooks:   unhealthyWebhooks,
			ActiveAlerts:        len(alerts),
			OverallUptime:       overallUptime,
			AverageResponseTime: metrics.AverageResponseTime.String(),
		},
		StatusCounts: make(map[WebhookHealthStatus]int),
		RecentAlerts: recentAlerts,
		TopIssues:    api.generateTopIssues(alerts),
		Trends:       api.generateTrends(),
	}

	// Convert status distribution
	for status, count := range metrics.StatusDistribution {
		overview.StatusCounts[status] = int(count)
	}

	api.sendJSONResponse(w, overview)
}

func (api *WebhookDashboardAPI) handleOrganizationDashboard(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	org := vars["org"]

	if org == "" {
		api.sendErrorResponse(w, "Organization parameter is required", http.StatusBadRequest)
		return
	}

	allWebhooks := api.monitor.GetAllWebhookStatuses()
	orgWebhooks := make([]WebhookSummary, 0)
	orgAlerts := make([]WebhookAlert, 0)

	healthyCount := 0
	totalCount := 0

	for _, webhook := range allWebhooks {
		if webhook.Organization == org {
			totalCount++
			if webhook.Status == WebhookStatusHealthy {
				healthyCount++
			}

			summary := WebhookSummary{
				ID:           webhook.ID,
				URL:          webhook.URL,
				Organization: webhook.Organization,
				Repository:   webhook.Repository,
				Status:       webhook.Status,
				Active:       webhook.Active,
				Events:       webhook.Events,
				LastChecked:  webhook.LastChecked,
				ErrorRate:    webhook.Metrics.ErrorRate,
				Uptime:       webhook.Metrics.Uptime,
				ActiveAlerts: len(webhook.Alerts),
			}
			orgWebhooks = append(orgWebhooks, summary)

			// Collect organization alerts
			for _, alert := range webhook.Alerts {
				if alert.ResolvedAt == nil {
					orgAlerts = append(orgAlerts, alert)
				}
			}
		}
	}

	// Calculate organization uptime
	orgUptime := 100.0
	if totalCount > 0 {
		orgUptime = float64(healthyCount) / float64(totalCount) * 100
	}

	orgMetrics := api.monitor.GetMetrics().OrganizationMetrics[org]
	if orgMetrics == nil {
		orgMetrics = &OrganizationMetrics{}
	}

	orgDashboard := OrganizationDashboard{
		Organization: org,
		Summary: DashboardSummary{
			TotalWebhooks:       totalCount,
			HealthyWebhooks:     healthyCount,
			UnhealthyWebhooks:   totalCount - healthyCount,
			ActiveAlerts:        len(orgAlerts),
			OverallUptime:       orgUptime,
			AverageResponseTime: orgMetrics.AverageResponseTime.String(),
		},
		Webhooks: orgWebhooks,
		Alerts:   orgAlerts,
		Metrics:  orgMetrics,
	}

	api.sendJSONResponse(w, orgDashboard)
}

func (api *WebhookDashboardAPI) handleListWebhooks(w http.ResponseWriter, r *http.Request) {
	org := r.URL.Query().Get("organization")
	status := r.URL.Query().Get("status")

	allWebhooks := api.monitor.GetAllWebhookStatuses()
	webhooks := make([]WebhookSummary, 0)

	for _, webhook := range allWebhooks {
		// Filter by organization if specified
		if org != "" && webhook.Organization != org {
			continue
		}

		// Filter by status if specified
		if status != "" && string(webhook.Status) != status {
			continue
		}

		summary := WebhookSummary{
			ID:           webhook.ID,
			URL:          webhook.URL,
			Organization: webhook.Organization,
			Repository:   webhook.Repository,
			Status:       webhook.Status,
			Active:       webhook.Active,
			Events:       webhook.Events,
			LastChecked:  webhook.LastChecked,
			ErrorRate:    webhook.Metrics.ErrorRate,
			Uptime:       webhook.Metrics.Uptime,
			ActiveAlerts: len(webhook.Alerts),
		}
		webhooks = append(webhooks, summary)
	}

	api.sendJSONResponse(w, webhooks)
}

func (api *WebhookDashboardAPI) handleGetWebhook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	webhook, err := api.monitor.GetWebhookStatus(id)
	if err != nil {
		api.sendErrorResponse(w, err.Error(), http.StatusNotFound)
		return
	}

	api.sendJSONResponse(w, webhook)
}

func (api *WebhookDashboardAPI) handleGetWebhookStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	webhook, err := api.monitor.GetWebhookStatus(id)
	if err != nil {
		api.sendErrorResponse(w, err.Error(), http.StatusNotFound)
		return
	}

	statusInfo := map[string]interface{}{
		"id":            webhook.ID,
		"status":        webhook.Status,
		"last_checked":  webhook.LastChecked,
		"metrics":       webhook.Metrics,
		"active_alerts": len(webhook.Alerts),
	}

	api.sendJSONResponse(w, statusInfo)
}

func (api *WebhookDashboardAPI) handleGetWebhookHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	webhook, err := api.monitor.GetWebhookStatus(id)
	if err != nil {
		api.sendErrorResponse(w, err.Error(), http.StatusNotFound)
		return
	}

	// Parse query parameters for pagination
	limitStr := r.URL.Query().Get("limit")
	limit := 50 // default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	history := webhook.History
	if len(history) > limit {
		history = history[:limit]
	}

	api.sendJSONResponse(w, history)
}

func (api *WebhookDashboardAPI) handleGetMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := api.monitor.GetMetrics()
	api.sendJSONResponse(w, metrics)
}

func (api *WebhookDashboardAPI) handleGetOrganizationMetrics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	org := vars["org"]

	metrics := api.monitor.GetMetrics()
	orgMetrics, exists := metrics.OrganizationMetrics[org]
	if !exists {
		api.sendErrorResponse(w, "Organization not found", http.StatusNotFound)
		return
	}

	api.sendJSONResponse(w, orgMetrics)
}

func (api *WebhookDashboardAPI) handleGetTrends(w http.ResponseWriter, r *http.Request) {
	// This would typically query a time-series database
	// For now, return mock trend data
	trends := api.generateTrends()
	api.sendJSONResponse(w, trends)
}

func (api *WebhookDashboardAPI) handleListAlerts(w http.ResponseWriter, r *http.Request) {
	severity := r.URL.Query().Get("severity")
	alertType := r.URL.Query().Get("type")

	alerts := api.monitor.GetActiveAlerts()
	filteredAlerts := make([]WebhookAlert, 0)

	for _, alert := range alerts {
		// Filter by severity if specified
		if severity != "" && string(alert.Severity) != severity {
			continue
		}

		// Filter by type if specified
		if alertType != "" && string(alert.Type) != alertType {
			continue
		}

		filteredAlerts = append(filteredAlerts, alert)
	}

	api.sendJSONResponse(w, filteredAlerts)
}

func (api *WebhookDashboardAPI) handleAcknowledgeAlert(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	alertID := vars["id"]

	err := api.monitor.AcknowledgeAlert(alertID)
	if err != nil {
		api.sendErrorResponse(w, err.Error(), http.StatusNotFound)
		return
	}

	api.sendJSONResponse(w, map[string]string{"status": "acknowledged"})
}

func (api *WebhookDashboardAPI) handleGetActiveAlerts(w http.ResponseWriter, r *http.Request) {
	alerts := api.monitor.GetActiveAlerts()
	api.sendJSONResponse(w, alerts)
}

func (api *WebhookDashboardAPI) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"version":   "1.0.0",
		"uptime":    time.Since(time.Now()).String(), // This would be actual uptime
	}

	api.sendJSONResponse(w, health)
}

func (api *WebhookDashboardAPI) handleSystemStatus(w http.ResponseWriter, r *http.Request) {
	metrics := api.monitor.GetMetrics()
	alerts := api.monitor.GetActiveAlerts()

	status := map[string]interface{}{
		"webhook_monitor": "running",
		"total_webhooks":  metrics.TotalWebhooks,
		"active_alerts":   len(alerts),
		"last_updated":    metrics.LastUpdated,
	}

	api.sendJSONResponse(w, status)
}

// Helper methods

func (api *WebhookDashboardAPI) sendJSONResponse(w http.ResponseWriter, data interface{}) {
	response := DashboardResponse{
		Success:   true,
		Data:      data,
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (api *WebhookDashboardAPI) sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	response := DashboardResponse{
		Success:   false,
		Error:     message,
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

func (api *WebhookDashboardAPI) generateTopIssues(alerts []WebhookAlert) []WebhookIssue {
	issueMap := make(map[WebhookAlertType]WebhookIssue)

	for _, alert := range alerts {
		if issue, exists := issueMap[alert.Type]; exists {
			issue.Count++
			if alert.CreatedAt.After(issue.LastSeen) {
				issue.LastSeen = alert.CreatedAt
			}
			issueMap[alert.Type] = issue
		} else {
			issueMap[alert.Type] = WebhookIssue{
				Type:        string(alert.Type),
				Count:       1,
				Description: api.getIssueDescription(alert.Type),
				Severity:    string(alert.Severity),
				LastSeen:    alert.CreatedAt,
			}
		}
	}

	issues := make([]WebhookIssue, 0, len(issueMap))
	for _, issue := range issueMap {
		issues = append(issues, issue)
	}

	return issues
}

func (api *WebhookDashboardAPI) generateTrends() DashboardTrends {
	// Mock trend data - in production this would come from time-series data
	now := time.Now()
	trends := DashboardTrends{
		DeliverySuccessRate: make([]TrendPoint, 24),
		AverageResponseTime: make([]TrendPoint, 24),
		AlertFrequency:      make([]TrendPoint, 24),
	}

	for i := 0; i < 24; i++ {
		timestamp := now.Add(time.Duration(-i) * time.Hour)
		trends.DeliverySuccessRate[i] = TrendPoint{
			Timestamp: timestamp,
			Value:     95.0 + float64(i%10), // Mock data
		}
		trends.AverageResponseTime[i] = TrendPoint{
			Timestamp: timestamp,
			Value:     200 + float64(i*10), // Mock data in ms
		}
		trends.AlertFrequency[i] = TrendPoint{
			Timestamp: timestamp,
			Value:     float64(i % 5), // Mock data
		}
	}

	return trends
}

func (api *WebhookDashboardAPI) getIssueDescription(alertType WebhookAlertType) string {
	descriptions := map[WebhookAlertType]string{
		AlertTypeHighErrorRate:       "Webhook error rate is above threshold",
		AlertTypeSlowResponse:        "Webhook response time is slower than expected",
		AlertTypeConsecutiveFailures: "Multiple consecutive webhook delivery failures",
		AlertTypeConfigurationIssue:  "Webhook configuration issues detected",
		AlertTypeDeliveryFailure:     "Webhook delivery failures",
		AlertTypeEndpointDown:        "Webhook endpoint is not responding",
	}

	if desc, exists := descriptions[alertType]; exists {
		return desc
	}
	return "Unknown issue type"
}

// Middleware

func (api *WebhookDashboardAPI) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if api.config.EnableCORS {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (api *WebhookDashboardAPI) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		api.logger.Debug("HTTP request",
			"method", r.Method,
			"path", r.URL.Path,
			"duration", time.Since(start))
	})
}

func (api *WebhookDashboardAPI) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if api.config.AuthToken == "" {
			next.ServeHTTP(w, r)
			return
		}

		token := r.Header.Get("Authorization")
		if token != "Bearer "+api.config.AuthToken {
			api.sendErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func getDefaultDashboardAPIConfig() *DashboardAPIConfig {
	return &DashboardAPIConfig{
		Port:            8080,
		Host:            "0.0.0.0",
		EnableCORS:      true,
		RequestTimeout:  30 * time.Second,
		MaxRequestSize:  1024 * 1024, // 1MB
		EnableAuth:      false,
		EnableRateLimit: false,
		RateLimit:       100, // requests per minute
	}
}
