// Package doctor implements system health check and diagnostic commands for the gz CLI.
// It provides comprehensive analysis of the development environment, identifying
// potential issues and suggesting remediation steps.
//
// The doctor command checks:
//   - Git configuration and connectivity
//   - Required tools and their versions
//   - Network connectivity to Git services
//   - Authentication and API tokens
//   - File system permissions
//   - Development environment setup
//   - Container runtime availability
//   - IDE integration status
//
// Each check provides detailed output with:
//   - Current status (OK, WARNING, ERROR)
//   - Diagnostic information
//   - Suggested fixes for identified issues
//   - Links to relevant documentation
package doctor
