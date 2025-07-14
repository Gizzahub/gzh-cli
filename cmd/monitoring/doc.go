// Package monitoring implements comprehensive monitoring and observability features
// for the GZH Manager system.
//
// This package provides real-time monitoring capabilities including:
//   - System performance metrics collection
//   - Application health monitoring
//   - Repository operation tracking
//   - Network environment change detection
//   - Custom metrics and alerting
//   - WebSocket-based real-time dashboards
//
// Key Components:
//
// Metrics Collection:
//   - Prometheus-compatible metrics export
//   - Custom business metrics (clone operations, API calls, errors)
//   - System resource monitoring (CPU, memory, disk)
//   - Application performance metrics
//
// Alerting System:
//   - Multi-channel alert delivery (Slack, Discord, Teams, Email)
//   - Configurable alert rules and thresholds
//   - Alert suppression and escalation
//   - Integration with external monitoring systems
//
// Real-time Dashboard:
//   - WebSocket-based live updates
//   - Interactive performance charts
//   - System status visualization
//   - Alert management interface
//
// Example usage:
//
//	gz monitoring start --port 8080
//	gz monitoring dashboard --web-port 3000
//	gz monitoring metrics --prometheus-port 9090
//
// The monitoring system integrates with popular observability stacks
// including Prometheus, Grafana, and various alerting platforms.
package monitoring
