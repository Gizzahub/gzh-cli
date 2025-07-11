package monitoring

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// GrafanaIntegration manages integration with Grafana for dashboards and alerting
type GrafanaIntegration struct {
	logger     *zap.Logger
	config     *GrafanaConfig
	httpClient *http.Client
	mutex      sync.RWMutex

	// Dashboard and alert rule cache
	dashboards map[string]*Dashboard
	alertRules map[string]*GrafanaAlertRule
}

// GrafanaConfig represents Grafana integration configuration
type GrafanaConfig struct {
	Enabled    bool                   `yaml:"enabled" json:"enabled"`
	BaseURL    string                 `yaml:"base_url" json:"base_url"`
	APIKey     string                 `yaml:"api_key" json:"api_key"`
	Username   string                 `yaml:"username" json:"username"`
	Password   string                 `yaml:"password" json:"password"`
	OrgID      int                    `yaml:"org_id" json:"org_id"`
	Timeout    time.Duration          `yaml:"timeout" json:"timeout"`
	Dashboards []DashboardTemplate    `yaml:"dashboards" json:"dashboards"`
	AlertRules []AlertRuleTemplate    `yaml:"alert_rules" json:"alert_rules"`
	Variables  map[string]interface{} `yaml:"variables" json:"variables"`
	Tags       []string               `yaml:"tags" json:"tags"`
}

// Dashboard represents a Grafana dashboard
type Dashboard struct {
	ID          int                    `json:"id,omitempty"`
	UID         string                 `json:"uid,omitempty"`
	Title       string                 `json:"title"`
	Description string                 `json:"description,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Timezone    string                 `json:"timezone,omitempty"`
	Panels      []Panel                `json:"panels"`
	Variables   []Variable             `json:"templating"`
	Time        TimeRange              `json:"time"`
	Refresh     string                 `json:"refresh,omitempty"`
	Version     int                    `json:"version,omitempty"`
	Metadata    map[string]interface{} `json:"meta,omitempty"`
}

// Panel represents a dashboard panel
type Panel struct {
	ID          int                    `json:"id"`
	Title       string                 `json:"title"`
	Type        string                 `json:"type"`
	GridPos     GridPosition           `json:"gridPos"`
	Targets     []QueryTarget          `json:"targets"`
	Options     map[string]interface{} `json:"options,omitempty"`
	FieldConfig FieldConfig            `json:"fieldConfig,omitempty"`
	Alert       *PanelAlert            `json:"alert,omitempty"`
}

// GridPosition represents panel grid position
type GridPosition struct {
	H int `json:"h"`
	W int `json:"w"`
	X int `json:"x"`
	Y int `json:"y"`
}

// QueryTarget represents a data source query
type QueryTarget struct {
	RefID      string                 `json:"refId"`
	Expr       string                 `json:"expr"`
	Interval   string                 `json:"interval,omitempty"`
	Legend     string                 `json:"legendFormat,omitempty"`
	Datasource map[string]interface{} `json:"datasource"`
}

// FieldConfig represents panel field configuration
type FieldConfig struct {
	Defaults  FieldDefaults          `json:"defaults"`
	Overrides []FieldOverride        `json:"overrides,omitempty"`
	Custom    map[string]interface{} `json:"custom,omitempty"`
}

// FieldDefaults represents default field settings
type FieldDefaults struct {
	Unit       string                 `json:"unit,omitempty"`
	Min        *float64               `json:"min,omitempty"`
	Max        *float64               `json:"max,omitempty"`
	Decimals   *int                   `json:"decimals,omitempty"`
	Thresholds *Thresholds            `json:"thresholds,omitempty"`
	Mappings   []ValueMapping         `json:"mappings,omitempty"`
	Custom     map[string]interface{} `json:"custom,omitempty"`
}

// Thresholds represents threshold configuration
type Thresholds struct {
	Mode  string      `json:"mode"`
	Steps []Threshold `json:"steps"`
}

// Threshold represents a single threshold step
type Threshold struct {
	Color string   `json:"color"`
	Value *float64 `json:"value"`
}

// ValueMapping represents value mapping configuration
type ValueMapping struct {
	Type    string                 `json:"type"`
	Value   string                 `json:"value,omitempty"`
	Text    string                 `json:"text,omitempty"`
	Options map[string]interface{} `json:"options,omitempty"`
}

// FieldOverride represents field override configuration
type FieldOverride struct {
	Matcher    FieldMatcher    `json:"matcher"`
	Properties []FieldProperty `json:"properties"`
}

// FieldMatcher represents field matcher configuration
type FieldMatcher struct {
	ID      string `json:"id"`
	Options string `json:"options"`
}

// FieldProperty represents field property configuration
type FieldProperty struct {
	ID    string      `json:"id"`
	Value interface{} `json:"value"`
}

// PanelAlert represents panel alert configuration
type PanelAlert struct {
	ID                  int                   `json:"id,omitempty"`
	Name                string                `json:"name"`
	Message             string                `json:"message,omitempty"`
	Frequency           string                `json:"frequency"`
	Conditions          []AlertCondition      `json:"conditions"`
	ExecutionErrorState string                `json:"executionErrorState,omitempty"`
	NoDataState         string                `json:"noDataState,omitempty"`
	Notifications       []NotificationChannel `json:"notifications,omitempty"`
}

// Variable represents a dashboard template variable
type Variable struct {
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Label      string                 `json:"label,omitempty"`
	Query      string                 `json:"query,omitempty"`
	Datasource map[string]interface{} `json:"datasource,omitempty"`
	Refresh    int                    `json:"refresh,omitempty"`
	Multi      bool                   `json:"multi,omitempty"`
	IncludeAll bool                   `json:"includeAll,omitempty"`
	AllValue   string                 `json:"allValue,omitempty"`
	Options    []VariableOption       `json:"options,omitempty"`
	Current    VariableOption         `json:"current,omitempty"`
	Hide       int                    `json:"hide,omitempty"`
}

// VariableOption represents a template variable option
type VariableOption struct {
	Text     string `json:"text"`
	Value    string `json:"value"`
	Selected bool   `json:"selected,omitempty"`
}

// TimeRange represents dashboard time range
type TimeRange struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// GrafanaAlertRule represents a Grafana alert rule
type GrafanaAlertRule struct {
	ID           int               `json:"id,omitempty"`
	UID          string            `json:"uid,omitempty"`
	Title        string            `json:"title"`
	Condition    string            `json:"condition"`
	Data         []AlertQuery      `json:"data"`
	NoDataState  string            `json:"noDataState"`
	ExecErrState string            `json:"execErrState"`
	For          time.Duration     `json:"for"`
	Annotations  map[string]string `json:"annotations,omitempty"`
	Labels       map[string]string `json:"labels,omitempty"`
	FolderUID    string            `json:"folderUID,omitempty"`
}

// AlertQuery represents an alert rule query
type AlertQuery struct {
	RefID             string                 `json:"refId"`
	QueryType         string                 `json:"queryType,omitempty"`
	Model             map[string]interface{} `json:"model"`
	DatasourceUID     string                 `json:"datasourceUid"`
	RelativeTimeRange *TimeRange             `json:"relativeTimeRange,omitempty"`
}

// NotificationChannel represents a notification channel
type NotificationChannel struct {
	ID   int    `json:"id,omitempty"`
	UID  string `json:"uid,omitempty"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// DashboardTemplate represents a dashboard template configuration
type DashboardTemplate struct {
	Name        string                 `json:"name"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Tags        []string               `json:"tags"`
	Variables   []VariableTemplate     `json:"variables"`
	Panels      []PanelTemplate        `json:"panels"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// VariableTemplate represents a variable template
type VariableTemplate struct {
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	Label      string   `json:"label"`
	Query      string   `json:"query"`
	Multi      bool     `json:"multi"`
	IncludeAll bool     `json:"includeAll"`
	Options    []string `json:"options"`
}

// PanelTemplate represents a panel template
type PanelTemplate struct {
	Title      string                 `json:"title"`
	Type       string                 `json:"type"`
	GridPos    GridPosition           `json:"gridPos"`
	Queries    []string               `json:"queries"`
	Unit       string                 `json:"unit"`
	Thresholds []ThresholdTemplate    `json:"thresholds"`
	Options    map[string]interface{} `json:"options"`
}

// ThresholdTemplate represents a threshold template
type ThresholdTemplate struct {
	Color string  `json:"color"`
	Value float64 `json:"value"`
}

// AlertRuleTemplate represents an alert rule template
type AlertRuleTemplate struct {
	Title       string            `json:"title"`
	Condition   string            `json:"condition"`
	Queries     []string          `json:"queries"`
	For         time.Duration     `json:"for"`
	Annotations map[string]string `json:"annotations"`
	Labels      map[string]string `json:"labels"`
}

// NewGrafanaIntegration creates a new Grafana integration instance
func NewGrafanaIntegration(logger *zap.Logger, config *GrafanaConfig) *GrafanaIntegration {
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &GrafanaIntegration{
		logger: logger,
		config: config,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		dashboards: make(map[string]*Dashboard),
		alertRules: make(map[string]*GrafanaAlertRule),
	}
}

// Start initializes the Grafana integration
func (gi *GrafanaIntegration) Start(ctx context.Context) error {
	if !gi.config.Enabled {
		gi.logger.Info("Grafana integration disabled")
		return nil
	}

	gi.logger.Info("Starting Grafana integration",
		zap.String("base_url", gi.config.BaseURL))

	// Test connection
	if err := gi.testConnection(ctx); err != nil {
		return fmt.Errorf("failed to connect to Grafana: %w", err)
	}

	// Deploy dashboards
	if err := gi.deployDashboards(ctx); err != nil {
		gi.logger.Error("Failed to deploy dashboards", zap.Error(err))
	}

	// Deploy alert rules
	if err := gi.deployAlertRules(ctx); err != nil {
		gi.logger.Error("Failed to deploy alert rules", zap.Error(err))
	}

	return nil
}

// Stop shuts down the Grafana integration
func (gi *GrafanaIntegration) Stop(ctx context.Context) error {
	gi.logger.Info("Stopping Grafana integration")
	return nil
}

// testConnection tests the connection to Grafana
func (gi *GrafanaIntegration) testConnection(ctx context.Context) error {
	url := fmt.Sprintf("%s/api/health", gi.config.BaseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	gi.setAuthHeaders(req)

	resp, err := gi.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Grafana health check failed: %d", resp.StatusCode)
	}

	return nil
}

// setAuthHeaders sets authentication headers for Grafana API requests
func (gi *GrafanaIntegration) setAuthHeaders(req *http.Request) {
	if gi.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+gi.config.APIKey)
	} else if gi.config.Username != "" && gi.config.Password != "" {
		req.SetBasicAuth(gi.config.Username, gi.config.Password)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if gi.config.OrgID > 0 {
		req.Header.Set("X-Grafana-Org-Id", fmt.Sprintf("%d", gi.config.OrgID))
	}
}

// deployDashboards deploys dashboard templates to Grafana
func (gi *GrafanaIntegration) deployDashboards(ctx context.Context) error {
	for _, template := range gi.config.Dashboards {
		dashboard, err := gi.generateDashboard(&template)
		if err != nil {
			gi.logger.Error("Failed to generate dashboard",
				zap.String("name", template.Name),
				zap.Error(err))
			continue
		}

		if err := gi.deployDashboard(ctx, dashboard); err != nil {
			gi.logger.Error("Failed to deploy dashboard",
				zap.String("title", dashboard.Title),
				zap.Error(err))
			continue
		}

		gi.logger.Info("Dashboard deployed successfully",
			zap.String("title", dashboard.Title))
	}

	return nil
}

// generateDashboard generates a dashboard from a template
func (gi *GrafanaIntegration) generateDashboard(template *DashboardTemplate) (*Dashboard, error) {
	dashboard := &Dashboard{
		Title:       template.Title,
		Description: template.Description,
		Tags:        template.Tags,
		Timezone:    "browser",
		Time: TimeRange{
			From: "now-1h",
			To:   "now",
		},
		Refresh:   "30s",
		Panels:    []Panel{},
		Variables: []Variable{},
	}

	// Generate variables
	for _, varTemplate := range template.Variables {
		variable := gi.generateVariable(&varTemplate)
		dashboard.Variables = append(dashboard.Variables, variable)
	}

	// Generate panels
	for i, panelTemplate := range template.Panels {
		panel := gi.generatePanel(i+1, &panelTemplate)
		dashboard.Panels = append(dashboard.Panels, panel)
	}

	return dashboard, nil
}

// generateVariable generates a template variable
func (gi *GrafanaIntegration) generateVariable(template *VariableTemplate) Variable {
	variable := Variable{
		Name:       template.Name,
		Type:       template.Type,
		Label:      template.Label,
		Query:      template.Query,
		Multi:      template.Multi,
		IncludeAll: template.IncludeAll,
		Refresh:    1,
		Hide:       0,
	}

	// Add datasource for query variables
	if template.Type == "query" {
		variable.Datasource = map[string]interface{}{
			"type": "prometheus",
			"uid":  "prometheus",
		}
	}

	// Generate options
	for _, opt := range template.Options {
		variable.Options = append(variable.Options, VariableOption{
			Text:  opt,
			Value: opt,
		})
	}

	return variable
}

// generatePanel generates a dashboard panel from a template
func (gi *GrafanaIntegration) generatePanel(id int, template *PanelTemplate) Panel {
	panel := Panel{
		ID:      id,
		Title:   template.Title,
		Type:    template.Type,
		GridPos: template.GridPos,
		Targets: []QueryTarget{},
		FieldConfig: FieldConfig{
			Defaults: FieldDefaults{
				Unit: template.Unit,
			},
		},
	}

	// Generate query targets
	for i, query := range template.Queries {
		target := QueryTarget{
			RefID:    fmt.Sprintf("%c", 'A'+i),
			Expr:     gi.substituteVariables(query),
			Interval: "30s",
			Datasource: map[string]interface{}{
				"type": "prometheus",
				"uid":  "prometheus",
			},
		}
		panel.Targets = append(panel.Targets, target)
	}

	// Generate thresholds
	if len(template.Thresholds) > 0 {
		thresholds := &Thresholds{
			Mode:  "absolute",
			Steps: []Threshold{},
		}

		for _, th := range template.Thresholds {
			thresholds.Steps = append(thresholds.Steps, Threshold{
				Color: th.Color,
				Value: &th.Value,
			})
		}

		panel.FieldConfig.Defaults.Thresholds = thresholds
	}

	// Set panel options
	if template.Options != nil {
		panel.Options = template.Options
	}

	return panel
}

// substituteVariables substitutes template variables in queries
func (gi *GrafanaIntegration) substituteVariables(query string) string {
	result := query
	for key, value := range gi.config.Variables {
		placeholder := fmt.Sprintf("{{%s}}", key)
		replacement := fmt.Sprintf("%v", value)
		result = strings.ReplaceAll(result, placeholder, replacement)
	}
	return result
}

// deployDashboard deploys a dashboard to Grafana
func (gi *GrafanaIntegration) deployDashboard(ctx context.Context, dashboard *Dashboard) error {
	payload := map[string]interface{}{
		"dashboard": dashboard,
		"overwrite": true,
		"message":   "Deployed by gzh-manager",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/api/dashboards/db", gi.config.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	gi.setAuthHeaders(req)

	resp, err := gi.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to deploy dashboard: %d - %s", resp.StatusCode, string(body))
	}

	// Cache the deployed dashboard
	gi.mutex.Lock()
	gi.dashboards[dashboard.Title] = dashboard
	gi.mutex.Unlock()

	return nil
}

// deployAlertRules deploys alert rule templates to Grafana
func (gi *GrafanaIntegration) deployAlertRules(ctx context.Context) error {
	for _, template := range gi.config.AlertRules {
		alertRule, err := gi.generateAlertRule(&template)
		if err != nil {
			gi.logger.Error("Failed to generate alert rule",
				zap.String("title", template.Title),
				zap.Error(err))
			continue
		}

		if err := gi.deployAlertRule(ctx, alertRule); err != nil {
			gi.logger.Error("Failed to deploy alert rule",
				zap.String("title", alertRule.Title),
				zap.Error(err))
			continue
		}

		gi.logger.Info("Alert rule deployed successfully",
			zap.String("title", alertRule.Title))
	}

	return nil
}

// generateAlertRule generates an alert rule from a template
func (gi *GrafanaIntegration) generateAlertRule(template *AlertRuleTemplate) (*GrafanaAlertRule, error) {
	alertRule := &GrafanaAlertRule{
		Title:        template.Title,
		Condition:    template.Condition,
		NoDataState:  "NoData",
		ExecErrState: "Alerting",
		For:          template.For,
		Annotations:  template.Annotations,
		Labels:       template.Labels,
		Data:         []AlertQuery{},
	}

	// Generate alert queries
	for i, query := range template.Queries {
		alertQuery := AlertQuery{
			RefID:     fmt.Sprintf("%c", 'A'+i),
			QueryType: "",
			Model: map[string]interface{}{
				"expr":     gi.substituteVariables(query),
				"interval": "30s",
				"refId":    fmt.Sprintf("%c", 'A'+i),
				"datasource": map[string]interface{}{
					"type": "prometheus",
					"uid":  "prometheus",
				},
			},
			DatasourceUID: "prometheus",
		}
		alertRule.Data = append(alertRule.Data, alertQuery)
	}

	return alertRule, nil
}

// deployAlertRule deploys an alert rule to Grafana
func (gi *GrafanaIntegration) deployAlertRule(ctx context.Context, alertRule *GrafanaAlertRule) error {
	jsonData, err := json.Marshal(alertRule)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/api/ruler/grafana/api/v1/rules/default", gi.config.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	gi.setAuthHeaders(req)

	resp, err := gi.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to deploy alert rule: %d - %s", resp.StatusCode, string(body))
	}

	// Cache the deployed alert rule
	gi.mutex.Lock()
	gi.alertRules[alertRule.Title] = alertRule
	gi.mutex.Unlock()

	return nil
}

// GetDashboards returns all deployed dashboards
func (gi *GrafanaIntegration) GetDashboards() map[string]*Dashboard {
	gi.mutex.RLock()
	defer gi.mutex.RUnlock()

	dashboards := make(map[string]*Dashboard)
	for k, v := range gi.dashboards {
		dashboards[k] = v
	}
	return dashboards
}

// GetAlertRules returns all deployed alert rules
func (gi *GrafanaIntegration) GetAlertRules() map[string]*GrafanaAlertRule {
	gi.mutex.RLock()
	defer gi.mutex.RUnlock()

	alertRules := make(map[string]*GrafanaAlertRule)
	for k, v := range gi.alertRules {
		alertRules[k] = v
	}
	return alertRules
}

// HealthCheck performs a health check of the Grafana integration
func (gi *GrafanaIntegration) HealthCheck(ctx context.Context) error {
	if !gi.config.Enabled {
		return nil
	}

	return gi.testConnection(ctx)
}
