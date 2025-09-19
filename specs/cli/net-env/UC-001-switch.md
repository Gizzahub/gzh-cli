# Command: gz net-env switch

## Scenario: Switch network environment configuration

### Input

**Command**:

```bash
gz net-env switch --profile office
```

**Prerequisites**:

- [ ] Network environment profiles configured
- [ ] Administrative privileges for network changes
- [ ] VPN client software installed (if required)

### Expected Output

**Success Case**:

```text
🌐 Switching to network environment: office

📋 Network Configuration Changes:
✅ DNS Servers:
   • Primary: 8.8.8.8 → 10.0.1.53
   • Secondary: 8.8.4.4 → 10.0.1.54
   • Search domains: → corp.example.com, internal.local

✅ HTTP Proxy:
   • Proxy: none → proxy.corp.example.com:8080
   • No proxy: → localhost,127.0.0.1,*.local,*.corp.example.com
   • Authentication: enabled

✅ VPN Connection:
   • Profile: home-vpn → office-vpn
   • Endpoint: home.vpn.example.com → office.vpn.corp.com
   • Status: connected → connected

✅ Network Routes:
   • Corporate subnet: 10.0.0.0/8 via VPN gateway
   • Development servers: 192.168.100.0/24 via proxy
   • External traffic: default gateway

🎉 Successfully switched to office network environment!

💡 Active configuration:
   - Corporate DNS for internal resolution
   - HTTP/HTTPS traffic through corporate proxy
   - VPN tunnel established for secure access
   - Custom routes for internal services

⏰ Auto-switch enabled: will switch to 'home' when office WiFi disconnects

stderr: (empty)
Exit Code: 0
```

**VPN Connection Failed**:

```text
🌐 Switching to network environment: office

📋 Network Configuration Changes:
✅ DNS Servers: 8.8.8.8 → 10.0.1.53
✅ HTTP Proxy: configured for proxy.corp.example.com:8080

❌ VPN Connection Failed:
   • Profile: office-vpn
   • Error: authentication failed
   • Server: office.vpn.corp.com
   • Reason: invalid certificate or expired credentials

💡 Manual VPN setup required:
   - Check VPN credentials: gz net-env profile edit office
   - Test VPN manually: openvpn --config ~/.gzh/net-env/office-vpn.conf
   - Update certificate: openssl x509 -in cert.pem -noout -dates

⚠️  Partial network switch completed. VPN connection pending.

stderr: VPN authentication failed
Exit Code: 1
```

**Profile Not Found**:

```text
🔍 Searching for network environment: office

❌ Network profile 'office' not found!

📋 Available profiles:
   • home (currently active)
   • mobile-hotspot
   • public-wifi
   • guest-network

💡 Create new profile:
   gz net-env create --profile office

🚫 Network environment switch failed.

stderr: profile not found
Exit Code: 1
```

**Permission Denied**:

```text
🌐 Switching to network environment: office

❌ Insufficient permissions for network changes:
   • DNS configuration: requires admin/root
   • System proxy settings: requires admin/root
   • VPN profile management: requires admin/root

💡 Solutions:
   - Run with sudo: sudo gz net-env switch --profile office
   - macOS: grant Full Disk Access in System Preferences
   - Windows: run as Administrator
   - Linux: ensure user in netdev group or use sudo

🚫 Network environment switch failed.

stderr: permission denied
Exit Code: 2
```

### Side Effects

**Files Created**:

- `~/.gzh/net-env/current.yaml` - Active network state
- `~/.gzh/net-env/switch.log` - Network change log
- `/tmp/gz-net-backup-*.conf` - Original settings backup

**Files Modified**:

- System DNS configuration (`/etc/resolv.conf`, Registry, etc.)
- HTTP proxy settings (system-wide or user-specific)
- VPN client configuration files
- Network routing tables

**State Changes**:

- DNS resolution servers changed
- HTTP/HTTPS proxy configuration applied
- VPN connection established/terminated
- Network routes added/removed
- Firewall rules updated (if configured)

### Validation

**Automated Tests**:

```bash
# Test network switch (requires test environment)
result=$(gz net-env switch --profile test-net 2>&1)
exit_code=$?

assert_contains "$result" "Switching to network environment"
# Exit code varies: 0 (success), 1 (partial), 2 (permission)

# Check state file creation
assert_file_exists "$HOME/.gzh/net-env/current.yaml"
current_profile=$(yq r "$HOME/.gzh/net-env/current.yaml" 'active_profile')
assert_equals "$current_profile" "test-net"

# Verify DNS resolution works
nslookup_result=$(nslookup google.com 2>&1)
assert_exit_code 0
```

**Manual Verification**:

1. Switch between different network profiles
1. Verify DNS resolution uses correct servers
1. Test HTTP traffic goes through proxy (if configured)
1. Confirm VPN connection establishes properly
1. Check network routes are applied correctly
1. Test auto-switch behavior with WiFi changes

### Edge Cases

**WiFi Network Detection**:

- Automatic profile switching based on WiFi SSID
- Multiple SSIDs mapped to same profile
- Hidden network handling
- Network priority when multiple available

**Connectivity Issues**:

- DNS server unreachable
- Proxy server down or unreachable
- VPN server connection timeout
- Internet connectivity loss during switch

**Configuration Conflicts**:

- Overlapping network routes
- DNS server conflicts
- Proxy authentication failures
- VPN certificate expiration

**System Integration**:

- Network manager conflicts (NetworkManager, systemd-networkd)
- VPN client compatibility (OpenVPN, WireGuard, IKEv2)
- Platform-specific proxy settings
- Firewall integration

### Performance Expectations

**Response Time**:

- DNS changes: < 2 seconds
- Proxy configuration: < 3 seconds
- VPN connection: < 15 seconds
- Full network switch: < 20 seconds

**Resource Usage**:

- Memory: < 50MB
- Network: Initial connection tests
- CPU: Low during normal operation, higher during VPN handshake

**Network Impact**:

- Brief connectivity interruption during DNS switch
- Temporary proxy authentication delays
- VPN tunnel establishment time

## Notes

- Cross-platform network configuration (macOS, Windows, Linux)
- Automatic WiFi-based profile switching
- VPN client integration (OpenVPN, WireGuard, built-in clients)
- System proxy configuration (HTTP, HTTPS, SOCKS)
- DNS-over-HTTPS and DNS-over-TLS support
- Network troubleshooting and diagnostics
- Backup and restore of original network settings
- Integration with corporate network policies
