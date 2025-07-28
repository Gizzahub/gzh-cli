// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

// Package main implements the git-synclone command, a Git extension for
// enhanced repository cloning with multi-provider support.
//
// git-synclone provides intelligent repository cloning capabilities for
// GitHub, GitLab, Gitea, and Gogs platforms. It supports:
//
//   - Bulk cloning of entire organizations or groups
//   - Parallel execution for faster cloning
//   - Resume capability for interrupted operations
//   - Multiple clone strategies (reset, pull, fetch)
//   - Configuration file support
//   - State management for complex operations
//
// Installation:
//
// The git-synclone binary should be placed in your PATH. Once installed,
// it can be invoked as a Git extension:
//
//	git synclone github -o myorg
//	git synclone gitlab -g mygroup --recursive
//	git synclone all -c config.yaml
//
// The command integrates with the existing gz synclone functionality while
// providing a more Git-native interface.
package main
