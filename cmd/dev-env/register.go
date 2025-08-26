package devenv

import (
	"github.com/spf13/cobra"

	"github.com/Gizzahub/gzh-cli/cmd/registry"
)

type devEnvCmdProvider struct{}

func (devEnvCmdProvider) Command() *cobra.Command {
	return NewDevEnvCmd()
}

func init() {
	registry.Register(devEnvCmdProvider{})
}
