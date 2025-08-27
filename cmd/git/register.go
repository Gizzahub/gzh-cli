package git

import (
	"github.com/spf13/cobra"

	"github.com/Gizzahub/gzh-cli/cmd/registry"
)

type gitCmdProvider struct{}

func (gitCmdProvider) Command() *cobra.Command {
	return NewGitCmd()
}

func init() {
	registry.Register(gitCmdProvider{})
}
