package quality

import (
	"github.com/spf13/cobra"

	"github.com/Gizzahub/gzh-cli/cmd/registry"
)

type qualityCmdProvider struct{}

func (qualityCmdProvider) Command() *cobra.Command {
	return NewQualityCmd()
}

func init() {
	registry.Register(qualityCmdProvider{})
}
