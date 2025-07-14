// Package debug provides debugging and diagnostic functionality
// for the GZH Manager CLI tool.
//
// This package implements comprehensive debugging capabilities including
// system diagnostics, performance profiling, logging control, and
// troubleshooting utilities for development and production environments.
//
// Key Components:
//
// System Diagnostics:
//   - System health checks and status
//   - Resource usage monitoring
//   - Network connectivity testing
//   - Service dependency validation
//
// Performance Profiling:
//   - CPU and memory profiling
//   - Performance bottleneck detection
//   - Execution time analysis
//   - Resource utilization tracking
//
// Debug Information:
//   - Detailed error reporting
//   - Stack trace analysis
//   - Configuration dump and validation
//   - Runtime state inspection
//
// Troubleshooting Tools:
//   - Interactive debugging sessions
//   - Log analysis and filtering
//   - Configuration verification
//   - Environment validation
//
// Features:
//   - Real-time monitoring dashboards
//   - Debug mode activation
//   - Verbose logging controls
//   - Performance metrics collection
//   - Automated issue detection
//
// Example usage:
//
//	debugger := debug.NewDebugger()
//
//	status := debugger.SystemCheck()
//	profile := debugger.ProfilePerformance()
//
//	err := debugger.EnableVerboseLogging()
//	report := debugger.GenerateReport()
//
// The package provides essential debugging and diagnostic capabilities
// for maintaining and troubleshooting the GZH Manager system.
package debug
