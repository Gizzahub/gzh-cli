// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package testutil

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecutablePathForGOOS(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		goos string
		want string
	}{
		{name: "Unix", goos: "linux", want: filepath.Join("bin", "gz")},
		{name: "Windows", goos: "windows", want: filepath.Join("bin", "gz.exe")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, executablePathForGOOS("bin", "gz", tt.goos))
		})
	}
}
