//nolint:testpackage // White-box testing needed for internal function access
package cmd

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Gizzahub/gzh-cli/internal/app"
)

func TestRootCommandOutput(t *testing.T) {
	cmd := NewRootCmd(context.Background(), "", app.NewTestAppContext())
	b := bytes.NewBufferString("")

	cmd.SetArgs([]string{"-h"})
	cmd.SetOut(b)

	cmdErr := cmd.RunE(cmd, nil)
	require.NoError(t, cmdErr)
}
