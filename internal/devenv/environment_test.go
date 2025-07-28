// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package devenv

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadEnvironment(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantErr     bool
		wantName    string
		wantService bool
	}{
		{
			name: "valid environment",
			input: `
name: test-env
description: Test environment
services:
  aws:
    profile: test-profile
    region: us-west-2
dependencies:
  - aws -> kubernetes
`,
			wantErr:     false,
			wantName:    "test-env",
			wantService: true,
		},
		{
			name: "missing name",
			input: `
services:
  aws:
    profile: test-profile
`,
			wantErr: true,
		},
		{
			name:    "invalid yaml",
			input:   `invalid: yaml: content:`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env, err := LoadEnvironment([]byte(tt.input))

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantName, env.Name)

			if tt.wantService {
				assert.True(t, env.HasService("aws"))
				assert.Equal(t, []string{"aws"}, env.GetServiceNames())
			}
		})
	}
}

func TestEnvironmentValidate(t *testing.T) {
	tests := []struct {
		name    string
		env     Environment
		wantErr bool
	}{
		{
			name: "valid environment",
			env: Environment{
				Name: "test",
				Services: map[string]ServiceConfig{
					"aws": {
						AWS: &AWSConfig{Profile: "test"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing name",
			env: Environment{
				Services: map[string]ServiceConfig{
					"aws": {AWS: &AWSConfig{Profile: "test"}},
				},
			},
			wantErr: true,
		},
		{
			name: "no services",
			env: Environment{
				Name:     "test",
				Services: map[string]ServiceConfig{},
			},
			wantErr: true,
		},
		{
			name: "invalid dependency",
			env: Environment{
				Name: "test",
				Services: map[string]ServiceConfig{
					"aws": {AWS: &AWSConfig{Profile: "test"}},
				},
				Dependencies: []string{"aws -> nonexistent"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.env.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
