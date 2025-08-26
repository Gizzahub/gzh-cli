package ide

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/Gizzahub/gzh-cli/cmd/registry"
)

type ideCmdProvider struct{}

func (ideCmdProvider) Command() *cobra.Command {
	return NewIDECmd(context.Background())
}

func init() {
	registry.Register(ideCmdProvider{})
}
