package monitoring

import "time"

// GetDefaultDashboardTemplates returns default dashboard templates for gzh-manager
func GetDefaultDashboardTemplates() []DashboardTemplate {
	return []DashboardTemplate{
		{
			Name:        "gzh-manager-overview",
			Title:       "GZH Manager - System Overview",
			Description: "Overview of gzh-manager system metrics and performance",
			Tags:        []string{"gzh-manager", "system", "overview"},
			Variables: []VariableTemplate{
				{
					Name:       "instance",
					Type:       "query",
					Label:      "Instance",
					Query:      "label_values(gzh_manager_monitoring_tasks_total, instance)",
					Multi:      true,
					IncludeAll: true,
				},
				{
					Name:       "organization",
					Type:       "query",
					Label:      "Organization",
					Query:      "label_values(gzh_manager_monitoring_tasks_total, organization)",
					Multi:      true,
					IncludeAll: true,
				},
			},
			Panels: []PanelTemplate{
				{
					Title: "Task Execution Rate",
					Type:  "stat",
					GridPos: GridPosition{
						H: 8, W: 6, X: 0, Y: 0,
					},
					Queries: []string{
						"sum(rate(gzh_manager_monitoring_tasks_total{instance=~\"$instance\"}[5m]))",
					},
					Unit: "ops",
					Thresholds: []ThresholdTemplate{
						{Color: "green", Value: 0},
						{Color: "yellow", Value: 10},
						{Color: "red", Value: 50},
					},
				},
				{
					Title: "Active Tasks by Type",
					Type:  "piechart",
					GridPos: GridPosition{
						H: 8, W: 6, X: 6, Y: 0,
					},
					Queries: []string{
						"sum by (type) (gzh_manager_monitoring_tasks_total{instance=~\"$instance\"})",
					},
				},
				{
					Title: "System CPU Usage",
					Type:  "timeseries",
					GridPos: GridPosition{
						H: 8, W: 12, X: 0, Y: 8,
					},
					Queries: []string{
						"gzh_manager_monitoring_cpu_usage_percent{instance=~\"$instance\"}",
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
						H: 8, W: 12, X: 12, Y: 8,
					},
					Queries: []string{
						"gzh_manager_monitoring_memory_usage_bytes{instance=~\"$instance\"} / 1024 / 1024",
					},
					Unit: "MB",
				},
				{
					Title: "Task Duration Distribution",
					Type:  "heatmap",
					GridPos: GridPosition{
						H: 8, W: 24, X: 0, Y: 16,
					},
					Queries: []string{
						"sum(rate(gzh_manager_monitoring_task_duration_seconds_bucket{instance=~\"$instance\"}[5m])) by (le)",
					},
				},
				{
					Title: "HTTP Request Rate",
					Type:  "timeseries",
					GridPos: GridPosition{
						H: 8, W: 12, X: 0, Y: 24,
					},
					Queries: []string{
						"sum(rate(gzh_manager_monitoring_http_requests_total{instance=~\"$instance\"}[5m])) by (endpoint)",
					},
					Unit: "reqps",
				},
				{
					Title: "Alert Rate by Severity",
					Type:  "timeseries",
					GridPos: GridPosition{
						H: 8, W: 12, X: 12, Y: 24,
					},
					Queries: []string{
						"sum(rate(gzh_manager_monitoring_alerts_total{instance=~\"$instance\"}[5m])) by (severity)",
					},
					Unit: "alerts/s",
				},
			},
		},
		{
			Name:        "gzh-manager-performance",
			Title:       "GZH Manager - Performance Metrics",
			Description: "Detailed performance metrics for gzh-manager components",
			Tags:        []string{"gzh-manager", "performance", "monitoring"},
			Variables: []VariableTemplate{
				{
					Name:       "instance",
					Type:       "query",
					Label:      "Instance",
					Query:      "label_values(gzh_manager_monitoring_tasks_total, instance)",
					Multi:      false,
					IncludeAll: false,
				},
				{
					Name:    "interval",
					Type:    "interval",
					Label:   "Interval",
					Options: []string{"30s", "1m", "5m", "15m", "30m", "1h"},
				},
			},
			Panels: []PanelTemplate{
				{
					Title: "Task Execution Time (95th percentile)",
					Type:  "timeseries",
					GridPos: GridPosition{
						H: 8, W: 12, X: 0, Y: 0,
					},
					Queries: []string{
						"histogram_quantile(0.95, sum(rate(gzh_manager_monitoring_task_duration_seconds_bucket{instance=\"$instance\"}[$interval])) by (le, type))",
					},
					Unit: "s",
				},
				{
					Title: "HTTP Response Time (99th percentile)",
					Type:  "timeseries",
					GridPos: GridPosition{
						H: 8, W: 12, X: 12, Y: 0,
					},
					Queries: []string{
						"histogram_quantile(0.99, sum(rate(gzh_manager_monitoring_http_request_duration_seconds_bucket{instance=\"$instance\"}[$interval])) by (le, endpoint))",
					},
					Unit: "s",
				},
				{
					Title: "Rule Evaluation Success Rate",
					Type:  "stat",
					GridPos: GridPosition{
						H: 8, W: 8, X: 0, Y: 8,
					},
					Queries: []string{
						"sum(rate(gzh_manager_monitoring_rule_evaluations_total{instance=\"$instance\",result=\"true\"}[$interval])) / sum(rate(gzh_manager_monitoring_rule_evaluations_total{instance=\"$instance\"}[$interval])) * 100",
					},
					Unit: "percent",
					Thresholds: []ThresholdTemplate{
						{Color: "red", Value: 0},
						{Color: "yellow", Value: 95},
						{Color: "green", Value: 99},
					},
				},
				{
					Title: "Error Rate by Type",
					Type:  "timeseries",
					GridPos: GridPosition{
						H: 8, W: 16, X: 8, Y: 8,
					},
					Queries: []string{
						"sum(rate(gzh_manager_monitoring_tasks_total{instance=\"$instance\",status=\"error\"}[$interval])) by (type)",
					},
					Unit: "errors/s",
				},
			},
		},
		{
			Name:        "gzh-manager-alerts",
			Title:       "GZH Manager - Alert Dashboard",
			Description: "Alert monitoring and analysis for gzh-manager",
			Tags:        []string{"gzh-manager", "alerts", "monitoring"},
			Variables: []VariableTemplate{
				{
					Name:       "severity",
					Type:       "query",
					Label:      "Severity",
					Query:      "label_values(gzh_manager_monitoring_alerts_total, severity)",
					Multi:      true,
					IncludeAll: true,
				},
				{
					Name:       "rule_id",
					Type:       "query",
					Label:      "Rule ID",
					Query:      "label_values(gzh_manager_monitoring_alerts_total, rule_id)",
					Multi:      true,
					IncludeAll: true,
				},
			},
			Panels: []PanelTemplate{
				{
					Title: "Active Alerts by Severity",
					Type:  "stat",
					GridPos: GridPosition{
						H: 8, W: 24, X: 0, Y: 0,
					},
					Queries: []string{
						"sum by (severity) (gzh_manager_monitoring_alerts_total{severity=~\"$severity\",status=\"firing\"})",
					},
				},
				{
					Title: "Alert Rate Over Time",
					Type:  "timeseries",
					GridPos: GridPosition{
						H: 8, W: 12, X: 0, Y: 8,
					},
					Queries: []string{
						"sum(rate(gzh_manager_monitoring_alerts_total{severity=~\"$severity\"}[5m])) by (severity)",
					},
					Unit: "alerts/s",
				},
				{
					Title: "Top Alerting Rules",
					Type:  "table",
					GridPos: GridPosition{
						H: 8, W: 12, X: 12, Y: 8,
					},
					Queries: []string{
						"topk(10, sum by (rule_id) (increase(gzh_manager_monitoring_alerts_total{rule_id=~\"$rule_id\"}[1h])))",
					},
				},
				{
					Title: "Alert Resolution Time",
					Type:  "histogram",
					GridPos: GridPosition{
						H: 8, W: 24, X: 0, Y: 16,
					},
					Queries: []string{
						"histogram_quantile(0.95, sum(rate(alert_resolution_duration_seconds_bucket[5m])) by (le))",
					},
					Unit: "s",
				},
			},
		},
	}
}

// GetDefaultAlertRuleTemplates returns default alert rule templates for gzh-manager
func GetDefaultAlertRuleTemplates() []AlertRuleTemplate {
	return []AlertRuleTemplate{
		{
			Title:     "High CPU Usage",
			Condition: "A",
			Queries: []string{
				"gzh_manager_monitoring_cpu_usage_percent > 90",
			},
			For: 5 * time.Minute,
			Annotations: map[string]string{
				"summary":     "GZH Manager instance has high CPU usage",
				"description": "CPU usage is above 90% for more than 5 minutes on instance {{ $labels.instance }}",
				"runbook":     "https://wiki.company.com/runbooks/high-cpu-usage",
			},
			Labels: map[string]string{
				"severity":   "warning",
				"team":       "infrastructure",
				"component":  "gzh-manager",
				"alert_type": "system",
			},
		},
		{
			Title:     "High Memory Usage",
			Condition: "A",
			Queries: []string{
				"gzh_manager_monitoring_memory_usage_bytes / 1024 / 1024 / 1024 > 8",
			},
			For: 5 * time.Minute,
			Annotations: map[string]string{
				"summary":     "GZH Manager instance has high memory usage",
				"description": "Memory usage is above 8GB for more than 5 minutes on instance {{ $labels.instance }}",
				"runbook":     "https://wiki.company.com/runbooks/high-memory-usage",
			},
			Labels: map[string]string{
				"severity":   "warning",
				"team":       "infrastructure",
				"component":  "gzh-manager",
				"alert_type": "system",
			},
		},
		{
			Title:     "Task Execution Failure Rate High",
			Condition: "A",
			Queries: []string{
				"sum(rate(gzh_manager_monitoring_tasks_total{status=\"error\"}[5m])) / sum(rate(gzh_manager_monitoring_tasks_total[5m])) * 100 > 10",
			},
			For: 2 * time.Minute,
			Annotations: map[string]string{
				"summary":     "High task execution failure rate",
				"description": "Task execution failure rate is above 10% for more than 2 minutes",
				"runbook":     "https://wiki.company.com/runbooks/task-failures",
			},
			Labels: map[string]string{
				"severity":   "critical",
				"team":       "platform",
				"component":  "gzh-manager",
				"alert_type": "application",
			},
		},
		{
			Title:     "HTTP Error Rate High",
			Condition: "A",
			Queries: []string{
				"sum(rate(gzh_manager_monitoring_http_requests_total{status=~\"5..\"}[5m])) / sum(rate(gzh_manager_monitoring_http_requests_total[5m])) * 100 > 5",
			},
			For: 3 * time.Minute,
			Annotations: map[string]string{
				"summary":     "High HTTP error rate",
				"description": "HTTP 5xx error rate is above 5% for more than 3 minutes",
				"runbook":     "https://wiki.company.com/runbooks/http-errors",
			},
			Labels: map[string]string{
				"severity":   "warning",
				"team":       "platform",
				"component":  "gzh-manager",
				"alert_type": "application",
			},
		},
		{
			Title:     "Alert Manager Down",
			Condition: "A",
			Queries: []string{
				"up{job=\"gzh-manager\"} == 0",
			},
			For: 1 * time.Minute,
			Annotations: map[string]string{
				"summary":     "GZH Manager instance is down",
				"description": "GZH Manager instance {{ $labels.instance }} has been down for more than 1 minute",
				"runbook":     "https://wiki.company.com/runbooks/service-down",
			},
			Labels: map[string]string{
				"severity":   "critical",
				"team":       "infrastructure",
				"component":  "gzh-manager",
				"alert_type": "availability",
			},
		},
		{
			Title:     "Slow Task Execution",
			Condition: "A",
			Queries: []string{
				"histogram_quantile(0.95, sum(rate(gzh_manager_monitoring_task_duration_seconds_bucket[5m])) by (le, type)) > 300",
			},
			For: 10 * time.Minute,
			Annotations: map[string]string{
				"summary":     "Task execution is slow",
				"description": "95th percentile of task execution duration is above 5 minutes for task type {{ $labels.type }}",
				"runbook":     "https://wiki.company.com/runbooks/slow-tasks",
			},
			Labels: map[string]string{
				"severity":   "warning",
				"team":       "platform",
				"component":  "gzh-manager",
				"alert_type": "performance",
			},
		},
		{
			Title:     "Rule Evaluation Failure",
			Condition: "A",
			Queries: []string{
				"sum(rate(gzh_manager_monitoring_rule_evaluations_total{result=\"false\"}[5m])) / sum(rate(gzh_manager_monitoring_rule_evaluations_total[5m])) * 100 > 20",
			},
			For: 5 * time.Minute,
			Annotations: map[string]string{
				"summary":     "High rule evaluation failure rate",
				"description": "Rule evaluation failure rate is above 20% for more than 5 minutes",
				"runbook":     "https://wiki.company.com/runbooks/rule-evaluation-failures",
			},
			Labels: map[string]string{
				"severity":   "warning",
				"team":       "platform",
				"component":  "gzh-manager",
				"alert_type": "monitoring",
			},
		},
		{
			Title:     "No Recent Tasks",
			Condition: "A",
			Queries: []string{
				"sum(increase(gzh_manager_monitoring_tasks_total[1h])) == 0",
			},
			For: 30 * time.Minute,
			Annotations: map[string]string{
				"summary":     "No tasks executed recently",
				"description": "No tasks have been executed in the last hour, which may indicate a problem",
				"runbook":     "https://wiki.company.com/runbooks/no-activity",
			},
			Labels: map[string]string{
				"severity":   "warning",
				"team":       "platform",
				"component":  "gzh-manager",
				"alert_type": "activity",
			},
		},
	}
}

// GetGrafanaConfigExample returns an example Grafana configuration
func GetGrafanaConfigExample() *GrafanaConfig {
	return &GrafanaConfig{
		Enabled: true,
		BaseURL: "http://localhost:3000",
		APIKey:  "your-grafana-api-key-here",
		OrgID:   1,
		Timeout: 30 * time.Second,
		Variables: map[string]interface{}{
			"instance":     "localhost:8080",
			"environment":  "production",
			"datacenter":   "us-west-2",
			"cluster_name": "gzh-manager-prod",
		},
		Tags:       []string{"gzh-manager", "monitoring", "automated"},
		Dashboards: GetDefaultDashboardTemplates(),
		AlertRules: GetDefaultAlertRuleTemplates(),
	}
}
