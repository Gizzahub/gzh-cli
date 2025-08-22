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
	version := "v1.0.0"
	cmd := NewVersionCmd(version) // 공개 함수 사용
	b := bytes.NewBufferString("")
	cmd.SetOut(b)

	err := cmd.Execute()
	require.NoError(t, err)

	out, err := io.ReadAll(b)
	require.NoError(t, err)

	assert.Equal(t, fmt.Sprintf("gz version %s\n", version), string(out))
}
