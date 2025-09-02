# Network Environment Management Specification

## Overview

The `net-env` command provides network environment management capabilities, enabling automatic detection of network changes and executing appropriate configuration adjustments. It manages network configuration transitions, VPN connections, DNS settings, proxy configurations, and basic network operations based on network context.

## Commands

### Core Commands (Currently Implemented)

- `gz net-env` - Launch interactive TUI dashboard
- `gz net-env status` - Show comprehensive network status
- `gz net-env profile` - Manage network profiles
- `gz net-env actions` - Execute network configuration actions
- `gz net-env cloud` - Cloud provider network management

### Interactive TUI Dashboard (`gz net-env`)

**Purpose**: Provides a visual interface for managing network configurations

**Features**:

- Real-time network status display
- Basic network information display
- Keyboard shortcuts for common actions
- Network component overview

**TUI Layout**:

```
┌─ GZH Network Environment Manager ─────────────────────────────────────────┐
│ Current Network: Corporate WiFi                 Status: Connected          │
├─────────────────────────────────────────────────────────────────────────┤
│ Component    │ Status      │ Details                    │ Health         │
├──────────────┼─────────────┼────────────────────────────┼────────────────┤
│ WiFi         │ ● Connected │ Corporate WiFi (5GHz)      │ Excellent     │
│ DNS          │ ● Active    │ 1.1.1.1, 8.8.8.8          │ Responsive    │
│ Proxy        │ ○ Disabled  │ -                          │ -             │
├─────────────────────────────────────────────────────────────────────────┤
│ Actions: [r]efresh [s]tatus [?]help [Q]uit                             │
└─────────────────────────────────────────────────────────────────────────┘
```

**Usage**:

```bash
gz net-env                    # Launch TUI dashboard
```

### Network Status (`gz net-env status`)

**Purpose**: Display comprehensive network configuration and status

**Features**:

- Unified status view of network components
- Network interface and WiFi details
- Basic network connectivity information
- System network configuration display

**Usage**:

```bash
gz net-env status                  # Show current network status
gz net-env status --verbose       # Show detailed network information
gz net-env status --json          # Output in JSON format
```

**Output Example**:

```
Network Environment Status
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Network: Corporate WiFi (5GHz, -45 dBm)
Interface: wlan0 (192.168.1.100/24)
Gateway: 192.168.1.1

Components:
  WiFi      ✓ Connected     Corporate WiFi
  DNS       ✓ Configured    1.1.1.1, 8.8.8.8
  Proxy     ○ Disabled      -

Network Health: Good
```

### Profile Management (`gz net-env profile`)

**Purpose**: Basic network profile management

**Features**:

- List available profiles
- Basic profile operations
- Profile configuration viewing

**Usage**:

```bash
gz net-env profile list           # List all profiles
gz net-env profile show <name>    # Show profile details
```

### Network Actions (`gz net-env actions`)

**Purpose**: Execute predefined network configuration actions

**Features**:

- VPN connection management (basic)
- DNS server configuration
- Proxy settings management
- Network configuration actions

**Usage**:

```bash
gz net-env actions run            # Execute network actions
gz net-env actions config init    # Create example configuration
gz net-env actions vpn connect --name <vpn-name>   # Connect to VPN
gz net-env actions dns set --servers 1.1.1.1,1.0.0.1 # Set DNS servers
gz net-env actions proxy set --http proxy.company.com:8080 # Set proxy
```

### Cloud Provider Management (`gz net-env cloud`)

**Purpose**: Manage cloud provider network configurations

**Features**:

- Cloud provider network settings
- Cloud-specific network configurations
- Multi-cloud network management

**Usage**:

```bash
gz net-env cloud list             # List cloud providers
gz net-env cloud configure <provider> # Configure cloud network settings
```

## Configuration

### Global Configuration

Network environment configurations are stored in:

- `~/.config/gzh-manager/net-env.yaml` - User-specific settings
- `/etc/gzh-manager/net-env.yaml` - System-wide settings
- Environment variable: `GZH_NET_ENV_CONFIG`

### Configuration Structure

```yaml
# Network Environment Configuration
version: "1.0.0"

net_env:
  # Network Profiles
  profiles:
    office:
      name: "Office Network"
      description: "Corporate office network configuration"
      dns:
        servers:
          - "10.0.0.1"
          - "10.0.0.2"
      proxy:
        http: "proxy.company.com:8080"
        https: "proxy.company.com:8080"
        no_proxy: "localhost,127.0.0.1,*.company.com"

    home:
      name: "Home Network"
      description: "Home network configuration"
      dns:
        servers:
          - "192.168.1.1"
          - "8.8.8.8"
      proxy:
        enabled: false

  # Network Actions
  actions:
    dns_set_office:
      type: "dns"
      action: "set"
      servers:
        - "10.0.0.1"
        - "10.0.0.2"

    proxy_enable_corporate:
      type: "proxy"
      action: "set"
      http: "proxy.company.com:8080"
      https: "proxy.company.com:8080"
      no_proxy: "localhost,127.0.0.1,*.company.com"
```

### Environment Variables

- `GZH_NET_ENV_CONFIG` - Path to configuration file
- `GZH_NET_ENV_PROFILE` - Override default network profile
- `GZH_DNS_OVERRIDE` - Enable/disable DNS override functionality
- `GZH_PROXY_AUTO_CONFIG` - Enable/disable automatic proxy configuration

## Examples

### Interactive TUI Mode

```bash
# Launch interactive network manager
gz net-env
```

### Network Status Checking

```bash
# Show current network status
gz net-env status

# Show detailed network information
gz net-env status --verbose

# Get status in JSON format
gz net-env status --json
```

### Profile Management

```bash
# List all profiles
gz net-env profile list

# Show specific profile
gz net-env profile show office
```

### Network Configuration Actions

```bash
# Execute network actions from configuration
gz net-env actions run

# Create example configuration
gz net-env actions config init

# Connect to VPN
gz net-env actions vpn connect --name office-vpn

# Set DNS servers
gz net-env actions dns set --servers 1.1.1.1,1.0.0.1

# Configure HTTP proxy
gz net-env actions proxy set --http proxy.company.com:8080
```

### Cloud Provider Management

```bash
# List configured cloud providers
gz net-env cloud list

# Configure AWS network settings
gz net-env cloud configure aws
```

## Integration Points

- **Development Environment**: Coordinates with `dev-env` for network-specific development configurations
- **Repository Management**: Integrates with `synclone` for network-aware repository access and proxy settings
- **Docker Management**: Basic coordination with Docker network settings
- **Package Management**: Network-aware package downloads

## Security Considerations

- **Configuration Security**: Secure storage of network credentials
- **Action Validation**: Basic validation of network actions before execution
- **Audit Logging**: Logging of network configuration changes
- **Proxy Authentication**: Secure handling of proxy credentials

## Platform Support

- **Linux**: Support for basic NetworkManager operations
- **macOS**: Support for basic system network configuration
- **Windows**: Limited support through basic network commands

## Future Enhancements

The following features are planned for future releases but are not currently implemented:

### Advanced Network Management
- **Smart Network Detection**: Automatic profile selection based on network environment
- **Network Profile Switching**: Advanced profile switching with auto-detection
- **Quick Actions**: Single commands for common network tasks
- **Network Monitoring**: Real-time network metrics and health monitoring

### VPN Management
- **Hierarchical VPN Management**: Manage layered VPN connections
- **VPN Profile Management**: Advanced VPN profiles and network mappings
- **VPN Failover Management**: VPN failover and backup connections

### Container and Cloud Integration
- **Docker Network Management**: Docker network profiles and configurations
- **Kubernetes Network Management**: Kubernetes network policies and configurations
- **Container Environment Detection**: Detect and analyze container environments
- **Advanced Cloud Integration**: Multi-cloud network management

### Network Analysis and Optimization
- **Network Topology Analysis**: Analyze and visualize network topology
- **Network Performance Monitoring**: Monitor network performance metrics
- **Network Analysis**: Analyze network performance and identify issues
- **Optimal Routing Management**: Manage and optimize network routing

### Advanced Features
- **Auto-Recovery**: Automatic reconnection and failback mechanisms
- **Profile Presets**: Ready-to-use configurations for common scenarios
- **Health Monitoring**: Continuous network component monitoring
- **Network Metrics Dashboard**: Real-time network performance visualization

## Summary

The current `net-env` command provides essential network management capabilities with a focus on simplicity and reliability. It offers:

### Current Capabilities
- **Interactive TUI**: Visual dashboard for network status viewing
- **Status Monitoring**: Comprehensive network status reporting
- **Basic Profile Management**: Simple network profile operations
- **Network Actions**: Essential network configuration tasks
- **Cloud Integration**: Basic cloud provider network management

### Design Philosophy
- **Simplicity**: Easy-to-use interface with clear command structure
- **Reliability**: Stable core functionality without complex dependencies
- **Extensibility**: Foundation for future advanced features
- **Safety**: Basic validation and error handling

The command serves as a solid foundation for network environment management while maintaining a clean, maintainable codebase that can be extended with additional features as needed.
