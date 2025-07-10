package monitoring

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// MonitoringClient is a client for the monitoring API
type MonitoringClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewMonitoringClient creates a new monitoring client
func NewMonitoringClient(baseURL string) *MonitoringClient {
	return &MonitoringClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetSystemStatus gets the system status
func (c *MonitoringClient) GetSystemStatus(ctx context.Context) (*SystemStatus, error) {
	resp, err := c.get(ctx, "/api/v1/status")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", resp.Status)
	}

	var status SystemStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &status, nil
}

// GetHealth gets the health status
func (c *MonitoringClient) GetHealth(ctx context.Context) (map[string]interface{}, error) {
	resp, err := c.get(ctx, "/api/v1/health")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", resp.Status)
	}

	var health map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return health, nil
}

// GetMetrics gets the metrics in the specified format
func (c *MonitoringClient) GetMetrics(ctx context.Context, format string) (string, error) {
	url := "/api/v1/metrics"
	if format != "" {
		url += "?format=" + format
	}

	resp, err := c.get(ctx, url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return string(body), nil
}

// GetTasks gets the list of tasks
func (c *MonitoringClient) GetTasks(ctx context.Context, limit, offset int, status string) (*TaskListResponse, error) {
	url := fmt.Sprintf("/api/v1/tasks?limit=%d&offset=%d", limit, offset)
	if status != "" {
		url += "&status=" + status
	}

	resp, err := c.get(ctx, url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", resp.Status)
	}

	var taskList TaskListResponse
	if err := json.NewDecoder(resp.Body).Decode(&taskList); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &taskList, nil
}

// GetTask gets a specific task by ID
func (c *MonitoringClient) GetTask(ctx context.Context, taskID string) (*Task, error) {
	resp, err := c.get(ctx, "/api/v1/tasks/"+taskID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", resp.Status)
	}

	var task Task
	if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &task, nil
}

// StopTask stops a task by ID
func (c *MonitoringClient) StopTask(ctx context.Context, taskID string) error {
	resp, err := c.post(ctx, "/api/v1/tasks/"+taskID+"/stop", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API error: %s", resp.Status)
	}

	return nil
}

// GetAlerts gets the list of alerts
func (c *MonitoringClient) GetAlerts(ctx context.Context) ([]*AlertInstance, error) {
	resp, err := c.get(ctx, "/api/v1/alerts")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", resp.Status)
	}

	var response struct {
		Alerts []*AlertInstance `json:"alerts"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Alerts, nil
}

// CreateAlert creates a new alert
func (c *MonitoringClient) CreateAlert(ctx context.Context, alert *Alert) error {
	resp, err := c.post(ctx, "/api/v1/alerts", alert)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("API error: %s", resp.Status)
	}

	return nil
}

// UpdateAlert updates an existing alert
func (c *MonitoringClient) UpdateAlert(ctx context.Context, alert *Alert) error {
	resp, err := c.put(ctx, "/api/v1/alerts/"+alert.ID, alert)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API error: %s", resp.Status)
	}

	return nil
}

// DeleteAlert deletes an alert
func (c *MonitoringClient) DeleteAlert(ctx context.Context, alertID string) error {
	resp, err := c.delete(ctx, "/api/v1/alerts/"+alertID)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API error: %s", resp.Status)
	}

	return nil
}

// GetConfig gets the configuration
func (c *MonitoringClient) GetConfig(ctx context.Context) (map[string]interface{}, error) {
	resp, err := c.get(ctx, "/api/v1/config")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", resp.Status)
	}

	var config map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return config, nil
}

// UpdateConfig updates the configuration
func (c *MonitoringClient) UpdateConfig(ctx context.Context, config map[string]interface{}) error {
	resp, err := c.put(ctx, "/api/v1/config", config)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API error: %s", resp.Status)
	}

	return nil
}

// TestNotification sends a test notification
func (c *MonitoringClient) TestNotification(ctx context.Context, notifType, target, message string) error {
	req := map[string]string{
		"type":    notifType,
		"target":  target,
		"message": message,
	}

	resp, err := c.post(ctx, "/api/v1/notifications/test", req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API error: %s", resp.Status)
	}

	return nil
}

// Helper methods for HTTP requests

func (c *MonitoringClient) get(ctx context.Context, path string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}

	return c.httpClient.Do(req)
}

func (c *MonitoringClient) post(ctx context.Context, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+path, reqBody)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.httpClient.Do(req)
}

func (c *MonitoringClient) put(ctx context.Context, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", c.baseURL+path, reqBody)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.httpClient.Do(req)
}

func (c *MonitoringClient) delete(ctx context.Context, path string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "DELETE", c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}

	return c.httpClient.Do(req)
}

// Response types

// TaskListResponse represents the response for task listing
type TaskListResponse struct {
	Tasks  []Task `json:"tasks"`
	Total  int    `json:"total"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
}

// Ping tests the connection to the monitoring server
func (c *MonitoringClient) Ping(ctx context.Context) error {
	resp, err := c.get(ctx, "/api/v1/health")
	if err != nil {
		return fmt.Errorf("failed to connect to monitoring server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("monitoring server returned status: %s", resp.Status)
	}

	return nil
}

// GetMetricsSummary gets a summary of key metrics
func (c *MonitoringClient) GetMetricsSummary(ctx context.Context) (map[string]interface{}, error) {
	metrics, err := c.GetMetrics(ctx, "json")
	if err != nil {
		return nil, err
	}

	var metricsData map[string]interface{}
	if err := json.Unmarshal([]byte(metrics), &metricsData); err != nil {
		return nil, fmt.Errorf("failed to parse metrics: %w", err)
	}

	// Extract key metrics for summary
	summary := map[string]interface{}{
		"active_tasks":    metricsData["active_tasks"],
		"total_requests":  metricsData["total_requests"],
		"memory_usage_mb": float64(metricsData["memory_usage_bytes"].(float64)) / 1024 / 1024,
		"cpu_usage":       metricsData["cpu_usage_percent"],
		"uptime":          fmt.Sprintf("%.0f seconds", metricsData["uptime_seconds"]),
	}

	return summary, nil
}
