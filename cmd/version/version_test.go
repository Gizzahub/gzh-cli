//nolint:testpackage // White-box testing needed for internal function access
package version

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionCommand(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected string
	}{
		{"release version", "v1.0.0", fmt.Sprintf("gz version %s\n", "v1.0.0")},
		{"dev version", "dev", "gz version dev\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewVersionCmd(tt.version)
			b := bytes.NewBufferString("")
			cmd.SetOut(b)

			err := cmd.Execute()
			require.NoError(t, err)

			out, err := io.ReadAll(b)
			require.NoError(t, err)

			assert.Equal(t, tt.expected, string(out))
		})
	}
}
