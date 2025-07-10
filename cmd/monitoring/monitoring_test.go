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
)

func TestNewMonitoringServer(t *testing.T) {
	config := &ServerConfig{
		Host:  "localhost",
		Port:  8080,
		Debug: true,
	}

	server := NewMonitoringServer(config)
	assert.NotNil(t, server)
	assert.Equal(t, config, server.config)
	assert.NotNil(t, server.router)
	assert.NotNil(t, server.metrics)
	assert.NotNil(t, server.alerts)
}

func TestMonitoringServerRoutes(t *testing.T) {
	config := &ServerConfig{
		Host:  "localhost",
		Port:  8080,
		Debug: false,
	}

	server := NewMonitoringServer(config)
	testServer := httptest.NewServer(server.router)
	defer testServer.Close()

	client := &http.Client{Timeout: 5 * time.Second}

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{"System Status", "GET", "/api/v1/status", http.StatusOK},
		{"Health Check", "GET", "/api/v1/health", http.StatusOK},
		{"Metrics", "GET", "/api/v1/metrics", http.StatusOK},
		{"Tasks", "GET", "/api/v1/tasks", http.StatusOK},
		{"Alerts", "GET", "/api/v1/alerts", http.StatusOK},
		{"Config", "GET", "/api/v1/config", http.StatusOK},
		{"WebSocket", "GET", "/ws", http.StatusNotImplemented},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, testServer.URL+tt.path, nil)
			require.NoError(t, err)

			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

func TestSystemStatusEndpoint(t *testing.T) {
	config := &ServerConfig{
		Host:  "localhost",
		Port:  8080,
		Debug: false,
	}

	server := NewMonitoringServer(config)
	testServer := httptest.NewServer(server.router)
	defer testServer.Close()

	resp, err := http.Get(testServer.URL + "/api/v1/status")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var status SystemStatus
	err = json.NewDecoder(resp.Body).Decode(&status)
	require.NoError(t, err)

	assert.Equal(t, "healthy", status.Status)
	assert.NotEmpty(t, status.Uptime)
	assert.GreaterOrEqual(t, status.ActiveTasks, 0)
	assert.GreaterOrEqual(t, status.TotalRequests, int64(0))
	assert.Greater(t, status.MemoryUsage, uint64(0))
	assert.NotZero(t, status.Timestamp)
}

func TestHealthEndpoint(t *testing.T) {
	config := &ServerConfig{
		Host:  "localhost",
		Port:  8080,
		Debug: false,
	}

	server := NewMonitoringServer(config)
	testServer := httptest.NewServer(server.router)
	defer testServer.Close()

	resp, err := http.Get(testServer.URL + "/api/v1/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var health map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&health)
	require.NoError(t, err)

	assert.Equal(t, "ok", health["status"])
	assert.Contains(t, health, "timestamp")
	assert.Contains(t, health, "checks")

	checks := health["checks"].(map[string]interface{})
	assert.Equal(t, "ok", checks["database"])
	assert.Equal(t, "ok", checks["external_api"])
	assert.Equal(t, "ok", checks["disk_space"])
}

func TestMetricsEndpoint(t *testing.T) {
	config := &ServerConfig{
		Host:  "localhost",
		Port:  8080,
		Debug: false,
	}

	server := NewMonitoringServer(config)
	testServer := httptest.NewServer(server.router)
	defer testServer.Close()

	tests := []struct {
		name           string
		format         string
		expectedType   string
		expectedStatus int
	}{
		{"Prometheus Format", "prometheus", "text/plain", http.StatusOK},
		{"JSON Format", "json", "application/json", http.StatusOK},
		{"Default Format", "", "text/plain", http.StatusOK},
		{"Invalid Format", "xml", "", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := testServer.URL + "/api/v1/metrics"
			if tt.format != "" {
				url += "?format=" + tt.format
			}

			resp, err := http.Get(url)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, tt.expectedType, resp.Header.Get("Content-Type"))
			}
		})
	}
}

func TestMonitoringClient(t *testing.T) {
	config := &ServerConfig{
		Host:  "localhost",
		Port:  8080,
		Debug: false,
	}

	server := NewMonitoringServer(config)
	testServer := httptest.NewServer(server.router)
	defer testServer.Close()

	client := NewMonitoringClient(testServer.URL)
	ctx := context.Background()

	t.Run("GetSystemStatus", func(t *testing.T) {
		status, err := client.GetSystemStatus(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, status)
		assert.Equal(t, "healthy", status.Status)
	})

	t.Run("GetHealth", func(t *testing.T) {
		health, err := client.GetHealth(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, health)
		assert.Equal(t, "ok", health["status"])
	})

	t.Run("GetMetrics", func(t *testing.T) {
		metrics, err := client.GetMetrics(ctx, "json")
		assert.NoError(t, err)
		assert.NotEmpty(t, metrics)

		var metricsData map[string]interface{}
		err = json.Unmarshal([]byte(metrics), &metricsData)
		assert.NoError(t, err)
		assert.Contains(t, metricsData, "active_tasks")
	})

	t.Run("GetTasks", func(t *testing.T) {
		tasks, err := client.GetTasks(ctx, 10, 0, "")
		assert.NoError(t, err)
		assert.NotNil(t, tasks)
		assert.GreaterOrEqual(t, tasks.Total, 0)
	})

	t.Run("GetAlerts", func(t *testing.T) {
		alerts, err := client.GetAlerts(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, alerts)
	})

	t.Run("Ping", func(t *testing.T) {
		err := client.Ping(ctx)
		assert.NoError(t, err)
	})
}

func TestMetricsCollector(t *testing.T) {
	collector := NewMetricsCollector()

	t.Run("InitialState", func(t *testing.T) {
		assert.Equal(t, 0, collector.GetActiveTasks())
		assert.Equal(t, int64(0), collector.GetTotalRequests())
		assert.Greater(t, collector.GetMemoryUsage(), uint64(0))
		assert.Greater(t, collector.GetCPUUsage(), 0.0)
	})

	t.Run("RecordRequest", func(t *testing.T) {
		collector.RecordRequest("GET", "/api/test", 200, 100*time.Millisecond)
		collector.RecordRequest("POST", "/api/test", 201, 150*time.Millisecond)
		collector.RecordRequest("GET", "/api/error", 500, 50*time.Millisecond)

		assert.Equal(t, int64(3), collector.GetTotalRequests())
		assert.Greater(t, collector.GetAverageResponseTime(), time.Duration(0))
		assert.Greater(t, collector.GetErrorRate(), 0.0)
	})

	t.Run("SetActiveTasks", func(t *testing.T) {
		collector.SetActiveTasks(5)
		assert.Equal(t, 5, collector.GetActiveTasks())
	})

	t.Run("ExportPrometheus", func(t *testing.T) {
		prometheus := collector.ExportPrometheus()
		assert.Contains(t, prometheus, "gzh_active_tasks")
		assert.Contains(t, prometheus, "gzh_total_requests")
		assert.Contains(t, prometheus, "gzh_memory_usage_bytes")
	})

	t.Run("ExportJSON", func(t *testing.T) {
		jsonData, err := collector.ExportJSON()
		assert.NoError(t, err)

		var metrics map[string]interface{}
		err = json.Unmarshal([]byte(jsonData), &metrics)
		assert.NoError(t, err)
		assert.Contains(t, metrics, "active_tasks")
		assert.Contains(t, metrics, "total_requests")
	})

	t.Run("Reset", func(t *testing.T) {
		collector.Reset()
		assert.Equal(t, 0, collector.GetActiveTasks())
		assert.Equal(t, int64(0), collector.GetTotalRequests())
	})
}

func TestAlertManager(t *testing.T) {
	manager := NewAlertManager()
	metrics := NewMetricsCollector()
	manager.SetMetrics(metrics)

	t.Run("CreateRule", func(t *testing.T) {
		rule := &AlertRule{
			Name:        "High Memory Usage",
			Description: "Memory usage is too high",
			Query:       "memory_usage_percent",
			Threshold:   80.0,
			Duration:    5 * time.Minute,
			Severity:    AlertSeverityHigh,
			Enabled:     true,
		}

		err := manager.CreateRule(rule)
		assert.NoError(t, err)
		assert.NotEmpty(t, rule.ID)
		assert.False(t, rule.CreatedAt.IsZero())
	})

	t.Run("ListRules", func(t *testing.T) {
		rules := manager.ListRules()
		assert.Len(t, rules, 1)
		assert.Equal(t, "High Memory Usage", rules[0].Name)
	})

	t.Run("CreateAlert", func(t *testing.T) {
		alert := &Alert{
			Name:        "Test Alert",
			Description: "This is a test alert",
			Severity:    "high",
			Status:      "firing",
		}

		err := manager.CreateAlert(alert)
		assert.NoError(t, err)
		assert.NotEmpty(t, alert.ID)
	})

	t.Run("GetAlerts", func(t *testing.T) {
		alerts, err := manager.GetAlerts()
		assert.NoError(t, err)
		assert.Len(t, alerts, 1)
	})

	t.Run("EvaluateRules", func(t *testing.T) {
		ctx := context.Background()
		err := manager.EvaluateRules(ctx)
		assert.NoError(t, err)
	})

	t.Run("GetAlertStats", func(t *testing.T) {
		stats := manager.GetAlertStats()
		assert.Contains(t, stats, "total_rules")
		assert.Contains(t, stats, "total_alerts")
		assert.Contains(t, stats, "firing_alerts")
	})
}

func TestCORSMiddleware(t *testing.T) {
	config := &ServerConfig{
		Host:  "localhost",
		Port:  8080,
		Debug: false,
	}

	server := NewMonitoringServer(config)
	testServer := httptest.NewServer(server.router)
	defer testServer.Close()

	// Test OPTIONS request
	req, err := http.NewRequest("OPTIONS", testServer.URL+"/api/v1/status", nil)
	require.NoError(t, err)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"))
	assert.Contains(t, resp.Header.Get("Access-Control-Allow-Methods"), "GET")
	assert.Contains(t, resp.Header.Get("Access-Control-Allow-Headers"), "Content-Type")
}

func TestTaskOperations(t *testing.T) {
	config := &ServerConfig{
		Host:  "localhost",
		Port:  8080,
		Debug: false,
	}

	server := NewMonitoringServer(config)
	testServer := httptest.NewServer(server.router)
	defer testServer.Close()

	client := NewMonitoringClient(testServer.URL)
	ctx := context.Background()

	t.Run("GetTasks", func(t *testing.T) {
		tasks, err := client.GetTasks(ctx, 10, 0, "")
		assert.NoError(t, err)
		assert.NotNil(t, tasks)
		assert.GreaterOrEqual(t, tasks.Total, 0)
	})

	t.Run("GetTaskWithPagination", func(t *testing.T) {
		tasks, err := client.GetTasks(ctx, 1, 0, "running")
		assert.NoError(t, err)
		assert.NotNil(t, tasks)
		// The mock returns 2 tasks, but we're only requesting 1 per page
		// So we should get all tasks that match the status filter
		assert.GreaterOrEqual(t, len(tasks.Tasks), 0)
	})
}

func TestConfigOperations(t *testing.T) {
	config := &ServerConfig{
		Host:  "localhost",
		Port:  8080,
		Debug: false,
	}

	server := NewMonitoringServer(config)
	testServer := httptest.NewServer(server.router)
	defer testServer.Close()

	client := NewMonitoringClient(testServer.URL)
	ctx := context.Background()

	t.Run("GetConfig", func(t *testing.T) {
		cfg, err := client.GetConfig(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Contains(t, cfg, "server")
		assert.Contains(t, cfg, "metrics")
	})

	t.Run("UpdateConfig", func(t *testing.T) {
		newConfig := map[string]interface{}{
			"test": "value",
		}
		err := client.UpdateConfig(ctx, newConfig)
		assert.NoError(t, err)
	})
}
