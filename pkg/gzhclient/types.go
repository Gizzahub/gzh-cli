// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package gzhclient

import (
	"fmt"
	"time"
)

// HealthStatus represents the overall health of the client.
type HealthStatus struct {
	Overall    StatusType                 `json:"overall"`
	Components map[string]ComponentHealth `json:"components"`
	Timestamp  time.Time                  `json:"timestamp"`
}

// ComponentHealth represents the health of a specific component.
type ComponentHealth struct {
	Status  StatusType             `json:"status"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// StatusType represents health status.
type StatusType string

const (
	// StatusHealthy indicates the service is operating normally.
	StatusHealthy StatusType = "healthy"
	// StatusDegraded indicates the service is operational but with reduced performance.
	StatusDegraded StatusType = "degraded"
	// StatusUnhealthy indicates the service is not operational.
	StatusUnhealthy StatusType = "unhealthy"
)

// BulkCloneRequest represents a bulk clone operation request.
type BulkCloneRequest struct {
	Platforms      []PlatformConfig `json:"platforms" yaml:"platforms"`
	OutputDir      string           `json:"output_dir" yaml:"outputDir"`
	Concurrency    int              `json:"concurrency" yaml:"concurrency"`
	Strategy       string           `json:"strategy" yaml:"strategy"` // reset, pull, fetch
	IncludePrivate bool             `json:"include_private" yaml:"includePrivate"`
	Filters        CloneFilters     `json:"filters" yaml:"filters"`
}

// PlatformConfig defines configuration for a Git platform.
type PlatformConfig struct {
	Type          string   `json:"type" yaml:"type"` // github, gitlab, gitea, gogs
	URL           string   `json:"url" yaml:"url"`
	Token         string   `json:"token" yaml:"token"`
	Organizations []string `json:"organizations" yaml:"organizations"`
	Users         []string `json:"users" yaml:"users"`
}

// CloneFilters defines filtering options for repositories.
type CloneFilters struct {
	IncludeRepos []string  `json:"include_repos" yaml:"includeRepos"`
	ExcludeRepos []string  `json:"exclude_repos" yaml:"excludeRepos"`
	Languages    []string  `json:"languages" yaml:"languages"`
	MinSize      int64     `json:"min_size" yaml:"minSize"`
	MaxSize      int64     `json:"max_size" yaml:"maxSize"`
	UpdatedAfter time.Time `json:"updated_after" yaml:"updatedAfter"`
}

// BulkCloneResult represents the result of a bulk clone operation.
type BulkCloneResult struct {
	TotalRepos   int                     `json:"total_repos"`
	SuccessCount int                     `json:"success_count"`
	FailureCount int                     `json:"failure_count"`
	SkippedCount int                     `json:"skipped_count"`
	Results      []RepositoryCloneResult `json:"results"`
	Duration     time.Duration           `json:"duration"`
	Summary      map[string]interface{}  `json:"summary"`
}

// RepositoryCloneResult represents the result of cloning a single repository.
type RepositoryCloneResult struct {
	RepoName  string        `json:"repo_name"`
	Platform  string        `json:"platform"`
	URL       string        `json:"url"`
	LocalPath string        `json:"local_path"`
	Status    string        `json:"status"` // success, failed, skipped
	Error     string        `json:"error,omitempty"`
	Duration  time.Duration `json:"duration"`
	Size      int64         `json:"size"`
}

// DevEnvRequest represents a development environment setup request.
type DevEnvRequest struct {
	Profile     string            `json:"profile" yaml:"profile"`
	Services    []ServiceConfig   `json:"services" yaml:"services"`
	Environment map[string]string `json:"environment" yaml:"environment"`
	Volumes     []VolumeMount     `json:"volumes" yaml:"volumes"`
	Networks    []NetworkConfig   `json:"networks" yaml:"networks"`
}

// ServiceConfig defines a service in the development environment.
type ServiceConfig struct {
	Name    string            `json:"name" yaml:"name"`
	Image   string            `json:"image" yaml:"image"`
	Ports   []PortMapping     `json:"ports" yaml:"ports"`
	Env     map[string]string `json:"env" yaml:"env"`
	Volumes []string          `json:"volumes" yaml:"volumes"`
	Command []string          `json:"command" yaml:"command"`
}

// PortMapping defines port forwarding configuration.
type PortMapping struct {
	Host      int    `json:"host" yaml:"host"`
	Container int    `json:"container" yaml:"container"`
	Protocol  string `json:"protocol" yaml:"protocol"`
}

// VolumeMount defines volume mounting configuration.
type VolumeMount struct {
	Source string `json:"source" yaml:"source"`
	Target string `json:"target" yaml:"target"`
	Type   string `json:"type" yaml:"type"` // bind, volume, tmpfs
}

// NetworkConfig defines network configuration.
type NetworkConfig struct {
	Name   string `json:"name" yaml:"name"`
	Driver string `json:"driver" yaml:"driver"`
	IPAM   IPAM   `json:"ipam" yaml:"ipam"`
}

// IPAM defines IP address management configuration.
type IPAM struct {
	Driver string       `json:"driver" yaml:"driver"`
	Config []IPAMConfig `json:"config" yaml:"config"`
}

// IPAMConfig defines IPAM configuration.
type IPAMConfig struct {
	Subnet  string `json:"subnet" yaml:"subnet"`
	Gateway string `json:"gateway" yaml:"gateway"`
}

// DevEnvResult represents the result of development environment setup.
type DevEnvResult struct {
	Profile   string            `json:"profile"`
	Status    string            `json:"status"`
	Services  []ServiceResult   `json:"services"`
	Networks  []NetworkResult   `json:"networks"`
	Duration  time.Duration     `json:"duration"`
	Endpoints map[string]string `json:"endpoints"`
}

// ServiceResult represents the result of service setup.
type ServiceResult struct {
	Name      string        `json:"name"`
	Status    string        `json:"status"`
	Error     string        `json:"error,omitempty"`
	Container string        `json:"container"`
	Ports     []PortMapping `json:"ports"`
	Duration  time.Duration `json:"duration"`
}

// NetworkResult represents the result of network setup.
type NetworkResult struct {
	Name     string        `json:"name"`
	Status   string        `json:"status"`
	Error    string        `json:"error,omitempty"`
	ID       string        `json:"id"`
	Duration time.Duration `json:"duration"`
}

// MonitoringData represents monitoring information.
type MonitoringData struct {
	System   SystemMetrics  `json:"system"`
	Services ServiceMetrics `json:"services"`
	Network  NetworkMetrics `json:"network"`
	Tasks    []TaskStatus   `json:"tasks"`
}

// SystemMetrics represents system-level metrics.
type SystemMetrics struct {
	CPU       CPUMetrics    `json:"cpu"`
	Memory    MemoryMetrics `json:"memory"`
	Disk      DiskMetrics   `json:"disk"`
	LoadAvg   []float64     `json:"load_avg"`
	Uptime    time.Duration `json:"uptime"`
	Timestamp time.Time     `json:"timestamp"`
}

// CPUMetrics represents CPU metrics.
type CPUMetrics struct {
	Usage      float64 `json:"usage"`
	Cores      int     `json:"cores"`
	UserTime   float64 `json:"user_time"`
	SystemTime float64 `json:"system_time"`
	IdleTime   float64 `json:"idle_time"`
}

// MemoryMetrics represents memory metrics.
type MemoryMetrics struct {
	Total     uint64  `json:"total"`
	Used      uint64  `json:"used"`
	Available uint64  `json:"available"`
	Usage     float64 `json:"usage"`
	Cached    uint64  `json:"cached"`
	Buffers   uint64  `json:"buffers"`
}

// DiskMetrics represents disk metrics.
type DiskMetrics struct {
	Total      uint64  `json:"total"`
	Used       uint64  `json:"used"`
	Available  uint64  `json:"available"`
	Usage      float64 `json:"usage"`
	ReadOps    uint64  `json:"read_ops"`
	WriteOps   uint64  `json:"write_ops"`
	ReadBytes  uint64  `json:"read_bytes"`
	WriteBytes uint64  `json:"write_bytes"`
}

// ServiceMetrics represents service-level metrics.
type ServiceMetrics struct {
	Services []ServiceInfo `json:"services"`
	Total    int           `json:"total"`
	Running  int           `json:"running"`
	Stopped  int           `json:"stopped"`
	Failed   int           `json:"failed"`
}

// ServiceInfo represents information about a service.
type ServiceInfo struct {
	Name     string            `json:"name"`
	Status   string            `json:"status"`
	Image    string            `json:"image"`
	Ports    []PortMapping     `json:"ports"`
	CPU      float64           `json:"cpu"`
	Memory   uint64            `json:"memory"`
	Networks []string          `json:"networks"`
	Labels   map[string]string `json:"labels"`
}

// NetworkMetrics represents network metrics.
type NetworkMetrics struct {
	Interfaces []NetworkInterface `json:"interfaces"`
	TotalRx    uint64             `json:"total_rx"`
	TotalTx    uint64             `json:"total_tx"`
}

// NetworkInterface represents network interface information.
type NetworkInterface struct {
	Name      string   `json:"name"`
	Addresses []string `json:"addresses"`
	RxBytes   uint64   `json:"rx_bytes"`
	TxBytes   uint64   `json:"tx_bytes"`
	RxPackets uint64   `json:"rx_packets"`
	TxPackets uint64   `json:"tx_packets"`
	Status    string   `json:"status"`
}

// TaskStatus represents the status of a running task.
type TaskStatus struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Status    string                 `json:"status"`
	Progress  float64                `json:"progress"`
	StartTime time.Time              `json:"start_time"`
	Duration  time.Duration          `json:"duration"`
	Result    map[string]interface{} `json:"result,omitempty"`
	Error     string                 `json:"error,omitempty"`
}

// PluginInfo represents information about a plugin - DISABLED (plugins package removed)
// type PluginInfo struct {
//	Name         string    `json:"name"`
//	Version      string    `json:"version"`
//	Description  string    `json:"description"`
//	Author       string    `json:"author"`
//	Status       string    `json:"status"`
//	Capabilities []string  `json:"capabilities"`
//	LoadTime     time.Time `json:"load_time"`
//	LastUsed     time.Time `json:"last_used"`
//	CallCount    int64     `json:"call_count"`
//	ErrorCount   int64     `json:"error_count"`
// }

// PluginExecuteRequest represents a plugin execution request - DISABLED (plugins package removed)
// type PluginExecuteRequest struct {
//	PluginName string                 `json:"plugin_name"`
//	Method     string                 `json:"method,omitempty"`
//	Args       map[string]interface{} `json:"args"`
//	Timeout    time.Duration          `json:"timeout,omitempty"`
// }

// PluginExecuteResult represents the result of plugin execution - DISABLED (plugins package removed)
// type PluginExecuteResult struct {
//	PluginName string        `json:"plugin_name"`
//	Method     string        `json:"method"`
//	Result     interface{}   `json:"result"`
//	Error      string        `json:"error,omitempty"`
//	Duration   time.Duration `json:"duration"`
//	Timestamp  time.Time     `json:"timestamp"`
// }

// Event represents a system event.
type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
	Level     string                 `json:"level"` // info, warn, error
}

// EventSubscription represents an event subscription.
type EventSubscription struct {
	ID      string                 `json:"id"`
	Types   []string               `json:"types"`
	Sources []string               `json:"sources"`
	Filters map[string]interface{} `json:"filters"`
	Webhook string                 `json:"webhook,omitempty"`
	Active  bool                   `json:"active"`
	Created time.Time              `json:"created"`
}

// APIError represents an API error response.
type APIError struct {
	Code      string                 `json:"code"`
	Message   string                 `json:"message"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	RequestID string                 `json:"request_id,omitempty"`
}

// Error implements the error interface.
func (e APIError) Error() string {
	return fmt.Sprintf("API Error %s: %s", e.Code, e.Message)
}

// PaginationInfo represents pagination information.
type PaginationInfo struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// ListOptions represents options for list operations.
type ListOptions struct {
	Page    int               `json:"page,omitempty"`
	PerPage int               `json:"per_page,omitempty"`
	Sort    string            `json:"sort,omitempty"`
	Order   string            `json:"order,omitempty"` // asc, desc
	Filters map[string]string `json:"filters,omitempty"`
}

// Response represents a paginated API response.
type Response struct {
	Data       interface{}            `json:"data"`
	Pagination *PaginationInfo        `json:"pagination,omitempty"`
	Meta       map[string]interface{} `json:"meta,omitempty"`
}
