# Network Environment Management Specification

## Overview

The `net-env` command provides comprehensive network environment management capabilities, enabling automatic detection of network changes and executing appropriate configuration adjustments. It manages WiFi transitions, VPN connections, DNS settings, proxy configurations, and Docker environment switches based on network context.

## Commands

### Core Commands

- `gz net-env status` - Show current network status
- `gz net-env switch` - Switch between network profiles
- `gz net-env actions` - Execute network configuration actions
- `gz net-env docker-network` - Manage Docker network profiles
- `gz net-env kubernetes-network` - Manage Kubernetes network policies
- `gz net-env container-detection` - Detect container environments
- `gz net-env network-topology` - Analyze network topology
- `gz net-env vpn-hierarchy` - Manage hierarchical VPN connections
- `gz net-env vpn-profile` - Manage VPN profiles
- `gz net-env vpn-failover` - Manage VPN failover
- `gz net-env network-metrics` - Monitor network performance
- `gz net-env network-analysis` - Analyze network performance
- `gz net-env optimal-routing` - Manage optimal routing

### Network Status (`gz net-env status`)

**Purpose**: Display current network configuration and status

**Features**:
- Show network interface information
- Display WiFi connection details
- Show VPN connection status
- Display DNS server configuration
- Show proxy settings

**Usage**:
```bash
gz net-env status                  # Show current network status
gz net-env status --verbose       # Show detailed network information
gz net-env status --json          # Output in JSON format
```

### Network Profile Switching (`gz net-env switch`)

**Purpose**: Switch between predefined network profiles

**Features**:
- Load network profiles from configuration files
- Switch between different network environments
- Apply multiple network configurations at once
- Support for environment-specific profiles

**Usage**:
```bash
gz net-env switch office          # Switch to office network profile
gz net-env switch home           # Switch to home network profile
gz net-env switch --list         # List available network profiles
gz net-env switch --init         # Create example configuration
```

### Network Actions (`gz net-env actions`)

**Purpose**: Execute predefined network configuration actions

**Features**:
- VPN connection management
- DNS server configuration
- Proxy settings management
- Hosts file modifications
- Custom script execution

**Usage**:
```bash
gz net-env actions run            # Execute network actions
gz net-env actions vpn connect --name office # Connect to VPN
gz net-env actions dns set --servers 1.1.1.1,1.0.0.1 # Set DNS servers
gz net-env actions proxy set --http proxy.company.com:8080 # Set proxy
```

### Docker Network Management (`gz net-env docker-network`)

**Purpose**: Manage Docker network profiles and configurations

**Features**:
- Create and manage Docker network profiles
- Apply network configurations to Docker environments
- Detect Docker Compose projects
- Generate profiles from Docker Compose files

**Usage**:
```bash
gz net-env docker-network list    # List Docker network profiles
gz net-env docker-network create myapp --network mynet # Create profile
gz net-env docker-network apply myapp # Apply Docker network profile
gz net-env docker-network detect  # Detect Docker Compose projects
```

### Kubernetes Network Management (`gz net-env kubernetes-network`)

**Purpose**: Manage Kubernetes network policies and configurations

**Features**:
- Create and manage Kubernetes network profiles
- Apply network policies to namespaces
- Manage service mesh configurations
- Handle network policy validation

**Usage**:
```bash
gz net-env kubernetes-network list # List Kubernetes network profiles
gz net-env kubernetes-network create prod-policies --namespace production # Create profile
gz net-env kubernetes-network policy add prod-policies web-policy \
  --pod-selector app=web --allow-from pod:app=api --ports TCP:8080 # Add policy
gz net-env kubernetes-network apply prod-policies # Apply profile
```

### Container Environment Detection (`gz net-env container-detection`)

**Purpose**: Detect and analyze container environments

**Features**:
- Detect running containers and networks
- Analyze container resource usage
- Identify container orchestration platforms
- Monitor container environment changes

**Usage**:
```bash
gz net-env container-detection detect # Detect container environments
gz net-env container-detection list   # List running containers
gz net-env container-detection monitor # Monitor container changes
gz net-env container-detection stats  # Show container statistics
```

### Network Topology Analysis (`gz net-env network-topology`)

**Purpose**: Analyze and visualize network topology

**Features**:
- Discover network services and connections
- Map container network relationships
- Export topology visualizations
- Identify network bottlenecks

**Usage**:
```bash
gz net-env network-topology analyze # Analyze network topology
gz net-env network-topology summary # Show topology summary
gz net-env network-topology export --format dot --output topology.dot # Export visualization
gz net-env network-topology services # Discover network services
```

### Hierarchical VPN Management (`gz net-env vpn-hierarchy`)

**Purpose**: Manage layered VPN connections

**Features**:
- Configure hierarchical VPN connections
- Manage parent-child VPN relationships
- Handle VPN failover scenarios
- Auto-connect based on network environment

**Usage**:
```bash
gz net-env vpn-hierarchy show     # Show VPN hierarchy
gz net-env vpn-hierarchy connect --root corp-vpn # Connect hierarchical VPN
gz net-env vpn-hierarchy auto-connect # Auto-connect for current environment
gz net-env vpn-hierarchy status   # Show hierarchy status
```

### VPN Profile Management (`gz net-env vpn-profile`)

**Purpose**: Manage VPN profiles and network mappings

**Features**:
- Create and manage VPN profiles
- Map networks to VPN connections
- Set connection priorities
- Handle network-specific VPN configurations

**Usage**:
```bash
gz net-env vpn-profile list       # List VPN profiles
gz net-env vpn-profile create office --network "Office WiFi" --vpn corp-vpn --priority 100 # Create profile
gz net-env vpn-profile map --network "Home WiFi" --vpn home-vpn --priority 50 # Map network to VPN
gz net-env vpn-profile show office # Show VPN profile details
```

### VPN Failover Management (`gz net-env vpn-failover`)

**Purpose**: Manage VPN failover and backup connections

**Features**:
- Configure backup VPN connections
- Handle connection loss scenarios
- Test failover configurations
- Monitor VPN connection health

**Usage**:
```bash
gz net-env vpn-failover start     # Start failover monitoring
gz net-env vpn-failover backup add --primary corp-vpn --backup home-vpn --priority 50 # Add backup
gz net-env vpn-failover test --scenario connection-loss # Test failover
gz net-env vpn-failover status    # Show failover status
```

### Network Performance Monitoring (`gz net-env network-metrics`)

**Purpose**: Monitor network performance metrics

**Features**:
- Monitor latency and bandwidth
- Track network interface usage
- Generate performance reports
- Set performance thresholds

**Usage**:
```bash
gz net-env network-metrics monitor # Start monitoring
gz net-env network-metrics show    # Show current metrics
gz net-env network-metrics latency --targets 8.8.8.8,1.1.1.1 # Test latency
gz net-env network-metrics bandwidth --interface eth0 # Monitor bandwidth
gz net-env network-metrics report --duration 1h # Generate report
```

### Network Analysis (`gz net-env network-analysis`)

**Purpose**: Analyze network performance and identify issues

**Features**:
- Latency analysis and monitoring
- Bandwidth utilization analysis
- Comprehensive network analysis
- Performance trend analysis
- Bottleneck detection

**Usage**:
```bash
gz net-env network-analysis latency --duration 10m --targets 8.8.8.8,1.1.1.1 # Latency analysis
gz net-env network-analysis bandwidth --interface eth0 --duration 5m # Bandwidth analysis
gz net-env network-analysis comprehensive --duration 15m # Comprehensive analysis
gz net-env network-analysis trends --period 24h # Performance trends
gz net-env network-analysis bottleneck # Detect bottlenecks
```

### Optimal Routing Management (`gz net-env optimal-routing`)

**Purpose**: Manage and optimize network routing

**Features**:
- Analyze optimal routes to destinations
- Discover best network interfaces
- Apply routing policies
- Enable auto-optimization

**Usage**:
```bash
gz net-env optimal-routing analyze --destination 8.8.8.8 # Analyze routes
gz net-env optimal-routing discover --targets google.com,cloudflare.com # Discover routes
gz net-env optimal-routing apply --policy latency-optimized # Apply routing
gz net-env optimal-routing auto-optimize --enable # Enable auto-optimization
gz net-env optimal-routing load-balance --interfaces eth0,wlan0 # Configure load balancing
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
net_env:
  # Network Profiles
  profiles:
    office:
      name: "Office Network"
      description: "Corporate office network configuration"
      vpn:
        name: "corporate-vpn"
        auto_connect: true
      dns:
        servers:
          - "10.0.0.1"
          - "10.0.0.2"
      proxy:
        http: "proxy.company.com:8080"
        https: "proxy.company.com:8080"
        no_proxy: "localhost,127.0.0.1,*.company.com"
      docker:
        context: "office"
      kubernetes:
        context: "office-cluster"

    home:
      name: "Home Network"
      description: "Home network configuration"
      vpn:
        name: "personal-vpn"
        auto_connect: false
      dns:
        servers:
          - "192.168.1.1"
          - "8.8.8.8"
      proxy:
        enabled: false
      docker:
        context: "home"
      kubernetes:
        context: "home-cluster"

    public:
      name: "Public Network"
      description: "Public WiFi network configuration"
      vpn:
        name: "secure-vpn"
        auto_connect: true
      dns:
        servers:
          - "1.1.1.1"
          - "9.9.9.9"
      proxy:
        enabled: false
      docker:
        context: "secure"
      kubernetes:
        context: "secure-cluster"

  # VPN Hierarchies
  vpn_hierarchies:
    corporate:
      root: "corporate-vpn"
      children:
        - "personal-vpn"
        - "backup-vpn"
      auto_connect: true

    public:
      root: "public-vpn"
      children: []
      auto_connect: true

  # Network Actions
  actions:
    vpn_connect_office:
      type: "vpn"
      action: "connect"
      vpn_name: "corporate-vpn"

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

  # Docker Network Profiles
  docker_profiles:
    office:
      networks:
        - name: "office-net"
          driver: "bridge"
          subnet: "172.20.0.0/16"
      services:
        - name: "database"
          network: "office-net"
          ports:
            - "5432:5432"

    home:
      networks:
        - name: "home-net"
          driver: "bridge"
          subnet: "172.21.0.0/16"

  # Kubernetes Network Policies
  kubernetes_profiles:
    production:
      namespace: "production"
      policies:
        - name: "web-policy"
          pod_selector:
            match_labels:
              app: "web"
          ingress:
            - from:
                - pod_selector:
                    match_labels:
                      app: "api"
              ports:
                - protocol: "TCP"
                  port: 8080

    staging:
      namespace: "staging"
      policies:
        - name: "allow-all"
          pod_selector: {}
          ingress:
            - {}

  # Monitoring Settings
  monitoring:
    enabled: true
    interval: "30s"
    metrics_retention: "24h"

  # Notification Settings
  notifications:
    enabled: true
    methods:
      - "stdout"
      - "desktop"
    events:
      - "network_change"
      - "vpn_connect"
      - "vpn_disconnect"
```

### Environment Variables

- `GZH_NET_ENV_CONFIG` - Path to configuration file
- `GZH_NET_ENV_PROFILE` - Override default network profile
- `GZH_VPN_AUTO_CONNECT` - Enable/disable automatic VPN connection
- `GZH_DNS_OVERRIDE` - Enable/disable DNS override functionality
- `GZH_PROXY_AUTO_CONFIG` - Enable/disable automatic proxy configuration

## Examples

### Network Profile Switching

```bash
# Switch to office network profile
gz net-env switch office

# Switch to home network profile
gz net-env switch home

# List available profiles
gz net-env switch --list

# Create example configuration
gz net-env switch --init
```

### VPN Management

```bash
# Connect to corporate VPN
gz net-env actions vpn connect --name corporate-vpn

# Disconnect from VPN
gz net-env actions vpn disconnect --name corporate-vpn

# Show VPN status
gz net-env vpn-hierarchy show

# Auto-connect hierarchical VPN
gz net-env vpn-hierarchy auto-connect
```

### DNS and Proxy Configuration

```bash
# Set DNS servers
gz net-env actions dns set --servers 1.1.1.1,1.0.0.1

# Set HTTP proxy
gz net-env actions proxy set --http proxy.company.com:8080

# Disable proxy
gz net-env actions proxy disable
```

### Container Network Management

```bash
# Create Docker network profile
gz net-env docker-network create myapp --network mynet --driver bridge

# Apply Docker network profile
gz net-env docker-network apply myapp

# Detect Docker Compose projects
gz net-env docker-network detect

# Create Kubernetes network profile
gz net-env kubernetes-network create prod-policies --namespace production

# Add network policy to profile
gz net-env kubernetes-network policy add prod-policies web-policy \
  --pod-selector app=web --allow-from pod:app=api --ports TCP:8080

# Apply Kubernetes network profile
gz net-env kubernetes-network apply prod-policies
```

### Network Monitoring and Analysis

```bash
# Monitor network metrics
gz net-env network-metrics monitor

# Test latency to specific targets
gz net-env network-metrics latency --targets 8.8.8.8,1.1.1.1,google.com

# Monitor bandwidth usage
gz net-env network-metrics bandwidth --interface eth0

# Generate performance report
gz net-env network-metrics report --duration 1h

# Analyze network latency
gz net-env network-analysis latency --duration 10m --targets 8.8.8.8,1.1.1.1

# Comprehensive network analysis
gz net-env network-analysis comprehensive --duration 15m

# Discover optimal routes
gz net-env optimal-routing discover --targets google.com,cloudflare.com

# Apply optimal routing
gz net-env optimal-routing apply --policy latency-optimized
```

## Integration Points

- **Development Environment**: Coordinates with `dev-env` for network-specific development configurations
- **Repository Management**: Integrates with `synclone` for network-aware repository access and proxy settings
- **IDE Settings**: Works with `ide` command for network-specific IDE configurations
- **Docker Management**: Synchronizes with Docker contexts and registry configurations
- **Package Management**: Coordinates with `always-latest` for network-aware package downloads

## Security Considerations

- **VPN Kill Switch**: Prevents traffic leakage when VPN connection drops
- **DNS Leak Protection**: Ensures DNS queries go through intended servers
- **Proxy Authentication**: Secure storage and handling of proxy credentials
- **Configuration Encryption**: Sensitive configuration data encryption at rest
- **Action Validation**: Verification of network actions before execution
- **Audit Logging**: Complete logging of all network configuration changes
- **Network Isolation**: Support for isolated network environments
- **Certificate Management**: Automatic handling of network-specific certificates

## Platform Support

- **Linux**: Full support for NetworkManager, systemd-resolved, iptables
- **macOS**: Support for Airport, System Configuration framework
- **Windows**: Limited support through netsh and PowerShell modules
- **Container Environments**: Docker and Kubernetes network management
- **Cloud Platforms**: AWS, GCP, Azure network service integration
