package profile

import (
	"github.com/spf13/cobra"

	"github.com/Gizzahub/gzh-cli/cmd/registry"
)

type profileCmdProvider struct{}

func (profileCmdProvider) Command() *cobra.Command {
	return NewProfileCmd()
}

func init() {
	registry.Register(profileCmdProvider{})
}
