# Command: gz net-env profiles

## Scenario: Manage network environment profiles

### Input

**Command**:

```bash
gz net-env profiles list
```

**Prerequisites**:

- [ ] Network environment system initialized
- [ ] Access to profile configuration directory

### Expected Output

**List Profiles**:

```text
🌐 Network Environment Profiles

📋 Available Profiles (4):

● office [ACTIVE]
   📶 Auto-switch: CorpNet-5G, CorpNet-Guest
   🔍 DNS: 10.0.1.53, 10.0.1.54 (corp-dns.example.com)
   🌐 Proxy: proxy.corp.example.com:8080 (NTLM)
   🔒 VPN: office-vpn (OpenVPN) → office.vpn.corp.com
   📊 Usage: 15 switches, last used 2h ago
   ⏰ Created: 2024-12-15, Modified: 2025-08-20

  home
   📶 Auto-switch: HomeWiFi-5G, HomeWiFi-2.4G
   🔍 DNS: 1.1.1.1, 1.0.0.1 (Cloudflare DNS)
   🌐 Proxy: none (direct connection)
   🔒 VPN: home-vpn (WireGuard) → home.vpn.example.com
   📊 Usage: 45 switches, last used 8h ago
   ⏰ Created: 2024-10-01, Modified: 2025-07-12

  mobile
   📶 Auto-switch: iPhone-Hotspot, AndroidAP-*, Hotspot-*
   🔍 DNS: 8.8.8.8, 8.8.4.4 (Google DNS)
   🌐 Proxy: none
   🔒 VPN: mobile-vpn (IKEv2) → mobile.vpn.example.com
   📊 Usage: 23 switches, last used 1d ago
   ⏰ Created: 2024-11-20, Modified: 2025-06-05

  public-wifi
   📶 Auto-switch: Guest-*, Public-*, Starbucks-WiFi
   🔍 DNS: 9.9.9.9, 149.112.112.112 (Quad9 DNS)
   🌐 Proxy: none
   🔒 VPN: secure-vpn (OpenVPN) → secure.vpn.example.com
   📊 Usage: 8 switches, last used 3d ago
   ⏰ Created: 2024-12-01, Modified: 2025-05-15

💡 Manage profiles:
   gz net-env profiles create <name>    # Create new profile
   gz net-env profiles edit <name>      # Edit existing profile
   gz net-env profiles delete <name>    # Delete profile
   gz net-env profiles export <name>    # Export profile config

stderr: (empty)
Exit Code: 0
```

**Create Profile**:

```text
# Command: gz net-env profiles create cafe --interactive

🌐 Creating network environment profile: cafe

📋 Profile Configuration:

📶 WiFi Auto-Switch Configuration:
   WiFi networks for auto-switching (comma-separated):
   > StarBucks-WiFi, CoffeeBean-Guest, Cafe-Free

🔍 DNS Configuration:
   Primary DNS server [8.8.8.8]: 1.1.1.1
   Secondary DNS server [8.8.4.4]: 1.0.0.1
   DNS-over-HTTPS [y/N]: y
   Search domains (optional):

🌐 HTTP Proxy Configuration:
   Enable HTTP proxy [y/N]: n

🔒 VPN Configuration:
   Enable VPN [y/N]: y
   VPN type (openvpn/wireguard/ikev2) [openvpn]: wireguard
   VPN server: public.vpn.example.com
   VPN port [51820]:
   Configuration file: ~/.gzh/net-env/vpn/cafe-vpn.conf

📊 Profile Summary:
   Name: cafe
   WiFi networks: StarBucks-WiFi, CoffeeBean-Guest, Cafe-Free
   DNS: 1.1.1.1, 1.0.0.1 (DoH enabled)
   Proxy: none
   VPN: WireGuard → public.vpn.example.com:51820

✅ Profile 'cafe' created successfully!

💡 Test profile: gz net-env switch --profile cafe --test
💡 Use profile: gz net-env switch --profile cafe

stderr: (empty)
Exit Code: 0
```

**Edit Profile**:

```text
# Command: gz net-env profiles edit office

🌐 Editing network environment profile: office

📋 Current Configuration:

📶 WiFi Auto-Switch Networks:
   Current: CorpNet-5G, CorpNet-Guest
   Update [CorpNet-5G, CorpNet-Guest]: CorpNet-5G, CorpNet-Guest, CorpNet-Legacy

🔍 DNS Configuration:
   Primary DNS [10.0.1.53]:
   Secondary DNS [10.0.1.54]:
   Search domains [corp.example.com, internal.local]:

🌐 HTTP Proxy Configuration:
   HTTP proxy [proxy.corp.example.com:8080]:
   Authentication type [NTLM]:
   Username [corp\username]: corp\newusername

🔒 VPN Configuration:
   VPN type [openvpn]:
   VPN server [office.vpn.corp.com]:
   Configuration file [~/.gzh/net-env/vpn/office-vpn.conf]:

📊 Updated Configuration:
   WiFi networks: CorpNet-5G, CorpNet-Guest, CorpNet-Legacy
   DNS: 10.0.1.53, 10.0.1.54
   Proxy: proxy.corp.example.com:8080 (NTLM: corp\newusername)
   VPN: OpenVPN → office.vpn.corp.com

✅ Profile 'office' updated successfully!

💡 Test changes: gz net-env switch --profile office --test

stderr: (empty)
Exit Code: 0
```

**Delete Profile**:

```text
# Command: gz net-env profiles delete old-office

🌐 Deleting network environment profile: old-office

⚠️  WARNING: This action cannot be undone!

📋 Profile Details:
   Name: old-office
   Created: 2024-08-15
   Last used: 45d ago
   Usage count: 3 switches

🗂️  Files to be removed:
   • ~/.gzh/net-env/profiles/old-office.yaml
   • ~/.gzh/net-env/vpn/old-office-vpn.conf
   • ~/.gzh/net-env/logs/old-office-*.log

Confirm deletion [y/N]: y

✅ Profile 'old-office' deleted successfully!

📊 Remaining profiles: 4 (office, home, mobile, public-wifi)

stderr: (empty)
Exit Code: 0
```

**Export Profile**:

```text
# Command: gz net-env profiles export office --format yaml

🌐 Exporting network environment profile: office

📄 Profile exported to: ~/.gzh/exports/net-env-office-20250902.yaml

📋 Export contents:
---
profile:
  name: office
  version: 2.1.0
  created: 2024-12-15T10:30:00Z
  modified: 2025-08-20T14:25:00Z

wifi:
  auto_switch_networks:
    - CorpNet-5G
    - CorpNet-Guest
    - CorpNet-Legacy

dns:
  primary: 10.0.1.53
  secondary: 10.0.1.54
  search_domains:
    - corp.example.com
    - internal.local
  doh_enabled: true

proxy:
  http_proxy: proxy.corp.example.com:8080
  https_proxy: proxy.corp.example.com:8080
  auth_type: ntlm
  username: corp\username
  no_proxy:
    - localhost
    - 127.0.0.1
    - "*.local"
    - "*.corp.example.com"

vpn:
  enabled: true
  type: openvpn
  server: office.vpn.corp.com
  port: 1194
  config_file: ~/.gzh/net-env/vpn/office-vpn.conf

💡 Import profile: gz net-env profiles import ~/.gzh/exports/net-env-office-20250902.yaml

stderr: (empty)
Exit Code: 0
```

### Side Effects

**Files Created**:

- Profile creation: `~/.gzh/net-env/profiles/<name>.yaml`
- VPN configs: `~/.gzh/net-env/vpn/<name>-vpn.conf`
- Export files: `~/.gzh/exports/net-env-<name>-<timestamp>.yaml`

**Files Modified**:

- Profile edits: existing profile YAML files
- Profile registry: `~/.gzh/net-env/profiles.json`

**State Changes**:

- Profile database updated
- VPN client configurations modified
- Auto-switch rules registry updated

### Validation

**Automated Tests**:

```bash
# Test profile listing
result=$(gz net-env profiles list 2>&1)
exit_code=$?

assert_contains "$result" "Network Environment Profiles"
assert_exit_code 0

# Test profile creation
gz net-env profiles create test-profile --dns "1.1.1.1,1.0.0.1" --no-vpn --no-proxy
assert_exit_code 0
assert_file_exists "$HOME/.gzh/net-env/profiles/test-profile.yaml"

# Test profile deletion
gz net-env profiles delete test-profile --confirm
assert_exit_code 0
assert_file_not_exists "$HOME/.gzh/net-env/profiles/test-profile.yaml"
```

**Manual Verification**:

1. List profiles and verify information accuracy
1. Create new profile with interactive prompts
1. Edit existing profile and confirm changes
1. Export profile and verify YAML format
1. Delete profile and confirm cleanup
1. Test auto-switch network pattern matching

### Edge Cases

**Profile Name Validation**:

- Invalid characters in profile names
- Reserved names (system, default, current)
- Duplicate profile name handling
- Case sensitivity considerations

**Configuration Validation**:

- Invalid IP addresses for DNS/proxy
- Unreachable VPN servers
- Invalid WiFi network patterns
- Malformed configuration files

**File System Issues**:

- Permission denied for profile directory
- Disk space exhaustion
- Corrupted profile configuration files
- Missing VPN configuration files

**Concurrent Operations**:

- Multiple profile edits simultaneously
- Profile deletion while in use
- Export during active network switching
- Import conflicts with existing profiles

### Performance Expectations

**Response Time**:

- List profiles: < 1 second
- Create profile: < 3 seconds
- Edit profile: < 2 seconds
- Export profile: < 1 second
- Delete profile: < 2 seconds

**Resource Usage**:

- Memory: < 30MB
- Disk: Profile files typically < 5KB each
- Network: No network calls for profile management

**Storage**:

- Profile configs: < 5KB per profile
- VPN configs: Variable size (typically < 10KB)
- Export files: < 10KB per export
- Log files: Rotated automatically

## Notes

- Profile validation during creation and editing
- Backup system for profile configurations
- Import/export functionality for profile sharing
- Template system for common profile types
- Integration with external VPN clients
- WiFi network pattern matching with wildcards
- Automatic profile cleanup for unused profiles
- Version control for profile configuration changes
