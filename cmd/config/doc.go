// Package config provides configuration management commands for the gz CLI.
// It handles loading, validation, and manipulation of configuration files
// used throughout the gzh-manager-go tool suite.
//
// The package supports multiple configuration formats and sources:
//   - YAML configuration files
//   - Environment variables with GZH_ prefix
//   - Command-line flag overrides
//   - Configuration file discovery and precedence
//   - Schema validation for configuration integrity
//
// Configuration precedence (highest to lowest):
//  1. Command-line flags
//  2. Environment variables
//  3. Configuration file in current directory
//  4. User configuration (~/.config/gzh-manager/)
//  5. System configuration (/etc/gzh-manager/)
package config