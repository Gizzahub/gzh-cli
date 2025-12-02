// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

// Package registry provides command registration and lifecycle management.
package registry

import (
	"fmt"
	"os"
	"os/exec"
)

// LifecycleManager handles command lifecycle checks and warnings.
type LifecycleManager struct {
	allowExperimental bool
	allowDeprecated   bool
}

// NewLifecycleManager creates a new lifecycle manager.
func NewLifecycleManager() *LifecycleManager {
	return &LifecycleManager{
		allowExperimental: os.Getenv("GZ_EXPERIMENTAL") == "1",
		allowDeprecated:   true, // Always allow but warn
	}
}

// CheckCommand validates if a command can be executed based on its lifecycle stage.
func (lm *LifecycleManager) CheckCommand(meta CommandMetadata) error {
	switch meta.Lifecycle {
	case LifecycleExperimental:
		if !lm.allowExperimental {
			return fmt.Errorf(
				"command '%s' is experimental and disabled by default\n"+
					"To enable experimental features, set: export GZ_EXPERIMENTAL=1",
				meta.Name,
			)
		}
		lm.showExperimentalWarning(meta)

	case LifecycleDeprecated:
		lm.showDeprecationWarning(meta)

	case LifecycleBeta:
		lm.showBetaWarning(meta)
	}

	return nil
}

// showExperimentalWarning displays a warning for experimental features.
func (lm *LifecycleManager) showExperimentalWarning(meta CommandMetadata) {
	fmt.Fprintf(os.Stderr, "⚠️  Warning: Command '%s' is experimental and may change or be removed\n", meta.Name)
	fmt.Fprintf(os.Stderr, "   Version: %s | Status: %s\n\n", meta.Version, meta.Lifecycle)
}

// showDeprecationWarning displays a warning for deprecated features.
func (lm *LifecycleManager) showDeprecationWarning(meta CommandMetadata) {
	fmt.Fprintf(os.Stderr, "⚠️  DEPRECATED: Command '%s' is deprecated and will be removed in a future version\n", meta.Name)
	fmt.Fprintf(os.Stderr, "   Current Version: %s | Please migrate to alternatives\n\n", meta.Version)
}

// showBetaWarning displays a warning for beta features.
func (lm *LifecycleManager) showBetaWarning(meta CommandMetadata) {
	fmt.Fprintf(os.Stderr, "ℹ️  Info: Command '%s' is in beta testing\n", meta.Name)
	fmt.Fprintf(os.Stderr, "   Version: %s | Please report any issues\n\n", meta.Version)
}

// CheckDependencies validates that all required external tools are available.
func (lm *LifecycleManager) CheckDependencies(meta CommandMetadata) []string {
	var missing []string

	for _, dep := range meta.Dependencies {
		if !isCommandAvailable(dep) {
			missing = append(missing, dep)
		}
	}

	return missing
}

// ShowDependencyWarning displays a warning for missing dependencies.
func (lm *LifecycleManager) ShowDependencyWarning(cmdName string, missing []string) {
	if len(missing) == 0 {
		return
	}

	fmt.Fprintf(os.Stderr, "⚠️  Warning: Command '%s' requires missing dependencies:\n", cmdName)
	for _, dep := range missing {
		fmt.Fprintf(os.Stderr, "   - %s\n", dep)
	}
	fmt.Fprintf(os.Stderr, "\n")
}

// isCommandAvailable checks if a command is available in PATH.
func isCommandAvailable(cmd string) bool {
	// 절대 경로 또는 상대 경로인 경우 직접 확인
	_, err := os.Stat(cmd)
	if err == nil {
		return true
	}

	// PATH에서 명령어 검색
	_, err = exec.LookPath(cmd)
	return err == nil
}

// FilterCommands filters command providers based on lifecycle settings.
func (lm *LifecycleManager) FilterCommands(providers []CommandProvider) []CommandProvider {
	result := make([]CommandProvider, 0, len(providers))

	for _, p := range providers {
		meta := GetMetadata(p)

		// Filter experimental commands if not enabled
		if meta.Lifecycle == LifecycleExperimental && !lm.allowExperimental {
			continue
		}

		result = append(result, p)
	}

	return result
}

// EnableExperimental enables experimental features.
func (lm *LifecycleManager) EnableExperimental() {
	lm.allowExperimental = true
}

// DisableExperimental disables experimental features.
func (lm *LifecycleManager) DisableExperimental() {
	lm.allowExperimental = false
}

// IsExperimentalEnabled returns whether experimental features are enabled.
func (lm *LifecycleManager) IsExperimentalEnabled() bool {
	return lm.allowExperimental
}
