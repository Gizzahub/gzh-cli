// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package registry

import (
	"sync"

	"github.com/spf13/cobra"
)

// CommandProvider defines an interface that exposes a Cobra command.
type CommandProvider interface {
	Command() *cobra.Command
}

var (
	mu        sync.RWMutex
	providers []CommandProvider
)

// Register adds a command provider to the registry.
func Register(p CommandProvider) {
	mu.Lock()
	providers = append(providers, p)
	mu.Unlock()
}

// List returns all registered command providers.
func List() []CommandProvider {
	mu.RLock()
	defer mu.RUnlock()
	return append([]CommandProvider(nil), providers...)
}
