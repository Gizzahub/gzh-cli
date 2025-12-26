// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

//go:build !devenv_external
// +build !devenv_external

package cmd

// NOTE: When devenv_external build tag is NOT set, the dev-env command
// is registered via the cmd/dev-env package (see root.go).
//
// The external wrapper (dev_env_wrapper.go) is available for future migration
// when all dev-env subcommands are moved to the gzh-cli-dev-env library.
//
// To test the external wrapper, build with: go build -tags devenv_external
//
// Current status:
// - Library provides: status, tui, switch-all commands
// - Old package provides: kubeconfig, docker, aws, gcloud, azure, ssh commands
//
// Migration plan:
// 1. Move remaining subcommands to gzh-cli-dev-env library
// 2. Enable devenv_external build tag
// 3. Remove cmd/dev-env package
