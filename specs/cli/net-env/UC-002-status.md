# Command: gz net-env status

## Scenario: Display current network environment status

### Input

**Command**:

```bash
gz net-env status
```

**Prerequisites**:

- [ ] Network access for connectivity tests
- [ ] Read permissions for system network configuration

### Expected Output

**Active Network Environment**:

```text
🌐 Network Environment Status

📋 Active Profile: office
   ⏰ Switched: 2025-09-02 09:15:30 KST
   📶 WiFi: CorpNet-5G (auto-detected)
   🔄 Duration: 5h 45m 20s

🔍 DNS Configuration
✅ Primary DNS: 10.0.1.53 (corp-dns-01.example.com)
   • Response time: 12ms
   • Status: healthy
✅ Secondary DNS: 10.0.1.54 (corp-dns-02.example.com)
   • Response time: 15ms
   • Status: healthy
✅ Search domains: corp.example.com, internal.local
✅ DNS-over-HTTPS: enabled (corporate policy)

🌐 HTTP Proxy Configuration
✅ HTTP Proxy: proxy.corp.example.com:8080
   • Authentication: NTLM (authenticated)
   • Connection: healthy
   • Response time: 89ms
✅ HTTPS Proxy: proxy.corp.example.com:8080
   • SSL inspection: enabled
   • Certificate: valid
✅ No proxy for: localhost, 127.0.0.1, *.local, *.corp.example.com

🔒 VPN Status
✅ Connection: office-vpn (connected)
   • Server: office.vpn.corp.com
   • Protocol: OpenVPN UDP
   • IP assigned: 10.8.0.156
   • Connected: 5h 44m
   • Data transferred: 234MB ↓ / 45MB ↑

🛣️  Network Routes
✅ Corporate subnet: 10.0.0.0/8 via 10.8.0.1 (VPN gateway)
✅ Development: 192.168.100.0/24 via 10.0.1.1 (local gateway)
✅ Internet: default via 192.168.1.1 (WiFi gateway)

📊 Connectivity Tests
✅ Internet: google.com (23ms)
✅ Corporate: intranet.corp.example.com (45ms)
✅ Development: dev.internal.local (12ms)
✅ VPN tunnel: vpn-test.corp.example.com (67ms)

🔄 Auto-Switch Rules
✅ CorpNet-5G → office (current)
✅ HomeWiFi → home
✅ iPhone-Hotspot → mobile
✅ Guest-WiFi → public-wifi

stderr: (empty)
Exit Code: 0
```

**No Network Environment**:

```text
🌐 Network Environment Status

❌ No active network environment configured

📶 Current Network: HomeWiFi-2.4G
   • IP: 192.168.1.45
   • Gateway: 192.168.1.1
   • DNS: 8.8.8.8, 8.8.4.4 (system default)

📋 Available profiles:
   • office - Corporate office environment
   • home - Home network optimized
   • mobile - Mobile hotspot configuration
   • public-wifi - Public WiFi with security

💡 Auto-switch available:
   gz net-env switch --auto

💡 Manual switch:
   gz net-env switch --profile home

🚫 Using system default network configuration.

stderr: (empty)
Exit Code: 1
```

**Network Issues Detected**:

```text
🌐 Network Environment Status

⚠️  Active Profile: office (degraded)
   ⏰ Switched: 2025-09-02 09:15:30 KST
   📶 WiFi: CorpNet-5G
   ⚠️  Issues detected: 3

🔍 DNS Configuration
❌ Primary DNS: 10.0.1.53 (timeout)
   • Last response: 3m ago
   • Status: unreachable
✅ Secondary DNS: 10.0.1.54 (active)
   • Response time: 18ms
   • Status: healthy (fallback active)
⚠️  Fallback to public DNS: 8.8.8.8 (temporary)

🌐 HTTP Proxy Configuration
❌ HTTP Proxy: proxy.corp.example.com:8080
   • Error: connection refused
   • Last successful: 15m ago
   • Fallback: direct connection (temporary)

🔒 VPN Status
⚠️  Connection: office-vpn (reconnecting)
   • Server: office.vpn.corp.com
   • Status: connection interrupted
   • Last connected: 2m ago
   • Retry attempt: 3/5

📊 Connectivity Tests
⚠️  Internet: google.com (245ms - using backup route)
❌ Corporate: intranet.corp.example.com (timeout)
✅ Development: dev.internal.local (34ms)
❌ VPN tunnel: vpn-test.corp.example.com (unreachable)

⚠️  Network environment degraded!

💡 Troubleshooting:
   - DNS issues: check network connectivity
   - Proxy failures: contact IT support
   - VPN reconnection: wait for auto-retry or manual restart

💡 Quick fixes:
   gz net-env switch --profile office --force-reconnect
   gz net-env diagnose --verbose

stderr: network issues detected
Exit Code: 2
```

### Side Effects

**Files Created**:

- `~/.gzh/net-env/status-cache.json` - Network status cache
- `~/.gzh/net-env/connectivity-log.json` - Connection test results

**Files Modified**: None
**State Changes**: Status cache and connectivity logs updated

### Validation

**Automated Tests**:

```bash
# Test network status display
result=$(gz net-env status 2>&1)
exit_code=$?

assert_contains "$result" "Network Environment Status"
# Exit code: 0 (healthy), 1 (no profile), 2 (issues)

# Check cache file creation
assert_file_exists "$HOME/.gzh/net-env/status-cache.json"
cache_content=$(cat "$HOME/.gzh/net-env/status-cache.json")
assert_contains "$cache_content" '"dns":'
assert_contains "$cache_content" '"timestamp":'
```

**Manual Verification**:

1. Check status with healthy network environment
1. Verify all connectivity tests are accurate
1. Test status with network issues
1. Confirm DNS resolution times are realistic
1. Validate VPN status information
1. Check auto-switch rule detection

### Edge Cases

**Multiple Network Interfaces**:

- Ethernet + WiFi simultaneously active
- VPN over multiple physical connections
- Cellular backup connection
- USB tethering scenarios

**Network Transitions**:

- WiFi network switching mid-test
- VPN disconnection during status check
- DNS server failover scenarios
- Proxy server rotation

**System Configuration Changes**:

- External network manager changes
- Manual DNS/proxy modifications
- VPN client updates or restarts
- Firewall rule changes

**Performance Monitoring**:

- Slow network connections (timeouts)
- High latency environments
- Bandwidth-limited connections
- Intermittent connectivity issues

### Performance Expectations

**Response Time**:

- Cached status: < 500ms
- DNS tests: < 2 seconds per server
- Connectivity tests: < 5 seconds total
- Full status check: < 10 seconds

**Resource Usage**:

- Memory: < 30MB
- Network: Minimal test traffic (< 1KB per test)
- CPU: Low impact status collection

**Test Coverage**:

- DNS resolution: 2-4 servers
- Connectivity: 4-6 endpoints
- Proxy: authentication and performance
- VPN: tunnel integrity and performance

## Notes

- Real-time network performance monitoring
- Automatic failover detection and reporting
- Historical connectivity trend analysis
- Integration with network monitoring tools
- Export capabilities for network diagnostics
- Proactive issue detection with alerting
- Cross-platform network interface detection
- Support for complex network topologies (corporate environments)
