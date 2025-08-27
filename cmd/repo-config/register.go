package repoconfig

import (
	"github.com/spf13/cobra"

	"github.com/Gizzahub/gzh-cli/cmd/registry"
)

type repoConfigCmdProvider struct{}

func (repoConfigCmdProvider) Command() *cobra.Command {
	return NewRepoConfigCmd()
}

func init() {
	registry.Register(repoConfigCmdProvider{})
}
