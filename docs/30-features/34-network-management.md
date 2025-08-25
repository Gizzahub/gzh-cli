# 🌐 Network Environment Management

The `gz net-env` command provides intelligent network environment management, including WiFi profile detection, VPN automation, proxy configuration, and DNS management.

## 📋 Table of Contents

- [Overview](#overview)
- [WiFi Profile Management](#wifi-profile-management)
- [VPN Management](#vpn-management)
- [Proxy Configuration](#proxy-configuration)
- [DNS Management](#dns-management)
- [Network Monitoring](#network-monitoring)
- [Integration Examples](#integration-examples)

## 🎯 Overview

Modern development often involves working across different network environments - home, office, coworking spaces, or remote locations. The `gz net-env` command automatically detects network changes and applies appropriate configurations for optimal development experience.

### Key Features

- **Automatic Network Detection** - Detects WiFi network changes
- **Profile-Based Configuration** - Different settings per network
- **VPN Automation** - Automatic VPN connection management
- **Proxy Management** - Dynamic proxy configuration
- **DNS Optimization** - Custom DNS settings per environment
- **Security Policies** - Network-based security rules

## 📡 WiFi Profile Management

### Network Detection and Profiles

```bash
# Show current network information
gz net-env status

# List available WiFi networks
gz net-env wifi scan

# Create network profile
gz net-env profile create --name office \
  --ssid "Company-WiFi" \
  --type corporate

# List network profiles
gz net-env profile list

# Apply specific profile
gz net-env profile apply --name office
```

### Automatic Network Switching

```bash
# Enable automatic profile switching
gz net-env auto-switch --enable

# Monitor network changes
gz net-env monitor

# Configure switch triggers
gz net-env trigger add --ssid "Home-WiFi" --profile home

# Test profile switching
gz net-env test-switch --from office --to home
```

### Network Profile Configuration

```bash
# Configure office network profile
gz net-env profile config --name office \
  --vpn company-vpn \
  --proxy http://proxy.company.com:8080 \
  --dns 8.8.8.8,1.1.1.1

# Configure home network profile
gz net-env profile config --name home \
  --vpn disabled \
  --proxy disabled \
  --dns 1.1.1.1,8.8.4.4

# Import profile from file
gz net-env profile import --file office-profile.yaml
```

## 🔐 VPN Management

### VPN Connection Management

```bash
# List available VPN connections
gz net-env vpn list

# Connect to VPN
gz net-env vpn connect --name company-vpn

# Disconnect from VPN
gz net-env vpn disconnect

# Check VPN status
gz net-env vpn status

# Test VPN connectivity
gz net-env vpn test --endpoint internal.company.com
```

### VPN Configuration

```bash
# Add VPN configuration
gz net-env vpn add --name company-vpn \
  --type openvpn \
  --config /path/to/company.ovpn

# Add WireGuard VPN
gz net-env vpn add --name wg-office \
  --type wireguard \
  --config /path/to/wg-office.conf

# Update VPN configuration
gz net-env vpn update --name company-vpn \
  --auto-connect true \
  --retry-attempts 3

# Test VPN configuration
gz net-env vpn validate --name company-vpn
```

### Automatic VPN Management

```bash
# Configure automatic VPN connection
gz net-env vpn auto --profile office --vpn company-vpn

# Set VPN connection rules
gz net-env vpn rule add \
  --trigger "network.ssid != 'Company-WiFi'" \
  --action "connect company-vpn"

# Monitor VPN health
gz net-env vpn monitor --auto-reconnect
```

## 🔄 Proxy Configuration

### Proxy Management

```bash
# Configure HTTP proxy
gz net-env proxy set --http http://proxy.company.com:8080

# Configure HTTPS proxy
gz net-env proxy set --https https://proxy.company.com:8080

# Configure SOCKS proxy
gz net-env proxy set --socks socks5://proxy.company.com:1080

# Set proxy bypass rules
gz net-env proxy bypass --add "*.company.com,localhost,127.0.0.1"

# Test proxy connectivity
gz net-env proxy test --url https://github.com
```

### Automatic Proxy Configuration

```bash
# Enable proxy auto-configuration
gz net-env proxy auto --profile office

# Configure proxy rules per network
gz net-env proxy rule add \
  --profile office \
  --proxy http://proxy.company.com:8080

# Disable proxy for specific networks
gz net-env proxy rule add \
  --profile home \
  --proxy disabled
```

### Proxy Authentication

```bash
# Configure proxy with authentication
gz net-env proxy auth --username employee \
  --password-cmd "pass proxy/company"

# Use system keychain for proxy auth
gz net-env proxy auth --keychain

# Test authenticated proxy
gz net-env proxy test --auth
```

## 🌍 DNS Management

### DNS Configuration

```bash
# Set custom DNS servers
gz net-env dns set --servers 8.8.8.8,1.1.1.1

# Add DNS server for specific domain
gz net-env dns add --domain company.com --server 10.0.0.1

# Configure DNS over HTTPS
gz net-env dns doh --enable --provider cloudflare

# Test DNS resolution
gz net-env dns test --domain github.com
```

### Dynamic DNS Configuration

```bash
# Configure network-specific DNS
gz net-env dns profile --name office \
  --servers 10.0.0.1,10.0.0.2 \
  --search company.com,corp.local

# Configure public DNS for untrusted networks
gz net-env dns profile --name public \
  --servers 1.1.1.1,8.8.8.8 \
  --doh-provider cloudflare

# Apply DNS profile automatically
gz net-env dns auto --untrusted-networks public
```

## 📊 Network Monitoring

### Real-time Network Monitoring

```bash
# Monitor network changes
gz net-env monitor

# Monitor with detailed logging
gz net-env monitor --verbose

# Monitor specific metrics
gz net-env monitor --metrics connectivity,latency,bandwidth

# Export monitoring data
gz net-env monitor --export network-metrics.json
```

### Network Performance Testing

```bash
# Test network connectivity
gz net-env test connectivity

# Test network speed
gz net-env test speed

# Test latency to specific hosts
gz net-env test latency --hosts github.com,google.com

# Comprehensive network test
gz net-env test all --output network-report.json
```

### Network Security Scanning

```bash
# Scan for open ports
gz net-env security scan-ports

# Check for suspicious network activity
gz net-env security monitor

# Validate security policies
gz net-env security validate --profile office

# Generate security report
gz net-env security report --output security-scan.json
```

## ⚙️ Configuration

### Basic Configuration

Add network environment settings to your `~/.config/gzh-manager/gzh.yaml`:

```yaml
commands:
  net_env:
    # Enable automatic network detection
    auto_detect: true

    # Default monitoring interval
    monitor_interval: "30s"

    # Network profiles
    profiles:
      office:
        ssid: "Company-WiFi"
        vpn: "company-vpn"
        proxy: "http://proxy.company.com:8080"
        dns: ["10.0.0.1", "10.0.0.2"]

      home:
        ssid: "Home-WiFi"
        vpn: disabled
        proxy: disabled
        dns: ["1.1.1.1", "8.8.8.8"]

      public:
        type: fallback
        vpn: "public-vpn"
        proxy: disabled
        dns: ["1.1.1.1", "8.8.4.4"]
        security_level: high

    # VPN configurations
    vpn:
      company-vpn:
        type: openvpn
        config_file: "/etc/openvpn/company.ovpn"
        auto_connect: true
        retry_attempts: 3

      public-vpn:
        type: wireguard
        config_file: "/etc/wireguard/public.conf"
        auto_connect: false
```

### Advanced Configuration

```yaml
commands:
  net_env:
    # Security policies
    security:
      # Require VPN on unknown networks
      require_vpn_on_unknown: true

      # Block certain domains on public networks
      public_network_blocks:
        - "*.company.com"
        - "internal.*"

      # Enable firewall rules
      firewall_rules:
        - rule: "block incoming on public networks"
          networks: ["public", "unknown"]

    # Monitoring settings
    monitoring:
      # Network change detection
      detect_changes: true
      change_threshold: "5s"

      # Performance monitoring
      performance_tests:
        enabled: true
        interval: "5m"
        targets: ["8.8.8.8", "1.1.1.1"]

      # Security monitoring
      security_monitoring:
        enabled: true
        scan_interval: "10m"
        alert_on_suspicious: true

    # Integration settings
    integration:
      # Notify on network changes
      notifications:
        enabled: true
        methods: ["desktop", "log"]

      # External integrations
      webhook_url: "http://monitoring.company.com/network-events"

      # Sync with external services
      sync:
        enabled: false
        service: "company-network-config"
```

## 🚀 Integration Examples

### Development Workflow Integration

```bash
# Automatic environment setup on network change
gz net-env monitor --on-change "./setup-dev-env.sh"

# Integration with IDE monitoring
gz net-env auto-switch --enable
gz ide monitor &

# Combined with repository sync
gz net-env profile apply --name office
gz synclone github --org company
```

### CI/CD Pipeline Integration

```yaml
# GitHub Actions example
- name: Configure Network Environment
  run: |
    gz net-env profile apply --name ci
    gz net-env test connectivity --required
```

### Container Development

```bash
# Configure proxy for Docker
gz net-env proxy docker-config --output docker-proxy.env

# Set up network for Kubernetes
gz net-env profile apply --name k8s-dev
gz dev-env k8s context use development
```

## 🔧 Automation and Scripting

### Network Change Automation

```bash
# Script executed on network change
cat > ~/.config/gzh-manager/on-network-change.sh << 'EOF'
#!/bin/bash
NETWORK_PROFILE=$(gz net-env detect-profile)
case $NETWORK_PROFILE in
  office)
    gz dev-env aws --profile work
    gz net-env vpn connect company-vpn
    ;;
  home)
    gz dev-env aws --profile personal
    gz net-env vpn disconnect
    ;;
  public)
    gz net-env vpn connect public-vpn
    gz net-env security enable-strict
    ;;
esac
EOF

gz net-env monitor --on-change ~/.config/gzh-manager/on-network-change.sh
```

### Scheduled Network Tasks

```bash
# Daily network health check
gz net-env test all --cron "0 9 * * *"

# Weekly security scan
gz net-env security scan --cron "0 0 * * 0"

# Monitor and restart VPN if needed
gz net-env vpn monitor --auto-reconnect --daemon
```

## 📋 Output Formats

All network commands support multiple output formats:

```bash
# JSON output for automation
gz net-env status --output json

# YAML for configuration
gz net-env profile export --output yaml

# Table format (default)
gz net-env wifi scan --output table

# CSV for analysis
gz net-env monitor --output csv
```

## 🆘 Troubleshooting

### Network Connection Issues

```bash
# Diagnose network problems
gz net-env diagnose

# Test basic connectivity
gz net-env test connectivity --verbose

# Reset network configuration
gz net-env reset --confirm

# Check network interface status
gz net-env interfaces --status
```

### VPN Connection Problems

```bash
# Debug VPN connection
gz net-env vpn debug --name company-vpn

# Test VPN configuration
gz net-env vpn validate --config /path/to/config

# Reset VPN state
gz net-env vpn reset --name company-vpn
```

### Proxy Configuration Issues

```bash
# Test proxy connectivity
gz net-env proxy test --detailed

# Debug proxy authentication
gz net-env proxy debug --auth

# Reset proxy configuration
gz net-env proxy reset
```

______________________________________________________________________

**Network Detection**: Automatic WiFi profile detection
**VPN Support**: OpenVPN, WireGuard, system VPN
**Proxy Types**: HTTP, HTTPS, SOCKS5 with authentication
**DNS**: Custom DNS, DNS over HTTPS, per-network configuration
**Security**: Network-based security policies and monitoring
