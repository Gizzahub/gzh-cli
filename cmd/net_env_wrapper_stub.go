// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

//go:build !netenv_external
// +build !netenv_external

package cmd

// NOTE: When netenv_external build tag is NOT set, the net-env command
// is registered via the cmd/net-env package (see root.go).
//
// The external wrapper (net_env_wrapper.go) is available for future migration
// when all net-env subcommands are moved to the gzh-cli-net-env library.
//
// To test the external wrapper, build with: go build -tags netenv_external
//
// Current status:
// - Library provides: status, watch, profile commands (via gzh-cli-net-env)
// - Old package provides: actions, cloud, tui commands (cmd/net-env)
//
// Migration plan:
// 1. Move remaining subcommands to gzh-cli-net-env library
// 2. Enable netenv_external build tag
// 3. Remove cmd/net-env package
//
// Library packages:
// - pkg/wifi: WiFi detection (macOS/Linux)
// - pkg/vpn: VPN detection (macOS/Linux, WireGuard)
// - pkg/dns: DNS configuration detection
// - pkg/proxy: Proxy configuration detection
// - pkg/profile: Network profile management
// - pkg/config: Configuration loading/saving
// - pkg/monitor: Network change monitoring
// - pkg/tui: Terminal UI dashboard
