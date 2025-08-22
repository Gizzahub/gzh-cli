// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package event

import "github.com/spf13/cobra"

// Adapter to keep old constructor available via package git
func NewEventCmdAdapter() *cobra.Command { return NewEventCmd() }
