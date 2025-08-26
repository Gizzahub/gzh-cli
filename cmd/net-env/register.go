package netenv

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/Gizzahub/gzh-cli/cmd/registry"
)

type netEnvCmdProvider struct{}

func (netEnvCmdProvider) Command() *cobra.Command {
	return NewNetEnvCmd(context.Background())
}

func init() {
	registry.Register(netEnvCmdProvider{})
}
