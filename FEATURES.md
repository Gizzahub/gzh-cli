# Features

This document describes the implemented functionality of gzh-manager-go (gz CLI tool).

## Repository Management

### Bulk Repository Cloning
- **Multi-platform Git hosting support**: Clone entire organizations from GitHub, GitLab, Gitea, and Gogs
- **Flexible cloning strategies**: Choose between reset, pull, or fetch strategies for existing repositories
- **Protocol flexibility**: Support for both HTTPS and SSH protocols with automatic authentication
- **Private repository support**: Token-based authentication for accessing private repositories
- **Configuration-driven**: YAML configuration files with environment-specific overrides (home, work, etc.)
- **Kustomize-style configuration**: Layer multiple configuration files for different environments

### SSH Configuration Management
- **Automated SSH config generation**: Create SSH configurations for Git repositories
- **Multi-service support**: Generate configs for GitHub, GitLab, Gitea, and Gogs
- **Key management**: Automatic SSH key association and configuration

## Package Management

### Always-Latest Package Updates
- **Multi-package manager support**: Automated updates for asdf, Homebrew, SDKMAN, MacPorts, APT, and rbenv
- **Flexible update strategies**: 
  - Minor latest: Update to latest minor version within the same major version
  - Major latest: Update to the absolute latest version
- **Bulk package operations**: Update multiple packages and tools simultaneously
- **Cross-platform compatibility**: Works across Linux, macOS, and Windows where applicable

## Development Environment Management

### Configuration Backup and Restore
- **Cloud service configurations**: Save and restore AWS, Google Cloud (gcloud) configurations and credentials
- **Container configurations**: Docker configuration management
- **Kubernetes integration**: kubeconfig backup and restore for cluster management
- **SSH configuration**: Complete SSH config save/load functionality
- **Metadata tracking**: Track save dates, descriptions, and source paths for all configurations
- **Safe operations**: Automatic backups before loading configurations

## Network Environment Management

### System Service Monitoring
- **Comprehensive daemon monitoring**: Monitor and manage system services (daemons) with real-time status updates
- **Network service filtering**: Identify and monitor network-related services specifically
- **Service dependency tracking**: Understand service relationships and dependencies
- **Live monitoring**: Real-time service status updates with configurable intervals
- **Cross-platform support**: Works with systemctl, service managers across different operating systems

### WiFi Network Automation
- **WiFi change detection**: Automatically detect network connections, disconnections, and network switches
- **Event-driven actions**: Trigger customizable actions based on network state changes
- **YAML-based action configuration**: Define network-specific actions using flexible configuration files
- **Daemon mode support**: Run as background service for continuous monitoring
- **Dry-run testing**: Test configurations safely without executing actual commands

### Network Configuration Actions
- **VPN management**: Connect/disconnect VPN connections (OpenVPN, WireGuard, NetworkManager)
- **DNS configuration**: Switch DNS servers based on network environment using resolvectl, NetworkManager
- **Proxy management**: Configure HTTP/HTTPS/SOCKS proxies with environment variables
- **Hosts file management**: Add/remove entries from system hosts file with automatic backups
- **Integrated automation**: Execute network configurations automatically when WiFi networks change
- **Safety features**: Automatic backups, dry-run mode, and validation before making system changes

### Network Environment Transitions
- **Seamless environment switching**: Automatically adapt system configuration when moving between networks (home, office, public WiFi)
- **Profile-based configurations**: Define different network profiles with specific VPN, DNS, proxy, and host settings
- **Event correlation**: Link WiFi network changes to appropriate system configuration changes
- **Rollback capabilities**: Safe configuration changes with automatic backup and restore functionality

## Configuration Management

### YAML Configuration System
- **Hierarchical configurations**: Layer multiple YAML files for different environments and contexts
- **Example configurations**: Built-in templates and examples for all major features
- **Configuration validation**: Syntax checking and validation for all configuration files
- **Environment-specific overrides**: Separate configurations for home, work, and other environments

### CLI Interface
- **Comprehensive help system**: Detailed help documentation for all commands and options
- **Consistent command structure**: Logical command hierarchy across all functionality
- **Rich output formatting**: Color-coded, emoji-enhanced output for better user experience
- **Verbose and dry-run modes**: Detailed logging and safe testing options across all commands

## IDE and Development Tools

### JetBrains IDE Settings Management
- **Cross-platform IDE detection**: Automatic detection of JetBrains products on Linux, macOS, and Windows
- **Real-time settings monitoring**: Track configuration changes across all JetBrains IDE installations using fsnotify
- **Settings synchronization fixes**: Detect and repair common sync issues, particularly with filetypes.xml corruption
- **Multi-IDE support**: Compatible with IntelliJ IDEA, PyCharm, WebStorm, PhpStorm, RubyMine, CLion, GoLand, DataGrip, Android Studio, and Rider
- **Smart file filtering**: Ignore temporary files and focus on meaningful configuration changes
- **Installation discovery**: List all detected JetBrains IDE installations with detailed information
- **Backup and recovery**: Automatic backup creation before applying sync fixes

## Cross-Platform Support
- **Operating system compatibility**: Linux, macOS, and Windows support where applicable
- **Multiple backend support**: Fallback mechanisms for different system tools and package managers
- **Flexible authentication**: Support for various authentication methods across different services