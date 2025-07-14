// Package net_env provides intelligent network environment management and
// transition capabilities for development workflows.
//
// This package implements the net-env command that automatically detects
// network environment changes and manages development configurations
// accordingly, enabling seamless transitions between different network
// contexts (home, office, coffee shop, etc.).
//
// Key Features:
//
// Network Detection:
//   - WiFi network change monitoring
//   - Network profile identification
//   - Automatic environment classification
//   - Connection quality assessment
//
// Environment Management:
//   - VPN connection management and automation
//   - DNS configuration switching
//   - Proxy settings management
//   - Docker network profile switching
//   - Kubernetes context management
//
// Configuration Profiles:
//   - Location-based configuration profiles
//   - Security policy enforcement
//   - Network-specific development settings
//   - Automated environment provisioning
//
// Security Features:
//   - Automatic VPN connection on untrusted networks
//   - DNS over HTTPS configuration
//   - Network security policy enforcement
//   - Sensitive data protection on public networks
//
// Example usage:
//
//	gz net-env monitor --verbose
//	gz net-env switch --profile office
//	gz net-env vpn --auto-connect
//	gz net-env docker --profile home
//
// The package maintains a hierarchy of network profiles and automatically
// applies the most appropriate configuration based on the detected network
// environment, ensuring optimal development experience across different
// locations and network conditions.
package netenv
