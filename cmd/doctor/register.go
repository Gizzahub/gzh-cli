// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package doctor

import (
	"github.com/spf13/cobra"

	"github.com/gizzahub/gzh-cli/cmd/registry"
)

type doctorCmdProvider struct{}

func (doctorCmdProvider) Command() *cobra.Command {
	DoctorCmd.Hidden = true
	return DoctorCmd
}

func (doctorCmdProvider) Metadata() registry.CommandMetadata {
	return registry.CommandMetadata{
		Name:         "doctor",
		Category:     registry.CategoryQuality,
		Version:      "1.0.0",
		Priority:     70,
		Experimental: false,
		Dependencies: []string{},
		Tags:         []string{"doctor", "health", "check", "diagnostics", "system"},
		Lifecycle:    registry.LifecycleStable,
	}
}

func init() {
	registry.Register(doctorCmdProvider{})
}
