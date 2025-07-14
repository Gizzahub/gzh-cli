// Package ssh_config provides SSH configuration management for Git operations
// and repository access across multiple platforms.
//
// This package implements the ssh-config command that manages SSH keys,
// configurations, and authentication for Git platforms, enabling secure
// and efficient repository operations.
//
// Key Features:
//
// SSH Key Management:
//   - SSH key generation and rotation
//   - Multiple key support for different platforms
//   - Key format conversion and compatibility
//   - Secure key storage and access
//
// Configuration Generation:
//   - Automatic SSH config file generation
//   - Platform-specific SSH settings
//   - Host alias configuration
//   - Connection optimization settings
//
// Platform Integration:
//   - GitHub SSH key management
//   - GitLab SSH configuration
//   - Gitea/Gogs SSH setup
//   - Enterprise Git platform support
//
// Security Features:
//   - SSH agent integration
//   - Key passphrase management
//   - Connection security validation
//   - Access audit and logging
//
// Example usage:
//
//	gz ssh-config generate --platforms github,gitlab
//	gz ssh-config add-key --platform github --key-file ~/.ssh/id_rsa.pub
//	gz ssh-config validate --config ~/.ssh/config
//	gz ssh-config rotate-keys --all-platforms
//
// The package ensures secure and reliable SSH connectivity to Git platforms,
// handling the complexity of managing multiple SSH keys and configurations
// for different platforms and use cases.
package sshconfig
