# net-env Command Reference

Network environment transitions for WiFi changes, VPN management, DNS configuration, and proxy settings.

## Synopsis

```bash
gz net-env <action> [flags]
gz net-env <action> --config <config-file>
```

## Description

The `net-env` command automates network environment transitions, managing VPN connections, DNS settings, proxy configurations, and environment-specific network policies.

## Network Components

- **WiFi** - Network detection and transition triggers
- **VPN** - VPN connection management and policies
- **DNS** - DNS server configuration per environment
- **Proxy** - HTTP/HTTPS proxy settings
- **Firewall** - Environment-specific firewall rules

## Actions

### `gz net-env transition`

Execute network environment transition.

```bash
gz net-env transition <environment> [flags]
```

**Arguments:**
- `environment` - Target environment: home, work, travel, public

**Flags:**
- `--auto-detect` - Auto-detect current environment
- `--force` - Force transition even if already in target environment
- `--dry-run` - Show what would be changed without executing

**Examples:**
```bash
# Transition to work environment
gz net-env transition work

# Auto-detect and transition
gz net-env transition --auto-detect

# Preview changes
gz net-env transition work --dry-run
```

### `gz net-env monitor`

Monitor network changes and auto-transition.

```bash
gz net-env monitor [flags]
```

**Flags:**
- `--interval` - Check interval (default: 5s)
- `--daemon` - Run as background daemon
- `--log-changes` - Log all network changes

**Examples:**
```bash
# Start monitoring
gz net-env monitor

# Run as daemon
gz net-env monitor --daemon

# Monitor with logging
gz net-env monitor --log-changes
```

### `gz net-env status`

Show current network environment status.

```bash
gz net-env status [flags]
```

**Flags:**
- `--detailed` - Show detailed network information
- `--output` - Output format: table, json, yaml

**Examples:**
```bash
# Basic status
gz net-env status

# Detailed information
gz net-env status --detailed

# JSON output
gz net-env status --output json
```

### `gz net-env configure`

Configure network environment settings.

```bash
gz net-env configure <environment> [flags]
```

**Arguments:**
- `environment` - Environment to configure

**Flags:**
- `--interactive` - Interactive configuration mode
- `--template` - Use configuration template

**Examples:**
```bash
# Configure work environment
gz net-env configure work --interactive

# Use template
gz net-env configure travel --template secure
```

### `gz net-env test`

Test network connectivity and configuration.

```bash
gz net-env test [flags]
```

**Flags:**
- `--environment` - Test specific environment
- `--component` - Test specific component: vpn, dns, proxy
- `--verbose` - Verbose test output

**Examples:**
```bash
# Test current environment
gz net-env test

# Test work environment
gz net-env test --environment work

# Test VPN connectivity
gz net-env test --component vpn
```

## Configuration

```yaml
version: "1.0"

# Environment definitions
environments:
  home:
    wifi_networks:
      - "HomeNetwork"
      - "HomeGuest"
    dns_servers:
      - "1.1.1.1"
      - "8.8.8.8"
    vpn:
      enabled: false
    proxy:
      enabled: false

  work:
    wifi_networks:
      - "CompanyWiFi"
      - "CompanyGuest"
    dns_servers:
      - "10.0.0.10"
      - "10.0.0.11"
    vpn:
      enabled: true
      connection: "company-vpn"
      auto_connect: true
    proxy:
      enabled: true
      http: "http://proxy.company.com:8080"
      https: "http://proxy.company.com:8080"
      exceptions:
        - "localhost"
        - "*.company.com"

  travel:
    wifi_networks: []  # Any network
    dns_servers:
      - "1.1.1.1"
      - "9.9.9.9"
    vpn:
      enabled: true
      connection: "secure-vpn"
      auto_connect: true
    proxy:
      enabled: false

# Detection rules
detection:
  methods:
    - wifi_ssid
    - gateway_ip
    - dns_servers

  rules:
    - environment: "work"
      conditions:
        wifi_ssid: "CompanyWiFi"
        gateway_ip: "10.0.0.1"

    - environment: "home"
      conditions:
        wifi_ssid: "HomeNetwork"
        gateway_ip: "192.168.1.1"

# Monitoring settings
monitoring:
  enabled: true
  interval: "10s"
  auto_transition: true
  notify_changes: true
```

## Environment Examples

### Work Environment Setup

```bash
# Configure work environment
gz net-env configure work

# Test work connectivity
gz net-env test --environment work

# Transition to work
gz net-env transition work
```

### Travel/Public WiFi Security

```bash
# Configure secure travel environment
gz net-env configure travel --template secure

# Auto-transition when on public WiFi
gz net-env monitor --auto-transition
```

### Home Office Setup

```bash
# Configure home environment
gz net-env configure home

# Transition to home settings
gz net-env transition home
```

## Network Components Management

### VPN Management

```bash
# List VPN connections
gz net-env vpn list

# Connect to VPN
gz net-env vpn connect company-vpn

# Disconnect VPN
gz net-env vpn disconnect

# Check VPN status
gz net-env vpn status
```

### DNS Configuration

```bash
# Set DNS servers
gz net-env dns set 1.1.1.1 8.8.8.8

# Reset to DHCP
gz net-env dns reset

# Check current DNS
gz net-env dns status
```

### Proxy Settings

```bash
# Configure proxy
gz net-env proxy set --http http://proxy:8080 --https http://proxy:8080

# Clear proxy settings
gz net-env proxy clear

# Show proxy status
gz net-env proxy status
```

## Automation Examples

### Automatic Transitions

```bash
# Start background monitoring
gz net-env monitor --daemon

# Check daemon status
gz net-env monitor --status

# Stop daemon
gz net-env monitor --stop
```

### Script Integration

```bash
#!/bin/bash
# Work setup script

# Transition to work environment
gz net-env transition work

# Wait for VPN connection
while ! gz net-env test --component vpn --quiet; do
    sleep 2
done

# Start work applications
echo "Work environment ready"
```

## Troubleshooting

### Common Issues

1. **VPN Connection Fails**
   ```bash
   # Check VPN configuration
   gz net-env test --component vpn --verbose

   # Reset VPN connection
   gz net-env vpn reset company-vpn
   ```

2. **DNS Resolution Issues**
   ```bash
   # Test DNS resolution
   gz net-env test --component dns

   # Flush DNS cache
   gz net-env dns flush
   ```

3. **Proxy Configuration Problems**
   ```bash
   # Test proxy connectivity
   gz net-env test --component proxy

   # Clear and reconfigure proxy
   gz net-env proxy clear
   gz net-env proxy set --http http://proxy:8080
   ```

### Debug Mode

```bash
# Enable debug logging
gz net-env transition work --debug

# Monitor with verbose output
gz net-env monitor --verbose --debug
```

## Security Considerations

### Public WiFi Safety

```yaml
travel:
  dns_servers:
    - "1.1.1.1"  # Cloudflare DNS over HTTPS
    - "9.9.9.9"  # Quad9 secure DNS
  vpn:
    enabled: true
    auto_connect: true
    kill_switch: true  # Disconnect internet if VPN fails
  firewall:
    enabled: true
    default_deny: true
    allowed_ports: [80, 443]
```

### Corporate Network Compliance

```yaml
work:
  dns_servers:
    - "10.0.0.10"  # Corporate DNS
  proxy:
    enabled: true
    authentication: true
    cert_validation: true
  monitoring:
    compliance_check: true
    audit_log: true
```

## Related Commands

- [`gz dev-env`](dev-env.md) - Development environment management
- [`gz profile`](profile.md) - Performance profiling

## See Also

- [Network Management Guide](../03-core-features/network-management/)
- [Network Examples](../../examples/network/)
