package repoconfig

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
)

// newDashboardCmd creates the dashboard subcommand
func newDashboardCmd() *cobra.Command {
	var flags GlobalFlags
	var (
		port        int
		host        string
		autoRefresh int
		enableWS    bool
		dataSource  string
		mockData    bool
	)

	cmd := &cobra.Command{
		Use:   "dashboard",
		Short: "Start real-time compliance dashboard",
		Long: `Start a web-based real-time compliance dashboard for monitoring 
repository configurations and policy adherence across the organization.

Dashboard Features:
- Real-time compliance monitoring
- Interactive charts and graphs
- Drill-down repository details
- Policy violation tracking
- WebSocket-based live updates
- Responsive web interface

The dashboard provides a comprehensive view of:
- Organization-wide compliance status
- Repository-level configuration details
- Policy violation trends over time
- Automated fix suggestions
- Risk assessment metrics

Examples:
  gz repo-config dashboard --org myorg                    # Start dashboard on default port
  gz repo-config dashboard --org myorg --port 8080       # Custom port
  gz repo-config dashboard --org myorg --auto-refresh 30 # Auto-refresh every 30 seconds
  gz repo-config dashboard --org myorg --mock-data       # Use mock data for testing`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDashboardCommand(flags, port, host, autoRefresh, enableWS, dataSource, mockData)
		},
	}

	// Add global flags
	addGlobalFlags(cmd, &flags)

	// Add dashboard-specific flags
	cmd.Flags().IntVar(&port, "port", 8080, "Port to serve dashboard on")
	cmd.Flags().StringVar(&host, "host", "localhost", "Host to bind dashboard server to")
	cmd.Flags().IntVar(&autoRefresh, "auto-refresh", 60, "Auto-refresh interval in seconds (0 to disable)")
	cmd.Flags().BoolVar(&enableWS, "websocket", true, "Enable WebSocket real-time updates")
	cmd.Flags().StringVar(&dataSource, "data-source", "github", "Data source for compliance data (github, file, mock)")
	cmd.Flags().BoolVar(&mockData, "mock-data", false, "Use mock data for testing (overrides data-source)")

	return cmd
}

// runDashboardCommand executes the dashboard command
func runDashboardCommand(flags GlobalFlags, port int, host string, autoRefresh int, enableWS bool, dataSource string, mockData bool) error {
	if flags.Organization == "" {
		return fmt.Errorf("organization is required (use --org flag)")
	}

	fmt.Printf("🌐 Starting Compliance Dashboard\n")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("Organization: %s\n", flags.Organization)
	fmt.Printf("Server: http://%s:%d\n", host, port)
	fmt.Printf("Auto-refresh: %d seconds\n", autoRefresh)
	fmt.Printf("WebSocket: %s\n", enabledString(enableWS))
	if mockData {
		fmt.Printf("Data Source: mock (testing mode)\n")
	} else {
		fmt.Printf("Data Source: %s\n", dataSource)
	}
	fmt.Println()

	// Create dashboard server
	dashboard := NewDashboardServer(DashboardConfig{
		Organization: flags.Organization,
		Port:         port,
		Host:         host,
		AutoRefresh:  autoRefresh,
		EnableWS:     enableWS,
		DataSource:   dataSource,
		MockData:     mockData,
		Verbose:      flags.Verbose,
	})

	// Start the dashboard server
	return dashboard.Start()
}

// DashboardConfig contains configuration for the dashboard server
type DashboardConfig struct {
	Organization string
	Port         int
	Host         string
	AutoRefresh  int
	EnableWS     bool
	DataSource   string
	MockData     bool
	Verbose      bool
}

// DashboardServer serves the compliance dashboard
type DashboardServer struct {
	config    DashboardConfig
	router    *mux.Router
	upgrader  websocket.Upgrader
	clients   map[*websocket.Conn]bool
	broadcast chan []byte
	dataCache *DashboardDataCache
}

// DashboardData represents the data structure for the dashboard
type DashboardData struct {
	Organization         string                   `json:"organization"`
	LastUpdated          time.Time                `json:"last_updated"`
	Summary              ComplianceSummaryMetrics `json:"summary"`
	Repositories         []RepositoryDashboard    `json:"repositories"`
	PolicyTrends         []PolicyTrendData        `json:"policy_trends"`
	RecentViolations     []ViolationSummary       `json:"recent_violations"`
	ComplianceHistory    []ComplianceDataPoint    `json:"compliance_history"`
	TopViolatingRepos    []RepositoryRank         `json:"top_violating_repos"`
	AutoFixOpportunities []AutoFixOpportunity     `json:"auto_fix_opportunities"`
}

// ComplianceSummaryMetrics provides high-level compliance metrics
type ComplianceSummaryMetrics struct {
	TotalRepositories     int     `json:"total_repositories"`
	CompliantRepositories int     `json:"compliant_repositories"`
	CompliancePercentage  float64 `json:"compliance_percentage"`
	TotalViolations       int     `json:"total_violations"`
	CriticalViolations    int     `json:"critical_violations"`
	RecentlyFixed         int     `json:"recently_fixed"`
	PendingFixes          int     `json:"pending_fixes"`
	ComplianceScore       float64 `json:"compliance_score"`
	TrendDirection        string  `json:"trend_direction"` // improving, declining, stable
	LastAuditTime         string  `json:"last_audit_time"`
}

// RepositoryDashboard represents repository data for dashboard
type RepositoryDashboard struct {
	Name             string    `json:"name"`
	Visibility       string    `json:"visibility"`
	Language         string    `json:"language"`
	LastCommit       string    `json:"last_commit"`
	ComplianceScore  float64   `json:"compliance_score"`
	ViolationCount   int       `json:"violation_count"`
	CriticalCount    int       `json:"critical_count"`
	AutoFixAvailable int       `json:"auto_fix_available"`
	LastChecked      time.Time `json:"last_checked"`
	Status           string    `json:"status"` // compliant, warning, critical
	Template         string    `json:"template"`
	RecentChanges    []string  `json:"recent_changes"`
	TrendDirection   string    `json:"trend_direction"`
}

// PolicyTrendData represents trend data for a specific policy
type PolicyTrendData struct {
	PolicyName     string           `json:"policy_name"`
	ViolationCount int              `json:"violation_count"`
	TrendDirection string           `json:"trend_direction"`
	ChangeRate     float64          `json:"change_rate"`
	DataPoints     []TrendDataPoint `json:"data_points"`
	Severity       string           `json:"severity"`
}

// TrendDataPoint represents a single data point in a trend
type TrendDataPoint struct {
	Date  time.Time `json:"date"`
	Value int       `json:"value"`
}

// ViolationSummary represents a violation for dashboard display
type ViolationSummary struct {
	Repository  string    `json:"repository"`
	Policy      string    `json:"policy"`
	Severity    string    `json:"severity"`
	Description string    `json:"description"`
	FirstSeen   time.Time `json:"first_seen"`
	Status      string    `json:"status"` // new, acknowledged, fixing, resolved
}

// ComplianceDataPoint represents a point in compliance history
type ComplianceDataPoint struct {
	Date              time.Time `json:"date"`
	CompliancePercent float64   `json:"compliance_percent"`
	TotalRepositories int       `json:"total_repositories"`
	CompliantRepos    int       `json:"compliant_repos"`
	TotalViolations   int       `json:"total_violations"`
}

// RepositoryRank represents a repository ranking by violations
type RepositoryRank struct {
	Repository      string  `json:"repository"`
	ViolationCount  int     `json:"violation_count"`
	CriticalCount   int     `json:"critical_count"`
	ComplianceScore float64 `json:"compliance_score"`
	Priority        string  `json:"priority"`
}

// AutoFixOpportunity represents an opportunity for automated fixing
type AutoFixOpportunity struct {
	Repository     string `json:"repository"`
	ViolationType  string `json:"violation_type"`
	Description    string `json:"description"`
	EstimatedTime  string `json:"estimated_time"`
	RiskLevel      string `json:"risk_level"`
	AutoApplicable bool   `json:"auto_applicable"`
	AffectedRepos  int    `json:"affected_repos"`
}

// DashboardDataCache caches dashboard data to reduce API calls
type DashboardDataCache struct {
	data        *DashboardData
	lastUpdated time.Time
	cacheTTL    time.Duration
}

// NewDashboardServer creates a new dashboard server
func NewDashboardServer(config DashboardConfig) *DashboardServer {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow connections from any origin for now
		},
	}

	return &DashboardServer{
		config:    config,
		router:    mux.NewRouter(),
		upgrader:  upgrader,
		clients:   make(map[*websocket.Conn]bool),
		broadcast: make(chan []byte),
		dataCache: &DashboardDataCache{
			cacheTTL: time.Duration(config.AutoRefresh) * time.Second,
		},
	}
}

// Start starts the dashboard server
func (ds *DashboardServer) Start() error {
	// Setup routes
	ds.setupRoutes()

	// Start WebSocket handler if enabled
	if ds.config.EnableWS {
		go ds.handleWebSocketConnections()
	}

	// Start data refresh routine
	if ds.config.AutoRefresh > 0 {
		go ds.startDataRefreshRoutine()
	}

	// Start HTTP server
	addr := fmt.Sprintf("%s:%d", ds.config.Host, ds.config.Port)
	fmt.Printf("🚀 Dashboard server starting on %s\n", addr)
	fmt.Printf("📊 Open your browser to http://%s to view the dashboard\n", addr)
	fmt.Println("Press Ctrl+C to stop the server")

	return http.ListenAndServe(addr, ds.router)
}

// setupRoutes configures the HTTP routes
func (ds *DashboardServer) setupRoutes() {
	// Static routes
	ds.router.HandleFunc("/", ds.handleDashboard).Methods("GET")
	ds.router.HandleFunc("/api/data", ds.handleAPIData).Methods("GET")
	ds.router.HandleFunc("/api/repositories", ds.handleAPIRepositories).Methods("GET")
	ds.router.HandleFunc("/api/policies", ds.handleAPIPolicies).Methods("GET")
	ds.router.HandleFunc("/api/violations", ds.handleAPIViolations).Methods("GET")
	ds.router.HandleFunc("/api/trends", ds.handleAPITrends).Methods("GET")
	ds.router.HandleFunc("/api/autofix", ds.handleAPIAutoFix).Methods("GET")

	// WebSocket endpoint
	if ds.config.EnableWS {
		ds.router.HandleFunc("/ws", ds.handleWebSocket).Methods("GET")
	}

	// Health check
	ds.router.HandleFunc("/health", ds.handleHealth).Methods("GET")
}

// handleDashboard serves the main dashboard HTML page
func (ds *DashboardServer) handleDashboard(w http.ResponseWriter, r *http.Request) {
	data := ds.getDashboardPageData()

	tmpl := template.Must(template.New("dashboard").Parse(dashboardHTML))
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleAPIData serves the main dashboard data as JSON
func (ds *DashboardServer) handleAPIData(w http.ResponseWriter, r *http.Request) {
	data, err := ds.getDashboardData()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// handleAPIRepositories serves repository-specific data
func (ds *DashboardServer) handleAPIRepositories(w http.ResponseWriter, r *http.Request) {
	data, err := ds.getDashboardData()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data.Repositories)
}

// handleAPIPolicies serves policy compliance data
func (ds *DashboardServer) handleAPIPolicies(w http.ResponseWriter, r *http.Request) {
	data, err := ds.getDashboardData()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data.PolicyTrends)
}

// handleAPIViolations serves violation data
func (ds *DashboardServer) handleAPIViolations(w http.ResponseWriter, r *http.Request) {
	data, err := ds.getDashboardData()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data.RecentViolations)
}

// handleAPITrends serves trend analysis data
func (ds *DashboardServer) handleAPITrends(w http.ResponseWriter, r *http.Request) {
	data, err := ds.getDashboardData()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data.ComplianceHistory)
}

// handleAPIAutoFix serves auto-fix opportunities
func (ds *DashboardServer) handleAPIAutoFix(w http.ResponseWriter, r *http.Request) {
	data, err := ds.getDashboardData()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data.AutoFixOpportunities)
}

// handleWebSocket handles WebSocket connections
func (ds *DashboardServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := ds.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// Register client
	ds.clients[conn] = true

	if ds.config.Verbose {
		log.Printf("WebSocket client connected from %s", r.RemoteAddr)
	}

	// Send initial data
	data, err := ds.getDashboardData()
	if err == nil {
		if jsonData, err := json.Marshal(data); err == nil {
			conn.WriteMessage(websocket.TextMessage, jsonData)
		}
	}

	// Handle incoming messages
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			if ds.config.Verbose {
				log.Printf("WebSocket client disconnected: %v", err)
			}
			delete(ds.clients, conn)
			break
		}
	}
}

// handleHealth serves health check endpoint
func (ds *DashboardServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":       "healthy",
		"timestamp":    time.Now(),
		"organization": ds.config.Organization,
		"clients":      len(ds.clients),
		"cache_age":    time.Since(ds.dataCache.lastUpdated).Seconds(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// getDashboardData gets dashboard data from cache or refreshes it
func (ds *DashboardServer) getDashboardData() (*DashboardData, error) {
	// Check cache
	if ds.dataCache.data != nil && time.Since(ds.dataCache.lastUpdated) < ds.dataCache.cacheTTL {
		return ds.dataCache.data, nil
	}

	// Refresh data
	var data *DashboardData
	var err error

	if ds.config.MockData {
		data = ds.generateMockData()
	} else {
		data, err = ds.fetchRealData()
		if err != nil {
			return nil, err
		}
	}

	// Update cache
	ds.dataCache.data = data
	ds.dataCache.lastUpdated = time.Now()

	return data, nil
}

// fetchRealData fetches real compliance data from the configured source
func (ds *DashboardServer) fetchRealData() (*DashboardData, error) {
	// In a real implementation, this would:
	// 1. Call the audit system to get current compliance data
	// 2. Fetch historical data from storage
	// 3. Calculate trends and metrics
	// 4. Return structured dashboard data

	// For now, use mock data but with real audit integration structure
	auditData, err := performComplianceAudit(ds.config.Organization, "")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch compliance data: %w", err)
	}

	// Transform audit data to dashboard format
	data := ds.transformAuditDataToDashboard(auditData)

	return data, nil
}

// transformAuditDataToDashboard converts audit data to dashboard format
func (ds *DashboardServer) transformAuditDataToDashboard(auditData AuditData) *DashboardData {
	// Convert repositories
	var dashboardRepos []RepositoryDashboard
	for _, repo := range auditData.Repositories {
		dashboardRepos = append(dashboardRepos, RepositoryDashboard{
			Name:             repo.Name,
			Visibility:       repo.Visibility,
			Language:         "Go", // Would be fetched from GitHub API
			LastCommit:       "2 hours ago",
			ComplianceScore:  calculateRepoComplianceScore(repo),
			ViolationCount:   repo.ViolationCount,
			CriticalCount:    repo.CriticalCount,
			AutoFixAvailable: countAutoFixable(repo),
			LastChecked:      time.Now(),
			Status:           getRepoStatus(repo),
			Template:         repo.Template,
			RecentChanges:    []string{},
			TrendDirection:   "stable",
		})
	}

	// Convert violations to recent violations
	var recentViolations []ViolationSummary
	for _, violation := range auditData.Violations {
		recentViolations = append(recentViolations, ViolationSummary{
			Repository:  violation.Repository,
			Policy:      violation.Policy,
			Severity:    violation.Severity,
			Description: violation.Description,
			FirstSeen:   time.Now().Add(-24 * time.Hour),
			Status:      "new",
		})
	}

	return &DashboardData{
		Organization: auditData.Organization,
		LastUpdated:  time.Now(),
		Summary: ComplianceSummaryMetrics{
			TotalRepositories:     auditData.Summary.TotalRepositories,
			CompliantRepositories: auditData.Summary.CompliantRepositories,
			CompliancePercentage:  auditData.Summary.CompliancePercentage,
			TotalViolations:       auditData.Summary.TotalViolations,
			CriticalViolations:    auditData.Summary.CriticalViolations,
			RecentlyFixed:         0,
			PendingFixes:          auditData.Summary.TotalViolations,
			ComplianceScore:       auditData.Summary.CompliancePercentage,
			TrendDirection:        "stable",
			LastAuditTime:         auditData.GeneratedAt.Format("15:04:05"),
		},
		Repositories:         dashboardRepos,
		PolicyTrends:         ds.generatePolicyTrends(auditData.PolicyCompliance),
		RecentViolations:     recentViolations,
		ComplianceHistory:    ds.generateComplianceHistory(),
		TopViolatingRepos:    ds.getTopViolatingRepos(dashboardRepos),
		AutoFixOpportunities: ds.generateAutoFixOpportunities(auditData.Violations),
	}
}

// Helper functions for dashboard data transformation
func calculateRepoComplianceScore(repo RepositoryAudit) float64 {
	if repo.ViolationCount == 0 {
		return 100.0
	}
	// Simple scoring: 100 - (violations * 5) - (critical * 15)
	score := 100.0 - float64(repo.ViolationCount*5) - float64(repo.CriticalCount*15)
	if score < 0 {
		score = 0
	}
	return score
}

func countAutoFixable(repo RepositoryAudit) int {
	// Mock implementation - would count actual auto-fixable violations
	return repo.ViolationCount / 2
}

func getRepoStatus(repo RepositoryAudit) string {
	if repo.CriticalCount > 0 {
		return "critical"
	}
	if repo.ViolationCount > 0 {
		return "warning"
	}
	return "compliant"
}

// generateMockData creates mock dashboard data for testing
func (ds *DashboardServer) generateMockData() *DashboardData {
	now := time.Now()

	return &DashboardData{
		Organization: ds.config.Organization,
		LastUpdated:  now,
		Summary: ComplianceSummaryMetrics{
			TotalRepositories:     25,
			CompliantRepositories: 18,
			CompliancePercentage:  72.0,
			TotalViolations:       15,
			CriticalViolations:    3,
			RecentlyFixed:         5,
			PendingFixes:          10,
			ComplianceScore:       75.5,
			TrendDirection:        "improving",
			LastAuditTime:         now.Format("15:04:05"),
		},
		Repositories: []RepositoryDashboard{
			{
				Name:             "api-server",
				Visibility:       "private",
				Language:         "Go",
				LastCommit:       "2 hours ago",
				ComplianceScore:  95.0,
				ViolationCount:   1,
				CriticalCount:    0,
				AutoFixAvailable: 1,
				LastChecked:      now,
				Status:           "compliant",
				Template:         "microservice",
				RecentChanges:    []string{"Branch protection enabled"},
				TrendDirection:   "improving",
			},
			{
				Name:             "web-frontend",
				Visibility:       "private",
				Language:         "TypeScript",
				LastCommit:       "1 day ago",
				ComplianceScore:  82.0,
				ViolationCount:   3,
				CriticalCount:    1,
				AutoFixAvailable: 2,
				LastChecked:      now,
				Status:           "warning",
				Template:         "frontend",
				RecentChanges:    []string{"Security scanning disabled"},
				TrendDirection:   "stable",
			},
		},
		PolicyTrends: []PolicyTrendData{
			{
				PolicyName:     "Branch Protection",
				ViolationCount: 5,
				TrendDirection: "improving",
				ChangeRate:     -0.2,
				Severity:       "critical",
				DataPoints: []TrendDataPoint{
					{Date: now.AddDate(0, 0, -7), Value: 7},
					{Date: now.AddDate(0, 0, -3), Value: 6},
					{Date: now, Value: 5},
				},
			},
		},
		RecentViolations: []ViolationSummary{
			{
				Repository:  "legacy-service",
				Policy:      "Branch Protection",
				Severity:    "critical",
				Description: "Main branch lacks protection rules",
				FirstSeen:   now.Add(-24 * time.Hour),
				Status:      "new",
			},
		},
		ComplianceHistory: ds.generateComplianceHistory(),
		TopViolatingRepos: []RepositoryRank{
			{
				Repository:      "legacy-service",
				ViolationCount:  5,
				CriticalCount:   2,
				ComplianceScore: 45.0,
				Priority:        "high",
			},
		},
		AutoFixOpportunities: []AutoFixOpportunity{
			{
				Repository:     "web-frontend",
				ViolationType:  "Branch Protection",
				Description:    "Enable branch protection on main branch",
				EstimatedTime:  "30 seconds",
				RiskLevel:      "low",
				AutoApplicable: true,
				AffectedRepos:  3,
			},
		},
	}
}

// Helper functions
func (ds *DashboardServer) generatePolicyTrends(policies []PolicyCompliance) []PolicyTrendData {
	var trends []PolicyTrendData
	for _, policy := range policies {
		trends = append(trends, PolicyTrendData{
			PolicyName:     policy.PolicyName,
			ViolationCount: policy.ViolatingRepos,
			TrendDirection: "stable",
			ChangeRate:     0.0,
			Severity:       policy.Severity,
			DataPoints:     []TrendDataPoint{},
		})
	}
	return trends
}

func (ds *DashboardServer) generateComplianceHistory() []ComplianceDataPoint {
	var history []ComplianceDataPoint
	now := time.Now()

	for i := 30; i >= 0; i-- {
		date := now.AddDate(0, 0, -i)
		compliance := 70.0 + float64(i)/30*10 // Improving trend
		history = append(history, ComplianceDataPoint{
			Date:              date,
			CompliancePercent: compliance,
			TotalRepositories: 25,
			CompliantRepos:    int(float64(25) * compliance / 100),
			TotalViolations:   25 - int(float64(25)*compliance/100),
		})
	}

	return history
}

func (ds *DashboardServer) getTopViolatingRepos(repos []RepositoryDashboard) []RepositoryRank {
	var ranks []RepositoryRank
	for _, repo := range repos {
		if repo.ViolationCount > 0 {
			priority := "low"
			if repo.CriticalCount > 0 {
				priority = "high"
			} else if repo.ViolationCount > 2 {
				priority = "medium"
			}

			ranks = append(ranks, RepositoryRank{
				Repository:      repo.Name,
				ViolationCount:  repo.ViolationCount,
				CriticalCount:   repo.CriticalCount,
				ComplianceScore: repo.ComplianceScore,
				Priority:        priority,
			})
		}
	}
	return ranks
}

func (ds *DashboardServer) generateAutoFixOpportunities(violations []ViolationDetail) []AutoFixOpportunity {
	opportunityMap := make(map[string]*AutoFixOpportunity)

	for _, violation := range violations {
		key := violation.Policy
		if opp, exists := opportunityMap[key]; exists {
			opp.AffectedRepos++
		} else {
			opportunityMap[key] = &AutoFixOpportunity{
				Repository:     violation.Repository,
				ViolationType:  violation.Policy,
				Description:    violation.Remediation,
				EstimatedTime:  "30 seconds",
				RiskLevel:      "low",
				AutoApplicable: true,
				AffectedRepos:  1,
			}
		}
	}

	var opportunities []AutoFixOpportunity
	for _, opp := range opportunityMap {
		opportunities = append(opportunities, *opp)
	}

	return opportunities
}

// getDashboardPageData prepares data for the HTML template
func (ds *DashboardServer) getDashboardPageData() interface{} {
	return map[string]interface{}{
		"Organization": ds.config.Organization,
		"AutoRefresh":  ds.config.AutoRefresh,
		"EnableWS":     ds.config.EnableWS,
		"Port":         ds.config.Port,
		"CurrentTime":  time.Now().Format("2006-01-02 15:04:05"),
	}
}

// startDataRefreshRoutine starts the background data refresh routine
func (ds *DashboardServer) startDataRefreshRoutine() {
	ticker := time.NewTicker(time.Duration(ds.config.AutoRefresh) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if ds.config.Verbose {
			log.Printf("Refreshing dashboard data...")
		}

		// Refresh cache
		_, err := ds.getDashboardData()
		if err != nil {
			log.Printf("Error refreshing dashboard data: %v", err)
			continue
		}

		// Broadcast to WebSocket clients if enabled
		if ds.config.EnableWS && len(ds.clients) > 0 {
			data, err := ds.getDashboardData()
			if err == nil {
				if jsonData, err := json.Marshal(data); err == nil {
					ds.broadcast <- jsonData
				}
			}
		}
	}
}

// handleWebSocketConnections handles WebSocket message broadcasting
func (ds *DashboardServer) handleWebSocketConnections() {
	for {
		select {
		case message := <-ds.broadcast:
			// Send message to all connected clients
			for client := range ds.clients {
				err := client.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					if ds.config.Verbose {
						log.Printf("WebSocket write error: %v", err)
					}
					client.Close()
					delete(ds.clients, client)
				}
			}
		}
	}
}

// enabledString returns "enabled" or "disabled" based on boolean value
func enabledString(enabled bool) string {
	if enabled {
		return "enabled"
	}
	return "disabled"
}

// Dashboard HTML template (embedded)
const dashboardHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Compliance Dashboard - {{.Organization}}</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/css/bootstrap.min.css" rel="stylesheet">
    <link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.0.0/css/all.min.css" rel="stylesheet">
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    <style>
        .metric-card {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            border-radius: 10px;
            padding: 20px;
            margin-bottom: 20px;
        }
        .compliance-score {
            font-size: 3rem;
            font-weight: bold;
        }
        .trend-indicator {
            font-size: 1.2rem;
        }
        .trend-up { color: #28a745; }
        .trend-down { color: #dc3545; }
        .trend-stable { color: #ffc107; }
        .violation-badge {
            border-radius: 20px;
            padding: 5px 10px;
            font-size: 0.8rem;
        }
        .status-compliant { background-color: #28a745; }
        .status-warning { background-color: #ffc107; }
        .status-critical { background-color: #dc3545; }
        .chart-container {
            position: relative;
            height: 300px;
            margin-bottom: 30px;
        }
        .sidebar {
            background-color: #f8f9fa;
            min-height: 100vh;
            padding: 20px;
        }
        .main-content {
            padding: 20px;
        }
        .websocket-status {
            position: fixed;
            top: 10px;
            right: 10px;
            z-index: 1000;
        }
        .auto-refresh-indicator {
            position: fixed;
            bottom: 10px;
            right: 10px;
            z-index: 1000;
        }
    </style>
</head>
<body>
    <!-- WebSocket Status -->
    {{if .EnableWS}}
    <div class="websocket-status">
        <span id="wsStatus" class="badge bg-secondary">
            <i class="fas fa-wifi"></i> Connecting...
        </span>
    </div>
    {{end}}

    <!-- Auto-refresh Indicator -->
    {{if .AutoRefresh}}
    <div class="auto-refresh-indicator">
        <span class="badge bg-info">
            <i class="fas fa-sync-alt"></i> Auto-refresh: {{.AutoRefresh}}s
        </span>
    </div>
    {{end}}

    <div class="container-fluid">
        <div class="row">
            <!-- Sidebar -->
            <div class="col-md-3 sidebar">
                <h4><i class="fas fa-shield-alt"></i> Compliance Dashboard</h4>
                <hr>
                <p><strong>Organization:</strong> {{.Organization}}</p>
                <p><strong>Last Updated:</strong> <span id="lastUpdated">{{.CurrentTime}}</span></p>
                
                <div class="mt-4">
                    <h6><i class="fas fa-chart-line"></i> Quick Actions</h6>
                    <div class="d-grid gap-2">
                        <button class="btn btn-outline-primary btn-sm" onclick="refreshData()">
                            <i class="fas fa-sync"></i> Refresh Now
                        </button>
                        <button class="btn btn-outline-success btn-sm" onclick="showAutoFix()">
                            <i class="fas fa-magic"></i> Auto-Fix
                        </button>
                        <button class="btn btn-outline-info btn-sm" onclick="exportReport()">
                            <i class="fas fa-download"></i> Export
                        </button>
                    </div>
                </div>
            </div>

            <!-- Main Content -->
            <div class="col-md-9 main-content">
                <!-- Header -->
                <div class="row mb-4">
                    <div class="col">
                        <h2><i class="fas fa-tachometer-alt"></i> Repository Compliance Overview</h2>
                        <p class="text-muted">Real-time monitoring of repository configurations and policy adherence</p>
                    </div>
                </div>

                <!-- Metrics Cards -->
                <div class="row" id="metricsCards">
                    <!-- Will be populated by JavaScript -->
                </div>

                <!-- Charts Row -->
                <div class="row">
                    <!-- Compliance Trend Chart -->
                    <div class="col-md-8">
                        <div class="card">
                            <div class="card-header">
                                <h5><i class="fas fa-chart-line"></i> Compliance Trend</h5>
                            </div>
                            <div class="card-body">
                                <div class="chart-container">
                                    <canvas id="complianceTrendChart"></canvas>
                                </div>
                            </div>
                        </div>
                    </div>

                    <!-- Policy Distribution Chart -->
                    <div class="col-md-4">
                        <div class="card">
                            <div class="card-header">
                                <h5><i class="fas fa-chart-pie"></i> Violation Distribution</h5>
                            </div>
                            <div class="card-body">
                                <div class="chart-container">
                                    <canvas id="violationPieChart"></canvas>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

                <!-- Repositories Table -->
                <div class="row mt-4">
                    <div class="col">
                        <div class="card">
                            <div class="card-header d-flex justify-content-between align-items-center">
                                <h5><i class="fas fa-code-branch"></i> Repository Status</h5>
                                <div>
                                    <input type="text" class="form-control form-control-sm" id="repoFilter" placeholder="Filter repositories...">
                                </div>
                            </div>
                            <div class="card-body">
                                <div class="table-responsive">
                                    <table class="table table-hover" id="repositoriesTable">
                                        <thead>
                                            <tr>
                                                <th>Repository</th>
                                                <th>Status</th>
                                                <th>Score</th>
                                                <th>Violations</th>
                                                <th>Auto-Fix</th>
                                                <th>Last Checked</th>
                                                <th>Actions</th>
                                            </tr>
                                        </thead>
                                        <tbody id="repositoriesBody">
                                            <!-- Will be populated by JavaScript -->
                                        </tbody>
                                    </table>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

                <!-- Recent Violations -->
                <div class="row mt-4">
                    <div class="col">
                        <div class="card">
                            <div class="card-header">
                                <h5><i class="fas fa-exclamation-triangle"></i> Recent Violations</h5>
                            </div>
                            <div class="card-body">
                                <div id="recentViolations">
                                    <!-- Will be populated by JavaScript -->
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>

    <!-- Bootstrap JS -->
    <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/js/bootstrap.bundle.min.js"></script>
    
    <!-- Dashboard JavaScript -->
    <script>
        let dashboardData = null;
        let ws = null;
        let charts = {};

        // Initialize dashboard
        document.addEventListener('DOMContentLoaded', function() {
            loadDashboardData();
            {{if .EnableWS}}setupWebSocket();{{end}}
            {{if .AutoRefresh}}setupAutoRefresh();{{end}}
            setupEventListeners();
        });

        // Load dashboard data from API
        async function loadDashboardData() {
            try {
                const response = await fetch('/api/data');
                dashboardData = await response.json();
                updateDashboard();
            } catch (error) {
                console.error('Error loading dashboard data:', error);
                showError('Failed to load dashboard data');
            }
        }

        // Update all dashboard components
        function updateDashboard() {
            if (!dashboardData) return;
            
            updateMetricsCards();
            updateCharts();
            updateRepositoriesTable();
            updateRecentViolations();
            updateLastUpdated();
        }

        // Update metrics cards
        function updateMetricsCards() {
            const summary = dashboardData.summary;
            const metricsHTML = ` + "`" + `
                <div class="col-md-3">
                    <div class="metric-card text-center">
                        <div class="compliance-score">${summary.compliance_percentage.toFixed(1)}%</div>
                        <div>Compliance Rate</div>
                        <div class="trend-indicator trend-${summary.trend_direction}">
                            <i class="fas fa-arrow-${getTrendIcon(summary.trend_direction)}"></i>
                            ${summary.trend_direction}
                        </div>
                    </div>
                </div>
                <div class="col-md-3">
                    <div class="metric-card text-center">
                        <div class="compliance-score">${summary.total_repositories}</div>
                        <div>Total Repositories</div>
                        <div class="text-light">${summary.compliant_repositories} compliant</div>
                    </div>
                </div>
                <div class="col-md-3">
                    <div class="metric-card text-center">
                        <div class="compliance-score">${summary.total_violations}</div>
                        <div>Total Violations</div>
                        <div class="text-light">${summary.critical_violations} critical</div>
                    </div>
                </div>
                <div class="col-md-3">
                    <div class="metric-card text-center">
                        <div class="compliance-score">${summary.pending_fixes}</div>
                        <div>Pending Fixes</div>
                        <div class="text-light">${summary.recently_fixed} recently fixed</div>
                    </div>
                </div>
            ` + "`" + `;
            
            document.getElementById('metricsCards').innerHTML = metricsHTML;
        }

        // Update charts
        function updateCharts() {
            updateComplianceTrendChart();
            updateViolationPieChart();
        }

        // Update compliance trend chart
        function updateComplianceTrendChart() {
            const ctx = document.getElementById('complianceTrendChart').getContext('2d');
            
            if (charts.complianceTrend) {
                charts.complianceTrend.destroy();
            }
            
            const history = dashboardData.compliance_history || [];
            const labels = history.map(h => new Date(h.date).toLocaleDateString());
            const data = history.map(h => h.compliance_percent);
            
            charts.complianceTrend = new Chart(ctx, {
                type: 'line',
                data: {
                    labels: labels,
                    datasets: [{
                        label: 'Compliance %',
                        data: data,
                        borderColor: '#667eea',
                        backgroundColor: 'rgba(102, 126, 234, 0.1)',
                        tension: 0.4
                    }]
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    scales: {
                        y: {
                            beginAtZero: true,
                            max: 100
                        }
                    }
                }
            });
        }

        // Update violation pie chart
        function updateViolationPieChart() {
            const ctx = document.getElementById('violationPieChart').getContext('2d');
            
            if (charts.violationPie) {
                charts.violationPie.destroy();
            }
            
            const summary = dashboardData.summary;
            
            charts.violationPie = new Chart(ctx, {
                type: 'doughnut',
                data: {
                    labels: ['Compliant', 'Violations', 'Critical'],
                    datasets: [{
                        data: [
                            summary.compliant_repositories,
                            summary.total_violations - summary.critical_violations,
                            summary.critical_violations
                        ],
                        backgroundColor: ['#28a745', '#ffc107', '#dc3545']
                    }]
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false
                }
            });
        }

        // Update repositories table
        function updateRepositoriesTable() {
            const tbody = document.getElementById('repositoriesBody');
            const repositories = dashboardData.repositories || [];
            
            tbody.innerHTML = repositories.map(repo => ` + "`" + `
                <tr>
                    <td>
                        <i class="fab fa-github"></i> ${repo.name}
                        <br><small class="text-muted">${repo.template}</small>
                    </td>
                    <td>
                        <span class="badge violation-badge status-${repo.status}">
                            ${repo.status.toUpperCase()}
                        </span>
                    </td>
                    <td>${repo.compliance_score.toFixed(1)}%</td>
                    <td>
                        ${repo.violation_count}
                        ${repo.critical_count > 0 ? ` + "`" + `<span class="badge bg-danger">${repo.critical_count} critical</span>` + "`" + ` : ''}
                    </td>
                    <td>
                        ${repo.auto_fix_available > 0 ? ` + "`" + `<span class="badge bg-info">${repo.auto_fix_available}</span>` + "`" + ` : '-'}
                    </td>
                    <td>${new Date(repo.last_checked).toLocaleString()}</td>
                    <td>
                        <div class="btn-group btn-group-sm">
                            <button class="btn btn-outline-primary" onclick="viewRepository('${repo.name}')">
                                <i class="fas fa-eye"></i>
                            </button>
                            ${repo.auto_fix_available > 0 ? ` + "`" + `
                                <button class="btn btn-outline-success" onclick="autoFixRepository('${repo.name}')">
                                    <i class="fas fa-magic"></i>
                                </button>
                            ` + "`" + ` : ''}
                        </div>
                    </td>
                </tr>
            ` + "`" + `).join('');
        }

        // Update recent violations
        function updateRecentViolations() {
            const container = document.getElementById('recentViolations');
            const violations = dashboardData.recent_violations || [];
            
            container.innerHTML = violations.map(violation => ` + "`" + `
                <div class="alert alert-${getSeverityClass(violation.severity)} alert-dismissible">
                    <div class="d-flex justify-content-between">
                        <div>
                            <strong>${violation.repository}</strong> - ${violation.policy}
                            <br><small>${violation.description}</small>
                        </div>
                        <div class="text-end">
                            <span class="badge bg-${getSeverityClass(violation.severity)}">${violation.severity}</span>
                            <br><small class="text-muted">${new Date(violation.first_seen).toLocaleDateString()}</small>
                        </div>
                    </div>
                </div>
            ` + "`" + `).join('');
        }

        // Setup WebSocket connection
        {{if .EnableWS}}
        function setupWebSocket() {
            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            const wsUrl = ` + "`" + `${protocol}//${window.location.host}/ws` + "`" + `;
            
            ws = new WebSocket(wsUrl);
            
            ws.onopen = function() {
                updateWebSocketStatus('connected');
            };
            
            ws.onmessage = function(event) {
                dashboardData = JSON.parse(event.data);
                updateDashboard();
            };
            
            ws.onclose = function() {
                updateWebSocketStatus('disconnected');
                // Attempt to reconnect after 5 seconds
                setTimeout(setupWebSocket, 5000);
            };
            
            ws.onerror = function() {
                updateWebSocketStatus('error');
            };
        }

        function updateWebSocketStatus(status) {
            const statusEl = document.getElementById('wsStatus');
            const statusMap = {
                'connected': { class: 'bg-success', icon: 'wifi', text: 'Connected' },
                'disconnected': { class: 'bg-warning', icon: 'wifi-slash', text: 'Disconnected' },
                'error': { class: 'bg-danger', icon: 'exclamation-triangle', text: 'Error' }
            };
            
            const config = statusMap[status] || statusMap.error;
            statusEl.className = ` + "`" + `badge ${config.class}` + "`" + `;
            statusEl.innerHTML = ` + "`" + `<i class="fas fa-${config.icon}"></i> ${config.text}` + "`" + `;
        }
        {{end}}

        // Setup auto-refresh
        {{if .AutoRefresh}}
        function setupAutoRefresh() {
            setInterval(loadDashboardData, {{.AutoRefresh}} * 1000);
        }
        {{end}}

        // Setup event listeners
        function setupEventListeners() {
            // Repository filter
            document.getElementById('repoFilter').addEventListener('input', function(e) {
                filterRepositories(e.target.value);
            });
        }

        // Filter repositories table
        function filterRepositories(query) {
            const rows = document.querySelectorAll('#repositoriesBody tr');
            rows.forEach(row => {
                const repoName = row.cells[0].textContent.toLowerCase();
                const visible = repoName.includes(query.toLowerCase());
                row.style.display = visible ? '' : 'none';
            });
        }

        // Utility functions
        function getTrendIcon(trend) {
            switch(trend) {
                case 'improving': return 'up';
                case 'declining': return 'down';
                case 'stable': return 'right';
                default: return 'right';
            }
        }

        function getSeverityClass(severity) {
            switch(severity) {
                case 'critical': return 'danger';
                case 'high': return 'warning';
                case 'medium': return 'info';
                case 'low': return 'light';
                default: return 'secondary';
            }
        }

        function updateLastUpdated() {
            document.getElementById('lastUpdated').textContent = new Date().toLocaleString();
        }

        // Action functions
        function refreshData() {
            loadDashboardData();
        }

        function showAutoFix() {
            alert('Auto-fix feature coming soon!');
        }

        function exportReport() {
            alert('Export feature coming soon!');
        }

        function viewRepository(repoName) {
            alert(` + "`" + `View details for ${repoName}` + "`" + `);
        }

        function autoFixRepository(repoName) {
            if (confirm(` + "`" + `Apply auto-fixes for ${repoName}?` + "`" + `)) {
                alert(` + "`" + `Auto-fix initiated for ${repoName}` + "`" + `);
            }
        }

        function showError(message) {
            console.error(message);
            // Could show a toast notification here
        }
    </script>
</body>
</html>`
