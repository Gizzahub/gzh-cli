package pm

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/Gizzahub/gzh-cli/cmd/registry"
)

type pmCmdProvider struct{}

func (pmCmdProvider) Command() *cobra.Command {
	return NewPMCmd(context.Background())
}

func init() {
	registry.Register(pmCmdProvider{})
}
