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
	username   string
	password   string
	token      string
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

// SetAuth sets the authentication credentials and obtains a token
func (c *MonitoringClient) SetAuth(username, password string) error {
	c.username = username
	c.password = password

	// Login and get token
	credentials := map[string]string{
		"username": username,
		"password": password,
	}

	jsonData, err := json.Marshal(credentials)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.baseURL+"/auth/login", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed: %s", resp.Status)
	}

	var loginResponse struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&loginResponse); err != nil {
		return err
	}

	c.token = loginResponse.Token
	return nil
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

	// Add token authorization if token is set
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
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

	// Add token authorization if token is set
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
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

	// Add token authorization if token is set
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	return c.httpClient.Do(req)
}

func (c *MonitoringClient) delete(ctx context.Context, path string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "DELETE", c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}

	// Add token authorization if token is set
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
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

// InstanceInfo represents monitoring instance information
type InstanceInfo struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Host         string            `json:"host"`
	Port         int               `json:"port"`
	Status       string            `json:"status"`
	Version      string            `json:"version"`
	LastSeen     time.Time         `json:"last_seen"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	Capabilities []string          `json:"capabilities,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

// ClusterStatus represents the status of the monitoring cluster
type ClusterStatus struct {
	ID              string          `json:"id"`
	Name            string          `json:"name"`
	Status          string          `json:"status"`
	TotalInstances  int             `json:"total_instances"`
	ActiveInstances int             `json:"active_instances"`
	Instances       []InstanceInfo  `json:"instances"`
	Leader          *InstanceInfo   `json:"leader,omitempty"`
	LastUpdated     time.Time       `json:"last_updated"`
	HealthChecks    map[string]bool `json:"health_checks,omitempty"`
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

// Instance management methods

// GetInstances gets the list of monitoring instances
func (c *MonitoringClient) GetInstances(ctx context.Context) ([]*InstanceInfo, error) {
	resp, err := c.get(ctx, "/api/v1/instances")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", resp.Status)
	}

	var response struct {
		Instances []*InstanceInfo `json:"instances"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Instances, nil
}

// GetInstance gets a specific instance by ID
func (c *MonitoringClient) GetInstance(ctx context.Context, instanceID string) (*InstanceInfo, error) {
	resp, err := c.get(ctx, "/api/v1/instances/"+instanceID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("instance not found: %s", instanceID)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", resp.Status)
	}

	var response struct {
		Instance *InstanceInfo `json:"instance"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Instance, nil
}

// DiscoverInstance discovers and adds a remote monitoring instance
func (c *MonitoringClient) DiscoverInstance(ctx context.Context, host string, port int) error {
	req := map[string]interface{}{
		"host": host,
		"port": port,
	}

	resp, err := c.post(ctx, "/api/v1/instances/discover", req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("API error: %s", resp.Status)
	}

	return nil
}

// RemoveInstance removes an instance from the registry
func (c *MonitoringClient) RemoveInstance(ctx context.Context, instanceID string) error {
	resp, err := c.delete(ctx, "/api/v1/instances/"+instanceID)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API error: %s", resp.Status)
	}

	return nil
}

// GetClusterStatus gets the cluster status
func (c *MonitoringClient) GetClusterStatus(ctx context.Context) (*ClusterStatus, error) {
	resp, err := c.get(ctx, "/api/v1/instances/cluster/status")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", resp.Status)
	}

	var status ClusterStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &status, nil
}
