package monitoring

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestGrafanaIntegration_Creation(t *testing.T) {
	logger := zap.NewNop()
	config := &GrafanaConfig{
		Enabled: true,
		BaseURL: "http://localhost:3000",
		APIKey:  "test-api-key",
		OrgID:   1,
		Timeout: 30 * time.Second,
	}

	gi := NewGrafanaIntegration(logger, config)

	assert.NotNil(t, gi)
	assert.Equal(t, config, gi.config)
	assert.NotNil(t, gi.httpClient)
	assert.NotNil(t, gi.dashboards)
	assert.NotNil(t, gi.alertRules)
}

func TestGrafanaIntegration_TestConnection(t *testing.T) {
	logger := zap.NewNop()

	// Create mock Grafana server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/health" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "ok"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	config := &GrafanaConfig{
		Enabled: true,
		BaseURL: server.URL,
		APIKey:  "test-api-key",
		Timeout: 5 * time.Second,
	}

	gi := NewGrafanaIntegration(logger, config)
	ctx := context.Background()

	t.Run("Successful connection", func(t *testing.T) {
		err := gi.testConnection(ctx)
		assert.NoError(t, err)
	})

	t.Run("Failed connection", func(t *testing.T) {
		// Change URL to cause failure
		gi.config.BaseURL = "http://invalid-url:9999"
		err := gi.testConnection(ctx)
		assert.Error(t, err)
	})
}

func TestGrafanaIntegration_DashboardGeneration(t *testing.T) {
	logger := zap.NewNop()
	config := &GrafanaConfig{
		Enabled: false, // Don't need actual connection for this test
		Variables: map[string]interface{}{
			"instance": "localhost:9090",
			"job":      "gzh-manager",
		},
	}

	gi := NewGrafanaIntegration(logger, config)

	t.Run("Generate dashboard from template", func(t *testing.T) {
		template := &DashboardTemplate{
			Name:        "system-monitoring",
			Title:       "System Monitoring",
			Description: "System metrics dashboard",
			Tags:        []string{"system", "monitoring"},
			Variables: []VariableTemplate{
				{
					Name:       "instance",
					Type:       "query",
					Label:      "Instance",
					Query:      "label_values(up, instance)",
					Multi:      true,
					IncludeAll: true,
				},
			},
			Panels: []PanelTemplate{
				{
					Title: "CPU Usage",
					Type:  "stat",
					GridPos: GridPosition{
						H: 8, W: 12, X: 0, Y: 0,
					},
					Queries: []string{
						"100 - (avg by (instance) (rate(node_cpu_seconds_total{mode=\"idle\",instance=\"{{instance}}\"}[5m])) * 100)",
					},
					Unit: "percent",
					Thresholds: []ThresholdTemplate{
						{Color: "green", Value: 0},
						{Color: "yellow", Value: 70},
						{Color: "red", Value: 90},
					},
				},
				{
					Title: "Memory Usage",
					Type:  "timeseries",
					GridPos: GridPosition{
						H: 8, W: 12, X: 12, Y: 0,
					},
					Queries: []string{
						"(1 - (node_memory_MemAvailable_bytes{instance=\"{{instance}}\"} / node_memory_MemTotal_bytes{instance=\"{{instance}}\"})) * 100",
					},
					Unit: "percent",
				},
			},
		}

		dashboard, err := gi.generateDashboard(template)
		require.NoError(t, err)

		assert.Equal(t, "System Monitoring", dashboard.Title)
		assert.Equal(t, "System metrics dashboard", dashboard.Description)
		assert.Equal(t, []string{"system", "monitoring"}, dashboard.Tags)
		assert.Len(t, dashboard.Variables, 1)
		assert.Len(t, dashboard.Panels, 2)

		// Verify variable generation
		variable := dashboard.Variables[0]
		assert.Equal(t, "instance", variable.Name)
		assert.Equal(t, "query", variable.Type)
		assert.Equal(t, "Instance", variable.Label)
		assert.True(t, variable.Multi)
		assert.True(t, variable.IncludeAll)

		// Verify panel generation
		cpuPanel := dashboard.Panels[0]
		assert.Equal(t, "CPU Usage", cpuPanel.Title)
		assert.Equal(t, "stat", cpuPanel.Type)
		assert.Len(t, cpuPanel.Targets, 1)
		assert.Equal(t, "percent", cpuPanel.FieldConfig.Defaults.Unit)

		memoryPanel := dashboard.Panels[1]
		assert.Equal(t, "Memory Usage", memoryPanel.Title)
		assert.Equal(t, "timeseries", memoryPanel.Type)
		assert.Len(t, memoryPanel.Targets, 1)
	})

	t.Run("Variable substitution", func(t *testing.T) {
		query := "up{instance=\"{{instance}}\",job=\"{{job}}\"}"
		result := gi.substituteVariables(query)
		expected := "up{instance=\"localhost:9090\",job=\"gzh-manager\"}"
		assert.Equal(t, expected, result)
	})
}

func TestGrafanaIntegration_AlertRuleGeneration(t *testing.T) {
	logger := zap.NewNop()
	config := &GrafanaConfig{
		Enabled: false,
		Variables: map[string]interface{}{
			"instance": "localhost:9090",
		},
	}

	gi := NewGrafanaIntegration(logger, config)

	t.Run("Generate alert rule from template", func(t *testing.T) {
		template := &AlertRuleTemplate{
			Title:     "High CPU Usage",
			Condition: "A",
			Queries: []string{
				"100 - (avg by (instance) (rate(node_cpu_seconds_total{mode=\"idle\",instance=\"{{instance}}\"}[5m])) * 100) > 80",
			},
			For: 5 * time.Minute,
			Annotations: map[string]string{
				"description": "CPU usage is above 80% for more than 5 minutes",
				"summary":     "High CPU usage detected",
			},
			Labels: map[string]string{
				"severity": "warning",
				"team":     "infrastructure",
			},
		}

		alertRule, err := gi.generateAlertRule(template)
		require.NoError(t, err)

		assert.Equal(t, "High CPU Usage", alertRule.Title)
		assert.Equal(t, "A", alertRule.Condition)
		assert.Equal(t, 5*time.Minute, alertRule.For)
		assert.Equal(t, "NoData", alertRule.NoDataState)
		assert.Equal(t, "Alerting", alertRule.ExecErrState)
		assert.Len(t, alertRule.Data, 1)

		// Verify alert query
		alertQuery := alertRule.Data[0]
		assert.Equal(t, "A", alertQuery.RefID)
		assert.Equal(t, "prometheus", alertQuery.DatasourceUID)
		assert.Contains(t, alertQuery.Model, "expr")

		// Verify annotations and labels
		assert.Equal(t, "High CPU usage detected", alertRule.Annotations["summary"])
		assert.Equal(t, "warning", alertRule.Labels["severity"])
	})
}

func TestGrafanaIntegration_DeployDashboard(t *testing.T) {
	logger := zap.NewNop()

	// Create mock Grafana server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/health":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "ok"}`))
		case "/api/dashboards/db":
			if r.Method == "POST" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"id": 1, "uid": "test-uid", "url": "/d/test-uid/test-dashboard"}`))
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	config := &GrafanaConfig{
		Enabled: true,
		BaseURL: server.URL,
		APIKey:  "test-api-key",
		Dashboards: []DashboardTemplate{
			{
				Name:  "test-dashboard",
				Title: "Test Dashboard",
				Variables: []VariableTemplate{
					{
						Name:  "instance",
						Type:  "query",
						Query: "label_values(up, instance)",
					},
				},
				Panels: []PanelTemplate{
					{
						Title:   "Test Panel",
						Type:    "stat",
						GridPos: GridPosition{H: 8, W: 12, X: 0, Y: 0},
						Queries: []string{"up"},
					},
				},
			},
		},
	}

	gi := NewGrafanaIntegration(logger, config)
	ctx := context.Background()

	t.Run("Deploy dashboard successfully", func(t *testing.T) {
		err := gi.Start(ctx)
		assert.NoError(t, err)

		// Verify dashboard was cached
		dashboards := gi.GetDashboards()
		assert.Len(t, dashboards, 1)
		assert.Contains(t, dashboards, "Test Dashboard")
	})
}

func TestGrafanaIntegration_DeployAlertRule(t *testing.T) {
	logger := zap.NewNop()

	// Create mock Grafana server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/health":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "ok"}`))
		case "/api/ruler/grafana/api/v1/rules/default":
			if r.Method == "POST" {
				w.WriteHeader(http.StatusAccepted)
				w.Write([]byte(`{"message": "rule created"}`))
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	config := &GrafanaConfig{
		Enabled: true,
		BaseURL: server.URL,
		APIKey:  "test-api-key",
		AlertRules: []AlertRuleTemplate{
			{
				Title:     "Test Alert",
				Condition: "A",
				Queries:   []string{"up == 0"},
				For:       1 * time.Minute,
				Annotations: map[string]string{
					"summary": "Service is down",
				},
				Labels: map[string]string{
					"severity": "critical",
				},
			},
		},
	}

	gi := NewGrafanaIntegration(logger, config)
	ctx := context.Background()

	t.Run("Deploy alert rule successfully", func(t *testing.T) {
		err := gi.Start(ctx)
		assert.NoError(t, err)

		// Verify alert rule was cached
		alertRules := gi.GetAlertRules()
		assert.Len(t, alertRules, 1)
		assert.Contains(t, alertRules, "Test Alert")
	})
}

func TestGrafanaIntegration_AuthHeaders(t *testing.T) {
	logger := zap.NewNop()

	t.Run("API key authentication", func(t *testing.T) {
		config := &GrafanaConfig{
			Enabled: false,
			APIKey:  "test-api-key",
			OrgID:   5,
		}

		gi := NewGrafanaIntegration(logger, config)
		req, _ := http.NewRequest("GET", "http://example.com", nil)

		gi.setAuthHeaders(req)

		assert.Equal(t, "Bearer test-api-key", req.Header.Get("Authorization"))
		assert.Equal(t, "5", req.Header.Get("X-Grafana-Org-Id"))
		assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
		assert.Equal(t, "application/json", req.Header.Get("Accept"))
	})

	t.Run("Basic authentication", func(t *testing.T) {
		config := &GrafanaConfig{
			Enabled:  false,
			Username: "admin",
			Password: "password",
		}

		gi := NewGrafanaIntegration(logger, config)
		req, _ := http.NewRequest("GET", "http://example.com", nil)

		gi.setAuthHeaders(req)

		// Basic auth should be set (we can't easily test the exact value)
		assert.NotEmpty(t, req.Header.Get("Authorization"))
		assert.Contains(t, req.Header.Get("Authorization"), "Basic")
	})
}

func TestGrafanaIntegration_DisabledIntegration(t *testing.T) {
	logger := zap.NewNop()
	config := &GrafanaConfig{
		Enabled: false,
	}

	gi := NewGrafanaIntegration(logger, config)
	ctx := context.Background()

	t.Run("Disabled integration should not start", func(t *testing.T) {
		err := gi.Start(ctx)
		assert.NoError(t, err)

		dashboards := gi.GetDashboards()
		assert.Len(t, dashboards, 0)

		alertRules := gi.GetAlertRules()
		assert.Len(t, alertRules, 0)
	})

	t.Run("Health check should pass when disabled", func(t *testing.T) {
		err := gi.HealthCheck(ctx)
		assert.NoError(t, err)
	})
}

func TestGrafanaIntegration_ErrorHandling(t *testing.T) {
	logger := zap.NewNop()

	// Create mock server that returns errors
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/health":
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"error": "service unavailable"}`))
		case "/api/dashboards/db":
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "invalid dashboard"}`))
		case "/api/ruler/grafana/api/v1/rules/default":
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "internal server error"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	config := &GrafanaConfig{
		Enabled: true,
		BaseURL: server.URL,
		APIKey:  "test-api-key",
		Dashboards: []DashboardTemplate{
			{
				Name:   "test-dashboard",
				Title:  "Test Dashboard",
				Panels: []PanelTemplate{},
			},
		},
		AlertRules: []AlertRuleTemplate{
			{
				Title:   "Test Alert",
				Queries: []string{"up == 0"},
			},
		},
	}

	gi := NewGrafanaIntegration(logger, config)
	ctx := context.Background()

	t.Run("Health check failure", func(t *testing.T) {
		err := gi.HealthCheck(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "503")
	})

	t.Run("Start with connection failure", func(t *testing.T) {
		err := gi.Start(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to connect to Grafana")
	})
}

func TestDashboardStructures(t *testing.T) {
	t.Run("Dashboard JSON serialization", func(t *testing.T) {
		dashboard := Dashboard{
			Title:       "Test Dashboard",
			Description: "Test description",
			Tags:        []string{"test", "monitoring"},
			Timezone:    "UTC",
			Time: TimeRange{
				From: "now-1h",
				To:   "now",
			},
			Panels: []Panel{
				{
					ID:    1,
					Title: "Test Panel",
					Type:  "stat",
					GridPos: GridPosition{
						H: 8, W: 12, X: 0, Y: 0,
					},
					Targets: []QueryTarget{
						{
							RefID: "A",
							Expr:  "up",
							Datasource: map[string]interface{}{
								"type": "prometheus",
								"uid":  "prometheus",
							},
						},
					},
				},
			},
		}

		jsonData, err := json.Marshal(dashboard)
		assert.NoError(t, err)
		assert.Contains(t, string(jsonData), "Test Dashboard")
		assert.Contains(t, string(jsonData), "Test Panel")

		// Test deserialization
		var deserializedDashboard Dashboard
		err = json.Unmarshal(jsonData, &deserializedDashboard)
		assert.NoError(t, err)
		assert.Equal(t, dashboard.Title, deserializedDashboard.Title)
	})

	t.Run("Alert rule JSON serialization", func(t *testing.T) {
		alertRule := GrafanaAlertRule{
			Title:     "Test Alert",
			Condition: "A",
			For:       5 * time.Minute,
			Data: []AlertQuery{
				{
					RefID: "A",
					Model: map[string]interface{}{
						"expr": "up == 0",
					},
					DatasourceUID: "prometheus",
				},
			},
			Annotations: map[string]string{
				"summary": "Test alert",
			},
			Labels: map[string]string{
				"severity": "warning",
			},
		}

		jsonData, err := json.Marshal(alertRule)
		assert.NoError(t, err)
		assert.Contains(t, string(jsonData), "Test Alert")
		assert.Contains(t, string(jsonData), "up == 0")

		// Test deserialization
		var deserializedAlertRule GrafanaAlertRule
		err = json.Unmarshal(jsonData, &deserializedAlertRule)
		assert.NoError(t, err)
		assert.Equal(t, alertRule.Title, deserializedAlertRule.Title)
	})
}

func TestGrafanaConfig_Validation(t *testing.T) {
	testCases := []struct {
		name   string
		config *GrafanaConfig
		valid  bool
	}{
		{
			name: "Valid configuration with API key",
			config: &GrafanaConfig{
				Enabled: true,
				BaseURL: "http://localhost:3000",
				APIKey:  "test-api-key",
				OrgID:   1,
			},
			valid: true,
		},
		{
			name: "Valid configuration with basic auth",
			config: &GrafanaConfig{
				Enabled:  true,
				BaseURL:  "http://localhost:3000",
				Username: "admin",
				Password: "password",
				OrgID:    1,
			},
			valid: true,
		},
		{
			name: "Disabled configuration",
			config: &GrafanaConfig{
				Enabled: false,
			},
			valid: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger := zap.NewNop()
			gi := NewGrafanaIntegration(logger, tc.config)
			assert.NotNil(t, gi)

			if tc.valid {
				assert.Equal(t, tc.config, gi.config)
			}
		})
	}
}
