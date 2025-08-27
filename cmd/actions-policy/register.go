package actionspolicy

import (
	"github.com/spf13/cobra"

	"github.com/Gizzahub/gzh-cli/cmd/registry"
)

type actionsPolicyCmdProvider struct{}

func (actionsPolicyCmdProvider) Command() *cobra.Command {
	return NewActionsPolicyCmd()
}

func init() {
	registry.Register(actionsPolicyCmdProvider{})
}
