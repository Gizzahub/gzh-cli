# Network Actions Documentation

The `gz net-env actions` command provides concrete implementations for common network configuration changes that are typically needed when switching between different network environments.

## Overview

Network actions allow you to automate VPN connections, DNS configuration, proxy settings, and hosts file management. This is particularly useful when combined with WiFi monitoring to automatically configure your network environment based on the current network.

## Usage

```bash
# Execute all actions from configuration file
gz net-env actions run

# Test configuration without executing
gz net-env actions run --dry-run

# Create example configuration
gz net-env actions config init

# Validate configuration
gz net-env actions config validate

# Individual action management
gz net-env actions vpn connect --name office
gz net-env actions dns set --servers 1.1.1.1,1.0.0.1
gz net-env actions proxy set --http http://proxy:8080
gz net-env actions hosts add --ip 192.168.1.100 --host server.local
```

## Configuration

### Configuration File Location

The network actions configuration is stored by default at:
- `~/.gz/network-actions.yaml`

You can specify a custom location with the `--config` flag.

### Configuration Structure

```yaml
# VPN Configuration
vpn:
  connect:
    - name: "office"
      type: "networkmanager"
      service: "office-vpn"
    - name: "home"
      type: "openvpn"
      config: "/etc/openvpn/home.conf"
    - name: "mobile"
      type: "wireguard"
      config: "/etc/wireguard/mobile.conf"
  disconnect:
    - "old-vpn"
    - "temp-vpn"

# DNS Configuration
dns:
  servers:
    - "1.1.1.1"
    - "1.0.0.1"
  interface: "wlan0"
  method: "resolvectl"  # resolvectl, networkmanager, manual

# Proxy Configuration
proxy:
  http: "http://proxy.company.com:8080"
  https: "http://proxy.company.com:8080"
  socks: "socks5://proxy.company.com:1080"
  no_proxy:
    - "localhost"
    - "127.0.0.1"
    - "*.local"
    - "company.internal"

# Hosts File Management
hosts:
  add:
    - ip: "192.168.1.100"
      host: "printer.local"
    - ip: "10.0.0.50"
      host: "dev-server.local"
  remove:
    - "old-server.local"
    - "deprecated.local"
```

## Actions

### VPN Actions

#### Connect VPN
```bash
gz net-env actions vpn connect --name office --type networkmanager
gz net-env actions vpn connect --name home --type openvpn --config /etc/openvpn/home.conf
gz net-env actions vpn connect --name mobile --type wireguard
```

**Supported VPN Types:**
- `networkmanager`: Uses NetworkManager VPN connections
- `openvpn`: Uses OpenVPN with systemd service
- `wireguard`: Uses WireGuard with wg-quick

#### Disconnect VPN
```bash
gz net-env actions vpn disconnect --name office
```

#### VPN Status
```bash
gz net-env actions vpn status
```

### DNS Actions

#### Set DNS Servers
```bash
gz net-env actions dns set --servers 1.1.1.1,1.0.0.1
gz net-env actions dns set --servers 8.8.8.8,8.8.4.4 --interface wlan0
```

#### DNS Status
```bash
gz net-env actions dns status
```

#### Reset DNS
```bash
gz net-env actions dns reset
```

### Proxy Actions

#### Set Proxy
```bash
gz net-env actions proxy set --http http://proxy:8080
gz net-env actions proxy set --https https://proxy:8080 --socks socks5://proxy:1080
```

#### Clear Proxy
```bash
gz net-env actions proxy clear
```

#### Proxy Status
```bash
gz net-env actions proxy status
```

### Hosts Actions

#### Add Host Entry
```bash
gz net-env actions hosts add --ip 192.168.1.100 --host server.local
```

#### Remove Host Entry
```bash
gz net-env actions hosts remove --host server.local
```

#### Show Hosts File
```bash
gz net-env actions hosts show
```

## Configuration Management

### Initialize Configuration
```bash
gz net-env actions config init
```

Creates an example configuration file with all sections populated.

### Validate Configuration
```bash
gz net-env actions config validate
```

Validates the configuration file syntax and structure.

## Integration with WiFi Monitoring

Network actions can be integrated with WiFi monitoring to automatically execute actions when network changes occur. See the WiFi monitoring documentation for details on how to configure actions to run automatically.

Example WiFi configuration with network actions:

```yaml
# ~/.gz/wifi-hooks.yaml
actions:
  - name: "office-network-setup"
    description: "Configure settings for office network"
    conditions:
      ssid: ["Office-WiFi"]
      state: ["connected"]
    commands:
      - "gz net-env actions vpn connect --name office"
      - "gz net-env actions dns set --servers 10.0.0.1,10.0.0.2"
      - "gz net-env actions proxy set --http http://proxy.company.com:8080"

  - name: "home-network-setup"
    description: "Configure settings for home network"
    conditions:
      ssid: ["Home-WiFi"]
      state: ["connected"]
    commands:
      - "gz net-env actions vpn disconnect --name office"
      - "gz net-env actions dns reset"
      - "gz net-env actions proxy clear"
```

## Command Reference

### Main Commands

| Command | Description |
|---------|-------------|
| `gz net-env actions run` | Execute all actions from configuration file |
| `gz net-env actions config init` | Create example configuration file |
| `gz net-env actions config validate` | Validate configuration file |

### VPN Commands

| Command | Description |
|---------|-------------|
| `gz net-env actions vpn connect` | Connect to VPN |
| `gz net-env actions vpn disconnect` | Disconnect from VPN |
| `gz net-env actions vpn status` | Show VPN status |

### DNS Commands

| Command | Description |
|---------|-------------|
| `gz net-env actions dns set` | Set DNS servers |
| `gz net-env actions dns status` | Show DNS configuration |
| `gz net-env actions dns reset` | Reset DNS to default |

### Proxy Commands

| Command | Description |
|---------|-------------|
| `gz net-env actions proxy set` | Set proxy configuration |
| `gz net-env actions proxy clear` | Clear proxy configuration |
| `gz net-env actions proxy status` | Show proxy status |

### Hosts Commands

| Command | Description |
|---------|-------------|
| `gz net-env actions hosts add` | Add entry to hosts file |
| `gz net-env actions hosts remove` | Remove entry from hosts file |
| `gz net-env actions hosts show` | Show hosts file contents |

## Flags

### Global Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--config` | Path to configuration file | `~/.gz/network-actions.yaml` |
| `--dry-run` | Show what would be executed without running | `false` |
| `--verbose` | Enable verbose logging | `false` |
| `--backup` | Create backup files before modifications | `true` |

### VPN Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--name` | VPN connection name | Required |
| `--type` | VPN type (networkmanager, openvpn, wireguard) | `networkmanager` |
| `--config` | VPN configuration file path | Auto-detected |

### DNS Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--servers` | Comma-separated DNS servers | Required |
| `--interface` | Network interface | Auto-detected |

### Proxy Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--http` | HTTP proxy URL | |
| `--https` | HTTPS proxy URL | |
| `--socks` | SOCKS proxy URL | |

### Hosts Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--ip` | IP address | Required for add |
| `--host` | Hostname | Required |

## Examples

### Basic Usage

```bash
# Create configuration
gz net-env actions config init

# Edit configuration
vim ~/.gz/network-actions.yaml

# Test configuration
gz net-env actions run --dry-run

# Execute actions
gz net-env actions run
```

### Office Network Setup

```bash
# Connect to office VPN
gz net-env actions vpn connect --name office --type networkmanager

# Set office DNS
gz net-env actions dns set --servers 10.0.0.1,10.0.0.2

# Configure proxy
gz net-env actions proxy set --http http://proxy.company.com:8080

# Add office servers to hosts
gz net-env actions hosts add --ip 10.0.0.50 --host dev-server.local
gz net-env actions hosts add --ip 10.0.0.100 --host build-server.local
```

### Home Network Setup

```bash
# Disconnect from office VPN
gz net-env actions vpn disconnect --name office

# Reset DNS to default
gz net-env actions dns reset

# Clear proxy
gz net-env actions proxy clear

# Remove office servers from hosts
gz net-env actions hosts remove --host dev-server.local
gz net-env actions hosts remove --host build-server.local
```

### Mobile/Travel Setup

```bash
# Connect to mobile VPN
gz net-env actions vpn connect --name mobile --type wireguard

# Use public DNS
gz net-env actions dns set --servers 1.1.1.1,1.0.0.1

# Clear proxy (direct connection)
gz net-env actions proxy clear
```

## Security Considerations

1. **Backup Files**: The tool creates backups before modifying system files like `/etc/hosts`
2. **Permissions**: Some actions require sudo privileges (VPN, DNS, hosts file changes)
3. **Configuration Security**: Store sensitive information like VPN configs securely
4. **Network Isolation**: Be aware that network changes affect all applications

## Troubleshooting

### Common Issues

1. **Permission Denied**: Some actions require sudo privileges
   ```bash
   sudo gz net-env actions vpn connect --name office
   ```

2. **VPN Connection Fails**: Check VPN configuration and credentials
   ```bash
   gz net-env actions vpn status
   ```

3. **DNS Changes Not Taking Effect**: Restart network service
   ```bash
   sudo systemctl restart systemd-resolved
   ```

4. **Proxy Settings Not Applied**: Check environment variables
   ```bash
   gz net-env actions proxy status
   ```

### Debug Mode

Use `--verbose` flag for detailed output:
```bash
gz net-env actions run --verbose
```

### Dry Run

Test configuration without making changes:
```bash
gz net-env actions run --dry-run
```

## Platform Support

- **Linux**: Full support with NetworkManager, systemd-resolved, systemd services
- **macOS**: Limited support (some features may require adaptation)
- **Windows**: Not currently supported

## Related Commands

- `gz net-env wifi monitor`: Monitor WiFi changes and trigger actions
- `gz net-env daemon`: Monitor system daemons and services
- `gz dev-env`: Development environment configuration
- `gz ssh-config`: SSH configuration management